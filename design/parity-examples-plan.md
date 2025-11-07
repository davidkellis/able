# Parity Examples Expansion Plan

## Goals
- Provide curated Able programs under `interpreter10/testdata/examples/` that exercise language features end-to-end (parse → typecheck → interpret) while remaining deterministic and fast.
- Keep the curated suite aligned with the parser/AST readiness checklist and the Go runtime capabilities so parity tests stay green.
- Capture blockers discovered while prototyping new examples (pattern guards, advanced pipes) so future contributors know what needs to land before expanding the suite.

## Current Coverage (2025-11-07)
| Example | Feature Focus | Status |
|---------|---------------|--------|
| `hello_world` | basic string concat + `print` | ✅ Runs in parity suite |
| `control_flow` | `for` loop, arithmetic accumulation | ✅ |
| `errors_rescue` | `raise`/`rescue` control flow | ✅ |
| `structs_translate` | struct construction + destructuring assignment | ✅ |
| `concurrency_proc` | proc memoization + `value()` reuse | ✅ |
| `generics_typeclass` | generic helpers + type arguments | ✅ |

## Pending Examples & Blockers
| Candidate | Desired Focus | Blockers | Notes |
|-----------|---------------|----------|-------|
| `patterns_match` | `match` clauses with guards and struct patterns | Parser still rejects guard syntax in TypeScript harness (tree-sitter mapping error). Track `design/parser-ast-coverage.md` – section “Match expression (struct guard)” still `TODO` for parser tests with `source.able`. | Revisit after parser sweep re-enables guard support. |
| `pipes_topics` | `%` topic references, placeholder-generated callables, method pipes | Go runtime still reports arity/type errors for placeholders when chained via `%`. Follow `design/concurrency-fixture-plan.md` and `fixtures/pipes/topic_placeholder` parity work. | Add once Go topic helper achieves parity with TS. |
| `patterns_destructure` | struct destructuring with conditional logic | Parser rejects the current `Point { x: x, y: y } := ...` sample due to open parser TODO; see `design/parser-ast-coverage.md` entry for “Binding patterns (struct)”. | Wait until parser+mapper agree on pattern binding metadata. |

## Next Steps
1. Land the outstanding parser fixes called out above (`design/parser-ast-coverage.md` rows for match guards and struct patterns).
2. Mirror the fixes in Go runtime (topic pipes) so both interpreters evaluate the new programs identically.
3. Add the new curated samples once the respective blockers are closed:
   - `patterns_match` exercising guards and wildcard fallbacks.
   - `pipes_topics` covering `%`/`@` placeholders + topic methods.
   - `patterns_destructure` showcasing `Point { x, y }` binding with simple control flow.
4. Update `interpreter10/test/parity/examples_parity.test.ts` and this plan whenever examples are added/removed.

## Notes
- Each curated program should complete in milliseconds, avoid randomness, and print concise output so parity diffs stay readable.
- Don’t reuse the repository’s `examples/` directory; that content remains documentation-focused and may rely on features still under development.
