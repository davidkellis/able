# Mode-aware evidence and release gate retained

Date: 2026-07-30

## Decision

Retain no production performance change and launch no new profile or A/B
cohort. The mode-aware authority reports 23 current closures, zero
invalidations, and no selected compiled or bytecode owner. The 132-row
scoreboard and zero-actionable frontier remain current.

This tranche repaired release metadata and test orchestration only. It did not
change compiler output, generated runtime behavior, the tree-walker, bytecode
execution, canonical `able-stdlib`, dependencies, the language, benchmarks,
or deferred WASM paths.

## Checked evidence reconciliation

The ordinary fast lane exposed a content-addressed cascade beginning with the
current v12 spec hash. The bytecode native-hot-tier budget, cross-engine
structural reconciliation, portable-VM ADR, shared-runtime semantic-ABI
feasibility report, and shared-runtime closed-region cutover decision were
regenerated into a disk-backed temporary workspace before their checked JSON
was updated.

Every difference was a source fingerprint. Decision-bearing summaries,
routes, runtime-kind mappings, ownership and effect obligations, migration
constraints, next lanes, and checked Markdown remained identical. Focused
tests and each generator's `--check` mode pass.

The existing two-package dictionary-capture fixture was also added to the exec
coverage index. Coverage now reports 273 seeded entries, zero planned entries,
and 274 fixture directories; the one extra directory is the intentionally
retired fixture-target-policy control.

## Bounded test orchestration

The former fast lane ran the entire interpreter package and both aggregate
fixture tables in one 60-second Go process. It timed out while the fixture
parity table was still progressing. The fixture collector now accepts a
validated deterministic batch index/count, and the normal runner uses eight
shards for each of:

- tree-walker exec fixtures;
- tree-walker/bytecode parity; and
- bytecode exec fixtures.

Ordinary interpreter tests run once with the two aggregate tables excluded.
The complete compiler inventory is similarly discovered and run in batches of
ten. The canonical stdlib expectation regression remains explicitly isolated
and passed in 14.648 seconds. This retains all 844 compiler tests and keeps the
same 60-second per-process bound.

The normal `./run_all_tests.sh` entry point completes successfully. The
ordinary interpreter package passed in 33.801 seconds, fixture shards stayed
below 5.4 seconds, and all compiler batches plus the isolated canonical test
passed. A noisy canonical-String batch reached 49.406 seconds once; a verbose
repeat completed in 26.064 seconds and its slowest individual test took 4.88
seconds.

## Canonical stdlib and strict dependency gates

The complete external canonical stdlib suite passes in both modes:

- tree-walker: 22 seconds;
- bytecode: 17 seconds.

A fresh ten-application unlike-workload set cover then compiled with
`--no-fallbacks`, passed every public verifier, and inspected every final Go
dependency graph. All ten graphs omit
`able/interpreter-go/pkg/interpreter`, and all ten dynamic-boundary telemetry
rows record zero runtime-service calls.

The rows cover Reverse Complement, Concurrent Event Routing, Distance Field,
Array Slice Window, Quicksort, Policy Record Dispatch, Fib, k-Nucleotide,
Fixed Width 128, and Sudoku Masks. The compact machine record is
`2026-07-30-mode-aware-evidence-release-gate-retained.json`.

The exact 3,941,000 KiB task-created `/var/tmp` workspace and its pointer were
removed after final validation. It contained the shared Go cache, disposable
generated modules, and comparison candidates; no task-owned large artifact
was retained.

## Next recommendation

Automate the topological refresh and structural comparison of the checked
architecture-evidence chain. Keep production performance mutation paused and
rerun the mode-aware selector only after a compiler, runtime, stdlib, spec,
benchmark, or observer change genuinely invalidates a closure.

Why: this tranche found no actionable performance owner, but it did find that
one legitimate upstream fingerprint update currently requires five manual,
ordered report refreshes.

What it entails: declare the checked report dependencies in one manifest,
regenerate every affected node into `/var/tmp`, compare all non-source fields
and Markdown, and update checked source records only when those comparisons
are identical. Fail closed when a decision-bearing field changes.

Why it matters: reliable evidence propagation ensures a real native-lowering
or bytecode opportunity is selected promptly while preventing stale hashes
from being confused with a performance invalidation. It also protects the
interpreter-free compiled architecture without manufacturing another cold,
named-container, or already-native experiment.
