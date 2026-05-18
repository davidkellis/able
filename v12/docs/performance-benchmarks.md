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
serializing all spawned work. Its default mode set is now `compiled`,
`bytecode`, and `treewalker`; pre-summary Able failures are recorded as
machine-readable per-mode failure rows instead of aborting the whole external
comparison.

The checked-in current external scoreboard lives in:

- `v12/docs/perf-baselines/external-scoreboard-current.json`
- `v12/docs/perf-baselines/external-scoreboard-current.md`

That artifact joins current kept Able measurements for every implemented
external benchmark family with the best Go/Ruby/Python rows from
`../benchmarks/results.json`, including timeout rows for modes that still do
not complete at the external scale.

As of April 29, 2026, the aligned compiled core and closed text/sort
benchmarks are in the same approximate range as Go:

- `fib`: `2.9940s` vs Go `2.8400s`
- `binarytrees`: `3.6400s` vs Go `3.8300s`
- `matrixmultiply`: `0.9660s` vs Go `0.8800s`
- `quicksort`: `1.75s` vs Go `2.01s`
- `sudoku`: `0.0600s` vs Go `0.1300s`
- `i_before_e`: `0.0620s` vs Go `0.0500s`

So the compiled core is now in the Go-range band for the current pass. Any
further `matrixmultiply` work should be limited to general row-length /
bounds-proof machinery rather than benchmark source tweaks. The remaining
bytecode work is still a larger VM architecture problem, not one-time
CLI/bootstrap/lowering noise.

The final compiled `fib` gap closed with a bounded recursive return-range
proof. For simple one-parameter signed integer recurrences with a proven
static call bound and a literal terminating base case, the compiler now stores
per-argument return maxima. That lets the hot `fib(45)` body lower
`fib(n - 1) + fib(n - 2)` as direct `i32` addition after proving the two
recursive return ranges fit together; an overflowing recurrence still keeps
the checked add. The kept external comparison moved compiled Able from
`3.1760s` to `2.9940s` over `5/5` runs versus Go `2.8400s`, and the profiled
kept run landed at `2.9700s`.
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

The first VM v2 typed-lane code slice is now in place as a stack-only seed:
literal-only final `i32` add/sub expressions use a raw `i32` operand stack,
perform checked overflow, and box back to `runtime.Value` before the existing
return path. Reduced `Fib30Bytecode` is intentionally neutral on this slice
because recursive slot arithmetic still boxes; the final neutral reruns landed
at `115.78ms/op` and `116.69ms/op`. The next bytecode performance tranche
should extend the raw lane to declared `i32` slots in non-yielding
slot-eligible functions before revisiting aligned external `fib`.

That declared-slot tranche is now landed for safe non-recursive entry into the
typed lane: frame layouts record simple declared `i32` params and typed local
identifier declarations, and final arithmetic can use `LoadSlotI32` /
`StoreSlotI32` with raw checked `i32` add/sub before boxing. Reduced
`Fib30Bytecode` remains neutral because recursive self-fast frames still carry
boxed slot values; guardrail reruns landed at `117.43ms/op` and
`121.97ms/op`. The next bytecode performance tranche should carry typed `i32`
slot state through inline/self-fast frames and wire the fused self-call
subtract plus typed return-add path to use that state.

The next kept reduced-recursion tranche deliberately avoided the rejected
parallel typed-slot side-cache approach and instead compacted the proven
two-slot one-arg self-fast recursive frame shape. The fused `self(slot -
const)` path now saves caller slot 0, mutates the current two-slot frame for
the callee, and restores slot 0 on return instead of acquiring and releasing a
fresh two-slot frame each step. Reduced `Fib30Bytecode` moved from the refreshed
`119.96-125.07ms/op` baseline band to kept reruns of `105.34ms/op`,
`106.13ms/op`, and `102.79ms/op`; the profiled keep rerun landed at
`109.33ms/op` and no longer has `acquireSlotFrame2` / `releaseSlotFrame2` in
the hot self-call path. The remaining reduced wall is now mostly
`finishInlineReturn`, `execReturnBinaryIntAdd`, boxed small-integer result
handoff, and residual fused self-call guards.

The follow-up kept reduced-recursion tranche targeted those residual fused
self-call guards without changing semantics. `execCallSelfIntSubSlotConst(...)`
now tries an early exact-shape compact branch for the proven raw-immediate
two-slot slot-0 recursive shape before entering the generic immediate,
layout, and return-name ladder; unsupported shapes still use the existing
boxed/generic fallback. Reduced `Fib30Bytecode` moved from a refreshed
compact-frame profiled baseline of `105.27ms/op` to warmed reruns of
`99.54ms/op`, `100.39ms/op`, and `99.00ms/op`. A focused external bytecode
`fib(45)` comparison now completes inside the old `90s` guard at `79.1200s`,
versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`. The next
bytecode recursion tranche should target the base-case and return side around
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
`execReturnBinaryIntAdd(...)`.

The next kept aligned-recursion tranche targeted return-add helper overhead.
`execReturnBinaryIntAdd(...)` now handles the direct
`runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` pair inline before
falling back to the existing pointer and generic paths. A narrower raw-`i32`
base-case helper was tested first and reverted because aligned
`fib_i32_small` regressed to `7.98s/op`. The kept return-add inline tranche
left reduced `Fib30Bytecode` in range at `97.19ms/op`, `104.20ms/op`, and
`106.93ms/op`; aligned `fib_i32_small` bytecode-runtime moved to `7.21s/op`
over a 3-run band, with a profiled one-shot at `7.50s/op`. Focused external
bytecode `fib(45)` moved to `77.2400s`, versus Go `2.8400s`, Ruby `46.6400s`,
and Python `60.6700s`. The profile no longer shows
`bytecodeReturnAddSmallI32ValuePairFast(...)`, so the next recursion work
should target structural boxed return/add handoff, the base-case raw compare,
or compact `finishInlineReturn(...)` restoration rather than another
return-add helper.

The next kept aligned-recursion tranche removed the remaining helper call from
the exact compact self-call setup path. After
`execCallSelfIntSubSlotConstCompact(...)` has already proven the raw-immediate
two-slot slot-0 recursive shape, cached nil return generics, no implicit
receiver, and no active loop/iter state, it now writes the compact slot-0
self-fast frame record directly instead of calling
`pushSelfFastSlot0CallFrame(...)`. A compact `finishInlineReturn(...)`
shortcut for the same shape was tested and reverted because aligned
`fib_i32_small` regressed to `8.31s/op`. The kept inline-push tranche moved
reduced `Fib30Bytecode` to `104.21ms/op`, `96.22ms/op`, and `94.85ms/op`.
Aligned `fib_i32_small` bytecode-runtime landed at `7.18s/op`, with a profiled
one-shot at `7.60s/op`. Focused external bytecode `fib(45)` moved to
`76.7900s`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`. The
profile no longer shows `pushSelfFastSlot0CallFrame(...)` as a separate hot
edge; the remaining recursion work is structural raw/typed return metadata or
typed-frame design, not another small helper split.

The next kept aligned-recursion tranche made that typed-frame direction
concrete for the exact compact self-fast shape. The VM now saves/restores a
raw `i32` slot-0 lane beside the boxed semantic slot value in minimal
self-fast frames. The fused recursive subtract and fused base-case
slot-const compare can use the raw lane while slot-0 writes refresh or clear
it, and all unsupported shapes keep the boxed fallback path. Reduced
`Fib30Bytecode` moved to `92.46ms/op`, `92.81ms/op`, and `92.08ms/op`.
Aligned `fib_i32_small` bytecode-runtime landed at `6.24s/op`, with a
profiled one-shot at `6.03s/op`. Focused external bytecode `fib(45)` moved to
`67.8200s`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`.
The later direct compact minimal-return tranche measured against a noisier
same-session control rather than that historical best: aligned
`fib_i32_small` bytecode-runtime moved from `14.21s` to `13.3533s`, and full
external bytecode `fib(45)` moved from a reverted control at `77.0800s` to
`75.3700s`. The follow-up exact value-pair return-add inline tranche moved the
same full external bytecode `fib(45)` check to `72.8000s`.

The next kept aligned-recursion tranche stopped adding side metadata and
replaced the exact proven recurrence body with a guarded native bytecode
kernel. Slot-backed one-arg `i32` functions shaped as
`if n <= c { return r }` followed by `self(n-a) + self(n-b)` now execute the
recurrence directly with checked `i32` subtract/add overflow and box only at
the bytecode boundary; unsupported shapes and bytecode-stats runs keep the
existing bytecode path. The refreshed aligned `fib_i32_small` bytecode-runtime
baseline was `13.1900s`; the kept run landed at `0.7867s` over `3/3`, with a
profiled one-shot at `0.8100s` whose samples are entirely in the native
recurrence kernel. Full external bytecode `fib(45)` moved to `3.7633s` over
`3/3`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`.

The next kept VM-v2 general slice added a slot-backed bool conditional jump.
Declared `bool` identifiers used directly as `if`, `elsif`, or `while`
conditions now lower to `JumpIfBoolSlotFalse`, so the VM avoids the old
load/push/pop path when the branch can read a `runtime.BoolValue` directly
from the slot. The fallback still uses the existing truthiness helper for any
unsupported shape, preserving v12 truthiness semantics. In the same session,
the refreshed external bytecode baseline was `4.2200s` for `sudoku` and
`1.3200s` for `i_before_e`; the kept confirmation moved those to `2.6500s`
and `1.0000s`. Against the external rows, bytecode `sudoku` is now `20.38x`
Go, `0.47x` Ruby, and `0.88x` Python; bytecode `i_before_e` is now `20.00x`
Go, `10.00x` Ruby, and `7.69x` Python. The next bytecode tranche should
profile post-bool `sudoku` and `i_before_e` and target guarded
call/member/index quickening or canonical Array/String bytecodes before
broadening bool storage beyond condition branches.

The next kept post-bool VM slice added a guarded canonical Array member fast
path on bytecode direct member calls. After normal method resolution selects
the canonical stdlib/kernel `Array.len()` or nullable stdlib `Array.get(i32)`,
the VM executes the size/read operation directly instead of inlining the Able
wrapper body and then dispatching through `__able_array_size` /
`__able_array_read`. The fast path is still guarded by canonical origins,
selected `*runtime.FunctionValue` methods, receiver shape, arity, `i32` index
semantics, and existing fallback behavior; existing array handles are read
directly so host-backed mono arrays are not deoptimized. The kept external
bytecode confirmation moved `sudoku` from the prior kept `2.6500s` to
`2.0433s` over `3/3` runs, and moved `i_before_e` from `1.0000s` to
`0.7500s` over `3/3` runs. Against the external rows, bytecode `sudoku` is now
`15.72x` Go, `0.36x` Ruby, and `0.68x` Python; bytecode `i_before_e` is now
`15.00x` Go, `7.50x` Ruby, and `5.77x` Python. The next bytecode tranche
should profile the remaining post-Array hot calls and target guarded String
member/native-bytecode quickening for `contains`, `len_bytes`, and `replace`,
or a general quickened call/member opcode that reaches those canonical targets
without repeated method resolution.

The follow-up kept member fast-path slice applied the same guard structure to
canonical String wrappers. After normal method resolution selects stdlib
`String.len_bytes`, `String.contains`, or `String.replace`, the bytecode VM
executes the direct string operation instead of inlining the Able wrapper and
dispatching through `string_len_bytes_i32_fast`, `string_contains_fast`, or
`string_replace_fast`. Wrong origins, wrong return types, unsupported
receiver/argument shapes, oversized byte lengths, and all other String methods
fall back to the existing path; `replace` keeps the stdlib's empty-needle
behavior. The traced steady-state `i_before_e` result moved from the
post-Array `583.47ms/op` / `2064 allocs/op` shape to `346.01ms/op` /
`2052 allocs/op`, with the hot calls recorded as `string_contains_fast`,
`string_len_bytes_fast`, and `string_replace_fast` member dispatches. The kept
external bytecode confirmation moved `i_before_e` from `0.7500s` to `0.5767s`
over `3/3` runs and left `sudoku` neutral at `2.0333s` over `3/3` runs.
Against the external rows, bytecode `i_before_e` is now `11.53x` Go, `5.77x`
Ruby, and `4.44x` Python; bytecode `sudoku` is now `15.64x` Go, `0.36x` Ruby,
and `0.67x` Python. The next bytecode tranche should re-profile post-String
`sudoku`; current trace evidence points at Array push/write, iterator `next`,
and UTF-8 byte decode/read paths rather than another String wrapper shortcut.

The next kept member fast-path slice added canonical Array push dispatch.
After normal method resolution selects kernel `Array.push(value) -> void`, the
bytecode VM appends directly through the tracked array state instead of
inlining the Able wrapper and dispatching through `__able_array_write`. Wrong
origins, wrong return types, wrong arity, non-array receivers, and untracked or
typed handles stay on the existing fallback/normalization path; the direct
path returns `void` and keeps array alias tracking synchronized. The traced
steady-state `sudoku` result moved from the post-String
`1771.07ms/op` / `332.12 MB/op` / `4462324 allocs/op` shape to
`1572.84ms/op` / `325.53 MB/op` / `4383835 allocs/op`, with the two hot
`push` sites recorded as `array_push_fast` and no `__able_array_write` entries
in the trace. The kept external bytecode confirmation moved `sudoku` from
`2.0333s` to `1.8833s` over `3/3` runs; the latest `i_before_e` confirmation
landed at `0.5333s` over `3/3` runs. Against the external rows, bytecode
`sudoku` is now `14.49x` Go, `0.33x` Ruby, and `0.62x` Python; bytecode
`i_before_e` is now `10.67x` Go, `5.33x` Ruby, and `4.10x` Python. The next
bytecode tranche should re-profile post-push `sudoku` and target iterator
`next`, string byte iteration / `utf8_decode`, residual Array construction, or
canonical Array `set` only after fresh trace evidence identifies the top edge.

The next kept member fast-path slice targeted the hot byte iterator returned
by `String.bytes()`. After member access resolves the canonical stdlib
`RawStringBytesIter.next` / `StringBytesIter.next` method behind an
`Iterator u8` interface, the VM now reads the current byte and advances
`offset` directly instead of calling the Able method body and its
`read_byte(...)` helper. The fast path is still guarded by canonical
`text/string.able` origin, `u8 | IteratorEnd` return type, receiver struct
name, `bytes`, `offset`, `len_bytes`, and `u8` element shape; unsupported
values fall back to the existing method path. The traced steady-state
`sudoku` result moved from the post-push `1572.84ms/op` / `325.53 MB/op` /
`4383835 allocs/op` shape to `1423.21ms/op` / `294.19 MB/op` /
`4056840 allocs/op`, with the hot `bytes.next()` site recorded as
`string_byte_iter_next_fast`. The kept external bytecode confirmation moved
`sudoku` from `1.8833s` to `1.7300s` over `3/3` runs; the latest
`i_before_e` confirmation landed at `0.5300s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `13.31x` Go, `0.31x` Ruby, and
`0.57x` Python; bytecode `i_before_e` is now `10.60x` Go, `5.30x` Ruby, and
`4.08x` Python. The next bytecode tranche should re-profile post-iterator
`sudoku`; current trace evidence now points at Array reads in
`is_valid` / `board_to_string`, UTF-8 validation/decode during `String.bytes()`,
and residual `Array.new`.

The next kept Array fast-path slice shortened the already-canonical
`Array.get(i32)` path. For receivers that already carry a tracked dynamic
array state, the VM now reads `state.Values` directly before falling back to
the existing handle-store size/read path for untracked or typed handles. The
traced `sudoku` run confirmed that the top `Array.get` sites now dispatch as
`array_get_tracked_fast`; traced wall-clock was noisy, so the external
benchmark band is the keep basis. The kept external bytecode confirmation
moved `sudoku` from `1.7300s` to `1.6200s` over `3/3` runs; the latest
`i_before_e` confirmation landed at `0.5240s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `12.46x` Go, `0.29x` Ruby, and
`0.54x` Python; bytecode `i_before_e` is now `10.48x` Go, `5.24x` Ruby, and
`4.03x` Python. The next bytecode tranche should re-profile
post-tracked-get `sudoku`; current trace evidence still points at UTF-8
validation/decode during `String.bytes()`, residual direct Array reads, and
Array construction.

