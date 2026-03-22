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
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib-deprecated-do-not-use/`, `design/`, `docs/`). Canonical stdlib source is being moved to the external `able-stdlib` repo and cached via `able setup`.

## Active Priorities
- **Compiler completion**: finish the Go AOT compiler first. This is the top
  priority and the main body of this plan is organized around it.
- **Bytecode performance**: once the compiler is in release shape, focus on
  making the Go bytecode interpreter fast.
- **Everything else**: parser/tooling/WASM/stdlib/testing-framework work stays
  in backlog unless it is directly required to unblock compiler completion or
  bytecode performance work.

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

The compiler already has major native-lowering work landed:
- large static slices of arrays, structs, interfaces, callables, joins, and
  control-flow now stay native;
- explicit dynamic-boundary audits exist;
- reduced benchmark fixtures exist for matrix, iterator, heap, and array-heavy
  paths;
- no-bootstrap / no-fallback enforcement exists for large parts of the static
  fixture set.

The compiler is still not done because these conditions are not yet all true:
- every statically representable type still does not map cleanly to one native
  carrier in all contexts;
- all static patterns/control-flow joins still do not stay native in all cases;
- all static dispatch still does not lower directly in all cases;
- compiled runtime helpers are still too interpreter-shaped in some areas;
- the full compiler validation/perf gate is not yet at release quality.

#### Production definition of done

The compiler is production-ready when all of these are true together:
- every statically representable Able type expression lowers to a native Go
  carrier;
- static control flow and pattern binding stay on native carriers;
- static field/method/interface/index/call dispatch lowers to direct compiled Go
  dispatch;
- dynamic/runtime carriers remain only at explicit dynamic or ABI boundaries;
- compiled runtime helpers implement language semantics directly in Go instead
  of modeling normal static execution in terms of interpreter behavior;
- compiler fixture parity is green under no-bootstrap/no-fallback enforcement;
- the compiled benchmark family is materially faster and free of already-known
  avoidable scaffolding on hot paths.

#### Milestone 1: Centralize Compiler Lowering Knowledge

Goal:
- make every carrier, dispatch, control, pattern, and boundary decision come
  from one shared synthesis point instead of emitter-local rules.

Why this is first:
- without this, the compiler keeps accumulating one-off fixes and nominal-type
  drift.

Required work:
- [ ] establish one canonical type-normalization path used by all codegen
      stages;
- [ ] establish one canonical carrier-synthesis path used by all emitters;
- [ ] establish one canonical join/pattern-synthesis path;
- [ ] establish one canonical dispatch-synthesis path;
- [ ] establish one canonical control-envelope synthesis path;
- [ ] establish one canonical boundary-adapter synthesis path;
- [ ] audit emitters and remove local fallback decisions that bypass those
      shared paths.

Proof required:
- generated-source audits that fail on new ad hoc carrier or dispatch fallback
  patterns.

#### Milestone 2: Native Carrier Completeness

Goal:
- every statically representable Able type expression maps to a native Go
  carrier everywhere it appears.

Required work:
- [ ] remove remaining representable `runtime.Value` / `any` fallbacks from
      type mapping and carrier synthesis;
- [ ] finish carrier synthesis for nullable, result, union, interface,
      callable, and fully bound generic nominal types;
- [ ] ensure alias-expanded and recovered type expressions use the same carrier
      path as directly written types;
- [ ] ensure representable branch/join locals do not regress to dynamic
      carriers just because a value was temporarily recovered from a broader
      path;
- [ ] ensure residual `runtime.Value` union members exist only for true dynamic
      payloads.

Current concrete queue under this milestone:
- [ ] close remaining carrier-synthesis gaps in `types.go` and
      `generator_native_unions.go`;
- [ ] remove representable `runtime.Value` / `any` locals from generated static
      code for union/result/nullable/interface/callable joins;
- [ ] add compile-shape regressions for named alias and parameterized union/
      result locals that still degrade.

Proof required:
- representative carrier audits over arrays, structs, unions, interfaces,
  callables, results, and generics;
- generated-source scans for representable `runtime.Value` / `any` locals.

#### Milestone 3: Native Pattern And Control-Flow Completeness

Goal:
- all static pattern matching, branch joins, loop results, rescue/or-propagation,
  and non-local control stay native when the types are statically representable.

Required work:
- [ ] remove runtime typed-pattern fallback for recoverable static subjects;
- [ ] keep `if`, `match`, `rescue`, `or {}`, `loop`, and `breakpoint` joins on
      native carriers in all representable cases;
- [ ] keep typed bindings and recovered branch bindings on native carriers;
- [ ] finish explicit control-envelope propagation for helper boundaries and
      remove any remaining panic/recover/IIFE-style static control scaffolding;
- [ ] ensure `raise`, `rethrow`, `ensure`, `!`, and `or {}` are implemented as
      proper compiled control/data semantics instead of interpreter-shaped
      escape hatches.

Current concrete queue under this milestone:
- [ ] remove recovered-subject typed-pattern fallback in `generator_match.go`,
      `generator_match_runtime_types.go`, and `generator_rescue.go`;
- [ ] eliminate representable join/binding fallback locals in
      `generator_or_else.go`, `generator_rescue.go`, and
      `generator_join_types.go`;
- [ ] add generated-source audits for `__able_try_cast(...)`,
      `bridge.MatchType(...)`, panic/recover, and IIFE-style static control
      lowering.

Proof required:
- fixture slices for `match`, `rescue`, `or {}`, loops, breakpoints, typed
  patterns, and propagation paths;
- generated-source audits that fail when static paths regress to runtime-type
  helpers.

#### Milestone 4: Native Dispatch Completeness

Goal:
- all statically resolved operations compile to direct Go dispatch.

Required work:
- [ ] finish static field access/assignment lowering;
- [ ] finish static method and default-method lowering;
- [ ] finish static interface/default-method dispatch lowering for all
      representable generic cases;
- [ ] finish static callable/bound-method/partial application lowering;
- [ ] finish static index/get/set/apply lowering without dynamic helper
      dispatch;
- [ ] remove residual dynamic helper dispatch from static call/member/index
      paths.

Current concrete queue under this milestone:
- [ ] audit representative static bodies for `__able_call_value(...)`,
      `__able_method_call_node(...)`, `__able_member_get*`, and `__able_index*`
      outside explicit dynamic boundaries;
- [ ] close remaining generic interface/default-method dispatch cases that
      still re-enter runtime dispatch on statically representable paths;
- [ ] close remaining static callable/index assignment gaps on recovered native
      carriers.

Proof required:
- combined-source dispatch audits;
- fixture slices covering structs, interfaces, callables, indexable types, and
  generic/default-method paths.

#### Milestone 5: Compiled Runtime Core Independence

Goal:
- compiled runtime helpers used on static paths must implement Able semantics as
  normal Go logic, not as thin wrappers around interpreter-style machinery.

Required work:
- [ ] audit every compiled runtime helper family used on static paths;
- [ ] replace helpers whose normal static behavior is still modeled too closely
      after interpreter operations;
- [ ] keep only true dynamic-boundary helpers dependent on runtime/interpreter
      object-model values;
- [ ] ensure array/map/range/string/concurrency runtime services used by
      compiled code are direct Go implementations;
- [ ] keep helper control propagation aligned with the explicit control-envelope
      model.

Proof required:
- source audit over emitted helper families;
- static fixture slices that exercise helper families without dynamic features.

#### Milestone 6: Boundary Containment And Static Cleanliness

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
- [ ] enumerate and document the final allowed boundary helper set;
- [ ] tighten adapters so conversion happens exactly at the edge and returns to
      native carriers immediately after;
- [ ] remove residual dynamic leakage from static fixtures;
- [ ] keep strict no-bootstrap/no-fallback/no-boundary audits green for static
      fixture families.

Proof required:
- boundary-marker harnesses;
- strict static fixture audits;
- generated-source audits around boundary helper usage.

#### Milestone 7: Compiler Performance Completion

Goal:
- make the compiler-generated code fast on the checked-in benchmark family
  without violating the lowering rules.

Required work:
- [ ] keep using reduced checked-in benchmarks to isolate hot shared lowering
      gaps;
- [ ] remove only shared primitive/control/array/callable/dispatch scaffolding,
      never by adding named non-primitive fast paths;
- [ ] remeasure after each material compiler workstream;
- [ ] keep benchmark proofs paired with generated-source shape audits.

Primary benchmark families:
- matrix / array hot paths
- iterator / generic container pipelines
- nominal generic method/container paths
- recursion/call overhead microbenchmarks

Definition of done for this milestone:
- [ ] hot compiled benchmark paths no longer carry already-identified avoidable
      scaffolding;
- [ ] checked-in benchmark baselines and current numbers are up to date.

#### Milestone 8: Compiler Release Validation

Goal:
- turn the above architecture into a hard release gate.

Required work:
- [ ] keep one authoritative lowering spec and one authoritative completion plan
      in sync with implementation;
- [ ] keep compiler fixture parity green under no-bootstrap/no-fallback rules;
- [ ] run the full compiler matrix and stdlib suite in compiled mode as a
      release gate;
- [ ] ensure diagnostics and failure behavior are stable enough for production
      use;
- [ ] confirm reproducible build trees and clean-checkout behavior for the
      compiler toolchain.

Release gate checklist:
- [ ] `./run_all_tests.sh` green
- [ ] `./run_stdlib_tests.sh` green
- [ ] full compiled fixture matrix green
- [ ] strict static no-fallback/no-boundary audits green
- [ ] benchmark baselines updated
- [ ] no known representable static-path regressions remaining in PLAN

#### Immediate compiler queue (start here)

1. [ ] Centralize remaining carrier-synthesis decisions so representable static
       types stop falling back to `runtime.Value` / `any`.
2. [ ] Remove remaining static typed-pattern fallback to runtime type-match
       helpers.
3. [ ] Tighten remaining static dispatch paths that still route through dynamic
       helper dispatch.
4. [ ] Audit compiled runtime helper families that still model static semantics
       too closely after interpreter-oriented helpers.
5. [ ] Re-run boundary and benchmark gates after each completed milestone.

### Bytecode Performance Program (second priority; start after compiler release work is closed or paused explicitly)

Goal:
- make the Go bytecode interpreter fast enough to be a practical execution mode
  after the compiler is finished.

Current state snapshot:
- a large amount of call-dispatch, lookup-cache, frame-pool, integer-op, and
  hotspot work is already landed;
- benchmark harnesses and counters already exist;
- the remaining work is no longer “find obvious first wins”, but a disciplined
  second phase focused on the remaining hot-path costs.

Milestones:

#### Bytecode Milestone 1: Hot Dispatch And Lookup Closure
- [ ] remove remaining high-frequency environment/path lookups in hot loops;
- [ ] extend slot coverage and direct-call lowering where it still materially
      affects benchmarked workloads;
- [ ] keep inline caches precise and cheap under mutation/concurrency.

#### Bytecode Milestone 2: Allocation Pressure Reduction
- [ ] cut remaining boxed integer churn, collection allocation churn, iterator
      churn, and closure scaffolding in hot paths;
- [ ] reduce transient arg-slice and method-resolution allocations that still
      survive the first optimization wave.

#### Bytecode Milestone 3: Collections / Containers / Async Hot Paths
- [ ] optimize remaining array/hash-map/iterator hot paths that still dominate
      benchmark traces;
- [ ] audit async/future execution overhead under realistic workloads.

#### Bytecode Milestone 4: Benchmark Gates
- [ ] keep benchmark baselines current for treewalker vs bytecode vs compiled;
- [ ] add report-first perf guardrails, then optional thresholds once noise is
      characterized.

### Backlog (not active until compiler + bytecode priorities permit)

These items remain important, but they are not active priorities right now.

#### Integration / Tooling backlog
- staged integration cleanup and clean-checkout reproducibility follow-ups
- stdlib externalization follow-ups
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
