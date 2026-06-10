# Compiled small-union-payload census

Date: 2026-07-18

## Decision

Complete the coverage-wide structural and dynamic gate and retain no compiler,
generated-runtime, canonical-stdlib, benchmark, fixture, or language change.

A generic caller-owned union-result ABI is not admitted. The generated source
contains nine eligible small nominal payload definitions, but exact execution
evidence reduces them to one allocating definition: `Utf8DecodeResult`. The
other eight definitions occur only in regex clone helpers and contribute zero
observed heap allocations in Regex Suffix, Regex Set, and Regex Stream. The
required breadth of at least two material nominal definitions is therefore not
met.

## Restored sub-minute regression guard

The former monolithic `TestCompilerSpecNullableExpectationBuilds` mixed two
expensive concerns and timed out in two independent 60-second launches. It is
now decomposed without dropping either concern:

- the matcher/interface, generic expectation, nullable result, equality
  matcher, and nil matcher shape receives a real generated-Go build; and
- the exact canonical `import able.spec.*` program still receives complete
  stdlib-aware compiler lowering.

The structural build completed in 1.982 seconds. Three independent canonical
lowering launches completed in 50.790, 49.273, and 51.298 seconds, for a
50.454-second mean. Each test is independently below the one-minute rule. This
is test-only decomposition; compiler output and canonical stdlib are unchanged.

## Structural census

One normal compiler executable generated Go for all 35 selected portable
applications. Each source-generation process used the catalog working
directory, canonical external stdlib, normal source-root policy, and a
55-second cap. All 35 completed.

A temporary Go-AST analyzer applied the existing caller-owned-result scalar
eligibility rule: a non-singleton nominal struct with one or two scalar words,
freshly pointer-constructed inside a statically known union result returned by
a compiled Able function. It excluded dynamic wrappers, host adapters,
ordinary pointer results, singleton tags, non-scalar fields, and fresh payloads
nested in some other returned aggregate.

The result splits into three exact groups:

| Generated application group | Applications | Eligible fresh union payloads |
| --- | ---: | --- |
| no qualifying site | 18 | none |
| text/import closure | 14 | `Utf8DecodeResult` at three sites in two functions |
| regex closure | 3 | `Utf8DecodeResult` plus eight regex payload definitions; 11 sites in four functions total |

The 14 text/import-closure applications are QuickSort, Sudoku Masks,
I-Before-E, JSON, PiDigits, Mandelbrot, Reverse Complement, K-Nucleotide,
TapeLang Alphabet, Word Frequency, Document Audit, Lexical Rollup, Channel
Rollup, and Unicode Scalar Pipeline. Static presence is not treated as runtime
materiality.

The regex-only definitions are `AutomataNodeAny`, `AutomataNodeLiteral`,
`AutomataNodeSimpleFoldLiteral`, `NFAAny`, `NFAChar`, `NFAGroupEnd`,
`NFAGroupStart`, and `NFASimpleFoldChar`. Their qualifying sites are confined
to `clone_node(...)` and `regex_program_clone_symbol(...)`. This is a shared
structural rule, not a named-regex eligibility decision; the names are reported
only to make the census auditable.

## Dynamic allocation attribution

Fresh normal binaries for Regex Suffix, Regex Set, and Regex Stream each
received three independent exact main-phase allocation-profile processes.
Every process used one logical CPU, `GOGC=50`, `GOMEMLIMIT=1GiB`, the catalog
working directory and arguments, a 55-second cap, and its public Ruby verifier.
All 9/9 executions verified, and each application produced one stable stdout
SHA-256 across repeats.

| Application | Runs | Mean profiled wall | Main allocations | Main bytes | `Utf8DecodeResult` objects | Payload bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix | 3 | 24.327 s | 6,397,570 | 249,832,824 | 292,567 mean | 4,681,072 mean |
| Regex Set | 3 | 2.445 s | 103,780 | 6,942,016 | 8,624 | 137,984 |
| Regex Stream | 3 | 2.782 s | 240,703 | 8,587,456 | 10,243 | 163,888 |

Main counters are exact and repeated; Regex Suffix bytes varied by only 16
bytes on either side of the reported mean. Exact line-attributed allocation
profile differences report no allocation row for either regex clone helper in
any of the nine processes. Thus all eight additional eligible definitions have
zero measured heap allocation in every dynamic guard. Only
`Utf8DecodeResult`, already identified in three unlike text applications by
the preceding attribution tranche, remains live and material.

The profiler's own serialization allocations are outside the authoritative
main counters and outside the generated payload source-line intersection.

## Verification and cleanup

- 35/35 selected applications completed bounded generated-source census.
- 3/3 regex applications generated and built fresh profile binaries.
- 9/9 exact allocation-profile executions passed their public verifiers.
- 9/9 regex-clone intersections contained no allocation row.
- Focused structural build and canonical lowering tests are independently
  below one minute.
- The temporary analyzer, generated trees, binaries, profiles, and stdout
  captures are removed after this record is written.

## Next recommendation

Run a coverage-wide dynamic feasibility census for statically proven
multi-operation primitive-integer evaluation regions in the bytecode VM.

Why: only 3 of 27 selected bytecode rows currently meet both interpreter
targets. Repeated profiles continue to expose integer extraction, boxing,
operand transport, and dispatch as broad parents, while isolated helper and
stack-carrier changes have been neutral or regressive. The retained typed-float
region demonstrates that batching a proven expression can remove several
dispatches and intermediate materializations without changing the general
operand-stack representation. The union-result branch is now closed and
should not consume another prototype.

What it entails: add temporary lowering counters for eligible primitive
integer expression trees and dynamic execution counts across the selected
suite, split by operation count and integer kind, then remove the counters.
Advance only if multi-operation regions are material in at least three unlike
applications and at least two primitive integer kinds. A prototype would reuse
the float-region proof structure but must implement exact per-kind overflow,
Euclidean division/remainder, shift, and normalization semantics, materialize
one ordinary value at region exit, and retain a cold fallback when runtime
slots violate the static proof. Measure it with repeated preserved-binary A/B
cohorts across admitted applications plus text, float, wide-integer, and
allocation-heavy guards. Continue to defer WASM.
