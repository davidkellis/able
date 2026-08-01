# Exact-`Self` benchmark coverage audit closure

Date: 2026-07-30

## Decision

**Close the native-dictionary exact-`Self` performance direction until a
natural application or production identity makes the path materially
reachable. Retain no benchmark or production code.**

The complete 66-application compiled/bytecode selection contains 12
application-defined interfaces across 11 sources, but none has a method whose
result is exact top-level `Self`. The sibling external suite contains 67
verifier-backed benchmark directories: the same 66 portable applications plus
the diagnostic legacy Sudoku workload. Excluding copied build output under
`target/`, it also contains zero exact-`Self` result declarations.

The audit therefore confirms a sustained performance-coverage gap, but it
does not identify a natural existing workload to admit. Adding or rewriting
an application solely to execute the new carrier path would be the synthetic
timing fixture prohibited by the admission rule.

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency, benchmark, fixture, frozen workspace,
or WASM source changed.

The exact 16 KiB of task-created `/var/tmp` audit lists were removed after the
compact evidence was retained. No `/tmp/able-*` entry or newly generated
Python cache remained.

## Scope and method

The authoritative selection manifest has SHA-256
`5e05af7e06f328284c91639fc58e8a4ff822c637fd641b36943505dfb3822097`.
Its compiled and bytecode modes each name the same 66 unique applications.
The current scoreboard has SHA-256
`3652695dc7b1576ed4729ef30a7688b171114cda9b4ce269132fd868b37849f3`.

The audit:

1. scanned every selected Able source, including multiline result syntax, for
   exact top-level `Self` results;
2. enumerated all application-defined interface declarations and their method
   results;
3. scanned all sibling `run.able` sources while excluding generated
   `target/` trees;
4. checked imported canonical `Clone`, `Default`, `Extend`, `Numeric`, and
   `Fractional` exact-`Self` surfaces for application calls through interface
   values; and
5. ranked the strongest state-machine, visitor, transformation, generic, and
   numeric candidates using current five-run compiled measurements, direct
   native-interface counts, reference availability, and verifier quality.

All 67 sibling benchmark directories have a public `verify.rb`. The only
benchmark not in the portable 66-row selection is legacy `sudoku`, already
classified as diagnostic because it is not source-equivalent to the selected
exact-cover references. `able-base` is suite infrastructure, not a benchmark.

## Catalog result

| Measure | Result |
| --- | ---: |
| Selected applications | 66 |
| Application sources declaring an interface | 11 |
| Application-defined interfaces | 12 |
| Exact top-level `Self` interface results | 0 |
| Sibling verifier-backed benchmark directories | 67 |
| Sibling source exact-`Self` results, excluding `target/` | 0 |

The 12 application interfaces are `Oscillator`, `GraphVisitor`, `PacketCodec`,
`DecisionPolicy`, `DistanceField`, `SignalKernel`, `StateHandler`,
`StatefulStage`, `FoldAlgebra`, `ReadableBuffer`, `CalibrationPolicy`, and
`ReadableWindow`. Their methods return primitives, optionals, or named state
and report values. None returns its receiver type.

Generated build trees do contain exact-`Self` declarations from the kernel and
canonical stdlib. Those are dependency copies, not application coverage.
Counting them would repeat the error this audit was intended to prevent:
generated support recurrence is not runtime reach.

## Imported exact-`Self` surfaces

The canonical stdlib exposes exact-`Self` through `Clone`, `Default`, `Extend`,
`Numeric`, and `Fractional`. Selected application source contains no
`.clone()`, `.extend()`, or iterator `.collect()` call.

`wide_integer_records` calls `min`, `max`, and `abs` on statically concrete
`UInt128` and `Int128` receivers. Those calls resolve through concrete
inherent/implementation methods; the application never stores a `Numeric`
interface value or returns the result through a captured interface
dictionary. This is useful nominal numeric coverage, but it cannot measure
exact-`Self` dictionary preservation.

## Candidate ranking

All six leading candidates have exact Able copies in the sibling suite,
Go/Python/Ruby references, public verifiers, five successful current compiled
runs, and zero current timeouts or failures.

| Rank | Application | Able / Go | Direct native adapters | Why it looked plausible | Admission result |
| ---: | --- | ---: | ---: | --- | --- |
| 1 | Versioned Telemetry Pipeline | 2.0680 / 0.2078 s | 3 | Sustained policy dispatch, generic window, iterator, `Result`, and 16.5M updates | Reject: its interfaces return `i64`, `i32`, or `?T`; changing them to `Self` would replace the workload contract |
| 2 | Generic Slot Buffer | 0.0760 / 0.0052 s | 1 | Generic interface constraint, nominal storage, aliasing, and iteration | Reject: `ReadableBuffer.read` naturally returns `?T`, not the buffer |
| 3 | Concurrent Stateful Pipeline | 0.0660 / 0.0056 s | 3 | Stateful interface dispatch, callbacks, nominal frames, and channels | Reject: each stage naturally returns `StageStep`; making the stage itself the result changes state ownership and pipeline semantics |
| 4 | Concurrent Graph Visitors | 0.0300 / 0.0044 s | 4 | Interface values, callbacks, nominal graph state, and futures | Reject: visitors naturally return `VisitState`, not themselves |
| 5 | Concurrent State Machines | 0.0260 / 0.0048 s | 4 | Interface-selected transitions, callbacks, immutable state, and futures | Reject: handlers naturally return `MachineState`; exact `Self` would redesign the machine |
| 6 | Wide Integer Records | 0.0740 / 0.0247 s | 2 | Sustained exact-`Self` numeric method names on nominal values | Reject: all receivers and results remain statically concrete, so no interface dictionary return is exercised |

The first five are genuinely broad applications, but none naturally owns the
receiver state that an exact-`Self` method would return. The sixth reaches
exact-`Self` method semantics only through concrete dispatch. Rewriting any
one would make the benchmark answer a compiler question rather than model its
existing application domain.

## Admission result

No candidate satisfies both required conditions:

- a sustained natural application contract; and
- repeated exact-`Self` return through an interface value.

Accordingly:

- no new benchmark was added;
- no existing benchmark was rewritten;
- no compiler candidate was implemented;
- no baseline/candidate/reference A/B cohort was manufactured; and
- the retained exact-`Self` dictionary correctness capability remains
  unchanged.

The machine-readable companion is
`2026-07-30-exact-self-benchmark-coverage-no-go.json`.

## Next

Refresh interpreter-free main-phase CPU and exact-allocation profiles for
`versioned_telemetry_pipeline`, `k_nucleotide`, and `sudoku_masks`, the three
largest current compiled absolute-excess workloads.

Why: the exact-`Self` direction has no natural runtime reach, while these
three sustained, unlike applications contribute the largest current compiled
time above their Go references: 1.8602 s, 1.5251 s, and 1.2476 s
respectively.

What it entails: rebuild normal strict binaries, verify that final dependency
graphs omit `pkg/interpreter`, collect repeated main-only CPU and exact
allocation evidence under the established affinity/memory protocol, and
select only one exact generated-code or generated-runtime owner that repeats
in all three. Exclude every already-closed checked-arithmetic, Array, frame,
stack, register, call/member/index, GC, launch-floor, and execution-context
route. Advance a change only through balanced five-or-more-pair
baseline/candidate/Go verification.

Why it matters: this returns compiled work to sustained application cost
rather than generated capability. A shared owner across numeric/text,
constraint-solving, and telemetry workloads would be strong evidence for a
general lowering rule; the absence of one should again result in no retained
code.
