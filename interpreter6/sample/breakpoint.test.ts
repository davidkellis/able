import { describe, it, expect, vi } from "vitest";
import { interpret } from "../interpreter";
import breakpointModule from "./breakpoint";
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
    // Interpret catches the signal, so no error should propagate here
    fn();
  } finally {
    logSpy.mockRestore();
    errorSpy.mockRestore();
  }
  return { stdout: stdout.join("\n"), stderr: stderr.join("\n") };
}

describe("Interpreter Sample - breakpoint", () => {
  it("should pause and continue execution past breakpoint", () => {
    const output = captureOutput(() => {
      interpret(breakpointModule as AST.Module);
    });

    // Check that the message before the breakpoint is logged
    expect(output.stdout).toContain("Before breakpoint");

    // Check the log message from the breakpoint expression itself
    expect(output.stdout).toContain("--- Breakpoint Reached (Execution will pause if debugger attached) ---");

    // Check that the print(x) part WAS executed after the breakpoint
    expect(output.stdout).toContain("10");

    // Stderr should be empty in this case (no uncaught exceptions)
    expect(output.stderr).toBe("");

  });
});
