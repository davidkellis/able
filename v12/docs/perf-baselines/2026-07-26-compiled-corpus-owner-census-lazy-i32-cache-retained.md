# Compiled corpus owner census and lazy common-`i32` bridge cache

Date: 2026-07-26

## Decision

Complete the 61-row strict compiled exact-owner census and retain one general
primitive bridge optimization. Common `i32` values crossing from generated
native Go into `runtime.Value` now initialize their immutable cached boxes
independently. A process pays for only the values it reaches instead of
allocating all 4,224 entries on its first in-range conversion.

This is a primitive compiler-boundary rule. It does not inspect an application,
benchmark, container name, or non-primitive nominal type, and it does not
change generated Go. The cache remains in `pkg/compiler/bridge`, physically
separate from the bytecode VM's raw-`i32` cache. Each slot preserves the
existing concrete `runtime.IntegerValue` dynamic type and uses `sync.Once` for
concurrent first access.

No canonical-stdlib, general runtime, interpreter, language, dependency, or
WASM change is needed.

## Current corpus refresh

The refresh built all 61 current coverage applications with
`--no-fallbacks`. Every smoke process passed its public Ruby verifier. All 61
compiler copies had SHA-256
`6ea26d6ee1e9b6b5447371bb01adc8052738ec08b13177db28b972b8c30c7bcc`,
and every final Go dependency graph omitted
`able/interpreter-go/pkg/interpreter`.

Adaptive repeated CPU profiling reused each exact executable and accumulated
at least one second of smoke-user-time per row, bounded from three through
fifty processes. The refresh retained:

- 61 verified build-smoke application processes;
- 1,639 verified main-only CPU-profile processes;
- 183 verified lightweight main-allocation processes; and
- 61 nonempty exact-binary merged CPU profiles.

All 1,883 processes passed their public verifier and stdout-hash check.
Profiles were merged only within one exact executable. The allocation pass ran
three processes per application with `GOMEMLIMIT=1GiB`, `GOGC=50`, and the
catalog's serial or goroutine executor policy.

## Exact-owner census

The breadth rule admitted only exact compiler/generated-runtime owners that
were material in at least three unlike target misses and absent or
non-material in suitable controls. Ranking used material application count
and aggregate scorecard excess. Broad parents and one-sample short-row ratios
were not selection evidence.

The leading repeated families were:

| Exact owner or family | Material misses | Decision |
| --- | ---: | --- |
| checked signed multiply | 11 | Already-closed primitive arithmetic |
| signed div/mod | 9 | Already-closed primitive arithmetic |
| `bridge.currentGID` / environment / channel / await | 3-8 | Already-rejected execution-context ABI family |
| `bridge.ToInt` | 7 | Admitted for exact allocation attribution |
| generated hash lookup/set/equality | 3 | Named-container family; excluded |

Other generated owners were application-specific or failed the three-unlike
threshold. After the admitted `ToInt` cache cost, the census contains no
additional non-closed shared compiled owner.

## Candidate admission

Main-phase exact allocation profiles found the same
`bridge.initCommonI32Boxes` owner creating exactly 4,225 objects in six unlike
applications:

| Application | Baseline main allocations | Eager-cache objects | Candidate initialized slots |
| --- | ---: | ---: | ---: |
| Sensor Calibration | 57,534 | 4,225 | 3 |
| concurrent Policy Callbacks | 46,938 | 4,225 | 7 |
| concurrent Tree Folds | 28,992 | 4,225 | 1 |
| concurrent Stencil Reduction | 29,731 | 4,225 | 35 |
| concurrent Stateful Pipeline | 196,998 | 4,225 | 5 |
| TapeLang Alphabet | 4,274 | 4,225 | 1 |

This is distinct from the rejected 263,168-entry bytecode raw-`i32` cache
experiments. Those trials changed or resized the VM's central operand cache
and harmed hot dispatch or GC pacing. The compiled bridge already performed a
`sync.Once` fast-path check on every cached conversion; the retained design
moves that check to the selected slot without changing the interface dynamic
type or adding a second lookup layer.

New guards prove that one request initializes one slot, cached reuse remains
allocation-free, range and suffix fallbacks are unchanged, and concurrent
first access is race-free.

## Repeated allocation result

Three lightweight main-phase processes per state produced:

| Application | Baseline bytes / allocations | Candidate bytes / allocations | Byte change | Allocation change |
| --- | ---: | ---: | ---: | ---: |
| Sensor Calibration | 3,610,317 / 57,536 | 3,333,973 / 53,312 | -7.65% | -7.34% |
| concurrent Policy Callbacks | 2,979,955 / 45,703 | 2,700,451 / 41,434 | -9.38% | -9.34% |
| concurrent Tree Folds | 1,210,629 / 29,006 | 929,261 / 24,764 | -23.24% | -14.62% |
| concurrent Stencil Reduction | 1,813,469 / 29,058 | 1,538,936 / 24,866 | -15.14% | -14.43% |
| concurrent Stateful Pipeline | 10,880,744 / 196,989 | 10,611,136 / 192,790 | -2.48% | -2.13% |
| TapeLang Alphabet | 282,592 / 4,272 | 6,112 / 45 | -97.84% | -98.95% |
| K-Nucleotide | 614,265,056 / 16,232,604 | 614,185,104 / 16,232,471 | -0.01% | -0.001% |
| Fib | 144 / 6 | 144 / 6 | neutral | neutral |
| Binary Trees | 9,820,593,061 / 613,771,226 | 9,820,316,899 / 613,767,026 | -0.003% | -0.001% |

