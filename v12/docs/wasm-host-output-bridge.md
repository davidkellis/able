# WASM host output bridge — 2026-07-14

## Scope and decision

The first executable `able_host` slice is complete: the pre-parsed-AST Go/WASM
evaluator now forwards observable output to JavaScript through
`globalThis.able_host.write_stdout(string)` and evaluation failures through
`globalThis.able_host.write_stderr(string)`.

This is a host-portability boundary, not a compiler or bytecode performance
change. It adds no benchmark, scheduler rule, nominal-type special case, or
canonical-stdlib source change.

## Bridge contract

Go's `js/wasm` runtime owns the low-level WebAssembly import object, so it
cannot directly add an arbitrary `able_host` import module beside the Go
runtime imports. The prototype therefore maps the ABI's existing method names
through JavaScript values:

```js
globalThis.able_host = {
  write_stdout(message) { /* UTF-8 text */ },
  write_stderr(message) { /* UTF-8 text */ },
};
```

The messages are complete UTF-8 strings. `print(value)` sends the rendered
value plus one newline to `write_stdout`; a malformed request or evaluation
failure sends its structured diagnostic plus one newline to `write_stderr`.
The JSON response remains authoritative if a host stderr writer is absent or
throws, so host reporting cannot erase the evaluation diagnostic.

The evaluator installs this `print` binding only when an `OutputSink` is
provided by a host. It is the existing fixture-style host print facility, not
a new v12 builtin or a canonical-stdlib API. Ordinary native interpreter runs
and AST evaluation without an output sink retain their prior globals.

The raw pointer/length import form in [wasm-host-abi.md](wasm-host-abi.md)
remains the intended contract for a non-Go embedding. Retaining the method
names now lets that future implementation replace only the adapter, without
changing Able observable output semantics.

## Verification

`just wasm-smoke` now builds one generated `ablewasm.wasm` binary and checks:

- the existing addition AST in tree-walker and bytecode mode;
- `host-output.ast.json` in both modes, requiring exactly
  `wasm host output\n` on host stdout; and
- `host-error.ast.json`, requiring a structured failure and its exact decoded
  diagnostic on host stderr.

Native `pkg/wasmhost` tests cover the same successful output path in both
interpreters and the JSON error-output path without requiring Node.

## Deliberately deferred

- source parsing and static/dynamic module loading;
- host filesystem and module-root callbacks;
- timer wakeups and scheduler integration; and
- browser extern callbacks.

## Follow-up

The planned static source/module closure is now complete: the portable resolver
resolves `dep -> math -> app`, sends the first two modules in ordered,
origin-labelled setup entries, and both interpreter modes import and print the
same result. A Node-only filesystem provider serves the current CLI, while a
browser host can inject approved virtual sources without duplicating resolver
rules. The portable request-composition layer now joins that parser/provider
pair into the evaluator request, and a headless Firefox smoke now proves that
path with Go/WASM, actual Able source parsing, and named output callbacks. The
next step is benchmark-guided runtime work, not unused browser host stubs. See
[wasm-source-module-loader.md](wasm-source-module-loader.md).
