import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - impl resolution semantics", () => {
  test("nested generic constraints pick most specific impl", () => {
    const I = new InterpreterV10();

    const readable = AST.interfaceDefinition("Readable", [
      AST.functionSignature(
        "read",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(readable);

    const comparable = AST.interfaceDefinition("Comparable", [
      AST.functionSignature(
        "cmp",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self")), AST.functionParameter("other", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("i32")
      ),
    ]);
    I.evaluate(comparable);

    const container = AST.structDefinition(
      "Container",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(container);

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "show",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const fancyNumDef = AST.structDefinition(
      "FancyNum",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named"
    );
    I.evaluate(fancyNumDef);

    const fancyReadable = AST.implementationDefinition(
      "Readable",
      AST.simpleTypeExpression("FancyNum"),
      [
        AST.functionDefinition(
          "read",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancyNum"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.stringInterpolation([
                AST.stringLiteral("#"),
                AST.memberAccessExpression(AST.identifier("self"), "value"),
              ])
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyReadable);

    const fancyComparable = AST.implementationDefinition(
      "Comparable",
      AST.simpleTypeExpression("FancyNum"),
      [
        AST.functionDefinition(
          "cmp",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("FancyNum")),
            AST.functionParameter("other", AST.simpleTypeExpression("FancyNum")),
          ],
          AST.blockExpression([
            AST.returnStatement(AST.integerLiteral(0)),
          ]),
          AST.simpleTypeExpression("i32")
        ),
      ]
    );
    I.evaluate(fancyComparable);

    const baseShow = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Container"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "show",
          [AST.functionParameter("self", AST.simpleTypeExpression("Container"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "value"),
                  "read"
                ),
                []
              )
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T")],
      undefined,
      [
        AST.whereClauseConstraint("T", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Readable")),
        ]),
      ]
    );
    I.evaluate(baseShow);

    const specificShow = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Container"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "show",
          [AST.functionParameter("self", AST.simpleTypeExpression("Container"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("comparable")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T")],
      undefined,
      [
        AST.whereClauseConstraint("T", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Readable")),
          AST.interfaceConstraint(AST.simpleTypeExpression("Comparable")),
        ]),
      ]
    );
    I.evaluate(specificShow);

    const value = AST.structLiteral([
      AST.structFieldInitializer(
        AST.structLiteral([
          AST.structFieldInitializer(AST.integerLiteral(42), "value"),
        ], false, "FancyNum"),
        "value"
      ),
    ], false, "Container", undefined, [AST.simpleTypeExpression("FancyNum")]);

    const call = AST.functionCall(
      AST.memberAccessExpression(value, "show"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "comparable" });
  });

  test("incomparable impls trigger ambiguity error", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const copyIface = AST.interfaceDefinition("Copyable", [
      AST.functionSignature(
        "duplicate",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("Self")
      ),
    ]);
    I.evaluate(copyIface);

    const pointDef = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    I.evaluate(pointDef);

    const pointMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("Point")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(pointMethods);

    const showPoint = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("Point")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(showPoint);

    const copyPoint = AST.implementationDefinition(
      "Copyable",
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "duplicate",
          [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
          AST.blockExpression([
            AST.returnStatement(AST.identifier("self")),
          ]),
          AST.simpleTypeExpression("Point")
        ),
      ]
    );
    I.evaluate(copyPoint);

    const wrapperDef = AST.structDefinition(
      "Wrapper",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const showImpl = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("show-constrained")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
    );
    I.evaluate(showImpl);

    const copyImpl = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("copy-constrained")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Copyable"))])]
    );
    I.evaluate(copyImpl);

    const pointLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(3), "x"),
      AST.structFieldInitializer(AST.integerLiteral(4), "y"),
    ], false, "Point");

    const wrapperPointLiteral = AST.structLiteral(
      [AST.structFieldInitializer(pointLiteral, "value")],
      false,
      "Wrapper",
      undefined,
      [AST.simpleTypeExpression("Point")]
    );

    const call = AST.functionCall(
      AST.memberAccessExpression(wrapperPointLiteral, "to_string"),
      []
    );
    expect(() => I.evaluate(call)).toThrow(/ambiguous implementations of Show/);
  });

  test("impl with stricter constraints wins over looser one", () => {
    const I = new InterpreterV10();

    const copyable = AST.interfaceDefinition("Copyable", [
      AST.functionSignature(
        "duplicate",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("Self")
      ),
    ]);
    I.evaluate(copyable);

    const itemDef = AST.structDefinition(
      "Item",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named"
    );
    I.evaluate(itemDef);

    const itemDisplayImpl = AST.implementationDefinition(
      "Display",
      AST.simpleTypeExpression("Item"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Item"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.stringInterpolation([
                AST.stringLiteral("Item("),
                AST.memberAccessExpression(AST.identifier("self"), "value"),
                AST.stringLiteral(")"),
              ])
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(itemDisplayImpl);

    const itemCopyableImpl = AST.implementationDefinition(
      "Copyable",
      AST.simpleTypeExpression("Item"),
      [
        AST.functionDefinition(
          "duplicate",
          [AST.functionParameter("self", AST.simpleTypeExpression("Item"))],
          AST.blockExpression([
            AST.returnStatement(AST.identifier("self")),
          ]),
          AST.simpleTypeExpression("Item")
        ),
      ]
    );
    I.evaluate(itemCopyableImpl);

    const wrapperDef = AST.structDefinition(
      "Wrapper",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const genericImpl = AST.implementationDefinition(
      "Display",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("generic")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Display"))])]
    );
    I.evaluate(genericImpl);

    const specificImpl = AST.implementationDefinition(
      "Display",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("copyable")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [
        AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
        AST.interfaceConstraint(AST.simpleTypeExpression("Copyable")),
      ])]
    );
    I.evaluate(specificImpl);

    const itemLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(3), "value"),
    ], false, "Item");

    const wrapperItem = AST.structLiteral([
      AST.structFieldInitializer(itemLiteral, "value")],
      false,
      "Wrapper",
      undefined,
      [AST.simpleTypeExpression("Item")]
    );

    const call = AST.functionCall(
      AST.memberAccessExpression(wrapperItem, "to_string"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "copyable" });
  });

  test("ambiguous methods can be disambiguated via named impls", () => {
    const I = new InterpreterV10();

    const service = AST.structDefinition(
      "Service",
      [],
      "named"
    );
    I.evaluate(service);

    const A = AST.interfaceDefinition("A", [
      AST.functionSignature(
        "act",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    const B = AST.interfaceDefinition("B", [
      AST.functionSignature(
        "act",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(A);
    I.evaluate(B);

    const implA = AST.implementationDefinition(
      "A",
      AST.simpleTypeExpression("Service"),
      [
        AST.functionDefinition(
          "act",
          [AST.functionParameter("self", AST.simpleTypeExpression("Service"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("A")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(implA);

    const implB = AST.implementationDefinition(
      "B",
      AST.simpleTypeExpression("Service"),
      [
        AST.functionDefinition(
          "act",
          [AST.functionParameter("self", AST.simpleTypeExpression("Service"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("B")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(implB);

    const namedA = AST.implementationDefinition(
      "A",
      AST.simpleTypeExpression("Service"),
      [
        AST.functionDefinition(
          "act",
          [AST.functionParameter("self", AST.simpleTypeExpression("Service"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("A.named")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      "ActA"
    );
    I.evaluate(namedA);

    const namedB = AST.implementationDefinition(
      "B",
      AST.simpleTypeExpression("Service"),
      [
        AST.functionDefinition(
          "act",
          [AST.functionParameter("self", AST.simpleTypeExpression("Service"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("B.named")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      "ActB"
    );
    I.evaluate(namedB);

    const svc = AST.structLiteral([], false, "Service");

    const ambiguousCall = AST.functionCall(
      AST.memberAccessExpression(svc, "act"),
      []
    );
    expect(() => I.evaluate(ambiguousCall)).toThrow(/ambiguous implementations of A for Service/);

    const callA = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("ActA"), "act"),
      [svc]
    );
    expect(I.evaluate(callA)).toEqual({ kind: "string", value: "A.named" });

    const callB = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("ActB"), "act"),
      [svc]
    );
    expect(I.evaluate(callB)).toEqual({ kind: "string", value: "B.named" });
  });
});
