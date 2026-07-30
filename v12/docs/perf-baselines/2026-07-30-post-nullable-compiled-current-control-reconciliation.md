# Post-nullable compiled current-control reconciliation

## Decision

Reconcile `compiled-current-control` as causally current and retain no
production change.

The primitive nullable value carrier does not reach the material generated
bodies in Fib, Matrix Multiply, or TapeLang Alphabet. Fib and Matrix have no
nullable carrier reference anywhere in their generated application functions.
TapeLang reaches the carrier only while parsing 14 loop closures once before
execution; the generated `execute` and `Tape` methods that own its runtime have
no carrier reference.

No new CPU or allocation profile was warranted. The retained current owner
profiles remain causally valid, and no owner shared by three unlike
applications was exposed.

## Strict build and boundary gate

All three applications were rebuilt from the post-carrier compiler with
`--no-fallbacks`.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Fib | 96 | absent | 1/1 |
| Matrix Multiply | 96 | absent | 1/1 |
| TapeLang Alphabet | 96 | absent | 1/1 |

The smoke processes are execution checks, not timing evidence. The
authoritative post-carrier scorecard already contains five verifier-backed
Able and Go processes for each row.

## Causal reach

The generated application function and entry-wrapper census is:

| Application | Generated functions/entries | Carrier references | Material-owner references |
| --- | ---: | ---: | ---: |
| Fib | 8 | 0 | 0 |
| Matrix Multiply | 10 | 0 | 0 |
| TapeLang Alphabet | 26 | 2 | 0 |

Dormant support definitions in the complete generated module were excluded
from reach. Every generated module defines conversion helpers for supported
nullable primitives, but definitions alone cannot affect an application that
does not invoke them.

Fib's current owner is direct generated recursion. Its retained main profile
places 99.69% flat CPU in `fib`, and the current recursive body contains only
native `i32` operations and its two direct recursive calls.

Matrix's current owner is its direct generated nested `f64` multiplication
loop. The post-truthiness/cast profile places 99.31% flat CPU in `matmul`; that
body has no nullable construction, conversion, or callee.

TapeLang contains one carrier site in `parse_program`: popping a loop start
from the parser stack produces a nullable `i32`. The 275-byte public workload
has 14 `]` bytes, so this construction occurs only 14 times before the
execution loop. The retained main profile attributes 64.57% flat CPU to
`execute`, 26.97% to `Tape.inc`, 6.30% to `Tape.get`, and 0.98% to
`Tape.move`; `parse_program` has zero samples. Those material bodies contain
zero carrier references.

## Current row state

The authoritative post-carrier five-process means are:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Fib | 4.0080 s | 3.1597 s | 1.2685x |
| Matrix Multiply | 1.1860 s | 1.0241 s | 1.1581x |
| TapeLang Alphabet | 3.7680 s | 3.0243 s | 1.2459x |

These remain target misses, but the nullable carrier cannot explain their
gaps. Their exact owners remain separate native generated bodies, so the
three-unlike-application admission gate still has no candidate.

## Scope

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No checked-arithmetic, Array, execution-context, named-container,
non-primitive nominal, or benchmark-specific route was reopened.

The machine-readable record is
`2026-07-30-post-nullable-compiled-current-control-reconciliation.json`.
After evidence verification, the exact 458 MiB generated-module/binary
workspace, 125 MiB focused-test cache, and generated Python cache were
removed. No matching `/var/tmp` artifact remains.

## Next

Reconcile `compiled-target-guards` against the nullable carrier.

Why: the six compiled target guards protect the small portion of the compiled
suite already delivering at least 95% of equivalent Go performance. Their
closure is invalidated by the same broad compiler hash, and protecting
existing wins has priority over selecting another miss.

What it entails: rebuild Base64, Binary Trees, JSON, Monte Carlo Pi, PiDigits,
and QuickSort strict; confirm interpreter-free graphs; audit generated
material paths for primitive nullable reach; and use the current five-run
scorecard plus existing A/B evidence when reach is absent. Reprofile only a
guard with material reach, and require repeated verifier-backed measurements
before advancing its closure.

Why it matters: restoring guard coverage ensures the general nullable
optimization did not buy improvements by silently weakening applications that
already meet the compiled goal.
