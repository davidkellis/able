# Concurrent Document Pipeline Option reconciliation — 2026-07-21

## Decision

Recognize Concurrent Document Pipeline as portable coverage for
`option_result_exceptions`. Retain no benchmark, compiler, runtime, bytecode
VM, canonical-stdlib, language, or WASM code change, and retain the existing
timing/profile evidence without remeasurement.

This is a metadata correction, not a new claim about executable performance.
The application, verifier, selected rows, generated binaries, and all
reference implementations are unchanged.

## Semantic evidence

The v12 language contract defines `?T` as `nil | T`. `Channel.receive()` has
return type `?T` and returns `nil` when a channel is closed and drained. The
canonical kernel exposes the same signature.

Concurrent Document Pipeline calls `jobs.receive()` inside each worker and
exhaustively matches the nullable result:

- `case task: DocumentTask` processes every received task;
- `case nil` records that one closed, drained worker has finished.

At the fixed benchmark scale, 32 input lines are replayed for 32 rounds, so
the workers execute 1,024 successful nullable matches. Four workers then each
execute the `nil` branch once. This is 1,028 dynamic Option matches in the
ordinary application path, not a declaration-only or incidental use.

## Frontier effect

The reconstructed baseline explicitly removes the newly recognized
membership, preserving the historical comparison. Current coverage changes
as follows:

| Measure | Before | After |
| --- | ---: | ---: |
| improved three-family interactions | 152 | 154 |
| minimum interaction depth | 2 | 2 |
| interactions at minimum depth | 6 | 4 |
| concurrency × expressions/files × Option depth | 2 | 3 |

The target interaction now has three unlike applications: Concurrent Document
Pipeline, Concurrent Event Routing, and Concurrent Text Index. The performance
frontier is unchanged at 89 selected rows, eight snapshot meets, 81 misses,
146.470105 seconds of target excess, and zero actionable groups.

## Evidence integrity

No timing was rerun because no executable input changed. The prior 60 verified
timing processes and profiles remain exact evidence for the same source and
binaries. The performance-evidence selector is empty and all 21 closures are
current. Repeating unchanged programs would add workstation variance without
testing a changed performance claim.

## Next recommendation

Audit the four remaining depth-two interactions together before adding a new
application. They are the closure/callable and interface-dispatch variants of
concurrency with files/text and real entry, currently represented by either
Concurrent Document Pipeline plus Concurrent Event Routing or Concurrent Event
Routing plus Concurrent Text Index.

Why: one existing application may already execute one of those families
substantially, and one honest concurrent ingestion/validation application could
naturally close more than one remaining gap. The audit should inspect exact
dynamic operations in Concurrent Text Index, Validated Job Pipeline, and
Dependency Wave Validation first. Add a source-equivalent Able/Go/Python/Ruby
application only if no existing workload qualifies. Performance-candidate
admission remains independent: require one concrete non-parent leaf reproduced
in three unlike applications and broad non-regressing guards. Do not begin
WASM work.