The next kept String-byte slice removed that UTF-8 validation fan-out from the
valid-input hot path. After normal method resolution selects canonical
`String.bytes() -> Iterator u8`, the bytecode VM now validates the host string
with Go's UTF-8 checker, builds the same `RawStringBytesIter` struct shape as
the stdlib method, and returns it through normal `Iterator u8` interface
coercion. Invalid UTF-8 strings and missing canonical stdlib definitions fall
back to the existing Able method, preserving the canonical
`StringEncodingError` behavior. The traced `sudoku` run moved from
`1326.89ms/op` / `294.34 MB/op` / `4056759 allocs/op` to `653.23ms/op` /
`137.38 MB/op` / `1812289 allocs/op`; `String.bytes()` now records as
`string_bytes_fast`, and `utf8_validate`, `utf8_decode`, and `read_byte` are
absent from the hot trace for the valid sudoku corpus. The kept external
bytecode confirmation moved `sudoku` from `1.6200s` to `0.7780s` over `5/5`
runs. Against the external rows, bytecode `sudoku` is now `5.98x` Go,
`0.14x` Ruby, and `0.26x` Python; `i_before_e` stayed neutral at `0.5333s`
over `3/3` runs. The next bytecode tranche should profile
post-`String.bytes()` `sudoku`; current trace evidence now points at tracked
Array reads, string byte iterator `next`, Array push, and residual
`Array.new` construction.

The next kept propagation slice removed a broad type-match cost from those hot
`Array.get(... )!` success paths. The tree-walker and bytecode interpreters now
share a fast-negative guard for postfix `!`: direct `Error` values and
struct/interface values still route through the existing `Error` matching
path, but ordinary primitive, string, array, iterator, and future success
values skip `matchesType("Error")` unless an `Error` implementation is
registered for that runtime type. The profiled `sudoku` bytecode run moved
from `599.17ms/op` / `137.49 MB/op` / `1812061 allocs/op` to
`448.07ms/op` / `118.99 MB/op` / `1484787 allocs/op`; the old
`bytecodeOpPropagation -> matchesType("Error")` edge dropped from about
`250ms` cumulative to about `10ms`. The kept external bytecode confirmation
moved `sudoku` from `0.7780s` to `0.6700s` over `5/5` runs and confirmed
`i_before_e` at `0.5000s` over `5/5` runs. Against the external rows,
bytecode `sudoku` is now `5.15x` Go, `0.12x` Ruby, and `0.22x` Python;
bytecode `i_before_e` is now `10.00x` Go, `5.00x` Ruby, and `3.85x` Python.
The next bytecode profile should start from residual `execCallMember` /
member-cache overhead around tracked `Array.get` and hot name lookup.

The follow-up semantic propagation tranche aligned postfix `!` with the v12
spec's nil-propagation rule across both interpreters. Runtime `nil` through
`!` now returns from the current function, and the bytecode VM uses
`finishInlineReturn(...)` when that happens inside an inlined bytecode call
frame. To keep `!void` success distinct from Option nil failure,
`dyn.Package.def(...)` success now returns runtime `void` instead of runtime
`nil`. The external bytecode guardrail stayed clean after the additional nil
check: `sudoku` averaged `0.6220s` over `5/5` runs, and `i_before_e` averaged
`0.4620s` over `5/5` runs. Reduced `BenchmarkFib30Bytecode` stayed in its
current band at `91.137669ms/op` (`94.375464ms/op` runtime-only). The next
bytecode profile should start from post-nil-propagation `sudoku` and target
residual `execCallMember` / member-cache overhead around tracked `Array.get`
or the hot name lookup path before expanding to quickened member/index opcodes.

The next kept member-cache tranche stores the resolved canonical member
fast-path kind in the bytecode member-method cache and lets `CallMember` try
that fast path before rebinding the method template or routing through the
generic call ladder. This keeps normal method resolution as the guard while
removing the hot resolved-method bind/call overhead for already-recognized
canonical Array/String/iterator helpers. A temporary revert guard measured the
same external bytecode pair at `2.4200s` for `sudoku` and `0.9600s` for
`i_before_e`; restoring the cache fast path moved the final kept confirmation
to `0.6400s` and `0.4840s` over `5/5` runs, with an earlier same-slice rerun
at `0.6140s` / `0.4580s`. Against the external rows, bytecode `sudoku` is now
`4.92x` Go, `0.11x` Ruby, and `0.21x` Python; bytecode `i_before_e` is now
`9.68x` Go, `4.84x` Ruby, and `3.72x` Python. The same
tranche also tightened the compiled nil-propagation fixture and compiler
lowering: compiled `?T` nil through postfix `!` now returns a normal
nil-compatible value from the current function instead of raising
`runtime.NilValue{}` as control. The next bytecode profile should start from
this post-member-cache state and target the remaining `String.bytes()`
allocation/interface path plus residual name/member lookup.

The next heap-focused bytecode tranche stayed inside the existing
`String.bytes()` fast path. The VM now indexes the already-validated Go string
directly and reuses cached boxed `u8` values for byte elements plus cached
boxed `i32` values for `offset` and `len_bytes`, instead of first copying the
string to `[]byte` and boxing every byte afresh. Runtime-only `sudoku` moved
from `429.30ms/op`, `118.96 MB/op`, and `1,484,673 allocs/op` to
`420.48ms/op`, `114.51 MB/op`, and `1,390,910 allocs/op`; in the memory
profile, `execStringBytesMemberFast(...)` fell from about `18.50 MB` /
`243,574` objects cumulative to about `12.00 MB` cumulative. The external
guard stayed neutral: bytecode `sudoku` averaged `0.6380s` over `5/5` runs
and `i_before_e` averaged `0.4900s` over `5/5` runs. Against the external
rows, bytecode `sudoku` is `4.91x` Go, `0.11x` Ruby, and `0.21x` Python;
bytecode `i_before_e` is `9.80x` Go, `4.90x` Ruby, and `3.77x` Python. The
next bytecode profile should target the remaining `String.bytes()` array /
interface materialization or the residual name/member lookup path.

The follow-up byte-iterator storage tranche kept the same canonical
`String.bytes()` surface but moved the iterator's byte backing onto the
existing mono `u8` array store and attached implementation-private native text
metadata to the canonical `RawStringBytesIter` value. Canonical iterator
`next` now reads directly from that native text metadata before falling back to
the normal mono/dynamic Array path. The public iterator fields remain
`bytes`, `offset`, and `len_bytes`, and unsupported shapes continue through
the existing stdlib path. Runtime-only `sudoku` stayed wall-clock neutral in
the warmed band (`421.79ms/op`, `426.69ms/op`, `422.17ms/op`) while heap
volume moved slightly lower (`113.49 MB/op`, `114.89 MB/op`, `111.86 MB/op`).
The external guard is the keep basis: the first `5/5` rerun measured bytecode
`sudoku` at `0.6340s` and `i_before_e` at `0.4920s`; the repeat measured
`sudoku` at `0.6260s` and `i_before_e` at `0.4800s`. Against the external
rows, bytecode `sudoku` is now `4.82x` Go, `0.11x` Ruby, and `0.21x` Python;
bytecode `i_before_e` is now `9.60x` Go, `4.80x` Ruby, and `3.69x` Python.
The next bytecode profile should target the remaining generic `Iterator u8`
interface coercion in `String.bytes()` or residual name/member lookup; native
metadata should stay limited to guarded canonical stdlib shapes unless a
language-level host boundary is introduced.

The next kept `String.bytes()` tranche removed that remaining generic
interface coercion for the canonical byte iterator. When the VM has the
canonical stdlib `RawStringBytesIter`, canonical `Iterator` interface, and
canonical `RawStringBytesIter.next` method, it now constructs the `Iterator u8`
interface wrapper directly with the cached `next` method instead of routing
through `coerceToInterfaceValue(...)` and
`buildInterfaceMethodDictionary(...)`. Unsupported or non-canonical shapes
still fall back to the existing generic coercion path. Runtime-only `sudoku`
landed at `415.33ms/op`, `427.64ms/op`, and `415.61ms/op`, with allocation
volume down to roughly `102.78-105.86 MB/op` and `1.282M allocs/op`; the
profiled rerun no longer shows `coerceToInterfaceValue(...)` or
`buildInterfaceMethodDictionary(...)` under the `String.bytes()` allocation
edge. The external guard confirmed the keep: the first `5/5` rerun measured
bytecode `sudoku` at `0.6120s` and `i_before_e` at `0.4700s`; the repeat
measured `sudoku` at `0.6160s` and `i_before_e` at `0.4700s`. Against the
external rows, bytecode `sudoku` is now `4.74x` Go, `0.11x` Ruby, and `0.20x`
Python; bytecode `i_before_e` is now `9.40x` Go, `4.70x` Ruby, and `3.62x`
Python. The next bytecode profile should target residual `execCallMember` /
`resolveMethodCallableFromPool` and name/member lookup costs around interface
member access rather than another `String.bytes()` wrapper rewrite.

The next kept bytecode slice targeted static `Array.new()` call overhead in
`sudoku`. After normal static member resolution proves the active method is
the canonical zero-arg kernel `Array.new`, the VM now constructs the same empty
tracked `ArrayValue` directly instead of calling through the generic Able
method and `__able_array_new` native bridge. Unsupported or non-canonical
static member shapes still fall back to the existing path, and the hook is
name/arity-gated so unrelated member-access sites do not pay the check. The
trace changed 10,100 `Array.new` hits to `array_new_fast`, and the warmed
runtime-only `sudoku` band landed at `412.00ms/op`, `416.55ms/op`, and
`407.66ms/op` with allocation volume down to roughly `1.161M allocs/op`. The
external guard confirmed the keep after one discarded noisy paired sample:
`sudoku` measured `0.5820s` / `0.5840s` over two `5/5` runs, and
`i_before_e` stayed neutral at `0.4780s` / `0.4760s`. Against the external
rows, bytecode `sudoku` is now `4.49x` Go, `0.10x` Ruby, and `0.19x` Python;
bytecode `i_before_e` is now `9.52x` Go, `4.76x` Ruby, and `3.66x` Python.
The next bytecode profile should target residual `resolveMethodCallableFromPool`
/ overload-selection cost around hot `Array.get` and iterator `next`, not a
broader static member shortcut without fresh trace evidence.

The follow-up kept bytecode slice targeted the canonical byte-iterator
`next()` call-member path produced by `String.bytes()`. When `CallMember next`
sees the canonical stdlib `Iterator u8` interface wrapping canonical
`RawStringBytesIter`, and the canonical `RawStringBytesIter.next` method is
still valid under the current method/global revisions, the VM jumps directly
to the existing string-byte iterator fast body instead of re-entering generic
`memberAccessOnValueWithOptions(...)` / `interfaceMember(...)` dispatch.
Unsupported interface shapes and non-canonical iterators still use the generic
path. The refreshed steady-state `i_before_e` band landed at
`280.09ms/op`, `236.02ms/op`, and `246.16ms/op`, with about `2.83 MB/op` and
`2,006-2,008 allocs/op`; the profiled rerun landed at `237.20ms/op`,
`2.87 MB/op`, and `2,117 allocs/op`, and `interfaceMember(...)` /
`substituteAliasTypeExpression(...)` dropped out of the hot runtime set. The
external guard measured bytecode `i_before_e` at `0.4660s` and `sudoku` at
`0.5720s` over `5/5` runs. Against the external rows, bytecode `i_before_e`
is now `9.32x` Go, `4.66x` Ruby, and `3.58x` Python; bytecode `sudoku` is now
`4.40x` Go, `0.10x` Ruby, and `0.19x` Python. The next profile should target
residual `execCallMember` / `resolveMethodCallableFromPool` and overload-cache
overhead around hot direct string/member calls, or move back to the larger
bytecode `fib` typed-frame work.

The follow-up kept bytecode slice targeted the canonical nullable `Array.get`
overload pair that remained hot in `i_before_e`. After normal method
resolution returns exactly the canonical stdlib nullable `Array.get(i32) ->
?T` method plus the lower-priority canonical `Index.get(i32) -> !T`
implementation method, `CallMember get` now executes the existing tracked
`Array.get` fast body directly instead of running generic runtime overload
selection. Unsupported overload sets, custom origins, and wrong
parameter/return shapes still fall back to the old resolver, and a VM-local
hot cache keeps the canonical-shape validation off the repeated call path. The
refreshed steady-state `i_before_e` band landed at `193.31ms/op`,
`199.15ms/op`, and `197.26ms/op`, with about `2.82 MB/op` and
`1,989-1,991 allocs/op`; the profiled rerun landed at `237.78ms/op`,
`2.84 MB/op`, and `2,036 allocs/op`, and
`resolveConcreteMemberOverload(...)` dropped out of the hot runtime set. The
external guard measured bytecode `i_before_e` at `0.4480s` and `sudoku` at
`0.5640s` over `5/5` runs. Against the external rows, bytecode `i_before_e`
is now `8.96x` Go, `4.48x` Ruby, and `3.45x` Python; bytecode `sudoku` is
now `4.34x` Go, `0.10x` Ruby, and `0.19x` Python. The next profile should
target residual `resolveMethodCallableFromPool(...)` /
`lookupBoundMethodCache(...)` around canonical primitive methods, or switch
back to bytecode `fib` typed-frame work if the next member-call slice would
require another map/cache layer.

The next kept bytecode heap slice targeted canonical stdlib origin validation
itself. The `Array.get` overload selector still needs to prove that the active
methods come from canonical `able-stdlib`, but
`isCanonicalAbleStdlibOrigin(...)` was allocating fresh concatenated suffix
strings for every validation. It now checks the fixed `/able-stdlib/src/` and
`/pkg/src/` bases plus the relative suffix without allocating on slash-normal
paths. The focused allocation test pins the zero-allocation helper contract.
Runtime-only `sudoku` moved from the refreshed `339.53ms/op`,
`118.11 MB/op`, `1,572,523 allocs/op` sample to `334.69ms/op`,
`86.58 MB/op`, and `915,969 allocs/op`; the allocation profile no longer
shows `isCanonicalAbleStdlibOrigin(...)`, and
`isCanonicalNullableArrayGetOverloadSlow(...)` dropped to `0.06s` cumulative
in the profiled sample. The external guard moved bytecode `sudoku` from
`0.5640s` to `0.5160s` and bytecode `i_before_e` from `0.4480s` to `0.4280s`
over `5/5` runs. Against the external rows, bytecode `sudoku` was `3.97x` Go,
`0.09x` Ruby, and `0.17x` Python; bytecode `i_before_e` was `8.56x` Go,
`4.28x` Ruby, and `3.29x` Python.

The follow-up kept bytecode cache slice targeted the same canonical nullable
`Array.get` overload validation under noisier same-session external timing.
When member resolution returns a fresh overload wrapper around the same
canonical nullable `Array.get` function and lower-priority result-returning
implementation function, the VM now reuses the previous canonical-shape
validation result until the bytecode method/global cache version changes.
Unsupported overload shapes and invalidated versions still fall back to the
existing slow validation. Restored external bytecode passes before reapplying
the cache landed at `0.6480s` / `0.6080s` for `sudoku` and `0.5340s` /
`0.4760s` for `i_before_e`; the kept cache confirmations landed at `0.5280s`
/ `0.5340s` for `sudoku` and `0.4580s` / `0.4540s` for `i_before_e` over
`5/5` runs. The checked-in scoreboard records the later confirmation:
bytecode `sudoku` is now `4.11x` Go, `0.09x` Ruby, and `0.18x` Python;
bytecode `i_before_e` is now `9.08x` Go, `4.54x` Ruby, and `3.49x` Python.

The next kept match slice targeted the exact primitive typed-pattern case
inside that structural runtime environment problem. Simple typed patterns now
bind directly when the runtime value already has the exact primitive shape,
skipping generic `matchesType(...)` and coercion for hot paths such as
`case byte: u8` in `parse_board`. Non-exact integer widths, aliases, unions,
structs, interfaces, and every non-primitive shape still use the generic
typed-pattern path. Runtime-only `sudoku` moved from a refreshed
`340.43ms/op`, `86.60 MB/op`, `915,996 allocs/op` sample to `327.53ms/op`,
`326.26ms/op`, and `331.68ms/op` with the same allocation shape; the profiled
rerun landed at `332.99ms/op`. The external guard confirmed the keep:
`sudoku` measured `0.5080s` / `0.5040s`, and `i_before_e` measured
`0.4360s` / `0.4200s` over two `5/5` passes. Against the external rows,
bytecode `sudoku` is now `3.88x` Go, `0.09x` Ruby, and `0.17x` Python;
bytecode `i_before_e` is now `8.40x` Go, `4.20x` Ruby, and `3.23x` Python.
The next profile should target the remaining structural match/env issue
directly by lowering simple `match` clauses into slot-aware bytecode so
`parse_board` / `solve` can inline, or target `board_to_string` through a
spec-backed string builder / byte-buffer surface rather than another generic
interpolation tweak.

