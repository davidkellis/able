# Regex post-closure allocation gate — 2026-07-12

## Decision

Keep no additional runtime, compiler, or stdlib implementation change in this
evidence tranche. Exact post-change allocation captures confirm that the kept
matcher-local closure-stack change reduced every application, then identify a
new concrete shared matcher allocation: each successful NFA thread insertion
creates an identical second `RegexNFAThread` solely for closure traversal.

The next candidate may return the newly inserted private thread from
`regex_nfa_upsert_thread` and place that same immutable record on the closure
stack. This is a general tagged-NFA data-ownership improvement: it applies to
ordinary matching, RegexSet, and scanners without inspecting patterns,
captures, containers, or input corpora. It must preserve replacement order and
never reuse a thread whose state/start/captures were not just accepted.

## Method and limits

Each generated binary was rebuilt with canonical external `able-stdlib` under
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The opt-in
`ABLE_GO_PHASE_PROFILE_DIR` records exact start/end allocation snapshots around
registered `main`. Suffix audit and independent matching used their established
bounded `--profile` input; RegexSet used its default verifier-sized input.

All three generated profile runs completed successfully, and RegexSet passed
its Ruby verifier. The allocation hook itself allocates while serializing
profiles, so `runtime/pprof` and `profilehook` paths are excluded from source
attribution. The phase totals are scope checks, not normal-runtime timing
measurements.

| Workload | Before closure reuse | Post-change main delta | Difference |
| --- | ---: | ---: | ---: |
| suffix audit (`--profile`) | 23,104,880 B / 287,518 allocs | 22,838,056 B / 279,253 allocs | -266,824 B / -8,265 allocs |
| independent match (`--profile`) | 13,818,088 B / 250,810 allocs | 13,635,048 B / 242,731 allocs | -183,040 B / -8,079 allocs |
| `RegexSet` audit (default) | 80,851,712 B / 1,678,779 allocs | 78,927,152 B / 1,619,085 allocs | -1,924,560 B / -59,694 allocs |

## Attribution

The regenerated source profiles show that allocating a fresh next-state Array
in `regex_nfa_move` remains real but is not the leading shared ownership
problem: it accounts for 7,528, 11,244, and 34,176 objects in suffix,
ordinary matching, and RegexSet respectively.

More importantly, the two private thread-record sites have exactly matching
object counts in every workload:

| Workload | Closure-stack thread records | Upserted active-thread records |
| --- | ---: | ---: |
| suffix audit | 41,352 | 41,352 |
| independent match | 14,532 | 14,532 |
| `RegexSet` audit | 283,000 | 283,000 |

The source confirms the relationship. On a successful upsert,
`regex_nfa_upsert_thread` materializes `RegexNFAThread { state, start,
captures }` for `threads`; `regex_nfa_add_closure` immediately materializes
the same values again for `scratch`. Since neither record is mutated after
acceptance, returning the inserted record is a narrow, semantics-preserving
way to remove the duplicate rather than pooling state across executions.

The exact profiles also expose a separate compiler-side primitive boundary:
`__able_char_to_codepoint` plus its runtime cast contributes 118,512,
111,468, and 853,768 objects. That is potentially a general kernel-helper
lowering opportunity, but all current evidence comes from regex execution.
It is deliberately not selected here until a non-regex character-processing
application shows the same compiled leaf.

## Retained captures

`v12/interpreters/go/.profiles/` retains per workload:

- `20260712_regex_*_post_closure_main_start.allocs.pprof`
- `20260712_regex_*_post_closure_main_end.allocs.pprof`
- `20260712_regex_*_post_closure_phase_alloc_stats.json`

Generated source trees and binaries were removed after those small snapshots
were retained.

## Next recommendation

Evaluate the private newly-upserted-thread reuse candidate. It entails changing
the internal upsert result from a boolean to an optional newly accepted thread,
then pushing that returned thread into closure traversal rather than creating a
second identical struct. Verify all regex fixtures in tree-walker and bytecode
modes, generic automata execution, and the same three bytecode and compiled
application gates. Why: it is the largest exact shared matcher allocation that
remains after stack-array reuse, and its proof is independent of capture count,
regex API, pattern shape, or corpus.
