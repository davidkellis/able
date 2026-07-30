# Post-nullable cross-family architecture ownership reconciliation

## Decision

Close `cross-family-architecture-ownership`, retain no production change, and
rebase the 12 jointly reviewed compiler-scope closures.

The ten post-nullable compiled family records partition all 65 current
compiled frontier rows exactly once. Every row has current strict
verifier-backed evidence and no interpreter dependency. The primitive
nullable carrier materially reaches 22 applications, but the common reached
operation is already a direct, allocation-free validity check.

`bridge.ToInt` is the only exact material leaf that crosses three unlike
compiled families. It is not an open candidate: its general global integer
cache already slowed the allocation-light TapeLang guard by 4.17%, and its
remaining calls occur at explicit semantic boundaries. No changed
generated-code or generated-runtime owner passes both the breadth and
open-mechanism gates.

## Complete compiled partition

The current family evidence reconciles exactly to the 65-row compiled
frontier:

| Closure | Rows | Material carrier reach | Post-carrier residual |
| --- | ---: | ---: | --- |
| Target guards | 6 | 0 | Protected native owners remain outside the carrier |
| Current control | 3 | 0 | Direct recursion, matrix loop, and Tape dispatch stay separate |
| Iterator/control | 8 | 4 | Native results plus split semantic boxes and nominal work |
| Text/map | 9 | 7 | Named-map dynamic boundaries plus unrelated text/nominal work |
| Regex | 6 | 6 | Direct validity checks; NFA execution remains the owner |
| Float numeric | 4 | 1 | One ignored native loop result; floats remain direct `float64` |
| Wide numeric | 3 | 0 | Direct wide arithmetic and general nominal results |
| Byte output | 2 | 0 | Native byte Arrays, borrowed output, cold host boundary |
| Sudoku quotient | 1 | 0 | One absent result after work; quotient remains one-application |
| Concurrency | 23 | 4 | Three native receives and one explicit callable ABI |
| **Total** | **65** | **22** | **No unreviewed row** |

The partition audit reports 65 rows, 65 unique rows, no missing row, no extra
row, and zero interpreter dependencies. Forty-one recorded graphs contain 96
packages and Base64's host-backed graph contains 119; the concurrency
telemetry record separately establishes 96 packages for each of its 23 rows.

## Cross-family exact-owner gate

### Retained carrier

The former primitive-nullable allocation owner is gone. Generic Slot Buffer,
Inventory Reconciliation, and Transaction Ledger Audit retain verifier-backed
five-run A/B improvements of 39.29%, 9.84%, and 16.67%, with exact object
reductions of 99.60%, 48.88%, and 8.88%.

Across iterator/control, text/map, Regex, and concurrency, present/absent
inspection is now a native `.valid` read and present values remain in their Go
scalar carriers. Recurrence of an operation that is already direct and
allocation-free is not a residual optimization owner.

### `bridge.ToInt`

Retained current profiles make `bridge.ToInt` material in three distinct
families:

- iterator/control: Binary Event Log and Dependency Plan;
- text/map: K-Nucleotide and other explicit dynamic map paths; and
- concurrency: Await Channel Mux, Validated Job Pipeline, Concurrent Stateful
  Pipeline, and Concurrent Event Routing.

This is real three-family reach, but the route is closed. These calls convert
runtime semantic values at explicit dynamic, interface, callable, nominal, or
service boundaries. A general global integer cache was tested and slowed the
unrelated allocation-light TapeLang guard by 4.17%. Repeating that cache or
narrowing it by benchmark, family, runtime count, or nominal type would violate
the retention rules.

### Other recurrent parents

| Owner | Family breadth | Disposition |
| --- | ---: | --- |
| Named `HashMap` dynamic conversion boundary | 2 | Insufficient breadth; named-container compiler rule forbidden |
| Await nullable callback conversion | 1 | Explicit runtime-callable ABI; insufficient breadth |
| Checked signed arithmetic | 2 material, one sparse | Already closed by mixed broad A/B results |
| `bridge.ToString` and positional semantic structs | 2 | Insufficient breadth |
| Output/entry host conversion | several | Once per process and not material |
| Allocation and GC | 5 | Aggregate Go parents with different Able descendants |
| `currentGID` | 1 family, 2 applications material | Broad execution-context alternatives rejected |

Non-primitive nominal construction, regex NFA storage, wide nominal results,
and Future/Channel/Mutex semantic representations remain distinct mechanisms.
They cannot be combined into one candidate merely because their Go ancestry
eventually reaches allocation or GC.

## Measurement decision

No CPU profile, allocation profile, wall-time cohort, or A/B implementation
was added.

Why: all changed carrier paths have one of four complete dispositions:
eliminated allocation, direct native consumption, an explicit boundary with
less than three-family breadth, or a non-material entry/host path. The sole
exact three-family leaf is unchanged and already failed its broad guard.
Another measurement cannot convert a rejected implementation into an open
general rule.

No compiler, generated runtime, runtime package, tree-walker, bytecode VM,
canonical stdlib, benchmark, language, dependency, nominal special case, or
WASM source changed.

## Ledger transaction and verification

The cross-family record completes the causal review required by the
compiler-production identity change. The named partial-advance path correctly
refused to replace a drifted shared scope. After all affected closures were
reviewed, the authoritative ledger was therefore rebased in one complete
transaction so no closure could inherit a compiler scope that another had not
reviewed:

- `compiled-architecture-target-budget`;
- all ten compiled frontier closures; and
- `cross-family-architecture-ownership`.

The resulting ledger has 23 current closures, zero invalidations, and an empty
selector. All five frontier tests and seven ledger tests passed, with the
ledger suite's one expected skip. Both direct generated-file checks pass, and
`go test ./cmd/ablec` passed in 5.682 seconds. JSON and whitespace checks are
recorded with the final handoff. The machine-readable companion is:
`2026-07-30-post-nullable-cross-family-architecture-ownership-reconciliation.json`.

No raw build or profile workspace was created in this evidence-only tranche.
The exact 122 MiB disk-backed Go test cache was removed after verification,
and no matching disposable artifact remains under `/var/tmp` or `/tmp`.

## Next

Do not start another performance implementation until a production,
benchmark, or semantic change invalidates a closure. Use the next bounded
tranche for v12 correctness and release verification.

Why: the current 130-row frontier has zero actionable groups and the complete
23-entry closure ledger is current. Reprofiling unchanged programs or
reopening the rejected integer cache would add noise without providing an
admissible mechanism.

What it entails: keep the scorecard, frontier, and ledger checks green; run
bounded correctness/release guards within the one-minute per-test rule; and
address a concrete v12 compiler, interpreter, spec, or canonical-stdlib defect
if one is found. A later legitimate production or benchmark change will
re-open only the affected performance closures.

Why it matters: this preserves the native-carrier and interpreter-free gains,
prevents failed or benchmark-shaped routes from returning without new
evidence, and ensures the next optimization starts from a real changed
mechanism.
