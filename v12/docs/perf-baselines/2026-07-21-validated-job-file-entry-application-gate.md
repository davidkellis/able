# Validated job file-entry application gate — 2026-07-21

## Decision

Retain the file-driven evolution of `validated_job_pipeline`, its source-equivalent
Go/Python/Ruby references, verifier and input contract, feature memberships,
repeated measurements, and bounded profiles. Retain no compiler, generated-runtime,
bytecode VM, canonical-stdlib, language, or WASM change.

The evolved application closes the four remaining depth-two portable feature
triples, raising minimum three-family interaction depth to three. The compiled
profile repeats the already-closed goroutine-identity owner. The bytecode profile
has no new concrete VM leaf with enough CPU weight and breadth to admit a generic
candidate, so the 89-row performance frontier remains zero actionable.

## Application contract

The application reads 32 job labels from `jobs.txt`, replays them for 64 rounds,
and sends 2,048 nominal `JobTask` values through four workers. File-derived label
text affects validation, recovery, and the successful recurrence; the workload
therefore cannot be reduced to an unused input fixture. It retains captured
`Result` mapping/unwrapping callables and `Error` interface dispatch. Its
schedule-independent result is:

```text
2048:2048:2048:1540:508:775362991:850772
```

Able tree-walker, Able bytecode, Able compiled, Go 1.26, Python 3.14, and Ruby
4.0 produce that output through the same verifier. The catalog declares
`jobs.txt` as a real program argument and records its SHA-256 fingerprint in
every retained comparison row.

An initial 512-round scale was rejected before retained measurement because the
tree-walker exceeded the bounded correctness window. The 64-round scale keeps
the tree-walker near 6.7 seconds and bytecode near 0.3–0.5 seconds, comfortably
below the one-minute project test cap.

## Feature-interaction result

Adding honest file input and program-entry handling to the existing application
raises all four former depth-two triples to depth three:

- concurrency × expressions/files × closures/callables;
- concurrency × expressions/files × interfaces/dispatch;
- concurrency × closures/callables × program entry;
- concurrency × interfaces/dispatch × program entry.

The full weighted result is:

| Measure | Reconstructed baseline | Current |
| --- | ---: | ---: |
| portable families | 11 | 11 |
| three-family interactions | 165 | 165 |
| zero-depth triples | 0 | 0 |
| minimum depth | 1 | 3 |
| depth-one triples | 8 | 0 |
| improved triples | — | 158 |

No synthetic syntax or benchmark-only language feature was added. The change
uses existing `able.fs`, `able.os.args`, nominal values, interfaces, callables,
`Result`, channels, and futures in an application-shaped flow.

## Repeated measurements

The retained exact-contract evidence comprises two independent five-process
cohorts per lane. All successful samples are retained; no outlier was removed.

| Lane | Processes | Pooled mean | CV | Limiting ratio |
| --- | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 1.184000 s | 10.23% | 245.49× Go |
| Go | 10 | 0.00482309 s | 9.11% | — |
| Able bytecode | 10 | 0.463000 s | 17.31% | 14.61× Python / 7.97× Ruby |
| Python | 10 | 0.0316964 s | 17.46% | — |
| Ruby | 10 | 0.0581155 s | 19.40% | — |

The checked scoreboard uses the second normal five-process cohort: compiled
1.130 seconds versus Go 0.0048 seconds (235.417×), and bytecode 0.414 seconds
versus Python 0.0368 seconds (11.250×) and Ruby 0.0622 seconds (6.656×). The
pooled values are the better workstation estimate; both cohorts have the same
target-miss classification.

Two earlier setup cohorts were intentionally excluded. They preceded the
catalog correction that made `jobs.txt` an explicit, fingerprinted execution
contract, so they are not eligible scorecard evidence even though their output
verified.

## Ownership and admission

Three verified compiled main-phase profiles merge to 3.08 seconds of CPU
samples. `bridge.currentGID` is 94.16% cumulative and `runtime.traceback2` is
84.42% cumulative. This independently repeats the exact `runtime.Stack`-based
goroutine identity owner across unlike concurrency applications. The prior
generic fixed-context alternative failed an unrelated N-Body guard, so this
tranche does not revive it without a design that avoids cost outside spawned
call graphs.

Three verified bytecode profiles merge to 940 ms of CPU samples. The largest
flat item is `runtime.cgocall` at 20.21%, dominated by parsing/tree-sitter
boundaries. VM work is dispersed: the largest named VM leaves are
`lookupCachedMemberMethodEntry` and `bytecodeRawIntegerValueInfo`, each only
2.13% flat; Go map access is 5.32% cumulative. No new child is both material
and shown in three unlike workload families. A workload-specific fast path is
therefore inadmissible.

## Verification

- exact output parity across both interpreters, compiled Able, and all three
  reference languages;
- two verifier-backed five-process cohorts per timed lane;
- explicit verifier, argument, input-file, and source fingerprints;
- three verified compiled and three verified bytecode profiles;
- complete catalog, feature-coverage, pair- and triple-interaction checks;
- selection, scoreboard, performance-frontier, and evidence-ledger checks;
- source files remain below 1,000 lines;
- `git diff --check`.

## Next recommendation

Prototype an explicitly scoped generated execution context for compiled
`spawn`/Future call graphs, then gate it across at least three unlike concurrency
applications plus unrelated N-Body, text/map, and iterator/control programs.

Why: repeated profiles now place 94.16% of this application's compiled CPU under
the same `bridge.currentGID` / `runtime.Stack` owner already seen across the
concurrency suite. That is a large, generic language-runtime wall, while the
new bytecode profile is too diffuse to support a responsible candidate.

What it entails: thread a context token only through generated functions proven
reachable from a spawn boundary; preserve the existing bridge fallback at
dynamic/host boundaries; add semantic tests for nested spawn, await, cancellation,
and environment restoration; then use repeated verifier-backed cohorts and
average every successful workstation sample. Keep the change only if concurrency
improves without regressing the unrelated guards. Do not begin WASM work.
