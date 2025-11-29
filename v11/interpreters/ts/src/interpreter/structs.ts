import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { makeIntegerFromNumber, numericToNumber } from "./numeric";

export function evaluateStructDefinition(ctx: InterpreterV10, node: AST.StructDefinition, env: Environment): V10Value {
  env.define(node.id.name, { kind: "struct_def", def: node });
  ctx.registerSymbol(node.id.name, { kind: "struct_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "struct_def", def: node }); } catch {}
  }
  return { kind: "nil", value: null };
}

export function evaluateStructLiteral(ctx: InterpreterV10, node: AST.StructLiteral, env: Environment): V10Value {
  if (!node.structType) throw new Error("Struct literal requires explicit struct type");
  const defVal = env.get(node.structType.name);
  if (defVal.kind !== "struct_def") throw new Error(`'${node.structType.name}' is not a struct type`);
  const structDef = defVal.def;
  const generics = structDef.genericParams;
  const constraints = ctx.collectConstraintSpecs(generics, structDef.whereClause);
  if (node.isPositional) {
    const vals: V10Value[] = node.fields.map(f => ctx.evaluate(f.value, env));
    let typeArguments: AST.TypeExpression[] | undefined = node.typeArguments;
    let typeArgMap: Map<string, AST.TypeExpression> | undefined;
    if (generics && generics.length > 0) {
      if (typeArguments && typeArguments.length > 0 && typeArguments.length !== generics.length) {
        throw new Error(`Type '${structDef.id.name}' expects ${generics.length} type arguments, got ${typeArguments.length}`);
      }
      if (!typeArguments || typeArguments.length === 0) {
        typeArguments = inferStructTypeArguments(ctx, structDef, vals);
      }
      if (!typeArguments || typeArguments.length === 0) {
        typeArguments = generics.map(() => AST.wildcardTypeExpression());
      }
      typeArgMap = ctx.mapTypeArguments(generics, typeArguments, `instantiating ${structDef.id.name}`);
      if (constraints.length > 0) {
        ctx.enforceConstraintSpecs(constraints, typeArgMap, `struct ${structDef.id.name}`);
      }
    } else if (node.typeArguments && node.typeArguments.length > 0) {
      throw new Error(`Type '${structDef.id.name}' does not accept type arguments`);
    }
    return {
      kind: "struct_instance",
      def: structDef,
      values: vals,
      typeArguments,
      typeArgMap,
    };
  }

  const map = new Map<string, V10Value>();
  const legacySource = (node as any).functionalUpdateSource as AST.Expression | undefined;
  const updateSources = node.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
  let canonicalBaseTypeArgs: AST.TypeExpression[] | undefined;
  for (const src of updateSources) {
    const base = ctx.evaluate(src, env);
    if (base.kind !== "struct_instance") throw new Error("Functional update source must be a struct instance");
    if (base.def.id.name !== structDef.id.name) throw new Error("Functional update source must be same struct type");
    if (!(base.values instanceof Map)) throw new Error("Functional update only supported for named structs");
    if (canonicalBaseTypeArgs && base.typeArguments) {
      if (canonicalBaseTypeArgs.length !== base.typeArguments.length) {
        throw new Error("Functional update sources must share type arguments");
      }
      for (let i = 0; i < canonicalBaseTypeArgs.length; i++) {
        const expected = canonicalBaseTypeArgs[i]!;
        const actual = base.typeArguments[i]!;
        if (!ctx.typeExpressionsEqual(expected, actual)) {
          throw new Error("Functional update sources must share type arguments");
        }
      }
    } else if (!canonicalBaseTypeArgs && base.typeArguments) {
      canonicalBaseTypeArgs = base.typeArguments;
    } else if (canonicalBaseTypeArgs && !base.typeArguments && canonicalBaseTypeArgs.length > 0) {
      throw new Error("Functional update sources must share type arguments");
    }
    for (const [key, value] of base.values.entries()) {
      map.set(key, value);
    }
  }
  for (const field of node.fields) {
    let fieldName = field.name?.name;
    if (!fieldName && field.isShorthand && field.value.type === "Identifier") {
      fieldName = (field.value as AST.Identifier).name;
    }
    if (!fieldName) throw new Error("Named struct field initializer must have a field name");
    const val = ctx.evaluate(field.value, env);
    map.set(fieldName, val);
  }

  let typeArguments: AST.TypeExpression[] | undefined = node.typeArguments;
  let typeArgMap: Map<string, AST.TypeExpression> | undefined;
  if (generics && generics.length > 0) {
    if (typeArguments && typeArguments.length > 0 && typeArguments.length !== generics.length) {
      throw new Error(`Type '${structDef.id.name}' expects ${generics.length} type arguments, got ${typeArguments.length}`);
    }
    if (!typeArguments || typeArguments.length === 0) {
      if (canonicalBaseTypeArgs) {
        typeArguments = canonicalBaseTypeArgs;
      } else {
        typeArguments = inferStructTypeArguments(ctx, structDef, map);
      }
    }
    if (canonicalBaseTypeArgs && canonicalBaseTypeArgs.length > 0 && typeArguments) {
      if (canonicalBaseTypeArgs.length !== typeArguments.length) {
        throw new Error("Functional update must use same type arguments as source");
      }
      for (let i = 0; i < canonicalBaseTypeArgs.length; i++) {
        if (!ctx.typeExpressionsEqual(canonicalBaseTypeArgs[i]!, typeArguments[i]!)) {
          throw new Error("Functional update must use same type arguments as source");
        }
      }
    }
    if (!typeArguments || typeArguments.length === 0) {
      typeArguments = generics.map(() => AST.wildcardTypeExpression());
    }
    typeArgMap = ctx.mapTypeArguments(generics, typeArguments, `instantiating ${structDef.id.name}`);
    if (constraints.length > 0) {
      ctx.enforceConstraintSpecs(constraints, typeArgMap, `struct ${structDef.id.name}`);
    }
  } else if (node.typeArguments && node.typeArguments.length > 0) {
    throw new Error(`Type '${structDef.id.name}' does not accept type arguments`);
  }

  return {
    kind: "struct_instance",
    def: structDef,
    values: map,
    typeArguments,
    typeArgMap,
  };
}

function inferStructTypeArguments(ctx: InterpreterV10, def: AST.StructDefinition, values: V10Value[] | Map<string, V10Value>): AST.TypeExpression[] {
  const generics = def.genericParams ?? [];
  if (generics.length === 0) return [];
  const bindings = new Map<string, AST.TypeExpression>();
  const genericNames = new Set(generics.map(g => g.name.name));
  if (Array.isArray(values)) {
    for (let i = 0; i < def.fields.length && i < values.length; i++) {
      const field = def.fields[i];
      if (!field?.fieldType) continue;
      const actual = ctx.typeExpressionFromValue(values[i]!);
      if (!actual) continue;
      ctx.matchTypeExpressionTemplate(field.fieldType, actual, genericNames, bindings);
    }
  } else {
    for (const field of def.fields) {
      if (!field?.name || !field.fieldType) continue;
      if (!values.has(field.name.name)) continue;
      const actual = ctx.typeExpressionFromValue(values.get(field.name.name)!);
      if (!actual) continue;
      ctx.matchTypeExpressionTemplate(field.fieldType, actual, genericNames, bindings);
    }
  }
  return generics.map(gp => bindings.get(gp.name.name) ?? AST.wildcardTypeExpression());
}

export function memberAccessOnValue(
  ctx: InterpreterV10,
  obj: V10Value,
  member: AST.Identifier | AST.IntegerLiteral,
  env: Environment,
  opts?: { preferMethods?: boolean },
): V10Value {
  if (obj.kind === "struct_def" && member.type === "Identifier") {
    const typeName = obj.def.id.name;
    const method = ctx.findMethod(typeName, member.name);
    if (!method) throw new Error(`No static method '${member.name}' for ${typeName}`);
    if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
      throw new Error(`Method '${member.name}' on ${typeName} is private`);
    }
    return method;
  }
  if (obj.kind === "package") {
    if (member.type !== "Identifier") throw new Error("Package member access expects identifier");
    const sym = obj.symbols.get(member.name);
    if (!sym) throw new Error(`No public member '${member.name}' on package ${obj.name}`);
    return sym;
  }
  if (obj.kind === "dyn_package") {
    if (member.type !== "Identifier") throw new Error("Dyn package member access expects identifier");
    const bucket = ctx.packageRegistry.get(obj.name);
    const sym = bucket?.get(member.name);
    if (!sym) throw new Error(`dyn package '${obj.name}' has no member '${member.name}'`);
    if (sym.kind === "function" && sym.node.type === "FunctionDefinition" && sym.node.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "struct_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "interface_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "union_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    return { kind: "dyn_ref", pkg: obj.name, name: member.name };
  }
  if (obj.kind === "impl_namespace") {
    if (member.type !== "Identifier") throw new Error("Impl namespace member access expects identifier");
    if (member.name === "interface") {
      return { kind: "string", value: obj.meta.interfaceName };
    }
    if (member.name === "target") {
      return { kind: "string", value: ctx.typeExpressionToString(obj.meta.target) };
    }
    if (member.name === "interface_args") {
      const args = obj.meta.interfaceArgs ?? [];
      return ctx.makeArrayValue(args.map(a => ({ kind: "string", value: ctx.typeExpressionToString(a) } as V10Value)));
    }
    const sym = obj.symbols.get(member.name);
    if (!sym) throw new Error(`No method '${member.name}' on impl ${obj.def.implName?.name ?? "<unnamed>"}`);
    return sym;
  }
  if (obj.kind === "interface_value") {
    if (member.type !== "Identifier") throw new Error("Interface member access expects identifier");
    const underlying = obj.value;
    const typeName = ctx.getTypeNameForValue(underlying);
    if (!typeName) throw new Error(`No method '${member.name}' for interface ${obj.interfaceName}`);
    const typeArgs = underlying.kind === "struct_instance" ? underlying.typeArguments : undefined;
    const typeArgMap = underlying.kind === "struct_instance" ? underlying.typeArgMap : undefined;
    const method = ctx.findMethod(typeName, member.name, {
      typeArgs,
      typeArgMap,
      interfaceName: obj.interfaceName,
    });
    if (!method) throw new Error(`No method '${member.name}' for interface ${obj.interfaceName}`);
    if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
      throw new Error(`Method '${member.name}' on ${typeName} is private`);
    }
    return { kind: "bound_method", func: method, self: underlying };
  }
  if (obj.kind === "proc_handle") {
    if (member.type !== "Identifier") throw new Error("Proc handle member access expects identifier");
    const fn = (ctx.procNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown proc handle method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "future") {
    if (member.type !== "Identifier") throw new Error("Future member access expects identifier");
    const fn = (ctx.futureNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown future method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "iterator") {
    if (member.type !== "Identifier") throw new Error("Iterator member access expects identifier");
    ctx.ensureIteratorBuiltins();
    const fn = (ctx.iteratorNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown iterator method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "error") {
    if (member.type !== "Identifier") throw new Error("Error member access expects identifier");
    if (member.name === "value") {
      return obj.value ?? { kind: "nil", value: null };
    }
    const fn = (ctx.errorNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[member.name];
    if (fn) {
      return ctx.bindNativeMethod(fn, obj);
    }
    const ufcs = ctx.tryUfcs(env, member.name, obj);
    if (ufcs) return ufcs;
    throw new Error(`No field or method named '${member.name}' on error value`);
  }

  if (member.type === "Identifier" && obj.kind !== "struct_instance" && obj.kind !== "array" && obj.kind !== "string") {
    const ufcs = ctx.tryUfcs(env, member.name, obj);
    if (ufcs) return ufcs;
    const typeName = ctx.getTypeNameForValue(obj);
    if (typeName) {
      const method = ctx.findMethod(typeName, member.name);
      if (method) {
        if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
          throw new Error(`Method '${member.name}' on ${typeName} is private`);
        }
        return { kind: "bound_method", func: method, self: obj };
      }
    }
    throw new Error(`Member access only supported on structs/arrays in this milestone (got ${obj.kind})`);
  }
  if (obj.kind === "struct_instance") {
    if (member.type === "Identifier") {
      if (!(obj.values instanceof Map)) throw new Error("Expected named struct instance");
      const typeName = obj.def.id.name;
      const method = ctx.findMethod(typeName, member.name, { typeArgs: obj.typeArguments, typeArgMap: obj.typeArgMap });
      if (opts?.preferMethods && method) {
        if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
          throw new Error(`Method '${member.name}' on ${typeName} is private`);
        }
        return { kind: "bound_method", func: method, self: obj };
      }
      if (obj.values.has(member.name)) {
        return obj.values.get(member.name)!;
      }
      if (method) {
        if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
          throw new Error(`Method '${member.name}' on ${typeName} is private`);
        }
        return { kind: "bound_method", func: method, self: obj };
      }
      const ufcs = ctx.tryUfcs(env, member.name, obj);
      if (ufcs) return ufcs;
      throw new Error(`No field or method named '${member.name}'`);
    }
    const idx = Number(member.value);
    if (!Array.isArray(obj.values)) throw new Error("Expected positional struct instance");
    if (idx < 0 || idx >= obj.values.length) throw new Error("Struct field index out of bounds");
    const val = obj.values[idx];
    if (val === undefined) throw new Error("Internal error: positional field is undefined");
    return val;
  }
  if (obj.kind === "array") {
    ctx.ensureArrayState(obj);
    if (member.type === "Identifier") {
      const name = member.name;
      if (name === "storage_handle") {
        const handle = obj.handle ?? 0;
        return makeIntegerFromNumber("i64", handle);
      }
      if (name === "length") {
        const state = ctx.ensureArrayState(obj);
        return makeIntegerFromNumber("i32", state.values.length);
      }
      if (name === "capacity") {
        const state = ctx.ensureArrayState(obj);
        return makeIntegerFromNumber("i32", state.capacity);
      }
      const ufcsFirst = ctx.tryUfcs(env, name, obj);
      if (ufcsFirst) return ufcsFirst;
      const method = ctx.findMethod("Array", name, { typeArgs: [AST.wildcardTypeExpression()] });
      if (method) {
        return { kind: "bound_method", func: method, self: obj };
      }
      const ufcs = ctx.tryUfcs(env, name, obj);
      if (ufcs) return ufcs;
      throw new Error(`Array has no member '${name}' (import able.collections.array for stdlib helpers)`);
    }
    if (member.type === "IntegerLiteral") {
      const idx = Number(member.value);
      const state = ctx.ensureArrayState(obj);
      if (idx < 0 || idx >= state.values.length) throw new Error("Array index out of bounds");
      const el = state.values[idx];
      if (el === undefined) throw new Error("Internal error: array element undefined");
      return el;
    }
    throw new Error("Array member access expects identifier or positional index");
  }
  if (obj.kind === "string") {
    if (member.type !== "Identifier") {
      throw new Error("String member access expects identifier");
    }
    const name = member.name;
    const ufcsFirst = ctx.tryUfcs(env, name, obj);
    if (ufcsFirst) return ufcsFirst;
    const method = ctx.findMethod("string", name);
    if (method) {
      return { kind: "bound_method", func: method, self: obj };
    }
    throw new Error(`String has no member '${name}' (import able.text.string for stdlib helpers)`);
  }
  if (member.type === "Identifier") {
    const ufcs = ctx.tryUfcs(env, member.name, obj);
    if (ufcs) return ufcs;
  }
  throw new Error("Member access only supported on structs/arrays in this milestone");
}

export function evaluateMemberAccessExpression(ctx: InterpreterV10, node: AST.MemberAccessExpression, env: Environment): V10Value {
  const obj = ctx.evaluate(node.object, env);
  if (node.isSafe && obj.kind === "nil") {
    return obj;
  }
  return memberAccessOnValue(ctx, obj, node.member, env);
}

export function evaluateImplicitMemberExpression(ctx: InterpreterV10, node: AST.ImplicitMemberExpression, env: Environment): V10Value {
  if (node.member.type !== "Identifier") throw new Error("Implicit member expects identifier");
  if (ctx.implicitReceiverStack.length === 0) {
    throw new Error(`Implicit member '#${node.member.name}' used outside of function with implicit receiver`);
  }
  const receiver = ctx.implicitReceiverStack[ctx.implicitReceiverStack.length - 1];
  if (receiver === undefined) {
    throw new Error(`Implicit member '#${node.member.name}' used outside of function with implicit receiver`);
  }
  return memberAccessOnValue(ctx, receiver, node.member, env);
}

function toSafeIndex(value: V10Value | undefined, label: string): number {
  const raw = numericToNumber(value ?? { kind: "nil", value: null }, label, { requireSafeInteger: true });
  const idx = Math.trunc(raw);
  if (idx < 0) throw new Error(`${label} must be non-negative`);
  return idx;
}
