# External Application Scoreboard

- Source measurements through: `2026-07-16T06:54:18.046118Z`
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
| `fib` | `compiled` | `verified` | 3.4400 | 3.0684 / 1.12x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 29.6520 | 5.8238 / 5.09x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1320 | 0.9564 / 1.18x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.8140 | 2.5033 / 0.72x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 8.5240 | 0.5722 / 14.90x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1020 | 0.0584 / 1.75x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.3800 | 2.4569 / 0.97x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7240 | 1.4255 / 0.51x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2020 | 0.2185 / 0.92x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3320 | 1.1753 / 1.13x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1280 | 0.0486 / 2.63x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1200 | 0.0147 / 8.16x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 4.5180 | 0.0568 / 79.54x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.4560 | 0.0326 / 13.99x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.4720 | 1.9528 / 1.78x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1540 | n/a | 57.5715 / 0.00x | 45.5928 / 0.00x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 12.7937 / n/a | 54.8039 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.7120 | n/a | 49.2453 / 0.10x | 51.8666 / 0.09x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.3114 / n/a | 14.7283 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.1785 / n/a | 21.7267 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5460 | n/a | 0.0906 / 6.03x | 0.1171 / 4.66x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9200 | n/a | 3.8149 / 0.77x | 2.3754 / 1.23x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.7840 | n/a | 2.5440 / 0.31x | 1.6323 / 0.48x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.6780 | n/a | 1.4469 / 1.85x | 1.5530 / 1.72x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3260 | n/a | 3.8748 / 0.60x | 9.7157 / 0.24x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.6880 | n/a | 1.1832 / 5.65x | 1.8865 / 3.55x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 7.0380 | n/a | 0.0250 / 281.52x | 0.0707 / 99.55x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 41.5300 | n/a | 1.3074 / 31.77x | 1.2687 / 32.73x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0049 / n/a | 3.0318 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | 57.2333 / n/a | 73.7805 / n/a | `unranked` | Able timed out |
| `channel_rollup` | `compiled` | `verified` | 1.2320 | 0.0045 / 273.78x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.6840 | 0.0042 / 162.86x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1240 | 0.0032 / 38.75x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3620 | 0.0040 / 90.50x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.5240 | 0.0038 / 137.89x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.4440 | 0.0033 / 134.55x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.6340 | n/a | 0.0389 / 16.30x | 0.0508 / 12.48x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4900 | n/a | 0.0613 / 7.99x | 0.0725 / 6.76x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1600 | n/a | 0.0309 / 5.18x | 0.0528 / 3.03x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2640 | n/a | 0.1235 / 2.14x | 0.0979 / 2.70x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.7180 | n/a | 0.0317 / 22.65x | 0.0519 / 13.83x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2540 | n/a | 0.0195 / 13.03x | 0.0468 / 5.43x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 8.4320 | 0.0050 / 1686.40x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 2.3940 | 0.0122 / 196.23x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2620 | 0.0046 / 56.96x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0900 | 0.0034 / 26.47x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1160 | 0.0035 / 33.14x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 2.7440 | 0.0322 / 85.22x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.2080 | 0.0041 / 50.73x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1980 | 0.0041 / 48.29x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0920 | 0.0037 / 24.86x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0900 | 0.0032 / 28.12x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2100 | 0.0030 / 70.00x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.8080 | n/a | 0.3428 / 22.78x | 0.6456 / 12.09x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0620 | n/a | 0.0987 / 41.16x | 0.1317 / 30.84x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4840 | n/a | 0.0187 / 79.36x | 0.0512 / 28.98x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3300 | n/a | 0.0131 / 25.19x | 0.0436 / 7.57x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.5100 | n/a | 0.0163 / 31.29x | 0.0505 / 10.10x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0396 / n/a | 0.0748 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 5.3800 | n/a | 0.0181 / 297.24x | 0.0416 / 129.33x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.5680 | n/a | 0.0181 / 252.38x | 0.0427 / 106.98x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7380 | n/a | 0.0270 / 27.33x | 0.0606 / 12.18x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5400 | n/a | 0.0156 / 34.62x | 0.0473 / 11.42x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 3.6600 | n/a | 0.0164 / 223.17x | 0.0423 / 86.52x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-01.json` — `custom` (`2026-07-16T05:27:23.942229Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-02.json` — `custom` (`2026-07-16T05:30:14.876612Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-03.json` — `custom` (`2026-07-16T05:33:06.519436Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-04.json` — `custom` (`2026-07-16T05:35:22.989506Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-05.json` — `custom` (`2026-07-16T05:37:50.439848Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-compiled-06.json` — `custom` (`2026-07-16T05:39:59.030422Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-01.json` — `custom` (`2026-07-16T05:48:00.217787Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-02.json` — `custom` (`2026-07-16T06:03:04.811877Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-03.json` — `custom` (`2026-07-16T06:03:44.245277Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-04.json` — `custom` (`2026-07-16T06:04:49.824325Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-05.json` — `custom` (`2026-07-16T06:08:57.258089Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-generality-bytecode-06.json` — `custom` (`2026-07-16T06:24:01.276318Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-compiled-01.json` — `custom` (`2026-07-16T06:26:13.240619Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-compiled-02.json` — `custom` (`2026-07-16T06:27:36.914970Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-bytecode-01.json` — `custom` (`2026-07-16T06:27:49.903705Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-async-bytecode-02.json` — `custom` (`2026-07-16T06:28:02.653579Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-compiled-01.json` — `custom` (`2026-07-16T06:31:21.054128Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-compiled-02.json` — `custom` (`2026-07-16T06:37:10.421365Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-compiled-03.json` — `custom` (`2026-07-16T06:42:57.744615Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-compiled-04.json` — `custom` (`2026-07-16T06:43:58.168687Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-bytecode-01.json` — `custom` (`2026-07-16T06:45:11.507236Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-bytecode-02.json` — `custom` (`2026-07-16T06:52:52.331210Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-bytecode-03.json` — `custom` (`2026-07-16T06:53:52.536277Z`)
- `v12/docs/perf-baselines/2026-07-15-strict-cohort-b-coverage-extra-bytecode-04.json` — `custom` (`2026-07-16T06:54:18.046118Z`)

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
