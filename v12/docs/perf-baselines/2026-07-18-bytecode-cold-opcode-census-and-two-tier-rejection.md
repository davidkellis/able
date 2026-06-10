# Bytecode Cold-Opcode Census and Two-Tier Rejection

Date: 2026-07-18

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, workload, fixture, or
language change from this tranche.

A full selected-bytecode main-phase opcode census completed for 25 of 27
applications and recorded 954,883,252 instructions. K-Nucleotide and Regex Set
Audit exceeded the 55-second cap under the deliberately expensive stats
observer, so their missing snapshots are reported as incomplete and never used
as evidence that an opcode is absent.

The complete snapshots exposed a coherent cold declaration/module family. A
single generic two-tier prototype moved function, struct, union, alias,
methods, interface, implementation, extern, import, and dynamic-import cases
to an out-of-line secondary handler. It preserved every opcode's semantics and
reduced `runResumable` from 35,909 to 31,789 bytes (11.47%), with a 1,061-byte
cold handler. The repeated broad gate nevertheless moved unlike applications
in both directions and consistently regressed multiple guards. The prototype
was fully reverted.

No WASM work was performed.

## Census contract

- selected manifest: all 27 reviewed bytecode applications;
- manifest SHA-256:
  `19c0b7c5c9a41226cfff851b99ffeca46317ff7f8ab608378deb2c66153c06fe`;
- exact observer/runtime binary SHA-256:
  `b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`;
- CPU 0, `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`;
- canonical external `able-stdlib` and catalog source/input/executor contracts;
- main-only stats reset immediately before `main()`;
- 55-second process cap and public Ruby output verification.

Twenty-five applications completed and verified with stats. K-Nucleotide and
Regex Set Audit timed out at 55.01 and 55.00 seconds respectively because the
per-instruction observer substantially slows them; neither produced a usable
stats file. Regex Stream Audit narrowly completed in 49.30 seconds. The exact
run statuses and hashes are retained in
`2026-07-18-bytecode-cold-opcode-census-runs.tsv`.

Across the 25 complete applications, 57 of 143 opcodes had zero executions.
Another eight nonzero opcodes remained below 0.01% in every complete
application: `AssignPattern`, `BinaryI32Add`, `BinaryIntSubSlotConst`,
`CallSelf`, `ConstI32`, `IteratorLiteral`, `StoreSlotI32`, and
`StoreSlotIntMulConstAddFromSlot`. This is a suite-wide main-phase
classification, not a claim that these operations are unreachable in arbitrary
programs.

The aggregate is dominated by broadly used cases: `LoadSlot` contributes
198,815,171 executions across all 25 applications, `Pop` 75,429,212, `Jump`
68,793,684, `StoreSlotBinaryIntSlotConst` 55,733,936, and
`JumpIfBinaryCompareFalse` 48,846,996. The complete per-opcode totals,
application counts, and maximum per-application shares are retained in
`2026-07-18-bytecode-cold-opcode-census-summary.tsv`.

## Prototype structure

Only the coherent declaration/module cases were moved. All hot arithmetic,
stack, branch, call, return, collection, matching, and concurrency cases stayed
in the primary switch. The default arm invoked one secondary switch, and the
existing definition/import implementations remained the semantic authority.
Dynamic definitions and imports therefore still worked; they merely paid the
out-of-line call on their already-cold path.

Focused definition, import, dynamic-import, return, and call-name tests passed.
The candidate kept the primary source file below the 1,000-line limit at 935
lines. Symbol sizes from separately preserved binaries were:

| Symbol | Baseline | Candidate |
| --- | ---: | ---: |
| `runResumable` | 35,909 B | 31,789 B |
| `execColdOpcode` | absent | 1,061 B |

The candidate binary SHA-256 was
`9b2e4ac8b2a0eae7e4d598150fa777727dea1c8f8acd1170b2999a503298792e`.

## Repeated causal gate

Nine unlike selected applications received five order-balanced baseline/
candidate pairs. All 90 processes completed, passed their public verifiers,
and retained every workstation outlier. Stats were disabled. The exact ledger
is retained in `2026-07-18-bytecode-cold-opcode-two-tier-ab.tsv`.

| Application | Samples/variant | Baseline mean | Candidate mean | Change | Slower pairs |
| --- | ---: | ---: | ---: | ---: | ---: |
| Word Frequency | 5 | 1.449 s | 1.483 s | +2.32% | 4/5 |
| Array Slice Window | 5 | 0.622 s | 0.651 s | +4.67% | 5/5 |
| Mandelbrot | 5 | 6.425 s | 6.396 s | -0.45% | 4/5 |
| Option/Result Config | 5 | 0.858 s | 0.850 s | -0.88% | 3/5 |
| Mutex Ledger | 5 | 0.347 s | 0.339 s | -2.51% | 1/5 |
| Unicode Scalar Pipeline | 5 | 3.407 s | 3.492 s | +2.49% | 5/5 |
| Reverse Complement | 5 | 6.808 s | 6.896 s | +1.29% | 2/5 |
| Rational Series | 5 | 4.146 s | 4.145 s | -0.04% | 2/5 |
| JSON | 5 | 0.858 s | 0.989 s | +15.26% | 2/5 |

The JSON mean includes the 1.505-second candidate outlier. Even without using
that volatile row, Array Slice Window and Unicode were slower in every pair,
and Word Frequency was slower in four of five. The design therefore fails the
project's cross-family bar. A smaller dispatcher is not itself a product win
when linked-code layout still exchanges performance among real programs.

The full 27-application scorecard was not rerun because the causal admission
gate had already rejected the candidate; doing so could not make the rejected
code eligible.

## Restoration and verification

- The secondary handler and extracted definition helpers are fully removed.
- `bytecode_vm_run.go` is restored to its exact pre-candidate SHA-256:
  `aac8114450eda1c49a743036103503a5d2d1c84fa3258e6f127425329a1a3def`.
- Focused definition/import/return/call tests pass on the restored source.
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passes in 28.764 seconds under the memory/GC/CPU guardrails.
- A clean restored CLI exactly reproduces the preserved baseline SHA-256
  `b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`.
- The selection-manifest check passes with 35 compiled and 27 bytecode rows.
- The external canonical stdlib was not changed.
- Raw binaries, stats JSON, output captures, and runners are cleanup-only;
  only the compact census and A/B ledgers are retained.

## Next recommendation

Run a bounded cross-suite Go PGO feasibility gate for the bytecode CLI, using
one diverse feature-complete subset only to train the profile and a disjoint
set of selected applications to decide whether it generalizes.

Why: source-level changes inside or immediately around the large Go dispatcher
have now repeatedly shifted unrelated programs in opposite directions, even
when they remove 11.5% of the function. PGO is the next generic way to let the
host compiler place and inline the existing semantic paths using aggregate
language-feature evidence, without encoding an Able benchmark, named
container, or opcode fusion in runtime source.

What it entails: merge CPU profiles from unlike text, numeric, collection,
matching/error, concurrency, and iterator training applications; build one
otherwise identical CLI with Go's `-pgo` input; and compare it first on
disjoint verifier-backed applications with repeated order-balanced processes.
Only if unseen validation is neutral-to-better should it advance to the full
27-application bytecode scorecard and build/tooling integration. Reject it if
any family consistently regresses, and continue to defer WASM.
