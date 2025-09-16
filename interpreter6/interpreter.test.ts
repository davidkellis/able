import { describe, it, expect, vi } from "vitest";
import { interpret } from "./interpreter";
import functionsSampleModule from "./sample/functions";
import loopsModule from "./sample/loops";
import conditionalsModule from "./sample/conditionals";
import functionsWithDestructuringModule from "./sample/functions-with-destructuring";
import arithmeticModule from "./sample/arithmetic";
import assignmentsModule from "./sample/assignments";

// Helper to capture console output
function captureConsole(fn: () => void): string[] {
  const output: string[] = [];
  const spy = vi.spyOn(console, "log").mockImplementation((...args) => {
    output.push(args.join(" "));
  });
  try {
    fn();
  } finally {
    spy.mockRestore();
  }
  return output;
}

describe("Able Interpreter Sample Programs", () => {
  it("should run the functions sample and print expected outputs", () => {
    const output = captureConsole(() => interpret(functionsSampleModule));
    expect(output.some((line) => line.includes("Test 1: add(2, 3) = 5"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 2: inner(5) = 10"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 3: factorial(5) = 120"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 4: sumThree(1, 2, 3) = 6"))).toBeTruthy();
  });

  it("should run the loops sample and print expected loop traces", () => {
    const output = captureConsole(() => interpret(loopsModule));
    expect(output.some((line) => line.includes("Test 1: While Loop Start"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 1: While Loop End"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 2: For Loop (Array) Start"))).toBeTruthy();
    expect(output.some((line) => line.includes("For item: apple"))).toBeTruthy();
    expect(output.some((line) => line.includes("For item: banana"))).toBeTruthy();
    expect(output.some((line) => line.includes("For item: cherry"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 3: For Loop (Inclusive Range 1..3) Start"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 3: Range Sum = 6"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 4: For Loop (Exclusive Range 5...8) Start"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 4: Range Sum = 18"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 6: While Loop with Break Start"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 6: While Loop with Break End"))).toBeTruthy();
  });

  it("should run the conditionals sample and print expected conditional traces", () => {
    const output = captureConsole(() => interpret(conditionalsModule));
    expect(output.some((line) => line.includes("Test 1: Simple if (true) executed"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 3: if/or - First branch executed"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 4: if/or - Second branch executed"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 5: if/or/else - Else branch executed"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 6: if result = Result A"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 7: Outer if true"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 7: Inner else executed"))).toBeTruthy();
  });

  it("should run the functions-with-destructuring sample and print expected destructuring traces", () => {
    const output = captureConsole(() => interpret(functionsWithDestructuringModule));
    expect(output.some((line) => line.includes("Test 1: Struct Destructuring"))).toBeTruthy();
    expect(output.some((line) => line.includes("x = 3"))).toBeTruthy();
    expect(output.some((line) => line.includes("y = 4"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 2: Array Destructuring"))).toBeTruthy();
    expect(output.some((line) => line.includes("first = 5"))).toBeTruthy();
    expect(output.some((line) => line.includes("second = 6"))).toBeTruthy();
    expect(output.some((line) => line.includes("rest = [7]"))).toBeTruthy();
  });

  it("should run the arithmetic sample and print expected arithmetic traces", () => {
    const output = captureConsole(() => interpret(arithmeticModule));
    expect(output.some((line) => line.includes("1 + 2: 3"))).toBeTruthy();
    expect(output.some((line) => line.includes("5 - 3: 2"))).toBeTruthy();
    expect(output.some((line) => line.includes("4 * 6: 24"))).toBeTruthy();
    expect(output.some((line) => line.includes("10 / 2: 5"))).toBeTruthy();
    expect(output.some((line) => line.includes("11 / 3: 3"))).toBeTruthy();
    expect(output.some((line) => line.includes("10 % 3: 1"))).toBeTruthy();
    expect(output.some((line) => line.includes("-5: -5"))).toBeTruthy();
    expect(output.some((line) => line.includes("2 + 3 * 4: 14"))).toBeTruthy();
    expect(output.some((line) => line.includes("(2 + 3) * 4: 20"))).toBeTruthy();
  });

  it("should run the assignments sample and print expected assignment traces", () => {
    const output = captureConsole(() => interpret(assignmentsModule));
    expect(output.some((line) => line.includes("Test 1: Simple Assignment"))).toBeTruthy();
    expect(output.some((line) => line.includes("x = 42"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 2: Reassignment"))).toBeTruthy();
    expect(output.some((line) => line.includes("y = 20"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 3: Destructuring Assignment (Struct)"))).toBeTruthy();
    expect(output.some((line) => line.includes("a = 1"))).toBeTruthy();
    expect(output.some((line) => line.includes("b = 2"))).toBeTruthy();
    expect(output.some((line) => line.includes("Test 4: Destructuring Assignment (Array)"))).toBeTruthy();
    expect(output.some((line) => line.includes("first = 3"))).toBeTruthy();
    expect(output.some((line) => line.includes("second = 4"))).toBeTruthy();
    expect(output.some((line) => line.includes("rest = [5]"))).toBeTruthy();
  });
});
