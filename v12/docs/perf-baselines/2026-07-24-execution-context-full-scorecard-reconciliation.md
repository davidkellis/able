# Compiled execution-context full-scorecard reconciliation — 2026-07-24

## Decision

Keep the generated execution-context ABI behind
`--experimental-execution-context`. Do not make it the compiler default and do
not propagate it through the generated callable ABI yet.

The candidate is broadly promising, but it is not broadly safe for
performance. Across the 61 portable application benchmarks it reduces the
unweighted geometric-mean runtime by 7.39% and improves 38 rows by more than
2%. It also produces repeat-confirmed material regressions in unrelated
numeric, text, and concurrency programs. A global default would therefore
violate the project's rule that compiler optimizations must help general
programs rather than trade selected benchmark wins for real-world losses.

The retained native-interface implementation remains useful experimental
infrastructure. The default compiler path is unchanged when the option is
absent.

## Protocol

All cohorts used the same fresh Go reference, source-equivalent verifiers,
canonical `able-stdlib`, CPU budgets from the benchmark catalog, CPU pool
`0-3`, `GOMEMLIMIT=1GiB`, `GOGC=50`, five independent processes per row, a
60-second runtime limit, and a separate 300-second build-only limit.

The complete measurements were:

- 61 Go reference rows, 305/305 verified processes;
- 61 default-ABI compiled rows, 303/305 verified processes and two timeouts;
- 61 experimental-ABI compiled rows, 305/305 verified processes;
- ten reverse-order experimental repeats, 50/50 verified processes;
- twelve reverse-order default repeats, 60/60 verified processes.

The two default first-cohort timeouts were one run each in
`future_pipeline` and `concurrent_stencil_reduction`. Their reverse-order
supplements completed 5/5 at 0.3200 seconds and 0.2200 seconds. Those complete
means replace the incomplete first-cohort means only in the suite aggregate.
The original reports remain unchanged so the timeout evidence is preserved.

## Verification

The performance-evidence ledger tests, scorecard-selection tests, bounded
scorecard-refresh tests, and checked ledger all pass. The final
`GOMEMLIMIT=1GiB GOGC=50 ./run_all_tests.sh` gate also passes every governance
check, Go package, compiler batch, and bytecode fixture.

Several existing aggregate partitions exceeded the project's preferred
one-minute duration under workstation load: compiler batches peaked at
285.319 seconds and the bytecode fixture aggregate took 266.160 seconds. They
remained CPU-active and completed successfully. This is test-runner
partitioning/resource evidence, not a benchmark result or a correctness
failure; preserve it for a future infrastructure tranche.

## Full-suite result

Using the complete default supplements for the two timed-out rows:

| Measure | Result |
| --- | ---: |
| Applications | 61 |
| Candidate wins greater than 2% | 38 |
| Within plus or minus 2% | 6 |
| Candidate losses greater than 2% | 17 |
| Candidate wins of at least 10% | 24 |
| Candidate losses of at least 10% | 9 |
| Unweighted geometric-mean candidate/default ratio | 0.9261 |
| Summed default measured time | 58.742 seconds |
| Summed candidate measured time | 58.070 seconds |
| Summed-time candidate/default ratio | 0.9886 |
| Default rows at or within 1.05x Go | 5/61 |
| Candidate rows at or within 1.05x Go | 5/61 |

The geometric mean shows a real broad opportunity, especially in interface-
and concurrency-heavy applications. The nearly neutral summed time and
unchanged Go-target count show that the candidate does not move the product
goal enough to justify its regressions.

Representative first-cohort wins include:

- `future_pipeline`: 0.3200 to 0.1520 seconds, 52.5% faster;
- `concurrent_state_machines`: 0.3620 to 0.2300 seconds, 36.5% faster;
- `concurrent_text_index`: 0.9840 to 0.6460 seconds, 34.3% faster;
- `channel_rollup`: 0.6100 to 0.4060 seconds, 33.4% faster;
- `concurrent_signal_dispatch`: 0.2860 to 0.1920 seconds, 32.9% faster;
- `concurrent_scene_tiles`: 0.4460 to 0.3400 seconds, 23.8% faster;
- `concurrent_audio_voices`: 0.9640 to 0.8100 seconds, 16.0% faster;
- `concurrent_packet_codecs`: 0.5480 to 0.4620 seconds, 15.7% faster.

