# Pinned Go Reference Refresh

- Generated: `2026-07-18T16:06:34Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-3`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `fib` | 5/5 (timeouts 0, failures 0) | verified | 3.3636 | 79a2f518ad5cefdb74f7fe7fffd343031b162fbc790cf1ed2ab98dbfc1bd088e | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `binarytrees` | 5/5 (timeouts 0, failures 0) | verified | 12.4042 | 27d067b4bfe5f501f1fe2d6a3e0254d699759c750296d9414161c2b5d3623b9f | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 |
| `matrixmultiply` | 5/5 (timeouts 0, failures 0) | verified | 1.1461 | 4b77c1e4f0add1763c830c9da22a9d92c43d14e48432bab021e2146e3ccb1e42 | 0dfcf69f5c73589f22465d7054ec20cd1aa43a7a1829c57673b147a49290fc13 |
| `quicksort` | 5/5 (timeouts 0, failures 0) | verified | 3.2474 | f0bc08270a8f666cb9df5fc21fbbebba5fabd6375e7dc7b1313a5949e7ad485a | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `sudoku_masks` | 5/5 (timeouts 0, failures 0) | verified | 0.6899 | 0a925cd66382c7162c8dab61c6fc4f95b895528ea580afb0f504e531983db223 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `i_before_e` | 5/5 (timeouts 0, failures 0) | verified | 0.0977 | 63386a111f2fd35ff949092a419c99ce7dcf21e81b24575ece330f7729df65c0 | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | 5/5 (timeouts 0, failures 0) | verified | 3.0974 | 61a41a0ec45d3b3a8c890853d0a7839ce371b758ded8d8b8d0129a5b28390af6 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | 5/5 (timeouts 0, failures 0) | verified | 1.7014 | 6ea05b659c17322d9fd009c1b6f7ca3862e4621d2d854bbae25cd2eaccbc91c1 | 16d35d81fed277412d781cad8942b00f198974d4de5981bfb61142fe73b7e8e1 |
| `monte_carlo_pi` | 5/5 (timeouts 0, failures 0) | verified-nondeterministic | 0.2495 | 34b6ad655126f97e453b17350d49552144c2bcd332fb3b3ca7192382554e3877 | 88c2a692c03bac7d6958f8f8b009abfbf1d93d31ac6bb66722348126204b817e,9ded48d63b153da0ddf157cdff5edadb4cf4b4c8a1db5faec3716c3c922ab97d,b1da4b80c8cb50887d0ded44d0bb9063c2dcf76e887fd766076d18572b222308,bccead5db1668cc1b137ab82a4e6ef2f0bc9eda9a0e4c4919424f1af0e597cb7,e89d5b2ef35c4ac23b9edbd1911692aa38a0d0d116b1b8b6c76cbd7092c4a14a |
| `pidigits` | 5/5 (timeouts 0, failures 0) | verified | 1.4199 | c8669a71e52ce32ed4e6852547efec42f67e8b8f5656e88653720d74b37da58c | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | 5/5 (timeouts 0, failures 0) | verified | 0.0576 | c0a81d428de3e5b86ec9980b514441c8388dadfc550ed345a1f6271dcd8f0b4c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | 5/5 (timeouts 0, failures 0) | verified | 0.0242 | d6ab5b73111cf5dbc06e1f3879ccc8548d082d88b778cfcba731a2fa3aaacd74 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | 5/5 (timeouts 0, failures 0) | verified | 0.0843 | 8f1ec0923f819b16a7a63adbb1f3d7165c526e7df73462e8edf865d6a39c9a29 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `nbody` | 5/5 (timeouts 0, failures 0) | verified | 0.0429 | ecbfb6dbba972da47cd6b2a4e377d053eb47b1fe1427236c625574aec82d294d | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `tapelang_alphabet` | 5/5 (timeouts 0, failures 0) | verified | 2.3799 | dcbe38fbf5c452ac955899f01362d0e395944e15e018d66c0a52479f73cdfb32 | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `distance_field` | 5/5 (timeouts 0, failures 0) | verified | 0.0181 | 418832deca45aa420b724aeca415542cc5d3eac0ce5b9ad1c6305654e52a03b8 | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `rms_norm` | 5/5 (timeouts 0, failures 0) | verified | 0.0133 | 70f4c52128ffdffef690939025f0bf4bbf56ef11a72ca69f669a1aa31fd0dd83 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
