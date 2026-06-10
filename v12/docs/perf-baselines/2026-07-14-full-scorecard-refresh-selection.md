# Full verifier-backed scorecard refresh selection

## Scope and method

This refresh reran the complete current scorecard scope in short independent
groups so the harness runner could retain each verifier-backed result rather
than losing one large report at its wall-clock boundary.  It includes:

- the 16-program generality suite in compiled and bytecode modes;
- the six standalone channel, Future, await, and Mutex applications in both
  modes; and
- fresh Go references for compiled mode plus fresh Python/Ruby references for
  bytecode mode.

Every launch used CPU 15 with `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, one timed process, a 45-second execution cap, and the
canonical Ruby verifier when one exists.  The detailed, machine-readable
results are the dated inputs listed by
`2026-07-14-full-scorecard-refresh.md`; the combined report deliberately does
not overwrite the separately maintained current scoreboard.

This is a one-run selection screen.  It identifies material, verified current
misses but is not a claim of a statistically significant regression relative
to the previous three-run report.

## Result

The combined report has 4 of 21 rankable compiled rows at or better than the
95%-of-Go target and 2 of 14 rankable bytecode rows at or better than both
95%-of-Python and 95%-of-Ruby targets.  Timed-out or reference-unavailable
rows remain unranked; none are treated as passes.

The six asynchronous applications are all material misses in both modes:

- compiled is 31.58x to 257.69x the fresh Go time; and
- bytecode is 2.48x to 19.33x the fresh Python time.

That is broad status evidence, not by itself a source-change authorization.
The completed concurrent-profile gates already cover the required unlike
shapes: channel rollup, Future pipeline/race, channel Awaitable selection, and
both public Mutex acquisition styles.  They establish that bytecode shares
only VM/executor parent frames before diverging into text, numeric, channel,
or mutex children.  Compiled profiles instead repeat the generic
`bridge.currentGID` / `runtime.Stack` identity bridge wall.  The only known
general remedy, the fixed execution-context ABI, remains opt-in because its
broad default gate regressed N-body and its package-linkage refinement
regressed K-Nucleotide.  Reopening either experiment or specializing a
scheduler, channel, Future, Mutex, task count, or benchmark shape would not
meet the broad-performance rule.

The large generality misses likewise confirm already separated routes:
K-Nucleotide map/counting and Reverse Complement tracked-byte work do not
share a concrete VM descendant; numeric, recursive-search, and text cases
remain distinct.  No canonical `able-stdlib` change is selected.

## Decision and next gate

Keep no VM, compiler, bridge, runtime, stdlib, or benchmark source change.
Do not rerun this unchanged scorecard as a substitute for a candidate.  The
next performance tranche is eligible only after a material cross-cutting
semantic/compiler change, or after a newly needed spec-defined portable
application exposes a concrete descendant repeated in at least three unlike
verified applications.  Profile that descendant with current bounded controls
and require broad compiled and bytecode guards before keeping any change.
