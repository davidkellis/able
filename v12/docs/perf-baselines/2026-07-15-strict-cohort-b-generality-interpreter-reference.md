# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-16T05:22:49Z`
- Suite: `generality`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU affinity: `none`

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fib` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 57.5715 | 2f6016de0be3f90fc98b7020626d2e99b6a7737648dd100abefc9f3ecc94758c | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `fib` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 45.5928 | 41d417c909521f90c7dca7fe2653f5043646d0755d1828d2eb9350cfcadb65f8 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `binarytrees` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 12.7937 | 8c8439ef4f9bb056b1d880045d8c33fd6b8a6aa587769fc14855f87a5f02dc9e | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 |
| `binarytrees` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 54.8039 | 39ad43b7a851c37819f0e7ffcc662d9e46983da913d353029cf6512031e741d0 | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 |
| `matrixmultiply` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 49.2453 | db91f1b169f289258edfd3497815731c042e21f51037442291e6f07348740da4 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `matrixmultiply` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 51.8666 | 0f24effadf62b5a38281b13267b06d33cda192d377adf0d024fd371df38b3d37 | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c |
| `quicksort` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 24.3114 | 188870db19753d1245a39fc3ecc971bf788f3c889e8602893d46030bdc6f124f | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `quicksort` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 14.7283 | 62b24797a800c7aa47a2d97335239e3aca9e1e9e02e653718a294d41efb110e8 | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `sudoku_masks` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 17.1785 | b108bfa273342d263666719c550b2ea3b4a99ee7fbdb2d08a88eb0e88343bcd7 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `sudoku_masks` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 21.7267 | d20ad64fac37d7688b05c6d63af6442c0a4b928ffe964bf86939057af512c6b6 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0906 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1171 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.8149 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.3754 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.5440 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.6323 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.4469 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 047b8a945eecda84c50e417968ec0964c9b53240e7705bb95e7c46410a0bf05b,0eb00de28599a1e1a54a5b5416b6e17167daed85443e4a26aa520bae3c2b5829,bc7aefdc68ae2de256b5b4cc69484cacf559ad021d54e406ede7509f4e6638ba,ca717f4635e69a1b761733b5051ec69fd7ad469bedd940b119ec79b3e57357ac,f3bc9f634c6a765e4f45f67c723c275b95b1e7c49cdffddfcdb8e45579880c45 |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5530 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 00b8c0e9cdae949df32dc66324c3939b2feaad023925f3827b11053519c1b021,487373e4fea8b5d504cd6854b84ffe25125245de0833118d57ce5f25fa6a5ff1,56114e0bf2d92aba7d8ed934e95fc5dc24b64f01a89ee3df453f4791fad05a31,87329b305606eab898e95f49da67c34437e413d8d6ea91fd8b224c41ecda99be,d296045cdd5aed80d7ca4aebdec1726c31389d6702c7035c79c2567461fff97e |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.8748 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 9.7157 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1832 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.8865 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0250 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0707 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3074 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2687 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `nbody` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.0049 | 13e3640944e9cb31985e622e38223e1c2093dffcdd9b32cc9c022f1dfd1a695b | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `nbody` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.0318 | 009b17efa869fad1ec8d49638a469ab3642b35a9d3d686b21ad9a52fd39f99f0 | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `tapelang_alphabet` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 57.2333 | 0bfb786c312b67354a391d59f00cdb9a13844419c46199274e95b5372c66687f | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `tapelang_alphabet` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 73.7805 | f77051cd0751f82e73ece8ff9981b340b085a9ab6c02061f9e6a861dc15ddfc0 | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
