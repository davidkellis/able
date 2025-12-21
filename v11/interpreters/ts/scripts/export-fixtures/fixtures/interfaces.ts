import { AST } from "../../context";
import type { Fixture } from "../../types";

const showInterface = AST.iface("Show", [
  AST.fnSig(
    "describe",
    [AST.param("self", AST.ty("Self"))],
    AST.ty("String"),
  ),
]);

const fancyStruct = AST.structDef(
  "Fancy",
  [AST.fieldDef(AST.ty("String"), "label")],
  "named",
);

const basicStruct = AST.structDef(
  "Basic",
  [AST.fieldDef(AST.ty("String"), "label")],
  "named",
);

const counterStruct = AST.structDef(
  "Counter",
  [AST.fieldDef(AST.ty("i32"), "value")],
  "named",
);

const entryStruct = AST.structDef(
  "Entry",
  [
    AST.fieldDef(AST.ty("String"), "key"),
    AST.fieldDef(AST.ty("Show"), "value"),
  ],
  "named",
);

const fancyDescribe = AST.fn(
  "describe",
  [AST.param("self", AST.ty("Fancy"))],
  [
    AST.ret(
      AST.interp([
        AST.str("fancy:"),
        AST.member(AST.id("self"), "label"),
      ]),
    ),
  ],
  AST.ty("String"),
);

const basicDescribe = AST.fn(
  "describe",
  [AST.param("self", AST.ty("Basic"))],
  [
    AST.ret(
      AST.interp([
        AST.str("basic:"),
        AST.member(AST.id("self"), "label"),
      ]),
    ),
  ],
  AST.ty("String"),
);

const counterDescribe = AST.fn(
  "describe",
  [AST.param("self", AST.ty("Counter"))],
  [
    AST.ret(
      AST.interp([
        AST.str("counter:"),
        AST.member(AST.id("self"), "value"),
      ]),
    ),
  ],
  AST.ty("String"),
);

const unionDescribe = AST.fn(
  "describe",
  [AST.param("self")],
  [AST.ret(AST.str("union"))],
  AST.ty("String"),
);

const fancyImpl = AST.impl("Show", AST.ty("Fancy"), [fancyDescribe]);
const basicImpl = AST.impl("Show", AST.ty("Basic"), [basicDescribe]);
const counterImpl = AST.impl("Show", AST.ty("Counter"), [counterDescribe]);

const unionImpl = AST.impl(
  "Show",
  AST.unionT([
    AST.ty("Fancy"),
    AST.ty("Basic"),
    AST.ty("Counter"),
  ]),
  [unionDescribe],
);

const bindFancy = AST.assign(
  AST.typedP(AST.id("fancy_value"), AST.ty("Show")),
  AST.structLiteral(
    [AST.fieldInit(AST.str("f"), "label")],
    false,
    "Fancy",
  ),
);

const bindBasic = AST.assign(
  AST.typedP(AST.id("basic_value"), AST.ty("Show")),
  AST.structLiteral(
    [AST.fieldInit(AST.str("b"), "label")],
    false,
    "Basic",
  ),
);

const entriesAssign = AST.assign(
  "entries",
  AST.arr(
    AST.structLiteral(
      [
        AST.fieldInit(AST.str("f"), "key"),
        AST.fieldInit(AST.id("fancy_value"), "value"),
      ],
      false,
      "Entry",
    ),
    AST.structLiteral(
      [
        AST.fieldInit(AST.str("b"), "key"),
        AST.fieldInit(AST.id("basic_value"), "value"),
      ],
      false,
      "Entry",
    ),
  ),
);

const initRangeResult = AST.assign("range_result", AST.str(""));
const initMapResult = AST.assign("map_result", AST.str(""));

const rangeLoop = AST.forIn(
  "range_idx",
  AST.range(AST.int(0), AST.int(1), true),
  AST.assign(
    "counter",
    AST.structLiteral(
      [AST.fieldInit(AST.id("range_idx"), "value")],
      false,
      "Counter",
    ),
  ),
  AST.assign(
    AST.typedP(AST.id("range_value"), AST.ty("Show")),
    AST.id("counter"),
  ),
  AST.assign(
    "range_result",
    AST.bin(
      "+",
      AST.id("range_result"),
      AST.call(AST.member(AST.id("range_value"), "describe")),
    ),
    "=",
  ),
);

const mapLoop = AST.forIn(
  "entry",
  AST.id("entries"),
  AST.assign(
    "map_result",
    AST.bin(
      "+",
      AST.id("map_result"),
      AST.call(
        AST.member(
          AST.member(AST.id("entry"), "value"),
          "describe",
        ),
      ),
    ),
    "=",
  ),
);

