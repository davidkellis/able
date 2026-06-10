# Truthiness/cast wide-text closure refresh

Date: 2026-07-21

## Decision

The `compiled-wide-numeric` and `bytecode-text-map` closures are current
against the post-truthiness/cast shared interpreter semantics. Keep no
compiler, VM, runtime, canonical-stdlib, benchmark, reference, language, or
WASM change from this tranche.

Fresh repeated timing confirms large remaining product gaps, but exact reach
does not admit a candidate. All three compiled-wide programs make zero calls
to generated truthiness or explicit-cast bridges. All five bytecode text/map
programs make zero changed Error fallbacks and zero cast failures. Only
K-Nucleotide reaches the generic cast wrapper, with 28 successful calls per
main; that is neither CPU-plausible nor shared across three unlike programs.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every retained timing process passed
  its public verifier and used its catalog directory, arguments, serial CPU
  budget, and a 55-second cap.
- Arithmetic means retain every successful sample. Inventory Reconciliation
  and Unicode Scalar Pipeline received a second Python cohort because their
  first cohorts were volatile; no sample was removed or replaced.
- Able executions use `GOMEMLIMIT=1GiB`, `GOGC=50`, and the catalog-resolved
  serial CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Fixed Width 128 | compiled | 5 | 0.200 s | Go | 5 | 0.005724 s | 34.939x |
| Rational Series | compiled | 5 | 0.128 s | Go | 5 | 0.015109 s | 8.472x |
| Wide Integer Records | compiled | 5 | 0.188 s | Go | 5 | 0.047009 s | 3.999x |
| I Before E | bytecode | 5 | 0.502 s | Python | 5 | 0.085586 s | 5.865x |
| Inventory Reconciliation | bytecode | 5 | 2.640 s | Python | 10 | 0.085564 s | 30.854x |
| K-Nucleotide | bytecode | 5 | 42.826 s | Ruby | 5 | 1.312821 s | 32.621x |
| Unicode Scalar Pipeline | bytecode | 5 | 3.654 s | Python | 10 | 0.246245 s | 14.839x |
| Word Frequency | bytecode | 5 | 1.372 s | Python | 5 | 0.019756 s | 69.447x |

All 115 retained timing processes represented by these decisions verified
with zero failures and zero timeouts: 30 compiled/Go processes and 85
bytecode/Python/Ruby processes. The additional Python cohorts are pooled with
the originals. Their retained pooled CVs are 33.51% for Inventory and 28.26%
for Unicode; the means therefore describe this workstation rather than
pretending those lanes are noise-free.

## Exact reach

Temporary debug counters were placed immediately at the changed semantic
boundaries, used only in untimed processes, and then removed.

| Application | Mode | Census processes | Truthy checks/process | Changed Error fallback | Explicit casts/process | Cast failures |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Fixed Width 128 | compiled | 1 | 0 | 0 | 0 | 0 |
| Rational Series | compiled | 1 | 0 | 0 | 0 | 0 |
| Wide Integer Records | compiled | 1 | 0 | 0 | 0 | 0 |
| I Before E | bytecode | 2 | 694,932 | 0 | 0 | 0 |
| Inventory Reconciliation | bytecode | 2 | 0 | 0 | 0 | 0 |
| K-Nucleotide | bytecode | 2 | 63,124,394 | 0 | 28 | 0 |
| Unicode Scalar Pipeline | bytecode | 2 | 1,769,472 | 0 | 0 | 0 |
| Word Frequency | bytecode | 2 | 415,293 | 0 | 0 | 0 |

Every repeated bytecode count reproduced exactly and every successful census
output passed its verifier. The initial all-opcode K-Nucleotide census was too
intrusive and reached the guard twice without producing a snapshot. A narrowed
semantic-only debug binary then completed and verified twice under the same
guard. Those two full-scale results, not the guard-capped attempts, supply the
reported counts.

The four text workloads with truthiness calls use primitive checks and never
enter the corrected non-primitive Error matcher. The 28 K-Nucleotide casts all
succeed; the new error-to-raise branch is never entered. Because the only
reached cast wrapper is one tiny single-application leaf, the profile admission
gate fails and no CPU profile or candidate is warranted.

The concise census is retained in
`2026-07-21-truthiness-cast-wide-text-closure-reach.json`; raw compiled
telemetry is retained in
`2026-07-21-truthiness-cast-wide-text-closure-compiled-reach.json`. No
diagnostic counter, binary, raw profile, or generated package remains in
production code.

## Exact timing artifacts

- Compiled Able and Go reference reports:
  `2026-07-21-truthiness-cast-wide-text-closures-compiled.json`
  (`96c841a5ff94a5c93a3e69b7fda31e1970ad4a40e049fe0de0c712f0ff45a44c`),
  `2026-07-21-truthiness-cast-wide-text-closures-go-reference.json`
  (`f18bff7d59c3d7b4d0f33c048188e0feeb0646b3aba548a0a81f11b2d1d98698`).
- Bytecode Able and initial interpreter references:
  `2026-07-21-truthiness-cast-wide-text-closures-bytecode.json`
  (`a4f2cdbe9a17e061b4d6c912abf5df14059c2316462c8aa459eaf088f4cb1723`),
  `2026-07-21-truthiness-cast-wide-text-closures-interpreter-reference.json`
  (`d4f66e4c4fafb12a8de20e801e3646d8b041db06768fcf184e1af3385d52a406`).
- Inventory/Unicode second Python report:
  `2026-07-21-truthiness-cast-wide-text-closures-python-c2-reference.json`
  (`1433ab52573d362f2672376eefc485b38c09a0937a6b467fd0a9b49b9ad0f7bd`).
- Compiled reach telemetry:
  `2026-07-21-truthiness-cast-wide-text-closure-compiled-reach.json`
  (`c0ef0b9a25828e4be170f935109a2b3f54ab3d643806dddd68844290c7cd1924`).

## Next recommendation

Refresh `compiled-text-map` and `bytecode-byte-output` next.

Why: compiled text/map is the same dynamic protocol family just measured in
the VM, but its compiler ownership and generated nominal/map lowering remain
independently invalidated. Bytecode byte/output is the nearest remaining
unrefreshed VM closure and tests a substantially different Array/u8, string,
and output path. Pairing them continues the checked revalidation without
mistaking one mode's evidence for the other's or over-weighting numeric work.

What it entails: reuse the frozen sources and current toolchains; collect five
verifier-backed processes per ordinary lane; add matched cohorts when retained
workstation variance is material; run exact changed-path reach before any
profile; and profile only a materially reached concrete leaf. Advance only
those two closures. Build a candidate only if one generic mechanism is
material in at least three unlike applications and preserves the current
target guards. No WASM work is involved.
