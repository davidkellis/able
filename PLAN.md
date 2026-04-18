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
- For compiler/AOT work, only primitive Able types may use primitive-specific Go lowering. All non-primitive nominal types must lower through shared struct/union/interface/generic translation rules and semantic encoding rules. Avoid new per-structure lowering branches for named non-primitive types unless the special handling is required by a language-level syntax or kernel ABI boundary rather than by the nominal type itself.

## Existing Assets
- `spec/full_spec_v10.md` + `spec/full_spec_v11.md`: authoritative semantics for archived toolchains. Keep them untouched unless a maintainer requests errata.
- `spec/full_spec_v12.md`: active specification for all current work; every behavioral change must be described here.
- `v10/` + `v11/`: frozen workspaces (read-only unless hotfix required).
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib-deprecated-do-not-use/`, `design/`, `docs/`). Canonical stdlib source lives in the external `able-stdlib` repo and is cached locally via `able setup`.

## Active Priorities
- **Compiler completion**: finish the Go AOT compiler first. This is the top
  priority and the main body of this plan is organized around it.
- **Bytecode performance**: closed on 2026-04-16. The Go bytecode interpreter
  now has the checked-in cross-mode baseline plus report-first guardrails
  described later in this plan.
- **Everything else**: parser/tooling/WASM/stdlib/testing-framework work stays
  in backlog unless it is directly required to unblock compiler completion, or
  unless a new post-compiler/post-bytecode priority is selected explicitly.

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
- Compiler/AOT optimization work must not introduce new bespoke lowering rules for specific non-primitive structures or containers. The correct fix direction is to improve the general lowering machinery so user-defined and stdlib nominal types fall out of the same rules.

## TODO (working queue: tackle in order, move completed items to LOG.md)

## Top Priorities

Priority order is fixed until changed explicitly:
1. Finish the compiler.
2. Make the bytecode interpreter fast.
3. Everything else is backlog.

The active plan below is intentionally organized around getting from the current
codebase to a production-grade compiler that correctly and efficiently compiles
Able to Go. Historical completed slices belong in `LOG.md`, not here.

### Compiler Completion Program (highest priority)

Goal:
- ship a production-grade Go AOT compiler that compiles non-dynamic Able code to
  direct Go implementations with no interpreter execution on static paths, keeps
  dynamic behavior behind explicit boundaries, and performs well on the checked-
  in benchmark family.

Canonical architecture docs:
- `v12/design/compiler-go-lowering-spec.md`
- `v12/design/compiler-go-lowering-plan.md`
- `v12/design/compiler-native-lowering.md`
- `v12/design/compiler-aot.md`

#### Current state snapshot

Compiler-native encoding completion is closed on 2026-04-14:
- large static slices of arrays, structs, interfaces, callables, joins, and
  control-flow now stay native;
- explicit dynamic-boundary audits exist;
- reduced benchmark fixtures exist for matrix, iterator, heap, and array-heavy
  paths;
- the array-native lowering tranche is complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is now limited to explicit dynamic
  or ABI edges plus the unspecialized wildcard-array ABI;
- imported shadowed nominal match bindings now preserve foreign package
  context through carrier reconstruction too, so direct field access stays
  native instead of round-tripping through nominal/runtime helpers;
- mixed imported/local shadowed nominal joins now keep distinct native union
  members instead of collapsing on unqualified type-expression strings, and
  shadowed callable joins built from those nominal returns now stay on native
  callable-union carriers instead of silently collapsing to
  `fn(...) -> runtime.Value`;
- lambda literals and placeholder lambdas now narrow through expected callable
  members inside native union/result carriers, and semantic `Result` carrier
  synthesis preserves that callable member's resolved package context too, so
  imported semantic `Option` / `Result` aliases and direct union aliases
  built from shadowed callable returns stay on native callable carriers
  instead of failing with `lambda expression type mismatch` or
  `placeholder lambda type mismatch`;
- imported semantic `Result` aliases over shadowed callable members now stay
  native under outer `Result` carriers too, because raw imported selector
  aliases nested inside function type expressions keep lexical caller-package
  normalization instead of being re-normalized under stale foreign package
  context and collapsing to `__able_fn_*_to_runtime_Value`; the same outer-
  result native-carrier path is now pinned for imported semantic `Option`
  aliases and imported generic union aliases over those shadowed callable
  members too;
- local parameterized union/result aliases now also have proof coverage for
  imported shadowed interface/callable actuals, so `Choice (RemoteReader i32)`
  and `Outcome (() -> RemoteThing)` style locals are pinned to native carriers
  instead of widening through `runtime.Value` / `any`;
- generic specialization now also has proof coverage for those imported
  shadowed alias actuals, so specialized helpers over `Choice (RemoteReader
  i32)` and `Outcome (() -> RemoteThing)` stay on native signatures instead
  of widening to `runtime.Value` / `any`;
- the shared carrier mapper and generic interface-method dispatch now also
  have proof coverage for those same imported shadowed alias actuals, so
  direct `lowerCarrierTypeInPackage(...)` lookups and existential
  `pass<T>(...)` style calls stay on native union/result/interface/callable
  carriers instead of broadening locals or helper signatures;
- imported generic interface-method calls now also normalize explicit type
  arguments in the lexical caller package and retry representable-carrier
  recovery while computing concrete param/return helper signatures, so
  imported `Echo.pass<Choice(RemoteReader i32)>(...)` and imported default
  generic method calls stay on native union/result/interface/callable
  carriers instead of widening through `runtime.Value` / `any`;
- imported generic interface default methods now also resolve nested
  selector-imported members inside those explicit type arguments before
  specializing the interface-package body, so calls like
  `tagged.tagged<Outcome(() -> RemoteThing)>(...)` synthesize the concrete
  native `Tagged<...>` carrier and avoid `__able_method_call_node(...)`;
- native union synthesis now also retries representable-carrier recovery in
  the member's resolved package before accepting a residual `runtime.Value`
  member, and the imported shadowed interface/callable alias families above
  are now pinned against hidden `_runtime_Value` union variants too;
- generic specialization now also retries representable-carrier recovery
  before rejecting a fully bound actual as broad, and the remaining imported
  shadowed result/interface plus union/callable specialization quadrants are
  pinned to native helper signatures too;
- imported shadowed nullable interface/callable aliases now stay on those
  native carriers through generic specialization too, and imported shadowed
  callable union aliases like `Choice (() -> RemoteThing)` now specialize
  through native union helpers instead of falling back through
  `runtime.Value`;
- proof coverage now also pins the adjacent broader imported-shadowed
  three-member alias surface, so `Choice3(RemoteReader i32)` and
  `Outcome3(() -> RemoteThing)` already stay on native
  union/interface/callable carriers too, with no hidden `_runtime_Value`
  helper variants;
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families through imported generic-interface dispatch and
  imported default generic methods, so
  `echo.pass<Choice3(RemoteReader i32)>(...)` and
  `tagged.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening helper
  signatures or falling back through `__able_method_call_node(...)`;
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families when the generic-interface receiver itself is a
  join-produced existential across concrete implementers, so joined
  `echo.pass<Choice3(RemoteReader i32)>(...)` and joined
  `tagger.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening the join
  local or defaulting the call back through runtime helpers;
- proof coverage now also pins broader outer-result shapes over those same
  imported-shadowed three-member alias families, so
  `!(Choice3(RemoteReader i32))` and `!(Outcome3(() -> RemoteThing))`
  already flatten/collapse to native union/result/interface/callable carriers
  too instead of regressing broader result/error families to
  `runtime.Value` / `any`;
- proof coverage now also pins the same broader outer-result families through
  generic interface shape synthesis itself, so parameterized carriers like
  `Keeper(!(Choice3(RemoteReader i32)))` and
  `Keeper(!(Outcome3(() -> RemoteThing)))` keep native union/result/interface
  signatures too instead of widening interface helper params/returns to
  `runtime.Value` / `any`;
- proof coverage now also pins the closed local interface existential family
  itself, so local aliases like `Either = (Reader i32) | Echo`,
  `Outcome = !Either`, and `Keeper<Either>` / `Keeper<Outcome>` helper
  synthesis stay on native union/result/interface carriers too instead of
  broadening local helper params/returns to `runtime.Value` / `any`;
- proof coverage now also pins the broader local multi-member
  interface/callable existential family, so local aliases like
  `Choice3 = (Reader i32) | Echo | String`,
  `Outcome3 = Error | (() -> Thing) | String`, generic interface dispatch
  over those same local families, and `Keeper<Choice3>` /
  `Keeper<Outcome3>` helper synthesis all stay on native
  union/result/interface/callable carriers too instead of broadening local
  params, locals, or helper signatures to `runtime.Value` / `any`;
- proof coverage now also pins the last local analogs of that broader family,
  so joined existential receivers calling `Echo.pass<Choice3>(...)` /
  `Tagger.tagged<Outcome3>(...)` stay on native carriers too, and local
  outer-result helper synthesis like `Keeper<!(Choice3)>` /
  `Keeper<!(Outcome3)>` also keeps native union/result/interface signatures
  instead of widening joined locals or helper params/returns to
  `runtime.Value` / `any`;
- native union synthesis no longer materializes partially native helper
  families when any member still only maps to `runtime.Value` / `any`, so
  hidden `_runtime_Value` variants stop leaking into adjacent imported-
  shadowed specialization slices;
- native interface method-shape collection, native interface impl-signature
  synthesis, and native callable signature materialization now also retry
  representable-carrier recovery after raw package-scoped mapping, so
  substituted imported shadowed alias families stay on native
  union/result/interface/callable carriers instead of broadening internally
  to `runtime.Value` / `any`;
- imported generic struct members with shadowed nominal type arguments now
  stay on specialized native carriers inside result and nested union/result
  families too, because fully bound imported selector arguments count as
  concrete in the caller package and foreign generic-struct specialization
  keeps that caller-side package context through field substitution; proof
  coverage now also pins that same native-carrier behavior through generic
  specialization over imported shadowed generic-struct result/union aliases,
  and imported generic-struct result/union aliases now also keep specialized
  native carriers when the generic argument is a native interface or native
  callable instead of falling back to base `*Box` plus dynamic member calls;
- local generic nominal carriers over normalized nullable/union/result members
  are now pinned too, so `Box MaybeReader`, `Box Choice`, `Box Outcome`,
  `Box !(Choice)`, and `Box !(Outcome)` all stay on specialized native
  `Box<...>` carriers instead of falling back to base `*Box`,
  `runtime.Value`, or `any`; this closes the deeper `types.go` /
  `generator_native_unions.go` carrier-synthesis cleanup;
- error-wrapped nominal struct typed matches now stay on those native struct
  carriers too, because generated `__able_struct_*_try_from(...)` /
  `__able_struct_*_from(...)` helpers unwrap through the shared
  `__able_struct_instance(...)` path before enforcing the nominal definition
  check, fixing static `case _: IndexError` matches on array helper bounds
  results under the no-bootstrap boundary harness;
- imported shadowed interface selector aliases now preserve their source
  package through generic type normalization too, so nullable / union / result
  aliases built from foreign `Interface<T>` members stay on native interface
  carriers even when the caller shadows the same interface name locally;
- raw imported selector-alias typed patterns now normalize in their lexical
  caller package before any recorded foreign package is reapplied, so generic
  `Result` / semantic-result carrier matches like
  `Outcome(RemoteIface<i32>) match { case value: RemoteIface<i32> => ... }`
  resolve back onto the native imported interface carrier instead of widening
  to `runtime.Value`;
- imported semantic `Result` aliases nested under outer `Result` carriers now
  stay native across shadowed foreign interfaces too, because alias expansion
  preserves alias-source normalization and builtin `Error` identities collapse
  across package contexts during nested carrier synthesis;
- generic specializations now also keep native interface carriers for both
  local and imported-shadowed interface actuals, including the duplicate-
  member collapse case where a fully bound generic union normalizes to the
  same foreign interface type twice, because imported selector source-package
  context now survives specialization materialization too instead of being
  dropped by no-op type substitution;
- join carrier selection now treats interfaces nominally instead of
  structurally, so unrelated same-shape interfaces stay on distinct native
  carriers and join through a native union rather than silently collapsing
  onto one interface family;
- propagated rescue identifier joins built from statically typed call failures
  now recover native callable plus imported shadowed nominal/interface
  carriers instead of widening back through `runtime.Value`, native-error
  unions, or member access fallback, and higher-order/unknown rescue call
  failures no longer misinfer from the callable return type so `err.value`
  style handlers stay on the dynamic error path instead of collapsing to
  arbitrary static return carriers;
- raised imported shadowed nominal struct literals now prefer the compiler's
  syntax-aware struct-literal type reconstruction during failure inference, so
  propagated rescue joins keep the foreign native struct carrier instead of
  collapsing onto same-named local nominals, while no-bootstrap raised
  non-`Error` bridge fallback stringification now stays aligned with
  interpreter-visible output for compiled rescue/string paths;
- nested struct-pattern field bindings now restore field-local expected type
  context before recursing into subpatterns, which keeps persistent map/set
  helper patterns on the correct native carriers, and dynamic typed-pattern
  casts now allocate temps on the caller stream so iterator-end /
  generator-yield matches stop colliding with surrounding temps during
  codegen;
- implicit and explicit return expressions now preserve the declared Able
  return `TypeExpr` while compiling the returned expression too, so generic
  return paths like `unwrap(io_read(...))` keep nullable success carriers
  instead of collapsing to their nil-capable Go carrier only;
- static nullable typed matches on nil-capable native carriers now guard both
  the non-nil typed branch and the `case nil` branch directly, so native
  interface and result-family whole-carrier matches no longer compile to dead
  `true`/`false` conditions ahead of the real nil arm;
- the concrete nullable typed-match nil-guard path is now narrowed to actual
  in-scope generics only, so `?Interface<T>` / `?Result<T>` typed arms regain
  their required non-nil guard instead of compiling as unconditional typed
  branches;
- nested native nullable/result outer unions now keep representable literals,
  struct literals, and typed-match clause ordering on native carriers too,
  because nested member wrapping is direct and match narrowing only removes a
  union member when the pattern exhausts that whole v12 member type rather
  than a non-nil subcase inside it;
- representable nested union/result members now flatten during carrier
  synthesis too, so outer unions like `(!T) | U` and `(A | B) | U`, plus
  direct nested result families like `!!T` and `!(A | B)`, lower to a single
  native union family instead of nesting one native union carrier inside
  another;
- fully bound duplicate union/result members now collapse to their single
  native carrier during mapping too, so generic specializations like
  `T | String` at `T = String` and `!T` at `T = Error` no longer miss native
  specialization behind synthetic union/runtime carriers;
- no-bootstrap / no-fallback enforcement is green across the release gate;
- the compiler release gates currently pass:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- the current dirty-tree compiler release rerun is green end-to-end again:
  bridge tests, all compiler core batches, outliers, the compiled exec-fixture
  matrix, strict-dispatch audit, interface-lookup audit, boundary audit, and
  `./run_stdlib_tests.sh` all pass after the latest rescue/failure-inference,
  persistent-collection, iterator-end, and bridge-fallback fixes;
- no-interpreter generic-interface dispatch now performs alias expansion and
  interface-constraint revalidation through compiler-emitted runtime metadata
  and generated helpers instead of interpreter registries or bridge fallback;
- the aggregated observed wall clock for the latest green compiler release
  rerun is now `52m51s` (`real 3171.27`), so the dominant remaining release-
  path issue is test runtime pressure rather than a known correctness blocker;
- the stronger compiler-native completion program and the bytecode performance
  program are now both closed; the remaining work returns to backlog /
  tooling-priority selection rather than another compiler or bytecode closure
  milestone.
- the clean-checkout reproducibility follow-up is now closed on 2026-04-17:
  `run_stdlib_tests.sh` self-bootstraps through `able setup` when needed, and
  the remaining active tooling/helper paths (`cmd/fixture`, interpreter
  fixture loading, `ablec`, generated compiled wrappers, and the repo fixture
  harnesses) now prefer explicit or cached stdlib installs before sibling
  `able-stdlib` probing; the rebased tree is green again at the top level with
  `/usr/bin/time -p ./run_all_tests.sh` = `real 1193.78` and
  `/usr/bin/time -p ./run_stdlib_tests.sh` = `real 36.36`.

#### Production definition of done

The compiler is only fully done for the stronger native-encoding goal when all
of these are true together:
- every statically representable Able type expression lowers to a native Go
  carrier;
- static control flow and pattern binding stay on native carriers;
- static field/method/interface/index/call dispatch lowers to direct compiled Go
  dispatch;
- dynamic/runtime carriers remain only at explicit dynamic or ABI boundaries;
- compiled runtime helpers implement language semantics directly in Go instead
  of modeling normal static execution in terms of interpreter behavior;
- no staged hybrid carrier architecture remains on static paths for arrays,
  unions, or other nominal values that should have final host-native compiled
  representations;
- experimental transition machinery is either removed or reduced to explicit
  boundary-only helpers instead of serving as the general static lowering path;
- compiler fixture parity is green under no-bootstrap/no-fallback enforcement;
- the compiled benchmark family is materially faster and free of already-known
  avoidable scaffolding on hot paths.

#### Milestone 1: Centralize Compiler Lowering Knowledge

Goal:
- make every carrier, dispatch, control, pattern, and boundary decision come
  from one shared synthesis point instead of emitter-local rules.

Status:
- complete on 2026-03-21.
- canonical lowering facade landed in:
  - `v12/interpreters/go/pkg/compiler/generator_lowering_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_boundaries.go`
- source-audit enforcement landed in:
  - `v12/interpreters/go/pkg/compiler/compiler_lowering_facade_audit_test.go`

Why this is first:
- without this, the compiler keeps accumulating one-off fixes and nominal-type
  drift.

Required work:
- [x] establish one canonical type-normalization path used by all codegen
      stages;
- [x] establish one canonical carrier-synthesis path used by all emitters;
- [x] establish one canonical join/pattern-synthesis path;
- [x] establish one canonical dispatch-synthesis path;
- [x] establish one canonical control-envelope synthesis path;
- [x] establish one canonical boundary-adapter synthesis path;
- [x] audit emitters and remove local fallback decisions that bypass those
      shared paths.

Proof required:
- generated-source audits that fail on new ad hoc carrier or dispatch fallback
  patterns.

#### Milestone 2: Native Carrier Completeness

Goal:
- every statically representable Able type expression maps to a native Go
  carrier everywhere it appears.

Required work:
- [x] remove remaining representable `runtime.Value` / `any` fallbacks from
      type mapping and carrier synthesis;
- [x] finish carrier synthesis for nullable, result, union, interface,
      callable, and fully bound generic nominal types;
- [x] ensure alias-expanded and recovered type expressions use the same carrier
      path as directly written types;
- [x] ensure representable branch/join locals do not regress to dynamic
      carriers just because a value was temporarily recovered from a broader
      path;
- [x] ensure residual `runtime.Value` union members exist only for true dynamic
      payloads.

Status:
- complete on 2026-03-22.
- landed the remaining shared carrier-synthesis closure in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_builtin_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_result_void.go`
- moved built-in `DivMod T` onto the shared nominal struct path instead of the
  old `any` fallback;
