# Compiled forward callable binding retained

Date: 2026-07-25

## Decision

Retain the general compiler rule that contextually types an unannotated local
lambda from its later direct use as an argument to a statically resolved,
fully bound callable parameter.

The rule is deliberately conservative:

- every observed use must be a direct identifier argument with a statically
  known callable parameter type;
- all inferred callable constraints must be identical and compatible with the
  lambda's arity and any explicit parameter or return annotations;
- indirect calls, conflicting signatures, reassignment, escape, dynamic
  parameters, and irreducibly polymorphic constraints keep the existing
  erased `runtime.Value` carrier and adapters.

This is a contextual typing correction in the shared assignment/callable
pipeline. It adds no application, benchmark, container, or non-primitive
nominal rule.

## Attribution

The census retained Concurrent Document Pipeline and Concurrent Event Routing
and selected Policy Record Dispatch as the unlike third guard. All three had
the exact same generated round-trip:

1. an unannotated captured scorer was first emitted with `runtime.Value`
   parameters;
2. the local callable was converted to `runtime.Value`;
3. the value was immediately converted back to the concrete callable required
   by a statically resolved function argument;
4. the hot callback therefore entered `__able_call_value`.

Concurrent Policy Callbacks was not used because its callbacks were already
fully typed. Manifest Normalization contained callable boxing, but not the
same immediate erased-to-concrete round-trip.

## Generated-code effect

The three scorer bindings now use these native carriers from their declaration
through their typed call:

- Document: `__able_fn_string_int64_to__DocumentScore`;
- Event: `__able_fn__EventRecord_int64_to__AcceptedRoute`;
- Policy Record: `__able_fn__PolicyRecord_int64_to__AcceptedDecision`.

Their enclosing generated functions contain no scorer
`to_runtime_value`/`from_runtime_value` conversion and no scorer
`__able_call_value`. Explicitly dynamic and conflicting callable uses retain
the erased carrier and both runtime adapters.

All three strict generated dependency graphs remain interpreter-free.

## Repeated A/B gate

