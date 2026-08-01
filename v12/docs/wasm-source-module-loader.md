# WASM static source/module loader — 2026-07-14

## Scope and decision

The Go/WASM prototype can now execute a small static source-module graph in
both interpreter modes. The JavaScript adapter parses an entry source file,
resolves its static import closure, and supplies dependencies in deterministic
dependency-first order through `EvaluateRequest.setupModules`. Each setup
module and the entry carry its source origin for host diagnostics.

The resolver is now portable JavaScript: it has no Node filesystem dependency
and accepts host-approved source bytes through a small provider contract. The
checked-in command-line prototype supplies a Node provider; a browser can
supply virtual, sandboxed, or user-granted sources without reimplementing the
import graph rules.

This is browser-host preparation, not a claim that the Go/WASM filesystem ABI
is wired and not a compiler/bytecode performance change. No canonical-stdlib,
language specification, benchmark, or runtime optimization changed.

## Current source layout

For an import such as `import math.{answer}`, the loader checks ordered roots
for:

1. `<root>/math.able`
2. `<root>/math/main.able`

The entry's directory is always the first root; `--module-root <path>` adds
ordered fallbacks. An imported source file must declare the exact imported
package path. The graph rejects cycles, unsafe import path segments, missing
packages, and `dynimport`, because dynamic module semantics must not silently
be treated as static browser imports.

The checked-in three-module smoke uses `dep.able`, `math.able`, and
`main.able`. Its closure is transmitted as:

```text
setupModules: dep.able, math.able
entry:        main.able
```

Both the tree-walker and bytecode interpreter resolve the imports, send `42\n`
through `able_host.write_stdout`, and return `42`.

## Source-provider contract

`loadStaticModuleClosure` owns layout selection, deterministic root precedence,
dependency-first ordering, import validation, and cycle detection. A host
supplies only source retrieval:

```js
const sourceProvider = {
  async readSource(origin) {
    // Return source text, null when this approved origin is absent,
    // or throw when host policy denies access or a real read fails.
  },
};

const closure = await loadStaticModuleClosure({
  entryPath: "/workspace/main.able",
  moduleRoots: ["/approved-stdlib"],
  parseSource,
  sourceProvider,
});
```

Origins are slash-separated virtual paths. The resolver appends only validated
package segments to the caller-supplied entry directory and roots; a host may
map those origins to real files, browser file handles, IndexedDB, or an
in-memory map. Returning `null` lets the resolver continue its ordered search;
throwing preserves a policy/read error and stops loading. `parseSource` also
receives the origin as its optional second argument for source diagnostics.

`node_source_provider.mjs` is the narrow Node-only adapter used by
`run_prototype.mjs`. It maps an absent file to `null` and propagates other
filesystem failures. It is intentionally outside the portable resolver.

## Portable request composition

`source_request.mjs` is the browser-facing layer above the resolver. It has no
Node imports and combines a caller-owned parser bridge with a source provider
into the exact object accepted by `__able_eval_request_json`:

```js
import { buildSourceEvaluationRequest } from "./source_request.mjs";

const request = await buildSourceEvaluationRequest({
  entryPath: "/workspace/main.able",
  moduleRoots: ["/approved-stdlib"],
  execMode: "bytecode",
  sourceProvider,
  parseSource(source, origin) {
    return parseSourceToAstModule(parser, source);
  },
});

const response = JSON.parse(evaluate(JSON.stringify(request)));
```

The browser owns parser creation, virtual-source permission, output callbacks,
and Go/WASM startup. This layer owns only request composition, so the Node CLI
uses it too rather than maintaining a second request shape. It preserves the
resolver's dependency-first setup order and source origins, while leaving the
host free to select either `treewalker` or `bytecode`.

## Browser execution smoke

`browser_smoke.html` and `browser_smoke.mjs` exercise that request path inside
an actual browser. They start the Go/WASM binary with `wasm_exec.js`, expose
the named `able_host` output callbacks, and fetch the checked-in
`dep -> math -> app` Able files through approved virtual origins. The portable
AST mapper is shared with the Node CLI; only parser startup differs. The
browser page runs both interpreter modes and requires result `42` plus exactly
`42\n` on host stdout.

`browser_smoke.test.mjs` is a local HTTP/WebDriver driver for the page. It
uses headless Firefox and geckodriver, serves only the explicit source files,
WASM smoke assets, Go's installed `wasm_exec.js`, and the lockfile-pinned
tree-sitter runtime/grammar. It reads the page's pass/fail state and needs no
Go filesystem callback:

```bash
cd v12/wasm && npm ci
cd ../..
just wasm-browser-smoke
```

Before loading `web-tree-sitter`, the browser adapter supplies an empty
`process.versions` object when Go's `wasm_exec.js` has installed its browser
`process` shim. This avoids that package's Node-loader probe mistaking the Go
shim for Node. The smoke therefore proves browser Go/WASM startup, actual Able
source parsing, output wiring, virtual-source policy, resolver/request
composition, and both runtime modes.

## Request and diagnostics boundary

`wasmhost.EvaluateRequest` retains the original unlabelled `setup` array for
compatibility, and adds:

```json
{
  "setupModules": [{"origin": "/workspace/dep.able", "module": {}}],
  "entryOrigin": "/workspace/main.able"
}
```

The evaluator executes those modules in supplied order. Decode/evaluation
errors name a setup origin or entry origin, while the direct pre-parsed JSON
bridge retains its prior origin-free diagnostics.

## Verification

- `node v12/wasm/module_loader.test.mjs` uses an in-memory source provider to
  test dependency order, root precedence, dynamic-import rejection, cycle
  rejection, and preservation of host policy failures without any npm
  dependency.
- `node v12/wasm/source_request.test.mjs` uses a virtual source map and parser
  callback to prove the exact bytecode request shape, dependency ordering, and
  origin-labelled parser diagnostics without Node filesystem or WASM startup.
- `just wasm-browser-smoke` builds `ablewasm.wasm` and drives Firefox through
  actual checked-in source modules in both runtime modes. It requires the
  `web-tree-sitter` package from `npm ci`, local Firefox, and geckodriver, but
  no Go filesystem ABI callback.
- `just wasm-source-module-smoke` runs the actual web-tree-sitter adapter plus
  the dependency closure in tree-walker and bytecode mode. It requires the
  lockfile-pinned `web-tree-sitter` package (`cd v12/wasm && npm ci`).
- Native `go test ./pkg/wasmhost` proves the ordered setup request and
  origin-labelled errors in both interpreter modes.

## Deliberately deferred

- broader JavaScript AST mapping beyond the current small expression/import
  subset;
- `dynimport`, runtime source evaluation, and browser externs;
- Go-native parsing/typechecking inside the WASM module; and
- `able_host` filesystem, module-root, timer, and scheduler callbacks.

## Next recommendation

Keep the completed WASM boundary narrow and return to the verifier-backed
performance gate. Browser execution now parses the existing source fixture in
both runtime modes; broader browser grammar or host ABI work has no benchmark
evidence to justify it. Refresh bounded profiles only for three unlike material
performance misses, then pursue a candidate only if one concrete leaf repeats
across them without violating existing regression guards.
