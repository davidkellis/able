# Compiled static-native carrier closure — 2026-07-24

## Decision

Retain a general compiler correction and make native static Array lowering a
permanent compiler contract.

The compiler previously used native Go scalar signatures and slice-backed
Array carriers by default, but two holes remained:

1. the legacy `ExperimentalMonoArrays` option, CLI flags, and environment
   setting could still select the runtime-store Array representation; and
2. `_ = expression` routed every statically typed result through
   `runtime.Value` before discarding it.

The first escape hatch can no longer disable native static arrays. The old
option and tooling spellings remain accepted for compatibility, but they are
no-ops and generated source now advertises
`__able_native_static_arrays = true` rather than an experimental mode.

Wildcard discard now evaluates the right-hand side once, preserves its native
carrier and control/error result, and discards that value directly. This is a
shared assignment-lowering correction for primitives, Strings, arrays,
nominals, interfaces, unions, and callables; it is not an application,
benchmark, or named-container rule.

## Enforced representation

New generated-source guards cover every primitive family:

- `bool`, fixed-width signed/unsigned integers, `isize`, and `usize` use their
  matching Go scalar types;
- `f32` and `f64` use `float32` and `float64`;
- `char` and `String` use `rune` and `string`;
- `i128` and `u128` use the native two-word `runtime.Int128` and
  `runtime.Uint128` value structs; and
- representative primitive Array families use compiler-owned
  `Elements []T` wrappers and avoid `runtime.Value`, `runtime.ArrayValue`,
  `runtime.ArrayStore`, bridge boxing, and dynamic dispatch in the static
  function body.

The two-word wide integers live in the runtime package for shared ownership,
but they are values rather than store handles or boxed interpreter objects.

## Verifier-backed reach audit

The opt-in typed-boundary diagnostic now distinguishes runtime-to-Array and
Array-to-runtime conversions. Normal binaries still contain no telemetry
counters, atomics, or environment branches.

Six unlike portable applications were compiled, executed, and checked with
their public verifiers:

| Application | Runtime → Array | Array → runtime | Classification |
| --- | ---: | ---: | --- |
| Matrix Multiply | 0 | 0 | static numeric arrays remain native |
| Distance Field | 0 | 0 | static nested numeric arrays remain native |
| Array Slice Window | 0 | 0 | static slice/copy operations remain native |
| Reverse Complement | 1 | 0 | explicit `fs.read_bytes` return boundary |
| Option/Result Config | 0 | 0 | no Array boundary |
| Concurrent Signal Dispatch | 33 | 0 | one bounded line read plus 32 `string_to_bytes` returns |

All six rows completed and verified. No application converted a native Array
back to runtime form during its measured main phase. The 34 runtime-to-Array
events are explicit filesystem/string kernel returns, not internal Array
execution or storage.

The complete diagnostic totals were:

| Category | Count |
| --- | ---: |
| Any → runtime | 312,866 |
| Runtime → integer | 8,207 |
| Runtime → Array / Array → runtime | 34 / 0 |
| Runtime → struct / struct → runtime | 4,104 / 10,532 |
| Runtime → union / union → runtime | 49,154 / 98,112 |
| Runtime → interface / interface → runtime | 0 / 0 |
| Runtime → callable / callable → runtime | 0 / 122,880 |
| Error → control / control → error | 391,171 / 0 |

Those other crossings reproduce already-classified Option/Result and
concurrency/host boundaries. They are reach counts, not timing or CPU
attribution. The raw diagnostic JSON SHA-256 was
`98ce1f27343b1797de257decdc3467a700041da5431d78a9b839888799e7acfe`.

## Performance interpretation

This tranche makes no benchmark-speed claim. The portable applications do not
currently use top-level wildcard discard in their hot paths, and native Array
lowering was already the default, so a normal-binary A/B timing gate would be
measuring noise. The retained correction removes a broadly applicable hidden
boxing path from real programs that do use `_ = expression` and prevents
future callers from accidentally compiling static arrays through runtime
stores.

Any later speed claim still requires repeated verifier-backed normal-binary
measurements and the existing broad scorecard gate.

## Verification

- focused static primitive, native Array, wildcard-discard, native touchpoint,
  typed-boundary telemetry, extern, and Array-boundary tests pass;
- legacy compiler option, `ablec`, and `able build` settings cannot disable
  native static arrays;
- focused `cmd/able` and `cmd/ablec` tests pass;
- six of six portable audit applications execute and verify;
- the two broad short-mode batching attempts reached pre-existing cumulative
  60-second package deadlines rather than reporting a test assertion failure;
  the interrupted iterator test passes alone in 11.928 seconds, and the two
  interrupted closed-benchmark subtests pass together in 44.047 seconds; and
- modified source files remain below 1,000 lines
  (`generator_assignments.go` is 999 lines after extracting discard lowering).

## Next recommendation

Run a generated static-closure size/reach census before another runtime fast
path.

Why: primitive and Array representation is now mandatory and guarded, while
the full typed-boundary work has already shown that remaining conversion
category parents split into different CPU descendants. Generated packages
still emit large registration, compatibility, and dynamic helper surfaces
even for closed static applications. Removing provably unreachable generated
code could reduce build cost and binary/instruction footprint broadly without
changing Able value representation or naming a nominal container.

What it entails: conservatively map generated functions, registrations, and
linked symbols reachable from main, package initialization, exports, dynamic
calls, interface adapters, externs, and concurrency entry points across at
least three serial and three concurrent applications. Advance only a shared
unreachable closure that repeats broadly, then use five complete
verifier-backed normal-binary A/B processes per side and correctness guards
for dynamic fallback and package initialization. If no shared linked/startup
or instruction-cache cost appears, close this direction and return to the
largest unresolved bytecode architecture wall.
