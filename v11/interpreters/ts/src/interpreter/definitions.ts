import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { ImplMethodEntry, V10Value } from "./values";

const NIL: V10Value = { kind: "nil", value: null };

function canonicalizeIdentifier(env: Environment, name: string): string {
  if (!env.has(name)) {
    return name;
  }
  const binding = env.get(name);
  if (binding.kind === "struct_def" || binding.kind === "interface_def") {
    return binding.def.id.name ?? name;
  }
  return name;
}

function canonicalizeTypeExpression(ctx: InterpreterV10, env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  const expanded = ctx.expandTypeAliases(expr);
  switch (expanded.type) {
    case "SimpleTypeExpression": {
      const name = canonicalizeIdentifier(env, expanded.name.name);
      if (name === expanded.name.name) {
        return expanded;
      }
      return AST.simpleTypeExpression(name);
    }
    case "GenericTypeExpression":
      return AST.genericTypeExpression(
        canonicalizeTypeExpression(ctx, env, expanded.base),
        (expanded.arguments ?? []).map((arg) => (arg ? canonicalizeTypeExpression(ctx, env, arg) : arg)),
      );
    case "NullableTypeExpression":
      return AST.nullableTypeExpression(canonicalizeTypeExpression(ctx, env, expanded.innerType));
    case "ResultTypeExpression":
      return AST.resultTypeExpression(canonicalizeTypeExpression(ctx, env, expanded.innerType));
    case "UnionTypeExpression":
      return AST.unionTypeExpression((expanded.members ?? []).map((member) => canonicalizeTypeExpression(ctx, env, member)));
    case "FunctionTypeExpression":
      return AST.functionTypeExpression(
        (expanded.paramTypes ?? []).map((param) => canonicalizeTypeExpression(ctx, env, param)),
        canonicalizeTypeExpression(ctx, env, expanded.returnType),
      );
    default:
      return expanded;
  }
}

