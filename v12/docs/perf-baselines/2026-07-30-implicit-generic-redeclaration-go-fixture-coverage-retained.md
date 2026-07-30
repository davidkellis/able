# Implicit generic redeclaration Go fixture coverage retained

Date: 2026-07-30

## Decision

Retain the removal of the stale `go` target exclusion from
`v12/fixtures/ast/errors/implicit_generic_redeclaration/manifest.json`.

Able v12 section 7.1.5 requires a compile-time diagnostic when a function body
declares a type with the same name as an inferred generic parameter. The Go
typechecker already implements and unit-tests that rule, but the canonical
source fixture still declared `skipTargets: ["go"]`. The active Go fixture
harness therefore returned before parsing, checking, or comparing that case.

Removing the stale exclusion makes the shared source-level fixture enforce the
existing language rule. No language or runtime behavior changed.

## Reproduction

The canonical source is:

```able
fn wrap(value: T) -> T {
  struct T {
    value: i32
  }
  value;
}
```

The fixture expectation is:

```text
typechecker: v12/fixtures/ast/errors/implicit_generic_redeclaration/source.able:2:3 typechecker: cannot redeclare inferred type parameter 'T' inside fn wrap (inferred at v12/fixtures/ast/errors/implicit_generic_redeclaration/source.able:1:16)
```

Before the correction, `cmd/fixture --describe` reported `skipped: true` for
the active Go target. After the correction it reports `skipped: false`, strict
typechecking, a nil execution result, and exactly the expected diagnostic with
both source locations.

The retained file identities are:

- manifest SHA-256:
  `d4e23cb70285385cc78194ac25f81bb1d42d29f2a4148eeaef39ca795de3374d`;
- source SHA-256:
  `9b6052de0ec8bfb61ee9c417ebb3429d60e3141250f7dea6a9cf119c7d0cce3a`;
  and
- AST module SHA-256:
  `022f02eb71f3c3283c0aadd343112c466fa2818ca81f0d6c1fb34ea1977d9c66`.

No active v12 fixture now excludes the Go target. The only remaining
`skipTargets` entry is the historical `ts` exclusion on `String_methods`; the
TypeScript interpreter is not part of the active v12 toolchain.

## Focused verification

- Exact strict fixture subtest:
  `TestFixtureParityStringLiteral/errors/implicit_generic_redeclaration`
  passes in 0.066 seconds.
- Both focused inferred-generic redeclaration checker tests pass in 0.002
  seconds.
- All `errors/*` source fixtures pass in 0.133 seconds.
- The complete typechecker package passes in 0.025 seconds.
- `cmd/fixture` tests pass in 0.070 seconds.
- Direct fixture CLI replay reports the expected diagnostic and
  `skipped: false`.

The first full-run attempt intentionally set a 55-second package-process
timeout. It stopped cumulative `cmd/able` and interpreter packages even though
the active tests shown were only 22 and 23 seconds. This was not a correctness
failure. The runner was repeated with its designed package timeout.

## Broad verification

The default `./run_all_tests.sh` v12 suite passes:

- execution-coverage index;
- 130-row external scoreboard and five-sample evidence;
- feature/application coverage;
- selection and execution contracts;
- threshold, cleanup, and embedded-kernel checks;
- standalone parser Go binding;
- all non-compiler Go packages in short mode;
- all 34 bounded compiler short-mode batches; and
- the complete bytecode fixture pass.

The large interpreter, bytecode, and compiler-batch durations are cumulative
package/group times, not one fixture reproduction. The changed focused fixture
remains well below one second.

The performance ledger remains current with 23 closures, zero invalidations,
and an empty selector. Fixture metadata is outside the compiler, runtime,
interpreter-production, canonical-stdlib, benchmark, and v12-spec performance
scopes, so no performance remeasurement is required.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
parser, canonical stdlib, benchmark, language, dependency, or WASM source
changed. No performance implementation was selected.

All Go build state used disk-backed `/var/tmp`. The exact 2,660,460 KiB
workspace was removed after verification; no task-owned artifact remains.

The machine-readable companion is
`2026-07-30-implicit-generic-redeclaration-go-fixture-coverage-retained.json`.

## Next

Add a general fixture-inventory guard that rejects an active `go` target
exclusion unless it is explicitly classified and allowlisted.

Why: a stale manifest flag silently prevented the only active runtime fixture
harness from enforcing a canonical v12 diagnostic.

What it entails: extend repository-owned fixture validation to enumerate
`skipTargets`, fail on unclassified `go` exclusions, retain the retired `ts`
entry as historical metadata, and add focused positive/negative tests. Keep
the guard independent of any one fixture name or language feature.

Why it matters: future Go coverage cannot disappear through an unnoticed
manifest edit, and genuine target exceptions remain reviewable rather than
implicit.
