# Post-publication compiled profile admission closure

Date: 2026-07-30

## Decision

Retain no production change and collect no duplicate profile cohort.

The requested interpreter-free compiled CPU/allocation refresh is already
complete under the current production identity. The checked invalidation
ledger has 23 current closures, zero invalidations, and an empty selector. Its
recommendation is explicit: do not rerun a closed performance tranche when no
checked evidence identity changed.

No compiler, generated runtime, runtime, canonical-stdlib, specification,
benchmark source, or frontier-row identity reopens the completed profiles.
There is therefore no admissible shared owner to implement or measure in an
A/B cohort.

## Current identity

- Local `HEAD`, `origin/master`, and remote `master`:
  `9c32f2777536da2c948327720acc75187973a6d9`.
- Selection manifest SHA-256:
  `6bbe6579df9857a791a2f30d55792bba0827994766d4f27beebbf0dba24ec628`.
- Current scoreboard SHA-256:
  `43c9a48d92ecaf02069655e6c1e78cc81bf1025997e578708da3c23acac8a4d8`.
- Frontier-evidence SHA-256:
  `6edd75cc356a4bdeb856d46d028e70cdb1d1774e8bad57b62f6b7c21fa7883c2`.
- Frontier SHA-256:
  `f4fa5b26a9f6229d0cb61a27af0ff8b965ba424732f73998463fd8640e5d312b`.
- Checked closure-ledger SHA-256:
  `aaf867b3b78add52fbf5b36efae45a4c097b9ae9e142f7bf826c754c942d3171`.
- Evaluated ledger SHA-256:
  `2d0f331971c938a7c086d8be95aa255bc9c34dfb6d7943686ce7d7debdc423b2`.

The checked production scopes remain:

| Scope | Files | Tree SHA-256 |
| --- | ---: | --- |
| compiler production | 287 | `f28265f71ff2e67df56d447aaeb90a26fd3af9927e2f719c630ed403e9c240ca` |
| runtime production | 40 | `73b8d6fcd4ee148ca8f928e139b0c69f37f7fcecc7b778dfecfcc9561ec8209d` |
| shared interpreter semantics | 137 | `0ba3255d7a3db4a798e616a30db8db9940038748804aa715d1d90b598403d9bf` |
| canonical stdlib | 70 | `382d256e2fb380220dcdd62a5cf83109fa72297f23d70bdd1ffe2d8daebed047` |
| v12 specification | 1 | `7083f1656a3452236a372c9b20e8efdcdf6f122681e04f7d6d8607099603e71f` |

## Existing profile coverage

The current compiler identity already has fresh strict profiles for six
unlike applications:

- Concurrent Event Routing;
- Distance Field;
- Fixed Width 128;
- K-Nucleotide;
- Manifest Normalization; and
- Policy Record Dispatch.

Every application was built with `--no-fallbacks`; each final graph contained
96 packages and omitted `able/interpreter-go/pkg/interpreter`. The retained
cohort has ten independent CPU-profile processes and three exact
allocation-profile processes per application, all verifier-backed. Its 42
readable reports and 18 exact phase-stat records remain under
`2026-07-30-post-nullable-compiled-architecture-owner-profiles/`.

The cohort found no concrete compiler or generated-runtime owner material in
three unlike families. The later complete ownership partition reconciled all
65 compiled rows. Its only exact leaf with three-family reach,
`bridge.ToInt`, is unchanged and closed: remaining uses are explicit semantic
boundaries, and the general global integer cache slowed the allocation-light
TapeLang guard by 4.17%.

The retained owner closure SHA-256 values are:

- Markdown:
  `d7fafabc99ba2266e99622264689824a0e22b3461f47b3ba4fd962a446621144`;
- JSON:
  `a726f0b0f252a7ef70669bb87e1b5b4792e36454bdbd212f2e9af26d4816fa08`;
  and
- complete cross-family reconciliation:
  `5fd2e70dd932871c163fab1fdcb1a9c374d66b1674800a3712ef7a42364d7832`.

## Admission and verification

- The scoreboard check passes for all 130 selected rows.
- All rows retain five successful Able and reference samples and 31 retained
  source/reference reports.
- All four path/evidence tests pass.
- The frontier check passes with 130 rows and zero actionable groups.
- All five frontier tests pass.
- The ledger reports 23 current closures and zero invalidations.
- The selector emits zero rows.
- All ten ledger tests pass with one intentional conditional skip.
- `go test ./cmd/ablec` passes in 6.090 seconds; the complete command took
  23.46 seconds with a cold disk-backed cache, below the one-minute limit.
- Whitespace validation passes.

Because the selector is empty, no strict rebuild, new pprof process, prototype,
or A/B timing cohort was admitted. Existing interpreter-free graph and profile
evidence remains current by exact checked identity.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
parser, canonical stdlib, benchmark, language, dependency, fixture, or WASM
source changed. No benchmark-specific, named-container, non-primitive nominal,
or broad execution-context rule was introduced.

The focused compiler run used an exact 127,280 KiB disk-backed workspace under
`/var/tmp`. It was removed after verification; no task-owned artifact remains.
A six-file generated Python cache created by a later validation pass was also
removed.

The machine-readable companion is
`2026-07-30-post-publication-compiled-profile-admission-closure.json`.

## Next

Audit the active non-WASM v12 correctness backlog and bounded test signals,
then select one concrete compiler, interpreter, specification, or canonical
stdlib defect if one exists.

Why: the performance frontier and all checked closures are current, so further
performance measurement requires a real production, semantic, benchmark, or
contradictory-evidence trigger.

What it entails: review `spec/TODO_v12.md`, current non-WASM skips and focused
test surfaces, run bounded checks under the one-minute rule, and choose only a
reproducible v12 defect. If a correction changes a checked performance scope,
let the ledger select the exact profiles that must be refreshed.

Why it matters: correctness work can supply a legitimate changed mechanism
without reopening rejected optimizations, while exact invalidation keeps the
next performance tranche tied to current code.
