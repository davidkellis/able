# WASM Pre-Parsed-AST Prototype Status

Date: 2026-07-14

## Purpose

The current WASM slice proves a deliberately narrow boundary: JavaScript can
give the Go runtime a fixture-style Able AST JSON module and receive the result
from either the tree-walker or bytecode interpreter. It is not a source-loading
or browser-hosted application runtime yet.

## What is supported

- `GOOS=js GOARCH=wasm go build ./cmd/ablewasm` produces the bridge.
- `__able_eval_request_json(requestJson)` accepts `EvaluateRequest` JSON.
- The request selects `treewalker` or `bytecode`, accepts setup AST modules,
  and returns the rendered result or a structured error.
- The portable static source resolver can parse a small dependency closure into
  ordered, origin-labelled setup AST modules. It resolves
  `<root>/<package>.able` and `<root>/<package>/main.able` from the entry
  directory plus ordered roots, rejects cycles and `dynimport`, and runs the
  closure in both interpreter modes. The Node CLI supplies a filesystem
  provider; a browser host can inject an approved virtual-source provider
  without duplicating resolver rules.
- The portable request-composition adapter turns a caller-owned parser bridge
  plus that provider into the exact evaluator request (`execMode`, ordered
  setup modules, entry module, and entry origin). The Node CLI uses this same
  adapter; a virtual-source-map test covers the browser-facing contract without
  Node filesystem or Go/WASM startup.
- `just wasm-browser-smoke` starts the Go/WASM binary in headless Firefox via
  a local WebDriver harness. It serves only the virtual `dep -> math -> app`
  source origins plus the parser runtime and grammar, then verifies `42` plus
  `42\n` in both runtime modes. The portable AST mapper is shared with the
  Node CLI; browser and Node parser bootstraps remain intentionally separate.
- The JS host can provide `globalThis.able_host.write_stdout(string)` and
  `write_stderr(string)`. The evaluator's host-provided `print` helper and
  structured failures forward there; the Node smoke verifies both channels.
- `just wasm-smoke` builds the ignored `v12/wasm/ablewasm.wasm` artifact and
  validates `1 + 2 == 3` in both execution modes through Node.
- `just clean` and `just cleanup-apply` remove that generated WASM artifact.

## Deliberately unsupported at this boundary

- Go-native source parsing and typechecking: the Go parser depends on the
  native tree-sitter language package. JavaScript supplies ASTs instead. The
  current JavaScript source adapter is deliberately limited; its portable AST
  mapper runs in the Node CLI and browser smoke, while parser bootstraps remain
  host-specific.
- Typechecking and typechecker-derived bytecode proof metadata: callers must
  provide an AST appropriate for the runtime boundary; enabling the native
  typechecker reports an explicit unsupported error.
- Dynamic source evaluation: it reports an explicit `js/wasm` unsupported
  error.
- Go-plugin extern functions: they report an explicit browser-host-callback
  unsupported error. `just wasm-smoke` evaluates an `extern go` call in both
  interpreter modes and asserts both the structured error and its stderr
  forwarding, so this deliberate boundary cannot silently become a missing
  binding or a target-specific fallback.
- The proposed `able_host` filesystem, timer, output, and module-root ABI:
  only the Go/JS output adapter is executable; filesystem, timer, and
  module-root work remain design-only in [wasm-host-abi.md](wasm-host-abi.md).
- Broader JavaScript AST mapping beyond the current small expression/import
  subset. The shared mapper and browser bootstrap currently cover the checked-
  in source smoke only.

## Boundary rationale

The AST evaluator is shared code. Native-only services now sit behind
platform-specific files, so the browser build does not import the native parser
or `plugin` package merely to execute an already decoded AST. Generic value
coercion and byte-array helpers remain shared; only the unavailable plugin
loader and its native-only service APIs are excluded. This prevents a browser
target from silently acquiring different execution semantics just to link.

## Verification

```bash
cd v12/interpreters/go
GOOS=js GOARCH=wasm go build ./cmd/ablewasm
GOOS=js GOARCH=wasm go test -c ./pkg/wasmhost
go test ./pkg/wasmhost -count=1 -timeout 60s

cd ../..
just wasm-smoke
```

The Node smoke intentionally uses `samples/addition.ast.json`, so it does not
need `web-tree-sitter` or installed npm dependencies. The normal source route
continues to use the existing JavaScript parser adapter when those dependencies
are installed.

The same smoke also runs the host-output, host-error, and unavailable Go extern
AST samples under the Go/JS adapter. See
[wasm-host-output-bridge.md](wasm-host-output-bridge.md) for the exact boundary
and its intentionally deferred host services.

`just wasm-smoke` also runs the dependency-free virtual-source request test.
`just wasm-source-module-smoke` adds the optional web-tree-sitter-backed
three-module source closure check after `cd v12/wasm && npm ci`. See
[wasm-source-module-loader.md](wasm-source-module-loader.md) for its current
layout, request contract, and bounded-performance-gate follow-up.
