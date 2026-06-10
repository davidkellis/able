# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-15T08:17:54Z`
- Suite: `generality`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `3`
- Timeout: `45s`
- CPU affinity: `14`

| Benchmark | Language | Status | Validation | Real (s) | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- |
| `fib` | python | 0/3 (attempted 1, timeouts 1, failures 0) | unavailable | n/a |  |
| `fib` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 40.1801 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `binarytrees` | python | 0/3 (attempted 1, timeouts 1, failures 0) | unavailable | n/a |  |
| `binarytrees` | ruby | 0/3 (attempted 1, timeouts 1, failures 0) | unavailable | n/a |  |
| `matrixmultiply` | python | 1/3 (attempted 2, timeouts 1, failures 0) | verified | 44.0134 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `matrixmultiply` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 41.9293 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `quicksort` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 22.6439 | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `quicksort` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 13.4652 | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `sudoku` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 2.6042 | 70f621ee8eb896bb208139b31d46eca3c081fe9bdb4b55275e3c20b6db1bed29 |
| `sudoku` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 5.2667 | 70f621ee8eb896bb208139b31d46eca3c081fe9bdb4b55275e3c20b6db1bed29 |
| `sudoku_masks` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 15.2837 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `sudoku_masks` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 18.4365 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `i_before_e` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0733 | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.1017 | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 3.4421 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 2.1771 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 2.3201 | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.4611 | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified-nondeterministic | 1.2851 | 0b3023bfb6b045e1dd27846f40ab40e65781df56446a72641c232b08a61ca47a,4e285c77524414d28bb746850659f8ac9d257e69ad739e3b93d882b33d91753a,90b2bb8a0dce8570de089e11180a57ec0685941d911694a8e7133096b20c8893 |
| `monte_carlo_pi` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified-nondeterministic | 1.3522 | 0c5728fe873d255d98749de15e2e4e1440907cc11ee4ba9c627cd896ce9b0286,0cef17b8af30936f579a83e4c3b9a54e2d92e599493104a8ca135bfdd70d9491,b41bc8ee03673790e275f257a05d219de812b8981a38fbf43170968f8a04768b |
| `pidigits` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 3.6258 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 8.8207 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.0302 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.6353 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0233 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0661 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.1287 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.1156 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `nbody` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 1.7544 | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `nbody` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 2.7624 | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `tapelang_alphabet` | python | 0/3 (attempted 1, timeouts 1, failures 0) | unavailable | n/a |  |
| `tapelang_alphabet` | ruby | 0/3 (attempted 1, timeouts 1, failures 0) | unavailable | n/a |  |
