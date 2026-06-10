# Cross-mode exact-leaf selection reconciliation

Date: 2026-07-18

## Decision

Complete the K-Nucleotide, Fixed Width 128, Regex Set, Distance Field, and
Future Pipeline cross-mode selection pass and retain no compiler, bytecode VM,
runtime, canonical-stdlib, benchmark, fixture, or language change.

Current retained profiles already covered both modes for K-Nucleotide, Fixed
Width, and Distance Field, plus the compiled side of Regex Set and Future
Pipeline. Two fresh ordinary-process bytecode profiles filled the only useful
gaps. The resulting five-family set does not expose one previously untried
exact operation that is material in three unlike applications.

The bytecode profiles share dispatcher, load/stack, call, and return parents.
Their concrete children are map/text and integer work, checked wide-nominal
operations, Array/member and regex-NFA work, float geometry, and async numeric
work. Every shared representation below those parents is either immaterial or
belongs to a design already rejected by broad application timing. Generated
compiled owners diverge even earlier.

## Admission rule and evidence reuse

A candidate required the same removable language/runtime operation to be
material in at least three unlike applications. Broad Go runtime parents,
`runResumable`, generic call/return frames, related variants of one library
algorithm, and operations already rejected by repeated A/B gates did not
qualify.

No unchanged profile was recollected merely to produce new artifacts:

| Application | Current bytecode evidence | Current compiled evidence |
| --- | --- | --- |
| K-Nucleotide | post-quickening verified CPU profile | post-result/current-scorecard map, String, hash/equality and conversion attribution |
| Fixed Width 128 | post-quickening verified CPU profile | post-wide-carrier checked `UInt128`, extraction and residual nominal attribution |
| Distance Field | post-quickening verified CPU profile | retained generated numeric/sqrt attribution after imported raw-call lowering |
| Regex Set | **fresh profile in this tranche** | retained post-regex-carrier generated-main profiles and current scorecard attribution |
| Future Pipeline | **fresh profile in this tranche** | equal-CPU generated-main profile with `bridge.currentGID` / `runtime.Stack` attribution |

The later `write_all` change cannot affect Fixed Width, Regex Set, Distance
Field, or Future Pipeline. K-Nucleotide does not use the changed bulk-output
path. The retained source and executable evidence is therefore compatible
with the current product state.

## Fresh bytecode profiles

Both missing profiles used the ordinary bytecode process, canonical external
stdlib, CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, the catalog
executor/input/source-root contract, and a 55-second process cap. Each output
passed its public Ruby verifier.

| Application | Wall / CPU samples | Output verification | Material concrete descendants |
| --- | --- | --- | --- |
| Future Pipeline | 0.890 s / 0.670 s | `3db937fd...d98` | async task execution; `execBinary` 32.84%; arithmetic fast path 13.43%; integer extraction and store/branch children |
| Regex Set | 4.230 s / 4.110 s | `3d8f861a...da2` | `execCallMemberArraySlot` 26.52%; Array read path 9.98%; member access 9.00%; named-field lookup and regex call traffic |

The Future profile is necessarily coarse at 10 ms samples, but it is enough
to reject an overlap: its material bytecode descendant is numeric worker
arithmetic, not Regex Set's Array/member/NFA path. Both profiles use Go build
ID `90188e44e89b96ea07662facec614ecb166b4bb1`, exactly matching the retained
post-quickening K-Nucleotide, Fixed Width, and Distance Field profiles.

## Five-application bytecode intersection

At a 2% cumulative threshold, these exact interpreter symbols occur in all
five profiles:

| Exact symbol | K-Nucleotide | Fixed Width | Distance | Regex Set | Future Pipeline | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| `runResumable` cumulative | 95.94% | 87.57% | 95.93% | 88.08% | 80.60% | dispatcher parent; concrete opcodes differ |
| `execLoadSlotOpcode` cumulative | 4.23% | 5.73% | 5.00% | 7.30% | 2.99% | load/stack wrapper; carrier/order trials rejected |
| `appendSlotStackValueChecked` cumulative | 3.38% | 3.63% | 3.89% | 4.14% | 2.99% | same closed stack-carrier family |
| `execCallOpcode` cumulative | 30.23% | 49.16% | 44.63% | 35.52% | 7.46% | call parent over five different consumers |
| `finishInlineReturn` cumulative | 15.59% | 6.84% | 8.52% | 2.92% | 2.99% | return parent; guard/frame variants rejected |

