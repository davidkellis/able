# Composite-interface self-pattern contract retained

Date: 2026-07-30

## Decision

Retain the canonical composite-interface self-pattern contract and its shared
typechecker enforcement.

A composite's admitted target set must be a subset of every explicit base
interface's admitted target set after substituting the base reference's
generic arguments. An implicit composite may contain only implicit bases. An
explicit composite may contain implicit bases, which adopt the composite
pattern for that composition. A missing `for` clause and `for Self` are both
implicit.

This is one structural language/typechecker rule. It adds no tree-walker,
bytecode VM, compiler generator, runtime carrier, bridge, named-container,
non-primitive nominal, benchmark, application, canonical stdlib, dependency,
or WASM special case.

## Reproduced defect

Before the change, all three programs passed `able check`:

- an implicit composite combining an explicit `for T` base and an implicit
  base;
- an explicit `for T` composite combining those same bases; and
- `interface StringComposite for String = IntegerOnly` where `IntegerOnly`
  was declared `for i32`.

The third relationship is impossible: a legal `StringComposite` target cannot
satisfy the base's `i32`-only contract. The first is under-specified because
the omitted composite clause silently inherits an explicit restriction.

After the change, the explicit compatible form remains valid. The
under-specified and incompatible forms fail with source-attributed diagnostics
that name the composite, base, and effective pattern or patterns.

## Implementation

`pkg/typechecker/composite_interface_self_patterns.go` validates composites
after the declaration refresh pass. This makes the rule independent of source
declaration order and reuses the existing structural implementation
self-pattern matcher.

Base generic arguments are taken from the resolved interface application, not
only from source syntax. This preserves correct substitution through generic
type aliases such as an alias that maps `T` to `Base<Array T>`.

The parser and AST already preserve the required pattern and base arguments.
Valid programs continue through the existing shared nominal interface
pipeline; invalid programs stop before interpreter execution or native adapter
and dictionary synthesis.

The canonical specification is §10.1.5 of `spec/full_spec_v12.md`, with design
rationale in `v12/design/composite-interface-self-patterns.md`.

## Coverage

Added focused guards for:

- parser preservation of an explicit generic composite pattern and generic
  base arguments;
- compatible explicit plus implicit bases;
- direct generic base substitution;
- generic interface-alias substitution;
- rejection of an implicit composite with an explicit base; and
- rejection of incompatible explicit patterns.

Added the source-attributed AST error fixture
`errors/composite_interface_self_pattern_mismatch`.

Expanded `exec/10_01_interface_defaults_composites` so one value satisfies an
explicit composite formed from an explicit and an implicit base. Its methods
execute through tree-walker, bytecode, parity, and strict fallback-free
compiled modes.

## Verification

Passed:

- focused parser and typechecker guards;
- complete `pkg/typechecker`;
- complete `pkg/parser`, `pkg/typechecker`, `pkg/interpreter`, and
  `cmd/ablec` package run;
- AST fixture export integrity and typechecker baseline replay;
- tree-walker and bytecode execution of
  `10_01_interface_defaults_composites`;
- tree-walker/bytecode parity for that fixture;
- strict compiled execution with `RequireNoFallbacks`;
- the standalone compiler regression that happened to be active when the
  aggregate package timeout fired;
- exec coverage: 272 seeded, zero planned, 273 fixture directories;
- external scoreboard evidence: 132 rows, five Able/reference samples each;
- frontier: zero actionable groups; and
- selection, execution-contract, preserved-report, and scorecard-refresh
  protocol tests.

The unbatched short compiler package reached its existing ten-minute package
timeout while
`TestCompilerNoFallbacksVectorStringCompiledBuildKeepsNativeStringReceiver`
had run for only two seconds. Replayed alone, that test passed in 6.652
seconds. The tranche's strict composite compiler fixture passed in 3.028
seconds. This is aggregate batching pressure, not an individual-test or
semantic failure.

## Performance evidence state

The scorecard rows and retained measurements are unchanged, but the
performance evidence ledger correctly refuses `--check`: the v12 spec is a
shared production scope and its content changed.

A fresh report written outside the repository selects all 23 closures:

- 11 bytecode closures;
- 11 compiled closures; and
- the cross-family architecture closure.

All 23 have the single reason `scope-content-drift:v12-spec`. There is no
compiler, runtime, VM, stdlib, benchmark-source, reference-source, scorecard,
or measurement drift. The canonical ledger was deliberately not advanced and
old measurements were not relabeled current.

## Worktree and cleanup

The pre-tranche worktree contained 44 paths: the prior ten-path non-WASM
handoff and the unchanged 34-path deferred WASM boundary. The index was empty.
All task builds and probes used a disk-backed workspace under `/var/tmp`.

No existing dirty or untracked path was reset, reverted, overwritten, staged,
or removed. The deferred WASM boundary was not touched. The task-owned
workspace is removed after final verification. The final fully expanded
worktree contains 57 paths: the unchanged 34-path deferred WASM boundary and
the reviewed 23-path non-WASM handoff. The index remains empty.

The guarded project cleanup removed the task-created 335 MiB Go cache and its
empty temp directory. Final verification then removed the exact 2.0 GiB
disk-backed task workspace. The final cleanup preview reports no generated
project artifacts.

## Next

Run a post-contract performance-evidence reconciliation tranche.

Why: all 23 closures are formally selected solely because the canonical v12
language contract changed. Production execution code and benchmark sources
did not change, but the project must not claim current performance evidence
until the affected scope is reviewed and measured.

What it entails: prove the benchmark corpus contains no newly invalid
composite relationship; recapture strict native-boundary and dependency
guards; run repeated balanced Able/reference measurements for the affected
compiled and bytecode closure families; advance only reviewed closure
snapshots; then rerun the ledger, scoreboard, frontier, and selector gates.
Retain no performance implementation change unless the refreshed evidence
exposes one material owner shared by at least three unlike applications.

Why it matters: this restores trustworthy evidence for the core goal—native
compiled lowering with minimal boundary crossings—without using a spec-only
change to waive measurement discipline or manufacture an optimization.
