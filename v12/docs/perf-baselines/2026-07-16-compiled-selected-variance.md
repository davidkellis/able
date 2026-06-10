# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `array_slice_window` | `compiled` | 5 | median=0.0800, mean=0.0940, range=0.0600, CV=27.74% | go: median=0.0036, mean=0.0036, range=0.0002, CV=2.57% | go: median=26.1111, mean=26.1111, range=0.0000, CV=0.00% |
| `await_channel_mux` | `compiled` | 5 | median=0.3500, mean=0.3480, range=0.0200, CV=2.40% | go: median=0.0039, mean=0.0039, range=0.0002, CV=1.99% | go: median=89.2308, mean=89.2308, range=0.0000, CV=0.00% |
| `base64` | `compiled` | 5 | median=2.2800, mean=2.3260, range=0.2800, CV=4.80% | go: median=2.2735, mean=2.2857, range=0.0743, CV=1.31% | go: median=1.0176, mean=1.0176, range=0.0000, CV=0.00% |
| `binarytrees` | `compiled` | 5 | median=27.5400, mean=27.8140, range=2.8000, CV=3.88% | go: median=5.2607, mean=5.2735, range=0.0852, CV=0.63% | go: median=5.2743, mean=5.2743, range=0.0000, CV=0.00% |
| `channel_rollup` | `compiled` | 5 | median=1.2100, mean=1.1940, range=0.0800, CV=3.05% | go: median=0.0047, mean=0.0047, range=0.0003, CV=2.11% | go: median=254.0426, mean=254.0426, range=0.0000, CV=0.00% |
| `dependency_plan` | `compiled` | 5 | median=0.0900, mean=0.0920, range=0.0100, CV=4.86% | go: median=0.0031, mean=0.0031, range=0.0003, CV=4.18% | go: median=29.6774, mean=29.6774, range=0.0000, CV=0.00% |
| `document_audit` | `compiled` | 5 | median=0.0900, mean=0.0960, range=0.0600, CV=26.15% | go: median=0.0035, mean=0.0034, range=0.0002, CV=2.41% | go: median=28.2353, mean=28.2353, range=0.0000, CV=0.00% |
| `fib` | `compiled` | 5 | median=3.4400, mean=3.4020, range=0.1400, CV=1.88% | go: median=2.8887, mean=2.9384, range=0.2714, CV=3.99% | go: median=1.1578, mean=1.1578, range=0.0000, CV=0.00% |
| `fixed_width_128` | `compiled` | 5 | median=8.2000, mean=9.2880, range=3.7600, CV=17.97% | go: median=0.0049, mean=0.0049, range=0.0002, CV=1.74% | go: median=1895.5102, mean=1895.5102, range=0.0000, CV=0.00% |
| `future_await_race` | `compiled` | 5 | median=0.1900, mean=0.1960, range=0.0600, CV=12.81% | go: median=0.0031, mean=0.0031, range=0.0003, CV=4.23% | go: median=63.2258, mean=63.2258, range=0.0000, CV=0.00% |
| `future_pipeline` | `compiled` | 5 | median=0.7100, mean=0.7280, range=0.0600, CV=3.69% | go: median=0.0043, mean=0.0043, range=0.0003, CV=2.84% | go: median=169.3023, mean=169.3023, range=0.0000, CV=0.00% |
| `i_before_e` | `compiled` | 5 | median=0.1000, mean=0.1020, range=0.0100, CV=4.38% | go: median=0.0542, mean=0.0545, range=0.0021, CV=1.70% | go: median=1.8716, mean=1.8716, range=0.0000, CV=0.00% |
| `json` | `compiled` | 5 | median=0.6800, mean=0.6980, range=0.1000, CV=5.85% | go: median=1.3141, mean=1.3144, range=0.0623, CV=1.83% | go: median=0.5310, mean=0.5310, range=0.0000, CV=0.00% |
| `k_nucleotide` | `compiled` | 5 | median=3.5100, mean=3.5880, range=0.5600, CV=6.21% | go: median=0.0511, mean=0.0514, range=0.0050, CV=3.53% | go: median=69.8054, mean=69.8054, range=0.0000, CV=0.00% |
| `lexical_rollup` | `compiled` | 5 | median=0.0900, mean=0.0940, range=0.0100, CV=5.83% | go: median=0.0035, mean=0.0036, range=0.0001, CV=1.63% | go: median=26.1111, mean=26.1111, range=0.0000, CV=0.00% |
| `mandelbrot` | `compiled` | 5 | median=0.1300, mean=0.1280, range=0.0100, CV=3.49% | go: median=0.0462, mean=0.0467, range=0.0029, CV=2.54% | go: median=2.7409, mean=2.7409, range=0.0000, CV=0.00% |
| `matrixmultiply` | `compiled` | 5 | median=1.1200, mean=1.1600, range=0.2700, CV=9.42% | go: median=0.9013, mean=0.9030, range=0.0284, CV=1.17% | go: median=1.2846, mean=1.2846, range=0.0000, CV=0.00% |
| `monte_carlo_pi` | `compiled` | 5 | median=0.2000, mean=0.2000, range=0.0200, CV=5.00% | go: median=0.1900, mean=0.1898, range=0.0071, CV=1.54% | go: median=1.0537, mean=1.0537, range=0.0000, CV=0.00% |
| `mutex_await_journal` | `compiled` | 5 | median=0.5300, mean=0.5360, range=0.1500, CV=10.52% | go: median=0.0031, mean=0.0032, range=0.0003, CV=4.59% | go: median=167.5000, mean=167.5000, range=0.0000, CV=0.00% |
| `mutex_ledger` | `compiled` | 5 | median=0.5100, mean=0.5100, range=0.0600, CV=5.88% | go: median=0.0036, mean=0.0036, range=0.0002, CV=2.68% | go: median=141.6667, mean=141.6667, range=0.0000, CV=0.00% |
| `nbody` | `compiled` | 5 | median=0.3900, mean=0.3920, range=0.0300, CV=3.33% | go: median=0.0303, mean=0.0302, range=0.0007, CV=1.09% | go: median=12.9801, mean=12.9801, range=0.0000, CV=0.00% |
| `option_result_config` | `compiled` | 5 | median=0.3000, mean=0.3140, range=0.1000, CV=12.04% | go: median=0.0030, mean=0.0030, range=0.0003, CV=3.48% | go: median=104.6667, mean=104.6667, range=0.0000, CV=0.00% |
| `pidigits` | `compiled` | 5 | median=1.2600, mean=1.2800, range=0.1500, CV=4.72% | go: median=1.1059, mean=1.1118, range=0.0788, CV=2.94% | go: median=1.1513, mean=1.1513, range=0.0000, CV=0.00% |
| `quicksort` | `compiled` | 5 | median=1.7200, mean=1.7400, range=0.0900, CV=2.40% | go: median=2.2881, mean=2.3459, range=0.2912, CV=5.11% | go: median=0.7417, mean=0.7417, range=0.0000, CV=0.00% |
| `rational_series` | `compiled` | 5 | median=2.6300, mean=2.6260, range=0.4000, CV=5.63% | go: median=0.0119, mean=0.0118, range=0.0004, CV=1.43% | go: median=222.5424, mean=222.5424, range=0.0000, CV=0.00% |
| `regex_set_audit` | `compiled` | 5 | median=0.1100, mean=0.1080, range=0.0200, CV=7.75% | go: median=0.0042, mean=0.0042, range=0.0004, CV=3.48% | go: median=25.7143, mean=25.7143, range=0.0000, CV=0.00% |
| `regex_stream_audit` | `compiled` | 5 | median=0.1300, mean=0.1400, range=0.0500, CV=15.97% | go: median=0.0040, mean=0.0041, range=0.0005, CV=4.31% | go: median=34.1463, mean=34.1463, range=0.0000, CV=0.00% |
| `regex_suffix_audit` | `compiled` | 5 | median=1.1900, mean=1.2020, range=0.0800, CV=2.78% | go: median=0.0307, mean=0.0306, range=0.0011, CV=1.37% | go: median=39.2810, mean=39.2810, range=0.0000, CV=0.00% |
| `reverse_complement` | `compiled` | 5 | median=0.1000, mean=0.1140, range=0.0600, CV=22.87% | go: median=0.0139, mean=0.0138, range=0.0012, CV=3.30% | go: median=8.2609, mean=8.2609, range=0.0000, CV=0.00% |
| `sudoku_masks` | `compiled` | 5 | median=8.7900, mean=8.7580, range=0.4200, CV=1.94% | go: median=0.5270, mean=0.5292, range=0.0163, CV=1.25% | go: median=16.5495, mean=16.5495, range=0.0000, CV=0.00% |
| `tapelang_alphabet` | `compiled` | 5 | median=3.2600, mean=3.2940, range=0.1700, CV=2.20% | go: median=1.8088, mean=1.8126, range=0.0936, CV=2.38% | go: median=1.8173, mean=1.8173, range=0.0000, CV=0.00% |
| `word_frequency` | `compiled` | 5 | median=0.2000, mean=0.1960, range=0.0100, CV=2.79% | go: median=0.0046, mean=0.0046, range=0.0003, CV=2.41% | go: median=42.6087, mean=42.6087, range=0.0000, CV=0.00% |

## Inputs

- `v12/docs/perf-baselines/2026-07-16-compiled-selected-comparison.json` — generated `2026-07-16T16:17:09.784078Z`, CPU `None`
