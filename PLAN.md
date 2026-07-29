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

- `v12/docs/perf-baselines/2026-07-28-active-go-dependency-security-refresh.md`
- `v12/docs/perf-baselines/2026-07-28-post-frontier-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-28-v12-correctness-release-readiness-closure.md`
- `v12/docs/perf-baselines/2026-07-28-bytecode-semantic-boundary-reach-closure.md`
- `v12/docs/perf-baselines/2026-07-28-bytecode-current-source-profile-coverage-closure.md`
- `v12/docs/perf-baselines/2026-07-28-bytecode-current-profile-coverage.md`
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-closure.md`
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-stability-compiled.md`
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-stability-bytecode.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-26-compiled-aot-native-carrier-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological retained state is in `LOG.md` and `v12/LOG.md`. Strict
fallback-free applications must remain interpreter-free, and the existing
native-carrier, nominal-encoding, and dynamic-boundary guards remain
authoritative.

### Authoritative performance frontier

The portable catalog and reviewed selection now contain 63 applications and
126 rows: 63 compiled and 63 bytecode. All 126 rows are verifier-backed and
ranked with five current Able processes and five current processes for every
required Go/Python/Ruby reference. Compiled canonical workloads remain
unchanged; the seven formerly unranked bytecode rows use the documented
portable mode contract.

Compiled currently has 6 target passes, a 5.575597× geometric-mean Able/Go
ratio, and 5.675368 seconds of positive target excess. Bytecode has 4 target
passes, a 12.780200× geometric mean over its Python/Ruby ratios, and
221.503684 seconds of positive target excess. A second independent matched
five-process cohort confirms all ten snapshot passes as established guards.
All 63 strict compiled graphs remain interpreter-free, and the 23-closure
performance-evidence ledger has zero invalidations.

There is no admitted production optimization candidate. All 59 current
bytecode misses have source-identity-checked CPU and allocation coverage. The
operation-level semantic-boundary map then covered eight unlike family
representatives and 25,994,536 main-only primitive transitions with no
dropped sites. Its only three broad exact shapes are the already-rejected
generic nominal-field carrier rule or its aggregate member-write form. No
production code was retained. The 23-closure ledger remains current with zero
invalidations.

Current correctness and release-readiness gates are complete. The ordinary
suite, canonical stdlib in both interpreter modes, compiled CLI integration,
and all 640 generated-program compiler audit partitions pass. No individual
test exceeds one minute when measured from exact serial JSON events.

The post-frontier release inventory's immutable snapshot contains 212 paths:
178 retained and 34 deferred WASM paths. Its exact 181-path retained boundary
was consolidated as commit `d42aab3b004ba121481bba4503d1635be5556b7d` and
published to `origin/master`; the deferred WASM hold remained outside that
commit.

The active v12 Go dependency-security refresh is complete locally. Go 1.26.5,
`go-git` v5.19.1, `x/crypto` v0.52.0, and their tidied safe transitive graph
reduce the active module from 26 symbol-reachable advisories to zero.
Module verification, vet/build, the ordinary handoff, the complete compiler
package, three strict interpreter-free application builds, and five-process
compiled/Go smoke cohorts pass. QuickSort and Binary Trees remain faster than
Go; Matrix Multiply remains its existing target miss with an essentially
unchanged Able mean. No production compiler, runtime, interpreter, VM, stdlib,
language, benchmark, fixture, archived-version, nominal-special-case, or WASM
code changed.

The maintainer has authorized exact consolidation and publication of the
six-path active-v12 security tranche. That boundary is the only non-WASM
working-tree change; the deferred WASM complement must remain outside the
index and commit.

After publication, next refresh the complete 63-row compiled Able/Go
scorecard with Go 1.26.5 before reopening production performance mutation.

Why: the security toolchain patch changes the Go reference version from
1.26.4 to 1.26.5. The three-application retention smoke is green, but the
authoritative compiled frontier still records Go 1.26.4.
What it entails: rebuild all 63 Go references with Go 1.26.5, run five current
strict compiled Able processes and five matching Go processes per application
under the existing verifier and execution contracts, regenerate the compiled
scorecard/frontier evidence, and inspect target-status or owner-rank changes.
Retain no production optimization unless the refreshed evidence admits a
general owner across at least three unlike programs.
Why it matters: this restores a single patched-toolchain provenance for the
95%-of-Go goal and distinguishes a real performance opportunity from
workstation noise or a stale reference baseline.

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