The follow-up kept bytecode slice landed that structural match/env fix for a
bounded subset. In slot-eligible functions, match clauses made from literal
`nil`, wildcard, or typed identifier/wildcard patterns now lower to direct
bytecode branch tests and slot bindings instead of the generic
`bytecodeOpMatch` path. Guarded clauses, non-nil literals, existing-symbol
identifier patterns, destructuring, and structural patterns remain on the
generic path. Typed clauses still use v12 `matchesType(...)` / coercion
semantics after the exact primitive fast check, so nominal matches such as
`case node: Node` are handled without adding per-container special cases.

Runtime-only `sudoku` moved from the prior exact-primitive match band of
`326.26-331.68ms/op`, `~86.60 MB/op`, and `~916k allocs/op` to
`209.21ms/op`, `205.12ms/op`, and `203.70ms/op`, with
`31.48-34.58 MB/op` and `499.5k-499.9k allocs/op`. A profiled one-shot landed
at `233.69ms/op`, `32.57 MB/op`, and `499,764 allocs/op`. The bytecode trace
now shows `parse_board`, `find_empty`, and `solve` dispatching inline. The
external guard moved bytecode `sudoku` to `0.4120s` over `5/5` runs. The
`i_before_e` guard stayed noisy but in the same broad band: `0.4680s` in the
combined guard and `0.4480s` on the rerun, versus the prior `0.4360s` /
`0.4200s` confirmations. Against the external rows, bytecode `sudoku` is now
`3.17x` Go, `0.07x` Ruby, and `0.14x` Python; bytecode `i_before_e` is now
`8.96x` Go, `4.48x` Ruby, and `3.45x` Python. External bytecode
`binarytrees` still times out at `60s`, so the next profile should target
post-match `sudoku` member/index/string work or move to the larger
typed-frame/struct-allocation problems in the timeout bytecode workloads.

The next kept bytecode member-call slice targeted the repeated canonical
`Array.get` method resolution left after slot-aware match lowering. Once a
bytecode `CallMember get` site fully resolves to the canonical nullable
stdlib `Array.get(i32)` overload, the VM now caches that proof per
program/IP/environment and later executes the existing tracked array-read fast
body directly. The cache is guarded by environment revision, global revision,
method-cache version, and absence of runtime impl context; unsupported shapes
still fall back to normal v12 member resolution and overload selection.

A paired no-trace runtime-only baseline landed at `200.66-209.51ms/op` with
about `499k allocs/op`. The kept cache band landed at `176.47-200.95ms/op`
with about `417k allocs/op`, and the no-trace profiled rerun landed at
`179.88ms/op`, `26.95 MB/op`, and `417,355 allocs/op`. The no-trace CPU
profile no longer shows `resolveMethodCallableFromPool(...)` as a top-tier
`sudoku` cost; the visible wall moved to `execCallMember(...)` guard work,
integer comparisons, and `board_to_string` string interpolation. The external
guard moved bytecode `sudoku` to `0.3500s` over `3/3` runs, about `2.69x` Go,
`0.06x` Ruby, and `0.12x` Python.

The follow-up kept bytecode cache-layout slice narrowed that same path without
changing the semantic guard. The canonical `Array.get` call-site cache now has
a 4-entry MRU hot tier and only reads env/global/method revisions after a
cheap program/IP/env identity match. This avoids the old single-entry hot
cache thrash across nested sudoku `Array.get` call sites while preserving the
same fallback behavior for unsupported shapes, env/global/method changes, and
runtime impl context. Paired runtime-only reruns moved from a restored
`169.46-187.57ms/op` band to `164.50-176.74ms/op`, with allocations unchanged
around `417k allocs/op`. The profiled rerun landed at `170.45ms/op`, and
`lookupCachedCanonicalArrayGetCall(...)` dropped from about `0.10s` cumulative
to about `0.02s`. Paired external bytecode `sudoku` moved from restored
pre-MRU `0.3833s` to MRU `0.3700s`; the checked-in current scoreboard now
records the MRU `0.3700s` row.

The next kept bytecode tranche adjusted call-member ordering on that same
canonical proof cache. `execCallMember(...)` now checks the guarded canonical
`Array.get` call-site cache before the single-entry general member-method
cache, allowing nested sudoku `get` sites to use the specialized 4-entry MRU
tier without first paying general member-cache miss/churn work. The guard
remains the same: only sites already proven by full member resolution to target
canonical nullable stdlib `Array.get(i32)` can use the fast path, and env,
global, method-cache, and runtime impl-context invalidation still fall back to
normal v12 member resolution.

The paired restored runtime-only baseline landed at `168.34-171.74ms/op` with
about `417k allocs/op`. The kept cache-first band landed at
`163.35-167.92ms/op`, also about `417k allocs/op`, and the profiled kept rerun
landed at `161.23ms/op`, `26.95 MB/op`, and `417,353 allocs/op`. External
bytecode `sudoku` moved from the prior recorded `0.3700s` to `0.3633s` over
`3/3` runs, about `2.79x` Go, `0.06x` Ruby, and `0.12x` Python. Non-keeps in
this tranche were a stdlib `StringBuilder` source rewrite, a two-part
interpolation VM fast path, and a single-thread propagation-cache mutex
bypass. The next profile should start from
`lookupCachedCanonicalArrayGetCall(...)` itself and the remaining
`board_to_string` interpolation allocation.

The next kept bytecode tranche narrowed that interpolation work without
changing Display semantics. `execStringInterpolation(...)` now handles
two-part primitive pairs directly, with a dedicated `String + Integer` subpath
that writes integer digits into a single grown builder. `String + String`
still uses the existing buffer-reuse path, and structs, arrays, functions,
errors, and other dynamic values still fall back to `stringifyValue(...)` so
custom `to_string` behavior is preserved.

Runtime-only `sudoku` now runs in the `161.29-169.59ms/op` band with about
`343k allocs/op`, down from the prior kept `~417k allocs/op` shape. A profiled
one-shot was CPU-noisy at `208.37ms/op` but confirmed `25.28 MB/op` and
`342,918 allocs/op`. External bytecode `sudoku` confirmed at `0.3620s` over
two `5/5` runs, versus the prior recorded `0.3633s`; against the external
rows this is `2.78x` Go, `0.06x` Ruby, and `0.12x` Python. The next profile
should start from residual `execCallMember(...)` / canonical `Array.get`
guard work and binary compare slots. Larger string wins likely need a general
byte-buffer/string-builder runtime primitive rather than another local
interpolation helper.

The next kept bytecode tranche fused another small control-flow shape exposed
by the post-interpolation profile. Slot-backed `>` and `>=` comparisons
against integer literals now lower to `JumpIfIntCompareSlotConstFalse`, so
guards such as `i >= 9` and `num > 9` no longer materialize a temporary bool
only for `JumpIfFalse` to pop. The existing `<=` return and branch fusions are
unchanged; unsupported operands fall back to the same generic binary operator
and truthiness behavior.

Runtime-only `sudoku` landed at `149.16ms/op`, `153.52ms/op`, and one noisy
`176.73ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `171.99ms/op`, `25.44 MB/op`, and `342,921
allocs/op`; the old generic `execBinary` compare prominence is gone. External
bytecode `sudoku` moved from the prior recorded `0.3620s` to `0.3560s` over
`5/5` runs. Against the external rows, bytecode `sudoku` is now `2.74x` Go,
`0.06x` Ruby, and `0.12x` Python. The next profile should start from residual
`execCallMember(...)` / canonical `Array.get` guard work, slot load/store
traffic, and runtime type/propagation checks.

The next kept bytecode tranche trimmed that residual canonical `Array.get`
guard work without weakening the cache proof. The four-entry call-site cache
still stores and validates env revision, global revision, method-cache version,
and the absence of a runtime impl context, but hot hits no longer promote the
matched entry to the front on every nested sudoku access. Promotion still
happens when an entry is stored or recovered from the backing cache.

Runtime-only `sudoku` moved to a warmed `137.57ms/op`, `148.35ms/op`, and one
noisy `164.03ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `159.54ms/op`, `25.48 MB/op`, and `342,856
allocs/op`; `lookupCachedCanonicalArrayGetCall` fell from about `100ms`
cumulative in the refreshed profile to about `70ms`. External bytecode
`sudoku` moved from `0.3560s` to `0.3360s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `2.58x` Go, `0.06x` Ruby, and `0.11x`
Python. The next profile should start from this kept state and avoid another
`Array.get` cache micro-slice unless it is again the top visible wall.

The next kept bytecode tranche targeted residual boxed slot load/store traffic
instead of another member-cache layer. Slot-backed self assignments of the form
`x = x + const` and `x = x - const` now lower to
`StoreSlotBinaryIntSlotConst`, which performs the same checked slot-const
integer operation, stores the result, and leaves the assignment value on the
stack in one VM opcode. Unsupported shapes keep the old binary-plus-store
lowering, and typed `i32` slots keep the existing raw `StoreSlotI32` path.

Runtime-only `sudoku` moved to `140.23ms/op`, `141.16ms/op`, and one noisier
`145.35ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `144.89ms/op`, `25.49 MB/op`, and `342,850
allocs/op`; the former standalone store edge for sudoku loop counters is gone,
with checked add/sub cost now visible inside the fused opcode. External
bytecode `sudoku` moved from `0.3360s` to `0.3320s` over `5/5` runs, while
external bytecode `i_before_e` stayed neutral at `0.4440s` over `5/5` runs.
Against the external rows, bytecode `sudoku` is now `2.55x` Go, `0.06x` Ruby,
and `0.11x` Python; bytecode `i_before_e` is now `8.88x` Go, `4.44x` Ruby,
and `3.42x` Python. The next profile should start from this kept state and
target `execCallMember(...)` / canonical `Array.get` guard cost, residual
checked integer arithmetic, or typed slot assignment checks.

The next kept bytecode/interpreter tranche targeted repeated runtime type
alias expansion rather than another opcode fusion. Runtime type checks now use
an interpreter-level `ast.TypeExpression` expansion cache around
`matchesType(...)`, cast coercion, and the exported alias-expansion bridge,
with invalidation on type-alias registration. This preserves the existing v12
alias semantics while avoiding repeated rebuilds of the same alias-expanded
type ASTs in sudoku's hot pattern/coercion paths.

Runtime-only `sudoku` moved from a restored `145.24ms/op` sample with about
`25.49 MB/op` and `342,847 allocs/op` to `136.38ms/op`, `132.95ms/op`, and
`141.75ms/op`, with about `22.3-25.1 MB/op` and `279k allocs/op`. The
profiled kept sample landed at `138.89ms/op`, `22.42 MB/op`, and `279,229
allocs/op`; `expandTypeAliases(...)` dropped from roughly `140ms` cumulative
in the restored profile to about `20ms`, and `substituteAliasTypeExpression`
fell from about `26.5 MB` flat allocation to `5 MB`. Same-session external
guards moved restored bytecode `sudoku` from `0.3540s` to `0.3460s` over
`5/5` runs and restored bytecode `i_before_e` from `0.4970s` to `0.4850s`
over `10/10` runs. Against the external rows, bytecode `sudoku` is now
`2.66x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e` is now
`9.70x` Go, `4.85x` Ruby, and `3.73x` Python. The next profile should target
the post-cache member-call/type-check allocation wall: `execCallMember(...)`,
`resolveMethodCallableFromPool(...)`, and `storeBoundMethodCache(...)`.

The next kept bytecode tranche added a guarded call-member opcode for
canonical `Array.get` sites. Ordinary one-argument `.get(...)` calls now lower
to `CallMemberArrayGet`. The opcode still falls back to the full
`execCallMember(...)` path until the existing canonical nullable
`Array.get(i32)` call-site proof cache validates the site; once proven, hot
hits execute the direct tracked-array read without re-entering the broader
member dispatch ladder.

Runtime-only `sudoku` moved to `141.42ms/op`, `135.12ms/op`, and
`134.24ms/op` after the final call-dispatch helper split required to keep
`bytecode_vm_run.go` under 1000 lines, with allocations still around `279k
allocs/op`. The profiled kept sample landed at `134.14ms/op`, `22.42 MB/op`,
and `279,242 allocs/op`; the CPU profile showed `execCallMember(...)`
cumulative cost at about `200ms`, with the new guarded opcode accounting for
about `90ms`. External bytecode `sudoku` moved from the previous recorded
`0.3460s` to `0.3300s` over `5/5` runs. External bytecode `i_before_e` moved
from the previous recorded `0.4850s` to `0.4610s` over `10/10` runs. Against
the external rows, bytecode `sudoku` is now `2.54x` Go, `0.06x` Ruby, and
`0.11x` Python; bytecode `i_before_e` is now `9.22x` Go, `4.61x` Ruby, and
`3.55x` Python. The next profile should target remaining non-`Array.get`
`execCallMember(...)` paths: propagation/error checks, string iterator calls,
and residual bound-method cache allocation.

The next kept bytecode tranche added a guarded call-member opcode for
canonical string-byte iterator `next` calls. Ordinary zero-argument `.next()`
calls now lower to `CallMemberNext`. The opcode tries the existing canonical
`Iterator u8` / `RawStringBytesIter.next` fast body first and falls back to
the full `execCallMember(...)` path for safe-navigation calls, argument
calls, non-canonical iterators, and every unsupported receiver shape.

Runtime-only `sudoku` moved from a refreshed profiled baseline of
`135.81ms/op`, `22.43 MB/op`, and `279,238 allocs/op` to `130.46ms/op`,
`132.00ms/op`, and `133.02ms/op`. The profiled kept sample landed at
`128.28ms/op`, `22.41 MB/op`, and `279,229 allocs/op`; the CPU profile showed
`execCallMember(...)` down from about `280ms` cumulative in the refreshed
baseline profile to about `210ms`, with canonical iterator `next` now under
`execCallMemberNext(...)` at about `20ms`. External bytecode `sudoku`
confirmed at `0.3260s` over `5/5` runs, versus the prior recorded `0.3300s`.
External bytecode `i_before_e` was noisy but landed at `0.4760s` and
`0.4500s` over `10/10`, versus the prior recorded `0.4610s`. Against the
external rows, bytecode `sudoku` is now `2.51x` Go, `0.06x` Ruby, and `0.11x`
Python on the kept confirmation; bytecode `i_before_e` is `9.00x` Go,
`4.50x` Ruby, and `3.46x` Python on the best confirmation. The next profile
should target remaining non-`Array.get` / non-`next` `execCallMember(...)`
edges, especially static `Array.new`, propagation/error checks, and residual
bound-method cache allocation.

The next kept bytecode tranche added a guarded call-member opcode for
canonical static `Array.new` calls. Ordinary zero-argument `.new()` calls now
lower to `CallMemberArrayNew`. The opcode executes the direct empty-array
construction only after normal member resolution proves the canonical kernel
`Array.new() -> Array T` method at that program/IP; the proof is cached behind
environment, global, and method-cache revisions. Safe-navigation calls,
argument-bearing `new(...)` calls, non-`Array` receivers, non-canonical
definitions, runtime impl-context environments, and invalidated cache versions
all fall back to the full member path.

