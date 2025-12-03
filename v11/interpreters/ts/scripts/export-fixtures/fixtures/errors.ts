import { AST } from "../../context";
import type { Fixture } from "../../types";

const errorsFixtures: Fixture[] = [
  {
      name: "errors/implicit_generic_where_ambiguity",
      module: AST.module([
        AST.interfaceDefinition(
          "Show",
          [
            AST.functionSignature(
              "render",
              [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
              AST.simpleTypeExpression("string"),
            ),
          ],
          undefined,
          AST.simpleTypeExpression("Self"),
        ),
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.fn(
          "format_point",
          [AST.param("value", AST.simpleTypeExpression("Point"))],
          [AST.identifier("value")],
          AST.simpleTypeExpression("Point"),
          undefined,
          [
            AST.whereClauseConstraint("Point", [
              AST.interfaceConstraint(AST.simpleTypeExpression("Show")),
            ]),
          ],
        ),
      ]),
      manifest: {
        description: "Where clause referencing known type names requires explicit generics",
        expect: {
          typecheckDiagnostics: [
            "typechecker: ../fixtures/ast/errors/implicit_generic_where_ambiguity/source.able:9:46 typechecker: cannot infer type parameter 'Point' because a type with the same name exists; declare it explicitly or qualify the type",
          ],
        },
      },
    },

  {
      name: "errors/implicit_generic_redeclaration",
      module: AST.module([
        AST.fn(
          "wrap",
          [AST.param("value", AST.simpleTypeExpression("T"))],
          [
            AST.structDefinition(
              "T",
              [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
              "named",
            ),
            AST.identifier("value"),
          ],
          AST.simpleTypeExpression("T"),
        ),
      ]),
      manifest: {
        description: "Redeclaring inferred generics inside a function body is rejected",
        skipTargets: ["go"],
        expect: {
          typecheckDiagnostics: [
            "typechecker: ../fixtures/ast/errors/implicit_generic_redeclaration/source.able typechecker: cannot redeclare inferred type parameter 'T' inside fn wrap (inferred at ../../fixtures/ast/errors/implicit_generic_redeclaration/source.able:0:0)",
          ],
        },
      },
    },

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

  {
      name: "errors/interface_self_pattern_mismatch",
      module: AST.module([
        AST.interfaceDefinition(
          "PointDisplay",
          [
            AST.functionSignature(
              "describe",
              [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
              AST.simpleTypeExpression("string"),
            ),
          ],
          undefined,
          AST.simpleTypeExpression("Point"),
        ),
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.structDefinition(
          "Line",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "start"),
            AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "end"),
          ],
          "named",
        ),
        AST.implementationDefinition(
          "PointDisplay",
          AST.simpleTypeExpression("Line"),
          [
            AST.functionDefinition(
              "describe",
              [AST.functionParameter("self", AST.simpleTypeExpression("Line"))],
              AST.blockExpression([AST.stringLiteral("line")]),
              AST.simpleTypeExpression("string"),
            ),
          ],
        ),
        AST.nil(),
      ]),
      manifest: {
        description: "Impl targeting mismatched concrete type violates explicit interface self pattern",
        expect: {
          result: { kind: "nil" },
          typecheckDiagnostics: [
            "typechecker: ../fixtures/ast/errors/interface_self_pattern_mismatch/source.able:13:1 typechecker: impl PointDisplay for Line must match interface self type 'Point'",
          ],
        },
      },
    },

  {
      name: "errors/interface_hkt_constructor_mismatch",
      module: AST.module([
        AST.interfaceDefinition(
          "Mapper",
          [
            AST.functionSignature(
              "wrap",
              [
                AST.functionParameter(
                  "self",
                  AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [
                    AST.simpleTypeExpression("T"),
                  ]),
                ),
                AST.functionParameter("value", AST.simpleTypeExpression("T")),
              ],
              AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [
                AST.simpleTypeExpression("T"),
              ]),
            ),
          ],
          [AST.genericParameter("T")],
          AST.genericTypeExpression(AST.simpleTypeExpression("F"), [AST.wildcardTypeExpression()]),
        ),
        AST.structDefinition("Array", [], "named", [AST.genericParameter("T")]),
        AST.implementationDefinition(
          "Mapper",
          AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [
            AST.simpleTypeExpression("i32"),
          ]),
          [
            AST.functionDefinition(
              "wrap",
              [
                AST.functionParameter(
                  "self",
                  AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [
                    AST.simpleTypeExpression("i32"),
                  ]),
                ),
                AST.functionParameter("value", AST.simpleTypeExpression("i32")),
              ],
              AST.blockExpression([AST.identifier("self")]),
              AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [
                AST.simpleTypeExpression("i32"),
              ]),
            ),
          ],
          undefined,
          undefined,
          [AST.simpleTypeExpression("i32")],
        ),
        AST.nil(),
      ]),
      manifest: {
        description: "Higher-kinded self pattern rejects implementations targeting concrete instantiations",
        expect: {
          result: { kind: "nil" },
          typecheckDiagnostics: [
            "typechecker: ../fixtures/ast/errors/interface_hkt_constructor_mismatch/source.able:8:1 typechecker: impl Mapper for Array i32 must match interface self type 'F _'",
          ],
        },
      },
    },

  {
      name: "errors/ufcs_static_method_not_found",
      module: AST.module([
        AST.structDefinition("Counter", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")], "named"),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Counter"),
          [
            AST.functionDefinition(
              "next",
              [AST.functionParameter("self", AST.simpleTypeExpression("Counter"))],
              AST.blockExpression([AST.bin("+", AST.member(AST.id("self"), "value"), AST.int(1))]),
              AST.simpleTypeExpression("i32"),
            ),
            AST.functionDefinition(
              "build",
              [AST.functionParameter("start", AST.simpleTypeExpression("i32"))],
              AST.blockExpression([
                AST.structLiteral([AST.structFieldInitializer(AST.id("start"), "value")], false, "Counter"),
              ]),
              AST.simpleTypeExpression("Counter"),
            ),
          ],
        ),
        AST.assign("c", AST.structLiteral([AST.structFieldInitializer(AST.int(1), "value")], false, "Counter")),
        AST.assign("bad", AST.functionCall(AST.id("build"), [AST.id("c")])),
        AST.id("bad"),
      ]),
      manifest: {
        description: "UFCS ignores static methods that lack a self parameter",
        expect: {
          errors: ["Undefined variable 'build'"],
          typecheckDiagnostics: [
            "typechecker: ../fixtures/ast/errors/ufcs_static_method_not_found/source.able:7:8 typechecker: undefined identifier 'build'",
          ],
        },
      },
    },
];

export default errorsFixtures;
