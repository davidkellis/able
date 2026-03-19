# Able Project Roadmap (v12 focus)

## Prompts

### New Session
Read AGENTS, PLAN, and the v12 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

### Next steps
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10/v11 toolchains available for historical reference while driving all new language, spec, and runtime work in v12.
- Use the Go tree-walking interpreter as the behavioral reference and keep the Go bytecode interpreter in strict semantic parity.
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical branches and the actively developed v12 runtime; document any deltas immediately in the v12 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md` + `spec/full_spec_v11.md`: authoritative semantics for archived toolchains. Keep them untouched unless a maintainer requests errata.
- `spec/full_spec_v12.md`: active specification for all current work; every behavioral change must be described here.
- `v10/` + `v11/`: frozen workspaces (read-only unless hotfix required).
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib-deprecated-do-not-use/`, `design/`, `docs/`). Canonical stdlib source is being moved to the external `able-stdlib` repo and cached via `able setup`.

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v12.md`; log discrepancies in `spec/TODO_v12.md`.
- **Go runtimes**: maintain tree-walker + bytecode interpreter parity; keep diagnostics and fixtures aligned.
- **Tooling**: build a Go-based fixture exporter; update harnesses to remove TS dependencies.
- **Performance**: expand bytecode VM coverage; add perf harnesses for tree-walker vs bytecode.
- **WASM**: run the Go runtime in WASM with JS tree-sitter parsing and a defined host ABI.
- **Stdlib externalization**: keep canonical stdlib in external git repo, auto-install into cache, and keep loader/resolver semantics aligned with spec.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `v12/design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `./run_all_tests.sh` (v12 default) must stay green for the Go suites and fixtures.
- `./run_stdlib_tests.sh` must stay green (tree-walker + bytecode).
- Diagnostics parity is mandatory: tree-walker and bytecode interpreters must emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`.
- After regenerating tree-sitter assets under `v12/parser/tree-sitter-able`, force Go to relink the parser by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- It is expected that some new fixtures will fail due to interpreter bugs/deficiencies. Implement fixtures strictly in accordance with the v12 spec semantics. Do not weaken or sidestep the behavior under test to "make tests pass".

## TODO (working queue: tackle in order, move completed items to LOG.md)
### Staged Integration Audit & Stabilization (active priority)
- Goal: land the currently staged runtime/compiler/CLI changes safely while keeping clean-checkout reproducibility and test guardrails intact.
- Immediate unit of work (execute in order):
  - [x] Remove accidental staged binaries (`v12/interpreters/go/able`, `v12/interpreters/go/able.test`) and ignore them.
  - [x] Make embedded kernel packaging reproducible from a clean checkout (track the `go:embed` payload files under `cmd/able/embedded/kernel`).
  - [x] Fix `v12/run_all_tests.sh` default mode so it remains green without hitting the default Go 10-minute timeout.
  - [x] Update `spec/full_spec_v12.md` and `spec/TODO_v12.md` for stdlib externalization (`able setup`, cache lookup order, global overrides).
  - [x] Update `AGENTS.md` + v12 README onboarding to reflect the new canonical stdlib flow and remove stale `v12/stdlib` assumptions.
  - [x] Fix compiler typed `=` assignment parity so `name: Type = value` reuses package/module bindings (when present) instead of always creating a local binding.
  - [x] Add a clean-environment smoke test that verifies `able setup` installs stdlib+kernel and both treewalker/bytecode can run a stdlib import fixture.
  - [x] Decide and document stdlib pinning policy (toolchain-pinned tag vs branch) and enforce it in dependency resolution + lockfile behavior.

### Runtime Performance Program (active priority: 10x targets)
- Goal: reach at least 10x speedup for bytecode and 10x for compiler relative to current benchmark baselines under `/home/david/sync/projects/benchmarks`.
- Execution order:
  - [x] Freeze a reproducible baseline snapshot (date, commit, machine profile, benchmark inputs) and check in structured results.
  - [x] Add a benchmark harness that emits machine-readable results for treewalker, bytecode, and compiled modes on the shared benchmark suite.
  - [x] Resolve baseline correctness blockers so benchmark statuses reflect performance limits rather than semantic/build failures (`matrixmultiply` stdlib helper import; `sudoku.solve` static-compileable body shape).
  - [x] Land first Bytecode Phase 1 call-dispatch fast path: inline `call`/`callname` now consume args directly from VM stack (no transient args slice on successful inline path).
  - [x] Add optional bytecode execution counters (`ABLE_BYTECODE_STATS`) with snapshot/reset APIs to guide hotspot work (opcode mix, name lookups, inline-call hit/miss).
  - [x] Add safe global-scope lookup cache for bytecode `LoadName`/`CallName`, keyed per VM instruction site with environment revision invalidation.
  - [x] Add inline slot-frame pooling for bytecode function calls to avoid per-call slot-slice allocation churn in recursive/hot-call workloads.
  - [x] Add regression coverage for lookup-cache invalidation on global rebinding plus runtime environment revision mutation semantics.
  - [x] Remove method-resolution miss-error allocations in hot paths (`Environment.Lookup` + resolver switch from `Get` to `Lookup` when probing scope callables).
  - [x] Reduce `resolveMethodFromPool` per-call allocation churn (drop map/closure candidate bookkeeping in favor of compact linear dedupe accumulator).
  - [x] Add conservative bytecode member-call inline cache (call-position member access only) with strict invalidation on global env revision + interpreter method-cache version.
  - [x] Add bytecode regression coverage proving member-call cache invalidates when unnamed `impl` definitions change dispatch without touching global bindings.
  - [x] Extend `ABLE_BYTECODE_STATS` with member-call cache hit/miss counters and wire instrumentation to cache lookup outcomes.
  - [x] Add bytecode regression coverage proving member-call cache counters snapshot/reset correctly (`BytecodeStats` / `ResetBytecodeStats`).
  - [x] Extend bytecode inline-call setup to support `BoundMethodValue` callees (inject receiver directly into slot frame) so call-position member dispatch can hit the no-args-slice inline fast path.
  - [x] Add bytecode regression coverage proving bound-method call sites record inline-call hits under `ABLE_BYTECODE_STATS`.
  - [x] Reduce call-dispatch fallback argument-slice churn in `callCallableValue` by avoiding unconditional `append(injected, args...)` allocation when no receiver injection is required.
  - [x] Update `invokeFunction` argument binding to use lazy writable copies (only when optional-arg fill/coercion needs mutation) and add regression coverage that host-provided arg slices are not mutated by coercion.
  - [x] Add conservative non-global current-scope lookup cache for bytecode `LoadName`/`CallName` (cache current-scope hits only, keyed per instruction site, invalidated by environment pointer + scope revision).
  - [x] Add regression coverage proving scoped lookup cache invalidates on local `=` rebinding (`CallName`) and local value reassignment (`LoadName`) within the same function activation.
  - [x] Extend dotted `CallName` fallback (`name.member`) to resolve the head receiver through the same cached-name path, reusing safe invalidation semantics for non-global/current-scope and global lookup sites.
  - [x] Add regression coverage proving dotted `CallName` receiver-head cache invalidates on local receiver rebinding at the same call site.
  - [x] Reduce partial-function dispatch merge churn by replacing two-step append concatenation with a single-pass merge buffer per partial call transition.
  - [x] Add single-overload runtime dispatch fast path (arity/type compatibility check + direct invoke) to skip full overload candidate scoring when only one callable overload exists.
  - [x] Remove unconditional generic-name-set map allocation in call dispatch (`functionGenericNameSet`) for non-generic functions/method sets.
  - [x] Share environment thread mode across lexical scope chains so child scopes stay lock-free in single-thread runs until first `spawn` flips the shared mode to multi-thread.
  - [x] Remove dotted `CallName` miss-error allocation churn by using non-error lookup probing before dotted fallback/head-member resolution.
  - [x] Expand slot coverage for recursive named calls by lowering stable self-recursive call sites to reserved slot loads (`LoadSlot+Call`) and bypassing `CallName` lookups in recursion hot paths.
  - [x] Skip `EnterScope`/`ExitScope` opcode emission for slot-enabled frames that do not require runtime environment scopes (`needsEnvScopes=false`).
  - [x] Add `bytecodeOpCallSelf` and lower stable self-recursive call sites to direct slot-indexed self calls (remove recursive callee `LoadSlot` stack churn while preserving inline-call fast paths).
  - [x] Memoize bytecode integer-literal range validation per instruction site (first execution only) so hot loops avoid repeated `ensureFitsInteger` checks while preserving lazy path semantics.
  - [x] Add `execBinary` numeric/operator fast path (`+`, `-`, `<`, `<=`, `>`, `>=`, `==`, `!=`) via `ApplyBinaryOperatorFast` with fallback to full operator dispatch for non-fast-path semantics.
  - [x] Add dedicated bytecode opcodes for hot integer binary operators (`+`, `-`, `<=`) and lower to them directly, with integer-specialized execution and fallback to full semantics for non-integer operands.
  - [x] Tighten `bytecodeOpCallSelf` inline setup with a dedicated self-call fast path (skip bound-method/general inline checks) and remove duplicate slot-frame clearing in frame-pool acquire/release cycle.
  - [x] Add slot+immediate integer opcodes for hot recursion forms (`slot - const`, `slot <= const`) and lower eligible slot-identifier/literal expressions directly to reduce repeated `LoadSlot`+`Const` dispatch.
  - [x] Add a single-parameter self-recursive inline shortcut in `tryInlineSelfCallFromStack` (skip generic param-loop setup when coercion is trivially unnecessary) for common `fib`-style recursion.
  - [x] Fuse self-recursive call lowering/execution for `f(slot - const)` into `bytecodeOpCallSelfIntSubSlotConst` to bypass arg stack traffic on the hot recursion shape.
  - [x] Harden interpreter method-cache synchronization for concurrent future execution (lock-guarded method cache map + version reads for bytecode member cache invalidation).
  - [x] Fast-path same-suffix integer arithmetic in bytecode hot ops (`+`/`-`) to skip repeated promotion/type-info lookups on recursive integer paths.
  - [x] Add hot-size slot-frame pool path to avoid per-call map lookup/insert churn in recursive bytecode call-frame reuse.
  - [x] Cache one-arg self-call inline metadata on frame layout (first-param type/simple-name) so fused self-recursive inline calls avoid repeated declaration/type-expression introspection.
  - [x] Add direct integer extraction fast path in bytecode specialized arithmetic (`bytecodeIntegerValue`) to avoid unnecessary unwrap work on common scalar cases.
  - [x] Bypass generic specialized-binary helper for slot+immediate recursion ops (`slot - const`, `slot <= const`) by executing direct integer paths before generic fallback.
  - [x] Execute hot specialized bytecode binary opcodes (`BinaryIntAdd/Sub/<=`) via direct `execBinary` opcode-specific paths with direct fallback operator dispatch, reducing helper/switch overhead in recursion loops.
  - [x] Restrict fused `CallSelfIntSubSlotConst` lowering to one-arg, no-type-arg, integer-coercion-safe layouts so runtime recursion can skip repeated generic/type-arg eligibility checks.
  - [x] Refactor bytecode call-frame push/pop into dedicated helpers and inline fused self-recursive `slot-const` frame setup against current layout to reduce per-call dispatch overhead.
  - [x] Add per-program sparse cache for decoded slot-const integer immediates and thread it through run-loop program switching; skip redundant same-program cache refreshes on inline self-recursive program switches.
  - [x] Reuse pre-boxed small integer runtime values in fused slot-const recursion paths (`CallSelfIntSubSlotConst`/`BinaryIntSubSlotConst`) to reduce repeated integer-interface boxing (`runtime.convT`) overhead.
  - [x] Extend boxed same-suffix int64 fast path to specialized integer `BinaryIntAdd`/`BinaryIntSub` opcodes and add direct slot-const immediate IP lookups (skip generic helper/switch on hot opcodes).
  - [x] Switch bytecode run-loop instruction fetch to pointer-based dispatch and keep hot handlers pointer-based (`execBinary`, fused self-call) to remove per-op `bytecodeInstruction` struct copies (`runtime.duffcopy`) in recursion loops.
  - [x] Inline binary stack pops in `execBinary` and remove call-frame tail clearing on pop; add direct-integer fast extraction for specialized binary opcodes to reduce hot loop dispatch overhead.
  - [x] Add `selfFast` inline-call frame flag for same-program/same-env recursion and skip redundant run-loop program switching on inline returns when the caller frame is known to remain in the current program/env.
  - [x] Reorder slot-const immediate decode in hot bytecode paths (`execBinary`, `execCallSelfIntSubSlotConst`) to read `instr.value` first and fall back to per-program immediate cache only when needed, removing hot-loop hash-map lookup pressure.
  - [x] Align bytecode serial scheduling with tree-walker synchronous-section semantics for non-async runs (`runResumable`), and add regression coverage proving `spawn` tasks do not run ahead of the main flow before explicit `future_flush`.
  - [x] Add `SerialExecutor.beginSynchronousSectionIfNeeded` and use it in bytecode run-loop entry so nested bytecode `vm.run` calls reuse the same sync section instead of repeatedly lock/unlock thrashing.
  - [x] Reduce call-frame hot-path overhead by preallocating bytecode VM call-frame capacity in `newBytecodeVM` and appending populated frame literals directly in `pushCallFrame`.
  - [x] Split `execBinary` dispatch by opcode class (slot-const specialized, specialized int opcodes, generic binary) so hot specialized integer opcodes bypass generic branch paths.
  - [x] Inline the `bytecodeOpReturn` stack pop in the VM run loop to remove `vm.pop()` call overhead on recursion return paths.
  - [x] Add conservative bytecode `Index.get` callsite cache for array receivers (keyed by instruction site + first-element type token, invalidated by global revision and interpreter method-cache version) to bypass repeated `findIndexMethod` churn without bypassing `Index` semantics.
  - [x] Extend conservative bytecode index-method caching to `IndexMut.set` + compound index assignment paths (`+=`, etc.), reusing strict invalidation semantics and preserving fallback array behavior when no index methods exist.
  - [x] Make bytecode call-frame backing storage lazy: allocate call-frame capacity on first push instead of eager `newBytecodeVM` preallocation to cut per-call VM allocation churn.
  - [x] Preserve bytecode per-program const caches (`validatedIntegerConstSlots`, slot-const immediate decode tables) across pooled VM resets so repeated calls avoid re-allocating instruction-sized cache state.
  - [x] Memoize function generic-name sets per `runtime.FunctionValue` so hot call dispatch avoids repeated generic-map allocation in `functionGenericNameSet`.
  - [x] Add cast identity fast paths for same-suffix primitive casts (`as i32`, `as f64`, etc.) and skip non-aliased primitive alias-expansion work in `castValueToType`.
  - [x] Return existing runtime values from same-type cast fast paths (`castValueToType`) to avoid repeated integer/float re-boxing allocations on hot `as` sites.
  - [x] Add allocation-light type-info/type-expression fast paths for primitive runtime values (cached simple AST nodes + direct `typeInfo` construction in `getTypeInfoForValue`) to reduce `ast.Ty`/parse churn in dispatch and type matching.
  - [x] Make runtime `Environment` map storage lazy (`values`/`structs` allocated on first write) to reduce per-call scope allocation pressure in bytecode function invocation.
  - [x] Refactor `expandTypeAliases` to preserve original type-expression nodes when alias expansion is a no-op (avoid unconditional generic/union/function node reconstruction in hot type-matching paths).
  - [x] Replace UFCS scope-membership map construction (`functionScopeSet`) with an allocation-light scope filter over existing function/overload pointers.
  - [x] Remove native call-dispatch boxing churn in `callCallableValue` (value-native fast path without pointer-escape temporaries; stack `NativeCallContext`; bound-native partial target normalization).
  - [x] Reuse cached boxed small integers for hot array scalar members (`storage_handle`, `length`, `capacity`) to remove repeated `runtime.NewSmallInt` allocation churn during member dispatch.
  - [x] Cache canonical alias-base expansion results for `canonicalTypeNames(...)` and invalidate on alias registration/import rebinding so hot method-resolution paths avoid repeated alias-chain reconstruction.
  - [x] Cache hot inferred generic type expressions for `Array<T>` and `Iterator<_>` in runtime type inference to eliminate repeated `ast.NewGenericTypeExpression` churn in dispatch/type matching.
  - [x] Reuse bytecode member-method callsite cache for dotted `CallName` fallback (`head.tail`) so repeated calls at the same instruction site avoid repeated bound-method reconstruction and resolver traversal.
  - [x] Replace hot type/call-path miss lookups (`Environment.Get`) with allocation-light probing (`Lookup`) in canonical type-name resolution and direct identifier call dispatch fallback, removing miss-error churn in bytecode-heavy runs.
  - [x] Reduce `runtime.NewEnvironment` churn by eliminating child-scope thread-mode throwaway allocations and reusing closure env for slot-enabled, non-generic bytecode calls that do not require runtime env scopes.
  - [x] Remove `typeExpressionToString`/`typeInfoToString` formatting churn (`fmt.Sprintf` + join slices) by switching to builder-based rendering for hot method/type paths.
  - [x] Remove generic type-argument slice copy churn in `parseTypeExpression` by treating AST argument slices as immutable in runtime resolution paths.
  - [x] Reduce bound-method allocation churn by returning value-form `runtime.BoundMethodValue`/`runtime.NativeBoundMethodValue` in hot method-resolution/member-cache paths.
  - [x] Add native-call arg borrowing metadata (`runtime.NativeFunctionValue.BorrowArgs`) and skip bytecode fallback arg-slice cloning for borrow-safe native call targets.
  - [x] Expand bytecode pre-boxed small-int cache upper bound (`4096` -> `16384`) to reduce hot integer boxing allocations in arithmetic recursion/loop paths.
  - [x] Add bounded dynamic boxed-int caching for out-of-range `i32`/`i64`/`isize` values and route specialized bytecode integer add/sub fast paths through it.
  - [x] Replace bytecode index-cache array element type key strings with compact numeric tokens to reduce hot cache-key hashing overhead while preserving element-type invalidation semantics.
  - [x] Return bytecode slot frames to the pool on all non-yield run exits (success + error unwind) and use pooled slot-frame acquire for top-level `invokeFunction` bytecode entry frames.
  - [x] Reuse bytecode string-interpolation part buffers across op executions and split literal/stack-op handlers out of `bytecode_vm_run.go` (keeps run loop under 1000 lines while reducing interpolation scratch allocations).
  - [x] Add direct `*runtime.FunctionValue` call-dispatch fast path in `callCallableValue(...)` (including bound-method function targets) to bypass overload flattening/scoring on common single-function calls while preserving mismatch diagnostics.
  - [x] Refactor method-resolution candidate accumulation to single-candidate-first storage (promote to slices only on ambiguity) to reduce `resolveMethodFromPool`/`callCallableValue` allocation churn in hot dispatch paths.
  - [x] Isolate quicksort hotloop memprofiles from one-time setup churn by suspending memory sampling during fixture/load/typecheck bootstrap and restoring before the timed call loop.
  - [x] Add an untimed quicksort hotloop warmup call before benchmark sampling/timer reset so one-time first-call cache/bootstrap work is excluded from steady-state perf and memprofile signals.
  - [x] Reduce call-dispatch allocation churn with zero-copy partial-arg merge shortcuts and overload-slice view reuse for `*runtime.FunctionOverloadValue` targets.
  - [x] Cache `typeInfo` generic signature strings used for method-cache keys so `findMethodCached(...)` avoids repeated `typeInfoToString(...)` allocations on hot generic receiver paths.
  - [x] Add capped pointer-receiver bound-method cache in `resolveMethodFromPool(...)` (keyed by receiver identity + method + interface filter + inherent gate) and clear it alongside method-cache invalidation.
  - [x] Add mutability-aware internal call dispatch (`callCallableValueMutable`) for bytecode-originated arg slices so `invokeFunction(...)` can skip defensive arg-slice cloning during coercion while preserving external/partial-call arg immutability guarantees.
  - [x] Add int64-first `div/mod` fast return path with boxed small-integer reuse in `evaluateDivMod(...)`/`evaluateDivModFast(...)` to reduce `%`/`//` result boxing churn in hot numeric loops.
  - [x] Reuse receiver-injection arg backing storage when bytecode passes mutable arg slices and pool `NativeCallContext` objects in call dispatch to reduce per-call allocation churn on hot native/member call paths.
  - [x] Extend bytecode slot fast paths to typed identifier declarations (`name: T := expr`) by lowering simple typed declarations to slot stores with typed-pattern coercion semantics preserved at runtime.
  - [x] Keep typed `=`/compound typed-pattern assignment semantics on `AssignPattern` paths (interface coercion + fallback binding behavior) while still enabling slot eligibility for typed `:=` declarations.
  - [x] Defer `getIntegerInfo(...)` map lookups off int64 arithmetic/div-mod fast paths and use `ensureFitsInt64Type(...)` directly until big-int fallback is needed.
  - [x] Cache boxed `u64` results for hot array metadata externs (`__able_array_size`, `__able_array_capacity`) and add an early primitive-receiver bound-method cache probe in `resolveMethodFromPool(...)` to reduce hot-loop allocation churn.
  - [x] Add specialized bytecode lowering/opcode for `(<int> / <int>) as <int>` with guarded fast execution + semantic fallback, and optimize dynamic `ArrayStoreWrite` append writes (`index == len`) to avoid nil-fill+overwrite churn.
  - [x] Pre-grow empty dynamic array append path to capacity 4 before first write so hot push loops avoid extra `cap=1`/`cap=2` realloc steps.
  - [ ] Bytecode Phase 1: remove remaining high-frequency environment/path lookups in hot loops (slot coverage expansion + call dispatch fast paths).
  - [ ] Bytecode Phase 2: cut allocation pressure (integer/array/hash map hot-path allocations, iterator churn, closure scaffolding in loops).
  - [ ] Compiler Phase 1: eliminate avoidable `runtime.Value` carriers in statically-typed locals, struct fields, and loop temporaries.
  - [ ] Compiler Phase 2: reduce bridge overhead at static call/member/index sites; prefer native typed paths and avoid dynamic helper round-trips.
  - [ ] Add perf guardrails (non-blocking CI report first, then optional thresholds once noise is characterized).
  - [ ] Publish per-phase progress in `LOG.md` with before/after timings for `fib`, `binarytrees`, `matrixmultiply`, `quicksort`, `sudoku`, and `i_before_e`.

