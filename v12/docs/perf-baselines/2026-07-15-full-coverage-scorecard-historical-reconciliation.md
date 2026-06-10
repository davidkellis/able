# Full-scorecard historical reconciliation — 2026-07-15

## Decision

Do not rerun a generated-source audit or bounded CPU profiles for Fixed Width
128, Rational Series, and K-Nucleotide merely because the full 32-application
scorecard makes their target misses visible together. That proposed follow-up
was stale: the current architecture already has the required independent
profile and lowering evidence, and it rejects a shared optimization.

The July 15 full scorecard remains the current reproducible selection and
regression baseline. It changes measurement coverage, not the source program,
compiler lowering, bytecode VM, generated runtime, canonical stdlib, or the
attribution of previously profiled work.

## Reconciled evidence

| Apparent grouping | Existing current-source evidence | Result |
| --- | --- | --- |
| Fixed Width 128 and Rational Series | The paired profile records Fixed Width's existing UInt128 checked-member path at 26.1% cumulative, while Rational is generic `execCallOpcode` (31.3%), `execCallName` (14.4%), and `finishInlineReturn` (10.0%). The lowering audit found only 37 ordinary `LoadSlot` instructions and no repeated specialized shape. | The common parent is not a concrete common leaf; previous call/frame variants also failed broad controls. |
| K-Nucleotide and other map/text controls | The generated-main follow-up finds `__able_hash_map_find_entry` at only 10.04% cumulative in K-Nucleotide, while conversion/allocation and raw map get/set dominate. Word Frequency has a materially different map/text shape, and the collision-safe and lazy index experiments regressed real K-Nucleotide. | No HashMap, key-type, conversion, or corpus-specific compiler/stdlib path is admissible. |
| Fixed Width, Rational Series, and K-Nucleotide together | The raw-value audit confirms that their recurring conversion samples cross different named-value contexts. A `RawValue` extension or compiler-only carrier would still materialize at ordinary storage, call, nominal, host, and Future boundaries. | A local carrier shortcut would be incomplete and benchmark-shaped. The full carrier proposal is separately rejected by the safe-Go representation feasibility audit. |

The source records are `2026-07-09-scorecard-tranche.md`,
`2026-07-14-compiled-static-metadata-main-profile-refresh.md`,
`design/raw-value-boundary-audit.md`, and
`design/primitive-value-representation-feasibility.md`.

## Consequence

No VM, compiler, generated-runtime, canonical-stdlib, fixture, or benchmark
change follows from this reconciliation. Repeating those profiles without a
material semantic or lowering change would consume the bounded measurement
budget while re-testing rejected named-workload routes.

Keep the full scorecard as the broad guard. The next performance candidate is
admissible only after a material, language-wide compiler/runtime or
semantic-portability change supplies a concrete non-nominal leaf in at least
three unlike verifier-backed applications. Until then, the next eligible
engineering work is an unfinished semantic or portability roadmap boundary,
with cross-runtime fixture parity first; its changed execution path may then
be profiled against the full scorecard.
