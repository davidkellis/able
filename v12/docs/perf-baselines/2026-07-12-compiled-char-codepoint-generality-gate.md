# Compiled char-to-codepoint generality gate — 2026-07-12

## Decision

Keep no compiler, runtime, kernel, stdlib, fixture, or benchmark-source change.
The repeated compiled `__able_char_to_codepoint` allocation observed in the
three tagged-NFA applications is not yet a general language-program cost.

Three non-regex controls either transport `char` values without converting
them or deliberately process byte data. Their generated-main allocation
profiles contain no material execution of the helper, so changing a primitive
boundary now would encode a regex-shaped optimization.

## Method

Each generated binary used canonical external `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The opt-in
`ABLE_GO_PHASE_PROFILE_DIR` captured exact allocation snapshots around user
`main`. The phase profiler itself allocates while serializing profiles; its
`runtime/pprof` and `profilehook` frames are excluded from source attribution.

| Workload | Program shape | Main phase delta | Material non-profiler allocation finding |
| --- | --- | ---: | --- |
| `zigzag_char_small` | Repeated `Array char` redistribution | 194,460,208 B / 182,874 allocs | `zigzag` and `flatten` Array growth; no scalar conversion helper. |
| `ascii_lower_small` | String byte iteration and ASCII conversion | 74,738,144 B / 1,552,973 allocs | `String.bytes` UTF-8 conversion and `bridge.ToUint`; no char-codepoint helper. |
| `reverse_complement_small` | FASTA `Array u8` transform | 2,411,216 B / 138 allocs | The short main phase is below allocation-profile resolution after profiler overhead; its source is byte-only and contains no char conversion. |

The generated source defines shared char equality, ordering, and hashing
implementations that call the helper, so textual presence of the helper is not
execution evidence. The exact profiles show no allocation attribution to it in
any control. In contrast, the ASCII workload's 1,336,625-object profile has
`bridge.ToUint` (555,247 objects), `__able_string_from_builtin_impl` (458,301),
and UTF-8 decode (251,090) as its material application leaves.

The retained artifacts are the nine small start/end/stats files in
`v12/interpreters/go/.profiles/` prefixed with:

- `20260712_zigzag_char_compiled_`
- `20260712_ascii_lower_compiled_`
- `20260712_reverse_complement_compiled_`

Each generated source tree approached 800 MB and was removed after copying
those snapshots.

## Verification

- All three generated binaries built and exited successfully under the bounded
  guardrail; their one-run compiled controls measured 1.72s, 4.52s, and 1.22s
  respectively. These are process controls, not comparative performance
  claims.
- `go test ./pkg/profilehook -count=1 -timeout 60s` passed.

## Next recommendation

Profile `run_length_encode_small`, `levenshtein_small`, and
`automata_dfa_small` in the same mode. Unlike this first set, each naturally
iterates or compares `char` scalars. A subsequent candidate requires the same
concrete `__able_char_to_codepoint` helper and caller to be material across all
three. If that proof still fails, use the repeated `String.bytes` conversion
path instead—but only after it clears an independent cross-workload gate of its
own.
