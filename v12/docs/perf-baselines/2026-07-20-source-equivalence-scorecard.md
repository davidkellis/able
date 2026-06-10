# External Application Scoreboard

- Source measurements through: `2026-07-20T10:47:37.834088Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `6/36` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/29` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `72` row fingerprints in JSON; `72` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `72` row fingerprints in JSON; `72` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `65` reviewed benchmark/mode rows, SHA-256 `e7b35985b05134e1619be193cbe21ddce846cc2392efe78560e629de048d97dc`; timeout rows remain in full status.
- Matched reference source fingerprints: `104` comparison fingerprints in JSON; `104` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.9760 | 3.2577 / 1.22x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.6600 | 10.6970 / 0.90x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1040 | 1.0892 / 1.01x | n/a | n/a | `meets` | — |
| `quicksort` | `compiled` | `verified` | 1.9620 | 2.6338 / 0.74x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.9240 | 0.5646 / 3.41x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1440 | 0.0580 / 2.48x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5360 | 2.4924 / 1.02x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7580 | 1.4731 / 0.51x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2040 | 0.2635 / 0.77x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3380 | 1.2323 / 1.09x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1640 | 0.0484 / 3.39x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1100 | 0.0167 / 6.59x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.5800 | 0.0532 / 67.29x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1200 | 0.0138 / 8.70x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1820 | 0.0358 / 5.08x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.9240 | 2.1690 / 1.81x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1080 | 0.0128 / 8.44x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0900 | 0.0106 / 8.49x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1500 | n/a | n/a | 49.4996 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | 54.2400 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.9200 | n/a | 48.5912 / 0.10x | 44.8640 / 0.11x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.2644 / n/a | 15.5419 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.2731 / n/a | 21.4984 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6320 | n/a | 0.0907 / 6.97x | 0.1231 / 5.13x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.0240 | n/a | 3.9152 / 0.77x | 2.4455 / 1.24x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.9100 | n/a | 2.8382 / 0.32x | 1.7117 / 0.53x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.7740 | n/a | 1.5116 / 1.84x | 1.6105 / 1.72x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.8540 | n/a | 4.1766 / 0.68x | 10.4766 / 0.27x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.6800 | n/a | 1.2087 / 5.53x | 1.8993 / 3.52x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.3340 | n/a | 0.0258 / 129.22x | 0.0786 / 42.42x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 43.5120 | n/a | 1.3955 / 31.18x | 1.3702 / 31.76x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.7760 | n/a | 0.2294 / 7.74x | 0.2957 / 6.01x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9948 / n/a | 2.9785 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.8300 | n/a | 0.5990 / 9.73x | 0.3381 / 17.24x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.7660 | n/a | 0.8990 / 5.30x | 0.6989 / 6.82x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5920 | 0.0062 / 95.48x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3580 | 0.0054 / 66.30x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1220 | 0.0043 / 28.37x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3540 | 0.0050 / 70.80x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8720 | 0.0046 / 189.57x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.8360 | 0.0040 / 209.00x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.6780 | n/a | 0.0416 / 16.30x | 0.0546 / 12.42x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4300 | n/a | 0.0605 / 7.11x | 0.0722 / 5.96x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1900 | n/a | 0.0300 / 6.33x | 0.0520 / 3.65x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.3660 | n/a | 0.1148 / 3.19x | 0.0956 / 3.83x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4900 | n/a | 0.0326 / 15.03x | 0.0535 / 9.16x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2400 | n/a | 0.0199 / 12.06x | 0.0454 / 5.29x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2300 | 0.0070 / 32.86x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1540 | 0.0152 / 10.13x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.2720 | 0.0069 / 39.42x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1040 | 0.0049 / 21.22x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1060 | 0.0054 / 19.63x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1260 | 0.0060 / 21.00x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1100 | 0.0056 / 19.64x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1700 | 0.0061 / 27.87x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0900 | 0.0044 / 20.45x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1100 | 0.0046 / 23.91x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2260 | 0.0042 / 53.81x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.3620 | 0.0120 / 30.17x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.1840 | n/a | 0.4208 / 19.45x | 0.6594 / 12.41x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.2200 | n/a | 0.1319 / 31.99x | 0.1864 / 22.64x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5080 | n/a | 0.0215 / 70.14x | 0.0547 / 27.57x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3380 | n/a | 0.0152 / 22.24x | 0.0438 / 7.72x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4800 | n/a | 0.0174 / 27.59x | 0.0511 / 9.39x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.7840 | n/a | 0.0191 / 198.12x | 0.0408 / 92.75x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.2680 | n/a | 0.0185 / 230.70x | 0.0424 / 100.66x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.7000 | n/a | 0.0180 / 205.56x | 0.0441 / 83.90x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7140 | n/a | 0.0306 / 23.33x | 0.0713 / 10.01x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4760 | n/a | 0.0210 / 22.67x | 0.0560 / 8.50x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8880 | n/a | 0.0197 / 45.08x | 0.0536 / 16.57x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6440 | n/a | 0.2807 / 12.98x | 0.3169 / 11.50x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-01-selected.json` — `custom` (`2026-07-20T10:00:17.378592Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-02-selected.json` — `custom` (`2026-07-20T10:02:44.262805Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-03-selected.json` — `custom` (`2026-07-20T10:05:50.356345Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-04-selected.json` — `custom` (`2026-07-20T10:08:19.348945Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-05-selected.json` — `custom` (`2026-07-20T10:11:55.988959Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-06-selected.json` — `custom` (`2026-07-20T10:13:55.485638Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-compiled-07-selected.json` — `custom` (`2026-07-20T10:14:56.742154Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-01-status.json` — `custom` (`2026-07-20T10:16:04.242012Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-02-status.json` — `custom` (`2026-07-20T10:17:58.593700Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T10:18:41.418699Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T10:19:50.371998Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T10:24:01.113922Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-06-status.json` — `custom` (`2026-07-20T10:25:55.636219Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T10:26:53.512446Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-async-compiled-01-selected.json` — `custom` (`2026-07-20T10:28:09.431606Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-async-compiled-02-selected.json` — `custom` (`2026-07-20T10:29:01.242241Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-async-bytecode-01-selected.json` — `custom` (`2026-07-20T10:29:15.287882Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-async-bytecode-02-selected.json` — `custom` (`2026-07-20T10:29:28.588542Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T10:31:50.990414Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T10:37:20.962638Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T10:42:29.224228Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T10:44:28.057494Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T10:45:44.714426Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T10:46:14.894912Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T10:47:05.671895Z`)
- `v12/docs/perf-baselines/2026-07-20-current-full-scorecard-cohort-a-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T10:47:37.834088Z`)

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
