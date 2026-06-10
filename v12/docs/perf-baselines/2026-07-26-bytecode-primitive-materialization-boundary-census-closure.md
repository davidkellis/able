# Bytecode primitive-materialization boundary census closure

Date: 2026-07-26

## Decision

Retain the opt-in, reason-aware materialization diagnostics. Reject and fully
revert the production experiment that stored raw primitive snapshots in
generic nominal struct fields.

The observer completed all 54 rankable bytecode applications, every public
verifier passed, and no counters were dropped. The broadest candidate boundary
was nominal/collection construction, but removing that boundary did not produce
a broad runtime win. Sensor Calibration regressed 6.93% and Rational Series
regressed 4.28%, so the experiment fails the unlike-program retention bar.

## Measurement contract

- CPU 6, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`.
- Disk-backed `/var/tmp` for build, profile, and result artifacts.
- Canonical external stdlib with source-root-only module resolution.
- Required serial/goroutine executor selected from the benchmark catalog.
- Main-only materialization statistics.
- Public verifier input through stdin for all 54 applications.
- Five rotated baseline/candidate/Python/Ruby processes per admitted A/B row.

The retained/restored CLI is
`f68f4e7314300999f257395da2a34d0c0d14e20557757f6535da89aa6466095c`.
The retained/restored interpreter test artifact is
`aa525490df683d83e9b26f59723e2122139cbf209023829e89ed9475bbf7dad2`.
Both match their pre-prototype artifacts byte for byte. The rejected candidate
hashes were `e1d8d63d3e32b2259ecde74be1d8552321df85b96944de67271786bb5822e845`
and `a6129932197e1e738918a4811f502866e50c81990112bc2f08fa36dec76e454a`.

## Corpus census

The 54 verified applications took 1,034.618 seconds under diagnostics and
recorded 18,696,646 materializations:

| Class / reason | Count | Applications | Applications with at least 1,000 |
|---|---:|---:|---:|
| candidate / collection value | 11,088,674 | 37 | 30 |
| required / interface or union | 3,917,679 | 14 | 14 |
| candidate / static return | 1,678,255 | 39 | 32 |
| required / environment | 1,472,008 | 30 | 20 |
| candidate / cast | 536,001 | 2 | 2 |
| candidate / static call | 4,029 | 1 | 1 |

Candidate-static transitions total 13,306,959; required-dynamic transitions
total 5,389,687. Base64, Distance Field, I Before E, JSON, Monte Carlo, and RMS
are zero-transition controls.

The largest concrete collection shapes were:

| Opcode / carrier | Count | Applications |
|---|---:|---:|
| `StructLiteralNamedFast` / `i32_slot_value` | 5,597,919 | 31 |
| `CallMemberArraySlot` / `i32_slot_value` | 2,731,820 | 9 |
| `StructLiteralNamedFast` / `integer_result<u64>` | 999,999 | 1 |
| `StructLiteralNamedFast` / `i64_slot_cell` | 846,269 | 10 |
| `StructLiteralNamedFast` / `integer_result<i64>` | 835,739 | 15 |

## Ordinary profile reconciliation

Fresh diagnostics-off measured-main profiles established that materialization
was real but not a universal dominant CPU owner.

- Sensor Calibration: 2.607 seconds, 117,023,144 bytes, 1,716,042
  allocations. `NewStructInstance` owned 11.54% of sampled allocation;
  integer boxing owned 2.83% and primitive materialization 1.41%.
- Word Frequency: 1.673 seconds, 48,404,144 bytes, 637,221 allocations.
  `NewStructInstance` owned 14.42% of sampled allocation; integer boxing and
  materialization each owned 1.01%.
- Concurrent Audio Voices: 1.941 seconds, 181,192,496 bytes, 4,103,008
  allocations. Integer boxing owned 10.45%, primitive materialization 10.07%,
  and `NewStructInstance` 4.90%.
- Array Slice Window: 0.506 seconds, 14,190,688 bytes, 422,243 allocations.
  `CallMemberArraySlot` was 22.64% cumulative CPU and primitive
  materialization was 3.77% cumulative.

This admitted one general experiment: keep immutable raw primitive snapshots
inside ordinary nominal struct positional storage until a later dynamic
consumer. It did not name or special-case any application or nominal type.

## Candidate A/B result

All 140 external processes passed their public verifier.

| Application | Baseline | Candidate | Change | Candidate / Python | Candidate / Ruby |
|---|---:|---:|---:|---:|---:|
| Sensor Calibration | 3.3577s | 3.5905s | +6.93% | 97.46x | 36.09x |
| Word Frequency | 1.4611s | 1.4625s | +0.10% | 79.26x | 27.73x |
| Concurrent Audio Voices | 1.3731s | 1.3601s | -0.95% | 10.67x | 11.34x |
| Policy Record Dispatch | 7.5975s | 7.2540s | -4.52% | 347.94x | 156.92x |
| Rational Series | 4.2387s | 4.4202s | +4.28% | 40.99x | 31.32x |
| Fixed Width 128 | 8.3480s | 8.0397s | -3.69% | 21.97x | 12.38x |
| Array Slice Window | 0.7532s | 0.7473s | -0.79% | 24.16x | 10.24x |

Three exact measured-main allocation samples per variant were completed before
Policy exposed that the reusable-main harness cannot warm that program because
of an ambiguous `map` overload. Sensor and Word reduced bytes by 2.36% and
2.26% while allocation counts were neutral. Concurrent Audio reduced bytes by
2.56% and allocation count by 2.36%. These partial memory gains cannot rescue
a candidate already rejected by verifier-backed runtime regressions.

## Retained implementation

The opt-in observer records:

- candidate-static versus required-dynamic class;
- semantic reason;
- raw carrier and primitive suffix;
- opcode and instruction offset;
- containing bytecode program;
- Able source origin, line, and column.

Per-VM pointers to atomic counters avoid a global lock on every recorded
transition. Reset and snapshot behavior are covered by focused tests. Ordinary
execution leaves statistics disabled.

No compiler, runtime-semantics, tree-walker, stdlib, benchmark, language,
dependency, or WASM change was required. The rejected nominal-field carrier
experiment was reverted, and rebuilt artifacts prove the restoration exactly.
The complete `./run_all_tests.sh` handoff passed every coverage, scoreboard,
threshold, non-compiler, and compiler-batch contract; the final bytecode
fixture corpus completed in 108.207 seconds.

The machine-readable companion is
`2026-07-26-bytecode-primitive-materialization-boundary-census-closure.json`.
Disposable raw census, profile, and A/B workspaces were removed after their
checksums and summarized results were recorded.

## Recommendation

Next reconcile the 2,731,820 `CallMemberArraySlot` `i32` materializations
across the nine affected applications with the existing Array raw-cache and
direct-store lanes. Profile at least three unlike material applications
(including Policy or a regex/log workload) plus Array Slice Window as the
known hot control. Prototype only a general primitive-Array element carrier
rule whose cost repeats in those profiles, then require five balanced
verifier-backed A/B runs.

This is next because it is the largest remaining concrete candidate boundary
after the generic nominal-field route failed. It is important because static
Arrays are explicitly required to use native carriers, while an Array-specific
primitive storage rule can avoid boxing without weakening the uniform encoding
of non-primitive nominal types or increasing compiler/interpreter crossings.
