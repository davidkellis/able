# Compiler spawn-gated callable-context activation

## Status

Retained on 2026-07-29. This is the measured follow-on to
`compiler-callable-context-default-activation.md`.

The compiler now selects the existing callable execution-context ABI when any
statically loaded module contains either `await` or `spawn`. It does not
enable the ABI globally and does not change the ABI itself.

The performance decision is recorded in
`../docs/perf-baselines/2026-07-29-compiled-spawn-gated-callable-context-retained.md`.

## Decision

Treat `spawn` as a scheduler-context requirement at the loaded-program-graph
level.

This is a general syntax and scheduler rule:

```text
schedulerContextRequired =
    loadedProgramContainsAwaitOrSpawn(modules)
executionContextsEnabled =
    forceExperimentalExecutionContext || schedulerContextRequired
callableExecutionContextsEnabled =
    schedulerContextRequired
```

Scanning all loaded modules is intentionally conservative. It covers imported
and nested spawn expressions without depending on entry-package placement or
compiler options. A reachability refinement is not justified by the current
evidence and must not be introduced from application names or measured runtime
counts.

Spawn-free and await-free programs retain their previous generated source.
Dynamic, host-created, nil-context, and incompatible-environment entries keep
the compatibility path described in the original callable-context design.

## Generated-helper closure

Spawn-gated activation exposed a general generated-code dependency:
context-aware Channel, Mutex, Future, and other Awaitable surfaces can require
the native await-waker and registration helpers even when the source contains
no explicit `await` expression.

The helper emission condition therefore follows callable-context activation,
not only the compiler's list of source-level await expressions. An imported,
spawn-only regression test guards this dependency.

This is not a new runtime path. It makes the already-selected context-aware
native interfaces self-contained.

## Evidence and trade-off

The complete strict catalog proves the gate is narrow:

- all 63 baseline and 63 candidate applications pass their public verifier;
- all 126 generated graphs remain interpreter-free;
- exactly 20 generated sources change and select callable context;
- 39 spawn-free programs and the four already await-gated programs remain
  byte-identical; and
- all changed rows contain `spawn`.

Repeated paired timing improves 13 of the 20 reached applications. The
cross-application paired geometric ratio is 0.884258, or 11.57% faster.
Removing the dominant Mutex Ledger win leaves a 0.970408 ratio, or 2.96%
faster across the other 19.

Seven short applications measure 1.18%-7.20% slower after fifteen paired
samples. These counterexamples are retained in the decision record. They do
not justify a narrower production gate: choosing by application identity,
nominal type, source family, or observed boundary-count threshold would be a
benchmark-specific rule. The general syntax gate materially helps unlike
programs, removes the shared boundary owner in 19 of 20 programs, and has no
material broad allocation penalty.

Binary Trees is the long-running guard. Across fifteen rotating
baseline/candidate/Go cohorts, the candidate's paired A/B geometric ratio is
0.988248 and its candidate/Go ratio is 1.023852. It therefore delivers 97.67%
of equivalent Go performance and preserves the 95% target.

## Invariants

- Primitive and static Array carriers remain native Go values.
- The context pointer does not box arguments, returns, captures, or receivers.
- Spawned tasks own child context and payload state.
- Package-context localization preserves scheduler payload.
- Static context-aware calls do not recover goroutine identity.
- Compatibility entries remain available at actual dynamic and host
  boundaries.
- Strict fallback-free graphs do not link `pkg/interpreter`.
- Non-primitive nominal types continue through shared nominal translation.
- No application, benchmark, named-container, or stdlib-package predicate
  participates in activation.

## Closed alternatives

- Do not flip execution context on globally.
- Do not activate from a list of concurrency applications.
- Do not use measured `currentGID` counts as a compiler predicate.
- Do not add named-container or non-primitive nominal branches.
- Do not remove dynamic or host compatibility entries.
- Do not add another callable ABI.

Any future narrowing or expansion needs new source-shape, semantic, allocation,
and repeated cross-program timing evidence.
