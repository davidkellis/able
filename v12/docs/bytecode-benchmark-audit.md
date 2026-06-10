# Bytecode Benchmark Audit

This document records the current benchmark-breadth audit for speculative
bytecode opcodes. The purpose is to keep bytecode optimization work grounded in
generic workload patterns rather than in single-benchmark stencils.

## Policy

- Do not add long bespoke benchmark-only opcodes.
- Prefer generic improvements to lowering, slot/value representation, dispatch,
  quickening, or reusable opcode families.
- Before keeping a new speculative opcode, check whether the shape recurs
  across a broader benchmark set with `v12/bench_bytecode_audit`.

## Tooling

- List benchmark suites:
  - `./v12/bench_bytecode_audit --list-suites`
  - `./v12/bench_compare_external --list-suites`
- Statically validate the complete portable/local catalog without timing:
  - `just bench-catalog-check`
  - `just bench-catalog-check --suite dependency-plan` validates one named
    portable suite without running it; every name accepted by
    `bench_compare_external --list-suites` is accepted here too
  - each named portable suite is also checked as a subset of `coverage`, the
    authoritative application population for broad lowering audits
- Run the broad lowering audit:
  - `./v12/bench_bytecode_audit --suite corpus-full`
- Run a broad external comparison only after performance selection reopens:
  - `./v12/bench_compare_external --suite coverage --modes compiled,bytecode --cpu-affinity CPU --require-quiet-cpu`

Current suite presets:

- `core`: `fib,binarytrees,matrixmultiply,quicksort,sudoku,i_before_e`
- `generality` / `full`:
  `fib,binarytrees,matrixmultiply,quicksort,sudoku,sudoku_masks,i_before_e,base64,json,monte_carlo_pi,pidigits,mandelbrot,reverse_complement,k_nucleotide,nbody,tapelang_alphabet`
- `numeric-structural`:
  `fib,binarytrees,matrixmultiply,mandelbrot,monte_carlo_pi,pidigits,nbody`
- `fixed-width`:
  `fixed_width_128`
- `nominal-numeric`:
  `fixed_width_128,rational_series,wide_integer_records`
- `wide-integer-records`:
  `wide_integer_records`
- `text-bytes`:
  `base64,json,i_before_e,reverse_complement,k_nucleotide,quicksort,tapelang_alphabet`
- `sudoku-masks`:
  `sudoku_masks`, the bounded bit-mask/backtracking workload in `full` /
  `generality` that shares the canonical Sudoku corpus without replacing the
  historical `sudoku` lane
- `log-routing-redaction`:
  `log_routing_redaction`, the portable four-language log classification and
  privacy-redaction application using public `RegexSet` and `Regex.replace`
- `config-validation-extraction`:
  `config_validation_extraction`, the portable four-language deployment
  configuration validator using public Regex matching and capture extraction
- `fixture-core`:
  `array_map_i32_small,channel_roundtrip_i32_small,future_fanout_i32_small,graph_bfs_small,hash_set_i32_small,md5_hex_small,queue_i32_small,regex_is_match_small,string_builder_small,union_find_small`
- `fixture-full`: all `v12/fixtures/bench/*` directories (61 benchmarks as of
  2026-07-10; 79 current fixtures; use `just bench-catalog-check` for the
  current catalog rather than a historical count)
- `fixture-generality`:
  `array_map_i32_small,channel_roundtrip_i32_small,deque_i32_small,dijkstra_heap_small,future_fanout_i32_small,heap_i32_small,nbody_small,persistent_set_i32_small,persistent_sorted_set_i32_small,random_lcg_i64_small,regex_is_match_small,string_builder_small,sum_u32_small,word_count_small,zigzag_char_small`
- `fixture-collections`:
  `array_filter_i32_small,array_fold_i32_small,array_map_i32_small,bit_set_small,concurrent_queue_i32_small,deque_i32_small,hash_set_i32_small,hashmap_i32_small,heap_i32_small,lazy_seq_cache_i32_small,lazy_seq_take_i32_small,linked_list_enumerable_i32_small,linked_list_for_i32_small,linked_list_iterator_collect_i64_small,linked_list_iterator_filter_map_i64_small,linked_list_iterator_pipeline_i64_small,list_i32_small,persistent_queue_i32_small,queue_i32_small,vector_i32_small`
- `fixture-text`:
  `ascii_lower_small,automata_dfa_small,base64_roundtrip_small,i_before_e_small,json_means_small,k_nucleotide_small,md5_hex_small,regex_is_match_small,reverse_complement_small,string_builder_small,string_contains_small,string_split_join_small,tapelang_small,zigzag_char_small`
- `fixture-algorithms`:
  `binarytrees_small,fib_i32_small,graph_bfs_small,mandelbrot_small,matrixmultiply_f64_small,monte_carlo_pi_small,nbody_small,pidigits_small,quicksort_file_small,sieve_count,sieve_full,sudoku_file_small,sum_u32_small,union_find_small`
- `fixture-concurrency`:
  `await_batch_i64_small,binarytrees_small,channel_pipeline_i32_small,channel_roundtrip_i32_small,concurrent_queue_i32_small,future_fanout_i32_small,future_yield_i32_small,mutex_counter_i32_small`
- `fixture-numeric`:
  `bigint_add_mul_small,bigint_ref_newton_small,biguint_add_mul_small,fib_i32_small,int128_accumulate_small,mandelbrot_small,matrixmultiply_f64_small,monte_carlo_pi_small,nbody_small,pidigits_small,random_lcg_i64_small,rational_series_small,sum_u32_small,uint128_accumulate_small`
- `fixture-external-small`:
  `base64_roundtrip_small,binarytrees_small,i_before_e_small,json_means_small,k_nucleotide_small,mandelbrot_small,monte_carlo_pi_small,nbody_small,pidigits_small,quicksort_file_small,reverse_complement_small,sudoku_file_small,tapelang_small`
- `corpus-full`:
  all 41 portable external `coverage` applications plus `fixture-full` (120
  programs total in the current catalog)
- `option-result-config`:
  `option_result_config`, a deployment-capacity reconciliation application
  that uses generic named-union fallback, mapping, validation, and recovery
  without making a synthetic combinator loop
- `channel-rollup`: `channel_rollup`, an external-style buffered channel
  producer/worker/consumer application over the checked-in word list
- `concurrency`:
  `binarytrees,channel_rollup,future_pipeline,future_await_race,await_channel_mux,mutex_ledger,mutex_await_journal,mutex_work_queue`; local channel, Future, and Mutex coverage remains in `fixture-concurrency`

The current `corpus-full` lowering audit passes with 120 programs, 460 lowered
functions, and 21,955 instructions. This is static breadth evidence, not a
timing result or permission to retain a benchmark-shaped opcode.

The 2026-07-10 feature coverage audit is recorded in
`docs/perf-baselines/2026-07-10-feature-benchmark-coverage.md`. It adds
Channel-Rollup as the first ordinary cross-language channel application while
keeping reference timing publication separate from the dirty shared results
ledger.

## 2026-06-23 Generality Audit

Command:

`./v12/bench_bytecode_audit --suite generality`

Summary:

- benchmarks lowered: `15`
- functions lowered: `78`
- instructions lowered: `4118`

Tracked opcode coverage:

- `JumpIfFloatMulAddMulCompareConstFalse`: appears in `mandelbrot` and
  `monte_carlo_pi`
- `TryArrayPushF64AffineProduct`: appears in `matrixmultiply` only
- `TryArrayPushF64NestedGet`: appears in `matrixmultiply` only
- `TryFloatUpdatePair`: appears in `mandelbrot` only
- `StoreSlotFloatAddMulSlot`: appears in `mandelbrot` only
- `StoreSlotFloatAddSub`: appears in `mandelbrot` only
- `JumpIfFloatAddCompareConstFalse`: does not appear in the current broad suite
- `tapelang_alphabet`: adds a parser plus tape-machine dispatch workload and
  does not currently use any of the tracked speculative opcodes

Implication:

- the only currently-audited float/control-flow shape that recurs across more
  than one benchmark is `JumpIfFloatMulAddMulCompareConstFalse`
- the remaining tracked float and array-specialized opcodes are still
  single-benchmark shapes in the current benchmark corpus
- future bytecode work should bias toward broader mechanisms unless a
  single-benchmark opcode later shows up in additional workloads

## 2026-06-24 Generality Follow-Up

The first post-audit cleanup tranche kept moving in the same direction:
matrix-only lowering hooks were backed out instead of being preserved as
benchmark-shaped keeps. The resulting `matrixmultiply_f64_small` bytecode
regression turned out not to need a new matrix specialization at all. The real
bug was generic lowering correctness:

- `bytecodeBinaryCastSlotFloatConstInstructionForOperator(...)` was
  incorrectly accepting both operand orders for `/`
- that let `const / (slot as f64)` lower into the existing fused
  `(slot as f64) / const` opcode, silently changing program meaning for
  left-associated scale expressions such as
  `1.0 / (n as f64) / (n as f64)`
- the fix was to keep only the operand order the opcode actually implements
  and leave mirrored division on generic lowering

Implication:

- broadened benchmark coverage is paying for itself as a correctness audit, not
  only as a performance audit
- when a regression appears after removing a single-benchmark fusion, the next
  move should be to look for a generic lowering/runtime bug before inventing a
  replacement specialization
- `matrixmultiply_f64_small` should remain in the bounded validation slice for
  any future float-lowering changes that touch casts, division, or nested array
  construction

The same corpus is also useful for exposing reusable stdlib algorithm hotspots.
A later bounded `fixture-text` sweep showed `automata_dfa_small` timing out in
bytecode not because the VM needed a new automata-specific opcode, but because
the canonical stdlib DFA matcher was linearly scanning every transition for
every character. Sorting DFA transitions once at build time and binary
searching the state/symbol range inside `../able-stdlib/src/text/automata.able`
cut the direct bytecode run from about `21.25s` to `7.96s` and brought the
fixture back under the `20s` bounded validator budget without any compiler or
VM specialization. The remaining heavier text timeout after that tranche is
`zigzag_char_small`. A same-turn bounded `fixture-algorithms` sweep under
`compiled,bytecode` then came back clean for all 18 workloads, so the current
bounded outlier across those two broader suites is `zigzag_char_small` rather
than an algorithms benchmark.

The next `zigzag_char_small` follow-up was also useful because it separated two
different issues that could easily have been conflated. A broad primitive-array
tranche landed first:

- mono `char` storage was added beside the existing primitive mono carriers
- bytecode array push now has a general `char` promotion path
- exact-native `__able_array_read` now has a direct bytecode fast path instead
  of always paying full native-call setup overhead

Those are generally-applicable runtime/VM improvements and they are now
covered by focused tests plus a loader-backed reduced zigzag trace. That trace
showed the source workload now reaches `array_push_char_mono_fast`, which means
the benchmark is no longer primarily blocked on boxed `char` carriers.