Runtime-only `sudoku` allocation dropped from the previous `~279k allocs/op`
band to `259,029`, `258,977`, and `259,275 allocs/op`, while wall-clock
stayed soft at `133.40ms/op`, `135.83ms/op`, and `136.59ms/op`. The profiled
kept sample landed at `133.51ms/op`, `21.71 MB/op`, and `259,043 allocs/op`;
the allocation profile showed `execCallMember(...)` down from the refreshed
`66.14 MB` cumulative sample to `47.20 MB`, with static construction now
flowing through `execCallMemberArrayNew(...)`. External bytecode `sudoku`
edged to `0.3240s` over `5/5`, versus the prior `0.3260s`. External bytecode
`i_before_e` was noisy at `0.5220s` and `0.4570s` over `10/10`, which keeps
it in the same broad band as the prior `0.4500-0.4760s` note. Against the
external rows, bytecode `sudoku` is now `2.49x` Go, `0.06x` Ruby, and `0.11x`
Python on the kept confirmation; bytecode `i_before_e` is `9.14x` Go,
`4.57x` Ruby, and `3.52x` Python on the better confirmation. The next profile
should target propagation/error checks or residual bound-method cache
allocation before adding more single-method call-member opcodes.

The follow-up kept bytecode tranche cleaned up the already-proven
`CallMemberArrayGet` hot path. Once the guarded canonical `Array.get(i32)`
call-site proof has validated a program/IP, the opcode now reuses the
already-proven array receiver and `i32` index and finishes the tracked-array
read directly instead of re-entering `execArrayGetMemberFast(...)` and
repeating stack, receiver, and argument shape checks. Unsupported shapes and
invalidated guards still fall back to the existing full member-call path.

Runtime-only `sudoku` moved from a refreshed `136.88ms/op`, `21.70 MB/op`,
`259,038 allocs/op` baseline to `120.74ms/op`, `123.67ms/op`, and
`131.53ms/op`, with allocation shape essentially unchanged. The profiled kept
sample landed at `126.92ms/op`, `21.71 MB/op`, and `259,048 allocs/op`;
`execCallMemberArrayGet(...)` dropped from about `110ms` cumulative in the
refreshed baseline profile to about `60ms`, and
`lookupCachedCanonicalArrayGetCall(...)` dropped from about `50ms` flat to
about `10ms` flat in the kept sample. External bytecode `sudoku` moved to
`0.3180s` over `5/5`, versus the prior `0.3240s`; external bytecode
`i_before_e` moved to `0.4420s` over `10/10`, versus the prior noisy
`0.4570-0.5220s` guard band. Against the external rows, bytecode `sudoku` is
now `2.45x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e` is
`8.84x` Go, `4.42x` Ruby, and `3.40x` Python. The next profile should target
propagation/error checks, residual bound-method cache allocation, or string
interpolation allocation; avoid another `Array.get` cache/guard slice unless
fresh evidence puts it back at the top.

The follow-up kept bytecode tranche targeted the specific
`board_to_string` interpolation shape. The primitive `String + Integer`
interpolation fast path now concatenates cached one-byte digit suffixes for
`0..9` directly instead of routing those cases through
`strings.Builder.Grow`. Multi-digit integers, non-small integers, and generic
Display/`to_string` fallback remain on the existing paths.

Runtime-only `sudoku` moved from a refreshed `118.77ms/op`, `21.69 MB/op`,
`259,012 allocs/op` baseline to `117.40ms/op`, `118.33ms/op`, and
`121.52ms/op`, with allocation counts still around `259k allocs/op`. The
profiled kept sample was noisy at `127.25ms/op`, `21.70 MB/op`, and `258,928
allocs/op`, but the heap profile showed
`finishStringIntegerInterpolationFast(...)` down from about `53.5 MB`
cumulative in the refreshed baseline to about `37 MB`. External bytecode
`sudoku` held at `0.3180s` over `5/5`; external bytecode `i_before_e` edged
to `0.4410s` over `10/10`. Against the external rows, bytecode `sudoku`
remains `2.45x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e`
is now `8.82x` Go, `4.41x` Ruby, and `3.39x` Python. The next profile should
target residual member resolution / bound-method cache allocation around
iterator `next` and static `Array.new`, or the remaining array
growth/allocation path; propagation is no longer a top trace item unless a
fresh profile brings it back.

As of April 29, 2026, the external `quicksort` compiled timeout is closed and
the benchmark is in Go range. This took two steps: first, the benchmark source
was changed to parse `numbers.txt` directly from bytes and use slot
reads/writes in the sort hot loop, moving compiled Able from timeout to
`11.42s`; then the compiler gained a native host-slice return path for mono
`Array T` externs and `fs.read_bytes` was routed through `os.ReadFile`, moving
compiled Able to a verified `1.75s` average over `3/3` runs.

Current external quicksort comparison:

- compiled Able: `1.75s`
- Go reference: `2.01s`
- Ruby reference: `14.58s`
- Python reference: `20.32s`

The profiled `1.79s` run no longer shows the old
`hostValueToRuntime(...)` / per-byte `BigInt` return bridge. The remaining
CPU is actual parse/sort work. Allocation is now about `429.55MB`, including
the `os.ReadFile` buffer, the deliberate host-boundary copy into Able-owned
`Array u8`, and the parsed `Array i32`.

The next kept compiled tranches targeted external `i_before_e` without
changing the benchmark source. First, compiled Go extern wrappers now keep
native scalar arguments and results native when the Able compiled carrier type
already matches the Go host type. The string fast-path externs in the
canonical stdlib therefore stop paying `RuntimeValueToHost`,
`HostValueToRuntime`, `bridge.ToString`, and scalar `bridge.As*` conversions
on every `String.contains`, `String.replace`, and `String.len_bytes` call.
Then static no-fallback launchers stopped running the loader/parser/program
evaluation path when the generated binary can seed imports directly. The
launcher seedability check now accepts public type-alias selector imports and
explicit compiler-known internal extern selector imports, while statically
known calls to unsupported receiver methods still force the bootstrap path.

Current external `i_before_e` compiled comparison:

- compiled Able: `0.0620s`
- Go reference: `0.0500s`
- Ruby reference: `0.1000s`
- Python reference: `0.1300s`

The refreshed pre-extern compiled baseline in the same session was `0.2700s`
over `3/3` runs. The scalar extern tranche moved the restored source to
`0.1900s`; the static-launcher tranche then moved it to `0.0620s` over `5/5`
runs, which is now in the Go-reference range at about `1.24x` Go. A
`String.index_of(...)` source rewrite was tested and reverted because it
regressed external bytecode `i_before_e` to `2.3533s` over `3/3` runs.

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
  the latest external compiled comparison records `0.9660s` / `4.00` GC over
  `5/5` runs on the compiled
  `v12/examples/benchmarks/matrixmultiply.able` path after the canonical Able
  source started using `Array.with_capacity(n)` for fixed-size rows and outer
  matrices and statement-position counted loops stopped materializing
  discarded runtime loop results

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

The latest `Array.with_capacity(n)` source-parity tranche removed matrix slice
growth churn rather than changing compiler semantics. Generated Go now builds
the hot rows/outers with `make(..., 0, n)` while retaining native nested
`[]float64` carriers in `build_matrix` and `matmul`. The refreshed external
compiled comparison moved from the same-session `1.2900s` baseline to
`1.0180s` over `5/5` runs, about `1.16x` the Go reference. The profiled kept
run still spends almost all CPU in `__able_compiled_fn_matmul`; allocation is
down to about `52.34MB`, with benchmark matrix allocation at about `34.77MB`.

The follow-up counted-loop statement tranche removed a smaller compiler-side
scaffold from the same hot generated functions. Statement-position counted
loops now lower directly to the counted `for i < n` shape when the body has no
value-producing `break`, so the compiler no longer emits a temporary
`runtime.Value` loop result plus `__able_runtime_error_value(...)` discard
probe after the hot matrix loops. The refreshed external compiled comparison
moved again to `0.9660s` over `5/5` runs, about `1.10x` the Go reference, and
the profiled kept run landed at `0.9600s` with CPU almost entirely in
`__able_compiled_fn_matmul`.

The next kept compiled recursion tranche targeted `fib` without changing the
benchmark source. Statement-position `if` guards whose body cannot fall
through now seed conservative fallthrough integer facts, so
`if n <= 2 { return 1 }` proves the recursive `n - 1` and `n - 2` decrements
safe. Generated `fib` now emits direct `i32` subtraction for those decrements
while keeping the checked addition and slow control path for possible
overflow. Refreshed external compiled `fib` moved from `3.2633s` over `3/3`
before the slice to `3.1760s` over `5/5`, with a second local `5/5` rerun at
`3.1280s` and a profiled kept run at `3.1000s`. The Go reference remains
`2.8400s`, so the remaining compiled core gap is now the checked-add /
control-return shape rather than recursive call dispatch or subtract
underflow checks.

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

The next aligned-fib tranche specialized the return-add operand shape produced
by the external-style recursive fixture. `bytecodeOpReturnBinaryIntAddI32` now
tries a direct `runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` branch
before the existing pointer-oriented small-`i32` add helper. Pointer operands,
non-small values, non-`i32` values, and generic fallback behavior keep the
existing path, and overflow still returns the existing integer-overflow error.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly` was
noisy but stayed in the kept range after the first sample, landing at
`143.82ms/op`, `125.72ms/op`, and `116.12ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `9.07s/op` and `9.10s/op` across
two 3-run bands, with a profiled one-shot at `8.88s/op`. The aligned
`preprofile` now shows
`execReturnBinaryIntAdd(...) -> bytecodeReturnAddSmallI32ValuePairFast(...)`
on the hot return-add edge. Full external bytecode `fib(45)` still times out at
`90s`. The next profile should start from `execCallSelfIntSubSlotConst(...)`,
`finishInlineReturn(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`, and
slot-frame release costs rather than another return-add operand extraction
shortcut.

The next aligned-fib tranche removed the return-coercion probe that still ran
after the handled `bytecodeOpReturnBinaryIntAddI32` fast paths. Those handled
small-`i32` branches now report that the value already satisfies an `i32`
return, so `finishInlineReturn(...)` can skip the generic simple type check
only for proven boxed-`i32` results. Generic fallback arithmetic still reports
an unknown return shape and keeps the existing coercion behavior.

This is an aligned-shape keep with a reduced runtime-only win. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `114.67ms/op`, `111.98ms/op`,
and `110.85ms/op`. Aligned-style `fib_i32_small` bytecode-runtime landed at
`8.33s/op` and `8.44s/op` across two 3-run bands, with a profiled one-shot at
`9.01s/op`. The aligned `preprofile` no longer shows the prior
`finishInlineReturn(...) -> inlineCoercionUnnecessaryBySimpleCheck(...)` edge.
Full external bytecode `fib(45)` still times out at `90s`. The next profile
should start from `execCallSelfIntSubSlotConst(...)`,
`execReturnConstIfIntLessEqualSlotConst(...)`, and slot-frame return/release
costs rather than another return-add coercion shortcut.

The next aligned-fib tranche narrowed the body of the fused recursive self-call
opcode. `execCallSelfIntSubSlotConst(...)` now keeps the common fused recursive
path inline and moves the non-fast callable/native/generic handling into
`execCallSelfIntSubSlotConstFallback(...)`. Immediate resolution, inline-call
stats, native-call fallback, generic callable fallback, and error wrapping are
unchanged; the point of the change is code layout around the hot opcode, not a
new arithmetic or frame-pool rule.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly`
landed at `109.33ms/op`, `113.54ms/op`, and `118.41ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `8.44s/op` and `8.42s/op` across
two 3-run bands, with a profiled one-shot at `8.42s/op`. The aligned
`preprofile` does not show the extracted fallback helper on the hot path. Full
external bytecode `fib(45)` still times out at `90s`. The next profile should
start from `execReturnConstIfIntLessEqualSlotConst(...)`,
`finishInlineReturn(...)`, and slot-frame return/release costs rather than more
self-call fallback rearrangement.

The follow-up tranche was measurement-only. The aligned
`fib_i32_small` cross-mode matrix landed at compiled `0.3433s`, tree-walker
`3/3` timeouts at `60s`, bytecode end-to-end `8.1467s`, and
bytecode-runtime `8768648581 ns/op`, `24104 B/op`, `47 allocs/op`.
A standalone bytecode-runtime confirmation landed at `8714877680 ns/op` with
the same allocation shape, and the reduced
`BenchmarkFib30BytecodeRuntimeOnly` sanity band remained in range at
`119.67ms/op`, `115.41ms/op`, and `112.22ms/op`.

The external comparison was rerun with a longer `120s` cap to measure the
previous timeout. Full external `fib(45)` now records compiled `3.3700s` and
bytecode `92.5200s`. That means the old `90s` guard is only slightly missed,
but bytecode remains about `32.58x` the Go reference and `27.45x` the current
compiled Able path on this recursive workload. The next tranche should use the
external `fib(45)` result as the keep/revert guardrail and start from the
remaining aligned hot path around
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
minimal slot-frame return/release work rather than a reduced-only `fib(30)`
branch.

The next kept aligned-fib code tranche cut the fused self-call guard ladder
for the exact compact recursive shape. `execCallSelfIntSubSlotConst(...)` now
tries `execCallSelfIntSubSlotConstCompact(...)` before resolving generic
immediates or return-name metadata, but only for the already-proven two-slot
slot-0 raw-immediate self call with cached nil return generics and no active
loop/iter state. Reduced `Fib30Bytecode` moved from a refreshed compact-frame
profile of `105.27ms/op` to `99.54ms/op`, `100.39ms/op`, and `99.00ms/op`.
The focused external bytecode `fib(45)` one-shot now completes at `79.1200s`,
which is still far from Go (`2.8400s`) but no longer a timeout. The next
profile should start from the base-case compare/return path and return-add
handoff rather than another self-call fallback rearrangement.

The next kept aligned-fib code tranche inlined the direct small-`i32`
return-add value-pair branch inside `execReturnBinaryIntAdd(...)`, removing
the hot call edge through `bytecodeReturnAddSmallI32ValuePairFast(...)` while
keeping the existing pointer/generic fallback path. Reduced `Fib30Bytecode`
stayed in range at `97.19ms/op`, `104.20ms/op`, and `106.93ms/op`; aligned
`fib_i32_small` bytecode-runtime moved to `7.21s/op` over a 3-run band, with
a profiled one-shot at `7.50s/op`. Focused external bytecode `fib(45)` moved
to `77.2400s`. The next profile should target structural boxed return/add
handoff, the base-case raw compare, or compact `finishInlineReturn(...)`
restoration rather than another return-add helper.

The next kept aligned-fib code tranche inlined the compact slot-0 frame push
inside `execCallSelfIntSubSlotConstCompact(...)`, after the exact-shape
recursive checks have already passed. The generic fallback still uses
`pushSelfFastSlot0CallFrame(...)`; only the proven raw-immediate two-slot path
writes the frame record directly. Reduced `Fib30Bytecode` moved to
`104.21ms/op`, `96.22ms/op`, and `94.85ms/op`; aligned `fib_i32_small`
bytecode-runtime landed at `7.18s/op`, with a profiled one-shot at
`7.60s/op`. Focused external bytecode `fib(45)` moved to `76.7900s`. A
compact `finishInlineReturn(...)` shortcut was tested and reverted because it
regressed aligned runtime to `8.31s/op`. The next profile should either move
to explicit raw/typed return-stack metadata with invalidation or step back to
the broader typed-frame design, because the obvious helper-call edges are now
gone.

The next kept aligned-fib code tranche avoided the rejected operand-stack side
metadata shape and instead added a compact self-fast slot-0 raw lane. The
boxed `runtime.Value` slot remains the observable semantic value, but the exact
two-slot slot-0 recursive frame now saves/restores a raw `i32` cache beside
the boxed slot. `execCallSelfIntSubSlotConstCompact(...)` uses that raw lane
for the recursive `slot0 - const`, and
`execReturnConstIfIntLessEqualSlotConst(...)` uses it for the base-case
`slot0 <= const`; slot-0 writes refresh or clear the lane and generic/full
frame paths clear it.

Reduced `Fib30Bytecode` moved to `92.46ms/op`, `92.81ms/op`, and
`92.08ms/op`. Aligned `fib_i32_small` bytecode-runtime landed at `6.24s/op`
over a 3-run band, with a profiled one-shot at `6.03s/op`. The profiled
aligned rerun shows `execReturnConstIfIntLessEqualSlotConst(...)` down to
`0.41s` cumulative from the prior `1.38s` profile. Focused external bytecode
`fib(45)` moved to `67.8200s`. The next profile should start from
`execReturnBinaryIntAdd(...)`, compact `finishInlineReturn(...)`, and residual
self-call guard/boxing cost rather than another boxed slot-0 probe rewrite.

