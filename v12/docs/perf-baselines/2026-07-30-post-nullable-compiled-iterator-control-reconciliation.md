# Post-nullable compiled iterator/control reconciliation

## Decision

Reconcile `compiled-iterator-control` as causally current and retain no
production change.

The primitive nullable value carrier materially reaches four of the eight
rows: Binary Event Log, Dependency Plan, Generic Slot Buffer, and
Option/Result Config. The other four rows have no application or material
imported primitive-carrier reach.

The reached rows do not expose one post-carrier generated-code or
generated-runtime owner in three unlike applications. No new CPU, allocation,
or timing cohort was warranted under the tranche's selective profile gate.

## Strict boundary and execution gate

Every application was rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Array Slice Window | 96 | absent | 1/1 |
| Binary Event Log | 96 | absent | 1/1 |
| Dependency Plan | 96 | absent | 1/1 |
| Discrete Event Simulation | 96 | absent | 1/1 |
| Document Audit | 96 | absent | 1/1 |
| Generic Slot Buffer | 96 | absent | 1/1 |
| Lexical Rollup | 96 | absent | 1/1 |
| Option/Result Config | 96 | absent | 1/1 |

Document Audit and Lexical Rollup use input paths rooted at the shared
external benchmark directory. Their first smoke attempts used the
application directory and failed before verification because the input files
were absent there. The same preserved binaries passed from the exact catalog
directory. Failed setup attempts are excluded from the gate.

Smoke durations are execution checks, not timing evidence. The authoritative
scorecard retains five verifier-backed Able and Go processes per row, and the
Generic Slot Buffer nullable decision retains its own balanced five-process
baseline/candidate/Go cohorts.

## Causal carrier reach

Generated support definitions were excluded. Every generated module defines
primitive nullable conversion helpers, but a dormant definition cannot affect
an application.

| Application | Material reach | Exact reached path |
| --- | --- | --- |
| Array Slice Window | no | Array reads are directly unwrapped; `slice` returns `Result<Array<i32>>` |
| Binary Event Log | yes | `map_count -> HashMap<i64,i64>.raw_get -> __able_nullable[int64]` |
| Dependency Plan | yes | resolver loop calls `Queue<i32>.dequeue -> Deque<i32>.pop_front -> __able_nullable[int32]` |
| Discrete Event Simulation | no | `EventQueue.pop` returns nullable non-primitive `Scheduled<SimulationEvent>` |
| Document Audit | no | application text/filter/map/reduce path contains no primitive nullable carrier |
| Generic Slot Buffer | yes | `VersionedBuffer<i64>.get` and `ReadableBuffer<i64>.read` use `__able_nullable[int64]` |
| Lexical Rollup | no | application text/filter/map/reduce path contains no primitive nullable carrier |
| Option/Result Config | yes | service/regional `Option<i32>` and specialized Option methods use `__able_nullable[int32]` |

The public Binary Event Log result proves 53,248 accepted records. Its final
23-key source scan makes `map_count` reach the retained value carrier at least
53,271 times.

Dependency Plan processes all 1,024 services and one terminating absent
dequeue in each of 12 resolver rounds, for 12,300 queue-carrier returns.

Option/Result Config audits 1,024 services for 24 rounds. It therefore
constructs 24,576 service `Option<i32>` values. The regional fallback is
requested for the 94 service identifiers divisible by 11, or 2,256 additional
nullable returns.

Generic Slot Buffer is governed by the retained direct A/B rather than an
inferred reach count. Its five-run mean improved from 0.0560 to 0.0340
seconds, and its exact main allocation objects fell from 264,215 to 1,046.
The former 131,842 flat `VersionedBuffer_get_spec` pointer constructions
became zero.

## Residual-owner admission

The four reached rows separate below generic allocation and GC ancestry:

- Generic Slot Buffer's post-carrier main phase allocates only 17,872 bytes in
  1,046 objects. Its storage, interface adapter, iteration, and arithmetic are
  direct native Go carriers.
- Binary Event Log's retained concrete owners are nominal `EventRecord`
  conversion and ordinary integer-boundary work. Its earlier generic-union
  bound-method allocation was already removed by the retained direct-known
  method rule; the remaining nominal layout does not have three-application
  breadth.
- Dependency Plan remains direct graph/Queue work plus common integer boxes.
  Removing the nullable Queue return does not turn its Queue/graph descendants
  into a general leaf shared with the other reached rows.
- Option/Result Config remains led by static generic-union/match work and
  allocation. The direct-known method and static match-construction routes
  already completed broad gates and are not reopened by a carrier-shape
  change.

`runtime.mallocgc` and collector frames are consequences of those different
operations, not a shared compiler rule. Array, named-container, non-primitive
nominal, call/member, frame, stack, checked-arithmetic, GC, and global
execution-context routes remain closed. With no plausible exact successor at
breadth three, the selective profile gate correctly retained zero new
profiles and no speculative candidate.

## Current row state

The current full-scorecard snapshots remain:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Array Slice Window | 0.0300 s | 0.0044 s | 6.8182x |
| Binary Event Log | 0.1640 s | 0.0078 s | 21.0256x |
| Dependency Plan | 0.0200 s | 0.0038 s | 5.2632x |
| Discrete Event Simulation | 0.0420 s | 0.0138 s | 3.0435x |
| Document Audit | 0.0380 s | 0.0041 s | 9.2683x |
| Generic Slot Buffer | 0.0760 s | 0.0052 s | 14.6154x |
| Lexical Rollup | 0.0660 s | 0.0041 s | 16.0976x |
| Option/Result Config | 0.0460 s | 0.0038 s | 12.1053x |

The Generic Slot Buffer scorecard snapshot is not substituted for its direct
nullable A/B. The balanced retained A/B is the causal optimization gate; the
full scorecard remains the product-wide snapshot.

## Scope

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No named-container, non-primitive nominal, or benchmark-specific rule was
introduced.

The machine-readable record is
`2026-07-30-post-nullable-compiled-iterator-control-reconciliation.json`.
After verification, the exact 921 MiB disk-backed generated-module, binary,
compiler, and Go-cache workspace was removed. No matching tranche artifact
remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-text-map` against the retained nullable carrier.

Why: Inventory Reconciliation and Transaction Ledger Audit are two of the
three applications in the carrier's direct retained A/B, and both belong to
that closure. It is therefore the next invalidated closure with proven causal
reach.

What it entails: attach their retained timing and exact allocation evidence,
strictly rebuild the nine text/map rows, and audit application plus material
imported paths for primitive-carrier reach. Profile only a reached row whose
remaining exact owner could repeat across three unlike applications. Keep
explicit dynamic-map conversions classified as semantic boundaries, and do
not add a `HashMap` or other named-container lowering rule.

Why it matters: this advances the compiler evidence closest to the retained
change while testing whether the removed nullable boxes reveal a genuinely
general successor. It continues the path toward native-Go compiled
performance without manufacturing breadth from map-heavy applications.
