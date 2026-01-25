import * as AST from "../../ast";
import type { Interpreter } from "../index";
import { isFloatValue, isIntegerValue } from "../numeric";
import type { FloatKind, IntegerKind, RuntimeValue } from "../values";
import {
  FLOAT_KINDS,
  INTEGER_KINDS,
  hasConcreteTypeName,
  integerRangeWithin,
  integerValueWithinRange,
  isAwaitableStructInstance,
  isErrorValue,
  normalizeKernelAliasName,
} from "./helpers";

export function applyTypeUnionAugmentations(cls: typeof Interpreter): void {
  cls.prototype.matchesType = function matchesType(this: Interpreter, t: AST.TypeExpression, v: RuntimeValue): boolean {
    const target = this.expandTypeAliases(t);
    if (v.kind === "error" && v.value) {
      const targetIsError = target.type === "SimpleTypeExpression" && normalizeKernelAliasName(target.name.name) === "Error";
      if (!targetIsError && this.matchesType(target, v.value)) {
        return true;
      }
    }
    const valueTypeExpr = this.typeExpressionFromValue(v);
    if (valueTypeExpr) {
      const canonicalValue = this.expandTypeAliases(valueTypeExpr);
      if (this.typeExpressionsEqual(target, canonicalValue)) {
        return true;
      }
    }
    switch (target.type) {
      case "WildcardTypeExpression":
        return true;
      case "SimpleTypeExpression": {
        const name = normalizeKernelAliasName(target.name.name);
        if (name === "Self") return true;
        if (/^[A-Z]$/.test(name)) return true;
        if (name === "Value") return true;
        if (name === "String") return v.kind === "String";
        if (name === "bool") return v.kind === "bool";
        if (name === "char") return v.kind === "char";
        if (name === "nil") return v.kind === "nil";
        if (name === "void") return v.kind === "void";
        if (this.unions.has(name)) {
          const unionDef = this.unions.get(name)!;
          return (unionDef.variants ?? []).some((variant) => this.matchesType(variant, v));
        }
        if (INTEGER_KINDS.includes(name as IntegerKind)) {
          if (!isIntegerValue(v)) {
            return false;
          }
          const actualKind = v.kind as IntegerKind;
          const expectedKind = name as IntegerKind;
          return (
            actualKind === expectedKind ||
            integerRangeWithin(actualKind, expectedKind) ||
            integerValueWithinRange(v.value, expectedKind)
          );
        }
        if (FLOAT_KINDS.includes(name as FloatKind)) {
          if (isFloatValue(v)) return true;
          return isIntegerValue(v);
        }
        if (name === "Error" && v.kind === "error") return true;
        if ((name === "IoHandle" || name === "ProcHandle") && v.kind === "host_handle") {
          return v.handleType === name;
        }
        if (name === "IteratorEnd" && v.kind === "iterator_end") return true;
        if (name === "Iterator" && v.kind === "iterator") return true;
        if (name === "Awaitable" && isAwaitableStructInstance(v)) return true;
        if (this.interfaces.has(name) && !hasConcreteTypeName(this, name)) {
          if (v.kind === "interface_value") return v.interfaceName === name;
          const typeName = this.getTypeNameForValue(v);
          const canonicalName = typeName
            ? this.parseTypeExpression(this.expandTypeAliases(AST.simpleTypeExpression(typeName)))?.name
            : null;
          if (!typeName && !canonicalName) return false;
          if (canonicalName && canonicalName !== typeName && canonicalName === name) {
            return true;
          }
          if (!typeName) return false;
          const subjectTypeArgs = v.kind === "struct_instance" ? v.typeArguments : undefined;
          return this.typeImplementsInterface(typeName, name, { subjectTypeArgs });
        }
        if (v.kind === "struct_instance") {
          const typeName = v.def.id.name;
          if (typeName === name) return true;
          const canonicalName = this.parseTypeExpression(this.expandTypeAliases(AST.simpleTypeExpression(typeName)))?.name;
          return canonicalName === name;
        }
        return false;
      }
      case "GenericTypeExpression": {
        if (target.base.type === "SimpleTypeExpression") {
          const baseName = normalizeKernelAliasName(target.base.name.name);
          if (baseName === "Self" || /^[A-Z]$/.test(baseName)) {
            return true;
          }
        }
        if (
          target.base.type === "SimpleTypeExpression" &&
          normalizeKernelAliasName(target.base.name.name) === "Array"
        ) {
          const isArrayValue = v.kind === "array" || (v.kind === "struct_instance" && v.def.id.name === "Array");
          if (!isArrayValue) return false;
          if (!target.arguments || target.arguments.length === 0) return true;
          const elemT = target.arguments[0]!;
          if (v.kind === "array") {
            return v.elements.every((el) => this.matchesType(elemT, el));
          }
          return true;
        }
        if (target.base.type === "SimpleTypeExpression") {
          const baseName = normalizeKernelAliasName(target.base.name.name);
          if (v.kind === "interface_value" && v.interfaceName === baseName) {
            return true;
          }
          if (this.unions.has(baseName)) {
            const unionDef = this.unions.get(baseName)!;
            return (unionDef.variants ?? []).some((variant) => this.matchesType(variant, v));
          }
          if (baseName === "Iterator" && v.kind === "iterator") {
            return true;
          }
          if (baseName === "Awaitable" && isAwaitableStructInstance(v)) {
            return true;
          }
          const valueTypeName = this.getTypeNameForValue(v);
          const canonicalValueName = valueTypeName
            ? this.parseTypeExpression(this.expandTypeAliases(AST.simpleTypeExpression(valueTypeName)))?.name
            : null;
          const matchesBase =
            valueTypeName === baseName ||
            (!!canonicalValueName && canonicalValueName === baseName);
          if (!matchesBase) {
            if (this.interfaces.has(baseName)) {
              const subjectTypeArgs = v.kind === "struct_instance" ? v.typeArguments : undefined;
              const interfaceArgs = target.arguments ?? [];
              if (valueTypeName && this.typeImplementsInterface(valueTypeName, baseName, { subjectTypeArgs, interfaceArgs })) {
                return true;
              }
            }
            return false;
          }
          if (v.kind === "struct_instance") {
            if (!target.arguments || target.arguments.length === 0) return true;
            const actualArgs = v.typeArguments ?? [];
            if (actualArgs.length === 0) return true;
            if (actualArgs.length !== target.arguments.length) return false;
            return target.arguments.every((expected, idx) => {
              const actual = actualArgs[idx]!;
              if (actual.type === "WildcardTypeExpression") return true;
              if (expected.type === "WildcardTypeExpression") return true;
              if (actual.type === "SimpleTypeExpression") {
                const name = actual.name.name;
                if (name === "Self" || /^[A-Z]/.test(name)) return true;
              }
              if (expected.type === "SimpleTypeExpression") {
                const name = expected.name.name;
                if (name === "Self" || /^[A-Z]/.test(name)) return true;
              }
              return this.typeExpressionsEqual(expected, actual);
            });
          }
        }
        return true;
      }
      case "FunctionTypeExpression":
        return (
          v.kind === "function" ||
          v.kind === "function_overload" ||
          v.kind === "bound_method" ||
          v.kind === "native_function" ||
          v.kind === "native_bound_method" ||
          v.kind === "partial_function"
        );
      case "NullableTypeExpression":
        if (v.kind === "nil") return true;
        return this.matchesType(target.innerType, v);
      case "ResultTypeExpression":
        if (isErrorValue(this, v)) return true;
        return this.matchesType(target.innerType, v);
      case "UnionTypeExpression":
        return target.members.some((member) => this.matchesType(member, v));
      default:
        return true;
    }
  };
}
