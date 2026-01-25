import { AST } from "../../../context";
import type { Fixture } from "../../../types";

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

export const dynamicInterfaceCollectionsFixture: Fixture = {
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
};
