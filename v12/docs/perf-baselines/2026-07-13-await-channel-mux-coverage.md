# Await Channel Mux Cross-Language Coverage — 2026-07-13

## Decision

Add `await-channel-mux` and retain two generic standalone-compiler correctness
repairs; add no performance optimization and make no canonical-stdlib change.
The completed Future Await Race application exercises Future-backed await
arms, but did not exercise selection among public Channel Awaitables or an
`Await.default` arm. This application fills that separate kernel/runtime
boundary.

During validation, standalone generated binaries rejected both a real
`Future<T>` and a kernel-shaped `Awaitable<T>` at static runtime boundaries.
The compiler bridge's no-interpreter matcher now accepts ordinary runtime
Future and Iterator values and recognizes the documented Awaitable callable
shape (`is_ready`, `register`, `commit`, and `is_default`). This mirrors the
interpreter's existing semantic treatment of those runtime-only values. It is
not a named-container lowering rule: Awaitable is the language's kernel
protocol used by `await` syntax.

## Application

Each of 512 rounds blocks a receiver on either of two public
`Channel.await_receive` arms, uses the existing `future_flush()` progress
barrier to establish that blocked state, then sends one lane-specific numeric
event. A second task selects the same two empty receive arms with
`Await.default`, contributing `-1`. The result is independent of scheduling:

```
512:258379980:-512
```

The canonical Able source is
`v12/examples/benchmarks/await_channel_mux/await_channel_mux.able`. The
sibling `../benchmarks/await-channel-mux` suite contains equivalent Able, Go
1.26, Python 3.14, and Ruby 4.0 programs and one shared Ruby verifier.

This is distinct from Channel Rollup's producer/worker pipeline, Future
Pipeline's Future cancellation flow, and Future Await Race's Future-only join
arms. It deliberately uses no named-container lowering rule or benchmark-only
runtime path.

## Harness Integration and Verification

`v12/bench_external_catalog.sh` registers `await_channel_mux` in its own
suite and in `concurrency` and `coverage`, with goroutine execution and
source-root isolation. `generality` remains unchanged.

- Bytecode and tree-walker Able both completed the canonical source under the
  goroutine executor with the shared result.
- A compiled Able binary completed with the same result.
- Go, Python, and Ruby reference implementations each completed with the same
  result.
- `bench_bytecode_audit --benchmarks await_channel_mux` passed.
- `bench_compiled_boundary_audit --benchmarks await_channel_mux --timeout 45`
  completed one verified goroutine-executor run with no timeout or failure.
- `TestMatchTypeWithoutInterpreterAcceptsGenericRuntimeAsyncValues` and
  `TestMatchTypeWithoutInterpreterRecognizesKernelAwaitableShape` pass in the
  compiler bridge suite.

The canonical external stdlib already exposes the public Channel and Await
APIs, so `../able-stdlib` required no change.

## Performance status

The completed three-process CPU-15 reference screen makes this the 25th
bytecode ledger row. Bytecode is 0.2700 seconds (2.26x Python and 2.74x Ruby),
while compiled is 0.4367 seconds (92.91x Go); every row is verifier-backed.
Because both product lanes are material misses, eight normal-scheduler,
output-verified profiles were merged per mode. They show bytecode loader/parser
work and generic VM/executor parents, while compiled repeats the known generic
`bridge.currentGID` -> `runtime.Stack` wall. There is no new concrete leaf
shared with Channel Rollup, Future Pipeline, and Future Await Race, so no
performance change is justified.

The complete reference and profile decision is
`2026-07-13-await-channel-mux-reference-profile-gate.md`. The next work should
return to an independently useful unfinished language boundary, because this
Channel/Awaitable application now has semantic, compiler, cross-language, and
profile coverage without identifying a broadly applicable optimization.
