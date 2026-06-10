# Compiled String split environment-entry closure

Date: 2026-07-25

## Decision

Retain no compiler, generated-runtime, canonical-stdlib, or runtime change from
this tranche. The exact imported-method environment-entry owner does not repeat
across the required three unlike applications.

Concurrent Event Routing pays for
`__able_compiled_entry_method_String_split -> bridge.SwapEnvIfNeeded ->
bridge.currentGID -> runtime.Stack`. Word Frequency and Sensor Calibration call
the same generated entry symbol, but their runtimes remain in serial
environment mode and neither profile contains a sampled environment swap,
goroutine lookup, or stack recovery beneath that entry.

The shared `String.split` entry symbol is therefore only a cumulative parent of
the language-level split body in the two serial applications. Treating it as an
exact shared boundary owner would violate the admission rule.

No production code was added, removed, or changed. No Able syntax, language,
interpreter, tree-walker, bytecode VM, dependency, or WASM work was performed.

Machine-readable samples are in
`2026-07-25-compiled-string-split-environment-entry-closure.json`.

## Cohort and settings

Fresh `--no-fallbacks` binaries were built for:

- Concurrent Event Routing;
- Word Frequency; and
- Sensor Calibration.

All three public verifiers passed, and `go version -m` confirmed that every
strict binary omitted `able/interpreter-go/pkg/interpreter`.

Five clean main CPU profiles and three exact main allocation samples were
captured per application with `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`,
and `ABLE_EXECUTOR=goroutine`.

## Exact CPU attribution

| Application | Merged CPU | `String.split` entry | `SwapEnvIfNeeded` | `currentGID` | `runtime.Stack` |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 800 ms | 670 ms (83.75%) | 140 ms (17.50%) | 240 ms (30.00%) | 230 ms (28.75%) |
| Word Frequency | 370 ms | 340 ms (91.89%) | 0 | 0 | 0 |
| Sensor Calibration | 800 ms | 580 ms (72.50%) | 0 | 0 | 0 |

The entry's large cumulative percentage in Word and Sensor belongs to its
callee, `__able_compiled_method_String_split`, and that method's string/byte
conversion, UTF-8, Array, and allocation descendants. It is not time spent
establishing a package environment.

The difference follows the existing bridge contract. `Runtime.MarkConcurrent`
switches Event to per-goroutine environments, so its entry swap must recover a
goroutine identity. Word and Sensor never enter concurrent runtime mode;
`Runtime.SwapEnv` uses the directly stored process environment and does not call
`currentGID`.

The existing environment-effect proof also remains conservative in the correct
direction. It already proves ordinary imported methods that depend only on
arguments and their proven-independent generated callees. `String.split`
retains its compatibility entry through its larger generated callee/control
graph. This tranche supplied no three-program evidence that broadening that
proof is semantically safe or useful.

## Allocation evidence

The exact allocation profiles reinforce the CPU distinction:

| Application | Mean bytes | Mean allocations | Mean GC | `currentGID` objects |
| --- | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 44,657,909 | 955,870 | 21.33 | 8,207 |
| Word Frequency | 29,474,259 | 690,057 | 31.33 | 0 |
| Sensor Calibration | 63,185,499 | 1,521,370 | 64.00 | 0 |

The `String.split` entry is again only a cumulative allocation parent in Word
and Sensor. Event's 8,207 `currentGID` allocation objects are real, but an
Event-only environment-entry correction would not clear the broad benchmark
bar.

The profiles do expose a different exact shared boundary inside the split
implementation:

| Exact generated allocation site | Event objects | Word objects | Sensor objects |
| --- | ---: | ---: | ---: |
| `__able_string_from_builtin_impl` | 164,736 | 134,549 | 293,376 |
| `__able_array_u8_to` | 40,320 | 43,927 | 98,048 |
| `__able_string_to_builtin_impl` | 27,952 | 25,338 | 57,084 |

These are primitive String/static `Array u8` conversion boundaries, not the
rejected package-entry hypothesis. They are recorded as evidence for the next
tranche, not optimized here.

## Current Go gap

After two warmups per binary, twenty alternating Able/Go process pairs ran on
CPU 9. All 120 timed processes passed the public verifiers:

| Application | Able mean | Go mean | Able / Go | Go performance |
| --- | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 178.479 ms | 2.865 ms | 62.291x | 1.61% |
| Word Frequency | 80.726 ms | 3.891 ms | 20.744x | 4.82% |
| Sensor Calibration | 185.858 ms | 3.639 ms | 51.069x | 1.96% |

The Go gaps remain far outside the 1.052632x target. They justify continued
compiled lowering work, but not an unsupported environment-effect change.

## Verification

- 24 Able profile processes passed their public verifiers.
- 120 repeated Able/Go timing processes passed their public verifiers.
- All three strict dependency graphs omit the interpreter.
- JSON evidence validates successfully.
- No production candidate existed, so no candidate-specific correctness or A/B
  gate was necessary.

## Next

Trace and localize the exact static primitive String/static `Array u8` kernel
conversion boundary shared by all three applications. In current generated
code, `validated_bytes` boxes a native Go string through
`__able_String_from_builtin`, constructs a runtime-value Array, and immediately
converts it back into a native `*__able_array_u8`. Split segments take the
inverse route through `__able_array_u8_to` and `__able_String_to_builtin`.

The next tranche should establish a general typed kernel-helper path that keeps
primitive String and statically typed `Array u8` values on their native Go
carriers. Dynamic, externally supplied, invalid, nullable, and erased calls
must retain the current runtime-value compatibility path and exact errors. Do
not special-case `String.split`; the rule must apply to every statically typed
call to the same language/kernel conversion boundary.

This is next because `__able_string_from_builtin_impl` alone owns 18.16%-21.42%
of allocation objects in all three unlike programs, and the paired Array/String
conversion sites repeat exactly in every profile. The work entails tracing
static helper registration and result coercion, adding native-carrier helpers
with verifier-backed invalid-input guards, and running repeated A/B cohorts
against equivalent Go. It is important because this is a concrete remaining
case where already-native primitive String and Array values are unnecessarily
boxed across the compiled/runtime boundary.
