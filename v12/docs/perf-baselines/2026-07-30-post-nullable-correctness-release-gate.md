# Post-nullable correctness and release gate

Date: 2026-07-30

## Decision

The retained post-nullable v12 state passes the bounded correctness and
release gate. Retain no additional production change.

The complete default v12 runner passed every deterministic contract,
non-compiler package, all 34 compiler short-mode batches, and the complete
bytecode fixture corpus. The canonical external stdlib then passed in both
tree-walker and bytecode modes. No correctness failure, semantic mismatch,
timeout, verifier failure, or swap event occurred.

This tranche verifies the existing dirty worktree. It does not authorize or
begin deferred WASM work, and it does not change compiler, runtime,
interpreter, bytecode VM, canonical stdlib, language, dependency, benchmark,
fixture, or nominal-lowering behavior.

## Environment

All build state was placed under the disk-backed workspace
`/var/tmp/able-v12-release-20260730.he0nIY`:

- Go toolchain: Go 1.26.5, as pinned by `v12/interpreters/go/go.mod`;
- fixture typechecking: strict;
- explicit disk-backed `GOCACHE`, `GOTMPDIR`, and `TMPDIR`;
- `PYTHONDONTWRITEBYTECODE=1`;
- canonical stdlib: sibling `../able-stdlib`;
- canonical stdlib source tree:
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`;
- canonical stdlib Git head:
  `219eff222c28406487231713753641bc49ee5b9a`; and
- unrelated host workloads remained active and were not paused or modified.

The freshly captured stdlib source state exactly matched
`2026-07-29-post-spawn-context-stdlib-source-state.json`. The stdlib worktree
was already dirty and this tranche made no change there.

## Complete v12 gate

Command:

```sh
GOTOOLCHAIN=go1.26.5 \
GOCACHE=/var/tmp/able-v12-release-20260730.he0nIY/gocache \
GOTMPDIR=/var/tmp/able-v12-release-20260730.he0nIY/gotmp \
TMPDIR=/var/tmp/able-v12-release-20260730.he0nIY/tmp \
PYTHONDONTWRITEBYTECODE=1 \
./run_all_tests.sh
```

Result:

| Measure | Result |
| --- | ---: |
| Exit status | 0 |
| Wall time | 15:50.29 |
| User time | 2,149.31s |
| System time | 147.67s |
| CPU utilization | 241% |
| Peak RSS | 4,711,072 KB |
| Swaps | 0 |
| Compiler tests listed | 836 |
| Compiler short-mode batches | 34 / 34 passed |
| Final bytecode fixture pass | 89.806s, passed |

The pre-Go contract layer passed:

- execution coverage index: 272 seeded rows, zero planned rows, 273 fixture
  directories;
- authoritative external scoreboard and five-sample evidence;
- feature-to-application coverage: 15 families, 16 normative sections, 65
  portable applications, and three local-only applications;
- 130-row benchmark selection, execution, and refresh contracts;
- external threshold controls;
- generated-artifact cleanup policy; and
- canonical/embedded kernel synchronization.

Every non-compiler package passed. The short interpreter package aggregate
took 103.960 seconds. All 34 compiler batches passed, and the final complete
bytecode fixture pass took 89.806 seconds.

## Individual-test timing audit

Four 25-test compiler aggregates exceeded one minute under normal host load:

| Batch | Aggregate |
| ---: | ---: |
| 5 | 60.649s |
| 20 | 106.686s |
| 29 | 71.763s |
| 30 | 105.171s |

They were replayed with the same toolchain, fixture mode, and disk-backed
caches using `go test -short -json`. Every selected test passed or explicitly
skipped in short mode, and no individual test approached one minute:

| Batch | Passed | Short-mode skips | Slowest individual test | Elapsed |
| ---: | ---: | ---: | --- | ---: |
| 5 | 19 | 6 | `TestCompilerConcreteEnumerableGenericMethodsExecute` | 3.420s |
| 20 | 21 | 4 | `TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete` | 24.040s |
| 29 | 20 | 5 | `TestCompilerCanonicalSpecNullableExpectationLowers` | 15.310s |
| 30 | 25 | 0 | `TestCompilerNoFallbacksStringBuilderUsesNativeArrayPushAll` | 10.720s |

The two interpreter aggregates were also replayed with JSON timing:

| Run | Slowest named test | Elapsed | Named tests over 60s |
| --- | --- | ---: | ---: |
| Short-mode interpreter | `TestExecFixtureParity` | 34.590s | 0 |
| Complete bytecode fixtures | `TestExecFixtureParity` | 34.070s | 0 |

The aggregate durations therefore do not violate the project rule that one
test must not take more than one minute.

## Canonical stdlib gate

Command:

```sh
GOTOOLCHAIN=go1.26.5 \
GOCACHE=/var/tmp/able-v12-release-20260730.he0nIY/gocache \
GOTMPDIR=/var/tmp/able-v12-release-20260730.he0nIY/gotmp \
TMPDIR=/var/tmp/able-v12-release-20260730.he0nIY/tmp \
PYTHONDONTWRITEBYTECODE=1 \
./run_stdlib_tests.sh
```

Result:

| Mode | Result | Elapsed |
| --- | --- | ---: |
| Tree-walker | passed | 19s |
| Bytecode | passed | 14s |
| Complete command | passed | 34.68s |

The command used 895,340 KB peak RSS and recorded no swaps.

## Derived evidence checks

Post-suite checks passed:

- canonical stdlib source-state recapture and exact comparison;
- authoritative 130-row external scorecard with five successful Able and
  reference processes per row;
- 130-row frontier with zero actionable groups;
- 23-entry closure ledger with zero invalidations and an empty selector; and
- JSON validation and final whitespace checks.

No implementation fix was required. The only retained changes from this
tranche are this record and the forward-looking handoff/log updates. The
machine-readable companion is
`2026-07-30-post-nullable-correctness-release-gate.json`.

The exact disk-backed workspace was removed after final verification; no
tranche build, test, timing, or source-state artifact remains.

## Next recommendation

Refresh the exact retained/deferred worktree release inventory without
staging, committing, or pushing.

Why: the behavioral release gate is green, but the worktree contains a large
accumulation of retained compiler/evidence paths alongside separately
deferred WASM work and disposable generated artifacts. A stale boundary is
now the largest release risk.

What it entails: capture fully expanded porcelain plus HEAD, origin, and index
identities; classify every changed path as retained, deferred, generated, or
untracked; validate JSON, gzip, formatting, syntax, source-size, secret, and
path-complement checks; and write an exact non-mutating manifest. Staging,
committing, and pushing remain outside that tranche without explicit
authorization.

Why it matters: an exact inventory protects the verified native-carrier,
interpreter-free, scorecard, and correctness state from accidental inclusion
of deferred WASM or machine-local artifacts.
