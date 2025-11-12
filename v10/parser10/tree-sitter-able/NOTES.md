# Able v10 Parsing Notes

This workspace tracks reference data pulled from `spec/full_spec_v10.md` while
the grammar comes together. Nothing here is authoritative; the spec remains the
source of truth.

## Reserved keywords

`fn`, `struct`, `union`, `interface`, `impl`, `methods`, `type`, `package`,
`import`, `dynimport`, `extern`, `prelude`, `private`, `Self`, `do`, `return`,
`if`, `or`, `else`, `while`, `for`, `in`, `match`, `case`, `breakpoint`,
`break`, `continue`, `raise`, `rescue`, `ensure`, `rethrow`, `proc`, `spawn`,
`as`, `nil`, `void`, `true`, `false`, `where`.

The grammar should treat these as reserved tokens when parsing identifiers.

## Operator precedence (v10)

Pulled from §6.3.1. Highest precedence is listed first.

| Prec | Operators                                                                 | Notes                                |
|------|---------------------------------------------------------------------------|--------------------------------------|
| 15   | `.`                                                                        | Member access                        |
| 14   | `()`, `[]`, postfix `!`                                                   | Calls, indexing, error propagation   |
| 13   | `^`                                                                        | Exponentiation (right-associative)   |
| 12   | unary `-`, `!`, `~`                                                       | Arithmetic / logical / bitwise NOT   |
| 11   | `*`, `/`, `%`                                                             | Multiplicative                       |
| 10   | `+`, `-`                                                                  | Additive                             |
| 9    | `<<`, `>>`                                                                | Shifts                               |
| 8    | `&`                                                                       | Bitwise AND                          |
| 7    | `\\xor`                                                                   | Bitwise XOR                          |
| 6    | `|`                                                                       | Bitwise OR                           |
| 6    | `>`, `<`, `>=`, `<=`                                                      | Comparisons (non-associative)        |
| 5    | `==`, `!=`                                                                | Equality (non-associative)           |
| 4    | `&&`                                                                      | Logical AND                          |
| 3    | `||`                                                                      | Logical OR                           |
| 2    | `..`, `...`                                                               | Range constructors                   |
| 1    | `:=`, `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `\\xor=`, `<<=`, `>>=` | Assignment family (right-assoc)      |
| 0    | `\\|>`                                                                    | Pipe-forward                          |

Notes:
- Safe navigation (`?.`) is *not* part of v10; postpone until the v11 grammar.
- Equality/comparison operators are non-associative (no chaining without
  explicit parentheses).
- Pipe-forward is the only operator defined below assignment.

## Immediate parser TODOs

1. Confirm the v10 story for `type` aliases / module metadata before adding
   syntax; declaration generics + composite interfaces + host interop are now
   covered.
2. Thread placeholder/topic tokens and enriched async/error nodes through the
   AST builders once the parser integration layer is in place.
3. Expand corpus coverage once grammar stabilises (declarations, async/error
   expressions, advanced patterns) and start mapping parser output to AST
   builders for integration tests.
