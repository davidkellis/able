import type { Interpreter } from "./index";
import { isIntegerValue, makeIntegerValue, ratioFromFloat } from "./numeric";

const f32Buffer = new ArrayBuffer(4);
const f32View = new DataView(f32Buffer);
const f64Buffer = new ArrayBuffer(8);
const f64View = new DataView(f64Buffer);

function float32Bits(value: number): number {
  f32View.setFloat32(0, value, false);
  return f32View.getUint32(0, false);
}

function float64Bits(value: number): bigint {
  f64View.setFloat64(0, value, false);
  return f64View.getBigUint64(0, false);
}

function expectU64(value: unknown): bigint {
  if (!value || !isIntegerValue(value) || value.kind !== "u64") {
    throw new Error("__able_u64_mul expects u64 inputs");
  }
  return value.value;
}

declare module "./index" {
  interface Interpreter {
    ensureNumericBuiltins(): void;
    numericBuiltinsInitialized?: boolean;
  }
}

export function applyNumericHostAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureNumericBuiltins = function ensureNumericBuiltins(this: Interpreter): void {
    if (this.numericBuiltinsInitialized) return;
    this.numericBuiltinsInitialized = true;

    const defineIfMissing = (name: string, thunk: () => ReturnType<Interpreter["makeNativeFunction"]>) => {
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

    defineIfMissing("__able_f32_bits", () =>
      this.makeNativeFunction("__able_f32_bits", 1, (_interp, args) => {
        if (args.length !== 1) {
          throw new Error("__able_f32_bits expects one argument");
        }
        const value = args[0];
        if (!value || value.kind !== "f32") {
          throw new Error("__able_f32_bits expects f32 input");
        }
        return makeIntegerValue("u32", BigInt(float32Bits(value.value)));
      }),
    );

    defineIfMissing("__able_f64_bits", () =>
      this.makeNativeFunction("__able_f64_bits", 1, (_interp, args) => {
        if (args.length !== 1) {
          throw new Error("__able_f64_bits expects one argument");
        }
        const value = args[0];
        if (!value || value.kind !== "f64") {
          throw new Error("__able_f64_bits expects f64 input");
        }
        return makeIntegerValue("u64", float64Bits(value.value));
      }),
    );

    defineIfMissing("__able_u64_mul", () =>
      this.makeNativeFunction("__able_u64_mul", 2, (_interp, args) => {
        if (args.length !== 2) {
          throw new Error("__able_u64_mul expects two arguments");
        }
        const lhs = expectU64(args[0]);
        const rhs = expectU64(args[1]);
        const result = (lhs * rhs) & ((1n << 64n) - 1n);
        return makeIntegerValue("u64", result);
      }),
    );
  };
}
