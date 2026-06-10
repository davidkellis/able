# Compiled-test cache bounded lifecycle retained

Date: 2026-07-27

## Decision

Retain explicit inspection and bounded pruning for the opt-in
`able test --compiled` executable cache.

The cache remains disabled unless the caller supplies a disk-backed root
through `ABLE_TEST_COMPILED_CACHE_DIR` or passes `--dir` to a lifecycle
command. There is no default cache location, automatic pruning, background
deletion, or RAM-backed fallback.

No production compiler lowering, generated runtime, runtime, interpreter,
bytecode VM, language, canonical stdlib, dependency, benchmark, or WASM
behavior changed.

## Interface

The CLI now exposes:

```text
able cache compiled-tests inspect \
  [--dir PATH] [--json] [--verbose]

able cache compiled-tests prune \
  [--dir PATH] [--max-bytes SIZE] [--max-age DURATION] \
  [--dry-run] [--json]
```

`--dir` takes precedence over `ABLE_TEST_COMPILED_CACHE_DIR`. A root is
required explicitly through one of those two interfaces.

Byte limits accept raw bytes and decimal or binary suffixes such as `1.5GB`
and `1536MiB`. Age limits accept Go durations and day suffixes such as `24h`
and `7d`. A dry run performs the same checksum inventory and candidate
selection without deleting anything.

Inspection classifies root content as:

- valid current-schema entries;
- corrupt current-schema entries;
- interrupted atomic-publication staging directories;
- obsolete `able-compiled-test-*` schema directories; or
- unknown caller-owned content.

Pruning always selects corrupt, staging, and obsolete-schema entries. It then
selects valid entries older than the explicit age limit and, if necessary,
least-recently-used valid entries until the explicit byte limit is met.
Unknown content is reported but never deleted.

## Concurrency and use contract

Every cache-enabled compiled-test invocation holds a shared operating-system
file lock from before lookup or publication until after the selected cached
executable exits. Inspection also takes a shared lock. Pruning takes an
exclusive nonblocking lock and returns a busy result without deleting
anything when a publisher or executable is active.

Unix uses `flock`; Windows uses `LockFileEx`. Kernel-managed file locks are
released when a process exits, including abnormal termination. Successful
hits refresh the entry-directory modification time, which is the explicit LRU
and age-pruning clock.

The production path continues to verify the manifest and executable SHA-256
before every hit. Lifecycle inspection uses the same validator, so a corrupt
entry cannot be counted as retained valid data.

## Focused guards

Focused tests cover:

- valid, corrupt, staging, obsolete-schema, and unknown classification;
- dry-run and actual byte-bound selection;
- age pruning and hit-driven last-use refresh;
- preservation of unknown root content;
- refusal to remove a target outside the configured root;
- JSON CLI inspection and pruning;
- byte-size and duration parsing; and
- a real helper process holding the shared lock while an exclusive prune
  returns busy and removes zero entries, followed by successful pruning after
  release.

The focused command passed:

```text
go test ./cmd/able \
  -run 'TestCompiledTestCache|TestParseCompiledTestCache|TestRunCache' \
  -count=1
```

The Windows lock implementation also compiled in isolation for
`GOOS=windows GOARCH=amd64 CGO_ENABLED=0`. A full Windows CLI cross-build is
not currently possible because the pre-existing tree-sitter language package
has no Windows build files.

## Canonical disk-backed measurement

All measurements used:

```text
TMPDIR=/var/tmp
GOCACHE=/var/tmp/able-go-cache
ABLE_TEST_COMPILED_CACHE_DIR=/var/tmp/able-compiled-test-cache
./v12/run_all_tests.sh --compiled-cli
```

The starting cache contained 32 valid entries and 1,283,077,424 bytes, with no
corrupt, staging, obsolete, or unknown content.

The lifecycle source change deliberately invalidated the prior build
identity. The complete canonical cold population passed:

- wall time: 950.71 seconds;
- peak RSS: 3,308,200 KB;
- trace: 42 misses and zero hits.

The canonical command contains 32 stable stdlib cases and ten temporary-path
compiled regression cases. Real Able source paths remain part of the cache
key to preserve diagnostic ownership, so the latter ten correctly miss in a
new Go test process.

Three complete warm repetitions passed:

| Repetition | Wall | Peak RSS | Trace |
| --- | ---: | ---: | --- |
| 1 | 181.82 s | 2,739,772 KB | 32 hits, 10 misses |
| 2 | 186.39 s | 2,738,108 KB | 32 hits, 10 misses |
| 3 | 180.79 s | 2,751,396 KB | 32 hits, 10 misses |
| Mean | 183.00 s | 2,743,092 KB | 32 hits, 10 misses |

The full canonical command was 5.20 times faster warm despite rebuilding all
ten process-local cases. This number is intentionally separate from the
previous 22.403-second mean for the 32-case stable stdlib-only lane.

Four accumulated generations occupied 104 valid entries and 3,756,646,712
bytes. The reviewed command:

```text
able cache compiled-tests prune \
  --dir /var/tmp/able-compiled-test-cache \
  --max-bytes 1536MiB
```

first produced a dry-run plan and then removed exactly 62 stale entries and
2,175,938,648 bytes. It retained 42 entries and 1,580,708,064 bytes: one
complete most-recent canonical working set below the 1,610,612,736-byte cap.

A post-prune canonical run passed in 183.21 seconds at 2,822,452 KB peak RSS
with exactly 32 hits and ten expected temporary-path misses. Reapplying the
same bound removed the ten newly obsolete temporary entries
(297,620,408 bytes). Final inspection again found:

- 42 valid entries;
- 1,580,708,064 bytes;
- zero corrupt entries;
- zero staging entries;
- zero obsolete-schema entries; and
- zero unknown entries.

The removed content consisted only of reproducible cached executables and
manifests. It can be regenerated by rerunning the corresponding compiled
tests. The reusable `/var/tmp/able-go-cache` was intentionally retained.

## Full verification

`./run_all_tests.sh` passed with disk-backed `/var/tmp` scratch space and
without enabling the compiled-test cache:

- total wall time: 692.07 seconds;
- peak RSS: 4,674,440 KB;
- `cmd/able`, including the new lifecycle guards: 47.589 seconds;
- every non-compiler package passed;
- all 33 bounded compiler batches passed; and
- the final bytecode fixture pass completed in 86.265 seconds.

No canonical stdlib source changed. The v12 spec remained byte-for-byte
unchanged.

## Next recommendation

Keep the compiled-test cache opt-in and explicitly disk-backed. Use inspection
and a reviewed byte or age bound in routine release validation; do not add an
implicit `/tmp` default or hidden automatic deletion.

Why: the stable 32-case cohort remains highly reusable, but the canonical
command also creates ten path-specific entries per process and one current
working set occupies 1.47 GiB. Explicit lifecycle control preserves the speed
benefit without unbounded disk growth.

What comes next: keep production performance mutation paused until the
authoritative frontier is invalidated by a new broad application, a
generated-execution change, a correctness failure, or a report-only
observation of one exact non-closed owner in at least three unlike
applications. When that evidence exists, refresh only the affected compiled
profiles and advance one general native-carrier or boundary rule through
verifier-backed repeated A/B/Go-reference measurements.

Why it is important: this keeps the project aimed at Go-equivalent generated
execution without manufacturing a benchmark-specific optimization or
reopening already-closed boundary routes. WASM remains deferred.
