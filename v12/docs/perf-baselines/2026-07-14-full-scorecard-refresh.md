# External Application Scoreboard

- Source measurements through: `2026-07-14T17:04:27.472609Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: CPU-pinned, one Go process, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and the source scorecard's per-process timeout.
- Compiled: `4/21` rankable rows meet the 95%-of-Go target.
- Bytecode: `2/14` rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- `unranked` means a timeout, failed run, or unavailable matched foreign reference; it is never counted as a pass or fail.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target |
| --- | --- | --- | ---: | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.5500 | 3.1392 / 1.13x | n/a | n/a | `miss` |
| `binarytrees` | `compiled` | `verified` | 30.2700 | 33.3653 / 0.91x | n/a | n/a | `meets` |
| `matrixmultiply` | `compiled` | `verified` | 1.2100 | 0.9940 / 1.22x | n/a | n/a | `miss` |
| `quicksort` | `compiled` | `verified` | 1.9300 | 2.6809 / 0.72x | n/a | n/a | `meets` |
| `sudoku` | `compiled` | `timeout` | n/a | 0.1534 / n/a | n/a | n/a | `unranked` |
| `sudoku_masks` | `compiled` | `verified` | 9.1100 | 0.5779 / 15.76x | n/a | n/a | `miss` |
| `i_before_e` | `compiled` | `verified` | 0.1800 | 0.0682 / 2.64x | n/a | n/a | `miss` |
| `base64` | `compiled` | `verified` | 2.5000 | 2.5024 / 1.00x | n/a | n/a | `meets` |
| `json` | `compiled` | `verified` | 0.8000 | 1.4478 / 0.55x | n/a | n/a | `meets` |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2800 | 0.2063 / 1.36x | n/a | n/a | `miss` |
| `pidigits` | `compiled` | `verified` | 1.4000 | 1.2090 / 1.16x | n/a | n/a | `miss` |
| `mandelbrot` | `compiled` | `verified` | 0.2300 | 0.0493 / 4.67x | n/a | n/a | `miss` |
| `reverse_complement` | `compiled` | `verified` | 0.1800 | 0.0155 / 11.61x | n/a | n/a | `miss` |
| `k_nucleotide` | `compiled` | `verified` | 3.5800 | 0.0602 / 59.47x | n/a | n/a | `miss` |
| `nbody` | `compiled` | `verified` | 0.4800 | 0.0329 / 14.59x | n/a | n/a | `miss` |
| `tapelang_alphabet` | `compiled` | `verified` | 3.7900 | 1.8370 / 2.06x | n/a | n/a | `miss` |
| `fib` | `bytecode` | `verified` | 0.2300 | n/a | n/a | 44.8618 / 0.01x | `unranked` |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` |
| `matrixmultiply` | `bytecode` | `verified` | 4.4600 | n/a | n/a | n/a | `unranked` |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 30.0546 / n/a | 16.0435 / n/a | `unranked` |
| `sudoku` | `bytecode` | `timeout` | n/a | n/a | 3.0391 / n/a | 6.3565 / n/a | `unranked` |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 23.8410 / n/a | 26.7593 / n/a | `unranked` |
| `i_before_e` | `bytecode` | `verified` | 0.6200 | n/a | 0.0939 / 6.60x | 0.1464 / 4.23x | `miss` |
| `base64` | `bytecode` | `verified` | 3.4900 | n/a | 4.4688 / 0.78x | 2.8551 / 1.22x | `miss` |
| `json` | `bytecode` | `verified` | 1.4400 | n/a | 2.9429 / 0.49x | 1.9970 / 0.72x | `meets` |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.6700 | n/a | 1.8059 / 1.48x | 1.8591 / 1.44x | `miss` |
| `pidigits` | `bytecode` | `verified` | 2.8400 | n/a | 4.6003 / 0.62x | 11.9968 / 0.24x | `meets` |
| `mandelbrot` | `bytecode` | `verified` | 7.0200 | n/a | 1.4024 / 5.01x | 2.2404 / 3.13x | `miss` |
| `reverse_complement` | `bytecode` | `verified` | 7.0900 | n/a | 0.0291 / 243.64x | 0.0999 / 70.97x | `miss` |
| `k_nucleotide` | `bytecode` | `verified` | 42.3800 | n/a | 1.6141 / 26.26x | 1.4842 / 28.55x | `miss` |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.4146 / n/a | 3.7831 / n/a | `unranked` |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` |
| `channel_rollup` | `compiled` | `verified` | 1.3400 | 0.0052 / 257.69x | n/a | n/a | `miss` |
| `future_pipeline` | `compiled` | `verified` | 0.7100 | 0.0049 / 144.90x | n/a | n/a | `miss` |
| `future_await_race` | `compiled` | `verified` | 0.1200 | 0.0038 / 31.58x | n/a | n/a | `miss` |
| `await_channel_mux` | `compiled` | `verified` | 0.4600 | 0.0047 / 97.87x | n/a | n/a | `miss` |
| `mutex_ledger` | `compiled` | `verified` | 0.5400 | 0.0043 / 125.58x | n/a | n/a | `miss` |
| `mutex_await_journal` | `compiled` | `verified` | 0.4500 | 0.0035 / 128.57x | n/a | n/a | `miss` |
| `channel_rollup` | `bytecode` | `verified` | 0.6500 | n/a | 0.0398 / 16.33x | 0.0556 / 11.69x | `miss` |
| `future_pipeline` | `bytecode` | `verified` | 0.4800 | n/a | 0.0660 / 7.27x | 0.0695 / 6.91x | `miss` |
| `future_await_race` | `bytecode` | `verified` | 0.1600 | n/a | 0.0294 / 5.44x | 0.0549 / 2.91x | `miss` |
| `await_channel_mux` | `bytecode` | `verified` | 0.2800 | n/a | 0.1128 / 2.48x | 0.0956 / 2.93x | `miss` |
| `mutex_ledger` | `bytecode` | `verified` | 0.6900 | n/a | 0.0357 / 19.33x | 0.0502 / 13.75x | `miss` |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2600 | n/a | 0.0190 / 13.68x | 0.0514 / 5.06x | `miss` |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-01.json` — `custom` (`2026-07-14T16:34:39.676550Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-02.json` — `custom` (`2026-07-14T16:39:20.904267Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-03.json` — `custom` (`2026-07-14T16:41:57.900561Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-04.json` — `custom` (`2026-07-14T16:44:16.335060Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-05.json` — `custom` (`2026-07-14T16:46:30.814037Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-compiled-06.json` — `custom` (`2026-07-14T16:48:26.959325Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-01.json` — `custom` (`2026-07-14T16:49:51.581843Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-02.json` — `custom` (`2026-07-14T16:53:04.301919Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-03.json` — `custom` (`2026-07-14T16:54:13.723738Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-04.json` — `custom` (`2026-07-14T16:55:24.211801Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-05.json` — `custom` (`2026-07-14T16:57:46.160428Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-generality-bytecode-06.json` — `custom` (`2026-07-14T16:59:15.204431Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-async-compiled-01.json` — `custom` (`2026-07-14T17:01:46.912855Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-async-compiled-02.json` — `custom` (`2026-07-14T17:03:41.395818Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-async-bytecode-01.json` — `custom` (`2026-07-14T17:04:08.025465Z`)
- `v12/docs/perf-baselines/2026-07-14-full-scorecard-async-bytecode-02.json` — `custom` (`2026-07-14T17:04:27.472609Z`)

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
