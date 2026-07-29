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

- `v12/docs/perf-baselines/2026-07-29-post-security-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-29-v12-x-net-security-refresh.md`
- `v12/docs/perf-baselines/2026-07-29-dependency-alert-attribution.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-correctness-release-gate.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-scorecard-and-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-scorecard-refresh.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-frontier.md`
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-closure-ledger.md`
- `v12/docs/perf-baselines/2026-07-29-compiled-spawn-gated-callable-context-retained.md`
- `v12/design/compiler-spawn-gated-callable-context-activation.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological retained state is in `LOG.md` and `v12/LOG.md`. Strict
fallback-free applications must remain interpreter-free, and the existing
native-carrier, nominal-encoding, and dynamic-boundary guards remain
authoritative.

### Authoritative performance frontier

The reviewed scorecard contains 63 compiled and 63 bytecode rows, each backed
by five successful Able and reference processes. Compiled has 7 target passes,
a 4.320152× geometric-mean Able/Go ratio, and 4.751789 seconds of positive
target excess. Bytecode remains at 4 target passes, a 12.780200× geometric
mean over Python/Ruby ratios, and 221.503684 seconds of positive target
excess. All eleven passes remain established guards. Binary Trees averages
10.2700 seconds versus Go at 11.0316 seconds, or 107.42% of Go throughput.

The selective post-spawn refresh promotes new Go 1.26.5 evidence only for the
20 source-changed compiled rows and preserves the 43 source-identical rows.
All 290 accepted Able/reference timing processes passed their public
verifiers. Nine measured source-identical controls drifted by a 10.31%
geometric mean under unrelated host load, so their noisy replacement evidence
was rejected. An earlier 15-process mixed-toolchain attempt was also rejected.

Fresh profiles across Await Channel Mux, Validated Job Pipeline, and
Concurrent Stateful Pipeline find `bridge.ToInt` in all three. It is ordinary
semantic-boundary materialization whose global cache already regressed the
TapeLang guard by 4.17%. `bridge.currentGID` and nominal struct construction
each remain material in only two applications. The combined frontier now has
11 established guards, 115 misses, 226.255474 seconds of positive target
excess, and zero actionable groups. All 23 closure-ledger entries are current
with zero invalidations. No additional production optimization was retained.

The post-spawn correctness gate is green. `./run_all_tests.sh` passed every
contract, non-compiler package, all 34 compiler batches, and the complete
bytecode fixture pass in 15:50.94 with no swaps. The canonical external stdlib
passed in tree-walker mode in 19 seconds and bytecode mode in 15 seconds.
Individual timing replays of the three compiler aggregates over one minute
found maxima of 22.910, 15.150, and 10.740 seconds, so no individual test
exceeds the one-minute limit. No correctness fix or production change was
required.

The exact 147-path post-spawn consolidation is published as
`6efad0a53120129510fdfbab7fbcc84dcd081768`; local `HEAD` and
`origin/master` match. The subsequent dependency-alert attribution finds zero
reachable active-v12 Go vulnerabilities at the required Go 1.26.5 toolchain
and zero npm vulnerabilities. Seven fixable `x/net` advisories and the
unfixable, unimported `x/crypto/openpgp` advisory remain module-only. Frozen
v10 and v11 each retain 25 reachable historical advisories and remain
untouched.

The retained bounded security refresh advances `x/net` to `v0.56.0` with
resolver-required `x/crypto v0.53.0` and `x/sys v0.46.0`. Go 1.26.5
production and test scans remain at zero reachable and zero imported-package
vulnerabilities; all seven fixable `x/net` advisories are gone. Only the
unfixable, unimported `x/crypto/openpgp` module advisory remains. Every Go
package compiles, focused dependency-resolver tests pass, and the performance
frontier and 23-entry closure ledger remain current.

### Next tranche

Obtain explicit maintainer authorization before publishing the exact local
post-census consolidation.

Why: the maintainer authorized exact staging and one local commit for the
eight-path retained candidate, but did not authorize a push.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized.

Why it matters: the evidence-backed stopping decision can reach shared history
without publishing deferred WASM or unintended local history.

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
