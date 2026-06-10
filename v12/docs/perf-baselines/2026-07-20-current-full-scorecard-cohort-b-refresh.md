# External Application Scoreboard

- Source measurements through: `2026-07-20T11:52:22.495468Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `5/36` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/29` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `72` row fingerprints in JSON; `72` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `72` row fingerprints in JSON; `72` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `65` reviewed benchmark/mode rows, SHA-256 `e7b35985b05134e1619be193cbe21ddce846cc2392efe78560e629de048d97dc`; timeout rows remain in full status.
- Matched reference source fingerprints: `103` comparison fingerprints in JSON; `103` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.7060 | 3.2228 / 1.15x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 10.1640 | 10.7063 / 0.95x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1860 | 1.0698 / 1.11x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.9100 | 2.7207 / 0.70x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.8780 | 0.5754 / 3.26x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1320 | 0.0611 / 2.16x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5720 | 2.4789 / 1.04x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.8240 | 1.4693 / 0.56x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1940 | 0.2686 / 0.72x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3940 | 1.2188 / 1.14x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1840 | 0.0564 / 3.26x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1100 | 0.0164 / 6.71x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.7340 | 0.0642 / 58.16x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1220 | 0.0162 / 7.53x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1820 | 0.0392 / 4.64x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.0680 | 1.9829 / 2.05x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1120 | 0.0127 / 8.82x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0900 | 0.0115 / 7.83x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1600 | n/a | n/a | 49.5633 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.8000 | n/a | 50.1725 / 0.10x | 46.6357 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.6353 / n/a | 15.0188 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.0974 / n/a | 21.8181 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5640 | n/a | 0.1105 / 5.10x | 0.1187 / 4.75x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.7900 | n/a | 3.8955 / 0.72x | 2.4923 / 1.12x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8060 | n/a | 2.9484 / 0.27x | 1.8978 / 0.42x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4580 | n/a | 1.5080 / 1.63x | 1.5772 / 1.56x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.4000 | n/a | 4.0316 / 0.60x | 10.4068 / 0.23x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.3720 | n/a | 1.2386 / 5.14x | 1.9265 / 3.31x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.5540 | n/a | 0.0291 / 122.13x | 0.0946 / 37.57x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 42.4280 | n/a | 1.4176 / 29.93x | 1.1996 / 35.37x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.7420 | n/a | 0.2172 / 8.02x | 0.3071 / 5.67x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0987 / n/a | 3.3071 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.5040 | n/a | 0.5533 / 9.95x | 0.3271 / 16.83x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.5240 | n/a | 0.8486 / 5.33x | 0.5668 / 7.98x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5960 | 0.0059 / 101.02x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3880 | 0.0062 / 62.58x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1100 | 0.0041 / 26.83x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3640 | 0.0046 / 79.13x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8340 | 0.0049 / 170.20x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7560 | 0.0050 / 151.20x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5180 | n/a | 0.0432 / 11.99x | 0.0590 / 8.78x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4980 | n/a | 0.0654 / 7.61x | 0.0747 / 6.67x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1820 | n/a | 0.0409 / 4.45x | 0.0814 / 2.24x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2760 | n/a | 0.1332 / 2.07x | 0.1004 / 2.75x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4320 | n/a | 0.0325 / 13.29x | 0.0619 / 6.98x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2240 | n/a | 0.0193 / 11.61x | 0.0504 / 4.44x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2700 | 0.0055 / 49.09x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.2240 | 0.0141 / 15.89x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2440 | 0.0062 / 39.35x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1040 | 0.0045 / 23.11x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1340 | 0.0047 / 28.51x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1580 | 0.0051 / 30.98x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1280 | 0.0056 / 22.86x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1280 | 0.0055 / 23.27x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1100 | 0.0047 / 23.40x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0940 | 0.0043 / 21.86x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2100 | 0.0041 / 51.22x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2560 | 0.0115 / 22.26x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.8020 | n/a | 0.5487 / 14.22x | 0.7738 / 10.08x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.9260 | n/a | 0.1178 / 33.33x | 0.1402 / 28.00x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4640 | n/a | 0.0189 / 77.46x | 0.0512 / 28.59x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3320 | n/a | 0.0143 / 23.22x | 0.0455 / 7.30x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4860 | n/a | 0.0186 / 26.13x | 0.0514 / 9.46x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.4660 | n/a | 0.0189 / 183.39x | 0.0449 / 77.19x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.2400 | n/a | 0.0192 / 220.83x | 0.0457 / 92.78x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5580 | n/a | 0.0190 / 187.26x | 0.0423 / 84.11x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7340 | n/a | 0.0277 / 26.50x | 0.0613 / 11.97x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4880 | n/a | 0.0158 / 30.89x | 0.0467 / 10.45x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8940 | n/a | 0.0175 / 51.09x | 0.0470 / 19.02x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.4540 | n/a | 0.2206 / 15.66x | 0.3515 / 9.83x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-01-selected.json` — `custom` (`2026-07-20T11:05:16.648004Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-02-selected.json` — `custom` (`2026-07-20T11:07:46.135037Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-03-selected.json` — `custom` (`2026-07-20T11:10:54.072950Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-04-selected.json` — `custom` (`2026-07-20T11:13:23.872254Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-05-selected.json` — `custom` (`2026-07-20T11:17:00.152357Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-06-selected.json` — `custom` (`2026-07-20T11:18:59.274053Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-compiled-07-selected.json` — `custom` (`2026-07-20T11:20:00.471165Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-01-status.json` — `custom` (`2026-07-20T11:21:07.132526Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-02-status.json` — `custom` (`2026-07-20T11:23:01.518870Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T11:23:40.593107Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T11:24:43.833043Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T11:28:49.637559Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-06-status.json` — `custom` (`2026-07-20T11:30:43.778498Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T11:31:38.694073Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-async-compiled-01-selected.json` — `custom` (`2026-07-20T11:32:55.324318Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-async-compiled-02-selected.json` — `custom` (`2026-07-20T11:33:47.904650Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-async-bytecode-01-selected.json` — `custom` (`2026-07-20T11:34:01.124624Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-async-bytecode-02-selected.json` — `custom` (`2026-07-20T11:34:13.067209Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T11:36:37.896578Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T11:42:10.817631Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T11:47:24.604859Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T11:49:19.826912Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T11:50:32.914022Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T11:51:01.249078Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T11:51:51.062360Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-b-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T11:52:22.495468Z`)

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
