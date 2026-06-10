# Full compiled scorecard refreshed

Date: 2026-07-25

## Decision

Retain three general correctness/evidence changes discovered while refreshing
all 61 portable compiled applications:

1. `bench_compare_external` can forward repeatable compiled build arguments
   and, when requested, retains each inner `bench_perf` artifact directory.
   This makes strict `--no-fallbacks` measurement and dependency auditing part
   of the normal comparison contract.
2. `bench_perf` writes `/usr/bin/time` metrics to a dedicated file instead of
   sharing program stderr with concurrent Go GC traces. This removes byte-level
   interleaving such as `gc real 0.08` without changing the timed program.
3. Generated channel receive uses a non-blocking receive `select` instead of
   racing `len(ch) > 0` against a later direct receive. With multiple receivers,
   two workers could observe one buffered item; the loser then blocked outside
   the close-aware select after the producer closed the channel.

Retain the deterministic source guards for all three contracts. Retain no
application, benchmark, named-container, non-primitive nominal, canonical
stdlib, interpreter, bytecode, language, dependency, or WASM change.

The consolidated machine-readable scorecard is
`2026-07-25-full-compiled-scorecard-refreshed.json`.

## Governing protocol

- All 61 applications in the portable `coverage` catalog were compiled with
  `--no-fallbacks`.
- Source emission and generated Go compilation each retained an independent
  60-second limit. Each program process retained a 60-second limit.
- Every Able and Go row used five fresh processes, its catalog working
  directory, arguments, verifier, executor, logical CPU budget, and a resolved
  CPU from the fixed `5,10,15,11` pool.
- `GOMEMLIMIT=1GiB` and `GOGC=50` were shared across the Able refresh.
- All generated modules, binaries, caches, and temporary work stayed under
  disk-backed `/var/tmp`.
- The initial compiler SHA-256 was
  `8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`.
  The compiler after the channel correction was
  `6ea26d6ee1e9b6b5447371bb01adc8052738ec08b13177db28b972b8c30c7bcc`.

The authoritative consolidated cohort uses:

- 45 unaffected rows from the initial full pass;
- 14 channel-reaching rows refreshed after the receive correction; and
- clean timing-stream repair cohorts for Mutex Ledger and Concurrent Scene
  Tiles, which do not execute channel receive.

That produces 305/305 verified Able processes. The fresh Go reference also
produced 305/305 verified processes.

## Correctness findings

### Timing output interleaving

Mutex Ledger and Concurrent Scene Tiles each produced five correct,
verifier-approved outputs, but one timing sample was rejected. Their retained
stderr showed concurrent GC text prefixed directly to the `real` line.
Separating time output from program stderr produced clean independent 5/5
cohorts for both rows.

### Multi-receiver channel hang

Concurrent Text Index had one real 60-second timeout in the initial cohort.
A clean five-run retry was not sufficient evidence: a 20-process reused-binary
stress check reproduced three timeouts.

A SIGQUIT dump caught `future_flush()` waiting for one worker. The worker was
blocked on the direct receive immediately below `if len(ch) > 0`; the producer
and the other workers had finished. The length check and receive were not
atomic, so another receiver could consume the item between them.

The generated fast path now uses:

```go
select {
case value := <-ch:
    // return the immediately available value
default:
    // continue to the close-aware blocking path
}
```

The corrected strict binary passed a five-run governing cohort and 100/100
additional verifier-backed process executions with a five-second diagnostic
bound. All 14 catalog rows that reach channel operations were then rebuilt and
remeasured: 70/70 passed without timeout or metric loss.

## Scorecard summary

- Six of 61 rows meet the 1.052632x target: Binary Trees, Quicksort, Base64,
  JSON, Monte Carlo Pi, and Pidigits.
- Five rows are faster than Go; Pidigits is effectively at parity.
- Geometric-mean Able/Go ratio: **6.2847x**.
- Median Able/Go ratio: **6.9767x**.
- Range: **0.4730x-214.5455x**.
- The four largest absolute target misses are Tapelang Alphabet (+1.7821 s),
  K-Nucleotide (+1.6704 s), Sudoku Masks (+1.1878 s), and Mutex Work Queue
  (+0.9396 s).

## Consolidated rows

`Excess` is Able mean minus Go mean. `Cohort` identifies the raw measurement
source selected into the consolidated report.

