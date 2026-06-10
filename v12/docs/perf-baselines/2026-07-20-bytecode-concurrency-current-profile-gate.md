# Bytecode concurrency current-profile gate

## Outcome

This gate retains one generic runtime allocation improvement shared by all six
verifier-backed concurrency applications. Each lexical `Environment` now
publishes one lazy `environmentState` containing its `sync.RWMutex` and
metadata instead of publishing a mutex and metadata as two independent heap
objects. Atomic publication, lock scope, shared thread-mode behavior, and
metadata semantics are unchanged.

No benchmark, verifier, reference implementation, language syntax, compiler
lowering, named nominal path, or canonical stdlib source changed.

## Reproducibility contract

- Go: `go1.26.4 linux/amd64`
- repository HEAD: `237406eccdfb025a519d898daedadee1c8d13a7b`
- frozen baseline test binary SHA-256:
  `e5bc433baa1438b48a4cf83d28c8fdbacbfeb2a17d72df0aca9e9933cc0b0bed`
- measured retained-candidate test binary SHA-256:
  `b35aa510834b44ac4bfe098be9bf8a635865d4e3a2a3e431478b6450c0802125`
- final validation binary after the source-only module split SHA-256:
  `c7ed95ee343def90f2baf6239f18aa881c41faa8f479600fdeef26fa4e19e3a1`
- canonical stdlib tree SHA-256:
  `3df668f5f6287679e3161ddebbfdee9aed76f8bb390d47781931e236c3d0dd8e`
- executor: goroutine for the application cohort; the two fixture guards use
  their normal serial executor
- measurement controls: CPU 0, `GOMAXPROCS=1`, `GOGC=50`,
  `GOMEMLIMIT=1GiB`, a warmed measured region, and a 55-second per-process
  ceiling
- selection protocol: ten order-balanced preserved-binary A/B processes per
  concurrency application; five per unlike guard
- CPU and allocation profiles were collected in separate processes.

Application source SHA-256 values were:

- Channel Rollup: `cc7ddd6dca16348087e17b89f07a30afa536750d088606744f9c22ce0704808a`
- Future Pipeline: `823bcdb46878c9d57ab6438bed3767861974227b96318572dc823f3da69067b7`
- Future Await Race: `7950a425ac6225a577d7c234395ee4293d5c89c2ddfcb19fd87732e8d4d0335b`
- Await Channel Mux: `31bd5d760fff06cdd813c5520f961fbaadb5222908608f2630c183c8598f5b0c`
- Mutex Ledger: `2975ac5ae318fd528fd353ee8846f4cb86104e8525a43c2aa1c31e05dfdb1a61`
- Mutex Await Journal: `3d17db99f2a06f2d87868dcb00e129d6f0cb18d4209b329e09bac281adaf8696`

## Current attribution and candidate

Fresh baseline CPU profiles repeated VM execution/call parents, GC work, and
raw integer handling, but their exact children split by application. Separate
allocation profiles exposed one concrete cross-program boundary:

- `newEnvironmentBase`, `Environment.mutex`, and `ensureMetaNoLock` occurred
  in every application;
- the separate mutex and metadata allocations were material in five unlike
  applications and still present in the sixth;
- depending on the application, the two allocations represented roughly 11%
  to 63% of non-process-initialization allocation objects.

The retained state is a runtime-wide environment representation rule, not a
concurrency-type or workload special case. The state remains lazy and uses the
same atomic compare-and-swap publication. In single-thread mode callers remain
lock-free; in multi-thread mode the same embedded `RWMutex` protects the same
fields.

## Repeated A/B results

Negative wall-time percentages are improvements. Means retain all valid
workstation samples.

