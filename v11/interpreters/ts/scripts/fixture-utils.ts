import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, V10 } from "../index";
import { mapSourceFile } from "../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../src/parser/tree-sitter-loader";
import { makeIntegerValue, numericToNumber } from "../src/interpreter/numeric";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT_ORIGIN = path.join("..", "..", "..", "..", "fixtures", "ast");
const FIXTURE_ROOT = path.resolve(__dirname, "../../../fixtures/ast");

export type Manifest = {
  description?: string;
  entry?: string;
  setup?: string[];
  skipTargets?: string[];
  expect?: {
    result?: { kind: string; value?: unknown };
    stdout?: string[];
    stderr?: string[];
    exit?: number;
    errors?: string[];
    typecheckDiagnostics?: string[];
  };
};

export async function collectFixtures(root: string): Promise<string[]> {
  const dirs: string[] = [];
  async function walk(current: string) {
    const entries = await fs.readdir(current, { withFileTypes: true });
    let hasModule = false;
    for (const entry of entries) {
      if (entry.isFile() && entry.name === "module.json") {
        hasModule = true;
      }
    }
    if (hasModule) {
      dirs.push(current);
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(current, entry.name));
      }
    }
  }
  await walk(root);
  return dirs.sort();
}

export async function readManifest(dir: string): Promise<Manifest> {
  const manifestPath = path.join(dir, "manifest.json");
  try {
    const contents = await fs.readFile(manifestPath, "utf8");
    return JSON.parse(contents) as Manifest;
  } catch (err: any) {
    if (err && err.code === "ENOENT") return {};
    throw err;
  }
}

export async function loadModuleFromFixture(dir: string, relativePath: string): Promise<AST.Module> {
  const absolute = path.join(dir, relativePath);
  return loadModuleFromPath(absolute);
}

export async function loadModuleFromPath(filePath: string): Promise<AST.Module> {
  if (filePath.endsWith(".json")) {
    try {
      const raw = JSON.parse(await fs.readFile(filePath, "utf8"));
      const module = hydrateNode(raw) as AST.Module;
      const sibling = await findSourceSibling(filePath);
      if (sibling) {
        const spanSource = await parseModuleFromSource(sibling);
        if (spanSource) {
          copyNodeSpans(module, spanSource);
        }
      }
      const originPath = normalizeFixtureOrigin(sibling ?? filePath);
      annotateModuleOrigin(module, originPath);
      return module;
    } catch (err: any) {
      if (err && err.code !== "ENOENT") {
        throw err;
      }
      // fall through to attempt source-based parse
    }
    const directory = path.dirname(filePath);
    const base = path.basename(filePath, ".json");
    const candidates = [path.join(directory, `${base}.able`)];
    if (base === "module") {
      candidates.push(path.join(directory, "source.able"));
    }
    for (const candidate of candidates) {
      const fromSource = await parseModuleFromSource(candidate);
      if (fromSource) {
        annotateModuleOrigin(fromSource, normalizeFixtureOrigin(candidate));
        return fromSource;
      }
    }
    throw new Error(`unable to load module from ${filePath}`);
  }
  const module = await parseModuleFromSource(filePath);
  if (!module) {
    throw new Error(`unable to parse source module ${filePath}`);
  }
  annotateModuleOrigin(module, normalizeFixtureOrigin(filePath));
  return module;
}

async function findSourceSibling(jsonPath: string): Promise<string | null> {
  const dir = path.dirname(jsonPath);
  const base = path.basename(jsonPath, ".json");
  const candidates = [path.join(dir, `${base}.able`)];
  if (base === "module") {
    candidates.push(path.join(dir, "source.able"));
  }
  for (const candidate of candidates) {
    try {
      await fs.access(candidate);
      return candidate;
    } catch {
      // continue
    }
  }
  return null;
}

export async function parseModuleFromSource(filePath: string): Promise<AST.Module | null> {
  try {
    const source = await fs.readFile(filePath, "utf8");
    const parser = await getTreeSitterParser();
    const tree = parser.parse(source);
    const mapped = mapSourceFile(tree.rootNode, source, filePath);
    const module = hydrateNode(mapped) as AST.Module;
    annotateModuleOrigin(module, filePath);
    return module;
  } catch (err: any) {
    if (err && err.code === "ENOENT") return null;
    throw err;
  }
}

