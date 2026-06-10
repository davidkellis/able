# Compiled Dynamic `i64` Boundary Cache

Date: 2026-07-21

## Decision

Retain a primitive-specific cache only at the generated native-`i64` to
dynamic-`runtime.Value` boundary. The candidate improves three unlike primary
applications, clears the extended TapeLang, Binary Trees, short-program, and
startup guards, and preserves all tested semantics.

This is not a HashMap, Result, application, or other non-primitive nominal
lowering rule. The compiler already routes every statically native value that
must enter a dynamic semantic operation through `runtimeValueExpr`; the change
only selects a suffix-specific bridge helper for that existing `int64`
language boundary. Ordinary `bridge.ToInt`, other integer kinds, dynamic
fallback behavior, and non-primitive translation are unchanged.

## Site and escape classification

Fresh generated programs covered K-Nucleotide, Inventory Reconciliation,
Validated Job Pipeline, TapeLang Alphabet, Option/Result Config, and Unicode
Scalar Pipeline. Generated call-site classification and Go `-m=2` output found
the same concrete mechanism in the three primaries:

- K-Nucleotide's specialized map calls box native `i64` handles while also
  boxing `u64` keys;
- Inventory's specialized map calls box native `i64` handles, keys, and
  values; and
- Validated Job's generic Result/operator path boxes native `i64` constants
  and loop values before dynamic binary operations.

The generated `[]runtime.Value` slices at the map sites do not escape, but the
interface payloads still escape conservatively through their generic
consumers. Validated Job's inlined `runtime.NewSmallInt` results likewise
escape at every hot dynamic binary-operation site. The common shape is the
compiler's native-primitive to dynamic-value conversion, not either consuming
nominal or helper name.

Escape analysis also corrected the previous TapeLang interpretation. A clean
3.52-second CPU profile contains only native Tape execution (`execute`, `inc`,
`get`, and `move`) and has no `bridge.ToInt` match. Its measured main allocation
is 282,552 bytes/4,274 objects with either prior cache binary. The earlier
global-cache rejection remains correct under its broad timing gate, but its
TapeLang movement cannot be attributed to an executed range/cache branch.

## Rejected inlining detour

The first prototype moved the retained `i32` cache logic behind a helper. This
made `bridge.ToInt` inlineable at Go cost 79 instead of non-inlineable cost 89.
It did not change allocation, because every profiled hot interface payload
still escaped through the generic consumer. Inlining also duplicated code:
K-Nucleotide grew 26,216 bytes and Inventory grew 19,160 bytes.

Repeated verified means rejected that formulation:

| Application | Pairs | Baseline | Inline candidate | Change |
| --- | ---: | ---: | ---: | ---: |
| K-Nucleotide | 10 | 3.505 s | 3.531 s | +0.7% |
| Inventory Reconciliation | 10 | 0.236 s | 0.241 s | +2.1% |

The inlining changes and their test were fully removed before the retained
candidate was built.

## Retained implementation

`bridge.ToDynamicI64` owns a separate lazy, `sync.Once`-initialized set of
immutable `i64` runtime values for `-128..4095`. Values outside the range use
the existing `runtime.NewSmallInt` representation. The suffix remains
observable as `i64`; the retained `i32` cache remains distinct and unchanged.

Only `generator.runtimeValueExpr(..., "int64")` emits the new helper. Other
generated `bridge.ToInt` calls, including runtime returns, struct
materialization, nullable conversion, and non-`i64` primitives, retain their
old path. Tests cover range endpoints, suffix/value preservation, fallback,
steady-state zero allocation, and generator routing while proving `int32`
routing is unchanged.

## Allocation gate

Every profiled output passed its external verifier. Inventory and Validated
Job use exact measured-main snapshots. K-Nucleotide uses ordinary cumulative
allocation profiles because exact profiling exceeds the one-minute bound.

| Application | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: |
| K-Nucleotide sampled objects | 29,725,713 | 22,547,800 | -24.1% |
| K-Nucleotide sampled `bridge.ToInt` objects | 11,491,168 | 3,692,029 | -67.9% |
| Inventory main bytes / objects | 68,744,928 / 1,630,307 | 17,116,640 / 553,186 | -75.1% / -66.1% |
| Validated Job main bytes / objects | 75,558,184 / 1,983,291 | 60,564,808 / 1,669,518 | -19.8% / -15.8% |

After the change, `bridge.ToUint` is K-Nucleotide's largest sampled allocation
leaf at 7,799,139 objects. That is selection evidence for a future breadth
census, not authorization to add an unsigned cache from one application.

## Repeated wall-time gate

All application processes ran their external Ruby verifier. Pair order
alternated. Volatile or short rows received ten or twenty pairs; Noop used 60
order-balanced launches with nanosecond timing. Means are arithmetic means.

| Application | Pairs | Baseline | Candidate | Change | Mean GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| K-Nucleotide | 5 | 3.394 s | 2.866 s | -15.6% | 61.2 -> 40.8 |
| Inventory Reconciliation | 10 | 0.265 s | 0.177 s | -33.2% | 8.0 -> 5.0 |
| Validated Job Pipeline | 10 | 3.060 s | 3.001 s | -1.9% | 11.0 -> 10.0 |
| TapeLang Alphabet | 10 | 4.160 s | 3.721 s | -10.6% guard-only | 4.2 -> 4.0 |
| Option/Result Config | 20 | 0.2115 s | 0.2050 s | -3.1% | 7.25 -> 7.30 |
| Unicode Scalar Pipeline | 10 | 0.236 s | 0.235 s | -0.4% | 5.9 -> 5.8 |
| Sudoku Masks | 5 | 1.824 s | 1.780 s | -2.4% | 13.8 -> 14.6 |
| Binary Trees, 4 CPU/goroutine | 10 | 7.319 s | 7.243 s | -1.0% | 64.1 -> 63.9 |
| Noop startup | 60 | 65.262 ms | 65.210 ms | -0.1% | 3.82 -> 3.80 |

TapeLang's host spread was unusually wide (3.40-6.10 seconds baseline), and
its hot loop does not execute the changed boundary, so its favorable mean is
reported only as a no-regression guard. Binary Trees initially appeared 2.5%
slower after five pairs; the required extension reversed that movement and
finished 1.0% faster over ten, with six candidate pair wins. Option/Result's
first ten-pair mean was distorted by two adjacent candidate outliers; the
twenty-pair cohort resolves to neutral/favorable. Exact Sudoku allocation is
identical at 156,370,688 bytes/7,802,594 objects and 11 measured-main GCs, so
its process-level GC variation is not structural.

## Verification

- `go test -race ./pkg/compiler/bridge -count=1 -timeout 60s` passes.
- Focused generator, integer, HashMap, union, and dynamic-boundary compiler
  controls pass in 19.482 seconds.
- `go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s` passes in
  24.338 seconds.
- Every measured benchmark process and allocation-profile output verified.
- No canonical stdlib, Able benchmark, verifier, reference implementation,
  bytecode VM, language, spec, or WASM source changed.

## Next direction

Refresh a bounded unsigned dynamic-boundary census across at least six unlike
compiled misses. K-Nucleotide now exposes `bridge.ToUint` as its largest
allocation leaf, but one program cannot justify a cache. Reconcile suffix,
value-range reuse, generated boundary site, and actual escape/materiality in
unsigned-heavy text, numeric, and wide-integer applications. Advance only a
primitive boundary rule repeated materially in at least three unlike programs,
then repeat startup, TapeLang, Binary Trees, dynamic-fallback, and bytecode
guards. Do not add named-container rules or resume WASM work.
