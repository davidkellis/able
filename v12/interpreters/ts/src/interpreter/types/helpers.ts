import * as AST from "../../ast";
import type { Interpreter } from "../index";
import { getIntegerInfo } from "../numeric";
import type { FloatKind, IntegerKind, RuntimeValue } from "../values";

export const INTEGER_KINDS: IntegerKind[] = ["i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128"];
export const FLOAT_KINDS: FloatKind[] = ["f32", "f64"];
export const INTEGER_KIND_SET: Set<IntegerKind> = new Set(INTEGER_KINDS);

export function isSingletonStructDef(def: AST.StructDefinition): boolean {
  if (!def || (def.genericParams && def.genericParams.length > 0)) return false;
  if (def.kind === "singleton") return true;
  return def.kind === "named" && def.fields.length === 0;
}

export function integerRangeWithin(source: IntegerKind, target: IntegerKind): boolean {
  const sourceInfo = getIntegerInfo(source);
  const targetInfo = getIntegerInfo(target);
  return sourceInfo.min >= targetInfo.min && sourceInfo.max <= targetInfo.max;
}

export function integerValueWithinRange(raw: bigint, target: IntegerKind): boolean {
  const targetInfo = getIntegerInfo(target);
  return raw >= targetInfo.min && raw <= targetInfo.max;
}

export function normalizeKernelAliasName(name: string): string {
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

export function hasConcreteTypeName(ctx: Interpreter, name: string): boolean {
  return ctx.structs.has(name) || ctx.unions.has(name);
}

export type TypeImplementsInterfaceOptions = {
  subjectTypeArgs?: AST.TypeExpression[];
  interfaceArgs?: AST.TypeExpression[];
  subjectType?: AST.TypeExpression;
};

export function normalizeTypeImplementsInterfaceOptions(
  opts?: AST.TypeExpression[] | TypeImplementsInterfaceOptions,
): TypeImplementsInterfaceOptions {
  if (!opts) return {};
  if (Array.isArray(opts)) return { subjectTypeArgs: opts };
  return opts;
}

export function isErrorValue(ctx: Interpreter, value: RuntimeValue): boolean {
  if (value.kind === "error") return true;
  if (value.kind === "interface_value" && value.interfaceName === "Error") return true;
  const typeName = ctx.getTypeNameForValue(value);
  if (!typeName) return false;
  const typeArgs = value.kind === "struct_instance" ? value.typeArguments : undefined;
  return ctx.typeImplementsInterface(typeName, "Error", typeArgs);
}

export function isAwaitableStructInstance(value: RuntimeValue): boolean {
  if (value.kind !== "struct_instance") return false;
  const values = value.values as Map<string, RuntimeValue> | undefined;
  if (!(values instanceof Map)) return false;
  return values.has("is_ready") && values.has("register") && values.has("commit");
}

export function collectInterfaceDispatchSets(ctx: Interpreter, interfaceName: string): {
  baseInterfaces: Set<string>;
  implInterfaces: Set<string>;
} {
  const baseInterfaces = new Set<string>();
  const implInterfaces = new Set<string>();
  const addExpanded = (name: string, target: Set<string>): void => {
    if (!name) return;
    const expanded = ctx.collectInterfaceConstraintExpressions(AST.simpleTypeExpression(name));
    for (const expr of expanded) {
      const info = ctx.parseTypeExpression(expr);
      if (info?.name) {
        target.add(info.name);
      }
    }
  };
  addExpanded(interfaceName, baseInterfaces);
  const implEntries = ctx.implMethods.get(interfaceName) ?? [];
  for (const entry of implEntries) {
    const name = entry?.def?.interfaceName?.name;
    if (!name) continue;
    addExpanded(name, implInterfaces);
  }
  for (const name of baseInterfaces) {
    implInterfaces.delete(name);
  }
  return { baseInterfaces, implInterfaces };
}

export function buildInterfaceMethodDictionary(
  ctx: Interpreter,
  interfaceName: string,
  interfaceArgs: AST.TypeExpression[] | undefined,
  rawValue: RuntimeValue,
  typeName: string,
  typeArguments?: AST.TypeExpression[],
  typeArgMap?: Map<string, AST.TypeExpression>,
): Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>> {
  const dictionary = new Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>>();
  const ifaceTypeArgs = interfaceArgs && interfaceArgs.length > 0 ? interfaceArgs : typeArguments;
  const baseInterfaceArgs = new Map<string, AST.TypeExpression[]>();
  const rootIface = ctx.interfaces.get(interfaceName);
  if (rootIface?.baseInterfaces) {
    for (const base of rootIface.baseInterfaces) {
      const info = ctx.parseTypeExpression(base);
      if (info?.name) {
        baseInterfaceArgs.set(info.name, info.typeArgs ?? []);
      }
    }
  }
  const dispatchSets = collectInterfaceDispatchSets(ctx, interfaceName);
  const collect = (
    ifaceName: string,
    targetTypeName: string,
    targetTypeArgs?: AST.TypeExpression[],
    targetTypeArgMap?: Map<string, AST.TypeExpression>,
    ifaceArgs?: AST.TypeExpression[],
  ): void => {
    const ifaceDef = ctx.interfaces.get(ifaceName);
    if (!ifaceDef || !Array.isArray(ifaceDef.signatures)) return;
    const ifaceEnv = ctx.interfaceEnvs.get(ifaceName) ?? ctx.globals;
    const targetTypeExpr = targetTypeArgs && targetTypeArgs.length > 0
      ? AST.genericTypeExpression(AST.simpleTypeExpression(targetTypeName), targetTypeArgs)
      : AST.simpleTypeExpression(targetTypeName);
    for (const sig of ifaceDef.signatures) {
      const methodName = sig?.name?.name;
      if (!methodName || dictionary.has(methodName)) continue;
      const method = ctx.findMethod(targetTypeName, methodName, {
        typeArgs: targetTypeArgs,
        typeArgMap: targetTypeArgMap,
        interfaceName: ifaceName,
        interfaceArgs: ifaceArgs,
        includeInherent: false,
      });
      if (method) {
        dictionary.set(methodName, method);
        continue;
      }
      if (ifaceName === "Iterator" && rawValue.kind === "iterator") {
        ctx.ensureIteratorBuiltins();
        const native = ctx.iteratorNativeMethods?.[methodName as "next" | "close"];
        if (native) {
          dictionary.set(methodName, native);
          continue;
        }
      }
      if (sig.defaultImpl) {
        const typeBindings = ctx.buildSelfTypePatternBindings(ifaceDef, targetTypeExpr);
        const defaultFunc = ctx.createDefaultMethodFunction(sig, ifaceEnv, targetTypeExpr, typeBindings);
        if (defaultFunc) {
          dictionary.set(methodName, defaultFunc);
          continue;
        }
      }
      if (ifaceName === interfaceName) {
        throw new Error(`No method '${methodName}' for interface ${interfaceName}`);
      }
    }
  };
  for (const ifaceName of dispatchSets.baseInterfaces) {
    const args = ifaceName === interfaceName ? interfaceArgs : baseInterfaceArgs.get(ifaceName);
    collect(ifaceName, typeName, typeArguments, typeArgMap, args);
  }
  for (const ifaceName of dispatchSets.implInterfaces) {
    collect(ifaceName, interfaceName, ifaceTypeArgs);
  }
  return dictionary;
}
