# Regex tagged-NFA outgoing-transition index — 2026-07-12

## Decision

Keep the canonical-stdlib implementation. Every compiled `Regex` and
`RegexSet` now owns an immutable `RegexNFAIndex`: prefix offsets and stable
transition indices grouped by source state. Tagged movement and epsilon closure
visit only the current state's outgoing transitions, while retaining the
previous transition order (and reverse order for closure) exactly.

The index is deliberately private to compiled regex execution plans. Public
`NFA` values and `RegexProgram.nfa_snapshot()` remain mutable, use their
existing transition-array semantics, and carry no stale cache. This is a
general tagged-NFA improvement for ordinary matching, replacement/splitting,
iterators, scanners, and `RegexSet`; it is not compiler lowering,
API-specific logic, or pattern/corpus recognition.

## Verification

- All `exec/14_*` regex fixtures passed in both tree-walker and bytecode modes
  under `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.
- `regex_suffix_audit` and `regex_set_audit` compiled binaries passed their
  Ruby verifiers on every profiled launch; independent `regex_is_match_small`
  retained its deterministic `199960002` output.
- Generic `automata_dfa_small` bytecode execution retained `120000000`.
- Generated source trees and binaries were removed after retaining only the
  merged main CPU profiles.

## Bytecode runtime gate

The warmed profile lane uses the same bounded inputs as the preceding evidence
gate, excluding parsing and bootstrap. Lower time is better.

| Workload | Before | Indexed plan | Change |
| --- | ---: | ---: | ---: |
| suffix audit | 3,821,488,072 ns/op | 1,571,017,595 ns/op | -58.9% |
| independent `is_match` | 1,079,147,288 ns/op | 741,606,031 ns/op | -31.3% |
| `RegexSet` audit | 10,027,294,045 ns/op | 2,320,082,691 ns/op | -76.9% |

Allocation counts stay effectively flat: the index removes transition scans,
not Regex thread/tag/value allocation. The default verifier-backed `RegexSet`
application also fell from 40.0900s to 10.4400s in the final external bytecode
lane.

Retained final bytecode profiles:

- `20260712_regex_suffix_audit_bytecode_outgoing_index.cpu.pprof`
- `20260712_regex_is_match_bytecode_outgoing_index.cpu.pprof`
- `20260712_regex_set_audit_bytecode_outgoing_index.cpu.pprof`

## Compiled runtime gate

Each application was rebuilt with the canonical external stdlib under the
same one-core, 1 GiB guardrail. Seed process times are directional, since the
generated binaries are expensive to build; the merged profiles give the
stable attribution.

| Workload | Before seed | Indexed seed | Profiled NFA change |
| --- | ---: | ---: | --- |
| suffix audit | 4.9200s | 3.8700s | move 6.10s → 4.06s; closure 3.20s → 2.12s. |
| independent `is_match` | 2.0500s | 1.9600s | effectively neutral: move 0.39s → 0.47s; closure 0.44s → 0.48s. |
| `RegexSet` audit | 0.3900s | 0.3300s | move 2.65s → 1.45s; closure 1.53s → 0.51s. |

The independent small-pattern case is neutral rather than a material win, so
the optimization clears the broad gate without claiming a universal speedup.
The remaining shared compiled wall is allocation and code-point/tag work,
not repeated unrelated-transition scanning.

Retained final merged main profiles:

- `20260712_regex_suffix_audit_compiled_outgoing_index_main_merged.cpu.pprof`
- `20260712_regex_is_match_compiled_outgoing_index_main_merged.cpu.pprof`
- `20260712_regex_set_audit_compiled_outgoing_index_main_merged.cpu.pprof`

## Next recommendation

Collect allocation-space attribution for the same three applications before
editing again. Allocation is now roughly half of each generated profile, but
the sources may be different: code-point records, tagged threads/captures,
and boundary conversion. A next candidate must identify one concrete shared
allocation child and remain neutral or better across all three applications
and regex fixtures; do not add a capture-free, `RegexSet`, or corpus-only
shortcut without that evidence.