- tightened concrete union/result carrier synthesis so fully bound
  representable members fail fast instead of silently widening to
  `runtime.Value` / `any`;
- added shared native `Error | void` carrier support so `!void` signatures no
  longer regress to runtime carriers;
- added focused regressions for:
  - concrete `DivMod i32` carrier lowering
  - parameterized union/result alias locals containing `DivMod i32`
  - native `!void` return lowering
  - existing no-fallback fixture coverage including `06_01_compiler_divmod`
    and `06_01_compiler_match_patterns`

Proof required:
- representative carrier audits over arrays, structs, unions, interfaces,
  callables, results, and generics;
- generated-source scans for representable `runtime.Value` / `any` locals.

#### Milestone 3: Native Pattern And Control-Flow Completeness

Goal:
- all static pattern matching, branch joins, loop results, rescue/or-propagation,
  and non-local control stay native when the types are statically representable.

Status:
- complete on 2026-03-22.
- shared join inference now keeps `if`, `match`, `rescue`, `or {}`, `loop`,
  and `breakpoint` results on native carriers across representable mixed
  branches, nil-capable joins, common existential joins, and recovered
  `TypeExpr`-backed locals.
- typed pattern bindings now stay native when the narrowed carrier is
  recoverable, including rescue bindings and native-union whole-value
  interface bindings.
