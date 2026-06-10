# Compiled Generated-Helper Execution Census

Date: 2026-07-18

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, benchmark,
fixture, or language change from this tranche.

An opt-in execution census covered all 35 selected compiled applications and
counted thirteen general generated-helper families. Every application
completed under the 55-second process cap and passed its public verifier.

The census rules out most proposed generic runtime parents: boxed binary,
boxed unary, generic interface dispatch, checked generic Array indexing,
String conversion, and generic-union fallback all recorded zero executions.
The broad residual-call counter is dominated by Option/Result Configuration,
while method/fast-call work repeats the already-closed Option/concurrency
split.

`__able_int64_from_value` was the only new exact helper with both broad and
material execution counts: 280,439 calls across eleven applications. Fresh
normal-binary profiles then rejected it before a code candidate. Across 80
verified profile launches it was below 1% flat in three unlike applications
and absent from a fourth 17.01-second profile. The surrounding CPU owners
split into HashMap lookup, regex allocation/GC, and channel scheduling.

All temporary counters were removed. No WASM work was performed.

## Census implementation and contract

The existing debug-only `--call-path-telemetry` compiler option was extended
temporarily. Normal compiler output remained unchanged because counter calls,
atomic variables, and stderr snapshots were emitted only in telemetry builds.

The temporary categories were:

- residual dynamic calls and generic method calls;
- boxed binary and unary operators;
- arbitrary Go-to-Able value conversion;
- generic interface dispatch;
- checked generic Array index get/set;
- byte-array-to-String and runtime-integer conversion;
- shared nominal instance decoding;
- the three existing fast-method/generic-union counters.

Each selected application was built once with telemetry and run once on CPU 0
with `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, its catalog executor/input
contract, canonical external `able-stdlib`, and a 55-second execution cap.
Counter-build timings are deliberately not performance evidence because every
hit performs an atomic increment. Even Binary Trees, which rose to 29.81
seconds under instrumentation, completed and verified.

The exact 35-row matrix is retained in
`2026-07-18-compiled-generated-helper-census.tsv`.

## Coverage-wide results

| Helper family | Total calls | Applications | Interpretation |
| --- | ---: | ---: | --- |
| `any_to_value` | 313,094 | 4 | 312,864 belong to Option/Result; not broad material work. |
| integer conversion | 280,439 | 11 | Only newly admissible profile target. |
| fast method call | 165,476 | 7 | 147,456 Option/Result plus the previously profiled concurrency family. |
| generic union method | 147,456 | 1 | Option/Result only. |
| residual call | 63,565 | 34 | Broad reach, but 57,649 are Option/Result and most applications make 1-3 calls. |
| nominal decode | 20,468 | 15 | 20,432 occur in four related concurrency applications. |
| method call | 18,020 | 6 | The same concurrency family as fast-method calls. |
| generic union fallback | 0 | 0 | Absent. |
| boxed binary | 0 | 0 | Static lowering avoids this helper. |
| boxed unary | 0 | 0 | Static lowering avoids this helper. |
| interface dispatch | 0 | 0 | Selected hot paths use static/compiled dispatch. |
| checked generic Array index | 0 | 0 | Selected hot paths use specialized lowering. |
| String conversion | 0 | 0 | This wrapper is not an executed shared wall. |

The residual-call breadth is therefore misleading: only Option/Result,
Mutex/Await Journal, I-Before-E, Await Channel Mux, and PiDigits exceed 1,000
calls, and the first two concrete descendant groups were already separated by
the 2026-07-15 normal-binary call-path profile gate. The census provides no new
reason to reopen that parent or the rejected fixed execution-context design.

## Integer-conversion profile gate

The leading `__able_int64_from_value` consumers were deliberately unlike:

| Application | Census calls | Shape |
| --- | ---: | --- |
| Word Frequency | 105,125 | file/text plus HashMap aggregation |
| Regex Suffix Audit | 70,676 | canonical regex NFA and text conversion |
| Channel Rollup | 42,430 | goroutine/channel pipeline plus text input |
| Future Pipeline | 32,786 | futures and channel transport |

The instrumentation was fully removed before building normal binaries. Each
application then received twenty independent whole-process CPU-profile
launches under the same memory/GC/CPU/executor contract. All 80 launches
verified and each application retained one stable output hash. Profiles were
merged only within the same executable. The exact launch ledger is retained
in `2026-07-18-compiled-int64-profile-runs.tsv`.

| Application | Merged CPU samples | Helper flat | Helper cumulative | Material owner |
| --- | ---: | ---: | ---: | --- |
| Word Frequency | 2.93 s | 0.68% | 0.68% | `__able_hash_map_find_entry`, 45.73% flat |
| Regex Suffix Audit | 17.01 s | 0% | 0% | allocation/GC and regex execution |
| Channel Rollup | 9.25 s | 0.11% | 0.11% | channel send/receive and Go runtime stack metadata |
| Future Pipeline | 4.49 s | 0.22% | 0.22% | channel send/receive and Go runtime stack metadata |

This current evidence supersedes the apparent 6.55%-8.17% cumulative helper
weight in the July 10 fixed-context/file-ingestion profiles. The helper now has
broad execution counts but not broad CPU materiality. Its static call sites
also remain limited to byte/String and character conversion plus opaque
channel/mutex handle decoding; it is not the ordinary numeric, Array, HashMap,
or compiled arithmetic path.

No optimization was implemented. Inlining another carrier switch or adding a
source/structure-specific conversion would optimize an immaterial helper and
reopen a previously closed path.

## Restoration and verification

- All temporary generated-helper variables, increments, snapshot fields, and
  audit-script columns were removed.
- The pre-existing three-field call-path telemetry format is restored.
- The focused call-path telemetry compiler test passes before and after the
  census instrumentation.
- The canonical external stdlib was unchanged.
- Raw generated trees, binaries, CPU profiles, and output captures are
  cleanup-only; only the compact census and profile-run ledgers are retained.

## Next recommendation

Run a coverage-wide generated-Go escape and bounds-check census for the
selected compiled applications, using normal source output and Go compiler
diagnostics rather than runtime instrumentation.

Why: the execution census shows that remaining compiled costs largely bypass
generic generated runtime helpers. They live in directly generated Go,
application functions, allocation, and scheduler/runtime work. Escape-analysis
and bounds-check diagnostics can identify a repeated generator-owned code
shape before another speculative runtime helper edit.

What it entails: build the selected generated sources with normalized
`-gcflags=-m=2` escape diagnostics and SSA bounds-check reporting, map findings
back to generator templates, and intersect them with current CPU/allocation
profiles. Advance only an exact primitive or shared nominal-encoding pattern
that repeats materially in at least three unlike applications. Do not add a
HashMap, Array, regex, concurrency-type, or benchmark-specific lowering, and
change canonical stdlib only for a demonstrated reusable API boundary.
Continue to defer WASM.
