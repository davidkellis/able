# Bytecode Array-slot value-role census

Date: 2026-07-22

Decision: retain the opt-in operand-role diagnostic and retain no runtime
candidate.

## Result

The previously shared small-integer-value to `CallMemberArraySlot` shape is
not index extraction. All 2,241,366 dynamic loads reconcile to a value
operand: 2,241,334 feed `push` and 32 feed `write_slot`. There are no receiver,
index, unknown-role, or dropped-program counts.

| Application | Exact shape | Operation and role | Direct route |
| --- | ---: | --- | ---: |
| Concurrent Event Routing | 136,192 | `push` value | 136,192 `array_push_u8_mono_fast` |
| Reverse Complement | 2,000,049 | 2,000,017 `push` values; 32 `write_slot` values | 4,033,607 total `array_push_u8_mono_fast` |
| Word Frequency | 105,125 | `push` value | 105,125 `array_push_u8_mono_fast` |

The Event Routing and Word Frequency exact counts match the direct `u8` push
dispatch exactly. Reverse Complement has additional direct `u8` pushes from
constants and other carriers; every traced push uses the same monomorphic
route. The shared exact sites are typed `u8`, and the fast path extracts that
small scalar once before calling the primitive array-store append. It does not
materialize the value through the generic integer path.

## Clean CPU attribution

Each application received a separate one-process runtime-only CPU profile
under the same 55-second, one-process memory guardrails. The profiles are
attribution evidence, not timing comparisons.

| Application | Total samples | `execArrayPushMemberFast` cumulative |
| --- | ---: | ---: |
| Concurrent Event Routing | 2.79s | no samples |
| Reverse Complement | 3.29s | 0.23s (6.99%) |
| Word Frequency | 1.23s | 0.01s (0.81%) |

Reverse Complement's cost is mainly the synchronized monomorphic array-store
append. That cost is not CPU-material in the other two families. A lock/cache
change would therefore fail the same-material-wall admission rule and would
also require a broader concurrency/aliasing audit than this operand-transport
shape can justify.

## Decision

No candidate was written. An index shortcut would optimize a path these loads
never execute. A raw-value transport shortcut would duplicate an already
direct `u8` kernel path. No compiler, tree-walker, stdlib, benchmark, language,
or WASM code changed.

## Next recommendation

Refresh the broad six-application clean CPU owner matrix and select the largest
exact, non-closed VM leaf that is CPU-material in at least three unlike
families.

Why: carrier, consumer, operation, role, trace, and profile evidence now close
this Array-slot lead without a general candidate. The next tranche should
avoid descending further into Reverse Complement's array-store lock because
that owner is absent or immaterial in the other guards. It entails bounded
one-process profiles for Fixed Width 128, Distance Field, Concurrent Event
Routing, Word Frequency, Array Slice Window, and Reverse Complement; exact
source/call-path attribution for repeated owners; and a repeated averaged A/B
only if one non-closed leaf clears the three-family breadth gate.
