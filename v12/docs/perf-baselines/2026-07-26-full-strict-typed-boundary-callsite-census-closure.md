# Full strict typed-boundary callsite census closure

Date: 2026-07-26

## Decision

Retain the report-only typed-boundary observer extension and no production
compiler, generated-runtime, runtime, interpreter, bytecode VM, stdlib,
benchmark, language, dependency, or WASM performance change.

All 61 current coverage applications compiled with `--no-fallbacks` and the
opt-in observer, ran once, and passed their public verifiers. All 61 final Go
dependency graphs omit `able/interpreter-go/pkg/interpreter`; dependency
counts range from 96 to 119 packages.

The census found 107 observed concrete conversion shapes. Fourteen reached at
least three applications, but none is both new and material in three unlike
Go-target misses. The only repeatedly executed primitive leaf with broad
reach is opaque channel-handle recovery. It belongs exclusively to the
explicit concurrency-service ABI and is already closed by normal-binary
profiles in which the enclosing `__able_int64_from_value` helper was 0.11%
flat in Channel Rollup and 0.22% in Future Pipeline. Mutex handles and
Awaitable/Future adapters repeat only in their related concurrency family.

The machine-readable companion is
`2026-07-26-full-strict-typed-boundary-callsite-census-closure.json`.

## Observer retained

The existing opt-in `--typed-boundary-telemetry` mode now records both category
totals and nonzero concrete shapes. Each shape reports:

- generated conversion function;
- Able source path and span, or an explicit compiler-runtime semantic site;
- source carrier;
- immediate consumer;
- semantic reason; and
- exact execution count.

Generated diagnostic code uses fixed counters and compile-time metadata.
There is no runtime stack walk or dynamic attribution map in a conversion hot
path. The ordinary diagnostics-off compiler output remains free of the
observer, enforced by the focused opt-in regression guard.

The initial census exposed one parent that was still too broad:
`runtime.Value` to `int64`. The observer was refined by the helper's semantic
label and the 17 affected applications were rebuilt and publicly verified.
This separated channel handles, channel capacity, and mutex handles rather
than ranking them as one primitive-conversion parent.

## Frozen protocol

- Repository commit:
  `237406eccdfb025a519d898daedadee1c8d13a7b`.
- Full-corpus compiler SHA-256:
  `7a8f7386d0d3e623008a660ced6b594e9cd6488b9539f78e3f375c56044d63be`.
- Refined-observer compiler SHA-256:
  `b828cdfd9fe7ebfc479e132235625eccc64047d699fd34cc8575d62fb69455a0`.
- Full-corpus comparison JSON SHA-256:
  `dbeaa5b54f7cc88fcb07a7110406697a7c8727ab1fa3d94ec6589537940b32f7`.
- Integer-refinement comparison JSON SHA-256:
  `2d63b99f883451afd944840bfe6a77f21ed572b3e773fb7051920f9f80ee114f`.
- All generated work used disk-backed
  `/var/tmp/able-typed-boundary-census-20260726`.

The diagnostic runs are reachability evidence, not performance measurements:
every observed transition performs atomic increments. Governing performance
remains the diagnostics-off five-run strict scorecard and the retained
normal-binary CPU/allocation profiles. The workstation-noise rule therefore
does not permit interpreting the one-run diagnostic wall times.

## Recurrent exact shapes

| Concrete shape | Applications | Events | Disposition |
| --- | ---: | ---: | --- |
| error to compiled control | 44 | 15,649,215 | already closed; successful path is a nil check and absent as a shared sampled owner |
| `Array<String>` from runtime | 33 | 33 | one ingress conversion per application |
| `IOError or String` from runtime | 19 | 19 | one file/result boundary per application |
| channel handle to `int64` | 14 | 238,838 | explicit concurrency-service ABI; already profiled below 1% |
| channel capacity to `int64` | 14 | 30 | one to four construction validations per application |
| `any` to runtime | 10 | 15 | startup-only reach |
| compiled control to error | 4 | 18,945 | explicit runtime-callable ABI; already closed |
| `IOError or i32` from runtime | 4 | 4 | one program-result boundary per application |
| mutex handle to `int64` | 3 | 43,016 | three related mutex applications only |
| `Awaitable<i64>` from runtime | 3 | 8,704 | three related scheduler applications only |
| `Awaitable<i64>` to runtime | 3 | 8,704 | three related scheduler applications only |
| `Array<Awaitable<i64>>` to runtime | 3 | 7,168 | same related scheduler family |
| `Future<i64>` from runtime | 3 | 1,544 | one material application; four events in each of two controls |
| `Error or i64` from runtime | 3 | 1,544 | one material application; four events in each of two controls |

The source-attributed nominal shapes do not reproduce the prior broad
`runtime.NewStructInstancePositionalSized` parent across three concrete Able
types. Each escaping nominal value remains attached to its own definition and
semantic boundary. Grouping those distinct types under allocation or
`runtime.Value` would violate the admission rule.

## Why profiling stopped at the gate

No new exact shape cleared all three prerequisites: unlike-program breadth,
plausible materiality, and an unclosed general compiler/runtime mechanism.

- The broad control route and primitive integer helper were already rejected
  by diagnostics-off profiles.
- Channel, mutex, Future, and Awaitable leaves are explicit scheduler service
  boundaries and remain one feature family.
- String Array and I/O union conversions occur once per application.
- `any` conversion is startup noise.
- Concrete nominal encode/decode shapes are definition-specific and do not
  repeat broadly.

Refreshing CPU/allocation profiles for these same closed or startup-only
leaves would contradict the tranche's closed-route exclusion. Because no new
profile-material shape reached the breadth gate, no prototype or five-run A/B
cohort was warranted.

## Verification

- 61/61 strict telemetry builds completed and publicly verified.
- 17/17 semantic-refinement builds completed and publicly verified.
- 61/61 generated top-level source trees omit the interpreter import.
- 61/61 final `go list -deps` graphs omit `pkg/interpreter`.
- Dependency graph sizes are 96 packages for ordinary rows and 119 for the
  existing Base64 dependency surface.
- Focused observer opt-in and integer-shim guards pass.
- The complete `./run_all_tests.sh` handoff passes every contract,
  non-compiler package, all 32 compiler batches, and the final 86.857-second
  bytecode fixture pass. The three previously audited aggregate compiler
  batches took 186.551, 75.252, and 86.757 seconds.
- No canonical `able-stdlib` source changed.

## Next recommendation

Audit semantic-work equivalence for strict misses that are already pure native
Go, beginning with Tapelang Alphabet, Fib, and Sudoku Masks, with Matrix
Multiply and Pidigits as parity controls.

Why: the full corpus now shows that residual compiled/runtime conversions do
not provide one new cross-family owner. The remaining large misses spend their
time in generated native application functions, Go allocation, `math/big`, or
required runtime services.

What it entails: compare generated Go and equivalent Go source plus assembly;
count loop iterations, calls, allocations, bounds/overflow checks, and
semantic adapters; reconcile input/output work; and advance only one
compiler-added redundant operation that repeats in at least three unlike
applications.

Why it is important: native carriers can match Go only if the generated code
performs equivalent work. This audit tests that remaining possibility without
weakening nominal semantics, reopening the compiler/interpreter boundary, or
adding benchmark, named-container, or non-primitive nominal rules. Continue
to defer WASM.
