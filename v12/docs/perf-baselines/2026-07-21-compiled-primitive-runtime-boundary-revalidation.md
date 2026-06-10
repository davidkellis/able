# Compiled primitive-runtime boundary revalidation

Date: 2026-07-21

## Decision

Reject the interpreter-independent primitive operator package again and keep
the current generated calls to `interpreter.ApplyBinaryOperatorFast(...)` and
`interpreter.ApplyUnaryOperatorFast(...)`.

The recovered candidate remained generic and semantically clean. It removed
the final static interpreter dependency, approximately 39% of the smallest
generated binaries, four startup GC cycles, and roughly 38 MiB of startup RSS.
However, it again changed Go GC pacing for a sustained allocation-heavy
application. Verified Binary Trees arithmetic means regressed under both the
historical single-CPU stress guard and the current four-CPU/goroutine execution
policy. No compiler, runtime, VM, stdlib, workload, or verifier change remains.

## Recovery and semantic proof

The exact July 17 candidate and its follow-up parity tests were recovered from
the local Codex session log rather than reconstructed from memory. The
temporary `runtimeops` package covered stable integer, float, string,
comparison, bitwise, shift, and unary operations while leaving dynamic and
nominal fallback on the existing compiler bridge.

Parity tests against the current interpreter fast path passed for:

- all twelve fixed-width and pointer-width signed/unsigned integer kinds;
- mixed signed/unsigned promotion and wide `u64`, `i128`, and `u128` values;
- arithmetic, Euclidean division/remainder, comparison, bitwise, shift,
  overflow, division by zero, and interface-wrapped primitive behavior; and
- floats, strings, NaN comparisons, unary negation, and bitwise not.

Focused static-launcher, no-bootstrap, dynamic callback, and bridge tests also
passed. A freshly generated static Noop application depended on
`able/interpreter-go/pkg/runtimeops` but not
`able/interpreter-go/pkg/interpreter`; its binary contained no
interpreter-named symbols.

## Source-matched artifacts

Noop, Array Slice Window, TapeLang Alphabet, and Binary Trees were generated
once with the candidate. The candidate binary was preserved, then a baseline
was built from the same generated source after changing only the runtimeops
import and its two calls back to the current interpreter package. This excludes
compiler/source drift from the comparison.

| Application | Baseline bytes | Candidate bytes | Change |
| --- | ---: | ---: | ---: |
| Noop | 14,003,520 | 8,420,904 | -39.9% |
| Array Slice Window | 15,470,608 | 9,722,216 | -37.2% |
| TapeLang Alphabet | 20,064,536 | 13,734,272 | -31.5% |
| Binary Trees | 14,261,536 | 8,695,136 | -39.0% |

## Startup and allocation-light results

Runs alternated baseline/candidate order. They used `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`, and every non-Noop output passed its external
Ruby verifier. Arithmetic means follow the workstation measurement policy.

| Application | Samples per side | Baseline mean | Candidate mean | Baseline GC | Candidate GC | Baseline RSS | Candidate RSS |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Noop | 10 | 0.0600 s | below 0.01 s timer resolution | 4.00 | 0.00 | 45,021 KiB | 6,910 KiB |
| Array Slice Window | 7 | 0.0600 s | below 0.01 s timer resolution | 4.00 | 1.00 | 47,467 KiB | 9,773 KiB |
| TapeLang Alphabet | 5 | 3.9760 s | 3.6640 s (-7.85%) | 3.00 | 0.00 | 47,934 KiB | 10,158 KiB |

TapeLang was noisy but favored the candidate in four of five pairs. It was not
extended after both Binary Trees contracts produced a stable structural
rejection signal.

## Binary Trees decision gate

The historical stress contract reproduced the same mechanism as the earlier
rejection:

| Contract | Samples per side | Baseline mean | Candidate mean | Delta | Baseline GC | Candidate GC |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 CPU, `GOGC=50`, 1 GiB limit | 3 | 30.1767 s | 32.9267 s | **+9.11%** | 172.67 | 250.67 |
| 4 CPUs, goroutine executor, default GC policy | 3 | 6.5667 s | 6.9767 s | **+6.24%** | 64.00 | 86.33 |

Every one of the six pairs moved against the candidate. GC counts were tightly
clustered: 172-173 versus 250-252 in the stress lane and exactly 64 versus
86-87 in the current mode-aware lane. The candidate lowered peak RSS by about
24%-25%, but removing the interpreter initializer also lowered the initial live
heap and the Go GC goal, causing 35%-45% more collections. This is the same
causal mechanism seen on July 17, now confirmed under the current four-CPU
execution policy as well as the historical stress guard.

The fact that current Binary Trees is much faster under its normal parallel
scorecard contract did not make the package cut safe. Its relative regression
survived and remains too large for a broadly applicable optimization.

## Restored verification

After removing the complete candidate, the retained tree again has only the two
documented generated interpreter roots. These commands pass:

```text
go test ./pkg/compiler/bridge -count=1 -timeout 60s
go test ./pkg/compiler -run 'TestCompilerStaticGeneratedCodeRootsLimitedToOperators|TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerDynamicBoundaryCallbackRoundtrip|TestCompilerNoBootstrapStaticFixturesStayBoundaryClean' -count=1 -timeout 60s
go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s
```

## Next recommendation

Do not retry interpreter-package isolation, unused-helper omission, cache
packing, heap ballast, or GC-policy adjustment. Build a generated-allocation
shape matrix over current compiled misses before selecting another compiler
candidate.

Why: the remaining static interpreter dependency is now proven to be a large
startup/footprint cost that cannot be removed safely until sustained generated
programs allocate less. The latest exact-leaf sweep saw shared Go GC parents
but different application-named descendants; exact symbol intersection alone
cannot tell whether different generated functions still share one compiler-
induced escape shape such as interface boxing, variadic argument storage,
closure capture, or temporary backing allocation.

What it entails: preserve generated sources for a fresh unlike subset of the
current compiled misses, combine main-only allocation attribution with Go
escape-analysis output, and classify escapes by lowering mechanism rather than
by generated symbol name. Advance only if the same compiler-controlled shape
is material in at least three unlike applications and can be removed under the
general primitive/nominal rules. Then require source-matched repeated means for
those applications plus Binary Trees, TapeLang, short-startup, dynamic-fallback,
and bytecode controls. This directly attacks the GC wall that blocked this
boundary without introducing a named type, benchmark, application, executor,
GC, or WASM special case.