- representative static pattern/control bodies are now mechanically audited
  against `__able_try_cast(...)`, `bridge.MatchType(...)`, `panic`, `recover`,
  and IIFE-style scaffolding.

Required work:
- [x] remove runtime typed-pattern fallback for recoverable static subjects;
- [x] keep `if`, `match`, `rescue`, `or {}`, `loop`, and `breakpoint` joins on
      native carriers in all representable cases;
- [x] keep typed bindings and recovered branch bindings on native carriers;
- [x] finish explicit control-envelope propagation for helper boundaries and
      remove any remaining panic/recover/IIFE-style static control scaffolding;
- [x] ensure `raise`, `rethrow`, `ensure`, `!`, and `or {}` are implemented as
      proper compiled control/data semantics instead of interpreter-shaped
      escape hatches.

Proof required:
- fixture slices for `match`, `rescue`, `or {}`, loops, breakpoints, typed
  patterns, and propagation paths;
- generated-source audits that fail when static paths regress to runtime-type
  helpers.

#### Milestone 4: Native Dispatch Completeness

Status:
- complete on 2026-03-22.
- shared dispatch recovery now converts recoverable `runtime.Value` / `any`
  call/member/index targets back onto native carriers before dispatch.
- local concrete and interface `Apply` bindings now route through the shared
  static apply path instead of `__able_call_value(...)`.
