# Current Compiled Main Collector-Free Profiles

## Decision

Keep no compiler, runtime, stdlib, or benchmark-source change. Normal compiled
program profiles remove the phase collector's allocation snapshots, but they do
not identify a concrete generated-lowering or runtime-bridge leaf shared by
two independent applications. In particular, Channel-Rollup exposes a very
large future-executor cost, but it is one concurrency application and not
enough evidence for a general scheduler change.

## Method

- Reused current generated binaries for Word-Frequency, Document-Audit,
  Lexical-Rollup, Channel-Rollup, and Fib. Each launch used ordinary
  `ABLE_GO_CPU_PROFILE`, not `ABLE_GO_PHASE_PROFILE_DIR`; no allocation or
  phase collector was enabled.
- Used `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1` for serial programs.
  Channel-Rollup retained its normal goroutine executor.
- Repeated normal process launches, output-checking every result, then merged
  each workload's CPU profiles. Launch counts were Word-Frequency 10,
  Document-Audit 100, Lexical-Rollup 10, Channel-Rollup 5, and Fib 3.
- Outputs were Word-Frequency `1937:11878177`, Document-Audit
  `1937:102:83257`, Lexical-Rollup and Channel-Rollup `16384:4828:502100`,
  and Fib `1134903170`.
- Retained merged profiles as
  `.profiles/20260710_{word_frequency,document_audit,lexical_rollup,
  channel_rollup,fib}_compiled_main_collector_free.cpu.pprof`.

## Results

| Workload | Merged samples | Material descendants | Interpretation |
| --- | ---: | --- | --- |
| Word-Frequency | 1.52 s | `__able_hash_map_find_entry` 42.8% flat; `String_split` 41.4% cumulative; `bridge.ToUint` 9.9% cumulative | Text/map lookup path; neither a general generated-binary leaf nor a basis for a named-container rule. |
| Document-Audit | 0.35 s | registration/`DecodeNodeJSON` 28.6% cumulative; generator 20.0%; `String_contains` 14.3% | Its short process remains bootstrap-heavy even without the phase collector. The generic metadata boundary was already rejected on build-throughput and process guards. |
| Lexical-Rollup | 0.30 s | `fs_read_lines` 43.3% cumulative; generator 26.7%; `strings.Split` 20.0%; `bridge.SwapEnvIfNeeded` 10.0% | Filesystem/iterator work, with a small environment bridge branch that does not repeat as the same leaf elsewhere. |
| Channel-Rollup | 19.67 s | `bridge.currentGID`/`runtime.Stack` 56.7% cumulative; executor `Flush` 39.3% cumulative; `SwapEnv` 24.4%; `time.Sleep` 18.8% | A material compiled concurrency-executor boundary, but observed in only one concurrent application. |
| Fib control | 9.67 s | generated `fib` 99.9% flat | Recursive numeric control, not a shared lowering boundary. |

The profiles intentionally distinguish a generic *component* from sufficient
selection evidence. `currentGID`, environment swapping, and executor flushing
are runtime-general mechanisms, but only Channel-Rollup currently demonstrates
their material cost. Conversely, Word-Frequency's map descendants are material
only beneath its map/text workload, so they must not become HashMap-specific
compiler lowering.

## Next recommendation

Profile a second independently authored compiled concurrency program—start
with the current BinaryTrees `spawn`/`future_flush` application—using the same
normal, collector-free, output-checked merged-profile method, with
Lexical-Rollup as a sequential guard. Why: Channel-Rollup makes the executor
bridge the only remaining broadly plausible compiled target, but a change must
repeat below the same concrete `currentGID`/task/flush descendants in another
concurrent program before it can be considered general. The work entails
building and verifying the current BinaryTrees generated binary, collecting
bounded repeated profiles, comparing exact bridge/executor leaves against
Channel-Rollup, and only then evaluating a semantics-preserving executor
candidate across both concurrency programs and the sequential guard. Do not
special-case channels, BinaryTrees, task counts, or any nominal container.
