# Performance Competitiveness: Active Selection Policy

## Goals

Able performance work serves two measurable outcomes while preserving the v12
language contract:

- **Compiled Able:** each rankable, statically representable application should
  reach at least 95% of the equivalent Go application's throughput.
- **Bytecode Able:** each rankable application should reach at least 95% of
  both the equivalent Python and Ruby application's throughput where both
  references exist. A missing reference or timeout is unranked, never a pass.

The Go tree-walker remains the behavioral reference. The compiler, bytecode
VM, external canonical `able-stdlib`, fixtures, and benchmark implementations
must preserve the same observable Able semantics. Faster benchmark code that
changes those semantics is not progress.

## Current scorecard and selection status

The reviewed selected frontier contains 49 compiled applications and 42
bytecode applications. Seven additional bytecode rows retain bounded status
probes. The 2026-07-22 current scorecards record exactly five verifier-backed
Able and applicable reference samples for every selected row; excluded rows
retain bounded status rather than manufactured ratios. Records:
`docs/perf-baselines/2026-07-22-current-compiled-scorecard.json`,
`docs/perf-baselines/2026-07-22-current-bytecode-scorecard.json`, and
`docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json`.

| Mode | Rankable rows meeting target | Current conclusion |
| --- | ---: | --- |
| compiled vs Go | 5 / 49 current; 3 established | Far from the 95% target; Base64 and Monte Carlo Pi are volatile snapshot crossings, while Matrix Multiply now clearly misses |
| bytecode vs Python and Ruby | 3 / 42 current; 2 established | Far from the 95% target; Await Channel Mux is a volatile snapshot crossing, while Base64 now clearly misses |

The current compiled meets are Binary Trees, QuickSort, Base64, JSON, and
Monte Carlo Pi. The current bytecode meets are Await Channel Mux, JSON, and
PiDigits.
Independent reconciliation establishes Binary Trees, QuickSort, and JSON as
compiled guards and JSON/PiDigits as bytecode guards. Compiled Base64, Monte
Carlo Pi, and bytecode Await Channel Mux are variance-sensitive snapshot
crossings. Record:
`docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md`.
The schema-2 frontier now enforces this distinction through
`bench-performance-stability.json`: every selected snapshot meet must carry a
reviewed cross-cohort classification, pooled/cohort ratios, sample counts,
current Able/reference fingerprints, stdlib identity, and hashed evidence.
Record:
`docs/perf-baselines/2026-07-20-performance-stability-manifest.md`.
The follow-up source-exact refresh confirms all five established guards—Binary
Trees, QuickSort, compiled JSON, bytecode JSON, and bytecode PiDigits—in two
independent cohorts against the promoted stdlib tree. Record:
`docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md`.
The complete evidence-backed frontier reports 83 misses, 156.334 seconds
above the aggregate per-row budget, and zero unclosed implementation groups.
The refresh therefore does not authorize another cache, frame, map, raw-cell,
typed-lane, scheduler, regex, or named-container experiment. Reopen a group
only when changed semantics or new exact profiles invalidate its recorded
closure.

The subsequent cross-family architecture audit reaches the same conclusion at
the next level up. Compiled hot paths largely execute direct generated Go, so
shared generated-runtime helpers are not the common residual wall. Bytecode
dispatch, scalar encoding, call/return, and nominal-allocation costs recur, but
their general mechanisms have already failed broad wall-time, reach,
deployability, or lifetime gates. Go map, allocation, and GC parents divide
into different semantic owners. The machine-readable eight-boundary census
admits zero new candidates and changes no frontier disposition. Record:
`docs/perf-baselines/2026-07-20-cross-family-architecture-ownership-audit.md`.

