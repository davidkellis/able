# External Application Scoreboard

- Source measurements through: `2026-07-16T04:36:44.827042Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/32` rankable rows meet the 95%-of-Go target.
- Bytecode: `4/26` rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `64` row fingerprints in JSON; `64` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `64` row fingerprints in JSON; `64` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `69` `.able` files, tree SHA-256 `785a6fd058c179379b1a153529fb340151a11b96d9014394cc40dbd87e1882ab`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `58` reviewed benchmark/mode rows, SHA-256 `11caaee63c66fa2e235249640fe0ce44833dc6ed9946d7d0a1e840997345c132`; timeout rows remain in full status.
- Matched reference source fingerprints: `96` comparison fingerprints in JSON; `96` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.2440 | 2.9049 / 1.12x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 28.6320 | 4.9766 / 5.75x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0300 | 0.8885 / 1.16x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.7100 | 2.3142 / 0.74x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 8.1900 | 0.5154 / 15.89x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1000 | 0.0549 / 1.82x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.2160 | 2.1890 / 1.01x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.6760 | 1.2785 / 0.53x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1980 | 0.2022 / 0.98x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.2360 | 1.0872 / 1.14x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1380 | 0.0458 / 3.01x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1180 | 0.0147 / 8.03x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.3220 | 0.0541 / 61.40x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.3900 | 0.0312 / 12.50x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.3520 | 1.8119 / 1.85x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1460 | n/a | 55.3957 / 0.00x | 41.6611 / 0.00x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 10.9425 / n/a | 50.2250 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.3380 | n/a | 46.8189 / 0.09x | 43.8530 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 22.7769 / n/a | 14.0915 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 15.9696 / n/a | 21.8573 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5300 | n/a | 0.0772 / 6.87x | 0.1064 / 4.98x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9300 | n/a | 3.5481 / 0.83x | 2.2160 / 1.32x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8680 | n/a | 2.3799 / 0.36x | 1.5424 / 0.56x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.3060 | n/a | 2.1822 / 1.06x | 2.2658 / 1.02x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3260 | n/a | 5.8706 / 0.40x | 13.2183 / 0.18x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.1180 | n/a | 1.2988 / 4.71x | 2.3284 / 2.63x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 6.5600 | n/a | 0.0293 / 223.89x | 0.0781 / 83.99x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 40.0240 | n/a | 1.3912 / 28.77x | 1.3166 / 30.40x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9283 / n/a | 2.9236 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | 61.3975 / n/a | 69.4130 / n/a | `unranked` | Able timed out |
| `channel_rollup` | `compiled` | `verified` | 1.2420 | 0.0045 / 276.00x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.6580 | 0.0042 / 156.67x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0031 / 38.71x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3540 | 0.0039 / 90.77x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.5120 | 0.0035 / 146.29x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.4120 | 0.0031 / 132.90x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.6160 | n/a | 0.0362 / 17.02x | 0.0455 / 13.54x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4900 | n/a | 0.0578 / 8.48x | 0.0697 / 7.03x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1580 | n/a | 0.0288 / 5.49x | 0.0479 / 3.30x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2500 | n/a | 0.1157 / 2.16x | 0.0899 / 2.78x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.7140 | n/a | 0.0289 / 24.71x | 0.0461 / 15.49x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2520 | n/a | 0.0183 / 13.77x | 0.0391 / 6.45x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 8.0980 | 0.0048 / 1687.08x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 2.4000 | 0.0118 / 203.39x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2440 | 0.0045 / 54.22x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1060 | 0.0032 / 33.12x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1240 | 0.0034 / 36.47x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 2.7320 | 0.0303 / 90.17x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.2340 | 0.0041 / 57.07x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.3900 | 0.0038 / 102.63x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1500 | 0.0035 / 42.86x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1800 | 0.0030 / 60.00x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2440 | 0.0030 / 81.33x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.1580 | n/a | 0.3117 / 26.17x | 0.5768 / 14.14x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.1040 | n/a | 0.0918 / 44.71x | 0.1194 / 34.37x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5340 | n/a | 0.0174 / 88.16x | 0.0449 / 34.16x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3380 | n/a | 0.0120 / 28.17x | 0.0366 / 9.23x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.5180 | n/a | 0.0145 / 35.72x | 0.0418 / 12.39x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0342 / n/a | 0.0662 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 5.9700 | n/a | 0.0167 / 357.49x | 0.0386 / 154.66x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.9220 | n/a | 0.0164 / 300.12x | 0.0380 / 129.53x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.8100 | n/a | 0.0249 / 32.53x | 0.0535 / 15.14x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5340 | n/a | 0.0142 / 37.61x | 0.0412 / 12.96x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 3.8380 | n/a | 0.0157 / 244.46x | 0.0416 / 92.26x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-01.json` — `custom` (`2026-07-16T03:09:38.729157Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-02.json` — `custom` (`2026-07-16T03:12:23.217176Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-03.json` — `custom` (`2026-07-16T03:15:01.068122Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-04.json` — `custom` (`2026-07-16T03:17:07.850846Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-05.json` — `custom` (`2026-07-16T03:19:24.650039Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-compiled-06.json` — `custom` (`2026-07-16T03:21:06.933308Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-01.json` — `custom` (`2026-07-16T03:29:05.713237Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-02.json` — `custom` (`2026-07-16T03:44:10.030441Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-03.json` — `custom` (`2026-07-16T03:44:48.808278Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-04.json` — `custom` (`2026-07-16T03:45:48.696312Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-05.json` — `custom` (`2026-07-16T03:49:45.887535Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-generality-bytecode-06.json` — `custom` (`2026-07-16T04:04:49.596777Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-async-compiled-01.json` — `custom` (`2026-07-16T04:06:54.051778Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-async-compiled-02.json` — `custom` (`2026-07-16T04:08:14.514486Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-async-bytecode-01.json` — `custom` (`2026-07-16T04:08:27.258557Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-async-bytecode-02.json` — `custom` (`2026-07-16T04:08:39.860728Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-compiled-01.json` — `custom` (`2026-07-16T04:11:41.217863Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-compiled-02.json` — `custom` (`2026-07-16T04:17:30.394180Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-compiled-03.json` — `custom` (`2026-07-16T04:24:44.676705Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-compiled-04.json` — `custom` (`2026-07-16T04:26:15.263002Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-bytecode-01.json` — `custom` (`2026-07-16T04:27:31.338568Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-bytecode-02.json` — `custom` (`2026-07-16T04:35:12.369962Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-bytecode-03.json` — `custom` (`2026-07-16T04:36:18.220587Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-a-coverage-extra-bytecode-04.json` — `custom` (`2026-07-16T04:36:44.827042Z`)

Regenerate after a new verifier-backed source scorecard with:

```sh
just bench-scoreboard
```

To replace the selected sources, pass each new scorecard explicitly, for example
`just bench-scoreboard --input path/to/compiled.json --input path/to/bytecode.json`.

Validate the checked-in report without running performance workloads with:

```sh
just bench-scoreboard-check
```
