# Go source-hygiene modularization retained

Date: 2026-07-27

## Decision

Retain the behavior-preserving Go source modularization and the alignment-only
typechecker formatting correction.

The release-consolidation inventory identified four modified Go files over the
1,000-line project limit. A repository-wide active-Go audit then found one
additional pre-existing tracked compiler file at 1,616 lines. All five were
split at existing function/test boundaries without rewriting function bodies,
assertions, call paths, lowering rules, or runtime behavior.

No compiler lowering, generated output, runtime semantics, interpreter/VM
behavior, language rule, canonical stdlib implementation, benchmark,
dependency, fixture expectation, reference application, or WASM behavior
changed.

## Source layout

| Original file | Before | After | Extracted file | Lines |
| --- | ---: | ---: | --- | ---: |
| `generator_controlflow.go` | 1,135 | 830 | `generator_controlflow_loop_control.go` | 313 |
| `compiler_array_intrinsics_test.go` | 1,054 | 848 | `compiler_array_intrinsics_boundary_test.go` | 212 |
| `interpreter_arrays.go` | 1,083 | 923 | `interpreter_array_members.go` | 168 |
| `bytecode_vm_slot_const_assignment_test.go` | 1,084 | 985 | `bytecode_vm_slot_const_assignment_overflow_test.go` | 108 |
| `generator_specialized_impl_calls.go` | 1,616 | 907 | `generator_specialized_impl_types.go` | 712 |

The extracted responsibilities are:

- break, continue, loop-expression, loop-statement, and breakpoint lowering;
- Array rest-binding, nullable-intrinsic, and explicit handle-boundary guards;
- Array direct members, metadata reads, index conversion, and `IndexError`
  construction;
- slot-constant subtract and overflow execution guards;
- specialized-implementation type-expression selection, compatibility,
  template matching, normalization, and specialization keys.

`typechecker/type_utils.go` remains 381 lines. `gofmt` changed only alignment
in its primitive integer-bounds table.

After the split:

- every active tracked or non-ignored Go file is below 1,000 lines;
- every visible changed/untracked Go file is `gofmt`-clean;
- `git diff --check` reports no errors.

Generated Go under disposable build workspaces was excluded from the source
limit and removed by the final cleanup.

## Focused verification

| Gate | Result | Elapsed | Peak RSS |
| --- | --- | ---: | ---: |
| Compiler loop/break and Array-boundary guards | pass | 11.42 s | 1,056,064 KB |
| Interpreter Array/member and slot-constant guards | pass | 10.94 s | 1,865,072 KB |
| Typechecker package | pass | 1.61 s | 399,556 KB |
| 15 specialization/interface guards | pass | 37.75 s | 1,260,676 KB |
| Focused compiler/interpreter/typechecker vet | pass | 1.89 s | 405,880 KB |

The specialization group covers generic nominal methods, named unions,
Option/Result, Heap, bound generic fields, generic interface adapters,
canonical specialization keys, execution, and callback specialization.

## Broad verification

`./run_all_tests.sh` passed in 701.68 seconds at 4,727,904 KB peak RSS:

- coverage, scorecard, feature, selection, and threshold contracts passed;
- all non-compiler packages passed;
- all 33 compiler batches passed;
- specialization-heavy batch 19 completed in 131.566 seconds, versus the
  preceding 132.800-second release audit;
- batch 29 completed in 74.666 seconds with unchanged tests;
- the bytecode fixture corpus passed in 89.857 seconds.

The earlier individual-timing audit remains authoritative because no test body
changed: batches 19 and 29 have respective individual maxima of 33.35 and
10.58 seconds.

Canonical stdlib verification used `TMPDIR=/var/tmp` and the disk-backed Go
cache. It passed in both modes:

- tree-walker: 17 seconds;
- bytecode: 15 seconds;
- complete command: 39.30 seconds at 887,880 KB peak RSS.

## Generated artifact cleanup

After verification, the reviewed default `scripts/cleanup.sh --apply` policy
removed 79.57 GiB across ten reproducible project-local paths:

- root `tmp` and `target`;
- `v12/tmp`;
- `v12/interpreters/go/.gocache`, `.gotmp`, `target`, and `tmp`;
- `able.test` and `compiler.test`;
- `v12/__pycache__`.

This removed the 57 visible `.gotmp` files and 36 Python bytecode files
identified by the inventory. The artifacts were deleted rather than moved to
trash, but are reproducible from source. No selected path contained a tracked
file. A post-cleanup dry run reports no default cleanup candidate.

Profile evidence was retained. The reusable disk-backed caches outside the
project remain:

- `/var/tmp/able-go-cache`: 48 GiB;
- `/var/tmp/able-compiled-test-cache`: 1.5 GiB and previously verified below
  its 1536 MiB policy ceiling.

## Recommendation

Keep production performance mutation paused.

Next perform a parser-specific modularization of
`v12/parser/tree-sitter-able/grammar.js`, the remaining known handwritten
source file above 1,000 lines.

Why: the Go source boundary is now compliant, but the 1,650-line grammar still
violates the same project limit and governs a 197,036-line generated parser.

What it entails: extract cohesive grammar helpers/rule families into JS
modules while preserving the exported grammar, regenerate tree-sitter assets,
force the Go parser harness to relink the generated parser, and run the grammar
corpus, Go parser tests, AST fixture harness, and complete handoff. Do not
touch WASM behavior.

Why it is important: it clears the last known handwritten size blocker while
keeping parser generation reproducible and avoiding manual edits to generated
`parser.c`.
