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

- `v12/docs/perf-baselines/2026-07-30-post-parser-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-30-post-parser-release-inventory.json`
- `v12/docs/perf-baselines/2026-07-30-post-parser-release-inventory.tsv`
- `v12/docs/perf-baselines/2026-07-30-parser-go-binding-contract-retained.md`
- `v12/docs/perf-baselines/2026-07-30-parser-go-binding-contract-retained.json`
- `v12/docs/perf-baselines/2026-07-30-post-publication-security-reconciliation.md`
- `v12/docs/perf-baselines/2026-07-30-post-publication-security-reconciliation.json`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-release-inventory.md`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-release-inventory.json`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-release-inventory.tsv`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-correctness-release-gate.md`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-correctness-release-gate.json`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-cross-family-architecture-ownership-reconciliation.md`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-cross-family-architecture-ownership-reconciliation.json`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-compiled-concurrency-reconciliation.md`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-compiled-concurrency-reconciliation.json`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-compiled-architecture-owner-closure.md`
- `v12/docs/perf-baselines/2026-07-30-post-nullable-compiled-architecture-owner-closure.json`
- `v12/docs/perf-baselines/2026-07-30-compiled-primitive-nullable-value-carrier-retained.md`
- `v12/design/compiler-primitive-nullable-value-carrier.md`
- `v12/docs/perf-baselines/2026-07-30-nullable-scalar-retained-frontier.md`
- `v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.md`
- `v12/docs/perf-baselines/external-scoreboard-current.md`
- `v12/docs/perf-baselines/2026-07-24-static-interpreter-package-cut-retained.md`

The chronological retained state is in `LOG.md` and `v12/LOG.md`. Strict
fallback-free applications must remain interpreter-free, and the existing
native-carrier, nominal-encoding, and dynamic-boundary guards remain
authoritative.

### Authoritative performance frontier

The reviewed scorecard contains 65 compiled and 65 bytecode rows, each backed
by five successful Able and reference processes. Compiled has 6 established
target guards, a 5.740300× geometric-mean Able/Go ratio, and 6.696421 seconds
of positive target excess. Bytecode has 4 established guards, a 17.733801×
geometric mean over the per-row limiting Python/Ruby ratios, and 266.723789
seconds of positive target excess. The combined frontier has 10 guards, 120
misses, 273.420211 seconds of positive target excess, no unestablished
snapshot meets, and zero actionable frontier groups.

Primitive nullable scalars now use generated `value + valid` Go carriers.
Five-run frozen A/B means improved Generic Slot Buffer from 0.0560 to 0.0340
seconds, Inventory Reconciliation from 0.1220 to 0.1100 seconds, and
Transaction Ledger Audit from 0.0480 to 0.0400 seconds. Exact main allocation
objects fell 99.60%, 48.88%, and 8.88% respectively. Present primitive zero
remains distinct from absent. `?Error`, non-primitive nominal nullables, and
explicit runtime/dynamic boundaries retain their existing representation.

All three strict application graphs omit `pkg/interpreter`. No stdlib,
tree-walker, bytecode VM, runtime package, language, dependency, or WASM
change was required. `go test ./cmd/ablec` and the complete
`./run_all_tests.sh` suite pass; the latter completed every preflight,
non-compiler package, all 34 compiler batches, and bytecode fixtures.

All 12 closures invalidated by the compiler-production identity change have
now been causally reviewed. The ten compiled family records partition all 65
compiled frontier rows exactly once; every row has current strict
verifier-backed evidence and no interpreter dependency. The complete
23-entry ledger is current with zero invalidations and an empty selector.

The bounded post-nullable correctness/release gate is complete. The full
default v12 runner passed every preflight, non-compiler package, all 34
compiler batches, and the complete bytecode fixture pass. Every audited named
test remained below one minute. The canonical external stdlib passed in both
tree-walker and bytecode modes. No production correction was required.

The fully expanded release inventory now classifies 324 pre-record dirty
paths exactly: 290 retained and the unchanged 34-path deferred WASM boundary.
The three inventory metadata files form an exact 293-path post-record retained
candidate whose complement is precisely those 34 deferred files. All format,
syntax, identity, size, secret, scope, scorecard, frontier, ledger, and cleanup
checks pass. The maintainer authorized and the repository now contains one
exact local consolidation commit for those 293 retained paths. Commit
`be9ecc505161085e1ec11f704571f589b3366c13` is published; local `HEAD`,
`origin/master`, and remote `master` agree with zero divergence.

The post-publication security reconciliation retains no dependency or
production change. Go 1.26.5 production and test scans have zero reachable
symbol and imported-package findings; only the unimported, unfixable
`x/crypto/openpgp` module advisory remains. All four canonical npm lockfiles
are clean. GitHub's 74-alert repository banner is one moderate alert lower
than the previously attributed banner and does not identify an active-v12
finding.

The standalone parser Go binding contract is now reproducible. The obsolete
nested module boundary was removed; the grammar-root module owns
`bindings/go`, retains the official `go-tree-sitter v0.24.0` runtime, and has
a complete sum identity. The routine v12 runner now tests the explicit
binding package with a one-minute limit. Root module integrity, standalone
grammar loading, the official Able parser packages, and a test-inclusive Go
1.26.5 vulnerability scan all pass. The complete v12 runner also passes every
preflight, non-compiler package, all 34 compiler batches, and the bytecode
fixture corpus with zero swaps. No parser semantic or performance input
changed.

The post-parser release inventory classifies all 45 pre-record dirty paths
exactly: 11 retained security/parser/handoff paths and the unchanged 34-path
deferred WASM boundary. The three inventory metadata files form an exact
14-path post-record retained candidate. Every manifest identity, format,
module, scope, secret, source-size, scoreboard, frontier, ledger, canonical
stdlib, index, and cleanup check passes. The maintainer authorized exact
staging and one local commit; the repository now contains that consolidation
as one local commit, with only the 34 deferred WASM paths left dirty. No push
is authorized.

### Next tranche

Obtain explicit maintainer authorization before publishing the local
post-parser consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized. Do not publish deferred WASM or unrelated local history.

Why it matters: the retained security evidence and parser binding correction
can reach shared history without publishing deferred WASM or unintended
commits.

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
