# Able Test CLI and Framework Protocol: Active Contract

## Status and ownership

`able test` is an implemented user-facing Go CLI command. Its framework
protocol and the default `able.spec` framework live in the canonical external
stdlib (`../able-stdlib/src/test` and `../able-stdlib/src/spec.able`). The Go
CLI loads those modules and renders their discovery and execution events; it
does not define a second host-only testing model.

This is separate from Able implementation verification. AST/exec fixtures,
tree-walker/bytecode parity, compiler audits, and benchmark checks remain Go
test harnesses under `v12/fixtures`, `v12/interpreters/go`, and
`v12/run_all_tests.sh`. User test modules never replace or configure those
implementation fixtures.

## User-facing command

```text
able [--exec-mode=treewalker|bytecode] test [--compiled] [options] [targets...]
```

Without targets, discovery starts at the current directory. A target may be a
directory, test module, or ordinary source file (whose parent directory is
searched). Discovery is recursive for `.test.able` and `.spec.able` files,
skipping `quarantine`, `node_modules`, and `.git`. Test modules are loaded in
the same package/typechecking session as their production code; ordinary
`run`, `check`, and `build` retain their explicit `--with-tests` profile.

The default execution mode is the normal tree-walker mode. `--exec-mode` may
select bytecode. `--compiled` runs the discovered test plan through generated
Go; `--compiled --list` and `--compiled --dry-run` deliberately use the same
shared discovery path without compiling a runner.

Supported options are:

- Discovery/filter request: `--path`, `--exclude-path`, `--name`,
  `--exclude-name`, `--tag`, and `--exclude-tag`.
- Execution request: `--shuffle [seed]`, `--fail-fast`, `--parallel N`, and
  `--repeat N`.
- Output/control: `--list`, `--dry-run`, and
  `--format doc|progress|tap|json`.

`--list` emits the selected descriptors without execution. `--dry-run` is a
descriptor list with metadata. The CLI rejects unknown flags and invalid
positive counts. Success and an empty discovery both exit zero; test failures
and command/target errors exit one; loader, typechecking, framework, and
runtime/discovery failures exit two.

## Framework protocol

Test-module evaluation registers ordinary Able `Framework` values through
`able.test.registry`. The harness loads every registered framework, invokes
`discover` with a `DiscoveryRequest`, then groups the returned descriptors by
framework and invokes `run` with a `TestPlan`, `RunOptions`, and `Reporter`.
The framework owns its discovery semantics for the supplied path/name/tag
filters; the harness applies the registered-framework grouping and the shared
`focus` tag rule before execution.

The stable protocol types are defined in canonical source:

- `DiscoveryRequest` and `RunOptions` carry CLI filters and execution hints.
- `TestDescriptor` carries framework/id/module/display/location/tags/metadata.
- `TestPlan`, `Failure`, `TestEvent`, `Reporter`, and `Framework` define the
  shared plan, failure, event, and callback boundaries.

The default `able.spec` framework is registered by importing it. It provides
suite/example registration, focus/skip, hooks, matchers, and descriptor/event
emission through this protocol. Other frameworks can implement the same
interface; they must not require a compiler, VM, or named-container special
case.

The CLI renders doc and progress reporters through canonical stdlib helpers;
TAP and JSON are direct structured event streams. A framework error produces
exit code two, while a `case_failed` event produces exit code one.

## Verification boundary

The Go command implementation owns command parsing, target discovery,
module/typecheck setup, interpreted/compiled dispatch, and output decoding.
Canonical stdlib owns framework registration, the protocol data structures,
the default spec framework, filter-aware discovery, focus handling, and
reporter behavior. Changes on either side require CLI tests plus canonical
stdlib suites; they must not be smuggled in as benchmark or runtime work.

The checked Go tests cover empty discovery, filters and targets, interpreted
and compiled shared discovery, doc/progress/TAP/JSON output, descriptor
metadata/location preservation, skips/failures/framework errors, and compiled
stdlib suites. The command's wrappers forward the test flags unchanged.

## Deferred scope

There is no active work item for manifest test opt-outs, framework option
schemas, remote execution, artifact attachments, hierarchical plan transport,
or new CLI flags. The earlier proposed shapes and extensions are retained in
[the historical record](./testing-cli-protocol-historical.md); select any new
testing feature only from a concrete language/tooling requirement.
