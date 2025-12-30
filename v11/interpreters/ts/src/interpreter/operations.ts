import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { collectTypeDispatches, type TypeDispatch } from "./type-dispatch";
import { callCallableValue } from "./functions";
import {
  applyArithmeticBinary,
  applyBitwiseBinary,
  applyBitwiseNot,
  applyComparisonBinary,
  applyNumericUnaryMinus,
  isIntegerValue,
  isNumericValue,
  numericToNumber,
  makeIntegerValue,
} from "./numeric";
import { makeIntegerFromNumber } from "./numeric";
import { valuesEqual } from "./value_equals";

export function resolveIndexFunction(
  ctx: Interpreter,
  receiver: RuntimeValue,
  methodName: string,
  interfaceName: string,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const dispatches = collectTypeDispatches(ctx, receiver);
  const allDispatches: TypeDispatch[] = [...dispatches];
  const seen = new Set(dispatches.map((entry) => `${entry.typeName}::${entry.typeArgs.map((arg) => ctx.typeExpressionToString(arg)).join("|")}`));
  for (const dispatch of dispatches) {
    const base = AST.simpleTypeExpression(dispatch.typeName);
    const expr = dispatch.typeArgs.length > 0
      ? AST.genericTypeExpression(base, dispatch.typeArgs)
      : base;
    const expanded = ctx.expandTypeAliases(expr);
    const expandedInfo = ctx.parseTypeExpression(expanded);
    if (expandedInfo) {
      const key = `${expandedInfo.name}::${expandedInfo.typeArgs.map((arg) => ctx.typeExpressionToString(arg)).join("|")}`;
      if (!seen.has(key)) {
        seen.add(key);
        allDispatches.push({ typeName: expandedInfo.name, typeArgs: expandedInfo.typeArgs });
      }
    }
  }
  for (const dispatch of allDispatches) {
    const method = ctx.findMethod(dispatch.typeName, methodName, {
      typeArgs: dispatch.typeArgs,
      interfaceName,
      includeInherent: false,
    });
    if (method) return method;
  }
  return null;
}

function resolveIndexErrorStruct(ctx: Interpreter): AST.StructDefinition {
  const name = "IndexError";
  const cached = ctx.standardErrorStructs.get(name);
  if (cached) return cached;
  const candidates = [
    name,
    `able.core.errors.${name}`,
    `core.errors.${name}`,
    `errors.${name}`,
  ];
  for (const candidate of candidates) {
    try {
      const val = ctx.globals.get(candidate);
      if (val && val.kind === "struct_def") {
        ctx.standardErrorStructs.set(name, val.def);
        return val.def;
      }
    } catch {
      // ignore lookup errors
    }
  }
  for (const bucket of ctx.packageRegistry.values()) {
    const val = bucket.get(name);
    if (val && val.kind === "struct_def") {
      ctx.standardErrorStructs.set(name, val.def);
      return val.def;
    }
  }
  const placeholder = AST.structDefinition(name, [], "named");
  ctx.standardErrorStructs.set(name, placeholder);
  return placeholder;
}

export function makeIndexErrorInstance(ctx: Interpreter, index: number, length: number): RuntimeValue {
  const def = resolveIndexErrorStruct(ctx);
  return ctx.makeNamedStructInstance(def, [
    ["index", makeIntegerValue("i64", BigInt(index))],
    ["length", makeIntegerValue("i64", BigInt(length))],
  ]);
}

type OperatorDispatch = { interfaceName: string; methodName: string };

const OPERATOR_INTERFACES: Record<string, OperatorDispatch> = {
  "+": { interfaceName: "Add", methodName: "add" },
  "-": { interfaceName: "Sub", methodName: "sub" },
  "*": { interfaceName: "Mul", methodName: "mul" },
  "/": { interfaceName: "Div", methodName: "div" },
  "%": { interfaceName: "Rem", methodName: "rem" },
};