| Benchmark | Able s | Go s | Able/Go | Excess s | Cohort | Dependencies | Interpreter |
| --- | ---: | ---: | ---: | ---: | --- | ---: | --- |
| `fib` | 3.624 | 3.3056 | 1.10x | 0.3184 | comparison | 96 | no |
| `binarytrees` | 10.4 | 10.8498 | 0.96x | -0.4498 | comparison | 96 | no |
| `matrixmultiply` | 1.162 | 1.0156 | 1.14x | 0.1464 | comparison | 96 | no |
| `quicksort` | 1.976 | 2.8002 | 0.71x | -0.8242 | comparison | 96 | no |
| `sudoku_masks` | 1.826 | 0.6382 | 2.86x | 1.1878 | comparison | 96 | no |
| `i_before_e` | 0.072 | 0.0681 | 1.06x | 0.0039 | comparison | 96 | no |
| `base64` | 2.23 | 2.7164 | 0.82x | -0.4864 | comparison | 119 | no |
| `binary_event_log` | 0.164 | 0.0084 | 19.52x | 0.1556 | comparison | 96 | no |
| `json` | 0.704 | 1.4883 | 0.47x | -0.7843 | comparison | 96 | no |
| `monte_carlo_pi` | 0.178 | 0.2084 | 0.85x | -0.0304 | comparison | 96 | no |
| `pidigits` | 1.174 | 1.1715 | 1.00x | 0.0025 | comparison | 96 | no |
| `mandelbrot` | 0.088 | 0.0473 | 1.86x | 0.0407 | comparison | 96 | no |
| `reverse_complement` | 0.052 | 0.0155 | 3.35x | 0.0365 | comparison | 96 | no |
| `k_nucleotide` | 1.726 | 0.0556 | 31.04x | 1.6704 | comparison | 96 | no |
| `nbody` | 0.08 | 0.0358 | 2.23x | 0.0442 | comparison | 96 | no |
| `tapelang_alphabet` | 3.794 | 2.0119 | 1.89x | 1.7821 | comparison | 96 | no |
| `distance_field` | 0.048 | 0.012 | 4.00x | 0.0360 | comparison | 96 | no |
| `rms_norm` | 0.032 | 0.0106 | 3.02x | 0.0214 | comparison | 96 | no |
| `fasta_generation` | 0.044 | 0.0132 | 3.33x | 0.0308 | comparison | 96 | no |
| `fixed_width_128` | 0.092 | 0.0054 | 17.04x | 0.0866 | comparison | 96 | no |
| `rational_series` | 0.066 | 0.0139 | 4.75x | 0.0521 | comparison | 96 | no |
| `wide_integer_records` | 0.082 | 0.024 | 3.42x | 0.0580 | comparison | 96 | no |
| `word_frequency` | 0.046 | 0.0069 | 6.67x | 0.0391 | comparison | 96 | no |
| `document_audit` | 0.042 | 0.0051 | 8.24x | 0.0369 | comparison | 96 | no |
| `lexical_rollup` | 0.054 | 0.0046 | 11.74x | 0.0494 | comparison | 96 | no |
| `channel_rollup` | 0.05 | 0.0057 | 8.77x | 0.0443 | channel-refresh | 96 | no |
| `future_pipeline` | 0.032 | 0.0052 | 6.15x | 0.0268 | channel-refresh | 96 | no |
| `future_await_race` | 0.054 | 0.0042 | 12.86x | 0.0498 | channel-refresh | 96 | no |
| `await_channel_mux` | 0.29 | 0.0049 | 59.18x | 0.2851 | channel-refresh | 96 | no |
| `mutex_ledger` | 0.094 | 0.0048 | 19.58x | 0.0892 | repair-a | 96 | no |
| `mutex_await_journal` | 0.43 | 0.0042 | 102.38x | 0.4258 | comparison | 96 | no |
| `mutex_work_queue` | 0.944 | 0.0044 | 214.55x | 0.9396 | comparison | 96 | no |
| `regex_suffix_audit` | 0.06 | 0.0062 | 9.68x | 0.0538 | comparison | 96 | no |
| `regex_set_audit` | 0.06 | 0.0053 | 11.32x | 0.0547 | comparison | 96 | no |
| `regex_stream_audit` | 0.07 | 0.0046 | 15.22x | 0.0654 | comparison | 96 | no |
| `log_routing_redaction` | 0.06 | 0.0045 | 13.33x | 0.0555 | comparison | 96 | no |
| `config_validation_extraction` | 0.042 | 0.0037 | 11.35x | 0.0383 | comparison | 96 | no |
| `unicode_scalar_pipeline` | 0.114 | 0.0094 | 12.13x | 0.1046 | comparison | 96 | no |
| `array_slice_window` | 0.03 | 0.0041 | 7.32x | 0.0259 | comparison | 96 | no |
| `dependency_plan` | 0.02 | 0.0036 | 5.56x | 0.0164 | comparison | 96 | no |
| `inventory_reconciliation` | 0.118 | 0.0081 | 14.57x | 0.1099 | comparison | 96 | no |
| `option_result_config` | 0.054 | 0.0047 | 11.49x | 0.0493 | comparison | 96 | no |
| `concurrent_text_index` | 0.04 | 0.0058 | 6.90x | 0.0342 | channel-refresh | 96 | no |
| `validated_job_pipeline` | 0.06 | 0.0038 | 15.79x | 0.0562 | channel-refresh | 96 | no |
| `dependency_wave_validation` | 0.042 | 0.0044 | 9.55x | 0.0376 | channel-refresh | 96 | no |
| `concurrent_event_routing` | 0.034 | 0.0056 | 6.07x | 0.0284 | channel-refresh | 96 | no |
| `concurrent_document_pipeline` | 0.028 | 0.0049 | 5.71x | 0.0231 | channel-refresh | 96 | no |
| `concurrent_stencil_reduction` | 0.032 | 0.0063 | 5.08x | 0.0257 | channel-refresh | 96 | no |
| `concurrent_signal_dispatch` | 0.038 | 0.0055 | 6.91x | 0.0325 | channel-refresh | 96 | no |
| `concurrent_transform_chain` | 0.034 | 0.0059 | 5.76x | 0.0281 | channel-refresh | 96 | no |
| `concurrent_policy_callbacks` | 0.032 | 0.0046 | 6.96x | 0.0274 | channel-refresh | 96 | no |
| `concurrent_graph_visitors` | 0.034 | 0.0046 | 7.39x | 0.0294 | comparison | 96 | no |
| `concurrent_audio_voices` | 0.03 | 0.0043 | 6.98x | 0.0257 | comparison | 96 | no |
| `concurrent_packet_codecs` | 0.03 | 0.0041 | 7.32x | 0.0259 | comparison | 96 | no |
| `concurrent_scene_tiles` | 0.03 | 0.0038 | 7.89x | 0.0262 | repair-a | 96 | no |
| `concurrent_tree_folds` | 0.024 | 0.0043 | 5.58x | 0.0197 | comparison | 96 | no |
| `concurrent_state_machines` | 0.034 | 0.004 | 8.50x | 0.0300 | comparison | 96 | no |
| `concurrent_stateful_pipeline` | 0.046 | 0.005 | 9.20x | 0.0410 | channel-refresh | 96 | no |
| `manifest_normalization` | 0.054 | 0.0045 | 12.00x | 0.0495 | comparison | 96 | no |
| `policy_record_dispatch` | 0.094 | 0.0047 | 20.00x | 0.0893 | comparison | 96 | no |
| `sensor_calibration` | 0.048 | 0.0055 | 8.73x | 0.0425 | comparison | 96 | no |

