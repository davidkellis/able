# Current bytecode-versus-interpreter scorecard reconciliation

## Purpose

Refresh every selected bytecode application against same-run Python and Ruby
references after the retained VM work, then reconcile the full current
cross-mode target and ownership frontier. The bytecode product target requires
Able to be no slower than `1 / 0.95 = 1.052632x` both equivalent interpreters.

## Measurement contract

- Selection: 42 reviewed bytecode rows receive performance evidence; seven
  excluded rows receive status-only probes.
- Ranked repetition: five independent verifier-backed bytecode processes and
  five independent verifier-backed Python and Ruby processes per selected row.
- Ranked result: 210/210 Able and 420/420 reference executions pass with zero
  failures or timeouts.
- Status result: the seven Able probes produce two verified completions and
  five bounded timeouts; the fourteen reference probes produce eight verified
  completions and six bounded timeouts. There are no failures.
- Resource contract: bytecode, Python, and Ruby use one logical CPU per catalog
  row. Concurrency rows retain the goroutine/external-concurrency executor
  policy while remaining single-CPU interpreter comparisons.
- Cap: each process is limited to 55 seconds. No quiet-CPU prerequisite is
  imposed; all ranked values are arithmetic means of five workstation samples.
- Runtime source: canonical `../able-stdlib/src`, 70 Able files, source-tree
  SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Evidence:
  [selected bytecode samples](2026-07-22-current-bytecode-scorecard.json),
  [selected references](2026-07-22-current-bytecode-interpreter-reference.json),
  [status bytecode samples](2026-07-22-current-bytecode-status-scorecard.json),
  and
  [status references](2026-07-22-current-bytecode-interpreter-status-reference.json).

## Result

Three of 42 selected bytecode applications meet the snapshot target against
both interpreters: Await Channel Mux, JSON, and PiDigits. Thirty-nine miss.
Summing one mean per selected application gives 147.786 seconds Able versus
17.2168 seconds for the faster applicable reference in each row, or 8.584x.
Total wall time above the per-row target budgets is 132.965684 seconds. As with
the compiled scorecard, that aggregate is a coverage diagnostic rather than a
representative application mix.

JSON and PiDigits remain established cross-cohort guards. Await Channel Mux is
not established: the preceding cohort was 1.726x the limiting Ruby reference,
this cohort is 0.932x, and their pooled limiting ratio is 1.218x. It is
therefore a variance-sensitive snapshot meet. Base64 now clearly misses the
limiting Ruby target at 1.113x, so its obsolete snapshot-meet stability entry
was retired.

## Complete selected bytecode ledger

