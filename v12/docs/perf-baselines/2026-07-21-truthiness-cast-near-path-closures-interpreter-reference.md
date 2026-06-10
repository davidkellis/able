# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-21T22:29:21Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `rms_norm` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.0289 | 91cd92bc83b0e0e5635a1c17bf64cbbd1d1883fec81f78a0fbe8ba22c299f2d0 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `rms_norm` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6665 | 55b6a731691291a2d1b7293cc4ba140939ba6bf661b896337632dd11396b273b | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 2.1717 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 1d8c704bf03cc42d1ce05cab589a1ec9ad7d558f44b4ca0b379807111643f9e1,2d07812b7a33c4901cb76c31cfaf4581024892fb5b9801be689af0602fb53e77,3b3e1de3630db550b3017ec661a491c303d558393e3924fadd8d0a7ec96ff22e,46885d6f4ef44bb63217eeff294464186c91bec6bb8b24b1806a008484c2b0ae,a2b2c5f7c599a537b41f3767b517f0ab6bc4d36f61871cb3a152b4ea5bbfe7c1 |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 2.1583 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 063163569831ebbae16fbd8e7a46686d705e1d74523ab3a774d61a9c5308c858,876a01078042994c3543b2003ee33415082e258cdc26c57bbf47a3aae96d558f,9996a0daf4d1166d10b17c8804638315aa77fc307298ef495829eb90685fb4e8,b345e3cee921952020e0b16a61fd708a6b49187405db5394667bedacecf42368,ebaacd3f14960acb371824763851a7208ccbb45b1f7a726680863bc354773061 |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3980 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.1426 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `distance_field` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6968 | 96edeaa2e1ca51879516235f8650439549bade854a466efad1cd42ae7f51b331 | 49ade5dafc8840964c43278ce4e186532d45583e19bea41f6510a90d7c918f88 |
| `distance_field` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.4783 | 6b5dbf5ed985f485c811eb0deb383cbb3356ffdcf68280b16f025029c29fa92e | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