### File-Size Maintainability (active)
- Goal: keep v12 implementation files under 1000 lines by splitting by cohesive responsibilities without semantic changes.
- Progress:
  - [x] Split compiler generator helpers out of `pkg/compiler/generator.go` (extern/diag and module-binding constant evaluation).
  - [x] Split compiler expression codegen out of `pkg/compiler/generator_exprs.go` (core dispatch vs call helpers vs cast/range/lambda helpers).
  - [x] Split IR lowering code out of `pkg/compiler/lowerer.go` (core lowering vs spawn/literal/loop lowering vs scope+emit helpers).
  - [x] Split interface runtime rendering out of `pkg/compiler/generator_render_runtime_interface.go` (main dispatch render path vs compiled-resolver emission block).
  - [x] Split concurrency runtime rendering out of `pkg/compiler/generator_render_runtime_concurrency.go` (core helper emission vs extern-wrapper emission block).
  - [x] Split IR Go emission out of `pkg/compiler/ir_codegen.go` (core instruction/literal emitters vs interpolation/destructure/terminator tail emitters).
  - [x] Split bytecode VM runtime loop out of `pkg/interpreter/bytecode_vm.go` (opcode definitions/entrypoint vs resumable loop + unwind/finalize helpers).
  - [x] Split runtime call rendering out of `pkg/compiler/generator_render_runtime_calls.go` (front half vs compiled-call/boundary/env/runtime-tail emission block).
  - [x] Split compiler exec fixture tests out of `pkg/compiler/exec_fixtures_compiler_test.go` (main fixture harness vs no-bootstrap fixture harness).
  - [x] Split bytecode VM tests out of `pkg/interpreter/bytecode_vm_test.go` (core VM tests vs async/member/collection tail tests and shared helpers).
  - [x] Split call-overload resolution helpers out of `pkg/interpreter/eval_expressions_calls.go`.
  - [x] Split dynamic/interface/package member resolution out of `pkg/interpreter/interpreter_members.go`.
  - [x] Split bytecode lowering support helpers out of `pkg/interpreter/bytecode_lowering.go` and run-loop program-switch helper out of `pkg/interpreter/bytecode_vm_run.go`.
  - [x] No remaining >1000-line `.go/.ts/.able` files under `v12/` (including tests), verified via `fd -e go -e ts -e able . v12 -x wc -l {} | grep -E '^[0-9]{4}'`.