The checked compiled target-budget follow-up makes that architecture result
quantitative. Five unlike high-excess applications require 7.54x-640.06x
speedups to reach 95% of Go. An intentionally favorable model makes each
application's largest attributed exact owner completely free, yet still leaves
4.28x-55.94x required. String and unsigned bridge conversion are material in
only two families; escaping nominal results, regex storage, and goroutine
identity are single-family owners. Allocation/GC and direct generated bodies
remain aggregate parents over different semantics. No current compiler
architecture mechanism is eligible. Record:
`docs/perf-baselines/2026-07-21-compiled-architecture-target-budget.md`.

The 2026-07-22 cross-engine refresh supersedes those timing ranges while
preserving the decision. The compiled frontier contributes 23.368 seconds of
target excess and bytecode contributes 132.966 seconds, or 85.05% of the
total. Across five unlike compiled applications, current required gains range
from 6.42x to 492.57x; making each largest attributed exact owner free still
leaves at least 3.60x. Across six unlike bytecode applications, making every
stack-transport operation free at equal average cost yields at most 1.80x and
still leaves at least 7.79x. No local compiler or VM mechanism is admitted.
The proposed semantic-region follow-up is now reconciled against completed
executable work. A coverage-wide typed safe-region census already found five
unlike material families; its one generic out-of-line executor regressed all
three governing applications. Whole-function register execution and
cross-suite Go PGO also failed their broad gates, and the current six-family
semantic audit admits no exact operation family. A uniform-instruction-cost
sizing model makes every previously admitted typed region free and closes only
Monte Carlo Pi; the other three current rows retain modeled 2.25x-21.58x gaps.
Because the census measured instruction rather than wall-time share, this is
not claimed as an upper bound; the repeated executable regressions close a
second Go dispatcher or executor.
Records:
`docs/perf-baselines/2026-07-22-current-cross-engine-architecture-target-budget-reconciliation.md`
and
`docs/perf-baselines/2026-07-22-bytecode-semantic-region-tier-feasibility.md`.

The non-WASM native hot-code design and budget is now complete. Its safe first
tier is pointer-free leaf code: Go owns every boxed or identity-bearing root,
and allocation, calls, dynamic/nominal operations, errors, unwinding, externs,
and suspension side-exit to the exact ordinary-VM instruction. Backedges poll
for cooperative scheduling; unsupported platforms and evicted code use the
ordinary VM; the process-local cache is content-addressed, W^X, and bounded.

Current compiled Able provides the only complete semantics-preserving native
planning proxy. Substituting its time for all 42 selected bytecode applications
removes 88.68% of target excess, but only 11 rows meet and 31 still miss. The
known typed regions clear a 25% equal-cost target-excess-reduction gate only in
Monte Carlo Pi; RMS Norm and Fixed Width 128 lack sufficient reach, and Future
Await Race misses even at whole-application compiled-proxy speed. No backend or
prototype is admitted. The next lane is an opt-in deterministic per-function
reach census. It must attribute entries, dynamic instructions, primitive work,
effect exits, backedges, boxed boundaries, and captures to source functions in
unlike applications before a backend ADR. Record:
`docs/perf-baselines/2026-07-22-bytecode-native-hot-tier-design-budget.md`.

This is an architecture boundary, not permission for named nominal, container,
benchmark, or application special cases.

The first feature-interaction tranche adds Concurrent Text Index and Validated
Job Pipeline. A reproducible 55-pair matrix across 11 discriminating families
shows that they reduce empty interactions from 29 to 15 and improve 32 pairs.
Five-run verifier-backed baselines remain far outside both product targets.
Their exact compiled profiles reproduce the rejected goroutine-identity
boundary; exact bytecode-main profiles reproduce completed member-cache,
return/type-match, allocation, and GC owners. No performance code is admitted.
Record:
`docs/perf-baselines/2026-07-20-feature-interaction-application-gate.md`.

The subsequent Policy Record Dispatch tranche strengthens 45 interaction
pairs and reduces depth-one coverage from nine pairs to one without admitting
performance code. The final reconciliation recognizes Dependency Wave
Validation's already-hot typed Channel and payload-union patterns, raising the
minimum depth across all 55 portable pairs from one to two and eliminating the
last depth-one pair. It adds no application or implementation code. Records:
`docs/perf-baselines/2026-07-21-feature-interaction-coverage-depth-gate.md` and
`docs/perf-baselines/2026-07-21-concurrency-lexical-depth-reconciliation.md`.

