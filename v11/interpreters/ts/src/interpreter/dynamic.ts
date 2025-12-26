import * as AST from "../ast";
import { mapSourceFile } from "../parser/tree-sitter-mapper";
import { getLoadedTreeSitterParser } from "../parser/tree-sitter-loader";
import { RaiseSignal } from "./signals";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

const NIL: RuntimeValue = { kind: "nil", value: null };

declare module "./index" {
  interface Interpreter {
    dynamicBuiltinsInitialized: boolean;
    dynPackageDefMethod: Extract<RuntimeValue, { kind: "native_function" }>;
    ensureDynamicBuiltins(): void;
    evaluateDynamicDefinition(pkgName: string, source: string): RuntimeValue;
  }
}

const asString = (interp: Interpreter, value: RuntimeValue | undefined, context: string): string | RuntimeValue => {
  if (value && value.kind === "String") {
    return value.value;
  }
  return interp.makeRuntimeError(`${context} expects String`);
};

const resolveDynamicPackage = (base: string[], target: string[]): string[] => {
  if (target.length === 0) return base;
  if (target[0] === "root") return target;
  if (base.length > 0 && target.length >= base.length) {
    const matches = base.every((part, idx) => target[idx] === part);
    if (matches) return target;
  }
  return [...base, ...target];
};

export function applyDynamicAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureDynamicBuiltins = function ensureDynamicBuiltins(this: Interpreter): void {
    if (this.dynamicBuiltinsInitialized) return;
    this.dynamicBuiltinsInitialized = true;

    this.dynPackageDefMethod = this.makeNativeFunction("dyn.Package.def", 2, (interp, args) => {
      const self = args[0];
      if (!self || self.kind !== "dyn_package") {
        return interp.makeRuntimeError("dyn.Package.def called on non-dyn package");
      }
      const source = asString(interp, args[1], "dyn.Package.def");
      if (typeof source !== "string") return source;
      return interp.evaluateDynamicDefinition(self.name, source);
    });

    const packageFn = this.makeNativeFunction("dyn.package", 1, (interp, args) => {
      const name = asString(interp, args[0], "dyn.package");
      if (typeof name !== "string") return name;
      if (!interp.packageRegistry.has(name)) {
        return interp.makeRuntimeError(`dyn.package: package '${name}' not found`);
      }
      return { kind: "dyn_package", name };
    });

    const defPackageFn = this.makeNativeFunction("dyn.def_package", 1, (interp, args) => {
      const name = asString(interp, args[0], "dyn.def_package");
      if (typeof name !== "string") return name;
      if (!interp.packageRegistry.has(name)) {
        interp.packageRegistry.set(name, new Map());
      }
      return { kind: "dyn_package", name };
    });

    const symbols = new Map<string, RuntimeValue>([
      ["package", packageFn],
      ["def_package", defPackageFn],
    ]);

    const dynPackage: RuntimeValue = { kind: "package", name: "dyn", symbols };
    try {
      this.globals.define("dyn", dynPackage);
    } catch {
      // ignore if already defined
    }
  };

  cls.prototype.evaluateDynamicDefinition = function evaluateDynamicDefinition(
    this: Interpreter,
    pkgName: string,
    source: string,
  ): RuntimeValue {
    if (!pkgName) {
      return this.makeRuntimeError("dyn.def requires package name");
    }
    const parser = getLoadedTreeSitterParser();
    if (!parser) {
      return this.makeRuntimeError("dyn.def requires tree-sitter to be initialized");
    }
    let moduleAst: AST.Module | null = null;
    try {
      const tree = parser.parse(source);
      if (tree.rootNode.type !== "source_file") {
        return this.makeRuntimeError(`dyn.def parse error: root ${tree.rootNode.type}`);
      }
      if ((tree.rootNode as unknown as { hasError?: boolean }).hasError) {
        return this.makeRuntimeError("dyn.def parse error: syntax errors");
      }
      moduleAst = mapSourceFile(tree.rootNode, source, `<dyn:${pkgName}>`);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      return this.makeRuntimeError(`dyn.def parse error: ${message}`);
    }
    if (!moduleAst) {
      return this.makeRuntimeError("dyn.def parse error: unable to map source");
    }

    const baseParts = pkgName.split(".").filter((part) => part.length > 0);
    const targetParts = resolveDynamicPackage(
      baseParts,
      moduleAst.package ? moduleAst.package.namePath.map((part) => part.name) : [],
    );
    moduleAst.package = AST.packageStatement(targetParts, moduleAst.package?.isPrivate);

    const prevDynamicMode = this.dynamicDefinitionMode;
    this.dynamicDefinitionMode = true;
    try {
      this.evaluate(moduleAst, this.globals);
    } catch (err) {
      if (err instanceof RaiseSignal) {
        return err.value;
      }
      const message = err instanceof Error ? err.message : String(err);
      return this.makeRuntimeError(`dyn.def error: ${message}`);
    } finally {
      this.dynamicDefinitionMode = prevDynamicMode;
    }
    return NIL;
  };
}
