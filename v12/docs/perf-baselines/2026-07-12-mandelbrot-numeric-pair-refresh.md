# Mandelbrot numeric-pair refresh (2026-07-12)

## Method

Mandelbrot refreshed Go 1.26.4, Ruby 4.0.5, Python 3.14.5, compiled Able, and
bytecode Able with three CPU-2-pinned, verifier-backed processes each. The
lane used `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap.

| Mode | Able (s) | Go ratio | Ruby ratio | Python ratio | Status |
| --- | ---: | ---: | ---: | ---: | --- |
| compiled | 0.1567 | 3.16x | 0.08x | 0.13x | verified 3/3 |
| bytecode | 7.4833 | 150.87x | 3.91x | 5.98x | verified 3/3 |

Fresh references are Go `0.0496s`, Ruby `1.9137s`, and Python `1.2517s`.
Together with the immediately preceding fresh Monte Carlo Pi row (bytecode
`2.6400s`, 1.62x Ruby and 1.68x Python), this confirms two independent
numeric bytecode misses.

## Profile gate

No new profile was taken. The bytecode VM source has not changed since the
current paired numeric audit in
`2026-07-11-bytecode-numeric-reference-gate.md`. That audit already captures
the same concrete state: the only material shared leaf is
`execJumpIfFloatMulAddMulCompareConstFalse(...)`; Mandelbrot additionally
spends material time in normalized per-pixel stores/general binary work, while
Monte Carlo Pi spends it in cast/divide stores and the integer PRNG recurrence.
The shared float compare/store and raw-float representation variants were
tested broadly and were neutral or regressive on their guards.

## Decision

Keep no VM, compiler, tree-walker, or `able-stdlib` change. Do not reopen the
float compare/store, raw-float carrier, or quickened-plan-table variants from
fresh ratios alone. The scorecard validates the numeric gap, but its concrete
shared leaf remains an exhausted lane rather than new evidence.

## Next recommendation

Return to a distinct bytecode boundary: refresh the file/byte-transform family
with Reverse Complement and K-Nucleotide, retaining Base64 or JSON as a
control. Why: the numeric family has now repeated and rejected its only shared
leaf, while file/byte inputs can reveal a separate generic materialization or
call/return cost. The work entails fresh three-family rows first; profile only
if two completed misses share a non-nominal helper and remain neutral on the
control.
