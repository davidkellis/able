# Compiled qualified-struct registry fast path retained

Date: 2026-07-24

## Decision

Retain the general generated-package and compiler-bridge change that lets
qualified static struct lookups use the generated registry before attempting
goroutine-local package-environment recovery.

The rule has two parts:

- generated package registrars preserve the loader-qualified package identity
  and also register the normalized identity used by generated
  struct-to-runtime converters when the two differ;
- `bridge.Runtime.StructDefinition` checks the exact qualified registry before
  recovering the current environment, while retaining the existing
  environment, cache, interpreter, and global-fallback paths for dynamic or
  unregistered lookups.

This is a package-identity and static-registry rule. It does not distinguish
specific structs, applications, containers, or non-primitive nominal types,
and it does not change the execution-context ABI.

## Immediate architectural goal

Compiled Able should lower statically knowable Able execution into typed
generated Go and cross package/runtime/interpreter boundaries only when
semantics require them. A generated converter for an exactly named,
already-registered struct definition is a static boundary: it does not need
goroutine-local interpreter state merely to recover that definition.

The prior callable-entry tranche reduced package-entry recovery, but
caller-level profiles still showed a second shared owner below
`bridge.Runtime.StructDefinition`. This tranche closes that exact owner.

## Residual attribution

The refresh used three unlike strict, interpreter-free applications:

- Dependency Wave Validation;
- Validated Job Pipeline;
- Concurrent Document Pipeline.

Five merged baseline profiles and generated-source attribution showed
qualified struct lookup in every application:

| Application | Profile total | `StructDefinition` cumulative | Share | Generated callers |
|---|---:|---:|---:|---|
| Dependency Wave Validation | 1.67 s | 0.93 s | 55.69% | `WaveResult`, `Accepted`, `WaveTask`, `Recovered` converters |
| Validated Job Pipeline | 1.72 s | 0.15 s | 8.72% | `JobResult`, `JobTask`, `ValidationError` converters |
| Concurrent Document Pipeline | 0.64 s | 0.16 s | 25.00% | `DocumentScore`, `DocumentTask` converters |

In each case the generated converter supplied a fully qualified static name,
yet `StructDefinition` called `currentEnv` before consulting
`qualifiedStructs`. In concurrent applications that meant parsing
`runtime.Stack` through `bridge.currentGID`.

The same profiles also separated a noncandidate:
`Runtime.Env`/`__able_current_payload` is a genuine task/channel payload
runtime-service boundary. It remains outside this change.

## Rejected diagnostic precursor

The first candidate moved the qualified-registry check ahead of environment
recovery but did not change registration. It was neutral or slower and
profiles showed that `StructDefinition` still reached `currentGID`.

Generated-source inspection found a general identity mismatch. For example:

- package registration used
  `dependency_wave_validation.dependency_wave_validation.WaveTask`;
- generated conversion requested
  `dependency_wave_validation.WaveTask`.

The diagnostic candidate therefore missed the registry. It was not retained.
Its binaries, timings, and profiles remain under
`rejected-qualified-key-miss*` in the temporary evidence root.

## Retained implementation

Generated struct registrars now register both package identities when
normalization changes the name:

1. the original loader-qualified identity remains registered;
2. the normalized runtime lookup identity is additionally registered;
3. already-present definitions and newly constructed definitions follow the
   same rule.

`Runtime.StructDefinition` performs a locked read of the exact qualified
registry before calling `currentEnv`. The later qualified/cache check remains
in place so concurrent registration and all dynamic fallback behavior retain
their previous semantics.

A regression test constructs a bridge with no interpreter or environment,
registers one qualified definition, verifies the lookup, and requires no more
than 0.1 allocation per cached run. Compiler tests cover repeated-leaf
normalization and preservation of both package identities.

## Repeated A/B gate

Baseline and candidate binaries were built once, frozen, and measured in five
order-balanced pairs on quiet CPU 7 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, and `GOGC=50`. The benchmark executor used a goroutine so
the measured path retained its concurrency behavior. Every Able process
passed the sibling public verifier.

| Application | Baseline samples (s) | Candidate samples (s) | Mean change | Mean GC | Verified |
|---|---|---|---:|---:|---:|
| Dependency Wave Validation | 0.43, 0.37, 0.34, 0.35, 0.35 | 0.18, 0.18, 0.16, 0.16, 0.17 | 0.368 -> 0.170 (-53.80%) | 10.8 -> 10.0 | 5/5 both |
| Validated Job Pipeline | 0.37, 0.37, 0.38, 0.34, 0.34 | 0.32, 0.32, 0.34, 0.31, 0.31 | 0.360 -> 0.320 (-11.11%) | 7.0 -> 6.8 | 5/5 both |
| Concurrent Document Pipeline | 0.15, 0.14, 0.14, 0.14, 0.14 | 0.12, 0.12, 0.12, 0.12, 0.12 | 0.142 -> 0.120 (-15.49%) | 3.0 -> 3.0 | 5/5 both |

The breadth gate passes: the rule improves dependency-wave record conversion,
validated Result/channel processing, and a concurrent document-scoring
pipeline.

