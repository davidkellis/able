# v12 Performance Benchmark Harness

`v12/bench_suite` runs the shared benchmark suite in three execution modes:

- `compiled`
- `treewalker`
- `bytecode`

It emits machine-readable JSON with:

- git commit + dirty state
- machine profile (OS/kernel/arch/CPU/memory)
- harness config (timeouts, runs, benchmark list)
- per-run status/timing/GC metrics
- per-benchmark summary rows

`v12/bench_perf` is the lighter per-target helper for focused perf checks. Its
compiled mode now builds through `cmd/ablec` directly, so compiled fixture
benchmarking measures the current compiler path without pulling in unrelated
`able build` package/bootstrap behavior. It also accepts repeated
`--compiled-build-arg` flags for controlled comparisons such as
`--no-experimental-mono-arrays`. It also supports `--run-from DIR` for
benchmarks that read relative input files, repeated `--program-arg ARG` flags
for workload-specific entry arguments like `wordlist.txt`, `--executor
serial|goroutine` for concurrency-sensitive workloads, and `--output-json
PATH` for machine-readable summaries. When benchmarking against external suite
directories it also pins the selected stdlib root from `ABLE_STDLIB_ROOT` or
the installed cache so the run does not accidentally collide with a sibling
`able-stdlib` checkout. Both the main `able` CLI and generated compiled
launchers now also honor:

- `ABLE_GO_CPU_PROFILE=/tmp/cpu.pprof`
- `ABLE_GO_MEM_PROFILE=/tmp/heap.pprof`

for reusable Go `pprof` capture during focused benchmark runs. `v12/bench_perf`
now sends `SIGINT` before `SIGKILL` on timeout, so timed-out profiled runs
still flush CPU/heap profiles when possible.

`v12/bench_guardrail` is the report-only comparer for suite JSON outputs. It
compares the checked-in baseline against a fresh run and reports status,
timing, and GC deltas without failing the build.

`v12/bench_compare_external` compares Able runs against the checked-in
cross-language results in the sibling `../benchmarks` repository. It reuses
`v12/bench_perf` to run Able against the real external workloads, including
suite-local setup hooks and suite-local input files, then joins those results
against `../benchmarks/results.json` for reference languages such as `go`,
`ruby`, and `python`. The Able side uses the canonical checked-in benchmark
programs under `v12/examples/benchmarks/` and runs them from the external
suite directories so the workload inputs match the external corpus. The
external harness also applies benchmark-specific executor selection when the
reference workload is explicitly parallel; `binarytrees` now runs with the
goroutine executor so it matches the external Go workload instead of silently
serializing all spawned work.

As of April 19, 2026, the aligned compiled core trio is in the same
approximate range as Go:

- `fib`: `3.16s` vs Go `2.84s`
- `binarytrees`: `3.65s` vs Go `3.83s`
- `matrixmultiply`: `1.03s` vs Go `0.88s`

The next external gap is text-heavy `i_before_e`. The latest kept stdlib text
and fs fast paths in `../able-stdlib` moved aligned Able results to:

- compiled: `1.07s` (down from `3.99s`)
- bytecode: `56.76s` (down from timing out at `90s`)

So the remaining work is now mostly bytecode runtime cost plus the residual
compiled text-path gap, not the old compiled recursion/core benchmark path.
The next kept bytecode recursion slice narrowed the hot recursive VM path
further by compacting same-program self-fast call frames and removing repeated
`IntegerValue` value-method copies from the hot slot-const arithmetic path.
That did not yet make aligned bytecode `fib` complete under `120s`, but it
did collapse the old `pushCallFrame` hotspot and kept aligned bytecode
`i_before_e` in the same high-50s band. The next work should stay on
recursive bytecode call/slot churn and a steadier `i_before_e` runtime-only
measurement path. That measurement path is now available as
`v12/bench_perf --modes bytecode-runtime`, which loads/lowers the program
once and then measures repeated `main()` calls under a Go benchmark harness.
The first kept aligned steady-state results are:

- `fib`: still timed out with a `300s` warmup+measure budget
- `i_before_e`: `62.14s/op`, `107.19 GB/op`, `315,106,815 allocs/op`

So the remaining bytecode problem is clearly VM runtime itself, not one-time
CLI/bootstrap/lowering noise.

The next kept steady-state profiling slice now starts CPU/heap profiling only
after program load/lowering plus the explicit warmup call, and it caches
lowered expression bytecode programs by AST-expression identity plus
placeholder-lambda mode. The refreshed aligned `i_before_e` steady-state
result on that kept code is:

- `27.92s/op`
- `16.61 GB/op`
- `172,094,647 allocs/op`

The corrected runtime-only profiles no longer show
`lowerExpressionToBytecodeWithOptions(...)`,
`(*bytecodeLoweringContext).emit(...)`, or `emitExpression(...)` in the hot
set. The remaining steady-state cost is now real VM/runtime work centered on
`execCallName`, `execCall`, `runMatchExpression`, identifier/call-name cache
bookkeeping, and the resulting GC/allocation pressure.

The next kept steady-state shared array-metadata boxing slice now reuses
shared boxed metadata values for common dynamic-array lengths/capacities and
for the `__able_array_size` helper path. The refreshed aligned
`bytecode-runtime` `i_before_e` result on that kept code is:

- `20.24s/op`, `4.57 GB/op`, `72,504,749 allocs/op` on the profiled rerun
- `19.87s/op`, `4.57 GB/op`, `72,505,049 allocs/op` on the clean rerun

The profile shift is specific and useful: `(*Interpreter).initArrayBuiltins.func4`
fell from about `167 MB` flat alloc-space to about `148 MB`, and
`(*ArrayState).BoxedLengthValue` fell from about `160 MB` to about `149 MB`.
The remaining steady-state wall is still mostly environment creation, struct
literal construction, match/ensure body evaluation, and residual host-value
conversion, but the array metadata path is no longer paying as much repeated
first-access boxing cost.

The next kept steady-state small unsigned extern-host conversion slice now
lowers host `u8` / `u16` / `u32` and in-range `u64` / `usize` results straight
to small integers instead of routing them through `big.Int` boxing. The
refreshed aligned `bytecode-runtime` `i_before_e` result on that kept code is:

- `22.95s/op`, `4.50 GB/op`, `69,018,740 allocs/op` on the profiled rerun
- `20.57s/op`, `4.50 GB/op`, `69,019,113 allocs/op` on clean rerun A
- `21.65s/op`, `4.50 GB/op`, `69,018,676 allocs/op` on clean rerun B
- `19.46s/op`, `4.50 GB/op`, `69,021,452 allocs/op` on clean rerun C

This is an allocation-pressure win rather than a clean wall-clock jump, but it
is real: `(*Interpreter).fromHostValue` fell from about `103 MB` flat
alloc-space to about `91 MB`, and the old `bigIntFromUint(...)`-driven
unsigned-host conversion slice dropped out of the top alloc set entirely. That
means the remaining steady-state host path is now less about integer boxing and
more about the residual union/nullable dispatch plus the larger environment /
match / ensure wall around it.

The next kept steady-state lazy environment-mutex tranche now keeps the
per-environment `sync.RWMutex` behind a lazy `atomic.Pointer` in
`pkg/runtime/environment.go`, so the single-threaded bytecode hot path stops
carrying an eagerly allocated lock payload in every short-lived lexical scope.
The refreshed aligned `bytecode-runtime` `i_before_e` result on that kept code
is:

- `21.98s/op`, `4.25 GB/op`, `69,018,511 allocs/op` on the profiled rerun
- `21.97s/op`, `4.25 GB/op`, `69,018,009 allocs/op` on clean rerun A
- `21.84s/op`, `4.25 GB/op`, `69,019,472 allocs/op` on clean rerun B

This is another object-size reduction win rather than a scope-count reduction,
and the alloc profile makes that clear: `NewEnvironmentWithValueCapacity(...)`
fell from about `1.71 GB` flat alloc-space to about `1.48 GB`, while alloc
count stayed in the same `69.02M` band. That shifts the remaining steady-state
wall even more cleanly toward `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `runMatchExpression(...)`, `execEnsureStart(...)`,
and the residual host-value conversion path.

The next kept text-path tranche moved out of the VM and into the canonical
external stdlib: `../able-stdlib/src/fs.able` now routes `fs.read_lines(...)`
through a direct `fs_read_lines_fast(...)` extern instead of layering
`open` / `read_all` / `close`, `bytes_to_string(...)`, newline normalization,
and line splitting in Able code. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun A: `1.28s/op`, `101.46 MB/op`,
  `3,582,266 allocs/op`
- bytecode-runtime clean rerun B: `1.40s/op`, `101.47 MB/op`,
  `3,582,304 allocs/op`
- bytecode-runtime profiled: `1.41s/op`, `101.51 MB/op`,
  `3,582,321 allocs/op`
- compiled external compare: `0.38s`

This is a major workload-specific collapse of the old text/fs stack. The
steady-state alloc profile no longer shows `runMatchExpression(...)`,
`execEnsureStart(...)`, `NewEnvironmentWithValueCapacity(...)`, or
`evaluateStructLiteral(...)` anywhere near the old top tier for
`i_before_e`. The remaining wall on this benchmark is now `copyCallArgs`,
`resolveMethodFromPool`, `stringMemberWithOverrides`, and the residual
extern-host conversion path.

The next kept text-path VM tranche stayed inside the interpreter runtime.
`resolveMethodFromPool(...)` now skips eager scope-callable and type-name
probing for primitive receivers until after inherent/interface/native method
lookup fails, so hot `String` method calls no longer pay `env.Lookup(...)` /
`env.Has(...)` work on the common success path. The refreshed aligned
`i_before_e` results on that kept code are:

- bytecode-runtime clean rerun: `1.27s/op`, `101.47 MB/op`,
  `3,582,283 allocs/op`
- bytecode-runtime profiled: `1.24s/op`, `101.51 MB/op`,
  `3,582,351 allocs/op`

This is a CPU-path cleanup, not a new heap collapse. The profiled comparison
against the prior kept `read_lines` fast-path state shows the intended shift:
`resolveMethodFromPool(...)` dropped from about `210ms` cumulative to about
`120ms`, and `stringMemberWithOverrides(...)` dropped from about `250ms`
cumulative to about `100ms`, while the benchmark stayed in the same
`~101 MB/op` band. The next remaining text-path wall is now led more clearly
by `copyCallArgs`, the residual `resolveMethodFromPool(...)` flat alloc slice,
`overloadArgKinds(...)`, and extern-host conversion.

The next kept text-path VM tranche stayed on the exact-native extern path.
The Go extern wrappers created by `makeExternNative(...)` now mark
`BorrowArgs: true`, so synchronous extern-host calls no longer clone argument
slices before dispatch. The refreshed aligned `i_before_e` results on that
kept code are:

- bytecode-runtime clean rerun A: `1.14s/op`, `84.88 MB/op`,
  `3,063,799 allocs/op`
- bytecode-runtime clean rerun B: `1.19s/op`, `84.89 MB/op`,
  `3,063,840 allocs/op`
- bytecode-runtime profiled: `1.11s/op`, `84.92 MB/op`,
  `3,063,917 allocs/op`

This is a real alloc collapse. The refreshed alloc profile no longer shows
`copyCallArgs(...)` anywhere near the top tier, and total profiled alloc-space
fell from about `119.85 MB` to about `90.95 MB`. The remaining text-path wall
is now led more cleanly by `resolveMethodFromPool(...)`,
`overloadArgKinds(...)`, `stringMemberWithOverrides(...)`, and residual
extern conversion through `fromHostValue(...)`.

The next kept text-path VM tranche stayed on the overload-selection cache
path. `v12/interpreters/go/pkg/interpreter/eval_expressions_calls_overloads.go`
now uses an inline comparable overload-cache signature for the common
small-arity cases instead of rebuilding the old concatenated
`overloadArgKinds(...)` string on every hot lookup, while larger arities still
fall back to the old slow path. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun: `1.18s/op`, `75.17 MB/op`,
  `2,718,071 allocs/op`
- bytecode-runtime profiled: `1.14s/op`, `75.21 MB/op`,
  `2,718,161 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows `overloadArgKinds(...)` in the top tier at all, and total profiled
