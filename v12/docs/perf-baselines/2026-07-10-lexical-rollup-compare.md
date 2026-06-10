# Lexical Rollup External Benchmark

Lexical Rollup is an independent public-Iterable application benchmark. It
reads the checked-in ENABLE word list, deterministically takes its first
16,384 records, filters them, maps selected values to weighted `i64` scores,
collects the scores, and reduces the total. All implementations must print
`16384:4828:502100`.

The input is `i-before-e/wordlist.txt` (1,743,363 bytes). The bounded record
count keeps the tree-walker verification practical while retaining an ordinary
file/Array/Iterable/callback/collection pipeline. It does not authorize a
container-, corpus-, or callback-specific optimization.

## Verified publication

The canonical sources live in `v12/examples/benchmarks/lexical_rollup` and
`../benchmarks/lexical-rollup`. Go 1.26, Python 3.14, Ruby 4.0, and Able v12
bytecode, compiled, and tree-walker implementations built, ran, and passed
the verifier in the isolated standard-runner publication:
`../benchmarks-lexical-rollup-publish-20260710/results.json`.

| Mode | Able (s) | Go (s) | Able/Go | Ruby (s) | Able/Ruby | Python (s) | Able/Python |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| compiled | 0.059276659 | 0.003364905 | 17.62x | 0.055081010 | 1.08x | 0.022383732 | 2.65x |
| bytecode | 4.519517367 | 0.003364905 | 1343.13x | 0.055081010 | 82.05x | 0.022383732 | 201.91x |
| tree-walker | 19.046687919 | 0.003364905 | 5660.39x | 0.055081010 | 345.79x | 0.022383732 | 850.92x |

These are one-run process-wall measurements from one Docker publication; they
are comparable only within this workload.

## Paired bytecode evidence

Fresh warmed one-process profiles used the canonical stdlib pin and
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, with no trace collection. The retained
artifacts are `20260710_document_audit_lexical_pair.cpu.pprof` and
`20260710_lexical_rollup_bytecode.cpu.pprof` under
`v12/interpreters/go/.profiles/`.

| Workload | Repetitions | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 400 | 7,840,008 | 363,035 | 199 |
| Lexical Rollup | 20 | 137,925,863 | 10,238,102 | 259 |

Both profiles have material `lookupCachedMemberMethodEntry` (19.2% Document
Audit; 11.7% Lexical Rollup) and `finishInlineReturn` (13.1%; 12.8%). Those are
already-rejected generic candidates: removing a member-cache write regressed
all guards, while the raw-cell and return-guard experiments were neutral or
mixed. The remaining work does not agree: Document Audit is member-cache and
cached-next heavy, while Lexical Rollup is additionally led by type matching,
exact-native calls, iterator-next, and GC. There is no new concrete shared
descendant to optimize.

No interpreter, compiler, or stdlib behavior changed in this tranche.

The subsequent compiled-pair attribution is recorded in
`2026-07-10-compiled-pipeline-pair.md`; it found no shared eligible lowering
boundary.
