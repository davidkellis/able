# Bytecode main-dispatch attribution gate

Date: 2026-07-21

## Decision

Keep no code. Three merged restored-VM profiles and separate opcode censuses do
not expose one exact ordinary-dispatch guard, lookup, or stack transition that
is material in Future Pipeline, Concurrent Text Index, and Word Frequency.

The common `runResumable(...)` and `execCallOpcode(...)` frames are dispatcher
parents over different semantic work. The only exact source lines sampled in
all three are the irreducible loop condition, instruction fetch, opcode switch,
and once-per-run statistics flag. No individual line reaches 1% in Concurrent
Text. A candidate would therefore optimize profile attribution noise or reopen
an already rejected stack/raw-integer family.

No VM, runtime, compiler, stdlib, application, fixture, reference, scorecard,
language, or WASM change is retained.

## Protocol

A restored bytecode test binary was frozen before profiling. Each application
received three independent clean CPU processes. Future and Concurrent Text ran
ten warmed `main()` calls per process with the goroutine executor; Word
Frequency ran three calls with the serial executor. Profiles were merged per
application. Bytecode statistics were disabled in every CPU process.

Separate one-call untimed census processes recorded opcode counts. All
processes used the canonical external stdlib, `GOMAXPROCS=1`, `GOGC=50`, a
1-GiB memory limit, and a 60-second cap.

| Application | Merged CPU samples | Profile-process mean | Bytecode ops / call |
| --- | ---: | ---: | ---: |
| Future Pipeline | 8.73 s | 291,612,495 ns | 9,650,629 |
| Concurrent Text Index | 11.55 s | 386,324,126 ns | 1,976,250 |
| Word Frequency | 10.55 s | 1,178,122,057 ns | 17,591,889 |

The process means are profile-run diagnostics, not promoted scorecard timings.

## Opcode ownership

The top opcode mixes are materially different:

| Application | Leading opcode counts |
| --- | --- |
| Future Pipeline | `LoadSlot` 2,203,692; `BinaryIntAdd` 1,589,255; `Pop` 1,155,134; `Jump` 573,461; fused integer stores/branches about 0.52-0.54M each |
| Concurrent Text Index | `LoadSlot` 447,780; `Pop` 202,141; `CallMember` 113,162; `Return` 108,944; `MemberAccess` 101,069; `CallName` 91,865 |
| Word Frequency | `LoadSlot` 4,406,730; `Pop` 2,465,647; `StoreSlotNew` 1,859,411; `Jump` 1,025,404; `CallName` 796,567; typed-pattern jump 778,084; `Return` 673,761 |

`LoadSlot` and `Pop` occur everywhere, but their concrete stack-carrier
descendants belong to a repeatedly rejected family and do not supply a new
transport rule. Future is dominated by integer arithmetic/fused loops,
Concurrent Text by member/call dispatch, and Word by calls, typed patterns, and
map/string work.

## Exact `runResumable` lines

Merged line profiles put `runResumable(...)` at 13.97%, 3.12%, and 6.54% flat
in Future, Concurrent Text, and Word respectively. Its common exact lines are:

| Source operation | Future | Concurrent Text | Word Frequency |
| --- | ---: | ---: | ---: |
| once-per-run statistics flag | 2.18% | 0.43% | 0.57% |
| loop condition | 1.60% | 0.26% | 1.42% |
| instruction fetch by `vm.ip` | 2.52% | 0.35% | 1.23% |
| opcode switch | 1.26% | 0.61% | 0.76% |
| combined irreducible dispatch lines | 7.56% | 1.65% | 3.98% |

No single line is material in all three. Combining four different operations
would not establish one removable owner, and localizing `vm.ip` or replacing
the switch would require rewriting helper/jump semantics without evidence that
such an architecture repays the low-share Concurrent Text path.

## Call-parent check

The opcode census and line profiles also put the call-family branch at 13.86%,
64.24%, and 34.98% cumulative. Its own dispatcher is only 0.11%, 0.52%, and
1.04% flat. Descendants split as follows:

- Future: named, member, and static calls beneath arithmetic loops;
- Concurrent Text: member resolution/cache work, then named/static/Array calls;
- Word Frequency: named calls, Array gets, type matching, and return coercion.

This reproduces the existing `execCallOpcode` closure: it is a common parent,
not a common removable leaf. The nil guard and opcode switch inside that helper
are not sampled materially in all three. Call-name, member-cache, return,
type-match, stack-carrier, and raw-integer micro-variants already have broad
rejection records and are not reopened here.

## Correctness and cleanup

No candidate or diagnostic code was added. The restored full `TestBytecode`
family passes in 23.720s. Temporary binaries, profiles, census JSON, and local
generated artifacts are removed after this record is written.

## Next recommendation

Run a fresh cross-corpus exact-leaf selection sweep rather than continuing on
this exhausted three-application intersection.

Why: the restored main dispatcher has no new shared removable leaf across the
current cohort, while the bytecode product frontier still contains large gaps
in concurrency, text/map, numeric, regex, wide-integer, and byte-output
families. Selection now needs broader evidence to find a different generic
runtime owner, not another variant of closed call/return/stack machinery.

What it entails: collect bounded, clean current profiles for one high-excess
application from at least six unlike families—using Future Pipeline, Word
Frequency, Distance Field, Regex Set Audit, Fixed Width 128, and Reverse
Complement as the initial set. Build an exact-symbol/leaf intersection matrix,
exclude dispatcher parents and every documented closed family, and require one
concrete compiler/VM/runtime-controlled leaf in at least three unlike
applications before implementing anything. Guard any candidate with repeated
arithmetic-mean application timings plus established JSON, Pidigits, and
Base64 bytecode target rows. Continue to defer WASM.
