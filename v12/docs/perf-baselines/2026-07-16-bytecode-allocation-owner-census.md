# Bytecode coverage-wide allocation-owner census

Date: 2026-07-16

## Decision

Keep no runtime, compiler, stdlib, or benchmark code from this tranche. Twenty-five
ordinary warmed runs produced valid bounded allocation profiles and exact benchmark
allocation counters. K-Nucleotide produced four consistent exact counter rows, but
its allocation-profile process did not exit inside the 55-second guard, so its
partial profiles are excluded from symbol clustering.

After process-initialization noise and previously closed allocation families are
removed, no one exact Able allocation owner remains material in at least three
unlike applications. The widest new-looking Go symbol, `copyCallArgs`, reconciles
to the benchmark `print` observer in I-Before-E and a related group of concurrency
natives in the other five applications. It is not one Able owner, and the
non-borrowing native contract requires stable argument storage. No candidate meets
the generality and semantic-safety bar.

## Method

The cohort is the bytecode section of `bench-selection-manifest.json`, matching the
preceding 26-application scorecard and CPU-symbol census. Each normal profile used
the ordinary loader and typechecker, canonical `../able-stdlib`, one warm call and
one measured call, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Applications
ran sequentially in separate processes. Sampling rates were selected per workload
to keep the profile overhead bounded; rates ranged from 1 for low-allocation
programs to 8192 for the highest-allocation numeric programs.

The exact measured allocation rows were:

| Application | Sample rate | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Array Slice Window | 64 | 14,158,600 | 422,250 |
| Await Channel Mux | 32 | 9,942,328 | 166,653 |
| Base64 | 1 | 2,201,617,824 | 584 |
| Channel Rollup | 64 | 22,951,704 | 303,287 |
| Dependency Plan | 4 | 1,462,576 | 11,970 |
| Document Audit | 1 | 624,928 | 911 |
| Fib | 1 | 38,384 | 39 |
| Fixed Width 128 | 4096 | 1,242,236,544 | 30,858,345 |
| Future Await Race | 8 | 1,286,592 | 34,847 |
| Future Pipeline | 128 | 14,437,816 | 688,824 |
| I-Before-E | 1 | 9,118,000 | 2,072 |
| JSON | 1 | 114,800,448 | 223 |
| Lexical Rollup | 4 | 2,443,400 | 14,919 |
| Mandelbrot | 8192 | 618,887,032 | 76,303,147 |
| Matrix Multiply | 4096 | 308,601,904 | 14,032,566 |
| Monte Carlo Pi | 4096 | 177,825,992 | 22,222,137 |
| Mutex Await Journal | 32 | 8,933,792 | 191,170 |
| Mutex Ledger | 64 | 17,332,560 | 436,411 |
| Option/Result Configuration | 256 | 76,362,576 | 1,305,012 |
| PiDigits | 256 | 336,063,032 | 1,013,276 |
| Rational Series | 256 | 129,951,880 | 1,405,666 |
| Regex Set Audit | 256 | 120,812,240 | 1,734,090 |
| Regex Stream Audit | 256 | 123,060,112 | 1,908,333 |
| Reverse Complement | 4096 | 705,411,536 | 10,894,483 |
| Word Frequency | 128 | 47,555,528 | 631,435 |

K-Nucleotide cannot fit its warm call, measured call, profile finalization, and
process shutdown under the cap. A temporary first-call diagnostic binary was tried
at rates 4096, 65536, and 1048576, followed by an exact-only run. The four measured
rows were stable at a mean 1,369,758,536 B/op and 24,014,661 allocs/op, but all four
processes exceeded the exit guard. The source switch was reverted immediately, and
the partial profiles were not admitted to the census.