const UNARY_INTERFACES: Record<string, OperatorDispatch> = {
  "-": { interfaceName: "Neg", methodName: "neg" },
  ".~": { interfaceName: "Not", methodName: "not" },
};

const EQUALITY_INTERFACES: OperatorDispatch[] = [
  { interfaceName: "Eq", methodName: "eq" },
  { interfaceName: "PartialEq", methodName: "eq" },
];

const ORDERING_INTERFACES: OperatorDispatch[] = [
  { interfaceName: "Ord", methodName: "cmp" },
  { interfaceName: "PartialOrd", methodName: "partial_cmp" },
];

function resolveOperatorFunction(
  ctx: Interpreter,
  receiver: RuntimeValue,
  op: string,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const dispatch = OPERATOR_INTERFACES[op];
  if (!dispatch) return null;
  const dispatches = collectTypeDispatches(ctx, receiver);
  for (const entry of dispatches) {
    const method = ctx.findMethod(entry.typeName, dispatch.methodName, {
      typeArgs: entry.typeArgs,
      interfaceName: dispatch.interfaceName,
      includeInherent: false,
    });
    if (method) return method;
  }
  return null;
}

function resolveComparisonFunction(
  ctx: Interpreter,
  receiver: RuntimeValue,
  dispatch: OperatorDispatch,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const dispatches = collectTypeDispatches(ctx, receiver);
  for (const entry of dispatches) {
    const method = ctx.findMethod(entry.typeName, dispatch.methodName, {
      typeArgs: entry.typeArgs,
      interfaceName: dispatch.interfaceName,
      includeInherent: false,
    });
    if (method) return method;
  }
  return null;
}

function resolveUnaryFunction(
  ctx: Interpreter,
  receiver: RuntimeValue,
  dispatch: OperatorDispatch,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const dispatches = collectTypeDispatches(ctx, receiver);
  for (const entry of dispatches) {
    const method = ctx.findMethod(entry.typeName, dispatch.methodName, {
      typeArgs: entry.typeArgs,
      interfaceName: dispatch.interfaceName,
      includeInherent: false,
    });
    if (method) return method;
  }
  return null;
}

function orderingToCmp(ctx: Interpreter, value: RuntimeValue): number | null {
  const typeName = ctx.getTypeNameForValue(value);
  switch (typeName) {
    case "Less":
      return -1;
    case "Equal":
      return 0;
    case "Greater":
      return 1;
    default:
      return null;
  }
}

function compareFromCmp(op: string, cmp: number): boolean {
  switch (op) {
    case "<":
      return cmp < 0;
    case "<=":
      return cmp <= 0;
    case ">":
      return cmp > 0;
    case ">=":
      return cmp >= 0;
    default:
      return false;
  }
}

declare module "./index" {
  interface Interpreter {
    computeBinaryForCompound(op: string, left: RuntimeValue, right: RuntimeValue): RuntimeValue;
    ensureDivModStruct(): AST.StructDefinition;
    ensureRatioStruct(): AST.StructDefinition;
    divModStruct?: AST.StructDefinition;
    ratioStruct?: AST.StructDefinition;
  }
}

export function evaluateUnaryExpression(ctx: Interpreter, node: AST.UnaryExpression, env: Environment): RuntimeValue {
  const v = ctx.evaluate(node.operand, env);
  if (node.operator === "-") {
    if (isNumericValue(v)) {
      return applyNumericUnaryMinus(v);
    }
    const method = resolveUnaryFunction(ctx, v, UNARY_INTERFACES["-"]);
    if (method) {
      return callCallableValue(ctx, method, [v], env);
    }
    return applyNumericUnaryMinus(v);
  }
  if (node.operator === "!") {
    return { kind: "bool", value: !ctx.isTruthy(v) };
  }
  if (node.operator === ".~") {
    const dispatch = UNARY_INTERFACES[".~"];
    if (dispatch) {
      const method = resolveUnaryFunction(ctx, v, dispatch);
      if (method) {
        return callCallableValue(ctx, method, [v], env);
      }
    }
    return applyBitwiseNot(v);
  }
  throw new Error(`Unknown unary operator ${node.operator}`);
}

