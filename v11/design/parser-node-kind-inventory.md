# Tree-sitter Node Inventory (Able v10)

This note records the tree-sitter node kinds we must handle in the Go parser for every feature tracked in `design/parser-ast-coverage.md`. Use it as the handoff checklist while we extend `parseExpression`, `parseStatement`, and related helpers.

## Expressions & Literals
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Integer literals (suffix variants) | `number_literal` | Grammar only recognises digits/underscores; add suffix/hex/float handling. |
| Float literals (`f32`, `f64`) | `number_literal` | Decimal points/exponents are not parsed yet; grammar update required. |
| Boolean literals | `boolean_literal` | Present. |
| Character literals | — | No `character_literal` rule yet. |
| String literals | `string_literal` | Present, no interpolation support. |
| Nil literal | `nil_literal` | Present. |
| Array literal | `array_literal` | Present. |
| Struct literal (named fields) | — | No struct literal production yet (needs `struct_literal`). |
| Struct literal (positional fields) | — | Same as above; grammar lacks tuple-style literal. |
| Unary expressions | `unary_expression` | Present. |
| Binary operators (arithmetic/bitwise) | `logical_or_expression`, `logical_and_expression`, `bitwise_or_expression`, `bitwise_xor_expression`, `bitwise_and_expression`, `equality_expression`, `comparison_expression`, `shift_expression`, `additive_expression`, `multiplicative_expression`, `exponent_expression` | All precedence tiers exist; parser must fold into AST operator nodes. |
| Block expression (`do {}`) | `do_expression`, `block` | Present. |
| If / elsif / else | `if_expression`, `elsif_clause`, `block` | Present. |
| Match expression (identifier + literal) | `match_expression`, `match_clause`, `match_guard`, `pattern`, `literal_pattern` | Present. |
| Match expression (struct guard) | `match_expression`, `match_clause`, `match_guard`, `struct_pattern`, `struct_pattern_field` | Present. |
| Lambda expression (inline) | `lambda_expression`, `lambda_parameter_list`, `lambda_parameter` | Present. |
| Trailing lambda call | `postfix_expression`, `call_suffix`, optional `lambda_expression` | Present; need trailing lambda flag in AST. |
| Function call with type arguments | `postfix_expression`, `call_suffix`, `type_arguments` | Present for expression calls. |
| Range expression | `range_expression` | Present (`..` / `...`). |
| Member access | `member_access` | Present. |
| Index expression | `index_suffix` | Present. |
| String interpolation | — | Grammar currently emits plain `string_literal`; interpolation rule missing. |
| Propagation expression (`!`) | `propagate_suffix` | Present. |
| Or-else expression (`else {}`) | — | No dedicated rule; needs grammar support (likely `or_else_expression`). |
| Rescue / ensure expressions | `rescue_expression`, `rescue_block`, `match_clause`, `ensure_expression` | Present. |
| Breakpoint expression | `breakpoint_expression`, `label`, `block` | Present. |
| Proc expression | `proc_expression`, `block` / `do_expression` / `call_target` | Present. |
| Spawn expression | `spawn_expression`, `block` / `do_expression` / `call_target` | Present. |
| Generator literal (`Iterator {}`) | — | No iterator literal rule yet. |

## Patterns & Assignment
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Declaration assignment (`:=`) | `assignment_expression`, `assignment_operator (:=)`, `assignment_target` | Present. |
| Reassignment (`=`) | `assignment_expression`, `assignment_operator (=)`, `assignment_target` | Present. |
| Compound assignments (`+=`, etc.) | `assignment_expression`, `assignment_operator (+=/-=/...)`, `assignment_target` | Present. |
| Identifier pattern | `pattern`, `pattern_base`, `identifier` | Present. |
| Wildcard pattern (`_`) | `pattern_base` (anonymous `_`) | Present. |
| Struct pattern (named fields) | `struct_pattern`, `struct_pattern_field` | Present. |
| Struct pattern (positional) | `struct_pattern` with bare `pattern` elements | Present. |
| Array pattern | `array_pattern`, `array_pattern_rest` | Present. |
| Nested patterns | Combination of `struct_pattern`, `array_pattern`, `typed_pattern`, etc. | Present. |
| Typed patterns | `typed_pattern`, `type_expression` | Present. |

## Declarations
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Function definition (with params, return type) | `function_definition`, `type_parameter_list`, `parameter_list`, `parameter`, `return_type`, `where_clause` | Present. |
| Function generics (`fn<T>`) | `function_definition`, `type_parameter_list` | Present. |
| Private functions (`private fn`) | `function_definition` with leading `private` token | Present. |
| Struct definition (named fields) | `struct_definition`, `struct_record`, `struct_field` | Present. |
| Struct definition (positional fields) | `struct_definition`, `struct_tuple` | Present. |
| Union definition | `union_definition` | Present. |
| Interface definition | `interface_definition`, `interface_member`, `function_signature`, `interface_composition` | Present. |
| Implementation definition (`impl`) | `implementation_definition`, `named_implementation_definition` | Present. |
| Methods definition (`methods for`) | `methods_definition` | Present. |
| Extern function body | `extern_function`, `host_code_block`, `host_code_chunk` | Present. |
| Prelude statement | `prelude_statement`, `host_code_block` | Present. |

