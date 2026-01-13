import { promises as fs } from "node:fs";
import path from "node:path";
import os from "node:os";
import crypto from "node:crypto";
import { pathToFileURL } from "node:url";

import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeFloatValue, makeIntegerFromNumber, isFloatValue, isIntegerValue, integerKinds } from "./numeric";
import { RaiseSignal } from "./signals";

type HostModule = Record<string, (...args: any[]) => any>;

type ExternTargetState = {
  preludes: string[];
  externs: AST.ExternFunctionBody[];
  externIndex: Map<string, number>;
};

type ExternPackageState = {
  targets: Map<AST.HostTarget, ExternTargetState>;
  cache: Map<AST.HostTarget, { hash: string; module?: HostModule; promise?: Promise<HostModule> }>;
};

const CACHE_VERSION = "v1";
const CACHE_DIR = path.join(os.tmpdir(), "able-v11-extern-ts");
const INTEGER_KINDS = new Set(integerKinds());

const KERNEL_EXTERN_PREFIX = "__able_";

declare module "./index" {
  interface Interpreter {
    externHostPackages: Map<string, ExternPackageState>;
    registerExternStatements(module: AST.Module): void;
    invokeExternHostFunction(
      pkgName: string,
      extern: AST.ExternFunctionBody,
      args: RuntimeValue[],
    ): RuntimeValue | Promise<RuntimeValue>;
    isKernelExtern(name: string): boolean;
  }
}

function lookupStructDef(ctx: Interpreter, name: string): AST.StructDefinition | null {
  try {
    const binding = ctx.globals.get(name);
    if (binding && binding.kind === "struct_def") return binding.def;
  } catch {}
  return null;
}

function isNullable(typeExpr?: AST.TypeExpression): boolean {
  return !!typeExpr && typeExpr.type === "NullableTypeExpression";
}

function structToHostValue(ctx: Interpreter, raw: RuntimeValue, def: AST.StructDefinition): any {
  if (def.kind === "singleton" || def.fields.length === 0) {
    return def.id.name;
  }
  if (raw.kind === "struct_def") {
    return def.id.name;
  }
  if (raw.kind !== "struct_instance") {
    throw new Error(`expected ${def.id.name} struct instance`);
  }
  if (Array.isArray(raw.values)) {
    const positional = raw.values.map((val, idx) => toHostValue(ctx, val, def.fields[idx]?.fieldType));
    return { __able_struct: def.id.name, __able_positional: positional };
  }
  if (!(raw.values instanceof Map)) {
    throw new Error(`expected ${def.id.name} struct fields`);
  }
  const result: Record<string, any> = { __able_struct: def.id.name };
  for (const field of def.fields) {
    if (!field?.name) continue;
    const fieldName = field.name.name;
    const fieldValue = raw.values.get(fieldName);
    if (fieldValue === undefined) {
      if (isNullable(field.fieldType)) {
        result[fieldName] = null;
      } else {
        throw new Error(`missing ${def.id.name}.${fieldName}`);
      }
      continue;
    }
    result[fieldName] = toHostValue(ctx, fieldValue, field.fieldType);
  }
  return result;
}

