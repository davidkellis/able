# RegexSet cross-application profile gate — 2026-07-12

## Decision

The third public regex application confirms a reusable canonical-stdlib
candidate: the tagged-NFA transition walk and epsilon-closure work recur in
three independently shaped regex applications. Keep the present code in this
tranche; the next tranche should evaluate one general NFA representation or
traversal improvement against all three applications and the existing regex
fixtures. It must not add compiler lowering, a `RegexSet`-only branch, or a
benchmark/corpus-specific shortcut.

## Workload and references

`regex-set-audit` compiles four public `RegexSet` patterns once, then performs
four deterministic classification passes over the first 512 ENABLE words. It
exercises combined-NFA matching, anchored character classes, file input, Array
results, and aggregation. Its verifier requires `2048:104:20840`.

The Able source is mirrored under the v12 examples and sibling benchmark
corpus. Fresh three-run, verifier-checked references were:

| Reference | Mean real time |
| --- | ---: |
| Go | 0.0042s |
| Python | 0.0183s |
| Ruby | 0.0437s |

`--profile` is an explicit bounded mode only: it processes 128 words and
returns `512:28:5660`; default verifier runs remain at 512 words.

The external comparison harness also completed one verifier-checked default
bytecode run in 40.0900s. That large gap to the reference runtimes is baseline
evidence, not justification for a set-, word-list-, or pattern-specific path.

## Bytecode evidence

All bytecode captures used `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and
the canonical external stdlib. The warmed runtime benchmark excludes parser
and bootstrap work. One profile-mode call measured `10,027,294,045 ns/op`,
`113,892,088 B/op`, and `924,887 allocs/op`; its 9.98-second CPU sample is
retained as:

- `v12/interpreters/go/.profiles/20260712_regex_set_audit_bytecode.cpu.pprof`

Cached identifier resolution is 28.26% cumulative, including 27.96% through
its cached lookup child. That concrete VM path also appears in suffix audit
(24.65%) and independent regex (11.76%), so it is a separate future generic
VM candidate. It is not coupled to the compiled NFA candidate and should not
be changed in the same experiment.

## Compiled evidence

The generated binary was built once under the same one-core, 1 GiB guardrail.
Its seed verifier-checked launch took 0.3900s with eight GCs. Twenty-three
additional verified default launches produced a 5.04-second merged `main`
CPU sample; only `main` phase profiles were merged, excluding bootstrap:

- `v12/interpreters/go/.profiles/20260712_regex_set_audit_compiled_main_merged.cpu.pprof`

| Application | Repeated generic descendants |
| --- | --- |
| `regex_suffix_audit` | `regex_nfa_move` 43.63% cumulative; `regex_nfa_add_closure` 22.89%; allocation 42.35%. |
| `regex_is_match_small` | `find_nfa_span` 68.66% cumulative; generic NFA/alloc work beneath it. |
| `regex_set_audit` | `regex_nfa_move` 52.58%; `regex_nfa_add_closure` 30.36%; allocation 39.88%. |

The direct source of the repeated work is general: `regex_nfa_move` and
`regex_nfa_add_closure` repeatedly scan the NFA transition list while moving
and closing active threads. A per-source-state outgoing-transition index (or
equivalent general traversal representation) is therefore the next candidate
to measure. It would benefit arbitrary regexes, sets, and scanner paths,
rather than recognizing a named API or these benchmark patterns.

## Disk hygiene

The generated source tree and temporary binary were removed after the merged
profile was copied into the existing profiles directory. No generated build
tree remains under the project or `/tmp`.
