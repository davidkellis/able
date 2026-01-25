import { describe, expect, test } from "bun:test";

import { buildSyntaxErrorDiagnostic } from "../../src/parser/diagnostics";
import { getTreeSitterParser } from "../../src/parser/tree-sitter-loader";
import { formatParserDiagnostic } from "../../scripts/typecheck-utils";

describe("parser diagnostics", () => {
  test("formats syntax errors with location and expectations", async () => {
    const source = "fn main() -> void {";
    const parser = await getTreeSitterParser();
    const tree = parser.parse(source);
    const diagnostic = buildSyntaxErrorDiagnostic(tree.rootNode, "main.able");
    const formatted = formatParserDiagnostic(diagnostic);
    expect(formatted).toContain("parser:");
    expect(formatted).toContain("main.able:1:");
    expect(formatted).toContain("expected");
  });
});
