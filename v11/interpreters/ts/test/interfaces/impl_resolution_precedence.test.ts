import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - impl resolution semantics", () => {
  test("inherent methods take precedence over impl methods", () => {
    const I = new InterpreterV10();

    const speakable = AST.interfaceDefinition("Speakable", [
      AST.functionSignature(
        "speak",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(speakable);

    const botDef = AST.structDefinition(
      "Bot",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "id")],
      "named"
    );
    I.evaluate(botDef);

    const botMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("Bot"),
      [
        AST.functionDefinition(
          "speak",
          [AST.functionParameter("self", AST.simpleTypeExpression("Bot"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("beep inherent")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(botMethods);

    const botImpl = AST.implementationDefinition(
      "Speakable",
      AST.simpleTypeExpression("Bot"),
      [
        AST.functionDefinition(
          "speak",
          [AST.functionParameter("self", AST.simpleTypeExpression("Bot"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("beep impl")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(botImpl);

    const botLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(42), "id")],
      false,
      "Bot"
    );

    const call = AST.functionCall(AST.memberAccessExpression(botLiteral, "speak"), []);
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "beep inherent" });
  });

  test("more specific impl wins over generic", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

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
            AST.returnStatement(AST.stringLiteral("Point inherent")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(pointMethods);

    const wrapperDef = AST.structDefinition(
      "Wrapper",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const genericImpl = AST.implementationDefinition(
      "Show",
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
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
    );
    I.evaluate(genericImpl);

    const specificImpl = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("Point")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("specific")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(specificImpl);

    const pointLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
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
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "specific" });
  });

  test("impl with superset constraints is more specific", () => {
    const I = new InterpreterV10();

    const traitA = AST.interfaceDefinition("TraitA", [
      AST.functionSignature(
        "trait_a",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(traitA);

    const traitB = AST.interfaceDefinition("TraitB", [
      AST.functionSignature(
        "trait_b",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(traitB);

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const wrapperDef = AST.structDefinition(
      "Wrapper",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const fancyTraitA = AST.implementationDefinition(
      "TraitA",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "trait_a",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("A:Fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyTraitA);

    const fancyTraitB = AST.implementationDefinition(
      "TraitB",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "trait_b",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("B:Fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyTraitB);

    const basicTraitA = AST.implementationDefinition(
      "TraitA",
      AST.simpleTypeExpression("Basic"),
      [
        AST.functionDefinition(
          "trait_a",
          [AST.functionParameter("self", AST.simpleTypeExpression("Basic"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("A:Basic")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(basicTraitA);

    const genericShow = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "value"),
                  "trait_a"
                ),
                []
              )
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("TraitA"))])]
    );
    I.evaluate(genericShow);

    const specificShow = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "value"),
                  "trait_b"
                ),
                []
              )
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [
        AST.genericParameter("T", [
          AST.interfaceConstraint(AST.simpleTypeExpression("TraitA")),
          AST.interfaceConstraint(AST.simpleTypeExpression("TraitB")),
        ]),
      ]
    );
    I.evaluate(specificShow);

    const fancyLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    const fancyWrapper = AST.structLiteral([
      AST.structFieldInitializer(fancyLiteral, "value"),
    ], false, "Wrapper", undefined, [AST.simpleTypeExpression("Fancy")]);

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyWrapper, "to_string"),
      []
    );
    expect(I.evaluate(fancyCall)).toEqual({ kind: "string", value: "B:Fancy" });

    const basicLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("b"), "label"),
    ], false, "Basic");

    const basicWrapper = AST.structLiteral([
      AST.structFieldInitializer(basicLiteral, "value"),
    ], false, "Wrapper", undefined, [AST.simpleTypeExpression("Basic")]);

    const basicCall = AST.functionCall(
      AST.memberAccessExpression(basicWrapper, "to_string"),
      []
    );
    expect(I.evaluate(basicCall)).toEqual({ kind: "string", value: "A:Basic" });
  });

  test("impl with where-clause superset constraints is more specific", () => {
    const I = new InterpreterV10();

    const readable = AST.interfaceDefinition("Readable", [
      AST.functionSignature(
        "read",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(readable);

    const writable = AST.interfaceDefinition("Writable", [
      AST.functionSignature(
        "write",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(writable);

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const wrapperDef = AST.structDefinition(
      "Wrap",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const fancyReadable = AST.implementationDefinition(
      "Readable",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "read",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("read-fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyReadable);

    const fancyWritable = AST.implementationDefinition(
      "Writable",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "write",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("write-fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyWritable);

    const basicReadable = AST.implementationDefinition(
      "Readable",
      AST.simpleTypeExpression("Basic"),
      [
        AST.functionDefinition(
          "read",
          [AST.functionParameter("self", AST.simpleTypeExpression("Basic"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("read-basic")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(basicReadable);

    const baseShow = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrap"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrap"))],
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
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrap"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrap"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "value"),
                  "write"
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
          AST.interfaceConstraint(AST.simpleTypeExpression("Writable")),
        ]),
      ]
    );
    I.evaluate(specificShow);

    const fancyWrap = AST.structLiteral([
      AST.structFieldInitializer(
        AST.structLiteral([
          AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
        ], false, "Fancy"),
        "value"
      ),
    ], false, "Wrap", undefined, [AST.simpleTypeExpression("Fancy")]);

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyWrap, "to_string"),
      []
    );
    expect(I.evaluate(fancyCall)).toEqual({ kind: "string", value: "write-fancy" });

    const basicWrap = AST.structLiteral([
      AST.structFieldInitializer(
        AST.structLiteral([
          AST.structFieldInitializer(AST.stringLiteral("b"), "label"),
        ], false, "Basic"),
        "value"
      ),
    ], false, "Wrap", undefined, [AST.simpleTypeExpression("Basic")]);

    const basicCall = AST.functionCall(
      AST.memberAccessExpression(basicWrap, "to_string"),
      []
    );
    expect(I.evaluate(basicCall)).toEqual({ kind: "string", value: "read-basic" });
  });

});