export function evaluateBinaryExpression(ctx: Interpreter, node: AST.BinaryExpression, env: Environment): RuntimeValue {
  const b = node;
  if (b.operator === "&&" || b.operator === "||") {
    const lv = ctx.evaluate(b.left, env);
    if (b.operator === "&&") {
      if (!ctx.isTruthy(lv)) return lv;
      const rv = ctx.evaluate(b.right, env);
      return rv;
    }
    if (ctx.isTruthy(lv)) return lv;
    const rv = ctx.evaluate(b.right, env);
    return rv;
  }

  if (b.operator === "|>" || b.operator === "|>>") {
    const subject = ctx.evaluate(b.left, env);
    ctx.implicitReceiverStack.push(subject);
    try {
      const placeholderCallable = ctx.tryBuildPlaceholderFunction(b.right, env);
      if (placeholderCallable) {
        try {
          return callCallableValue(ctx, placeholderCallable, [subject], env);
        } catch (err) {
          const message = err instanceof Error ? err.message : String(err);
          throw new Error(`pipe RHS must be callable: ${message}`);
        }
      }
      if (b.right.type === "FunctionCall") {
        const calleeVal = ctx.evaluate(b.right.callee, env);
        const evaluatedArgs = b.right.arguments.map((arg) => ctx.evaluate(arg, env));
        const callArgs =
          calleeVal.kind === "bound_method" || calleeVal.kind === "native_bound_method"
            ? evaluatedArgs
            : [subject, ...evaluatedArgs];
        try {
          return callCallableValue(ctx, calleeVal, callArgs, env);
        } catch (err) {
          const message = err instanceof Error ? err.message : String(err);
          throw new Error(`pipe RHS must be callable: ${message}`);
        }
      }
      const rhsVal = ctx.evaluate(b.right, env);
      const callArgs = (rhsVal.kind === "bound_method" || rhsVal.kind === "native_bound_method") ? [] : [subject];
      try {
        return callCallableValue(ctx, rhsVal, callArgs, env);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        throw new Error(`pipe RHS must be callable: ${message}`);
      }
    } finally {
      ctx.implicitReceiverStack.pop();
    }
  }

  const left = ctx.evaluate(b.left, env);
  const right = ctx.evaluate(b.right, env);

  if (b.operator === "+" && left.kind === "String" && right.kind === "String") {
    return { kind: "String", value: left.value + right.value };
  }

  if (["+","-","*","/","//","%","/%","^"].includes(b.operator)) {
    if (isNumericValue(left) && isNumericValue(right)) {
      return applyArithmeticBinary(b.operator, left, right, {
        makeDivMod: (kind, parts) => {
          const structDef = ctx.ensureDivModStruct();
          const typeArg = AST.simpleTypeExpression(kind);
          const typeArgMap = ctx.mapTypeArguments(structDef.genericParams ?? [], [typeArg], "DivMod");
          return {
            kind: "struct_instance",
            def: structDef,
            values: new Map([
              ["quotient", parts.quotient],
              ["remainder", parts.remainder],
            ]),
            typeArguments: [typeArg],
            typeArgMap,
          };
        },
        makeRatio: (parts) => {
          const structDef = ctx.ensureRatioStruct();
          return {
            kind: "struct_instance",
            def: structDef,
            values: new Map([
              ["num", makeIntegerValue("i64", parts.num)],
              ["den", makeIntegerValue("i64", parts.den)],
            ]),
          };
        },
      });
    }
    const method = resolveOperatorFunction(ctx, left, b.operator);
    if (method) {
      return callCallableValue(ctx, method, [left, right], env);
    }
    return applyArithmeticBinary(b.operator, left, right, {
      makeDivMod: (kind, parts) => {
        const structDef = ctx.ensureDivModStruct();
        const typeArg = AST.simpleTypeExpression(kind);
        const typeArgMap = ctx.mapTypeArguments(structDef.genericParams ?? [], [typeArg], "DivMod");
        return {
          kind: "struct_instance",
          def: structDef,
          values: new Map([
            ["quotient", parts.quotient],
            ["remainder", parts.remainder],
          ]),
          typeArguments: [typeArg],
          typeArgMap,
        };
      },
      makeRatio: (parts) => {
        const structDef = ctx.ensureRatioStruct();
        return {
          kind: "struct_instance",
          def: structDef,
          values: new Map([
            ["num", makeIntegerValue("i64", parts.num)],
            ["den", makeIntegerValue("i64", parts.den)],
          ]),
        };
      },
    });
  }

  if ([">","<",">=","<=","==","!="].includes(b.operator)) {
    if (isNumericValue(left) && isNumericValue(right)) {
      return applyComparisonBinary(b.operator, left, right);
    }
    if (left.kind === "String" && right.kind === "String") {
      switch (b.operator) {
        case ">": return { kind: "bool", value: left.value > right.value };
        case "<": return { kind: "bool", value: left.value < right.value };
        case ">=": return { kind: "bool", value: left.value >= right.value };
        case "<=": return { kind: "bool", value: left.value <= right.value };
        case "==": return { kind: "bool", value: left.value === right.value };
        case "!=": return { kind: "bool", value: left.value !== right.value };
      }
    }
    if (b.operator === "==" || b.operator === "!=") {
      for (const dispatch of EQUALITY_INTERFACES) {
        const method = resolveComparisonFunction(ctx, left, dispatch);
        if (!method) continue;
        const result = callCallableValue(ctx, method, [left, right], env);
        if (result.kind !== "bool") {
          throw new Error(`Comparison '${b.operator}' requires a bool result from ${dispatch.interfaceName}.${dispatch.methodName}`);
        }
        return { kind: "bool", value: b.operator === "!=" ? !result.value : result.value };
      }
      if (b.operator === "==") return { kind: "bool", value: valuesEqual(left, right) };
      return { kind: "bool", value: !valuesEqual(left, right) };
    }
    for (const dispatch of ORDERING_INTERFACES) {
      const method = resolveComparisonFunction(ctx, left, dispatch);
      if (!method) continue;
      const result = callCallableValue(ctx, method, [left, right], env);
      const cmp = orderingToCmp(ctx, result);
      if (cmp === null) {
        throw new Error(`Comparison '${b.operator}' requires Ordering result from ${dispatch.interfaceName}.${dispatch.methodName}`);
      }
      return { kind: "bool", value: compareFromCmp(b.operator, cmp) };
    }
    throw new Error("Unsupported comparison operands");
  }

  if ([".&",".|",".^",".<<",".>>"].includes(b.operator)) {
    return applyBitwiseBinary(b.operator, left, right);
  }

  throw new Error(`Unknown binary operator ${b.operator}`);
}

