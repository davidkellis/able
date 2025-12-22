import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerFromNumber, numericToNumber } from "./numeric";

function isCallable(value: RuntimeValue): boolean {
  switch (value.kind) {
    case "function":
    case "function_overload":
    case "native_function":
    case "native_bound_method":
    case "bound_method":
    case "partial_function":
    case "dyn_ref":
      return true;
    default:
      return false;
  }
}

export function evaluateStructDefinition(ctx: Interpreter, node: AST.StructDefinition, env: Environment): RuntimeValue {
  env.define(node.id.name, { kind: "struct_def", def: node });
  ctx.registerSymbol(node.id.name, { kind: "struct_def", def: node });
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, { kind: "struct_def", def: node }); } catch {}
  }
  return { kind: "nil", value: null };
}

export function evaluateStructLiteral(ctx: Interpreter, node: AST.StructLiteral, env: Environment): RuntimeValue {
  if (!node.structType) throw new Error("Struct literal requires explicit struct type");
  const defVal = env.get(node.structType.name);
  if (defVal.kind !== "struct_def") throw new Error(`'${node.structType.name}' is not a struct type`);
  const structDef = defVal.def;
  const generics = structDef.genericParams;
  const constraints = ctx.collectConstraintSpecs(generics, structDef.whereClause);
  if (node.isPositional) {
    const vals: RuntimeValue[] = node.fields.map(f => ctx.evaluate(f.value, env));
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

  const map = new Map<string, RuntimeValue>();
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

function inferStructTypeArguments(ctx: Interpreter, def: AST.StructDefinition, values: RuntimeValue[] | Map<string, RuntimeValue>): AST.TypeExpression[] {
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
  ctx: Interpreter,
  obj: RuntimeValue,
  member: AST.Identifier | AST.IntegerLiteral,
  env: Environment,
  opts?: { preferMethods?: boolean },
): RuntimeValue {
  if (obj.kind === "struct_def" && member.type === "Identifier") {
    const typeName = obj.def.id.name;
    const method = ctx.findMethod(typeName, member.name);
    if (!method) throw new Error(`No static method '${member.name}' for ${typeName}`);
    const methodNode = method.kind === "function_overload" ? method.overloads[0]?.node : method.node;
    if (methodNode?.type === "FunctionDefinition" && methodNode.isPrivate) {
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
    if (sym.kind === "function_overload") {
      const first = sym.overloads[0];
      if (first?.node.type === "FunctionDefinition" && first.node.isPrivate) {
        throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
      }
    }
    if (sym.kind === "function" && sym.node.type === "FunctionDefinition" && sym.node.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "struct_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "interface_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    if (sym.kind === "union_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${member.name}' is private`);
    return { kind: "dyn_ref", pkg: obj.name, name: member.name };
  }
  if (obj.kind === "impl_namespace") {
    if (member.type !== "Identifier") throw new Error("Impl namespace member access expects identifier");
    if (member.name === "interface") {
      return { kind: "String", value: obj.meta.interfaceName };
    }
    if (member.name === "target") {
      return { kind: "String", value: ctx.typeExpressionToString(obj.meta.target) };
    }
    if (member.name === "interface_args") {
      const args = obj.meta.interfaceArgs ?? [];
      return ctx.makeArrayValue(args.map(a => ({ kind: "String", value: ctx.typeExpressionToString(a) } as RuntimeValue)));
    }
    const sym = obj.symbols.get(member.name);
    if (!sym) throw new Error(`No method '${member.name}' on impl ${obj.def.implName?.name ?? "<unnamed>"}`);
    return sym;
  }
  if (obj.kind === "interface_value") {
    if (member.type !== "Identifier") throw new Error("Interface member access expects identifier");
    const underlying = obj.value;
    if (obj.interfaceName === "Error" && member.type === "Identifier") {
      if (member.name === "value") {
        if (underlying.kind === "error") {
          return underlying.value ?? { kind: "nil", value: null };
        }
        return { kind: "nil", value: null };
      }
      if (underlying.kind === "error") {
        const fn = (ctx.errorNativeMethods as Record<string, Extract<RuntimeValue, { kind: "native_function" }>>)[member.name];
        if (fn) {
          return ctx.bindNativeMethod(fn, underlying);
        }
      }
    }
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
    const methodNode = method.kind === "function_overload" ? method.overloads[0]?.node : method.node;
    if (methodNode?.type === "FunctionDefinition" && methodNode.isPrivate) {
      throw new Error(`Method '${member.name}' on ${typeName} is private`);
    }
    return { kind: "bound_method", func: method, self: underlying };
  }
  if (obj.kind === "proc_handle") {
    if (member.type !== "Identifier") throw new Error("Proc handle member access expects identifier");
    const fn = (ctx.procNativeMethods as Record<string, Extract<RuntimeValue, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown proc handle method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "future") {
    if (member.type !== "Identifier") throw new Error("Future member access expects identifier");
    const fn = (ctx.futureNativeMethods as Record<string, Extract<RuntimeValue, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown future method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "iterator") {
    if (member.type !== "Identifier") throw new Error("Iterator member access expects identifier");
    ctx.ensureIteratorBuiltins();
    const fn = (ctx.iteratorNativeMethods as Record<string, Extract<RuntimeValue, { kind: "native_function" }>>)[member.name];
    if (!fn) throw new Error(`Unknown iterator method '${member.name}'`);
    return ctx.bindNativeMethod(fn, obj);
  }
  if (obj.kind === "error") {
    if (member.type !== "Identifier") throw new Error("Error member access expects identifier");
    if (member.name === "value") {
      return obj.value ?? { kind: "nil", value: null };
    }
    const fn = (ctx.errorNativeMethods as Record<string, Extract<RuntimeValue, { kind: "native_function" }>>)[member.name];
    if (fn) {
      return ctx.bindNativeMethod(fn, obj);
    }
    const bound = ctx.resolveMethodFromPool(env, member.name, obj);
    if (bound) return bound;
    throw new Error(`No field or method named '${member.name}' on error value`);
  }

  if (member.type === "Identifier" && obj.kind !== "struct_instance" && obj.kind !== "array" && obj.kind !== "String") {
    const bound = ctx.resolveMethodFromPool(env, member.name, obj);
    if (bound) return bound;
    const typeName = ctx.getTypeNameForValue(obj);
    if (typeName) {
      const method = ctx.findMethod(typeName, member.name);
      if (method) {
        const methodNode = method.kind === "function_overload" ? method.overloads[0]?.node : method.node;
        if (methodNode?.type === "FunctionDefinition" && methodNode.isPrivate) {
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
      if (obj.values.has(member.name)) {
        const fieldVal = obj.values.get(member.name)!;
        if (opts?.preferMethods) {
          if (isCallable(fieldVal)) {
            return fieldVal;
          }
          const bound = ctx.resolveMethodFromPool(env, member.name, obj);
          if (bound) return bound;
        }
        return fieldVal;
      }
      const bound = ctx.resolveMethodFromPool(env, member.name, obj);
      if (bound) return bound;
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
      if (opts?.preferMethods) {
        const bound = ctx.resolveMethodFromPool(env, name, obj);
        if (bound) return bound;
      }
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
      const bound = ctx.resolveMethodFromPool(env, name, obj);
      if (bound) return bound;
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
  if (obj.kind === "String") {
    if (member.type !== "Identifier") {
      throw new Error("String member access expects identifier");
    }
    const name = member.name;
    const bound = ctx.resolveMethodFromPool(env, name, obj);
    if (bound) return bound;
    throw new Error(`String has no member '${name}' (import able.text.string for stdlib helpers)`);
  }
  if (member.type === "Identifier") {
    const bound = ctx.resolveMethodFromPool(env, member.name, obj);
    if (bound) return bound;
  }
  throw new Error("Member access only supported on structs/arrays in this milestone");
}

export function evaluateMemberAccessExpression(ctx: Interpreter, node: AST.MemberAccessExpression, env: Environment): RuntimeValue {
  const obj = ctx.evaluate(node.object, env);
  if (node.isSafe && obj.kind === "nil") {
    return obj;
  }
  return memberAccessOnValue(ctx, obj, node.member, env);
}

export function evaluateImplicitMemberExpression(ctx: Interpreter, node: AST.ImplicitMemberExpression, env: Environment): RuntimeValue {
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

function toSafeIndex(value: RuntimeValue | undefined, label: string): number {
  const raw = numericToNumber(value ?? { kind: "nil", value: null }, label, { requireSafeInteger: true });
  const idx = Math.trunc(raw);
  if (idx < 0) throw new Error(`${label} must be non-negative`);
  return idx;
}
