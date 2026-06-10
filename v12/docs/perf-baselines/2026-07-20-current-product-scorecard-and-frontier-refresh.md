# Current product scorecard and frontier refresh

Date: 2026-07-20

## Decision

Promote a fresh verifier-backed external scorecard and regenerate the complete
cross-mode performance frontier. Retain no compiler, bytecode VM, runtime,
parser, canonical-stdlib, benchmark, fixture, language, nominal-lowering, or
WASM performance change.

The selected product score is now 5/41 compiled rows at the 95%-of-Go target
and 3/34 bytecode rows at both the 95%-of-Python and 95%-of-Ruby targets. The
joined frontier therefore has 8 target meets, 67 misses, 121.969 seconds above
the aggregate per-row budget, and zero unclosed implementation groups.

## Measurement contract

`bench_refresh_external_scorecard` ran the authoritative mode-aware selection
with `--runs 5`, a 55-second per-process timeout, `GOMEMLIMIT=1GiB`, and
`GOGC=50`. It used the catalog's CPU budget, executor, working directory,
arguments, source-root policy, and public verifier for every application.

- all 75 selected rows have five successful Able samples;
- every successful Able sample passed its public verifier;
- every selected row has five successful fresh applicable Go or Python/Ruby
  reference samples;
- the scorecard contains all 82 compiled/bytecode status rows and preserves
  the seven unselected bytecode probes;
- the five unselected Able timeouts remain timeouts rather than inferred
  ratios;
- Able sources, foreign sources, declared inputs, and verifiers have recorded
  fingerprints; and
- the canonical external stdlib records 70 `.able` files and tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.

The evidence checker passes with selection SHA-256
`f849d4ec36406fbb5c6739ccb41203e6d6b39450c61b0b894211e0d5863519a3`.
The canonical aggregate is
`2026-07-20-current-product-scorecard-refresh.{json,md}` and is promoted to
`external-scoreboard-current.{json,md}`.

## Target changes

The preceding frontier contained five target meets and 125.769 seconds of
aggregate excess. The current refresh contains eight meets and 121.969
seconds of excess, a 3.800-second reduction. Three rows crossed from miss to
meet:

| Application | Mode | Previous ratio | Current ratio | Current interpretation |
| --- | --- | ---: | ---: | --- |
| Base64 | compiled | 1.057x Go | 1.028x Go | meet, 2.35% inside the allowed ratio |
| Base64 | bytecode | 1.329x worst reference | 1.048x Ruby | meet, only 0.48% inside the allowed ratio |
| Monte Carlo Pi | compiled | 1.063x Go | 0.927x Go | meet, but historically threshold-sensitive |

These are current five-run arithmetic means, not permission to weaken their
guards. Base64 bytecode in particular is too close to the 1.052632 ratio
budget to call established from one refreshed cohort. Binary Trees,
QuickSort, Base64, JSON, and Monte Carlo Pi are the five current compiled
meets; Base64, JSON, and PiDigits are the three selected bytecode meets.
Matrix Multiply bytecode also completes faster than both references, but it
remains an unselected full-status probe and is not counted in 3/34.

## Remaining frontier

The largest absolute excess remains K-Nucleotide bytecode at 43.839 seconds.
The largest aggregate groups are:

| Group | Misses | Excess seconds | Disposition |
| --- | ---: | ---: | --- |
| bytecode text/map | 5 | 51.681 | closed: no shared material leaf after the retained hash index |
| bytecode wide numeric | 3 | 16.595 | closed: generic carrier/extractor candidates rejected by broad guards |
| bytecode float numeric | 4 | 15.317 | closed: typed-lane/carrier candidates rejected by broad guards |
| bytecode regex | 5 | 15.092 | closed: generic NFA/Array/call alternatives already retained or rejected |
| bytecode byte output | 2 of 3 | 4.877 | closed: retained raw-u8 path leaves different residual owners |
| compiled text/map | 5 | 4.376 | closed: no shared material leaf after the retained hash index |
| compiled concurrency | 7 | 4.104 | closed: fixed-context candidate regressed unlike guards |

Fresh timings change the product ranking but do not invalidate any recorded
closure. The frontier still has no exact, non-nominal, previously untried
descendant that is material in at least three unlike applications. Reopening
raw-cell, return/frame, lookup-cache, map, typed-lane, regex, scheduler, or
named-container work from a parent ratio would repeat an already failed gate.

## Verification

- scorecard evidence check: 75 selected rows, 82 full-status rows, five
  successful Able/reference samples for every selected row;
- regenerated frontier check: 75 selected rows and zero actionable groups;
- promoted scoreboard structural check;
- focused scorecard-selection/frontier unit tests; and
- `git diff --check`.

The deferred project-cleanup follow-up is also complete. The conservative
`just cleanup-apply` path now discovers reproducible Python `__pycache__`,
pytest, mypy, and Ruff caches inside active v12 in addition to the existing Go
caches, benchmark workspaces, fixture targets, test binaries, and optional
profile archives. It still refuses tracked paths and never traverses the
frozen v10/v11 workspaces or external `able-stdlib`. This tranche reclaimed
the 213 MiB local Go build cache plus its newly generated Python bytecode
cache; a final preview reports no remaining selected artifact.

## Next recommendation

Follow-up completed by
`2026-07-20-threshold-stability-reconciliation.md`. Fifteen-sample pools make
compiled Base64 an established guard, classify bytecode Base64 as a
variance-sensitive miss, and leave compiled Monte Carlo volatile.

The completed recommendation was to run a threshold-stability reconciliation
for the three new target crossings: compiled Base64, bytecode Base64, and
compiled Monte Carlo Pi.

Why: all three changed classification relative to the preceding current
frontier, and bytecode Base64 is only 0.48% inside the allowed ratio. On this
workstation, one five-sample cohort is the correct current score but is not
enough to decide whether these rows are durable target guards or variance-
sensitive provisional meets. That distinction matters before any future
candidate is allowed to regress them.

What it entails: preserve one current compiled binary and one current
bytecode binary, refresh the matching Go/Python/Ruby executables, run at least
three independently order-balanced five-sample cohorts under the same CPU,
memory, timeout, input, and verifier contracts, retain every non-timeout sample
in arithmetic averages, and pool all fifteen samples per lane. Update only the
guard classification and frontier evidence; do not tune benchmark sources or
advance implementation code unless a separate three-unlike-application
profile invalidates a closed group.
