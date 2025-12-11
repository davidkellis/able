import type { InterpreterV10 } from "./index";
import { makeIntegerValue, ratioFromFloat } from "./numeric";

declare module "./index" {
  interface InterpreterV10 {
    ensureNumericBuiltins(): void;
    numericBuiltinsInitialized?: boolean;
  }
}

export function applyNumericHostAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureNumericBuiltins = function ensureNumericBuiltins(this: InterpreterV10): void {
    if (this.numericBuiltinsInitialized) return;
    this.numericBuiltinsInitialized = true;

    const defineIfMissing = (name: string, thunk: () => ReturnType<InterpreterV10["makeNativeFunction"]>) => {
      if (this.globals.has(name)) return;
      this.globals.define(name, thunk());
    };

    defineIfMissing("__able_ratio_from_float", () =>
      this.makeNativeFunction("__able_ratio_from_float", 1, (_interp, args) => {
        if (args.length !== 1) {
          throw new Error("__able_ratio_from_float expects one argument");
        }
        const value = args[0];
        if (!value || (value.kind !== "f32" && value.kind !== "f64")) {
          throw new Error("__able_ratio_from_float expects float input");
        }
        const ratioDef = this.ensureRatioStruct();
        const parts = ratioFromFloat(value.value, value.kind);
        return {
          kind: "struct_instance",
          def: ratioDef,
          values: new Map<string, any>([
            ["num", makeIntegerValue("i64", parts.num)],
            ["den", makeIntegerValue("i64", parts.den)],
          ]),
        };
      }),
    );
  };
}
