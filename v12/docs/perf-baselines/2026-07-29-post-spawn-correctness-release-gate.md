# Post-spawn correctness and release gate

Date: 2026-07-29

## Decision

The retained post-spawn v12 state passes the bounded correctness and release
gate. Retain no additional production change.

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
`/var/tmp/able-v12-correctness-release-20260729`:

- Go toolchain: Go 1.26.5;
- fixture typechecking: strict;
- explicit disk-backed `GOCACHE`, `GOTMPDIR`, and `TMPDIR`;
- canonical stdlib: sibling `../able-stdlib`;
- canonical stdlib source tree:
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`;
- canonical stdlib Git head:
  `219eff222c28406487231713753641bc49ee5b9a`;
- unrelated host workloads remained active and were not paused or modified.

The stdlib source-state capture exactly matched
`2026-07-29-post-spawn-context-stdlib-source-state.json`. The stdlib worktree
was already dirty and this tranche made no change there.

## Complete v12 gate

Command:

```sh
GOTOOLCHAIN=go1.26.5 \
GOCACHE=/var/tmp/able-v12-correctness-release-20260729/gocache \
GOTMPDIR=/var/tmp/able-v12-correctness-release-20260729/gotmp \
TMPDIR=/var/tmp/able-v12-correctness-release-20260729/tmp \
PYTHONDONTWRITEBYTECODE=1 \
./run_all_tests.sh
```

Result:

| Measure | Result |
| --- | ---: |
| Exit status | 0 |
| Wall time | 15:50.94 |
| User time | 2,143.07s |
| System time | 132.99s |
| CPU utilization | 239% |
| Peak RSS | 4,651,676 KB |
| Swaps | 0 |
| Compiler tests listed | 835 |
| Compiler short-mode batches | 34 / 34 passed |
| Final bytecode fixture pass | 83.886s, passed |

The pre-Go contract layer passed:

- execution coverage index;
- authoritative scorecard and five-sample evidence;
- feature-to-application coverage;
- benchmark selection, execution, and refresh contracts;
- external threshold controls;
- generated-artifact cleanup policy; and
- canonical/embedded kernel synchronization.

Every non-compiler package passed. The short tree-walker interpreter package
aggregate took 97.915 seconds, and the final complete bytecode fixture pass
took 83.886 seconds.

## Individual-test timing audit

Three 25-test compiler aggregates exceeded one minute under unrelated host
load:

| Batch | Aggregate |
| ---: | ---: |
| 20 | 109.930s |
| 29 | 75.451s |
| 30 | 123.490s |

They were replayed with the same toolchain, fixture mode, and disk-backed
caches using `go test -short -json`. Every selected test passed or explicitly
skipped in short mode, and no individual test approached one minute:

| Batch | Passed | Short-mode skips | Slowest individual test | Elapsed |
| ---: | ---: | ---: | --- | ---: |
| 20 | 21 | 4 | `TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete` | 22.910s |
| 29 | 20 | 5 | `TestCompilerCanonicalSpecNullableExpectationLowers` | 15.150s |
| 30 | 25 | 0 | `TestCompilerNoFallbacksStringBuilderUsesNativeArrayPushAll` | 10.740s |

The aggregate durations therefore do not violate the project rule that one
test must not take more than one minute.

## Canonical stdlib gate

Command:

```sh
GOTOOLCHAIN=go1.26.5 \
GOCACHE=/var/tmp/able-v12-correctness-release-20260729/gocache \
GOTMPDIR=/var/tmp/able-v12-correctness-release-20260729/gotmp \
TMPDIR=/var/tmp/able-v12-correctness-release-20260729/tmp \
PYTHONDONTWRITEBYTECODE=1 \
./run_stdlib_tests.sh
```

Result:

| Mode | Result | Elapsed |
| --- | --- | ---: |
| Tree-walker | passed | 19s |
| Bytecode | passed | 15s |
| Complete command | passed | 35.40s |

The command used 871,740 KB peak RSS and recorded no swaps.

## Derived evidence checks

Post-suite checks also passed:

- `git diff --check`;
- authoritative external scorecard;
- 126-row frontier with zero actionable groups; and
- 23-entry closure ledger with zero invalidations.

No implementation fix was required. The only changes from this tranche are
this record and the forward-looking handoff/log updates.

## Next recommendation

Refresh the exact retained/deferred worktree release inventory without
staging, committing, or pushing.

Why: the correctness gate is green, but the worktree contains many retained
compiler/evidence paths accumulated after the previous inventory alongside
the separately deferred WASM boundary. A stale inventory is now the largest
release risk.

What it entails: capture fully expanded porcelain, HEAD/origin/index
identities, classify every changed file as retained, deferred, or generated;
validate JSON, gzip, formatting, syntax, source-size, secret, and path
complement checks; and write an exact non-mutating manifest. Any staging,
commit, or push still requires separate explicit authorization.

Why it is important: an exact boundary protects the verified native-carrier,
interpreter-free, scorecard, and correctness state from accidental inclusion
of deferred WASM or disposable machine artifacts.