- mixed-source pure-generic interface dispatch now prefers the more concrete
  compiled specialization instead of falling back to runtime method dispatch.
- representative dispatch coverage now lives in:
  - `v12/interpreters/go/pkg/compiler/generator_dispatch_recovery.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dispatch_completeness_test.go`

Goal:
- all statically resolved operations compile to direct Go dispatch.

Required work:
- [x] finish static field access/assignment lowering;
- [x] finish static method and default-method lowering;
- [x] finish static interface/default-method dispatch lowering for all
      representable generic cases;
- [x] finish static callable/bound-method/partial application lowering;
- [x] finish static index/get/set/apply lowering without dynamic helper
      dispatch;
- [x] remove residual dynamic helper dispatch from static call/member/index
      paths.

Proof required:
- combined-source dispatch audits;
- fixture slices covering structs, interfaces, callables, indexable types, and
  generic/default-method paths.

#### Milestone 5: Compiled Runtime Core Independence

Status:
- complete on 2026-03-22.
- static compiled kernel/runtime helper families now call direct Go `_impl`
  helpers on static paths instead of routing through `__able_extern_*` or
  helper-to-helper `__able_extern_call(...)` chains.
- zero-arg callable syntax and `Await.default` zero-arg callback
  specialization now stay on native callable carriers in compiled static code.

Goal:
- compiled runtime helpers used on static paths must implement Able semantics as
  normal Go logic, not as thin wrappers around interpreter-style machinery.

Required work:
- [x] audit every compiled runtime helper family used on static paths;
- [x] replace helpers whose normal static behavior is still modeled too closely
      after interpreter operations;
- [x] keep only true dynamic-boundary helpers dependent on runtime/interpreter
      object-model values;
- [x] ensure array/map/range/string/concurrency runtime services used by
      compiled code are direct Go implementations;
- [x] keep helper control propagation aligned with the explicit control-envelope
      model.

Proof required:
- source audit over emitted helper families;
- static fixture slices that exercise helper families without dynamic features.

#### Milestone 6: Boundary Containment And Static Cleanliness

Status:
- complete on 2026-03-22.
- the final explicit boundary helper set is now mechanically locked to:
  - `call_value` via `__able_call_value(...)`
  - `call_named` via `__able_call_named(...)`
  - `call_original` via generated original-wrapper calls
- representative static no-bootstrap fixture execution is now audited for:
  - zero fallback boundary calls
  - zero explicit dynamic boundary calls
  - zero interface/member lookup fallback calls
  - zero global lookup fallback calls
- representative static fixture batches now remain green under:
  - no-fallback boundary-marker harnesses
  - no-bootstrap boundary/lookup-marker harnesses
  - static generated-source boundary audits
- representative boundary-containment coverage now lives in:
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_containment_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_main_bootstrap_test.go`

Goal:
- make the dynamic boundary explicit, narrow, and mechanically enforced.

Allowed boundary categories only:
- `dynimport`
- dynamic package mutation / definition
- dynamic evaluation / metaprogramming
- extern / host ABI conversion
- explicit compiled <-> dynamic callback boundaries
- values already originating from dynamic runtime payloads

Required work:
- [x] enumerate and document the final allowed boundary helper set;
- [x] tighten adapters so conversion happens exactly at the edge and returns to
      native carriers immediately after;
- [x] remove residual dynamic leakage from static fixtures;
- [x] keep strict no-bootstrap/no-fallback/no-boundary audits green for static
      fixture families.

Proof required:
- boundary-marker harnesses;
- strict static fixture audits;
- generated-source audits around boundary helper usage.

#### Milestone 7: Compiler Performance Completion

Status:
- complete on 2026-03-22.
- added the reduced checked-in recursion benchmark family:
  - `v12/fixtures/bench/fib_i32_small/main.able`
- shared compiled callable/runtime env scaffolding now swaps package envs only
  when the target env differs from the current env, via:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_env_swap.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_env_helpers.go`
- representative generated code now uses the conditional env-swap path across:
  - compiled functions/methods
  - native callable wrappers
  - native array core methods
  - native interface generic dispatch helpers
  - iterator collect mono-array helpers
  - compiled future task entry
- representative performance proof and current numbers are now recorded in:
  - `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`

Goal:
- make the compiler-generated code fast on the checked-in benchmark family
  without violating the lowering rules.

Required work:
- [x] keep using reduced checked-in benchmarks to isolate hot shared lowering
      gaps;
- [x] remove only shared primitive/control/array/callable/dispatch scaffolding,
      never by adding named non-primitive fast paths;
- [x] remeasure after each material compiler workstream;
- [x] keep benchmark proofs paired with generated-source shape audits.

Primary benchmark families:
- matrix / array hot paths
- iterator / generic container pipelines
- nominal generic method/container paths
- recursion/call overhead microbenchmarks

