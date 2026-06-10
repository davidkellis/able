# External Benchmark Comparison

- Generated: `2026-07-10T15:50:25.573430Z`
- External results: `/home/david/sync/projects/benchmarks-document-audit-publish-20260710/results.json`
- Suite: `document-audit`
- Able benchmarks: `document_audit`
- Able modes: `compiled, bytecode, treewalker`
- Reference languages: `go, ruby, python`
- CPU affinity: `0`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (1) | 0.2100 | 0.0030 | 69.29x | 0.0434 | 4.84x | 0.0190 | 11.05x |
| `document_audit` | `bytecode` | ok (1) | 0.3300 | 0.0030 | 108.88x | 0.0434 | 7.61x | 0.0190 | 17.36x |
| `document_audit` | `treewalker` | ok (1) | 24.0900 | 0.0030 | 7948.44x | 0.0434 | 555.44x | 0.0190 | 1267.19x |
