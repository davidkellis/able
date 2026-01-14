import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import { GeneratorStopSignal, GeneratorYieldSignal, ProcYieldSignal, ReturnSignal } from "./signals";
import type { IteratorStep, IteratorValue, RuntimeValue } from "./values";
import type {
  BlockState,
  ContinuationContext,
  ForLoopState,
  IfExpressionState,
  MatchExpressionState,
  LoopExpressionState,
  WhileLoopState,
  StringInterpolationState,
  EnsureState,
} from "./continuations";

class GeneratorContext implements ContinuationContext {
  readonly kind = "generator" as const;
  readonly iteratorValue: IteratorValue;

  private index = 0;
  private started = false;
  private busy = false;
  private done = false;
  private closed = false;
  private pendingValue: RuntimeValue | null = null;
  private storedError: unknown = null;
  private resumeCurrentStatement = false;
  private currentStatementIndex = -1;

  private readonly blockStates = new WeakMap<AST.BlockExpression, BlockState[]>();
  private readonly forLoopStates = new WeakMap<AST.ForLoop, ForLoopState[]>();
  private readonly whileLoopStates = new WeakMap<AST.WhileLoop, WhileLoopState[]>();
  private readonly loopStates = new WeakMap<AST.LoopExpression, LoopExpressionState[]>();
  private readonly ifStates = new WeakMap<AST.IfExpression, IfExpressionState[]>();
  private readonly matchStates = new WeakMap<AST.MatchExpression, MatchExpressionState[]>();
  private readonly stringInterpolationStates = new WeakMap<AST.StringInterpolation, StringInterpolationState[]>();
  private readonly ensureStates = new WeakMap<AST.EnsureExpression, EnsureState[]>();

  constructor(
    private readonly interpreter: Interpreter,
    readonly env: Environment,
    private readonly body: AST.Statement[],
  ) {
    this.iteratorValue = {
      kind: "iterator",
      iterator: {
        next: () => this.next(),
        close: () => this.close(),
      },
    };
  }

  controllerValue(): RuntimeValue {
    const yieldFn = this.interpreter.makeNativeFunction("Generator.yield", 1, (_interp, args) => {
      const value = args[0] ?? { kind: "nil", value: null };
      this.emit(value);
      return { kind: "nil", value: null };
    });
    const stopFn = this.interpreter.makeNativeFunction("Generator.stop", 0, () => {
      this.stop();
      return { kind: "nil", value: null };
    });
    const map = new Map<string, RuntimeValue>();
    map.set("yield", yieldFn);
    map.set("stop", stopFn);
    return {
      kind: "struct_instance",
      def: this.interpreter.generatorControllerStruct,
      values: map,
    };
  }

  emit(value: RuntimeValue): never {
    if (this.closed) {
      throw new Error("Cannot yield from a closed iterator");
    }
    this.pendingValue = value;
    throw new GeneratorYieldSignal();
  }

  stop(): never {
    this.closed = true;
    this.done = true;
    this.index = this.body.length;
    throw new GeneratorStopSignal();
  }

  close(): void {
    if (this.closed) return;
    this.closed = true;
    this.done = true;
  }

  next(): IteratorStep {
    if (this.closed) {
      this.done = true;
      if (this.storedError) {
        throw this.storedError;
      }
      return { done: true, value: this.interpreter.iteratorEndValue };
    }
    if (this.busy) {
      throw new Error("iterator.next re-entered while suspended at yield");
    }
    if (this.done) {
      if (this.storedError) {
        throw this.storedError;
      }
      return { done: true, value: this.interpreter.iteratorEndValue };
    }

    this.busy = true;
    this.interpreter.pushGeneratorContext(this);
    try {
      if (!this.started) {
        this.started = true;
      }
      while (this.index < this.body.length) {
        const stmt = this.body[this.index]!;
        this.beginStatement(this.index);
        try {
          this.interpreter.evaluate(stmt, this.env);
        } catch (err) {
          if (err instanceof GeneratorYieldSignal) {
            const value = this.pendingValue ?? { kind: "nil", value: null };
            this.pendingValue = null;
            const repeat = this.resumeCurrentStatement;
            if (!repeat) {
              this.index += 1;
            }
            this.finishStatement();
            return { done: false, value };
          }
          if (err instanceof ProcYieldSignal) {
            this.pendingValue = null;
            const repeat = this.resumeCurrentStatement;
            if (!repeat) {
              this.index += 1;
            }
            this.finishStatement();
            throw err;
          }
          if (err instanceof GeneratorStopSignal) {
            this.pendingValue = null;
            this.done = true;
            this.finishStatement();
            return { done: true, value: this.interpreter.iteratorEndValue };
          }
          if (err instanceof ReturnSignal) {
            this.pendingValue = null;
            this.done = true;
            this.finishStatement();
            return { done: true, value: this.interpreter.iteratorEndValue };
          }
          this.pendingValue = null;
          this.done = true;
          this.storedError = err;
          this.finishStatement();
          throw err;
        }
        this.index += 1;
        this.finishStatement();
      }
      this.done = true;
      return { done: true, value: this.interpreter.iteratorEndValue };
    } finally {
      this.interpreter.popGeneratorContext(this);
      this.busy = false;
    }
  }

