# Compiled execution-context recommendation reconciliation — 2026-07-22

## Decision

Do not implement or time another spawn-scoped generated execution context.
The recommendation at the end of the Validated Job Pipeline tranche was stale:
the proposed design is the already-completed spawn-selected fixed-context
experiment, and its narrower descendants have also completed broad gates.

Retain one test-only generated-source guard proving that ordinary Able `spawn`
syntax does not silently select the rejected experimental ABI. Retain no
compiler, bridge, generated-runtime, VM, canonical-stdlib, benchmark,
reference, language, or WASM performance change.

## Historical identity

The 2026-07-14 spawn-selected gate did exactly what the stale recommendation
proposed: it scanned statically loaded user packages for language-level `spawn`,
excluded dormant canonical-stdlib declarations, enabled the fixed-pointer
context ABI only for matching programs, and left serial programs on the
compatibility ABI.

That generic candidate improved Channel Rollup by 9.1% and Future Pipeline by
14.5%, but regressed the unlike Mutex Ledger application by 10.0% in matched
three-run verifier-backed cohorts. Mutex Await Journal was also 1.4% slower.
It was reverted. A Mutex-, channel-, callback-, or task-shape exception would
be benchmark/application specialization rather than a language-level solution.

Earlier and adjacent variants close the remaining unchanged descendants:

- program-wide fixed context regressed N-Body by 54.7% across matched three-run
  cohorts because package-context rebinding allocated in dense static calls;
- allocation-free payload-only package linkage repaired N-Body but regressed
  K-Nucleotide by 16.6% across six retained runs and increased GC;
- payload, package-linkage, and compatibility-boundary variants passed semantic
  gates but failed the broad application gate.

The new Validated Job Pipeline profile placing 94.16% cumulative CPU under
`bridge.currentGID` / `runtime.Stack` confirms the old owner. It does not
invalidate the completed candidate gates: it is another concurrency
application exercising the same predicted path, not a new mechanism or a
contradictory guard result.

## Active-code audit and retained guard

The active generator still enables the context ABI only when
`Options.ExperimentalExecutionContext` is explicitly true. Both `ablec` and
`able build` expose that state only through
`--experimental-execution-context`, whose default is false.

`TestCompilerSpawnDoesNotSelectExperimentalExecutionContextByDefault` now
compiles an ordinary program containing `spawn` and requires:

- no generated `__able_execution_context` type;
- no `_ctx` main entry;
- no `__able_spawn_context` call; and
- the ordinary compiled main and `__able_spawn` compatibility path.

The existing explicit-opt-in test continues to require the inverse generated
surface. This protects the historical decision without deleting the useful
experimental implementation or turning a rejected candidate into a default.

## Verification

Focused generated-source checks passed in 0.123 seconds. The race-built nested
spawn executable passed in 12.714 seconds. Spawn/await, fairness/cancellation,
and nested native-context fixture parity passed in 9.893 seconds. Concurrent
same-environment, distinct-environment, goroutine-local, and nested restoration
bridge tests passed in 0.039 seconds. Every individual test command remained
below one minute.

No timing rerun was performed. Production sources, applications, references,
verifiers, stdlib, and execution contracts are unchanged. The active evidence
policy explicitly says that an old recommendation is not an invalidation
trigger, and repeating the unchanged candidate would only measure workstation
noise against a conclusive prior rejection.

## Next recommendation

Add one real, non-concurrent binary event-log decoder application before
selecting another implementation candidate.

Why: all 21 implementation closures are current, the 89-row frontier has zero
actionable groups, and another concurrency application will predictably land
on the rejected goroutine-identity path. A binary record workload is unlike the
current concurrency, regex, numeric-loop, and line-oriented text applications;
it can provide honest cross-family evidence for byte access, integer decoding,
nominal records, interface/error handling, maps, and file I/O without inventing
new syntax or a named-container compiler rule.

What it entails: define one deterministic binary fixture and public output
verifier; implement source-equivalent Able, Go, Python, and Ruby decoders with
the same validation/checksum algorithm; keep tree-walker correctness below the
one-minute cap; add it to the catalog only after exact parity; and retain two
five-process averaged cohorts per selected lane. Admit a compiler or VM profile
candidate only if one concrete non-parent leaf is also CPU-material in two
existing unlike application families. Update canonical `able-stdlib` only for
a reusable specified API gap. Do not begin WASM work.
