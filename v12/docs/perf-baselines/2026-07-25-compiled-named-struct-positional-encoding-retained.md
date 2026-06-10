# Compiled named-struct positional encoding retained

Date: 2026-07-25

## Decision

Retain the general generated named-struct runtime encoding change.

When a generated native nominal struct must cross an explicit runtime boundary,
its `runtime.StructInstanceValue` now stores fields in definition-ordered
positional slots instead of allocating a `map[string]runtime.Value`. The
existing runtime helper keeps payloads of up to three fields in the struct
instance's inline storage. Generated decoding, named patterns, assignment
patterns, and IR patterns retrieve named fields through the shared by-name
helper, which accepts both positional and legacy map-backed representations.

The rule does not name `Error`, `Result`, any application type, or a container.
It applies to every generated named nominal struct through the shared semantic
encoding. Map-backed values originating at compatible dynamic or host
boundaries remain readable.

Twenty final order-balanced baseline/candidate/Go cohorts improved all three
unlike applications: Concurrent Event Routing by 22.61%, Manifest
Normalization by 15.75%, and Binary Event Log by 20.22%, for a 19.58%
geometric-mean improvement. Normal-workload allocation bytes fell
32.91%-49.95%; intentionally error-heavy bytes fell 32.70%-48.91%.

Machine-readable evidence is in
`2026-07-25-compiled-named-struct-positional-encoding-retained.json`.

## Admission and rejected route

Fresh normal and intentionally error-heavy profiles reproduced concrete error
encoding in all three strict interpreter-free applications. The old
map-backed conversion allocated exactly three flat objects per converted error
in each workload.

A diagnostic prototype that mapped `Error` through the ordinary generated
native-interface carrier was rejected and fully removed. It invalidated
general `Result` carrier completeness, nullable `cause` signatures, matcher
adapters, and fallback-free compilation. Making that route correct would
require the broad control/Result ABI change excluded from this tranche.

The admitted alternative is a shared nominal encoding rule already supported
by both interpreters: definition-ordered positional storage for named structs.
It reduces the unavoidable runtime representation without introducing an
`Error`-specific or application-specific carrier.

## Semantic encoding

Generated named-struct conversion now:

1. resolves the same runtime struct definition as before;
2. allocates one `StructInstanceValue` through
   `runtime.NewStructInstancePositionalSized`;
3. registers it in the existing cycle-preservation map before converting
   fields;
4. writes converted fields in definition order; and
5. returns the same runtime nominal identity.

The tree-walker and bytecode VM already create ordinary named structs with
this positional representation. Generated struct decoding and the common
field helper already accepted it. The candidate completed the symmetry for
generated-to-runtime conversion.

The semantic guard exposed one additional compiler obligation: runtime named
patterns had treated every non-nil positional slice as a positional struct and
indexed by pattern order. Generated match, assignment-pattern, and IR pattern
paths now distinguish named from positional patterns. Named patterns resolve
each requested field through `__able_struct_named_field_value`; positional
patterns remain index-based.

## Repeated wall time

After two verified warmups per lane, twenty order-balanced
baseline/candidate/Go cohorts ran on CPU 9 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, and `GOGC=50`. Every process passed the application's
public verifier.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 25.700 ms | 19.888 ms | -22.61% | 3.173 ms | 6.267x | 15.96% |
| Manifest Normalization | 15.569 ms | 13.117 ms | -15.75% | 2.787 ms | 4.706x | 21.25% |
| Binary Event Log | 170.152 ms | 135.743 ms | -20.22% | 6.297 ms | 21.557x | 4.64% |

The geometric-mean candidate/Go ratio is 8.599x, or 11.63% of Go
performance. This tranche materially reduces a shared boundary but does not
meet the 1.052632x compiled target.

## Exact allocation

Three verified exact main-phase allocation profiles were collected for every
application in both normal and error-heavy modes.

