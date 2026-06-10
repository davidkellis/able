# External Application Scoreboard

- Source measurements through: `2026-07-15T06:39:12.591807Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: CPU-pinned, one Go process, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and the source scorecard's per-process timeout.
- Compiled: `5/21` rankable rows meet the 95%-of-Go target.
- Bytecode: `3/15` rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- `unranked` means a timeout, failed run, or unavailable matched foreign reference; it is never counted as a pass or fail.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target |
| --- | --- | --- | ---: | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.2400 | 2.8923 / 1.12x | n/a | n/a | `miss` |
| `binarytrees` | `compiled` | `verified` | 26.8600 | 29.8219 / 0.90x | n/a | n/a | `meets` |
| `matrixmultiply` | `compiled` | `verified` | 1.1100 | 0.9080 / 1.22x | n/a | n/a | `miss` |
| `quicksort` | `compiled` | `verified` | 1.7500 | 2.3420 / 0.75x | n/a | n/a | `meets` |
| `sudoku` | `compiled` | `timeout` | n/a | 0.1326 / n/a | n/a | n/a | `unranked` |
| `sudoku_masks` | `compiled` | `verified` | 7.7900 | 0.5257 / 14.82x | n/a | n/a | `miss` |
| `i_before_e` | `compiled` | `verified` | 0.1300 | 0.0568 / 2.29x | n/a | n/a | `miss` |
| `base64` | `compiled` | `verified` | 2.2300 | 2.2743 / 0.98x | n/a | n/a | `meets` |
| `json` | `compiled` | `verified` | 0.7400 | 1.3256 / 0.56x | n/a | n/a | `meets` |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1900 | 0.1890 / 1.01x | n/a | n/a | `meets` |
| `pidigits` | `compiled` | `verified` | 1.2700 | 1.0658 / 1.19x | n/a | n/a | `miss` |
| `mandelbrot` | `compiled` | `verified` | 0.1800 | 0.0466 / 3.86x | n/a | n/a | `miss` |
| `reverse_complement` | `compiled` | `verified` | 0.1500 | 0.0157 / 9.55x | n/a | n/a | `miss` |
| `k_nucleotide` | `compiled` | `verified` | 3.1200 | 0.0526 / 59.32x | n/a | n/a | `miss` |
| `nbody` | `compiled` | `verified` | 0.3700 | 0.0303 / 12.21x | n/a | n/a | `miss` |
| `tapelang_alphabet` | `compiled` | `verified` | 3.4300 | 1.7337 / 1.98x | n/a | n/a | `miss` |
| `fib` | `bytecode` | `verified` | 0.1400 | n/a | n/a | 39.8737 / 0.00x | `unranked` |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` |
| `matrixmultiply` | `bytecode` | `verified` | 3.8100 | n/a | 43.4082 / 0.09x | 41.0455 / 0.09x | `meets` |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 21.1651 / n/a | 13.4884 / n/a | `unranked` |
| `sudoku` | `bytecode` | `timeout` | n/a | n/a | 2.7193 / n/a | 5.3886 / n/a | `unranked` |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 15.5988 / n/a | 21.9489 / n/a | `unranked` |
| `i_before_e` | `bytecode` | `verified` | 0.4900 | n/a | 0.1025 / 4.78x | 0.1407 / 3.48x | `miss` |
| `base64` | `bytecode` | `verified` | 2.6800 | n/a | 3.8567 / 0.69x | 2.2977 / 1.17x | `miss` |
| `json` | `bytecode` | `verified` | 0.7200 | n/a | 2.5651 / 0.28x | 2.6154 / 0.28x | `meets` |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.2100 | n/a | 1.4420 / 1.53x | 3.3516 / 0.66x | `miss` |
| `pidigits` | `bytecode` | `verified` | 2.0600 | n/a | 4.9168 / 0.42x | 10.9244 / 0.19x | `meets` |
| `mandelbrot` | `bytecode` | `verified` | 5.6400 | n/a | 1.2363 / 4.56x | 2.0226 / 2.79x | `miss` |
| `reverse_complement` | `bytecode` | `verified` | 5.9600 | n/a | 0.0270 / 220.74x | 0.0766 / 77.81x | `miss` |
| `k_nucleotide` | `bytecode` | `verified` | 36.1800 | n/a | 1.1866 / 30.49x | 1.1236 / 32.20x | `miss` |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.8162 / n/a | 2.8494 / n/a | `unranked` |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` |
| `channel_rollup` | `compiled` | `verified` | 1.1700 | 0.0053 / 220.75x | n/a | n/a | `miss` |
| `future_pipeline` | `compiled` | `verified` | 0.6100 | 0.0047 / 129.79x | n/a | n/a | `miss` |
| `future_await_race` | `compiled` | `verified` | 0.1100 | 0.0035 / 31.43x | n/a | n/a | `miss` |
| `await_channel_mux` | `compiled` | `verified` | 0.3400 | 0.0043 / 79.07x | n/a | n/a | `miss` |
| `mutex_ledger` | `compiled` | `verified` | 0.6100 | 0.0041 / 148.78x | n/a | n/a | `miss` |
| `mutex_await_journal` | `compiled` | `verified` | 0.3900 | 0.0037 / 105.41x | n/a | n/a | `miss` |
| `channel_rollup` | `bytecode` | `verified` | 0.6100 | n/a | 0.0396 / 15.40x | 0.0511 / 11.94x | `miss` |
| `future_pipeline` | `bytecode` | `verified` | 0.4500 | n/a | 0.0569 / 7.91x | 0.0631 / 7.13x | `miss` |
| `future_await_race` | `bytecode` | `verified` | 0.1400 | n/a | 0.0279 / 5.02x | 0.0467 / 3.00x | `miss` |
| `await_channel_mux` | `bytecode` | `verified` | 0.2300 | n/a | 0.1094 / 2.10x | 0.0904 / 2.54x | `miss` |
| `mutex_ledger` | `bytecode` | `verified` | 0.6100 | n/a | 0.0312 / 19.55x | 0.0492 / 12.40x | `miss` |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2400 | n/a | 0.0193 / 12.44x | 0.0402 / 5.97x | `miss` |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-01.json` — `custom` (`2026-07-15T06:18:50.641994Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-02.json` — `custom` (`2026-07-15T06:22:26.637042Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-03.json` — `custom` (`2026-07-15T06:24:47.116043Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-04.json` — `custom` (`2026-07-15T06:26:41.620956Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-05.json` — `custom` (`2026-07-15T06:28:35.099439Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-compiled-06.json` — `custom` (`2026-07-15T06:29:56.641505Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-01.json` — `custom` (`2026-07-15T06:30:52.168902Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-02.json` — `custom` (`2026-07-15T06:33:12.533478Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-03.json` — `custom` (`2026-07-15T06:33:23.856870Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-04.json` — `custom` (`2026-07-15T06:33:39.165310Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-05.json` — `custom` (`2026-07-15T06:34:24.940379Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-bytecode-06.json` — `custom` (`2026-07-15T06:35:58.432761Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-async-compiled-01.json` — `custom` (`2026-07-15T06:37:45.657197Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-async-compiled-02.json` — `custom` (`2026-07-15T06:38:59.473276Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-async-bytecode-01.json` — `custom` (`2026-07-15T06:39:06.063402Z`)
- `v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-async-bytecode-02.json` — `custom` (`2026-07-15T06:39:12.591807Z`)

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
