# External Application Scoreboard

- Source measurements through: `2026-07-16T09:42:53.639840Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/32` rankable rows meet the 95%-of-Go target.
- Bytecode: `4/26` rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `64` row fingerprints in JSON; `64` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `64` row fingerprints in JSON; `64` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `69` `.able` files, tree SHA-256 `44a1adeafa85b2aec82fa18b4adb1d2903f8103aa9c58953c4b89767f20c3052`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `58` reviewed benchmark/mode rows, SHA-256 `11caaee63c66fa2e235249640fe0ce44833dc6ed9946d7d0a1e840997345c132`; timeout rows remain in full status.
- Matched reference source fingerprints: `96` comparison fingerprints in JSON; `96` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.3040 | 2.9564 / 1.12x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 29.1060 | 5.3740 / 5.42x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0720 | 0.9282 / 1.15x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.7960 | 2.3469 / 0.77x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 8.2760 | 0.5401 / 15.32x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1120 | 0.0570 / 1.96x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.3000 | 2.3283 / 0.99x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7040 | 1.4079 / 0.50x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1960 | 0.1957 / 1.00x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3000 | 1.1586 / 1.12x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1340 | 0.0473 / 2.83x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1160 | 0.0137 / 8.47x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.4180 | 0.0551 / 62.03x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.4060 | 0.0306 / 13.27x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.4200 | 1.8406 / 1.86x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1480 | n/a | 54.3191 / 0.00x | 43.2450 / 0.00x | `meets` | — |
| `matrixmultiply` | `bytecode` | `verified` | 4.6460 | n/a | 48.4717 / 0.10x | 44.2432 / 0.11x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 11.6958 / n/a | 52.6225 / n/a | `unranked` | Able timed out |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 22.7834 / n/a | 14.7736 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 16.0351 / n/a | 20.5417 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5400 | n/a | 0.0760 / 7.11x | 0.1070 / 5.05x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.8260 | n/a | 3.5852 / 0.79x | 2.3263 / 1.21x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8700 | n/a | 2.4928 / 0.35x | 1.5667 / 0.56x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4480 | n/a | 1.4348 / 1.71x | 1.4870 / 1.65x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.2460 | n/a | 3.8071 / 0.59x | 9.6658 / 0.23x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.2180 | n/a | 1.1577 / 5.37x | 1.7797 / 3.49x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 6.7720 | n/a | 0.0243 / 278.68x | 0.0684 / 99.01x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 40.6020 | n/a | 1.2149 / 33.42x | 1.1941 / 34.00x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.8558 / n/a | 3.0844 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | 54.1871 / n/a | 68.5713 / n/a | `unranked` | Able timed out |
| `channel_rollup` | `compiled` | `verified` | 1.2000 | 0.0045 / 266.67x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.6700 | 0.0043 / 155.81x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0033 / 36.36x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3460 | 0.0041 / 84.39x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.5120 | 0.0035 / 146.29x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.4140 | 0.0031 / 133.55x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5720 | n/a | 0.0368 / 15.54x | 0.0477 / 11.99x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4960 | n/a | 0.0593 / 8.36x | 0.0689 / 7.20x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1580 | n/a | 0.0294 / 5.37x | 0.0499 / 3.17x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2520 | n/a | 0.1181 / 2.13x | 0.0927 / 2.72x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.6840 | n/a | 0.0298 / 22.95x | 0.0479 / 14.28x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2480 | n/a | 0.0184 / 13.48x | 0.0405 / 6.12x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 8.3500 | 0.0048 / 1739.58x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 2.2720 | 0.0120 / 189.33x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2160 | 0.0045 / 48.00x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0960 | 0.0035 / 27.43x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1020 | 0.0036 / 28.33x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 2.5940 | 0.0309 / 83.95x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1900 | 0.0041 / 46.34x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1820 | 0.0040 / 45.50x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0860 | 0.0036 / 23.89x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0800 | 0.0032 / 25.00x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2000 | 0.0030 / 66.67x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.6840 | n/a | 0.3418 / 22.48x | 0.5958 / 12.90x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0140 | n/a | 0.0963 / 41.68x | 0.1256 / 31.96x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4460 | n/a | 0.0173 / 83.58x | 0.0473 / 30.57x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3140 | n/a | 0.0125 / 25.12x | 0.0389 / 8.07x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4440 | n/a | 0.0154 / 28.83x | 0.0448 / 9.91x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0380 / n/a | 0.0749 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 4.9720 | n/a | 0.0178 / 279.33x | 0.0394 / 126.19x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.3440 | n/a | 0.0169 / 257.04x | 0.0406 / 107.00x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7140 | n/a | 0.0260 / 27.46x | 0.0566 / 12.61x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4740 | n/a | 0.0146 / 32.47x | 0.0437 / 10.85x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 3.3580 | n/a | 0.0162 / 207.28x | 0.0408 / 82.30x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-01-selected.json` — `custom` (`2026-07-16T08:55:38.926525Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-02-selected.json` — `custom` (`2026-07-16T08:58:23.428345Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-03-selected.json` — `custom` (`2026-07-16T09:01:10.124483Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-04-selected.json` — `custom` (`2026-07-16T09:03:23.542115Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-05-selected.json` — `custom` (`2026-07-16T09:05:41.808365Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-compiled-06-selected.json` — `custom` (`2026-07-16T09:07:28.202407Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-01-selected.json` — `custom` (`2026-07-16T09:07:56.521681Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-01-status.json` — `custom` (`2026-07-16T09:09:28.600679Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-02-status.json` — `custom` (`2026-07-16T09:12:32.544605Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-03-selected.json` — `custom` (`2026-07-16T09:13:11.369579Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-04-selected.json` — `custom` (`2026-07-16T09:14:12.469447Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-05-selected.json` — `custom` (`2026-07-16T09:18:13.796129Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-bytecode-06-status.json` — `custom` (`2026-07-16T09:21:17.574597Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-async-compiled-01-selected.json` — `custom` (`2026-07-16T09:23:14.480130Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-async-compiled-02-selected.json` — `custom` (`2026-07-16T09:24:35.837863Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-async-bytecode-01-selected.json` — `custom` (`2026-07-16T09:24:48.428029Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-async-bytecode-02-selected.json` — `custom` (`2026-07-16T09:25:00.855479Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-16T09:27:59.147940Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-16T09:33:08.258254Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-16T09:37:47.030475Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-16T09:38:41.493065Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-16T09:39:53.687017Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-16T09:40:01.831980Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-16T09:41:33.761156Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-16T09:42:30.246454Z`)
- `v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-16T09:42:53.639840Z`)

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
