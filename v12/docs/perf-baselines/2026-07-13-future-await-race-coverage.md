# Future Await Race Cross-Language Coverage — 2026-07-13

## Decision

Add the `future-await-race` application and one compiler correctness repair;
add no performance optimization or canonical-stdlib change. The prior gap
check overstated the remaining concurrency hole: Future Pipeline already
exercises cancellation and cooperative yielding. The missing cross-language
semantic boundary was the Future `Awaitable` protocol itself.

The application exposed that generated binaries running the goroutine executor
could only suspend `await` through the serial executor's resume channels. The
shared generated Await runtime now waits on the Await waker or the task's
Future cancellation context when those serial channels are absent, and reports
blocked/unblocked state to the executor. This is executor-correctness work for
every compiled pending Await, not an application or benchmark fast path.

## Application

Each of 96 rounds launches two yielding numeric tasks and a third task that
joins their values through two `await` expressions. A separate probe yields
until it observes a direct Future cancellation. The final summary is
independent of interleaving:

```
96:96412966:1
```

The canonical Able source is
`v12/examples/benchmarks/future_await_race/future_await_race.able`; the
sibling `../benchmarks/future-await-race` suite contains matching Able,
Go 1.26, Python 3.14, and Ruby 4.0 programs and one Ruby verifier.

The benchmark is deliberately unlike Future Pipeline. It has no producer or
worker pool, no named-container-specific path, and no privileged runtime API;
its material language boundary is repeated Future-await joining.

## Harness Integration

`v12/bench_external_catalog.sh` registers `future_await_race` in its own
suite and in `concurrency` and `coverage`. It uses the goroutine executor and
source-root isolation, because the canonical source and sibling Docker source
share a package name. The established `generality` selection suite remains
unchanged, and no timing scorecard is amended until fresh references exist.

## Verification

- Able bytecode and tree-walker completed the canonical source with the shared
  verifier.
- Compiled Able completed the sibling source under the goroutine executor and
  passed the shared verifier.
- The compiled-boundary audit completed one verified goroutine-executor run
  with no timeout or failure.
- Go, Python, and Ruby sources each completed and passed the same verifier.
- The compiler's goroutine parity lane now covers both a pending Future await
  and the existing fairness/cancellation fixture.

## Performance follow-up

The first pinned reference screen and the conditional profile comparison are
complete in
`v12/docs/perf-baselines/2026-07-13-future-await-race-reference-profile-gate.md`.
They keep no performance change: bytecode shares only executor/VM parents with
Channel Rollup and Future Pipeline, while compiled code repeats the known
`bridge.currentGID` / `runtime.Stack` wall whose generic fixed-context
candidate already failed broad serial guards.

## Next

Treat this as a completed application row and return to feature-led coverage.
The next boundary must be independently useful and receive the same shared-
fixture, cross-language, verifier-backed process before it becomes performance
selection evidence.
