# Able Project Log

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

# 2025-12-31 — Stdlib/runtime test fixes (v11)
- Resolved Iterable/Enumerable method ambiguity by preferring explicit impl methods over default interface methods (fixes Vector.each).
- Aligned hasher host builtins to return i64 handles/hashes (TS runtime + stubs + typechecker) and updated stdlib string integration expectations for subString errors.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-31 — Add idiomatic Able style guide (v11)
- Added a new documentation guide covering idiomatic Able conventions with examples at `v11/docs/idiomatic-able.md`.
- Cleared completed PLAN TODO entries now that the queue is empty.
- Tests: not run (docs-only changes).

# 2025-12-31 — Fix stdlib spec test discovery filters (v11)
- Corrected `able.spec` discovery filtering so empty exclude lists no longer filter every example; regenerated `v11/stdlib/src/spec.able`.
- Tests: `cd v11/interpreters/ts && ABLE_TYPECHECK_FIXTURES=off bun run scripts/run-module.ts test ../../stdlib/tests --list`.

# 2025-12-30 — Stdlib test syntax cleanup (v11)
- Rewrote stdlib tests to avoid non-spec syntax: replaced `let`/`let mut` with `:=`, switched prefix `match` to postfix form, wrapped `raise` in match/lambda bodies, removed `TestEvent.` qualifiers, and used struct-literal constructors for NFAChar.
- Updated static calls in tests from `Type::method` to `Type.method` and replaced `_` lambda parameters with `_ctx`.
- Tests: `bun run scripts/run-module.ts test ../../stdlib/tests` (fails: `able.text.automata` package missing from stdlib src; still in quarantine).

# 2025-12-30 — Verbose anonymous function syntax (v11)
- Added tree-sitter support for `fn(...) { ... }` anonymous functions and mapped them to `LambdaExpression` with generics/where clauses in TS/Go parsers.
- Added exec fixture `exec/07_02_01_verbose_anonymous_fn` plus conformance/coverage index updates.
- Tests: `bun test test/parser/fixtures_parser.test.ts`; `ABLE_FIXTURE_FILTER=07_02_01_verbose_anonymous_fn bun run scripts/run-fixtures.ts`; `go test -a ./pkg/interpreter -run TestExecFixtures/07_02_01_verbose_anonymous_fn$`.

# 2025-12-30 — Go parser interface-arg test fix (v11)
- Fixed Go parser test coverage so `TestParseImplInterfaceArgsParentheses` is a real test (no longer embedded in a raw string for the breakpoint fixture).
- Regenerated the tree-sitter parser/wasm and forced a Go rebuild so the cgo parser picks up the updated grammar.
- Tests: `bun test test/parser/fixtures_parser.test.ts`; `go test -a ./pkg/parser -run TestParseImplInterfaceArgsParentheses`; `./run_all_tests.sh --version=v11`.

# 2025-12-30 — Interface arg parentheses (Option 2b)
- Updated `spec/full_spec_v11.md` to clarify that interface args are space-delimited type expressions and generic applications only form when parenthesized.
- Updated tree-sitter grammar + TS/Go type parsers for `interface_type_*` nodes, and removed interface-arg arity grouping in TS/Go interpreter + typechecker.
- Added TS/Go parser tests asserting interface-arg splitting vs parenthesized generic args; rebuilt tree-sitter parser/wasm.
- Docs: noted parenthesized generic applications in `v11/parser/README.md`.
- Tests: not run (parser build via `npm run build`).

# 2025-12-30 — Go `able test` CLI wiring + integration tests (v11)
- Implemented Go `able test` CLI (args/targets, discovery, harness/run, reporters, exit codes) plus loader support for include packages and interpreter helpers for array/method access.
- Added lightweight Go CLI tests for `able test` list/dry-run using a stub stdlib harness to keep the suite fast and isolated.
- Tests: `cd v11/interpreters/go && go test ./cmd/able -run TestTestCommand`.

# 2025-12-30 — TypeScript `able test` CLI wiring (v11)
- Completed TS CLI `able test` wiring end-to-end (discovery/filter/run/report/exit), including kernel `Array` decoding for reporters/list output and array-length handling for discovery results.
- Bound trailing lambdas inside expression lists so `suite.it("...") { ... }` attaches correctly within `describe` bodies.
- Updated stdlib tests to `import able.spec.*` so suite method names are in scope; fixed a malformed `able.spec.assertions` import.
- Shifted testing protocol `Framework` methods to return `?Failure` and regenerated `v11/stdlib/src/spec.able`.
- Tests: `cd v11/interpreters/ts && bun test test/cli/run_module_cli.test.ts`.

# 2025-12-30 — Drop snapshots from testing framework (v11)
- Removed snapshot matchers, stores, and CLI flag references from the stdlib testing DSL; deleted the snapshot store design doc and scrubbed snapshot mentions from testing docs/design notes/spec.
- Updated stdlib tests and the TS CLI skeleton tests to drop `match_snapshot`/`--update-snapshots`.
- Tests: not run.

# 2025-12-30 — Testing module suffix policy (v11)
- Codified test modules as `.test.able` or `.spec.able` in `spec/full_spec_v11.md` (new Tooling: Testing Framework section) and resolved the open decision in `v11/design/testing-plan.md`.
- Removed the suffix-policy TODO from `PLAN.md` now that the rule is set.
- Tests: not run.

# 2025-12-30 — Split stdlib testing into able.test + able.spec (v11)
- Moved testing protocol/harness/reporters/snapshots into `v11/stdlib/src/test` and the user DSL into `v11/stdlib/src/spec`/`v11/stdlib/src/spec.able`, renaming packages and framework id to `able.spec`.
- Migrated quarantine stdlib tests into `v11/stdlib/tests` and updated imports to `able.spec` + `able.test.*`.
- Updated testing docs/spec references to the new namespace split and renamed internal `RspecFramework` to `SpecFramework`.
- Tests: not run (stdlib/CLI integration pending).

# 2025-12-29 — Alias re-export follow-up validation (v11)
- Progress: ran `./run_all_tests.sh --version=v11`; all green.
- State: parity report refreshed at `v11/tmp/parity-report.json`.
- Next: resume PLAN backlog (regex stdlib expansion + tutorial cleanup).

# 2025-12-29 — Typechecker method-set dedupe + alias impl diag alignment (v11)
- Prevented duplicate method-set candidates by tagging method-set function infos, skipping them during member lookup, and marking static method-set entries as type-qualified.
- Kept method-set targets in generic form to preserve where-clause constraint enforcement, and deduped method sets/impl records across session preludes.
- Aligned alias re-export impl ambiguity diagnostics to the canonical target label and updated the baseline accordingly.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-29 — Alias re-export impl ambiguity fixture (v11)
- Added AST error fixture `errors/alias_reexport_impl_ambiguity` covering duplicate impl registration when alias-attached impls target the same canonical type.
- Filled `module.json` for alias re-export method/impl ambiguity fixtures and updated the typecheck baseline for the new diagnostic.
- Tests: `cd v11/interpreters/ts && bun test test/parser/fixtures_mapper.test.ts`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`.

# 2025-12-29 — Document Euclidean division (v11)
- Clarified in `spec/full_spec_v11.md` that `//`, `%`, and `/%` use Euclidean integer division (non-negative remainder), with examples and negative divisor behavior.

# 2025-12-29 — Add floor division helpers (v11)
- Added stdlib `div_floor`/`mod_floor`/`div_mod_floor` functions and methods for `i32`, `i64`, `u32`, `u64`, plus updated numeric helper fixture coverage and spec text.
- Removed the integer floor-division TODO from `spec/TODO_v11.md` now that helpers are implemented.

# 2025-12-29 — Spec TODO audit (v11)
- Trimmed `spec/TODO_v11.md` to the remaining items with detailed scope and open questions (alias/re-export method propagation).

# 2025-12-29 — Quarantine host regex hooks (v11)
- Removed TS/Go regex host hooks and wiring; stdlib `able.text.regex` now raises `RegexUnsupportedFeature` instead of calling host engines.
- Quarantined the exec regex fixture (`exec/14_02_regex_core_match_streaming`) and marked coverage as planned; TS regex integration test now skipped pending the stdlib engine.
- Updated conformance/testing/manual docs and regex design notes to reflect stdlib-only regex plans.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-29 — v11 fixture sweep
- Progress: ran `./run_all_tests.sh --version=v11 --fixture`.
- Current state: TypeScript fixtures, parity harness, and Go tests all green; parity report at `v11/tmp/parity-report.json`.
- Next: continue PLAN backlog (regex stdlib expansion and tutorial cleanup).

# 2025-12-29 — Regex core match/streaming fixture (v11)
- Implemented stdlib regex core helpers (`Regex.compile`, `is_match`, `match`, `find_all`, `scan` + streaming `RegexScanner.feed/next`) with host-backed compile/find externs.
- Added TS/Go regex host builtins and runtime state for compiled handles, plus span/match struct construction (empty groups/named groups for now).
- Added exec fixture `exec/14_02_regex_core_match_streaming`, updated conformance plan + coverage index, and removed the PLAN backlog item.
- Updated regex stdlib integration test and testing matcher docs now that regex helpers are live.
- Go runtime now treats `IteratorEnd {}` as matching the IteratorEnd sentinel during return type checks, aligning iterator method returns with pattern matching behavior.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_02_regex_core_match_streaming bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_02_regex_core_match_streaming$`.

# 2025-12-29 — Interface dispatch fixture + IteratorEnd runtime alignment (v11)
- Added exec fixture `exec/14_01_language_interfaces_index_apply_iterable` covering Index/IndexMut, Iterable/Iterator, and Apply dispatch; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- Stdlib `iteration.each` now matches `IteratorEnd {}` before visiting values, keeping IteratorEnd from flowing into visitor callbacks.
- Go runtime now treats `IteratorEnd` as a first-class type in `matchesType`/type inference and equality comparisons, aligning match behavior with TS.
- Added exec fixture `exec/14_01_operator_interfaces_arithmetic_comparison` covering arithmetic/comparison operator interface dispatch with Display/Clone/Default helpers; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- TS/Go runtimes now route unary `-` through `Neg` interface impls when operands are non-numeric, and both runtimes dispatch comparison operators via `Eq`/`PartialEq` and `Ord`/`PartialOrd` when available.
- Go static method lookup now includes impl methods so interface-provided statics like `Default.default()` resolve on types.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_06_stdlib_package_resolution bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_07_search_path_env_override bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_01_language_interfaces_index_apply_iterable bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_06_stdlib_package_resolution$`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_07_search_path_env_override$`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_01_language_interfaces_index_apply_iterable$`.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_01_operator_interfaces_arithmetic_comparison bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_01_operator_interfaces_arithmetic_comparison$`.
- Added stdlib `os` module with `args()` and runtime `__able_os_args` builtins for TS/Go; added exec fixture `exec/15_02_entry_args_signature` plus coverage/conformance updates and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_02_entry_args_signature bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_02_entry_args_signature$`.
- Added `os.exit` runtime support (TS/Go) with CLI/fixture harness handling; added exec fixture `exec/15_03_exit_status_return_value` plus coverage/conformance updates and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_03_exit_status_return_value bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_03_exit_status_return_value$`.
- Added exec fixture `exec/15_04_background_work_flush` to assert background tasks are not awaited on exit; updated coverage/conformance and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_04_background_work_flush bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_04_background_work_flush$`.
- Added exec fixture `exec/16_01_host_interop_inline_extern` covering extern host bindings; updated coverage/conformance and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=16_01_host_interop_inline_extern bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/16_01_host_interop_inline_extern$`.

# 2025-12-28 — Exec fixtures for errors + concurrency (v11)
- Added `exec/11_03_rescue_rethrow_standard_errors` covering arithmetic/indexing runtime errors, rescue/ensure, and rethrow semantics; updated the exec coverage index and removed the PLAN backlog item.
- Added `exec/12_02_future_fairness_cancellation` covering `future_yield` fairness, cancellation via `future_cancelled`, and `future_flush` queue drains; updated the exec coverage index + conformance plan and removed the PLAN backlog item.
- Added `exec/12_03_spawn_future_status_error` and `exec/12_04_future_handle_value_view` for future status/value/error propagation and handle/value behaviour; updated the exec coverage index + conformance plan and cleared the PLAN items.
- Added `exec/12_05_mutex_lock_unlock` and `exec/12_06_await_fairness_cancellation` for mutex/await semantics; updated the exec coverage index + conformance plan and cleared the PLAN items.
- TS/Go runtimes now raise standard errors (`DivisionByZeroError`, `OverflowError`, `ShiftOutOfRangeError`, `IndexError`) with `Error.value` payloads for rescue matching; `!` propagation now raises any `Error` value and index fallback returns `IndexError` payloads.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=11_03_rescue_rethrow_standard_errors bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_02_future_fairness_cancellation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_03_spawn_future_status_error bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_04_future_handle_value_view bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_05_mutex_lock_unlock bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_06_await_fairness_cancellation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.
- Added `exec/12_07_channel_mutex_error_types` and `exec/13_03_package_config_prelude` for channel/mutex error payloads and package.yml root-name/prelude parsing; updated the exec coverage index + conformance plan and cleared the PLAN items.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_07_channel_mutex_error_types bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_03_package_config_prelude bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(12_07_channel_mutex_error_types|13_03_package_config_prelude)$'`.
- Added `exec/13_04_import_alias_selective_dynimport` covering import aliases, selective renames, and dynimport bindings; updated the exec coverage index + conformance plan and removed the PLAN backlog item.
- TS runtime now treats primitive types as satisfying `Hash`/`Eq` constraints for interface enforcement and returns `IndexError` values on out-of-bounds array assignments to align IndexMut semantics.
- Updated TS division/ratio tests to assert `RaiseSignal` error payloads rather than raw error messages.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-26 — Impl specificity exec fixture + array type-arg dispatch (v11)
- Added `exec/10_02_impl_specificity_named_overrides` covering impl specificity ordering, named impl disambiguation, and HKT targets; updated the exec coverage index, conformance plan, and removed the PLAN backlog item.
- TS runtime now derives array element type arguments during method resolution so concrete vs generic impl selection matches spec intent.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_02_impl_specificity_named_overrides bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod go test ./pkg/interpreter -run TestExecFixtures/10_02_impl_specificity_named_overrides$`.

# 2025-12-24 — Truthiness exec fixture + boolean context alignment (v11)
- Added `exec/06_11_truthiness_boolean_context` to cover truthiness rules, unary `!`, and `&&`/`||` operand returns; updated coverage index + conformance plan and cleared the PLAN item.
- TS/Go runtimes now evaluate `!`, `&&`, and `||` via truthiness (returning operands) and dynimport supports late-bound packages without eager loader failures; dyn refs now re-check privacy at call time.
- Typecheckers no longer require bool for conditions/guards or logical operands; Go typechecker/interpreter tests updated to match truthiness semantics.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_11_truthiness_boolean_context bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_11_truthiness_boolean_context$`; `cd v11/interpreters/go && go test ./pkg/typechecker -run Truthiness`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestLogicalOperandsTruthiness$`; `cd v11/interpreters/go && go test ./pkg/driver -run TestLoaderDynImportDependencies$`.

# 2025-12-24 — Dynamic metaprogramming exec fixture + runtime support (v11)
- Added `exec/06_10_dynamic_metaprogramming_package_object` to cover dyn package creation/lookup, dynamic definitions, and late-bound dynimport redefinitions; updated coverage index + conformance plan and cleared the PLAN item.
- Implemented dyn runtime helpers in TS/Go: `dyn.package`, `dyn.def_package`, and `dyn.Package.def` parse/evaluate dynamic code and replace prior definitions without overload merging.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_10_dynamic_metaprogramming_package_object bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_10_dynamic_metaprogramming_package_object$`.

# 2025-12-24 — Lexical line-join + trailing commas exec fixture (v11)
- Added `exec/06_09_lexical_trailing_commas_line_join` to cover delimiter line-joining and trailing commas in arrays/structs/imports; updated conformance plan and exec coverage index.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_09_lexical_trailing_commas_line_join bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_09_lexical_trailing_commas_line_join$`.

# 2025-12-24 — Array ops exec fixture + IndexMut error surfacing (v11)
- Added `exec/06_08_array_ops_mutability` to cover array mutation, bounds handling, and iteration, plus updated the conformance plan and coverage index.
- Index assignment now returns IndexError values from IndexMut implementations instead of silently discarding them (TS + Go interpreters).
- Tests: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_08_array_ops_mutability$`.

# 2025-12-24 — Exec fixtures for structs, unions, methods, interfaces, and packages (v11)
- Added exec fixtures for numeric literal contextual typing (plus overflow diag), positional structs, nullable truthiness, Option/Result construction, union guarded match coverage (plus a non-exhaustive diag), union payload patterns, method imports/UFCS instance-vs-static, interface dynamic dispatch, and directory-based package structure.
- Interpreters now coerce numeric values to float parameter contexts at runtime; TS/Go typecheckers accept integer literals in float contexts per spec.
- Updated `v11/fixtures/exec/coverage-index.json`, `v11/docs/conformance-plan.md`, and pruned completed PLAN items.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_05_03_struct_positional_named_tuple bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_01_union_payload_patterns bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_02_nullable_truthiness bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_03_union_construction_result_option bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_04_union_guarded_match_exhaustive bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_01_literals_numeric_contextual bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=09_02_methods_instance_vs_static bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_03_interface_type_dynamic_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_01_package_structure_modules bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(04_05_03_struct_positional_named_tuple|04_06_01_union_payload_patterns|04_06_02_nullable_truthiness|04_06_03_union_construction_result_option|04_06_04_union_guarded_match_exhaustive(_diag)?|06_01_literals_numeric_contextual(_diag)?|09_02_methods_instance_vs_static|10_03_interface_type_dynamic_dispatch|13_01_package_structure_modules)$'`.

# 2025-12-23 — Core exec fixtures + literal escape parsing (v11)
- Added exec fixtures for struct named updates (plus diagnostic), string/char literal escapes, control-flow expression values, and lambda closures with explicit return; updated exec coverage index + conformance plan and removed completed PLAN items.
- TS/Go parsers now unescape string/char literals with spec escapes (including `\'` and `\u{...}`) to align literal parsing across runtimes.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_05_02_struct_named_update_mutation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_01_literals_string_char_escape bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_05_control_flow_expr_value bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=07_02_lambdas_closures_capture bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(04_05_02_struct_named_update_mutation|04_05_02_struct_named_update_mutation_diag|06_01_literals_string_char_escape|06_05_control_flow_expr_value|07_02_lambdas_closures_capture)$'`.

# 2025-12-22 — Division-by-zero exec fixture (v11)
- Added `exec/04_02_primitives_truthiness_numeric_diag` to assert division-by-zero errors, plus inline semantics comment in the package-visibility fixture module.
- Normalized TS numeric division errors to use lowercase `division by zero` for parity with the stdlib error message and Go runtime.
- Updated exec coverage index + conformance plan to include the new diagnostic fixture.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-22 — Kernel alias normalization for typed patterns (v11)
- Normalized runtime type matching to map KernelChannel/KernelMutex/KernelRange/KernelRangeFactory/KernelRatio/KernelAwaitable/AwaitWaker/AwaitRegistration to their stdlib names so typed patterns match kernel aliases.
- Tests: `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod go test ./pkg/interpreter -run TestStdlibChannelMutexModuleLoader`; `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod ./run_all_tests.sh --version=v11`.

# 2025-12-22 — Singleton struct exec fixture (v11)
- Added `exec/04_05_01_struct_singleton_usage` covering singleton struct tags and pattern matching; updated exec coverage index and PLAN backlog.
- Treated singleton struct definitions as runtime values in TS/Go (typed pattern checks + type-name reporting) and matched singleton identifiers as constant patterns; Go `valuesEqual` now handles struct definition pointers.
- Tests: `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod ./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-20 — Reserved underscore type aliases (v11)
- Added `exec/04_04_reserved_underscore_types` for `_` placeholder type expressions and updated the exec coverage index + PLAN backlog.
- Added AST fixture `errors/type_alias_underscore_reserved` and enforced runtime/typechecker rejection of type alias name `_` in TS + Go.
- Refreshed the typecheck baseline to include the new alias diagnostic.
- Tests: `./run_all_tests.sh --version=v11 --fixture`; `ABLE_TYPECHECK_FIXTURES=warn bun run scripts/run-fixtures.ts -- --write-typecheck-baseline`.

# 2025-12-20 — Go type matching + constraint parity fixes (v11)
- Normalized runtime type matching for kernel alias names (`KernelArray`, `KernelHashMap`) and treated generic type args as wildcards during concrete matches to align alias recursion fixtures.
- Treated primitives as satisfying `Hash`/`Eq` method presence during impl resolution and expanded intrinsic Hash/Eq coverage to include all integer/float types.
- Stopped Go typechecker from typechecking impl/method bodies to mirror TS diagnostics for interface conformance tests.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestFixtureParityStringLiteral/strings/String_methods -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtures/04_07_05_alias_recursion_termination -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtures/06_01_literals_array_map_inference -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestFixtureParityStringLiteral/interfaces/apply_index_missing_impls -count=1`; `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-19 — Type expression syntax exec fixture (v11)
- Added `exec/04_03_type_expression_syntax` to cover nested type expressions with generic application and unions, including inline semantics comments.
- Updated `v11/fixtures/exec/coverage-index.json` and removed the completed PLAN backlog item.
- Tests not run (not requested).

# 2025-12-19 — Primitives truthiness exec fixture (v11)
- Added `exec/04_02_primitives_truthiness_numeric` to cover literal forms, Euclidean `//`/`%`, and truthiness for `nil`/`false`/`Error`/`void`.
- Updated `v11/fixtures/exec/coverage-index.json` and removed the completed PLAN backlog item.
- Tests not run (not requested).

# 2025-12-19 — Exec fixture reimplementation pass (v11)
- Reimplemented exec fixtures for §2, §3, §4.1, §4.7.2–§4.7.5, §5.0–§5.3, §6.1.7–§6.1.9, §6.2, and §6.3.1 with fresh programs aligned to spec semantics; refreshed manifests/expectations accordingly.
- Updated `PLAN.md` note to emphasize spec-first fixtures even when interpreters fail.
- Tests not run (exec fixtures pending re-sweep).

# 2025-12-19 — Exec fixture expansion (v11)
- Seeded exec fixtures for §2 (lexical comments/identifiers/trailing commas), §3 (block expression separation/value), and §4.1.4–§4.1.6 (generic inference with constrained interface methods), each with inline semantics notes.
- Coverage index and PLAN backlog updated to reflect the new fixtures; exec coverage guard stays green. Go runtime updated to treat generic-parameter types as wildcards during overload selection so the new §4.1 fixture runs.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-18 — Exec fixtures renamed to spec sections (v11)
- Renamed exec fixtures/packages to use spec section prefixes, refreshed manifests/comments, and regenerated the coverage index + docs with the spec-based naming scheme.
- Added/updated exec fixtures for async proc/spawn scheduling, methods/UFCS, option/result handling via raise/rescue, package visibility with explicit aliases, and alias/union generic combos; each case documents the exercised semantics inline.
- Conformance docs now carry the seeded coverage matrix and exec fixtures guide keyed to the new IDs, with the JSON coverage index tracked alongside.
- `run_all_tests.sh` runs `scripts/check-exec-coverage.mjs` before the suites; `./run_all_tests.sh --version=v11` passes after the exec fixture sweep.
- Seeded new exec fixtures for §4.7.2–§4.7.5 (generic alias substitution, import visibility, alias-backed methods/impls, and recursive alias termination via nominal indirection) with manifests and package wiring.
- Coverage index and PLAN backlog updated to reflect the new alias fixtures.
- Added exec fixtures for §5.0–§5.2 (mutability declaration vs assignment, identifier/wildcard typed patterns, struct pattern rename with typed nesting) and brought coverage index + PLAN in sync.

# 2025-12-13 — Host interop tutorial unblocked + prefix match guard (v11)
- Added the Go-side `read_text` extern for tutorial 14 (imports `os`, returns `host_error` on failure) so the host interop example now runs on the Go runtime alongside the existing TS path.
- Tree-sitter grammar usage is now enforced to reject prefix-style `match <expr> { ... }`: added Go parser and TS tree-sitter/mapper tests that expect syntax errors for the legacy ordering, matching the spec requirement of `<expr> match { ... }`.
- Tests: `cd v11/interpreters/go && go test ./pkg/parser -run PrefixMatch`; `cd v11/interpreters/ts && bun test test/parser/fixtures_parser.test.ts`.

