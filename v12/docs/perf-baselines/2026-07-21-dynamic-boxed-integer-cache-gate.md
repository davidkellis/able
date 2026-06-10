# Dynamic boxed-integer cache gate

## Decision

The dynamic boxed-integer cache is a material shared VM wall, but the tested
generic lock simplification does not clear the broad benchmark bar. No
production runtime change is retained.

The ordinary package-global `sync.RWMutex` and bounded per-kind maps remain in
place. The only retained change is test-harness plumbing: the fresh-process
retention probe can now profile exactly one `main` call and records its
`main_duration_ns`, excluding load/lowering and final collections from paired
wall comparisons.

## Protocol

All application probes used separate processes with `GOMAXPROCS=1`,
`GOGC=50`, `GOMEMLIMIT=1GiB`, skipped benchmark typechecking, and canonical
`able-stdlib`. Each application ran through the ordinary loader and bytecode
interpreter. Diagnostic counts were collected with the existing opt-in
`able_bytecode_box_reuse` build; CPU profiles and timings used builds without
the diagnostic event mutex.

The retention harness resets diagnostic counts after program evaluation and
before `main`. CPU profiling now starts at that same boundary. The wall timer
surrounds only `interp.CallFunction(mainValue, nil)`.

## Reuse census

| Application / kind | Lookups | Hits | Inserts | Capacity misses | Interpretation |
| --- | ---: | ---: | ---: | ---: | --- |
| Fixed Width 128 / `i32` | 737,856 | 0 | 262,144 | 475,712 | high-cardinality stream |
| Fixed Width 128 / `u64` | 1,625,001 | 0 | 262,144 | 1,362,857 | high-cardinality stream |
| Reverse Complement / `i32` | 3,804,572 | 2,033,361 | 262,144 | 1,509,067 | 53.45% hits, then saturation |
| Rational Series / `i128` | 16,021 | 12,652 | 3,369 | 0 | 78.97% hits |
| K-Nucleotide / `i32` | 5,690,960 | 5,450,412 | 240,548 | 0 | 95.77% hits |
| K-Nucleotide / `u64` | 6 | 2 | 4 | 0 | immaterial auxiliary kind |

Rational Series also made 5,494 intentional large-`i64` bypasses;
K-Nucleotide made 25. Matrix Multiply made no dynamic-cache request, while
Word Frequency made 13,633 large-`i64` bypasses and did not populate a cache.
Those two programs are therefore negative guards rather than evidence for a
map mechanism.

The census establishes three unlike reuse-heavy consumers, while Fixed Width
proves that unconditional admission also serves a large no-reuse stream. A
candidate therefore had to improve both hits and misses without specializing
on an application, nominal type, or integer value distribution.

## Clean CPU attribution

One-process, measured-main profiles attribute
`bytecodeBoxedIntegerValue(...)` as follows:

| Application | CPU duration | Boxed-integer cumulative | `mapaccess2_fast64` cumulative |
| --- | ---: | ---: | ---: |
| Fixed Width 128 | 7.39 s | 9.62% | 7.18% |
| Reverse Complement | 3.61 s | 15.51% | 7.76% |
| Rational Series | 4.41 s | 2.72% | 1.13% |
| K-Nucleotide | 42.76 s | 1.36% | 2.23% |

This clears the three-unlike-program materiality rule. It also identifies the
exact shared owner as bounded-map synchronization and lookup, rather than the
fixed/extended eager cache or an application-specific numeric operation.

## Candidate

The reversible prototype replaced the global `RWMutex` read-then-write miss
transaction with one ordinary `Mutex` transaction and one map lookup. It
preserved bounded growth, stable cached value identity, concurrent safety, and
all integer-kind semantics. It was deliberately generic: hits and misses for
every dynamically cached integer kind used the same mechanism.

Focused cache tests and race-enabled cache tests passed. The candidate was then
frozen into a test binary; an otherwise identical restored-`RWMutex` binary
served as baseline. Fresh-process pairs alternated order.

| Application | Pairs | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 5 | 7.8160 s | 7.8043 s | -0.15% |
| Reverse Complement | 5 | 3.1645 s | 3.1293 s | -1.11% |
| Rational Series, first cohort | 5 | 3.8554 s | 4.2216 s | +9.50% |
| Rational Series, confirmation | 5 | 3.9479 s | 3.8307 s | -2.97% |
| Rational Series, pooled | 10 | 3.9017 s | 4.0262 s | +3.19% |

The independent Rational cohort reversed direction, demonstrating workstation
volatility, but the requested arithmetic mean across all ten pairs remains a
3.19% regression. Fixed Width and Reverse Complement are effectively neutral.
One preliminary K-Nucleotide pair favored the candidate, but the prototype had
already failed its independent guard; that single pair is not promoted as
timing evidence and a long second cohort was not run.

An ordinary mutex would also give up concurrent cache-hit reads without a
material single-process win. The candidate is therefore rejected and fully
removed rather than traded for a benchmark-local improvement. No `sync.Map`,
admission heuristic, enlarged fixed cache, named-container rule, GC policy, or
WASM change was attempted.

## Verification

After restoration:

```text
go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s
ok   able/interpreter-go/pkg/interpreter  24.072s
```

The benchmark harness remains under the project limit at 952 lines.

## Next

Do not iterate on dynamic-cache lock or map variants. The exact wall is real,
but the simplest general mechanism was neutral-to-regressive and the reuse
distributions conflict. The next tranche should refresh architecture-level
selection across high-excess bytecode applications and require one semantic
operation—not a Go runtime parent or previously closed raw/boxed integer,
frame, call/return, member, dispatcher, or cache variant—to be material in at
least three unlike applications. If none appears, shift the same selection
discipline to generated compiler code. This entails new bounded main-only
profiles and operation/owner reconciliation before any implementation, because
the remaining distance to the product targets is too large for another noisy
cache micro-variant.