Definition of done for this milestone:
- [x] hot compiled benchmark paths no longer carry already-identified avoidable
      scaffolding;
- [x] checked-in benchmark baselines and current numbers are up to date.

#### Milestone 8: Compiler Release Validation

Goal:
- turn the above architecture into a hard release gate.

Status:
- complete on 2026-03-30.
- release validation is now green under the real top-level gates:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- the milestone-closing fixes were shared semantic fixes, not nominal
  special-cases:
  - range expressions inferred through `Iterable<T>` instead of incorrectly
    recoercing compiled ranges through nominal `Range<T>` carriers
  - native interface default-method dispatch now preserves concrete wrapped
    receiver overrides instead of eagerly short-circuiting to default helpers
  - numeric operators now accept unions whose members are all numeric and
    resolve them through pairwise promotion/normalization instead of rejecting
    them as non-numeric

Required work:
- [x] keep one authoritative lowering spec and one authoritative completion plan
      in sync with implementation;
- [x] keep compiler fixture parity green under no-bootstrap/no-fallback rules;
- [x] run the full compiler matrix and stdlib suite in compiled mode as a
      release gate;
- [x] ensure diagnostics and failure behavior are stable enough for production
      use;
- [x] confirm reproducible build trees and clean-checkout behavior for the
      compiler toolchain.

Release gate checklist:
- [x] `./run_all_tests.sh` green
- [x] `./run_stdlib_tests.sh` green
- [x] full compiled fixture matrix green
- [x] strict static no-fallback/no-boundary audits green
- [x] benchmark baselines updated
- [x] no known representable static-path regressions remaining in PLAN

#### Compiler Program Status

Compiler release validation is closed, but compiler-native encoding completion
remains the active highest-priority work. Bytecode performance does not start
until the staged hybrid/static-lowering gaps below are closed.

#### Compiler Native Encoding Completion (active follow-on)

Goal:
- finish the stronger compiler end-state where statically representable Able
  constructs lower to final Go-native encoded constructs rather than to staged
  hybrid carriers.

Status:
- array-native lowering tranche complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is limited to explicit dynamic or
  ABI edges plus the unspecialized wildcard-array ABI.
- residual representable union/result/interface carrier-synthesis cleanup
  complete on 2026-04-14; representable static carrier families now only
  admit `runtime.Value` / `any` at explicit dynamic/open or ABI edges.
- mono-array transitional runtime-store scaffolding cleanup complete on
  2026-04-14; compiler-generated mono-array wrappers are now pure slice
  carriers, mono-array field access (`length`, `capacity`, `storage_handle`)
  stays native on those wrappers, and `runtime.ArrayValue` / `ArrayStore*`
  remain only explicit dynamic or ABI boundary machinery.

Required work:
- [x] finish eliminating residual representable union/result/interface lowering
      paths that still rely on `runtime.Value` / `any` members outside explicit
      dynamic or ABI boundaries;
- [x] retire the remaining transitional mono-array/runtime-typed-store
      scaffolding now that static compiled arrays use compiler-native carriers
      by default;
- [x] decide and document the final no-interpreter policy for alias /
      constraint revalidation in generic interface dispatch, then make the
      implementation match that policy;
- [x] rerun the compiled release gates after each material native-encoding
      closure so the stronger finish line is enforced, not just the milestone-8
      release gate.

Proof required:
- source audits showing no staged hybrid carrier shapes remain on static paths
  where final host-native encodings are expected;
- fixture and generated-source coverage proving arrays, unions, interfaces,
  patterns, and dispatch stay on final native carriers;
- top-level release gates still green after each closure step.

Status:
- compiler-native completion is closed on 2026-04-14;
- bytecode performance is closed on 2026-04-16.

### Bytecode Performance Program (second priority; start after compiler-native completion work is closed or paused explicitly)

Goal:
- make the Go bytecode interpreter fast enough to be a practical execution mode
  after the compiler is finished.

Current state snapshot:
- Bytecode Milestone 1 is closed on 2026-04-16:
  - the remaining high-frequency dispatch/lookup scaffolding work is now
    closed across name lookup, member/index caches, slot stores, inline-call
    setup, direct raw-array access, and hot integer compare/add slot-const
    paths;
  - the last correctness blocker on that path was a bytecode-only inline
    return coercion gap for impl-generic interface returns, now fixed by
    threading the callee's generic-name set through bytecode call frames so
    return coercion keeps method-set generics instead of validating
    `Iterator T` / `Enumerable T` style returns as fully concrete;
  - focused interpreter coverage, the full `pkg/interpreter` Go test run, and
    `./run_stdlib_tests.sh` are green on this tree;
  - a fresh local quicksort hotloop spot-check after the fix is
    `10735729 ns/op` (`go test ./pkg/interpreter -run '^$' -bench
    '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchtime=50x -count=1`);
  - Bytecode Milestones 2, 3, and 4 are now also closed on 2026-04-16:
    allocation-pressure cleanup is done, collection/async-specific hidden
    overhead is reduced, the checked-in bytecode-core cross-mode baseline is
    current, and report-first guardrail tooling now compares fresh runs
    against that baseline without enforcing premature thresholds;
- the old compiled-mode CLI blockers in `cmd/able` / `cmd/ablec` are now
  cleared in this tree, and the downstream compiler follow-on regressions
  they exposed are now fixed too:
  - matcher/interface boundary helper finalization now materializes concrete
    sibling interface families before emitting native boundary helpers, so
    compiled matcher coercions keep concrete sibling adapters instead of
    falling back at boundary sites;
  - assignment binding metadata now reconciles concrete native carriers back
    into stored type expressions when rescue/join inference had only retained
    a broader wrong-package union/result wrapper, so imported shadowed nominal
    rescue joins no longer regress later member dispatch to stale local
    carrier metadata;
  - the repo-level `./run_all_tests.sh` gate is green again on this tree
    after those fixes (`real 1399.73`);
  - after the rebase exposed compiler package timeout pressure again, the
    heavy independent compiler fixture/parity/audit subtests now run in
    parallel, which brought `TestCompilerExecFixtures` down to about
    `real 203.36`, the full `pkg/compiler` package back under the default
    30-minute timeout (`1230.138s`), and the top-level repo gate back to
    green on this tree (`real 1287.77`);
- a large amount of call-dispatch, lookup-cache, frame-pool, integer-op, and
  hotspot work is already landed;
- the first post-compiler bytecode tranche is now landed too: repeated array
  index-method sites use a single-entry hot inline cache before the broader
  map cache, and the quicksort hotloop CPU profile no longer shows
  `indexMethodCacheKey(...)` among the active runtime hotspots;
- the next bytecode tranche is landed too: array-handle tracking now keeps a
  single tracked `ArrayValue` fast path and only promotes to an alias set when
  multiple wrappers share a handle, so `syncArrayValues(...)` dropped out of
  the same quicksort hotspot profile;
