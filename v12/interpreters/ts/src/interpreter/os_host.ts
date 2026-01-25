import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { numericToNumber } from "./numeric";
import { ExitSignal } from "./signals";

declare module "./index" {
  interface Interpreter {
    ensureOsBuiltins(): void;
    osBuiltinsInitialized: boolean;
    osArgs: string[];
  }
}

export function applyOsHostAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureOsBuiltins = function ensureOsBuiltins(this: Interpreter): void {
    if (this.osBuiltinsInitialized) return;
    this.osBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<RuntimeValue, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_os_args", () =>
      this.makeNativeFunction("__able_os_args", 0, (interp, args) => {
        if (args.length !== 0) throw new Error("__able_os_args expects no arguments");
        const values = interp.osArgs.map((arg) => ({ kind: "String", value: arg } as RuntimeValue));
        return interp.makeArrayValue(values);
      }),
    );

    defineIfMissing("__able_os_exit", () =>
      this.makeNativeFunction("__able_os_exit", 1, (_interp, args) => {
        if (args.length !== 1) throw new Error("__able_os_exit expects one argument");
        const raw = numericToNumber(args[0], "exit code", { requireSafeInteger: true });
        const code = Math.trunc(raw);
        if (Number.isNaN(code) || !Number.isFinite(code)) {
          throw new Error("exit code must be a finite integer");
        }
        if (code < 0) {
          throw new Error("exit code must be non-negative");
        }
        throw new ExitSignal(code);
      }),
    );
  };
}
