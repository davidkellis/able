import { describe, it, expect, vi } from "vitest";
import { interpret } from "../interpreter";
import assignmentsAndClosuresModule from "./assignments-and-closures";
import * as AST from "../ast";

// Helper to capture console output
function captureConsole(fn: () => void): string {
  const output: string[] = [];
  const spy = vi.spyOn(console, "log").mockImplementation((...args) => {
    output.push(args.map(String).join(" "));
  });
  try {
    fn();
  } finally {
    spy.mockRestore();
  }
  return output.join("\n");
}

// Descriptive name based on the original module name
describe("Interpreter Sample - assignments-and-closures", () => {
  it("should produce the expected output", () => {
    const output = captureConsole(() => {
      // Cast might be needed if the imported module isn't automatically typed as AST.Module
      interpret(assignmentsAndClosuresModule as AST.Module);
    });
    // Compare captured output against the stored snapshot
    expect(output).toMatchSnapshot();
  });
});