export function evaluateRangeExpression(ctx: Interpreter, node: AST.RangeExpression, env: Environment): RuntimeValue {
  const start = ctx.evaluate(node.start, env);
  const end = ctx.evaluate(node.end, env);
  const viaInterface = ctx.tryInvokeRangeImplementation(start, end, node.inclusive, env);
  if (viaInterface) {
    return viaInterface;
  }
  if (!isIntegerValue(start) || !isIntegerValue(end)) {
    throw new Error("Range boundaries must be numeric");
  }
  let startNum: number;
  let endNum: number;
  try {
    startNum = numericToNumber(start, "Range start");
    endNum = numericToNumber(end, "Range end");
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    if (message.includes("Range start must be numeric") || message.includes("Range end must be numeric")) {
      throw new Error("Range boundaries must be numeric");
    }
    throw err;
  }
  if (!Number.isFinite(startNum) || !Number.isFinite(endNum)) {
    throw new Error("Range endpoint must be finite");
  }
  const startInt = Math.trunc(startNum);
  const endInt = Math.trunc(endNum);
  const step = startInt <= endInt ? 1 : -1;
  const elements: RuntimeValue[] = [];
  for (let current = startInt; ; current += step) {
    if (step > 0) {
      if (node.inclusive) {
        if (current > endInt) break;
      } else if (current >= endInt) {
        break;
      }
    } else if (node.inclusive) {
      if (current < endInt) break;
    } else if (current <= endInt) {
      break;
    }
    elements.push(makeIntegerFromNumber("i32", current));
  }
  return ctx.makeArrayValue(elements);
}