| Application / mode | Bytes baseline -> candidate | Change | Allocations baseline -> candidate | Change | GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| Event Routing / normal | 8,959,858.67 -> 4,484,386.67 | -49.95% | 104,584.00 -> 77,944.33 | -25.47% | 5.00 -> 4.00 |
| Event Routing / error-heavy | 11,261,325.33 -> 5,753,624.00 | -48.91% | 119,160.00 -> 86,376.33 | -27.51% | 7.00 -> 5.00 |
| Manifest / normal | 5,824,397.33 -> 3,907,458.67 | -32.91% | 93,125.67 -> 84,933.67 | -8.80% | 5.00 -> 3.00 |
| Manifest / error-heavy | 4,208,890.67 -> 2,832,554.67 | -32.70% | 41,156.00 -> 32,963.33 | -19.91% | 3.00 -> 2.00 |
| Binary / normal | 94,117,293.33 -> 62,725,448.00 | -33.35% | 1,209,078.00 -> 1,078,006.00 | -10.84% | 89.67 -> 60.00 |
| Binary / error-heavy | 55,070,773.33 -> 33,050,490.67 | -39.99% | 463,033.00 -> 331,961.00 | -28.31% | 52.00 -> 31.00 |

The change also benefits non-error named-struct crossings in these
applications, which is why whole-application savings exceed the selected
error-only owner.

## CPU profiles

Ten verified final main-phase CPU profiles were merged per application and
mode. The short Event and Manifest processes make 10 ms samples coarse, so
these totals are supporting evidence:

| Application | Normal baseline -> candidate | Error-heavy baseline -> candidate |
| --- | ---: | ---: |
| Concurrent Event Routing | 170 ms -> 150 ms | 150 ms -> 120 ms |
| Manifest Normalization | 90 ms -> 80 ms | 40 ms -> 50 ms |
| Binary Event Log | 1.61 s -> 1.29 s | 500 ms -> 400 ms |

Binary's longer profiles show the error converter falling from 220 ms to
80 ms cumulative in error-heavy mode. Manifest's one-sample fluctuation
reinforces why retention is based on exact allocations and repeated wall time,
not short-profile percentages.

## Verification

- Normal and error-heavy outputs match the equivalent Go applications for all
  three final binaries.
- All final strict dependency graphs omit
  `able/interpreter-go/pkg/interpreter`.
- Generated-source guards require arbitrary named structs and a canonical
  error fixture to use positional runtime storage while retaining map-backed
  decode support.
- An executable guard covers two error implementations, nested causes,
  raise/rescue/rethrow, interface method dispatch, and named-field dynamic
  inspection.
- Existing imported-interface, equality, dynamic error, and map-fallback
  guards pass in the broader focused compiler group.
- All 180 final timed processes, 18 warmups, 36 exact-allocation processes,
  and 120 CPU-profile processes passed verification.
- The broader error/interface/pattern group passes in 8.853 seconds.
- The focused new semantic/source group passes in 2.578 seconds.
- `go test ./cmd/ablec -count=1 -timeout=60s` passes in 6.111 seconds.
- All touched source and test files remain below 1,000 lines.
- No canonical stdlib, runtime package, interpreter, bytecode VM, language,
  dependency, or WASM change was needed.
- Final binary SHA-256:
  - Event Routing:
    `ba9f56fed856bcea681fe77364dee23aa435fec2c746e7f17a83136c0921f574`
  - Manifest Normalization:
    `fcf9852ad2841401e58a3a6e71a1c2ecfc50ebc5539784e8d0929c641f9777a7`
  - Binary Event Log:
    `efa50d8502ff542cc6c208120faf54c8a1f301e42c64d394d83c394f2c126a7c`
- Raw artifacts:
  `/tmp/able-static-interface-errors-20260725.fkwojK`

## Next

Qualify caller-owned recovery of generated native structs from runtime
`StructInstanceValue` across Concurrent Event Routing, Manifest Normalization,
and Binary Event Log.

This is next because the final exact profiles now expose the inverse side of
the same general boundary in all three applications:
`EventTask`/`RoutedEvent` recovery in Event Routing, `ManifestRecord` recovery
in Manifest Normalization, and `EventRecord` recovery in Binary Event Log.
Each `__able_struct_*_from` helper eagerly allocates a generated native pointer
before an immediate static consumer.

The work entails proving where a caller-owned stack result or shared
`from_into` helper preserves nominal identity, aliases, cycles, nullable and
generic fields, failure atomicity, dynamic values, and imported definitions.
It must profile exact callsites first and admit at most one general
nominal-recovery rule; it must not name these records, a container, or another
non-primitive nominal type.

This is important because it is the next representation allocation repeated
in the same three unlike strict programs. Removing it where lifetime is
statically bounded would keep recovered Able records in ordinary Go storage
for their immediate compiled consumers and further reduce crossings toward
native Go performance.

Do not begin WASM work.
