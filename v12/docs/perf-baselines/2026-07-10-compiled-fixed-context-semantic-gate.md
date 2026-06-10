# Fixed-Context ABI Semantic Gate

## Decision

Keep the fixed-pointer execution-context ABI opt-in.  The option now passes
the compiler semantic gate, including the remaining dynamic bridge forms, but
the compiler default remains unchanged until the broader compiler test suite
is run with the option enabled.

The shared fixture/audit follow-up is complete; see
`2026-07-10-compiled-fixed-context-expanded-fixture-gate.md` for its 114
fixture runs, marker audits, race result, and the generic bound-method wrapper
correction it exposed.

## Method

`runCompiledFixtureOutcomeWithOptions` extends the existing fixture outcome
runner without changing its default behavior.  The experimental gate compiles
each selected fixture with:

```text
Options{PackageName: "main", ExperimentalExecutionContext: true}
```

and compares the compiled binary's exit status, stdout, and stderr to the
tree-walker result.  The fixture batch is deliberately sequential: generated
Go builds must not overlap under the project memory guardrails.

The dynamic-boundary batch also enables the existing marker interface.  It
requires no fallback calls and at least one explicit dynamic boundary call.
Two temporary programs directly assert the two compatibility wrappers left by
the fixed ABI:

- a callable value (`call_value`) captures a lexical variable and is invoked
  by dynamically defined Able code;
- an unresolved named call (`call_named:missing_runtime_fn`) retains the
  error/exit behavior through the named dynamic bridge.

## Coverage and result

The following reference-parity fixtures passed under the option:

- map literals; generic inference; interface defaults/generics; rescue/ensure;
  package prelude; and inline Go externs;
- temporary-file I/O;
- spawn/await, fairness/cancellation, channel ping-pong, nested spawn through
  a native channel boundary, and background-work flush;
- package-object metaprogramming, selective dynimport, dynimport interface
  dispatch, and search-path environment overrides.

The two direct dynamic probes also passed, with zero fallback marker calls and
the expected explicit marker in each case.  The full option-aware run was:

```text
go test ./pkg/compiler \
  -run 'TestCompilerExperimentalExecutionContext(DynamicNamedAndValueBoundaries|FixtureParity|DynamicBoundaryParity)' \
  -count=1 -timeout 60s
```

Result: `ok able/interpreter-go/pkg/compiler 38.876s`.

This is semantic evidence only.  It adds no generated-runtime, benchmark, or
`able-stdlib` behavior, and it introduces no benchmark-, task-count-, or
nominal-container-specific branch.

## Next recommendation

Run the complete existing compiler suite with the fixed ABI selected wherever
it emits and executes a generated binary, then repeat the targeted generated
race build.  The category matrix proves the compatibility boundaries directly;
the remaining rollout condition is broad regression coverage across every
compiler harness rather than more performance tuning.  Parameterize or wrap
the shared test construction at that boundary, compare failures against the
default path, and keep the flag opt-in unless the option-aware suite is green.
