# Bytecode Stack Diagnostics and Loop Result Discard

## Result

This tranche adds opt-in bytecode diagnostics and retains one generic lowering
repair. `ABLE_BYTECODE_STATS=1` now records maximum value-stack depth, maximum
value-stack capacity, capacity-growth count, and maximum call-frame depth.
`ABLE_BYTECODE_STATS_OUT=/path/stats.json` writes a snapshot at normal CLI exit
and, with Go profiling active, on the bounded profiling interrupt path.

Tapelang reached 2,071,446 stack values in five seconds with only four call
frames. The independent recursive named-struct BinaryTrees run reached only
24 stack values, capacity 32, and 23 call frames. Both used CPUs `2-3`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and real external inputs. The first 120-second
diagnostic Tapelang attempt was OOM-killed around 37 seconds before it could
flush; controlled 25-second probes retained the same guardrails.

The growing shape was a final conditional expression in a loop body. The old
lowering selected one trailing discard after all branches, allowing other
branches to retain one result per iteration. The repair lowers effect-only
final expressions with statement semantics for `loop`, `while`, and `for`, so
each branch discards its own result. It is a language-level stack-balance rule,
not a program, benchmark, or nominal-type rule.

The same five-second Tapelang probe then peaked at 471,733 values (4.4x lower).
A 30-second uninstrumented candidate retained 104 MB; the earlier run had
already reached roughly 483 MB near that point. A separate residual
store-result path remains, so this is not a complete Tapelang solution.

An explicit 11-program bytecode guard completed (array, channel, deque,
Dijkstra/heap, future, NBody, persistent sorted set, random, regex, and text)
at one run and a 45-second cap. Results are in
`v12/tmp/loop-discard-candidate-guard.{json,md}`. The broader preset reached
twelve members before its pre-existing `persistent_set_i32_small` unsupported
integer-kind runtime error, so it is not counted as all-green.

Focused diagnostics/lowering tests passed. A later full combined Go-package
command was environment-terminated before reporting and is deliberately not
claimed as a pass. No compiler or `able-stdlib` source changed.

## Next

## Per-op Delta Follow-up

The follow-up added signed per-op net value-stack deltas to the same opt-in
snapshot. It attributes a change at each instruction boundary to the preceding
opcode, including an instruction immediately before return or error.

Five-second pinned probes produced these totals:

| Workload | Net stack delta | Peak depth | Interpretation |
| --- | ---: | ---: | --- |
| Tapelang | +446,064 | 446,057 | Residual growth remains after the loop-body repair. |
| BinaryTrees | +11 | 24 | Normal recursive operand traffic returns to a bounded stack. |

Tapelang's largest positive opcode totals were `LoadSlot` (+13,235,643),
`LoadSlotStructField` (+9,863,690), and `LoadName` (+6,149,293). Its largest
consumers were `CallMemberArraySlot` (-7,584,671), `JumpIfFalse`
(-7,237,798), `Binary` (-7,237,792), `CallMember` (-3,370,841), and `Pop`
(-3,272,982). This is ordinary operand traffic, not a proof that any one
opcode is unbalanced; notably, a store can preserve an already-produced top
value while reporting a net delta of zero.

Focused VM, CLI, and profile-hook tests covering the new accounting passed.
No new runtime, compiler, or `able-stdlib` behavior changed in this follow-up.

## Next

## Instruction-site Follow-up

The next opt-in layer attributes only stack high-watermark extensions to a
capped set of instruction sites. Entries report bytecode opcode/IP plus source
origin, line, column, and the number of values by which that site extended a
VM's maximum depth.

In the five-second pinned Tapelang probe, 309,745 of 309,751 high-watermark
values were credited to `LoadSlotStructField pos` at
`tapelang_alphabet.able:50:58`, the `self.pos` field load in the nested
`self.cells.read_slot(self.pos)` expression inside `Tape.inc`. The same
BinaryTrees probe remained bounded; its largest site grew only 21 values.

This does **not** authorize a `Tape`, field-name, or struct-load special case:
a high-watermark site is the first push after a pre-existing residual, not
necessarily the later instruction that fails to consume the value. The two
workloads do not share the same site or call shape. Focused VM, CLI, and
profile-hook tests passed. No runtime, compiler, or `able-stdlib` behavior
changed beyond diagnostics.

## Next

## Inline-frame Balance Follow-up

The next opt-in layer records the stack base after inline-call arguments are
removed, then aggregates any excess stack entries immediately before the
callee's return value is appended. It covers full and self-fast inline frames
and reports the callee return source location only when excess is nonzero.

Both five-second pinned probes produced an empty excess report:

| Workload | Peak depth | Inline-frame excess sites |
| --- | ---: | ---: |
| Tapelang | 312,882 | 0 |
| BinaryTrees | 23 | 0 |

Therefore the Tapelang residual does not cross an inline return boundary. No
return-frame reorder or special return path is authorized. Focused VM, CLI,
and profile-hook tests passed. No runtime, compiler, or `able-stdlib` behavior
changed beyond diagnostics.

## Next

