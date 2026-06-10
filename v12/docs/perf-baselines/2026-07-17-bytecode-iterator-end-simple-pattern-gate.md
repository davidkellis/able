# Bytecode IteratorEnd simple-pattern gate — 2026-07-17

## Decision

Keep a simple typed-pattern check for the core `IteratorEnd` iteration
sentinel. Bytecode lowering now records `IteratorEnd` in the existing simple
type-check metadata, and the VM can decide the common sentinel match or
non-sentinel miss before entering general nominal/type matching.

This is a language-level iteration protocol boundary, not a special case for
an application, benchmark, or stdlib container. Section 14 of the v12 spec
defines `IteratorEnd` as the singleton that terminates all core iterators.
Interface wrappers still defer to the semantic matcher, and the existing
exact-definition path remains ahead of the simple check. No compiler, stdlib,
fixture, or language source changed.

## Profile and category admission

Fresh baseline profiles used one preserved test binary, canonical external
`able-stdlib`, CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, skipped
benchmark typechecking, and a separate bounded process per program.

| Workload | Iterations | ns/op | CPU samples | Typed-pattern cumulative |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 3 | 540,805,345 | 2.41 s | no samples |
| Unicode Scalar Pipeline | 1 | 4,481,239,543 | 4.64 s | 15.95% |
| Run-length encode | 1 | 1,553,913,081 | 1.72 s | 27.33% |
| String Split/Join | 1 | 1,243,383,377 | 1.41 s | 12.06% |
| Iterator Collect | 3 | 549,441,230 | 2.45 s | 8.16% |
| Numeric Array Map | 20 | 104,605,283 | 2.38 s | 5.88% |

Temporary exact branch/type counters then classified one main call per
program. Boolean executed no typed-pattern jumps. Split/Join was dominated by
its own `Utf8DecodeResult`, `StringEncodingError`, and `Error` result patterns;
Array Map used no-runtime-value decisions; and the remaining generic Iterator
Collect categories did not repeat broadly. The one material category shared
by three unlike programs was the core iteration sentinel:

| Workload | IteratorEnd matches | IteratorEnd misses |
| --- | ---: | ---: |
| Unicode Scalar Pipeline | 16 | 1,769,472 |
| Run-length encode | 48 | 960,000 |
| Iterator Collect | 12 | 94,000 |
| **Total** | **76** | **2,823,472** |

All counters and diagnostic binaries were removed after admission.

## Candidate and rejected placement

`bytecodeSimpleTypeCheckForName(...)` now recognizes `IteratorEnd`.
`bytecodeMatchSimpleTypedPattern(...)` handles that check before raw-integer
extraction, slot-cell reading, or value materialization. It recognizes the
runtime sentinel and compatible singleton/instance representations, returns a
definite miss for ordinary values, and defers interface wrappers to the
existing full matcher.

The first candidate placed this branch after `bytecodeSlotReadValue(...)`.
That was semantically correct but caused rejected iterator elements to be
materialized before the definite miss: Iterator Collect rose from about
214,686 to 276,342 allocations/op, and allocation profiles doubled
`bytecodeStackSnapshotValue(...)` objects from 30,830 to 61,657 per benchmark
process. Moving the already-decided sentinel check ahead of materialization
restored repeated allocation pairs to 214,684–214,686 allocations/op. The
allocation-producing placement was not retained.

## Repeated workstation gate

Every timing is an independent process. Pairs alternated baseline-first and
candidate-first order; every valid workstation outlier remains in the mean.
The short, volatile Array Map control was expanded to ten pairs.

| Workload | Pairs | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 659.090 ms | 664.147 ms | 0.77% slower; neutral |
| Unicode Scalar Pipeline | 5 | 4.51470 s | 3.82484 s | 15.28% faster |
| Run-length encode | 5 | 1.43016 s | 1.11512 s | 22.03% faster |
| String Split/Join | 5 | 1.17372 s | 1.21280 s | 3.33% slower; guarded |
| Iterator Collect | 5 | 530.776 ms | 491.218 ms | 7.45% faster |
| Numeric Array Map | 10 | 109.015 ms | 102.853 ms | 5.65% faster |

The Boolean mean retains its 833.831 ms baseline and 853.077 ms candidate
outliers. Split/Join retains the 1.333 s candidate sample. Array Map retains
its full roughly 92.6–122.7 ms candidate spread. No control crosses the 5%
broad regression guard.

Allocation spot checks support no allocation-reduction claim. Unicode is
exactly unchanged at 5,226,407 allocations/op and Array Map at 14,646.
Run-length differs by 15 setup-sensitive allocations and Iterator by two;
allocated bytes remain effectively unchanged.

## Mechanism and correctness

Final candidate profiles reduce cumulative typed-pattern execution from
15.95% to 4.82% in Unicode, 27.33% to 8.38% in Run-length, and 8.16% to 3.17%
in Iterator Collect. General fallback matching falls from 11.42% to 2.13% in
Unicode, 23.84% to 5.24% in Run-length, and disappears from the Iterator
profile. The other residual pattern categories remain below the three-program
admission rule and received no candidate.

Focused typed-pattern, matching, IteratorEnd, iterator-collect, type-match,
coercion, canonicalization, bytecode VM, primitive-kernel, runtime, alias, and
fixture-parity tests pass. The broadest parity/alias group completed in
55.833 seconds, below the one-minute test cap. All touched source files remain
below 1,000 lines and `git diff --check` passes.

## Next recommendation

Reconcile the residual bytecode call-execution family across Unicode,
Run-length, Split/Join, Iterator Collect, and Numeric Array Map, centered on
`execCallOpcode(...)`.

Why: clean post-change profiles put call execution at 27.81%–50.26%
cumulative in those five unlike programs, while typed-pattern execution is now
substantially reduced. The visible call descendants differ—cached name calls,
iterator/member calls, static-member calls, Array calls, native calls, and
inline frame setup—so the aggregate alone does not justify another call fast
path.

What it entails: collect temporary counts and bounded call trees by opcode,
resolved callable kind, arity, receiver carrier, inline/native path, argument
materialization, and result transport. Admit a candidate only for one material
child repeated in at least three unlike programs, then run the same independent
paired workload gate with all outliers retained. Preserve dynamic dispatch,
aliases, coercion, user-defined callables, and both Go interpreters; do not add
a named-container rule, and continue to defer WASM.
