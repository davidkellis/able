# Relocatable performance-evidence gates retained

Date: 2026-07-30

## Decision

Retain two general evidence-tooling corrections:

1. repository-owned provenance paths recorded from another checkout resolve
   against the active repository's retained `v12/...` path; and
2. native `bytecode-production` closure hashing excludes `*_wasm.go`, matching
   the existing shared-interpreter native scope boundary.

Both the dirty developer worktree and a relocated source-only export now pass
the complete checked performance-evidence gates with 23 current closures,
zero invalidations, and an empty selector. No production, parser, compiler,
runtime, interpreter, VM, language, stdlib, benchmark-measurement, fixture, or
WASM source changed.

## Repository-owned path relocation

Historical selected reports legitimately preserve the absolute checkout paths
that produced their measurements. Those paths are provenance, not an
instruction to read a different Able worktree during validation.

The new shared helper `v12/bench_repository_paths.py` applies one rule:

- relative repository paths resolve under the active repository;
- absolute paths already under the active repository remain unchanged;
- absolute repository-owned paths from another checkout follow their retained
  `v12/...` suffix under the active repository; and
- paths without a `v12` repository suffix remain external.

The scoreboard uses this rule for measured Able sources and retained
reference-report JSON. The five-sample checker uses it before enforcing that
selected reports and their reference reports live under the active
`perf-baselines` directory.

This preserves external Go/Python/Ruby implementation and verifier paths while
preventing an old Able checkout from becoming a hidden validation dependency.

## Native bytecode scope

The `bytecode-production` evidence scope previously included all
`bytecode*.go` and `interpreter_bytecode*.go` files except tests. Three
untracked deferred WASM proof files therefore entered the checked baseline:

- `bytecode_i32_frame_proof_wasm.go`;
- `bytecode_scalar_proof_wasm.go`; and
- `bytecode_type_proofs_wasm.go`.

The scope now also excludes `*_wasm.go`. This is an evidence-boundary
correction, not WASM implementation work. It prevents target-specific,
deferred files from invalidating native Go bytecode-performance closures.

The corrected scope is:

| Property | Before | After |
| --- | ---: | ---: |
| files | 206 | 203 |
| tree SHA-256 | `b8472c5b17856a9eed8eecb4dcecd2e3030d6024e7f50ed050f5ae06902de197` | `4c6b98eafd856f6bda13884e8b32004bf5af2842b4fcf6d2990e8c9321cd39a1` |
| current closures | 23 in the dirty tree, 11 in a clean export | 23 in both |
| invalidated closures | 0 in the dirty tree, 12 in a clean export | 0 in both |

The rebuilt closure ledger has SHA-256
`15fcba70f930f002e7ef9e00084c4986973248db817e989266a693cc1c40f9c2`.
The evaluated JSON and Markdown reports remain byte-identical because the
corrected committed scope still selects nothing.

## Focused regression coverage

Four fast path tests cover:

- an ordinary retained reference report;
- a relocated retained source and reference report;
- rejection of a genuinely external report; and
- relocation of a measured Able source in the scoreboard.

The ledger suite now contains a focused WASM-boundary regression. It
bootstraps a native bytecode scope containing one native file and one
`*_wasm.go` file, verifies that only the native file is hashed, modifies the
WASM file, and proves that no closure is selected.

Developer-worktree results:

- scoreboard check: 130 rows, pass;
- five-sample check: 130 rows, five successful Able/reference samples each,
  31 retained source reports, and 31 retained reference reports;
- path relocation tests: four pass;
- frontier: 130 rows, zero actionable groups;
- frontier tests: five pass;
- ledger: 23 current closures, zero invalidations;
- ledger tests: eight run, seven pass and one conditional test skipped;
- Python compilation, JSON parsing, whitespace, file-length, and guarded
  cleanup checks: pass.

No changed source reaches 1,000 lines. The largest is
`v12/bench_external_scoreboard` at 995 lines.

## Relocated-export proof

A fresh Git archive of published `e7a05403` was overlaid only with the
candidate tooling, tests, and corrected ledger. It contains 10,959 regular
files and none of the three deferred WASM proof files. The canonical external
stdlib was supplied as the expected adjacent source dependency.

Inside that relocated export:

- the scoreboard passes all 130 rows;
- the five-sample checker passes all 130 rows and 31 retained source/reference
  report pairs;
- all four path relocation tests pass;
- the frontier passes with zero actionable groups;
- the ledger passes with all 23 closures current and zero invalidations; and
- the eight-test ledger suite passes, with only its intentionally conditional
  invalidated-Markdown test skipped.

This closes both failures from the shared-tree reconciliation. The checked
performance state now depends on committed native inputs, not the checkout
location or deferred WASM files.

## Cleanup

All validation, Python cache, and relocated-export state lived under
disk-backed `/var/tmp`. The exact 140,068 KiB (136.79 MiB) workspace was
removed after preserving this record. No build state used RAM-backed `/tmp`;
the guarded project cleanup dry run is empty, and no task workspace remains.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: the worktree now combines retained publication records and
evidence-tooling corrections with the unchanged 34-path deferred WASM
boundary.

What it entails: expand all untracked files, record exact state, line, byte,
and SHA-256 identities, reproduce the 34-path deferred complement, and define
the retained candidate without repository mutation.

Why it matters: an exact release boundary lets the self-contained evidence
correction enter history without absorbing deferred WASM work.
