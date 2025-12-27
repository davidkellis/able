# v11 Exec Conformance Plan

This document tracks end-to-end exec fixtures for Able v11 and maps them to the specification. Every fixture should focus on a narrow semantic slice, emit deterministic output, and include inline `##` comments that describe the exact behaviour being asserted.

## Naming
- Place fixtures under `v11/fixtures/exec`.
- Use `exec/<section>_<feature>[_variation]/` (e.g., `exec/08_control_flow_fizzbuzz`, `exec/11_error_rescue_ensure`). Sections are zero-padded when helpful for sorting and can group related headings (e.g., `06` for expressions/types, `11` for errors).
- Use `_diag` for diagnostic/error-only cases and `_combo` for multi-feature compositions.

## Coverage Matrix (seeded + planned)

| Spec section(s) | Semantic focus | Fixture id(s) | Status / notes |
| --- | --- | --- | --- |
| 15.1–15.6, 6.6 | Program entry, print output, implicit `void` main | `exec/15_01_program_entry_hello_world` | Seeded (renamed from `exec_hello_world`); covers entry point signature + stdout |
| 8.1.1, 8.2.2, 6.3 | `if/elsif/else` + `for` range iteration | `exec/08_01_control_flow_fizzbuzz` | Seeded (renamed from `exec_fizzbuzz`); ensures branch ordering + range inclusivity |
| 8.1.1, 6.11 | `if` truthiness and expression result values | `exec/08_01_if_truthiness_value` | Seeded; truthiness drives branch selection and nil when no else |
| 8.2.1, 8.2.4 | `while` loop with continue/break control flow | `exec/08_02_while_continue_break` | Seeded; continue skips body tail and break exits early |
| 6.1, 6.3, 8.2.2 | Integer literals, arithmetic, accumulation in loops | `exec/08_02_numeric_sum_loop` | Seeded (renamed from `exec_math_sum`); validates `%`/`+` semantics and loop exit |
| 8.2.3, 8.3.5 | `loop` expression break payloads | `exec/08_02_loop_expression_break_value` | Seeded; break returns payload or nil |
| 8.2.5 | Inclusive/exclusive range iteration | `exec/08_02_range_inclusive_exclusive` | Seeded; ascending/descending range bounds |
| 8.3 | Breakpoint non-local jumps | `exec/08_03_breakpoint_nonlocal_jump` | Seeded; labeled breaks unwind to breakpoint with payload |
| 6.3.2, 11.3.3 | Division by zero raises a runtime error | `exec/04_02_primitives_truthiness_numeric_diag` | Seeded; confirms division-by-zero error propagation |
| 6.3.2-6.3.3, 14.1.1, 14.1.4 | Operator dispatch via Add/Index/IndexMut interfaces | `exec/06_03_operator_overloading_interfaces` | Seeded; custom structs participate in `+` and `[]`/`[]=` |
| 6.3.4 | Safe navigation short-circuits on nil | `exec/06_03_safe_navigation_nil_short_circuit` | Seeded; receiver evaluated once and argument evaluation skipped on nil |
| 6.4, 7.4.1, 7.4.2 | Function call argument order + trailing lambda | `exec/06_04_function_call_eval_order_trailing_lambda` | Seeded; arguments evaluate left-to-right and trailing lambda is equivalent to explicit argument |
| 6.6 | String interpolation with escapes + multiline | `exec/06_06_string_interpolation` | Seeded; interpolation evaluates expressions and escapes `\\``/`\\$` with multiline literals preserved |
| 6.7 | Generator yield/stop + IteratorEnd | `exec/06_07_generator_yield_iterator_end` | Seeded; next() yields values then returns IteratorEnd repeatedly after stop |
| 6.1.8, 6.3, 6.4 | Function calls + pipeline, generator iterator protocol (`Iterator { ... }`) | `exec/06_07_iterator_pipeline` | Seeded (renamed from `exec_iterator_pipeline`); exercises `gen.yield`, `Iterator.next`, `|>` |
| 6.8 | Array ops and mutability (size/push/pop/get/set/index/iteration) | `exec/06_08_array_ops_mutability` | Seeded; aliases share storage and bounds errors surface via Index/IndexMut |
| 6.9 | Line joining and trailing commas | `exec/06_09_lexical_trailing_commas_line_join` | Seeded; joins within delimiters with trailing commas in arrays/structs/imports |
| 6.10 | Dynamic package objects + late-bound imports | `exec/06_10_dynamic_metaprogramming_package_object` | Seeded; dyn.def_package/def and dynimport observe redefinitions |
| 6.11, 6.3.2 | Truthiness + logical operators in boolean contexts | `exec/06_11_truthiness_boolean_context` | Seeded; nil/false/error falsy with `!`, `&&`, `||` returning operands |
| 6.12.1 | String helper API (lengths, substring, split, replace, prefix/suffix) | `exec/06_12_01_stdlib_string_helpers` | Seeded; byte/char/grapheme counts, substring bounds, split/replace helpers |
| 6.12.2 | Array helper API (size, push/pop, get/set, clear) | `exec/06_12_02_stdlib_array_helpers` | Seeded; bounds checks, pop nil, and clear resets size |
| 6.12.3 | Numeric helpers (Ratio normalization, divmod) | `exec/06_12_03_stdlib_numeric_ratio_divmod` | Seeded; Ratio normalization and Euclidean /% results |
| 6.6, 4.6, 8.1.2 | Union match expression with wildcard fallthrough | `exec/08_01_union_match_basic` | Seeded (renamed from `exec_match_union`); match ordering and wildcard coverage |
| 8.1.2 | Match guards, default arm behavior, and exhaustiveness errors | `exec/08_01_match_guards_exhaustiveness` | Seeded; guard fallthrough uses default and missing coverage errors |
| 4.6.1, 4.6.3-4.6.4, 8.1.2 | Payload-bearing union variants matched via struct patterns; generic union alias expansion | `exec/04_06_01_union_payload_patterns` | Seeded; union payload patterns + generic alias expansion |
| 4.6.2, 6.11 | Nullable shorthand and truthiness of nil vs non-nil values | `exec/04_06_02_nullable_truthiness` | Seeded; ?T behaves as nil | T in boolean contexts |
| 4.6.3 | Constructing Option/Result unions without propagation | `exec/04_06_03_union_construction_result_option` | Seeded; match handles nil/error variants directly |
| 4.6.4, 8.1.2 | Union matches with guards and ordered clauses | `exec/04_06_04_union_guarded_match_exhaustive` | Seeded; guard/ordering behavior across union variants |
| 4.6.4, 8.1.2 | Non-exhaustive union match raises error | `exec/04_06_04_union_guarded_match_exhaustive_diag` | Seeded; runtime error on missing variant coverage |
| 11.3 | Unhandled `raise` causes non-zero exit with message | `exec/11_03_raise_exit_unhandled` | Seeded (renamed from `exec_raise_exit`); captures default exception propagation |
| 11.3 | `rescue` + `ensure` ordering with side effects | `exec/11_03_rescue_ensure` | Seeded; ensures rescue handler runs before ensure |
| 8.1.2, 8.2.2, 11.3 | Composition: match + loop + rescue | `exec/11_00_errors_match_loop_combo` | Seeded; raise/rescue within loop with match-driven branches |
| 12.5 | Channels: buffered send/receive, close terminates iteration | `exec/12_05_concurrency_channel_ping_pong` | Seeded; validates rendezvous order and nil on close |
| 12.2, 12.3, 12.6 | `proc` vs `spawn` vs `await` scheduling | `exec/12_02_async_proc_spawn_combo` | Seeded: cooperative scheduling with `proc_yield` + `proc_flush`, future status/value |
| 7.4, 9, 10 | Method call syntax + UFCS + impl dispatch | `exec/09_04_methods_ufcs_basics` | Seeded: inherent method sugar, UFCS free function, type-qualified static |
| 9.5 | Method-set generics and where-clause enforcement | `exec/09_05_method_set_generics_where` | Seeded; constraints gate instance and UFCS calls |
| 10.1 | Interface defaults, implicit vs explicit `Self`, and composite aliases | `exec/10_01_interface_defaults_composites` | Seeded; default methods flow through composite interface types |
| 10.2.1–10.2.5 | Impl specificity ordering, named impl disambiguation, HKT targets | `exec/10_02_impl_specificity_named_overrides` | Seeded; overlapping impls pick most-specific and named impls require explicit calls |
| 7.4.3, 9.1-9.3, 13.4 | Instance vs static methods across imports; UFCS uses imported free function | `exec/09_02_methods_instance_vs_static` | Seeded; instance/static method sets with UFCS via selective import |
| 7.1, 7.4, 9, 13.4 | Composition: generic helpers + inherent methods + imports | `exec/09_00_methods_generics_imports_combo` | Seeded; combines generics, method calls, and module imports |
| 10.3 | Dynamic dispatch via interface-typed values | `exec/10_03_interface_type_dynamic_dispatch` | Seeded; interface-typed calls target concrete impls |
| 13.1-13.2, 13.4 | Directory-derived package paths, hyphen normalization, and package-segment imports | `exec/13_01_package_structure_modules` | Seeded; directory prefix + package segment mapping |
| 13.2–13.5 | Packages/import visibility and private types | `exec/13_02_packages_visibility_diag` | Seeded: alias vs selective import scope; private helpers encapsulated |
| 4.7, 4.6, 7.1 | Type aliases/unions with generic functions | `exec/04_07_types_alias_union_generic_combo` | Seeded: alias applied to union + generic function inference fallback |
| 4.3 | Type expression syntax (nested generics + unions) | `exec/04_03_type_expression_syntax` | Seeded: multi-arg generics and nested type expressions |
| 4.5.2 | Named struct literals + functional update vs mutation | `exec/04_05_02_struct_named_update_mutation`, `exec/04_05_02_struct_named_update_mutation_diag` | Seeded: functional update copies base; mismatched update source errors |
| 4.5.3, 5.2.4 | Positional structs: ordered literals, indexed access/mutation, positional destructuring | `exec/04_05_03_struct_positional_named_tuple` | Seeded: positional field access, mutation, and pattern binding |
| 6.1.4–6.1.5 | Char/string escapes + non-interpolated double-quoted literals | `exec/06_01_literals_string_char_escape` | Seeded: escape sequences and literal `${}` text in standard strings |
| 6.1.1–6.1.2 | Contextual integer/float literal typing and rounding | `exec/06_01_literals_numeric_contextual` | Seeded; numeric literals adopt target types when within range |
| 6.1.1–6.1.2 | Out-of-range numeric literal diagnostics | `exec/06_01_literals_numeric_contextual_diag` | Seeded; overflow in typed context raises error |
| 6.5, 8.1 | `if`/`match` expression values | `exec/06_05_control_flow_expr_value` | Seeded: branches evaluate to selected values |
| 7.1.3-7.1.5 | Function definitions: explicit/implicit generics + return inference | `exec/07_01_function_definition_generics_inference` | Seeded; implicit generics inferred from signatures with return type inference |
| 7.3, 11.1 | Explicit return short-circuits nested blocks; void return allowed | `exec/07_03_explicit_return_flow` | Seeded; early return exits the function from inner blocks |
| 11.1 | Return type enforcement and unreachable tail expressions after early return | `exec/11_01_return_statement_type_enforcement` | Seeded; bare return only for void, early return skips tail |
| 7.2–7.3 | Lambdas, closures, explicit return | `exec/07_02_lambdas_closures_capture` | Seeded: closure capture and explicit return in factory |
| 7.4.1-7.4.3 | Method call syntax + trailing lambdas vs UFCS | `exec/07_04_trailing_lambda_method_syntax` | Seeded; trailing lambdas match explicit calls for free functions and methods |
| 7.4.4 | Apply interface call syntax for non-function callables | `exec/07_04_apply_callable_interface` | Seeded; Apply implementors can be invoked directly or via interface types |
| 7.5 | Placeholder-based partial application | `exec/07_05_partial_application` | Seeded; placeholder arity + evaluation order for partial application |
| 7.6 | Shorthand member access and placeholder lambdas | `exec/07_06_shorthand_member_placeholder_lambdas` | Seeded; #member, fn #method, and @/@n lambdas |
| 7.7 | Overload resolution, nullable tail omission, and ambiguity | `exec/07_07_overload_resolution_runtime` | Seeded; runtime overload selection and ambiguity diagnostics |
| 11.2 | `!`/`or` propagation for `Option`/`Result` | `exec/11_02_option_result_propagation` | Seeded: Option unwrap + error handling via `! or {}` |
| 11.2.3 | `or {}` handlers for `Option`/`Result` + `Error` unions | `exec/11_02_option_result_or_handlers` | Seeded; `or` binds errors and treats `Error`-implementers as failure values |

The matrix is also materialized as `v11/fixtures/exec/coverage-index.json` to enable tooling/CI checks.

## Execution + Reporting
- TS harness: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`.
- Go harness: `cd v11/interpreters/go && go test ./pkg/interpreter -run ExecFixtures`.
- Coverage index lives at `v11/fixtures/exec/coverage-index.json`; `v11/scripts/check-exec-coverage.mjs` (invoked by `run_all_tests.sh`) validates seeded IDs stay in sync with the fixture directories.
