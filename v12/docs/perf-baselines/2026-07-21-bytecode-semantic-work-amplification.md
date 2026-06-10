# Bytecode semantic-work amplification audit

## Decision

**no-go-current-six-semantic-candidate**.

Do not build a semantic optimization from these six rows. Next test the only concrete two-family survivor, the direct Array semantic boundary, against Matrix Multiply as a third unlike application and numeric/text guards. Admit work only if one exact Array storage or member operation is material in all three and its end-to-end savings are large enough; otherwise close that boundary and return to the compiled residual frontier.

The audit distinguishes a repeated broad parent from one exact operation. Admission requires the same concrete mechanism to be materially amplified in at least three unlike application families.

## Logical-work normalization

| Application | Logical unit | Units | Semantic ops / unit | Call ops / unit | Allocations / unit | Bytes / unit | Dominant semantic category |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `fixed_width_128` | hot loop iteration | 2,000,000 | 20.613 | 2.110 | 15.956 | 674.2 | `control-return-scope-error` |
| `distance_field` | distance update | 2,000,000 | 23.000 | 3.000 | 13.000 | 184.0 | `scalar-numeric-conversion` |
| `concurrent_event_routing` | routed task | 4,096 | 2,966.453 | 444.424 | 688.609 | 70,641.4 | `control-return-scope-error` |
| `word_frequency` | nonempty word | 17,979 | 544.775 | 87.847 | 34.811 | 3,019.8 | `control-return-scope-error` |
| `array_slice_window` | window element | 288,000 | 20.424 | 3.251 | 1.442 | 59.0 | `scalar-numeric-conversion` |
| `reverse_complement` | sequence base | 2,000,000 | 21.184 | 4.034 | 2.035 | 133.3 | `control-return-scope-error` |

The six processes execute 157,421,139 semantic operations after 89,338,836 transport operations. Their measured mains allocate 2,343,825,960 bytes in 65,842,215 allocations. These totals are comparison parents only; neither semantic operations nor allocations have uniform cost.

## Reference-boundary reconciliation

- **`fixed_width_128`:** Python tuples and Ruby arrays implement the same two-word add/subtract/compare algorithm; this row has no bulk-library shortcut that explains the whole gap.
- **`distance_field`:** All three programs perform the same scalar recurrence and call a native hypot implementation once per update.
- **`concurrent_event_routing`:** Python and Ruby use native String split, queues, threads, and built-in maps/arrays; Able performs equivalent work through language-level Result, HashMap, Channel, and worker code.
- **`word_frequency`:** Python and Ruby use native String split and built-in hash tables; Able uses canonical stdlib split, iteration, Result, and HashMap implementations.
- **`array_slice_window`:** Python and Ruby copy each window with a native slice operation, then perform the weighted element loop at language level; Able uses canonical Array.slice and language-level iteration.
- **`reverse_complement`:** Python bytes.translate plus reverse slicing and Ruby String.tr plus reverse are native bulk operations; Able explicitly reads, translates, appends, and wraps each byte.

The native-boundary mismatch is real in four programs, but it is not one operation: concurrency primitives, splitting/hashing, slice copying, and byte translation/reversal have different contracts. Fixed Width and Distance Field also remain far from target without a comparable bulk shortcut, so native library use cannot explain the corpus-wide gap by itself.

## Mechanism admission

| Mechanism | Kind | Unlike families | Material families | Eligible | Disposition |
| --- | --- | ---: | ---: | --- | --- |
| `new-local-binding-store` | exact-opcode | 4 | 2 | no | `rejected-insufficient-amplified-breadth` |
| `named-call-return` | exact-opcode-family | 4 | 2 | no | `rejected-closed-and-insufficient-amplified-breadth` |
| `array-slot-member-dispatch` | exact-opcode-family | 2 | 2 | no | `rejected-insufficient-unlike-breadth` |
| `typed-result-pattern-control` | exact-opcode-family | 2 | 2 | no | `rejected-insufficient-unlike-breadth` |
| `native-bulk-library-boundary` | aggregate-boundary | 4 | 4 | no | `rejected-not-one-mechanism` |
| `allocation-volume` | aggregate-runtime-parent | 6 | 6 | no | `rejected-not-one-mechanism` |

Reasons:

- **`new-local-binding-store`:** StoreSlotNew is about two operations per logical unit in Fixed Width and Reverse Complement, but 538.8 per routed task and 103.4 per word in the two stdlib-heavy programs. Only two unlike families establish amplification, and their language-level pipelines differ.
- **`named-call-return`:** CallName and Return recur, but numeric rows perform expected per-iteration calls while the two amplified rows traverse different stdlib paths. Direct call, frame, continuation, and return candidates have already failed broad gates.
- **`array-slot-member-dispatch`:** CallMemberArraySlot and ArrayReadSlot are material in two unlike applications, below the three-family threshold; the existing direct Array-slot path reports no inline misses.
- **`typed-result-pattern-control`:** Typed-pattern and nullable Result control is large in Event Routing and Word Frequency but does not recur materially in a third unlike row.
- **`native-bulk-library-boundary`:** The reference interpreters cross native boundaries, but the exact operations split among queues/threads, String split and hash lookup, Array slicing, and byte translation/reversal. Treating all native library calls as one optimization would hide distinct semantics.
- **`allocation-volume`:** Allocation is broad, but exact owner reports divide among wide nominal values, float materialization, concurrency environments, String/HashMap state, Array slicing, and byte Array conversion/snapshots.

## Reproduction

Opcode counts came from fresh, bounded stats processes and are deterministic diagnostics, not wall-time measurements. Allocation counts come from two independent measured-main processes in the cited refresh reports. Performance decisions continue to use repeated-process arithmetic means.

```sh
v12/bench_semantic_work_amplification \
  --json-out v12/docs/perf-baselines/2026-07-21-bytecode-semantic-work-amplification.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-bytecode-semantic-work-amplification.md
python3 v12/bench_semantic_work_amplification_test.py
```