However, the full bounded `zigzag_char_small` run still times out at `60s`.
The remaining hot shape in the reduced trace is canonical array indexing:
repeated `array_len_fast` checks followed by exact-native `__able_array_read`.

Implication:

- the next generic bytecode tranche for this workload should target canonical
  `Index.get` / array-indexing quickening or lowering cleanup
- do not answer this benchmark with a zigzag-only opcode, row-pattern fusion,
  or any other single-workload stencil

That next generic tranche has now landed as well:

- direct bytecode `Index.get` / `IndexSet` fast paths now keep mono `char` and
  `u8` arrays on handle-backed storage instead of materializing boxed state
  when the source uses canonical `arr[idx]` forms
- successful canonical index results now skip a following propagation opcode
  whenever the concrete runtime value cannot implement `Error`, which removes
  generic success-side `...!` overhead from hot `arr[idx]!` and
  `arr.get(idx)!` shapes
- a reduced-zigzag profile after that keep showed the remaining generic wall
  much more clearly:
  - stdlib-backed index lookup was still materializing mono arrays during type
    inspection
  - the bytecode index-method cache only keyed primitive element tokens, so
    canonical indexing on nominal arrays like `Array (Array char)` still paid
    full generic `Index.get` call/type-binding cost
- the next keep addressed those broader runtime issues instead of adding any
  new zigzag-specific lowering:
  - added mono-handle element-type recovery so array type inference can see
    `char`, `u8`, `f64`, `i32`, `i64`, and `bool` without deopting to boxed
    array state
  - fixed shared `typeExpressionToString(...)` handling for result types like
    `!void`, which also made canonical `IndexMut.set` recognition structurally
    correct
  - extended canonical array-index quickening from primitive-only element
    tokens to a general element-type cache key, which lets `Array (Array char)`
    and other nominal array workloads reach the same reusable direct
    `Index.get` / `IndexMut.set` path

Measured result:

- reduced zigzag runtime:
  - before targeted profiling follow-up:
    about `9.80s`, `2.23GB/op`, `48.37M allocs/op`
  - after mono-handle type inference + canonical primitive index quickening:
    about `5.12s`, `1.04GB/op`, `22.01M allocs/op`
  - after nominal array element-type cache identities:
    about `0.358s`, `16.1MB/op`, `631k allocs/op`
- the reduced source trace no longer reports hot kernel `__able_array_read`
  fallback calls; dominant traced work is now ordinary `array_push_*_fast`
  plus `array_len_fast`
- the bounded full bytecode workload
  `zigzag_char_small` now completes in about `9.9s` instead of timing out at
  `60s`
- a bounded `fixture-text` bytecode sweep now comes back clean for all 17 text
  workloads under the same `60s` guard

Updated implication:

- `zigzag_char_small` is no longer the bounded text blocker
- the next follow-up should move to the new heaviest text fixtures
  (`string_builder_small`, `automata_dfa_small`, `string_contains_small`,
  `word_count_small`) or another broader corpus sweep
- continue rejecting any benchmark-specific opcode or fusion that is justified
  only by a single workload

## 2026-06-25 Canonical Numeric Struct + `Random` Fast Paths

The next bounded follow-up stayed on the same generality rule and targeted the
new post-array-sync leaders with CPU profiles first:

- `int128_accumulate_small` and `uint128_accumulate_small` were dominated by
  repeated canonical stdlib method dispatch, struct construction, and
  primitive conversion churn around `Int128` / `UInt128`
- `random_lcg_i64_small` showed the same broader pattern on
  `able.random.Random`: repeated tiny canonical methods paying full member
  lookup/call overhead
- `sum_u32_small` did not share that shape; it stayed mostly about integer
  boxing/casting plus array/index traffic, which made it a good control case

The kept fix therefore extended the existing canonical stdlib member fast-path
family instead of adding any new opcode or lowering rule:

- canonical bytecode member fast-path detection now recognizes
  `able.random.Random`, `able.numbers.int128.Int128`, and
  `able.numbers.uint128.UInt128`
- direct static/member fast paths now cover the hot constructor, state, and
  arithmetic methods used by those canonical stdlib structs
- edge cases still fall back to the generic path when semantics require it,
  such as out-of-range numeric conversion methods or negative
  `UInt128.from_i64(...)`

Measured result from the bounded bytecode spot-check:

- `int128_accumulate_small`: `11.30s` -> `1.32s`
- `uint128_accumulate_small`: `10.22s` -> `1.44s`
- `random_lcg_i64_small`: `5.49s` -> `2.40s`
- `sum_u32_small`: `6.25s` -> `5.25s`

Implication:

- this keep is still general VM work: canonical stdlib struct dispatch is now
  much cheaper without introducing any benchmark-only arithmetic stencil
- the remaining leaders are now pointing more clearly at generic integer
  boxing/cast/index costs rather than at more numeric-stdlib dispatch
- the next follow-up should therefore target `sum_u32_small`,
  `nbody_small`, `persistent_sorted_set_i32_small`, `word_count_small`, and a
  bounded full refresh instead of returning to `Int128` / `UInt128` /
  `Random` dispatch

## 2026-06-25 Mono `u32` / `u64` Array Tranche

The next follow-up stayed on that same generality rule and moved to the
remaining unsigned-array gap rather than adding any `sum_u32_small`-specific
opcode or lowering rule:

- the missing shared runtime surface was primitive-array storage itself:
  - mono array carriers existed for `i32`, `i64`, `bool`, `char`, `u8`, and
    `f64`, but not for `u32` or `u64`
  - generic `Array.with_capacity(...)` therefore still allocated boxed dynamic
    storage for `Array u32` / `Array u64`
  - bytecode direct reads and canonical `Array.push` had no matching
    mono-unsigned fast path family
- the landed keep was fully general:
  - added runtime mono-array support for `u32` / `u64` across handle
    allocation, read/write, reserve, clone, deopt, and type metadata
  - taught generic array construction to allocate mono unsigned handles when
    `T` is `u32` / `u64`
  - added bytecode mono-unsigned fast reads for direct index, canonical
    `get`, slot reads, and exact-native `__able_array_read`
  - added mono-unsigned canonical `Array.push` fast paths so repeated unsigned
    appends stay on the primitive carrier

Measured result from bounded spot-checks:

- post-runtime-only `sum_u32_small` moved the wrong way to about `6.13s` over
  `2/2`, which exposed append conversion cost as the next shared wall
- after the general mono-unsigned append fast path landed,
  `sum_u32_small` improved to about `5.81s` over `2/2`
- `nbody_small` stayed effectively flat at `5.90s` versus the prior `5.92s`

Implication:

- the unsigned primitive-array storage path is now structurally present and no
  longer the main missing runtime family
- the remaining `sum_u32_small` gap is still broader unsigned
  arithmetic/cast/index loop overhead rather than array materialization alone
- the next follow-up should profile or instrument the shared unsigned loop
  path and target generic coercion/update/index costs that should transfer to
  adjacent integer-heavy workloads

## 2026-06-25 Raw Small-Integer VM Lane

The next follow-up stayed on that same generality rule and started with a
bounded CPU profile of `sum_u32_small` instead of another source-shaped guess.

Profile result:

- the earlier mono-unsigned array storage/read/push work had already dropped
  out of the hot tier
- the dominant shared cost had moved into VM integer representation work:
  `bytecodeBoxedIntegerValue(...)`, canonical integer casts, small-integer
  binary fast paths, and slot/index churn

The kept fix therefore stayed inside the VM representation rather than adding
any new workload-shaped lowering or opcode:

- added raw small-integer stack/slot carriers for supported primitive integer
  kinds (`i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`, `isize`,
  `usize`)
- taught generic slot loads/stores, exact typed integer stores, and typed
  integer casts to preserve those values in raw form inside the bytecode VM
- kept specialized small-integer arithmetic results raw until a real runtime
  boundary is crossed
- materialized raw values only at shared interpreter/runtime boundaries such
  as overload selection, generic calls, returns, and runtime array writes
- widened canonical array member/index helpers to accept raw integer indices
  so control workloads like `nbody_small` do not fall back to generic overload
  dispatch

Measured result from bounded spot-checks:

- `sum_u32_small`: about `5.81s` -> `3.80s`
- `nbody_small`: about `5.90s` -> `5.99s`

Implication:

- this is a generally-applicable primitive representation keep, not a
  benchmark-only fusion
- the unsigned loop wall is much smaller now
- the next follow-up should broaden the bounded integer-heavy corpus again and
  target the remaining generic leaders (`persistent_sorted_set_i32_small`,
  `word_count_small`, nearby call/container cases, and any residual
  `nbody_small` cost) instead of returning to unsigned-array storage or to a
  `sum_u32_small`-only path

## 2026-06-25 Canonical `StringBuilder` Fast Paths

The next bounded text pass stayed on that same generality rule and moved to
the new heaviest text cases instead of adding any builder-only opcode:

- pre-tranche bounded bytecode spot-check:
  - `string_builder_small`: `7.62s`
  - `automata_dfa_small`: `7.05s`
  - `string_contains_small`: `5.57s`
  - `word_count_small`: `3.50s`
- the recurring generic cost was canonical stdlib text construction through
  `StringBuilder.push_string`, `push_char`, `push_byte`, `push_bytes`, and
  `append_builder`, all of which still expanded into stdlib
  `validated_bytes(...)`, `char_to_utf8(...)`, and `Array.push_all(...)`
- the landed keep was a general runtime + VM cut:
  - runtime mono-`u8` append helpers for single bytes, byte slices, and host
    strings, with dynamic-to-mono promotion when the existing contents remain
    representable as `u8`
  - canonical bytecode fast paths for stdlib
    `StringBuilder.push_char`, `push_byte`, `push_bytes`, `push_string`,
    `append_builder`, and `finish`
  - `finish` returns the canonical stdlib `String` struct value, not a
    benchmark-only host string shortcut
- bounded post-tranche spot-check:
  - `string_builder_small`: `3.96s`
  - `automata_dfa_small`: `6.31s`
  - `string_contains_small`: `5.07s`
  - `word_count_small`: `3.03s`

Updated implication:

- the builder-side write cut is general enough to help adjacent text
  workloads, not only `string_builder_small`
- the next text follow-up should target `automata_dfa_small` and
  `string_contains_small`, or expand the corpus again, rather than revisiting
  `StringBuilder` write-side lowering

## 2026-06-25 Canonical `String` Fast Paths

The next bounded text pass stayed on the same generality rule and moved from
builder-side writes to canonical stdlib string dispatch:

- the residual hot shape was still generic stdlib work:
  - canonical stdlib `String.contains`, `replace`, `len_bytes`, and `bytes`
    only had VM fast paths for host `runtime.StringValue`, not for the
    canonical `able.text.string.String` struct values that the benchmarks
    actually produce
  - canonical `String.chars().iterator().next()` still fell through the
    generic iterator/member machinery because there was no char-iterator fast
    path