Until new evidence invalidates an implementation closure, audit portable
three-family interactions, weighted by semantic importance and current target
excess. Inspect existing applications first; add one bounded source-equivalent
application only for a material missing combination, and profile code only if
the audit exposes a new concrete generic leaf in at least three unlike
applications.

That audit now records all 165 triples as nonempty. Recognizing Concurrent Text
Index's repeated nullable/pattern/control paths reduces depth-one triples from
eight to two; both residual callable/concurrency triples have substantial
Concurrent Event Routing coverage and do not justify a duplicate application.
Coverage expansion is therefore paused. Next build a cross-mode residual cost
model across unlike high-excess applications, using generated-code shape,
semantic-operation/opcode counts, allocations, and reference work to select an
architectural compiler or VM mechanism only when it explains material excess
in at least three programs. Record:
`docs/perf-baselines/2026-07-21-weighted-feature-interaction-triple-audit.md`.

The subsequent application-depth passes add Concurrent Document Pipeline and
Manifest Normalization. The latter raises files/text × captured callables ×
Option/Result to three unlike applications and makes compiled
`String.to_builtin` CPU-material in a third program. That trigger admitted one
generic indexed-error-formatting experiment, but its repeated wall gate was
mixed: roughly 1.5% faster K-Nucleotide, 2.9% slower Policy Record Dispatch,
and 1% slower Manifest Normalization. The candidate was reverted. Bytecode
reproduces only completed call, return, raw-integer, slot/member, map, and GC
families. The 89-row frontier remains zero actionable. Records:
`docs/perf-baselines/2026-07-21-concurrent-document-pipeline-application-gate.md`
and `docs/perf-baselines/2026-07-21-manifest-normalization-application-gate.md`.

The Validated Job Pipeline file-entry pass then raises minimum portable
three-family depth from two to three by making real file text and program-entry
arguments drive 2,048 concurrent validation tasks. Its compiled profile again
places 94.16% cumulative CPU below `bridge.currentGID` / `runtime.Stack`; its
bytecode profile has no new material shared VM child. The tempting follow-up—a
spawn-selected context ABI—was already completed and rejected after a 10.0%
Mutex Ledger regression, so the stale recommendation was reconciled without
another timing run. Records:
`docs/perf-baselines/2026-07-21-validated-job-file-entry-application-gate.md`
and
`docs/perf-baselines/2026-07-22-compiled-execution-context-recommendation-reconciliation.md`.

The scorecard's bytecode column is deliberately full-process: normal CLI load,
lowering, typechecking, bootstrap, and one `main()` invocation all remain in
the row. It is not interchangeable with the warmed `bytecode-runtime` lane,
which validates once and measures repeated `main()` execution. Trusted
`able run --skip-typecheck` is available only after a separate successful
check and is not enabled for the scorecard. The CPU-8-pinned, three-process
Sudoku/Word-Frequency/Future-Pipeline comparison found the same Sudoku caps,
1.33 s versus 1.32 s for Word Frequency, and 0.41 s in both Future Pipeline
lanes. Typechecking and loading therefore are not the material common wall;
the execution-only lane stays separate and must not silently replace the
full-process target. Record:
`docs/perf-baselines/2026-07-15-bytecode-execution-lane-reconciliation.md`.

## Evidence invalidation triggers

A closed performance mechanism is reconsidered only when at least one checked
trigger applies:

1. the v12 spec changes the relevant semantic boundary;
2. canonical `able-stdlib` adds or changes a reusable specified operation at
   that boundary;
3. compiler, runtime, bridge, or VM implementation changes invalidate the
   source/artifact identities used by the causal evidence;
4. a new source-equivalent portable application makes the same concrete leaf
   CPU-material in a third unlike family; or
5. current verifier-backed scorecard evidence contradicts the closure's
   predicted reach or guard behavior.

