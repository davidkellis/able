# Post-nullable compiled float-numeric reconciliation

## Decision

Reconcile `compiled-float-numeric` as causally current and retain no
production change.

All four strict applications already lower their material floating-point
work to native Go `float64` values. Their generated application bodies
contain zero `__able_nullable[float64]` values and zero float recovery through
`runtime.Value`. N-Body's `Array<f64>` reads and writes are direct native
`Elements[index]` operations.

The retained primitive nullable carrier reaches only one different shape:
N-Body's `advance` function returns an allocation-free
`__able_nullable[int32]` because its final loop expression is optional. The
caller ignores that value. No corresponding changed carrier occurs in
Distance Field, Mandelbrot, or RMS Norm, so no post-carrier residual has the
required three-unlike-application breadth.

## Strict boundary and execution gate

Every application was rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Distance Field | 96 | absent | 1/1 |
| Mandelbrot | 96 | absent | 1/1 |
| N-Body | 96 | absent | 1/1 |
| RMS Norm | 96 | absent | 1/1 |

Smoke durations were 0.03, 0.08, 0.08, and 0.04 seconds respectively. They
are execution checks, not timing evidence. The authoritative scorecard
retains five verifier-backed Able and Go processes per row.

## Generated native-carrier audit

Only generated application functions and their materially called numeric
helpers count as reach. Dormant conversion support emitted into every module
does not.

| Application | Nullable `f64` sites | Material lowering |
| --- | ---: | --- |
| Distance Field | 0 | scalar `float64` loop; direct native `hypot`/`sqrt`; one explicit output conversion |
| Mandelbrot | 0 | fused `float64` pixel loop; direct native `Array<u8>` writes; no bridge in application bodies |
| N-Body | 0 | native `__able_array_f64` storage and direct `float64` element reads/writes |
| RMS Norm | 0 | scalar `float64` loop and direct native `sqrt`; two explicit output conversions |

N-Body's six material application functions contain 148 direct `.Elements`
references and no nullable-float carrier. Array reads compile to a native
bounds check followed by a `float64` load. Array writes compile to a bounds
check followed by direct assignment.

The hot `advance` body returns `__able_nullable[int32]` only to preserve the
language result of its final loop. It contains no nullable `float64`, no
float bridge call, and no result allocation. Its caller receives and discards
the value directly. This is causal reach of the changed representation in one
application, but not a shared float-numeric owner.

The remaining `bridge.ToFloat64` sites occur only when Distance Field, N-Body,
and RMS Norm pass their final results to the dynamic `print` host boundary.
They execute once or twice per application, outside the numeric loops.
Mandelbrot's byte output uses the static I/O path and has no application-body
bridge call.

## Retained profile evidence

The current post-nullable architecture gate already profiles Distance Field
from the same source and compiler state:

- ten verified CPU-profile processes merge to 160 ms;
- ownership is native generated `main`, `hypot`, `sqrt`, and `math.Sqrt`;
- three exact main-phase allocation processes average 512 bytes, 11 objects,
  and zero GC cycles; and
- no material Able allocation owner exists.

The earlier cross-mode numeric gate separates the other compiler owners:
Mandelbrot spends 95.42% cumulative CPU in its generated primitive
`pixel_byte` loop, while Distance Field and RMS Norm are native square-root
geometry. N-Body now visibly uses the same native `float64` arithmetic and
square root over direct primitive Arrays.

This evidence leaves no lower representation beneath the shared arithmetic:
`math.Sqrt` and Go `float64` operations are already the target native
operations. The prior normalized raw-float allocation, operand-lane,
sidecar, scalar-return, and producer-fusion candidates apply to the bytecode
VM. Their multi-program wall-time gates regressed guarded applications and
are not invalidated by the compiled nullable carrier.

## Selective profile and candidate gate

Fresh profiling was not admitted:

- three rows have no changed-carrier reach;
- N-Body's changed nullable result is unique, ignored, and allocation-free;
- native `float64` arithmetic and `math.Sqrt` expose no missing compiler or
  generated-runtime conversion to remove; and
- explicit output conversions are required host boundaries and execute at
  negligible frequency.

No exact changed residual therefore reaches three unlike rows. A fresh CPU,
allocation, timing, or A/B cohort would repeat native ownership evidence
without identifying an implementable general rule.

## Current row state

The current five-process scorecard means remain:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Distance Field | 0.0380 s | 0.0147 s | 2.5850x |
| Mandelbrot | 0.1100 s | 0.0533 s | 2.0638x |
| N-Body | 0.0980 s | 0.0354 s | 2.7684x |
| RMS Norm | 0.0300 s | 0.0140 s | 2.1429x |

These remain product misses, but their material generated numeric structures
already correspond to native Go primitives. The current evidence cannot
attribute the residual gap to float boxing or interpreter fallback.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No named-container, non-primitive nominal, wide-type, or benchmark-specific
rule was introduced.

`go test ./cmd/ablec` passed in 5.430 seconds. The machine-readable record is
`2026-07-30-post-nullable-compiled-float-numeric-reconciliation.json`.

After retaining this evidence, the exact 564 MiB disk-backed generated
module, binary, audit, and Go-cache workspace was removed. Four accidentally
created standalone application-body audit files were also removed. No
matching tranche artifact remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-wide-numeric` against the primitive nullable carrier.

Why: Fixed Width 128, Rational Series, and Wide Integer Records are the
smallest remaining invalidated closure with exactly three unlike rows.
Primitive `i128` and `u128` are covered by the value-carrier architecture,
while these programs also exercise shared nominal Result and record
translation that must remain generic.

What it entails: strictly rebuild all three applications, verify their graphs
remain interpreter-free, and distinguish primitive wide nullable values from
non-primitive Rational, Result, and record encodings. Trace material
DivMod/quotient/result paths and reuse the retained wide profiles; profile
again only if the changed carrier reaches one still-open concrete residual in
all three rows.

Why it matters: the rows remain roughly 3.00x-17.97x Go. Causal review can
either locate a general primitive-wide boundary or prove that the remaining
cost belongs to nominal algorithms and already-closed package/context work,
without introducing forbidden special cases for named wide structures.
