# Truthiness/cast cross-family architecture ownership audit

Date: 2026-07-21

## Decision

**No cross-family implementation candidate is admitted.**

The current 85-row frontier contains 46 compiled and 39 bytecode rows. Eight
meet target, 77 miss, and the selected rows carry 143.927053 seconds of modeled
excess. The 18 refreshed family closures retain 1,710 verified timing
processes. Reconciliation changes no frontier disposition and retains no
compiler, VM, runtime, stdlib, benchmark, reference, language, or WASM change.

## What changed since the prior census

The prior census covered 75 rows before shared interpreter semantics became an
explicit closure scope. The corrected production semantics make qualified
`able.core.interfaces.Error` values falsy and make failed explicit runtime
casts catchable. That change invalidated every old closure even where a
benchmark never reached the changed path.

Nine staged refreshes now cover every family closure. Exact evidence finds:

- Compiled Option/Result reaches the cast bridge, Policy Record Dispatch
  reaches truthiness, and seven concurrency applications reach truthiness,
  but admitted normal-build profiles sample zero CPU in all generated/shared
  truthiness and cast helpers.
- Bytecode Rational Series and Wide Integer Records reach successful casts at
  high counts, but the catchable wrapper is CPU-flat zero; K-Nucleotide has
  only 28 successful calls per process. Every refreshed bytecode census records
  zero entries into the corrected non-primitive Error fallback.
- The admitted concurrency profiles reproduce the already-known
  `currentGID`/`runtime.Stack` owner at 74.07%-96.82% cumulative. Its general
  execution-context replacement already failed unlike guards.

Reach therefore broadened, but material ownership did not. A path is not a
performance candidate merely because it executes; it must own sampled CPU in
at least three unlike applications.

## Architecture partition

| Boundary | Mode | Current result | Disposition |
| --- | --- | --- | --- |
| Corrected truthiness/cast paths | Cross-mode | Reach without a three-family sampled leaf | `closed-no-shared-material-leaf` |
| Generated runtime/boxed fallback | Compiled | Direct generated bodies dominate; bridges sample zero | `closed-no-shared-material-leaf` |
| Allocation and nominal results | Compiled | Text, wide, regex, and concurrency lifetimes differ | `closed-split-semantic-owners` |
| Runtime context publication | Compiled | `currentGID` is material but its general replacement failed guards | `closed-rejected-candidate` |
| Primary opcode dispatch | Bytecode | Shared parent; executable general mechanisms already rejected | `closed-rejected-mechanisms` |
| Scalar encoding/transport | Bytecode | Recurrent exact leaves reverse sign under carrier variants | `closed-rejected-candidate` |
| Call/return continuation | Bytecode | Descendants divide by call and coercion route | `closed-rejected-or-no-shared-child` |
| Nominal construction | Bytecode | Hot definitions/lifetimes do not cross three unlike apps | `closed-insufficient-lifetime-breadth` |
| Map/allocation parents | Cross-mode | Aggregate Go symbols split under caller attribution | `closed-no-shared-leaf` |

The machine-readable census is
`2026-07-21-truthiness-cast-cross-family-architecture-ownership-census.json`.
It freezes the current frontier totals, admission rule, boundary partition,
and exact invalidation condition.

## Measurement policy

No timing was added in this reconciliation because every representative
family already has post-fix retained repeated-process evidence. Re-running
unchanged binaries would add workstation noise without changing reach or
candidate admission. All successful samples remain in their arithmetic means;
the family reports document any second cohorts used for volatility.

The architecture rule remains deliberately strict: one concrete, non-nominal
mechanism must be CPU-material in at least three unlike verifier-backed
applications and must not already have failed broad guards. Aggregate ancestry
such as generated bodies, map access, malloc, or GC cannot satisfy that rule.

## Next recommendation

Refresh `compiled-sudoku-quotient`, the only remaining invalidated closure.

Why: it is the last unchecked consequence of the truthiness/cast semantic
change. Closing it will make the 21-closure ledger current and distinguish a
real new performance lead from stale evidence. It is intentionally narrow and
cannot authorize a compiler optimization on its own.

What it entails: verify post-fix compiled/Go timing for Sudoku Masks with
repeated retained samples, run exact quotient and corrected-path reach only if
the current artifacts do not already answer it, and retain no candidate unless
the same concrete quotient mechanism is independently material in at least two
additional unlike applications. Advance only that closure. Do not begin WASM
work.