### Compiler AOT
- Status: **COMPLETE**. All definition-of-done criteria met. History in `LOG.md`.
- No-bootstrap execution: non-dynamic programs run fully compiled (`interpreter.New()` instantiated for runtime services, `EvaluateProgram()` never called). Validated via `TestCompilerNoBootstrapExecFixtures`: 222 pass, 13 fail (12 inherently dynamic/IO + 1 pre-existing), 5 skip out of 240 total.
- Bootstrap skip detection: `TestCompilerMainSkips` (7 tests) verifies generated `main.go` omits `EvaluateProgram()` for static programs.
- Fallback audit: clean (`TestCompilerExecFixtureFallbacks` runs by default).
- Full matrix: `v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict` green.
- Spec: compiler AOT contract fully documented in `spec/full_spec_v12.md`.
### Compiler AOT Boundary Hardening (active priority)
- Goal: enforce v12 AOT contract that compiled static code does not use interpreter execution paths; interpreter usage is allowed only for explicit dynamic features (`dynimport`, `dyn.def_package`, `dyn.eval`, etc.).
- Immediate unit of work (execute in order):
  - [x] Make static fallback rejection the default for `able build` (require-no-fallbacks on by default for non-dynamic builds; keep explicit override for migration/debug).
  - [x] Wire compile-time policy: when dynamic features are not present, any collected fallback is a hard compile error (not warning/runtime boundary).
  - [x] Add/strengthen tests so static fixtures assert zero boundary fallback calls by default (remove env-gated audit behavior for core static checks).
  - [x] Keep dynamic fixtures explicit: dynamic-boundary tests must prove boundary calls only occur for explicit dynamic operations.
  - [x] Remove static fallback sites that currently route through interpreter evaluation (starting with local `methods` / `impl` statement evaluation paths).
  - [x] Remove static named/value call fallback to bridge interpreter dispatch; unresolved static calls must fail compile.
  - [x] Eliminate unconditional interpreter bootstrap in static generated `main.go`; static path must not require interpreter initialization.
  - [x] Update `spec/full_spec_v12.md` and `spec/TODO_v12.md` to reflect enforcement status and any temporary implementation limits.
