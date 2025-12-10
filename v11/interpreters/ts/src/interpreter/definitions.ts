import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { ImplMethodEntry, V10Value } from "./values";

const NIL: V10Value = { kind: "nil", value: null };

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
  const typeName = (() => {
    let current: AST.TypeExpression = node.targetType;
    while (current.type === "GenericTypeExpression") {
      current = current.base;
    }
    if (current.type === "SimpleTypeExpression") {
      return current.name.name;
    }
    throw new Error("Only simple target types supported in methods");
  })();
  if (!ctx.inherentMethods.has(typeName)) ctx.inherentMethods.set(typeName, new Map());
  const bucket = ctx.inherentMethods.get(typeName)!;
  for (const def of node.definitions) {
    insertFunction(bucket, def.id.name, { kind: "function", node: def, closureEnv: env }, 0);
  }
  return NIL;
}

export function evaluateImplementationDefinition(ctx: InterpreterV10, node: AST.ImplementationDefinition, env: Environment): V10Value {
  const variants = ctx.expandImplementationTargetVariants(node.targetType);
  const unionVariantSignatures = node.targetType.type === "UnionTypeExpression"
    ? [...new Set(variants.map(v => v.signature))].sort()
    : undefined;
  const funcs = new Map<string, Extract<V10Value, { kind: "function" | "function_overload" }>>();
  for (const def of node.definitions) {
    insertFunction(funcs, def.id.name, { kind: "function", node: def, closureEnv: env }, -1);
  }
  ctx.attachDefaultInterfaceMethods(node, funcs);
  if (node.implName) {
    const name = node.implName.name;
    const symMap = new Map<string, V10Value>();
    for (const [k, v] of funcs.entries()) symMap.set(k, v);
    const implVal: V10Value = {
      kind: "impl_namespace",
      def: node,
      symbols: symMap,
      meta: { interfaceName: node.interfaceName.name, target: node.targetType, interfaceArgs: node.interfaceArgs },
    };
    env.define(name, implVal);
    ctx.registerSymbol(name, implVal);
    const qn = ctx.qualifiedName(name);
    if (qn) { try { ctx.globals.define(qn, implVal); } catch {} }
  } else {
    const constraintSpecs = ctx.collectConstraintSpecs(node.genericParams, node.whereClause);
    const baseConstraintSig = constraintSpecs
      .map(c => `${c.typeParam}->${ctx.typeExpressionToString(c.ifaceType)}`)
      .sort()
      .join("&") || "<none>";
    const genericNames = new Set((node.genericParams ?? []).map(gp => gp.name.name));
    for (const variant of variants) {
      const typeName = variant.typeName;
      const targetArgTemplates = variant.argTemplates;
      const isGenericTarget = genericNames.has(typeName);
      const key = `${node.interfaceName.name}::${typeName}`;
      if (!ctx.unnamedImplsSeen.has(key)) ctx.unnamedImplsSeen.set(key, new Map());
      const interfaceArgSig = (node.interfaceArgs ?? []).map(arg => ctx.typeExpressionToString(arg)).join("|") || "<none>";
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
        throw new Error(`Unnamed impl for (${node.interfaceName.name}, ${ctx.typeExpressionToString(node.targetType)}) already exists`);
      }
      constraintSet.add(constraintKey);
      const implEntry: ImplMethodEntry = {
        def: node,
        methods: funcs,
        targetArgTemplates,
        genericParams: node.genericParams ?? [],
        whereClause: node.whereClause,
        unionVariantSignatures,
      };
      if (isGenericTarget) {
        ctx.genericImplMethods.push(implEntry);
      } else {
        if (!ctx.implMethods.has(typeName)) ctx.implMethods.set(typeName, []);
        ctx.implMethods.get(typeName)!.push(implEntry);
        if (node.interfaceName.name === "Range") {
          ctx.registerRangeImplementation(implEntry, node.interfaceArgs);
        }
      }
    }
  }
  return NIL;
}
