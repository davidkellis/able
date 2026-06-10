# RawValue boundary audit

## Decision

`runtime.RawValue` is not a safe compiler-to-runtime value carrier today. It
is an internal bytecode transport for unboxed integer and float payloads, not a
replacement for `runtime.Value`. Do not lower compiled Able functions,
interfaces, collections, structs, or returns through it without a language-wide
runtime-value design.

## Current representation

`RawValue` has three states:

- a materialized `runtime.Value`;
- an integer suffix plus an `int64` bit pattern; or
- a float suffix plus a `float64` payload.

It deliberately does not implement `runtime.Value`: its `Kind()` returns
`RawValueKind`, whereas every language value must return `runtime.Kind`. It
also has no unboxed bool, char, nil, nominal, or reference representation.

Unsigned `u64`, `u128`, and `usize` retain their full bit pattern while raw.
When materialized, a negative `int64` bit pattern becomes a `big.Int` with the
same unsigned value. That preserves integer width, but it is necessarily a
materialization and allocation boundary.

## Existing raw-only paths

The raw path is confined to the bytecode interpreter:

- Bytecode stack/slot cells carry selected integer and float results without
  boxing, then `bytecodeMaterializeRawValue(...)` boxes them for generic work.
- `NativeFunctionValue.RawImpl` receives `[]runtime.RawValue` and returns one;
  an ordinary native `Impl` receives only `[]runtime.Value`.
- `IteratorValue.NextRaw()` and generator yield can retain primitive payloads;
  `IteratorValue.Next()` materializes before exposing a regular value.
- Exact native-call results can append a raw value back to the bytecode stack.

Raw values materialize at every generic-language boundary: ordinary calls,
function returns, member/implementation dispatch, type matching and casts,
patterns, array/struct storage, native calls without `RawImpl`, and host
conversion. The tree-walker has no equivalent raw carrier. The compiler has no
`RawValue` references and its generated static functions and bridge helpers
communicate with the runtime through `runtime.Value`.

## Performance evidence

The compiled allocation profiles show `bridge.ToUint`/`ToInt` boxing is
material, but not a local shared lowering rule:

- K-Nucleotide spends about 55.7% of cumulative allocation bytes in those
  helpers while passing primitive keys and counts through raw map operations.
- Word-frequency's smaller `ToUint` allocation is from string UTF-8 and byte
  array conversion, not the map calls.
- The independent fixed-width-128 control also boxes heavily, including u64
  values above `MaxInt64`; its dominant cost is the required public UInt128
  conversion and BigInt semantics, not a container path.

This confirms that a fast path for any named collection, string helper, FASTA
input, or UInt128 would be benchmark-specific. It does not demonstrate a
semantically complete general raw boundary.

### 2026-07-12 source-aligned confirmation

Fresh compiled Word Frequency and K-Nucleotide profiles confirm the same
decision under the canonical external stdlib. `bridge.ToUint` is material in
both: 8.1% cumulative CPU and 20.5 MB across eight Word Frequency captures,
and 16.8% cumulative CPU and 1.075 GB across two K-Nucleotide captures.
That recurrence remains below different named-value contexts. Word Frequency
is dominated by `String_split` and `__able_hash_map_find_entry`; K-Nucleotide
is dominated by primitive conversion/allocation together with raw HashMap and
string-validation work. The JSON control is faster than fresh Go. This is
stronger evidence that the helper is telemetry for the existing general-value
boundary, not authorization for a HashMap, string-key, FASTA, small-count, or
`RawValue` compiler shortcut.

## Required proof before any future prototype

A representation proposal must first make the same carrier work for all of
these without a type-name branch:

1. Direct functions and closures: arguments, returns, recursion, and captured
   values.
2. Primitive widths: signed/unsigned casts, u64 high-bit values, u128 values,
   overflow, and BigInt promotion.
3. Nominal boundaries: struct fields, interfaces, unions, nullable values, and
   generic type parameters.
4. Collections and callbacks: arrays, maps, iterators, native calls, extern
   calls, and dynamic dispatch.
5. Tree-walker/bytecode/compiler parity through shared fixtures, including
   observable `Kind`, type-match, equality, and error behavior.

Until that proof exists, retain `RawValue` as a bytecode-local optimization and
keep compiled lowering on `runtime.Value`.

## 2026-07-12 feasibility conclusion

The cross-runtime feasibility tranche rejects a `RawValue` expansion or a
compiler-only carrier as the next performance change. `RawValue` cannot
implement `runtime.Value` without replacing its incompatible `Kind()` method,
and a local extension would still materialize at ordinary storage, call,
nominal, host, and future boundaries. A universal tagged value remains a
possible future runtime architecture, but it requires a specified carrier and
the complete tree-walker/bytecode/compiler proof matrix before any prototype.
The detailed candidate analysis and matrix are in
`v12/design/primitive-value-representation-feasibility.md`.
