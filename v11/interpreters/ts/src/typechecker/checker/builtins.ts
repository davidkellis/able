import type * as AST from "../../ast";
import { buildStandardInterfaceBuiltins } from "../../builtins/interfaces";
import { arrayType, iteratorType, primitiveType, unknownType, type TypeInfo } from "../types";
import type { Environment } from "../environment";
import type { FunctionInfo } from "./types";

type BuiltinContext = {
  env: Environment;
  functionInfos: Map<string, FunctionInfo>;
  registerInterfaceDefinition(definition: AST.InterfaceDefinition): void;
  collectImplementationDefinition(definition: AST.ImplementationDefinition): void;
};

export function installBuiltins(context: BuiltinContext): void {
  const voidType = primitiveType("void");
  const boolType = primitiveType("bool");
  const i32Type = primitiveType("i32");
  const i64Type = primitiveType("i64");
  const u64Type = primitiveType("u64");
  const stringType = primitiveType("string");
  const charType = primitiveType("char");
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

  register("__able_array_new", [], i64Type);
  register("__able_array_with_capacity", [i32Type], i64Type);
  register("__able_array_size", [i64Type], u64Type);
  register("__able_array_capacity", [i64Type], u64Type);
  register("__able_array_set_len", [i64Type, i32Type], voidType);
  register("__able_array_read", [i64Type, i32Type], unknown);
  register("__able_array_write", [i64Type, i32Type, unknown], voidType);
  register("__able_array_reserve", [i64Type, i32Type], u64Type);
  register("__able_array_clone", [i64Type], i64Type);

  register("__able_string_from_builtin", [stringType], arrayType(i32Type));
  register("__able_string_to_builtin", [arrayType(i32Type)], stringType);
  register("__able_char_from_codepoint", [i32Type], charType);

  register("__able_hasher_create", [], i64Type);
  register("__able_hasher_write", [i64Type, stringType], voidType);
  register("__able_hasher_finish", [i64Type], i64Type);
  installBuiltinInterfaces(context);
}

function registerBuiltinFunction(
  env: Environment,
  functionInfos: Map<string, FunctionInfo>,
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
  functionInfos.set(name, {
    name,
    fullName: name,
    parameters: params,
    genericConstraints: [],
    genericParamNames: [],
    whereClause: [],
    returnType,
  });
}

function installBuiltinInterfaces(context: BuiltinContext): void {
  const { interfaces, implementations } = buildStandardInterfaceBuiltins();
  for (const iface of interfaces) {
    context.registerInterfaceDefinition(iface);
  }
  for (const impl of implementations) {
    // helper delegates to the shared implementation flow without duplicating logic here
    context.collectImplementationDefinition(impl);
  }
}
