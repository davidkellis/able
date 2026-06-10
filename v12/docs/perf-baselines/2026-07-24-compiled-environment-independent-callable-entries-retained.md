# Compiled environment-independent callable entries retained

Date: 2026-07-24

## Decision

Retain the general compiler change that propagates package-environment
independence through statically typed callable invocations and generated
lambdas, then selects raw generated Go bodies at proven-independent call
sites.

The rule is semantic and program-independent:

- a typed callable invocation no longer makes its enclosing function
  package-environment dependent by definition;
- a generated lambda receives its own environment-effect record and installs
  an environment guard only when its body or transitive callees require one;
- adapters without a lexical caller package may select a raw generated body
  when the compiler's fixed-point proof marks that body independent;
- unproven functions, package bindings, unknown externs, and package-reading
  lambdas retain their package-entry guards.

This is a narrow extension of the existing default-ABI independence proof. It
does not enable the rejected execution-context ABI, change the callable ABI,
or add a rule for a benchmark, container, or non-primitive nominal type.

## Immediate architectural goal

Compiled Able should keep statically knowable execution in typed generated Go.
Package-entry wrappers, boxed `runtime.Value` carriers, dynamic bridge helpers,
and the interpreter should appear only at genuine dynamic, host/ABI, or
runtime-service boundaries.

The interpreter-package cut proved that strict generated applications can be
fallback-free. It did not prove that every static call avoids the remaining
package/runtime bridge. This tranche closes one shared static call category.

## Boundary refresh

The post-generic-callable, post-interpreter-cut telemetry refresh covered
Option/Result Config, Dependency Wave Validation, Concurrent Document
Pipeline, and, for typed-boundary confirmation, Validated Job Pipeline. Every
application compiled with `--no-fallbacks`, passed its public verifier, and
omitted the interpreter dependency.

The call-path audit reported zero fast-method, generic-union-method, or
generic-union-fallback calls in each application. The dynamic-boundary audit
reported only explicit dynamic, residual polymorphic, host, or runtime-service
sites:

| Application | Explicit dynamic | Residual polymorphic | Host ABI | Runtime service |
|---|---:|---:|---:|---:|
| Option/Result Config | 1 | 1 | 1 | 0 |
| Dependency Wave Validation | 2 | 1 | 1 | 4 |
| Concurrent Document Pipeline | 3 | 1,026 | 1 | 4 |

The remaining material shared runtime owner was not interpreter dispatch or
boxed generic-callable dispatch. Five merged CPU profiles showed repeated
goroutine identity recovery below package-entry guards:

| Application | Profile total | `bridge.currentGID` cumulative | Share |
|---|---:|---:|---:|
| Dependency Wave Validation | 2.94 s | 2.74 s | 93.20% |
| Validated Job Pipeline | 2.87 s | 2.73 s | 95.12% |
| Concurrent Document Pipeline | prior refresh | prior refresh | 95.27% |

Generated-source inspection found the same avoidable owner in all three:
statically typed, package-independent bodies were reached through entry guards
that called `SwapEnvIfNeeded`, which recovered a goroutine ID through
`runtime.Stack`. Examples included generic `Result.map` specializations,
native interface adapters, and parameter-only typed lambdas.

## Generated-code effect

After the change:

- environment-independent `Result.map` and `Option.map` specializations call
  their raw typed Go bodies;
- proven-independent native interface adapters may call their raw
  implementation bodies even when the adapter has no lexical caller package;
- pure typed lambdas no longer emit `bridge.SwapEnvIfNeeded`;
- a lambda that reads a package binding still emits the guard;
- runtime entry wrappers remain generated and available for dynamic callers.

The `Iterator.filter_map` and `Enumerable.reduce` guards independently confirm
the raw-body selection. The still package-dependent `Iterator.collect` path
continues through its entry wrapper.

## Repeated A/B gate

Baseline and candidate binaries were built once, frozen, and measured in five
order-balanced pairs on quiet CPU 14 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, and `GOGC=50`. Every process passed the sibling public
verifier.

| Application | Baseline samples (s) | Candidate samples (s) | Mean change | Verified |
|---|---|---|---:|---:|
| Option/Result Config | 0.05, 0.04, 0.04, 0.04, 0.04 | 0.04, 0.04, 0.03, 0.04, 0.03 | 0.0420 -> 0.0360 (-14.29%) | 5/5 both |
| Dependency Wave Validation | 0.60, 0.57, 0.58, 0.59, 0.61 | 0.33, 0.33, 0.32, 0.32, 0.37 | 0.5900 -> 0.3340 (-43.39%) | 5/5 both |
| Validated Job Pipeline | 0.42, 0.41, 0.42, 0.41, 0.40 | 0.36, 0.34, 0.36, 0.34, 0.35 | 0.4120 -> 0.3500 (-15.05%) | 5/5 both |

