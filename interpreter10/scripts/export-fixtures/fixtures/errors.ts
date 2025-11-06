import { AST } from "../../context";
import type { Fixture } from "../../types";

const errorsFixtures: Fixture[] = [
  {
      name: "errors/rescue_guard",
      module: AST.module([
        AST.rescue(
          AST.block(AST.raise(AST.str("boom"))),
          [
            AST.mc(AST.litP(AST.str("ignore")), AST.str("ignored")),
            AST.mc(
              AST.id("msg"),
              AST.block(
                AST.ifExpression(
                  AST.bin("==", AST.id("msg"), AST.str("boom")),
                  AST.block(AST.str("handled")),
                ),
              ),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Rescue guard selects matching clause",
        expect: {
          result: { kind: "nil" },
        },
      },
    },

  {
      name: "errors/raise_manifest",
      module: AST.module([
        AST.raise(AST.str("boom")),
      ]),
      manifest: {
        description: "Fixture raises error",
        expect: {
          errors: ["boom"],
        },
      },
    },

  {
      name: "errors/rescue_catch",
      module: AST.module([
        AST.rescue(
          AST.block(AST.raise(AST.str("boom"))),
          [
            AST.mc(
              AST.id("err"),
              AST.block(AST.call("print", AST.id("err")), AST.str("handled")),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Rescue expression catches raise",
        expect: {
          stdout: ["[error]"],
          result: { kind: "string", value: "handled" },
        },
      },
    },

  {
      name: "errors/rescue_typed_pattern",
      module: AST.module([
        AST.rescue(
          AST.block(AST.raise(AST.str("boom"))),
          [
            AST.matchClause(
              AST.typedPattern(AST.identifier("err"), AST.simpleTypeExpression("Error")),
              AST.stringLiteral("caught"),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Typed pattern catches raised error",
        expect: {
          result: { kind: "string", value: "caught" },
        },
      },
    },

  {
      name: "errors/or_else_handler",
      module: AST.module([
        AST.orElseExpression(
          AST.propagationExpression(
            AST.blockExpression([AST.raiseStatement(AST.stringLiteral("boom"))]),
          ),
          AST.blockExpression([AST.stringLiteral("handled")]),
          "err",
        ),
      ]),
      manifest: {
        description: "Or else handler runs when propagation raises",
        expect: {
          result: { kind: "string", value: "handled" },
        },
      },
    },

  {
      name: "errors/ensure_runs",
      module: AST.module([
        AST.ensureExpression(
          AST.rescueExpression(
            AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
            [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("rescued"))],
          ),
          AST.blockExpression([AST.call("print", AST.stringLiteral("ensure"))]),
        ),
      ]),
      manifest: {
        description: "Ensure block executes regardless of rescue",
        expect: {
          stdout: ["ensure"],
          result: { kind: "string", value: "rescued" },
        },
      },
    },
];

export default errorsFixtures;
