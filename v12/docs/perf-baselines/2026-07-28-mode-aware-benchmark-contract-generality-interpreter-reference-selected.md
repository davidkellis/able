# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-28T19:28:44Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fib` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 5.2074 | 713f87ed003b1d6e99f2fdc0ff5b4e21d52935fcf3110b6a02d3fb378cd2ccac | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 |
| `fib` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.3390 | 8a05f59f64e61e6514be4b200bf27953b75f6f97224bd9b6dec3c5a4a40bb61d | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 |
| `binarytrees` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5610 | 8c8439ef4f9bb056b1d880045d8c33fd6b8a6aa587769fc14855f87a5f02dc9e | 92b6df65f712164fc10a53dbc1085312406b233110001316a85b78ed0a16cfab |
| `binarytrees` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5744 | 39ad43b7a851c37819f0e7ffcc662d9e46983da913d353029cf6512031e741d0 | 92b6df65f712164fc10a53dbc1085312406b233110001316a85b78ed0a16cfab |
| `matrixmultiply` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.3338 | db91f1b169f289258edfd3497815731c042e21f51037442291e6f07348740da4 | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 |
| `matrixmultiply` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.1064 | 0f24effadf62b5a38281b13267b06d33cda192d377adf0d024fd371df38b3d37 | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 |
| `quicksort` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6732 | a74b4a71a6e7cba638d608fc8cd2e4f923a296ff5cafd1395f51d5d312ef5924 | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 |
| `quicksort` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6615 | b0c7538265806dd43747a821480395696f913ea7ef77601aa0664bd1fb16f193 | 88148e21399796b608b9762acf30ecb3a1d938a57a60945c20653dc74c6b3e60 |
| `sudoku_masks` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.7962 | d193633f77cd4fe0f9396df76e7f6564bd4a64f95faa6ef78c82bfd795db7175 | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 |
| `sudoku_masks` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.1343 | 019a9b44da6e17af014ebf5c8af15e5064ddfb082ab50a1ed097611a9b5124a9 | 9354bc257cae59f24fce2f106308db1c36a10976f52089fa54a6d50b7e50b506 |
| `i_before_e` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0843 | 8bfaa1f26c92a881e9d1cfe59a1c5b46898c2a21e3dcc5dbb4b7aadacae28ccc | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `i_before_e` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1191 | eb1b632013f8655315acc77c500a5623048b114ca9fa810d4c659f480d57257e | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `base64` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 3.8207 | c19ec8e11eb48b54f50f73d40729338dddffb89db89575d1e8470fe36144bfb3 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `base64` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.4968 | 29c3d989b9f6bcdd4fe9def690e017bbd270b767f68c0c4446e61c43783f2fda | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `json` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 2.5929 | 1368eca8a57392ec4397ad755449848fef914b155ad0541a58fb9205b7a0605b | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `json` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.6712 | 6d72bd6d7656ff81ceb481cd621eb2eef60e12d6ad602433c30a223dbbd6420c | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e |
| `monte_carlo_pi` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.4459 | 5ec04624423829c5ffb1ff6c7b92706631eda752bbd634a806850a4edc2c5efc | 00af5ae835aae7c208a589306bd8f8b3a25c7625b626b46be382c6a90d990abd,628b02eec667f35165dc34ebd8a12b9c531eb377225b10593d1d4d451af5ae08,6c9a75d2ba9fd89a9458cc9697baef7b36ebdb2ce49f9027991998ad924edb67,7ff3eca1590b815e0d06b7d46df784a4df5edf8e9d8bb52a7bd2a57929ac2487,febb039b777da02d338c1e6d26e7af2aa8ff27d5f2dc9def8218ee8bdf45c7e7 |
| `monte_carlo_pi` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified-nondeterministic | 1.5447 | e87ea1fcbc9a312b8a96d30d3437beb969fe0c71abec4b696a456224a88a59d6 | 22465a6604833454496f3b094b4d03775b7e7b87a1d6c74956cb3221c3d0e1d0,76685fb10f91a61271a4e7ad445c2831a1b5bdde31d051f6dd02bc9c87c8f8b0,8a7e128f9997d36d436a20169e59b4089b7bebb005b36125248646edaabbf563,935143b071487996b19a05f76bb7a00e4dcec6a044d91f658431875a04323d42,997d660bc4994125c91a8fe60badbb588a6dd1a09c2be18b7a9c35f235898458 |
| `pidigits` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 4.0342 | 98090e1898327ebc9133554bd91609d1732698d76b0ca8b30696ea65726ff0f9 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `pidigits` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 10.0133 | 2926a844870ab5241b2694db2bae830501ef1dea0219758d05962e45f945e3d0 | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3706 | 6a3aaece33adafca79ed1ed3aa2d6e7861f4a45d7c5416668ea4a58f503f9365 | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `mandelbrot` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.9661 | 7f814f0ccfd009685befd6bc57e9984e3fd9b2990c88dd1264e9f8676ea7aa2c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `reverse_complement` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0263 | 0ebe3d5c98324ba53472cd75208b8ba13c311134cf3f1caae47077929a5ca1a1 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `reverse_complement` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0748 | cd44c852891f5c29601171b001e47f7569f2d3b74d9a5a47d49e90f51ed422bc | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `k_nucleotide` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.3037 | d382e44e51a7d27f2ad28ed9c9ae1bf4645502a95dfa0bacd455d926dce96806 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `k_nucleotide` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.2331 | de3ae3289db5ab5bc6eb3735042a5209be173ccb9f86f872d8cd88deac0e1745 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `nbody` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2002 | 3e6fe954e3b03c77fda6a4a25e8b09ed4450817ade3e5e5b013de73cc54bea2c | bdcf7a5967f944dc85b65e0e03ed5fd5daf6b699793224d6b03c7b2c75ea8790 |
| `nbody` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3351 | 4f4032cfe736cd4053bb3fd27fa156e5a07e04d5b355fcdb2afad66a5a0ce2be | bdcf7a5967f944dc85b65e0e03ed5fd5daf6b699793224d6b03c7b2c75ea8790 |
| `tapelang_alphabet` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5897 | 0bfb786c312b67354a391d59f00cdb9a13844419c46199274e95b5372c66687f | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `tapelang_alphabet` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.7746 | f77051cd0751f82e73ece8ff9981b340b085a9ab6c02061f9e6a861dc15ddfc0 | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `distance_field` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5807 | 96edeaa2e1ca51879516235f8650439549bade854a466efad1cd42ae7f51b331 | 49ade5dafc8840964c43278ce4e186532d45583e19bea41f6510a90d7c918f88 |
| `distance_field` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3396 | 6b5dbf5ed985f485c811eb0deb383cbb3356ffdcf68280b16f025029c29fa92e | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `rms_norm` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.7947 | 91cd92bc83b0e0e5635a1c17bf64cbbd1d1883fec81f78a0fbe8ba22c299f2d0 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `rms_norm` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5263 | 55b6a731691291a2d1b7293cc4ba140939ba6bf661b896337632dd11396b273b | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `fasta_generation` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2215 | 8679fbc7e84f9f39545949489a5c3ad4d22ef12fea3efa66e122d8357dcb1510 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
| `fasta_generation` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2976 | 7e7cda57ae6212a85f75bbe6627c813aae11d59ac37a06e1e8f1cccd60ca8eb9 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
