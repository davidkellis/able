# Generic Interface Registration Reconciliation

## Decision

Retain a shared compiled-thunk registration repair. Compiled metadata names
the canonical interface `able.kernel.Clone`, while the loaded generic
`impl Clone for Array T` can retain the public type-alias spelling `Clone`.
Registration compared those strings literally and rejected the implementation
before the compiled program entered `main`.

The compiler fixture `06_12_28_stdlib_fs_lines` now passes. No Array,
filesystem, benchmark, or nominal-type special case was added.

## Root cause and repair

The failure was:

`no matching impl method for able.kernel.Clone.clone target=Array<T>`

Diagnostic reduction proved that the target type, empty interface arguments,
constraints, method name, and parameter types aligned. The candidate was
discarded solely because its stored interface name was `Clone` and compiled
registration supplied `able.kernel.Clone`.

`RegisterCompiledImplMethodOverload` now normalizes both identities through
the interpreter's type-alias expansion and canonical interface registry before
comparison. This preserves package identity: an explicit negative regression
proves that an unrelated `other.Marker` interface is not merged with an
unqualified `Marker` implementation merely because their final components
match.

## Correctness verification

- Focused unit regressions pass for:
  - canonical compiled identity matching a stored generic interface alias;
  - unrelated same-short-name interfaces remaining distinct.
- Compiled `06_12_28_stdlib_fs_lines` passes with its complete expected output
  in 29.641 seconds.
- Compiled controls pass for:
  - generic alias substitution;
  - canonical Array helpers;
  - Array slicing;
  - user-defined operator/Clone interfaces;
  - StringBuilder iterator cloning and specialized Array self returns;
  - Grapheme cloning retaining native Array clone lowering; and
  - primitive Clone and interface lookup/cache behavior.
- Diff hygiene passes and both changed Go files remain below one thousand
  lines.

The v12 spec and canonical stdlib did not change: this reconciles the compiler
bootstrap with the existing nominal interface contract.

## Repeated startup performance guard

Compiled thunk registration occurs once during application bootstrap, which is
included in external benchmark wall time. Four unlike compiled applications
therefore received fresh five-process verifier-backed runs. Their baseline is
the immediately preceding promoted bounded-lines cohort, with identical Able
application and canonical stdlib sources.

| Application | Baseline (s) | Candidate (s) | Change |
| --- | ---: | ---: | ---: |
| Document Audit | 0.096 | 0.094 | -2.1% |
| Dependency Plan | 0.080 | 0.082 | +2.5% |
| Option/Result Config | 0.200 | 0.198 | -1.0% |
| Channel Rollup | 1.200 | 1.194 | -0.5% |

Document Audit had the same first-process workstation outlier in a second
five-run batch; its combined ten-process mean remained 0.094 seconds. The
other rows were stable. This is neutral startup performance around the existing
measurement resolution, with no broad regression signal. The current full
scorecard was not relabeled; the focused candidate reports retain every sample
and verifier result.

## Next recommendation

Collect fresh bounded compiled and bytecode profiles for Regex Suffix Audit,
Regex Set Audit, and Regex Stream Audit, then advance a candidate only if the
same canonical NFA leaf repeats across all three.

Why: source equivalence has ruled out the eager-input mismatch, yet this family
still misses by roughly 45x-84x compiled and 107x-279x bytecode, with Regex
Suffix bytecode still timing out. All three exercise the reusable regex engine,
so a shared transition, epsilon-closure, capture, or allocation repair would
benefit real regex programs rather than one benchmark. The numeric families,
by contrast, still mix canonical checked/nominal obligations with much thinner
foreign representations.

What it entails: produce one-process CPU and allocation profiles under the
normal memory/timeout guards; reconcile hot paths by source function and
operation rather than parent dispatcher; select only a leaf present in all
three profiles; add focused semantic regex controls; and require repeated
five-process A/B results across all three applications plus unrelated text and
non-regex controls before retaining code.
