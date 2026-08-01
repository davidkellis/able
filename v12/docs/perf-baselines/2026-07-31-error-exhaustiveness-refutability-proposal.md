# Open-set Error exhaustiveness/refutability proposal

Date: 2026-07-31

## Outcome

**Proposal complete; no execution or language rule changed.**

The recommended v12 rule is a conservative, positive static proof:

- only unguarded clauses contribute coverage;
- wildcard and bare binding patterns cover the complete remaining domain;
- typed binding and typed discard patterns have identical coverage;
- an unguarded `case error: Error` or `case _: Error` covers the whole open
  `Error` component;
- concrete implementations never collectively exhaust an open interface;
- closed unions are covered member by member using semantic subtyping;
- ordinary partial matches remain valid and retain their runtime
  `Non-exhaustive match` diagnostic;
- partial rescues remain valid and re-propagate the original raised value; and
- refutable declaration/assignment patterns remain valid and preserve their
  runtime mismatch result.

The complete proposal is
`v12/design/error-exhaustiveness-refutability-proposal.md`. It requires a
maintainer decision before specification or implementation work begins.

## Current behavior

The typechecker checks match/rescue subjects, patterns, guards, clause bodies,
and branch joins. It does not currently prove exhaustiveness or refutability.

Tree-walker and bytecode ordinary matches report `Non-exhaustive match` when no
clause accepts the subject. Compiled matches emit the corresponding dynamic
guard.

Tree-walker, bytecode, and compiled rescue evaluation all re-propagate the
original raised value when no handler matches. This makes a partial rescue a
useful, intentional construct rather than an invalid expression.

`or {}` structurally handles its failure alternative and `!` propagates it;
neither has a pattern list requiring an exhaustiveness proof.

## Compatibility census

The audit parsed all `.able` sources under `v12/fixtures`, `v12/examples`, and
`../able-stdlib/src` with the current v12 parser:

- files: 759;
- parse failures: zero;
- match expressions: 915; and
- rescue expressions: 71.

Match syntax:

- 351 expressions contain an unguarded wildcard or bare binding;
- 102 contain an unguarded typed `Error` pattern;
- 84 rely on typed `Error` coverage without a universal;
- 480 contain neither syntactic form, including many closed-union or
  otherwise type-specific matches; and
- four expressions contain guarded universal patterns, comprising six such
  clauses.

Rescue syntax:

- 26 expressions contain an unguarded universal;
- 42 contain typed `Error` coverage without a universal; and
- three contain only concrete/refutable handlers.

Pattern-node totals:

| Context | Array | Identifier | Literal | Struct | Typed | Wildcard |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| match | 4 | 110 | 405 | 356 | 872 | 300 |
| rescue | 0 | 16 | 1 | 2 | 44 | 10 |

This census is intentionally syntactic. It does not claim that the 480
no-catch-all matches are semantically non-exhaustive; many use closed union
members or other patterns that need resolved type information.

Two canonical fixtures explicitly require a runtime non-exhaustive-match
diagnostic:

- `exec/08_01_match_guards_exhaustiveness`; and
- `exec/04_06_04_union_guarded_match_exhaustive_diag`.

The canonical stdlib and compiler regression corpus commonly bind the whole
error through `case error: Error`. Requiring only the discard spelling
`case _: Error` would misclassify equivalent programs.

## Alternatives

The proposal records four choices:

1. conservative positive proof plus runtime fallback — recommended;
2. hard compile-time exhaustiveness — incompatible with current fixtures;
3. runtime-only behavior — compatible but provides no lowering proof; and
4. closed-world enumeration of visible `Error` implementations — unsound and
   rejected.

## Focused verification

The unchanged behavior passed:

```text
go test ./pkg/typechecker \
  -run 'Test(Match|Rescue|OrElse|Propagation)' -count=1
ok .../pkg/typechecker 0.003s

ABLE_TYPECHECK_FIXTURES=strict go test ./pkg/interpreter \
  -run 'TestExecFixtureParity/(04_06_04_union_guarded_match_exhaustive_diag|08_01_match_guards_exhaustiveness|11_03_rescue_ensure|06_12_30_stdlib_option_result)$' \
  -count=1
ok .../pkg/interpreter 0.338s

ABLE_COMPILER_EXEC_FIXTURES='04_06_04_union_guarded_match_exhaustive_diag,08_01_match_guards_exhaustiveness,11_03_rescue_ensure,06_12_30_stdlib_option_result' \
ABLE_COMPILER_REQUIRE_NO_FALLBACKS=1 \
go test ./pkg/compiler -run '^TestCompilerExecFixtures$' -count=1
ok .../pkg/compiler 4.798s
```

All task builds and parser census artifacts used a disk-backed `/var/tmp`
workspace.

The complete `./run_all_tests.sh` release gate then passed in 12:45.48 at
1,837,420 KiB peak RSS:

- 273 seeded exec fixtures, zero planned fixtures, and 274 fixture
  directories;
- 132 current scorecard rows and zero actionable frontier groups;
- 23 current performance closures and zero invalidations;
- the five-node architecture evidence chain current;
- parser, tree-walker, bytecode, and cross-mode fixture coverage;
- all 844 compiler tests;
- canonical compiler outlier: 13.986 seconds; and
- slowest bounded compiler batch: 45.516 seconds.

The final ledger check reports 23 current closures and zero invalidations; the
mode-aware selection command emits no compiled or bytecode owner. The exact
2,700,016 KiB task workspace was removed after verification. No `/tmp/able-*`,
`/var/tmp/able-*`, or new Python cache remains.

## Scope

No specification, parser, AST, typechecker, interpreter, VM, compiler,
generated runtime, stdlib, dependency, fixture, benchmark, frozen workspace,
or WASM behavior changed.

This proposal does not settle operand admissibility for `raise`, the exact
runtime carrier/catchability of a non-exhaustive-match diagnostic, handler
effect inference, or constant guard evaluation.

## Next

Obtain a maintainer decision on option A, the conservative positive-proof rule.

Why: implementation before selection could either reject currently valid
partial programs or remove a runtime path without a canonical soundness rule.

What it entails: review the open-interface, guarded-clause, ordinary-match,
partial-rescue, and destructuring decisions in the proposal. If option A is
selected, canonicalize it in the v12 spec, implement one shared semantic fact
analysis, add cross-runtime fixtures, and let compiled lowering consume only
positive proofs.

Why it matters: the shared proof is what can safely eliminate redundant
match/rescue fallback branches around native `Error`/`Result` carriers while
keeping static compiled programs on ordinary Go control flow.
