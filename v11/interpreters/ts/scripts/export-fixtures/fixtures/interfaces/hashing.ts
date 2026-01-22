import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const hashKernelImport = AST.importStatement(
  ["able", "kernel"],
  false,
  [
    AST.importSelector("KernelHasher"),
    AST.importSelector("Hash"),
    AST.importSelector("Hasher"),
  ],
);

const hashValueFn = AST.fn(
  "hash_value",
  [AST.param("value", AST.ty("T"))],
  [
    AST.assign(
      AST.typedP(AST.id("hasher"), AST.ty("Hasher")),
      AST.call(AST.member(AST.id("KernelHasher"), "new")),
    ),
    AST.call(AST.member(AST.id("value"), "hash"), AST.id("hasher")),
    AST.call(AST.member(AST.id("hasher"), "finish")),
  ],
  AST.ty("u64"),
  [AST.genericParameter("T")],
  [
    AST.whereClauseConstraint("T", [
      AST.interfaceConstraint(AST.ty("Hash")),
    ]),
  ],
);

const hashCheck = (value, expected) =>
  AST.iff(
    AST.bin(
      "!=",
      AST.call("hash_value", value),
      AST.int(expected, "u64"),
    ),
    AST.assign("ok", AST.bool(false), "="),
  );

export const primitiveHashingFixture: Fixture = {
  name: "interfaces/primitive_hashing",
  module: AST.mod(
    [
      hashValueFn,
      AST.assign("ok", AST.bool(true)),
      hashCheck(AST.bool(true), 12638152016183539244n),
      hashCheck(AST.int(-42, "i32"), 11047133508193088422n),
      hashCheck(AST.int(123n, "u64"), 12162020487158469588n),
      hashCheck(AST.chr("A"), 5559048874771775234n),
      hashCheck(AST.str("Able"), 9016399465655874229n),
      AST.id("ok"),
    ],
    [hashKernelImport],
  ),
  manifest: {
    description: "Hashes primitive values via KernelHasher and Hash impls",
    expect: {
      result: { kind: "bool", value: true },
    },
  },
};

export const kernelInterfaceFixture: Fixture = {
  name: "interfaces/kernel_interface_availability",
  module: AST.mod(
    [
      AST.fn(
        "probe",
        [AST.param("hasher", AST.ty("Hasher"))],
        [
          AST.call(
            AST.member(AST.id("hasher"), "write_u16"),
            AST.int(515, "u16"),
          ),
          AST.call(AST.member(AST.id("hasher"), "finish")),
        ],
        AST.ty("u64"),
      ),
      AST.call(
        "probe",
        AST.call(AST.member(AST.id("KernelHasher"), "new")),
      ),
    ],
    [
      AST.importStatement(
        ["able", "kernel"],
        false,
        [
          AST.importSelector("KernelHasher"),
          AST.importSelector("Hasher"),
        ],
      ),
    ],
  ),
  manifest: {
    description: "Kernel Hasher interface and KernelHasher are available for dispatch",
    expect: {
      result: { kind: "u64", value: "592598317564770290" },
    },
  },
};

const customHashEqStruct = AST.structDef(
  "Key",
  [AST.fieldDef(AST.ty("i32"), "id")],
  "named",
);

const customHashEqImpl = AST.impl(
  "Hash",
  AST.ty("Key"),
  [
    AST.fn(
      "hash",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("hasher", AST.ty("Hasher")),
      ],
      [
        AST.call(
          AST.member(AST.id("hasher"), "write_i32"),
          AST.member(AST.id("self"), "id"),
        ),
        AST.ret(AST.nil()),
      ],
      AST.ty("void"),
    ),
  ],
);

const customEqImpl = AST.impl(
  "Eq",
  AST.ty("Key"),
  [
    AST.fn(
      "eq",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("other", AST.ty("Self")),
      ],
      [
        AST.ret(
          AST.bin(
            "==",
            AST.member(AST.id("self"), "id"),
            AST.member(AST.id("other"), "id"),
          ),
        ),
      ],
      AST.ty("bool"),
    ),
  ],
);

const keyLiteral = (value: number) =>
  AST.structLiteral(
    [AST.fieldInit(AST.int(value), "id")],
    false,
    "Key",
  );

export const customHashEqFixture: Fixture = {
  name: "interfaces/custom_hash_eq",
  module: AST.mod(
    [
      customHashEqStruct,
      customHashEqImpl,
      customEqImpl,
      AST.assign(
        AST.typedP(
          AST.id("map"),
          AST.gen(AST.ty("HashMap"), [AST.ty("Key"), AST.ty("i32")]),
        ),
        AST.call(AST.member(AST.id("HashMap"), "new")),
      ),
      AST.call(
        AST.member(AST.id("map"), "raw_set"),
        keyLiteral(1),
        AST.int(10),
      ),
      AST.call(
        AST.member(AST.id("map"), "raw_set"),
        keyLiteral(2),
        AST.int(20),
      ),
      AST.assign("sum", AST.int(0)),
      AST.match(
        AST.call(
          AST.member(AST.id("map"), "raw_get"),
          keyLiteral(1),
        ),
        [
          AST.mc(
            AST.typedP(AST.id("value"), AST.ty("i32")),
            AST.block(
              AST.assign(
                "sum",
                AST.bin("+", AST.id("sum"), AST.id("value")),
                "=",
              ),
            ),
          ),
          AST.mc(AST.wc(), AST.block()),
        ],
      ),
      AST.match(
        AST.call(
          AST.member(AST.id("map"), "raw_get"),
          keyLiteral(2),
        ),
        [
          AST.mc(
            AST.typedP(AST.id("value"), AST.ty("i32")),
            AST.block(
              AST.assign(
                "sum",
                AST.bin("+", AST.id("sum"), AST.id("value")),
                "=",
              ),
            ),
          ),
          AST.mc(AST.wc(), AST.block()),
        ],
      ),
      AST.match(
        AST.call(
          AST.member(AST.id("map"), "raw_get"),
          keyLiteral(3),
        ),
        [
          AST.mc(
            AST.litP(AST.nil()),
            AST.block(
              AST.assign(
                "sum",
                AST.bin("+", AST.id("sum"), AST.int(1)),
                "=",
              ),
            ),
          ),
          AST.mc(AST.wc(), AST.block()),
        ],
      ),
      AST.bin("==", AST.id("sum"), AST.int(31)),
    ],
    [
      AST.importStatement(
        ["able", "kernel"],
        false,
        [
          AST.importSelector("HashMap"),
          AST.importSelector("Hash"),
          AST.importSelector("Eq"),
          AST.importSelector("Hasher"),
        ],
      ),
    ],
  ),
  manifest: {
    description: "Custom Hash/Eq impls drive kernel HashMap lookups",
    expect: {
      result: { kind: "bool", value: true },
    },
  },
};
