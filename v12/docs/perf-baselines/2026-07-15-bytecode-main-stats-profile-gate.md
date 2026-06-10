# Bytecode main-stats profile gate — 2026-07-15

## Question

The main-only counter inventory records high `LoadSlot` counts in unlike Word
Frequency, Future Pipeline, and Base64 runs. Counts are not CPU attribution, so
this gate asked whether one concrete load/materialization descendant is also
material in all three applications.

## Method

Each target ran once through the warmed `bytecode-runtime` harness with one Go
process, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, its catalog input and
executor/source-root contract, and an opt-in CPU profile. The Word Frequency,
Future Pipeline, and Base64 main outputs had already passed their Ruby
verifiers in the preceding counter launch. These are unpinned attribution
profiles, not wall-clock evidence or scorecard input.

## Attribution

| Application | Material samples | `LoadSlot` result | Decision |
| --- | --- | --- | --- |
| Word Frequency | `hashMapFindEntryWithHash` 11.9% flat; call/return and map helpers | no material `execLoadSlotOpcode` sample | Map/call work, not a generic slot leaf. |
| Future Pipeline | `bytecodeRawIntegerValueInfo` 9.4%; scheduler atomics 9.4%; binary/call work | `execLoadSlotOpcode` 3.1% flat | A local integer/scheduler route; raw-integer changes already lost broad guards. |
| Base64 | host `base64.Encode` 33.8%; `Decode` 16.0%; `md5.block` 14.2% | no material generic load sample; Array member call is 9.3% cumulative | Host byte-codec work, not VM load/materialization. |

`bytecodeRawIntegerValueInfo` appears at only 2.8% in Word Frequency and 1.3%
in Base64, so it is not a three-way material leaf either. The existing
execution-lane, three-shape, and candidate-admission decisions remain valid.

## Decision

Reject `LoadSlot`, raw-integer extraction, scheduler, generic call/return, and
Array-member work as a new shared candidate. Do not add a representation,
opcode, cache, or host-codec optimization from this gate. Delete the generated
profiles after recording this result; future performance selection still
requires a new concrete non-nominal leaf across three unlike verifier-backed
applications.
