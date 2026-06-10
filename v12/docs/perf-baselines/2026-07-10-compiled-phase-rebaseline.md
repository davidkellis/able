# Current Compiled-Phase Rebaseline

## Decision

Keep no compiler, runtime, stdlib, or benchmark-source change. Current
generated binaries repeat the known generic bootstrap AST-decoding boundary,
but the two general replacements for that boundary have already failed broad
throughput/process guards. The exact-allocation phase profiler intentionally
perturbs `main`, so it cannot select a compiled-runtime lowering candidate.

## Method

- Built current generated binaries for Word-Frequency, Document-Audit,
  Lexical-Rollup, Channel-Rollup, and Fib with the canonical stdlib,
  `GOMEMLIMIT=1GiB`, and `GOGC=50`.
- Ran each binary with `ABLE_GO_PHASE_PROFILE_DIR`, which records distinct
  bootstrap and `main` CPU profiles plus exact allocation deltas. Serial
  applications used one P; Channel-Rollup retained the goroutine executor.
- Verified output from every generated binary: Word-Frequency
  `1937:11878177`, Document-Audit `1937:102:83257`, Lexical-Rollup and
  Channel-Rollup `16384:4828:502100`, and Fib `1134903170`.
- Retained bootstrap/main profiles and phase stats as
  `.profiles/20260710_{word_frequency,document_audit,lexical_rollup,
  channel_rollup,fib}_compiled_phase_rebaseline_{bootstrap,main}.cpu.pprof`
  and matching `_stats.json` files.

## Repeated bootstrap boundary

Generated registration calls `interpreter.DecodeNodeJSON` while restoring the
exported AST metadata needed by definitions, methods, implementations, and
defaults. It repeats below `RegisterIn` in every current profile.

| Workload | Bootstrap allocation | Allocations | `DecodeNodeJSON` CPU / bootstrap samples |
| --- | ---: | ---: | ---: |
| Word-Frequency | 4,530,080 B | 17,580 | 120 ms / 150 ms (80.0%) |
| Document-Audit | 4,483,808 B | 16,757 | 100 ms / 140 ms (71.4%) |
| Lexical-Rollup | 4,487,568 B | 16,792 | 90 ms / 150 ms (60.0%) |
| Channel-Rollup | 4,519,104 B | 16,821 | 90 ms / 140 ms (64.3%) |
| Fib control | 2,801,928 B | 4,952 | 30 ms / 50 ms (60.0%) |

This is a broad compiler metadata boundary, not a filesystem, Channel, map,
or nominal-container condition. It is nevertheless not a new candidate:

- The complete direct-Go-constructor replacement made ordinary generated
  application builds exceed six CPU minutes in source parsing.
- The compact tagged-codec replacement was neutral on Document-Audit and
  Lexical-Rollup and regressed i-before-e by 2.49% under guarded process-wall
  comparison.

Those rejected designs and their full guard evidence remain in
`2026-07-10-compiled-phase-profiles.md`. The rebaseline provides no new
representation, call boundary, or allocation subpath that could change that
decision.

## Main-phase limitation

`ABLE_GO_PHASE_PROFILE_DIR` deliberately enables exact allocation sampling and
writes allocation snapshots at phase boundaries. It therefore puts
`runtime.profilealloc`, stack unwinding, and profile-writer work into the
instrumented `main` CPU samples. Main allocation deltas also vary with each
application's ordinary workload (for example 2.8 MB in Document-Audit versus
44.8 MB in Channel-Rollup). They are useful for phase attribution, not for
ranking generated-binary runtime bridge or lowering cost.

Do not infer a candidate from `bridge.ToUint`, string conversion, generated
function names, or runtime allocation helpers in these instrumented main
profiles. That would optimize collector behavior or one application path.

## Next recommendation

Collect collector-free, merged CPU profiles of normal compiled `main`
execution across the same applications and Fib control. Why: the phase
rebaseline validates the known bootstrap boundary but cannot rank main runtime
cost without its exact-allocation instrumentation. The work entails repeated
output-checked launches with ordinary `ABLE_GO_CPU_PROFILE`, merging profiles
per workload, and selecting a change only when one generic generated-lowering
or runtime-bridge descendant repeats across independent applications. Do not
reopen AST codec replacements or add a Channel, filesystem, map, or
nominal-container special case.