## Profile confirmation

Five verifier-backed candidate profiles per application were merged from the
exact timed candidate binaries:

| Application | Candidate total | Qualified `StructDefinition` | Remaining `currentGID` | Exact remaining source |
|---|---:|---:|---:|---|
| Dependency Wave Validation | 0.74 s | 0 sampled | 0.59 s (79.73%) | `Runtime.Env` task payload |
| Validated Job Pipeline | 1.58 s | 0 sampled | 1.51 s (95.57%) | `Runtime.SwapEnv`, `SwapEnvIfNeeded`, `Runtime.Env` |
| Concurrent Document Pipeline | 0.56 s | 0 sampled | 0.50 s (89.29%) | `Runtime.SwapEnv`, `SwapEnvIfNeeded`, `Runtime.Env` |

The selected exact owner disappears in all three. The remaining ledgers are
not identical: Dependency Wave is now dominated by the genuine runtime task
payload boundary, while Validated Job and Concurrent Document retain
substantial static package-entry guards.

## Equivalent Go comparison

Equivalent Go binaries were built once and measured in five high-resolution
processes on the same pinned CPU. All 15 outputs passed the same public
verifiers.

| Application | Candidate Able mean | Go samples (s) | Go mean | Able / Go |
|---|---:|---|---:|---:|
| Dependency Wave Validation | 0.170 s | 0.002237721, 0.002107611, 0.002173661, 0.002143060, 0.002343064 | 0.002201023 | 77.24x |
| Validated Job Pipeline | 0.320 s | 0.001884003, 0.001842097, 0.001943577, 0.001931657, 0.001956458 | 0.001911558 | 167.40x |
| Concurrent Document Pipeline | 0.120 s | 0.001978990, 0.001618984, 0.001905996, 0.001903402, 0.001864624 | 0.001854399 | 64.71x |

The candidate clears the breadth and repeated-improvement gates but remains
far from the compiled 1.052632x target. Continued static boundary closure is
therefore required.

## Artifact identity

| Application | Baseline SHA-256 | Candidate SHA-256 | Go SHA-256 |
|---|---|---|---|
| Dependency Wave Validation | `eb3622753d2413896e410bbd51c202d70c3ca1cf5496a0c57927fa24e5a76e59` | `f37d46bbe302e3b24151d73411c47699f1c1f5c2928d2536246038056632761c` | `cda38aee36fd8dcb17cbd236a21b20403cb761687a6638c376a54c5e5ad850f6` |
| Validated Job Pipeline | `0487e3173491c9fa28ec43a6731c98545dccf6c6b0ac5ec94514ef7c4ea01cb4` | `1f49ce4ab8d64041684601790f52663fd2cafa9571db4e6a562ce53519545e47` | `e5b9725db1379d2bc2ab6918a86c4730c45ef49680a5fd6cb3f5092e5193a9d6` |
| Concurrent Document Pipeline | `c7b28b7f257cb1b03d657cbaaf3d387a7e944c5c2b7e87a48acc0e36dacaa9bd` | `84d48681c6a247c9e73fef69227cb82508562041498bd4e7e7f367b9f4bde6be` | `a46a7558e9fbc14a7204a1e37c35f418aae900134d62433c029414772954aa34` |

The aggregate machine-readable record is
`2026-07-24-compiled-qualified-struct-registry-fast-path-retained.json`.
Raw temporary evidence is retained under
`/tmp/able-aot-struct-definition-20260724.yc8lAD`.

## Verification

Passing bounded guards cover:

- runtime package-name normalization and dual qualified registration;
- zero-allocation interpreter-free qualified struct lookup;
- imported canonical Matcher struct conversion;
- environment-independent and dependent imported method behavior;
- static interpreter-root exclusion;
- nested spawn, mutex, await, goroutine-future, and future-flush parity;
- `go test ./cmd/ablec`.

All three generated candidate dependency graphs omit
`able/interpreter-go/pkg/interpreter`. Modified files remain below 1,000
lines, and `git diff --check` passes.

No canonical stdlib, tree-walker, bytecode VM, language/specification,
dependency, or WASM change was required. The implementation is confined to
the compiler generator and its generated bridge support.

## Next

Refresh exact residual `SwapEnvIfNeeded` callers across Validated Job
Pipeline, Concurrent Document Pipeline, and one unlike current strict
String/callback-heavy concurrent application. Select only a statically pure
String method, callable entry, or other exact semantic category that repeats
in all three; do not optimize the genuine
`Runtime.Env`/`__able_current_payload` task-payload boundary.

This is next because the retained lookup rule removes the shared
`StructDefinition` owner, leaving static entry guards as the largest avoidable
category in two applications but not in Dependency Wave. The work entails new
main-only caller profiles, generated-source attribution, fixed-point
environment-effect proof, concurrency parity guards, and five order-balanced
verifier-backed A/B pairs against equivalent Go applications. It is important
because the current programs remain 64.71x-167.40x behind Go, and crossing
into goroutine-local runtime state for a provably package-independent static
entry is now the clearest measured lowering gap.
