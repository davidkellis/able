# Compiled runtime-value boundary census closure

Date: 2026-07-25

## Decision

Retain no compiler, generated-runtime, bridge, runtime, canonical-stdlib,
application, benchmark, interpreter, bytecode, language, dependency, or WASM
performance change.

Seven unlike current strict target misses were generated both with the
existing typed-boundary diagnostic and as normal telemetry-free binaries.
Every generated dependency graph contains 96 packages and omits
`able/interpreter-go/pkg/interpreter`. Five repeated diagnostic executions per
application produced stable main-only counts and passed the public verifier.

One exact allocation leaf is material in three unlike applications:
`runtime.NewStructInstancePositionalSized`. It is the single escaping
`runtime.StructInstanceValue` required when a native nominal value crosses an
explicit runtime boundary. Option/Result Config performs 6,432 instrumented
struct encodes per main, Sensor Calibration 1,280, and Manifest Normalization
7,168. The current positional representation allocates exactly one 128-byte
instance for each small encoded struct; the retained July 25 positional
encoding already removed the avoidable field map and uses inline field slots.

There is no admissible local candidate below that leaf. Removing the remaining
instance would require keeping the native nominal carrier through the broad
Result/control/runtime ABI or recovering it through caller-owned nominal
storage. Both architectures are already closed and explicitly excluded from
this tranche. The other large exact owner, repeated construction of AST type
expressions while decoding a generic union, is material only in Option/Result
Config. High `control_from_error` counts again compile to effectively free
successful nil checks and do not form a shared sampled owner.

The machine-readable companion is
`2026-07-25-compiled-runtime-value-boundary-census-closure.json`.

## Frozen protocol

The compiler was built once at SHA-256
`8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`.
All generated and profiling work used disk-backed
`/var/tmp/able-boundary-census-20260725`.

For every application:

1. emit a diagnostic runnable with `--no-fallbacks`,
   `--typed-boundary-telemetry`, `--main`, and `--pkg main`;
2. emit a corresponding normal runnable without telemetry so atomics cannot
   contaminate CPU/allocation attribution;
3. prove both generated sources and final dependency graphs omit
   `pkg/interpreter`;
4. run the diagnostic binary five times with the catalog working directory,
   arguments, executor, and public verifier;
5. run 20 independent normal main-only CPU profiles and merge only within the
   application; and
6. run three independent exact main-allocation profiles and retain the
   lightweight phase counter as the allocation-total authority.

Serial applications used CPU 5 and `GOMAXPROCS=1`. Concurrent Scene Tiles used
CPUs `5,10,15,11`, the goroutine executor, and `GOMAXPROCS=4`. All processes
used `GOMEMLIMIT=1GiB`.

Allocation-profile serialization allocates heavily inside `runtime/pprof`.
Those observer frames are excluded from application attribution. Main-phase
`runtime.MemStats` deltas are captured before the end snapshot is serialized,
so they remain the exact total authority.

## Stable main-only boundary reach

Only nonzero categories are shown.

| Application | Exact nonzero typed-boundary counts |
| --- | --- |
| Concurrent Scene Tiles | Array from runtime 1; struct from/to runtime 4/4; error to control 9 |
| Option/Result Config | struct to runtime 6,432; union from runtime 24,384; error to control 55,200 |
| Sensor Calibration | Array from runtime 1; struct to runtime 1,280; union from runtime 1; error to control 25,089 |
| Dependency Plan | none |
| Document Audit | Array from runtime 1; struct from runtime 2; interface from runtime 2; error to control 207 |
| Manifest Normalization | Array from runtime 1; struct from/to runtime 3,072/7,168; union from runtime 1; error to control 16,385 |
| Array Slice Window | none |

All five repetitions of each row were identical. Their stdout SHA-256 values
match the current scorecard:

| Application | Stdout SHA-256 |
| --- | --- |
| Concurrent Scene Tiles | `2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe` |
| Option/Result Config | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| Sensor Calibration | `e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb` |
| Dependency Plan | `96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38` |
| Document Audit | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |
| Manifest Normalization | `2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb` |
| Array Slice Window | `155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e` |

## CPU and allocation intersection

Twenty verified main-only CPU profiles per application produced:

| Application | Merged samples | Boundary attribution |
| --- | ---: | --- |
| Concurrent Scene Tiles | 80 ms | No nominal/control boundary sample; checked arithmetic and application work dominate |
| Option/Result Config | 280 ms | Generic union decode 180 ms cumulative; AST type construction 140 ms cumulative |
| Sensor Calibration | 120 ms | `CalibrationError` encoding 30 ms cumulative; allocation descendants, zero flat converter CPU |
| Dependency Plan | 10 ms | No typed boundary; checked addition is the only sample |
| Document Audit | 20 ms | No typed boundary sample; UTF-8 validation and GC |
| Manifest Normalization | 180 ms | `ManifestRecord`/`ManifestError` encoding 100 ms cumulative; positional instance creation 60 ms cumulative |
| Array Slice Window | 80 ms | No typed boundary; native Array slicing, allocation, and checked addition |

