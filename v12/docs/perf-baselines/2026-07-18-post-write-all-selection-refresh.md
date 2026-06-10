# External Application Scoreboard

- Source measurements through: `2026-07-19T03:05:10.585553Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `6/36` selected rankable rows meet the 95%-of-Go target.
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
| `fib` | `compiled` | `verified` | 3.3680 | 3.7407 / 0.90x | n/a | n/a | `meets` | — |
| `binarytrees` | `compiled` | `verified` | 10.3240 | 11.0044 / 0.94x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.2140 | 0.9964 / 1.22x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 2.1020 | 2.7133 / 0.77x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 9.4240 | 0.5971 / 15.78x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1180 | 0.0642 / 1.84x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5440 | 2.4856 / 1.02x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.8880 | 1.5514 / 0.57x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2080 | 0.2150 / 0.97x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.5560 | 1.1622 / 1.34x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1440 | 0.0502 / 2.87x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1020 | 0.0219 / 4.66x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.5720 | 0.0786 / 45.45x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1080 | 0.0147 / 7.35x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1700 | 0.0564 / 3.01x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.8820 | 2.0734 / 1.87x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1040 | 0.0131 / 7.94x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0880 | 0.0107 / 8.22x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1700 | n/a | n/a | 44.2791 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.6000 | n/a | 51.5769 / 0.09x | 47.2458 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.6492 / n/a | 15.8552 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.6339 / n/a | 23.1227 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5500 | n/a | 0.1265 / 4.35x | 0.1231 / 4.47x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9500 | n/a | 3.9161 / 0.75x | 2.4969 / 1.18x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8220 | n/a | 2.6856 / 0.31x | 1.8851 / 0.44x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.7780 | n/a | 1.5207 / 1.83x | 1.5319 / 1.81x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.6340 | n/a | 4.1174 / 0.64x | 10.4254 / 0.25x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.7400 | n/a | 1.2431 / 5.42x | 1.9676 / 3.43x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 4.7740 | n/a | 0.0278 / 171.73x | 0.0749 / 63.74x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 43.1140 | n/a | 1.3609 / 31.68x | 1.2740 / 33.84x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 2.1020 | n/a | 0.2060 / 10.20x | 0.2891 / 7.27x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0046 / n/a | 3.2710 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 6.0120 | n/a | 0.5966 / 10.08x | 0.3611 / 16.65x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.7800 | n/a | 0.9586 / 4.99x | 0.5322 / 8.98x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5560 | 0.0059 / 94.24x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3760 | 0.0054 / 69.63x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1020 | 0.0042 / 24.29x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3800 | 0.0051 / 74.51x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.9020 | 0.0048 / 187.92x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.8880 | 0.0045 / 197.33x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5460 | n/a | 0.0434 / 12.58x | 0.0546 / 10.00x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.6320 | n/a | 0.0646 / 9.78x | 0.0762 / 8.29x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.2040 | n/a | 0.0343 / 5.95x | 0.0572 / 3.57x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2160 | n/a | 0.1181 / 1.83x | 0.0946 / 2.28x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4280 | n/a | 0.0357 / 11.99x | 0.0599 / 7.15x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2460 | n/a | 0.0205 / 12.00x | 0.0467 / 5.27x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2540 | 0.0077 / 32.99x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1520 | 0.0150 / 10.13x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2740 | 0.0058 / 47.24x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0980 | 0.0046 / 21.30x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1100 | 0.0050 / 22.00x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 1.0720 | 0.0349 / 30.72x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1520 | 0.0049 / 31.02x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1400 | 0.0052 / 26.92x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1240 | 0.0046 / 26.96x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1340 | 0.0036 / 37.22x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2600 | 0.0039 / 66.67x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2980 | 0.0128 / 23.28x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 9.0920 | n/a | 0.4874 / 18.65x | 0.7503 / 12.12x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.3000 | n/a | 0.1126 / 38.19x | 0.1457 / 29.51x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5820 | n/a | 0.0211 / 74.98x | 0.0643 / 24.60x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3540 | n/a | 0.0155 / 22.84x | 0.0465 / 7.61x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.5600 | n/a | 0.0180 / 31.11x | 0.0543 / 10.31x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `timeout` | n/a | n/a | 0.0422 / n/a | 0.0842 / n/a | `unranked` | Able timed out |
| `regex_set_audit` | `bytecode` | `verified` | 4.4540 | n/a | 0.0210 / 212.10x | 0.0477 / 93.38x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.8060 | n/a | 0.0205 / 185.66x | 0.0667 / 57.06x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6840 | n/a | 0.0316 / 21.65x | 0.0859 / 7.96x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5320 | n/a | 0.0155 / 34.32x | 0.0567 / 9.38x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 1.1260 | n/a | 0.0255 / 44.16x | 0.0534 / 21.09x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6760 | n/a | 0.2315 / 15.88x | 0.3270 / 11.24x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-01-selected.json` — `custom` (`2026-07-19T02:16:02.565680Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-02-selected.json` — `custom` (`2026-07-19T02:19:14.849615Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-03-selected.json` — `custom` (`2026-07-19T02:22:25.159436Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-04-selected.json` — `custom` (`2026-07-19T02:24:53.182981Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-05-selected.json` — `custom` (`2026-07-19T02:28:25.730778Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-06-selected.json` — `custom` (`2026-07-19T02:30:24.148153Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-compiled-07-selected.json` — `custom` (`2026-07-19T02:31:23.490700Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-01-status.json` — `custom` (`2026-07-19T02:32:29.945952Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-02-status.json` — `custom` (`2026-07-19T02:34:24.173149Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-03-selected.json` — `custom` (`2026-07-19T02:35:04.534369Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-04-selected.json` — `custom` (`2026-07-19T02:36:12.748702Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-05-selected.json` — `custom` (`2026-07-19T02:40:30.262217Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-06-status.json` — `custom` (`2026-07-19T02:42:24.516630Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-generality-bytecode-07-selected.json` — `custom` (`2026-07-19T02:43:23.461298Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-async-compiled-01-selected.json` — `custom` (`2026-07-19T02:44:40.854124Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-async-compiled-02-selected.json` — `custom` (`2026-07-19T02:45:34.691981Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-async-bytecode-01-selected.json` — `custom` (`2026-07-19T02:45:48.859837Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-async-bytecode-02-selected.json` — `custom` (`2026-07-19T02:46:00.562726Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-19T02:48:24.219392Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-19T02:54:04.388850Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-19T02:59:15.496338Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-19T03:01:16.255811Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-19T03:02:38.462821Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-19T03:02:47.903952Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-bytecode-02-status.json` — `custom` (`2026-07-19T03:03:45.088427Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-19T03:04:36.889020Z`)
- `v12/docs/perf-baselines/2026-07-18-post-write-all-selection-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-19T03:05:10.585553Z`)

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
