# External Application Scoreboard

- Source measurements through: `2026-07-18T17:23:34.111526Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `7/35` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `3/27` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `70` row fingerprints in JSON; `70` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `70` row fingerprints in JSON; `70` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `62` reviewed benchmark/mode rows, SHA-256 `9976ccea0e85b2acf92b019727e81b0ce88a347828b3ffb675d869cae81eca7c`; timeout rows remain in full status.
- Matched reference source fingerprints: `98` comparison fingerprints in JSON; `98` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `channel_rollup` | `bytecode` | `verified` | 0.7120 | n/a | 0.0431 / 16.52x | 0.0581 / 12.25x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.6060 | n/a | 0.0641 / 9.45x | 0.0837 / 7.24x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.2020 | n/a | 0.0436 / 4.63x | 0.0661 / 3.06x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.3120 | n/a | 0.1247 / 2.50x | 0.1048 / 2.98x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.5080 | n/a | 0.0416 / 12.21x | 0.0528 / 9.62x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.3000 | n/a | 0.0261 / 11.49x | 0.0626 / 4.79x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.6700 | 0.0084 / 79.76x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3720 | 0.0077 / 48.31x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1460 | 0.0054 / 27.04x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.5180 | 0.0072 / 71.94x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.9240 | 0.0057 / 162.11x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.8780 | 0.0052 / 168.85x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.0580 | n/a | 0.3823 / 21.08x | 0.7323 / 11.00x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.1540 | n/a | 0.1162 / 35.75x | 0.1669 / 24.89x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.7580 | n/a | 0.0228 / 77.11x | 0.0636 / 27.64x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3760 | n/a | 0.0153 / 24.58x | 0.0529 / 7.11x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4820 | n/a | 0.0282 / 17.09x | 0.0608 / 7.93x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0491 / n/a | 0.1086 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 4.6240 | n/a | 0.0224 / 206.43x | 0.0497 / 93.04x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.8340 | n/a | 0.0229 / 167.42x | 0.0563 / 68.10x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7580 | n/a | 0.0322 / 23.54x | 0.0710 / 10.68x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5440 | n/a | 0.0193 / 28.19x | 0.0558 / 9.75x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.9140 | n/a | 0.0190 / 48.11x | 0.0552 / 16.56x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6260 | n/a | 0.2903 / 12.49x | 0.3431 / 10.57x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2860 | 0.0097 / 29.48x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1700 | 0.0159 / 10.69x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.3080 | 0.0067 / 45.97x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1140 | 0.0050 / 22.80x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1200 | 0.0055 / 21.82x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 1.1400 | 0.0411 / 27.74x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1620 | 0.0066 / 24.55x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1280 | 0.0081 / 15.80x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1200 | 0.0070 / 17.14x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1280 | 0.0069 / 18.55x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2380 | 0.0048 / 49.58x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2720 | 0.0141 / 19.29x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `bytecode` | `verified` | 5.0520 | n/a | n/a | n/a | `unranked` | Python and Ruby references unavailable |
| `fib` | `bytecode` | `verified` | 0.2900 | n/a | n/a | 46.9796 / 0.01x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 27.9447 / n/a | 15.9612 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 27.4226 / n/a | 24.1037 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6820 | n/a | 0.1475 / 4.62x | 0.2052 / 3.32x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.1780 | n/a | 5.2732 / 0.60x | 3.3672 / 0.94x | `meets` | — |
| `json` | `bytecode` | `verified` | 0.9860 | n/a | 2.9249 / 0.34x | 1.7617 / 0.56x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 3.1340 | n/a | 2.0108 / 1.56x | 1.7294 / 1.81x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.6700 | n/a | 4.3496 / 0.61x | 11.0683 / 0.24x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 7.5760 | n/a | 1.2740 / 5.95x | 1.9700 / 3.85x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 7.5020 | n/a | 0.0269 / 278.88x | 0.0808 / 92.85x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 44.9340 | n/a | 1.3879 / 32.38x | 1.4065 / 31.95x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.3279 / n/a | 3.6897 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 6.6680 | n/a | 0.6158 / 10.83x | 0.3452 / 19.32x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 5.5580 | n/a | 0.8594 / 6.47x | 0.5674 / 9.80x | `miss` | — |
| `quicksort` | `compiled` | `verified` | 2.0840 | 3.2474 / 0.64x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 10.2320 | 0.6899 / 14.83x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1640 | 0.0242 / 6.78x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.8300 | 0.0843 / 45.43x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.2220 | 0.0429 / 5.17x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 5.5520 | 2.3799 / 2.33x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1640 | 0.0181 / 9.06x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1320 | 0.0133 / 9.92x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 10.8240 | 12.4042 / 0.87x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.9200 | 1.7014 / 0.54x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.4540 | 0.2495 / 1.82x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1920 | 0.0576 / 3.33x | n/a | n/a | `miss` | — |
| `fib` | `compiled` | `verified` | 3.3460 | 3.3636 / 0.99x | n/a | n/a | `meets` | — |
| `i_before_e` | `compiled` | `verified` | 0.1100 | 0.0977 / 1.13x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1540 | 1.1461 / 1.01x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3460 | 1.4199 / 0.95x | n/a | n/a | `meets` | — |
| `base64` | `compiled` | `verified` | 2.5360 | 3.0974 / 0.82x | n/a | n/a | `meets` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-bytecode-01-selected.json` — `custom` (`2026-07-18T16:55:17.442602Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-bytecode-02-selected.json` — `custom` (`2026-07-18T16:55:32.409534Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-compiled-01-selected.json` — `custom` (`2026-07-18T16:53:58.672893Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-async-compiled-02-selected.json` — `custom` (`2026-07-18T16:55:00.557230Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-18T17:13:08.515573Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-18T17:13:17.999084Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-18T17:14:15.153111Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-18T17:15:08.662668Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-18T17:15:41.296961Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-18T16:58:18.387803Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-18T17:04:22.782604Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-18T17:09:48.955392Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-18T17:11:51.067004Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-01-selected.json` — `custom` (`2026-07-18T16:40:09.961851Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-01-status.json` — `custom` (`2026-07-18T16:41:10.084275Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-02-status.json` — `custom` (`2026-07-18T16:43:04.928407Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-03-selected.json` — `custom` (`2026-07-18T16:43:50.283485Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-04-selected.json` — `custom` (`2026-07-18T16:45:04.899091Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-05-selected.json` — `custom` (`2026-07-18T16:49:32.550368Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-06-status.json` — `custom` (`2026-07-18T16:51:27.044630Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-bytecode-07-selected.json` — `custom` (`2026-07-18T16:52:33.854392Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-compiled-02-selected.json` — `custom` (`2026-07-18T16:25:57.288631Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-compiled-05-selected.json` — `custom` (`2026-07-18T16:35:47.755010Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-compiled-06-selected.json` — `custom` (`2026-07-18T16:38:23.047080Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-generality-compiled-07-selected.json` — `custom` (`2026-07-18T16:39:40.716916Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-classification-carryover-binarytrees.json` — `custom` (`2026-07-18T16:22:30.777211Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-classification-carryover-json.json` — `custom` (`2026-07-18T16:29:31.883331Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-compiled-selection-classification-carryover-monte-mandel.json` — `custom` (`2026-07-18T16:32:43.368533Z`)
- `v12/docs/perf-baselines/2026-07-18-post-string-classification-compiled-b.json` — `custom` (`2026-07-18T17:23:34.111526Z`)

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
