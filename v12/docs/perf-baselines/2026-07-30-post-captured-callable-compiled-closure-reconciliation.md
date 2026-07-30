# Post-captured-callable compiled closure reconciliation

Date: 2026-07-30

## Decision

Retain no additional compiler or runtime change.

The retained statically monomorphic captured-callable rule has exact current
source reach in three of the 66 compiled applications: Versioned Telemetry
Pipeline, Manifest Normalization, and Binary Event Log. Ten verifier-backed
main-only CPU profiles per reached application, three lightweight exact
allocation measurements per application, and one exact allocation-shape
profile per application expose no removable generated-code or generated-runtime
owner that is material in all three.

Checked arithmetic repeats in Binary Event Log and Versioned Telemetry, but not
in Manifest Normalization, and its general helper alternatives are already
closed by broad A/B evidence. Allocation and GC ancestry reaches all three, but
the concrete descendants and required lifetimes differ. Replacing escaping
identity-bearing Able structs with copied Go values would change alias-visible
mutation semantics; it is not a sound general lowering rule.

No canonical stdlib, runtime, interpreter, bytecode VM, language, dependency,
benchmark, or WASM source changed in this tranche.

## Exact reach census

The current benchmark and canonical-stdlib sources contain six unannotated
local lambda declarations that can enter the retained fresh-lambda inference
path. Frozen `HEAD` and current compilers generated all six with
`--no-fallbacks`. Exact callable-declaration comparison found:

| Application | Callable carrier changed | Current carrier |
| --- | --- | --- |
| Binary Event Log | yes | `EventRecord, i64 -> i64` |
| Manifest Normalization | yes | `ManifestRecord, i64 -> NormalizedManifest` |
| Versioned Telemetry Pipeline | yes | `Sample, Sample -> i64` |
| Concurrent Event Routing | no | byte-identical to frozen baseline |
| Concurrent Document Pipeline | no | byte-identical to frozen baseline |
| Policy Record Dispatch | no | byte-identical to frozen baseline |

The other 60 compiled application trees and the canonical stdlib contain no
declaration that enters this local fresh-lambda path. The compiler call graph
also confines the retained correction to
`forwardFreshLambdaTypeExpr` through static callable-argument inference.

The exact generated-module SHA-256 values are retained in the machine-readable
companion. The three reached modules changed; the three zero-reach controls
have identical frozen/current hashes.

## Strict binary and dependency controls

All six current modules built, ran, and passed their public verifier. Every
`go list -deps` graph contains 96 packages and omits the exact package
`able/interpreter-go/pkg/interpreter`.

| Application | Binary bytes | Binary SHA-256 |
| --- | ---: | --- |
| Binary Event Log | 13,931,224 | `7880784210fe406cbd689b4ad5400ea469b335d38d247b32ebab2e70360a1667` |
| Manifest Normalization | 13,891,248 | `97003d4c990b52d77606864a76b4f967be3f3d563226e618bd6db6e476f8f621` |
| Versioned Telemetry Pipeline | 12,962,624 | `ff9301c782b53f82cb008a44701628a80fd5fdf72be5d8533b3723cd3108f6fc` |
| Concurrent Event Routing | 14,943,760 | `2458cf3f34800f010df652c02976e5ec5e49a2c89d53763e9fb5a8fa9bacce0d` |
| Concurrent Document Pipeline | 13,626,896 | `49bcf837c5bc78e14b67b4188c9f66a50ef63878ef56e24b8c2527a204f9abad` |
| Policy Record Dispatch | 22,118,720 | `5fc84a4cce030db2e6f61936e88a9638ad675a5f89d82dc025e45cf2b42734d2` |

## Current CPU ownership

Each row combines ten independent main-only profiles on CPU 12 under
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, and `GOGC=50`. All 30 launches passed the
public verifier.

| Application | Merged samples | Largest current owners |
| --- | ---: | --- |
| Binary Event Log | 0.51 s | checked multiply 19.61%; divmod 9.80%; checked add 7.84%; `parse_event` 56.86% cumulative; allocation 17.65% cumulative |
| Manifest Normalization | 0.04 s | String split/UTF-8 validation and allocation; the process is below useful leaf-level CPU resolution |
| Versioned Telemetry Pipeline | 22.85 s | checked multiply 22.93%; divmod 18.77%; checked add 10.81%; allocation 19.52% cumulative; policy adapters 8.27% and 5.08% cumulative |

