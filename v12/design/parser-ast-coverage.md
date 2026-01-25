# Parser & AST Coverage Checklist (Able v12)

## Purpose
- Track every surface feature defined in [`spec/full_spec_v12.md`](../spec/full_spec_v12.md) and confirm that:
  - the parser produces the correct concrete syntax tree and canonical AST, and
  - the shared AST fixture suite exercises the feature (both interpreters consume these fixtures).
- Serve as the canonical backlog for parser/AST gaps until every item is verified by dedicated tests/fixtures.

## Status Legend
- `TODO` – No targeted coverage yet (parser + fixtures missing).
- `Partial` – Some coverage exists but is incomplete (e.g., only one interpreter, limited scenarios).
- `Done` – Parser tests and AST fixtures both cover the feature with representative scenarios.

> **Note:** As of the latest audit, every feature now has AST fixtures (`TODO` appears only in the “Parser Tests” column). Remaining `TODO` entries indicate missing parser assertions, not fixture gaps.

> **Note:** “Parser Tests” refers to focused assertions in `interpreter-go/pkg/parser/*`. “AST Fixtures” refers to entries under `fixtures/ast` exported via the Go fixture exporter (TODO).

> **Note:** Fixtures that include a `source.able` file should be round-tripped by the Go parser harness once the Go exporter + fixture driver are wired for v12.

---

