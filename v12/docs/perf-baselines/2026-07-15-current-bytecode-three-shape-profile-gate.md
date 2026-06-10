# Current bytecode three-shape profile gate — 2026-07-15

## Decision

Keep no bytecode VM, compiler, canonical-stdlib, or benchmark-source change.
The current warmed profiles repeat VM dispatch parents and two helpers whose
generic candidates have already failed broad guards, but do not expose a new
material concrete descendant across text, iterator, and numeric work.

## Method

Each target was loaded once, warmed once, and then measured in one
`BenchmarkBytecodeProgramRuntime` process with the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.  A three-sample quiet-host
preflight immediately preceded each profile. CPU 1 was quiet for the text and
iterator samples; CPU 11 was quiet for the numeric sample after CPU 1 became
busy. The rows are therefore profile evidence, not a cross-row timing
comparison.

`--source-root-only` and the fixture directory as `--run-from` were required
so the repository root could not be mistaken for a second stdlib source root.
The initial default-root launch was rejected by that collision before its
benchmark started; it produced no profile and is not evidence.

| Workload | Calls | Result | CPU profile |
| --- | ---: | --- | --- |
| `string_split_join_small` | 5 | 924,484,804 ns/op; 49,445,435 B/op; 555,701 allocs/op | `20260715_current_string_split_join.cpu.pprof` |
| `linked_list_iterator_collect_i64_small` | 20 | 394,920,717 ns/op; 8,416,665 B/op; 192,810 allocs/op | `20260715_current_iterator_collect.cpu.pprof` |
| `array_map_i32_small` | 75 | 62,702,459 ns/op; 805,763 B/op; 95 allocs/op | `20260715_current_array_map_i32.cpu.pprof` |

The CPU profiles and matching runner JSON records live in
`v12/interpreters/go/.profiles/` and are generated, cleanup-eligible evidence.
Direct bytecode output checks completed with `191484`, `382455000`, and
`1097192358`, respectively.

## Attribution

| Shared-looking frame | Current evidence | Decision |
| --- | --- | --- |
| `execCallOpcode` | 27.17% cumulative in text, 64.21% in iterator collect, and 39.32% in numeric mapping | This is a dispatcher parent. Its descendants split into named-call return/coercion, iterator/generator member dispatch, and Array-slot calls. |
| `finishInlineReturn` | 19.57%, 7.87%, and 10.04% cumulative | It is present in all three but reaches different semantic returns. The broad slotless-return-guard experiment was neutral to mixed and was reverted; this refresh provides no reason to retry it. |
| `bytecodeRawIntegerValueInfo` | 8.91%, 3.17%, and 4.27% flat | It is a small shared helper with different callers and value shapes. The raw-integer extraction candidate already regressed split/join and iterator guards despite a numeric-map gain, so it remains closed. |

The text route also has named-call/type-match and return coercion work; the
iterator route is dominated by `next`/`yield` generator and static-member
dispatch; the numeric route is dominated by `execCallMemberArraySlot`, Array
get/push, and primitive transport. String-map lookup and type-match leaves are
not material in all three. A named Array, String, iterator, map, or benchmark
optimization would therefore be an unsupported specialization.

## Next recommendation

Use the current verifier-backed scorecard to select a genuinely new pair or
trio of material bytecode misses, and profile it only when it crosses an
unrepresented language boundary. This three-shape refresh reconfirms that the
remaining local VM micro-candidates are either parent frames or previously
rejected generic changes; new evidence is needed before modifying the VM or
canonical stdlib.