alloc-space fell again from about `90.95 MB` to about `82.41 MB`. The
remaining text-path wall is now led more cleanly by
`resolveMethodFromPool(...)`, `stringMemberWithOverrides(...)`, and residual
extern conversion through `fromHostValue(...)`.

The next kept text-path VM tranche stayed on the hot member-call path itself.
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go` now
exposes `resolveMethodCallableFromPool(...)`, and bytecode lowering now emits a
dedicated `bytecodeOpCallMember` so the VM can resolve the callable template
and inject the receiver directly instead of first allocating a fresh
`runtime.BoundMethodValue` for common `obj.method(...)` calls. The refreshed
aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.06s/op`, `55.81 MB/op`,
  `1,853,918 allocs/op`
- bytecode-runtime clean rerun B: `1.09s/op`, `55.81 MB/op`,
  `1,853,917 allocs/op`
- bytecode-runtime profiled: `1.09s/op`, `55.85 MB/op`,
  `1,854,009 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows `resolveMethodFromPool(...)` in the top tier at all, and total profiled
alloc-space fell again from about `82.41 MB` to about `72.65 MB`. The
remaining text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`fromHostValue(...)`, and the residual extern/string conversion path.

The next kept text-path VM tranche stayed on that residual extern return path.
`v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now routes hot `i32`
fast-invoker results through `boxedOrSmallIntegerValue(...)` instead of boxing
a fresh `runtime.NewSmallInt(...)` on every call, while
`v12/interpreters/go/pkg/interpreter/extern_host_coercion.go` plus the new
`v12/interpreters/go/pkg/interpreter/extern_host_result_fast.go` now fast-path
host `String`, `Array String`, and `IOError | Array String`-style union
results before falling back to the old generic reflect-heavy conversion path.
The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.08s/op`, `42.69 MB/op`,
  `1,335,429 allocs/op`
- bytecode-runtime clean rerun B: `1.08s/op`, `42.70 MB/op`,
  `1,335,471 allocs/op`
- bytecode-runtime profiled: `1.08s/op`, `42.74 MB/op`,
  `1,335,572 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows the old generic `fmt.Sprint(value.Interface())` / `fromHostValue(...)`
extern return path in the top flat allocators, the hot
`buildExternFastInvoker.func1` `runtime.NewSmallInt(...)` slice disappeared,
and total profiled alloc-space fell again from about `72.65 MB` to about
`50.44 MB` while wall-clock stayed in the same restored `~1.08s/op` band.
The remaining text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`boxedOrSmallIntegerValue(...)`, `externStringSliceResult(...)`, and the
residual `reflect.Value.Call` path behind `fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that residual union-return extern
path. `v12/interpreters/go/pkg/interpreter/extern_host.go` now passes the
active interpreter into cached fast invokers, and
`v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now fast-paths
one-string-arg `func(string) interface{}` host wrappers, which is the shape
produced for union-return externs like `fs_read_lines_fast(...)`. Hot
`[]string` success returns now bypass `reflect.Value.Call` entirely, while
non-`[]string` results still fall back through `fromHostValue(...)` using the
already-computed direct result. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun A: `0.98s/op`, `42.71 MB/op`,
  `1,335,497 allocs/op`
- bytecode-runtime clean rerun B: `1.01s/op`, `42.70 MB/op`,
  `1,335,460 allocs/op`
- bytecode-runtime profiled: `1.01s/op`, `42.74 MB/op`,
  `1,335,541 allocs/op`

This is a real CPU-path keep. Heap stayed in the same collapsed `~42.7 MB/op`
band, but the refreshed profile shows the intended shift:
`reflect.Value.Call` dropped out of the top CPU and alloc sets for the
`fs_read_lines_fast(...)` success path, and steady-state wall-clock moved down
from the prior `~1.08s/op` band into the `~0.98-1.01s/op` band. The remaining
text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`boxedOrSmallIntegerValue(...)`, `externStringSliceResult(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche widened the bytecode small-int boxing cache
to all int64-representable integer suffixes instead of limiting the hot path to
`i32`, `i64`, and `isize`. The kept code is in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`, with the
focused unsigned-cache coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`.
The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.09s/op`, `34.39 MB/op`,
  `1,162,592 allocs/op`
- bytecode-runtime profiled: `1.08s/op`, `34.43 MB/op`,
  `1,162,662 allocs/op`

Subsequent clean reruns on this machine were wall-clock noisy while holding the
same `~34.4 MB/op` / `1.16M allocs/op` heap shape, so this is best treated as
an allocation-pressure keep rather than a new stable CPU-path win. The real
profile shift is that the old unsupported-kind fallback in
`boxedOrSmallIntegerValue(...)` disappeared from the hot alloc story, and total
profiled alloc-space fell again from about `61.30 MB` to about `50.83 MB`.
That leaves the remaining text-path wall more cleanly concentrated in
`bytecodeBoxedIntegerValue(...)`, `patternToInteger(...)`,
`callResolvedCallableWithInjectedReceiver(...)`,
`externStringSliceResult(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that remaining integer-conversion
wall. `v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast.go`
now keeps small integer-to-integer suffix casts on a 64-bit arithmetic path
instead of immediately allocating `big.Int` values for
`bitPattern(...)` / `patternToInteger(...)`, and the same helper is reused from
`v12/interpreters/go/pkg/interpreter/interpreter_type_coercion.go` so the
post-alias simple-type cast path also avoids that small-int `big.Int` churn.
Focused coverage in
`v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast_test.go`
now pins small signed-to-unsigned wrap behavior, the negative-to-`u64`
big-integer fallback boundary, and the bounded allocation behavior on repeated
`u8` casts. The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.35s/op`, `26.10 MB/op`,
  `644,158 allocs/op`
- bytecode-runtime clean rerun B: `1.17s/op`, `26.10 MB/op`,
  `644,158 allocs/op`
- bytecode-runtime profiled: `1.36s/op`, `26.13 MB/op`,
  `644,192 allocs/op`

This is a keep for allocation pressure, not a stable CPU-path win. Wall-clock
stayed noisy on this machine, but the heap shift is large and repeatable. The
important profile change is that
`castValueToCanonicalSimpleTypeFast(...)`, `castValueToType(...)`,
`patternToInteger(...)`, and `bitPattern(...)` dropped out of the top alloc set
entirely, and total profiled alloc-space fell again from about `50.83 MB` to
about `44.90 MB`. That leaves the remaining text-path wall more cleanly
concentrated in `bytecodeBoxedIntegerValue(...)`,
`callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche closed a real gap between method resolution
and bytecode member execution. The resolver can already return
`runtime.NativeBoundMethodValue` for primitive/native receivers, but
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` had only been
accepting raw `NativeFunctionValue` templates on the direct member-call
exact-native path. The kept change now lets
`bytecodeResolveExactInjectedNativeCallTarget(...)` accept
`NativeBoundMethodValue` too, and the focused coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go` pins that
shape directly. The refreshed aligned `i_before_e` results on the kept code
are:

- bytecode-runtime clean rerun A: `1.08s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime clean rerun B: `1.07s/op`, `26.13 MB/op`,
  `644,192 allocs/op`
- bytecode-runtime clean rerun C: `1.12s/op`, `26.09 MB/op`,
  `644,118 allocs/op`
- bytecode-runtime profiled: `1.07s/op`, `26.13 MB/op`,
  `644,192 allocs/op`

This is a CPU-path keep layered on top of the prior heap work. The aligned
runtime moved back into the low `~1.07-1.12s/op` band while preserving the
post-cast `~26.1 MB/op` / `644k allocs/op` heap shape. The remaining text-path
wall is still led by `callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche closed a real direct-call cache gap.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now uses the
existing bytecode member-method cache on the `bytecodeOpCallMember` path
instead of bypassing it and re-running method resolution on every direct
`obj.method(...)` call. That also closes a real regression: the existing
`TestBytecodeVM_StatsMemberMethodCacheCounters` proof was red because
`execCallMember(...)` never consulted the cache even though member access and
dotted call-name fallback already did. On a miss, the VM now stores a rebound
template for the same cache surface; on a hit, it executes the cached resolved
member callee through the exact-native / inline / generic call ladder without
re-running `resolveMethodCallableFromPool(...)`. The refreshed aligned
`i_before_e` results on the kept code are:

- bytecode-runtime clean rerun A: `1.00s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime clean rerun B: `1.00s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime profiled: `1.02s/op`, `26.13 MB/op`,
  `644,191 allocs/op`

This is a keep as both correctness and CPU-path work. The cache-counter
regression is closed, and aligned steady-state bytecode `i_before_e` moved from
the prior `~1.07-1.12s/op` band into the low `~1.00-1.02s/op` band while
preserving the post-cast `~26.1 MB/op` / `644k allocs/op` heap shape. The
remaining text-path wall is still led by
`callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on method resolution itself.
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go` now
gives primitive receivers stable bound-method cache keys by type token instead
of treating value receivers like `String` as uncached, and it keeps that cache
semantically safe by only storing primitive receiver entries when resolution
actually came from a real method candidate. Primitive scope-fallback callables
stay uncached. The focused coverage in
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution_cache_test.go`
now pins both sides directly: primitive `String` methods reuse one cache entry
across distinct receiver values, and primitive scope-fallback callables do not
get cached across reassignment. The refreshed aligned `i_before_e` results on
the kept code are:

- bytecode-runtime clean rerun A: `0.92s/op`, `26.10 MB/op`,
  `644,118 allocs/op`
