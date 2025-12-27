# Able Project Roadmap (v11 focus)

## Standard onboarding prompt
Read AGENTS, PLAN, and the v11 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

## Standard next steps prompt
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10 toolchain available for historical reference while driving all new language, spec, and runtime work in v11.
- Keep the Go interpreter as the behavioural reference and ensure the TypeScript runtime + future ports match feature-for-feature (the concurrency implementation strategy may differ).
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical v10 branch and the actively developed v11 runtime; document any deltas immediately in the v11 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics for the archived toolchain. Keep it untouched unless a maintainer requests an errata fix.
- `spec/full_spec_v11.md`: active specification for all current work; every behavioural change must be described here.
- `v10/interpreters/{ts,go}/`: Frozen interpreters that match the v10 spec. Treat them as read-only unless a blocking support request lands.
- `v11/interpreters/{ts,go}/`, `v11/parser`, `v11/fixtures`, `v11/stdlib`: active development surface for Able v11.
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v11.md`; log discrepancies in `spec/TODO_v11.md`. Reference the v10 spec only when clarifying historical behaviour.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types; track string/regex bring-up via `design/regex-plan.md` and the new spec TODOs covering byte-based strings with char/grapheme iterators.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones now live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.
- Status: interface alignment is in place (Apply/Index/Awaitable + stdlib helpers); parity harness is green with the latest sweep saved to `tmp/parity-report.json`; kernel/stdlib loader auto-detection is bundled; Ratio/numeric conversions are landed; exec fixtures are spec-prefixed with coverage enforced via `scripts/check-exec-coverage.mjs` in `run_all_tests.sh`; next focus is regex stdlib expansions and tutorial cleanup.

## Guardrails (must stay true)
- `v11/interpreters/ts/scripts/run-parity.ts` remains the authoritative entry point for fixtures/examples parity; `./run_all_tests.sh --version=v11` must stay green (TS + Go unit tests, fixture suites, parity CLI). Run the v10 suite only when explicitly asked to investigate archival regressions.
- `v11/interpreters/ts/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (working queue: tackle in order, move completed items to LOG.md)

### v11 Conformance Exec Fixture Suite (new workstream)
- Reference: `v11/docs/conformance-plan.md` for the coverage matrix + naming scheme.
- [x] Draft a coverage matrix mapping every v11 spec heading/feature (and key variations) to one or more exec fixtures; land as `v11/docs/conformance-plan.md`. Include naming scheme (`exec/<section>_<feature>[_variation]/`) and mark which headings share fixtures.
- [x] Ensure `scripts/export-fixtures`/`run_all_tests` include the new exec fixtures and add a simple coverage index (JSON/YAML) generated from the matrix to prevent gaps; consider a CI check to flag missing headings.

