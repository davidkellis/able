import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    tryUfcs(env: Environment, funcName: string, receiver: V10Value): Extract<V10Value, { kind: "bound_method" }> | null;
  }
}

export function applyMemberAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.tryUfcs = function tryUfcs(this: InterpreterV10, env: Environment, funcName: string, receiver: V10Value): Extract<V10Value, { kind: "bound_method" }> | null {
    try {
      const candidate = env.get(funcName);
      if (candidate && (candidate.kind === "function" || candidate.kind === "function_overload")) {
        return { kind: "bound_method", func: candidate, self: receiver };
      }
    } catch {}
    const typeName = this.getTypeNameForValue(receiver);
    if (!typeName) return null;
    const bucket = this.inherentMethods.get(typeName);
    if (!bucket) return null;
    const method = bucket.get(funcName);
    if (!method) return null;
    const instanceCallable = selectInstanceCallable(method);
    if (!instanceCallable) return null;
    return { kind: "bound_method", func: instanceCallable, self: receiver };
  };
}

function selectInstanceCallable(func: Extract<V10Value, { kind: "function" | "function_overload" }>): Extract<V10Value, { kind: "function" | "function_overload" }> | null {
  if (func.kind === "function") {
    return functionExpectsSelf(func.node) ? func : null;
  }
  const instanceOverloads = func.overloads.filter((entry) => entry?.node && functionExpectsSelf(entry.node));
  if (!instanceOverloads.length) {
    return null;
  }
  if (instanceOverloads.length === 1) {
    return instanceOverloads[0] ?? null;
  }
  return { kind: "function_overload", overloads: instanceOverloads };
}

function functionExpectsSelf(def: AST.FunctionDefinition | AST.LambdaExpression): boolean {
  if (def.type !== "FunctionDefinition") return false;
  if (def.isMethodShorthand) return true;
  const firstParam = Array.isArray(def.params) ? def.params[0] : undefined;
  if (!firstParam) return false;
  if (firstParam.name?.type === "Identifier" && firstParam.name.name?.toLowerCase() === "self") return true;
  if (firstParam.paramType?.type === "SimpleTypeExpression" && firstParam.paramType.name?.name === "Self") return true;
  return false;
}
