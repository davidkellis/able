# Bytecode execution-lane reconciliation — 2026-07-15

## Decision

Keep the promoted 32-application bytecode scorecard unchanged. Its ordinary
CLI rows intentionally include loading, lowering, typechecking, bootstrap,
and one `main()` execution, just as the command a user invokes does. They
remain the current full-process status and regression baseline.

Do not interpret a warmed `bytecode-runtime` measurement as that same metric.
The runtime benchmark loads and validates a program once, then calls warmed
`main()` repeatedly; it intentionally skips duplicate typechecking. This
explains why focused VM measurements can be far below the full-process cap
without contradicting the scorecard.

## Trusted execution mode

`able run --skip-typecheck <target-or-file>` now provides an explicit mode for
an already validated source graph. It skips the program-wide typechecker and
therefore emits no typecheck diagnostics during that run. Ordinary `able run`
is unchanged and continues to typecheck by default. `able check` rejects the
flag, so validation remains an explicit, separately observable operation.

The flag is a measurement and trusted-execution boundary, not a language
semantic optimization. It must not be used to execute unvalidated source, and
it is not enabled by the normal CLI, fixtures, or current scorecard.

`bench_perf` and `bench_compare_external` now expose this as the separately
named `bytecode-prechecked` mode. They run `able check` once before timing and
then run each bytecode process with the explicit trusted flag. The external
scoreboard accepts only `compiled` and ordinary `bytecode` source rows, so it
rejects this mode rather than allowing it to replace the full-process baseline.

## Verification

The focused CLI tests prove all three contracts:

- the default run stops before `main()` when another function has a type error;
- a trusted bytecode run can execute an otherwise unreachable, already-known
  type error without pretending that it performed validation; and
- `able check --skip-typecheck` fails with a clear diagnostic.

Shell syntax checks pass for the benchmark scripts. A local fixture and one
external JSON application both completed the precheck plus verifier-backed
`bytecode-prechecked` execution path. Those functional launches were
intentionally unpinned and are not timing evidence or scorecard input.

The initial CPU ranking found no eligible core. Its least-busy candidate, CPU
15, peaked at 6.93% busy; its immediate formal preflight then measured 9.90%,
9.00%, and 6.00% busy against the 5% limit. A later ranking found CPU 8 below
the gate (0.99% peak busy) and its immediate preflight passed at 0.00%, 0.00%,
and 1.98% busy, enabling the bounded paired study below.

## Next measurement gate

CPU 8 ran three verifier-backed processes per lane with the standard 45
second/1-GiB/single-Go-process guard:

| Application | Full process | Prechecked execution | Result |
| --- | ---: | ---: | --- |
| Sudoku | timed out (3/3) | timed out (3/3) | no boundary effect |
| Word Frequency | 1.33 s | 1.32 s | 0.01 s (clock-resolution noise) |
| Future Pipeline | 0.41 s | 0.41 s | no difference |

The verifier hashes match between the two lanes for both completing
applications. Sudoku has the same timeout status in both lanes. The exact
machine-readable and Markdown records are
`2026-07-15-bytecode-full-vs-prechecked-paired.json` and
`2026-07-15-bytecode-full-vs-prechecked-paired.md`.

## Decision

Reject a loader, typechecker, or source-graph-cache performance investigation.
This three-application decomposition finds no material execution-boundary
phase, so it cannot justify a change to the default diagnostic path or a
trusted-mode-specific cache. The full-process scorecard remains the product
guard, and the prechecked lane remains a non-scorecard measurement aid.

The next performance investigation requires a newly admitted concrete
non-nominal VM or compiler leaf repeated in three unlike verifier-backed
applications. Neither result authorizes a benchmark-specific skip, a changed
default diagnostic policy, or a scorecard redefinition.