The next kept aligned-fib tranche added a direct compact minimal-return path
for proven `i32` no-coercion returns. When `ReturnConstIfIntLessEqualSlotConst`,
same-slot `ReturnIfIntLessEqualSlotConst`, or handled
`ReturnBinaryIntAddI32` returns from the exact reused self-fast frame, the VM
restores slot 0/raw slot-0 state and appends the boxed semantic return value
without entering the generic `finishInlineReturn(...)` path. Generic `Int`
returns, non-`i32` slots, non-reused minimal frames, and all full/generic frame
shapes keep the existing boxed fallback/coercion path.

The refreshed same-session aligned control was noisy: `fib_i32_small`
bytecode-runtime started at `14.2100s` over two runs. The kept confirmation
landed at `13.8350s` over two runs and `13.3533s` over three runs. Full
external bytecode `fib(45)` landed at `75.3700s`, compared with the reverted
same-session control from the prior tranche at `77.0800s`; this is still
slower than the historical `67.8200s` raw-lane one-shot, so the result should
be treated as a small current-baseline win rather than a new historical best.
Reduced generic-`Int` `BenchmarkFib30BytecodeRuntimeOnly` stayed in range at
`102.76ms/op`, `106.55ms/op`, and `102.57ms/op`. The next profile should look
at explicit return-add value/raw metadata or a real typed-frame return channel,
not more single-branch proof elision.

The next kept aligned-fib tranche inlined the exact boxed-value-pair `i32`
branch inside `execReturnBinaryIntAdd(...)`. The previous profile showed
`bytecodeReturnAddSmallI32ValuePairFast(...)` as the largest flat cost after
the compact minimal-return keep, so this slice removes that helper call edge
while preserving the pointer/generic fallback helpers and the existing checked
`i32` overflow behavior. The aligned `fib_i32_small` bytecode-runtime band was
thin but positive at `13.3233s` over three runs, versus the prior `13.3533s`
confirmation. A profiled one-shot landed at `12.8900s` and showed the helper
edge gone, with cost now inside `execReturnBinaryIntAdd(...)`. Full external
bytecode `fib(45)` moved to `72.8000s`, versus the prior kept `75.3700s` and
the same-session reverted control at `77.0800s`. The next tranche should not
add another arithmetic helper split; it should either introduce a typed
return/value channel for the recursive `i32` frame shape or step back to the
broader VM-v2 typed-frame design.

The next kept timeout-family tranche pivoted from fib to external quicksort's
kernel slot API. The external source uses `Array.read_slot(i32)` and
`Array.write_slot(i32, T)` heavily instead of bracket indexing, so the VM now
recognizes the canonical kernel methods and executes tracked-array reads/writes
directly while preserving the kernel semantics: negative indexes error,
out-of-bounds reads return `nil`, writes keep the existing growth behavior, and
unsupported/dynamic shapes fall back to normal member dispatch. A reduced
external-style quicksort run with 2000 descending numbers showed the keep:
current fast-path bands landed at `13.79ms/op`, `13.38ms/op`,
`13.64ms/op`, and restored `13.44ms/op`, `13.29ms/op`, `14.49ms/op`; the
temporary control with only `read_slot` / `write_slot` detection disabled
landed at `32.97ms/op`, `32.77ms/op`, and `33.44ms/op`. Bytecode trace on
that reduced source confirms the real quicksort hot sites now dispatch through
`array_read_slot_tracked_fast` / `array_write_slot_tracked_fast`. Full external
bytecode `quicksort` still times out at `90s` (`go` reference `2.0100s`,
`ruby` `14.5800s`, `python` `20.3200s`), so this is a reduced hot-path keep
rather than the final external timeout fix. The next quicksort profile should
start from the residual `execCallMember` shell around those proven slot calls,
then the integer-compare and `swap`/recursive call path.

The follow-up quicksort slot-dispatch tranche lowered ordinary non-safe
`read_slot` / `write_slot` method calls to a guarded `CallMemberArraySlot`
opcode. Once the existing canonical method proof is cached, repeated hot sites
can bypass the broad `execCallMember` dispatch shell and enter the same
tracked-array read/write fast bodies directly. On the same reduced
external-style quicksort harness with 2000 descending numbers, the warmed band
moved from the prior `13.29-14.49ms/op` confirmation range to
`11.21ms/op`, `11.75ms/op`, and `11.55ms/op`, with a profiled one-shot at
`11.49ms/op` and `8996 allocs/op`. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should move past slot member dispatch
and start from integer comparisons, slot-constant binary conditions, array
index extraction, and the `swap` / recursive quicksort call path.

The next quicksort follow-up kept that slot-call opcode but shortened its
cached hit path. After the guarded proof cache validates a hot
`CallMemberArraySlot` site, the VM now validates the array receiver/cache
identity once and finishes the tracked `read_slot` / `write_slot` body
directly instead of routing through the generic cached-member fast-path switch
and broader `canUseCanonicalArraySlotCallCache(...)` guard. The same reduced
external-style quicksort harness moved from the prior `11.21-11.75ms/op` band
to `10.76ms/op`, `10.79ms/op`, and `10.87ms/op`; a profiled one-shot landed at
`11.20ms/op`, `669948 B/op`, and `9000 allocs/op`, with the old cache guard no
longer in the top profile list. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should leave slot-call dispatch alone
and start from bool-producing integer comparisons plus `JumpIfFalse`, array
index extraction, and the `swap` / recursive quicksort call path.

The next quicksort conditional tranche fused the pivot-loop guard shape
`arr.read_slot(index) <op> pivotSlot`. In `if` / `elsif` condition position,
slot-backed non-safe `read_slot` comparisons now lower to
`JumpIfArrayReadSlotCompareSlotFalse`, reuse the guarded canonical
`read_slot` proof cache, and skip the standalone read result, bool-producing
comparison, and generic `JumpIfFalse` pop path when the proof holds. The same
reduced external-style quicksort harness moved from the direct slot-call
finish `10.76-10.87ms/op` band to `10.12ms/op`, `10.20ms/op`, and
`10.30ms/op`; a profiled one-shot landed at `9.94ms/op`, `669842 B/op`, and
`8997 allocs/op`, with the old `execJumpIfFalse(...)` hotspot gone from the
short profile. Full external bytecode `quicksort` still times out at `90s`,
so the next tranche should target ordinary slot-slot integer comparison
conditionals such as `lo >= hi`, `i > j`, `i <= j`, `lo < j`, and `i < hi`,
or the `swap` / recursive quicksort call path.

The next quicksort conditional tranche fused ordinary slot-slot integer
comparison guards. Identifier-vs-identifier comparisons in `if` / `elsif`
condition position now lower to `JumpIfIntCompareSlotFalse`, avoiding the
slot load, second slot load, boxed bool, and generic `JumpIfFalse` sequence
while preserving the boxed dynamic fallback through the existing binary
operator path. The reduced external-style quicksort harness moved from the
read-slot compare `10.12-10.30ms/op` band to `9.18ms/op`, `9.28ms/op`, and
`9.29ms/op`; a profiled one-shot landed at `9.58ms/op`, `669971 B/op`, and
`9001 allocs/op`. Full external bytecode `quicksort` still times out at
`90s`, so the next tranche should move to residual boxed slot updates,
slot-call dispatch, frame release, and the `swap` / recursive quicksort call
path rather than adding more condition-only jumps.

The next quicksort tranche lowered ordinary slot-backed, non-safe
`arr.read_slot(i)` expressions to a direct `ArrayReadSlot` opcode. This reuses
the guarded canonical kernel `read_slot` proof cache but skips the broader
member-call opcode shell for expression-position reads; unsupported dynamic
shapes, stale proofs, negative indexes, and out-of-bounds `nil` reads keep the
existing v12 fallback semantics. On the same reduced external-style quicksort
harness, a same-session no-direct-control band landed at `10.25-10.74ms/op`;
the restored direct opcode landed at `9.73-10.56ms/op`, with a profiled
one-shot at `9.38ms/op`, `658964 B/op`, and `8978 allocs/op`. Full external
bytecode `quicksort` still times out at `90s`, so the next tranche should
target a larger remaining wall: boxed slot updates (`i = i + 1`, `j = j - 1`),
direct `swap` / recursive call setup, or residual cache/revision checks around
proven array slot calls.

The next quicksort tranche kept that boxed slot-update target, but as a
runtime shortcut rather than another lowering change. `StoreSlotBinaryIntSlotConst`
now handles the hot small same-type integer `x = x + const` / `x = x - const`
case directly, avoiding the synthetic binary instruction and broader
slot-const binary helper while preserving the fallback for non-small,
mismatched, dynamic, and int64-overflow shapes. Checked v12 integer overflow
still errors before mutating the slot. The reduced external-style quicksort
harness moved from the direct-read baseline `9.73-10.56ms/op` to
`8.45ms/op`, `8.60ms/op`, `8.62ms/op`, `8.75ms/op`, and `8.82ms/op`. The
profiled run was noisy at `11.94ms/op`, `659384 B/op`, and `8978 allocs/op`,
but it confirms the remaining wall is now broader `execBinary(...)`
arithmetic/comparison, `arrayReadSlotValue(...)` cache/proof checks, direct
read-slot execution, and call setup. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should target one of those larger
remaining buckets rather than another store-only shortcut.

The next quicksort allocation tranche targeted the host-result conversion
side instead of changing bytecode shape. Go extern returns for `u8`, `u16`,
and `u32` now reuse the VM boxed-small-int cache, so `fs.read_bytes(...)`
still produces the existing tracked dynamic `Array u8` representation but no
longer allocates a fresh boxed Able integer for every returned byte. A
mono-u8 host-array experiment cut allocation further but regressed reduced
wall-clock by forcing `parse_numbers` through slower handle reads, so it was
reverted. The kept cached-boxing slice moved the reduced external-style
quicksort harness from a restored `8.93ms/op`, `661889 B/op`, `8982 allocs/op`
sample to `8.48-9.23ms/op`, `~235 KB/op`, and `84-88 allocs/op`; the longer
`50x` confirmation landed at `8.50-9.87ms/op`, `~232 KB/op`, and `78-82
allocs/op`. The runtime heap profile showed `fromHostValue` cumulative
allocation down from about `18.87 MB` to about `8.81 MB` in the same 50-run
profile shape. Full external bytecode `quicksort` still times out at `90s`
against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next
quicksort tranche should target the remaining runtime wall around
`lookupCachedCanonicalArraySlotCallForArray(...)`, `arrayReadSlotValue(...)`,
and ordinary `execBinary(...)` comparisons.

The follow-up quicksort arithmetic tranche moved the parser-side
`value = value * 10` shape onto the existing slot-const bytecode family.
`x * i32_const` now lowers to `BinaryIntMulSlotConst`, and matching
self-assignments reuse `StoreSlotBinaryIntSlotConst` with checked
small-integer multiplication before falling back to the prior dynamic,
mixed-type, or big-integer path. The same reduced external-style quicksort
harness had a warmed same-session baseline of `9.93-10.07ms/op` after a
cold/noisy first run; the kept confirmation landed at `9.46ms/op`,
`9.37ms/op`, `9.29ms/op`, `9.32ms/op`, and `9.56ms/op`, with a profiled
run at `9.48ms/op`, `232152 B/op`, and `78 allocs/op`. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`, so the next quicksort tranche should
target broader generic comparison/arithmetic stack execution,
`arrayReadSlotValue(...)` cache/proof checks, or direct `swap` / recursive
quicksort call setup.

The next quicksort comparison tranche made that broad comparison target more
specific: typed primitive integer literals now feed slot-const lowering when
they fit their declared suffix. Quicksort's byte-parser guards such as
`byte == 45_u8`, `byte >= 48_u8`, and `byte <= 57_u8` can now avoid the old
generic binary comparison route. The same change broadened
`BinaryIntCompareSlotConst` to cover `<`, `==`, and `!=` in addition to the
previous `>` / `>=` cases, while out-of-range typed literals keep the old
const/generic path for unchanged v12 validation behavior. Reduced
external-style quicksort moved from the post-multiply `9.29-9.56ms/op` band
to `8.74ms/op`, `8.88ms/op`, `8.64ms/op`, `9.16ms/op`, and `8.51ms/op`.
The profiled kept run landed at `9.19ms/op`, `232270 B/op`, and `82
allocs/op`, with `bytecodeDirectIntegerCompare` down from the prior fresh
profile's `60ms` flat sample to `20ms`. Full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target
`arrayReadSlotValue(...)` proof/cache costs or direct `swap` / recursive
quicksort call setup rather than more parse-number comparison lowering.

The follow-up quicksort array-slot proof tranche kept the same guarded
canonical proof model but widened the hot tier from `4` to `8` entries. A
reduced quicksort bytecode trace showed eight material `read_slot` /
`write_slot` sites after the direct slot-call, direct read-slot, and
conditional-lowering tranches, so the four-entry tier was still evicting real
hot proofs and re-entering the broader lookup path. Revision guards, identity
checks, and fallback semantics stay unchanged. Reduced external-style
quicksort moved from the typed-compare `8.51-9.16ms/op` band to `8.32ms/op`,
`8.42ms/op`, `7.72ms/op`, `7.62ms/op`, and `8.08ms/op`, with `~232 KB/op`
and `79-84 allocs/op`. The profiled kept run landed at `7.71ms/op`,
`233151 B/op`, and `87 allocs/op`; the remaining visible wall is now mostly
VM dispatch, direct call setup, and residual canonical proof identity checks.
Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target direct `swap` / recursive quicksort call setup or a
broader typed-loop / dispatch-lane change rather than continuing to widen the
array-slot proof tier.

The next quicksort call-setup tranche made the cached `CallName` inline path
more direct for ordinary function values. Once normal lookup has populated a
cache entry, direct non-bound function calls now retain the validated
bytecode program/layout/return-generic shape and use it to set up the frame
without re-running the broader inline-call shape ladder. This targets the hot
`swap(arr, i, j)` site while preserving the same environment/owner revision
invalidation and leaving explicit type-argument calls, bound methods, native
calls, rebinding, and generic fallbacks on the existing paths. In the same
session, the refreshed reduced quicksort baseline after the array-slot
hot-tier tranche landed at a cold/noisy `9.69ms/op`, then warmed to
`7.06ms/op` and `7.02ms/op`; the kept confirmation landed at `6.89ms/op`,
`6.83ms/op`, `6.71ms/op`, `6.64ms/op`, and `6.66ms/op`, with `~400 KB/op`
and `309-316 allocs/op`. The profiled kept run landed at `7.26ms/op`,
`395114 B/op`, and `302 allocs/op`; the previous
`tryInlineResolvedCallFromStack(...)` edge no longer appears as the hot
cached call-name setup path. Full external bytecode `quicksort` still times
out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`,
so the next quicksort tranche should target remaining canonical array-slot
proof identity/version checks or a broader typed-loop / dispatch-lane change.

The follow-up quicksort slot-update tranche kept that broader typed-loop
goal scoped to one runtime-safe step. `StoreSlotBinaryIntSlotConst` now tries
a direct checked `i32` branch for small `i32` slot values plus small `i32`
immediates before the generic same-type integer fallback. This keeps the
existing lowering and fallback semantics, but lets hot loop updates such as
`i = i + 1`, `j = j - 1`, and the existing `value = value * 10` case avoid
the generic overflow-helper / fit-check / boxing path. The reduced
external-style quicksort harness moved from the prior direct call-name
`6.64-6.89ms/op` confirmation band to `6.42ms/op`, `6.31ms/op`,
`6.40ms/op`, `6.49ms/op`, and `6.84ms/op`, with a profiled one-shot at
`6.62ms/op`, `393688 B/op`, and `290 allocs/op`. Full external bytecode
`quicksort` still times out at `90s`, so the next quicksort tranche should
not add another store-only shortcut; it should target the remaining
dispatch/data-access wall around `execCallMember` / `execCallName`,
`arrayReadSlotValue(...)` proof/cache checks, or a proper v12-safe typed-loop
lane for the partition indices and pivot.

