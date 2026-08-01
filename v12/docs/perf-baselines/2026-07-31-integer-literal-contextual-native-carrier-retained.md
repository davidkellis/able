# Fixed integer default and contextual native carriers retained

Date: 2026-07-31

## Decision

**Retain a fixed, non-configurable unconstrained `i32` default and the general
contextual-literal corrections exposed while proving that contract.**

An unsuffixed integer literal with no resolving context has the v12 type
`i32`. A valid concrete numeric context may adopt an unsuffixed literal when
the value fits. A suffix fixes the literal's source type: ordinary permitted
widening remains available, but contextual adoption cannot silently retarget
the suffix.

The earlier `configurable/TBC` phrase was stale and has been removed. The
resolved TODO is also gone. The detailed diagnostic text now records the
already-retained lazy behavior for an unconstrained value outside `i32`: it is
still an `i32` literal and raises `OverflowError` if evaluated. This agrees
with the longstanding tree-walker, bytecode constant-validation, and strict
compiled behavior, including recoverable deliberately out-of-range literals.

## Proof findings and retained corrections

The proof found three real implementation gaps:

1. A typed binding could accept an explicitly suffixed integer whose source
   range was not assignable to the annotation. The shared checker now reports
   the explicit source and expected types. A new diagnostic fixture covers
   tree-walker, bytecode, and compiler preparation.
2. Generic inference treated an unsuffixed literal's provisional `i32` label
   as final even after an explicit type argument or expected return bound the
   parameter to another numeric type. Pending literals now adopt a fitting
   resolved binding, while true incompatibilities remain diagnostics.
3. The compiler's generic-specialization repair pass could overwrite a
   resolved call-site type argument with the argument literal's provisional
   `i32` carrier. Resolved call-site bindings now outrank repair facts.
   Checked interpreter execution likewise consumes the resolved numeric
   inference fact before either engine materializes an unsuffixed literal.

These are general primitive and generic rules. They contain no application,
benchmark, container, nominal-type, stdlib-family, or source-name condition.
No runtime representation, canonical stdlib, dependency, benchmark,
reference implementation, frozen workspace, or WASM behavior changed.

## Native-carrier proof

A single control covered:

- unconstrained `42` as native Go `int32`;
- `200` in a `u8` context as native Go `uint8`;
- `3_000_000_000` in an `i64` context as native Go `int64`;
- `42_u64` as native Go `uint64`;
- inferred `identity(7)` as native Go `int32`; and
- explicit and expected-result generic `i64` calls with values above the
  `i32` range as native Go `int64`.

Tree-walker, bytecode, and strict no-fallback compiled execution produced the
same six values. Generated Go contained separate native `int32` and `int64`
generic specializations and no `runtime.Value`, `big.Int`, or bridge
conversion in the contextual generic call body.

The focused parser prefix/suffix corpus, complete typechecker package, lazy
bytecode overflow guard, wide explicit-literal compiler guard, contextual
generic compiler guard, and three literal fixture families all passed. The
fixture corpus now contains 281 seeded rows, zero planned rows, and 282
fixture directories.

The first complete gate exposed an over-broad compiler precedence rule: keeping
every checker-populated call-site type label after specialization repair broke
three imported/shadowed callable-alias guards. The retained rule now protects
only generic parameters reached by a direct unsuffixed integer literal.
Package-qualified nominal and callable facts continue to win for every other
argument. The three failing guards plus the contextual wide-integer controls
then passed three consecutive focused runs.

The clean complete gate passed in 636.54 seconds at 1,283,560 KiB peak RSS.
All 86 compiler batches passed; batch 76 was slowest at 44.557 seconds, and the
isolated canonical compiler test took 14.866 seconds. Canonical stdlib passed
in tree-walker mode in 19 seconds and bytecode mode in 15 seconds; the combined
gate took 36.16 seconds at 859,144 KiB peak RSS.

This correctness tranche did not claim a benchmark speedup and therefore did
not relabel any timing. Before the fix, the wide contextual programs failed;
there is no meaningful baseline runtime to compare. The relevant performance
closures were reconciled structurally after the complete release gates.

## Evidence reconciliation

The canonical spec SHA-256 moved from
`0965a0a48b49f5eaed9392d75b7dbf6e74965a04c2547286d39397dc0812bdcd`
to
`a7444934c9c4f334d33064b1a64cc91afbe4b02d3be86dc7a0947605c8124382`.
Its checked scope moved from
`5f3a8573e88d3ec612a05a429948fcb1bb63981cd7dd7f6983a83bcc71cf1095`
to
`b826ae4cb5ce7f808c1a09970bc94ef4e5887c3417a08246aaeaab2669cd5934`.

The compiler, bytecode, and shared-interpreter production scopes also moved.
The pre-reconciliation selector therefore invalidated all 23 closures for
the exact applicable scope reasons. This does not imply a benchmark
candidate: the source and evidence identities, benchmark measurements,
closure decisions, and target classifications are unchanged.

After reconciliation, all 23 closures are current and none is selected. The
ordered architecture chain is current across all five nodes, with no decision
change. The final compiler-production scope is
`03b8e875814cc407d30db52ed2ccaf32fb6ae9f26a955a8a41a172d37cf1ab37`.
The exact 3,243,448 KiB owned build workspace and regenerated extern cache were
removed from `/var/tmp` and the RAM-backed `/tmp`; no owned Able task directory
remains in either location.

## Next

Resolve the remaining §6.3.2 names and contracts for wrapping, saturating, and
checked primitive integer arithmetic.

Why: the language already fixes checked operators and native primitive
carriers, but the alternative arithmetic API is still marked `names TBD`.
That leaves performance-sensitive code without one canonical, portable way
to request non-raising arithmetic.

What it entails: inventory existing runtime, compiler, bytecode, and canonical
stdlib helper surfaces; compare their behavior for every fixed-width integer;
write a spec-first naming and result contract; then implement only the shared
primitive rule with cross-engine fixtures if the chosen API is not already
complete.

Why it matters: explicit wrapping, saturating, and optional checked operations
let optimized Able code express the same intent as native Go without changing
default overflow safety or introducing dynamic/boxed work. A single canonical
contract also prevents engine-specific helper names from becoming an ABI.
