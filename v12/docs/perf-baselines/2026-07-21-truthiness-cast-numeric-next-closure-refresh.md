# Truthiness/cast numeric-next closure refresh

Date: 2026-07-21

## Decision

The `compiled-float-numeric` and `bytecode-wide-numeric` closures are current
against the post-truthiness/cast shared interpreter semantics. Keep no
compiler, VM, runtime, canonical-stdlib, benchmark, reference, or WASM change
from this tranche.

Fresh repeated timing confirms the product gaps but does not identify a new
generic candidate. All five compiled applications have zero reach into either
changed semantic path. The three bytecode applications have zero changed Error
fallbacks. Rational Series and Wide Integer Records reach the explicit-cast
boundary heavily, but make zero failing casts; four current main profiles put
zero flat CPU in the new catchable wrapper. Their measurable descendants are
the already-closed raw conversion and primitive-target canonicalization work,
and recur in only two selected programs.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every timed process passed its public
  verifier and used its catalog directory, arguments, CPU budget, and a
  55-second cap.
- Arithmetic means retain every successful sample. Distance Field and N-body
  received matched second compiled/Go cohorts after volatile first lanes.
  Fixed Width Python received a second reference cohort. No sample was removed
  or replaced.
- Able executions use `GOMEMLIMIT=1GiB`, `GOGC=50`, and the catalog-resolved
  serial CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Distance Field | compiled | 10 | 0.106 s | Go | 10 | 0.013843 s | 7.657x |
| Mandelbrot | compiled | 5 | 0.122 s | Go | 5 | 0.051568 s | 2.366x |
| Monte Carlo Pi | compiled | 5 | 0.206 s | Go | 5 | 0.216336 s | 0.952x |
| N-body | compiled | 10 | 0.182 s | Go | 10 | 0.042016 s | 4.332x |
| RMS Norm | compiled | 5 | 0.118 s | Go | 5 | 0.014422 s | 8.182x |
| Fixed Width 128 | bytecode | 5 | 9.704 s | Python | 10 | 0.510847 s | 18.996x |
| Rational Series | bytecode | 5 | 4.636 s | Ruby | 5 | 0.177608 s | 26.102x |
| Wide Integer Records | bytecode | 5 | 5.814 s | Python | 5 | 0.074452 s | 78.091x |

All 120 timing processes represented by these pooled decisions verified with
zero failures and zero timeouts: 70 compiled/Go processes and 50
bytecode/Python/Ruby processes. The additional cohorts are averaged with the
first cohorts. The remaining workstation variance is retained explicitly:
Distance Field Able has 33.04% pooled CV, N-body Go 19.11%, and Fixed Width
Python 26.00%.

Monte Carlo currently meets the 95%-of-Go target at 0.952x. As in the existing
threshold-stability record, it remains a volatile crossing rather than an
established guard; this focused cohort does not rewrite the full scoreboard.

## Exact reach

Temporary opt-in counters were placed immediately at the changed semantic
boundaries, used only in untimed processes, and then removed.

| Application | Mode | Census processes | Truthy checks/process | Changed Error fallback | Explicit casts/process | Cast failures |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Distance Field | compiled | 1 | 0 | 0 | 0 | n/a |
| Mandelbrot | compiled | 1 | 0 | 0 | 0 | n/a |
| Monte Carlo Pi | compiled | 1 | 0 | 0 | 0 | n/a |
| N-body | compiled | 1 | 0 | 0 | 0 | n/a |
| RMS Norm | compiled | 1 | 0 | 0 | 0 | n/a |
| Fixed Width 128 | bytecode | 2 | 2,000,000 | 0 | 0 | 0 |
| Rational Series | bytecode | 2 | 650,002 | 0 | 1,000,002 | 0 |
| Wide Integer Records | bytecode | 2 | 71,996 | 0 | 768,003 | 0 |

Every bytecode count reproduced exactly. All bytecode truthiness calls are
primitive checks and return before the corrected Error matcher. The compiled
telemetry rows also show only 10 ordinary dynamic-boundary events in total;
none enters generated `__able_truthy` or `__able_cast`.

