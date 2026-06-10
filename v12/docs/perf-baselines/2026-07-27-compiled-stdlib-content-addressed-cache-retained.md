# Compiled stdlib content-addressed cache retained

Date: 2026-07-27

## Decision

Retain an opt-in content-addressed executable cache for `able test
--compiled`.

The cache makes the complete 32-case compiled stdlib lane practical for
repeated validation without changing Able compilation, generated execution,
test discovery, assertions, or production lowering. A cold cache still
performs the complete Able compile and Go build. A hit skips only those two
steps; the cached executable still runs the selected tests and reporters on
every invocation.

No production compiler, generated runtime, runtime, interpreter, bytecode VM,
language, canonical stdlib, dependency, benchmark, or WASM change was
required.

## Interface

The cache is disabled unless the caller sets:

```text
ABLE_TEST_COMPILED_CACHE_DIR=/disk/backed/path
```

Use a disk-backed path such as `/var/tmp/able-compiled-test-cache`, not the
RAM-backed `/tmp` on this workstation.

Two diagnostic controls are available:

- `ABLE_TEST_COMPILED_CACHE_TRACE=/path/to/trace` appends `hit KEY` or
  `miss KEY` after each successful lookup or publication.
- `ABLE_TEST_COMPILED_CACHE_SALT=VALUE` deliberately creates a different key
  and is useful for invalidation tests or manual cache epochs.

The cache does not yet prune old entries. The caller owns the configured
directory and may remove it when its entries are no longer useful.

## Cache contract

Each key includes:

- the cache schema;
- every loaded Able module name, ordinary import, dynamic import, source path,
  and source byte;
- included test-package names and ordered search-root kind/provenance;
- the generated Able runner and Go reporter harness;
- all compiler options used by the compiled-test lane;
- the Go module root;
- the selected Go executable bytes and the relevant `go env` identity
  (`GOOS`, `GOARCH`, `GOVERSION`, `CGO_ENABLED`, `GOFLAGS`, `GOEXPERIMENT`,
  `GOTOOLCHAIN`, `CC`, and `CXX`);
- `go.mod`, `go.sum`, non-test Go/C/header/assembly build sources under
  `cmd/` and `pkg/`, and the embedded kernel inputs; and
- the optional caller-provided salt.

Source paths are deliberately part of the key. This preserves real-source
diagnostic ownership and prevents an artifact built for one relocated source
tree from reporting locations from another.

The synthetic runner is the sole exception. Its unique parent workspace now
contains a stable `compiled-test-runner` child, so its qualified Able package
does not change on every invocation. Cache-enabled builds normalize only the
synthetic runner origin to `compiled-test-runner/runner.able`, build in a
private stable `able/compiled` Go module, and use `-trimpath`. Real Able source
origins remain unchanged.

Entries contain an executable and a JSON manifest with the schema, key, and
executable SHA-256. Lookup verifies the manifest, executable mode, and checksum
before returning a hit. Publication copies and syncs into a same-filesystem
staging directory and then renames atomically. A concurrent valid winner is
reused; a corrupt exact-key entry is rejected and repaired.

## Focused correctness guards

The new focused tests prove:

- identical semantic inputs produce identical keys;
- Able dependency-source changes invalidate the key;
- harness/reporter configuration changes invalidate the key;
- compiler-option, build-identity, salt, and source-path changes invalidate
  the key;
- the cache is explicitly opt-in;
- repeat publication reuses a verified artifact;
- checksum corruption becomes a miss and is repaired; and
- only the synthetic runner origin is normalized, while real user origins are
  preserved.

Commands:

```text
go test ./cmd/able -run 'TestCompiledTestCache|TestNormalizeCompiledTestRunnerOrigins' -count=1
go test ./cmd/able -count=1
```

Both passed. The complete `cmd/able` package took 39.898 seconds in the final
focused handoff.

## Three-unlike-family evidence

An initial three-run cold/warm selection matrix used a distinct salt for each
round and preserved every existing expected-output assertion:

| Case | Cold wall mean | Warm wall mean | Reduction |
| --- | ---: | ---: | ---: |
| Vector | 55.937 s | 1.607 s | 97.13% |
| HashSet | 45.790 s | 1.617 s | 96.47% |
| LazySeq | 39.847 s | 1.663 s | 95.83% |

The matrix trace contained exactly nine misses and nine hits. Mean peak RSS
fell 89.60%-90.16%.

After the final stable-module and trim-path correction, a fresh salt for each
family proved configuration invalidation on the exact retained code:

| Case | Salted miss wall | Immediate hit wall |
| --- | ---: | ---: |
| Vector | 39.52 s | 1.59 s |
| HashSet | 28.77 s | 1.50 s |
| LazySeq | 24.69 s | 1.48 s |

The final trace contained exactly three misses and three hits. The lower final
miss times reflect a hot shared Go build cache; the trace is the authoritative
miss/hit evidence.

## Complete 32-case lane

Final cold population:

- all 32 cases passed;
- trace: 32 misses, zero hits;
- wall time: 1,045.16 seconds;
- peak RSS: 3,265,372 KB;
- worst individual case: Vector at 58.18 seconds;
- every individual case remained below one minute; and
- the retained uncached reference was 1,038.379 seconds, so cache-enabled cold
  overhead was 0.65%, within normal workstation noise.

All 32 cached binaries were then scanned:

- zero contained an `able-test-work-*` path; and
- all 32 contained the stable `compiled-test-runner/runner.able` origin.

Three complete warm repetitions passed:

| Repetition | Wall | Peak RSS |
| --- | ---: | ---: |
| 1 | 22.39 s | 314,572 KB |
| 2 | 21.85 s | 315,200 KB |
| 3 | 22.97 s | 317,972 KB |
| Mean | 22.403 s | 315,915 KB |

The trace contained exactly 96 hits and zero misses. Relative to the final
cold population, the warm mean is:

- 97.86% lower wall time;
- 46.65 times faster; and
- 90.33% lower peak RSS.

The 32 verified executables occupy 1.2 GiB on disk.

## Full verification

`./run_all_tests.sh` passed using disk-backed `/var/tmp` scratch space:

- coverage, external scoreboard, selection, and threshold checks passed;
- all non-compiler Go packages passed;
- all 33 bounded compiler batches passed;
- the final bytecode fixture pass completed in 86.259 seconds; and
- total wall time was 696.74 seconds.

No canonical stdlib source changed. No production generated-code ownership or
compiler/interpreter boundary changed, so the frozen application performance
profiles remain authoritative.

## Next recommendation

Keep production performance mutation paused and add a bounded lifecycle for
the compiled-test cache before considering default use in
`run_all_tests.sh --compiled-cli`.

Why: one useful 32-case identity occupies 1.2 GiB, and compiler/source changes
will naturally leave older content-addressed identities behind.

What it entails: add explicit inspect/prune controls with byte and age limits;
distinguish current-schema valid entries from corrupt or obsolete entries;
prove that pruning cannot delete an artifact being published or executed; and
then measure the canonical script with an explicitly configured disk-backed
cache. Do not add hidden background deletion or a RAM-backed default.

Why it is important: bounded disk use preserves the 46.65-times faster routine
validation without recreating the stale temporary-file pressure this tranche
was intended to avoid. Production lowering remains paused until a new
three-unlike-application owner invalidates the closed performance frontier.
Do not begin WASM work.
