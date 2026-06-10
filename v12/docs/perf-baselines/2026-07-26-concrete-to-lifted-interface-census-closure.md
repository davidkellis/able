# Concrete-to-lifted-interface census closure

Date: 2026-07-26

## Decision

Retain the report-only typed-boundary observer extension and no production
compiler, generated-runtime, runtime, interpreter, bytecode VM, stdlib,
benchmark, language, dependency, or WASM performance change.

All 61 current coverage applications compiled with `--no-fallbacks` and the
opt-in observer, ran once, and passed their public verifiers. All 61 final Go
dependency graphs omit `able/interpreter-go/pkg/interpreter`; dependency
counts range from 96 to 119 packages.

The exact `concrete Go carrier -> runtime.Value -> lifted native interface`
route has zero executable callsites and zero runtime events across the strict
corpus. No shape therefore reaches even one application, much less the
required three unlike applications. No performance prototype or repeated A/B
cohort was admissible.

The machine-readable companion is
`2026-07-26-concrete-to-lifted-interface-census-closure.json`.

## Observer refinement

The opt-in typed-boundary observer now has an
`interface_lift_via_runtime` category at the final native-interface fallback.
Candidate metadata includes:

- generated function and Able source;
- fully bound concrete carrier after current type bindings;
- lifted target interface; and
- the semantic reason for runtime boxing and recovery.

The first smoke run incorrectly treated generic definitions such as
`ArrayIterator<T> -> Iterator<T>` as concrete. The observer was corrected to
require a fully bound source carrier before registration. Ordinary
diagnostics-off generated code remains unchanged.

Expression lowering can compare speculative coercion paths before choosing a
native adapter. Candidate metadata is consequently not reachability evidence
by itself. An emitted marker call proves generated reachability; a nonzero
snapshot count proves execution. The focused guard now records this
distinction explicitly.

## Strict application result

The complete application pass produced:

- 61 comparison rows, all successful and publicly verified;
- 61 telemetry snapshots, exactly one per application;
- 2,927 generated top-level Go files scanned;
- zero emitted `interface_lift_via_runtime` marker calls beyond the marker
  definition;
- zero `interface_lift_via_runtime` runtime events; and
- 61 successful `go list -deps` checks with zero interpreter-backed graphs.

Seventeen metadata candidates appeared during speculative lowering:

| Application | Candidate registrations | Concrete families |
| --- | ---: | --- |
| Inventory Reconciliation | 1 | `HashMap<i64, i64> -> Map<i64, i64>` |
| Concurrent Graph Visitors | 4 | `ThresholdVisitor` / `StripeVisitor -> GraphVisitor` |
| Concurrent Packet Codecs | 4 | `DeltaCodec` / `RunCodec -> PacketCodec` |
| Concurrent Tree Folds | 4 | `AdditiveFold` / `MultiplicativeFold -> FoldAlgebra` |
| Concurrent State Machines | 4 | `GateHandler` / `CycleHandler -> StateHandler` |

None reached emitted code. The compiler selected the already-generated direct
native interface adapters instead.

## Representative compiled stdlib result

The strict corpus itself compiled representative canonical stdlib packages:

- Inventory Reconciliation included `able.collections.hash_map` and
  `able.collections.map`;
- Rational Series included `able.numbers.rational`; and
- Unicode Scalar Pipeline included `able.text.string`.

Those applications remained interpreter-free, verified, and free of the
target executable route.

A retained compiled foundational stdlib build additionally covered core
interfaces/options/iteration, collections, `able.spec` assertions, the test
harness, and text/string/regex packages. Its binary exits zero. Five candidate
registrations came from `able-stdlib/tests/assertions.test.able`
(`BeBetweenMatcher<i64>` and `CustomMatcher<i64>` toward `Matcher<i32>`), but
again there were zero emitted marker calls: sibling/direct native matcher
adapters replaced the speculative runtime recovery path.

The surrounding Go integration test reported failure only because
`ABLE_TEST_KEEP_WORKDIR=1` intentionally wrote the retained path to stderr
while the helper expected empty stderr. The instrumented compilation took
138.88 seconds; this is recorded as a possible test-lane issue, not as an Able
semantic or application-performance failure.

## Frozen protocol

- Repository commit:
  `237406eccdfb025a519d898daedadee1c8d13a7b`.
- Full-corpus compiler SHA-256:
  `d03978d57340528de0f941d262581e1f2486f7dd250ecf024eefc8de85d7e3c2`.
- Full-corpus comparison JSON SHA-256:
  `61218a963274f131d56e34c6b56c8c07d46b7da7f6dba3390f8db0f250e4c31b`.
- Foundational compiled-test binary SHA-256:
  `b26bcbdeef548a1cabb6b95d9dca11a2c1c777b5deab24b3352dfeca0111a205`.
- Corpus artifacts used disk-backed
  `/var/tmp/able-interface-lift-census-20260726`.

The diagnostic application runs are reachability evidence, not performance
measurements: the observer adds atomic counters. The one-run wall times are
not used for performance conclusions.

## Why the tranche stops here

There is no executable owner to profile. Repeated measurements cannot
establish a speedup for code that is absent, and implementing a new direct
coercion would duplicate the native adapter path the compiler already chose.
Any matcher-, Result-, stdlib-, container-, nominal-type-, or
benchmark-specific change would also violate the generality rules.

No canonical `able-stdlib` source changed. No production runtime or lowering
changed. WASM remains deferred.

## Verification

- Focused observer guards pass in 0.994 seconds.
- Focused CLI flag/environment guards pass in 0.036 seconds.
- `go vet ./...` and `go build ./...` pass.
- The complete `./run_all_tests.sh` handoff passes every contract,
  non-compiler package, all 33 compiler batches, and the final 87.571-second
  bytecode fixture corpus.
- The three known heavy aggregate compiler batches took 187.586, 84.907, and
  110.522 seconds without an individual timeout.
- The ordinary diagnostics-off `cmd/able` package completed in 52.608 seconds.
- Canonical stdlib tests pass in tree-walker mode in 18 seconds and bytecode
  mode in 15 seconds.
- No canonical `able-stdlib` source changed.

## Next recommendation

Keep production performance mutation paused. First reproduce the compiled
foundational stdlib integration duration without telemetry and without
retained-workdir stderr.

Why: the concrete-interface census did not invalidate the closed performance
frontier. The only new actionable signal is the 138.88-second diagnostic
integration test, while project policy requires individual tests to complete
within one minute.

What it entails: run at least three warm-cache diagnostics-off repetitions,
separate Able compilation, generated-Go build, and execution time, and change
the test lane only if the ordinary median still exceeds one minute. Prefer
compile/build reuse or narrower test shards; do not weaken semantic coverage.

Why it is important: fast, reliable release gates make future compiler
correctness and performance changes cheaper to validate, while the
evidence-first reproduction avoids mistaking observer overhead or retained
artifact plumbing for a production compiler regression. Continue to defer
WASM.