The three exact main-allocation repetitions were extremely stable:

| Application | Mean bytes | Mean objects | Selected exact attribution |
| --- | ---: | ---: | --- |
| Concurrent Scene Tiles | 1,192,512 | 37,058.00 | Typed nominal conversions are single-digit noise |
| Option/Result Config | 9,936,637.33 | 182,896.00 | 823,296 bytes/run in 6,432 positional instances; about 6.24 MB/run in union type-expression construction |
| Sensor Calibration | 3,609,397.33 | 57,535.67 | 163,840 bytes/run in 1,280 positional instances |
| Dependency Plan | 199,072 | 14,410.00 | No typed boundary |
| Document Audit | 373,840 | 1,956.33 | Typed nominal/interface conversions are single-digit noise |
| Manifest Normalization | 3,907,373.33 | 84,933.33 | At least 917,504 bytes/run for the 7,168 instrumented small positional instances; record decode separately allocates 245,760 bytes/run |
| Array Slice Window | 1,441,720 | 24,016.00 | No typed boundary |

Generated caller and stack attribution resolves the material families:

- Option/Result Config:
  `__able_struct_ConfigError_to_seen` and
  `__able_union_int32_or_runtime_ErrorValue_from_value`;
- Sensor Calibration: `__able_struct_CalibrationError_to_seen`; and
- Manifest Normalization:
  `__able_struct_ManifestRecord_to_seen`,
  `__able_struct_ManifestError_to_seen`, and
  `__able_struct_ManifestRecord_from`.

The same runtime instance allocator is below the three encoding callers, but
the allocation is the semantic payload rather than avoidable wrapper work.
The union/type-expression construction does not repeat materially in Sensor or
Manifest, and the inverse record recovery is material only in Manifest within
this cohort.

## Why no candidate advanced

- Dependency Plan and Array Slice Window execute no typed/runtime conversion
  at all while remaining 4.53x and 12.20x slower than their current Go rows.
  A universal compiled/runtime crossing tax cannot explain their gaps.
- `control_from_error` reaches four applications substantially but is absent
  from the shared CPU leaves, reproducing the earlier nil-check result.
- `NewStructInstancePositionalSized` is exact and broad, but its one escaping
  object is the remaining semantic runtime representation. The allocation
  budget test already enforces one allocation for small escaping instances.
- Eliminating that object would re-open a broad Result/control carrier ABI,
  lazy generic nominal representation, or caller-owned nominal lifetime
  architecture. This tranche explicitly excludes those closed routes.
- Generic union AST/type construction is highly material in Option/Result
  Config alone. A one-application cache or `Result` rule would fail the
  unlike-program and shared-nominal constraints.
- Array, interface, and other nominal conversions occur at only single-digit
  reach outside the three material encoding rows.

No candidate was implemented, so no A/B wall-time claim is made. The current
scorecard remains the governing Able/Go timing authority.

## Verification

- 14/14 strict generated modules built: seven telemetry and seven normal.
- All 14 generated-source and dependency checks omit
  `able/interpreter-go/pkg/interpreter`; each graph contains 96 packages.
- 35/35 telemetry executions passed their public verifiers with stable JSON.
- 140/140 normal CPU-profile executions passed their public verifiers.
- 21/21 exact-allocation executions passed their public verifiers.
- All generated files, binaries, profiles, and caches were kept outside the
  RAM-backed `/tmp`.
- No production, test, stdlib, interpreter, bytecode, language, dependency, or
  WASM file changed.

## Next recommendation

Refresh and close strict-compiled scorecard availability for the eight rows
that still report failures in the July 25 repair report: Mutex Ledger, the
three regex audits, Log Routing Redaction, Config Validation Extraction,
Concurrent State Machines, and Policy Record Dispatch.

Why: this tranche closes the broad runtime-value boundary as a selectable
performance route, but the compiled scorecard still has unavailable or flaky
feature rows. Correctness and availability must be established before the
next performance owner can be selected, and the retained compiler-cache work
may have changed which former build failures now complete under the one-minute
guardrail.

What it entails: rebuild each row strictly with the current compiler and
catalog contract; classify generation timeout, Go-build failure, deterministic
runtime failure, and concurrency flake separately; require interpreter-free
dependency graphs and repeated public-verifier passes; and correct only a
shared language/compiler/runtime cause. Do not disguise a build timeout as a
runtime result or add regex, container, nominal, benchmark, or application
special cases.

Why it is important: a performance claim against Go is meaningful only for a
broad executable feature set. Closing these missing rows restores the evidence
base needed to choose the next general compiled lowering owner and prevents
optimization work from proceeding around unmeasured correctness holes. Do not
begin WASM work.
