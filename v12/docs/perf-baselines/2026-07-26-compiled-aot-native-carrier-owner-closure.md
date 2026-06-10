# Compiled AOT native-carrier owner closure

Date: 2026-07-26

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, reference, language, dependency, or WASM change
from this tranche.

Retain one test-only extern-plugin lifecycle correction discovered by the full
handoff: the invalid-artifact regression now builds its initially valid plugin
in a short-lived helper process before corrupting the file on disk. The old
test loaded the plugin and then truncated its own memory-mapped artifact,
which can cause a legitimate Linux `SIGBUS`. Production plugin loading and
rebuild behavior is unchanged.

Fresh diagnostics-off generated-code CPU and allocation evidence for N-Body,
Inventory Reconciliation, and Unicode Scalar Pipeline contains no exact
representation owner material in all three applications. The selected work
already remains in native generated Go:

- N-Body uses `*__able_array_f64` and `float64` through its hot call graph.
- Inventory uses native `int64` at application call sites, then intentionally
  enters the runtime-backed generic Map semantic boundary.
- Unicode uses native `string`, `rune`, `int32`, and `int64`, with shared
  nominal union/interface encoding at iterator results.

The three-way admission gate therefore fails before a production prototype or
five-pair A/B/Go cohort is justified.

The compact machine-readable companion is
`2026-07-26-compiled-aot-native-carrier-owner-closure.json`.

## Compiler-batch audit

The preceding full handoff reported short-mode compiler batches 19, 28, and
29 at 185.073, 75.304, and 93.208 seconds. Two fresh shard repetitions
produced:

| Batch | Handoff | Repeat 1 | Repeat 2 | Three-sample mean |
| --- | ---: | ---: | ---: | ---: |
| 19 | 185.073s | 183.610s | 187.110s | 185.264s |
| 28 | 75.304s | 76.300s | 77.810s | 76.471s |
| 29 | 93.208s | 89.600s | 87.900s | 90.236s |

Verbose per-test timing found no individual test above one minute. Batch 19's
largest guard was
`TestCompilerTypedArrayDefaultMethodsKeepConcreteReceivers` at 46.23 seconds;
its other canonical-stdlib guards accumulate within the same 25-test shard.
Batch 28's largest guard was the canonical nullable Spec lowering test at
23.33 seconds. Batch 29's largest was the imported Matcher struct-converter
guard at 22.79 seconds.

This is stable shard composition, not a reproduced compiler regression. No
compiler or test-runner change was made.

## Frozen application contract

All three applications were emitted with `-no-fallbacks` by one current
compiler binary:

```text
28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab
```

That is byte-identical to the compiler artifact used by the current strict
scorecard. Every smoke, profile, counter, and telemetry execution passed its
public Ruby verifier. Each final graph contains 96 packages and zero matches
for `able/interpreter-go/pkg/interpreter`.

The current five-run scorecard selected three unlike target misses:

| Application | Surface | Able mean | Go mean | Ratio |
| --- | --- | ---: | ---: | ---: |
| N-Body | static `Array f64`, numeric direct calls | 0.0820s | 0.0354s | 2.3164x |
| Inventory Reconciliation | generic Map, `i64` calls/stores | 0.1640s | 0.0082s | 20.0000x |
| Unicode Scalar Pipeline | String/char/iterator/union | 0.1120s | 0.0101s | 11.0891x |

All build and profile artifacts lived under disk-backed `/var/tmp`.

## Diagnostics-off CPU profiles

Twenty independently launched main-only profiles were publicly verified and
merged per Able application. Equivalent Go references were copied
byte-for-byte into disposable modules, with only `main` renamed behind a
profile wrapper; twenty independently launched one-workload profiles were
also verified and merged.

| Application | Able samples | Dominant Able owners | Go samples | Dominant Go owners |
| --- | ---: | --- | ---: | --- |
| N-Body | 1.11s | generated `advance` 87.39% flat; compiled `sqrt` 9.91% | 0.59s | native `advance` 93.22% flat |
| Inventory | 1.82s | generic map equality 18.13%; hash/search; `ToDynamicI64`; allocation | 0.11s | native Go map access/assignment |
| Unicode | 1.60s | scalar checksum; UTF-8 iterator; allocation; checked multiply/divmod | 0.14s | native Go rune loop |

N-Body contains no material `runtime.Value`, interface conversion, boxing, or
allocation owner. Its generated hot signature is
`(*__able_array_f64, ..., float64)`, and its element loads and stores use
`[]float64`.

Inventory's material conversion is below the explicit generic Map boundary:
`bridge.ToDynamicI64` feeds shared runtime map hashing/equality and
`__able_nullable_i64_from_value` recovers the nullable Map result. The
application's arithmetic, direct call arguments, and stores remain `int64`.

