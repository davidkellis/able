# Able v12 Roadmap

This file is the forward-looking roadmap and current handoff. Completed work
belongs in `LOG.md`, `v12/LOG.md`, and the dated records under
`v12/docs/perf-baselines/`; it should not accumulate here.

## Session Start

1. Read `AGENTS.md`, this file, and `spec/full_spec_v12.md` completely.
2. Read the latest entries in `LOG.md` and `v12/LOG.md`, plus every record
   linked from the current handoff.
3. Inspect the dirty worktree before editing. Preserve all existing changes and
   untracked files.
4. Work only in v12 unless a maintainer explicitly requests an archival v10 or
   v11 fix.
5. Follow the authoritative performance frontier below. Fix an existing
   correctness failure before continuing performance work.

## Mission

- Compiled Able applications must reach at least 95% of equivalent Go
  performance and should exceed Go where practical.
- Bytecode Able applications must reach at least 95% of equivalent Python and
  Ruby performance and should exceed both where comparable.
- Use broad, feature-covering benchmark applications. Retain only general
  compiler, runtime, VM, or canonical-stdlib rules that help unlike programs.
- Primitive Able types and static Arrays use native Go carriers. Non-primitive
  nominal types use shared nominal translation and semantic encoding; never
  add named-container or other non-primitive nominal compiler rules.
- Update `../able-stdlib` only for a general stdlib correction or optimization.
  Keep `v12/stdlib-deprecated-do-not-use` read-only.
- Do not begin WASM work until compiled and bytecode performance goals are met.

The architectural overview is
`v12/design/performance-competitiveness-vision.md`. The combined compiled and
bytecode comparison is
`v12/docs/perf-baselines/external-scoreboard-current.{json,md}`; the current
strict compiled refresh is
`v12/docs/perf-baselines/2026-07-28-current-strict-compiled-scorecard-owner-closure.{json,md}`.

## Current Handoff

Start from:

