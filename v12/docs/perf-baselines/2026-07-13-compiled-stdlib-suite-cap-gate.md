# Compiled Stdlib Suite Cap Gate

Date: 2026-07-13

## Purpose

Expose the release-style `cmd/able` compiled-stdlib integration lane one named
case at a time. The cap is a test-health guard, not a benchmark score: a case
that does not finish within one minute is recorded as a compiler-selection
signal, not counted as a passed performance result.

## Test isolation

The test-only stdlib resolver now prefers the adjacent canonical
`able-stdlib` checkout. Build tests use a temporary `ABLE_HOME` override as
well. Those changes prevent an empty developer cache from triggering automatic
git bootstrap; production `able` cache/bootstrap behavior is unchanged.

## Bounded command

```sh
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 GOPROXY=off GOSUMDB=off \
  GOCACHE="$(pwd)/.gocache" \
  go test ./cmd/able -run '^TestTestCommandCompiledRunsStdlib…$' \
  -count=1 -timeout=60s
```

## Results

| Named case | Result | Relevant timeout state |
| --- | --- | --- |
| BigintAndBiguintSuites | cap at 60s | `refreshNativeInterfaceAdapters` → type-expression normalization |
| ExtendedNumericSuites | cap at 60s | waiting for generated Go-build subprocess |
| NumbersNumericSuite | cap at 60s | timeout; confirms numeric failure is not only the grouped suite |
| FoundationalSuites | cap at 60s | `refreshNativeInterfaceAdapters` → concrete interface/impl binding shapes |
| CollectionsListVectorSuites | cap at 60s | timeout; independent collection-family confirmation |

The remaining named cases were deliberately not run in this selection gate:
three unrelated semantic families already exceeded the cap, and further
timeouts would not determine which compiler operation is responsible.

A minimal compiled CLI smoke (`TestTestCommandCompiledRuns`) also exceeded the
cap at one Go runtime thread while waiting for its generated Go-build
subprocess. With `GOMAXPROCS=2` and the same memory/offline limits it passed in
71.57 seconds. The cap therefore measures release-lane throughput, including
generated Go compilation; it is not an Able application-runtime benchmark.

## Decision

No compiler, VM, runtime, canonical-stdlib, or benchmark change is justified
yet. `refreshNativeInterfaceAdapters(...)` is a shared parent in numeric and
foundational stacks, but the concrete descendants differ and end-to-end timing
also includes generated Go build cost. This is test-infrastructure evidence,
not an Able application-runtime benchmark. Keep generated-Go CLI coverage in
an explicit release lane, restore fast ordinary verification, and resume
runtime optimization only from repeated application profiles.

## Release invocation

Generated-Go CLI execution tests are skipped by ordinary `go test` runs. Run
them deliberately with:

```sh
./v12/run_all_tests.sh --compiled-cli
```

That command sets `ABLE_RUN_COMPILED_CLI_INTEGRATION=1`. Fast `--dry-run`
discovery and invalid-configuration tests remain in the ordinary package path.