export function hydrateNode(node: any): AST.Node {
  if (node === null || typeof node !== "object") return node;
  if (Array.isArray(node)) {
    return node.map((entry) => hydrateNode(entry)) as unknown as AST.Node;
  }
  const { type, ...rest } = node as { type?: string };
  if (!type) {
    const hydrated: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(rest)) {
      hydrated[key] = hydrateNode(value);
    }
    return hydrated as unknown as AST.Node;
  }
  const hydrated: Record<string, unknown> = { type };
  for (const [key, value] of Object.entries(rest)) {
    hydrated[key] = hydrateNode(value);
  }
  return hydrated as unknown as AST.Node;
}

export function annotateModuleOrigin(module: AST.Module, filePath: string): void {
  if (!module) return;
  const queue: AST.Node[] = [module];
  while (queue.length > 0) {
    const current = queue.pop()!;
    if (typeof current !== "object" || current === null) continue;
    (current as any).origin = filePath;
    for (const value of Object.values(current)) {
      if (typeof value === "object" && value) {
        if (Array.isArray(value)) {
          for (const entry of value) {
            if (entry && typeof entry === "object") queue.push(entry as AST.Node);
          }
        } else {
          queue.push(value as AST.Node);
        }
      }
    }
  }
}

function normalizeFixtureOrigin(filePath: string): string {
  const absolute = path.resolve(filePath);
  const relative = path.relative(FIXTURE_ROOT, absolute);
  if (relative && !relative.startsWith("..") && !path.isAbsolute(relative)) {
    const combined = path.join(FIXTURE_ROOT_ORIGIN, relative);
    return combined.split(path.sep).join("/");
  }
  return absolute.split(path.sep).join("/");
}

function copyNodeSpans(target: AST.Node | undefined, source: AST.Node | undefined): void {
  if (!target || !source) return;
  if (typeof target !== "object" || typeof source !== "object") return;
  if (Array.isArray(target) && Array.isArray(source)) {
    const length = Math.min(target.length, source.length);
    for (let index = 0; index < length; index += 1) {
      copyNodeSpans(target[index] as AST.Node, source[index] as AST.Node);
    }
    return;
  }
  if (Array.isArray(target) || Array.isArray(source)) {
    return;
  }
  const targetRecord = target as Record<string, unknown>;
  const sourceRecord = source as Record<string, unknown>;
  const targetType = typeof targetRecord.type === "string" ? targetRecord.type : null;
  const sourceType = typeof sourceRecord.type === "string" ? sourceRecord.type : null;
  if (targetType && sourceType && targetType !== sourceType) {
    return;
  }
  if (sourceRecord.span) {
    (targetRecord as { span?: AST.Span }).span = sourceRecord.span as AST.Span;
  }
  const keys = new Set([...Object.keys(targetRecord), ...Object.keys(sourceRecord)]);
  for (const key of keys) {
    copyNodeSpans(targetRecord[key] as AST.Node, sourceRecord[key] as AST.Node);
  }
}

export function ensurePrint(interpreter: V10.InterpreterV10): void {
  try {
    const existing = interpreter.globals.get("print");
    if (existing && typeof existing === "object") return;
  } catch {
    // fall through and define print
  }
  interpreter.globals.define(
    "print",
    interpreter.makeNativeFunction("print", 1, (_interp, args) => {
      const parts = args.map((value) => formatValue(value));
      console.log(parts.join(" "));
      return { kind: "nil", value: null };
    }),
  );
}

