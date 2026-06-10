# Compiler Union ABI: Current Guardrails and Historical Record

## Status and authority

This is the source-backed record for the compiler's union, nullable, and
result carriers. It is not an active bring-up plan. Able semantics remain in
spec §4.6 and §11; the direct AST-to-Go compiler, its strict no-fallback tests,
and `compiler-native-lowering-guardrails.md` define current implementation
responsibility.

Only a language type shape—not a named struct, stdlib container, or benchmark
application—may select a carrier. `Array` has its separate language-syntax /
kernel boundary rules; every other nominal member uses the shared nominal
carrier machinery.

## Current static carrier contract

The `TypeMapper` first normalizes aliases and package context. It then uses
the shared native-union synthesizer for representable members:

| Able type shape | Static compiled representation |
| --- | --- |
| `?T` | a nil-capable native carrier or a generated typed nullable carrier |
| `!T` / normalized `Error | T` | `runtime.ErrorValue` when it is the whole value, otherwise a generated native union carrier |
| closed or multi-member union | a generated `__able_union_*` interface plus typed variant wrappers |
| fully bound interface member | its generated `__able_iface_*` carrier inside the union |
| `Error` member | `runtime.ErrorValue` or the shared concrete-error carrier path |

The union synthesizer flattens representable nested union/result members,
deduplicates them by normalized package-aware type identity, and reuses one
carrier for an equivalent normalized shape. It creates `*_wrap_*`, `*_as_*`,
`*_from_value`, `*_try_from_value`, and `*_to_value` helpers as needed. The
names are generated implementation details; shared shape synthesis and direct
static representation are the architectural requirements.

A native union is materialized only when every member can be represented by a
native carrier. The compiler must not create a partially native union with a
broad `runtime.Value` or `any` member merely to keep one branch fast. A
non-representable value stays at an explicit dynamic/ABI boundary, or strict
no-fallback compilation rejects the static path.

## Dispatch, matching, and boundaries

Static union matching uses generated extraction helpers and native control
flow. Nullable matches use their native nil/payload operations. Static union,
result, and typed-error paths must not regress to `bridge.MatchType(...)`,
`__able_try_cast(...)`, or general interpreter member dispatch.

When a value crosses an explicit dynamic feature, callback, or host ABI edge,
the generated conversion helpers translate between the native carrier and
`runtime.Value` at that edge. They do not justify retaining runtime values in
static locals, parameters, returns, or direct calls. A result returned from a
dynamic edge is converted back immediately when its static type is
representable.

The same rule applies to interface members: a fully bound static interface
branch remains a native interface carrier, then dispatches directly. It is not
an interpreter dictionary or an excuse to add a union rule for a particular
interface, collection, or user type.

## Change guardrails

- Keep carrier selection in shared normalization, carrier synthesis, and
  boundary helpers—not emitter-local checks for named nominal types.
- Keep static paths direct and prove them with strict no-fallback,
  generated-source, and behavioral tests appropriate to the changed shape.
- Keep dynamic conversion visible and adjacent to its language/host boundary.
- Treat a new carrier optimization as a performance candidate only when the
  same material non-nominal leaf recurs across at least three unlike verified
  applications. No union-specific optimization is selected by this record.

## Historical staging record

The former “first slice,” target ABI, bring-up order, implementation targets,
and non-goals described work that is now complete. The completed chronology is:

1. Native nullable scalar and nil-capable carrier lowering.
2. Closed-union carriers, wrappers, boundary conversions, and native pattern
   extraction.
3. Native result/error carrier handling and propagation.
4. Multi-member unions, generic aliases, nested representable unions, and
   package-aware/shadowed nominal resolution.
5. Native interface/callable member carriers, shared join inference, typed
   patterns, and explicit dynamic callback conversions.

The detailed dated implementation evidence remains in
`compiler-native-lowering.md`; it is historical context, not a reopened
union-lowering queue. New work in this area requires a new specified semantic
gap or a broadly selected performance leaf, not completion of an old stage.

## Focused evidence

The compiler tests cover closed union params/returns and matches, multi-member
unions, generic `Option`/`Result` aliases, `Error` result propagation, native
interface branches, and dynamic callback conversion. They prove both sides of
the boundary: static code retains generated carriers, while explicit dynamic
paths use checked conversion helpers.
