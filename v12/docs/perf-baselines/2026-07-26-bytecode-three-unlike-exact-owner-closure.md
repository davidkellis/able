# Bytecode three-unlike exact-owner closure

Date: 2026-07-26

## Decision

Retain no VM, runtime, compiler, tree-walker, canonical-stdlib, benchmark,
fixture, language, dependency, or WASM change from this tranche.

Fresh current-artifact CPU and allocation evidence across Word Frequency, RMS
Norm, and Concurrent Signal Dispatch exposes no new concrete non-nominal owner
that is material in all three applications. The only exact CPU symbols at or
above 1% flat in every merged profile are the aggregate dispatcher, an
already-closed slot-to-stack carrier parent, and a Go collector scan leaf.
There is no allocation site at or above 1% in all three.

No candidate reached the implementation gate, so no A/B experiment was
started. The machine-readable companion is
`2026-07-26-bytecode-three-unlike-exact-owner-closure.json`.

## Workload selection

The applications are current verifier-backed bytecode misses with complete
Python and Ruby references and deliberately different dominant semantics:

| Application | Family | Main semantics |
| --- | --- | --- |
| Word Frequency | file/text/map/Result | UTF-8 decode, String splitting, generic HashMap counting, typed Result control |
| RMS Norm | float/Array numeric | f64 Array traversal, arithmetic, square root, scalar reduction |
| Concurrent Signal Dispatch | concurrency/interface/nominal/i64 | Futures, Channels, user interface dispatch, positional tasks/results, signed Array work |

The cohort does not treat three text, three numeric, or three concurrency
programs as unlike evidence.

## Frozen artifacts and execution contract

The dirty worktree was preserved. Both measurement artifacts were built once
from repository commit `237406eccdfb025a519d898daedadee1c8d13a7b` with Go
1.26.4:

| Artifact | SHA-256 |
| --- | --- |
| ordinary `cmd/able` CLI | `5f1108bc9596e74dd37e29fdb863bf8fa517e91935fd7db83ceecc940b896666` |
| `pkg/interpreter` benchmark binary | `5069b6dff944d7e68aeb38fb9b85dab990b4d29c842a6cfed04fe66897cb01ab` |

The benchmark binary is byte-for-byte identical to the July 24 three-unlike
refresh artifact. The intervening retained changes were compiler-only and do
not invalidate the VM/runtime closure ledger.

The canonical external stdlib was the dirty 70-file Able source tree at
commit `219eff222c28406487231713753641bc49ee5b9a`. Source, verifier,
declared-input, and foreign-reference hashes are recorded in the JSON
companion.

