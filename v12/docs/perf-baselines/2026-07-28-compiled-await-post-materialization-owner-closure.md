# Compiled Await post-materialization owner closure

Date: 2026-07-28

## Decision

Retain no production change.

Fresh retained-state CPU and exact allocation profiles across Await Channel
Mux, Future Await Race, Mutex Await Journal, and Mutex Work Queue expose no
open generated compiler/runtime owner that is material in at least three
unlike applications. Every repeated large owner is one of:

- the semantically required distinct native waker;
- an already-rejected typed Awaitable Array/protocol conversion;
- ordinary primitive boxing at a semantic dynamic boundary whose global cache
  already failed a broad application guard;
- the previously closed reverse execution-context/current-goroutine route; or
- mutex/channel work confined to one concurrency family.

The three-unlike-application admission rule therefore stops before a
prototype or A/B timing candidate. This is a no-code closure, not evidence
that the four applications meet the product goal: the latest retained timing
still ranges from 2.57x to 24.30x equivalent Go.

## Measurement

All applications were regenerated from the retained compiler with:

```text
-no-fallbacks -experimental-execution-context -build -main
```

The compiler SHA-256 was
`55c7e9cb6f911406510b67a9f09ccece55cc4cb9a111cf8ceebb234adfc13871`.
Each application ran with `ABLE_EXECUTOR=goroutine`, `GOMAXPROCS=4`, and
`taskset -c 0-3`. Every process output passed its external public verifier,
and every generated dependency graph omitted
`able/interpreter-go/pkg/interpreter`.

Twenty-five measured-main CPU profiles were aggregated per application.
Three exact measured-main allocation snapshots were collected separately per
application:

| Application | CPU duration / samples | Exact main bytes, mean | Exact main objects, mean |
| --- | ---: | ---: | ---: |
| Await Channel Mux | 1.87s / 2.01s | 5,175,597.3 | 96,764.7 |
| Future Await Race | 225.35ms / 320ms | 681,728.0 | 10,979.0 |
| Mutex Await Journal | 73.89ms / 70ms | 965,866.7 | 22,784.3 |
| Mutex Work Queue | 210.02ms / 300ms | 2,117,917.3 | 49,520.3 |

The very short Journal and Queue runs remain CPU-sample sparse even after 25
profiles. Exact allocation ownership is stable across the three independent
snapshots and supplies the stronger selection signal for those applications.

## Owner classification

The representative counts below are exact flat object counts from allocation
snapshot one. A dash means the leaf was absent or below the profile reporting
threshold.

| Owner or conversion family | Channel | Future | Journal | Queue | Disposition |
| --- | ---: | ---: | ---: | ---: | --- |
| distinct native Await waker | 1,024 | 192 | 2,048 | 4,096 | required identity; endpoint reuse is unsafe |
| `bridge.currentGID` | 8,195 | 1,074 | — | — | material in two; broad context routes already closed |
| `bridge.ToInt` | 4,098 | — | 4,101 | 8,197 | ordinary semantic-boundary boxing; global `i64` cache rejected |
| callback-to-runtime wrapper | 4,096 | — | 4,096 | 8,192 | closed Awaitable protocol-arm conversion |
| typed Awaitable Array conversion | 2,048 | 192 Array clones | 4,096 | 8,192 | closed Array round trip/protocol-arm representation |
| runtime Awaitable interface wrapper | 2,560 | — | 2,048 | 4,096 | closed Awaitable protocol-arm conversion |
| mutex waiter notification | — | — | 2,048 | 8,196 | only two mutex applications |
| mutex Await lock | — | — | 2,048 | 4,096 | only two mutex applications |
| channel arm/receive work | material | — | — | — | one application family |

Future Await Race has 287 `bridge.ToDynamicI64` objects, but that helper is
the already-retained cache specifically for proven native-`i64` to dynamic
`runtime.Value` conversions. The remaining three-application `bridge.ToInt`
leaf is ordinary semantic-boundary materialization. Extending the cache to
ordinary `ToInt` previously removed large allocations but slowed the
allocation-light TapeLang guard by 4.17%, so this profile does not reopen that
route.

The CPU profiles agree with the allocation classification:

- Channel Mux spends 1.52s cumulative, 75.62% of samples, in
  `bridge.currentGID`, almost entirely through `runtime.Stack`.
- Future Race spends 230ms cumulative, 71.88%, in the same closed route.
- Journal has only seven scattered 10ms flat samples and no shared dominant
  owner.
- Queue has a 40ms typed Awaitable Array conversion and 90ms cumulative Await
  initialization, but their allocation descendants are the closed protocol
  family. Its `currentGID` sample is only 10ms.

No open owner survives both the breadth and admissibility filters. No
benchmark-specific, mutex-only, named-container, non-primitive nominal,
global boxing-cache, execution-context ABI, interpreter, runtime, stdlib,
application, dependency, bytecode, language, or WASM change was made.

## Preserved evidence and verification

Machine-readable inputs, exact counts, hashes, and the owner dispositions are
in:

- `2026-07-28-compiled-await-post-materialization-owner-closure.json`

Human-readable `pprof` CPU, cumulative CPU, allocation-object, and
allocation-space tops are retained under:

- `2026-07-28-compiled-await-post-materialization-owner-profiles/`

Post-classification verification passed:

- all 112 captured CPU/allocation outputs passed their public verifiers;
- all four generated dependency graphs omit `pkg/interpreter`;
- focused lazy-carrier, user-Awaitable materialization, and stale-waker
  compiler guards pass within the one-minute test limit;
- the performance evidence ledger remains current with the unchanged
  compiler-production hash
  `8d69533a81ce44f58ec8921abbd9867cbeb935aeab3ec9b39312f10aee1f7433`.

## Next

Refresh the broad strict compiled scorecard and owner ranking outside this
now-closed Await sub-frontier.

Why: these four applications remain far below Go, but their remaining shared
costs are either semantically required or already rejected by broader
evidence. Continuing to tunnel into Await would favor one concurrency family
or reopen a failed boundary design.

What it entails: rebuild a feature-covering compiled cohort from the retained
interpreter-free compiler, repeat verifier-backed Able/Go timing, refresh CPU
and allocation profiles for materially changed or poorly explained misses,
and admit only one exact open native-lowering or boundary owner repeated in
at least three unlike programs.

Why it is important: the 95%-of-Go goal requires eliminating general boxing
and semantic adaptation boundaries where native Go carriers can soundly
survive. A broad refresh is the shortest evidence-based route to the next
such boundary after the Await-local owner set has closed.
