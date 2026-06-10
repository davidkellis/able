# Bytecode concurrency fourteen-application refresh

## Outcome

This refresh retains no production runtime, compiler, benchmark, or canonical
stdlib change. Eleven verifier-backed concurrency applications and three
unlike controls completed two independent current main-only CPU profiles and
two independent exact timed-main allocation processes from one frozen
bytecode test artifact.

The only new exact CPU leaf with broad concurrency-specific presence is
`sync/atomic.(*Int32).Add`. Caller traces show that every sampled call is an
implementation detail of `sync.RWMutex`, spread across environment lookup,
revision, runtime-data, parent, struct-definition, alias-cache, and method-
cache stores. It is not a removable executor counter. The allocation profiles
repeat the retained lazy combined `environmentState` family and already-closed
raw-integer, call-argument, stack, and type-match families. No new concrete VM
operation clears the three-unlike-family admission rule.

## Reproducibility contract

- Go: `go1.26.4 linux/amd64`
- repository HEAD: `237406eccdfb025a519d898daedadee1c8d13a7b`
- frozen test binary SHA-256:
  `6758a13355a1adeebe0984098679c3ad344b0f8cf1a8642694e873e3dd12d53e`
- canonical stdlib tree SHA-256:
  `8eebba62436c8dd21bf154405da81f885bb5f8d07032aeffdfefd80903fd030c`
- executor: goroutine for the eleven concurrency applications; normal serial
  execution for Mandelbrot, Word Frequency, and JSON
- CPU: CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, warmed measured
  `main()`, one call per process, and a 59-second process ceiling
- allocation: two independent exact `runtime.MemStats` deltas around the same
  warmed measured `main()` region per application
- ownership supplement: one sampled allocation process for each row except
  Mandelbrot, whose 76-million-allocation exact result made another sampled
  process unnecessary for the bounded decision
- selection rule: admit at most one exact runtime/VM child material in at
  least three unlike concurrency families and absent or semantically distinct
  in unlike controls

The bytecode benchmark loads, typechecks, lowers, and warms the program before
the CPU-measured region. The frozen artifact was reused for every profile and
exact allocation process.

## Current measurements

Arithmetic means retain both independent workstation processes. Bytes,
allocations, frees, and GC counts are exact measured-main deltas rather than
sampled-profile estimates.

| Application | CPU 1 ns | CPU 2 ns | CPU mean ns | Bytes mean | Allocs mean | Frees mean | GC mean |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 46,028,656 | 45,482,531 | 45,755,594 | 10,389,448 | 161,604.5 | 62,430 | 0.5 |
| Channel Rollup | 174,774,726 | 177,734,284 | 176,254,505 | 23,755,656 | 261,135 | 125,508 | 1.0 |
| Concurrent Event Routing | 2,815,973,055 | 3,046,225,425 | 2,931,099,240 | 289,347,356 | 2,820,542 | 2,112,738.5 | 10.0 |
| Concurrent Text Index | 306,570,244 | 361,959,547 | 334,264,896 | 39,621,052 | 497,354 | 242,382 | 2.0 |
| Dependency Wave Validation | 310,738,388 | 293,799,362 | 302,268,875 | 23,921,584 | 383,201 | 197,230 | 1.5 |
| Future Await Race | 11,277,527 | 10,219,103 | 10,748,315 | 1,380,268 | 34,220 | 34,101.5 | 0.5 |
| Future Pipeline | 283,447,657 | 307,835,694 | 295,641,676 | 14,413,020 | 656,074.5 | 335,877.5 | 1.0 |
| Mutex Await Journal | 66,639,881 | 70,610,333 | 68,625,107 | 9,803,548 | 184,239.5 | 46,010 | 0.0 |
| Mutex Ledger | 341,772,329 | 234,155,047 | 287,963,688 | 17,723,792 | 400,705.5 | 198,100 | 1.0 |
| Mutex Work Queue | 290,056,211 | 223,492,606 | 256,774,409 | 26,235,232 | 423,084.5 | 205,808 | 1.5 |
| Validated Job Pipeline | 953,653,275 | 875,874,787 | 914,764,031 | 90,421,096 | 1,388,645 | 893,081.5 | 4.5 |
| Mandelbrot control | 7,360,156,160 | 7,268,053,920 | 7,314,105,040 | 615,170,772 | 76,303,285 | 75,405,891 | 32.5 |
| Word Frequency control | 1,095,327,724 | 1,309,463,694 | 1,202,395,709 | 54,292,720 | 625,866 | 336,035 | 3.0 |
| JSON control | 503,162,695 | 538,322,130 | 520,742,413 | 114,821,872 | 471 | 50,022 | 2.0 |