This is stronger breadth evidence for the parents, but not for a removable
child. In particular:

- K-Nucleotide is led by bitwise/map/call/type work;
- Fixed Width is led by checked `UInt128` members and result construction;
- Distance Field is led by static/native float geometry and raw-float
  normalization;
- Regex Set is led by Array/member access plus regex-NFA calls; and
- Future Pipeline is led by integer arithmetic inside goroutine tasks.

`bytecodeRawIntegerValueInfo` repeats in K-Nucleotide, Regex Set, and Future,
but its carrier, metadata, and extraction alternatives have already failed
repeated collection, numeric, and control guards. Distance is float-led and
Fixed Width's material integer work is nominal wide arithmetic. Reopening the
raw-integer design from three new samples would ignore the causal A/B record.

`finishInlineReturn` is likewise not a new leaf. Regex line attribution splits
its 120 ms among return coercion probing, frame pop, program/lookup switching,
slot release, balance bookkeeping, and return-stack append. Future contributes
only 20 ms in different frame-release children. Earlier slotless-guard reorder,
caller-owned frame-result, raw-cell, and minimal-frame variants were neutral or
regressive across unlike programs.

## Compiled intersection

The current generated-main owners do not form a three-application cohort:

| Application | Concrete compiled owner | Cross-mode interpretation |
| --- | --- | --- |
| K-Nucleotide | HashMap equality/hash, String conversion, primitive conversion | same application family only; no other selected compiled owner matches |
| Fixed Width | checked `UInt128` dispatch/extraction and residual nominal construction | one wide-numeric application; primitive-wide and caller-owned-result gains are already retained |
| Distance Field | generated float geometry and imported `sqrt` | `sqrt` is the previously closed Distance/RMS/NBody family |
| Regex Set | canonical regex-NFA transition/closure | related regex algorithms, not an unlike three-program compiler mechanism |
| Future Pipeline | `bridge.currentGID` through `runtime.Stack` | compiled concurrency-only wall; fixed-context ABI regressed unrelated NBody |

No exact compiler-owned descendant recurs in three rows, and no exact
cross-mode operation joins three applications. Generic allocator, GC,
dispatcher, and scheduler frames are consequences of the different owners.

## Reconciliation and cleanup

No candidate passed admission, so building a baseline/candidate cohort would
not answer a supported question. This tranche deliberately does not retry raw
integer/float carriers, slot/stack ordering, call/return frames, primitive
Array growth, regex indexing, fixed execution context, compiled package
ballast, or a named nominal/container rule.

- Both fresh bytecode processes completed and verified below one minute.
- The selection, catalog, and checked-in scoreboard contracts pass.
- Focused bytecode tests pass on the unchanged production tree.
- Raw profiles, two 45.8 MB CLI binaries, stdout/stderr captures, and temporary
  workspaces are removed after this record is written.
- No WASM work was performed.

## Next recommendation

Run a correctness-first feasibility gate for a true register bytecode IR,
starting with static translation and dynamic instruction-elimination accounting
rather than another executable VM candidate.

Why: the current pass completes the exact-leaf search and shows why local
helper changes cannot close interpreter gaps that remain roughly 5x to more
than 100x. Across the full opcode census, slot loads, pops, and stores are a
large shared instruction class, but their downstream semantics differ. The
rejected typed-block prototype entered a second dispatcher while retaining the
original stack operations; it did not test an IR that actually removes
transient stack traffic.

What it entails: define stack effects and side-effect barriers for the existing
bytecodes, translate only statically safe basic blocks into a temporary
register-form model, and combine that model with retained dynamic opcode/block
counts to measure dispatches, snapshots, loads, pops, and stores that would
really disappear. Calls, dynamic dispatch, allocation/identity boundaries,
raises, rescue/ensure, yields, spawn/await, and unknown stack effects must end a
block and fall back to ordinary bytecode. Admit an executable prototype only
if the model removes a material share—preferably at least 15%—of dynamically
executed instructions in at least three unlike verified applications without
specializing an application, source sequence, primitive Array, regex, or named
nominal type. This is a larger architectural test, but the completed evidence
shows that another branch reorder or carrier substitution is unlikely to move
the product targets.
