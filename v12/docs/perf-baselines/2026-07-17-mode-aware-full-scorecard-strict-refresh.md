# External Application Scoreboard

- Source measurements through: `2026-07-17T15:13:05.070936Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `3/34` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `3/27` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `68` row fingerprints in JSON; `68` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `68` row fingerprints in JSON; `68` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `69` `.able` files, tree SHA-256 `f37de0ac91abf02ab7c2af47e66cc06c9a37b9e32d618f4b12aee6ff11587f1d`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `61` reviewed benchmark/mode rows, SHA-256 `d829d5ae1a06dd346e1a9b9a0e8f4d33405bc0bca74c630ac858cf3912b35bf5`; timeout rows remain in full status.
- Matched reference source fingerprints: `98` comparison fingerprints in JSON; `98` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.2620 | 2.9887 / 1.09x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.7260 | 10.1090 / 0.96x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.4640 | 1.0017 / 1.46x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.8380 | 2.4316 / 0.76x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 9.2300 | 0.5342 / 17.28x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.2100 | 0.0574 / 3.66x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5400 | 2.3465 / 1.08x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.9860 | 1.3146 / 0.75x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2860 | 0.1982 / 1.44x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.4480 | 1.1046 / 1.31x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1480 | 0.0484 / 3.06x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1320 | 0.0146 / 9.04x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.4860 | 0.0527 / 66.15x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.2640 | 0.0314 / 8.41x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.2520 | 1.8379 / 2.31x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1020 | 0.0162 / 6.30x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1400 | 0.0104 / 13.46x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1840 | n/a | 54.7512 / 0.00x | 44.0366 / 0.00x | `meets` | — |
| `matrixmultiply` | `bytecode` | `verified` | 4.9800 | n/a | 47.8571 / 0.10x | 46.3074 / 0.11x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.4656 / n/a | 15.2190 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 21.5935 / n/a | 22.9846 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6840 | n/a | 0.0788 / 8.68x | 0.1130 / 6.05x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9340 | n/a | 3.7912 / 0.77x | 2.4701 / 1.19x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.9140 | n/a | 2.6341 / 0.35x | 1.8039 / 0.51x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.6040 | n/a | 1.5449 / 1.69x | 1.5596 / 1.67x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.1740 | n/a | 4.2562 / 0.51x | 10.3098 / 0.21x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.0360 | n/a | 1.1789 / 5.12x | 1.8385 / 3.28x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 6.0020 | n/a | 0.0260 / 230.85x | 0.0756 / 79.39x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 38.7740 | n/a | 1.3276 / 29.21x | 1.3517 / 28.69x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0515 / n/a | 3.2673 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.3280 | n/a | 0.5631 / 9.46x | 0.3364 / 15.84x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.4200 | n/a | 0.9028 / 4.90x | 0.5449 / 8.11x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.8060 | 0.0052 / 155.00x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3620 | 0.0051 / 70.98x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0820 | 0.0038 / 21.58x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3140 | 0.0044 / 71.36x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.6540 | 0.0042 / 155.71x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6360 | 0.0038 / 167.37x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.4760 | n/a | 0.0387 / 12.30x | 0.0497 / 9.58x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4140 | n/a | 0.0572 / 7.24x | 0.0691 / 5.99x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1480 | n/a | 0.0353 / 4.19x | 0.0501 / 2.95x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2000 | n/a | 0.1158 / 1.73x | 0.0899 / 2.22x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3480 | n/a | 0.0313 / 11.12x | 0.0502 / 6.93x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2060 | n/a | 0.0193 / 10.67x | 0.0443 / 4.65x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.1980 | 0.0055 / 36.00x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1240 | 0.0131 / 9.47x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2000 | 0.0067 / 29.85x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0940 | 0.0038 / 24.74x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1100 | 0.0043 / 25.58x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 1.3060 | 0.0419 / 31.17x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1480 | 0.0046 / 32.17x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1380 | 0.0045 / 30.67x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0880 | 0.0041 / 21.46x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1120 | 0.0035 / 32.00x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2360 | 0.0038 / 62.11x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.8380 | n/a | 0.3863 / 20.29x | 0.6491 / 12.08x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.1320 | n/a | 0.1183 / 34.93x | 0.1251 / 33.03x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.6580 | n/a | 0.0178 / 93.15x | 0.0481 / 34.47x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3980 | n/a | 0.0129 / 30.85x | 0.0392 / 10.15x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.5100 | n/a | 0.0160 / 31.88x | 0.0528 / 9.66x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0382 / n/a | 0.0714 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 5.7540 | n/a | 0.0195 / 295.08x | 0.0403 / 142.78x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.5880 | n/a | 0.0177 / 259.21x | 0.0410 / 111.90x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6960 | n/a | 0.0271 / 25.68x | 0.0586 / 11.88x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4920 | n/a | 0.0158 / 31.14x | 0.0450 / 10.93x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 1.0400 | n/a | 0.0168 / 61.90x | 0.0426 / 24.41x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-01-selected.json` — `custom` (`2026-07-17T14:27:47.951550Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-02-selected.json` — `custom` (`2026-07-17T14:30:50.926570Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-03-selected.json` — `custom` (`2026-07-17T14:34:25.635827Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-04-selected.json` — `custom` (`2026-07-17T14:37:00.888092Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-05-selected.json` — `custom` (`2026-07-17T14:39:34.494262Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-06-selected.json` — `custom` (`2026-07-17T14:41:31.964295Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-compiled-07-selected.json` — `custom` (`2026-07-17T14:42:31.888788Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-01-selected.json` — `custom` (`2026-07-17T14:43:03.008802Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-01-status.json` — `custom` (`2026-07-17T14:44:00.153459Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-02-status.json` — `custom` (`2026-07-17T14:45:54.362441Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-03-selected.json` — `custom` (`2026-07-17T14:46:35.644982Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-04-selected.json` — `custom` (`2026-07-17T14:47:36.072503Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-05-selected.json` — `custom` (`2026-07-17T14:51:24.331226Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-06-status.json` — `custom` (`2026-07-17T14:53:18.152034Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-generality-bytecode-07-selected.json` — `custom` (`2026-07-17T14:54:11.037708Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-async-compiled-01-selected.json` — `custom` (`2026-07-17T14:55:16.773881Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-async-compiled-02-selected.json` — `custom` (`2026-07-17T14:56:01.025956Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-async-bytecode-01-selected.json` — `custom` (`2026-07-17T14:56:12.416231Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-async-bytecode-02-selected.json` — `custom` (`2026-07-17T14:56:22.598188Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-17T14:58:27.598540Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-17T15:03:32.199053Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-17T15:08:30.539598Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-17T15:09:28.315827Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-17T15:10:43.578469Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-17T15:10:53.166917Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-17T15:11:50.406216Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-17T15:12:52.624552Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-17T15:13:05.070936Z`)

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
