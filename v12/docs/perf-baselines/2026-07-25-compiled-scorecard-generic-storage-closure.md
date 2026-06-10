# Current compiled scorecard and generic-storage closure

Date: 2026-07-25

## Decision

Retain no compiler, runtime, or stdlib code from this tranche.

The refreshed compiled scorecard found a material shared owner in
K-Nucleotide, Inventory Reconciliation, and Word Frequency:

```text
native specialized key/value carriers
  -> generated nominal method
  -> runtime.Value key/value encoding
  -> runtime-backed generic HashMap storage
```

The narrow optimization would be a compiler or runtime rule for the named
stdlib `HashMap` type. That is explicitly forbidden. The general semantic
parent is typed storage for generic nominal values, including a representation
and identity design that applies to unlike user-defined and stdlib nominals.
That is an architectural storage change, not an admissible local lowering rule
for this tranche. The candidate therefore stopped before implementation,
semantic guards, or A/B measurement.

Machine-readable evidence is in
`2026-07-25-compiled-scorecard-generic-storage-closure.json`.

## Current compiled scorecard

Fresh Go references and strict compiled Able rows used five measured,
verifier-backed runs, the fixed CPU pool `5,10,15,11`, the catalog's logical
CPU budget for each application, `GOMEMLIMIT=1GiB`, and `GOGC=50`.

- All 61 Go references completed: 305 of 305 outputs verified.
- The latest Able row for each application gives 53 complete comparisons.
- Seven complete rows meet the 95%-of-Go target ratio of at most `1.052632x`:
  Binary Trees, Quicksort, I Before E, Base64, JSON, Monte Carlo Pi, and
  Pidigits.
- Forty-six complete rows miss the target.
- Eight rows remain incomplete. Six regex/policy-heavy rows and their
  dependents did not finish preparation even with a 90-second repair budget;
  Mutex Ledger and Concurrent State Machines produced verified output but
  intermittently lost timing metrics.

The material-excess ranking changed substantially from the stale scoreboard:

| Application | Able mean | Go mean | Ratio | Excess | Mean GC |
| --- | ---: | ---: | ---: | ---: | ---: |
| Tapelang Alphabet | 3.8180 s | 2.0027 s | 1.906x | 1.8153 s | 0.0 |
| K-Nucleotide | 1.6460 s | 0.0556 s | 29.604x | 1.5904 s | 285.4 |
| Sudoku Masks | 1.8120 s | 0.6968 s | 2.600x | 1.1152 s | 133.0 |
| Mutex Work Queue | 0.9880 s | 0.0047 s | 210.213x | 0.9833 s | 13.6 |
| Fib | 3.6340 s | 3.3510 s | 1.084x | 0.2830 s | 0.0 |
| Matrix Multiply | 1.1820 s | 0.9801 s | 1.206x | 0.2019 s | 8.0 |
| Binary Event Log | 0.1940 s | 0.0085 s | 22.824x | 0.1855 s | 60.0 |
| Inventory Reconciliation | 0.1520 s | 0.0138 s | 11.014x | 0.1382 s | 14.0 |

Binary Trees now averages `1.010x` Go. Several current concurrency rows that
were historically large now finish in roughly 30-50 milliseconds. This
confirms that the previous aggregate was too stale to select new work.

The main and repair reports remain dated evidence rather than replacing
`external-scoreboard-current`. They cover compiled Able versus Go only and
have eight incomplete rows, while the canonical aggregate also governs
bytecode comparisons.

## Profile admission

The first profile cohort selected K-Nucleotide, Binary Event Log, and Inventory
Reconciliation by material allocation and wall-time excess. Binary was
disqualified from the shared candidate: ten merged CPU runs attribute
`0.65 s` cumulative time to `EventRecord_to_seen`, while the map equality path
accounts for only `0.05 s`. Its dominant owner is the already-studied nominal
record conversion/writeback boundary.

Word Frequency replaced Binary because it exercises the same generic
collection through different primitive carriers and an interface adapter.
Each final row was rebuilt strictly, publicly verified, and confirmed to have
a 96-package dependency graph with no `pkg/interpreter` dependency.