# 2025-12-12 — Inherent methods as functions (v11)
- Exported method functions now carry their method-set obligations: Go binds implicit `self` for method shorthand, substitutes the receiver into exported signatures, and preserves method-set context so free-call constraints fail when receivers/where-clauses are violated.
- TypeScript attaches method-set obligations/substitutions to exported method infos so direct calls enforce receiver typing and block missing `where` constraints; §9 in the spec now spells out the export + sugar model for inherent methods.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`.

# 2025-12-11 — UFCS overload priority (v11)
- Added overload priority metadata so inherent methods outrank interface/default impls without masking UFCS ambiguities; runtime dispatch now tags impl/default entries with lower priority and tolerates impl ambiguity when other candidates are present.
- Typechecker mirrors the priority model, keeping the highest-priority candidate per signature and sorting overloads by score then priority to stay aligned with runtime selection.
- Go interpreter/runtime carry the same priority metadata for parity with the TypeScript path.
- Module loaders now auto-discover bundled kernel/stdlib roots (no `ABLE_STD_LIB` knob); the `able` namespace is treated as a normal dependency resolved through the standard search paths/lockfile.
- Tests: `bun test test/stdlib/string_stdlib_integration.test.ts`; `bun test test/stdlib/concurrency_stdlib_integration.test.ts test/stdlib/await_stdlib_integration.test.ts test/stdlib/hash_set_stdlib_integration.test.ts test/stdlib/bit_set_stdlib_integration.test.ts test/stdlib/hash_map_stdlib_integration.test.ts`; `./run_all_tests.sh --version=v11`.

# 2025-12-10 — Conditional/unwrap syntax shift (v11)
- Landed the new `if/elsif/else` syntax and `{ expr or { err => ... } }` handlers across grammar/AST/TS+Go interpreters/typecheckers; regenerated tree-sitter artifacts and aligned AST fixtures/printers with the new `elseIfClauses`/`elseBody` layout.
- Updated fixtures, examples, tutorials, and tests (TS + Go) to drop the legacy `if/or` + `| err |` handler forms; handling blocks now use `binding =>` with optional binding, and parser mappers no longer misclassify handler statements as bindings.
- Exported the fixture corpus and reran the full parity + CLI + Go/TS suites to green.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-09 — Pipe/placeholder alignment (v11)
- Made bare `@` placeholders behave as `@1` across parsers/interpreters/typecheckers; placeholder lambdas now reuse the first argument when unnumbered tokens repeat, with new runtime tests in TS/Go to lock the behaviour.
- Pipe expressions now typecheck as callable invocations (including low-precedence `|>>`), emitting non-callable diagnostics instead of silently returning `unknown`, and placeholder-driven pipes follow the callable-only model.
- Exported the AST fixture corpus after the placeholder changes so shared fixtures/exports include explicit placeholder indices.
- Tests: `cd v11/interpreters/ts && bun test test/runtime/pipes.test.ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run Placeholder && go test ./pkg/typechecker`.

## 2025-12-08 — Parity harness green (v11)
- Normalized Go fixture results (String kind) and fixed native Error methods to use receiver-bound arity so messages propagate correctly through fixture CLI output and parity reporting.
- Cleaned up TS fixture/parity scripts (type checks and stringify usage) and aligned stringification on both runtimes: TS honours `to_string`, Go bound-method stringification now includes targets, and stdout capture no longer misreports types.
- Parity sweep now fully green with the latest fixture export; report saved to `tmp/parity-report.json`.
- Updated the v11 spec + manuals/onboarding to canonicalize the `String` type (lowercase reserved for scalar primitives) and remove the remaining lowercase `string` references in type signatures/examples.
- Tests: `cd v11/interpreters/go && go test ./...`; `cd v11/interpreters/ts && bun run scripts/run-parity.ts`.

## 2025-12-07 — Operators tutorial + xor parity (v11)
- Added tutorial `02a_operators_and_builtin_types.able` to showcase built-in scalars, arithmetic (`/ // %% /%`), comparisons, bitwise/shift ops, and boolean logic; confirmed it runs in TS+Go.
- Enabled `\xor` in the Go interpreter/typechecker and added a shared fixture (`functions/bitwise_xor_operator`) so bitwise xor stays covered across runtimes.
- Go/TS suites remain green after exporting fixtures and rerunning the operators tutorial with both CLIs.

## 2025-12-06 — Option/Result `else` handling fixed (v11)
- Implemented spec-compliant `else {}` handling for `?T`/`!T` across TS/Go interpreters and typecheckers: failures now trigger handler blocks on `nil` or `Error` values, with error bindings and optional early-return narrowing.
- Updated the Go runtime type test so interface-backed errors (e.g., `MathError`) satisfy `Error`, letting `or-else` handlers bind user-defined errors instead of treating them as successes.
- The tutorial 09 example now runs to completion in both interpreters (`Handled error: need two numbers`), clearing the option-narrowing runtime/typechecker regression from the PLAN backlog.

## 2025-12-05 — Go stdlib typecheck parity (v11)
- Go ProgramChecker sweep (driver.NewLoader RootStdlib against `v11/stdlib/src`) now reports zero diagnostics; stdlib modules typecheck cleanly with the Go checker.
- `./run_all_tests.sh --version=v11` stays green (TS + Go units, fixtures, parity, CLI); Go unit suites no longer flag await manual waker.
- Removed the Go stdlib/typechecker parity worklist from PLAN since the stdlib checker is stable again.

## 2025-12-05 — UFCS method-style fallback finalized (v11)
- Clarified UFCS spec wording (§7.4/§9.4) covering pipe equivalence, receiver-compatible overload selection, and self-parameter trimming; marked the TODO as complete and removed the PLAN item.
- Added shared UFCS fixtures: `functions/ufcs_generic_overloads` (generic + overloaded free-function binding parity across pipe/member syntax) and `errors/ufcs_overload_ambiguity` (method-style ambiguity diagnostics). Exporter now emits these fixtures.
- Hardened the TS typechecker runtime: pipelines no longer emit spurious undefined-identifier diagnostics, and duplicate typecheck diagnostics are deduped in the fixture runner; fixed a missing runtime import in pattern handling.
- Updated the typecheck baseline to reflect new fixtures and existing diagnostics, keeping UFCS coverage aligned across interpreters.

## 2025-12-04 — Struct pattern shorthand + rename operator (v11)
- Codified `::` as the pattern rename operator in §5.2 (shorthand `field` = `field::field`, rename, rename+type, nested patterns) while keeping dot for namespace traversal; added examples for assignment/match and aligned import syntax guidance.
- Documented import renaming via `::` (`import pkg::alias`, `import pkg.{item::alias}`) in the spec and queued implementation work in PLAN.
- Updated the tree-sitter grammar and TS/Go ASTs to carry rename + type metadata, track struct kinds for positional patterns, and propagate type annotations into nested struct patterns when present.
- TS/Go runtimes and typecheckers now evaluate/check shorthand/rename/type struct patterns consistently across `=`/`:=`/`match`/`rescue`; fixtures/tutorials/exporter outputs refreshed (including `patterns/struct_pattern_rename` and updated baselines).
- Tests: `bun test` (v11/interpreters/ts); `cd v11/interpreters/go && go test ./...`.

## 2025-12-03 — Impl specificity + ambiguity parity (v11)
- Implemented the full impl-specificity lattice across Go/TS typecheckers and runtimes: concrete > generic, constraint superset, union subset, and more-instantiated generic tie-breaks with ambiguity diagnostics; named impls stay opt-in.
- Go/TS runtimes now register generic-target impls, bind `Self` during matching, and surface consistent ambiguity errors; fixtures/manifests (`interfaces/impl_specificity*`) and the typecheck baseline reflect the new coverage.
- Tests: `cd v11/interpreters/go && go test ./...`; `cd v11/interpreters/ts && bun test`.

## 2025-12-02 — Pipe precedence + low-precedence pipe (v11)
- Raised pipe (`|>`) precedence above assignment and added a low-precedence `|>>` operator in the grammar (tree-sitter regenerated) with matching TS/Go parser mappings.
- TS/Go interpreters and typecheckers now treat `|>>` identically to `|>` (topic/callable fallback, placeholder guards) so pipeline semantics stay in sync across runtimes.
- Added fixture `pipes/low_precedence_pipe` to cover assignment vs pipe grouping and `||` interactions; refreshed exported fixtures, and `bun test` (v11/interpreters/ts) plus `cd v11/interpreters/go && go test ./...` are green.

## 2025-12-02 — UFCS inherent instance methods (v11)
- UFCS resolution now considers inherent instance methods (excluding static/interface/named impl methods) when the first argument can serve as the receiver; TS/Go interpreters bind the receiver accordingly and fall back from identifier calls, with TS/Go typecheckers mirroring the UFCS candidate search and call handling.
- Added fixtures for UFCS calls into inherent methods plus a static-method negative case, refreshed the typecheck baseline, and documented the UFCS expansion in the spec (§7.4/§9.4).
- `bun test` (v11/interpreters/ts) and `cd v11/interpreters/go && go test ./...` remain green after the UFCS sweep.

## 2025-12-02 — Function/method overloading implemented (v11)
- Added function/method overload sets across the Go and TS interpreters (runtime dispatch with nullable-tail omission, bound/native/UFCS/dyn refs) and TS typechecker (arity filter, specificity scoring, ambiguity diagnostics); envs now merge duplicate names into overload sets.
- Updated the spec (§7.4.1, §7.7) to codify nullable trailing parameter omission and overload eligibility/specificity, removed the overloading TODO, and exported shared fixtures for overload success/ambiguity (functions and methods).
- `bun test` (v11/interpreters/ts) and `cd v11/interpreters/go && go test ./...` are green after the overload sweep.

## 2025-12-01 — Interface dispatch alignment + Awaitable stdlib
- Codified the language-backed interfaces in the stdlib (Apply, Index/IndexMut, Iterable defaults, Awaitable/Proc/Future handles, channel/mutex awaitables) and wired both interpreters/typecheckers to route callable invocation and `[]`/`[]=` through these impls, surfacing Apply/IndexMut diagnostics when missing.
- Added shared fixtures/tests for callable values and index assignment dispatch (`interfaces/apply_index_dispatch`, missing-impl diagnostics, Go/TS unit tests) plus new awaitable fixtures covering channel/mutex/timer arms and stdlib helpers (`concurrency/await_*`, ModuleLoader integration for Channel/Mutex/Await.default).
- Refreshed the stdlib concurrency surface (awaitable wrappers, channel/mutex helpers) and kept parity harnesses green; `./run_all_tests.sh --version=v11` now passes end-to-end after the interface alignment sweep (TS unit/CLI/tests+fixtures+parity, Go unit tests).

## 2025-11-27 — Channel for-loop iteration fixed (Go runtime)
- Channel iterators now terminate correctly after closure/empty reads (stdlib `ChannelIterator.next` returns `IteratorEnd` before the typed arm), so Go for-loops over channels no longer hang.
- The Go ModuleLoader concurrency smoke test now iterates a channel with a for-loop (summing values) via the stdlib surface, and the stdlib channel/mutex smoke test mirrors the for-loop path.
- `go test ./...` (v11/interpreters/go) and the TS stdlib integration test for concurrency remain green.

## 2025-11-26 — Kernel vs stdlib layering complete (v11)
- Primitive strings now live entirely in the stdlib: helpers/iterators/wrap-unwrap operate on built-in strings, with only the three kernel bridges left native. Go/TS runtimes and typecheckers resolve string members via the stdlib and surface import hints when missing.
- Arrays now prefer stdlib methods end-to-end (size inline, minimal native shims retained for compatibility), with runtimes/typecheckers pointing diagnostics at `able.collections.array`.
- Go runtime member access now consults stdlib method sets before native fallbacks, and primitive `string` is treated as iterable (`u8`) in the Go typechecker, matching the TS path. ModuleLoader + Go test suites stay green.

## 2025-11-20 — Typed-pattern match reachability fix
- Added a reachability guard for typed `match` patterns so clauses whose annotations cannot match the subject still typecheck but are treated as unreachable, preventing spurious return/branch diagnostics in the Go checker (mirrors TS intent).
- Full v11 suites run green after the fix (`go test ./pkg/typechecker`, `go test ./...`, `./run_all_tests.sh --version=v11`).
- Extended the v11 stdlib with a `collections.set` interface and `HashSet` implementation (plus smoke + TS ModuleLoader integration tests) and refreshed stdlib docs/PLAN so the restored surface tracks the new module.
- Go typechecker now recognises stdlib collections (`List`, `Vector`, `HashSet`) as valid for-loop iterables; iterable helpers moved out of `type_utils.go` to stay under the 1k-line guardrail, and a new regression test covers generic element inference for these types.

## 2025-11-19 — File modularization cleanup
- Split Go typechecker declaration/type utility stacks into dedicated files (`decls_*`, `type_substitution.go`) and shrank `type_utils.go` beneath the 1k-line guardrail; Go AST definitions now live across `ast.go`, `type_expressions.go`, and `patterns.go` so each file stays lean.
- Broke the long TS fixture exporter (`proc_scheduling.ts`) into `proc_scheduling_part{1,2}.ts` with a tiny aggregate shim to keep per-file size under 1000 lines.
- `go test ./...` (v11/interpreters/go) remains green after the split.

## 2025-11-18 — String helper surface complete
- Added the full string helper set to both interpreters: `len_chars`/`len_graphemes`, `substring` (code-point offsets with `RangeError` on invalid bounds), `split` (empty delimiter splits graphemes), `replace`, `starts_with`/`ends_with`, and `chars`/`graphemes` iterators (Segmenter-aware in TS, rune fallback in Go). Shared helpers keep `len_*` in sync with iterator `size()`.
- Typecheckers now understand the string surface (Go member access signatures, TS call resolution for optional substring length and string primitives), and Go array literal inference merges element types instead of rejecting unions.
- Added the `strings/string_methods` AST fixture covering length/slicing/split/replace/prefix/suffix cases; `./run_all_tests.sh --version=v11` runs green after the fixture/typechecker/runtime updates.

## 2025-11-17 — Loop/range constructs & continue semantics complete
- Locked down the full §8.2–§8.3 feature set across both runtimes: `loop {}` expressions, `while`, and `for` now return the last `break` payload (or `nil` on exhaustion) while rejecting labeled `break`/`continue` targets. The Go and TypeScript interpreters share identical behavior, including iterator-driven loops and generator continuations.
- Added the Range interface runtime registries (`src/interpreter/range.ts`, `pkg/interpreter/range_runtime.go`) so range literals first delegate to stdlib implementations before falling back to synthesized arrays, matching the spec requirement that ranges materialize via the Range interface.
- Tree-sitter mappers + AST builders already exposed `LoopExpression`, `ContinueStatement`, and `RangeExpression`; we confirmed parser precedence and metadata are round-tripping, so the AST contract is now exercised end-to-end by fixtures such as `control/loop_expression` and `control/for_range_break`.
- Typechecker metadata is in place: both checkers push loop contexts, record break payload types, and emit diagnostics for labeled `continue`, so assignments using loop expressions inherit the correct type.
- Fresh runtime + parity coverage landed in TS (`test/control_flow/{while,for}.test.ts`, `test/runtime/iterators.test.ts`) and Go (`pkg/interpreter/interpreter_control_flow_test.go`). `./run_all_tests.sh --version=v11` runs green (TS unit + CLI suites, fixtures/parity harness, Go parser/interpreter/typechecker tests), so PLAN item 6 is officially complete and moved here from the TODO list.

# 2025-11-17 — Typed declaration + literal adoption work finalized
- Locked down the binding semantics for `:=`/`=` by keeping typed patterns intact across AST parsing/mapping (declaration + fallback assignment), enforcing “`:=` introduces at least one new binding” in both interpreters/typecheckers, and ensuring runtime evaluation order stays deterministic (RHS once, receivers/indexers evaluated exactly once) even for compound assignments and safe-navigation forms.
- Verified the runtime/typechecker literal-adoption flow now covers every context listed in the v11 spec (arrays, maps, ranges, iterator yields, async bodies, function bodies/returns/arguments, struct literals, typed patterns), and refreshed the AST fixtures (`patterns/typed_destructuring`, `patterns/typed_equals_assignment`) so typed destructuring + typed `=` assignments keep parity between TypeScript and Go.
- Fixed a regression uncovered by the fixture sweep where the TS AST builder always serialized `isSafe: false` on member-access nodes; it now omits the flag unless `?.` is present, matching the Go AST (`json:"safe,omitempty"`) and letting fixtures/tests pass again. Full `./run_all_tests.sh --version=v11` run is green.

## 2025-11-16 — Safe navigation operator implemented
- Tree-sitter grammar/parser now treat `?.` as part of member-access, the TypeScript/Go AST mappers expose a `safe` flag on `MemberAccessExpression` nodes, and the generated parser artifacts (`grammar.json`, `parser.c`, node types, WASM) have been regenerated so fixtures and tooling pick up the new operator.
- TypeScript + Go interpreters short-circuit safe member access/calls (returning `nil` when receivers are `nil`, skipping argument evaluation, and mirroring dot semantics otherwise) while rejecting assignments that attempt to use `?.`.
- The Go typechecker wraps safe-navigation results in `NullableType` only when the receiver may be `nil`, so redundant usage on non-optional receivers still typechecks as plain dot access. New unit tests in both runtimes (`test/runtime/safe_navigation.test.ts`, `interpreter_safe_navigation_test.go`) cover the runtime semantics, and `bun test …`, `go test ./pkg/typechecker ./pkg/interpreter`, plus `./v11/ablego ./v11/examples/rosettacode/factorial.able` all remain green.

## 2025-11-15 — Map literal support landed
- Extended the shared tree-sitter grammar plus both AST layers (TS + Go) with `MapLiteral`/entry/spread nodes, regenerated parser artifacts, and wired the fixtures/exporter so `#{ ... }` forms round-trip through `module.json` + source generation.
- Implemented map literal evaluation in the TypeScript interpreter (new hash_map value kind, literal/spread semantics, fixture assertions) and added runtime+typechecker unit tests covering spreads, duplicates, and diagnostics.
- Mirrored the same behavior in the Go interpreter/typechecker (hash map insertion helper, literal evaluation, new `MapType`, diagnostics) and added Go unit tests—`./v11/run_all_tests.sh` now exercises the new fixtures end-to-end.

## 2025-11-14 — Type alias declarations wired end-to-end
- The shared tree-sitter grammar gained a `type_alias_definition` rule (space-delimited generics + optional `where` clauses) plus corpus coverage, and both the TypeScript + Go AST mappers now surface `TypeAliasDefinition` statements.
- TypeScript + Go parsers/interpreters ignore alias declarations at runtime while the TypeScript typechecker/summary plumbing keeps tracking them; the TypeScript fixture exporter/pretty-printer now emits `type Foo T where … = Expr` syntax.
- Added the `types/type_alias_definition` AST fixture so both runtimes/typecheckers exercise alias declarations, updated the fixture baseline (`bun run scripts/export-fixtures.ts && ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts -- --write-typecheck-baseline`), and kept Go green with `go test ./pkg/parser ./pkg/interpreter`.

## 2025-11-13 — v11 Spec Expansion Complete
- Closed the v11 spec TODO slate: every deferred item now lives in `spec/full_spec_v11.md`, covering mutable `=` semantics (§5.3.1), map literals (§6.1.9), struct functional updates (§4.5.2), type aliases (§4.7), safe navigation (§6.3.4), typed `=` declarations (§5.1–§5.1.1), contextual literal typing (§6.1.1/§6.3.2), optional generic parameters (§7.1.5/§4.5/§10.1), await/async coordination plus channel error surfaces (§12.6–§12.7), the `loop` expression (§8.2.3), Array/String APIs (§6.8/§6.1.5/§6.12.1), stdlib packaging + module search paths (§13.6–§13.7), and the regex/text modules (§14.2). `spec/TODO_v11.md` now reflects that completion and will track only newly scheduled language work.
- The root PLAN no longer lists “Expand the v11 specification” as an open milestone; remaining TODOs focus on implementing the documented features across the interpreters, parser, and stdlib.

## 2025-11-12 — Versioned Workspace Split
- Introduced a dedicated `v10/` workspace that now owns the frozen Able v10 assets (`design/`, `docs/`, `examples/`, `fixtures/`, `parser/`, `stdlib/`, `interpreter10` → `v10/interpreters/ts`, `interpreter10-go` → `v10/interpreters/go`, plus helper scripts). This removes ambiguity about where new work should land and keeps the archived toolchain intact for maintenance.
- Added a version-dispatching `run_all_tests.sh` at the repo root (`./run_all_tests.sh --version=v10|v11 --typecheck-fixtures=...`) so CI and contributors can target either workspace without remembering individual paths.
- Updated repo-wide onboarding docs (`README.md`, `AGENTS.md`, `PLAN.md`) to describe the multi-version layout, explain how to run tests per version, and drop completed roadmap items covering the initial workspace bootstrap/freeze.
- Copied the legacy v10 docs (`v10/README.md`, `v10/AGENTS.md`, `v10/PLAN.md`, `v10/LOG.md`) so historical context remains close to the frozen code while the root docs focus on cross-version coordination.

## 2025-11-11 — Stdlib Module Search Paths
- **Pipe semantics parity**: Added the `pipes/multi_stage_chain` AST fixture so multi-stage pipelines that mix `%` topic steps, placeholder-built callables, and bound methods stay covered; `bun run scripts/run-fixtures.ts` (TypeScript) and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` (Go) stay green with no parity divergences observed.
- **Typechecker strict fixtures**: TypeScript’s checker now hoists struct identifiers for static calls, predeclares `:=` bindings (so proc handles can reference themselves), binds iterator driver aliases plus struct/array pattern destructures, and hides private package members behind the standard “has no symbol” diagnostic. The full fixture suite now passes under `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, and manifests/baselines were updated where diagnostics are expected.
- **Dynamic interface collections & iterables**: added the shared fixture `interfaces/dynamic_interface_collections` (plus exporter wiring) so both interpreters prove that range-driven loops and map-like containers storing interface values still choose the most specific impl even when union targets overlap. `bun run scripts/run-fixtures.ts` is green; Go parity remains blocked by existing fixture failures (`functions/hkt_interface_impl_ok`, `imports/static_alias_private_error`).
- **Privacy & import spec gaps**: enforced package privacy across both interpreters/typecheckers (selectors, aliases, wildcard bindings), tightened dynimport placeholder handling (TypeScript + Go parity plus fixture coverage), and updated `spec/full_spec_v10.md` with the canonical `Proc`/`Future` handle definitions so runtimes and tooling share the same ABI (`ProcStatus`, `ProcError`, `status/value/cancel` contracts).
- **Interface & impl completeness**: higher-kinded/visibility edge cases now have end-to-end coverage. Overlapping impl ambiguity diagnostics land in both interpreters/typecheckers, TypeScript enforces method-set backed constraints with parity coverage, both interpreters honour interface self-type patterns (including higher-kinded targets) and reject bare constructors unless `for …` is declared, and parser/exporter propagate the `for …` patterns so fixtures surface the diagnostics automatically. Shared AST fixtures (`errors/interface_self_pattern_mismatch`, `errors/interface_hkt_constructor_mismatch`, plus the positive `functions/hkt_interface_impl_ok`) keep both runtimes + typecheckers aligned.

- Proc/future cancellation coverage is complete: both interpreters expose `Future.cancel()`, cancellation transitions produce the `Cancelled` status, and the new `concurrency/future_cancel_nested` fixture exercises a proc awaiting a spawned future that is cancelled mid-flight. The exporter and strict harness generate the fixture automatically, keeping the TypeScript + Go runtimes/typecheckers in lockstep for nested cancellation chains across both executors.
- Channel error helpers now emit the stdlib error structs in both runtimes. TypeScript + Go runtimes route every channel error through `ChannelClosed`, `ChannelNil`, or `ChannelSendOnClosed`, and new Bun/Go tests cover the behaviour (`interpreter10/test/concurrency/channel_mutex.test.ts`, `interpreter10-go/pkg/interpreter/interpreter_channels_mutex_test.go`).
- Lightweight executor diagnostics are now wired into both runtimes: the new `proc_pending_tasks()` helper surfaces cooperative queue length via `CooperativeExecutor.pendingTasks`, Go’s serial and goroutine executors expose the same data, and coverage comes from a dedicated AST fixture (`concurrency/proc_executor_diagnostics`) plus Bun/Go unit tests. The spec, manuals, and `design/concurrency-executor-contract.md` now describe the helper so fixtures can assert drain behaviour deterministically.
- String host bridge externs (`__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint`) are wired into both interpreters and typecheckers with dedicated tests (`interpreter10/test/string_host.test.ts`, `interpreter10-go/pkg/interpreter/interpreter_string_host_test.go`).
- Hasher externs (`__able_hasher_create/write/finish`) now back the stdlib hash maps across TypeScript and Go, complete with parity tests and stub support for tooling.
- Added the `concurrency/channel_error_rescue` AST fixture and exposed `Error` member access (message/value) in both interpreters so Able code can assert the struct payloads produced by the channel helpers.
- Go parity now runs the new `concurrency/channel_error_rescue` fixture under the goroutine executor, and a dedicated Go test (`TestChannelErrorRescueExposesStructValue`) verifies that rescuing channel errors exposes the struct payload via `err.value`.
- Added the `errors/result_error_accessors` AST fixture so both interpreters exercise `err.message()/cause()/value` inside `!T else { |err| ... }` flows; fixture exporter + TS harness updated accordingly.
- Go typechecker now recognises `Error.message()`, `.cause()`, and `.value`, and the spec documents the runtime-provided `Error.value` payload hook; the typechecker baseline entry for `channel_error_rescue` was removed once diagnostics cleared.
- Proc/future runtime errors now record their cause payloads in both interpreters, the new `concurrency/proc_error_cause` fixture exercises `err.cause()` end-to-end, and matching Bun/Go tests keep the regression harness green.
- Generator laziness parity closed: iterator continuations now cover if/while/for/match across both runtimes, stdlib helpers (`stdlib/src/concurrency/channel.able`, `stdlib/src/collections/range.able`) use generator literals, and new fixtures (`fixtures/ast/control/iterator_*`, `fixtures/ast/stdlib/channel_iterator`, `fixtures/ast/stdlib/range_iterator`) keep the shared harness authoritative.
- Automatic time slicing verified for long-running procs: the new `concurrency/proc_time_slicing` fixture proves that handles without explicit `proc_yield()` still progress under repeated `proc_flush()` calls, capturing both the intermediate `Pending` status and the eventual resolved value across runtimes.
### AST → Parser → Typechecker Completion Plan _(reopen when new AST work appears)_
- Full sweep completed 2025-11-06 (strict fixture run, Go interpreter suite, Go parser harness, and `bun test` all green). Archive details in `LOG.md`; bring this plan back only if new AST/syntax changes introduce regressions.