| Application | Able (s) | Python (s) | Ruby (s) | Worst ratio | Target | Excess (s) |
| --- | ---: | ---: | ---: | ---: | --- | ---: |
| `array_slice_window` | 0.7440 | 0.0610 | 0.1294 | 12.197x | miss | 0.6798 |
| `await_channel_mux` | 0.2000 | 0.2202 | 0.2145 | 0.932x | meet | 0.0000 |
| `base64` | 2.8400 | 6.6027 | 2.5509 | 1.113x | miss | 0.1548 |
| `binary_event_log` | 7.0260 | 0.1779 | 0.2649 | 39.494x | miss | 6.8387 |
| `channel_rollup` | 0.5380 | 0.0471 | 0.0574 | 11.423x | miss | 0.4884 |
| `config_validation_extraction` | 1.3800 | 0.0218 | 0.0489 | 63.303x | miss | 1.3571 |
| `concurrent_document_pipeline` | 0.2900 | 0.0229 | 0.0512 | 12.664x | miss | 0.2659 |
| `concurrent_event_routing` | 3.2560 | 0.0336 | 0.0611 | 96.905x | miss | 3.2206 |
| `concurrent_text_index` | 0.7680 | 0.0973 | 0.1070 | 7.893x | miss | 0.6656 |
| `dependency_plan` | 0.5560 | 0.0193 | 0.0542 | 28.808x | miss | 0.5357 |
| `dependency_wave_validation` | 0.5620 | 0.0353 | 0.0551 | 15.921x | miss | 0.5248 |
| `distance_field` | 5.9140 | 0.5798 | 0.3877 | 15.254x | miss | 5.5059 |
| `document_audit` | 0.2960 | 0.0149 | 0.0428 | 19.866x | miss | 0.2803 |
| `fasta_generation` | 1.9000 | 0.2082 | 0.3160 | 9.126x | miss | 1.6808 |
| `fixed_width_128` | 8.5220 | 0.3504 | 0.6762 | 24.321x | miss | 8.1532 |
| `future_await_race` | 0.1520 | 0.0332 | 0.0598 | 4.578x | miss | 0.1171 |
| `future_pipeline` | 0.4580 | 0.0626 | 0.0709 | 7.316x | miss | 0.3921 |
| `i_before_e` | 0.5340 | 0.0841 | 0.1269 | 6.350x | miss | 0.4455 |
| `inventory_reconciliation` | 2.6240 | 0.0750 | 0.0896 | 34.987x | miss | 2.5451 |
| `json` | 0.8940 | 2.8646 | 1.9273 | 0.464x | meet | 0.0000 |
| `k_nucleotide` | 46.5300 | 1.4742 | 1.4639 | 31.785x | miss | 44.9891 |
| `lexical_rollup` | 0.4480 | 0.0269 | 0.0606 | 16.654x | miss | 0.4197 |
| `log_routing_redaction` | 3.1280 | 0.0219 | 0.0574 | 142.831x | miss | 3.1049 |
| `manifest_normalization` | 1.5820 | 0.0249 | 0.0781 | 63.534x | miss | 1.5558 |
| `mandelbrot` | 6.8140 | 1.4879 | 2.0400 | 4.580x | miss | 5.2478 |
| `monte_carlo_pi` | 2.6440 | 1.7064 | 1.8248 | 1.549x | miss | 0.8478 |
| `mutex_await_journal` | 0.1960 | 0.0302 | 0.0513 | 6.490x | miss | 0.1642 |
| `mutex_ledger` | 0.3760 | 0.0345 | 0.0567 | 10.899x | miss | 0.3397 |
| `mutex_work_queue` | 0.3400 | 0.0290 | 0.0534 | 11.724x | miss | 0.3095 |
| `option_result_config` | 0.8340 | 0.0191 | 0.0538 | 43.665x | miss | 0.8139 |
| `pidigits` | 2.4540 | 4.3663 | 12.5234 | 0.562x | meet | 0.0000 |
| `policy_record_dispatch` | 7.4520 | 0.0318 | 0.0675 | 234.340x | miss | 7.4185 |
| `rational_series` | 4.3260 | 0.1895 | 0.2003 | 22.828x | miss | 4.1265 |
| `regex_set_audit` | 4.5180 | 0.0258 | 0.0766 | 175.116x | miss | 4.4908 |
| `regex_stream_audit` | 3.5960 | 0.0292 | 0.0719 | 123.151x | miss | 3.5653 |
| `regex_suffix_audit` | 3.5520 | 0.0319 | 0.0744 | 111.348x | miss | 3.5184 |
| `reverse_complement` | 3.7580 | 0.0379 | 0.1142 | 99.156x | miss | 3.7181 |
| `rms_norm` | 4.7540 | 1.1947 | 0.7739 | 6.143x | miss | 3.9394 |
| `unicode_scalar_pipeline` | 3.6500 | 0.3329 | 0.3499 | 10.964x | miss | 3.2996 |
| `validated_job_pipeline` | 0.3800 | 0.0263 | 0.0479 | 14.449x | miss | 0.3523 |
| `wide_integer_records` | 5.5860 | 0.0782 | 0.1611 | 71.432x | miss | 5.5037 |
| `word_frequency` | 1.4140 | 0.0234 | 0.0517 | 60.427x | miss | 1.3894 |

## Status-only rows

The excluded rows remain unranked. Fib and Matrix Multiply complete their
single Able probe; Binary Trees, QuickSort, Sudoku Masks, N-Body, and TapeLang
hit the 55-second bytecode cap. Reference outcomes are likewise retained as
status rather than variance evidence. These rows do not enter target counts,
averages, ownership selection, or candidate gates.

## Ownership reconciliation

The fresh target ordering does not open a generic VM candidate. The same-day
cross-feature CPU and allocation-owner matrices already cover text/file,
numeric, collection/iterator, nominal/union, and concurrency programs on the
retained VM. They find no exact non-dispatch CPU leaf or removable semantic
allocation in three unlike applications. Current target-excess groups remain
closed because their descendants split by semantics, lack sufficient breadth,
or correspond to candidates already rejected by broad averaged guards.

No additional profile was run: the admission rule requires a materially
missing group with one open concrete VM descendant in at least three unlike
programs, and the reconciled evidence has none. This tranche retains no VM,
compiler, runtime, stdlib, benchmark, language, or WASM change.

## Recommendation

Build a current cross-engine architecture target-budget reconciliation next.
Both complete scorecards are now current, but every known local/shared leaf is
closed while 44 compiled and 39 selected bytecode rows still miss. Join the
fresh per-row target excess with existing exact owner evidence, separate fixed
process/runtime cost from sustained application work, and calculate the
maximum plausible savings of each remaining generic architecture boundary
before implementing another candidate. The audit should select only a
language-general mechanism with enough theoretical budget and measured reach
in at least three unlike programs; otherwise it should explicitly identify
which larger compiler or VM boundary must change. This avoids retrying local
optimizations that cannot close the measured gaps and continues to exclude
named-container, application-specific, benchmark-specific, and WASM paths.
