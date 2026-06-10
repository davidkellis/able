# Concurrent State Machines application gate — 2026-07-23

## Decision

Retain the portable `concurrent_state_machines` application, its
source-equivalent Go/Python/Ruby implementations, exact verifier, catalog and
coverage memberships, two complete measurement cohorts, and bounded profiles.
Retain no compiler, generated-runtime, bytecode VM, tree-walker,
canonical-stdlib, language, dependency, or WASM change.

The workload raises both shallow callable/interface interactions using four
independently spawned state machines rather than another worker pool or
ordered channel pipeline. Compiled execution reproduces the closed
goroutine-identity owner. Bytecode reproduces established environment/cache
locking, integer extraction and arithmetic, member lookup, nominal-field,
frame/return, and dispatch families. No new generic implementation candidate
passes the broad admission rule.

## Application contract

Four independent state machines each carry an immutable nominal
`MachineState`. Every transition calls the state's inherent `advance` method,
then selects either `GateHandler` or `CycleHandler` through the user-defined
`StateHandler` interface. The handler receives a captured adjustment closure
on every transition. Four Futures are joined only after all state machines
complete, so output is schedule-independent.

The final result verifies all completion and transition counts, accepted and
redirected totals, four independent checksums, four final values, and
cross-machine aggregates:

```text
8192:8192:7166:1026:732185,669476,501158,348726:567906,647364,128740,255598:251539:653272
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that exact output. Its SHA-256 is
`96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103`.
The catalog uses the goroutine executor, assigns four logical CPUs to
compiled/Go and one to interpreter lanes, and isolates the explicit source
root. Canonical and external Able sources are byte-identical with SHA-256
`a8270e45edd0da619ec0c49ff1c09bf80a920e37ceff046e1887963059ea221e`.
No external input file or new stdlib API is required.

## Coverage result

The application genuinely covers lexical bindings; nominal values and
construction; inherent methods; captured first-class functions; loop,
branch, and match control flow; user-defined interfaces and implementations;
interface-selected method calls; spawn and Future lifecycles; packages and
imports; stdlib protocols; and real program entry.

Before scorecard promotion, the catalog contains 56 portable applications.
Both modes are selected, so the strict selection manifest contains 56
compiled and 49 bytecode rows. Both concurrency × functions/closures ×
interface dispatch and functions/closures × inherent methods × interface
dispatch increase from five to six independent applications.

## Repeated measurements

Every lane received two independent five-process cohorts, and every sample was
retained. All 50 timed processes passed the exact verifier with zero failures
and zero timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.262000 s | 0.2640 s | 0.2600 s | 66.405× Go |
| Go 1.26 | 10 | 0.003946 s | 0.003746 s | 0.004145 s | — |
| Able bytecode | 10 | 0.292000 s | 0.3020 s | 0.2820 s | 4.923× Python / 5.430× Ruby |
| Python 3.14 | 10 | 0.059316 s | 0.060228 s | 0.058403 s | — |
| Ruby 4.0 | 10 | 0.053779 s | 0.054800 s | 0.052757 s | — |

Cohort means moved by 1.53% for compiled, 10.10% for Go, 6.85% for bytecode,
3.08% for Python, and 3.80% for Ruby. The Go reference is the noisiest lane,
so the decision uses all ten retained samples as requested. Every selected and
pooled ratio remains an unambiguous target miss.

## Ownership and admission

Three goroutine-executor compiled profiles merge to 1.50 seconds of CPU
samples. `bridge.currentGID` owns 96.67% cumulatively and descends through
`runtime.Stack`; all four `run_machine` activations, both interface adapters,
both implementations, and all four captured adjustment closures sit below
that wall. This independently reproduces the exact generic
compiled-concurrency owner. Its fixed-context replacement already failed
broad concurrent and serial guards, so it is not retried.

Three clean warmed bytecode profiles run ten application calls apiece and
average 89,875,426 ns/op, 12,286,543 B/op, and 210,112 allocations/op. They
merge to 2.69 seconds of CPU samples. `runResumable` owns 97.77%
cumulatively, `execCallOpcode` 30.86%, `execCallMember` 27.51%,
`execBinary` 24.16%, and `execLoadSlotStructField` 6.69%. The largest
non-dispatch leaf, `sync/atomic.(*Int32).Add` at 8.92%, is attributed by the
call tree to `RWMutex` reader bookkeeping. Integer extraction,
named-struct-field planning, arithmetic, inline calls, and member caches are
the same completed generic families found in unlike prior applications.

The separately instrumented process records 24,586 inline-call hits with zero
misses and 16,384 resolved member-inline hits. The remaining four resolved
fallbacks are setup-scale, not a hot missed lowering. There is no new exact
three-application descendant, so no implementation candidate is admitted.

## Evidence

- two Go, Python/Ruby, and Able cohorts:
  `2026-07-23-concurrent-state-machines-{go-reference,interpreter-reference,comparison}-{a,b}.{json,md}`;
- clean compiled merged profile:
  `.profiles/20260723_concurrent_state_machines_compiled_merged.cpu.pprof`;
- clean warmed bytecode merged profile:
  `.profiles/20260723_concurrent_state_machines_bytecode_runtime_merged.cpu.pprof`;
- readable profile tables and separate inline counters:
  `2026-07-23-concurrent-state-machines-{compiled,bytecode}-profile-top.txt`
  and
  `2026-07-23-concurrent-state-machines-bytecode-runtime-stats.json`.

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

Promote and reconcile the governed performance and interaction frontiers,
then choose the highest-excess minimum-depth interaction from the resulting
weighted frontier.

Why: this application raises both previously shallow interactions to depth
six and reproduces only closed owners. The scorecard ratios can change which
remaining minimum-depth interaction has the greatest adjacent target excess,
so the next application must be selected from regenerated evidence.

What it entails: add the two selected rows to the current scorecard, refresh
performance and interaction frontiers, review the affected concurrency
closures, and regenerate the deterministic architecture/ABI dependency chain.
Then build another source-equivalent application only if it adds a materially
different semantic shape. Take two five-process cohorts and profile only after
six-lane parity. Admit runtime or compiler work only for an exact generic
owner repeated across at least three unlike applications. Update canonical
`able-stdlib` only for a reusable API or correctness defect, and do not begin
WASM work.
