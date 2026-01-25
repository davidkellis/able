import type { RuntimeValue } from "../interpreter/values";
import { applyArithmeticBinary } from "../interpreter/numeric";

export type BytecodeInstruction =
  | { op: "const"; value: RuntimeValue }
  | { op: "load"; slot: number }
  | { op: "store"; slot: number }
  | { op: "pop" }
  | { op: "add" }
  | { op: "jump"; target: number }
  | { op: "jump_if_false"; target: number }
  | { op: "return" };

export type BytecodeProgram = {
  instructions: BytecodeInstruction[];
  locals: number;
};

const NIL_VALUE: RuntimeValue = { kind: "nil", value: null };

export class BytecodeVM {
  private stack: RuntimeValue[] = [];
  private locals: RuntimeValue[] = [];
  private ip = 0;

  run(program: BytecodeProgram): RuntimeValue {
    this.stack = [];
    this.locals = new Array(program.locals).fill(NIL_VALUE);
    this.ip = 0;

    const instructions = program.instructions;
    while (this.ip < instructions.length) {
      const instr = instructions[this.ip]!;
      switch (instr.op) {
        case "const":
          this.stack.push(instr.value);
          this.ip += 1;
          break;
        case "load":
          this.stack.push(this.locals[instr.slot] ?? NIL_VALUE);
          this.ip += 1;
          break;
        case "store": {
          const value = this.pop();
          this.locals[instr.slot] = value;
          this.stack.push(value);
          this.ip += 1;
          break;
        }
        case "pop":
          this.pop();
          this.ip += 1;
          break;
        case "add": {
          const right = this.pop();
          const left = this.pop();
          this.stack.push(applyArithmeticBinary("+", left, right));
          this.ip += 1;
          break;
        }
        case "jump":
          this.ip = instr.target;
          break;
        case "jump_if_false": {
          const value = this.pop();
          const isFalse = value.kind === "bool" ? !value.value : value.kind === "nil";
          this.ip = isFalse ? instr.target : this.ip + 1;
          break;
        }
        case "return":
          return this.pop();
      }
    }
    return NIL_VALUE;
  }

  private pop(): RuntimeValue {
    if (this.stack.length === 0) {
      throw new Error("bytecode stack underflow");
    }
    return this.stack.pop()!;
  }
}
