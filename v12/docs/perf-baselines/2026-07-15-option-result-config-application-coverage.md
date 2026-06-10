# Option/Result Configuration Application Coverage — 2026-07-15

## Scope

`option-result-config` is a portable deployment-capacity reconciliation
application. It audits 1,024 services across 24 passes, resolves a
service-local capacity through a regional fallback, validates the resolved
capacity, and substitutes a deterministic safe value for absent or invalid
configuration. Its verified result is `1024:18221610432`.

Able expresses the ordinary data flow with generic named-union members:
`Option.or_else`, `Option.map`, `Option.ok_or_else`, `Result.and_then`,
`Result.unwrap_or_else`, and `Result.is_ok`. Go uses value/valid pairs and
Python/Ruby use `None`/`nil`; each reference performs the same resolution and
checksum calculation. This is an application-shaped configuration workflow,
not a synthetic combinator loop.

## Compiled-binary repair

The first standalone compiled run exposed that generic named-union member
dispatch relied on the in-process interpreter used by fixture harnesses. A
standalone binary deliberately omits that bootstrap. Generated calls now carry
the checked generic-union target and first consult the generated compiled
method table. That supports every generic named union with generated methods;
it does not name or encode `Option`, `Result`, or any container.

`TestCompilerStandaloneGenericNamedUnionMethods` builds and executes a binary
using a user-defined `Choice T` generic union with chained `or_else`, `map`,
and `unwrap_or`, preventing this no-bootstrap regression from being hidden by
the stdlib aliases.

## Verification

- Go, Python, and Ruby references each pass `verify.rb`.
- `bench_perf` executes the canonical Able entry in compiled, bytecode, and
  tree-walker modes; every one passes the same Ruby verifier.
- `just bench-catalog-check` reports 33 portable applications, 77 fixtures,
  and 110 corpus programs.
- `bench_bytecode_audit --suite corpus-full` reports 110 programs, 420
  functions, and 20,379 instructions.

No pinned comparison, profile, scorecard, VM optimization, compiler
optimization, or stdlib performance claim was made. A later candidate still
needs one concrete shared leaf across this and at least two unlike
verifier-backed applications after the quiet-host gate.