Every process used CPU 6, one logical CPU, `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, source-root-only loading, normal typechecking,
the canonical external stdlib, and a 59-second cap. Word Frequency and RMS
Norm used the serial executor. Concurrent Signal Dispatch used the goroutine
executor required by its language semantics.

The runtime benchmark loaded and typechecked once, warmed `main()` once,
forced GC, and measured one main call. CPU, exact allocation counters, and
64-KiB sampled allocation profiles ran in separate processes. Five ordinary
frozen-CLI launches per Able/Python/Ruby lane passed the public verifiers.

## Matched end-to-end timing

Five interleaved processes per lane all passed the public verifier:

| Application | Able samples (s) | Able mean | Python mean / ratio | Ruby mean / ratio |
| --- | --- | ---: | ---: | ---: |
| Word Frequency | 1.93, 1.28, 1.22, 1.66, 1.37 | 1.492 | 0.012 / 124.333x | 0.054 / 27.630x |
| RMS Norm | 4.36, 4.13, 4.39, 4.65, 4.59 | 4.424 | 0.854 / 5.180x | 0.532 / 8.316x |
| Concurrent Signal Dispatch | 1.47, 1.37, 1.34, 1.38, 1.91 | 1.494 | 0.052 / 28.731x | 0.070 / 21.343x |

All rows remain far outside the required `1.052632x` maximum against both
interpreters. These comparisons establish current target-miss status; sampled
profile durations are not promoted as timing evidence.

## Exact main allocation counters

Three independent unprofiled measured-main processes per application
produced:

| Application | Mean ns/main | Mean bytes/main | Mean allocations/main |
| --- | ---: | ---: | ---: |
| Word Frequency | 1,074,405,553 | 48,402,851 | 637,264 |
| RMS Norm | 4,237,827,601 | 288,052,888 | 20,000,125 |
| Concurrent Signal Dispatch | 1,182,795,113 | 125,900,952 | 4,584,109 |

Word Frequency's allocation-count span is seven objects, RMS Norm is exact
across all three runs, and Concurrent Signal Dispatch spans 57 objects under
goroutine scheduling.

## Main-only CPU intersection

Three independent profiles per application merged to 3.29 seconds of Word
Frequency samples, 12.81 seconds of RMS Norm samples, and 3.61 seconds of
Signal Dispatch samples.

The complete exact flat-symbol intersection at or above 1% in every
application is:

| Exact symbol | Word | RMS | Signal | Disposition |
| --- | ---: | ---: | ---: | --- |
| `(*bytecodeVM).runResumable` | 7.29% | 8.67% | 6.65% | aggregate opcode dispatcher |
| `runtime.tryDeferToSpanScan` | 3.65% | 2.81% | 4.71% | collector symptom over different allocations |
| `(*bytecodeVM).appendSlotStackValueChecked` | 2.13% | 1.17% | 1.39% | already-closed carrier parent |

The slot-to-stack parent does not hide one common removable child:

- Word Frequency mixes ordinary values, pooled/value-sidecar i32 values,
  typed-pattern observations, Array/member results, and return values.
- RMS Norm's child is `bytecodeStackSnapshotValue`, which severs aliases by
  copying mutable float slot state into a stable operand value.
- Concurrent Signal Dispatch uses concurrency-safe raw-i64 stack/result
  values. Mutable reuse across Future VMs would violate the retained
  concurrency ownership rule.

`runResumable` names the dispatch loop, not one semantic operation. The exact
dispatcher descendants split into map/call/type work, float regions/static
math calls, and Array/interface/concurrency work. A second dispatcher,
register executor, and broad stack-carrier changes have already failed their
unlike-program gates.

## Allocation-owner intersection

The separate sampled profiles identify disjoint concrete leaves:

- Word Frequency is led by positional UTF-8 decode results, String host
  results, Array lease/view/capacity work, and integer materialization.
- RMS Norm is led by raw-float slot normalization, float materialization,
  stable float stack snapshots, unary float results, and the native math
  result path.
- Concurrent Signal Dispatch is led by raw-i64 slots/stack/results,
  monomorphic primitive Array values, async environments and their mutexes,
  and bound interface/member callables.

There is no exact allocation leaf at or above 1% in all three applications.
The apparent collector intersection therefore cannot authorize pooling,
carrier reuse, or GC-policy tuning. Positional nominal construction in Word
Frequency is also an observable nominal value and is absent from RMS Norm and
materially below 1% in Signal Dispatch; it does not reopen the completed
nominal-lifetime route.

## Candidate admission

No candidate passes the active contract:

- the dispatcher is an aggregate parent;
- slot append is a completed carrier family whose concrete values differ;
- GC scanning is caused by different semantic allocation owners;
- Word's text/map/nominal work, RMS's float work, and Signal's
  concurrency/interface/i64 work are each material in only one row.

The correct retained result is therefore no code.

## Verification

- 45/45 interleaved Able/Python/Ruby timing processes passed their exact
  public verifiers.
- 9/9 CPU processes, 9/9 exact-allocation processes, and 3/3 sampled
  allocation processes completed under the declared cap.
- The focused unchanged bytecode family passed:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 36.148s
```

The complete `./run_all_tests.sh` handoff passed: coverage and scoreboard
contracts, every non-compiler package, all 32 bounded compiler batches, and
the complete bytecode fixture pass. The aggregate non-compiler
`pkg/interpreter` run completed in 283.711s and the final bytecode fixture pass
completed in 261.476s; these are package/corpus aggregates, not individual
test durations.

Raw profiles, caches, binaries, and outputs were placed under the disk-backed
`/var/tmp/able-bytecode-three-unlike-20260726.*` workspace and are disposable
after this record.

## Next recommendation

Build a corpus-wide exact-owner frequency census over the current 54-row
rankable bytecode scorecard.

Why: two independent three-unlike cohorts now terminate in aggregate or
already-closed parents, but a hand-selected intersection can miss a smaller
exact leaf repeated across many other applications. The compiled corpus-wide
census found one such general owner after its largest-miss cohort did not.

What it entails: inventory current compatible measured-main CPU and allocation
profiles by frozen interpreter/source/stdlib identity; refresh only missing or
stale rows; normalize exact symbols by concrete semantic caller; rank
non-closed owners by unlike-program reach and total target excess; and admit a
prototype only for one general non-nominal leaf material in at least three
unlike applications. If no owner survives, declare profile-driven local VM
optimization exhausted for the current corpus and require an explicit
architecture or semantic invalidation before more VM code.

Why it is important: bytecode still misses Python and Ruby broadly. A
corpus-wide census is the strongest remaining way to find a general local VM
tax without repeating typed lanes, stack transport, Array, float, register
executor, named-container, benchmark-specific, or GC-policy experiments. Do
not begin WASM work.
