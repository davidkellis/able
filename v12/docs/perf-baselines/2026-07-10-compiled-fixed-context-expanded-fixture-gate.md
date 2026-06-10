# Fixed-Context ABI Expanded Fixture Gate

## Decision

Keep the fixed-pointer execution-context ABI opt-in.  The complete shared
fixture and marker-audit gate now passes under the option, including a generic
dynamic bound-method defect found and fixed during this run.  This closes the
fixture/audit rollout condition; it does not silently change the compiler
default.

## Test-only option seam

`ABLE_COMPILER_FIXTURE_EXPERIMENTAL_EXECUTION_CONTEXT=1` is a test-only switch
that adds `ExperimentalExecutionContext` at shared generated-binary harnesses.
With the variable absent, each harness preserves its previous options and the
production compiler remains controlled only by `Options` and
`ablec -experimental-execution-context`.

The switch covers the ordinary and no-bootstrap fixture runners, fixture
outcome parity, dynamic-boundary parity, static boundary audits, interface and
global-lookup audits, and strict-dispatch fixture runs.  It is not a production
environment override.

## Coverage and result

All runs used `GOMEMLIMIT=1GiB`, `GOGC=50`, `-parallel=1`, and a 60-second test
timeout.  The 114 curated compiler execution fixtures were distributed across
12 serial batches, so generated Go builds did not overlap.  Every batch passed
with the candidate enabled.

The candidate also passed:

- all 42 existing dynamic-boundary tests (callable values, named calls,
  composite/interface/union/nullable conversion, mono arrays, and ordinary and
  native bound methods);
- all eight static fallback-marker batches, all four interface/global-lookup
  batches, and both strict-dispatch heavy-fixture batches;
- the generated nested-spawn program under `go build -race`, plus the bridge
  race suite.

The targeted post-gate checks were:

```text
go test ./pkg/compiler -run '^TestCompilerExperimentalExecutionContextNestedSpawnExecutes$' -count=1 -timeout 60s
go test -race ./pkg/compiler/bridge -count=1 -timeout 60s
go test ./pkg/interpreter -run '^$' -count=1 -timeout 60s
```

All passed.

## Generic defect found and fixed

The ordinary dynamic bound-method tests initially failed to build: a native
callable closure selected `Counter.add_ctx`, although that closure is invoked
through a function-value/dynamic bridge and has no lexical execution-context
parameter.  This was not a `Counter` rule.  Any compiled source method exposed
as a native callable value had the same invalid assumption.

`compileNativeBoundMethodValue` now calls the existing stable compatibility
entry for every such method value.  That wrapper derives its context at the
native bridge boundary, which is the established ABI contract for dynamic
calls.  A source guard proves candidate bound-method values use the
compatibility entry, and both ordinary and native bound-method callback tests
pass afterward.  No benchmark-, task-count-, or nominal-container-specific
branch was added.

No `able-stdlib` source changed.

## Next recommendation

Run the remaining direct generated-binary helper suite with the same test-only
option: the shared `compileAndRunSource`, `compileAndRunSourceWithOptions`, and
`compileAndRunExecSourceWithOptions` families cover bespoke native-interface,
extern, generic, union, nullable, and collection execution tests outside the
fixture harness.  This is the remaining semantic rollout evidence before a
default-flip decision.  The work entails option-normalizing those three test
helpers, splitting their existing tests into bounded groups, and retaining
opt-in status unless every candidate result is green.
