import { AST } from "../../context";
import type { Fixture } from "../../types";

const structsFixtures: Fixture[] = [
  {
      name: "structs/named_literal",
      module: AST.module([
        AST.structDefinition(
          "Point",
          [
            AST.fieldDef(AST.ty("i32"), "x"),
            AST.fieldDef(AST.ty("i32"), "y"),
          ],
          "named",
        ),
        AST.assign(
          "point",
          AST.structLiteral(
            [
              AST.fieldInit(AST.int(3), "x"),
              AST.fieldInit(AST.int(4), "y"),
            ],
            false,
            "Point",
          ),
        ),
        AST.member(AST.id("point"), "x"),
      ]),
      manifest: {
        description: "Named struct literal evaluates and exposes fields",
        expect: {
          result: { kind: "i32", value: 3 },
        },
      },
    },

  {
      name: "structs/positional_literal",
      module: AST.module([
        AST.structDefinition(
          "Pair",
          [
            AST.fieldDef(AST.ty("i32")),
            AST.fieldDef(AST.ty("i32")),
          ],
          "positional",
        ),
        AST.assign(
          "pair",
          AST.structLiteral(
            [
              AST.fieldInit(AST.int(7)),
              AST.fieldInit(AST.int(9)),
            ],
            true,
            "Pair",
          ),
        ),
        AST.member(AST.id("pair"), AST.int(1)),
      ]),
      manifest: {
        description: "Positional struct literal supports numeric member access",
        expect: {
          result: { kind: "i32", value: 9 },
        },
      },
    },
];

export default structsFixtures;
