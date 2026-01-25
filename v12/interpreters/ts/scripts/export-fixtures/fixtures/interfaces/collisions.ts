import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const collisionStruct = AST.structDef(
  "CollisionKey",
  [AST.fieldDef(AST.ty("i32"), "id")],
  "named",
);

const collisionHashImpl = AST.impl(
  "Hash",
  AST.ty("CollisionKey"),
  [
    AST.fn(
      "hash",
      [
        AST.param("self", AST.ty("Self")),
        AST.param("hasher", AST.ty("Hasher")),
      ],
      [
        AST.call(
          AST.member(AST.id("hasher"), "write_u64"),
          AST.int(1n, "u64"),
        ),
        AST.ret(AST.nil()),
      ],
      AST.ty("void"),
    ),
  ],
);

const collisionEqImpl = AST.impl(
  "Eq",
  AST.ty("CollisionKey"),
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

const collisionLiteral = (value: number) =>
  AST.structLiteral(
    [AST.fieldInit(AST.int(value), "id")],
    false,
    "CollisionKey",
  );

export const collisionFixture: Fixture = {
  name: "interfaces/hash_collision_handling",
  module: AST.mod(
    [
      collisionStruct,
      collisionHashImpl,
      collisionEqImpl,
      AST.assign(
        AST.typedP(
          AST.id("map"),
          AST.gen(AST.ty("HashMap"), [AST.ty("CollisionKey"), AST.ty("String")]),
        ),
        AST.call(AST.member(AST.id("HashMap"), "new")),
      ),
      AST.call(
        AST.member(AST.id("map"), "raw_set"),
        collisionLiteral(1),
        AST.str("first"),
      ),
      AST.call(
        AST.member(AST.id("map"), "raw_set"),
        collisionLiteral(2),
        AST.str("second"),
      ),
      AST.assign("score", AST.int(0)),
      AST.iff(
        AST.bin(
          "==",
          AST.call(AST.member(AST.id("map"), "raw_size")),
          AST.int(2),
        ),
        AST.assign(
          "score",
          AST.bin("+", AST.id("score"), AST.int(1)),
          "=",
        ),
      ),
      AST.match(
        AST.call(
          AST.member(AST.id("map"), "raw_get"),
          collisionLiteral(1),
        ),
        [
          AST.mc(
            AST.typedP(AST.id("value"), AST.ty("String")),
            AST.block(
              AST.iff(
                AST.bin("==", AST.id("value"), AST.str("first")),
                AST.assign(
                  "score",
                  AST.bin("+", AST.id("score"), AST.int(1)),
                  "=",
                ),
              ),
            ),
          ),
          AST.mc(AST.wc(), AST.block()),
        ],
      ),
      AST.match(
        AST.call(
          AST.member(AST.id("map"), "raw_get"),
          collisionLiteral(2),
        ),
        [
          AST.mc(
            AST.typedP(AST.id("value"), AST.ty("String")),
            AST.block(
              AST.iff(
                AST.bin("==", AST.id("value"), AST.str("second")),
                AST.assign(
                  "score",
                  AST.bin("+", AST.id("score"), AST.int(1)),
                  "=",
                ),
              ),
            ),
          ),
          AST.mc(AST.wc(), AST.block()),
        ],
      ),
      AST.bin("==", AST.id("score"), AST.int(3)),
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
    description: "Kernel HashMap resolves collisions using Eq",
    expect: {
      result: { kind: "bool", value: true },
    },
  },
};
