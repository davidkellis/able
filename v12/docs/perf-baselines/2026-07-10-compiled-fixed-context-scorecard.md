# Fixed-Context Default Comparison Scorecard

## Decision

Keep the fixed-pointer execution-context ABI opt-in for now. Its broad matched
performance scorecard has no regression and removes the intended goroutine
identity wall, but switching the compiler default requires a separate full
semantic parity gate for dynamic compatibility boundaries.

## Method

Fresh default binaries were built from the current source beside the retained
`-experimental-execution-context` binaries. Generated source trees were
discarded after each build and only binaries retained. Every binary was run
before timing and produced its expected output.

Wall controls used three ordinary launches per mode with `GOMEMLIMIT=1GiB` and
`GOGC=50`; serial programs also used `GOMAXPROCS=1`, while Channel-Rollup and
BinaryTrees used `ABLE_EXECUTOR=goroutine`. File-backed programs received their
canonical input paths.

Default normal CPU profiles were collected with the same launch counts as the
retained candidate profiles: Channel-Rollup 5, Lexical-Rollup 10,
Word-Frequency 10, and Document-Audit 100. They are retained as:

- `.profiles/20260710_default_{channel_rollup,lexical_rollup,word_frequency,document_audit}_main_collector_free.cpu.pprof`

The matching opt-in artifacts use the `fixed_context` prefix.

## Output and wall scorecard

| Workload | Expected output | Default, 3 runs | Fixed context, 3 runs | Reading |
| --- | --- | ---: | ---: | --- |
| Channel-Rollup | `16384:4828:502100` | 3.81, 3.70, 3.90 s (3.80 mean) | 2.29, 2.18, 2.20 s (2.22 mean) | 41.5% lower |
| Lexical-Rollup | `16384:4828:502100` | 3.82, 3.55, 3.95 s (3.77 mean) | 3.71, 3.52, 3.56 s (3.60 mean) | 4.7% lower |
| Word-Frequency | `1937:11878177` | 0.49, 0.54, 0.55 s | 0.52, 0.54, 0.49 s | within short-run noise |
| Document-Audit | `1937:102:83257` | 0.46, 0.37, 0.35 s | 0.37, 0.35, 0.34 s | lower; short-run evidence |
| BinaryTrees small | checked tree report | 0.05, 0.05, 0.05 s | 0.05, 0.05, 0.05 s | neutral |
| Array-map i32 | `1097192358` | 0.08, 0.07, 0.07 s | 0.07, 0.06, 0.06 s | short-control improvement |
| LinkedList iterator collect | `382455000` | 0.12, 0.13, 0.10 s | 0.12, 0.13, 0.10 s | neutral |

No workload regressed. The short controls do not establish a separate speed
claim; they are regression guards for the compiler-wide ABI.

## Profile comparison

The fixed context removes the selected generic bridge wall in Channel-Rollup:

| Descendant | Default cumulative | Fixed-context cumulative |
| --- | ---: | ---: |
| `bridge.currentGID` | 32.25% | 1.32% |
| `runtime.Stack` beneath it | 32.10% | 1.32% |
| `bridge.SwapEnvIfNeeded` | 12.69% | not material in the candidate top |
| `runtime.procyield` | 8.11% | 0.84% |

Lexical-Rollup remains structurally neutral: its byte/string conversion and
integer helper descendants are nearly unchanged (`String_to_builtin` 18.47%
default versus 18.32% candidate; `__able_int64_from_value` 8.23% versus
8.08%). Word-Frequency and Document-Audit remain dominated by their existing
map or allocation/conversion work in both modes. This matches the wall data:
the ABI removes the generic direct-call bridge tax without trading it for a
serial application cost.

No compiler, runtime, benchmark, or `able-stdlib` source changed in this
measurement tranche.

## Next recommendation

Run the experimental ABI through the full compiler fixture/parity matrix,
including dynamic named/value calls, imported packages, host externs, dynamic
imports, nested concurrency/cancellation, and generated race builds. Why: the
performance gate now passes broadly, while default enablement depends on
proving compatibility-wrapper semantics at every remaining dynamic boundary.
The work entails parameterizing the compiler fixture harness with the option,
adding any missing boundary assertions, and comparing the experimental output
against the tree-walker/reference results. Do not flip the default until that
semantic gate is green.
