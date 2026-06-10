# Raw-i32 cache distribution gate

Date: 2026-07-17

## Decision

Reject a smaller eager raw-`i32` cache bound and retain no runtime candidate.
Selected numeric applications materially reuse values through the current
262,143 upper bound. Reducing the bound enough to remove a meaningful fraction
of startup boxes would turn millions of existing cache hits into repeated
interface boxes in K-Nucleotide and tens of thousands in Rational Series.

All threshold counters, JSON output plumbing, runner code, and raw process
artifacts were removed. The cache size, lookup behavior, benchmarks, stdlib,
fixtures, and language remain unchanged.

## Diagnostic protocol

Temporary counters recorded every call to
`bytecodeRawI32SlotCachedValue(...)` after program loading and immediately
before `main`. They did not change cache lookup or returned values. Exclusive
bins ended at -1025, 255, 1,023, 4,095, 16,383, 65,535, 98,303, 131,071,
147,455, 196,607, 262,143, and `MaxInt32`; each process also recorded its exact
minimum and maximum request.

The corpus deliberately mixed:

- selected text/map: Word Frequency;
- selected iterator/text: Document Audit;
- selected numeric Array: Array Slice Window;
- selected nominal numeric: Rational Series;
- selected recursion/control: Fib;
- selected allocation-heavy text/map: K-Nucleotide;
- selected Queue/graph: Dependency Plan;
- focused Array map, split/join, iterator collect, Binary Trees, and TapeLang
  fixture controls.

Eleven external processes passed their Ruby verifiers. Five focused fixtures
ran twice and exited successfully. Rational Series, Word Frequency, Array
Slice Window, Dependency Plan, and all five fixtures produced byte-identical
counter JSON in their second process, so the request distributions are
deterministic. K-Nucleotide's single instrumented process was bounded but took
39.38 seconds; its 25.2 million requests make the selection result decisive
without another expensive duplicate.

## Distribution result

Coverage is the percentage of requests at or below a proposed upper bound:

| Workload | Requests | Maximum | <=65,535 | <=131,071 | <=147,455 | <=196,607 | <=262,143 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Rational Series | 855,324 | 262,143 | 86.4514% | 91.0622% | 92.2149% | 95.6730% | 100.0000% |
| Array Slice Window | 967,121 | 262,138 | 98.5979% | 99.0653% | 99.1821% | 99.5326% | 100.0000% |
| K-Nucleotide | 25,216,889 | 502,690 | 55.8329% | 68.0477% | 71.1014% | 80.2625% | 92.4773% |
| Array Map fixture | 348,026 | 140,073 | 87.9650% | 98.6966% | 100.0000% | 100.0000% | 100.0000% |
| Binary Trees fixture | 1,683,489 | 131,056 | 99.8421% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Word Frequency | 1,118,913 | 128,161 | 99.9999% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Split/Join fixture | 933,801 | 24,000 | 100.0000% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Iterator Collect fixture | 16,006 | 8,000 | 100.0000% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Dependency Plan | 697,619 | 1,031 | 100.0000% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Document Audit | 1,938 | 1,937 | 100.0000% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| TapeLang fixture | 1,659 | 141 | 100.0000% | 100.0000% | 100.0000% | 100.0000% | 100.0000% |
| Fib | 0 | - | - | - | - | - | - |

The admission rule required at least 99.9% coverage in every unlike workload.
A 196,607 upper bound would remove 65,536 eager boxes, only about 262 KB of raw
box payload, but would add misses for:

- 37,010 Rational Series requests, or 4.33%;
- 4,520 Array Slice Window requests, or 0.47%; and
- 3,080,192 K-Nucleotide requests, or 12.21%, in addition to the 1,896,987
  requests already above the current cache.

A 147,455 bound would save more startup objects but is still worse: it covers
only 92.21% of Rational Series and 71.10% of K-Nucleotide, while Array Map needs
values through 140,073. Bounds very near 262,143 could avoid only a small number
of startup objects and do not address the measured 57-58 ms package wall
materially. No common, materially smaller threshold passes.

## Verification and restoration

- 11/11 selected external processes verified with no timeout or failure.
- 10/10 focused fixture processes exited successfully.
- Nine repeated workloads produced identical counter files across processes.
- Focused raw integer, slot, frame, Array index, and comparison tests pass after
  diagnostic removal.
- Focused CLI run and bytecode-stats output tests pass after removal.
- No cache-size candidate was built, so timing changes from atomic diagnostic
  counters are not performance evidence.

## Next recommendation

Attribute high-value cache requests to their production call sites across
Rational Series, K-Nucleotide, Array Slice Window, and Array Map before changing
the cache again.

Why: the cache is serving real repeated values, so representation and sizing
changes both fail. The remaining generic opportunity is to avoid asking for an
interface value when an `i32` can remain in an existing raw stack cell, register,
or typed slot. That could improve execution and reduce dependence on the eager
cache rather than trading startup against runtime allocations.

What it entails: add temporary call-site counters around the existing cache
requests, not stack inspection or workload detection. Require one producer or
transport boundary to dominate in at least three unlike programs before
building a candidate. Any candidate must preserve ordinary boxed/value
semantics at escape boundaries and pass repeated text/map, iterator, numeric
Array, recursion, Binary Trees, TapeLang, and compiled-startup controls. Remove
all counters afterward and reject a split caller distribution rather than
adding per-program or named-container paths.
