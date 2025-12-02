import * as AST from "../../ast";
import { buildStandardInterfaceBuiltins } from "../../builtins/interfaces";
import { arrayType, iteratorType, primitiveType, unknownType, type TypeInfo } from "../types";
import type { Environment } from "../environment";
import type { FunctionInfo } from "./types";

type BuiltinContext = {
  env: Environment;
  functionInfos: Map<string, FunctionInfo>;
  implementationContext?: unknown;
  registerStructDefinition(definition: AST.StructDefinition): void;
  registerTypeAlias(definition: AST.TypeAliasDefinition): void;
  registerInterfaceDefinition(definition: AST.InterfaceDefinition): void;
  collectImplementationDefinition(definition: AST.ImplementationDefinition): void;
  collectMethodsDefinition(definition: AST.MethodsDefinition): void;
};

export function installBuiltins(context: BuiltinContext): void {
  const voidType = primitiveType("void");
  const boolType = primitiveType("bool");
  const i32Type = primitiveType("i32");
  const i64Type = primitiveType("i64");
  const u64Type = primitiveType("u64");
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

  register("__able_string_from_builtin", [stringType], arrayType(i32Type));
  register("__able_string_to_builtin", [arrayType(i32Type)], stringType);
  register("__able_char_from_codepoint", [i32Type], charType);
  installAwaitBuiltins(context);

  register("__able_hasher_create", [], i64Type);
  register("__able_hasher_write", [i64Type, stringType], voidType);
  register("__able_hasher_finish", [i64Type], i64Type);
  installOrderingBuiltins(context);
  installIterationBuiltins(context);
  installBuiltinInterfaces(context);
  installStdlibStubs(context);
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

function installStdlibStubs(context: BuiltinContext): void {
  buildArrayMethodStubs().forEach((methods) => context.collectMethodsDefinition(methods));
  context.collectMethodsDefinition(buildListMethods());
  context.collectMethodsDefinition(buildLinkedListMethods());
  context.collectMethodsDefinition(buildLazySeqMethods());
  context.collectMethodsDefinition(buildVectorMethods());
  context.collectMethodsDefinition(buildHashMapMethods());
  context.collectMethodsDefinition(buildHashSetMethods());
  context.collectMethodsDefinition(buildBitSetMethods());
  context.collectMethodsDefinition(buildTreeMapMethods());
  context.collectMethodsDefinition(buildTreeSetMethods());
  context.collectImplementationDefinition(buildChannelIterableImpl());
}

function installAwaitBuiltins(context: BuiltinContext): void {
  const awaitWakerStruct = AST.structDefinition("AwaitWaker", [], "named");
  const awaitRegistrationStruct = AST.structDefinition("AwaitRegistration", [], "named");
  context.registerStructDefinition(awaitWakerStruct);
  context.registerStructDefinition(awaitRegistrationStruct);

  const awaitableInterface = AST.interfaceDefinition(
    "Awaitable",
    [
      AST.functionSignature(
        "is_ready",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("bool"),
      ),
      AST.functionSignature(
        "register",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("waker", AST.simpleTypeExpression("AwaitWaker")),
        ],
        AST.simpleTypeExpression("AwaitRegistration"),
      ),
      AST.functionSignature(
        "commit",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("Output"),
      ),
      AST.functionSignature(
        "is_default",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("bool"),
      ),
    ],
    [AST.genericParameter("Output")],
  );
  context.registerInterfaceDefinition(awaitableInterface);

  const wakerMethods = AST.methodsDefinition(AST.simpleTypeExpression("AwaitWaker"), [
    AST.functionDefinition(
      "wake",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    ),
  ]);
  const registrationMethods = AST.methodsDefinition(AST.simpleTypeExpression("AwaitRegistration"), [
    AST.functionDefinition(
      "cancel",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    ),
  ]);
  context.collectMethodsDefinition(wakerMethods);
  context.collectMethodsDefinition(registrationMethods);

  const awaitableUnknown: TypeInfo = { kind: "interface", name: "Awaitable", typeArguments: [unknownType] };
  const callbackType: TypeInfo = { kind: "function", parameters: [], returnType: unknownType };
  registerBuiltinFunction(context.env, context.functionInfos, "__able_await_default", [callbackType], awaitableUnknown);
  registerBuiltinFunction(
    context.env,
    context.functionInfos,
    "__able_await_sleep_ms",
    [primitiveType("i64"), { kind: "nullable", inner: callbackType }],
    awaitableUnknown,
  );
}

