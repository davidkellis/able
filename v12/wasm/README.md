# Able WASM Prototype

This directory contains a minimal end-to-end prototype for:

1. Parsing Able source in JavaScript with tree-sitter (`web-tree-sitter` + `tree-sitter-able.wasm`).
2. Converting the parse tree into fixture-style AST JSON for a narrow syntax subset.
3. Sending that JSON into the Go/WASM runtime via `__able_eval_request_json`.

The adapter currently supports a deliberately small subset:

- package statements (`package name`)
- static/dynamic import statements
- identifiers, numbers, booleans, strings, nil
- assignment expressions
- binary expressions
- postfix member access + function calls

If the source uses unsupported constructs, the adapter exits with an explicit error.

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

The runner prints a JSON payload containing:

- the request (AST JSON sent to wasm)
- the runtime response (`ok`, `result`, `error`, diagnostics)
