# 2026-07-09 Performance Scorecard Tranche

This is a bounded refresh, not a replacement for
`external-scoreboard-current.md`. Each Able measurement uses three runs through
`v12/bench_compare_external` against the current sibling
`../benchmarks/results.json` references. Times are wall-clock seconds; ratios
below `1.00x` are faster than the reference.

| Benchmark | Mode | Able | Go | Able/Go | Ruby | Able/Ruby | Python | Able/Python |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| i_before_e | compiled | 1.2133 | 0.0500 | 24.27x | 0.1000 | 12.13x | 0.1300 | 9.33x |
| i_before_e | bytecode | 0.4533 | 0.0500 | 9.07x | 0.1000 | 4.53x | 0.1300 | 3.49x |
| base64 | compiled | 2.2200 | 2.2000 | 1.01x | 2.2100 | 1.00x | 3.3100 | 0.67x |
| base64 | bytecode | 3.1467 | 2.2000 | 1.43x | 2.2100 | 1.42x | 3.3100 | 0.95x |
| monte_carlo_pi | compiled | 0.1767 | 0.1800 | 0.98x | 1.4200 | 0.12x | 1.6800 | 0.11x |
| monte_carlo_pi | bytecode | 2.3633 | 0.1800 | 13.13x | 1.4200 | 1.66x | 1.6800 | 1.41x |

The compiler floor holds for `base64` and `monte_carlo_pi`, but not the
file-backed `i_before_e` workload. Bytecode misses at least one Ruby/Python
floor in all three rows. The result is intentionally descriptive: three
programs are insufficient to replace the full suite scorecard.

## Paired VM profiles

