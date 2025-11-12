import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { moduleToSource } from "../../scripts/export-fixtures";

describe("export-fixtures printer", () => {
  test("includes interface self type pattern in generated source", () => {
    const iface = AST.interfaceDefinition(
      "Display",
      [
        AST.functionSignature(
          "show",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("string"),
        ),
      ],
      undefined,
      AST.simpleTypeExpression("Point"),
    );
    const module = AST.module([iface]);
    const output = moduleToSource(module).trim();
    expect(output).toContain("interface Display for Point");
    expect(output).toContain("fn show (self: Self) -> string");
  });

  test("prints higher-kinded self type patterns with wildcard arguments", () => {
    const iface = AST.interfaceDefinition(
      "Mapper",
      [
        AST.functionSignature(
          "map",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("value", AST.simpleTypeExpression("T")),
          ],
          AST.simpleTypeExpression("Self"),
          [AST.genericParameter("T")],
        ),
      ],
      [AST.genericParameter("T")],
      AST.genericTypeExpression(AST.simpleTypeExpression("M"), [AST.wildcardTypeExpression()]),
    );
    const module = AST.module([iface]);
    const output = moduleToSource(module).trim();
    expect(output).toContain("interface Mapper <T> for M _");
    expect(output).toContain("fn map <T> (self: Self, value: T) -> Self");
  });
});
