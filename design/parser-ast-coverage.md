# Parser & AST Coverage Checklist (Able v10)

## Purpose
- Track every surface feature defined in `spec/full_spec_v10.md` and confirm that:
  - the parser produces the correct concrete syntax tree and canonical AST, and
  - the shared AST fixture suite exercises the feature (both interpreters consume these fixtures).
- Serve as the canonical backlog for parser/AST gaps until every item is verified by dedicated tests/fixtures.

## Status Legend
- `TODO` – No targeted coverage yet (parser + fixtures missing).
- `Partial` – Some coverage exists but is incomplete (e.g., only one interpreter, limited scenarios).
- `Done` – Parser tests and AST fixtures both cover the feature with representative scenarios.

> **Note:** “Parser Tests” refers to focused assertions in `interpreter10-go/pkg/parser/*`. “AST Fixtures” refers to entries under `fixtures/ast` exported via `interpreter10/scripts/export-fixtures.ts`.

---

## Expressions & Literals
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Integer literals (suffix variants) | §6.1.1 | `expressions/int_addition`, `expressions/integer_suffix` | TODO | Partial | Parser assertions still needed; interpreter currently normalises to i32. |
| Float literals (`f32`, `f64`) | §6.1.2 | `expressions/mixed_numeric_arithmetic`, `expressions/float_suffix` | TODO | Partial | Parser coverage pending; interpreter normalises to f64 today. |
| Boolean literals | §6.1.3 | `basics/bool_literal`, `expressions/bool_literal_true` | TODO | Partial | Parser coverage pending. |
| Character literals | §6.1.4 | `basics/char_literal`, `basics/char_escape` | TODO | Partial | Add parser tests (including additional escape cases). |
| String literals | §6.1.5 | `basics/string_literal` | `TestParseModuleImports` | Partial | Parser test is incidental; add dedicated literal-focused test. |
| Nil literal | §6.1.6 | `basics/nil_literal` | TODO | Partial | Parser coverage pending. |
| Array literal | §6.8 | `patterns/array_destructuring`, `expressions/array_literal_empty`, `expressions/array_literal_typed` | TODO | Partial | Parser coverage pending; fixtures cover empty and typed array cases. |
| Struct literal (named fields) | §4.5.2 | `structs/named_literal` | TODO | Partial | Parser test missing. |
| Struct literal (positional fields) | §4.5.3 | `structs/positional_literal` | TODO | Partial | Parser coverage pending. |
| Unary expressions | §6.3.2 | `expressions/unary_negation`, `expressions/unary_not`, `expressions/unary_bitwise`, `expressions/unary_double_negation` | TODO | Partial | Parser tests pending; fixtures now cover nested unary sequences. |
| Binary operators (arithmetic) | §6.3.2 | `expressions/int_addition` | `TestParseModuleImports` | Partial | Expand fixtures/tests for precedence matrix. |
| Block expression (`do {}`) | §6.2 | `expressions/block_expression` | TODO | Partial | Parser coverage pending for standalone blocks and nested scopes. |
| If / if-or / else | §6.5 | `control/if_else_branch`, `control/if_or_else` | TODO | Partial | Parser tests should cover or-clauses and trailing else blocks. |
| Match expression (identifier + literal) | §6.5 / §4.6 | `match/identifier_literal`, `match/guard_clause`, `match/wildcard_pattern` | TODO | Partial | Parser tests should cover literal, guard, and wildcard fallback clauses. |
| Match expression (struct guard) | §6.5 | `match/struct_guard`, `match/struct_positional_pattern` | TODO | Partial | Parser coverage pending for guard expressions and positional struct patterns. |
| Lambda expression (inline) | §6.4 | `functions/lambda_expression` | TODO | Partial | Parser tests pending; add typed/closure variants. |
| Trailing lambda call | §6.4 | `functions/trailing_lambda_call` | TODO | Partial | Parser assertions for `isTrailingLambda` pending. |
| Function call with type arguments | §6.4 | `functions/generic_application` | `TestParseModuleImports` | Partial | Add dedicated coverage for multiple type args + parser tests. |
| Range expression | §6.8 | `control/for_range_break`, `control/range_inclusive` | TODO | Partial | Parser assertions for inclusive/exclusive bounds pending. |
| Member access | §4.5 / §6.4 | `strings/interpolation_struct_to_string` | `TestParseModuleImports` | Partial | Expand to positional access, chained access. |
| Index expression | §6.8 | `expressions/index_access` | TODO | Partial | Parser coverage pending; add bounds/error variants. |
| String interpolation | §6.6 | `strings/interpolation_basic`, `strings/string_literal_escape` | TODO | Partial | Parser test missing; add multi-part + escape cases. |
| Propagation expression (`!`) | §11.2.2 | `errors/or_else_handler`, `expressions/or_else_success` | TODO | Partial | Parser tests pending; covers both failure and success flows. |
| Or-else expression (`else {}`) | §11.2.3 | `errors/or_else_handler`, `expressions/or_else_success` | TODO | Partial | Parser coverage pending; add guard/binding variations. |
| Rescue / ensure expressions | §11.3.2 / §11.3.2 | `errors/rescue_catch`, `errors/ensure_runs`, `expressions/rescue_success`, `expressions/ensure_success` | TODO | Partial | Parser tests pending; covers success and failure cases. |
| Breakpoint expression | §6.5 | `expressions/breakpoint_value` | TODO | Partial | Parser coverage pending; add labeled + nested use cases. |
| Proc expression | §12.2 | `concurrency/proc_cancel_value` | TODO | Partial | Parser coverage pending. |
| Spawn expression | §12.3 | `concurrency/future_memoization` | TODO | Partial | Parser coverage pending. |
| Generator literal (`Iterator {}`) | §6.7 | `expressions/iterator_literal` | TODO | Partial | Parser coverage pending; handles basic yield sequences. |

