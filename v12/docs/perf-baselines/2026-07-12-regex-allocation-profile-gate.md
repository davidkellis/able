# Regex cross-application allocation profile gate — 2026-07-12

## Decision

Keep no additional runtime, compiler, bridge, or stdlib change in this
tranche. Exact allocation-space captures identify one credible next candidate:
the per-closure temporary stack created by `regex_nfa_add_closure`, together
with tagged-thread insertion in `regex_nfa_upsert_thread`. Those descendants
recur in public captured matching, ordinary matching, and `RegexSet` matching.

The next experiment may thread a matcher-local, reusable closure work stack
through tagged movement/closure. It must keep the stack local to one execution
(or scanner instance), preserve closure order, and be measured across the same
three applications. It must not recognize capture-free patterns, `RegexSet`, a
specific corpus, or an individual pattern shape.

## Method and limits

Each generated binary was rebuilt with canonical external `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The opt-in phase profiler
writes exact start/end `allocs` snapshots around registered `main`. Suffix and
independent matching use their normal `--profile` modes because an exact
full-input suffix capture exceeded the 90-second diagnostic guard; `RegexSet`
uses its verifier-backed default workload.

The profiler itself allocates while writing an exact profile and starting its
CPU hook. Attribution below excludes `runtime/pprof` and `profilehook` stacks;
the phase-stat byte/allocation deltas are retained as scope checks, not as
normal-runtime performance measurements.

| Workload | Main allocation delta | Main allocation count | Common application allocators |
| --- | ---: | ---: | --- |
| suffix audit (`--profile`) | 23,104,880 B | 287,518 | closure 1.90 MB; upsert 0.80 MB; fresh captures 0.22 MB; codepoints 1.96 MB. |
| independent match (`--profile`) | 13,818,088 B | 250,810 | closure 0.50 MB; upsert 0.29 MB; fresh captures 0.16 MB; codepoints 2.99 MB. |
| `RegexSet` audit (default) | 80,851,712 B | 1,678,779 | closure 11.46 MB; upsert 5.12 MB; fresh captures 1.90 MB; codepoints 8.43 MB. |

`RegexSet` patterns have no captures, so its `regex_nfa_new_captures` allocation
shows that an empty capture array is still built at every start boundary. That
is useful follow-up evidence, but it is not by itself the next candidate:
suffix matching has real capture state and needs a different semantic gate.

The stronger shared first change is closure scratch reuse. Every current call
to `regex_nfa_add_closure` creates an `Array RegexNFAThread` stack; movement
invokes closure repeatedly, and all three shapes repeatedly allocate both that
stack and tagged thread records. A local reusable stack can remove the former
without changing capture/tag values or sharing mutable state across matchers.

## Retained captures

The start/end allocation profiles and phase stats are retained in
`v12/interpreters/go/.profiles/` with the prefixes:

- `20260712_regex_suffix_audit_profile_*allocs.pprof`
- `20260712_regex_is_match_profile_*allocs.pprof`
- `20260712_regex_set_audit_*allocs.pprof`

## Next recommendation

Evaluate a generic matcher-local closure-stack reuse candidate. It entails
passing one scratch `Array RegexNFAThread` through `find_nfa_span`,
`RegexSet`, and scanners, proving no thread/capture escapes from a completed
closure, then rerunning all regex fixtures plus before/after profiles and
verifier-backed measurements for suffix audit, independent match, and
`RegexSet`. Why: it directly targets the first concrete allocation child that
repeats across all three public API shapes, unlike capture-free or set-only
shortcuts.