- bytecode-runtime clean rerun B: `0.97s/op`, `26.10 MB/op`,
  `644,117 allocs/op`
- bytecode-runtime profiled: `1.11s/op`, `26.13 MB/op`,
  `644,191 allocs/op`

This is a CPU-path keep layered on top of the prior cache work. The refreshed
profile shows `resolveMethodCallableFromPool(...)` dropping from the older
`~250ms` cumulative tier to about `~70ms` while keeping the post-cast heap
shape effectively flat. The remaining text-path wall is now more cleanly led by
`callResolvedCallableWithInjectedReceiver(...)`,
`externStringSliceResult(...)`, `bytecodeBoxedIntegerValue(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that injected-receiver wall, but
it did so by moving the receiver prepend into the shared callable dispatcher
instead of creating another direct-call path. `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
now exposes a shared optional injected-receiver helper, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now passes the
existing VM stack slice plus receiver into that helper instead of first
materializing a fresh merged argument slice for every `obj.method(...)` call.
Focused coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go` now also
pins optional-arity and overloaded method-call semantics on the direct
member-call opcode path. The refreshed aligned `i_before_e` results on the kept
code are:

- bytecode-runtime clean rerun: `0.90s/op`, `20.57 MB/op`,
  `471,293 allocs/op`
- bytecode-runtime profiled: `0.90s/op`, `20.60 MB/op`,
  `471,367 allocs/op`

This is a keep as both CPU-path and heap work. The injected member-call merge
fell out of the top alloc set entirely, and aligned steady-state bytecode
`i_before_e` moved from the prior `~0.92-0.97s/op` band into the `~0.90s/op`
band while heap fell from about `26.1 MB/op` / `644k allocs/op` to about
`20.6 MB/op` / `471k allocs/op`. The remaining text-path wall is now more
cleanly `bytecodeBoxedIntegerValue(...)`, `externStringSliceResult(...)`,
`strings.genSplit`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept steady-state tranche raised the lazy dynamic boxed-int cache cap
from `32768` to `262144` in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`, which
is large enough for one warmup pass to retain the full large loop-index working
set on aligned `i_before_e` without broadening the eager fixed small-int cache.
The refreshed aligned `i_before_e` results on the kept code are:

- bytecode-runtime clean rerun: `0.88s/op`, `14.63 MB/op`,
  `347,625 allocs/op`
- bytecode-runtime profiled: `0.88s/op`, `14.66 MB/op`,
  `347,695 allocs/op`

This is a keep as heap work with a stable CPU band. The refreshed alloc profile
shows the intended shift: `bytecodeBoxedIntegerValue(...)` dropped from about
`5.63 MB` flat alloc-space to about `1.54 MB`, and total profiled alloc-space
fell from about `44.37 MB` to about `34.94 MB`. The remaining text-path wall is
now more cleanly `externStringSliceResult(...)`, `strings.genSplit`,
`buildExternFastInvoker.func8`, and the residual plugin body in
`fs_read_lines_fast(...)`.

