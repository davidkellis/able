# Bytecode Frame-local Array Ownership: Historical Investigation

## Status

This record preserves the release-disabled observer work and the rejected
explicit-release experiment. The active contract is
[bytecode-frame-array-ownership.md](./bytecode-frame-array-ownership.md).
Nothing here authorizes a VM lease-release path.

## Former production model

The investigated model used an ownership context per eligible inline VM frame:

```text
active frame
  owned: set[*ArrayValue]      # pointer wrappers created in this frame
  escaped: set[*ArrayValue]    # wrappers crossing a conservative barrier
  parent: frame context        # restored at inline return
```

It was intentionally a VM sidecar, not a new Array representation or a
replacement for ArrayStore leases. Empty contexts were needed along an enabled
inline call chain so a canonical `Array.with_capacity` result could transfer
through a caller that had not created an Array itself. The maps were lazy.

The proposed provenance rules were:

1. Track only bytecode-created wrappers at an Array literal, verified
   canonical `Array.new`, or canonical kernel Array materialization.
2. At an inline return, walk returned dynamic Arrays and supported aggregate
   wrappers; transfer reachable owned wrappers to the parent. A public return
   would leave them unowned by the sidecar.
3. Mark a wrapper escaped before cleanup when it enters a public return,
   environment/global/capture, aggregate/interface/map/Future/channel,
   extern/dynamic/unknown/spawn call, or a borrowed Array write.
4. A hypothetical cleanup would exclude returned and escaped wrappers and
   release only every remaining tracked wrapper before the VM cleared the
   frame, including error unwind.

The model deliberately rejected source-assignment provenance, slot-only
provenance, raw-handle view release, and rules for named stdlib containers.
Those shortcuts cannot preserve Able reference semantics.

## Observer evidence

The release-disabled observer validated that provenance bookkeeping can make
the above distinctions:

- A generic nested `Array (Array i32)` control completed one 25-wrapper local
  graph per marker.
- Each MatrixMultiply iteration created 100 wrappers: the 25-wrapper
  transpose was frame-local, while current `a`, `b`, and returned `d` graphs
  remained live until `main` completed.
- A flat `Array.map` control kept its source graph local but recorded six map
  result graphs as closure escapes, making it a guard rather than a release
  candidate.

The first profile exposed an observer bug: a canonical factory result through
an empty inline caller looked public because that caller had no context. The
observer now installs an empty context for enabled active callers and focused
coverage proves the transfer. Detailed snapshots and guard settings are in
`docs/perf-baselines/2026-07-11-bytecode-array-ownership-observer-profile.md`.

Program-finalization metadata later marked the three canonical creation
boundaries and capture/dynamic/spawn/aggregate barriers. It prevents observer
allocation until a diagnostic run reaches an eligible program. It does not
prove ownership or alter ordinary execution.

## Rejected explicit-release experiment

An opt-in follow-up released only a tracked wrapper classified as unreturned
and unescaped when an inline frame completed, including error unwind. Focused
checks retained returned/unknown-call aliases and verified release for the
frame-local/error-unwound controls. The normal-loader nested, Matrix, and flat
controls preserved output and the expected classifications.

However, paired externally verified one-run bytecode controls were mixed:

| Benchmark | Default | Explicit release | Result |
| --- | ---: | ---: | --- |
| `base64` | 3.130s | 3.070s | 1.9% faster |
| `i_before_e` | 0.560s | 0.610s | 8.9% slower |

Both used pinned one-process guardrails and passed their verifiers. The patch
was removed: it had no three-application shared leaf, regressed an unlike
text/iterator workload, and would add complex live-frame work to a broadly
used execution path. The observer and diagnostic lowering metadata remain;
they perform no release.

## Reconsideration requirements

A future candidate must satisfy the active performance policy, preserve all
alias/return/aggregate/closure/Future/extern/error behavior, prove no normal
VM overhead outside its eligibility conditions, and pass the broad bounded
coverage/performance gate. The historical model is a starting point for proof
design only, never a reason to bypass those requirements.
