# Compiled Imported Environment-Independent Raw Calls

Date: 2026-07-16

## Outcome

Retain the conservative interprocedural environment-independence proof and
the resulting imported raw-body call selection.

Compiled functions still expose both a raw `__able_compiled_*` body and an
environment-restoring `__able_compiled_entry_*` wrapper. Runtime adapters and
unproven imported callers continue to use the wrapper. A source-level static
caller may now use an imported raw body only when a greatest-fixed-point proof
shows that the callee and every transitive generated static callee are
independent of package environment state.

## Proof boundary

The proof starts pessimistically. It excludes functions with package/module
binding reads or writes, unresolved environment lookup, dynamic named/value or
method calls, callable-value application, lambdas, spawn/await, runtime lookup
helpers, or extern bodies. It also scans generated bodies for direct compiled
callees that bypass the ordinary named-call lowering path, so error/interface
method helpers participate in the same fixed point. Newly discovered generic
specializations are analyzed before the fixed point closes.

Same-package static calls remain raw as before. Cross-package raw selection is
only an optimization; every entry wrapper is still emitted for runtime and
dynamic consumers.

Focused tests cover:

- direct and transitive pure imports;
- imported aliases and generic specializations;
- checked arithmetic/error propagation;
- a local parameter shadowing a package binding;
- immutable and mutable package bindings;
- dynamic callable application and lambdas;
- unknown Go externs;
- imported calls inside spawned work; and
- retained runtime entry wrappers.

## Repeated A/B gate

The baseline and candidate binaries came from the same generated packages.
The baseline changed only the hot imported `hypot`/`sqrt` call sites back to
their entry wrappers. Each binary passed the public benchmark verifier before
timing. Twenty pairs per application alternated which variant ran first.

| Application | Baseline mean | Candidate mean | Change | Baseline CV | Candidate CV |
| --- | ---: | ---: | ---: | ---: | ---: |
| Distance Field | 70.144 ms | 41.736 ms | -40.50% | 4.02% | 12.00% |
| RMS Norm | 70.531 ms | 41.403 ms | -41.30% | 3.23% | 9.99% |
| NBody | 166.307 ms | 101.746 ms | -38.82% | 6.85% | 4.62% |

The retained means are roughly 3.88x, 4.38x, and 3.32x the previously
refreshed matched Go means for Distance, RMS, and NBody respectively. This is a
large compiler improvement, but it does not yet meet the 1.05x target.

The normal external harness then ran five independent verified processes for
each of Distance Field, RMS Norm, NBody, and the environment-sensitive
I-Before-E guard. All 20 processes passed. Reported means were 0.064, 0.066,
0.134, and 0.082 seconds respectively; the guard remains faster than its stored
Ruby and Python references and 1.64x its stored Go reference.

## Profile and build observations

The old environment-swap/atomic-store wall disappears from the refreshed
profiles. Short Distance and RMS profiles land entirely in generated math/main
code at 10 ms sampling resolution. NBody records 71.4% in generated `advance`
and 28.6% in generated `sqrt`; there is no `SwapEnvIfNeeded`, restore, or
atomic environment store sample.

Five fresh Distance compiler-generation processes average 1.030 seconds. Its
generated `compiled.go` is 1,859,354 bytes. Entry wrappers remain present, so
the optimization does not remove the runtime ABI or trade runtime correctness
for code-size reduction.

## Verification

- Focused compiler tests for environment proof, imports, aliases, generics,
  shadowing, package bindings, primitive helpers, and spawned calls pass.
- Compiler bridge and `cmd/ablec` tests pass.
- Focused interpreter call/return/hash tests pass.
- `go build ./cmd/able ./cmd/ablec` passes.
- Strict no-fallback generated binaries for the three math applications pass
  their public verifiers.
- The broad compiler `-short` package command still exceeds the mandated
  one-minute package limit in unrelated sequential test setup; bounded focused
  groups pass within the limit.

## Next recommendation

Refresh bounded bytecode profiles for Distance Field, RMS Norm, and a reduced
or steady-state NBody lane, then target the first generic float/static-call
cost that repeats across all three.

Why: native sqrt reduced the bytecode applications sharply, but Distance and
RMS still take about 6.3 and 6.0 seconds against sub-second Python/Ruby
references, while full bytecode NBody remains beyond the bounded process
window. The compiler's former shared environment wall is gone; the largest
remaining gap with the clearest cross-program evidence is now bytecode numeric
execution.

What it entails: collect one-process CPU/call/allocation profiles under the
existing OOM and one-minute guardrails; use reduced or steady-state NBody only
to make the same operation observable; compare raw f64 extraction/storage,
static call setup/return, and type-match costs; implement a candidate only if
the same concrete VM operation is material in all three; then gate repeated
normal processes plus unlike float-heavy and error-domain guards. Continue to
defer WASM.