  markStatementIncomplete(): void {
    this.resumeCurrentStatement = true;
  }

  private beginStatement(index: number): void {
    this.currentStatementIndex = index;
    this.resumeCurrentStatement = false;
  }

  private finishStatement(): void {
    this.currentStatementIndex = -1;
    this.resumeCurrentStatement = false;
  }

  getBlockState(node: AST.BlockExpression): BlockState | undefined {
    const stack = this.blockStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setBlockState(node: AST.BlockExpression, state: BlockState): void {
    const stack = this.blockStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.blockStates.set(node, [state]);
    }
  }

  clearBlockState(node: AST.BlockExpression): void {
    const stack = this.blockStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.blockStates.delete(node);
  }

  getForLoopState(node: AST.ForLoop): ForLoopState | undefined {
    const stack = this.forLoopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setForLoopState(node: AST.ForLoop, state: ForLoopState): void {
    const stack = this.forLoopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.forLoopStates.set(node, [state]);
    }
  }

  clearForLoopState(node: AST.ForLoop): void {
    const stack = this.forLoopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.forLoopStates.delete(node);
  }

  getWhileLoopState(node: AST.WhileLoop): WhileLoopState | undefined {
    const stack = this.whileLoopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setWhileLoopState(node: AST.WhileLoop, state: WhileLoopState): void {
    const stack = this.whileLoopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.whileLoopStates.set(node, [state]);
    }
  }

  clearWhileLoopState(node: AST.WhileLoop): void {
    const stack = this.whileLoopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.whileLoopStates.delete(node);
  }

  getLoopExpressionState(node: AST.LoopExpression): LoopExpressionState | undefined {
    const stack = this.loopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setLoopExpressionState(node: AST.LoopExpression, state: LoopExpressionState): void {
    const stack = this.loopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.loopStates.set(node, [state]);
    }
  }

  clearLoopExpressionState(node: AST.LoopExpression): void {
    const stack = this.loopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.loopStates.delete(node);
  }

  getIfState(node: AST.IfExpression): IfExpressionState | undefined {
    const stack = this.ifStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setIfState(node: AST.IfExpression, state: IfExpressionState): void {
    const stack = this.ifStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.ifStates.set(node, [state]);
    }
  }

  clearIfState(node: AST.IfExpression): void {
    const stack = this.ifStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.ifStates.delete(node);
  }

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined {
    const stack = this.matchStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void {
    const stack = this.matchStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.matchStates.set(node, [state]);
    }
  }

  clearMatchState(node: AST.MatchExpression): void {
    const stack = this.matchStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.matchStates.delete(node);
  }

  getStringInterpolationState(node: AST.StringInterpolation): StringInterpolationState | undefined {
    const stack = this.stringInterpolationStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setStringInterpolationState(node: AST.StringInterpolation, state: StringInterpolationState): void {
    const stack = this.stringInterpolationStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.stringInterpolationStates.set(node, [state]);
    }
  }

  clearStringInterpolationState(node: AST.StringInterpolation): void {
    const stack = this.stringInterpolationStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.stringInterpolationStates.delete(node);
  }

