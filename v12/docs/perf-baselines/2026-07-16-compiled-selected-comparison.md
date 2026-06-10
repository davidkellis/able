# External Benchmark Comparison

- Generated: `2026-07-16T16:17:09.784078Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-16-compiled-selected-go-reference.json`
- Suite: `custom`
- Able benchmarks: `array_slice_window, await_channel_mux, base64, binarytrees, channel_rollup, dependency_plan, document_audit, fib, fixed_width_128, future_await_race, future_pipeline, i_before_e, json, k_nucleotide, lexical_rollup, mandelbrot, matrixmultiply, monte_carlo_pi, mutex_await_journal, mutex_ledger, nbody, option_result_config, pidigits, quicksort, rational_series, regex_set_audit, regex_stream_audit, regex_suffix_audit, reverse_complement, sudoku_masks, tapelang_alphabet, word_frequency`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0940 | 0.0036 | 26.11x |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3480 | 0.0039 | 89.23x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.3260 | 2.2857 | 1.02x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 27.8140 | 5.2735 | 5.27x |
| `channel_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 1.1940 | 0.0047 | 254.04x |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0920 | 0.0031 | 29.68x |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0960 | 0.0034 | 28.24x |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.4020 | 2.9384 | 1.16x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 9.2880 | 0.0049 | 1895.51x |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1960 | 0.0031 | 63.23x |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.7280 | 0.0043 | 169.30x |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1020 | 0.0545 | 1.87x |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.6980 | 1.3144 | 0.53x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 3.5880 | 0.0514 | 69.81x |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.0940 | 0.0036 | 26.11x |
| `mandelbrot` | `compiled` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1280 | 0.0467 | 2.74x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1600 | 0.9030 | 1.28x |
| `monte_carlo_pi` | `compiled` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.2000 | 0.1898 | 1.05x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.5360 | 0.0032 | 167.50x |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.5100 | 0.0036 | 141.67x |
| `nbody` | `compiled` | ok (5) | verified (5) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.3920 | 0.0302 | 12.98x |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.3140 | 0.0030 | 104.67x |
| `pidigits` | `compiled` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.2800 | 1.1118 | 1.15x |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.7400 | 2.3459 | 0.74x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 2.6260 | 0.0118 | 222.54x |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.1080 | 0.0042 | 25.71x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1400 | 0.0041 | 34.15x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a | 1.2020 | 0.0306 | 39.28x |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1140 | 0.0138 | 8.26x |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 8.7580 | 0.5292 | 16.55x |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.2940 | 1.8126 | 1.82x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.1960 | 0.0046 | 42.61x |
