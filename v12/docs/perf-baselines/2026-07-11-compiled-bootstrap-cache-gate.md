# Compiled bootstrap small-integer cache gate (2026-07-11)

## Purpose

The fair pinned process refresh left I-Before-E and ReverseComplement as
short-lived compiled application misses. ReverseComplement's generated `main`
was too short to sample, so this tranche measured initialization before trying
to optimize any application loop. The guard set was the real external
I-Before-E, JSON, and ReverseComplement applications with their suite
verifiers, plus the representative warmed bytecode string, iterator, and
numeric programs.

## Shared cold boundary

`GODEBUG=inittrace=1`, CPU `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1` show that all three generated binaries eagerly initialized the
interpreter package before generated `main`:

| Application | `pkg/interpreter` init clock | Bytes | Allocations |
| --- | ---: | ---: | ---: |
| I-Before-E | 59 ms | 37,994,624 | 707,143 |
| JSON | 63 ms | 37,994,544 | 707,143 |
| ReverseComplement | 60 ms | 37,993,872 | 707,133 |

The only material package initializer is the bytecode small-integer box cache.
It prebuilds values for the fixed `-256..16384` range and the extended `i32`
range, even when a static compiled program never executes bytecode.

Thirty verifier-checked ReverseComplement launcher profiles give the other
part of the boundary. The 50 ms merged bootstrap sample places 30 ms cumulative
(60%) in `RegisterIn`, with 20 ms (40%) in compiled package registration and
its kernel definition decode. Existing I-Before-E bootstrap samples agree:
`RegisterIn` is 50 ms cumulative (33.3%) and package registration 30 ms
(20%). The package initializer precedes both profiles and is therefore not in
their sample window.

## Candidate and compiled result

The candidate deferred the table through a concurrency-safe one-time
initializer. On the three verified generated binaries it reduced the
interpreter-package init cost to 18--19 ms, 9.51 MB, and 263,215 allocations.
The same candidate's normal verified process rows were encouraging:

| Application | Baseline | Candidate | Method |
| --- | ---: | ---: | --- |
| I-Before-E | 0.0930 s | 0.0450 s | 10 pinned verifier-checked launches per side |
| ReverseComplement | 0.0930 s | 0.0460 s | 10 pinned verifier-checked launches per side |
| JSON | 0.7760 s | 0.6880 s | 5 interleaved pinned verifier-checked pairs |

The first deferral placed the one-time check on hot integer-boxing helpers and
was rejected immediately by the warmed VM guard. A refined form restored eager
setup for normal tree-walker/bytecode constructors and used a cache-free
constructor only for generated compiled launchers. It retained the application
startup benefit, but the safe cache-absence checks still changed the hot VM
shape.

## Bytecode gate and decision

For an exact same-tree comparison, the refined candidate was removed only long
enough to build a baseline interpreter test binary, then restored to build the
candidate binary. Each binary ran the same loaded program under CPU `2`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

| Warmed bytecode application | Baseline | Candidate | Result |
| --- | ---: | ---: | --- |
| `string_split_join_small` | 986,189,531 ns/op | 1,029,091,656 ns/op | +4.35% |
| `linked_list_iterator_collect_i64_small` | 242,880,369 ns/op | 245,630,156 ns/op | +1.13% |
| `array_map_i32_small` | 71,055,513 ns/op | 73,789,592 ns/op | +3.85% |

Three more split/join repetitions confirmed the material regression:
1,049,860,064 ns/op baseline versus 1,094,916,707 ns/op candidate (+4.29%).
Array-map was noisy but not a compensating win (73,379,977 ns/op baseline
versus 74,252,741 ns/op candidate across its three repeats).

The candidate is therefore reverted. It would make a few short compiled
programs faster by slowing a generic bytecode value path, which violates the
cross-program performance rule. No compiler, interpreter, or `able-stdlib`
performance code is retained.

## Next recommendation

Profile and count allocation sources inside `RegisterIn` across the same three
generated binaries, then compare their generated registration tables. The
common cache initialization is not safely separable within the current shared
interpreter package; launcher registration is the remaining sampled shared
boundary. A follow-up may change only an application-independent registration
representation or default builtin registration step that repeats in both
misses and remains neutral on JSON and warmed bytecode. Do not retry lazy
small-integer cache checks without a design that physically separates bytecode
cache state from compiled-launcher runtime state and proves the VM hot path is
unchanged.
