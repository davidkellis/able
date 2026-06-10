# External Benchmark Comparison

> Historical comparison for the pre-repair LinkedList-source variant. Use
> `2026-07-10-document-audit-array-compare.md` for the current direct-Array
> source and refreshed isolated Docker results.

- Generated: `2026-07-10T15:23:48.684848Z`
- External results: `/home/david/sync/projects/benchmarks-document-audit-publish-20260710/results.json`
- Suite: `document-audit`
- Able benchmarks: `document_audit`
- Able modes: `compiled, bytecode, treewalker`
- Reference languages: `go, ruby, python`
- CPU affinity: `0`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (1) | 0.1600 | 0.0032 | 49.73x | 0.0466 | 3.44x | 0.0228 | 7.00x |
| `document_audit` | `bytecode` | ok (1) | 0.3300 | 0.0032 | 102.57x | 0.0466 | 7.09x | 0.0228 | 14.45x |
| `document_audit` | `treewalker` | ok (1) | 25.3000 | 0.0032 | 7863.90x | 0.0466 | 543.42x | 0.0228 | 1107.47x |