function structFromHostValue(ctx: Interpreter, value: any, def: AST.StructDefinition): RuntimeValue {
  const makeNamed = (entries: Array<[string, RuntimeValue]>): RuntimeValue =>
    ctx.makeNamedStructInstance(def, entries);
  if (def.kind === "singleton" || def.fields.length === 0) {
    if (typeof value === "string") {
      if (value === def.id.name) return makeNamed([]);
      throw new Error(`extern struct tag ${value} does not match ${def.id.name}`);
    }
    if (value && typeof value === "object" && "__able_struct" in value) {
      if ((value as any).__able_struct === def.id.name) return makeNamed([]);
      throw new Error(`extern struct tag ${(value as any).__able_struct} does not match ${def.id.name}`);
    }
    throw new Error(`expected ${def.id.name} singleton`);
  }
  if (value === null || value === undefined) {
    throw new Error(`expected ${def.id.name} struct value`);
  }
  if (typeof value === "string" && value === def.id.name) {
    return makeNamed([]);
  }
  if (typeof value === "object" && value && "__able_struct" in value) {
    const tag = (value as any).__able_struct;
    if (tag && tag !== def.id.name) {
      throw new Error(`extern struct tag ${tag} does not match ${def.id.name}`);
    }
    if ((value as any).__able_positional && def.kind === "positional") {
      const items = Array.isArray((value as any).__able_positional) ? (value as any).__able_positional : [];
      const converted = items.map((entry, idx) => fromHostValue(ctx, entry, def.fields[idx]?.fieldType));
      return { kind: "struct_instance", def, values: converted };
    }
  }
  if (def.kind === "positional" && Array.isArray(value)) {
    const converted = value.map((entry, idx) => fromHostValue(ctx, entry, def.fields[idx]?.fieldType));
    return { kind: "struct_instance", def, values: converted };
  }
  const entries: Array<[string, RuntimeValue]> = [];
  const record = typeof value === "object" && value ? (value as Record<string, any>) : {};
  for (const field of def.fields) {
    if (!field?.name) continue;
    const fieldName = field.name.name;
    if (!(fieldName in record)) {
      if (isNullable(field.fieldType)) {
        entries.push([fieldName, { kind: "nil", value: null }]);
        continue;
      }
      throw new Error(`missing ${def.id.name}.${fieldName}`);
    }
    entries.push([fieldName, fromHostValue(ctx, record[fieldName], field.fieldType)]);
  }
  return makeNamed(entries);
}

function ensureTargetState(state: ExternPackageState, target: AST.HostTarget): ExternTargetState {
  let entry = state.targets.get(target);
  if (!entry) {
    entry = { preludes: [], externs: [], externIndex: new Map() };
    state.targets.set(target, entry);
  }
  return entry;
}

function sanitizePackageName(pkg: string): string {
  return pkg.replace(/[^a-zA-Z0-9._-]/g, "_") || "pkg";
}

function signatureKey(extern: AST.ExternFunctionBody): string {
  const sig = extern.signature;
  if (!sig) return "<missing>";
  const params = sig.params?.map((param) => typeKey(param.paramType)).join(",");
  const returnType = typeKey(sig.returnType);
  return `${sig.id?.name ?? "<anon>"}(${params})->${returnType}`;
}

function typeKey(typeExpr?: AST.TypeExpression): string {
  if (!typeExpr) return "void";
  switch (typeExpr.type) {
    case "SimpleTypeExpression":
      return typeExpr.name.name;
    case "GenericTypeExpression":
      return `${typeKey(typeExpr.base)}<${(typeExpr.arguments ?? []).map(typeKey).join(",")}>`;
    case "NullableTypeExpression":
      return `?${typeKey(typeExpr.innerType)}`;
    case "ResultTypeExpression":
      return `!${typeKey(typeExpr.innerType)}`;
    case "UnionTypeExpression":
      return (typeExpr.members ?? []).map(typeKey).join("|");
    case "FunctionTypeExpression":
      return `(${(typeExpr.paramTypes ?? []).map(typeKey).join(",")})->${typeKey(typeExpr.returnType)}`;
    case "WildcardTypeExpression":
      return "_";
    default:
      return typeExpr.type;
  }
}

function hashHostModule(target: AST.HostTarget, state: ExternTargetState): string {
  const hasher = crypto.createHash("sha256");
  hasher.update(`target:${target}\nversion:${CACHE_VERSION}\n`);
  for (const prelude of state.preludes) {
    hasher.update(`prelude:${prelude}\n`);
  }
  for (const extern of state.externs) {
    hasher.update(`extern:${signatureKey(extern)}\n`);
    hasher.update(`${extern.body}\n`);
  }
  return hasher.digest("hex");
}

function externParamName(param: AST.FunctionParameter, index: number): string {
  if (param?.name?.type === "Identifier") return param.name.name;
  return `arg${index}`;
}

function renderExternFunction(extern: AST.ExternFunctionBody): string {
  const sig = extern.signature;
  const name = sig?.id?.name ?? "extern_fn";
  const params = (sig?.params ?? []).map((param, idx) => `${externParamName(param, idx)}: any`).join(", ");
  const body = extern.body.trim();
  return `export function ${name}(${params}): any {\n${indent(body)}\n}`;
}