## Patterns & Assignment
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Declaration assignment (`:=`) | §5.1 | `expressions/assignment_declare` | TODO | Partial | AST fixture in place; interpreter assertions + parser tests still needed. |
| Reassignment (`=`) | §5.1 | `expressions/reassignment` | TODO | Partial | Parser coverage pending. |
| Compound assignments (`+=`, etc.) | §5.1 | `expressions/compound_assignment`, `expressions/compound_assignment_variants` | TODO | Partial | Parser tests pending; fixtures now cover arithmetic, bitwise, and shift operators. |
| Identifier pattern | §5.2.1 | `patterns/named_struct_destructuring` | TODO | Partial | Ensure explicit parser assertions. |
| Wildcard pattern (`_`) | §5.2.2 | `patterns/wildcard_assignment`, `match/wildcard_pattern` | TODO | Partial | Parser tests pending; fixtures now include match fallback usage. |
| Struct pattern (named fields) | §5.2.3 | `match/struct_guard`, `match/nested_struct_pattern`, `patterns/nested_struct_destructuring` | TODO | Partial | Parser coverage pending; fixtures include assignment and nested match cases. |
| Struct pattern (positional) | §5.2.4 | `patterns/struct_positional_destructuring`, `match/struct_positional_pattern` | TODO | Partial | Parser coverage pending; fixtures cover assignment and match destructuring. |
| Array pattern | §5.2.5 | `patterns/array_destructuring`, `match/array_pattern` | TODO | Partial | Parser coverage pending; fixtures cover assignment and match forms. |
| Nested patterns | §5.2.6 | `patterns/named_struct_destructuring`, `patterns/nested_struct_destructuring`, `patterns/nested_array_destructuring`, `match/nested_struct_pattern` | TODO | Partial | Parser assertions for nested destructuring still needed. |
| Typed patterns | §5.2.7 | `patterns/typed_assignment` | TODO | Partial | Parser tests pending. |