export function evaluateIndexExpression(ctx: Interpreter, node: AST.IndexExpression, env: Environment): RuntimeValue {
  const obj = ctx.evaluate(node.object, env);
  const idxVal = ctx.evaluate(node.index, env);
  const viaInterface = resolveIndexFunction(ctx, obj, "get", "Index");
  if (viaInterface) {
    return callCallableValue(ctx, viaInterface, [obj, idxVal], env);
  }
  if (obj.kind !== "array") throw new Error("Indexing is only supported on arrays in this milestone");
  const state = ctx.ensureArrayState(obj);
  const idx = Math.trunc(numericToNumber(idxVal, "Array index", { requireSafeInteger: true }));
  if (idx < 0 || idx >= state.values.length) return makeIndexErrorInstance(ctx, idx, state.values.length);
  const el = state.values[idx];
  if (el === undefined) throw new Error("Internal error: array element undefined");
  return el;
}

export function applyOperationsAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureDivModStruct = function ensureDivModStruct(this: Interpreter): AST.StructDefinition {
    if (this.divModStruct) return this.divModStruct;
    const divModDef = AST.structDefinition(
      "DivMod",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("T"), "quotient"),
        AST.structFieldDefinition(AST.simpleTypeExpression("T"), "remainder"),
      ],
      "named",
      [AST.genericParameter("T")],
    );
    this.evaluate(divModDef, this.globals);
    const resolved = this.globals.get("DivMod");
    if (resolved && resolved.kind === "struct_def") {
      this.divModStruct = resolved.def;
    } else {
      this.divModStruct = divModDef;
    }
    return this.divModStruct;
  };

  cls.prototype.ensureRatioStruct = function ensureRatioStruct(this: Interpreter): AST.StructDefinition {
    if (this.ratioStruct) return this.ratioStruct;
    try {
      const existing = this.globals.get("Ratio");
      if (existing?.kind === "struct_def") {
        this.ratioStruct = existing.def;
        return this.ratioStruct;
      }
    } catch {}
    const ratioDef = AST.structDefinition(
      "Ratio",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "num"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "den"),
      ],
      "named",
    );
    this.evaluate(ratioDef, this.globals);
    const resolved = this.globals.get("Ratio");
    if (resolved && resolved.kind === "struct_def") {
      this.ratioStruct = resolved.def;
    } else {
      this.ratioStruct = ratioDef;
    }
    return this.ratioStruct;
  };

  cls.prototype.computeBinaryForCompound = function computeBinaryForCompound(this: Interpreter, op: string, left: RuntimeValue, right: RuntimeValue): RuntimeValue {
    if (["+","-","*","/","%"].includes(op)) {
      return applyArithmeticBinary(op, left, right, {
        makeRatio: (parts) => {
          const structDef = this.ensureRatioStruct();
          return {
            kind: "struct_instance",
            def: structDef,
            values: new Map([
              ["num", makeIntegerValue("i64", parts.num)],
              ["den", makeIntegerValue("i64", parts.den)],
            ]),
          };
        },
      });
    }

    if ([".&",".|",".^",".<<",".>>"].includes(op)) {
      return applyBitwiseBinary(op, left, right);
    }

    throw new Error(`Unsupported compound operator ${op}`);
  };
}
