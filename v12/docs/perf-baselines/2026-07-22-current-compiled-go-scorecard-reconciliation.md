# Current compiled-versus-Go scorecard reconciliation

## Purpose

Refresh every portable compiled application against a same-run Go reference,
then reconcile the target and ownership frontier after the retained
direct-known-method and split-receiver ABI changes. The product target is Able
wall time no greater than `1 / 0.95 = 1.052632x` the equivalent Go program.

## Measurement contract

- Coverage: all 49 portable applications and all 49 selected compiled rows.
- Repetition: five independent verifier-backed processes per Able row and five
  independent verifier-backed processes per fresh Go row.
- Process result: 245/245 Able and 245/245 Go executions pass; zero failures
  and zero timeouts.
- Resource contract: 36 serial applications use one logical CPU; 13 explicitly
  parallel applications use four logical CPUs and the goroutine executor.
  Both implementations use the same resolved `0-3` CPU pool per row.
- Cap: each timed process has a 55-second timeout. Build time is outside the
  timed application process.
- Workstation policy: no quiet-CPU prerequisite. The retained result is the
  arithmetic mean of repeated processes, as requested for a normally loaded
  workstation.
- Runtime source: canonical `../able-stdlib/src`, 70 Able files, source-tree
  SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Evidence:
  [compiled samples](2026-07-22-current-compiled-scorecard.json),
  [Go samples](2026-07-22-current-compiled-go-reference.json), and
  [stdlib state](2026-07-22-current-compiled-stdlib-source-state.json).

## Result

Five of 49 applications meet the snapshot target: Binary Trees, QuickSort,
Base64, JSON, and Monte Carlo Pi. Forty-four miss. Summing the independently
weighted application means gives 49.526 seconds Able versus 28.1438 seconds Go,
or 1.760x; total wall time above the per-row target budgets is 23.368 seconds.
That aggregate is a coverage diagnostic, not an estimate of a typical
application mix, because each benchmark contributes once and many Go programs
finish in only 4-10 milliseconds.

Three snapshot meets are established cross-cohort guards: Binary Trees,
QuickSort, and JSON. Base64 is a volatile crossing: its immediately preceding
five-process cohort was 1.061x Go, this cohort is 0.922x, and their pooled ratio
is 0.989x. Monte Carlo Pi was already classified as a volatile crossing.
Matrix Multiply now clearly misses at 1.132x, so its obsolete
snapshot-meet-only stability entry was retired. PiDigits is the nearest miss at
1.066x; its five Able samples are 1.25-1.38 seconds and its five Go samples are
1.215-1.266 seconds, so the present miss is narrow but not caused by a single
outlier.

## Complete compiled ledger

