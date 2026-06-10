# Concurrent document-pipeline performance gate

## Purpose

Concurrent Document Pipeline is a portable application benchmark for two
previously shallow feature interactions:

- concurrency × expressions/files × closures/callables;
- concurrency × closures/callables × real program entry.

It is an application-coverage addition, not a benchmark-specific optimization
vehicle. Its source shape is a file-driven transform with four long-lived
workers, captured scoring callbacks, nominal task/result values, and a
schedule-independent collector.

## Portable contract

The canonical Able source and the sibling Able, Go, Python, and Ruby programs
read the same 32-line input and perform 32 rounds. Each implementation emits
the same counts, total, and checksum through one public verifier. Four workers
are part of the contract; ordering of worker completion is not.

The sibling Able package deliberately has a different package name from the
canonical example. The current loader can see both roots during benchmark
harness execution, so identical package names would collide even though the
program bodies are source-equivalent.

The scale was fixed before retained measurement. A 128-round tree-walker smoke
approached the project timeout, so the workload was reduced to 32 rounds. This
keeps every correctness execution below one minute while retaining 1,024 tasks.

## Evidence rule

Feature interaction depth and performance-candidate admission remain separate
decisions:

1. A real application may be retained when it closes a portable coverage gap
   and passes cross-runtime correctness.
2. Repeated timings determine current performance, with every successful
   volatile-workstation sample retained in the arithmetic mean.
3. A code candidate still requires one concrete mechanism reproduced across
   unlike applications and broad guards that do not regress.

For this application, compiled profiling reproduced `bridge.currentGID` /
`runtime.Stack`, already a shared concurrency owner. Its generic fixed-context
alternative was previously rejected because it regressed N-Body. The bytecode
timing did not identify a new exact VM child beyond already-profiled channel,
call, member, return, and typed-match paths. Therefore neither mode admitted a
new code candidate.

## Consequence

The benchmark remains because it improves portable application coverage and
adds current performance evidence. No compiler, runtime, VM, stdlib, language,
or WASM change follows merely from its large ratios. The next application
should broaden the depth-two expressions/files × closures × Option/Result
frontier with a non-routing validation/transformation shape, then apply the
same independent candidate gate.

## 2026-07-21 nullable-result reconciliation

The later interaction audit found that this application already executes a
substantial Option path that the initial feature annotation omitted.
`Channel.receive()` returns `?DocumentTask`; the four workers match 1,024 task
values and four closed-channel `nil` values during every normal run. The
`option_result_exceptions` membership is therefore part of the existing
application semantics, not a new workload feature.

Only the coverage manifest, reconstructed-baseline contract, and derived
triple frontier change. Executable performance evidence does not: no source,
input, verifier, runtime, compiler, VM, or stdlib dependency changed. This
raises concurrency × expressions/files × Option/Result from depth two to three
and leaves four depth-two concurrency interactions for the next source audit.
See
`docs/perf-baselines/2026-07-21-concurrent-document-pipeline-option-reconciliation.md`.
