import type { V10Value } from "./values";

export class ReturnSignal extends Error {
  constructor(public value: V10Value) {
    super("ReturnSignal");
  }
}

export class RaiseSignal extends Error {
  constructor(public value: V10Value) {
    super("RaiseSignal");
  }
}

export class BreakSignal extends Error {
  constructor(public label: string | null, public value: V10Value) {
    super("BreakSignal");
  }
}

export class BreakLabelSignal extends Error {
  constructor(public label: string, public value: V10Value) {
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
