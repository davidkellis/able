# External Application Scoreboard

- Source measurements through: `2026-07-20T13:50:04.426360Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/37` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/30` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `74` row fingerprints in JSON; `74` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `74` row fingerprints in JSON; `74` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `67` reviewed benchmark/mode rows, SHA-256 `15fe3f6c76ba1fa495d565ff6dd79aac3f886c28a72a511714fedea671fdc4d8`; timeout rows remain in full status.
- Matched reference source fingerprints: `106` comparison fingerprints in JSON; `106` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.6260 | 3.3287 / 1.09x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.7100 | 10.7807 / 0.90x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0960 | 0.9734 / 1.13x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.9780 | 2.4466 / 0.81x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.9280 | 0.5698 / 3.38x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1300 | 0.0584 / 2.23x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5420 | 2.4287 / 1.05x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7680 | 1.5618 / 0.49x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.3400 | 0.1989 / 1.71x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.3760 | 1.1665 / 1.18x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1620 | 0.0513 / 3.16x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1000 | 0.0165 / 6.06x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.5380 | 0.0615 / 57.53x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1120 | 0.0145 / 7.72x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1620 | 0.0331 / 4.89x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.6800 | 1.8996 / 1.94x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1080 | 0.0124 / 8.71x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0900 | 0.0122 / 7.38x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1800 | n/a | n/a | 44.1456 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.7700 | n/a | 48.9848 / 0.10x | 46.2509 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.2550 / n/a | 15.1683 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.3745 / n/a | 20.7802 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5640 | n/a | 0.1094 / 5.16x | 0.1128 / 5.00x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9460 | n/a | 3.8227 / 0.77x | 2.3747 / 1.24x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.9180 | n/a | 2.5392 / 0.36x | 1.8617 / 0.49x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4320 | n/a | 1.5864 / 1.53x | 1.5736 / 1.55x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3640 | n/a | 3.9047 / 0.61x | 9.9726 / 0.24x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.3680 | n/a | 1.3239 / 4.81x | 1.8933 / 3.36x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.3880 | n/a | 0.0256 / 132.34x | 0.0718 / 47.19x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 43.3620 | n/a | 1.2992 / 33.38x | 1.2436 / 34.87x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.8060 | n/a | 0.1990 / 9.08x | 0.2876 / 6.28x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0295 / n/a | 2.9963 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.7420 | n/a | 0.5776 / 9.94x | 0.3518 / 16.32x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.5880 | n/a | 0.8272 / 5.55x | 0.5110 / 8.98x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5500 | 0.0061 / 90.16x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3800 | 0.0057 / 66.67x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1120 | 0.0040 / 28.00x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3420 | 0.0047 / 72.77x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.7900 | 0.0048 / 164.58x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6060 | 0.0040 / 151.50x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5060 | n/a | 0.0396 / 12.78x | 0.0552 / 9.17x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4260 | n/a | 0.0729 / 5.84x | 0.0855 / 4.98x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1620 | n/a | 0.0304 / 5.33x | 0.0559 / 2.90x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2120 | n/a | 0.1158 / 1.83x | 0.0906 / 2.34x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3780 | n/a | 0.0324 / 11.67x | 0.0493 / 7.67x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2140 | n/a | 0.0195 / 10.97x | 0.0434 / 4.93x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2140 | 0.0055 / 38.91x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1340 | 0.0138 / 9.71x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1800 | 0.0052 / 34.62x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1280 | 0.0041 / 31.22x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1000 | 0.0042 / 23.81x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1320 | 0.0048 / 27.50x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1120 | 0.0050 / 22.40x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1320 | 0.0048 / 27.50x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1180 | 0.0047 / 25.11x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1000 | 0.0037 / 27.03x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2980 | 0.0092 / 32.39x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2200 | 0.0035 / 62.86x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.3200 | 0.0096 / 33.33x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.2940 | n/a | 0.3510 / 23.63x | 0.6821 / 12.16x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0800 | n/a | 0.1516 / 26.91x | 0.1694 / 24.09x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5160 | n/a | 0.0202 / 75.05x | 0.0602 / 25.18x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3620 | n/a | 0.0147 / 24.63x | 0.0475 / 7.62x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4680 | n/a | 0.0187 / 25.03x | 0.0574 / 8.15x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.7300 | n/a | 0.0213 / 175.12x | 0.0469 / 79.53x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.3800 | n/a | 0.0216 / 202.78x | 0.0512 / 85.55x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5820 | n/a | 0.0208 / 172.21x | 0.0533 / 67.20x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6980 | n/a | 0.0372 / 18.76x | 0.0716 / 9.75x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5780 | n/a | 0.0161 / 35.90x | 0.0489 / 11.82x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.6500 | n/a | 0.0681 / 38.91x | 0.0919 / 28.84x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8360 | n/a | 0.0204 / 40.98x | 0.0521 / 16.05x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6180 | n/a | 0.2438 / 14.84x | 0.3479 / 10.40x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-01-selected.json` — `custom` (`2026-07-20T13:02:30.648883Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-02-selected.json` — `custom` (`2026-07-20T13:05:00.801433Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-03-selected.json` — `custom` (`2026-07-20T13:08:13.298541Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-04-selected.json` — `custom` (`2026-07-20T13:10:37.539994Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-05-selected.json` — `custom` (`2026-07-20T13:14:11.164458Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-06-selected.json` — `custom` (`2026-07-20T13:16:06.958277Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-compiled-07-selected.json` — `custom` (`2026-07-20T13:17:07.629522Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-01-status.json` — `custom` (`2026-07-20T13:18:14.031954Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-02-status.json` — `custom` (`2026-07-20T13:20:08.267695Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T13:20:50.051035Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T13:21:53.104553Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T13:26:03.126943Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-06-status.json` — `custom` (`2026-07-20T13:27:57.295458Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T13:28:53.792845Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-async-compiled-01-selected.json` — `custom` (`2026-07-20T13:30:06.973889Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-async-compiled-02-selected.json` — `custom` (`2026-07-20T13:30:55.639447Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-async-bytecode-01-selected.json` — `custom` (`2026-07-20T13:31:08.287663Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-async-bytecode-02-selected.json` — `custom` (`2026-07-20T13:31:19.324537Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T13:33:38.223720Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T13:39:05.171757Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T13:44:10.652463Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T13:46:38.922106Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T13:47:55.428631Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T13:48:25.830363Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T13:49:16.226914Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-a-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T13:50:04.426360Z`)

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
