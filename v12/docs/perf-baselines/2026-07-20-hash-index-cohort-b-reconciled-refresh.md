# External Application Scoreboard

- Source measurements through: `2026-07-20T15:02:20.374405Z`
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
| `fib` | `compiled` | `verified` | 3.5720 | 3.3169 / 1.08x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.7020 | 10.6741 / 0.91x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1840 | 1.0726 / 1.10x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.8280 | 2.7404 / 0.67x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.7960 | 0.6017 / 2.98x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1180 | 0.0656 / 1.80x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5900 | 2.5744 / 1.01x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.8940 | 1.4453 / 0.62x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2160 | 0.1999 / 1.08x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.5520 | 1.1930 / 1.30x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1480 | 0.0520 / 2.85x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1080 | 0.0153 / 7.06x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.6020 | 0.0626 / 57.54x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1140 | 0.0141 / 8.09x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1840 | 0.0341 / 5.40x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.0760 | 1.9146 / 2.13x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1100 | 0.0187 / 5.88x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1360 | 0.0125 / 10.88x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2200 | n/a | n/a | 45.1513 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.5900 | n/a | 54.7614 / 0.08x | 46.6638 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 27.2326 / n/a | 17.1609 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.7748 / n/a | 20.5753 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6500 | n/a | 0.0977 / 6.65x | 0.1391 / 4.67x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.4780 | n/a | 3.9066 / 0.89x | 2.6873 / 1.29x | `miss` | — |
| `json` | `bytecode` | `verified` | 1.0780 | n/a | 2.6641 / 0.40x | 1.6810 / 0.64x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 3.1400 | n/a | 1.4468 / 2.17x | 1.5691 / 2.00x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 3.1460 | n/a | 4.1200 / 0.76x | 10.4571 / 0.30x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 10.2080 | n/a | 1.2365 / 8.26x | 1.9086 / 5.35x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.1540 | n/a | 0.0238 / 132.52x | 0.0685 / 46.04x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 39.7380 | n/a | 1.3274 / 29.94x | 1.3516 / 29.40x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.6040 | n/a | 0.2040 / 7.86x | 0.2849 / 5.63x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9489 / n/a | 3.0733 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.8420 | n/a | 0.5416 / 10.79x | 0.3241 / 18.03x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 5.3360 | n/a | 0.8145 / 6.55x | 0.5121 / 10.42x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.6260 | 0.0057 / 109.82x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3880 | 0.0054 / 71.85x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0940 | 0.0039 / 24.10x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3460 | 0.0048 / 72.08x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8060 | 0.0047 / 171.49x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7200 | 0.0039 / 184.62x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5440 | n/a | 0.0457 / 11.90x | 0.0720 / 7.56x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4680 | n/a | 0.0619 / 7.56x | 0.0797 / 5.87x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1640 | n/a | 0.0305 / 5.38x | 0.0606 / 2.71x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2440 | n/a | 0.1285 / 1.90x | 0.1003 / 2.43x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3580 | n/a | 0.0304 / 11.78x | 0.0581 / 6.16x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2100 | n/a | 0.0241 / 8.71x | 0.0448 / 4.69x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2460 | 0.0056 / 43.93x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1460 | 0.0132 / 11.06x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1580 | 0.0051 / 30.98x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1000 | 0.0038 / 26.32x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1180 | 0.0041 / 28.78x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1640 | 0.0047 / 34.89x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1220 | 0.0050 / 24.40x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1120 | 0.0065 / 17.23x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0800 | 0.0050 / 16.00x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0900 | 0.0045 / 20.00x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2580 | 0.0106 / 24.34x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2000 | 0.0042 / 47.62x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2420 | 0.0123 / 19.67x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.7960 | n/a | 0.3568 / 21.85x | 0.6178 / 12.62x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.9080 | n/a | 0.1069 / 36.56x | 0.1328 / 29.43x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.3800 | n/a | 0.0189 / 73.02x | 0.0521 / 26.49x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3300 | n/a | 0.0133 / 24.81x | 0.0397 / 8.31x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4740 | n/a | 0.0160 / 29.62x | 0.0441 / 10.75x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.2920 | n/a | 0.0178 / 184.94x | 0.0437 / 75.33x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 3.9900 | n/a | 0.0200 / 199.50x | 0.0477 / 83.65x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.4660 | n/a | 0.0196 / 176.84x | 0.0456 / 76.01x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7120 | n/a | 0.0302 / 23.58x | 0.0591 / 12.05x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5620 | n/a | 0.0222 / 25.32x | 0.0536 / 10.49x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.5100 | n/a | 0.0677 / 37.08x | 0.0800 / 31.37x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8700 | n/a | 0.0158 / 55.06x | 0.0426 / 20.42x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6020 | n/a | 0.2150 / 16.75x | 0.3141 / 11.47x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-01-selected.json` — `custom` (`2026-07-20T14:07:38.515185Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-02-selected.json` — `custom` (`2026-07-20T14:10:08.080363Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-03-selected.json` — `custom` (`2026-07-20T14:13:24.611366Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-04-selected.json` — `custom` (`2026-07-20T14:16:23.194260Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-05-selected.json` — `custom` (`2026-07-20T14:20:24.369135Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-06-selected.json` — `custom` (`2026-07-20T14:22:27.992448Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-compiled-07-selected.json` — `custom` (`2026-07-20T14:23:38.746916Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-01-status.json` — `custom` (`2026-07-20T14:24:46.377199Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-02-status.json` — `custom` (`2026-07-20T14:26:41.016231Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T14:27:29.266708Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T14:29:02.990878Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-05-replacement.json` — `custom` (`2026-07-20T15:02:20.374405Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-06-status.json` — `custom` (`2026-07-20T14:35:40.243623Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T14:36:41.332606Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-async-compiled-01-selected.json` — `custom` (`2026-07-20T14:38:01.812799Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-async-compiled-02-selected.json` — `custom` (`2026-07-20T14:38:54.028575Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-async-bytecode-01-selected.json` — `custom` (`2026-07-20T14:39:07.002951Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-async-bytecode-02-selected.json` — `custom` (`2026-07-20T14:39:18.242705Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T14:41:40.248878Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T14:47:07.906031Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T14:52:42.076672Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T14:55:01.373836Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T14:56:13.647550Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T14:56:41.074531Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T14:57:28.797455Z`)
- `v12/docs/perf-baselines/2026-07-20-hash-index-cohort-b-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T14:58:16.131000Z`)

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
