# Bytecode cross-feature exact-leaf matrix

Date: 2026-07-22

## Decision

Retain no VM, runtime, compiler, canonical-stdlib, benchmark, fixture,
language, or WASM change from this tranche. Five fresh bounded CPU profiles
covering unlike feature families expose no new exact interpreter-owned flat
leaf that is material in at least three applications.

The only exact interpreter symbols with broad visible flat samples are
`(*bytecodeVM).runResumable`, an aggregate dispatch loop, and
`bytecodeRawIntegerValueInfo`. The raw extractor clears 1% flat CPU only in
Binary Event Log and iterator collect in this cohort. Its direct parents split
among comparisons, casts, coercion checks, and generic integer conversion; no
direct parent clears 1% in three applications. The repeated Go map/hash leaf
resolves primarily to the already-closed primitive integer metadata table plus
environment/type/cache maps with different semantics.

## Workload and profiling contract

The representatives deliberately span five different feature families:

| Family | Application | Reason selected |
| --- | --- | --- |
| Text/file | Binary Event Log | Binary file input, parsing, strings, records, Result and HashMap updates |
| Numeric | Matrix Multiply | Allocation-heavy nested Arrays plus the general monomorphic-f64 dot loop |
| Collection/iterator | Linked-list iterator collect | Generic lazy map/filter/collect/reduce and member dispatch |
| Nominal/union | Option/Result Config | Generic unions, Error, captured callables and nominal dispatch |
| Concurrency | Validated Job Pipeline | Files, channels, futures, workers, Result and nominal messages |

The fixed interpreter test binary SHA-256 was
`883a6ff6eaa828076a32dd5738f8844e2bec1803778bf2be86fb7e19ec650a00`.
Every process used normal typechecking, canonical external `able-stdlib`,
source-root-only loading, one warmup, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and a 59-second cap.

Initial one-call calibrations all passed. Iterator and Validated Job Pipeline
were too short for reliable 10 ms CPU samples, so the final profiles increased
measured calls within one warmed process. This is attribution, not an A/B wall-
time claim; repeated calls improve sampling resolution without treating the
profiled timing as a promoted scorecard result.

| Application | Measured calls | ns/op | Bytes/op | Objects/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 1 | 6,424,270,525 | 465,602,864 | 8,361,497 | 6.40 s |
| Matrix Multiply | 2 | 4,667,423,656 | 306,823,384 | 14,034,078 | 9.31 s |
| Iterator collect | 20 | 404,939,483 | 8,366,802 | 192,563 | 8.06 s |
| Option/Result Config | 8 | 800,227,069 | 66,994,835 | 1,167,548 | 6.39 s |
| Validated Job Pipeline | 100 | 139,396,074 | 13,578,029 | 201,104 | 13.88 s |

All five final profile processes passed.

## Exact flat-leaf matrix

Percentages are flat CPU shares in Binary, Matrix, Iterator, Option, and
Validated order. The materiality rule is at least 1% flat in at least three
unlike applications. Values below 0.5% are omitted from the compact matrix.

| Exact symbol | Binary | Matrix | Iterator | Option | Validated | Disposition |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| `(*bytecodeVM).runResumable` | 4.69% | 6.34% | 6.45% | 3.76% | 2.45% | Aggregate opcode dispatcher, not a semantic candidate |
| `bytecodeRawIntegerValueInfo` | 3.12% | 0.97% | 3.85% | 0.94% | 0.86% | Only two rows clear 1%; carrier family already broadly gated |
| `(*bytecodeVM).execCallOpcode` | 0.62% | 0.75% | 1.61% | 0.63% | 0.65% | Aggregate call dispatcher; only Iterator clears 1% |
| `finishInlineReturn` | 0.78% | <0.5% | 1.36% | 0.63% | 0.58% | Closed by the preceding shape census and return gates |
| `appendSlotStackValueChecked` | 0.62% | 0.86% | 0.74% | <0.5% | <0.5% | Closed stack-carrier family; no material breadth |
| `tryInlineResolvedCallFromStack` | <0.5% | <0.5% | 1.24% | 1.56% | 0.58% | Two material rows; no three-family admission |
| `Environment.lookupCurrentValueNoLock` | 0.62% | 0% | 1.12% | 1.10% | <0.5% | Two material flat rows; cumulative callers diverge |