export function evaluateInterfaceDefinition(ctx: InterpreterV10, node: AST.InterfaceDefinition, env: Environment): V10Value {
  ctx.interfaces.set(node.id.name, node);
  ctx.interfaceEnvs.set(node.id.name, env);
  env.define(node.id.name, { kind: "interface_def", def: node });
  ctx.registerSymbol(node.id.name, { kind: "interface_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "interface_def", def: node }); } catch {}
  }
  return NIL;
}

function insertFunction(
  bucket: Map<string, Extract<V10Value, { kind: "function" | "function_overload" }>>,
  name: string,
  fn: Extract<V10Value, { kind: "function" }>,
  priority = 0,
): void {
  (fn as any).methodResolutionPriority = priority;
  const existing = bucket.get(name);
  if (!existing) {
    bucket.set(name, fn);
    return;
  }
  if (existing.kind === "function") {
    bucket.set(name, { kind: "function_overload", overloads: [existing, fn] });
    return;
  }
  bucket.set(name, { kind: "function_overload", overloads: [...existing.overloads, fn] });
}

function functionExpectsSelf(def: AST.FunctionDefinition): boolean {
  if (def.isMethodShorthand) return true;
  const firstParam = Array.isArray(def.params) ? def.params[0] : undefined;
  if (!firstParam) return false;
  if (firstParam.name?.type === "Identifier" && firstParam.name.name?.toLowerCase() === "self") return true;
  if (firstParam.paramType?.type === "SimpleTypeExpression" && firstParam.paramType.name?.name === "Self") return true;
  return false;
}

export function evaluateUnionDefinition(ctx: InterpreterV10, node: AST.UnionDefinition, env: Environment): V10Value {
  env.define(node.id.name, { kind: "union_def", def: node });
  ctx.unions.set(node.id.name, node);
  ctx.registerSymbol(node.id.name, { kind: "union_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "union_def", def: node }); } catch {}
  }
  return NIL;
}

export function evaluateMethodsDefinition(ctx: InterpreterV10, node: AST.MethodsDefinition, env: Environment): V10Value {
  const targetType = canonicalizeTypeExpression(ctx, env, node.targetType);
  const typeName = (() => {
    let current: AST.TypeExpression = targetType;
    while (current.type === "GenericTypeExpression") {
      current = current.base;
    }
    if (current.type === "SimpleTypeExpression") {
      return canonicalizeIdentifier(env, current.name.name);
    }
    throw new Error("Only simple target types supported in methods");
  })();
  if (!ctx.inherentMethods.has(typeName)) ctx.inherentMethods.set(typeName, new Map());
  const bucket = ctx.inherentMethods.get(typeName)!;
  for (const def of node.definitions) {
    const expectsSelf = functionExpectsSelf(def);
    const exportedName = expectsSelf ? def.id.name : `${typeName}.${def.id.name}`;
    (def as any).structName = typeName;
    const fnValue: Extract<V10Value, { kind: "function" }> = { kind: "function", node: def, closureEnv: env };
    (fnValue as any).structName = typeName;
    if (!expectsSelf) {
      (fnValue as any).typeQualified = true;
    }
    insertFunction(bucket, def.id.name, fnValue, 0);
    env.define(exportedName, fnValue);
    ctx.registerSymbol(exportedName, fnValue);
    const qn = ctx.qualifiedName(exportedName);
    if (qn) {
      try { ctx.globals.define(qn, fnValue); } catch {}
    }
  }
  return NIL;
}

export function evaluateImplementationDefinition(ctx: InterpreterV10, node: AST.ImplementationDefinition, env: Environment): V10Value {
  const canonicalInterfaceName = canonicalizeIdentifier(env, node.interfaceName.name);
  const canonicalTarget = canonicalizeTypeExpression(ctx, env, node.targetType);
  const implNode: AST.ImplementationDefinition = {
    ...node,
    interfaceName: AST.identifier(canonicalInterfaceName),
    targetType: canonicalTarget,
  };
  const variants = ctx.expandImplementationTargetVariants(implNode.targetType);
  const unionVariantSignatures = implNode.targetType.type === "UnionTypeExpression"
    ? [...new Set(variants.map(v => v.signature))].sort()
    : undefined;
  const funcs = new Map<string, Extract<V10Value, { kind: "function" | "function_overload" }>>();
  for (const def of implNode.definitions) {
    insertFunction(funcs, def.id.name, { kind: "function", node: def, closureEnv: env }, -1);
  }
  ctx.attachDefaultInterfaceMethods(implNode, funcs);
  if (implNode.implName) {
    const name = implNode.implName.name;
    const symMap = new Map<string, V10Value>();
    for (const [k, v] of funcs.entries()) symMap.set(k, v);
    const implVal: V10Value = {
      kind: "impl_namespace",
      def: implNode,
      symbols: symMap,
      meta: { interfaceName: implNode.interfaceName.name, target: implNode.targetType, interfaceArgs: implNode.interfaceArgs },
    };
    env.define(name, implVal);
    ctx.registerSymbol(name, implVal);
    const qn = ctx.qualifiedName(name);
    if (qn) { try { ctx.globals.define(qn, implVal); } catch {} }
  } else {
    const constraintSpecs = ctx.collectConstraintSpecs(implNode.genericParams, implNode.whereClause);
    const baseConstraintSig = constraintSpecs
      .map(c => `${c.typeParam}->${ctx.typeExpressionToString(c.ifaceType)}`)
      .sort()
      .join("&") || "<none>";
    const genericNames = new Set((implNode.genericParams ?? []).map(gp => gp.name.name));
    for (const variant of variants) {
      const typeName = canonicalizeIdentifier(env, variant.typeName);
      const targetArgTemplates = variant.argTemplates;
      const isGenericTarget = genericNames.has(typeName);
      const key = `${implNode.interfaceName.name}::${typeName}`;
      if (!ctx.unnamedImplsSeen.has(key)) ctx.unnamedImplsSeen.set(key, new Map());
      const interfaceArgSig = (implNode.interfaceArgs ?? []).map(arg => ctx.typeExpressionToString(arg)).join("|") || "<none>";
      const templateArgSig = targetArgTemplates.length === 0
        ? "<none>"
        : targetArgTemplates.map(t => ctx.typeExpressionToString(t)).join("|");
      const templateKeyBase = `${interfaceArgSig}::${templateArgSig}`;
      const templateKey = unionVariantSignatures ? `${unionVariantSignatures.join("|")}::${templateKeyBase}` : templateKeyBase;
      const templateBucket = ctx.unnamedImplsSeen.get(key)!;
      if (!templateBucket.has(templateKey)) templateBucket.set(templateKey, new Set());
      const constraintKey = unionVariantSignatures ? `${unionVariantSignatures.join("|")}::${baseConstraintSig}` : baseConstraintSig;
      const constraintSet = templateBucket.get(templateKey)!;
      if (constraintSet.has(constraintKey)) {
        if (ctx.implDuplicateAllowlist.has(key)) continue;
        throw new Error(`Unnamed impl for (${implNode.interfaceName.name}, ${ctx.typeExpressionToString(implNode.targetType)}) already exists`);
      }
      constraintSet.add(constraintKey);
      const implEntry: ImplMethodEntry = {
        def: implNode,
        methods: funcs,
        targetArgTemplates,
        genericParams: implNode.genericParams ?? [],
        whereClause: implNode.whereClause,
        unionVariantSignatures,
      };
      if (isGenericTarget) {
        ctx.genericImplMethods.push(implEntry);
      } else {
        if (!ctx.implMethods.has(typeName)) ctx.implMethods.set(typeName, []);
        ctx.implMethods.get(typeName)!.push(implEntry);
        if (implNode.interfaceName.name === "Range") {
          ctx.registerRangeImplementation(implEntry, implNode.interfaceArgs);
        }
      }
    }
  }
  return NIL;
}
