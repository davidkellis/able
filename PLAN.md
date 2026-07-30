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

- `v12/docs/perf-baselines/2026-07-30-guarded-generated-local-cleanup-retained.md`
- `v12/docs/perf-baselines/2026-07-30-post-readiness-release-consolidation-inventory.md`
- `v12/docs/perf-baselines/2026-07-30-extern-plugin-toolchain-release-readiness-retained.md`
- `v12/docs/perf-baselines/2026-07-30-full-compiled-static-native-boundary-census.md`
- `v12/docs/perf-baselines/2026-07-30-post-captured-callable-compiled-closure-reconciliation.md`
- `v12/docs/perf-baselines/2026-07-30-statically-monomorphic-captured-callable-retained.md`
- `v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological completed state is in `LOG.md` and `v12/LOG.md`. Detailed
performance decisions live under `v12/docs/perf-baselines/`; completed task
narratives do not belong in this plan.

Strict fallback-free compiled applications must remain interpreter-free.
Primitive native carriers, shared nominal lowering, explicit dynamic-boundary
boxing, the fail-closed active fixture-target policy, and the full-mode parser
fixture lane remain authoritative.

The guarded cleanup removed the four inventoried generated paths and
1,742,116 KiB of reproducible cache state after proving that they contained no
tracked files and had no active process owner. The final release candidate is
112 paths after the cleanup records are included. Its complement is exactly
the unchanged 34-path deferred WASM boundary; no generated-local or unmatched
path remains.

The current scorecard contains 66 compiled and 66 bytecode rows. The combined
frontier has ten guards, 122 misses, zero actionable groups, and 277.200421
seconds of aggregate target excess. All 23 checked closures are current with
zero invalidations. The completed boundary census admits no new safe lowering
mechanism shared by three unlike applications.

### Next tranche

Verify the authorized exact local consolidation, then obtain explicit
maintainer authorization before publishing it.

Why: the exact 112-path staging and one local commit are authorized, but no
remote mutation is authorized.

What it entails: confirm the local commit is exactly one ahead of
`origin/master`, has parent `418886c70aee64b92b5bb3266ee5fe6453ac4320`,
contains exactly the reviewed 112 paths, leaves an empty index, and preserves
the unchanged 34-path deferred WASM worktree. Then, only if separately
authorized, push that one commit to the intended remote branch. Do not use a
broad refspec, reset, rewrite history, or touch WASM.

Why it matters: independent post-commit verification protects the exact
native-carrier and interpreter-free boundary, while separate publication
authorization prevents deferred WASM or an unintended ref from reaching
shared history.

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
