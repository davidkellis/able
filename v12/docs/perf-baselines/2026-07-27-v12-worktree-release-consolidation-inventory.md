# V12 worktree release-consolidation inventory

Date: 2026-07-27

## Decision

Do not reconstruct the long-running v12 effort as one commit per historical
tranche. Consolidate it through dependency-ordered subsystem review
boundaries, while retaining the dated logs and performance records as the
chronological evidence.

No source, fixture, generated artifact, cache, Git index, commit, or history
state changed during this inventory. Only this record, its machine-readable
summary, and the forward handoff/log documentation were added or updated.

## Baseline

The read-only inventory was captured on `master` at
`237406eccdfb025a519d898daedadee1c8d13a7b`, which exactly matched
`origin/master`. The base commit is dated 2026-06-10 and is named `wip`; the
visible worktree primarily contains the July 9–27 effort.

The pre-record `git status --porcelain=v1 --untracked-files=all` snapshot had:

- 5,142 visible changed paths;
- 559 tracked modifications;
- 4,583 untracked files;
- no tracked deletions or renames.

Its SHA-256 was
`e9c19c9bb6f8f27ee71ee10591196460d9336d7eb6bafa5712e5b54e670d8374`.
A deterministic path classification covered all 5,142 entries with zero
unmatched paths; its temporary map had SHA-256
`f51e02ce5dcb3e22746d4829e979a49ff94c00bbdebbb5dbfa329210f315dd67`.
The temporary snapshot and map were removed after this record was validated.

Tracked changes total 231,755 inserted and 169,548 deleted lines. The generated
tree-sitter C parser accounts for 147,634 additions and 147,281 deletions;
excluding it, the other 558 tracked files contain 84,121 additions and 22,267
deletions.

## Complete path classification

| Boundary | Tracked modified | Untracked | Total |
| --- | ---: | ---: | ---: |
| Language, AST, parser, typechecker | 52 | 24 | 76 |
| Runtime, tree-walker, bytecode VM | 277 | 387 | 664 |
| Compiler, AOT CLI, semantic ABI | 174 | 161 | 335 |
| Deferred WASM isolation | 7 | 19 | 26 |
| Benchmarks, evidence, scoreboards | 7 | 3,857 | 3,864 |
| Design, documentation, release tooling | 42 | 40 | 82 |
| Generated local state; exclude from review | 0 | 95 | 95 |
| **Total** | **559** | **4,583** | **5,142** |

The main source families are:

- language: the v12 spec/TODO, AST fixtures, tree-sitter sources, Go AST,
  parser, and typechecker;
- runtime: 514 interpreter/VM paths, 57 runtime paths, 78 execution-fixture
  paths, runtime support, and the tree-walker/bytecode wrappers;
- compiler: 256 compiler paths, 36 semantic-ABI paths, 42 command paths, the
  embedded kernel, and Go module metadata;
- benchmark/evidence: 3,412 performance-baseline paths, 102 visible profile
  files, 195 benchmark-fixture paths, 56 examples, 97 benchmark-tool paths,
  and two benchmark documentation paths.

The performance archive contains 1,734 untracked Markdown files. Of those, 250
are explicitly referenced from `PLAN.md`, `LOG.md`, or `v12/LOG.md`; 1,484 are
supporting reports rather than handoff records. There are 1,260 stems with
multiple evidence extensions and 890 single-extension stems. These files are
not declared stale, but decision records and raw evidence should be reviewed
as separate archive boundaries.

## Why historical tranche commits are unsafe

The same source files were changed across many logged tranches. For example:

- `generator_controlflow.go` appears in 16 log entries;
- `interpreter_arrays.go` appears in 15;
- `compiler_array_intrinsics_test.go` appears in 10;
- `bytecode_vm_slot_const_assignment_test.go` appears in three.

There is no commit-level provenance between those edits. Assigning the final
form of a shared file to one historical record would therefore invent history
and could produce intermediate commits that never represented a tested state.
The logs and dated records remain the chronological narrative; review commits
should represent coherent final subsystem states.

## Proposed review boundaries

No commits were created. If a maintainer later authorizes consolidation, use
this order:

1. **Language contract, AST, parser, and typechecker.** Review spec changes,
   parser/AST alignment, typechecking behavior, and their fixtures together.
   Regenerate parser assets only from the reviewed grammar and verify the
   parser harness.
