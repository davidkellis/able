# Compiler/AOT release-readiness no-code closure

Date: 2026-07-27

## Decision

The current v12 compiler/AOT state passes the selected release-readiness
surface. Retain no production code for this tranche.

The preceding generated-artifact audit found exact current/frozen machine-code
parity for all eight recorded hot functions across Tapelang Alphabet, Sudoku
Masks, and Fib. With no changed execution owner or native-carrier boundary to
optimize, this tranche tested shippability rather than admitting another
speculative lowering change.

No compiler, generated runtime, runtime, interpreter, VM, canonical stdlib,
language, dependency, benchmark, fixture, reference application, or WASM
behavior changed.

## Gates

All commands used the existing disk-backed Go cache at
`/var/tmp/able-go-cache`. Compile-heavy commands also set
`TMPDIR=/var/tmp`.

| Gate | Result | Elapsed | Peak RSS |
| --- | --- | ---: | ---: |
| `go vet ./...` | pass | 1.62 s | 238,944 KB |
| `go build ./...` | pass | 2.81 s | 383,456 KB |
| `./v12/run_all_tests.sh --compiled-cli` | pass | 181.72 s | 2,736,252 KB |
| `./run_all_tests.sh` | pass | 637.05 s | 4,655,052 KB |
| `./run_stdlib_tests.sh` | pass | 31.26 s | 853,928 KB |

The complete handoff passed:

- coverage, scorecard, selection-integrity, and threshold-control checks;
- every non-compiler Go package;
- all 33 bounded compiler batches;
- the final bytecode fixture pass in 88.039 seconds;
- canonical stdlib tests in both tree-walker and bytecode modes, each reported
  at 15 seconds.

The v12 specification remained unchanged during the audit. Its SHA-256 was
`4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.

## Individual-test timing check

Two aggregate compiler batches exceeded one minute, so their top-level tests
were timed from `go test -json` events rather than treating package duration as
individual-test duration:

| Batch | Full-suite package time | Isolated package time | Longest individual test |
| --- | ---: | ---: | ---: |
| 19 | 132.800 s | 129.10 s | 33.35 s |
| 29 | 68.617 s | 67.32 s | 10.58 s |

Batch 19's longest case was
`TestCompilerTypedArrayDefaultMethodsKeepConcreteReceivers`. Batch 29's was
`TestCompilerNoFallbacksStringBuilderUsesNativeArrayPushAll`. No individual
test violated the one-minute project limit.

## Compiled-test cache lifecycle

The explicit compiled-CLI lane produced the expected 42 trace events:

- 32 stable cache hits;
- 10 temporary-path-sensitive misses.

The misses temporarily raised the cache to 1,878,328,432 bytes. The retained
LRU pruning command removed exactly 10 entries and 297,620,408 bytes, leaving
42 valid entries and 1,580,708,024 bytes below the 1536 MiB ceiling. No active
user work or unrelated temporary files were removed.

The disposable compiled trace, canonical-stdlib log, and the stdlib runner's
temporary build workspace were removed. The bounded compiled-test cache and
the reusable disk-backed Go build cache were retained.

## Interpretation

The audited compiler state is green for the current v12 release surface:
static compiled applications retain their native-carrier and
interpreter-free guarantees, while dynamic/bootstrap behavior remains covered
by the existing runtime paths. The audit exposed neither a correctness
failure nor a new general production owner material in three unlike
applications, so a production prototype and repeated A/B/Go cohort were not
justified.

## Recommendation

Keep production performance mutation paused.

Next prepare a non-mutating release-consolidation inventory of the very dirty
v12 worktree, grouped by the dated retained records, without resetting,
rewriting history, or creating commits unless a maintainer explicitly
authorizes it.

Why: the execution and release gates are green, but the long-running worktree
must be made reviewable before those verified results can become a reliable
release candidate.

What it entails: map changed and untracked v12 files to their retained
tranches, identify unmatched or stale artifacts without deleting them, and
propose safe review/commit boundaries. If new performance evidence appears
instead, refresh only the affected profile or boundary census before
considering code.

Why it is important: this preserves every existing change while turning the
verified compiler state into auditable, releasable work and prevents another
speculative optimization from weakening the general Go-native lowering
architecture.
