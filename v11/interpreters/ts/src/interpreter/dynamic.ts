import * as AST from "../ast";
import { mapSourceFile } from "../parser/tree-sitter-mapper";
import { getLoadedTreeSitterParser } from "../parser/tree-sitter-loader";
import { makeIntegerValue } from "./numeric";
import { RaiseSignal } from "./signals";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

const NIL: RuntimeValue = { kind: "nil", value: null };

declare module "./index" {
  interface Interpreter {
    dynamicBuiltinsInitialized: boolean;
    dynPackageDefMethod: Extract<RuntimeValue, { kind: "native_function" }>;
    dynPackageEvalMethod: Extract<RuntimeValue, { kind: "native_function" }>;
    ensureDynamicBuiltins(): void;
    evaluateDynamicDefinition(pkgName: string, source: string): RuntimeValue;
    evaluateDynamicEval(pkgName: string, source: string): RuntimeValue;
  }
}

const asString = (interp: Interpreter, value: RuntimeValue | undefined, context: string): string | RuntimeValue => {
  if (value && value.kind === "String") {
    return value.value;
  }
  if (value && value.kind === "struct_instance" && value.def.id.name === "String") {
    let bytesVal: RuntimeValue | undefined;
    if (value.values instanceof Map) {
      bytesVal = value.values.get("bytes");
    } else if (Array.isArray(value.values)) {
      bytesVal = value.values[0];
    }
    const toBuiltin = bytesVal ? interp.globals.get("__able_String_to_builtin") : undefined;
    if (bytesVal && toBuiltin && toBuiltin.kind === "native_function") {
      const result = toBuiltin.impl(interp, [bytesVal]);
      if (result && (result as RuntimeValue).kind === "String") {
        return (result as Extract<RuntimeValue, { kind: "String" }>).value;
      }
    }
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

const PARSE_ERROR_MESSAGE = "parse error: syntax errors";
const textEncoder = new TextEncoder();

type ParseErrorInfo = {
  message: string;
  startByte: number;
  endByte: number;
  isIncomplete: boolean;
};

const isWhitespaceByte = (value: number): boolean =>
  value === 0x20 || value === 0x09 || value === 0x0a || value === 0x0d || value === 0x0b || value === 0x0c;

const lastNonWhitespaceByte = (source: string): number => {
  const bytes = textEncoder.encode(source);
  let idx = bytes.length;
  while (idx > 0 && isWhitespaceByte(bytes[idx - 1]!)) {
    idx -= 1;
  }
  return idx;
};

const parseErrorInfoFromTree = (root: { childCount: number; child: (idx: number) => any }, source: string): ParseErrorInfo => {
  const endByte = lastNonWhitespaceByte(source);
  let firstError: any = null;
  let incomplete: any = null;
  const stack: any[] = [root];
  while (stack.length > 0) {
    const node = stack.pop();
    if (!node) continue;
    const isMissing = typeof node.isMissing === "function" ? node.isMissing() : node.isMissing;
    const isError = node.type === "ERROR";
    if (isMissing || isError) {
      if (!firstError) firstError = node;
      const start = typeof node.startIndex === "number" ? node.startIndex : 0;
      const end = typeof node.endIndex === "number" ? node.endIndex : start;
      if ((isMissing && start >= endByte) || (isError && end >= endByte)) {
        incomplete = node;
      }
    }
    for (let i = 0; i < node.childCount; i += 1) {
      const child = node.child(i);
      if (child) stack.push(child);
    }
  }
  const target = incomplete ?? firstError ?? root;
  const start = typeof target.startIndex === "number" ? target.startIndex : 0;
  const end = typeof target.endIndex === "number" ? target.endIndex : start;
  return {
    message: PARSE_ERROR_MESSAGE,
    startByte: start,
    endByte: end,
    isIncomplete: Boolean(incomplete),
  };
};

const resolveStructDef = (interp: Interpreter, name: string, fields: AST.StructFieldDefinition[]): AST.StructDefinition => {
  const candidates = [
    name,
    `able.core.errors.${name}`,
    `core.errors.${name}`,
    `errors.${name}`,
  ];
  for (const candidate of candidates) {
    try {
      const val = interp.globals.get(candidate);
      if (val && val.kind === "struct_def") return val.def;
    } catch {
      // ignore lookup errors
    }
  }
  for (const bucket of interp.packageRegistry.values()) {
    const val = bucket.get(name);
    if (val && val.kind === "struct_def") return val.def;
  }
  return AST.structDefinition(name, fields, "named");
};

const buildParseErrorValue = (
  interp: Interpreter,
  message: string,
  startByte: number,
  endByte: number,
  isIncomplete: boolean,
): RuntimeValue => {
  const spanDef = resolveStructDef(
    interp,
    "Span",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("u64"), "start"),
      AST.structFieldDefinition(AST.simpleTypeExpression("u64"), "end"),
    ],
  );
  const parseErrorDef = resolveStructDef(
    interp,
    "ParseError",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("String"), "message"),
      AST.structFieldDefinition(AST.simpleTypeExpression("Span"), "span"),
      AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "is_incomplete"),
    ],
  );
  const spanValue = interp.makeNamedStructInstance(spanDef, [
    ["start", makeIntegerValue("u64", BigInt(Math.max(0, startByte)))],
    ["end", makeIntegerValue("u64", BigInt(Math.max(0, endByte)))],
  ]);
  const parseErrorValue = interp.makeNamedStructInstance(parseErrorDef, [
    ["message", { kind: "String", value: message }],
    ["span", spanValue],
    ["is_incomplete", { kind: "bool", value: isIncomplete }],
  ]);
  return interp.makeRuntimeError(message, parseErrorValue);
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

    this.dynPackageEvalMethod = this.makeNativeFunction("dyn.Package.eval", 2, (interp, args) => {
      const self = args[0];
      if (!self || self.kind !== "dyn_package") {
        return interp.makeRuntimeError("dyn.Package.eval called on non-dyn package");
      }
      const source = asString(interp, args[1], "dyn.Package.eval");
      if (typeof source !== "string") return source;
      return interp.evaluateDynamicEval(self.name, source);
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

    const evalFn = this.makeNativeFunction("dyn.eval", 1, (interp, args) => {
      const source = asString(interp, args[0], "dyn.eval");
      if (typeof source !== "string") return source;
      return interp.evaluateDynamicEval("dyn.eval", source);
    });

    const symbols = new Map<string, RuntimeValue>([
      ["package", packageFn],
      ["def_package", defPackageFn],
      ["eval", evalFn],
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

  cls.prototype.evaluateDynamicEval = function evaluateDynamicEval(
    this: Interpreter,
    pkgName: string,
    source: string,
  ): RuntimeValue {
    if (!pkgName) {
      return this.makeRuntimeError("dyn.eval requires package name");
    }
    const parser = getLoadedTreeSitterParser();
    if (!parser) {
      return this.makeRuntimeError("dyn.eval requires tree-sitter to be initialized");
    }
    let moduleAst: AST.Module | null = null;
    let tree: { rootNode: any } | null = null;
    try {
      tree = parser.parse(source);
      if (tree.rootNode.type !== "source_file") {
        return buildParseErrorValue(this, `parse error: root ${tree.rootNode.type}`, 0, 0, false);
      }
      if ((tree.rootNode as unknown as { hasError?: boolean }).hasError) {
        const info = parseErrorInfoFromTree(tree.rootNode, source);
        return buildParseErrorValue(this, info.message, info.startByte, info.endByte, info.isIncomplete);
      }
      moduleAst = mapSourceFile(tree.rootNode, source, `<dyn-eval:${pkgName}>`);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      return buildParseErrorValue(this, `parse error: ${message}`, 0, 0, false);
    }
    if (!moduleAst) {
      return buildParseErrorValue(this, "parse error: unable to map source", 0, 0, false);
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
      return this.evaluate(moduleAst, this.globals);
    } catch (err) {
      if (err instanceof RaiseSignal) {
        return err.value;
      }
      const message = err instanceof Error ? err.message : String(err);
      return this.makeRuntimeError(`dyn.eval error: ${message}`);
    } finally {
      this.dynamicDefinitionMode = prevDynamicMode;
    }
  };
}
