# Performance-stability manifest and frontier integration

Date: 2026-07-20

## Decision

Retain a versioned cross-cohort stability manifest and render its result as a
separate layer in the complete performance frontier. Retain no compiler,
bytecode VM, runtime, parser, canonical-stdlib, benchmark, fixture, language,
nominal-lowering, or WASM performance change.

The promoted scorecard remains the sole authority for current snapshot status:
8 of 75 selected rows meet and 67 miss. The stability layer classifies all
eight snapshot meets and establishes six candidate-admission guards: four
compiled and two bytecode. Compiled Monte Carlo Pi remains a volatile crossing,
and bytecode Base64 remains a variance-sensitive pooled miss.

## Manifest contract

`v12/bench-performance-stability.json` is deliberately sparse: it contains
exactly the selected rows whose current scorecard status is `meets`. Each entry
records:

- benchmark and mode;
- `established-meet`, `volatile-crossing`, or `variance-sensitive-miss`;
- pooled limiting ratio and every independent cohort ratio;
- Able and limiting-reference sample counts;
- current Able and every applicable reference source SHA-256;
- the canonical-stdlib tree SHA-256 used by the stability evidence;
- evidence paths, which the generated frontier expands to path/hash records;
  and
- a concise reviewer rationale.

The manifest also pins the reviewed selection SHA-256 and current promoted
scorecard's canonical-stdlib tree SHA-256. The frontier rejects an entry outside
the selection, duplicate or missing snapshot meets, entries for snapshot
misses, stale source fingerprints, stale current stdlib identity, missing
evidence, fewer than ten samples per implementation, and classifications that
contradict the cohort/pooled ratios. A newly promoted snapshot crossing cannot
silently become a regression guard: it makes the fast frontier check fail until
independent evidence is reviewed.

## Generated frontier

The frontier schema is now version 2. Its summary retains the raw scorecard
counts and separately reports:

- 6 established guards;
- 4 compiled and 2 bytecode established guards; and
- 2 unestablished snapshot meets.

Every selected ledger row includes snapshot status, established-guard status,
and stability classification. The JSON carries the complete stability record;
the Markdown renders a focused eight-row evidence section with pooled/cohort
ratios, samples, source/stdlib fingerprints, rationales, and evidence links.
Ownership dispositions, target excess, actionable-group ranking, and raw
scorecard status are unchanged: 75 rows, 121.969 seconds aggregate excess, and
zero actionable groups.

## Verification

- five focused frontier tests pass, covering complete ownership evidence,
  snapshot-meet stability coverage, established versus volatile status, and
  source-fingerprint invalidation;
- `just bench-frontier-check` reproduces the schema-2 JSON and Markdown exactly;
- `just bench-scoreboard-check` confirms the promoted scorecard was not changed;
- all manifest and generated JSON parses successfully;
- both Python files remain below 1,000 lines; and
- `git diff --check` passes.

## Source-identity finding and next recommendation

Completed by `2026-07-20-source-exact-established-guard-refresh.md`. All five
lanes remain established after two independent current-source cohorts, so all
six durable guards now use the promoted canonical-stdlib tree identity.
