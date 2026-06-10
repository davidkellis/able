# Standard-library protocol contract audit

**Status:** reconciled with the canonical external stdlib on 2026-07-14

## Purpose

Section 14 of the v12 specification formerly called the standard-library
interfaces “Conceptual / TBD,” even though they are active language contracts.
That wording invited new work to treat established syntax protocols as
unsettled.

The canonical source is the external `able-stdlib` checkout:

- `src/core/interfaces.able` defines the operator, callable, indexing, error,
  `Default`, and `Extend` protocols.
- `src/core/iteration.able` defines `Iterator`, `Iterable`, and `Range`.
- `able.kernel` provides the primitive comparison, hashing, display, clone,
  and default implementations imported by the canonical interfaces module.

The v12 specification remains the semantic authority. This audit records that
its Section 14 now names those protocols as a defined language contract and
links its signatures and lifecycle semantics to their canonical source.

## Iterator lifecycle cross-check

The canonical `Iterator T` protocol includes:

```able
fn next(self: Self) -> T | IteratorEnd;
fn close(self: Self) -> void {}
```

`close()` is idempotent. Resource-owning iterators override it, and a closed
iterator must return `IteratorEnd` without resuming work. Default `each`,
`filter`, `map`, `filter_map`, and `collect` close their captured input on
completion, early exit, and errors. This agrees with the required-runtime
protocol in §6.12 and with the public source contract.

`exec/06_12_27_stdlib_iterator_close` exercises that lifecycle in the
tree-walker, bytecode VM, and compiled no-fallback modes. It is part of the
default compiler fixture set, so it guards direct idempotent close, generator
close, `for` break, and all default combinators on every compiler fixture run.

## Scope

This is a specification and handoff correction only. It changes no parser,
AST, interpreter, compiler, stdlib source, benchmark, or runtime behavior.
Unresolved language-design choices such as block-comment syntax, variance
syntax, and Array slice ownership remain explicit specification design work;
this audit does not choose their semantics.
