# Treewalker/Bytecode Parity Reporting

v12 uses two Go interpreters (tree-walker + bytecode). Parity reporting focuses on
ensuring identical outputs, diagnostics, and exit behavior across both modes.

## Status
The parity harness is TODO. Once wired, it should:
- run each AST/exec fixture through both execution modes,
- diff stdout/stderr/exit and typechecker diagnostics,
- emit a machine-readable JSON report for CI.

## Planned CLI contract
- `--exec-mode=treewalker|bytecode` selects the interpreter mode.
- `./run_all_tests.sh --fixture` runs the Go fixture suite in both exec modes.
- A dedicated parity runner should write `v12/tmp/parity-report.json`.

## CI integration (planned)
When the parity runner lands, CI should archive `v12/tmp/parity-report.json` so
regressions are easy to triage.