function installIterationBuiltins(context: BuiltinContext): void {
  context.registerStructDefinition(AST.structDefinition("IteratorEnd", [], "singleton"));

  const iterParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const nextSignature = AST.functionSignature(
    "next",
    [selfParam],
    AST.unionTypeExpression([AST.simpleTypeExpression("T"), AST.simpleTypeExpression("IteratorEnd")]),
  );
  const iteratorInterface = AST.interfaceDefinition("Iterator", [nextSignature], [iterParam]);
  context.registerInterfaceDefinition(iteratorInterface);

  const iterableParam = AST.genericParameter("T");
  const iterableSelf = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const eachSignature = AST.functionSignature(
    "each",
    [
      iterableSelf,
      AST.functionParameter(
        "visit",
        AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("void")),
      ),
    ],
    AST.simpleTypeExpression("void"),
  );
  const iteratorSignature = AST.functionSignature(
    "iterator",
    [iterableSelf],
    AST.genericTypeExpression(AST.simpleTypeExpression("Iterator"), [AST.simpleTypeExpression("T")]),
  );
  const iterableInterface = AST.interfaceDefinition("Iterable", [eachSignature, iteratorSignature], [iterableParam]);
  context.registerInterfaceDefinition(iterableInterface);
}

function installOrderingBuiltins(context: BuiltinContext): void {
  const voidBlock = AST.blockExpression([]);
  const orderingUnion = AST.unionTypeExpression([
    AST.simpleTypeExpression("Less"),
    AST.simpleTypeExpression("Equal"),
    AST.simpleTypeExpression("Greater"),
  ]);

  context.registerStructDefinition(AST.structDefinition("Less", [], "singleton"));
  context.registerStructDefinition(AST.structDefinition("Equal", [], "singleton"));
  context.registerStructDefinition(AST.structDefinition("Greater", [], "singleton"));
  context.registerTypeAlias(AST.typeAliasDefinition("Ordering", orderingUnion));

  const rhsParam = AST.genericParameter("Rhs");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const otherParam = AST.functionParameter("other", AST.simpleTypeExpression("Rhs"));
  const cmpSignature = AST.functionSignature("cmp", [selfParam, otherParam], AST.simpleTypeExpression("Ordering"));
  const ordInterface = AST.interfaceDefinition("Ord", [cmpSignature], [rhsParam]);
  context.registerInterfaceDefinition(ordInterface);

  const orderingMethodsBlock = (typeName: string): AST.MethodsDefinition => {
    const target = AST.simpleTypeExpression(typeName);
    const method = AST.functionDefinition(
      "cmp",
      [selfParam, AST.functionParameter("other", AST.simpleTypeExpression(typeName))],
      voidBlock,
      AST.simpleTypeExpression("Ordering"),
    );
    return AST.methodsDefinition(target, [method]);
  };

  context.collectMethodsDefinition(orderingMethodsBlock("i32"));
  context.collectMethodsDefinition(orderingMethodsBlock("string"));
}

