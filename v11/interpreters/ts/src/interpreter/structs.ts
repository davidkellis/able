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
  let typeArguments: AST.TypeExpression[] | undefined = node.typeArguments;
  let typeArgMap: Map<string, AST.TypeExpression> | undefined;
  if (generics && generics.length > 0) {
    const appliedTypeArgs = typeArguments ?? generics.map(() => AST.wildcardTypeExpression());
    typeArguments = appliedTypeArgs;
    typeArgMap = ctx.mapTypeArguments(generics, appliedTypeArgs, `instantiating ${structDef.id.name}`);
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
    ctx.ensureArrayState(obj);
    if (member.type === "Identifier") {
      const name = member.name;
      const ufcsFirst = ctx.tryUfcs(env, name, obj);
      if (ufcsFirst) return ufcsFirst;
      const state = ctx.ensureArrayState(obj);
      if (name === "storage_handle") {
        const handle = obj.handle ?? 0;
        return makeIntegerFromNumber("i64", handle);
      }
      if (name === "length") {
        return makeIntegerFromNumber("i32", state.values.length);
      }
      if (name === "capacity") {
        return makeIntegerFromNumber("i32", state.capacity);
      }
      const methods = ctx.arrayNativeMethods;
      if (name === "size") {
        const fn = (methods.size ??= ctx.makeNativeFunction("array.size", 1, (_interp, [self]) => {
          if (!self || self.kind !== "array") throw new Error("size receiver must be an array");
          const state = ctx.ensureArrayState(self);
          return makeIntegerFromNumber("u64", state.values.length);
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "push") {
        const fn = (methods.push ??= ctx.makeNativeFunction("array.push", 2, (_interp, [self, value]) => {
          if (!self || self.kind !== "array") throw new Error("push receiver must be an array");
          const state = ctx.ensureArrayState(self);
          state.values.push(value ?? { kind: "nil", value: null });
          state.capacity = Math.max(state.capacity, state.values.length);
          return { kind: "nil", value: null };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "pop") {
        const fn = (methods.pop ??= ctx.makeNativeFunction("array.pop", 1, (_interp, [self]) => {
          if (!self || self.kind !== "array") throw new Error("pop receiver must be an array");
          const state = ctx.ensureArrayState(self);
          if (state.values.length === 0) return { kind: "nil", value: null };
          const value = state.values.pop();
          return value ?? { kind: "nil", value: null };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "get") {
        const fn = (methods.get ??= ctx.makeNativeFunction("array.get", 2, (_interp, [self, index]) => {
          if (!self || self.kind !== "array") throw new Error("get receiver must be an array");
          const state = ctx.ensureArrayState(self);
          const idx = toSafeIndex(index, "index");
          if (idx < 0 || idx >= state.values.length) return { kind: "nil", value: null };
          const value = state.values[idx];
          return value ?? { kind: "nil", value: null };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "set") {
        const fn = (methods.set ??= ctx.makeNativeFunction("array.set", 3, (interp, [self, index, value]) => {
          if (!self || self.kind !== "array") throw new Error("set receiver must be an array");
          const state = ctx.ensureArrayState(self);
          const idx = toSafeIndex(index, "index");
          if (idx < 0 || idx >= state.values.length) {
            return interp.makeRuntimeError(`index ${idx} out of bounds for length ${state.values.length}`);
          }
          state.values[idx] = value ?? { kind: "nil", value: null };
          return { kind: "nil", value: null };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "clear") {
        const fn = (methods.clear ??= ctx.makeNativeFunction("array.clear", 1, (_interp, [self]) => {
          if (!self || self.kind !== "array") throw new Error("clear receiver must be an array");
          const state = ctx.ensureArrayState(self);
          state.values.length = 0;
          return { kind: "nil", value: null };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      if (name === "iterator") {
        const fn = (methods.iterator ??= ctx.makeNativeFunction("array.iterator", 1, (_interp, [self]) => {
          if (!self || self.kind !== "array") throw new Error("iterator receiver must be an array");
          const state = ctx.ensureArrayState(self);
          let idx = 0;
          return {
            kind: "iterator",
            iterator: {
              next: () => {
                const current = ctx.ensureArrayState(self);
                if (idx >= current.values.length) return { done: true, value: ctx.iteratorEndValue };
                const value = current.values[idx];
                idx += 1;
                return { done: false, value: value ?? { kind: "nil", value: null } };
              },
              close: () => {},
            },
          };
        }));
        return ctx.bindNativeMethod(fn, obj);
      }
      const method = ctx.findMethod("Array", name, { typeArgs: [AST.wildcardTypeExpression()] });
      if (method) {
        return { kind: "bound_method", func: method, self: obj };
      }
      const ufcs = ctx.tryUfcs(env, name, obj);
      if (ufcs) return ufcs;
      throw new Error(`Array has no member '${name}'`);
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
    const methods = ctx.stringNativeMethods;
    if (name === "len_bytes") {
      const fn = (methods.len_bytes ??= ctx.makeNativeFunction("string.len_bytes", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("len_bytes receiver must be a string");
        const encoded = new TextEncoder().encode(self.value);
        return makeIntegerFromNumber("u64", encoded.length);
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "len_chars") {
      const fn = (methods.len_chars ??= ctx.makeNativeFunction("string.len_chars", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("len_chars receiver must be a string");
        return makeIntegerFromNumber("u64", Array.from(self.value).length);
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "len_graphemes") {
      const fn = (methods.len_graphemes ??= ctx.makeNativeFunction("string.len_graphemes", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("len_graphemes receiver must be a string");
        return makeIntegerFromNumber("u64", segmentGraphemes(self.value).length);
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "substring") {
      const fn = (methods.substring ??= ctx.makeNativeFunction("string.substring", -1, (interp, args) => {
        if (args.length < 2 || args.length > 3) throw new Error("substring expects start and optional length");
        const self = args[0];
        if (!self || self.kind !== "string") throw new Error("substring receiver must be a string");
        const start = numericToNumber(args[1] ?? { kind: "nil", value: null }, "start", { requireSafeInteger: true });
        if (!Number.isInteger(start) || start < 0) {
          return interp.makeRuntimeError("substring start must be a non-negative integer");
        }
        let length: number | undefined;
        if (args.length === 3 && args[2] && args[2].kind !== "nil") {
          const parsed = numericToNumber(args[2], "length", { requireSafeInteger: true });
          if (parsed < 0) {
            return interp.makeRuntimeError("substring length must be non-negative");
          }
          length = parsed;
        }
        const codepoints = Array.from(self.value);
        if (start > codepoints.length) {
          return interp.makeRuntimeError("substring start out of range");
        }
        const end = length === undefined ? codepoints.length : start + length;
        if (end > codepoints.length) {
          return interp.makeRuntimeError("substring range out of bounds");
        }
        return { kind: "string", value: codepoints.slice(start, end).join("") };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "split") {
      const fn = (methods.split ??= ctx.makeNativeFunction("string.split", 2, (_interp, [self, delimiter]) => {
        if (!self || self.kind !== "string") throw new Error("split receiver must be a string");
        if (!delimiter || delimiter.kind !== "string") throw new Error("split delimiter must be a string");
        const parts =
          delimiter.value === ""
            ? segmentGraphemes(self.value)
            : self.value.split(delimiter.value);
        return {
          kind: "array",
          elements: parts.map((part) => ({ kind: "string", value: part } as V10Value)),
        };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "replace") {
      const fn = (methods.replace ??= ctx.makeNativeFunction("string.replace", 3, (_interp, [self, oldPart, newPart]) => {
        if (!self || self.kind !== "string") throw new Error("replace receiver must be a string");
        if (!oldPart || oldPart.kind !== "string") throw new Error("replace target must be a string");
        if (!newPart || newPart.kind !== "string") throw new Error("replace replacement must be a string");
        return { kind: "string", value: self.value.split(oldPart.value).join(newPart.value) };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "starts_with") {
      const fn = (methods.starts_with ??= ctx.makeNativeFunction("string.starts_with", 2, (_interp, [self, prefix]) => {
        if (!self || self.kind !== "string") throw new Error("starts_with receiver must be a string");
        if (!prefix || prefix.kind !== "string") throw new Error("starts_with prefix must be a string");
        return { kind: "bool", value: self.value.startsWith(prefix.value) };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "ends_with") {
      const fn = (methods.ends_with ??= ctx.makeNativeFunction("string.ends_with", 2, (_interp, [self, suffix]) => {
        if (!self || self.kind !== "string") throw new Error("ends_with receiver must be a string");
        if (!suffix || suffix.kind !== "string") throw new Error("ends_with suffix must be a string");
        return { kind: "bool", value: self.value.endsWith(suffix.value) };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "chars") {
      const fn = (methods.chars ??= ctx.makeNativeFunction("string.chars", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("chars receiver must be a string");
        const cps = Array.from(self.value);
        let idx = 0;
        return {
          kind: "iterator",
          iterator: {
            next: () => {
              if (idx >= cps.length) return { done: true, value: ctx.iteratorEndValue };
              const cp = cps[idx];
              idx += 1;
              return { done: false, value: { kind: "char", value: cp } };
            },
            close: () => {},
          },
        };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "graphemes") {
      const fn = (methods.graphemes ??= ctx.makeNativeFunction("string.graphemes", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("graphemes receiver must be a string");
        const pieces = segmentGraphemes(self.value);
        let idx = 0;
        return {
          kind: "iterator",
          iterator: {
            next: () => {
              if (idx >= pieces.length) return { done: true, value: ctx.iteratorEndValue };
              const piece = pieces[idx];
              idx += 1;
              return { done: false, value: { kind: "string", value: piece } };
            },
            close: () => {},
          },
        };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    if (name === "bytes") {
      const fn = (methods.bytes ??= ctx.makeNativeFunction("string.bytes", 1, (_interp, [self]) => {
        if (!self || self.kind !== "string") throw new Error("bytes receiver must be a string");
        const encoded = new TextEncoder().encode(self.value);
        let idx = 0;
        return {
          kind: "iterator",
          iterator: {
            next: () => {
              if (idx >= encoded.length) return { done: true, value: ctx.iteratorEndValue };
              const value = makeIntegerFromNumber("u8", encoded[idx] ?? 0);
              idx += 1;
              return { done: false, value };
            },
            close: () => {},
          },
        };
      }));
      return ctx.bindNativeMethod(fn, obj);
    }
    const ufcs = ctx.tryUfcs(env, name, obj);
    if (ufcs) return ufcs;
    throw new Error(`String has no member '${name}'`);
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

const segmentGraphemes = (text: string): string[] => {
  const segmenterCtor = (Intl as any)?.Segmenter;
  if (typeof segmenterCtor === "function") {
    const segmenter = new segmenterCtor(undefined, { granularity: "grapheme" });
    return Array.from(segmenter.segment(text)).map((entry: any) => entry.segment as string);
  }
  return Array.from(text);
};