## Dependency and artifact verification

- All 61 strict generated modules were retained and audited.
- Sixty graphs contain 96 dependencies; Base64 contains 119 because of its
  host crypto dependencies.
- No graph contains `able/interpreter-go/pkg/interpreter`.
- Raw consolidated JSON SHA-256:
  `5ef3f0102dfa3a2be274dc0eb418a48ee794ffd7bfb79ce710d660bbac4ef554`.
- Raw Go-reference JSON SHA-256:
  `20cdb244cd4a4ca0e3ef45e1821ea9166a9a870d8544e24277e5bcaeada86569`.
- Dependency-audit TSV SHA-256:
  `ce6090d9a83f26fff959d24833abe346230a6057798640e3a780a76597c37620`.
- The oversized raw reports were represented by those hashes rather than
  checked in as files exceeding 1,000 lines.

Focused channel-generation, compiled concurrency parity, compiler CLI,
benchmark contract, JSON, syntax, and whitespace checks passed. No spec or
canonical-stdlib update was required.

## Next recommendation

Profile the four largest absolute target misses: Tapelang Alphabet,
K-Nucleotide, Sudoku Masks, and Mutex Work Queue. Use Pidigits as the
single-threaded parity control and Binary Trees as the concurrent parity
control.

Why: the four misses contribute 5.5799 seconds of absolute excess per
scorecard pass and cover unlike primitive-loop, text/hash, search/Array, and
mutex/task workloads. Ratio alone overweights short launch-dominated rows;
absolute excess selects costs large enough to produce stable profiles.

What it entails: rebuild all six rows strictly with the corrected compiler;
collect repeated main-only CPU and exact-allocation profiles; attribute hot
leaves back through generated callers; intersect only exact compiler or
generated-runtime owners; and advance a candidate only if one material owner
repeats across at least three unlike misses while remaining absent or
non-material in the controls. Any prototype then requires verifier-backed
repeated A/B/Go measurements.

Why it is important: only six rows currently meet the Go target, but the next
change must improve general lowering or runtime behavior rather than one
benchmark family. An exact owner intersection across the largest unlike
absolute misses is the shortest evidence-backed route toward native-carrier
Go parity without reopening closed boundary, nominal-special-case, or
benchmark-specific paths. Do not begin WASM work.