function renderHostModule(state: ExternTargetState): string {
  const sections: string[] = [];
  if (state.preludes.length > 0) {
    sections.push(state.preludes.join("\n"));
  }
  sections.push(
    [
      "type IoHandle = any;",
      "type ProcHandle = any;",
      "class AbleHostError extends Error {",
      "  __able_host_error = true;",
      "  constructor(message: string) { super(message); }",
      "}",
      "export function host_error(message: string): never {",
      "  throw new AbleHostError(message);",
      "}",
    ].join("\n"),
  );
  for (const extern of state.externs) {
    sections.push(renderExternFunction(extern));
  }
  return sections.filter(Boolean).join("\n\n");
}

function indent(body: string): string {
  if (!body) return "  ";
  return body.split("\n").map((line) => `  ${line}`).join("\n");
}

function isThenable(value: unknown): value is Promise<any> {
  return !!value && typeof (value as any).then === "function";
}

function coerceHostError(ctx: Interpreter, err: unknown): RaiseSignal {
  if (err instanceof RaiseSignal) return err;
  if (err && typeof err === "object" && (err as any).__able_host_error) {
    return new RaiseSignal(ctx.makeRuntimeError((err as Error).message));
  }
  if (err instanceof Error) {
    return new RaiseSignal(ctx.makeRuntimeError(err.message));
  }
  return new RaiseSignal(ctx.makeRuntimeError(String(err)));
}

function coerceStringValue(ctx: Interpreter, value: RuntimeValue): string {
  if (value.kind === "String") return value.value;
  if (value.kind === "struct_instance" && value.def.id.name === "String") {
    const toBuiltin = ctx.globals.get("__able_String_to_builtin");
    if (toBuiltin && toBuiltin.kind === "native_function") {
      const result = toBuiltin.impl(ctx, [value]);
      if (result && (result as RuntimeValue).kind === "String") {
        return (result as Extract<RuntimeValue, { kind: "String" }>).value;
      }
    }
  }
  throw new Error("expected String value");
}

function coerceArrayValue(ctx: Interpreter, value: RuntimeValue): Extract<RuntimeValue, { kind: "array" }> {
  if (value.kind === "array") return value;
  if (value.kind === "struct_instance" && value.def.id.name === "Array") {
    let handleVal: RuntimeValue | undefined;
    if (value.values instanceof Map) {
      handleVal = value.values.get("storage_handle");
    } else if (Array.isArray(value.values)) {
      handleVal = value.values[2];
    }
    if (handleVal && isIntegerValue(handleVal)) {
      const handle = Number(handleVal.value);
      const state = ctx.arrayStates.get(handle);
      if (state) {
        return ctx.makeArrayValue(state.values, state.capacity);
      }
    }
  }
  throw new Error("expected Array value");
}

function toHostValue(ctx: Interpreter, value: RuntimeValue, typeExpr?: AST.TypeExpression): any {
  const raw = value.kind === "interface_value" ? value.value : value;
  if (!typeExpr) return raw;
  const target = ctx.expandTypeAliases(typeExpr);
  switch (target.type) {
    case "SimpleTypeExpression": {
      const name = target.name.name;
      if (name === "String") return coerceStringValue(ctx, raw);
      if (name === "bool") return raw.kind === "bool" ? raw.value : Boolean(raw);
      if (name === "char") return raw.kind === "char" ? raw.value : String(raw);
      if (name === "nil") return null;
      if (name === "void") return undefined;
      if (name === "IoHandle" || name === "ProcHandle") {
        if (raw.kind === "host_handle" && raw.handleType === name) return raw.value;
        throw new Error(`expected ${name}`);
      }
      if (isIntegerValue(raw)) return Number(raw.value);
      if (isFloatValue(raw)) return raw.value;
      const def = lookupStructDef(ctx, name);
      if (def) return structToHostValue(ctx, raw, def);
      return raw;
    }
    case "GenericTypeExpression": {
      if (target.base.type === "SimpleTypeExpression") {
        const baseName = target.base.name.name;
        if (baseName === "Array" || baseName === "KernelArray") {
          const elemType = target.arguments?.[0];
          const arr = coerceArrayValue(ctx, raw);
          return arr.elements.map((el) => toHostValue(ctx, el, elemType));
        }
      }
      return raw;
    }
    case "NullableTypeExpression":
      if (raw.kind === "nil") return null;
      return toHostValue(ctx, raw, target.innerType);
    case "ResultTypeExpression":
      return toHostValue(ctx, raw, target.innerType);
    case "UnionTypeExpression":
      for (const member of target.members ?? []) {
        if (!member) continue;
        if (ctx.matchesType(member, raw)) {
          return toHostValue(ctx, raw, member);
        }
      }
      return raw;
    default:
      return raw;
  }
}

