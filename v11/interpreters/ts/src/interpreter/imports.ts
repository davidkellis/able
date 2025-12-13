import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

const NIL: V10Value = { kind: "nil", value: null };

export function evaluateModule(ctx: InterpreterV10, node: AST.Module): V10Value {
  const moduleEnv = node.package ? new Environment(ctx.globals) : ctx.globals;
  const prevPkg = ctx.currentPackage;
  if (node.package) {
    ctx.currentPackage = node.package.namePath.map(p => p.name).join(".");
    if (!ctx.packageRegistry.has(ctx.currentPackage)) ctx.packageRegistry.set(ctx.currentPackage, new Map());
  } else {
    ctx.currentPackage = null;
  }
  for (const imp of node.imports) {
    ctx.evaluate(imp, moduleEnv);
  }
  let last: V10Value = NIL;
  for (const stmt of node.body) {
    last = ctx.evaluate(stmt, moduleEnv);
  }
  ctx.currentPackage = prevPkg;
  return last;
}

export function evaluatePackageStatement(): V10Value {
  return NIL;
}

export function evaluateImportStatement(ctx: InterpreterV10, node: AST.ImportStatement, env: Environment): V10Value {
  if (!node.isWildcard && (!node.selectors || node.selectors.length === 0) && node.alias) {
    const pkg = node.packagePath.map(p => p.name).join(".");
    const bucket = ctx.packageRegistry.get(pkg);
    if (!bucket) throw new Error(`Import error: package '${pkg}' not found`);
    const filtered = new Map<string, V10Value>();
    for (const [name, val] of bucket.entries()) {
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
      if (val.kind === "struct_def" && val.def.isPrivate) continue;
      if (val.kind === "interface_def" && val.def.isPrivate) continue;
      if (val.kind === "union_def" && val.def.isPrivate) continue;
      filtered.set(name, val);
    }
    env.define(node.alias.name, { kind: "package", name: pkg, symbols: filtered });
  } else if (node.isWildcard) {
    const pkg = node.packagePath.map(p => p.name).join(".");
    const bucket = ctx.packageRegistry.get(pkg);
    if (!bucket) throw new Error(`Import error: package '${pkg}' not found`);
    for (const [name, val] of bucket.entries()) {
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
      if (val.kind === "struct_def" && val.def.isPrivate) continue;
      if (val.kind === "interface_def" && val.def.isPrivate) continue;
      if (val.kind === "union_def" && val.def.isPrivate) continue;
      try { env.define(name, val); } catch {}
    }
  } else if (node.selectors && node.selectors.length > 0) {
    const pkg = node.packagePath.map(p => p.name).join(".");
    for (const sel of node.selectors) {
      const original = sel.name.name;
      const alias = sel.alias ? sel.alias.name : original;
      let val: V10Value | null = null;
      const fq = pkg ? `${pkg}.${original}` : original;
      const reexports: Record<string, string> = {
        "able.collections.array.Array": "able.kernel.Array",
        "able.collections.range.Range": "able.kernel.Range",
        "able.collections.range.RangeFactory": "able.kernel.RangeFactory",
        "able.core.numeric.Ratio": "able.kernel.Ratio",
        "able.concurrency.Channel": "able.kernel.Channel",
        "able.concurrency.Mutex": "able.kernel.Mutex",
        "able.concurrency.Awaitable": "able.kernel.Awaitable",
        "able.concurrency.AwaitWaker": "able.kernel.AwaitWaker",
        "able.concurrency.AwaitRegistration": "able.kernel.AwaitRegistration",
      };
      if (pkg) {
        try { val = ctx.globals.get(`${pkg}.${original}`); } catch {}
      }
      if (!val) {
        try { val = ctx.globals.get(original); } catch {}
      }
      if (!val && pkg) {
        try { val = ctx.globals.get(`${pkg}.${original}`); } catch {}
      }
      if (!val) {
        const fallback = reexports[fq];
        if (fallback) {
          try { val = ctx.globals.get(fallback); } catch {}
        }
      }
      if (!val) throw new Error(`Import error: symbol '${original}'${pkg ? ` from '${pkg}'` : ""} not found in globals`);
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) {
        throw new Error(`Import error: function '${original}' is private`);
      }
      if (val.kind === "struct_def" && val.def.isPrivate) {
        throw new Error(`Import error: struct '${original}' is private`);
      }
      if (val.kind === "interface_def" && val.def.isPrivate) {
        throw new Error(`Import error: interface '${original}' is private`);
      }
      if (val.kind === "union_def" && val.def.isPrivate) {
        throw new Error(`Import error: union '${original}' is private`);
      }
      env.define(alias, val);
    }
  }
  return NIL;
}

export function evaluateDynImportStatement(ctx: InterpreterV10, node: AST.DynImportStatement, env: Environment): V10Value {
  const pkg = node.packagePath.map(p => p.name).join(".");
  const bucket = ctx.packageRegistry.get(pkg);
  if (!bucket) throw new Error(`dynimport error: package '${pkg}' not found`);
  if (node.isWildcard) {
    for (const [name, val] of bucket.entries()) {
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
      if (val.kind === "struct_def" && val.def.isPrivate) continue;
      if (val.kind === "interface_def" && val.def.isPrivate) continue;
      if (val.kind === "union_def" && val.def.isPrivate) continue;
      try { env.define(name, { kind: "dyn_ref", pkg, name }); } catch {}
    }
  } else if (node.selectors && node.selectors.length > 0) {
    for (const sel of node.selectors) {
      const original = sel.name.name;
      const alias = sel.alias ? sel.alias.name : original;
      const val = bucket.get(original);
      if (!val) throw new Error(`dynimport error: '${original}' not found in '${pkg}'`);
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) throw new Error(`dynimport error: function '${original}' is private`);
      if (val.kind === "struct_def" && val.def.isPrivate) throw new Error(`dynimport error: struct '${original}' is private`);
      if (val.kind === "interface_def" && val.def.isPrivate) throw new Error(`dynimport error: interface '${original}' is private`);
      if (val.kind === "union_def" && val.def.isPrivate) throw new Error(`dynimport error: union '${original}' is private`);
      env.define(alias, { kind: "dyn_ref", pkg, name: original });
    }
  } else if (node.alias) {
    env.define(node.alias.name, { kind: "dyn_package", name: pkg });
  }
  return NIL;
}
