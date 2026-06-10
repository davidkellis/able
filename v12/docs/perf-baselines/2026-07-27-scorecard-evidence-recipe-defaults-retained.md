# Scorecard evidence recipe defaults retained

Date: 2026-07-27

Decision: retain one general benchmark-tooling correction, retain no production
performance change, and keep the evidence-gated performance pause.

## Trigger

The second post-consolidation invalidation audit found no changed compiler,
runtime, VM, language, canonical-stdlib, benchmark, reference, or evidence
input. The worktree complement remained exactly the 34 deferred WASM paths.

The audit did expose one general integrity defect. The documented
`just bench-scorecard-evidence-check` recipe invoked
`bench_scorecard_evidence_check.py` without its three required arguments, so
the no-argument gate exited with an argument-parser error instead of checking
the current scorecard.

## Retained correction

The recipe now supplies the canonical current defaults:

- `v12/docs/perf-baselines/external-scoreboard-current.json`;
- `v12/bench-selection-manifest.json`;
- five required successful Able/reference runs.

Trailing recipe arguments remain available for an explicit alternate review.
No benchmark program, measurement, selection, performance disposition, or
production execution path changed.

## Verification

- portable catalog: 61 applications;
- feature coverage: 15 families and 16 normative sections;
- operation-depth frontier: zero actionable entries;
- reviewed selection: 115 rows with SHA-256
  `1dc70106786a7e668982f070428fed3a81f77ba2abb4adf72d97848265f9dead`;
- scorecard evidence: 115 selected rows and 122 full-status rows, with five
  successful Able/reference samples for every selected row;
- current frontier: 115 rows and zero actionable groups;
- evidence ledger: 21 current closures and zero invalidations;
- complete deterministic architecture budget: pass;
- `git diff --check`: pass.

The deterministic gate run completed in approximately 13 seconds. Final
cleanup removed 309 MiB of reproducible project-local Go cache, 60 KiB of
Python bytecode, and an empty local temporary directory. The reusable 73 MiB
disk-backed extern cache was retained.

## Next recommendation

Keep production performance mutation paused until an authoritative
invalidation appears.

Why: the current source identities, complete repeated evidence, frontier, and
closure ledger agree that no general production candidate is open.

What it entails: when a new broad application, retained semantic/source
change, correctness failure, or new non-closed owner across three unlike
applications appears, refresh only the affected evidence before evaluating one
general lowering/runtime rule.

Why it is important: the corrected gate now enforces the five-run admission
standard directly, while the pause prevents benchmark-specific work or
unnecessary crossings of the compiled/interpreted boundary. Do not begin WASM
work.
