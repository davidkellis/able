# Bytecode text/map five-application refresh

Date: 2026-07-21

## Decision

Keep no VM, compiler, stdlib, language, or workload change. Two independent
main-only CPU profiles and two exact main-allocation processes for each of
K-Nucleotide, Word Frequency, Inventory Reconciliation, Reverse Complement,
and JSON expose no new concrete Able-owned leaf that is material in three
unlike applications.

The broad intersection is real only at dispatcher, stack, raw-integer,
call-frame, Go map, and Go GC symbols. The VM families have already completed
broad candidate gates. Caller reconciliation splits the Go map leaves among
integer boxing, type/member/call caches, Array bookkeeping, language HashMap
work, and JSON's native host parser. Treating the runtime map implementation
as one Able optimization target would combine unrelated semantics.

This closes the largest frontier ownership group without a speculative
candidate. The only retained source change is a test-harness diagnostic that
adds exact `runtime.MemStats` deltas for the timed `main()` phase to the
existing opt-in retention report.

## Protocol

A current interpreter test binary was frozen and every program ran in its own
process with the canonical external stdlib, `GOMAXPROCS=1`, `GOGC=50`, a
1-GiB memory limit, and a 59-second test cap. Loading, typechecking, lowering,
and final forced collections were outside the CPU and allocation measurement
window. Public benchmark inputs and working directories were preserved.

CPU profiles were collected twice per application and merged. Allocation
counters were then collected twice without CPU profiling. Means below are
arithmetic means, following the workstation-noise policy. All ten CPU runs and
all ten allocation runs completed successfully; the source files match the
current verifier-backed scorecard cohort, and the canonical 70-file stdlib
tree remained
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.

| Application | CPU runs | CPU mean | Merged CPU samples |
| --- | --- | ---: | ---: |
| K-Nucleotide | 42.522 s, 46.650 s | 44.586 s | 88.82 s |
| Word Frequency | 1.141 s, 1.181 s | 1.161 s | 2.32 s |
| Inventory Reconciliation | 2.300 s, 2.341 s | 2.320 s | 4.63 s |
| Reverse Complement | 2.836 s, 3.096 s | 2.966 s | 5.89 s |
| JSON | 0.518 s, 0.537 s | 0.528 s | 0.99 s |

## Exact CPU intersection

The admission threshold was at least 1% flat CPU in at least three unlike
applications.

| Exact symbol | Breadth | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `runtime.tryDeferToSpanScan` | 5 | 1.08%-3.88% | Go GC machinery |
| `(*bytecodeVM).runResumable` | 4 | 5.83%-8.65% | dispatcher parent; JSON runs mostly in a native host kernel |
| `internal/runtime/maps.ctrlGroup.matchH2` | 4 | 3.02%-8.83% | different map shapes and Able owners |
| `bytecodeRawIntegerValueInfo` | 4 | 1.72%-3.20% | previously closed raw-integer family |
| `aeshashbody` | 4 | 1.72%-4.32% | different string-key maps; JSON is native-host work |
| `appendSlotStackValueChecked` | 4 | 1.35%-2.55% | previously closed stack-carrier family |
| `execCallOpcode` | 3 | 1.94%-2.59% | call dispatcher parent already closed |
| `runtime.mapaccess2_faststr` | 3 | 1.73%-2.59% | different caches and language operations |
| `popCallFrameFields` | 3 | 1.29%-2.59% | previously closed call-frame family |
| `execBinary` | 3 | 1.29%-1.62% | binary dispatcher parent already closed |
| `execStoreSlot` | 3 | 1.07%-2.16% | different value and container paths |

JSON is an important discriminator rather than a VM candidate: about 95% of
its samples are in the native JSON field-means kernel, led by float parsing.
Its string map activity belongs to the generated native parser, not to the VM
caches seen in the other applications.

## Map/hash owner reconciliation

`ctrlGroup.matchH2` does not identify one data structure or operation:

- K-Nucleotide divides primarily among integer type information, boxed
  integer values, type aliases/matches, call-name caches, and Array state.
- Word Frequency divides among string and integer-key cache maps; its AES
  samples include known-type, named-struct, integer-info, and environment
  lookups.
- Inventory divides among member-method cache keys, type-expression keys,
  ordinary string maps, and the language-level integer HashMap operation.
- Reverse Complement's fast-integer map access is 86% integer boxing, with
  the remainder in Array backing/size and monomorphic-u8 state maps.
- JSON's relevant string-map path belongs to the native JSON field parser.

The same runtime leaf therefore has different keys, values, lifetimes, and
semantic owners. No shared VM map representation or lookup site reaches the
three-unlike-program admission threshold.

## Exact main allocation

The two-process exact counters are much more stable than wall time. Byte
spreads are 0.0029%-0.0239%; object-count spreads are 0.0010%-0.0185% for the
four allocation-heavy VM programs. JSON performs only about 443 measured Go
allocations, so its 55-object difference is a large percentage but immaterial
beside its 115-MiB native input buffer.

| Application | Mean allocated bytes | Mean allocations | Mean frees | Mean GCs |
| --- | ---: | ---: | ---: | ---: |
| K-Nucleotide | 1,245,836,872 | 23,947,521 | 21,935,718 | 35.5 |
| Word Frequency | 54,289,380 | 625,808 | 339,431 | 3.0 |
| Inventory Reconciliation | 41,668,988 | 2,096,718 | 1,762,356 | 3.0 |
| Reverse Complement | 266,675,724 | 4,069,525 | 1,837,915 | 7.0 |
| JSON | 114,820,196 | 443 | 71,728 | 1.5 |

Allocation caller inspection also diverges: K-Nucleotide spreads through
ordinary VM/runtime machinery; Word Frequency includes structs, Array state,
String hosts, and caches; Inventory includes raw integer results, interface
coercion, and call arguments; Reverse Complement is led by host-byte Array
conversion and stack snapshots; JSON is dominated by the native input buffer.
There is no common concrete allocator to implement.

## Verification

The focused harness test and the full bytecode test family pass:

```text
go test ./pkg/interpreter -run '^TestBytecodeProgramRuntimeRetention$' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.034s

go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  24.118s
```

The frozen binaries and raw/merged profiles are temporary and are removed
after the record and generated frontier are verified.

## Next recommendation

Refresh the `bytecode-wide-numeric` ownership group across Fixed Width 128,
Rational, and Wide Integer Records, with two unrelated numeric/bytecode
discriminators.

Why: after closing the 51.681-second text/map group, wide numeric is the next
largest bytecode frontier group at 16.595 target-excess seconds. Its current
disposition rests on profiles from before the latest VM and runtime changes,
and the three core programs exercise checked unsigned arithmetic, rational
division/casts, parsing, comparison, and bitwise work rather than one named
container.

What it entails: freeze one current binary, collect two bounded main-only CPU
and exact-allocation processes per application, and intersect concrete leaves
at the same three-unlike-program threshold. Exclude the already rejected
raw-integer carrier/extractor variants unless a new concrete owner invalidates
that evidence. Implement at most one general primitive or VM operation and
keep it only after repeated, order-balanced target and unrelated guard means.
No named wide-type rule, stdlib special case, benchmark-specific opcode, or
WASM work is admissible.
