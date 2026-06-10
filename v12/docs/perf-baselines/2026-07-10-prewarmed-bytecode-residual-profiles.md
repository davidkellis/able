# Prewarmed Bytecode Residual Profiles

## Decision

Keep no bytecode, runtime, compiler, stdlib, or benchmark-source change. The
extern-host prewarm removes cold child plugin compilation, but the residual
bytecode target misses do not share a new concrete VM leaf. The only repeated
return/member families are already-rejected generic candidates.

## Method

- Created an isolated `ABLE_EXTERN_CACHE_DIR` and populated it through
  `able cache prewarm`. The workspace-visible discovery found 68 packages and
  built 10 Go extern-host modules.
- Used canonical `able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
  `GOMAXPROCS=1` for the serial runtime profiles. A current
  `pkg/interpreter` benchmark binary loaded and warmed each program before
  sampling its repeated `main()` calls. Exact iteration counts were selected
  only to obtain useful profile duration; they do not change program logic.
- Channel-Rollup retained its goroutine executor and was captured as one
  normal prewarmed bytecode process. Repeated `main()` calls retain
  executor-sensitive state and are not a valid Channel measurement.
- Separate normal bytecode runs under the same cache verified output: Fib
  `1134903170`, Word-Frequency `1937:11878177`, Document-Audit
  `1937:102:83257`, and Lexical-Rollup/Channel-Rollup `16384:4828:502100`.

| Workload | Boundary | Iterations | Runtime result | CPU samples |
| --- | --- | ---: | ---: | ---: |
| Word-Frequency | warmed bytecode runtime | 3 | 1,398,382,368 ns/op; 48,402,578 B/op; 508,503 allocs/op | 4.17s |
| Document-Audit | warmed bytecode runtime | 100 | 9,277,553 ns/op; 364,927 B/op; 206 allocs/op | 0.92s |
| Lexical-Rollup | warmed bytecode runtime | 10 | 158,527,644 ns/op; 10,251,699 B/op; 291 allocs/op | 1.58s |
| Fib control | warmed bytecode runtime | 1,000,000 | 2,354 ns/op; 448 B/op; 3 allocs/op | 2.35s |
| Channel-Rollup | normal goroutine bytecode process | 1 | output checked | 0.67s across goroutines |

The retained profiles are
`20260710_{word_frequency,document_audit,lexical_rollup,fib}_prewarmed_bytecode_runtime.cpu.pprof`
and `20260710_channel_rollup_prewarmed_bytecode_process.cpu.pprof` under
`v12/interpreters/go/.profiles/`.

## Evidence

| Workload | Material concrete work | Selection consequence |
| --- | --- | --- |
| Word-Frequency | call opcode 36.69% cumulative; inline return 13.43%; HashMap find 6.47% flat; raw integer information 4.80% flat | map/raw-integer descendants are unique to this workload |
| Document-Audit | generator execution 81.52%; call opcode 55.43%; cached member lookup 17.39%; inline return 8.70% | iterator/member-heavy, not map or numeric recurrence work |
| Lexical-Rollup | generator execution 57.59%; inline return 15.19%; type matching 13.92%; cached member lookup 9.49% | public pipeline/type-match/external-host mix differs below dispatch |
| Channel-Rollup | goroutine task work 22.39%; loader 19.40%; parser 14.93%; cached member lookup 7.46% | executor and normal-process initialization remain distinct from the serial runtime samples |
| Fib control | recurrence kernel 52.34%; numeric fit check 13.62% | no-extern numeric control, not an application-target leaf |

`finishInlineReturn(...)` recurs in Word-Frequency, Document-Audit, and
Lexical-Rollup, but the already-tested slotless-return guard reorder was
neutral-to-mixed and was reverted. This profile offers no new semantic return
subpath. Cached member lookup recurs in Document-Audit, Lexical-Rollup, and a
small Channel process sample, but is not material in Word-Frequency and its
previous broad member-cache candidate was rejected. `execCallOpcode` and
`runResumable` are broad dispatcher parents, not candidates.

Do not optimize HashMap lookup, raw integer extraction, generator execution,
the scheduler, parser, `fs`, a named container, or a benchmark call shape from
one descendant. The cache prewarm is retained as packaging work, not evidence
that any runtime operation became universally hot.

## Next recommendation

Profile compiler-generated binary phases across the same independent
applications, with Fib as a no-extern control. Why: the bytecode target misses
now diverge below shared VM parents, while compiled application rows still
have material gaps to Go and often Python/Ruby. The work entails phase-separated
bootstrap/main CPU and allocation captures for Word-Frequency, Document-Audit,
Lexical-Rollup, and Channel-Rollup; output checks; and selection only when a
generic compiler/lowering or runtime-bridge descendant repeats. Do not add
Channel-, filesystem-, map-, or nominal-container-specific lowering.