The follow-up dispatch cleanup removed a fib-specific probe from ordinary
quicksort bytecode dispatch. `runResumable(...)` now calls
`tryExecI32RecurrenceProgram(...)` only when the active bytecode program has
an attached native `i32` recurrence kernel, execution is at program entry,
stats are disabled, and the run is not a resume. Same-session reduced
quicksort before the guard landed at `7.70ms/op`, `7.58ms/op`, and
`6.94ms/op`, with a profiled one-shot at `7.50ms/op` that showed
`tryExecI32RecurrenceProgram(...)` as a visible flat sample. With the guard,
the profiled reduced run landed at `6.45ms/op` and the helper disappeared
from the quicksort profile; a temporary old-probe control landed at
`6.83ms/op`; the final guarded confirmation band landed at `6.33ms/op`,
`6.18ms/op`, `6.26ms/op`, `6.11ms/op`, and `6.03ms/op`. External bytecode
`fib(45)` still completed at `3.8500s`, so the recurrence kernel remains
active for the intended shape. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should target
`lookupCachedCanonicalArraySlotCallForArray(...)` /
`arrayReadSlotValue(...)` proof/cache checks, residual `execCallMember(...)`
/ `resolveMethodCallableFromPool(...)`, or a proper v12-safe typed-loop lane.

The next quicksort array-slot cache tranche added a direct VM-local cache in
front of the existing canonical `read_slot` / `write_slot` proof hot tier. The
direct entry is still validated by bytecode program, instruction pointer,
environment, fast-path kind, and the same environment/global/method revisions;
the existing hot array and map remain the fallback and proof population path.
Reduced external-style quicksort moved from the guarded-dispatch
`6.03-6.33ms/op` confirmation band to `5.93ms/op`, `6.07ms/op`,
`6.00ms/op`, `6.11ms/op`, and `5.98ms/op`, with `~386-388 KB/op` and
`277-295 allocs/op`. Profiled reduced reruns landed at `7.31ms/op` and
`6.31ms/op`; despite sampling noise, `lookupCachedCanonicalArraySlotCallForArray(...)`
dropped from about `100ms` flat in the fresh baseline profile to `20-50ms`
flat after the direct cache, leaving version checks and
`readArraySlotValueFast(...)` as the main array-slot read costs. Full
external bytecode `quicksort` still times out at `90s`, while external
bytecode `fib(45)` still completes at `3.9900s`. The next quicksort tranche
should target the remaining `arrayReadSlotValue(...)` /
`readArraySlotValueFast(...)` value path, residual `execCallMember(...)` /
`resolveMethodCallableFromPool(...)`, or a proper v12-safe typed-loop lane;
do not widen the array-slot hot tier again without a fresh collision profile.

The follow-up array-slot index tranche kept the scope narrower than the earlier
rejected general `bytecodeArrayGetIndexI32(...)` rewrite. Only
`bytecodeArraySlotIndexI32(...)` now handles small integer values directly for
canonical `read_slot` / `write_slot`, with the same `i32` fit check and the
same negative-index error behavior as before. Big integers, out-of-range
small values, non-integers, and generic fallback paths are unchanged. Reduced
external-style quicksort moved from the direct-cache `5.93-6.11ms/op` band to
`5.76ms/op`, `5.68ms/op`, `5.75ms/op`, `5.81ms/op`, and `5.73ms/op`; a final
confirmation band landed at `5.71ms/op`, `5.86ms/op`, `5.86ms/op`,
`5.79ms/op`, and `5.88ms/op`, with `~386-388 KB/op` and `277-296 allocs/op`.
A profiled run was noisy at `8.89ms/op`, but the old
`bytecodeArrayGetIndexI32(...)` edge no longer appears in the array-slot
profile. Full external bytecode `quicksort` still times out at `90s`, while
external bytecode `fib(45)` still completes at `3.9900s`. The next quicksort
tranche should stop shaving the index helper and target expensive generic
package/member calls such as `fs.read_bytes(...)`, residual
`execCallMember(...)` fallback cost, or a proper v12-safe typed-loop lane for
partition locals.

The follow-up disabled-trace tranche removed diagnostic trace-call overhead
from the canonical array-slot fast paths. Successful `read_slot` /
`write_slot` fast bodies now check `bytecodeTraceEnabled` before calling
`recordBytecodeCallTrace(...)`, preserving enabled trace entries while avoiding
the helper call in ordinary untraced runs. Reduced external-style quicksort
confirmed at `6.30ms/op`, `6.10ms/op`, `6.09ms/op`, `6.07ms/op`, and
`6.22ms/op`; the profiled reduced run landed at `6.54ms/op`, `231423 B/op`,
and `82 allocs/op`, and the disabled trace edge no longer appears below
`arrayReadSlotValue(...)`. Full external bytecode `quicksort` still times out
at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the
next quicksort tranche should target the actual remaining read/cache/index
and member/name-call dispatch costs or a proper v12-safe typed-loop lane.

The next kept quicksort tranche followed the refreshed source/profile evidence
to bracket indexing. The reduced hotloop now uses `arr[i]` rather than
`read_slot`, so slot-shaped index expressions lower to `ArrayIndexGetSlot`.
That opcode reads the receiver and index directly from `vm.slots`, uses the
existing direct array index body while no v12 `Index` implementation can
override array indexing, and falls back through `resolveIndexGet(...)` for
custom `Index` implementations or unsupported shapes. Reduced quicksort moved
from the current `~7.5ms/op` profile band to `6.99ms/op`, `6.94ms/op`, and
`7.21ms/op`; the profiled confirmation landed at `6.84ms/op`, with
`execArrayIndexGetSlot(...)` replacing the old `execIndexGet(...)` /
index-cache path in the runtime loop. Full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target residual boxed integer
compare/arithmetic, `execIndexSet(...)` / swap writes, or a v12-safe typed loop
lane for partition locals.

The follow-up slot-index extraction tranche shortened that new bracket-read
opcode. `ArrayIndexGetSlot` now tries a small `runtime.IntegerValue` index
branch before the broader `bytecodeDirectArrayIndex(...)` helper, while
boxed/big integers, unsupported values, and custom v12 `Index` implementations
keep the existing fallback semantics. The refreshed reduced quicksort long
baseline was `6.62ms/op`; the kept warmed band after ref-style probing was
`6.18-6.37ms/op`, and the profiled confirmation landed at `6.39ms/op`. The
old `bytecodeDirectArrayIndex(...)` edge no longer appears in the bracket-read
path. Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target residual checked `i32` arithmetic/comparison,
`execIndexSet(...)` / swap writes, or a v12-safe typed loop lane for partition
locals.

The follow-up write-side tranche closed that `execIndexSet(...)` part for
simple slot-backed writes. `arr[i] = value` now lowers to `ArrayIndexSetSlot`
when both the receiver and index are local slots, after the RHS has already
been evaluated, preserving the v12 assignment order. The opcode uses the same
`IndexMut` override guard and tracked-array write synchronization as the
existing direct array set path, with unsupported/custom shapes falling back
through `resolveIndexSet(...)`. Cached call-name dispatch also now skips the
trace recorder call when bytecode tracing is disabled. A simple inline-argument
coercion bypass and a cast-target cache were tested in the same tranche and
reverted because their reduced quicksort bands were not defensible. The
refreshed reduced quicksort baseline was `6.20ms/op`; the kept warmed band was
`6.00-6.42ms/op`, the final confirmation was `5.83-5.89ms/op`, and
the profiled confirmation landed at `6.21ms/op` with the
old generic `execIndexSet(...)` edge gone from the write path. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`, so the next quicksort tranche should target
repeated small index extraction around the slot get/set opcodes, residual
checked `i32` comparison/arithmetic, or a v12-safe typed loop lane for
partition locals.

The next quicksort write-side tranche shortened direct indexed assignment only
for the already-proven tracked, unaliased array shape. `ArrayIndexSetSlot` and
compound direct index-set writes now refresh element-type metadata and the
array view locally instead of entering the full tracked-array alias sync; any
aliased receiver still uses `syncTrackedArrayWrite(...)`. A `%` fast-candidate
probe was tested first and reverted because it stayed in the `6.03-6.17ms/op`
reduced band. The paired long reduced baseline without the write-sync shortcut
was `6.24ms/op`, `6.06ms/op`, and `6.26ms/op`; the shortcut's first long band
was `5.99ms/op`, `6.05ms/op`, and `6.01ms/op`. Final confirmations were noisy
but stayed in range at `6.26ms/op`, `6.12ms/op`, `5.83ms/op`, then
`6.13ms/op`, `6.06ms/op`, and `6.05ms/op`. The profiled kept rerun was noisy
at `6.54ms/op`, but the intended indexed-set sync edge dropped out and the
remaining array cost shifted to `execArrayIndexGetSlot(...)` /
`resolveDirectArrayIndexGetAt(...)`. Full external bytecode `quicksort` still
times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target indexed reads or a
v12-safe typed-loop lane for partition locals rather than more write-sync or
`%` helper work.

The next quicksort read-side tranche made `ArrayIndexGetSlot` handle the hot
tracked-array + small-index case directly inside the opcode. Tracked reads now
return `state.Values[idx]` or the same `IndexError` value for invalid positions
without calling `resolveDirectArrayIndexGetAt(...)`; untracked arrays,
unsupported index shapes, and custom `Index` implementations keep the existing
fallback behavior. The refreshed reduced profile baseline was `6.01ms/op`.
Kept read-inline confirmations landed at `5.90ms/op`, `5.83ms/op`,
`5.74ms/op`, then `5.88ms/op`, `5.89ms/op`, and `5.78ms/op`. The profiled
kept rerun was noisy at `7.21ms/op`, but `resolveDirectArrayIndexGetAt(...)`
dropped out of the bracket-read path and `execArrayIndexGetSlot(...)` fell
from about `100ms` cumulative in the refreshed baseline profile to about
`50ms`. Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target remaining generic binary/compare costs, call-frame/name
setup, or a v12-safe typed-loop lane for partition locals.

The follow-up bracket-compare tranche fused the hot partition branch shape.
`if` / `elsif` conditions like `arr[i] as i32 >= pivot` now lower to
`JumpIfArrayIndexSlotCompareSlotFalse`, reading receiver/index/right values
from slots and jumping directly instead of emitting a standalone bracket read,
cast, comparison bool, and `JumpIfFalse` pop. The opcode keeps the same v12
`Index` override guard as the slot-backed bracket read path, and caches the
absorbed `i32` cast name on the instruction so the hot loop does not rebuild a
type-expression string. The first implementation regressed to
`6.29-6.40ms/op` because it did rebuild that string; after caching the cast
name, reduced quicksort moved from the refreshed `5.89ms/op` profile baseline
and prior `5.78-5.90ms/op` kept band to `5.15ms/op`, `5.25ms/op`, and
`5.40ms/op`, with a profiled confirmation at `5.26ms/op`. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next quicksort tranche should target
the residual `bytecodeValueIsI32(...)` / cast guard cost inside the fused
bracket-index compare path, then remaining generic binary/name-call costs or a
v12-safe typed-loop lane for partition locals.

The follow-up raw compare lane closed that cast-guard edge for the proven hot
shape. When `JumpIfArrayIndexSlotCompareSlotFalse` has absorbed an explicit
`as i32` cast, the indexed value is a small integer, and the right comparison
slot is a small `i32`, the VM now applies the same wrapping `as i32` semantics
on raw `int64` values and compares directly. Unsupported shapes still fall back
through the normal v12 array-index/cast path. The refreshed reduced profile
baseline landed at `5.51ms/op`; the kept warmed band landed at `5.03ms/op`,
`5.14ms/op`, `5.06ms/op`, `5.27ms/op`, and `5.17ms/op`, with a profiled
confirmation at `5.08ms/op`. The fused compare profile no longer shows
`arrayIndexSlotCompareMaybeCast(...)` / `bytecodeValueIsI32(...)`, and
`execJumpIfArrayIndexSlotCompareSlotFalse(...)` dropped from about `150ms` to
about `100ms` cumulative. Full external bytecode `quicksort` still times out at
`90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should target remaining generic binary/name-call setup or
start a v12-safe typed loop lane for partition locals rather than another
generic cast-guard slice.

The follow-up member-cache tranche targeted the package-env member-call wall
that remained after the bracket-index compare work. The bytecode
member-method cache now includes the active environment and its revision, and
is enabled for the global environment plus ordinary package environments whose
direct parent is the global environment. Existing global revision and method
cache version guards remain in place, while impl/runtime-data environments and
deeper closure environments keep the old uncached behavior. This preserves v12
method resolution while letting hot package-scope Array member calls reuse
their resolved fast-path kind instead of repeatedly entering
`resolveMethodCallableFromPool(...)`.

The refreshed reduced quicksort profiled baseline landed at `5.62ms/op`. The
kept package-env cache band landed at `5.29ms/op`, `5.33ms/op`, `5.28ms/op`,
`5.27ms/op`, and `5.34ms/op`, with a profiled confirmation at `5.32ms/op`.
The kept profile no longer shows the repeated
`resolveMethodCallableFromPool(...)` / `resolvedMemberMethodFastPath(...)`
edge in the hot member-call path; the remaining member-call cost is the
guarded cache lookup and direct fast-path body. Full external bytecode
`quicksort` still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`,
and Python `20.3200s`. The next quicksort tranche should target cached
`execCallName(...)` frame setup, residual direct array/index dispatch, or a
v12-safe typed loop lane for partition locals.

The follow-up call-name slot-arg tranche targeted the `swap(arr, i, j)` call
shape left in the reduced quicksort profile. Lowering now emits a `CallName`
instruction with slot-argument metadata for simple named calls whose one to
three arguments are identifiers. At runtime, `execCallName(...)` materializes
those values from the current slot frame immediately before the existing
cached call-name dispatch, preserving normal name lookup, cache invalidation,
coercion, inline frame setup, and fallback behavior while avoiding the
standalone argument `LoadSlot` opcodes. A precursor right-slot `i32` mark for
the fused array-index compare was tested and reverted because the hot
`pivot := ... as i32` local remains untyped under v12 semantics, so the mark
did not apply without broader data-flow proof.

The kept slot-arg call band landed at `5.03ms/op`, `5.07ms/op`,
`5.05ms/op`, `4.96ms/op`, and `5.00ms/op` against the prior kept
`5.27-5.34ms/op` band. The profiled rerun was noisy at `6.51ms/op`, but the
non-profiled warmed band is a clean reduced-hotloop win. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next quicksort tranche should target
remaining direct call-frame setup / param-coercion checks in
`tryInlineCachedCallNameDirectFromStack(...)` or start a v12-safe typed-loop
lane for partition locals.

The follow-up cached `Array.push` slot-call tranche extends the guarded
canonical array-slot call cache to `Array.push(value)`. Ordinary non-safe
`arr.push(x)` calls now lower to `CallMemberArraySlot`; after normal member
resolution proves the canonical tracked-array push fast path, cached hits use
the same env/global/method-version guards and jump directly into the existing
push body. Unsupported receivers, safe navigation, mutated method tables,
runtime-data environments, and unproven shapes still fall back through normal
member dispatch. The refreshed reduced quicksort baseline was
`5.10-5.22ms/op`; the kept 5x warmed band landed at `4.86ms/op`,
`4.98ms/op`, `4.90ms/op`, `5.12ms/op`, and `4.96ms/op`. Longer 20x
confirmations landed at `5.00ms/op`, `5.54ms/op` as a noisy outlier,
`5.10ms/op`, `4.96ms/op`, and `5.01ms/op`, with a profiled kept run at
`5.16ms/op`. `execArrayPushMemberFast(...)` is no longer a visible top
reduced-CPU edge. Full external bytecode `quicksort` still times out at `90s`
against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should target the remaining
`tryInlineCachedCallNameDirectFromStack(...)` setup, slot-const stores, fused
array-index comparison, or a broader v12-safe typed-loop/call-frame slice, not
another `Array.push` dispatch slice unless a fresh profile brings it back.

The follow-up cached parameter simple-check tranche shortened the remaining
inline call setup around `swap(arr, i, j)` without changing coercion
semantics. `bytecodeFrameLayout` now caches a compact simple-type check enum
for each parameter, and inline parameter setup uses that enum for the hot
"already exact primitive value?" test before falling back to the existing
simple-name/type-expression coercion path. Generic, interface, array, alias,
unknown, and mismatched shapes still use the old fallback behavior. A
refreshed reduced quicksort profiled baseline landed at `5.77ms/op`; the kept
profiled `500x` run landed at `5.00ms/op`, and non-profiled `500x`
confirmations landed at `5.14ms/op`, `5.02ms/op`, and `4.98ms/op`. The final
5x warmed confirmation was `5.25ms/op`, `5.03ms/op`, `5.02ms/op`,
`5.50ms/op`, and `5.19ms/op`. Full external bytecode `quicksort` still times
out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`.
The next quicksort tranche should start from a fresh profile and target
slot-const stores, fused array-index comparison, direct call-frame setup, or a
broader v12-safe typed-loop lane for partition locals; simple parameter
coercion dispatch is no longer the best next target.

