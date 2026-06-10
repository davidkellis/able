# Regex matcher-local closure-stack reuse — 2026-07-12

## Decision

Keep the canonical-stdlib implementation. Each ordinary NFA match and
`RegexSet` operation now owns one reusable `Array RegexNFAThread` work stack;
a `RegexScanner` owns one for its stream lifetime. `regex_nfa_add_closure`
borrows that stack and drains it before returning, while `regex_nfa_move`
passes the same empty stack to each closure it invokes.

The scratch array holds only temporary thread values. Active thread arrays,
next-state arrays, and capture-tag arrays retain their existing ownership and
allocation behavior. Consequently no state can cross a match, set operation,
or scanner instance, and public NFA/RegexProgram snapshots are untouched.
This is a general tagged-NFA execution improvement, not a capture-free,
`RegexSet`, corpus, pattern, VM, or generated-code special case.

## Verification

- All `exec/14_*` regex fixtures passed in tree-walker and bytecode modes with
  canonical external `able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
  `GOMAXPROCS=1`.
- The compiled no-fallback scanner-compaction fixture passed.
- `automata_dfa_small` still printed `120000000` in bytecode mode.
- Default generated suffix and RegexSet applications passed their Ruby
  verifiers; default generated `regex_is_match_small` retained `199960002`.

## Bytecode runtime gate

The warmed one-call benchmark lane retained the prior outgoing-index inputs.
Each value below is a fresh process; the two readings show the repeatability
of the directional result. Lower is better.

| Workload | Before | Reuse readings | Allocation change |
| --- | ---: | ---: | --- |
| suffix audit | 1,571,017,595 ns/op | 1,406,961,286; 1,390,048,956 ns/op | 701,170 -> 611,406; 611,337 allocs/op |
| independent `is_match` | 741,606,031 ns/op | 675,468,763; 683,784,911 ns/op | 517,692 -> 437,782; 437,739 allocs/op |
| `RegexSet` audit | 2,320,082,691 ns/op | 2,230,061,648; 2,193,324,551 ns/op | 924,798 -> 723,980; 723,892 allocs/op |

That is roughly 5--11% lower elapsed time and 13--22% fewer allocations in
every representative public API shape. Allocation bytes also fell in every
lane: suffix 80.38 MB to 79.19/79.39 MB, ordinary matching 47.27 MB to
45.20/45.39 MB, and RegexSet 114.04 MB to 104.28/103.95 MB.

The CPU profiles retained with this decision are:

- `20260712_regex_suffix_audit_bytecode_closure_stack.cpu.pprof`
- `20260712_regex_is_match_bytecode_closure_stack.cpu.pprof`
- `20260712_regex_set_audit_bytecode_closure_stack.cpu.pprof`

## Compiled application gate

Each default application was rebuilt from the same canonical stdlib under the
one-core, 1 GiB guardrail. These single seed runs are directional, but none
regressed materially: suffix improved from 3.8700s to 3.7700s, ordinary
matching was neutral at 1.9600s to 1.9900s, and RegexSet improved from
0.3300s to 0.3200s. The verification outputs above establish that all three
are the same applications as their baselines.

## Next recommendation

Refresh the exact allocation profiles after this kept change before adding
another buffer-reuse rule. The next likely shared wall is construction of
next-state/thread arrays in `regex_nfa_move`, but that inference needs fresh
evidence now that closure-stack allocations are gone. The gate entails the
same bounded start/end generated-main captures for suffix, ordinary matching,
and RegexSet, followed by a candidate only if one concrete allocator repeats
across all three. This preserves capture and scanner isolation and prevents a
workload-shaped optimization.