- Definition of done for this workstream:
  - [x] Non-dynamic compiled programs execute without interpreter evaluation fallback calls (`__ABLE_BOUNDARY_FALLBACK_CALLS=0` in static audit runs).
  - [x] Non-dynamic compiled `main.go` omits interpreter bootstrap/eval paths and does not require interpreter-backed bridge operations for static semantics.
  - [x] Dynamic programs still function with explicit boundary transitions and retain parity with tree-walker/bytecode behavior.
  - [x] `./run_all_tests.sh` and compiler fixture audits stay green with the new strict policy.
### Compiler Native Lowering Contract (active priority)
- Goal: lower Able source to native Go encodings wherever semantics are statically representable; interpreter/runtime carriers are allowed only at explicit dynamic boundaries and ABI edges.
- Current status: the recent compiler-side `Array -> Elements []runtime.Value` rewrite is under audit. Treat it as an experimental intermediate state, not as an accepted final architecture.
- Vision / non-negotiable constraints:
  - [x] Record the target architecture in design docs and the plan (`v12/design/compiler-native-lowering.md`, `v12/design/compiler-monomorphization.md`, `v12/design/compiler-no-panic-flow-control.md`, `v12/design/compiler-union-abi.md`, `spec/TODO_v12.md`).
  - [ ] Arrays: compiled representation must be native Go array-backed storage (slice or compiler-owned wrapper), not `runtime.ArrayValue`, `ArrayStore*`, or kernel `storage_handle` plumbing on static paths.
  - [ ] Structs: compiled locals, fields, params, and returns must remain native Go structs/pointers; do not auto-box them into `runtime.Value` to preserve identity or dispatch.
  - [ ] Unions: compile to generated Go interfaces plus native variant carriers, not `any` or `runtime.Value`, except at explicit dynamic boundaries.
  - [ ] Flow control: compiled control flow must use ordinary Go conditionals and returns; do not use `panic`/`recover` or IIFEs for regular returns, breaks, continues, breakpoints, raises, or rescues.
  - [ ] Boundary discipline: when compiled code must cross into explicit dynamic behavior, perform a narrow adapter conversion at that edge and return to native Go carriers immediately after the boundary.
