# Full 32-application scorecard decision — 2026-07-15

## Measurement

`bench_refresh_external_scorecard` completed the first source-aligned full
portable cohort after CPU 14 passed the immediate quiet-host preflight. The
run rebuilt fresh Go, Python, and Ruby reference reports, then measured every
compiled and bytecode application with three requested processes, a 45-second
cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

The retained aggregate is
`2026-07-15-full-coverage-scorecard-refresh.{json,md}`. It joins 24 source
scorecards into 64 application/mode rows for all 32 portable applications.

## Result

| Mode | Rankable | Meets 95% target | Unranked |
| --- | ---: | ---: | ---: |
| Compiled vs Go | 31 | 4 | 1 timeout |
| Bytecode vs Python and Ruby | 23 | 3 | 9 |

The nine bytecode unranked rows are seven timeouts, one incomplete result, and
one row without both matched foreign references. The incomplete K-Nucleotide
bytecode row has two verified samples plus one timeout. It remains visible but
cannot be used to claim either a pass or a ratio-derived failure.

The final aggregation exposed this valid mixed outcome: the scoreboard had
previously rejected any row containing both successes and timeouts. The
generic report contract now labels such rows `incomplete` and makes them
`unranked`; it does not alter VM, compiler, runtime, stdlib, or benchmark
behavior.

## Decision

Keep no performance implementation change. The full scorecard establishes a
larger, reproducible selection baseline, but ratios alone do not identify an
inner optimization. In particular, do not respond to Fixed Width, Rational,
map, regex, Array, or scheduler rows with a named-type, API, or benchmark
special case.

## Historical reconciliation and next eligible direction

The initially proposed Fixed Width 128/Rational Series/K-Nucleotide profile
triad is not new evidence. The existing paired and generated-main profiles
already split it into UInt128 checked-member work, generic call/frame work,
and map/conversion/GC work; the shared parents are not a common removable
leaf. The existing raw-value and safe-Go carrier audits also reject turning
those named routes into a `RawValue`, compiler-only carrier, or nominal
container shortcut.

No new profile or implementation is authorized from this unchanged scorecard
data. Keep the scorecard as the broad selection guard. The next performance
investigation may begin only when a material language-wide compiler/runtime or
semantic-portability change creates a concrete non-nominal leaf in at least
three unlike verifier-backed applications. Until then, the next eligible work
is an unfinished semantic or portability roadmap boundary with fixture parity
first. The evidence and correction are recorded in
`2026-07-15-full-coverage-scorecard-historical-reconciliation.md`.
