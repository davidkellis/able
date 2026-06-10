# Compiled nullable-scalar allocation attribution

Date: 2026-07-18

## Decision

Complete the exact main-phase allocation attribution gate and retain no
compiler, generated-runtime, canonical-stdlib, benchmark, fixture, or language
change.

The proposed generic internal value-plus-present ABI for nullable primitive
results does not pass the project's three-unlike-application admission rule.
Across I-Before-E, Word Frequency, and Unicode Scalar Pipeline, exact
allocation profiles attribute no allocation to any generated `__able_ptr`
source line. The only observed nullable primitive allocation is Dependency
Plan's generic `Deque<i32>.pop_front() -> ?i32`: 3,072 objects and 48 KiB in
each of three runs. One application is not sufficient evidence for an ABI
rewrite.

## Measurement contract

The current compiler built one normal executable per application before any
profile launch. Each executable then received three independent allocation-only
phase-profile processes with:

- CPU 0, `GOMAXPROCS=1`, `GOGC=50`, and `GOMEMLIMIT=1GiB`;
- the catalog working directory and arguments;
- the canonical external stdlib selected by the normal compiler driver;
- a 55-second process cap; and
- the application's public Ruby verifier.

All twelve processes completed and verified. Each application produced one
stable stdout SHA-256 across its three runs. The exact `runtime.MemStats`
main-phase counters were identical across repeats:

| Application | Mean profiled wall | Main bytes | Main allocations | `__able_ptr` allocation |
| --- | ---: | ---: | ---: | --- |
| I-Before-E | 1.137561 s | 9,630,064 | 15,901 | none observed |
| Word Frequency | 2.815553 s | 31,184,904 | 720,431 | none observed |
| Unicode Scalar Pipeline | 4.921004 s | 44,129,432 | 3,068,210 | none observed |
| Dependency Plan | 0.891396 s | 475,192 | 18,631 | 3,072 objects / 49,152 bytes |

The wall values describe exact-profiler overhead and are not product timing
claims. Repetition establishes stable counters and attribution on a variable
workstation rather than treating one process as decisive.

## Exact caller attribution

Every start/end profile pair was subtracted twice, once for allocated objects
and once for allocated bytes, with line attribution and no node-count cutoff.
The set of generated source lines containing `__able_ptr` was then intersected
with the nonzero allocation lines.

- I-Before-E: no intersection in any of three object or byte profiles.
- Word Frequency: no intersection in any of three object or byte profiles.
- Unicode Scalar Pipeline: no intersection in any of three object or byte
  profiles.
- Dependency Plan: all six intersections identify only
  `__able_compiled_method_Deque_pop_front_spec` returning
  `__able_ptr(value)`. Each object profile reports exactly 3,072 allocations;
  each byte profile reports 48 KiB.

Relative to Dependency Plan's authoritative main counters, that is 16.49% of
allocation objects and 10.34% of allocated bytes. It is material locally, but
it does not repeat across applications. The generated `read_byte(Array<u8>,
i32) -> ?u8` helper remains reachable in the three text binaries, but its
success pointer either is not entered on the native valid-UTF-8 path or is
eliminated by Go; it contributes no measured heap allocation here.

The start/end profile difference includes allocation-profile serialization,
especially in short applications. Those profiler frames are outside the
authoritative main-phase counters and were not interpreted as application
work. They also cannot create a false generated `__able_ptr` source-line
sample.

## Separate nominal-union result

The text profiles expose a different repeated allocation:

| Application | `Utf8DecodeResult` objects | Bytes |
| --- | ---: | ---: |
| I-Before-E | 854 | 13.34 KiB |
| Word Frequency | 65,490 | 1,023.28 KiB |
| Unicode Scalar Pipeline | 1,119,810 | 17.09 MiB |

These allocations are the pointer-backed nominal payload of
`Utf8DecodeResult | StringEncodingError`, not a nullable primitive result and
not a `read_byte` allocation. The existing caller-owned nominal-result ABI
does not apply because the function returns a union carrier. Special-casing
`Utf8DecodeResult`, String, or the benchmark applications would violate the
shared nominal-lowering rule, so this finding does not admit a code candidate
in the present tranche.

## Verification and cleanup

- 4/4 preserved compiled builds completed under their bounded build process.
- 12/12 exact allocation-only profile processes passed their public verifier.
- 12/12 main allocation counters and per-application stdout hashes repeated
  exactly.
- 24/24 object/byte profile subtractions completed; the nullable source-line
  result repeated exactly across all three runs per application.
- `go test ./pkg/profilehook -count=1 -timeout 60s` passed.
- The focused caller-owned-result group and the focused `^TestCompilerNullable`
  group passed in 1.294 seconds and 5.160 seconds.
- `TestCompilerSpecNullableExpectationBuilds` is not green under the mandated
  one-minute limit: two independent standalone launches timed out at 60
  seconds, one during generated-code compilation and one in its isolated-cache
  nested Go build. This tranche did not cause the failure because it changes no
  production source, but the existing test must be made sub-minute before it
  can be used as a broad handoff guard.
- No production source or canonical stdlib change was made.

Raw generated trees, executables, stdout captures, and profiles were temporary
decision evidence and are removed after this compact record is written.

## Next recommendation

Run a coverage-wide structural census for fresh small nominal payloads returned
inside statically known unions, followed by bounded dynamic attribution only
for repeated shapes.

Why: the nullable-scalar branch is now closed on current evidence, while the
same pointer-backed `Utf8DecodeResult` payload is a concrete allocation owner
in three applications. That is enough to investigate a shared mechanism, but
not enough to optimize one named stdlib struct. A full-suite census can show
whether the existing caller-owned nominal-result proof can be generalized to
union payload transport across at least two nominal definitions and three
unlike applications.

What it entails: first make the existing spec-nullable build guard complete
within one minute without weakening its compile-and-build assertion. Then
inventory generated union-return functions whose success or error payload is a
fresh small scalar nominal, classify their static callers, and measure only
shapes that immediately pattern-test/deconstruct the result. Admit a prototype
only if multiple nominal definitions are materially represented. Any candidate
must be structural, preserve ordinary pointer carriers at
dynamic/interface/host boundaries, let Go escape analysis decide stack versus
heap placement, and pass retained-identity, alias, early-return, error-union,
and non-text controls. If definition breadth or dynamic materiality fails,
close this family and select the largest remaining shared compiled or bytecode
scorecard wall. Continue to defer WASM.