Add opt-in signed **instruction-site** net stack deltas, capped to the top
positive sites. Unlike a high-watermark report, this will show whether a
specific direct call or control-flow instruction fails to consume values at a
site. Collect the same bounded Tapelang and independent-program probes; make a
repair only if the resulting sequence is language-level, shared, and passes
the broad guard plus an external scorecard.

## Signed Instruction-site Delta Follow-up

The final diagnostic layer now aggregates each instruction site's signed net
value-stack contribution, capped to 512 sites. It also supports
`ABLE_BYTECODE_STATS_MAIN_ONLY=1`: the CLI resets opt-in counters immediately
before invoking the entrypoint, preventing bootstrap and module setup from
using the capped site budget. This changes diagnostics only; it does not alter
normal execution or bytecode behavior.

Fresh five-second external probes used CPUs `2-3`, `GOMEMLIMIT=1GiB`,
`GOGC=50`, the canonical external stdlib, and `timeout --signal=INT
--kill-after=30 5`. The profiling-interrupt hook wrote these snapshots:

| Workload | Net instruction delta | Peak depth | Capacity / growths | Call frames |
| --- | ---: | ---: | ---: | ---: |
| Tapelang | +169,817 | 169,816 | 184,832 / 26 | 4 |
| BinaryTrees | +8 | 23 | 32 / 2 | 23 |

Tapelang's largest positive sites were ordinary operand producers: `kinds`,
`ip`, `kind`, and `OP_INC` loads in `execute` line 128–129 (+792,620 each),
followed by the field and argument loads within `Tape.inc` line 50
(+377,394 each). Its `CallMemberArraySlot`, `CallMember`, `Binary`,
`JumpIfFalse`, and `Pop` opcode totals were all net consumers. Thus the
previous high-watermark `self.pos` load remains an observation point after a
residual rather than evidence that the field load is defective.

BinaryTrees exhibits the control pattern expected of a balanced workload:
anonymous struct-construction constants contribute +7,050,220 values and its
named-struct fast instruction consumes -7,050,184, leaving only the bounded
entry-state residue. Its report shares neither Tapelang's retained growth nor
a positive direct-call/control site.

The site report therefore does not authorize a field-load, direct-call,
array, `Tape`, `BinaryTrees`, or named-container special case. Focused VM,
CLI, and profile-hook tests passed. No compiler or `able-stdlib` source
changed.

## Call-operand-region Balance Follow-up

The next opt-in layer records the operand base for every bytecode call,
expects one result value, and reports only a post-call excess. It covers
regular, named, member, cached, static-member, Array-slot, and self calls.
Inline calls are deferred until their shared return path appends the callee
result, so the report checks the same boundary independent of dispatch form.
The 256-site report is disabled unless bytecode statistics are enabled; normal
call dispatch keeps its prior direct path.

Fresh probes used the established canonical stdlib, CPUs `2-3`,
`GOMEMLIMIT=1GiB`, and `GOGC=50`. Tapelang and BinaryTrees ran for five
seconds with profiling-interrupt snapshots; the finite array-fold and
linked-list iterator collect fixtures wrote normal-exit snapshots from an
isolated fixture symlink so the repository stdlib collision guard was not
part of their measurements.

| Workload | Net stack delta | Peak depth | Call-operand excess sites |
| --- | ---: | ---: | ---: |
| Tapelang | +167,456 | 167,455 | 0 |
| BinaryTrees | +13 | 23 | 0 |
| Array fold | 0 | 3 | 0 |
| Linked-list iterator collect | 0 | 4 | 0 |

Tapelang recorded 819,203 inline calls, including 818,996 resolved member
calls; direct member and named generic-fallback counts were zero. Its empty
operand and inline-frame reports therefore rule out receiver/argument
truncation, result append, and inline return as the retained-stack source.
The array and iterator controls also exercised canonical Array and generic
nominal collection calls without a call-boundary excess.

This authorizes no runtime fast path or repair: all measured call boundaries
are balanced, including the workload that still grows. No compiler or
`able-stdlib` source changed. Focused VM, CLI, and profile-hook tests passed.

## Backward-loop-edge Balance Follow-up

The final stack diagnostic records the first target-depth baseline for each
unconditional backward jump, then aggregates later depth above that baseline.
The report is capped to 512 sites and runs only with bytecode statistics
enabled. Loop, `while`, `for`, and `continue` back jumps now retain their
source node for diagnostic attribution; this does not change their execution
semantics.

Before the repair, fresh five-second external probes found this result:

| Workload | Net stack delta | Peak depth | Loop-edge excess sites |
| --- | ---: | ---: | ---: |
| Tapelang | +157,522 | 157,513 | 2 |
| BinaryTrees | +2 | 23 | 0 |
| Array fold | 0 | 3 | 0 |
| Linked-list iterator collect | 0 | 4 | 0 |

The dominant Tapelang site was the `execute` loop at
`tapelang_alphabet.able:126:3`: 735,211 backedges began at depth 2, reached a
maximum +157,507 excess, and accumulated 57,886,243,957 excess values. A
small `Tape.ensure` report was downstream of that pre-existing caller residue.
This proves that the retained value crosses the outer loop iteration, while
the independent array, iterator, and recursive controls remain balanced.

