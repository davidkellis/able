# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-21T23:05:06Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fixed_width_128` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.5909 | 9b1e7b1cb5f737c0726e0ef0490152098786925ccb090a40ca1caea7543279f7 | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `fixed_width_128` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.1344 | 7655ea9f640810e2cdb796e9323389ec22367e63bb50c2715577aa25a0f51e1f | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `rational_series` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1815 | 88feefdf0cb2d5d729a992e91e2b31da5c5528ed4e27c51ffb498446437f4df8 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `rational_series` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1776 | 7cddf887e05d9c2795420f2bd2ba3150fe5ca544b3a1d208882192d53aa0f135 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `wide_integer_records` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0745 | 7a26e96bc7aab8344ec9da95037d7feb31935ff297b4a02b702599a337d15b47 | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
| `wide_integer_records` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2103 | 5d9c0a97b32ffab5dc4106f0d9baf50ba9f38213a8e2a535ef0553e880854059 | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
