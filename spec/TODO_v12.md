# Able v12 Open Work

This file is a forward-only index of unresolved v12 language decisions and
selected implementation gaps. Completed work belongs in `LOG.md`, `v12/LOG.md`,
the dated records under `v12/docs/perf-baselines/`, and the relevant design
notes.

An item listed here is not implementation authorization by itself. Select work
through `PLAN.md`, cite the canonical specification rule or obtain the missing
language decision, and add cross-runtime coverage where behavior is observable.

## Currently selected implementation gaps

- None.

The mode-aware performance selector also currently selects no compiled or
bytecode owner. Do not infer a performance workstream from historical status
notes or residual helper names.

## Unresolved specification decisions

The following markers still describe genuine language-design choices. Each
needs a maintainer decision before implementation work is selected:

- shared-data race and ownership guidance for concurrency.

Resolve one item at a time in `spec/full_spec_v12.md`. The same tranche must
add parser/AST coverage when syntax changes and tree-walker, bytecode,
typechecker, compiler, and fixture coverage wherever the decision affects
those surfaces.

## Closed tracked surfaces

- Compiler AOT correctness gaps: none tracked.
- Canonical-stdlib externalization gaps: none tracked.
- Regex engine/API gaps: none tracked. Unsupported regex options remain
  specified errors, not silently ignored features.
- Parser/AST coverage: complete for the current v12 syntax surface; see
  `v12/design/parser-ast-coverage.md`.
- Typechecker implementation queue: empty pending a specification decision;
  see `v12/design/typechecker-plan.md`.
- Compiler-native performance queue: empty pending a genuine mode-aware
  invalidation and a general owner shared by unlike programs.
- Canonical-spec contradictions identified by the 2026-07-31 open-work audit:
  none tracked.

## Active implementation authorities

- Forward roadmap and performance admission:
  `PLAN.md`
- Language contract:
  `spec/full_spec_v12.md`
- Compiler lowering contract:
  `v12/design/compiler-go-lowering-spec.md`
- Compiler lowering process:
  `v12/design/compiler-go-lowering-plan.md`
- Native-carrier guardrails:
  `v12/design/compiler-native-lowering-guardrails.md`
- Generic specialization guardrails:
  `v12/design/compiler-monomorphization.md`
- Explicit compiled control-flow contract:
  `v12/design/compiler-no-panic-flow-control.md`
- Union carrier contract:
  `v12/design/compiler-union-abi.md`
- Historical container staging record:
  `v12/design/monomorphized-container-abi.md`

## Maintenance rule

Add an item only when the canonical spec or a reproduced implementation
failure identifies a concrete unresolved requirement. Remove it when the
decision and implementation land. Preserve completed measurements and
rationale in the logs, dated evidence, or design record rather than appending
them here.
