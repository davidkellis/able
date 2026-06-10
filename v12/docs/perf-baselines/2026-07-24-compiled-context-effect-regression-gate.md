# Compiled context-effect regression gate

Date: 2026-07-24

## Decision

Retain the conservative imported inherent-method environment-effect proof in
the default compiler. Retain the caller-owned package-context and proven-pure
native-interface adapter corrections only behind the existing
`--experimental-execution-context` option. Do not make that ABI the default or
propagate it through all generated callables.

The default-ABI method proof clears every measured guard and has no
benchmark-, primitive-, or nominal-type selection. The broader execution-
context experiment still regresses a concurrency guard and therefore does not
clear the production admission bar.

No Able syntax, language semantics, bytecode VM, tree-walker, canonical
`able-stdlib`, dependency, or WASM change was needed.

## Root cause

The first allocation profile localized the large serial regressions to
cross-package context entry wrappers. The original experimental helper
returned a newly allocated `__able_execution_context` whenever the package
environment changed. In Wide Integer Records, one measured process attributed
1,747,661 flat allocations to `__able_context_with_environment`; default mode
had no measured-main allocations.

The package entry wrapper now supplies a caller-owned local context. A repeat
Wide profile attributed no measured-main allocation to the context helper.
This removes a general generated-call allocation rather than recognizing any
benchmark or data type.

A second omission was in the existing conservative package-environment proof.
Its fixed point covered free functions and interface implementations but not
inherent `methods Type` bodies, because those bodies are intentionally absent
from `allFunctionInfos`. Consequently, imported methods that only operate on
their arguments still crossed a package-environment entry wrapper.

`environmentEffectFunctionInfos` now extends that same fixed point to inherent
methods without changing unrelated generator passes. A method enters its raw
generated body only when its generated body and every generated callee are
proven unable to observe package or dynamic runtime state. A method that reads
a package binding retains the entry wrapper. Native-interface context adapters
use the same proof; unproven implementations retain their context entry.

## Default-ABI admission cohort

Each side used independently built binaries, the public Ruby verifier, CPU 0,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 60-second runtime limit. Cohort A ran old
then candidate; cohort B reversed the order. The two initially noisy rows,
Wide Integer Records and Concurrent Packet Codecs, received a third ten-run
reverse-order cohort. Values below pool all successful process samples.

| Application | Processes/side | Old mean | Candidate mean | Change | Old GC | Candidate GC |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Wide Integer Records | 20 | 0.1745 s | 0.1725 s | -1.15% | 3.6 | 3.8 |
| Fixed Width 128 | 10 | 0.2150 s | 0.1940 s | -9.77% | 4.8 | 4.9 |
| Rational Series | 10 | 0.1270 s | 0.1250 s | -1.57% | 3.0 | 2.9 |
| Unicode Scalar Pipeline | 10 | 0.2520 s | 0.2500 s | -0.79% | 5.0 | 5.0 |
| Concurrent Packet Codecs | 20 | 0.4600 s | 0.4560 s | -0.87% | 4.5 | 4.3 |
| Mutex Work Queue | 10 | 0.7770 s | 0.7220 s | -7.08% | 4.1 | 4.0 |

All 80 candidate and 80 old processes passed verification. The unweighted
geometric mean across these six unlike guards improves by 3.61%. No guard
regresses in the pooled wall-time result.

Raw process records are
`2026-07-24-method-proof-default-{a,b}-<application>-{old,candidate}.json`;
Wide and Packet also have cohort `c`.

## Experimental-ABI gate

The caller-owned context plus inherent-method and native-interface proof
subcandidate substantially repaired the original serial losses in its first
five-process cohort:

| Application | Default | Experimental | Change |
| --- | ---: | ---: | ---: |
| Wide Integer Records | 0.212 s | 0.188 s | -11.3% |
| Fixed Width 128 | 0.210 s | 0.202 s | -3.8% |
| Rational Series | 0.120 s | 0.154 s | +28.3% |
| Unicode Scalar Pipeline | 0.258 s | 0.252 s | -2.3% |
| Concurrent Packet Codecs | 0.666 s | 0.482 s | -27.6% |
| Mutex Work Queue | 1.478 s | 1.680 s | +13.7% |

Rational was volatile across exploratory cohorts, but the Mutex regression
reproduced the earlier full-scorecard loss. Therefore this is useful
experimental infrastructure, not an admitted ABI rollout.

A broader subcandidate passed `nil` context to functions proven package-
environment-independent. It made Mutex Work Queue materially worse
(2.672 s versus 1.846 s) and was reverted. The package-environment proof does
not prove scheduler-context independence.

Mutex CPU profiles explain the distinction. Experimental processes spent
about 92% cumulative CPU below `bridge.currentGID`/`runtime.Stack`, reached
through dynamic await/callable paths including `__able_call_value_fast`,
`__able_invoke_awaitable_method`, and `__able_await_with_state`. Static package
purity cannot safely remove or reconstruct that task-local context.

## Correctness gate

Focused generated-code tests cover:

- pure imported methods using raw generated bodies;
- package-binding methods retaining their entry wrappers;
- the runtime entry wrapper remaining emitted;
- caller-owned package context generation;
- pure native-interface implementations using context bodies directly;
- package-dependent native-interface implementations retaining guarded entry;
- default ABI output containing no experimental context sibling; and
- static spawn, dynamic-call, bound-method, and captured-callback context
  behavior.

The focused method/interface gate passes in 0.216 seconds. The broader
non-fixture execution-context batch passes in 10.815 seconds. A single regex
that also selected all fixture and dynamic-boundary parity cases exceeded the
one-minute aggregate test limit while CPU-active; the focused batches are the
accepted bounded gate.

## Next

Profile and model the generated dynamic callable/await boundary shared by
Mutex Work Queue and Concurrent Event Routing. Add an explicit scheduler-
context effect only if it can distinguish callables that require task-local
state from ordinary static package-pure calls.

This is next because the remaining context-ABI loss is no longer an allocation
or primitive representation problem: it is repeated goroutine identity
recovery at dynamic await boundaries. The next tranche should attribute
callable creation, capture, invocation, and await completion; prototype a
general context-carrying callable ABI with compatibility wrappers only at
truly dynamic/external boundaries; and gate it with repeated, reverse-order
cohorts over Packet Codecs, Audio Mixing, Scene Updates, Graph Traversal,
Mutex Work Queue, and Concurrent Event Routing. Do not select benchmarks,
named containers, or nominal types, and do not begin WASM work.
