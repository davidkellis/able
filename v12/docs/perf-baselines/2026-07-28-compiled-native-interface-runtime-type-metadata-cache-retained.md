# Compiled native-interface runtime type metadata cache retained

Date: 2026-07-28

## Decision

Retain one immutable generated runtime type expression per native-interface
specialization and reuse it at the final semantic `bridge.MatchType` boundary.

Previously, every recovery call rebuilt expressions such as
`Awaitable<i64>` with `ast.Gen(ast.Ty(...))`. The expression is metadata, not
application state. Moving its construction to package initialization removes
repeated AST allocation without changing matching, coercion, dynamic fallback,
or the native Go interface carrier.

The rule applies uniformly to generated native interfaces. It does not name an
application, container, user nominal type, or non-primitive representation.

## Refreshed owner selection

Five exact diagnostic runs per application after receiver-free kernel
callables found:

| Application | Context calls | Native function | Native bound | Method lookup | Residual `currentGID` |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 888.0 | 104.0 | 784.0 | 888.0 | 1,289.2 |
| Await Channel Mux | 9,216.0 | 9,216.0 | 0.0 | 8,704.0 | 8,707.0 |
| Mutex Await Journal | 8,506.4 | 8,506.4 | 0.0 | 6,458.4 | 128.8 |
| Mutex Work Queue | 16,807.0 | 16,807.0 | 0.0 | 12,711.0 | 168.8 |

Ten main-only CPU profiles per application left residual environment lookup
material only in Future Await Race and Await Channel Mux. Native-bound work is
absent from the two mutex programs. Neither is a shared three-program owner.

Three exact main-allocation profiles per application instead repeated
`ast.NewIdentifier`, `ast.NewSimpleTypeExpression`, and
`ast.NewGenericTypeExpression` below native-interface recovery in Await
Channel Mux and both mutex applications. In one Await Channel Mux profile,
`__able_iface_Awaitable_i64_try_from_value` attributed 7,680 flat objects to
the inline type-expression construction.

## Allocation A/B

Five rotating, public-verifier-backed main-phase allocation processes per side
measured:

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 908,548.8 | 907,185.6 | -0.15% | 14,830.4 | 14,796.8 | -0.23% |
| Await Channel Mux | 10,025,206.4 | 8,516,681.6 | -15.05% | 168,964.2 | 144,398.0 | -14.54% |
| Mutex Await Journal | 5,159,265.6 | 4,405,400.0 | -14.61% | 85,778.0 | 73,484.2 | -14.33% |
| Mutex Work Queue | 10,511,736.0 | 9,012,971.2 | -14.26% | 175,663.8 | 151,272.6 | -13.89% |

The three hot applications improve strongly in both measures. Future Await
Race is the expected low-reach control and remains neutral.

## Balanced timing and Go comparison

Fifteen rotating baseline/candidate/Go cohorts per application measured:

| Application | Baseline | Candidate | Raw change | Paired 95% interval | Go | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 0.014734s | 0.014977s | +1.65% | -3.69% to +8.26% | 0.001577s | 9.50x |
| Await Channel Mux | 0.090586s | 0.088377s | -2.44% | -6.06% to +1.68% | 0.002523s | 35.03x |
| Mutex Await Journal | 0.009721s | 0.009454s | -2.75% | -8.14% to +3.36% | 0.001620s | 5.84x |
| Mutex Work Queue | 0.020230s | 0.018998s | -6.09% | -12.71% to +1.97% | 0.002054s | 9.25x |

Every interval includes zero, so wall time is recorded as neutral. The
compiled applications remain well below the 95%-of-Go product goal.

## Correctness and scope

- A generated-source regression guard requires a package-level cached type
  expression and forbids reconstruction at the final interface match.
- Four strict experimental applications build and verify, and every dependency
  graph omits `pkg/interpreter`.
- All 224 retained census, allocation, and balanced timing executions passed
  their public verifier.
- `go test ./cmd/ablec ./pkg/compiler/bridge` and focused interpreter
  Await/Mutex/Future/Channel tests pass.
- The full `go test ./pkg/compiler` invocation reached its ten-minute
  aggregate package timeout while `TestCompilerExecFixtureFallbacks` was still
  compiling. That exact test passed alone in 167.402 seconds. No test produced
  a semantic failure.

No canonical stdlib, interpreter, bytecode VM, language, dependency,
application source, named-container/non-primitive nominal, or WASM change was
made.

Machine-readable aggregate:

- `2026-07-28-compiled-native-interface-runtime-type-metadata-cache-retained.json`

## Next

Measure generated Awaitable protocol-carrier materialization after this cache.

Why: `channel_awaitable.toStruct`, `mutex_awaitable.toStruct`, and Await waker
construction are now the largest allocation families recurring across Await
Channel Mux and both mutex applications.

What it entails: count native-interface-to-runtime conversions and protocol
carrier construction, trace their consumers through static Array and await
selection, and prototype only a general language-kernel boundary rule that
keeps the typed Go interface carrier until a genuinely dynamic consumer
requires `runtime.StructInstanceValue`. Use five-or-more balanced
baseline/candidate/Go cohorts and preserve dynamic Awaitable behavior.

Why it is important: this is the next measured opportunity to remove boxing
and compiled/runtime crossings. It directly advances the goal of making
fallback-free compiled Able behave like native Go without adding a
benchmark-specific or nominal-type shortcut.