- Immediate unit of work (execute in order):
  - [ ] Audit and back out the staged array lowering that overrides kernel `Array` to `Elements []runtime.Value` while still depending on `runtime.ArrayValue` / `ArrayStore*`.
  - [ ] Define the compiler-native array ABI for static code (`literal`, `len`, `push`, `pop`, `get`, `set`, indexing, iteration, cloning, bounds/error semantics) without interpreter structural helpers.
  - [ ] Restore and strengthen compiler array-lowering regression coverage; static-path audits should fail if generated code reintroduces `runtime.ArrayValue`, `ArrayStore*`, IIFEs, or `panic`/`recover` flow control outside explicit dynamic boundaries.
  - [x] Remove struct-local auto-boxing to `runtime.Value` and preserve native identity through static field/method dispatch.
  - [ ] Design the native union ABI (generated Go interfaces, variant wrappers, type switches, pattern-match lowering, nullable/result interaction).
    - Audit status (2026-03-10/2026-03-11): the first native result slice is now landed for `ResultTypeExpression` shapes that normalize to the current two-member `runtime.ErrorValue | T` carrier form, plain `Error` type positions now also use that native carrier instead of `runtime.Value`, and `?Error` now lowers to `*runtime.ErrorValue`. The broader remaining typed union/result pattern and wrapper surface still relies on `bridge.MatchType(...)` / `__able_try_cast(...)`. Nullable pointer/slice forms are native, the compiler-native scalar nullable family lowers natively (`?bool`, `?String`, `?char`, `?f32`, `?f64`, `?isize`, `?usize`, `?i8`, `?i16`, `?i32`, `?i64`, `?u8`, `?u16`, `?u32`, `?u64`), and the first closed two-branch union slice is now native for direct `UnionTypeExpression` shapes plus named union definitions that normalize to the same two-member native carrier form.
    - Initial implementation order is now captured in `v12/design/compiler-union-abi.md`: start with nullable value carriers and closed two-branch native unions before widening to more generic shapes.
    - [x] Complete the broader native-union carrier widening tranche for this phase: multi-member nominal unions, generic alias unions that normalize to native nullable/result carriers, and interface/open unions that can keep non-native payloads in explicit `runtime.Value` residual union members now stay on generated native union carriers instead of collapsing the whole union to `any`.
    - Progress update (2026-03-10/2026-03-18): the native nullable-value slice now covers the compiler-native scalar family (`?bool`, `?String`, `?char`, `?f32`, `?f64`, `?isize`, `?usize`, `?i8`, `?i16`, `?i32`, `?i64`, `?u8`, `?u16`, `?u32`, `?u64`) plus `?Error -> *runtime.ErrorValue`. These shapes now use native Go pointer carriers on compiled static paths, explicit generated boundary helpers for wrapper/lambda conversion, native typed-`match` nil/payload lowering, and native `or {}` nil-branching instead of `any`. The first closed union pass is now also landed for two-member native unions: the compiler synthesizes generated Go interfaces plus wrapper carrier structs, static params/returns stay native on those shapes, wrappers/lambdas use explicit generated boundary helpers, typed union `match` lowering on static paths no longer falls back to `__able_try_cast(...)` for those carrier shapes, `or {}` now recognizes native `Error`-implementer failure branches on those carriers without routing the whole expression back through runtime-value probing, `case err: Error => ...` branches on those same carriers now discriminate natively too, converting the matched value to `runtime.Value` only at the branch binding edge when the whole error value is bound, and the first native `!T` slice now maps `ResultTypeExpression` to that same carrier family when the result shape normalizes to `runtime.ErrorValue | T`. Plain `Error` positions now also use `runtime.ErrorValue` on compiled static paths, and `?Error` uses `*runtime.ErrorValue`, so explicit `Error` returns, explicit `String | Error` unions, and nullable error paths no longer fall back to `runtime.Value`. Direct compiled `Error.message()` / `Error.cause()` calls now also stay native on the `runtime.ErrorValue` carrier, native concrete-error normalization preserves both compiled message and cause payloads, and struct field conversion now supports `Error` / `?Error` carriers without falling back to unsupported-field codegen. The broader carrier-widening tranche is now also landed for this phase: multi-member nominal unions, generic alias unions like `Option T = nil | T` / `Result T = Error | T`, and interface/open unions like `String | Tag` now use generated native union carriers; non-native residual members stay explicit as `runtime.Value` branches inside those carriers instead of forcing the whole union back to `any`, and singleton struct boundary converters now accept runtime `StructDefinitionValue` payloads so interpreted callers can pass bare singleton values into compiled native struct/union params. Fully bound object-safe interfaces now also stay on generated native Go interface carriers across static params/returns, typed local assignment, struct fields, direct method dispatch, concrete receiver `Index` / `IndexMut`, default-interface method calls, `Apply`, wrapper/lambda ABI conversion, and dynamic callback boundaries; the strict no-fallback interface fixture audit is green again end-to-end, `06_12_26_stdlib_test_harness_reporters` now has a dedicated regression harness, native interface runtime adapters now round-trip `void` as `struct{}` and write back mutated pointer-backed interface args after runtime dispatch, and `generator_match.go` has been split back under the file-size cap. The non-object-safe/generic interface existential tranche is now closed too: pure-generic interfaces now keep generated native carriers instead of collapsing to `runtime.Value`, generic interface/default-interface methods now keep the receiver on that native carrier and cross into runtime only at the explicit generic dispatch edge, inferred runtime results convert back to the best-known native carrier, and the strict interface lookup audit is green with total interface/global lookup counts forced to zero. The next category is now callable/function-type existentials plus further runtime-boundary tightening around callable-driven generic inference surfaces.
  - [x] Design and stage the explicit control-result envelope for non-local jumps and exceptions so callers can branch on normal return vs jump vs raise without `panic`/`recover`.
  - [x] Remove the residual panic-based dynamic helper internals (`__able_member_get_method`, `__able_member_get`, `__able_member_set`, and similar bridge-era helpers) so explicit dynamic boundaries no longer need narrow recover containment shims.
  - [x] Complete native lowering for non-object-safe/generic interface existentials without regressing the completed object-safe carrier slice back to `runtime.Value`.
  - [x] Design and stage a native callable/function-type existential ABI so interface method values and other callable-typed existential surfaces stop defaulting to dynamic carrier paths.
  - [x] Tighten the remaining runtime-boundary interface/callable surfaces so residual `runtime.Value` usage is explicit, minimal, and mechanically audited.
  - [ ] Re-audit Claude's recent compiler work against this contract before merging staged compiler changes.
  - Progress note (2026-03-10/2026-03-18): static compiled array locals now synthesize/use a built-in compiler `Array` carrier with spec-visible `length` / `capacity` / `storage_handle` fields plus hidden native `Elements []runtime.Value` storage; array literals, `push`, `write_slot`, direct index assignment, `clear`, and static array destructuring/rest bindings now lower to native slice mutation/binding plus metadata sync. Match expressions no longer blanket-convert struct subjects to `runtime.Value` before pattern dispatch, so static `Array` pattern conditions and rest tails stay native on direct compiled paths. The generated `Array` boundary helpers are also narrower now: compiled `Array -> runtime.ArrayValue` conversion keeps plain runtime array boundaries handle-free unless an existing handle is already present or a `StructInstanceValue` target explicitly requires one, while `Array <- runtime.ArrayValue` now reads existing handle state without re-ensuring the store. Residual dynamic array helpers (`__able_index`, `__able_index_set`, `__able_member_get`, `__able_member_set`) are now normalized around the shared runtime-array unwrapping shim, and current static compiler slices continue to bypass them for native `*Array` paths. Reachability coverage now explicitly proves the opposite side too: representative static native programs avoid the residual helper layer, while explicit dynamic package/member/index programs still reach it. Unannotated local struct declarations also now stay native instead of being declaration-time boxed into `runtime.Value`, targeted regression coverage confirms direct compiled struct params, returns, mutation-through-call sites, static array rest-pattern lowering, narrowed array boundary helpers stay native on static paths, and wrapper returns for native struct/array values now use explicit `__able_struct_*_to` boundary conversion instead of routing through `__able_any_to_value`. The native union/nullable slice now covers the compiler-native scalar nullable family (`?bool`, `?String`, `?char`, `?f32`, `?f64`, `?isize`, `?usize`, `?i8`, `?i16`, `?i32`, `?i64`, `?u8`, `?u16`, `?u32`, `?u64`) plus the native `Error` carrier family (`Error -> runtime.ErrorValue`, `?Error -> *runtime.ErrorValue`), the first native `!T` slice for `runtime.ErrorValue | T` result shapes, the broader carrier-widening tranche for multi-member nominal unions, generic alias unions, and interface/open unions with explicit residual `runtime.Value` members, the fully bound object-safe native interface tranche, the non-object-safe/generic interface existential tranche for pure-generic interfaces plus generic/default-interface methods, and now the callable/function-type existential tranche too. Function types now map to generated native callable carriers instead of `any`; direct lambdas, local function definitions, placeholder lambdas, bound method values, function-typed params, struct fields, wrapper arg/return boundaries, and native interface conversions all stay on those carriers on static compiled paths. Callable-heavy generic surfaces exercised by `Iterator<T>` / iterable fixtures and the stdlib reporters fixture now stay on the narrowed carrier/boundary design without falling back to broad dynamic helper paths. Singleton struct boundary shims now also accept runtime `StructDefinitionValue` payloads for bare singleton arguments. Explicit control-result propagation is now also landed for compiled static flow and explicit dynamic call boundaries: compiled functions plus generated `call_value` / `call_named` sites branch on `*__ableControl` instead of raw panic, callback-bearing runtime helpers convert that control back into ordinary Go `error` results, and dynamic boundary failure markers now survive callback failures instead of being lost to raw panics. The residual dynamic helper cleanup tranche is now complete too: generated `__able_member_get`, `__able_member_set`, `__able_member_get_method`, and `__able_method_call*` helpers now use explicit error/control returns instead of raw panic, the temporary `recover`-based bridge wrappers are gone, lambda callback arg conversion now preserves native interface carriers and nullable pointer-struct carriers while still rejecting nil for non-nullable struct params, and the full dynamic-boundary suite is green again. The strict interface/global lookup audit has also now been split across four deterministic default batch tests so each stays below the repo's one-minute target, while the unsuffixed selector test remains available for explicit fixture subsets via `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES`. Remaining work is now a different category: full compiler-work re-audit against this contract, broader performance-oriented specialization/monomorphization, and mechanical enforcement of the allowed dynamic-carrier touchpoints.
