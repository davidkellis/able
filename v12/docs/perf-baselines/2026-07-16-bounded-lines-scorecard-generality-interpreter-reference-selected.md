# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-16T08:46:54Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU affinity: `none`

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fib` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 54.3191 | 2f6016de0be3f90fc98b7020626d2e99b6a7737648dd100abefc9f3ecc94758c | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `fib` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 43.2450 | 41d417c909521f90c7dca7fe2653f5043646d0755d1828d2eb9350cfcadb65f8 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `matrixmultiply` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 48.4717 | db91f1b169f289258edfd3497815731c042e21f51037442291e6f07348740da4 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `matrixmultiply` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 44.2432 | 0f24effadf62b5a38281b13267b06d33cda192d377adf0d024fd371df38b3d37 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0760 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1070 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.5852 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.3263 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.4928 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.5667 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.4348 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 11c331cd8e55b63c9a68e6f1fdfa7b11a08807e869f352f85051c7b19f9b9cd7,51be5bc1c5bbf8eca5bb3e6963b913dd128051c22eb52fa9b4b82ecadd406887,63c9e60cf02bad441fda44ee5b851c3e1cef11dfd68dcc5454a21ca097e57911,7829faffdfb7ba13ef74a9bbc423f57078f26f3564f62063d90cca67618b9735,9c7bfe4c87b9dfcde2b6cae7b9fe0f875a9c468f8e0fda0407d3cc427f1d6800 |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.4870 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 34c3c14eed3ab2243ea6c44a16461d12ab9c579879cabb5c518e638222d2c471,35b0827c36edff6540e8afb0d8956e3d172bbaf7cd7f022a477a3dcde31c15da,582ed5c910922286cc7bb71996f9286564a0e9eebc851f8073f342292fe6e8fb,6d7cf56795f513232f3a29068b174d9095f2801031a90d0eec72aa225b919239,da75c28a61332e017d897ec84efd00e021c4d8f07da76cd8952ae97d6de5c190 |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.8071 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 9.6658 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1577 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.7797 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0243 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0684 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2149 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1941 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
