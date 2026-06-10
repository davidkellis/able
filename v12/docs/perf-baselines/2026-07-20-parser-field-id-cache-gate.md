# Parser field-ID cache gate

## Decision

Keep the generic Tree-sitter field-ID cache. The Go AST mapper previously used
`Node.ChildByFieldName(...)` at 120 production call sites. The Go binding
allocates a C string, resolves that invariant grammar field name, calls
Tree-sitter, and frees the C string on every lookup. Able now enumerates the
active grammar's field table once and routes those same lookups through
`Node.ChildByFieldId(...)`.

Numeric IDs are not hard-coded, so regenerating the grammar may reorder fields
without invalidating the parser. Unknown names retain the original
name-lookup fallback. This is a universal parser operation used by every Able
source file; it does not recognize a benchmark, stdlib type, nominal
container, Regex operation, or application shape.

## Implementation and parity

`pkg/parser/node_field.go` builds one immutable field-name-to-ID table from
`Language.FieldCount()` and `Language.FieldNameForId(...)`. It deliberately
does not use the binding's `FieldIdForName(...)` helper, which creates a C
string without freeing it in the current dependency version. A nil-safe
`childByFieldName(...)` helper uses the cached ID when known.

A grammar-inventory test proves that every field ID has exactly the cached
name and ID. It then compares cached and native name lookup for every grammar
field on every node in a parsed tree, including nil and missing-child results.
Focused parser, loader, and interpreter parity tests pass.

## Interleaved exact gate

The baseline binary contains the previously retained node-kind cache and the
candidate differs only by this field-ID change. Ten independent
baseline/candidate process pairs per application were order-balanced on the
workstation. Both sides parsed exactly the same modules and source bytes.

| Application | Baseline mapping | Candidate mapping | Mapping change | Mapping wins | Baseline load | Candidate load | Load change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 59.505 ms | 54.130 ms | -9.0% | 9/10 | 133.556 ms | 125.118 ms | -6.3% |
| Dependency Plan | 28.562 ms | 26.664 ms | -6.6% | 10/10 | 65.720 ms | 62.823 ms | -4.4% |
| Future Await Race | 18.047 ms | 16.829 ms | -6.8% | 7/10 | 39.763 ms | 38.872 ms | -2.2% |
| Configuration Validation/Extraction | 150.998 ms | 136.966 ms | -9.3% | 8/10 | 347.708 ms | 325.418 ms | -6.4% |

AST mapping improved in all four programs and won 34/40 exact pairs. Total
loading improved in all four and won 31/40 pairs. Mapping plus origin
annotation also improved in every program: 77.775 to 71.845 ms, 38.306 to
35.810 ms, 23.640 to 22.587 ms, and 211.570 to 193.882 ms. This combined
measure accounts for GC moving across the observer boundary.

Ten merged baseline and candidate loader profiles per application confirm the
mechanism. Baseline `ChildByFieldName` was 4.93%, 4.48%, and 6.10% cumulative
in the longer Document, Dependency, and Configuration sets; the shorter
Future set did not sample it reliably in this cohort. The candidate profiles
contain no `ChildByFieldName`, `CString`, name-resolution, or corresponding
free subtree. The replacement helper plus `ChildByFieldId` is 0%-3.23%
cumulative.

## Normal process controls

The verifier-backed harness completed ten cold-bytecode and ten warmed-runtime
processes for every application. All 80 processes succeeded without timeout;
all 40 cold outputs passed their external verifiers.

| Application | Cold bytecode mean | Verification | Warm `main()` mean |
| --- | ---: | ---: | ---: |
| Document Audit | 0.317 s | 10/10 | 13.337 ms/op |
| Dependency Plan | 0.487 s | 10/10 | 289.697 ms/op |
| Future Await Race | 0.152 s | 10/10 | 9.826 ms/op |
| Configuration Validation/Extraction | 1.349 s | 10/10 | 818.699 ms/op |

These normal cohorts are correctness and product controls, not the A/B
selection evidence. The parser does not run inside warmed `main()` or a
generated application binary, and independent cold process means visibly move
with workstation load. The order-balanced exact loader gate above isolates
the candidate.

Machine-readable and rendered normal-process records:

- `2026-07-20-parser-field-id-cache-candidate.json`
- `2026-07-20-parser-field-id-cache-candidate-scorecard.md`

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

Next audit redundant origin annotation in package combination. The loader
first traverses every file AST to assign its correct file path, then traverses
the combined module containing those same nodes again. `ast.annotateValue`
remains 8.1%-15.4% cumulative in the four post-candidate profile sets, making
this a larger shared pure-Go loader wall than the remaining individual parser
accessors.

The next tranche should first count new versus already-annotated nodes in each
pass across the same four applications. If the second pass is materially
redundant in at least three, add an explicit traversal mode that may skip a
subtree only when its root is already present and the caller guarantees that
the prior traversal was complete. Multi-file tests must prove that original
nodes keep their own file paths while newly constructed module/package nodes
receive the primary path. Then repeat the interleaved loading gate and normal
verifier controls. Do not weaken the public partial-table semantics of
`AnnotateOrigins`, add application rules, or begin WASM work.
