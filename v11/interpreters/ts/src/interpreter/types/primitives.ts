import * as AST from "../../ast";
import type { Interpreter } from "../index";
import { getIntegerInfo, isFloatValue, isIntegerValue, makeFloatValue, makeIntegerValue } from "../numeric";
import type { FloatKind, IntegerKind, RuntimeValue } from "../values";
import {
  FLOAT_KINDS,
  INTEGER_KIND_SET,
  INTEGER_KINDS,
  hasConcreteTypeName,
  integerValueWithinRange,
  isSingletonStructDef,
} from "./helpers";

export function applyTypePrimitiveAugmentations(cls: typeof Interpreter): void {
  cls.prototype.typeExpressionFromValue = function typeExpressionFromValue(
    this: Interpreter,
    value: RuntimeValue,
    seen?: WeakSet<object>,
  ): AST.TypeExpression | null {
    const seenValues = seen ?? new WeakSet<object>();
    if (seenValues.has(value as object)) {
      if (process.env.ABLE_TRACE_ERRORS) {
        const typeName = this.getTypeNameForValue(value) ?? value.kind;
        console.error(`[trace] typeExpressionFromValue cycle on ${typeName}`);
      }
      return AST.wildcardTypeExpression();
    }
    seenValues.add(value as object);
    try {
      switch (value.kind) {
      case "String":
        return AST.simpleTypeExpression("String");
      case "bool":
        return AST.simpleTypeExpression("bool");
      case "char":
        return AST.simpleTypeExpression("char");
      case "nil":
        return AST.simpleTypeExpression("nil");
      case "error":
        return AST.simpleTypeExpression("Error");
      case "host_handle":
        return AST.simpleTypeExpression(value.handleType);
      case "struct_instance": {
        const base = AST.simpleTypeExpression(value.def.id.name);
        const generics = value.def.genericParams ?? [];
        if (generics.length > 0) {
          const genericNames = new Set(generics.map(gp => gp.name.name));
          let typeArgs = value.typeArguments ?? [];
          let needsInference = typeArgs.length !== generics.length;
          if (!needsInference) {
            for (const arg of typeArgs) {
              if (!arg || arg.type === "WildcardTypeExpression") {
                needsInference = true;
                break;
              }
              if (arg.type === "SimpleTypeExpression" && genericNames.has(arg.name.name)) {
                needsInference = true;
                break;
              }
            }
          }
          if (needsInference) {
            const bindings = new Map<string, AST.TypeExpression>();
            if (Array.isArray(value.values)) {
              for (let i = 0; i < value.def.fields.length && i < value.values.length; i++) {
                const field = value.def.fields[i];
                const actualVal = value.values[i];
                if (!field?.fieldType || !actualVal) continue;
                const actual = this.typeExpressionFromValue(actualVal, seenValues);
                if (!actual) continue;
                this.matchTypeExpressionTemplate(field.fieldType, actual, genericNames, bindings);
              }
            } else {
              for (const field of value.def.fields) {
                if (!field?.name || !field.fieldType) continue;
                const actualVal = value.values.get(field.name.name);
                if (!actualVal) continue;
                const actual = this.typeExpressionFromValue(actualVal, seenValues);
                if (!actual) continue;
                this.matchTypeExpressionTemplate(field.fieldType, actual, genericNames, bindings);
              }
            }
            typeArgs = generics.map(gp => bindings.get(gp.name.name) ?? AST.wildcardTypeExpression());
          }
          if (typeArgs.length > 0) {
            return AST.genericTypeExpression(base, typeArgs);
          }
        }
        return base;
      }
      case "iterator_end":
        return AST.simpleTypeExpression("IteratorEnd");
      case "iterator":
        return AST.simpleTypeExpression("Iterator");
      case "struct_def": {
        if (!isSingletonStructDef(value.def)) return null;
        return AST.simpleTypeExpression(value.def.id.name);
      }
      case "interface_value":
        return AST.simpleTypeExpression(value.interfaceName);
      case "array": {
        const state = this.ensureArrayState ? this.ensureArrayState(value) : { values: value.elements };
        const elems = state.values ?? value.elements ?? [];
        let elemType: AST.TypeExpression | null = null;
        for (const el of elems) {
          const inferred = this.typeExpressionFromValue(el, seenValues);
          if (!inferred) continue;
          if (!elemType) {
            elemType = inferred;
            continue;
          }
          if (!this.typeExpressionsEqual(elemType, inferred)) {
            elemType = AST.wildcardTypeExpression();
            break;
          }
        }
        if (!elemType) elemType = AST.wildcardTypeExpression();
        return AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [elemType]);
      }
      case "hash_map":
        return AST.simpleTypeExpression("Map");
      default: {
        if (value.kind === "interface_def" && (value as any).def?.id?.name) {
          return AST.simpleTypeExpression((value as any).def.id.name);
        }
        if ((value as any).kind === "i8" || (value as any).kind === "i16" || (value as any).kind === "i32" || (value as any).kind === "i64" || (value as any).kind === "i128" || (value as any).kind === "u8" || (value as any).kind === "u16" || (value as any).kind === "u32" || (value as any).kind === "u64" || (value as any).kind === "u128") {
          return AST.simpleTypeExpression((value as any).kind);
        }
        if ((value as any).kind === "f32" || (value as any).kind === "f64") {
          return AST.simpleTypeExpression((value as any).kind);
        }
        return null;
      }
      }
    } finally {
      seenValues.delete(value as object);
    }
  };

  cls.prototype.getTypeNameForValue = function getTypeNameForValue(this: Interpreter, value: RuntimeValue): string | null {
    if (INTEGER_KINDS.includes(value.kind as IntegerKind)) {
      return value.kind;
    }
    if (FLOAT_KINDS.includes(value.kind as FloatKind)) {
      return value.kind;
    }
    switch (value.kind) {
      case "struct_instance":
        return value.def.id.name;
      case "iterator_end":
        return "IteratorEnd";
      case "struct_def":
        return isSingletonStructDef(value.def) ? value.def.id.name : null;
      case "type_ref":
        return value.typeName;
      case "interface_value":
        return this.getTypeNameForValue(value.value);
      case "error":
        return "Error";
      case "host_handle":
        return value.handleType;
      case "String":
        return "String";
      case "bool":
        return "bool";
      case "char":
        return "char";
      case "void":
        return "void";
      case "array":
        return "Array";
      case "iterator":
        return "Iterator";
      default:
        return null;
    }
  };

  cls.prototype.coerceValueToType = function coerceValueToType(
    this: Interpreter,
    typeExpr: AST.TypeExpression | undefined,
    value: RuntimeValue,
  ): RuntimeValue {
    if (!typeExpr) return value;
    const canonical = this.expandTypeAliases(typeExpr);
    if (value.kind === "error" && value.value) {
      const targetIsError = canonical.type === "SimpleTypeExpression" && canonical.name.name === "Error";
      if (!targetIsError && this.matchesType(canonical, value.value)) {
        return this.coerceValueToType(canonical, value.value);
      }
    }
    if (canonical.type === "SimpleTypeExpression") {
      const name = canonical.name.name;
      if (INTEGER_KIND_SET.has(name as IntegerKind) && isIntegerValue(value)) {
        const targetKind = name as IntegerKind;
        const actualKind = value.kind as IntegerKind;
        if (actualKind !== targetKind && integerValueWithinRange(value.value, targetKind)) {
          return makeIntegerValue(targetKind, value.value);
        }
      }
      if (this.interfaces.has(name) && !hasConcreteTypeName(this, name)) {
        return this.toInterfaceValue(name, value);
      }
      if (FLOAT_KINDS.includes(name as FloatKind)) {
        const targetKind = name as FloatKind;
        if (isFloatValue(value) && value.kind !== targetKind) {
          return makeFloatValue(targetKind, value.value);
        }
        if (isIntegerValue(value)) {
          return makeFloatValue(targetKind, Number(value.value));
        }
      }
    }
    if (canonical.type === "GenericTypeExpression" && canonical.base.type === "SimpleTypeExpression") {
      const baseName = canonical.base.name.name;
      if (this.interfaces.has(baseName) && !hasConcreteTypeName(this, baseName)) {
        return this.toInterfaceValue(baseName, value, canonical.arguments ?? []);
      }
    }
    return value;
  };

  cls.prototype.castValueToType = function castValueToType(
    this: Interpreter,
    typeExpr: AST.TypeExpression,
    value: RuntimeValue,
  ): RuntimeValue {
    const canonical = this.expandTypeAliases(typeExpr);
    if (this.matchesType(canonical, value)) return value;
    const rawValue = value.kind === "interface_value" ? value.value : value;
    if (canonical.type === "SimpleTypeExpression") {
      const name = canonical.name.name;
      if (INTEGER_KIND_SET.has(name as IntegerKind)) {
        const targetKind = name as IntegerKind;
        if (isIntegerValue(rawValue)) {
          const info = getIntegerInfo(targetKind);
          const modulus = 1n << BigInt(info.bits);
          let normalized = rawValue.value % modulus;
          if (normalized < 0n) {
            normalized += modulus;
          }
          if (info.signed) {
            const signBit = 1n << BigInt(info.bits - 1);
            if (normalized >= signBit) {
              normalized -= modulus;
            }
          }
          return makeIntegerValue(targetKind, normalized);
        }
        if (isFloatValue(rawValue)) {
          if (!Number.isFinite(rawValue.value)) {
            throw new Error(`cannot cast non-finite float to ${targetKind}`);
          }
          const truncated = BigInt(Math.trunc(rawValue.value));
          if (!integerValueWithinRange(truncated, targetKind)) {
            throw new Error(`value out of range for ${targetKind}`);
          }
          return makeIntegerValue(targetKind, truncated);
        }
      }
      if (FLOAT_KINDS.includes(name as FloatKind)) {
        const targetKind = name as FloatKind;
        if (isFloatValue(rawValue)) {
          return makeFloatValue(targetKind, rawValue.value);
        }
        if (isIntegerValue(rawValue)) {
          return makeFloatValue(targetKind, Number(rawValue.value));
        }
      }
      if (name === "Error" && rawValue.kind === "error") {
        return rawValue;
      }
      if (this.interfaces.has(name) && !hasConcreteTypeName(this, name)) {
        return this.toInterfaceValue(name, value);
      }
    }
    if (canonical.type === "GenericTypeExpression" && canonical.base.type === "SimpleTypeExpression") {
      const baseName = canonical.base.name.name;
      if (this.interfaces.has(baseName) && !hasConcreteTypeName(this, baseName)) {
        return this.toInterfaceValue(baseName, value, canonical.arguments ?? []);
      }
    }
    throw new Error(`cannot cast ${this.getTypeNameForValue(rawValue) ?? rawValue.kind} to ${this.typeExpressionToString(canonical)}`);
  };
}
