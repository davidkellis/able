# await-channel-mux

Await Channel Mux is a deterministic two-lane dispatch application. Each round
sends one numeric event to either lane, then a separate task selects it through
two Channel Awaitable receive arms. A subsequent idle task uses `Await.default`
only after both lanes are empty.

Each receiver reaches its blocked await before its sender starts, using the
runtime's existing `future_flush()` progress barrier. That makes the workload
deterministic without replacing channel selection with a benchmark-only path.

It covers Channel-backed `Awaitable` registration, commit callbacks, multi-arm
selection, and default selection. It deliberately differs from Channel Rollup's
producer/worker pipeline and Future Await Race's Future-only joins.