The intersection has no admissible exact leaf. Arithmetic is absent from the
Manifest sample and was already rejected by unlike-program timing gates.
String work is Manifest-specific, HashMap work is Binary-specific, and
interface-policy dispatch is Telemetry-specific.

## Exact allocation ownership

Three lightweight main-phase measurements per application avoid allocation
profiler serialization costs:

| Application | Allocated-byte mean | Allocation mean | GC mean |
| --- | ---: | ---: | ---: |
| Binary Event Log | 9,321,773.33 | 171,069.33 | 8 |
| Manifest Normalization | 1,397,200.00 | 27,078.00 | 1 |
| Versioned Telemetry Pipeline | 430,788,165.33 | 13,325,303.00 | 351 |

The byte ranges are 16 bytes, zero bytes, and 224 bytes, respectively.
Allocation-count ranges are one, zero, and zero. These totals reproduce the
post-carrier retained measurements.

One separate `runtime.MemProfileRate=1` process per application passed the same
verifier. Start/end profile subtraction includes the known profile-writer
serialization allocation, so selection used generated-source line attribution
plus the independent totals above:

- Telemetry allocates 13,208,878 `Sample` objects (403.10 MiB) at its hot
  nominal literal before storing them in its generic window.
- Binary allocates 53,248 `EventRecord` objects (2.44 MiB) on the parse return
  path; its other material allocation descendants are error conversion,
  generic-union wrapping, HashMap storage, and one dynamic integer boundary.
- Manifest allocates 3,072 `ManifestRecord` objects (240 KiB), while its other
  descendants are String split/conversion, normalization results, Result and
  Option wrappers, and error conversion.

These are not one removable ABI. Telemetry's values escape into generic
storage; Binary and Manifest return different larger records through control
and union paths. The retained caller-owned nominal-result ABI already removes
proven nonescaping fresh small results while preserving the pointer ABI at
escaping, aggregate, interface, and dynamic boundaries. Extending copied value
semantics to these escaping records would violate Able reference identity and
field-mutation behavior.

## Closure result

The retained compiler rule has exact zero reach in the compiled target guards,
current control, Sudoku quotient, float, wide-numeric, byte-output, regex, and
concurrency groups. It reaches Binary Event Log and Manifest Normalization in
the text/map and iterator/control surface, and Versioned Telemetry in
iterator/control. The current three-application profiles close that changed
surface with no shared admissible successor.

The compiled frontier therefore returns from `refresh-required` to
`closed-no-shared-leaf`. This record also supports rebasing the compiled
architecture-target and cross-family ownership closures on the reviewed
compiler production identity.

## Verification and cleanup

- 30 main-only CPU processes passed public verifiers.
- Nine lightweight exact allocation processes passed public verifiers.
- Three exact allocation-shape processes passed public verifiers and stayed
  below one minute; Telemetry, the longest, completed in 28 seconds.
- Six strict current binaries passed public verifiers and omitted the
  interpreter package.
- All large generated modules, binaries, raw profiles, and the shared Go cache
  remained under disk-backed `/var/tmp`.

## Next

Run a full 66-row static native-boundary census before selecting another
compiled candidate.

Why: the three refreshed profiles expose no safe shared successor, while the
remaining compiled misses still span 1.07x-23.80x equivalent Go. Reprofiling the
same three rows or reopening checked arithmetic, GC, or nominal identity would
repeat closed work.

What it entails: generate current strict modules for all selected compiled
rows and count remaining `runtime.Value` conversions, bridge calls, dynamic
call sites, interface adapters, union conversions, and heap nominal literals
by semantic boundary. Join those static counts to current timing excess and
existing profiles, then admit only an exact lowerable boundary with material
reach in three unlike applications. Preserve escaping and dynamic paths and do
not add application, container, or non-primitive nominal-name rules.

Why it matters: this searches directly for places where Able still crosses
from native Go carriers into erased runtime representation, which is the most
credible remaining route toward Go-native compiled performance.