## Declarations
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Function definition (with params, return type) | §6.4 / §10 | `functions/generic_application`, `strings/interpolation_struct_to_string` | `TestParseModuleImports` | Partial | Need fixtures for where clauses + parser tests. |
| Function generics (`fn<T>`) | §10.2 | `functions/generic_application`, `functions/generic_multi_parameter` | TODO | Partial | Parser coverage pending; fixtures cover single and multi-parameter generics. |
| Private functions (`private fn`) | §13.5 | `privacy/private_static_method` | TODO | Partial | Parser coverage pending. |
| Struct definition (named fields) | §4.5.2 | `structs/named_literal` | TODO | Partial | Parser tests missing. |
| Struct definition (positional fields) | §4.5.3 | `structs/positional_literal` | TODO | Partial | Parser coverage pending. |
| Union definition | §4.6 | `unions/simple_match`, `unions/generic_result` | TODO | Partial | Parser tests pending; fixtures cover simple and generic multi-variant unions. |
| Interface definition | §10.1 | `errors/generic_constraint_unsatisfied`, `declarations/interface_impl_success` | TODO | Partial | Parser tests pending; success + error cases covered. |
| Implementation definition (`impl`) | §10.3 | `declarations/interface_impl_success` | TODO | Partial | Parser coverage pending; add generics/named impl variants. |
| Methods definition (`methods for`) | §10.3 | `strings/interpolation_struct_to_string`, `privacy/private_static_method` | TODO | Partial | Parser coverage pending. |
| Extern function body | §16.1.2 | `interop/prelude_extern` | TODO | Partial | Parser coverage pending; fixture exercises Go target metadata round-trip. |
| Prelude statement | §16.1.1 | `interop/prelude_extern` | TODO | Partial | Parser coverage pending; fixture covers per-target storage and no-op execution. |

## Control Flow Statements
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| While loop | §6.5 | `control/while_sum` | TODO | Partial | Parser coverage pending. |
| For loop (`for x in`) | §6.5 / Iteration | `control/for_sum` | TODO | Partial | Need parser assertions for iterator patterns. |
| Break statement (with/without label) | §6.5 | `control/for_range_break` | TODO | Partial | Add parser coverage for labeled breaks. |
| Continue statement | §6.5 | `control/for_continue` | TODO | Partial | Parser tests pending. |
| Return statement | §11.1 | `strings/interpolation_struct_to_string` (method) | `TestParseModuleImports` | Partial | Add tests covering bare `return`, value returns. |
| Raise statement | §11.3.1 | `errors/raise_manifest` | TODO | Partial | Parser coverage pending. |
| Rescue block | §11.3.2 | `errors/rescue_catch`, `errors/rescue_guard` | TODO | Partial | Parser tests pending. |
| Ensure block | §11.3.2 | `errors/ensure_runs` | TODO | Partial | Parser coverage pending. |
| Rethrow statement | §11.3.2 | `errors/rethrow_propagates` | TODO | Partial | Parser tests pending; fixture ensures rethrow surfaces original error. |

## Types & Generics
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Simple type expression | §4.1.2 | `functions/generic_application`, `errors/generic_constraint_unsatisfied` | `TestParseModuleImports` | Partial | Need explicit coverage for builtins + parser assertions. |
| Generic type expression (`Array<T>`) | §4.1.2 | `types/generic_type_expression` | TODO | Partial | Parser coverage pending; ensure parser emits type arg nodes. |
| Function type expression (`(T) -> U`) | §4.1.2 | `types/function_type_expression` | TODO | Partial | Parser tests pending for arrow syntax + lambda inference. |
| Nullable type (`T?`) | §4.6.2 | `types/nullable_type_expression` | TODO | Partial | Parser coverage pending; add non-nil variant. |
| Result type (`!T`) | §11.2.1 | `types/result_type_expression` | TODO | Partial | Parser tests needed for success + error propagation signatures. |
| Union type syntax | §4.6 | `types/union_type_expression` | TODO | Partial | Parser coverage pending; include multi-branch unions & literals. |
| Type parameter constraints (`where`) | §4.1.5 | `types/generic_where_constraint` | TODO | Partial | Parser tests should assert clause ordering + multiple constraints. |
| Interface constraint (`T: Display`) | §4.1.5 | `errors/generic_constraint_unsatisfied`, `types/generic_where_constraint` | TODO | Partial | Success-case now covered; add parser assertions for constraint lists. |

