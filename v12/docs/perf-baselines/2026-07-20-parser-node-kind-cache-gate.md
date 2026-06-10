# Parser node-kind cache gate

## Decision

Keep the generic Tree-sitter node-kind cache. Go AST mapping previously called
`Node.Kind()` at 139 parser sites. The binding implements that operation as a
C call followed by `C.GoString`, even though every returned name belongs to the
fixed Able grammar. The parser now resolves the grammar's symbol names once
and maps each node's stable `KindId()` through that immutable table.

This applies to every Able source file loaded by the tree-walker, bytecode
interpreter, compiler, and test tooling. It does not recognize a benchmark,
stdlib type, nominal container, or application shape, and it does not change
the AST or language semantics.

## Exact decomposition and selection

Opt-in diagnostics now split program loading into native Tree-sitter parsing,
Go AST mapping, and origin annotation. They run only when the bytecode runtime
phase-output observer is enabled. Ten fresh processes for each workload gave
the following pre-candidate means:

| Application | Parsed modules | Source bytes | Native parse | AST mapping | Origin annotation | Program load |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 13 | 125,096 | 47.506 ms | 70.638 ms | 19.144 ms | 159.755 ms |
| Dependency Plan | 9 | 49,383 | 24.146 ms | 36.455 ms | 10.478 ms | 76.503 ms |
| Future Await Race | 2 | 28,078 | 12.518 ms | 17.486 ms | 4.851 ms | 37.772 ms |
| Configuration Validation/Extraction | 27 | 270,984 | 121.470 ms | 180.192 ms | 43.044 ms | 367.996 ms |

AST mapping was the largest measured loader subphase in all four unrelated
programs. Ten merged loader profiles per program then identified the same
concrete mapping child: `Node.Kind()` was 11.76%, 9.72%, 11.11%, and 16.11%
cumulative respectively. Its `ts_node_type` and Go string-conversion subtree
was present in every profile set.

## Interleaved candidate gate

An initial sequential cohort was contaminated by changing workstation load:
Document Audit became slower in native parsing and origin annotation, neither
of which the candidate changes. The decision therefore uses ten
order-balanced, interleaved baseline/candidate process pairs per application.
All source-byte and parsed-module counts matched exactly.

| Application | Baseline mapping | Candidate mapping | Mapping change | Mapping wins | Baseline load | Candidate load | Load change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 63.632 ms | 56.124 ms | -11.8% | 9/10 | 134.598 ms | 126.202 ms | -6.2% |
| Dependency Plan | 31.588 ms | 27.702 ms | -12.3% | 10/10 | 67.102 ms | 62.962 ms | -6.2% |
| Future Await Race | 17.326 ms | 15.968 ms | -7.8% | 8/10 | 37.123 ms | 36.792 ms | -0.9% |
| Configuration Validation/Extraction | 167.738 ms | 138.199 ms | -17.6% | 10/10 | 333.813 ms | 320.195 ms | -4.1% |

Configuration Validation shifted some garbage-collection time across the
mapping/origin observer boundary. The combined mapping-plus-origin mean still
improved from 206.011 ms to 194.359 ms (-5.7%). The same combined measure
improved 9.5%, 9.8%, and 3.2% in the other three programs. Native parsing was
effectively unchanged, as expected.

Post-candidate ten-profile sets contain no `Node.Kind()`, `ts_node_type`, or
per-query Go string conversion. The replacement `KindId()` subtree is smaller
and allocation-free but remains a C accessor; no private binding or unsafe
Tree-sitter representation was introduced.

## Normal process controls

The verifier-backed harness completed ten fresh processes for cold bytecode
and warmed bytecode runtime in all four applications. All 40 cold executions
verified and no process failed or timed out.

| Application | Cold bytecode mean | Verification | Warm `main()` mean |
| --- | ---: | ---: | ---: |
| Document Audit | 0.272 s | 10/10 | 11.543 ms/op |
| Dependency Plan | 0.479 s | 10/10 | 321.266 ms/op |
| Future Await Race | 0.148 s | 10/10 | 13.997 ms/op |
| Configuration Validation/Extraction | 1.392 s | 10/10 | 823.194 ms/op |

The parser executes during cold source loading and compiler front-end work,
not inside generated application binaries or the warmed `main()` loop. Warmed
results are therefore correctness/load controls rather than performance
selection evidence. The previous non-interleaved normal cohort and this one
also show ordinary workstation movement; the exact interleaved loader gate is
the retained A/B evidence.

Machine-readable and rendered normal-process records:

- `2026-07-20-parser-node-kind-cache-candidate.json`
- `2026-07-20-parser-node-kind-cache-candidate-scorecard.md`

## Verification

- `go test ./pkg/parser ./pkg/driver -count=1 -timeout 60s`
- `go test ./pkg/interpreter -run '^(TestLoadBytecodeProgramRuntimeBenchConfig|TestExecFixtureParity/14_05_regex_nfa_wildcards_classes)$' -count=1 -timeout 60s`
- `go test ./cmd/able -run '^TestBuildEnvFalseAllowsFallbacks$' -count=1 -timeout 60s`
- `go test ./pkg/compiler -run '^TestCompilerIteratorFilterMapBenchmarkShapeExecutesFromNonMainSourcePackage$' -count=1 -timeout 60s`
- Four applications x two modes x ten normal processes: 80/80 successful;
  all 40 cold bytecode outputs verified.

No spec, AST contract, benchmark source, generated application code, or
canonical `able-stdlib` source changed.

## Next selection

Next test cached Tree-sitter field IDs for the parser's 120
`ChildByFieldName(...)` call sites. The binding currently allocates and frees a
C string, then asks Tree-sitter to resolve the field name on every access;
`ChildByFieldName` remains a concrete 2.4%-4.8% cumulative child across all
four post-candidate profile sets. The grammar exposes stable field IDs and
`ChildByFieldId`, so the candidate can remain a universal parser operation.

The tranche should precompute the Able grammar's field-name-to-ID table once,
replace name-based child lookup through one nil-safe helper, prove helper
parity in parser tests, and rerun the same interleaved four-program mapping and
loader gate. Retain it only if the exact mapping phase improves broadly; do
not select a cursor rewrite, private binding layout, benchmark rule, named
nominal rule, or WASM work from this evidence.
