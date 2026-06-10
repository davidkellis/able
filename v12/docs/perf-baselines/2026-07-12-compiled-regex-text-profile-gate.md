# Compiled regex/text profile gate — 2026-07-12

## Decision

Keep no compiler, bridge, runtime, or canonical-stdlib performance change from
this tranche. The two independently authored regex programs repeat generic NFA
execution and allocation work, but the non-regex split/join control instead
spends in String conversion and joining. No concrete AOT bridge/value/dispatch
leaf repeats across all three binaries.

## Method

Each binary was built once with canonical external `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The generated source tree
is about 742–747 MB per build, so it was deleted immediately after its binary
and merged profile were retained. This leaves no generated build tree in the
project or `/tmp`.

Repeated binary launches used `ABLE_GO_PHASE_CPU_PROFILE_DIR`, the generated
launcher's CPU-only phase hook. Only `main.cpu.pprof` files were merged;
bootstrap, parser, registration, and allocation-snapshot costs are excluded.

| Workload | Verified launches | Seed compiled run | Merged main samples |
| --- | ---: | --- | ---: |
| `regex_suffix_audit` | 3 Ruby-verifier passes | 4.9200s real, 73 GCs | 13.98s |
| `regex_is_match_small` | 3 identical stdout hashes (`199960002`) | 2.0500s real, 38 GCs | 6.03s |
| `string_split_join_small` | 20 identical stdout hashes (`191484`) | 0.3900s real, 8 GCs | 4.18s |

Retained merged main profiles:

- `20260712_regex_suffix_audit_compiled_main_merged.cpu.pprof`
- `20260712_regex_is_match_compiled_main_merged.cpu.pprof`
- `20260712_string_split_join_compiled_main_merged.cpu.pprof`

## Attribution

| Workload | Material descendants | Interpretation |
| --- | --- | --- |
| Suffix audit | `regex_nfa_move` 43.63% cumulative, `regex_nfa_add_closure` 22.89%, allocation 42.35%, bridge `Cast` 11.80%. | Public named-capture matching is dominated by generic tagged-NFA thread work and its value allocation. |
| Independent regex | `find_nfa_span` 68.66%, allocation 51.91%, bridge `Cast` 13.43%, simple type match 12.94%. | Ordinary `is_match` repeats the general NFA/bridge family with a different pattern/result shape. |
| Split/join control | `String_join` 56.22%, allocation 46.65%, `String_split` 20.81%, `bridge.ToUint` 23.68%. | The non-regex text path is String/UTF-8 conversion and aggregation, not NFA or regex type matching. |

Allocation and GC are broad runtime parents, but their allocation sources are
not shared: NFA threads/tags in the regex applications versus strings, byte
arrays, and joins in the text control. Likewise, `Cast`/type matching repeat
only in regex while `ToUint` is material only in split/join. A bridge-wide or
compiler-wide change therefore has no evidence gate.

## Next recommendation

Add a third independently shaped public regex application, preferably using
`RegexSet` or `RegexScanner`, with verifier-backed Go/Python/Ruby references.
Then compare its compiled main profile with these two regex profiles. This
will distinguish a generally useful NFA-state/allocation improvement from a
two-workload pattern artifact; only a repeated NFA move/closure/allocation
child may justify a general canonical-stdlib change. Do not add regex-specific
compiler lowering, named-container rules, or benchmark-shaped VM paths.
