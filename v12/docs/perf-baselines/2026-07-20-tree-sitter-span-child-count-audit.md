# Tree-sitter span and child-count audit

## Decision

Keep no parser performance change from this tranche.

Two general Tree-sitter wrapper-call ideas were audited. The proposed
`spanFromNode` conversion to `Node.Range()` was rejected before implementation
because the pinned Go binding does not provide one combined range operation.
A separate candidate hoisted immutable child counts out of parser loops, but
it failed the broad application gate and was fully reverted.

No benchmark, application, standard-library, compiler, VM, generated runtime,
nominal lowering, language syntax, spec, or WASM behavior changed.

## `Node.Range()` binding audit

The project pins `github.com/tree-sitter/go-tree-sitter@v0.25.0`. In that
binding, `Node.Range()` is ordinary Go composition of four public methods:

1. `StartByte()`
2. `EndByte()`
3. `StartPosition()`
4. `EndPosition()`

Each method invokes a separate Tree-sitter C function. Tree-sitter's public C
node API has no combined `ts_node_range` function. Current `spanFromNode` calls
only `StartPosition()` and `EndPosition()`, so replacing it with `Range()`
would increase the operation from two C/Go crossings to four while discarding
the returned byte offsets. No range candidate was eligible for measurement.

This invalidates the `Node.Range()` premise recorded as the next selection in
the preceding loader-walk report; that report now points here for the resolved
audit.

## Child-count candidate

The adjacent parser audit found approximately sixty loops whose conditions
repeatedly called `NamedChildCount()` or `ChildCount()` while traversing an
immutable syntax tree. The candidate evaluated each count once in the loop
initializer. It was mechanical, parser-wide, independent of grammar feature or
application, and did not alter child ordering or mapping semantics.

Focused parser, AST, driver, dynamic-import, and regex fixture controls passed
before measurement.

## Twenty-pair phase gate

Separate frozen baseline and candidate test binaries ran twenty independent,
order-balanced process pairs per application. The selection metric is the
loader's `program_load` phase. Means include every sample because this is a
normal workstation and repeated cohorts, rather than idle-CPU assumptions,
are the noise control.

| Application | Baseline load | Candidate load | Change | Candidate wins |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 134.507 ms | 135.005 ms | +0.37% | 10/20 |
| Dependency Plan | 66.086 ms | 62.648 ms | -5.20% | 14/20 |
| Future Await Race | 34.805 ms | 37.087 ms | +6.56% | 7/20 |
| Configuration Validation/Extraction | 296.917 ms | 283.848 ms | -4.40% | 13/20 |

Pairwise geometric changes were +0.44%, -5.55%, +5.66%, and -4.14%. The
candidate therefore produced two useful-looking improvements, one neutral
result, and one clear regression. It fails the requirement that a retained
optimization help broadly without making an unlike guard application slower.

## Profile mechanism check

Ten loader-only CPU profiles per side and application were merged. CPU sample
resolution is coarse for these short processes, but the Future Await Race
profiles provide the decisive mechanism check: the cumulative
`NamedChildCount` share falls from 14.29% in the baseline aggregate to no
sample in the candidate aggregate, while the twenty-pair wall-clock result
regresses 6.56%. Removing the intended sampled wrapper wall therefore did not
produce an application win.

Other aggregates are mixed: sampled `NamedChildCount` cumulative time falls
for Document Audit (4.38% to 2.33%) and Configuration Validation (5.47% to
4.60%), but rises within sampling error for Dependency Plan (3.28% to 6.15%).
Total `runtime.cgocall` remains the dominant cumulative owner on both sides.
The profiles support rejecting the isolated loop rewrite rather than claiming
that one wrapper subtotal predicts end-to-end loading.

## Verification after revert

- `go test ./pkg/parser -count=1 -timeout 60s`
- `go test ./pkg/ast ./pkg/driver -count=1 -timeout 60s`
- `go test ./pkg/interpreter -run '^(TestLoadBytecodeProgramRuntimeBenchConfig|TestExecFixtureParity/(06_10_dynamic_metaprogramming_package_object|14_05_regex_nfa_wildcards_classes))$' -count=1 -timeout 60s`
- `git diff --check`

All focused tests pass after the full candidate revert. No candidate artifacts
or temporary instrumentation remain in production code.

## Next selection

The subsequent duplicate-span audit found a 34.53%-39.78% duplicate share in
all four applications. More than 99% came from dispatcher layers annotating the
same AST object from the same syntax node after a delegated parser had already
done so. A zero-cache dispatcher cleanup removed 28.26%-32.72% of all span
requests and passed the broad gate. See
`2026-07-20-parser-redundant-span-dispatch-gate.md` for the retained change and
next selection.
