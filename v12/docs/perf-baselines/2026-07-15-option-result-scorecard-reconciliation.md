# Option/Result Scorecard Reconciliation

## Scope

The portable `option_result_config` application was present in the 33-entry
coverage catalog but was absent from the current external scoreboard. The
first fresh five-run scorecard attempt exposed a semantic error in both Go
interpreters, so the failed report is retained as evidence and was not
promoted.

The failure occurred in the canonical `Result<T>.and_then` implementation:
an imported `case err: Error` typed pattern incorrectly used the short name
`Error` on an optimized matching path. The ordinary matching path already
resolved that name in its lexical environment to
`able.core.interfaces.Error`.

## Repair

All optimized typed-pattern entry points now use the same lexical type
canonicalization as the ordinary path:

- transient tree-walker clause matching;
- regular fast pattern matching; and
- bytecode typed-match execution.

The focused regression executes the complete application twice in both
tree-walker and bytecode modes. It validates the public generic
`Option`/`Result` API and the imported-interface branch without adding a
named type, benchmark, or stdlib special case.

## Fresh measurement

Fresh five-process references were reused because their source fingerprints
remain current. The repaired verifier-backed comparison is
`2026-07-15-option-result-scorecard-coverage-postfix.json`.

| Mode | Able mean | Go ratio | Python ratio | Ruby ratio | Result |
| --- | ---: | ---: | ---: | ---: | --- |
| compiled | 0.1960 s | 65.33x | 12.89x | 4.79x | verified, target miss |
| bytecode | 3.3880 s | 1129.33x | 222.89x | 82.84x | verified, target miss |

These means are workstation samples across independent processes. They close
the scorecard correctness gap; they do not constitute a before/after
performance claim for the semantic repair.

## Promotion

`external-scoreboard-current.{json,md}` now has 33 applications, 66
compiled/bytecode rows, and 25 explicit input reports. The current scorecard
has four of 32 rankable compiled rows and three of 24 rankable bytecode rows
meeting their respective 95% targets. All incomplete rows remain unranked.

`bench_external_scoreboard --check` validates the promoted report exactly.

## Decision

Retain the semantic repair and the complete scorecard. Do not select an
Option/Result performance candidate from this single application: the current
record does not show a concrete shared leaf across three unlike verified
applications.
