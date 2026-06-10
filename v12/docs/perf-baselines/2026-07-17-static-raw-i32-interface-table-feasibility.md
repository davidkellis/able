# Static raw-i32 interface-table feasibility

Date: 2026-07-17

## Decision

Reject and fully remove the generated static interface-table candidate. Static
interface literals preserve the existing `bytecodeRawI32SlotValue` dynamic type
and do eliminate the corresponding runtime boxes, but Go emits one static
temporary symbol per entry. Two bounded sizes show linear executable growth
that would add about 44.7 MB to the ordinary test binary or 26.1 MB after
stripping if extended to the current 263,168-entry cache.

That is disproportionate to avoiding about 1.05 MB and 262,910 startup
allocations. No production, generated source, stdlib, fixture, benchmark, or
language change is retained.

## Trial design

A generated `[...]runtime.Value` literal covered the first 4,096 and then
8,192 entries of the eager raw-`i32` cache. Cache construction copied those
already-boxed interfaces into the existing cache array and constructed all
remaining entries normally. Consequently:

- `bytecodeRawI32SlotCachedValue(...)` and every runtime lookup path were
  unchanged;
- the cached interface dynamic type remained `bytecodeRawI32SlotValue`;
- the existing identity/type sentinel test exercised a value inside the static
  segment; and
- each generated file remained 520 lines, below the project limit.

The baseline binary was built before either candidate. Each state reused its
binary for five initialization traces. The 4,096-entry baseline/candidate
micro gate used two opposite-order cohorts of seven fixed one-million-iteration
samples per sub-benchmark.

## Initialization result

| Static entries | Mean init clock | Mean init bytes | Mean init allocations | Allocations removed |
| ---: | ---: | ---: | ---: | ---: |
| 0 | 32.8 ms | 38,003,448.0 | 707,336.4 | - |
| 4,096 | 30.0 ms | 37,988,132.8 | 703,498.4 | 3,838.0 |
| 8,192 | 28.6 ms | 37,971,774.4 | 699,402.6 | 7,933.8 |

The allocation result is exact apart from unrelated one-allocation trace
variation. Go already supplies static storage for 258 values in this range, so
the 4,096-entry literal removes 4,096 - 258 = 3,838 boxes and doubling the
literal removes another 4,096. Avoided allocation bytes scale at four bytes per
remaining raw value.

The five-trace clocks move in the expected direction, but initialization clock
is workstation-noisy and is not sufficient to offset the executable-size
result.

## Executable and compiler cost

| Static entries | Unstripped bytes | Delta | Stripped bytes | Delta | Extra `.stmp` symbols |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 37,495,960 | - | 28,352,680 | - | - |
| 4,096 | 38,165,904 | +669,944 | 28,733,640 | +380,960 | 4,096 |
| 8,192 | 38,887,480 | +1,391,520 | 29,163,720 | +811,040 | 8,192 |

The ELF data section grows by about twenty bytes per literal and the text plus
symbol/debug representation adds substantially more. From the larger bounded
point, linear extension to all 263,168 entries projects +44,702,580 unstripped
bytes and +26,054,660 stripped bytes. It would also add about 8.1 MB of
generated Go source and 263,168 compiler temporaries.

Three forced baseline builds averaged 23.01 seconds; three later 4,096-entry
builds averaged 21.28 seconds. Because these were ordered rather than
interleaved, they establish no measured build regression but are not evidence
of an improvement. Executable growth independently fails the admission gate.

## Lookup controls

The source lookup path and concrete dynamic type were identical. Fourteen
samples per case were pooled because the workstation results were volatile:

| Path | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| raw integer info, cached raw `i32` | 2.3541 ns | 2.4429 ns | +3.77% |
| direct small `i32`, cached raw `i32` | 0.9984 ns | 0.8840 ns | -11.46% |
| validated slot, cached raw `i32` | 2.9394 ns | 3.2150 ns | +9.38% |
| direct small `i32`, boxed `i32` | 6.1366 ns | 6.3702 ns | +3.81% |

Unchanged unrelated controls ranged from -16% to +17% across the same pooled
cohorts. The mixed direction is binary-layout/workstation noise rather than a
dispatch effect, unlike the rejected pointer-carrier design. No application
A/B was warranted after the static-size gate failed.

## Verification and cleanup

- The cached identity/concrete-type sentinel test passed at both segment sizes.
- Focused raw integer, slot, frame, array-index, and comparison guards pass
  after restoration.
- The restored test binary is exactly 37,495,960 bytes, matching the preserved
  baseline size.
- The restored interpreter init trace returns to about 38.00 MB and 707,336
  allocations.
- The generated literal, generator, cache-construction branch, test binaries,
  traces, and raw benchmark outputs were removed after this aggregate record
  was written.

## Next recommendation

Measure actual raw-`i32` cache value distribution across the selected bytecode
suite, then test a smaller eager upper bound only if the distribution supports
one common threshold.

Why: both representation-preserving static literals and pointer packing are
now ruled out. The remaining generic question is whether the current upper
bound of 262,143 is materially larger than real program demand. Right-sizing
would keep the concrete type and hot lookup unchanged while eliminating unused
startup boxes without increasing binaries.

What it entails: add temporary threshold counters—not a retained workload
heuristic—for cache hits and uncached raw results across unlike string/map,
iterator, numeric Array, recursion, allocation-heavy, and allocation-light
bytecode programs. Select no candidate unless one bound covers essentially all
observed reuse across the suite. Then compare repeated bytecode walls and
allocations plus compiled startup, Binary Trees GC pacing, TapeLang, and
numeric/text controls. Revert if misses add runtime boxing or the smaller live
heap recreates the prior lazy-cache GC regression. Do not retain profiling
counters, change GC policy, add ballast, specialize named containers, or begin
WASM work.