## Expressions & Literals
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Integer literals (suffix variants) | §6.1.1 | `expressions/int_addition`, `expressions/integer_suffix` | `TestParseExpressionFixtures` | Done | Parser confirms underscore-suffixed integer literals (e.g., `42_i64`). |
| Float literals (`f32`, `f64`) | §6.1.2 | `expressions/mixed_numeric_arithmetic`, `expressions/float_suffix` | `TestParseExpressionFixtures` | Done | Parser handles suffixes (`3.5_f32`) and emits typed float literals. |
| Boolean literals | §6.1.3 | `basics/bool_literal`, `expressions/bool_literal_true` | `TestParseLiteralFixtures` | Done | Literal suite now exercises boolean parsing. |
| Character literals | §6.1.4 | `basics/char_literal`, `basics/char_escape` | `TestParseLiteralFixtures` | Done | Literal suite covers simple character parsing; escape variants still available via fixtures. |
| String literals | §6.1.5 | `basics/string_literal` | `TestParseLiteralFixtures` | Done | Dedicated literal fixture verifies parser output. |
| Nil literal | §6.1.6 | `basics/nil_literal` | `TestParseLiteralFixtures` | Done | Parser asserts `nil` literal shape. |
| Array literal | §6.8 | `patterns/array_destructuring`, `expressions/array_literal_empty`, `expressions/array_literal_typed` | `TestParseExpressionFixtures` | Done | Empty-array case covered; typed scenario still represented via fixtures. |
| Struct literal (named fields) | §4.5.2 | `structs/named_literal` | `TestParseStructFixtures` | Done | Parser covers record-style structs with field initialisers. |
| Struct literal (positional fields) | §4.5.3 | `structs/positional_literal` | `TestParseStructFixtures` | Done | Parser exercises tuple-style literals and numeric member access. |
| Struct literal (functional updates / spreads) | §4.5.2 | `structs/functional_update` | `TestParseStructAndArrayLiterals` | Done | Fixture exercises `Point { ...p, x: 5 }`; parser suite now asserts spread ordering alongside array literals. |
| Unary expressions | §6.3.2 | `expressions/unary_negation`, `expressions/unary_not`, `expressions/unary_bitwise`, `expressions/unary_double_negation` | `TestParseExpressionFixtures` | Done | Unary negation fixture parsed; extend as needed for other operators. |
| Binary operators (arithmetic) | §6.3.2 | `expressions/int_addition` | `TestParseArithmeticPrecedence`, `TestParseModuleImports` | Done | Precedence test now covers additive/multiplicative ordering. |
| Pipe operator (`|>`) — basic chain | §6.3.2 | `pipes/member_topic` | `TestParsePipeChainExpression` | Done | Chain parsing covered, but topic/placeholders tracked separately. |
| Pipe operator (`|>`) — topic / placeholder combos | §6.3.2, §7.6.1 | `pipes/topic_placeholder` | `TestParsePipeTopicAndPlaceholderSteps` | Done | Covers `%` arithmetic, placeholder-generated callables, and topic method calls; extend when new pipe forms land. |
| Block expression (`do {}`) | §6.2 | `expressions/block_expression` | `TestParseExpressionFixtures` | Done | Block fixture parsed via expression suite. |
| If / elsif / else | §6.5 | `control/if_else_branch`, `control/if_or_else` | `TestParseIfExpression` | Done | Parser test exercises elsif chains with a trailing default block. |
| Match expression (identifier + literal) | §6.5 / §4.6 | `match/identifier_literal`, `match/guard_clause`, `match/wildcard_pattern` | `TestParseMatchFixtures`, `TestParseMatchExpression` | Done | Parser fixtures now cover literal, identifier, guard, and wildcard fallback match clauses. |
| Match expression (struct guard) | §6.5 | `match/struct_guard`, `match/struct_positional_pattern` | `TestParseMatchFixtures`, `TestParseMatchExpression` | Done | Fixture suite now covers guarded named matches and positional struct destructuring. |
| Lambda expression (inline) | §6.4 | `functions/lambda_expression` | `TestParseLambdaExpressionLiteral` | Done | Parser now asserts standalone lambdas assigned to locals. |
| Trailing lambda call | §6.4 | `functions/trailing_lambda_call` | `TestParseModuleImports`, `TestParseTrailingLambdaCallSimple` | Done | Coverage checks `isTrailingLambda` for method calls with trailing blocks. |
| Function call with type arguments | §6.4 | `functions/generic_application` | `TestParseFunctionCallWithTypeArguments`, `TestParseModuleImports` | Done | Dedicated test asserts multi-argument type applications. |
| Range expression | §6.8 | `control/for_range_break`, `control/range_inclusive` | `TestParseControlFlowFixtures`, `TestParseRangeExpressions` | Done | Range test asserts exclusive/inclusive literals outside fixture coverage. |
| Member access | §4.5 / §6.4 | `strings/interpolation_struct_to_string`, `structs/positional_literal` | `TestParseMemberAccessChaining`, `TestParseStructFixtures` | Done | Test covers chained property access, method calls, and assignments. |
| Index expression | §6.8 | `expressions/index_access` | `TestParseExpressionFixtures`, `TestParseIndexExpressions` | Done | Parser test now exercises identifier and call-derived index expressions. |
| String interpolation | §6.6 | `strings/interpolation_basic`, `strings/string_literal_escape` | `TestParseStringFixtures` | Done | Parser now asserts multi-part interpolation with embedded expressions. |
| Propagation expression (`!`) | §11.2.2 | `errors/or_else_handler`, `expressions/or_else_success` | `TestParsePropagationExpression`, `TestParsePropagationAndOrElse` | Done | Parser suite covers propagation with and without handlers. |
| Or-else expression (`else {}`) | §11.2.3 | `errors/or_else_handler`, `expressions/or_else_success` | `TestParsePropagationAndOrElse`, `TestParseErrorHandlingExpressions` | Done | Binding + guard cases now asserted via parser tests. |
| Rescue / ensure expressions | §11.3.2 / §11.3.2 | `errors/rescue_catch`, `errors/ensure_runs`, `expressions/rescue_success`, `expressions/ensure_success` | `TestParseRescueAndEnsure`, `TestParseErrorHandlingExpressions` | Done | Parser tests cover monitored blocks, rescue guards, and ensure clauses. |
| Breakpoint expression | §6.5 | `expressions/breakpoint_value` | `TestParseBreakpointExpression` | Done | Parser test exercises labeled breakpoint blocks. |
| Spawn expression (Future handle) | §12.2–§12.3 | `concurrency/future_cancel_value`, `concurrency/future_yield_flush`, `concurrency/future_value_reentrancy`, `concurrency/future_value_memoization`, `concurrency/future_value_cancel_memoization` | `TestParseSpawnExpressionForms`, `TestParseConcurrencyFixtures` | Done | Parser tests validate spawn expressions with block/call targets plus nested `future.value()` re-entrancy and cancellation/memoization scenarios. |
| Generator literal (`Iterator {}`) | §6.7 | `control/iterator_for_loop`, `control/iterator_while_loop`, `control/iterator_if_match` | `TestParseIteratorLiteral` | Done | Parser test now covers typed and untyped iterator literals; fixtures exercise yield/stop paths. |

