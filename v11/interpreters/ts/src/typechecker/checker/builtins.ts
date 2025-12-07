import { arrayType, primitiveType, unknownType, type TypeInfo } from "../types";
import {
  functionSignature,
  functionParameter,
  genericParameter,
  interfaceDefinition,
  nullableTypeExpression,
  simpleTypeExpression,
  structDefinition,
  structFieldDefinition,
} from "../../ast";
import type { Environment } from "../environment";
import type { FunctionInfo } from "./types";

type BuiltinContext = {
  env: Environment;
  functionInfos: Map<string, FunctionInfo[]>;
  implementationContext?: unknown;
  registerStructDefinition(definition: unknown): void;
  registerTypeAlias(definition: unknown): void;
  registerInterfaceDefinition(definition: unknown): void;
  collectImplementationDefinition(definition: unknown): void;
  collectMethodsDefinition(definition: unknown): void;
};

export function installBuiltins(context: BuiltinContext): void {
  const voidType = primitiveType("void");
  const boolType = primitiveType("bool");
  const i32Type = primitiveType("i32");
  const i64Type = primitiveType("i64");
  const u64Type = primitiveType("u64");
  const u8Type = primitiveType("u8");
  const stringType = primitiveType("string");
  const charType = primitiveType("char");
  const anyType = unknownType;
  const unknown = unknownType;

  const register = (name: string, params: TypeInfo[], returnType: TypeInfo) => {
    registerBuiltinFunction(context.env, context.functionInfos, name, params, returnType);
  };

  register("print", [unknown], voidType);
  register("proc_yield", [], voidType);
  register("proc_cancelled", [], boolType);
  register("proc_flush", [], voidType);
  register("proc_pending_tasks", [], i32Type);

  register("__able_channel_new", [i32Type], i64Type);
  register("__able_channel_send", [unknown, unknown], voidType);
  register("__able_channel_receive", [unknown], unknown);
  register("__able_channel_try_send", [unknown, unknown], boolType);
  register("__able_channel_try_receive", [unknown], unknown);
  register("__able_channel_await_try_recv", [unknown, unknown], unknown);
  register("__able_channel_await_try_send", [unknown, unknown, unknown], unknown);
  register("__able_channel_close", [unknown], voidType);
  register("__able_channel_is_closed", [unknown], boolType);

  register("__able_mutex_new", [], i64Type);
  register("__able_mutex_lock", [i64Type], voidType);
  register("__able_mutex_unlock", [i64Type], voidType);
  register("__able_mutex_await_lock", [i64Type, anyType], anyType);

  register("__able_array_new", [], i64Type);
  register("__able_array_with_capacity", [i32Type], i64Type);
  register("__able_array_size", [i64Type], u64Type);
  register("__able_array_capacity", [i64Type], u64Type);
  register("__able_array_set_len", [i64Type, i32Type], voidType);
  register("__able_array_read", [i64Type, i32Type], unknown);
  register("__able_array_write", [i64Type, i32Type, unknown], voidType);
  register("__able_array_reserve", [i64Type, i32Type], u64Type);
  register("__able_array_clone", [i64Type], i64Type);

  register("__able_string_from_builtin", [stringType], arrayType(u8Type));
  register("__able_string_to_builtin", [arrayType(u8Type)], stringType);
  register("__able_char_from_codepoint", [i32Type], charType);

  const awaitableUnknown: TypeInfo = { kind: "interface", name: "Awaitable", typeArguments: [unknown] };
  register("__able_await_default", [{ kind: "function", parameters: [], returnType: unknown }], awaitableUnknown);
  register(
    "__able_await_sleep_ms",
    [primitiveType("i64"), { kind: "nullable", inner: { kind: "function", parameters: [], returnType: unknown } }],
    awaitableUnknown,
  );

  register("__able_hasher_create", [], i64Type);
  register("__able_hasher_write", [i64Type, stringType], voidType);
  register("__able_hasher_finish", [i64Type], i64Type);

  const divModStruct = structDefinition(
    "DivMod",
    [
      structFieldDefinition(simpleTypeExpression("T"), "quotient"),
      structFieldDefinition(simpleTypeExpression("T"), "remainder"),
    ],
    "named",
    [genericParameter("T")],
  );
  (divModStruct as any)._builtin = true;
  (divModStruct as any).origin = "<builtin>";
  context.registerStructDefinition(divModStruct as any);

  const errorInterface = interfaceDefinition(
    "Error",
    [
      functionSignature("message", [functionParameter("self", simpleTypeExpression("Self"))], stringType),
      functionSignature(
        "cause",
        [functionParameter("self", simpleTypeExpression("Self"))],
        nullableTypeExpression(simpleTypeExpression("Error")),
      ),
    ],
  );
  (errorInterface as any)._builtin = true;
  (errorInterface as any).origin = "<builtin>";
  context.registerInterfaceDefinition(errorInterface as any);
}

function registerBuiltinFunction(
  env: Environment,
  functionInfos: Map<string, FunctionInfo[]>,
  name: string,
  params: TypeInfo[],
  returnType: TypeInfo,
): void {
  const fnType: TypeInfo = {
    kind: "function",
    parameters: params,
    returnType,
  };
  env.define(name, fnType);
  const existing = functionInfos.get(name) ?? [];
  functionInfos.set(name, [
    ...existing,
    {
      name,
      fullName: name,
      parameters: params,
      genericConstraints: [],
      genericParamNames: [],
      whereClause: [],
      returnType,
    },
  ]);
}
