import { describe, it, expect, vi } from "vitest";
import { interpret } from "../interpreter";
import exceptionsModule from "./exceptions"; // Import the AST module
import * as AST from "../ast";

// Helper to capture console output (including errors)
function captureOutput(fn: () => void): { stdout: string; stderr: string } {
  const stdout: string[] = [];
  const stderr: string[] = [];
  const logSpy = vi.spyOn(console, "log").mockImplementation((...args) => {
    stdout.push(args.map(String).join(" "));
  });
  const errorSpy = vi.spyOn(console, "error").mockImplementation((...args) => {
    stderr.push(args.map(String).join(" "));
  });

  try {
    fn();
  } finally {
    logSpy.mockRestore();
    errorSpy.mockRestore();
  }
  return { stdout: stdout.join("\n"), stderr: stderr.join("\n") };
}

// Descriptive name based on the original module name
describe("Interpreter Sample - exceptions", () => {
  it("should produce the expected output (including errors)", () => {
    // Exceptions might print to console.error, so capture both
    const output = captureOutput(() => {
      // Cast might be needed
      interpret(exceptionsModule as AST.Module);
    });
    // Snapshot both stdout and stderr
    expect(output.stdout).toMatchSnapshot("stdout");
    expect(output.stderr).toMatchSnapshot("stderr");
  });
});
