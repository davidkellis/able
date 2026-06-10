# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-17T14:19:58Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fib` | python | 1/5 (attempted 2, timeouts 1, failures 0) | verified | 54.7512 | 2f6016de0be3f90fc98b7020626d2e99b6a7737648dd100abefc9f3ecc94758c | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `fib` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 44.0366 | 41d417c909521f90c7dca7fe2653f5043646d0755d1828d2eb9350cfcadb65f8 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `matrixmultiply` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 47.8571 | db91f1b169f289258edfd3497815731c042e21f51037442291e6f07348740da4 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `matrixmultiply` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 46.3074 | 0f24effadf62b5a38281b13267b06d33cda192d377adf0d024fd371df38b3d37 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0788 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1130 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.7912 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.4701 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.6341 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.8039 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5449 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 078211e948b9cb4cb502168914a56238f8c40102fa284d16fe15ac855d7147c6,24ef076b7c3dc38c03159af7ac70e8f50f05cd25301fb7bd4ff9b2e43a753d88,5df5db2fecdc38faabc246e8089f11c6ecbfc614a77a546f1c2fcb8d11be5965,a6bad3b75f78f4c2a6adab7857b14ff0cf5031fbfc35134102ef320affdc624f,e48260104e1a1443c9f84eb8ccbea3a4dd06e2cbf02476669ffe720aa96a71ea |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5596 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 1a03094b5b7603a269cd0a2a1d6d8e0414633cf60570d84b05d144904ad24a34,1dfe817c59fb9a3fb15f7705bed5ce0659709889ae52fb8b1455d4681a77c4ac,21d0638825a4643bcd8ca475e6bf2d904183d60407a5252fb0af73b0d5a11931,886d99e4184fb447c1f484bcb3afc9ad3aa53e56189a868be711fc2a74ff8e29,d9c4b48814f41bf0d42fcc1ec4d2bd6601ab697ab08d7bc5a6bc36002f9aa8fe |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.2562 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 10.3098 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1789 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.8385 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0260 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0756 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3276 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3517 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `distance_field` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5631 | 96edeaa2e1ca51879516235f8650439549bade854a466efad1cd42ae7f51b331 | 49ade5dafc8840964c43278ce4e186532d45583e19bea41f6510a90d7c918f88 |
| `distance_field` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3364 | 6b5dbf5ed985f485c811eb0deb383cbb3356ffdcf68280b16f025029c29fa92e | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `rms_norm` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.9028 | 91cd92bc83b0e0e5635a1c17bf64cbbd1d1883fec81f78a0fbe8ba22c299f2d0 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `rms_norm` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5449 | 55b6a731691291a2d1b7293cc4ba140939ba6bf661b896337632dd11396b273b | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
