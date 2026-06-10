# Compiled generic-union direct-known-method gate — 2026-07-22

## Decision

Retain the general generated-runtime change. Statically resolved native methods
on generic named unions now receive the receiver directly instead of first
boxing a `runtime.NativeBoundMethodValue` into `runtime.Value` and redispatching
through the generic fast-call switch.

The rule names no union, record, container, benchmark, or stdlib type. It
preserves argument-zero receiver injection, the active environment and
`RuntimeData` in `NativeCallContext`, native error/Able control conversion, nil
normalization, existing arity behavior, and the unchanged dynamic fallback.
No canonical-stdlib, VM, language, or WASM change was needed.

## Verification protocol

Current binaries were preserved before the source edit. Candidate binaries were
built once and preserved afterward. Every run used the same binary for its
variant and the public application verifier. The gate contains 252 successful
processes with zero failures or timeouts:

- 200 current/candidate timing processes;
- 20 lightweight candidate allocation-stat processes;
- four exact candidate allocation-profile processes;
- 28 forward/reverse build-smoke processes.

Every workstation sample is retained. Owner applications use two
direction-reversed cohorts of five processes per variant. The initial five
guards use two cohorts of three; volatile N-Body, K-Nucleotide, and Matrix
Multiply received two additional reversed cohorts of five. Timing processes
used `GOMEMLIMIT=1GiB`, `GOGC=50`, their catalog CPU budget, and a 55-second
per-process cap.

## Owner wall gate

| Application | Samples/variant | Current mean | Candidate mean | Delta |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 10 | 0.685 s | 0.631 s | -7.88% |
| Option/Result Config | 10 | 0.217 s | 0.176 s | -18.89% |
| Manifest Normalization | 10 | 0.196 s | 0.185 s | -5.61% |
| Policy Record Dispatch | 10 | 0.211 s | 0.195 s | -7.58% |

All four unlike owner applications improve. This is substantially broader than
the three-application admission minimum and includes binary decoding, numeric
configuration, text/record normalization, and regex/interface policy dispatch.

## Exact allocation gate

| Application | Current objects | Candidate objects | Object delta | Current bytes | Candidate bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,850,462.8 | 4,465,436.8 | -7.94% | 290,399,603.2 | 273,425,644.8 | -5.85% |
| Option/Result Config | 1,630,658.4 | 1,360,322.0 | -16.58% | 57,447,832.0 | 46,044,529.6 | -19.85% |
| Manifest Normalization | 1,034,216.4 | 997,350.6 | -3.56% | 47,288,388.8 | 45,704,526.4 | -3.35% |
| Policy Record Dispatch | 970,530.0 | 961,314.8 | -0.95% | 49,166,376.0 | 48,748,643.2 | -0.85% |

The candidate source contains no old bound-method composite. Four exact
main-start/main-end allocation differences contain the new direct helper and
no allocation attributed to the deleted expression. The reductions exceed the
original 64-byte box counts because the direct call also lets Go keep more of
the now-concrete call setup off the heap.

## Unrelated guard gate

| Guard | Samples/variant | Current mean | Candidate mean | Delta | Interpretation |
| --- | ---: | ---: | ---: | ---: | --- |
| Binary Trees | 6 | 9.948 s | 9.720 s | -2.30% | improves |
| N-Body | 16 | 0.1656 s | 0.1688 s | +1.89% | timer/noise interval crosses zero |
| K-Nucleotide | 16 | 2.923 s | 2.979 s | +1.90% | cohort directions disagree; interval crosses zero |
| Matrix Multiply | 16 | 1.196 s | 1.231 s | +2.87% | cohort drift; interval crosses zero |
| Mutex Ledger | 6 | 0.8567 s | 0.7450 s | -13.04% | improves |

The three positive point estimates are not confirmed regressions: approximate
95% difference intervals all include zero after the supplemental cohorts. The
two longer or concurrent guards improve, while owner improvements are
consistent and much larger. That combination meets the broad-retention bar.

## Tests and durable evidence

Focused generated-source and executable tests cover the absence of the box,
presence of the direct helper and fallback, receiver order, a void/nil result,
and raised-control propagation. The standalone generic-named-union executable
test also passes. No test exceeds one minute.

The companion JSON retains every timing and allocation observation needed to
recompute these means after temporary build/profile artifacts are cleaned:
`2026-07-22-compiled-generic-union-direct-known-method-gate.json`.

## Next direction

Refresh the four owner CPU profiles from the retained implementation and use
them to select the next exact shared compiled leaf. The old bound-method box is
gone, so a fresh profile is required before considering call-context setup,
receiver-slice injection, nominal record conversion, or another allocation
wall. Admit a candidate only if the same concrete child repeats across at least
three unlike applications; keep bytecode work on its separately profiled
frontier and continue to defer WASM.
