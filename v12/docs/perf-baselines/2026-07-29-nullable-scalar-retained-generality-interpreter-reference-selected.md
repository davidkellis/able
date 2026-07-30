# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-29T22:49:39Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fib` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 5.4730 | 713f87ed003b1d6e99f2fdc0ff5b4e21d52935fcf3110b6a02d3fb378cd2ccac | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 |
| `fib` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.2566 | 8a05f59f64e61e6514be4b200bf27953b75f6f97224bd9b6dec3c5a4a40bb61d | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 |
| `binarytrees` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5667 | 8c8439ef4f9bb056b1d880045d8c33fd6b8a6aa587769fc14855f87a5f02dc9e | 92b6df65f712164fc10a53dbc1085312406b233110001316a85b78ed0a16cfab |
| `binarytrees` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6081 | 39ad43b7a851c37819f0e7ffcc662d9e46983da913d353029cf6512031e741d0 | 92b6df65f712164fc10a53dbc1085312406b233110001316a85b78ed0a16cfab |
| `matrixmultiply` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.1299 | db91f1b169f289258edfd3497815731c042e21f51037442291e6f07348740da4 | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 |
| `matrixmultiply` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.2610 | 0f24effadf62b5a38281b13267b06d33cda192d377adf0d024fd371df38b3d37 | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 |
| `quicksort` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1138 | a74b4a71a6e7cba638d608fc8cd2e4f923a296ff5cafd1395f51d5d312ef5924 | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 |
| `quicksort` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2566 | b0c7538265806dd43747a821480395696f913ea7ef77601aa0664bd1fb16f193 | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 |
| `sudoku_masks` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.8424 | d193633f77cd4fe0f9396df76e7f6564bd4a64f95faa6ef78c82bfd795db7175 | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 |
| `sudoku_masks` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.5671 | 019a9b44da6e17af014ebf5c8af15e5064ddfb082ab50a1ed097611a9b5124a9 | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0956 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1638 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.4302 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.5588 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.6856 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.6625 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5028 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 11fe7a8ed5731bed707cc91d0ed73bf920fa2d8704f8f4b6774be2a39a2b363a,494433302355ee84b2c6c7e53e4e219a15045b6a80aa49bd831360803fa29d59,6d300094d4deb4564da117d2a8c1b83b4b50eea99800603e57a5d2932df5efee,756c886275e0df84151290659a2828135b0b3c38c75e90563119fcece5d483d7,f2760e4504a5eb76ddcfb3049bccec9642c7db7c39dc611cc927d0a1fc85473b |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.6149 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 28ebc02a69faf915498984e67e72cdd3edb1e813682b3009f7f825bfffe8ebad,5c6ff19381be7dee1fb3ec04db462f28f8264dffeabbe70df013c49ae65eb2a5,a8c5e922f415cf5292715421d722bcb31460eca1feacca1a28e1689e519be34b,d3c1fb66c3fca8da76aadd939b5b7110dbed7660ca989447d669ac8f204fff97,e25fa670a50c8cdb55bf29895af48a36f6d8f11d7e80cd57bdd106a1c50de369 |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.2060 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 10.4238 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2522 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.9620 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0254 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0747 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3863 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2642 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `nbody` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2174 | 3e6fe954e3b03c77fda6a4a25e8b09ed4450817ade3e5e5b013de73cc54bea2c | bdcf7a5967f944dc85b65e0e03ed5fd5daf6b699793224d6b03c7b2c75ea8790 |
| `nbody` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3560 | 4f4032cfe736cd4053bb3fd27fa156e5a07e04d5b355fcdb2afad66a5a0ce2be | bdcf7a5967f944dc85b65e0e03ed5fd5daf6b699793224d6b03c7b2c75ea8790 |
| `tapelang_alphabet` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6164 | 0bfb786c312b67354a391d59f00cdb9a13844419c46199274e95b5372c66687f | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `tapelang_alphabet` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.7878 | f77051cd0751f82e73ece8ff9981b340b085a9ab6c02061f9e6a861dc15ddfc0 | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `distance_field` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5909 | 96edeaa2e1ca51879516235f8650439549bade854a466efad1cd42ae7f51b331 | 49ade5dafc8840964c43278ce4e186532d45583e19bea41f6510a90d7c918f88 |
| `distance_field` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3238 | 6b5dbf5ed985f485c811eb0deb383cbb3356ffdcf68280b16f025029c29fa92e | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `rms_norm` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.8329 | 91cd92bc83b0e0e5635a1c17bf64cbbd1d1883fec81f78a0fbe8ba22c299f2d0 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `rms_norm` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5018 | 55b6a731691291a2d1b7293cc4ba140939ba6bf661b896337632dd11396b273b | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `fasta_generation` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2002 | 8679fbc7e84f9f39545949489a5c3ad4d22ef12fea3efa66e122d8357dcb1510 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
| `fasta_generation` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3027 | 7e7cda57ae6212a85f75bbe6627c813aae11d59ac37a06e1e8f1cccd60ca8eb9 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