There is no exact non-dispatch interpreter leaf at or above 1% flat in three
of the five applications.

## Raw-integer caller reconciliation

The extractor itself remains broadly visible, but its direct callers explain
why a new global carrier change is not admitted:

| Application | Extractor | Leading direct callers as shares of extractor samples |
| --- | ---: | --- |
| Binary Event Log | 200 ms / 3.12% | immediate comparison 25%; same-type pair 20%; immediate arithmetic 20%; cast/direct value 10% each |
| Matrix Multiply | 90 ms / 0.97% | fast cast 44%; u8 conversion 33%; generic integer conversion 22% |
| Iterator collect | 310 ms / 3.85% | cast 35%; immediate comparison 19%; same-type pair 16%; native raw conversion 13% |
| Option/Result Config | 60 ms / 0.94% | direct integer value 67%; immediate arithmetic and coercion check 17% each |
| Validated Job Pipeline | 120 ms / 0.86% | same-type pair 50%; direct integer value 25%; three smaller 8% paths |

No one direct parent contributes 1% of whole-profile CPU in three
applications. The existing direct same-type i32/i64 pair path, exhaustive
single-pass raw slot store, and non-integer comparison rejection already own
the previously proven common cases. The remaining generic-interface,
pre-switch, carrier, store, and producer alternatives have averaged broad
rejections. The current samples add no new carrier or coherence rule that
would invalidate those results.

## Map/hash and allocation reconciliation

Go's `internal/runtime/maps.ctrlGroup.matchH2` is 2.58%-4.38% flat in all five
profiles. Its callers do not represent one Able map:

- `lookupIntegerInfo` is 3.28%, 3.33%, 1.61%, 1.10%, and 0.58% cumulative.
  It is the only fixed semantic table with broad reach, but both the full
  metadata switch and the narrower membership switch already regressed unlike
  application guards and remain closed.
- Environment lookup is cumulative in Binary, Iterator, Option, and Validated,
  but divides among identifier lookup, call-cache ownership, member lexical
  state, ordinary `Lookup`/`Has`, and concurrency-safe locked access. Matrix
  has no corresponding environment leaf.
- Integer-keyed Array storage/handle maps dominate Matrix, while nominal type,
  interface, struct-definition, known-type, and canonical-name caches appear
  in different subsets of the other applications.

The broad Go GC leaves are also not one implementation candidate. For example,
`runtime.tryDeferToSpanScan` is 6.56% Binary, 1.40% Matrix, 0.99% Iterator,
7.04% Option, and 5.84% Validated, while `runtime.mallocgc` is 8.81%-17.94%
cumulative in four applications but only 2.11% in Iterator. CPU profiles do
not identify whether those allocations share a removable Able constructor,
so no allocation candidate is inferred from GC ancestry alone.

## Verification

No temporary production instrumentation or candidate code was added. The
unchanged bytecode family passes:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 27.676s
```

Raw calibration/final profiles and pprof text are cleanup-eligible under
`/tmp/able-bytecode-cross-feature-20260722`.

## Next recommendation

Run a sampled allocation-owner matrix for these same five applications before
leaving the bytecode frontier.

Why: exact VM CPU leaves are now exhausted, but allocation and GC consume a
large share in four unlike applications. The CPU stacks cannot tell whether
that pressure comes from one reusable runtime constructor or merely different
application data. An allocation matrix is the remaining evidence needed to
distinguish those cases without guessing from Go GC internals.

What it entails: capture separate clean `alloc_objects` and `alloc_space`
profiles plus exact measured-main allocation counters for each representative;
intersect fully qualified allocation sites and their interpreter callers; and
admit a candidate only if the same removable semantic allocation is material
in at least three unlike applications. Treat required user result objects,
Array backing growth, and process/bootstrap allocation as non-candidates.
If one owner qualifies, implement a general lifetime/ownership rule and gate
it with repeated alternating processes whose complete averages retain every
workstation sample. Continue to defer WASM.
