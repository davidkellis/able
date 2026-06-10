# Compiler correctness gate reconciled

Date: 2026-07-25

## Decision

Retain two general corrections required by the interpreter-free compiled
runtime:

1. the bridge can evaluate primitive binary operators without a concrete
   interpreter when generated static code still has an erased
   `runtime.Value` carrier;
2. generic type identity derives package ownership from the canonical nominal
   base rather than the caller context, so an imported generic alias and its
   canonical target share one specialization key.

The primitive path handles only built-in numeric, string, boolean, character,
and nil semantics. It deliberately declines nominal values and unsupported
dotted operations so user-defined operator dispatch continues to require the
interpreter/dynamic boundary.

No application, benchmark, named-container, or non-primitive nominal rule was
added.

## Failure attribution

The six failures exposed after the forward callable-binding tranche had three
causes:

- `TestCompilerInferredNilWithValueBranchPreservesValue` and
  `TestCompilerPipePlaceholderLambdaExecutes` reached `__able_binary_op` with
  primitive boxed values after the static interpreter-package cut. The only
  remaining bridge path required a concrete interpreter and returned
  `compiler bridge: missing interpreter`.
- `TestCompilerSpecializedImplCanonicalKeyPreventsDuplicateContainAllBodies`
  generated identical `ContainAllMatcher<String>.matches` bodies under
  `able.kernel.Array<String>` and the `able.collections.array.Array<String>`
  alias. The generic identity key redundantly prefixed the caller package even
  though the nominal base already carried its defining package.
- The execution-context, Iterator filter/map, and no-self interface tests
  expected entry-wrapper calls that the retained environment-independence
  proof had correctly localized to raw generated Go bodies. Their assertions
  were updated to require the direct body calls while preserving the guarded
  entry definitions for dependent callers.

## Suite-timeout attribution

The optional single-invocation compiler package contains 785 top-level tests
plus large parallel fixture matrices. A 30-minute run capped with
`-parallel=2` reached the package timeout while
`13_05_dynimport_interface_dispatch` and
`14_02_regex_core_match_streaming` had been active for only 1 and 29 seconds.
The timeout stack also contained idle `SerialExecutor` loops from completed
tests, but none owned the active wait.

The two named fixtures pass alone in 2.536 seconds and 28.011 seconds. The
ordinary compiler suite, excluding only the eight large fixture/parity
matrices, passes in 566.264 seconds with `-short -parallel=2`. The timeout is
therefore aggregate package scheduling, not a per-test semantic deadlock.
Future handoffs should keep compiler logic and fixture matrices as bounded
separate gates; no individual test in this reconciliation exceeded one minute.

## Verification

Passed:

- all six reconciled compiler regressions together: 26.774 seconds;
- complete `pkg/compiler/bridge`: 0.048 seconds;
- ordinary compiler tests excluding the large fixture/parity matrices:
  566.264 seconds;
- `go test ./cmd/ablec`: 5.404 seconds;
- three representative no-bootstrap boundary audits: 3.518 seconds;
- the same three interpreter-free fixture executions: 4.002 seconds;
- the two fixture subtests named by the aggregate timeout: 2.536 seconds and
  28.011 seconds;
- file-length audit: the new primitive bridge implementation is 407 lines;
  existing near-limit generator files remain at 997, 994, and 998 lines.

No canonical stdlib, tree-walker, bytecode VM, language, dependency, or WASM
change was required. This correctness tranche makes no performance claim.

## Next

Refresh CPU and allocation profiles for Concurrent Document Pipeline,
Concurrent Event Routing, and Concurrent Policy Callbacks after the retained
lowering changes. Confirm whether the exact
`__able_current_payload -> bridge.Runtime.Env -> bridge.currentGID` path still
repeats materially in all three.

Advance a candidate only if it is a general way to preserve or reuse scheduler
payload at generated dynamic callable/await boundaries, without reopening the
broad execution-context ABI. Gate it with verifier-backed repeated A/B
measurements against the equivalent Go applications. This matters because
these applications remain far from Go after ordinary static calls and
callables have been lowered directly, making scheduler-context recovery the
next plausible shared compiled boundary.
