import * as AST from "./ast";
import { Environment } from "./environment";
import type { Interpreter } from "./interpreter";
import type { AbleFunction, AbleValue } from "./runtime";
import {
  createError,
  hasKind,
  isAbleArray,
  isAbleFunction,
  isAbleStructInstance,
} from "./runtime";
import { BreakSignal, RaiseSignal, ReturnSignal } from "./signals";

export function evaluateFunctionCall(this: Interpreter, node: AST.FunctionCall, environment: Environment): AbleValue {
  const self = this as any;
  const callee = self.evaluate(node.callee, environment);
  if (!isAbleFunction(callee)) {
    throw new Error(`Interpreter Error: Cannot call non-function type ${callee?.kind ?? typeof callee}.`);
  }

  const func = callee as AbleFunction;
  const args = node.arguments.map((arg) => self.evaluate(arg, environment));

  if (func.isBoundMethod && typeof (func as any).apply === "function") {
    try {
      return (func as any).apply(args);
    } catch (e: any) {
      if (e instanceof ReturnSignal || e instanceof RaiseSignal || e instanceof BreakSignal) {
        throw e;
      }
      throw createError(e.message || "Bound method execution error", e);
    }
  }

  if (func.node === null && typeof (func as any).apply === "function") {
    try {
      return (func as any).apply(args);
    } catch (e: any) {
      throw createError(e.message || "Native function error", e);
    }
  }

  const funcDef = func.node;
  if (!funcDef) throw new Error("Interpreter Error: Function definition node is missing.");

  if (!func.isBoundMethod && args.length !== funcDef.params.length) {
    const funcName = funcDef.type === "FunctionDefinition" && funcDef.id ? funcDef.id.name : "(anonymous)";
    throw new Error(`Interpreter Error: Expected ${funcDef.params.length} arguments but got ${args.length} for function '${funcName}'.`);
  }

  const funcEnv = new Environment(func.closureEnv);
  for (let i = 0; i < funcDef.params.length; i++) {
    const param = funcDef.params[i];
    const argValue = args[i];
    if (!param) {
      throw new Error(`Interpreter Error: Parameter definition missing at index ${i}`);
    }
    if (argValue === undefined) {
      throw new Error(`Interpreter Error: Argument value undefined for parameter at index ${i}`);
    }

    if (param.name.type === "Identifier") {
      funcEnv.define(param.name.name, argValue);
    } else if (
      param.name.type === "StructPattern" ||
      param.name.type === "ArrayPattern" ||
      param.name.type === "WildcardPattern" ||
      param.name.type === "LiteralPattern"
    ) {
        self.evaluatePatternAssignment(param.name, argValue, funcEnv, true);
    } else {
      const unknownPattern: never = param.name;
      throw new Error(`Interpreter Error: Unsupported parameter pattern type: ${(unknownPattern as any).type}`);
    }
  }

  try {
    let lastValue: AbleValue;
    if (funcDef.body.type === "BlockExpression") {
        lastValue = self.evaluateStatements(funcDef.body.body, funcEnv);
      } else {
        lastValue = self.evaluate(funcDef.body, funcEnv);
      }
    return lastValue;
  } catch (signal) {
    if (signal instanceof ReturnSignal) {
      return signal.value;
    }
    throw signal;
  }
}

export function executeFunction(this: Interpreter, func: AbleFunction, args: AbleValue[], callSiteEnv: Environment): AbleValue {
  const self = this as any;
  const funcDef = func.node;
  if (!funcDef) throw new Error("Interpreter Error: Function definition node is missing.");

  if (args.length !== funcDef.params.length) {
    const funcName = funcDef.type === "FunctionDefinition" && funcDef.id ? funcDef.id.name : "(method/lambda)";
    throw new Error(
      `Interpreter Error: Argument count mismatch during direct function execution for '${funcName}'. Expected ${funcDef.params.length}, got ${args.length}.`,
    );
  }
  const funcEnv = new Environment(func.closureEnv);
  for (let i = 0; i < funcDef.params.length; i++) {
    const param = funcDef.params[i];
    const argValue = args[i];
    if (!param) {
      throw new Error(`Interpreter Error: Parameter definition missing at index ${i} in executeFunction.`);
    }
    if (argValue === undefined) {
      throw new Error(`Interpreter Error: Argument value undefined for parameter at index ${i} in executeFunction.`);
    }

    if (param.name.type === "Identifier") {
      funcEnv.define(param.name.name, argValue);
    } else if (
      param.name.type === "StructPattern" ||
      param.name.type === "ArrayPattern" ||
      param.name.type === "WildcardPattern" ||
      param.name.type === "LiteralPattern"
    ) {
        self.evaluatePatternAssignment(param.name, argValue, funcEnv, true);
    } else {
      const unknownPattern: never = param.name;
      throw new Error(`Interpreter Error: Unsupported parameter pattern type in executeFunction: ${(unknownPattern as any).type}`);
    }
  }
  try {
    let lastValue: AbleValue;
    if (funcDef.body.type === "BlockExpression") {
        lastValue = self.evaluateStatements(funcDef.body.body, funcEnv);
      } else {
        lastValue = self.evaluate(funcDef.body, funcEnv);
      }
    return lastValue;
  } catch (signal) {
    if (signal instanceof ReturnSignal) return signal.value;
    throw signal;
  }
}

export function findMethod(this: Interpreter, object: AbleValue, methodName: string): AbleFunction | null {
  const self = this as any;
  let typeName: string | null = null;
  if (isAbleStructInstance(object)) {
    typeName = object.definition.name;
  } else if (isAbleArray(object)) {
    typeName = "Array";
  } else if (hasKind(object, "string")) {
    typeName = "string";
  }

  if (!typeName) return null;

  const inherent = self.inherentMethods.get(typeName);
  if (inherent && inherent.methods.has(methodName)) {
    return inherent.methods.get(methodName)!;
  }

  const typeImplsMap = self.implementations.get(typeName);
  if (typeImplsMap) {
    for (const impl of typeImplsMap.values()) {
      if (impl.methods.has(methodName)) {
        return impl.methods.get(methodName)!;
      }
    }
  }

  return null;
}

export function bindMethod(this: Interpreter, selfValue: AbleValue, method: AbleFunction): AbleFunction {
  const self = this as any;
  const boundMethod: AbleFunction = {
    kind: "function",
    node: method.node,
    closureEnv: method.closureEnv,
    isBoundMethod: true,
    apply: (args: AbleValue[]) => {
      const finalArgs = [selfValue, ...args];
      const expectedParamCount = method.node?.params.length ?? 0;
      if (finalArgs.length !== expectedParamCount) {
        const funcName = method.node?.type === "FunctionDefinition" && method.node.id ? method.node.id.name : "(method)";
        console.error(
          `Internal Error: Bound method argument count mismatch for '${funcName}'. Expected ${expectedParamCount}, got ${finalArgs.length} (including self).`,
        );
        throw new Error(`Internal Interpreter Error: Bound method argument count mismatch for '${funcName}'.`);
      }
        return self.executeFunction(method, finalArgs, method.closureEnv);
      },
  } as any;
  return boundMethod;
}
