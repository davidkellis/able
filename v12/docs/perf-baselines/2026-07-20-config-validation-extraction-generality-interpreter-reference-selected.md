# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-20T20:50:39Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0838 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1125 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.8809 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.3808 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.6995 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.7315 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5017 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 1bf05a67140b09985167cb2351ab09129e92b150afa38e1931fade84ee5f18ab,6cc41bbf72f9a0b8d747d9e0ca24d1ae996cd11898ef25c5a74cea16ce1896b3,6d191872c81cddf13996ae53ee1899f233cf7c054fb4785dd5f912065d48356e,c7eacf4fc5a59f0565019ebc1bff55034176c3d4789ae3752eb284fbf4ff8dc4,d75a063d0bb92e98f1f6ba565dfebfb47b403c115dfe61d90b069fed393cec3b |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5246 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 106b6b69458ac93e37f66a76b0b26c33b8360d4c7b9fa845aa45fd30290154ec,40bca5d8e617ff4b7e0dbfc68f51728ed7b1f829d04da5cdb383b08be0cf09f2,805572ddfc30cfd82e769876ba4817417a89785da4bf3adef377590cbfb31e40,a9cd13f559bcef3067cc8d14464ab9cdda3e723ef1f23afe9f66ad5bf231aaeb,c58e254fec234e69c1075db1069e9e6fbcfb3dc48ea439738ca8ae94a468ae6c |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.0333 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 10.2466 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1668 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.8306 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0293 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0819 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3555 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2526 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `distance_field` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5432 | 96edeaa2e1ca51879516235f8650439549bade854a466efad1cd42ae7f51b331 | 49ade5dafc8840964c43278ce4e186532d45583e19bea41f6510a90d7c918f88 |
| `distance_field` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3103 | 6b5dbf5ed985f485c811eb0deb383cbb3356ffdcf68280b16f025029c29fa92e | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `rms_norm` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.8306 | 91cd92bc83b0e0e5635a1c17bf64cbbd1d1883fec81f78a0fbe8ba22c299f2d0 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `rms_norm` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5339 | 55b6a731691291a2d1b7293cc4ba140939ba6bf661b896337632dd11396b273b | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `fasta_generation` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1978 | 8679fbc7e84f9f39545949489a5c3ad4d22ef12fea3efa66e122d8357dcb1510 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
| `fasta_generation` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2889 | 7e7cda57ae6212a85f75bbe6627c813aae11d59ac37a06e1e8f1cccd60ca8eb9 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
