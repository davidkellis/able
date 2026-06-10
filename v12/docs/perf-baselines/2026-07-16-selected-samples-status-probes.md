# Selected Samples and Excluded Status Probes

## Decision

Keep no compiler, bytecode VM, bridge, generated-runtime, canonical-stdlib, or
benchmark-source change from this tranche. The proposed generated execution-
context direction first collapsed to a program-level `spawn` selector for the
existing fixed-context ABI. Repository reconciliation found that exact generic
candidate had already been measured and rejected on 2026-07-14: it improved
Channel Rollup and Future Pipeline, but regressed Mutex Ledger by 10.0% and was
fully reverted. This tranche's duplicate implementation was also fully removed.

The current profile/candidate inventory closes the remaining unchanged
fixed-context variants and the visible compiler/VM leaves. Ratios alone do not
justify another performance patch. The retained change instead removes a
measurement-system cost while preserving the evidence contract.

## Collection contract

`bench_refresh_external_scorecard` now reads the reviewed mode-aware selection
manifest before building its commands and partitions both Able runs and fresh
Python/Ruby references:

- all 58 strict-selection rows receive the requested five independent timed
  processes;
- all six excluded bytecode rows receive one fresh bounded status probe;
- all 64 benchmark/mode rows remain in the aggregate full-status scorecard;
- all 32 compiled rows remain five-run because compiled selection is required
  to equal the complete portable catalog; and
- strict cohort variance remains unchanged: it requires exactly five verified
  Able and applicable reference samples for every selected row and ignores an
  excluded row only after proving identical full-status coverage.

The partition preserves catalog order and is derived from the manifest rather
than from benchmark names. Mixed bounded groups are split into separately named
`selected` and `status` source reports, so aggregation never combines different
run policies in one command. A changed valid manifest automatically changes the
commands.

## Expected collection-time reduction

The six excluded Able bytecode rows previously consumed five 90-second timeout
attempts apiece. Retaining one probe avoids 24 attempts, or 36 minutes per
cohort. In promoted cohort B, the excluded Python/Ruby reference means imply
another 18.8 minutes avoided by retaining one rather than five processes. The
expected reduction is therefore about 54.8 minutes per full cohort, or roughly
1 hour 50 minutes for the normal independent two-cohort refresh, before small
launch overheads. This estimate changes no measured performance result.

## Verification

- `just bench-selection-check`: eight tests pass (six selection/variance tests
  and two refresh-partition dry-run tests).
- The dry-run tests prove every one of the 64 rows appears exactly once, the 58
  selected rows use five runs, and the six excluded rows and their interpreter
  references use one run.
- A second dry-run test removes Fib bytecode from a temporary valid manifest and
  proves the generated commands move it from five samples to one, guarding
  against hard-coded exclusions.
- `bench_selection_manifest_check --status-scorecard ...`: all 58 selected
  current rows are verified.
- `just bench-scoreboard-check`: current promoted scoreboard replay passes.
- Shell syntax and Python compilation checks pass.

No full cohort was launched because the change affects collection scheduling,
not runtime performance, and its complete command graph is deterministically
testable with `--dry-run`.

## Follow-up resolution

The feature-led audit found no uncovered portable v12 family, matching the
retained July 15 reconciliation. It added an executable 16-section/32-
application coverage contract rather than inventing a benchmark, and retained
no performance code. The result and next source-equivalence recommendation are
recorded in `2026-07-16-feature-coverage-contract.md`.