- the next bytecode tranche is landed too: repeated call-position member
  lookups now hit a single-entry inline member-method cache before the broader
  per-VM map cache, and receiver identity is pinned so same-name methods on
  different struct definitions cannot cross-hit;
- the next bytecode tranche is landed too: repeated non-local `LoadName` /
  `CallName` sites now cache lexical-owner hits (including captured parent and
  global bindings) with a single-entry inline hot path, so runtime
  `Environment.Lookup` is no longer a meaningful quicksort hotspot;
- the next bytecode tranche is landed too: already-tracked array wrappers now
  reuse their tracked array state directly instead of re-entering
  `ArrayStoreEnsure(...)` on hot reads, which moved array-state/store work out
  of the quicksort hotspot set;
- the next bytecode tranche is landed too: cached plain-array index sites now
  stay on a bytecode raw-array fast path instead of bouncing through the
  generic interpreter index dispatcher, which materially cut the quicksort
  index hot path again;
- the next bytecode tranche is landed too: exact-arity native calls and
  native bound-method calls now stay on a VM-side fast path instead of always
  routing through the generic callable dispatcher, which cut the remaining
  quicksort call-dispatch path again and keeps receiver injection plus
  non-borrowing arg stability pinned under bytecode;
- the next bytecode tranche is landed too: index-method caching now stores
  per-program / per-IP cache slots instead of routing hot array index sites
  back through a composite-key hash map, which removes the remaining
  `bytecodeIndexMethodCacheKey` hashing cost from the quicksort profile;
- the next bytecode tranche is landed too: array receiver identity now reuses
  a cached element-type token on shared array state, and single-threaded
  bytecode cache probes now read the method-cache version without taking the
  method-cache lock, which removed `currentMethodCacheVersion()` from the
  quicksort hotspot set and shrank the remaining index-identity path again;
- the next bytecode tranche is landed too: `execCall` and `execCallName` now
  resolve exact native call targets before attempting inline bytecode frame
  setup, so exact native call sites stop paying the inline-probe miss path and
  focused stats coverage now pins that those sites no longer contribute inline
  hit/miss counters;
- the next bytecode tranche is landed too: direct raw-array index `get` / `set`
  now reuse the tracked shared `ArrayState` pointer when the receiver already
  carries a valid tracked handle, and direct integer index decoding now stays
  on the small-int fast path, which moved the quicksort hot loop back into the
  high-17ms band without changing cache or invalidation semantics;
- the next bytecode tranche is landed too: name, member-method, and
  index-method VM cache probes now read environment/global revision state
  through a single-thread runtime hint, and bytecode method-cache probes read
  the interpreter method-cache version directly on the same single-thread path,
  which cut the remaining revision/version bookkeeping around hot cache checks
  and pushed the quicksort hot loop back down to roughly `17.5ms/op` on the
  longer local spot-check;
- the next bytecode tranche is landed too: already-cached raw-array no-method
  index sites now bypass `resolveCachedIndexMethod(...)` entirely and jump
  straight to direct array access under the same inline cache guards, and a
  same-session 5x50x local A/B check moved the quicksort hotloop distribution
  from roughly `20.0-24.3ms/op` without the bypass to `19.8-20.4ms/op` with
  the bypass while keeping recursive array-parameter self-call coverage pinned;
- the next bytecode tranche is landed too: when no `Array` `Index` /
  `IndexMut` impls exist at all, raw-array bytecode indexing now skips the
  index-method cache layer completely and jumps straight to direct array
  access, while focused “impl appears later” coverage stays green; this pushed
  the quicksort hot loop down again to roughly `15.9ms/op` on the 50x local
  spot-check;
- the next bytecode tranche is landed too: successful bytecode call paths no
  longer eagerly materialize eval-state/runtime-context data just to complete a
  call; `execCall`, `execCallName`, and related exact-native paths now resolve
  `stateFromEnv(...)` only if an actual error needs runtime-context attachment,
  which moved the quicksort hot loop to roughly `16.2ms/op` on a fresh 50x
  local spot-check while leaving error semantics unchanged;
- the next bytecode tranche is landed too: bound generic method calls whose
  receiver is already concretely injected may now use the existing bytecode
  inline call fast path instead of being conservatively forced through full
  `invokeFunction(...)` dispatch; that specifically unlocked concrete generic
  receiver calls like the hot `Array.push(...)` path and moved the quicksort
  hot loop to roughly `13.4ms/op` on a fresh 50x local spot-check;
- the next bytecode tranche is landed too: simple `LoadName` / `CallName`
  sites now record at lowering time that they are plain identifier lookups, so
  the VM can skip repeated runtime dotted-name/cacheability checks and go
  straight to the hot lexical/global/scope cache path; that cut the remaining
  named-call overhead again and moved the quicksort hot loop to roughly
  `13.3ms/op` on a clean 50x local rerun while pushing `execCallName` out of
  the top-tier hotspot set;
- the next bytecode tranche is landed too: simple named-call cache entries now
  live behind stable pointers, so hot `CallName` hits stop copying full
  dispatch records back and forth through the inline cache and map cache; the
  focused call-name invalidation slice stayed green, repeated 50x quicksort
  reruns landed around `10.6-11.2ms/op`, and `lookupCachedCallName(...)`
  dropped out of the top hotspot set on the next profiled run;
- the next bytecode tranche is landed too: direct small-int pair helpers now
  sit in front of the bytecode integer compare/add/sub hot paths, so common
  small-int pairs stop paying the older repeated `IntegerValue` extraction work
  before comparison and same-type arithmetic; repeated 50x quicksort reruns
  landed around `10.4-10.8ms/op`, and the old direct integer-compare hotspot
  no longer shows up as a top flat frame on the next profiled run;
- the next bytecode tranche is landed too: small slot-frame layouts now batch
  prefill the bytecode hot frame pool, so recursive inline calls can draw
  several same-size frames from the pool on the way down instead of allocating
  one frame per depth step; focused slot-frame/call coverage stayed green,
  repeated 50x quicksort reruns stayed in the same `10.5-11.2ms/op` band, and
  the profile cut `acquireSlotFrame(...)`, `tryInlineResolvedCallFromStack(...)`,
  and `execCallName(...)` materially on the next run;
- the next bytecode tranche is landed too: untyped `StoreSlot` sites now take
  a direct VM fast path instead of always routing through the typed-assignment
  helper, so ordinary local slot stores stop paying the typed pattern/coercion
  helper call on every write; focused slot-store coverage stayed green,
  repeated 50x quicksort reruns stayed roughly in the `10.2-11.3ms/op` band,
  and `execStoreSlot(...)` dropped out of the hotspot tier on the next profile;
