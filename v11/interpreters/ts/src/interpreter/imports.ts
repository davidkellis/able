import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

const NIL: RuntimeValue = { kind: "nil", value: null };

export function evaluateModule(ctx: Interpreter, node: AST.Module): RuntimeValue {
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
  let last: RuntimeValue = NIL;
  for (const stmt of node.body) {
    last = ctx.evaluate(stmt, moduleEnv);
  }
  ctx.currentPackage = prevPkg;
  return last;
}

export function evaluatePackageStatement(): RuntimeValue {
  return NIL;
}

export function evaluateImportStatement(ctx: Interpreter, node: AST.ImportStatement, env: Environment): RuntimeValue {
  if (!node.isWildcard && (!node.selectors || node.selectors.length === 0)) {
    const pkg = node.packagePath.map(p => p.name).join(".");
    const bucket = ctx.packageRegistry.get(pkg);
    if (!bucket) throw new Error(`Import error: package '${pkg}' not found`);
    const filtered = new Map<string, RuntimeValue>();
    for (const [name, val] of bucket.entries()) {
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
      if (val.kind === "struct_def" && val.def.isPrivate) continue;
      if (val.kind === "interface_def" && val.def.isPrivate) continue;
      if (val.kind === "union_def" && val.def.isPrivate) continue;
      filtered.set(name, val);
    }
    const alias = node.alias?.name ?? defaultPackageAlias(pkg);
    env.define(alias, { kind: "package", name: pkg, symbols: filtered });
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
      let val: RuntimeValue | null = null;
      if (pkg) {
        const bucket = ctx.packageRegistry.get(pkg);
        if (bucket?.has(original)) {
          val = bucket.get(original) ?? null;
        }
      }
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
      if (!val) {
        const aliasDef = ctx.typeAliases.get(original);
        if (aliasDef) {
          if (aliasDef.isPrivate) {
            throw new Error(`Import error: type alias '${original}' is private`);
          }
          if (alias !== original) {
            ctx.typeAliases.set(alias, { ...aliasDef, id: AST.identifier(alias) });
          }
          continue;
        }
        throw new Error(`Import error: symbol '${original}'${pkg ? ` from '${pkg}'` : ""} not found in globals`);
      }
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
      if (env.hasInCurrentScope(alias)) {
        continue;
      }
      env.define(alias, val);
    }
  }
  return NIL;
}

export function evaluateDynImportStatement(ctx: Interpreter, node: AST.DynImportStatement, env: Environment): RuntimeValue {
  const pkg = node.packagePath.map(p => p.name).join(".");
  if (node.isWildcard) {
    const bucket = ctx.packageRegistry.get(pkg);
    if (!bucket) throw new Error(`dynimport error: package '${pkg}' not found`);
    for (const [name, val] of bucket.entries()) {
      if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
      if (val.kind === "struct_def" && val.def.isPrivate) continue;
      if (val.kind === "interface_def" && val.def.isPrivate) continue;
      if (val.kind === "union_def" && val.def.isPrivate) continue;
      try { env.define(name, { kind: "dyn_ref", pkg, name }); } catch {}
    }
  } else if (node.selectors && node.selectors.length > 0) {
    const bucket = ctx.packageRegistry.get(pkg);
    for (const sel of node.selectors) {
      const original = sel.name.name;
      const alias = sel.alias ? sel.alias.name : original;
      const val = bucket?.get(original);
      if (val?.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) throw new Error(`dynimport error: function '${original}' is private`);
      if (val?.kind === "struct_def" && val.def.isPrivate) throw new Error(`dynimport error: struct '${original}' is private`);
      if (val?.kind === "interface_def" && val.def.isPrivate) throw new Error(`dynimport error: interface '${original}' is private`);
      if (val?.kind === "union_def" && val.def.isPrivate) throw new Error(`dynimport error: union '${original}' is private`);
      if (env.hasInCurrentScope(alias)) {
        continue;
      }
      env.define(alias, { kind: "dyn_ref", pkg, name: original });
    }
  } else {
    const alias = node.alias?.name ?? defaultPackageAlias(pkg);
    env.define(alias, { kind: "dyn_package", name: pkg });
  }
  return NIL;
}

function defaultPackageAlias(pkg: string): string {
  if (!pkg) return pkg;
  const parts = pkg.split(".");
  return parts[parts.length - 1] ?? pkg;
}
