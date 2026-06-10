# Compiled callable environment localization retained

Date: 2026-07-24

## Decision

Retain the general compiler rule that localizes package-environment
requirements to the generated entry wrapper that actually establishes the
callee package, and treats successfully statically lowered immutable module
bindings as environment-independent.

The rule has two conservative parts:

- only a raw generated-body call propagates the callee's environment
  dependency to its caller; a package-entry call keeps the dependency inside
  that entry wrapper;
- an immutable module binding that the compiler has already replaced with its
  literal native carrier does not make the enclosing body environment
  dependent. Mutable or otherwise unresolved bindings retain the environment
  requirement.

This extends the existing default-ABI fixed-point proof. It does not change the
execution-context ABI, remove any dynamic adapter, or add a benchmark,
container, or non-primitive nominal rule.

## Attribution

The post-`String.len_bytes` refresh covered Concurrent Document Pipeline,
Concurrent Event Routing, and Concurrent Policy Callbacks. All three were
built with `--no-fallbacks`, ran with the goroutine executor, passed their
public verifiers, and remained interpreter-free.

Generated-source inspection corrected the initial boxed-callable hypothesis:

- Document and Event each inferred a local captured scorer with erased
  `runtime.Value` parameters, then converted it through `runtime.Value` into
  the typed callable accepted by `process_task` or `route_task`.
- Policy already stored and invoked its four-`i64` adjustment callbacks as
  native typed Go functions. It had no equivalent callable boxing in the hot
  path.

The exact category shared by all three was instead a package-environment guard
inside the generated hot closure. Document and Event inherited the requirement
from calls already routed through cross-package entry wrappers. Policy's three
callbacks inherited it solely from the statically lowered immutable
`MODULUS` binding.

An initial immutable-binding-only subcandidate removed Policy's three guards
but did not reach Document or Event, so it did not clear the three-program
bar. Localizing dependencies at existing callee entry wrappers made the same
legal closure-entry category disappear in all three.

## Generated-code effect

After the retained change:

- the hot Document and Event scorer closures no longer emit
  `bridge.SwapEnvIfNeeded`;
- Policy's three typed adjustment closures no longer emit that guard;
- cross-package calls whose bodies need package state still target their
  package-entry wrappers;
- same-package raw-body calls continue to propagate dependency through the
  fixed point;
- mutable module-binding reads, unknown raw callees, dynamic calls, host
  boundaries, and runtime services remain environment-dependent;
- Document and Event still contain their separately attributable erased
  scorer `to_runtime_value` / typed `from_runtime_value` round-trip.

## Repeated A/B gate

Baseline and candidate binaries were each built once and frozen. Five
order-balanced pairs ran on quiet CPU 9 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `ABLE_EXECUTOR=goroutine`. Every one of the
30 Able processes passed its sibling public verifier.

| Application | Baseline samples (s) | Candidate samples (s) | Mean change |
|---|---|---|---:|
| Concurrent Document Pipeline | 0.08, 0.09, 0.09, 0.08, 0.10 | 0.06, 0.06, 0.06, 0.06, 0.06 | 0.088 -> 0.060 (-31.82%) |
| Concurrent Event Routing | 0.52, 0.51, 0.55, 0.53, 0.59 | 0.42, 0.42, 0.43, 0.42, 0.46 | 0.540 -> 0.430 (-20.37%) |
| Concurrent Policy Callbacks | 0.24, 0.25, 0.25, 0.25, 0.25 | 0.08, 0.08, 0.09, 0.09, 0.08 | 0.248 -> 0.084 (-66.13%) |

The geometric-mean improvement is 43.13%. Mean peak RSS was effectively flat:
Document changed 13,710.4 to 13,661.6 KB, Event 27,168.8 to 27,733.6 KB, and
Policy 15,895.2 to 15,569.6 KB.

## Profile confirmation

Five main-phase profiles per candidate were merged. The exact generated
closure guard is absent from all three candidate sources.

| Application | Baseline profile | Candidate profile | Selected closure caller |
|---|---:|---:|---:|
| Concurrent Document Pipeline | 0.41 s | 0.37 s | `work.func1` 0.18 s -> no sample |
| Concurrent Event Routing | 2.27 s | 2.24 s | `work.func1` 0.44 s -> 0.04 s body-only sample |
| Concurrent Policy Callbacks | 1.11 s | 0.41 s | three `make_adjusters` lambdas 0.49 s aggregate -> no sample |

