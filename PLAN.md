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
`v12/docs/perf-baselines/external-scoreboard-current.{json,md}`.

## Current Handoff

Start from:

- `v12/docs/perf-baselines/2026-07-31-rank1-interface-hkt-static-boundary-retained.md`
- `v12/docs/perf-baselines/2026-07-31-fixed-integer-alternative-arithmetic-retained.md`
- `v12/docs/perf-baselines/2026-07-31-integer-literal-contextual-native-carrier-retained.md`
- `v12/docs/perf-baselines/2026-07-31-variance-invariant-v12-editorial-reconciliation-retained.md`
- `v12/docs/perf-baselines/2026-07-31-block-comment-contract-editorial-reconciliation-retained.md`
- `v12/docs/perf-baselines/2026-07-31-mode-aware-release-correctness-gate-retained.md`
- `v12/docs/perf-baselines/2026-07-31-current-default-primitive-boxing-boundary-census-no-go.md`
- `v12/docs/perf-baselines/2026-07-31-current-default-three-application-owner-no-go.md`
- `v12/docs/perf-baselines/2026-07-31-compiled-closure-refresh-retained.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological completed state is in `LOG.md` and `v12/LOG.md`. Detailed
performance decisions live under `v12/docs/perf-baselines/`; completed task
narratives do not belong in this plan.

All 23 performance-evidence closures are current. The default compiled corpus
has 66/66 verifier-backed rows, 66/66 resolved strict dependency graphs, and
zero interpreter links. Nine compiled rows now have established target guards;
57 remain misses. The full compiled Able/Go geometric-mean ratio is 4.2187x,
while 8 of the 11 rows with Go runtimes above 100 ms meet the target.

### Next tranche

Resolve the remaining shared-data race and ownership guidance for v12
concurrency.

Why: `spawn` already maps to Go goroutines and Able nominal values already use
native carriers, but the language does not yet state whether unsynchronized
shared mutation is invalid, undefined, or dynamically checked. Leaving that
open makes native compiled concurrency behavior and interpreter behavior hard
to align without accidentally introducing defensive copies or runtime checks.

What it entails: inventory captures, mutable nominal carriers, channels,
mutexes, interpreter scheduling, and compiled goroutine lowering; select the
race-free source contract in the spec first; then add only the diagnostics,
cross-engine fixtures, or native-lowering guards required by that decision.

Why it matters: a source-level race and ownership rule can preserve direct
native Go references and synchronization while avoiding implicit copying,
boxing, or a runtime ownership layer at the compiled/interpreted boundary.

The completed inference/HKT tranche is recorded in
`v12/docs/perf-baselines/2026-07-31-rank1-interface-hkt-static-boundary-retained.md`.

Do not repeat closed checked-arithmetic, Array, frame, stack, register,
call/member/index, GC, launch-floor, or global default execution-context
rollout experiments. Do not narrow spawn activation using application names,
nominal types, source families, or measured runtime-count thresholds.
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
