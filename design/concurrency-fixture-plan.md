# Concurrency Fixture Plan

Date: 2025-10-19

- Target fixtures: concurrency/proc_cancel_value, future_memoization, proc_cancelled_outside_error, proc_cancelled_helper.
- Added Go test harness to run these fixtures under goroutine executor (see pkg/interpreter/fixtures_concurrency_test.go); harness now reuses shared fixture helpers so parity pipelines stay consistent.
- Next: integrate fixture execution into shared parity harness; avoid TS fairness cases.