## Patterns & Assignment
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Declaration assignment (`:=`) | §5.1 | `expressions/assignment_declare` | `TestParseWhileLoopWithBreakAndContinue`, `TestParseForLoopWithAssignment` | Done | Parser tests now cover declaration assignment in control-flow contexts. |
| Reassignment (`=`) | §5.1 | `expressions/reassignment` | `TestParseForLoopWithAssignment` | Done | Parser test exercises reassignment inside `for` loop body. |
| Compound assignments (`+=`, etc.) | §5.1 | `expressions/compound_assignment`, `expressions/compound_assignment_variants` | `TestParseWhileLoopWithBreakAndContinue` | Done | While-loop parser test asserts additive compound assignment form. |
| Identifier pattern | §5.2.1 | `patterns/named_struct_destructuring` | `TestParsePatternFixtures` | Done | Parser asserts plain identifier bindings via struct destructuring fixture. |
| Wildcard pattern (`_`) | §5.2.2 | `patterns/wildcard_assignment`, `match/wildcard_pattern` | `TestParseMatchFixtures` | Done | Parser fixture suite asserts wildcard fallback behaviour in match expressions. |
| Struct pattern (named fields) | §5.2.3 | `match/struct_guard`, `match/nested_struct_pattern`, `patterns/nested_struct_destructuring` | `TestParseMatchExpression`, `TestParsePatternFixtures` | Done | Named struct patterns covered via guarded matches and nested destructuring assignment. |
| Struct pattern (positional) | §5.2.4 | `match/struct_positional_pattern`, `patterns/struct_positional_destructuring` | `TestParseMatchFixtures`, `TestParsePatternFixtures` | Done | Parser asserts positional struct matches and destructuring assignments. |
| Array pattern | §5.2.5 | `patterns/array_destructuring`, `match/array_pattern`, `patterns/nested_struct_destructuring` | `TestParsePatternFixtures` | Done | Parser asserts array patterns with rest bindings inside nested struct destructuring. |
| Nested patterns | §5.2.6 | `patterns/named_struct_destructuring`, `patterns/nested_struct_destructuring`, `patterns/nested_array_destructuring`, `match/nested_struct_pattern` | `TestParsePatternFixtures` | Done | Nested struct+array destructuring now exercised via parser fixtures. |
| Typed patterns | §5.2.7 | `patterns/typed_assignment` | `TestParsePatternFixtures` | Done | Standalone typed identifier assignments now parsed and validated. |

## Declarations
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Function definition (with params, return type) | §6.4 / §10 | `functions/generic_application`, `strings/interpolation_struct_to_string` | `TestParseFunctionDefinitionWithReturnType`, `TestParseModuleImports` | Done | Parser test covers parameter types and explicit return type annotations. |
| Function generics (`fn<T>`) | §10.2 | `functions/generic_application`, `functions/generic_multi_parameter` | `TestParseGenericFunctionDefinition` | Done | Unit test validates generic parameters and where-clause constraints. |
| Private functions (`private fn`) | §13.5 | `privacy/private_static_method` | `TestParsePrivateFunctionDefinition` | Done | Parser test confirms private flag propagation. |
| Struct definition (named fields) | §4.5.2 | `structs/named_literal` | `TestParseStructDefinitions` | Done | Parser test covers named-field struct declarations. |
| Struct definition (positional fields) | §4.5.3 | `structs/positional_literal` | `TestParseStructDefinitions` | Done | Same test asserts positional tuple struct declarations. |
| Union definition | §4.6 | `unions/simple_match`, `unions/generic_result` | `TestParseUnionAndInterface` | Done | Parser test covers union declarations alongside interface definitions. |
| Interface definition | §10.1 | `errors/generic_constraint_unsatisfied`, `interfaces/composite_generic`, `declarations/interface_impl_success` | `TestParseInterfaceCompositeGenerics` | Done | Parser covers generic headers and composite base lists. |
| Implementation definition (`impl`) | §10.3 | `declarations/interface_impl_success` | `TestParseImplicitMethods` | Done | Coverage includes implicit method bodies within `impl` blocks. |
| Methods definition (`methods for`) | §10.3 | `strings/interpolation_struct_to_string`, `privacy/private_static_method` | `TestParseImplicitMethods` | Done | Methods shorthand, parameters, and implicit member access verified by parser tests. |
| Extern function body | §16.1.2 | `interop/prelude_extern` | `TestParsePreludeAndExtern` | Done | Test asserts extern bodies and target metadata map correctly. |
| Prelude statement | §16.1.1 | `interop/prelude_extern` | `TestParsePreludeAndExtern` | Done | Test confirms host-target preludes parse into the canonical AST. |

