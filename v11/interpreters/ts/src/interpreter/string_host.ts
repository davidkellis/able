import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerValue, numericToNumber } from "./numeric";

const encoder = new TextEncoder();
const decoder = new TextDecoder();

declare module "./index" {
  interface Interpreter {
    ensureStringHostBuiltins(): void;
    stringHostBuiltinsInitialized: boolean;
  }
}

function expectString(value: RuntimeValue, label: string): string {
  if (value.kind !== "String") {
    throw new Error(`${label} must be a string`);
  }
  return value.value;
}

function expectArray(interp: Interpreter, value: RuntimeValue, label: string): Extract<RuntimeValue, { kind: "array" }> {
  if (value.kind === "array") {
    return value;
  }
  if (value.kind === "struct_instance" && value.def.id.name === "Array") {
    let handleVal: RuntimeValue | undefined;
    if (value.values instanceof Map) {
      handleVal = value.values.get("storage_handle");
    } else if (Array.isArray(value.values)) {
      handleVal = (value.values as RuntimeValue[])[2];
    }
    let handle = 0;
    if (handleVal) {
      try {
        handle = Math.trunc(numericToNumber(handleVal, "array handle", { requireSafeInteger: true }));
      } catch {
        handle = 0;
      }
    }
    if (handle) {
      const state = interp.arrayStates.get(handle);
      if (state) {
        return interp.makeArrayValue(state.values, state.capacity);
      }
    }
  }
  throw new Error(`${label} must be an array`);
}

function expectNumeric(value: RuntimeValue, label: string): number {
  return Math.trunc(numericToNumber(value, label, { requireSafeInteger: true }));
}

function toByte(value: RuntimeValue, index: number): number {
  const num = expectNumeric(value, `Array element ${index}`);
  if (Number.isNaN(num) || num < 0 || num > 0xff) {
    throw new Error(`Array element ${index} must be in range 0..255`);
  }
  return num;
}

export function applyStringHostAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureStringHostBuiltins = function ensureStringHostBuiltins(this: Interpreter): void {
    if (this.stringHostBuiltinsInitialized) return;
    this.stringHostBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<RuntimeValue, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_String_from_builtin", () =>
      this.makeNativeFunction("__able_String_from_builtin", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_String_from_builtin expects one argument");
        const input = expectString(args[0], "String");
        const encoded = encoder.encode(input);
        const elements = Array.from(encoded, (byte): RuntimeValue => makeIntegerValue("u8", BigInt(byte)));
        return this.makeArrayValue(elements);
      }),
    );

    defineIfMissing("__able_String_to_builtin", () =>
      this.makeNativeFunction("__able_String_to_builtin", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_String_to_builtin expects one argument");
        const arr = expectArray(this, args[0], "bytes array");
        const bytes = Uint8Array.from(arr.elements.map((element, idx) => toByte(element, idx)));
        let decoded: string;
        try {
          decoded = decoder.decode(bytes);
        } catch (e) {
          const message = e instanceof Error ? e.message : "invalid UTF-8 bytes";
          throw new Error(message);
        }
        return { kind: "String", value: decoded };
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
