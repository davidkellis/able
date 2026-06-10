# Compiled iterator/control and dynamic-i64 lazy-slot gate

## Decision

Retain one general primitive-boundary optimization. The generated dynamic
`i64` cache now initializes each common-value slot independently instead of
allocating all 4,224 boxed values on the first common `i64` conversion.
The represented range, returned `runtime.Value` semantics, uncommon-value
fallback, and generated-site selection are unchanged. Each slot uses
`sync.Once`, preserving concurrent first access without a map, a named nominal
rule, or application-specific policy.

No canonical-stdlib, language, workload, reference, bytecode VM, or WASM code
changed.

## Current iterator/control selection

The current compiler built preserved binaries for Array Slice Window,
Dependency Plan, Document Audit, Lexical Rollup, and Option/Result Config.
Every launch used the catalog run directory and arguments, CPU 0,
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a 55-second process cap.
All outputs passed their public Ruby verifiers.

CPU-only main-phase evidence merged 100 independent profiles for each of the
first four programs and 50 for Option/Result: 450/450 verified processes.

| Application | Main CPU samples | Material exact owners |
| --- | ---: | --- |
| Array Slice Window | 790 ms | rolling checksum 82.28% cumulative; `Array.slice` 43.04%; checked multiply 21.52% flat |
| Dependency Plan | 90 ms | deployment resolver 100%; `bridge.ToInt` 33.33%; Queue/Deque and graph work |
| Document Audit | 80 ms | generator 62.50%; first `ToDynamicI64` cache initialization 37.50%; substring/iterator work |
| Lexical Rollup | 810 ms | iterator next 25.93%; generator 18.52%; channel receive 11.11%; environment swap 8.64% |
| Option/Result Config | 6.11 s | static generic-union calls 58.43%; call-value fast dispatch 43.04%; allocation 58.76% |

The raw five-way exact intersection contains only main wrappers and Go
allocation/GC helpers. Caller reconciliation splits `runtime.makeslice` among
required Array slice backing, String splitting, and generic call-conversion
slices. It splits `runtime.convT` between the dynamic-`i64` cache in Document
and Lexical and generic union/function conversions in Option/Result.

The new admissible owner appears after joining this evidence with the retained
dynamic-`i64` tranche: `initDynamicI64Boxes` is exact in Document and Lexical,
while K-Nucleotide, Inventory Reconciliation, and Validated Job previously
established the same compiler boundary. That supplies five unlike programs,
not a manufactured third benchmark.

## Allocation proof

Separate allocation-only phase processes supplied authoritative main-phase
counter deltas. The candidate only changes programs that reach a common
dynamic `i64` value.

| Application | Baseline bytes / allocations | Candidate bytes / allocations | Change |
| --- | ---: | ---: | ---: |
| Array Slice Window | 1,441,352 / 24,012 | 1,441,352 / 24,012 | identical |
| Dependency Plan | 475,192 / 18,631 | 475,192 / 18,631 | identical |
| Document Audit | 645,784 / 6,091 | 373,432 / 1,952 | -42.2% bytes, -68.0% allocations |
| Lexical Rollup | 2,437,696 / 30,377 | 2,161,856 / 26,166 | -11.3% bytes, -13.9% allocations |
| Option/Result Config | 57,447,912 / 1,630,660 | 57,448,064 / 1,630,661 | effectively identical |

The removed Document/Lexical owner is the eager 270 KiB/4,225-object cache
initialization. The retained per-slot form creates only values actually
requested and keeps repeated common-value calls allocation-free.

## Repeated wall and guard gate

Preserved baseline/candidate binaries ran in alternating order. Every process
completed and verified; all samples, including workstation outliers, remain in
the arithmetic means. Inventory received ten extra pairs after its first five
pairs moved against the candidate. K-Nucleotide and Validated Job received five
extra confirmation pairs as additional original-boundary guards.

| Application | Pairs | Baseline mean | Candidate mean | Candidate change | Mean GCs |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 10 | 0.0600 s | 0.0590 s | -1.67% | 4.00 -> 2.90 |
| Lexical Rollup | 10 | 0.0810 s | 0.0750 s | -7.41% | 4.20 -> 3.30 |
| Array Slice Window | 5 | 0.0600 s | 0.0600 s | 0.00% | 4.00 -> 4.00 |
| Dependency Plan | 5 | 0.0600 s | 0.0600 s | 0.00% | 4.00 -> 3.60 |
| Option/Result Config | 5 | 0.1860 s | 0.1880 s | +1.08% | 7.00 -> 7.00 |
| K-Nucleotide | 10 | 2.8030 s | 2.7940 s | -0.32% | 40.60 -> 42.50 |
| Inventory Reconciliation | 15 | 0.1827 s | 0.1780 s | -2.55% | 5.07 -> 4.07 |
| Validated Job Pipeline | 10 | 2.7330 s | 2.7210 s | -0.44% | 9.00 -> 9.00 |
| TapeLang Alphabet | 5 | 3.7300 s | 3.7100 s | -0.54% | 4.00 -> 3.00 |
| Binary Trees, four CPUs | 3 | 10.1367 s | 9.5933 s | -5.36% | 150.00 -> 151.67 |

Option/Result's two-millisecond difference is below the timer resolution and
comes with identical main allocation/GC behavior. Binary Trees and TapeLang do
not attribute their wall differences to this boundary and are treated only as
no-regression guards. K-Nucleotide performs about two more collections but its
ten-pair wall mean remains neutral-to-favorable; the much larger cache-object
removal does not create a broad wall regression.

The wall gate contains 156 verified paired processes. Including selection
profiles, build smokes, and allocation processes, 621 generated-binary
processes verified in this tranche.

## Verification

- Race-enabled complete compiler-bridge tests pass in 1.153 s.
- Focused generated runtime-value boundary tests pass.
- Profile-hook tests pass.
- Full `TestBytecode` passes in 24.335 s.
- Every current, candidate, and paired application output passed its external
  verifier.

## Next recommendation

Refresh the complete compiled product scorecard on the retained per-slot cache
while reusing the same-day fresh Go references. Run five verifier-backed Able
processes for all 45 selected compiled applications under catalog contracts,
then rebuild stability and the cross-mode frontier.

Why: the retained change improves two newly exposed iterator consumers and
touches the compiler bridge linked by every generated binary. Focused guards
show no regression, but the promoted product ratios still describe the eager
cache. A complete Able-side refresh is the honest way to quantify the broad
effect and detect any threshold crossing; rebuilding unchanged Go references
again would add workstation noise rather than source freshness.
