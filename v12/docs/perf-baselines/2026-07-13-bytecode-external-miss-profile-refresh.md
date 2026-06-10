# Bytecode External Miss Profile Refresh — 2026-07-13

## Decision

Keep no bytecode VM, tree-walker, compiler, runtime, canonical-stdlib, or
benchmark-source change from this tranche. Fresh bounded profiles again divide
the material cost among unrelated mechanisms: host codec and MD5 work in
Base64, the previously rejected raw-float slot lane in Monte Carlo Pi, and
text-member dispatch/cache work in I-Before-E. `runResumable` and
`execCallOpcode` are common parent frames, not a common removable leaf.

Accordingly, do not retry the raw-float/frame, call-name/raw-cell,
inline-return, or member-cache variants that previously failed the broad
benchmark bar. Do not add codec, MD5, benchmark, or nominal-container-specific
special cases.

## Method

The public bytecode comparison used the canonical external stdlib, one pinned
CPU (`15`), `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. It ran once per
benchmark/reference pair with the normal 60-second timeout and completed all
external output and exit-status checks.

| Benchmark | Normal bytecode elapsed |
| --- | ---: |
| Base64 | 3.21 s |
| Monte Carlo Pi | 2.70 s |
| I-Before-E | 0.57 s |

The runtime harness then warmed each program before profiling on the same CPU.
Base64 ran twice, Monte Carlo Pi three times, and I-Before-E thirty times to
obtain bounded samples while preserving a useful profile duration.

| Benchmark | Warmed `ns/op` | Profile samples | Allocations/op |
| --- | ---: | ---: | ---: |
| Base64 | 2,664,270,904 | 5.31 s | 581 |
| Monte Carlo Pi | 2,487,936,512 | 7.43 s | 22,222,114 |
| I-Before-E | 238,740,053 | 7.13 s | 1,923 |

Retained profiles:

- `v12/interpreters/go/.profiles/20260713_external_base64_bytecode_refresh.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_external_monte_carlo_pi_bytecode_refresh.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_external_i_before_e_bytecode_refresh.cpu.pprof`

The temporary harness binary, raw comparison output, profile summaries, and
runtime statistics remain under
`v12/tmp/perf/2026-07-13-bytecode-miss-profile-refresh/`; they are explicitly
cleanup-eligible.

## Attribution

| Benchmark | Material profile evidence | Interpretation |
| --- | --- | --- |
| Base64 | `execAndFinishExactNativeCall`: 4.45 s cumulative (83.8%); Go Base64 encode/decode and MD5 block work are the prominent descendants | Host codec and digest work, not a shared VM leaf |
| Monte Carlo Pi | `execStoreSlotCastSlotFloatConstDiv`: 2.80 s cumulative (37.7%); `storeReusableFloatSlotRaw`: 1.70 s (22.9%) | The raw-float lane already tested broadly and rejected |
| I-Before-E | `execCallOpcode`: 4.20 s cumulative (58.9%); `execCallMember`: 2.06 s (28.9%); cached member lookup and `finishInlineReturn` are prominent descendants | Text member/call dispatch, distinct from the other two |

The dispatcher frames recur only because every VM program passes through them.
Their non-overlapping descendants fail the required repeated-concrete-leaf bar
for a general optimization.

## Next Recommendation

Profile two unlike compiled target misses—N-body (numeric/package-heavy) and
I-Before-E (file/text-heavy)—with the same bounded, verifier-backed procedure.
The compiled-performance target remains materially farther from Go than the
bytecode target is from Python/Ruby, while this bytecode refresh supplies no
safe source candidate. The next tranche should collect fresh reference-aware
compiled measurements and warmed generated-program CPU profiles, then consider
code only if both applications expose the same concrete compiler/runtime leaf.
