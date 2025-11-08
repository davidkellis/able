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

  {
      name: "errors/result_error_accessors",
      module: AST.module([
        AST.structDefinition("ChannelClosed", [], "named"),
        AST.structDefinition("ChannelNil", [], "named"),
        AST.structDefinition("ChannelSendOnClosed", [], "named"),
        AST.assignmentExpression(
          ":=",
          AST.identifier("failure_description"),
          AST.orElseExpression(
            AST.propagationExpression(
              AST.rescueExpression(
                AST.blockExpression([
                  AST.functionCall(AST.identifier("__able_channel_close"), [AST.integerLiteral(0)]),
                  AST.stringLiteral("ok"),
                ]),
                [
                  AST.matchClause(
                    AST.typedPattern(AST.identifier("err"), AST.simpleTypeExpression("Error")),
                    AST.identifier("err"),
                  ),
                ],
              ),
            ),
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("payload"),
                AST.memberAccessExpression(AST.identifier("err"), "value"),
              ),
              AST.assignmentExpression(
                ":=",
                AST.identifier("payload_tag"),
                AST.matchExpression(
                  AST.identifier("payload"),
                  [
                    AST.matchClause(
                      AST.structPattern([], false, "ChannelNil"),
                      AST.stringLiteral("ChannelNil"),
                    ),
                    AST.matchClause(
                      AST.structPattern([], false, "ChannelClosed"),
                      AST.stringLiteral("ChannelClosed"),
                    ),
                    AST.matchClause(
                      AST.structPattern([], false, "ChannelSendOnClosed"),
                      AST.stringLiteral("ChannelSendOnClosed"),
                    ),
                    AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Unknown")),
                  ],
                ),
              ),
              AST.assignmentExpression(
                ":=",
                AST.identifier("cause"),
                AST.functionCall(AST.memberAccessExpression(AST.identifier("err"), "cause"), []),
              ),
              AST.assignmentExpression(
                ":=",
                AST.identifier("cause_tag"),
                AST.matchExpression(
                  AST.identifier("cause"),
                  [
                    AST.matchClause(
                      AST.typedPattern(AST.identifier("inner"), AST.simpleTypeExpression("Error")),
                      AST.stringLiteral("cause"),
                    ),
                    AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("nil")),
                  ],
                ),
              ),
              AST.stringInterpolation([
                AST.functionCall(AST.memberAccessExpression(AST.identifier("err"), "message"), []),
                AST.stringLiteral("|"),
                AST.identifier("payload_tag"),
                AST.stringLiteral("|"),
                AST.identifier("cause_tag"),
              ]),
            ]),
            "err",
          ),
        ),
        AST.identifier("failure_description"),
      ]),
      manifest: {
        description: "Result handlers can call Error.message()/cause()/value",
        expect: {
          result: { kind: "string", value: "close of nil channel|ChannelNil|nil" },
        },
      },
    },
];

export default errorsFixtures;
