# Compiled native Await service carriers retained

Date: 2026-07-28

## Decision

Retain lazy native carriers for generated `AwaitWaker` and
`AwaitRegistration` service values under the existing experimental callable
execution-context gate.

Compiled Await now keeps its scheduler wake and cancellation operations in
ordinary Go objects. The established map-backed semantic structs and native
callable fields are created only when an arbitrary user-defined Awaitable,
explicit protocol call, or interpreter/dynamic boundary needs them. Native
channel and mutex registration paths invoke the Go operation directly.

This is a general language-syntax/kernel rule. It does not name a benchmark,
Channel, Mutex, stdlib container, or non-primitive user nominal type.

## Fresh owner attribution

Ten main-only CPU profiles and three exact allocation profiles per application
were collected from fresh fallback-free binaries before the prototype. Five
diagnostic processes per application then counted construction directly:

| Application | Await wakers per run | Await registrations per run |
| --- | ---: | ---: |
| Await Channel Mux | 1,024 | 1,024 |
| Mutex Await Journal | 2,048 | 50-82 |
| Mutex Work Queue | 4,096 | 198-214 |

Exact allocation attribution showed about five objects at both constructor
sites for every service value: a struct instance, field map, callable value,
closure, and associated storage. The candidate profiles reduce both sites to
one escaping carrier object per construction. Escape diagnostics confirm that
the carrier is the intended single escape and that semantic map/callable
allocation remains deferred to `MaterializeRuntimeValue`.

## Allocation A/B

Five rotating, public-verifier-backed main-phase processes per side measured:

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 6,617,873.6 | 5,555,548.8 | -16.05% | 112,672.6 | 101,911.2 | -9.55% |
| Mutex Await Journal | 2,820,513.6 | 1,798,961.6 | -36.22% | 46,018.2 | 37,606.2 | -18.28% |
| Mutex Work Queue | 5,797,483.2 | 3,776,486.4 | -34.86% | 95,528.0 | 79,056.6 | -17.24% |

Both allocation measures improve in all three unlike programs.

## Balanced timing and Go comparison

Fifteen rotating baseline/candidate/equivalent-Go cohorts per application
measured:

| Application | Baseline | Candidate | Raw change | Paired 95% interval | Go | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.083843s | 0.082677s | -1.39% | -6.22% to +2.84% | 0.003226s | 25.63x |
| Mutex Await Journal | 0.008299s | 0.006087s | -26.66% | -31.08% to -21.24% | 0.001853s | 3.28x |
| Mutex Work Queue | 0.016371s | 0.012476s | -23.80% | -26.51% to -20.62% | 0.002332s | 5.35x |

Raw means improve in all three. Journal and Queue have clear paired wall-time
gains; Channel is neutral within workstation noise. The rule is retained for
its broad allocation reduction plus two strong unlike-program timing wins.
No Channel timing improvement is claimed.

The remaining 3.90%, 30.49%, and 18.69% of equivalent-Go performance still
miss the 95% product target.

## Semantic and scope gates

- A generated-source guard requires both native service carriers only under
  the reached experimental gate.
- A user-defined Awaitable receives a materialized `AwaitWaker`, calls
  `wake()`, and completes correctly.
- An explicit native Awaitable `register` call returns a materializable
  registration whose `cancel()` works through ordinary member dispatch.
- Default builds of all three applications and an experimental await-free
  N-body build remain byte-identical to the pre-change compiler.
- All three strict candidate graphs omit `pkg/interpreter`; every public
  verifier passes.
- Focused closure-owned callable, goroutine, Future, Mutex, execution-context,
  dynamic-boundary, cross-package/interface, bridge, and compiled-CLI tests
  pass.
- Full experimental execution-context fixture parity passes in 49.743
  seconds.
- Every touched source file remains below 1,000 lines.
- The scope-only ledger update changes only the `compiler-production` tree
  hash. Its seven tests pass with 21 current closures and zero invalidations.

No canonical stdlib, interpreter, bytecode VM, language, dependency,
application source, non-primitive nominal rule, or WASM change was needed.

Machine-readable aggregate:

- `2026-07-28-compiled-native-await-service-carriers-retained.json`

## Next

Count and attribute `__able_context_from_native` and
`__able_context_from_environment` construction across the same three
applications.

Why: after removing four allocations from every waker and registration, the
fresh candidate profiles show context reconstruction as the largest remaining
generated allocation owner shared by all three.

What it entails: add exact caller and construction counters, distinguish task
creation from native-call reverse adaptation, and prototype only an
allocation-free general reverse context view that preserves captured,
cross-package, nested, error, and dynamic compatibility semantics. Repeat the
same public-verifier allocation and balanced equivalent-Go protocol.

Why it is important: compiled Await remains 3.28x-25.63x slower than Go.
Removing a repeated synthetic execution-context wrapper would carry the
already-native Go state farther through the kernel call path, while exact
attribution prevents reopening the previously rejected broad context ABI.
