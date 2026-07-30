# Transaction Ledger Audit application gate

## Decision

Retain the new portable application, its references, selection coverage, and
performance evidence. Retain no compiler, generated-runtime, interpreter, VM,
stdlib, language, dependency, or WASM change.

The compiled owner refresh did not find one concrete general mechanism that
is material in all three unlike applications. `runtime.mallocgc` is the only
broad common parent, but the allocating Able descendants differ. That is not
enough evidence to authorize a production rule.

## Application contract

Transaction Ledger Audit processes 32 delimited records for 512 rounds. It
covers valid and malformed transactions, signed integer parsing, nominal
records, `Result` validation, native `Array` region counters, a generic
string-keyed map, and aggregation.

The canonical and external Able sources are byte-identical at
`efed16e88aed7c6b9d4b7aef115eb1274824c5b725bbdbf845993218c9273738`.
Able, Go, Python, and Ruby all produce:

```text
16384:10240:2048:2048:2048:3584,3584,3072:785408:836090112:991490
```

The output SHA-256 is
`aa5a4fe7f85ce13998797ef506647a93f16e0ee747613683268fd801d609c812`.
Strict compiled construction passed with `--no-fallbacks`; its dependency
graph and binary symbols contain no interpreter package.

## Repeated baseline

All measurements used CPU 15, Go 1.26.5 where applicable, five successful
Able processes, five successful reference processes per comparison, and the
public verifier.

| Mode | Able mean | Reference mean | Able/reference |
| --- | ---: | ---: | ---: |
| compiled | 0.0500 s | Go 0.0073 s | 6.8493× |
| bytecode | 4.5900 s | Python 0.0344 s | 133.4302× |
| bytecode | 4.5900 s | Ruby 0.0914 s | 50.2188× |

The new application is therefore a target miss in both modes. Its evidence
expands the frontier; it does not establish a candidate by itself.

## Three-application compiled owner refresh

Each application was rebuilt strictly with Go 1.26.5 and verified without an
interpreter dependency. CPU profiles came from independent verified
processes. Allocation counts are main-phase deltas.

| Application | CPU profiles / samples | Main allocated bytes | Main allocations | Material Able owners |
| --- | ---: | ---: | ---: | --- |
| Transaction Ledger Audit | 20 / 410 ms | 5,784,248 | 115,269 | split/parsing, `ToString`, `ToInt`, `ToDynamicI64`, key and error construction |
| Inventory Reconciliation | 5 / 540 ms | 17,037,072 | 553,059 | generic map hash/equality/lookup, `ToDynamicI64`, nullable integer recovery |
| Sensor Calibration | 30 / 220 ms | 3,333,264 | 53,315 | split/parsing, `ToInt`, `Result`/error union and nominal construction |

Map storage is material only in Transaction Ledger Audit and Inventory
Reconciliation. Parsing and integer conversion are material only in
Transaction Ledger Audit and Sensor Calibration. Nominal error construction
is likewise absent from Inventory Reconciliation's leading owners. The only
common ancestor is Go allocation/GC machinery, not an exact compiler or
generated-runtime leaf.

Consequently:

- `compiled-text-map` remains `closed-no-shared-leaf`;
- `bytecode-text-map` remains `closed-rejected-candidate`: the new row repeats
  an already-closed text/map/Result surface and provides no new compiled
  mechanism that would justify reopening a VM route;
- no named-container or non-primitive nominal special case is admissible;
- no broad execution-context or compiled/interpreted-boundary change is
  warranted.

Readable CPU, cumulative CPU, allocation-object, allocation-space, and phase
records are retained in
`2026-07-29-transaction-ledger-audit-compiled-owner-profiles/`.

## Cohort result

The portable corpus now contains 64 applications and 128 selected
compiled/bytecode rows. The selection identity is
`96d86cdad06f52286e4423ecad7d47ffaf75968fdc51a7e8010f65369291549e`.
The dated scorecard refresh contains 35 retained source reports and 36
reference reports, with five successful Able/reference samples for every
selected row.

## Next recommendation

Add a portable user-defined generic nominal storage workload before reopening
typed nominal storage architecture.

Why: the generic map boundary remains important in two programs, but a
`HashMap`-specific compiler rule is forbidden. Current evidence cannot tell
whether the cost is a general semantic-encoding boundary shared by
user-defined nominals or merely this stdlib container's implementation.

What it entails: build source-equivalent Able, Go, Python, and Ruby programs
that exercise a user-defined generic mutable nominal through identity,
aliasing, interface, and iteration boundaries. Verify every engine, take at
least five balanced measurements, run the evidence selector, and profile only
the closure groups it invalidates.

Why it is important: this is the shortest evidence path to a safe general
native-Go nominal lowering rule. It directly tests whether compiled Able can
avoid dynamic materialization without sacrificing language identity or
crossing the compiled/interpreted boundary.
