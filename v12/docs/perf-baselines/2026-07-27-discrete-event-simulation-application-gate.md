# Discrete Event Simulation application gate

Date: 2026-07-27

## Decision

Retain the portable `discrete_event_simulation` application, its
source-equivalent Go/Python/Ruby implementations, exact verifier, catalog and
coverage memberships, repeated selected measurements, and bounded profiles.
Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, language, dependency, or WASM production change.

This application closes the missing application-defined generic-storage
control. Its strict generated program lowers nested `EventQueue T` and
`Scheduled T` specializations directly to Go structs, pointers, native integer
fields, and `[]*Scheduled_SimulationEvent`. The hot main and heap methods call
those specialized Go functions directly. They never box into `runtime.Value`
or cross into the interpreter.

The compact machine-readable companion is
`2026-07-27-discrete-event-simulation-application-gate.json`.

## Application contract

The workload is a binary-min-heap discrete-event scheduler. It seeds 4,096
events, processes 50,000 in timestamp/sequence order, and schedules a
deterministic successor for every processed event. Its application-defined
storage is:

```able
struct Scheduled T {
  at: i64,
  sequence: i64,
  value: T
}

struct EventQueue T {
  items: Array (Scheduled T)
}
```

The concrete program uses `EventQueue SimulationEvent`; an alias binding
observes the same mutated queue, so the benchmark checks reference identity as
well as field representation. It does not use HashMap, Channel, `dynamic`, or
a benchmark-only API.

Able tree-walker, Able bytecode, strict compiled Able, Go 1.26, Python 3.14,
and Ruby 4.0 all emit:

```text
50000:3364:4096:4096:12513:695126337
```

The output SHA-256 is
`6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405`.
The canonical and external Able sources are byte-identical with SHA-256
`f4229da6b02e7e79c49b81f266a08efb87476edade1255277a55fda77f10f22f`.

## Repeated measurements

All selected and reference measurements used independent processes on fixed
CPU 15 with public-verifier output checks. No selected sample was discarded.
Three Able cohorts and two reference cohorts give:

| Lane | Cohort means | Pooled mean | Limiting pooled ratio |
| --- | --- | ---: | ---: |
| Able compiled | 0.0540 / 0.0520 / 0.0580 s | 0.05467 s | 3.919x Go |
| Go 1.26 | 0.0138 / 0.0141 s | 0.01395 s | — |
| Able bytecode | 4.804 / 4.584 / 4.736 s | 4.708 s | 25.435x Python / 21.964x Ruby |
| Python 3.14 | 0.1698 / 0.2004 s | 0.1851 s | — |
| Ruby 4.0 | 0.2095 / 0.2192 s | 0.21435 s | — |

The compiled workload is 25.52% of equivalent Go performance, while bytecode
is 3.93% of Python and 4.55% of Ruby performance. Both are clear target
misses; neither miss is caused by an accidental interpreter link.

## Strict compiled lowering

The strict final graph contains 96 dependencies and no
`able/interpreter-go/pkg/interpreter` dependency. Generated carriers are:

```go
type __able_array_Scheduled_SimulationEvent struct {
    Elements []*Scheduled_SimulationEvent
}
type EventQueue_SimulationEvent struct {
    Items *__able_array_Scheduled_SimulationEvent
}
type Scheduled_SimulationEvent struct {
    At int64
    Sequence int64
    Value *SimulationEvent
}
type SimulationEvent struct {
    Kind int32
    Payload int64
}
```

`main` calls `EventQueue_new_spec`, `push_spec`, `pop_spec`, and `len_spec`
directly. Those specialized methods accept and return only the carriers above
plus native Go control results. The separately generated generic entry
wrappers retain `runtime.Value` compatibility for dynamic/irreducible use, but
the static hot path never calls them.

Thirty independently verified main-only CPU profiles merge to 540 ms.
`EventQueue_pop_spec` owns 22.22% flat/64.81% cumulative,
`comes_before_spec` owns 11.11% flat, and Go GC span scanning owns 29.63%
flat. The only compiler helpers sampled are the already-closed checked
arithmetic operations. There is no bridge, interpreter, generic-value
conversion, or interface-dispatch owner.

Three exact main allocation processes each report 2,319,168 bytes, 108,224
allocations, and two collections. Source-line attribution assigns 108,192
objects exactly to the 54,096 `SimulationEvent` and 54,096
`Scheduled_SimulationEvent` constructions also present in the Go
implementation. Only 32 phase allocations remain outside those
source-semantic objects. This is direct evidence against hidden boxing.

## Bytecode ownership

Three warmed one-main profiles merge to 13.65 CPU seconds and report identical
128,289,384 B/op and 3,904,826 allocs/op. Flat ownership is dispersed across
the ordinary VM dispatcher, cached static-member calls, propagation/type
matching, frame/stack operations, and Go maps used by semantic caches.

These are the existing `bytecode-iterator-control` families. The new row adds
no storage-specific leaf and does not invalidate the rejected broad
frame/stack/register, member/call, or semantic-region routes. Its inclusion is
valuable coverage, but it is not authorization to retry those routes.

## Evidence and verification

- repeated selected reports:
  `2026-07-27-discrete-event-simulation-{able,able-repeat}.json` and
  `2026-07-27-discrete-event-simulation-comparison.{json,md}`;
- repeated reference reports:
  `2026-07-27-discrete-event-simulation-{go-reference,go-reference-repeat,interpreter-reference,interpreter-reference-repeat}.{json,md}`;
- readable ownership:
  `2026-07-27-discrete-event-simulation-{compiled,bytecode}-profile-top.txt`
  and
  `2026-07-27-discrete-event-simulation-compiled-allocation-top.txt`;
- exact six-language output parity and five verifier-backed selected samples
  per lane;
- strict native-carrier inspection and an interpreter-free dependency graph;
- catalog, coverage, operation-depth, selection, scoreboard, frontier,
  evidence-ledger, and deterministic architecture checks.

All large build/profile state lived under disk-backed `/var/tmp`.

## Next recommendation

Return to the evidence-gated production pause. Expand or invalidate the corpus
only with another genuine application or semantic change; admit performance
code only when current profiles identify one non-closed exact owner material
in at least three unlike applications.

Why: this tranche supplies the missing user-defined generic-storage control
and proves its compiled representation is already the intended native Go
representation. The remaining compiled costs are the application heap,
source-equivalent object allocation/GC, and already-closed checked operations,
not boxing or an interpreter boundary.

What it entails: keep the new application as a permanent lowering guard,
maintain exact scorecard/frontier identities, and inspect new authoritative
invalidations before profiling or changing production code. If one appears,
refresh three unlike owners, then require focused semantic guards and at least
five balanced verifier-backed baseline/candidate/Go or interpreter cohorts.

Why it is important: this prevents a correct native generic representation
from being destabilized in pursuit of a benchmark-specific speedup and keeps
effort aimed at genuinely general owners capable of moving the product target.
Do not begin WASM work.
