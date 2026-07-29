# V12 correctness and release-readiness closure

Date: 2026-07-28

## Decision

The current v12 release surface is green. Retain no production code for this
tranche.

The audit found that several derived architecture artifacts still encoded the
predecessor 119-row scorecard totals after the portable bytecode promotion
expanded the selected frontier to 126 rows. The repair updated the residual
cohort rationale and expected totals, removed one hard-coded closure count,
and mechanically regenerated the dependent residual, interaction,
architecture-budget, structural-strategy, portable-backend, shared-runtime,
and bytecode-tier artifacts. Their decisions did not change, and no prototype
became admissible.

The audit also repaired the release harness so its generated-program matrices
are partitioned finely enough for the one-minute individual-test policy, and
split or narrowed three oversized compiler tests without weakening their
assertions.

No compiler production path, generated runtime, runtime, interpreter, bytecode
VM, canonical stdlib, language, dependency, benchmark, fixture, reference
application, or WASM behavior changed.

## Current release gates

All build-heavy work used the exact disk-backed workspace
`/var/tmp/able-v12-release-7SwdTv`; no large generated build was placed on the
RAM-backed `/tmp`.

- `go vet ./...` and `go build ./...` pass.
- The ordinary `./run_all_tests.sh` handoff passes every preflight,
  non-compiler package, interpreter fixture, bounded compiler batch, and the
  bytecode fixture corpus.
- Canonical stdlib tests pass in both reference modes: tree-walker in 18
  seconds and bytecode in 15 seconds.
- The cold `./run_all_tests.sh --compiled-cli` lane passes; the generated-Go
  CLI package reports 1,190.720 seconds.
- Architecture-budget, scoreboard, scorecard-evidence, closure-ledger, and
  frontier reproduction checks pass.
- Shell syntax, JSON validity, diff whitespace, and maintained Go source-size
  checks pass.

The compiler release matrix passes:

- compiler bridge and all bounded core batches;
- the concurrency, diagnostics, and dynamic-boundary parity outliers;
- 128/128 fallback-audit partitions;
- 128/128 compiled-execution partitions;
- 128/128 strict-dispatch partitions;
- 128/128 interface-lookup partitions; and
- 128/128 boundary-marker partitions.

That is 640 green generated-program audit partitions. Together they verify
output, fallback classification, strict generated dispatch, static
interface-lookup bypass, and absence of fallback markers across the full
fixture corpus.

## Individual-test timing

The audit found one genuine individual duration violation:
`TestCompilerTypedArrayDefaultMethodsKeepConcreteReceivers` took 117.87
seconds. It compiled two independent canonical stdlib graphs and imported the
broad `able.spec.*` surface.

The test-only repair:

- split the String and i32 cases into independently reported tests;
- factored their shared assertion helper;
- imported the actual canonical owner
  `able.collections.enumerable.{Enumerable}`; and
- preserved generated-Go compilation, specialized `drop` helper assertions,
  and concrete lazy `Iterator<String>`/`Iterator<i32>` assertions.

The new tests take 7.40 and 7.17 seconds. The adjacent canonical
specialization guard plus both replacements pass together in 39.54 seconds.

Concurrency parity is now four top-level batches and diagnostics parity two
top-level batches. Their exact fixture subtests remain unchanged; the longest
observed parity subtest was 28.22 seconds.

The full compiler core JSON audit found a 53.09-second maximum and zero
individual tests at or above one minute. Release matrices now use 128
partitions instead of 24:

| Matrix | Result | Maximum under four-worker contention | Serial replay when needed |
| --- | ---: | ---: | ---: |
| fallback | 128/128 | 58.707 s | not needed |
| compiled execution | 128/128 | 123.709 s | 46.896 s package; 24.00 s fixture |
| strict dispatch | 128/128 | 126.613 s | 16.839 s package; 16.70 s fixture |
| interface lookup | 128/128 | 67.466 s | 19.541 s package; 19.43 s fixture |
| boundary marker | 128/128 | 60.809 s | 17.598 s package; 17.50 s fixture |

The over-one-minute figures are whole processes competing with three other
compiler processes. Exact serial JSON events demonstrate that no individual
fixture crosses the policy limit. The unchanged compiled-CLI test code retains
its prior warm JSON-event maximum of 31.09 seconds.

## Evidence reconciliation

The final deterministic pass confirms:

- 63 portable applications and 126 selected rows: 63 compiled and 63
  bytecode;
- complete five-sample Able/reference scorecard evidence;
- 23 current performance closures and zero invalidations;
- zero actionable frontier groups;
- 6/63 compiled and 4/63 bytecode target passes;
- a 5.575597x compiled Able/Go geometric-mean ratio;
- a 12.780200x bytecode geometric mean over Python/Ruby ratios; and
- all architecture-budget and threshold-control checks.

The reconciled common 63-application architecture model says a theoretical
compiled-native bytecode proxy would remove 93.010928% of measured bytecode
target excess, with 34 projected target passes and 29 misses. However, the
checked feasibility chain still finds zero concrete eligible hot-function
classes and zero target closures. This remains strategic evidence, not an
admissible production implementation.

The v12 specification did not change.

## Cache and temporary-file cleanup

The cold release work raised the disposable disk-backed workspace to 28 GiB.
It contains the isolated Go cache, generated application workspaces, and logs
used for this record. After recording the results, the exact workspace, the
29 MiB `/tmp/able-v12-extern-go` directory, and 284 KiB of test-created Python
bytecode were deleted. No broad `/tmp`, repository, user-cache, or
caller-owned cleanup was performed.

## Recommendation

Keep production performance mutation paused and next refresh the non-mutating
post-frontier release-consolidation inventory.

Why: all current CPU/allocation coverage and causal boundary evidence agree
that there is no open general owner spanning three unlike programs, while the
126-row release surface and 640-partition compiler matrix are now green.

What it entails: map changed and untracked v12 paths added since the preceding
consolidation to their dated records and dependency-ordered review boundaries;
identify unmatched or generated-local paths without deleting, staging,
committing, resetting, or touching deferred WASM work.

Why it is important: this turns the verified native-lowering and
interpreter-free state into an auditable release candidate, protects the
extremely dirty worktree, and prevents a benchmark-specific experiment from
reopening a measured regression. Resume production performance work only when
a checked evidence invalidation or a new exact open owner reaches three unlike
families.
