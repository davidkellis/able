import * as AST from "../ast";
import type { Interpreter } from "./index";
import { getIntegerInfo, isFloatValue, isIntegerValue, makeFloatValue, makeIntegerValue } from "./numeric";
import type { FloatKind, IntegerKind, RuntimeValue } from "./values";

const INTEGER_KINDS: IntegerKind[] = ["i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128"];
const FLOAT_KINDS: FloatKind[] = ["f32", "f64"];
const INTEGER_KIND_SET: Set<IntegerKind> = new Set(INTEGER_KINDS);

function isSingletonStructDef(def: AST.StructDefinition): boolean {
  if (!def || (def.genericParams && def.genericParams.length > 0)) return false;
  if (def.kind === "singleton") return true;
  return def.kind === "named" && def.fields.length === 0;
}

function integerRangeWithin(source: IntegerKind, target: IntegerKind): boolean {
  const sourceInfo = getIntegerInfo(source);
  const targetInfo = getIntegerInfo(target);
  return sourceInfo.min >= targetInfo.min && sourceInfo.max <= targetInfo.max;
}

function integerValueWithinRange(raw: bigint, target: IntegerKind): boolean {
  const targetInfo = getIntegerInfo(target);
  return raw >= targetInfo.min && raw <= targetInfo.max;
}

function normalizeKernelAliasName(name: string): string {
  switch (name) {
    case "string":
      return "String";
    case "KernelArray":
      return "Array";
    case "KernelChannel":
      return "Channel";
    case "KernelHashMap":
      return "HashMap";
    case "KernelMutex":
      return "Mutex";
    case "KernelRange":
      return "Range";
    case "KernelRangeFactory":
      return "RangeFactory";
    case "KernelRatio":
      return "Ratio";
    case "KernelAwaitable":
      return "Awaitable";
    case "KernelAwaitWaker":
      return "AwaitWaker";
    case "KernelAwaitRegistration":
      return "AwaitRegistration";
    default:
      return name;
  }
}

function isErrorValue(ctx: Interpreter, value: RuntimeValue): boolean {
  if (value.kind === "error") return true;
  if (value.kind === "interface_value" && value.interfaceName === "Error") return true;
  const typeName = ctx.getTypeNameForValue(value);
  if (!typeName) return false;
  const typeArgs = value.kind === "struct_instance" ? value.typeArguments : undefined;
  return ctx.typeImplementsInterface(typeName, "Error", typeArgs);
}

function isAwaitableStructInstance(value: RuntimeValue): boolean {
  if (value.kind !== "struct_instance") return false;
  const values = value.values as Map<string, RuntimeValue> | undefined;
  if (!(values instanceof Map)) return false;
  return values.has("is_ready") && values.has("register") && values.has("commit");
}