function fromHostValue(ctx: Interpreter, value: any, typeExpr?: AST.TypeExpression): RuntimeValue {
  if (!typeExpr) return { kind: "nil", value: null };
  const target = ctx.expandTypeAliases(typeExpr);
  switch (target.type) {
    case "SimpleTypeExpression": {
      const name = target.name.name;
      if (name === "String") return { kind: "String", value: String(value) };
      if (name === "bool") return { kind: "bool", value: Boolean(value) };
      if (name === "char") return { kind: "char", value: String(value) };
      if (name === "nil") return { kind: "nil", value: null };
      if (name === "void") return { kind: "void" };
      if (name === "IoHandle" || name === "ProcHandle") {
        return { kind: "host_handle", handleType: name, value };
      }
      if (name === "f32" || name === "f64") {
        return makeFloatValue(name as "f32" | "f64", Number(value));
      }
      if (INTEGER_KINDS.has(name as any)) {
        return makeIntegerFromNumber(name as any, Number(value));
      }
      const def = lookupStructDef(ctx, name);
      if (def) {
        return structFromHostValue(ctx, value, def);
      }
      throw new Error(`unsupported extern return type ${name}`);
    }
    case "GenericTypeExpression": {
      if (target.base.type === "SimpleTypeExpression") {
        const baseName = target.base.name.name;
        if (baseName === "Array" || baseName === "KernelArray") {
          const elemType = target.arguments?.[0];
          const raw = Array.isArray(value) ? value : [];
          const elements = raw.map((entry) => fromHostValue(ctx, entry, elemType));
          return ctx.makeArrayValue(elements);
        }
      }
      throw new Error(`unsupported extern return type ${target.base.type}`);
    }
    case "NullableTypeExpression":
      if (value === null || value === undefined) return { kind: "nil", value: null };
      return fromHostValue(ctx, value, target.innerType);
    case "ResultTypeExpression":
      return fromHostValue(ctx, value, target.innerType);
    case "UnionTypeExpression": {
      const tagged = value && typeof value === "object" && (value as any).__able_struct;
      if (tagged) {
        for (const member of target.members ?? []) {
          if (!member || member.type !== "SimpleTypeExpression") continue;
          if (member.name.name === (value as any).__able_struct) {
            return fromHostValue(ctx, value, member);
          }
        }
      }
      let lastErr: unknown = null;
      for (const member of target.members ?? []) {
        if (!member) continue;
        try {
          return fromHostValue(ctx, value, member);
        } catch (err) {
          lastErr = err;
        }
      }
      if (lastErr) throw lastErr;
      return { kind: "nil", value: null };
    }
    default:
      return { kind: "nil", value: null };
  }
}

