# Fixed-Context Candidate Residual Profiles

## Decision

Keep the opt-in fixed-pointer execution-context ABI unchanged. The candidate
continues to satisfy its Channel-Rollup gain and serial guard, but the first
post-change profiles do not yet justify another implementation change.

## Method

Built the current `-experimental-execution-context` generated binaries and
verified their ordinary outputs before profiling:

| Workload | Input | Verified output | Profile launches |
| --- | --- | --- | ---: |
| Channel-Rollup | 1,743,363-byte ENABLE word list | `16384:4828:502100` | 5 |
| Lexical-Rollup | same word list | `16384:4828:502100` | 10 |
| Word-Frequency | 131,072-byte `corpus.md` | `1937:11878177` | 10 |
| Document-Audit | same `corpus.md` | `1937:102:83257` | 100 |

Each normal process used `ABLE_GO_CPU_PROFILE`, `GOMEMLIMIT=1GiB`, and
`GOGC=50`; the serial workloads also used `GOMAXPROCS=1`, while
Channel-Rollup retained `ABLE_EXECUTOR=goroutine`. Per-process profiles were
merged into:

- `.profiles/20260710_fixed_context_{channel_rollup,lexical_rollup,word_frequency,document_audit}_main_collector_free.cpu.pprof`

An initial `ABLE_GO_PHASE_PROFILE_DIR` Channel-Rollup attempt was discarded:
its first profiled process exceeded 90 seconds at full CPU with its `main`
profile still active, versus the candidate's normal 2.04-second mean. Exact
allocation phase sampling distorts this scheduler-intensive workload, as it
already does ordinary `main` timing. Bootstrap remains classified by the
existing phase analysis; the normal profiles are used only for residual runtime
attribution.

## Results

| Workload | Merged profile samples | Material local work | Shared-looking leaf |
| --- | ---: | --- | --- |
| Channel-Rollup | 22.74 CPU s | allocation/GC; byte-array/string conversion | `__able_int64_from_value` 6.55% cumulative |
| Lexical-Rollup | 34.39 CPU s | allocation/GC; `String_to_builtin` / `Array u8` conversion | `__able_int64_from_value` 8.08% cumulative |
| Word-Frequency | 4.11 CPU s | `__able_hash_map_find_entry` 19.71% flat | `__able_int64_from_value` 7.06% cumulative |
| Document-Audit | 25.71 CPU s | allocation/GC; `Array u8` conversion 20.73% cumulative | `__able_int64_from_value` 8.17% cumulative |

The removed bridge goroutine-identity wall does not recur materially in the
Channel profile. That confirms the retained fixed-context ABI is doing its
intended generic work; remaining top entries are allocation and conversion
costs rather than `currentGID`/`runtime.Stack` lookup.

`__able_int64_from_value` is the only exact leaf visible in all four profiles.
It is not yet a selected candidate. In the serial applications it is reached
through the common file/string byte-conversion path (`String_to_builtin` and
`Array u8` conversion), and all four workloads ingest files. Word-Frequency's
map lookup is workload-local, while the surrounding string/array conversion
and GC descendants do not recur as one independent leaf across non-file
programs. Optimizing a file-input-shaped path now would violate the benchmark
selection rule.

No compiler, runtime, benchmark, or `able-stdlib` source changed in this
profiling tranche.

## Next recommendation

Audit the primitive byte-array/string conversion boundary and profile it
against at least one maintained in-memory numeric or collection program that
does not read files, alongside the existing four applications. Why: the
repeated integer-extraction sample may be a broadly useful primitive conversion
cost, but current evidence is confounded by shared file ingestion. The work
entails call-path counting, output-checked bounded profiles, and a candidate
only if the same semantic conversion leaf remains material without that input
path. Do not specialize `HashMap`, files, word lists, channels, or a benchmark.