- the next bytecode tranche is landed too: direct array index decoding now
  inlines the 64-bit small-int case at the outer value switch, so common value
  and pointer-backed integer indexes stop bouncing through the extra
  `IntegerValue -> int` helper on hot array `get` / `set` sites; repeated 50x
  quicksort reruns stayed roughly in the `10.2-11.0ms/op` band, and the old
  `bytecodeDirectArrayIndexFromInteger(...)` hotspot dropped out of the top set
  on the next profile;
- the next bytecode tranche is landed too: slot-const integer bytecode
  instructions now carry typed integer-immediate metadata directly, so hot
  `x +/- const` and `x <= const` loops stop re-decoding `instr.value` through
  the generic runtime-value path on every iteration; the focused lowering and
  cache slice stayed green, the profiled 50x quicksort run reached
  `9917425 ns/op`, and the old slot-const immediate decode hotspot dropped out
  of the top tier on the next profile;
- the next bytecode tranche is landed too: direct bytecode array writes now
  use an alias-aware tracked-array write sync path, so exclusive wrappers skip
  the old post-write alias-broadcast walk while shared aliases still stay
  coherent and element-type tokens stay current; the focused tracking/index
  slice stayed green, repeated 50x quicksort reruns landed around
  `9.90-10.00ms/op`, and the old `syncArrayValues(...)` / `resolveIndexSet(...)`
  hotspot pair dropped out of the top tier on the next profile;
- the next bytecode tranche is landed too: identifier member-access sites now
  carry their member name directly in bytecode, and `execMemberAccess` only
  probes/stores the member-method cache when a site can actually use it; that
  removes the remaining futile member-cache work from plain field/property
  accesses and drops `execMemberAccess` out of the hotspot tier, while the
  residual named-call cost is now dominated by inline-call value coercion
  rather than by ordinary/member dispatch scaffolding;
- the next bytecode tranche is landed too: slot-layout analysis now caches
  simple parameter type names for all inlineable parameters, and inline bytecode
  calls now use a primitive-only fast coercion path for simple integer widening
  plus integer/float coercions before falling back to the general type
  coercion machinery; that cuts the targeted `tryInlineCallFromStack(...)`
  coercion profile materially even though whole-benchmark wall-clock remains in
  the same noisy low-13ms band on this machine;
- the next bytecode tranche is landed too: ordinary general integer
  comparisons now bypass the shared fast-operator dispatcher in the bytecode VM
  and stay on a direct integer compare path, while the shared fast evaluator
  also shortcuts exact integer/string comparisons before it falls back to the
  generic comparison machinery; that moved the quicksort hot loop down again
  into the low-12ms band on local 20x/50x spot-checks and pushed the remaining
  visible cost toward slot coercion / cast work rather than comparison
  dispatch;
- the next bytecode tranche is landed too: inline bytecode calls now skip
  `coerceValueToType(...)` entirely when the declared parameter type is a
  guaranteed no-op coercion shape (for example `Array i32`), and simple
  primitive casts now short-circuit before alias/type-metadata checks on hot
  same-type paths such as `as i32`; that kept the quicksort hot loop in the
  high-11ms band on local 50x runs and pushed `castValueToType(...)` out of
  the hotspot tier;
- the next bytecode tranche is landed too: slot layouts now cache per-parameter
  inline-call coercion metadata, and the VM’s hot inline call path now uses
  that cached data instead of recomputing generic/no-op coercion guards on
  every recursive call while also separating the already-bound receiver path;
  this pushed the quicksort hot loop down again to roughly `11.0ms/op` on the
  50x profiled local run and dropped `tryInlineCallFromStack(...)` out of the
  hotspot tier;
- the next bytecode tranche is landed too: slot store instructions now carry
  typed-identifier assignment metadata directly in bytecode, and the VM’s
  store-slot path uses that metadata instead of reopening assignment AST nodes
  and pattern helpers on every store just to discover most stores are untyped;
  this removed `execStoreSlot(...)` / `typedSlotAssignmentValues(...)` from the
  hotspot tier and kept the quicksort hot loop in the mid-11ms band on local
  50x reruns;
- the next bytecode tranche is landed too: simple `CallName` sites now keep a
  revision-guarded call-site cache for the resolved callee plus its dispatch
  shape (`exact native`, `inline bytecode`, or `generic`), so repeated named
  calls no longer redo lexical lookup validation and call-target
  classification on every hit; focused invalidation and dispatch-kind rebinding
  coverage is green, the profiled 50x quicksort run reached roughly
  `10.7ms/op`, and repeated clean 50x reruns stayed around `11.1-11.4ms/op`
  with one noisy outlier;
- the next bytecode tranche is landed too: the direct integer-comparison fast
  path and the shared comparison helper now short-circuit small-int pairs
  before they fall through the general `ToInt64()` / big-int comparison route,
  which trims the remaining generic comparison helper overhead without widening
  the bytecode opcode surface; focused parity coverage is green,
  `execBinaryDirectIntegerComparisonFast(...)` dropped from roughly `100ms`
  cumulative to roughly `80ms` on the profiled quicksort run, and repeated
  local 50x reruns stayed in the same low-11ms band;
- the next bytecode tranche is landed too: hot `x + const` integer sites now
  lower onto the existing slot-const immediate path instead of reloading and
  unboxing the constant operand through the ordinary specialized binary route;
  focused lowering/parity coverage is green, the profiled quicksort run moved
  `execBinarySpecializedOpcode(...)` from roughly `70ms` cumulative down to
  roughly `30ms`, and repeated local 50x reruns stayed in the low-11ms band
  with the best reruns landing around `10.8ms/op`;
- the next bytecode tranche is landed too: primitive-receiver method
  resolution now checks the existing bound-method cache before paying
  `env.Lookup(...)`, and single-threaded bytecode runs now read/write that
  bound-method cache without taking the interpreter method-cache lock; this
  removed `resolveMethodFromPool(...)` / array member dispatch from the
  hotspot tier on the next profiled quicksort run, while repeated local 50x
  reruns landed in the roughly `10.4-11.2ms/op` band;
- the next bytecode tranche is landed too: direct array index decoding now
  keeps concrete integer receivers on a narrower small-int/int64 path and
  skips the old extra generic integer extraction on already-concrete index
  values; on 64-bit builds that also removes the redundant int-range guard for
  small ints. Focused index-cache coverage is green, the profiled quicksort
  run moved `bytecodeDirectArrayIndex(...)` from roughly `70ms` flat down to
  roughly `40ms` cumulative, and repeated local 50x reruns stayed in the
  roughly `10.6-11.1ms/op` band;
