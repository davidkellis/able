# Able WASM Prototype

This directory contains a minimal end-to-end prototype for:

1. Parsing Able source in JavaScript with tree-sitter (`web-tree-sitter` + `tree-sitter-able.wasm`).
2. Converting the parse tree into fixture-style AST JSON for a narrow syntax subset.
3. Sending that JSON into the Go/WASM runtime via `__able_eval_request_json`.

The runtime accepts pre-parsed AST modules and therefore does not bundle the
native source loader, typechecker, or Go-plugin extern host. Dynamic source
evaluation and Go extern functions report an explicit unsupported error on
`js/wasm`. The initial browser host bridge forwards the evaluator's injected
`print` helper and evaluation failures through `globalThis.able_host` methods
`write_stdout(string)` and `write_stderr(string)`; filesystem, module, timer,
and browser-extern callbacks remain future ABI work.

The adapter currently supports a deliberately small subset:

- package statements (`package name`)
- static/dynamic import statements
- identifiers, numbers, booleans, strings, nil
- assignment expressions
- binary expressions
- postfix member access + function calls

If the source uses unsupported constructs, the adapter exits with an explicit error.

## Static source modules

When using `--source`, the prototype resolves static imports before invoking
WASM. It loads dependencies in dependency-first order and sends them as
origin-labelled `setupModules` before the entry AST. The current intentionally
narrow source layout is one package per `<root>/<package>.able` or
`<root>/<package>/main.able`; the directory containing the entry is the first
root and `--module-root` adds ordered fallbacks. A dependency's declared
`package` must exactly match the imported package path. `dynimport` is rejected
at this browser source-loading boundary rather than silently acquiring static
semantics.

The resolver itself is portable and has no Node filesystem import. It takes a
`sourceProvider.readSource(origin)` function that returns source text, `null`
for an absent approved origin, or throws for host policy/read failures. The
Node CLI wires this to `node_source_provider.mjs`; a browser host can provide
virtual paths and approved source bytes while retaining the same resolver
ordering and validation. This is not yet a browser application entry point or
the Go/WASM filesystem ABI.

`source_request.mjs` composes that provider and a caller-owned parser function
into the JSON-ready request for `__able_eval_request_json`. It is also portable
and is used by the Node CLI, so a browser host need only provide parser setup,
source policy, output callbacks, and its Go/WASM bootstrap. Its virtual-source
map contract is checked without npm dependencies:

```bash
node source_request.test.mjs
```

For a real browser Go/WASM smoke that parses the checked-in three-module Able
sources, run the following on a machine with Firefox and geckodriver:

```bash
cd v12/wasm && npm ci
cd ../.. && just wasm-browser-smoke
```

This test serves the parser runtime, Able grammar, and only the approved source
origins to the browser. It validates browser startup, real source parsing,
source-policy composition, output forwarding, and both interpreters; it does
not add the Go/WASM filesystem ABI.

The standalone resolver test has no npm dependency:

```bash
node module_loader.test.mjs
```

To execute the source/module smoke, install the optional parser dependency and
run the root target:

```bash
cd v12/wasm && npm ci
cd ../.. && just wasm-source-module-smoke
```

## Build the runtime

```bash
cd v12/interpreters/go
GOOS=js GOARCH=wasm go build -o ../../wasm/ablewasm.wasm ./cmd/ablewasm
```

## Install JS dependencies

```bash
cd v12/wasm
npm install
```

## Run the prototype

```bash
cd v12/wasm
node run_prototype.mjs --source ./samples/addition.able --wasm ./ablewasm.wasm
```

For a parser-independent smoke test (and for hosts that already produce fixture
AST JSON), run:

```bash
node run_prototype.mjs --module-json ./samples/addition.ast.json --wasm ./ablewasm.wasm --exec-mode bytecode
```

The runner prints a JSON payload containing:

- the request (AST JSON sent to wasm)
- the runtime response (`ok`, `result`, `error`, diagnostics)
- captured `able_host` stdout/stderr writes

The Go `js/wasm` runtime owns WebAssembly's low-level import object, so this
prototype maps the ABI method names to UTF-8 JavaScript strings rather than
direct pointer/length imports. A non-Go embedding can implement the raw-memory
form described in `../docs/wasm-host-abi.md` without changing Able semantics.
