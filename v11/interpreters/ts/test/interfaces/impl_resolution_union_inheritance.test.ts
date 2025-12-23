import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - impl resolution semantics", () => {
  test("where-clause superset across multiple type params is preferred", () => {
    const I = new Interpreter();

    const readable = AST.interfaceDefinition("Readable", [
      AST.functionSignature(
        "read",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(readable);

    const writable = AST.interfaceDefinition("Writable", [
      AST.functionSignature(
        "write",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(writable);

    const combiner = AST.interfaceDefinition("Combine", [
      AST.functionSignature(
        "combine",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(combiner);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
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
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
    expect(I.evaluate(fancyCombineCall)).toEqual({ kind: "String", value: "write-fancy" });

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
    expect(I.evaluate(mixedCombineCall)).toEqual({ kind: "String", value: "read-fancyread-basic" });
  });

  test("union impl specificity prefers smaller subset", () => {
    const I = new Interpreter();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_String",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const extraDef = AST.structDefinition(
      "Extra",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
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
          "to_String",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("pair")),
          ]),
          AST.simpleTypeExpression("String")
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
          "to_String",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("triple")),
          ]),
          AST.simpleTypeExpression("String")
        ),
      ]
    );
    I.evaluate(tripleImpl);

    const fancyValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyValue, "to_String"),
      []
    );
    expect(I.evaluate(fancyCall)).toEqual({ kind: "String", value: "pair" });

    const extraValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("e"), "label"),
    ], false, "Extra");

    const extraCall = AST.functionCall(
      AST.memberAccessExpression(extraValue, "to_String"),
      []
    );
    expect(I.evaluate(extraCall)).toEqual({ kind: "String", value: "triple" });
  });

  test("overlapping union impls without subset remain ambiguous", () => {
    const I = new Interpreter();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_String",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const extraDef = AST.structDefinition(
      "Extra",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
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
          "to_String",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("pair")),
          ]),
          AST.simpleTypeExpression("String")
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
          "to_String",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("other")),
          ]),
          AST.simpleTypeExpression("String")
        ),
      ]
    );
    I.evaluate(otherPairImpl);

    const fancyValue = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
    ], false, "Fancy");

    const fancyCall = AST.functionCall(
      AST.memberAccessExpression(fancyValue, "to_String"),
      []
    );

    expect(() => I.evaluate(fancyCall)).toThrowError(/ambiguous implementations of Show.*Fancy \| Basic.*Fancy \| Extra/);
  });

  test("dynamic interface value uses union-target impl", () => {
    const I = new Interpreter();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_String",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const fancyImpl = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "to_String",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("fancy")),
          ]),
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
    expect(I.evaluate(call)).toEqual({ kind: "String", value: "union" });
  });

  test("interface inheritance constraints prefer deeper hierarchy", () => {
    const I = new Interpreter();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_String",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    I.evaluate(show);

    const fancyShow = AST.interfaceDefinition(
      "FancyShow",
      [
        AST.functionSignature(
          "fancy",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancyBase);

    const fancySpecial = AST.structDefinition(
      "FancySpecial",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "label")],
      "named"
    );
    I.evaluate(fancySpecial);

    const fancyBaseShow = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("FancyBase"),
      [
        AST.functionDefinition(
          "to_String",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancyBase"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("base")),
          ]),
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
        ),
      ]
    );
    I.evaluate(fancyBaseFancyShow);

    const fancySpecialShow = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("FancySpecial"),
      [
        AST.functionDefinition(
          "to_String",
          [AST.functionParameter("self", AST.simpleTypeExpression("FancySpecial"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("special")),
          ]),
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
          AST.simpleTypeExpression("String")
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
          "to_String",
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
          AST.simpleTypeExpression("String")
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
          "to_String",
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
          AST.simpleTypeExpression("String")
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
      AST.memberAccessExpression(wrappedSpecial, "to_String"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "String", value: "shine" });
  });

});
