# External Application Scoreboard

- Source measurements through: `2026-07-20T18:28:52.284241Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `5/39` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/32` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `78` row fingerprints in JSON; `78` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `78` row fingerprints in JSON; `78` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `71` reviewed benchmark/mode rows, SHA-256 `52ac0405ac8c22824ed3ee0a442247eaf37c2ab7299087f296ecd3b151c80f48`; timeout rows remain in full status.
- Matched reference source fingerprints: `116` comparison fingerprints in JSON; `116` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.4860 | 3.0507 / 1.14x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.3800 | 10.0992 / 0.93x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0600 | 0.9363 / 1.13x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.6800 | 2.4184 / 0.69x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.7180 | 0.5599 / 3.07x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1000 | 0.0608 / 1.64x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.1520 | 2.4440 / 0.88x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.6460 | 1.4782 / 0.44x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1720 | 0.1980 / 0.87x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.1920 | 1.1138 / 1.07x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1280 | 0.0487 / 2.63x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.0900 | 0.0195 / 4.62x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.0060 | 0.0699 / 43.00x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.0900 | 0.0143 / 6.29x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1420 | 0.0347 / 4.09x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.4060 | 1.8333 / 1.86x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1020 | 0.0119 / 8.57x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0900 | 0.0107 / 8.41x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1500 | n/a | 54.0722 / 0.00x | 44.6619 / 0.00x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 55.8364 / n/a | 54.3417 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.1600 | n/a | 47.3536 / 0.09x | 42.9331 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 23.4669 / n/a | 14.3315 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 16.5050 / n/a | 20.6226 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.4940 | n/a | 0.0781 / 6.33x | 0.1100 / 4.49x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.5640 | n/a | 3.5960 / 0.71x | 2.2845 / 1.12x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.7340 | n/a | 2.4170 / 0.30x | 1.5637 / 0.47x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4700 | n/a | 1.4211 / 1.74x | 1.4794 / 1.67x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3520 | n/a | 3.8446 / 0.61x | 9.5775 / 0.25x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.2380 | n/a | 1.1458 / 5.44x | 1.7585 / 3.55x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.3100 | n/a | 0.0241 / 137.34x | 0.0681 / 48.60x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 40.3900 | n/a | 1.2220 / 33.05x | 1.2127 / 33.31x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.8380 | n/a | 0.1998 / 9.20x | 0.2910 / 6.32x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9849 / n/a | 3.3666 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | 55.0529 / n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.6160 | n/a | 0.5517 / 10.18x | 0.3417 / 16.44x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.4440 | n/a | 0.7745 / 5.74x | 0.5178 / 8.58x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5720 | 0.0054 / 105.93x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3480 | 0.0057 / 61.05x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0800 | 0.0037 / 21.62x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3420 | 0.0046 / 74.35x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.7580 | 0.0044 / 172.27x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7360 | 0.0037 / 198.92x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.6000 | 0.0040 / 400.00x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5280 | n/a | 0.0377 / 14.01x | 0.0482 / 10.95x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4160 | n/a | 0.0558 / 7.46x | 0.0678 / 6.14x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1600 | n/a | 0.0315 / 5.08x | 0.0533 / 3.00x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2020 | n/a | 0.1082 / 1.87x | 0.0900 / 2.24x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3780 | n/a | 0.0297 / 12.73x | 0.0510 / 7.41x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2180 | n/a | 0.0198 / 11.01x | 0.0422 / 5.17x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3280 | n/a | 0.0244 / 13.44x | 0.0480 / 6.83x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2220 | 0.0053 / 41.89x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1320 | 0.0131 / 10.08x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.2060 | 0.0245 / 8.41x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1480 | 0.0051 / 29.02x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0900 | 0.0040 / 22.50x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1000 | 0.0050 / 20.00x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1100 | 0.0044 / 25.00x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1240 | 0.0050 / 24.80x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1100 | 0.0049 / 22.45x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0800 | 0.0050 / 16.00x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0800 | 0.0040 / 20.00x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2700 | 0.0116 / 23.28x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2020 | 0.0040 / 50.50x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2420 | 0.0092 / 26.30x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.2940 | n/a | 0.3580 / 20.37x | 0.6022 / 12.11x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.8500 | n/a | 0.0963 / 39.98x | 0.1297 / 29.68x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.1160 | n/a | 0.0632 / 80.95x | 0.1384 / 36.97x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4080 | n/a | 0.0186 / 75.70x | 0.0488 / 28.85x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.2940 | n/a | 0.0134 / 21.94x | 0.0407 / 7.22x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4040 | n/a | 0.0166 / 24.34x | 0.0445 / 9.08x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.1860 | n/a | 0.0166 / 191.93x | 0.0405 / 78.67x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 3.8660 | n/a | 0.0198 / 195.25x | 0.0444 / 87.07x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.3380 | n/a | 0.0185 / 180.43x | 0.0419 / 79.67x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6360 | n/a | 0.0283 / 22.47x | 0.0604 / 10.53x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4620 | n/a | 0.0168 / 27.50x | 0.0451 / 10.24x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.3940 | n/a | 0.0650 / 36.83x | 0.0789 / 30.34x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.7580 | n/a | 0.0158 / 47.97x | 0.0424 / 17.88x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.3120 | n/a | 0.2183 / 15.17x | 0.3224 / 10.27x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-01-selected.json` — `custom` (`2026-07-20T17:42:50.580511Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-02-selected.json` — `custom` (`2026-07-20T17:45:01.387164Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-03-selected.json` — `custom` (`2026-07-20T17:47:42.083920Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-04-selected.json` — `custom` (`2026-07-20T17:49:49.246020Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-05-selected.json` — `custom` (`2026-07-20T17:52:57.685970Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-06-selected.json` — `custom` (`2026-07-20T17:54:39.363279Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-compiled-07-selected.json` — `custom` (`2026-07-20T17:55:34.688661Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-01-status.json` — `custom` (`2026-07-20T17:56:43.981351Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-02-status.json` — `custom` (`2026-07-20T17:58:45.856295Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T17:59:21.561478Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T18:00:24.063025Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T18:04:18.719393Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-06-status.json` — `custom` (`2026-07-20T18:06:20.773368Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T18:07:15.985703Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-async-compiled-01-selected.json` — `custom` (`2026-07-20T18:08:28.899990Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-async-compiled-02-selected.json` — `custom` (`2026-07-20T18:09:40.758437Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-async-bytecode-01-selected.json` — `custom` (`2026-07-20T18:09:53.320571Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-async-bytecode-02-selected.json` — `custom` (`2026-07-20T18:10:07.871772Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T18:13:15.120124Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T18:18:17.717555Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T18:23:00.768701Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T18:25:20.564415Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T18:26:57.609039Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T18:27:23.429033Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T18:28:09.012949Z`)
- `v12/docs/perf-baselines/2026-07-20-wide-integer-records-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T18:28:52.284241Z`)

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