Unicode's material representation is different. `utf8_decode` constructs the
shared nominal `Utf8DecodeResult`, and iterator `next` returns the shared
`IteratorEnd | rune` union. Its primitive character and checksum work remains
`rune`, `int32`, and `int64`.

No exact CPU leaf, static return conversion, primitive store conversion,
interface adapter, or boxing operation is material in all three.

## Allocation evidence

Three independent lightweight main-phase counters per application were
stable:

| Application | Able bytes | Able allocations | Able GC | Go bytes | Go allocations |
| --- | ---: | ---: | ---: | ---: | ---: |
| N-Body | 1,096 | 41 | 0 | 376 | 6 |
| Inventory | 17,038,155 mean | 553,063 | 5 | 296,000 | 41 |
| Unicode | 24,808,379 mean | 2,629,753 | 9 | 197,032 | 10 |

Three exact allocation-profile processes per Able and Go application also
verified. Main-boundary profile serialization is allocation-heavy, so the
lightweight counters are authoritative for totals. Exact Able attribution
still separates the application shapes:

- N-Body has no material application allocation leaf.
- Inventory records 270,336 flat `ToDynamicI64` objects and 135,169 nullable
  result-recovery objects in the representative exact profile.
- Unicode records 1,016,741 flat UTF-8 decode objects and 293,986 iterator
  union wrappers in the representative exact profile.

The allocations do not share one carrier, owner, or semantic reason.

## Report-only boundary reconciliation

Separate telemetry builds used the already-retained opt-in typed-boundary
observer; diagnostics remained disabled in every governing CPU and allocation
run.

| Application | Material observed categories |
| --- | --- |
| N-Body | `any_to_runtime=2`, both at final output |
| Inventory | `control_from_error=811,010` |
| Unicode | `control_from_error=1,769,472` |

All other typed-boundary categories were zero. The shared control category
reaches only Inventory and Unicode, and its concrete descendants differ:
fallible generic Map semantics versus fallible character/iterator semantics.
N-Body's two output conversions are neither hot nor avoidable.

## Admission decision

| Candidate | Reach | Disposition |
| --- | --- | --- |
| preserve primitive arguments/results through static calls | already true in all three | no missing lowering rule |
| remove `runtime.Value` at generic Map operations | Inventory only | one-family semantic boundary; previously closed |
| unbox iterator/UTF-8 nominal union results | Unicode only | one-family nominal representation; no shared reach |
| remove successful control checks | Inventory and Unicode only | concrete operations can fail; N-Body has no matching boundary |
| remove Array checks | N-Body only | prior relational proof found material reach in only N-Body |

No general primitive/static rule repeats in all selected applications.
Consequently no production experiment or balanced A/B cohort was entered.

## Verification

- The invalid-artifact extern-plugin regression passes three consecutive
  focused runs in 0.47-0.53 seconds, and the related extern-plugin group
  passes.
- The complete `./run_all_tests.sh` handoff passes every coverage, scorecard,
  selection, and threshold contract; all non-compiler packages; all 32
  compiler batches; and the final bytecode fixture corpus in 86.579 seconds.
  The audited aggregate batches took 187.613, 75.397, and 88.214 seconds in
  this final run.
- Three strict builds and 162 application processes passed public
  verification: three smoke processes, 60 Able CPU-profile processes, 60 Go
  CPU-profile processes, nine Able counter processes, nine Go counter
  processes, nine Able allocation-profile processes, nine Go
  allocation-profile processes, and three telemetry processes.
- All three final dependency graphs remain interpreter-free.
- The existing typed-boundary observer was not changed.
- No stdlib, runtime semantics, interpreter, language, dependency, or WASM
  work was needed.

## Next recommendation

Run a full strict-corpus, report-only typed-boundary callsite census.

Why: this three-program cohort disproves a shared owner here, but it cannot
show whether another set of unlike target misses shares one concrete residual
primitive boundary. The existing category totals are too broad to select a
production change by themselves.

What it entails: extend or compose the opt-in observer so each native/runtime
conversion is attributed to generated function, source site, carrier,
consumer, and semantic reason; execute all 61 strict applications under their
public verifiers; rank only independently recurring callsite shapes; then
refresh diagnostics-off profiles for any exact shape material in at least
three unlike applications. A production prototype still requires focused
guards and five-or-more balanced verifier-backed A/B/Go pairs.

Why it is important: it tests the native-carrier hypothesis across the entire
compiled corpus while preserving Able errors, nominal encoding, and explicit
dynamic boundaries. It also provides a defensible stopping condition if
strict applications already keep primitives native everywhere except required
semantic boundaries. Do not begin WASM work.
