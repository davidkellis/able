# Compiled native UTF-8 validation retained

Date: 2026-07-25

## Decision

Retain the general canonical primitive String validation path.

When `able.text.string.utf8_validate` receives its inferred native
`*__able_array_u8` carrier, generated code now uses Go's allocation-free
`utf8.Valid` for valid input. Nil or invalid input falls through to the
unchanged Able implementation, preserving the precise `StringEncodingError`
type, message, and byte offset.

Twenty order-balanced baseline/candidate/Go cohorts improved all three unlike
applications: Concurrent Event Routing by 2.37%, Word Frequency by 16.03%,
and Sensor Calibration by 11.67%, for a 10.20% geometric-mean improvement.
Exact main-phase allocation bytes fell 7.84%-16.70%, and allocation counts
fell 37.15%-49.19%.

No application, benchmark, `String.split`, named container, or non-primitive
nominal type is selected by the compiler rule. No canonical stdlib, runtime
package, interpreter, tree-walker, bytecode VM, language, dependency, or WASM
change was needed.

Machine-readable evidence is in
`2026-07-25-compiled-string-utf8-validation-retained.json`.

## Owner analysis

The previous profile attributed 77,595 Event, 65,485 Word, and 105,994 Sensor
allocation objects to successful `utf8_decode` results. The decoder already
used native `utf8.DecodeRune`; its remaining allocation was the mutable
`Utf8DecodeResult` nominal record carried through the union ABI.

A caller-owned-result prototype was tested and rejected. On Word Frequency it
moved the 65,484-object profile leaf from `utf8_decode` to its
`utf8_validate` caller, while exact main allocation remained 6,269,888 bytes
and 266,044 allocations. Go correctly heap-promoted the caller slot because
the nominal result pointer entered the union interface. None of that prototype
is retained.

Call-graph localization then showed that the largest repeated decode owner was
the full-input `utf8_validate` pass. Valid native byte Arrays need only a
boolean validity decision; they do not need an observable decode-result
record. The retained prefix therefore avoids constructing a nominal record or
union on the overwhelmingly common valid path. The original Able loop remains
the invalid-input diagnostic path.

`slice_bytes` remains unchanged because it returns a fresh mutable Array and
the language requires slice ownership and mutation isolation.

## Repeated wall time

After two verified warmups per lane, twenty order-balanced
baseline/candidate/Go process cohorts ran on CPU 9 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and the application-appropriate executor.
Every timed process passed its public verifier.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 107.140 ms | 104.598 ms | -2.37% | 3.077 ms | 33.990x | 2.94% |
| Word Frequency | 22.223 ms | 18.661 ms | -16.03% | 3.765 ms | 4.957x | 20.17% |
| Sensor Calibration | 28.859 ms | 25.491 ms | -11.67% | 3.259 ms | 7.821x | 12.79% |

The retained rule improves all three rows, but these applications remain
outside the 1.052632x compiled target. It removes one allocation-heavy native
String semantic loop; it does not close the remaining generic runtime bridge
costs.

## Exact main-phase allocation

Three exact main-phase samples per candidate were compared with the
immediately preceding retained three-sample means.

| Application | Bytes baseline -> candidate | Change | Allocations baseline -> candidate | Change | GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 15,936,829 -> 14,687,085 | -7.84% | 421,382 -> 264,834 | -37.15% | 10.00 -> 9.33 |
| Word Frequency | 6,269,920 -> 5,222,677 | -16.70% | 266,044 -> 135,165 | -49.19% | 6.33 -> 5.67 |
| Sensor Calibration | 11,938,891 -> 10,255,365 | -14.10% | 534,977 -> 324,801 | -39.29% | 11.33 -> 9.33 |

The candidate profiles contain no `utf8_validate` or validation-owned
`utf8_decode` allocation leaf. The next exact allocation owners common to all
three are `bridge.ToString` with 38,274, 35,959, and 48,898 objects, and
`slice_bytes` with 33,101, 33,684, and 63,991 objects.

## Verification

- The generated-body guard requires the canonical package/name/signature and
  the native `utf8.Valid` prefix; a differently named function is rejected.
- Existing executable guards preserve valid multibyte character iteration and
  invalid-byte `StringEncodingError` behavior.
- All 180 timed processes and all 18 warmups passed public verification.
- All 30 candidate CPU-profile and nine exact-allocation-profile processes
  passed public verification.
- Ten candidate main-phase CPU profiles and three exact allocation profiles
  were captured per application.
- All three fresh `--no-fallbacks` dependency graphs omit `pkg/interpreter`.
- Focused UTF-8 generated/executable guards pass in 20.955 seconds.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes in 8.099 seconds.
- Touched source files remain below 1,000 lines.

## Next

Localize the exact `bridge.ToString` allocation owner through all three
post-retention call graphs and determine whether static String carriers cross
a generic nominal/interface boundary unnecessarily.

This is next because `bridge.ToString` is now the largest explicit
compiled/runtime conversion site that repeats materially in every
application, while `slice_bytes` has a known fresh-mutable-Array ownership
obligation. The work entails tracing every `bridge.ToString` caller, separating
host formatting from generic Map/interface conversion, and admitting at most
one shared nominal/interface carrier rule with generated-source, semantic,
strict-dependency, exact-allocation, and repeated A/B gates. It is important
because this is the clearest remaining shared place where already-native
compiled String values may still be boxed solely to cross a generic runtime
boundary.

Do not begin WASM work.
