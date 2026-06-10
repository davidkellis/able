# Compiled bootstrap allocation gate

Date: 2026-07-16

## Decision

Retain one profiling-infrastructure correction and no compiled-performance
candidate. Generated binaries now accept
`ABLE_GO_PHASE_ALLOC_PROFILE_DIR`, which records exact bootstrap/main
allocation snapshots and phase counters without starting Go's CPU profiler.
Normal binaries remain unchanged when the variable is unset.

The new mode proved that the previously observed 2.7–3.1 MB bootstrap band was
mostly profiling machinery. Real generated registration allocates about
219–612 KiB across the four applications. A shared simple-type metadata cache
removed 17–21% of those bytes, but two concurrency-safe implementations both
regressed Dependency Plan normal-process time. The compiler trial was fully
removed. No runtime, bytecode VM, canonical-stdlib, application, verifier,
reference, or benchmark source changed.

## Measurement correction

`ABLE_GO_PHASE_PROFILE_DIR` intentionally combines exact allocation snapshots
with CPU profiles. Its bootstrap counters therefore included
`runtime/pprof.StartCPUProfile`, which allocated about 1.15 MiB, plus CPU
profile buffers and related setup. The first start/end allocation difference
also contains allocations made while serializing the start snapshot because a
profile cannot include the allocations required to write itself.

The allocation-only mode retains exact `runtime.MemProfileRate=1` snapshots,
the two-GC boundary drain, `phase-stats.json`, and restoration of the prior
sample rate. It does not create `bootstrap.cpu.pprof` or `main.cpu.pprof` and
does not call `pprof.StartCPUProfile`. Phase counters begin after the start
snapshot is written, so the JSON deltas exclude snapshot serialization too.
Start/end profile attribution is restricted to stacks below generated
`RegisterIn`.

The mode is opt-in and mutually exclusive with the combined phase mode, the
CPU-only phase mode, and the standalone CPU profiler.

## Corrected baseline

Current binaries were built once with canonical `../able-stdlib` and ran from
their catalog working directories with `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`.

| Application | Combined-mode bootstrap bytes | Allocation-only bootstrap bytes | Real bootstrap allocations | `RegisterIn` attributed bytes |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 3,106,952 | 627,064 | 7,754 | 611.52 KiB |
| Dependency Plan | 2,858,200 | 396,392 | 4,989 | 386.25 KiB |
| Array Slice Window | 2,746,224 | 283,880 | 3,624 | 276.38 KiB |
| Base64 | 2,675,080 | 225,008 | 2,796 | 218.88 KiB |

All four runs reproduced their canonical stdout hashes:

| Application | SHA-256 |
| --- | --- |
| Document Audit | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |
| Dependency Plan | `96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38` |
| Array Slice Window | `155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e` |
| Base64 | `5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316` |

## Shared allocation attribution

The corrected `RegisterIn` differences repeat two concrete families:

| Application | `NewIdentifier` | `NewSimpleTypeExpression` | interface-dispatch registration |
| --- | ---: | ---: | ---: |
| Document Audit | 96.00 KiB | 59.31 KiB | 100.27 KiB cumulative |
| Dependency Plan | 72.44 KiB | 46.44 KiB | 100.26 KiB cumulative |
| Array Slice Window | 53.50 KiB | 32.19 KiB | 62.88 KiB cumulative |
| Base64 | 40.06 KiB | 23.69 KiB | 44.37 KiB cumulative |

The type nodes come from the same generic `renderTypeExpression` path in
generated package definitions, callable signatures, method registration, and
interface dispatch. Existing substitution and alias expansion treat simple
type nodes as immutable values, and the interpreter already shares equivalent
nodes internally. This admitted a general compiler candidate rather than a
named-interface or application rule.

## Rejected static simple-type cache

The trial changed generated simple type expressions to obtain shared metadata
from the compiler bridge. The cache was keyed only by the type name and
retained distinct nodes for distinct names. It did not change generic
substitution, alias expansion, interface identity, source spans, or dynamic
fallbacks.

The final mutex-backed formulation produced these exact bootstrap deltas:

| Application | Baseline bytes / allocations | Candidate bytes / allocations | Byte change | Allocation change |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 627,064 / 7,754 | 521,832 / 6,031 | -16.8% | -22.2% |
| Dependency Plan | 396,392 / 4,989 | 314,232 / 3,626 | -20.7% | -27.3% |
| Array Slice Window | 283,880 / 3,624 | 230,440 / 2,710 | -18.8% | -25.2% |
| Base64 | 225,008 / 2,796 | 183,056 / 2,114 | -18.6% | -24.4% |

Main-phase byte/allocation counts were unchanged. Every candidate output hash
matched its baseline.

Normal binaries were then measured with alternating launch order and
nanosecond wall timing. The three short applications used 60 independent
launches per side; Base64 used five per side.

| Application | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Document Audit | 60.452 ms | 61.048 ms | +1.0% |
| Dependency Plan | 58.847 ms | 60.995 ms | +3.7% |
| Array Slice Window | 64.090 ms | 64.018 ms | -0.1% |
| Base64 | 2.1992 s | 2.1954 s | -0.2% |

The earlier `sync.Map` formulation had independently produced the same mixed
shape over 60 short-application launches and five Base64 launches per side:
Document Audit +0.3%, Dependency Plan +4.1%, Array Slice Window +0.8%, and
Base64 -0.1%. Reducing allocations is not sufficient when every metadata use
pays a synchronization and lookup cost. The repeated Dependency Plan
regression fails the broad benchmark bar, so both cache formulations and their
tests were removed.

## Verification and cleanup

- `go test ./pkg/profilehook -count=1 -timeout 60s` passes, including proof
  that allocation-only mode writes four allocation snapshots plus phase stats,
  writes no CPU profiles, and restores `runtime.MemProfileRate`.
- Focused compiler definition/constraint, static-main, bootstrap-main, and
  generic-interface metadata tests pass after the candidate revert.
- `go build ./cmd/ablec` passes.
- Diff hygiene passes for the touched source.
- All baseline/candidate application invocations completed inside the normal
  project limit with stable output.

Generated source trees, binaries, profiles, and timing logs are temporary and
removed after this record.

## Next recommendation

Attribute the remaining generated interface-dispatch registration allocation
across the same short applications, with Base64 as the guard. Distinguish
necessary immutable dispatch entries from avoidable map/slice growth and
error-producing probe paths. Advance only if the generator knows an exact
cross-application capacity or can eliminate a repeated failed lookup without
adding work to each registered entry or runtime dispatch.

Why: after correcting the profiler and rejecting shared type-node lookup,
`__able_register_interface_dispatch` is the largest concrete compiler-owned
allocation descendant repeated in all four programs (about 44–100 KiB
cumulative). It is a language-level interface mechanism rather than a named
container or benchmark kernel, so a proven construction-only improvement
would benefit ordinary compiled applications.

What it entails: add temporary off-timing counts for dispatch tables, per-key
entry counts, capacity growth, and failed environment/type probes; reconcile
those counts with the allocation stacks; remove the counters; then trial only
a generic registration-time capacity/probe change. Gate it with static and
dynamic interface semantics, alias/package-identity controls, and two
independent alternating timing batches across the short applications. Do not
reopen literal interface-name comparison, default-AST codecs, or a runtime
dispatch fast path from startup evidence.
