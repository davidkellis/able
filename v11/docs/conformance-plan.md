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
| 6.1, 6.3, 8.2.2 | Integer literals, arithmetic, accumulation in loops | `exec/08_02_numeric_sum_loop` | Seeded (renamed from `exec_math_sum`); validates `%`/`+` semantics and loop exit |
| 6.1.8, 6.3, 6.4 | Function calls + pipeline, generator iterator protocol (`Iterator { ... }`) | `exec/06_07_iterator_pipeline` | Seeded (renamed from `exec_iterator_pipeline`); exercises `gen.yield`, `Iterator.next`, `|>` |
| 6.6, 4.6, 8.1.2 | Union match expression with wildcard fallthrough | `exec/08_01_union_match_basic` | Seeded (renamed from `exec_match_union`); match ordering and wildcard coverage |
| 11.3 | Unhandled `raise` causes non-zero exit with message | `exec/11_03_raise_exit_unhandled` | Seeded (renamed from `exec_raise_exit`); captures default exception propagation |
| 11.3 | `rescue` + `ensure` ordering with side effects | `exec/11_03_rescue_ensure` | Seeded; ensures rescue handler runs before ensure |
| 12.5 | Channels: buffered send/receive, close terminates iteration | `exec/12_05_concurrency_channel_ping_pong` | Seeded; validates rendezvous order and nil on close |
| 12.2, 12.3, 12.6 | `proc` vs `spawn` vs `await` scheduling | `exec/12_02_async_proc_spawn_combo` | Seeded: cooperative scheduling with `proc_yield` + `proc_flush`, future status/value |
| 7.4, 9, 10 | Method call syntax + UFCS + impl dispatch | `exec/09_04_methods_ufcs_basics` | Seeded: inherent method sugar, UFCS free function, type-qualified static |
| 13.2–13.5 | Packages/import visibility and private types | `exec/13_02_packages_visibility_diag` | Seeded: alias vs selective import scope; private helpers encapsulated |
| 4.7, 4.6, 7.1 | Type aliases/unions with generic functions | `exec/04_07_types_alias_union_generic_combo` | Seeded: alias applied to union + generic function inference fallback |
| 11.2 | `!`/`or` propagation for `Option`/`Result` | `exec/11_02_option_result_propagation` | Seeded: Option unwrap + error handling via `! or {}` |

The matrix is also materialized as `v11/fixtures/exec/coverage-index.json` to enable tooling/CI checks.

## Execution + Reporting
- TS harness: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`.
- Go harness: `cd v11/interpreters/go && go test ./pkg/interpreter -run ExecFixtures`.
- Coverage index lives at `v11/fixtures/exec/coverage-index.json`; `v11/scripts/check-exec-coverage.mjs` (invoked by `run_all_tests.sh`) validates seeded IDs stay in sync with the fixture directories.