const dynamicInterfaceCollectionsModule = AST.mod([
  showInterface,
  fancyStruct,
  basicStruct,
  counterStruct,
  entryStruct,
  fancyImpl,
  basicImpl,
  counterImpl,
  unionImpl,
  bindFancy,
  bindBasic,
  entriesAssign,
  initRangeResult,
  initMapResult,
  rangeLoop,
  mapLoop,
  AST.arr(AST.id("range_result"), AST.id("map_result")),
]);

const applyIndexDispatchStruct = AST.structDef(
  "Pair",
  [
    AST.fieldDef(AST.gen(AST.ty("Array"), [AST.ty("i32")]), "values"),
  ],
  "named",
);

const applyInterface = AST.iface(
  "Apply",
  [
    AST.fnSig(
      "apply",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("idx", AST.ty("Args")),
      ],
      AST.ty("Result"),
    ),
  ],
  [AST.genericParameter("Args"), AST.genericParameter("Result")],
);

const indexInterface = AST.iface(
  "Index",
  [
    AST.fnSig(
      "get",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("idx", AST.ty("Idx")),
      ],
      AST.ty("Val"),
    ),
  ],
  [AST.genericParameter("Idx"), AST.genericParameter("Val")],
);

const indexMutInterface = AST.iface(
  "IndexMut",
  [
    AST.fnSig(
      "set",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("idx", AST.ty("Idx")),
        AST.param("value", AST.ty("Val")),
      ],
      AST.ty("void"),
    ),
  ],
  [AST.genericParameter("Idx"), AST.genericParameter("Val")],
);

const pairGet = AST.fn(
  "get",
  [
    AST.param("self", AST.ty("Self")),
    AST.param("idx", AST.ty("i32")),
  ],
  [
    AST.ret(AST.index(AST.member(AST.id("self"), "values"), AST.id("idx"))),
  ],
  AST.ty("i32"),
  undefined,
  undefined,
  false,
);

const pairSet = AST.fn(
  "set",
  [
    AST.param("self", AST.ty("Self")),
    AST.param("idx", AST.ty("i32")),
    AST.param("value", AST.ty("i32")),
  ],
  [
    AST.assignIndex(AST.member(AST.id("self"), "values"), AST.id("idx"), AST.id("value")),
    AST.ret(AST.nil()),
  ],
  AST.ty("void"),
  undefined,
  undefined,
  false,
);

const pairApply = AST.fn(
  "apply",
  [
    AST.param("self", AST.ty("Self")),
    AST.param("idx", AST.ty("i32")),
  ],
  [AST.ret(AST.call(AST.member(AST.id("self"), "get"), AST.id("idx")))],
  AST.ty("i32"),
  undefined,
  undefined,
  false,
);

const pairIndexImpl = AST.impl(
  "Index",
  AST.ty("Pair"),
  [pairGet],
  undefined,
  undefined,
  [AST.ty("i32"), AST.ty("i32")],
);

const pairIndexMutImpl = AST.impl(
  "IndexMut",
  AST.ty("Pair"),
  [pairSet],
  undefined,
  undefined,
  [AST.ty("i32"), AST.ty("i32")],
);

const pairApplyImpl = AST.impl(
  "Apply",
  AST.ty("Pair"),
  [pairApply],
  undefined,
  undefined,
  [AST.ty("i32"), AST.ty("i32")],
);

const bindPair = AST.assign(
  "pair",
  AST.structLiteral(
    [
      AST.fieldInit(AST.arr(AST.int(2), AST.int(5)), "values"),
    ],
    false,
    "Pair",
  ),
);

const writeLeft = AST.assignIndex(
  AST.id("pair"),
  AST.int(0),
  AST.bin("+", AST.call(AST.id("pair"), AST.int(1)), AST.int(3)),
);

const writeRight = AST.assignIndex(
  AST.id("pair"),
  AST.int(1),
  AST.bin("+", AST.index(AST.id("pair"), AST.int(0)), AST.call(AST.id("pair"), AST.int(1))),
);

const applyIndexDispatch: Fixture = {
  name: "interfaces/apply_index_dispatch",
  module: AST.mod([
    applyInterface,
    indexInterface,
    indexMutInterface,
    applyIndexDispatchStruct,
    pairIndexImpl,
    pairIndexMutImpl,
    pairApplyImpl,
    bindPair,
    writeLeft,
    writeRight,
    AST.bin("+", AST.index(AST.id("pair"), AST.int(0)), AST.call(AST.id("pair"), AST.int(1))),
  ]),
  manifest: {
    description: "Callable values dispatch to Apply.apply and []/[]= dispatch to Index/IndexMut implementations",
    expect: {
      result: {
        kind: "i32",
        value: "21",
      },
    },
  },
};