2. **Runtime, tree-walker, and bytecode VM.** Review shared runtime carriers,
   interpreter/VM changes, execution fixtures, and wrappers after the language
   contract is stable.
3. **Compiler, AOT CLI, and semantic ABI.** Review native-carrier lowering,
   static/dynamic boundary behavior, compiler caches, generated-runtime
   package cuts, CLI behavior, and focused guards against the first two
   boundaries.
4. **Benchmark applications, fixtures, and tooling.** Review workload sources,
   selection/coverage contracts, verifier tooling, and refresh commands
   independently from their generated results.
5. **Decision records and current scoreboards.** Review the 250 log-referenced
   untracked Markdown records plus current scoreboards as the governing
   evidence.
6. **Raw measurement archive.** Review the remaining reports, JSON, TSV, text,
   compressed profiles, and `.profiles` separately so generated evidence
   cannot obscure source review.
7. **Design, manuals, logs, and release tooling.** Reconcile documentation
   against the final reviewed source and land the forward-only handoff last.

The 26 WASM paths are a separate deferred boundary and must not be included in
the compiler/AOT release series while the no-WASM performance directive
remains active.

## Objective blockers

### Oversized handwritten files

Five modified handwritten files exceed the 1,000-line project limit:

| File | Lines |
| --- | ---: |
| `v12/parser/tree-sitter-able/grammar.js` | 1,650 |
| `v12/interpreters/go/pkg/compiler/generator_controlflow.go` | 1,135 |
| `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_assignment_test.go` | 1,084 |
| `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go` | 1,083 |
| `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go` | 1,054 |

The 197,036-line generated `tree-sitter-able/src/parser.c` is a generated
asset and must not be split manually. The grammar that produces it is
handwritten and requires a parser-specific modularization decision.

### Formatting

`gofmt -l` over all 1,029 visible changed/untracked Go files outside `.gotmp`
reported one file:

`v12/interpreters/go/pkg/typechecker/type_utils.go`

The difference is alignment-only in the primitive integer-bounds table. It was
not changed during this inventory. `git diff --check` reported no whitespace
errors in tracked diffs.

### Generated and local state

Ninety-five visible untracked paths, totaling 20,250,349 bytes, are local
generated state and must not be staged:

- 57 files under `v12/interpreters/go/.gotmp`;
- 36 Python bytecode files under `v12/__pycache__`;
- the 15,127,680-byte `v12/interpreters/go/ablec` binary;
- `.stats_persistent_sorted_set_enabled.json`.

The `.gotmp` tree is a compiled-boundary-audit generated workspace. The binary
and stats file are not covered by the current default cleanup candidates,
which is a tooling-policy gap to resolve before consolidation.

A dry run of `scripts/cleanup.sh` found 79.52 GiB across ten reproducible
project-local paths, dominated by:

- 67 GiB in `v12/interpreters/go/.gocache`;
- 14 GiB in `v12/tmp`;
- 275 MiB in root `target`;
- 46 MiB in `able.test`;
- 34 MiB in `compiler.test`;
- 25 MiB in `v12/interpreters/go/target`.

It also selected `.gotmp` and Python bytecode. Nothing was deleted. The
disk-backed compiled-test cache and `/var/tmp/able-go-cache` are outside this
project-local cleanup selection and remain intentionally retained and bounded.

The profile archive currently has 196 physical files totaling 1,858,845 bytes;
102 non-ignored files appear in status. Profiles are intentionally excluded
from default cleanup and require an explicit `--include-profiles` decision.

## Recommendation

Do not stage or commit the current worktree yet.

Next perform a behavior-preserving Go source-hygiene tranche: split the four
oversized handwritten Go files below 1,000 lines and run `gofmt` on
`type_utils.go`, without changing semantics, generated output, benchmark
selection, or WASM.

Why: these are objective repository-policy failures in otherwise green source
boundaries, and they are safer to resolve before attempting review splits.

What it entails: extract cohesive helpers and test groups within their current
packages, preserve all existing dirty changes, run focused compiler,
interpreter/VM, Array, and typechecker guards, then run the complete handoff.
The tree-sitter grammar should remain a separate parser-specific follow-up
because it controls generated assets.

Why it is important: it creates reviewable source units without altering the
native-carrier or compiler/interpreter boundary behavior that the performance
work has already verified.