Both profiles used `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, the serial
executor, a five-iteration warmed `BenchmarkBytecodeProgramRuntime`, and no
trace instrumentation.

- `i_before_e`: `247213898 ns/op`, `9079200 B/op`, `1948 allocs/op`.
  The largest cumulative VM costs were generic call dispatch (`execCallOpcode`,
  59%), member dispatch (`execCallMember`, 33%), member-cache lookup (17%), and
  type matching (7%).
- `monte_carlo_pi`: `2710078099 ns/op`, `177794678 B/op`, `22222104 allocs/op`.
  Float-slot stores and normalization dominate: `execStoreSlotOpcode` (42%),
  cast/divide float stores (41%), and raw float slot normalization (24%).

These profiles do not share a material implementation wall. No VM candidate
was taken from them.

## Rejected primitive-string experiment

The generated `i_before_e` binary profile showed `String.contains` spending
substantial time in byte-array conversion, UTF-8 validation, and GC. A generic
canonical-stdlib experiment redirected valid primitive strings to host substring
search while retaining the old validation path for unchecked invalid Go strings.
It improved compiled `i_before_e` to `0.5867s`, but bytecode regressed to
`0.7867s`; restoring the code recovered bytecode to `0.4533s`. The experiment
and its supporting extern-dispatch changes were fully reverted.

The next compiler candidate, if pursued, must be compiler-local primitive
lowering that preserves the existing stdlib source and bytecode path. It must
be guarded by multiple ordinary string workloads plus `base64` and a numeric
control before it can be kept.

## Kept compiler-local `String.contains` lowering

The follow-up compiler tranche keeps that restricted primitive lowering. Only
the canonical `able.text.string.String.contains` implementation with the exact
primitive carrier signature `string, string -> bool` is eligible. Generated Go
uses `strings.Contains` only when both arguments pass `utf8.ValidString`; the
existing generated Able implementation remains the fallback for unchecked
invalid strings. User-defined `String` types and every bytecode/tree-walker
path remain on ordinary semantics.

CPU-pinned five-run A/B readings (`--cpu-affinity 0`) show the result is not
specific to the file benchmark:

| Workload | Lowered compiled | Restored compiled | Change |
| --- | ---: | ---: | ---: |
| `i_before_e` | 0.784s | 1.680s | 53% faster |
| `string_contains_small` | 0.296s | 0.458s | 35% faster |
| `base64` | 2.294s | 2.296s | neutral |
| `monte_carlo_pi` | 0.198s | 0.202s | neutral |

The three-run external `i_before_e` sample was `0.4800s` compiled and
`0.4933s` bytecode. It remains well short of Go, but the compiler improvement
is real while bytecode stays in its normal band. Focused compiler rendering and
interpreter return/call guards pass. The next performance work should profile
the remaining compiled file/text path after this allocation wall, while
continuing to require these text/bytes/numeric controls.

## Kept compiler-local `String.len_bytes` lowering

The post-`contains` paired profiles showed one further shared primitive cost:
canonical `String.len_bytes` plus `validated_bytes` accounted for 400 ms (70%)
of a 570 ms `i_before_e` profile and 50 ms (25%) of the independent
`string_contains_small` profile. The compiler now recognizes only the exact
`able.text.string.String.len_bytes` primitive signature `string -> uint64`.
For valid UTF-8 strings of at most `2,147,483,647` bytes it emits
`uint64(len(self))`; invalid and out-of-range inputs retain the generated Able
method body. The range guard preserves the stdlib's `i32` byte-array-length
semantics. No external stdlib, bytecode, or tree-walker source changed.

| Workload | Lowered compiled | Restored compiled | Result |
| --- | ---: | ---: | --- |
| `i_before_e` | 0.122s | 0.674s | 82% faster |
| `string_contains_small` | 0.222s | 0.484s | 54% faster |
| `base64` | 2.297s | 2.323s | neutral |
| `monte_carlo_pi` | 0.197s | 0.193s | neutral |

All readings are CPU-pinned and use five runs for text workloads and three for
controls. Captured stdout from candidate and restored binaries was identical
for every text run. A new three-run external `i_before_e` check measured
compiled `0.0933s` (Go `0.0500s`, Ruby `0.1000s`, Python `0.1300s`) and
bytecode `0.4600s`; the bytecode source path is unchanged. This materially
narrows the compiled file/text gap but does not yet meet the 95%-of-Go target.

The resulting `i_before_e` profile is too short for stable attribution. The
longer independent string workload now emphasizes `String.replace`, but it is
not material in `i_before_e`; it is therefore not a valid next lowering target.
Refresh a broader compiled text/bytes slice and require a cost to recur in at
least two ordinary programs before pursuing another primitive boundary.

## Broader compiled text/bytes reconnaissance — no keep

A one-run external sweep widened the evidence beyond the string fixtures:

| Benchmark | Compiled Able | Go | Observation |
| --- | ---: | ---: | --- |
| `base64` | 2.270s | 2.200s | near Go; host crypto/base64 kernel dominates |
| `json` | 0.700s | 1.360s | faster than Go; host numeric parser dominates |
| `i_before_e` | 0.070s | 0.050s | close enough to be too short for useful profiling |
| `reverse_complement` | 0.080s | 0.010s | short run; no reusable sample yet |
| `k_nucleotide` | 3.100s | n/a | generic raw-map operations and prefix checks dominate |
| `quicksort` | 2.520s | 2.010s | non-text control |
| `tapelang_alphabet` | 3.730s | 1.750s | program-defined tape operations dominate |

CPU-pinned generated-binary profiles established that these are not one shared
compiler wall. K-nucleotide spent about 1.17s in `HashMap.raw_get`, 0.93s in
`HashMap.raw_set`, and 0.84s in canonical `String.starts_with` / validation;
JSON spent 0.49s in its host field-mean parser and 0.21s in `strconv.ParseFloat`;
Tapelang spent 1.02s in its application-defined `Tape.inc`; and base64 spent
0.99s encoding, 0.80s decoding, and 0.35s in MD5. An independent word-count
program was too short for reliable map attribution and instead sampled
`String.split`.

No lowering was attempted. `String.starts_with` appears materially in only the
k-nucleotide program, while a HashMap-specific compiler branch would both fail
the repeatability bar and violate the nominal-container lowering policy. The
next measurement should pair k-nucleotide with a longer ordinary generic-map
program and look only for a shared non-nominal conversion or runtime boundary.

## Paired HashMap representation trial — rejected

The normal `hashmap_i32_small` shape has only 3,000 lookups, so it was
temporarily scaled from 1,000 to 20,000 entries for profiling without changing
its algorithm; the fixture was restored immediately after measurement. A
200,000-entry attempt exceeded the 50-second bounded guard. The 20,000-entry
profile and k-nucleotide both showed linear scanning of
`HashMapValue.Entries` after each hash calculation.

A collision-safe bucket index in the shared kernel representation was tried,
including generated compiled applications plus tree-walker and bytecode paths.
It preserved insertion iteration order and retained equality checks within a
hash collision bucket. Focused collision/removal, interpreter HashMap, and
compiled HashMap execution tests passed. The large-map result was strong:

| Workload | Indexed trial | Restored baseline | Result |
| --- | ---: | ---: | --- |
| scaled `hashmap_i32` (20,000 entries) | 0.140s | 0.977s | 86% faster |
| external `k_nucleotide` | 4.077s | 3.85s | 6% slower |

Candidate and restored scaled-map stdout matched exactly. The representation
was nevertheless fully reverted: k-nucleotide exercises only 4/16-entry maps
at very high frequency, so the index bookkeeping or lazy-index branch costs
more than its short scans. Do not retry this design by merely retuning its
threshold. Any future map work must reduce a cost shared by both small and
large maps without adding a lookup-time representation decision.

## Non-map cross-suite profile refresh — no keep

Pinned compiled quicksort measured 2.030s. Its profile spent 1.50s
cumulatively in the program's recursive sort, with 0.30s flat in a checked
signed multiply used by decimal parsing. Pinned matrix multiplication measured
1.180s and spent 98% of samples in its floating-point matrix kernel; it did
not repeat the checked-integer cost. The previously captured Tapelang compiled
profile is dominated by the program-defined `Tape.inc` and `Tape.get` methods.
The external bytecode Tapelang run exceeded the 50-second guard, so it did not
produce an actionable profile.

No source change was attempted. The evidence rejects a quicksort parsing
shortcut, a `Tape` lowering, or a checked-multiply relaxation without a
cross-program proof. The next bytecode investigation should profile the
feature-equivalent local `tapelang_small` beside an array-heavy local control,
then use external Tapelang and quicksort only as final guards.

## Local Tapelang/array VM follow-up — generic VM correctness repair

The minimal `+o` probe exposed a VM error rather than an application-specific
hotspot: a reused transient scope could retain a cached outer `idx` value after
the owner had been assigned. This made slotless loop functions reload zero and
spin. Scope-cache refresh now validates the owner revision, using the direct
single-thread revision accessor so the repair does not add a lock or regress
the ordinary bytecode hot path. The inline-return fallback also now preserves
the actual returned value when its fast no-coercion check misses.

The real fixture now completes with checksum `2048` in both bytecode and
tree-walker modes. Its bounded bytecode runtime is `980061 ns/op`, `67955
B/op`, and `377 allocs/op` (10 iterations). This is a general lexical-scope
and call-return repair: no `Tape` lowering, parser shortcut, or stdlib change
was added.

The completed local controls used the same bounded runtime settings:

| Workload | Warmed reading | Shared finding |
| --- | ---: | --- |
| `array_map_i32_small` | 60,402,299 ns/op; 855,771 B/op; 327 allocs/op | Calls include program switching and integer store work. |
| `array_filter_i32_small` | 73,819,282 ns/op; 1,018,184 B/op; 314 allocs/op | `execCallOpcode` is 43% cumulative, principally member/call dispatch. |
| `linked_list_iterator_collect_i64_small` | 239,124,191 ns/op; 3,254,764 B/op; 29,094 allocs/op | `execCallOpcode` is 62% cumulative, with member calls, inline setup, and returns. |

The refreshed control readings after the repair were `58060034 ns/op`,
`852591 B/op`, `321 allocs/op` for map; `78321152 ns/op`, `1015005 B/op`,
`308 allocs/op` for filter; and `234086852 ns/op`, `3248395 B/op`, `29082
allocs/op` for collection. `bytecodeRawIntegerValueInfo` remains only 1--4% of
the profiles. The next performance candidate still needs the same generic
subpath to repeat across independently shaped programs.

## Member/dispatch refresh — no keep

Fresh bounded one-process CPU profiles tested whether the remaining generic
call parent concealed a shared member/field wall:

| Workload | Runtime sample | Call composition |
| --- | ---: | --- |
| `string_builder_small` | 245,370,955 ns/op (12x) | `execCallOpcode` 56%; string-byte iterator `next` and cached-member lookup dominate. |
| `array_filter_i32_small` | 71,991,444 ns/op (30x) | `execCallOpcode` 39%; generic member dispatch, exact-native array calls, inline setup, and returns split the cost. |
| `tapelang_small` | 938,404 ns/op (2500x) | `execCallOpcode` 51%; array-slot/member calls, name lookup, and inline named calls dominate. |

The cache is not missing in the two cache-heavy programs: instrumented
single-run samples recorded 373,710 hits / 21 misses in the builder and 40,003
/ 11 in filter. Tapelang recorded 317 / 66, consistent with its much smaller
method-call volume. The same parent therefore represents different dispatch
lanes, not a safe common optimization target. No runtime, compiler, or stdlib
change was kept.

The next slice should intentionally profile two ordinary programs sharing one
dispatch lane, then validate a generic candidate against a different lane plus
Tapelang and quicksort guards.

## Exact-native and inline nominal lane follow-up — no keep

Two deliberately paired follow-ups both rejected their apparent candidate.

`array_filter_i32_small` records 40,000 resolved exact-native member calls,
but `array_map_i32_small` records none. Its fresh 30-iteration profile measured
62,053,199 ns/op, 849,158 B/op, and 314 allocations/op; the exact-native target
resolver itself was only 0.46% of the filter profile. Caching or reshaping that
resolver would therefore neither span the pair nor remove a material cost.

The inline nominal-method pair likewise diverged:

| Workload | Bounded sample | Relevant attribution |
| --- | ---: | --- |
| `linked_list_iterator_collect_i64_small` | 235,371,739 ns/op; 3,223,147 B/op; 29,036 allocs/op (10x) | calls 56%; `inlineResolvedCallEnvForBindings(...)` 6.4% |
| `hashmap_i32_small` | 19,522,473 ns/op; 1,582,538 B/op; 9,044 allocs/op (120x) | `hashMapFindEntryWithHash(...)` 14%; exact-native calls 18%; inline environment setup 2.1% |

The collector spends materially in inline generic-call setup; `HashMap` does
not. Adding an inline-binding cache based on the first profile would target a
single workload, while a named-container fast path would violate the nominal
lowering policy. No runtime, compiler, or `able-stdlib` change was made.

The next tranche should refresh a broader bounded profile slice and select a
candidate only when the same generic VM leaf, rather than merely a call parent,
repeats materially in independently shaped programs. Retain Tapelang and
quicksort as different-lane guards for any eventual candidate.

## Wider VM profile sweep and generic-pattern trial — no keep

The requested wider bounded slice used the same one-process guardrails across
four different families:

| Workload | Profiled runtime sample | Repeated match attribution |
| --- | ---: | --- |
| `string_split_join_small` | 936,843,798 ns/op; 51,779,240 B/op; 520,018 allocs/op (2x) | `JumpIfNotTypedPattern` 16.7%; matcher 11.8% |
| `linked_list_iterator_collect_i64_small` | 230,356,500 ns/op; 3,223,163 B/op; 29,036 allocs/op (10x) | 12.2%; matcher 6.6% |
| `array_map_i32_small` | 63,467,230 ns/op; 848,986 B/op; 313 allocs/op (35x) | 6.3%; `matchesTypeWithoutRuntimeValue(...)` 4.5% |
| `byte_histogram_small` | 123,313,738 ns/op; 167,769 B/op; 9,791 allocs/op (15x) | 22.8%; matcher 14.7% |

This was a genuine cross-family VM wall, so a narrowly scoped semantic cache
was tested: lowering marked only typed patterns whose declared generic or
`Self` type makes the existing `matchesTypeWithoutRuntimeValue(...)` result
unconditionally true. The VM then preserved the existing stack-snapshot path
without re-running that query. It was generic to declared type variables and
did not target a particular container, function, or benchmark.

Same-session restoration rejected it:

| Workload | Candidate | Restored | Result |
| --- | ---: | ---: | --- |
| `string_split_join_small` | 921,636,060 ns/op | 924,659,285 ns/op | neutral |
| `linked_list_iterator_collect_i64_small` | 236,265,405 ns/op | 231,068,183 ns/op | 2.2% slower |
| `array_map_i32_small` | 58,795,011 ns/op | 58,020,312 ns/op | 1.3% slower |
| `byte_histogram_small` | 122,743,073 ns/op | 129,056,976 ns/op | 4.9% faster |

Allocation shapes were unchanged. The candidate therefore helps one
match-heavy family but loses two representative controls, so every source and
test change was reverted. No compiler or `able-stdlib` work is justified by
this VM-only result.

Do not retry the generic-pattern metadata marker as a small refactor. The next
profile slice should move beyond typed-pattern dispatch and identify a larger
leaf that recurs in non-pattern numeric/map/graph code plus an unrelated
string or iterator guard.

## Non-pattern raw-integer extraction follow-up — no keep

The non-pattern Dijkstra/Heap, BitSet, and StringBuilder profiles exposed one
shared VM leaf without relying on the earlier match-heavy workloads:

| Workload | Profiled sample | `bytecodeRawIntegerValueInfo(...)` |
| --- | ---: | ---: |
| `dijkstra_heap_small` | 30,768,588 ns/op; 561,513 B/op; 13,925 allocs/op (50x) | 6.5% flat |
| `bit_set_small` | 836,127,219 ns/op; 25,573,504 B/op; 1,162,217 allocs/op (2x) | 5.4% flat |
| `string_builder_small` | 266,371,315 ns/op; 1,712,973 B/op; 79,576 allocs/op (8x) | 6.1% flat |

The leaf's callers were importantly different. Dijkstra and StringBuilder
reached it mostly from raw slot storage; BitSet reached it mostly through
generic integer conversion and binary evaluation. A generic candidate moved
the ordinary small `runtime.IntegerValue` cases to the first type switch,
ahead of every VM-internal raw carrier. It preserved every carrier and passed
focused raw-integer/type-pattern coverage, but the extra failed assertion on
raw values outweighed any boxed-integer win:

| Workload | Candidate | Restored | Result |
| --- | ---: | ---: | --- |
| `dijkstra_heap_small` | 35,322,568 ns/op | 33,729,533 ns/op | 4.7% slower |
| `bit_set_small` | 950,281,010 ns/op | 838,481,457 ns/op | 13.3% slower |
| `string_builder_small` | 264,407,299 ns/op | 262,257,772 ns/op | 0.8% slower |

All allocation counts were effectively unchanged. The candidate was completely
reverted; it did not justify an `able-stdlib` change. The lesson is that a
shared leaf name is not enough: its dominant carriers and callers must also
repeat. Do not revisit global `bytecodeRawIntegerValueInfo(...)` type-switch
ordering. Profile generic integer binary/store caller paths next only when the
same caller recurs across independently shaped workloads.

## Raw-slot cell pool reuse — kept

The next caller-level check found the required alignment. Dijkstra/Heap,
StringBuilder, and ByteHistogram all execute the generic slot-store sequence
`execStoreSlot(...)` -> `tryStoreRawIntegerSlotValue(...)`; its sampled share
was 4.6%, 7.1%, and 2.2%, respectively. BitSet exercises different raw
conversion/binary callers and served as a guard.

The VM already returns raw i64 and general raw-integer slot cells to per-VM
pools as callee slot frames are released. The first store into an eligible slot
did not draw from those pools, however: `storeRawI64Slot(...)` and the general
non-cached raw-integer branch allocated a cell directly. Both now acquire the
existing pooled cell. Reuse of a matching live slot, invalid-target result
behavior, raw stack ownership, integer kinds, and normal boxing paths are
unchanged. Focused tests cover reuse of both the i64 and large unsigned raw
cell paths.

Same-session A/B measured broad allocation and runtime wins:

| Workload | Pooled candidate | Restored control | Result |
| --- | ---: | ---: | --- |
| `dijkstra_heap_small` | 30,370,087 ns/op; 346,344 B/op; 9,858 allocs/op | 30,590,240 ns/op; 560,877 B/op; 13,924 allocs/op | 0.7% faster; 38% fewer bytes; 29% fewer allocs |
| `string_builder_small` | 238,895,206 ns/op; 1,708,716 B/op; 79,560 allocs/op | 243,600,816 ns/op; 1,709,062 B/op; 79,570 allocs/op | 1.9% faster |
| `byte_histogram_small` | 123,519,601 ns/op; 165,240 B/op; 9,784 allocs/op | 133,974,544 ns/op; 165,287 B/op; 9,786 allocs/op | 7.8% faster |
| `bit_set_small` | 818,644,225 ns/op; 22,821,720 B/op; 1,121,882 allocs/op | 848,446,602 ns/op; 25,557,816 B/op; 1,162,189 allocs/op | 3.5% faster; 11% fewer bytes |

This is a shared VM frame-allocation fix, not a benchmark-shaped optimization:
it benefits every slot-indexed function that stores eligible raw integers after
a prior frame has released reusable cells. No compiler, tree-walker, or
`able-stdlib` source changed. Next, refresh the generic integer binary/store
profiles on the kept allocation state and pursue a dispatch change only if the
same caller and carrier family recur across independently shaped workloads.

## Post-pool binary/store refresh — no keep

Fresh profiles on the kept pool-reuse state tested whether the remaining
generic integer binary/store cost now shared a concrete dispatch lane:

| Workload | Profiled sample | Binary/store composition |
| --- | ---: | --- |
| `dijkstra_heap_small` | 37,107,329 ns/op; 346,979 B/op; 9,859 allocs/op (50x) | `execBinary(...)` 8.2%; direct same-small-integer pairs about 2.2%; raw store 3.3% |
| `bit_set_small` | 996,062,961 ns/op; 22,837,400 B/op; 1,121,910 allocs/op (2x) | `execBinary(...)` 19.2%; `ApplyBinaryOperatorFast(...)` 20.7%; bitwise evaluation 11.1% |
| `string_builder_small` | 285,360,979 ns/op; 1,712,625 B/op; 79,567 allocs/op (8x) | iterator/member/field work dominates; binary 5.3%, store 4.9%, direct pairs about 1.3% |
| `byte_histogram_small` | 167,792,982 ns/op; 167,657 B/op; 9,786 allocs/op (15x) | iterator/index/type-pattern work dominates; binary 4.4%, raw store 4.4%, direct pairs about 1.2% |

The broad `execBinary(...)` parent represents divergent children: BitSet goes
through generic bitwise operator evaluation, Dijkstra primarily hits
slot-constant/direct raw pairs, and the string workloads spend more time in
dispatch, indexing, and typed-pattern machinery. The direct raw-pair helper is
too small and does not repeat in the BitSet-dominant lane. No source change was
attempted; a parent-only rewrite would violate the cross-workload criterion.

Next, profile two independently written bitwise/numeric kernels—such as
BitSet and CRC32—and require the same concrete operator and carrier lane before
considering a generic dispatch change. Keep Dijkstra and StringBuilder as
different-lane guards. No `able-stdlib` change is justified by this result.

## Dotted raw-integer bitwise lane — kept

The BitSet and CRC32 profiles met the cross-workload bar: both repeatedly
entered `execBinary(...)` for dotted bitwise operators with two raw, same-kind
small integers. CRC32 spent 49.5% in generic bitwise evaluation, while BitSet
spent 20.7% in `ApplyBinaryOperatorFast(...)` and 11.1% in bitwise evaluation.

Lowering marks only `.&`, `.|`, `.^`, `.<<`, and `.>>`. The VM uses that mark
to evaluate valid same-kind integer carriers up to 64 bits directly as raw bit
patterns. It preserves the generic path for all other operators and values,
including non-dotted operations, mixed kinds, boxed/big values, wider integer
types, invalid shifts, and overflow errors. That instruction-local marker is
important: the rejected unconditional check had made ordinary binary work,
including StringBuilder, slower.

Same-session A/B (bounded to one process with `GOMEMLIMIT=1GiB`, `GOGC=50`,
and `GOMAXPROCS=1`) supports keeping the scoped generic path:

| Workload | Marked candidate | Restored control | Result |
| --- | ---: | ---: | --- |
| `crc32_small` | 987,565,584 ns/op; 13,862,328 B/op; 288,057 allocs/op (2x) | 1,469,649,058 ns/op; 13,862,188 B/op; 288,055 allocs/op | 32.8% faster |
| `bit_set_small` | 805,516,978 ns/op; 22,821,712 B/op; 1,121,882 allocs/op (2x) | 817,967,702 ns/op; 22,821,752 B/op; 1,121,883 allocs/op | 1.5% faster |
| `dijkstra_heap_small` | 28,539,341 ns/op; 346,344 B/op; 9,858 allocs/op (50x) | 28,821,682 ns/op; 346,344 B/op; 9,858 allocs/op | 1.0% faster |
| `string_builder_small` | 236,972,154 ns/op; 1,708,734 B/op; 79,561 allocs/op (8x) | 236,724,022 ns/op; 1,708,698 B/op; 79,559 allocs/op | statistically unchanged (0.1% slower) |

A final candidate CRC32 CPU profile retained 26.1% cumulatively in the direct
same-kind raw bitwise helper, replacing the dominant generic bitwise evaluator.
Focused raw-result and lowering-marker coverage passes. The current dirty
workspace's full package and strict bytecode fixture matrix remain red on
independent Array/call-member/raw-i32 expectations and a
`TestExecFixtureParity` timeout; this dotted-bitwise lane is not on those
paths. No compiler, tree-walker, named-container behavior, or `able-stdlib`
source changed.

Next, refresh profiles for additional independent numeric/bitwise programs on
this state. Pursue raw extraction or map lookup only when the same caller and
carrier family repeats outside the current pair.

## Post-bitwise refresh — no keep

`k_nucleotide_small` provided a third, independently written rolling `u64`
bitwise kernel. It uses the same dotted shifts, ANDs, and ORs as the prior pair,
but, after the kept lane, direct raw bitwise evaluation accounted for only 1.4%
flat of its 8,445,768 ns/op profile (250 iterations). Calls, returns, and map
work are now larger there, so widening the bitwise lane would not meet the
cross-workload bar.

The next numeric profiles also ruled out a cast rewrite: `sum_u32_small`
(1,902,774,336 ns/op; 2 iterations) spent 29.0% cumulatively in
`execCastOpcode(...)`, while `byte_histogram_small` (119,703,243 ns/op; 15
iterations) did not repeat that wall. Its largest sampled work was string-byte
iteration, struct fields, and array indexing.

Array Fold and Array Map did share one raw-i32 leaf, respectively 7.3% and
10.4% flat in fresh profiles: `bytecodeRawI32SlotCachedValue(...)`. Its callers
were generic arithmetic-result and primitive-array read paths. A general trial
returned a freshly converted raw-i32 carrier directly rather than using the
bounded cached carrier. It eliminated the lookup but replaced allocation-free
reuse with interface boxes, so it failed the broad performance requirement:

| Workload | Direct-carrier candidate | Restored cache | Result |
| --- | ---: | ---: | --- |
| `array_fold_i32_small` | 55,580,720 ns/op; 1,360,035 B/op; 327,022 allocs/op (25x) | 54,780,333 ns/op; 235,756 B/op; 45,959 allocs/op | 1.5% slower; 5.8x bytes; 7.1x allocs |
| `array_map_i32_small` | 68,327,218 ns/op; 2,359,165 B/op; 378,138 allocs/op (35x) | 61,044,372 ns/op; 847,840 B/op; 304 allocs/op | 11.9% slower; 2.8x bytes; 1,244x allocs |

The candidate was fully reverted. No compiler, tree-walker, named-container,
or `able-stdlib` source changed. Do not replace the bounded raw-i32 cache with
direct interface conversion. Next, profile a larger generic call/return,
primitive-array read, or exact map-lookup leaf across workloads from distinct
benchmark families before making another change.

## Cross-key map lookup — no keep

The map-lookup check deliberately paired distinct key representations rather
than optimizing the existing primitive benchmark in isolation:

| Workload | Profiled sample | Lookup result |
| --- | ---: | --- |
| `hashmap_i32_small` | 18,940,326 ns/op; 1,582,583 B/op; 9,044 allocs/op (120x) | `hashMapFindEntryWithHash(...)` 10.6% flat, reached from raw lookup and insertion. |
| `word_count_small` | 874,143,774 ns/op; 32,332,428 B/op; 354,613 allocs/op (2x) | Entry search only 0.6% flat, reached during insertion; call/return, strings, and allocation dominate. |

Although both workloads enter `execCallOpcode(...)`, they do not share a
material lookup or hashing leaf. A primitive-key fast path would be a named
representation optimization, while a generic lookup rewrite would tax the
string-heavy program without addressing its current wall. No runtime,
compiler, tree-walker, or `able-stdlib` source changed.

Next, deliberately pair call/return-heavy programs and require the same
concrete frame setup, frame release, or return-coercion operation across both
before considering a shared VM dispatch change.

## Aligned call/return slice — no keep

K-Nucleotide and WordCount both use direct slot-argument inline calls, so this
pair tested a concrete dispatch operation rather than the broad call parent:

| Operation | `k_nucleotide_small` | `word_count_small` | Result |
| --- | ---: | ---: | --- |
| `tryInlineCachedCallNameDirectFromSlots(...)` | 5.2% cumulative | 6.9% cumulative | shared, but modest |
| `pushCallFrame(...)` / `popCallFrameFields(...)` | 5.7% / 3.8% cumulative | 4.6% / 2.9% cumulative | shared, but modest |
| return work | `finishInlineReturn(...)` 16.1% cumulative | program return coercion 6.9% cumulative | divergent dominant paths |

The only shared operations are modest frame/slot mechanics. The material
return paths differ, and the two nearby generic ideas—mixed-coercion raw-cell
preservation and return-guard reordering—already failed broad guard benchmarks.
Reopening either would optimize noise rather than a shared wall. No runtime,
compiler, tree-walker, or `able-stdlib` source changed.

Next, pair primitive-array reads across distinct element types and outer
program shapes. Change dispatch only if the same direct-read or materialization
leaf repeats materially.

## Primitive-array reads — no keep

The direct-read check paired an `Array<u32>` checksum with an `Array<i32>`
arithmetic fold. Both route indexed reads through
`ArrayStoreMonoPrimitiveReadInfoIntoFresh(...)`, but the concrete common leaf
is too small to justify VM-wide changes:

| Workload | Read-path evidence | Why no change |
| --- | --- | --- |
| `sum_u32_small` | indexed-read path 10.8% cumulative; shared materialization leaf 2.9% flat | `execCastOpcode(...)` is its 29.0% dominant wall. |
| `array_fold_i32_small` | indexed-read path 12.8% cumulative; shared materialization leaf 1.2% flat | arithmetic, member-slot calls, and raw-i32 work dominate. |

The same leaf exists, but is not material in both workloads. A generic array
cache or dispatch rewrite would put cost on every primitive array without
addressing the larger divergent work. No runtime, compiler, tree-walker, or
`able-stdlib` source changed.

Next, refresh the broader corpus and select a concrete caller/callee or
allocation leaf that is material in at least two benchmark families, rather
than optimizing another shared parent.

## Cross-family allocation scan — no keep

Allocation profiles of the primitive HashMap and string WordCount workloads
also diverge rather than exposing a reusable VM allocation wall:

| Workload | Allocation profile | Result |
| --- | --- | --- |
| `hashmap_i32_small` | `assignPatternExpression(...)` 42.2% and declaration-name analysis 40.2% of sampled allocation | Pattern/declaration machinery dominates. |
| `word_count_small` | positional struct construction 27.6%, interface coercion 9.6%, array/string backing allocation | String/container work dominates. |

The small-int cache appears in both profiles only as benchmark-process
initialization, not a per-program allocation site. There is no shared
allocation leaf, so a map-, String-, or struct-specific optimization would
violate the generality rule. No runtime, compiler, tree-walker, or
`able-stdlib` source changed.

Next, act only on a newly repeated concrete leaf. If the broader corpus still
does not expose one, refresh the cross-runtime scorecard and direct work toward
the materially lagging benchmark family rather than tuning noise.

## Cross-runtime refresh and float/text triage — no keep

A fresh three-run external scorecard was collected without changing the
checked-in broader scoreboard. It identifies the current relative gaps while
keeping compiler and bytecode goals separate:

| Benchmark | Compiled Able / Go | Bytecode Able / Ruby | Bytecode Able / Python |
| --- | ---: | ---: | ---: |
| `i_before_e` | 0.1200s / 0.0500s (2.40x) | 0.5900s / 0.1000s (5.90x) | 0.5900s / 0.1300s (4.54x) |
| `base64` | 2.4600s / 2.2000s (1.12x) | 3.0333s / 2.2100s (1.37x) | 3.0333s / 3.3100s (0.92x) |
| `monte_carlo_pi` | 0.2100s / 0.1800s (1.17x) | 2.6967s / 1.4200s (1.90x) | 2.6967s / 1.6800s (1.61x) |

Monte Carlo was initially selected because it is the largest numeric bytecode
miss in this slice. Its one-process CPU profile measured `2441319938 ns/op`,
`177854672 B/op`, and `22222194 allocs/op`. The largest concrete VM work was
the specialized float cast/divide path (35.8% cumulative), raw-float store
normalization (26.3%), and
`execJumpIfFloatMulAddMulCompareConstFalse(...)` (15.2%).

Two independent float programs were then used to distinguish a generic
opportunity from a Monte-specific one. N-body measured `3706879358 ns/op`,
`176993496 B/op`, and `10615878 allocs/op`; it is instead dominated by general
call execution, lookup, frame management, and array access. Mandelbrot did
repeat the float comparison opcode (21.2% cumulative) and float store traffic,
but its `652266768 ns/op`, `55719288 B/op`, `6862841 allocs/op` profile lands
on the exact raw-float compare/store boundary that has already had multiple
broad f64-representation, owned-cell, and direct-decode candidates rejected.
The third program therefore confirms existing coverage rather than authorizing
another micro-variant of that lane.

The scorecard's worst bytecode miss, file-backed `i_before_e`, was also
refreshed. Its fixture profile measured `94346033 ns/op`, `4835044 B/op`, and
`46395 allocs/op` over 30 iterations. It is call/return and type-pattern
dominated: `execCallOpcode(...)` is 48.1% cumulative,
`finishInlineReturn(...)` 20.5%, and `execJumpIfNotTypedPattern(...)` 11.7%.
Those concrete paths were already deliberately paired with WordCount, iterator,
numeric, and string guards; the nearby generic variants either diverge by
workload or have been broad-benchmark rejects. There is no isolated substring
or file-read VM leaf here that could justify a text-specific fast path.

No runtime, compiler, tree-walker, or `able-stdlib` source changed. The next
tranche should refresh a broader external scorecard in comparable family-sized
chunks, then act only on a newly repeated concrete VM leaf. Compiler work must
pair a file/text program with a non-text primitive guard before adding another
primitive-boundary lowering; no nominal-container or benchmark-specific rule is
permitted.

## Broader external text/bytes slice — no keep

The next three-run external scorecard chunk deliberately mixed JSON parsing,
byte-array file transformation, and a language interpreter. JSON is already
competitive in both modes; the gaps are reverse-complement and compiled
Tapelang:

| Benchmark | Compiled Able / Go | Bytecode Able / Go | Outcome |
| --- | ---: | ---: | --- |
| `json` | 0.7533s / 1.3600s (0.55x) | 1.2133s / 1.3600s (0.89x) | already ahead of Go |
| `reverse_complement` | 0.1400s / 0.0100s (14.00x) | 5.9600s / 0.0100s (596.00x) | largest completed gap |
| `tapelang_alphabet` | 3.7567s / 1.7500s (2.15x) | all 3 runs exceeded 55s | bounded-harness timeout, not a comparative score |

The reverse-complement fixture was too short for CPU sampling, so the profile
used the unchanged external program and its normal corpus. It measured
`6280784722 ns/op`, `868018904 B/op`, and `10913206 allocs/op`. The material
work is canonical array-slot access and result traffic:

| Profile lane | Sampled share |
| --- | ---: |
| `execCallMemberArraySlot(...)` | 26.8% cumulative |
| `appendSlotStackValueChecked(...)` | 19.7% cumulative |
| `execArrayPushMemberFast(...)` | 13.6% cumulative |
| `execCallOpcode(...)` | 33.6% cumulative |

The independent byte-transform guard, `byte_histogram_small`, does not repeat
that lane. Its `131830366 ns/op`, `167666 B/op`, `9786 allocs/op` profile is
instead driven by canonical String-byte iterator calls (21.8% cumulative),
typed-pattern dispatch (23.9%), and indexed i32 counter writes (12.2%). The
common `execCallOpcode(...)` parent is insufficient evidence for a dispatch
rewrite.

This specifically rules out a DNA/file/`Array u8` shortcut. It also does not
justify revisiting canonical array-slot cache/proof changes, raw-carrier
representations, or stack-append micro-variants: those already failed broader
array, string, list, or external-workload guards. No runtime, compiler,
tree-walker, or `able-stdlib` source changed.

Next, pair reverse-complement with another external-scale direct primitive-array
transform. Consider a compiler or VM candidate only when that pair identifies
the same primitive language boundary and preserves an unrelated text/bytes
guard; do not add a named-container or program-shaped rule.

## Reverse-complement primitive-array pairing — no keep

The next external-scale pairing started with K-Nucleotide because it is an
independent file-backed `Array<u8>` program that also calls canonical
`read_slot(...)`. One bytecode profile iteration exceeded the project
55-second cap and was interrupted; it is not used as a performance result.
The cap-compliant independent guard was full external Mandelbrot, which writes
an ordinary `Array<u8>` output but has completely different numeric work.

| Workload | Bounded profile | Material lane |
| --- | --- | --- |
| `reverse_complement` | 6,280,784,722 ns/op; 868,018,904 B/op; 10,913,206 allocs/op | array-slot reads 26.8% cumulative, stack append 19.7%, array push 13.6% |
| `mandelbrot` | 7,876,603,111 ns/op; 618,921,256 B/op; 76,303,136 allocs/op | float compare jump 20.4%, normalized float stores 14.3%; stack append only 5.7% |

Mandelbrot does not have a material array-push or canonical `read_slot` lane,
so it cannot validate an array primitive candidate. The shared VM parent and a
small stack-append slice are insufficient, particularly because the prior
array-slot cache/proof, raw-carrier, float-store, and stack-append micro-cuts
already failed broader guards. No generated compiler profile was taken: there
is no shared primitive language boundary to investigate.

No runtime, compiler, tree-walker, or `able-stdlib` source changed. Next,
return to the external scorecard and choose a family with two cap-compliant,
independently shaped programs before profiling another candidate. Keep JSON
and Base64 as controls; do not add a primitive-array, named-container, file,
or program-shaped specialization.

## Structural cap-compliant pair and compiler triage — no keep

The BinaryTrees idea was rejected before measurement because full external
bytecode is a known timeout. Sudoku and MatrixMultiply form a cap-compliant
nested-array pair with different algorithms. Their three-run scorecard shows
that neither needs bytecode work for the Python/Ruby objective:

| Benchmark | Compiled Able / Go | Bytecode Able / Ruby | Bytecode Able / Python |
| --- | ---: | ---: | ---: |
| `sudoku` | 0.1333s / 0.1300s (1.03x) | 0.5100s / 5.6700s (0.09x) | 0.5100s / 3.0200s (0.17x) |
| `matrixmultiply` | 1.1400s / 0.8800s (1.30x) | 3.8933s / 42.9300s (0.09x) | 3.8933s / 56.2900s (0.07x) |

The only remaining competitive concern in this pair is compiled MatrixMultiply.
Its generated-binary CPU profile has 98.1% of samples in the already-direct
primitive f64 `__able_compiled_fn_matmul` loop. Sudoku completes too quickly
for a deep compiler profile; its small sample is instead String-byte validation
and allocation, not matrix arithmetic. Thus the two programs do not expose a
shared lowering helper or runtime primitive boundary.

No compiler, runtime, tree-walker, or `able-stdlib` source changed. The matrix
result does not authorize a Matrix/nested-array/nominal-container special case;
any future primitive f64 work must first repeat in another ordinary numeric
program and improve the general lowering path. Next, select a new
cap-compliant pair with a common concrete generic runtime helper before
profiling candidates, while retaining JSON, Base64, Sudoku, and MatrixMultiply
as scorecard controls.

## Full opcode and call-trace discovery — no keep

The next discovery pass deliberately audited lowering before taking another CPU
profile. The fresh full external audit covers 15 benchmarks, 78 lowered
functions, and 4,076 instructions. The only repeated speculative float opcode
is still `JumpIfFloatMulAddMulCompareConstFalse` in Monte Carlo and Mandelbrot;
the broad f64 compare/store variants on that lane have already been rejected.

Bounded diagnostic call traces then covered i-before-e, Monte Carlo, Sudoku,
MatrixMultiply, Mandelbrot, and reverse-complement. They do find common
dispatch names, but not a common material runtime leaf:

| Shared trace dispatch | Why it is not a candidate |
| --- | --- |
| `array_get_tracked_fast` | The callers are String UTF-8 helpers, Sudoku's nested tracked i32 rows, and MatrixMultiply's f64 row access; their profiles and carrier work diverge. |
| `array_push_fast` / `array_push_handle_fast` | These are existing canonical fast paths; reverse-complement uses u8 transforms, MatrixMultiply uses f64 rows, and i-before-e spends its remaining time in String helpers. |
| `array_len_fast` | It is bookkeeping around distinct programs, not a material shared CPU leaf. |
| inline/exact-native call dispatch | Monte Carlo has no material inner-loop calls, while the other programs reach different callees. |

Trace hit counts are not timing data. In particular, i-before-e's trace added
substantial map-accounting overhead, so it was used solely to rank callsites.
The untraced profiles already show the divergent downstream work. No compiler,
runtime, tree-walker, or `able-stdlib` source changed.

Next, refresh one unprofiled cap-compliant benchmark family at a time. Profile
a candidate only after both scorecard results and untraced CPU samples show the
same material generic leaf in two independent programs; retain the current
external rows as regression controls.

## Pidigits value-representation reconnaissance — no keep

Pidigits was refreshed because it is a remaining compiled-scorecard gap:

| Benchmark | Compiled Able / GMP-backed Go | Bytecode Able / Ruby |
| --- | ---: | ---: |
| `pidigits` | 1.3467s / 0.7400s (1.82x) | 2.5600s / 9.1800s (0.28x) |

The bytecode target is met, but compiled Pidigits remains behind Go. The
proposed primitive-integer companion, `sum_u32_small`, was rejected from the
candidate set after source and prior-profile review: Pidigits' hot work is
explicit `BigIntRef` host division/multiplication over `math/big`, while
`sum_u32_small` is primitive u32-to-u64 casting and mono-array access. They do
not share a boxing, conversion, slot, return, or host-ABI leaf.

No `BigIntRef` compiler special case is permitted, and one Pidigits benchmark
does not justify a nominal stdlib optimization. A future improvement may only
be a reusable `able-stdlib` host-kernel/API redesign, validated by a second
independently written BigIntRef benchmark that calls the same arithmetic API.
No runtime, compiler, tree-walker, or stdlib source changed.

Next, add that independent benchmark fixture before trying any BigInt
multiplication/division change. This will establish whether an optimization is
useful to ordinary programs using the library rather than to Pidigits alone.

## Independent BigIntRef confirmation — no keep

`bigint_ref_newton_small` is now a second, independently written BigIntRef
workload. It builds a 1024-digit integer, runs Newton integer-square-root
iterations, and uses the mutable `mul_i64`, `add`, `div`, `div_i64`, `mul`, and
`compare` API rather than Pidigits' digit-extraction algorithm. Its bounded
bytecode run is `3,315,269 ns/op`, `107,295 B/op`, and `259 allocs/op` over
500 iterations. Tree-walker and bytecode produce the identical result, and the
compiled one-run check exits successfully in `0.1600s` wall time.

Fresh CPU profiles reject a shared material optimization. The Newton workload
is dispatch-led (`execCallOpcode` 72.1% cumulative, `execCallMember` 38.8%)
and has only `math/big.divWW` at 3.0% flat. At the Pidigits workload size,
`math/big.mulAddVWW` alone is 28.9% flat, with large multiplication and
division paths dominating. Thus the common BigIntRef API surface does not yet
identify a common expensive leaf. No compiler, VM, tree-walker, or
`able-stdlib` source changed; the new fixture is retained as an independent
regression guard and is included in `fixture-numeric`.

Next, do not tune BigIntRef yet. Select a new cap-compliant pair only when
their untraced profiles expose the same material generic runtime or library
leaf; this prevents turning a Pidigits-shaped math kernel into a nominal-type
optimization without broad evidence.

## Nominal numeric triage — no keep

The bounded `Int128`, `UInt128`, and `Rational` fixture sweep was run to test
for a shared primitive-value or general nominal-runtime cost. These are all
nominal stdlib structs, not primitive Able types, so their names cannot justify
a compiler or VM fast path. The three-run bytecode results were:

| Fixture | Bytecode process | Bytecode steady state | Compiled process (one run) |
| --- | ---: | ---: | ---: |
| `int128_accumulate_small` | 3.5700s | 3.2048s, 329.6MB, 11.87M allocs | 7.6400s |
| `uint128_accumulate_small` | 2.1067s | 1.7180s, 230.4MB, 7.80M allocs | 5.6700s |
| `rational_series_small` | 0.4200s | 0.2096s, 6.8MB, 56.5k allocs | 0.3100s |

The profiles diverge. Int128 spends 36.0% cumulatively in the existing
`execInt128BinaryMemberFast` path and 12.2% in struct reconstruction;
UInt128 spends 76.6% cumulatively in member calls, 29.4% in UInt128 instance
construction, and 22.8% in boxed-integer conversion. Rational instead is
ordinary VM frame/call/cast work (`execCallOpcode` 25.3% cumulative,
`execCallName` 13.8%, and `finishInlineReturn` 12.4%). The shared
raw-integer-metadata samples are small and do not outweigh those distinct
costs. The lowering audit likewise finds only 23 ordinary `LoadSlot`
instructions across the three programs and no repeated specialized shape.

There is no checked-in Go/Python/Ruby equivalent for these local fixtures, so
these rows are diagnostic controls rather than a claim against either
competitiveness target. No compiler, VM, tree-walker, or `able-stdlib` code
changed. A broad three-mode validation attempt was stopped after crossing the
one-minute aggregate guard in the existing Int128 tree-walker path; each
compiled and bytecode measurement above completed under its individual
55-second cap.

Next, add cross-language equivalents for a fixed-width, two-word numeric
workload before pursuing this family. That will turn the local diagnostics into
a real compiler/VM scorecard, while a future optimization must still be a
generic nominal-lowering/runtime improvement demonstrated by at least two
unrelated nominal types.

## Fixed-width two-word cross-language control — no keep

`fixed-width-128` is now an external benchmark suite with canonical Able,
Go 1.26, Python 3.14, and Ruby 4.0 implementations. It has two independent
shapes in one executable: 1,000,000 carry-aware modular additions/reductions
and 1,000,000 two-word constructions/lexicographic selections. All four languages,
plus Able tree-walker, bytecode, and compiled execution, produce:

```text
9:8872316649623652056
15:48270715034
```

The first local development scorecard is deliberately explicit about its
scope: Go is `0.001368s/run` over 25 runs, Python `0.080005s/run` over five,
and Ruby `0.164176s/run` over five. Able's capped three-run process averages
are compiled `1.8767s` and bytecode `1.9300s`; bytecode steady state is
`1,720,939,385 ns/op`, `337.9MB/op`, and `6.27M allocs/op`. The one-run
tree-walker control completes in `10.5100s`. This is a real gap (compiled is
about 1,372x the Go row and bytecode about 24.1x / 11.8x the Python/Ruby rows),
but it is one nominal type and therefore cannot justify extending an existing
UInt128-specific path.

The `fixed-width` external and lowering-audit suite is registered without
adding it to `full` before a controlled Docker refresh publishes the reference
rows. Dockerfiles and a verifier are present in `../benchmarks/fixed-width-128`;
the shared `results.json` was intentionally untouched because it already has
unrelated local changes and the required base images are not cached here. No
compiler, VM, tree-walker, or `able-stdlib` source changed.

Next, add an external rational/value-normalization workload with the same
Go/Python/Ruby/Able matrix. Only if that unrelated nominal type exposes the
same material generic lowering, frame, or allocation leaf should a runtime or
compiler candidate be considered; otherwise these rows remain controls.

## Rational cross-language control and paired profile — no keep

`rational-series` now supplies the unrelated nominal control. Its two 50,000
iteration shapes are bounded GCD normalization and normalized add/multiply/
divide series arithmetic. Go 1.26, Python 3.14, Ruby 4.0, and all three Able
modes agree on:

```text
1810114:1663937
176266036:78300349
```

Local reference measurements are Go `0.010258s/run` over 25 runs, Python
`0.104552s/run` over five, and Ruby `0.136748s/run` over five. Able's capped
three-run rows are compiled `2.3933s` and bytecode `3.8167s`; bytecode steady
state is `3,625,031,522 ns/op`, `104.4MB/op`, and `1.01M allocs/op`. The
one-run tree-walker control completes in `12.5600s`. This is about 233x Go for
compiled execution and 36.5x / 27.9x Python/Ruby for bytecode, so it is a
useful regression control but not evidence that either target is met.

The paired profile comparison rejects a generic candidate. Fixed-width spends
26.1% cumulatively in its existing UInt128 checked-member fast path, while
Rational is generic call/frame/cast work (`execCallOpcode` 31.3% cumulative,
`execCallName` 14.4%, `finishInlineReturn` 10.0%). `runResumable` is a broad
umbrella rather than a leaf, and previous call-frame/call-name variants have
already failed broad controls. The `nominal-numeric` lowering audit passes with
37 ordinary `LoadSlot` instructions across both programs and no repeated
specialized shape. No compiler, VM, tree-walker, or `able-stdlib` source
changed.

The 200,000-iteration fixed-width local sample and the earlier local Rational
reference rows above remain profile controls, not the calibrated scorecard.

## Isolated nominal-numeric Docker publication — no keep

A detached benchmark-repository worktree at
`../benchmarks-nominal-publish-20260710` recorded a clean one-process Docker
snapshot for both suites. All twelve build/run/verifier combinations passed.
The shared runner now emits a generic nanosecond elapsed-time metric and its
parser prefers that metric over GNU `time`'s centisecond display, preventing
the Go rows from rounding to `0.00s`. This applies to every external benchmark;
it does not alter an Able runtime, compiler, or stdlib path.

| Suite | Able compiled | Able bytecode | Go | Python | Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| `fixed-width-128` | 5.428066437s | 6.107921918s | 0.004639621s | 0.458474346s | 0.661396768s |
| `rational-series` | 1.489584237s | 3.704359985s | 0.011639764s | 0.117506846s | 0.138126643s |

The normal comparison harness separately completed both catalogue mappings
against that snapshot. Its fresh Able runs put fixed-width at 1,170x compiled
Go and 9.00x bytecode Ruby, and rational-series at 137x compiled Go and 26.43x
bytecode Ruby. Neither workload is close to the 95% goals.

The snapshot deliberately does not overwrite `../benchmarks/results.json`,
which already contains unrelated local work. Keep both suites outside `full`
until that file can be reconciled cleanly; their dedicated `fixed-width` and
`nominal-numeric` suites remain usable for explicit audits and comparisons.

## Broad VM refresh: text, iterator, and numeric-array controls — no keep

The next bounded one-process refresh used the established OOM guardrails
(`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`). It intentionally paired
three unrelated fixture families rather than retrying a prior call/frame or
typed-pattern micro-variant:

| Workload | Profiled sample | Dominant concrete work |
| --- | ---: | --- |
| `string_split_join_small` | 986,574,216 ns/op; 49,653,456 B/op; 513,966 allocs/op (1x) | call-name/return and typed-pattern dispatch |
| `linked_list_iterator_collect_i64_small` | 230,802,093 ns/op; 3,254,294 B/op; 29,076 allocs/op (5x) | member/iterator dispatch and exact-native calls |
| `array_map_i32_small` | 55,298,661 ns/op; 850,837 B/op; 310 allocs/op (20x) | integer binary/cast work and primitive-array reads |

The profiles rule out a shared candidate. Split/join spends 27.3% cumulatively
in `execCallOpcode(...)`, 21.2% in `finishInlineReturn(...)`, and 13.1% in
typed-pattern dispatch. Collection spends 58.3% cumulatively in call execution
and 33.0% in member calls, with `lookupCachedMemberMethodEntry(...)` only 4.4%
flat. Array-map instead spends 17.4% cumulatively in `execBinary(...)`, 11.9%
in casts, and 9.2% in array-member reads. `runResumable(...)` is only the common
dispatcher parent; raw-integer metadata is 2.0%, 4.4%, and 7.3% respectively,
not a sufficiently shared material leaf.

Accordingly no VM, compiler, tree-walker, or `able-stdlib` code changed. The
profiles are retained under `.profiles/20260710-refresh-*.cpu.prof`. Do not
reopen call-name/frame, typed-pattern, raw-integer, or primitive-array
micro-variants from this evidence. The next investigation should profile a
small compiled external-scorecard pair with a non-text primitive control, then
consider only a shared generated-Go lowering/runtime boundary; that directly
serves the compiler-versus-Go goal while retaining this bytecode trio as guards.

## Compiled text/bytes with numeric control refresh — no keep

Fresh one-process compiled runs completed at `0.0800s` for `i_before_e`,
`2.2200s` for Base64, and `0.1700s` for Monte Carlo. Against the current
external Go rows (`0.0500s`, `2.2000s`, and `0.1800s`), Base64 and Monte Carlo
remain at or near the compiler target; i-before-e is a real but very short
1.6x gap. It is too short for a reliable CPU profile, as is Monte Carlo at this
input, so neither was artificially repeated or scaled to manufacture samples.

The profiled long-running Base64 binary instead has 89.9% cumulative time in
direct host kernels: base64 encode 40.6%, decode 34.1%, and MD5 15.3%. The
remaining visible GC work is 10.0%. This is already the intended primitive
extern boundary, not compiler bridge or generic generated-runtime overhead.
It cannot validate a change for i-before-e, and Monte Carlo exercises direct
numeric helpers rather than these byte kernels. No compiler, VM, tree-walker,
or `able-stdlib` source changed. The profile is retained as
`.profiles/20260710-compiled-base64.cpu.prof`.

The next compiler investigation should select two independently shaped,
profile-long external programs that are both materially outside the Go target
and inspect only shared generated-Go/runtime work. This avoids perturbing
already-direct host kernels or extrapolating from sub-second controls.

## Compiled recursive-pair scorecard follow-up — no keep

Fresh core runs measured Fib at `3.2100s` versus Go `2.8400s` and BinaryTrees
at `4.0900s` versus Go `3.8300s`. Generated-binary profiles reject a shared
compiler candidate: Fib is 100% in its already-direct generated Fibonacci loop,
whereas BinaryTrees is dominated by tree allocation/GC (`make_tree` 64.0%
cumulative; allocation 53.8%) and the goroutine executor. No bridge, lowering
helper, or runtime boundary repeats. Both binaries completed and were verified
independently; no compiler, VM, tree-walker, or stdlib code changed.

## Compiled-scorecard pair selection — no keep

No existing pair meets the evidence bar. Reverse-complement and i-before-e are
too short; Base64, MatrixMultiply, Monte Carlo, and Fib are already direct
primitive kernels; Pidigits is explicit BigIntRef library work; K-Nucleotide
lacks a Go reference; and Tapelang is program-defined Tape behavior. Quicksort
is profile-long but its recursive sort/decimal parsing cost has no matching
material leaf. No source change is justified. The next scorecard investment is
to add missing Go references for profile-long application workloads, then use
the completed comparison to select a genuinely shared primitive boundary.

## Missing Go-reference validation — no Able change

The existing Go 1.26 K-Nucleotide and N-body implementations, Docker entries,
and verifiers now pass both local and Docker validation. Their Docker runs took
`0.053801632s` and `0.030855248s` respectively. The generic high-resolution
timer was corrected to write its metric to stderr, so benchmark verification
continues to consume only program stdout. The dirty shared results file was not
overwritten; publish reconciled rows separately before using these references as
scorecard gates. No Able runtime, compiler, tree-walker, or stdlib code changed.

## Isolated K-Nucleotide/N-body comparison — no keep

Fresh compiled Able is `2.8900s` versus Go 1.26 `0.053543927s` for
K-Nucleotide and `0.4400s` versus `0.054752445s` for N-body. The generated
profiles diverge: K-Nucleotide is HashMap representation/conversion and GC
work, while N-body is direct sqrt/abs numeric work plus bridge environment
swaps. No shared primitive boundary exists, so no HashMap, FASTA, N-body, or
other benchmark-shaped optimization was attempted.

## Map-companion discovery — no keep

`word_count_small` is the only current application-shaped map control, but its
fresh compiled run is `0.1100s`, too short for a reliable generated-binary
profile. Its prior map lookup attribution is also not material beside string
and struct work. It cannot validate K-Nucleotide's HashMap conversion/GC cost;
the fixture was not artificially scaled. A future companion must be an
independently written external word-frequency application with a realistic
checked-in corpus and Go/Python/Ruby/Able references.

## Checked-in word-list companion trial — rejected

An independent full-word-count trial used the checked-in 1.7 MiB word list.
Go, Python, and Ruby agree on `172823:4508531`, but the canonical Able program
does not reach a bounded result while building the mostly-distinct 172,823-word
map. The incomplete sources were removed. This is a useful stress observation,
not a valid scorecard benchmark; a future companion needs repeated natural
vocabulary and a bounded Able baseline.

## Repeated-vocabulary corpus qualification — no keep

The checked-in v12 language specification (316 KiB) is a suitable realistic
prose corpus: a temporary independent full-word-count probe completes in
`0.5200s` compiled and `3.0400s` bytecode. Unlike the word list, it has
repeated natural vocabulary and stays inside the guardrail. The probe was
removed; a later self-contained external suite must copy this corpus and add
all reference implementations before it becomes a scorecard member.
