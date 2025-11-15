import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

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
  let typeArguments: AST.TypeExpression[] | undefined = node.typeArguments;
  let typeArgMap: Map<string, AST.TypeExpression> | undefined;
  if (generics && generics.length > 0) {
    typeArgMap = ctx.mapTypeArguments(generics, typeArguments, `instantiating ${structDef.id.name}`);
    if (constraints.length > 0) {
      ctx.enforceConstraintSpecs(constraints, typeArgMap, `struct ${structDef.id.name}`);
    }
  } else if (node.typeArguments && node.typeArguments.length > 0) {
    throw new Error(`Type '${structDef.id.name}' does not accept type arguments`);
  }
  if (node.isPositional) {
    const vals: V10Value[] = node.fields.map(f => ctx.evaluate(f.value, env));
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
    if (typeArguments && base.typeArguments) {
      if (typeArguments.length !== base.typeArguments.length) {
        throw new Error("Functional update must use same type arguments as source");
      }
      for (let i = 0; i < typeArguments.length; i++) {
        const expected = typeArguments[i]!;
        const actual = base.typeArguments[i]!;
        if (!ctx.typeExpressionsEqual(expected, actual)) {
          throw new Error("Functional update must use same type arguments as source");
        }
      }
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
  return {
    kind: "struct_instance",
    def: structDef,
    values: map,
    typeArguments,
    typeArgMap,
  };
}

export function memberAccessOnValue(ctx: InterpreterV10, obj: V10Value, member: AST.Identifier | AST.IntegerLiteral, env: Environment): V10Value {
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
      return {
        kind: "array",
        elements: args.map(a => ({ kind: "string", value: ctx.typeExpressionToString(a) } as V10Value)),
      };
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

  if (member.type === "Identifier" && obj.kind !== "struct_instance" && obj.kind !== "array") {
    const ufcs = ctx.tryUfcs(env, member.name, obj);
    if (ufcs) return ufcs;
    throw new Error("Member access only supported on structs/arrays in this milestone");
  }
  if (obj.kind === "struct_instance") {
    if (member.type === "Identifier") {
      if (!(obj.values instanceof Map)) throw new Error("Expected named struct instance");
      if (obj.values.has(member.name)) {
        return obj.values.get(member.name)!;
      }
      const typeName = obj.def.id.name;
      const method = ctx.findMethod(typeName, member.name, { typeArgs: obj.typeArguments, typeArgMap: obj.typeArgMap });
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
    const idx = Number(member.value);
    if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
    const el = obj.elements[idx];
    if (el === undefined) throw new Error("Internal error: array element undefined");
    return el;
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
