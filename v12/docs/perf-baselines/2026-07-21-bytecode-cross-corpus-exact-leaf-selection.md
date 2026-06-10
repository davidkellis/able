# Bytecode cross-corpus exact-leaf selection gate

Date: 2026-07-21

## Decision

Keep no runtime or compiler code. Fresh profiles of six unlike, high-excess
bytecode applications do not expose one new Able-owned CPU leaf that is
material in at least three applications.

The exact flat-symbol intersection contains the VM dispatcher, previously
closed stack/raw-integer/call-frame families, Go allocation/GC machinery, and
two Go map/hash leaves. Caller-tree reconciliation shows that the map/hash
leaves combine different Able maps and operations rather than one shared
lookup. No candidate clears the preimplementation generality gate, so no A/B
timing or guard-row promotion is warranted.

No VM, compiler, generated runtime, canonical stdlib, benchmark, fixture,
language, reference, scorecard, or WASM code changed.

## Profile protocol

A clean current interpreter test binary was frozen before collection. Every
application received two independent CPU-profile processes, and the two
profiles were merged per application. Future Pipeline ran ten warmed `main()`
calls per process, Word Frequency ran three, and the four longer programs ran
one. Bytecode statistics were disabled.

Each process used the canonical external stdlib, `GOMAXPROCS=1`, `GOGC=50`, a
1-GiB memory limit, and a 60-second cap. Programs used their catalog working
directories, inputs, and executor modes. The reported means are arithmetic
means of the two profile processes and are diagnostics, not promoted external
scorecard timings.

| Application | Merged CPU samples | Profile-process mean |
| --- | ---: | ---: |
| Future Pipeline | 5.57 s | 279,248,084 ns |
| Word Frequency | 6.78 s | 1,133,998,780 ns |
| Distance Field | 11.25 s | 5,637,316,484 ns |
| Regex Set Audit | 7.51 s | 3,764,764,278 ns |
| Fixed Width 128 | 16.20 s | 8,156,288,163 ns |
| Reverse Complement | 5.78 s | 2,901,942,664 ns |

## Exact flat-symbol intersection

A symbol had to own at least 1% flat CPU in an application. The complete
intersection with reach in three or more applications is:

| Exact symbol | Applications at or above 1% | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 6 | 3.70%-16.38% | dispatcher parent; exact-line gate already closed |
| `(*bytecodeVM).appendSlotStackValueChecked` | 5 | 1.73%-4.66% | closed stack-carrier family |
| `bytecodeRawIntegerValueInfo` | 5 | 1.86%-3.98% | closed raw-integer family |
| `runtime.tryDeferToSpanScan` | 5 | 1.62%-5.99% | Go GC machinery, not an Able operation |
| `aeshashbody` | 4 | 1.20%-2.65% | different string-key maps after caller reconciliation |
| `internal/runtime/maps.ctrlGroup.matchH2` | 4 | 2.93%-9.69% | different maps and key shapes after caller reconciliation |
| `(*bytecodeVM).execCallOpcode` | 3 | 1.33%-2.13% | call dispatcher parent already closed |
| `(*bytecodeVM).popCallFrameFields` | 3 | 1.54%-2.22% | closed call-frame family |
| `(*bytecodeVM).pushCallFrame` | 3 | 1.07%-1.47% | closed call-frame family |
| `runtime.nextFreeFast` | 3 | 1.62%-2.31% | Go allocator machinery, not one Able allocation site |

This broad pass confirms the preceding main-dispatch result: shared parent
frames are large, but their concrete semantic children divide by application.
Reopening a stack, raw-integer, call, return, or frame micro-variant would
repeat an already rejected family rather than use new evidence.

## Map and hash caller reconciliation

`ctrlGroup.matchH2` initially looked new because it is material in Word
Frequency, Regex Set Audit, Fixed Width 128, and Reverse Complement. Its direct
and immediate Able owners are not shared:

- Word Frequency splits samples among Array-store lease/backing maps, Array
  tracking, and type/alias caches. No one caller owns more than 1.03% of the
  complete profile.
- Regex Set Audit is led by `bytecodeNamedStructMemberPlanAt`, with separate
  Array-store handle/size maps.
- Fixed Width 128 is dominated by `bytecodeBoxedIntegerValue`, with a smaller
  named-struct member-plan component.
- Reverse Complement is dominated by `bytecodeBoxedIntegerValue`, with smaller
  Array backing and mono-u8 state maps.

Thus the strongest immediate owner, boxed-integer lookup, occurs materially in
only two of the four applications. The named-struct plan also occurs in only
two. Combining them under the Go map matcher would falsely treat unrelated
semantics as one optimization target.

AES string hashing likewise divides below the runtime leaf. Word Frequency
mixes ordinary string-key maps with interface/member cache keys; Distance
Field is led by simple type-coercion name lookup; Regex Set Audit mixes cached
call-name, Array-slot, and static-receiver keys; Reverse Complement is led by
integer target-kind/name lookup. There is no common Able-owned string-key map
that reaches 1% in three applications.

## Verification and cleanup

No candidate or temporary diagnostics were added to the source tree. The full
restored bytecode family passes:

```text
go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  25.378s
```

The frozen binary, raw profiles, merged profiles, and aggregation tables are
temporary and are removed after this record is written.

## Next recommendation

Give the compiler target the next selection tranche: run a current
generated-main exact-leaf sweep across six unlike non-concurrency misses,
initially Matrix Multiply, Mandelbrot, Reverse Complement, Fixed Width 128,
Word Frequency, and Regex Set Audit.

Why: this bytecode sweep found no new shared concrete leaf, while only 5 of 45
rankable compiled rows currently meet the 95%-of-Go target. The compiler has
large misses across numeric, byte-output, wide-integer, text/map, and regex
programs. Looking across those families can identify a shared generated-runtime
or lowering cost without continuing to subdivide exhausted VM parents.

What it entails: freeze current generated binaries and matching Go references,
verify every output, collect bounded user-main CPU profiles separately from
cold startup, and build the same exact-symbol/leaf intersection with a
three-unlike-application admission threshold. Exclude already closed fixed
startup, boxed-cache, nominal-result, interface, and generated-dispatch
families. Implement at most one concrete owner, using only primitive lowering
or the shared nominal pipeline, then require repeated order-balanced arithmetic
means for all admitted applications plus unrelated compiler and bytecode
guards. Continue to defer WASM.