export function applyExternHostAugmentations(cls: typeof Interpreter): void {
  cls.prototype.isKernelExtern = function isKernelExtern(this: Interpreter, name: string): boolean {
    return name.startsWith(KERNEL_EXTERN_PREFIX);
  };

  cls.prototype.registerExternStatements = function registerExternStatements(this: Interpreter, module: AST.Module): void {
    const pkgName = this.currentPackage ?? "<root>";
    if (!this.externHostPackages) {
      this.externHostPackages = new Map();
    }
    let state = this.externHostPackages.get(pkgName);
    if (!state) {
      state = { targets: new Map(), cache: new Map() };
      this.externHostPackages.set(pkgName, state);
    }
    for (const stmt of module.body ?? []) {
      if (stmt.type === "PreludeStatement") {
        const targetState = ensureTargetState(state, stmt.target);
        targetState.preludes.push(stmt.code);
      } else if (stmt.type === "ExternFunctionBody") {
        const name = stmt.signature?.id?.name;
        if (!name) continue;
        const targetState = ensureTargetState(state, stmt.target);
        const existingIndex = targetState.externIndex.get(name);
        if (existingIndex === undefined) {
          targetState.externIndex.set(name, targetState.externs.length);
          targetState.externs.push(stmt);
        } else {
          targetState.externs[existingIndex] = stmt;
        }
      }
    }
  };

  cls.prototype.invokeExternHostFunction = function invokeExternHostFunction(
    this: Interpreter,
    pkgName: string,
    extern: AST.ExternFunctionBody,
    args: RuntimeValue[],
  ): RuntimeValue | Promise<RuntimeValue> {
    const name = extern.signature?.id?.name;
    if (!name) {
      return { kind: "nil", value: null };
    }
    const pkgState = this.externHostPackages?.get(pkgName);
    if (!pkgState) {
      throw new RaiseSignal(this.makeRuntimeError(`extern package '${pkgName}' is not registered`));
    }
    const targetState = pkgState.targets.get(extern.target);
    if (!targetState) {
      throw new RaiseSignal(this.makeRuntimeError(`extern target '${extern.target}' is not available`));
    }
    const hash = hashHostModule(extern.target, targetState);
    const cached = pkgState.cache.get(extern.target);
    const moduleReady = cached && cached.hash === hash && cached.module;
    const loadPromise =
      moduleReady ? Promise.resolve(cached.module) : cached?.promise ?? loadHostModule(pkgName, extern.target, targetState, pkgState);

    const invokeWithModule = (hostModule: HostModule): RuntimeValue | Promise<RuntimeValue> => {
      const fn = hostModule[name];
      if (typeof fn !== "function") {
        throw new RaiseSignal(this.makeRuntimeError(`extern function ${name} is not exported`));
      }
      const params = extern.signature?.params ?? [];
      const hostArgs = params.map((param, idx) => toHostValue(this, args[idx] ?? { kind: "nil", value: null }, param.paramType));
      try {
        const result = fn(...hostArgs);
        if (isThenable(result)) {
          return result.then(
            (value) => fromHostValue(this, value, extern.signature?.returnType),
            (err) => {
              throw coerceHostError(this, err);
            },
          );
        }
        return fromHostValue(this, result, extern.signature?.returnType);
      } catch (err) {
        throw coerceHostError(this, err);
      }
    };

    if (moduleReady) {
      return invokeWithModule(cached!.module!);
    }
    return loadPromise.then((module) => invokeWithModule(module));
  };
}

async function loadHostModule(
  pkgName: string,
  target: AST.HostTarget,
  targetState: ExternTargetState,
  pkgState: ExternPackageState,
): Promise<HostModule> {
  const hash = hashHostModule(target, targetState);
  const existing = pkgState.cache.get(target);
  if (existing && existing.hash === hash && existing.module) {
    return existing.module;
  }
  const cacheEntry = existing && existing.hash === hash ? existing : { hash };
  const promise = (async () => {
    await fs.mkdir(CACHE_DIR, { recursive: true });
    const fileName = `${sanitizePackageName(pkgName)}-${hash}.ts`;
    const filePath = path.join(CACHE_DIR, fileName);
    try {
      await fs.access(filePath);
    } catch {
      const source = renderHostModule(targetState);
      await fs.writeFile(filePath, source, "utf8");
    }
    const moduleUrl = pathToFileURL(filePath).href;
    const loaded = await import(moduleUrl);
    cacheEntry.module = loaded as HostModule;
    cacheEntry.promise = undefined;
    pkgState.cache.set(target, cacheEntry as { hash: string; module?: HostModule; promise?: Promise<HostModule> });
    return loaded as HostModule;
  })();
  cacheEntry.promise = promise;
  pkgState.cache.set(target, cacheEntry as { hash: string; module?: HostModule; promise?: Promise<HostModule> });
  return promise;
}
