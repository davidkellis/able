import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

const encoder = new TextEncoder();
const decoder = new TextDecoder();

declare module "./index" {
  interface InterpreterV10 {
    ensureStringHostBuiltins(): void;
    stringHostBuiltinsInitialized: boolean;
  }
}

function expectString(value: V10Value, label: string): string {
  if (value.kind !== "string") {
    throw new Error(`${label} must be a string`);
  }
  return value.value;
}

function expectArray(value: V10Value, label: string): Extract<V10Value, { kind: "array" }> {
  if (value.kind !== "array") {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function expectNumeric(value: V10Value, label: string): number {
  if (value.kind === "i32" || value.kind === "f64") {
    return Math.trunc(value.value);
  }
  throw new Error(`${label} must be numeric`);
}

function toByte(value: V10Value, index: number): number {
  const num = expectNumeric(value, `Array element ${index}`);
  if (Number.isNaN(num) || num < 0 || num > 0xff) {
    throw new Error(`Array element ${index} must be in range 0..255`);
  }
  return num;
}

export function applyStringHostAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureStringHostBuiltins = function ensureStringHostBuiltins(this: InterpreterV10): void {
    if (this.stringHostBuiltinsInitialized) return;
    this.stringHostBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<V10Value, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_string_from_builtin", () =>
      this.makeNativeFunction("__able_string_from_builtin", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_string_from_builtin expects one argument");
        const input = expectString(args[0], "string");
        const encoded = encoder.encode(input);
        const elements = Array.from(encoded, (byte): V10Value => ({ kind: "i32", value: byte }));
        return { kind: "array", elements };
      }),
    );

    defineIfMissing("__able_string_to_builtin", () =>
      this.makeNativeFunction("__able_string_to_builtin", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_string_to_builtin expects one argument");
        const arr = expectArray(args[0], "bytes array");
        const bytes = Uint8Array.from(arr.elements.map((element, idx) => toByte(element, idx)));
        let decoded: string;
        try {
          decoded = decoder.decode(bytes);
        } catch (e) {
          const message = e instanceof Error ? e.message : "invalid UTF-8 bytes";
          throw new Error(message);
        }
        return { kind: "string", value: decoded };
      }),
    );

    defineIfMissing("__able_char_from_codepoint", () =>
      this.makeNativeFunction("__able_char_from_codepoint", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_char_from_codepoint expects one argument");
        const codepoint = expectNumeric(args[0], "codepoint");
        if (codepoint < 0 || codepoint > 0x10ffff) {
          throw new Error("codepoint must be within Unicode range 0..0x10FFFF");
        }
        return { kind: "char", value: String.fromCodePoint(codepoint) };
      }),
    );
  };
}
