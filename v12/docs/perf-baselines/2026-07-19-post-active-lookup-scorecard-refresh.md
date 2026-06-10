# External Application Scoreboard

- Source measurements through: `2026-07-19T07:49:32.054624Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/36` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/28` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `72` row fingerprints in JSON; `72` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `72` row fingerprints in JSON; `72` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `64` reviewed benchmark/mode rows, SHA-256 `f46fccb1a156cb0574e4c0418683454feb8412b19089b778d7bc7820c57325b0`; timeout rows remain in full status.
- Matched reference source fingerprints: `103` comparison fingerprints in JSON; `103` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.3560 | 3.1759 / 1.06x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 10.1340 | 10.9270 / 0.93x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.2740 | 1.0300 / 1.24x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.8960 | 2.5337 / 0.75x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 9.5620 | 0.6139 / 15.58x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1340 | 0.0722 / 1.86x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.7860 | 2.5237 / 1.10x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.8480 | 1.5461 / 0.55x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2340 | 0.2742 / 0.85x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.4560 | 1.3091 / 1.11x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1740 | 0.0544 / 3.20x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1140 | 0.0171 / 6.67x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.6180 | 0.0575 / 62.92x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1560 | 0.0135 / 11.56x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1940 | 0.0349 / 5.56x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.7700 | 1.9210 / 1.96x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1340 | 0.0117 / 11.45x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1360 | 0.0106 / 12.83x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2100 | n/a | n/a | 47.6815 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 5.2200 | n/a | 51.2257 / 0.10x | 46.4778 / 0.11x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.7130 / n/a | 15.7057 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.4509 / n/a | 21.6375 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5520 | n/a | 0.0802 / 6.88x | 0.1154 / 4.78x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.2580 | n/a | 3.9203 / 0.83x | 2.5728 / 1.27x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8720 | n/a | 2.7564 / 0.32x | 1.8232 / 0.48x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.8260 | n/a | 1.7163 / 1.65x | 1.6925 / 1.67x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.8460 | n/a | 4.2279 / 0.67x | 10.4097 / 0.27x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.3540 | n/a | 1.2181 / 5.22x | 2.0232 / 3.14x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 4.6640 | n/a | 0.0260 / 179.38x | 0.0764 / 61.05x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 43.4980 | n/a | 1.3393 / 32.48x | 1.2754 / 34.11x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 2.2820 | n/a | 0.2081 / 10.97x | 0.3124 / 7.30x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0295 / n/a | 3.0979 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.9060 | n/a | 0.5866 / 10.07x | 0.3276 / 18.03x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.9080 | n/a | 0.8974 / 5.47x | 0.5496 / 8.93x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.6240 | 0.0055 / 113.45x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.4320 | 0.0056 / 77.14x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1260 | 0.0041 / 30.73x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3620 | 0.0046 / 78.70x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8080 | 0.0047 / 171.91x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7700 | 0.0040 / 192.50x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.6240 | n/a | 0.0413 / 15.11x | 0.0551 / 11.32x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4260 | n/a | 0.0624 / 6.83x | 0.0697 / 6.11x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1700 | n/a | 0.0338 / 5.03x | 0.0546 / 3.11x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2740 | n/a | 0.1146 / 2.39x | 0.0966 / 2.84x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4560 | n/a | 0.0325 / 14.03x | 0.0543 / 8.40x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2320 | n/a | 0.0254 / 9.13x | 0.0533 / 4.35x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2360 | 0.0056 / 42.14x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1760 | 0.0139 / 12.66x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2220 | 0.0053 / 41.89x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1120 | 0.0041 / 27.32x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1440 | 0.0041 / 35.12x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 1.0320 | 0.0449 / 22.98x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1160 | 0.0053 / 21.89x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1180 | 0.0052 / 22.69x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0940 | 0.0047 / 20.00x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1060 | 0.0046 / 23.04x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2440 | 0.0035 / 69.71x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2840 | 0.0096 / 29.58x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.1700 | n/a | 0.4314 / 18.94x | 0.6846 / 11.93x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.9240 | n/a | 0.1035 / 37.91x | 0.1328 / 29.55x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.6540 | n/a | 0.0221 / 74.84x | 0.0746 / 22.17x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3360 | n/a | 0.0173 / 19.42x | 0.0490 / 6.86x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4780 | n/a | 0.0232 / 20.60x | 0.0513 / 9.32x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0381 / n/a | 0.0764 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 4.2720 | n/a | 0.0198 / 215.76x | 0.0451 / 94.72x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5780 | n/a | 0.0200 / 178.90x | 0.0518 / 69.07x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7600 | n/a | 0.0490 / 15.51x | 0.0702 / 10.83x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5620 | n/a | 0.0170 / 33.06x | 0.0485 / 11.59x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.9880 | n/a | 0.0176 / 56.14x | 0.0461 / 21.43x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.5180 | n/a | 0.2559 / 13.75x | 0.3333 / 10.56x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-01-selected.json` — `custom` (`2026-07-19T07:00:21.583654Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-02-selected.json` — `custom` (`2026-07-19T07:03:31.219061Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-03-selected.json` — `custom` (`2026-07-19T07:06:39.425492Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-04-selected.json` — `custom` (`2026-07-19T07:09:09.001158Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-05-selected.json` — `custom` (`2026-07-19T07:12:43.715664Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-06-selected.json` — `custom` (`2026-07-19T07:14:43.056973Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-compiled-07-selected.json` — `custom` (`2026-07-19T07:15:44.917809Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-01-status.json` — `custom` (`2026-07-19T07:17:04.838748Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-02-status.json` — `custom` (`2026-07-19T07:18:58.971407Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-03-selected.json` — `custom` (`2026-07-19T07:19:41.013457Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-04-selected.json` — `custom` (`2026-07-19T07:20:48.181465Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-05-selected.json` — `custom` (`2026-07-19T07:25:07.920661Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-06-status.json` — `custom` (`2026-07-19T07:27:02.314094Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-generality-bytecode-07-selected.json` — `custom` (`2026-07-19T07:28:01.313287Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-async-compiled-01-selected.json` — `custom` (`2026-07-19T07:29:18.826494Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-async-compiled-02-selected.json` — `custom` (`2026-07-19T07:30:11.628657Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-async-bytecode-01-selected.json` — `custom` (`2026-07-19T07:30:25.119658Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-async-bytecode-02-selected.json` — `custom` (`2026-07-19T07:30:37.089660Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-19T07:32:59.791896Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-19T07:38:36.147275Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-19T07:43:49.765626Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-19T07:45:47.173781Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-19T07:47:03.144967Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-19T07:47:11.981629Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-19T07:48:09.126347Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-19T07:48:59.307601Z`)
- `v12/docs/perf-baselines/2026-07-19-post-active-lookup-scorecard-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-19T07:49:32.054624Z`)

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
