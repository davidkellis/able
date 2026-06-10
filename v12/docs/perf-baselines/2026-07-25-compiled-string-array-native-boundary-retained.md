# Compiled String/static Array native boundary retained

Date: 2026-07-25

## Decision

Retain the general typed kernel-helper path for
`__able_String_from_builtin` and `__able_String_to_builtin`.

When the argument is already proven to use a native primitive String or static
`Array u8` carrier, generated code now converts directly between Go `string`
and `*__able_array_u8`. It no longer boxes every byte into `runtime.Value`,
constructs a runtime Array, and immediately converts that Array back to its
native carrier. Calls with erased or dynamic arguments keep the existing
runtime compatibility helpers.

The same rule improved all three unlike strict applications. Twenty
order-balanced baseline/candidate/Go process cohorts improved by 45.68%-85.35%,
with a 72.47% geometric mean. Main-phase allocation bytes fell
64.31%-81.11%, with a 75.70% geometric mean, and allocation counts fell
55.92%-64.84%, with a 60.90% geometric mean.

No application, benchmark, `String.split`, named container, or non-primitive
nominal type is selected by the compiler rule. No canonical stdlib, runtime
package, interpreter, tree-walker, bytecode VM, language, dependency, or WASM
change was needed.

Machine-readable evidence is in
`2026-07-25-compiled-string-array-native-boundary-retained.json`.

## Lowering change

The shared primitive helper table now recognizes the two String kernel
conversions only when their argument expressions already compile to the exact
native carriers:

- Go `string` -> `*__able_array_u8`;
- `*__able_array_u8` -> Go `string`.

The forward helper copies string bytes into a mutable native Array. The inverse
helper preserves nil rejection and UTF-8 validation before producing a Go
string. Computed arguments retain normal evaluation order and generated error
control.

The exact-native admission check is deliberate. An `any`/`runtime.Value`
argument is not coerced merely to qualify for this path; it continues through
`__able_string_from_builtin_impl` or
`__able_string_to_builtin_impl`. Extern registration and dynamic compatibility
therefore remain unchanged.

Fresh generated sources for Concurrent Event Routing, Word Frequency, and
Sensor Calibration contain no call to either runtime implementation outside
the retained helper definitions. Their static String implementations use the
native helper calls throughout.

## Repeated wall time

After two verified warmups per lane, twenty order-balanced
baseline/candidate/Go process measurements ran on CPU 9 with
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and the application-appropriate
executor. Every timed process passed its public verifier.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 194.722 ms | 105.768 ms | -45.68% | 3.109 ms | 34.016x | 2.94% |
| Word Frequency | 80.971 ms | 21.228 ms | -73.78% | 3.792 ms | 5.598x | 17.86% |
| Sensor Calibration | 190.589 ms | 27.923 ms | -85.35% | 3.369 ms | 8.289x | 12.06% |

The retained change is large and uniform, but these applications remain
outside the 1.052632x compiled target. It removes one concrete compiled/runtime
boundary rather than closing the full Go gap.

## Exact main-phase allocation

Three exact main-phase samples per candidate were compared with the immediately
preceding three-sample baselines:

| Application | Bytes baseline -> candidate | Change | Allocations baseline -> candidate | Change | GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 44,657,909 -> 15,936,829 | -64.31% | 955,870 -> 421,382 | -55.92% | 21.33 -> 10.00 |
| Word Frequency | 29,474,259 -> 6,269,920 | -78.73% | 690,057 -> 266,044 | -61.45% | 31.33 -> 6.33 |
| Sensor Calibration | 63,185,499 -> 11,938,891 | -81.11% | 1,521,370 -> 534,977 | -64.84% | 64.00 -> 11.33 |

The former `__able_string_from_builtin_impl`, `__able_array_u8_to`, and
`__able_string_to_builtin_impl` allocation sites disappear from static
application execution. The native helpers still perform required ownership
copies: Able byte Arrays are mutable, while Go strings are immutable.

## Verification

- Generated-source guards cover literal and computed static arguments and
  reject runtime boxing/Array bridge calls.
- Erased forward and inverse calls retain the runtime implementation path.
- An executable invalid-byte guard preserves the UTF-8 error.
- Existing dynamic String compatibility and normalized runtime-value guards
  pass.
- Existing static Array byte access and direct kernel-call guards pass.
- All 180 timed baseline/candidate/Go processes passed public verification.
- All 39 candidate CPU/allocation profile processes passed public
  verification.
- Ten candidate CPU profiles and three exact allocation profiles were captured
  per application.
- All three fresh `--no-fallbacks` dependency graphs omit
  `pkg/interpreter`.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes.
- A combined 44-test String/runtime selector reached the cumulative 60-second
  package deadline without an assertion failure. Its in-flight
  `String.contains` invalid-receiver guard passes alone in 11.177 seconds;
  candidate-specific guards pass together in under two seconds.

## Next

Trace the shared native String split pipeline rooted at
`__able_compiled_fn_utf8_decode`, then localize its exact
`bridge.ToString` and `__able_compiled_fn_slice_bytes` children.

This is next because the refreshed exact allocation profiles now agree across
all three unlike applications: `utf8_decode` is the largest generated-code
allocation leaf with 77,595, 65,485, and 105,994 objects, while
`bridge.ToString` and `slice_bytes` also repeat materially in every row. The
work entails following their inferred parameter/result carriers and
Result/error encodings, separating required byte ownership from avoidable
runtime conversion, and admitting at most one general typed lowering rule
behind generated-source, semantic, strict-dependency, and repeated A/B gates.
It is important because this is now the largest common place where the
interpreter-free compiled String pipeline may still leave native Go carriers.

Do not begin WASM work.