## Reached bytecode profile gate

The uninstrumented current interpreter test binary has SHA-256
`8cd3807f7d42c4ea3905398a9f2825e50986be62d138fe081d994989caae8749`.
Two bounded main-only CPU processes were collected for each reached row after
load, typecheck, lowering, warmup, and forced GC.

| Application | Merged samples | Cast opcode cumulative | Explicit wrapper flat | Raw cast cumulative | Target canonicalization cumulative |
| --- | ---: | ---: | ---: | ---: | ---: |
| Rational Series | 8.62 s | 10.21% | 0% | 2.78% | 2.09% |
| Wide Integer Records | 12.13 s | 5.11% | 0% | 2.31% | 0.74% |

`castValueToType` has zero flat samples in all four profiles; its cumulative
share is inherited from the raw conversion below it. Because every measured
cast succeeds, the new error-to-raise branch is not entered. The remaining
raw conversion and cast-target work is material in two, not three, selected
applications. Moreover, the general reserved-primitive target bypass already
passed semantics and allocation gates before reversing sign across complete
and CPU-0 broad wall cohorts; this refresh supplies recurrence, not a new fact
that reopens it.

The concise census is retained in
`2026-07-21-truthiness-cast-numeric-next-closure-reach.json`; raw compiled
telemetry is retained in
`2026-07-21-truthiness-cast-numeric-next-closure-compiled-reach.json`. No
diagnostic counter, binary, raw profile, or generated package remains in
production code.

## Exact timing artifacts

- Initial compiled Able and Go reference reports:
  `2026-07-21-truthiness-cast-numeric-next-closures-compiled.json`
  (`c53474dd6f2e2c8faaae3868e4a3170e883109d974f5da4fe3ab2d6004c8de95`),
  `2026-07-21-truthiness-cast-numeric-next-closures-go-reference.json`
  (`2a4ee4dcd99df2c95da9b7b27e012fa044d3c636117e9c4d2287b9759cc805e6`).
- Distance/N-body second Able and Go reports:
  `2026-07-21-truthiness-cast-numeric-next-closures-compiled-c2.json`
  (`e489825222985389b394a71179aa246d519e67ddcd55d08f9d3aa5f9af368858`),
  `2026-07-21-truthiness-cast-numeric-next-closures-compiled-c2-go-reference.json`
  (`f958fbdb011842587486b6ec47b2e612cd2e0dc80deb3a89e13cf711e6cdcec1`).
- Bytecode Able and initial interpreter references:
  `2026-07-21-truthiness-cast-numeric-next-closures-bytecode.json`
  (`34abe088c348247ed0e97e5d4eeb89945b82a14c43176c8276033304365ea85b`),
  `2026-07-21-truthiness-cast-numeric-next-closures-interpreter-reference.json`
  (`44c295b7dd327bf7e1a8a908277c24622c17205438df6c4763b9b927893f7516`).
- Fixed Width second Python report:
  `2026-07-21-truthiness-cast-numeric-next-closures-fixed-python-c2-reference.json`
  (`ab17e52ca60f35fbda837a540122f9678c891ac66ea8d4401a7fdcec0b1a0a57`).

## Next recommendation

Refresh `compiled-wide-numeric` and `bytecode-text-map` next.

Why: compiled wide-numeric is the nearest remaining closure to the heavily
reached cast/type-matching machinery measured here, but exercises the compiled
nominal lowering side rather than the VM. Bytecode text/map is the nearest
unrefreshed dynamic-protocol family where canonical Error truthiness could
plausibly occur through lookup, iterator, and result-bearing paths. Together
they test both remaining causal directions without reopening closed aggregate
allocation or adding named-container special cases.

What it entails: reuse the frozen sources and current reference toolchains;
collect five verifier-backed processes per ordinary lane; retain additional
cohorts for volatile workstation rows; run exact main-only truthiness/cast
reach before profiling; and profile only materially reached changed paths.
Advance only those two closures. Build a candidate only if one concrete,
generic leaf is material in at least three unlike applications and preserves
the current compiled and bytecode target guards. No WASM work is involved.