For the 25 valid profiles, `go tool pprof` supplied normalized flat
`alloc_objects` and `alloc_space` percentages. A symbol had to account for at least
1% flat in an application and recur in three applications before reconciliation.
Standard Go memory profiles include allocations made before the benchmark timer is
reset, so process-init owners such as `initBytecodeSmallIntBoxCache`, `init.func1`,
and `strings.genSplit` are profile contamination; the benchmark `B/op` and
`allocs/op` rows remain the exact measured-call counters.

## Reconciled clusters

The main exact allocation-object clusters, with K-Nucleotide excluded, are:

| Exact symbol | Applications | Sum flat | Decision |
| --- | ---: | ---: | --- |
| `bytecodeBoxedIntegerValue` | 13 | 95.43% | closed raw-boxing family |
| `bytecodeRawIntegerResultValue` | 8 | 221.26% | closed raw-result family |
| Array lease cleanup | 9 | 59.71% | closed primitive-Array shell/lease family |
| `newEnvironmentBase` | 8 | 81.16% | closed call-environment family |
| `copyCallArgs` | 6 | 74.36% | different owners after caller reconciliation |
| positional struct result construction | 5 | 86.65% | closed positional-result family |
| `bytecodeStackSnapshotValue` | 4 | 30.39% | closed stack-carrier family |
| `bytecodeNormalizedRawFloatSlotValue` | 3 | 216.21% | exact substitution already failed broad wall-time gates |
| `copyInlineCallArgToSlot` | 3 | 21.16% | raw-carrier call-frame snapshot family |
| `ArrayEnsureCapacity` | 3 | 10.74% | related Array/regex growth; prior growth trials rejected |

Allocation-space clustering reaches the same conclusion. Raw float normalization
accounts for 198.90 summed flat percentage across Mandelbrot, Matrix Multiply, and
Monte Carlo Pi, but this exact substitution and several related owned-cell/direct
setter designs previously reduced allocation while regressing wall time. Repeating
that family would optimize the profiler metric rather than application performance.
Array capacity, environment, raw integer, positional-result, and primitive-Array
owners are likewise documented closed families.

`copyCallArgs` received caller-tree reconciliation because it was the only remaining
symbol with apparent unlike reach. All six samples arise in the exact-native path.
I-Before-E's 1,628 sampled copies come from the benchmark `print` observer. Await
Channel Mux, Channel Rollup, Future Pipeline, Mutex Await Journal, and Mutex Ledger
come from the related channel/future/mutex native family. A native marked
`BorrowArgs == false` is allowed to retain its argument slice, so replacing the
copy with shared scratch storage is not a semantics-preserving generic change.
Auditing individual concurrency natives could remove copies for that one related
feature family, but it does not satisfy this tranche's three-unlike-application
admission rule.

## Verification and cleanup

The ordinary runtime benchmark source contains no first-call/skip-warmup diagnostic
switch. No production or benchmark behavior changed, so no candidate-specific test
suite was required; the restored benchmark configuration receives its focused Go
test before handoff. All temporary binaries, partial profiles, top reports, and
aggregation tables are removed after this record is complete.

## Next recommendation

Refresh the selected-program compiled scorecard against the equivalent Go programs,
then profile the largest compiler misses that repeat across unlike applications.

Why: two coverage-wide bytecode censuses now show that the remaining shared CPU and
allocation symbols are parents, process noise, semantically required boundaries,
or families already rejected by application wall time. Continuing to mine those
same bytecode leaves risks optimizing profiler metrics. The compiler still has the
separate product target of at least 95% of equivalent Go performance and should now
receive the same current, verifier-backed evidence.

What it entails: run each selected compiled Able and Go pair in separate processes,
use at least five verified workstation repetitions and arithmetic means, preserve
output hashes, and rank both ratio and absolute time. For the largest misses, collect
bounded CPU and allocation profiles of the generated binaries and require one
general lowering/runtime owner across at least three unlike programs before building
a candidate. Any change must use primitive lowering or the shared nominal pipeline,
never a benchmark, stdlib-container, regex, or source-shape special case. Do not
begin WASM work.
