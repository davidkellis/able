# Bytecode Array ownership observer profile (2026-07-11)

## Purpose

Validate the release-disabled bytecode frame-ownership observer against the
same source-program controls used for the preceding live-state and escape
attribution. This is not an ArrayStore reclamation experiment: the observer
only counts pointer identities and never releases a lease.

## Method

The opt-in `TestBytecodeProgramRuntimeArrayOwnership` harness loads each
program through the ordinary bytecode loader, resets its observer after module
setup, invokes `main` through a lowered bytecode call expression, and records
snapshots before `main`, at each program `print`, and after `main`. Its only
print override records a marker; program output is otherwise immaterial to the
ownership counts.

Each control was a fresh, bounded process with:

```text
taskset -c 2
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1
GOCACHE=/tmp/able-retention-go-cache.BqynbE
ABLE_BENCH_SKIP_TYPECHECK=1
```

The exact targets were the existing generic nested-Array, MatrixMultiply, and
flat Array-map marker drivers in
`v12/tmp/live-array-ownership-2026-07-11/drivers/`.

## Observer correction

The initial nested reading reported 25 public returns per marker. A trace
showed that both `round` and canonical `Array.with_capacity` were inline VM
calls. The issue was observer bookkeeping: an inline caller that had not yet
created an Array had no context, so a child factory result had no parent and
was mislabeled public. The observer now gives every enabled inline boundary an
empty context while retaining lazy owned/escaped maps. Its regression test
proves a returned Array transfers through an empty caller and becomes
frame-local when that caller completes.

## Results

| Control | Per-print cumulative result | After `main` | Interpretation |
| --- | --- | --- | --- |
| Nested `Array (Array i32)` | At marker *n*: `created=25n`, `transferred=25n`, `frame_local=25n`, no escapes or public returns. | `created=200`, `frame_local=200`. | Every 25-wrapper graph completes during its `round` call; this matches the prior stale-per-round retention shape. |
| MatrixMultiply `Array (Array f64)` | At marker *n*: `created=100n`, `transferred=175n`, `frame_local=25n`, no escapes or public returns. | `created=300`, `frame_local=300`. | The temporary transpose `c` graph is local at each marker; current `a`, `b`, and returned `d` graphs complete only when `main` exits. |
| Flat `Array.map` i32 | First marker: `created=2`, `transferred=4`; each later marker adds one created result and two transfers; no frame-local result at markers. | `created=7`, `transferred=14`, `frame_local=1`, `escaped=6` with reason `closure`. | The initial source graph is local. Each map result reaches a conservative closure boundary and must remain outside a release candidate. |

`public_returned` and `error_unwound` were zero for every marker and final
snapshot. The observer also now handles an explicit bytecode `return` as a
return boundary rather than an error unwind; focused tests cover both public
and detached explicit returns.

## Decision

Keep the profile hook and observer, but keep release disabled. The controls
confirm a broadly useful ownership model and reject a simplistic "release every
finished Array" rule.

## Lowering-time gated rerun

The next tranche added program-finalization metadata for canonical Array
creation boundaries: Array literals, canonical Array-new, and canonical
kernel-shaped Array literals. It also records capture, dynamic import, spawn,
and aggregate barriers; those are future release rejections, not source- or
container-specific optimizations. The profile now attaches no observer at VM
acquisition. It activates only when execution enters a program carrying a
canonical creation marker and reconstructs empty contexts for active inline
callers so factory provenance can still return through them.

The gated fresh-process rerun reproduced the table above exactly for all three
controls. Focused Array alias-through-return parity, explicit return handling,
error-unwind coverage, and the bytecode return/call-name guards passed. Normal
VM execution sees no observer allocation or ownership context.

## Explicit-release experiment (rejected)

A separately opt-in follow-up released only a tracked Array wrapper that was
unreturned and unescaped when its inline frame completed, including error
unwind. Focused checks confirmed that a frame-local wrapper and an
error-unwound wrapper lost their ArrayStore leases, while a returned wrapper
and aliases sent through an unknown-call boundary remained readable. The
normal-loader controls also preserved output: nested `Array i32` released all
25 completed wrappers per marker; MatrixMultiply released its 25-wrapper
transpose while retaining its three current 25-wrapper graphs; flat `map`
released only its source graph and retained six conservative closure escapes.

Both independently verified external bytecode rows preserved their stdout,
but the paired pinned controls failed the generality bar:

| Benchmark | Default | Explicit release | Result |
| --- | ---: | ---: | --- |
| `base64` | 3.130s | 3.070s | 1.9% faster |
| `i_before_e` | 0.560s | 0.610s | 8.9% slower |

Each was a one-run `taskset -c 2`, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1` comparison, with the existing upstream Ruby verifier passing
in both configurations. The experiment was removed rather than retained as
an opt-in knob: a small codec win does not justify a regression on an
independent text/iterator workload. The release-disabled observer and its
lowering metadata remain useful diagnostic infrastructure; they do not alter
normal bytecode lifetime behavior.