## Control Flow Statements
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| While loop | `while_statement` | Present. |
| For loop (`for x in`) | `for_statement`, `pattern`, `expression`, `block` | Present. |
| Break statement (with/without label) | `break_statement`, optional `label`, optional value `expression` | Present. |
| Continue statement | `continue_statement` | Present. |
| Return statement | `return_statement` | Present. |
| Raise statement | `raise_statement` | Present. |
| Rescue block | `rescue_block`, `match_clause`, typically inside `rescue_expression` | Present. |
| Ensure block | `ensure_expression`, `block` | Present. |
| Rethrow statement | `rethrow_statement` | Present. |

## Types & Generics
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Simple type expression | `type_expression`, `type_identifier`, `qualified_identifier` | Present. |
| Generic type expression (`Array<T>`) | — | Grammar lacks type-application (`<...>`) inside types; needs new production. |
| Function type expression (`(T) -> U`) | `type_arrow`, `parenthesized_type`, `type_expression` | Present. |
| Nullable type (`T?`) | `type_prefix` with leading `"?"` | Present. |
| Result type (`!T`) | `type_prefix` with leading `"!"` | Present. |
| Union type syntax | `type_union` | Present. |
| Type parameter constraints (`where`) | `where_clause`, `where_constraint`, `type_bound_list` | Present. |
| Interface constraint (`T: Display`) | `type_bound_list`, `qualified_identifier` | Present (same machinery as `where`). |

## Modules & Imports
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Package statement | `package_statement`, `identifier` | Present. |
| Import selectors (`import pkg.{Foo, Bar::B}`) | `import_statement`, `import_clause`, `import_selector` | Present. |
| Wildcard import (`*`) | `import_statement`, `import_wildcard_clause` | Present. |
| Alias import (`import pkg::alias`) | `import_statement`, `import_clause`, alias `identifier` | Present. |
| Dynamic import (`dynimport`) | `import_statement` with `kind` token `dynimport` | Present. |
| Prelude statement (`prelude {}`) | `prelude_statement`, `host_code_block` | Present. |

## Concurrency & Async
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| `proc` block/expression | `proc_expression`, `block` / `do_expression` / `call_target` | Present. |
| `proc` helpers (`proc_yield`, etc.) | Standard `postfix_expression` + `identifier` calls | No dedicated grammar; handled as regular call expressions. |
| `spawn` expression | `spawn_expression`, `block` / `do_expression` / `call_target` | Present. |
| Channel literal/ops | — | No send/receive (`<-`) rules yet; grammar addition required. |
| Mutex helper (`mutex`) | Standard `postfix_expression` call | No dedicated grammar; regular call mapping. |

## Error Handling & Options
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Option/Result propagation (`!`) | `propagate_suffix` | Present. |
| Option/Result handlers (`else {}`) | — | Grammar missing dedicated handler expression. |
| Exception raising/rescue/ensure | `raise_statement`, `rescue_expression`, `rescue_block`, `ensure_expression`, `match_clause` | Present. |
| Breakpoint labels | `breakpoint_expression`, `break_statement`, `label` | Present. |

## Host Interop
| Feature | Tree-sitter node kinds | Grammar notes |
|---------|-----------------------|---------------|
| Host preludes (`prelude { ... }`) | `prelude_statement`, `host_code_block`, nested `host_code_chunk` | Present. |
| Extern function bodies | `extern_function`, `host_code_block`, `host_code_chunk` | Present. |

## Mapping Helper Backlog (parser.go)
- **Expression literals:** add handlers for `boolean_literal`, `nil_literal`, `array_literal`, and extend `parseNumberLiteral` for floats/suffixes; introduce `parseCharacterLiteral` and interpolation-aware string parsing once grammar lands.
- **Struct/Array literals:** implement `parseStructLiteral` (named + positional) and upgrade `parseArrayLiteral` to emit typed initialisers.
- **Control-flow expressions:** add dedicated parsers for `if_expression`, `match_expression`, `rescue_expression`, `ensure_expression`, `proc_expression`, `spawn_expression`, `breakpoint_expression`, and future `or_else_expression` / `iterator_literal`.
- **Binary/unary operators:** replace the current "first child" shortcut with operator-aware folding that covers every infix node (`additive_expression` … `logical_or_expression`) and prefix unary forms.
- **Call chains:** expand trailing lambda handling to ensure type arguments + lambda arguments survive multi-call pipelines.
- **Patterns:** extend `parsePattern` to handle `struct_pattern`, `array_pattern`, `typed_pattern`, nested bindings, and guards.
- **Statements:** support `while_statement`, `for_statement`, `break_statement` (labels + values), `continue_statement`, `raise_statement`, `rethrow_statement`, `expression_statement` fallbacks, and eventually `prelude_statement` / `extern_function`.
- **Declarations:** flesh out `parseFunctionDefinition` (generics, where clause, return types), and add helpers for `struct_definition`, `union_definition`, `interface_definition`, `implementation_definition`, and `methods_definition`.
- **Types:** introduce full `parseTypeExpression` covering unions, arrows, nullable/result modifiers, and future generic type applications.
- **Error handling & options:** once grammar exposes option handlers and channel ops, add corresponding AST constructors.