export function installRuntimeStubs(interpreter: V10.InterpreterV10): void {
  const globals = interpreter.globals;

  const defineStub = (
    name: string,
    arity: number,
    impl: (interp: V10.InterpreterV10, args: V10.V10Value[]) => V10.V10Value | null,
  ) => {
    try {
      globals.define(
        name,
        interpreter.makeNativeFunction(name, arity, (innerInterp, args) => {
          const result = impl(innerInterp as V10.InterpreterV10, args);
          return result ?? { kind: "nil", value: null };
        }),
      );
    } catch {
      // ignore redefinition attempts
    }
  };

  const hasGlobal = (name: string): boolean => {
    try {
      globals.get(name);
      return true;
    } catch {
      return false;
    }
  };

  let nextHandle = 1;
  const makeHandle = (): V10.V10Value => makeIntegerValue("i32", BigInt(nextHandle++));

  type ChannelState = {
    capacity: number;
    queue: V10.V10Value[];
    closed: boolean;
  };

  const channels = new Map<number, ChannelState>();
  type MutexState = {
    locked: boolean;
  };
  const mutexes = new Map<number, MutexState>();

  const toNumber = (value: V10.V10Value, label = "numeric value"): number => {
    return numericToNumber(value, label);
  };
  const toHandle = (value: V10.V10Value): number => Math.trunc(toNumber(value, "handle"));

  const checkCancelled = (interp: V10.InterpreterV10): boolean => {
    try {
      const cancelled = interp.procCancelled();
      return cancelled.kind === "bool" && cancelled.value;
    } catch {
      return false;
    }
  };

  const blockOnNilChannel = (interp: V10.InterpreterV10): V10.V10Value | null => {
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return null;
  };

  const awaitableDef = AST.structDefinition("ChannelAwaitable", [], "named");
  const awaitRegistrationDef = AST.structDefinition("AwaitRegistration", [], "named");

  const makeAwaitRegistration = (interp: V10.InterpreterV10, cancel?: () => void): V10.V10Value => {
    const inst: V10.V10Value = { kind: "struct_instance", def: awaitRegistrationDef, values: new Map() };
    const cancelNative = interp.makeNativeFunction("AwaitRegistration.cancel", 1, () => {
      if (cancel) cancel();
      return { kind: "nil", value: null };
    });
    (inst.values as Map<string, V10.V10Value>).set("cancel", interp.bindNativeMethod(cancelNative, inst));
    return inst;
  };

  const makeAwaitable = (
    interp: V10.InterpreterV10,
    handle: number,
    op: "send" | "receive",
    payload: V10.V10Value | null,
    callback?: V10.V10Value,
  ): V10.V10Value => {
    const inst: V10.V10Value = { kind: "struct_instance", def: awaitableDef, values: new Map() };
    const isReady = interp.makeNativeFunction("Awaitable.is_ready", 1, () => {
      const channel = channels.get(handle);
      if (!channel) return { kind: "bool", value: false };
      if (op === "receive") {
        const ready = channel.queue.length > 0 || (channel.capacity === 0 && channel.queue.length > 0) || channel.closed;
        return { kind: "bool", value: ready };
      }
      if (channel.closed) {
        throw new Error("send on closed channel");
      }
      const ready = channel.capacity === 0 ? true : channel.queue.length < channel.capacity;
      return { kind: "bool", value: ready };
    });
    const register = interp.makeNativeFunction("Awaitable.register", 2, () => makeAwaitRegistration(interp));
    const commit = interp.makeNativeFunction("Awaitable.commit", 1, () => {
      const channel = channels.get(handle);
      if (!channel) return { kind: "nil", value: null };
      if (op === "receive") {
        const value = channel.queue.shift() ?? { kind: "nil", value: null };
        if (callback && callback.kind === "native_function") {
          callback.impl(interp, [value]);
        }
        return value;
      }
      if (channel.closed) {
        throw new Error("send on closed channel");
      }
      if (op === "send") {
        if (channel.capacity === 0 || channel.queue.length < channel.capacity) {
          channel.queue.push(payload ?? { kind: "nil", value: null });
        }
        if (callback && callback.kind === "native_function") {
          callback.impl(interp, []);
        }
      }
      return { kind: "nil", value: null };
    });
    const isDefault = interp.makeNativeFunction("Awaitable.is_default", 1, () => ({ kind: "bool", value: false }));
    const values = inst.values as Map<string, V10.V10Value>;
    values.set("is_ready", interp.bindNativeMethod(isReady, inst));
    values.set("register", interp.bindNativeMethod(register, inst));
    values.set("commit", interp.bindNativeMethod(commit, inst));
    values.set("is_default", interp.bindNativeMethod(isDefault, inst));
    return inst;
  };

  if (!hasGlobal("__able_channel_new")) defineStub("__able_channel_new", 1, (_interp, [capacityArg]) => {
    const capacity = Math.max(0, Math.trunc(toNumber(capacityArg)));
    const handleValue = makeHandle();
    const handle = toHandle(handleValue);
    channels.set(handle, { capacity, queue: [], closed: false });
    return handleValue;
  });

  if (!hasGlobal("__able_channel_send")) defineStub("__able_channel_send", 2, (interp, [handleArg, value]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) {
      const blocked = blockOnNilChannel(interp);
      if (blocked) return blocked;
      return { kind: "nil", value: null };
    }
    if (channel.closed) {
      throw new Error("send on closed channel");
    }
    if (channel.capacity === 0) {
      channel.queue = [value];
    } else if (channel.queue.length < channel.capacity) {
      channel.queue.push(value);
    } else {
      channel.queue[channel.queue.length - 1] = value;
    }
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_receive")) defineStub("__able_channel_receive", 1, (interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) {
      const blocked = blockOnNilChannel(interp);
      if (blocked) return blocked;
      return { kind: "nil", value: null };
    }
    if (channel.queue.length > 0) {
      return channel.queue.shift()!;
    }
    if (channel.closed) {
      return { kind: "nil", value: null };
    }
    const blocked = blockOnNilChannel(interp);
    if (blocked) return blocked;
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_try_send")) defineStub("__able_channel_try_send", 2, (_interp, [handleArg, value]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) return { kind: "bool", value: false };
    if (channel.closed) throw new Error("send on closed channel");
    if (channel.capacity === 0) {
      channel.queue = [value];
      return { kind: "bool", value: true };
    }
    if (channel.queue.length < channel.capacity) {
      channel.queue.push(value);
      return { kind: "bool", value: true };
    }
    return { kind: "bool", value: false };
  });

  if (!hasGlobal("__able_channel_try_receive")) defineStub("__able_channel_try_receive", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) return { kind: "nil", value: null };
    if (channel.queue.length > 0) {
      return channel.queue.shift()!;
    }
    return channel.closed ? { kind: "nil", value: null } : { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_await_try_recv")) {
    defineStub("__able_channel_await_try_recv", 2, (interp, [handleArg, callback]) => {
      const handle = toHandle(handleArg);
      return makeAwaitable(interp, handle, "receive", null, callback);
    });
  }

  if (!hasGlobal("__able_channel_await_try_send")) {
    defineStub("__able_channel_await_try_send", 3, (interp, [handleArg, value, callback]) => {
      const handle = toHandle(handleArg);
      return makeAwaitable(interp, handle, "send", value, callback);
    });
  }

  if (!hasGlobal("__able_channel_close")) defineStub("__able_channel_close", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (channel) channel.closed = true;
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_is_closed")) defineStub("__able_channel_is_closed", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    return { kind: "bool", value: channel ? channel.closed : false };
  });

  if (!hasGlobal("__able_mutex_new")) defineStub("__able_mutex_new", 0, () => makeHandle());
  if (!hasGlobal("__able_mutex_lock")) defineStub("__able_mutex_lock", 1, (interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    let state = mutexes.get(handle);
    if (!state) {
      state = { locked: false };
      mutexes.set(handle, state);
    }
    if (!state.locked) {
      state.locked = true;
      return { kind: "nil", value: null };
    }
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_mutex_unlock")) defineStub("__able_mutex_unlock", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const state = mutexes.get(handle);
    if (!state || !state.locked) {
      throw new Error("unlock of unlocked mutex");
    }
    state.locked = false;
    return { kind: "nil", value: null };
  });
}

