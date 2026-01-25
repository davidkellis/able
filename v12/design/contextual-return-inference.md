# Contextual Return-Type Inference for Call Chains

## Problem
We want generic call sites like `iter.collect()` to infer their type arguments from
surrounding context when arguments alone are insufficient. In particular, a call
that appears in a return position should be able to use the enclosing function's
return type to infer the call's generic parameters, mirroring Rust-style
"collect" inference.

## Goals
- Allow call-site inference to use the expected type from return positions
  (explicit `return` and implicit final-expression returns).
- Apply the expected type to generic method calls and call chains where the
  outermost expression is a call.
- Keep inference deterministic: ambiguous or underconstrained cases should
  produce diagnostics that request explicit type arguments.
- Maintain parity between the TS and Go typecheckers.

## Non-Goals
- Full bidirectional inference across arbitrary expressions (only the call
  expression itself participates in expected-type inference).
- Inferring type arguments when the expected type is a type constructor with
  unbound parameters.

## Spec Alignment
- §7.1.4 clarifies that expected types from return positions participate in
  call-site inference.
- §6.1.8 already requires contextual inference for struct literals; the same
  expected-type mechanism should be shared with calls.

## Typechecker Plan

### TypeScript
- Extend function-call inference to accept an optional expected return type and
  apply it during generic substitution (not just post-hoc refinement).
- Use this expected type when checking explicit `return` statements and when
  typechecking the final expression in a function body with a declared return
  type.
- Ensure method-call resolution and overload selection incorporate the inferred
  generic substitutions so the return type is computed with the expected args.

### Go
- `checkExpressionWithExpectedType` already injects expected return types for
  explicit `return` statements; extend function-body checking so implicit
  returns also use `checkExpressionWithExpectedType`.
- Ensure `instantiateFunctionCall` keeps using expected-return inference for
  generics and that the resulting return type uses those substitutions.

## Test Plan
- Add an exec fixture that returns a `HashSet u8` built via `collect()` without
  explicit type arguments.
- Add typechecker unit tests (TS + Go) that assert generic inference in return
  contexts for call chains and method calls.
- Run `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict` after
  implementation.

## Open Questions
- Should expected-type inference flow into non-call expressions (e.g., blocks or
  match expressions that yield a call)? For now, only the outermost call uses
  expected types.
