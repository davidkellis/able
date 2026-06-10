# Truthiness and explicit-cast runtime alignment

Date: 2026-07-21

## Contract

The v12 specification makes `false`, `nil`, and every value implementing
`Error` falsy. It also defines `as` as an explicit runtime cast whose failure
raises a value implementing `Error`; a surrounding `rescue` must therefore be
able to catch failed interface checks and numeric conversions.

These rules apply equally to tree-walker evaluation, bytecode execution, and
compiled dynamic boundaries. Qualified stdlib interface identity does not
change the language-level `Error` protocol.

## Defects found

The interpreter truthiness helper checked a short `Error` impl name directly.
Canonical stdlib errors registered against
`able.core.interfaces.Error` therefore remained truthy in several exec
fixtures even though propagation and typed matching already recognized them as
errors.

The explicit-cast evaluator returned an ordinary Go error when its runtime
check failed. That surfaced the expected diagnostic at an uncaught boundary,
but skipped Able's raise signal, so a `rescue` expression could not catch it.
Specialized bytecode cast paths called the same raw conversion routine and had
the same control-flow mismatch.

## Resolution

Truthiness now delegates non-primitive values to the canonical Error matcher,
which resolves short and qualified interface identities, inherited impls, and
compiled impl tables through the existing cached interface machinery.

The public explicit-cast boundary now converts conversion failures to an Able
raise signal. Tree-walker evaluation and every bytecode cast path use that
boundary. Assignment-style coercion and typed-pattern matching retain their
non-raising APIs because they are not explicit `as` expressions.

## Verification

- Focused cast and four affected exec fixtures pass in tree-walker and bytecode.
- Two full `TestExecFixtures` runs average 19.780 seconds tree-walker
  (`19.374`, `20.186`) and 18.108 seconds bytecode (`18.220`, `17.995`), all
  below the one-minute test limit.
- The four truthiness/cast fixtures pass through the compiled fixture harness.
- Focused compiler truthiness and cast tests pass.

## Performance evidence

This is a semantic correctness change, not a performance candidate. The
checked evidence ledger now includes non-bytecode interpreter semantics as a
shared production dependency because compiled dynamic boundaries call the
same helpers.

The first revalidation tranche refreshed the compiled and bytecode target
guards with 80 verified processes and fresh matched references. Only those two
closure baselines were advanced; the other 19 remain invalidated. The ledger's
partial-advance operation refuses to overwrite a changed scope shared with
unrefreshed closures, preventing a targeted measurement from silently making
unmeasured evidence current. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-target-guard-refresh.md`.

The second tranche refreshed compiled current-control and bytecode
iterator/control with 115 verified timing processes. Exact reach evidence shows
that their ordinary mains never enter the changed non-primitive Error
truthiness fallback or explicit runtime-cast boundary. Their prior ownership
profiles therefore remain causal, no candidate was admitted, and only those
two additional closures were advanced. Seventeen closures remain invalidated.
See `../docs/perf-baselines/2026-07-21-truthiness-cast-control-closure-refresh.md`.

The third tranche refreshed compiled iterator/control and bytecode
float/numeric with 155 verified timing processes, retaining additional cohorts
for volatile workstation rows. Exact reach again finds zero changed Error
fallbacks in all nine applications and zero explicit casts in the four
bytecode applications. Option/Result reaches the compiled cast bridge 24,384
times, but two current main CPU profiles sample none of the cast path. No new
shared leaf or candidate was admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closure-refresh.md`.

The fourth tranche refreshed compiled float/numeric and bytecode wide/numeric
with 120 verified timing processes. All five compiled applications have zero
shared truthiness/cast bridge reach. The bytecode rows have zero changed Error
fallbacks; Rational and Wide make 1,000,002 and 768,003 successful casts per
main, but four current profiles put zero flat CPU in the catchable wrapper.
The remaining raw conversion and target-canonicalization work is a two-row,
previously rejected family, so no candidate was admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closure-refresh.md`.

The fifth tranche refreshed compiled wide/numeric and bytecode text/map with
115 verified timing processes, retaining second Python cohorts for volatile
Inventory and Unicode lanes. All three compiled applications have zero shared
truthiness/cast bridge reach. Two exact bytecode censuses per row reproduce up
to 63,124,394 truthiness checks but zero changed Error fallbacks and zero cast
failures; only K-Nucleotide makes 28 successful generic casts. That tiny
single-application leaf does not justify profiling or a candidate. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-wide-text-closure-refresh.md`.

The sixth tranche refreshed compiled text/map and bytecode byte/output with
115 verified timing processes, retaining matched second compiled/Go cohorts
for volatile I Before E and Word Frequency. All five compiled rows have zero
shared truthiness/cast bridge reach. Two exact bytecode censuses per row find
only four primitive checks in Reverse Complement and zero changed Error
fallbacks or explicit casts across all three applications. No path passes the
profile admission rule and no candidate was admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closure-refresh.md`.

The seventh tranche refreshed compiled byte/output and bytecode regex with 160
verified timing processes, retaining matched second compiled/Go cohorts for
Base64, FASTA, and Pi Digits. All four compiled rows have zero shared
truthiness/cast bridge reach. Two exact bytecode censuses per row reproduce
208,060-1,565,212 primitive truthiness checks but zero changed Error fallbacks
and zero explicit casts. No path passes the profile admission rule and no
candidate was admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closure-refresh.md`.

The eighth tranche refreshed compiled regex and bytecode concurrency with 450
verified timing processes, retaining matched second cohorts for every row.
Only compiled Policy reaches generated truthiness, 2,048 times; this
single-application path fails the three-application breadth rule. Two exact
bytecode censuses per row reproduce up to 541,058 primitive checks but zero
changed Error fallbacks or explicit casts. No path passes profile admission
and no candidate was admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closure-refresh.md`.

The ninth tranche refreshed compiled concurrency and bytecode register
architecture with 400 verified timing processes and balanced ten-sample means.
Compiled truthiness/cast reach is broad, but six normal-build profiles across
Await Channel Mux, Event Routing, and Mutex Work Queue sample none of those
bridges; `bridge.currentGID` remains 74.07%-96.82% cumulative. The bytecode
architecture rows reuse deterministic post-fix family censuses and have zero
changed Error-fallback or cast reach. No new concrete leaf or candidate was
admitted. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-architecture-closure-refresh.md`.

The tenth tranche refreshed the two broad architecture closures without
launching duplicate timing. The checked compiled budget now consumes the
post-fix family cohorts by fingerprint and pools every retained successful
sample. Its five unlike applications remain 7.27x-498.07x short of the
95%-of-Go budget; perfect removal of each row's largest exact owner still
leaves 3.55x-43.53x. The refreshed 85-row cross-family census adds the
corrected semantic paths as an explicit boundary, but admitted profiles sample
zero shared truthiness/cast CPU and bytecode reaches zero corrected Error
fallbacks. No candidate was admitted and only Sudoku quotient remains
invalidated. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-cross-family-architecture-ownership-audit.md`.

The eleventh tranche refreshed the final compiled Sudoku quotient closure.
Ten retained Able and ten retained Go processes average 2.073 and 0.637319
seconds. All nine generated Sudoku bodies avoid shared truthiness/cast helpers.
Two current main-only profiles merge to 3.91 seconds and put 12.53% cumulative
in signed Euclidean division, still only one-application breadth; perfect
removal would leave a 2.70x target requirement. No candidate was admitted. All
21 closures are now current with zero actionable frontier groups. See
`../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-closure-refresh.md`.
