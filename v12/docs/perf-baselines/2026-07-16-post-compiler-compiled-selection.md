# External Benchmark Comparison

- Generated: `2026-07-17T05:15:00.293429Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-16-post-compiler-go-reference.json`
- Suite: `custom`
- Able benchmarks: `array_slice_window, await_channel_mux, base64, binarytrees, channel_rollup, dependency_plan, distance_field, document_audit, fib, fixed_width_128, future_await_race, future_pipeline, i_before_e, json, k_nucleotide, lexical_rollup, mandelbrot, matrixmultiply, monte_carlo_pi, mutex_await_journal, mutex_ledger, nbody, option_result_config, pidigits, quicksort, rational_series, regex_set_audit, regex_stream_audit, regex_suffix_audit, reverse_complement, rms_norm, sudoku_masks, tapelang_alphabet, word_frequency`
- Able modes: `compiled`
- Reference languages: `go, ruby, python`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0900 | 0.0036 | 25.00x | n/a | n/a | n/a | n/a |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3980 | 0.0039 | 102.05x | n/a | n/a | n/a | n/a |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.3600 | 2.3670 | 1.00x | 2.2100 | 1.07x | 3.3100 | 0.71x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 30.6420 | 6.3222 | 4.85x | 20.3900 | 1.50x | 12.2500 | 2.50x |
| `channel_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.9240 | 0.0049 | 188.57x | n/a | n/a | n/a | n/a |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0860 | 0.0031 | 27.74x | n/a | n/a | n/a | n/a |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.0920 | 0.0110 | 8.36x | n/a | n/a | n/a | n/a |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0920 | 0.0036 | 25.56x | n/a | n/a | n/a | n/a |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.8700 | 3.0973 | 1.25x | 46.6400 | 0.08x | 60.6700 | 0.06x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.2100 | 0.0049 | 42.86x | n/a | n/a | n/a | n/a |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1220 | 0.0031 | 39.35x | n/a | n/a | n/a | n/a |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.3660 | 0.0044 | 83.18x | n/a | n/a | n/a | n/a |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1400 | 0.0591 | 2.37x | 0.1000 | 1.40x | 0.1300 | 1.08x |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.7840 | 1.4080 | 0.56x | 1.5600 | 0.50x | 2.8700 | 0.27x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 5.6440 | 0.0667 | 84.62x | n/a | n/a | n/a | n/a |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1120 | 0.0039 | 28.72x | n/a | n/a | n/a | n/a |
| `mandelbrot` | `compiled` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1500 | 0.0551 | 2.72x | n/a | n/a | n/a | n/a |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1380 | 1.0848 | 1.05x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `monte_carlo_pi` | `compiled` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.2080 | 0.2693 | 0.77x | 1.4200 | 0.15x | 1.6800 | 0.12x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.5280 | 0.0042 | 125.71x | n/a | n/a | n/a | n/a |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 1.0000 | 0.0079 | 126.58x | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.2200 | 0.0655 | 3.36x | n/a | n/a | n/a | n/a |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.2740 | 0.0050 | 54.80x | n/a | n/a | n/a | n/a |
| `pidigits` | `compiled` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.4100 | 1.6853 | 0.84x | 9.1800 | 0.15x | n/a | n/a |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 2.0500 | 3.0451 | 0.67x | 14.5800 | 0.14x | 20.3200 | 0.10x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.1600 | 0.0131 | 12.21x | n/a | n/a | n/a | n/a |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.3000 | 0.0046 | 65.22x | n/a | n/a | n/a | n/a |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.2820 | 0.0042 | 67.14x | n/a | n/a | n/a | n/a |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a | 2.0920 | 0.0414 | 50.53x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1220 | 0.0151 | 8.08x | n/a | n/a | n/a | n/a |
| `rms_norm` | `compiled` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 0.1100 | 0.0100 | 11.00x | n/a | n/a | n/a | n/a |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 9.3260 | 0.5674 | 16.44x | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 5.2060 | 1.9656 | 2.65x | 67.8300 | 0.08x | 58.9900 | 0.09x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.2500 | 0.0046 | 54.35x | n/a | n/a | n/a | n/a |
