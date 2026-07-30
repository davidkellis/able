# Silent Go test-suppression audit closure

Date: 2026-07-30

## Decision

Retain no code.

The active non-WASM v12 Go test corpus has no unclassified successful
early-return, build-tag, package-selection, or compiler-sharding mechanism
that silently removes normal semantic coverage. The preceding parser-lane
correction remains the only required change.

## Top-level return audit

A Go AST audit parsed all 617 active non-WASM test files and inspected 3,226
top-level `Test*` functions while ignoring returns inside nested function
literals. It found 21 returns capable of ending the top-level test:

| Classification | Returns | Why it is not a silent opt-out |
| --- | ---: | --- |
| isolated child/helper process coordination | 12 | the parent starts and validates the child, or the child performs one requested setup/rebuild action |
| successful search before an unconditional failure | 7 | failure follows the loop if the expected instruction, event, module, or descriptor is absent |
| bounded polling/concurrency success | 2 | timeout or final-state assertions fail when the expected condition is not reached |

There is no unconditional success return before semantic work and no
environment-controlled feature exclusion.

The wider environment-sensitive audit found 33 functions that both read
environment state and contain returns. They are:

- child-process coordinators;
- compiled-CLI release-lane selection already enforced through an explicit
  `Skipf`;
- compiler fixture selection, batching, strictness, and diagnostics
  resolvers whose empty cases already produce explicit skips or failures;
- benchmark/profile configuration helpers; and
- stdlib/search-path/GOCACHE configuration helpers.

None converts a normal semantic test into an invisible successful no-op.

## Build constraints

Only three active non-WASM test files carry Go build constraints:

- `fib_bench_test.go` uses `!(js && wasm)` and is included in the native
  default build;
- `fixture_replay_test.go` uses `!(js && wasm)` and is included in the native
  default build; and
- `bytecode_vm_small_int_boxing_reuse_enabled_test.go` uses
  `able_bytecode_box_reuse`.

The latter is a documented diagnostic-only counter implementation. Normal
production selects `bytecode_vm_small_int_boxing_reuse_disabled.go`, whose
false compile-time constant erases the diagnostic guards. The normal
interpreter test binary lists 1,620 tests; the tagged binary lists 1,622 and
adds exactly:

- `TestBytecodeDynamicIntBoxCacheReuseRecordsLookupHitAndInsert`; and
- `TestBytecodeDynamicIntBoxCacheReuseRecordsI64Bypass`.

Both tagged tests pass. No normal semantic test disappears under the default
build.

JS/WASM-only constraints and the deferred WASM workspace were excluded from
the audit.

## Package and runner topology

The native non-WASM module contains 27 packages:

- 20 own Go test files; and
- seven are support/command packages with no test files.

The seven zero-test-file packages are `cmd/parse-module`, four
`internal/semanticabi/cmd/*` report/generator commands,
`pkg/parser/language`, and `pkg/testcli`. They are ordinary support packages,
not packages whose tests were removed by build selection. Their behavior is
covered through parser tests, semantic-ABI checks/generators, and `cmd/able`
integration respectively.

The default compiler listing contains 836 unique top-level tests. A batch size
of 25 partitions them into exactly 34 shards with no duplicate or omitted
name. The default runner then provides:

- the standalone parser binding;
- the explicit full-mode parser fixture-corpus lane;
- every remaining non-compiler package in short mode;
- all 34 compiler shards; and
- the complete bytecode fixture pass.

No runner selection produced an unexpected zero-test semantic package.

## Dynamic verification

Representative tests for every return/build classification pass:

- cross-process compiled-test cache locking: 0.152 seconds;
- locally scoped dynamic-import discovery: 0.003 seconds;
- concurrent ArrayStore statistics polling: 0.016 seconds;
- bytecode instruction-search and iterator-close polling group: 0.070
  seconds;
- extern-plugin child/build/rebuild coordination: 6.879 seconds; and
- both `able_bytecode_box_reuse` diagnostic tests: 0.067 seconds.

Every focused invocation used a one-minute bound.

The default runner SHA-256 remains
`d81d9cbc49fe661bf50c5b7662cb76609923a2853a1edda09378382c11036e6f`,
identical to the immediately preceding complete green handoff. No executable
source changed after that full run, so repeating the multi-minute identical
suite would add no new evidence.

The independent scoreboard, frontier, and ledger checks pass with 130 fully
sampled rows, zero actionable groups, 23 current closures, zero invalidations,
and an empty selector.

## Scope and cleanup

No test, test runner, parser, AST, compiler, generated runtime, runtime
package, interpreter, bytecode VM, canonical stdlib, benchmark, language,
dependency, fixture, or WASM source changed.

Local `HEAD` and `origin/master` remain
`9c32f2777536da2c948327720acc75187973a6d9` with zero divergence. The index is
empty. Nothing was staged, committed, or pushed.

All audit and focused-test state used an exact 458,224 KiB disk-backed
`/var/tmp` workspace. It was removed, and no task-owned cache remains.

The machine-readable companion is
`2026-07-30-silent-go-test-suppression-audit-closure.json`.

## Next

Create an exact, non-mutating retained/deferred release inventory for the
current worktree.

Why: the completed correctness tranches now form 18 intentional non-WASM
paths beside the unchanged 34-path deferred WASM boundary.

What it entails: record every current path's state, disposition, governing
record, line count, byte count, and SHA-256 in JSON, Markdown, and TSV; verify
the 34 deferred identities against their published inventory; rerun format,
size, scope, evidence, Git, and cleanup gates. Do not stage, commit, or push.

Why it matters: a later consolidation can include exactly the retained
correctness work without absorbing deferred WASM files or relying on a
hand-maintained path list.
