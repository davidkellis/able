# Loader AST-walk fusion gate

## Decision

Keep one general source-loader improvement: each parsed file AST is now walked
once to assign source origins and discover nested dynamic imports. Previously,
`collectDynImports` reflectively traversed the complete AST during `parseFile`,
then origin annotation reflectively traversed the same AST again during package
combination.

The AST package now owns the reusable reflection walker. Ordinary origin
annotation and skip-known package-wrapper annotation use the same traversal
contract. The loader stores the completed origin table on each `fileModule`,
adopts the first sorted file's table, merges additional file tables, and only
annotates newly constructed combined-package wrappers.

This is independent of application, benchmark, nominal type, standard-library
container, and execution mode. It changes cold source loading and compiler
frontend work, not warmed bytecode `main()` or generated application runtime.

## Dynamic-import correctness

The spec permits `dynimport` at package or local scope. The new loader test
exposed an existing parser mapping gap: the grammar accepted an
`import_statement` inside a block, but `parseStatement` discarded it. Import
mapping now uses one shared parser for module and block scopes. A local
`dynimport` therefore remains in the AST and the fused walk discovers its
package dependency.

The `06_10_dynamic_metaprogramming_package_object` shared fixture now keeps one
dynamic import at package scope and performs another inside `main`. It passes
tree-walker/bytecode parity and the compiler dynamic-fallback path. This aligns
the implementation with existing v12 spec text; no spec or AST shape changed.

## Twenty-pair phase gate

Separate baseline and candidate test binaries were built before and after the
change. Twenty independent baseline/candidate process pairs per application
were order-balanced. Means include all samples, including workstation
outliers.

| Application | Baseline load | Candidate load | Change | Candidate wins |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 129.617 ms | 129.459 ms | -0.12% | 12/20 |
| Dependency Plan | 62.208 ms | 61.848 ms | -0.58% | 13/20 |
| Future Await Race | 35.480 ms | 34.049 ms | -4.03% | 10/20 |
| Configuration Validation/Extraction | 293.088 ms | 282.079 ms | -3.76% | 15/20 |

The two smaller source graphs remain near the timing noise floor, while the
larger and structurally unlike Future and Configuration graphs improve about
4%. No application mean regresses. Pairwise geometric means improve 0.25%,
1.29%, 3.60%, and 3.65%, respectively. The candidate wins 50/80 exact pairs.

The origin subphase is not directly comparable after fusion: baseline timing
contains origin assignment alone, while candidate timing contains both origin
assignment and dynamic-import inspection. Total `program_load` is therefore
the application-facing selection measurement.

## CPU-profile mechanism check

Ten loader-only CPU profiles per side and application were merged. The
baseline's origin and dynamic-import traversals are sequential, so their
cumulative shares can be added. The candidate's one `ast.walkValue` subtree
includes both responsibilities.

| Application | Baseline two walks | Candidate fused walk |
| --- | ---: | ---: |
| Document Audit | 10.61% | 8.66% |
| Dependency Plan | 8.06% | 5.36% |
| Future Await Race | 17.64% | 8.57% |
| Configuration Validation/Extraction | 10.90% | 3.93% |

`driver.collectDynImports` disappears from every candidate profile. The
remaining walk is the required per-file traversal, and its cumulative share is
lower in all four applications. This confirms that the measured change comes
from replacing two general AST walks with one rather than moving work to a new
helper.

## Normal verifier controls

The ordinary external-comparison harness completed ten cold-bytecode and ten
warmed-runtime processes for each application. All 80 processes succeeded and
all 40 cold outputs passed their external verifiers.

| Application | Cold bytecode mean | Verification | Warm `main()` mean |
| --- | ---: | ---: | ---: |
| Document Audit | 0.294 s | 10/10 | 13.101 ms/op |
| Dependency Plan | 0.446 s | 10/10 | 301.521 ms/op |
| Future Await Race | 0.126 s | 10/10 | 9.773 ms/op |
| Configuration Validation/Extraction | 1.268 s | 10/10 | 777.091 ms/op |

Machine-readable and rendered records:

- `2026-07-20-loader-walk-fusion-candidate.json`
- `2026-07-20-loader-walk-fusion-candidate-scorecard.md`

Warm timings are correctness controls, not evidence for this loader-only
change.

## Verification

- `go test ./pkg/parser ./pkg/ast ./pkg/driver -count=1 -timeout 60s`
- Focused dynamic-import interpreter tests and fixtures, including
  `TestExecFixtureParity/06_10_dynamic_metaprogramming_package_object`
- Focused compiler dynamic-feature, dynamic-fallback, and non-main-source
  execution tests
- `go test ./cmd/able -run '^TestBuildEnvFalseAllowsFallbacks$' -count=1 -timeout 60s`
- Four applications x two modes x ten normal processes: 80/80 successful;
  all 40 cold bytecode outputs verified

Every focused test completed in under 19 seconds. No canonical `able-stdlib`,
VM runtime, generated runtime, benchmark source, nominal lowering, or WASM code
changed.

## Next selection

The next selection was to audit whether `Node.Range()` could replace the two
Tree-sitter position crossings in `spanFromNode`. The subsequent pinned-binding
audit found that `Range()` performs four separate C calls rather than one, so
that premise is closed without a candidate. The same tranche also measured and
reverted a parser-wide child-count hoist after a broad application regression.
See `2026-07-20-tree-sitter-span-child-count-audit.md` for the resolved evidence
and next selection.