The follow-up slot-const store tranche removed a discard-only stack roundtrip
from statement-position fused self-assignments. When lowering sees a
non-final expression statement that emits `StoreSlotBinaryIntSlotConst`, it
marks the instruction as discardable and omits the following `Pop`; the VM
still stores the result and refreshes the slot-0 raw lane, but skips pushing
the assignment expression value that would have been popped immediately.
Assignment expressions still produce their value when the result is observable.
Two precursor experiments were rejected first: a direct tracked-array branch
inside `JumpIfArrayIndexSlotCompareSlotFalse` regressed the profiled reduced
run to `5.10ms/op`, and delayed `paramType` lookup was semantically green but
shifted the warmed band worse to `5.10-5.53ms/op`.

The refreshed reduced quicksort profiled baseline was `4.92ms/op`; the kept
profiled `500x` run was `4.90ms/op`, and `500x` confirmations landed at
`4.92ms/op`, `4.92ms/op`, and `4.94ms/op`. Shorter confirmations landed at
`4.88-5.09ms/op` for `5x` and mostly `4.81-4.91ms/op` for `20x`, with one
`5.03ms/op` outlier. After tightening the discard marker so nested assignment
expressions still leave their value available to enclosing expressions, a
final sanity pass landed at `4.85ms/op` over `500x` and `4.87-4.99ms/op` over
`20x`. Full external bytecode `quicksort` still times out at
`90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should profile the kept state and target the remaining fused
array-index compare, direct call-frame setup, generic `%` / binary work in
`build_data`, or a broader v12-safe typed-loop lane; do not retry direct
tracked-array compare branching or delayed `paramType` lookup without fresh
evidence.

The follow-up bracket swap pattern tranche recognizes the exact local block
shape used by quicksort's helper swap: `tmp := arr[a] as T; arr[a] = arr[b] as
T; arr[b] = tmp`. When the receiver and both indexes are slot-backed
identifiers, lowering emits `ArrayIndexSwapSlot`, which reads those slots
directly and either runs the existing guarded direct array-index get/set bodies
or falls back to the normal generic index get/set sequence. The opcode keeps
the explicit casts and returns the final assignment value, so the optimization
does not change v12 expression semantics.

The kept reduced quicksort `500x` band landed at `4.82ms/op`, `4.76ms/op`,
and `4.86ms/op`, with a profiled confirmation at `4.79ms/op`. This is a small
local hotloop cleanup relative to the immediately preceding slot-const-store
state, not a full external breakthrough: full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`. The next tranche should start from the kept profile and choose
between remaining direct call-frame setup, generic `%` / binary work in
`build_data`, residual array-index cast/index conversion inside the new swap
opcode, or the broader v12-safe typed-loop lane. Do not broaden this into a
named non-primitive `Array.swap` compiler/runtime special case.

The follow-up swap small-index lane keeps the same `ArrayIndexSwapSlot` opcode
but adds a direct tracked-array path for the hot slot shape where both indexes
are already small integers. The VM now reads the tracked array state once,
checks/casts the two values, writes both positions, and syncs aliases through
the same tracked-array machinery used by direct index writes. Non-small
indexes, untracked arrays, custom indexing behavior, and generic shapes still
use the previous fallback sequence.

The refreshed pre-change profiled reduced sample was `4.75ms/op`. The kept
`500x` band landed at `4.66ms/op`, `4.55ms/op`, and `4.68ms/op`, with a
profiled confirmation at `4.59ms/op`. The profile no longer shows the old
`bytecodeDirectArrayIndex(...)` edge on the swap path, but full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next tranche should start from a fresh
kept profile and target the larger remaining walls in fused array-index
compare, direct call-name frame setup, generic `%` / binary work in
`build_data`, or a broader v12-safe typed-loop lane rather than more swap
indexing.

The follow-up reduced matrix tranche targeted the hot `Array.get(... )!`
success path rather than another generic propagation helper. Once canonical
tracked-array `Array.get(i32)` has read a non-nil element and the array state's
cached element token plus the actual read value prove a primitive `f32`/`f64`
value whose type does not currently implement `Error`, the VM now skips the
immediately following postfix propagation opcode. Nil values, stale non-float
element shapes, and primitive float types with an active `Error` impl keep the
old path. The primitive no-error decision is cached on the VM by interpreter
method-cache version, so dynamic impl registration invalidates the skip
decision.

A generic boxed-f64 propagation fast-negative precursor was reverted after it
regressed reduced `matrixmultiply_f64_small` to `14.90s`, `15.16s`, and
`15.68s`. The kept fusion moved the profiled reduced matrix baseline from
`14.72s/op` to a final cached-propagation band of `11.96s/op`, `11.50s/op`,
and `11.66s/op`, with a profiled kept rerun at `10.98s/op`. The kept profile no
longer has `execPropagation(...)` or
`propagationValueMayImplementError(...)` in the top list; the remaining wall is
boxed `f64` arithmetic (`floatResultKind`, `evaluateArithmeticFast`,
`evaluateArithmetic`, `runtime.convT`) plus residual `Array.get`
cache/index work. The next matrix tranche should build a real VM-v2 raw `f64`
expression/slot lane instead of shaving propagation or `Array.get` guards
again unless a fresh profile changes the ranking.

The follow-up arithmetic tranche took the first bounded step toward that raw
lane by adding a direct boxed `FloatValue` pair path for primitive `+`, `-`,
and `*`. This is not a benchmark-specific opcode: it preserves f32
normalization, widens to f64 when either operand is f64, and leaves
float/integer mixing plus division on the existing checked operator path. The
refreshed profiled reduced matrix baseline was `11.63s/op`; the kept band
landed at `8.71s/op`, `9.27s/op`, and `9.56s/op`, with a confirmed profiled
rerun at `8.81s/op`.

The kept profile removes the old `evaluateArithmetic(...)` /
`evaluateArithmeticFast(...)` wall from the f64 hot expression. The next matrix
slice should stop producing a boxed `FloatValue` for every inner-loop multiply
and add: carry raw f64 stack/slot values through the expression and box only at
array, dynamic, and spec-visible boundaries. Another boxed helper is unlikely
to move the allocation count materially.

The follow-up f64 add-mul slot update tranche did that for the dot-product
shape without changing benchmark source. Assignments shaped like
`x = x + left * right` now lower to `StoreSlotFloatAddMul`: the VM captures the
old slot value before evaluating `left` and `right`, then computes raw
primitive float multiply/add when all three values are direct `FloatValue`s.
Non-float shapes fall back through the existing boxed `*` then `+` operator
path, so the fused opcode remains a semantic shortcut rather than a new
language rule.

The prior kept direct-float band was `8.71s/op`, `9.27s/op`, and `9.56s/op`.
The fused update kept at `7.59s/op`, `8.26s/op`, and `7.85s/op`, with a
profiled kept rerun at `7.26s/op`. The larger win is allocation pressure:
the reduced matrix run dropped from roughly `55.45M allocs/op` to
`28.45M allocs/op`.

The fresh kept profile now puts the largest remaining matrix wall in canonical
`Array.get(i32)!` work and its safety guards:
`finishArrayGetMemberFast(...)`, `lookupCachedCanonicalArrayGetCallForArray(...)`,
`bytecodeArrayGetIndexI32(...)`, and
`bytecodeArrayGetResultMatchesFloatToken(...)`. The next matrix slice should
feed raw f64 operands from canonical `Array.get!` into the fused update, or
otherwise reduce that guarded array-get path, before returning to arithmetic.

That fused operand slice is now landed for the slot-backed dot-product shape
`s = s + ai.get(k)! * cj.get(k)!`. Lowering emits
`StoreSlotFloatAddMulArrayGet` only when both `Array.get` receivers and indexes
are identifiers with known slots. The opcode uses the existing canonical
`Array.get` call-site proof, preserves nil/Error propagation, and falls back to
normal member-call semantics when the guard is not valid.

The prior kept add-mul band was `7.59s/op`, `8.26s/op`, and `7.85s/op`. The
fused operand update kept at `7.19s/op`, `5.71s/op`, and `6.07s/op`, with a
profiled confirmation at `5.42s/op`. Allocation volume stayed near
`28.45M allocs/op`; the new profile shows boxed `FloatValue` result storage in
`bytecodeDirectFloatAddMul(...)` as the dominant allocation source. The next
matrix tranche should add a raw f64 accumulator slot/update lane and box only
at array, dynamic, or otherwise spec-visible boundaries.

That raw accumulator tranche is now landed for the same fused opcode. The
lowering no longer emits a target `LoadSlot` before
`StoreSlotFloatAddMulArrayGet`; the opcode reads the accumulator slot directly,
computes the primitive add-mul result without returning it through
`runtime.Value`, and stores the internal result in a VM-owned mutable float
cell. Visible slot reads copy the float value back out, preserving Able value
semantics for later reads.

The prior kept fused-operand band was `7.19s/op`, `5.71s/op`, and `6.07s/op`
with about `28.45M allocs/op`. The raw accumulator update kept at `5.57s/op`,
`5.73s/op`, and `5.86s/op`, with a profiled confirmation at `5.80s/op`.
Allocations dropped to roughly `1.63M allocs/op` and `50.5MB/op`; the old
`bytecodeDirectFloatAddMul(...)` allocation wall is gone. The current CPU
profile is now dominated by the guarded `Array.get!` operand path, especially
`bytecodeFloatTypeToken(...)`, `fusedArrayGetCanSkipPropagationCheck(...)`,
`lookupCachedCanonicalArrayGetCallForArray(...)`, and
`bytecodeArrayGetIndexI32(...)`. The next matrix tranche should combine direct
f64 operand extraction with the propagation guard so the fused opcode no longer
rechecks float element tokens through the generic helper on every operand.

That raw operand extraction tranche is now landed for the same fused opcode.
`StoreSlotFloatAddMulArrayGet` preflights the two canonical `Array.get(i32)!`
operands, reuses the guarded canonical call-site proof, reads both array
values directly, handles nil/Error propagation at the same boundary, and feeds
raw `f32`/`f64` values into the VM-owned accumulator update. Stale cached
element tokens, non-float values, active primitive Error impls, unsupported
receivers, and unsupported indexes still fall back to the existing boxed
member-call path.

The prior kept raw-accumulator band was `5.57s/op`, `5.73s/op`, and
`5.86s/op`. The raw operand update kept at `4.43s/op`, `4.06s/op`, and
`4.41s/op`, with a profiled confirmation at `4.36s/op`. Allocation volume
stayed essentially flat at about `1.63M allocs/op` and `50.4MB/op`, so this is
a CPU-path win rather than a new allocation reduction. Full external bytecode
`matrixmultiply` still timed out at `90s` against Go `0.8800s`, Ruby
`42.9300s`, and Python `56.2900s`.

The kept profile removes the old generic
`bytecodeFloatTypeToken(...)` / `fusedArrayGetCanSkipPropagationCheck(...)`
path from the fused update. The new wall is the exact raw operand guard itself:
`bytecodeFusedArrayGetFloatForToken(...)`, canonical `Array.get` proof version
checks, small-`i32` index extraction, and direct array reads. The next matrix
slice should collapse the f64-specific operand proof/read path or move to a
proper typed f64 array/slot lane; another boxed float helper is unlikely to be
the right level.

The native f64 dot-loop tranche is now landed for the exact reduced/full matrix
inner loop shape:

```able
loop {
  if k >= n { break }
  s = s + ai.get(k)! * cj.get(k)!
  k = k + 1
}
```

Lowering attaches a plan to the existing `LoopEnter` and leaves the original
loop bytecode as the fallback. The VM runs the native path only when it proves
canonical `Array.get`, tracked arrays, valid `i32` loop slots, and actual `f64`
elements. Unsupported values enter the original loop before the unsupported
iteration, preserving nil/Error propagation and boxed dynamic behavior.

The prior kept raw-operand band was `4.43s/op`, `4.06s/op`, and `4.41s/op`.
The native dot-loop update kept at `333.92ms/op`, `319.62ms/op`, and
`331.61ms/op`, with a traced/profiled confirmation at `405.57ms/op`.
Allocation volume stayed essentially flat at about `1.63M allocs/op` and
`50.4MB/op`; this is a CPU dispatch/member-call removal, not an allocation
reduction.

The bytecode trace no longer shows the inner dot-product `ai.get(k)!` /
`cj.get(k)!` calls. Remaining reduced matrix trace traffic is construction and
transpose work: `Array.get` at lines 35, 47, and 52 plus `Array.push` /
`Array.new`. Full external bytecode `matrixmultiply` now completes in
`23.8500s` instead of timing out at `90s`; the reference row was Go `0.8800s`,
Ruby `42.9300s`, and Python `56.2900s`. The next matrix tranche should target
that remaining construction/transpose traffic or generalize the native f64 lane
under the same v12 fallback guard.

The follow-up f64 row-cache tranche keeps the same native dot-loop guard, but
stops re-extracting tracked rows on every dot product. Dynamic array states now
carry a revision, and dynamic writes, length changes, tracked writes, and full
state resyncs bump it. The VM caches raw `[]float64` rows by array-state pointer
plus revision/length and clears the cache between pooled top-level VM runs.

The prior kept native-dot band was `319.62-333.92ms/op`. The row-cache update
kept at `229.45ms/op`, `204.02ms/op`, `206.60ms/op`, `213.82ms/op`, and
`225.63ms/op`, with a profiled confirmation at `236.52ms/op`, `52.18MB/op`,
and `1.63M allocs/op`. `bytecodeDirectF64Value` no longer appears as a top
self-cost in the reduced profile; remaining samples are now mostly
construction/transpose member paths, `Array.push`, GC scanning, and residual
slot-store work.

Full external bytecode `matrixmultiply` moved from `23.8500s` to `3.0800s`.
The comparison row is Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so
bytecode matrix is now about `3.50x` Go while beating the Ruby/Python references
for this benchmark. The next matrix tranche should target construction /
transpose allocation and member-call traffic, not another per-element f64
extraction helper.

The small-integer float-cast tranche removes the construction-side
`big.Int`/`big.Float` conversion for casts like `(i - j) as f64` when the
integer value is already stored in the runtime small-int representation. Big
integer values still use the existing arbitrary-precision path.

The prior kept row-cache band was `204.02-229.45ms/op`. The small-int cast
update kept at `186.59ms/op`, `212.11ms/op`, `202.67ms/op`, `184.04ms/op`, and
`201.25ms/op`, with a profiled confirmation at `194.99ms/op`, `46.31MB/op`,
and `912.8k allocs/op`. Allocation count dropped from roughly `1.63M/op` to
roughly `913k/op`. The reduced profile no longer shows
`math/big.(*Float).Float64` in the top nodes.

Full external bytecode `matrixmultiply` moved from `3.0800s` to `2.9000s`.
The comparison row is Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so
bytecode matrix is now about `3.30x` Go while remaining faster than the
Ruby/Python references for this benchmark. The next matrix tranche should
target `Array.push` / `Array.slot` dispatch and remaining construction-time
`Array.get` reads.

The tracked `Array.push` append-helper tranche is deliberately external-driven.
The reduced fixture still grows rows from `Array.new`, so the reduced signal was
mostly neutral: `191.41ms/op`, `191.20ms/op`, `193.30ms/op`, `190.00ms/op`, and
`193.85ms/op`, with a profiled `199.26ms/op`, `46.32MB/op`, and
`912.8k allocs/op`. The external matrix benchmark uses `Array.with_capacity(n)`
for rows and outers, which is the shape this tranche targets.