| Application | Native specialization | Main allocated bytes | Main allocations | Main GC | CPU evidence |
| --- | --- | ---: | ---: | ---: | --- |
| K-Nucleotide | `HashMap_u64_i32` | 614,264,498.67 | 16,232,599 | 335.00 | `raw_get` 6.51 s; `raw_set` 8.14 s; primitive key equality 3.00 s |
| Inventory Reconciliation | `HashMap_i64_i64` | 17,037,077.33 | 553,059 | 15.33 | map adapter/get 0.57 s; `raw_set` 0.36 s; key equality 0.21 s |
| Word Frequency | `HashMap_String_i32` | 3,347,368.00 | 60,186 | 3.67 | map adapter/get 0.13 s; `raw_set` 0.11 s; `ToString` 0.08 s |

Allocation values are three-run exact means. CPU profiles merge ten verified
runs for K and Inventory and thirty for the shorter Word process.

Generated source proves the same semantic transition in all three:

- K accepts `uint64`/`int32`, calls `bridge.ToUint` and `bridge.ToInt`, and
  passes `[]runtime.Value` to `__able_hash_map_get_impl` or
  `__able_hash_map_set_impl`.
- Inventory accepts `int64`/`int64`, calls `bridge.ToDynamicI64`, and reaches
  the same runtime helpers through its generated `Map` adapter.
- Word accepts `string`/`int32`, calls `bridge.ToString` and `bridge.ToInt`,
  and reaches the same runtime helpers through a different generated `Map`
  adapter.

The native carriers are therefore being preserved through compiled calls and
generic specialization, then deliberately materialized at the runtime-backed
storage boundary. Removing only these conversions would name `HashMap`;
removing the boundary generally requires shared typed generic-nominal storage,
including fallback materialization, identity, mutation, aliasing, iteration,
interface, generic, imported, and dynamic behavior.

## Verification and provenance

- The current Go refresh verified 305 processes.
- The main Able report verified 270 processes; the repair report verified 20
  additional processes.
- Twelve exact-allocation and 60 CPU-profile processes verified.
- All final profile dependency graphs contain 96 packages and omit
  `able/interpreter-go/pkg/interpreter`.
- No candidate was implemented because it failed the general-rule admission
  gate; twenty-cohort A/B/Go measurement was therefore not warranted.
- No compiler, runtime, interpreter, VM, stdlib, language, dependency, or WASM
  code was changed.
- Completed raw Able build/profile trees were removed from RAM-backed `/tmp`;
  only 3.8 MiB of small `able-*` files remained afterward. Future large
  workspaces are required to use disk-backed `/var/tmp`.
- Retained compiler SHA-256:
  `8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`.
- Dated scorecard evidence:
  `2026-07-25-current-compiled-go-reference.{json,md}`,
  `2026-07-25-current-compiled-scorecard.{json,md}`, and
  `2026-07-25-current-compiled-scorecard-repair.{json,md}`.

## Next

Profile Tapelang Alphabet, Matrix Multiply, and Fib as a low-allocation
compiled compute/control cohort, then admit at most one shared primitive or
control-flow lowering owner.

This is next because these unlike applications have material absolute excess
without being dominated by the forbidden named-container storage route or by
mandatory scheduler services. Tapelang and Fib report no main GC in the
scorecard, while Matrix reports only eight collections, so their remaining
time is more likely to expose native arithmetic, loop, checked-operation, or
compiled-call overhead.

The work entails fresh strict generated-source inspection, exact allocation
profiles, repeated CPU profiles, and attribution of checked arithmetic,
control propagation, loops, recursion, and native Array access. A candidate
advances only if the same semantic owner is material in all three and can be
corrected by a general primitive/native-carrier or control-flow rule. If
admitted, add focused overflow, evaluation-order, error, recursion, loop, and
dynamic-boundary guards before twenty order-balanced baseline/candidate/Go
cohorts.

This is important because these programs test the central performance claim
directly: when Able primitives and static Arrays already lower to native Go
carriers and no interpreter boundary is present, generated code should
converge on equivalent Go performance. Do not begin WASM work.
