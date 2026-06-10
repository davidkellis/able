# Compiled scalar-character generality gate — 2026-07-12

## Decision

Keep no compiler, runtime, kernel, or canonical-stdlib performance change.
The generated `__able_char_to_codepoint` helper is material for generic DFA
matching, but it is not the same material operation across ordinary scalar
character programs. Optimizing that helper now would favor one dynamic
dispatch shape rather than improve character processing generally.

This gate repairs one benchmark-source type error only:
`run_length_encode_small` now explicitly narrows `text.len_bytes()` from `u64`
to `i32` for `StringBuilder.with_capacity`. Its fixed 20,000-byte input is in
range. The repair makes both interpreters and the compiled control executable;
it does not specialize the implementation or change a runtime path.

## Method

Each generated binary used canonical external `able-stdlib`, `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. `ABLE_GO_PHASE_PROFILE_DIR` captured exact
allocation snapshots around user `main`; profile-writer frames are excluded
from application attribution. Generated source trees were removed after their
small start/end snapshots and phase stats were retained.

| Workload | Main phase delta | Material helper finding | Dominant application path |
| --- | ---: | --- | --- |
| `run_length_encode_small` | 330,228,008 B / 7,870,882 allocs | 174,848 objects (2.71%) | `String.chars`, UTF-8 conversion, and byte bridge; the helper is reached while re-encoding output chars. |
| `levenshtein_small` | 2,646,744 B / 4,865 allocs | No material helper attribution | Concrete `Array char` comparisons lower directly; the short control is below useful helper-resolution after profiling overhead. |
| `automata_dfa_small` | 94,302,704 B / 2,190,713 allocs | 839,405 objects (42.60%) | Dynamic DFA step and `String.chars`; the helper is reached through character predicate dispatch. |

The two observed helper paths are not the same caller or cost shape:
run-length encoding is dominated by producing UTF-8 output, while DFA matching
uses dynamic scalar predicates. Levenshtein uses the same language primitive
but reaches neither path materially. Thus the required three-workload common
leaf/caller proof is absent.

The retained artifacts in `v12/interpreters/go/.profiles/` are prefixed with:

- `20260712_run_length_encode_compiled_`
- `20260712_levenshtein_compiled_`
- `20260712_automata_dfa_compiled_`

## Verification

- The repaired run-length control executed successfully in both tree-walker
  and bytecode modes.
- All three bounded compiled profiling controls built, ran, wrote their phase
  snapshots, and removed their temporary generated trees.

## Next recommendation

Profile the direct `String.bytes` boundary in `ascii_lower_small`,
`byte_histogram_small`, and `md5_hex_small` under the same bounded generated
main-phase mode. The first prior control already showed `String.bytes`, UTF-8
conversion, and `bridge.ToUint` as the leading application allocation family;
these three programs exercise that public byte-iteration API with independent
algorithm shapes. Only a repeated material leaf and caller across all three
would justify a general primitive-boundary candidate.