## 2025-11-09 — Executor Diagnostics Helper
- Added the `proc_pending_tasks()` runtime helper so Able programs/tests can observe cooperative executor queue depth. TypeScript wires it through `CooperativeExecutor.pendingTasks()` while the Go runtime surfaces counts from both the serial and goroutine executors (best-effort on the latter via atomic counters). The helper is registered with both typecheckers so Warn/Strict fixture runs understand the signature.
- New coverage keeps the helper honest: Bun unit tests exercise the helper directly (`interpreter10/test/concurrency/proc_spawn_scheduling.test.ts`), Go gains matching tests (`TestProcPendingTasksSerialExecutor`, `TestProcPendingTasksGoroutineExecutor`), and a serial-only AST fixture (`fixtures/ast/concurrency/proc_executor_diagnostics/`) ensures both interpreters prove that `proc_flush` drains the cooperative queue.
- Spec + docs now describe the helper alongside `proc_yield`/`proc_flush` (see `spec/full_spec_v10.md`, `docs/manual/*.md/html`, `design/concurrency-executor-contract.md`, and `AGENTS.md`), and the PLAN TODO for “Concurrency ergonomics” is officially closed out.

## 2025-11-08 — Fixture Diagnostics Parity Enforcement
- `interpreter10/scripts/run-parity.ts` now diffs typechecker diagnostics for every AST fixture so warn/strict parity runs catch unexpected checker output even when manifests do not declare expectations. The parity JSON report captures the mismatched diagnostics to speed up triage.
- Added `test/scripts/parity/fixtures_compare.test.ts` to cover the helper logic that determines when diagnostics mismatches should fail parity.
- Go’s typechecker now treats unannotated `self` parameters inside `methods {}` / `impl` blocks as `Self`/concrete receiver types and seeds iterator literals with the implicit `gen` binding, eliminating the extra diagnostics that used to appear only in Go’s warn/strict runs.
- Added Go regression tests (`pkg/typechecker/checker_impls_test.go`, `pkg/typechecker/checker_iterators_test.go`) so implicit `self` bindings and iterator generator helpers stay covered.
- Bun’s typechecker mirrors the same behaviour: iterator literals now predefine the implicit `gen` binding, implicit `self` parameters default to `Self`, and new tests (`test/typechecker/method_sets.test.ts`, `test/typechecker/iterators.test.ts`) keep the coverage locked in.
- The Go fixture runner and CLI now continue evaluating programs in warn/strict modes even when diagnostics are reported (matching the Bun harness), so parity checks capture both the expected runtime results and the shared diagnostics payloads.

## 2025-11-07 — AST → Parser → Typechecker Cycle Revalidated
- Added proc handle memoization fixtures (success + cancellation) and ensured both interpreters plus the Go parser harness run them under strict typechecking (`bun run scripts/run-fixtures.ts`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`).
- Verified the full suite remains green (`./run_all_tests.sh --typecheck-fixtures=strict`, `bun test`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`), keeping the Priority 0 gate satisfied.
- Updated PLAN.md to mark the current AST → Parser → Typechecker cycle complete and advance the focus to Phase α (Channel & Mutex runtime bring-up).
- Added stdlib specs for channel/mutex behaviour (`stdlib/tests/concurrency/channel_mutex.test.able`) so the Phase α bullet “add unit tests covering core operations” is now satisfied.

## 2025-11-07 — Serial Executor Future Reentrancy
- Go’s SerialExecutor now exposes a `Drive` helper that runs pending proc/future tasks inline, so nested `future.value()` / `proc_handle.value()` calls no longer deadlock and match the TypeScript scheduler semantics. The helper steals the targeted handle from the deterministic queue, executes it re-entrantly (including repeated slices when `proc_yield` fires), and restores the outer task context once the awaited handle resolves.
- Fixtures `concurrency/future_value_reentrancy` and `concurrency/proc_flush_fairness` now pass under the Go interpreter’s serial executor, keeping the newly added fairness/re-entrancy corpus green for both runtimes (`ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`).
- PLAN immediate actions trimmed: the “Concurrency stress coverage & docs” placeholder has been cleared now that the blocking fixtures run cleanly; follow-up concurrency work can graduate to design docs instead of the top-level plan.
- Added a Go-side regression test (`TestSerialExecutorFutureValueReentrancy`) that mirrors the new fixture to ensure nested `future.value()` calls stay green even if future contributors touch the executor; the design note `design/go-concurrency-scheduler.md` now documents the inline-driving behaviour and follow-up doc work.
- Added a companion fixture (`concurrency/proc_value_reentrancy`) plus TypeScript exporter wiring + Go regression test (`TestSerialExecutorProcValueReentrancy`) so both interpreters exercise nested `proc.value()` waits under the serial executor; parser coverage tables were updated accordingly.
- Documented the goroutine-executor fairness contract (what `proc_yield`/`proc_flush` mean under `GoroutineExecutor`, how we rely on Go’s scheduler, and when tests must fall back to the serial executor) to close out the remaining PLAN follow-up (`design/go-concurrency-scheduler.md`).
- Updated the v10 spec to codify re-entrant `proc.value()` / `future.value()` semantics so both interpreters (and future targets) must guarantee deadlock-free nested waits (§12.2.5 “Re-entrant waits”).
- Added direct Go unit tests (`TestProcHandleValueMemoizesResult`, `TestProcHandleValueCancellationMemoized`) to ensure repeated `value()` calls return memoized results/errors even after cancellation, satisfying the remaining “exercise repeated value() paths” item from the concurrency plan.
- Introduced the `concurrency/proc_value_memoization` fixture (plus exporter wiring) so both interpreters prove proc handles memoize values, and updated the Go parity harness to run it with the goroutine executor.
- Added `concurrency/proc_value_cancel_memoization` to assert that cancelled proc handles return the same error for repeated `value()` calls without re-running their bodies; the fixture exporter, AST corpus, and fixture run all cover this scenario now.

## 2025-11-06 — Tree-Sitter Mapper Modularization Complete
- Go declarations/patterns/imports all run through the shared `parseContext`, removing the last raw-source helpers so both runtimes share one parser contract (`interpreter10-go/pkg/parser/{declarations,patterns,statements,expressions}_parser*.go`).
- TypeScript parser README now calls out the shared context, and `PLAN.md` logged the Step 6 regression sweep so future contributors know the refactor is locked in (`interpreter10/src/parser/README.md`, `PLAN.md`).
- Wrapper exports like `parseExpression(node, source)` / `parseBlock` / `parsePattern` were removed; all Go parser consumers now flow through the context pipeline.
- Tests: `./run_all_tests.sh --typecheck-fixtures=warn` and `cd interpreter10-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`.
- Follow-up: confirm any remaining helpers that still accept raw source (e.g., host-target parsing) genuinely require it before migrating them to `parseContext`.

