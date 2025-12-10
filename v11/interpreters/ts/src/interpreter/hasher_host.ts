import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { makeIntegerValue, numericToNumber } from "./numeric";

const FNV_OFFSET = 0x811c9dc5;
const FNV_PRIME = 0x01000193;

const encoder = new TextEncoder();

declare module "./index" {
  interface InterpreterV10 {
    ensureHasherBuiltins(): void;
    hasherBuiltinsInitialized: boolean;
    nextHasherHandle: number;
    hasherStates: Map<number, number>;
  }
}

function expectString(value: V10Value, label: string): string {
  if (value.kind !== "String") {
    throw new Error(`${label} must be a string`);
  }
  return value.value;
}

function updateHash(current: number, bytes: Uint8Array): number {
  let hash = current >>> 0;
  for (const byte of bytes) {
    hash ^= byte;
    hash = Math.imul(hash, FNV_PRIME) >>> 0;
  }
  return hash >>> 0;
}

export function applyHasherHostAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureHasherBuiltins = function ensureHasherBuiltins(this: InterpreterV10): void {
    if (this.hasherBuiltinsInitialized) return;
    this.hasherBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<V10Value, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_hasher_create", () =>
      this.makeNativeFunction("__able_hasher_create", 0, (interp) => {
        const handle = interp.nextHasherHandle++;
        interp.hasherStates.set(handle, FNV_OFFSET);
        return makeIntegerValue("i32", BigInt(handle));
      }),
    );

    defineIfMissing("__able_hasher_write", () =>
      this.makeNativeFunction("__able_hasher_write", 2, (interp, args) => {
        const handle = Math.trunc(numericToNumber(args[0], "hasher handle", { requireSafeInteger: true }));
        if (handle <= 0) throw new Error("Hasher handle must be positive");
        const current = interp.hasherStates.get(handle);
        if (current === undefined) {
          throw new Error("Unknown hasher handle");
        }
        const chunk = expectString(args[1], "bytes");
        const updated = updateHash(current, encoder.encode(chunk));
        interp.hasherStates.set(handle, updated);
        return { kind: "nil", value: null };
      }),
    );

    defineIfMissing("__able_hasher_finish", () =>
      this.makeNativeFunction("__able_hasher_finish", 1, (interp, args) => {
        const handle = Math.trunc(numericToNumber(args[0], "hasher handle", { requireSafeInteger: true }));
        if (handle <= 0) throw new Error("Hasher handle must be positive");
        const current = interp.hasherStates.get(handle);
        if (current === undefined) {
          throw new Error("Unknown hasher handle");
        }
        interp.hasherStates.delete(handle);
        return makeIntegerValue("i32", BigInt(current >>> 0));
      }),
    );
  };
}
