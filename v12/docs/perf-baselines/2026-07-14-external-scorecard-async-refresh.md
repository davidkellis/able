# External scorecard and async coverage refresh — 2026-07-14

## Decision

Refresh the verifier-backed, CPU-pinned application scorecards and keep no
compiler, bytecode VM, runtime, or canonical-stdlib performance change. The
new measurements make the product gaps clearer, especially in compiled
concurrency applications, but a ratio is selection evidence rather than a
license for a channel-, Future-, Mutex-, or benchmark-specific fast path.

The report-only scoreboard's default inputs now point at these fresh source
scorecards, so `just bench-scoreboard-check` validates the published current
report rather than the superseded July 13 set.

## Method

Every process was pinned to CPU 15 with the canonical external stdlib and:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1
```

The stable 16-application `generality` suite used one verifier-backed run per
reference and Able mode, matching the prior scorecard's status-screen method
and 45-second cap. The non-overlapping async coverage rows used three runs:
Channel Rollup, Future Pipeline, Future Await Race, Await Channel Mux, Mutex
Ledger, and Mutex Await Journal. Every completed process passed its suite's
exact Ruby verifier; timeout rows remain unranked.

## Stable target status

The refreshed scoreboard has 22 rows per target mode: 16 stable generality
applications plus six async coverage applications.

| Target | Rankable rows meeting target | Material status |
| --- | ---: | --- |
| Compiled / Go | 5 / 21 | Fib, BinaryTrees, QuickSort, Base64, and JSON meet; Sudoku remains cap-bound |
| Bytecode / Python and Ruby | 2 / 14 | JSON and PiDigits meet; six generality rows are cap-bound or lack a fair completed pair |

The largest completed stable misses remain heterogeneous: compiled
K-Nucleotide is 80.32x Go, Sudoku Masks 15.47x, N-body 14.48x, and Reverse
Complement 8.75x. Bytecode Reverse Complement is 242.01x Python / 91.71x
Ruby, while K-Nucleotide is 30.97x / 32.55x. They do not by themselves identify
a common concrete implementation leaf.

## Async coverage status

All six fresh async rows completed and verified. They establish a broad,
portable performance deficit across spawn, channel, future, await, Mutex, and
callback/ensure behavior.

| Application | Compiled / Go | Bytecode / Python | Bytecode / Ruby |
| --- | ---: | ---: | ---: |
| Channel Rollup | 598.21x | 17.60x | 13.53x |
| Future Pipeline | 167.30x | 9.53x | 7.98x |
| Future Await Race | 32.43x | 5.64x | 3.19x |
| Await Channel Mux | 88.41x | 2.42x | 2.83x |
| Mutex Ledger | 136.50x | 23.15x | 13.81x |
| Mutex Await Journal | 117.50x | 13.37x | 6.07x |

The earlier Awaitable interpreter profile already showed that its dispatcher
does not share a concrete hot descendant across mutex, channel, and future
applications. These scorecard rows therefore do not reopen an Awaitable fast
path. They do make generated compiled concurrency runtime the clearest next
profile family: it misses every async application by at least 32x Go.

## Artifacts and verification

- `2026-07-14-{compiled,bytecode}-generality-scorecard.{json,md}` and their
  Go/Python/Ruby reference refreshes record the stable suite.
- `2026-07-14-{compiled,bytecode}-async-coverage-scorecard.{json,md}` and
  matching reference refreshes record the six three-run async applications.
- [The current external scoreboard](external-scoreboard-current.md) is rebuilt
  from exactly those four source scorecards.
- `just bench-scoreboard-check` passes, as does `git diff --check`.

## Compiled async main-phase follow-up

The planned generated-main profile gate is complete. Channel Rollup, Future
Pipeline, and Future Await Race all repeat the same `bridge.currentGID` /
`runtime.Stack` identity leaf (93.74%, 94.04%, and 64.29% cumulative in the
merged captures). It is already addressed only by the opt-in fixed-context
ABI; its default and package-linkage forms were separately rejected by
verifier-backed N-body and K-Nucleotide generality regressions. Keep no source
change and do not retry that unchanged candidate. The method and evidence are
in `2026-07-14-compiled-async-main-phase-refresh.md`.

## Next recommendation

The roadmap reconciliation found no missing v12 semantic boundary that can
honestly become a new portable timing row. Do not create a synthetic workload
for dynamic modules, externs, or host callbacks. The next active product slice
is instead WASM `able_host` stdout/stderr forwarding with Node smoke coverage;
it is a portability feature, not performance evidence. Continue to use this
verifier-backed scorecard and the existing broad corpus to select future
compiler/VM candidates.