Timing volatility, aggregate GC/allocation recurrence, ancestry labels, or an
old “next recommendation” are not invalidation. When no trigger applies, do
not rerun unchanged cohorts or construct a local candidate. The next selection
tool should make these identities and source scopes mechanical so the roadmap
cannot silently reopen current evidence.

That selector is now checked in. It covers 18 frontier groups plus the
cross-family ownership, bytecode-register architecture, and compiled target-
budget closures. All 21 are current and none is selected. Production scope
drift is mode-aware, evidence and benchmark identities are closure-local, and
test-only changes are excluded. `just bench-evidence-ledger-check` runs its
contract tests and exact artifact check; `just bench-scoreboard-check` includes
the non-mutating ledger check. Record:
`docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.md`.

## Non-negotiable implementation rules

- `spec/full_spec_v12.md` is semantic authority; the shared AST and Go
  tree-walker are the parity baseline.
- Static compiled paths use direct Go carriers, control, and dispatch. Dynamic
  `runtime.Value` paths are explicit language/host boundaries only.
- Only primitive types may have primitive-specific compiler lowering. Every
  non-primitive nominal type, including stdlib and user containers, uses shared
  nominal/carrier/dispatch lowering.
- Array and String may have VM/compiler treatment only as language/kernel
  boundaries, never because one benchmark uses them.
- `able-stdlib` changes require a reusable specified API or behavior gap; they
  may not be benchmark shims. The external repository is canonical.
- Do not start WASM work until these compiled and bytecode targets are met.

## Candidate admission gate

An implementation candidate is admissible only when all of the following are
true:

1. A material semantic/compiler change or new spec-defined portable
   application makes an investigation necessary.
2. A bounded profile identifies the same concrete, non-nominal material leaf
   in at least three unlike verifier-backed applications.
3. The proposed change belongs to shared runtime, VM, compiler, or stdlib
   machinery—not a benchmark, source shape, named container, algorithm, task
   count, or host-input special case.
4. Focused semantic, generated-source/boundary, and fallback/invalidation tests
   prove the changed contract.
5. The complete bounded coverage/performance gate shows no material regression
   outside the original workloads.

If the leaf is only a dispatcher/envelope parent, or its material children
diverge by application, retain no code. Refresh evidence or complete the next
language/stdlib requirement instead.

## Measurement and verification

Use verifier-backed application launches with fresh matched references, one
process per sample, `GOMEMLIMIT=1GiB`, `GOGC=50`, the configured timeout, and
the catalog's mode-aware execution contract. `--cpu-affinity` supplies an
ordered pool; each row records its logical CPU budget, resolved taskset,
executor policy, and matching `GOMAXPROCS` for Go/Able processes. Preserve
source verifiers and foreign references; do not tune an Able program or
reference merely to improve a ratio.

`just bench-scorecard-refresh` creates bounded grouped evidence for all 49
portable `coverage` applications and promotes only after the aggregate passes
the exact five-run selected-evidence check. `just bench-scoreboard-check`
validates the checked-in scoreboard without executing workloads. Excluded
bytecode rows remain visible as bounded status probes; full-scorecard reruns
follow material changes, not idle time.

Every kept change needs the proportionate fast test group. The full matrix is
confidence evidence, not the default sub-minute test. Performance evidence is
valid only when the same source, verifier, references, and guard settings are
recorded with the result.

## Definition of done

The performance program is complete when the full rankable scorecard reaches
the two 95% targets, timeouts are either resolved or documented as language
limitations, and source/boundary audits show that static compiled paths and
optimized VM paths retain their intended representations. New benchmark
families need an Able implementation or an explicit spec/stdlib blocker.

Until then, the next work is chosen exclusively by the candidate admission
gate. Historical benchmark readings, kept optimizations, rejected experiments,
and former “next slices” are retained in
[the historical performance record](./performance-competitiveness-historical.md)
and dated `PLAN.md`/`v12/LOG.md` entries; they are not current assignments.