### v11 Exec Fixture Backlog (heading-specific coverage)
Note: It is expected that some new fixtures will fail due to interpreter bugs/deficiencies.We should implement fixtures strictly in accordance with the v11 spec semantics. Do not weaken or sidestep the behavior under test to “make tests pass”. If a fixture fails under a given interpreter, follow up by fixing the interpreter so the implementation honors the spec. The interpreters should perfectly implement all the semantics described in the v11 spec.
- [x] `exec/02_lexical_comments_identifiers` — §2/§6.9: identifiers vs reserved placeholders, comments ignored, trailing commas/line join tolerated.
- [x] `exec/03_blocks_expr_separation` — §3: newline vs semicolon separation and last-expression block value with scoped bindings.
- [x] `exec/04_01_type_inference_constraints` — §4.1.4–4.1.6: polymorphic function calls with inferred type params and constraint satisfaction.
- [x] `exec/04_07_02_alias_generic_substitution` — §4.7.2: generic alias expansion, substitution, and inference through alias chains.
- [x] `exec/04_07_03_alias_scope_visibility_imports` — §4.7.3: alias visibility across packages with `import`/`private` and alias re-export.
- [x] `exec/04_07_04_alias_methods_impls_interaction` — §4.7.4: aliases preserving method/impl dispatch and type-based feature lookups.
- [x] `exec/04_07_05_alias_recursion_termination` — §4.7.5: rejecting/handling recursive aliases per termination rules (diagnostic path).
- [x] `exec/05_00_mutability_declaration_vs_assignment` — §5.0/§5.1: `:=` introduces new bindings, `=` requires existing, mutation rules.
- [x] `exec/05_02_identifier_wildcard_typed_patterns` — §5.2.1–5.2.2/§5.2.7: identifier vs `_` binding and typed patterns in declarations.
- [x] `exec/05_02_struct_pattern_rename_typed` — §5.2.3: struct pattern with `::` renames, optional fields, and typed destructuring.
- [x] `exec/05_02_array_nested_patterns` — §5.2.5–§5.2.6: array/nested destructuring with mixed patterns and rest handling (if applicable).
- [x] `exec/05_03_assignment_evaluation_order` — §5.3/§5.3.1: RHS evaluated once, compound patterns, mutable reassignment side effects.
- [x] `exec/06_01_literals_array_map_inference` — §6.1.7–§6.1.9: array/map literal inference, spread/entry order, mixed element unification.
- [x] `exec/06_02_block_expression_value_scope` — §6.2: block-as-expression value, inner scope bindings, void-return blocks.
- [x] `exec/06_03_operator_precedence_associativity` — §6.3.1: precedence/associativity across pipes, assignment, arithmetic/boolean ops.
- [ ] `exec/11_01_return_statement_type_enforcement` — §11.1: return type enforcement and unreachable tail expressions after early return.
- [ ] `exec/11_02_option_result_or_handlers` — §11.2.3: `or {}` handlers with/without error binding, option vs result divergence.
- [ ] `exec/11_03_rescue_rethrow_standard_errors` — §11.3: Arithmetic/Indexing errors, `rescue` with `ensure`, rethrow semantics.
- [ ] `exec/12_02_proc_fairness_cancellation` — §12.2.5: cooperative scheduling fairness, `proc_yield`, cancellation and `proc_flush`.
- [ ] `exec/12_03_spawn_future_status_error` — §12.3: future status/value/error propagation, repeated await, cancellation request.
- [ ] `exec/12_04_proc_vs_spawn_differences` — §12.4: side-by-side proc vs spawn behaviours (blocking vs immediate future, cancellation semantics).
- [ ] `exec/12_05_mutex_lock_unlock` — §12.5: mutex lock/unlock, helper ensuring unlock on scope exit, double-lock error.
- [ ] `exec/12_06_await_fairness_cancellation` — §12.6: await on multiple awaitables, fairness, cancellation propagation through awaitable protocol.
- [ ] `exec/12_07_channel_mutex_error_types` — §12.7: channel/mutex error variants surfaced to user code and handled via rescue.
- [ ] `exec/13_03_package_config_prelude` — §13.3: `package.yml` fields (name, prelude) influencing resolution and visibility.
- [ ] `exec/13_04_import_alias_selective_dynimport` — §13.4: `import` with `::` aliases/selectors plus `dynimport` late binding.
- [ ] `exec/13_06_stdlib_package_resolution` — §13.6: resolving `able.*` stdlib packages without extra module paths.
- [ ] `exec/13_07_search_path_env_override` — §13.7: module search path/env override (`ABLE_MODULE_PATHS`) resolution precedence.
- [ ] `exec/14_01_language_interfaces_index_apply_iterable` — §14.1.1–§14.1.3: Index/IndexMut/Apply/Iterator interfaces driving dispatch.
- [ ] `exec/14_01_operator_interfaces_arithmetic_comparison` — §14.1.4–§14.1.7: operator interfaces (arithmetic/bitwise/compare/display/error/clone/default) backing runtime calls.
- [ ] `exec/14_02_regex_core_match_streaming` — §14.2: regex compile/match/set/streaming semantics (pending stdlib implementation).
- [ ] `exec/15_02_entry_args_signature` — §15.2: main signature/args handling, default void return enforced.
- [ ] `exec/15_03_exit_status_return_value` — §15.3: exit status based on return vs runtime error, stdout/stderr ordering.
- [ ] `exec/15_04_background_work_flush` — §15.4: background work (proc/spawn) completion expectations at program exit.
- [ ] `exec/16_01_host_interop_inline_extern` — §16: extern host function bodies, prelude usage, error mapping for host failures.