const applyIndexDiagnosticsStruct = AST.structDef(
  "ReadOnlyPair",
  [
    AST.fieldDef(
      AST.gen(AST.ty("Array"), [AST.ty("i32")]),
      "values",
    ),
  ],
  "named",
);

const applyIndexDiagnosticsIndexImpl = AST.impl(
  "Index",
  AST.ty("ReadOnlyPair"),
  [
    AST.fn(
      "get",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("idx", AST.ty("i32")),
      ],
      [
        AST.ret(
          AST.index(
            AST.member(AST.id("self"), "values"),
            AST.id("idx"),
          ),
        ),
      ],
      AST.ty("i32"),
      undefined,
      undefined,
      false,
    ),
  ],
  undefined,
  undefined,
  [AST.ty("i32"), AST.ty("i32")],
);

const callMissingApply = AST.fn(
  "call_missing_apply",
  [AST.param("pair", AST.ty("ReadOnlyPair"))],
  [AST.call(AST.id("pair"))],
  AST.ty("void"),
  undefined,
  undefined,
  false,
);

const assignMissingIndexMut = AST.fn(
  "assign_missing_index_mut",
  [AST.param("pair", AST.ty("ReadOnlyPair"))],
  [
    AST.assign(
      AST.typedP(
        AST.id("indexed"),
        AST.gen(
          AST.gen(AST.ty("Index"), [AST.ty("i32")]),
          [AST.ty("i32")],
        ),
      ),
      AST.id("pair"),
    ),
    AST.assignIndex(AST.id("indexed"), AST.int(0), AST.int(7)),
  ],
  AST.ty("i32"),
  undefined,
  undefined,
  false,
);

const applyIndexDiagnostics: Fixture = {
  name: "interfaces/apply_index_missing_impls",
  module: AST.mod([
    applyInterface,
    indexInterface,
    indexMutInterface,
    applyIndexDiagnosticsStruct,
    applyIndexDiagnosticsIndexImpl,
    callMissingApply,
    assignMissingIndexMut,
    AST.str("done"),
  ]),
  manifest: {
    description:
      "Calling non-Apply values and assigning through Index-only values surface Apply/IndexMut diagnostics",
    expect: {
      result: { kind: "String", value: "done" },
      typecheckDiagnostics: [
        "typechecker: ../../../fixtures/ast/interfaces/apply_index_missing_impls/source.able:23:3 typechecker: cannot call non-callable value ReadOnlyPair (missing Apply implementation)",
        "typechecker: ../../../fixtures/ast/interfaces/apply_index_missing_impls/source.able:27:10 typechecker: cannot assign via [] without IndexMut implementation on type Index i32 i32",
      ],
    },
  },
};

const interfacesFixtures: Fixture[] = [
  {
    name: "interfaces/dynamic_interface_collections",
    module: dynamicInterfaceCollectionsModule,
    manifest: {
      description:
        "Ranges and map-like collections of interface values dispatch the most specific impls even with union targets",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "String", value: "counter:0counter:1" },
            { kind: "String", value: "fancy:fbasic:b" },
          ],
        },
      },
    },
  },
  {
    name: "interfaces/imported_impl_dispatch",
    setupModules: {
      "package.json": AST.module(
        [
          AST.structDefinition(
            "Talker",
            [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "name")],
            "named",
          ),
          AST.interfaceDefinition(
            "Speaker",
            [
              AST.functionSignature(
                "speak",
                [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
                AST.simpleTypeExpression("String"),
              ),
            ],
          ),
          AST.implementationDefinition(
            "Speaker",
            AST.simpleTypeExpression("Talker"),
            [
              AST.functionDefinition(
                "speak",
                [AST.functionParameter("self", AST.simpleTypeExpression("Talker"))],
                AST.blockExpression([AST.memberAccessExpression(AST.id("self"), "name")]),
                AST.simpleTypeExpression("String"),
              ),
            ],
          ),
        ],
        [],
        AST.packageStatement(["speaklib"]),
      ),
    },
    module: AST.module(
      [
        AST.assign(
          "talker",
          AST.structLiteral(
            [AST.structFieldInitializer(AST.stringLiteral("hi"), "name")],
            false,
            "Talker",
          ),
        ),
        AST.functionCall(AST.memberAccessExpression(AST.id("talker"), "speak"), []),
      ],
      [AST.importStatement(["speaklib"], true)],
    ),
    manifest: {
      description: "Wildcard imports bring interface impls into scope for method dispatch",
      setup: ["package.json"],
      expect: {
        result: { kind: "String", value: "hi" },
      },
    },
  },
  applyIndexDispatch,
  applyIndexDiagnostics,
];

export default interfacesFixtures;