## Control Flow Statements
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| While loop | §6.5 | `control/while_sum` | `TestParseWhileLoopWithBreakAndContinue` | Done | Parser test now asserts loop condition, compound assignment, and nested control statements. |
| For loop (`for x in`) | §6.5 / Iteration | `control/for_sum` | `TestParseForLoopWithAssignment` | Done | Parser test covers `for-in` iteration with identifier pattern and reassignment in body. |
| Break statement (with/without label) | §6.5 | `control/for_range_break` | `TestParseWhileLoopWithBreakAndContinue` | Done | Break with value handled inside while-loop parser test. |
| Continue statement | §6.5 | `control/for_continue` | `TestParseWhileLoopWithBreakAndContinue` | Done | Continue statement parsing verified in while-loop test. |
| Return statement | §11.1 | `strings/interpolation_struct_to_string` (method) | `TestParseReturnStatements` | Done | Parser test covers bare `return` and returning a value after control flow. |
| Raise statement | §11.3.1 | `errors/raise_manifest` | `TestParseErrorHandlingFixtures` | Done | Fixture parsed via new error-handling suite. |
| Rescue block | §11.3.2 | `errors/rescue_catch`, `errors/rescue_guard`, `expressions/rescue_success` | `TestParseErrorHandlingFixtures` | Done | Parser tests now cover success + match-clause rescue forms. |
| Ensure block | §11.3.2 | `errors/ensure_runs`, `expressions/ensure_success` | `TestParseErrorHandlingFixtures` | Done | Ensure expressions parsed for both rescue + success paths. |
| Rethrow statement | §11.3.2 | `errors/rethrow_propagates` | `TestParseRaiseAndRethrowStatements` | Done | Unit test now verifies rethrow statements emit the canonical AST node. |

## Types & Generics
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Simple type expression | §4.1.2 | `functions/generic_application`, `errors/generic_constraint_unsatisfied` | `TestParseSimpleTypeExpressions`, `TestParseModuleImports` | Done | Simple boolean/string annotations verified by dedicated parser test. |
| Generic type expression (`Array<T>`) | §4.1.2 | `types/generic_type_expression` | `TestParseTypeExpressionFixtures` | Done | Fixture-backed parser test covers struct fields annotated with `Array i32`. |
| Function type expression (`(T) -> U`) | §4.1.2 | `types/function_type_expression` | `TestParseTypeExpressionFixtures`, `TestParseFunctionTypeMultiParam` | Done | Parser asserts arrow types in parameter positions via fixture + dedicated multi-param test. |
| Nullable type (`T?`) | §4.6.2 | `types/nullable_type_expression` | `TestParseTypeExpressionFixtures` | Done | Fixture ensures the `?string` shorthand on params and returns round-trips. |
| Result type (`!T`) | §11.2.1 | `types/result_type_expression` | `TestParseTypeExpressionFixtures` | Done | Parser test confirms `!i32` return annotations map to the canonical AST. |
| Union type syntax | §4.6 | `types/union_type_expression` | `TestParseTypeExpressionFixtures` | Done | Fixture exercises `string | i32` unions across parameter/return annotations. |
| Type parameter constraints (`where`) | §4.1.5 | `types/generic_where_constraint` | `TestParseTypeExpressionFixtures` | Done | Fixture covers multi-generic `where` clauses with stacked interface constraints. |
| Interface constraint (`T: Display`) | §4.1.5 | `errors/generic_constraint_unsatisfied`, `types/generic_where_constraint` | `TestParseTypeExpressionFixtures` | Done | Fixture round-trip covers interface constraints inside explicit `where` clauses. |

## Modules & Imports
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Package statement | §13.2 | `imports/package_statement` | `TestParsePackageStatement`, `TestParseModuleImports` | Done | Unit tests assert package qualifiers parse into the canonical AST. |
| Import selectors (`import pkg.{Foo, Bar::B}`) | §13.4 | `imports/static_alias_public` | `TestParseImportSelectors`, `TestParseModuleImports` | Done | Targeted test asserts selector lists and per-item aliasing. |
| Wildcard import (`*`) | §13.4 | `imports/static_wildcard` | `TestParseWildcardImport`, `TestParseModuleImports` | Done | Dedicated parser test now verifies wildcard imports without extra statements. |
| Alias import (`import pkg::alias`) | §13.4 | `imports/static_alias_public` | `TestParseImportAlias` | Done | Parser coverage confirms module-level aliasing matches fixture expectations. |
| Dynamic import (`dynimport`) | §13.4 | `imports/dynimport_wildcard`, `imports/dynimport_selector_alias` | `TestParseDynImportSelectors`, `TestParseModuleImports` | Done | Parser tests cover aliasing, selector lists, and wildcard dynimports. |
| Prelude statement (`prelude {}`) | §16.1.1 | `interop/prelude_extern` | `TestParsePreludeAndExtern` | Done | Parser test verifies prelude bodies for host targets. |

