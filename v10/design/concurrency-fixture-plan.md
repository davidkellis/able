# Concurrency Fixture Plan

Date: 2025-10-19

- Target fixtures: concurrency/proc_cancel_value, future_memoization, proc_cancelled_outside_error, proc_cancelled_helper.
- Added Go test harness to run these fixtures under the goroutine executor (initially via `pkg/interpreter/fixtures_concurrency_test.go`, now folded into `pkg/interpreter/fixtures_parity_test.go`) so parity pipelines stay consistent.
- 2025-10-23: Integrated fixture execution into the shared parity harness by selecting the goroutine executor inside `TestFixtureParityStringLiteral`; removed the dedicated concurrency harness.
- 2025-10-23 (later): TypeScript interpreter now uses the shared executor contract (`src/interpreter/executor.ts`), so the fixture harness exercises the same cancellation/yield scenarios by default; no additional TS-specific wiring required beyond the cooperative executor.
- 2025-10-23 (later): Added Bun unit coverage (`test/concurrency/channel_mutex.test.ts`) for unbuffered rendezvous, buffered back pressure, close wake-ups, and cancellation cleanup to guard the TS runtime against regressions while the fixtures stay focused on AST wiring.
- 2025-10-24: Serial executor gained an explicit yield path so fixtures can assert deterministic round-robin traces. Added fairness fixtures (`concurrency/fairness_proc_round_robin`, `concurrency/fairness_proc_future`) and wired them into both harnesses.

# Fairness Fixture Status

- **Serial executor coordination**: Implemented; `proc_yield` now requeues serial tasks without touching production goroutine behaviour.
- **Fixture coverage**: Round-robin proc and proc/future fairness fixtures live under `fixtures/ast/concurrency/` and run green in TS + Go parity suites.