Mean GC cycles moved from 9.0 to 9.0, 12.8 to 10.8, and 7.4 to 6.8
respectively.

The three shapes are unlike: serial configuration with Option/Result,
concurrent dependency-wave mapping, and concurrent channel/text validation.

## Profile confirmation

Five candidate profiles per application were merged:

| Application | Baseline total | Candidate total | Baseline `currentGID` | Candidate `currentGID` | Absolute owner change |
|---|---:|---:|---:|---:|---:|
| Dependency Wave Validation | 2.94 s | 1.67 s | 2.74 s | 1.49 s | -45.62% |
| Validated Job Pipeline | 2.87 s | 1.72 s | 2.73 s | 1.59 s | -41.76% |

Option/Result Config produced only 60-90 ms of merged samples and is too sparse
for symbol-level conclusions. Its repeated end-to-end result is retained as
the governing evidence for that row.

The profile confirms that the candidate reduced the selected exact owner. It
also shows that package-entry recovery remains the dominant cumulative cost in
the two concurrent programs, so AOT boundary closure is not complete.

## Equivalent Go comparison

The equivalent Go binaries were also built once and verified in five
high-resolution processes:

| Application | Candidate Able mean | Go mean | Able / Go |
|---|---:|---:|---:|
| Option/Result Config | 0.0360 s | 0.002164 s | 16.64x |
| Dependency Wave Validation | 0.3340 s | 0.002457 s | 135.93x |
| Validated Job Pipeline | 0.3500 s | 0.002232 s | 156.80x |

The change clears the breadth and improvement gates but does not meet the
compiled 95%-of-Go goal. The large remaining ratios make continued static
boundary attribution the governing work.

## Artifact identity

| Application | Baseline SHA-256 | Candidate SHA-256 | Go SHA-256 |
|---|---|---|---|
| Option/Result Config | `03117d4947f1435b80a822be7c4e758e917e3ba413e99ee7105b3c25f605eba7` | `6f2978a22a90f78a2e0978affc81f7874dae83a69aa05f9c1e1c74c211305cf2` | `5d1d88194ad4cb04ebf27498b09d5986c57ec48db51f68cc269d718ea788ab6e` |
| Dependency Wave Validation | `9a1d878210adc912f6b78c2b59fe3d92e5cf998b3887e8bff8222179e4d961a1` | `eb3622753d2413896e410bbd51c202d70c3ca1cf5496a0c57927fa24e5a76e59` | `0f227dac75454acb4cc3ebb460f262d4709c0543bd778916b82109f687d65a64` |
| Validated Job Pipeline | `ac46c865b758274d10f73d61ebbde47a7fa84c54e3eee2bc532ad111694bc4a8` | `0487e3173491c9fa28ec43a6731c98545dccf6c6b0ac5ec94514ef7c4ea01cb4` | `dee94a7d62df04b2ec8d3c2cfbe83f948e74013bf6dec5abe0fa359904c6c683` |

The aggregate machine-readable record is
`2026-07-24-compiled-environment-independent-callable-entries-retained.json`.
Raw temporary evidence was retained under
`/tmp/able-aot-lowering-refresh-20260724.YOsC5D`.

## Verification

Passing bounded guards cover:

- fixed-point environment independence and dependent imported calls;
- pure and package-reading typed lambdas;
- unknown-caller native interface adapters;
- generic named unions and canonical Option/Result specialization;
- Enumerable, Iterator, filter-map, and HashMap native lowering;
- strict generated-code interpreter-root exclusion;
- nested-spawn, mutex, await, goroutine-future, and future-flush parity;
- `go test ./cmd/ablec`.

The initial aggregate generic and concurrency invocations exceeded the
one-minute command timeout. No semantic mismatch was reported. The affected
tests passed when divided into bounded semantic groups; no individual retained
verification exceeded one minute.

No canonical stdlib, runtime, interpreter, language, dependency, or WASM
change was required.

## Next

Refresh exact residual package-entry attribution across at least three unlike
strict compiled applications after this retained change. Select the largest
exact generated caller or semantic category that repeats in all three and
prove whether its transitive package-environment dependency is real or an
avoidable static escape.

This is next because `currentGID` fell 41.76%-45.62% in the profiled concurrent
applications yet remains 89%-92% cumulative. The work entails caller-level
profiles and generated-source tracing, then a narrow extension of static
effect/lowering analysis only if one exact general category clears the
three-program gate. It must retain entry guards for package bindings, dynamic
calls, host ABI, and runtime services, and it must not revive the rejected
broad execution-context ABI. This is important because removing these repeated
static-to-runtime crossings is the most direct remaining route toward
generated Go performance.