The earlier K-Nucleotide and N-body guard failures do not recur:
`k_nucleotide` is within 1.0% and `nbody` is within 2.7% in the matched full
cohort.

## Reverse-order regression gate

Ten suspected losses were repeated in reverse order. Each pooled mean below
contains ten default and ten candidate processes:

| Benchmark | Default | Candidate | Change | Default GC | Candidate GC |
| --- | ---: | ---: | ---: | ---: | ---: |
| `wide_integer_records` | 0.1970 s | 0.3220 s | +63.5% | 3.4 | 7.0 |
| `fixed_width_128` | 0.2270 s | 0.3420 s | +50.7% | 5.1 | 8.6 |
| `unicode_scalar_pipeline` | 0.2880 s | 0.3660 s | +27.1% | 6.1 | 8.1 |
| `rational_series` | 0.1440 s | 0.1740 s | +20.8% | 3.3 | 4.0 |
| `mutex_work_queue` | 1.6560 s | 1.8760 s | +13.3% | 7.6 | 7.9 |
| `concurrent_event_routing` | 2.7470 s | 3.0570 s | +11.3% | 10.8 | 10.9 |
| `distance_field` | 0.1100 s | 0.1170 s | +6.4% | 3.0 | 3.7 |
| `matrixmultiply` | 1.1810 s | 1.2130 s | +2.7% | 4.6 | 4.8 |
| `base64` | 2.4790 s | 2.5440 s | +2.6% | 22.6 | 22.6 |
| `manifest_normalization` | 0.2040 s | 0.2030 s | -0.5% | 6.2 | 6.2 |

The first six rows are disqualifying. They span multiple primitive numeric
representations, Unicode processing, and two distinct concurrent scheduling
shapes. The large GC increases in the numeric/text rows indicate a general
context-carrier or context-threaded call-graph allocation cost rather than
workstation timing noise.

## Scorecard and closure handling

The experimental report is not a production scorecard source because the
option is not the compiler default. The checked external scorecard is not
promoted from this tranche: the candidate was rejected, while the default
first cohort intentionally retains its two timeout samples. The complete
default supplement proves that both affected rows remain executable and
verifier-correct.

The compiler-production closure drift is nevertheless reconciled. The changed
production files add option-gated generated code; they do not change generated
output under the default option. The full default cohort covers every portable
compiled application, the reverse supplement completes the two initially
volatile rows, and the prior focused and full correctness gates cover both
option states. Those facts support an atomic shared-scope rebaseline of the
ledger. The tool correctly refuses per-closure advancement when an existing
shared scope has drifted, because that could validate unrelated closures
accidentally. The pre-reconciliation report is therefore preserved, then the
reviewed ledger is bootstrapped once: all 21 closures are current and zero are
invalidated, without changing scorecard selection or any architecture
decision.

## Next selection

Profile the confirmed `wide_integer_records`, `fixed_width_128`,
`rational_series`, and `unicode_scalar_pipeline` regressions under both ABIs,
with one favorable interface-heavy control. Attribute generated call sites,
escape analysis, allocations, and GC to the exact context carrier and wrapper
boundary before building another candidate.

This is next because those four rows give the clearest repeated signal and
share elevated GC despite exercising unlike language features. The work should
seek a general semantic effect/escape rule that omits context transport only
from generated call graphs proven unable to observe package, dynamic, native,
or concurrency context. It must not select benchmarks, named containers, or
nominal types. A candidate is admissible only if it preserves the current
interface-heavy wins while clearing all four regression guards and the two
concurrency guards. Until that proof exists, do not extend the ABI to all
callables.

## Primary evidence

- `2026-07-24-execution-context-go-reference.json`
- `2026-07-24-execution-context-default-compiled.json`
- `2026-07-24-execution-context-candidate-compiled.json`
- `2026-07-24-execution-context-default-regression-repeat.json`
- `2026-07-24-execution-context-candidate-regression-repeat.json`
- `2026-07-24-execution-context-pre-reconciliation-invalidation.json`
- `2026-07-24-compiled-native-interface-execution-context.md`
- `2026-07-21-performance-evidence-invalidation-ledger.json`
