# Bytecode static primitive-return consumer closure

Date: 2026-07-26

## Decision

Retain caller/return-frame attribution in the opt-in primitive-materialization
observer and add a dedicated low-overhead observer mode. Retain no production
return-carrier change.

All 54 rankable applications passed their public verifiers. The census
reconciled all 1,678,255 candidate-static return materializations across 39
applications to their first semantic consumer, with no dropped sites.

## Census

The temporary stack-lifetime probe was used only for this report and removed
before profiling. It distinguished the next instruction from the first opcode
that actually reads the returned stack cell.

| First consumer | Count | Applications |
|---|---:|---:|
| `StoreSlot` | 828,865 | 3 |
| `StoreSlotNew` | 262,858 | 17 |
| `BinaryIntAdd` | 140,290 | 17 |
| chained `Return` | 95,325 | 9 |
| `StructLiteralNamedFast` | 68,358 | 8 |
| comparison jump | 61,440 | 1 |
| `MemberSet` | 52,370 | 2 |
| generic `Binary` | 45,144 | 4 |
| calls | 79,845 | 8 |
| other instructions | 21,007 | 9 |
| superseded type precheck | 12,001 | 1 |
| public escape | 96 | 1 |

The 800,000 `ReturnBinary` `i32`-to-`StoreSlot` transitions are Fasta-only.
The broadest ordinary owned-slot leaf is `Return` `integer_result<i64>` to
`StoreSlotNew`: 210,974 transitions in Concurrent Audio Voices (129,050),
Concurrent Packet Codecs (57,344), Concurrent Scene Tiles (20,484), and Mutex
Work Queue (4,096).

Array Slice Window exposed 12,001 redundant materializing prechecks followed by
12,001 coercion-path materializations. That is a real duplicate transition but
only one application, so it does not clear the breadth gate.

## Diagnostics-off profile gate

One measured-main CPU and allocation profile per high-volume application
showed repeated boxing allocation:

| Application | Main bytes | Allocations | Sampled integer boxing |
|---|---:|---:|---:|
| Concurrent Audio Voices | 181,197,472 | 4,103,106 | 11 MB |
| Concurrent Packet Codecs | 85,525,488 | 2,023,217 | 9 MB |
| Concurrent Scene Tiles | 67,436,480 | 1,105,870 | 3 MB |

This admitted one local experiment: immutable raw integer results could cross
an inline return without materialization only when the saved caller
immediately stored them into an untyped VM-owned slot. The existing store path
already accepts raw integers; public, typed, aggregate, dynamic,
interface/union, error/control, and callback boundaries were unchanged.

## Rejected experiment

Three fresh interleaved exact one-main samples per variant rejected the rule:

| Application | Baseline | Candidate | Time change | Bytes change | Allocation-count change |
|---|---:|---:|---:|---:|---:|
| Concurrent Audio Voices | 1.0572s | 1.0744s | +1.62% | -1.72% | +2.29% |
| Concurrent Packet Codecs | 0.5225s | 0.5532s | +5.87% | -2.37% | +1.82% |
| Concurrent Scene Tiles | 0.3794s | 0.3978s | +4.84% | +0.07% | +0.56% |

The immutable raw value-form carrier reduced byte volume in two applications
but escaped through interface-backed slot storage often enough to create more
objects in all three. Packet and Scene also regressed materially. The
candidate therefore failed before the five-run public-language A/B gate and
was fully reverted.

The restored interpreter test artifact
`7f56e96e8719888b3922fe4a66337fa737195ded1bade7613f6f8400914a0dbf`
matches the pre-candidate artifact byte for byte. The rejected candidate was
`1ff1c734778c88f3923fa6b5fe5e0f6158fe842d7e526ee723e6bdb56f1e23c4`.

## Retained diagnostics

`ABLE_BYTECODE_PRIMITIVE_MATERIALIZATION_STATS=1` now enables only the
primitive-materialization observer. It avoids the cost of unrelated global VM
counters. Static-return rows additionally report the saved frame kind, caller
program, return IP, next opcode, and caller source location. Normal execution
remains diagnostics-off.

No compiler, runtime semantics, tree-walker, stdlib, benchmark, fixture,
language, dependency, or WASM change was required. The machine-readable
companion is
`2026-07-26-bytecode-static-return-consumer-closure.json`.

The complete `./run_all_tests.sh` handoff passed every contract, all 32
compiler batches, and the 88.052-second final bytecode fixture corpus.
Compiler batches 19, 28, and 29 passed but took 185.073, 75.304, and 93.208
seconds; these unchanged-code duration anomalies should be audited before the
next heavy profiling run.

## Recommendation

Next briefly repeat and identify the three over-one-minute compiler batches,
then return to compiled AOT lowering and refresh diagnostics-off generated-code
CPU and allocation profiles for at least three unlike current strict target
misses. Intersect residual `runtime.Value`, interface conversion, and boxing
owners across their generated Go, then admit only a general primitive/static
carrier rule that repeats.

This is next because the broad bytecode candidate-static materialization
reasons are now closed at their concrete consumers, while compiled parity with
Go remains the first roadmap priority. It entails tracing generated static
calls and stores from Able types to Go carriers without reopening fallback or
interpreter paths. It is important because interpreter-free compiled programs
should be able to approach native Go when primitive values remain native
through the entire generated call graph. Do not begin WASM work.