export function interceptStdout(buffer: string[], fn: () => void): void {
  const originalLog = console.log;
  try {
    console.log = (...args: unknown[]) => {
      buffer.push(args.map((arg) => (typeof arg === "string" ? arg : JSON.stringify(arg))).join(" "));
    };
    fn();
  } finally {
    console.log = originalLog;
  }
}

export function formatValue(value: V10.V10Value): string {
  switch (value.kind) {
    case "String":
    case "char":
      return String(value.value);
    case "bool":
      return value.value ? "true" : "false";
    case "i32":
    case "f32":
    case "f64":
      return String(value.value);
    default:
      if (typeof (value as any).value === "number" || typeof (value as any).value === "bigint") {
        return String((value as any).value);
      }
      return `[${value.kind}]`;
  }
}

export function extractErrorMessage(err: unknown): string {
  if (!err) return "";
  if (typeof err === "string") return err;
  if (err instanceof Error) {
    const anyErr = err as any;
    if (anyErr.value && typeof anyErr.value === "object" && "message" in anyErr.value) {
      return String(anyErr.value.message);
    }
    return err.message;
  }
  if (typeof err === "object" && err) {
    const anyErr = err as any;
    if (typeof anyErr.message === "string") return anyErr.message;
  }
  return String(err);
}
