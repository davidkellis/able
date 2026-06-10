# Scorecard evidence hermeticity and full correctness gate

Date: 2026-07-27

## Decision

Retain a general scorecard-evidence hermeticity gate and make the canonical
stdlib test runner honor an explicit disk-backed `GOCACHE`. Retain no compiler,
generated-runtime, runtime, interpreter, bytecode VM, canonical-stdlib,
language, dependency, benchmark workload, or WASM production change.

The performance invalidation selector remains empty: all 21 closures are
current. `spec/TODO_v12.md` has no tracked compiler AOT or stdlib
externalization gap. The appropriate next tranche was therefore the full v12
correctness gate, not another speculative production optimization.

## Failure found

The first `./run_all_tests.sh` attempt stopped before Go tests:

```text
bench_external_scoreboard: input file not found:
/var/tmp/able-backup-dedup-evidence/go-reference.json
```

The selected Backup Dedup comparison reports embedded absolute references to
disposable `/var/tmp` Go and interpreter reports even though exact retained
copies already existed under `v12/docs/perf-baselines`. Cleaning temporary
files therefore made the checked-in current scoreboard non-reproducible.

Both comparison JSON reports now name their retained repository artifacts:

- `2026-07-27-backup-dedup-go-reference.json`;
- `2026-07-27-backup-dedup-interpreter-reference.json`.

Their readable Markdown companions record the same durable locations.

## General retained guards

`bench_scorecard_evidence_check.py` now validates dependency hermeticity for a
scorecard retained under `v12/docs/perf-baselines`. Every source report and
each `go_reference_json` or `reference_json` dependency must resolve to an
existing file under that retained directory. An existing `/tmp`, `/var/tmp`,
or other external file is rejected before cleanup can expose the defect.

The check is part of both the ordinary v12 test runner and
`just bench-scoreboard-check`. Its focused regression proves that a retained
reference is accepted and an existing external temporary reference is
rejected. The current cohort verifies:

- 119 selected rows and 126 full-status rows;
- five successful Able/reference samples for every selected row;
- 20 retained comparison sources;
- 33 retained Go/Python/Ruby reference reports;
- selection SHA-256
  `e6a6ccacc9620f9e1b89e2510cab52a85114ddbf2f41e33abae7c6d8a70241f8`.

`run_stdlib_tests.sh` now uses an explicitly supplied `GOCACHE` instead of
unconditionally creating `v12/interpreters/go/.gocache`. Its default remains
unchanged when no environment override is supplied.

## Verification

The default full v12 gate passed after the evidence repair:

- all deterministic scoreboard, coverage, selection, execution-contract,
  threshold, cleanup, and embedded-kernel checks;
- all non-compiler Go packages in short mode;
- all 33 bounded compiler short-mode batches;
- the complete bytecode fixture pass.

Three compiler batches had aggregate package times of 137.417, 70.068, and
113.622 seconds. Per-test JSON timing confirmed that these totals accumulated
multiple sub-minute tests. Their slowest individual tests were 37.97, 16.05,
and 10.95 seconds respectively, so the repository's one-minute individual
test limit remains satisfied.

The canonical external stdlib also passed:

- tree-walker: 20 seconds;
- bytecode: 15 seconds.

Both ran with `TMPDIR` under disk-backed `/var/tmp` and
`GOCACHE=/var/tmp/able-go-build-cache`; no repository-local `.gocache` was
created. Focused scorecard hermeticity, five-run evidence, ledger, shell syntax,
and file-length checks pass. The 999-line
`bench_external_scoreboard` was not enlarged.

Cleanup removed the exact 66 MiB tranche workspace under `/var/tmp` and a
stale, unopened 4.1 MiB Able extern cache under RAM-backed `/tmp`. The reusable
disk-backed Go build cache was preserved. Deferred WASM paths were not touched.

## Next recommendation

Keep production performance mutation paused until a genuine application,
semantic correction, canonical-stdlib change, or source-identity change
invalidates the checked frontier.

Why: the full repository and both canonical-stdlib engines are green, the
scorecard is now self-contained, all 21 performance closures are current, and
no tracked AOT gap or shared non-closed performance owner exists.

What it entails: run the evidence selector after an authoritative change,
refresh only the invalidated rows and closures, and profile at least three
unlike applications. Admit production code only when one exact non-closed
owner repeats across all three and survives balanced verifier-backed
baseline/candidate/reference measurements.

Why it is important: this keeps cleanup safe, makes retained measurements
reproducible, and prevents correct native Go lowering from being destabilized
by another closed or benchmark-specific experiment. Do not begin WASM work.
