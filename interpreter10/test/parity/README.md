# Parity Test Notes

The suites in this directory (`fixtures_parity.test.ts` and `examples_parity.test.ts`) run both the TypeScript **and** Go interpreters for every case so we can diff their observable behaviour.

- Each Bun test first evaluates the program/fixture via the TypeScript interpreter.
- It then shells out to a compiled Go helper (`interpreter10-go/cmd/fixture` for AST fixtures, `interpreter10-go/cmd/able` for the curated example programs) and compares results, stdout, and diagnostics.
- To keep per-test overhead low we build each Go helper once at suite startup (the initial `go build` accounts for the ~10 s run-up you see in the logs) and reuse the resulting binary for every test. After that, individual cases run in a few milliseconds.
- Example programs live under `interpreter10/testdata/examples/` (currently: `hello_world`, `control_flow`, `errors_rescue`, `structs_translate`, `concurrency_proc`, `generics_typeclass`). Add new coverage there rather than using the docs-facing `examples/` directory so tests stay deterministic, and only check in programs that the current parser + both interpreters fully support (e.g., match guards and advanced pipes are still blocked—track readiness in `design/parser-ast-coverage.md` / `design/concurrency-fixture-plan.md` before adding those cases).

If you add new parity suites, follow the same pattern: build the relevant Go binary once, cache it for the duration of the test file, and keep the shared explanation here in mind when folks wonder why TypeScript tests need Go artifacts.***
