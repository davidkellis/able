import * as AST from "./ast";
import { Environment } from "./environment";
import type { Interpreter } from "./interpreter";
import type { AbleIterator, AbleValue } from "./runtime";
import { IteratorEnd, createArrayIterator, createRangeIterator, isAbleArray, isAbleRange } from "./runtime";
import { BreakSignal } from "./signals";

export function evaluateWhileLoop(this: Interpreter, node: AST.WhileLoop, environment: Environment): AbleValue {
  const self = this as any;
  let lastValue: AbleValue = { kind: "nil", value: null };
  while (true) {
    const conditionValue = self.evaluate(node.condition, environment);
    if (!self.isTruthy(conditionValue)) break;
    try {
      lastValue = self.evaluate(node.body, environment);
    } catch (signal) {
      if (signal instanceof BreakSignal) {
        if (!signal.label) {
          return signal.value;
        }
        throw signal;
      }
      throw signal;
    }
  }
  return lastValue;
}

export function evaluateForLoop(this: Interpreter, node: AST.ForLoop, environment: Environment): AbleValue {
  const self = this as any;
  const iterableValue = self.evaluate(node.iterable, environment);
  let iterator: AbleIterator;
  if (isAbleArray(iterableValue)) {
    iterator = createArrayIterator(iterableValue);
  } else if (isAbleRange(iterableValue)) {
    iterator = createRangeIterator(iterableValue);
  } else {
    throw new Error(`Interpreter Error: Unsupported iterable type ${iterableValue?.kind ?? typeof iterableValue}`);
  }

  let lastValue: AbleValue = { kind: "nil", value: null };
  while (true) {
    const nextValue = iterator.next();
    if (nextValue === IteratorEnd) {
      break;
    }
    const loopEnv = new Environment(environment);
    self.evaluatePatternAssignment(node.pattern, nextValue, loopEnv, true);
    try {
      lastValue = self.evaluate(node.body, loopEnv);
    } catch (signal) {
      if (signal instanceof BreakSignal) {
        if (!signal.label) {
          return signal.value;
        }
        throw signal;
      }
      throw signal;
    }
  }
  return lastValue;
}
