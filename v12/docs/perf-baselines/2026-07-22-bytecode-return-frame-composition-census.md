# Bytecode return-frame composition census

Date: 2026-07-22

## Decision

Retain no VM, runtime, compiler, stdlib, benchmark, fixture, language, or WASM
change from this tranche. The stats-only census completed, but no immutable
return-frame shape passed the predeclared admission rule.

All four owners overwhelmingly use full inline frames. That fact does not
identify removable work: the same full-frame, program-switch, and slot-frame
release mechanics also dominate all three unrelated controls. The narrower
raw-cell, coercion, slotless, and self-fast shapes either lack breadth across
three owners or are materially selected by a control. `already_balanced` and
self-fast-minimal returns never occurred.

## Method

A temporary, release-disabled observer counted semantic return shapes inside
`finishInlineReturn`. It was enabled independently with
`ABLE_BYTECODE_RETURN_CENSUS=1`, so the all-opcode bytecode statistics observer
remained disabled. Counts were also grouped by instruction origin and shape.

The fixed census binary SHA-256 was
`a18ecbdd4f7ed4240cd803d50d2df4059f79241f8bef32d21e4734079d610721`.
Each workload loaded and typechecked normally, ran one untimed warmup, and
measured one `main()` call with source-root-only loading under
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 59-second process cap. All
seven processes passed. Timing is diagnostic only because this tranche makes
its decision from exact event composition rather than census-instrumented wall
time.

| Role | Workload | ns/op | Bytes/op | Objects/op |
| --- | --- | ---: | ---: | ---: |
| Owner | Binary Event Log | 7,133,198,446 | 465,565,496 | 8,361,435 |
| Owner | Option/Result Config | 791,509,286 | 67,914,992 | 1,167,817 |
| Owner | Manifest Normalization | 1,189,342,526 | 63,998,264 | 916,218 |
| Owner | Policy Record Dispatch | 6,902,429,806 | 132,690,480 | 1,402,534 |
| Control | String split/join | 1,103,358,441 | 50,592,592 | 577,767 |
| Control | Iterator collect | 434,247,350 | 8,701,856 | 193,006 |
| Control | Numeric Array map | 74,749,894 | 875,320 | 281 |

## Aggregate composition

Percentages are shares of all observed inline returns. A return can contribute
to several columns: for example, a full return can also have raw input, need
generic coercion, switch programs, and release a slot frame.

| Workload | Returns | Full | Self-fast | Slotless | Raw in | Raw out | Fast/no coercion | Generic coercion | Program switch | Slot release |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 946,230 | 100% | 0% | 0% | 45.36% | 33.24% | 58.01% | 41.99% | 100% | 100% |
| Option/Result Config | 293,784 | 100% | 0% | 8.37% | 41.57% | 31.08% | 10.55% | 81.07% | 100% | 91.63% |
| Manifest Normalization | 596,485 | 100% | 0% | 1.20% | 0% | 0% | 1.44% | 97.36% | 100% | 99.31% |
| Policy Record Dispatch | 2,534,369 | 100.00% | 0.003% | 0.20% | 1.01% | 0% | 85.41% | 14.39% | 100.00% | 99.92% |
| String split/join | 616,587 | 100% | 0% | 0% | 0.001% | 0% | 0.98% | 99.02% | 100% | 100% |
| Iterator collect | 164,046 | 100% | 0% | 0.007% | 37.80% | 37.80% | 0.005% | 99.99% | 100% | 100% |
| Numeric Array map | 36,020 | 100% | 0% | 0% | 99.96% | 0% | 99.96% | 0.039% | 100% | 100% |

Self-fast-minimal and already-balanced counts were zero in every workload.
Policy's 66 self-fast returns were 0.003% of its total and occurred in no other
owner. Slotless coercion exactly matched the slotless counts wherever it
occurred.

## Admission reconciliation

The declared rule required one immutable semantic shape to dominate at least
three unlike owners without materially selecting the controls.

- Full frames, program switching, and slot-frame release have owner breadth,
  but are essentially universal in the controls too. These counters describe
  required call restoration, not a proven redundant branch or state copy.
- Generic coercion is substantial in every owner, but it is also 99.02% of
  split/join and 99.99% of iterator collect. A coercion shortcut would select
  the exact controls that previously exposed mixed return changes.
- Fast/no-coercion is substantial in Binary and Policy, not three owners, and
  it is 99.96% of numeric Array map.
- Raw input has breadth in Binary, Option, and Policy, but is negligible in
  Policy and dominates numeric Array map; raw output is substantial in only
  Binary and Option and also occurs throughout iterator collect.
- Slotless returns are meaningful only in Option and are small elsewhere.
  Self-fast is negligible, while self-fast-minimal and already-balanced paths
  are absent.

No candidate was admitted. In particular, the zero already-balanced count
confirms that reordering its guard cannot remove a frequent path, and the lack
of a dominant self-fast/minimal shape rules out another small frame-specialized
return tweak. The temporary observer and its tests were removed after the
census.

## Verification and restoration

The restored focused semantic suite passes:

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*Return|TestBytecodeVM_SlotlessInline|TestBytecodeVM_MinimalReturn|TestBytecodeVM_InlineReturnRestoresCallerActiveLookupCaches|TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm|TestBytecodeGenericUnionCallCacheInvalidatesMemberChanges' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 4.232s
```

The three inspected interpreter files remain gofmt-clean and under 1,000
lines. Raw census artifacts are cleanup-eligible under
`/tmp/able-return-frame-census-20260722`.

## Next recommendation

Return to the broad bytecode feature frontier and build a cross-workload exact-
leaf matrix from a bounded set of unlike applications rather than continuing
to tune return handling.

Why: the return/type-resolution corridor has now had profiles, shape censuses,
and several general candidates; its remaining shapes either fail owner breadth
or are equally dominant in controls. Continuing locally would optimize a
common workload shape without evidence that it explains the language-wide
performance gap. The next useful evidence is an exact interpreter-owned leaf
that recurs across unrelated feature families.

What it entails: choose representative text/file, numeric, collection/
iterator, nominal/union, and concurrency applications from the current suite;
run one bounded clean CPU profile for each under the same one-process
guardrails; compare exact flat leaves and their caller closures; and admit a
candidate only when the same mechanism is material in at least three unrelated
applications with appropriate controls. Start with raw integer extraction,
map/hash lookup, and type-match descendants only if the exact caller—not merely
the cumulative family—repeats. Continue to defer WASM.