- the next bytecode tranche is landed too: slot-enabled frame layouts now
  cache summary `any runtime coercion needed?` flags, and the hot inline-call
  setup paths bulk-copy arguments straight into slot frames whenever the
  layout proves no runtime coercion is possible; bound-receiver inline calls
  now do the same for explicit arguments after seeding the injected receiver.
  Focused slot-analysis/quicksort coverage is green, the profiled 50x
  quicksort run reached `9778187 ns/op`, repeated clean 50x reruns landed at
  `9930751`, `9916588`, and `10300807 ns/op`, and
  `tryInlineResolvedCallFromStack(...)` is down to roughly `30ms` cumulative
  instead of sitting in the top hotspot tier;
- the next bytecode tranche is landed too: the direct small-int comparison
  path no longer pays an extra tuple-return helper on every hot comparison.
  `bytecodeDirectIntegerCompare(...)` now decodes and compares concrete
  small-int pairs directly, which removed `bytecodeDirectSmallIntPair(...)`
  from the current hot profile entirely. Focused parity/quicksort coverage is
  green, the profiled 50x quicksort run moved from `10722194` to
  `10384074 ns/op`, repeated clean 50x reruns landed at `10067934`,
  `9961095`, and `9989424 ns/op`, and the compare fast path collapsed from
  the old `80ms` flat / `90ms` cumulative chain to roughly `50ms`
  cumulative total;
- the first post-Milestone-1 allocation-pressure tranche is landed too: hot
  bytecode member/index/binary result sites now reuse existing stack slots in
  place instead of doing repeated pop/pop/append reshaping, and the VM’s hot
  execution paths use unchecked top-slot replacement helpers after their
  existing stack-depth guards. Focused stack/quicksort coverage is green, the
  profiled 50x quicksort run reached `9886874 ns/op`, repeated clean 50x
  reruns landed at `10203423`, `10273281`, and `9753472 ns/op`, and the
  temporary `replaceTop2(...)` hotspot introduced by the first helper rewrite
  dropped back out of the top tier once the hot paths switched to the
  unchecked inlinable helpers;
- the next Milestone 2 tranche is landed too: bytecode programs now cache
  their resolved return generic-name sets, so hot inline/named call paths no
  longer re-enter `FunctionValue.GenericNameSet()` on every frame push, and
  the bytecode small-int boxing cache is now eagerly initialized so hot
  arithmetic no longer pays `sync.Once` on every small boxed integer hit.
  Focused coverage is green; repeated local 50x quicksort reruns landed at
  `9189777`, `9094444`, and `9161513 ns/op`, and the kept profile shows the
  old `FunctionValue.GenericNameSet()` frame gone with small-int boxing
  reduced to noise;
- benchmark harnesses and counters already exist;
- the remaining work is no longer “find obvious first wins”, but a disciplined
  second phase focused on the remaining hot-path costs.

Milestones:

#### Bytecode Milestone 1: Hot Dispatch And Lookup Closure
Closed on 2026-04-16.
- high-frequency environment/path lookup churn is removed from the hot-loop
  dispatch path;
- slot coverage and direct-call lowering are extended far enough that the
  remaining top quicksort costs are no longer ordinary dispatch/lookup
  scaffolding;
- inline caches stay precise under rebinding/mutation and cheap on the current
  single-thread bytecode execution path;
- remaining bytecode optimization work now moves to allocation pressure and
  collection/async-specific hot paths rather than more dispatch/lookup closure.

#### Bytecode Milestone 2: Allocation Pressure Reduction
Closed on 2026-04-16.
- hot stack reshaping, repeated generic-name recomputation, and per-hit
  small-int box-cache initialization are removed from the bytecode VM’s hot
  path;
- transient arg-slice and generic/method-call metadata churn are reduced far
  enough that the remaining benchmark allocation cost is no longer generic VM
  scaffolding;
- the remaining visible allocation/throughput work is now dominated by
  collection-specific paths such as `ArrayStoreWrite`, plus the core integer
  compare/index loop, so that work moves to Milestone 3.

#### Bytecode Milestone 3: Collections / Containers / Async Hot Paths
Closed on 2026-04-16.
- dynamic array append growth now goes through the bytecode runtime's explicit
  reserve policy instead of the Go slice heuristic, which cut the quicksort
  benchmark from roughly `106.6 KB/op / 51 allocs/op` to roughly
  `75.6-75.9 KB/op / 49-50 allocs/op` while keeping the hot loop in the
  low-`9ms/op` band;
- a dedicated bytecode `spawn` / `future_yield` / `future_flush` benchmark is
  now in-tree, and the accidental per-spawn bytecode re-lowering path is
  removed by caching lowered spawn bodies in bytecode instructions;
- serial future scheduling also avoids the old queue-insert copy churn for
  not-yet-started tasks, and the new async benchmark moved from roughly
  `1.61-1.71ms/op`, `~1.07 MB/op`, and `3537-3539 allocs/op` down to roughly
  `1.23-1.43ms/op`, `~279 KB/op`, and `2666-2668 allocs/op`;
- the remaining bytecode costs are now core compare/call execution plus
  fundamental scheduler/context creation work, so further work moves to
  Milestone 4 benchmark gates rather than more hidden collection/async
  scaffolding cleanup.

#### Bytecode Milestone 4: Benchmark Gates
Closed on 2026-04-16.
- [x] keep benchmark baselines current for treewalker vs bytecode vs compiled;
- [x] add report-first perf guardrails, then optional thresholds once noise is
      characterized.
- `v12/bench_suite --suite bytecode-core` now records the checked-in
  cross-mode baseline for:
  - `quicksort`
  - `future_yield_i32_small`
  - `sum_u32_small`
- the current baseline artifacts are:
  - `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json`
  - `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md`
- `v12/bench_guardrail` now compares a fresh suite JSON against a baseline and
  reports status/timing/GC deltas without failing the build, so benchmark
  regressions are visible before any hard thresholds are introduced.

### Backlog (not active until compiler + bytecode priorities permit)

These items remain important, but they are not active priorities right now.

#### Integration / Tooling backlog
- stdlib externalization follow-ups
  - remaining open slice: collision/error semantics when multiple `name: able`
    roots are simultaneously visible through lockfiles, overrides, or
    `ABLE_MODULE_PATHS`
- fixture exporter and other tooling cleanup
- testing CLI / user-facing testing framework work

#### Language / Runtime backlog
- WASM runtime work
- regex syntax and engine work
- broader parser/tree-sitter coverage work not required for compiler completion
- additional stdlib redesign/migration work not required for compiler release
- concurrency feature expansion beyond current spec/runtime requirements

#### Documentation backlog
- continue reconciling older design notes against the active Go-first toolchain
  and the compiler lowering spec as work resumes in those areas
- keep `spec/TODO_v12.md` and relevant design notes current when compiler or
  bytecode work resolves remaining language/implementation gaps
