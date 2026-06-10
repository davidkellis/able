# Compiled generic-union post-direct profile gate — 2026-07-22

## Decision

The retained direct-known-method change remains in place. Keep no additional
compiler/runtime change from this tranche.

Fresh profiles show that the apparent compiled-method registration wall is a
cumulative parent, not overhead in the registration closure: the closure has
zero or effectively zero flat CPU in every owner. Its children are the
different Able method bodies. Removing that forwarding frame therefore fails
the exact-leaf admission rule.

The one exact compiler-controlled leaf that does repeat is allocation in
`__able_call_known_native_method_fast`: construction of a
`NativeCallContext` and the receiver-prefixed argument slice. A general
`sync.Pool` experiment removed one context allocation per call but failed the
four-application wall gate, so it was reverted.

No canonical-stdlib, bytecode VM, benchmark, language, or WASM change was
made.

## Profile protocol

One retained binary per application was reused for every launch. Fifty
independent, verifier-backed main-only CPU profiles ran with one logical CPU,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 55-second process cap:

- Binary Event Log: 5 profiles, 2.68 seconds of merged samples;
- Option/Result Config: 15 profiles, 1.77 seconds of merged samples;
- Manifest Normalization: 15 profiles, 1.79 seconds of merged samples;
- Policy Record Dispatch: 15 profiles, 1.94 seconds of merged samples.

All 50 processes passed their public verifier. Only the main phase was merged;
bootstrap and allocation-snapshot collector work were excluded.

## Shared-path interpretation

| Application | Static union call cumulative | Known helper cumulative | Registration closure cumulative | Registration closure flat |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 26.49% | 24.63% | 17.54% | 0.00% |
| Option/Result Config | 50.85% | 45.20% | 31.64% | 0.00% |
| Manifest Normalization | 20.11% | 18.99% | 18.44% | 0.56% |
| Policy Record Dispatch | 11.86% | 11.34% | 10.31% | 0.00% |

The closure's cumulative samples descend into different generated Option,
Result, record, conversion, text, and policy work. The shared runtime wall is
allocation instead: `runtime.mallocgc` is 43.02–58.76% cumulative across the
four profiles, and the known helper owns 610,304, 417,792, 57,344, and 14,848
flat allocation objects in the exact current profiles.

## Rejected context-pool experiment

The experiment reused `NativeCallContext` through the same acquire/reset/release
shape used by the Go interpreter. It preserved active `Env`/`State`, receiver
ordering, error propagation, and the dynamic fallback. Five exact allocation
runs per application confirmed that it removed almost exactly one allocation
per generic-union method call:

| Application | Current objects | Pool objects | Object delta | Current bytes | Pool bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,465,436.8 | 4,240,199.8 | -5.04% | 273,425,644.8 | 268,022,041.6 | -1.98% |
| Option/Result Config | 1,360,322.0 | 1,212,875.4 | -10.84% | 46,044,529.6 | 42,506,240.0 | -7.68% |
| Manifest Normalization | 997,350.6 | 976,834.4 | -2.06% | 45,704,526.4 | 44,611,422.4 | -2.39% |
| Policy Record Dispatch | 961,314.8 | 955,694.0 | -0.58% | 48,748,643.2 | 48,619,145.6 | -0.27% |

Because the workstation signal was volatile, every owner was expanded to 20
verified processes per variant with direction-reversed cohorts:

| Application | Current mean | Pool mean | Delta | Current median | Pool median |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 0.7325 s | 0.7150 s | -2.39% | 0.700 s | 0.700 s |
| Option/Result Config | 0.1975 s | 0.2180 s | +10.38% | 0.210 s | 0.210 s |
| Manifest Normalization | 0.2245 s | 0.2040 s | -9.13% | 0.220 s | 0.195 s |
| Policy Record Dispatch | 0.2160 s | 0.2280 s | +5.56% | 0.210 s | 0.225 s |

The heap reduction is real, but synchronization/repopulation cost is not a
broad wall win. Option contains two long candidate pauses and Policy shows a
consistent median regression. The candidate therefore fails the broad gate
and has been fully reverted. The companion JSON retains all 160 valid timing
samples and allocation/profile summaries. An earlier Policy setup attempt used
the wrong input filename; its 20 symmetric pre-verifier failures are excluded
from the gate and recorded as harness error, not program evidence.

## Verification

After the revert, the focused generated-source, executable semantics, and
standalone generic-named-union tests pass. The retained direct-known-method
source is unchanged from the prior tranche.

## Next direction

Audit a split-receiver compiled-method wrapper ABI before implementing it. The
remaining exact allocation in all four owners is the fresh
`[receiver] + args` slice. A general wrapper that accepts the receiver
separately could remove that allocation without the rejected synchronization
cost, and could benefit every compiled instance method rather than naming
Option, Result, or any container. The audit must trace arity/partial-call,
package environment, control/error, static, UFCS, and dynamic fallback
semantics; prototype only if one shared representation can cover them. Gate
any prototype against these four owners plus unlike compiled controls, again
using repeated verifier-backed averages.
