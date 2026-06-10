# Bytecode three-unlike semantic-owner refresh

Date: 2026-07-24

## Decision

Retain no VM, runtime, compiler, tree-walker, canonical-stdlib, benchmark,
fixture, language, dependency, or WASM change from this tranche.

Fresh warmed main-only CPU and allocation evidence across Policy Record
Dispatch, Mandelbrot, and Concurrent Audio Voices exposes no new exact
semantic owner that is material in all three applications. The only four
symbols at or above 1% flat CPU in every merged profile are the aggregate
dispatcher, two already-closed stack-carrier parents, and a Go GC scan leaf.
There is no allocation site at or above 1% flat allocation objects in all
three.

No candidate reached the implementation gate, so no A/B experiment was
started. The detailed machine-readable record is
`2026-07-24-bytecode-three-unlike-semantic-owner-refresh.json`.

## Workload selection

The applications are current verifier-backed bytecode misses with complete
Python and Ruby references and deliberately unlike semantics:

| Application | Family | Main semantics |
| --- | --- | --- |
| Policy Record Dispatch | regex/text/map/nominal | file input, regex captures, Result, HashMap, unions, callbacks |
| Mandelbrot | float/byte Array | scalar float loops, fused conditions, u8 output |
| Concurrent Audio Voices | concurrency/interface/callback | Futures, interface calls, closures, nominal state, i64 arithmetic |

This cohort does not treat three regex programs, three numeric programs, or
three concurrency programs as unlike evidence.

## Frozen artifacts and execution contract

The current dirty worktree was preserved. Both measurement artifacts were
built once from repository commit
`237406eccdfb025a519d898daedadee1c8d13a7b` with Go 1.26.4:

| Artifact | SHA-256 |
| --- | --- |
| ordinary `cmd/able` CLI | `c43d0e475fba52be3df1001607df300239858156f77558c91e8a822a91a99026` |
| `pkg/interpreter` benchmark binary | `5069b6dff944d7e68aeb38fb9b85dab990b4d29c842a6cfed04fe66897cb01ab` |

The canonical external stdlib was the dirty retained 70-file Able source tree
at commit `219eff222c28406487231713753641bc49ee5b9a`, with deterministic source
hash `7fd5c9baf427f10ff612f592b5a7820a7f93abeba6959725175a0c7748b2fb4e`.
Source, verifier, input, and foreign-reference hashes are recorded in the JSON
artifact.

All Go processes used:

- CPU affinity 6 after the declared quiet-host preflight passed;
- one logical CPU and `GOMAXPROCS=1`;
- `GOMEMLIMIT=1GiB` and `GOGC=50`;
- canonical external `able-stdlib`;
- source-root-only loading and normal typechecking;
- a 59-second per-process cap;
- the serial executor for Policy and Mandelbrot;
- the goroutine executor for Concurrent Audio Voices.

The runtime benchmark loaded and typechecked once, warmed `main()` once,
forced GC, then measured one main call. CPU, exact allocation counters, and
sampled allocation profiles were collected in separate processes. Three
ordinary frozen-CLI smoke runs passed the public verifiers before profiling.

## Matched end-to-end timing

Five interleaved Able, Python 3.14.5, and Ruby 4.0.5 processes ran on the same
CPU. All 45 processes passed the exact public verifier. These are current
selection measurements, not a partial promotion of the checked-in full
scoreboard.

| Application | Able samples (s) | Able mean | Python mean / ratio | Ruby mean / ratio |
| --- | --- | ---: | ---: | ---: |
| Policy Record Dispatch | 7.32, 6.77, 6.64, 6.73, 6.94 | 6.880 | 0.012 / 573.333x | 0.040 / 172.000x |
| Mandelbrot | 6.28, 6.08, 6.14, 6.11, 6.38 | 6.198 | 1.172 / 5.288x | 1.860 / 3.332x |
| Concurrent Audio Voices | 1.20, 1.22, 1.27, 1.29, 1.25 | 1.246 | 0.112 / 11.125x | 0.104 / 11.981x |

Every row remains far outside the required `1.052632x` maximum against both
interpreters.

## Main-only CPU intersection

Three independent profiles per application merged to 19.08 seconds of Policy
samples, 17.26 seconds of Mandelbrot samples, and 3.34 seconds of Audio
samples.

The complete exact flat-symbol intersection at or above 1% in every
application is:

| Exact symbol | Policy | Mandelbrot | Audio | Disposition |
| --- | ---: | ---: | ---: | --- |
| `(*bytecodeVM).runResumable` | 15.62% | 17.44% | 6.59% | Aggregate opcode dispatcher, not one semantic operation |
| `(*bytecodeVM).appendSlotStackValueChecked` | 3.25% | 4.06% | 1.80% | Already-closed stack-carrier parent |
| `runtime.tryDeferToSpanScan` | 1.21% | 4.92% | 6.89% | Go collector symptom with different allocation owners |
| `(*bytecodeVM).appendStackValue` | 1.05% | 1.74% | 1.80% | Already-closed stack-carrier parent |

The two stack parents do not hide one shared removable child:

- Policy mixes ordinary value snapshots, Array/member results, return values,
  and typed-pattern values.
- Mandelbrot is dominated by float slot loads and float snapshot/write-barrier
  work.
- Audio uses concurrency-safe immutable raw integer transport. Reusing mutable
  raw cells across Future VMs would recreate the race fixed by the retained
  Concurrent Audio Voices correctness rule.

The same stack parents were already broad in the July 21 regex, float,
wide-numeric, and cross-corpus refreshes and the July 22 six-application clean
CPU refresh. Two general ordering/carrier trials failed their unlike
wall-time guards. The current callers supply no new ownership or coherence
rule that invalidates those closures.

## Exact and sampled allocation

Three independent unprofiled counter processes per application produced:

| Application | Mean bytes/main | Mean allocations/main | Counter span |
| --- | ---: | ---: | --- |
| Policy Record Dispatch | 132,283,224 | 1,402,470 | 405,000 bytes / 30 objects |
| Mandelbrot | 615,168,077 | 76,303,082 | 584 bytes / 9 objects |
| Concurrent Audio Voices | 181,159,312 | 4,103,041 | 320 bytes / 4 objects |

One separate 64 KiB sampled allocation profile per application attributes the
objects:

- Policy is led by positional nominal structs, Array lease/view/capacity
  objects, String builder output, and integer materialization.
- Mandelbrot assigns 99.65% of focused sampled objects and 96.55% of focused
  sampled space to `bytecodeNormalizedRawFloatSlotValue`. That is the
  application-specific fused float loop and the already-rejected raw-float
  carrier family.
- Audio is led by callee environments and their concurrency runtime data,
  immutable raw-i64 stack/slot/result transport, integer materialization, and
  positional nominal structs.

Policy and Audio share positional structs and boxed integers, but Mandelbrot
does not. Mandelbrot's float objects do not occur materially in the other two.
There is therefore no exact allocation owner at or above 1% in all three.
The apparent GC intersection cannot authorize a pooling, carrier, or lifetime
change.

## Candidate admission

No candidate passes all required gates:

- `runResumable` is a dispatcher parent, not a semantic rule.
- stack append is an already-rejected carrier family whose concrete values
  split across ordinary, float, and concurrent integer ownership;
- GC scanning is caused by different semantic allocations;
- Policy's regex/Array/nominal work, Mandelbrot's float loop, and Audio's
  environments/integer transport are each material in only one or two rows.

Consequently the tranche keeps no code and does not manufacture an A/B
candidate from a parent symbol.

## Verification

The frozen CLI smoke run for each application passed its public verifier.
All nine CPU processes, nine exact-allocation processes, and three sampled
allocation processes completed under the declared cap. The unchanged broad
bytecode family also passes:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 27.069s
```

Raw profiles, outputs, and timing files are cleanup-eligible under
`/tmp/able-bytecode-three-unlike-20260724.3pIhq6`.

## Next recommendation

Reconcile the active bytecode VM v2 roadmap against the complete current
closure ledger before another implementation tranche.

Why: the roadmap still lists typed registers, quickened dispatch, native
Array/String opcodes, and compact frames as candidate directions, but the
current three-unlike gate and the earlier full-corpus architecture budgets
admit none as a present shared owner. Selecting one directly would repeat a
rejected representation or one-family experiment.

What it entails: inventory each active VM v2 direction against the current
scoreboard, semantic-work amplification reports, exact CPU/allocation owners,
and recorded A/B rejections; remove or explicitly defer disproven items; then
model any surviving cohesive primary-VM change across at least three unlike
applications. Admit a prototype only if it preserves one semantic authority
and models at least 25% target-excess reduction in every governing row.

Why it is important: Able remains 3.3x-573x behind the governing interpreter
in this cohort. Local helper changes cannot close those budgets, while an
unreconciled architecture queue risks repeating months of negative trials.
This design gate makes the next implementation decision explicit and
evidence-backed. Continue to defer WASM.