| Application | CPUs | Able mean (s) | Go mean (s) | Able/Go | Target | Excess (s) |
| --- | ---: | ---: | ---: | ---: | --- | ---: |
| `fib` | 1 | 4.9340 | 3.4101 | 1.447x | miss | 1.3444 |
| `binarytrees` | 4 | 11.4440 | 11.9385 | 0.959x | meet | 0.0000 |
| `matrixmultiply` | 1 | 1.2140 | 1.0729 | 1.132x | miss | 0.0846 |
| `quicksort` | 1 | 1.9120 | 2.6813 | 0.713x | meet | 0.0000 |
| `sudoku_masks` | 1 | 1.9060 | 0.5782 | 3.296x | miss | 1.2974 |
| `i_before_e` | 1 | 0.1220 | 0.0637 | 1.915x | miss | 0.0549 |
| `base64` | 1 | 2.4480 | 2.6537 | 0.922x | meet | 0.0000 |
| `binary_event_log` | 1 | 0.5360 | 0.0090 | 59.556x | miss | 0.5265 |
| `json` | 1 | 0.7420 | 1.6387 | 0.453x | meet | 0.0000 |
| `monte_carlo_pi` | 1 | 0.1920 | 0.2826 | 0.679x | meet | 0.0000 |
| `pidigits` | 1 | 1.3200 | 1.2384 | 1.066x | miss | 0.0164 |
| `mandelbrot` | 1 | 0.1280 | 0.0527 | 2.429x | miss | 0.0725 |
| `reverse_complement` | 1 | 0.1140 | 0.0171 | 6.667x | miss | 0.0960 |
| `k_nucleotide` | 1 | 2.8980 | 0.0809 | 35.822x | miss | 2.8128 |
| `nbody` | 1 | 0.1720 | 0.0378 | 4.550x | miss | 0.1322 |
| `tapelang_alphabet` | 1 | 4.0460 | 2.1407 | 1.890x | miss | 1.7926 |
| `distance_field` | 1 | 0.0900 | 0.0133 | 6.767x | miss | 0.0760 |
| `rms_norm` | 1 | 0.0880 | 0.0119 | 7.395x | miss | 0.0755 |
| `fasta_generation` | 1 | 0.1100 | 0.0171 | 6.433x | miss | 0.0920 |
| `fixed_width_128` | 1 | 0.2060 | 0.0058 | 35.517x | miss | 0.1999 |
| `rational_series` | 1 | 0.1280 | 0.0145 | 8.828x | miss | 0.1127 |
| `wide_integer_records` | 1 | 0.1840 | 0.0267 | 6.891x | miss | 0.1559 |
| `word_frequency` | 1 | 0.1800 | 0.0059 | 30.508x | miss | 0.1738 |
| `document_audit` | 1 | 0.1020 | 0.0053 | 19.245x | miss | 0.0964 |
| `lexical_rollup` | 1 | 0.1200 | 0.0054 | 22.222x | miss | 0.1143 |
| `channel_rollup` | 4 | 0.5980 | 0.0063 | 94.921x | miss | 0.5914 |
| `future_pipeline` | 4 | 0.3960 | 0.0057 | 69.474x | miss | 0.3900 |
| `future_await_race` | 4 | 0.1060 | 0.0049 | 21.633x | miss | 0.1008 |
| `await_channel_mux` | 4 | 0.3600 | 0.0055 | 65.455x | miss | 0.3542 |
| `mutex_ledger` | 4 | 0.8200 | 0.0051 | 160.784x | miss | 0.8146 |
| `mutex_await_journal` | 4 | 0.8120 | 0.0045 | 180.444x | miss | 0.8073 |
| `mutex_work_queue` | 4 | 2.3300 | 0.0052 | 448.077x | miss | 2.3245 |
| `regex_suffix_audit` | 1 | 0.1680 | 0.0064 | 26.250x | miss | 0.1613 |
| `regex_set_audit` | 1 | 0.1440 | 0.0070 | 20.571x | miss | 0.1366 |
| `regex_stream_audit` | 1 | 0.1600 | 0.0062 | 25.806x | miss | 0.1535 |
| `log_routing_redaction` | 1 | 0.1440 | 0.0058 | 24.828x | miss | 0.1379 |
| `config_validation_extraction` | 1 | 0.1300 | 0.0047 | 27.660x | miss | 0.1251 |
| `unicode_scalar_pipeline` | 1 | 0.3060 | 0.0111 | 27.568x | miss | 0.2943 |
| `array_slice_window` | 1 | 0.0980 | 0.0047 | 20.851x | miss | 0.0931 |
| `dependency_plan` | 1 | 0.0940 | 0.0051 | 18.431x | miss | 0.0886 |
| `inventory_reconciliation` | 1 | 0.2120 | 0.0096 | 22.083x | miss | 0.2019 |
| `option_result_config` | 1 | 0.1760 | 0.0052 | 33.846x | miss | 0.1705 |
| `concurrent_text_index` | 4 | 0.9960 | 0.0064 | 155.625x | miss | 0.9893 |
| `validated_job_pipeline` | 4 | 1.0460 | 0.0038 | 275.263x | miss | 1.0420 |
| `dependency_wave_validation` | 4 | 1.4440 | 0.0043 | 335.814x | miss | 1.4395 |
| `concurrent_event_routing` | 4 | 2.9140 | 0.0056 | 520.357x | miss | 2.9081 |
| `concurrent_document_pipeline` | 4 | 0.2840 | 0.0051 | 55.686x | miss | 0.2786 |
| `manifest_normalization` | 1 | 0.2260 | 0.0047 | 48.085x | miss | 0.2211 |
| `policy_record_dispatch` | 1 | 0.2260 | 0.0087 | 25.977x | miss | 0.2168 |

## Ownership reconciliation

The current evidence manifest now records the completed post-ABI owner refresh.
The former `compiled-iterator-control` candidate is closed: the retained
direct-known-method and split-receiver ABI work removed its shared escaping
bound-method box, and the fresh eight-application CPU/counter/allocation pass
found no successor concrete generated/runtime leaf in three unlike programs.
The remaining compiled groups retain their established dispositions:

- target guards remain protected;
- Sudoku quotient remains exact but only one application wide;
- float, wide-numeric, regex, and concurrency candidates have already failed
  broad unlike-program guards;
- byte-output, text/map, current-control, and post-ABI iterator/control split
  into different concrete descendants.

Consequently this tranche does not profile another subset and retains no
compiler, runtime, stdlib, benchmark, language, or WASM change. Selecting a
candidate from the largest ratios alone would optimize application families or
startup-dominated short programs without evidence of a generally reusable
leaf.

## Recommendation

Refresh the complete selected bytecode-versus-Python/Ruby scorecard next. The
compiled half is now current, but the bytecode half of the canonical frontier
still comes from earlier runtime artifacts. A bytecode refresh will run five
verified processes for each of the 42 selected rows against fresh five-process
Python and Ruby references, retain one bounded status probe for each excluded
row, update target/stability classifications, and profile only a materially
missing group with one open concrete VM descendant in at least three unlike
applications. This is the shortest evidence-driven route to deciding whether
the next broadly applicable performance change belongs in the VM or whether
both engines need a larger architectural step.
