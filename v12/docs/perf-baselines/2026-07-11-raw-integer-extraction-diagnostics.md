# Raw Integer Extraction Diagnostic Tranche

## Decision

Keep no runtime change. The raw-integer extractor is shared, but its dominant
callers and carrier mix do not support a profitable direct-carrier store
shortcut. A candidate that bypassed the generic extraction and range switches
for already-raw i32/i64 store values helped only numeric Array map and
regressed the text, iterator, and non-Array controls. It was removed together
with the temporary diagnostic instrumentation.

No compiler lowering, stdlib source, source-shape opcode, or named-container
rule changed in this tranche.

## Method

One temporary, opt-in probe counted every input carrier to
`bytecodeRawIntegerValueInfo(...)` and sampled its direct callers once per
1,024 calls. It was run once per source program with canonical `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The source and harness
instrumentation were removed before the candidate and restored-control
benchmarks. JSON evidence remains under
`.profiles/20260711-raw-integer-diagnostics/`.

| Program | Extractor calls | raw i32/i64 carriers | boxed integers | unsupported inputs | `tryStoreRawIntegerSlotValue` caller samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| String split/join | 5,085,706 | 1,673,000 (32.9%) | 2,772,045 (54.5%) | 640,661 (12.6%) | 2,116 / 4,966 (42.6%) |
| Iterator collect | 920,714 | 465,655 (50.6%) | 213,010 (23.1%) | 180,049 (19.6%) | 361 / 899 (40.2%) |
| Numeric Array map | 696,113 | 611,945 (87.9%) | 84,161 (12.1%) | 7 | 115 / 679 (16.9%) |
| String builder control | 2,056,034 | 911,615 (44.3%) | 1,144,339 (55.7%) | 64 | 813 / 2,007 (40.5%) |

The raw i32/i64 carriers are common, and `tryStoreRawIntegerSlotValue(...)` is
the largest common direct caller except in numeric Array map. But a substantial
fraction of text and iterator extraction remains boxed or speculative failure,
so adding a direct pre-switch to that store helper makes the nonmatching path
more expensive.

## Rejected candidate

The candidate handled `bytecodeRawI32SlotValue`,
`*bytecodeRawI32StackCell`, `bytecodeRawI64ResultValue`, and
`*bytecodeRawI64SlotCell` directly in
`tryStoreRawIntegerSlotValue(...)`. It used the exact existing storage helpers,
so focused raw-slot semantics passed. The same-session bounded A/B, however,
rejected it:

| Program | Candidate | Restored control | Result |
| --- | ---: | ---: | --- |
| String split/join (5x) | 1,195,853,648 ns/op | 1,034,551,572 ns/op | 15.6% slower |
| Iterator collect (5x) | 289,153,479 ns/op | 252,269,809 ns/op | 14.6% slower |
| Numeric Array map (20x) | 64,242,781 ns/op | 65,511,649 ns/op | 1.9% faster only here |
| String builder (5x) | 264,519,015 ns/op | 249,954,889 ns/op | 5.8% slower |

Allocation shapes stayed effectively unchanged. The direct branches merely
moved type switching ahead of the existing compact extractor; they do not
remove a common enough operation. Keep the simpler generic helper.

## Next direction

Profile the callers of the repeated Go `mapaccess2_faststr` cost rather than
retrying raw-carrier branching. It appears in text, iterator, and numeric
profiles, but may represent different interpreter maps (method/cache/type
lookups). First classify the owning map lookups and their hit/miss behavior
with diagnostics outside timing. Only consider a candidate if the same
interpreter map and validation rule recur across workloads; otherwise keep no
change.
