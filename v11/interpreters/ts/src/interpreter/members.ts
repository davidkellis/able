import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

declare module "./index" {
  interface Interpreter {
    resolveMethodFromPool(
      env: Environment,
      funcName: string,
      receiver: RuntimeValue,
      opts?: { interfaceName?: string },
    ): Extract<RuntimeValue, { kind: "bound_method" }> | null;
  }
}

export function applyMemberAugmentations(cls: typeof Interpreter): void {
  cls.prototype.resolveMethodFromPool = function resolveMethodFromPool(
    this: Interpreter,
    env: Environment,
    funcName: string,
    receiver: RuntimeValue,
    opts?: { interfaceName?: string },
  ): Extract<RuntimeValue, { kind: "bound_method" }> | null {
    const typeName = this.getTypeNameForValue(receiver);
    let typeArgs = receiver.kind === "struct_instance" ? receiver.typeArguments : undefined;
    const typeArgMap = receiver.kind === "struct_instance" ? receiver.typeArgMap : undefined;
    const canonicalTypeName = typeName
      ? this.parseTypeExpression(this.expandTypeAliases(AST.simpleTypeExpression(typeName)))?.name
      : null;
    const candidateTypeNames = new Set<string>();
    if (typeName) candidateTypeNames.add(typeName);
    if (canonicalTypeName && canonicalTypeName !== typeName) candidateTypeNames.add(canonicalTypeName);
    if (!typeArgs && receiver.kind === "array") {
      const inferred = this.typeExpressionFromValue(receiver);
      const info = inferred ? this.parseTypeExpression(inferred) : null;
      if (info?.typeArgs && info.typeArgs.length > 0) {
        typeArgs = info.typeArgs;
      } else {
        typeArgs = [AST.wildcardTypeExpression()];
      }
    }
    const nameInScope = env?.has(funcName) ?? false;
    const typeNameInScope = Array.from(candidateTypeNames).some((name) => env?.has(name));
    const seen = new Set<Extract<RuntimeValue, { kind: "function" }>>();
    const candidates: Array<Extract<RuntimeValue, { kind: "function" }>> = [];
    const allowInherent = nameInScope || typeNameInScope || isPrimitiveReceiver(receiver, typeName);

    const addCandidate = (
      callable: Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null,
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

    for (const name of candidateTypeNames) {
      const bucket = this.inherentMethods.get(name);
      const inherent = allowInherent ? bucket?.get(funcName) : null;
      const instanceCallable = inherent ? selectInstanceCallable(inherent, receiver, this) : null;
      addCandidate(instanceCallable, name);
      const preExisting = candidates.length;
      let method: Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null = null;
      try {
        method = this.findMethod(name, funcName, {
          typeArgs,
          typeArgMap,
          interfaceName: opts?.interfaceName,
          includeInherent: false,
        });
      } catch (err) {
        if (!preExisting) throw err;
      }
      addCandidate(method, name);
    }

    const hasMethodCandidate = candidates.length > 0;
    try {
      const candidate = nameInScope ? env.get(funcName) : null;
      if (candidate && (candidate.kind === "function" || candidate.kind === "function_overload")) {
        const ufcs = selectUfcsCallable(candidate, receiver, this);
        if (!hasMethodCandidate) {
          addCandidate(ufcs);
        }
      }
    } catch {}

    if (!candidates.length) return null;
    const callable: Extract<RuntimeValue, { kind: "function" | "function_overload" }> =
      candidates.length === 1 ? candidates[0]! : { kind: "function_overload", overloads: candidates };
    return { kind: "bound_method", func: callable, self: receiver };
  };
}

function isPrimitiveReceiver(receiver: RuntimeValue, typeName?: string | null): boolean {
  switch (receiver.kind) {
    case "String":
    case "array":
    case "integer":
    case "float":
    case "bool":
    case "char":
    case "nil":
      return true;
    case "struct_instance":
      return typeName === "Array";
    default:
      return false;
  }
}

function selectInstanceCallable(
  func: Extract<RuntimeValue, { kind: "function" | "function_overload" }>,
  receiver?: RuntimeValue,
  ctx?: Interpreter,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
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
  func: Extract<RuntimeValue, { kind: "function" | "function_overload" }>,
  receiver: RuntimeValue,
  ctx: Interpreter,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  if (func.kind === "function") {
    if ((func as any).typeQualified) {
      return null;
    }
    return firstParamMatches(func.node, receiver, ctx) ? func : null;
  }
  const ufcsOverloads = func.overloads.filter(
    (entry) => entry?.node && !(entry as any).typeQualified && firstParamMatches(entry.node, receiver, ctx),
  );
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
  receiver: RuntimeValue | undefined,
  ctx: Interpreter | undefined,
): boolean {
  if (def.type !== "FunctionDefinition") return false;
  const structName = (def as any).structName ?? (def as any).struct_name;
  const receiverTypeName = receiver && ctx ? ctx.getTypeNameForValue(receiver) : null;
  if (structName && receiverTypeName && receiverTypeName !== structName) {
    return false;
  }
  if (def.isMethodShorthand) return true;
  const params = Array.isArray(def.params) ? def.params : [];
  if (!params.length) return false;
  if (!receiver || !ctx?.matchesType) return true;
  const firstType = params[0]?.paramType ?? AST.wildcardTypeExpression();
  return ctx.matchesType(firstType, receiver);
}
