# Parity Reporting & CI Integration

The cross-interpreter parity harness (`v11/interpreters/ts/scripts/run-parity.ts`) runs inside `./run_all_tests.sh`. Every run produces `tmp/parity-report.json`, a machine-readable summary that records the Bun vs. Go parity status for fixtures and curated examples.

## Generating reports locally

```bash
cd v11/interpreters/ts
bun run scripts/run-parity.ts --suite fixtures --suite examples --json
```

- `--suite` selects which suites to execute (fixtures, examples, or both).
- `ABLE_PARITY_MAX_FIXTURES` limits the number of AST fixtures processed (handy while debugging).
- `ABLE_TYPECHECK_FIXTURES=warn|strict` keeps diagnostics aligned between interpreters during parity runs.
- The CLI always writes `tmp/parity-report.json` relative to the repo root; pass `--report <path>` if you want to override the destination while optionally mirroring the payload to stdout with `--json`.

## CI workflow

Before calling `./run_all_tests.sh` (or running the parity CLI directly), set one (or both) of:

- `ABLE_PARITY_REPORT_DEST=/path/to/artifacts/parity-report.json`
- `CI_ARTIFACTS_DIR=/path/to/artifacts` (the script writes `${CI_ARTIFACTS_DIR}/parity-report.json`)

That ensures every full test run copies `tmp/parity-report.json` into a location your CI system can archive. Example GitHub Actions steps:

```yaml
- name: Run Able parity suite
  run: |
    export ABLE_PARITY_REPORT_DEST="$GITHUB_WORKSPACE/artifacts/parity-report.json"
    ./run_all_tests.sh --typecheck-fixtures=warn
- name: Upload parity report
  uses: actions/upload-artifact@v4
  with:
    name: able-parity
    path: artifacts/parity-report.json
```

## Report structure

Each suite contributes an entry:

```jsonc
{
  "suite": "fixtures",
  "root": "/abs/path/fixtures/ast",
  "total": 200,
  "passed": 200,
  "failed": 0,
  "skipped": 0,
  "entries": [
    { "name": "structs/functional_update", "status": "ok" }
  ]
}
```

Failed entries include a `reason` plus structured diffs (`diff`, `tsOutcome`, `goOutcome`) so downstream consumers can display precise mismatches.

## Troubleshooting tips

- Use `ABLE_PARITY_MAX_FIXTURES` to focus on the first failing AST case.
- Re-run `bun run scripts/run-parity.ts --suite fixtures --json` locally to verify fixes before pushing.
- The CLI caches the Go binaries per invocation; clear your build cache if Go dependencies change and the harness starts failing to compile.