Frozen baseline and candidate binaries ran in ten order-balanced pairs on CPU
9 with `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`ABLE_EXECUTOR=goroutine`. Elapsed time used the shell's nanosecond-resolution
clock. Every one of the 60 Able processes passed its sibling public verifier.

| Application | Baseline samples (s) | Candidate samples (s) | Mean change |
|---|---|---|---:|
| Concurrent Document Pipeline | 0.085630, 0.062941, 0.060314, 0.061043, 0.054996, 0.055778, 0.054078, 0.054517, 0.056398, 0.061073 | 0.034676, 0.037063, 0.037353, 0.034824, 0.036088, 0.035570, 0.034455, 0.034632, 0.034844, 0.045513 | 0.060677 -> 0.036502 (-39.84%) |
| Concurrent Event Routing | 0.375935, 0.379452, 0.392133, 0.377325, 0.374788, 0.369608, 0.368845, 0.387344, 0.373446, 0.381660 | 0.303913, 0.318852, 0.312531, 0.307910, 0.296356, 0.295731, 0.293730, 0.295551, 0.295963, 0.293763 | 0.378054 -> 0.301430 (-20.27%) |
| Policy Record Dispatch | 0.117242, 0.108712, 0.113386, 0.105695, 0.105587, 0.102197, 0.102523, 0.118567, 0.101754, 0.106100 | 0.102024, 0.102319, 0.109241, 0.099142, 0.099982, 0.096566, 0.096914, 0.111256, 0.095371, 0.096774 | 0.108176 -> 0.100959 (-6.67%) |

The geometric-mean improvement is 23.50%. A separate five-pair GNU `time`
sample found mean peak RSS changes of 13,580.8 to 12,708.8 KB for Document,
28,580.8 to 28,588.8 KB for Event, and 20,898.4 to 20,482.4 KB for Policy
Record.

## Profile confirmation

Five candidate main-phase profiles were merged per application. The selected
scorer `__able_call_value` and concrete scorer `from_runtime_value` owners are
absent from every candidate profile:

| Application | Baseline selected owner | Candidate total samples | Candidate selected owner |
|---|---:|---:|---:|
| Concurrent Document Pipeline | `__able_call_value` 0.17 s (45.95%) | 0.14 s | no sample |
| Concurrent Event Routing | `__able_call_value` 0.40 s (17.86%) | 1.49 s | no sample |
| Policy Record Dispatch | concrete wrapper 0.04 s (7.27%); `__able_call_value` 0.01 s (1.82%) | 0.48 s | no sample |

The remaining Document and Event profiles are dominated by generated
scheduler payload recovery through `bridge.currentGID`. Policy Record is
instead dominated by regex/GC work, confirming that this tranche removed the
shared callable owner without claiming those unrelated residuals.

## Equivalent Go comparison

Each source-equivalent Go program was built with `-trimpath` and ran ten
verified processes under the same affinity and runtime settings.

| Application | Candidate Able mean | Go mean | Able / Go |
|---|---:|---:|---:|
| Concurrent Document Pipeline | 0.036502 s | 0.002585 s | 14.12x |
| Concurrent Event Routing | 0.301430 s | 0.003900 s | 77.29x |
| Policy Record Dispatch | 0.100959 s | 0.003753 s | 26.90x |

The candidate clears the breadth gate but does not meet the compiled
95%-of-Go goal.

## Artifact identity

| Application | Baseline SHA-256 | Measured candidate SHA-256 | Final rebuild SHA-256 | Go SHA-256 |
|---|---|---|---|---|
| Concurrent Document Pipeline | `6980f136817e02d0cafe437737322cdddd5cc40ddf088a5678538d8e3f36ee37` | `f3b174743471778973e0613b05a3bbb1470bbb8ab4687c670cc69d1ae0d78786` | `ba86258c41c92a71314c5ed238fd786b095928fbb379b8fc3df3d4f35f2f476e` | `a46a7558e9fbc14a7204a1e37c35f418aae900134d62433c029414772954aa34` |
| Concurrent Event Routing | `32573925aa6a2fb6151dcd75b20b547e1b2945dea7809cb1c57fa1ddffb0c4a3` | `d22fdcc528ea9efce9d7c4d6c854b74a1c0b3562d07497feddd8504924f609ca` | `d54436662b93f7d948724dc9ce5ca6c69cad0a2f91d867016b5bacf201553b81` | `e08c89e8f55ee0e09834e4c4498248dd61cfb5683c542b044151271d37ad16d7` |
| Policy Record Dispatch | `2e92aa3744beebf17bab1b006cb8c57cdee701947cb6c97c67528ebade271ad6` | `e2297a042b9d3537dbcef84301f3ee1740ba876c721709be31245638e0ea4876` | `453bfe3a1d89269c881357c67d333c68838a8b0a61ed8b61bae9915e560cd19e` | `b011eae0a14d2687320bc7baf51521d7af985b4c0eb876ad14ca1e2e7f6e2f0a` |

The final refactor moved inference mechanics out of
`generator_assignments.go`, restoring that file to 997 lines. Root generated
application sources from the measured and final builds are byte-for-byte
identical; only build identity changed.

The machine-readable record is
`2026-07-25-compiled-forward-callable-binding-retained.json`. Temporary raw
evidence is under `/tmp/able-aot-native-callable-flow-20260725.J4pdLM`.

## Verification

Passing bounded guards cover:

- native contextual callable generated source through a nested use;
- explicitly dynamic and conflicting callable constraints;
- generic Enumerable/Iterator and interface callable boundaries;
- imported/shadowed callable joins and callable return adapters;
- native callable struct fields and execution;
- spawn capture, channels, mutexes, goroutine futures, and array tasks;
- strict interpreter-free builds and public verification for all three
  applications;
- `go test ./cmd/ablec`.

An additional full `go test ./pkg/compiler -count=1 -timeout=20m` sweep is not
green in the current worktree. It reported six deterministic failures before
timing out:

- `TestCompilerInferredNilWithValueBranchPreservesValue`;
- `TestCompilerExperimentalExecutionContextGuardsCrossPackageInterfaceBodies`;
- `TestCompilerConcreteIteratorFilterMapStayNativeWithExperimentalMonoArrays`;
- `TestCompilerPipePlaceholderLambdaExecutes`;
- `TestCompilerSpecializedImplCanonicalKeyPreventsDuplicateContainAllBodies`;
- `TestCompilerNoSelfInterfaceConcreteAdapterKeepsDeclaredArity`.

The six-test bounded rerun reproduced them in 27.284 seconds. They cover nil
joins, experimental context guards, an inline Iterator lambda, a pipe
placeholder, specialization deduplication, and no-self interface arity; none
enters the new direct local-lambda assignment inference hook. The full sweep
timed out while
`TestCompilerGenericStaticNominalMethodInfersInterfaceParamBindingExecutes`
was waiting on a child process, but that test passes alone in 0.734 seconds.
This indicates current suite interaction or leaked executor/process state and
must be reconciled before more performance work.

No canonical stdlib, runtime, interpreter, tree-walker, bytecode, language,
dependency, or WASM change was required.

## Next

Reconcile the six current full-compiler failures and the suite-level
executor/process interaction timeout first. Preserve this tranche's generated
shape and repeat its bounded/application gates during that work.

This is next because the semantic gate must be green before another
performance candidate is admissible. Once it is green, refresh strict profiles
for Concurrent Document Pipeline, Concurrent Event Routing, and Concurrent
Policy Callbacks and attribute the repeated
`__able_current_payload -> bridge.Runtime.Env -> bridge.currentGID` scheduler
path. That remains the next performance target because it accounts for 92.86%
cumulative CPU in Document, 69.13% in Event, and 95.12% in the prior Policy
Callbacks profile. Advance only a narrow general static-payload rule; do not
reopen the rejected broad execution-context ABI route. Do not begin WASM work.
