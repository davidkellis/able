# External Application Scoreboard

- Source measurements through: `2026-07-20T16:55:19.324456Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `6/38` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/31` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `76` row fingerprints in JSON; `76` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `76` row fingerprints in JSON; `76` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `69` reviewed benchmark/mode rows, SHA-256 `9f1d7eae4a05b7cdb4c378906ebc894648092c61f486a5db4cfd4da5aa1ee7ec`; timeout rows remain in full status.
- Matched reference source fingerprints: `113` comparison fingerprints in JSON; `113` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.5460 | 3.2012 / 1.11x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.4820 | 10.2334 / 0.93x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1180 | 0.9513 / 1.18x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.8580 | 2.4006 / 0.77x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.9640 | 0.5497 / 3.57x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1260 | 0.1083 / 1.16x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.4640 | 2.4286 / 1.01x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7220 | 1.6141 / 0.45x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2300 | 0.2492 / 0.92x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3500 | 1.3842 / 0.98x | n/a | n/a | `meets` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1400 | 0.0578 / 2.42x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1140 | 0.0174 / 6.55x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.1220 | 0.0905 / 34.50x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1060 | 0.0141 / 7.52x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1720 | 0.0457 / 3.76x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.4780 | 2.1681 / 1.60x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.0860 | 0.0128 / 6.72x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0880 | 0.0110 / 8.00x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1400 | n/a | 57.6431 / 0.00x | 45.2561 / 0.00x | `meets` | — |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 55.6113 / n/a | 55.5659 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.6800 | n/a | 49.5857 / 0.09x | 45.3768 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.2582 / n/a | 14.7737 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.0840 / n/a | 20.1946 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5360 | n/a | 0.1091 / 4.91x | 0.1219 / 4.40x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.7700 | n/a | 3.9303 / 0.70x | 2.4816 / 1.12x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8820 | n/a | 2.6015 / 0.34x | 1.6757 / 0.53x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.5500 | n/a | 1.4779 / 1.73x | 1.4934 / 1.71x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3420 | n/a | 4.1673 / 0.56x | 10.2967 / 0.23x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.3280 | n/a | 1.3088 / 4.83x | 1.8618 / 3.40x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.4540 | n/a | 0.0294 / 117.48x | 0.0832 / 41.51x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 40.6580 | n/a | 1.3240 / 30.71x | 1.2237 / 33.23x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.7620 | n/a | 0.2017 / 8.74x | 0.3117 / 5.65x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9403 / n/a | 3.1281 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | 56.2835 / n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.4720 | n/a | 0.5255 / 10.41x | 0.3179 / 17.21x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.7160 | n/a | 0.8350 / 5.65x | 0.5088 / 9.27x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5620 | 0.0059 / 95.25x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3140 | 0.0060 / 52.33x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0900 | 0.0050 / 18.00x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3860 | 0.0047 / 82.13x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.7200 | 0.0044 / 163.64x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6780 | 0.0042 / 161.43x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.4260 | 0.0046 / 310.00x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5260 | n/a | 0.0377 / 13.95x | 0.0503 / 10.46x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4820 | n/a | 0.0553 / 8.72x | 0.0688 / 7.01x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1680 | n/a | 0.0317 / 5.30x | 0.0571 / 2.94x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2260 | n/a | 0.1143 / 1.98x | 0.0917 / 2.46x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3660 | n/a | 0.0321 / 11.40x | 0.0521 / 7.02x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2360 | n/a | 0.0201 / 11.74x | 0.0487 / 4.85x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3500 | n/a | 0.0384 / 9.11x | 0.0448 / 7.81x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2460 | 0.0074 / 33.24x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1380 | 0.0134 / 10.30x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1740 | 0.0067 / 25.97x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0960 | 0.0043 / 22.33x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1200 | 0.0046 / 26.09x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1560 | 0.0055 / 28.36x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1200 | 0.0052 / 23.08x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1240 | 0.0052 / 23.85x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0940 | 0.0054 / 17.41x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1080 | 0.0041 / 26.34x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2720 | 0.0100 / 27.20x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2040 | 0.0057 / 35.79x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2700 | 0.0097 / 27.84x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.7920 | n/a | 0.3508 / 22.21x | 0.6049 / 12.88x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.9260 | n/a | 0.0959 / 40.94x | 0.1517 / 25.88x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4540 | n/a | 0.0262 / 55.50x | 0.0496 / 29.31x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3360 | n/a | 0.0135 / 24.89x | 0.0416 / 8.08x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.5280 | n/a | 0.0164 / 32.20x | 0.0472 / 11.19x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.3260 | n/a | 0.0169 / 196.80x | 0.0431 / 77.17x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.1260 | n/a | 0.0177 / 233.11x | 0.0399 / 103.41x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.8760 | n/a | 0.0186 / 208.39x | 0.0471 / 82.29x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.8260 | n/a | 0.0296 / 27.91x | 0.0574 / 14.39x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.6440 | n/a | 0.0161 / 40.00x | 0.0459 / 14.03x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.6260 | n/a | 0.0634 / 41.42x | 0.0882 / 29.77x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.9360 | n/a | 0.0181 / 51.71x | 0.0464 / 20.17x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6120 | n/a | 0.2128 / 16.97x | 0.3258 / 11.09x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-01-selected.json` — `custom` (`2026-07-20T16:08:10.421371Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-02-selected.json` — `custom` (`2026-07-20T16:10:33.612132Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-03-selected.json` — `custom` (`2026-07-20T16:13:36.370063Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-04-selected.json` — `custom` (`2026-07-20T16:15:57.569321Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-05-selected.json` — `custom` (`2026-07-20T16:19:17.709504Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-06-selected.json` — `custom` (`2026-07-20T16:21:03.873861Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-compiled-07-selected.json` — `custom` (`2026-07-20T16:21:59.748520Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-01-status.json` — `custom` (`2026-07-20T16:23:09.442785Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-02-status.json` — `custom` (`2026-07-20T16:25:11.522631Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T16:25:50.895685Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T16:26:53.873158Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T16:30:50.108578Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-06-status.json` — `custom` (`2026-07-20T16:32:52.131265Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T16:33:47.714813Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-async-compiled-01-selected.json` — `custom` (`2026-07-20T16:34:58.551304Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-async-compiled-02-selected.json` — `custom` (`2026-07-20T16:36:08.658949Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-async-bytecode-01-selected.json` — `custom` (`2026-07-20T16:36:21.611554Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-async-bytecode-02-selected.json` — `custom` (`2026-07-20T16:36:36.923279Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T16:38:58.073170Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T16:44:34.146352Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T16:49:35.538786Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T16:51:57.825491Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T16:53:10.628100Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T16:53:38.407335Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T16:54:30.053213Z`)
- `v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T16:55:19.324456Z`)

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
