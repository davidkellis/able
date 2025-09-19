import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - impl resolution semantics", () => {
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

  test("where-clause superset across multiple type params is preferred", () => {
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

    const combiner = AST.interfaceDefinition("Combine", [
      AST.functionSignature(
        "combine",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(combiner);

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

    const pairDef = AST.structDefinition(
      "Pair",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("A"), "left"),
        AST.structFieldDefinition(AST.simpleTypeExpression("B"), "right"),
      ],
      "named",
      [AST.genericParameter("A"), AST.genericParameter("B")]
    );
    I.evaluate(pairDef);

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

    const baseCombine = AST.implementationDefinition(
      "Combine",
      AST.genericTypeExpression(AST.simpleTypeExpression("Pair"), [
        AST.simpleTypeExpression("A"),
        AST.simpleTypeExpression("B"),
      ]),
      [
        AST.functionDefinition(
          "combine",
          [AST.functionParameter("self", AST.simpleTypeExpression("Pair"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.binaryExpression(
                "+",
                AST.functionCall(
                  AST.memberAccessExpression(
                    AST.memberAccessExpression(AST.identifier("self"), "left"),
                    "read"
                  ),
                  []
                ),
                AST.functionCall(
                  AST.memberAccessExpression(
                    AST.memberAccessExpression(AST.identifier("self"), "right"),
                    "read"
                  ),
                  []
                )
              )
            ),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      [AST.genericParameter("A"), AST.genericParameter("B")],
      undefined,
      [
        AST.whereClauseConstraint("A", [AST.interfaceConstraint(AST.simpleTypeExpression("Readable"))]),
        AST.whereClauseConstraint("B", [AST.interfaceConstraint(AST.simpleTypeExpression("Readable"))]),
      ]
    );
    I.evaluate(baseCombine);

    const specificCombine = AST.implementationDefinition(
      "Combine",
      AST.genericTypeExpression(AST.simpleTypeExpression("Pair"), [
        AST.simpleTypeExpression("A"),
        AST.simpleTypeExpression("B"),
      ]),
      [
        AST.functionDefinition(
          "combine",
          [AST.functionParameter("self", AST.simpleTypeExpression("Pair"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "right"),
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
      [AST.genericParameter("A"), AST.genericParameter("B")],
      undefined,
      [
        AST.whereClauseConstraint("A", [AST.interfaceConstraint(AST.simpleTypeExpression("Readable"))]),
        AST.whereClauseConstraint("B", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Readable")),
          AST.interfaceConstraint(AST.simpleTypeExpression("Writable")),
        ]),
      ]
    );
    I.evaluate(specificCombine);

    const fancyPair = AST.structLiteral([
      AST.structFieldInitializer(AST.structLiteral([
        AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
      ], false, "Fancy"), "left"),
      AST.structFieldInitializer(AST.structLiteral([
        AST.structFieldInitializer(AST.stringLiteral("f2"), "label"),
      ], false, "Fancy"), "right"),
    ], false, "Pair", undefined, [
      AST.simpleTypeExpression("Fancy"),
      AST.simpleTypeExpression("Fancy"),
    ]);

    const fancyCombineCall = AST.functionCall(
      AST.memberAccessExpression(fancyPair, "combine"),
      []
    );
    expect(I.evaluate(fancyCombineCall)).toEqual({ kind: "string", value: "write-fancy" });

    const mixedPair = AST.structLiteral([
      AST.structFieldInitializer(AST.structLiteral([
        AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
      ], false, "Fancy"), "left"),
      AST.structFieldInitializer(AST.structLiteral([
        AST.structFieldInitializer(AST.stringLiteral("b"), "label"),
      ], false, "Basic"), "right"),
    ], false, "Pair", undefined, [
      AST.simpleTypeExpression("Fancy"),
      AST.simpleTypeExpression("Basic"),
    ]);

    const mixedCombineCall = AST.functionCall(
      AST.memberAccessExpression(mixedPair, "combine"),
      []
    );
    expect(I.evaluate(mixedCombineCall)).toEqual({ kind: "string", value: "read-fancyread-basic" });
  });

  test("union impl specificity prefers smaller subset", () => {
    const I = new InterpreterV10();

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

    const extraDef = AST.structDefinition(
      "Extra",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(extraDef);

    const pairImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Basic"),
      ]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("pair")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(pairImpl);

    const tripleImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Basic"),
        AST.simpleTypeExpression("Extra"),
      ]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("triple")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(tripleImpl);

    const fancyValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyValue, "to_string"),
      []
    );
    expect(I.evaluate(fancyCall)).toEqual({ kind: "string", value: "pair" });

    const extraValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("e"), "label"),
    ], false, "Extra");

    const extraCall = AST.functionCall(
      AST.memberAccessExpression(extraValue, "to_string"),
      []
    );
    expect(I.evaluate(extraCall)).toEqual({ kind: "string", value: "triple" });
  });

  test("overlapping union impls without subset remain ambiguous", () => {
    const I = new InterpreterV10();

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

    const extraDef = AST.structDefinition(
      "Extra",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(extraDef);

    const pairImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Basic"),
      ]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("pair")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(pairImpl);

    const otherPairImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Extra"),
      ]),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("other")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(otherPairImpl);

    const fancyValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyValue, "to_string"),
      []
    );

    expect(() => I.evaluate(fancyCall)).toThrowError(/candidates:\s*Fancy \| Basic, Fancy \| Extra/);
  });

  test("dynamic interface value uses union-target impl", () => {
    const I = new InterpreterV10();

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

    const fancyImpl = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyImpl);

    const unionImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Basic"),
      ]),
      [
        AST.functionDefinition(
          "describe",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("union")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(unionImpl);

    const fancyValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    // Coerce into interface-typed binding
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(AST.identifier("item"), AST.simpleTypeExpression("Show")),
        fancyValue
      )
    );

    const call = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("item"), "describe"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "union" });
  });

  test("interface inheritance constraints prefer deeper hierarchy", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const fancyShow = AST.interfaceDefinition(
      "FancyShow",
      [
        AST.functionSignature(
          "fancy",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      AST.simpleTypeExpression("Show"),
      undefined,
      [AST.simpleTypeExpression("Show")]
    );
    I.evaluate(fancyShow);

    const shinyShow = AST.interfaceDefinition(
      "ShinyShow",
      [
        AST.functionSignature(
          "shine",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("string")
        ),
      ],
      undefined,
      AST.simpleTypeExpression("Show"),
      undefined,
      [AST.simpleTypeExpression("FancyShow")]
    );
    I.evaluate(shinyShow);

    const fancyBase = AST.structDefinition(
      "FancyBase",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(fancyBase);

    const fancySpecial = AST.structDefinition(
      "FancySpecial",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(fancySpecial);

    const fancyBaseShow = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("FancyBase"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancyBase"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("base")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyBaseShow);

    const fancyBaseFancyShow = AST.implementationDefinition(
      "FancyShow",
      AST.simpleTypeExpression("FancyBase"),
      [
        AST.functionDefinition(
          "fancy",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancyBase"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("fancy-base")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyBaseFancyShow);

    const fancySpecialShow = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("FancySpecial"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancySpecial"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("special")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancySpecialShow);

    const fancySpecialFancyShow = AST.implementationDefinition(
      "FancyShow",
      AST.simpleTypeExpression("FancySpecial"),
      [
        AST.functionDefinition(
          "fancy",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancySpecial"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("fancy-special")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancySpecialFancyShow);

    const shinyImpl = AST.implementationDefinition(
      "ShinyShow",
      AST.simpleTypeExpression("FancySpecial"),
      [
        AST.functionDefinition(
          "shine",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancySpecial"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("shine")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(shinyImpl);

    const container = AST.structDefinition(
      "Wrap",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(container);

    const baseImpl = AST.implementationDefinition(
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
                  "fancy"
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
          AST.interfaceConstraint(AST.simpleTypeExpression("FancyShow")),
        ]),
      ]
    );
    I.evaluate(baseImpl);

    const shinySpecific = AST.implementationDefinition(
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
                  "shine"
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
          AST.interfaceConstraint(AST.simpleTypeExpression("ShinyShow")),
        ]),
      ]
    );
    I.evaluate(shinySpecific);

    const wrappedSpecial = AST.structLiteral([
      AST.structFieldInitializer(
        AST.structLiteral([
          AST.structFieldInitializer(AST.stringLiteral("s"), "label"),
        ], false, "FancySpecial"),
        "value"
      ),
    ], false, "Wrap", undefined, [AST.simpleTypeExpression("FancySpecial")]);

    const call = AST.functionCall(
      AST.memberAccessExpression(wrappedSpecial, "to_string"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "shine" });
  });

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
    expect(() => I.evaluate(call)).toThrow(/Ambiguous method 'to_string' for type 'Wrapper'/);
  });

  test("impl with stricter constraints wins over looser one", () => {
    const I = new InterpreterV10();

    const display = AST.interfaceDefinition("Display", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(display);

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
    expect(() => I.evaluate(ambiguousCall)).toThrow(/Ambiguous method 'act' for type 'Service'/);

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
