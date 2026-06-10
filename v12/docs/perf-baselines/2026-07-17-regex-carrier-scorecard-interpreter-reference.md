# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-17T19:48:10Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `regex_set_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0236 | 3dfa49b7d1b958dc2ac7e278cbdf207e70ee9cb524c2d19f483d31baa1331b73 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_set_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0536 | 914b1064f4f14c3c55c6bf50c3ca9a73b4f627fd3854c84c4feb68db7c4be1b1 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0198 | 711b2bfd7229437c0e6da454f303a890cd9a2bfdba0ed67cd0f83c84c0c94414 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_stream_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0496 | 132ebe36d07b64b7f8542a0c1150c70735b47be31c44a924a66c967c42cec8e5 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `array_slice_window` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0298 | 81a554ebbdae62eef642e766d86f3d49d33666842435eb4e9a6765d979af55e9 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `array_slice_window` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0606 | 5fa9c8c230542bb1bc30cbfc44302181df6880c6e3bb0646cb90178d4b06d017 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
