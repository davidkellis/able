# Bytecode typed-`i32` slot storage closure

Date: 2026-07-26

Decision: retain no production VM, runtime, compiler, tree-walker, stdlib,
benchmark, fixture, language, dependency, or WASM change.

The requested typed-slot architecture is already present in the current VM.
Fresh source, carrier, CPU, and allocation evidence found no remaining boxed
slot transition that is material in three unlike applications. A new slot
representation would therefore duplicate retained native lanes or reopen
previously rejected operand-stack/frame ABI work.

## Current artifact and retained architecture

Fresh binaries built from repository commit
`237406eccdfb025a519d898daedadee1c8d13a7b` with the disk-backed Go cache are
byte-for-byte identical to the artifacts used by the July 26 three-application
and 54-row bytecode owner censuses:

| Artifact | SHA-256 |
| --- | --- |
| `pkg/interpreter` benchmark binary | `5069b6dff944d7e68aeb38fb9b85dab990b4d29c842a6cfed04fe66897cb01ab` |
| ordinary `cmd/able` CLI | `5f1108bc9596e74dd37e29fdb863bf8fa517e91935fd7db83ceecc940b896666` |

No VM/runtime source changed after those artifacts. Their full-corpus CPU and
allocation closure remains current rather than historical.

The source audit confirms that the general bytecode slot layer already has:

- typechecker-proved `i32` register frames backed by `[]int32` plus validity
  bits;
- value-slot `i32` sidecars backed by `[]int32` plus validity bits, with the
  corresponding `runtime.Value` slot kept nil;
- raw `i32` slot values and allocation-free cached stack carriers;
- direct typed and untyped integer stores, including in-place owned-cell
  updates for other primitive integer widths;
- direct primitive reads for arithmetic, comparison, casts, Array indexes,
  calls, and returns where the static/runtime proof permits them; and
- explicit materialization at environment, generic return, host/native,
  polymorphic, interface, nominal, error, and other dynamic boundaries.

Call-frame save/restore state already preserves register and sidecar ownership.
Unknown or incompatible values invalidate the native fact and follow the
ordinary boxed semantic path.

## Three-unlike carrier census

One ordinary CLI process per application used CPU 6, one logical CPU,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, normal typechecking, the
canonical external stdlib, and main-only opt-in bytecode statistics. Every
process passed its public Ruby verifier.

| Application | Family | Proven integer loads | Native/raw carriers | Ordinary small values |
| --- | --- | ---: | ---: | ---: |
| Array Slice Window | Array/slice | 1,020,268 | 624,260 raw `i32` | 396,008 |
| Word Frequency | text/map/Result | 944,384 | 576,412 raw `i32`; 75 raw `i64` | 367,897 |
| Fixed Width 128 | wide numeric/nominal | 5,000,006 | 2,262,142 raw integer carriers | 2,737,862 small/big values |

Distance Field is the control: it recorded zero proven integer loads because
its corresponding typed lane is `f64`.

The exact carrier/consumer shapes do not intersect across all three:

- Array is led by raw-`i32` comparisons and raw/ordinary integer casts.
- Word is led by raw-`i32` arithmetic/comparison and ordinary integer
  Array-value stores.
- Fixed Width is led by mixed-width struct construction, static calls, and
  signed/unsigned/wide casts; only 262,144 of its loads are raw `i32`.

This also reproduces the July 22 carrier census exactly for all three
applications. Concurrent Event Routing supplied an additional verifier-backed
one-shot observation—633,856 raw-`i32` and 452,352 ordinary loads—but its
shared Array-value route was already proved to use the direct monomorphic
`u8` kernel path and to be CPU-material only in Reverse Complement.

## Fresh ordinary CPU and allocation evidence

Diagnostics were disabled for three independent measured-main CPU and exact
allocation-counter processes per selected application. The same frozen
benchmark binary loaded and typechecked once, warmed `main` once, forced GC,
then measured one main call.

| Application | Mean main time | Mean allocated bytes | Mean allocations |
| --- | ---: | ---: | ---: |
| Array Slice Window | 556.946 ms | 14,190,688 | 422,243.333 |
| Word Frequency | 1,248.759 ms | 48,545,547 | 637,236.667 |
| Fixed Width 128 | 8,297.743 ms | 1,242,275,800 | 30,858,402.667 |

