# Origin skip-known package-combination gate

## Decision

Keep the explicit skip-known origin traversal for package combination. The
loader first annotates every parsed file AST with that file's path, then builds
a combined module whose imports, exports, and body point back to those same
AST nodes. The old second traversal recognized existing table entries but
still reflectively revisited every descendant.

`ast.AnnotateOrigins(...)` retains its original partial-table semantics. A new
`ast.AnnotateOriginsSkippingKnown(...)` operation may stop at an existing node
only when its caller guarantees an earlier complete traversal of that subtree.
Only the loader's combined-module pass makes that guarantee. This is a general
multi-file source-loader rule, not an application, benchmark, stdlib, or
nominal-type fast path.

## Duplication admission audit

Temporary counters measured unique nodes added versus nodes already present
in the origin table across one complete load. The counters were removed before
the performance candidate was built.

| Application | Origin visits | New nodes | Existing-node revisits | Existing share |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 37,018 | 18,569 | 18,449 | 49.8% |
| Dependency Plan | 19,506 | 9,796 | 9,710 | 49.8% |
| Future Await Race | 10,688 | 5,352 | 5,336 | 49.9% |
| Configuration Validation/Extraction | 83,157 | 41,734 | 41,423 | 49.8% |

The skip-known pass reduces total visits to 19,120, 10,109, 5,555, and
42,627 respectively. Only 551, 313, 203, and 893 known subtree roots must be
examined to connect the new combined-module wrappers; their descendants are
already complete. The duplicated shape therefore repeats materially in all
four unrelated programs.

## Semantics

The original traversal continues through an existing root and fills missing
descendants, preserving callers that supply partial tables. A focused AST test
locks that behavior. A separate test documents that skip-known intentionally
requires a complete subtree.

A two-file package-combination test proves that nodes from `a.able` and
`b.able`, including their package paths and function identifiers, keep their
own file origins. The newly constructed combined module, copied package
statement, and copied package identifier receive the sorted primary path
`a.able`.

## Final interleaved gate

The baseline binary contains the retained Tree-sitter allocator, node-kind,
and field-ID improvements. After removing every temporary counter, ten
independent baseline/candidate process pairs per application were
order-balanced on the workstation.

| Application | Baseline origin | Candidate origin | Origin change | Origin wins | Baseline load | Candidate load | Load change | Load wins |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 18.857 ms | 10.144 ms | -46.2% | 10/10 | 134.717 ms | 121.862 ms | -9.5% | 8/10 |
| Dependency Plan | 9.454 ms | 4.942 ms | -47.7% | 10/10 | 64.716 ms | 56.628 ms | -12.5% | 8/10 |
| Future Await Race | 4.832 ms | 2.764 ms | -42.8% | 10/10 | 34.115 ms | 31.881 ms | -6.5% | 8/10 |
| Configuration Validation/Extraction | 55.128 ms | 25.364 ms | -54.0% | 10/10 | 304.705 ms | 283.458 ms | -7.0% | 8/10 |

The exact target improves in all four programs and wins all 40 pairs. Total
program loading improves 6.5%-12.5% and wins 32/40 pairs. Native parsing and
AST mapping, which this candidate cannot change, moved in both directions;
their samples remain in the total means.

Ten merged loader profiles per side confirm the mechanism. Cumulative
`ast.annotateValue` falls from 11.57% to 2.59% in Document Audit, 11.29% to
8.62% in Dependency Plan, 13.89% to 6.25% in Future Await Race, and 15.36% to
11.62% in Configuration Validation. The remaining subtree belongs primarily
to the necessary first per-file traversal.

## Normal process controls

The verifier-backed harness completed ten cold-bytecode and ten warmed-runtime
processes for each application. All 80 processes succeeded without timeout;
all 40 cold outputs passed their external verifiers.

| Application | Cold bytecode mean | Verification | Warm `main()` mean |
| --- | ---: | ---: | ---: |
| Document Audit | 0.262 s | 10/10 | 10.880 ms/op |
| Dependency Plan | 0.446 s | 10/10 | 261.729 ms/op |
| Future Await Race | 0.146 s | 10/10 | 8.923 ms/op |
| Configuration Validation/Extraction | 1.224 s | 10/10 | 735.737 ms/op |

Compared with the immediately preceding ten-process field-ID control, all
four cold means are lower, but the exact interleaved origin/load gate remains
the selection evidence. Origin annotation does not execute inside warmed
`main()` or generated application binaries.

Machine-readable and rendered normal-process records:

- `2026-07-20-origin-skip-known-candidate.json`
- `2026-07-20-origin-skip-known-candidate-scorecard.md`

## Verification

- `go test ./pkg/ast ./pkg/driver -count=1 -timeout 60s`
- `go test ./pkg/interpreter -run '^(TestLoadBytecodeProgramRuntimeBenchConfig|TestExecFixtureParity/14_05_regex_nfa_wildcards_classes)$' -count=1 -timeout 60s`
- `go test ./cmd/able -run '^TestBuildEnvFalseAllowsFallbacks$' -count=1 -timeout 60s`
- `go test ./pkg/compiler -run '^TestCompilerIteratorFilterMapBenchmarkShapeExecutesFromNonMainSourcePackage$' -count=1 -timeout 60s`
- Four applications x two modes x ten normal processes: 80/80 successful;
  all 40 cold bytecode outputs verified.

No spec, AST contract, benchmark source, generated application code, or
canonical `able-stdlib` source changed.

## Next selection

Next audit fusing dynamic-import discovery with the remaining mandatory
per-file origin traversal. `collectDynImports` reflectively walks each parsed
AST before package combination, while origin annotation later walks that same
AST again. Post-candidate profiles put dynamic-import discovery at 3.17%-5.17%
cumulative in Document Audit, Dependency Plan, and Configuration Validation;
the remaining origin traversal is 6.25%-11.62% across all four.

The tranche should introduce one general AST node-walk primitive, use a single
per-file walk to collect nested `DynImportStatement` nodes and build that
file's origin table, store the completed table on `fileModule`, and merge those
tables during combination before annotating only new wrappers. This preserves
the spec's local-scope `dynimport` semantics, unlike a top-level-only scan.
Advance only if traversal counts and profiles confirm one walk replaces two
in at least three programs. Validate nested dynimports, multi-file origins,
ordinary imports, and the same interleaved loading/verifier gates. Do not add a
benchmark rule, alter the AST contract, or begin WASM work.