The offending shape is generic nested control flow. In an effect-only outer
`if` branch whose final statement is an inner `if`, the old
`bytecodeDiscardTrailingBlockResult(...)` shortcut inspected only the last
emitted instruction. If that instruction was the inner `else` branch's store,
it marked that one store as discarded and omitted the outer `Pop`; the inner
then branch could still leave its result on every iteration.

The kept repair requires the candidate trailing store's source node to be the
outer source block's actual final statement. Direct final assignments retain
the existing discard fast path. Nested `if`, match, loop, and other control
expressions instead use the existing whole-expression `Pop`, which discards
exactly one result on every falling-through branch. This is a language-level
effect-only block rule, not a Tapelang, field, Array, or nominal-container
special case.

The nested-conditional regression test stays bounded and preserves
tree-walker parity. The repaired five-second Tapelang probe reached peak 7 / a
net +8 delta with no loop-edge report. Repaired BinaryTrees, array fold, and
iterator collect remained balanced. The 11-program bytecode guard completed
11/11 (array, channel, deque, Dijkstra, future, heap, NBody, persistent sorted
set, random, regex, and text) at a 45-second cap; artifacts are in
`v12/tmp/loop-edge-candidate-guard.{json,md}`.

An uninstrumented external Tapelang run under `GOMEMLIMIT=1GiB` and `GOGC=50`
used the full 45-second wall-clock cap rather than completing, but it no longer
has the diagnosed retained-stack growth. No compiler or `able-stdlib` source
changed. Focused VM, CLI, and profile-hook tests passed.

## Post-stack CPU and Allocation Refresh

This refresh used the canonical stdlib at
`/home/david/sync/projects/able-stdlib/src`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.
The repaired external Tapelang process was sampled for ten seconds on CPUs
`2-3`; its timeout was intentional. The warmed finite controls ran under
`GOMAXPROCS=1` on CPU `2`: linked-list iterator collect for five iterations,
and numeric Array map for twenty. Retained profiles are
`20260711_external_tapelang_poststack_10s.{cpu,alloc,heap}.pprof` at the
repository root and
`20260711_{iterator_collect_poststack_5x,array_map_poststack_20x}.{cpu,mem}.pprof`
under `v12/interpreters/go/.profiles/`.

| Workload | Result | Material CPU evidence |
| --- | --- | --- |
| External Tapelang | bounded ten-second sample; does not finish within the cap | Array-slot member/read path (26.2% / 15.8% cumulative), slot loads (13.1%), then inline return (7.1%) |
| Iterator collect | 230,659,694 ns/op; 3,247,876 B/op; 29,063 allocs/op | generator/member/iterator dispatch, typed-pattern work, inline return (7.9%) |
| Numeric Array map | 62,509,804 ns/op; 849,247 B/op; 307 allocs/op | Array-get/member calls, binary/raw-integer work, inline return (8.7%) |

`execCallOpcode(...)` remains a shared dispatcher parent rather than a common
leaf. Tapelang is dominated by Array slot reads and cached member dispatch;
iterator collect by generators, iterator-next, and type matching; and numeric
map by Array reads, arithmetic, and raw integer transport. `finishInlineReturn`
recurs at 7--9%, but the descendants differ by all three workloads and the
slotless-return guard reorder was already rejected by the broad guard set. This
refresh supplies no reason to retry it.

The allocation captures show `runtime.ArrayEnsureCapacity(...)` at 10.33 MB
(14.8%) for iterator collect and 10.67 MB (21.2%) for numeric map. The external
all-process allocation profile is startup-dominated, so it neither confirms nor
contradicts a Tapelang allocation wall. One generic policy experiment widened
post-4096 Array growth from 1.5x to 1.75x. It was not source- or
container-specific: `grownCapacity(...)` is shared by dynamic Array storage and
its related caches. It increased iterator allocation from about 3.25 MB/op to
about 3.54 MB/op, left the text guard neutral, and had only noise-level numeric
movement (candidate 58.2--61.0 ms/op versus restored 59.1--61.9 ms/op at ten
iterations). The candidate was reverted.

Keep no runtime, compiler, or canonical-stdlib code from this refresh. The
repaired Tapelang stack stays bounded, but the three workloads do not expose a
new material concrete VM leaf that repeats across their independent feature
families. The CPU/heap profiles also show no retained Tapelang VM-heap growth;
the external allocation and heap captures are mostly initialization and parsing.

## Next

Refresh the verified cross-language scorecard in bounded family-sized shards:
compare compiler output with Go, and bytecode with Python and Ruby, across the
existing benchmarkable-program suite. The VM profiles have exhausted the
currently shared helper candidates, so an observed target miss is the stronger
way to select the next general change. Preserve the pinned toolchain and OOM
guardrails, record repeatable multi-run rows, and profile the selected miss plus
an unrelated guard before considering a generic interpreter, compiler, or
canonical-stdlib change.