  getEnsureState(node: AST.EnsureExpression): EnsureState | undefined {
    const stack = this.ensureStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setEnsureState(node: AST.EnsureExpression, state: EnsureState): void {
    const stack = this.ensureStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.ensureStates.set(node, [state]);
    }
  }

  clearEnsureState(node: AST.EnsureExpression): void {
    const stack = this.ensureStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.ensureStates.delete(node);
  }
}

declare module "./index" {
  interface Interpreter {
    iteratorNativeMethods: {
      next: Extract<RuntimeValue, { kind: "native_function" }>;
      close: Extract<RuntimeValue, { kind: "native_function" }>;
    };
    iteratorEndValue: Extract<RuntimeValue, { kind: "iterator_end" }>;
    generatorControllerStruct: AST.StructDefinition;
    generatorBuiltinsInitialized: boolean;
    generatorStack: GeneratorContext[];
    ensureIteratorBuiltins(): void;
    createIteratorFromLiteral(node: AST.IteratorLiteral, env: Environment): IteratorValue;
    pushGeneratorContext(ctx: GeneratorContext): void;
    popGeneratorContext(ctx: GeneratorContext): void;
    currentGeneratorContext(): GeneratorContext | null;
  }
}

export function applyIteratorAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureIteratorBuiltins = function ensureIteratorBuiltins(this: Interpreter): void {
    if (this.generatorBuiltinsInitialized) return;
    this.generatorBuiltinsInitialized = true;
    this.iteratorEndValue = { kind: "iterator_end" };
    this.generatorStack = [];
    this.generatorControllerStruct = AST.structDefinition(
      "__GeneratorController",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("Function"), "yield"),
        AST.structFieldDefinition(AST.simpleTypeExpression("Function"), "stop"),
      ],
      "named",
    );
    this.iteratorNativeMethods = {
      next: this.makeNativeFunction("Iterator.next", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "iterator") throw new Error("Iterator.next called on non-iterator");
        const step = self.iterator.next();
        if (step.done) return interp.iteratorEndValue;
        return step.value;
      }),
      close: this.makeNativeFunction("Iterator.close", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "iterator") throw new Error("Iterator.close called on non-iterator");
        self.iterator.close();
        return { kind: "nil", value: null };
      }),
    };
  };

  cls.prototype.pushGeneratorContext = function pushGeneratorContext(this: Interpreter, ctx: GeneratorContext): void {
    this.ensureIteratorBuiltins();
    this.generatorStack.push(ctx);
  };

  cls.prototype.popGeneratorContext = function popGeneratorContext(this: Interpreter, ctx: GeneratorContext): void {
    if (!this.generatorStack || this.generatorStack.length === 0) return;
    const top = this.generatorStack[this.generatorStack.length - 1];
    if (top === ctx) {
      this.generatorStack.pop();
    }
  };

  cls.prototype.currentGeneratorContext = function currentGeneratorContext(this: Interpreter): GeneratorContext | null {
    if (!this.generatorStack || this.generatorStack.length === 0) return null;
    return this.generatorStack[this.generatorStack.length - 1]!;
  };

  cls.prototype.createIteratorFromLiteral = function createIteratorFromLiteral(this: Interpreter, node: AST.IteratorLiteral, env: Environment): IteratorValue {
    this.ensureIteratorBuiltins();
    const generatorEnv = new Environment(env);
    const context = new GeneratorContext(this, generatorEnv, node.body);
    const controller = context.controllerValue();
    const bindingName = node.binding?.name ?? "gen";
    generatorEnv.define(bindingName, controller);
    if (bindingName !== "gen") {
      generatorEnv.define("gen", controller);
    }
    return context.iteratorValue;
  };
}

export function evaluateIteratorLiteral(ctx: Interpreter, node: AST.IteratorLiteral, env: Environment): RuntimeValue {
  return ctx.createIteratorFromLiteral(node, env);
}

export function evaluateYieldStatement(ctx: Interpreter, node: AST.YieldStatement, env: Environment): RuntimeValue {
  const generator = ctx.currentGeneratorContext();
  if (!generator) throw new Error("yield must be used inside a generator literal");
  const value = node.expression ? ctx.evaluate(node.expression, env) : { kind: "nil", value: null };
  generator.emit(value);
  return { kind: "nil", value: null };
}
