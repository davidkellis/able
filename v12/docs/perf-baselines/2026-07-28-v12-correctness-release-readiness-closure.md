# V12 correctness and release-readiness closure

Date: 2026-07-28

## Decision

The current v12 release surface is green. Retain no production code for this
tranche.

The audit found one genuine release-evidence failure: the cross-engine
structural-strategy snapshot still pinned the predecessor closure-ledger
fingerprint. The current ledger itself was healthy at 21 closures and zero
invalidations. The repair updated the current source identities and
mechanically regenerated the dependent structural strategy, portable-backend
ADR, shared-runtime semantic-ABI feasibility, and closed-region decision
artifacts. Their decisions did not change, and no prototype became admissible.

No compiler, generated runtime, runtime, interpreter, bytecode VM, canonical
stdlib, language, dependency, benchmark, fixture, reference application, or
WASM behavior changed.

## Release gates

All build-heavy work used disk-backed `TMPDIR=/var/tmp` and
`GOCACHE=/var/tmp/able-go-cache`.

| Gate | Result | Elapsed | Peak RSS |
| --- | --- | ---: | ---: |
| `go vet ./...` | pass | 8.73 s | 554,804 KB |
| `go build ./...` | pass | 2.55 s | 378,808 KB |
| `./run_all_tests.sh` | pass | 880.18 s | 4,440,504 KB |
| `./run_all_tests.sh --compiler` | pass | 4,947 s | 2,782,396 KB |
| cold `./run_all_tests.sh --compiled-cli` | pass | 1,276.93 s | 3,312,312 KB |
| warm JSON-event compiled-CLI replay | pass | 184.74 s | 2,733,408 KB |
| `./run_stdlib_tests.sh` | pass | 33.62 s | 840,004 KB |

The ordinary handoff passed every deterministic contract, every non-compiler
package, 34 compiler batches, and the 85.938-second bytecode fixture corpus.

The separate compiler matrix passed:

- 33 compiler core batches;
- three explicit parity outliers;
- all 24 fallback-audit shards;
- all 24 compiled-execution shards;
- all 24 strict-dispatch shards;
- all 24 interface-lookup shards; and
- all 24 boundary-marker shards.

That is 120 green generated-program audit shards. Together they verify output,
fallback classification, strict generated dispatch, static interface-lookup
bypass, and absence of fallback markers.

Canonical stdlib tests passed in both reference modes: tree-walker in 17
seconds and bytecode in 16 seconds.

## Individual-test timing

Eight compiler aggregates crossed one minute during the noisy full runs.
Exact `go test -json` event timing found these individual maxima:

| Lane | Batch | Longest individual test |
| --- | ---: | ---: |
| short | 5 | 3.41 s |
| short | 20 | 34.47 s |
| short | 29 | 15.13 s |
| short | 30 | 14.52 s |
| full core | 7 | 2.54 s |
| full core | 9 | 23.05 s |
| full core | 19 | 34.94 s |
| full core | 29 | 10.44 s |

The cold compiled-CLI package aggregate took 1,271.838 seconds, while its warm
JSON replay put the longest individual integration test at 31.09 seconds.

Two single-shard samples narrowly crossed one minute under sustained pressure
from the 82-minute compiler matrix. Neither reproduced:

- interface-lookup shard 8/24: 19.38, 19.54, and 19.98 seconds; 19.63-second
  mean;
- boundary shard 22/24: 17.72, 17.50, and 18.45 seconds; 17.89-second mean.

No individual test violates the one-minute project limit, so no coverage or
sharding change is justified.

## Evidence repair

The repaired chain is:

| Artifact | SHA-256 |
| --- | --- |
| performance closure ledger | `f00e9fafd2ad222c018de0748093957687bfd555f3bceb711825ef23191c3ffc` |
| cross-engine structural strategy | `69679874a7d9666a32411bb126c726349cdfeff692055d6c0d13bf53b5d1f78e` |
| portable VM backend ADR | `af7eb2539bb16be719af1ead3eb276b8f912e788a15de84371b30436c3d57f2e` |
| shared-runtime semantic-ABI feasibility | `b7d4ffb2e26f4f88dc317bd456e7f84d7e865d902f661f80ebcd0a867960acfb` |
| shared-runtime closed-region decision | `83e9792145e78d48bd55e53ecae68c51f4965666302db1e2b14217616f00dea5` |

The final deterministic pass confirms:

- 63 portable applications and 119 selected rows: 63 compiled and 56
  bytecode;
- complete five-sample scorecard evidence;
- 21 current closures and zero invalidations;
- zero actionable frontier groups;
- all architecture-budget and threshold-control checks;
- `git diff --check`; and
- no maintained Go source file at or above 1,000 lines.

The v12 specification remained unchanged with SHA-256
`4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.

## Cache and temporary-file cleanup

The cold compiled-CLI run raised the explicit disk-backed compiled-test cache
to 94 valid entries and 3,464,331,864 bytes. Checksum-aware LRU pruning removed
52 entries and 1,878,969,544 bytes, leaving 42 valid entries and
1,585,362,320 bytes below the 1536 MiB ceiling.

There were no corrupt, staging, or obsolete-schema entries. The cache tool
classified 258 entries and 190,478,898 bytes as unknown and preserved them;
no caller-owned data was deleted.

The disposable release workspace under `/var/tmp` contains only logs and
event traces needed to write this record and will be deleted after final hash
verification.

## Recommendation

Keep production performance mutation paused and next refresh the non-mutating
release-consolidation inventory for work retained after the prior July 27
inventory.

Why: the current execution and evidence gates are green and no shared
performance owner is admitted, while the long-running dirty worktree has
accumulated substantial retained July 27–28 work.

What it entails: map current changed and untracked v12 paths to their dated
records and dependency-ordered review boundaries. Identify unmatched or
generated-local paths without deleting them. Do not reset, commit, rewrite
history, or touch deferred WASM work.

Why it is important: this turns the verified native-lowering and
interpreter-free state into an auditable release candidate without inventing
a narrow optimization or risking the existing dirty worktree.
