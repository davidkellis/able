# Compiled split-receiver method ABI

## Status

Active for generated compiled instance methods. The ordinary runtime callable
ABI remains the compatibility boundary; a second optional entry accepts the
receiver separately for statically resolved calls.

## Problem

The compiled generic-union direct-call path knew both the receiver and the
compiled method entry, but the registered wrapper still accepted one combined
`[]runtime.Value`. The caller therefore allocated a new
`[receiver] + explicit arguments` slice for every call. Exact allocation
profiles reproduced that compiler-owned leaf in Binary Event Log,
Option/Result Config, Manifest Normalization, and Policy Record Dispatch.

## Contract audit

Replacing the ordinary callable ABI would be incorrect. These paths require
the combined argument representation and remain unchanged:

- bound method values and native bound method values;
- partial application and later argument merging;
- UFCS, callable values, interface dictionaries, and overload groups;
- package/runtime symbols and interpreter thunks;
- dynamically resolved methods and the generic bridge fallback.

Generated compiled instance methods instead expose two compatible entries:

1. The ordinary wrapper accepts `receiver + explicit arguments`, checks that a
   receiver exists, and delegates to the direct core.
2. The direct core accepts `receiver` and `explicit arguments` separately. It
   performs the same optional-last-argument normalization, arity check,
   package-environment swap, value conversion, compiled-body call, control to
   error conversion, mutable-struct receiver writeback, and result conversion.

The compiled-method registry stores the direct core as optional metadata on a
single instance-method entry. Static generic-union dispatch uses it when
present. Builtins, overload groups, and any entry without it retain the old
receiver-prefix fallback. This is a representation rule for every generated
compiled instance method; it contains no named nominal, container, stdlib, or
benchmark branch.

## Evidence

All four owner means improved across 20 verifier-backed processes per variant,
by 1.0% to 3.5%. Five exact candidate allocation runs per owner reduced main
objects by 0.6% to 10.8% and bytes by 0.3% to 9.4%. Five unlike guards also
had non-regressive means; the three volatile guards received 16 samples per
variant. Focused generated-source and executable tests cover receiver order,
optional parameters, bound/dynamic compatibility paths, nominal methods,
control conversion, and fallback semantics.

The durable measurements and every retained timing sample are in
`../docs/perf-baselines/2026-07-22-compiled-split-receiver-method-abi-gate.md`
and its companion JSON.

## Lazy native environment refinement

The direct helper no longer constructs `NativeCallContext` on the successful
path. The semantic audit found that direct generated method bodies and raised-
control conversion consume `Env` but not `State`; runtime data remains
reachable through that environment. The optional direct ABI therefore passes
`*runtime.Environment`. It constructs a context containing that environment
only inside the uncommon control-to-error branch. The ordinary compatibility
wrapper extracts `Env` through one shared inlinable helper, while dynamic and
fallback paths preserve the original full native context.

This synchronization-free form removes another exact per-call object in four
unlike owners. The 488-process final gate has one significant owner
improvement, neutral remaining owners, and no statistically confirmed guard
regression. Full evidence is in
`../docs/perf-baselines/2026-07-22-compiled-lazy-native-environment-direct-abi-gate.md`.

Do not retry `sync.Pool`, separate environment/state parameters, by-value
environment/state transport, or the closed program-wide/spawn-scoped execution
context designs without a new invalidating trigger. Refresh profiles before
selecting another generated-call change.
