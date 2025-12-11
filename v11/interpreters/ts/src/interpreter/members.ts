import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    resolveMethodFromPool(
      env: Environment,
      funcName: string,
      receiver: V10Value,
      opts?: { interfaceName?: string },
    ): Extract<V10Value, { kind: "bound_method" }> | null;
  }
}

export function applyMemberAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.resolveMethodFromPool = function resolveMethodFromPool(
    this: InterpreterV10,
    env: Environment,
    funcName: string,
    receiver: V10Value,
    opts?: { interfaceName?: string },
  ): Extract<V10Value, { kind: "bound_method" }> | null {
    const typeName = this.getTypeNameForValue(receiver);
    let typeArgs = receiver.kind === "struct_instance" ? receiver.typeArguments : undefined;
    const typeArgMap = receiver.kind === "struct_instance" ? receiver.typeArgMap : undefined;
    if (!typeArgs) {
      if (receiver.kind === "array") {
        typeArgs = [AST.wildcardTypeExpression()];
      }
    }
    const seen = new Set<Extract<V10Value, { kind: "function" }>>();
    const candidates: Array<Extract<V10Value, { kind: "function" }>> = [];

    const addCandidate = (
      callable: Extract<V10Value, { kind: "function" | "function_overload" }> | null,
      privacyContext?: string,
    ): void => {
      if (!callable) return;
      const functions = callable.kind === "function_overload" ? callable.overloads : [callable];
      for (const fn of functions) {
        if (!fn) continue;
        if (fn.node?.type === "FunctionDefinition" && fn.node.isPrivate) {
          if (privacyContext) {
            throw new Error(`Method '${funcName}' on ${privacyContext} is private`);
          }
          continue;
        }
        if (seen.has(fn)) {
          continue;
        }
        seen.add(fn);
        candidates.push(fn);
      }
    };

    if (typeName) {
      const bucket = this.inherentMethods.get(typeName);
      const inherent = bucket?.get(funcName);
      const instanceCallable = inherent ? selectInstanceCallable(inherent, receiver, this) : null;
      addCandidate(instanceCallable, typeName);
      const preExisting = candidates.length;
      let method: Extract<V10Value, { kind: "function" | "function_overload" }> | null = null;
      try {
        method = this.findMethod(typeName, funcName, {
          typeArgs,
          typeArgMap,
          interfaceName: opts?.interfaceName,
          includeInherent: false,
        });
      } catch (err) {
        if (!preExisting) throw err;
      }
      addCandidate(method, typeName);
    }

    const hasMethodCandidate = candidates.length > 0;
    try {
      const candidate = env.get(funcName);
      if (candidate && (candidate.kind === "function" || candidate.kind === "function_overload")) {
        const ufcs = selectUfcsCallable(candidate, receiver, this);
        if (!hasMethodCandidate) {
          addCandidate(ufcs);
        }
      }
    } catch {}

    if (!candidates.length) return null;
    const callable: Extract<V10Value, { kind: "function" | "function_overload" }> =
      candidates.length === 1 ? candidates[0]! : { kind: "function_overload", overloads: candidates };
    return { kind: "bound_method", func: callable, self: receiver };
  };
}

function selectInstanceCallable(
  func: Extract<V10Value, { kind: "function" | "function_overload" }>,
  receiver?: V10Value,
  ctx?: InterpreterV10,
): Extract<V10Value, { kind: "function" | "function_overload" }> | null {
  if (func.kind === "function") {
    return functionExpectsSelf(func.node) && firstParamMatches(func.node, receiver, ctx) ? func : null;
  }
  const instanceOverloads = func.overloads.filter(
    (entry) => entry?.node && functionExpectsSelf(entry.node) && firstParamMatches(entry.node, receiver, ctx),
  );
  if (!instanceOverloads.length) {
    return null;
  }
  if (instanceOverloads.length === 1) {
    return instanceOverloads[0] ?? null;
  }
  return { kind: "function_overload", overloads: instanceOverloads };
}

function selectUfcsCallable(
  func: Extract<V10Value, { kind: "function" | "function_overload" }>,
  receiver: V10Value,
  ctx: InterpreterV10,
): Extract<V10Value, { kind: "function" | "function_overload" }> | null {
  if (func.kind === "function") {
    return firstParamMatches(func.node, receiver, ctx) ? func : null;
  }
  const ufcsOverloads = func.overloads.filter((entry) => entry?.node && firstParamMatches(entry.node, receiver, ctx));
  if (!ufcsOverloads.length) {
    return null;
  }
  if (ufcsOverloads.length === 1) {
    return ufcsOverloads[0] ?? null;
  }
  return { kind: "function_overload", overloads: ufcsOverloads };
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

function firstParamMatches(
  def: AST.FunctionDefinition | AST.LambdaExpression,
  receiver: V10Value | undefined,
  ctx: InterpreterV10 | undefined,
): boolean {
  if (def.type !== "FunctionDefinition") return false;
  if (def.isMethodShorthand) return true;
  const params = Array.isArray(def.params) ? def.params : [];
  if (!params.length) return false;
  if (!receiver || !ctx?.matchesType) return true;
  const firstType = params[0]?.paramType ?? AST.wildcardTypeExpression();
  return ctx.matchesType(firstType, receiver);
}
