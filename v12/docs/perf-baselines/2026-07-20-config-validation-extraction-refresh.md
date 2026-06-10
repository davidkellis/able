# External Application Scoreboard

- Source measurements through: `2026-07-20T21:56:44.721513Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `3/41` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/34` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `82` row fingerprints in JSON; `82` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `82` row fingerprints in JSON; `82` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `75` reviewed benchmark/mode rows, SHA-256 `f849d4ec36406fbb5c6739ccb41203e6d6b39450c61b0b894211e0d5863519a3`; timeout rows remain in full status.
- Matched reference source fingerprints: `118` comparison fingerprints in JSON; `118` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.5180 | 3.2054 / 1.10x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 10.0840 | 10.3476 / 0.97x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1200 | 0.9704 / 1.15x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 2.0840 | 2.4984 / 0.83x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.8720 | 0.5449 / 3.44x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1560 | 0.0578 / 2.70x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.6020 | 2.4625 / 1.06x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.8340 | 1.5125 / 0.55x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2320 | 0.2182 / 1.06x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.3960 | 1.2680 / 1.10x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1420 | 0.0541 / 2.62x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.0940 | 0.0179 / 5.25x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.5060 | 0.0617 / 56.82x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1160 | 0.0130 / 8.92x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1720 | 0.0329 / 5.23x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.7580 | 1.9219 / 1.96x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1140 | 0.0119 / 9.58x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1080 | 0.0103 / 10.49x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2200 | n/a | n/a | 47.1157 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.8200 | n/a | 51.8158 / 0.09x | 47.3598 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 26.2040 / n/a | 15.7993 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.5133 / n/a | 22.3665 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5760 | n/a | 0.0838 / 6.87x | 0.1125 / 5.12x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.1640 | n/a | 3.8809 / 0.82x | 2.3808 / 1.33x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.9100 | n/a | 2.6995 / 0.34x | 1.7315 / 0.53x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4440 | n/a | 1.5017 / 1.63x | 1.5246 / 1.60x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3500 | n/a | 4.0333 / 0.58x | 10.2466 / 0.23x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 7.1460 | n/a | 1.1668 / 6.12x | 1.8306 / 3.90x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.5200 | n/a | 0.0293 / 120.14x | 0.0819 / 42.98x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 43.3140 | n/a | 1.3555 / 31.95x | 1.2526 / 34.58x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.7940 | n/a | 0.1978 / 9.07x | 0.2889 / 6.21x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.7135 / n/a | 3.3038 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.5640 | n/a | 0.5432 / 10.24x | 0.3103 / 17.93x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.6240 | n/a | 0.8306 / 5.57x | 0.5339 / 8.66x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.6080 | 0.0053 / 114.72x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3380 | 0.0052 / 65.00x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0038 / 31.58x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3640 | 0.0045 / 80.89x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.6780 | 0.0043 / 157.67x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7200 | 0.0041 / 175.61x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.5400 | 0.0051 / 301.96x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5120 | n/a | 0.0436 / 11.74x | 0.0679 / 7.54x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4540 | n/a | 0.0780 / 5.82x | 0.0738 / 6.15x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1620 | n/a | 0.0300 / 5.40x | 0.0543 / 2.98x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2000 | n/a | 0.1119 / 1.79x | 0.0972 / 2.06x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3540 | n/a | 0.0355 / 9.97x | 0.0579 / 6.11x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2240 | n/a | 0.0228 / 9.82x | 0.0500 / 4.48x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3260 | n/a | 0.0273 / 11.94x | 0.0491 / 6.64x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2080 | 0.0055 / 37.82x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1640 | 0.0139 / 11.80x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.1820 | 0.0253 / 7.19x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1500 | 0.0052 / 28.85x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1200 | 0.0053 / 22.64x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1000 | 0.0050 / 20.00x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1100 | 0.0057 / 19.30x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1000 | 0.0057 / 17.54x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1300 | 0.0050 / 26.00x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.1100 | 0.0049 / 22.45x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0940 | 0.0048 / 19.58x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1080 | 0.0042 / 25.71x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2840 | 0.0090 / 31.56x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2360 | 0.0035 / 67.43x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.3180 | 0.0101 / 31.49x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.1180 | 0.0043 / 27.44x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.8000 | n/a | 0.3331 / 23.42x | 0.6513 / 11.98x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0740 | n/a | 0.1084 / 37.58x | 0.1439 / 28.31x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.3520 | n/a | 0.0631 / 84.82x | 0.1369 / 39.09x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5220 | n/a | 0.0201 / 75.72x | 0.0532 / 28.61x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.4200 | n/a | 0.0135 / 31.11x | 0.0442 / 9.50x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4760 | n/a | 0.0180 / 26.44x | 0.0506 / 9.41x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.3800 | n/a | 0.0188 / 179.79x | 0.0452 / 74.78x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.3540 | n/a | 0.0194 / 224.43x | 0.0467 / 93.23x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.0600 | n/a | 0.0202 / 200.99x | 0.0498 / 81.53x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 3.8540 | n/a | 0.0189 / 203.92x | 0.0437 / 88.19x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7700 | n/a | 0.0300 / 25.67x | 0.0637 / 12.09x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5720 | n/a | 0.0172 / 33.26x | 0.0526 / 10.87x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 3.0940 | n/a | 0.0721 / 42.91x | 0.1146 / 27.00x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.9060 | n/a | 0.0219 / 41.37x | 0.0504 / 17.98x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 4.1180 | n/a | 0.2431 / 16.94x | 0.3123 / 13.19x | `miss` | — |
| `config_validation_extraction` | `bytecode` | `verified` | 1.7020 | n/a | 0.0196 / 86.84x | 0.0445 / 38.25x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-01-selected.json` — `custom` (`2026-07-20T21:02:00.831315Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-02-selected.json` — `custom` (`2026-07-20T21:04:29.117501Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-03-selected.json` — `custom` (`2026-07-20T21:07:37.036863Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-04-selected.json` — `custom` (`2026-07-20T21:10:00.113772Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-05-selected.json` — `custom` (`2026-07-20T21:13:30.916306Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-06-selected.json` — `custom` (`2026-07-20T21:15:24.495491Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-compiled-07-selected.json` — `custom` (`2026-07-20T21:16:33.245757Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-01-status.json` — `custom` (`2026-07-20T21:17:39.938490Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-02-status.json` — `custom` (`2026-07-20T21:19:34.436300Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T21:20:18.662692Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T21:21:25.362854Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T21:25:35.972612Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-06-status.json` — `custom` (`2026-07-20T21:27:30.047876Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T21:28:25.658112Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-async-compiled-01-selected.json` — `custom` (`2026-07-20T21:29:37.223865Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-async-compiled-02-selected.json` — `custom` (`2026-07-20T21:30:46.492384Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-async-bytecode-01-selected.json` — `custom` (`2026-07-20T21:30:59.059241Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-async-bytecode-02-selected.json` — `custom` (`2026-07-20T21:31:13.755474Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T21:34:25.272012Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T21:39:48.717026Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T21:47:16.838967Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T21:49:54.431817Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-compiled-05-selected.json` — `custom` (`2026-07-20T21:52:12.924041Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T21:53:55.868153Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T21:54:24.619749Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T21:55:39.867753Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T21:56:33.338724Z`)
- `v12/docs/perf-baselines/2026-07-20-config-validation-extraction-coverage-extra-bytecode-05-selected.json` — `custom` (`2026-07-20T21:56:44.721513Z`)

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
