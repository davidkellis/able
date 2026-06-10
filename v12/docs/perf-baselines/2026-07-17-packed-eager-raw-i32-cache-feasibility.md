# Packed eager raw-i32 cache feasibility

Date: 2026-07-17

## Decision

Reject and fully remove the pointer-backed eager raw-`i32` cache prototype.
The representation removes nearly all per-entry interface boxing during
package initialization, but changing cached values from the concrete
`bytecodeRawI32SlotValue` dynamic type to pointers makes the VM's central raw
integer extraction and slot paths materially slower. The candidate therefore
fails the hot-path admission gate and does not proceed to application timing.

No production, stdlib, benchmark, fixture, or language change is retained.

## Representation audit

The eager raw cache contains 263,168 `runtime.Value` interface entries covering
`-1024...262143`. Constructing each interface from the concrete raw value causes
one box per entry. The trial added one concrete backing array and pointed each
interface entry at its corresponding cell. That preserves eager storage and
cache identity, while changing the interface's dynamic type from
`bytecodeRawI32SlotValue` to `*bytecodeRawI32SlotValue`.

The dynamic type is observed by roughly two dozen production type switches
covering integer extraction/materialization, typed slots, comparison, array
indexing, float conversion, and type inspection. A centralized carrier helper
was tested first; a narrower version then put the pointer case directly in the
existing hot switches. Both retained numeric semantics in the focused suite,
but neither retained dispatch cost.

## Initialization result

Both test binaries were built before their traces and run with
`GODEBUG=inittrace=1` and no test body:

| State | Interpreter init bytes | Init allocations | One trace clock |
| --- | ---: | ---: | ---: |
| Preserved baseline | 38,000,968 | 707,321 | 57 ms |
| Pointer-backed candidate | 39,065,272 | 444,426 | 31 ms |
| Restored source | 38,003,400 | 707,336 | 29 ms |

The candidate removes 262,895 allocations (37.17%) but adds 1,064,304 bytes
(2.80%) for the concrete backing array. The restored trace also demonstrates
that a single initialization clock is workstation-noisy; object and byte
counts establish the representation effect, while the dispatch benchmarks
decide admission.

## Controlled hot-path result

The preserved baseline and candidate test executables each ran seven fixed
one-million-iteration samples per sub-benchmark. These are the candidate's
narrowest direct-switch revision, not the more expensive initial carrier-helper
revision:

| Path | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| raw integer info, cached raw `i32` | 2.5800 ns | 2.7443 ns | +6.37% |
| direct small `i32`, cached raw `i32` | 1.1615 ns | 2.3426 ns | +101.69% |
| validated slot, cached raw `i32` | 2.4377 ns | 3.4360 ns | +40.95% |
| direct small `i32`, boxed `i32` | 5.9519 ns | 9.0727 ns | +52.43% |
| direct small `i32`, string miss | 0.6005 ns | 3.0713 ns | +411.46% |

Every path remained allocation-free. The cached pointer is slower even before
application-level effects, while adding its recognition also taxes unrelated
boxed and miss cases. These are too large and too broad to average away or
justify with the startup allocation reduction.

## Verification and cleanup

- Focused raw-`i32`, slot, integer, array-index/compare, and float-conversion
  tests passed with the candidate.
- After removal, the focused representation and formerly failing store/fused
  frame guards pass again.
- The package-wide test attempt reached the repository's one-minute limit in
  the existing regex fixture lane; it is not used as candidate evidence.
- All pointer-carrier production cases, helper calls, backing storage, and
  representational test edits were removed.
- Raw test binaries, traces, and benchmark outputs were deleted after this
  aggregate record was written.

## Next recommendation

Test a statically initialized interface-table representation on a bounded
cache segment before attempting another full raw cache.

Why: the allocation result proves that per-entry interface boxing is the exact
startup mechanism, while the rejection proves that changing the dynamic type
is not viable. A generated static interface literal may let the Go linker place
the same concrete `bytecodeRawI32SlotValue` boxes in read-only data, preserving
all existing type switches and hot paths.

What it entails: generate a small segmented table outside production first and
measure compiler/linker behavior, binary-size growth, init allocations, lookup
identity, and the same seven-run extraction suite. Keep generated source files
under 1,000 lines and reject the idea if static boxes are still initialized at
runtime, compile/link cost or binary growth is disproportionate, or any lookup
path changes. Only a successful bounded segment should advance to the full
cache and then the broad compiled/bytecode application controls. Do not use
unsafe representation tricks, lazy initialization, GC-policy changes, dummy
ballast, workload detection, or named-container exceptions.