- the landed keep was a general VM change:
  - fast-path classification for text stdlib methods now keys off canonical
    struct definitions so `String`, `StringBuilder`, and raw string iterators
    are distinguished by type rather than by workload
  - canonical stdlib `String.len_bytes`, `contains`, `replace`, `bytes`, and
    `chars` now reuse the shared string fast-path machinery while preserving
    stdlib UTF-8 validation by falling back on invalid data
  - added canonical char-iterator `next` fast paths for
    `RawStringCharsIter` / `StringCharsIter`
  - fast-path-created iterator interface values now also expose the
    self-returning `iterator()` method so stdlib chains like
    `text.chars().iterator().next()` keep working without any special lowering
- focused guardrails:
  - source-level bytecode trace coverage for canonical `String.contains` /
    `replace` / `len_bytes` / `bytes` / `chars`
  - invalid-UTF-8 fallback coverage for canonical `String.chars`
- bounded post-tranche spot-check:
  - `string_builder_small`: `0.73s`
  - `automata_dfa_small`: `3.63s`
  - `string_contains_small`: `0.17s`
  - `word_count_small`: `3.14s`

Updated implication:

- `string_contains_small` is no longer a meaningful leader in the bounded text
  set
- `automata_dfa_small` remains relevant, but the front of the queue has moved
  enough that the next tranche should start with a broader bounded corpus
  refresh rather than assuming text-only work is still dominant
- continue rejecting any benchmark-specific opcode, lowering branch, or fusion
  that cannot be justified as a reusable stdlib/runtime dispatch improvement

## 2026-06-25 Generic Array Sync Instead Of Metadata-Triggered Full Rescans

The broader bounded corpus refresh did in fact expose a different generic wall,
but it was still shared runtime machinery rather than a benchmark-only shape:

- pre-keep bounded `fixture-full` leaders included:
  - `deque_i32_small`: `6.65s`
  - `queue_i32_small`: `5.95s`
  - `persistent_queue_i32_small`: `3.52s`
  - `int128_accumulate_small`: `11.37s`
  - `zigzag_char_small`: `10.71s`
  - `uint128_accumulate_small`: `9.70s`
  - `sum_u32_small`: `5.55s`
  - `random_lcg_i64_small`: `5.25s`
- an intermediate tracked-`Array i32` partial-cache branch was profiled and
  rejected because it made `deque_i32_small` / `queue_i32_small` worse
  (`7.30s` / `6.44s`) while `pprof` showed that the real cost was still
  `syncArrayValues(...)` rebuilding the full tracked `i32` cache from
  metadata-only kernel `Array.length`, `Array.capacity`, `Array.reserve`, and
  other handle-backed array sync sites
- the kept fix stayed on the generic array machinery instead:
  - added shared handle-based sync helpers for array writes, length changes,
    and metadata-only updates
  - direct array index/member writes now use incremental sync instead of the
    full `syncArrayValues(...)` refresh
  - native `__able_array_write`, `__able_array_set_len`,
    `__able_array_reserve`, and handle-backed bytecode write/push fallbacks
    now use those same helpers
  - no benchmark-specific opcode, fused expression pattern, or special
    lowering branch was added
- focused guardrails stayed green:
  - tracked-array cache/token tests, array slot/index write parity tests, and
    array swap cache tests under `env GOMEMLIMIT=1GiB GOGC=50 go test`
- bounded post-keep container check:
  - `deque_i32_small`: `1.99s`
  - `queue_i32_small`: `2.59s`
  - `persistent_queue_i32_small`: `3.57s`
  - `linked_list_enumerable_i32_small`: `2.02s`
  - `list_i32_small`: `1.21s`
  - `vector_i32_small`: `1.33s`
- broader bounded post-keep `fixture-full` confirmation:
  - `deque_i32_small`: `1.95s`
  - `queue_i32_small`: `2.62s`
  - `heap_i32_small`: still timeout
  - next generic leaders now cluster around integer/numeric/text workloads:
    - `int128_accumulate_small`: `11.30s`
    - `zigzag_char_small`: `10.36s`
    - `uint128_accumulate_small`: `10.22s`
    - `sum_u32_small`: `6.25s`
    - `nbody_small`: `6.07s`
    - `random_lcg_i64_small`: `5.49s`
    - `persistent_sorted_set_i32_small`: `4.12s`
    - `word_count_small`: `3.27s`

Updated implication:

- the right fix direction for the queue/deque wall was a reusable array sync
  correction, not a container-specific VM path
- continue rejecting any optimization justified only by one benchmark or one
  stdlib container name
- the next tranche should move to the remaining generic numeric / text leaders
  (`heap_i32_small`, `int128_accumulate_small`,
  `uint128_accumulate_small`, `zigzag_char_small`, `sum_u32_small`,
  `random_lcg_i64_small`, `nbody_small`, `persistent_sorted_set_i32_small`,
  `word_count_small`) rather than revisiting array metadata sync

## 2026-06-23 Fixture Corpus Expansion

The local fixture benchmark corpus now covers 76 workloads instead of 30. The
latest tranche added reduced external-style kernels (`base64`, `json`,
`monte_carlo_pi`, `nbody`, `pidigits`, `reverse_complement`,
`k_nucleotide`, `quicksort_file`, `sudoku_file`, `i_before_e`,
`binarytrees`, `mandelbrot`, `tapelang`), broader collection/concurrency
shapes (`array_*`, `list`, `lazy_seq_*`, `channel_*`, `mutex_counter`), and
numeric/text helpers (`random_lcg`, `md5_hex`, `ascii_lower`, `bigint`,
`bigint_ref_newton`, `biguint`, `rational`, `int128`, `uint128`).

The same update also retired stale non-runnable fixture directories
(`persistent_map_i32_small`, `tree_map_i32_small`, `tree_set_i32_small`) and
validated the formerly suspect new workloads one benchmark at a time under
bounded-memory settings so the corpus stays broad without reintroducing
benchmark-specific compiler or VM work.

A follow-on bounded validation pass also hardened the fixture validator itself:
`v12/bench_fixture_validate` now uses hard `timeout -k` build/run limits and
catalog-driven file arguments, which keeps broad corpus checks reliable under
memory pressure. The first post-hardening collections hotspot rerun closed one
real VM bug (`deque_i32_small` now matches in all three modes after the
bytecode `%` slot-const path learned to handle raw typed integer slots) and
left the remaining partials concentrated in broader compiled iterator/helper
and runtime interface-conformance gaps rather than in benchmark harness noise.

Implication:

- lowering audits no longer need to rely only on the external-style programs
  to get workload diversity
- `corpus-full` is the right default audit lane when deciding whether a new
  speculative opcode family is genuinely recurring

## 2026-06-25 Shared Raw-Value Boundary Follow-Up

The next broader-corpus follow-up after the raw small-integer VM lane stayed
inside the same generality rule and treated newly unblocked benchmarks as a
correctness audit before doing any more speed work.

What the broader slice exposed:

- `word_count_small` was not slow first; it was broken in bytecode:
  - the first failure came from canonical stdlib `String.split(...)` because
    shared typed-pattern matching and shared type coercion still assumed
    boxed primitive integers, while the VM now legitimately carries raw
    `u8`/`u32`/`u64` values in hot paths
  - after that fix, the same workload failed again because shared array
    construction still passed raw VM integer carriers directly into the
    runtime array store
- those are both shared interpreter/runtime boundary bugs, not workload-
  specific lowering opportunities

Kept fix direction:

- materialize raw VM values before shared pattern matching, type matching,
  type coercion, and casts
- materialize raw VM values before shared array-state creation hands elements
  to the runtime array store
- do not add any benchmark-specific lowering rule, opcode, or stdlib
  container fast path

Measured outcome:

- direct bytecode `word_count_small` now completes and matches treewalker
  output (`127152`) instead of failing inside stdlib `String.split(...)`
- bounded bytecode runtime spot-checks under `GOMEMLIMIT=1GiB GOGC=50`:
  - `word_count_small`: `3.23s`, `1.287GB/op`, `15.11M allocs/op`
  - `random_lcg_i64_small`: `2.39s`, `263MB/op`, `8.98M allocs/op`
  - `persistent_sorted_set_i32_small`: `3.98s`, `1.323GB/op`,
    `17.26M allocs/op`
  - `nbody_small`: `5.98s`, `996MB/op`, `27.19M allocs/op`

Bounded CPU-profile implication across `word_count_small`,
`persistent_sorted_set_i32_small`, and `nbody_small`:

- the next common wall is no longer integer boxing or unsigned array storage
- the recurring shared costs are now:
  - `expandTypeAliases(...)`
  - string-keyed runtime map lookup / hashing (`mapaccess2_faststr`,
    `aeshashbody`)
  - bytecode scope/cache bookkeeping such as `storeCachedScopeValue(...)`
  - GC scanning pressure caused by the same allocation-heavy shared paths

Updated implication:

- keep the next tranche centered on shared alias-expansion and
  string-keyed dispatch/cache lookup churn across call-heavy workloads
- do not answer these benchmarks with arithmetic-stencil opcodes or
  container-specific special cases

## 2026-06-25 Shared Alias / Interface Cache Follow-Up

The next broader-corpus follow-up stayed inside the same generality rule and
treated the pre-keep profiles as a shared runtime problem, not a benchmark
shaping opportunity.

What the prior profiles exposed:

- `word_count_small`, `persistent_sorted_set_i32_small`, and `nbody_small`
  all spent real time in the same shared call-heavy machinery:
  - repeated alias expansion during call-local type binding, method-set
    constraint enforcement, and impl matching
  - repeated interface method-dictionary construction for the same concrete
    receiver/interface pairs
  - the resulting string-keyed lookup churn and allocation pressure

Kept fix direction:

- move hot call-local binding, method-set constraint, and impl-resolution
  alias-expansion sites onto the existing shared `expandTypeAliasesCached(...)`
  path
- cache interface method dictionaries by `(concrete type, interface name,
  interface args)` behind the existing method-cache invalidation boundary
- clone cached method maps before attaching them to a fresh interface value so
  later per-value method memoization cannot alias shared cache state
- do not add any benchmark-specific opcode, lowering rule, or named-container
  special case

Added guardrails:

- `interpreter_method_resolution_cache_test.go` now pins interface method
  dictionary cache reuse, invalidation, and clone-on-handoff behavior

Measured outcome:

- bounded bytecode runtime spot-checks:
  - `word_count_small`: `2.35s`, `583MB/op`, `9.81M allocs/op`
    (from `3.23s`, `1.287GB/op`, `15.11M allocs/op`)
  - `nbody_small`: `5.72s`, `876MB/op`, `24.93M allocs/op`
    (from `5.98s`, `996MB/op`, `27.19M allocs/op`)