The merged profiles contain one apparent common CPU parent:
`bytecodeRawIntegerValueInfo` is 2.41% flat in Array, 1.08% in Word, and
2.62% in Fixed Width. Its concrete callers are disjoint:

- Array: casts, typed stores, same-type comparisons, and integer extraction;
- Word: same-type comparisons and typed-pattern matching; and
- Fixed Width: wide-integer extraction, direct integer conversion, casts, and
  simple-check coercion.

No caller is CPU-material in all three. The slot-to-stack carrier
`appendSlotStackValueChecked` is 2.96% flat in Word but only 0.60% in Array
and 0.57% in Fixed Width.

Sampled allocation ownership is likewise split:

- `bytecodeBoxedIntegerValue` is 0.87% of sampled objects in Array, 2.49% in
  Word, and 7.35% in Fixed Width;
- `bytecodeRawIntegerResultValue` is 41.77% in Array and 3.80% in Fixed Width,
  but is not material in Word; and
- the dominant remaining objects are Array storage/leases in Array,
  String/nominal/Array values in Word, and big-integer/nominal work in Fixed
  Width.

The sampled profiles include process-initialization cache samples; the exact
measured-main counters above do not use those samples as allocation totals.

## Admission decision

| Proposed parent | Material applications | Decision |
| --- | ---: | --- |
| New typed `i32` slot storage | already implemented | close as completed infrastructure |
| Generic raw-integer inspection | 3 | reject parent: no common concrete caller |
| Slot-to-stack transport | 1 | reject: below three |
| Boxed integer allocation | 2 | reject: below three |
| Raw integer result allocation | 2 | reject: below three and mixed suffix semantics |
| Broad tagged slot/frame ABI replacement | not admitted | previously rejected broad ABI/operand-stack route |

No production prototype cleared the three-unlike-program gate, so a
baseline/candidate/Python/Ruby cohort was not warranted. The current
five-process public scorecard remains the baseline:

| Application | Able | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: |
| Array Slice Window | 0.744s | 12.197x | 5.750x |
| Word Frequency | 1.414s | 60.427x | 27.350x |
| Fixed Width 128 | 8.522s | 24.321x | 12.603x |

All rows remain far outside the `1.052632x` target, but the gap cannot be
attributed to one missing typed-`i32` slot carrier.

## Verification and cleanup

- Three-run focused native-lane, invalidation, overflow, materialization,
  call-boundary, carrier-observer, and frame-proof tests pass in 0.070s.
- The three selected carrier processes and the Distance Field control pass
  their public verifiers.
- `git diff --check` passes for the retained handoff.
- The complete `./run_all_tests.sh` handoff passes every contract,
  non-compiler package, all 32 compiler batches, and the 107.961-second final
  bytecode fixture corpus.
- All generated profiles, binaries, statistics, and output captures under
  `/var/tmp/able-bytecode-i32-slot-20260726` were removed after their hashes
  and aggregate evidence were recorded. The reusable disk-backed Go cache was
  retained and the task TMPDIR was emptied.

## Next recommendation

Run a report-only bytecode primitive-materialization boundary census over the
54 rankable applications.

Why: typed primitive register, sidecar, raw-slot, stack, and direct-store lanes
already exist, while symbol-level profiling groups semantically different
materializations under broad helpers. The remaining question is whether an
avoidable semantic boundary is hidden across different callers.

What it entails: add opt-in counters at raw-to-boxed transitions, recording
the primitive carrier/suffix, originating opcode/source site, and reason such
as static call, return, cast, pattern, environment, collection value,
interface/union, host/native, error/control, or public escape. Classify
required dynamic boundaries separately from candidate static boundaries,
reconcile counts with CPU/allocation profiles across the full corpus, and
admit production work only when one avoidable reason is material in three
unlike applications.

Why it is important: this tests the actual architectural objective—keep
primitive values native until semantics require boxing—without replacing the
whole VM slot ABI or optimizing one application. Do not begin WASM work.
