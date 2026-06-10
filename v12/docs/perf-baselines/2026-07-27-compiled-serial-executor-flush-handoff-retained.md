# Compiled serial-executor Flush handoff retained

Date: 2026-07-27

## Decision

Retain one general compiler-generated runtime correctness correction:
`__able_serial_executor.Flush` now observes the worker's locked
dequeue-to-active handoff.

This mirrors the reference interpreter correction in
`2026-07-27-serial-executor-flush-handoff-retained.md`. It is not a
performance optimization and does not add a compiler/interpreter boundary,
container-specific lowering, benchmark branch, or new runtime ABI.

## Root cause

The generated worker removed a runnable task from `queue` under the executor
lock, released that lock, and only later published the task through `active`.
During that interval, `Flush` could observe an empty queue and no active task
and return before the already-dequeued future ran.

The earlier interpreter audit corrected the same state transition in the
reference serial executor. Reviewing the compiler-owned runtime found that
the generated copy still had the old transition, invalidating the assumption
that both executor implementations shared the same Flush semantics.

## Retained rule

The generated executor now:

1. increments `workerInFlight` while dequeuing a worker-owned task under the
   executor lock;
2. makes `Flush` wait while that handoff count is nonzero;
3. publishes `current`, `active`, and `paused` under the same lock before
   decrementing the handoff count; and
4. broadcasts after the task becomes visible as active.

Only the invisible handoff is counted. A task that later pauses on a channel,
mutex, timer, await registration, or other external wake remains blocked work
that `Flush` may leave. `Drive`-stolen tasks do not increment the worker
handoff count.

The task becomes visible before its terminal-status check. This also closes
the handoff for an already-terminal future without leaving a stale count.

## Source organization

The future renderer was already near the 1,000-line project limit. Its
declaration prelude moved intact to
`generator_render_runtime_future_declarations.go`, leaving
`generator_render_runtime_future.go` at 992 lines. The extraction has no
generated semantic effect beyond preserving the same declaration order.

The concurrency design record now explicitly states that the compiler emits
the same locked handoff invariant as the reference executor.

## Verification

Focused generated-runtime verification passed:

- the source-shape guard confirms the generated field, dequeue increment,
  Flush predicate, worker/Drive distinction, active publication, and
  decrement;
- a generated Go test deterministically pauses between dequeue and active
  publication and proves that Flush waits through both phases;
- the same generated test proves that Flush returns once the handed-off task
  is paused;
- the emitted test passed under `go test -race -count=20` in 28.145 seconds;
- the generated goroutine-Flush notifier and blocked-task parity tests pass;
- native primitive, static Array, shared nominal, direct-kernel,
  no-bootstrap, dynamic-boundary, and unlike benchmark-source guards pass;
- all changed or untracked compiler-boundary Go files are `gofmt` clean and
  below 1,000 lines; and
- `go vet` passes for the compiler, bridge, semantic ABI, and compiler-facing
  CLIs.

Three unlike strict applications—Fib, Sudoku Masks, and Tapelang Alphabet—
compiled with `--no-fallbacks`. Each final graph contains 96 packages and
omits `able/interpreter-go/pkg/interpreter`; their top-level generated Go
sources contain no interpreter import. Nine independent executions passed the
public verifiers. A second Fib emission reproduced all 14 generated Go files
byte-for-byte.

The explicit compiled-CLI lane rebuilt and passed all 42 cases with the new
compiler identity in 19:31.59 at 3,399,964 KB peak RSS. The complete
`./run_all_tests.sh` handoff passed in 15:14.81 at 4,545,104 KB peak RSS,
including all 33 compiler batches and the 97.109-second bytecode fixture
corpus. Canonical stdlib tests passed in 50.25 seconds at 856,468 KB peak RSS;
tree-walker reported 20 seconds and bytecode 19 seconds.

## Scope

No Able stdlib, external reference, benchmark, fixture expectation, language,
dependency, interpreter, bytecode VM, or WASM source changed. The generated
runtime correction is the only production behavior authored by this compiler
boundary review.

## Next recommendation

Perform the fourth dependency-ordered release-consolidation review for
canonical documentation, examples, benchmarks, performance evidence, and
shared held spec/design paths.

Why: language/parser/typechecker, runtime/reference engines, and
compiler/AOT/semantic-ABI boundaries are now audited. The remaining visible
delta is evidence and cross-boundary material that must be classified before
any maintainer-authorized staging or commit.

What it entails: inventory the remaining changed and untracked documentation,
examples, benchmark contracts, generated evidence, and held shared paths;
verify hashes and references; distinguish authoritative records from
disposable artifacts; and emit the final manifest without touching deferred
WASM or Git history.

Why it is important: it completes the review chain needed to preserve the
native-Go, interpreter-free compiled contract while making the extremely
dirty long-running worktree safe for maintainer review. Performance mutation
remains paused unless new evidence identifies one exact general owner in
three unlike applications.
