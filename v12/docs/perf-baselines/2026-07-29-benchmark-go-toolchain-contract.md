# Benchmark Go toolchain provenance contract

Date: 2026-07-29

## Decision

Retain the benchmark-tooling changes. They are a general evidence-integrity
rule and do not alter compiler, generated-runtime, runtime, interpreter, VM,
stdlib, language, benchmark, fixture, dependency, nominal-lowering, or WASM
behavior.

Promoted full-scorecard refreshes now require an exact Go patch selector such
as `go1.26.5`. The selector is resolved before collection, propagated to both
external Go references and Able's CLI/generated-Go build path, recorded with
the actual `go version`, and checked again when the comparison joins the two
sides.

## Retained contract

- `bench_refresh_external_scorecard` accepts `--go-toolchain`, rejects
  promotion without it, resolves the selector once, and exports both
  `GOTOOLCHAIN` and the exact expected `go version` to every reference and
  Able comparison process.
- `bench_refresh_go_refs` accepts the same exact-patch selector, exports it
  before building, rejects selector/resolved-version disagreement, and records
  selector, actual version, and expected version in its JSON report.
- `bench_perf_json_report.py` records the selector and actual Go version used
  by the Able build lane. It refuses to write a report when the resolved
  version differs from the refresh driver's expectation.
- `bench_compare_external` preserves both contracts and rejects missing,
  mixed-version, mixed-selector, or mislabeled fresh compiled comparisons
  before writing promotable comparison evidence.
- The shared validation lives in `bench_go_toolchain_contract.py`; the
  comparison driver remains below the 1,000-line source ceiling.

Dry, non-promoting exploratory runs may omit the selector. Any refresh that
writes `external-scoreboard-current` must provide an exact patch selector.

## Verification

Fast contracts cover:

- exact-patch selector acceptance and rejection;
- resolution and recording of the requested selector;
- rejection when `go version` disagrees with the selector or expected value;
- rejection of version, selector, and missing-selector mismatches;
- refusal to promote a full refresh without `--go-toolchain`;
- propagation of the selector and resolved version through every Go-reference
  and Able comparison command in the 63-application dry-run graph.

All 31 `v12/bench_*_test.py` files pass with a 60-second per-test ceiling.
Shell syntax checks and Python compilation pass, and every maintained source
remains below 1,000 lines.

## End-to-end smoke

A quiet-CPU, verifier-backed Fibonacci smoke used:

- selector: `go1.26.5`;
- resolved version on both sides:
  `go version go1.26.5 linux/amd64`;
- one fresh Go process and one strict compiled Able process;
- `--no-fallbacks`, CPU 7, a 60-second run cap, and a 60-second build cap.

Both processes verified with zero timeouts and failures. The reference report,
Able measurement, comparison report, and embedded Go comparison all recorded
the same selector and resolved version. The observed 3.1300/2.8565-second
ratio is a one-run contract smoke, not performance evidence and not a
scorecard update.

The three exact Marketlab process groups paused for a quiet measurement were
resumed by the cleanup trap and verified running afterward. The smoke used
59 MiB under
`/var/tmp/able-v12-go-toolchain-contract-20260729`; that exact workspace and
the test-created Python bytecode cache were deleted after recording results.

## Next

Refresh the non-mutating release-boundary inventory for the completed Go
1.26.5 scorecard and toolchain-provenance tranches.

Why: these tranches add a large evidence cohort plus a small benchmark-tooling
change to an intentionally dirty tree that still contains 34 deferred WASM
files.

What it entails: enumerate and hash every current non-WASM change, map each
path to its governing dated record and dependency order, verify JSON, shell
syntax, Python contracts, whitespace, source-size, scope, and exact
WASM/non-WASM separation, and produce an auditable candidate list without
staging, committing, resetting, or deleting user work.

Why it matters: the compiled baseline and its measurement contract are now
trustworthy; an exact inventory preserves that boundary and prevents deferred
WASM work or unrelated dirty state from entering a future performance release.
