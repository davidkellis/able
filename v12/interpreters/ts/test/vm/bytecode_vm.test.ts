import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { valuesEqual } from "../../src/interpreter/value_equals";
import { BytecodeVM } from "../../src/vm/bytecode";
import { lowerExpression, lowerModule } from "../../src/vm/lowering";

describe("bytecode VM prototype", () => {
  test("executes integer literals", () => {
    const expr = AST.integerLiteral(7);
    const interp = new Interpreter();
    const treeResult = interp.evaluate(expr);
    const program = lowerExpression(expr);
    const vmResult = new BytecodeVM().run(program);
    expect(valuesEqual(vmResult, treeResult)).toBeTrue();
  });

  test("executes block with assignment + add", () => {
    const expr = AST.blockExpression([
      AST.assign("x", AST.integerLiteral(1)),
      AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(2)),
    ]);
    const interp = new Interpreter();
    const treeResult = interp.evaluate(expr);
    const program = lowerExpression(expr);
    const vmResult = new BytecodeVM().run(program);
    expect(valuesEqual(vmResult, treeResult)).toBeTrue();
  });

  test("executes module body with reassignment", () => {
    const module = AST.module([
      AST.assign("x", AST.integerLiteral(2)),
      AST.assignmentExpression("=", AST.identifier("x"), AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(3))),
      AST.identifier("x"),
    ]);
    const interp = new Interpreter();
    const treeResult = interp.evaluate(module);
    const program = lowerModule(module);
    const vmResult = new BytecodeVM().run(program);
    expect(valuesEqual(vmResult, treeResult)).toBeTrue();
  });
});
