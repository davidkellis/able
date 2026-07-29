# Bytecode current-source profile coverage closure

Date: 2026-07-28

## Decision

Retain no VM, compiler, runtime, stdlib, language, benchmark, dependency, or
WASM change. The five bytecode portable-workload rows were the only current
misses without exact current-source CPU and measured-main allocation coverage.
Two fresh main-only CPU processes and two separate exact measured-main
allocation processes for each row close that gap without exposing a new
concrete Able-owned owner in three unlike applications.

Together with the retained group records, all 59 current bytecode target
misses now have source-identity-checked CPU and allocation evidence. The
generated coverage ledger records every selected source SHA-256 and rejects
missing, duplicated, stale, or non-miss rows.

## Source and execution identity

The interpreter test binary was frozen at SHA-256
`d349288a85a68b4282f7936a549631e1910ccbb2760e98297d5aab18d16f4d3d`.
Each process used one pinned logical CPU, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, the catalog-selected serial/goroutine executor, the
canonical external stdlib, skipped repeated setup typechecking, and had a
59-second cap. Workspaces and the Go cache lived under disk-backed
`/var/tmp`.

The host was active during collection. Durations are therefore arithmetic
means used only to contextualize the merged ownership profiles; the current
five-process public-verifier scorecard remains the timing authority.
Allocation counters came from independent unprofiled processes and were
stable to 0-824 bytes and 0-9 allocations within each pair.

| Application | Current source SHA-256 | Catalog arguments | Executor |
| --- | --- | --- | --- |
| Binary Trees | `b2e5cd3b3f439960e39e04b2da675321e9bc45c1b5d811bdb9d61069bb168eeb` | `15` | goroutine |
| N-Body | `7d0cb3f9291be2577e7726e6868deaaf2399223acfe48a666226d7ac697f8e49` | `50000` | serial |
| QuickSort | `13dc68cc43b87d80d21943a45ca24f0722614f3b151f2ed26ef4f1025b103338` | `numbers-bytecode.txt` | serial |
| Sudoku Masks | `222b321f579d7b2a84f4bc0fd379064a7ebe554bd83169782b28d04eaaab90e0` | `1 10` | serial |
| TapeLang Alphabet | `426a40e33840f3a0e9e62d5f9b9519a6840edd2733b8031df6a280e4b782fdb8` | `benchmark-bytecode.tape` | serial |

The first four hashes differ from the calibrated source-copy hashes in the
earlier portable-scale profile manifest. Those older profiles established the
portable workload choice, but they do not substitute for this exact
current-source refresh. TapeLang's source hash was already identical.

## Repeated CPU profiles

The two CPU profiles per application cover only `main()` and were merged
before exact-symbol intersection.

| Application | Main runs | Mean | Current scorecard target excess |
| --- | --- | ---: | ---: |
| Binary Trees | 12.500350 s, 12.200631 s | 12.350490 s | 11.707474 s |
| N-Body | 9.278728 s, 9.097090 s | 9.187909 s | 8.947263 s |
| QuickSort | 13.116637 s, 12.378406 s | 12.747522 s | 11.481684 s |
| Sudoku Masks | 25.455372 s, 25.327144 s | 25.391258 s | 22.503263 s |
| TapeLang Alphabet | 25.750599 s, 22.756962 s | 24.253780 s | 19.909263 s |

These rows account for 74.548947 seconds of the current 221.503684-second
bytecode target excess.

## Exact CPU intersection

Admission required the same exact symbol to have at least 1% flat CPU in at
least three unlike applications. The refresh reproduced only completed or
non-actionable routes.

| Exact symbol | Breadth | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 5 | 4.82%-14.79% | aggregate dispatcher |
| `runtime.tryDeferToSpanScan` | 3 | 2.41%-12.59% | Go GC machinery |
| `sync/atomic.(*Int32).Add` | 3 | 2.00%-6.38% | divergent environment/type-cache and Array lock callers |
| `(*bytecodeVM).appendSlotStackValueChecked` | 3 | 1.48%-3.29% | closed stack-carrier route |
| `bytecodeRawIntegerValueInfo` | 3 | 1.16%-3.44% | closed raw-integer route |
| `(*bytecodeVM).execCallOpcode` | 3 | 1.50%-3.00% | aggregate call dispatcher |
| `(*bytecodeVM).finishInlineReturn` | 3 | 1.25%-2.62% | closed return/frame route |
| `(*bytecodeVM).appendStackValue` | 3 | 1.49%-2.02% | closed stack-carrier route |
| `(*bytecodeVM).execStoreSlot` | 3 | 1.10%-1.55% | aggregate store dispatcher |

