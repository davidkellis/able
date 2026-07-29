# Compiled callable-context default activation retained

Date: 2026-07-29

## Decision

Retain automatic activation of the existing allocation-free callable
execution-context ABI when any loaded Able module contains `await`.

This is an activation rule, not a new ABI. Await-free programs retain their
old generated shape, and `--experimental-execution-context` remains a
force-on diagnostic for the broader execution-context machinery. Dynamic and
host-created callables retain their compatibility entries.

The change is general compiler lowering. It does not special-case a benchmark,
container, non-primitive nominal type, or stdlib package, and it does not
change Able language semantics.

## Broad source and semantic gate

Frozen baseline and candidate compilers were built from commit
`ada2a21c751baf51200149e9dac2d175e29aa222` and the candidate worktree,
respectively, with Go 1.26.5. Their SHA-256 identities were:

- baseline:
  `3a60b1daf01a9d260277f437504998871593a6aaa8131a1b9c5d7ae6c9e8578f`;
- candidate:
  `3d0af8e165e573f3525c2667587be6d489a9299dea9a17487775f1698b5e1570`.

The complete 63-application strict catalog was emitted, linked, run, and
checked on both sides:

- 126/126 generated applications linked without `pkg/interpreter`;
- 126/126 processes completed and passed their public verifier;
- stdout identities matched between baseline and candidate for every
  application;
- 59/63 generated `compiled.go` files were byte-identical; and
- exactly Future Await Race, Await Channel Mux, Mutex Await Journal, and Mutex
  Work Queue changed and selected the callable-context ABI.

The current final emitter reproduced all four measured candidate
`compiled.go` files byte for byte after formatting and test edits. Therefore
the performance binaries represent the retained lowering behavior.

## Repeated timing

Seven rotating baseline/candidate/Go cohorts produced seven independent
processes per lane. Every process used the catalog executor and CPU settings
and passed its public verifier.

| Application | Baseline | Candidate | Change | Go | Candidate as Go performance |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 0.036571s | 0.015092s | -58.73% | 0.004160s | 27.57% |
| Await Channel Mux | 0.208787s | 0.113219s | -45.77% | 0.005682s | 5.02% |
| Mutex Await Journal | 0.247917s | 0.010365s | -95.82% | 0.005256s | 50.71% |
| Mutex Work Queue | 0.573129s | 0.016687s | -97.09% | 0.005611s | 33.62% |

The four-application candidate/baseline geometric-mean time ratio is
0.128471, an 87.15% reduction. Because the other 59 generated programs are
byte-identical and all four changed programs improve, no established compiled
guard can regress through this activation.

The raw timing sample SHA-256 is
`0379ce72a6fd7d9720e4f091420a0504b7d9a7847cb450f3054f8f65047ade3e`.

## Allocation and exact boundary evidence

Five independent lightweight `runtime.MemStats` main-phase processes per lane
showed that the speedups did not merely move cost to another allocation owner:

| Application | Allocated bytes | Allocation count |
| --- | ---: | ---: |
| Future Await Race | -21.89% | -10.62% |
| Await Channel Mux | -20.69% | -14.85% |
| Mutex Await Journal | -77.32% | -66.93% |
| Mutex Work Queue | -78.50% | -69.24% |

The allocation sample SHA-256 is
`1bb81f3ae3c91233821c809245b1bb30daa5279a48ea1730ea25c826d35bb4f3`.

A diagnostic-only bridge overlay then counted `currentGID` calls in five
verified processes per lane:

| Application | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Future Await Race | 2,476.8 | 1,074.8 | -56.61% |
| Await Channel Mux | 15,875.0 | 8,195.0 | -48.38% |
| Mutex Await Journal | 16,174.2 | 14.0 | -99.91% |
| Mutex Work Queue | 38,243.2 | 14.0 | -99.96% |

The residual calls belong to retained compatibility or scheduler services;
the hot generated callable path no longer recovers its context through
goroutine identity. The boundary sample SHA-256 is
`6995a821ed9443cfb271a5e132d712c8d3aa652a41fbf8e7a21dbba99fe5b5ad`.

The reach overlay also proved actual default execution of 922, 512, 2,048,
and 4,096 context-aware callable calls and 192, 1,024, 2,048, and 4,096
context-aware await calls in the four applications. Its TSV SHA-256 is
`a4b0aba7b206dd19eeac273d195526a19678e586e3b01289641f64c8bcc5c0d2`.

## Implementation

- Scheduler-context detection is option-independent and scans all loaded
  modules, so imported await bodies activate the ABI.
- Context machinery activates by default only for scheduler-requiring program
  graphs.
- The existing option force-enables execution context for diagnostic
  comparison but does not make await-free programs emit the scheduler-only
  callable surface.
- The callable-context census instruments every emitting mode now that
  default mode can legitimately select the ABI.
- Focused source guards cover default, imported, forced, and await-free
  selection. Existing closure-owned method and native Awaitable expectations
  now assert the scheduler-selected default.

No runtime, interpreter, bytecode VM, stdlib, language, dependency,
application, fixture, non-primitive nominal, or WASM behavior changed.

## Verification

The following bounded gates pass:

- default/imported/forced/await-free callable-context source guards;
- nested spawn, captured interface callback, closure capture, nested await,
  native Awaitable/waker, and stale-waker execution;
- the full experimental execution-context fixture parity test;
- dynamic named/value/bound-method and host-context compatibility tests;
- `go test ./cmd/ablec ./pkg/compiler/bridge`;
- the closed-benchmark static-kernel guard together with the new activation
  guards; and
- shell syntax, formatting, source-size, and whitespace checks.

A monolithic `go test -short ./pkg/compiler` invocation was deliberately
capped at one aggregate minute and exhausted that package-level cap while
running many tests; it reported no semantic assertion failure. Its active
closed-benchmark test and the changed focused tests passed together in
14.664 seconds. The relevant broader semantic groups had already passed in
bounded commands, each under one minute.

## Next

Refresh and promote the authoritative 63-application compiled scorecard with
the retained default.

Why: this tranche proves the four affected applications improve and the other
59 are structurally unchanged, but the checked scorecard still reports the
pre-activation frontier. The next optimization must be selected from current,
not stale, end-to-end residuals.

What it entails: rebuild all 63 strict applications with the retained
compiler, collect five matched verifier-backed Able/Go processes using the
locked Go 1.26.5 contract, regenerate the compiled frontier and closure
ledger, and profile the largest post-activation owner that repeats in at least
three unlike misses.

Why it is important: the four applications still deliver only 5.02%-50.71% of
Go performance. Updating the governing evidence will show whether remaining
cost is shared scheduler machinery, a different general lowering boundary, or
application-specific work, and prevents spending the next tranche on an owner
this change already removed.
