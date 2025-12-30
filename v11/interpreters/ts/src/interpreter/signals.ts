import type { RuntimeValue } from "./values";

export class ReturnSignal extends Error {
  constructor(public value: RuntimeValue) {
    super("ReturnSignal");
  }
}

export class RaiseSignal extends Error {
  constructor(public value: RuntimeValue) {
    super("RaiseSignal");
  }
}

export class BreakSignal extends Error {
  constructor(public label: string | null, public value: RuntimeValue) {
    super("BreakSignal");
  }
}

export class BreakLabelSignal extends Error {
  constructor(public label: string, public value: RuntimeValue) {
    super("BreakLabelSignal");
  }
}

export class ContinueSignal extends Error {
  constructor(public label: string | null) {
    super("ContinueSignal");
  }
}

export class ProcYieldSignal extends Error {
  constructor() {
    super("ProcYieldSignal");
  }
}

export class GeneratorYieldSignal extends Error {
  constructor() {
    super("GeneratorYieldSignal");
  }
}

export class GeneratorStopSignal extends Error {
  constructor() {
    super("GeneratorStopSignal");
  }
}

export class ExitSignal extends Error {
  constructor(public code: number) {
    super("ExitSignal");
  }
}
