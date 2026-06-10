# Fixed-context direct generated-binary gate — 2026-07-11

## Scope

The opt-in compiler execution-context ABI already passed the fixture, dynamic
boundary, audit, and race gates. This tranche extended the same test-only
option switch to the three existing direct generated-binary helpers:

- `compileAndRunSource`
- `compileAndRunSourceWithOptions`
- `compileAndRunExecSourceWithOptions`

The default helper paths and production compiler default remain unchanged.

## Candidate result

All 103 existing direct executable compiler tests passed with
`ABLE_COMPILER_FIXTURE_EXPERIMENTAL_EXECUTION_CONTEXT=1`, `GOMEMLIMIT=1GiB`,
`GOGC=50`, `-parallel=1`, and a per-command `-timeout 60s`. The selection is
derived from direct call sites of the three helpers, not from a bespoke
benchmark list. It covers benchmark-shaped iterator and collection programs,
generic calls, interfaces, unions, nullable values, callable values,
recursion, ordered and persistent collections, range, filesystem I/O, string
implementations, struct patterns, and truthiness.

The candidate also passed the focused VM/bridge/compiler gate:

```sh
GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^(TestBytecodeVM_EnvironmentNameWritesMaterializeRawI32Values|TestMaterializeRuntimeValueBoxesBytecodeRawI32|TestBytecodeVM_CallName|TestBytecodeVM_I32RegisterFramePreseedsInlineCallNameDirect|TestBytecodeVM_.*Return|TestBytecodeVM_SlotlessInline|TestBytecodeVM_MinimalReturn|TestBytecodeVM_InlineReturnRestoresCallerActiveLookupCaches|TestBytecodeVM_Native.*MaterializesComputedIntegerArg)$' -count=1 -timeout 60s
GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/compiler/bridge -count=1 -timeout 60s
ABLE_COMPILER_FIXTURE_EXPERIMENTAL_EXECUTION_CONTEXT=1 GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/compiler -run '^(TestCompilerSafeNavigationMethodAndFieldExecute|TestCompilerPersistentMapEachExecutes|TestCompilerPersistentMapEachInfersClosureParamType|TestCompilerPersistentMapEachInferredClosureIgnoresOuterCallbackExpectedType|TestCompilerSpawnSiblingCapturedReceiverDispatchBuildsAndStaysNative)$' -count=1 -parallel=1 -timeout 60s
```

The safe-navigation and persistent-map regressions also pass with the
experimental option explicitly disabled.

## Correctness fixes found by the expanded gate

The direct suite found a raw bytecode i32 carrier crossing from an environment
or interpreter result into generated compiled code. It reproduced with the
execution-context option disabled, so it was not an ABI regression.

The retained general fixes are:

- Bytecode name writes materialize VM-only raw scalar carriers before storing
  them in an `Environment`; slots and stacks keep their existing raw fast path.
- The compiler bridge materializes raw carriers at its runtime-value boundary,
  including calls, lookup, member/index operations, operators, casts, ranges,
  awaits, and interpreter-evaluated definitions. Stable values take the
  allocation-free pass-through path.
- The generated runtime now resolves ordinary named fields of any
  `runtime.StructInstanceValue` before falling back to an interpreter-only
  member bridge. It handles both map-backed and positional named storage. This
  fixes static generated callbacks receiving runtime struct values; it is not a
  PersistentMap-specific lowering rule.

No benchmark-specific optimization, named-container compiler rule, production
ABI default change, or `able-stdlib` source change was made.

## Next gate

Run the remaining compiler package tests with the test-only candidate switch
under the same serial memory limits, then repeat the targeted race build. This
is the last broad semantic gate before deciding whether the opt-in ABI can
become the compiler default.
