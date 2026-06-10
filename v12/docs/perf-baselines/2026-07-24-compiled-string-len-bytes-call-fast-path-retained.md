# Compiled String length call fast path retained

Date: 2026-07-24

## Decision

Retain the compiler-owned call-site lowering for statically resolved canonical
primitive `String.len_bytes()` calls.

The generated call:

1. evaluates the receiver exactly once;
2. checks the same UTF-8 and maximum-length conditions as the canonical
   compiled method;
3. returns Go's native `len` result directly for the valid native String path;
4. enters the original package-aware compiled method for invalid or oversized
   values.

The fallback preserves `StringEncodingError`, package ownership, and all
existing dynamic semantics. The rule applies only to the language primitive
String method identified by package, receiver, signature, and self semantics.
It does not apply to `StringBuilder`, another nominal type with a method named
`len_bytes`, a benchmark, or a named container.

## Immediate architectural goal

Statically knowable Able operations should remain typed generated Go
operations. A valid native Go string does not need to become a runtime value
or recover goroutine-local package state to compute its byte length.

The prior qualified-struct tranche removed one shared environment-recovery
owner. This tranche selected the next exact three-way owner rather than
optimizing aggregate `currentGID` or `SwapEnvIfNeeded`.

## Selection evidence

The current strict, interpreter-free profile set covered three unlike
concurrent applications:

- Validated Job Pipeline combines Result mapping, callbacks, channels, and
  text validation.
- Concurrent Document Pipeline combines captured typed callables, channels,
  and repeated document scoring.
- Concurrent Event Routing combines parsing, Result mapping, unions, maps,
  captured callables, and repeated routing.

Five merged main-only profiles showed the same exact generated entry:

| Application | Profile total | `String.len_bytes` entry cumulative | Share |
|---|---:|---:|---:|
| Validated Job Pipeline | 1.58 s | 1.07 s | 67.72% |
| Concurrent Document Pipeline | 0.56 s | 0.15 s | 26.79% |
| Concurrent Event Routing | 5.05 s | 2.69 s | 53.27% |

The raw compiled method already contained a native Go fast path, but callers
entered its package wrapper first. In concurrent execution that wrapper
called `SwapEnvIfNeeded`, which recovered the goroutine identity through
`runtime.Stack` before the native length test.

This was an avoidable static crossing. It was distinct from
`Runtime.Env`/`__able_current_payload`, which remains an intentional task
payload boundary.

## Generated-code effect

Before:

```go
length, control :=
    __able_compiled_entry_method_String_len_bytes(value)
```

After, schematically:

```go
length, control := func(value string) (uint64, *__ableControl) {
    if utf8.ValidString(value) && len(value) <= 2147483647 {
        return uint64(len(value)), nil
    }
    return __able_compiled_entry_method_String_len_bytes(value)
}(value)
```

The IIFE parameter guarantees single evaluation even for a computed receiver.
It also keeps the exceptional package-aware path reachable instead of
assuming that every Go string is a valid Able String.

All three generated candidate dependency graphs continue to omit
`able/interpreter-go/pkg/interpreter`.

## Repeated A/B gate

