# Truthiness/cast text-byte next closure refresh

Date: 2026-07-21

## Decision

The `compiled-text-map` and `bytecode-byte-output` closures are current against
the post-truthiness/cast shared interpreter semantics. Keep no compiler, VM,
runtime, canonical-stdlib, benchmark, verifier, reference, language, or WASM
change from this tranche.

Fresh repeated timing confirms the remaining product gaps but does not admit a
generic candidate. All five compiled text/map programs have zero reach into
generated truthiness or explicit-cast helpers. The three bytecode byte/output
programs have zero changed Error fallbacks and zero explicit casts. Reverse
Complement makes only four primitive truthiness checks per main; Base64 and
FASTA make none.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every timing process passed its public
  verifier and used its catalog directory, arguments, serial CPU budget, and
  a 55-second cap.
- Arithmetic means retain every successful sample. I Before E and Word
  Frequency received matched second compiled/Go cohorts after volatile first
  lanes; no sample was removed or replaced.
- Able executions use `GOMEMLIMIT=1GiB`, `GOGC=50`, and the catalog-resolved
  serial CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| I Before E | compiled | 10 | 0.135 s | Go | 10 | 0.084251 s | 1.602x |
| Inventory Reconciliation | compiled | 5 | 0.196 s | Go | 5 | 0.009098 s | 21.544x |
| K-Nucleotide | compiled | 5 | 2.670 s | Go | 5 | 0.068445 s | 39.009x |
| Unicode Scalar Pipeline | compiled | 5 | 0.262 s | Go | 5 | 0.013081 s | 20.029x |
| Word Frequency | compiled | 10 | 0.157 s | Go | 10 | 0.006175 s | 25.425x |
| Base64 | bytecode | 5 | 2.790 s | Ruby | 5 | 2.578351 s | 1.082x |
| FASTA Generation | bytecode | 5 | 1.812 s | Python | 5 | 0.207096 s | 8.750x |
| Reverse Complement | bytecode | 5 | 3.250 s | Python | 5 | 0.025976 s | 125.116x |

All 115 timing processes represented by these decisions verified with zero
failures and zero timeouts: 70 compiled/Go processes and 45
bytecode/Python/Ruby processes. The pooled I Before E lanes remain noisy at
39.24% Able CV and 26.66% Go CV; pooled Word Frequency is 22.07% Able CV. All
samples remain in the means, as required for this workstation.

Base64 is faster than Python but is 1.082x Ruby, so it remains narrowly outside
the strict faster-reference target. FASTA and Reverse Complement remain large
misses. This focused refresh does not rewrite the full external scoreboard.

## Exact reach

Temporary debug counters were placed immediately at the changed semantic
boundaries, used only in untimed processes, and then removed.

| Application | Mode | Census processes | Truthy checks/process | Changed Error fallback | Explicit casts/process | Cast failures |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| I Before E | compiled | 1 | 0 | 0 | 0 | 0 |
| Inventory Reconciliation | compiled | 1 | 0 | 0 | 0 | 0 |
| K-Nucleotide | compiled | 1 | 0 | 0 | 0 | 0 |
| Unicode Scalar Pipeline | compiled | 1 | 0 | 0 | 0 | 0 |
| Word Frequency | compiled | 1 | 0 | 0 | 0 | 0 |
| Base64 | bytecode | 2 | 0 | 0 | 0 | 0 |
| FASTA Generation | bytecode | 2 | 0 | 0 | 0 | 0 |
| Reverse Complement | bytecode | 2 | 4 | 0 | 0 | 0 |

Every census process completed under its guard and passed the public verifier;
all repeated bytecode counts reproduced exactly. The four Reverse Complement
checks are primitive and return before the corrected non-primitive Error
matcher. Because no changed path is reached in either closure, the profile
admission gate fails and no CPU profile or candidate is warranted.

The concise census is retained in
`2026-07-21-truthiness-cast-text-byte-next-closure-reach.json`; raw compiled
telemetry is retained in
`2026-07-21-truthiness-cast-text-byte-next-closure-compiled-reach.json`. No
diagnostic counter, binary, raw profile, or generated package remains in
production code.

## Exact timing artifacts

- Initial compiled Able and Go reports:
  `2026-07-21-truthiness-cast-text-byte-next-closures-compiled.json`
  (`fa60ea9a5810085e554a92ab84a94e7ff3e7ccc2c6ae93d215e6f4e72f413101`),
  `2026-07-21-truthiness-cast-text-byte-next-closures-go-reference.json`
  (`4b5af7bdc91abfc97bcb4b63aa490a3be6eaea898bf62428817234e104c2aae0`).
- I Before E/Word Frequency second compiled and Go reports:
  `2026-07-21-truthiness-cast-text-byte-next-closures-compiled-c2.json`
  (`27805b64b6424c6e80662228976944a284ca7f299cba1a176a8985c90cc5dfd0`),
  `2026-07-21-truthiness-cast-text-byte-next-closures-compiled-c2-go-reference.json`
  (`bf76c6ffca7c26564085be796dfbc2dc604b0614ce2571dba24797d02912de99`).
- Bytecode Able and interpreter references:
  `2026-07-21-truthiness-cast-text-byte-next-closures-bytecode.json`
  (`0b16eacb8943dc9c1f3a850155e5b0b32980d7fe71d8e0cc9fa8c7aff9bb2c00`),
  `2026-07-21-truthiness-cast-text-byte-next-closures-interpreter-reference.json`
  (`84cbe514b1b55f40b05d30c0f1da9998479cdd052f2507534d2572e94a47a0ce`).
- Compiled reach telemetry:
  `2026-07-21-truthiness-cast-text-byte-next-closure-compiled-reach.json`
  (`32a849b86f7cfb4dea2abc77da4b8a8f22a6b12a15e2fdbd82a043411d908c26`).

## Next recommendation

Refresh `compiled-byte-output` and `bytecode-regex` next.

Why: compiled byte/output is the independently invalidated compiler side of
the Array/u8, string, and output family just measured in the VM. Bytecode regex
is the nearest remaining VM closure whose result-bearing, matching, and Error
protocol paths could plausibly reach the corrected semantics. This pair keeps
mode-specific evidence separate while moving into a different semantic family.

What it entails: reuse the frozen sources and current toolchains; collect five
verifier-backed processes per ordinary lane; add matched cohorts when retained
workstation variance is material; run exact changed-path reach before any
profile; and profile only a materially reached concrete leaf. Advance only
those two closures. Build a candidate only if one generic mechanism is
material in at least three unlike applications and preserves current target
guards. No WASM work is involved.
