# Bytecode candidate-admission inventory — 2026-07-15

## Scope

This is a read-only reconciliation of the retained bounded bytecode profile
reports. It does not rerun an unchanged workload, create a synthetic timing
loop, or claim a new timing result. Its purpose is to decide whether the
completed execution-lane study leaves an already-recorded VM candidate that
can honestly be reopened.

## Evidence

| Unlike workload set | Repeated-looking frame | Material descendants | Admission result |
| --- | --- | --- | --- |
| split/join text, iterator collect, numeric Array map | VM call-dispatch parents | named-call return/coercion and type matching; generator/member dispatch; mono-Array slot transport | Parent only; no shared leaf. |
| Word Frequency, Document Audit, Dependency Plan | `finishInlineReturn` | string-key map/raw-integer; lazy iterator/member cache; Array/Queue/member/index | The only three-way leaf is the already-rejected return family. |
| Document Audit, Lexical Rollup, Word Frequency | cached member/`next()` work | iterator protocols in the first two; map/counting in Word Frequency | Repeats in two applications only, and the member-cache family is already guarded. |
| Fixed Width 128, Rational Series, K-Nucleotide | call/frame and conversion parents | UInt128 checked members; generic call/return; map/conversion/GC | Neither a concrete leaf nor a common representation boundary. |

The source records are:

- `2026-07-15-current-bytecode-three-shape-profile-gate.md`
- `2026-07-14-dependency-plan-profile-gate.md`
- `2026-07-14-bytecode-coverage-reference-profile-gate.md`
- `2026-07-15-full-coverage-scorecard-historical-reconciliation.md`

The current CPU-8-pinned execution-lane comparison adds no contrary result:
moving typechecking outside the measured process left Sudoku capped in both
lanes, Word Frequency within 0.01 seconds, and Future Pipeline unchanged.
See `2026-07-15-bytecode-full-vs-prechecked-paired.md`.

## Decision

No bytecode VM candidate is admitted. In particular, do not reopen raw
integer extraction, inline-return guard order, member-cache, call/frame, or
typecheck/cache work. Each either lacks three unlike material descendants or
already lost the broad benchmark guard.

The next eligible performance tranche requires a material cross-cutting
semantic/compiler/runtime change or a newly needed portable application, then
a bounded three-unlike-application profile that identifies a concrete
non-nominal leaf. Until then, retain the current scorecard and reports as
regression evidence rather than spending host budget on unchanged profiles.