| Application | Runs | Baseline ns/op | Retained ns/op | Change | Baseline allocs/op | Retained allocs/op | Objects removed |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Channel Rollup | 10 + 10 | 202,176,457 | 195,501,478 | -3.30% | 283,708 | 224,894 | 58,814 |
| Future Pipeline | 10 + 10 | 288,152,062 | 290,966,747 | +0.98% | 688,293 | 655,503 | 32,790 |
| Future Await Race | 10 + 10 | 10,596,838 | 10,698,682 | +0.96% | 33,510 | 33,026 | 484 |
| Await Channel Mux | 10 + 10 | 59,100,040 | 58,151,236 | -1.61% | 153,759 | 148,640 | 5,119 |
| Mutex Ledger | 10 + 10 | 194,083,852 | 196,479,713 | +1.23% | 433,107 | 400,330 | 32,777 |
| Mutex Await Journal | 10 + 10 | 71,974,306 | 72,245,970 | +0.38% | 189,651 | 183,499 | 6,152 |

The combined object can use more bytes where an environment needs a mutex but
never metadata: the largest observed steady increase was 2.30% in Mutex
Ledger. This trade remains inside the broad memory and wall-time guards while
removing tens of thousands of GC-visible objects in the material programs.

Unlike controls also pass:

| Guard | Runs | Baseline ns/op | Retained ns/op | Change |
| --- | ---: | ---: | ---: | ---: |
| JSON target meet | 5 + 5 | 523,001,587 | 527,736,404 | +0.91% |
| PiDigits target meet | 5 + 5 | 2,159,306,145 | 2,210,090,003 | +2.35% |
| numeric Array map | 5 + 5 | 88,097,184 | 83,338,682 | -5.40% |
| linked-list iterator collect | 5 + 5 | 432,324,929 | 438,649,731 | +1.46% |

## Post-change attribution

The separate `ensureMetaNoLock` allocation leaf disappears. One lazy state
allocation remains attributed to `Environment.mutex`, as expected. Residual
CPU and allocation ownership no longer establishes another shared candidate:

- Channel Rollup is member lookup/call, GC, and frame work.
- Future Pipeline and Future Await Race are raw-integer arithmetic plus
  future scheduling.
- Await Channel Mux is await-arm, waker, receiver-argument, and channel-value
  construction.
- Mutex Ledger is raw integers, mutex notification, boxing, and call traffic.
- Mutex Await Journal is mutex-awaitable/waker construction, raw integers,
  maps, and GC.

Already-closed raw integer, call/frame, return, and execution-context designs
were not reopened because the refreshed profiles supplied no new exact child.
The bytecode concurrency group is therefore current and closed after this
retained allocation rule.

## Correctness and public-output verification

The normal external harness ran five fresh bytecode processes per application
and guard. Every successful output passed its public verifier:

| Application | Result | Mean real time |
| --- | --- | ---: |
| Channel Rollup | 5/5, verified 5/5 | 0.582 s |
| Future Pipeline | 5/5, verified 5/5 | 0.394 s |
| Future Await Race | 5/5, verified 5/5 | 0.138 s |
| Await Channel Mux | 5/5, verified 5/5 | 0.214 s |
| Mutex Ledger | 5/5, verified 5/5 | 0.372 s |
| Mutex Await Journal | 5/5, verified 5/5 | 0.204 s |
| JSON | 5/5, verified 5/5 | 0.858 s |
| PiDigits | 5/5, verified 5/5 | 2.442 s |

JSON remains faster than both stored Python and Ruby references; PiDigits
remains faster than its stored Ruby reference. The focused environment suite,
environment race suite, and bytecode Future/channel/mutex/await slices pass.
After the source-only module split, the final binary reproduced Channel
Rollup's retained allocation fingerprint at 224,893 allocs/op over five calls.
The whole interpreter-package command did not return a usable completion on
this workstation, so it was not treated as evidence and was not extended past
the repository's sub-minute testing rule.

The bounded artifacts remain under
`/tmp/able-bytecode-concurrency-20260720-a` for this session.

## Next recommendation

Refresh compiled byte/output ownership for Base64, FASTA Generation,
PiDigits, and Reverse Complement. It is now the only stale actionable group in
the generated frontier. Preserve current compiled binaries, collect separate
bounded main-phase CPU/allocation profiles, and advance only an exact generated
runtime/compiler descendant shared by at least three unlike applications.
Base64 and PiDigits are near the target, so retained changes must also preserve
the compiled target-meeting controls. This is the right next gate because all
selected bytecode groups now have current or closed evidence; no WASM work is
needed or authorized.
