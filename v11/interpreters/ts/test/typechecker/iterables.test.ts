import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("TypeChecker iterables", () => {
  test("for-loop typed pattern rejects mismatched Iterable element types", () => {
    const displayInterface = AST.interfaceDefinition("Display", [
      AST.functionSignature(
        "describe",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string"),
      ),
    ]);

    const iterableInterface = AST.interfaceDefinition(
      "Iterable",
      [
        AST.functionSignature(
          "iterator",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("Iterator"),
        ),
      ],
      [AST.genericParameter("T")],
    );

    const iterableType = AST.genericTypeExpression(AST.simpleTypeExpression("Iterable"), [
      AST.simpleTypeExpression("string"),
    ]);

    const makeItemsFn = AST.functionDefinition("make_items", [], AST.blockExpression([AST.nilLiteral()]), iterableType);

    const bindItems = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("items"), iterableType),
      AST.functionCall(AST.identifier("make_items"), []),
    );

    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("item"), AST.simpleTypeExpression("Display")),
      AST.identifier("items"),
      AST.blockExpression([AST.identifier("item")]),
    );

    const moduleAst = AST.module(
      [displayInterface, iterableInterface, makeItemsFn, bindItems, loop],
      [],
      AST.packageStatement(["app"]),
    );

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics.length).toBeGreaterThanOrEqual(1);
    const loopDiag = diagnostics.find((diag) =>
      diag.message.includes("for-loop pattern expects type Display"),
    );
    expect(loopDiag?.message).toContain("got string");
  });

  test("accepts stdlib collection iterables", () => {
    const checker = new TypeChecker();
    const listStruct = AST.structDefinition("List", [], "named", [AST.genericParameter("T")]);
    const linkedListStruct = AST.structDefinition("LinkedList", [], "named", [AST.genericParameter("T")]);
    const lazySeqStruct = AST.structDefinition("LazySeq", [], "named", [AST.genericParameter("T")]);
    const vectorStruct = AST.structDefinition("Vector", [], "named", [AST.genericParameter("T")]);
    const hashSetStruct = AST.structDefinition("HashSet", [], "named", [AST.genericParameter("T")]);
    const dequeStruct = AST.structDefinition("Deque", [], "named", [AST.genericParameter("T")]);
    const queueStruct = AST.structDefinition("Queue", [], "named", [AST.genericParameter("T")]);
    const bitSetStruct = AST.structDefinition("BitSet", [], "named");

    const listLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("string")),
      AST.identifier("items"),
      AST.blockExpression([AST.identifier("value") as unknown as AST.Statement]),
    );
    const linkedListLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("linked"), AST.simpleTypeExpression("i32")),
      AST.identifier("linkedValues"),
      AST.blockExpression([AST.identifier("linked") as unknown as AST.Statement]),
    );
    const lazySeqLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("lazyValue"), AST.simpleTypeExpression("string")),
      AST.identifier("lazyItems"),
      AST.blockExpression([AST.identifier("lazyValue") as unknown as AST.Statement]),
    );
    const vectorLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("item"), AST.simpleTypeExpression("i32")),
      AST.identifier("values"),
      AST.blockExpression([AST.identifier("item") as unknown as AST.Statement]),
    );
    const setLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("entry"), AST.simpleTypeExpression("string")),
      AST.identifier("entries"),
      AST.blockExpression([AST.identifier("entry") as unknown as AST.Statement]),
    );
    const dequeLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("dequeValue"), AST.simpleTypeExpression("string")),
      AST.identifier("dequeItems"),
      AST.blockExpression([AST.identifier("dequeValue") as unknown as AST.Statement]),
    );
    const queueLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("queueValue"), AST.simpleTypeExpression("i32")),
      AST.identifier("queueItems"),
      AST.blockExpression([AST.identifier("queueValue") as unknown as AST.Statement]),
    );
    const bitSetLoop = AST.forLoop(
      AST.typedPattern(AST.identifier("bit"), AST.simpleTypeExpression("i32")),
      AST.identifier("bitset"),
      AST.blockExpression([AST.identifier("bit") as unknown as AST.Statement]),
    );

    const listFn = AST.functionDefinition(
      "consume_list",
      [AST.functionParameter("items", AST.genericTypeExpression(AST.simpleTypeExpression("List"), [AST.simpleTypeExpression("string")]))],
      AST.blockExpression([listLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const linkedListFn = AST.functionDefinition(
      "consume_linked_list",
      [
        AST.functionParameter(
          "linkedValues",
          AST.genericTypeExpression(AST.simpleTypeExpression("LinkedList"), [AST.simpleTypeExpression("i32")]),
        ),
      ],
      AST.blockExpression([linkedListLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const lazySeqFn = AST.functionDefinition(
      "consume_lazy_seq",
      [
        AST.functionParameter(
          "lazyItems",
          AST.genericTypeExpression(AST.simpleTypeExpression("LazySeq"), [AST.simpleTypeExpression("string")]),
        ),
      ],
      AST.blockExpression([lazySeqLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const vectorFn = AST.functionDefinition(
      "consume_vector",
      [AST.functionParameter("values", AST.genericTypeExpression(AST.simpleTypeExpression("Vector"), [AST.simpleTypeExpression("i32")]))],
      AST.blockExpression([vectorLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const setFn = AST.functionDefinition(
      "consume_hash_set",
      [AST.functionParameter("entries", AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [AST.simpleTypeExpression("string")]))],
      AST.blockExpression([setLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const dequeFn = AST.functionDefinition(
      "consume_deque",
      [AST.functionParameter("dequeItems", AST.genericTypeExpression(AST.simpleTypeExpression("Deque"), [AST.simpleTypeExpression("string")]))],
      AST.blockExpression([dequeLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const queueFn = AST.functionDefinition(
      "consume_queue",
      [AST.functionParameter("queueItems", AST.genericTypeExpression(AST.simpleTypeExpression("Queue"), [AST.simpleTypeExpression("i32")]))],
      AST.blockExpression([queueLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );
    const bitSetFn = AST.functionDefinition(
      "consume_bit_set",
      [AST.functionParameter("bitset", AST.simpleTypeExpression("BitSet"))],
      AST.blockExpression([bitSetLoop, AST.returnStatement(AST.integerLiteral(0))]),
      AST.simpleTypeExpression("i32"),
    );

    const moduleAst = AST.module(
      [
        listStruct,
        linkedListStruct,
        lazySeqStruct,
        vectorStruct,
        hashSetStruct,
        dequeStruct,
        queueStruct,
        bitSetStruct,
        listFn,
        linkedListFn,
        lazySeqFn,
        vectorFn,
        setFn,
        dequeFn,
        queueFn,
        bitSetFn,
      ],
      [],
      AST.packageStatement(["app"]),
    );

    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);
  });
});
