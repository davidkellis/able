# Compiled primitive runtime boundary gate

Date: 2026-07-17

## Decision

Reject the interpreter-independent primitive operator package and restore the
two generated references to `interpreter.ApplyBinaryOperatorFast` and
`interpreter.ApplyUnaryOperatorFast`.

The candidate was generic and semantically clean. It removed the final static
interpreter import roots, but it failed the allocation-heavy application guard.
No production code or test from this candidate remains. The previously retained
compiled-runtime interface boundary remains unchanged.

## Candidate and semantic proof

The temporary `runtimeops` package implemented stable integer, float, string,
comparison, bitwise, shift, and unary operations without importing the
interpreter. Dynamic and nominal operations still fell through the bridge.
Direct parity tests covered:

- every signed and unsigned integer suffix, including `isize` and `usize`;
- mixed signed/unsigned promotion;
- wide `u64`, `i128`, and `u128` values;
- arithmetic, Euclidean division/remainder, comparisons, bitwise operations,
  shifts, overflow, division by zero, and interface-wrapped primitives;
- float, string, NaN, unary negation, and bitwise-not behavior.

The runtimeops parity suite, compiler bridge suite, static/no-bootstrap tests,
dynamic callback fallback, and generated-main selection tests all passed. Every
measured external application output passed its Ruby verifier.

## Dependency and startup result

The experiment proved that the boundary is technically sufficient:

- `go list -deps` for a static generated application omitted
  `able/interpreter-go/pkg/interpreter`;
- the generated binary contained zero interpreter-named symbols;
- the Noop binary fell from 13,985,560 to 8,414,632 bytes (39.8% smaller);
- interpreter initialization disappeared from `inittrace`;
- the remaining runtimeops initialization was 0.009 ms, 4,648 bytes, and 112
  allocations;
- a fresh five-process Noop cohort averaged 0.0300 seconds and zero GC cycles,
  versus 0.0660 seconds and three cycles in the retained state.

These are real startup and footprint improvements, but they are not sufficient
to admit a change that slows a sustained application.

## Repeated application gate

Candidate and baseline binaries were built from the same generated source. The
baseline copy changed only the primitive helper import/calls back to the current
interpreter package. Runs alternated order under `GOMEMLIMIT=1GiB`, `GOGC=50`,
and `GOMAXPROCS=1`. Each process had a 60-second cap and its stdout was checked
by the catalog Ruby verifier.

| Application | Samples per side | Baseline mean | Candidate mean | Delta | Baseline GC | Candidate GC |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | 5 | 0.074 s | below 0.01 s timer resolution | large win | 3.8 | 1.0 |
| TapeLang Alphabet | 5 | 4.004 s | 4.048 s | +1.10% | 3.0 | 0.0 |
| Binary Trees | 5 | 31.534 s | 33.132 s | **+5.07%** | 173.8 | 250.2 |

Binary Trees was extended to five pairs because one baseline process was noisy.
Its candidate range was narrow (32.68-33.64 seconds), and all five candidate
runs performed 249-252 collections. The baseline performed 173-175 collections;
its wall-time outlier does not affect the stable GC diagnosis. Median wall time
also rejects the candidate: 33.10 seconds versus 30.83 seconds (+7.36%).

The candidate reduced average Binary Trees peak RSS from 279,758 KiB to 252,918
KiB, but the smaller initial live heap lowered Go's GC goal and caused 44% more
collections. This independently reproduces the earlier lazy-cache gate. Keeping
unused heap ballast, changing GC policy, or selecting behavior by workload would
violate the project performance guardrails.

## Restored-state verification

After the full candidate revert, these focused suites pass:

```text
go test ./pkg/compiler/bridge -count=1 -timeout 60s
go test ./pkg/compiler -run 'TestCompilerStaticGeneratedCodeRootsLimitedToOperators|TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerDynamicBoundaryCallbackRoundtrip|TestCompilerNoBootstrapStaticFixturesStayBoundaryClean' -count=1 -timeout 60s
```

## Next direction

Do not continue splitting static interpreter roots yet. First reduce sustained
generated-program allocation/GC cost through a general compiler/runtime
mechanism that benefits unlike nominal-heavy applications without changing Able
identity. The next bounded step is an interprocedural non-capture/effect
feasibility audit over existing call and nominal-result lowering. It should
identify which facts are already available, add proof fixtures for retained-old
values and captured parameters, and build a candidate only if the same safe
caller-owned result-storage opportunity appears in multiple unlike programs.
This addresses a large measured compiler wall and supplies the safety proof that
the previously rejected nominal-reuse experiment lacked.
