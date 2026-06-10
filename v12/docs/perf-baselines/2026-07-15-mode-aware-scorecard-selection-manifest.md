# Mode-aware scorecard selection manifest

## Decision

Keep the full external scorecard as the complete status ledger, including
timeouts, failures, partial rows, source fingerprints, and verifier/input
contracts. Use the separate reviewed mode-aware manifest only to decide which
rows must carry exactly five successful samples in strict two-cohort variance.

This resolves the all-or-nothing failure exposed by bytecode BinaryTrees
without hiding it. A timeout remains visible and unranked in each full status
cohort; it is merely outside candidate-selection variance until a bounded
process completes and a reviewer admits it.

No compiler, bytecode VM, canonical stdlib, benchmark workload, foreign
reference, or promoted scorecard changed.

## Reviewed manifest

`v12/bench-selection-manifest.json` is schema version 1 and lists benchmarks
explicitly under `compiled` and `bytecode`. Its semantic SHA-256 is
`11caaee63c66fa2e235249640fe0ce44833dc6ed9946d7d0a1e840997345c132`.

- compiled: all 32 portable candidate-selection applications;
- bytecode: 26 currently bounded applications;
- full status only for now: bytecode BinaryTrees, QuickSort, Sudoku Masks,
  NBody, TapeLang Alphabet, and Regex Suffix Audit.

The historical current scoreboard classifies 57 of the 58 selected rows as
verified and K-Nucleotide bytecode as incomplete (two successes and one
45-second timeout). The new refresh default is 90 seconds, so K-Nucleotide
remains selected pending a current five-run confirmation. None of the six
excluded rows is currently verified and therefore none is suggested for
re-admission.

## Cohort lifecycle

`bench_refresh_external_scorecard` validates the manifest before launching
any timing process. Every generated aggregate embeds:

- the manifest path and exact file SHA-256;
- a canonical semantic selection SHA-256;
- the normalized per-mode benchmark lists and row count;
- the complete full-status comparison-source manifest and canonical stdlib
  state.

The aggregate still contains all 64 candidate application/mode status rows.
Promotion through `bench_external_scoreboard --cohort ... --write-current`
replays the embedded selection record along with the exact comparison sources
and stdlib fingerprint.

Strict variance now uses:

```sh
just bench-variance-report \
  --scorecard v12/docs/perf-baselines/candidate-a-refresh.json \
  --scorecard v12/docs/perf-baselines/candidate-b-refresh.json \
  --selection-manifest v12/bench-selection-manifest.json \
  --require-runs 5 \
  --output-json /tmp/able-selection-variance.json \
  --output-md /tmp/able-selection-variance.md
```

It rejects the comparison unless:

1. both aggregates embed the exact reviewed manifest record;
2. both retain identical complete full-status coverage;
3. both name disjoint comparison-source reports;
4. both use the same canonical stdlib source state;
5. every selected Able row retains exactly five successful verifier-backed
   samples and no timeout/failure;
6. every selected reference component retains exactly five measured,
   verifier-backed samples and a usable ratio.

Unselected timeout rows are not passed through the successful-run check, but
they remain present in both aggregates and their source reports. Focused
`--input` variance retains its existing behavior and does not need a manifest.
Strict complete-scorecard variance now requires one.

## Re-admission

The fast status check is:

```sh
./v12/bench_selection_manifest_check \
  --status-scorecard v12/docs/perf-baselines/external-scoreboard-current.json
```

It validates that selected benchmarks are part of portable `coverage`, that
compiled selection exactly equals all 32 coverage applications, and that each
selected row exists in the status scorecard. Any excluded portable row whose
Able status becomes `verified` is printed as `review for re-admission`; the
tool never edits the reviewed manifest automatically.

## Fast verification

`just bench-selection-check` runs the static checker and six protocol tests in
about 0.05 seconds. The tests prove:

- a timeout row remains in full status while absent from selected variance;
- a timeout selected by the manifest fails strict evidence;
- aggregate construction embeds the manifest without dropping timeout rows;
- manifest identity and full-status coverage must match between cohorts;
- canonical stdlib source state must match between cohorts;
- a completed excluded row is surfaced for review and re-admission.

The same check is part of `v12/run_all_tests.sh`. Current scoreboard replay,
catalog validation, refresh dry-run, Python compilation, and diff hygiene also
pass.

## Next recommendation

Run one focused five-process K-Nucleotide bytecode eligibility check under the
current 90-second/1-GiB guard before paying for two full cohorts. Why: it is
the manifest's only historically incomplete selected row, and a single failed
selection row would still invalidate both expensive strict cohorts. This
entails fresh verifier-backed Python/Ruby references or reuse of the exact
current-source five-run reference artifact, five independent Able bytecode
processes, and the manifest status check against a temporary full-status
aggregate.

If it completes 5/5, collect the two independent tagged cohorts and run strict
selected variance. If it does not, retain its timeout in full status, remove it
from selection through review, and profile it only if the same concrete VM
descendant recurs in at least two other unlike applications; do not optimize a
map/text parent from K-Nucleotide alone.
