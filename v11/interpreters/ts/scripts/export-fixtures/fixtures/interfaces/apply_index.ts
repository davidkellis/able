import { AST } from "../../../context";
import type { Fixture } from "../../../types";

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

export const applyIndexDispatch: Fixture = {
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
      typecheckDiagnostics: [
        "typechecker: v11/fixtures/ast/interfaces/apply_index_dispatch/source.able:37:11 typechecker: '+' requires numeric operands (got Result Unknown and i32)",
        "typechecker: v11/fixtures/ast/interfaces/apply_index_dispatch/source.able:38:1 typechecker: '+' requires numeric operands (got Result Unknown and i32)",
      ],
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
        AST.gen(AST.ty("Index"), [AST.ty("i32"), AST.ty("i32")]),
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

export const applyIndexDiagnostics: Fixture = {
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
        "typechecker: v11/fixtures/ast/interfaces/apply_index_missing_impls/source.able:23:3 typechecker: cannot call non-callable value ReadOnlyPair (missing Apply implementation)",
        "typechecker: v11/fixtures/ast/interfaces/apply_index_missing_impls/source.able:27:3 typechecker: cannot assign via [] without IndexMut implementation on type Index i32 i32",
      ],
    },
  },
};