K-Nucleotide is the hot-reuse control: it reaches almost the entire common
range, so lazy initialization is intentionally neutral rather than turning
millions of cache hits back into boxes. Fib proves zero-reach neutrality.
Binary Trees retains essentially identical application allocation despite
avoiding the fixed cache.

## Repeated A/B/Go timing

Every timed process used the catalog CPU budget and affinity plus
`GOMEMLIMIT=1GiB` and `GOGC=50`. Order rotated among preserved baseline,
candidate, and freshly built Go 1.26.4. The public Ruby verifier ran after
every timed process. Short applications received 130-150 samples per variant;
long applications received eight or ten.

| Application | Samples per variant | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sensor Calibration | 130 | 0.011706s | 0.012036s | +2.82%, neutral interval | 0.003534s | 3.406x |
| concurrent Policy Callbacks | 150 | 0.010111s | 0.009518s | -5.87% | 0.002742s | 3.471x |
| concurrent Tree Folds | 150 | 0.005861s | 0.005857s | -0.08%, neutral | 0.002471s | 2.371x |
| concurrent Stencil Reduction | 150 | 0.010572s | 0.010268s | -2.87% | 0.002909s | 3.529x |
| concurrent Stateful Pipeline | 130 | 0.020502s | 0.019886s | -3.01% | 0.003302s | 6.022x |
| TapeLang Alphabet | 8 | 3.880099s | 3.707224s | -4.46%, noisy interval | 1.952433s | 1.899x |
| K-Nucleotide | 10 | 1.606295s | 1.560417s | -2.86%, noisy interval | 0.057821s | 26.987x |
| Fib | 8 | 3.563191s | 3.325890s | -6.66% noise/layout control | 3.225784s | 1.031x |
| Binary Trees | 8 | 10.284289s | 10.401889s | +1.14%, neutral interval | 10.773678s | 0.966x |

Sensor's paired 95% mean-difference interval is
`[-0.0839ms, +0.7436ms]`; Tree Folds is
`[-0.1644ms, +0.1553ms]`; Binary Trees is
`[-83.3347ms, +318.5360ms]`. They are treated as neutral. Fib never reaches
the cache and therefore establishes that favorable long-row movement is
workstation or binary-layout noise, not a claimed cache speedup.

All 2,232 A/B/Go timed processes verified. Across corpus refresh, exact
profiles, candidate smoke/stats, and timing, this tranche completed 4,163
verified application processes.

The candidate's final executable-size changes range from -232 bytes through
+3,920 bytes across the nine measured applications, at most 0.03%. Generated
source outside the linked local Able module is byte-identical in every A/B
pair.

## Verification and known baseline failure

Passed:

- full `pkg/compiler/bridge` tests;
- focused bridge race tests;
- generic interface-boundary, Option/Result specialization,
  Iterator/filter-map execution, and strict Vector compiler guards;
- `go test ./cmd/ablec`;
- nine candidate `--no-fallbacks` dependency audits;
- nine generated-source equivalence audits; and
- `git diff --check` for the retained files.

`TestCompilerHashMapNativeCarrierExecutes` currently fails with
`hash map value missing handle`. A narrow source A/B restored only the eager
cache and reproduced the identical failure, proving it is baseline-equal and
not caused by this candidate. The failure remains a correctness blocker and
must be repaired before another performance tranche.

## Cleanup

All measurements used the disk-backed
`/var/tmp/able-corpus-owner-census-20260726` workspace. Compiler copies,
generated local module copies, Go caches, allocation profiles, CPU profiles,
binaries, and per-process output are deleted after this aggregate record is
written. No stale `able-*` directory from this tranche remains on tmpfs.

## Next recommendation

Repair `TestCompilerHashMapNativeCarrierExecutes` before starting more
performance work.

Why: the failure is baseline-equal, so it does not invalidate the retained
primitive cache, but the roadmap requires correctness failures to take
priority. A generated native map handle is being lost or rejected at a shared
nominal/dynamic boundary.

What it entails: reproduce the failing generated module, trace the handle
through the general nominal encoding and bridge conversion path, and fix the
shared carrier/identity rule without a `HashMap` or other named-container
compiler branch. Then rerun the exact test, generic dynamic-boundary and
nominal-interface guards, strict K-Nucleotide/Inventory/Word Frequency
applications, and `go test ./cmd/ablec`.

Why it is important: a fast compiler cannot be considered closed while a
native carrier can disappear across a legal boundary. Once that correctness
gate is green, begin typed bytecode `i32` slot storage. The compiled census has
closed its one newly admissible shared owner and the remaining repeated owners
are already-closed families; typed slots are the next general opportunity
toward the bytecode 95% target. Do not begin WASM work.