## Modules & Imports
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Package statement | §13.2 | `imports/package_statement` | TODO | Partial | Parser coverage pending; fixture registers package namespace for imports. |
| Import selectors (`import pkg.{Foo, Bar as B}`) | §13.4 | `imports/static_alias_public` | `TestParseModuleImports` | Partial | Add standalone parser tests. |
| Wildcard import (`*`) | §13.4 | `imports/static_wildcard` | `TestParseModuleImports` | Partial | Parser roundtrip coverage pending. |
| Alias import (`import pkg as alias`) | §13.4 | `imports/static_alias_public` | TODO | Partial | Parser coverage pending. |
| Dynamic import (`dynimport`) | §13.4 | `imports/dynimport_wildcard`, `imports/dynimport_selector_alias` | `TestParseModuleImports` | Partial | Add parser assertions for dynimport selectors. |
| Prelude statement (`prelude {}`) | §16.1.1 | `interop/prelude_extern` | TODO | Partial | Parser coverage pending; need fixtures for additional targets + parser assertions. |

## Concurrency & Async
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| `proc` block/expression | §12.2 | `concurrency/proc_cancel_value`, `concurrency/proc_yield_flush` | TODO | Partial | Parser coverage pending; fixtures exercise cancellation and resolution paths. |
| `proc` helpers (`proc_yield`, etc.) | §12.2.4 | `concurrency/proc_cancelled_helper`, `concurrency/proc_yield_flush` | TODO | Partial | Parser tests should cover helper invocation sites (`proc_yield`, `proc_cancelled`, `proc_flush`). |
| `spawn` expression | §12.3 | `concurrency/future_memoization` | TODO | Partial | Parser assertions pending for `spawn` syntax and future member calls. |
| Channel literal/ops | §12.5 | `concurrency/channel_basic_ops` | TODO | Partial | Fixture exercises new/send/receive/try/close/is_closed semantics; parser assertions still needed. |
| Mutex helper (`mutex`) | §12.5 | `concurrency/mutex_locking` | TODO | Partial | Fixture covers sequential lock/unlock usage; extend once concurrency helpers exist. |

## Error Handling & Options
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Option/Result propagation (`!`) | §11.2.2 | `expressions/or_else_success`, `errors/or_else_handler`, `types/result_type_expression`, `unions/generic_result` | TODO | Partial | Parser coverage pending; fixtures cover success + failure propagation paths. |
| Option/Result handlers (`else {}`) | §11.2.3 | `expressions/or_else_success`, `errors/or_else_handler`, `types/result_type_expression`, `unions/generic_result` | TODO | Partial | Parser tests should assert handler binding + multi-clause cases. |
| Exception raising/rescue/ensure | §11.3 | `errors/raise_manifest`, `errors/rescue_catch`, `errors/rescue_guard`, `errors/rescue_typed_pattern`, `errors/rethrow_propagates`, `errors/ensure_runs`, `expressions/ensure_success` | TODO | Partial | Add parser assertions for rescue guards, ensure blocks, and typed patterns. |
| Breakpoint labels | §6.5 | `expressions/breakpoint_value`, `control/for_range_break` | TODO | Partial | Parser coverage pending for labeled break statements within loops/expressions. |

## Host Interop
| Feature | Spec Reference | AST Fixtures | Parser Tests | Status | Notes |
|---------|----------------|--------------|--------------|--------|-------|
| Host preludes (`prelude { ... }`) | §16.1.1 | `interop/prelude_extern` | TODO | Partial | Fixture validates Go-target prelude storage; parser tests still needed. |
| Extern function bodies | §16.1.2 | `interop/prelude_extern` | TODO | Partial | Fixture covers extern bodies stored in interpreter extras; add parser assertions + multi-target cases. |

---

## Next Steps (Execution Order)
1. **AST Fixtures First** – Expand `fixtures/ast` (via `export-fixtures.ts`) until every feature in the table has a representative module that can be consumed by both interpreters.
2. **Interpreter Verification Second** – Extend the interpreter test suites (TS + Go) so each fixture is evaluated and its behavior/assertions are checked, guaranteeing runtime support for every AST form.
3. **Parser Coverage Third** – Once fixtures/interpreter checks exist, add focused parser unit tests that parse the surface syntax into the canonical AST and confirm round-trips for each feature.
4. Keep this checklist current—mark rows `Done` only when all three stages above are satisfied and validated (`bun run scripts/run-fixtures.ts`, interpreter suites, `go test ./...`).