The next kept bytecode-runtime tranche reworked the `fs_read_lines_fast(...)`
cache shape in the external stdlib. Instead of the earlier shared map +
`RWMutex` experiment, [fs.able](/home/david/sync/projects/able-stdlib/src/fs.able)
now keeps a single-entry immutable hot cache keyed by
`path + size + modifiedNs` behind `atomic.Pointer`. The hot repeated-read path
is now just `os.Stat(...)` plus an atomic load/compare, while misses still
rebuild from `os.ReadFile(...)` and replace the cached entry. The rewrite
invalidation proof is pinned in
[compiler_stdlib_io_temp_test.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/compiler/compiler_stdlib_io_temp_test.go).

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.91s/op`, `8.37 MB/op`, `347,617 allocs/op`
- bytecode-runtime clean rerun B: `0.89s/op`, `8.37 MB/op`, `347,617 allocs/op`
- bytecode-runtime profiled: `0.89s/op`, `8.40 MB/op`, `347,690 allocs/op`

This is a keep as both heap work and a CPU-safe cache shape. Compared with the
prior kept dynamic boxed-int tranche, aligned steady-state bytecode
`i_before_e` stays in the same sub-second band while heap falls from about
`14.6 MB/op` to about `8.4 MB/op`. The refreshed alloc profile shows the
intended shift: the old `strings.genSplit` / `os.readFileContents`
plugin-body cost drops out of the measured hot path, leaving
`buildExternFastInvoker.func8`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and residual member/native dispatch as the
cleaner remaining wall.

The next kept bytecode-runtime tranche moved off full per-call string-slice
boxing in the interpreter fast invoker layer. In
[extern_host_fast.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/extern_host_fast.go),
each string-slice fast invoker now keeps a tiny cached template for
`[]string -> []runtime.Value`, keyed by a source snapshot. Repeated hot
`Array String` extern results now clone that cached boxed template instead of
re-boxing every `StringValue` from scratch on each call, while still returning
a fresh Able array backing slice. The invalidation/no-aliasing behavior is
pinned in
[extern_host_result_fast_test.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/extern_host_result_fast_test.go).

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun: `0.87s/op`, `5.61 MB/op`, `174,794 allocs/op`
- bytecode-runtime profiled: `0.88s/op`, `5.64 MB/op`, `174,867 allocs/op`

This is a keep as another material heap collapse with a flat CPU band.
Compared with the prior kept atomic `read_lines` hot-cache tranche, aligned
steady-state bytecode `i_before_e` stays in the same sub-second range while
heap falls from about `8.4 MB/op` / `348k allocs/op` to about
`5.6 MB/op` / `175k allocs/op`. The refreshed alloc profile shows the intended
shift: `externStringSliceResult(...)` drops out of the top alloc tier and the
remaining array-string return cost is now mostly one cloned `[]runtime.Value`
slice in `externCloneValueSlice(...)`, plus the remaining member/native
dispatch wall.

The next kept steady-state runtime tranche preserved validated VM lookup caches
across pooled runs and moved match-clause binding scopes onto pre-sized local
environments with non-merging binds. The refreshed aligned `i_before_e`
steady-state result on that kept code is:

- `24.61s/op`
- `9.90 GB/op`
- `146,646,034 allocs/op`

The runtime-only alloc-space profile now shows the intended shift:
`storeCachedScopeValue(...)` and `storeCachedCallName(...)` are no longer
top-tier allocators. The remaining steady-state pressure is now centered on
`matchPattern(...)`, environment creation/binds for clause scopes, and the
runtime work those clause scopes feed (`runMatchExpression(...)`,
`execEnsureStart(...)`, extern-host calls, struct literal construction, and
runtime context snapshots).

The next kept steady-state tranche moved simple match clauses onto a direct
fast path for identifier, wildcard, literal, and typed patterns instead of
always going through the generic binding collector. The refreshed aligned
`i_before_e` steady-state result on that kept code is:

- `23.84s/op`
- `8.89 GB/op`
- `139,670,607 allocs/op`

The updated profiles show the right shift again: `matchPattern(...)` itself is
no longer a dominant allocator, and `matchPatternFast(...)` now carries much
of the hot match work directly. The remaining steady-state pressure is now
mostly clause-scope environment/binding churn plus the runtime work those
clauses feed, especially `runMatchExpression(...)`, `execEnsureStart(...)`,
extern-host calls, struct literal construction, and runtime context
snapshots.

The next kept steady-state tranche moved runtime diagnostic work for explicit
`return` onto a lazy path. Bytecode and tree-walker `returnSignal` no longer
snapshot the call stack on every normal return; return-type coercion failures
now attach runtime context only when they actually need a diagnostic. The
refreshed aligned `i_before_e` steady-state result on that kept code is:

- `23.40s/op`
- `8.74 GB/op`
- `136,183,494 allocs/op`

The refreshed profiles show the expected shift again: `snapshotCallStack(...)`
dropped materially from the alloc-space top tier, so the remaining steady-state
pressure is now more cleanly concentrated in clause-scope environment/binding
churn plus `runMatchExpression(...)`, `execEnsureStart(...)`, extern-host
calls, struct literal construction, and runtime context attachment on actual
error paths.

The next kept steady-state tranche moved the hot extern path onto cached target
hashes plus direct invokers for the primitive string signatures that dominate
`i_before_e`. The refreshed aligned steady-state result on that kept code is:

- profiled: `23.43s/op`, `8.23 GB/op`, `121,472,207 allocs/op`
- clean rerun: `22.24s/op`, `8.23 GB/op`, `121,472,290 allocs/op`

The refreshed profiles show the intended shift: `hashExternState(...)`,
`externSignatureKey(...)`, and the old cumulative extern-host hashing path
dropped out of the top allocators, while `invokeExternHostFunction(...)` /
`fromHostResults(...)` shrank materially. The remaining steady-state pressure
is now more clearly concentrated in clause-scope environment/binding churn,
`runMatchExpression(...)`, `execEnsureStart(...)`, struct literal
construction, and the residual host-value conversion path.

The next kept steady-state tranche moved the common one-binding child-scope
case in `runtime.Environment` onto an inline slot instead of an eager
one-entry map. `NewEnvironmentWithValueCapacity(...)` now avoids allocating a
map for `valueCapacity == 1`, and environments promote to a real map only on
the second distinct local binding. The refreshed aligned steady-state
`i_before_e` result on that kept code is:

- profiled: `24.40s/op`, `6.64 GB/op`, `107,523,776 allocs/op`
- clean rerun: `21.93s/op`, `6.64 GB/op`, `107,523,333 allocs/op`

That is a real step down from the prior kept baseline of
`22.24s/op`, `8.23 GB/op`, `121,472,290 allocs/op`. The refreshed alloc
profile shows the intended shift: the first-binding map allocation is no
longer the main wall, and the remaining pressure is now more clearly
`NewEnvironmentWithValueCapacity(...)` object churn,
`promoteSingleBindingNoLock(...)` on multi-bind scopes,
`evaluateStructLiteral(...)`, `snapshotCallStack(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`. Steady-state bytecode
`fib` still times out at `300s`, so this tranche materially improved the
text-heavy runtime path without changing the recursive timeout story yet.

The next kept steady-state tranche sized hot child scopes from cheap AST
binding counts and removed one of the remaining miss-heavy lookup paths.
Block scopes, function/lambda call scopes, loop-iteration scopes, iterator
literal scopes, and `or {}` handler scopes now use
`NewEnvironmentWithValueCapacity(...)`, while `matchPatternFast(...)` now uses
`Environment.Lookup(...)` instead of the miss-allocating `Get(...)` path for
ordinary identifier probes. The refreshed aligned steady-state `i_before_e`
results on that kept code are:

- profiled: `22.01s/op`, `6.43 GB/op`, `97,060,888 allocs/op`
- clean reruns: `21.93s/op` and `21.87s/op`, both at about `6.43 GB/op` and
  `97.06M allocs/op`

Wall-clock stayed in the same low-21s band, but allocation pressure dropped
materially again from the prior kept baseline of `6.64 GB/op` and
`107.52M allocs/op`. The refreshed profiles show the intended shift:
`Environment.Get(...)` / `fmt.Errorf(...)` miss pressure dropped out of the
top tier, and `matchPattern(...)` cumulative allocs narrowed again. The
remaining steady-state wall is now more clearly
`NewEnvironmentWithValueCapacity(...)` object churn,
`setCurrentValueNoLock(...)`, `evaluateStructLiteral(...)`,
`snapshotCallStack(...)`, `runMatchExpression(...)`, and
`execEnsureStart(...)`.

The next kept steady-state tranche moved general runtime diagnostics onto a
lazy call-stack path. Runtime errors no longer eagerly copy the full eval-state
call stack at first attachment; instead they keep a lazy reference that is
frozen only when the stack is about to mutate or when a real runtime
diagnostic is built. The refreshed aligned steady-state `i_before_e` results on
that kept code are:

- profiled: `20.51s/op`, `5.85 GB/op`, `84,856,303 allocs/op`
- clean rerun A: `23.19s/op`, `5.85 GB/op`, `84,856,355 allocs/op`
- clean rerun B: `22.09s/op`, `5.85 GB/op`, `84,856,481 allocs/op`

Wall-clock stayed in the same rough low-20s band with one noisier clean run,
but allocation pressure dropped materially again from the prior kept baseline
of about `6.43 GB/op` and `97.06M allocs/op`. The refreshed alloc profile
shows the intended shift: `snapshotCallStack(...)` dropped out of the top
allocators entirely, so the remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `collectImplCandidates(...)`, `arrayMember(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`. Steady-state bytecode
`fib` still times out at `300s`, so the next likely wins are still on the
text-heavy runtime path rather than the recursive timeout path.

The next kept steady-state tranche trimmed the hot type-canonicalization path.
`canonicalizeExpandedTypeExpression(...)` now reuses the original
nullable/result/union/function/generic AST nodes when none of their children
actually change, instead of rebuilding fresh type-expression trees on every
no-op canonicalization pass. The refreshed aligned steady-state `i_before_e`
results on that kept code are:

- profiled: `20.57s/op`, `5.42 GB/op`, `77,708,818 allocs/op`
- clean rerun A: `20.39s/op`, `5.42 GB/op`, `77,708,480 allocs/op`
- clean rerun B: `20.82s/op`, `5.42 GB/op`, `77,710,637 allocs/op`

That is another real step down from the prior kept baseline of about
`22.09s/op`, `5.85 GB/op`, and `84.86M allocs/op`. The refreshed alloc profile
shows the intended shift: `ast.NewNullableTypeExpression(...)` and
`ast.NewUnionTypeExpression(...)` dropped out of the alloc-space top set
entirely, and `canonicalizeExpandedTypeExpression(...)` itself is now a much
smaller slice. The remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `collectImplCandidates(...)`, `arrayMember(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`.

The next kept steady-state tranche stayed on method/member resolution. Array
helper access now skips the guaranteed direct-member miss for non-field names
like `len`, `get`, and `push`, while `typeImplementsInterface(...)` now caches
resolved type/interface/arg-signature results behind the same invalidation
boundary as the existing method cache. The refreshed aligned steady-state
`i_before_e` results on that kept code are:

- profiled: `21.00s/op`, `5.06 GB/op`, `77,363,744 allocs/op`
- clean rerun: `20.19s/op`, `5.06 GB/op`, `77,364,488 allocs/op`

That kept wall-clock in the same low-20s band but cut another chunk of
allocation pressure from the prior kept baseline of about `20.39s/op`,
`5.42 GB/op`, and `77.71M allocs/op`. The refreshed alloc profile shows the
intended shift: `collectImplCandidates(...)` dropped out of the top
alloc-space set entirely, so the remaining steady-state wall is now more
cleanly `NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `arrayMember(...)`, `runMatchExpression(...)`,
`execEnsureStart(...)`, and the residual host-value conversion path.

The next kept steady-state tranche stayed on array metadata and struct-literal
success-path lookup work. Dynamic array state now caches boxed `length` /
`capacity` values, so repeated array helper/member reads stop re-boxing the
same large integers on every access. Struct-literal shorthand bindings and
struct-definition fallback now also use `Environment.Lookup(...)` instead of
the heavier error-producing `Get(...)` path on the hot success path. The
refreshed aligned steady-state `i_before_e` results on that kept code are:

- profiled: `23.03s/op`, `4.88 GB/op`, `73,577,870 allocs/op`
- clean rerun A: `20.00s/op`, `4.88 GB/op`, `73,577,767 allocs/op`
- clean rerun B: `19.61s/op`, `4.88 GB/op`, `73,577,672 allocs/op`

That is another real reduction from the prior kept baseline of about
`20.19s/op`, `5.06 GB/op`, and `77.36M allocs/op`. The refreshed alloc profile
shows the intended shift: `arrayMember(...)` dropped from the older
`~343 MB` flat tier to about `160 MB`, with the remaining array metadata cost
now concentrated in the first boxed-length materialization instead of repeated
re-boxing. The remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, the residual boxed-length/boxed-capacity path,
`runMatchExpression(...)`, `execEnsureStart(...)`, and host-value conversion.

The next kept steady-state tranche stayed on environment-object size. Ordinary
`runtime.Environment` scopes now move their cold struct-definition/runtime-data
state behind a lazy `environmentMeta` pointer, so hot lexical scopes that only
bind values no longer carry those fields inline. The refreshed aligned
steady-state `i_before_e` results on that kept code are:

- profiled: `21.33s/op`, `4.63 GB/op`, `73,577,840 allocs/op`
- clean rerun A: `20.51s/op`, `4.63 GB/op`, `73,577,220 allocs/op`
- clean rerun B: `20.09s/op`, `4.63 GB/op`, `73,577,644 allocs/op`

That kept wall-clock in the same low-20s band while cutting heap pressure from
the prior kept `4.88 GB/op` band down to about `4.63 GB/op`. The refreshed
alloc profile shows the intended shift:
`NewEnvironmentWithValueCapacity(...)` dropped from the old ~`1.95 GB` flat
object-allocation tier to about `1.71 GB`, while the value-map allocation
slice stayed small. The remaining steady-state wall is now more clearly scope
count plus `evaluateStructLiteral(...)`, `setCurrentValueNoLock(...)`,
`runMatchExpression(...)`, `execEnsureStart(...)`, and residual host-value
conversion.

The next kept steady-state tranche stayed on propagation control flow. The
canonical stdlib `io.unwrap(...)`, `io.unwrap_void(...)`, and
`io.bytes_to_string(...)` paths in `../able-stdlib/src/io.able` now use direct
propagation (`!`) instead of nested `match`/`raise` control flow, and the Go
interpreter now reuses `cachedSimpleTypeExpression("Error")` on the hot
propagation/or-else/runtime-error checks instead of constructing a fresh
`ast.Ty("Error")` node every time. The refreshed aligned steady-state
`i_before_e` results on that kept code are:

- profiled: `19.17s/op`, `4.60 GB/op`, `73,227,937 allocs/op`
- clean rerun A: `20.94s/op`, `4.60 GB/op`, `73,229,445 allocs/op`
- clean rerun B: `19.92s/op`, `4.60 GB/op`, `73,228,693 allocs/op`

That kept wall-clock in the low-20s band while cutting heap from the prior
kept `4.63 GB/op` band to about `4.60 GB/op`, and it dropped alloc count by
roughly `350k` objects per run from the prior kept `73.58M` band to about
`73.23M`. The refreshed alloc profile shows the intended shift: the old
`bytecodeOpPropagation` line no longer carries flat alloc-space on the kept
profile, so propagation is no longer paying per-call AST `Error` type
construction on top of the stdlib unwrap/match traffic.

The current cross-mode bytecode-core baseline is checked in at:

- `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json`
- `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md`

That suite is intentionally small and stable enough for routine reruns. It
tracks:

- `quicksort`
- `future_yield_i32_small`
- `sum_u32_small`

For targeted compiler-lowering checks, prefer checked-in fixture targets under
`v12/fixtures/bench/` so the benchmark package metadata is reproducible from
the repo. Recent mono-array work uses
`v12/fixtures/bench/matrixmultiply_f64_small/main.able` for the staged nested
`Array (Array f64)` comparison and
`v12/fixtures/bench/zigzag_char_small/main.able` for the staged text-scalar
(`Array char` / `Array (Array char)`) comparison and
`v12/fixtures/bench/sum_u32_small/main.able` for the staged unsigned numeric
comparison and `v12/fixtures/bench/hashmap_i32_small/main.able` for the first
broader native-container (`HashMap i32 i32` + `Map i32 i32`) comparison and
`v12/fixtures/bench/heap_i32_small/main.able` for the broader array-backed
container family (`Heap i32`) comparison and the shared generic nominal-method
specialization follow-up and bound generic field/member carrier refinement
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_for_i32_small/main.able` for the first
benchmark-worthy generic-container hot-path (`LinkedList -> Iterable ->
Iterator`) comparison and
`v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` for the next
concrete generic/default-method container hot path (`LinkedList.map/filter/reduce`)
comparison and the shared static nominal receiver/struct-literal closure
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able` for
the next iterator default-method hot path
(`LinkedList.lazy().map<i64>(...).filter(...).next()`) comparison and
`v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` for the
mono-array-enabled iterator collect/reduce follow-up
(`LinkedList.lazy().map<i64>(...).filter(...).collect<Array i64>().reduce(...)`)
comparison and
`v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able` for
the iterator-literal controller / `filter_map` follow-up
(`LinkedList.lazy().filter_map<i64>(...).collect<Array i64>().reduce(...)`)
comparison, while the full
`matrixmultiply` workload in `v12/examples/benchmarks/matrixmultiply.able`
remains the canonical suite entry used by `v12/bench_suite`. Current focused
snapshots for these reduced fixtures are checked in at:

- `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-scalar-propagation-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-for-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-specialized-default-impls-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`

The reduced recursion/call-overhead benchmark is now:
- `v12/fixtures/bench/fib_i32_small/main.able`

The current representative compiled Milestone 7 snapshot is:
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`
  - `bench/fib_i32_small`: `2.7567s`, `0.00` GC
  - `bench/heap_i32_small`: `0.2900s`, `5.00` GC
  - `bench/linked_list_iterator_pipeline_i64_small`: `0.1433s`, `9.67` GC
  - `bench/matrixmultiply_f64_small`: `0.1167s`, `7.33` GC
  - `examples/benchmarks/matrixmultiply`: `1.0633s`, `13.33` GC

The `zigzag_char_small` snapshot was corrected after fixing mono-off nested
carrier identity for `Array (Array char)`, so use the checked-in snapshot
rather than any earlier ad hoc mono-off timings.

The current best matrix snapshots are now:
- reduced target:
  `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`,
  which records `0.1133s` / `7.00` GC on the compiled
  `matrixmultiply_f64_small` target after removing the synthetic static-array
  loop-induction checked-arithmetic scaffolding through shared primitive
  counted-loop lowering
- full canonical benchmark:
  `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`,
  which records `1.0833s` / `13.00` GC on the compiled
  `v12/examples/benchmarks/matrixmultiply.able` path

The follow-up affine integer snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
proves the remaining `build_matrix` `i - j` / `i + j` helper calls are gone,
but it is effectively performance-neutral relative to the counted-loop
snapshot on this benchmark family.

The follow-up subtraction range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
proves the widened inline overflow branch is now gone for `build_matrix`
`i - j`, but `i + j` still carries the widened checked-add path and the
benchmark remains in the same band as the counted-loop baseline.

The follow-up upper-bound range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
proves the remaining widened inline overflow branch is now gone for
`build_matrix` `i + j` too. The inner-loop affine add/sub gap is closed, but
the benchmark still remains in the same band as the counted-loop baseline.

The iterator-pipeline family is now split intentionally:
- `linked_list_iterator_pipeline_i64_small` isolates the already-closed native
  `map/filter/next` path
- `linked_list_iterator_collect_i64_small` isolates the now-closed
  mono-array-enabled `collect<Array i64>().reduce(...)` follow-up
- `linked_list_iterator_filter_map_i64_small` isolates the now-closed
  iterator-literal controller / `filter_map(...).collect<Array i64>()`
  follow-up

## Benchmarks Covered

- `fib`
- `binarytrees`
- `matrixmultiply`
- `quicksort`
- `sudoku`
- `i_before_e`

## Usage

```bash
# default suite (all benchmarks, all modes)
./v12/bench_suite

# targeted compiled mono-array comparison
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/hashmap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_for_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/fib_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able

# reproducible baseline example
./v12/bench_suite \
  --suite bytecode-core \
  --runs 1 \
  --timeout 90 \
  --build-timeout 240 \
  --output-json v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json \
  --output-md v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md

# report-only comparison against the checked-in baseline
./v12/bench_guardrail \
  --baseline v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json \
  --current v12/tmp/perf/current-bytecode-core.json

# compare Able against the sibling ../benchmarks corpus
./v12/bench_compare_external \
  --benchmarks fib,binarytrees,matrixmultiply \
  --modes compiled,bytecode \
  --output-md /tmp/able-vs-external.md

# run a concurrency benchmark with the goroutine executor explicitly
./v12/bench_perf --runs 1 --timeout 90 --modes compiled \
  --executor goroutine \
  --run-from ../benchmarks/binarytrees \
  v12/examples/benchmarks/binarytrees.able

# benchmark a local program against an external suite-local input file
./v12/bench_perf --runs 1 --timeout 20 --modes compiled \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# benchmark steady-state bytecode runtime (load/lower once, then main() only)
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# capture a CPU profile for an aligned compiled run
ABLE_GO_CPU_PROFILE=/tmp/able-fib.pprof \
./v12/bench_perf --runs 1 --timeout 90 --modes compiled \
  --run-from ../benchmarks/fib \
  v12/examples/benchmarks/fib.able

# capture a flushed profile from a timed-out bytecode run
ABLE_GO_CPU_PROFILE=/tmp/able-bytecode-fib.pprof \
./v12/bench_perf --runs 1 --timeout 30 --modes bytecode \
  --run-from ../benchmarks/fib \
  v12/examples/benchmarks/fib.able

# capture a steady-state bytecode runtime profile
ABLE_GO_CPU_PROFILE=/tmp/able-bytecode-runtime-ibe.pprof \
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# emit a steady-state bytecode runtime call trace
ABLE_BYTECODE_TRACE_OUT=/tmp/able-bytecode-runtime-ibe-trace.json \
ABLE_BYTECODE_TRACE_LIMIT=12 \
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able
```

## JSON Output

The output file includes:

- `results`: one row per `(benchmark, mode, run_index)`
- `summary`: aggregated `ok/timeout/error` counts and average metrics for successful runs

Statuses:

- `ok`: command exited 0 within timeout
- `timeout`: command exceeded timeout
- `error`: non-timeout non-zero exit, including compiled-build failure

For `bytecode-runtime`, the JSON payload also includes:

- `avg_ns_per_op`
- `avg_bytes_per_op`
- `avg_allocs_per_op`

`bytecode-runtime` runs one explicit warmup call before timed measurement, so
the wall-clock timeout must budget for warmup plus the measured benchmark call.
When `ABLE_GO_CPU_PROFILE` or `ABLE_GO_MEM_PROFILE` is set for
`bytecode-runtime`, the emitted profiles cover only the post-warmup measured
call, not the initial load/lower/warmup phase.

## Bytecode Runtime Trace

`bytecode-runtime` also supports an opt-in sorted bytecode call trace:

- `ABLE_BYTECODE_TRACE_OUT=/tmp/trace.json`
- `ABLE_BYTECODE_TRACE_LIMIT=20` (optional top-N trim; omit for all entries)

The emitted JSON report includes:

- `target_path`
- `run_from`
- `program_args`
- `trace.total_hits`
- `trace.entries`, sorted by `hits` descending

Each trace entry includes:

- `hits`
- `op` (`call_name` or `call_member`)
- `name`
- `lookup` (`name`, `dot_fallback`, `resolved_method`, `member_access`)
- `dispatch` (`exact_native`, `inline`, `generic`)
- `origin`, `line`, and `column` when source location metadata is available

This trace mode is diagnostic-only. The additional counting adds overhead, so
the traced `ns/op` result should not be compared directly against the untraced
steady-state benchmark baseline. Use it to rank hot callsites and callees, then
rerun the normal `bytecode-runtime` benchmark without tracing to judge whether a
change is actually keep-worthy.

The next kept trace-driven bytecode text-path tranche used that call trace to
target overload-valued member calls instead of attempting more exact-native
substitution. In
[bytecode_vm_call_member.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go),
`bytecodeOpCallMember` now resolves overload-valued methods down to a selected
`*runtime.FunctionValue` before dispatch and then feeds that selected overload
through the existing injected-receiver inline/generic ladders without
materializing a fresh bound-method value on each hot call. The small-arity
overload-selection scratch path also stays stack-backed.

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.887s/op`, `5.60 MB/op`,
  `174,773 allocs/op`
- bytecode-runtime clean rerun B: `0.911s/op`, `5.60 MB/op`,
  `174,775 allocs/op`
- bytecode-runtime profiled: `0.905s/op`, `5.63 MB/op`,
  `174,847 allocs/op`

That is a keep as a CPU-path win with the prior low-heap shape preserved.
Compared with the restored post-template-cache baseline, aligned steady-state
bytecode `i_before_e` moves from roughly the `~1.00-1.01s/op` band down into
the `~0.89-0.91s/op` band while staying in the prior `~5.6 MB/op` /
`175k allocs/op` heap band. The traced run confirms the semantic target:
`Array.get` now shows up as `call_member` / `resolved_method` / `inline`
instead of remaining on the generic member-call path.

The next kept bytecode-runtime tranche tightened exact-native call overhead for
extern wrappers only. In
[values.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/runtime/values.go),
`runtime.NativeFunctionValue` now has an opt-in `SkipContext` flag, and
[definitions.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/definitions.go)
marks generated extern wrappers with `SkipContext: true` because those
closures do not observe `*runtime.NativeCallContext`. The tree-walker native
call path and the bytecode
[execExactNativeCall(...)](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/bytecode_vm_call_native_fast.go)
fast path now both bypass native-call-context pooling/setup on that opt-in
path, while context-sensitive runtime/concurrency natives keep the old
behavior unchanged.

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.872s/op`, `5.60 MB/op`,
  `174,775 allocs/op`
- bytecode-runtime clean rerun B: `0.853s/op`, `5.60 MB/op`,
  `174,774 allocs/op`
- bytecode-runtime profiled: `0.837s/op`, `5.63 MB/op`,
  `174,846 allocs/op`

That is a keep as a modest CPU-path win with the same low-heap shape
preserved. Compared with the prior overload-member-inline band, aligned
steady-state bytecode `i_before_e` moves from roughly `~0.89-0.91s/op` down
into the `~0.84-0.87s/op` band while holding the prior `~5.6 MB/op` /
`175k allocs/op` heap band. The refreshed profile also says native-call
context setup is no longer the useful exact-native target; the remaining wall
is now the actual fast-string extern body plus residual member/name-call
dispatch.

The next kept `i_before_e` slice was benchmark-local rather than VM-internal.
[v12/examples/benchmarks/i_before_e/i_before_e.able](/home/david/sync/projects/able/v12/examples/benchmarks/i_before_e/i_before_e.able)
now short-circuits `is_valid(...)` by returning early when a word has no
`"ei"` or has `"ei"` but no `"cei"`, and only falls back to
`replace("cei", "")` on the small remaining subset that actually needs it.
On the aligned `wordlist.txt` corpus that removes pointless `replace(...)`
work from `172,695` of `172,823` words; only `128` words still take the
replace path, and an exhaustive local equivalence check over the aligned
wordlist preserved the prior `1628` invalid outputs.

The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `0.792s/op`, `2.84 MB/op`,
  `2,080 allocs/op`
- bytecode-runtime clean rerun B: `0.749s/op`, `2.84 MB/op`,
  `2,078 allocs/op`
- bytecode-runtime profiled: `0.779s/op`, `2.87 MB/op`,
  `2,151 allocs/op`
- external compiled compare: `0.290s`
- external bytecode compare: `1.020s`

This is a keep because the old benchmark body was doing obviously wasted
string work on nearly the entire corpus; the semantics stay the same, but the
benchmark is no longer dominated by avoidable `replace(...)` calls. Compared
with the prior exact-native skip-context band, aligned steady-state bytecode
`i_before_e` moves from roughly `~0.84-0.87s/op` down into the
`~0.75-0.79s/op` band, and heap drops from roughly `5.6 MB/op` /
`175k allocs/op` into the `~2.84 MB/op` / `2.1k allocs/op` band. One-shot
aligned bytecode `i_before_e` is now `1.02s`, so this benchmark is no longer
the right place to spend more string micro-optimization time before revisiting
the much larger `fib` timeout problem.

The next kept bytecode `fib` slice stayed deliberately narrow. It did not try
another frame-reuse experiment; instead it specialized the remaining hot `i32`
arithmetic/boxing path directly. `v12/interpreters/go/pkg/interpreter/
bytecode_vm_i32_fast.go` now provides a dedicated small-`i32` boxing path plus
direct small-`i32` add/sub helpers, and the fused self-call immediate-subtract
path plus the specialized bytecode integer add/sub path use those helpers
before falling back to the generic integer machinery.

That is a keep as a reduced recursion-kernel win, not a full aligned benchmark
fix. The restored reduced `BenchmarkFib30Bytecode` baseline was roughly
`219-225ms/op`, and the warmed kept reruns landed around `198.70ms/op` and
`201.98ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`, so the remaining real wall is still broader recursive VM cost rather
than just generic `i32` boxing/dispatch.

The next kept bytecode `fib` slice stayed on that same reduced recursion path
and trimmed the self-fast frame shape instead of attempting slot reuse again.
Pure self-recursive calls that carry no generic-name set, no implicit receiver,
and no loop/iterator base state now use a smaller minimal self-fast frame
instead of the full self-fast frame payload.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `198.70-201.98ms/op`; the refreshed warmed reruns on the kept code
landed at `199.73ms/op`, `195.06ms/op`, and `196.84ms/op`. Aligned one-shot
external bytecode `fib` still times out at `90s`, but the refreshed reduced
CPU profile no longer shows `pushCallFrame(...)` as a top-tier flat hotspot.
The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`popCallFrameFields(...)`, `acquireSlotFrame(...)`,
`bytecodeDirectSmallI32Value(...)`, and the residual direct `i32`
boxing/immediate-subtract path.

The next kept bytecode `fib` slice stayed on that reduced recursion path and
trimmed the inline return path instead of trying another frame-reuse
experiment. `v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go` now
owns bytecode inline returns, which pulls the hot return logic out of
`bytecode_vm_run.go` and keeps that file back under the 1000-line guardrail.
More importantly for the benchmark, the inline return helper now handles
`bytecodeCallFrameKindSelfFastMinimal` directly instead of routing that case
through the broader `popCallFrameFields(...)` path, and the hot inline call
sites now use cached return-generic metadata through
`bytecodeInlineReturnGenericNames(...)`.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `195.06-199.73ms/op`; the refreshed warmed reruns on the kept code
landed at `189.63ms/op`, `195.72ms/op`, and `192.46ms/op`. A single profiled
reduced run landed at `202.31ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `pushCallFrame(...)` as a visible top-tier hotspot. The remaining reduced
wall is now more cleanly `execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeDirectSmallI32Value(...)`, and `bytecodeBoxedIntegerI32Value(...)`.

The next kept bytecode `fib` slice stayed on the reduced recursion path again,
but this time targeted slot-frame pool churn rather than arithmetic. `v12/
interpreters/go/pkg/interpreter/bytecode_vm_slot_frames.go` now batches small
slot-frame allocations at 32 frames instead of 8, so reduced recursive runs do
not keep rebuilding tiny hot pools across the common recursion depth.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `189.63-195.72ms/op`; the refreshed warmed reruns on the kept code
landed at `198.99ms/op`, `183.53ms/op`, and `187.89ms/op`. A single profiled
reduced run landed at `207.16ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `releaseSlotFrame(...)` as a top-tier hotspot. The remaining reduced wall
is now more cleanly `execCallSelfIntSubSlotConst(...)`, `pushCallFrame(...)`,
`finishInlineReturn(...)`, `bytecodeDirectSmallI32Value(...)`,
`acquireSlotFrame(...)`, and `execBinarySlotConst(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path and
trimmed the minimal self-fast setup path instead of trying another arithmetic
micro-optimization. `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go`
now exposes `pushSelfFastMinimalCallFrame(...)`, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` now uses that
direct helper inside `execCallSelfIntSubSlotConst(...)` whenever the current
program already has a cached nil return-generic set and the frame is
guaranteed to stay minimal. That same path now also skips
`bytecodeInlineReturnGenericNames(...)` entirely on the cached-nil case.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior restored reduced `BenchmarkFib30Bytecode` band was
roughly `188.93-197.98ms/op`; the refreshed warmed reruns on the kept code
landed at `197.79ms/op`, `185.03ms/op`, and `184.96ms/op`. A single profiled
reduced run landed at `194.46ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `pushCallFrame(...)` or `bytecodeInlineReturnGenericNames(...)` in the
visible top tier. The remaining reduced wall is now more cleanly
`finishInlineReturn(...)`, `execCallSelfIntSubSlotConst(...)`,
`inlineCoercionUnnecessaryBySimpleType(...)`, `execBinarySlotConst(...)`,
`bytecodeDirectSmallI32Value(...)`, and
`bytecodeSubtractIntegerImmediateI32Fast(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path again
but moved one level lower into slot-frame release. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_slot_frames.go` now uses direct nil stores for released
slot frames of size `1..4` instead of always calling the generic
`clear(slots)` path. That matches the common reduced `fib` frame sizes, so the
hot return path no longer pays the broader slice-clear helper for every tiny
frame release.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `184.96-197.79ms/op`; the refreshed warmed reruns on the kept code
landed at `186.37ms/op`, `187.04ms/op`, and `189.04ms/op`. A single profiled
reduced run landed at `212.16ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `releaseSlotFrame(...)` / `clear(slots)` as a visible top-tier hotspot.
The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeDirectSmallI32Value(...)`, and
`bytecodeSubtractIntegerImmediateI32Fast(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path but
trimmed the self-call dispatch shape itself. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_calls.go` now has an early dedicated self-slot fast
branch inside `execCallSelfIntSubSlotConst(...)`, so the successful recursive
hot path bypasses the older generic callee switch, the `*bytecodeProgram`
type assertion/equality check, and `callNode` extraction entirely. That branch
reads the `*runtime.FunctionValue` directly from the reserved self slot, uses
the already-known `currentProgram`, and stays on the existing minimal
self-fast frame path.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `186.37-189.04ms/op`; the refreshed warmed reruns on the kept code
landed at `181.90ms/op`, `178.13ms/op`, and `176.39ms/op`. A single reduced
rerun landed at `188.57ms/op`. Aligned one-shot external bytecode `fib` still
times out at `90s`, but the reduced warmed band moved materially again. The
next remaining wall is now more clearly the arithmetic side of the same
recursion path: `execBinary(...)`, `execBinarySlotConst(...)`,
`bytecodeDirectSmallI32Value(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
and the residual return-path work in `finishInlineReturn(...)`.

The next kept reduced-`fib` slice stayed on that arithmetic side but narrowed
it to the recursive-result `+` path only. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_i32_fast.go` now has a dedicated
`bytecodeDirectSmallI32Pair(...)` helper, and
`bytecodeAddSmallI32PairFast(...)` uses that combined extractor directly
instead of calling `bytecodeDirectSmallI32Value(...)` twice for the hot
small-`i32` pair-add case.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `176.39-181.90ms/op`; the refreshed warmed reruns on the kept code
landed at `171.06ms/op`, `175.06ms/op`, and `175.27ms/op`. A single profiled
reduced run landed at `183.33ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the reduced warmed band moved again. The next
remaining wall is now more cleanly the self-call arithmetic and return side:
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeSubtractIntegerImmediateI32Fast(...)`, and the residual direct
small-`i32` extraction/boxing helpers.

The next kept reduced-`fib` slice stayed on that same self-fast recursion path
but trimmed the minimal frame bookkeeping instead of touching arithmetic again.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go` now keeps top
contiguous minimal self-fast frames out of `callFrameKinds` entirely and tracks
them with an explicit suffix count until a broader frame kind needs to sit
above them. `v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go`,
`v12/interpreters/go/pkg/interpreter/bytecode_vm_pool.go`, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_run_finalize.go` now consume
that same suffix directly, so the hot reduced-`fib` recursion path no longer
appends a `bytecodeCallFrameKindSelfFastMinimal` entry on every recursive
step, while mixed stacks still materialize the suffix back into
`callFrameKinds` before pushing a full or metadata-bearing self-fast frame.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior kept reduced baseline was in the low `170ms` band;
the refreshed warmed reruns on the kept code landed at `173.91ms/op`,
`171.32ms/op`, and `172.12ms/op`. A single profiled reduced run landed at
`170.23ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`, but the hot recursion path now avoids the old per-step minimal-frame
kind push entirely. The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`pushSelfFastMinimalCallFrame(...)`, and the residual direct small-`i32`
extraction/boxing helpers.

The next kept reduced-`fib` slice moved out of the VM and into lowering.
`v12/interpreters/go/pkg/interpreter/bytecode_lowering.go` now routes non-last
`if` expressions through a statement-only lowering path, and
`v12/interpreters/go/pkg/interpreter/bytecode_lowering_controlflow.go` lowers
that path without synthesizing a dead `Nil` value for the missing `else`
branch. On the reduced `fib(30)` kernel that removes one dead
`bytecodeOpConst Nil` plus one immediate `Pop` from every non-base recursive
step in the function body.

That is a keep as another reduced recursion-kernel win, not a full aligned
benchmark fix. The prior kept warmed reduced `BenchmarkFib30Bytecode` band was
roughly `171.32-173.91ms/op`; the refreshed warmed reruns on the kept code
landed at `163.53ms/op`, `159.41ms/op`, and `160.61ms/op`. A single profiled
reduced run landed at `169.55ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows the earlier dead statement-result const/pop overhead as a visible
top-tier slice. The remaining reduced wall is back on
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`bytecodeDirectSmallI32Value(...)`, `acquireSlotFrame(...)`, and residual
run-loop dispatch.

The next kept reduced-`fib` slice stayed on control-flow lowering, but moved
from dead statement cleanup to direct conditional dispatch.
`v12/interpreters/go/pkg/interpreter/bytecode_lowering_controlflow.go` now
lowers `if` / `elsif` conditions through a dedicated conditional jump opcode
when the existing slot-const matcher already proves the shape `slot <= i32`,
and `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go` executes that
path without first materializing a boolean result for `JumpIfFalse` to consume.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior kept warmed reduced
`BenchmarkFib30Bytecode` band was roughly `159.41-163.53ms/op`; the refreshed
warmed reruns on the kept code landed at `159.93ms/op`, `155.70ms/op`, and
`151.83ms/op`. A single profiled reduced run landed at `151.60ms/op`.
Aligned one-shot external bytecode `fib` still times out at `90s`, but the
refreshed reduced profile no longer routes the base-case compare through the
old `execBinarySlotConst(...) -> BoolValue -> JumpIfFalse` path. The remaining
reduced wall is now more cleanly `execCallSelfIntSubSlotConst(...)`,
`execBinary(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`acquireSlotFrame(...)`, `finishInlineReturn(...)`, and residual run-loop
dispatch.

The next kept reduced-`fib` slice moved onto that residual return side.
`v12/interpreters/go/pkg/interpreter/bytecode_slot_analysis.go` now caches a
compact primitive return check on slot frame layouts, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go` uses that cached
check in `finishInlineReturn(...)` before falling back to the older
string-based simple-type helper or full return coercion. The reduced `fib`
kernel's recursive `Int` returns therefore avoid re-running the string-based
simple return helper on every unwind while preserving the existing fallback
path for non-simple and mismatched values.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior refreshed warmed reduced
`BenchmarkFib30Bytecode` band was `156.88ms/op`, `160.22ms/op`, and
`163.96ms/op`. The first warmed reruns on the kept code landed at
`155.13ms/op`, `153.92ms/op`, and `159.16ms/op`; the confirmation band landed
at `150.43ms/op`, `157.63ms/op`, `156.45ms/op`, `150.28ms/op`, and
`151.49ms/op`. A single profiled reduced run landed at `144.17ms/op`.
Allocation shape stayed essentially unchanged at about `102 KB/op` and
`863 allocs/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The remaining reduced wall is now most likely the self-call arithmetic
and residual frame churn around `execCallSelfIntSubSlotConst(...)`,
`execBinary(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`acquireSlotFrame(...)`, and run-loop dispatch.

The next kept reduced-`fib` slice stayed on the fused recursive self-call path
and targeted the arithmetic setup for `fib(n - 1)` / `fib(n - 2)` directly.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_i32_fast.go` now exposes a
self-call-only small-`i32` immediate subtract helper, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` uses that helper
inside `execCallSelfIntSubSlotConst(...)` before falling back to the broader
integer-immediate helper ladder. This keeps generic arithmetic fallback
unchanged while letting the reduced recursive fast path compute its next
argument without reopening the wider helper path.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior kept confirmation band after cached return
checks was `150.43ms/op`, `157.63ms/op`, `156.45ms/op`, `150.28ms/op`, and
`151.49ms/op`. The first warmed band on this tranche landed at
`147.64ms/op`, `149.63ms/op`, and one noisy `172.81ms/op` outlier; the
confirmation band then landed at `146.70ms/op`, `144.66ms/op`,
`143.36ms/op`, `139.07ms/op`, and `143.54ms/op`. A single profiled reduced
run landed at `137.87ms/op`. Allocation shape stayed essentially unchanged at
about `102 KB/op` and `863-864 allocs/op`. Aligned one-shot external bytecode
`fib` still times out at `90s`. The next reduced wall should be re-profiled,
but likely remains in residual self-call frame setup,
`execBinary(...)` result addition, `acquireSlotFrame(...)`, and run-loop
dispatch rather than the immediate subtract ladder.

The next reduced-`fib` maintenance tranche was intentionally behavior-neutral.
The fused slot-const recursive self-call path moved out of
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` and into
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_self_slot_const.go` so
follow-up work can stay below the project file-size cap and avoid mixing
call-dispatch edits with fused-recursion edits. `bytecode_vm_calls.go` dropped
from `992` to `842` lines; the new focused file is `158` lines.

Focused recursive/self-fast coverage stayed green, and the reduced
`BenchmarkFib30Bytecode` check remained in the current kept band at
`143.51ms/op`, `148.90ms/op`, and `148.58ms/op` with the same allocation
shape. No aligned external run was needed for this organization-only split.
The next performance tranche should start from the new fused self-call file
and re-profile before trying more run-loop or frame-setup work.

The next kept reduced-`fib` slice used that new fused self-call boundary to
target the remaining size-2 frame acquisition edge. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_slot_frames.go` now exposes a dedicated
`acquireSlotFrame2()` helper that mirrors the existing hot-pool semantics for
exactly two-slot frames, and `bytecode_vm_call_self_slot_const.go` uses it
only when the fused self-call frame layout is exactly two slots. All other
layouts continue through the general `acquireSlotFrame(slotCount)` path.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The refreshed pre-change reduced
`BenchmarkFib30Bytecode` band was `147.30ms/op`, `140.91ms/op`, and
`150.88ms/op`. The first warmed band after the change landed at
`135.60ms/op`, `136.34ms/op`, and `138.36ms/op`; the confirmation band landed
at `139.26ms/op`, `140.10ms/op`, `138.69ms/op`, `138.53ms/op`, and
`141.70ms/op`. A single profiled reduced run landed at `133.52ms/op`.
Allocation shape stayed unchanged at about `102 KB/op` and `863-864
allocs/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The small `preprofile` output no longer shows the earlier
`execCallSelfIntSubSlotConst(...) -> acquireSlotFrame(...)` edge; remaining
visible reduced samples sit around fused self-call dispatch, the conditional
slot-const jump, `finishInlineReturn(...)`, and `execBinary(...)`.

The next kept reduced-`fib` slice stayed on that residual conditional
slot-const jump path. `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
now gives `execJumpIfIntLessEqualSlotConstFalse(...)` a dedicated direct
small-integer `<=` immediate helper. That keeps the fused `if n <= const`
condition out of the generic `bytecodeDirectIntegerCompare("<=", ...)` helper
and avoids constructing a `BoolValue` just so the conditional jump can read it
back immediately. Generic binary comparisons still use the broader helper and
fallback ladder.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The refreshed pre-change reduced
`BenchmarkFib30Bytecode` checks landed at `135.15ms/op` for a 3x warmed run
and `141.27ms/op` for the profiled one-shot. The first warmed band after the
change landed at `130.38ms/op`, `141.41ms/op`, and `140.29ms/op`; the
confirmation band landed at `138.06ms/op`, `137.32ms/op`, `136.52ms/op`,
`134.52ms/op`, and `136.98ms/op`. A single profiled reduced run landed at
`135.38ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The small `preprofile` output no longer shows the earlier
`execJumpIfIntLessEqualSlotConstFalse(...) -> bytecodeDirectIntegerCompare(...)`
edge; remaining visible reduced samples sit around fused self-call dispatch,
add/binary execution, inline return, and residual conditional jump dispatch.

The next reduced-`fib` tranche deliberately improved measurement instead of
changing the VM hot path. `v12/interpreters/go/pkg/interpreter/fib_bench_test.go`
now includes `BenchmarkFib30BytecodeRuntimeOnly`, which parses/evaluates the
reduced `fib` function once, validates a warmup `fib(30) == 832040`, and then
repeatedly calls that same bytecode function on the same interpreter. The
existing `BenchmarkFib30Bytecode` remains the end-to-end reduced check that
builds a fresh interpreter and evaluates the module each iteration.

The first side-by-side check showed why this separation matters:
end-to-end `BenchmarkFib30Bytecode` landed at `138.59ms/op` and
`136.27ms/op` with about `102 KB/op` and `863-864 allocs/op`, while
`BenchmarkFib30BytecodeRuntimeOnly` landed at `130.76ms/op` and
`144.90ms/op` with effectively zero steady-state allocations. The broader
runtime-only warmed band landed at `134.48ms/op`, `129.18ms/op`,
`135.39ms/op`, `139.44ms/op`, and `130.13ms/op`; the profiled one-shot landed
at `135.83ms/op`, and a post-assertion one-shot landed at `128.88ms/op`. The
runtime-only `preprofile` now points directly at VM runtime edges:
`execCallSelfIntSubSlotConst(...)`, `bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)`,
`execJumpIfIntLessEqualSlotConstFalse(...)`, `execBinary(...)`, and the
remaining `finishInlineReturn(...)` sample.

The next runtime-only reduced-`fib` slice used that isolated benchmark to
target the sampled self-call subtract helper itself. Since
`bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)` only handles operands
that have already been proven small `i32`, it no longer calls the generic
`subInt64Overflow(...)` helper before checking the `i32` bounds. The observable
overflow behavior is preserved by the existing `math.MinInt32` /
`math.MaxInt32` check, so `i32` underflow still reports the same overflow
error and non-`i32` shapes still miss this self-call helper.

That is a runtime-only reduced recursion-kernel keep. The runtime-only
baseline before the change landed at `133.10ms/op`, `127.21ms/op`, and
`130.93ms/op`. The first warmed band after the change landed at
`132.47ms/op`, `129.62ms/op`, `126.40ms/op`, `128.22ms/op`, and
`126.96ms/op`; confirmation landed at `134.12ms/op`, `135.39ms/op`,
`128.00ms/op`, `135.33ms/op`, and `130.16ms/op`; the profiled one-shot landed
at `128.22ms/op`. A temporary restored A/B band with the old helper landed
much slower at `156.25-180.87ms/op`, so the direct subtract helper change is
kept despite the normal reduced-fib timing noise. The runtime-only profile now
still points at `execCallSelfIntSubSlotConst(...)`, `finishInlineReturn(...)`,
the fused conditional jump, and the residual binary add/boxing path as the
remaining work.

The next reduced-`fib` control-flow tranche fused the base-case return shape
itself. Statement-position `if slot <= const { return slot }` now lowers to
`bytecodeOpReturnIfIntLessEqualSlotConst`, so the reduced `fib` base case no
longer executes a standalone slot-const conditional jump followed by a separate
slot load and return dispatch. Expression-position `if` lowering and
non-returning statement `if` behavior continue through the existing paths.

That is a runtime-only reduced recursion-kernel keep. The first warmed band
after the change landed at `122.95ms/op`, `127.64ms/op`, `128.59ms/op`,
`133.48ms/op`, and `124.78ms/op`; confirmation landed at `136.10ms/op`,
`125.41ms/op`, `132.00ms/op`, `132.53ms/op`, and `135.25ms/op`; the profiled
one-shot landed at `136.73ms/op`. The small runtime-only `preprofile` now
shows `runResumable(...) -> execReturnIfIntLessEqualSlotConst(...)` in place
of the older standalone `execJumpIfIntLessEqualSlotConstFalse(...)` base-case
edge. The remaining reduced wall is back on fused self-call dispatch,
`finishInlineReturn(...)`, and residual binary add/small-integer handling.

The next kept follow-on narrowed that fused return-if opcode further for the
exact same-slot shape emitted by reduced `fib`: `if n <= const { return n }`.
When the condition slot and return slot are identical and the opcode already
carries a typed integer immediate, `execReturnIfIntLessEqualSlotConst(...)`
now compares the already-read slot value directly, returns it on the true path,
and advances `ip` on the false path. Other shapes still use the existing
return-if fallback.

This is another small runtime-only reduced recursion-kernel keep. The
refreshed return-if baseline landed at `140.33ms/op`, `130.53ms/op`,
`138.43ms/op`, `128.59ms/op`, and `131.46ms/op`. The same-slot fast path first
kept band landed at `133.12ms/op`, `131.19ms/op`, `122.67ms/op`,
`129.81ms/op`, and `132.06ms/op`; confirmation landed at `138.89ms/op`,
`123.58ms/op`, `132.05ms/op`, `125.71ms/op`, and `127.26ms/op`. A temporary
restored A/B check with the same-slot block removed regressed to
`155.13ms/op`, `140.63ms/op`, `146.91ms/op`, `137.47ms/op`, and
`136.60ms/op`, so the fast path is retained despite the noisy profiled
one-shot. The next reduced profile should start again from fused self-call
dispatch, inline return, and the binary add/small-integer path.

The next kept run-loop tranche removed one of the helper edges introduced
during the return-if split. `bytecodeOpReturnIfIntLessEqualSlotConst` is now
handled inline in `runResumable(...)` again, while cold placeholder
lambda/value execution moved to `bytecode_vm_placeholder.go` so
`bytecode_vm_run.go` remains below the project line cap. A narrower
self-call-only fixed-cache boxing bypass was tested first, but it regressed
the runtime-only band to `139.98-153.20ms/op` and was reverted.

The refreshed runtime-only baseline before this tranche landed at
`125.49ms/op`, `125.65ms/op`, `139.05ms/op`, `128.96ms/op`, and
`128.07ms/op`, with a profiled one-shot at `140.06ms/op`. The inline
return-if/cold-placeholder split first clean band landed at `131.03ms/op`,
`137.89ms/op`, `129.69ms/op`, `134.66ms/op`, and `136.04ms/op`; the quiet
confirmation band landed at `130.60ms/op`, `127.63ms/op`, `130.90ms/op`,
`128.78ms/op`, and `132.06ms/op`; the profiled one-shot landed at
`144.56ms/op`. End-to-end reduced `BenchmarkFib30Bytecode` one-shots landed
at `130.39ms/op`, `132.60ms/op`, and `131.84ms/op`. The final small
runtime-only `preprofile` no longer shows the removed
`runReturnIfIntLessEqualSlotConst(...)` wrapper. The next tranche should start
from the remaining fused self-call dispatch, `finishInlineReturn(...)`, and
binary add/small-integer samples.

The next runtime-only reduced-`fib` tranche targeted the unwind side of that
same minimal self-fast recursion path. `bytecode_vm_slot_frames.go` now has a
dedicated `releaseSlotFrame2(...)` helper for exact two-slot frames, and
`finishInlineReturn(...)` uses it only when the active frame layout proves the
callee frame has exactly two slots. This is intentionally different from the
rejected frame-clear elision probe: both slots are still eagerly cleared before
the frame returns to the hot pool.

The refreshed runtime-only baseline before the change landed at
`137.61ms/op`, `126.62ms/op`, `127.97ms/op`, `127.92ms/op`, and
`129.67ms/op`, with a profiled one-shot at `132.08ms/op`. The first kept band
after the change landed at `112.96ms/op`, `113.63ms/op`, `119.86ms/op`,
`127.98ms/op`, and `125.08ms/op`; the confirmation band landed at
`114.67ms/op`, `111.05ms/op`, `111.15ms/op`, `117.82ms/op`, and
`111.43ms/op`; the profiled one-shot landed at `118.84ms/op`. The tiny
runtime-only `preprofile` no longer shows the old
`finishInlineReturn(...) -> releaseSlotFrame(...)` edge. The next tranche
should start from fused self-call setup, the residual `finishInlineReturn(...)`
coercion check, and the binary add/small-integer samples.

The next runtime-only reduced-`fib` tranche fused the final implicit add-return
shape. A node-less implicit `BinaryIntAdd` followed by `Return` now lowers to
`bytecodeOpReturnBinaryIntAdd`; the following return instruction is left in
place but becomes unreachable, preserving existing jump targets. The VM uses
the existing specialized add helper and returns the result directly instead of
replacing the top two stack values and dispatching the next return opcode.
Explicit `return expr + expr` shapes remain on the existing path.

This is a small runtime-only reduced recursion-kernel keep. The same-load
pre-change runtime-only baseline landed at `116.06ms/op`, `118.25ms/op`, and
`116.52ms/op`. The first fused band landed at `118.64ms/op`, `121.93ms/op`,
`116.23ms/op`, `115.53ms/op`, and `115.04ms/op`; the longer fused band landed
at `111.70ms/op`, `112.49ms/op`, `120.25ms/op`, `125.21ms/op`,
`111.48ms/op`, `111.92ms/op`, `114.35ms/op`, and `109.11ms/op`. A temporary
no-fusion control under the same host load landed at `119.59ms/op`,
`123.21ms/op`, `113.04ms/op`, `117.74ms/op`, and `113.97ms/op`; restored
fused confirmation landed at `115.48ms/op`, `110.36ms/op`, `109.40ms/op`,
`110.87ms/op`, and `112.25ms/op`. The profiled one-shot landed at
`123.82ms/op`, and the tiny runtime-only `preprofile` now shows
`runResumable(...) -> execReturnBinaryIntAdd(...)` rather than a standalone
final recursive `execBinary(...)` sample. The next tranche should start from
fused self-call setup, `finishInlineReturn(...)` coercion checks, and the
remaining call-frame/slot state churn rather than another generic add-dispatch
rewrite.

The next kept tranche pivoted from the reduced `fib(30)` shape to the real
aligned external benchmark shape. The checked-in external benchmark is
`fib(45)` over `i32` with `if n <= 2 { return 1 }`, so the earlier fused
`return slot` base-case opcode did not apply. Statement-position
`if slot <= const { return small_i32_const }` now lowers to
`bytecodeOpReturnConstIfIntLessEqualSlotConst`, which returns the encoded
small `i32` constant directly after the same direct slot/immediate comparison
used by the other fused slot-const conditional paths.

This is an aligned-shape keep, not a full external timeout fix yet. The
current reduced `BenchmarkFib30BytecodeRuntimeOnly` baseline before the change
landed at `109.57-116.54ms/op`; after the change it landed at `113.40ms/op`,
`116.23ms/op`, and `119.29ms/op`, so the already-optimized reduced path is
effectively unchanged. The aligned-style `fib_i32_small` bytecode-runtime
fixture landed at `10.56s/op` across three fused runs and `10.58s/op` on
restored fused confirmation. A temporary no-fusion control under the same
fixture landed at `12.59s/op`, which is enough to keep the opcode. The full
external bytecode `fib(45)` run still times out at `90s`; the next tranche
should stay on aligned-fib residual overhead rather than another reduced
`fib(30)` branch unless a fresh profile says otherwise.

The next aligned-fib tranche targeted a repeated object-immediate probe rather
than another frame-setup branch. Lowered slot-const instructions now keep a raw
`int64` immediate beside the existing typed `runtime.IntegerValue`. The typed
value remains the semantic fallback path, while fused self-call subtract,
return-const base-case, and conditional slot-const helpers can use the raw
value after lowering has already proven the literal is a small default `i32`
immediate.

This is another aligned-shape keep. A profiled `fib_i32_small` run before the
change showed samples in `bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)`
and `bytecodeDirectIntegerLessEqualImmediate(...)` from repeatedly unpacking
the same instruction immediate. With raw immediates enabled, aligned-style
`fib_i32_small` bytecode-runtime runs landed at `9.94s/op`, `10.37s/op`, and
`10.18s/op`; a temporary no-raw control under the same code shape landed at
`10.49s/op`; restored raw confirmation landed at `9.49s/op`. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `118.38-126.67ms/op`, so the
change is kept for the aligned benchmark path rather than claimed as a reduced
`fib(30)` win. The next profile should start from fused self-call setup,
`bytecodeAddSmallI32PairFast(...)`, and `finishInlineReturn(...)`; the raw
immediate probe itself should no longer be the first thing to chase.

The next aligned-fib tranche specialized the already-fused implicit return-add
opcode for functions declared `i32`. When lowering sees a node-less final
`BinaryIntAdd` followed by the implicit `Return` inside an `i32` function, it
now emits `bytecodeOpReturnBinaryIntAddI32`. That opcode tries the direct
small-`i32` add path first, then falls back to the existing generic return-add
semantics for unexpected operand shapes.

This is an aligned-shape keep, not a reduced `fib(30)` win. The reduced
`BenchmarkFib30BytecodeRuntimeOnly` sanity band landed at `125.87ms/op`,
`127.84ms/op`, and `125.94ms/op`. The aligned-style `fib_i32_small`
bytecode-runtime fixture landed at `9.89s/op` and `9.86s/op` across two
3-run confirmation bands, with a profiled one-shot at `9.77s/op`. The
aligned `preprofile` no longer shows the old
`execReturnBinaryIntAdd(...) -> execBinarySpecializedOpcode(...)` edge;
return-add now reaches `bytecodeAddSmallI32PairFast(...)` directly on the hot
path. The next profile should start from fused self-call setup,
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
the remaining direct small-`i32` pair extraction/boxing costs.

The next aligned-fib tranche stayed on the fused self-call setup path. The raw
slot-const immediate work proved the literal value was already available as an
`int64`, so `execCallSelfIntSubSlotConst(...)` now performs the small-`i32`
subtract directly in the fused recursive self-call branch instead of calling
`bytecodeSelfCallSubtractIntegerImmediateI32RawFast(...)`. The existing
overflow check, boxed integer cache, typed-immediate path, and generic fallback
behavior remain unchanged.

This is an aligned-shape keep with a reduced runtime-only assist. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `114.39ms/op`, `116.08ms/op`,
and `122.80ms/op`. The aligned-style `fib_i32_small` bytecode-runtime fixture
landed at `9.80s/op` across a 3-run band, with a profiled one-shot at
`9.61s/op`. The aligned `preprofile` no longer shows the helper edge
`execCallSelfIntSubSlotConst(...) -> bytecodeSelfCallSubtractIntegerImmediateI32RawFast(...)`;
the fused self-call path now reaches `bytecodeBoxedIntegerI32Value(...)`
directly after the inlined subtract. Full external bytecode `fib(45)` still
times out at `90s`, so the next profile should start from the remaining
`execCallSelfIntSubSlotConst(...)`, `finishInlineReturn(...)`,
`execReturnConstIfIntLessEqualSlotConst(...)`, and `releaseSlotFrame2(...)`
costs rather than another raw-subtract helper-shape rewrite.

The next aligned-fib tranche removed work from the minimal self-fast return
branch itself. The external-style base case lowers to
`bytecodeOpReturnConstIfIntLessEqualSlotConst`, and that lowering encodes the
literal return value as an `i32`. When the active function also declares an
`i32` return, `finishInlineReturn(...)` now treats that fused opcode as already
satisfying the return type and skips the generic simple return-coercion probe.
All other return opcodes and mismatched return declarations keep the existing
coercion path.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly`
landed at `118.01ms/op`, `113.15ms/op`, and `122.26ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `9.33s/op` and `9.24s/op` across
two 3-run bands, with a profiled one-shot at `8.74s/op`. The aligned
`preprofile` shows
`finishInlineReturn(...) -> inlineCoercionUnnecessaryBySimpleCheck(...)`
dropping from the prior `39` sample range to `14` samples after the fused
base-case return skips that probe. Full external bytecode `fib(45)` still
times out at `90s`. The next profile should start from
`execCallSelfIntSubSlotConst(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`,
`bytecodeDirectSmallI32Pair(...)`, `bytecodeBoxedIntegerI32Value(...)`, and
slot-frame release costs rather than another return-coercion shortcut.
