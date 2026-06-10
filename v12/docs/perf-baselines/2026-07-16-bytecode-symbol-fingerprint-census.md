# Bytecode coverage-wide CPU symbol fingerprint census

Date: 2026-07-16

## Decision

Keep no runtime, compiler, or stdlib code. A normalized CPU-symbol census now
covers every one of the 26 reviewed portable bytecode applications, but after
removing shared parents and documented rejected families it does not identify
a new concrete leaf with enough material headroom in three unlike programs.

The largest surviving exact VM leaf is canonical Array slot-call proof
validation. It is already a 2.24 ns allocation-free direct lookup, and its
complete receiver/runtime-data/version/cache prefix is 9.50 ns. The required
checks preserve dynamic runtime data, global-definition invalidation, and
method-cache invalidation. No further slot-call candidate was built from a
1-2% local leaf.

## Method

The cohort is the bytecode section of `bench-selection-manifest.json`, the same
26 applications used by the preceding five-process scorecard. Each profile
used the ordinary loader, typechecker, canonical `../able-stdlib`, source-root
isolation where required, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.
Applications ran sequentially in separate test processes.

After one warm call, fixed measured-call counts targeted roughly 1-6 seconds
of CPU samples while keeping every process below 60 seconds. Three short
programs were rescaled from their observed warmed cost: Document Audit used
200 calls, Future Await Race 300, and Fib 750,000. K-Nucleotide cannot fit its
roughly 38-second warm call plus a profiled call under the mandatory cap. A
temporary test-harness-only switch therefore profiled its first `main` call
after load/typecheck in 35.96 seconds. That switch was removed from source
immediately after the diagnostic binary was built.

All 26 final profiles contain CPU samples:

| Application | Calls | CPU samples |
| --- | ---: | ---: |
| Array Slice Window | 6 | 2.99 s |
| Await Channel Mux | 20 | 1.14 s |
| Base64 | 2 | 4.55 s |
| Channel Rollup | 8 | 1.47 s |
| Dependency Plan | 8 | 1.98 s |
| Document Audit | 200 | 2.29 s |
| Fib | 750,000 | 1.61 s |
| Fixed Width 128 | 1 | 6.67 s |
| Future Await Race | 300 | 2.81 s |
| Future Pipeline | 8 | 2.34 s |
| I-Before-E | 8 | 1.81 s |
| JSON | 5 | 2.21 s |
| K-Nucleotide | 1 first call | 35.89 s |
| Lexical Rollup | 10 | 1.22 s |
| Mandelbrot | 1 | 5.69 s |
| Matrix Multiply | 1 | 4.09 s |
| Monte Carlo Pi | 2 | 4.34 s |
| Mutex Await Journal | 20 | 1.45 s |
| Mutex Ledger | 10 | 1.78 s |
| Option/Result Configuration | 5 | 3.91 s |
| PiDigits | 2 | 3.98 s |
| Rational Series | 1 | 3.43 s |
| Regex Set Audit | 1 | 4.21 s |
| Regex Stream Audit | 1 | 3.64 s |
| Reverse Complement | 1 | 6.43 s |
| Word Frequency | 3 | 3.55 s |

`go tool pprof -top -nodecount=0` supplied exact fully qualified symbols. The
census grouped a symbol only when its flat weight was at least 1% in an
application, then required at least three applications. Cumulative-only shared
frames were retained for reconciliation but could not qualify as concrete
leaves.

## Coverage-wide clusters

The largest exact clusters are:

| Exact symbol/family | Applications at >=1% flat | Sum flat weight | Decision |
| --- | ---: | ---: | --- |
| `bytecodeVM.runResumable` | 24 | 194.11% | VM parent, not a semantic operation |
| `runtime.tryDeferToSpanScan` | 23 | 95.35% | Go allocation/GC scanning; callers diverge |
| `bytecodeRawIntegerValueInfo` | 18 | 79.77% | already rejected by repeated broad raw-integer gates |
| `internal/runtime/maps.ctrlGroup.matchH2` | 15 | 58.54% | ArrayStore, environment/cache, and language-map ownership diverge |
| `sync/atomic.(*Int32).Add` | 13 | 59.08% | Array leases, executor/channel, and GC paths diverge |
| `appendSlotStackValueChecked` | 13 | 31.17% | two generic ordering/carrier trials already failed broad guards |
| `finishInlineReturn` | 13 | 20.15% | multiple return-guard trials already failed broad guards |
| `popCallFrameFields` | 10 | 22.37% | call-frame parent with unlike consumers |
| `execBinary` | 9 | 17.69% | arithmetic parent; concrete operators/types diverge |
| `lookupCachedMemberMethodEntry` | 6 | 21.75% | member policy just completed its dependency/admission gates |

