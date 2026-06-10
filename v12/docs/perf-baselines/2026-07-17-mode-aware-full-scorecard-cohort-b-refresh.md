# External Application Scoreboard

- Source measurements through: `2026-07-17T16:32:55.228163Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `7/34` selected rankable rows meet the 95%-of-Go target.
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
| `fib` | `compiled` | `verified` | 3.5100 | 6.4866 / 0.54x | n/a | n/a | `meets` | — |
| `binarytrees` | `compiled` | `verified` | 11.1300 | 13.4281 / 0.83x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.3640 | 2.0850 / 0.65x | n/a | n/a | `meets` | — |
| `quicksort` | `compiled` | `verified` | 2.4780 | 5.6066 / 0.44x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 10.9260 | 1.2525 / 8.72x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1260 | 0.1747 / 0.72x | n/a | n/a | `meets` | — |
| `base64` | `compiled` | `verified` | 3.3280 | 2.5649 / 1.30x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.7760 | 1.4724 / 0.53x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2060 | 0.2336 / 0.88x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.5020 | 1.2071 / 1.24x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1300 | 0.0544 / 2.39x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1100 | 0.0148 / 7.43x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.5300 | 0.0702 / 50.28x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1800 | 0.0381 / 4.72x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.7160 | 2.2589 / 2.09x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1000 | 0.0163 / 6.13x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0880 | 0.0121 / 7.27x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `bytecode` | `verified` | 5.0420 | n/a | 51.2647 / 0.10x | 48.9405 / 0.10x | `meets` | — |
| `fib` | `bytecode` | `verified` | 0.1500 | n/a | n/a | 45.9127 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | 55.1481 / n/a | `unranked` | Able timed out |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.3514 / n/a | 14.8379 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.6773 / n/a | 23.6062 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6960 | n/a | 0.1099 / 6.33x | 0.1339 / 5.20x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.7940 | n/a | 4.4650 / 0.85x | 3.0291 / 1.25x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.9700 | n/a | 2.9555 / 0.33x | 1.6967 / 0.57x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.6540 | n/a | 1.5134 / 1.75x | 2.1670 / 1.22x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.2460 | n/a | 4.2015 / 0.53x | 9.9913 / 0.22x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.4680 | n/a | 1.2516 / 5.17x | 2.0392 / 3.17x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 6.9280 | n/a | 0.0292 / 237.26x | 0.0727 / 95.30x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 44.9580 | n/a | 1.2718 / 35.35x | 1.2332 / 36.46x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9372 / n/a | 3.0972 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 6.8860 | n/a | 0.5620 / 12.25x | 0.3312 / 20.79x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 7.5620 | n/a | 0.8473 / 8.92x | 0.5368 / 14.09x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 1.0500 | 0.0056 / 187.50x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3960 | 0.0062 / 63.87x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1140 | 0.0039 / 29.23x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3640 | 0.0051 / 71.37x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.6920 | 0.0048 / 144.17x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6980 | 0.0057 / 122.46x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5500 | n/a | 0.0442 / 12.44x | 0.0542 / 10.15x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4560 | n/a | 0.0573 / 7.96x | 0.0700 / 6.51x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1600 | n/a | 0.0301 / 5.32x | 0.0509 / 3.14x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2100 | n/a | 0.1167 / 1.80x | 0.0970 / 2.16x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3600 | n/a | 0.0311 / 11.58x | 0.0514 / 7.00x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2120 | n/a | 0.0192 / 11.04x | 0.0448 / 4.73x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2060 | 0.0076 / 27.11x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1240 | 0.0163 / 7.61x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2140 | 0.0061 / 35.08x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1020 | 0.0048 / 21.25x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1020 | 0.0058 / 17.59x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 1.2220 | 0.0457 / 26.74x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1120 | 0.0053 / 21.13x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1440 | 0.0053 / 27.17x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1600 | 0.0049 / 32.65x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0960 | 0.0040 / 24.00x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2080 | 0.0038 / 54.74x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.7500 | n/a | 0.3479 / 22.28x | 0.7390 / 10.49x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0120 | n/a | 0.1150 / 34.89x | 0.1477 / 27.16x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5100 | n/a | 0.0212 / 71.23x | 0.0575 / 26.26x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3760 | n/a | 0.0182 / 20.66x | 0.0469 / 8.02x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4560 | n/a | 0.0242 / 18.84x | 0.0556 / 8.20x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0431 / n/a | 0.0801 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 5.1860 | n/a | 0.0273 / 189.96x | 0.0555 / 93.44x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.5220 | n/a | 0.0236 / 191.61x | 0.0483 / 93.62x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7040 | n/a | 0.0337 / 20.89x | 0.0806 / 8.73x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4800 | n/a | 0.0214 / 22.43x | 0.0518 / 9.27x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8280 | n/a | 0.0176 / 47.05x | 0.0503 / 16.46x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-01-selected.json` — `custom` (`2026-07-17T15:43:41.441402Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-02-selected.json` — `custom` (`2026-07-17T15:47:13.898042Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-03-selected.json` — `custom` (`2026-07-17T15:51:20.728474Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-04-selected.json` — `custom` (`2026-07-17T15:53:54.374278Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-05-selected.json` — `custom` (`2026-07-17T15:56:25.734239Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-06-selected.json` — `custom` (`2026-07-17T15:58:35.546058Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-compiled-07-selected.json` — `custom` (`2026-07-17T15:59:34.253455Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-01-selected.json` — `custom` (`2026-07-17T16:00:01.887108Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-01-status.json` — `custom` (`2026-07-17T16:01:05.352068Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-02-status.json` — `custom` (`2026-07-17T16:03:08.158913Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-03-selected.json` — `custom` (`2026-07-17T16:03:56.860628Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-04-selected.json` — `custom` (`2026-07-17T16:05:00.686596Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-05-selected.json` — `custom` (`2026-07-17T16:09:24.820456Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-06-status.json` — `custom` (`2026-07-17T16:11:26.829419Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-generality-bytecode-07-selected.json` — `custom` (`2026-07-17T16:12:45.423158Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-async-compiled-01-selected.json` — `custom` (`2026-07-17T16:14:25.567142Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-async-compiled-02-selected.json` — `custom` (`2026-07-17T16:15:19.254898Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-async-bytecode-01-selected.json` — `custom` (`2026-07-17T16:15:31.851074Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-async-bytecode-02-selected.json` — `custom` (`2026-07-17T16:15:42.687915Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-17T16:18:01.054989Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-17T16:23:33.712163Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-17T16:28:22.801910Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-17T16:29:22.423867Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-17T16:30:35.647249Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-17T16:30:44.350278Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-17T16:31:45.342871Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-17T16:32:44.279961Z`)
- `v12/docs/perf-baselines/2026-07-17-mode-aware-full-scorecard-cohort-b-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-17T16:32:55.228163Z`)

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