Baseline and candidate binaries were built once and frozen. Five
order-balanced pairs ran on CPU 13 after the core passed the three-sample
quiet-host preflight. Every process used `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and the goroutine executor. All 30 processes
passed the sibling public Ruby verifier.

| Application | Baseline samples (s) | Candidate samples (s) | Mean change | Mean GC |
|---|---|---|---:|---:|
| Validated Job Pipeline | 0.32, 0.32, 0.32, 0.32, 0.33 | 0.12, 0.12, 0.11, 0.11, 0.12 | 0.322 -> 0.116 (-63.98%) | 6.2 -> 4.0 |
| Concurrent Document Pipeline | 0.11, 0.12, 0.13, 0.12, 0.11 | 0.08, 0.08, 0.09, 0.09, 0.08 | 0.118 -> 0.084 (-28.81%) | 3.0 -> 2.6 |
| Concurrent Event Routing | 0.98, 0.98, 0.99, 0.98, 0.96 | 0.46, 0.44, 0.46, 0.46, 0.44 | 0.978 -> 0.452 (-53.78%) | 25.0 -> 19.6 |

The three-application geometric-mean improvement is 50.88%.

## Profile confirmation

Five verified candidate main-only profiles per application were merged:

| Application | Candidate total | `String.len_bytes` entry | Remaining `currentGID` | Largest exact residuals |
|---|---:|---:|---:|---|
| Validated Job Pipeline | 0.59 s | 0 sampled | 0.54 s (91.53%) | task payload, Result callback entry, `String.contains` |
| Concurrent Document Pipeline | 0.41 s | 0 sampled | 0.37 s (90.24%) | typed-callable runtime adapter, task payload |
| Concurrent Event Routing | 2.27 s | 0 sampled | 1.79 s (78.85%) | `String.split`, typed-callable runtime adapter, task payload |

The selected owner disappeared rather than moving below another
`String.len_bytes` symbol. The new exact residuals differ, so this tranche
does not authorize a broad `currentGID` or execution-context ABI change.

## Equivalent Go comparison

Equivalent Go binaries were frozen and measured in five high-resolution
processes on the same CPU. All 15 outputs passed the same public verifiers.

| Application | Candidate Able mean | Go samples (s) | Go mean | Able / Go |
|---|---:|---|---:|---:|
| Validated Job Pipeline | 0.116 s | 0.002181391, 0.002297184, 0.002071322, 0.002370786, 0.002087038 | 0.002201544 | 52.69x |
| Concurrent Document Pipeline | 0.084 s | 0.002150435, 0.002243019, 0.002159591, 0.002066072, 0.002227681 | 0.002169360 | 38.72x |
| Concurrent Event Routing | 0.452 s | 0.002793330, 0.003202969, 0.003180621, 0.003150294, 0.002982165 | 0.003061876 | 147.62x |

The candidate clears the breadth and repeated-improvement gates but remains
far from the compiled 1.052632x target.

## Artifact identity

| Application | Baseline SHA-256 | Candidate SHA-256 | Go SHA-256 |
|---|---|---|---|
| Validated Job Pipeline | `1f49ce4ab8d64041684601790f52663fd2cafa9571db4e6a562ce53519545e47` | `0f3c596dbcda2f517ce677945d773be0e16f0c1ed606b224fadfdc70a6bc7442` | `e5b9725db1379d2bc2ab6918a86c4730c45ef49680a5fd6cb3f5092e5193a9d6` |
| Concurrent Document Pipeline | `84d48681c6a247c9e73fef69227cb82508562041498bd4e7e7f367b9f4bde6be` | `23c5ac49b7655156ffb692be90be6f46a89937c09f86bc9b77dd71d3b7b66364` | `a46a7558e9fbc14a7204a1e37c35f418aae900134d62433c029414772954aa34` |
| Concurrent Event Routing | `d5c0c7640a835786f60fa478e903481b2d5b1245cb1ad0de7a38dfbcaef162cc` | `9aef17771985fa0fb01732fb00f97e67090de6a060e4f977b65b3b4a7bde1343` | `3195e2c5bcf93904e9bde33d2641a0683c516e3faa12799191014b8f92ba277e` |

The aggregate machine-readable record is
`2026-07-24-compiled-string-len-bytes-call-fast-path-retained.json`.
Raw evidence is retained under
`/tmp/able-aot-static-entry-20260724.NcLrdz`.

## Verification

Passing bounded guards cover:

- canonical primitive recognition and native call generation;
- single evaluation of computed String receivers;
- valid native length and invalid unchecked UTF-8 fallback behavior;
- existing literal `String.contains` invalid-input behavior;
- imported environment-independent and package-dependent calls;
- typed callables and nested-spawn imported calls;
- default execution-context ABI selection;
- strict generated interpreter-root exclusion;
- nested spawn, mutex, await/future, and future-flush concurrency parity;
- `go test ./cmd/ablec`.

Generated-source inspection confirms the native call form in all three
applications. Every candidate output and profile process passed its public
verifier.

No canonical stdlib, runtime package, tree-walker, bytecode VM,
language/specification, dependency, or WASM change was required.

## Next

Refresh the boxed typed-callable adapter boundary across Concurrent Document
Pipeline, Concurrent Event Routing, and Concurrent Policy Callbacks. Attribute
the exact `__able_call_value` descendants and prove which callable parameters,
captures, generic specializations, or interface dispatches already have a
fully known Go signature.

This is next because the current profiles leave `__able_call_value` at 63.41%
cumulative in Concurrent Document and 37.89% in Event Routing after the
String-length owner is removed. The work entails preserving typed Go callable
carriers through the relevant parameter/adapter path, retaining
`runtime.Value` only for genuinely dynamic callables, adding generated-source
and dynamic-fallback guards, and repeating the five-pair A/B gate against Go
across three unlike applications. It is important because callable boxing is
now the clearest measured violation of the goal that statically knowable Able
should lower to equivalent native Go without crossing the compiled/runtime
boundary.