### Compiler AOT Performance and `runtime.Value` Reduction (active priority)
- Goal: minimize `runtime.Value` usage in static compiled code; keep it only where semantically required (explicit dynamic boundary crossing, interface/runtime-polymorphic dispatch, and ABI conversion points).
- Native-lowering requirement: static Able semantics should lower to host-native Go constructs (concrete structs/scalars/collections) rather than interpreter object-model execution paths; generic dynamic carriers are reserved for explicit boundary/ABI/polymorphic residual cases.
- Kickoff changes landed:
  - [x] Map statically-typed `Array ...` locals to compiled `*Array` instead of defaulting to `runtime.Value`.
  - [x] Fix local `=` declaration fallback so unbound local assignments do not compile to `__able_global_set/__able_global_get`.
  - [x] Lower typed `Array` index read/write (`arr[idx]`, `arr[idx] = v`) through direct `runtime.ArrayStore*` paths for compiled `*Array` receivers.
  - [x] Keep static no-fallback enforcement active while applying these optimizations.
- Immediate unit of work (execute in order):
  - [x] Add a static native-lowering audit for user-defined nominal types (struct/union/interface views) and primitive locals to identify/remove avoidable `runtime.Value` carriers in compiled function bodies.
    - Origin-type tracking (`OriginGoType` in `paramInfo`) enables static method/field resolution on `runtime.Value` struct locals.
    - `compileOriginStructMethodCall`: extracts struct, calls compiled method directly, writes back.
    - `compileOriginStructFieldAccess`: extracts struct, accesses Go field directly.
    - Default impl sibling dispatch: direct compiled calls via `compileResolvedMethodCall` instead of `__able_impl_self_method` + `__able_call_value`.
    - Struct literal → runtime.Value IIFE eliminated (inline lines in assignments).
  - [x] Add regression fixtures that assert struct-heavy static programs emit concrete Go typed locals/field access paths and avoid dynamic dispatch helpers (`__able_member_get_method`, `__able_call_value`, `__able_call_named`) outside explicit dynamic-boundary wrappers.
    - `TestCompilerStructMethodStaticDispatch`: struct method calls → direct compiled dispatch.
    - `TestCompilerStructFieldStaticAccess`: struct field access → direct Go field access.
    - `TestCompilerDefaultImplSiblingDirectCall`: default impl sibling → direct compiled call.
  - [ ] Encode/enforce allowed dynamic-carrier touchpoints in codegen (explicit dynamic boundary adapters, residual runtime-polymorphic dispatch, extern ABI conversion), with compile-time errors for new static misuse.
  - [x] Add call-site intrinsics for typed `Array` methods in hot paths (`push`, `len`, `get`, `set`) to bypass dynamic member lookup / `__able_call_value`.
  - [x] Add compiler regression fixtures proving typed-array locals in static code emit no `__able_global_get/__able_global_set` in compiled function bodies.
  - [x] Add compiler regression fixtures proving typed-array loops emit no `__able_member_get_method`/`__able_call_value` for `push/get/set/len`.
  - [x] Add fast-path loop lowering for `while` loops without explicit `break`/`continue`/`rescue` needs (avoid per-iteration closure + `defer` scaffolding).
  - [x] Extend array literal lowering so typed contexts keep native compiled `*Array` paths and avoid unnecessary struct<->runtime boxing.
  - [x] Audit `Array` stdlib compiled methods (`push`, `len`, `get`, `set`, `refresh_metadata`) for redundant runtime round-trips; remove avoidable metadata refresh churn.
  - [x] Add benchmark fixtures (`noop`, `sieve_count`, `sieve_full`) and track real/user/sys + GC count in CI-adjacent perf script (non-blocking, report-only).
  - [x] Document required/allowed `runtime.Value` usage categories in `spec/full_spec_v12.md` and note staged limits in `spec/TODO_v12.md`.
  - [x] Design and stage monomorphized container ABI proposal (`Array<T>` native element typing) with compatibility constraints before implementation.
  - [x] Add stage-0 mono-array flag scaffolding (`--experimental-mono-arrays`, `ABLE_EXPERIMENTAL_MONO_ARRAYS`) through compiler options and generated feature marker.
  - [ ] Revise the staged mono-array plan so the final static-path representation is compiler-native Go storage, not a `runtime.ArrayValue` / `ArrayStore*` hybrid.
  - [ ] Implement staged monomorphized array lowering behind a compiler flag once design/spec update is approved.
    - Partial stage-1 landed: runtime typed stores (`i32`, `i64`, `bool`, `u8`) + compiler lowering for typed array literals, index read/write, and `Array.push/len/get/set` intrinsics when static element type is known.
    - Runtime array capacity growth now uses amortized expansion to remove per-push reallocation thrash in both dynamic and mono paths.
    - Added mono-array boundary regression coverage for explicit dynamic calls (compiled callback conversion + compiled->dynamic->compiled array roundtrip under `--experimental-mono-arrays`).
    - Added mono-array boundary coverage for nullable/union/interface callback conversion shapes (success + failure) under `--experimental-mono-arrays`.
    - Captured compiled-only perf snapshot (5-run avg, 2026-02-26, after native index-int + propagation/cast de-boxing):
      - `bench/noop`: mono `0.060s` real / `3.20` GC vs default `0.060s` / `3.00` GC.
      - `bench/sieve_count`: mono `0.072s` real / `5.00` GC vs default `0.094s` / `10.80` GC.
      - `bench/sieve_full`: mono `0.156s` real / `22.40` GC vs default `0.376s` / `54.40` GC.
    - Landed index-int lowering optimization for array read/write/get/set paths: native integer index types now avoid `bridge.ToInt` + `bridge.AsInt` boxing round-trips.
    - Landed mono propagation/cast de-boxing for typed index reads (`arr[idx]! as i64` style paths) so mono read values stay native where semantically safe.
    - Landed compatibility fixes for mixed array carriers and interface-typed assignment coercion:
      - `Array` struct converters now accept/synchronize raw `*runtime.ArrayValue` carriers at explicit runtime boundaries.
      - Interface-annotated local assignment now applies `bridge.MatchType` coercion so interface args are preserved in compiled dispatch.
    - This staged runtime-backed path is not the final contract: the target design is compiler-native Go storage for static paths, with explicit boundary adapters only where dynamic behavior is entered.
    - Default-on rollout criteria (stage-1 gate):
      - strict compiler fixture audits pass in `ABLE_TYPECHECK_FIXTURES=strict` (`ExecFixtures`, `StrictDispatch`, `InterfaceLookupBypass`, `BoundaryFallbackMarker`, `ExecFixtureFallbacks`);
      - dynamic-boundary suites pass, including mono-array boundary tests;
      - compiled-only perf stability (5-run avg) stays within guardrails: mono real-time regression <= 10%, mono GC regression <= 15% vs default on `bench/noop`, `bench/sieve_count`, `bench/sieve_full`.
    - Gate status (2026-02-26):
      - strict fixture audits: PASS.
      - dynamic-boundary suites (`TestCompilerDynamicBoundary*`): PASS.
      - compiled-only perf (5-run avg):
        - `bench/noop`: default `0.062s` / `3.20` GC; mono `0.060s` / `3.20` GC.
        - `bench/sieve_count`: default `0.072s` / `5.40` GC; mono `0.074s` / `5.20` GC.
        - `bench/sieve_full`: default `0.164s` / `23.20` GC; mono `0.164s` / `23.00` GC.
      - Result: stage-1 rollout gate satisfied; default-on enabled in CLI flows with explicit opt-out (`--no-experimental-mono-arrays` / `ABLE_EXPERIMENTAL_MONO_ARRAYS=off`); remaining work is staged rollout mechanics and eventual flag retirement criteria.
