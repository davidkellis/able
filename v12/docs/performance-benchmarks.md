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