function buildArrayMethodStubs(): AST.MethodsDefinition[] {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const elementParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const indexParam = AST.functionParameter("idx", AST.simpleTypeExpression("i32"));
  const arrayType = AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], block, arrayType),
    AST.functionDefinition(
      "with_capacity",
      [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
      block,
      arrayType,
    ),
    AST.functionDefinition("size", [selfParam], block, AST.simpleTypeExpression("u64")),
    AST.functionDefinition("push", [selfParam, elementParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition("pop", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("clear", [selfParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition("read_slot", [selfParam, indexParam], block, AST.simpleTypeExpression("T")),
    AST.functionDefinition("write_slot", [selfParam, indexParam, elementParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition("len", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("capacity", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("get", [selfParam, indexParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition(
      "set",
      [selfParam, indexParam, elementParam],
      block,
      AST.resultTypeExpression(AST.simpleTypeExpression("nil")),
    ),
    AST.functionDefinition("first", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("last", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("push_all", [selfParam, AST.functionParameter("values", arrayType)], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition(
      "map",
      [
        selfParam,
        AST.functionParameter(
          "f",
          AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("U")),
        ),
      ],
      block,
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("U")]),
      [AST.genericParameter("U")],
    ),
  ];

  const baseMethods = AST.methodsDefinition(arrayType, methods, [typeParam]);

  const filterMethods = AST.methodsDefinition(
    arrayType,
    [
      AST.functionDefinition(
        "filter",
        [
          selfParam,
          AST.functionParameter(
            "predicate",
            AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("bool")),
          ),
        ],
        block,
        arrayType,
      ),
    ],
    [typeParam],
    [AST.whereClauseConstraint("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Clone"))])],
  );

  return [baseMethods, filterMethods];
}

function buildListMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const valueParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const indexParam = AST.functionParameter("index", AST.simpleTypeExpression("i32"));
  const listType = AST.genericTypeExpression(AST.simpleTypeExpression("List"), [AST.simpleTypeExpression("T")]);
  const arrayOfT = AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("empty", [], block, listType),
    AST.functionDefinition("len", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("prepend", [selfParam, valueParam], block, listType),
    AST.functionDefinition("head", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("tail", [selfParam], block, listType),
    AST.functionDefinition("first", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("last", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("nth", [selfParam, indexParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("concat", [selfParam, AST.functionParameter("other", listType)], block, listType),
    AST.functionDefinition("append", [selfParam, valueParam], block, listType),
    AST.functionDefinition("reverse", [selfParam], block, listType),
    AST.functionDefinition("to_array", [selfParam], block, arrayOfT),
  ];

  return AST.methodsDefinition(listType, methods, [typeParam]);
}

function buildLinkedListMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const valueParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const nodeType = AST.genericTypeExpression(AST.simpleTypeExpression("ListNode"), [AST.simpleTypeExpression("T")]);
  const listType = AST.genericTypeExpression(AST.simpleTypeExpression("LinkedList"), [AST.simpleTypeExpression("T")]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], block, listType),
    AST.functionDefinition("len", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("head_node", [selfParam], block, AST.nullableTypeExpression(nodeType)),
    AST.functionDefinition("tail_node", [selfParam], block, AST.nullableTypeExpression(nodeType)),
    AST.functionDefinition("push_front", [selfParam, valueParam], block, nodeType),
    AST.functionDefinition("push_back", [selfParam, valueParam], block, nodeType),
    AST.functionDefinition("pop_front", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("pop_back", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("insert_after", [selfParam, AST.functionParameter("node", nodeType), valueParam], block, nodeType),
    AST.functionDefinition("remove_node", [selfParam, AST.functionParameter("node", nodeType)], block, AST.simpleTypeExpression("T")),
    AST.functionDefinition("clear", [selfParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition(
      "for_each",
      [
        selfParam,
        AST.functionParameter(
          "visit",
          AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("void")),
        ),
      ],
      block,
      AST.simpleTypeExpression("void"),
    ),
  ];

  return AST.methodsDefinition(listType, methods, [typeParam]);
}

function buildLazySeqMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const indexParam = AST.functionParameter("index", AST.simpleTypeExpression("i32"));
  const countParam = AST.functionParameter("count", AST.simpleTypeExpression("i32"));
  const iteratorOfT = AST.genericTypeExpression(AST.simpleTypeExpression("Iterator"), [AST.simpleTypeExpression("T")]);
  const iterableOfT = AST.genericTypeExpression(AST.simpleTypeExpression("Iterable"), [AST.simpleTypeExpression("T")]);
  const arrayOfT = AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]);
  const lazySeqType = AST.genericTypeExpression(AST.simpleTypeExpression("LazySeq"), [AST.simpleTypeExpression("T")]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("from_iterator", [AST.functionParameter("iterator", iteratorOfT)], block, lazySeqType),
    AST.functionDefinition("from_iterable", [AST.functionParameter("iterable", iterableOfT)], block, lazySeqType),
    AST.functionDefinition("len", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("get", [selfParam, indexParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("head", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("tail", [selfParam], block, lazySeqType),
    AST.functionDefinition("to_array", [selfParam], block, arrayOfT),
    AST.functionDefinition("take", [selfParam, countParam], block, arrayOfT),
    AST.functionDefinition(
      "for_each",
      [
        selfParam,
        AST.functionParameter(
          "visit",
          AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("void")),
        ),
      ],
      block,
      AST.simpleTypeExpression("void"),
    ),
  ];

  return AST.methodsDefinition(lazySeqType, methods, [typeParam]);
}

function buildVectorMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const valueParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const indexParam = AST.functionParameter("index", AST.simpleTypeExpression("i32"));
  const vectorType = AST.genericTypeExpression(AST.simpleTypeExpression("Vector"), [AST.simpleTypeExpression("T")]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], block, vectorType),
    AST.functionDefinition("len", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("tail_offset", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("get", [selfParam, indexParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("first", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("last", [selfParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("set", [selfParam, indexParam, valueParam], block, vectorType),
    AST.functionDefinition("push", [selfParam, valueParam], block, vectorType),
    AST.functionDefinition("pop", [selfParam], block, vectorType),
  ];

  return AST.methodsDefinition(vectorType, methods, [typeParam]);
}

function buildHashMapMethods(): AST.MethodsDefinition {
  const keyParam = AST.functionParameter("key", AST.simpleTypeExpression("K"));
  const valueParam = AST.functionParameter("value", AST.simpleTypeExpression("V"));
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const hashMapType = AST.genericTypeExpression(AST.simpleTypeExpression("HashMap"), [
    AST.simpleTypeExpression("K"),
    AST.simpleTypeExpression("V"),
  ]);
  const block = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], block, hashMapType),
    AST.functionDefinition(
      "with_capacity",
      [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
      block,
      hashMapType,
    ),
    AST.functionDefinition("clear", [selfParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition("capacity", [selfParam], block, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("get", [selfParam, keyParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("V"))),
    AST.functionDefinition("set", [selfParam, keyParam, valueParam], block, AST.simpleTypeExpression("void")),
    AST.functionDefinition("remove", [selfParam, keyParam], block, AST.nullableTypeExpression(AST.simpleTypeExpression("V"))),
    AST.functionDefinition("contains", [selfParam, keyParam], block, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("size", [selfParam], block, AST.simpleTypeExpression("i32")),
  ];

  return AST.methodsDefinition(hashMapType, methods, [
    AST.genericParameter("K"),
    AST.genericParameter("V"),
  ]);
}

function buildHashSetMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const elementParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const hashSetType = AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [
    AST.simpleTypeExpression("T"),
  ]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], AST.blockExpression([]), hashSetType),
    AST.functionDefinition(
      "with_capacity",
      [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
      AST.blockExpression([]),
      hashSetType,
    ),
    AST.functionDefinition("add", [selfParam, elementParam], AST.blockExpression([]), AST.simpleTypeExpression("bool")),
    AST.functionDefinition(
      "remove",
      [selfParam, elementParam],
      AST.blockExpression([]),
      AST.simpleTypeExpression("bool"),
    ),
    AST.functionDefinition(
      "contains",
      [selfParam, elementParam],
      AST.blockExpression([]),
      AST.simpleTypeExpression("bool"),
    ),
    AST.functionDefinition("size", [selfParam], AST.blockExpression([]), AST.simpleTypeExpression("i32")),
    AST.functionDefinition("clear", [selfParam], AST.blockExpression([]), AST.simpleTypeExpression("void")),
    AST.functionDefinition(
      "is_empty",
      [selfParam],
      AST.blockExpression([]),
      AST.simpleTypeExpression("bool"),
    ),
  ];

  return AST.methodsDefinition(hashSetType, methods, [typeParam]);
}

function buildBitSetMethods(): AST.MethodsDefinition {
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const bitParam = AST.functionParameter("bit", AST.simpleTypeExpression("i32"));
  const visitorParam = AST.functionParameter(
    "visit",
    AST.functionTypeExpression([AST.simpleTypeExpression("i32")], AST.simpleTypeExpression("void")),
  );
  const bitSetType = AST.simpleTypeExpression("BitSet");
  const iteratorOfI32 = AST.genericTypeExpression(AST.simpleTypeExpression("Iterator"), [AST.simpleTypeExpression("i32")]);
  const voidBlock = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], voidBlock, bitSetType),
    AST.functionDefinition("set", [selfParam, bitParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("reset", [selfParam, bitParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("flip", [selfParam, bitParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("contains", [selfParam, bitParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("clear", [selfParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("each", [selfParam, visitorParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("iterator", [selfParam], voidBlock, iteratorOfI32),
  ];

  return AST.methodsDefinition(bitSetType, methods);
}

function buildTreeMapMethods(): AST.MethodsDefinition {
  const keyParam = AST.genericParameter("K");
  const valueParam = AST.genericParameter("V");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const keyParamDecl = AST.functionParameter("key", AST.simpleTypeExpression("K"));
  const valueParamDecl = AST.functionParameter("value", AST.simpleTypeExpression("V"));
  const treeMapType = AST.genericTypeExpression(AST.simpleTypeExpression("TreeMap"), [
    AST.simpleTypeExpression("K"),
    AST.simpleTypeExpression("V"),
  ]);
  const treeEntryType = AST.genericTypeExpression(AST.simpleTypeExpression("TreeEntry"), [
    AST.simpleTypeExpression("K"),
    AST.simpleTypeExpression("V"),
  ]);
  const treeEntryArray = AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [treeEntryType]);
  const voidBlock = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], voidBlock, treeMapType),
    AST.functionDefinition("len", [selfParam], voidBlock, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("set", [selfParam, keyParamDecl, valueParamDecl], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("get", [selfParam, keyParamDecl], voidBlock, AST.nullableTypeExpression(AST.simpleTypeExpression("V"))),
    AST.functionDefinition("contains", [selfParam, keyParamDecl], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("remove", [selfParam, keyParamDecl], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("first", [selfParam], voidBlock, AST.nullableTypeExpression(treeEntryType)),
    AST.functionDefinition("last", [selfParam], voidBlock, AST.nullableTypeExpression(treeEntryType)),
    AST.functionDefinition("to_array", [selfParam], voidBlock, treeEntryArray),
    AST.functionDefinition(
      "keys",
      [selfParam],
      voidBlock,
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("K")]),
    ),
    AST.functionDefinition(
      "values",
      [selfParam],
      voidBlock,
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("V")]),
    ),
    AST.functionDefinition(
      "each",
      [
        selfParam,
        AST.functionParameter(
          "visit",
          AST.functionTypeExpression([treeEntryType], AST.simpleTypeExpression("void")),
        ),
      ],
      voidBlock,
      AST.simpleTypeExpression("void"),
    ),
    AST.functionDefinition(
      "each_key",
      [
        selfParam,
        AST.functionParameter(
          "visit",
          AST.functionTypeExpression([AST.simpleTypeExpression("K")], AST.simpleTypeExpression("void")),
        ),
      ],
      voidBlock,
      AST.simpleTypeExpression("void"),
    ),
  ];

  const keyConstraints = [
    AST.interfaceConstraint(
      AST.genericTypeExpression(AST.simpleTypeExpression("Ord"), [AST.simpleTypeExpression("K")]),
    ),
    AST.interfaceConstraint(AST.simpleTypeExpression("Clone")),
  ];
  const valueConstraints = [AST.interfaceConstraint(AST.simpleTypeExpression("Clone"))];
  const whereClause = [
    AST.whereClauseConstraint("K", keyConstraints),
    AST.whereClauseConstraint("V", valueConstraints),
  ];

  return AST.methodsDefinition(treeMapType, methods, [keyParam, valueParam], whereClause);
}

function buildTreeSetMethods(): AST.MethodsDefinition {
  const typeParam = AST.genericParameter("T");
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const valueParam = AST.functionParameter("value", AST.simpleTypeExpression("T"));
  const treeSetType = AST.genericTypeExpression(AST.simpleTypeExpression("TreeSet"), [AST.simpleTypeExpression("T")]);
  const arrayOfT = AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]);
  const voidBlock = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("new", [], voidBlock, treeSetType),
    AST.functionDefinition("len", [selfParam], voidBlock, AST.simpleTypeExpression("i32")),
    AST.functionDefinition("is_empty", [selfParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("insert", [selfParam, valueParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("remove", [selfParam, valueParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("contains", [selfParam, valueParam], voidBlock, AST.simpleTypeExpression("bool")),
    AST.functionDefinition("first", [selfParam], voidBlock, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("last", [selfParam], voidBlock, AST.nullableTypeExpression(AST.simpleTypeExpression("T"))),
    AST.functionDefinition("to_array", [selfParam], voidBlock, arrayOfT),
    AST.functionDefinition("clear", [selfParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition(
      "each",
      [
        selfParam,
        AST.functionParameter(
          "visit",
          AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("void")),
        ),
      ],
      voidBlock,
      AST.simpleTypeExpression("void"),
    ),
  ];

  const ordConstraint = AST.interfaceConstraint(
    AST.genericTypeExpression(AST.simpleTypeExpression("Ord"), [AST.simpleTypeExpression("T")]),
  );
  const cloneConstraint = AST.interfaceConstraint(AST.simpleTypeExpression("Clone"));
  const whereClause = [AST.whereClauseConstraint("T", [ordConstraint, cloneConstraint])];

  return AST.methodsDefinition(treeSetType, methods, [typeParam], whereClause);
}

function buildChannelIterableImpl(): AST.ImplementationDefinition {
  const typeParam = AST.genericParameter("T");
  const channelType = AST.genericTypeExpression(AST.simpleTypeExpression("Channel"), [AST.simpleTypeExpression("T")]);
  const selfParam = AST.functionParameter("self", AST.simpleTypeExpression("Self"));
  const visitParam = AST.functionParameter(
    "visit",
    AST.functionTypeExpression([AST.simpleTypeExpression("T")], AST.simpleTypeExpression("void")),
  );
  const iteratorReturn = AST.genericTypeExpression(AST.simpleTypeExpression("Iterator"), [AST.simpleTypeExpression("T")]);
  const voidBlock = AST.blockExpression([]);

  const methods: AST.FunctionDefinition[] = [
    AST.functionDefinition("each", [selfParam, visitParam], voidBlock, AST.simpleTypeExpression("void")),
    AST.functionDefinition("iterator", [selfParam], voidBlock, iteratorReturn),
  ];

  return AST.implementationDefinition(
    "Iterable",
    channelType,
    methods,
    undefined,
    [typeParam],
    [AST.simpleTypeExpression("T")],
  );
}