The census also found several smaller apparent leaves:

- `ensureFitsInt64Type` is above 1% in Future Pipeline, Mutex Await Journal,
  and Mutex Ledger, but those applications repeat the same authored LCG/checksum
  modulo pattern. Fib's larger percentage is only a 2.1 microsecond `main` and
  has negligible absolute product cost. The raw-store caller was already
  consolidated in an earlier retained change.
- `evaluateDivModFast` is material in Await Channel Mux, Future Await Race,
  Mutex Await Journal, and Mutex Ledger. Their sources deliberately repeat the
  same `48271`/`1_000_003` checksum recurrence, so they are not unlike evidence
  for this leaf. Fixed Width, Option/Result, and Rational use it below the
  material threshold.
- `memberMethodLexicalStateHeader` clears 1% in Dependency Plan, Document
  Audit, and I-Before-E, but it is precisely the validation family whose
  threshold and same-parent shortcuts were just rejected.
- `bytecodeDirectIntegerCompare` clears 1% in six applications, but it is the
  already-specialized raw comparison family with different caller shapes.

## Canonical Array slot proof follow-up

After the rejected/parent filters, the highest remaining exact VM leaf was
`lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions`:

| Application | Flat | Cumulative |
| --- | ---: | ---: |
| Array Slice Window | 1.67% | 2.01% |
| Regex Set Audit | 1.19% | 3.33% |
| Regex Stream Audit | 1.10% | 2.20% |
| Reverse Complement | 1.09% | 1.40% |
| Matrix Multiply | 0.73% | 0.73% |

Regex Set and Stream count as one related implementation family; Array Slice,
Reverse Complement, and Matrix Multiply provide the unlike controls. The leaf
has small but real reach, so it received focused evidence rather than an
automatic rejection.

Array Slice Window's stats-enabled measured call made 888,266 Array slot-call
lookups: 888,257 cache hits, nine cache misses, no receiver miss, and no
fast-path failure. Thus the sampled cost is overwhelmingly the successful
proof path, not recoverable failed probing.

The existing allocation-free microbenchmarks were repeated five times for two
seconds per row:

| Microbenchmark | Mean | Range | Allocations |
| --- | ---: | ---: | ---: |
| Direct cache lookup with precomputed versions | 2.243 ns | 2.179-2.385 ns | 0 |
| Full cached prefix with runtime/global/method versions | 9.502 ns | 9.062-9.882 ns | 0 |

The direct entry performs an indexed read and exact program, IP, fast-path,
global-revision, and method-version comparisons. The remaining prefix verifies
that no dynamic runtime data changes member semantics and obtains the current
global/method revisions. Removing those checks would make the optimization
incorrect; moving or duplicating them would add state to every VM/instruction
for at most the measured 1-2% local leaf. The previous quicksort work already
closed further slot-call dispatch shaving in favor of larger walls. No
candidate was justified.

## Verification and cleanup

The benchmark catalog remains complete. The temporary K-Nucleotide harness
switch is absent from source, and the focused runtime-benchmark configuration
test passes on the restored harness. No production or test behavior changed.
All census binaries, profiles, top reports, stats outputs, and microbenchmark
artifacts are removed after this record.

## Next recommendation

Run a coverage-wide allocation-owner fingerprint census for the same selected
bytecode cohort, using exact benchmark allocation counters plus bounded sampled
allocation profiles.

Why: the widest unresolved CPU symbols are Go allocation/GC scanning and map
internals. CPU stacks show that those runtime leaves recur, but not which Able
operation owns the allocations or maps. A cross-application allocation-owner
census can distinguish a real shared Able allocation boundary from unrelated
ArrayStore, environment, nominal-result, string, and language-map work. This
is the remaining evidence needed before touching GC-facing runtime structure.

What it entails: run one measured warmed call per application in its own
process, use exact `B/op` and `allocs/op` from the benchmark, and choose a
sampling rate that keeps each profile below 60 seconds. Normalize flat
`alloc_objects` and `alloc_space` by application, group exact Able allocation
owners appearing materially in at least three unlike programs, and remove the
already-closed positional-result, primitive-Array shell, raw boxing, lambda,
member-cache, and call-environment candidates. Only the highest surviving
owner receives lifetime/consumer tracing and a generic trial. Any retained
candidate still needs repeated verifier-backed workstation A/B averages. Do
not begin WASM work or add benchmark, stdlib-type, nominal-container, regex,
or source-shape special cases.
