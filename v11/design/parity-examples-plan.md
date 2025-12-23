# Parity Examples Expansion Plan

## Goals
- Provide curated Able programs under `v11/interpreters/ts/testdata/examples/` that exercise language features end-to-end (parse → typecheck → interpret) while remaining deterministic and fast.
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
| `patterns_match` | `match` guards + struct patterns + ordered fallbacks | ✅ Added to parity suite (parser + mapper now stable) |
| `patterns_destructure` | struct destructuring inside loops + matches | ✅ |
| `pipes_topics` | pipeline callables, placeholder-based transforms, method pipes | ✅ |
| `dynimport_parity` | `dynimport` aliasing, dyn packages, selector imports | ✅ |
| `dynimport_multiroot` | Multi-root `dynimport` (external package roots via `ABLE_MODULE_PATHS`) | ✅ Added with shared deps under `v11/interpreters/ts/testdata/examples/deps/` |

## Pending Examples & Blockers
| Candidate | Desired Focus | Blockers | Notes |
|-----------|---------------|----------|-------|
| `channel_select` | Multi-branch `select` exercising channel send/receive readiness, default clauses, and cancellation helpers (`proc_cancelled`, `proc_flush`). | Tree-sitter + Go/TS interpreters lack `select` syntax/evaluator support; spec work tracked in `spec/todo.md` (`Channel select semantics`). | Draft the program once the parser + runtimes expose the `select` AST nodes. Should cover buffered/unbuffered channels plus cancellation. |

## Next Steps
1. Keep the new parity samples green—update them whenever the underlying semantics change (match guards, destructuring assignment rules, or pipe semantics).
2. Brainstorm the next curated additions once upcoming work (e.g., channel select, advanced dynimport scenarios) lands and the parser/Go runtime expose the new constructs.
3. Continue updating `v11/interpreters/ts/test/parity/examples_parity.test.ts` and this plan whenever examples are added/removed so downstream contributors know the current coverage.

## Notes
- Each curated program should complete in milliseconds, avoid randomness, and print concise output so parity diffs stay readable.
- Don’t reuse the repository’s `examples/` directory; that content remains documentation-focused and may rely on features still under development.
