import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { GeneratorStopSignal, GeneratorYieldSignal, ReturnSignal } from "./signals";
import type { IteratorStep, IteratorValue, V10Value } from "./values";

type BlockState = {
  env: Environment;
  index: number;
  result: V10Value;
};

type ForLoopState = {
  mode: "static" | "iterator";
  values?: V10Value[];
  iterator?: IteratorValue;
  baseEnv: Environment;
  index: number;
  result: V10Value;
  iterationEnv?: Environment;
  awaitingBody: boolean;
  pendingValue?: V10Value;
  iteratorClosed?: boolean;
};

type WhileLoopState = {
  baseEnv: Environment;
  result: V10Value;
  inBody: boolean;
  loopEnv?: Environment;
  conditionInProgress: boolean;
};

type IfExpressionState = {
  stage: "if_condition" | "if_body" | "or_condition" | "or_body";
  orIndex: number;
  result?: V10Value;
};

type MatchExpressionState = {
  stage: "subject" | "clause" | "guard" | "body";
  clauseIndex: number;
  subject?: V10Value;
  matchEnv?: Environment;
};

class GeneratorContext {
  readonly iteratorValue: IteratorValue;

  private index = 0;
  private started = false;
  private busy = false;
  private done = false;
  private closed = false;
  private pendingValue: V10Value | null = null;
  private storedError: unknown = null;
  private resumeCurrentStatement = false;
  private currentStatementIndex = -1;

  private readonly blockStates = new WeakMap<AST.BlockExpression, BlockState>();
  private readonly forLoopStates = new WeakMap<AST.ForLoop, ForLoopState>();
  private readonly whileLoopStates = new WeakMap<AST.WhileLoop, WhileLoopState>();
  private readonly ifStates = new WeakMap<AST.IfExpression, IfExpressionState>();
  private readonly matchStates = new WeakMap<AST.MatchExpression, MatchExpressionState>();

  constructor(
    private readonly interpreter: InterpreterV10,
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

  controllerValue(): V10Value {
    const yieldFn = this.interpreter.makeNativeFunction("Generator.yield", 1, (_interp, args) => {
      const value = args[0] ?? { kind: "nil", value: null };
      this.emit(value);
      return { kind: "nil", value: null };
    });
    const stopFn = this.interpreter.makeNativeFunction("Generator.stop", 0, () => {
      this.stop();
      return { kind: "nil", value: null };
    });
    const map = new Map<string, V10Value>();
    map.set("yield", yieldFn);
    map.set("stop", stopFn);
    return {
      kind: "struct_instance",
      def: this.interpreter.generatorControllerStruct,
      values: map,
    };
  }

  emit(value: V10Value): never {
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
    return this.blockStates.get(node);
  }

  setBlockState(node: AST.BlockExpression, state: BlockState): void {
    this.blockStates.set(node, state);
  }

  clearBlockState(node: AST.BlockExpression): void {
    this.blockStates.delete(node);
  }

  getForLoopState(node: AST.ForLoop): ForLoopState | undefined {
    return this.forLoopStates.get(node);
  }

  setForLoopState(node: AST.ForLoop, state: ForLoopState): void {
    this.forLoopStates.set(node, state);
  }

  clearForLoopState(node: AST.ForLoop): void {
    this.forLoopStates.delete(node);
  }

  getWhileLoopState(node: AST.WhileLoop): WhileLoopState | undefined {
    return this.whileLoopStates.get(node);
  }

  setWhileLoopState(node: AST.WhileLoop, state: WhileLoopState): void {
    this.whileLoopStates.set(node, state);
  }

  clearWhileLoopState(node: AST.WhileLoop): void {
    this.whileLoopStates.delete(node);
  }

  getIfState(node: AST.IfExpression): IfExpressionState | undefined {
    return this.ifStates.get(node);
  }

  setIfState(node: AST.IfExpression, state: IfExpressionState): void {
    this.ifStates.set(node, state);
  }

  clearIfState(node: AST.IfExpression): void {
    this.ifStates.delete(node);
  }

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined {
    return this.matchStates.get(node);
  }

  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void {
    this.matchStates.set(node, state);
  }

  clearMatchState(node: AST.MatchExpression): void {
    this.matchStates.delete(node);
  }
}

declare module "./index" {
  interface InterpreterV10 {
    iteratorNativeMethods: {
      next: Extract<V10Value, { kind: "native_function" }>;
      close: Extract<V10Value, { kind: "native_function" }>;
    };
    iteratorEndValue: Extract<V10Value, { kind: "iterator_end" }>;
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

export function applyIteratorAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureIteratorBuiltins = function ensureIteratorBuiltins(this: InterpreterV10): void {
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

  cls.prototype.pushGeneratorContext = function pushGeneratorContext(this: InterpreterV10, ctx: GeneratorContext): void {
    this.ensureIteratorBuiltins();
    this.generatorStack.push(ctx);
  };

  cls.prototype.popGeneratorContext = function popGeneratorContext(this: InterpreterV10, ctx: GeneratorContext): void {
    if (!this.generatorStack || this.generatorStack.length === 0) return;
    const top = this.generatorStack[this.generatorStack.length - 1];
    if (top === ctx) {
      this.generatorStack.pop();
    }
  };

  cls.prototype.currentGeneratorContext = function currentGeneratorContext(this: InterpreterV10): GeneratorContext | null {
    if (!this.generatorStack || this.generatorStack.length === 0) return null;
    return this.generatorStack[this.generatorStack.length - 1]!;
  };

  cls.prototype.createIteratorFromLiteral = function createIteratorFromLiteral(this: InterpreterV10, node: AST.IteratorLiteral, env: Environment): IteratorValue {
    this.ensureIteratorBuiltins();
    const generatorEnv = new Environment(env);
    const context = new GeneratorContext(this, generatorEnv, node.body);
    generatorEnv.define("gen", context.controllerValue());
    return context.iteratorValue;
  };
}

export function evaluateIteratorLiteral(ctx: InterpreterV10, node: AST.IteratorLiteral, env: Environment): V10Value {
  return ctx.createIteratorFromLiteral(node, env);
}

export function evaluateYieldStatement(ctx: InterpreterV10, node: AST.YieldStatement, env: Environment): V10Value {
  const generator = ctx.currentGeneratorContext();
  if (!generator) throw new Error("yield must be used inside a generator literal");
  const value = node.expression ? ctx.evaluate(node.expression, env) : { kind: "nil", value: null };
  generator.emit(value);
  return { kind: "nil", value: null };
}