## 2025-11-03 — AST → Parser → Typechecker Completion Sweep
Status: ✅ Completed. Parser coverage table reads fully `Done`, TypeScript + Go checkers share the bool/comparison semantics, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` executes cleanly after wiring the remaining builtin signatures, iterator annotations, and error fixture diagnostics into the manifests.

Open items (2025-11-02 audit):
- [x] Iterator literals dropped both the optional binding identifier and optional element type annotation. Update both ASTs + parsers/interpreters so the metadata survives round-trips and execution.
- [x] Teach both typecheckers to honor the iterator element annotation so every `yield` matches the declared element type (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/literals.go` + iterator tests).
- [x] Carry iterator element metadata through `for` loops so typed loop patterns validate across array/range/iterator inputs (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/{control_flow,literals,type_utils}.go` + new cross-loop tests).
- [x] Give the TS checker parity with Go for block/proc/spawn typing and array/range literal inference so all three stages (AST → parser → typechecker) agree on the element/result metadata (§6.8, §6.10, §12.2).
- [x] Enforce async-only builtins (`proc_yield`) and add concurrency smoke tests so TS emits the same diagnostics as Go when authors call scheduler helpers outside `proc`/`spawn`.
- [x] Implement `if`/`while` diagnostics + inference in the TS checker so control-flow expressions match the Go implementation (§8.1/§8.2).
- [x] Mirror Go's match/rescue guard enforcement in the TS checker (§8.1.2 / §11.3).
- [x] Enforce package privacy + import diagnostics in the TS checker so private packages/definitions behave identically to Go (updated `imports.test.ts` + package summaries carry `visibility` metadata).

## Historical Status Notes

### 2025-10-30
- Comments are now ignored during parser → AST mapping for both interpreters.
  - ✅ Go: `ModuleParser` / helper utilities skip `comment`, `line_comment`, `block_comment` nodes and `TestParseModuleIgnoresComments` asserts the behaviour.
  - ✅ TypeScript: `tree-sitter-mapper` filters the same node types; `fixtures_mapper.test.ts` covers the mapping path and `fixtures_parser.test.ts` ensures the raw grammar parses comment-heavy sources.
- TODO: audit remaining parser/mapping gaps per `design/parser-ast-coverage.md` (pipes/topic combos, functional updates, etc.) and backfill fixtures/tests.
- DONE: comment skipping now wired through struct literals, struct patterns, and related mapper helpers across both runtimes.
- TODO: Build end-to-end coverage across **all three facets** (parsing, tree → AST mapping, AST evaluation) for both interpreters. Use the coverage table to drive fixture additions, parser assertions, and runtime tests until every spec feature is green.
- TODO: Extend the **typechecker** suites (Go + TS) so they verify type rules and inference across modules. Assemble an exhaustive inference corpus exercising expression typing, generics, interfaces/impls, and cross-module reconciliation; ensure these scenarios are evaluated alongside runtime fixtures.

### 2025-10-31
- Regenerated the tree-sitter-able artifacts with the freshly rebuilt grammar (interface-composition fix now baked into `parser.c`/`.wasm`) using the local Emscripten toolchain; no diff surfaced, confirming the repo already carried the correct bits.
- Cleared local Go build caches (`.gocache`, `interpreter10-go/.gocache`) and re-ran `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` to mimic CI picking up the refreshed grammar without stale entries.
- ACTION: propagate the cache-trim guidance to CI docs if flakes recur; otherwise move on to the remaining parser fixture gaps (`design/parser-ast-coverage.md`).
- Mirrored the TypeScript placeholder auto-lift guardrails inside the Go interpreter so pipe placeholders evaluate eagerly, keeping the shared `pipes/topic_placeholder` fixture green.
- Parser sweep: both the TypeScript mapper and Go parser now skip inline comments when traversing struct literals, call/type argument lists, and struct definitions, with fresh TS/Go tests guarding the behaviour.
- TypeScript checker scaffold landed: basic environment/type utilities exist under `interpreter10/src/typechecker`, exported via the public index, and fixtures respect `ABLE_TYPECHECK_FIXTURES` ahead of the full checker port.
- TypeScript checker now emits initial diagnostics for logical operands, range bounds, struct pattern validation, and generic interface constraints so the existing error fixtures pass under `ABLE_TYPECHECK_FIXTURES=warn`.

### 2025-11-01
- TypeScript checker grew a declaration-collection sweep that registers interfaces, structs, methods, and impl blocks before expression analysis, mirroring the Go checker’s phase ordering.
- Implementation validation now checks that every `impl` supplies the interface’s required methods, that method generics/parameters/returns mirror the interface signature, and flags stray definitions; only successful impls count toward constraint satisfaction.
- Canonicalised type formatting to the v10 spec (`Array i32`, `Result string`) and keyed implementation lookups by the fully-instantiated type so generic targets participate in constraint checks.
- Extended the TypeScript checker’s type info to capture `nullable`, `result`, `union`, and function signatures, and taught constraint resolution to recognise specialised impls like `Show` for `Result string`, with fresh tests covering the new cases.
- Added focused Bun tests under `test/typechecker/` plus the `ABLE_TYPECHECK_FIXTURES` harness run to lock in the new behaviour and guard future regressions.
- Mirrored the Go checker and test suite to use the same spec-facing type formatter and wrapper-aware constraint logic so diagnostics now reference `Array i32`, `string?`, etc., and added parity tests (`interpreter10-go/pkg/typechecker/*`).

### 2025-11-02
- The Bun CLI suite (`interpreter10/test/cli/run_module_cli.test.ts`) now covers multi-file packages, custom loader search paths via the new `ABLE_MODULE_PATHS` env, and strict vs warn typecheck enforcement so the ModuleLoader refactor stays covered without pulling `stdlib/` into every run.
- Introduced the `able test` skeleton inside `scripts/run-module.ts`: it parses the planned flags/filters, materialises run options + reporter selection, and prints a deterministic plan summary before exiting with code `2` while the stdlib testing packages remain unparsable. (See `design/testing-cli-design.md` / `design/testing-cli-protocol.md`.)
- Extracted the shared package-scanning helpers (`discoverRoot`, `indexSourceFiles`, etc.) into `scripts/module-utils.ts` so other tooling (fixtures runner, future harnesses) can reuse the multi-module discovery logic without duplicating it.
- **Deferral noted:** full stdlib/testing integration is still on pause until the parser accepts `stdlib/src/test/*`; once that unblocks, wire the CLI skeleton into the `able.test` harness per the design notes.

### 2025-11-05
- Step 6 regression sweep ran end-to-end: `./run_all_tests.sh --typecheck-fixtures=warn` stayed green post-refactor, and `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` uncovered/validated the Go-side AST gaps.
- Go AST parity improvements: `FunctionCall.arguments` now serialises as an empty array when no args exist, `NilLiteral` carries `value: null`, and break/continue statements omit label/value when not supplied (matching the TS mapper contract).
- Parser helper normalisation no longer nukes empty slices before fixture comparison, and parameter lists default to `[]`, which brought the channel/mutex fixtures back in sync.
- Fixture `structs/functional_update` gained explicit `\"isShorthand\": false` flags so both interpreters agree on struct literal metadata going forward.

### 2025-11-06
- Go CLI now exposes `able check` alongside `able run`, sharing the manifest/target resolution pipeline. Both commands surface ProgramChecker diagnostics + package summaries and fail fast when typechecking reports issues, keeping the TypeScript + Go CLIs aligned.
- Added dedicated CLI tests to cover `able check` success/failure cases so future refactors keep the new mode wired through manifest resolution, typechecker reporting, and exit codes.

### 2025-11-07 — Phase α (Channel/Mutex) Completion
- Audited channel/mutex stdlib wiring across both runtimes: helper registration, typechecker signatures, fixtures, and prelude exports now match; no AST or scheduler drift detected.
- Added Bun smoke tests for nil-channel cancellation and mutex re-entry errors (`interpreter10/test/concurrency/channel_mutex.test.ts`) to mirror the Go parity suite.
- Documented the audit and captured the remaining TODO (map native errors to `ChannelClosed`/`ChannelNil`/`ChannelSendOnClosed`) in `design/channels-mutexes.md` and `spec/TODO.md`.
- Cleared `Phase α` from the active roadmap; next milestone is Phase 4 (cross-interpreter parity/tooling).

### 2025-11-07 — Fixture Parity Harness (Phase 4 Kick-off)
- Added a Go CLI entry point (`cmd/fixture`) that evaluates a single AST fixture and emits normalized JSON (result kind/value, stdout, diagnostics) with respect to `ABLE_TYPECHECK_FIXTURES`. The helper reuses the interpreter infrastructure and supports serial/goroutine executors while sandboxing the Go build cache.
- Refactored the TypeScript fixture loader into `scripts/fixture-utils.ts` so both the CLI harness and Bun tests can hydrate modules, install runtime stubs, and intercept `print` output consistently.
- Rebuilt `run-fixtures.ts` on top of the shared utilities (no behavior change) to keep fixture execution logic single-sourced.
- Introduced a Bun parity test (`test/parity/fixtures_parity.test.ts`) that exercises a representative slice of the shared fixture corpus (currently 20 fixtures across basics + concurrency) against both interpreters and asserts matching results/stdout via the new Go CLI.

### 2025-11-08 — Parity CLI Reporting
- Added `interpreter10/scripts/run-parity.ts`, reusable parity helpers, and Bun parity suites that now share the same execution/diffing logic across AST fixtures and curated examples.
- `run_all_tests.sh` now invokes the parity CLI so local + CI runs execute the same cross-interpreter verification and drop a JSON report at `tmp/parity-report.json` for machine-readable diff tracking; `tmp/` landed in `.gitignore` to keep artifacts out of commits.
- The helper script also honors `ABLE_PARITY_REPORT_DEST` or `CI_ARTIFACTS_DIR` so pipelines can copy the parity JSON into their artifact buckets without bespoke wrapper scripts.
- Updated `interpreter10/README.md` and `interpreter10-go/README.md` with parity CLI instructions, env knobs (`ABLE_PARITY_MAX_FIXTURES`, `ABLE_PARITY_REPORT_DEST`), and guidance on keeping the cross-interpreter harness green.
- Added Go package docs (`pkg/interpreter/doc.go`, `pkg/typechecker/doc.go`) plus README guidance on regenerating `go doc`/pkg.go.dev pages so the documentation workstream is unblocked.
- Landed `dynimport_parity` and `dynimport_multiroot` in `interpreter10/testdata/examples/` to cover dynamic package aliasing, selector imports, and multi-root dynimport scenarios end-to-end; the parity README + plan now list them alongside the other curated programs, and the Go CLI + Bun harness honor `ABLE_MODULE_PATHS` when resolving shared deps.
- Authored `docs/parity-reporting.md` and linked it from the workspace README so CI pipelines know how to persist `tmp/parity-report.json` via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR`.
- The Go CLI (`cmd/able`) now honors `ABLE_MODULE_PATHS` in addition to `ABLE_PATH`, with new tests ensuring the search-path env works; stdlib docs reference the alias so multi-root dynimport scenarios can rely on a single env knob across interpreters.
- Fixed the `..`/`...` range mapping bug in both parsers so inclusive ranges now follow the spec (TS + Go parser updates, new parser unit tests, interpreter for-loop regression tests, and fizzbuzz-style parity coverage).

### Phase 5 Foundations — Parser Alignment
- Canonical AST mapping now mirrors the fixture corpus across both runtimes. The TypeScript mapper’s fixture parity suite (`bun test test/parser/fixtures_mapper.test.ts`) and the Go parser harness (`go test ./pkg/parser`) stay green, so every tree-sitter node shape maps to the shared AST contract with span/origin metadata.
- `tree-sitter-able` grammar coverage is complete for the v10 surface (see `design/parser-ast-coverage.md`); new syntax is added directly with fixture+cXPath tests so the grammar remains authoritative.
- Translators and loaders are live in both interpreters: TypeScript’s `ModuleLoader` and Go’s `driver.Loader` now ingest `.able` source via tree-sitter, hydrate canonical AST modules, and feed them to their respective typechecker/interpreter pipelines.
- End-to-end parse → typecheck → interpret tests exercise both runtimes: `ModuleLoader pipeline with typechecker` (Bun) covers the TS path, and `pkg/interpreter/program_pipeline_test.go` drives the Go loader/interpreter via `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`.
- Diagnostic coverage now rides on the same pipelines: the new Bun test asserts missing import selectors surface typechecker errors before evaluation, and the Go suite verifies that `EvaluateProgram` halts (or proceeds when `AllowDiagnostics` is set) when return-type violations are reported.

### 2025-11-22
- TS typechecker imports now seed struct/interface bindings from package summaries (with generic placeholders) so stdlib types preserve their shape when referenced from dependent modules.
- Added builtin HashSet method stubs (new/with_capacity/add/remove/contains/size/clear/is_empty) to the TS typechecker, letting the hash_set ModuleLoader integration typecheck without ignoring bool-condition diagnostics.
- HashSet stdlib integration test now expects zero diagnostics; the stdlib Bun suite remains green.

### 2025-11-23
- Quarantine stdlib iterators now return the explicit `IteratorEnd {}` sentinel and match on the sentinel type to avoid pattern collisions; iterator imports across array/list/linked_list/vector/lazy_seq/string/automata DSL modules now pull from `core.iteration`.
- TS stdlib integration suite rerun to confirm no regressions in the active modules; Go unit tests remain green.
- TS typechecker now treats `Iterator` interface annotations as structural for literal overflow checks, so iterator literals annotated as `Iterator u8` surface integer-bound diagnostics on yielded values; `run_all_tests.sh --version=v11` is green.

### 2025-11-24
- Primitive `string` is now treated as an iterable of `u8` across both runtimes: TS/Go typecheckers recognise string for-loops, diagnostics reference `array, range, string, or iterator`, and new tests cover typed-pattern matches plus runtime iteration backed by the stdlib string module.
- Added ModuleLoader + Go runtime tests to ensure string iteration requires importing `able.text.string` and yields byte values; stdlib README documents the import requirement.

### 2025-11-25
- Added a guardrail fixture for missing Apply/IndexMut implementations (`interfaces/apply_index_missing_impls`) and regenerated exports/baseline so warn/strict runs expect the shared diagnostics.
- Aligned Go typechecker diagnostics with TS for non-callable Apply targets and Index-only [] assignments (now report “non-callable … missing Apply implementation” and “cannot assign via [] without IndexMut …”), keeping the parity suite green.
- Full v11 sweep rerun after the additions (`./run_all_tests.sh --version=v11`) is green.

### 2025-12-03
- Swapped import alias syntax to the `::` rename operator across the tree-sitter grammar, TS/Go parsers, and module loaders, rejecting legacy `as` aliases while keeping dot traversal unchanged.
- Updated fixtures/docs/tests to the new syntax (package aliases, selective aliases for static/dynimport, struct-pattern rename coverage) and re-exported the shared fixture corpus.
- Ran `./run_all_tests.sh --version=v11` (TS+Go units, fixtures, parity); all suites passed and parity report saved to `v11/tmp/parity-report.json`.

### 2025-12-09
- Finalized the v11 operator surface: `%`/`//`/`/%` follow Euclidean semantics, `%=` compounds and the dot-prefixed bitwise set (`.& .| .^ .<< .>>`) are supported, `^` acts as exponent (bitwise xor remains dotted), and the operator interfaces in `core.interfaces` mirror the runtime behavior. Parser/AST/typechecker/stdlib fixtures all align and legacy `%%` syntax is gone.
- Verified the full sweep (`./run_all_tests.sh --version=v11`) stays green after the operator updates (TS + Go units, fixtures, parity harness).

### 2025-12-10
- Kernel/stdlib discovery now covers the v11 layout: search-path collectors scan `v11/kernel/src` and `v11/stdlib/src`, TS ModuleLoader/Go CLI auto-load kernel packages when bundled, and new tests in both runtimes assert the v11 scan paths.
- TS/Go CLI/module loader tests updated to exercise the bundled scan; stdlib README notes the expanded auto-detection coverage.
- Manifested runs now pin the bundled boot packages: Go `able deps install` injects both stdlib and kernel into `package.lock`, kernel search-path discovery honors module-path env entries, and the TS CLI reads `package.yml`/`package.lock` to add dependency roots before falling back to bundled detection. New CLI tests cover lock-required runs plus pinned stdlib/kernel boot without env overrides.

### 2025-12-10 — Ratio & Numeric Conversions
- Implemented exact `Ratio` struct and normalization helpers in the stdlib kernel, added `to_r` conversions for integers/floats, and expanded numeric smoke tests (`v11/stdlib/src/core/numeric.able`, `v11/stdlib/tests/core/numeric_smoke.test.able`).
- TypeScript runtime/typechecker now treat `Ratio` as numeric: builtin `__able_ratio_from_float`, exact ratio arithmetic/comparisons, NumericConversions support, and new ratio tests (`v11/interpreters/ts/src/interpreter/{numeric.ts,operations.ts,numeric_host.ts}`, `v11/interpreters/ts/src/typechecker/**`, `v11/interpreters/ts/test/{typechecker/numeric.test.ts,basics/ratio.test.ts}`).
- Go runtime/typechecker mirror the Ratio struct/builtin and conversion helpers with adjusted constraint diagnostics plus coverage in interpreter/typechecker suites (`v11/interpreters/go/pkg/interpreter/*`, `v11/interpreters/go/pkg/typechecker/*`).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/numeric.test.ts`; `cd v11/interpreters/ts && bun test test/basics/ratio.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`; `cd v11/interpreters/go && go test ./pkg/interpreter`.

### 2025-12-24
- Added exec fixture `exec/06_03_operator_overloading_interfaces` to cover Add/Index/IndexMut operator dispatch and updated the conformance plan + coverage index.
- Go interpreter now dispatches arithmetic/bitwise operators to interface implementations when operands are non-numeric.
- Range expressions now enforce integer bounds in both runtimes/typecheckers; updated range diagnostics/tests to match the v11 spec.
- Cleared the operator-overloading exec fixture item from the v11 PLAN backlog.
- Added composition exec fixtures for combined behavior: `exec/09_00_methods_generics_imports_combo` (imports + generics + methods) and `exec/11_00_errors_match_loop_combo` (match + loop + rescue), with coverage index + conformance plan updates.
- Added exec fixture `exec/06_03_safe_navigation_nil_short_circuit` to cover `?.` short-circuiting, receiver evaluation, and argument skipping, and updated coverage + conformance docs.
- Added exec fixture `exec/06_04_function_call_eval_order_trailing_lambda` to cover left-to-right call argument evaluation and trailing lambda equivalence, with coverage + conformance updates.
- Added exec fixture `exec/06_06_string_interpolation` to cover interpolation escapes and multiline string literals, with coverage + conformance updates.
- Added exec fixture `exec/06_07_generator_yield_iterator_end` to cover yield/stop semantics and IteratorEnd exhaustion behavior, with coverage + conformance updates.
- Tree-sitter grammar now allows multiline double-quoted strings; TS/Go parsers unescape interpolation text for `\\$`/`\\`` (and `\\\\`) so backtick escapes follow the v11 spec.
- Pattern matching now treats `IteratorEnd {}` as a match for the iterator end sentinel in both interpreters.
- Added exec fixture `exec/06_12_01_stdlib_string_helpers` covering required string helper semantics (lengths, substring bounds, split/replace, prefix/suffix) and updated coverage/conformance tracking; cleared the PLAN backlog item.
- Added exec fixture `exec/06_12_02_stdlib_array_helpers` for the required array helper API (size, push/pop, get/set, clear) with coverage/conformance updates; cleared the PLAN backlog item.
- Added exec fixture `exec/06_12_03_stdlib_numeric_ratio_divmod` covering Ratio normalization/to_r and Euclidean /% results with coverage/conformance updates; cleared the PLAN backlog item.
- Added `as` cast expressions to the grammar + AST contract and implemented explicit numeric/interface casts in both interpreters and typecheckers.
- Stdlib numeric cleanup: replaced unsupported `const`/`mut`/`else if`, normalized i128 constants, removed duplicate Ratio numerator/denominator methods in favor of kernel definitions, and added statement terminators where the parser requires them.
- Inherent methods now resolve without requiring the method name in the caller scope (TS + Go) so stdlib extensions work through package imports; refreshed the numeric ratio/divmod exec fixture import to use stdlib `Ratio`.

### 2025-12-26
- Added exec fixtures `exec/07_01_function_definition_generics_inference` (implicit/explicit generics + return inference), `exec/07_03_explicit_return_flow` (explicit return flow), `exec/07_04_trailing_lambda_method_syntax` (method call syntax + trailing lambda parity), `exec/07_04_apply_callable_interface` (Apply callables), `exec/07_05_partial_application` (placeholder partial application), and `exec/07_06_shorthand_member_placeholder_lambdas` (implicit member/method shorthand + placeholder lambdas); updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- TS parser/typechecker now accept implicit member assignments as valid assignment targets, matching the runtime semantics.
- Tests: `ABLE_FIXTURE_FILTER=07_01_function_definition_generics_inference bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_01_function_definition_generics_inference$`; `ABLE_FIXTURE_FILTER=07_03_explicit_return_flow bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_03_explicit_return_flow$`; `ABLE_FIXTURE_FILTER=07_04_trailing_lambda_method_syntax bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_04_trailing_lambda_method_syntax$`; `ABLE_FIXTURE_FILTER=07_04_apply_callable_interface bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_04_apply_callable_interface$`; `ABLE_FIXTURE_FILTER=07_05_partial_application bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_05_partial_application$`; `ABLE_FIXTURE_FILTER=07_06_shorthand_member_placeholder_lambdas bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_06_shorthand_member_placeholder_lambdas$`.
- Method resolution now honors name/type-in-scope gating without breaking kernel/primitive method access; fixed UFCS name lookup and aligned TS/Go member resolution with type-name visibility.
- Updated TS typechecker tests to match truthiness semantics (if/while/match/rescue guards) and swapped the diagnostic location test to an undefined identifier.
- Refreshed fixtures for truthiness (`errors/logic_operand_type`) and string/numeric exec imports; updated the AST typecheck baseline for the logic operand fixture.
- Tests: `./run_all_tests.sh --version=v11` (TS + Go units, fixtures, parity) with parity report in `v11/tmp/parity-report.json`.

### 2025-12-27
- Added exec fixtures `exec/07_07_overload_resolution_runtime`, `exec/08_01_if_truthiness_value`, and `exec/08_01_match_guards_exhaustiveness`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- `if` expressions now return `nil` (not `void`) when no branch matches and there is no else, aligning TS/Go runtimes with the v11 spec; updated `errors/rescue_guard` fixture expectation.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`); `./run_all_tests.sh --version=v11`.
- Next: continue exec fixture backlog starting at `exec/08_02_while_continue_break` and `exec/08_02_loop_expression_break_value`.

### 2025-12-28
- Added exec fixtures `exec/08_02_while_continue_break`, `exec/08_02_loop_expression_break_value`, and `exec/08_02_range_inclusive_exclusive`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- Updated the generated AST fixture expectation for `errors/rescue_guard` in the TS export fixture source to keep it aligned with nil-returning `if` expressions.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Next: continue exec fixture backlog starting at `exec/08_03_breakpoint_nonlocal_jump`.
- Added exec fixtures `exec/13_06_stdlib_package_resolution` and `exec/13_07_search_path_env_override`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- Exec fixture runners now honor manifest-provided env overrides for module search paths; TS uses CLI-style search path resolution and Go exec fixtures mirror the env-driven roots.

### 2025-12-29
- Added exec fixture `exec/08_03_breakpoint_nonlocal_jump` and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- Tests: `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Next: continue exec fixture backlog starting at `exec/09_05_method_set_generics_where`.
- Codified alias/re-export method propagation and conflict semantics in the v11 spec and cleared the remaining spec TODO.
- Added AST fixture `errors/alias_reexport_method_ambiguity` (with setup packages) plus baseline entry; TS/Go typecheckers now surface ambiguous overloads when multiple method sets attach the same signature.
- Tests: `ABLE_FIXTURE_FILTER=alias_reexport_method_ambiguity bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestFixtureParityStringLiteral/errors/alias_reexport_method_ambiguity`.

### 2025-12-30
- Added exec fixture `exec/09_05_method_set_generics_where` covering method-set generics/where constraints for instance + UFCS calls; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- Go interpreter now enforces method-set generic/where constraints during method calls; call helpers were split into `v11/interpreters/go/pkg/interpreter/call_helpers.go` to keep call logic files under 1000 lines.
- Aligned TS + Go parser handling of impl interface args so space-delimited arg lists are not collapsed into a single generic application; fixture mapper parity is green again.
- Spec: documented interface-arg grouping by interface arity (greedy left-to-right grouping with parentheses to force a single argument).
- TS/Go typechecker + runtime now group impl interface args by interface arity, so unparenthesized generic applications like `Map K V` remain a single argument when the interface expects one.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).

### 2025-12-31
- Added exec fixture `exec/10_01_interface_defaults_composites` covering interface defaults, implicit vs explicit `Self`, and composite aliases; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- TS/Go runtimes now treat composite interfaces as base-interface bundles for interface coercion and method dispatch (including default methods), with interface checks honoring base interfaces.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Added exec fixture `exec/07_02_01_verbose_anonymous_fn` covering verbose anonymous functions (generics + where clauses).
- Stdlib fixes: added numeric interfaces to `core/numeric`, corrected rational i128 min/max constants, and added `Queue.is_empty` inherent method to avoid interface ambiguity.
- String stdlib: added `__able_char_to_codepoint` host builtin (TS/Go/kernel) and rewrote `char_to_utf8` to use it; string smoke tests now import `able.text.string`.
- Test cleanup: corrected array filter expectations, added semicolons to avoid `Array.new()`/`for` parse ambiguity, removed unsupported heap comparator test, and renamed lazy seq iterator to avoid duplicate impls.

### 2025-12-26
- Added exec fixture `exec/11_01_return_statement_type_enforcement` and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- TS/Go runtimes now enforce return type checks (including bare return for non-void), and the TS/Go typecheckers ignore unreachable tail expressions after return while rejecting bare returns in non-void functions.
- Updated `exec/11_00_errors_match_loop_combo` to use `//` so integer division stays within `i32` per the v11 spec.
- Added exec fixture `exec/11_02_option_result_or_handlers` (Option/Result `or {}` handlers) and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- Return type enforcement now treats iterator values as `Iterator` interface matches in TS/Go, and Go generic interface checks accept interface implementers; updated the pipeline diagnostics test to expect runtime mismatch under `AllowDiagnostics`.
- Tests: `./run_all_tests.sh --version=v11 --fixture`; `./run_all_tests.sh --version=v11`.

### 2026-01-10
- TS interpreter entrypoint tasks now preserve raw runtime errors, support re-entrant `proc_flush`, and prioritize generator continuations so iterator yields work inside proc contexts.
- CLI/fixture/parity runners now bind entrypoint `main` calls via a dedicated environment to avoid missing symbol errors; entrypoint-only async helpers still gate user-facing `proc_yield`/`proc_cancelled`.
- Tests: `bun test test/concurrency/native_suspend.test.ts`; `bun test test/cli/run_module_cli.test.ts`; `bun test test/parity/examples_parity.test.ts`; `bun test test/parity/fixtures_parity.test.ts`.

### 2026-01-11
- Added `dyn.Package.eval`/`dyn.eval` plumbing for REPL-style evaluation (parse errors return `ParseError` with `is_incomplete`), and mirrored dynamic-definition rebinding rules in Go imports/definitions.
- Added stdlib `able.repl` module (line editor, `:help`/`:quit`, prints non-`void` results) plus `able repl` CLI support for TS and Go.
- Spec: documented `ParseError`/`Span` plus dynamic eval APIs and REPL-oriented parse error semantics.
- Tests: `bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.

### 2026-01-13
- Ran the full v11 sweep; parity report refreshed at `v11/tmp/parity-report.json`.
- Tests: `./run_all_tests.sh --version=v11`.
- Added `able.io.temp` for temp file/dir creation + cleanup helpers, and added `io.puts`/`io.gets` wrappers in `able.io`.
- Extended stdlib IO tests with temp helper coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/io.test.able`; `./v11/ablego test v11/stdlib/tests/io.test.able`.
- Expanded Path tests for mixed separators and UNC roots, and fs tests for missing directory reads.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`; `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.
- Expanded stdlib IO/Path/fs edge coverage (non-positive reads, empty paths, remove missing, empty read_lines, differing roots).
- Tests: `./v11/ablets test v11/stdlib/tests/io.test.able`; `./v11/ablego test v11/stdlib/tests/io.test.able`; `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`; `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.
- Added a PermissionDenied stdlib fs test and fixed Go extern singleton struct decoding so union kinds map correctly from host strings; tightened permission error detection in Go stdlib error mapping.
- Tests: `./v11/ablego test v11/stdlib/tests/fs.test.able`.
- Ran the full v11 sweep; parity report refreshed at `v11/tmp/parity-report.json`.
- Tests: `./run_all_tests.sh --version=v11`.
- Lowercased the `able.io.path` package name and updated stdlib/tests/docs imports and call sites to use `path.*`.
- Tests: `./run_all_tests.sh --version=v11`.
- Next: finish stdlib IO coverage (errors, path normalization, IO handle edge cases) and keep `./run_all_tests.sh --version=v11` green.

### 2026-01-14
- Preserved entrypoint runtime diagnostic context in the TS scheduler so raw runtime errors keep locations/stack notes.
- Standardized TS diagnostics path normalization to the repo root, added shared path helpers, and normalized fixture origins accordingly.
- Corrected infix/postfix expression spans so runtime diagnostics point at full expressions instead of suffix/right operands.
- Propagated return-statement context into return type mismatch errors and refreshed fixture expectations/baselines for new paths/notes.
- Tests: `bun run scripts/run-fixtures.ts`.
- Split Go interpreter member resolution into smaller modules (`interpreter_members.go`, `interpreter_method_resolution.go`) to keep files under 900 lines.
- Tests: `./run_all_tests.sh --version=v11` (passed); `./run_stdlib_tests.sh` (TS stdlib failed at `v11/stdlib/tests/fs.test.able:202`, Go stdlib passed).
- Fixed `fs.copy_dir` overwrite behavior by clearing destination contents after a removal attempt when needed.
- Tests: `./run_stdlib_tests.sh`; `./run_all_tests.sh --version=v11`.
- Dropped redundant where-clauses from `HashMap` impls so stdlib typechecking stays clean in strict mode.
- Updated strict fixture manifests for expected typechecker diagnostics and refreshed the baseline for `expressions/map_literal_spread`.
- Removed the `PathType` alias from fs helpers; paths now use `Path | String` directly and fs helpers extend `Path`.
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict --fixture`; `./run_stdlib_tests.sh`.
- Fixed impl validation to compare interface method where clauses against method-level constraints instead of impl-level where clauses (TS + Go typecheckers).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/implementation_validation.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`.
- Added exec fixture `exec/10_02_impl_where_clause` to cover impl-level where clauses without method-level duplication and updated the coverage index.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_02_impl_where_clause bun run scripts/run-fixtures.ts`.
- Added exec fixture `exec/04_05_04_struct_literal_generic_inference`, updated the exec coverage index, and enforced exec-fixture typechecking when manifests specify diagnostics (TS + Go).
- Fixed struct literal generic type-argument handling in the TS and Go typecheckers (placeholder args in TS; inferred args in Go).
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict`.
- Clarified spec call-site inference to include return-context expected types and documented return-context inference design notes; updated PLAN work queue.

### 2026-01-15
- TS typechecker now uses expected return types to drive generic call inference (explicit + implicit return paths), with new return-context unit tests.
- Go typechecker now propagates expected return types through implicit return blocks, plus a focused unit test for implicit-return inference.
- Added exec fixture `exec/07_08_return_context_generic_call_inference` and updated the exec coverage index.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/return_context_inference.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker -run TestGenericCallInfersFromImplicitReturnExpectedType`.
- TS typechecker now treats method-shorthand exports as taking implicit self for overload resolution, and uses receiver substitutions when enforcing method-set where clauses on exported function calls.
- TS runtime now treats unresolved generic type arguments on struct instances as wildcard matches when comparing against concrete generic types.
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict`.
- Documented kernel Hash/Eq decisions (sink-style hashing, IEEE float equality, floats not Eq/Hash), updated spec wording, and expanded the PLAN work breakdown for interpreter alignment.
- Extended the kernel Hash/Eq plan to move the default `Hasher` implementation into Able with host bitcast helpers; updated spec TODOs and PLAN tasks accordingly.
- Added a kernel-level FNV-1a Hasher (Able code) with big-endian byte emission, introduced `__able_f32_bits`/`__able_f64_bits`/`__able_u64_mul` helpers in TS/Go, and updated stdlib hashing call sites + tests to use the new sink-style Hash API.

### 2026-01-16
- Added common `HashSet` set operations (union/intersect/difference/symmetric difference, subset/superset/disjoint) plus new spec coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/collections/hash_set.test.able`; `./v11/ablego test v11/stdlib/tests/collections/hash_set.test.able`.
- Spec: documented the always-loaded `able.kernel` contract (core interfaces, HashMap, KernelHasher, hash bridges) and clarified map literal key constraints plus hash container semantics.
- Spec: defined the `Hasher` interface and tied primitive Hash/Eq/Ord impls to the kernel library; aligned kernel string/char bridge names.
- Spec: enumerated kernel-resident types/interfaces/methods and listed the full `Hasher` helper surface with default semantics.
- TS interpreter: track struct definitions and treat concrete type names as taking precedence over interface names during runtime coercion/matching to fix `Range` vs `Range` interface collisions.
- Go typechecker: unwrap interface aliases (e.g., `Clone`, `Eq`, `Hash`) when collecting impls, validating impls, and solving constraints.
- Spec TODOs: cleared the kernel hashing contract items now captured in `spec/full_spec_v11.md`.
- Tests: `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh --version=v11`.
- Go typechecker: allow impl targets to be interface types (supporting `impl Iterable T for Iterator T` matches).
- Go interpreter: treat missing generic args as wildcards when matching impl targets; record iterator values as `Iterator _` for runtime type info.
- Stdlib: fixed iterable helper signatures in `able.core.iteration` after adding iterator-as-iterable support.
- Tests: `./v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Go typechecker: allow interface default methods to satisfy member access when an impl omits the method body.
- Tests: `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Stdlib: added `default<T: Default>()` helper in `able.core.interfaces`.
- Stdlib/spec: added `Extend` interface + `Iterable.collect` default, with Array/HashSet impls and iteration tests.
- Next: resume the PLAN work queue (regex parser + quantifiers).

### 2026-01-18
- Spec: clarified interface dynamic dispatch as dictionary-based (default methods + interface-impl method availability).
- TS typechecker: added type-parameter tracking in expressions, inference for interface-method generics, and base-interface method candidates.
- TS interpreter: interface values now carry method dictionaries (incl. iterator natives), interface-member binding handles native methods, and for-loops accept interface-wrapped iterators.
- Go typechecker: collect transitive impls/method sets for imports, preserve interface metadata on impls for default methods, and write inferred call type arguments into the AST.
- Stdlib: fixed `Iterable.map`/`filter_map` generic parameter syntax.
- Tests: `v11/ablets .examples/foo.able`; `v11/ablego .examples/foo.able`; `v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `v11/ablego test v11/stdlib/tests/core/iteration.test.able`; `v11/ablets test v11/stdlib/tests/collections/hash_set.test.able`; `v11/ablego test v11/stdlib/tests/collections/hash_set.test.able`; `v11/ablets test v11/stdlib/tests/collections/hash_set_smoke.test.able`; `v11/ablego test v11/stdlib/tests/collections/hash_set_smoke.test.able`.

### 2026-01-19
- Spec: documented package-qualified member access as yielding first-class values (type aliases remain type-only).
- TS typechecker: package member access now resolves symbol types from summaries (function values included), enabling `pkg.fn` usage in expressions.
- Tests: `v11/ablets .examples/foo.able`; `v11/ablego .examples/foo.able`.
- Stdlib tests: added iteration coverage for `collect` via Default/Extend and package-qualified function values.
- Tests: `v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Spec: documented the `Default` interface signature and its stdlib helper.
- Design: captured the eager vs lazy collections split (`Iterable` minimal, lazy adapters on `Iterator`, eager `Enumerable`).
- Spec: documented the `Enumerable` interface and lazy/eager split in the iteration section; updated core iteration protocol.
- Stdlib: made `Enumerable` parseable under current grammar (removed base-interface clause and HKT `where` constraints); documented `Iterable`’s "implement either each/iterator" intent; moved `Queue` operations to inherent methods.
- TS interpreter: added Error/Awaitable/Iterator interface-value handling for native values (default methods + await helpers), and allow generic interface values to satisfy `matchesType`.
- TS tests: aligned `Display` dispatch test with `to_string`.
- Tests: `bun test` in `v11/interpreters/ts`; `go test ./...` in `v11/interpreters/go`.
- Stdlib: added explicit `Enumerable.lazy` impl for `Array` to keep lazy iterators reachable under Go.
- Tests: `./v11/ablego test v11/stdlib/tests/enumerable.test.able --format tap`.

### 2026-01-20
- Go typechecker: instantiate generic unions when resolving type annotations and normalize applied union types for assignability.
- Parser/typechecker: where-clause subjects now accept type expressions; interface bases parse via `for ... : ...`; fixture printer updated and AST fixtures regenerated with new `where` shape.
- Typechecker: base interface signatures now participate in impl validation; self-type pattern names map to concrete `Self` substitutions.
- Kernel/stdlib: added `PartialEq`/`PartialOrd` impls for non-float primitives and big-number types; `Ord` impls now define `partial_cmp` to satisfy base interface contracts.
- Runtime: impl method resolution now prefers direct interface matches over base interface matches to avoid ambiguity (TS + Go).
- Fixtures: updated `implicit_generic_where_ambiguity` diagnostics + typecheck baseline; adjusted TreeMap stdlib test to include `partial_cmp` on custom `Ord` keys.
- Tests: `bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && bun test`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./...`.
- Kernel: compare String bytes for `PartialEq`/`Eq` to avoid recursive `==` on struct-backed strings.
- TS/Go interpreters: lambdas now treat `return` as local by catching return signals.
- Tests: `./v11/ablets test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_builder.test.able --format tap`.

### 2026-01-21
- Typechecker: higher-kinded self patterns now reject concrete targets unless the impl is still a type constructor (arity-aware in TS + Go).
- Parser: type application parsing prefers left-associative space-delimited arguments; tree-sitter assets + corpus refreshed and Go parser relinked.
- Stdlib: removed Array overrides from `Enumerable` impl to rely on interface defaults; added exec fixtures for type-arg arity + associativity diagnostics.
- Tests: `cd v11/parser/tree-sitter-able && tree-sitter test -u`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`; `./run_all_tests.sh --version=v11`.
- TS interpreter: separate interface vs subject type arguments in impl resolution (`findMethod`, `resolveInterfaceImplementation`, `matchImplEntry`, `typeImplementsInterface`) and widen runtime generic skipping for nested type expressions.
- TS interpreter: bind receiver type arguments into function env as `type_ref` and allow typed patterns to resolve generic type refs (fallback to wildcard for unknown generic names) to avoid non-exhaustive matches in generic matchers.
- Tests: `/home/david/sync/projects/able/v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests/assertions.test.able`; `/home/david/sync/projects/able/v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests` (timed out after 60s).
- Typechecker: overload resolution now prefers non-generic matches over generic, with generic specificity ranking aligned across TS + Go; unknown argument types no longer satisfy overload sets in Go to match TS (UFCS ambiguity case).
- CLI: test runner skips typechecking in `--list`/`--dry-run` modes to avoid spurious stdlib diagnostics.
- Fixtures: updated UFCS overload expectations and typecheck baseline entries; refreshed export-fixtures manifests for overload cases.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker` in `v11/interpreters/go`; `bun test test/parity/fixtures_parity.test.ts -t "functions/ufcs_generic_overloads"`; `ABLE_FIXTURE_FILTER=errors/ufcs_overload_ambiguity bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.
- Typechecker: bind higher-kinded self pattern placeholders to partially applied targets and apply interface args to `Self` when the impl target is still a type constructor (TS + Go); flatten applied types during substitution for nested constructor applications.
- Tests: `go test ./pkg/typechecker -run TestImplementationAllowsPartiallyAppliedConstructorWithSelfPattern` in `v11/interpreters/go`; `bun test test/typechecker/implementation_validation.test.ts -t "partially applied constructor"` in `v11/interpreters/ts`.
- Typechecker (Go): only apply explicit interface args to `Self` for constructor targets so inferred self-pattern args don't double-apply in method validation.
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker` in `v11/interpreters/go`; `bun test test/typechecker` in `v11/interpreters/ts`.
- Typechecker (TS): avoid overriding impl generic substitutions with self-pattern bindings during impl validation to prevent nested applied types in stdlib interface checks.
- Tests: `./run_all_tests.sh --version=v11`.
- Typechecker: add regression coverage to ensure self-pattern placeholders remain in-scope for interface method signatures (TS + Go).
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker -run TestImplementationAllowsSelfPatternPlaceholderInMethodSignature` in `v11/interpreters/go`; `bun test test/typechecker/implementation_validation.test.ts -t "self placeholders"` in `v11/interpreters/ts`.
- Typechecker: add package-scoped duplicate declaration coverage so same symbol names across packages do not conflict (TS session + Go program checker).
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker -run TestProgramCheckerAllowsDuplicateNamesAcrossPackages` in `v11/interpreters/go`; `bun test test/typechecker/duplicates.test.ts -t "same symbol name"` in `v11/interpreters/ts`.
- Fixtures: added exec coverage for typechecker return mismatch diagnostics to mirror runtime behavior checks (`exec/11_01_return_statement_typecheck_diag`).
- Tests: `ABLE_FIXTURE_FILTER=11_01_return_statement_typecheck_diag bun run scripts/run-fixtures.ts` in `v11/interpreters/ts`; `GOCACHE=/tmp/able-gocache go test ./pkg/interpreter -run TestExecFixtures/11_01_return_statement_typecheck_diag` in `v11/interpreters/go`.

### 2026-01-22
- Stdlib: moved `collect` to the `Iterator` interface and kept `Iterable` focused on `each`/`iterator`.
- Spec: documented `Iterator.collect` in both iteration sections and removed `collect` from `Iterable`.
- Design: updated the eager/lazy collections split doc to reflect `Iterator.collect`.
- Tests not run (docs + stdlib interface change only).

### 2026-01-23
- Typechecker (TS): scope duplicate declaration tracking by package (prelude-safe) and allow local bindings to shadow package aliases during member access.
- Kernel: added `__able_os_args`/`__able_os_exit` externs to `v11/kernel/src/kernel.able` for kernel/stdlib alignment.
- Tests: `./run_stdlib_tests.sh --version=v11`; `./run_all_tests.sh --version=v11`.
- Typechecker: enforce missing type-argument diagnostics for concrete type annotations (TS + Go), while allowing constructor targets for impls/method sets; avoid duplicate arity diagnostics for constraints.
- Fixtures: added builtin type-arity + partial-application regression fixtures; refreshed typecheck baseline.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`; `cd v11/interpreters/ts && bun test test/typechecker/constraint_arity.test.ts`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker`.
- Fixtures: updated nested struct destructuring + Apply/Index dispatch fixtures to unwrap index results (`!`) so happy-path tests typecheck cleanly; refreshed typecheck baseline.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`.
- Spec: unified async model around `spawn`/`Future`, removed `proc`, renamed helpers to `future_*`, and rewrote Section 12 accordingly.
- Design: added `v11/design/future-unification.md`; updated concurrency/AST/typechecker/stdlib design notes to align with unified Future semantics.
- Plan: added a comprehensive implementation breakdown for the unified Future change.
- Tests not run (docs/spec/plan changes only).
- Parser/AST: removed `proc` keyword/`proc_expression` from tree-sitter, regenerated grammar artifacts, updated parser corpus; removed `ProcExpression` from TS/Go AST schemas and parser mappers; fixture JSON now uses `SpawnExpression`.
- Runtime: await/channel helpers now accept future contexts (TS + Go), and `proc_cancelled` works inside spawned futures in Go.
- Typechecker: future `cancel()` is allowed (TS + Go), and concurrency/typechecker tests updated to use spawn/future semantics.
- Fixtures: `.able` sources updated from `proc` → `spawn`, and expected error strings updated from `Proc failed/cancelled` → `Future failed/cancelled`.
- Tests not run (parser + runtime + fixture changes only).

### 2026-01-30
- Compiler: added Go compiler scaffolding to emit Go struct types, literal-return function bodies, wrapper registrations, and struct conversion helpers; added a bridge runtime for interpreter/compiled interop and a new `ablec` CLI.
- Compiler: extended codegen to handle identifiers, `+/-/*` binary expressions, and simple compiled-function calls; added a compiler-backed exec harness that builds and runs a tiny compiled program.
- Compiler: added multi-statement bodies with implicit return, plus `:=`/`=` identifier bindings (typed patterns supported); expanded statement lowering to evaluate expressions via `_ = expr`.
- Compiler: compile array literals into runtime array values and support struct member access; map Array/Map/HashMap types to runtime values for safe pass-through; added compiler coverage for array literals and member access.
- Compiler: added panic/recover plumbing for compiled wrappers, plus bridge/interpreter helpers to convert panicked runtime values into raise signals.
- Compiler: added runtime bridge index helper, global runtime registration, and index-expression codegen (runtime.Value-only) with compiler coverage.
- Compiler: index-expression codegen now converts runtime values into expected primitive/struct types with panic-on-conversion failure.
- Fixtures: added exec coverage for compiler index expressions (statement form) and included it in the compiler fixture parity list.
- Compiler: added index assignment lowering via runtime bridge and new compiler exec fixture coverage for assignment.
- Compiler: array literal lowering now converts struct elements into runtime values for interop.
- Fixtures: added compiler exec fixture for array literals with struct elements and import fix for array helpers.
- Compiler: assignment expressions returning runtime values now coerce to expected types; added compiler exec fixture for index assignment return value.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler` (including `TestCompilerExecHarness`) in `v12/interpreters/go`; `GOCACHE=$(pwd)/.gocache go test ./cmd/ablec` in `v12/interpreters/go`; `GOCACHE=$(pwd)/v12/interpreters/go/.gocache ./run_all_tests.sh`; `GOCACHE=$(pwd)/v12/interpreters/go/.gocache ./run_stdlib_tests.sh`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler` in `v12/interpreters/go`.

### 2026-01-31
- Compiler: added `as` type-cast lowering via the interpreter bridge, including runtime helper emission and type-expression rendering for compiled code.
- Compiler: added `06_03_cast_semantics` to the compiled exec fixture list for coverage.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `10_04_interface_dispatch_defaults_generics` and `10_15_interface_default_generic_method` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_04_interface_dispatch_defaults_generics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_15_interface_default_generic_method go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `13_01_package_structure_modules`, `13_03_package_config_prelude`, `13_04_import_alias_selective_dynimport`, `13_05_dynimport_interface_dispatch`, `13_06_stdlib_package_resolution`, and `13_07_search_path_env_override` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_03_package_config_prelude,13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch,13_06_stdlib_package_resolution,13_07_search_path_env_override go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `04_05_04_struct_literal_generic_inference`, `04_07_02_alias_generic_substitution`, `04_07_03_alias_scope_visibility_imports`, `04_07_04_alias_methods_impls_interaction`, `04_07_05_alias_recursion_termination`, `04_07_06_alias_reexport_methods_impls`, and `04_07_types_alias_union_generic_combo` to compiler exec fixtures.
- Fixtures: added `06_10_dynamic_metaprogramming_package_object`, `14_02_regex_core_match_streaming`, and `16_01_host_interop_inline_extern` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_05_04_struct_literal_generic_inference,04_07_02_alias_generic_substitution,04_07_03_alias_scope_visibility_imports,04_07_04_alias_methods_impls_interaction,04_07_05_alias_recursion_termination,04_07_06_alias_reexport_methods_impls,04_07_types_alias_union_generic_combo go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_10_dynamic_metaprogramming_package_object go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=14_02_regex_core_match_streaming go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=16_01_host_interop_inline_extern go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added safe-navigation lowering for member access/method calls with nil short-circuiting and argument skip.
- Fixtures: added `06_01_compiler_safe_navigation` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Compiler: added match-expression lowering for simple patterns (wildcard/identifier/literal), plus while-loop and loop-expression lowering with break/continue signals.
- Fixtures: added `06_01_compiler_loops` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: lowered lambda expressions into native function values with capture support for compiled locals.
- Fixtures: added `06_01_compiler_lambda_closure` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added verbose anonymous-function lowering that accepts block bodies with explicit returns.
- Fixtures: added `06_01_compiler_verbose_anonymous_fn` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added range-expression lowering plus for-loop codegen over arrays and iterables.
- Fixtures: added `06_01_compiler_for_loop` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Compiler: expanded match lowering to handle struct/array patterns (including runtime-typed checks) with bindings.
- Fixtures: added `06_01_compiler_match_patterns` exec fixture + coverage index entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Compiler: match lowering now treats `ErrorValue` as struct-like data during pattern matches.
- Fixtures: expanded `06_01_compiler_match_patterns` to cover typed cases and rest bindings.
- Fixtures: added positional struct pattern coverage to `06_01_compiler_match_patterns`.
- Compiler: fixed positional-struct match bindings to keep identifiers in scope for clause bodies.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added raise/rescue lowering plus error-value conversion helpers for compiled code.
- Fixtures: added `06_01_compiler_rescue` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Compiler: added ensure-expression lowering and rethrow statement lowering for compiled code (rescue-aware rethrow).
- Fixtures: added `06_01_compiler_ensure_rethrow` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_raise_error_interface` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: expanded `06_01_compiler_raise_error_interface` to cover nil Error.cause handling.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Tests: `./run_stdlib_tests.sh` in repo root (treewalker + bytecode stdlib suites).
- Tests: `./run_all_tests.sh` in repo root (completed successfully).
- CLI: added `able build` to compile a target or entry file into a native binary (emits Go code + runs `go build`).
- Tests: `go test ./cmd/able -run TestBuildTargetFromManifest -count=1` in `v12/interpreters/go`.
- CLI/Loader: standard loads now skip `.test.able`/`.spec.able` modules unless `--with-tests` is provided; `able build` supports `--with-tests` and routes outputs under `target/test/compiled`.
- Tests: `go test ./cmd/able -run TestRunIgnoresTestModulesUnlessWithTests -count=1` in `v12/interpreters/go`.
- Tests: `go test ./cmd/able -run TestBuildTargetFromManifest -count=1` in `v12/interpreters/go` (after `--with-tests` changes).
- Fixtures: added `06_01_compiler_raise_non_error` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added lowering for propagation (`!`) and `or {}` with runtime error handling for raised interpreter errors.
- Compiler: added raise-signal extraction to panic with runtime values in compiled runtime helpers.
- Fixtures: added `06_01_compiler_or_else` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Fixtures: expanded `06_01_compiler_or_else` to cover error binding on nil and success cases.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: adjusted or-else type resolution to fall back to runtime.Value for mixed branch types.
- Fixtures: added `06_01_compiler_or_else_mixed` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_or_else_struct_mix` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_or_else_error_union` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: allow if/block/assignment expressions to compile in value positions by wrapping tail-expression lowering.
- Fixtures: added `06_01_compiler_if_block_exprs` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added breakpoint/labeled break lowering with breakpoint label tracking in codegen.
- Fixtures: added `06_01_compiler_breakpoint` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added placeholder lambda lowering and implicit member access support for compiled code.
- Fixtures: added `06_01_compiler_placeholder_lambda` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added pipe operator lowering with placeholder/implicit member handling.
- Fixtures: added `06_01_compiler_pipe` exec fixture and coverage entry.
- Interpreter + compiler: prevent placeholder analysis from lifting placeholders across pipe/call boundaries so pipe RHS placeholders don't suppress evaluation.
- Fixtures: updated `06_01_compiler_pipe` expected output for low-precedence pipe assignment.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run Placeholder -count=1` in `v12/interpreters/go`.
- Compiler: added positional struct literal lowering plus positional member access/assignment for compiled structs.
- Fixtures: added `06_01_compiler_struct_positional` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added positional struct pattern support for compiled match patterns on positional structs.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added named struct functional update lowering for compiled code.
- Fixtures: added `06_01_compiler_struct_update` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: expanded typed pattern lowering to support runtime Array/HashMap/DivMod type checks.
- Fixtures: updated `06_01_compiler_match_patterns` to cover typed Array match cases.
- Compiler: fixed positional struct pattern lowering to accept positional field ASTs with field names.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added pattern assignment lowering for struct/array/typed patterns with assignment-time matching and bindings in compiled code.
- Compiler: updated runtime typed-pattern checks for Map/HashMap to accept HashMap struct instances (and HashMapValue) during match/assignment conditions.
- Fixtures: added `06_01_compiler_assignment_patterns` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: expanded `06_01_compiler_match_patterns` to cover typed HashMap match cases.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: expanded `06_01_compiler_assignment_patterns` to cover typed HashMap pattern assignment.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_assignment_pattern_errors` exec fixture plus coverage entry for pattern assignment mismatch errors.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_assignment_pattern_typed_mismatch` exec fixture plus coverage entry for typed pattern assignment mismatches.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_assignment_pattern_rest_mismatch` exec fixture plus coverage entry for rest-pattern assignment mismatch errors.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_assignment_pattern_struct_mismatch` exec fixture plus coverage entry for struct pattern assignment mismatches.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_assignment_pattern_positional_mismatch` exec fixture plus coverage entry for positional struct pattern arity mismatches.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added iterator literal lowering via compiled generator helpers and controller methods.
- Fixtures: added `06_01_compiler_iterator_literal` exec fixture plus coverage entry for iterator literals.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added spawn/await lowering with runtime await keys plus env swapping for compiled lambdas/wrappers.
- Interpreter + bridge: added env-aware CallFunctionIn plus RunCompiledFuture/AwaitIterable and spawn/await bridge helpers.
- Fixtures: added `06_01_compiler_spawn_await` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: stabilize await expression identifiers across compile passes to keep await helpers declared once.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fix: future.await commit now respects async payload so serial executor doesn't deadlock (compiled await on Future handle).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_await_future go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: ensure lowering now rethrows any recovered panic after running ensure, matching tree-walker behavior on non-raise errors.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_ensure_error_passthrough` exec fixture and coverage entry for ensure running on non-raise errors.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_ensure_error_passthrough go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Bytecode: attach runtime context for call-by-name resolution failures to keep diagnostics aligned.
- Fixtures: `06_01_compiler_or_else` nil handler no longer references unbound err (spec-aligned).
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` and `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Tests: `./run_stdlib_tests.sh` (tree-walker + bytecode) in repo root.
- Compiler: lower for loops with pattern matching bindings, emitting runtime pattern mismatch errors on failed destructuring.
- Fixtures: added `06_01_compiler_for_loop_pattern` exec fixture and coverage entry.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_pattern go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_for_loop_pattern_mismatch` exec fixture and coverage entry for for-loop pattern errors.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_for_loop_struct_pattern` exec fixture and coverage entry for for-loop struct patterns.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_struct_pattern go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_for_loop_pattern_guard` exec fixture and coverage entry for guarded match use inside for-loop pattern bodies.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_pattern_guard go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_for_loop_typed_pattern` exec fixture and coverage entry for typed for-loop patterns.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_typed_pattern go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_for_loop_typed_pattern_mismatch` exec fixture and coverage entry for typed for-loop pattern errors.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_typed_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: allow runtime.Value expectations to accept primitive/void expressions (converted to runtime values), and map nullable/result/union types to runtime.Value.
- Fixtures: added `06_01_compiler_nullable_return` exec fixture and coverage entry for nullable return handling.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_nullable_return go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_union_return` exec fixture and coverage entry for union return handling.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_union_return go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_result_return` exec fixture and coverage entry for result return handling.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_result_return go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Bytecode: bind-pattern errors in for loops now attach runtime context to the for-loop node to match tree-walker spans.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_01_compiler_for_loop_pattern_mismatch -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtureParity/06_01_compiler_for_loop_typed_pattern_mismatch -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Compiler: allow expressions targeting runtime.Value to accept primitive/void values in direct compileExpr calls (e.g. union/nullable parameters).
- Fixtures: added `06_01_compiler_union_param` exec fixture and coverage entry for union-typed parameters.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_union_param go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_nullable_param` exec fixture and coverage entry for nullable parameters.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_nullable_param go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Fixtures: added `06_01_compiler_struct_param_bridge` exec fixture and coverage entry for struct param/return conversions.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_struct_param_bridge go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Compiler: boolean contexts now use runtime truthiness via new bridge.IsTruthy helper; logical operators and unary ! honor truthiness for non-bool operands.
- Fixtures: added `06_11_truthiness_boolean_context` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_11_truthiness_boolean_context go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Compiler: if expressions now allow missing else (yielding nil) when the result type is runtime.Value/void, matching truthiness semantics.
- Fixtures: added `08_01_if_truthiness_value` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_01_if_truthiness_value go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12 default) in repo root.
- Compiler: expanded exec fixture coverage to include nullable truthiness, union construction, and literal inference/escaping scenarios.
- Fixtures: added `04_06_02_nullable_truthiness`, `04_06_03_union_construction_result_option`, `06_01_literals_array_map_inference`, `06_01_literals_numeric_contextual`, `06_01_literals_string_char_escape` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_08_array_ops_mutability` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_08_array_ops_mutability go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `11_02_option_result_or_handlers` and `11_02_option_result_propagation` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_02_option_result_or_handlers go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_02_option_result_propagation go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_03_safe_navigation_nil_short_circuit`, `06_04_function_call_eval_order_trailing_lambda`, and `06_06_string_interpolation` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_03_safe_navigation_nil_short_circuit go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_04_function_call_eval_order_trailing_lambda go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_06_string_interpolation go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_03_cast_error_payload_recovery` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_03_cast_error_payload_recovery go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: treat `^` as exponentiation (bitwise XOR remains `.^`) and route exponent through runtime binary op handling.
- Fixtures: added `06_03_operator_precedence_associativity` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_03_operator_precedence_associativity go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: route unary operator fallback through runtime dispatch for non-primitive operands, and allow non-primitive binary operators/comparisons to use runtime operator interfaces.
- Bridge: added unary operator dispatch helper for compiled code.
- Fixtures: added `06_03_operator_overloading_interfaces` and `14_01_operator_interfaces_arithmetic_comparison` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_03_operator_overloading_interfaces go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=14_01_operator_interfaces_arithmetic_comparison go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: avoid unused temps for empty struct pattern bindings and treat `IteratorEnd` sentinel values as matching `IteratorEnd {}` struct patterns.
- Fixtures: added `14_01_language_interfaces_index_apply_iterable` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=14_01_language_interfaces_index_apply_iterable go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `10_01_interface_defaults_composites`, `10_02_impl_specificity_named_overrides`, `10_03_interface_type_dynamic_dispatch`, `10_05_interface_named_impl_defaults`, and `10_06_interface_generic_param_dispatch` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_01_interface_defaults_composites go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_02_impl_specificity_named_overrides go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_05_interface_named_impl_defaults go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_06_interface_generic_param_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `10_02_impl_where_clause`, `10_07_interface_default_chain`, `10_08_interface_default_override`, `10_09_interface_named_impl_inherent`, `10_10_interface_inheritance_defaults`, `10_11_interface_generic_args_dispatch`, `10_12_interface_union_target_dispatch`, `10_13_interface_param_generic_args`, `10_14_interface_return_generic_args`, and `10_16_interface_value_storage` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_02_impl_where_clause,10_07_interface_default_chain,10_08_interface_default_override,10_09_interface_named_impl_inherent,10_10_interface_inheritance_defaults,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_13_interface_param_generic_args,10_14_interface_return_generic_args,10_16_interface_value_storage go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `12_01_bytecode_spawn_basic`, `12_01_bytecode_await_default`, `12_02_async_spawn_combo`, `12_03_spawn_future_status_error`, and `12_04_future_handle_value_view` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_01_bytecode_spawn_basic go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_01_bytecode_await_default go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_02_async_spawn_combo go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_03_spawn_future_status_error go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_04_future_handle_value_view go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler runtime: compiled async tasks now cooperatively yield by running another queued task when `future_yield()` is called, avoiding hangs without resumable compiled frames.
- Compiler: compiled lambdas now recover raised errors from interpreter calls (while rethrowing break/continue signals) so rescue works when interpreted code invokes compiled lambdas.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_05_concurrency_channel_ping_pong,12_05_mutex_lock_unlock,12_06_await_fairness_cancellation,12_07_channel_mutex_error_types go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_12_01_stdlib_string_helpers`, `06_12_02_stdlib_array_helpers`, and `06_12_03_stdlib_numeric_ratio_divmod` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_01_stdlib_string_helpers,06_12_02_stdlib_array_helpers,06_12_03_stdlib_numeric_ratio_divmod go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added diagnostic node registration and call-frame tracking so compiled runtime errors (division/shift) report source locations and call sites.
- Fixtures: added `06_01_compiler_division_by_zero`, `06_01_compiler_shift_out_of_range`, and `15_02_entry_args_signature` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=15_02_entry_args_signature,06_01_compiler_division_by_zero,06_01_compiler_shift_out_of_range go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `08_02_loop_expression_break_value`, `08_02_numeric_sum_loop`, `08_02_range_inclusive_exclusive`, `08_02_while_continue_break`, and `08_03_breakpoint_nonlocal_jump` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_02_loop_expression_break_value,08_02_numeric_sum_loop,08_02_range_inclusive_exclusive,08_02_while_continue_break,08_03_breakpoint_nonlocal_jump go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: attach runtime context to non-exhaustive match failures in compiled code so diagnostics include source locations.
- Fixtures: added `08_01_control_flow_fizzbuzz`, `08_01_match_guards_exhaustiveness`, and `08_01_union_match_basic` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_01_control_flow_fizzbuzz,08_01_match_guards_exhaustiveness,08_01_union_match_basic go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: represent structs as pointers in compiled Go code; update struct literal lowering and struct <-> runtime conversions to use pointer semantics and handle nil.
- Compiler: allow pattern assignments to convert between runtime.Value and typed bindings (both directions), and suppress unused pattern binding vars.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_00_mutability_declaration_vs_assignment,05_02_array_nested_patterns,05_02_identifier_wildcard_typed_patterns,05_02_struct_pattern_rename_typed,05_03_assignment_evaluation_order go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Fixtures: added `04_05_01_struct_singleton_usage`, `04_05_02_struct_named_update_mutation`, `04_05_03_struct_positional_named_tuple`, `04_06_01_union_payload_patterns`, `04_06_04_union_guarded_match_exhaustive`, and `07_02_lambdas_closures_capture` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_05_01_struct_singleton_usage,04_05_02_struct_named_update_mutation,04_05_03_struct_positional_named_tuple,04_06_01_union_payload_patterns,04_06_04_union_guarded_match_exhaustive,07_02_lambdas_closures_capture go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `07_04_apply_callable_interface` and `07_04_trailing_lambda_method_syntax` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_04_apply_callable_interface,07_04_trailing_lambda_method_syntax go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: sync struct mutations after dynamic member calls by converting the runtime receiver back into the compiled struct pointer.
- Fixtures: added `07_06_shorthand_member_placeholder_lambdas` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_06_shorthand_member_placeholder_lambdas go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: allow missing nullable tail args in compiled calls/wrappers by injecting nil; attach runtime context for compiled call errors; populate call-node callees so overload errors report names.
- Fixtures: added `07_07_overload_resolution_runtime` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_07_overload_resolution_runtime go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `07_02_01_verbose_anonymous_fn` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_02_01_verbose_anonymous_fn go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `07_05_partial_application` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_05_partial_application go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `07_01_function_definition_generics_inference` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_01_function_definition_generics_inference go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `07_08_return_context_generic_call_inference` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_08_return_context_generic_call_inference go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: allow non-local returns inside nested blocks by lowering return statements to a compiled return signal and recovering at function boundaries.
- Fixtures: added `07_03_explicit_return_flow` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_03_explicit_return_flow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: treat wrapped generator stop errors as iterator completion in compiled iterator helpers (avoids stop bubbling through runtime diagnostics after call-context wrapping).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_iterator_literal go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Compiler/Runtime: compiled futures now use a resumable yield handshake under the serial executor so future_yield requeues compiled tasks fairly (channel-based resume/yield).
- Fixtures: added `12_02_future_fairness_cancellation` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_02_future_fairness_cancellation go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler harness: honor fixture manifest executor selection (serial vs goroutine) and add `12_08_blocking_io_concurrency` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` (v12) completed successfully.
- Compiler: raise/or-else/rescue now handle raised errors as error panics with runtime context; rescue/or-else recover both runtime.Value and raiseSignal errors; rethrow preserves original error when available.
- Compiler helpers: division-by-zero and shift-out-of-range now raise structured errors via RaiseWithContext; __able_panic_on_error now panics errors directly to keep diagnostics.
- Fixtures: added `06_07_generator_yield_iterator_end`, `06_07_iterator_pipeline`, `11_00_errors_match_loop_combo`, `11_03_raise_exit_unhandled`, `11_03_rescue_ensure`, `11_03_rescue_rethrow_standard_errors` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_result_return,06_01_compiler_or_else,11_02_option_result_propagation,06_07_generator_yield_iterator_end,06_07_iterator_pipeline,11_00_errors_match_loop_combo,11_03_raise_exit_unhandled,11_03_rescue_ensure,11_03_rescue_rethrow_standard_errors go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `14_02_hash_eq_primitives`, `14_02_hash_eq_float`, and `14_02_hash_eq_custom` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=14_02_hash_eq_primitives,14_02_hash_eq_float,14_02_hash_eq_custom go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `15_04_background_work_flush` to compiler exec fixtures.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=15_04_background_work_flush go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Runtime: UFCS resolution now accepts native functions by binding them with adjusted arity, so compiled native functions can participate in UFCS method syntax.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=09_04_methods_ufcs_basics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=09_00_methods_generics_imports_combo,09_02_methods_instance_vs_static,09_05_method_set_generics_where go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: lower struct literals to runtime values for unknown/unsupported types, allow functional updates, shorthand field initializers, and singleton positional literals; add runtime struct literal lowering with validation and type-arg handling.
- Compiler: allow generic calls by routing them through dynamic call helpers; add yield statement lowering for iterator literals.
- Compiler: relax block/lambda/match lowering for void contexts (empty blocks, raise/rethrow in blocks, mixed match clause types), add zero-value helper for unreachable branches, and ignore void lambda return mismatches; ensure lambda params are marked used; avoid wrapper arg name collisions on `result`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_bytecode_map_spread,06_07_iterator_pipeline,07_09_bytecode_iterator_yield,08_01_bytecode_if_indexing,14_02_regex_core_match_streaming,11_00_errors_match_loop_combo,06_01_compiler_or_else,06_01_compiler_result_return go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: allow top-level return statements to short-circuit compiled function/lambda bodies, ignoring trailing unreachable statements.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_01_return_statement_type_enforcement go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Interpreters: pattern assignment expressions now return `Error` values on mismatch without partial bindings; bytecode VM uses shared pattern assignment path.
- Compiler: non-simple or runtime-typed pattern assignments now fall back to interpreter to preserve error-value semantics; updated compiler exec fixtures to assert error values.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestInterfaceAssignmentMissingImplementation -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_assignment_pattern_errors,06_01_compiler_assignment_pattern_positional_mismatch,06_01_compiler_assignment_pattern_rest_mismatch,06_01_compiler_assignment_pattern_struct_mismatch,06_01_compiler_assignment_pattern_typed_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Interpreters: for-loop pattern mismatches now yield `Error` values and halt iteration without raising; loop bindings continue to shadow outer scopes.
- Compiler: for-loop pattern mismatches no longer panic; loops break on mismatch to preserve error-value behavior.
- Spec: typed pattern mismatch language now aligns with error-value semantics.
- Fixtures: updated for-loop mismatch exec fixtures to assert error values and continued execution.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_pattern_mismatch,06_01_compiler_for_loop_typed_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestBytecodeVM_ForLoopArraySum -count=1` in `v12/interpreters/go`.
- Fixtures: added block-expression coverage for for-loop pattern mismatches; updated spec for loop evaluation semantics.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Spec: clarify cast failures raise runtime exceptions; align loop result wording to void on normal completion (break supplies value, for-loop pattern mismatch yields Error); update while loop example result text.
- Compiler: pattern assignments now compile with runtime pattern checks and error values; typed pattern matching/binding uses match-type coercion (not cast) for runtime values; __able_try_cast now uses MatchType helper.
- Bridge/Interpreter: added MatchType helper plus exported MatchesType/CoerceValueToType to support compiler pattern matching.
- Audit: compiler fallbacks reduced to 4 (only overload-resolution cases, lambda generics, and float literal type in hash fixture remain).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_assignment_pattern_errors,06_01_compiler_assignment_pattern_typed_mismatch,06_01_compiler_assignment_pattern_rest_mismatch,06_01_compiler_assignment_pattern_struct_mismatch,06_01_compiler_assignment_pattern_positional_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_match_patterns go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: generate overload dispatchers for compiled functions (runtime type matching + specificity scoring, optional-arg penalty, ambiguity errors), register overload wrappers, and route calls through compiled overload helpers with runtime-context diagnostics.
- Compiler: allow verbose anonymous function generics/where clauses by inserting generic constraint checks via MatchType; lambda generics no longer force interpreter fallback.
- Compiler: untyped mixed numeric literals now prefer float types; float division no longer raises division-by-zero (IEEE semantics).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_07_overload_resolution_runtime,14_02_hash_eq_float go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_02_01_verbose_anonymous_fn go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Typechecker: allow casts to interface types (and nullable interface types) when the source type implements the interface.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_03_cast_semantics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: assignment statements now discard assignment expression results to avoid unused temps when pattern assignments synthesize temps.
- Compiler: match patterns treat singleton struct identifiers as literal matches (non-binding) when not shadowed, aligning singleton enum-like matching and union payload patterns.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_03_type_expression_syntax,04_05_01_struct_singleton_usage,04_05_03_struct_positional_named_tuple,04_06_01_union_payload_patterns,04_06_02_nullable_truthiness,04_06_03_union_construction_result_option,04_06_04_union_guarded_match_exhaustive,04_07_02_alias_generic_substitution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Bytecode VM: for-loop pattern binding now uses the for-loop mismatch semantics (ErrorValue + loop termination) instead of raising.
- Fixtures: array/struct/typed assignment mismatch AST fixtures now expect Error values; struct pattern missing-field/type mismatch fixtures expect typechecker diagnostics instead of runtime errors.
- Interpreter: type info for struct instances now infers generic type arguments when missing/wildcard, so compiled values without TypeArguments still match generic impls.
- Compiler: avoid Go identifier collisions with imported packages by reserving names (`fmt`, `runtime`, `ast`, `bridge`, `interpreter`, `errors`, `sync`).
- Compiler: interface-typed parameters and returns now coerce via `MatchType` in wrappers, and compiled direct calls coerce interface args before invoking compiled functions; interface-typed returns now wrap interface values in compiled returns.
- Interpreter: compiled await in serial executor now yields via compiled resume channels instead of surfacing `errSerialYield` to compiled code.
- Interpreter: `ApplyBinaryOperator` now mirrors string concatenation semantics for `+` (string+string) to match tree-walker behavior.
- Compiler: match pattern code now reuses field temps for nested struct checks to avoid unused locals; rescue lowering now marks recovered subject temp used.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_13_interface_param_generic_args,10_14_interface_return_generic_args go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `go test ./pkg/interpreter -run TestExecFixtures/12_06_await_fairness_cancellation -count=1` in `v12/interpreters/go`.
- Tests: compiler exec fixtures run in batches (batches 9–14) via `ABLE_COMPILER_EXEC_FIXTURES=...` in `v12/interpreters/go` (all passing after fixes).
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/parser -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_assignment_pattern_errors,06_01_compiler_assignment_pattern_positional_mismatch,06_01_compiler_assignment_pattern_rest_mismatch,06_01_compiler_assignment_pattern_struct_mismatch,06_01_compiler_assignment_pattern_typed_mismatch,06_01_compiler_for_loop_pattern_mismatch,06_01_compiler_for_loop_typed_pattern_mismatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Workspace: moved `v12/interpreters/go/tmp_saved` to `/tmp/able_tmp_saved` after cleaning compiler fixture temp output.
- Compiler/Runtime: compiled future cancellation now eagerly marks the handle cancelled when a compiled task completes with a cancel request, preventing pending status leaks under serial scheduling.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures/12_02_future_fairness_cancellation -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/v12/interpreters/go/.gocache ./run_all_tests.sh` at repo root.
- Workspace: moved `v12/interpreters/go/tmp` to `/tmp/ablec_tmp` after test run artifacts.
- Tests: `GOCACHE=$(pwd)/v12/interpreters/go/.gocache ./run_stdlib_tests.sh` at repo root.
- Audit: compiler fallback audit over the compiler exec fixture list reports 0 remaining fallbacks (script: `/tmp/ablec_fallback_audit.go`).
- Compiler: collect `methods` definitions for entry module, compile eligible method bodies, and emit direct compiled calls for static/type-qualified and instance methods when the receiver type is known (falls back to dynamic resolution otherwise).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_method_call,06_01_compiler_type_qualified_method,06_01_compiler_bound_method_value,09_02_methods_instance_vs_static,09_04_methods_ufcs_basics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler/Interpreter: compiled method thunks register into inherent method pools (via new interpreter `RegisterCompiledMethod` and `CompiledThunk`), so dynamic member dispatch can execute compiled method bodies while preserving method constraints/visibility checks.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_method_call,06_01_compiler_type_qualified_method,06_01_compiler_bound_method_value,09_02_methods_instance_vs_static,09_04_methods_ufcs_basics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler/Interpreter: compiled method thunk registration now matches overloads by signature (target type + param types), including implicit self for method shorthand, to avoid ambiguous method bindings.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_method_call,06_01_compiler_type_qualified_method,06_01_compiler_bound_method_value,09_02_methods_instance_vs_static,09_04_methods_ufcs_basics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_stdlib_tests.sh` (tree-walker + bytecode) at repo root.
- Compiler/Interpreter: compiled method registration now tolerates generic target mismatches (base name + generic params) and applies thunks to all matching overloads to avoid failures when kernel/fixture method sets overlap.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_05_concurrency_channel_ping_pong,12_06_await_fairness_cancellation,15_04_background_work_flush go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` at repo root.
- Tooling: `run_all_tests.sh` now reuses a shared Go build cache by default (set `ABLE_GOCACHE=tmp` for isolated runs) and propagates `ABLE_COMPILER_EXEC_GOCACHE`.
- Compiler tests: `go build` invocations now respect `ABLE_COMPILER_EXEC_GOCACHE`/`GOCACHE` for fixture builds.
- Tests: `go test ./pkg/compiler -run TestCompilerExecHarness -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go` (completed in ~211s; exceeds the 1-minute guideline).
- CLI: `able build` default output now roots under the current working directory (`./target/compiled`), even when building a file in another directory.
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- CLI: added `v12/able` wrapper for the main CLI; it passes `ABLE_BUILD_ROOT` so build outputs root at the caller's working directory.
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- CLI: skip `.gomodcache`/`.modcache` when copying the interpreter module for builds outside the module root.
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- Compiler: generated `main.go` now discovers search paths (stdlib/kernel/ABLE_PATH) and registers `print`, matching CLI behavior for compiled binaries.
- Tests: `go test ./pkg/compiler -run TestCompilerExecHarness -count=1` in `v12/interpreters/go`.
- Docs: added `v12/design/compiler-aot.md` with correctness-first compiler vision and full work breakdown; updated `PLAN.md` compiler AOT queue and flattened nested TODO bullets.
- Spec: added compiled execution boundary semantics in `spec/full_spec_v12.md` and added compiler AOT gaps in `spec/TODO_v12.md`; removed completed compiler vision item from `PLAN.md`.
- Compiler AOT: added program analysis with module dependency graph and preserved typechecker outputs (`program_analysis.go`) plus tests; removed completed items from `PLAN.md`.
- Tests: `go test ./pkg/compiler -run TestAnalyzeProgramBuildsGraphAndTypecheck -count=1` in `v12/interpreters/go`.
- Audit: compiler fallback audit over compiler exec fixture list reports 1 fallback (`04_05_02_struct_named_update_mutation_diag: main (identifier type mismatch)`).
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerFallbackAudit -count=1 -v` in `v12/interpreters/go`.
- Audit: compiler fallback audit excluding fixtures with expected typecheck diagnostics reports 0 fallbacks (ad-hoc script run).
- Compiler: struct literal functional updates now fall back to runtime lowering when update sources can't be compiled to the target struct type, avoiding identifier type mismatch fallbacks while preserving runtime validation for mismatched sources.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=04_05_02_struct_named_update_mutation_diag go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Audit: compiler fallback audit excluding fixtures with expected typecheck diagnostics reports 0 fallbacks (ad-hoc script run).
- Compiler bridge: accept `String` structs in `AsString` and allow integer-to-float coercions in `AsFloat` (with interface unwrapping for primitive conversions).
- Tests: `go test ./pkg/compiler/bridge -run TestAs -count=1` in `v12/interpreters/go`.
- Compiler: added a fallback-audit test for compiler exec fixtures (guarded by `ABLE_COMPILER_FALLBACK_AUDIT`) to surface uncompiled functions.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=15_01_program_entry_hello_world go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1` in `v12/interpreters/go`.
- Audit: compiler fallback audit across all exec fixtures reports 0 fallbacks (using `ABLE_COMPILER_EXEC_FIXTURES=all`).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1` in `v12/interpreters/go`.
- Compiler: compiled integer `+`/`-`/`*` now emit overflow-checked helpers (signed/unsigned, width-aware) and overflow errors attach runtime diagnostics context.
- Tests: `go test ./pkg/compiler -run TestDetectDynamicFeatures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_integer_overflow` to cover compiled integer overflow diagnostics.
- Fixtures: added `06_01_compiler_integer_overflow_sub` and `06_01_compiler_integer_overflow_mul` to cover compiled `-`/`*` overflow diagnostics.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_integer_overflow,06_01_compiler_integer_overflow_sub,06_01_compiler_integer_overflow_mul go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: compound assignments now pass diagnostic node context into compiled numeric helpers (division/overflow/shift) instead of using `nil`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_compound_assignment go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Fixtures: added `06_01_compiler_compound_assignment_overflow` to cover overflow errors on `+=` in compiled output.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_compound_assignment_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler/Interpreter: unary integer negation now checks overflow; negating the minimum value raises `OverflowError`.
- Fixtures: added `06_01_compiler_unary_overflow` for unary negation overflow diagnostics.
- Spec: documented unary negation overflow in `spec/full_spec_v12.md`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_unary_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `go test ./pkg/interpreter -run TestBytecodeVM_UnaryNegate -count=1` in `v12/interpreters/go`.
- Compiler: compiled `//`/`%` overflow now checks signed bounds (including `min_int // -1`) and raises `OverflowError`.
- Fixtures: added `06_01_compiler_divmod_overflow` for compiled division overflow diagnostics.
- Spec: documented `//`/`%` overflow behavior in `spec/full_spec_v12.md`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_divmod_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: `//`/`%` helpers now enforce overflow for signed/unsigned bounds (including `min_int // -1`) and `^` uses shared float pow helpers to keep imports consistent.
- Fixtures: added `06_01_compiler_divmod_overflow`, `06_01_compiler_pow_overflow`, and `06_01_compiler_pow_negative_exponent`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_pow_overflow,06_01_compiler_pow_negative_exponent go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_divmod go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: `/%` now lowers to compiled divmod helpers and constructs `DivMod` results without interpreter operator dispatch (falls back to placeholder struct definition if missing).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_divmod,06_01_compiler_divmod_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: IR iterator helper now enforces re-entrancy errors (`iterator.next re-entered while suspended at yield`) to match spec/interpreter semantics.
- Tests: `go test ./pkg/compiler -run TestIREmitFunctionIteratorLiteral -count=1` in `v12/interpreters/go`.
- Compiler: iterator helper in AOT generator now enforces re-entrancy errors to avoid deadlocks and match spec semantics.
- Fixtures: added exec fixture `07_10_iterator_reentrancy` for iterator re-entrancy errors.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=07_10_iterator_reentrancy go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: build output `main` now formats runtime errors via `DescribeRuntimeDiagnostic` (restores location/notes in compiled binaries).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_iterator_literal,06_07_generator_yield_iterator_end,06_07_iterator_pipeline,07_07_bytecode_implicit_iterator,07_09_bytecode_iterator_yield,07_10_iterator_reentrancy go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler IR: assignment lowering now implicitly declares missing bindings for `=` (matches spec), with pattern assignments only declaring when no existing binding is found.
- Tests: `go test ./pkg/compiler -run TestLowerAssignmentAllowsImplicitBinding -count=1` in `v12/interpreters/go`.
- Compiler IR: pattern destructure bindings now flow through IRDestructure bindings map, avoiding missing slot errors for `=` reassignments to existing bindings.
- Tests: `go test ./pkg/compiler -run TestIREmitPatternAssignmentExistingBinding -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_00_mutability_declaration_vs_assignment,06_01_compiler_assignment_patterns,06_01_compiler_assignment_pattern_errors,06_01_compiler_assignment_pattern_typed_mismatch,06_01_compiler_assignment_pattern_rest_mismatch,06_01_compiler_assignment_pattern_struct_mismatch,06_01_compiler_assignment_pattern_positional_mismatch,05_02_array_nested_patterns,05_02_identifier_wildcard_typed_patterns,05_02_struct_pattern_rename_typed go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_match_patterns,08_01_match_guards_exhaustiveness,08_01_union_match_basic,11_03_bytecode_rescue_basic,11_03_rescue_ensure,11_03_raise_exit_unhandled go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_for_loop,06_01_compiler_for_loop_pattern,06_01_compiler_for_loop_pattern_mismatch,06_01_compiler_for_loop_struct_pattern,06_01_compiler_for_loop_pattern_guard,06_01_compiler_for_loop_typed_pattern,06_01_compiler_for_loop_typed_pattern_mismatch,08_02_loop_expression_break_value,08_02_numeric_sum_loop,08_02_while_continue_break go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: added dynamic feature warnings during compile to flag modules using dynimport/dynamic calls.
- Tests: `go test ./pkg/compiler -run TestDetectDynamicFeatures -count=1` in `v12/interpreters/go`.
- Compiler: spawn lowering now captures visible slots, builds a nested IR function body, and emits `IRSpawn` with explicit error flow (codegen still unsupported).
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler: IR codegen now emits `IRSpawn` by generating a task closure and calling `bridge.Spawn`; IR emission now includes nested spawn body functions and memoized IR function names.
- Tests: `go test ./pkg/compiler -run TestIREmitFunctionSpawn -count=1` in `v12/interpreters/go`.
- Compiler: IR codegen now treats slots as mutable cells (`*runtime.Value`), loading via `__able_cell_value` and storing via `*slot` to preserve by-reference capture semantics for spawn bodies.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler: spawn bodies now capture slots by reference (captured slots passed as `*runtime.Value` into spawn IR functions); IR await codegen now calls `bridge.Await` with a per-site `ast.AwaitExpression`.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler: split IR Go codegen helpers into `ir_codegen_helpers.go` to keep files under 1000 lines.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler: iterator literals now lower into nested IR functions with by-reference captured slots; IR codegen emits iterator helpers and compiles iterator bodies via `__able_new_iterator`.
- Compiler: added IR await and iterator literal codegen tests.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: added IR-to-Go emission scaffold with helper runtime shims and minimal instruction coverage (compute/invoke/branch/iter next), plus parser-based smoke tests for emitted Go.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR codegen now emits array/struct/map literals, string interpolation, and limited destructuring; added literal-focused IR emit tests and fixed lowering to treat array/map literals as non-const IR nodes.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: expanded IR destructuring codegen to handle struct/array/nested patterns with typed/literal checks and rest binding temps; added error-to-struct matching helper for parity with interpreter pattern semantics.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR struct literal lowering/codegen now carries explicit type arguments (including functional update inheritance).
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR codegen now emits casts via bridge.Cast; added a cast-focused IR emit test.
- Tests: `go test ./pkg/compiler -run TestIREmitFunction -count=1` in `v12/interpreters/go`.
- Compiler AOT: dynamic feature detection now scans module/function bodies (dynimport, dyn member calls, control-flow blocks) and reports per-function usage.
- Tests: `go test ./pkg/compiler -run TestDetectDynamicFeatures -count=1` in `v12/interpreters/go`.
- Compiler AOT: added typed IR scaffolding (ANF blocks, instructions, explicit error flow terminators) with block invariant tests.
- Tests: `go test ./pkg/compiler -run TestIRBlock -count=1` in `v12/interpreters/go`.
- Compiler AOT: added initial AST-to-IR lowerer for control flow (`if`, `match`), assignments, and basic expressions.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Typechecker: program checks now expose per-package inferred type maps for compiler use.
- Compiler: lowering can consume `ProgramAnalysis` inference map; program analysis test asserts inference presence.
- Tests: `go test ./pkg/compiler -run TestAnalyzeProgramBuildsGraphAndTypecheck -count=1` in `v12/interpreters/go`.
- Tests: `go test ./pkg/typechecker -run TestProgramCheckerResolvesDependencies -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now handles `while`, `loop`, `break`, and `continue`, and skips writes on terminated control-flow paths.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now covers `for` loops with iterator steps and pattern destructuring, using `IRIterNext`.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now handles `raise`, `rescue`, `ensure`, `or`, and `!` (propagation), with explicit error handler routing.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now supports `spawn`, `await`, and `breakpoint`, plus global identifier references.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now handles casts, ranges, collections (array/map/struct), iterators, and string interpolation (scaffolded).
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR lowering now emits explicit literal nodes for arrays/maps/structs/string interpolation and iterator literals.
- Tests: `go test ./pkg/compiler -run TestLowerLiteralExpressions -count=1` in `v12/interpreters/go`.
- Compiler AOT: added basic IR validation (terminator presence) and enabled it across lowerer tests.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Compiler AOT: IR validation now checks reachability and missing block references.
- Tests: `go test ./pkg/compiler -run TestLower -count=1` in `v12/interpreters/go`.
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- Tests: `./run_all_tests.sh` at repo root.
- Compiler: collect structs/functions/overloads across all modules, emit per-package overload dispatchers, and register compiled function thunks by signature instead of overwriting env bindings (uses qualified original names for fallback calls).
- Interpreter: track per-package environments (new `PackageEnvironment`) and register compiled function overloads by param signature on existing runtime function values.
- CLI: `ablec` now supports `-build` (forces `-pkg=main`) and `-bin` to build a native binary after emitting Go code.
- Tests: `go test ./pkg/compiler -run TestCompilerEmitsStructsAndWrappers -count=1` in `v12/interpreters/go`.
- Tests: `go test ./pkg/interpreter -run TestInterpreterEvaluateProgramSuccess -count=1` in `v12/interpreters/go`.
- Interpreter: compiled member assignment now supports array receivers (length/capacity/storage_handle/index updates) to keep compiled stdlib methods from failing on runtime arrays.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `go test ./pkg/interpreter -run TestInterpreterEvaluateProgramSuccess -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go` (timed out at 120s).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch,13_06_stdlib_package_resolution,13_07_search_path_env_override go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_01_interface_defaults_composites,10_02_impl_specificity_named_overrides,10_02_impl_where_clause,10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_05_interface_named_impl_defaults,10_06_interface_generic_param_dispatch,10_07_interface_default_chain go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_08_interface_default_override,10_09_interface_named_impl_inherent,10_10_interface_inheritance_defaults,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_13_interface_param_generic_args,10_14_interface_return_generic_args,10_15_interface_default_generic_method,10_16_interface_value_storage go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Interpreter: compiled thunks now enter the serial executor's synchronous section when invoked from non-async contexts, preserving deterministic spawn ordering for compiled code.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_02_async_spawn_combo go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_01_bytecode_spawn_basic,12_01_bytecode_await_default,12_02_async_spawn_combo,12_02_future_fairness_cancellation,12_03_spawn_future_status_error,12_04_future_handle_value_view,12_05_concurrency_channel_ping_pong,12_05_mutex_lock_unlock,12_06_await_fairness_cancellation,12_07_channel_mutex_error_types,12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- CLI: `ablec` now loads programs with the same search path discovery (entry dir + `ABLE_PATH`/`ABLE_MODULE_PATHS` + stdlib/kernel roots) as the main CLI.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./cmd/ablec -count=1` in `v12/interpreters/go`.
- Compiler: compiled functions/methods now swap to their package environments; compiled struct method wrappers reapply mutated receivers via generated `__able_struct_<Name>_apply` helpers.
- Interpreter: `matchesType` now accepts array struct instances by converting to `ArrayValue` for type matching.
- Compiler: struct conversion rendering now scopes temp declarations to avoid redeclare errors.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_04_future_handle_value_view,12_05_concurrency_channel_ping_pong,12_05_mutex_lock_unlock,12_06_await_fairness_cancellation,12_07_channel_mutex_error_types,12_08_blocking_io_concurrency,06_11_truthiness_boolean_context,06_12_01_stdlib_string_helpers,06_12_02_stdlib_array_helpers,06_12_03_stdlib_numeric_ratio_divmod,08_01_if_truthiness_value,08_01_control_flow_fizzbuzz,08_01_bytecode_if_indexing,08_01_bytecode_match_basic,08_01_bytecode_match_subject,08_01_match_guards_exhaustiveness,08_01_union_match_basic,08_02_bytecode_loop_basics,08_02_loop_expression_break_value,08_02_numeric_sum_loop go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_02_range_inclusive_exclusive,08_02_while_continue_break,08_03_breakpoint_nonlocal_jump,05_00_mutability_declaration_vs_assignment go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_02_array_nested_patterns,05_02_identifier_wildcard_typed_patterns,05_02_struct_pattern_rename_typed,05_03_assignment_evaluation_order,05_03_bytecode_assignment_patterns,11_02_option_result_or_handlers,11_02_option_result_propagation,11_02_bytecode_or_else_basic go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_00_errors_match_loop_combo,11_03_bytecode_ensure_basic,11_03_bytecode_rescue_basic,11_03_raise_exit_unhandled,11_03_rescue_ensure,11_03_rescue_rethrow_standard_errors go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go` (completed in ~205s; exceeds the 1-minute guideline).
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` in `v12/interpreters/go`.
- CLI: `able test --compiled` now writes temp build output under the Go module root (optional `ABLE_TEST_KEEP_WORKDIR`), loads test packages via `IncludePackages`, and avoids invalid imports in the runner source.
- Compiler: wrapper arg/return interface MatchType checks now skip when unbound generic params are present (prevents false runtime mismatches like `Expectation.to`).
- Tests: `go test ./cmd/able -run TestTestCommandCompiledRuns -count=1` in `v12/interpreters/go`.
- CLI: `able build` and `ablec` now emit a `go.mod` into build outputs with a local `replace` to the interpreter module (self-contained builds outside the module tree).
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- Tests: `go test ./cmd/ablec -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go` (completed in ~211s; exceeds the 1-minute guideline).
- Tests: `./run_all_tests.sh` at repo root.
- CLI: `able build`/`ablec -build` now copy `v12/interpreters/go` plus parser sources into build outputs and wire `go.mod` `replace` to `./v12/interpreters/go` for self-contained binaries.
- Tests: `go test ./cmd/ablec -count=1` in `v12/interpreters/go`.
- Tests: `go test ./cmd/able -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go` (completed in ~211s; exceeds the 1-minute guideline).
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=all go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1` in `v12/interpreters/go`.
- Compiler: compiled runtime helpers now fast-path array indexing/assignment and array element access for handle-0 arrays, including index coercion + IndexError payload construction for direct array access.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_08_array_ops_mutability,06_12_02_stdlib_array_helpers go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler: split `generator_render.go` into smaller render files (`generator_render_runtime.go`, `generator_render_structs.go`, `generator_render_functions.go`, `generator_render_main.go`, `generator_render_helpers.go`) to keep files under 1000 lines.
- Tests: `go test ./pkg/compiler -run TestCompilerEmitsStructsAndWrappers -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_integer_overflow,06_01_compiler_integer_overflow_sub,06_01_compiler_integer_overflow_mul,06_01_compiler_unary_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_divmod_overflow,06_01_compiler_pow_overflow,06_01_compiler_pow_negative_exponent,06_01_compiler_compound_assignment_overflow go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_divmod,06_01_compiler_array_struct_literal,06_01_literals_array_map_inference go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_08_array_ops_mutability,06_12_02_stdlib_array_helpers go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_12_01_stdlib_string_helpers,06_12_03_stdlib_numeric_ratio_divmod go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_11_truthiness_boolean_context,08_01_if_truthiness_value,08_01_control_flow_fizzbuzz go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_01_bytecode_if_indexing,08_01_bytecode_match_basic,08_01_bytecode_match_subject go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_01_match_guards_exhaustiveness,08_01_union_match_basic,08_02_bytecode_loop_basics go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_02_loop_expression_break_value,08_02_numeric_sum_loop,08_02_range_inclusive_exclusive go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=08_02_while_continue_break,08_03_breakpoint_nonlocal_jump,05_00_mutability_declaration_vs_assignment go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_02_array_nested_patterns,05_02_identifier_wildcard_typed_patterns,05_02_struct_pattern_rename_typed go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=05_03_assignment_evaluation_order,05_03_bytecode_assignment_patterns,11_02_option_result_or_handlers go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_02_option_result_propagation,11_02_bytecode_or_else_basic,11_00_errors_match_loop_combo go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_03_bytecode_ensure_basic,11_03_bytecode_rescue_basic,11_03_raise_exit_unhandled go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=11_03_rescue_ensure,11_03_rescue_rethrow_standard_errors,12_02_async_spawn_combo go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_01_bytecode_spawn_basic,12_01_bytecode_await_default,12_02_future_fairness_cancellation go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_03_spawn_future_status_error,12_04_future_handle_value_view,12_05_concurrency_channel_ping_pong go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_05_mutex_lock_unlock,12_06_await_fairness_cancellation,12_07_channel_mutex_error_types go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Compiler AOT: enumerated remaining unnamed impl fallback blockers in stdlib-heavy fixtures; fixed compileability for `IndexMut Array<T>.set` and String iterator unnamed impl methods (`StringBytesIter`, `StringCharsIter`, `StringGraphemesIter`) to reduce strict-dispatch blockers.
- Compiler AOT: `Index Array<T>.get` now lowers to a compileable direct slot-read path; fallback audit for `13_06_stdlib_package_resolution`, `12_08_blocking_io_concurrency`, and `06_08_array_ops_mutability` no longer reports unnamed impl method fallbacks.
- Compiler AOT: impl-method param-type rendering now substitutes interface generic arguments before registration, and synthetic default impl methods are skipped for interpreter thunk registration (they are still available in compiled interface-dispatch tables).
- Compiler AOT: deferred strict interface dispatch when impl thunk registration mismatches are detected during registration (`__able_interface_dispatch_blocked`) to avoid unsafe strict-mode activation in programs with unresolved registration parity.
- Compiler AOT: moved diagnostic/await AST global emission to the end of generated source so nodes discovered during render are always declared (fixes missing `__able_binary_node_*` declarations in stdlib-heavy compiled outputs).
- Interpreter: improved compiled impl-thunk mismatch diagnostics and added interface-generic substitution during impl-thunk param matching.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_EXEC_FIXTURES=13_06_stdlib_package_resolution,06_08_array_ops_mutability,10_02_impl_where_clause,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=06_08_array_ops_mutability,08_01_bytecode_if_indexing,06_12_02_stdlib_array_helpers,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_02_impl_where_clause,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/compiler -run TestCompilerEmitsStructsAndWrappers -count=1` in `v12/interpreters/go`.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestInterpreterEvaluateProgramSuccess -count=1` in `v12/interpreters/go`.
- Blocker: `GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=60s` times out waiting on the compiled fixture subprocess (no stdout/stderr), so this fixture still needs dedicated follow-up before claiming full strict-dispatch parity in stdlib-heavy concurrency paths.
- Compiler AOT: compiled future runtime now selects executor mode from host interpreter (`bridge.ExecutorKind`), enabling compiled goroutine scheduling when fixtures/programs use goroutine executor (fixes compiled `12_08_blocking_io_concurrency` hang path while preserving serial default behavior).
- Compiler AOT: compiled async failure normalization now wraps generic async task errors into `Future failed: ...` runtime errors while preserving cancellation (`context.Canceled`) status, restoring parity for failed/cancelled future value/status behavior.
- Compiler bridge/interpreter: added `Interpreter.ExecutorKind()` plus bridge coverage (`TestExecutorKind`) so compiled runtime can safely query scheduler mode without fallback heuristics.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/compiler/bridge -count=1`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=70s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_03_spawn_future_status_error,12_04_future_handle_value_view,12_06_await_fairness_cancellation,12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=70s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=12_01_bytecode_spawn_basic,12_01_bytecode_await_default,12_02_async_spawn_combo,12_02_future_fairness_cancellation,12_03_spawn_future_status_error,12_04_future_handle_value_view,12_05_concurrency_channel_ping_pong,12_05_mutex_lock_unlock,12_06_await_fairness_cancellation,12_07_channel_mutex_error_types,12_08_blocking_io_concurrency go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=70s`.
- Tests: `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache ABLE_COMPILER_EXEC_FIXTURES=10_17_interface_overload_dispatch,10_02_impl_where_clause,10_12_interface_union_target_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout=70s`.
- Compiler AOT: completed dynamic boundary hardening audit coverage by running `TestCompilerBoundaryFallbackMarkerForStaticFixtures` across `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all` (full exec fixture corpus) with zero compiled->interpreter fallback marker regressions.
- Compiler AOT: generated `main.go` now includes a guarded static fast-path that skips `interp.EvaluateProgram(...)` for safe `main`-only, fallback-free, non-dynamic programs; dynamic/complex programs keep the interpreter bootstrap path.
- Compiler runtime: compiled-call registration now defines callable bindings directly in each package runtime environment (`env.Define(name, fn)`) in addition to compiled-call table registration.
- Compiler tests: added `pkg/compiler/compiler_main_bootstrap_test.go` to assert static launcher generation skips evaluation while dynamic launcher generation retains interpreter bootstrap.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerCompiledHashSetUnionStdlib|TestCompilerZeroFieldStructIdentifierValue|TestCompilerSingletonStaticOverloadDispatch|TestCompilerBuildsLargeI128AndU128Literals|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundary|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all go test -v ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1` (full-corpus audit; ~339s aggregate).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Compiler AOT: expanded compiled main no-bootstrap eligibility from `main`-only to static function graphs where all discovered functions are compileable and the module set remains fallback-free/non-dynamic with no struct/method/impl bootstrap requirements.
- Compiler AOT: generated registration now conditionally installs compiled function overload thunks only when a matching runtime declaration already exists in the package environment, avoiding no-bootstrap hard-fail paths while preserving strict checks for bootstrapped environments.
- Compiler tests: added `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers` in `pkg/compiler/compiler_main_bootstrap_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerCompiledHashSetUnionStdlib|TestCompilerZeroFieldStructIdentifierValue|TestCompilerSingletonStaticOverloadDispatch|TestCompilerBuildsLargeI128AndU128Literals|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundary|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Compiler AOT: expanded no-bootstrap launch support to safe single-package programs with compileable functions/methods/impls (still requiring no dynamic usage and no compiler fallbacks).
- Compiler AOT: `RegisterIn` now detects whether interpreter metadata was preloaded (`__able_bootstrapped_metadata`) and gates interpreter thunk-registration APIs accordingly, while always registering compiled call/method/interface dispatch tables.
- Compiler AOT: per-package struct definitions are now seeded into runtime environments during registration when missing, using generated `runtime.StructDefinitionValue` placeholders with preserved struct kind/field/generic shape metadata.
- Compiler AOT: overload thunk registration now executes before compiled-call env replacement, preserving bootstrap registration against interpreted declarations.
- Compiler tests: added `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls` in `pkg/compiler/compiler_main_bootstrap_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundary|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Compiler AOT: expanded no-bootstrap main gating from single-package-only to a statically-seedable multi-package path (`noBootstrapImportsSeedable`), still requiring non-dynamic code with zero fallbacks and fully compileable functions/methods/impls.
- Compiler AOT: generator now collects module static import metadata (`ImportStatement`) per package and records selector/package/wildcard forms for launch-time seeding decisions.
- Compiler runtime registration: added no-bootstrap-only static import seeding in `RegisterIn` that constructs per-package public symbol maps (public compiled callables + public struct defs) and replays package/selector/wildcard bindings into package envs without interpreter bootstrap.
- Compiler runtime registration: seeded package callable exports via generated `NativeFunctionValue` proxies that swap into source package env and dispatch through `__able_call_named`, preserving compiled-call dispatch semantics.
- Compiler tests: added `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports` to verify generated `main.go` skips `EvaluateProgram(...)` for a static multi-package import case and that generated registration emits import-seeding code.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundary|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Compiler AOT: expanded no-bootstrap import seedability to include wildcard imports and named impl namespace imports (in addition to package/selector imports), while preserving fallback/dynamic/compileability guards.
- Compiler generator: now tracks union definitions/packages alongside interface metadata so no-bootstrap import seeding can expose public unions/interfaces consistently.
- Compiler registration (`RegisterIn`): per-package env seeding now covers missing interface and union definitions in no-bootstrap flows (struct seeding remained in place).
- Compiler registration (`RegisterIn`): added named impl namespace seeding for packages without bootstrapped metadata, constructing `runtime.ImplementationNamespaceValue` with compiled native method wrappers.
- Compiler registration no-bootstrap import seeding: source package public maps now include public functions, structs, interfaces, unions, and named impl namespaces; wildcard seeding replays all importable names.
- Safety gate: no-bootstrap now rejects ambiguous named impl namespace shapes (mixed targets/interfaces per impl name or overloaded impl namespace methods). These cases continue to bootstrap.
- Compiler tests: added
  - `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport`
  - `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport`
  in `v12/interpreters/go/pkg/compiler/compiler_main_bootstrap_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerBoundaryFallbackMarkerForStaticFixtures|TestCompilerDynamicBoundary|TestCompilerStrictDispatchForStdlibHeavyFixtures' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_05_interface_named_impl_defaults,10_09_interface_named_impl_inherent,06_12_01_stdlib_string_helpers go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Refactor: split compiler render helpers to maintain file-size guardrails (<1000 lines) by moving struct-definition helpers into `generator_render_struct_defs.go` and thunk emitters into `generator_render_thunks.go`; `generator_render_functions.go` is now under 1000 lines.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler/bridge -count=1`.
- Compiler AOT: no-bootstrap named impl namespace seeding now supports overloaded methods by emitting per-namespace overload dispatch closures (type-match scoring + ambiguity/no-match errors + partial application via `runtime.PartialFunctionValue`).
- Compiler AOT: removed the no-bootstrap seedability rejection for overloaded named impl namespace methods; the safety gate now validates overload dispatchability (compileable entries + renderable concrete param types).
- Compiler tests: added `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads` in `v12/interpreters/go/pkg/compiler/compiler_main_bootstrap_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent|TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_05_interface_named_impl_defaults,10_09_interface_named_impl_inherent,10_17_interface_overload_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Compiler AOT: no-bootstrap import callable seeding now binds imported callables directly to compiled call table entries (`__able_lookup_compiled_call`) rather than routing through `__able_call_named`, reducing boundary indirection for static cross-package calls.
- Compiler tests: strengthened `TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports` to assert direct compiled-call binding in generated seeding logic.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Compiler AOT: started package-output partitioning by adding a dedicated generated artifact `compiled_packages.go` that emits per-package registrar helpers plus a shared `__able_register_compiled_packages(...)` entrypoint.
- Compiler AOT: `compiled.go` registration now delegates package seeding/callable registration to the generated package registrar entrypoint, preserving runtime behavior while separating package-scoped emission.
- Compiler tests: `compiler_main_bootstrap_test` now inspects combined generated compiled sources (`compiled.go` + `compiled_packages.go`) so no-bootstrap assertions cover split output.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Compiler AOT: continued package-output partitioning by moving function/overload callable registration emission into per-package generated files (`compiled_pkg_callables_*.go`) via package-scoped helpers (`__able_register_compiled_package_callables_*`).
- Compiler AOT: package registrars in `compiled_packages.go` now focus on package env/bootstrap seeding and delegate callable registration to per-package callable helpers.
- Compiler tests: bootstrap source inspection now aggregates all `compiled*.go` outputs so assertions stay valid as generated artifacts split across files.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Compiler AOT: continued package-output partitioning by moving method/impl registration emission into per-package generated files (`compiled_pkg_methods_impls_*.go`) via package-scoped helpers (`__able_register_compiled_package_methods_impls_*`).
- Compiler AOT: `RegisterIn` now invokes generated `__able_register_compiled_method_impl_packages(...)` for method/impl registration before interface-dispatch table setup, preserving registration order while removing inline monolithic emission.
- Compiler AOT: package registrars in `compiled_packages.go` now focus on package env/bootstrap/definition seeding and delegate methods+impls and callables to dedicated per-package generated helpers.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecHarness|TestCompilerSingletonStaticOverloadDispatch|TestCompilerZeroFieldStructIdentifierValue|TestCompilerCompiledHashSetUnionStdlib|TestCompilerEmitsStructsAndWrappers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Compiler AOT: continued generated-output partitioning by moving interface-dispatch and method-overload registration emission out of `compiled.go` into dedicated generated artifact `compiled_interface_dispatch.go`.
- Compiler AOT: `RegisterIn` now delegates dispatch setup to `__able_register_compiled_interface_dispatch(rt)` after per-package method/impl registration, preserving registration order while reducing monolithic `compiled.go` emission.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: normalized `InterfaceValue` member-get-method dispatch to a single shared path in generated runtime calls (value/pointer forms now flow through one local `iface` value), removing duplicated pointer/value interface dispatch branches while preserving strict interface miss behavior.
- Compiler tests: added `TestCompilerNormalizesInterfaceMemberGetMethodDispatch` in `v12/interpreters/go/pkg/compiler/compiler_interface_member_get_method_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.076s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 512.628s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.945s`).
- CLI/compiler integration: added shared `ABLE_COMPILER_REQUIRE_NO_FALLBACKS` parsing helper in `v12/interpreters/go/cmd/able/compiler_options.go` with strict token validation and explicit error messages for invalid values.
- CLI/compiler integration: `able test --compiled` now applies the same `RequireNoFallbacks` setting (via env) before compilation, matching `able build` strict fallback behavior.
- CLI tests: added `TestBuildNoFallbacksFlagFailsWhenFallbackRequired`, `TestBuildNoFallbacksEnvFailsWhenFallbackRequired`, and `TestBuildNoFallbacksInvalidEnvFailsArgumentParsing` in `v12/interpreters/go/cmd/able/build_test.go`.
- CLI tests: added `TestBuildAllowFallbacksOverridesEnv` in `v12/interpreters/go/cmd/able/build_test.go` and `TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv` in `v12/interpreters/go/cmd/able/test_cli_test.go`.
- Tests: `cd v12/interpreters/go && go test ./cmd/able -run 'TestBuildNoFallbacksFlagFailsWhenFallbackRequired|TestBuildNoFallbacksEnvFailsWhenFallbackRequired|TestBuildNoFallbacksInvalidEnvFailsArgumentParsing|TestBuildAllowFallbacksOverridesEnv|TestBuildTargetFromManifest|TestBuildOutputOutsideModuleRoot|TestTestCommandCompiledRuns|TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./cmd/able -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRequireNoFallbacksFails|TestCompilerEmitsStructsAndWrappers' -count=1`.
- Compiler AOT: continued generated-output partitioning by moving no-bootstrap static import seeding emission out of `compiled_packages.go` into dedicated generated artifact `compiled_import_seeding.go`.
- Compiler AOT: package registration now calls `__able_seed_no_bootstrap_imports(__able_bootstrapped_metadata)` so `compiled_packages.go` focuses on package registrar orchestration.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Compiler AOT: continued generated-output partitioning by moving per-package definition seeding (struct/interface/union + named impl namespace seeds) out of `compiled_packages.go` into per-package generated artifacts (`compiled_pkg_defs_*.go`).
- Compiler AOT: package registrars now delegate definition seeding to generated helpers (`__able_register_compiled_package_defs_*`) before callable registration, reducing monolithic package registrar emission.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: continued generated-output partitioning by moving per-package package-registrar emission out of `compiled_packages.go` into per-package generated artifacts (`compiled_pkg_registrar_*.go`).
- Compiler AOT: `compiled_packages.go` now focuses on aggregator orchestration (`__able_register_compiled_method_impl_packages`, `__able_register_compiled_packages`) and delegates per-package registration to generated registrar helpers.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: continued generated-output partitioning by moving register/run entrypoint emission (`Register`, `RegisterIn`, `RunMain`, `RunMainIn`, `RunRegisteredMain`) out of `compiled.go` into dedicated generated artifact `compiled_register.go`.
- Compiler AOT: `compiled.go` now focuses on runtime helpers, compiled bodies/wrappers, thunks, overload dispatchers, and diagnostics, while registration entrypoint orchestration is emitted separately.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: continued generated-output partitioning by moving package-registration aggregator emission out of `compiled_packages.go` into dedicated generated artifact `compiled_package_aggregators.go`.
- Compiler AOT: generated output no longer emits `compiled_packages.go`; aggregator orchestration now lives in `compiled_package_aggregators.go` while per-package registration remains in `compiled_pkg_registrar_*.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: added compiler option `RequireNoFallbacks` to fail compilation when any fallback wrappers would be emitted, providing an explicit no-silent-fallback guardrail for strict builds.
- Compiler tests: added `TestCompilerRequireNoFallbacksFails` in `v12/interpreters/go/pkg/compiler/compiler_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerEmitsStructsAndWrappers|TestCompilerRequireNoFallbacksFails' -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport|TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads|TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent' -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=13_01_package_structure_modules,13_04_import_alias_selective_dynimport,13_06_stdlib_package_resolution go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES=10_03_interface_type_dynamic_dispatch,10_04_interface_dispatch_defaults_generics,10_06_interface_generic_param_dispatch,10_11_interface_generic_args_dispatch,10_12_interface_union_target_dispatch,10_17_interface_overload_dispatch,13_05_dynimport_interface_dispatch go test ./pkg/compiler -run TestCompilerExecFixtures -count=1`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1`.
- Compiler AOT: normalized `InterfaceValue` member-get-method dispatch to a single shared path in generated runtime calls (value/pointer forms now flow through one local `iface` value), removing duplicated pointer/value interface dispatch branches while preserving strict interface miss behavior.
- Compiler tests: added `TestCompilerNormalizesInterfaceMemberGetMethodDispatch` in `v12/interpreters/go/pkg/compiler/compiler_interface_member_get_method_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.076s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 512.628s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.945s`).
- Compiler AOT: normalized `DynPackage` builtin member-call receiver handling to a single shared value path (value/pointer receiver assertions now populate one local `dyn` value before `bridge.MemberGet`), removing duplicated pointer/value receiver switch branches while preserving `DynPackage.def`/`DynPackage.eval` behavior.
- Compiler tests: added `TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver` in `v12/interpreters/go/pkg/compiler/compiler_dynpackage_member_call_receiver_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver|TestCompilerRegistersBuiltinDynPackageMemberMethods' -count=1` (`ok ... 0.040s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES='13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (`ok ... 3.753s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.602s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 235.950s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 507.252s`).
- Compiler AOT: normalized builtin `Error.message`/`Error.cause` receiver handling to a shared helper (`__able_builtin_error_receiver`) so value/pointer receivers now flow through one local `runtime.ErrorValue` normalization path.
- Compiler tests: added `TestCompilerNormalizesErrorBuiltinMemberReceivers` in `v12/interpreters/go/pkg/compiler/compiler_error_member_receiver_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesErrorBuiltinMemberReceivers|TestCompilerRegistersBuiltinErrorMemberMethods|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (`ok ... 0.058s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES='12_07_channel_mutex_error_types,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (`ok ... 3.856s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 58.021s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 232.297s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 500.790s`).
- Compiler AOT: normalized duplicated pointer/value native callable-dispatch branches in `__able_call_value` via shared helper `__able_call_native_function` so native function values, native bound-method values, and bound-method native members reuse one arity/error/partial-application path.
- Compiler tests: added `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesErrorBuiltinMemberReceivers|TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver' -count=1` (`ok ... 0.057s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES='12_07_channel_mutex_error_types,13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (`ok ... 6.302s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 61.363s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 530.283s`).
- Compiler AOT: normalized duplicated pointer/value interface + partial unwrapping branches in `__able_call_value` to shared helpers (`__able_callable_interface_value`, `__able_callable_partial_value`, `__able_merge_bound_args`) while preserving nil-pointer-to-nil-value behavior and bound-arg merge semantics.
- Compiler tests: added `TestCompilerNormalizesCallValueUnwrapBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_unwrap_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesErrorBuiltinMemberReceivers' -count=1` (`ok ... 0.058s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_EXEC_FIXTURES='12_07_channel_mutex_error_types,13_04_import_alias_selective_dynimport,13_05_dynimport_interface_dispatch' go test ./pkg/compiler -run TestCompilerExecFixtures -count=1` (`ok ... 5.703s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 75.289s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 243.026s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 529.914s`).
- Compiler AOT: normalized duplicated pointer/value callable-name unwrapping in `__able_callable_name` to shared helpers (`__able_callable_native_function_value`, `__able_callable_native_bound_method_value`, `__able_callable_bound_method_value`) while reusing shared partial/interface unwrapping helpers.
- Compiler tests: added `TestCompilerNormalizesCallableNameUnwrapBranches` in `v12/interpreters/go/pkg/compiler/compiler_callable_name_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallableNameUnwrapBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallValueNativeDispatchBranches' -count=1` (`ok ... 0.059s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 60.268s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 506.503s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 236.352s`).
- Compiler AOT: normalized duplicated pointer/value bound-method dispatch branches in `__able_call_value` into shared helper `__able_call_bound_method` while preserving compiled-thunk + native-member dispatch behavior.
- Compiler tests: added `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_bound_method_shim_regression_test.go`.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert the new bound-method helper path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.077s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.331s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 520.056s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 236.549s`).
- Compiler AOT: normalized duplicated pointer/value native-bound receiver-injection dispatch branches in `__able_call_value` into shared helper `__able_call_native_bound_method` while preserving native partial-target semantics.
- Compiler tests: added `TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_bound_method_shim_regression_test.go`.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert the new native-bound helper path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.097s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 56.348s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 504.781s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 245.303s`).
- Compiler AOT: normalized duplicated `*runtime.FunctionValue` compiled-thunk dispatch in `__able_call_value` and the bound-method thunk branch via shared helper `__able_call_function_thunk` while preserving runtime-error context handling and nil-value semantics.
- Compiler tests: added `TestCompilerNormalizesCallValueFunctionThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_value_function_thunk_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.113s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 57.563s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 492.656s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 229.642s`).
- Compiler AOT: normalized duplicated `runtime.NativeFunctionValue` / `*runtime.NativeFunctionValue` dispatch in `__able_call_value` via shared helper `__able_call_native_function_value` while preserving partial-target semantics for pointer/value targets.
- Compiler tests: added `TestCompilerNormalizesCallValueNativeFunctionDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_function_shim_regression_test.go`.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert the shared native-function helper path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.126s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.973s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 530.525s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 246.697s`).
- Compiler AOT: normalized duplicated `runtime.NativeBoundMethodValue` / `*runtime.NativeBoundMethodValue` switch dispatch in `__able_call_value` via shared unwrapping path `__able_callable_native_bound_method_value` with unified dispatch through `__able_call_native_bound_method(nativeBound, fn, ...)`.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_bound_method_shim_regression_test.go` to assert switch-branch removal and the new normalized unwrapping path.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert the new native-bound unwrapping path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.127s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 60.240s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 509.840s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 244.857s`).
- Compiler AOT: normalized duplicated `runtime.BoundMethodValue` / `*runtime.BoundMethodValue` switch dispatch in `__able_call_value` via shared unwrapping path `__able_callable_bound_method_value` with unified dispatch through `__able_call_bound_method(bound, fn, ...)`.
- Compiler tests: updated `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_bound_method_shim_regression_test.go` to assert switch-branch removal and the normalized bound-method unwrapping path.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert the normalized bound-method unwrapping path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.144s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.867s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 525.014s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 256.323s`).
- Compiler AOT: removed the now-single-case `switch v := fn.(type)` in `__able_call_value` by normalizing `*runtime.FunctionValue` dispatch to a direct assertion path that routes through `__able_call_function_thunk`.
- Compiler tests: updated `TestCompilerNormalizesCallValueFunctionThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_value_function_thunk_shim_regression_test.go` to assert switch removal and direct assertion path usage.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.133s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.023s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 590.436s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 234.021s`).
- Compiler tests: strengthened `TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod` in `v12/interpreters/go/pkg/compiler/compiler_string_method_registration_test.go` to assert generated wrappers do not use `__able_call_named("String.from_bytes_unchecked", ...)` fallback dispatch.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty' -count=1` (`ok ... 0.042s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 57.430s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 497.476s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 245.787s`).
- Compiler AOT: normalized `__able_call_bound_method` by removing `switch method := bound.Method.(type)` and routing native-function dispatch through `__able_callable_native_function_value(bound.Method)` while keeping direct `*runtime.FunctionValue` thunk assertion + shared helper dispatch.
- Compiler tests: updated `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_bound_method_shim_regression_test.go` to assert removal of the bound-method switch and helper-based native unwrapping.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert helper-based native unwrapping inside bound-method dispatch.
- Compiler tests: updated `TestCompilerNormalizesCallValueFunctionThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_value_function_thunk_shim_regression_test.go` for the renamed direct thunk assertion local.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.127s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.269s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 510.913s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 231.459s`).
- Compiler AOT: normalized `__able_interface_bind_receiver_method` by removing method-kind pointer/value switch dispatch and routing through shared callable unwrapping helpers (`__able_callable_native_function_value`, `__able_callable_native_bound_method_value`, `__able_callable_bound_method_value`) while preserving nil-pointer behavior and receiver binding semantics.
- Compiler tests: added `TestCompilerNormalizesInterfaceBindReceiverMethodDispatch` in `v12/interpreters/go/pkg/compiler/compiler_interface_bind_receiver_method_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch' -count=1` (`ok ... 0.090s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.495s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 496.160s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 241.374s`).
- Compiler AOT: normalized `__able_runtime_value_type_name` by replacing legacy pointer/value switch unwrapping for interface/type-ref/numeric paths with shared helpers (`__able_callable_interface_value`, `__able_runtime_struct_definition_value`, `__able_runtime_type_ref_value`, `__able_runtime_integer_value`, `__able_runtime_float_value`) while preserving nil-pointer behavior and type-name resolution semantics.
- Compiler tests: added `TestCompilerNormalizesRuntimeValueTypeNameUnwrapping` in `v12/interpreters/go/pkg/compiler/compiler_runtime_value_type_name_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesRuntimeValueTypeNameUnwrapping|TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch' -count=1` (`ok ... 0.129s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 60.871s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 559.009s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 241.198s`).
- Compiler AOT: normalized `__able_interface_dispatch_static_receiver` by removing explicit pointer/value switch branches and routing static-receiver detection through shared struct/type-ref unwrapping helpers (`__able_runtime_struct_definition_value`, `__able_runtime_type_ref_value`) with preserved nil-pointer semantics.
- Compiler tests: added `TestCompilerNormalizesInterfaceDispatchStaticReceiver` in `v12/interpreters/go/pkg/compiler/compiler_interface_dispatch_static_receiver_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceDispatchStaticReceiver|TestCompilerNormalizesRuntimeValueTypeNameUnwrapping|TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch' -count=1` (`ok ... 0.130s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.392s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 527.628s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 254.320s`).
- Compiler AOT: normalized `__able_interface_dispatch_member` candidate matching by extracting receiver-type/generic/`MatchType` resolution into shared helper `__able_interface_dispatch_match_entry`, preserving constraint enforcement and coerced-receiver semantics while removing duplicated inline match branches.
- Compiler tests: added `TestCompilerNormalizesInterfaceDispatchMemberMatchResolution` in `v12/interpreters/go/pkg/compiler/compiler_interface_dispatch_member_match_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceDispatchMemberMatchResolution|TestCompilerNormalizesInterfaceDispatchStaticReceiver|TestCompilerNormalizesRuntimeValueTypeNameUnwrapping|TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch' -count=1` (`ok ... 0.158s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.169s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 506.131s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 241.648s`).
- Compiler AOT: generalized static method lowering for nominal-struct returns by replacing the hardcoded `String.from_bytes_unchecked` return-type override with shared helper `staticMethodNominalStructReturnType` used by both method and impl lowering (`fillMethodInfo`, `fillImplMethodInfo`), preserving no-fallback compileability for `impl Default for String` static dispatch while allowing additional static constructors that return their target struct type to compile without fallback.
- Compiler tests: added `TestCompilerRegistersCompiledStringStaticStructReturnMethod` in `v12/interpreters/go/pkg/compiler/compiler_string_static_struct_return_registration_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod|TestCompilerRegistersCompiledStringStaticStructReturnMethod|TestCompilerNoFallbacksStringDefaultImplStaticEmpty|TestCompilerNormalizesInterfaceDispatchMemberMatchResolution' -count=1` (`ok ... 0.079s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 60.376s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 519.249s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 243.853s`).
- Compiler AOT: normalized interface runtime method binding by replacing legacy method-kind pointer/value switches in `__able_interface_method_receiver` and `__able_interface_bind_method` with shared callable unwrapping helpers (`__able_callable_bound_method_value`, `__able_callable_native_function_value`, `__able_callable_native_bound_method_value`), preserving nil-pointer behavior and bound receiver semantics while reducing branch-local duplication.
- Compiler tests: added `TestCompilerNormalizesInterfaceBindMethodDispatch` in `v12/interpreters/go/pkg/compiler/compiler_interface_bind_method_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceBindMethodDispatch|TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceDispatchMemberMatchResolution|TestCompilerNormalizesInterfaceDispatchStaticReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.096s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 57.544s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 526.627s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 239.138s`).
- Compiler AOT: normalized callable-kind detection in `__able_is_callable_value` by replacing the legacy pointer/value type switch with shared callable unwrapping helpers (`__able_callable_native_function_value`, `__able_callable_native_bound_method_value`, `__able_callable_bound_method_value`, `__able_callable_partial_value`) plus direct compiled-function assertions, preserving typed-nil callable semantics while reducing dispatch-kind shim duplication.
- Compiler tests: added `TestCompilerNormalizesIsCallableValueDispatch` in `v12/interpreters/go/pkg/compiler/compiler_is_callable_value_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.139s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 57.447s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 510.969s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 235.150s`).
- Compiler AOT: normalized member-name extraction in `__able_member_name` by replacing the legacy pointer/value string switch with shared helper `__able_runtime_string_value`, preserving nil-pointer string behavior while reducing branch-local string-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesMemberNameStringUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_member_name_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallableNameUnwrapBranches' -count=1` (`ok ... 0.167s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 63.281s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 536.874s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 236.355s`).
- Compiler AOT: normalized qualified-callable candidate nil filtering in `__able_resolve_qualified_callable` by replacing the local `switch candidate.(type)` branch with shared `__able_is_nil(candidate)` checking, preserving nil-sentinel behavior while reducing branch-local nil-kind dispatch duplication.
- Compiler tests: updated `TestCompilerRemovesNilPointerQualifiedCallableShim` in `v12/interpreters/go/pkg/compiler/compiler_nil_pointer_qualified_callable_shim_regression_test.go` to assert switch removal and shared nil-helper usage.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRemovesNilPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerQualifiedCallableShim|TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch' -count=1` (`ok ... 0.107s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.112s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 542.815s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 247.830s`).
- Compiler AOT: normalized integer unwrapping in runtime string helper `__able_int64_from_value` by replacing the local pointer/value integer switch with shared helper `__able_runtime_integer_value`, preserving nil-pointer and 64-bit fit checks while reducing branch-local integer-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesInt64FromValueIntegerUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_int64_from_value_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesRuntimeValueTypeNameUnwrapping|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerRemovesNilPointerQualifiedCallableShim' -count=1` (`ok ... 0.092s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 57.353s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 493.730s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 226.789s`).
- Compiler AOT: normalized string unwrapping in runtime helper `__able_string_from_builtin_impl` by replacing the local pointer/value string switch with shared helper `__able_runtime_string_value` while preserving struct-instance string handling (`__able_string_bytes_from_struct`), reducing branch-local string-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesStringFromBuiltinUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_string_from_builtin_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerRemovesNilPointerQualifiedCallableShim' -count=1` (`ok ... 0.091s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.878s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 502.838s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 230.086s`).
- Compiler AOT: normalized char unwrapping in runtime helper `__able_char_to_codepoint_impl` by replacing the local pointer/value char switch with shared helper `__able_runtime_char_value`, preserving nil-pointer char behavior while reducing branch-local char-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesCharToCodepointUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_char_to_codepoint_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCharToCodepointUnwrap|TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerRemovesNilPointerQualifiedCallableShim' -count=1` (`ok ... 0.108s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 58.222s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 500.646s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 230.275s`).
- Compiler AOT: normalized error unwrapping in core runtime helper `__able_struct_instance` by replacing the local error pointer/value switch with shared helper `__able_runtime_error_value`, preserving nil-pointer error handling while reducing branch-local error-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesStructInstanceErrorUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_struct_instance_error_unwrap_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesStructInstanceErrorUnwrap|TestCompilerNormalizesCharToCodepointUnwrap|TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerRemovesNilPointerQualifiedCallableShim' -count=1` (`ok ... 0.131s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 56.380s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 484.316s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 225.983s`).
- Compiler AOT: normalized nil detection in core runtime helper `__able_is_nil` by replacing the local `runtime.NilValue` pointer/value switch with shared helper `__able_runtime_nil_value`, preserving nil-pointer semantics while reducing branch-local nil-kind dispatch duplication.
- Compiler tests: added `TestCompilerNormalizesIsNilUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_is_nil_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesIsNilUnwrap|TestCompilerNormalizesStructInstanceErrorUnwrap|TestCompilerNormalizesCharToCodepointUnwrap|TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerRemovesNilPointerQualifiedCallableShim' -count=1` (`ok ... 0.193s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 59.157s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 505.884s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 234.512s`).
- Compiler AOT: normalized struct-instance unwrapping in core runtime helper `__able_struct_instance` by replacing direct `*runtime.StructInstanceValue` assertion with shared helper `__able_runtime_struct_instance_value`, preserving typed-nil behavior while keeping shared error unwrapping via `__able_runtime_error_value`.
- Compiler tests: updated `TestCompilerNormalizesStructInstanceErrorUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_struct_instance_error_unwrap_shim_regression_test.go` to assert shared struct-instance helper emission/usage and direct-assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesStructInstanceErrorUnwrap' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 61.170s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 501.871s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 235.305s`).
- Compiler AOT: normalized struct-instance unwrapping in runtime helper `__able_string_from_builtin_impl` by replacing direct `*runtime.StructInstanceValue` assertion with shared helper `__able_runtime_struct_instance_value`, preserving struct-instance byte extraction (`__able_string_bytes_from_struct`) while routing typed-nil pointers through the existing argument error path.
- Compiler tests: updated `TestCompilerNormalizesStringFromBuiltinUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_string_from_builtin_shim_regression_test.go` to assert shared struct-instance helper emission/usage and direct-assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesStringFromBuiltinUnwrap' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 57.838s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 500.554s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 233.138s`).
- Compiler AOT: normalized `Error` builtin receiver unwrapping in `__able_builtin_error_receiver` by replacing direct `runtime.ErrorValue`/`*runtime.ErrorValue` assertions with shared helper `__able_runtime_error_value`, preserving typed-nil pointer rejection and receiver validation semantics.
- Compiler tests: updated `TestCompilerNormalizesErrorBuiltinMemberReceivers` in `v12/interpreters/go/pkg/compiler/compiler_error_member_receiver_shim_regression_test.go` to assert shared runtime error helper emission/usage and direct pointer assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesErrorBuiltinMemberReceivers' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 61.417s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 521.518s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 245.068s`).
- Compiler AOT: normalized `DynPackage` builtin member-call receiver unwrapping in `__able_builtin_dynpackage_member_call` by replacing direct `runtime.DynPackageValue`/`*runtime.DynPackageValue` assertions with shared helper `__able_runtime_dynpackage_value`, preserving typed-nil pointer rejection and receiver validation semantics.
- Compiler tests: updated `TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver` in `v12/interpreters/go/pkg/compiler/compiler_dynpackage_member_call_receiver_shim_regression_test.go` to assert shared runtime dynpackage helper emission/usage and direct pointer assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesDynPackageBuiltinMemberCallReceiver' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 58.894s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 508.755s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 265.551s`).
- Compiler AOT: normalized `Future` builtin receiver unwrapping in `__able_builtin_future_receiver` by replacing direct `*runtime.FutureValue` assertion with shared helper `__able_runtime_future_value`, preserving typed-nil pointer rejection and receiver validation semantics.
- Compiler tests: added `TestCompilerNormalizesFutureBuiltinMemberReceiver` in `v12/interpreters/go/pkg/compiler/compiler_future_member_receiver_shim_regression_test.go`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesFutureBuiltinMemberReceiver' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.639s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 511.689s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 243.107s`).
- Compiler AOT: normalized interface-wrapped method dispatch in `__able_call_bound_method` by unwrapping `bound.Method` via shared helper `__able_callable_interface_value` before compiled-thunk/native dispatch checks, preserving typed-nil interface handling semantics.
- Compiler tests: updated `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_bound_method_shim_regression_test.go` to assert interface helper unwrapping and local-method dispatch checks.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_dispatch_shim_regression_test.go` to assert normalized bound-method native unwrapping after interface unwrapping.
- Compiler tests: updated `TestCompilerNormalizesCallValueFunctionThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_value_function_thunk_shim_regression_test.go` to assert normalized local-method thunk assertion path.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch' -count=1` (`ok ... 0.058s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.367s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 524.970s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 248.036s`).
- Compiler AOT: normalized interface member-get dispatch in `__able_member_get_method` by replacing legacy `isIface` tracking + direct `runtime.InterfaceValue`/`*runtime.InterfaceValue` assertions with shared helper `__able_callable_interface_value`, preserving strict interface-miss panic behavior and typed-nil interface-pointer handling semantics.
- Compiler tests: updated `TestCompilerNormalizesInterfaceMemberGetMethodDispatch` in `v12/interpreters/go/pkg/compiler/compiler_interface_member_get_method_shim_regression_test.go` to assert helper-based interface unwrapping and legacy assertion branch removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.021s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 62.491s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 540.533s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.675s`).
- Compiler AOT: normalized `Iterator.next` builtin receiver unwrapping in `__able_builtin_iterator_next` by replacing direct `*runtime.IteratorValue` assertion with shared helper `__able_runtime_iterator_value`, preserving typed-nil pointer rejection and receiver validation semantics.
- Compiler tests: added `TestCompilerNormalizesIteratorBuiltinMemberReceiver` in `v12/interpreters/go/pkg/compiler/compiler_iterator_member_receiver_shim_regression_test.go` to assert shared helper usage and legacy direct assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.040s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.366s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 527.393s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 257.697s`).
- Compiler AOT: normalized array receiver unwrapping in `__able_member_set` and `__able_member_get` by replacing duplicated direct `*runtime.ArrayValue` assertions with shared helper `__able_runtime_array_value`, preserving typed-nil pointer bypass behavior while removing branch-local array assertion duplication.
- Compiler tests: added `TestCompilerNormalizesArrayMemberReceiverUnwrap` in `v12/interpreters/go/pkg/compiler/compiler_array_member_receiver_shim_regression_test.go` to assert helper usage and legacy direct assertion removal in both member-set and member-get helpers.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.058s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.895s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 530.935s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 252.561s`).
- Compiler AOT: normalized direct callable pointer assertions in `__able_is_callable_value` by replacing legacy `*runtime.FunctionValue` / `*runtime.FunctionOverloadValue` checks with shared helpers (`__able_callable_function_value`, `__able_callable_function_overload_value`), preserving typed-nil callable semantics while removing branch-local assertion duplication.
- Compiler tests: updated `TestCompilerNormalizesIsCallableValueDispatch` in `v12/interpreters/go/pkg/compiler/compiler_is_callable_value_shim_regression_test.go` to assert shared helper emission/usage and legacy direct assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.076s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 62.788s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 531.135s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 242.430s`).
- Compiler AOT: normalized function-thunk dispatch unwrapping in `__able_call_bound_method` and `__able_call_value` by replacing direct `*runtime.FunctionValue` assertions with shared helper `__able_callable_function_value`, preserving typed-nil thunk semantics while removing branch-local assertion duplication.
- Compiler tests: updated `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` and `TestCompilerNormalizesCallValueFunctionThunkDispatch` to assert helper-based thunk unwrapping and direct assertion removal.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.131s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 58.581s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 521.272s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 232.812s`).
- Compiler AOT: normalized thunk helper internals by changing `__able_call_function_thunk` to accept `runtime.Value` and unwrap via shared helper `__able_callable_function_value`, then routing `__able_call_bound_method` and `__able_call_value` through direct `__able_call_function_thunk(method, ...)` / `__able_call_function_thunk(fn, ...)` delegation.
- Compiler tests: updated `TestCompilerNormalizesCallValueFunctionThunkDispatch` and `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` to assert helper-internal unwrapping and direct thunk-helper delegation.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.130s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.710s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 524.507s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 245.028s`).
- Compiler AOT: normalized compiled-thunk bytecode dispatch in `__able_call_compiled_thunk` by extracting bytecode-to-thunk resolution into shared helper `__able_compiled_thunk_value` for both `interpreter.CompiledThunk` and function-signature bytecode forms, removing duplicated direct bytecode assertion branches from the call helper.
- Compiler tests: added `TestCompilerNormalizesCallCompiledThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_compiled_thunk_shim_regression_test.go` to assert shared helper usage and direct assertion branch removal in `__able_call_compiled_thunk`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallCompiledThunkDispatch|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.147s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.547s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 526.948s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 257.062s`).
- Compiler AOT: normalized compiled-thunk helper parity by updating `__able_compiled_thunk_value` to the shared helper return contract `(value, ok, nilPtr)` (typed-nil signaling aligned with other runtime unwrapping helpers) and routing `__able_call_compiled_thunk` through a unified `if !ok || nilPtr` guard.
- Compiler tests: updated `TestCompilerNormalizesCallCompiledThunkDispatch` in `v12/interpreters/go/pkg/compiler/compiler_call_compiled_thunk_shim_regression_test.go` to assert helper signature change and typed-nil guard usage.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallCompiledThunkDispatch|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.160s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 62.343s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 524.781s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 242.480s`).
- Compiler AOT: normalized native-function helper guard ordering in `__able_call_native_function_value` to shared helper style (`if !ok || nilPtr`) after `__able_callable_native_function_value` lookup, aligning guard semantics with other helper-based unwrap paths.
- Compiler tests: updated `TestCompilerNormalizesCallValueNativeFunctionDispatchBranches` in `v12/interpreters/go/pkg/compiler/compiler_call_value_native_function_shim_regression_test.go` to assert normalized guard ordering and reject legacy ordering.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallCompiledThunkDispatch|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.178s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 66.466s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 522.997s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.271s`).
- Compiler AOT: normalized helper-order guard semantics in `__able_callable_name` by replacing nilPtr-first helper branches with `ok || nilPtr` checks followed by explicit `if !ok || nilPtr` handling across native/native-bound/bound/partial/interface unwrap paths.
- Compiler tests: updated `TestCompilerNormalizesCallableNameUnwrapBranches` in `v12/interpreters/go/pkg/compiler/compiler_callable_name_shim_regression_test.go` to assert legacy nilPtr-first branch removal and normalized guard usage.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallableNameUnwrapBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallCompiledThunkDispatch|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.203s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 64.010s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 546.965s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 253.677s`).
- Compiler AOT: normalized remaining nilPtr-first unwrap loops in `__able_call_bound_method` and `__able_call_value` by replacing interface/partial helper branches with `ok || nilPtr` checks followed by explicit `if !ok || nilPtr` handling.
- Compiler tests: updated `TestCompilerNormalizesCallValueBoundMethodDispatchBranches` and `TestCompilerNormalizesCallValueUnwrapBranches` to assert legacy nilPtr-first branch removal and normalized guard usage.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallValueFunctionThunkDispatch|TestCompilerNormalizesCallValueNativeDispatchBranches|TestCompilerNormalizesCallableNameUnwrapBranches|TestCompilerNormalizesCallValueNativeFunctionDispatchBranches|TestCompilerNormalizesCallCompiledThunkDispatch|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesArrayMemberReceiverUnwrap|TestCompilerNormalizesIteratorBuiltinMemberReceiver|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.211s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.994s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 527.428s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 238.177s`).
- Compiler AOT: normalized remaining ok-only typed-nil helper branches in `__able_builtin_error_receiver` and `__able_member_name` by replacing `if ...; ok { if nilPtr { ... } }` with `ok || nilPtr` checks followed by explicit `if !ok || nilPtr` handling.
- Compiler tests: updated `TestCompilerNormalizesErrorBuiltinMemberReceivers` and `TestCompilerNormalizesMemberNameStringUnwrap` to reject legacy ok-only guard branches and assert normalized typed-nil helper guards.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesErrorBuiltinMemberReceivers|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesIsCallableValueDispatch|TestCompilerNormalizesCallValueUnwrapBranches|TestCompilerNormalizesCallValueBoundMethodDispatchBranches|TestCompilerNormalizesCallableNameUnwrapBranches|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.140s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.058s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 510.034s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 239.319s`).
- Compiler AOT: normalized remaining ok-only typed-nil helper branches in interface/string renderers by updating `__able_interface_bind_method`, `__able_int64_from_value`, `__able_string_from_builtin_impl`, and `__able_char_to_codepoint_impl` from `if ...; ok { if nilPtr { ... } }` to `ok || nilPtr` checks followed by explicit `if !ok || nilPtr` handling.
- Compiler tests: updated `TestCompilerNormalizesInterfaceBindMethodDispatch`, `TestCompilerNormalizesInt64FromValueIntegerUnwrap`, `TestCompilerNormalizesStringFromBuiltinUnwrap`, and `TestCompilerNormalizesCharToCodepointUnwrap` to reject legacy ok-only guards and assert normalized typed-nil handling.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceBindMethodDispatch|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesCharToCodepointUnwrap|TestCompilerNormalizesErrorBuiltinMemberReceivers|TestCompilerNormalizesMemberNameStringUnwrap|TestCompilerNormalizesCallableNameUnwrapBranches|TestCompilerNormalizesCallValueUnwrapBranches' -count=1` (`ok ... 0.147s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 61.879s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 516.841s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 243.724s`).
- Compiler AOT: normalized remaining ok-only typed-nil helper branches in interface-member/runtime helpers by updating `__able_interface_bind_receiver_method`, `__able_interface_dispatch_static_receiver`, `__able_runtime_value_type_name`, and `__able_struct_instance` from `if ...; ok { if nilPtr { ... } }` to `ok || nilPtr` checks followed by explicit `if !ok || nilPtr` handling (or equivalent `ok && !nilPtr` return for static receiver checks).
- Compiler tests: updated `TestCompilerNormalizesInterfaceBindReceiverMethodDispatch`, `TestCompilerNormalizesInterfaceDispatchStaticReceiver`, `TestCompilerNormalizesRuntimeValueTypeNameUnwrapping`, and `TestCompilerNormalizesStructInstanceErrorUnwrap` to reject legacy ok-only guards and assert normalized typed-nil handling.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerNormalizesInterfaceBindReceiverMethodDispatch|TestCompilerNormalizesInterfaceDispatchStaticReceiver|TestCompilerNormalizesRuntimeValueTypeNameUnwrapping|TestCompilerNormalizesStructInstanceErrorUnwrap|TestCompilerNormalizesInterfaceBindMethodDispatch|TestCompilerNormalizesInt64FromValueIntegerUnwrap|TestCompilerNormalizesStringFromBuiltinUnwrap|TestCompilerNormalizesCharToCodepointUnwrap' -count=1` (`ok ... 0.150s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.195s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 515.407s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 236.694s`).
- Compiler AOT: removed redundant value-form static lookup shim branches for `runtime.StructDefinitionValue` and `runtime.TypeRefValue` inside `__able_member_get_method`, routing static method resolution through shared `__able_runtime_value_type_name(base)` + `__able_interface_dispatch_static_receiver(base)` path.
- Compiler tests: updated `TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim` and `TestCompilerRemovesTypeRefPointerMemberGetMethodShim` to assert direct branch/lookup shim removal in `__able_member_get_method` while preserving shared static-receiver lookup path assertions.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerNormalizesInterfaceMemberGetMethodDispatch|TestCompilerRemovesImplNamespacePointerMemberGetMethodShim|TestCompilerRemovesPackagePublicMemberGetMethodShim|TestCompilerRemovesErrorValueMemberGetMethodShim' -count=1` (`ok ... 0.116s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 59.635s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 526.853s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 236.553s`).
- Compiler AOT: removed redundant value-form static lookup shim branches for `runtime.StructDefinitionValue` and `runtime.TypeRefValue` inside `__able_resolve_qualified_callable` (`resolveReceiver`), routing receiver static resolution through shared `__able_member_get_method` + nil-filter path.
- Compiler tests: updated `TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim` and `TestCompilerRemovesTypeRefPointerQualifiedCallableShim` to assert direct resolver branch/lookup shim removal and shared member-get-method resolver path retention.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerQualifiedCallableShim|TestCompilerRemovesNilPointerQualifiedCallableShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.126s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.310s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 525.936s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 245.216s`).
- Compiler AOT: deduped `env.StructDefinition(head)` static lookup in `__able_resolve_qualified_callable` by normalizing the local struct-definition branch to a single canonical type-name path (`structTypeName`, derived from `def.Node.ID.Name` when present), removing duplicated local `lookupStatic(head)` dispatch while preserving resolver-level `lookupStatic(head)` fallback behavior.
- Compiler tests: updated `TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim` to assert canonical `structTypeName` lookup in the env struct-definition branch and exactly one remaining `lookupStatic(head)` fallback branch in the resolver helper segment.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerRemovesStructDefinitionPointerQualifiedCallableShim|TestCompilerRemovesTypeRefPointerQualifiedCallableShim|TestCompilerRemovesImplNamespacePointerQualifiedCallableShim|TestCompilerRemovesNilPointerQualifiedCallableShim|TestCompilerRemovesStructDefinitionPointerMemberGetMethodShim|TestCompilerRemovesTypeRefPointerMemberGetMethodShim|TestCompilerNormalizesInterfaceMemberGetMethodDispatch' -count=1` (`ok ... 0.136s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1` (`ok ... 60.165s`).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL=1 ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout=25m` (`ok ... 530.722s`).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (`ok ... 244.888s`).
- Planning hygiene: rewrote the `PLAN.md` Compiler AOT section from a long historical stream into a concise active backlog + definition-of-done checklist, with completed shim-slice history retained in `LOG.md`.
- Tests: `cd v12 && ./run_compiler_full_matrix.sh --typecheck-fixtures=strict` (`ABLE_COMPILER_EXEC_FIXTURES=all`, `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all`, `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all`, `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all`, fallback audit enabled) completed successfully with gate timings: `ok ... 530.066s`, `ok ... 530.185s`, `ok ... 536.592s`, `ok ... 480.375s`, `ok ... 31.196s`.
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerDynamicBoundary' -count=1` (`ok ... 60.913s`).
- Spec: closed tracked Compiler AOT contract gaps by expanding `spec/full_spec_v12.md` compiled-boundary coverage with explicit sections for static compile-failure semantics (no silent fallback), compiled runtime ABI contract (Array/BigInt/Ratio/String/Channel/Mutex/Future), compiled interface+overload dispatch model, compiled stdlib/kernel resolution requirements, and compiled<->dynamic boundary conversion/error rules.
- Spec TODO: cleared `spec/TODO_v12.md` Compiler AOT gap list after the above normative spec updates.
- Compiler AOT: **no-bootstrap execution path complete**. Non-dynamic programs now execute fully compiled — `interpreter.New()` is instantiated for runtime services, but `EvaluateProgram()` is never called. `TestCompilerNoBootstrapExecFixtures` validates the path: 222 pass, 13 fail (12 inherently dynamic/IO fixtures + 1 pre-existing 06_12_17 heap ordering), 5 skip out of 240 total.
- Compiler AOT: added `implSiblings` compile context for default interface methods in named impls. When a default method (e.g., `describe(self)`) calls `self.name()`, the compiler now detects the sibling method in the same named impl and generates a direct `__able_impl_self_method` call instead of falling through to `__able_member_get_method`. This fixed `10_05_interface_named_impl_defaults` in no-bootstrap mode.
- Compiler AOT: added `__able_impl_self_method` runtime helper that creates a `NativeBoundMethodValue` with correct bound-arity semantics (subtracts 1 from total arity to account for auto-injected receiver).
- Compiler AOT: `TestCompilerMainSkips` (7 tests) validates that generated `main.go` omits `EvaluateProgram()` for static programs importing stdlib.
- Compiler AOT: all definition-of-done criteria met — non-dynamic programs execute fully compiled with no interpreter execution; dynamic features execute only through explicit boundary paths; stdlib and kernel compile and execute directly; compiler fixture + stdlib compiled gates green in strict no-fallback mode; spec semantics parity preserved.
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_NO_BOOTSTRAP_FIXTURES=all go test ./pkg/compiler -run '^TestCompilerNoBootstrapExecFixtures$' -count=1 -timeout=20m` (222 pass, 13 fail, 5 skip).
- Tests: `cd v12/interpreters/go && ABLE_COMPILER_FALLBACK_AUDIT=1 ABLE_COMPILER_FIXTURES=all go test ./pkg/compiler -run '^TestCompilerExecFixtureFallbacks$' -count=1 -timeout=20m` (clean).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run '^TestCompilerMainSkips' -count=1` (7/7 pass).
- Tests: `cd v12/interpreters/go && go test ./pkg/compiler -run 'TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerBoundaryFallbackMarkerForStaticFixtures' -count=1` (pass).