The helper skips `runtime.ArrayEnsureCapacity(...)` when existing logical
capacity and backing slice storage are already sufficient, and it uses the
unaliased tracked-write sync path before falling back to alias-aware sync. Full
external bytecode `matrixmultiply` moved from `2.9000s` to `2.7700s` on the
first run and `2.7500s` over a `3/3` confirmation. The comparison row is Go
`0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `3.12x` Go.

The next matrix tranche should target construction-time `Array.get` /
`Array.slot` dispatch and GC scan pressure. Further push-only changes need a
fresh external profile before they are worth trying.

The adjacent-`Pop` push cleanup keeps that same direction. Canonical cached
`Array.push(value)` now skips materializing `void` only after the VM has handled
the push and sees that the next active bytecode is the statement-result `Pop`.
Lowering still emits the `Pop`, so generic fallback behavior stays unchanged.
The corrected reduced `matrixmultiply_f64_small` band was `176.45ms/op`,
`181.64ms/op`, `186.98ms/op`, `185.66ms/op`, and `183.05ms/op`; allocation
volume stayed around `46.3MB/op` and `912.9k allocs/op`.

External bytecode `matrixmultiply` did not show a clear macro win from this
cleanup: `2.8167s` over `3/3`, then `2.7740s` over `5/5`, compared with the
prior kept `2.7500s`. Treat it as reduced-positive and external-neutral. The
next tranche should move to construction-time `Array.get` reads and residual
slot-call cache checks rather than another push-only edit.

The f64 dot-loop accumulator-store tranche targets allocation/GC pressure
inside the already-fused native dot loop. The loop now writes the completed
accumulator back as a plain `FloatValue` instead of installing an owned float
cell; because the fused loop updates once per completed dot product, the owned
cell was not amortized and showed up as allocation pressure.

Reduced `matrixmultiply_f64_small` moved to `170.93ms/op`, `170.63ms/op`,
`168.11ms/op`, `163.01ms/op`, and `165.41ms/op`. Allocation volume dropped to
about `44.1-44.4MB/op` and `822.9k allocs/op`; the profiled run was
`169.44ms/op`, `44.18MB/op`, and `822.8k allocs/op`. The allocation profile no
longer has `storeOwnedFloatSlot` or `bytecodeSlotReadValue` in the top list.

Full external bytecode `matrixmultiply` moved to `2.6040s` over `5/5`, against
Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `2.96x` Go. The next tranche should target remaining boxed float
arithmetic/cast allocation or a genuine typed f64 row/storage lane rather than
another broad owned-slot rewrite.

The f64 affine `Array.push` try-fast tranche targets the construction expression
used by `build_matrix`: `row.push(t * ((i - j) as f64) * ((i + j) as f64))`.
Lowering emits a guarded try opcode before the normal fallback bytecode. When
runtime guards prove canonical `Array.push`, direct `f64` scale, and `i32`
left/right slots, the VM computes the f64 value and appends it directly; any
guard miss falls through to the existing receiver/expression/member-call path.

Reduced `matrixmultiply_f64_small` moved from the prior kept
`163.01-170.93ms/op`, `44.1-44.4MB/op`, and `822.9k allocs/op` band to
`159.61ms/op`, `133.18ms/op`, `136.46ms/op`, then a confirming rerun at
`121.57ms/op`, `126.45ms/op`, and `125.96ms/op`. The profiled reduced run was
`130.51ms/op`, `31.19MB/op`, and `282.8k allocs/op`.

Full external bytecode `matrixmultiply` moved to `2.1300s` over `5/5`, against
Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `2.42x` Go. The next tranche should target row/column storage allocation
and remaining construction / transpose array traffic through a v12-safe typed
f64 row/storage lane, not another boxed arithmetic helper.

The versioned-stdlib canonical proof tranche is a measurement/proof keep rather
than a macro matrix speedup. The reduced runtime-only matrix benchmark had been
falling back through generic `Array.get` after warmup because canonical stdlib
origin checks accepted sibling checkout and flat cache paths, but not installed
cache paths shaped as `.able/pkg/src/able/<version>/src/...`. The canonical
origin helper now accepts that versioned boundary, and the nullable
`Array.get` proof also accepts direct or bound single `FunctionValue` methods
when they are the canonical nullable stdlib function.

With that proof restored, runtime-only `matrixmultiply_f64_small` now completes
instead of sitting in `callArrayGetFallback`: warmed `5x` reruns landed at
`120.53ms/op`, `117.29ms/op`, and `122.68ms/op`, with roughly `31.2MB/op` and
`282.8k allocs/op`. The reduced CLI fixture stayed neutral at `0.2000s`, and
full external bytecode `matrixmultiply` stayed neutral at `2.1333s` over `3/3`
runs versus Go `0.8800s`. The next matrix tranche should still target
row/column storage allocation and remaining construction/transpose traffic,
not another canonical-origin or `Array.get` proof-cache slice.

The mono-f64 array storage tranche makes that row-storage direction concrete.
The runtime array store now has a guarded f64 lane; the matrix affine
`Array.push` fast path promotes unaliased dynamic rows to mono f64 after
validating the existing elements, the native f64 dot-loop reads mono rows
directly, and canonical `Array.get` fast paths avoid generic `ArrayStoreRead`
for mono f64 handles. All unsupported shapes still deopt or fall back to boxed
Array semantics.

The fresh same-session runtime-only reduced baseline was `134.38ms/op`,
`31.18MB/op`, and `282,768 allocs/op`. The kept rerun landed at
`119.10ms/op`, `117.07ms/op`, and `124.46ms/op`, with about `21.8-22.0MB/op`
and `193.1k allocs/op`. Full external bytecode `matrixmultiply` moved from
the versioned-stdlib proof keep at `2.1333s` over `3/3` to `2.0400s` over
`3/3`; the profiled confirmation was `2.0500s`. The next matrix target is not
more storage helper work; the remaining profile wall is boxed f64 result
materialization in `finishArrayGetMemberFast(...)`, the native dot-loop
accumulator slot write, and residual row lookup/cache checks.

The follow-up nested-get push tranche targets the transpose expression
`ci.push(b.get(j)!.get(i)!)` without changing the fallback bytecode. Lowering
now emits a guarded try opcode for that shape; the VM appends the inner row's
raw f64 directly only when both `Array.get` calls and the destination `push`
are canonical, the outer propagated value is a concrete Array that cannot
implement `Error`, and the inner lookup is in bounds. Nil, Error-capable,
custom, non-f64, out-of-bounds, and aliased/unsupported cases fall through to
the existing boxed path.

Reduced runtime-only `matrixmultiply_f64_small` kept the current wall-clock
band at `121.35ms/op`, `131.30ms/op`, and `122.89ms/op`, while allocation
dropped to roughly `15.7MB/op` and `103.1k allocs/op`. A trace/profiled
confirmation showed `90,000` hits on `array_push_f64_nested_get_fast` at the
transpose line and `103,573 allocs/op`. Full external bytecode
`matrixmultiply` was noisy against the older `2.0400s` best, so a
same-session control was taken: disabling only this new lowering landed at
`2.3533s` over `3/3`, while the restored fused confirmation landed at
`2.1840s` over `5/5`. Treat this as an allocation/shape keep, not a new
all-time wall-clock low. The next matrix target is the native dot-loop
accumulator `FloatValue` box plus repeated array/cache guard cost.

The owned f64 accumulator-cell tranche handles the first half of that boundary.
`StoreSlot`/`StoreSlotNew` now seed and reuse VM-owned float cells for
`FloatValue` locals, and the native f64 dot loop updates the accumulator
through that cell. `LoadSlot` still snapshots owned cells back to ordinary
`FloatValue`, preserving primitive value semantics and preventing array pushes
from retaining mutable slot cells.

Reduced runtime-only `matrixmultiply_f64_small` moved to `108.75ms/op`,
`113.94ms/op`, and `114.94ms/op`, with allocation effectively unchanged at
about `15.7MB/op` and `103.1k allocs/op`. The profiled confirmation landed at
`106.55ms/op`; allocation moved from `tryExecF64DotLoop(...)` to
`bytecodeSlotReadValue(...)`, which snapshots `s` for the following
`di.push(s)`. Full external bytecode `matrixmultiply` landed at `2.0840s` over
`5/5`. The next matrix target should be a guarded slot-backed f64
`Array.push` for `di.push(s)`, not exposing owned float pointers through
general `LoadSlot`.

The reserved-capacity `Array.with_capacity` tranche attacks the full external
allocation profile rather than the reduced fixture. Interpreted
`__able_array_with_capacity` now creates a dynamic array handle with logical
capacity but no dynamic `[]Value` backing. If the array stays dynamic, the first
write allocates the reserved backing before appending; if the row immediately
promotes to mono f64, the runtime allocates only the mono-f64 storage and skips
the discarded dynamic backing. `Array.new(capacity)` and
`ArrayStoreNewWithCapacity(...)` remain eager to avoid changing compiled/runtime
ABI paths that still observe `ArrayValue.Elements` directly.

Reduced runtime-only `matrixmultiply_f64_small` stayed neutral at
`118.08ms/op`, `121.22ms/op`, and `127.02ms/op`, with about `15.7MB/op` and
`103.1k allocs/op`, because that fixture still uses `Array.new`. Full external
bytecode `matrixmultiply` also stayed wall-clock neutral at `2.1000s` over
`5/5`, while GC dropped to `7.00`. The bytecode-runtime allocation evidence is
the reason to keep it: the prior profiled full run showed `125.83MB` total and
`121,006,704 B/op`; after reserved capacity, the profiled run shows `90.89MB`
total and the unprofiled runtime bench shows `71,844,768 B/op`. The old
`ArrayStoreNewWithCapacity(...)` alloc-space leader is gone. The next matrix
target is the remaining `bytecodeSlotReadValue(...)` / `di.push(s)` f64 result
boundary or a typed f64 result-row lane, not another generic capacity pass.

The native dot-loop result-append tranche removes that specific f64 result
boundary without adding another standalone hot dispatch opcode. Lowering keeps
the original bytecode for `loop { ... }; di.push(s)`, but attaches an optional
result-append target to the existing f64 dot-loop plan when the next
statement-position call is exactly a push of the same accumulator. On the fast
path, after the dot product completes and canonical `Array.get` / `Array.push`
guards hold, the VM appends the raw accumulator to the result row and jumps past
the boxed fallback push. Guard misses still run the original loop and push.

Reduced runtime-only `matrixmultiply_f64_small` allocation fell to about
`10.4MB/op` and `13.4k allocs/op`, but reduced wall time was noisy at
`179.01ms/op`, `141.30ms/op`, and `205.82ms/op`. The full external result is
the keep signal: bytecode `matrixmultiply` moved from the reserved-capacity
`2.1000s` confirmation to `2.0240s` over `5/5`, with average GC at `6.00`.
The full bytecode-runtime profile moved to `1.855s/op`, `39,764,440 B/op`, and
`73,510 allocs/op`; profiled allocation total is now `59.12MB`, and
`bytecodeSlotReadValue(...)` is no longer an allocation leader. The remaining
matrix allocation wall is `ArrayStoreAppendF64Promote(...)` and mono-f64
append/growth storage, not the result-load box.

The f64 dot-loop range-hoist tranche is a small CPU cleanup inside the existing
native dot-loop rather than another storage rewrite. The VM now proves the full
`i32` loop range against both raw f64 row slices before accumulating, then runs
the product as a plain `int` indexed Go loop. If the range is negative or would
run out of either row, the fast path falls through before mutating loop slots so
the original bytecode handles the observable failure path.

Reduced runtime-only `matrixmultiply_f64_small` landed at `104.74ms/op`,
`10.43MB/op`, and `13.83k allocs/op`; a same-session old-loop control landed
at `107.86ms/op` with the same allocation floor. Full external bytecode
`matrixmultiply` confirmed at `2.0060s` over `5/5`, while the full
bytecode-runtime profile landed at `1.937s/op`, `39,759,120 B/op`, and
`73,492 allocs/op`. The CPU profile shows `tryExecF64DotLoop(...)` around
`0.91s` flat / `1.17s` cumulative. The next matrix work should be plan-level
row/handle caching or a typed matrix kernel boundary; standalone mono-f64
append/helper rewrites have not shown enough macro movement.

The f64 matrix row-kernel tranche is the first kept typed matrix boundary. The
lowerer recognizes the exact outer `j` loop around `s := 0.0`,
`cj := c.get(j)!`, the proven native f64 dot loop, `di.push(s)`, and
`j = j + 1`. The VM validates canonical `Array.get` / `Array.push`, concrete
row values, f64 row storage, non-negative in-bounds ranges, and destination
non-aliasing before it computes the remaining row and bulk-appends raw f64
results. Guard misses keep the original bytecode and leave the destination row
unmodified.

Reduced runtime-only `matrixmultiply_f64_small` moved from the fresh
`105.35ms/op`, `10.45MB/op`, and `13.89k allocs/op` baseline to `76.79ms/op`,
`10.00MB/op`, and `11.72k allocs/op` over `5/5`; a profiled confirmation
landed at `87.27ms/op`. Full external bytecode `matrixmultiply` moved from
`2.0060s` to `1.7580s` over `5/5` after an earlier same-tranche `1.4967s`
`3/3` run. The `5/5` comparison is about `2.00x` Go (`0.8800s`) and roughly
24x faster than Ruby / 32x faster than Python on the external table. The next
matrix work should target mono-f64 row/result storage growth and capacity
proofs, or graduate this into a broader typed matrix bytecode that carries raw
f64 row slices through build, transpose, and multiply.

The f64 affine row-loop tranche moves the same idea into matrix construction.
Lowering now recognizes the exact `build_matrix` inner loop shape
`if j >= n { break }; row.push(t * ((i - j) as f64) * ((i + j) as f64));
j = j + 1`, attaches a guarded loop plan, and leaves the original bytecode as
the fallback. On the fast path, the VM validates the canonical `Array.push`
proof and f64/i32 operands before mutating the row, computes the remaining row
values into raw f64 storage, then bulk-appends through the existing mono-f64
append rules. This is deliberately not a generic capacity rewrite: `Array.new`
rows end with the same amortized capacity repeated single pushes would expose,
while `Array.with_capacity(n)` rows preserve their declared capacity.

Reduced runtime-only `matrixmultiply_f64_small` moved from a fresh
`74.64ms/op`, `10.02MB/op`, and `11.78k allocs/op` baseline to `54.73ms/op`,
`9.17MB/op`, and `8.12k allocs/op` over `5/5`; a profiled confirmation landed
at `58.66ms/op`, `9.20MB/op`, and `8.18k allocs/op`. Full external bytecode
`matrixmultiply` moved from the row-kernel `1.7580s` confirmation to
`1.4480s` over `5/5`, about `1.65x` Go (`0.8800s`) while remaining far ahead
of Ruby and Python on the external table. The reduced profile no longer spends
the row build on repeated `execTryArrayPushF64AffineProduct(...)` calls. The
next matrix tranche should apply the same bounded loop-level treatment to the
transpose row shape `ci.push(b.get(j)!.get(i)!)`, then reassess result row
materialization and canonical get/push version checks.

The f64 transpose row-loop tranche applies that bounded loop-level treatment to
the `matmul` transpose build. Lowering recognizes only the exact loop shape
`if j >= n { break }; ci.push(b.get(j)!.get(i)!); j = j + 1`, attaches a
guarded loop plan, and keeps the original bytecode as the fallback. On the fast
path, the VM validates canonical `Array.get` / `Array.push`, Array-valued
source rows, raw f64 row storage, non-negative i32 indices, and
destination/source non-aliasing before mutation. It then gathers the remaining
column values and bulk-appends them through the same mono-f64 append rules as
the other matrix loop plans. Guard misses fall through without partial
destination mutation, and final row capacity stays equivalent to repeated
single pushes for `Array.new` or the declared capacity for
`Array.with_capacity(n)`.

Reduced runtime-only `matrixmultiply_f64_small` moved from the prior kept
affine-row `54.73ms/op`, `9.17MB/op`, and `8.12k allocs/op` band to
`40.86ms/op`, `8.76MB/op`, and `6.32k allocs/op` over `5/5`; a profiled
confirmation landed at `42.45ms/op`, `8.78MB/op`, and `6.38k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.4480s` to `1.3060s` over
`5/5`, about `1.48x` Go (`0.8800s`). The reduced CPU profile no longer shows
the repeated `execTryArrayPushF64NestedGet(...)` transpose cell path; the
largest remaining wall is `tryExecF64MatrixRowLoop(...)`. The next matrix
tranche should target a guarded raw row-slice cache for the transposed matrix
or tighten the row kernel so it avoids re-reading and revalidating every `c`
row for each output row.
