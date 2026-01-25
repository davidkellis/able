import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { ImplMethodEntry, RuntimeValue } from "./values";
import { IMPL_NAMESPACE_BINDING } from "./impl_namespace";

const NIL: RuntimeValue = { kind: "nil", value: null };

function canonicalizeIdentifier(env: Environment, name: string): string {
  if (!env.has(name)) {
    return name;
  }
  const binding = env.get(name);
  if (binding.kind === "struct_def" || binding.kind === "interface_def" || binding.kind === "union_def") {
    return binding.def.id.name ?? name;
  }
  return name;
}

function canonicalizeExpandedTypeExpression(env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = canonicalizeIdentifier(env, expr.name.name);
      if (name === expr.name.name) {
        return expr;
      }
      return AST.simpleTypeExpression(name);
    }
    case "GenericTypeExpression":
      return AST.genericTypeExpression(
        canonicalizeExpandedTypeExpression(env, expr.base),
        (expr.arguments ?? []).map((arg) => (arg ? canonicalizeExpandedTypeExpression(env, arg) : arg)),
      );
    case "NullableTypeExpression":
      return AST.nullableTypeExpression(canonicalizeExpandedTypeExpression(env, expr.innerType));
    case "ResultTypeExpression":
      return AST.resultTypeExpression(canonicalizeExpandedTypeExpression(env, expr.innerType));
    case "UnionTypeExpression":
      return AST.unionTypeExpression((expr.members ?? []).map((member) => canonicalizeExpandedTypeExpression(env, member)));
    case "FunctionTypeExpression":
      return AST.functionTypeExpression(
        (expr.paramTypes ?? []).map((param) => canonicalizeExpandedTypeExpression(env, param)),
        canonicalizeExpandedTypeExpression(env, expr.returnType),
      );
    default:
      return expr;
  }
}

export function canonicalizeTypeExpression(ctx: Interpreter, env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  const expanded = ctx.expandTypeAliases(expr);
  return canonicalizeExpandedTypeExpression(env, expanded);
}

export function evaluateInterfaceDefinition(ctx: Interpreter, node: AST.InterfaceDefinition, env: Environment): RuntimeValue {
  ctx.interfaces.set(node.id.name, node);
  ctx.interfaceEnvs.set(node.id.name, env);
  ctx.defineInEnv(env, node.id.name, { kind: "interface_def", def: node });
  ctx.registerSymbol(node.id.name, { kind: "interface_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "interface_def", def: node }); } catch {}
  }
  return NIL;
}

function insertFunction(
  bucket: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>,
  name: string,
  fn: Extract<RuntimeValue, { kind: "function" }>,
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

export function evaluateUnionDefinition(ctx: Interpreter, node: AST.UnionDefinition, env: Environment): RuntimeValue {
  ctx.defineInEnv(env, node.id.name, { kind: "union_def", def: node });
  ctx.unions.set(node.id.name, node);
  ctx.registerSymbol(node.id.name, { kind: "union_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "union_def", def: node }); } catch {}
  }
  return NIL;
}

export function evaluateMethodsDefinition(ctx: Interpreter, node: AST.MethodsDefinition, env: Environment): RuntimeValue {
  const targetType = canonicalizeTypeExpression(ctx, env, node.targetType);
  const methodSetConstraints = ctx.collectConstraintSpecs(node.genericParams, node.whereClause);
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
    const fnValue: Extract<RuntimeValue, { kind: "function" }> = { kind: "function", node: def, closureEnv: env };
    (fnValue as any).methodSetConstraints = methodSetConstraints;
    (fnValue as any).methodSetTargetType = targetType;
    (fnValue as any).methodSetGenericParams = node.genericParams;
    (fnValue as any).structName = typeName;
    if (!expectsSelf) {
      (fnValue as any).typeQualified = true;
    }
    insertFunction(bucket, def.id.name, fnValue, 0);
    ctx.defineInEnv(env, exportedName, fnValue);
    ctx.registerSymbol(exportedName, fnValue);
    const qn = ctx.qualifiedName(exportedName);
    if (qn) {
      try { ctx.globals.define(qn, fnValue); } catch {}
    }
  }
  return NIL;
}

export function evaluateImplementationDefinition(ctx: Interpreter, node: AST.ImplementationDefinition, env: Environment): RuntimeValue {
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
  const funcs = new Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>();
  for (const def of implNode.definitions) {
    insertFunction(funcs, def.id.name, { kind: "function", node: def, closureEnv: env }, -1);
  }
  ctx.attachDefaultInterfaceMethods(implNode, funcs);
  if (implNode.implName) {
    const name = implNode.implName.name;
    const symMap = new Map<string, RuntimeValue>();
    for (const [k, v] of funcs.entries()) symMap.set(k, v);
    const implVal: RuntimeValue = {
      kind: "impl_namespace",
      def: implNode,
      symbols: symMap,
      meta: { interfaceName: implNode.interfaceName.name, target: implNode.targetType, interfaceArgs: implNode.interfaceArgs },
    };
    const wrapImplMethodEnv = (callable: Extract<RuntimeValue, { kind: "function" | "function_overload" }>): void => {
      const wrap = (fn: Extract<RuntimeValue, { kind: "function" }>): void => {
        const wrapped = new Environment(fn.closureEnv);
        wrapped.define(IMPL_NAMESPACE_BINDING, implVal);
        fn.closureEnv = wrapped;
      };
      if (callable.kind === "function") {
        wrap(callable);
        return;
      }
      for (const fn of callable.overloads) {
        if (!fn) continue;
        wrap(fn);
      }
    };
    for (const method of funcs.values()) {
      if (!method) continue;
      wrapImplMethodEnv(method);
    }
    ctx.defineInEnv(env, name, implVal);
    ctx.registerSymbol(name, implVal);
    const qn = ctx.qualifiedName(name);
    if (qn) { try { ctx.globals.define(qn, implVal); } catch {} }
  } else {
    const constraintSpecs = ctx.collectConstraintSpecs(implNode.genericParams, implNode.whereClause);
    const baseConstraintSig = constraintSpecs
      .map(c => `${ctx.typeExpressionToString(c.subjectExpr)}->${ctx.typeExpressionToString(c.ifaceType)}`)
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