- `v12/docs/perf-baselines/2026-07-28-post-consolidation-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-28-v12-correctness-release-readiness-closure.md`
- `v12/docs/perf-baselines/2026-07-28-compiled-semantic-work-recommendation-reconciliation.md`
- `v12/docs/perf-baselines/2026-07-28-current-strict-compiled-scorecard-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-28-compiled-await-post-materialization-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-27-typed-generic-nominal-storage-reach-census.md`
- `v12/docs/perf-baselines/2026-07-27-compiler-positional-boundary-zero-reach-closure.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.md`
- `v12/docs/perf-baselines/2026-07-25-compiled-scorecard-generic-storage-closure.md`
- `v12/docs/perf-baselines/2026-07-26-compiled-aot-native-carrier-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological retained state is in `LOG.md` and `v12/LOG.md`. Strict
fallback-free applications must remain interpreter-free, and the existing
native-carrier, nominal-encoding, and dynamic-boundary guards remain
authoritative.

### Authoritative performance frontier

There is no active production optimization candidate. The current portable
catalog has 63 applications and the reviewed selection has 119 rows:
63 compiled and 56 bytecode. The current compiled refresh has 9 target passes
and 54 misses, a 5.2579x geometric-mean Able/Go ratio, and 5.4254 seconds of
positive target excess. All 63 strict compiled graphs remain interpreter-free.
Detailed completed work is in `LOG.md`, `v12/LOG.md`, and the dated records
linked above.

The broad retained-state owner refresh retained no production change.
Fifteen CPU profiles and three exact allocation profiles for each of ten
poorly explained compiled misses found no open compiler, generated-runtime, or
semantic-boundary owner material in at least three unlike applications.
Repeated leaves are already-closed checked arithmetic or boxing, same-family
UTF-8/regex work, required generic map/nominal/concurrency semantics, or
already-native application and Go runtime work.

The proposed TapeLang/Sudoku/N-Body semantic-work follow-up was already
completed on 2026-07-26 and remains valid. Fresh current strict artifacts
retain the normalized source identities and interpreter-free graphs. Current
assembly preserves the earlier hot call/branch counts, and the relational
observer reproduces 0/13 safe reachable checks for TapeLang, 4/32 for Sudoku,
and 32/52 for N-Body. N-Body remains the only dynamically material proof
family, so no production candidate reaches three unlike applications.

The current correctness and release-readiness tranche is complete. Repository
vet/build, the ordinary handoff, the separate compiler release matrix, the
cold and warm generated-Go CLI lanes, canonical stdlib parity, and all
deterministic evidence contracts pass. The compiler matrix includes 120 green
fallback, execution, strict-dispatch, interface-lookup, and boundary-marker
shards. The longest measured individual test is 34.94 seconds.

The only failure was a stale predecessor fingerprint in the derived
architecture-evidence chain. Current source identities were propagated through
the structural strategy, portable-backend ADR, semantic-ABI feasibility, and
closed-region decision artifacts without changing any decision. The
performance ledger remains at 21 current closures and zero invalidations.
No production or canonical-stdlib change was required.

The post-consolidation non-mutating inventory is complete. Its immutable
pre-record snapshot contains 386 paths: 337 retained, 34 deferred WASM, and 15
generated-local exclusions, with zero unmatched paths. The retained paths map
to dependency-ordered compiler/runtime, benchmark, evidence, and handoff
boundaries governed by 28 dated records. JSON, gzip, whitespace, formatting,
source-size, scope, and secret-signature checks pass. The Git index remains
clean; no path was staged, committed, reset, deleted, or rewritten.

The maintainer authorized the exact-index consolidation. The sorted candidate
contains exactly 340 paths: 337 retained pre-record paths plus the three
inventory metadata files. The exact NUL-delimited pathspec, refreshed
non-self metadata identities, cached release checks, and resulting local
commit are recorded in the current inventory record. All 34 deferred WASM
paths and generated-local cache paths remain outside the commit.

Pause production performance mutation until one concrete admission
invalidation exists. Next request explicit maintainer authorization before
publishing the local branch. Why: the exact retained boundary is consolidated
locally, but pushing changes remote state and was not authorized by the local
commit request. What it entails: inspect the final commit and branch
divergence, confirm the remote destination and intended commit range, then
push only after explicit authorization. Why it is important: this prevents
publishing inactive WASM, machine-local artifacts, or an unintended commit
range while preserving the native-carrier and interpreter-free state.

Do not repeat closed checked-arithmetic, Array, frame, stack, register,
call/member/index, GC, launch-floor, or default execution-context rollout
experiments.
Any admitted production rule still requires focused semantic guards and
five-or-more balanced public-verifier baseline/candidate/reference pairs.
Keep large workspaces under disk-backed `/var/tmp`, clean disposable
artifacts, preserve the dirty deferred WASM paths, and do not begin WASM work.

## Standing Roadmap

These are continuing release conditions, not unchecked implementation tasks.
A candidate direction is not authorization to implement without the
three-unlike-program evidence gate.

### 1. Compiled AOT lowering closure

- Keep every ordinary static application under `--no-fallbacks`; final
  dependency graphs must omit `pkg/interpreter`.
- Preserve native Go carriers through static calls and use `runtime.Value`
  only at explicit dynamic, irreducibly polymorphic, host/ABI, or
  runtime-service boundaries.
- Close lowering categories generically, with no application, benchmark,
  named-container, or non-primitive nominal branches.
- Measure an admitted change against equivalent Go implementations across at
  least three unlike applications.

### 2. Bytecode VM competitiveness

- Preserve typed primitive register, slot, stack, Array, call, and return
  lanes with sound invalidation at polymorphic and dynamic boundaries.
- Reopen quickening, native operations, or frame work only when invalidated
  current profiles show one exact shared cost.
- Eliminate timeouts and compare against Python/Ruby only through broad,
  verifier-backed applications.

### 3. Scoreboard, coverage, and release gates

- Update scoreboards after every material execution change.
- Keep performance thresholds report-only until an independent clear-pass
  family repeats the five-or-more-pair protocol with a stable guard band.
- Add applications only when the external suite expands or a reusable
  language/stdlib surface lands; do not create benchmark-only APIs.

### 4. Correctness gates

- Keep `./run_all_tests.sh` and `./run_stdlib_tests.sh` green.
- Keep tree-walker and bytecode behavior, diagnostics, typechecker output, and
  shared AST fixtures aligned when language/runtime behavior changes.

## Deferred Backlog

- WASM host callbacks, loading, scheduler/filesystem ABI, and further WASM
  implementation or design.
- Parser/tree-sitter expansion not required by active performance or
  correctness work.
- Broad stdlib redesign or migration not required by a general benchmark
  surface.
- Concurrency feature expansion beyond current v12 semantics.
- Historical design-note reconciliation unrelated to active work.

## Guardrails

- Treat `spec/full_spec_v12.md` and canonical Go AST definitions as the active
  language contract.
- Preserve semantics, identity, errors, control flow, concurrency, and dynamic
  fallback boundaries.
- Do not retry rejected benchmark-specific, one-family, named-container,
  non-primitive nominal, broad execution-context ABI, duplicate-runtime,
  foreign-backend, JIT, or executable-memory routes without new cross-program
  evidence.
- Use repeated measurements and averages. Preserve verifier output, exact
  settings, baseline/candidate artifacts, and machine-readable results.
- Put large build/profile workspaces under disk-backed `/var/tmp`, not the
  RAM-backed `/tmp`; remove disposable raw artifacts after recording results.
- Keep source files below 1,000 lines and individual tests below one minute.
- Every handoff must recommend what comes next, why, what it entails, and why
  it is important.
- Remove completed items from this file and record outcomes in `LOG.md`,
  `v12/LOG.md`, and a dated decision record.

## Completed-Work Index

- Project-wide history: `LOG.md`
- v12 history: `v12/LOG.md`
- Performance decisions: `v12/docs/perf-baselines/`
- Architecture and rejected routes: `v12/design/`
- Compiler lowering authorities:
  `v12/design/compiler-go-lowering-spec.md`,
  `v12/design/compiler-go-lowering-plan.md`, and
  `v12/design/compiler-native-lowering-guardrails.md`