- Definition of done for this workstream:
  - [x] Static typed-array hot paths (`push/get/set/len` + index ops) compile without dynamic member dispatch in generated function bodies.
  - [x] Static local-variable fallback semantics (`=` declares when unbound) stay local-scope and do not route through global environment helpers.
  - [x] Sieve-style benchmark shows measurable runtime and GC reduction versus pre-work baseline, with unchanged program output.
  - [x] No regressions in compiler strict static checks, fixture parity, and dynamic-boundary behavior.
### WASM
- WASM: prototype JS tree-sitter parsing that feeds AST into the Go/WASM runtime (**in progress**).
  - Landed staging scaffold: `cmd/ablewasm` (`GOOS=js GOARCH=wasm`) + `pkg/wasmhost` JSON bridge and `v12/wasm/` Node prototype (`web-tree-sitter` subset adapter + runner).
  - Next: broaden AST adapter coverage beyond the initial expression/import subset and wire it to the host ABI path in `v12/docs/wasm-host-abi.md`.
- WASM: build a minimal `ablewasm` runner (Node + browser harness) once the Go runtime builds to WASM.
- WASM: document the WASM deployment contract in `v12/docs/`.
### Regex syntax
- Regex syntax: add regex AST nodes and grammar in tree-sitter (quantifiers, groups, classes, alternation).
- Regex syntax: wire AST mapping for regex nodes in Go parser.
- Regex syntax: add fixtures/tests for regex AST output and exec behavior; keep stdlib engine parity.
- Regex syntax: add parser corpus cases that cover nested groups, alternation, and escaped quantifiers.
- Regex syntax: update `spec/TODO_v12.md` with remaining regex syntax/semantics gaps.
- Regex syntax: align stdlib regex implementation with parser outputs as grammar coverage expands.
