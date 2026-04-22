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