- bounded direct bytecode runs still complete correctly:
  - `word_count_small`: output `127152`, `6.21s`, peak RSS `604364 KB`
  - `nbody_small`: output pair `-0.16907516382852453` /
    `-0.16901644126443466`, `6.13s`, peak RSS `111892 KB`
  - `random_lcg_i64_small`: output `500123.29219735786`, `2.61s`, peak RSS
    `102816 KB`

Refreshed post-keep profile implication (`word_count_small`):

- interface method-dictionary rebuilds are no longer a top wall
- impl-candidate churn is now much smaller
- the remaining shared leaders are:
  - string-keyed map/hash work (`mapaccess2_faststr`, `aeshashbody`)
  - GC scan pressure from allocation-heavy shared call/member paths
  - residual shared call-local type binding / type-matching work

Updated implication:

- the next tranche should target shared call/member dispatch caches, scope
  lookup/update structures, and type-binding environment churn
- do not answer this new wall with interface-specific tweaks, container
  special cases, or benchmark-shaped bytecodes

## 2026-06-25 Call-Local Type Binding Cache Follow-Up

The next broader-corpus follow-up stayed inside the same generality rule and
treated the refreshed post-alias-cache profile as a shared generic-dispatch
setup problem.

What the prior profile exposed:

- `execCallMember` / `execCachedCallName` were still spending real time inside
  `callCallableValueWithInjectedReceiver(...)`
- the visible shared call-setup cost was repeated
  `bindCallLocalTypeBindings(...)` work:
  - rebuilding the same type-binding payloads for the same `(function,
    receiver type)` pairs
  - re-running merge-capable environment writes for synthetic `name` /
    `name_type` bindings that never need overload semantics

Kept fix direction:

- cache call-local type binding payloads per `(function value, receiver type)`
- apply cached runtime payloads with `Environment.DefineWithoutMerge(...)`
  because these synthetic bindings are never overload sets
- route explicit generic call type arguments through the same helper path
- clear the new cache on both method-cache invalidation and type-alias updates
- do not add any benchmark-specific opcode, lowering rule, or container
  special case

Added guardrails:

- `interpreter_method_resolution_cache_test.go` now pins call-local type
  binding cache reuse and invalidation on both `invalidateMethodCache()` and
  `RegisterTypeAlias(...)`

Measured outcome:

- the first one-shot benchmark sample was noisy, so the keep decision comes
  from repeated clean reruns plus the refreshed profile
- repeated bounded clean reruns now settle at:
  - `word_count_small`: `2.12-2.47s`, about `542MB/op`,
    about `8.52M allocs/op`
    (from the prior `2.35s`, `583MB/op`, `9.81M allocs/op`)
  - `nbody_small`: `5.45-5.68s`, about `794MB/op`,
    about `19.80M allocs/op`
    (from the prior `5.72s`, `876MB/op`, `24.93M allocs/op`)
- the refreshed `word_count_small` CPU profile now drops
  `bindCallLocalTypeBindings(...)` from about `240ms` cumulative to about
  `80ms`

Updated implication:

- the generic-binding setup wall moved as intended
- the next common wall is now the remaining string-keyed lookup/update
  machinery plus GC scan pressure around shared `call_name` / `call_member`
  dispatch and scope caches
- the next tranche should target those shared lookup/cache structures rather
  than more generic-binding work or any benchmark-shaped opcode change

## 2026-06-25 Scope Lookup Metadata / Hot-Name Follow-Up

The next bounded follow-up stayed on that same generic wall and targeted the
shared lexical lookup/cache path rather than adding any new opcode or lowering
shape.

What the prior profile/code path exposed:

- `lookupCachedIdentifierNameEntry(...)` was validating a scope-cache hit and
  then re-reading the same `scopeLookupCache[key]` entry to recover metadata
- scope-cache hits were not reseeding `nameLookupHot`, unlike global-cache
  hits, so same-site lexical lookups could keep paying the string-keyed scope
  map path after the inline hot entry had been displaced
- shared lookup/store helpers were recomputing the same env/owner revision
  payload immediately after cache validation or store

Kept fix direction:

- make validated scope-cache hits return the resolved lookup metadata directly
- reseed `nameLookupHot` on scope-cache hits just like global-cache hits
- have the shared scope/global helper boundaries return the resolved lookup
  payload so callers do not immediately rebuild it
- do not add any benchmark-specific opcode, lowering rule, or named-stdlib
  special case

Added guardrails:

- `bytecode_vm_lookup_cache_test.go` now pins the scope-cache entry path and
  verifies that a scope-cache hit refreshes the hot inline name cache

Measured outcome:

- focused tests:
  - `go test ./pkg/interpreter -run 'TestBytecodeVM_(LookupCachedIdentifierNameUsesHotValueCache|ResolveCachedIdentifierNameUsesScopeCache|LookupCachedIdentifierNameEntryUsesScopeCacheAndSeedsHotCache)$' -count=1`
  - `go test ./pkg/interpreter -run '^$' -count=1`
- bounded spot-checks:
  - repeated `word_count_small`: `2.16-2.20s`,
    about `542.6-542.8MB/op`, about `8.53M allocs/op`
  - `nbody_small` control: `5.25s`, `794MB/op`, `19.80M allocs/op`

Updated implication:

- the lookup-entry duplication is closed without changing semantics or adding a
  workload-shaped bytecode
- the next shared wall is still the remaining string-keyed dispatch/cache
  update machinery around `call_name`, `call_member`,
  `lookupCachedScopeValue`, and `storeCachedScopeValue`
- the next tranche should stay there rather than revisiting this now-closed
  lookup-entry path

## 2026-06-25 Scope Cache Entry Reuse / Shared Hot Metadata Follow-Up

The next bounded follow-up stayed on that same shared cache wall and narrowed
in on the remaining cost inside `storeCachedScopeValue(...)`.

What the prior profile exposed:

- after the lookup-entry cleanup, `storeCachedScopeValue(...)` was still a
  meaningful flat cost center
- the cache still stored full scope-entry structs directly in the map, so
  repeated refreshes at the same bytecode site paid another map value
  write/copy
- `nameLookupHot` then copied the same metadata again after the store

Kept fix direction:

- make `scopeLookupCache` store stable entry objects and update them in place
- have `nameLookupHot` point at shared lookup metadata instead of copying the
  full entry payload on every reseed
- reuse a dedicated VM-owned entry object for global hot-name stores so the hot
  path uses the same shape for both lexical and global hits
- do not add any benchmark-specific opcode, lowering rule, or named-stdlib
  special case

Added guardrails:

- `bytecode_vm_lookup_cache_test.go` now pins same-site scope-cache entry reuse
  and verifies the hot-name cache points at the reused entry object

Measured outcome:

- focused tests:
  - `go test ./pkg/interpreter -run 'TestBytecodeVM_(LookupCachedIdentifierNameUsesHotValueCache|ResolveCachedIdentifierNameUsesScopeCache|LookupCachedIdentifierNameEntryUsesScopeCacheAndSeedsHotCache|StoreCachedScopeValueReusesEntryObject|ResetForRunPreservesLookupCaches)$' -count=1`
  - `go test ./pkg/interpreter -run '^$' -count=1`
- bounded spot-checks:
  - repeated `word_count_small`: `2.10-2.20s`,
    about `542.1-543.3MB/op`, about `8.53M allocs/op`
  - `nbody_small` control: `5.15s`, `794MB/op`, `19.80M allocs/op`
- refreshed profile deltas on `word_count_small`:
  - `storeCachedScopeValue(...)`: about `100ms flat / 180ms cum` down to about
    `10ms flat / 40ms cum`
  - `lookupCachedIdentifierNameEntry(...)`: about `320ms cum` down to about
    `140ms cum`

Updated implication:

- the targeted scope-cache store/update wall is materially smaller now without
  changing semantics or adding a workload-shaped bytecode
- the next shared wall is the remaining string-keyed dispatch/cache
  bookkeeping inside `call_name` / `call_member` and nearby interpreter
  call/coercion paths
- the next tranche should move there rather than revisiting scope-cache
  storage

## 2026-06-25 Cached Member-Method Template Dispatch Follow-Up

The next follow-up stayed on that same generality rule and kept the focus on
shared bytecode `call_member` dispatch rather than on any benchmark-shaped
opcode or lowering fusion.

What the prior profile and escape analysis exposed:

- after the scope-cache keep, `word_count_small` still spent meaningful time in
  generic `call_member` setup and injected-receiver call plumbing
- member-method cache hits were still validating a cached template and then
  rebuilding a bound wrapper callable before re-entering the generic member
  call shell
- a first attempt to inline the injected-receiver arg merge inside
  `callCallableValueWithInjectedReceiver(...)` was not trustworthy:
  - focused semantics stayed green
  - but bounded spot-checks regressed badly (`word_count_small` moved up into
    the `2.72-2.87s` range and `nbody_small` to about `6.13s`)
  - that branch was backed out and should stay rejected

Kept fix direction:

- keep only the generic cache-hit wrapper-elision half of the work
- make cached member-method hits carry the unbound method template plus the
  existing fast-path metadata instead of a pre-bound callable wrapper
- have cached `execCallMember(...)` hits feed that template plus the stack
  receiver directly into the shared exact-native / injected-receiver dispatch
  path
- do not add any benchmark-specific opcode, lowering rule, or named-stdlib
  special case

Added guardrails:

- `call_callable_native_bound_partial_test.go` now also pins direct
  `runtime.BoundMethodValue` receiver injection into interpreted function calls
- `bytecode_vm_member_cache_test.go` now pins that cached member-method entries
  keep an unbound native template that remains exact-native dispatchable

Measured outcome:

- focused tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters)|BytecodeVM_CallNameDotFallbackUsesMemberMethodCache|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
- uncapped spot-checks on the current tree:
  - `word_count_small`: about `2.18s`, `533MB/op`, `8.50M allocs/op`
  - `nbody_small`: about `5.41-5.44s`, `767MB/op`, `19.21M allocs/op`
- bounded anti-OOM controls for the same binary:
  - `word_count_small`: about `2.33s`, `533.6-533.8MB/op`,
    `8.50M allocs/op`
  - `nbody_small`: about `5.23s`, `767.6MB/op`, `19.21M allocs/op`
- refreshed `word_count_small` CPU profile after the keep:
  - `execCallMember(...)`: about `2.34s` cumulative
  - `callResolvedCallableWithInjectedReceiver(...)`: about `1.94s`
    cumulative
  - `lookupCachedMemberMethodEntry(...)`: about `10ms` cumulative
- filtered escape analysis still reports bound-wrapper escapes in the generic
  `lookupCachedMemberMethod(...)` / `bindMemberMethodTemplate(...)` helpers and
  in interpreter-side method resolution, but it no longer reports the old
  cached-hit wrapper-construction helper that used to sit directly on the
  member-method cache entry path

Updated implication:

- this keep is still generally-applicable dispatch/cache work: cached
  member-method hits no longer need their own bound wrapper object before
  re-entering the shared member-call path
- the main remaining cost is still the broader generic member-call shell, not
  member-method cache lookup itself
