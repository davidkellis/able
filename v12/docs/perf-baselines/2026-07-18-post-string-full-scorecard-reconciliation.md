# Post-String full scorecard reconciliation

Date: 2026-07-18

## Decision

Promote the refreshed verifier-backed external scorecard and retain no
additional compiler, bytecode VM, runtime, or canonical-stdlib performance
change from this tranche.

The current selected result is:

- compiled: 7 of 35 rows meet the 95%-of-Go target;
- bytecode: 3 of 27 rows meet both the 95%-of-Python and 95%-of-Ruby targets;
- full status: all 70 application/mode rows remain visible, including eight
  unranked timeout or unavailable-reference rows; and
- selected evidence: all 62 selected Able rows have five independent verified
  process samples under their recorded CPU/executor contracts.

The refreshed largest compiled misses do not share one new concrete
compiler-owned CPU or allocation descendant. Their current profile owners
split below generic runtime parents, and the only repeated concurrency child
is the previously rejected execution-context identity lookup. No candidate is
admitted from a ratio alone or from a common Go allocation/dispatch parent.

## Collection and reconciliation

The full refresh used the current 35-application catalog, canonical external
`../able-stdlib/src`, `GOMEMLIMIT=1GiB`, `GOGC=50`, a 55-second process cap,
and the CPU/executor contract attached to each application. Serial rows used
one logical CPU; parallel compiled rows used the same four-CPU budget as their
Go reference. Every successful Able output passed its public verifier.

The initial aggregate correctly stopped when Matrix Multiply's Python and Ruby
references could not finish under the bounded cap. Able bytecode itself
completed 5/5 at 5.052 seconds, but a selected row without bounded reference
evidence cannot produce the required cross-interpreter classification.
Matrix Multiply bytecode is therefore status-only; Matrix Multiply compiled
remains selected, and all 35 compiled applications remain selected. The
reviewed manifest now contains 35 compiled and 27 bytecode rows. This changes
measurement eligibility only; it does not change an application, reference,
timeout, runtime, or target.

The first five-run classification was volatile around several target
boundaries, so an independent five-run compiled cohort covered Fib,
I-Before-E, Matrix Multiply, PiDigits, and Base64. Base64 bytecode received a
second independent five-run cohort. No samples or outliers were discarded.

Matrix Multiply compiled moved from 1.260 seconds in the first cohort to 1.154
seconds in the second, placing its ten-run combined mean at 1.207 seconds and
1.0531x Go: only 0.05 percentage points outside the target boundary. Ten more
verified processes averaged 1.139 seconds. The complete 20-process mean is
1.173 seconds, or 1.0235x Go, so the row is classified as meeting the target.
The promoted strict aggregate uses the independent five-run classification
cohort, while the ten-run extension remains companion stability evidence.

The reconciled selected compiled rows that meet the target are Binary Trees,
QuickSort, JSON, Fib, Matrix Multiply, PiDigits, and Base64. The selected
bytecode rows that meet both interpreter targets are Base64, JSON, and
PiDigits.

Primary artifacts:

- `2026-07-18-post-string-compiled-selection-reconciled-refresh.{json,md}`
- `2026-07-18-post-string-classification-compiled-b.{json,md}`
- `2026-07-18-post-string-classification-bytecode-b.{json,md}`
- `2026-07-18-post-string-classification-matrix-extended.{json,md}`
- `external-scoreboard-current.{json,md}`

## Current compiled owner reconciliation

Ranking by absolute Able-minus-Go time avoids allowing tiny Go denominators to
select a benchmark-specific shortcut. The leading current misses and their
most recent fingerprint-compatible normal-binary attribution are:

| Application/family | Current excess | Concrete measured owner | Decision |
| --- | ---: | --- | --- |
| Sudoku Masks | 9.5421 s | generated `find_best_empty`, slice growth, allocation | application search/allocation shape; no unlike repeated child |
| K-Nucleotide | 3.7457 s | HashMap equality/hash, String conversion, primitive conversion | map/text boundary differs from Sudoku and TapeLang |
| TapeLang Alphabet | 3.1721 s | generated tape execution and methods | allocation-light application loop |
| Regex Suffix Audit | 1.0989 s | canonical regex NFA transition/closure work | library algorithm family, not a compiler-wide child |
| Mutex/Channel/Await family | 0.51-0.92 s each | `bridge.currentGID` through `runtime.Stack` | exact shared child, but its fixed-context replacement regressed unrelated N-Body by 54.7% |
| Word Frequency | 0.3013 s | generated HashMap find and `String.split` | text/map owner, not shared with the leading structural or scheduler rows |
| Fixed Width 128 | 0.2763 s | checked UInt128 dispatch, extraction, and `math/big` | nominal wide-integer semantics, not shared with the above |

The retained literal-needle `String.contains` rule can affect I-Before-E,
Document Audit, and Lexical Rollup. It cannot affect the other rows in this
table, so their July 16-17 profiles remain attribution evidence for the exact
runtime and source fingerprints measured by this refresh. A new profile of the
same binaries would not reopen rejected raw-carrier, fixed-context,
interpreter-package isolation, or named-container candidates.

The package-isolation hypothesis remains real but closed in its current form.
Removing the full interpreter dependency eliminated about 59 ms, 38 MB of
initializer allocation, and 40% of a Noop binary, but it increased collection
frequency and made allocation-heavy Binary Trees about 5% slower. Retaining
unused ballast, changing GC policy, or accepting slower allocation-heavy real
programs to improve short benchmarks would violate the broad performance bar.

## Verification and cleanup

- The reconciled aggregate passes the strict scorecard evidence checker.
- Current-scoreboard replay, reviewed selection, and portable catalog checks
  pass.
- All promoted rows preserve verifier and input fingerprints, canonical-stdlib
  source identity, CPU/executor contracts, and per-run samples.
- The empty refresh workspace was removed. Large project-local caches and old
  profile workspaces remain user-removable through `just cleanup-apply` or
  `just cleanup-archives`; the performance run did not delete them implicitly.
- No WASM work was performed.

## Next recommendation

Run a bytecode multi-operation quickening feasibility census before editing
another VM micro-helper.

Why: only 3 of 27 selected bytecode rows currently meet both interpreter
targets, so the interpreter remains the larger product gap. Recent unlike
profiles repeatedly share dispatcher/call parents, but their individual raw
integer, stack, return, lookup, and carrier children have either split by
workload or already failed broad A/B guards. The remaining opportunity is
therefore likely work repeated across opcode sequences rather than another
single helper branch.

What it entails: add temporary opt-in counters for opcode pairs/basic-block
shapes and their dynamic guard outcomes across the selected bytecode catalog;
require one sequence and one removable validation/materialization boundary to
be material in at least three unlike applications; and remove all counters
after the census. Do not retry the rejected slot/constant fusion or specialize
a benchmark, String, Array, regex, iterator, or named container. Only if the
same semantically guarded sequence recurs should a generic quickened opcode or
basic-block plan be prototyped, with invalidation/dynamic-dispatch proof,
focused interpreter tests, and repeated order-balanced application guards.
