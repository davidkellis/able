# Concurrent Stateful Pipeline application gate — 2026-07-23

## Decision

Retain the portable `concurrent_stateful_pipeline` application, its
source-equivalent Go/Python/Ruby implementations, exact verifier, catalog and
coverage memberships, two complete measurement cohorts, and bounded profiles.
Retain no compiler, generated-runtime, bytecode VM, tree-walker,
canonical-stdlib, language, dependency, or WASM change.

The workload raises concurrency × functions/closures × interface dispatch from
four to five applications using a different topology from the preceding
worker pools. Compiled execution reproduces the closed goroutine-identity
owner. Bytecode reproduces the established concurrent environment/cache-lock,
integer, call, nominal-field, frame/return, and dispatch families. No new
generic implementation candidate passes the broad admission rule.

## Application contract

The application generates 4,096 nominal frames in memory, then sends them
through one producer and three ordered concurrent stages. Each stage is held
through the user-defined `StatefulStage` interface. Its interface method
receives a captured reducer callback that consumes and returns the stage's
explicit `StageState`, so each stage carries count, total, and checksum state
from one frame to the next. One stage uses a state-sensitive window branch and
the other two use different affine parameters.

The final collector verifies every stage and Future count, four lane counts,
three independent stage states, and schedule-independent output aggregates:

```text
4096:4096:4096:4096,4096,4096:1024,1024,1024,1024:711726:931393:107993:489486,499007:726169,351558:141038,412661
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that exact output. Its SHA-256 is
`b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67`.
The catalog uses the goroutine executor, assigns four logical CPUs to
compiled/Go and one to interpreter lanes, and isolates the explicit source
root. Canonical and external Able sources are byte-identical with SHA-256
`99c6ecfb4eccd2c93925701f8a9845a53de88b9ac042ba6d34517ca073922bf5`.
No external input file or new stdlib API is required.

## Coverage result

The application genuinely covers lexical bindings and nullable channel
patterns; nominal values and generic arrays/channels; expressions and array
updates; captured functions passed as data; state-sensitive control flow;
user-defined interfaces and implementations; nullable Channel receive
handling; spawn, Future, and Channel lifecycles; packages/imports; stdlib
protocols; and real program entry.

The promoted catalog contains 55 portable applications and 110 status rows.
Both modes are selected, so the strict frontier contains 55 compiled and 48
bytecode rows. The priority concurrency × functions/closures × interface
dispatch triple increases from four to five independent applications.

## Repeated measurements

Every lane received two independent five-process cohorts, and every sample was
retained. All 50 timed processes passed the exact verifier with zero failures
and zero timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.816000 s | 0.8140 s | 0.8180 s | 187.849× Go |
| Go 1.26 | 10 | 0.004344 s | 0.004382 s | 0.004306 s | — |
| Able bytecode | 10 | 0.346000 s | 0.3500 s | 0.3420 s | 5.712× Python / 7.110× Ruby |
| Python 3.14 | 10 | 0.060577 s | 0.061585 s | 0.059570 s | — |
| Ruby 4.0 | 10 | 0.048661 s | 0.049203 s | 0.048120 s | — |

Cohort means moved by 0.49% for compiled, 1.75% for Go, 2.31% for bytecode,
3.33% for Python, and 2.23% for Ruby. The low movement is useful workstation
evidence, but every selected and pooled ratio remains an unambiguous target
miss.

## Ownership and admission

Three goroutine-executor compiled profiles merge to 5.58 seconds of CPU
samples. `bridge.currentGID`/`runtime.Stack` owns 96.42% cumulatively.
`run_stage`, both interface adapters, all three captured reducers, and Channel
send/receive paths sit beneath that wall. This independently reproduces the
exact generic compiled-concurrency owner. Its fixed-context replacement
already failed broad concurrent and serial guards, so it is not retried.

Three clean warmed bytecode profiles run ten application calls apiece and
average 181,874,040 ns/op, 24,909,291 B/op, and 448,764 allocs/op. They merge
to 5.44 seconds of CPU samples. `runResumable` is 91.91% cumulative,
`execCallOpcode` 26.10%, `execBinary` 14.52%, and
`structDefinitionNamedFieldIndex` 3.49%. The largest flat leaf,
`sync/atomic.(*Int32).Add`, is 9.56%; call-tree inspection attributes it to
`RWMutex` reader bookkeeping, not executor pending counters. The exact leaf
also appears in Concurrent Stencil Reduction, Concurrent Signal Dispatch,
Concurrent Transform Chain, Concurrent Policy Callbacks, and earlier
dependency-wave evidence. That environment/cache-lock family has already been
reconciled across unlike applications and does not justify retrying a rejected
cache or environment design.

The separately instrumented application records 57,363 inline-call hits and
zero misses, plus 45,061 resolved member inline hits and three fallbacks.
Residual integer extraction, arithmetic, nominal-field access, frames/returns,
and member lookup repeat completed generic families rather than exposing a new
exact three-application descendant.

## Evidence

- two Go, Python/Ruby, and Able cohorts:
  `2026-07-23-concurrent-stateful-pipeline-{go-reference,interpreter-reference,comparison}-{a,b}.{json,md}`;
- clean compiled merged profile:
  `.profiles/20260723_concurrent_stateful_pipeline_compiled_merged.cpu.pprof`;
- clean warmed bytecode merged profile:
  `.profiles/20260723_concurrent_stateful_pipeline_bytecode_runtime_merged.cpu.pprof`;
- readable profile tables and separate inline counters:
  `2026-07-23-concurrent-stateful-pipeline-{compiled,bytecode}-profile-top.txt`
  and
  `2026-07-23-concurrent-stateful-pipeline-bytecode-runtime-stats.json`.

## Verification

- exact output parity in tree-walker, bytecode, compiled Able, Go, Python, and
  Ruby;
- ten verifier-backed timed processes per compiled, bytecode, and reference
  lane;
- three clean compiled and three clean warmed bytecode profiles;
- focused catalog, selection, coverage, operation-depth, matrix, triple, and
  scorecard checks;
- every added source file remains below 1,000 lines;
- JSON, source-identity, syntax, and whitespace checks.

## Next recommendation

Recompute and promote the governed performance and interaction frontiers,
then choose the highest-excess minimum-depth interaction rather than assuming
the same triple remains uniquely shallow.

Why: this application raises the requested interaction to depth five and
reproduces closed owners. At depth five another interaction can tie the
frontier, so the next workload should improve the weighted frontier instead of
adding another variant of this ordered state pipeline.

What it entails: reconcile the 55-application scorecard and closure ledger,
rank every minimum-depth triple by adjacent target excess, then build one
source-equivalent deterministic Able/Go/Python/Ruby application only if it
adds a materially different semantic shape. Take two five-process cohorts and
profile after parity. Admit runtime or compiler work only for an exact generic
owner repeated across at least three unlike applications. Update canonical
`able-stdlib` only for a reusable API or correctness defect, and do not begin
WASM work.