The residual ledger is now clearer:

- Document retains 0.17 s (45.95%) under `__able_call_value`;
- Event retains 0.40 s (17.86%) under `__able_call_value`;
- Policy has no hot boxed callback path; its remaining 0.39 s
  `bridge.currentGID` owner is under channel scheduler payload lookup.

Thus this tranche removes one shared closure-entry crossing without claiming
that Document/Event callable boxing or Policy channel scheduling is solved.

## Equivalent Go comparison

The current source-equivalent Go binaries were built once with `-trimpath` and
each ran five verified high-resolution processes on the same CPU.

| Application | Candidate Able mean | Go samples (s) | Go mean | Able / Go |
|---|---:|---|---:|---:|
| Concurrent Document Pipeline | 0.060 s | 0.004086, 0.003719, 0.003997, 0.002718, 0.003368 | 0.003578 s | 16.77x |
| Concurrent Event Routing | 0.430 s | 0.003978, 0.004782, 0.004755, 0.004309, 0.003966 | 0.004358 s | 98.67x |
| Concurrent Policy Callbacks | 0.084 s | 0.003789, 0.004818, 0.004017, 0.003116, 0.003614 | 0.003871 s | 21.70x |

The candidate clears the breadth and improvement gates but does not meet the
compiled 95%-of-Go goal.

## Artifact identity

| Application | Baseline SHA-256 | Candidate SHA-256 | Go SHA-256 |
|---|---|---|---|
| Concurrent Document Pipeline | `23c5ac49b7655156ffb692be90be6f46a89937c09f86bc9b77dd71d3b7b66364` | `6980f136817e02d0cafe437737322cdddd5cc40ddf088a5678538d8e3f36ee37` | `a46a7558e9fbc14a7204a1e37c35f418aae900134d62433c029414772954aa34` |
| Concurrent Event Routing | `9aef17771985fa0fb01732fb00f97e67090de6a060e4f977b65b3b4a7bde1343` | `32573925aa6a2fb6151dcd75b20b547e1b2945dea7809cb1c57fa1ddffb0c4a3` | `e08c89e8f55ee0e09834e4c4498248dd61cfb5683c542b044151271d37ad16d7` |
| Concurrent Policy Callbacks | `decdda95f28b33eb076771722b72f68ebce99f4adda981a13041b9fe05e4ddc1` | `284b8719ba8751499bfae6b90feba2683748869c43de6f40006d22be2caf705e` | `46bd01493be186dc8cd46230b9f8b503e9db86c2511433e18f679af46bc66585` |

The machine-readable record is
`2026-07-24-compiled-callable-environment-localization-retained.json`.
Temporary raw evidence is under
`/tmp/able-aot-callable-carrier-20260724.170B6M`.

## Verification

Passing bounded guards cover:

- immutable and mutable module-binding closure effects;
- raw-body versus package-entry dependency propagation across three packages;
- imported free functions and inherent methods;
- typed lambda and native callable generated source and execution;
- nested spawn and explicit dynamic-call context construction;
- concurrency fixtures, goroutine await, flush, blocked-task, and mutex parity;
- `go test ./cmd/ablec`;
- strict generated execution and public verification for all three benchmark
  applications.

`TestCompilerPipePlaceholderLambdaExecutes` still fails with
`compiler bridge: missing interpreter`. Reverting this tranche's production
changes reproduces the same failure, while its generated-source guard passes,
so it is a pre-existing residual from the interpreter-package cut rather than
a candidate regression. No assertion failure occurred in the retained
candidate's focused or concurrency gates.

No canonical stdlib, runtime, interpreter, tree-walker, bytecode, language,
dependency, or WASM change was required.

## Next

Census the strict compiled benchmark corpus for the exact local-callable
carrier round-trip seen in Document and Event: a native callable with erased
`runtime.Value` parameters is immediately converted to `runtime.Value` and
back to a compatible concrete typed callable. Select a third unlike
application that genuinely executes the same path, then preserve the callable
as a native Go function through assignment and typed argument flow.

This is next because the residual boxed call is now 45.95% of Document and
17.86% of Event profiles, while Policy demonstrated that a fully typed
callback avoids it. The work entails carrier-flow census, one general
contextual lambda or native-callable adapter rule, dynamic-fallback guards,
and five order-balanced verified A/B pairs against equivalent Go. It is
important because this is the remaining measured point where statically
compatible Able callback signatures unnecessarily stop being native Go.
Do not begin WASM work.
