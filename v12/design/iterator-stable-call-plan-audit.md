# Generic stable `Iterator.next()` call-plan audit

Date: 2026-07-14

## Decision

Do not prototype a new bytecode `Iterator.next()` stable-call plan. The
existing `CallMemberNext` path already has the maximal safe specialization for
ordinary runtime iterators and canonical string iterators. The remaining path
is an ordinary zero-argument member call, and the checks a new plan would need
to preserve are the same ones already paid by the generic member cache.

This is not a rejection of iterator performance as a concern. It is a
rejection of a shortcut that would either change Able member-call semantics or
merely reproduce the existing cache with another data structure.

## Language constraints

`next()` is not a privileged language operation. A generator produces an
`Iterator T`, but advancing it is still a normal member call that must return a
value, `IteratorEnd`, or propagate an exception. See spec §6.7 and §14.1.
Normal member-call resolution gives a callable receiver field precedence over
the callable pool (spec §7.4.3 and §9.4). Therefore a cached implementation of
`next` must not survive adding or replacing a callable field named `next`.

Interface values are dynamic dispatch lenses. Their dictionary supplies the
method selected at upcast; runtime interface values also have a per-value
`Methods` overlay which can shadow a shared dictionary. Dynamic imports are
late-bound and runtime definitions are atomic at symbol-replacement granularity
(spec §6.10). Bytecode may not assume that the same call site, type name, or
interface dictionary remains valid without the corresponding invalidation
check.

The observable result protocol is also part of the boundary: `IteratorEnd`,
`nil`, errors, and raw primitive carriers must retain their current behavior.

## Current VM coverage

`bytecodeOpCallMemberNext` already tries, in order:

1. a generic runtime `IteratorValue` raw-result fast path;
2. canonical string byte/character iterator fast paths;
3. the normal inline member-method cache; and
4. general member dispatch.

The remaining cached path validates receiver identity, method-cache version,
lexical scope/binding state, interface dictionary/template identity, and then
executes the already-resolved call. These are deliberate semantic guards, not
accidental lookup work.

Existing tests cover the relevant boundary behavior:

- `TestBytecodeVM_CallMemberNextCacheHonorsCallableFieldShadow`
- `TestBytecodeVM_InterfaceMemberMethodCacheRejectsShadowedDictionaryEntry`
- `TestBytecodeVM_CallMemberInterfaceIteratorNextFastPathRequiresIteratorNative`
- `TestBytecodeVM_CallMemberIteratorNextRawI64KeepsBytecodeCarrier`
- `TestBytecodeVM_CallMemberGenericIteratorNextFastPathWrapsNilValue`

The tree-walker continues to use the same runtime member-resolution semantics;
there is no separate language rule to relax for bytecode.

## Evidence and feasibility

The fresh coverage profiles show cached `next()` work in two unlike lazy
pipelines:

| Application | Cached `next()` cumulative | Related member-cache cumulative |
| --- | ---: | ---: |
| Document Audit | 8.28% | 15.53% |
| Lexical Rollup | 5.97% | 10.91% |

That repetition makes the protocol worth auditing, but it does not make every
shortcut valid. The candidate outcomes are:

| Candidate | Result |
| --- | --- |
| Call a cached function unconditionally | Incorrect: bypasses callable-field precedence and per-interface overrides. |
| Trust a shared interface dictionary pointer | Incorrect: a value can have an owned overlay; current tests deliberately mutate one. |
| Recheck field/dictionary/lexical state at every call | Correct, but equivalent to the existing cache validation and does not remove its material cost. |
| Add universal field and interface-dictionary revision contracts | A language/runtime-wide redesign affecting both interpreters, dynamic code, and compiler boundaries; unjustified by this 6–8% slice alone. |

The profile also attributes substantial time below the cached-next wrapper to
resolved-call execution and return handling. Replacing the wrapper alone could
not credibly close the 6–50x reference gaps, and Word Frequency—currently the
largest coverage miss—does not use this iterator leaf.

## Reopen condition

Reopen this only with a broader semantic proposal, such as an explicitly
immutable interface-dictionary contract plus revisioned callable-field storage
that all runtimes, dynamic bridges, and compiler lowering can honor. Such a
proposal must first specify invalidation semantics, preserve the tests above in
both interpreters, and prove a broad win against the verified coverage,
text, numeric, map, and JSON guards. It must not encode knowledge of a named
stdlib iterator, lazy sequence, or benchmark.
