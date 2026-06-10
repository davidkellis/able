# External Application Scoreboard

- Source measurements through: `2026-07-15T21:32:43.761631Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/32` rankable rows meet the 95%-of-Go target.
- Bytecode: `3/24` rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `66` row fingerprints in JSON; `2` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `66` row fingerprints in JSON; `2` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Matched reference source fingerprints: `97` comparison fingerprints in JSON; `6` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.1600 | 2.8935 / 1.09x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 26.0733 | 29.8424 / 0.87x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0133 | 0.8872 / 1.14x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.6733 | 2.2544 / 0.74x | n/a | n/a | `meets` | — |
| `sudoku` | `compiled` | `timeout` | n/a | 0.1292 / n/a | n/a | n/a | `unranked` | Able timed out |
| `sudoku_masks` | `compiled` | `verified` | 9.8833 | 0.5141 / 19.22x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1100 | 0.0543 / 2.03x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.1500 | 2.1969 / 0.98x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.6467 | 1.2959 / 0.50x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1967 | 0.1837 / 1.07x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.2067 | 1.0615 / 1.14x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1400 | 0.0461 / 3.04x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1100 | 0.0141 / 7.80x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.1100 | 0.0500 / 62.20x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.3700 | 0.0301 / 12.29x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.5600 | 1.7124 / 2.08x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1500 | n/a | n/a | 40.1801 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 3.8100 | n/a | 44.0134 / 0.09x | 41.9293 / 0.09x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 22.6439 / n/a | 13.4652 / n/a | `unranked` | Able timed out |
| `sudoku` | `bytecode` | `timeout` | n/a | n/a | 2.6042 / n/a | 5.2667 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 15.2837 / n/a | 18.4365 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5100 | n/a | 0.0733 / 6.96x | 0.1017 / 5.01x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.6633 | n/a | 3.4421 / 0.77x | 2.1771 / 1.22x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.7267 | n/a | 2.3201 / 0.31x | 1.4611 / 0.50x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.1433 | n/a | 1.2851 / 1.67x | 1.3522 / 1.59x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.2667 | n/a | 3.6258 / 0.63x | 8.8207 / 0.26x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.1400 | n/a | 1.0302 / 5.96x | 1.6353 / 3.75x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 6.2000 | n/a | 0.0233 / 266.09x | 0.0661 / 93.80x | `miss` | — |
| `k_nucleotide` | `bytecode` | `incomplete` | 38.1550 | n/a | 1.1287 / 33.80x | 1.1156 / 34.20x | `unranked` | Able run incomplete |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.7544 / n/a | 2.7624 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `channel_rollup` | `compiled` | `verified` | 1.1367 | 0.0052 / 218.60x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.6167 | 0.0065 / 94.88x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0039 / 30.77x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3300 | 0.0044 / 75.00x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.5100 | 0.0041 / 124.39x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.4200 | 0.0038 / 110.53x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5600 | n/a | 0.0365 / 15.34x | 0.0471 / 11.89x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4100 | n/a | 0.0542 / 7.56x | 0.0650 / 6.31x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1467 | n/a | 0.0280 / 5.24x | 0.0489 / 3.00x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2300 | n/a | 0.1084 / 2.12x | 0.0847 / 2.72x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.6233 | n/a | 0.0361 / 17.27x | 0.0481 / 12.96x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2300 | n/a | 0.0186 / 12.37x | 0.0412 / 5.58x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 7.4033 | 0.0053 / 1396.85x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 2.1633 | 0.0123 / 175.88x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2133 | 0.0049 / 43.53x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0933 | 0.0038 / 24.55x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1100 | 0.0045 / 24.44x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 2.4000 | 0.0320 / 75.00x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1733 | 0.0060 / 28.88x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1733 | 0.0051 / 33.98x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0833 | 0.0041 / 20.32x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0767 | 0.0037 / 20.73x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 6.7000 | n/a | 0.3386 / 19.79x | 0.8400 / 7.98x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.3767 | n/a | 0.1033 / 32.69x | 0.1457 / 23.18x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.3000 | n/a | 0.0189 / 68.78x | 0.0593 / 21.92x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.2833 | n/a | 0.0143 / 19.81x | 0.0427 / 6.63x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4433 | n/a | 0.0165 / 26.87x | 0.0500 / 8.87x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0418 / n/a | 0.0724 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 4.7000 | n/a | 0.0187 / 251.34x | 0.0412 / 114.08x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.0733 | n/a | 0.0191 / 213.26x | 0.0414 / 98.39x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.5600 | n/a | 0.0270 / 20.74x | 0.0606 / 9.24x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4500 | n/a | 0.0187 / 24.06x | 0.0441 / 10.20x | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.1960 | 0.0030 / 65.33x | 0.0152 / 12.89x | 0.0409 / 4.79x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 3.3880 | 0.0030 / 1129.33x | 0.0152 / 222.89x | 0.0409 / 82.84x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-01.json` — `custom` (`2026-07-15T08:20:49.221407Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-02.json` — `custom` (`2026-07-15T08:26:29.162168Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-03.json` — `custom` (`2026-07-15T08:28:59.695529Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-04.json` — `custom` (`2026-07-15T08:30:56.124259Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-05.json` — `custom` (`2026-07-15T08:32:56.842687Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-compiled-06.json` — `custom` (`2026-07-15T08:34:27.123765Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-01.json` — `custom` (`2026-07-15T08:37:01.116353Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-02.json` — `custom` (`2026-07-15T08:43:51.684101Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-03.json` — `custom` (`2026-07-15T08:44:15.399511Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-04.json` — `custom` (`2026-07-15T08:44:53.048451Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-05.json` — `custom` (`2026-07-15T08:47:16.930117Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-bytecode-06.json` — `custom` (`2026-07-15T08:51:50.826208Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-compiled-01.json` — `custom` (`2026-07-15T08:53:42.961251Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-compiled-02.json` — `custom` (`2026-07-15T08:54:55.736896Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-bytecode-01.json` — `custom` (`2026-07-15T08:55:04.755425Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-async-bytecode-02.json` — `custom` (`2026-07-15T08:55:13.705737Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-compiled-01.json` — `custom` (`2026-07-15T08:57:35.897503Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-compiled-02.json` — `custom` (`2026-07-15T09:02:32.110648Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-compiled-03.json` — `custom` (`2026-07-15T09:07:32.155786Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-compiled-04.json` — `custom` (`2026-07-15T09:07:58.287362Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-bytecode-01.json` — `custom` (`2026-07-15T09:08:38.103786Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-bytecode-02.json` — `custom` (`2026-07-15T09:11:01.130738Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-bytecode-03.json` — `custom` (`2026-07-15T09:11:34.805359Z`)
- `v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-coverage-extra-bytecode-04.json` — `custom` (`2026-07-15T09:11:38.079353Z`)
- `v12/docs/perf-baselines/2026-07-15-option-result-scorecard-coverage-postfix.json` — `custom` (`2026-07-15T21:32:43.761631Z`)

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
