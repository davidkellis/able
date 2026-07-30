# Generic Slot Buffer application gate

## Decision

Retain Generic Slot Buffer as the 65th source-equivalent portable application,
along with its references, catalog/coverage entries, repeated measurements,
and closure-selected profiles. Retain no compiler, generated-runtime,
interpreter, VM, stdlib, language, dependency, or WASM change.

The application answers the tranche's question: user-defined generic nominal
storage already lowers to direct native Go structures. It does not reproduce
the generic-map materialization path from Transaction Ledger Audit and
Inventory Reconciliation, so there is no shared nominal-storage owner and no
authorization for typed nominal architecture work.

## Application contract

`VersionedBuffer<T>` owns an `Array<Slot<T>>`. The workload performs direct
writes, observes an alias mutation through the original buffer, reads through
a generic `ReadableBuffer<i64>` interface, and traverses the buffer through
its `Iterable`/`Iterator` implementation. It therefore covers identity,
aliasing, interface dispatch, iteration, static Array storage, and nullable
primitive reads without relying on a named stdlib container.

The canonical and external Able sources are byte-identical at
`6309f10f16478b34efaa12fb292b24cc22db7bbc700e6f77e4403b32dd37697b`.
Able, Go, Python, and Ruby all produce:

```text
512:512:278528:232173748:800413927:576434349:166504429:71319552
```

The output SHA-256 is
`149cd95dcb57f9309c82ccd148336280f98baa95ea3d91ba34be7989fdab06fe`.
Tree-walker, bytecode, strict compiled, Go, Python, and Ruby correctness checks
passed. The strict build used `--no-fallbacks`; its final dependency graph
omits `pkg/interpreter`.

## Repeated comparison

All scorecard measurements used CPU 15, Go 1.26.5 where applicable, five
successful Able processes, five successful reference processes per
comparison, and the public verifier. No sample was discarded.

| Mode | Able mean | Reference mean | Able/reference |
| --- | ---: | ---: | ---: |
| compiled | 0.0420 s | Go 0.0051 s | 8.2353× |
| bytecode | 2.2580 s | Python 0.1847 s | 12.2252× |
| bytecode | 2.2580 s | Ruby 0.1037 s | 21.7743× |

Both rows are target misses. They expand the evidence frontier but do not
admit a production candidate by themselves.

## Compiled lowering result

Thirty independent strict CPU-profiled processes contributed 710 ms of merged
samples. The exact main phase allocated 2,124,160 bytes in 264,218
allocations and performed one GC. Generated Go uses:

- `__able_array_Slot_i64` with `[]*Slot_i64` storage;
- `Slot_i64` with direct `int64` value and generation fields;
- `VersionedBuffer_i64` with the native Array carrier and an `int64` mutation
  field;
- a direct typed `ReadableBuffer_i64` adapter whose methods call compiled
  implementations.

There is no `runtime.Value` conversion or interpreter transition in the
application hot path. Identity and alias-observed mutation remain direct Go
pointer identity.

The exact allocation profile instead attributes 131,329 flat objects, or
81.87% of the captured objects, to
`VersionedBuffer_get_spec`. A successful `?i64` read constructs
`__able_ptr(slot.Value)` because the generated nullable scalar carrier is
`*int64`.

This is not the hypothesized nominal-storage materialization. Inventory
Reconciliation and Transaction Ledger Audit allocate through
`bridge.ToDynamicI64` and nullable recovery at their runtime map boundary,
whereas Generic Slot Buffer allocates its directly typed nullable result
inside `get`. The common fact is the pointer-backed primitive-nullable
representation, not one exact allocating leaf across the three programs.
A global nullable ABI rewrite is too broad to infer from this tranche and
therefore remains an evidence trigger rather than retained code.

Readable carrier, CPU, allocation, and main-phase summaries are in
`2026-07-29-generic-slot-buffer-owner-profiles/`.

## Bytecode closure

Three warmed serial runtime profiles contributed 7.11 seconds of CPU samples.
Their operation times were 2.1827, 2.2567, and 2.7345 seconds; allocation
counts were stable at approximately 39.64 MB and 2.172 million allocations
per operation.

The leading descendants are `runResumable`, cached member lookup,
raw-integer inspection, inline resolved calls, member calls, frames, stacks,
binary operations, and casts. These are already-closed dispatcher,
call/member, frame/stack, raw-integer, propagation/type, and cache families.
No storage-specific VM leaf appears, so no bytecode route is reopened.

## Cohort result

The portable corpus now contains 65 applications and 130 selected
compiled/bytecode rows. The selection identity is
`d450605f6b271fbddbda7bf31e9f61c1d87cbf1a407f9047304d79bd64ff1684`.
Every new Able and reference row has five verified samples.

## Verification

Catalog syntax and mapping, feature coverage, the 130-row selection manifest,
five-sample scorecard evidence, the performance frontier, and the 23-entry
closure ledger all pass their focused checks. `go test ./cmd/ablec` passes.

The complete `./run_all_tests.sh` handoff suite passes every preflight
contract, non-compiler package, all 34 bounded compiler batches, and the
bytecode fixture pass. The only issue found during the first invocation was a
stale test expectation for 64 applications; advancing that guard to the
correct 65-application corpus made the focused test and full rerun green.
Removed 190 MiB of exact `/var/tmp/able-generic-slot-buffer-*` build/profile
artifacts and generated Python caches after retaining the readable evidence.
No matching Generic Slot Buffer artifact remains in `/var/tmp` or `/tmp`.

## Next recommendation

Run a focused compiled nullable-scalar carrier reconciliation before changing
the generated ABI.

Why: pointer-backed primitive nullables are now dynamically material across
Generic Slot Buffer, Inventory Reconciliation, and Transaction Ledger Audit,
but the exact allocating leaves and required runtime conversions differ. The
evidence is strong enough to investigate the representation, not yet strong
enough to rewrite it.

What it entails: census every primitive nullable shape and its static,
interface, union, dynamic, and host boundaries; specify nil, equality, and
conversion behavior; then prototype a general value-plus-valid typed carrier
only if that census preserves the evidence gate. Any candidate needs focused
semantic guards plus at least five balanced baseline/candidate/Go pairs across
the three unlike applications.

Why it is important: a sound general nullable carrier could remove hundreds
of thousands of heap boxes without disturbing the already-native nominal
storage, adding a named-container special case, or crossing the
compiled/interpreted boundary.