- the next tranche should therefore stay on shared `call_member` /
  `call_name` / injected-receiver call-shell overhead rather than revisiting
  cache-hit wrapper construction or reviving the rejected inline receiver-arg
  experiment

## 2026-06-25 Member-Call Arg Preparation Follow-Up

The next follow-up stayed on that same generality rule and narrowed in on the
remaining allocation pressure inside shared bytecode `call_member` dispatch.

What the prior profile exposed:

- cached member-method template dispatch removed cache-hit wrapper rebuilds, but
  the remaining generic member-call shell was still paying for extra arg-slice
  cloning on the injected-receiver path
- the same pattern also showed up in shared overload selection for member calls:
  member-call overload resolution first cloned/materialized the explicit args
  and then built another eval-arg slice for the receiver-prefixed signature
- this is reusable call-shell work, not benchmark shape work

Kept fix direction:

- add a shared `bytecodePrepareCallArgs(...)` helper so member-call paths now
  copy only when the target actually needs stable borrowed args
- otherwise materialize raw VM values in place on the existing arg slice
- make `callResolvedCallableWithInjectedReceiver(...)` use that in-place
  prepared slice instead of pre-cloning all explicit args
- make member-call overload selection materialize directly into its inline or
  allocated eval-arg buffer instead of cloning the explicit arg slice first
- materialize exact-native member-call args in place right before the native
  handoff so computed raw integers/floats reach native code as ordinary runtime
  values without introducing another general copy

Added guardrails:

- `bytecode_vm_call_native_fastpath_test.go` now pins that an exact-native
  bound-member call receives a materialized computed integer argument, not a
  raw VM carrier

Measured outcome:

- focused tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallNameDotFallbackUsesMemberMethodCache|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
- uncapped spot-checks on the current tree:
  - `word_count_small`: about `2.16-2.33s`, `526.7-528.3MB/op`,
    about `8.32M allocs/op`
  - `nbody_small`: about `4.91-5.00s`, `701.3-701.4MB/op`,
    about `17.40M allocs/op`
- bounded anti-OOM controls:
  - `word_count_small`: repeated reruns settled around `2.21-2.25s`,
    `527.5-527.7MB/op`, about `8.32M allocs/op` after one noisy `2.55s`
    first sample
  - `nbody_small`: about `4.91s`, `701.8MB/op`, `17.40M allocs/op`
- refreshed `word_count_small` CPU profile after the keep:
  - `execCallMember(...)`: about `2.22s` cumulative
    (from about `2.34s`)
  - `callResolvedCallableWithInjectedReceiver(...)`: about `1.85s`
    cumulative (from about `1.94s`)
  - `execCachedCallName(...)`: still about `0.83s` cumulative, with about
    `0.50s` of that still flowing into the generic cached `call_name`
    execution path

Updated implication:

- this keep is still generally-applicable call-shell work: member-call arg
  slices are no longer cloned eagerly before overload selection or generic
  injected-receiver dispatch
- the large runtime/allocation win on `nbody_small` and the smaller but still
  useful allocation win on `word_count_small` both point at the same reusable
  cause: less shared call-arg churn, not a benchmark-only fusion
- the next tranche should therefore stay on the remaining generic cached
  `call_name` path plus the direct/member call shells
  (`execCachedCallName(...)`, `callCallableValueMutable(...)`,
  `execCallMember(...)`, `callResolvedCallableWithInjectedReceiver(...)`)
  rather than revisiting arg-slice cloning or introducing benchmark-specific
  bytecode

## 2026-06-25 Cached `call_name` / Direct-Call Arg Preparation Follow-Up

The next follow-up stayed on the same shared call-shell wall and removed the
remaining eager raw-value materialization from the non-member call paths.

What the prior profile exposed:

- cached `call_name` still materialized every stack arg before dispatch even
  when the call ended up on an inline path that could consume raw VM integer
  carriers directly
- uncached `call_name`, direct `call`, and `call_self` still had the same old
  pattern: materialize first, then decide whether the path was exact-native,
  inline, or generic
- this is still reusable VM dispatch overhead, not workload shaping

Kept fix direction:

- make cached `call_name` mirror the kept member-call behavior:
  - inline paths keep the existing raw stack args
  - exact-native paths materialize immediately before the native handoff
  - generic paths use `bytecodePrepareCallArgs(...)` so they copy only when
    the callee actually needs stable borrowed args
- apply that same dispatch-order cleanup to the shared direct `execCall(...)`,
  `execCallSelf(...)`, and uncached `execCallName(...)` paths instead of
  leaving eager materialization at the top of each opcode helper

Added guardrail:

- `bytecode_vm_call_native_fastpath_test.go` now also pins that an exact-native
  `call_name` site receives a materialized computed integer argument, not a raw
  VM carrier

Measured outcome:

- focused tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
- direct bounded runtime spot-checks on the current tree:
  - `word_count_small`: about `2.19-2.23s`, `527.6-527.8MB/op`,
    about `8.328M allocs/op`
  - `nbody_small`: about `4.96-5.12s`, `702.0MB/op`,
    about `17.40M allocs/op`
- refreshed sequential `word_count_small` CPU profile after the keep:
  - `execCachedCallName(...)`: about `0.82s` cumulative
    (from about `0.83s`)
  - the generic cached `call_name` execution path inside that drops to about
    `0.47s` cumulative (from about `0.50s`)
  - `callResolvedCallableWithInjectedReceiver(...)` still dominates at about
    `1.88s` cumulative, which is now larger than the remaining cached
    `call_name` wall by a wide margin

Updated implication:

- this keep is still generally-applicable dispatch cleanup: raw VM args now
  stay raw until the direct/cached call path knows whether it needs inline,
  exact-native, or generic handling
- `word_count_small` holds in the improved band with a slightly smaller cached
  `call_name` profile footprint, while `nbody_small` stays effectively flat in
  a noisy band; that is enough to keep the cleanup
- the next tranche should move one layer deeper into the shared interpreter
  call helper (`callCallableValueWithOptionalInjectedReceiver(...)` and the
  overload/native/direct-function setup it drives) rather than spending more
  time on outer VM arg materialization sites

## 2026-06-25 Rejected Follow-Up: Direct Bound-Function Injected-Receiver Helper

The next experiment stayed generic as well, but it did not earn its
complexity. It tried to bypass the receiver-prefixed arg slice for the
direct-function subset of injected-receiver calls inside
`callCallableValueWithOptionalInjectedReceiver(...)`.

What the prior profile suggested:

- after the kept `call_name` cleanup, the remaining wall was no longer outer
  VM arg materialization
- `callCallableValueWithOptionalInjectedReceiver(...)` and the direct
  `invokeFunction(...)` path it feeds were now the main shared call-shell cost
- the direct bound-function path looked like a plausible place to remove one
  more generic arg-slice allocation and slot-seeding step

Attempted direction:

- add a helper that keeps the receiver separate from explicit args for the
  direct-function injected-receiver path
- let that helper:
  - perform arity checks and method-set constraint checks
  - coerce receiver/explicit args without first building a receiver-prefixed
    slice
  - seed bytecode slot frames directly from receiver plus explicit args
- keep the old generic path as a fallback for thunks, generic inference, and
  any other case that still needed the shared path

Guardrail learned during implementation:

- the first version exposed a real correctness hazard: eager fallback
  preparation mutated caller arg backing storage when a bound direct-function
  call needed coercion
- that was pinned by the focused regression test used during the experiment,
  but the helper itself was still reverted because the runtime signal was not
  good enough

Verification on the candidate branch:

- focused tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction|DoesNotMutateBoundDirectFunctionArgsOnCoercion)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.046s`
- bounded runtime spot-checks on the candidate branch:
  - `word_count_small`: `2206199013-2327122313 ns/op`,
    `525152520-534403224 B/op`, `8327614-8328598 allocs/op`
  - `nbody_small`: `5323378624-5337067454 ns/op`,
    `701926768-702203192 B/op`, `17403775-17404459 allocs/op`
- candidate-branch profile implication:
  - the helper was definitely active
  - but the remaining wall still sat inside the shared direct-call path, and
    the broader runtime signal was mixed at best on `word_count_small` and
    clearly worse on `nbody_small`

Decision:

- reject and revert the direct bound-function injected-receiver helper
- reason:
  - it stayed generic and benchmark-agnostic, which was the right constraint
  - but it did not show a reusable corpus-level win
  - it added call-path complexity while moving `nbody_small` into a slower
    band, so keeping it would have lowered the bar on evidence

Post-revert current-tree controls:

- focused tests after the revert:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.046s`
- low-memory reruns on the reverted current tree:
  - `word_count_small`: `2001442608-2170823642 ns/op`,
    `428169736-428271256 B/op`, `6964964-6965268 allocs/op`
  - `nbody_small`: `4327730610-4389841497 ns/op`,
    `345602688-345646448 B/op`, `12272192-12272311 allocs/op`
- refreshed current-tree `word_count_small` CPU profile after the revert:
  - `callCallableValueWithOptionalInjectedReceiver(...)`: about `2.01s`
    cumulative
  - `invokeFunction(...)`: about `2.05s` cumulative
  - `execCachedCallName(...)`: now only about `0.76s` cumulative

Updated implication:

- the next productive tranche should stay generic and move deeper into the
  shared direct-call machinery itself, especially `invokeFunction(...)`
  parameter binding, coercion, local-env setup, and slot-frame seeding
- do not revive the rejected helper and do not introduce benchmark-shaped
  fusions or opcodes to compensate for it

## 2026-06-25 Kept Follow-Up: Shared Simple-Type Coercion Fast Path

The next follow-up stayed inside the shared call/coercion machinery instead of
returning to dispatch-only work. The current profile said the direct-call wall
still ran through `invokeFunction(...)`, and within that path the reusable
coercion helpers were costing real time on every call.

What the post-revert profile exposed:

- `coerceReturnValue(...)` still carried about `250ms` cumulative on
  `word_count_small`
- `coerceValueToType(...)` still carried about `250ms` cumulative there too
- the old simple-type coercion path still re-entered the generic matcher /
  canonicalizer even when the declared type was already a non-aliased simple
  type and the value only needed an ordinary primitive coercion

Kept fix direction:

- add a shared `tryFastSimpleTypeCoercion(...)` helper that:
  - reuses the existing exact named-struct fast path
  - reuses the existing inline primitive coercion helpers for integer/floating
    simple types
  - short-circuits exact simple `String` / `bool` / `char` / `Error` cases
- make `coerceValueToType(...)` use that helper before falling back to the
  general coercion path
- make `coerceReturnValue(...)` try the same helper before paying the full
  alias-canonicalization and generic matcher path when the declared return type
  is already a non-aliased simple type

Added guardrails:

- `interpreter_type_coercion_fast_test.go` now also pins:
  - repeated `coerceValueToType(i32 -> f64)` on a small integer stays at or
    below the current fast-path allocation floor
  - repeated `coerceReturnValue(i32 -> f64)` on a small integer stays at or
    below that same floor

