# External Benchmark Comparison

- Generated: `2026-07-29T19:33:10.278417Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows:
  `2026-07-29-transaction-ledger-audit-interpreter-reference.json`
- Fresh Go reference rows:
  `2026-07-29-transaction-ledger-audit-go-reference.json`
- Suite: `custom`
- Able benchmarks: `transaction_ledger_audit`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `transaction_ledger_audit` | `compiled` | ok (5) | verified (5) | aa5a4fe7f85ce13998797ef506647a93f16e0ee747613683268fd801d609c812 | 0.0500 | 0.0073 | 6.85x | n/a | n/a | n/a | n/a |
| `transaction_ledger_audit` | `bytecode` | ok (5) | verified (5) | aa5a4fe7f85ce13998797ef506647a93f16e0ee747613683268fd801d609c812 | 4.5900 | n/a | n/a | 0.0344 | 133.43x | 0.0914 | 50.22x |
