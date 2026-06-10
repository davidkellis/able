# External Application Scoreboard

- Source measurements through: `2026-07-20T20:15:15.358677Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `4/40` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `2/33` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `80` row fingerprints in JSON; `80` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `80` row fingerprints in JSON; `80` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `73` reviewed benchmark/mode rows, SHA-256 `c09a947ee6ab58da00c9a257b558268f5851ce1020efc7b2349f47636068c4f1`; timeout rows remain in full status.
- Matched reference source fingerprints: `117` comparison fingerprints in JSON; `117` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.5480 | 3.1205 / 1.14x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.5960 | 10.9347 / 0.88x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1700 | 1.0536 / 1.11x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.9280 | 2.6012 / 0.74x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.9880 | 0.6045 / 3.29x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.2000 | 0.0605 / 3.31x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5400 | 2.5724 / 0.99x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.7780 | 1.4403 / 0.54x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2620 | 0.2242 / 1.17x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.4120 | 1.2076 / 1.17x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1420 | 0.0517 / 2.75x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1080 | 0.0161 / 6.71x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.0280 | 0.0546 / 55.46x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1220 | 0.0138 / 8.84x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1880 | 0.0321 / 5.86x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.9840 | 1.8448 / 2.16x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1240 | 0.0118 / 10.51x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.1320 | 0.0113 / 11.68x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1500 | n/a | n/a | 44.2771 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | 53.1704 / n/a | 54.1106 / n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.9300 | n/a | 48.1957 / 0.10x | 45.2379 / 0.11x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 25.3076 / n/a | 15.2717 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 17.8622 / n/a | 21.2515 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.6400 | n/a | 0.1061 / 6.03x | 0.1269 / 5.04x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.1860 | n/a | 3.8755 / 0.82x | 2.4556 / 1.30x | `miss` | — |
| `json` | `bytecode` | `verified` | 1.0260 | n/a | 2.6028 / 0.39x | 1.6835 / 0.61x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 3.0760 | n/a | 1.4786 / 2.08x | 1.5627 / 1.97x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.5540 | n/a | 4.0382 / 0.63x | 9.8656 / 0.26x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 7.2300 | n/a | 1.1291 / 6.40x | 1.7638 / 4.10x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.8100 | n/a | 0.0268 / 142.16x | 0.0737 / 51.70x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 48.4720 | n/a | 1.2627 / 38.39x | 1.2578 / 38.54x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 2.0640 | n/a | 0.2063 / 10.00x | 0.3126 / 6.60x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.0151 / n/a | 3.1860 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.9740 | n/a | 0.5378 / 11.11x | 0.3312 / 18.04x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.8040 | n/a | 0.8392 / 5.72x | 0.5291 / 9.08x | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.6320 | 0.0062 / 101.94x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.4100 | 0.0059 / 69.49x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0038 / 31.58x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3820 | 0.0048 / 79.58x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8040 | 0.0046 / 174.78x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.7780 | 0.0044 / 176.82x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.9160 | 0.0048 / 399.17x | n/a | n/a | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5580 | n/a | 0.0415 / 13.45x | 0.0548 / 10.18x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4820 | n/a | 0.0582 / 8.28x | 0.0771 / 6.25x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1680 | n/a | 0.0360 / 4.67x | 0.0611 / 2.75x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2240 | n/a | 0.1204 / 1.86x | 0.1052 / 2.13x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4480 | n/a | 0.0382 / 11.73x | 0.0549 / 8.16x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2400 | n/a | 0.0208 / 11.54x | 0.0445 / 5.39x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.4000 | n/a | 0.0306 / 13.07x | 0.0698 / 5.73x | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2180 | 0.0058 / 37.59x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1620 | 0.0132 / 12.27x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.1940 | 0.0261 / 7.43x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1760 | 0.0052 / 33.85x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0940 | 0.0040 / 23.50x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1060 | 0.0041 / 25.85x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1100 | 0.0044 / 25.00x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1120 | 0.0050 / 22.40x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1160 | 0.0045 / 25.78x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.1000 | 0.0043 / 23.26x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0860 | 0.0041 / 20.98x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0800 | 0.0038 / 21.05x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2740 | 0.0081 / 33.83x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2060 | 0.0041 / 50.24x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2360 | 0.0117 / 20.17x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.9280 | n/a | 0.3406 / 23.28x | 0.6307 / 12.57x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 3.9300 | n/a | 0.1025 / 38.34x | 0.1254 / 31.34x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.4140 | n/a | 0.0603 / 89.78x | 0.1411 / 38.37x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4580 | n/a | 0.0202 / 72.18x | 0.0569 / 25.62x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3860 | n/a | 0.0149 / 25.91x | 0.0442 / 8.73x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4680 | n/a | 0.0177 / 26.44x | 0.0526 / 8.90x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.5200 | n/a | 0.0233 / 151.07x | 0.0524 / 67.18x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.3480 | n/a | 0.0222 / 195.86x | 0.0458 / 94.93x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.7880 | n/a | 0.0239 / 158.49x | 0.0500 / 75.76x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 2.9580 | n/a | 0.0181 / 163.43x | 0.0422 / 70.09x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6640 | n/a | 0.0299 / 22.21x | 0.0655 / 10.14x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5160 | n/a | 0.0174 / 29.66x | 0.0515 / 10.02x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.6400 | n/a | 0.0676 / 39.05x | 0.0910 / 29.01x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.9880 | n/a | 0.0163 / 60.61x | 0.0466 / 21.20x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6500 | n/a | 0.2347 / 15.55x | 0.3710 / 9.84x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-01-selected.json` — `custom` (`2026-07-20T19:22:14.622829Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-02-selected.json` — `custom` (`2026-07-20T19:24:44.789822Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-03-selected.json` — `custom` (`2026-07-20T19:27:52.178836Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-04-selected.json` — `custom` (`2026-07-20T19:30:21.263624Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-05-selected.json` — `custom` (`2026-07-20T19:34:01.278518Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-06-selected.json` — `custom` (`2026-07-20T19:36:07.963771Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-compiled-07-selected.json` — `custom` (`2026-07-20T19:37:12.895297Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-01-status.json` — `custom` (`2026-07-20T19:38:20.029311Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-02-status.json` — `custom` (`2026-07-20T19:40:14.692162Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-03-selected.json` — `custom` (`2026-07-20T19:40:59.571019Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-04-selected.json` — `custom` (`2026-07-20T19:42:11.512156Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-05-selected.json` — `custom` (`2026-07-20T19:46:51.160843Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-06-status.json` — `custom` (`2026-07-20T19:48:45.382373Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-generality-bytecode-07-selected.json` — `custom` (`2026-07-20T19:49:43.886376Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-async-compiled-01-selected.json` — `custom` (`2026-07-20T19:51:02.713279Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-async-compiled-02-selected.json` — `custom` (`2026-07-20T19:52:20.076500Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-async-bytecode-01-selected.json` — `custom` (`2026-07-20T19:52:33.741945Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-async-bytecode-02-selected.json` — `custom` (`2026-07-20T19:52:50.223047Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-20T19:56:05.525419Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-20T20:01:23.392016Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-20T20:08:39.997087Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-20T20:11:06.616236Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-20T20:12:49.409670Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-20T20:13:18.332910Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-20T20:14:26.720817Z`)
- `v12/docs/perf-baselines/2026-07-20-log-routing-redaction-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-20T20:15:15.358677Z`)

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