Measured outcome:

- focused tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(Interpreter(CoerceValueToTypeWouldBeNoOp|CastValueToCanonicalSimpleTypeFast|CoerceValueToTypeUnsignedIntegerWideningUsesValueRange|CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath|CoerceValueToTypeExactNamedStructReturnsSameValue|CoerceReturnValueExactNamedStructReturnsSameValue|CoerceValueToTypeExactNamedStructUnwrapsErrorPayload)|Inline(CoerceValueBySimpleTypeUnsignedIntegerWideningUsesValueRange|CoercionUnnecessaryWithInterpreterExactNamedStruct|CoercionUnnecessaryWithInterpreterDoesNotTreatErrorAsExactNamedStruct)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.043s`
- broader call/bytecode regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|Interpreter(CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath|CoerceValueToTypeExactNamedStructReturnsSameValue|CoerceReturnValueExactNamedStructReturnsSameValue)|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.046s`
- compile-only smoke:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.051s [no tests to run]`
- low-memory runtime spot-checks on the kept branch:
  - `word_count_small`: `1971397072-2015052143 ns/op`,
    `423842624-425997280 B/op`, `6924810-6925229 allocs/op`
  - `nbody_small`: `4240994370-4335043229 ns/op`,
    `345651088-345765608 B/op`, `12272315-12272589 allocs/op`
- profiled `word_count_small` rerun on the kept branch:
  - `1998048493 ns/op`, `426012200 B/op`, `6925206 allocs/op`

Refreshed profile implication:

- `coerceReturnValue(...)` still exists as a shared wall, but the expensive
  `matchesType(...)` portion inside it drops from about `190ms` to about
  `110ms` cumulative on the sampled rerun
- `coerceValueToType(...)` now enters the simple-type fast path directly on the
  sampled run, but the remaining major coercion cost has shifted toward generic
  interface coercion and repeated alias checks rather than small integer/float
  widening itself
- the overall runtime signal is enough to keep the change: `word_count_small`
  moves from the prior `2.00-2.17s`, `~428.2MB/op`, `~6.97M allocs/op` band
  down to about `1.97-2.02s`, `423.8-426.0MB/op`, `~6.925M allocs/op`, and
  `nbody_small` also edges down into about `4.24-4.34s`

Updated implication:

- keep this coercion cleanup
- the next productive tranche should stay on the remaining shared coercion
  wall, especially generic interface coercion inside `coerceValueToType(...)`
  plus repeated alias/canonicalization checks in `coerceReturnValue(...)`
  rather than returning to benchmark-shaped dispatch experiments

## 2026-06-25 Kept Follow-Up: Narrow Alias-Reference Cache For Shared Coercion Checks

The next follow-up stayed on the remaining coercion wall, but it narrowed the
scope of the cache after an initial broader version added too much churn. The
kept version caches alias-reference checks only where the same declared type
expressions are reused heavily in the shared call/coercion path, rather than
inside the generic alias-expansion helper itself.

What the prior kept profile exposed:

- after the simple-type coercion fast path, `typeExpressionReferencesAlias(...)`
  still cost about `180ms` cumulative on the profiled `word_count_small` rerun
- that alias-reference walk was now larger than the sampled interface-method
  dictionary construction cost on the same run
- the remaining coercion wall also pointed more clearly at
  `coerceToInterfaceValue(...)` / `lookupImplEntry(...)`, so any alias-cache
  change had to stay scoped and not add more dynamic churn

Kept fix direction:

- add a cached `typeExpressionReferencesAliasCached(...)` helper keyed by the
  existing interpreter/type-expression identity, invalidated alongside the
  existing alias-expansion cache on `RegisterTypeAlias(...)`
- use that cached predicate only at the shared call/coercion entry points where
  the same declared type expressions repeat:
  - `coerceReturnValue(...)`
  - `castValueToType(...)`
- do not use that cached predicate inside `expandTypeAliasesCached(...)`
  itself; the broader version regressed `word_count_small`, so the kept version
  leaves the generic expansion helper on the plain recursive predicate

Added guardrail:

- `interpreter_type_info_cache_test.go` now pins that the alias-reference cache
  stores results and clears on alias registration before recalculating against
  the updated alias set

Measured outcome:

- focused alias/coercion tests:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(ExpandTypeAliasesCachedReusesAndInvalidates|TypeExpressionReferencesAliasCachedReusesAndInvalidates|Interpreter(CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.043s`
- broader call/bytecode regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|Interpreter(CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath|CoerceValueToTypeExactNamedStructReturnsSameValue|CoerceReturnValueExactNamedStructReturnsSameValue)|ExpandTypeAliasesCachedReusesAndInvalidates|TypeExpressionReferencesAliasCachedReusesAndInvalidates|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.047s`
- compile-only smoke:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.039s [no tests to run]`
- low-memory runtime spot-checks on the kept branch:
  - `word_count_small`: `1979596260-2068814618 ns/op`,
    `423934480-425971728 B/op`, `6924801-6925189 allocs/op`
    - two of the three clean reruns held near `1.98-2.01s`; one slower sample
      landed at `2.07s`
  - `nbody_small`: `4290396228-4327382083 ns/op`,
    `345563032-345619272 B/op`, `12272092-12272230 allocs/op`
- profiled `word_count_small` rerun on the kept branch:
  - `2014554726 ns/op`, `426038968 B/op`, `6925318 allocs/op`

Refreshed profile implication:

- `typeExpressionReferencesAlias(...)` drops from about `180ms` to about
  `150ms` cumulative on the sampled rerun
- `coerceReturnValue(...)` stays in about the same general band as the prior
  keep, while the remaining dominant shared coercion cost now shows more
  clearly inside:
  - `coerceToInterfaceValue(...)`: about `220ms` cumulative
  - `lookupImplEntry(...)` inside that path: about `150ms` cumulative
- `buildInterfaceMethodDictionary(...)` is now a smaller sampled cost than the
  remaining interface lookup path

Updated implication:

- keep the narrowed alias-reference cache
- the runtime band is effectively flat-to-slightly-better with a real profile
  reduction on the alias-reference wall and no broad regression signal
- the next productive tranche should stay on generic interface coercion,
  especially `lookupImplEntry(...)`, `collectImplCandidates(...)`, and related
  impl/interface lookup setup, rather than spending more time on alias scans or
  dispatch-only experiments

## 2026-06-25 Kept Follow-Up: Shared Selected-Impl Cache For Interface Coercion

The next follow-up stayed on the same generic coercion wall and removed
duplicate impl-candidate scans from the shared interface satisfaction path.

What the prior kept profile exposed:

- `coerceToInterfaceValue(...)` still cost about `220ms` cumulative on the
  sampled `word_count_small` rerun
- inside that path, `lookupImplEntry(...)` and `collectImplCandidates(...)`
  still cost about `150ms` cumulative
- the work was structurally duplicated: `typeImplementsInterface(...)` first
  resolved the impl to answer yes/no, then `coerceToInterfaceValue(...)` asked
  `lookupImplEntry(...)` for the same `(type, interface, interface args)` tuple
  again

Kept fix direction:

- add a shared selected-impl cache keyed by the same
  `(type, interface, interface args)` tuple already used by the boolean
  interface-impl cache
- populate that cache from `lookupImplEntry(...)` and let
  `typeImplementsInterface(...)` prime it on the first miss
- clear the new cache on the same `invalidateMethodCache()` boundary as the
  existing method/interface caches
- keep the change generic: no benchmark-shaped opcode, lowering rule, or
  workload-specific dispatch path

Added guardrail:

- `interpreter_method_resolution_cache_test.go` now pins that
  `typeImplementsInterface(...)` primes the selected-impl lookup cache and that
  the selected cache clears together with the existing method caches

Measured outcome:

- focused cache regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(TypeImplementsInterface_(CachesAndInvalidatesWithMethodCache|PrimesSelectedImplLookupCache)|CoerceToInterfaceValue_MethodDictionaryCacheInvalidatesAndClones)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.052s`
- broader call/bytecode regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|Interpreter(CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath|CoerceValueToTypeExactNamedStructReturnsSameValue|CoerceReturnValueExactNamedStructReturnsSameValue)|ExpandTypeAliasesCachedReusesAndInvalidates|TypeExpressionReferencesAliasCachedReusesAndInvalidates|TypeImplementsInterface_(CachesAndInvalidatesWithMethodCache|PrimesSelectedImplLookupCache)|CoerceToInterfaceValue_MethodDictionaryCacheInvalidatesAndClones|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.057s`
- compile-only smoke:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.049s [no tests to run]`
- low-memory runtime spot-checks on the kept branch:
  - `word_count_small`: `1911151681-1931360505 ns/op`,
    `339217440-339930856 B/op`, `6644413-6644494 allocs/op`
  - `nbody_small`: `4642000705 ns/op`, `345748992 B/op`, `12272552 allocs/op`
- profiled `word_count_small` rerun on the kept branch:
  - `1911151681 ns/op`, `339930856 B/op`, `6644494 allocs/op`

Refreshed profile implication:

- `lookupImplEntry(...)` / `collectImplCandidates(...)` drop out of the leading
  sampled `word_count_small` hotspots after the keep
- `coerceToInterfaceValue(...)` falls to about `60ms` cumulative
- `typeImplementsInterface(...)` falls to about `30ms` cumulative
- the new generic leaders are now:
  - `matchesType(...)`, especially repeated value-to-type-expression inspection
    and alias expansion in shared type matching
  - bytecode scope/cache update work around `execCallMember(...)`,
    `storeCachedScopeValue(...)`, and raw-slot materialization helpers

Updated implication:

- keep the selected-impl cache
- this is a broad shared win rather than a benchmark-shaped one: it removes
  duplicate impl-resolution work for any repeated interface coercion or
  interface match on the same concrete type/interface pair
- the next productive tranche should target shared type-matching churn
  (`matchesType(...)`, `typeExpressionFromValue(...)`, related alias/type-info
  paths) and the remaining bytecode call/cache bookkeeping rather than
  returning to impl-candidate lookup

## 2026-06-25 Kept Follow-Up: Remove Eager Value-Type Reconstruction From Shared Type Matching

The next follow-up stayed on the new generic type-matching wall and removed a
shared pre-pass that was rebuilding full value type expressions for every
`matchesType(...)` call, even when the target shape already had cheaper
specialized logic.

What the prior kept profile exposed:

- `matchesType(...)` had become the new leading shared interpreter cost at
  about `450ms` cumulative on the sampled `word_count_small` rerun
- inside that path, the eager exact-match pre-pass was still paying for:
  - `typeExpressionFromValueWithSeen(...)`: about `100ms` cumulative
  - alias expansion on the reconstructed value type: about `80ms` cumulative
- those costs were now larger than the remaining exact-match benefit because
  the branch-specific `matchesType(...)` logic already knows how to handle the
  relevant primitive, array, interface, result/option, union, and nominal
  generic cases

Kept fix direction:

- remove the eager `typeExpressionFromValue(...)` equality pre-pass at the top
  of `matchesType(...)`
- keep the existing branch logic and existing
  `fastExactNamedStructTypeMatch(...)` intact so the exact-match behavior still
  comes from the type-shape-specific paths rather than from a generic
  reconstruct-and-compare pre-pass
- keep the change generic: no benchmark-shaped opcode, lowering rule, or
  workload-specific dispatch path

Added guardrail:

- `interpreter_type_info_cache_test.go` now pins that exact generic nominal
  matches still work both directly and through alias expansion:
  - `Box<i32>` matches an exact `Box<i32>` value
  - `Box<String>` still does not match that same value
  - an alias expanding to `Box<i32>` still matches correctly

Measured outcome:

- focused type/alias regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(MatchesType_GenericStructAndAliasRemainExact|ExpandTypeAliasesCachedReusesAndInvalidates|TypeExpressionReferencesAliasCachedReusesAndInvalidates)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.061s`
- broader call/bytecode regression set:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run 'Test(CallCallableValue_(NativeBoundMethodPartialDoesNotDoubleInjectReceiver|NativeSkipContextPassesNilContext|BoundMethodValueInjectsReceiverIntoDirectFunction)|CallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion|CallFunction_DoesNotMutateCallerArgsOnCoercion|Interpreter(CoerceValueToTypeSmallIntegerToFloatAvoidsBigIntPath|CoerceReturnValueSmallIntegerToFloatAvoidsBigIntPath|CoerceValueToTypeExactNamedStructReturnsSameValue|CoerceReturnValueExactNamedStructReturnsSameValue)|ExpandTypeAliasesCachedReusesAndInvalidates|TypeExpressionReferencesAliasCachedReusesAndInvalidates|MatchesType_GenericStructAndAliasRemainExact|TypeImplementsInterface_(CachesAndInvalidatesWithMethodCache|PrimesSelectedImplLookupCache)|CoerceToInterfaceValue_MethodDictionaryCacheInvalidatesAndClones|BytecodeVM_(CallMemberOpcodeExecutesMethodCall|CallMemberHandlesOptionalMethodArity|CallMemberHandlesOverloadedMethods|MemberMethodCacheTracksStructDefinition|MemberMethodCacheWorksInPackageEnv|LookupCachedMemberMethodEntryKeepsTemplateForExactNativeDispatch|NonMethodMemberAccessSkipsMemberMethodCacheCounters|NativeCallNameExactCallMaterializesComputedIntegerArg|NativeBoundMethodExactCallInjectsReceiverOnce|NativeBoundMethodExactCallMaterializesComputedIntegerArg|NativeBoundMethodArgsStayStableWhenBorrowDisabled|NativeExactCallsSkipInlineProbeStats|NativeExactCallSkipContextPassesNilContext)|BytecodeVM_CallName(CacheRecordsDirectInlineShape|CacheSkipsDirectInlineForTypeArguments|DotFallbackUsesMemberMethodCache)|BytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod)$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.047s`
- compile-only smoke:
  - `env GOMEMLIMIT=1GiB GOGC=50 go test ./pkg/interpreter -run '^$' -count=1`
  - result: `ok able/interpreter-go/pkg/interpreter 0.047s [no tests to run]`
- low-memory runtime spot-checks on the kept branch:
  - `word_count_small`: `1664816273-1683037845 ns/op`,
    `315857360-315948984 B/op`, `6363910-6364032 allocs/op`
  - `nbody_small`: `4143913318 ns/op`,
    `345614952 B/op`, `12272231 allocs/op`
- profiled `word_count_small` rerun on the kept branch:
  - `1683037845 ns/op`, `315948984 B/op`, `6364032 allocs/op`

Refreshed profile implication:

- `matchesType(...)` drops from about `450ms` cumulative to about `180ms`
- `typeExpressionFromValueWithSeen(...)` drops from about `100ms` cumulative to
  about `20ms`
- the new generic leaders are now more clearly in shared bytecode
  call/cache/scope bookkeeping and surrounding map/GC pressure:
  - `execCallMember(...)`
  - `storeCachedScopeValue(...)`
  - `lookupCachedIdentifierNameEntry(...)`
  - `lookupCachedScopeValue(...)`

Updated implication:

- keep this `matchesType(...)` change
- this is a broad shared win rather than a benchmark-shaped one: it removes an
  eager reconstruct-and-compare pre-pass from every runtime type match and lets
  the existing type-shape-specific logic decide only when needed
- the next productive tranche should target shared bytecode scope/cache update
  work and identifier/member dispatch bookkeeping rather than returning to
  type-expression reconstruction

The next follow-up is now landed too: top-level slot-backed bytecode returns
now go through the same frame-layout return checks that inline returns already
used, via a shared `coerceProgramReturnValue(...)` helper in the VM. That keeps
`returnSimpleCheck`, `returnSimpleType`, `returnExactStructDef`, and cached
generic-name sets on the fast path for both inline and top-level successful
returns, and lets `invokeFunction(...)` / bytecode lambda dispatch trust the
VM-returned value instead of re-running `coerceReturnValue(...)` outside the
VM. The keep stayed generic and benchmark-agnostic: it closes a shared
VM/interpreter asymmetry in return handling instead of adding any
benchmark-shaped opcode or lowering rule. New guardrails in
`bytecode_vm_top_level_return_coercion_test.go` pin typed top-level `return`,
top-level `returnSignal`, generic return-name passthrough, and top-level
`ReturnBinaryIntAddI32` coercion. Under the bounded `GOMEMLIMIT=1GiB GOGC=50`
spot-checks, `word_count_small` stays in the current band at
`1679344360 ns/op`, `315846104 B/op`, `6363847 allocs/op`, while a single
bounded `nbody_small` control run lands at
`4377922551 ns/op`, `345651296 B/op`, `12272314 allocs/op`. More importantly,
the refreshed bounded `word_count_small` CPU profile drops
`coerceReturnValue(...)` from about `290ms` cumulative to about `190ms`, with
`canonicalizeTypeExpression(...)` inside that path down from about `140ms` to
about `60ms`. The next follow-up should therefore target the remaining generic
return-type canonical alias/match work inside `coerceReturnValue(...)` and the
nearby shared `call_member` / scope-cache bookkeeping rather than returning to
top-level return handling or inventing benchmark-shaped bytecodes.

That next follow-up is now landed too: alias-free and alias-expanded return
canonicalization now stay on the shared cached type-expression path instead of
falling back to the raw alias-expansion walk. Specifically,
`expandTypeAliasesCached(...)` now uses the cached alias-reference predicate for
the negative path, and `coerceReturnValue(...)` now canonicalizes through a
shared `canonicalizeTypeExpressionCached(...)` helper that reuses cached alias
expansion and skips the old raw path entirely when the return type is already
alias-free. The keep stayed generic and benchmark-agnostic: it reduces repeated
type-expression work that applies across return coercion, matching, and any
other hot alias-aware path, without adding any benchmark-shaped opcode or
lowering rule. A new guardrail in `interpreter_type_info_cache_test.go` now
pins that `expandTypeAliasesCached(...)` seeds the negative alias-reference
cache entry too. Under the bounded `GOMEMLIMIT=1GiB GOGC=50` spot-checks,
`word_count_small` improves from `1727481234 ns/op`, `315961848 B/op`,
`6364143 allocs/op` to `1688287932 ns/op`, `308919632 B/op`,
`6284120 allocs/op`, while a bounded `nbody_small` control run stays in the
same band at `4273857073 ns/op`, `345851728 B/op`, `12272802 allocs/op`. The
refreshed bounded `word_count_small` CPU profile drops `coerceReturnValue(...)`
again from about `260ms` cumulative to about `170ms`, drops
`expandTypeAliasesCached(...)` from about `80ms` cumulative to about `40ms`,
and cuts the alias-reference lookup inside `coerceReturnValue(...)` from about
`40ms` to about `10ms`. The next follow-up should therefore move on to the now
clearer shared bytecode call/lookup wall (`execCallOpcode(...)`,
`lookupCachedIdentifierNameEntry(...)`, `lookupCachedScopeValue(...)`) plus the
remaining generic `matchesType(...)` work rather than revisiting alias
expansion or inventing benchmark-shaped bytecodes.

That next follow-up is now landed too: the kept lexical-cache change was not a
new benchmark-shaped opcode or a named-stdlib special case, but a lower-level
VM cache layout fix. A narrower direct-mapped overlay on top of the old hashed
lookup maps was explicitly tried and then rejected after bounded regressions, so
the kept tree instead replaces the shared hashed `(program, ip)` path itself.
`globalLookupCache` and `scopeLookupCache` now store per-program
instruction-indexed tables, matching the VM’s other indexed cache families and
removing repeated composite-key hash work from both lookup and store paths
without changing env/owner revision validation semantics. Focused
`bytecode_vm_lookup_cache_test.go` / `bytecode_vm_pool_cache_test.go` slices
stay green under `GOMEMLIMIT=1GiB GOGC=50`. The bounded spot-checks move
`word_count_small` from `1688287932 ns/op`, `308919632 B/op`, `6284120 allocs/op`
to `1655132462 ns/op`, `308924112 B/op`, `6284160 allocs/op`, and move the
`nbody_small` control from `4273857073 ns/op`, `345851728 B/op`, `12272802 allocs/op`
to `4016024170 ns/op`, `345696120 B/op`, `12272354 allocs/op`. The refreshed
bounded `word_count_small` CPU profile then drops
`lookupCachedIdentifierNameEntry(...)` from about `260ms` cumulative to about
`120ms`, `lookupCachedScopeValue(...)` from about `120ms` to about `20ms`, and
`storeCachedScopeValue(...)` from about `80ms` to about `40ms`. The next
follow-up should therefore move on to the remaining shared string-keyed
`call_name` / `call_member` dispatch bookkeeping and residual generic
`matchesType(...)` / coercion work rather than revisiting lexical cache
storage.

That next follow-up is now landed too, but it stayed deliberately inside the
shared direct-call shell rather than reviving the earlier rejected
injected-receiver helper. The kept change reuses existing `bytecodeFrameLayout`
metadata on non-inlined slot-backed calls: `invokeFunction(...)` now uses the
cached slot-layout coercion metadata instead of redoing the broader generic
per-param check loop, sizes slot-backed local env capacity with
`functionLocalBindingCapacityForLayout(...)` so params that stay in slots do
not bloat env capacity, and only pushes implicit-receiver state on that path
when the lowered body actually references `#member`. Focused
call/capacity/member regressions stay green under `GOMEMLIMIT=1GiB GOGC=50`,
including new guardrails that pin bytecode caller-arg stability during
coercion and the slot-backed env-capacity rule. Sequential bounded reruns move
`word_count_small` from the prior `1655132462 ns/op`, `308924112 B/op`,
`6284160 allocs/op` to `1619976174-1630897051 ns/op`,
`308772872-308916744 B/op`, `6283833-6284193 allocs/op`, while the
`nbody_small` control stays effectively flat in a
`3974311393-4048037193 ns/op`, `345737832-345768152 B/op`,
`12272459-12272528 allocs/op` band versus the prior `4016024170 ns/op`,
`345696120 B/op`, `12272354 allocs/op`. The refreshed bounded
`word_count_small` CPU profile still shows the direct-call shell dominated by
nested `vm.run(...)`, but the outer slot-backed binding/setup work shrinks
again and the new `invokeFunctionBindArgsForSlotLayout(...)` helper accounts
for only about `10ms` cumulative on the profiled rerun. The next follow-up
should therefore stay on the remaining generic direct-call shell costs,
especially cached call-shape classification, overload setup,
`matchesType(...)`, and GC scan pressure, rather than revisiting slot-layout
reuse or reopening the rejected injected-receiver shortcut.

That next follow-up is now landed too, and it stayed on the same generality
rule. A broader profile on `persistent_sorted_set_i32_small` showed that the
remaining explicit generic call-local binding work was still paying
merge-aware `Environment.Define(...)` overhead even though these synthetic
`name_type` / `name` bindings live only in a fresh call-local env. A first
rewrite that routed this path through a per-call
`callLocalTypeBindingRuntimeValue` slice was rejected within the same tranche
after it added heap churn and regressed allocations. The kept change is the
narrower one: `bindTypeArgumentsIfAny(...)` now uses direct
`Environment.DefineWithoutMerge(...)` for explicit generic call-local
bindings and a focused regression test now pins both the `T_type` string and
the `T` type-ref payload. Sequential bounded reruns keep `word_count_small`
around `1611743514 ns/op`, `308817880 B/op`, `6283923 allocs/op`, keep the
`nbody_small` control flat at `3941518706 ns/op`, `345803912 B/op`,
`12272605 allocs/op`, and keep `persistent_sorted_set_i32_small` in a noisy
but baseline-level `3345790376-3564166569 ns/op`,
`1223657752-1223862792 B/op`, `14239924-14240266 allocs/op` band. The
refreshed bounded `persistent_sorted_set_i32_small` CPU profile drops
`bindTypeArgumentsIfAny(...)` to about `130ms` cumulative, with the remaining
shared wall now leaning more on generic constraint checks, fresh env
allocation, pattern binding/matching, scope lookup, and GC scan pressure. The
next follow-up should therefore target those broader call-setup and
lookup/match costs rather than synthetic type-binding insertion itself.

That next follow-up is now landed too, again in a narrowed generic form rather
than as a broad new call-site specialization. A persistent-set profile showed
that constrained generic calls were still rebuilding the same reusable call
metadata on every invocation: generic lists, collected constraint specs, and
function-call error-context strings. The first version of the keep routed that
cache through both constrained-generic checks and the ordinary explicit
generic-binding path, but bounded controls showed that unconstrained generic
calls should not pay that extra lookup. The kept version therefore caches only
the constrained-generic metadata:

- function/lambda constrained-generic plans are cached by declaration node
- method-set constrained-generic plans are cached by `*runtime.MethodSet`
- `enforceGenericConstraintsIfAny(...)` and `enforceMethodSetConstraints(...)`
  now reuse those cached plans instead of rebuilding
  `collectConstraintSpecs(...)` and repeated context strings
- explicit generic binding still uses the direct `DefineWithoutMerge(...)`
  path without the broader plan lookup

Focused generic/call/cache regressions stay green under
`GOMEMLIMIT=1GiB GOGC=50`. Sequential bounded reruns keep
`word_count_small` in a `1626379701-1699339117 ns/op`,
`308861384-308890016 B/op`, `6284045-6284128 allocs/op` band, keep
`nbody_small` in a `4075536820-4100655103 ns/op`,
`345461344-345509184 B/op`, `12271814-12271928 allocs/op` band, and move
`persistent_sorted_set_i32_small` into a lower-allocation band around
`3368698530-3373423415 ns/op`, `1194132776-1194222000 B/op`,
`13383376-13383816 allocs/op` versus the prior roughly `1.223GB/op`,
`14.24M allocs/op` band. The refreshed bounded persistent-set profile confirms
the intended shift: `collectConstraintSpecs(...)` drops out of the hot view,
`enforceGenericConstraintsIfAny(...)` lands around `180ms` cumulative, and the
remaining shared wall now leans more on explicit generic binding
(`bindTypeArgumentsIfAny(...)`), fresh env allocation, pattern
binding/matching, scope lookup, and GC scan pressure. The next follow-up
should therefore target those broader generic call-shell costs rather than
revisiting constrained-generic metadata caching itself.

That next follow-up is now landed too and it stayed within the same general
call-shell boundaries. The next persistent-set profile showed that two more
reusable costs were still sitting in the generic call setup:

- repeated generic call sites were still rebuilding the same runtime
  `StringValue` / `TypeRefValue` payloads in `bindTypeArgumentsIfAny(...)`
- simple identifier parameter binds in fresh call-local envs still paid the
  general `assignPattern(...)` / `declareOrAssign(...)` path even though merge
  and fallback semantics are irrelevant there

The keep addresses those broader costs directly:

- explicit generic call-site type-binding payloads are now cached per
  `(function node, call node)` and cleared on the same mutation boundaries as
  the related receiver-derived call-local binding cache
- `bindTypeArgumentsIfAny(...)` now reuses those cached payloads through the
  existing shared runtime-binding application helper
- simple identifier and wildcard parameter patterns now bind straight into the
  fresh call-local env with `DefineWithoutMerge(...)`

Focused regressions stay green under `GOMEMLIMIT=1GiB GOGC=50`. Sequential
bounded reruns move `persistent_sorted_set_i32_small` from the prior roughly
`3.37s`, `~1.194GB/op`, `~13.38M allocs/op` band down to about
`3.20-3.22s`, `~1.182GB/op`, `~11.87M allocs/op`, while
`word_count_small` lands at `1582365124 ns/op`, `308843560 B/op`,
`6284004 allocs/op` and the `nbody_small` control stays effectively flat at
`4026531059 ns/op`, `345591240 B/op`, `12272112 allocs/op`. The refreshed
bounded persistent-set profile confirms the intended shift: the sampled
`bindTypeArgumentsIfAny(...)` time is now almost entirely cache-hit lookup plus
cached binding application instead of payload construction, `assignPattern(...)`
drops out of the hot path, and the next remaining shared wall is more clearly
fresh env allocation (`NewEnvironmentWithValueCapacity(...)`), cached binding
application (`DefineWithoutMerge(...)` into the fresh env), scope
lookup/cache bookkeeping, and GC scan pressure. The next follow-up should
therefore target those shared env/binding/lookup costs rather than revisiting
explicit type-binding caching itself.

That next follow-up is now landed too, and it stayed entirely at the shared
runtime environment boundary rather than adding any new benchmark-shaped VM or
lowering rule.

The persistent-set profile said the remaining binding-side wall was no longer
payload construction itself; it was the repeated replay of current-scope
inserts into brand-new environments. The kept change therefore adds reusable
runtime support instead of a call-site-specific shortcut:

- `runtime.EnvironmentBinding`
- `Environment.DefineWithoutMergeBindings(...)`
- `runtime.NewEnvironmentWithBindings(...)`

The explicit generic call binding caches now store reusable
`[]runtime.EnvironmentBinding` payloads, fresh function/lambda call-local envs
seed those bindings through the shared constructor, and fresh iterator-literal
controller envs in both interpreters use the same batched no-merge path for
their initial bindings. The first cut of the helper briefly double-scanned the
binding slice; that was narrowed within the same tranche to a single-pass
apply loop after the first bounded persistent-set rerun showed a noisy wall-
time regression despite lower allocations.

Focused runtime/interpreter regressions stay green under
`GOMEMLIMIT=1GiB GOGC=50`. Sequential bounded `-benchtime=1x` reruns keep
`word_count_small` roughly flat in a `1.62-1.73s`, `~309MB/op`,
`~6.284M allocs/op` band, while `persistent_sorted_set_i32_small` moves into a
lower-allocation band around `1.162GB/op`, `10.955M allocs/op` versus the
prior `~1.182GB/op`, `~11.87M allocs/op`; after narrowing the helper to a
single pass, reruns settle around `3.38-3.98s` instead of the first noisy
`5.04s` sample.

Implication:

- this keep is still general: it improves shared runtime scope seeding rather
  than encoding a benchmark-only call fusion
- the next reusable wall is now even more clearly fresh `Environment`
  object/map allocation plus the surrounding scope-cache and GC scan pressure

That next follow-up is now landed too, again in a narrower form after a
measured backout inside the same tranche.

The first pass broadened the runtime-side experiment into a two-inline-binding
`Environment` layout so two synthetic bindings could avoid immediate map
promotion entirely. That did reduce allocation counts, but repeated bounded
reruns showed a real wall-time regression: the larger environment object added
enough GC scan and hot-path cost to more than cancel out the map savings.
That variant was backed out and is not part of the kept tree.

The kept generally-applicable pieces are:

- `NewEnvironmentWithValueCapacity(...)` now stores a lazy capacity hint
  instead of eagerly allocating the map
- promotion applies that deferred hint only when a second distinct binding
  actually forces a map
- the bytecode VM now keeps hot per-program entry slices for both
  `scopeLookupCache` and `globalLookupCache`, so steady-state lexical lookups
  avoid repeated `map[*bytecodeProgram]...` lookups

Focused runtime/interpreter regressions stay green under
`GOMEMLIMIT=1GiB GOGC=50`. Sequential bounded `-benchtime=1x` reruns move
`persistent_sorted_set_i32_small` into a `3.13-3.48s`, `~1.1626GB/op`,
`~10.958M allocs/op` band versus the prior `3.38-3.98s`, `~1.162GB/op`,
`~10.955M allocs/op` band, move `word_count_small` from roughly
`1.62-1.73s`, `~309MB/op`, `~6.284M allocs/op` down to roughly
`1.56-1.57s`, `~302MB/op`, `~6.244M allocs/op`, and keep the `nbody_small`
control healthy at about `3.99s`, `312MB/op`, `12.07M allocs/op`.

The refreshed bounded persistent-set profile confirms the intended shift:

- `scopeLookupCacheEntries(...)` drops to about `30ms` cumulative
- `NewEnvironmentWithValueCapacity(...)` drops to about `80ms`
- the next shared wall is more clearly `Environment.RuntimeData()`,
  scope-entry validation itself, and the remaining GC scan pressure

Implication:

- the keep remains generic because it improves shared constructor/caching
  mechanics instead of adding any workload-shaped lowering
- the next reusable tranche should target runtime-data lookup and remaining
  scope-cache validation costs