The atomic leaf does not identify one removable synchronization boundary.
Binary Trees reaches it primarily through per-call environment runtime data
and type caches. QuickSort reaches it through Array reads/writes and integer
boxing. TapeLang reaches it through Array slot reads/writes. Removing those
locks broadly has already been closed, while special-casing either benchmark
or a named non-primitive container is prohibited.

The dominant concrete CPU owners remain distinct:

- Binary Trees: recursion environments, named-struct construction, and GC;
- N-Body: float materialization, stack snapshots, calls, and member lookup;
- QuickSort: integer boxing, raw integer/Array access, and map/lock work;
- Sudoku Masks: integer bitwise operations, recursive calls, frames, and
  Array reads;
- TapeLang: named-struct member plans, member calls, frames, and Array slots.

## Exact measured-main allocation

| Application | Mean bytes | Mean allocations | Mean frees | Mean GCs | Dominant sampled main owner |
| --- | ---: | ---: | ---: | ---: | --- |
| Binary Trees | 3,830,196,580 | 31,566,686.5 | 31,332,482.5 | 138 | call environments and named structs |
| N-Body | 463,353,152 | 29,301,246.5 | 28,624,309 | 22 | stack snapshots and float materialization |
| QuickSort | 1,136,209,200 | 41,058,116 | 39,811,797.5 | 28 | boxed/cached i32 values |
| Sudoku Masks | 25,048,580 | 196,569 | 181,800 | 1 | named-struct instances |
| TapeLang Alphabet | 163,968 | 851 | 56,552.5 | 0 | below the default allocation sampler's useful resolution |

The main allocation volumes and concrete owners do not repeat across three
unlike applications. TapeLang is a particularly strong discriminator: its
851 measured allocations are immaterial beside the tens of millions in the
numeric and recursive programs. Package-init allocation samples were excluded
from ownership decisions because the exact `main()` counters exclude setup.

## Full 59-row reconciliation

The current bytecode misses partition exactly as follows:

| Evidence group | Rows | CPU | Allocation |
| --- | ---: | --- | --- |
| portable workload admission | 5 | current exact-source refresh | current exact-source refresh |
| float numeric | 4 | retained current exact | retained current exact |
| wide numeric | 3 | retained current exact | retained current exact |
| byte output | 3 | retained current exact | retained current exact |
| text/map | 9 | retained current exact | retained current exact |
| regex | 6 | retained current exact | retained current exact |
| concurrency | 23 | retained current exact | retained current exact |
| iterator/control | 6 | retained current exact | retained current exact |
| **Total** | **59** | **59 current** | **59 current** |

The four bytecode target guards are intentionally excluded. The generated
ledger joins only current target misses, validates their selected source
hashes against the scorecard, and binds each row to the retained frontier
group and CPU/allocation evidence paths.

## Verification and cleanup

The frozen binary completed all 10 CPU processes, all 10 independent exact
allocation processes, and the five sampled allocation-owner processes. The
selected sources and arguments match the current 126-row verifier-backed
scorecard. Focused ledger tests and the existing frontier generator are run
after the manifest update.

Raw binaries, Go caches, pprof data, and temporary reports are disposable and
are removed after the durable ledger and documentation are generated.

## Next recommendation

Build a bytecode semantic-boundary reach map before selecting another
optimization candidate.

Why: CPU/allocation coverage is now complete, and every exact owner currently
repeating across three unlike misses is either aggregate, Go-owned, or already
closed. Another profile-only sweep would repeat evidence rather than identify
where native primitive carriers are being lost.

What it entails: add release-disabled counters at generic value
materialization boundaries, attribute them to primitive kind and consumer
operation without naming applications or containers, run the largest unlike
bytecode misses, and rank only boundaries reached materially in at least
three families. Advance at most one general VM carrier/boundary rule to
repeated verifier-backed A/B measurement.

Why it matters: the bytecode frontier still has 221.503684 seconds of target
excess. A semantic-boundary reach map connects that excess to places where
Able primitives cross into boxed/runtime representations, directly serving
the goal of preserving native carriers while avoiding speculative rewrites.