## Concurrency & Async
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| `spawn` block/expression | §12.2 | `concurrency/future_cancel_value`, `concurrency/future_yield_flush`, `concurrency/future_value_reentrancy`, `concurrency/future_value_memoization`, `concurrency/future_value_cancel_memoization` | `TestParseSpawnExpressionForms`, `TestParseConcurrencyFixtures` | Done | Dedicated parser test covers block/do/call-target forms plus nested/memoized `future.value()` usage, ensuring both successful and cancelled memoization fixtures stay mapped correctly. |
| `future_*` helpers (`future_yield`, etc.) | §12.2.4 | `concurrency/future_cancelled_helper`, `concurrency/future_yield_flush`, `concurrency/future_flush_fairness` | `TestParseFutureHelpers`, `TestParseConcurrencyFixtures` | Done | Fixtures cover cooperative yielding plus the `future_flush` fairness drain; parser tests assert the helper call shapes. |
| `spawn` expression | §12.3 | `concurrency/future_memoization`, `concurrency/future_value_reentrancy` | `TestParseSpawnExpressionForms`, `TestParseConcurrencyFixtures` | Done | Parser tests validate block/do/call-target spawn expressions and re-entrant `future.value()` usage. |
| Channel literal/ops | §12.5 | `concurrency/channel_basic_ops`, `concurrency/channel_receive_loop`, `concurrency/channel_send_on_closed_error`, `concurrency/channel_nil_send_cancel`, `concurrency/channel_nil_receive_cancel` | `TestParseChannelAndMutexHelpers` | Done | Channel helper invocations now exercised directly by the parser suite. |
| Mutex helper (`mutex`) | §12.5 | `concurrency/mutex_locking`, `concurrency/mutex_contention` | `TestParseChannelAndMutexHelpers` | Done | Parser test covers mutex creation and basic lock/unlock calls. |

## Error Handling & Options
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Option/Result propagation (`!`) | §11.2.2 | `expressions/or_else_success`, `errors/or_else_handler`, `types/result_type_expression`, `unions/generic_result` | `TestParsePropagationExpression`, `TestParsePropagationAndOrElse` | Done | Parser suite now asserts propagation suffix on plain assignments and `else` handlers. |
| Option/Result handlers (`else {}`) | §11.2.3 | `expressions/or_else_success`, `errors/or_else_handler`, `types/result_type_expression`, `unions/generic_result` | `TestParsePropagationAndOrElse`, `TestParseErrorHandlingExpressions` | Done | Coverage includes bound handlers and multi-clause rescue/ensure combinations. |
| Exception raising/rescue/ensure | §11.3 | `errors/raise_manifest`, `errors/rescue_catch`, `errors/rescue_guard`, `errors/rescue_typed_pattern`, `errors/rethrow_propagates`, `errors/ensure_runs`, `expressions/ensure_success` | `TestParseRaiseAndRethrowStatements`, `TestParseErrorHandlingExpressions` | Done | Parser tests cover raise/rethrow statements alongside rescue clauses and ensure blocks. |
| Breakpoint labels | §6.5 | `expressions/breakpoint_value`, `control/for_range_break` | `TestParseBreakpointWithLabel` | Done | Labeled breakpoint expressions now asserted by dedicated parser coverage. |

## Host Interop
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Host preludes (`prelude { ... }`) | §16.1.1 | `interop/prelude_extern` | `TestParsePreludeAndExtern` | Done | Parser test confirms host-target prelude bodies are captured in the AST. |
| Extern function bodies | §16.1.2 | `interop/prelude_extern` | `TestParsePreludeAndExtern` | Done | Parser test round-trips extern functions for the Go target. |

---

## Next Steps (Execution Order)
1. **AST Fixtures First** – Expand `fixtures/ast` (via `v12/export_fixtures.sh`) until every feature in the table has a representative module that can be consumed by the Go runtimes.
2. **Interpreter Verification Second** – Extend the Go tree-walker + bytecode test suites so each fixture is evaluated and its behavior/assertions are checked, guaranteeing runtime support for every AST form.
3. **Parser Coverage Third** – Once fixtures/interpreter checks exist, add focused parser unit tests that parse the surface syntax into the canonical AST and confirm round-trips for each feature.
4. Keep this checklist current—mark rows `Done` only when all three stages above are satisfied and validated (`./run_all_tests.sh`, interpreter suites, `go test ./...`).
