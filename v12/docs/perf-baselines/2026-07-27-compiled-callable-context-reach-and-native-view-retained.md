# Compiled callable-context reach and native-view tranche retained

## Decision

Retain two general additions:

1. a reusable default/experimental callable-context reach census with
   diagnostic-only execution overlays; and
2. an embedded `runtime.NativeCallContext` view in await-reached experimental
   execution contexts, replacing one heap allocation at every context-aware
   native call.

The native view is initialized when an execution context is created or
localized and is thereafter read-only. It is emitted only when
`ExperimentalExecutionContext` and the entry-package `await` gate are both
active. Default builds and await-free experimental builds retain the prior
renderer path.

No Channel, Mutex, Future, Awaitable, benchmark, named-container, or
non-primitive nominal rule was added. No canonical stdlib, interpreter,
bytecode VM, language, dependency, or WASM source changed.

## Full-catalog reach census

The first census emitted, linked, ran, and publicly verified default and
experimental forms of all 63 portable applications:

- 126/126 final dependency graphs omit `pkg/interpreter`;
- 126/126 runs and public verifiers pass;
- all 63 default/experimental stdout SHA-256 pairs match;
- default mode emits the new callable-context helper in 0/63 applications;
- experimental mode emits and executes it in exactly four applications.

| Experimental application | Context-aware callable calls | Context-aware awaits |
| --- | ---: | ---: |
| Future Await Race | 909 | 192 |
| Await Channel Mux | 9,216 | 1,024 |
| Mutex Await Journal | 8,531 | 2,048 |
| Mutex Work Queue | 17,320 | 4,096 |

After the embedded native view was added, all 63 default applications were
emitted, linked, and verified again. Compared with the pre-change census,
every generated-source hash, final-binary hash, stdout hash, dependency
decision, and verifier result is identical. The four final experimental
positive controls also remain interpreter-free and verifier-correct, with
954/192, 9,216/1,024, 8,549/2,048, and 17,200/4,096 diagnostic calls/awaits.

## Fourth reached application

The prior callable-context tranche had repeated A/B evidence for three of the
four reached applications. Five fresh verifier-backed runs completed the
fourth:

| Future Await Race | Able real | Go real | Able / Go |
| --- | ---: | ---: | ---: |
| Default | 0.0800 s | 0.0046 s | 17.39x |
| Retained callable context | 0.0360 s | 0.0046 s | 7.83x |

The callable context improves Future Await Race by 55.00%, confirming a fourth
unlike reached application. Its main allocation result is a tradeoff:
allocation objects increased in the initial sample even though wall time fell.

## Embedded native-call view A/B

Allocation profiles identified `__able_native_call_context` as a general
owner repeated across all four applications. It allocated a fresh pointer for
each context-aware native call. The retained rule embeds that value in the
already-required execution context.

Five exact main-phase allocation runs per side measured:

| Application | Retained bytes | Embedded bytes | Change | Retained objects | Embedded objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 911,644.8 | 894,347.2 | -1.90% | 15,541.0 | 14,536.0 | -6.47% |
| Await Channel Mux | 10,535,940.8 | 10,402,124.8 | -1.27% | 189,440.4 | 179,210.4 | -5.40% |
| Mutex Await Journal | 5,432,027.2 | 5,420,182.4 | -0.22% | 99,819.8 | 92,240.2 | -7.59% |
| Mutex Work Queue | 11,115,499.2 | 11,015,352.0 | -0.90% | 205,100.0 | 188,194.8 | -8.24% |

Fifteen rotating retained/embedded pairs per application found wall changes of
-0.45%, -5.57%, -8.33%, and +4.45%. Every paired 95% interval crossed zero:
[-24.18%, 23.28%], [-23.19%, 12.05%], [-23.19%, 6.52%], and
[-11.11%, 20.01%]. The rule is retained for its broad allocation improvement;
no wall-time improvement or regression is claimed.

A separate five-run final snapshot measured the embedded candidate at 8.80x,
21.56x, 9.23x, and 11.74x equivalent Go. These naturally noisy absolute
values are not used as the embedded-view A/B decision.

## Residual owner result

Ten verified main-only CPU profiles per application and three exact
`currentGID` overlays found:

| Application | Merged CPU samples | `currentGID` cumulative | Exact lookup mean |
| --- | ---: | ---: | ---: |
| Future Await Race | 0.20 s | 85.00% | 1,277.3 |
| Await Channel Mux | 0.88 s | 64.77% | 8,707.0 |
| Mutex Await Journal | 0.13 s | 38.46% | 157.7 |
| Mutex Work Queue | 0.34 s | 14.71% | 185.7 |

Environment recovery remains important in Future Await Race and Await Channel
Mux, but is no longer a material shared owner across all four. Context-aware
method/callable dispatch repeats in every profile, while its cost is split
among environment recovery, method lookup, native-bound-method argument
injection, await-state construction, and allocation. No second production
candidate was advanced without exact branch-level attribution.

## Evidence ledger and correctness

The ledger trial differed only in compiler-production scope identity:
`file_count` 281 -> 284 and tree SHA-256
`68d53484...8f54` -> `4487fff4...1ddd`. The exact 63-application default
identity proof justified that scope-only bootstrap. The checked ledger and all
seven ledger tests pass with 21 current closures and zero invalidations.

Correctness gates passed:

- focused callable ABI, await-free gate, nested/captured/cross-package,
  native-interface, bound-method, static spawn/kernel, fixed-helper, and public
  Mutex tests;
- full experimental fixture parity;
- dynamic boundary and dynamic named/value compatibility;
- `go test ./pkg/compiler/bridge`;
- `go test ./cmd/ablec`.

## Retained evidence

- `2026-07-27-compiled-callable-context-reach-census.tsv`
- `2026-07-27-compiled-callable-context-final-{default,experimental}-census.tsv`
- `2026-07-27-compiled-callable-context-future-{baseline,candidate}.json`
- `2026-07-27-compiled-callable-context-{future,reached}-go-reference.json`
- `2026-07-27-compiled-callable-context-native-view-candidate.json`
- `2026-07-27-compiled-callable-context-native-view-{balanced,allocation}.tsv`
- `2026-07-27-compiled-callable-context-residual-current-gid.tsv`
- `2026-07-27-compiled-callable-context-residual-profiles/`
- `bench_compiler_callable_context_reach_census`
- `bench_compiler_callable_context_reach_instrument.py`

## Next

Add diagnostic-only branch counters to the remaining context-aware
method/callable dispatch path and attribute each residual `currentGID` lookup
to its exact caller across the four reached applications.

Why: `__able_method_call_ctx` / `__able_call_value_fast_ctx` repeats in all
four profiles, but the sampled cost is currently an inseparable mixture and
environment recovery alone is not broad enough for another change. What it
entails: count native function, native bound method, compatibility/default,
method lookup, argument injection, environment swap, and error-control
branches; repeat verified runs; then profile only the largest branch shared by
at least three applications. Why it is important: the applications remain
8x-22x slower than Go, and exact attribution is required to continue lowering
Able callables toward native Go without another one-family boundary rule.
