# Able Project Log

# 2026-03-22 — Compiler Milestone 5 complete: compiled runtime core independence (v12)
- Closed `PLAN.md` Milestone 5 by moving remaining static compiled helper
  families onto direct Go runtime-core helpers instead of interpreter-oriented
  wrapper chains.
- Landed the Milestone 5 closure in:
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_call_resolution.go`
  - `v12/interpreters/go/pkg/compiler/generator_runtime_call_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_concurrency.go`
  - `v12/interpreters/go/pkg/compiler/generator_concurrency.go`
  - `v12/interpreters/go/pkg/compiler/generator_control_results.go`
  - `v12/interpreters/go/pkg/compiler/generator_function_type_normalize.go`
  - `v12/interpreters/go/pkg/compiler/generator_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_impl_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_helpers.go`
  - `v12/interpreters/go/pkg/compiler/compiler_runtime_core_independence_test.go`
- Static kernel/runtime helper calls now lower directly to `_impl` helpers on
  static paths for:
  - array helpers
  - hash-map helpers
  - string/char helpers
  - channel helpers
  - mutex helpers
- Helper-to-helper runtime-core bodies now avoid `__able_extern_call(...)`
  chaining in representative static families such as:
  - `__able_array_values`
  - `(__able_channel_awaitable).commit`
  - `(__able_mutex_awaitable).commit`
- Zero-arg callable syntax is now normalized consistently across:
  - direct callable lowering
  - static method specialization
  - spawned-task / await helper paths
- `Await.default({ => ... })` specialization now keeps the native zero-arg
  callback carrier instead of regressing to `runtime.Value -> T`.
- Spawned-task compiled control flow now returns through the runtime callback
  ABI using the explicit `runtime.Value, error` control envelope rather than
  compiled-function `(value, *__ableControl)` semantics.
- Added Milestone 5 regressions for:
  - direct static `_impl` helper usage
  - helper-body extern-chain elimination
  - zero-arg callable syntax normalization
  - `Await.default` zero-arg callback specialization
- Validation:
  - `go test ./pkg/compiler -run 'TestCompiler(StaticKernelHelpersUseDirectImplCalls|RuntimeHelperBodiesAvoidExternCallChaining|ZeroArgCallableTypeSyntaxStaysNative|AwaitDefaultZeroArgCallbackSpecializationStaysNative)$|TestCompilerExecFixtures/(06_01_compiler_spawn_await|06_12_02_stdlib_array_helpers|06_12_19_stdlib_concurrency_channel_mutex_queue)$' -count=1 -timeout 60s`
  - `git diff --check`

# 2026-03-22 — Compiler Milestone 4 complete: native dispatch completeness (v12)
- Closed `PLAN.md` Milestone 4 by removing the remaining shared static
  call/member/index/apply dispatch fallbacks instead of adding new
  nominal-type-specific lowering rules.
- Landed the Milestone 4 closure in:
  - `v12/interpreters/go/pkg/compiler/generator_dispatch_recovery.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections_static_array_access.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dispatch_completeness_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_test.go`
- Shared dispatch recovery now converts recoverable `runtime.Value` / `any`
  targets back onto native carriers before:
  - member access / assignment
  - method calls
  - index get / set
  - apply/call dispatch
- Local concrete and interface `Apply` bindings now stay on the shared static
  apply path instead of `__able_call_value(...)`.
- Mixed-source pure-generic interface dispatch now prefers the more concrete
  compiled specialization, which keeps representative generic interface calls
  off `__able_method_call_node(...)`.
- Validation:
  - `go test ./pkg/compiler -run 'TestCompiler(LocalConcreteApplyBindingStaysNative|LocalInterfaceApplyBindingStaysNative|DispatchTouchpointsStayNative|StaticNativePathsAvoidDynamicHelperReachability|BroadStaticNativeTouchpointsStayNative|GenericInterfaceTouchpointsStayNative|StaticIndexInterfacesStayNative|ConcreteReceiverInterfaceMethodStaysNative|ConcreteReceiverApplyStaysNative|InterfaceMethodWithLambdaArgStaysNative|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|ImportedGenericInterfaceAdapterRendersConcreteHelper|PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative)$|TestCompilerExecFixtures/(07_04_apply_callable_interface|10_04_interface_dispatch_defaults_generics|10_15_interface_default_generic_method|14_01_language_interfaces_index_apply_iterable)$' -count=1 -timeout 60s`
  - `git diff --check`

# 2026-03-22 — Compiler Milestone 3 complete: native pattern/control-flow completeness (v12)
- Closed `PLAN.md` Milestone 3 by finishing the remaining shared native
  pattern/control-flow gaps instead of adding new nominal-type-specific rules.
- Landed the final Milestone 3 closure in:
  - `v12/interpreters/go/pkg/compiler/generator_dynamic_typed_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_union_patterns.go`
  - `v12/interpreters/go/pkg/compiler/compiler_pattern_binding_native_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
- Recoverable typed pattern bindings now stay native, including:
  - rescue bindings onto `runtime.ErrorValue`
  - rescue bindings onto generated native interface carriers
  - native-union whole-value typed interface bindings
- Representative static pattern/control bodies are now source-audited against:
  - `__able_try_cast(...)`
  - `bridge.MatchType(...)`
  - `panic`
  - `recover`
  - IIFE-style scaffolding
- Validation:
  - `go test ./pkg/compiler -run 'TestCompiler(RescueTypedPatternBindingStaysNativeInterface|RescueTypedPatternBindingStaysNativeError|NativeUnionTypedPatternWholeValueBindingUsesNativeInterfaceCarrier|PatternControlTouchpointsStayNative|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|IfJoinRecoversTypeExprBackedInterfaceCarrier|IfJoinRecoversTypeExprBackedErrorCarrier|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|JoinExpressionConcreteGenericImplementersInferNativeInterface|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface|IfExpressionInterfaceAndNilInferNativeCarrier|MatchExpressionCallableAndNilInferNativeCarrier|RescueExpressionErrorAndNilInferNativeNullableError|NilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback|CommonExistentialJoinsExecuteWithoutDynamicFallback|ParameterizedInterfaceJoinsExecuteWithoutDynamicFallback|BroadStaticNativeTouchpointsStayNative|GenericInterfaceTouchpointsStayNative|StructPatternNamedFieldBinding(StaysNative|Executes))$|TestCompilerExecFixtures/(06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_rescue|06_01_compiler_or_else_error_union|06_01_compiler_loops|08_03_breakpoint_nonlocal_jump|11_02_option_result_or_handlers)$' -count=1 -timeout 60s`
  - `git diff --check`

# 2026-03-22 — Compiler Milestone 2 complete: native carrier completeness (v12)
- Closed `PLAN.md` Milestone 2 by removing the remaining shared
  representable-carrier widenings in the compiler type mapper and native union
  synthesis.
- Landed the carrier-completeness closure in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_builtin_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_result_void.go`
- Moved built-in `DivMod T` onto the shared nominal struct path and removed the
  old `any` fallback for concrete `DivMod` carriers.
- Tightened native union/result synthesis so fully bound representable members
  fail fast instead of silently collapsing to `runtime.Value` / `any`.
- Added shared native `Error | void` carrier support so `!void` signatures no
  longer regress to runtime result carriers.
- Added focused Milestone 2 regressions in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_carrier_completeness_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_result_test.go`
- Validation:
  - `go test ./pkg/compiler -run 'TestCompiler(DivModConcreteCarrierStaysNative|ParameterizedUnionAliasLocalStaysNative|ParameterizedResultAliasLocalStaysNative|ResultVoidReturnUsesNativeCarrier|StaticIndexInterfacesStayNative|GenericUnionAliasesStayNative|BroadNativeUnionExecutes|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|DirectErrorReturnUsesNativeCarrier|DirectErrorMessageAndCauseStayNative|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$' -count=1 -timeout 60s`
  - `go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_01_compiler_divmod|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|11_02_option_result_or_handlers)$' -count=1 -timeout 60s`
  - `git diff --check`

# 2026-03-21 — Compiler Milestone 1 complete: lowering facade centralized (v12)
- Closed compiler `PLAN.md` Milestone 1 by making lowering entrypoints explicit
  and shared instead of emitter-local.
- Landed canonical lowering facade files in:
  - `v12/interpreters/go/pkg/compiler/generator_lowering_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_boundaries.go`
- Refactored compiler emitters onto those entrypoints for:
  - type normalization and carrier synthesis
  - join/pattern synthesis
  - top-level dispatch synthesis
  - control-envelope checks
  - boundary/runtime conversion and wrapping
- Added `v12/interpreters/go/pkg/compiler/compiler_lowering_facade_audit_test.go`
  so future compiler changes fail if they bypass the lowering facade and call
  the lower-level synthesis helpers directly.
- Validation:
  - `go test ./pkg/compiler -run '^$'`
  - `go test ./pkg/compiler -run 'TestCompiler(LoweringFacadeSourceAudit|BroadStaticNativeTouchpointsStayNative|GenericInterfaceTouchpointsStayNative|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|StructPatternNamedFieldBinding(StaysNative|Executes))$|TestCompilerExecFixtures/(06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|11_02_option_result_or_handlers)$' -count=1 -timeout 60s`
  - `git diff --check`

# 2026-03-21 — PLAN reset around compiler completion first (v12)
- Rewrote `PLAN.md` so the working roadmap now reflects the actual priority
  order:
  1. finish the compiler
  2. make the bytecode interpreter fast
  3. backlog everything else
- Replaced the old mixed queue with a production-oriented compiler completion
  program with explicit milestones:
  - centralize lowering knowledge
  - native carrier completeness
  - native pattern/control-flow completeness
  - native dispatch completeness
  - compiled runtime core independence
  - boundary containment
  - compiler performance completion
  - compiler release validation
- Added a separate second-priority bytecode performance program and pushed
  integration/tooling/WASM/regex/other work into backlog.

# 2026-03-21 — v12 design-doc reconciliation pass (v12)
- Reconciled the active `v12/design` corpus around the current Go-first v12
  architecture.
- Updated:
  - `v12/design/README.md`
  - `v12/design/compiler-interpreter-vision.md`
  - `v12/design/interface-dispatch-dictionaries.md`
  - `v12/design/pattern-break-alignment.md`
  - `v12/design/testing-plan.md`
  - `v12/design/testing-cli-design.md`
  - `v12/design/channels-mutexes.md`
  - `v12/design/extern-host-execution.md`
  - `v12/design/compiler-native-lowering.md`
  - `v12/design/compiler-monomorphization.md`
  - `v12/design/monomorphized-container-abi.md`
  - `v12/design/compiler-union-abi.md`
  - `v12/design/stdlib-v12.md`
- Reconciliation results:
  - added a design-doc authority map so compiler work is explicitly governed by
    the new lowering spec/plan and by `compiler-aot.md`, not by older
    background notes;
  - downgraded stale TypeScript/Bun references in active Go-first docs to
    historical/future-runtime status where appropriate;
  - clarified that dictionary-dispatch notes govern interpreter/dynamic
    interface values, not the compiled static interface-lowering model;
  - clarified that named stdlib/container mentions in compiler docs are proof
    cases for shared lowering machinery, not licenses for nominal special cases;
  - removed stale `isize` / `usize` references from active compiler design docs
    and aligned them with the actual v12 spec primitive surface;
  - corrected the stale `Hash + PartialEq` wording in `stdlib-v12.md` to
    `Hash + Eq`, matching both the spec and `kernel-interfaces-hash-eq.md`.

# 2026-03-21 — Lowering docs/spec primitive-type reconciliation (v12)
- Reconciled the new compiler lowering docs against the actual v12 primitive
  type surface in `spec/full_spec_v12.md`.
- Corrected:
  - `v12/design/compiler-go-lowering-spec.md`
  - `spec/full_spec_v12.md`
- Changes:
  - removed the incorrect `isize` / `usize` mention from the new lowering spec;
  - added the actual spec primitives `i128` / `u128` to that table;
  - corrected the stray illustrative array-slice snippet from
    `start: usize, end: usize` to `start: u64, end: u64` so it matches the
    documented array helper signatures.

# 2026-03-21 — Compiler Go lowering spec + completion plan documented (v12)
- Wrote the canonical exhaustive Able -> Go lowering specification in:
  - `v12/design/compiler-go-lowering-spec.md`
- Wrote the separate ordered compiler completion plan in:
  - `v12/design/compiler-go-lowering-plan.md`
- Updated:
  - `PLAN.md`
  - `v12/design/compiler-native-lowering.md`
- Purpose of the reset:
  - compiler work must now be driven by one general lowering architecture and
    one ordered completion plan instead of a growing pile of emitter-local or
    type-specific fixes;
  - the compiler must compile Able semantics into direct Go implementations and
    may use the interpreters only as the semantic oracle and at explicit
    dynamic boundaries.

# 2026-03-21 — Compiler completion ladder reset (v12)
- Reframed the compiler roadmap around completion milestones instead of vague
  residual categories.
- The active native-lowering contract now tracks five high-level completion
  goals:
  - native carrier completeness
  - native pattern/control-flow completeness
  - native dispatch completeness
  - explicit dynamic-boundary containment
  - performance completeness
- Documented in:
  - `PLAN.md`
  - `v12/design/compiler-native-lowering.md`
- Also added a native compiler regression for named struct-pattern field
  renames (`field::binding`) so this control-flow/pattern milestone stays tied
  to concrete generated-code behavior rather than category labels.

# 2026-03-21 — Type-expression-backed native join inference tranche (v12)
- Closed the next shared `runtime.Value` / `any` proof gap on native compiled
  joins: when a branch/local still reports a dynamic Go carrier but retains a
  concrete normalized Able `TypeExpr`, the compiler now recovers the native
  carrier instead of widening the whole join back to `runtime.Value`.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_join_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_ident.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_result_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_rescue.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_union_patterns.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - shared join inference now accepts branch-local recovered `TypeExpr`
    metadata, not just the already-lowered Go carrier string;
  - typed-pattern bindings that currently have to travel through
    `runtime.Value` now preserve their concrete `TypeExpr`, including native
    union interface-branch bindings;
  - identifier lowering now prefers the recovered native carrier when a local
    is stored as `runtime.Value` but its bound `TypeExpr` is statically
    representable;
  - `if`, `match`, `rescue`, `or {}`, and loop/breakpoint result inference now
    all reuse that shared recovered-type path instead of treating
    `runtime.Value` / `any` as an immediate join failure.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfJoinRecoversTypeExprBackedInterfaceCarrier|IfJoinRecoversTypeExprBackedErrorCarrier|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|IfExpressionInterfaceAndNilInferNativeCarrier|MatchExpressionCallableAndNilInferNativeCarrier|RescueExpressionErrorAndNilInferNativeNullableError|NilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface)$' -count=1 -timeout 60s` (pass, `1.073s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfJoinRecoversTypeExprBackedInterfaceCarrier|IfJoinRecoversTypeExprBackedErrorCarrier|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|IfExpressionInterfaceAndNilInferNativeCarrier|MatchExpressionCallableAndNilInferNativeCarrier|RescueExpressionErrorAndNilInferNativeNullableError|NilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface|CommonExistentialJoinsExecuteWithoutDynamicFallback|ParameterizedInterfaceJoinsExecuteWithoutDynamicFallback|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$|TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1 -timeout 60s` (pass, `21.319s`)
  - `git diff --check` (pass)
- Follow-up backlog is now explicit instead of categorical:
  - close `types.go` / `generator_native_unions.go` carrier-synthesis fallbacks
    for fully bound normalized union/result/nullable `TypeExpr`s;
  - remove recovered-subject typed-pattern fallback to `__able_try_cast(...)`
    in `generator_match.go`, `generator_match_runtime_types.go`, and
    `generator_rescue.go`;
  - remove remaining representable join/binding fallback locals in
    `generator_or_else.go`, `generator_rescue.go`, and
    `generator_join_types.go`;
  - tighten `generator_native_unions.go` residual runtime-member gates so
    `runtime.Value` union members remain only for true dynamic payloads.

# 2026-03-21 — Nil-capable native join inference tranche (v12)
- Closed the next shared `runtime.Value` / `any` proof gap on native compiled
  joins: `nil` branches in `if`, `match`, `rescue`, `or {}`, and loop/
  breakpoint result inference no longer force the whole join back to
  `runtime.Value` when the non-nil branches already share a native carrier.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_join_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_result_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_rescue.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - shared join inference now recognizes nil-valued branch expressions
    (`any(nil)`, `runtime.NilValue{}`, and typed nils) separately from the
    concrete branch carriers;
  - the compiler now joins the non-nil carriers first and then preserves that
    native result when the joined carrier already has a nil zero value or a
    native nullable wrapper exists;
  - this closes nil-capable native joins for interface carriers, callable
    carriers, native nullable/error carriers, and the existing loop/breakpoint
    result probes through one shared rule rather than per-construct handling.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfExpressionInterfaceAndNilInferNativeCarrier|MatchExpressionCallableAndNilInferNativeCarrier|RescueExpressionErrorAndNilInferNativeNullableError|NilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback)$' -count=1 -timeout 60s` (pass, `1.080s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|ImportedGenericInterfaceAdapterRendersConcreteHelper|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteGenericImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface|IfExpressionInterfaceAndNilInferNativeCarrier|MatchExpressionCallableAndNilInferNativeCarrier|RescueExpressionErrorAndNilInferNativeNullableError|CommonExistentialJoinsExecuteWithoutDynamicFallback|ParameterizedInterfaceJoinsExecuteWithoutDynamicFallback|NilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$|TestCompilerGenericInterfaceTouchpointsStayNative$|TestCompilerInterfaceLookupGenericMethodFixturesRegression$|TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1 -timeout 60s` (pass, `26.868s`)
  - `git diff --check` (pass)
- Follow-up:
  - this category is now closed; the next remaining category is broader
    union/result/existential lowering beyond the native slices already landed.

# 2026-03-21 — Parameterized existential join-carrier inference tranche (v12)
- Closed the next shared residual existential/join gap on native compiled
  paths: fully bound parameterized interface joins now infer the common native
  interface carrier instead of degrading to a synthesized union or
  `runtime.Value`.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_join_types.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - shared join inference now materializes candidate native interface carriers
    from the actual branch impl metadata instead of depending only on
    already-loaded zero-arg interface infos;
  - that materialization walks bound base interfaces too, so concrete
    implementers of different parameterized child interfaces can still join on
    the common bound parent carrier;
  - direct bound joins like `Left | Right -> Reader i32` now stay on
    `__able_iface_Reader_i32`, and inherited bound joins like
    `LeftReader i32 | RightReader i32 -> Reader i32` do the same without
    adding any nominal-type-specific lowering rule.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface|ParameterizedInterfaceJoinsExecuteWithoutDynamicFallback)$' -count=1 -timeout 60s` (pass, `1.132s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteGenericImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface|CommonExistentialJoinsExecuteWithoutDynamicFallback|ParameterizedInterfaceJoinsExecuteWithoutDynamicFallback|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$|TestCompilerInterfaceLookupGenericMethodFixturesRegression$|TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1 -timeout 60s` (pass, `24.953s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(GenericInterfaceTouchpointsStayNative|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|ImportedGenericInterfaceAdapterRendersConcreteHelper|JoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface|JoinExpressionParameterizedInheritedImplementersInferSharedParentInterface)$' -count=1 -timeout 60s` (pass, `1.347s`)
  - `git diff --check` (pass)
- Follow-up:
  - the next residual existential-proof category is broader `runtime.Value` /
    `any` elimination on unresolved generic/result/existential joins, then
    the remaining union/result/existential lowering beyond the already-landed
    native slices.

# 2026-03-21 — Common existential join-carrier inference tranche (v12)
- Closed the next shared residual join/existential gap on native compiled
  paths: mixed concrete branches that share a native existential carrier no
  longer have to degrade into a union-plus-dynamic-dispatch path first.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_join_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - shared join inference now prefers a common native `runtime.ErrorValue`
    carrier when all concrete branches are native `Error` implementers,
    instead of synthesizing a native union and dispatching back through
    dynamic member helpers;
  - shared join inference now also prefers a common zero-arg native interface
    carrier when all concrete branches implement the same interface, so mixed
    concrete joins like `Cat | Dog -> Speak` stay on the interface carrier
    directly instead of materializing `__able_union_*` locals and then calling
    `__able_method_call_node(...)`;
  - the generic-interface dispatch path now keeps working after that join
    inference when multiple concrete adapters exist for the same generic-only
    interface: adapter-case specialization now refines the concrete receiver
    target before selecting the compiled impl, so joined `Echo` carriers still
    call the compiled generic dispatch helper rather than falling back through
    the runtime adapter layer.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|InterfaceParamAndReturnStayNative|TypedInterfaceAssignmentStaysNative|JoinExpressionConcreteImplementersInferNativeInterface|JoinExpressionConcreteGenericImplementersInferNativeInterface|JoinExpressionConcreteErrorsInferNativeErrorCarrier|CommonExistentialJoinsExecuteWithoutDynamicFallback|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$' -count=1 -timeout 60s` (pass, `2.887s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$|TestCompilerInterfaceLookupGenericMethodFixturesRegression$' -count=1 -timeout 60s` (pass, `20.843s`)
  - `git diff --check` (pass)
- Follow-up:
  - the next residual existential-proof category is broader common-carrier
    discovery for fully bound parameterized interface joins plus the remaining
    broader union/result/existential lowering beyond the already-landed native
    slices.

# 2026-03-21 — Cross-package generic-interface adapter retention tranche (v12)
- Closed the remaining shared generic-interface adapter-emission gap exposed by
  cross-package generic/default-method fixtures.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_coercions.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_generic_test.go`
- Tranche details:
  - generic-only native interface adapters that are proven during shared
    refresh are now retained in an explicit adapter set for later render-time
    use, instead of being dropped by a subsequent adapter refresh before code
    emission;
  - render-time native interface adapter emission now runs to a fixed point,
    so concrete adapters discovered late through shared direct-coercion helper
    generation are still emitted without adding new nominal-type rules;
  - the new regression `TestCompilerImportedGenericInterfaceAdapterRendersConcreteHelper`
    pins the cross-package `Tokenizer <- Prefixer` helper/type emission that
    previously built broken source.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerImportedGenericInterfaceAdapterRendersConcreteHelper$|TestCompilerInterfaceLookupGenericMethodFixturesRegression$' -count=1 -timeout 60s` (pass, `5.070s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorCollectGenericNominalAccumulatorStaysNative$|TestCompilerConcreteIteratorCollectGenericNominalAccumulatorExecutes$' -count=1 -timeout 60s` (pass, `38.415s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorGenericMethodsStayNativeWithExperimentalMonoArrays$|TestCompilerConcreteIteratorGenericMethodsExecuteWithExperimentalMonoArrays$' -count=1 -timeout 60s` (pass, `48.512s`)
  - `cd v12/interpreters/go && git diff --check` (pass)
- Follow-up:
  - the next genuinely different category remains broader residual
    `runtime.Value` / `any` elimination for unresolved
    generic/result/existential proof gaps beyond the now-closed generic
    adapter-retention hole.

# 2026-03-21 — Native loop/breakpoint join tranche (v12)
- Closed the remaining shared loop/breakpoint join gap on native compiled
  control-flow paths.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_compile_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_context.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_counted_loops.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_result_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_iterables.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - `loop { ... }` result typing now probes break payloads through the real
    compiled body context, so statically representable break-value shapes stay
    on native carriers instead of defaulting the loop result temp to
    `runtime.Value`;
  - labeled `breakpoint` expressions now do the same for mixed normal exits
    and labeled `break` payloads, including cases where the break payload
    depends on earlier locals in the breakpoint body;
  - labeled/non-local `break` with payloads now coerce directly onto the
    native result carrier, and bare `break` now writes the correct nil payload
    when a loop result temp is present instead of accidentally reusing the
    normal-completion `void` sentinel.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion)$' -count=1` (pass, `1.134s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(LoopExpressionBreakValuesInferNativeUnion|BreakpointExpressionMixedExitsInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion)$|TestCompilerExecFixtures/(06_01_compiler_loops|08_03_breakpoint_nonlocal_jump|06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1` (pass)
- Follow-up:
  - the next genuinely different category remains broader residual
    `runtime.Value` / `any` elimination for unresolved
    generic/result/existential proof gaps beyond the now-closed control-flow
    join family.

# 2026-03-21 — Native `or {}` join and handler-binding tranche (v12)
- Closed the residual shared `or {}` join/binding gap on native compiled
  paths.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - `or {}` now uses the same shared join-type inference as mixed-result
    `if`/`match`/`rescue`, so representable mixed success/handler result shapes
    stay on native carriers instead of defaulting to `runtime.Value`;
  - nullable success paths now join on the unwrapped payload carrier rather
    than the pointer carrier, which keeps mixed nullable success/handler joins
    native;
  - when `or { err => ... }` is handling a statically-known native failure
    branch, the handler binding now stays on that native failure carrier
    instead of being forced through `runtime.Value`, which fixes the compiled
    option/result handler fixture path.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|NullableI32ReturnAndOrElseStayNative|ResultReturnUsesNativeCarrier|OrElseOnErrorUnionUsesNativeCarrierDetection)$' -count=1` (pass, `1.070s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|OrElseOnNullableMixedBranchesInferNativeUnion|OrElseOnErrorUnionMixedBranchesInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|ResultReturnUsesNativeCarrier|OrElseOnErrorUnionUsesNativeCarrierDetection|InlineClosedUnionParamAndMatchStayNative)$|TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1` (pass, `15.725s`)
- Follow-up:
  - the next genuinely different category is broader residual `runtime.Value`
    / `any` elimination for unresolved generic/result/existential proof gaps
    beyond the now-closed join and `or {}` surfaces.

# 2026-03-21 — Native join-result inference tranche (v12)
- Closed the shared mixed-result join gap on native compiled control-flow
  expressions.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_join_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_rescue.go`
  - `v12/interpreters/go/pkg/compiler/compiler_join_native_test.go`
- Tranche details:
  - mixed-result `if`, `match`, and `rescue` expressions now try to infer a
    native compiled carrier when all branch result types are statically
    representable, instead of immediately collapsing the join local to
    `runtime.Value`;
  - branch coercion now reuses the shared native static-coercion path, so
    native unions/interfaces/callables/nullable carriers can serve as the join
    result directly;
  - static typed-pattern lowering now accepts same-family nominal carriers via
    shared receiver compatibility rather than requiring exact Go carrier
    identity, which keeps specialized native carriers on the static pattern
    path.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback)$' -count=1` (pass, `1.025s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|RescueExpressionMixedBranchesInferNativeUnion|JoinExpressionsExecuteWithoutRuntimeCarrierFallback|RescueStatementMixedResultTypesNoFallback|NullableI32ReturnAndOrElseStayNative|ResultReturnUsesNativeCarrier|InlineClosedUnionParamAndMatchStayNative)$|TestCompilerExecFixtures/(06_01_compiler_rescue|06_01_compiler_match_patterns|06_01_compiler_if_block_exprs)$' -count=1` (pass, `10.680s`)
- Follow-up:
  - the next genuinely different category is broader residual `runtime.Value`
    / `any` elimination for unresolved generic/result/join proof gaps, not
    more control-flow join cleanup.

# 2026-03-21 — Static generic-interface dispatch helper tranche (v12)
- Closed the residual static generic-interface runtime edge for the current
  compiler-native lowering work.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interface_generic_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_generic_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
- Tranche details:
  - statically-known generic interface calls such as `Echo.pass<T>(...)` now
    lower through generated compiled dispatch helpers on the native interface
    carrier instead of converting the receiver through
    `__able_iface_*_to_runtime_value(...)` plus `__able_method_call_node(...)`;
  - those helpers use the existing shared impl-specialization machinery to
    call specialized compiled impls directly for concrete adapter cases;
  - the runtime-adapter case remains in the generated helper as the explicit
    dynamic boundary for interface values that already originate from runtime
    payloads;
  - adapter refresh now preserves already-proven concrete adapters instead of
    dropping them when refreshing generic-only interface carrier state.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerPureGenericInterfaceAssignmentUsesNativeCarrier$' -count=1` (pass, `0.028s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(GenericInterfaceTouchpointsStayNative|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes)$' -count=1` (pass, `0.756s`)
  - `cd v12/interpreters/go && timeout 40s env GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupGenericMethodFixturesRegression$' -count=1` (pass, `4.655s`)
  - `git diff --check` (pass)
- Follow-up:
  - the next genuinely different category is broader residual `runtime.Value`
    / `any` elimination for unresolved generic/result/join surfaces, not
    another generic-interface runtime-edge cleanup.

# 2026-03-21 — Shared upper-bound addition range-proof tranche on the matrix path (v12)
- Closed the next shared primitive-lowering gap on the matrix benchmark family
  by proving fixed-width signed addition safe when both operands are
  statically non-negative and bounded, without adding any named non-primitive
  lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/model.go`
  - `v12/interpreters/go/pkg/compiler/generator_compile_context.go`
  - `v12/interpreters/go/pkg/compiler/generator_integer_facts.go`
  - `v12/interpreters/go/pkg/compiler/generator_integer_param_facts.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_counted_loops.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary_checked_inline.go`
  - `v12/interpreters/go/pkg/compiler/generator_render.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
- Tranche details:
  - the compiler now carries simple primitive upper-bound facts across
    statically resolved function calls and seeds them back into callee param
    contexts before render;
  - counted-loop induction variables now inherit a conservative upper bound
    from their loop guard when the bound is statically known;
  - inline checked signed addition now lowers directly when both operands are
    proven non-negative and their combined upper bound fits the target width;
  - the matrix compile-shape audit now proves `build_matrix` lowers both
    `i - j` and `i + j` as direct signed arithmetic with no widened
    `int64(...)` affine branch scaffolding left in the inner loop.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1267s`, `7.00` GC
  - `examples/benchmarks/matrixmultiply`: `1.1367s`, `13.00` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - the matrix inner-loop affine add/sub branch gap is now closed through one
    shared primitive/function range-proof rule;
  - wall-clock remains in the same band, so the current best matrix timings
    are still the earlier counted-loop snapshot;
  - the next worthwhile category is no longer affine loop arithmetic on the
    matrix path, but a different shared residual or benchmark family.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineCheckedIntegerAddSubStayStatic|InlineCheckedSignedSubWithNonNegativeOperandsElidesOverflowBranch|InlineCheckedSignedAddWithCallsiteUpperBoundElidesOverflowBranch|WhileLoopFastPath|CountedLoopFastPath|ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|MatrixMultiplyCountedLoopsStayNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes))$' -count=1` (pass, `4.390s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.1500s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1267s`, `7.00` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `1.3200s`, `13.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `1.1367s`, `13.00` GC)
  - `git diff --check` (pass)

# 2026-03-21 — Shared non-negative subtraction range-proof tranche on the matrix path (v12)
- Closed the next shared primitive-lowering gap on the matrix benchmark family
  by proving signed subtraction safe when both fixed-width operands are
  statically known non-negative, without adding any named non-primitive
  lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_context.go`
  - `v12/interpreters/go/pkg/compiler/generator_integer_facts.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_local_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary_checked_inline.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
- Tranche details:
  - the compile context now tracks simple primitive integer sign facts per Go
    binding and carries them into child scopes while clearing them on
    rebinding/shadowing;
  - static assignments from non-negative integer expressions now preserve that
    fact for later primitive lowering;
  - inline checked signed subtraction now lowers directly when both operands
    are proven non-negative, so it no longer emits widened `int64(...)`
    subtraction plus overflow-branch scaffolding;
  - the matrix compile-shape audit now proves `build_matrix` lowers `i - j`
    as a direct signed subtraction, while the widened checked `i + j` path
    remains as the next primitive residual.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1167s`, `7.00` GC
  - `examples/benchmarks/matrixmultiply`: `1.1000s`, `13.00` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - the subtraction-side inline overflow-branch gap is now closed through one
    shared primitive range-proof rule;
  - this remains effectively performance-neutral on the matrix family, so the
    next primitive residual is stronger upper-bound range proofs for affine
    addition like `i + j`, not subtraction.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineCheckedIntegerAddSubStayStatic|InlineCheckedSignedSubWithNonNegativeOperandsElidesOverflowBranch|WhileLoopFastPath|CountedLoopFastPath|ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|MatrixMultiplyCountedLoopsStayNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes))$' -count=1` (pass, `3.723s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.1200s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1167s`, `7.00` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `1.2000s`, `13.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `1.1000s`, `13.00` GC)
  - `git diff --check` (pass)

# 2026-03-21 — Shared inline affine integer-check tranche on the matrix path (v12)
- Closed the next shared primitive-lowering gap on the matrix benchmark family
  by inlining fixed-width checked integer `+` / `-` on static compiled paths,
  without adding any named non-primitive lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary_checked_inline.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
- Tranche details:
  - shared checked addition/subtraction for fixed-width integers under 64 bits
    now lowers inline on static compiled paths instead of calling
    `__able_checked_add_signed(...)`, `__able_checked_sub_signed(...)`, and
    their unsigned equivalents
  - this is intentionally narrow: `int`, `uint`, `i64`, and `u64` stay on the
    existing helper path because those widths still rely on the wider-width or
    runtime-width overflow machinery
  - the compile-shape audit now proves `build_matrix` no longer carries the
    signed checked helper calls; its affine integer ops now lower as inline
    `int64(...) +/- int64(...)` plus explicit range checks
  - this remains within the compiler contract documented in `PLAN.md`: only
    built-in `Array` semantics and primitive types receive special lowering;
    non-primitive nominal types still rely on shared
    struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1133s`, `7.00` GC
  - `examples/benchmarks/matrixmultiply`: `1.0867s`, `13.00` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - the affine checked-helper call gap is now closed through one shared
    primitive rule
  - this workload is effectively perf-neutral after the earlier counted-loop
    win, so the remaining primitive residual is now the inline overflow
    branches themselves where static range proofs are available
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineCheckedIntegerAddSubStayStatic|WhileLoopFastPath|CountedLoopFastPath|ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|MatrixMultiplyCountedLoopsStayNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes))$' -count=1` (pass, `3.929s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.1200s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1133s`, `7.00` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `1.0700s`, `13.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `1.0867s`, `13.00` GC)
  - `git diff --check` (pass)

# 2026-03-20 — Shared counted-loop fast-path tranche on the matrix path (v12)
- Closed the next shared primitive/control-flow lowering gap on the matrix
  benchmark family by lowering canonical counted `loop` forms directly to Go
  counted loops, without adding any named non-primitive lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_counted_loops.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
- Tranche details:
  - shared counted-loop recognition now lowers canonical
    `loop { if i >= n { break } ... i = i + 1 }` shapes to direct
    `for i < n { ... i++ }` when the induction variable and bound stay on
    primitive integer carriers and the loop body does not mutate the counter
    outside the trailing increment
  - the matcher is conservative by construction: it now traverses nested
    function/lambda/iterator/ensure bodies too, so the fast path rejects loop
    bodies that can still mutate the counter through nested control flow
  - the compile-shape audit now proves `build_matrix` and `matmul` stay on
    direct counted loops, and `matmul` no longer carries
    `__able_checked_add_signed(...)` for loop induction
  - this remains within the compiler contract documented in `PLAN.md`: only
    built-in `Array` semantics and primitive types receive special lowering;
    non-primitive nominal types still rely on shared
    struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1133s`, `7.00` GC
  - `examples/benchmarks/matrixmultiply`: `1.0833s`, `13.00` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - loop-induction checked arithmetic was the next shared primitive residual
    on the matrix family
  - that counted-loop gap is now closed
  - the next primitive residual is affine checked integer arithmetic such as
    `i - j` / `i + j` inside `build_matrix`, not loop-control scaffolding
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(WhileLoopFastPath|CountedLoopFastPath|ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|MatrixMultiplyCountedLoopsStayNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes))$' -count=1` (pass, `4.035s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.1100s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1133s`, `7.00` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `1.0600s`, `13.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `1.0833s`, `13.00` GC)
  - `git diff --check` (pass)

# 2026-03-20 — Shared propagated static-array accessor pointer-elision tranche on the matrix path (v12)
- Closed the next shared built-in `Array` lowering gap on the matrix benchmark
  family by removing pointer-backed nullable-carrier construction from
  propagated static array accessors, without adding any named non-primitive
  lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
- Tranche details:
  - propagated static built-in `Array` accessors (`get`, `first`, `last`,
    `read_slot`, `pop`) now lower as direct bounds-check + element-load paths
    with nil control transfer instead of manufacturing pointer-backed nullable
    carriers on the success path
  - the generated `build_matrix`, `matmul`, and `main` bodies for the matrix
    benchmark family are now free of `__able_ptr(...)` in those propagated
    static array access paths
  - this remains within the compiler contract documented in `PLAN.md`: only
    built-in `Array` semantics and primitive types receive special lowering;
    non-primitive nominal types still rely on shared
    struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1967s`, `7.33` GC
  - `examples/benchmarks/matrixmultiply`: `3.4367s`, `13.67` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - propagated static-array accessor pointer construction was the next shared
    macro-scale residual on the matrix family
  - that pointer-carrier gap is now closed
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes)|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative)$' -count=1` (pass, `2.497s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.2000s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1967s`, `7.33` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `3.5300s`, `14.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `3.4367s`, `13.67` GC)
  - `git diff --check` (pass)

# 2026-03-20 — Shared static built-in `Array` frame-elision tranche on the matrix path (v12)
- Closed the remaining macro-scale shared built-in `Array` lowering gap on the
  matrix benchmark family by removing synthetic call-frame scaffolding from
  static array factories and intrinsics, without adding any named
  non-primitive lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_intrinsics.go`
  - `v12/interpreters/go/pkg/compiler/generator_static_array_factories.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections_static_array_access.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
- Tranche details:
  - shared static built-in `Array` factories and intrinsics no longer emit
    synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()` pairs
    on compiled static paths
  - the generated `build_matrix` and `matmul` bodies are now free of that
    frame scaffolding while staying on the same shared `Array` lowering rules
  - this remains within the compiler contract documented in `PLAN.md`: only
    built-in `Array` semantics and primitive types receive special lowering;
    non-primitive nominal types still rely on shared
    struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `0.1933s`, `7.00` GC
  - `examples/benchmarks/matrixmultiply`: `4.2267s`, `13.00` GC
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`
  - direct compiled `ablec` output for
    `v12/examples/benchmarks/matrixmultiply.able`:
    `-95.58358333329998`
- Conclusion:
  - synthetic static-array frame churn was the dominant remaining macro-scale
    built-in `Array` cost on the matrix family
  - the current matrix built-in `Array` tranche is now closed
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes)|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative)$' -count=1` (pass, `2.910s`)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `0.2000s`, `7.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `0.1933s`, `7.00` GC)
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/examples/benchmarks/matrixmultiply.able` (pass, output `-95.58358333329998`, `4.0400s`, `13.00` GC)
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (pass, `4.2267s`, `13.00` GC)
  - `git diff --check` (pass)

# 2026-03-20 — Shared native float-to-int cast tranche on the matrix entry path (v12)
- Closed the next shared primitive-lowering gap on the matrix benchmark family
  by removing the remaining `float -> int` runtime-cast crossings from the
  compiled benchmark entry path, without adding any named non-primitive
  lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
- Tranche details:
  - shared primitive `float -> int` casts now lower through native
    `math.Trunc(...)` plus explicit `NaN` / `Inf` / range overflow checks on
    static compiled paths instead of round-tripping through `__able_cast(...)`
    and `bridge.AsInt(...)`
  - the full `matrixmultiply` compile-shape audit now proves the compiled
    `main` body avoids those runtime cast helpers while `build_matrix` and
    `matmul` stay free of the known scalar runtime carrier helpers
  - this remains within the compiler contract documented in `PLAN.md`: only
    built-in `Array` semantics and primitive types receive special lowering;
    non-primitive nominal types still rely on shared
    struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `1.7567s`, `7.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`.
- Macro smoke:
  - the full compiled `v12/examples/benchmarks/matrixmultiply.able` path still
    times out under the current `60s` harness budget.
- Conclusion:
  - the primitive entry-cast gap on the matrix family is now closed through
    shared primitive lowering rules
  - the next category is the remaining macro-scale built-in `Array` lowering
    work on the full matrix path beyond the entry casts
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArrays(MatrixMultiplyScalarLoopStaysNative|MatrixMultiplyMainStaysNative|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes)|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative)$' -count=1` (pass, `2.782s`).
  - `git diff --check` (pass).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `1.9700s`, `7.00` GC).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `1.7567s`, `7.00` GC).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (timeout).

# 2026-03-20 — Shared native scalar array-propagation tranche on the reduced matrix path (v12)
- Closed the next shared AOT lowering gap on the reduced matrix benchmark by
  removing residual built-in `Array` scalar propagation/runtime-cast crossings,
  without introducing any named-structure lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-scalar-propagation-compiled.md`
- Tranche details:
  - static built-in `Array` propagation now returns concrete success element
    types on the compiled path, so nested `get(...)!` / index propagation on
    staged scalar arrays no longer routes success values through
    `__able_nullable_*_to_value(...)`
  - the reduced matrix hot loop now stays on direct Go `float64` multiply/add
    instead of falling back through `bridge.AsFloat(...)` and
    `__able_binary_op(...)`
  - explicit primitive numeric casts such as `i32 -> f64` and float widening
    now lower directly to Go casts on static compiled paths instead of
    round-tripping through `__able_cast(...)`
  - this remains within the compiler contract documented in `PLAN.md`:
    primitive types may use primitive-specific lowering, while non-primitive
    nominal types still rely on shared struct/union/interface/generic rules
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-scalar-propagation-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: `1.9733s`, `7.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    `-28.500833332098754`.
- Macro smoke:
  - the full compiled `v12/examples/benchmarks/matrixmultiply.able` path still
    times out under the current `60s` harness budget.
- Conclusion:
  - the reduced matrix scalar-loop carrier gap is now closed through shared
    built-in `Array` and primitive lowering rules
  - the next category is the remaining macro-scale built-in `Array` lowering
    work on the full matrix path
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|ExperimentalMonoArrays(NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|MatrixMultiplyScalarLoopStaysNative))$' -count=1` (pass, `1.487s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, output `-28.500833332098754`, `1.6700s`, `7.00` GC).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass, `1.9733s`, `7.00` GC).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (timeout).

# 2026-03-20 — Shared static nominal receiver/struct-literal closure tranche (v12)
- Closed the remaining shared generic nominal default/static-method lowering
  gap on the reduced `LinkedList -> Enumerable -> LazySeq` path without adding
  any named-structure rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_type_substitute.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_nominal_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-shared-static-nominal-closure-compiled.md`
- Tranche details:
  - recursive type substitution now resolves chained bindings transitively,
    which prevents specialized nominal/default method bodies from stalling on
    placeholder bindings that still point at other placeholders
  - static nominal target refinement now upgrades bare targets and struct
    literals to the concrete expected carrier when the specialization context
    already knows the nominal target (for example `LazySeq<T>` -> `LazySeq<i32>`)
  - native interface concrete-impl matching now compares against the
    specialized target template, so specialized receivers like `*LinkedList_i32`
    satisfy compiled `Iterable<i32>` adapters through the shared path
  - the surfaced residual fallback gap is closed for:
    `ConcreteEnumerableGenericMethods...`,
    `LinkedListIterableAdapter...`, and
    `LazySeqIteratorCarrier...`
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-shared-static-nominal-closure-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_enumerable_i32_small`: `0.1633s`, `8.33` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`:
    `382455000`.
- Conclusion:
  - the shared generic nominal default/static receiver and struct-literal path
    is now closed on the reduced `LinkedList -> Enumerable -> LazySeq` family
  - the next category is the next benchmark-worthy generic
    container/runtime edge that still crosses residual runtime carriers
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerLinkedListIterableAdapterStaysNative$' -count=1` (pass, `12.668s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerLazySeqIteratorCarrierStaysNative$' -count=1` (pass, `31.403s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerConcreteEnumerableGenericMethodsStayNative$' -count=1` (pass, `30.409s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerConcreteEnumerableGenericMethodsExecute$' -count=1` (pass, `14.742s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerConcreteIteratorCollectGenericNominalAccumulatorExecutes$' -count=1` (pass, `12.071s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerConcreteIteratorMapFilterFunctionExecutes$' -count=1` (pass, `12.050s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `37.308s`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass, `0.1633s`, `8.33` GC).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass, output `382455000`).
  - `git diff --check` (pass).

# 2026-03-20 — Bound generic field/member carrier refinement tranche (v12)
- Closed the next shared AOT lowering gap by keeping fully bound generic
  struct fields/members on their concrete native carriers inside already-
  specialized nominal method bodies, without adding a named-structure rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_generic_nominal_inference.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections_static_array_access.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_impl_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_nominal_methods.go`
  - `v12/interpreters/go/pkg/compiler/compiler_nominal_method_specialization_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_array_family_native_test.go`
  - `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
- Tranche details:
  - Added a user-defined proof case (`Bucket T { items: Array T }`) that runs
    under `ExperimentalMonoArrays` and proves a fully bound generic field now
    lowers to its concrete native carrier (`Items *__able_array_i32`) inside
    specialized nominal method bodies.
  - Static array index lowering now returns concrete element types directly
    when an expected type is known, with explicit control transfer on bounds
    failures, instead of materializing a temporary `runtime.Value` and
    converting back.
  - Receiver-derived specialization bindings now upgrade placeholder self-
    bindings like `T -> T` to concrete bindings like `T -> i64`; that closes
    the remaining mono-array `Iterable.iterator` / `Iterable.each` execute gap
    inside specialized default/nominal method bodies.
  - The same shared fix closes the surfaced `PersistentSortedQueue`
    specialization regression: compiled main bodies now stay on
    `PersistentSortedSet_i32` / `PersistentQueue_i32` plus specialized impl
    siblings instead of falling back through unspecialized zero-arg helpers.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/heap_i32_small`: `0.7667s`, `91.33` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/heap_i32_small/main.able`: `-211812354`.
- Conclusion:
  - the bound generic field/member carrier refinement tranche is now closed
  - the next category is the next benchmark-worthy generic
    container/runtime edge that still crosses residual runtime carriers
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(BoundGenericFieldCarrierSpecialization(StaysNative|Executes)|GenericNominalMethodSpecialization(StaysNative|Executes)|HeapGenericMethodSpecializationStaysNative)$' -count=1` (pass, `6.175s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerPersistentSortedQueueMethodsStayNative$' -count=1` (pass, `8.404s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorGenericMethods(StayNativeWithExperimentalMonoArrays|ExecuteWithExperimentalMonoArrays)$' -count=1` (pass, `27.013s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorCollectGenericNominalAccumulator(StaysNative|Executes)$' -count=1` (pass, `21.305s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorMapFilterFunction(StaysNative|Executes)$' -count=1` (pass, `24.609s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `33.052s`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/heap_i32_small/main.able` (pass, `0.7667s`, `91.33` GC).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/heap_i32_small/main.able` (pass, output `-211812354`).
  - `git diff --check` (pass).

# 2026-03-20 — Shared generic nominal `methods` specialization tranche (v12)
- Closed the next shared AOT lowering gap by specializing generic nominal
  `methods` blocks when the concrete target type is statically known,
  without adding another named-structure rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_nominal_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_impl_calls.go`
  - `v12/interpreters/go/pkg/compiler/compiler_nominal_method_specialization_test.go`
  - `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
- Tranche details:
  - Method param/return mapping now honors bound type bindings during nominal
    method specialization, so a generic `methods Box T` or `methods Heap T`
    body can render concrete compiled signatures once the target is known
    statically (`T -> i32`, etc.).
  - Added shared nominal-method specialization parallel to the existing impl
    specialization path; callsites now pick specialized compiled method bodies
    for concrete nominal targets instead of reusing the unspecialized
    `runtime.Value` signatures.
  - Added a user-defined proof case (`Box T`) to pin that this is a general
    nominal-lowering rule, not a container-specific hack.
  - Re-measured the reduced `Heap i32` benchmark because it is the first hot
    constrained generic nominal-method path that materially exercises this
    lowering.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/heap_i32_small`: `4.2000s`, `1811.67` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/heap_i32_small/main.able`: `-211812354`.
- Conclusion:
  - the shared generic nominal-method specialization layer is now closed
  - the next category is bound generic field/member carrier refinement inside
    already-specialized nominal method bodies, not another per-structure rule
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(GenericNominalMethodSpecialization(StaysNative|Executes)|HeapGenericMethodSpecializationStaysNative)$' -count=1` (pass, `2.286s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(GenericNominalMethodSpecialization(StaysNative|Executes)|HeapGenericMethodSpecializationStaysNative|BitSetHeapMethodsStayNative|DequeQueueMethodsStayNative|PersistentSortedQueueMethodsStayNative|ConcreteIteratorCollectGenericNominalAccumulator(StaysNative|Executes)|ConcreteIteratorGenericMethods(StayNative|Execute|StayNativeWithExperimentalMonoArrays|ExecuteWithExperimentalMonoArrays)|ConcreteIteratorMapFilterFunction(StaysNative|Executes))$|TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `20.995s`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/heap_i32_small/main.able` (pass, `4.2000s`, `1811.67` GC).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/heap_i32_small/main.able` (pass, output `-211812354`).
  - `git diff --check` (pass).

# 2026-03-20 — Generic nominal `Iterator.collect<C>()` proof tranche (v12)
- Closed the next generic nominal lowering guardrail on the compiler path
  without adding a new named-structure rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_collect_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_render.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
- Tranche details:
  - Added a user-defined `SumCount` accumulator implementing only
    `Default + Extend i64` and pinned `Iterator.collect<SumCount>()` against
    `__able_method_call_node(...)`, `__able_call_value(...)`, and iterator
    runtime-value conversion on the compiled path.
  - This proves the shared generic default-method lowering now carries
    statically known user nominal accumulators through the compiled
    `Iterator.collect<C>()` helper without another nominal-type-specific
    lowering branch.
  - The dedicated `Array` collect helper remains, but only as a fallback
    behind the shared generic path and only for the built-in `Array`
    language/kernel exception.
- Conclusion:
  - the compiler now has an explicit proof that `collect<C>()` is shared
    generic nominal lowering first, not a pattern for adding more
    per-structure fast paths
  - the next category remains broader performance widening on the next hot
    generic-container/runtime edges, not more nominal special casing
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteIteratorCollectGenericNominalAccumulator(StaysNative|Executes)|ConcreteIteratorGenericMethods(StayNative|Execute|StayNativeWithExperimentalMonoArrays|ExecuteWithExperimentalMonoArrays)|ConcreteIteratorMapFilterFunction(StaysNative|Executes))$' -count=1` (pass, `13.585s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteIteratorCollectGenericNominalAccumulator(StaysNative|Executes)|ConcreteIteratorGenericMethods(StayNativeWithExperimentalMonoArrays|ExecuteWithExperimentalMonoArrays)|ConcreteIteratorMapFilterFunction(StaysNative|Executes))$|TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `13.425s`).
  - `git diff --check` (pass).

# 2026-03-20 — Native iterator `filter_map` / iterator-controller tranche (v12)
- Closed the remaining iterator-literal controller/runtime-value edge inside
  the generic iterator default-method family.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_iterators.go`
  - `v12/interpreters/go/pkg/compiler/generator_iterator_controller_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
  - `v12/interpreters/go/pkg/compiler/ir_codegen_iterators.go`
  - `v12/interpreters/go/pkg/compiler/compiler_iterator_filter_map_native_test.go`
  - `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`
  - `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/manifest.json`
  - `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
- Tranche details:
  - Compiled iterator literals now bind `gen` as a compiler-owned
    `*__able_generator` controller instead of an opaque runtime object.
  - `gen.yield(...)`, `gen.stop()`, and bound `gen.yield` callable captures
    now lower directly through that compiler-owned controller path instead of
    `__able_method_call_node(...)`.
  - Native nilable/static-carrier conditions now lower to direct nil checks
    (`expr != nil`) instead of broad `__able_truthy(...)` conversion, which
    keeps `Iterator.filter_map` on the static path.
  - The widened validation slice confirmed this also closes the fallback that
    had reopened in `LazySeq.iterator` via `self.each(gen.yield)`.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
- Compiled-only 3-run average (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_iterator_filter_map_i64_small`: `0.1267s`, `10.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`:
    `191952000`.
- Conclusion:
  - the iterator default-method family is now closed through `filter_map`
  - the next category is the next benchmark-worthy generic-container/runtime
    edge beyond iterator-literal controller cleanup
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteIteratorFilterMap(StayNative|Executes)(WithExperimentalMonoArrays)?$' -count=1` (pass, `6.875s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteIterator(GenericMethodsStayNative|GenericMethodsExecute|GenericMethodsStayNativeWithExperimentalMonoArrays|GenericMethodsExecuteWithExperimentalMonoArrays|MapFilterFunctionStaysNative|MapFilterFunctionExecutes)|ConcreteIteratorFilterMap(StayNative|Executes)(WithExperimentalMonoArrays)?|LazySeqIteratorCarrierStaysNative)$|TestCompilerExecFixtures/(06_12_14_stdlib_collections_linked_list_lazy_seq|06_12_18_stdlib_collections_array_range)$|TestIREmitFunctionIteratorLiteral$' -count=1` (pass, `24.641s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able` (pass, output `191952000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able` (pass, `0.1267s`, `10.00` GC).

# 2026-03-20 — Mono-array-enabled `Iterator.collect<Array T>()` tranche (v12)
- Closed the follow-up iterator/default-method slice that remained open after
  the earlier `map/filter -> next()` tranche.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_specialized_impl_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_collect_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_render.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`
  - `v12/fixtures/bench/linked_list_iterator_collect_i64_small/manifest.json`
  - `v12/fixtures/bench/linked_list_iterator_collect_i64_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
- Tranche details:
  - Fixed the sibling impl specialization binding bug so default-impl sibling
    specialization no longer reuses the caller's generic-name map blindly.
  - Fixed the specialization matcher recursion bug by normalizing once at the
    top-level instead of renormalizing on every recursive descent.
  - Closed the actual mono-array collect issue by lowering
    `Iterator.collect<Array i64>()` through a generated compiled helper with a
    specialized `*__able_array_i64` accumulator instead of the residual
    `__able_method_call_node(...)` + `__able_array_i64_from(...)` bridge.
  - This helper is intentionally compiler-owned and array-specific because
    `Array` is a language/kernel special form; it is not a precedent for
    adding bespoke lowering rules for arbitrary user-defined nominal types.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_iterator_collect_i64_small`: mono on `0.1833s`,
    `14.00` GC; mono off `0.1833s`, `13.33` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`:
    `382455000`.
- Conclusion:
  - the mono-array-enabled `Iterator.collect<Array T>()` correctness bug is
    closed on the reduced iterator-pipeline family
  - the result is performance-neutral on this reduced benchmark, so this is a
    correctness/native-carrier closure rather than a new speed step
  - the next category is broader performance widening on the next hot generic
    container/runtime edges
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteIteratorGenericMethodsStayNative|ConcreteIteratorGenericMethodsExecute|ConcreteIteratorGenericMethodsStayNativeWithExperimentalMonoArrays|ConcreteIteratorGenericMethodsExecuteWithExperimentalMonoArrays)$' -count=1` (pass, `6.948s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteEnumerableGenericMethodsStayNative|LinkedListIterableAdapterStaysNative|LazySeqIteratorCarrierStaysNative|ConcreteIteratorGenericMethodsStayNative|ConcreteIteratorGenericMethodsExecute|ConcreteIteratorGenericMethodsStayNativeWithExperimentalMonoArrays|ConcreteIteratorGenericMethodsExecuteWithExperimentalMonoArrays|ConcreteIteratorMapFilterFunctionStaysNative|ConcreteIteratorMapFilterFunctionExecutes)$|TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `16.994s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output /home/david/sync/projects/able/v12/tmp/iter-collect.SAtVIg/main.able` (pass, output `15`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` (pass, output `382455000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` (pass, `0.1833s`, `14.00` GC).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` (pass, `0.1833s`, `13.33` GC).

# 2026-03-20 — Native iterator default-method hot-path tranche (`LinkedList.lazy().map/filter -> next()`) (v12)
- Closed the next benchmark-worthy generic-container/runtime edge on the
  shared native carrier path.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_calls.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`
  - `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/manifest.json`
  - `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
- Tranche details:
  - Ordinary default native-interface methods now lower through the same
    direct compiled-helper path already used for default generic methods when
    the receiver stays on a native interface carrier.
  - On the representative iterator pipeline shape, compiled `Iterator.filter`
    no longer routes through the runtime adapter method layer after
    `LinkedList.lazy().map<i64>(...)`; it now resolves to the compiled default
    helper directly on the native iterator carrier.
  - New regressions pin both the direct callsite shape and a function-shaped
    `LinkedList.lazy().map/filter -> next()` loop body so the path fails if it
    reintroduces `__able_iface_Iterator_*_to_runtime_value(...)`,
    `__able_method_call_node(...)`, or `__able_call_value(...)`.
  - The first attempted reduced benchmark used
    `...collect<Array i64>().reduce(...)`, but that exposed a separate open
    issue on the mono-array-enabled CLI/default path: `Iterator.collect<Array
    T>()` still falls back through a residual dynamic/specialized-array bridge.
    This tranche intentionally closed the already-corrected `map/filter/next`
    edge without folding that distinct bug into the same benchmark.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_iterator_pipeline_i64_small`: `0.1800s`, `13.33` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`:
    `382455000`.
- Conclusion:
  - the `LinkedList.lazy().map/filter -> next()` hot path is now closed on the
    shared native iterator carrier design
  - the next category is narrower than “another generic container pass”: it is
    the mono-array-enabled `Iterator.collect<Array T>()` / specialized-array
    accumulator interaction surfaced by the first benchmark attempt
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|ConcreteIteratorGenericMethodsStayNative|ConcreteIteratorGenericMethodsExecute|ConcreteIteratorMapFilterFunctionStaysNative|ConcreteIteratorMapFilterFunctionExecutes|ConcreteEnumerableGenericMethodsStayNative|LazySeqIteratorCarrierStaysNative|LinkedListIterableAdapterStaysNative)$|TestCompilerExecFixtures/(06_12_14_stdlib_collections_linked_list_lazy_seq|06_12_16_stdlib_collections_deque_queue|06_12_17_stdlib_collections_bit_set_heap)$' -count=1` (pass, `20.237s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able` (pass, output `382455000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able` (pass, `0.1800s`, `13.33` GC).
  - `git diff --check` (pass).

# 2026-03-20 — Callback/runtime carrier cleanup inside generic default impls (v12)
- Closed the remaining callback/runtime-value carrier slice on the concrete
  `LinkedList` `Enumerable` default-impl hot path.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/model.go`
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_compile_context.go`
  - `v12/interpreters/go/pkg/compiler/generator_overloads.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_specialized_impl_calls.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-specialized-default-impls-compiled.md`
- Tranche details:
  - Specialized impl functions now retain bound generic type bindings through
    compileability checks and render, and the renderer now iterates to a fixed
    point so specialized siblings discovered during body compilation are
    emitted in the same pass.
  - Specialized sibling impls are cached early enough to break mutually
    recursive specialization loops during codegen.
  - Default-impl call selection now prefers specialized sibling impls before
    the ordinary concrete-receiver method path, which fixes the remaining
    `Enumerable.lazy()` regression: the specialized `LinkedList` helper now
    calls `__able_compiled_impl_Enumerable_iterator_1_spec(...)` directly
    instead of bridging `Iterator_A -> runtime.Value -> Iterator_i32`.
  - The linked-list benchmark no longer stack-overflows by recursively
    converting `LinkedListIterator -> ListNode` cycles through
    `__able_any_to_value(...)`.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-specialized-default-impls-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_enumerable_i32_small`: `0.1667s`, `15.33` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`:
    `382455000`.
- Conclusion:
  - the callback/runtime-value carrier regression on this hot path is closed
  - wall-clock is back in line with the previous linked-list baseline; this
    tranche is a correctness/native-carrier cleanup, not a new speed step
  - the next category is the next benchmark-worthy callback/generic-container
    runtime edge beyond this `LinkedList` default-impl path
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteEnumerableGenericMethodsStayNative|LinkedListIterableAdapterStaysNative|LazySeqIteratorCarrierStaysNative)$|TestCompilerExecFixtures/(06_12_14_stdlib_collections_linked_list_lazy_seq|06_12_16_stdlib_collections_deque_queue|06_12_17_stdlib_collections_bit_set_heap)$' -count=1` (pass, `11.957s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass, output `382455000`, no stack overflow).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass, `0.1667s`, `15.33` GC).
  - `git diff --check` (pass).

# 2026-03-20 — Concrete `Enumerable` default-method hot-path tranche (`LinkedList.map/filter/reduce`) (v12)
- Closed the next concrete generic/default container-method hot-path tranche
  on the shared native carrier path.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_impls.go`
  - `v12/interpreters/go/pkg/compiler/generator_compile_context.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_iterables.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_methods.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
  - `v12/fixtures/bench/linked_list_enumerable_i32_small/manifest.json`
  - `v12/fixtures/bench/linked_list_enumerable_i32_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-compiled.md`
- Tranche details:
  - The compiler now binds higher-kinded interface self patterns like
    `Enumerable A for C _` to the concrete target type on compiled impl paths,
    so concrete `LinkedList.map/filter/reduce` calls resolve to compiled impl
    functions instead of the narrow runtime generic-method boundary.
  - Bound type-constructor calls inside those compiled default impl bodies,
    such as `C.default()`, now resolve through static compiled impl lookup
    rather than `__able_env_get("C")`.
  - Native `Iterator<T>` carriers now satisfy compiled iterable lowering
    directly, which removes the `to_runtime_value -> from_value -> iterator()`
    round-trip that previously overflowed on larger `LinkedList` graphs inside
    default `Enumerable` loops.
  - New focused regressions now pin both the concrete `Enumerable`
    callsite shape and the compiled `Enumerable.map` loop body so those
    runtime iterator round-trips do not return.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_enumerable_i32_small`: `0.1667s`, `12.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`:
    `382455000`.
- Conclusion:
  - the next remaining hot edge on this family is callback/runtime-value
    carrier overhead inside generic default impl bodies, not container target
    resolution or iterator-fallback correctness
  - the next category is the next benchmark-worthy callback/generic-container
    runtime edge rather than another nominal container-correctness pass
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerConcreteEnumerableGenericMethodsStayNative$|TestCompilerConcreteIterableForLoopStaysNative$|TestCompilerInterfaceIterableForLoopStaysNative$' -count=1` (pass, `0.733s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ConcreteEnumerableGenericMethodsStayNative|LinkedListIterableAdapterStaysNative|LazySeqIteratorCarrierStaysNative|ConcreteIterableArgToInterfaceParamStaysNative|ConcreteIterableForLoopStaysNative|InterfaceIterableForLoopStaysNative|DequeQueueMethodsStayNative|BitSetHeapMethodsStayNative|PersistentSortedQueueMethodsStayNative|StaticSliceBuiltinsIgnoreLenShadowing|TreeMapStaticCarrierStaysNative|PersistentMapStaticCarrierStaysNative|HashMapStaticCarrierStaysNative)$|TestCompilerExecFixtures/(06_12_11_stdlib_collections_tree_map_set|06_12_12_stdlib_collections_persistent_map_set|06_12_13_stdlib_collections_persistent_sorted_queue|06_12_14_stdlib_collections_linked_list_lazy_seq|06_12_16_stdlib_collections_deque_queue|06_12_17_stdlib_collections_bit_set_heap|14_01_language_interfaces_index_apply_iterable)$' -count=1` (pass, `27.048s`).
  - direct compiled parity check for `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass, output `382455000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` (pass).

# 2026-03-20 — Benchmark-worthy generic container hot-path tranche (`LinkedList -> Iterable -> Iterator`) (v12)
- Closed the next generic-container hot-path tranche on the shared native
  carrier path.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_iterables.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_ident.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_inheritance.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/compiler_iterable_loop_native_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
  - `v12/fixtures/bench/linked_list_for_i32_small/main.able`
  - `v12/fixtures/bench/linked_list_for_i32_small/manifest.json`
  - `v12/fixtures/bench/linked_list_for_i32_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-20-linked-list-for-i32-small-compiled.md`
- Tranche details:
  - Static `for value in iterable` lowering now uses native concrete/interface
    receiver calls (`iterator`, `next`) instead of the old
    `__able_resolve_iterator(...)` runtime path when the iterable carrier is
    already statically representable.
  - Identifier coercion now prefers static expected-type coercion before
    broad runtime conversion, which keeps more native interface/callable
    argument paths on compiler-owned carriers.
  - Native interface adapter synthesis now honors interface inheritance, so
    containers that implement a derived interface (for example
    `Enumerable A for LinkedList`) now synthesize the corresponding native
    base-interface adapter (`Iterable A`) instead of falling back to
    `runtime.Value`.
  - Native interface concrete adapters now directly coerce compatible native
    interface return carriers instead of round-tripping through runtime values.
    That removes the recursive conversion bug on cyclic native container
    graphs like `LinkedListIterator -> ListNode`.
  - New focused regressions now pin both the static iterable loop lowering and
    the `LinkedList` native `Iterable` adapter path.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-20-linked-list-for-i32-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/linked_list_for_i32_small`: `0.2000s`, `15.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/linked_list_for_i32_small/main.able`: `1199940000`.
- Conclusion:
  - the first benchmark-worthy generic-container hot path is now closed on the
    shared native interface/container carrier design;
  - the next category is broader performance widening for the remaining
    generic container/runtime carrier edges that are still hot enough to
    matter.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(LinkedListIterableAdapterStaysNative|LazySeqIteratorCarrierStaysNative|ConcreteIterableArgToInterfaceParamStaysNative|ConcreteIterableForLoopStaysNative|InterfaceIterableForLoopStaysNative)$|TestCompilerExecFixtures/(14_01_language_interfaces_index_apply_iterable|06_12_14_stdlib_collections_linked_list_lazy_seq)$' -count=1` (pass, `6.253s`).
  - direct compiled parity check for `v12/fixtures/bench/linked_list_for_i32_small/main.able` (pass, output `1199940000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/linked_list_for_i32_small/main.able` (pass).

# 2026-03-20 — Deeper generic container native-carrier fix (`LazySeq` / typed `nil`) (v12)
- Closed the next deeper generic-container correctness tranche without adding
  a new container-specific lowering rule.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_generic_native_test.go`
- Tranche details:
  - The real remaining issue on this slice was shared native-carrier hygiene,
    not another stdlib container exception.
  - `mapNullableType(...)` now preserves native nilable carriers when the
    inner Go type already has a nil zero value, which keeps generic container
    fields like `LazySeq.Source: ?(Iterator T)` on the generated native
    interface carrier instead of collapsing to `any`.
  - `nil` expression lowering now emits typed Go nils for native nilable
    carriers (`(*ListNode)(nil)`, `__able_iface_Iterator_T(nil)`, etc.) rather
    than invalid untyped `nil` short declarations in compiled bodies.
  - This closes the compiled stdlib fixture regression in
    `06_12_14_stdlib_collections_linked_list_lazy_seq` and keeps the
    linked-list/lazy-seq path on the shared native container/interface ABI.
- Conclusion:
  - the next container/native-carrier category is no longer “fix deeper
    generic container correctness”; it is benchmark-worthy generic container
    hot paths and any remaining residual container/runtime fallback surfaces.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerLazySeqIteratorCarrierStaysNative$|TestCompilerExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1` (pass, `3.064s`).
  - `git diff --check` (pass).

# 2026-03-19 — Broader stdlib container audit + `Heap i32` benchmark tranche (v12)
- Closed the next broader container tranche without adding new type-specific
  lowering rules.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/compiler_container_array_family_native_test.go`
  - `v12/fixtures/bench/heap_i32_small/main.able`
  - `v12/fixtures/bench/heap_i32_small/manifest.json`
  - `v12/fixtures/bench/heap_i32_small/package.yml`
  - `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
- Tranche details:
  - The remaining stdlib container families in this slice already lower
    through the shared nominal/container path; the missing work was mechanical
    audit coverage, not more compiler exceptions.
  - New no-fallback compiler regressions now pin representative static method
    bodies for:
    - `Deque` / `Queue`
    - `BitSet` / `Heap`
    - `PersistentSortedSet` / `PersistentQueue`
  - Those tests assert native locals stay on compiler-native carriers and that
    representative compiled methods avoid `__able_call_value(...)`,
    `__able_member_get_method(...)`, `__able_method_call_node(...)`,
    `bridge.MatchType(...)`, and `__able_try_cast(...)`.
  - Shared compiled fixture gates are green for the same family:
    `06_12_13_stdlib_collections_persistent_sorted_queue`,
    `06_12_16_stdlib_collections_deque_queue`, and
    `06_12_17_stdlib_collections_bit_set_heap`.
  - Added a checked-in reduced benchmark target for `Heap i32`:
    `v12/fixtures/bench/heap_i32_small/main.able`.
  - The first version of that target was too large for a “small” checked-in
    fixture (`53.0167s` average compiled run), so it was trimmed before being
    recorded.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
- Final compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build
  path):
  - `bench/heap_i32_small`: `7.7533s`, `1105.00` GC.
- Direct parity:
  - direct compiled `ablec` output for
    `v12/fixtures/bench/heap_i32_small/main.able`: `-211812354`.
- Conclusion:
  - this broader array-backed/persistent container family is now mechanically
    covered by the shared native-lowering contract;
  - the next category is no longer “make these families native”, it is deeper
    generic container paths and any remaining container/runtime surfaces that
    still force generic carrier fallbacks.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DequeQueueMethodsStayNative|BitSetHeapMethodsStayNative|PersistentSortedQueueMethodsStayNative|StaticSliceBuiltinsIgnoreLenShadowing|TreeMapStaticCarrierStaysNative|PersistentMapStaticCarrierStaysNative)$|TestCompilerExecFixtures/(06_12_13_stdlib_collections_persistent_sorted_queue|06_12_16_stdlib_collections_deque_queue|06_12_17_stdlib_collections_bit_set_heap|06_12_11_stdlib_collections_tree_map_set|06_12_12_stdlib_collections_persistent_map_set)$' -count=1` (pass, `19.583s`).
  - direct compiled parity check for `v12/fixtures/bench/heap_i32_small/main.able` (pass, output `-211812354`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/heap_i32_small/main.able` (pass).
  - `git diff --check` (pass).

# 2026-03-19 — Nominal tree/persistent container follow-through (v12)
- Closed the next nominal container tranche without adding new type-specific
  lowering rules for tree/persistent map/set families.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_static_builtin_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_builtins.go`
  - `v12/interpreters/go/pkg/compiler/generator_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections_static_array_access.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_intrinsics.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/compiler_container_nominal_native_test.go`
- Tranche details:
  - The real blocker on `TreeMap` / `TreeSet` and `PersistentMap` /
    `PersistentSet` was generic codegen hygiene, not missing per-container
    lowering. Static compiled bodies were still emitting raw Go `len(...)` and
    `cap(...)`, which broke when Able code bound locals named `len`.
  - Static slice/string builtin use now routes through generated helpers
    (`__able_slice_len`, `__able_slice_cap`, `__able_string_len_bytes`) in
    compiled bodies, so nominal container code no longer depends on unshadowed
    Go predeclared identifiers.
  - Focused regressions now pin the builtin-shadowing case directly and assert
    that `TreeMap` / `PersistentMap` params and returns stay on native
    `*TreeMap` / `*PersistentMap` carriers rather than regressing to
    `runtime.Value`.
  - This closes the broader tree/persistent map/set follow-through via the
    shared nominal struct/interface pipeline instead of adding new container
    exceptions.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StaticSliceBuiltinsIgnoreLenShadowing|TreeMapStaticCarrierStaysNative|PersistentMapStaticCarrierStaysNative|HashMapStaticCarrierStaysNative)$|TestCompilerCompiledHashSetUnionStdlib$|TestCompilerExecFixtures/(06_12_11_stdlib_collections_tree_map_set|06_12_12_stdlib_collections_persistent_map_set)$' -count=1` (pass, `12.323s`).
  - `cd v12/interpreters/go && git diff --check` (pass).

# 2026-03-19 — Generic nominal container carrier follow-through (v12)
- Closed the remaining integration gaps exposed by the shared nominal-carrier
  refactor instead of backing out the generic path.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_control.go`
  - `v12/interpreters/go/pkg/compiler/compiler_control_bridge_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_type_alias_native_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_hashmap_native_test.go`
- Tranche details:
  - `TypeMapper` now expands simple type aliases before host-type mapping, so
    alias-backed nominal struct fields lower through the same generic carrier
    path as any other struct field.
  - That fixes the kernel/container case cleanly: `HashMap.handle` now lowers
    as `int64` via `HashMapHandle = i64` instead of silently becoming
    `runtime.Value`.
  - The explicit map-literal lowering edge now converts the runtime hash-map
    handle value back into the native `int64` carrier before constructing the
    compiled `HashMap` struct.
  - The compiled control bridge now preserves exit signals via
    `interpreter.ExitCodeFromError(control.Err)` before wrapping raised values,
    which fixes the false `runtime: runtime error` failure in the compiled
    stdlib `HashSet.union` integration path.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ControlToErrorPreservesExitSignals|TypeAliasStructFieldStaysNative|HashMapStaticCarrierStaysNative|HashMapLiteralStaysNative|HashMapCarrierArrayStaysSpecialized|GenericMapInterfaceSignatureStaysNative|HashMapNativeCarrierExecutes|HashSetStaticCarrierStaysNative|HashSetCarrierArrayStaysSpecialized|HashSetIteratorWrapsConcreteNativeIterator|HashSetIteratorNativeCarrierExecutes)$|TestCompilerDynamicBoundaryCallbackHashMapConversion(Success|Failure)Markers$|TestCompilerCompiledHashSetUnionStdlib$' -count=1` (pass, `18.066s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_01_compiler_map_literal|06_01_compiler_map_literal_typed|06_01_compiler_map_literal_spread|06_12_15_stdlib_collections_hash_map_set)$|TestCompilerExecHarness$' -count=1` (pass).
  - `git diff --check` (pass).

# 2026-03-19 — Native HashMap container carrier tranche (v12)
- Closed the next broader non-scalar/container lowering slice by keeping
  `HashMap K V` on native compiler carriers and fixing the residual
  generic-interface return matching exposed by stdlib `HashSet.iterator()`.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_coercions.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_methods.go`
  - `v12/interpreters/go/pkg/compiler/compiler_hashmap_native_test.go`
  - `v12/fixtures/bench/hashmap_i32_small/main.able`
  - `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
- Tranche details:
  - `HashMap K V` now maps to native `*HashMap` carriers on static compiled
    paths instead of collapsing to `any`;
  - typed and inferred map literals now build native `*HashMap` values in the
    compiled body, and `Array (HashMap K V)` outer carriers stay native too;
  - `Map K V` interface params stay on the generated native interface carrier;
  - generic interface return matching no longer pre-binds interface generics
    to themselves, which closes the old `HashSet.iterator()` /
    `Enumerable.iterator` fallback on shapes like `Iterator T` vs
    `Iterator A`;
  - the residual native-interface coercion path is now an explicit runtime
    roundtrip at the ABI edge instead of compiler fallback;
  - `generator_native_interfaces.go` is back under the file-size limit after
    splitting coercion/matching helpers into
    `generator_native_interface_coercions.go`.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/hashmap_i32_small`: `1.7633s`, `175.33` GC.
- Conclusion:
  - the first broader native container slice beyond arrays is now landed;
  - this closes the remaining correctness/native-lowering gaps around static
    `HashMap` carriers and generic `Iterator` interface returns, and it gives
    us a checked-in reduced benchmark target for future map/set performance
    work.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(HashMapStaticCarrierStaysNative|HashMapLiteralStaysNative|HashMapCarrierArrayStaysSpecialized|GenericMapInterfaceSignatureStaysNative|HashMapNativeCarrierExecutes|HashSetIteratorWrapsConcreteNativeIterator|HashSetIteratorNativeCarrierExecutes)$|TestCompilerDynamicBoundaryCallbackHashMapConversion(Success|Failure)Markers$' -count=1` (pass, `14.694s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_01_compiler_map_literal|06_01_compiler_map_literal_typed|06_01_compiler_map_literal_spread|06_12_15_stdlib_collections_hash_map_set)$' -count=1` (pass, `12.991s`).
  - direct `ablec` parity check for `v12/fixtures/bench/hashmap_i32_small/main.able` (pass, output `4498503`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/hashmap_i32_small/main.able` (pass).
  - `git diff --check` (pass).

# 2026-03-19 — Remaining primitive numeric mono-array tranche (v12)
- Closed the next widening slice by covering the remaining primitive numeric
  scalar family with staged compiler-owned array wrappers.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_numeric_family_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_mono_arrays_numeric_test.go`
  - `v12/fixtures/bench/sum_u32_small/main.able`
  - `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
- Tranche details:
  - staged mono-array wrappers now also cover `Array i8`, `Array i16`,
    `Array u16`, `Array u32`, `Array u64`, `Array isize`, `Array usize`, and
    `Array f32`;
  - focused compile-time regressions now pin the generated wrapper types for
    those families and a runtime execution slice exercises all of them in a
    single compiled program;
  - dynamic-boundary callback coverage now explicitly includes `Array u32` and
    `Array f32`.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/sum_u32_small`: mono on `1.0933s`, `185.33` GC; mono off
    `1.6800s`, `21.33` GC.
- Conclusion:
  - the staged primitive numeric scalar family is now materially more complete;
  - the checked-in `u32` benchmark shows a real wall-clock win on the
    specialized path even though raw GC count is not lower on this workload,
    which points to boxing/runtime dispatch as the main generic-path cost here.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArraysRemainingNumericFamilyUsesSpecializedWrappers|ExperimentalMonoArraysRemainingNumericFamilyExecutes|ExperimentalMonoArrays(TypedArrayUsesSpecializedWrapper|F64TypedArrayUsesSpecializedWrapper|CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper|TextTypedArraysExecute|F64Executes|TypedArrayExecutes))$|TestCompilerDynamicBoundaryMonoArray(U32CallbackConversionSuccessMarkers|F32CallbackConversionSuccessMarkers)$' -count=1` (pass, `8.351s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|CarrierArrayWrappersRemainAvailableWithoutExperimentalMonoArrays|CarrierArrayWrappersPreserveNestedRowIdentityWithoutExperimentalMonoArrays|ExperimentalMonoArrays(RemainingNumericFamilyUsesSpecializedWrappers|RemainingNumericFamilyExecutes|TypedArrayUsesSpecializedWrapper|F64TypedArrayUsesSpecializedWrapper|CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper|TypedArrayWrapperUsesSpecializedBoundaryConverters|TypedArrayExecutes|F64Executes|TextTypedArraysExecute|CharResultPropagationStaysNative|CharResultPropagationExecutes|StaticBodyStaysOnCompilerOwnedArrayCarrier|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized|WidenedSliceExecutes|PropagationComputedIndexStaysHoisted|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|NestedCharRowsStaySpecialized|InterfaceCarrierArrayStaysSpecialized|CallableCarrierArrayStaysSpecialized|CarrierArraysExecute)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerDynamicBoundaryMonoArray(U32CallbackConversionSuccessMarkers|F32CallbackConversionSuccessMarkers|CharCallbackConversionSuccessMarkers|StringCallbackConversionSuccessMarkers|F64CallbackConversionSuccessMarkers|CallbackConversionSuccessMarkers)' -count=1` (pass, `20.911s`).
  - direct `ablec` output parity check for `v12/fixtures/bench/sum_u32_small/main.able` with and without `--no-experimental-mono-arrays` (pass, both `17999997000003`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/sum_u32_small/main.able` (pass).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/sum_u32_small/main.able` (pass).

# 2026-03-19 — Mono-off carrier-array identity correction tranche (v12)
- Closed the next correctness/perf tranche on the widened mono-array work by
  keeping compiler-owned carrier-array wrappers available even when staged
  scalar mono arrays are disabled.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- Tranche details:
  - staged scalar specializations remain gated by `ExperimentalMonoArrays`;
  - carrier-array wrappers for already-native compiler carriers now synthesize
    regardless of that flag, including outer arrays of `*Array`,
    native interfaces, native callables, native unions, and other pointer-backed
    compiler-owned carriers;
  - this fixes the mono-off nested-text-row bug where `Array (Array char)`
    fell back to generic `*Array` / `[]runtime.Value` outer storage, boxed rows
    through `runtime.Value`, and then lost row mutation identity on
    `rows[idx]!.push(...)`;
  - the corrected mono-off `zigzag_char_small` compiled binary now prints the
    same result as mono-on (`8192000`), so the prior mono-on/mono-off timing
    comparison was invalid and has been replaced.
- Snapshot corrected in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- Corrected compiled-only 3-run averages (`v12/bench_perf`, direct `ablec`
  build path):
  - `bench/zigzag_char_small`: mono on `0.9567s`, `88.00` GC; mono off
    `1.0500s`, `384.00` GC.
- Conclusion:
  - the text-scalar widening slice is not regressing on the checked-in reduced
    benchmark once the mono-off path is corrected;
  - staged `Array char` specialization plus native outer carrier identity now
    show a modest wall-clock win and a large GC win on this workload;
  - the next mono-array category is broader carrier reduction/widening again,
    not text-path rollback.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(CarrierArrayWrappersRemainAvailableWithoutExperimentalMonoArrays|CarrierArrayWrappersPreserveNestedRowIdentityWithoutExperimentalMonoArrays|ExperimentalMonoArraysNestedCharRowsStaySpecialized|ExperimentalMonoArraysCharResultPropagation(StaysNative|Executes)|ExperimentalMonoArrays(TextTypedArraysExecute|CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper))$|TestCompilerDynamicBoundaryMonoArray(CharCallbackConversionSuccessMarkers|StringCallbackConversionSuccessMarkers)$' -count=1` (pass, `5.722s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|CarrierArrayWrappersRemainAvailableWithoutExperimentalMonoArrays|CarrierArrayWrappersPreserveNestedRowIdentityWithoutExperimentalMonoArrays|ExperimentalMonoArrays(TypedArrayUsesSpecializedWrapper|F64TypedArrayUsesSpecializedWrapper|CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper|TypedArrayWrapperUsesSpecializedBoundaryConverters|TypedArrayExecutes|F64Executes|TextTypedArraysExecute|CharResultPropagationStaysNative|CharResultPropagationExecutes|StaticBodyStaysOnCompilerOwnedArrayCarrier|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized|WidenedSliceExecutes|PropagationComputedIndexStaysHoisted|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|NestedCharRowsStaySpecialized|InterfaceCarrierArrayStaysSpecialized|CallableCarrierArrayStaysSpecialized|CarrierArraysExecute)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerDynamicBoundaryMonoArray' -count=1` (pass, `31.130s`).
  - direct `ablec` output parity check for `v12/fixtures/bench/zigzag_char_small/main.able` with and without `--no-experimental-mono-arrays` (pass, both `8192000`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/zigzag_char_small/main.able` (pass).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/zigzag_char_small/main.able` (pass).

# 2026-03-19 — Compiler text-scalar mono-array tranche (v12)
- Closed the next broader scalar-carrier slice by extending staged
  compiler-owned mono arrays to the text scalar family and fixing the result
  propagation bug that blocked `!Array char` on static compiled paths.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_specialized_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_mono_arrays_test.go`
  - `v12/fixtures/bench/zigzag_char_small/main.able`
  - `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- Tranche details:
  - `Array char` now lowers to `*__able_array_char` over `[]rune`;
  - `Array String` now lowers to `*__able_array_String` over `[]string`;
  - nested `Array (Array char)` now stays on a dedicated compiler-owned outer
    wrapper (`*__able_array_array_char`) instead of falling back to the generic
    `*Array` shell;
  - explicit dynamic-boundary callback coverage now includes specialized
    `Array char` and `Array String` payloads;
  - `compilePropagationExpression(...)` now re-wraps native result success
    branches through the static coercion path, fixing the emitted
    `!Array char` / `!Array String` bug where the compiler incorrectly called
    `_from_value(__able_runtime, specializedCarrier)` on an already-native
    success value.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- Initial mono-off comparison recorded here was later superseded by the
  mono-off carrier-array identity correction tranche above. That earlier
  baseline was invalid because the mono-off compiled path produced the wrong
  result on nested text rows.
- Conclusion:
  - the text-scalar carrier widening is architecturally complete for this
    slice, and `!Array char` now compiles/executes correctly on static paths
  - the benchmark interpretation for this slice now comes from the corrected
    mono-off identity tranche above, not the initial invalid mono-off run.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExperimentalMonoArrays(CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper|TextTypedArraysExecute|CharResultPropagationStaysNative|CharResultPropagationExecutes|NestedCharRowsStaySpecialized)$|TestCompilerDynamicBoundaryMonoArray(CharCallbackConversionSuccessMarkers|StringCallbackConversionSuccessMarkers)$' -count=1` (pass, `6.162s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArrays(TypedArrayUsesSpecializedWrapper|F64TypedArrayUsesSpecializedWrapper|CharTypedArrayUsesSpecializedWrapper|StringTypedArrayUsesSpecializedWrapper|TypedArrayWrapperUsesSpecializedBoundaryConverters|TypedArrayExecutes|F64Executes|TextTypedArraysExecute|CharResultPropagationStaysNative|CharResultPropagationExecutes|StaticBodyStaysOnCompilerOwnedArrayCarrier|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized|WidenedSliceExecutes|PropagationComputedIndexStaysHoisted|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|NestedCharRowsStaySpecialized|InterfaceCarrierArrayStaysSpecialized|CallableCarrierArrayStaysSpecialized|CarrierArraysExecute)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerDynamicBoundaryMonoArray' -count=1` (pass, `33.355s`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/zigzag_char_small/main.able` (pass).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/zigzag_char_small/main.able` (pass).

# 2026-03-19 — Compiler native carrier-array widening tranche (v12)
- Closed the next broader carrier-reduction slice by extending compiler-owned
  array wrapper synthesis beyond staged scalars / nested mono arrays to other
  native carrier element families the compiler already owns.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_mono_arrays_test.go`
- Tranche details:
  - mono-array wrapper synthesis is no longer limited to staged scalar element
    kinds plus nested mono-array wrappers;
  - arrays whose element type is already a compiler-owned native carrier now
    synthesize dedicated outer wrappers too, including:
    - generic inner arrays (`Array (Array char)` -> `*__able_array_Array`)
    - native interface carriers (`Array Greeter` ->
      `*__able_array_iface_Greeter`)
    - native callable carriers (`Array (i32 -> i32)` ->
      `*__able_array_fn_int32_to_int32`)
    - other representable pointer-backed carrier families covered by the same
      spec synthesis path;
  - rendered mono-array converters now reuse the compiler's existing native
    value conversion rules for these carrier elements instead of only the
    scalar/nested-mono special cases, and explicit dynamic-boundary callback
    coverage now includes both interface-carrier arrays and callable-carrier
    arrays.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExperimentalMonoArrays(GenericInnerArrayCarrierStaysSpecialized|InterfaceCarrierArrayStaysSpecialized|CallableCarrierArrayStaysSpecialized|CarrierArraysExecute|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes)$|TestCompilerDynamicBoundaryMonoArray(InterfaceCarrierCallbackConversionSuccessMarkers|CallableCarrierCallbackConversionSuccessMarkers|F64CallbackConversionSuccessMarkers|CallbackConversionSuccessMarkers)$' -count=1` (pass, `10.416s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InterfaceParamAndReturnStayNative|TypedInterfaceAssignmentStaysNative|PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative|ExperimentalMonoArrays(GenericInnerArrayCarrierStaysSpecialized|InterfaceCarrierArrayStaysSpecialized|CallableCarrierArrayStaysSpecialized|CarrierArraysExecute|NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes)|FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerDynamicBoundaryMonoArray(InterfaceCarrierCallbackConversionSuccessMarkers|CallableCarrierCallbackConversionSuccessMarkers|F64CallbackConversionSuccessMarkers|CallbackConversionSuccessMarkers)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass, `16.838s`).
  - `git diff --check` (pass).
- Benchmark note:
  - no new benchmark snapshot was added in this tranche because the current
    shared benchmark set does not materially exercise interface/callable or
    generic-inner-array carrier arrays; this slice is an architectural
    widening pass rather than a benchmark-driven hot-loop optimization.
- Handoff:
  - the next category is now broader performance specialization again:
    widen native storage coverage to more scalar/container families that still
    fall back to generic `*Array` / `runtime.Value`, then remeasure on a
    benchmark that actually exercises those paths.

# 2026-03-19 — Compiler nested mono-array outer-carrier tranche (v12)
- Closed the next mono-array carrier-reduction slice by removing the remaining
  generic outer `*Array` shell for nested staged rows on static compiled paths.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_array_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_methods.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- Tranche details:
  - `Array (Array f64)` and the same nested shape for other staged mono-array
    rows now lower to dedicated compiler-owned outer wrappers such as
    `*__able_array_array_f64` instead of the generic `*Array` carrier with
    `[]runtime.Value` row storage;
  - array literal/factory/inference/type-mapping paths now synthesize those
    nested wrappers from the inner mono-array carrier type, so static
    compiled bodies append rows directly as typed Go pointers;
  - rendered mono-array boundary helpers and native `Array` core methods now
    handle pointer-backed specialized element carriers directly, including nil
    propagation for `read_slot` / `pop` and runtime boundary conversion via
    explicit nested wrapper helpers instead of boxing through
    `runtime.Value` element slots.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: mono on `5.7233s`, `252.00` GC; mono
    off `44.5167s`, `3550.67` GC.
- Conclusion:
  - the outer nested carrier is now native and compiler-owned, which closes
    the remaining generic outer-row shell on representative static compiled
    paths;
  - this tranche did not produce a second visible wall-clock step on the
    reduced matrix benchmark beyond the earlier `f64` slice, so the next
    material performance category is broader carrier reduction beyond nested
    mono-array wrappers.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExperimentalMonoArrays(NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|F64TypedArrayUsesSpecializedWrapper|F64Executes|WidenedSliceExecutes|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized)$|TestCompilerDynamicBoundaryMonoArray(F64CallbackConversionSuccessMarkers|CallbackConversionSuccessMarkers)$' -count=1` (pass, `7.503s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArrays(NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|F64TypedArrayUsesSpecializedWrapper|F64Executes|WidenedSliceExecutes|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized)|FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerDynamicBoundaryMonoArray(F64CallbackConversionSuccessMarkers|CallbackConversionSuccessMarkers)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass, `12.151s`).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass).

# 2026-03-19 — Compiler mono-array `f64` nested carrier tranche (v12)
- Closed the next mono-array specialization slice by extending the staged
  specialized set to `f64` and fixing nested `Array (Array f64)` get/push
  propagation on static compiled paths.
- Landed in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_specialized_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_mono_arrays_test.go`
  - `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
- Tranche details:
  - `Array f64` is now part of the staged specialized wrapper set
    (`*__able_array_f64`);
  - nested `Array (Array f64)` expressions such as
    `rows.get(j)!.get(i)!` now stay on the static compiler path when a
    concrete `float64` is required, instead of forcing the surrounding
    `push(...)` call back through `__able_method_call_node(...)`;
  - the fix was in native nullable propagation: pointer-backed nullable
    carriers like `*float64` now coerce to the requested concrete static type
    through the existing nullable runtime helper path instead of being treated
    as an unrecoverable type mismatch;
  - the full compiled `v12/examples/benchmarks/matrixmultiply.able` path no
    longer aborts early with `runtime: runtime error` when mono arrays are
    enabled; at the harness's `60s` timeout it now matches mono-off and the
    historical compiled baseline by timing out rather than failing.
- Added a reproducible reduced compiler benchmark target:
  - `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - same algorithm and nested `Array (Array f64)` structure as the full
    `matrixmultiply` example, but scaled to `300x300` so both mono-on and
    mono-off compiled runs finish within the existing `60s` `bench_perf`
    budget.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
- Compiled-only 3-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/matrixmultiply_f64_small`: mono on `5.4833s`, `280.00` GC; mono
    off `45.3133s`, `3568.67` GC.
- Conclusion:
  - staged `f64` specialization plus the nested get/push propagation fix now
    produce a material wall-clock and GC win on a benchmark that directly
    exercises the newly lowered path;
  - the next category is no longer “finish the `f64` path”; it is broader
    nested/container carrier reduction beyond the current scalar-specialized
    inner-row strategy.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExperimentalMonoArrays(NestedF64RowsStaySpecialized|NestedF64RowsExecute|NestedF64GetPushStaysSpecialized|NestedF64GetPushExecutes|F64TypedArrayUsesSpecializedWrapper|F64Executes)$|TestCompilerDynamicBoundaryMonoArrayF64CallbackConversionSuccessMarkers$' -count=1` (pass, `4.994s`).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able` (timeout, no runtime failure).
  - `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/examples/benchmarks/matrixmultiply.able` (timeout).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass).
  - `./v12/bench_perf --runs 3 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/matrixmultiply_f64_small/main.able` (pass).

# 2026-03-19 — Compiler residual generic-array narrowing tranche (v12)
- Closed the residual generic-array narrowing tranche for the current
  compiler-native mono-array arc.
- Landed the remaining generic static-array narrowing in:
  - `v12/interpreters/go/pkg/compiler/generator_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections_static_array_access.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_intrinsics.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_origin_struct_calls.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
- Tranche details:
  - inferred local `Array T` bindings now preserve recoverable element-type
    metadata on static compiled paths, which lets generic static helper
    results such as `get`, `pop`, `first`, `last`, and `read_slot` stay on
    native nullable carriers where the element type is still representable;
  - array `set` / index-assignment success semantics are back in parity across
    static and residual runtime-backed paths: success now returns `nil`, while
    failure remains `IndexError`;
  - `06_08_array_ops_mutability` and `06_12_02_stdlib_array_helpers` were
    failing because the success path still returned the assigned value, which
    broke `match { case nil | case IndexError }` handling under compiled
    native lowering; those fixtures are green again after the parity fix;
  - `06_12_18_stdlib_collections_array_range` is green again too: iterator
    interface boundary helpers now accept raw `*runtime.IteratorValue`
    carriers directly, runtime-backed iterator adapters call `iter.Next()`
    directly on that carrier, and `__able_control_from_error_with_node(...)`
    now preserves `__able_generator_stop` as generator-completion control
    instead of downgrading it into a raised runtime error.
- Regression coverage added / tightened in:
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_generic_test.go`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IteratorInterfaceBoundaryAcceptsRuntimeIteratorDirectly|ArrayHelperFixtureNullableIntrinsicsStayNative|StaticNativeFixturesExecuteWithoutExplicitBoundaries/06_08_array_ops_mutability)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass, `4.727s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(IteratorInterfaceBoundaryAcceptsRuntimeIteratorDirectly|ArrayHelperFixtureNullableIntrinsicsStayNative|BroadStaticNativeTouchpointsStayNative|GenericResidualTouchpointsStayNarrow|StaticNativeFixturesExecuteWithoutExplicitBoundaries|ExperimentalMonoArrays(InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized|WidenedSliceExecutes|PropagationComputedIndexStaysHoisted)|InterfaceLookupGenericMethodFixturesRegression)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range|10_04_interface_dispatch_defaults_generics|10_15_interface_default_generic_method)$' -count=1` (pass, `17.397s`).
  - `git diff --check` (pass).
- Handoff:
  - this residual generic-array narrowing tranche is complete;
  - the next category is now genuinely different work: broaden
    specialization/monomorphization beyond the currently representable array
    residuals and other container/runtime-carrier surfaces, then re-measure
    once those broader carrier reductions land.

# 2026-03-19 — Compiler mono-array compiled remeasurement tranche (v12)
- Closed the compiled remeasurement tranche for the widened mono-array slice.
- Tightened the benchmark helper in:
  - `v12/bench_perf`
  - compiled mode now builds through `cmd/ablec` instead of `able build`, so
    direct fixture benchmarking measures the compiler path without unrelated
    package/bootstrap failures;
  - `--compiled-build-arg` can now be repeated to forward flags such as
    `--no-experimental-mono-arrays` to the compiled build step, which makes the
    mono-on versus mono-off comparison reproducible from the checked-in helper.
- Fixed a surfaced compiler codegen bug while measuring:
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - the propagated static-array index path (`arr[count - 1]!`) was still
    routing the computed index through `compileExpr(...)`, which wrapped
    checked-arithmetic setup in an IIFE and emitted invalid Go like
    `func() int32 { return struct{}{}, *__ableControl }()`;
  - the propagation path now hoists index setup lines normally, and a new
    regression test proves computed mono-array propagation indices avoid those
    residual IIFEs.
- Snapshot recorded in:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-widened-compiled.md`
- Compiled-only 5-run averages (`v12/bench_perf`, direct `ablec` build path):
  - `bench/noop`: mono on `0.0100s`, `0.00` GC; mono off `0.0100s`, `0.00` GC.
  - `bench/sieve_count`: mono on `0.0100s`, `0.00` GC; mono off `0.0100s`,
    `0.00` GC.
  - `bench/sieve_full`: mono on `0.0200s`, `1.00` GC; mono off `0.0200s`,
    `3.00` GC.
- Conclusion:
  - the widened mono-array slice does not yet produce a visible wall-clock win
    on these three compiled fixtures;
  - it does reduce timed GC pressure on the heaviest array fixture
    (`sieve_full`), which supports continuing the compiler-native
    specialization push rather than backing it out;
  - the next category is now shrinking residual generic `*Array` /
    `runtime.Value` paths and widening specialization coverage further, because
    the staged explicit-typed slice alone is not enough to move overall
    compiled runtime materially.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExperimentalMonoArrays(PropagationComputedIndexStaysHoisted|WidenedSliceExecutes|InferredLiteralLoopAndPatternStaySpecialized|FactoryCloneAndReserveStaySpecialized)$' -count=1` (pass, `1.068s`).
  - `./v12/bench_perf --runs 1 --timeout 30 --modes compiled v12/fixtures/bench/sieve_full/main.able` (pass after codegen fix).
  - `./v12/bench_perf --runs 5 --timeout 30 --modes compiled ...` for mono on / mono off across `noop`, `sieve_count`, `sieve_full` (all pass).
  - `git diff --check` (pass).

# 2026-03-19 — Compiler widened mono-array specialization tranche (v12)
- Closed the second mono-array implementation slice: specialization is no
  longer limited to explicit typed literals and first-hop intrinsics.
- Landed widening work in:
  - `v12/interpreters/go/pkg/compiler/generator_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_static_array_factories.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_static_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match_arrays.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_intrinsics.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_array_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
- Tranche details:
  - non-empty unannotated local array literals now infer staged specialized
    carriers (`[1, 2, 3] -> *__able_array_i32`, etc.) when
    `ExperimentalMonoArrays` is enabled and the element family is staged;
  - static array `for` loops now iterate directly over typed slices for both
    generic `*Array` and specialized wrappers, instead of routing through
    `__able_array_values(...)` on the specialized path;
  - static array pattern matching / destructuring now bind specialized element
    locals with their concrete Go element type and preserve specialized rest
    tails instead of dropping them back to generic `*Array`;
  - `Array.new()` / `Array.with_capacity()` now lower directly to compiler-owned
    static carriers on typed static paths, and `reserve()` / `clone_shallow()`
    now stay specialized too.
- Regression coverage added in:
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_widening_test.go`
  - updated `v12/interpreters/go/pkg/compiler/compiler_mono_array_contract_test.go`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier|ExperimentalMonoArraysTypedArrayUsesSpecializedWrapper|ExperimentalMonoArraysTypedArrayWrapperUsesSpecializedBoundaryConverters|ExperimentalMonoArraysTypedArrayExecutes|ExperimentalMonoArraysInferredLiteralLoopAndPatternStaySpecialized|ExperimentalMonoArraysFactoryCloneAndReserveStaySpecialized|ExperimentalMonoArraysWidenedSliceExecutes)$|TestCompilerDynamicBoundaryMonoArray' -count=1` (pass, `16.981s`).
  - `cd v12/interpreters/go && git diff --check` (pass).
  - touched compiler files remain under the 1000-line limit.
- Adjacent note:
  - `TestCompilerExecFixtures/06_12_18_stdlib_collections_array_range` still
    fails on the current branch with a runtime error after the array helper
    section; narrowing the new array-factory interception did not change that
    outcome, so it remains a separate compiler/runtime follow-up rather than a
    mono-array widening regression.

# 2026-03-19 — Compiler specialized mono-array carrier tranche (v12)
- Closed the first specialized-wrapper implementation slice for compiler-native
  mono arrays.
- Landed compiler-owned staged carriers plus direct typed lowering in:
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_specs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_mono_arrays.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - shared interface / struct / union conversion helpers
- Tranche details:
  - `Array i32`, `Array i64`, `Array bool`, and `Array u8` now map to
    compiler-owned specialized wrappers (`*__able_array_i32`, etc.) when
    `ExperimentalMonoArrays` is enabled and the array type is explicit;
  - typed array literals, `push`, `get`, `set`, `read_slot`, `write_slot`,
    direct `arr[idx]`, direct `arr[idx] = value`, and `or {}` propagation on
    those typed arrays now operate on native typed Go slices instead of
    `[]runtime.Value`;
  - explicit wrapper/lambda/interface/union/struct boundary helpers now
    convert specialized wrappers to/from runtime carriers directly, and the
    existing dynamic mono-array boundary suite is green again;
  - unannotated arrays intentionally remain on the generic compiler-owned
    `*Array` carrier for now, so pattern/destructuring/iterator surfaces are
    not half-migrated.
- Regression coverage added in:
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_specialized_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_contract_test.go`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier|ExperimentalMonoArraysTypedArrayUsesSpecializedWrapper|ExperimentalMonoArraysTypedArrayWrapperUsesSpecializedBoundaryConverters|ExperimentalMonoArraysTypedArrayExecutes)$' -count=1` (pass, `0.734s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryMonoArray' -count=1` (pass, `15.323s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier|ExperimentalMonoArraysTypedArrayUsesSpecializedWrapper|ExperimentalMonoArraysTypedArrayWrapperUsesSpecializedBoundaryConverters|ExperimentalMonoArraysTypedArrayExecutes|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters)$|TestCompilerDynamicBoundaryMonoArray' -count=1` (pass, `16.414s`).
- Handoff:
  - the next mono-array category is widening specialization coverage beyond the
    explicit typed slice: constructors / stdlib factories, unannotated local
    inference, clone/iteration/pattern paths, and then compiled perf
    re-measurement.

# 2026-03-19 — Compiler mono-array plan revision tranche (v12)
- Closed the mono-array plan-revision tranche for the compiler performance /
  native-lowering arc.
- Replaced the old handle-tagged typed-runtime-store plan with the accepted
  compiler-native direction in:
  - `v12/design/monomorphized-container-abi.md`
  - `v12/design/compiler-monomorphization.md`
  - `v12/design/compiler-native-lowering.md`
  - `PLAN.md`
  - `spec/TODO_v12.md`
- Added a compiler guardrail in:
  - `v12/interpreters/go/pkg/compiler/compiler_mono_array_contract_test.go`
- Tranche details:
  - the historical experimental mono-array work remains in-tree, but it is now
    explicitly documented as compatibility scaffolding / measurement reference,
    not the accepted final architecture;
  - the final target is now pinned to compiler-generated specialized wrappers
    over native Go slices for staged element kinds, with `runtime.ArrayValue`,
    `ArrayStore*`, and runtime typed stores reserved for explicit dynamic
    boundaries or residual compatibility only;
  - `TestCompilerExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier`
    now proves that enabling `ExperimentalMonoArrays` still keeps
    representative static array bodies on the compiler-owned array carrier and
    avoids `runtime.ArrayValue`, `runtime.ArrayStore*`, and broad dynamic
    helper dispatch.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(FeatureFlagMonoArraysDefaultsDisabled|FeatureFlagMonoArraysEnabledViaOptions|ExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier)$' -count=1` (pass).
  - `git diff --check` (pass).
- Handoff:
  - the mono-array plan-revision tranche should now be treated as closed;
  - the next category is the actual specialized-wrapper implementation and
    performance measurement work on top of that revised design.

# 2026-03-19 — Compiler native-lowering re-audit closure (v12)
- Closed the broader compiler re-audit tranche for the current native-lowering
  contract.
- Landed the final array/error-unwrap fix in:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
  - `v12/interpreters/go/pkg/compiler/ir_codegen.go`
- Tightened regression coverage in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_struct_instance_error_unwrap_shim_regression_test.go`
- Closure details:
  - native array bounds/error paths were already returning `runtime.ErrorValue`
    on compiled static paths, but `__able_struct_instance(...)` was still
    rehydrating those errors through an anonymous synthetic struct;
  - that dropped the wrapped `IndexError` definition, so static
    `case _: IndexError => ...` matches on array `set` / index-assignment
    results could fail under the boundary-marker harness with
    `Non-exhaustive match`;
  - `__able_error_to_struct(...)` now preserves the concrete wrapped struct
    payload when an `ErrorValue` carries one, instead of always synthesizing a
    fresh definition-less struct view;
  - the zero-explicit-boundary fixture audit now includes
    `06_08_array_ops_mutability`, so native array mutation/error semantics are
    covered alongside the previously-audited lambda/interface fixtures.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(BroadStaticNativeTouchpointsStayNative|GenericResidualTouchpointsStayNarrow|StaticNativeFixturesExecuteWithoutExplicitBoundaries|NormalizesStructInstanceErrorUnwrap)$' -count=1` (pass, `8.860s`).
  - `cd v12/interpreters/go && ABLE_COMPILER_FIXTURES=06_08_array_ops_mutability GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerStaticNativeFixturesExecuteWithoutExplicitBoundaries/06_08_array_ops_mutability$' -count=1` (pass, `2.461s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(BroadStaticNativeTouchpointsStayNative|GenericResidualTouchpointsStayNarrow|StaticNativeFixturesExecuteWithoutExplicitBoundaries|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|InterfaceParamAndReturnStayNative|TypedInterfaceAssignmentStaysNative|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative|NormalizesStructInstanceErrorUnwrap)$|TestCompilerExecFixtures/(06_01_compiler_placeholder_lambda|06_01_compiler_bound_method_value|06_08_array_ops_mutability|10_03_interface_type_dynamic_dispatch|10_15_interface_default_generic_method)$' -count=1` (pass, `12.058s`).
  - `git diff --check` (pass).
- Handoff:
  - the broader compiler re-audit tranche should now be treated as closed;
  - the next category is now genuinely different work: performance-oriented
    specialization/monomorphization on top of the enforced native-lowering
    contract.

# 2026-03-18 — Compiler native touchpoint enforcement tranche (v12)
- Closed the touchpoint-enforcement follow-up for the compiler-native lowering
  arc.
- Landed the broader static/generic touchpoint audits in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
- Tranche details:
  - broad static compiled paths now have a combined-source audit that covers
    arrays, structs, named unions, object-safe interfaces, and native
    callables in one no-fallback source and fails if generated compiled
    function bodies regress to `__able_call_value(...)`,
    `__able_member_get*`, `__able_index*`, `__able_method_call_node(...)`,
    `bridge.MatchType(...)`, `__able_try_cast(...)`, `__able_any_to_value(...)`,
    or panic/IIFE-style control scaffolding on those native paths;
  - the intentionally residual generic-interface edge now has a paired audit
    that proves the compiled body stays on the native carrier and narrows the
    runtime crossing to `__able_iface_*_to_runtime_value(...)` plus
    `__able_method_call_node(...)`, without regressing to the broader dynamic
    helper layer;
  - representative static fixtures for placeholder lambdas, bound method
    values, and object-safe interface dispatch now execute under the boundary
    marker harness with both fallback and explicit boundary counts at zero.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(BroadStaticNativeTouchpointsStayNative|GenericResidualTouchpointsStayNarrow|StaticNativeFixturesExecuteWithoutExplicitBoundaries)$' -count=1` (pass, `5.828s`).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(BroadStaticNativeTouchpointsStayNative|GenericResidualTouchpointsStayNarrow|StaticNativeFixturesExecuteWithoutExplicitBoundaries|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|InterfaceParamAndReturnStayNative|TypedInterfaceAssignmentStaysNative|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative)$|TestCompilerExecFixtures/(06_01_compiler_placeholder_lambda|06_01_compiler_bound_method_value|10_03_interface_type_dynamic_dispatch|10_15_interface_default_generic_method)$' -count=1` (pass, `10.747s`).
- Handoff:
  - the touchpoint-enforcement tranche should now be treated as closed;
  - the next category is now genuinely different work: merge-time re-audit of
    Claude's staged compiler changes against the contract, plus broader
    performance-oriented specialization/monomorphization.

# 2026-03-18 — Compiler interface/global lookup audit batching (v12)
- Closed the strict lookup-audit batching follow-up for the compiler-native
  lowering arc.
- Landed the audit split in:
  - `v12/interpreters/go/pkg/compiler/compiler_interface_lookup_audit_test.go`
- Batch-audit details:
  - the default strict interface/global lookup audit no longer runs as one
    oversized `TestCompilerInterfaceLookupBypassForStaticFixtures` sweep;
  - the default fixture set is now distributed across four top-level batch
    tests:
    - `TestCompilerInterfaceLookupBypassForStaticFixturesBatch1`
    - `TestCompilerInterfaceLookupBypassForStaticFixturesBatch2`
    - `TestCompilerInterfaceLookupBypassForStaticFixturesBatch3`
    - `TestCompilerInterfaceLookupBypassForStaticFixturesBatch4`
  - the unsuffixed `TestCompilerInterfaceLookupBypassForStaticFixtures`
    selector is now reserved for explicit fixture subsets via
    `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES`, which keeps targeted auditing
    available without forcing the whole default sweep through one test body;
  - default batching uses a deterministic round-robin split over the fixture
    list so the suite remains mechanically stable and stays under the repo's
    per-test one-minute guardrail.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1` (pass, selector now skips quickly when no explicit fixture subset is requested).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=10_03_interface_type_dynamic_dispatch GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1` (pass, `1.854s`).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixturesBatch1$' -count=1` (pass, `19.241s`).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixturesBatch2$' -count=1` (pass, `21.622s`).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixturesBatch3$' -count=1` (pass, `18.178s`).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixturesBatch4$' -count=1` (pass, `18.234s`).
- Handoff:
  - the strict lookup-audit batching follow-up should now be treated as
    closed;
  - the next category is the broader compiler re-audit/enforcement pass:
    tighten the allowed dynamic-carrier touchpoints mechanically, re-audit the
    larger staged compiler arc against the native-lowering contract, and then
    continue performance-oriented specialization/monomorphization work.

# 2026-03-18 — Compiler native callable/function-type existential lowering completion (v12)
- Closed the callable/function-type existential tranche for the current
  compiler-native lowering arc.
- Landed the callable carrier and boundary fixes in:
  - `v12/interpreters/go/pkg/compiler/generator_native_callables.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_callables.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_callable_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_bound_method_values.go`
  - `v12/interpreters/go/pkg/compiler/generator_placeholders.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/generator_local_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
- Added/updated regression coverage in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_callable_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_local_function_definition_no_fallback_test.go`
- Closure details:
  - `FunctionTypeExpression` now lowers to generated native Go callable
    carriers instead of defaulting to `any`;
  - direct lambdas, local function definitions, placeholder lambdas, and
    captured bound methods now emit those native callable carriers instead of
    `runtime.NativeFunctionValue` on static compiled paths;
  - function-typed params, struct fields, wrapper arg/return conversion, and
    native-interface conversion now round-trip through explicit generated
    callable boundary helpers instead of broad `runtime.Value` coercion;
  - callable-body control propagation no longer uses callback-style
    `error` returns internally; native callable bodies capture
    `*__ableControl` with labeled break blocks and return `(zero, control)`
    directly like compiled functions;
  - higher-order fixture surfaces including partial application, trailing
    lambdas, callable `apply`, iterator/default-interface higher-order calls,
    and the stdlib reporters fixture now stay on the native callable ABI.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative|NativeCallableExecutes)$|TestCompilerNoFallbacksForLocalFunctionDefinition(Statement|ShadowingTypedBinding)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PlaceholderLambdaStaysNative|BoundMethodValueStaysNative|FunctionTypedParamStaysNative|NativeCallableExecutes|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|InterfaceLookupGenericMethodFixturesRegression)$|TestCompilerNoFallbacksForLocalFunctionDefinition(Statement|ShadowingTypedBinding)$|TestCompilerExecFixtures/(06_01_compiler_placeholder_lambda|06_01_compiler_bound_method_value|06_04_function_call_eval_order_trailing_lambda|07_02_lambdas_closures_capture|07_02_01_verbose_anonymous_fn|07_04_apply_callable_interface|07_04_trailing_lambda_method_syntax|07_05_partial_application|06_12_26_stdlib_test_harness_reporters|14_01_language_interfaces_index_apply_iterable|10_04_interface_dispatch_defaults_generics|10_15_interface_default_generic_method)$' -count=1` (pass).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1` (pass, `72.154s`).
- Handoff:
  - the callable/function-type existential tranche should now be treated as
    closed;
  - the next category is genuinely different work: re-audit the broader
    staged compiler arc against the current native-lowering contract, keep
    shrinking dynamic-carrier touchpoints mechanically, and split/batch the
    strict compiler audits so they meet the repo's one-minute test target.

# 2026-03-18 — Compiler generic-interface existential lowering completion (v12)
- Closed the non-object-safe/generic interface existential tranche for the
  current compiler-native lowering arc.
- Landed the generic-interface carrier and dispatch-edge fixes in:
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_generic_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
- Added regression coverage in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_generic_test.go`
- Closure details:
  - pure-generic interfaces now keep generated native carriers instead of
    collapsing typed locals/params back to `runtime.Value`;
  - generic interface/default-interface methods now keep the receiver on that
    native carrier and cross into runtime only at the explicit generic
    dispatch edge via `__able_method_call_node(...)`;
  - generic dispatch results now convert back from `runtime.Value` into the
    best-known native Go carrier before the surrounding compiled code sees
    them again;
  - native-interface adapter population now requires actual generic-method
    implementers for generic-only carriers instead of treating every compiled
    type as a valid adapter when no object-safe methods are present.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes|InterfaceLookupGenericMethodFixturesRegression)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InterfaceParamAndReturnStayNative|NativeInterfaceExecutes|TypedInterfaceAssignmentStaysNative|GenericInterfaceReturnFromStructLiteralStaysNative|UnionTargetInterfaceAssignmentStaysNative|StaticIndexInterfacesStayNative|ConcreteReceiverInterfaceMethodStaysNative|ConcreteReceiverApplyStaysNative|InterfaceMethodWithLambdaArgStaysNative|NativeInterfaceRuntimeAdapterUsesStructZeroForVoidReturn|NativeInterfaceRuntimeAdapterWritesBackPointerArgs|PureGenericInterfaceAssignmentUsesNativeCarrier|DefaultGenericInterfaceMethodUsesNativeReceiverBoundary|GenericInterfaceExistentialExecutes)$|TestCompilerExecFixtures/(10_04_interface_dispatch_defaults_generics|10_15_interface_default_generic_method|14_01_language_interfaces_index_apply_iterable)$|TestCompilerDynamicBoundary(CallbackInterfaceConversion(Success|Failure)Markers|MonoArrayCallbackInterfaceConversion(Success|Failure)Markers)$' -count=1` (pass).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookup(BypassForStaticFixtures|GenericMethodFixturesRegression)$' -count=1` (pass).
  - `git diff --check` (clean for touched files after this tranche).
- Handoff:
  - this generic-interface existential tranche should now be treated as
    closed;
  - the next category is callable/function-type existentials, especially the
    callable-driven generic inference surfaces still present on `Iterator<T>`
    (`map`, `filter`, `filter_map`, `collect`), followed by residual
    runtime-boundary tightening around those callable surfaces.
  - process note: the full strict interface lookup audit passed with total
    lookup counters forced to zero, but the all-fixture run took about 88s and
    should be split or batched later to stay inside the repo's sub-minute test
    target.

# 2026-03-18 — Compiler object-safe interface tranche audit closure (v12)
- Closed the remaining cleanup and audit items for the object-safe native
  interface tranche.
- Landed the final codegen fixes in:
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_union_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_array_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match_arrays.go`
- Final closure details:
  - native interface runtime adapters now round-trip `void` as `struct{}` on
    the Go side instead of failing conversion at runtime;
  - pointer-backed native interface arguments are now written back after
    runtime-backed interface dispatch, which closes the `ProgressReporter`
    state-loss hole exposed by `06_12_26_stdlib_test_harness_reporters`;
  - native interface `*_from_value(...)` helpers now recover concrete compiled
    adapters directly before falling back to the generic runtime adapter path;
  - match-pattern lowering no longer emits unused native-union unwrap temps
    when a typed branch has no bindings;
  - `generator_match.go` is back under the repository file-size limit after
    splitting array-pattern helpers into `generator_match_arrays.go`.
- Regression coverage added/updated in:
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_interface_lookup_audit_test.go`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NativeInterfaceRuntimeAdapterUsesStructZeroForVoidReturn|NativeInterfaceRuntimeAdapterWritesBackPointerArgs|GenericInterfaceReturnFromStructLiteralStaysNative|UnionTargetInterfaceAssignmentStaysNative|StaticIndexInterfacesStayNative|ConcreteReceiverInterfaceMethodStaysNative|ConcreteReceiverApplyStaysNative|InterfaceMethodWithLambdaArgStaysNative)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupReportersFixtureRegression$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$' -count=1` (pass).
  - `git diff --check` (clean).
- Handoff:
  - the completed object-safe tranche should now be treated as closed;
  - the next category is genuinely different work: non-object-safe/generic
    interface existentials, callable/function-type existentials, and residual
    runtime-boundary tightening around those surfaces.

# 2026-03-11 — Compiler test build-cache stabilization (v12)
- Documented the next post-object-safe compiler lowering categories in
  `PLAN.md`:
  - non-object-safe/generic interface existentials;
  - native callable/function-type existential ABI;
  - tighter runtime-boundary auditing for residual interface/callable
    `runtime.Value` usage.
- Investigated the oversized repo-local Go build cache at
  `v12/interpreters/go/.gocache` and confirmed the dominant growth source was
  compiler tests building generated packages from fresh random module-local
  temp directories like `tmp/ablec-interface-lookup-fixture-2902499314`.
- Landed deterministic compiler test workdirs in
  `v12/interpreters/go/pkg/compiler/compiler_test_workdir_test.go` and updated
  the compiler exec/audit/benchmark harness tests to reuse stable module-local
  `tmp/ablec-*` paths instead of `os.MkdirTemp(...)` package roots.
- Fixed a native-interface synthesis recursion bug surfaced by the stricter
  rerun path:
  - `generator_native_interfaces.go` now installs an in-progress placeholder
    entry before recursively mapping interface method signatures;
  - `generator.go` tracks native interfaces currently under construction so
    recursive interface/union signatures reuse the existing carrier token
    instead of re-entering synthesis until stack overflow.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ExecHarness|InterfaceParamAndReturnStayNative|NativeInterfaceExecutes|TypedInterfaceAssignmentStaysNative)$|TestCompilerDynamicBoundary(CallbackInterfaceConversion(Success|Failure)Markers|MonoArrayCallbackInterfaceConversion(Success|Failure)Markers)$' -count=1` (pass).
  - Repeated identical compiled-harness runs now reuse the same cache entries:
    two back-to-back `TestCompilerExecHarness` runs produced
    `delta_bytes=0 delta_files=0` for `v12/interpreters/go/.gocache`.
- Current state:
  - the cache-growth mechanism is fixed for these compiler tests going forward,
    but the existing `v12/interpreters/go/.gocache` contents remain huge until
    explicitly pruned;
  - the broad `TestCompilerInterfaceLookupBypassForStaticFixtures` audit still
    has unrelated staged-compiler correctness/build failures, but the previous
    stack-overflow failure is resolved.

# 2026-03-18 — Compiler object-safe interface concrete-receiver closure (v12)
- Closed the remaining concrete-receiver holes in the object-safe native
  interface tranche:
  - concrete `Index` / `IndexMut` operator lowering now dispatches through
    compiled impls instead of bridging native structs into `__able_index` /
    `__able_index_set`;
  - concrete receiver default-interface method calls now dispatch through
    compiled impl bodies instead of falling back to `__able_method_call_node`;
  - concrete `Apply` calls now dispatch through compiled impls instead of
    `__able_call_value`;
  - native integer-result coercion for compiled bitwise/shift expressions now
    allows the stdlib `decode_multibyte` path to stay on the static compiler
    path instead of failing with `binary expression type mismatch`.
- Code landed in:
  - `v12/interpreters/go/pkg/compiler/generator_index_static.go`
  - `v12/interpreters/go/pkg/compiler/generator_apply_static.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_test.go`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(GenericInterfaceReturnFromStructLiteralStaysNative|UnionTargetInterfaceAssignmentStaysNative|StaticIndexInterfacesStayNative|ConcreteReceiverInterfaceMethodStaysNative|ConcreteReceiverApplyStaysNative|StaticCallReturningConcreteErrorCoercesToNativeResult)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupBypassForStaticFixtures/(06_12_21_stdlib_fs_path|06_12_22_stdlib_io_temp|06_03_operator_overloading_interfaces|14_01_language_interfaces_index_apply_iterable)$' -count=1` now leaves only:
    - `06_12_21_stdlib_fs_path`
    - `06_12_22_stdlib_io_temp`
    - the final `iter.next()` tail in `14_01_language_interfaces_index_apply_iterable`
- Boundary / handoff:
  - the remaining `14_01` dynamic surface is `Iterator<T>`: `counter.iterator()`
    now dispatches through the compiled `Iterable.iterator` impl, but its
    return still normalizes to `runtime.Value`, and `iter.next()` still routes
    through `__able_method_call_node`;
  - that is expected to move with the next category, not the completed
    object-safe tranche, because `Iterator<T>` is not object-safe as defined in
    the stdlib today (`filter`, `map`, `filter_map`, `collect` are generic
    methods on the interface).

# 2026-03-11 — Compiler object-safe native interface lowering completion (v12)
- Closed the fully bound object-safe native interface/existential tranche in:
  - `v12/interpreters/go/pkg/compiler/generator_native_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_interfaces.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_interface_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_interface_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_broad_test.go`
- Behavior / architecture changes:
  - fully bound object-safe Able interfaces now lower to generated native Go
    interface carriers plus concrete/runtime adapters instead of collapsing
    static compiled paths back to `runtime.Value` or `any`;
  - static params, returns, typed local assignment, struct fields, and direct
    method dispatch now stay on those native interface carriers;
  - residual concrete-receiver holes on that static path are now also closed:
    concrete `Index` / `IndexMut`, default-interface method calls, and
    concrete `Apply` lowering all dispatch through compiled impls instead of
    dynamic member/call helpers;
  - wrapper returns/args, lambda ABI conversion, and dynamic callback boundary
    conversion now use explicit generated interface adapters instead of
    implicitly binding interface-typed values as raw `runtime.Value`;
  - native interface adapter population now refreshes against the current impl
    set instead of freezing the first cached adapter view, which fixed the
    no-fallback static-fixture regression where typed interface locals rejected
    struct literals even though matching compiled impls existed;
  - interface/open unions such as `String | Tag` still keep residual
    non-native branches as explicit `runtime.Value` union members, but the
    interface branch itself now stays native end-to-end on the static path.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InterfaceParamAndReturnStayNative|NativeInterfaceExecutes|TypedInterfaceAssignmentStaysNative|InterfaceBranchUnionStaysOnNativeCarrier|BroadNativeUnionExecutes|GenericUnionAliasesStayNative|MultiMemberUnionMatchStaysNative)$|TestCompilerExecFixtures/(10_03_interface_type_dynamic_dispatch|10_10_interface_inheritance_defaults|10_16_interface_value_storage|10_17_interface_overload_dispatch|11_02_option_result_or_handlers)$' -count=1` (pass).
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES='10_03_interface_type_dynamic_dispatch,10_10_interface_inheritance_defaults,10_16_interface_value_storage,10_17_interface_overload_dispatch' GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(CallbackInterfaceConversion(Success|Failure)Markers|MonoArrayCallbackInterfaceConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - this object-safe interface carrier category is complete; the next category
    is different work: non-object-safe/generic interface existentials,
    callable/function-type existentials, and further runtime-boundary
    tightening around those residual surfaces.

# 2026-03-11 — Compiler residual dynamic-helper panic cleanup completion (v12)
- Closed the residual dynamic-helper panic cleanup tranche in:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls_tail.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_runtime_call_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_iterators.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_await.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_concurrency.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_helper_reachability_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_*shim_regression_test.go`
- Behavior / architecture changes:
  - generated `__able_member_get`, `__able_member_set`,
    `__able_member_get_method`, `__able_member`, and `__able_method_call*`
    helpers now use explicit ordinary Go `error` / `*__ableControl` returns
    instead of raw panic-based error signaling;
  - the temporary `recover`-based bridge wrappers
    (`__able_error_from_panic`, `__able_bridge_call_value_with_node`,
    `__able_bridge_call_named_with_node`) are gone, and generated bridge sites
    now call the runtime bridge directly;
  - generated compiler callsites that consume dynamic member/method helpers now
    branch through shared control-return helpers instead of embedding bare
    helper calls as value expressions;
  - lambda callback argument conversion now preserves nullable pointer-struct
    carriers by retaining the original Able parameter type expression, while
    still rejecting `nil` for non-nullable struct params.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DynamicHelpersRemovePanicBridgeWrappers|NormalizesArrayMemberReceiverUnwrap|RemovesTypeRefPointerMemberGetMethodShim|NormalizesInterfaceMemberGetMethodDispatch|PrefersInterfaceDispatchBeforeUFCSInMemberGetMethod|RemovesStructDefinitionPointerMemberGetMethodShim|RemovesStructDefinitionPointerQualifiedCallableShim|RemovesTypeRefPointerQualifiedCallableShim|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundary' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NormalizesCallValue(UnwrapBranches|NativeFunctionDispatchBranches|NativeDispatchBranches|NativeBoundMethodDispatchBranches|BoundMethodDispatchBranches|FunctionThunkDispatch)|Removes(StructDefinitionPointerQualifiedCallableShim|NilPointerQualifiedCallableShim|TypeRefPointerQualifiedCallableShim|StructDefinitionPointerMemberGetMethodShim|TypeRefPointerMemberGetMethodShim|PackagePublicMemberGetMethodShim|ImplNamespacePointerQualifiedCallableShim|ImplNamespacePointerMemberGetMethodShim|ErrorValueMemberGetMethodShim)|NormalizesInterfaceMemberGetMethodDispatch|PrefersInterfaceDispatchBeforeUFCSInMemberGetMethod|DynamicHelpersRemovePanicBridgeWrappers)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|DirectErrorMessageAndCauseStayNative|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionExecutes|ExplicitErrorUnionParamAndMatchStayNative|MultiMemberUnionMatchStaysNative|GenericUnionAliasesStayNative|InterfaceBranchUnionStaysOnNativeCarrier|SingletonStructBoundaryAcceptsRuntimeDefinition|BroadNativeUnionExecutes|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|DynamicHelpersRemovePanicBridgeWrappers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|NormalizesArrayMemberReceiverUnwrap)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns|06_01_compiler_or_else_error_union|06_01_compiler_or_else|06_01_compiler_rescue|06_01_compiler_ensure_rethrow|06_01_compiler_ensure_error_passthrough|06_01_compiler_result_return|06_01_compiler_raise_error_interface|06_01_compiler_raise_non_error|11_02_option_result_or_handlers)$' -count=1` (pass).
- Remaining gap:
  - this cleanup category is complete; the next category is broader
    interface/existential lowering and further runtime-boundary tightening, not
    more residual panic-wrapper cleanup.

# 2026-03-11 — Compiler dynamic-boundary control propagation completion (v12)
- Closed the explicit dynamic-call boundary control tranche in:
  - `v12/interpreters/go/pkg/compiler/generator_render_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_runtime_call_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls_tail.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_iterators.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_strings.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_hash_maps.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_await.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_concurrency.go`
  - `v12/interpreters/go/pkg/compiler/compiler_call_value_*_shim_regression_test.go`
- Behavior / architecture changes:
  - compiled functions and helper-generated static callsites now propagate
    explicit `*__ableControl` results through dynamic `call_value` /
    `call_named` paths instead of letting native-wrapper errors explode through
    `__able_panic_on_error(...)`;
  - `__able_call_value(...)` and `__able_call_named(...)` now return
    `(runtime.Value, *__ableControl)`, and compiled dynamic callsites convert
    that control with ordinary Go branching via generated `controlCheckLines`;
  - callback-bearing runtime helpers (`await`, mutex/channel callbacks,
    hashmap iteration helpers, string-byte extraction, iterator yield) now
    translate dynamic call-boundary control back into ordinary Go `error`
    returns instead of panicking;
  - plain non-nullable struct wrapper arg conversion now rejects `nil` before
    compiled bodies run, fixing the dynamic callback struct-conversion failure
    path;
  - explicit dynamic bridge calls now contain residual panic-based helper
    failures (`bridge.CallValueWithNode` / `bridge.CallNamedWithNode`) at the
    boundary and normalize them back into ordinary errors so boundary markers
    and diagnostics survive callback failures instead of being lost to raw
    panics.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryCallback(Array|Union|Nullable)ConversionFailureMarkers$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundary' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NormalizesCallValue(UnwrapBranches|NativeFunctionDispatchBranches|NativeDispatchBranches|NativeBoundMethodDispatchBranches|BoundMethodDispatchBranches|FunctionThunkDispatch)|Removes(StructDefinitionPointerQualifiedCallableShim|NilPointerQualifiedCallableShim|TypeRefPointerQualifiedCallableShim|StructDefinitionPointerMemberGetMethodShim|TypeRefPointerMemberGetMethodShim|PackagePublicMemberGetMethodShim|ImplNamespacePointerQualifiedCallableShim|ImplNamespacePointerMemberGetMethodShim|ErrorValueMemberGetMethodShim)|NormalizesInterfaceMemberGetMethodDispatch|PrefersInterfaceDispatchBeforeUFCSInMemberGetMethod)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|DirectErrorMessageAndCauseStayNative|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionExecutes|ExplicitErrorUnionParamAndMatchStayNative|MultiMemberUnionMatchStaysNative|GenericUnionAliasesStayNative|InterfaceBranchUnionStaysOnNativeCarrier|SingletonStructBoundaryAcceptsRuntimeDefinition|BroadNativeUnionExecutes|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap))$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns|06_01_compiler_or_else_error_union|06_01_compiler_or_else|06_01_compiler_rescue|06_01_compiler_ensure_rethrow|06_01_compiler_ensure_error_passthrough|06_01_compiler_result_return|06_01_compiler_raise_error_interface|06_01_compiler_raise_non_error|11_02_option_result_or_handlers)$' -count=1` (pass).
- Remaining gap:
  - this explicit dynamic-boundary control category is complete; the next,
    separate hardening category is to remove the residual panic-based dynamic
    helper internals (`__able_member_get_method`, `__able_member_get`, etc.)
    so the narrow boundary `recover` containment layer can be deleted.

# 2026-03-11 — Compiler broader native-union carrier widening (v12)
- Completed the next native-union widening tranche in:
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_ident.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_broad_test.go`
- Behavior / architecture changes:
  - closed native unions are no longer limited to the first two-member nominal
    slice; multi-member nominal unions now stay on generated native union
    carriers instead of falling back to `any`;
  - generic union aliases that normalize to native nullable/result carrier
    families, such as `Option T = nil | T` and `Result T = Error | T`, now stay
    on those same native carriers on compiled static paths;
  - broader interface/open unions such as `String | Tag` now also stay on
    generated native union carriers, with non-native residual members carried as
    explicit `runtime.Value` branches inside the generated carrier instead of
    forcing the whole union back to `any`;
  - singleton struct boundary converters now accept runtime
    `StructDefinitionValue` inputs as well as `StructInstanceValue`, which fixes
    interpreted callers passing bare singleton values like `Blue` into compiled
    native struct/union params and keeps the widened native union path
    executable, not just compile-time green.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(MultiMemberUnionMatchStaysNative|GenericUnionAliasesStayNative|InterfaceBranchUnionStaysOnNativeCarrier|SingletonStructBoundaryAcceptsRuntimeDefinition|BroadNativeUnionExecutes)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionExecutes|NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|MultiMemberUnionMatchStaysNative|GenericUnionAliasesStayNative|InterfaceBranchUnionStaysOnNativeCarrier|SingletonStructBoundaryAcceptsRuntimeDefinition|BroadNativeUnionExecutes)$|TestCompilerExecFixtures/(06_01_compiler_result_return|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers|06_01_compiler_nullable_param|06_01_compiler_nullable_return)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - this broader carrier-widening category is complete for the current native
    union ABI tranche; the next category is explicit control-result
    propagation, while deeper interface existential lowering beyond the current
    residual `runtime.Value` union branches remains later follow-on work.

# 2026-03-11 — Compiler native `Error` cleanup completion (v12)
- Closed the remaining narrow native-`Error` cleanup gaps in:
  - `v12/interpreters/go/pkg/compiler/generator_native_error_methods.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_calls_lambda.go`
  - `v12/interpreters/go/pkg/compiler/generator_overloads.go`
  - `v12/interpreters/go/pkg/compiler/generator_concurrency.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_result_test.go`
- Behavior / architecture changes:
  - direct compiled `Error.message()` now lowers to native
    `runtime.ErrorValue.Message` field reads instead of
    `__able_method_call_node(...)` / dynamic member lookup;
  - direct compiled `Error.cause()` now lowers to native
    `runtime.ErrorValue.Payload["cause"]` extraction plus normalized nullable
    error coercion only when needed, instead of dynamic method dispatch;
  - native concrete-error normalization now preserves compiled `cause()`
    payloads as well as compiled `message()` text when constructing
    `runtime.ErrorValue`;
  - struct field conversion/rendering now supports both `Error` and `?Error`
    native carriers, so error-bearing structs no longer fall back to
    unsupported-field codegen on static paths;
  - the generic expected-type guards used by dynamic-call / overload / await
    lowering now recognize compiler-native nullable carriers, so `*runtime.ErrorValue`
    no longer trips false "unknown expected type" compile failures.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|DirectErrorMessageAndCauseStayNative|DirectErrorCauseExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionExecutes|NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|DirectErrorMessageAndCauseStayNative|DirectErrorCauseExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|MatchOnErrorUnionExecutes|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_result_return|11_02_option_result_or_handlers|06_01_compiler_or_else_error_union|06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NormalizesStructInstanceErrorUnwrap|RemovesErrorValueMemberGetMethodShim|NormalizesErrorBuiltinMemberReceivers|RegistersBuiltinErrorMemberMethods)$' -count=1` (pass).
- Remaining gap:
  - the native `Error` special-case cleanup is complete; remaining union work
    now moves into a different category: broader interface/open/generic union
    lowering, followed by explicit control-result propagation.

# 2026-03-11 — Compiler native nullable `?Error` slice (v12)
- Extended the native error-carrier family to `?Error` in:
  - `v12/interpreters/go/pkg/compiler/generator_nullable_native.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_nullable.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_nullable_native_test.go`
- Behavior / architecture changes:
  - `?Error` now lowers to `*runtime.ErrorValue` on compiled static paths
    instead of falling back to `any` / `runtime.Value`;
  - concrete `Error` implementers flowing into nullable `Error` positions now
    normalize through the same compiled `runtime.ErrorValue` path used by the
    direct `Error` / `!T` slices, then wrap into the native nullable pointer
    carrier;
  - native nullable helper emission now includes explicit `Error` adapter
    helpers, so wrapper/lambda boundaries and native `match` on `?Error` stay
    on the compiled nullable path without `__able_try_cast(...)` /
    `bridge.MatchType(...)`.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableErrorReturnAndMatchStayNative|NullableErrorReturnExecutes|DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|MatchOnErrorUnionExecutes|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_result_return|11_02_option_result_or_handlers|06_01_compiler_or_else_error_union|06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - this is still part of the narrow native `Error` carrier family, not a
    general native lowering for arbitrary interface existential types.

# 2026-03-11 — Compiler native `Error` carrier widening (v12)
- Widened the first native union/result slice so plain `Error` type positions
  now use the native `runtime.ErrorValue` carrier in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_result_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_test.go`
- Behavior / architecture changes:
  - `Error` no longer maps to `runtime.Value` on compiled static paths; it now
    maps to `runtime.ErrorValue`, which lets explicit `Error` returns and
    explicit `String | Error` unions stay on the native carrier family already
    used by the first `!T` slice;
  - static coercion now normalizes concrete `Error` implementers into
    `runtime.ErrorValue` at the compiled edge, including direct `Error`
    returns and concrete-error values flowing into native `Error` union
    members;
  - wrapper argument conversion for native `Error` params now validates the
    runtime value against `Error` first, then normalizes it into the compiled
    `runtime.ErrorValue` carrier instead of defaulting back to `runtime.Value`.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionExecutes)$|TestCompilerExecFixtures/(06_01_compiler_result_return|06_01_compiler_or_else_error_union|11_02_option_result_or_handlers)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(DirectErrorReturnUsesNativeCarrier|DirectErrorReturnExecutes|ExplicitErrorUnionParamAndMatchStayNative|ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|MatchOnErrorUnionExecutes|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_result_return|11_02_option_result_or_handlers|06_01_compiler_or_else_error_union|06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - this is still only a native `Error` special-case, not a general native
    interface-carrier ABI; broader interface branches and open/generic unions
    still remain follow-on work.

# 2026-03-11 — Compiler native result `!T` slice (v12)
- Landed the first compiler-native `ResultTypeExpression` slice in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_rescue.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_result_test.go`
- Behavior / architecture changes:
  - `types.go` now maps the first native `!T` slice to generated native union
    carriers when the result shape normalizes to the current two-member
    `runtime.ErrorValue | T` ABI;
  - compiled returns, branch coercion, and propagation on those shapes now wrap
    and unwrap the native carrier directly instead of lowering the whole result
    path to `any`;
  - no-bootstrap concrete `Error` implementers now normalize into
    `runtime.ErrorValue` by calling the compiled `Error.message()` impl first,
    instead of depending on `bridge.ErrorValue(...)` to recover the message at
    runtime from interpreter-backed interface machinery;
  - `value(false)! or { err => ... }` now executes correctly on the compiled
    no-bootstrap path for the current native result carrier slice.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ResultReturnUsesNativeCarrier|ResultPropagationUsesNativeCarrier|ResultPropagationExecutes|InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|MatchOnErrorUnionExecutes|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_result_return|11_02_option_result_or_handlers|06_01_compiler_or_else_error_union|06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - the current result slice is still deliberately narrow: broader
    interface-branch, generic/open union, and explicit control-result lowering
    work is still pending.

# 2026-03-11 — Compiler native union `Error` match branches (v12)
- Extended the first closed native union slice so typed `Error` match branches
  on native union carriers now stay on the static path in:
  - `v12/interpreters/go/pkg/compiler/generator_native_union_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_test.go`
- Behavior / architecture changes:
  - native closed-union typed-pattern resolution now recognizes the case where
    exactly one concrete union member implements the requested interface name,
    starting with the `Error`-typed branch shape used by the current result /
    error-handling surface;
  - static `match` conditions on those branches now unwrap through the
    generated native union helper first instead of falling back to
    `bridge.MatchType(...)` / `__able_try_cast(...)`;
  - when the typed branch binds the whole matched error value
    (`case err: Error => ...`), the matched concrete value converts to
    `runtime.Value` only at that branch edge so existing dynamic `Error`
    method calls keep working without re-boxing the whole match path;
  - the same typed-pattern resolution is now shared with assignment-pattern
    lowering so the union/interface discrimination logic stays centralized.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|MatchOnErrorUnionUsesNativeCarrierDetection|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter|MatchOnErrorUnionExecutes)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns|06_01_compiler_or_else_error_union|06_01_compiler_result_return|11_02_option_result_or_handlers)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - `ResultTypeExpression` still lowers to `any`, and this new static typed
    branch support is still limited to the narrow “one concrete member
    implements the interface” closed-union case rather than a full native
    interface carrier ABI.

# 2026-03-11 — Compiler native union `or {}` failure detection (v12)
- Corrected `or {}` lowering for the first closed native union slice in:
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_test.go`
- Behavior / architecture changes:
  - `compileOrElseExpression(...)` now recognizes native two-branch union
    carriers whose failure branch is a native `Error` implementer and treats
    that branch as the failure path instead of only probing `runtime.Value` /
    nullable carriers;
  - the success path stays on the native success carrier and unwraps through the
    generated union helper instead of forcing the whole `or {}` expression back
    through runtime-value failure checks;
  - when an `err => ...` binding is present, the failure branch converts to
    `runtime.Value` only at the handler edge so handler code can keep using the
    existing `Error` method surface without re-boxing the whole expression path;
  - nullable `or {}` inference was tightened at the same time so unannotated
    handlers prefer the unwrapped payload type (`T`), not the native carrier
    type (`*T`), preserving nil-failure semantics for the existing native
    nullable slice.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|NullableI32ReturnAndOrElseStayNative)$|TestCompilerExecFixtures/(06_01_compiler_or_else_error_union|11_02_option_result_or_handlers|06_01_compiler_result_return)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|OrElseOnErrorUnionUsesNativeCarrierDetection|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns|06_01_compiler_or_else_error_union|06_01_compiler_result_return|11_02_option_result_or_handlers)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - `ResultTypeExpression` still lowers to `any` because a fully native `!T`
    slice is coupled to native interface/error carriers; the broader result
    ABI remains follow-on work.

# 2026-03-11 — Compiler closed native union slice (v12)
- Landed the first closed native union carrier pass in:
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_union_test.go`
- Behavior / architecture changes:
  - direct two-member `UnionTypeExpression` shapes whose branches already map to
    native Go carriers now lower to generated Go interfaces plus wrapper
    carrier structs instead of `any`;
  - named union definitions that normalize to the same two-branch native shape
    now map to those same generated carriers;
  - compiled params/returns, explicit wrapper/lambda boundary adapters, and
    runtime-value conversion helpers now use explicit generated union helpers
    instead of broad `any` conversion for those shapes;
  - typed `match` lowering on static native union carriers now unwraps branches
    through generated helper functions instead of routing through
    `__able_try_cast(...)`;
  - the named-union path now normalizes parser output where `union A = B | C`
    arrives as a single `UnionTypeExpression` inside `def.Variants`.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(InlineClosedUnionParamAndMatchStayNative|NamedClosedUnionReturnWrapsNativeVariants|NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers|CallbackUnionConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - `ResultTypeExpression` still lowers to `any`, and broader union surfaces
    such as generic/open unions, interface-branch unions, and non-typed union
    pattern forms still need native lowering.

# 2026-03-10 — Compiler native nullable scalar widening (v12)
- Broadened the compiler-native nullable carrier slice from the original trio to
  the full compiler-native scalar family in:
  - `v12/interpreters/go/pkg/compiler/generator_nullable_native.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_nullable.go`
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_literals.go`
  - `v12/interpreters/go/pkg/compiler/compiler_nullable_native_test.go`
- Behavior / architecture changes:
  - `mapNullableType(...)` now lowers compiler-native scalar nullable values to
    typed Go pointer carriers across the full currently supported scalar set:
    `?bool`, `?String`, `?char`, `?f32`, `?f64`, `?isize`, `?usize`, `?i8`,
    `?i16`, `?i32`, `?i64`, `?u8`, `?u16`, `?u32`, `?u64`;
  - generated runtime helper emission for those carriers now lives in
    `generator_render_runtime_nullable.go`, which keeps
    `generator_render_runtime.go` under the repo's 1000-line limit while
    centralizing the native nullable boundary adapter family;
  - nullable integer, float, and char literals now lower directly to typed
    native pointer carriers (`__able_ptr(...)`) when a static nullable scalar
    type is expected instead of falling back to `any`.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|NullableI64ReturnAndOrElseStayNative|NullableF64ReturnStayNative|NullableCharParamAndMatchStayNative|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - general unions/results still lower to `any`, and typed union/result pattern
    matching still needs a native carrier/type-switch ABI instead of the
    current staged dynamic coercion path.

# 2026-03-10 — Compiler native nullable value carriers (v12)
- Landed the first code-bearing union/nullable compiler slice in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_nullable_native.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
  - `v12/interpreters/go/pkg/compiler/generator_value_conversions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_ident.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_match_runtime_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_or_else.go`
  - `v12/interpreters/go/pkg/compiler/generator_exprs_lambda_cast_range.go`
  - `v12/interpreters/go/pkg/compiler/compiler_nullable_native_test.go`
- Behavior / architecture changes:
  - `?i32`, `?bool`, and `?String` now lower to native Go pointer carriers
    (`*int32`, `*bool`, `*string`) instead of `any`;
  - compiled wrappers/lambdas now convert those nullable carriers explicitly at
    dynamic boundaries via dedicated generated helpers instead of routing
    through `any`;
  - typed nullable `match` cases on static paths now lower through native nil
    checks plus direct payload dereference instead of `__able_try_cast(...)`;
  - `or {}` on native nullable values now branches on nil directly and unwraps
    the payload natively on success.
- Structural maintenance:
  - split conversion helpers out of `generator_collections.go` into
    `generator_value_conversions.go`;
  - split runtime/literal match tail helpers out of `generator_match.go` into
    `generator_match_runtime_types.go`;
  - both previously affected compiler files are back under the repo's
    1000-line limit.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return)$|TestCompilerDynamicBoundaryCallbackNullableConversion(Success|Failure)Markers$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NullableI32ParamAndMatchStayNative|NullableI32ReturnAndOrElseStayNative|StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_nullable_param|06_01_compiler_nullable_return|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundary(CallbackNullableConversion(Success|Failure)Markers|CallbackArrayConversion(Success|Failure)Markers)$' -count=1` (pass).
- Remaining gap:
  - this first native nullable pass landed the architecture and was then
    broadened later the same day; see the follow-on scalar-widening entry above
    for the current carrier surface.

# 2026-03-10 — Compiler union ABI kickoff audit (v12)
- Started the native union-lowering tranche by documenting the target ABI in:
  - `v12/design/compiler-union-abi.md`
  - `PLAN.md`
- Audit result captured:
  - current compiler lowering still maps `UnionTypeExpression` and
    `ResultTypeExpression` to `any`;
  - nullable lowering only keeps pointer/slice forms native today; nullable
    value types still fall back to `any`;
  - typed union/nullable pattern checks and wrapper coercions still rely on
    dynamic `bridge.MatchType(...)` / `__able_try_cast(...)`.
- First implementation order was fixed before code work starts:
  - native nullable value carriers first (`?i32`, `?bool`, `?String`);
  - then closed two-branch native unions over nominal types;
  - then native pattern dispatch over those carriers;
  - then wrapper/boundary adapters for those same shapes.
- Remaining gap:
  - no native union carrier code has landed yet; this tranche only established
    the implementation order and target ABI so follow-on compiler work can
    reduce `any` deliberately instead of continuing the staged fallback.

# 2026-03-10 — Compiler dynamic-helper reachability audit (v12)
- Added a stricter reachability audit for residual dynamic helpers in:
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_helper_reachability_test.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
- Behavior / enforcement changes:
  - static native array/struct compiled paths are now regression-audited to
    prove they do not call `__able_index`, `__able_index_set`,
    `__able_member_get`, `__able_member_set`, `__able_member_get_method`, or
    `__able_call_value`;
  - an explicit dynamic package/member/index example is regression-audited to
    prove the residual helper layer is still reachable when dynamic behavior is
    actually requested, specifically through `__able_member_get`,
    `__able_call_value`, and `__able_index`;
  - `__able_index` / `__able_index_set` now use the same normalized runtime
    array receiver shim as the member helpers.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StaticNativePathsAvoidDynamicHelperReachability|ExplicitDynamicPathsUseDynamicHelpers|NormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$|TestCompilerDynamicBoundaryCallbackArrayConversion(Success|Failure)Markers$' -count=1` (pass).
- Remaining gap:
  - this audit proves helper reachability from representative static vs dynamic
    cases, but it does not yet prove the same property for every remaining
    residual dynamic helper surface such as more exotic `member_set` paths.

# 2026-03-10 — Compiler residual array dynamic-helper audit (v12)
- Audited the remaining array dynamic helpers in:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_arrays.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_index_receiver_shim_regression_test.go`
- Behavior / architecture changes:
  - normalized `__able_index` and `__able_index_set` to use the shared
    `__able_runtime_array_value(...)` unwrapping helper instead of direct
    `*runtime.ArrayValue` pointer assertions;
  - confirmed current static array compiler slices continue to bypass
    `__able_index`, `__able_index_set`, `__able_member_get`, and
    `__able_member_set` on native `*Array` paths, leaving these helpers as
    residual dynamic/boundary machinery rather than the primary static lowering;
  - the remaining dynamic helper surface still preserves legacy handle-oriented
    behavior for interpreter-owned array values and `Array` struct instances,
    which is acceptable for now because those helpers are boundary-only.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(NormalizesArrayIndexReceiverUnwrap|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges)$|TestCompilerNormalizesArray(IndexReceiverUnwrap|MemberReceiverUnwrap)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$' -count=1` (pass).
- Remaining gap:
  - the helper layer still needs a broader reachability audit if we want to
    mechanically prove every remaining `__able_index` / `__able_member_*`
    emission site is tied only to explicit dynamic behavior.

# 2026-03-10 — Compiler array boundary helper narrowing (v12)
- Continued the compiler native-lowering work by tightening generated `Array`
  boundary adapters in:
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
- Behavior / architecture changes:
  - `__able_struct_Array_from` now reads existing `runtime.ArrayValue` handle
    state with `runtime.ArrayStoreState` instead of re-normalizing through
    `runtime.ArrayStoreEnsure`;
  - `__able_struct_Array_to` now routes through a shared
    `__able_struct_Array_runtime_value` helper and keeps plain
    `*runtime.ArrayValue` boundaries handle-free when no handle is already
    present on the compiled `Array`;
  - `__able_struct_Array_apply` now keeps raw `*runtime.ArrayValue` targets
    handle-free unless the raw target or compiled value already carries a
    handle, and allocates/synchronizes a handle only when applying back into an
    interpreter `Array` struct instance target that requires `storage_handle`
    semantics;
  - removed the legacy `Array_apply -> Array_to -> *runtime.StructInstanceValue`
    fallback for `Array` struct-instance targets; the Array-specific apply path
    now updates those targets directly.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|ArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range|06_01_compiler_match_patterns|06_01_compiler_assignment_patterns)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryCallbackArrayConversion(Success|Failure)Markers$' -count=1` (pass).
- Remaining gap:
  - explicit array dynamic runtime helpers outside struct conversion
    (`generator_render_runtime*.go`) still preserve legacy handle-oriented
    behavior and need a separate audit to confirm they are only used at true
    dynamic boundaries.

# 2026-03-10 — Compiler native array pattern lowering (v12)
- Continued the compiler native-lowering work by removing remaining
  `runtime.ArrayValue` rest-tail materialization from static compiled array
  destructuring in:
  - `v12/interpreters/go/pkg/compiler/generator_exprs_helpers.go`
  - `v12/interpreters/go/pkg/compiler/generator_match.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow.go`
  - `v12/interpreters/go/pkg/compiler/generator_controlflow_match.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
- Behavior / architecture changes:
  - static array pattern conditions and bindings now accept compiled `*Array`
    subjects directly instead of rejecting them and forcing runtime-only
    destructuring paths;
  - array rest bindings in both `match` and pattern assignment now synthesize a
    native compiled `*Array` tail via the compiler-owned `Array` carrier plus
    `__able_struct_Array_sync`, instead of materializing
    `&runtime.ArrayValue{Elements: ...}`;
  - pattern assignments on static subjects now match/bind against the native
    compiled subject type first and convert back to `runtime.Value` only at the
    assignment-expression result edge;
  - `match` expressions no longer blanket-convert struct subjects to
    `runtime.Value` before pattern dispatch, which keeps static `Array`
    destructuring native on direct compiled paths;
  - split match lowering into `generator_controlflow_match.go`, bringing
    `generator_controlflow.go` back under the repo's 1000-line limit.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|MatchArrayRestBindingStaysNative|PatternAssignmentArrayRestBindingStaysNative|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures/(06_01_compiler_match_patterns|06_01_compiler_assignment_patterns|05_02_array_nested_patterns|06_01_compiler_assignment_pattern_rest_mismatch)$' -count=1` (pass).
- Remaining gap:
  - compiled array boundaries still retain explicit `runtime.ArrayValue` /
    `ArrayStore*` adapters, which is acceptable only at explicit dynamic
    boundaries and still needs further narrowing.

# 2026-03-10 — Compiler native struct-local carrier correction (v12)
- Continued the compiler native-lowering work by removing declaration-time
  boxing of unannotated local struct values in:
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/compiler_struct_static_dispatch_test.go`
- Behavior / architecture changes:
  - unannotated local struct declarations no longer rewrite native compiled
    struct pointers back into `runtime.Value` just to preserve dispatch/identity;
  - static field access and static method dispatch now keep those locals as
    native Go struct pointers and use direct compiled field/method paths
    without `__able_struct_*_from` / `__able_struct_*_to` extract-writeback
    shims in the tested static cases;
  - direct compiled function calls on static struct paths are now regression-
    covered for native param passing, native returns, and native mutation-
    through-call behavior;
  - wrapper returns for native struct and array values now use explicit
    `__able_struct_*_to` conversion instead of routing through the broad
    `__able_any_to_value` helper;
  - the legacy `OriginGoType` runtime-value bridge remains available only for
    residual boxed cases, but static locals no longer depend on it by default.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata)$|TestCompilerExecFixtures/(06_01_compiler_bound_method_value|07_04_trailing_lambda_method_syntax|10_07_interface_default_chain|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata)$|TestCompilerExecFixtures/(06_01_compiler_bound_method_value|07_04_trailing_lambda_method_syntax|10_07_interface_default_chain|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass).
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata|ArrayWrapperUsesExplicitArrayBoundaryConverters|StructMethodStaticDispatch|StructFieldStaticAccess|DefaultImplSiblingDirectCall|StructFunctionParamAndReturnStayNative|StructMutationAcrossStaticFunctionCallStaysNative|StructWrapperReturnUsesExplicitStructConverter)$|TestCompilerExecFixtures/(06_01_compiler_bound_method_value|07_04_trailing_lambda_method_syntax|10_07_interface_default_chain|06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range)$' -count=1` (pass).
- Remaining gap:
  - the broader struct native-lowering contract is not complete yet; explicit
    dynamic-boundary conversions, params/returns crossing residual runtime
    carriers, and union lowering still need follow-through.

# 2026-03-10 — Compiler static array carrier correction (v12)
- Continued the compiler native-lowering reset by correcting the static `Array`
  path in:
  - `v12/interpreters/go/pkg/compiler/generator.go`
  - `v12/interpreters/go/pkg/compiler/generator_builtin_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_collections.go`
  - `v12/interpreters/go/pkg/compiler/generator_mono_array_intrinsics.go`
  - `v12/interpreters/go/pkg/compiler/generator_assignments.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `v12/interpreters/go/pkg/compiler/compiler_array_intrinsics_test.go`
- Behavior / architecture changes:
  - the compiler now synthesizes built-in `Array` struct metadata even when no
    source module defines `Array`, so array literals can lower to a native
    compiled carrier instead of immediately falling back to `runtime.ArrayValue`;
  - compiled `Array` values now keep the spec-visible metadata fields
    (`length`, `capacity`, `storage_handle`) and carry hidden native
    `Elements []runtime.Value` storage for static execution;
  - unannotated local `Array` declarations now stay as native compiled `*Array`
    values instead of being auto-boxed back into `runtime.Value`;
  - static array literal / `push` / `write_slot` / direct index assignment /
    `clear` lowering now mutates the native slice directly and synchronizes
    metadata after mutation.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompiler(WhileLoopFastPath|ArrayStructKeepsSpecFieldsAndNativeStorage|ArrayMutationsSyncMetadata)$|TestCompilerExecFixtures/(06_08_array_ops_mutability|06_12_02_stdlib_array_helpers|06_12_18_stdlib_collections_array_range|06_01_compiler_index_assignment|06_01_compiler_index_assignment_value)$' -count=1` (pass).
  - attempted `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -count=1`, but stopped the run after roughly 3 minutes because it exceeded the repo's stated quick-test expectation; no failures surfaced before cancellation.
- Remaining gap:
  - explicit compiled/interpreter array boundary adapters still use
    `runtime.ArrayValue` / `ArrayStore*`; that is acceptable only at the
    boundary, and this workstream remains open until those adapters are
    narrowed and the rest of the static nominal-value lowering contract is
    enforced.

# 2026-03-10 — Compiler native-lowering contract documented (v12)
- Recorded the active compiler direction in:
  - `v12/design/compiler-native-lowering.md`
  - `v12/design/compiler-monomorphization.md`
  - `v12/design/compiler-no-panic-flow-control.md`
  - `spec/TODO_v12.md`
  - `PLAN.md`
- Direction reset captured:
  - static compiled code should lower Able constructs to native Go carriers wherever semantics are representable;
  - arrays should use compiler-native Go array-backed storage on static paths, not `runtime.ArrayValue`, `ArrayStore*`, or kernel `storage_handle` plumbing;
  - structs should remain native Go structs/pointers on static paths rather than being auto-boxed to `runtime.Value`;
  - unions should target generated Go interfaces plus native variant carriers rather than stopping at `any`;
  - compiled flow should avoid IIFEs and `panic`/`recover`, using explicit control-result signaling for non-local jumps and exceptions.
- Current state note:
  - the recent compiler-side `Array -> Elements []runtime.Value` rewrite is being treated as an in-flight experiment under audit, not as an accepted final architecture.
- Next steps:
  - audit/back out the runtime-backed static array hybrid in the staged compiler diff;
  - define the compiler-native array ABI for static code;
  - remove struct-local boxing and design native union lowering plus explicit control-result propagation.

# 2026-03-05 — Dynamic array first-append pre-grow (v12)
- Reduced hot append-loop allocation churn in:
  - `v12/interpreters/go/pkg/runtime/array_store.go`
- Behavior/perf changes:
  - in `ArrayStoreWrite` dynamic append path (`index == len`), pre-grow empty arrays to capacity `4` before the first append.
  - this avoids the extra `cap=1` and `cap=2` backing-slice allocation steps on common push-heavy loops while preserving existing sparse-write and growth semantics.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestArrayStoreDynamicCapacityGrowthAmortized|TestArrayStoreDynamicSparseWritePreservesNilGap|TestArrayStoreMonoBoolRoundTripAndDynamicFallback|TestArrayStoreMonoI64RoundTripAndDynamicFallback' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtureParity/07_10_bytecode_quicksort_hotloop' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `53.803s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`, fixed `-benchtime=50x` A/B):
  - before pre-grow: `~106.93-107.12 B/op`, `~49-50 allocs/op`.
  - after pre-grow: `~106.87-106.98 B/op`, `~47 allocs/op`.
  - wall-time remained in the same host-noisy band (`~32.7-33.5ms/op` in the comparable runs).
  - profile snapshot: `/tmp/able-bytecode-quicksort.after20.mem.out`.
- Next steps (bytecode perf focus):
  - reduce `runtime.allocm` / scheduler allocation churn (`SerialExecutor` + bytecode call loop interaction).
  - reduce `sync.(*Pool).pinSlow` pressure from VM/context pooling paths (`acquireBytecodeVM`, native call context reuse).
  - continue trimming `ArrayStoreWrite` alloc-space share (now still top flat allocator in quicksort hotloop profiles).

# 2026-03-05 — Quicksort hotloop benchmark steady-state warmup isolation (v12)
- Improved benchmark signal quality in:
  - `v12/interpreters/go/pkg/interpreter/bytecode_quicksort_hotloop_bench_test.go`
- Behavior/perf harness changes:
  - kept memprofile sampling suspended through parser/load/typecheck/bootstrap and an explicit untimed `main()` warmup call, then resumed sampling immediately before the timed benchmark loop.
  - added a pre-timer warmup invocation so first-call cache/bootstrap costs are excluded from steady-state `ns/op`, `B/op`, and alloc profile signals.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^$' -bench '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchmem -count=5` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^$' -bench '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchmem -count=1 -memprofile /tmp/able-bytecode-quicksort.after16.mem.out` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtureParity/07_10_bytecode_quicksort_hotloop' -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before harness tweak: `32.37ms/op`, `107802 B/op`, `53 allocs/op` (single-run snapshot).
  - after harness tweak: `31.61ms/op`, `107281 B/op`, `50 allocs/op` (single-run snapshot).
  - latest alloc-space profile (`/tmp/able-bytecode-quicksort.after16.mem.out`) is now dominated by steady-state loop paths (`runtime.ArrayStoreWrite`, call dispatch), with one-time small-int cache init no longer at the top.

# 2026-03-05 — Int-div-cast bytecode opcode + dynamic array append write fast path (v12)
- Reduced bytecode hotloop allocation pressure in:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering_cast_div.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_binary_fastpath_test.go`
  - `v12/interpreters/go/pkg/runtime/array_store.go`
  - `v12/interpreters/go/pkg/runtime/array_store_mono_test.go`
- Behavior/perf changes:
  - added specialized lowering for `(<expr> / <expr>) as <int>` to emit new `bytecodeOpBinaryIntDivCast` instead of generic `Binary("/")` followed by `Cast`.
  - new VM opcode executes a guarded fast path for integer operands in safe range (including a `r==2 && l>=0` shift shortcut), then falls back to generic `/` + cast semantics when outside the guardrails.
  - optimized dynamic `ArrayStoreWrite` for append writes (`index == len`) to append the value directly (skipping nil-fill then overwrite), while preserving sparse-write gap semantics for `index > len`.
  - added regression coverage for lowering/parity/division-by-zero on the new opcode and sparse dynamic write gap behavior.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_Binary(IntDivCastFastPathParity|IntDivCastFloatFallbackParity|IntDivCastDivisionByZeroParity|LoweringEmitsIntegerDivCastOpcode)|TestExecFixtureParity/07_10_bytecode_quicksort_hotloop' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestArrayStoreDynamicCapacityGrowthAmortized|TestArrayStoreDynamicSparseWritePreservesNilGap|TestArrayStoreMonoBoolRoundTripAndDynamicFallback|TestArrayStoreMonoI64RoundTripAndDynamicFallback' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `52.699s`; one prior run observed a transient fixture parity mismatch that did not reproduce in targeted reruns and subsequent full run).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `33.75ms/op`, `155178 B/op`, `1823 allocs/op`.
  - after tranche: `32.51ms/op`, `112193 B/op`, `49 allocs/op` (single-run memprofile capture); repeated runs showed wall-time variance but stable low alloc counts (`~48-52 allocs/op`).
  - memprofile hotspot shift: `evaluateDivision` dropped out of top alloc nodes for this benchmark shape; profile snapshot at `/tmp/able-bytecode-quicksort.after12.mem.out`.

# 2026-03-05 — Array metadata boxing + primitive receiver cache probe (v12)
- Reduced bytecode hotloop allocation pressure in:
  - `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
- Behavior/perf changes:
  - added a dedicated cached boxed-`u64` path for `__able_array_size` / `__able_array_capacity` return values so repeated array metadata reads avoid per-call integer boxing allocations in hot loops.
  - kept the existing bytecode global small-int cache scope unchanged (`i32/i64/isize`) and moved array-metadata caching to a dedicated path to avoid inflating bytecode cache initialization overhead.
  - added an early primitive-receiver bound-method cache probe in `resolveMethodFromPool(...)` and relaxed scope-name cache gating (still guarded by impl-context checks) so hot method calls can reuse cached bound methods even when a callable of the same name exists in scope.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_BoxedSmallIntValueCache|TestBytecodeVM_BoxedIntegerValueDynamicCache|TestResolveMethodFromPool_BoundMethodCacheInvalidatesWithMethodCache|TestExecFixtureParity/07_10_bytecode_quicksort_hotloop|TestArrayBuiltins' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `53.045s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `33.60ms/op`, `314920 B/op`, `5822 allocs/op`.
  - after tranche: `33.75ms/op`, `155178 B/op`, `1823 allocs/op`.
  - memprofile alloc-space dropped from `~26.24MB` to `~19.02MB` for this benchmark shape.

# 2026-03-05 — Call-dispatch arg/context pooling + typed-declare slot coverage (v12)
- Reduced bytecode hotloop allocation pressure in:
  - `v12/interpreters/go/pkg/interpreter/call_helpers.go`
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_operations_arithmetic.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_operations_fast.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering_helpers.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_slot_analysis.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_store.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_typed_pattern_slot_test.go`
- Behavior/perf changes:
  - call dispatch now reuses mutable arg backing storage for bound-receiver injection (`prependReceiverCallArgs`) instead of always allocating a fresh `[]Value`.
  - native call dispatch now reuses pooled `runtime.NativeCallContext` objects per interpreter, removing per-call context allocation churn on hot native/member call paths.
  - int64 arithmetic/div-mod hot paths now use `ensureFitsInt64Type(...)` directly and defer `getIntegerInfo(...)` lookups until big-int fallback is actually required.
  - bytecode slot analysis/lowering now accepts typed identifier declarations (`name: T := expr`) as slot-eligible and lowers them through slot stores.
  - added VM typed-slot store coercion/mismatch handling so typed declaration semantics are preserved on slot paths; typed `=`/compound typed-pattern assignments remain on `AssignPattern` paths to preserve interface coercion/fallback behavior.
  - fixed parity regression in `10_11_interface_generic_args_dispatch` by keeping typed `=` assignment on `AssignPattern`.
- Regression coverage:
  - `TestBytecodeVM_TypedIdentifierDeclarationUsesSlotLowering`
  - `TestBytecodeVM_TypedIdentifierMismatchReturnsErrorValue`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_TypedIdentifierDeclarationUsesSlotLowering|TestBytecodeVM_TypedIdentifierMismatchReturnsErrorValue|TestBytecodeVM_AssignmentPattern|TestBytecodeVM_CompoundAssignmentPattern' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtureParity/10_11_interface_generic_args_dispatch' -count=1 -v` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `53.031s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `35.68ms/op`, `970300 B/op`, `30939 allocs/op`.
  - after tranche: `34.84ms/op`, `806269 B/op`, `28916 allocs/op`.
  - memprofile total alloc-space dropped from `~52.91MB` to `~41.71MB`; `runtime.NewEnvironment` fell out of top alloc-space nodes in this benchmark shape.

# 2026-03-05 — Integer div/mod boxed-result fast path (v12)
- Reduced hot `%`/`//` result boxing churn in:
  - `v12/interpreters/go/pkg/interpreter/interpreter_operations_arithmetic.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_operations_fast.go`
- Behavior/perf changes:
  - added an int64-first fast path in `evaluateDivMod(...)` that computes Euclidean quotient/remainder directly and returns cached boxed integers for supported hot kinds (`i32/i64/isize`) via existing boxed-int caches.
  - fallback div/mod paths now opportunistically return cached boxed integers for small int results instead of always returning fresh value boxing.
  - mirrored these result-boxing optimizations in `evaluateDivModFast(...)` so bytecode fast dispatch and generic arithmetic dispatch stay aligned.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(DivModEuclidean|DivModStructResult|DivisionByZeroDiagnostics|CallFunction_DoesNotMutateCallerArgsOnCoercion|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|ResolveMethodFromPool_BoundMethodCacheInvalidatesWithMethodCache)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `54.914s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `~1.352MB/op`, `~42.93k allocs/op`.
  - after tranche: `~1.257MB/op`, `~40.93k allocs/op`.
  - profile total alloc-space dropped to ~`22.86MB` for the latest `-benchtime=5x` hotloop shape.
- Bench harness spot-check:
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort`:
    - quicksort avg `4.260s`, `gc_avg=37.5` (`v12/tmp/perf/bench-suite-20260305T231745Z.json`).
  - macro wall-time remained within the current noise band while micro-allocation metrics improved.

# 2026-03-05 — Mutability-aware bytecode call dispatch for coercion paths (v12)
- Reduced call/invoke coercion-copy churn in:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/call_args_mutation_test.go`
- Behavior/perf changes:
  - introduced internal `callCallableValueMutable(...)` and shared implementation `callCallableValueWithMutability(...)`.
  - bytecode VM call sites now route through the mutable path so callee-parameter coercion can update borrowed arg slices in place (instead of always cloning on first coercion).
  - public/external call entry points remain immutable by default (`callCallableValue(...)` keeps `argsMutable=false`), preserving host caller argument stability.
  - partial-call chaining remains protected: recursive partial dispatch forces immutable handling to avoid mutating stored `PartialFunctionValue.BoundArgs`.
- Regression coverage:
  - added `TestCallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion`.
  - existing `TestCallFunction_DoesNotMutateCallerArgsOnCoercion` remains green for public API safety.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(CallFunction_DoesNotMutateCallerArgsOnCoercion|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallDispatch|FixtureParityStringLiteral/errors/ufcs_static_method_not_found|ResolveMethodFromPool_BoundMethodCacheInvalidatesWithMethodCache|BytecodeVM_BoundMethodInlineCallStatsHit)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `58.298s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `~1.448MB/op`, `~46.93k allocs/op`.
  - after tranche: `~1.352MB/op`, `~42.93k allocs/op`.
  - wall-time stayed in the same noisy band (`~34.9-38.4ms/op` in sampled runs).
- Bench harness snapshot (bytecode mode):
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort`:
    - quicksort avg `4.255s`, `gc_avg=35.0` (`v12/tmp/perf/bench-suite-20260305T230945Z.json`).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,binarytrees,matrixmultiply,quicksort,sudoku,i_before_e` (`v12/tmp/perf/bench-suite-20260305T231005Z.json`):
    - completed: `quicksort 4.080s (gc=36)`, `i_before_e 4.040s (gc=35)`;
    - timed out in this snapshot window: `fib`, `binarytrees`, `matrixmultiply`, `sudoku`.

# 2026-03-05 — Pointer-receiver bound-method cache in method resolution (v12)
- Reduced bound-wrapper churn in:
  - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution_cache_test.go`
- Behavior/perf changes:
  - added a capped bound-method cache (`boundMethodCacheMaxEntries=2048`) keyed by pointer receiver identity + method name + interface filter + `allowInherent` gate.
  - cache is consulted in `resolveMethodFromPool(...)` only when safe (no callable name in scope, no active `implMethodContext`) and currently targets pointer receivers.
  - cache entries are cleared together with method-cache invalidation (`invalidateMethodCache`) to preserve dispatch correctness when methods are redefined.
  - added regression coverage proving invalidate clears the cache and updated method targets are observed after redefinition.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(ResolveMethodFromPool_BoundMethodCacheInvalidatesWithMethodCache|FunctionScopeFilter|CallDispatch|FixtureParityStringLiteral/errors/ufcs_static_method_not_found)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `51.301s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `~1.512MB/op`, `~48.93k allocs/op`.
  - after tranche: `~1.448MB/op`, `~46.93k allocs/op`.
  - profile signal: `resolveMethodFromPool` dropped to low single-sample presence (`~512KB` cum in the latest `-benchtime=5x` profile shape) versus prior multi-MB prominence.

# 2026-03-05 — Hotloop harness setup isolation + type-info key cache (v12)
- Reduced quicksort hotloop setup/profile noise and method-cache key churn in:
  - `v12/interpreters/go/pkg/interpreter/bytecode_quicksort_hotloop_bench_test.go`
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `v12/interpreters/go/pkg/interpreter/function_overloads.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_interface_lookup.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_type_info_cache_test.go`
- Behavior/perf changes:
  - benchmark harness now suspends `runtime.MemProfileRate` during one-time fixture/load/typecheck bootstrap and restores sampling before the timed `main()` call loop; this keeps memprofiles focused on runtime hotloop allocations.
  - `mergePartialCallArgs(...)` now reuses caller slices when one side is empty, removing avoidable merge-buffer allocations in partial-call chaining.
  - added `functionOverloadsView(...)` and wired call dispatch to reuse existing overload slices for `*runtime.FunctionOverloadValue` targets instead of always flattening.
  - added `cachedTypeInfoName(...)` + interpreter cache map (`typeInfoCacheKey -> string`) and switched `findMethodCached(...)` to use it, eliminating repeated `typeInfoToString(...)` allocations for hot generic receiver keys.
  - added allocation regression guard: `TestCachedTypeInfoNameAvoidsRepeatedAllocationsForCommonGenericTypes`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(CachedTypeInfoNameAvoidsRepeatedAllocationsForCommonGenericTypes|CanonicalTypeNamesUsesAliasBaseWithoutASTExpansion|CallDispatch|CallFunction_DoesNotMutateCallerArgsOnCoercion|FixtureParityStringLiteral/errors/ufcs_static_method_not_found)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `51.473s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`):
  - before tranche: `~36.08-36.94ms/op`, `~1.608MB/op`, `~56.95k allocs/op`.
  - after tranche: `~36.01-36.89ms/op`, `~1.512MB/op`, `~48.93k allocs/op`.
  - memprofile total alloc-space dropped from ~`24.6MB` to ~`19.4MB` (`-~21%`) for the `-benchtime=5x` hotloop profile shape.

# 2026-03-05 — Method-resolution accumulator single-candidate fast path (v12)
- Reduced hot method/call dispatch allocation churn in:
  - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
- Behavior/perf changes:
  - refactored `methodResolutionAccumulator` to keep single function/native candidates in dedicated fields and only allocate/promote to slices when a second distinct candidate appears.
  - preserved existing ambiguity/dedup semantics (`Ambiguous overload` behavior and native-key deduping) while avoiding per-call slice appends on common single-candidate paths.
  - this primarily targets the `resolveMethodFromPool` path that feeds hot `callCallableValue` dispatch.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(FunctionScopeFilter|CallDispatch|FixtureParityStringLiteral/errors/ufcs_static_method_not_found|BytecodeVM_BoundMethodInlineCallStatsHit)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(IndexMethodCacheTracksArrayElementType|CallNameScopeCacheInvalidatesOnLocalRebind|BoxedIntegerValueDynamicCache|StatsMemberMethodCacheCounters)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `53.814s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`, 3 runs):
  - `~38.03-41.52ms/op` (host-noisy wall-time),
  - `~1.61MB/op`,
  - `~56.95k allocs/op`.
  - allocation profile improved materially versus immediate prior tranche (`~1.64MB/op`, `~60.95k allocs/op`).

# 2026-03-05 — Direct function call-dispatch fast path (v12)
- Reduced call-dispatch overhead in:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
- Behavior/perf changes:
  - `callCallableValue(...)` now detects direct `*runtime.FunctionValue` call targets (including function targets inside `BoundMethodValue`) and bypasses `functionOverloads(...)` flattening/select logic on this common path.
  - preserves partial-application behavior and mismatch diagnostics by:
    - running a pre-invoke mismatch check for `TypeQualified` direct functions, and
    - running mismatch mapping on invoke errors for other direct functions.
  - native and overload-value paths are unchanged.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestFixtureParityStringLiteral/errors/ufcs_static_method_not_found' -count=1 -v` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(CallDispatch|CallCallableValue_NativeBoundMethodPartialDoesNotDoubleInjectReceiver|BytecodeVM_BoundMethodInlineCallStatsHit|BytecodeVMExecStringInterpolationReusesPartsBuffer)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(IndexMethodCacheTracksArrayElementType|CallNameScopeCacheInvalidatesOnLocalRebind|BoxedIntegerValueDynamicCache)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `53.782s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`, 3 runs):
  - `~37.48-38.74ms/op`,
  - `~1.64MB/op`,
  - `~60.95k allocs/op`.
  - this pass is primarily structural/micro-allocation-focused; wall-time remained in the current noise band on this host.

# 2026-03-05 — Bytecode literal handler split + interpolation scratch reuse (v12)
- Reduced bytecode run-loop scratch allocations and trimmed oversized run-loop file:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_literals.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_stack_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_types.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_literals_test.go`
- Behavior/perf/maintainability changes:
  - moved `StringInterpolation`/`ArrayLiteral` bytecode handlers out of `runResumable(...)` into dedicated helpers.
  - added reusable VM interpolation-part scratch buffer (`stringInterpParts`) with per-call clear, removing per-op `make([]runtime.Value, argCount)` for interpolation.
  - switched array literal operand extraction to stack-segment copy + truncate helper path.
  - split `Dup`/`Pop` handler code into dedicated helpers; `bytecode_vm_run.go` is now `996` lines (back under 1000-line guardrail).
- Regression coverage:
  - `TestBytecodeVMExecStringInterpolationReusesPartsBuffer`
  - `TestBytecodeVMExecArrayLiteralCopiesStackSegment`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM(ExecStringInterpolationReusesPartsBuffer|ExecArrayLiteralCopiesStackSegment|ReleaseCompletedRunFramesReleasesActiveSlots|FinishRunResumableReleasesUnwoundCallFrames)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(Async|IndexMethodCacheTracksArrayElementType|CallNameScopeCacheInvalidatesOnLocalRebind|BoxedIntegerValueDynamicCache)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `60.219s` on this run).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`, 3 runs):
  - `~35.11-39.10ms/op`,
  - `~1.64MB/op`,
  - `~60.95k allocs/op`.

# 2026-03-05 — Bytecode slot-frame finalization cleanup + pooled top-level slot setup (v12)
- Reduced bytecode slot-frame allocation churn in:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run_finalize.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_frames_finalize_test.go`
- Behavior/perf changes:
  - switched `invokeFunction(...)` bytecode entry frame setup from `make([]runtime.Value, ...)` to `vm.acquireSlotFrame(...)`.
  - `finishRunResumable(...)` now returns active slot frames to the slot-frame pool on non-yield exits (normal completion and error unwind).
  - error unwind now releases callee slot frames as frames are popped, then releases the final active frame at run completion.
- Regression coverage:
  - `TestBytecodeVMReleaseCompletedRunFramesReleasesActiveSlots`
  - `TestBytecodeVMFinishRunResumableReleasesUnwoundCallFrames`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVMReleaseCompletedRunFramesReleasesActiveSlots|TestBytecodeVMFinishRunResumableReleasesUnwoundCallFrames|TestBytecodeVM_BoxedIntegerValueDynamicCache|TestBytecodeVM_IndexMethodCacheTracksArrayElementType|TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -count=1` (pass; `59.021s`).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (`BenchmarkBytecodeQuicksortHotloopRuntime`, 3 runs):
  - `~35.33-36.75ms/op`,
  - `~1.64MB/op`,
  - `~60.95k allocs/op`.

# 2026-03-05 — Bytecode index-cache key token compaction (v12)
- Reduced index-cache key overhead in:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_index_cache.go`
- Behavior/perf changes:
  - replaced array-element cache key strings (`"i32"`, `"String"`, etc.) with compact numeric tokens (`uint16`) for bytecode index-method cache keys.
  - preserved existing element-type invalidation semantics (including empty-array and impl-change cases); unsupported element kinds remain non-cacheable as before.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(IndexMethodCacheTracksArrayElementType|IndexSetCompoundCacheInvalidatesWhenImplAppears|StatsMemberMethodCacheCounters|CallNameDotFallbackUsesMemberMethodCache|BoxedIntegerValueDynamicCache|NativeCallArgsSliceStaysStable)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `69.158s` on this loaded host run).
- Perf signal checks:
  - quicksort hotloop allocation profile remained in the improved post-boxing band (`~1.83MB/op`, `~64.9k allocs/op`).
  - wall-time readings were noisy under concurrent host load during this pass; no deterministic `ns/op` claim is made from this micro-optimization alone.

# 2026-03-05 — Dynamic boxed-int cache for bytecode add/sub fast paths (v12)
- Reduced integer boxing churn in:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`
- Behavior/perf changes:
  - added a bounded dynamic boxed-int cache (`bytecodeIntBoxDynamicCacheLimit=32768`) for out-of-range `i32`/`i64`/`isize` values while preserving the existing fixed small-int cache.
  - wired specialized bytecode add/sub fast paths (`addIntegerSameTypeFast`, `subtractIntegerSameTypeFast`) to use the new boxed-int helper.
  - kept `bytecodeBoxedSmallIntValue` semantics unchanged; added regression coverage for dynamic cache behavior and cached no-allocation reuse.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(BoxedSmallIntValueCache|BoxedIntegerValueDynamicCache|NativeCallArgsSliceStaysStable)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `50.927s`).
- Perf signal checks:
  - pre-tranche quicksort hotloop: `~35.0ms/op`, `~2.02MB/op`, `~68.8k allocs/op`.
  - after this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`):
    - `~35.0-35.6ms/op`,
    - `~1.83MB/op`,
    - `~64.9k allocs/op`.

# 2026-03-05 — Native arg-borrow + small-int cache expansion (v12)
- Reduced bytecode call/boxing allocation pressure in:
  - `v12/interpreters/go/pkg/runtime/values.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
- Behavior/perf changes:
  - added `BorrowArgs bool` to `runtime.NativeFunctionValue`; bytecode fallback now only clones arg slices for native/dynamic targets that require stable arg storage.
  - marked hot array runtime natives (`__able_array_*`, `Array.new`, `array.iterator`) as borrow-safe, removing per-call fallback arg cloning on these sites.
  - expanded pre-boxed integer cache range from `[-256, 4096]` to `[-256, 16384]` for `i32/i64/isize` boxed reuse.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_NativeCallArgsSliceStaysStable|TestBytecodeVM_BoxedSmallIntValueCache' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `51.062s`).
- Perf signal checks:
  - pre-tranche quicksort hotloop (latest prior state): `~34.9-35.2ms/op`, `~2.15MB/op`, `~72.9k allocs/op`.
  - after this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`):
    - `~35.03-35.25ms/op`,
    - `~2.02MB/op`,
    - `~68.8k allocs/op`.
  - alloc-space hotspot `copyCallArgs` dropped out of top allocators in latest profile; current dominant runtime allocators are integer arithmetic (`addIntegerSameTypeFast`), method resolution, and environment creation.

# 2026-03-05 — Bound-method value-form allocation reduction (v12)
- Reduced bound-method heap churn in:
  - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_member_cache.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_members_dynamic.go`
- Behavior/perf changes:
  - switched hot method-resolution/member-cache/interface-member return paths from pointer-form bound wrappers to value-form runtime wrappers:
    - `runtime.BoundMethodValue` (value)
    - `runtime.NativeBoundMethodValue` (value)
  - preserves call semantics (dispatch already supports both value + pointer forms) while reducing per-call heap allocation pressure.
- Test compatibility update:
  - adjusted ordering test to accept both bound-wrapper forms:
    - `v12/interpreters/go/pkg/interpreter/interpreter_strings_ordering_test.go`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `49.845s`).
  - targeted cache/string/env tests remained passing in earlier per-tranche runs.
- Perf signal checks:
  - `BenchmarkBytecodeQuicksortHotloopRuntime` remained stable in wall time (`~35.10-35.98ms/op` band) with maintained low allocation profile (`~2.18MB/op`, `~74.9k allocs/op`).
  - alloc-space profile (`/tmp/bytecode-quicksort-hotloop.mem.after-boundvalue.out`) shows `resolveMethodFromPool` flat allocation down from recent `~3.5MB` to `~2.5MB`.

# 2026-03-05 — `parseTypeExpression` generic arg copy removal (v12)
- Reduced generic type parse allocation churn in:
  - `v12/interpreters/go/pkg/interpreter/impl_resolution_types.go`
- Behavior/perf changes:
  - `parseTypeExpression(...)` now reuses `*ast.GenericTypeExpression.Arguments` slices directly (treated as immutable in runtime resolution paths) instead of copying with `append([]TypeExpression(nil), ...)` on every parse.
  - preserves semantics while cutting repeated generic-argument copy overhead in method lookup/type coercion paths.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestTypeExpressionToStringStableForms|TestTypeInfoToStringStableForms|TestCanReuseFunctionClosureEnvForBytecode|TestBytecodeVM_CallNameDotFallbackUsesMemberMethodCache' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `51.611s`).
- Perf signal checks:
  - prior quicksort hotloop sample after type-string pass: `~35.19-35.70ms/op`, `~2.30MB/op`, `~82.3k allocs/op`.
  - after this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`): `~35.94-36.61ms/op`, `~2.18MB/op`, `~74.9k allocs/op`.

# 2026-03-05 — Type stringification allocation pass for method/type hot paths (v12)
- Reduced type formatting churn in:
  - `v12/interpreters/go/pkg/interpreter/impl_resolution_types.go`
- Behavior/perf changes:
  - rewrote `typeExpressionToString(...)` to use a recursive `strings.Builder` writer (`appendTypeExpressionString`) instead of `fmt.Sprintf` + `strings.Join` intermediate slices.
  - rewrote `typeInfoToString(...)` to use builder-based rendering for generic `Type<...>` signatures.
  - this keeps textual output stable while reducing per-call allocation pressure in hot dispatch/type paths (method-cache keys, diagnostics, interface checks).
- Regression coverage:
  - added `TestTypeExpressionToStringStableForms`
  - added `TestTypeInfoToStringStableForms`
  - file: `v12/interpreters/go/pkg/interpreter/impl_resolution_types_string_test.go`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestTypeExpressionToStringStableForms|TestTypeInfoToStringStableForms|TestCanReuseFunctionClosureEnvForBytecode|TestBytecodeVM_CallNameDotFallbackUsesMemberMethodCache' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `52.581s`).
- Perf signal checks:
  - prior quicksort hotloop sample after env-reuse tranche: best around `35.74ms/op`, `~2.40MB/op`, `~86.3k allocs/op`.
  - after this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`): best sample `35.19ms/op`, `~2.30MB/op`, `~82.3k allocs/op`.
  - latest alloc-space profile (`/tmp/bytecode-quicksort-hotloop.mem.after-type-string.out`) no longer shows `typeInfoToString` in top alloc-space nodes.

# 2026-03-05 — Environment allocation reduction for bytecode call hot paths (v12)
- Reduced hot call-path environment allocation pressure in:
  - `v12/interpreters/go/pkg/runtime/environment.go`
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
- Behavior/perf changes:
  - fixed `runtime.NewEnvironment(...)` child-scope construction so it no longer allocates a throwaway `atomic.Bool` when inheriting `threadMode` from a parent environment.
  - added conservative bytecode call fast-path env reuse in `invokeFunction(...)`: when a function has slot-enabled bytecode with `needsEnvScopes=false`, no generics, and no type-arg binding requirement, invocation now reuses `fn.Closure` instead of allocating a per-call child environment.
  - this keeps tree-walker semantics unchanged and only applies to bytecode functions whose frame analysis guarantees no call-local env bindings are required.
- Regression coverage:
  - added `TestCanReuseFunctionClosureEnvForBytecode`
    (`v12/interpreters/go/pkg/interpreter/eval_expressions_calls_env_reuse_test.go`).
  - added `TestEnvironmentChildReusesParentThreadModePointer` and `TestEnvironmentNewChildAllocationCount`
    (`v12/interpreters/go/pkg/runtime/environment_test.go`).
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestEnvironment' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestCanReuseFunctionClosureEnvForBytecode|TestBytecodeVM_CallNameDotFallbackUsesMemberMethodCache|TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind|TestBytecodeVM_LoadNameScopeCacheInvalidatesOnLocalAssign' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `53.436s`).
- Perf signal checks:
  - prior quicksort hotloop sample: `~36.56-38.39ms/op`, `~2.73MB/op`, `~96.3k allocs/op`.
  - after this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`): best sample `35.74ms/op`, `2.40MB/op`, `~86.3k allocs/op`.
  - alloc-space profile snapshot (`/tmp/bytecode-quicksort-hotloop.mem.after-env-reuse.out`) now shows `runtime.NewEnvironment` down to `~4.50MB` flat (from prior `~21.50MB` in the same benchmark shape).

# 2026-03-05 — Lookup miss-path churn reduction (`Get` -> `Lookup`) for hot type/call resolution (v12)
- Reduced environment lookup miss-allocation overhead on hot bytecode dispatch/type paths:
  - `v12/interpreters/go/pkg/interpreter/definitions.go`
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
- Behavior/perf changes:
  - `canonicalTypeName(...)` now probes via `env.Lookup(...)` instead of `env.Get(...)` when canonicalizing simple type names; this avoids allocating miss errors for common primitive/non-bound type names during repeated type canonicalization/matching.
  - direct identifier call resolution in treewalker call dispatch now uses `Lookup` probing for primary and dotted-head resolution, constructing undefined-variable errors only on true terminal misses.
  - `evaluateExternFunctionBody` existence-check path now uses `Lookup` probe semantics.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(BytecodeVM_CallNameDotFallbackUsesMemberMethodCache|CanonicalTypeNamesUsesAliasBaseWithoutASTExpansion|ArrayBuiltins|BytecodeVM_StatsMemberMethodCacheCounters)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `52.772s`).
- Perf signal checks:
  - prior tranche baseline: `BenchmarkBytecodeQuicksortHotloopRuntime` at roughly `~36.98-37.23ms/op`, `~2.97MB/op`, `~108.3k allocs/op`.
  - after lookup pass:
    - `~36.56-38.39ms/op`,
    - `~2.73MB/op`,
    - `~96.3k allocs/op`.
  - alloc-space profile (`/tmp/bytecode-quicksort-hotloop.mem.after-lookup-pass.out`) dropped to ~`108.97MB` total, with lower `fmt.Errorf`/`Environment.Get` pressure and `resolveMethodFromPool`/`callCallableValue` remaining primary optimization targets.

# 2026-03-05 — Array/member dispatch allocation pass + dotted `CallName` cache reuse (v12)
- Reduced hot bytecode member-dispatch allocations across array/member/type-alias/type-expression paths:
  - `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_type_info.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - `v12/interpreters/go/pkg/interpreter/imports.go`
  - `v12/interpreters/go/pkg/interpreter/eval_statements.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_definitions.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
- Behavior/perf changes:
  - `arrayMember` now reuses cached boxed small-int runtime values for `storage_handle`, `length`, and `capacity` when values are within cache range, avoiding repeated `runtime.NewSmallInt` allocations on hot array-member reads.
  - `canonicalTypeNames(...)` now caches alias-base expansion results (`typeAliasBaseCache`) and invalidates on alias writes/import alias rebinding through centralized `setTypeAlias(...)`.
  - type inference now caches hot generic type-expression instances for `Array<T>` (wildcard/simple element shapes) and `Iterator<_>` to avoid repeated `ast.NewGenericTypeExpression` churn during method/type matching.
  - dotted `CallName` fallback (`head.tail`) now reuses the existing bytecode member-method callsite cache (`lookupCachedMemberMethod`/`storeCachedMemberMethod`) before resolving through full member dispatch.
- Regression coverage:
  - added `TestTypeExpressionFromValueCachesArrayAndIteratorGenerics`
    (`v12/interpreters/go/pkg/interpreter/interpreter_type_info_cache_test.go`).
  - added `TestBytecodeVM_CallNameDotFallbackUsesMemberMethodCache`
    (`v12/interpreters/go/pkg/interpreter/bytecode_vm_scope_lookup_cache_test.go`).
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(CallNameDotFallbackScopeCacheInvalidatesOnHeadRebind|CallNameDotFallbackUsesMemberMethodCache|StatsMemberMethodCacheCounters|CallNameScopeCacheInvalidatesOnLocalRebind|LoadNameScopeCacheInvalidatesOnLocalAssign)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(TypeExpressionFromValueCachesArrayAndIteratorGenerics|TypeExpressionFromValueCachesStructAndHostHandleNames|CanonicalTypeNamesUsesAliasBaseWithoutASTExpansion|BytecodeVM_BoxedSmallIntValueCache|ArrayBuiltins)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `55.880s`).
- Perf signal checks:
  - baseline before this tranche (`BenchmarkBytecodeQuicksortHotloopRuntime`): `~38.03-38.57ms/op`, `~4.03MB/op`, `~134.3k allocs/op`.
  - after array/member/type-expression caching + dotted callname cache reuse:
    - `~36.97-37.23ms/op`,
    - `~2.97MB/op`,
    - `~108.3k allocs/op`.
  - alloc-space profile dropped major `arrayMember` small-int allocation hotspots and removed `ast.NewGenericTypeExpression` from top alloc-space nodes in latest snapshot.

# 2026-03-05 — UFCS scope-filter + native call-dispatch allocation pass (v12)
- Method-resolution allocation reduction:
  - replaced per-call UFCS scope map construction with allocation-light filter state:
    - `functionScopeFilterFromValue(...)`
    - `functionScopeFilter.contains(...)`
  - wired through `resolveMethodFromPool(...)` and `selectUfcsCallable(...)`.
  - file: `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`.
- Call-dispatch allocation reduction:
  - removed native pointer-escape churn in `callCallableValue(...)` by using value-form native tracking (`native` + `hasNative`) and normalizing bound-native partial targets to existing method values.
  - switched native call-context construction to stack form (`ctx := runtime.NativeCallContext{...}`).
  - file: `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`.
- Regression coverage:
  - added `TestCallCallableValue_NativeBoundMethodPartialDoesNotDoubleInjectReceiver`
    (`v12/interpreters/go/pkg/interpreter/call_callable_native_bound_partial_test.go`).
  - added `TestFunctionScopeFilterSingle`, `TestFunctionScopeFilterOverloads`, and `TestFunctionScopeFilterDisabledPassThrough`
    (`v12/interpreters/go/pkg/interpreter/function_scope_filter_test.go`).
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'Test(CallCallableValue_NativeBoundMethodPartialDoesNotDoubleInjectReceiver|FunctionScopeFilter|Call|BytecodeVM_Index|Cast|ExecFixtureParity/07_10_bytecode_quicksort_hotloop)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `~54-58s` in verification runs).
  - `./run_all_tests.sh` rerun passed end-to-end (`All tests completed successfully`).
- Perf signal checks:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^$' -bench '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchmem -count=3`
    -> `~39.37-40.06ms/op`, `~5.06MB/op`, `~150.3k allocs/op`
    (previous stage: `~40.48-41.85ms/op`, `~5.70MB/op`, `~162.3k allocs/op`).
  - `go tool pprof -top /tmp/bytecode-quicksort-hotloop.mem.after-callpartial.out`
    -> total alloc-space `~126.31MB` (down from prior `~144.70MB`), `callCallableValue` flat alloc down to `~4.50MB`, and legacy `functionScopeSet` allocator removed from hotspot list.

# 2026-03-04 — Inline `execBinary` stack pops + specialized integer extraction tightening (v12)
- Reduced remaining bytecode recursion-loop overhead in binary dispatch and call-frame unwind:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go`
- Behavior/perf changes:
  - `execBinary(...)` now performs direct stack pop operations inline (single bounds check + slice truncation) instead of two `vm.pop()` calls.
  - specialized integer opcode handling now prioritizes direct integer extraction (`bytecodeDirectIntegerValue`) and falls back to wider `bytecodeIntegerValue` only when needed.
  - `popCallFrameFields(...)` no longer clears the truncated tail slot on every pop; this removes unnecessary per-return frame writes in hot recursion loops.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_(Binary|SelfCallSlot|SlotConstImmediateCacheBuildsAndRefreshes|BoxedSmallIntValueCache)$' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `51.342s`).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (5x, `-benchtime=1x`): `245ms`, `249ms`, `247ms`, `261ms`, `244ms`.
  - pprof probe (`BenchmarkFib30Bytecode`, 1x `~257ms/op`) shows `vm.pop` removed from top hotspots; remaining top nodes are `runResumable`, `execBinary`, and `execCallSelfIntSubSlotConst`.
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup`: `6.28s`, `gc_avg~796.67`.

# 2026-03-04 — Pointer-based bytecode instruction dispatch in run loop (v12)
- Removed per-op `bytecodeInstruction` value copies from the hot VM loop and kept hot handlers pointer-based:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
- Behavior/perf changes:
  - `runResumable` now fetches instructions by pointer (`instr := &instructions[vm.ip]`) rather than copying the full instruction struct every iteration.
  - hot handler signatures now take pointers:
    - `execBinary(...)`
    - `execBinarySlotConst(...)`
    - `execBinarySpecializedOpcode(...)`
    - `callSelfIntSubSlotConstArg(...)`
    - `execCallSelfIntSubSlotConst(...)`
  - non-hot handlers in the run switch dereference as needed (`*instr`) to preserve existing behavior.
  - this removed `runtime.duffcopy` from the benchmark hotspot set for `fib30` recursion.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_(SelfCallSlot|Binary|SlotConstImmediateCacheBuildsAndRefreshes)$' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `54.456s` and follow-up `49.956s` in later run).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (5x, `-benchtime=1x`): `287ms`, `272ms`, `301ms`, `293ms`, `272ms`.
  - subsequent post-tuning run with direct IP immediate lookup remained in `304ms`-`329ms` band.
  - pprof snapshot (`BenchmarkFib30Bytecode`, 1x, `~284ms/op`) no longer shows `runtime.duffcopy` in top nodes.
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup` remained noisy in this window (`6.51s` and `6.79s` probe averages; single-run spot check `6.29s`), so macro quicksort movement is not yet claimed as deterministic from this change alone.

# 2026-03-04 — Specialized `BinaryIntAdd/Sub` boxed fast path + direct slot-const IP lookup (v12)
- Reduced remaining recursion-loop overhead in specialized integer arithmetic and slot-const immediate retrieval:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates.go`
- Behavior/perf changes:
  - added `addIntegerSameTypeFast(...)` companion to subtract fast path; both now return boxed `runtime.Value` fast results when safe.
  - `BinaryIntAdd`/`BinaryIntSub` specialized opcode execution now attempts same-suffix int64 fast-path with boxed small-int reuse before falling back to generic integer arithmetic.
  - added direct slot-const immediate lookup by instruction pointer (`bytecodeSlotConstImmediateAtIP`) and wired hot opcode handlers (`execBinary`, `execCallSelfIntSubSlotConst`) to use it directly, avoiding extra helper switch/fallback overhead on the fast path.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(SlotConstImmediateCacheBuildsAndRefreshes|BoxedSmallIntValueCache|SelfCallSlot|Binary)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `49.956s`).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (5x, `-benchtime=1x`): `329ms`, `304ms`, `317ms`, `304ms`, `322ms`.
  - post-change pprof probe (`BenchmarkFib30Bytecode`, 1x) around `~335ms/op`.
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup`: `6.15s`, `gc_avg~793.33`.

# 2026-03-04 — Pre-boxed small-int reuse for fused recursive slot-const paths (v12)
- Reduced integer-interface boxing overhead in hot fused recursion paths by reusing pre-boxed `runtime.Value` small ints:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`
- Behavior/perf changes:
  - added cached boxed small-int tables for `i32`, `i64`, and `isize` over `[-256, 4096]`.
  - `subtractIntegerSameTypeFast(...)` now returns `runtime.Value` and serves boxed cached values when in range.
  - fused `CallSelfIntSubSlotConst` inline setup and slot-const binary subtraction now consume the boxed result directly, avoiding repeated `runtime.IntegerValue -> interface` boxing on the hottest recursive path.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(SlotConstImmediateCacheBuildsAndRefreshes|BoxedSmallIntValueCache|SelfCallSlot|Binary|AwaitExpressionManualWaker|BoundMethodInline)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `50.315s`).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (5x, `-benchtime=1x`): `384ms`, `388ms`, `393ms`, `387ms`, `384ms` (material improvement vs prior ~`500ms` band).
  - pprof snapshot (`BenchmarkFib30Bytecode`, 1x) now reports `~386ms/op`; `runtime.convT` no longer appears in top sample list for this benchmark.
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup`: `6.34s`, `gc_avg~891`.

# 2026-03-04 — Slot-const immediate sparse cache + same-program switch skip (v12)
- Added a per-program sparse table for decoded slot-const integer immediates and wired run-loop switching to carry that table:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run_program_switch.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_types.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`
- Behavior/perf changes:
  - replaced eager per-instruction immediate arrays with sparse `(ip -> immediate)` entries for slot-const opcodes only, reducing cache memory/GC pressure on larger programs.
  - run-loop now fetches slot-const immediates from the per-program sparse table for:
    - `BinaryIntSubSlotConst`
    - `BinaryIntLessEqualSlotConst`
    - `CallSelfIntSubSlotConst`
  - `switchRunProgram(...)` now skips cache refresh work when `next == current program`, eliminating redundant map/cache reloads on hot self-recursive inline switches.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(SlotConstImmediateCacheBuildsAndRefreshes|SelfCallSlot|Binary|AwaitExpressionManualWaker|BoundMethodInline)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `52.117s`).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (5x, `-benchtime=1x`): `529ms`, `505ms`, `501ms`, `584ms`, `536ms`.
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup`:
    - probe A: `6.38s`, `gc_avg~893.67`
    - probe B: `6.28s`, `gc_avg~892.33`

# 2026-03-04 — Bytecode fused self-call contract tightening + call-frame helper refactor (v12)
- Tightened fused recursive self-call lowering/runtime path and reduced hot call-frame churn:
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering_callself_slot_const.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run_finalize.go`
- Behavior/perf changes:
  - fused `bytecodeOpCallSelfIntSubSlotConst` lowering now requires:
    - one arg,
    - no call-site type args,
    - and frame layouts where first-param coercion for integer recursion is statically safe.
  - removed per-call generic/type-arg guard checks from the fused runtime recursion path by relying on the stricter lowering contract.
  - fused self-recursive inline execution now sets up child slot frames directly from current frame layout for same-suffix integer `slot - const` cases before generic fallback.
  - added dedicated call-frame push/pop helpers (`pushCallFrame`, `popCallFrameFields`) and used them in run-loop return/unwind paths to avoid repeated struct-literal stack operations in hot paths.
  - expanded same-suffix integer subtract fast path reuse (`subtractIntegerSameTypeFast`) across fused self-call and slot+immediate binary execution.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(SelfCallSlot|BoundMethodInline|Binary)' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass; interpreter suite `53.302s`).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.
- Perf signal checks:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^$' -bench '^BenchmarkFib30Bytecode$' -benchtime=1x -count=5`:
    - `464ms`, `576ms`, `468ms`, `484ms`, `501ms` (best observed `~0.465s` in this batch).
  - `./v12/bench_suite --modes bytecode --benchmarks quicksort --runs 3 --timeout 45 --skip-setup`:
    - `quicksort` avg `6.04s`, `gc_avg~838.67` (single-run probes in same window: `6.13s`).

# 2026-03-04 — Bytecode recursive integer hot-path tightening (v12)
- Reduced per-call overhead on fused recursive integer call paths (`self(slot - const)`) and specialized integer arithmetic:
  - `v12/interpreters/go/pkg/interpreter/numeric_helpers.go`
  - `v12/interpreters/go/pkg/interpreter/int64_fast_arith.go`
  - `v12/interpreters/go/pkg/interpreter/interpreter_operations_fast.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_frames.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_types.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_slot_analysis.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
- Behavior/perf changes:
  - `promoteIntegerTypes` now returns immediately when operand suffixes already match, avoiding repeated integer-info lookups in hot arithmetic.
  - added `ensureFitsInt64Type(...)` and routed same-suffix integer fast paths through direct suffix checks (no integer-info map lookup on common non-overflow path).
  - extracted `evaluateIntegerArithmeticFast(...)` so bytecode specialized integer ops avoid re-entering generic `runtime.Value` arithmetic dispatch.
  - slot-frame pooling now has a hot-size fast pool (`slotFrameHotSize`/`slotFrameHotPool`) so recursive functions reuse frames without per-call `map[int]...` lookups.
  - frame layout now caches one-arg self-call inline metadata (`firstParamType` / `firstParamSimple`) and `tryInlineSelfCallWithArg(...)` uses cached metadata instead of re-reading function declaration shape each call.
  - `bytecodeIntegerValue(...)` now checks direct integer scalar forms before interface/scalar unwrap probing, reducing overhead on specialized bytecode integer ops.
  - slot+immediate recursion paths now execute direct integer logic first (`callSelfIntSubSlotConstArg`, `execBinarySlotConst`) instead of routing through `execBinaryIntegerSpecialized` helper for the hot common case.
  - hot specialized bytecode binary opcodes (`BinaryIntAdd/Sub/<=`) now use direct opcode-specific execution in `execBinary(...)` with explicit operator fallbacks, reducing helper dispatch overhead.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_SelfCallSlotAvoidsCallNameLookups|TestBytecodeVM_SelfCallSlotDisabledWhenFunctionNameAssigned|TestBytecodeVM_BinaryFastPath|TestBytecodeVM_BinarySlotConstTypeErrorParity|TestIntegerLiteralSuffixPreserved' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output (no files >1000 lines).
- Perf signal checks:
  - `BenchmarkFib30Bytecode` (single-iteration runs):
    - prior to this tranche: ~`1.015s/op`,
    - after same-suffix arithmetic + hot slot-frame pool + cached one-arg inline metadata + direct slot-const integer paths + direct specialized-opcode execution: consistently ~`0.47s–0.56s/op` in current runs (best observed `0.474s`; representative `~0.53s`).
  - benchmark suite snapshot:
    - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`
    - `fib`: timeout at `45s` (still not completing),
    - `quicksort`: `5.95s`–`5.99s`, `gc_avg~803-804` (improved from recent `6.74s` snapshot; still noisy).

# 2026-03-04 — Fused self-recursive call opcode + method-cache concurrency hardening (v12)
- Added fused recursion call opcode path for the common `self(slot - const)` shape:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering_callself_slot_const.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_binary_fastpath_test.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_self_call_slot_test.go`
- Behavior changes:
  - new opcode: `bytecodeOpCallSelfIntSubSlotConst`.
  - lowering now emits the fused self-call opcode when the recursive call argument is `slot-backed identifier - integer literal`.
  - VM executes the fused opcode by computing the arg directly from slot+immediate, then taking the dedicated self-inline path (`tryInlineSelfCallWithArg`) or normal call fallback.
  - recursive self-call regression now accepts either `bytecodeOpCallSelf` or the fused opcode as valid optimized execution.
- Concurrency hardening:
  - synchronized `Interpreter.methodCache` access and invalidation to prevent concurrent map writes under async/future execution:
    - `v12/interpreters/go/pkg/interpreter/interpreter_interface_lookup.go`
    - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
    - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - bytecode member-method cache now reads method-cache version via lock-guarded accessor:
    - `v12/interpreters/go/pkg/interpreter/bytecode_vm_member_cache.go`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - resolved prior nondeterministic panic in full interpreter package run:
    - `fatal error: concurrent map writes` at `findMethodCached` under async execution no longer reproduces.
- Perf signal checks:
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (still not completing),
    - `quicksort`: `6.74s`, `gc_avg=806` (still within noisy band; no deterministic macro gain yet).

# 2026-03-04 — Bytecode run/lowering helper splits to restore <1000-line guardrail (v12)
- Refactored without semantic changes to keep files under AGENTS line limit:
  - moved lowering support helpers out of `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go` into `v12/interpreters/go/pkg/interpreter/bytecode_lowering_helpers.go`.
  - moved run-loop program-switch helper out of `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go` into `v12/interpreters/go/pkg/interpreter/bytecode_vm_run_program_switch.go`.
- Validation:
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output.

# 2026-03-04 — Self-recursive single-parameter inline setup shortcut (v12)
- Added a tighter inline setup branch for the most common recursive shape (`fn f(n) -> ... f(n-1) ...`):
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
- Behavior changes:
  - `tryInlineSelfCallFromStack(...)` now has a single-parameter fast branch when:
    - no generics,
    - no implicit-member receiver usage,
    - and param coercion is trivially unnecessary.
  - this bypasses the generic per-parameter setup loop and generic-name handling for this shape, while retaining the existing fallback path when preconditions are not met.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_MemberMethodCacheInvalidatesOnImplChange|TestBytecodeVM_SelfCallSlot|TestBytecodeVM_BinaryFastPath|TestBytecodeVM_BinarySlotConstTypeErrorParity|TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` unchanged versus immediately prior slot-const pass (`InlineCallHits=1664078`, `CallNameLookups=0`, low `Const` opcode count maintained).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged),
    - `quicksort`: `6.08s`, `gc_avg=805` (still noisy; no deterministic macro gain signal).

# 2026-03-04 — Slot+immediate integer bytecode ops (`slot - const`, `slot <= const`) (v12)
- Reduced recursion hot-path dispatch for slot-backed identifiers with small integer literals:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering_binary_slot_const.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_binary_fastpath_test.go`
- Behavior changes:
  - new opcodes:
    - `bytecodeOpBinaryIntSubSlotConst`
    - `bytecodeOpBinaryIntLessEqualSlotConst`
  - lowering now emits these opcodes when an eligible expression matches:
    - left: slot-backed identifier,
    - right: untyped integer literal fitting `i32`,
    - operator: `-` or `<=`.
  - VM executes slot-const opcodes without stack operand pops:
    - reads left operand directly from slot index,
    - uses embedded integer immediate for right operand,
    - applies integer-specialized fast path first, then falls back to full `applyBinaryOperator` semantics for non-integer dynamic values.
  - additional call-frame setup cleanup:
    - `execCallSelf` now uses dedicated `tryInlineSelfCallFromStack(...)` path for direct function self-calls,
    - slot-frame pool no longer double-clears frames on both acquire and release (release-only clear).
- Regression/parity coverage:
  - `TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes`
  - `TestBytecodeVM_BinarySlotConstTypeErrorParity`
  - existing binary/self-call fast-path tests remain green.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_BinaryFastPath|TestBytecodeVM_LoweringEmitsIntegerBinaryHotOpcodes|TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes|TestBytecodeVM_BinarySlotConstTypeErrorParity|TestBytecodeVM_SelfCallSlot' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - file-size guardrail check (`fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}'`) produced no output.
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` now shows major const/slot opcode mix reduction for recursive probe shape:
    - previously (before slot-const ops): `op[0]` (`Const`) ~`4,992,236`,
    - after slot-const ops: `op[0]` ~`1,664,079` (with new slot-const opcodes active in the top mix).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged),
    - `quicksort`: `6.28s`, `gc_avg=805` (still noisy).
  - `./v12/bench_suite --runs 1 --timeout 120 --build-timeout 240 --modes bytecode --benchmarks fib`:
    - `fib` still timed out at `120s` (no completion yet).

# 2026-03-04 — `CallSelf` inline setup tightening + slot-frame pool clear dedupe (v12)
- Reduced recursive call-frame setup overhead on self-recursive bytecode call sites:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_frames.go`
- Behavior changes:
  - added `tryInlineSelfCallFromStack(...)` and wired `execCallSelf` to use it for `*runtime.FunctionValue` callees.
  - self-call inline setup now skips bound-method/general callee-shape checks used by the shared `tryInlineCallFromStack(...)` path, while preserving fallback semantics when inline preconditions are not met.
  - slot-frame pooling now clears frames once per reuse cycle (on release only); `acquireSlotFrame` no longer performs a second redundant `clear(...)`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_SelfCallSlot|TestBytecodeVM_BinaryFastPath|TestBytecodeVM_LoweringEmitsIntegerBinaryHotOpcodes' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/12_01_bytecode_await_default' -count=5` (pass; no reproduction of prior transient mismatch).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` remains stable for recursion probe shape (`CallNameLookups=0`, `InlineCallHits=1664078`, specialized int op mix unchanged).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged).
    - `quicksort`: `6.06s` real, `gc_avg=809` (still noisy, slight movement vs immediately prior run).
  - `./v12/bench_suite --runs 1 --timeout 90 --build-timeout 240 --modes bytecode --benchmarks fib` still times out (no measurable completion yet).

# 2026-03-04 — Dedicated integer binary opcodes for `+` / `-` / `<=` (v12)
- Added specialized bytecode opcodes and lowering for fib-style integer hot paths:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_binary_opcode.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_binary_fastpath_test.go`
- Behavior changes:
  - new opcodes:
    - `bytecodeOpBinaryIntAdd`
    - `bytecodeOpBinaryIntSub`
    - `bytecodeOpBinaryIntLessEqual`
  - lowering now emits these opcodes directly for plain `+`, `-`, and `<=` binary expressions.
  - VM dispatch handles these with integer-specialized execution first:
    - `+`/`-` use `evaluateArithmeticFast` with integer operands,
    - `<=` uses direct integer comparison (`int64` fast path, `big.Int` fallback).
  - when operands are not integers, behavior falls back to existing generic operator semantics (including string/numeric checks and full dispatch path).
- Regression/parity coverage:
  - `TestBytecodeVM_BinaryFastPathIntegerParity`
  - `TestBytecodeVM_BinaryFastPathFloatFallbackParity`
  - `TestBytecodeVM_BinaryFastPathOverflowParity`
  - `TestBytecodeVM_BinaryFastPathTypeErrorParity`
  - `TestBytecodeVM_LoweringEmitsIntegerBinaryHotOpcodes`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_BinaryFastPath|TestBytecodeVM_LoweringEmitsIntegerBinaryHotOpcodes|TestBytecodeVM_AssignmentAndBinary' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` now shows dedicated integer opcode execution in the recursion probe (`op[9]`, `op[10]`, `op[11]` counts active for the specialized integer ops).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged).
    - `quicksort`: `6.36s` real, `gc_avg=813` (still noisy; no deterministic macro-level gain signal yet).

# 2026-03-04 — Bytecode `execBinary` numeric fast path + parity coverage (v12)
- Reduced bytecode binary-operator dispatch overhead for common numeric/comparison operators:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_binary_fastpath_test.go`
- Behavior changes:
  - `execBinary` now tries `ApplyBinaryOperatorFast` first for `+`, `-`, `<`, `<=`, `>`, `>=`, `==`, `!=`.
  - when the fast evaluator does not handle the operand/operator combination, bytecode falls back to existing `applyBinaryOperator` behavior (interface dispatch and full semantics unchanged).
  - existing string-`+` guard semantics remain intact.
- Regression/parity coverage:
  - `TestBytecodeVM_BinaryFastPathIntegerParity` validates normal integer arithmetic/comparison parity vs tree-walker.
  - `TestBytecodeVM_BinaryFastPathOverflowParity` validates integer-overflow error parity (`integer overflow`) vs tree-walker.
  - `TestBytecodeVM_BinaryFastPathTypeErrorParity` validates arithmetic type-error parity (`Arithmetic requires numeric operands`) vs tree-walker.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_BinaryFastPath|TestBytecodeVM_AssignmentAndBinary' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output (no files over 1000 lines).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` remains functionally stable (`CallNameLookups=0`, recursive `CallSelf` opcode mix unchanged for the probe).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged).
    - `quicksort`: `6.11s` real, `gc_avg=804` (within current noisy range; no deterministic macro gain signal yet).

# 2026-03-04 — Bytecode const-validation memoization (v12)
- Reduced repeated integer-range checking overhead on bytecode `Const` instructions:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_types.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_const_validation.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_const_validation_test.go`
- Behavior changes:
  - bytecode VM now memoizes successful integer-literal range validation per `(program, instruction index)` and skips repeated `ensureFitsInteger(...)` checks for subsequent executions of the same `Const` instruction.
  - validation remains lazy and execution-path-dependent: instructions are validated only when reached at runtime.
- Regression coverage:
  - `TestBytecodeVM_IntegerLiteralValidationRemainsLazy` verifies:
    - unreachable overflow literal does not fail when an earlier branch `return`s,
    - overflow still raises when the literal path is executed.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_IntegerLiteralValidationRemainsLazy|TestBytecodeVM_SelfCallSlot' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` confirms opcode mix is unchanged semantically for the recursive probe after memoization.
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged).
    - `quicksort`: `5.98s` real, `gc_avg=804` (no deterministic improvement signal yet; still noisy).

# 2026-03-04 — Slot-frame scope-op elision + self-call opcode fast path (v12)
- Reduced recursive bytecode dispatch overhead in slot-enabled functions:
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_self_call_slot_test.go`
- Behavior changes:
  - lowerer now skips emitting `EnterScope`/`ExitScope` opcodes when slot analysis marks `needsEnvScopes=false` (runtime still preserves scope semantics for frames that require env scope chains).
  - added `bytecodeOpCallSelf`; stable self-recursive sites now lower to direct slot-indexed self calls instead of `LoadSlot+Call`, removing one high-frequency callee stack push/pop per recursive call.
  - `execCallSelf` shares the same inline-call fast path and fallback semantics as existing `Call`/`CallName` handling.
- Regression coverage:
  - `TestBytecodeVM_SelfCallSlotAvoidsCallNameLookups` now additionally asserts `bytecodeOpCallSelf` executes on recursive paths.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_SelfCallSlot' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
  - `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}' | sort` produced no output (no files over 1000 lines).
- Perf signal checks:
  - `ABLE_BYTECODE_STATS=1 go run /tmp/able_stats_probe_ops.go` now shows recursive `fib(30)` using `op[71]=1664078` (`bytecodeOpCallSelf`) and reduced `LoadSlot` dispatch (`op[67]=3328157` vs prior `4992235` for the same probe shape).
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks fib,quicksort`:
    - `fib`: timeout at `45s` (unchanged).
    - `quicksort`: `5.97s` real, `gc_avg=764` (within current noise band; no deterministic macro-level gain yet).

# 2026-03-04 — Slot coverage expansion for self-recursive calls (v12)
- Reduced high-frequency recursive name-resolution overhead in slot-enabled bytecode functions:
  - `v12/interpreters/go/pkg/interpreter/definitions.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_slot_analysis.go`
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - added `v12/interpreters/go/pkg/interpreter/bytecode_self_call_slot.go`.
- Behavior:
  - slot-enabled lowering now reserves a hidden self-call slot (`frameLayout.selfCallSlot`) for functions where the function identifier is not assigned anywhere in the function body (conservative AST scan).
  - recursive call sites (`f(...)` inside `fn f`) lower to `LoadSlot+Call` instead of `CallName` when that self-call slot is enabled.
  - runtime call setup seeds the reserved self slot for both standard bytecode invocation and inline call frames.
  - when the function name is assigned in the function body, the optimization is disabled and bytecode keeps regular `CallName` semantics.
- Regression coverage:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_self_call_slot_test.go`
  - `TestBytecodeVM_SelfCallSlotAvoidsCallNameLookups`
  - `TestBytecodeVM_SelfCallSlotDisabledWhenFunctionNameAssigned`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_SelfCallSlotAvoidsCallNameLookups|TestBytecodeVM_SelfCallSlotDisabledWhenFunctionNameAssigned' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
- Perf signal checks:
  - targeted stats probe (`fib(20)`/`fib(30)` style recursive run under `ABLE_BYTECODE_STATS=1`) now records `CallNameLookups=0` for recursive self calls while retaining full inline-call hits.
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort,fib` observed:
    - `quicksort`: `5.84s` real (`gc_avg=754`) — aligned with recent noise band.
    - `fib`: still timed out at `45s`, so additional recursion-path optimization is still required.

# 2026-03-04 — Environment thread-mode sharing + dotted `CallName` miss fast path (v12)
- Reduced high-frequency environment overhead in interpreter + bytecode hot loops:
  - `v12/interpreters/go/pkg/runtime/environment.go`
  - replaced per-environment `singleThread` bool with a shared thread-mode flag across lexical scope chains.
  - `NewEnvironment(parent)` now inherits the parent thread-mode handle so local scopes remain lock-free in single-thread execution.
  - `SetSingleThread`/`SetMultiThread` now flip shared mode for the full chain; existing child scopes observe the change immediately.
  - added single-thread fast paths for `Define`, `DefineStruct`, `Snapshot`, `StructSnapshot`, `SetRuntimeData`, and `HasInCurrentScope`.
- Added runtime regression coverage:
  - `v12/interpreters/go/pkg/runtime/environment_test.go`
  - `TestEnvironmentThreadModePropagatesToChildren` verifies child envs inherit and observe thread-mode transitions from parent/global.
- Reduced dotted-call dispatch miss churn:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_lookup_cache.go`
  - added non-error lookup path (`lookupCachedName`) so `CallName` probing can test existence without constructing miss errors.
  - `resolveCachedName` now wraps the lookup path and only constructs undefined-variable errors for terminal misses.
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - `execCallName` now uses non-error probing for primary and dotted-head resolution paths, building an undefined-variable error only when fallback is not applicable.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestEnvironmentLookupInCurrentScopeDoesNotWalkParent|TestEnvironmentLookupRespectsLexicalScope|TestEnvironmentThreadModePropagatesToChildren' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_CallNameDotFallbackScopeCacheInvalidatesOnHeadRebind|TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind|TestBytecodeVM_CallNameCacheInvalidatesOnRebind|TestBytecodeVM_StatsCounters' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime ./pkg/interpreter -count=1` (pass).
- Perf signal checks (bytecode, noisy single-host):
  - before this tranche:
    - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` -> `5.94s` avg real (`gc_avg=752.0`).
  - after environment thread-mode sharing:
    - same command -> `5.80s` avg real (`gc_avg=752.0`).
  - after dotted `CallName` non-error lookup path:
    - same command -> `5.785s` avg real (`gc_avg=752.5`).
  - result: modest but consistent movement in the right direction; still within expected run-to-run noise band, so larger phase-1 cuts remain necessary.

# 2026-03-04 — Non-inline call-dispatch fast path (single-overload + partial merge) (v12)
- Reduced fallback call-dispatch overhead in interpreter runtime paths used by bytecode non-inline calls:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - partial-function argument merge now uses a single-pass merge helper (`mergePartialCallArgs`) instead of chained appends.
  - added single-overload fast path in `callCallableValue`: when exactly one overload is present, do direct arity/type compatibility check + `invokeFunction(...)` instead of full overload candidate selection/scoring.
- Added compatibility helper for single-overload checks:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls_overloads.go`
  - `matchesSingleRuntimeOverload(...)` mirrors existing runtime compatibility behavior for function/lambda declarations.
- Removed unnecessary map allocations in generic-name extraction:
  - `v12/interpreters/go/pkg/interpreter/call_helpers.go`
  - `functionGenericNameSet(...)` now lazily allocates only when generic params actually exist.
- Regression coverage:
  - `v12/interpreters/go/pkg/interpreter/call_dispatch_fastpath_test.go`
    - `TestCallDispatchPartialChainPreservesBoundArgOrder`
    - `TestCallDispatchSingleOverloadMismatchReportsParameterType`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestCallDispatchPartialChainPreservesBoundArgOrder|TestCallDispatchSingleOverloadMismatchReportsParameterType|TestCallFunction_DoesNotMutateCallerArgsOnCoercion' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestEnvironmentLookupInCurrentScopeDoesNotWalkParent|TestEnvironmentLookupRespectsLexicalScope' -count=1` (pass).
- Perf signal checks (bytecode, noisy single-host):
  - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1` observed `1.832s/op` and `1.851s/op`.
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` observed `5.865s` real average (`gc_avg=747.5`).
  - Result: quicksort moved slightly toward the lower end of the recent noise band; improvements are incremental, not yet step-change.

# 2026-03-04 — Dotted `CallName` receiver-head cached resolution (v12)
- Extended dotted call fallback path in bytecode call dispatch to reuse cached name resolution for receiver heads:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - when `CallName` target lookup fails and falls back to dotted form (`head.tail`), receiver `head` is now resolved via `resolveCachedName(...)` instead of direct `env.Get(head)`.
  - this applies existing safe cache invalidation rules (global revision + non-global current-scope env pointer/revision) to dotted receiver-head lookups.
- Added regression coverage:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_scope_lookup_cache_test.go`
  - `TestBytecodeVM_CallNameDotFallbackScopeCacheInvalidatesOnHeadRebind`
  - scenario: local `s` receiver is rebound between repeated `CallName("s.get")` uses at the same instruction site; second call must observe the rebound receiver.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_CallNameDotFallbackScopeCacheInvalidatesOnHeadRebind|TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind|TestBytecodeVM_LoadNameScopeCacheInvalidatesOnLocalAssign|TestBytecodeVM_CallNameCacheInvalidatesOnRebind' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestEnvironmentLookupInCurrentScopeDoesNotWalkParent' -count=1` (pass).
- Perf signal checks (bytecode, noisy single-host sample):
  - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1` observed `1.874s/op` and `1.926s/op`.
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` observed `5.94s` real average (`gc_avg=756.0`).
  - Result: quicksort moved back toward the earlier noise band; no clear deterministic gain yet.

# 2026-03-04 — Bytecode scoped name lookup cache for non-global envs (v12)
- Reduced non-inline name-resolution overhead in bytecode function bodies that are not slot-eligible:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_lookup_cache.go`
  - added `resolveCachedName(...)` and conservative current-scope cache path for non-global `LoadName`/`CallName`.
  - cache eligibility: non-dotted names only; cache stores only bindings found in the **current** scope.
  - cache key/invalidation: `(program pointer, instruction pointer)` plus `(env pointer, env.Revision())`.
  - global-scope cache behavior remains unchanged.
- Runtime API support:
  - `v12/interpreters/go/pkg/runtime/environment.go`
  - added `(*Environment).LookupInCurrentScope(name) (Value, bool)` to avoid parent walks/error allocation for current-scope probes.
- VM wiring:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`: `bytecodeOpLoadName` now resolves through `resolveCachedName(...)`.
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`: `bytecodeOpCallName` now resolves through `resolveCachedName(...)` before dotted fallback handling.
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_types.go`: added `scopeLookupCache` map on VM state.
- Regression coverage:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_scope_lookup_cache_test.go`
    - `TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind`
    - `TestBytecodeVM_LoadNameScopeCacheInvalidatesOnLocalAssign`
  - `v12/interpreters/go/pkg/runtime/environment_test.go`
    - `TestEnvironmentLookupInCurrentScopeDoesNotWalkParent`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/runtime -run 'TestEnvironmentLookupRespectsLexicalScope|TestEnvironmentLookupInCurrentScopeDoesNotWalkParent' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind|TestBytecodeVM_LoadNameScopeCacheInvalidatesOnLocalAssign|TestBytecodeVM_CallNameCacheInvalidatesOnRebind|TestBytecodeVM_InlineBoundMethodCallStats' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
- Perf signal checks (bytecode, single-host noise expected):
  - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1` observed `1.850s/op` then `1.825s/op`.
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` observed `6.49s` real average (`gc_avg=756.5`).
  - Result: no deterministic speedup signal yet; change is retained for correctness + reduced lookup work in non-slot paths while we continue larger hotspot work.

# 2026-03-04 — Bytecode Phase 1 bound-method inline call fast path (v12)
- Extended bytecode inline call setup so call-position member dispatch can inline when the callee is a bound method:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
  - new helper `inlineCallFunctionValue(...)` now recognizes `*runtime.BoundMethodValue` / `runtime.BoundMethodValue` with `*runtime.FunctionValue` methods.
  - `tryInlineCallFromStack` now injects the bound receiver into slot `0` and maps stack arguments to the remaining parameter slots without allocating an argument slice on successful inline calls.
  - native-bound methods and overload-backed bound methods keep existing fallback behavior through `callCallableValue`.
- Added regression coverage:
  - `v12/interpreters/go/pkg/interpreter/bytecode_vm_bound_method_inline_test.go`
  - `TestBytecodeVM_InlineBoundMethodCallStats` verifies bytecode/treewalker parity and confirms inline-call hits are recorded for bound-method call sites with `ABLE_BYTECODE_STATS=1`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_InlineBoundMethodCallStats|TestBytecodeVM_StatsCounters|TestBytecodeVM_StatsMemberMethodCacheCounters|TestBytecodeVM_MemberMethodCacheInvalidatesOnImplChange' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (bytecode, single-workstation noise expected):
  - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1` observed `1.896s/op` then `1.717s/op` on back-to-back runs.
  - `./v12/bench_suite --runs 1 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` observed `6.18s` then `5.87s` real with similar GC counts (`736`/`738`).
  - Result: improvement is in the expected noise band for this benchmark mix; no regression signal outside prior variability.

# 2026-03-04 — Call-dispatch fallback arg-slice churn reduction (v12)
- Reduced runtime call-dispatch allocation pressure in non-inline paths:
  - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
  - `callCallableValue` now avoids unconditional `append(injected, args...)` allocation when no receiver injection is needed; it reuses `args` directly in that common case.
  - receiver-injected paths (`BoundMethod`/`NativeBoundMethod`) still allocate a combined slice as required.
- Preserved arg-slice immutability semantics while removing dispatcher copies:
  - `invokeFunction` now uses a lazy writable copy for parameter binding, only when mutation is required (optional-arg fill or actual coercion rewrite).
  - added fast skip for obvious no-op coercions via `inlineCoercionUnnecessary(...)` in the runtime invocation path.
- Added regression coverage:
  - `v12/interpreters/go/pkg/interpreter/call_args_mutation_test.go`
  - `TestCallFunction_DoesNotMutateCallerArgsOnCoercion` verifies host-provided argument slices are not mutated even when coercion occurs inside call dispatch.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestCallFunction_DoesNotMutateCallerArgsOnCoercion|TestBytecodeVM_InlineBoundMethodCallStats|TestBytecodeVM_StatsCounters|TestBytecodeVM_StatsMemberMethodCacheCounters' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
  - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
- Perf signal checks (bytecode, noisy single-host samples):
  - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1` observed `1.84s/op` and `1.78s/op` on sequential runs.
  - `./v12/bench_suite --runs 2 --timeout 45 --build-timeout 240 --modes bytecode --benchmarks quicksort` observed `6.135s` real average (`gc_avg=743.5`).
  - Result: no clear regression signal beyond existing run-to-run noise; gains are modest and need further hotspot work.

# 2026-03-03 — Runtime performance harness + baseline snapshot (v12)
- Implemented machine-readable suite harness `v12/bench_suite` covering:
  - benchmarks: `fib`, `binarytrees`, `matrixmultiply`, `quicksort`, `sudoku`, `i_before_e`
  - modes: `compiled`, `treewalker`, `bytecode`
  - per-run status (`ok`/`timeout`/`error`) + `real/user/sys` + GC count + summary aggregation
  - metadata capture: git commit/dirty state, machine profile, toolchain version, run config
  - bounded execution controls: run timeout (`--timeout`) and compiled build timeout (`--build-timeout`)
  - isolated cache/bootstrap path (`ABLE_HOME` sandbox + setup) for reproducible runs
- Added harness documentation:
  - `v12/docs/performance-benchmarks.md`
  - `v12/README.md` now points to the benchmark harness/docs
- Captured and checked in baseline snapshot:
  - `v12/docs/perf-baselines/2026-03-03-benchmark-baseline.json`
  - command:
    - `./v12/bench_suite --runs 1 --timeout 30 --build-timeout 240 --output-json v12/docs/perf-baselines/2026-03-03-benchmark-baseline.json`
- Baseline highlights (1 run/mode, 30s run timeout):
  - completed successfully in all modes: `quicksort`, `i_before_e`
  - timeouts across all modes: `fib`, `binarytrees`
  - `matrixmultiply`: `compiled`/`treewalker` timeout; `bytecode` error
  - `sudoku`: compiled build error (captured as `error`), interpreted modes timed out
  - quick triage diagnostics after baseline:
    - `matrixmultiply` bytecode error: `array has no member 'get' (import able.collections.array for stdlib helpers)` at `v12/examples/benchmarks/matrixmultiply.able:34`
    - `sudoku` compiled build error: `static fallback not allowed ... unsupported function body` for `sudoku.sudoku.solve`
- Follow-up fixes (same day):
  - benchmark source fixes:
    - `v12/examples/benchmarks/matrixmultiply.able`: added `import able.collections.array` to satisfy required `Array.get` helper resolution in interpreter modes.
    - `v12/examples/benchmarks/sudoku/sudoku.able`: rewrote `solve` match-clause bodies to expression-style results (removed in-clause `return` statements) so compiled no-fallback lowering accepts the function body.
  - focused validation:
    - `./v12/bench_suite --benchmarks matrixmultiply --modes bytecode --runs 1 --timeout 35 --build-timeout 120` now reports `timeout` (no runtime error).
    - `./v12/bench_suite --benchmarks sudoku --modes compiled --runs 1 --timeout 5 --build-timeout 300` now reports `timeout` (no static fallback compile error).
  - Bytecode Phase 1 kickoff optimization:
    - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`: added stack-based inline-call setup path for `bytecodeOpCall`/`bytecodeOpCallName` so successful inline calls avoid transient args-slice allocation.
  - Bytecode perf observability:
    - added optional runtime counters behind `ABLE_BYTECODE_STATS` with interpreter APIs:
      - `(*Interpreter).BytecodeStats()`
      - `(*Interpreter).ResetBytecodeStats()`
    - counters include opcode dispatch counts, `LoadName`/`CallName` lookup counts, dotted `CallName` fallback count, and inline-call hit/miss counts.
    - files:
      - `v12/interpreters/go/pkg/interpreter/interpreter_bytecode_stats.go`
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
      - `v12/interpreters/go/pkg/interpreter/interpreter.go`
  - Added bytecode stats regression coverage:
    - `TestBytecodeVM_StatsCounters` in `v12/interpreters/go/pkg/interpreter/bytecode_vm_async_test.go`.
  - Follow-up Bytecode Phase 1 optimizations:
    - Added per-VM global lookup cache for `LoadName`/`CallName` in global scope:
      - cache key: `(program pointer, instruction pointer)` to avoid shared-program mutation/races.
      - cache invalidation: runtime `Environment.Revision()` mutation counter checked on each cache read.
      - files:
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_lookup_cache.go`
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
        - `v12/interpreters/go/pkg/runtime/environment.go`
    - Added inline slot-frame pooling for bytecode call frames so recursive inline calls reuse `[]runtime.Value` slot frames instead of allocating per call:
      - files:
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_frames.go`
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go`
        - `v12/interpreters/go/pkg/interpreter/bytecode_vm_run.go`
    - Added regression coverage:
      - `TestBytecodeVM_CallNameCacheInvalidatesOnRebind` (`bytecode_vm_async_test.go`)
      - `TestEnvironmentRevisionIncrementsOnMutation` (`runtime/environment_test.go`)
    - Added runtime environment non-error lookup API for hot miss-heavy probes:
      - `(*Environment).Lookup(name) (runtime.Value, bool)` in `v12/interpreters/go/pkg/runtime/environment.go`.
      - `resolveMethodFromPool` now uses `Lookup` instead of `Get` when probing scope-callable names, avoiding per-miss error construction.
    - Refactored method-candidate accumulation in `resolveMethodFromPool` to reduce allocation churn:
      - replaced per-call map+closure bookkeeping with a compact linear-dedupe accumulator (`methodResolutionAccumulator`) in `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`.
      - preserves ambiguity/private-method semantics while avoiding high-frequency heap churn in member-call loops.
    - Added conservative bytecode member-method inline cache for call-position member access (`preferMethods=true`) in `v12/interpreters/go/pkg/interpreter/bytecode_vm_member_cache.go`:
      - caches method templates (not receiver-bound instances) keyed by `(program, ip, member, receiver-shape)` for `Array`, `String`, and struct-instance receivers.
      - rebinds the cached template to the current receiver at each hit.
      - strict invalidation requires both:
        - global environment revision unchanged (`Environment.Revision()`), and
        - interpreter method-cache epoch unchanged (`Interpreter.methodCacheVersion`, incremented by `invalidateMethodCache()`).
      - wired through `execMemberAccess` in `v12/interpreters/go/pkg/interpreter/bytecode_vm_members.go`.
    - Added cache-invalidation regression coverage for unnamed `impl` changes:
      - `TestBytecodeVM_MemberMethodCacheInvalidatesOnImplChange` (`bytecode_vm_async_test.go`).
      - scenario: `s.greet()` first resolves via UFCS scope function, then unnamed `impl Greeter for S` is introduced; second execution of the same call-site must resolve to impl dispatch (proves epoch-based invalidation beyond env-revision invalidation).
    - Extended bytecode stats observability for the new member-call cache:
      - added `MemberMethodCacheHits` / `MemberMethodCacheMiss` fields to `BytecodeStatsSnapshot` in `v12/interpreters/go/pkg/interpreter/interpreter_bytecode_stats.go`.
      - wired instrumentation in `lookupCachedMemberMethod` to record hit/miss outcomes (including eligible misses before cache map initialization and invalidation misses).
      - added `TestBytecodeVM_StatsMemberMethodCacheCounters` in `bytecode_vm_async_test.go` to verify:
        - counters increment on a repeated member-call site,
        - `ResetBytecodeStats()` clears the new counters.
  - Added environment lookup regression coverage:
    - `TestEnvironmentLookupRespectsLexicalScope` (`runtime/environment_test.go`).
  - post-change checks:
    - `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_' -count=1` (pass).
    - `cd v12/interpreters/go && go test ./pkg/runtime -count=1` (pass).
    - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_|TestResolveMethod|TestInterpreterMethod|TestMember' -count=1` (pass).
    - quick signal check: `quicksort` bytecode remained within noise of baseline (`5.86s` -> `5.91s`, 1 run each); `fib(45)` still exceeds the 30s run timeout.
    - ad-hoc finite workload check: `./v12/ablebc v12/tmp/fib30.able` (`fib(30)`) completed in `real=1.06s` on this workstation.
    - profiler-guided follow-up signal (bytecode):
      - `ABLE_BENCH_FIXTURE=v12/fixtures/bench/sieve_full go test ./pkg/interpreter -run '^$' -bench BenchmarkExecFixtureBytecode -benchtime=2x -count=1`
      - before: `~1.779s/op`; after resolver/lookup churn reduction: `~1.702s/op` (~4.4% improvement in this harness run).
      - `./v12/bench_suite --runs 1 --timeout 45 --modes bytecode --benchmarks quicksort`: `5.83s` real, `741` GC (down from earlier 5.88–6.07s single-run noise band).
      - after member-call cache landing, repeated quick checks stayed in the same noise band:
        - `sieve_full` bench fixture (`-benchtime=2x`) observed `1.91s/op` then `1.72s/op` on successive runs.
        - `bench_suite quicksort` single runs observed `6.41s` then `6.06s` real with similar GC counts.

# 2026-03-03 — Stdlib setup smoke coverage + toolchain-pinned stdlib resolution policy (v12)
- Closed remaining staged-integration stdlib items from `PLAN.md`:
  - added clean-environment setup smoke coverage for stdlib+kernel bootstrap and cross-interpreter execution.
  - enforced/documented toolchain-pinned stdlib resolution semantics for implicit `able` dependencies.
- CLI/runtime changes:
  - `able setup` now resolves stdlib using the toolchain default version pin (`defaultStdlibVersion`) instead of an unpinned branch fetch.
  - dependency installer now injects `able` as `version: <toolchain pin>` when absent from manifest, ensuring lockfile stdlib entries are pinned by default.
  - stdlib git resolution now uses canonical version tags (`v<version>`) rather than floating `main` for implicit/default resolution paths.
- Tests:
  - added/updated coverage in `v12/interpreters/go/cmd/able/dependency_installer_test.go`:
    - `TestDependencyInstaller_PinsBundledStdlib` now asserts the bundled stdlib path is used only when it matches the toolchain pin.
    - `TestDependencyInstaller_RejectsBundledStdlibVersionMismatch` verifies mismatched local bundled stdlib is ignored in favor of pinned cached stdlib.
  - setup smoke fixture (`v12/interpreters/go/cmd/able/setup_smoke_test.go`) now keys stdlib manifest version off `defaultStdlibVersion`.
- Spec/docs:
  - updated `spec/full_spec_v12.md` §13.6 to codify implicit toolchain-pinned stdlib tag resolution, setup/auto-install parity, and lockfile behavior with override opt-ins.
  - removed the completed stdlib version-selection TODO from `spec/TODO_v12.md`.
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestSetupInstallsStdlibAndKernelAndRunSupportsBothExecModes|TestDependencyInstaller_PinsBundledStdlib|TestDependencyInstaller_RejectsBundledStdlibVersionMismatch' -count=1` (pass).

# 2026-02-20 — Compiler no-bootstrap execution path: 85% pass rate (v12)
- Continued Phase 3 of the no-bootstrap execution plan (spicy-wobbling-cascade.md).
- Progress: 58 failures → 35 failures (205/240 = 85.4% pass rate).
- Changes:
  1. **Binary constant folding** (`generator.go`): Added `evalConstInt` helper and `*ast.BinaryExpression` case in `literalToRuntimeExpr` to handle `(-MAX_i64) - 1` patterns for `I64_MIN`/`I32_MIN`. Fixed: `06_12_04_stdlib_numbers_bigint`, `06_12_09_stdlib_numbers_primitives`, `06_12_08_stdlib_numbers_rational`.
  2. **`compiledImplChecker` in `ensureTypeSatisfiesInterface`** (`impl_resolution_types.go`): When an interface isn't in `i.interfaces` (no-bootstrap mode), falls back to compiled dispatch table. Fixed: `10_02_impl_where_clause`, `10_02_impl_specificity_named_overrides`.
  3. **`compiledImplChecker` in `typeHasMethod`** (`impl_resolution_types.go`): Added compiled dispatch fallback after `findMethod` fails, so constraint enforcement works without `i.implMethods`.
  4. **`compiledImplChecker` for generic type expressions** (`interpreter_type_matching.go`): Added fallback paths for generic interface matching (e.g., `Formatter<String>`) when interface isn't in `i.interfaces`.
- Remaining 35 failures categorized:
  - 5 inherently dynamic (dynamic imports, host interop) — require bootstrap
  - 5 extern native functions (io_stdout, os_env, pipe_reader) — require host registration
  - ~10 deep interface/impl dispatch (Error interface, operator overloading, generic args)
  - ~15 various (UFCS, iterator dispatch, struct update, named impl namespaces)
- All bootstrap tests pass; all interpreter tests pass.

# 2026-02-19 — Compiler AOT method receiver parity for `Self`-typed first params (v12)
- Closed a compiler/interpreter parity gap in method receiver detection:
  - compiler method lowering now treats a method as instance-receiver when its first parameter type is `Self`, even if that parameter is not named `self`.
  - this matches interpreter semantics and prevents misclassification as static methods in compiled registration/dispatch.
- Files:
  - `v12/interpreters/go/pkg/compiler/generator_methods.go`
  - `v12/interpreters/go/pkg/compiler/compiler_method_self_param_detection_test.go`
- Added focused regression coverage:
  - `TestCompilerTreatsSelfTypedFirstMethodParamAsInstanceReceiver`
  - asserts `methods Counter { fn bump(this: Self) ... }` registers as `__able_register_compiled_method("Counter", "bump", true, ...)` (not static).
- Validation (all bounded below 30 minutes per suite):
  - `cd v12/interpreters/go && timeout 900s go test ./pkg/compiler -run 'TestCompilerTreatsSelfTypedFirstMethodParamAsInstanceReceiver|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod' -count=1 -timeout=14m` (pass, `ok ... 0.046s`)
  - `cd v12/interpreters/go && timeout 900s go test ./pkg/compiler -run 'TestCompilerNoFallbacksForLocalTypeDefinitions|TestCompilerNoFallbacksForLocalFunctionDefinitionStatement|TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding' -count=1 -timeout=14m` (pass, `ok ... 0.061s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 92.448s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 566.381s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=1 timeout 1800s go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=28m` (pass, `ok ... 97.949s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1 -timeout=28m` (pass, `ok ... 10.250s`)

# 2026-02-19 — Compiler AOT interface default-impl body metadata preservation (v12)
- Closed the remaining metadata parity gap for rendered interface signatures by preserving default-implementation bodies instead of emitting `nil`.
- Implementation:
  - added shared default-impl block renderer that serializes AST blocks to JSON and decodes them in generated code:
    - `v12/interpreters/go/pkg/compiler/generator_export_defs.go`
  - wired default-impl body preservation into:
    - package-level interface definition rendering (`renderInterfaceDefinitionExpr`)
    - block-local interface definition rendering (`renderLocalInterfaceDefinitionExpr`)
  - files:
    - `v12/interpreters/go/pkg/compiler/generator_export_defs.go`
    - `v12/interpreters/go/pkg/compiler/generator_local_type_decls.go`
  - exported interpreter decoder helper used by generated metadata:
    - `v12/interpreters/go/pkg/interpreter/fixtures_public.go`
    - new: `DecodeNodeJSON(data []byte) (ast.Node, error)`
- Extended focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_definition_metadata_render_test.go`
  - package-level + local assertions now require signature default-impl metadata to contain decode-backed block construction (`interpreter.DecodeNodeJSON(...)`).
- Validation (all bounded below 30 minutes per suite):
  - `cd v12/interpreters/go && timeout 900s go test ./pkg/compiler -run 'TestCompilerPreservesDefinitionGenericConstraintsAndWhereClauses|TestCompilerNoFallbacksForLocalDefinitionConstraintsAndWhereClauses|TestCompilerNoFallbacksForLocalTypeDefinitions|TestCompilerNoFallbacksForLocalInterfaceDefinitionWithDefaultImpl' -count=1` (pass, `ok ... 0.073s`)
  - `cd v12/interpreters/go && timeout 600s go test ./pkg/interpreter -count=1` (pass, `ok ... 69.504s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 91.493s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 524.221s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=1 timeout 1800s go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=28m` (pass, `ok ... 94.285s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1 -timeout=28m` (pass, `ok ... 10.658s`)

# 2026-02-19 — Compiler AOT definition metadata parity for generics/where constraints (v12)
- Closed a definition-metadata parity gap in compiled package/local definition rendering:
  - generic parameter interface constraints are now preserved when emitting AST metadata for struct/union/interface definitions.
  - `where`-clause constraints are now preserved when emitting AST metadata for struct/union/interface definitions and interface signatures.
- Files:
  - `v12/interpreters/go/pkg/compiler/generator_export_defs.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_struct_defs.go`
  - `v12/interpreters/go/pkg/compiler/generator_local_type_decls.go`
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_definition_metadata_render_test.go`
  - `TestCompilerPreservesDefinitionGenericConstraintsAndWhereClauses`
  - `TestCompilerNoFallbacksForLocalDefinitionConstraintsAndWhereClauses`
- Validation (all bounded below 30 minutes per suite):
  - `cd v12/interpreters/go && timeout 900s go test ./pkg/compiler -run 'TestCompilerPreservesDefinitionGenericConstraintsAndWhereClauses|TestCompilerNoFallbacksForLocalDefinitionConstraintsAndWhereClauses|TestCompilerNoFallbacksForLocalTypeDefinitions|TestCompilerNoFallbacksForLocalInterfaceDefinitionWithDefaultImpl' -count=1` (pass, `ok ... 0.073s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 87.435s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 525.507s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=1 timeout 1800s go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=28m` (pass, `ok ... 94.918s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1 -timeout=28m` (pass, `ok ... 9.996s`)

# 2026-02-19 — Compiler AOT local interface default-impl signature no-fallback parity (v12)
- Closed the remaining local type-definition sub-gap by allowing block-local `interface` declarations with default-impl signatures to lower in compiled mode instead of being marked unsupported.
  - file: `v12/interpreters/go/pkg/compiler/generator_local_type_decls.go`
  - change: local interface signature rendering no longer rejects `sig.DefaultImpl != nil`.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_local_type_definition_no_fallback_test.go`
  - `TestCompilerNoFallbacksForLocalInterfaceDefinitionWithDefaultImpl`
  - validates local interface default-impl signatures compile under `RequireNoFallbacks: true` and avoid `CallOriginal("demo.main", ...)`.
- Validation (all bounded below 30 minutes per suite):
  - `cd v12/interpreters/go && timeout 600s go test ./pkg/compiler -run 'TestCompilerNoFallbacksForLocalTypeDefinitions|TestCompilerNoFallbacksForLocalInterfaceDefinitionWithDefaultImpl|TestCompilerNoFallbacksForLocalFunctionDefinitionStatement|TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding' -count=1` (pass, `ok ... 0.077s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 88.930s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=28m` (pass, `ok ... 504.874s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=1 timeout 1800s go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=28m` (pass, `ok ... 100.949s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURES=all timeout 1800s go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1 -timeout=28m` (pass, `ok ... 10.032s`)

# 2026-02-19 — Compiler AOT local type-definition statement no-fallback lowering (v12)
- Removed another unsupported-statement fallback source by compiling block-local type declarations (`type`/`struct`/`union`/`interface`) directly in compiled function bodies:
  - added local type statement lowering in `v12/interpreters/go/pkg/compiler/generator_local_type_decls.go`.
  - wired into statement compilation switch in `v12/interpreters/go/pkg/compiler/generator.go`.
- Lowering behavior:
  - local `struct` definitions emit `runtime.StructDefinitionValue` and bind both value + struct table in the current runtime env (`env.Define(...)`, `env.DefineStruct(...)`).
  - local `union` definitions emit `runtime.UnionDefinitionValue` bindings in the current runtime env.
  - local `interface` definitions emit `runtime.InterfaceDefinitionValue` bindings in the current runtime env.
  - local `type` alias statements are compile-time-only in compiled mode once target type rendering succeeds (no fallback wrappers).
  - at this milestone, interface signatures with default impl bodies were still conservatively rejected for the local-lowering path; that sub-gap was closed in the follow-up entry dated 2026-02-19 above.
- Added focused no-fallback regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_local_type_definition_no_fallback_test.go`
  - `TestCompilerNoFallbacksForLocalTypeDefinitions`
  - validates local `type`/`struct`/`union`/`interface` statements compile under `RequireNoFallbacks: true`, emit direct env/runtime bindings, and avoid `CallOriginal("demo.main", ...)`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNoFallbacksForLocalTypeDefinitions|TestCompilerNoFallbacksForLocalFunctionDefinitionStatement|TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding|TestCompilerNoFallbacksStringDefaultImplStaticEmpty|TestCompilerRequireNoFallbacksFails' -count=1` (pass, `ok ... 0.091s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1` (pass, `ok ... 11.632s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=12m` (pass, `ok ... 97.642s`)

# 2026-02-19 — Compiler AOT local function-definition statement no-fallback lowering (v12)
- Removed a remaining static-program fallback source by compiling block-local `fn` statements directly instead of marking them unsupported:
  - added local function statement lowering in `v12/interpreters/go/pkg/compiler/generator_local_functions.go`
    - local `fn name(...) { ... }` now lowers to a local `runtime.Value` callable binding using compiled lambda lowering.
    - binding is installed before body compilation so recursive local functions resolve without fallback.
  - wired into statement compilation switch:
    - `v12/interpreters/go/pkg/compiler/generator.go`
- Refactored compile-context helpers out of `generator.go` into:
  - `v12/interpreters/go/pkg/compiler/generator_context.go`
  - keeps `generator.go` below the 1000-line cap (now 900 lines).
- Added focused no-fallback regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_local_function_definition_no_fallback_test.go`
  - `TestCompilerNoFallbacksForLocalFunctionDefinitionStatement`
  - `TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding`
  - validates recursive local function definition compiles with `RequireNoFallbacks: true`, emits a local runtime function binding, and avoids `CallOriginal("demo.main", ...)`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNoFallbacksForLocalFunctionDefinitionStatement|TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding|TestCompilerNoFallbacksStringDefaultImplStaticEmpty|TestCompilerRequireNoFallbacksFails' -count=1` (pass, `ok ... 0.073s`)
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1` (pass, `ok ... 9.962s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures)$' -count=1 -timeout=12m` (pass, `ok ... 94.385s`)

# 2026-02-19 — Compiler AOT full-matrix timeout hardening (v12)
- Hardened compiler matrix runner to prevent indefinite/stalled suites:
  - `v12/run_compiler_full_matrix.sh` now applies:
    - `go test -timeout` via `ABLE_COMPILER_SUITE_TIMEOUT` (default `25m`),
    - hard wall timeout wrapper via `ABLE_COMPILER_SUITE_WALL_TIMEOUT` (default `30m`, through `timeout(1)` when available).
  - each gate now runs through a shared `run_suite` helper that prints the suite currently running.
- Wired timeout controls into manual CI runs:
  - `.github/workflows/compiler-full-matrix-nightly.yml` now exposes `suite_timeout` and `suite_wall_timeout` workflow-dispatch inputs and maps them to `ABLE_COMPILER_SUITE_TIMEOUT` / `ABLE_COMPILER_SUITE_WALL_TIMEOUT`.
- Updated operator docs:
  - `v12/docs/compiler-full-matrix.md` now documents the new timeout env vars and workflow inputs.
- Validation:
  - `cd v12 && ./run_compiler_full_matrix.sh --help` (shows the new timeout defaults).
  - `cd v12 && ABLE_COMPILER_EXEC_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_SUITE_TIMEOUT=4m ABLE_COMPILER_SUITE_WALL_TIMEOUT=6m ./run_compiler_full_matrix.sh --typecheck-fixtures=strict --skip-fallback-audit` (pass; all four suites complete in ~2s each with timeout controls active).

# 2026-02-19 — Compiler AOT compiled member-dispatch UFCS precedence fix (v12)
- Fixed a compiled-runtime recursion/hang in stdlib compiled CLI tests (`math` + `core/numeric_smoke`) caused by generated `__able_member_get_method(...)` attempting UFCS fallback before interface/member dispatch:
  - symptom: compiled `able-test` stalled after math cases; goroutine dump showed deep recursion in `__able_compiled_fn_floor` (`floor(value)` -> `value.floor()` -> UFCS to `floor(value)`).
  - root cause: in generated member-get-method order, UFCS partial binding ran before `__able_interface_dispatch_member(base, name)`.
  - fix: reordered generated dispatch so interface member resolution runs before UFCS fallback.
  - file: `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_member_get_method_ufcs_precedence_regression_test.go`
  - `TestCompilerPrefersInterfaceDispatchBeforeUFCSInMemberGetMethod`
  - asserts generated `__able_member_get_method` places interface member dispatch before UFCS fallback.
- Closed the remaining stdlib smoke strict-lookup follow-up by promoting math/io/os/process/term/harness fixtures into the default interface-lookup audit set:
  - `v12/interpreters/go/pkg/compiler/compiler_interface_lookup_audit_test.go`
  - added `06_12_20_stdlib_math_core_numeric`, `06_12_22_stdlib_io_temp`, `06_12_23_stdlib_os`, `06_12_24_stdlib_process`, `06_12_25_stdlib_term`, and `06_12_26_stdlib_test_harness_reporters` to `defaultCompilerInterfaceLookupAuditFixtures()`.
- Added bridge-level AOT hardening control for global lookup fallback behavior:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge.go`
  - new runtime toggle:
    - `SetGlobalLookupFallbackEnabled(enabled bool)`
    - guarded fallback sites in `Call`, `Get`, `StructDefinition`, and `CallNamedWithNode`.
  - default remains enabled to preserve current static fixture behavior until broader env seeding/lookup tightening lands.
- Added focused bridge regression coverage for the new fallback toggle:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_test.go`
  - `TestRuntimeCallCanDisableGlobalEnvironmentFallback`
  - `TestCallNamedCanDisableGlobalEnvironmentFallback`
  - `TestGetCanDisableGlobalEnvironmentFallback`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerPrefersInterfaceDispatchBeforeUFCSInMemberGetMethod|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerRemovesTypeRefPointerMemberGetMethodShim' -count=1` (pass, `ok ... 0.061s`)
  - `cd v12/interpreters/go && go test ./cmd/able -run '^TestTestCommandCompiledRunsStdlibMathAndCoreNumericSuites$' -count=1 -timeout=8m -v` (pass, `--- PASS ... (9.36s)`)
  - `cd v12/interpreters/go && go test ./cmd/able -run '^TestTestCommandCompiled' -count=1 -timeout=10m` (pass, `ok ... 212.134s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 556.492s`)
  - `cd v12/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1 -timeout=25m` (pass, `ok ... 34.552s`)
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 64.942s`)
  - `cd v12/interpreters/go && timeout 600s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_12_20_stdlib_math_core_numeric ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=9m` (pass, `ok ... 2.812s`)
  - `cd v12/interpreters/go && timeout 600s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_12_22_stdlib_io_temp,06_12_23_stdlib_os ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=9m` (pass, `ok ... 11.838s`)
  - `cd v12/interpreters/go && timeout 600s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_12_24_stdlib_process ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=9m` (pass, `ok ... 13.286s`)
  - `cd v12/interpreters/go && timeout 600s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_12_25_stdlib_term,06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=9m` (pass, `ok ... 9.415s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=12m` (pass, `ok ... 90.499s`)
  - `cd v12/interpreters/go && timeout 600s go test ./pkg/compiler/bridge -run 'TestRuntimeCallFallsBackToGlobalEnvironment|TestRuntimeCallCanDisableGlobalEnvironmentFallback|TestCallNamedFallsBackToGlobalEnvironment|TestCallNamedCanDisableGlobalEnvironmentFallback|TestGetCanDisableGlobalEnvironmentFallback' -count=1 -timeout=9m` (pass, `ok ... 0.003s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=12m` (pass, `ok ... 86.382s`)
  - bounded re-run (all commands capped below 30m with shell `timeout` + Go `-timeout`):
    - `timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=24m` (pass, `ok ... 514.653s`)
    - `timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -timeout=24m` (pass, `ok ... 567.705s`)
    - `timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=24m` (pass, `ok ... 494.815s`)
    - `timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1 -timeout=24m` (pass, `ok ... 461.014s`)
    - `timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1 -timeout=12m` (pass, `ok ... 30.724s`)
    - `timeout 1200s go test ./cmd/able -run '^TestTestCommandCompiled' -count=1 -timeout=15m` (pass, `ok ... 177.640s`)

# 2026-02-19 — Compiler AOT bridge global-lookup seeding hardening (v12)
- Extended compiler-generated `RegisterIn` initialization to seed `entryEnv` struct definitions from interpreter lookup for all compile-known structs (plus `Array` as a safety net):
  - file: `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - generated helper: `__able_seed_entry_struct_defs(interp, entryEnv)`, invoked during `RegisterIn(...)`.
- Result: strict-total global lookup no longer reports baseline `struct_registry:Array` fallback for static fixtures, and stdlib/fs/process-related registry lookups dropped where structs are compile-known.
- Validation:
  - `cd v12/interpreters/go && timeout 120s go test ./pkg/compiler/bridge -count=1` (pass, `ok ... 0.003s`)
  - `cd v12/interpreters/go && timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=24m` (pass, `ok ... 497.820s`)
  - `cd v12/interpreters/go && timeout 300s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=02_lexical_comments_identifiers ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -v -timeout=4m` (pass, `ok ... 1.957s`)
  - `cd v12/interpreters/go && timeout 420s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES='04_03_type_expression_syntax,04_04_reserved_underscore_types,05_02_array_nested_patterns,05_03_assignment_evaluation_order' ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -v -timeout=6m` (pass, `ok ... 7.961s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=12m` (pass, `ok ... 88.657s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerExecFixtureFallbacks)$' -count=1 -timeout=12m` (pass, `ok ... 106.741s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict go test ./cmd/able -run '^TestTestCommandCompiled' -count=1 -timeout=12m` (pass, `ok ... 187.073s`)
- Follow-up hardening landed for the residual `struct_registry:*` cases:
  - added interpreter bulk seeding helpers:
    - `Interpreter.SeedStructDefinitions(dst *runtime.Environment)` in `v12/interpreters/go/pkg/interpreter/extern_host_coercion.go`
    - `Environment.StructSnapshot()` in `v12/interpreters/go/pkg/runtime/environment.go`
  - updated bridge struct lookup to hydrate missing struct defs from `LookupStructDefinition(name)` into the current env before fallback accounting:
    - `v12/interpreters/go/pkg/compiler/bridge/bridge.go`
  - generated register seeding now calls `interp.SeedStructDefinitions(entryEnv)` before compile-known name seeding:
    - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - added regression coverage:
    - `v12/interpreters/go/pkg/compiler/bridge/bridge_test.go`:
      - `TestStructDefinitionHydratesFromInterpreterLookupWithoutFallbackCounters`
    - `v12/interpreters/go/pkg/interpreter/extern_host_coercion_lookup_struct_test.go`:
      - `TestSeedStructDefinitionsCopiesKnownStructsIntoDestinationEnv`
    - `v12/interpreters/go/pkg/runtime/environment_test.go`:
      - `TestEnvironmentStructSnapshotCopiesCurrentStructBindings`
- Validation (post-follow-up):
  - `cd v12/interpreters/go && timeout 600s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES='06_12_26_stdlib_test_harness_reporters,10_06_interface_generic_param_dispatch,10_16_interface_value_storage,14_01_language_interfaces_index_apply_iterable' ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -v -timeout=8m` (pass, `ok ... 12.196s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=12m` (pass, `ok ... 90.726s`)
  - `cd v12/interpreters/go && timeout 1500s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run '^TestCompilerInterfaceLookupBypassForStaticFixtures$' -count=1 -timeout=24m` (pass, `ok ... 538.970s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run '^(TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerExecFixtureFallbacks)$' -count=1 -timeout=12m` (pass, `ok ... 116.612s`)
  - `cd v12/interpreters/go && timeout 900s env ABLE_TYPECHECK_FIXTURES=strict go test ./cmd/able -run '^TestTestCommandCompiled' -count=1 -timeout=12m` (pass, `ok ... 199.072s`)
- Matrix tooling hardening:
  - `v12/run_compiler_full_matrix.sh` now enforces `ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1` by default for interface-lookup audits.
  - `.github/workflows/compiler-full-matrix-nightly.yml` now exposes and wires `global_lookup_strict_total` (default `1`).
  - `v12/docs/compiler-full-matrix.md` updated with the new env/input and runtime baseline.
  - sanity check: `cd v12 && env ABLE_COMPILER_EXEC_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=10_06_interface_generic_param_dispatch ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1 ./run_compiler_full_matrix.sh --typecheck-fixtures=strict --skip-fallback-audit` (pass, per-gate `ok ... ~2s`).

# 2026-02-16 — Compiler AOT nil-pointer qualified-callable candidate shim cleanup (v12)
- Reduced qualified-callable resolver candidate filtering shim surface by removing the pointer-form nil branch from generated `__able_resolve_qualified_callable(...)`:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - changed candidate type switch from `case runtime.NilValue, *runtime.NilValue` to `case runtime.NilValue`.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_nil_pointer_qualified_callable_shim_regression_test.go`
  - `TestCompilerRemovesNilPointerQualifiedCallableShim`
  - asserts within the resolver’s `switch candidate.(type)` segment that pointer-form nil is absent and value-form nil remains.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesNilPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerQualifiedCallableShim|TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.181s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 60.310s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.823s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 49.002s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 47.122s`)

# 2026-02-16 — Compiler AOT ImplementationNamespace pointer qualified-callable shim cleanup (v12)
- Reduced qualified-callable resolver shim surface by removing the pointer-form `ImplementationNamespace` branch from generated `__able_resolve_qualified_callable(...)` while preserving value-form lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.ImplementationNamespaceValue` branch in the `resolveReceiver` switch.
  - kept `case runtime.ImplementationNamespaceValue` lookup path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_impl_namespace_qualified_callable_shim_regression_test.go`
  - `TestCompilerRemovesImplNamespacePointerQualifiedCallableShim`
  - asserts value-form ImplementationNamespace branch remains and resolver emits exactly one `typed.Methods[tail]` method branch.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesImplNamespacePointerQualifiedCallableShim|TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.170s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.091s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 68.645s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 54.976s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 48.926s`)

# 2026-02-16 — Compiler AOT StructDefinition pointer qualified-callable shim cleanup (v12)
- Reduced qualified-callable resolver shim surface by removing the pointer-form `StructDefinition` branch from generated `__able_resolve_qualified_callable(...)` while preserving value-form static lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.StructDefinitionValue` branch in the `resolveReceiver` switch.
  - kept `case runtime.StructDefinitionValue` lookup path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_structdef_qualified_callable_shim_regression_test.go`
  - `TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim`
  - asserts value-form `StructDefinition` branch remains and resolver emits exactly one `lookupStatic(typed.Node.ID.Name)` branch.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.182s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.803s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 74.379s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 59.970s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 51.614s`)

# 2026-02-16 — Compiler AOT TypeRef pointer qualified-callable shim cleanup (v12)
- Reduced qualified-callable resolver shim surface by removing the pointer-form `TypeRef` branch from generated `__able_resolve_qualified_callable(...)` while preserving value-form static lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.TypeRefValue` branch in the `resolveReceiver` switch.
  - kept `case runtime.TypeRefValue` lookup path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_typeref_qualified_callable_shim_regression_test.go`
  - `TestCompilerRemovesTypeRefPointerQualifiedCallableShim`
  - asserts value-form `TypeRef` branch remains and the resolver emits exactly one `lookupStatic(typed.TypeName)` branch.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.142s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 64.505s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.831s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 50.132s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 47.505s`)

# 2026-02-16 — Compiler AOT ImplementationNamespace pointer member_get_method shim cleanup (v12)
- Reduced member-dispatch shim surface by removing the pointer-form `ImplementationNamespace` branch from generated `__able_member_get_method(...)` while preserving value-form method lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.ImplementationNamespaceValue` branch in member-get-method dispatch.
  - kept `case runtime.ImplementationNamespaceValue` lookup path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_impl_namespace_member_get_method_shim_regression_test.go`
  - `TestCompilerRemovesImplNamespacePointerMemberGetMethodShim`
  - asserts value-form ImplementationNamespace branch remains and exactly one member-get-method `typed.Methods[name]` branch is emitted.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.100s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 3.809s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 62.043s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 49.537s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 45.896s`)

# 2026-02-16 — Compiler AOT StructDefinition pointer member_get_method shim cleanup (v12)
- Reduced member-dispatch shim surface by removing the pointer-form `StructDefinition` lookup branch from generated `__able_member_get_method(...)` while preserving the value-form static lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.StructDefinitionValue` branch in member-get-method dispatch.
  - kept `case runtime.StructDefinitionValue` lookup path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_structdef_member_get_method_shim_regression_test.go`
  - `TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim`
  - asserts value-form StructDefinition branch remains and only one `typed.Node.ID.Name` compiled static lookup branch is emitted in member-get-method dispatch.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.082s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.068s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 111.414s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 86.957s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 51.126s`)

# 2026-02-16 — Compiler AOT TypeRef pointer member_get_method shim cleanup (v12)
- Reduced member-dispatch shim surface by removing the pointer-form `TypeRef` compiled-method lookup branch from generated `__able_member_get_method(...)` while preserving value-form static lookup:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `case *runtime.TypeRefValue` branch that duplicated the static lookup path.
  - preserved `case runtime.TypeRefValue` lookup path for static member resolution.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_typeref_member_get_method_shim_regression_test.go`
  - `TestCompilerRemovesTypeRefPointerMemberGetMethodShim`
  - asserts exactly one `typed.TypeName` compiled-method lookup branch remains in generated member-get-method dispatch and that value-form TypeRef handling is still emitted.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.061s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 61.855s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 58.144s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 57.902s`)

# 2026-02-16 — Compiler AOT package/dynpackage pointer member_get_method shim cleanup (v12)
- Reduced targeted member-dispatch shim surface in generated `__able_member_get_method(...)` while preserving strict lookup-bypass behavior:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - kept value-form package fast path (`case runtime.PackageValue`) for strict-total static fixture lookup bypass.
  - removed pointer-form package fast path (`case *runtime.PackageValue`) from this member-get-method dispatch path.
  - removed pointer-form dynpackage dyn-ref fast path (`case *runtime.DynPackageValue`) from this member-get-method dispatch path.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_package_member_get_method_shim_regression_test.go`
  - `TestCompilerRemovesPackagePublicMemberGetMethodShim`
  - asserts value-form package fast path remains, pointer-form package branch is absent, and bridge fallback path remains emitted.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (pass, `ok ... 0.041s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 56.515s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 4.075s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 44.352s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 44.222s`)

# 2026-02-16 — Compiler AOT Error.value member_get_method shim cleanup (v12)
- Removed the legacy `Error.value` hardcoded branch from generated `__able_member_get_method(...)` so method dispatch no longer bypasses callable/method lookup rules for error payload values:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed:
    - `errorValue := runtime.ErrorValue{}`
    - `hasErrorValue := false`
    - `if hasErrorValue && name == "value" { ... }`
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_error_value_member_get_method_shim_regression_test.go`
  - `TestCompilerRemovesErrorValueMemberGetMethodShim`
  - asserts the legacy shim branch string is absent and `Error.message`/`Error.cause` builtin compiled-method registration remains present.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRemovesErrorValueMemberGetMethodShim|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerInterfaceLookupBypassForStaticFixtures' -count=1` (pass, `ok ... 58.132s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_07_channel_mutex_error_types go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (pass, `ok ... 2.148s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 45.562s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (pass, `ok ... 45.452s`)

# 2026-02-16 — Compiler AOT HashMap member_set shim cleanup (v12)
- Audited `__able_member_set(...)` type-specific shims and removed an unreachable legacy `HashMap.handle` read branch that shadowed the actual setter branch:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed duplicate branch that read/returned current handle (`val, ok := inst.Fields["handle"]`) before the setter branch.
  - retained the actual setter branch (`hash map handle must be positive`, `HashMapStoreEnsureHandle`, and `inst.Fields["handle"] = value`).
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_hashmap_member_set_shim_regression_test.go`
  - `TestCompilerMemberSetHashMapHandleUsesSetterBranch`
  - asserts legacy read-branch pattern is absent and setter assignment/validation strings remain.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerMemberSetHashMapHandleUsesSetterBranch|TestCompilerRegistersBuiltinAwaitNamedCalls|TestCompilerRegistersBuiltinFutureNamedCalls|TestCompilerRegistersBuiltinFutureMemberMethods|TestCompilerRegistersBuiltinDynPackageMemberMethods|TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.203s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 51.146s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 64.003s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 288.489s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 553.109s`)

# 2026-02-16 — Compiler AOT await named-call shim replacement with compiled call registration (v12)
- Removed hardcoded await helper switch branches from generated `__able_call_named(...)` and moved both await helpers to builtin compiled-call registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed direct branches for:
    - `__able_await_default`
    - `__able_await_sleep_ms`
- Added builtin compiled-call wrappers + registration:
  - wrappers:
    - `__able_builtin_named_await_default(...)`
    - `__able_builtin_named_await_sleep_ms(...)`
  - registration entries in `__able_register_builtin_compiled_calls(...)`:
    - `__able_register_compiled_call(env, "__able_await_default", -1, 0, "", __able_builtin_named_await_default)`
    - `__able_register_compiled_call(env, "__able_await_sleep_ms", -1, 1, "", __able_builtin_named_await_sleep_ms)`
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_await_named_call_registration_test.go`
  - `TestCompilerRegistersBuiltinAwaitNamedCalls`
  - asserts helper emission + registration and absence of legacy await switch branches in `__able_call_named`.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinAwaitNamedCalls|TestCompilerRegistersBuiltinFutureNamedCalls|TestCompilerRegistersBuiltinFutureMemberMethods|TestCompilerRegistersBuiltinDynPackageMemberMethods|TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.181s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 48.844s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 61.663s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 269.375s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 553.991s`)

# 2026-02-16 — Compiler AOT future_* named-call shim replacement with compiled call registration (v12)
- Removed hardcoded `future_*` switch branches from generated `__able_call_named(...)` and moved those builtins to compiled-call registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed direct branches for:
    - `future_yield`
    - `future_cancelled`
    - `future_flush`
    - `future_pending_tasks`
- Added builtin compiled-call wrappers + registration:
  - wrappers:
    - `__able_builtin_named_future_yield(...)`
    - `__able_builtin_named_future_cancelled(...)`
    - `__able_builtin_named_future_flush(...)`
    - `__able_builtin_named_future_pending_tasks(...)`
  - registration helper:
    - `__able_register_builtin_compiled_calls(entryEnv, interp)`
    - seeds compiled calls via `__able_register_compiled_call(...)` for the four `future_*` names.
- Wired builtin compiled-call registration into startup:
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `RegisterIn(...)` now invokes `__able_register_builtin_compiled_calls(entryEnv, interp)` before builtin compiled method registration.
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_future_named_call_registration_test.go`
  - `TestCompilerRegistersBuiltinFutureNamedCalls`
  - asserts helper emission + registration + `RegisterIn` wiring and absence of legacy `future_*` `__able_call_named` switch branches.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinFutureNamedCalls|TestCompilerRegistersBuiltinFutureMemberMethods|TestCompilerRegistersBuiltinDynPackageMemberMethods|TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.169s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 55.395s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 68.294s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 267.627s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 541.308s`)

# 2026-02-16 — Compiler AOT Future member shim replacement with compiled registration (v12)
- Removed direct `__able_future_member_value(...)` shim call sites from generated member lookup paths and moved Future member handling to builtin compiled-method registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed call-site branches from:
    - `__able_member_get(...)`
    - `__able_member_get_method(...)`
- Added builtin compiled helpers and registrations for Future methods:
  - helpers:
    - `__able_builtin_future_receiver(...)`
    - `__able_builtin_future_status(...)`
    - `__able_builtin_future_value(...)`
    - `__able_builtin_future_cancel(...)`
    - `__able_builtin_future_is_ready(...)`
    - `__able_builtin_future_register(...)`
    - `__able_builtin_future_commit(...)`
    - `__able_builtin_future_is_default(...)`
  - registrations:
    - `Future.status`, `Future.value`, `Future.cancel`, `Future.is_ready`, `Future.register`, `Future.commit`, `Future.is_default`
- Updated runtime type-name mapping for compiled method dispatch:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_interface_member.go`
  - added `*runtime.FutureValue` => `"Future"`.
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_future_member_compiled_registration_test.go`
  - `TestCompilerRegistersBuiltinFutureMemberMethods`
  - asserts helper emission + registration and confirms legacy `__able_future_member_value` member-lookup call-site shim strings are absent.
- Removed the now-dead legacy helper implementation after call-site migration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_future.go`
  - deleted `__able_future_member_value(...)` to keep runtime codegen aligned with compiled-method registration.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinFutureMemberMethods|TestCompilerRegistersBuiltinDynPackageMemberMethods|TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.130s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 52.055s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 65.033s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 312.632s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 632.595s`)

# 2026-02-16 — Compiler AOT DynPackage def/eval shim replacement with compiled registration (v12)
- Removed direct `DynPackage.def` / `DynPackage.eval` bridge-member shim branches from `__able_member_get_method` and moved both to builtin compiled-method registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - dyn package branch now keeps `DynRefValue` handling for non-`def`/`eval` members while allowing compiled-method dispatch for `def`/`eval`.
- Added builtin compiled helpers and registration entries:
  - `__able_builtin_dynpackage_member_call(...)`
  - `__able_builtin_dynpackage_def(...)`
  - `__able_builtin_dynpackage_eval(...)`
  - `__able_register_compiled_method("DynPackage", "def", true, 1, 1, __able_builtin_dynpackage_def)`
  - `__able_register_compiled_method("DynPackage", "eval", true, 1, 1, __able_builtin_dynpackage_eval)`
  - helper delegates invocation through `bridge.CallValue(...)` so dynamic package method arity/behavior stays aligned with interpreter semantics.
- Extended runtime type-name mapping so compiled method lookup can bind on dynamic package receivers:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_interface_member.go`
  - added `runtime.DynPackageValue` / `*runtime.DynPackageValue` => `"DynPackage"`.
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_dynpackage_member_compiled_registration_test.go`
  - `TestCompilerRegistersBuiltinDynPackageMemberMethods`
  - asserts helper emission + registration and absence of legacy direct `def/eval` shim branch strings.
- Regression found and fixed during validation:
  - initial migration registered `DynPackage.def/eval` with arity `0`, which broke fixture `06_10_dynamic_metaprogramming_package_object` (`first 42` only); corrected to arity `1` with delegated call-through.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_10_dynamic_metaprogramming_package_object go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -v` (pass)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinDynPackageMemberMethods|TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.120s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 52.137s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 65.169s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 303.587s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 597.530s`)

# 2026-02-16 — Compiler AOT Iterator member shim replacement with compiled registration (v12)
- Removed legacy `Iterator.next` native-method shim construction from `__able_member_get_method` and moved it to builtin compiled-method registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed inline branch-local `runtime.NativeFunctionValue` construction for iterator `next`.
- Added builtin compiled helper and registration entry:
  - `__able_builtin_iterator_next(...)`
  - `__able_register_compiled_method("Iterator", "next", true, 0, 0, __able_builtin_iterator_next)`
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_iterator_member_compiled_registration_test.go`
  - `TestCompilerRegistersBuiltinIteratorMemberMethods`
  - asserts helper emission + method registration and absence of legacy `Iterator.next` member shim branch/constructor.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinIteratorMemberMethods|TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.108s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 49.985s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 63.456s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 277.878s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 565.078s`)

# 2026-02-16 — Compiler AOT Error member shim replacement with compiled registration (v12)
- Removed legacy `Error.message` / `Error.cause` native-method shim construction from `__able_member_get_method` and moved both to builtin compiled-method registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed inline branch-local `runtime.NativeFunctionValue` construction for:
    - `messageMethod := runtime.NativeFunctionValue{...}`
    - `causeMethod := runtime.NativeFunctionValue{...}`
  - preserved direct payload field behavior for `error.value` access.
- Added builtin compiled helpers and registration entries:
  - `__able_builtin_error_message(...)`
  - `__able_builtin_error_cause(...)`
  - `__able_register_compiled_method("Error", "message", true, 0, 0, __able_builtin_error_message)`
  - `__able_register_compiled_method("Error", "cause", true, 0, 0, __able_builtin_error_cause)`
- Added focused compiler regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_error_member_compiled_registration_test.go`
  - `TestCompilerRegistersBuiltinErrorMemberMethods`
  - asserts helper emission + method registration and absence of legacy `messageMethod`/`causeMethod` shim branches.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass, `ok ... 0.081s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 53.439s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 67.702s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 292.504s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 567.316s`)

# 2026-02-16 — Compiler AOT integer member shim replacement with compiled registration (v12)
- Removed hardcoded integer runtime member lookup shims for `clone`/`to_string` and replaced them with builtin compiled-method registration:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - removed `__able_member_get_method` integer branches for:
    - `if _, ok := base.(runtime.IntegerValue); ok { ... }`
    - `if intPtr, ok := base.(*runtime.IntegerValue); ok && intPtr != nil { ... }`
  - added generated builtin helpers:
    - `__able_builtin_integer_clone(...)`
    - `__able_builtin_integer_to_string(...)`
    - `__able_register_builtin_compiled_methods()`
  - registration now seeds integer method thunks for `i8`, `i16`, `i32`, `i64`, `i128`, `u8`, `u16`, `u32`, `u64`, `u128`, `isize`, `usize`.
- Wired builtin method registration into compiler startup:
  - `v12/interpreters/go/pkg/compiler/generator_render_functions.go`
  - `RegisterIn(...)` now calls `__able_register_builtin_compiled_methods()` before package method/impl registration.
- Added focused regression coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_integer_member_compiled_registration_test.go`
  - `TestCompilerRegistersBuiltinIntegerMemberMethods`
  - asserts generated source includes builtin helper emission + registration call and no longer emits legacy integer shim branches.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersBuiltinIntegerMemberMethods|TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (pass)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 51.155s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 64.358s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass, `ok ... 266.656s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 539.938s`)

# 2026-02-16 — Compiler AOT String static method lowering regression guard (v12)
- Updated method lowering so `methods String.fn from_bytes_unchecked(...) -> String` compiles as a typed static method registration path (struct return) instead of relying on runtime member-lookup shims:
  - `v12/interpreters/go/pkg/compiler/generator_methods.go`
  - removed the generic runtime-value return forcing path and replaced it with targeted typed return lowering for `String.from_bytes_unchecked`.
- Added a focused compiler regression test that asserts static compiled-method registration for this path:
  - `v12/interpreters/go/pkg/compiler/compiler_string_method_registration_test.go`
  - `TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod`
  - checks generated source contains `__able_register_compiled_method("String", "from_bytes_unchecked", false, ...)`.
- Added a focused no-fallback regression for String-shadowing impl dispatch:
  - `v12/interpreters/go/pkg/compiler/compiler_string_impl_regression_test.go`
  - `TestCompilerNoFallbacksStringDefaultImplStaticEmpty`
  - verifies `impl Default for String { fn default() -> String { String.empty() } }` compiles under `RequireNoFallbacks: true` (guard against `impl Default for String.default` fallback regressions).
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run 'TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 58.297s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (pass, `ok ... 495.357s`)
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (pass, `ok ... 46.704s`)

# 2026-02-16 — Compiler AOT strict-total lookup stabilization + all-fixture baseline (v12)
- Fixed compiled dyn-package method dispatch for `def`/`eval`:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - replaced the broken `dyn.def` shortcut path with direct `bridge.MemberGet(__able_runtime, dynPkg, "def"/"eval")` resolution so compiled code reuses interpreter-native `DynPackageValue` bound methods.
- Added direct static handling for `String.from_bytes_unchecked` to eliminate remaining strict-total member-lookup misses:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - generated helper `__able_static_string_from_bytes_unchecked_method(...)`
  - hooked `runtime.StructDefinitionValue` / `*runtime.StructDefinitionValue` member resolution for `"String"."from_bytes_unchecked"` before interpreter fallback lookup.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES='06_10_dynamic_metaprogramming_package_object,06_12_13_stdlib_collections_persistent_sorted_queue,06_12_21_stdlib_fs_path,06_12_22_stdlib_io_temp,06_12_24_stdlib_process,07_02_01_verbose_anonymous_fn,13_04_import_alias_selective_dynimport' ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=20m` (pass, `ok ... 644.285s`).
- Hardened default interface-lookup audit coverage (no env override required) by adding additional regression fixtures in `defaultCompilerInterfaceLookupAuditFixtures()`:
  - `06_10_dynamic_metaprogramming_package_object`
  - `06_12_21_stdlib_fs_path`
  - `13_04_import_alias_selective_dynimport`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (pass, `ok ... 58.411s`).

# 2026-02-14 — Compiler AOT strict interface-lookup bypass audit + markers (v12)
- Added compiler bridge instrumentation for interpreter member-lookup fallback paths:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge.go`
  - new counters and helpers:
    - `ResetMemberGetPreferMethodsCounters()`
    - `MemberGetPreferMethodsStats()`
  - `CallNamed` now supports a generated qualified-callable resolver hook (`SetQualifiedCallableResolver`) before interpreter member lookup, while still routing fallback qualified member lookup through `MemberGetPreferMethods(...)` when unresolved.
- Added bridge unit coverage for lookup counters:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_test.go`
  - `TestMemberGetPreferMethodsCounters`
  - `TestCallNamedWithQualifiedResolverBypassesMemberLookup`
- Extended compiler fixture harness marker support:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - new env-gated stderr markers:
    - `__ABLE_MEMBER_LOOKUP_CALLS`
    - `__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS`
  - counters are reset before `RunRegisteredMain(...)`.
- Tightened strict interface dispatch behavior in generated runtime calls:
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`
  - when `__able_interface_dispatch_strict` is enabled and an interface method cannot be resolved by compiled dispatch, code now raises immediately instead of falling through to interpreter member lookup.
  - added shared compiled-thunk invocation helper (`__able_call_compiled_thunk`) in `__able_call_value` that accepts both raw func thunks and `interpreter.CompiledThunk`, and expanded bound-method fast paths for `runtime.BoundMethodValue`/`*runtime.BoundMethodValue` when the wrapped callable carries compiled thunk metadata.
- Added dedicated static fixture audit gate:
  - `v12/interpreters/go/pkg/compiler/compiler_interface_lookup_audit_test.go`
  - `TestCompilerInterfaceLookupBypassForStaticFixtures`
  - defaults now cover interface-heavy static fixtures across:
    - `06_01`, `06_03`, `07_04`
    - `10_01` through `10_17`
    - `14_01` language/index-apply and operator arithmetic/comparison
  - configurable via `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES` (`all` supported), and now also asserts `__ABLE_BOUNDARY_FALLBACK_CALLS=0` for these static fixtures.
  - optional strict-total mode (`ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1`) additionally asserts `__ABLE_MEMBER_LOOKUP_CALLS=0`; current failures in that mode now show `__ABLE_BOUNDARY_EXPLICIT_CALLS=0` for the focused probe and are concentrated in non-interface member resolution (impl/interface method lookup) rather than `call_value` bridge crossings.
- Wired new audit into full-matrix tooling/CI:
  - `v12/run_compiler_full_matrix.sh`
  - `v12/run_all_tests.sh`
  - `.github/workflows/compiler-full-matrix-nightly.yml`
  - `v12/docs/compiler-full-matrix.md`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler/bridge -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_GOCACHE=$(pwd)/.gocache ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=10_02_impl_specificity_named_overrides ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1 ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (expected failure; confirms remaining non-interface member-lookup path with `__ABLE_BOUNDARY_EXPLICIT_CALLS=0`)
  - `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_01_compiler_type_qualified_method,06_03_operator_overloading_interfaces,07_04_apply_callable_interface,10_01_interface_defaults_composites,10_02_impl_specificity_named_overrides,10_02_impl_where_clause,10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_05_interface_named_impl_defaults,10_06_interface_generic_param_dispatch,10_07_interface_default_chain,10_08_interface_default_override,10_09_interface_named_impl_inherent,10_10_interface_inheritance_defaults,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_13_interface_param_generic_args,10_14_interface_return_generic_args,10_15_interface_default_generic_method,10_16_interface_value_storage,10_17_interface_overload_dispatch,14_01_language_interfaces_index_apply_iterable,14_01_operator_interfaces_arithmetic_comparison go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerDynamicBoundary -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_01_compiler_type_qualified_method,10_05_interface_named_impl_defaults,10_17_interface_overload_dispatch,14_01_language_interfaces_index_apply_iterable ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict`
  - `ABLE_FIXTURE_FILTER=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=06_01_compiler_type_qualified_method,10_05_interface_named_impl_defaults,10_17_interface_overload_dispatch,14_01_language_interfaces_index_apply_iterable ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./run_all_tests.sh --version=v12 --fixture --compiler-full-matrix --typecheck-fixtures=strict`

# 2026-02-14 — Compiler AOT text/string compiled strict coverage expansion (v12)
- Expanded compiled strict/no-fallback stdlib gate coverage to text/string suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - added `TestTestCommandCompiledRunsStdlibTextStringSuites` covering:
    - `v12/stdlib/tests/text/string_methods.test.able`
    - `v12/stdlib/tests/text/string_split.test.able`
    - `v12/stdlib/tests/text/string_builder.test.able`
    - `v12/stdlib/tests/text/string_smoke.test.able`
- Expanded build precompile discovery assertions for text packages:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go`
  - added expectations for:
    - `able.text.string`
    - `able.text.regex`
    - `able.text.ascii`
    - `able.text.automata`
    - `able.text.automata_dsl`
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibTextStringSuites' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `ABLE_FIXTURE_FILTER=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./run_all_tests.sh --version=v12 --fixture --compiler-full-matrix --typecheck-fixtures=strict`

# 2026-02-13 — Compiler full-matrix operator docs (v12)
- Added dedicated operator-facing docs:
  - `v12/docs/compiler-full-matrix.md`
  - covers:
    - local command paths (`run_compiler_full_matrix.sh`, `run_all_tests.sh --compiler-full-matrix`)
    - env override knobs for narrowed/full sweeps
    - workflow dispatch inputs
    - current runtime profile baseline for `...=all` sweeps
- Added doc pointer in:
  - `v12/README.md`
- Validation:
  - `rg -n "compiler-full-matrix\\.md|workflow_dispatch" v12/README.md v12/docs/compiler-full-matrix.md .github/workflows/compiler-full-matrix-nightly.yml`

# 2026-02-13 — CI workflow for compiler full-matrix sweeps (v12)
- Added GitHub Actions workflow:
  - `.github/workflows/compiler-full-matrix-nightly.yml`
  - schedule: daily (`20 6 * * *`) plus `workflow_dispatch`.
  - runs `v12/run_compiler_full_matrix.sh` with configurable fixture env overrides (defaults to `all`).
  - sets Go via `v12/interpreters/go/go.mod` and enables module cache.
- Validation:
  - `ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict`
  - `ABLE_FIXTURE_FILTER=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./run_all_tests.sh --version=v12 --fixture --compiler-full-matrix --typecheck-fixtures=strict`

# 2026-02-13 — Compiler full-matrix wrapper target for nightly/manual sweeps (v12)
- Added dedicated compiler full-matrix runner:
  - `v12/run_compiler_full_matrix.sh`
  - runs:
    - `TestCompilerExecFixtures`
    - `TestCompilerStrictDispatchForStdlibHeavyFixtures`
    - `TestCompilerBoundaryFallbackMarkerForStaticFixtures`
  - defaults to `ABLE_COMPILER_EXEC_FIXTURES=all`, `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all`, `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all`, with env overrides supported for narrowed local sweeps.
- Added `run_all_tests` target flag:
  - `v12/run_all_tests.sh --compiler-full-matrix`
  - executes normal v12 test flow, then invokes `run_compiler_full_matrix.sh`.
  - fixed option wiring to preserve caller fixture env overrides (`...=${...:-all}`) instead of force-overwriting to `all`.
- Documentation updates:
  - `README.md` and `v12/README.md` now include full-matrix command examples.
- Validation:
  - `bash -n v12/run_compiler_full_matrix.sh v12/run_all_tests.sh`
  - `./v12/run_compiler_full_matrix.sh --help`
  - `./v12/run_all_tests.sh --help`
  - `ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict`
  - `ABLE_FIXTURE_FILTER=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters ./v12/run_all_tests.sh --fixture --compiler-full-matrix --typecheck-fixtures=strict`

# 2026-02-13 — Compiler AOT full-matrix `...=all` sweep + strict runner expectation fix (v12)
- Ran explicit full-matrix compiler fixture sweeps (separate from reduced default CI-speed gates):
  - `ABLE_COMPILER_EXEC_FIXTURES=all` with `TestCompilerExecFixtures` (~506s) passed.
  - `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all` with `TestCompilerStrictDispatchForStdlibHeavyFixtures` (~533s) passed.
  - `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all` with `TestCompilerBoundaryFallbackMarkerForStaticFixtures` (~463s) passed.
- Fixed strict-dispatch runner behavior for full fixture coverage:
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `runCompilerStrictDispatchFixture` now:
    - enforces `__ABLE_STRICT=true` marker presence as before,
    - but validates fixture outcomes using manifest expectations (`stdout`, `stderr`, `exit`) instead of failing unconditionally on non-zero exits.
  - this allows strict-dispatch auditing across fixtures that intentionally assert runtime/type errors.
- Post-fix default gate sanity:
  - `go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (~160s) passed.
  - `go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1` (~63s) passed.
  - `go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (~63s) passed.

# 2026-02-13 — Compiler AOT strict/boundary default suite runtime reduction (v12)
- Reduced default fixture sets for strict-dispatch + boundary-audit gates:
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
  - both now use shared high-signal defaults from:
    - `v12/interpreters/go/pkg/compiler/compiler_heavy_fixture_defaults_test.go`
- Fixed full-matrix opt-in semantics:
  - `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all` and `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all` now use `collectExecFixtures(...)` directly (true full fixture discovery), independent of the reduced default exec suite.
- Improved fixture list parsing consistency:
  - strict-dispatch + boundary-audit selectors now accept comma/semicolon/whitespace-separated lists.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -count=1`
- Result:
  - strict-dispatch default gate ~54s.
  - boundary-audit default gate ~54s.
  - full `./pkg/compiler` package ~377s (previously ~386s after initial strict/boundary reduction, ~489s after exec-fixture reduction, and earlier timed out at default 10m).

# 2026-02-13 — Compiler AOT exec fixture default suite runtime reduction (v12)
- Reduced default `TestCompilerExecFixtures` matrix to a high-signal subset:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - extracted default fixture list into `defaultCompilerExecFixtures()`.
  - kept full fixture matrix available via existing env control:
    - `ABLE_COMPILER_EXEC_FIXTURES=all` (filesystem discovery with `collectExecFixtures`).
- Scope preserved in default suite:
  - entry/interop smoke fixtures.
  - core compiler control-flow/pattern/rescue/concurrency fixtures.
  - interface/import/regex heavy fixtures.
  - complete `06_12_01` through `06_12_26` stdlib compiled fixture set.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -count=1`
- Result:
  - `TestCompilerExecFixtures` completed in ~154s.
  - full `./pkg/compiler` package completed in ~489s (previously timed out at Go default 10m).

# 2026-02-13 — Compiler AOT boundary marker strictness fix for call_original parity (v12)
- Fixed dynamic boundary parity regression introduced by strict fixture no-fallback defaults:
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`
  - `TestCompilerDynamicBoundaryCallOriginalMarkers` now sets `ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=0` for that test only.
  - rationale: this test intentionally uses an uncompileable function body to exercise explicit `call_original` boundary markers; strict no-fallback compilation should stay enabled by default elsewhere.
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerDynamicBoundaryCallOriginalMarkers -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerConcurrencyParityFixtures -count=1`
- Follow-up note:
  - package timeout pressure from `TestCompilerExecFixtures` was subsequently reduced by narrowing the default exec fixture suite while keeping `ABLE_COMPILER_EXEC_FIXTURES=all` for full-matrix runs.

# 2026-02-13 — Compiler AOT stdlib harness/reporters strict smoke gate (v12)
- Added new stdlib smoke suite for strict compiled harness/reporters coverage:
  - `v12/stdlib/tests/harness_reporters_smoke.test.able`
  - smoke module exercises:
    - `able.test.harness` discovery and run flow (`discover_all`, `run_all`).
    - `able.test.reporters` doc/progress reporter output buffering.
  - smoke module now clears example registrations at start/end so `able test --compiled` remains deterministic (`able test: no tests to run`).
- Added new exec fixture `v12/fixtures/exec/06_12_26_stdlib_test_harness_reporters`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - harness discovery returns descriptors for fixture-defined examples.
    - `DocReporter` and `ProgressReporter` produce output through custom emit buffers.
    - reporter run paths complete without framework failures.
- Fixed reporter method selector lookup in fixture/smoke modules:
  - `v12/stdlib/tests/harness_reporters_smoke.test.able`
  - `v12/fixtures/exec/06_12_26_stdlib_test_harness_reporters/main.able`
  - both modules now import `finish` from `able.test.reporters` so `progress.finish()` resolves under interpreter/compiled execution.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_26_stdlib_test_harness_reporters`.
- Added strict compiled CLI gate for harness/reporters smoke suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibHarnessReportersSmokeSuite`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/harness_reporters_smoke.test.able`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_26_stdlib_test_harness_reporters go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibHarnessReportersSmokeSuite' -count=1`

# 2026-02-13 — Compiler AOT stdlib term strict smoke gate (v12)
- Added new stdlib smoke suite for fast strict compiled gating:
  - `v12/stdlib/tests/term_smoke.test.able`
  - smoke module validates `able.term` tty/size/raw-mode helper behavior with non-interactive checks.
- Added new exec fixture `v12/fixtures/exec/06_12_25_stdlib_term`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - `term.is_tty` boolean behavior.
    - `term.try_size` and `term.try_set_raw_mode` typed `IOError` fallback behavior.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_25_stdlib_term`.
- Added strict compiled CLI gate for term smoke suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibTermSmokeSuite`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/term_smoke.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.term`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_25_stdlib_term go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_25_stdlib_term go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_25_stdlib_term go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_25_stdlib_term go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibTermSmokeSuite' -count=1`

# 2026-02-13 — Compiler AOT stdlib process strict smoke gate (v12)
- Fixed strict compiled `process.spawn` host coercion panic:
  - `v12/interpreters/go/pkg/interpreter/extern_host_coercion.go`
  - array coercion for extern struct fields now tolerates interface-typed host targets (used by struct-to-map conversion), avoiding `reflect: Elem of invalid type interface {}`.
  - nullable field coercion now also tolerates interface-typed host targets by delegating to inner-type coercion for non-`nil` values.
- Added interpreter regression test for extern struct-array coercion:
  - `v12/interpreters/go/pkg/interpreter/interpreter_extern_test.go`
  - new test: `TestExternStructArrayFieldCoercesIntoHostMap`
  - new test: `TestExternStructNullableArrayFieldCoercesIntoHostMap`
- Added new stdlib smoke suite for strict compiled process coverage:
  - `v12/stdlib/tests/process_smoke.test.able`
  - covers spawn/wait/stdio output, method-chain process-spec setup (`with_cwd`, `with_env` with selector imports), and missing-command `IOError(NotFound)` mapping.
- Added new exec fixture `v12/fixtures/exec/06_12_24_stdlib_process`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - `able.process` spawn/wait/stdio behavior for a successful command with method-chain `ProcessSpec` setup.
    - typed `IOError(NotFound)` behavior from `process.try_spawn` on missing commands.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_24_stdlib_process`.
- Added strict compiled CLI gate for process smoke suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibProcessSmokeSuite`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/process_smoke.test.able`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestExternStructArrayFieldCoercesIntoHostMap|TestExternStructNullableArrayFieldCoercesIntoHostMap' -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_24_stdlib_process go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_24_stdlib_process go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_24_stdlib_process go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_24_stdlib_process go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibProcessSmokeSuite' -count=1`
- Note:
  - `with_cwd` / `with_env` member calls require selector imports in scope (per method lookup rules), so the smoke/fixture modules import `able.process.{with_cwd, with_env}` when exercising method-chain coverage.

# 2026-02-13 — Compiler AOT stdlib os strict smoke gate (v12)
- Added new stdlib smoke suite for fast strict compiled gating:
  - `v12/stdlib/tests/os_smoke.test.able`
  - smoke module validates `able.os` args/env/cwd/chdir/try_chdir/temp-dir behavior.
- Added new exec fixture `v12/fixtures/exec/06_12_23_stdlib_os`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - `able.os` env mutation/readback and cwd/chdir behavior.
    - typed `IOError(NotFound)` behavior from `os.try_chdir` on missing paths.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_23_stdlib_os`.
- Added strict compiled CLI gate for os smoke suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibOsSmokeSuite`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/os_smoke.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.os`
    - `able.process`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_23_stdlib_os go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_23_stdlib_os go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_23_stdlib_os go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_23_stdlib_os go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibOsSmokeSuite' -count=1`
- Follow-up status:
  - this blocker was resolved in the `06_12_24_stdlib_process` slice via extern host coercion fixes; remaining process work is method-chain coverage for `ProcessSpec.with_cwd` / `ProcessSpec.with_env` under strict compiled lookup.

# 2026-02-13 — Compiler AOT stdlib io/temp strict smoke gates (v12)
- Added new stdlib smoke suite for fast strict compiled gating:
  - `v12/stdlib/tests/io_smoke.test.able`
  - smoke module validates `able.io` read/write helpers plus `able.io.temp` temp file lifecycle.
- Added new exec fixture `v12/fixtures/exec/06_12_22_stdlib_io_temp`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - `able.io` string/bytes conversion plus `read_all`/`write_all` helper semantics.
    - `able.io.temp` temp directory/file creation and cleanup behavior.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_22_stdlib_io_temp`.
- Added strict compiled CLI gate for io smoke suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibIoSmokeSuite`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/io_smoke.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.io.temp`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_22_stdlib_io_temp go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_22_stdlib_io_temp go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_22_stdlib_io_temp go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_22_stdlib_io_temp go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibIoSmokeSuite' -count=1`

# 2026-02-13 — Compiler AOT stdlib fs/path strict smoke gates (v12)
- Added new stdlib smoke suites for fast strict compiled gating:
  - `v12/stdlib/tests/fs_smoke.test.able`
  - `v12/stdlib/tests/path_smoke.test.able`
  - both are non-framework smoke modules (assertion-style `main()`), so `able test --compiled` reports `able test: no tests to run` on success.
- Added new exec fixture `v12/fixtures/exec/06_12_21_stdlib_fs_path`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers:
    - `able.io.path` normalization/join/extension behavior.
    - `able.fs` write/read/rename/read_dir/remove behavior on temp paths.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_21_stdlib_fs_path`.
- Added strict compiled CLI gate for fs/path smoke suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibFsAndPathSmokeSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/fs_smoke.test.able`
    - `v12/stdlib/tests/path_smoke.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.fs`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_21_stdlib_fs_path go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_21_stdlib_fs_path go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_21_stdlib_fs_path go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_21_stdlib_fs_path go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibFsAndPathSmokeSuites' -count=1`

# 2026-02-13 — Compiler AOT stdlib math/core-numeric strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_20_stdlib_math_core_numeric`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers stdlib math/core numeric behavior:
    - `able.math` integer-safe helpers (`abs_i64`, `sign_i64`, `clamp_i64`, `gcd`, `lcm`).
    - `able.core.numeric` conversion helpers (`to_r`, `Ratio.to_i32`) including fractional conversion error path.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_20_stdlib_math_core_numeric`.
- Added strict compiled CLI gate for stdlib math/core numeric suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibMathAndCoreNumericSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/math.test.able`
    - `v12/stdlib/tests/core/numeric_smoke.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.math`
    - `able.core.numeric`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_20_stdlib_math_core_numeric go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_20_stdlib_math_core_numeric go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_20_stdlib_math_core_numeric go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_20_stdlib_math_core_numeric go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibMathAndCoreNumericSuites' -count=1`

# 2026-02-13 — Compiler AOT stdlib concurrency channel/mutex/concurrent_queue strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_19_stdlib_concurrency_channel_mutex_queue`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers stdlib concurrency wrappers:
    - `Channel` send/receive/close/iterable behavior through `able.concurrency`.
    - `Mutex` `with_lock` and manual lock/unlock behavior through `able.concurrency`.
    - `ConcurrentQueue` enqueue/dequeue/try/close semantics through `able.concurrency.concurrent_queue`.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_19_stdlib_concurrency_channel_mutex_queue`.
- Added strict compiled CLI gate for stdlib concurrency suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibConcurrencyChannelMutexAndQueueSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/concurrency/channel_mutex.test.able`
    - `v12/stdlib/tests/concurrency/concurrent_queue.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.concurrency`
    - `able.concurrency.concurrent_queue`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_19_stdlib_concurrency_channel_mutex_queue go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_19_stdlib_concurrency_channel_mutex_queue go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_19_stdlib_concurrency_channel_mutex_queue go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_19_stdlib_concurrency_channel_mutex_queue go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibConcurrencyChannelMutexAndQueueSuites' -count=1`

# 2026-02-13 — Compiler AOT collections array/range strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_18_stdlib_collections_array_range`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers array/range behavior through stdlib wrappers:
    - `Array` push/push_all/get/write_slot/pop/clear helpers and length/optional accessors.
    - `RangeFactory` inclusive/exclusive ranges via stdlib `able.collections.range` re-exports.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_18_stdlib_collections_array_range`.
- Added strict compiled CLI gate for array/range smoke suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsArrayAndRangeSmokeSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/collections/array_smoke.test.able`
    - `v12/stdlib/tests/collections/range_smoke.test.able`
  - asserts successful run and expected `able test: no tests to run` output for smoke modules.
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.array`
    - `able.collections.range`
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_FIXTURE_FILTER=06_12_18_stdlib_collections_array_range go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_18_stdlib_collections_array_range go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_18_stdlib_collections_array_range go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_18_stdlib_collections_array_range go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsArrayAndRangeSmokeSuites' -count=1`

# 2026-02-13 — Compiler AOT collections bit_set/heap strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_17_stdlib_collections_bit_set_heap`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers bit-set/heap behavior through stdlib wrappers:
    - `BitSet` set/reset/flip/contains, `each`, `Iterable` iteration, and clear semantics.
    - `Heap` min-heap push/pop ordering, `peek`, `len`, and empty-state semantics.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_17_stdlib_collections_bit_set_heap`.
- Added strict compiled CLI gate for bit_set/heap suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsBitSetAndHeapSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/bit_set.test.able`
    - `v12/stdlib/tests/heap.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.bit_set`
    - `able.collections.heap`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_17_stdlib_collections_bit_set_heap go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_17_stdlib_collections_bit_set_heap go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_17_stdlib_collections_bit_set_heap go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_17_stdlib_collections_bit_set_heap go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsBitSetAndHeapSuites' -count=1`

# 2026-02-12 — Compiler AOT collections deque/queue strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_16_stdlib_collections_deque_queue`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers deque/queue behavior through stdlib wrappers:
    - `Deque` push/pop from both ends, growth past initial capacity, and iterable traversal ordering.
    - `Queue` FIFO enqueue/dequeue/peek semantics, enumerable iteration, and clear behavior.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_16_stdlib_collections_deque_queue`.
- Added strict compiled CLI gate for deque/queue smoke suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsDequeAndQueueSmokeSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/collections/deque_smoke.test.able`
    - `v12/stdlib/tests/collections/queue_smoke.test.able`
  - asserts successful run and expected `able test: no tests to run` output for smoke modules.
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.deque`
    - `able.collections.queue`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_16_stdlib_collections_deque_queue go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_16_stdlib_collections_deque_queue go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_16_stdlib_collections_deque_queue go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_16_stdlib_collections_deque_queue go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsDequeAndQueueSmokeSuites' -count=1`

# 2026-02-12 — Compiler AOT collections hash_map/hash_set strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_15_stdlib_collections_hash_map_set`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers hash-backed collection behavior through stdlib wrappers:
    - `HashMap` set/get/remove/contains/for_each/map semantics.
    - `HashSet` add/remove/contains/union/intersect/difference/symmetric_difference/subset/superset/disjoint semantics.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_15_stdlib_collections_hash_map_set`.
- Added strict compiled CLI gate for hash collection suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsHashMapSmokeAndHashSetSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/collections/hash_map_smoke.test.able`
    - `v12/stdlib/tests/collections/hash_set.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.hash_map`
    - `able.collections.hash_set`
- Updated stdlib hash-map smoke test callback shape for strict compilation compatibility:
  - `v12/stdlib/tests/collections/hash_map_smoke.test.able`
  - replaced local named callback declaration in `check_for_each` with an inline lambda passed to `map.for_each`, preserving test semantics while avoiding compiler fallback on unsupported local function statements.
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_15_stdlib_collections_hash_map_set go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_15_stdlib_collections_hash_map_set go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_15_stdlib_collections_hash_map_set go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_15_stdlib_collections_hash_map_set go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsHashMapSmokeAndHashSetSuites' -count=1`

# 2026-02-12 — Compiler AOT collections linked_list/lazy_seq strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_14_stdlib_collections_linked_list_lazy_seq`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers linked and lazy collection behavior through stdlib implementations:
    - `LinkedList` push/pop on both ends, node-handle insert/remove, and deterministic traversal.
    - `LazySeq` cache-backed get/take/each/to_array behavior over array-seeded state.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_14_stdlib_collections_linked_list_lazy_seq`.
- Added strict compiled CLI gate for linked/lazy suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsLinkedListAndLazySeqSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/linked_list.test.able`
    - `v12/stdlib/tests/lazy_seq.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.linked_list`
    - `able.collections.lazy_seq`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/06_12_14_stdlib_collections_linked_list_lazy_seq$' -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_14_stdlib_collections_linked_list_lazy_seq go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_14_stdlib_collections_linked_list_lazy_seq go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_14_stdlib_collections_linked_list_lazy_seq go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsLinkedListAndLazySeqSuites' -count=1`

# 2026-02-12 — Compiler AOT collections persistent_sorted_set/persistent_queue strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_13_stdlib_collections_persistent_sorted_queue`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers persistent ordered/FIFO collection behavior through stdlib implementations:
    - `PersistentSortedSet` ordered uniqueness, first/last access, remove persistence, and range extraction.
    - `PersistentQueue` FIFO enqueue/dequeue/peek persistence plus deterministic iteration order.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_13_stdlib_collections_persistent_sorted_queue`.
- Added strict compiled CLI gate for persistent sorted/FIFO suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsPersistentSortedSetAndQueueSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/persistent_sorted_set.test.able`
    - `v12/stdlib/tests/persistent_queue.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.persistent_sorted_set`
    - `able.collections.persistent_queue`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_13_stdlib_collections_persistent_sorted_queue go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_13_stdlib_collections_persistent_sorted_queue go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_13_stdlib_collections_persistent_sorted_queue go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_13_stdlib_collections_persistent_sorted_queue go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsPersistentSortedSetAndQueueSuites' -count=1`

# 2026-02-12 — Compiler AOT collections persistent_map/persistent_set strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_12_stdlib_collections_persistent_map_set`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers persistent HAMT collection behavior through stdlib implementations:
    - `PersistentMap` insert/update/remove/get/contains semantics, collision handling, and builder-based construction.
    - `PersistentSet` structural-sharing insert/remove semantics plus union/intersect behavior.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_12_stdlib_collections_persistent_map_set`.
- Added strict compiled CLI gate for persistent collections suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsPersistentMapPersistentSetSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/persistent_map.test.able`
    - `v12/stdlib/tests/persistent_set.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.persistent_map`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_12_stdlib_collections_persistent_map_set go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_12_stdlib_collections_persistent_map_set go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_12_stdlib_collections_persistent_map_set go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_12_stdlib_collections_persistent_map_set go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsPersistentMapPersistentSetSuites' -count=1`

# 2026-02-12 — Compiler AOT collections tree_map/tree_set strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_11_stdlib_collections_tree_map_set`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers ordered collection behavior through stdlib tree collection impls:
    - `TreeMap` ordered insert/update/remove/get/contains plus `first`/`last` entry access.
    - `TreeSet` uniqueness-aware insertion plus ordered `first`/`last`, `contains`, and remove semantics.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_11_stdlib_collections_tree_map_set`.
- Added strict compiled CLI gate for ordered collections suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsTreeMapTreeSetSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/tree_map.test.able`
    - `v12/stdlib/tests/tree_set.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.tree_map`
    - `able.collections.tree_set`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_11_stdlib_collections_tree_map_set go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_11_stdlib_collections_tree_map_set go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_11_stdlib_collections_tree_map_set go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_11_stdlib_collections_tree_map_set go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsTreeMapTreeSetSuites' -count=1`

# 2026-02-12 — Compiler AOT collections list/vector strict gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_10_stdlib_collections_list_vector`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers persistent collection behavior through stdlib collection impls:
    - `List` accessors and structural transforms (`prepend`, `tail`, `last`, `nth`, `concat`, `reverse`)
    - `Vector` accessors and persistence operations (`push`, `set`, `pop`, `first/last/get`) with explicit old/new value assertions.
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_10_stdlib_collections_list_vector`.
- Added strict compiled CLI gate for stdlib collections suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibCollectionsListVectorSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/list.test.able`
    - `v12/stdlib/tests/vector.test.able`
- Extended build precompile discovery assertions:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now also checks:
    - `able.collections.list`
    - `able.collections.vector`
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_10_stdlib_collections_list_vector go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_10_stdlib_collections_list_vector go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_10_stdlib_collections_list_vector go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_10_stdlib_collections_list_vector go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibCollectionsListVectorSuites' -count=1`

# 2026-02-12 — Compiler AOT foundational stdlib compiled CLI gate (v12)
- Added strict compiled CLI coverage for foundational stdlib suites in `v12/interpreters/go/cmd/able/test_cli_test.go`:
  - new test: `TestTestCommandCompiledRunsStdlibFoundationalSuites`
  - runs `able test --compiled` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` against:
    - `v12/stdlib/tests/simple.test.able`
    - `v12/stdlib/tests/assertions.test.able`
    - `v12/stdlib/tests/enumerable.test.able`
  - asserts suite output markers are present and stderr is empty.
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRunsStdlibFoundationalSuites -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestTestCommandCompiledRunsStdlibFoundationalSuites|TestTestCommandCompiledRunsStdlibNumbersNumericSuite|TestTestCommandCompiledRunsStdlibExtendedNumericSuites|TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv|TestDiscoverPrecompilePackagesIncludesStdlibAndKernel' -count=1`

# 2026-02-12 — Compiler AOT precompile discovery assertions expanded for numeric + foundational stdlib sets (v12)
- Extended build precompile discovery assertions in `v12/interpreters/go/cmd/able/build_precompile_test.go`:
  - `TestDiscoverPrecompilePackagesIncludesStdlibAndKernel` now verifies:
    - `able.spec`
    - `able.collections.enumerable`
    - `able.test.protocol`
    - `able.test.harness`
    - `able.test.reporters`
    - `able.numbers.bigint`
    - `able.numbers.biguint`
    - `able.numbers.int128`
    - `able.numbers.uint128`
    - `able.numbers.rational`
    - `able.numbers.primitives`
  alongside existing `able.io`, `able.io.path`, and `able.kernel`.
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestTestCommandCompiledRunsStdlibFoundationalSuites|TestTestCommandCompiledRunsStdlibNumbersNumericSuite|TestTestCommandCompiledRunsStdlibExtendedNumericSuites|TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`

# 2026-02-12 — Compiler AOT numeric primitives strict gates (`numbers_numeric`) (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_09_stdlib_numbers_primitives`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers primitive numeric helpers from `able.numbers.primitives`:
    - i32 helpers (`abs`, `sign`, `div_mod`, `bit_count`, `bit_length`)
    - u32 bit helpers (`leading_zeros`, `trailing_zeros`)
    - f64 fractional helpers (`floor`, `ceil`, `round`, `fract`)
    - conversion/error paths (`to_u32`, `f64.to_i32`, reciprocal zero, invalid clamp bounds).
- Wired fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_09_stdlib_numbers_primitives`.
- Added compiled CLI stdlib gate for aggregate numeric suite:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibNumbersNumericSuite`
  - runs `able test --compiled v12/stdlib/tests/numbers_numeric.test.able` with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true`.
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_09_stdlib_numbers_primitives go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_09_stdlib_numbers_primitives go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_09_stdlib_numbers_primitives go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_09_stdlib_numbers_primitives go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestTestCommandCompiledRunsStdlibNumbersNumericSuite|TestTestCommandCompiledRunsStdlibExtendedNumericSuites|TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_04_stdlib_numbers_bigint,06_12_05_stdlib_numbers_biguint,06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational,06_12_09_stdlib_numbers_primitives go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_04_stdlib_numbers_bigint,06_12_05_stdlib_numbers_biguint,06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational,06_12_09_stdlib_numbers_primitives go test ./pkg/interpreter -run TestExecFixtures -count=1`

# 2026-02-12 — Compiler AOT extended numeric stdlib strict gates (int128/uint128/rational) (v12)
- Added new exec fixtures:
  - `v12/fixtures/exec/06_12_06_stdlib_numbers_int128`
  - `v12/fixtures/exec/06_12_07_stdlib_numbers_uint128`
  - `v12/fixtures/exec/06_12_08_stdlib_numbers_rational`
- Coverage:
  - `Int128`: arithmetic (`add/sub/mul/div/rem`), comparison, clamp, division-by-zero and conversion error paths.
  - `UInt128`: arithmetic (`add/sub/mul/div/rem`), comparison, clamp, bit helpers (`leading_zeros`, `trailing_zeros`), conversion/underflow/div-zero error paths.
  - `Rational`: normalization, arithmetic, comparison, clamp, floor/ceil/round, conversion/div-zero/clamp-order error paths.
- Wired all three fixtures into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` includes `exec/06_12_06_stdlib_numbers_int128`, `exec/06_12_07_stdlib_numbers_uint128`, `exec/06_12_08_stdlib_numbers_rational`.
- Added compiled CLI stdlib gate for extended numeric suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibExtendedNumericSuites`
  - runs `able test --compiled` against:
    - `v12/stdlib/tests/int128.test.able`
    - `v12/stdlib/tests/uint128.test.able`
    - `v12/stdlib/tests/rational.test.able`
  - enforces `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true`.
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestTestCommandCompiledRunsStdlibExtendedNumericSuites|TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_04_stdlib_numbers_bigint,06_12_05_stdlib_numbers_biguint,06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_04_stdlib_numbers_bigint,06_12_05_stdlib_numbers_biguint,06_12_06_stdlib_numbers_int128,06_12_07_stdlib_numbers_uint128,06_12_08_stdlib_numbers_rational go test ./pkg/interpreter -run TestExecFixtures -count=1`

# 2026-02-12 — Compiler AOT rescue mixed-result coercion + restored biguint error assertions (v12)
- Fixed compiler rescue lowering in `v12/interpreters/go/pkg/compiler/generator_rescue.go`:
  - rescue expression result typing now supports mixed monitored/clause result types in statement contexts by coercing branches to `runtime.Value` when required.
  - added explicit rescue-branch coercion helper `coerceRescueBranch`.
  - keeps strict `RequireNoFallbacks` compilation green for rescue flows that previously forced fallback via `rescue clause type mismatch`.
- Added compiler regression coverage in `v12/interpreters/go/pkg/compiler/compiler_test.go`:
  - `TestCompilerRescueStatementMixedResultTypesNoFallback`
  - asserts mixed-type rescue used as a statement compiles successfully with `RequireNoFallbacks: true` and emits zero fallbacks.
- Restored explicit BigUint error-path assertions in fixture `v12/fixtures/exec/06_12_05_stdlib_numbers_biguint`:
  - `from_i64` negative conversion rescue
  - `to_i64` overflow rescue
  - subtraction underflow rescue
  - updated `manifest.json` expected output and `v12/fixtures/exec/coverage-index.json` focus text accordingly.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRescueStatementMixedResultTypesNoFallback|TestCompilerRequireNoFallbacksFails' -count=1`
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_05_stdlib_numbers_biguint go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_rescue,11_03_bytecode_rescue_basic,11_03_rescue_ensure,11_03_rescue_rethrow_standard_errors,06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites -count=1`

# 2026-02-12 — Compiler AOT biguint stdlib fixture coverage under strict compiled gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_05_stdlib_numbers_biguint`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers `BigUint` arithmetic (`add/sub/mul`), comparison ordering, and clamp behavior with deterministic output assertions.
- Wired the biguint fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go` default fixture list now includes `06_12_05_stdlib_numbers_biguint`.
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go` default strict-dispatch fixture list now includes `06_12_05_stdlib_numbers_biguint`.
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go` default boundary-audit fixture list now includes `06_12_05_stdlib_numbers_biguint`.
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_05_stdlib_numbers_biguint`.
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_05_stdlib_numbers_biguint go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_05_stdlib_numbers_biguint go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`

# 2026-02-12 — Compiler AOT compiled stdlib bigint/biguint CLI gate (v12)
- Added strict compiled-mode CLI coverage for stdlib bigint/biguint suites:
  - `v12/interpreters/go/cmd/able/test_cli_test.go`
  - new test: `TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites`
  - runs `able test --compiled` against:
    - `v12/stdlib/tests/bigint.test.able`
    - `v12/stdlib/tests/biguint.test.able`
  - enforces `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` to keep this path as an AOT no-fallback gate.
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites|TestTestCommandCompiledRuns|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`

# 2026-02-12 — Compiler AOT bigint stdlib fixture coverage under strict compiled gates (v12)
- Added new exec fixture `v12/fixtures/exec/06_12_04_stdlib_numbers_bigint`:
  - files: `package.yml`, `manifest.json`, `main.able`
  - covers `BigInt` arithmetic (`add/sub/mul`), comparison ordering, clamp behavior, and conversion error paths (`to_u64`, `to_i64` overflow) with deterministic output assertions.
- Wired the bigint fixture into compiler strict/parity defaults:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go` default fixture list now includes `06_12_04_stdlib_numbers_bigint`.
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go` default strict-dispatch fixture list now includes `06_12_04_stdlib_numbers_bigint`.
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go` default boundary-audit fixture list now includes `06_12_04_stdlib_numbers_bigint`.
- Extended build precompile discovery assertion to include bigint package coverage:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go` now checks `able.numbers.bigint` is present in discovered precompile package sets.
- Updated fixture coverage index:
  - `v12/fixtures/exec/coverage-index.json` now includes `exec/06_12_04_stdlib_numbers_bigint`.
- Validation:
  - `cd v12/interpreters/go && ABLE_FIXTURE_FILTER=06_12_04_stdlib_numbers_bigint go test ./pkg/interpreter -run TestExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_12_04_stdlib_numbers_bigint go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_04_stdlib_numbers_bigint go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_04_stdlib_numbers_bigint go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestResolveBuildPrecompileStdlibFromEnvMissingDefaultsTrue' -count=1`

# 2026-02-12 — Compiler AOT build wiring for stdlib/kernel precompile + bundled sources (v12)
- `able build` now precompiles stdlib/kernel package graphs by default by discovering packages from stdlib search roots and passing them through loader `IncludePackages`:
  - added `v12/interpreters/go/cmd/able/build_precompile.go`
  - build toggle env: `ABLE_BUILD_PRECOMPILE_STDLIB=1|true|yes|on|0|false|no|off` (default: enabled)
  - build flags: `--precompile-stdlib` and `--no-precompile-stdlib`
- `able build` argument parsing and usage now include the stdlib precompile controls (`v12/interpreters/go/cmd/able/build.go`).
- External build outputs (outside module root) now bundle stdlib/kernel sources alongside copied interpreter/parser module trees:
  - `v12/interpreters/go/cmd/able/go_mod_root.go` now copies:
    - `v12/stdlib/src` -> `<out>/v12/stdlib/src`
    - `v12/kernel/src` -> `<out>/v12/kernel/src`
- Added coverage:
  - `v12/interpreters/go/cmd/able/build_precompile_test.go`
    - env parsing (default/explicit/invalid)
    - package discovery includes `able.io`, `able.io.path`, `able.kernel`
    - CLI arg override for `--no-precompile-stdlib`
  - updated `v12/interpreters/go/cmd/able/build_test.go` to assert bundled stdlib/kernel sources exist in external outputs.
- Validation:
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestResolveBuildPrecompileStdlibFromEnvExplicitValues|TestResolveBuildPrecompileStdlibFromEnvMissingDefaultsTrue|TestResolveBuildPrecompileStdlibFromEnvInvalid|TestDiscoverPrecompilePackagesIncludesStdlibAndKernel|TestParseBuildArgumentsPrecompileStdlibFlagOverridesEnv' -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestBuildTargetFromManifest|TestBuildOutputOutsideModuleRoot' -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run 'TestBuildNoFallbacksFlagFailsWhenFallbackRequired|TestBuildNoFallbacksEnvFailsWhenFallbackRequired|TestBuildNoFallbacksInvalidEnvFailsArgumentParsing|TestBuildAllowFallbacksOverridesEnv|TestTestCommandCompiledRuns|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRequireNoFallbacksFails|TestCompilerEmitsStructsAndWrappers' -count=1`

# 2026-02-12 — Compiler AOT strict no-fallback fixture/parity gates (v12)
- Added shared fixture-gate strictness helper in `v12/interpreters/go/pkg/compiler/compiler_fixture_strictness_test.go`:
  - fixture/parity compiler paths now default to `RequireNoFallbacks=true`;
  - optional local override via `ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=0|false|off|no`;
  - invalid override values fail fast with a clear test error.
- Applied strict compile options across fixture/parity harnesses:
  - `v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go` (`runCompilerExecFixture`)
  - `v12/interpreters/go/pkg/compiler/compiler_diagnostics_parity_test.go` (`runCompiledFixtureOutcome`; shared by diagnostics + concurrency parity)
  - `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go` (`runCompiledFixtureBoundaryOutcome`)
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go` (`runCompilerBoundaryAuditFixture`)
  - `v12/interpreters/go/pkg/compiler/compiler_strict_dispatch_test.go` (`runCompilerStrictDispatchFixture`)
  - `v12/interpreters/go/pkg/compiler/compiler_concurrency_parity_test.go` (`TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks`)
- Validation:
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_struct_positional go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_01_compiler_struct_positional go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_DYNAMIC_BOUNDARY_FIXTURES=13_04_import_alias_selective_dynimport go test ./pkg/compiler -run TestCompilerDynamicBoundaryParityFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_DIAGNOSTICS_FIXTURES=06_01_compiler_division_by_zero go test ./pkg/compiler -run TestCompilerDiagnosticsParityFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_CONCURRENCY_PARITY_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerConcurrencyParityFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_struct_positional ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_01_compiler_struct_positional ABLE_COMPILER_DYNAMIC_BOUNDARY_FIXTURES=13_04_import_alias_selective_dynimport ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=13_06_stdlib_package_resolution ABLE_COMPILER_DIAGNOSTICS_FIXTURES=06_01_compiler_division_by_zero ABLE_COMPILER_CONCURRENCY_PARITY_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundaryParityFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerDiagnosticsParityFixtures|TestCompilerConcurrencyParityFixtures|TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks' -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary native bound-method callback gates (v12)
- Added native-bound-method boundary coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_native_methods_test.go`:
  - `TestCompilerDynamicBoundaryNativeBoundMethodCallbackSuccessMarkers`
  - `TestCompilerDynamicBoundaryNativeBoundMethodCallbackFailureMarkers`
- These tests pass dynamic package native bound methods (e.g. `pkg.def`) through dynamic callback invocation and assert:
  - tree-walker vs compiled parity (success/failure),
  - explicit `call_value` marker presence,
  - fallback marker count remains zero.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(NativeBoundMethodCallbackSuccessMarkers|NativeBoundMethodCallbackFailureMarkers|BoundMethodCallbackSuccessMarkers|BoundMethodCallbackFailureMarkers|CallbackArrayConversionSuccessMarkers|CallbackArrayConversionFailureMarkers|CallbackHashMapConversionSuccessMarkers|CallbackHashMapConversionFailureMarkers|CallbackInterfaceConversionSuccessMarkers|CallbackInterfaceConversionFailureMarkers|CallbackUnionConversionSuccessMarkers|CallbackUnionConversionFailureMarkers|CallbackNullableConversionSuccessMarkers|CallbackNullableConversionFailureMarkers|CallbackNilStringConversionFailureMarkers|CallbackCharConversionFailureMarkers|CallbackStructConversionFailureMarkers|CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary bound-method callback gates (v12)
- Added method-value boundary coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_methods_test.go`:
  - `TestCompilerDynamicBoundaryBoundMethodCallbackSuccessMarkers`
  - `TestCompilerDynamicBoundaryBoundMethodCallbackFailureMarkers`
- These tests pass a bound method value (`counter.add`) through dynamic callback invocation and assert:
  - tree-walker vs compiled parity (success/failure),
  - explicit `call_value` marker presence,
  - fallback marker count remains zero.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(BoundMethodCallbackSuccessMarkers|BoundMethodCallbackFailureMarkers|CallbackArrayConversionSuccessMarkers|CallbackArrayConversionFailureMarkers|CallbackHashMapConversionSuccessMarkers|CallbackHashMapConversionFailureMarkers|CallbackInterfaceConversionSuccessMarkers|CallbackInterfaceConversionFailureMarkers|CallbackUnionConversionSuccessMarkers|CallbackUnionConversionFailureMarkers|CallbackNullableConversionSuccessMarkers|CallbackNullableConversionFailureMarkers|CallbackNilStringConversionFailureMarkers|CallbackCharConversionFailureMarkers|CallbackStructConversionFailureMarkers|CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary composite payload conversion gates (v12)
- Added container/composite boundary conversion coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_composites_test.go`:
  - `TestCompilerDynamicBoundaryCallbackArrayConversionSuccessMarkers`
  - `TestCompilerDynamicBoundaryCallbackArrayConversionFailureMarkers`
  - `TestCompilerDynamicBoundaryCallbackHashMapConversionSuccessMarkers`
  - `TestCompilerDynamicBoundaryCallbackHashMapConversionFailureMarkers`
- These tests exercise dynamic→compiled callback payloads with `Array i32` and `HashMap String i32` shapes and assert:
  - tree-walker vs compiled parity (success/failure),
  - explicit `call_value` boundary markers present,
  - fallback marker count remains zero.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryCallback(ArrayConversionSuccessMarkers|ArrayConversionFailureMarkers|HashMapConversionSuccessMarkers|HashMapConversionFailureMarkers|InterfaceConversionSuccessMarkers|InterfaceConversionFailureMarkers|UnionConversionSuccessMarkers|UnionConversionFailureMarkers|NullableConversionSuccessMarkers|NullableConversionFailureMarkers)|TestCompilerDynamicBoundary(CallbackNilStringConversionFailureMarkers|CallbackCharConversionFailureMarkers|CallbackStructConversionFailureMarkers|CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary interface/union/nullable conversion gates (v12)
- Added boundary conversion coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_type_shapes_test.go`:
  - `TestCompilerDynamicBoundaryCallbackInterfaceConversionSuccessMarkers`
  - `TestCompilerDynamicBoundaryCallbackInterfaceConversionFailureMarkers`
  - `TestCompilerDynamicBoundaryCallbackUnionConversionSuccessMarkers`
  - `TestCompilerDynamicBoundaryCallbackUnionConversionFailureMarkers`
  - `TestCompilerDynamicBoundaryCallbackNullableConversionSuccessMarkers`
  - `TestCompilerDynamicBoundaryCallbackNullableConversionFailureMarkers`
- All tests assert boundary marker behavior (`call_value` explicit markers present, fallback markers zero) plus tree-walker vs compiled parity for success/failure outcomes.
- Added local helper assertions/utilities in the same file:
  - `assertBoundaryCallValueMarkers`
  - `joinLines`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryCallback(InterfaceConversionSuccessMarkers|InterfaceConversionFailureMarkers|UnionConversionSuccessMarkers|UnionConversionFailureMarkers|NullableConversionSuccessMarkers|NullableConversionFailureMarkers)|TestCompilerDynamicBoundary(CallbackNilStringConversionFailureMarkers|CallbackCharConversionFailureMarkers|CallbackStructConversionFailureMarkers|CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary nil/char/struct conversion gates (v12)
- Added additional dynamic boundary conversion coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`:
  - `TestCompilerDynamicBoundaryCallbackNilStringConversionFailureMarkers`
    - dynamic callback passes `nil` to compiled `String` callback; asserts runtime failure parity + explicit `call_value` marker + zero fallback markers.
  - `TestCompilerDynamicBoundaryCallbackCharConversionFailureMarkers`
    - dynamic callback passes string literal to compiled `char` callback; asserts runtime failure parity + explicit `call_value` marker + zero fallback markers.
  - `TestCompilerDynamicBoundaryCallbackStructConversionFailureMarkers`
    - dynamic callback passes `nil` to compiled struct-typed callback; asserts runtime failure parity + explicit `call_value` marker + zero fallback markers.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(CallbackNilStringConversionFailureMarkers|CallbackCharConversionFailureMarkers|CallbackStructConversionFailureMarkers|CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary non-numeric conversion gates (v12)
- Added non-numeric dynamic boundary conversion coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`:
  - `TestCompilerDynamicBoundaryCallbackBoolConversionFailureMarkers`
    - dynamic callback passes integer to compiled `bool` callback; asserts runtime failure parity + explicit `call_value` marker + zero fallback markers.
  - `TestCompilerDynamicBoundaryCallbackStringConversionFailureMarkers`
    - dynamic callback passes `bool` to compiled `String` callback; asserts runtime failure parity + explicit `call_value` marker + zero fallback markers.
  - `TestCompilerDynamicBoundaryCallbackStringConversionSuccessMarkers`
    - dynamic callback passes string to compiled `String` callback; asserts successful parity (`able!`) + explicit `call_value` marker + zero fallback markers.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(CallbackBoolConversionFailureMarkers|CallbackStringConversionFailureMarkers|CallbackStringConversionSuccessMarkers|CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic boundary numeric conversion edge-case gates (v12)
- Expanded dynamic boundary conversion-failure coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`:
  - `TestCompilerDynamicBoundaryCallbackOverflowConversionFailureMarkers`
    - dynamic callback invokes compiled `i32` callback with out-of-range integer (`2147483648`) and asserts runtime failure parity plus explicit `call_value` marker emission.
  - `TestCompilerDynamicBoundaryCallbackUnsignedConversionFailureMarkers`
    - dynamic callback invokes compiled `u8` callback with negative integer (`-1`) and asserts runtime failure parity plus explicit `call_value` marker emission.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary(CallbackOverflowConversionFailureMarkers|CallbackUnsignedConversionFailureMarkers|CallbackConversionFailureMarkers|CallbackRoundtrip|CallNamedMarkers|CallOriginalMarkers|ParityFixtures)|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT call-family boundary marker coverage completion (v12)
- Extended dynamic boundary test coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go` with:
  - `TestCompilerDynamicBoundaryCallNamedMarkers`:
    - unresolved named call path (typecheck off) to exercise runtime `call_named` explicit marker emission.
  - `TestCompilerDynamicBoundaryCallOriginalMarkers`:
    - non-compileable function wrapper path to exercise runtime `call_original` explicit marker emission.
- Added helper utilities in the same test file:
  - `withTypecheckFixturesOff`
  - `hasBoundaryMarkerPrefix`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryCallNamedMarkers|TestCompilerDynamicBoundaryCallOriginalMarkers|TestCompilerDynamicBoundaryCallbackRoundtrip|TestCompilerDynamicBoundaryCallbackConversionFailureMarkers|TestCompilerDynamicBoundaryParityFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerEmitsStructsAndWrappers' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT callback conversion-failure boundary gate (v12)
- Added dynamic callback conversion-failure coverage in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`:
  - `TestCompilerDynamicBoundaryCallbackConversionFailureMarkers`
  - synthesizes a dynamic function that invokes a compiled callback with a bad argument type and asserts:
    - runtime failure occurs in both tree-walker and compiled runs,
    - zero fallback markers,
    - explicit boundary markers include `call_value`.
- Updated generated compiler test harness emission (`v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`) so boundary markers are printed for runtime-error exits after registration (not only successful exits), enabling boundary auditing for failing dynamic-boundary scenarios.
- Added compiler codegen assertions for boundary marker presence in generated output (`v12/interpreters/go/pkg/compiler/compiler_test.go`):
  - `call_original` wrapper marker emission
  - `call_named` bridge marker emission
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerEmitsStructsAndWrappers|TestCompilerDynamicBoundaryCallbackRoundtrip|TestCompilerDynamicBoundaryCallbackConversionFailureMarkers|TestCompilerDynamicBoundaryParityFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`

# 2026-02-12 — Compiler AOT dynamic callback roundtrip boundary coverage (v12)
- Added `TestCompilerDynamicBoundaryCallbackRoundtrip` in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`.
- The new test builds a synthetic dynamic program that:
  - defines a dyn package function `apply_twice` at runtime via `dyn.def_package(...).def(...)`,
  - passes a compiled callback (`fn(x: i32) -> i32`) into interpreted dynamic code,
  - validates compiled vs tree-walker output parity (`value 42`).
- Boundary marker assertions now cover callback roundtrip behavior by requiring explicit boundary markers (`call_value`) and zero fallback markers for the scenario.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryParityFixtures|TestCompilerDynamicBoundaryCallbackRoundtrip|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`

# 2026-02-12 — Compiler AOT explicit dynamic-boundary marker accounting (v12)
- Compiler generated runtime now tracks explicit compiled→interpreter bridge calls separately from fallback calls:
  - explicit counter/snapshot helpers in generated runtime call layer:
    - `__able_boundary_explicit_count_get()`
    - `__able_boundary_explicit_snapshot()`
  - explicit call-family markers:
    - `call_value`
    - `call_named`
    - `call_original`
  - fallback marker semantics remain focused on unexpected fallback routing.
- Harness marker output now includes explicit boundary markers when `ABLE_COMPILER_BOUNDARY_MARKER` is enabled:
  - `__ABLE_BOUNDARY_EXPLICIT_CALLS=...`
  - `__ABLE_BOUNDARY_EXPLICIT_NAMES=...` (verbose mode)
  (`v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`).
- Dynamic boundary parity gate now asserts:
  - tree-walker vs compiled parity for dynamic fixtures,
  - `fallback` marker count remains zero,
  - explicit boundary marker count is positive with non-empty names
  (`v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`).
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerDynamicBoundaryParityFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerDynamicBoundaryParityFixtures' -count=1`

# 2026-02-12 — Compiler AOT dynamic-boundary parity + bridge call fallback hardening (v12)
- Added dynamic boundary parity coverage for compiled mode with explicit dynamic fixtures (`06_10_dynamic_metaprogramming_package_object`, `13_04_import_alias_selective_dynimport`, `13_05_dynimport_interface_dispatch`, `13_07_search_path_env_override`) in `v12/interpreters/go/pkg/compiler/compiler_dynamic_boundary_parity_test.go`.
- The new gate compares tree-walker vs compiled outcomes (`stdout`, `stderr`, `exit`) and additionally asserts that these dynamic fixtures execute via explicit boundary paths without generic fallback-call marker hits.
- Compiler bridge call semantics now fall back to global environment lookup when current environment misses function symbols, aligning `Runtime.Call`/`CallNamedWithNode` behavior with existing `Get` fallback semantics (`v12/interpreters/go/pkg/compiler/bridge/bridge.go`).
- Added bridge regressions:
  - `TestRuntimeCallFallsBackToGlobalEnvironment`
  - `TestCallNamedFallsBackToGlobalEnvironment`
  in `v12/interpreters/go/pkg/compiler/bridge/bridge_test.go`.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerDynamicBoundaryParityFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundaryParityFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`

# 2026-02-11 — Compiled main dispatch path consolidation (v12)
- Compiler codegen now emits reusable entrypoint helpers in generated `compiled.go`:
  - `RunMain(interp)`
  - `RunMainIn(interp, env)`
  - `RunRegisteredMain(rt, interp, entryEnv)`
- `RunRegisteredMain` prefers compiled dispatch for `main` via the compiled call table (`__able_lookup_compiled_call`) and only falls back to interpreter callable invocation when no compiled entry is registered.
- Generated `main.go` now invokes `RunRegisteredMain` instead of directly branching between wrapper calls and `interp.CallFunction`.
- Updated compiled harness callers to use the same entrypoint helper:
  - compiler exec fixture harness (`v12/interpreters/go/pkg/compiler/exec_fixtures_compiler_test.go`)
  - `able test --compiled` harness source (`v12/interpreters/go/cmd/able/test_cli_compiled.go`)
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecHarness|TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks|TestCompilerConcurrencyParityFixtures|TestCompilerDiagnosticsParityFixtures' -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=15_01_program_entry_hello_world,12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`

# 2026-02-11 — Compiler performance baseline harness (v12)
- Added `BenchmarkCompilerExecFixtureBinary` for repeatable compiled-runtime execution baseline runs using exec fixtures (`v12/interpreters/go/pkg/compiler/compiler_performance_bench_test.go`).
- Benchmark flow:
  - resolve fixture (`ABLE_COMPILER_BENCH_FIXTURE`, default `v12/fixtures/exec/07_09_bytecode_iterator_yield`)
  - compile fixture once to generated Go
  - build one compiled benchmark binary
  - benchmark binary execution (`b.N` runs) with fixture env applied
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run '^$' -bench BenchmarkCompilerExecFixtureBinary -benchtime=1x -count=1`
  - sample result on this host:
    - `BenchmarkCompilerExecFixtureBinary-16  1  63310648 ns/op`
- Plan updates:
  - removed completed Compiler AOT TODO item for performance baseline from `PLAN.md`.

# 2026-02-11 — Compiler concurrency parity fixture gate expansion (v12)
- Added `TestCompilerConcurrencyParityFixtures` to compare tree-walker vs compiled outcomes (`stdout`, `stderr`, `exit`) for core concurrency/scheduler fixtures plus spawn/await compiler fixtures:
  - `06_01_compiler_spawn_await`
  - `06_01_compiler_await_future`
  - `12_01_bytecode_spawn_basic`
  - `12_01_bytecode_await_default`
  - `12_02_async_spawn_combo`
  - `12_02_future_fairness_cancellation`
  - `12_03_spawn_future_status_error`
  - `12_04_future_handle_value_view`
  - `12_05_concurrency_channel_ping_pong`
  - `12_05_mutex_lock_unlock`
  - `12_06_await_fairness_cancellation`
  - `12_07_channel_mutex_error_types`
  - `12_08_blocking_io_concurrency`
  - `15_04_background_work_flush`
- Added env override support via `ABLE_COMPILER_CONCURRENCY_PARITY_FIXTURES`.
- Expanded `TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks` to assert the same blocked-flush behavior in tree-walker goroutine mode for the synthetic nil-channel blocked-task program, keeping the compiler regression tied to reference runtime semantics.
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerConcurrencyParityFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks|TestCompilerConcurrencyParityFixtures' -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerConcurrencyParityFixtures|TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks|TestCompilerDiagnosticsParityFixtures' -count=1`
- Plan updates:
  - removed completed Compiler AOT TODO item for compiled concurrency semantics parity (including scheduler helpers) from `PLAN.md`.

# 2026-02-11 — Compiler goroutine `future_flush` blocked-task parity (v12)
- Compiler generated runtime: added blocked-task accounting to the goroutine future executor (`pending` + `blocked` + per-handle blocked state) and updated `Flush()` to short-circuit when all pending tasks are blocked, matching interpreter goroutine executor behavior (`v12/interpreters/go/pkg/compiler/generator_render_runtime_future.go`).
- Compiler generated concurrency helpers now mark async tasks blocked/unblocked around channel/mutex blocking waits and nil-channel waits, so goroutine executor accounting reflects real blocking states (`v12/interpreters/go/pkg/compiler/generator_render_runtime_concurrency.go`).
- Compiler generated nil-channel blocking now respects async context cancellation and reports an error outside async context, aligning with interpreter behavior (`v12/interpreters/go/pkg/compiler/generator_render_runtime_concurrency.go`).
- Added regression coverage to ensure compiled goroutine-mode `future_flush()` returns when pending tasks are blocked:
  - `v12/interpreters/go/pkg/compiler/compiler_concurrency_parity_test.go` (`TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks`)
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=12_05_concurrency_channel_ping_pong,12_05_mutex_lock_unlock,15_04_background_work_flush go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerDiagnosticsParityFixtures -count=1`
- Plan updates:
  - removed completed optional Compiler AOT follow-up for goroutine blocked-task accounting from `PLAN.md`.

# 2026-02-11 — Compiler diagnostics parity fixture gate (v12)
- Added `TestCompilerDiagnosticsParityFixtures` to compare tree-walker vs compiled runtime outcomes (stdout, stderr diagnostics, exit code) for arithmetic/runtime diagnostic fixtures (`v12/interpreters/go/pkg/compiler/compiler_diagnostics_parity_test.go`).
- The new diagnostics gate currently covers:
  - `04_02_primitives_truthiness_numeric_diag`
  - `06_01_compiler_division_by_zero`
  - `06_01_compiler_integer_overflow`
  - `06_01_compiler_integer_overflow_sub`
  - `06_01_compiler_integer_overflow_mul`
  - `06_01_compiler_unary_overflow`
  - `06_01_compiler_divmod_overflow`
  - `06_01_compiler_pow_overflow`
  - `06_01_compiler_pow_negative_exponent`
  - `06_01_compiler_shift_out_of_range`
  - `06_01_compiler_compound_assignment_overflow`
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerDiagnosticsParityFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
- Plan updates:
  - removed completed Compiler AOT diagnostics parity TODO item from `PLAN.md`.

# 2026-02-11 — Compiler AOT parity gates verified (v12)
- Verified compiler fixture parity and boundary behavior remain green across the current fixture set:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./v12/able test --compiled v12/stdlib/tests`
- Plan updates:
  - removed completed Compiler AOT TODO items for exec-fixture parity, compiled stdlib parity, and no-silent-fallback enforcement from `PLAN.md`.

# 2026-02-11 — Compiler AOT singleton static-overload dispatch parity (v12)
- Compiler runtime member lookup: fixed compiled `__able_member_get_method` fallback so singleton struct receivers can resolve compiled static overload wrappers (e.g. `AutomataDSL.union`) without redirecting all singleton instance lookups to static/type-ref mode (`v12/interpreters/go/pkg/compiler/generator_render_runtime_calls.go`).
- Compiler runtime interface dispatch: unwrap nested interface receivers before compiled interface method binding/selection to keep impl receiver typing stable (`v12/interpreters/go/pkg/compiler/generator_render_runtime_interface.go`).
- Tests added/updated:
  - `v12/interpreters/go/pkg/compiler/compiler_singleton_struct_test.go` (`TestCompilerSingletonStaticOverloadDispatch`)
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerZeroFieldStructIdentifierValue|TestCompilerSingletonStaticOverloadDispatch' -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`
  - `cd v12/interpreters/go && go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./v12/able test --compiled v12/stdlib/tests/assertions.test.able`
  - `./v12/able test --compiled v12/stdlib/tests/automata.test.able`
  - `./v12/able test --compiled v12/stdlib/tests/persistent_map.test.able`
  - `./v12/able test --compiled v12/stdlib/tests` (passes)

# 2026-02-11 — Compiler AOT stdlib-compiled-mode unblockers (v12)
- Compiler integer literal lowering: fixed `_i128`/`_u128` handling so runtime-value contexts now emit `runtime.IntegerValue` via `big.Int` parsing instead of narrow Go casts, preventing generated-code overflow failures during compiled stdlib builds (`v12/interpreters/go/pkg/compiler/generator_exprs.go`, `v12/interpreters/go/pkg/compiler/generator_types.go`).
- Compiler identifier lowering: zero-field struct identifiers now materialize direct struct instances in typed contexts (instead of loading struct definitions via global lookup), fixing singleton-style matcher constructors in compiled stdlib/spec paths (`v12/interpreters/go/pkg/compiler/generator_exprs_ident.go`).
- Compiler bridge: `StructDefinition` cache is now environment-scoped (`env pointer + name`) instead of bare-name scoped, avoiding cross-environment collisions for same-named structs (`v12/interpreters/go/pkg/compiler/bridge/bridge.go`).
- Tests added:
  - `v12/interpreters/go/pkg/compiler/compiler_integer_literals_test.go` (`TestCompilerBuildsLargeI128AndU128Literals`)
  - `v12/interpreters/go/pkg/compiler/compiler_singleton_struct_test.go` (`TestCompilerZeroFieldStructIdentifierValue`)
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_test.go` (`TestStructDefinitionCacheScopesByEnvironment`)
- Validation:
  - `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerBuildsLargeI128AndU128Literals -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerZeroFieldStructIdentifierValue -count=1`
  - `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `./run_all_tests.sh` (passes)
- Progress on compiled stdlib parity: `able test --compiled v12/stdlib/tests` now gets past previous literal/singleton constructor failures and advances to a later blocker (`v12/stdlib/src/collections/persistent_map.able:533:16 Ambiguous overload for insert`).

# 2026-02-10 — Compiler AOT boundary audit expansion (v12)
- Compiler runtime call dispatch: `__able_call_named` now attempts `env.Get(name)` and routes through `__able_call_value` before interpreter bridge fallback, eliminating avoidable named-call fallback when compiled call tables are not directly keyed for the current scope (`pkg/compiler/generator_render_runtime_calls.go`).
- Compiler runtime boundary lookup: compiled call lookup now walks environment parent chain (`__able_lookup_compiled_call`) to respect lexical scope nesting (`pkg/compiler/generator_render_runtime_calls.go`).
- Compiler runtime boundary diagnostics: boundary marker now supports an optional verbose breakdown (`ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1`) with per-name fallback counts (`__ABLE_BOUNDARY_FALLBACK_NAMES=...`) for targeted debugging (`pkg/compiler/generator_render_runtime_calls.go`, `pkg/compiler/exec_fixtures_compiler_test.go`).
- Boundary audit coverage: promoted previously failing fixtures into default zero-fallback audit set after fixes:
  - `12_08_blocking_io_concurrency`
  - `14_02_regex_core_match_streaming`
  (`pkg/compiler/compiler_boundary_audit_test.go`)
- Validation:
  - `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE=1 ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=12_08_blocking_io_concurrency,14_02_regex_core_match_streaming GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -v`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — Compiler AOT boundary fallback audit marker (v12)
- Compiler runtime call helpers now track fallback calls that route from compiled call sites into interpreter execution for names the compiler registered as compiled callables (`pkg/compiler/generator_render_runtime_calls.go`).
- Compiler fixture harness now supports `ABLE_COMPILER_BOUNDARY_MARKER=1` and emits `__ABLE_BOUNDARY_FALLBACK_CALLS=<count>` on stderr after execution (`pkg/compiler/exec_fixtures_compiler_test.go`).
- Added `TestCompilerBoundaryFallbackMarkerForStaticFixtures` with env override support (`ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all|a,b,c`) to assert zero unexpected compiled→interpreter fallback calls on a curated static fixture set (`pkg/compiler/compiler_boundary_audit_test.go`).
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -v`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — Compiler AOT dynamic boundary bridge hardening (v12)
- Bridge runtime: fixed goroutine env fallback semantics so `Runtime.Env()` no longer returns a sticky nil after `SwapEnv(nil)`; nil swaps now clear goroutine-local override and fall back to the registered base env (`pkg/compiler/bridge/bridge.go`).
- Bridge conversion: `AsString` now accepts interface-wrapped `Array` byte storage when decoding `String` struct values across compiled/interpreter boundaries (`pkg/compiler/bridge/bridge.go`).
- Tests: added bridge regressions for interface-wrapped String byte arrays and env fallback after nil env swap (`pkg/compiler/bridge/bridge_test.go`).
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler/bridge -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_10_dynamic_metaprogramming_package_object,13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -v`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — Compiler AOT strict dispatch hard-fail path (v12)
- Compiler: removed silent strict-dispatch downgrade in generated `RegisterIn`; compiled impl-thunk registration errors now fail immediately instead of flipping a hidden blocked flag (`pkg/compiler/generator_render_functions.go`).
- Compiler tests: strict-dispatch fixture audit now supports `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all|a,b,c` for broader gating while keeping a focused default set (`pkg/compiler/compiler_strict_dispatch_test.go`).
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -v`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency,13_06_stdlib_package_resolution,14_02_regex_core_match_streaming,14_01_language_interfaces_index_apply_iterable GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — Compiler AOT strict interface-dispatch registration parity (v12)
- Interpreter: fixed compiled impl-thunk registration for canonicalized impl targets by preserving source-form impl targets and matching registration against both source and canonical target expressions (`pkg/interpreter/impl_resolution.go`, `pkg/interpreter/definitions.go`, `pkg/interpreter/compiled_thunk.go`).
- Interpreter: compiled impl-thunk registration now accepts both raw and alias-expanded constraint signatures and substitutes interface bindings on both sides of param matching (`pkg/interpreter/compiled_thunk.go`).
- Compiler tests: added `TestCompilerStrictDispatchForStdlibHeavyFixtures` to assert `__able_interface_dispatch_strict == true` at runtime for stdlib-heavy compiled fixtures, and added a harness marker hook used by this audit (`pkg/compiler/compiler_strict_dispatch_test.go`, `pkg/compiler/exec_fixtures_compiler_test.go`).
- Plan: removed completed Compiler AOT TODO item for impl-thunk registration parity gaps (`PLAN.md`).
- Validation:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -v`
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency,13_06_stdlib_package_resolution,14_02_regex_core_match_streaming,14_01_language_interfaces_index_apply_iterable GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — Compiler AOT fallback closure + `!void` return parity (v12)
- Compiler AOT: removed remaining compiler fallback audit failures in stdlib-heavy and interface fixtures:
  - identifier lowering now supports typed-local coercion via runtime bridge conversion when an expected static type differs (`generator_exprs_ident.go`);
  - control-flow statement compilation now propagates nested failure reasons (`generator_controlflow.go`).
- Compiler AOT: added explicit `Result<void>` return handling for bare `return`:
  - compile-body return lowering now treats bare returns in `-> !void` functions as `runtime.VoidValue{}` (not missing-return fallback / nil value);
  - statement-mode `return` in `Result<void>` contexts now emits `__able_return{value: runtime.VoidValue{}}` (`generator.go`, `generator_types.go` helper).
- Validation:
  - `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`
  - `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_03_operator_overloading_interfaces,14_01_language_interfaces_index_apply_iterable GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -v`
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./cmd/able -run TestTestCommandCompiledRuns -count=1`
  - `./run_all_tests.sh` (passes)

# 2026-02-10 — v12 test blockers cleared (coverage index + compiler/runtime parity)
- Fixtures: added missing seeded entries to `v12/fixtures/exec/coverage-index.json` for new compiler/iterator/interface fixtures so the exec coverage guard passes.
- Compiler codegen:
  - fixed generated runtime errors to avoid `fmt.Errorf(message)` vet failures (`generator_render_runtime.go`);
  - fixed impl-method wrapper receiver writeback so compiled iterator state mutations persist (`generator_render_functions.go`);
  - fixed generic local-lambda calls with type arguments to call local values instead of unresolved global names (`generator_exprs.go`);
  - added call-frame push/pop in dynamic value calls to preserve caller notes in runtime diagnostics (`generator_render_runtime_calls.go`);
  - fixed match-binding temp declarations to avoid unused-temp compile failures without changing match semantics (`generator_match.go`).
- Stdlib: updated `Array` `Index.get` to return `IndexError` for out-of-bounds access (`v12/stdlib/src/collections/array.able`) so `arr[idx]!` rescue/rethrow fixtures behave per spec.
- Tests:
  - `ABLE_COMPILER_EXEC_FIXTURES=07_02_01_verbose_anonymous_fn ... go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `ABLE_COMPILER_EXEC_FIXTURES=07_10_iterator_reentrancy ... go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`
  - `go test ./pkg/interpreter -run 'TestExecFixtures/11_03_rescue_rethrow_standard_errors' -count=1 -exec-mode=treewalker`
  - `go test ./pkg/interpreter -run 'TestExecFixtures/11_03_rescue_rethrow_standard_errors' -count=1 -exec-mode=bytecode`
  - `./run_all_tests.sh` (passes)

# 2026-02-06 — Compiler match-statement lowering + stdlib explicit casts (v12)
- Compiler: treat match expressions used as statements as void blocks so clause bodies can be statement-only (fixes regex parse_tokens compilation).
- Stdlib: `to_u64` helpers now use explicit `u64` casts/literals to avoid implicit coercion.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_FALLBACK_AUDIT=1 go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1`.

# 2026-02-06 — Compiler struct arg writeback for runtime.Value callers (v12)
- Compiler: when passing runtime.Value struct bindings to compiled functions, convert once and apply mutations back to the runtime struct instance (fixes assignment evaluation order fixture).
- Tests:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_03_assignment_evaluation_order go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
  - Compiler exec fixtures sweep in 11 batches via `ABLE_COMPILER_EXEC_FIXTURES` (207 fixtures total).

# 2026-02-06 — Stdlib test run (v12)
- Tests: `./run_stdlib_tests.sh`.

# 2026-02-06 — Full test run (v12)
- Tests: `./run_all_tests.sh`.

# 2026-02-06 — Parser cast line breaks (v12)
- Parser: allow line breaks after `as` in cast expressions; restored cast fixture to newline form.
- Tests:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser -count=1`
  - `./v12/abletw v12/fixtures/exec/06_03_cast_semantics/main.able`
  - `./v12/ablebc v12/fixtures/exec/06_03_cast_semantics/main.able`

# 2026-02-06 — Full test run (v12, post cast fix)
- Tests: `./run_all_tests.sh`.

# 2026-02-06 — Ablec build integration test (v12)
- CLI: `ablec` now has a testable `run` entrypoint and a build test covering go.mod + binary output.
- Tests: `cd v12/interpreters/go && go test ./cmd/ablec -count=1`.

# 2026-02-06 — Compiler multi-package build + native binary output (v12)
- Compiler: collect/compile functions across packages, qualify overload helpers by package, and register compiled function thunks per package environment.
- Compiler: add struct apply helpers + per-package env swaps so compiled methods update runtime struct instances and execute under the right package env.
- Runtime bridge: track per-goroutine env in compiled bridge (`SwapEnv`/`Env`) to support async execution.
- Interpreter: track package environments, expose compiled function overload registration, and support array member assignment + interface matching by struct fields for compiled values.
- CLI: `able build` command + `ablec -build` now emit `go.mod` in build output and run `go build -mod=mod` for native binaries; `--with-tests` loads test modules for run/check/build; compiled test runner avoids importing package names directly.
- Tests:
  - `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (takes ~211s; exceeds 1-minute guideline)
  - `cd v12/interpreters/go && go test ./cmd/able -count=1`
  - `cd v12/interpreters/go && go test ./cmd/ablec -count=1`
  - `./run_stdlib_tests.sh`
  - `./run_all_tests.sh`

# 2026-02-04 — Compiler untyped param support (v12)
- Compiler: map missing type annotations to `runtime.Value`, removing param/return-type fallbacks for untyped parameters.
- Fallback audit (exec fixtures) after update:
  - Top reasons: unsupported function body (14), unknown struct literal type (10), unsupported struct literal (10).
  - Top functions: `main` still dominated by struct literal typing and unsupported bodies.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_07_types_alias_union_generic_combo,06_04_function_call_eval_order_trailing_lambda,06_07_generator_yield_iterator_end go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-02-04 — Compiler typed pattern lowering (v12)
- Compiler: typed patterns now lower via runtime casts for match/assignment/loop bindings; added `__able_try_cast` helper and global lookup bridge support.
- Fallback audit (exec fixtures) after typed-pattern changes:
  - Top reasons: unsupported param/return type (20), unsupported function body (12), unknown struct literal type (10), unsupported struct literal (9).
  - Top functions: `main` still dominates (struct literal typing + unsupported body), then `status_name`, `maybe_text`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_02_identifier_wildcard_typed_patterns,05_02_struct_pattern_rename_typed,06_01_compiler_assignment_pattern_typed_mismatch,06_01_compiler_match_patterns,06_01_compiler_for_loop_typed_pattern,06_01_compiler_for_loop_typed_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-02-04 — Compiler type mapping + global identifier lookup (v12)
- Compiler: broadened type mapping to treat unknown/simple/generic/function/wildcard types as `runtime.Value`.
- Compiler: unknown identifiers now lower to runtime global lookup with diagnostic context (bridge `Get` + compiled helper).
- Fallback audit (exec fixtures) after updates:
  - Top reasons: unsupported typed pattern (21), unsupported function body (20), unsupported param/return type (20), unknown/unsupported struct literal (9 each).
  - Top functions: `main` dominates (typed patterns, unsupported body, struct literal typing), then `status_name`, `maybe_text`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_01_type_inference_constraints,06_01_compiler_method_call,10_03_interface_type_dynamic_dispatch,13_01_package_structure_modules go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-02-04 — Compiler fallback audit (v12)
- Added `compiler.Result.Fallbacks` to track interpreter fallbacks (including overloads).
- Audit summary (exec fixtures):
  - Top reasons: unsupported param/return type (62), unknown identifier (25), unsupported function body (21), unsupported typed pattern (16), unknown struct literal type (9).
  - Top functions (by occurrences): `main` dominated (unknown identifier/unsupported body/typed patterns), plus `status_name`, `tick`, `describe`, `maybe_text`.
- Notes: prioritize param/return type support + typed pattern lowering; then unknown identifier + struct literal typing gaps.

# 2026-02-04 — Interpreter test run (v12)
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -count=1 -timeout 60s`.

# 2026-02-04 — Compiler exec bytecode fixtures (v12)
- Fixtures: added remaining bytecode exec fixtures to the compiler exec list for compiled parity coverage.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_03_bytecode_assignment_patterns,06_01_bytecode_map_spread,06_02_bytecode_unary_range_cast,07_02_bytecode_lambda_calls,07_07_bytecode_implicit_iterator,07_08_bytecode_placeholder_lambda,07_09_bytecode_iterator_yield,08_01_bytecode_if_indexing,08_01_bytecode_match_basic,08_01_bytecode_match_subject,08_02_bytecode_loop_basics,09_00_bytecode_member_calls,11_02_bytecode_or_else_basic,11_03_bytecode_ensure_basic,11_03_bytecode_rescue_basic go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-02-04 — Compiler exec diagnostics + pipe placeholder fix (v12)
- Compiler: pipe lowering now emits multiline blocks so switch/case stays valid; placeholder lambdas are emitted as `runtime.Value` for type switches.
- Compiler: return statements missing values in non-void functions now raise runtime diagnostics with source context; added return type mismatch helper.
- Compiler exec harness: expand expected stdout/stderr fixtures with embedded newlines.
- Fixtures: added remaining non-bytecode lexer/typecheck/diagnostic exec fixtures to the compiler exec list.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=02_lexical_comments_identifiers,03_blocks_expr_separation,04_01_type_inference_constraints,04_02_primitives_truthiness_numeric_diag,04_03_type_expression_syntax,04_03_type_expression_arity_diag,04_03_type_expression_associativity_diag,04_04_reserved_underscore_types,04_05_02_struct_named_update_mutation_diag,04_06_04_union_guarded_match_exhaustive_diag,06_01_literals_numeric_contextual_diag,06_09_lexical_trailing_commas_line_join,11_01_return_statement_type_enforcement,11_01_return_statement_typecheck_diag,13_02_packages_visibility_diag go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Full test run (v12)
- Tests: `./run_all_tests.sh`.

# 2026-01-31 — Compiler bound method value fixture (v12)
- Compiler: allowed struct member access to fall back to runtime so bound method values can be captured.
- Fixtures: added compiler exec fixture for bound method values; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Full test run (v12)
- Tests: `./run_all_tests.sh`.

# 2026-01-31 — Compiler dynamic member access fixture (v12)
- Compiler: allowed runtime member access expressions to lower via member-get bridge.
- Fixtures: added compiler exec fixture for dynamic member access; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Compiler type-qualified methods fixture (v12)
- Fixtures: added compiler exec fixture for type-qualified methods; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Compiler interpolation Display fixture (v12)
- Fixtures: added compiler exec fixture for struct to_string interpolation; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Full test run (v12)
- Tests: `./run_all_tests.sh`.

# 2026-01-31 — Compiler method calls + block scoping (v12)
- Compiler: lowered method call syntax via runtime dispatch; block expressions now compile into scoped closures to allow shadowing.
- Runtime bridge: added call-by-value and method-preferred member access helpers for compiled code.
- Fixtures: added compiler exec fixture for method call syntax; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Stdlib test run (v12)
- Tests: `./run_stdlib_tests.sh`.

# 2026-01-31 — Full test run (v12)
- Tests: `./run_all_tests.sh`.

# 2026-01-31 — Compiler string interpolation lowering (v12)
- Compiler: added string interpolation lowering using runtime stringify for Display conversions.
- Runtime bridge: exposed interpreter stringify for compiled code.
- Fixtures: added compiler exec fixture for string interpolation; updated exec coverage index.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-31 — Compiler exec harness typecheck parity (v12)
- Compiler: aligned compiled exec harness with fixture typecheck mode (allow diagnostics unless fixtures are typecheck-off), preventing silent skips when warnings exist.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_dynamic_member_compound go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler dynamic member compound fixture (v12)
- Fixtures: added compiler exec fixture for dynamic compound member assignment; updated exec coverage + compiler fixture list.

# 2026-01-30 — Compiler dynamic compound member assignment (v12)
- Compiler: added dynamic member get bridge and compound member assignment lowering for runtime values.
- Interpreter: exposed member-get wrapper for compiled interop.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Exec fixture stderr normalization (v12)
- Fixtures: normalized exec fixture stderr comparisons to split multi-line diagnostics; updated compiler error fixture manifests.
- Tests: `./run_all_tests.sh`.

# 2026-01-30 — Compiler compound assignment lowering (v12)
- Compiler: added compound assignment lowering (`+=`, `-=`, `*=`, `/=`, `%=`, `.&=`, `.|=`, `.^=`, `.<<=`, `.>>=`) with RHS-first evaluation for identifiers, index targets, and struct fields.
- Compiler: added runtime binary-op helper for dynamic compound assignments.
- Fixtures: added compiler exec fixture for compound assignments; updated exec coverage + compiler fixture list.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler /% divmod lowering (v12)
- Compiler: added /% lowering via runtime binary operator bridge for DivMod results.
- Compiler: map DivMod generic type to runtime values for compiled function signatures.
- Fixtures: added compiler exec fixture for DivMod results; updated exec coverage + compiler fixture list.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler bitwise/shift lowering (v12)
- Compiler: added bitwise and shift operator lowering with overflow/shift-range checks for compiled code.
- Runtime bridge: exposed standard overflow and shift-out-of-range error values for compiled helpers.
- Fixtures: added compiler exec fixtures for bitwise ops and shift out-of-range diagnostics; updated exec coverage + compiler fixture list.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler division ops lowering (v12)
- Compiler: added /, //, % lowering with division-by-zero raises and Euclidean integer helpers for compiled code.
- Runtime bridge: exposed DivisionByZeroError value for compiled helpers.
- Fixtures: added compiler exec fixtures for division ops and division-by-zero behavior; updated exec coverage and compiler fixture list.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler map literal spread lowering (v12)
- Compiler: added map literal spread lowering via HashMap for-each callbacks in compiled code.
- Compiler: infer HashMap type arguments from map literal entries/spreads; refactored generator helpers to keep files under 1000 lines.
- Fixtures: added compiler map literal spread exec fixture and updated exec coverage + compiler fixture list.

# 2026-01-30 — Compiler typed HashMap literal fixture (v12)
- Fixtures: added a typed HashMap compiler exec fixture to exercise map literal inference; updated exec coverage and compiler fixture list.

# 2026-01-30 — Compiler map literal lowering (v12)
- Compiler: added map literal lowering to runtime HashMap creation with explicit entry sets (no spread yet).
- Fixtures: added compiler exec fixture for map literals; updated exec coverage index and compiler fixture list.

# 2026-01-30 — WASM JS host ABI draft (v12)
- Docs: defined the initial JS host ABI for the WASM runtime (stdout/stderr, timers, filesystem, module search roots) in `v12/docs/wasm-host-abi.md`.

# 2026-01-30 — Exec coverage + full test run (v12)
- Fixtures: added compiler fixture entries to exec coverage index; adjusted index-assignment fixture manifest to omit empty stdout expectation.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`.

# 2026-01-30 — Compiler member assignment lowering (v12)
- Compiler: lowered struct field assignment to Go field writes with RHS-first evaluation; added runtime member assignment fallback for dynamic values.
- Runtime bridge: added member assignment bridge helper and interpreter wrapper.
- Fixtures: added compiler exec fixture for member assignment.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.

# 2026-01-30 — Compiler unary/comparison/control-flow codegen (v12)
- Compiler: added unary `-`, `!` (bool-only), and bitwise not `.~` codegen plus comparison operators for primitive types.
- Compiler: added bool-only `&&`/`||` and if/elsif/else codegen for boolean conditions with same-typed branches; block expressions now compile in tail positions.
- Compiler: allow untyped integer literals to adopt float contexts during codegen.
- Compiler: fixed `:=` handling to allow shadowing outer bindings while rejecting same-scope redeclarations.
- Compiler: split render/control-flow/type helpers into `generator_render.go`, `generator_controlflow.go`, and `generator_types.go` to keep `generator.go` under 1000 lines.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler`.

# 2026-01-30 — Compiler exec fixture parity runner (v12)
- Compiler: added exec fixture parity runner that builds and runs compiled wrappers against a configurable fixture subset (`ABLE_COMPILER_EXEC_FIXTURES`, defaulting to a small smoke list).
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler`.

# 2026-01-30 — IR track deferred (v12)
- Plan: removed the typed core IR + runtime ABI implementation track from `PLAN.md` (deferred in favor of direct Go codegen).

# 2026-01-30 — Bytecode VM expansion track completed (v12)
- Plan: removed the completed interpreter performance/bytecode VM expansion track from `PLAN.md`.

# 2026-01-30 — Error-payload cast typechecker + full test runs (v12)
- Typechecker: allow explicit `as` casts from `Error` values to struct targets (payload recovery) with runtime checks.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`.

# 2026-01-30 — Top-level else/elsif parsing fix (v12)
- Parser: attach `else`/`elsif` clause statements to the preceding `if` at module scope, matching block parsing and v12 semantics.
- Fixtures: re-exported v12 AST fixtures via `./v12/export_fixtures.sh` (Go exporter, full run).
- Tests: `cd v12/interpreters/go && go test ./pkg/parser -count=1`.

# 2026-01-30 — Bytecode doc + singleton payload cast AST fixture (v12)
- Docs: expanded bytecode VM format + calling convention details in `v12/design/compiler-interpreter-vision.md`.
- Fixtures: added AST fixture `errors/error_payload_cast_singleton` and exported its `module.json` via the Go fixture exporter.
- Tests: not run (fixture export only).

# 2026-01-30 — Error payload cast recovery fixture (v12)
- Fixtures: added exec coverage for error-payload cast recovery via `as`.
- Tests: `./run_all_tests.sh`.

# 2026-01-30 — Stdlib test run (v12)
- Tests: `./run_stdlib_tests.sh`.

# 2026-01-30 — Type-application newline fix (v12)
- Parser: added external type-application separator to prevent newline from binding space-delimited type applications, plus immediate parenthesized type application for `fn()` type forms.
- Scanner: emit type-application separators only for same-line type prefixes and avoid reserved keywords; keep newline continuation logic intact.
- Fixtures: removed semicolon workaround in AST error payload fixtures.
- Tests: `npx tree-sitter test`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`; `./run_all_tests.sh`.

# 2026-01-29 — Parser line-break operators + type-application newline guard (v12)
- Parser: treat newlines as statement separators and add line-break-aware operator tokens so line-leading operators parse without consuming trailing newlines.
- Parser: remove optional line breaks before assignment operators; keep line-break handling after operators and inside delimiters.
- Parser: regenerated tree-sitter artifacts.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`; `cd v12/interpreters/go && go test ./pkg/interpreter`.

# 2026-01-28 — Bytecode iterator literal pre-lowering (v12)
- Bytecode: pre-lower iterator literal bodies to bytecode when supported, falling back to tree-walker execution for unsupported nodes.
- Design: documented `iterator_literal` in the bytecode instruction set.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_IteratorLiteral -count=1`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_07_bytecode_implicit_iterator -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_09_bytecode_iterator_yield -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/08_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/11_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/12_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/13_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/14_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/05_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/09_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/10_ -count=1 -timeout 60s`.

# 2026-01-28 — Stdlib BigInt/BigUint (v12)
- Stdlib: added `able.numbers.bigint` and `able.numbers.biguint` with basic arithmetic, comparisons, formatting, and numeric conversions.
- Tests: added BigInt/BigUint stdlib tests under `v12/stdlib/tests`.

# 2026-01-28 — Bytecode ensure inline handler (v12)
- Bytecode: execute ensure blocks inline after evaluating the try expression via fallback, then rethrow any captured error or return the try result.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_EnsureExpression -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_03_bytecode_ensure_basic -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode rescue inline handler (v12)
- Bytecode: execute rescue clauses inline after evaluating the monitored expression via fallback, matching patterns/guards before returning or rethrowing.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_RescueExpression -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_03_bytecode_rescue_basic -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_03 -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode await iterable lowering (v12)
- Bytecode: lower await iterable expressions to bytecode before running the await protocol.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/12_01_bytecode_await_default -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/12_06 -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/12_ -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode breakpoint labeled break (v12)
- Bytecode: lower labeled break statements to a breakpoint-aware opcode for non-local exits.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_BreakpointExpression -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/08_03_breakpoint_nonlocal_jump -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/08_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_ -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode match subject lowering (v12)
- Bytecode: lower match subjects as bytecode expressions before clause dispatch, leaving guards/bodies on fallback eval.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_Match -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/08_01_bytecode_match_subject -count=1 -timeout 60s`.

# 2026-01-28 — Exec fixture parity (13_04 slice)
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/13_04 -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode or-else exec fixture (v12)
- Fixtures: added bytecode-friendly `or {}` exec fixture for nil fallback and error binding.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_02_bytecode_or_else_basic -count=1 -timeout 60s`.

# 2026-01-28 — Exec fixture parity (11_03 slice)
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_03 -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode or-else handler path (v12)
- Bytecode: route `or {}` handling through a dedicated opcode that catches raised errors, binds failures in a fresh scope, and evaluates the handler inline.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_OrElseExpression -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_02 -count=1 -timeout 60s`.

# 2026-01-28 — Bytecode import ops (v12)
- Bytecode: added native import/dynimport opcodes and moved spawn execution into the controlflow helper to keep the VM file under 1000 lines.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_ImportStatement -count=1`.

# 2026-01-28 — Bytecode pattern compound assignment guard (v12)
- Bytecode: lower compound pattern assignments to the pattern assignment opcode so the VM raises the expected runtime error.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_CompoundAssignmentPattern -count=1`.

# 2026-01-27 — Bytecode assignment fixtures (v12)
- Fixtures: added exec coverage for bytecode-friendly pattern assignments and identifier compound assignments.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/05_03_bytecode_assignment_patterns -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/05_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/06_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/07_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/08_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/09_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/10_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/11_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/12_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/13_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/14_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/15_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/16_ -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode definition ops (v12)
- Bytecode: added definition opcodes for unions, type aliases, methods, interfaces, implementations, and externs (with runtime context attached on errors).
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_DefinitionOpcodes -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/09_ -count=1 -timeout 60s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestExecFixtureParity/10_ -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode member/index diagnostics (v12)
- Bytecode: attach runtime context and standard error wrapping to member/index get/set errors for parity with tree-walker diagnostics.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_MemberAccess -count=1`.

# 2026-01-27 — Bytecode opcode docs (v12)
- Docs: documented assignment and member-set bytecode opcodes in the compiler/interpreter vision doc.

# 2026-01-27 — Bytecode name diagnostics (v12)
- Bytecode: attach runtime context for identifier loads and `:=` redeclaration errors by threading source nodes into load/declare opcodes.

# 2026-01-27 — Bytecode loop pattern diagnostics (v12)
- Bytecode: attach runtime context to loop pattern binding errors in bytecode for parity with tree-walker diagnostics.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_ForLoopArraySum -count=1`.

# 2026-01-27 — Bytecode delegated ops audit (v12)
- Plan: documented remaining delegated ops for future bytecode lowering.

# 2026-01-27 — Bytecode compound assignments (v12)
- Bytecode: lower identifier compound assignments (e.g., `+=`) to a native opcode that evaluates RHS first and reuses the current binding for correct semantics.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_CompoundAssignmentName -count=1`.

# 2026-01-27 — Bytecode pattern assignments (v12)
- Bytecode: lower non-identifier pattern assignments to a native opcode and execute via `assignPattern`, including typed patterns and `:=` new-binding checks.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_AssignmentPattern -count=1`.

# 2026-01-27 — Bytecode return statements (v12)
- Bytecode: lower return statements to a native opcode that emits return signals for function returns while preserving “return outside function” errors at module scope.
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_Return -count=1`.

# 2026-01-27 — Bytecode member assignment (v12)
- Bytecode: lower member/implicit-member assignments to new opcodes and implement VM handling for struct/array member mutations (kept member/index ops in a helper file to stay under the 1000-line limit).
- Tests: `cd v12/interpreters/go && go test ./pkg/interpreter -run TestBytecodeVM_MemberAssignment -count=1`.

# 2026-01-27 — Exec perf harness (v12)
- Tooling: added Go benchmarks to compare tree-walker vs bytecode execution over exec fixtures (configurable via `ABLE_BENCH_FIXTURE`).
- Benchmarks: `cd v12/interpreters/go && go test -bench ExecFixture ./pkg/interpreter -run '^$'`.

# 2026-01-27 — Bytecode format documentation (v12)
- Docs: documented the current bytecode VM instruction set and calling convention in `v12/design/compiler-interpreter-vision.md`.

# 2026-01-27 — Bytecode async resume + typed pattern assignment (v12)
- Bytecode: preserve VM state across `future_yield` in async tasks (resume VM on yield), and advance past yield calls so tasks don't restart; also route typed-pattern assignments through the tree-walker path to preserve type-driven coercions.
- Bytecode: wrap standard runtime errors (division by zero, etc.) and attach runtime context for raise/rethrow to match rescue behavior/diagnostics.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/10_11_interface_generic_args_dispatch -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/11_03_raise_exit_unhandled -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/11_03_rescue_rethrow_standard_errors -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/12_02_async_spawn_combo -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/12_02_future_fairness_cancellation -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode diagnostics parity (v12)
- Bytecode: attach runtime context to match/range/cast errors so fixture diagnostics include source locations.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/04_06_04_union_guarded_match_exhaustive_diag -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_02_bytecode_unary_range_cast -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestExecFixtureParity/06_' -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode loop signals + call diagnostics (v12)
- Bytecode: added loop-enter/exit tracking so delegated eval can honor break/continue, and attached runtime context to call errors for parity (moved call ops into helper file).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_07_iterator_pipeline -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_07_overload_resolution_runtime -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_ -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode placeholder block lowering (v12)
- Bytecode: lower named function bodies as blocks to avoid mistakenly treating blocks with placeholder lambdas as placeholder closures.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_08_bytecode_placeholder_lambda -count=1 -timeout 60s`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_Placeholder -count=1 -timeout 60s`.

# 2026-01-27 — Bytecode function bodies (v12)
- Bytecode: function and lambda bodies now execute via compiled bytecode when running in bytecode mode (with tree-walker fallback if lowering fails).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestBytecodeVM_(LambdaCalls|SpawnExpression|IteratorLiteral|ForLoopArraySum)$' -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_02_bytecode_lambda_calls -count=1`.

# 2026-01-27 — Bytecode iterator yield fixture (v12)
- Fixtures: added exec coverage for iterator literals that yield with loop control in bytecode mode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_09_bytecode_iterator_yield -count=1`.

# 2026-01-27 — Bytecode yield opcode (v12)
- Bytecode: yield statements now lower to a native opcode, letting iterator bodies run fully in bytecode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_IteratorLiteral -count=1`.

# 2026-01-27 — Bytecode for-loop lowering (v12)
- Bytecode: for loops now lower to native bytecode with iterator opcodes and pattern binding (no tree-walker delegation for the loop itself).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestBytecodeVM_ForLoop(ArraySum|BreakValue)$' -count=1`.

# 2026-01-27 — Bytecode await evaluation (v12)
- Bytecode: await opcode now evaluates the await-expression iterable via bytecode when possible (fallback per expression).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_AwaitExpressionManualWaker -count=1`.

# 2026-01-27 — Bytecode iterator + breakpoint evaluation (v12)
- Bytecode: iterator literal and breakpoint opcodes now execute their bodies via bytecode when lowering succeeds (fallback to tree-walker per-expression).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestBytecodeVM_(IteratorLiteral|BreakpointExpression)$' -count=1`.

# 2026-01-27 — Bytecode rescue/or-else/ensure evaluation (v12)
- Bytecode: rescue/or-else/ensure opcodes now evaluate inner expressions via bytecode when possible (fallback per expression).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run 'TestBytecodeVM_(MatchLiteralPatterns|MatchGuard|RescueExpression|EnsureExpression|OrElseExpression)$' -count=1`.

# 2026-01-27 — Bytecode match evaluation (v12)
- Bytecode: match opcode now evaluates subject, guards, and bodies via bytecode when possible (with tree-walker fallback per expression).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_MatchLiteralPatterns -count=1`.

# 2026-01-27 — Bytecode implicit member direct access (v12)
- Bytecode: implicit member opcode now resolves the implicit receiver directly in the VM without tree-walker delegation.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_ImplicitMemberExpression -count=1`.

# 2026-01-27 — Bytecode assignment pattern fallback (v12)
- Bytecode: assignment expressions that require pattern/compound handling now delegate via eval-expression opcode instead of failing lowering.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_AssignmentPatternFallback -count=1`.

# 2026-01-27 — Bytecode placeholder lambda execution (v12)
- Bytecode: placeholder lambda invocation now runs a bytecode program when available; placeholder expressions lower to a dedicated placeholder-value opcode to honor active placeholder frames.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_PlaceholderLambda -count=1`.

# 2026-01-26 — Bytecode placeholder lambda opcode (v12)
- Bytecode: added placeholder lambda opcode to construct @/@n callables in bytecode mode, with parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_PlaceholderLambda -count=1`.

# 2026-01-26 — Bytecode placeholder lambda fixture (v12)
- Fixtures: added exec fixture for placeholder lambdas in bytecode mode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_08_bytecode_placeholder_lambda -count=1`.

# 2026-01-26 — Bytecode implicit member + iterator fixture (v12)
- Fixtures: added exec fixture for implicit members and iterator literals in bytecode mode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_07_bytecode_implicit_iterator -count=1`.

# 2026-01-26 — Bytecode breakpoint opcode (v12)
- Bytecode: added breakpoint opcode delegation with parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_BreakpointExpression -count=1`.

# 2026-01-26 — Bytecode iterator literal opcode (v12)
- Bytecode: added iterator literal opcode delegation with parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_IteratorLiteral -count=1`.

# 2026-01-26 — Bytecode implicit member opcode (v12)
- Bytecode: added implicit member opcode delegation with parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_ImplicitMemberExpression -count=1`.

# 2026-01-26 — Bytecode await fixture (v12)
- Fixtures: added exec fixture for bytecode await default arm.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/12_01_bytecode_await_default -count=1`.

# 2026-01-26 — Bytecode await opcode (v12)
- Bytecode: added await opcode delegation with a manual-waker parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_AwaitExpressionManualWaker -count=1`.

# 2026-01-26 — Bytecode async task lowering (v12)
- Bytecode: spawned tasks now run bytecode when lowering succeeds (fallback to tree-walker on unsupported nodes).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_SpawnExpression -count=1`.

# 2026-01-26 — Bytecode spawn fixture (v12)
- Fixtures: added exec fixture for bytecode spawn + future.value.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/12_01_bytecode_spawn_basic -count=1`.

# 2026-01-26 — Bytecode spawn op (v12)
- Bytecode: added native spawn opcode/lowering plus parity test for future.value().
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_SpawnExpression -count=1`.

# 2026-01-26 — Bytecode or-else opcode (v12)
- Bytecode: added a dedicated or-else opcode that delegates evaluation to the tree-walker for correct raise handling.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM/OrElse -count=1`.

# 2026-01-26 — Bytecode unary/range/cast fixture (v12)
- Fixtures: added exec fixture to cover unary ops, ranges, type casts, and interpolation in bytecode mode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_02_bytecode_unary_range_cast -count=1`.

# 2026-01-26 — Bytecode propagation op (v12)
- Bytecode: added native propagation opcode/lowering so `!` raises in bytecode mode without eval delegation.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode string interpolation (v12)
- Bytecode: added native lowering + VM op for string interpolation with parity test.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode unary/range/typecast ops (v12)
- Bytecode: added native lowering + VM ops for unary, range, and type cast expressions, plus parity tests.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode short-circuit + pipe (v12)
- Bytecode: added native lowering for `&&`/`||` short-circuit and `|>`/`|>>` pipe operators, plus parity tests.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode statement delegation (v12)
- Bytecode: added eval-statement opcode to delegate definitions/imports/return/yield to the tree-walker during bytecode runs.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-26 — Bytecode eval delegation (v12)
- Bytecode: added a generic eval opcode to delegate propagation/or-else/unary/typecast/await/spawn/etc to the tree-walker, with parity tests for propagation and or-else.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode ensure/rethrow (v12)
- Bytecode: added ensure/rethrow opcode delegation and parity tests; added a bytecode-friendly ensure fixture.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/11_03_bytecode_ensure_basic -count=1`.

# 2026-01-26 — Bytecode rescue/raise (v12)
- Bytecode: added rescue/raise opcode delegation with parity tests plus a bytecode-friendly rescue fixture.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/11_03_bytecode_rescue_basic -count=1`.

# 2026-01-26 — Bytecode match fixture (v12)
- Fixtures: added exec fixture for bytecode-friendly match (literals, guards, wildcard).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/08_01_bytecode_match_basic -count=1`.

# 2026-01-26 — Bytecode match expressions (v12)
- Bytecode: added match-expression opcode delegation and parity tests for literal patterns + guards.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode lambda-call fixture (v12)
- Fixtures: added exec fixture for bytecode-friendly lambda calls, closure capture, and safe member access.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/07_02_bytecode_lambda_calls -count=1`.

# 2026-01-26 — Bytecode loop expression (v12)
- Bytecode: added loop-expression lowering with break/continue handling and parity tests.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode member-call fixture (v12)
- Fixtures: added exec fixture for bytecode-friendly member access, method calls, and safe navigation (tick suppression on nil).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/09_00_bytecode_member_calls -count=1`.

# 2026-01-26 — Bytecode if/index fixture (v12)
- Fixtures: added exec fixture for bytecode-friendly if/elsif/else with array/map index assignment and aggregation.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/08_01_bytecode_if_indexing -count=1`.

# 2026-01-26 — Bytecode loop fixture (v12)
- Fixtures: added exec fixture for bytecode-friendly while/for loops with continue and accumulation.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/08_02_bytecode_loop_basics -count=1`.

# 2026-01-26 — Bytecode map literal fixture (v12)
- Fixtures: added exec fixture to exercise bytecode-friendly map literal + spread evaluation (size, sum, contains checks).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_01_bytecode_map_spread -count=1`.

# 2026-01-26 — Bytecode for loops (v12)
- Bytecode: added for-loop opcode that delegates to tree-walker evaluation; parity tests cover array iteration and break payloads.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode while loops (v12)
- Bytecode: added while-loop lowering plus break/continue handling with scope unwinding; added parity tests for while loops (including break/continue).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`.

# 2026-01-26 — Bytecode map literals (v12)
- Bytecode: added map literal opcode/lowering; parity tests cover direct map literal and spread semantics using kernel HashMap helpers.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Bytecode array literals (v12)
- Bytecode: added array literal opcode/lowering and exercised with index access tests.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Bytecode index expressions (v12)
- Bytecode: added index get/set ops plus lowering for index expressions and index assignments, sharing interpreter index helpers.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Bytecode struct literals (v12)
- Bytecode: added struct definition/literal ops so named struct literals can be evaluated in bytecode mode.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Bytecode member access + calls (v12)
- Bytecode: added member access op + call-site handling for member callees, dotted identifiers, and safe `?.` calls; added dup/jump-if-nil op for short-circuiting.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Bytecode funcs + extern isolation (v12)
- Bytecode: added lowering/VM support for function definitions, lambda expressions, and direct function calls.
- Extern host: salted the extern cache hash per interpreter session to avoid Go plugin reuse across runs (prevents fixture parity cross-talk with stateful externs).
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM -count=1`; `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity -count=1`.

# 2026-01-25 — Go bytecode VM baseline (v12)
- Interpreter: added a minimal Go bytecode VM + lowering (literals, identifiers, :=/=, blocks, if/elsif/else, binary ops) with tree-walker fallback for unsupported nodes.
- CLI/tests: wired bytecode exec mode to the new backend; added bytecode unit parity tests.
- Tests not run (not requested).

# 2026-01-25 — Exec fixture parity harness (v12)
- Tests: added exec fixture parity checks that compare tree-walker vs bytecode stdout/stderr/exit and typecheck diagnostics (`TestExecFixtureParity`).
- Tests not run (not requested).

# 2026-01-25 — Exec-mode flag + fixture mode runs (v12)
- CLI: added `--exec-mode=treewalker|bytecode` global flag and wired treewalker/bytecode wrappers to pass it.
- Tests: added an exec-mode flag for interpreter fixture tests and updated `v12/run_all_tests.sh` to run fixtures in bytecode mode.
- Docs: parity reporting notes now reference the exec-mode flag.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`.

# 2026-01-25 — Go fixture exporter + interface runtime fixes (v12)
- Tooling: added a Go fixture exporter (`v12/interpreters/go/cmd/fixture-exporter`) and wired `v12/export_fixtures.sh` to use it; moved fixture normalization into the parser package and updated JSON encoding for integer literals.
- Runtime: fixed interface-method receiver selection + generic `Self` checks so interface dictionaries and generic impls no longer trip runtime type mismatches; relaxed impl matching to treat wildcard generics as compatible with concrete args.
- Docs: removed the outdated “Go exporter TODO” note from the manual.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`.

# 2026-01-25 — v12 fork + Go-only toolchain
- Forked v11 → v12 and added `spec/full_spec_v12.md` plus `spec/TODO_v12.md`.
- Removed the TypeScript interpreter from v12; introduced Go-only CLI wrappers `abletw` (tree-walker) and `ablebc` (bytecode).
- Updated root docs/scripts to default to v12 and Go-only test runners; v10/v11 frozen.
- Tests not run (workspace + docs refactor).

# 2026-01-25 — Interface dictionary fixture coverage (v12)
- Fixtures: added exec coverage for default generic interface methods and interface-value storage (`exec/10_15_interface_default_generic_method`, `exec/10_16_interface_value_storage`) and updated the exec coverage index.
- Plan: removed the interface dictionary fixture-expansion item from `PLAN.md`.
- Tests: `./run_all_tests.sh`; `./run_stdlib_tests.sh`.

# 2026-01-24 — Iterator interface returns + constraint-arity fixture cleanup (v11)
- Go interpreter: treat Iterator interface return values as iterators during for/each by accepting `IteratorValue` in `adaptIteratorValue`.
- Fixtures: removed duplicate constraint-interface-arity diagnostics from exported manifests via the TS fixture definitions; re-exported fixtures.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict go test ./pkg/interpreter -run 'TestFixtureParityStringLiteral/errors/constraint_interface_arity' -count=1`; `./run_all_tests.sh --version=v11`.

# 2026-01-24 — Interface dictionary arg dispatch + fixture expansion (v11)
- Go interpreter: coerce interface-typed generic values into interface dictionaries so interface arguments are preserved for bindings, params, and return coercions.
- Fixtures: added interface dictionary exec coverage for default chains, overrides, named impl + inherent method calls, interface inheritance, interface-arg dispatch (bindings/params/returns), and union-target dispatch; added AST error fixtures for ambiguous impl constraints + missing interface methods; updated exec coverage index and typecheck baseline.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_11_interface_generic_args_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_12_interface_union_target_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_13_interface_param_generic_args -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_14_interface_return_generic_args -count=1`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_13_interface_param_generic_args bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_14_interface_return_generic_args bun run scripts/run-fixtures.ts`.

# 2026-01-24 — Named impl method resolution fix (v11)
- Interpreters (TS + Go): attach named-impl context to impl methods so default methods (and peers) can resolve sibling methods via `self.method()` in the same impl.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_05_interface_named_impl_defaults bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_06_interface_generic_param_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_05_dynimport_interface_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_05_interface_named_impl_defaults`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_06_interface_generic_param_dispatch`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_05_dynimport_interface_dispatch`.

# 2026-01-24 — Unified Future regression passes + fs rename fix (v11)
- Stdlib fs (TS): switched `fs_remove`/`fs_rename` externs to sync Node calls to avoid flaky exists checks during stdlib runs.
- Plan: removed the unified Future regression-pass TODO after completing the full sweeps.
- Tests: `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-24 — Go test wording cleanup (v11)
- Go tests: updated concurrency/await diagnostics to say “future handle” instead of “proc handle”.
- Tests not run (text-only change).

# 2026-01-24 — Interface dictionary exec coverage expansion (v11)
- Exec fixtures: added coverage for named impl disambiguation across packages with defaults, generic-constraint dispatch across packages, and dynimport-returned interface values.
- Coverage index: registered `exec/10_05_interface_named_impl_defaults`, `exec/10_06_interface_generic_param_dispatch`, and `exec/13_05_dynimport_interface_dispatch`.
- Tests not run (fixture-only changes).

# 2026-01-24 — Unified Future model naming cleanup (v11)
- TypeScript tests: renamed `proc_spawn_*` concurrency tests/helpers to `future_spawn_*` and updated imports (including await tests).
- Docs: updated `v11/interpreters/ts/README.md` and `v11/stdlib/src/README.md` to use Future terminology.
- Plan: collapsed the unified Future model checklist to the remaining regression-pass item.
- Tests not run (rename + docs + plan updates only).

# 2026-01-24 — Bytecode VM prototype (v11)
- TypeScript interpreter: added a minimal stack-based bytecode VM plus a small AST->bytecode lowering path (literals, identifiers, `:=`/`=`, `+`, blocks).
- Tests: added VM-vs-tree-walker conformance checks for literals, assignment + arithmetic, and module bodies.
- Tests: `cd v11/interpreters/ts && bun test test/vm/bytecode_vm.test.ts`.

# 2026-01-24 — Core IR + runtime ABI design (v11)
- Design: expanded `v11/design/compiler-interpreter-vision.md` with a typed core IR outline and runtime ABI surface (interface dictionaries, concurrency, errors, dynamic hooks).
- Tests not run (design-only update).

# 2026-01-24 — Stdlib copy helpers speedup (v11)
- Stdlib fs: routed `copy_file`/`copy_dir` through host externs (Go + TS) and removed the Able-level directory traversal to keep `copy_dir` under the test-time budget.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able --format tap --name "able.fs::copies directory trees with overwrite control"` (≈59.6s); `./run_stdlib_tests.sh`; `./run_all_tests.sh --version=v11`.

# 2026-01-24 — Interface dispatch fixture coverage (v11)
- Exec fixtures: added `exec_10_04_interface_dispatch_defaults_generics` to cover cross-package default interface methods + generic interface method calls on interface-typed values.
- Coverage index: registered the new exec fixture entry.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_04_interface_dispatch_defaults_generics bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.

# 2026-01-24 — Future handle test/fixture cleanup (v11)
- TypeScript tests: align await stdlib integration helper with `Future.value() -> !T` by returning `!Array bool`.
- Exec fixtures: update `12_04_future_handle_value_view` stdout expectations to match the renamed output text.
- Go tests: disambiguate duplicate future-handle/serial-executor test names introduced during the future renames.
- Tests: `./v11/export_fixtures.sh`; `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-24 — Regex quantifier parsing + scan (v11)
- Stdlib regex: implemented literal-token parsing with quantifiers (`*`, `+`, `?`, `{m}`, `{m,}`, `{m,n}`), updated match/find_all/scan to use token spans, and fixed a match-case return syntax issue.
- Design: updated `v11/design/regex-plan.md` to reflect the partial Phase 1 status and active regex fixture.
- Tests: `./v11/ablets check v11/stdlib/src/text/regex.able`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_02_regex_core_match_streaming bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.

# 2026-01-23 — Constraint arity regression coverage (v11)
- Typechecker (TS): added constraint interface arity diagnostics for missing/mismatched interface type arguments.
- Fixtures/tests: added `errors/constraint_interface_arity` fixture (calls wrapped in a non-invoked helper to avoid runtime errors), plus new TS/Go typechecker regression tests for constraint arity.
- Baseline: regenerated `v11/fixtures/ast/typecheck-baseline.json` after the fixes.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`.

# 2026-01-23 — Typechecker default enforcement (v11)
- Harnesses: defaulted fixture/test/parity typechecking to strict, with explicit warn/off overrides; run_all_tests now always passes `ABLE_TYPECHECK_FIXTURES` (strict by default) into fixture/parity/Go test runs.
- Docs: refreshed v11 + interpreter readmes and parity reporting notes to reflect strict defaults and explicit overrides.
- Tests not run (docs + harness configuration only).

# 2026-01-23 — Stdlib typecheck verification (v11)
- Verified stdlib packages typecheck cleanly in TS + Go by importing all stdlib packages via a temporary stdlib typecheck harness.
- Verified `.examples/foo.able` runs with strict typechecking in both interpreters.
- Tests: `./v11/ablets check tmp/stdlib_typecheck.able`; `./v11/ablego check tmp/stdlib_typecheck.able`; `./v11/ablets .examples/foo.able`; `./v11/ablego .examples/foo.able`.

# 2026-01-22 — Stdlib Eq/Ord/Hash audit (v11)
- Stdlib: audited `v11/stdlib/src` for generic `Eq`/`Ord`/`Hash` constraints and kernel alias usage; no type-arg uses remain, so the PLAN item was cleared.
- Tests not run (audit + plan/log updates only).

# 2026-01-22 — Go parser nested generics fix (v11)
- Go parser: avoid flattening parenthesized generic applications in type argument lists so nested types like `Iterable (Option String)` stay nested; fixes `TestParseInterfaceCompositeNestedGenerics` and exec fixture `04_03_type_expression_syntax`.
- Tests: `cd v11/interpreters/go && go test ./pkg/parser -run TestParseInterfaceCompositeNestedGenerics -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/04_03_type_expression_syntax -count=1`; `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-22 — TS typechecker import symbol scoping (v11)
- Typechecker (TS): record symbol origins for imports/locals and filter unqualified/UFCS free-function resolution to the explicitly imported symbol source; builtins tagged for scope filtering.
- Tests: `bun test test/typechecker/function_calls.test.ts` and `bun test test/typechecker/duplicates.test.ts` in `v11/interpreters/ts`.
- Tests: `./run_stdlib_tests.sh --version=v11`.

# 2026-01-21 — TS typechecker import scoping for functions (v11)
- Typechecker (TS): track current package for function infos, filter call resolution to imported packages, and avoid cross-package duplicate declaration errors in stdlib/test runs.
- Tests: `./run_stdlib_tests.sh --version=v11`.

# 2026-01-21 — Full test runs (v11)
- Tests: `./run_all_tests.sh --version=v11`.

# 2026-01-21 — Full test runs (v11)
- Tests: `./run_all_tests.sh --version=v11` (passed).
- Tests: `./run_stdlib_tests.sh --version=v11` failed in TS stdlib due to duplicate declaration diagnostics (e.g., `able.text.regex` vs `able.text.string`, `able.spec` vs `able.spec.assertions`, and duplicated helpers across stdlib collection smoke tests); Go stdlib tests passed.

# 2026-01-21 — TS interpreter types refactor (v11)
- Refactored `v11/interpreters/ts/src/interpreter/types.ts` into `types/format.ts`, `types/primitives.ts`, `types/unions.ts`, `types/structs.ts`, and `types/helpers.ts` while keeping the public augmentations unchanged.
- Tests not run (refactor only).

# 2026-01-21 — TS impl resolution refactor (v11)
- Split `v11/interpreters/ts/src/interpreter/impl_resolution.ts` into stage modules for constraints, candidates, specificity ranking, defaults, and diagnostics under `v11/interpreters/ts/src/interpreter/impl_resolution/`.
- Tests not run (refactor only).

# 2026-01-21 — TS implementation collection refactor (v11)
- Split `v11/interpreters/ts/src/typechecker/checker/implementation-collection.ts` into collection vs validation/self-pattern helpers in `implementation-validation.ts`.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/implementation_validation.test.ts`.

# 2026-01-21 — TS function call refactor (v11)
- Split `v11/interpreters/ts/src/typechecker/checker/function-calls.ts` into call-shape parsing, overload resolution, and diagnostics helpers (`function-call-parse.ts`, `function-call-resolve.ts`, `function-call-errors.ts`).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/ts && bun test test/typechecker/method_sets.test.ts`; `cd v11/interpreters/ts && bun test test/typechecker/apply_interface.test.ts`.

# 2026-01-21 — TypeCheckerBase size trim (v11)
- Extracted `checkFunctionDefinition`/`checkReturnStatement` into `v11/interpreters/ts/src/typechecker/checker/checker_base_functions.ts` and trimmed `v11/interpreters/ts/src/typechecker/checker_base.ts` under 900 lines.
- Tests not run (refactor only).

# 2026-01-21 — Interfaces fixture refactor (v11)
- Split `v11/interpreters/ts/scripts/export-fixtures/fixtures/interfaces.ts` into modular fixture files under `v11/interpreters/ts/scripts/export-fixtures/fixtures/interfaces/`.
- Tests not run (fixture refactor only).

# 2026-01-20 — Go typechecker builtin arity fallback (v11)
- Go typechecker: prefer builtin arity when generic base names are not found in the env, fixing Array T without explicit imports in parity examples.
- State: `./run_all_tests.sh --version=v11` green; parity examples passing.
- Next: resume PLAN TODOs (regex parser + quantifiers).

# 2026-01-19 — Eager/lazy collections verification (v11)
- Design: noted that `String`/`StringBuilder` keep bespoke eager `map`/`filter` (char-only) until an HKT wrapper exists (`v11/design/collections-eager-lazy-split.md`).
- Plan/spec: removed the eager/lazy TODO item from `PLAN.md` and cleared the resolved spec TODO entry (`spec/TODO_v11.md`).
- Tests: `./v11/ablets test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_builder.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/text/string_builder.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found); `./v11/ablego test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found).

# 2026-01-19 — Iterator.map generator yield fix (v11)
- Stdlib: wrapped `Iterator.map` match-arm yields in a block so generator resumes advance instead of repeating the same value.
- Tests: `./v11/ablets test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found).

# 2026-01-19 — Eager/lazy collections wiring (v11)
- Stdlib: moved lazy adapters to `Iterator`, introduced HKT `Enumerable` + `lazy` bridge, and refactored collection `Enumerable` impls/`Iterable` overrides.
- Collections: added shared `MapEntry` in `able.collections.map`, switched `PersistentMap` to `Iterable` entries, removed `Enumerable` from `BitSet`, and added HashMap value iteration plus value-based `map` (keys preserved).
- Plan: removed completed eager/lazy sub-steps from `PLAN.md`.
- Tests not run (stdlib + plan updates only).

# 2026-01-17 — Iterable collect type refs (v11)
- Interpreters: bind generic type parameters as runtime type refs so static interface methods (like `C.default()`) resolve in TS + Go.
- Stdlib/tests: disambiguate `collect` by terminating the `C.default()` statement and import `able.collections.array` in iteration tests so `Array` impls load for the Go typechecker.
- Tests: `./v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.

# 2026-01-15 — Hash/Eq fixture and test coverage (v11)
- Fixtures: added AST fixtures for primitive hashing, kernel hasher availability, custom Hash/Eq, and collision handling; added exec fixtures for primitive hashing plus custom Hash/Eq + collisions; updated exec coverage index.
- Tests: added TS + Go unit coverage for hash helper builtins and kernel HashMap dispatch (custom + collision keys).
- Tests not run (edited code + fixtures only).

# 2026-01-15 — Remove host hasher bridges (v11)
- Kernel: dropped the `__able_hasher_*` extern declarations and the unused `HasherHandle` alias so hashing flows through `KernelHasher` only.
- Interpreters: removed host hasher state/builtins from Go + TypeScript, along with runtime stub and typechecker builtin entries.
- Docs/spec: scrubbed hasher bridge references from the kernel contract and extern execution/design notes.
- Tests not run (edited code + docs only).

# 2026-01-15 — Kernel Hash/Eq runtime alignment (v11)
- Kernel: added primitive `Eq`/`Ord`/`Hash` impls (ints/bool/char/String) plus float-only `PartialEq`/`PartialOrd`, and wired the Able-level FNV-1a hasher with raw byte helpers.
- Stdlib: trimmed duplicate interface impls and routed map hashing through the sink-style `Hash.hash` API.
- Interpreters: hash map kernels now dispatch `Hash.hash`/`Eq.eq`; numeric equality follows IEEE semantics; Go/TS typecheckers exclude floats from `Eq`/`Hash`.
- Fixtures: added float equality + hash-map key rejection exec coverage.
- Tests not run (edited code + docs only).

# 2026-01-15 — Kernel interfaces + Hash/Eq plan (v11)
- Documented the kernel-resident interface plan for primitive `Hash`/`Eq`, including runtime/stdlib/typechecker alignment, spec deltas, and fixture coverage (`v11/design/kernel-interfaces-hash-eq.md`, `v11/design/stdlib-v11.md`).
- Captured the spec update checklist in `spec/TODO_v11.md` and expanded the roadmap work breakdown in `PLAN.md`.
- Tests not run (planning/doc updates only).

# 2026-01-15 — Manual syntax alignment (v11)
- Manual docs now reflect v11 lexical rules and pipe semantics: line comments use `##`, string literals can be double-quoted or backticks with interpolation, and pipe docs no longer reference a `%` topic token (`v11/docs/manual/manual.md`, `v11/docs/manual/variables.html`).
- Tests not run (docs-only changes).

# 2026-01-15 — Primitive Hash/Eq constraints (v11)
- TypeScript typechecker now treats primitive numeric types as satisfying `Hash`/`Eq` constraints (matching Go) and the example iterates directly over `String` so `for` sees an `Iterable` (`v11/interpreters/ts/src/typechecker/checker/implementation-constraints.ts`, `.examples/foo.able`).
- Tests: `./v11/ablets .examples/foo.able`; `./v11/ablego .examples/foo.able`.

# 2026-01-13 — Runtime diagnostics formatting (v11)
- Runtime errors now emit `runtime:` diagnostics with locations + call-site notes in both interpreters; CLI/runtime harnesses share the same formatter.
- Added Go runtime diagnostic formatting test and updated exec fixture stderr expectations to include locations.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run RuntimeDiagnostics`; `cd v11/interpreters/ts && bun test test/runtime/runtime_diagnostics.test.ts`.

# 2026-01-13 — Parser diagnostics formatting (v11)
- Parser diagnostics: route syntax/mapping failures through shared formatting, add span/expectation extraction from tree-sitter/mapper nodes, and normalize parser error messages for CLI output.
- CLI/tests: TS + Go loaders now surface parser diagnostics in the same format as typechecker output; added parser diagnostic tests for localized expectation messages.
- Tests: `cd v11/interpreters/ts && bun test test/parser/diagnostics.test.ts`; `cd v11/interpreters/go && go test ./pkg/driver -run ParserDiagnostics`.

# 2026-01-13 — Diagnostics groundwork + union normalization (v11)
- Design: added diagnostics overhaul note with warning policy, span/notes shape, and union normalization rules (`v11/design/diagnostics-overhaul.md`).
- Typecheckers: normalized unions with nullable/alias expansion, redundant-member warnings, and single-member collapse (TS + Go); added warning severity plumbing and location end spans.
- CLI/fixtures: warning-prefixed formatting for typechecker diagnostics in TS + Go; Go CLI diagnostics now use location-first formatting.
- Tests: `bun test test/typechecker/union_normalization.test.ts`; `go test ./pkg/typechecker -run 'UnionNormalization|NormalizeUnionTypes'`.

# 2026-01-13 — Lowercase path package cleanup (v11)
- Stdlib: ensured the path module works under the lowercase package name by importing `Path` into stdlib tests and avoiding module shadowing in `fs.write_bytes`.
- Tests: `./v11/ablets test v11/stdlib/tests/path.test.able`; `./run_stdlib_tests.sh --version=v11`; `./run_all_tests.sh --version=v11`.

# 2026-01-13 — Stdlib fs convenience helpers (v11)
- Stdlib fs: added `read_lines`, `write_lines`, `copy_file`, `copy_dir`, `touch`, `remove_file`, and `remove_dir` helpers; `touch` now uses host `utimes`/`Chtimes`, `copy_dir` uses an explicit task stack to avoid iterator re-entrancy, and `fs_path` prioritizes string inputs to keep Go/TS behavior aligned.
- Tests: expanded `v11/stdlib/tests/fs.test.able` to cover line IO, copy helpers + overwrite behavior, touch, and explicit remove wrappers.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`.

# 2026-01-13 — Path API completion (v11)
- Stdlib Path: added `current`/`home`/`absolute`/`expand_home`/`normalize` helpers, `/` join sugar, and filesystem wrappers (`exists`, `is_file`, `is_dir`, `stat`, `read_text`, `write_text`).
- Go typechecker: allow `/` to resolve via `Div` interface implementations when operands are non-numeric.
- Tests: expanded `v11/stdlib/tests/path.test.able` with cwd/home/absolute/expand_home, join sugar, and fs helper coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.

# 2026-01-13 — Proc cancellation test alignment (v11)
- TS tests: move cooperative cancellation check to run after `proc_yield`, aligning with proc resume semantics.
- Tests: `./run_all_tests.sh --version=v11`.

# 2026-01-13 — TS call continuation yields (v11)
- TS interpreter: preserve function-call environments across proc yields so async extern calls resume without replaying earlier statements; stop resetting block/module indices on manual yields.
- Tests: `./run_stdlib_tests.sh`.

# 2026-01-11 — TS concurrency continuation fixes (v11)
- TS interpreter: added module-level continuation state so entrypoint yields resume without replaying declarations; await commit now resumes across cooperative yields; future/proc awaitBlocked handling unified; proc/future value waits handle immediate waker completion.
- Tests: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.

# 2026-01-11 — Stdlib OS/process/term scaffolding (v11)
- Stdlib: added `able.os`, `able.process`, and `able.term` modules with host externs and Able-level helpers (`ProcessSpec`, `ProcessStatus`, `ProcessSignal`, `TermSize`, line editor, env/cwd helpers).
- IO: extended TS `IoHandle` support to allow stream-backed handles (child process pipes) in `io_read`/`io_write`/`io_flush`/`io_close`.
- Go/TS hosts: implemented process spawning + wait/kill plumbing with cached status, plus term TTY hooks (Go raw mode/size limited to linux/amd64).
- Tests: not run.

# 2026-01-11 — Go extern host plugin exports (v11)
- Go interpreter: capture extern package name at definition time so host calls resolve after module evaluation; generate exported plugin wrappers for extern functions and bump extern cache version.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run ExecFixtures`.
- Full sweep: `./run_all_tests.sh --version=v11`.

# 2026-01-10 — Extern host execution (v11)
- TS: added per-package host module caching with `host_error`/handle aliases and routed externs through dynamic imports; empty non-kernel externs now raise runtime errors.
- Go: added per-package plugin generation + caching, with extern calls marshaled through host functions; introduced `HostHandle` runtime values for IoHandle/ProcHandle.
- Fixtures/tests: updated extern exec fixture expectation and extern unit tests.
- Tests: `cd v11/interpreters/ts && bun test test/basics/externs.test.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run Extern`.

# 2026-01-10 — String interpolation continuation (v11)
- TS interpreter: added continuation state for string interpolation so time-slice yields resume mid-interpolation without re-running completed parts (fixes exec string helpers under scheduler yields).
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_12_01_stdlib_string_helpers bun run scripts/run-fixtures.ts`.
- Follow-up: ran `./run_all_tests.sh --version=v11`; all suites green and parity report refreshed at `v11/tmp/parity-report.json`.
- TS interpreter: switched the `read_text` host extern stub to async `fs.promises.readFile` so blocking IO can suspend only the calling proc via the native Promise bridge.

# 2026-01-02 — Placeholder member access in typechecker (v11)
- Go typechecker: treat member access on placeholder expressions as unknown instead of rejecting the access, fixing `@.method()` in pipe shorthand.
- Tests: `bun test test/parity/examples_parity.test.ts -t pipes_topics/main.able`; `./run_all_tests.sh --version=v11`.

# 2026-01-02 — String host accepts struct-backed String (v11)
- Go interpreter: `__able_String_from_builtin` now accepts struct-backed `String` values by extracting the `bytes` field, fixing stdlib string helpers when `String` is a struct instance.
- Go interpreter tests: added coverage for struct-backed `String` conversions in the string host builtins.
- Tests: `go test ./pkg/interpreter -run StringFromBuiltin`; `./v11/ablego test v11/stdlib/tests/text/string_methods.test.able`.

# 2026-01-02 — Parse empty function types + numeric method resolution (v11)
- Go parser: treat empty parenthesized types as zero-arg function parameters when parsing `() -> T`, preventing `()->T` from collapsing into a simple type.
- Go typechecker: allow method-set lookup for integer/float values so numeric helpers like `to_r` resolve after importing `able.core.numeric`.
- Tests: `./v11/ablego test v11/stdlib/tests/core/numeric_smoke.test.able`; `go test ./pkg/parser ./pkg/typechecker`.

# 2026-01-02 — Type matching fixes + Clone primitives (v11)
- Go interpreter: expanded runtime type matching to compare alias-expanded value types, let unknown type names match struct instances (generic union members like `T`), and accept struct-backed `String` values; added Result/Option generic matching fallback.
- Go interpreter: primitives now satisfy `Clone` via built-in method lookup, fixing stdlib TreeSet String constraints without extra imports.
- Tests: `ABLE_TYPECHECK_FIXTURES=off ./v11/ablego test v11/stdlib/tests --list`.

# 2026-01-02 — Method resolution disambiguation (v11)
- Go interpreter: tightened UFCS/member lookup to filter candidates by in-scope overloads and disambiguate same-name types via method target matching, fixing local `Channel.send` collisions while preserving alias reexports and Ratio methods.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`.

# 2026-01-02 — Serial executor fairness ordering (v11)
- Go interpreter: queue freshly created serial tasks ahead of resumed ones so `proc_yield` round-robins even when procs are created in separate eval calls.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`; `./run_all_tests.sh --version=v11`.

# 2026-01-01 — Trailing lambdas + struct definition bindings (v11)
- Reattached trailing lambda bodies inside expression lists in the Go parser so call arguments match block parsing and `suite.it(...) { ... }` forms register correctly.
- Go interpreter now treats native/bound/partial callables as valid function values for `fn(...)` type checks.
- Go runtime records struct definitions separately from value bindings (including imports) so struct literals resolve even when constructors shadow names.
- Tests: `./run_all_tests.sh --version=v11`; `./v11/ablego test /home/david/sync/projects/able/v11/stdlib/tests/simple.test.able`; `./v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests/simple.test.able`.

# 2026-01-01 — Regex literal matching (v11)
- Implemented literal-only regex compile/match/find_all using byte-span search and stored literal bytes in the regex handle.
- Updated the stdlib matcher test to exercise substring matches and refreshed docs to reflect partial regex support.
- Tests: `cd v11/interpreters/ts && ABLE_TYPECHECK_FIXTURES=off bun run scripts/run-module.ts test ../../stdlib/tests`.
