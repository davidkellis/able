# Bytecode semantic-boundary reach closure

Date: 2026-07-28

## Decision

Retain no production change. The current operation-level reach map found
three primitive/consumer boundaries material in at least three unlike
families, but all three are the already-rejected generic nominal-field carrier
route or its member-write form. No open VM carrier or lowering rule reached
the candidate gate, so no prototype or timing A/B cohort was warranted.

The existing release-disabled primitive-materialization observer was already
sufficient. It records raw primitive kind, static/dynamic class, semantic
reason, materializing opcode and source, and the saved caller consumer for
static returns. An audit of direct raw-materialization helper calls found only
carrier implementation internals, shared runtime/tree-walker coercions, and
tests outside the observer's intentional bytecode-boundary wrappers. No new
instrumentation was needed.

## Measurement contract

- CPU 6, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`.
- One catalog-selected logical CPU and serial/goroutine executor.
- Canonical stdlib and source-root-only loading.
- Current scorecard sources, portable mode arguments, and public verifiers.
- Main-only counters with a 59-second process cap.
- Frozen CLI SHA-256
  `9d3559c1295b110ed95834f35cf08a040decbf4c2ea212370e47433a3becde4f`.
- Build, stats, output, and Go cache under disk-backed `/var/tmp`.

The exact boundary key is class, semantic reason, primitive suffix, and
consumer opcode. For `static_return`, consumer means the saved caller opcode;
for every other reason it is the opcode performing the materialization.
A boundary is material in an application only at 1,000 or more transitions
and at least 1% of that application's observed transitions. Candidate breadth
requires three unlike application families.

K-Nucleotide and Sudoku Masks were repeated. Both repeats had identical total
counts and byte-identical public output. K-Nucleotide's two JSON snapshots
differed only in the source ordering of two equal-count static-return rows;
operation-level aggregation was identical.

## Selected corpus

The largest current miss in each of eight unlike families was selected, except
that Concurrent Event Routing exceeded the diagnostic cap and emitted neither
output nor a snapshot. Concurrent Transform Chain, the next-largest bounded
concurrency miss, replaced it. All eight retained rows passed their public
verifiers.

| Application | Family | Transitions | Dominant exact boundary |
| --- | --- | ---: | --- |
| K-Nucleotide | text/map | 4,233,526 | `i32 -> StructLiteralNamedFast` 4,233,440 |
| Sudoku Masks | search | 18,395,575 | `i32 -> StoreSlotNew` 13,010,067 |
| Fixed Width 128 | wide numeric | 999,999 | `u64 -> StructLiteralNamedFast` 999,999 |
| Policy Record Dispatch | regex/policy | 2,072,340 | `i32 -> MemberSet` 1,045,570 |
| Distance Field | float numeric | 0 | zero-transition control |
| Discrete Event Simulation | iterator/control | 266,384 | `i64 -> StructLiteralNamedFast` 162,288 |
| Concurrent Transform Chain | concurrency | 26,711 | `i64 -> AssignName` 8,196 |
| Reverse Complement | byte output | 1 | immaterial control |
| **Total** | **8 families** | **25,994,536** | **0 dropped sites** |

Distance Field confirms that native float work does not necessarily cross an
observed runtime-value boundary. Reverse Complement similarly executes its
current byte-output workload with only one primitive materialization. These
controls prevent the high-volume integer programs from being mistaken for a
universal VM representation problem.

## Broad exact boundaries

| Boundary | Families | Transitions | Disposition |
| --- | ---: | ---: | --- |
| candidate collection `i32 -> StructLiteralNamedFast` | 4 | 4,499,649 | rejected generic nominal-field carrier |
| required interface/union `i32 -> MemberSet` | 3 | 2,263,704 | same nominal-field retention boundary |
| candidate collection `i64 -> StructLiteralNamedFast` | 3 | 234,011 | rejected generic nominal-field carrier |

The `StructLiteralNamedFast` rows span genuinely different source mechanisms:
String-derived entries, regex/NFA state, event scheduling, and concurrent
pipeline records. Their shared optimization would nevertheless be exactly the
general rule tested on 2026-07-26: retain immutable raw primitive snapshots in
ordinary nominal fields until a later dynamic consumer. That verifier-backed
five-run experiment regressed Sensor Calibration 6.93% and Rational Series
4.28% and was fully reverted.

`MemberSet` materializes a reusable raw stack cell before an aggregate write
so the stored field owns an independent stable value. Keeping that raw value
in the positional field would reopen the same generic nominal storage rule,
not establish a new primitive call/return lane. Its observer class is also
`required_dynamic/interface_union`; bypassing it broadly would weaken the
shared nominal/interface representation contract.

Exact primitive kind matters. Fixed Width 128's 999,999 `u64` struct
transitions do not create a three-family `u64` boundary, and grouping all
integers together would erase the carrier proof required by this tranche.

## Closure

No new primitive call, return, slot, stack, Array, cast, host, environment, or
public-escape boundary clears the breadth gate. The only broad shapes would
repeat a measured regression or remove required aggregate/interface
semantics. A fresh prototype would therefore violate the closed-route and
candidate-admission rules.

No compiler, generated-runtime, runtime-semantics, interpreter execution,
bytecode execution, stdlib, benchmark, language, dependency, nominal special
case, or WASM source changed. The compact machine-readable companion is
`2026-07-28-bytecode-semantic-boundary-reach-closure.json`.

## Verification and cleanup

- All eight retained census outputs passed their public verifiers.
- Focused primitive-materialization observer and CLI snapshot tests passed.
- The evidence-ledger, profile-coverage, frontier, selection, and
  mode-execution contracts passed 30 tests with one intentional
  no-invalidation skip.
- The checked ledger contains 23 current closures and zero invalidations.
- JSON, whitespace, line-count, and generated coverage reproducibility checks
  passed.
- Removed the exact 336 MiB disk-backed build/census workspace and a
  demonstrably idle 27 MiB `/tmp/able-v12-extern-go` cache. No process held
  either path open.

## Next recommendation

Run the current correctness and release-readiness gates, and pause production
performance mutation until a concrete admission invalidation exists.

Why: complete current CPU/allocation coverage and this causal boundary map now
agree that every shared owner is aggregate, required, already optimized, or
already rejected. Another local mutation would target one application or
reopen a measured regression.

What it entails: run v12 correctness, stdlib, fixture, strict dependency, and
release checks; repair only real shared-semantic failures; keep the
performance-evidence ledger current; and refresh performance only if a
retained compiler/runtime/language/stdlib change alters execution ownership.

Why it matters: this preserves interpreter-free compiled graphs and native
primitive carriers without manufacturing benchmark-specific wins. The 95%
compiled-Go and bytecode-Python/Ruby goals remain active; the pause defines
the evidence needed to resume production optimization safely. Do not begin
WASM work.
