# Compiled `String.bytes` native iterator — 2026-07-12

## Decision

Keep the compiler lowering for canonical primitive `String.bytes()`. On a
valid UTF-8 Go string whose byte length fits Able's `i32` iterator field, the
generated method now creates the monomorphized `__able_array_u8` directly from
`[]uint8(self)` and returns the existing `RawStringBytesIter` adapter.

The previous generated path converted the string through runtime values,
boxed every byte, rebuilt a generic array, converted it back to a
monomorphized `u8` array, and then validated it. The kept path validates with
`utf8.ValidString` and has the same length guard; invalid or oversized inputs
fall through to the existing Able implementation unchanged.

This is a primitive-language boundary only. It does not recognize a benchmark,
named container, user-defined `bytes` method, `StringBuilder`, or `Grapheme`.
It uses the existing iterator representation and therefore preserves iterator
and interface semantics. No canonical-stdlib source change is required.

## Generality gate

Bounded generated-main allocation profiles used canonical external
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Before the
change, all three independent public `String.bytes` callers had the same
material caller and leaf family: `String.bytes` -> `bridge.ToUint` ->
`__able_string_from_builtin_impl` -> UTF-8 decode.

| Workload | Baseline main allocations | Kept main allocations | Allocation reduction |
| --- | ---: | ---: | ---: |
| `ascii_lower_small` | 74,932,848 B / 1,553,129 | 19,679,240 B / 340,277 | 73.7% bytes / 78.1% objects |
| `byte_histogram_small` | 28,367,672 B / 545,495 | 14,403,320 B / 239,405 | 49.2% bytes / 56.1% objects |
| `md5_hex_small` | 32,034,088 B / 637,802 | 3,128,048 B / 24,469 | 90.2% bytes / 96.2% objects |

The costly conversion family disappears from the kept profiles. `String.bytes`
still allocates one owned `[]uint8` slice per iterator, which is required by
the existing raw iterator representation; it no longer allocates one runtime
value per byte or round-trips through a generic array.

## Compiled application gate

The control used three cold-process runs per application with the same one-core
and 1 GiB guardrail. Lower is better.

| Workload | Baseline real average | Kept real average | Change |
| --- | ---: | ---: | ---: |
| `ascii_lower_small` | 0.2767 s | 0.1633 s | 41.0% faster |
| `byte_histogram_small` | 0.1567 s | 0.1367 s | 12.8% faster |
| `md5_hex_small` | 0.2633 s | 0.1933 s | 26.6% faster |

Every run completed; each application's baseline and kept stdout SHA-256 was
identical. Average garbage collections also fell from 6.00 to 4.00, 4.00 to
3.33, and 5.00 to 3.00 respectively. The temporary JSON timing summaries and
generated source trees were removed after this record was written.

## Verification

- Focused compiler native-string, static iterable, and runtime-shim tests
  passed.
- Focused bytecode string-iterator tests passed, confirming that the unchanged
  reference iterator implementation remains correct.
- `git diff --check` passed.

## Next recommendation

Collect bounded CPU and allocation profiles for `StringBuilder` through
`string_builder_small`, `run_length_encode_small`, and
`byte_histogram_small`. After `String.bytes` is removed, builder append and
finish paths dominate the first two profile shapes. A new candidate still
requires one material builder leaf and caller across all three, rather than a
special case for construction-heavy benchmark setup.
