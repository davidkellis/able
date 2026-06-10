# Compiled post-ABI current profile refresh — 2026-07-22

## Decision

Retain no production change. Fresh main-only CPU, exact allocation-counter,
and allocation-owner evidence across eight unlike compiled applications does
not identify a new open generated/runtime leaf in at least three programs.
The union-heavy cohort repeats only mechanisms that have already completed a
broad gate, while the immediate non-union expansion separates into numeric,
map/text, and concurrency owners.

No compiler, bridge, runtime, canonical-stdlib, bytecode, language, benchmark,
or WASM behavior changed.

## Artifact identity correction

The first diagnostic pass used the final lazy-environment ABI artifacts. Its
profiles unexpectedly contained the old allocating `fmt.Sprintf`-based
`structCacheKey`, proving those binaries predated the subsequently retained
typed struct-definition cache key. All 59 otherwise verified processes from
that pass were excluded before selection.

The governing binaries are the preserved post-key artifacts under
`v12/tmp/struct-definition-key-candidate-rebuild-20260722/artifacts`. They
contain the retained direct/lazy environment ABI and comparable
`(environment pointer, name)` cache key. Their hashes and sizes are recorded in
the companion JSON. Current profiles contain no `structCacheKey` allocation.

## Protocol

Every governing process used logical CPU 0, `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, a 55-second cap, and its public Ruby verifier.
The tranche retained 139 successful governing processes with zero benchmark
failure or timeout:

- 20 independent ordinary timing processes for the four union-heavy owners;
- 50 main-only CPU profiles for those owners;
- 21 main-only CPU profiles for four unlike non-union misses;
- 40 lightweight exact main-allocation counter processes, five per program;
- eight allocation-owner processes.

K-Nucleotide's one-object end profile could not serialize within the cap. The
normal program had produced its output, but that diagnostic was excluded as a
timeout. It was replaced by a verified 64 KiB sampled cumulative allocation
profile plus five main-only exact counter processes. No test or accepted
measurement was allowed to run beyond one minute.

## Repeated current timings

These are arithmetic means over five independent verified processes, not
single-sample claims.

| Application | Samples | Mean | User | System | Mean GC |
| --- | --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 0.58, 0.51, 0.53, 0.54, 0.55 s | 0.542 s | 0.486 s | 0.038 s | 15.4 |
| Option/Result Config | 0.18, 0.18, 0.18, 0.17, 0.18 s | 0.178 s | 0.138 s | 0.030 s | 5.0 |
| Manifest Normalization | 0.19, 0.18, 0.18, 0.18, 0.18 s | 0.182 s | 0.142 s | 0.028 s | 5.0 |
| Policy Record Dispatch | 0.19, 0.21, 0.22, 0.23, 0.24 s | 0.218 s | 0.180 s | 0.030 s | 7.8 |

## Main-only CPU classification

| Application | Processes | Samples | Governing classification |
| --- | ---: | ---: | --- |
| Binary Event Log | 5 | 2.49 s | allocation/GC, `bridge.ToInt`, EventRecord conversion |
| Option/Result Config | 15 | 1.60 s | allocation/GC, already-rejected static match-expression construction |
| Manifest Normalization | 15 | 1.60 s | allocation/GC, String/byte conversion |
| Policy Record Dispatch | 15 | 1.75 s | allocation/GC, regex NFA bodies |
| N-Body | 10 | 0.77 s | direct generated floating-point advance/sqrt bodies |
| K-Nucleotide | 3 | 8.34 s | primitive map equality/hash and allocation/GC |
| Matrix Multiply | 3 | 3.20 s | direct generated matrix multiplication body |
| Mutex Ledger | 5 | 2.36 s | closed `currentGID`/`runtime.Stack` environment tracking |

The common `runtime.mallocgc` and GC frames are parents, not a shared compiler
mechanism. Caller reconciliation splits them among nominal conversion, union
conversion, String/bytes, regex storage, primitive map boxing, matrix backing,
and concurrency bookkeeping.

## Exact main allocation counters

Five independent processes per row produced the following arithmetic means.
The companion JSON retains every count.

| Application | Mean objects | Mean bytes | Stability |
| --- | ---: | ---: | --- |
| Binary Event Log | 3,457,793.0 | 244,031,243.2 | objects identical |
| Option/Result Config | 1,052,543.6 | 37,768,635.2 | one-object range |
| Manifest Normalization | 917,512.0 | 43,453,324.8 | 86-object range |
| Policy Record Dispatch | 921,369.8 | 47,323,872.0 | three-object range |
| N-Body | 41.0 | 1,096.0 | identical |
| Matrix Multiply | 8,017.6 | 32,897,305.6 | two-object range |
| Mutex Ledger | 147,664.4 | 8,270,760.0 | one-object range |
| K-Nucleotide | 27,734,581.0 | 1,224,424,377.6 | objects identical |

## Why no candidate was admitted

- `NewIdentifier` / `NewSimpleTypeExpression` repeats across the union-heavy
  cohort, but the immutable static expression candidate already removed those
  objects and then failed the broad wall gate. It remains reverted.
- `bridge.ToInt` is broad, but the retained dynamic-`i64` helper already covers
  the proven escaping common-value boundary. A global suffix-preserving cache
  removed allocations and still failed an unrelated averaged guard. Remaining
  ordinary boxes are observable dynamic/nominal boundary values.
- `bridge.ToUint` is large in K-Nucleotide and present in Manifest/Policy, but
  the completed eight-application unsigned census found material reach only in
  K-Nucleotide. Its hot callers are primitive HashMap operations, so extending
  it would effectively optimize one named-container workload.
- K-Nucleotide's primitive map hashing/equality has no matching owner in the
  other seven programs and cannot authorize a `HashMap` lowering branch.
- Mutex Ledger exactly reproduces the closed goroutine-ID/environment owner;
  the fixed execution-context alternatives already failed unlike controls.
- N-Body and Matrix Multiply spend their CPU in direct generated arithmetic
  bodies. They expose no general bridge, dispatch, or allocation helper to
  remove.

This satisfies the handoff's instruction to broaden immediately when the
union-heavy cohort had no new leaf. The broadened cohort also terminates
without a candidate, so no speculative A/B implementation was built.

## Verification

All 139 governing benchmark/profile processes passed their public verifier.
The focused bridge cache tests pass after the profile work, and no source file
was changed outside documentation/handoff records. The current bridge modules
remain below 1,000 lines (`bridge.go` 902; focused struct-definition module
157).

## Next recommendation

Refresh and reconcile the verifier-backed compiled-versus-Go scorecard using
the retained current binaries before selecting another compiler candidate.

Why: this eight-application screen has no open concrete leaf, while some
checked-in performance-frontier dispositions still describe pre-ABI profiles
or candidates that have since been retained, rejected, and reverted. Choosing
from those stale rankings risks repeating closed work.

This entails freezing current compiled binaries for the portable suite,
running repeated averaged Able and Go cohorts under the same verifier and
resource contract, updating target ratios and evidence dispositions, and then
profiling only a materially missing group whose same concrete generic
descendant appears in at least three unlike applications. Continue to forbid
named nominal/container/application lowerings and continue to defer WASM.