The wider wall variation in Event Routing, Text Index, Ledger, Work Queue,
and Word Frequency is why both processes are retained and averaged. Exact
allocation volumes stayed sufficiently stable to identify owners without
selecting on one noisy wall sample.

## CPU intersection and reconciliation

At a one-percent flat-share threshold, `sync/atomic.(*Int32).Add` occurs in
nine of the eleven concurrency applications:

| Application | Flat CPU share |
| --- | ---: |
| Channel Rollup | 5.56% |
| Concurrent Event Routing | 8.78% |
| Concurrent Text Index | 12.12% |
| Dependency Wave Validation | 4.92% |
| Future Pipeline | 5.17% |
| Mutex Await Journal | 7.14% |
| Mutex Ledger | 8.77% |
| Mutex Work Queue | 9.80% |
| Validated Job Pipeline | 5.49% |

It is below one percent in all three controls. Exact call traces initially make
this look like the requested common scheduler leaf, but every sample's direct
caller is `sync.RWMutex.RLock`, `RUnlock`, `Lock`, or `Unlock`. The first Able
callers then split:

- Event Routing: lookup/revision hints, runtime data, method and alias caches,
  and parent access;
- Text Index and Dependency Validation: lookup, revision, runtime data, and
  method-cache versioning;
- Future Pipeline: runtime-data read/write and lookup;
- mutex workloads: lookup/revision, alias cache, parent, struct definitions,
  and mutex notification-related environment work;
- Channel Rollup: runtime-data read/write.

These stores are mutable and shared after goroutine execution begins. Replacing
their `RWMutex` with a plain mutex would optimize this one-CPU measurement at
the possible expense of real parallel applications. Removing the lazy mutex
would require either embedding synchronization in every serial environment or
redesigning all mutable environment state. Neither is a bounded consequence of
these profiles, so no lock candidate is admitted.

The other repeated flat CPU symbols are known families: VM dispatch,
`aeshashbody`, raw-integer extraction/materialization, cached member lookup,
inline return, type fitting, and Go GC/map leaves. Their direct consumers split
across the applications or their generic candidates have already completed
broad acceptance/rejection gates. This refresh supplies no invalidating fact
that would justify reopening them.

## Allocation reconciliation

Sampled allocation ownership is used only to label the exact timed-main
counters. Process/bootstrap samples such as `initBytecodeSmallIntBoxCache` are
not treated as measured-main candidates.

`newEnvironmentBase` appears materially across nine concurrency applications;
the lazy `Environment.mutex` state appears materially in Event Routing, Text
Index, Dependency Validation, Future Pipeline, Ledger, Work Queue, and
Validated Job Pipeline. This is the residual object expected after the
2026-07-20 retained change combined the mutex and metadata into one lazy
`environmentState`. The object cannot be removed once the environment is
shared without removing the synchronization it provides.

The remaining repeated allocations are application-dependent combinations of
integer result boxing, argument copying/receiver prepending, stack snapshots,
typed-pattern construction, Array storage, awaitable/waker construction, and
struct instances. The exact children do not establish a new operation in
three unlike concurrency families whose general design is still open.

## Correctness and public-output verification

- `go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s` passed
  in 28.514 seconds.
- One normal CLI bytecode process for each of the fourteen applications
  succeeded and passed its independent public-output verifier.
- JSON completed in 0.91 seconds versus the stored 2.87-second Python
  reference, or 0.32x Python wall time. The concurrency benchmarks do not yet
  have comparable Python rows, so they are semantic/profile coverage rather
  than cross-language target claims.

No production file and no file in the canonical `../able-stdlib` repository
changed in this tranche.

## Next recommendation

Refresh `bytecode-iterator-control` across Array Slice Window, Dependency Plan,
Document Audit, Lexical Rollup, and Option/Result Configuration, with unlike
numeric and text/map controls. It is the only selected bytecode ownership group
whose evidence is still mixed and predates the recent VM/runtime changes; the
other selected bytecode groups now have current exact evidence or a completed
candidate gate.

This entails one frozen warmed-runtime artifact, two CPU and two exact
allocation processes per row, caller reconciliation beneath cached member
lookup, and at most one candidate only if a new concrete VM leaf repeats in
three unlike iterator/control families. This is preferable to reopening the
RWMutex, frame/return, Array growth, nullable, union, or named-container paths
without new evidence, and it advances bytecode performance without starting
WASM work.
