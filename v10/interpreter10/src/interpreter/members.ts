import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    tryUfcs(env: Environment, funcName: string, receiver: V10Value): Extract<V10Value, { kind: "bound_method" }> | null;
  }
}

export function applyMemberAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.tryUfcs = function tryUfcs(this: InterpreterV10, env: Environment, funcName: string, receiver: V10Value): Extract<V10Value, { kind: "bound_method" }> | null {
    try {
      const candidate = env.get(funcName);
      if (candidate && candidate.kind === "function") {
        return { kind: "bound_method", func: candidate, self: receiver };
      }
    } catch {}
    return null;
  };
}
