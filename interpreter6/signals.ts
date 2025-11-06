import type { AbleValue } from "./runtime";

export class ReturnSignal extends Error {
  constructor(public value: AbleValue) {
    super(`ReturnSignal: ${JSON.stringify(value)}`);
    this.name = "ReturnSignal";
  }
}

export class RaiseSignal extends Error {
  constructor(public value: AbleValue) {
    super(`RaiseSignal: ${JSON.stringify(value)}`);
    this.name = "RaiseSignal";
  }
}

export class BreakSignal extends Error {
  constructor(public label: string, public value: AbleValue) {
    super(`BreakSignal: '${label}' with ${JSON.stringify(value)}`);
    this.name = "BreakSignal";
  }
}
