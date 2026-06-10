# Post-Quickening Bytecode Scorecard Reconciliation (2026-07-18)

## Decision

Promote the complete selected-bytecode refresh to
`external-scoreboard-current.{json,md}`. Retain the generic
comparison/conditional-branch quickening from the preceding gate, but retain no
additional VM, compiler, runtime, benchmark, fixture, language, or canonical
`able-stdlib` change from this measurement tranche.

The promoted product result remains:

- compiled: 7 of 35 selected rows meet the 95%-of-Go target;
- bytecode: 3 of 27 selected rows meet both the 95%-of-Python and
  95%-of-Ruby targets;
- bytecode meets: Base64, JSON, and PiDigits; and
- full status: all 70 application/mode rows remain visible.

The quickening moved application time in the right direction broadly enough to
retain, but it did not move another application across the product threshold.
The remaining gap is still large and distributed across unlike workload
families.

## Collection protocol

The unchanged reviewed manifest selects 27 bounded bytecode rows. Each received
five independent Able processes under its catalog CPU/executor contract, CPU
pool 0-3, `GOMEMLIMIT=1GiB`, `GOGC=50`, the canonical external stdlib, and a
55-second per-process cap. Every successful process passed its public Ruby
verifier. Existing five-run Python/Ruby references were reused only when their
source, input, verifier, and execution-contract fingerprints matched.

The primary refresh comprises 135/135 verified Able processes in three
non-overlapping source reports:

- `2026-07-18-post-quickening-bytecode-generality.json`;
- `2026-07-18-post-quickening-bytecode-async.json`; and
- `2026-07-18-post-quickening-bytecode-coverage.json`.

Fifteen volatile or decision-sensitive rows received independent five-run
expansion cohorts: all six concurrency rows; I-Before-E, Base64, and JSON; and
Fixed Width 128, Rational Series, Word Frequency, Lexical Rollup, Array Slice
Window, and Option/Result Config. Those add 75/75 verified processes. No sample
or outlier was discarded, for 210/210 verified Able launches in this tranche.

The scorecard promotion uses exactly one complete five-run source per selected
row, as required by the strict evidence checker. Expansion reports remain
companion evidence and their complete ten-run pooled averages govern the
volatility discussion below.

The canonical stdlib state is unchanged from the preceding promoted scorecard:
70 `.able` files, tree SHA-256
`f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`,
Git `219eff222c28406487231713753641bc49ee5b9a`, dirty.

## Reconciliation

Against the preceding five-run bytecode cohort, 23 of 27 selected primary means
improve and four regress. The median row moves -7.65%; the unweighted sum of row
means moves from 114.656 seconds to 108.284 seconds (-5.56%). These cohort
statistics describe the current workstation scorecard; they do not by
themselves attribute every change to the quickened opcode. The preceding
order-balanced preserved-binary A/B remains the causal gate for that code.

Material or decision-sensitive primary movements include:

| Application | Previous | Primary refresh | Change | Reconciliation |
| --- | ---: | ---: | ---: | --- |
| Mutex/Await Journal | 0.300 s | 0.224 s | -25.33% | 10-run pooled 0.219 s; short-row cohort movement, not attributed to comparison quickening |
| Mutex Ledger | 0.508 s | 0.388 s | -23.62% | 10-run pooled 0.405 s |
| Channel Rollup | 0.712 s | 0.566 s | -20.51% | 10-run pooled 0.556 s |
| Await/Channel Mux | 0.312 s | 0.252 s | -19.23% | 10-run pooled 0.238 s |
| Monte Carlo Pi | 3.134 s | 2.648 s | -15.51% | five-run row; target still misses both references |
| I-Before-E | 0.682 s | 0.584 s | -14.37% | 10-run pooled 0.601 s |
| RMS Norm | 5.558 s | 4.764 s | -14.29% | five-run row |
| Array Slice Window | 0.758 s | 0.804 s | +6.07% | contradictory first cohort; expansion 0.704 s, pooled 0.754 s |
| Lexical Rollup | 0.482 s | 0.570 s | +18.26% | expansion 0.514 s, pooled 0.542 s; remains volatile and a clear miss |

The other expanded pooled means are Base64 3.188 s, JSON 0.845 s, Fixed Width
128 8.217 s, Rational Series 4.205 s, Word Frequency 1.605 s, Option/Result
Config 0.883 s, Future Pipeline 0.476 s, Future/Await Race 0.192 s, and the
remaining concurrency rows 0.214-0.556 s. Base64 remains safely inside both
interpreter targets in both independent cohorts; no target classification
flips.

## Remaining bytecode gap

Ranking misses by absolute time above the faster reference avoids letting tiny
reference denominators alone choose the next optimization:

| Application | Able | Faster reference | Absolute excess | Dominant family from current evidence |
| --- | ---: | ---: | ---: | --- |
| K-Nucleotide | 42.986 s | 1.388 s | 41.598 s | map/text/integer and call-return boundaries |
| Fixed Width 128 | 8.266 s | 0.382 s | 7.884 s | checked wide-integer operations |
| Reverse Complement | 7.070 s | 0.027 s | 7.043 s | byte-array/string construction and boxing |
| Distance Field | 6.158 s | 0.345 s | 5.813 s | numeric arrays and float work |
| Mandelbrot | 6.624 s | 1.274 s | 5.350 s | float conditions and numeric loop work |
| Regex Set | 4.334 s | 0.022 s | 4.312 s | canonical regex NFA work |

These are unlike owners below generic dispatcher/allocation parents. Prior
gates have already rejected generic raw-integer extraction, lookup policy,
return/frame ABI, named-container rules, regex state indexing, GC policy, and
benchmark-specific lowering. The scorecard alone therefore does not admit
another implementation candidate.

## Verification

- `bench_scorecard_evidence_check.py` accepts 62 selected rows, 70 full-status
  rows, and exactly five successful Able/reference samples per selected row.
- `bench_external_scoreboard --check` passes after promotion.
- Selection-manifest validation and all seven scorecard-selection protocol
  tests pass.
- All promoted and companion JSON artifacts parse successfully.
- Temporary benchmark workspaces were removed by the harness; no WASM work was
  performed.

## Next recommendation

Refresh bounded main-only CPU profiles for K-Nucleotide, Fixed Width 128,
Reverse Complement, Distance Field, and Mandelbrot before editing another VM
helper. They are the five largest absolute selected-bytecode misses and span
unlike map/text, wide-integer, byte/string, array/float, and numeric-loop
families.

Run one verifier-backed profile process per application under the same
55-second and memory guardrails, reconcile exact descendants rather than
aggregate Go runtime parents, and add allocation profiles only where CPU
evidence identifies allocation as the removable wall. Admit a prototype only
if the same concrete validation, materialization, or dispatch boundary is
material in at least three of the five. Do not reopen a previously rejected
carrier, frame, lookup, GC, named-container, or workload-specific rule without
new invalidating evidence. This is the shortest evidence path to another broad
VM improvement while continuing to defer WASM.
