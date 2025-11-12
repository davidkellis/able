import { AST } from "../../context";
import type { Fixture } from "../../types";

const basicsFixtures: Fixture[] = [
  {
      name: "basics/string_literal",
      module: AST.module([AST.str("hello")]),
      manifest: {
        description: "Evaluates a simple string literal module",
        expect: {
          result: { kind: "string", value: "hello" },
        },
      },
    },

  {
      name: "basics/bool_literal",
      module: AST.module([AST.bool(true)]),
      manifest: {
        description: "Evaluates a boolean literal",
        expect: {
          result: { kind: "bool", value: true },
        },
      },
    },

  {
      name: "basics/nil_literal",
      module: AST.module([AST.nil()]),
      manifest: {
        description: "Evaluates the nil literal",
        expect: {
          result: { kind: "nil" },
        },
      },
    },

  {
      name: "basics/char_literal",
      module: AST.module([AST.chr("a")]),
      manifest: {
        description: "Evaluates a character literal",
        expect: {
          result: { kind: "char", value: "a" },
        },
      },
    },
];

export default basicsFixtures;
