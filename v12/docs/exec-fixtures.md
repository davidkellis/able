# Exec Fixtures (v12)

Exec fixtures are end-to-end Able programs that run through each interpreter and assert observable behavior (stdout/stderr/exit). They live under `v12/fixtures/exec`, and the Go harness executes them:

- Go: `cd v12/interpreters/go && go test ./pkg/interpreter -run ExecFixtures`
- Combined (fixture-only): `./run_all_tests.sh --version=v12 --fixture`
- Coverage index guard: keep `v12/fixtures/exec/coverage-index.json` in sync with fixture directories; `v12/scripts/check-exec-coverage.mjs` (called by `run_all_tests.sh`) fails when entries drift.

## Layout
Each exec fixture directory contains:
- `package.yml` — package name (e.g., `name: exec_08_01_control_flow_fizzbuzz`).
- `main.able` — entry module with `fn main() -> void`.
- `manifest.json` — expectations:
  ```json
  {
    "description": "short summary",
    "expect": {
      "stdout": ["line 1", "line 2"],
      "stderr": ["error message"],      // optional
      "exit": 0                         // optional; default 0 when no runtime error
    }
  }
  ```

Naming follows the conformance plan (`v12/docs/conformance-plan.md`): directories use `exec/<section>_<heading>_<feature>[_variation]/` keyed to spec headings (zero-padded when helpful), and package names mirror the directory with an `exec_` prefix (e.g., `exec_08_01_control_flow_fizzbuzz`).

## Authoring Guidelines
- Keep programs minimal and focused on one semantic slice (loops, pattern matching, rescue/ensure, concurrency, etc.).
- Use standard Able syntax; no host-language features. `print` accepts one argument—build strings via interpolation (`\`value ${x}\``) when needed.
- Deterministic output: avoid nondeterministic ordering unless the expectation matches the scheduler semantics being tested.
- If the program raises and should exit non-zero, set `expect.exit` and optionally `expect.stderr` to the error message.

## Running and Debugging
- Filter to a specific exec fixture in Go with `go test ./pkg/interpreter -run ExecFixtures/<name>`.
- Go harness failures show up in `go test ./pkg/interpreter -run ExecFixtures`; rerun with `-v` for verbose output.

## Adding New Cases
Target breadth over complexity: add fixtures that cover control flow (loops/match/rescue), data types (unions/structs/arrays), interfaces/impls, pipelines/UFCS, and concurrency (spawn/future/channel/mutex). Ensure expectations reflect spec-defined behavior and stay interpreter-agnostic.
