# Compiled captured-callable cross-family closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime code for the captured-callable candidate.

Fresh strict generated source, exact allocation profiles, and merged CPU
profiles do not reproduce the same erased captured-callable boundary in all
three required applications:

- Manifest Normalization converts a native `ManifestRecord` to
  `runtime.Value` before invoking its erased captured `normalizer`.
- Binary Event Log converts a native `EventRecord` to `runtime.Value` before
  invoking its erased captured `scorer`.
- Concurrent Event Routing already carries its `scorer` as the concrete
  generated Go callable
  `__able_fn__EventRecord_int64_to__AcceptedRoute`. Its nested callback invokes
  that carrier directly and has no `EventRecord` runtime conversion callsite.

The two-family source resemblance therefore does not clear the required
three-unlike-application evidence gate. No A/B candidate was built, and the
required twenty baseline/candidate/Go cohorts were intentionally not run.

## Admission refresh

All three applications were freshly compiled with the current compiler using
`--no-fallbacks`. Their final Go dependency graphs omit
`able/interpreter-go/pkg/interpreter`, and all public verifiers pass.

| Application | Outer captured carrier | Nested argument carrier | Native-record conversion callsites |
| --- | --- | --- | ---: |
| Concurrent Event Routing | `__able_fn__EventRecord_int64_to__AcceptedRoute` | `*EventRecord`, `int64` | 0 |
| Manifest Normalization | `__able_fn_runtime.Value_runtime.Value_to__NormalizedManifest` | native `*ManifestRecord` converted to `runtime.Value` | 1 |
| Binary Event Log | `__able_fn_runtime.Value_runtime.Value_to_int64` | native `*EventRecord` converted to `runtime.Value` | 1 |

Event Routing's current shape is the decisive rejection. Its local `scorer`
is passed directly to the statically typed `route_task` parameter, so the
retained forward fresh-lambda binding inference proves and preserves its
concrete signature. Manifest and Binary use their local callables only from
inside a nested lambda, which the deliberately conservative forward-use
analysis treats as an escape and leaves erased.

This difference is semantic, not noise: the generated Event source contains
neither an `EventRecord` conversion at the nested invocation nor an erased
callable carrier for `scorer`.

## Exact allocation refresh

Three verified exact main-phase allocation runs were collected per
application with:

```text
GOMAXPROCS=1
GOMEMLIMIT=1GiB
GOGC=50
ABLE_GO_PHASE_ALLOC_PROFILE_DIR=<run-directory>
```

| Application | Allocated-byte samples | Mean bytes | Allocation samples | Mean allocations | Mean GC |
| --- | --- | ---: | --- | ---: | ---: |
| Concurrent Event Routing | 8,959,944; 8,959,784; 8,959,960 | 8,959,896 | 104,585; 104,584; 104,585 | 104,584.67 | 5.67 |
| Manifest Normalization | 5,824,360; 5,824,296; 5,824,360 | 5,824,338.67 | 93,126; 93,125; 93,126 | 93,125.67 | 5.00 |
| Binary Event Log | 94,117,224; 94,117,192; 94,117,208 | 94,117,208 | 1,209,078; 1,209,078; 1,209,078 | 1,209,078 | 89.33 |

Exact start/end `alloc_objects` profile subtraction attributes the candidate
conversion as follows:

| Application | Conversion owner | Flat objects | Cumulative objects | Result |
| --- | --- | ---: | ---: | --- |
| Concurrent Event Routing | `__able_struct_EventRecord_to_seen` at nested `scorer` invocation | 0 | 0 | absent |
| Manifest Normalization | `__able_struct_ManifestRecord_to_seen` | 18,432 | 49,152 | material |
| Binary Event Log | `__able_struct_EventRecord_to_seen` | 319,488 | 851,968 | material |

The normal verified workloads execute 3,072 successful Manifest callbacks and
53,248 successful Binary callbacks. Event executes 3,072 successful routing
callbacks, but those calls remain entirely on native Go carriers.

The exact profiles also reinforce why a two-family implementation would be
attractive but inadmissible: Manifest's conversion owns 2,784 KiB flat and
3,264 KiB cumulative allocation space, while Binary's owns 48,256 KiB flat
and 73,216 KiB cumulative allocation space. Event has no corresponding owner.

## CPU refresh

Ten verified main-phase CPU profiles were collected per application using
`ABLE_GO_PHASE_CPU_PROFILE_DIR` and merged for inspection.

| Application | Merged CPU samples | Candidate conversion cumulative CPU |
| --- | ---: | ---: |
| Concurrent Event Routing | 190 ms | absent |
| Manifest Normalization | 100 ms | 60 ms in `__able_struct_ManifestRecord_to_seen` |
| Binary Event Log | 1,600 ms | 920 ms in `__able_struct_EventRecord_to_seen` |

The short Event and Manifest runs make 10 ms CPU samples coarse, so CPU
percentages are supporting rather than admission evidence. The generated
source and exact allocation attribution are decisive. Binary's longer profile
independently confirms that its conversion is material.

## Verification and provenance

- All three strict builds pass their public verifiers.
- All three final dependency graphs omit `pkg/interpreter`.
- All nine exact allocation runs pass public verification.
- All thirty CPU-profile runs pass public verification.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes in 5.686 seconds.
- No compiler, runtime, interpreter, stdlib, language, dependency, or WASM
  code was changed for this candidate.
- Baseline binary SHA-256:
  - Event Routing:
    `79c53727a425e6359326bbfd8875b66fe7b8f651a363dc037d01ecb61dde4708`
  - Manifest Normalization:
    `91d0e018792d81927ed2b013cce3f0bd68547733e71d4338a12f3bef7897cab2`
  - Binary Event Log:
    `0fd2085de8914545197f02cbe3fdef573315fa8de2cd4066df022d11cf175f6b`
- Raw artifacts:
  `/tmp/able-captured-callable-20260725.nyhWWS`

## Next

Qualify the native nominal-to-static-interface conversion used while building
errors across Concurrent Event Routing, Manifest Normalization, and Binary
Event Log.

This is next because the refreshed exact profiles expose one genuine owner in
all three applications: each concrete application error is converted into a
runtime struct payload while it is encoded through the statically known
`Error` interface. The concrete-error conversion itself owns 3,072 to 36,864
flat allocation objects in the current exact profiles.

The work entails tracing the shared nominal/interface semantic encoding,
determining whether a fully static native payload can remain native without
changing observable error identity, `message`, `cause`, rescue, or dynamic
inspection behavior, and profiling at least one success-heavy and one
error-heavy workload mode. Any implementation must be a general static
interface/semantic-encoding rule; it must not name the three application error
types, `Result`, a container, or another non-primitive nominal type.

This is important because it is now the strongest representation crossing
proven in the same three unlike programs. Closing it generally would reduce
allocation and GC work on error paths while preserving the project rule that
non-primitive nominal types use the shared nominal translation pipeline.

Do not begin WASM work.
