# Bytecode Float/Slot Profile Gate

Date: 2026-07-16

## Outcome

The bounded Distance Field, RMS Norm, and reduced NBody bytecode profile gate
is complete. Keep no VM, compiler, benchmark, or stdlib production change from
this tranche.

The profiles do identify two real shared costs: ordinary slot-to-stack movement
and call setup/return. A direct float/value type-switch shortcut did not clear
the repeated broad benchmark gate, and immediately enabling the existing raw
float slot sidecar failed focused correctness tests because not every fused
float consumer reads through the sidecar-aware accessors.

## Bounded baseline

All warmed runs used `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, one measured
`main()` call after one warmup call, source-root-only loading, and a 50-second
per-process timeout.

| Program | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Distance Field | 6,134,317,804 | 512,083,872 | 38,000,201 |
| RMS Norm | 5,766,723,903 | 592,083,376 | 52,000,223 |
| reduced NBody (10,000 advances) | 1,700,506,406 | 97,645,352 | 6,161,836 |

The CPU profiles were captured in separate single warmed processes. Their
shared material operations were:

- `appendSlotStackValueChecked`: 4.09% cumulative in Distance, 4.17% in RMS,
  and 7.10% in reduced NBody;
- `execCallOpcode`: 45.01%, 34.43%, and 42.60% cumulative respectively; and
- allocation/runtime work: `runtime.mallocgc` was 16.37% cumulative in
  Distance, 19.30% in RMS, and 5.92% in reduced NBody.

Float arithmetic/storage is material in the two scalar applications, but it
does not recur as the same dominant leaf in reduced NBody. RMS in particular
attributes 13.22% cumulative to `bytecodeNormalizedRawFloatSlotValue`, mostly
the interface conversion beneath its f64 return. Call setup/return is the
larger cross-program family, but its concrete descendants still differ between
native static math calls and Able inline calls.

## Rejected direct slot-load shortcut

The first candidate returned raw f32, raw f64, and value-form `FloatValue`
directly from `appendSlotStackValueChecked` instead of entering the generic
snapshot helper. This is semantics-preserving and generic, but it did not
remove any allocation.

Distance received a second A/B repetition because the first result disagreed
with the other programs. Its combined six-sample means were 6,146,343,934
ns/op baseline and 6,209,068,734 ns/op candidate, a 1.02% regression. The
three-sample RMS result improved 0.52%, and reduced NBody improved 2.58%.
Bytes and allocation counts were unchanged to measurement precision in every
case. This is mixed workstation-scale movement, not a broad retained win, so
the shortcut was reverted.

## Rejected direct sidecar activation

The second experiment made discarded fused float stores use the existing
frame-owned raw-float slot sidecar. This would avoid repeatedly boxing a raw
float into a `runtime.Value`, and therefore directly addresses the scalar
allocation profile.

It did not reach performance measurement. Focused tests immediately found two
correctness failures:

- `TestBytecodeVM_FloatAddSubSlotUpdateParity` failed with an arithmetic
  operand error; and
- `TestBytecodeVM_StoreSlotFloatAffineParity` reached return coercion with a
  nil value and panicked.

The sidecar's call-frame save/restore and generic slot accessors exist, but
some fused float consumers still read `vm.slots` directly. Making the sidecar
authoritative before reconciling all readers therefore hides live values. The
activation was reverted, and the focused float/return/frame suite passes again.

## Verification

After both reverts:

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*Float|TestBytecodeVM_.*Return|TestBytecodeVM_.*Slot.*Restore|TestBytecodeVM_.*CallFrame' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter
```

## Next recommendation

Reconcile the raw-float slot sidecar across every bytecode slot reader before
trying to activate it in production.

Why: the cumulative allocation evidence is much larger and more stable than
the rejected branch shortcut, and the VM already has frame ownership,
pooling, and return restoration for raw float sidecars. The failed experiment
pinpointed a correctness completeness problem rather than disproving the
optimization. A complete sidecar path would be a generic primitive-float VM
improvement shared by scalar math, arrays, methods, and user programs.

What it entails: inventory direct `vm.slots[...]` float reads in fused
arithmetic, cast, call-argument, return, match/type, and array/member paths;
route semantically eligible reads through sidecar-aware helpers; add snapshot,
aliasing, call-frame, error, and mixed f32/f64 regression tests; then enable the
sidecar only for discarded raw-float stores and repeat the three-program A/B
gate. If the audit would require broad speculative churn, stop and instead
profile the shared static-call setup/return descendants. Continue to defer
WASM.
