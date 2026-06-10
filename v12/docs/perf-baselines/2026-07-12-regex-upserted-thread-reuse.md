# Regex upserted-thread reuse — 2026-07-12

## Decision

Keep the canonical-stdlib change. A successful private
`regex_nfa_upsert_thread` now returns the exact immutable `RegexNFAThread`
that it stored in the active set. `regex_nfa_add_closure` pushes that returned
record onto its local work stack instead of constructing a second record with
the same state, start, and capture tags.

An upsert still returns no record when the earlier active thread wins. A
replacement still stores and returns only the new earlier-start record. The
stack is drained before the closure returns, so this change neither pools nor
shares state between ordinary matching, `RegexSet`, and scanner executions.
Public NFA snapshots remain unchanged.

## Bytecode runtime gate

The warmed one-call lane used canonical external `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Suffix audit and ordinary
matching used their established 128-input `--profile` modes; RegexSet used
its matching 128-word `--profile` mode. Lower is better.

| Workload | Previous closure-stack readings | Reuse readings | Allocation change |
| --- | ---: | ---: | ---: |
| suffix audit | 1,406,961,286; 1,390,048,956 ns/op | 1,165,084,432; 1,147,043,850 ns/op | 611,406/611,337 -> 490,472/490,449 allocs/op |
| independent `is_match` | 675,468,763; 683,784,911 ns/op | 574,214,026; 561,888,653 ns/op | 437,782/437,739 -> 345,280/345,212 allocs/op |
| `RegexSet` audit | 2,230,061,648; 2,193,324,551 ns/op | 1,795,066,480 ns/op | 723,980/723,892 -> 531,145 allocs/op |

This is a 16--19% elapsed-time improvement and a 20--27% allocation-count
reduction in all three independent public tagged-NFA API shapes. The first
RegexSet run accidentally used its 512-word default workload; it measured
7.63s and is intentionally excluded because it is not comparable with the
established 128-word profile lane.

## Correctness and generated-code gate

- All `exec/14_*` regex fixtures passed in tree-walker and bytecode modes.
- The independent generic `automata_dfa_small` bytecode control still printed
  `120000000`.
- Default compiled suffix and RegexSet applications passed their Ruby
  verifiers; the independent compiled application completed with its existing
  deterministic output. The generated RegexSet seed was `0.3000s` real,
  versus the previous `0.3200s` seed.
- Temporary generated source trees reached roughly 800 MB per application and
  were removed immediately after the checks; no generated binaries or source
  trees are retained by this decision.

## Next recommendation

Profile `zigzag_char_small`, `reverse_complement_small`, and
`ascii_lower_small` in bounded compiled allocation mode before changing
`__able_char_to_codepoint`. The helper has a large repeated contribution in
the three regex captures, but that is insufficient evidence for a language
primitive/compiler optimization. Independent character workloads establish
whether the same concrete helper and caller are broadly material; only then
should a general kernel-boundary candidate be designed and gated across the
suite.
