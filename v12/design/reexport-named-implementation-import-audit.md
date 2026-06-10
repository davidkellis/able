# Re-export and Named-Implementation Import Audit

Status: completed named-implementation and source re-export tranche, 2026-07-14
Authority: spec §10.2.5, §10.3.3, §13.4, and §13.5; the shared AST remains the
implementation contract.

This audit records package-import and explicit source-re-export semantics. It
does not add a nominal type, container, benchmark, or source-shape special case.

## Implemented named-implementation namespace contract

`Name = impl Interface for Target { ... }` parses to one
`ast.ImplementationDefinition` with `ImplName = Name`. A public named impl is
an importable explicit-dispatch namespace; it is never added to implicit impl
selection. Its methods remain selected only with function-position syntax such
as `Name.method(value, ...)`.

| Situation | Required outcome | Go implementation coverage |
| --- | --- | --- |
| `import pkg.{Fancy}` | Binds public named implementation `Fancy` for explicit dispatch. | `ProgramChecker` package surface; static runtime imports; existing cross-package named-impl fixture. |
| Two static imports bind the same name and either binding is a named implementation | Diagnostic/error; neither import may silently overwrite or skip the other. | Selective and wildcard import paths in checker and runtime. |
| `import second.{Fancy::OtherFancy}` | Valid distinct binding; both namespaces remain callable. | Checker and runtime regression tests. |
| A local named impl collides with an imported binding, or a local declaration collides with an imported named impl | Diagnostic; the imported binding is not replaced. | Declaration collector regression test. |
| `private` named impl | Not importable outside its package; public/package import maps omit it. | Checker private-symbol summary and runtime private-value filtering. |
| `dynimport` | Retains its documented dynamic shadowing behavior. | Deliberately unchanged: this static import rule does not redefine dynamic packages. |

The diagnostic directs callers to a selector alias. This is the exact recovery
mechanism specified by §13.5 and is preferable to load-order-dependent winner
selection. The collision check is based on the shared
`ImplementationNamespace` representation, not a package name or a named
stdlib type.

## Existing identity/re-export behavior

Aliases are already resolved to their canonical target for method-set and impl
lookup. The shared AST fixtures cover alias-defined method/impl propagation and
ambiguity, while the exec fixture `04_07_06_alias_reexport_methods_impls`
proves that methods and implementations attached through an alias remain
available through the underlying/re-exported type. The kernel compatibility
re-export map is also deliberately identity-preserving.

## Explicit source re-export contract

The grammar and shared AST now model `export Name;` and `export * from
package;` as `ast.ExportStatement`. The tree-sitter corpus establishes both
forms before the module parser maps them into `Module.Exports`; the AST fixture
`imports/source_reexport_syntax` records the JSON contract.

Named exports publish an existing visible binding. A locally private binding,
or a selectively imported private binding, is rejected and never added to the
wrapper's package surface. Wildcard exports publish only the source package's
public surface. The loader treats a wildcard export source as a dependency, so
the checker sees it in dependency order and package-import cycle checks remain
unchanged.

Both Go interpreters publish the original runtime value after evaluating the
module body. Thus a re-exported function, struct, union, interface, alias, or
named implementation remains the same definition/value; it is not a wrapper.
The compiler tracks the original static import binding, forwards the exact
runtime value when seeding no-bootstrap package maps, and resolves static calls
and nominal struct/interface lookup through re-export chains. No named
container or benchmark-specific lowering branch is involved.

## Evidence

- `pkg/parser`: named implementation source maps to `ImplementationDefinition`
  with its `ImplName`.
- `pkg/typechecker`: direct, wildcard, renamed, local, and private
  named-implementation import cases.
- `pkg/interpreter`: the same static import cases at runtime.
- `pkg/interpreter` exec fixture `10_05_interface_named_impl_defaults`: explicit
  named dispatch across packages in tree-walker and bytecode modes.
- `pkg/compiler`: named-namespace bootstrap and member-call lowering controls.
- `pkg/parser`: source `export` CST, AST, and generated fixture coverage.
- `pkg/typechecker`: named/wildcard package-surface propagation and private
  named-source rejection.
- `pkg/interpreter`: named and wildcard identity tests in tree-walker and
  bytecode modes.
- `pkg/compiler`: named/wildcard function forwarding plus nominal struct
  carrier forwarding without interpreter bootstrap.

No performance candidate follows from this semantics work. The active
cross-application performance gate remains unchanged.