declare module "./index" {
  interface Interpreter {
    expandTypeAliases(t: AST.TypeExpression): AST.TypeExpression;
    typeExpressionToString(t: AST.TypeExpression): string;
    parseTypeExpression(t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null;
    typeExpressionsEqual(a: AST.TypeExpression, b: AST.TypeExpression): boolean;
    cloneTypeExpression(t: AST.TypeExpression): AST.TypeExpression;
    typeExpressionFromValue(value: RuntimeValue): AST.TypeExpression | null;
    matchesType(t: AST.TypeExpression, v: RuntimeValue): boolean;
    getTypeNameForValue(value: RuntimeValue): string | null;
    typeImplementsInterface(typeName: string, interfaceName: string, typeArgs?: AST.TypeExpression[]): boolean;
    coerceValueToType(typeExpr: AST.TypeExpression | undefined, value: RuntimeValue): RuntimeValue;
    castValueToType(typeExpr: AST.TypeExpression, value: RuntimeValue): RuntimeValue;
    toInterfaceValue(interfaceName: string, rawValue: RuntimeValue): RuntimeValue;
  }
}

export function applyTypesAugmentations(cls: typeof Interpreter): void {
  cls.prototype.expandTypeAliases = function expandTypeAliases(this: Interpreter, t: AST.TypeExpression, seen = new Set<string>()): AST.TypeExpression {
    if (!t) return t;
    switch (t.type) {
      case "SimpleTypeExpression": {
        const name = t.name.name;
        const alias = this.typeAliases.get(name);
        if (!alias?.targetType || seen.has(name)) {
          return t;
        }
        seen.add(name);
        const expanded = this.expandTypeAliases(alias.targetType, seen);
        seen.delete(name);
        return expanded ?? t;
      }
      case "GenericTypeExpression": {
        const baseName = t.base.type === "SimpleTypeExpression" ? t.base.name.name : null;
        const expandedBase = this.expandTypeAliases(t.base, seen);
        const expandedArgs = (t.arguments ?? []).map((arg) => (arg ? this.expandTypeAliases(arg, seen) : arg));
        if (!baseName) {
          return { ...t, base: expandedBase, arguments: expandedArgs };
        }
        const alias = this.typeAliases.get(baseName);
        if (!alias?.targetType || seen.has(baseName)) {
          return { ...t, base: expandedBase, arguments: expandedArgs };
        }
        const substitutions = new Map<string, AST.TypeExpression>();
        (alias.genericParams ?? []).forEach((param, index) => {
          const paramName = param?.name?.name;
          if (!paramName) return;
          substitutions.set(paramName, expandedArgs[index] ?? AST.wildcardTypeExpression());
        });
        const substitute = (expr: AST.TypeExpression): AST.TypeExpression => {
          switch (expr.type) {
            case "SimpleTypeExpression": {
              const name = expr.name.name;
              if (substitutions.has(name)) {
                return this.expandTypeAliases(substitutions.get(name)!, seen);
              }
              return this.expandTypeAliases(expr, seen);
            }
            case "GenericTypeExpression":
              return {
                ...expr,
                base: substitute(expr.base),
                arguments: (expr.arguments ?? []).map((arg) => (arg ? substitute(arg) : arg)),
              };
            case "NullableTypeExpression":
              return { ...expr, innerType: substitute(expr.innerType) };
            case "ResultTypeExpression":
              return { ...expr, innerType: substitute(expr.innerType) };
            case "UnionTypeExpression":
              return { ...expr, members: (expr.members ?? []).map((member) => substitute(member)) };
            case "FunctionTypeExpression":
              return {
                ...expr,
                paramTypes: (expr.paramTypes ?? []).map((param) => substitute(param)),
                returnType: substitute(expr.returnType),
              };
            default:
              return expr;
          }
        };
        seen.add(baseName);
        const substituted = substitute(alias.targetType);
        const expanded = this.expandTypeAliases(substituted, seen);
        seen.delete(baseName);
        return expanded ?? substituted;
      }
      case "NullableTypeExpression":
        return { ...t, innerType: this.expandTypeAliases(t.innerType, seen) };
      case "ResultTypeExpression":
        return { ...t, innerType: this.expandTypeAliases(t.innerType, seen) };
      case "UnionTypeExpression":
        return { ...t, members: (t.members ?? []).map((member) => this.expandTypeAliases(member, seen)) };
      case "FunctionTypeExpression":
        return {
          ...t,
          paramTypes: (t.paramTypes ?? []).map((param) => this.expandTypeAliases(param, seen)),
          returnType: this.expandTypeAliases(t.returnType, seen),
        };
      default:
        return t;
    }
  };

  cls.prototype.typeExpressionToString = function typeExpressionToString(this: Interpreter, t: AST.TypeExpression): string {
    const canonical = this.expandTypeAliases(t);
    switch (canonical.type) {
      case "SimpleTypeExpression":
        return canonical.name.name;
      case "GenericTypeExpression":
        return `${this.typeExpressionToString(canonical.base)}<${(canonical.arguments ?? []).map(arg => this.typeExpressionToString(arg)).join(",")}>`;
      case "NullableTypeExpression":
        return `${this.typeExpressionToString(canonical.innerType)}?`;
      case "ResultTypeExpression":
        return `Result<${this.typeExpressionToString(canonical.innerType)}>`;
      case "FunctionTypeExpression":
        return `(${canonical.paramTypes.map(pt => this.typeExpressionToString(pt)).join(", ")}) -> ${this.typeExpressionToString(canonical.returnType)}`;
      case "UnionTypeExpression":
        return canonical.members.map(member => this.typeExpressionToString(member)).join(" | ");
      case "WildcardTypeExpression":
        return "_";
      default:
        return "<type>";
    }
  };

  cls.prototype.parseTypeExpression = function parseTypeExpression(this: Interpreter, t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null {
    const canonical = this.expandTypeAliases(t);
    if (canonical.type === "SimpleTypeExpression") {
      return { name: canonical.name.name, typeArgs: [] };
    }
    if (canonical.type === "GenericTypeExpression" && canonical.base.type === "SimpleTypeExpression") {
      return { name: canonical.base.name.name, typeArgs: canonical.arguments ?? [] };
    }
    return null;
  };

  cls.prototype.cloneTypeExpression = function cloneTypeExpression(this: Interpreter, t: AST.TypeExpression): AST.TypeExpression {
    switch (t.type) {
      case "SimpleTypeExpression":
        return { type: "SimpleTypeExpression", name: AST.identifier(t.name.name) };
      case "GenericTypeExpression":
        return {
          type: "GenericTypeExpression",
          base: this.cloneTypeExpression(t.base),
          arguments: (t.arguments ?? []).map(arg => this.cloneTypeExpression(arg)),
        };
      case "FunctionTypeExpression":
        return {
          type: "FunctionTypeExpression",
          paramTypes: t.paramTypes.map(pt => this.cloneTypeExpression(pt)),
          returnType: this.cloneTypeExpression(t.returnType),
        };
      case "NullableTypeExpression":
        return { type: "NullableTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "ResultTypeExpression":
        return { type: "ResultTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "UnionTypeExpression":
        return { type: "UnionTypeExpression", members: t.members.map(member => this.cloneTypeExpression(member)) };
      case "WildcardTypeExpression":
      default:
        return { type: "WildcardTypeExpression" };
    }
  };

  cls.prototype.typeExpressionsEqual = function typeExpressionsEqual(this: Interpreter, a: AST.TypeExpression, b: AST.TypeExpression): boolean {
    return this.typeExpressionToString(a) === this.typeExpressionToString(b);
  };

  cls.prototype.typeExpressionFromValue = function typeExpressionFromValue(this: Interpreter, value: RuntimeValue): AST.TypeExpression | null {
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
                const actual = this.typeExpressionFromValue(actualVal);
                if (!actual) continue;
                this.matchTypeExpressionTemplate(field.fieldType, actual, genericNames, bindings);
              }
            } else {
              for (const field of value.def.fields) {
                if (!field?.name || !field.fieldType) continue;
                const actualVal = value.values.get(field.name.name);
                if (!actualVal) continue;
                const actual = this.typeExpressionFromValue(actualVal);
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
          const inferred = this.typeExpressionFromValue(el);
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
  };

  cls.prototype.matchesType = function matchesType(this: Interpreter, t: AST.TypeExpression, v: RuntimeValue): boolean {
    const target = this.expandTypeAliases(t);
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
        if (name === "IteratorEnd" && v.kind === "iterator_end") return true;
        if (name === "Iterator" && v.kind === "iterator") return true;
        if (name === "Awaitable" && isAwaitableStructInstance(v)) return true;
        if (this.interfaces.has(name)) {
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
          const typeArgs = v.kind === "struct_instance" ? v.typeArguments : undefined;
          return this.typeImplementsInterface(typeName, name, typeArgs);
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
              const typeArgs = v.kind === "struct_instance" ? v.typeArguments : undefined;
              if (valueTypeName && this.typeImplementsInterface(valueTypeName, baseName, typeArgs)) {
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
      case "interface_value":
        return this.getTypeNameForValue(value.value);
      case "error":
        return "Error";
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

  cls.prototype.typeImplementsInterface = function typeImplementsInterface(this: Interpreter, typeName: string, interfaceName: string, typeArgs?: AST.TypeExpression[]): boolean {
    const interfaceInfoFromExpr = (expr: AST.TypeExpression | undefined): { name: string; args?: AST.TypeExpression[] } | null => {
      if (!expr) return null;
      if (expr.type === "SimpleTypeExpression") {
        return { name: expr.name.name };
      }
      if (expr.type === "GenericTypeExpression" && expr.base.type === "SimpleTypeExpression") {
        return { name: expr.base.name.name, args: expr.arguments ?? [] };
      }
      return null;
    };

    const directImplements = (name: string, ifaceName: string, args?: AST.TypeExpression[]): boolean => {
      if (ifaceName === "Error" && name === "Error") {
        return true;
      }
      const base: AST.SimpleTypeExpression = { type: "SimpleTypeExpression", name: AST.identifier(name) };
      const subjectType: AST.TypeExpression = args && args.length > 0 ? { type: "GenericTypeExpression", base, arguments: args } : base;
      const entries = [
        ...(this.implMethods.get(name) ?? []),
        ...this.genericImplMethods,
      ];
      if (entries.length === 0) return false;
      for (const entry of entries) {
        if (entry.def.interfaceName.name !== ifaceName) continue;
        if (this.matchImplEntry(entry, { subjectType, typeArgs: args })) return true;
      }
      return false;
    };

    const check = (name: string, ifaceName: string, args: AST.TypeExpression[] | undefined, visited: Set<string>): boolean => {
      const argKey = (args ?? []).map((arg) => this.typeExpressionToString(arg)).join("|");
      const key = `${name}::${ifaceName}::${argKey}`;
      if (visited.has(key)) return true;
      visited.add(key);
      const ifaceDef = this.interfaces.get(ifaceName);
      if (ifaceDef?.baseInterfaces && ifaceDef.baseInterfaces.length > 0) {
        for (const base of ifaceDef.baseInterfaces) {
          const info = interfaceInfoFromExpr(base);
          if (!info) return false;
          if (!check(name, info.name, info.args, visited)) return false;
        }
        if (!ifaceDef.signatures || ifaceDef.signatures.length === 0) {
          return true;
        }
      }
      return directImplements(name, ifaceName, args);
    };

    return check(typeName, interfaceName, typeArgs, new Set());
  };

  cls.prototype.coerceValueToType = function coerceValueToType(this: Interpreter, typeExpr: AST.TypeExpression | undefined, value: RuntimeValue): RuntimeValue {
    if (!typeExpr) return value;
    const canonical = this.expandTypeAliases(typeExpr);
    if (canonical.type === "SimpleTypeExpression") {
      const name = canonical.name.name;
      if (INTEGER_KIND_SET.has(name as IntegerKind) && isIntegerValue(value)) {
        const targetKind = name as IntegerKind;
        const actualKind = value.kind as IntegerKind;
        if (actualKind !== targetKind && integerValueWithinRange(value.value, targetKind)) {
          return makeIntegerValue(targetKind, value.value);
        }
      }
      if (this.interfaces.has(name)) {
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
    return value;
  };

  cls.prototype.castValueToType = function castValueToType(this: Interpreter, typeExpr: AST.TypeExpression, value: RuntimeValue): RuntimeValue {
    const canonical = this.expandTypeAliases(typeExpr);
    if (this.matchesType(canonical, value)) return value;
    const rawValue = value.kind === "interface_value" ? value.value : value;
    if (canonical.type === "SimpleTypeExpression") {
      const name = canonical.name.name;
      if (INTEGER_KIND_SET.has(name as IntegerKind)) {
        const targetKind = name as IntegerKind;
        if (isIntegerValue(rawValue)) {
          if (!integerValueWithinRange(rawValue.value, targetKind)) {
            throw new Error(`value out of range for ${targetKind}`);
          }
          return makeIntegerValue(targetKind, rawValue.value);
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
      if (this.interfaces.has(name)) {
        return this.toInterfaceValue(name, value);
      }
    }
    throw new Error(`cannot cast ${this.getTypeNameForValue(rawValue) ?? rawValue.kind} to ${this.typeExpressionToString(canonical)}`);
  };

  cls.prototype.toInterfaceValue = function toInterfaceValue(this: Interpreter, interfaceName: string, rawValue: RuntimeValue): RuntimeValue {
  if (!this.interfaces.has(interfaceName)) {
    throw new Error(`Unknown interface '${interfaceName}'`);
  }
  if (rawValue.kind === "interface_value") {
      if (rawValue.interfaceName === interfaceName) return rawValue;
      return this.toInterfaceValue(interfaceName, rawValue.value);
    }
    const typeName = this.getTypeNameForValue(rawValue);
  if (!typeName || !this.typeImplementsInterface(typeName, interfaceName, rawValue.kind === "struct_instance" ? rawValue.typeArguments : undefined)) {
    throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
  }
  let typeArguments: AST.TypeExpression[] | undefined;
  let typeArgMap: Map<string, AST.TypeExpression> | undefined;
  if (rawValue.kind === "struct_instance") {
    typeArguments = rawValue.typeArguments;
    typeArgMap = rawValue.typeArgMap;
  }
  const resolution = this.resolveInterfaceImplementation(typeName, interfaceName, {
    typeArgs: typeArguments,
    typeArgMap,
  });
  if (!resolution.ok) {
    if (resolution.error) {
      throw resolution.error;
    }
    throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
  }
  return { kind: "interface_value", interfaceName, value: rawValue, typeArguments, typeArgMap };
};
}
