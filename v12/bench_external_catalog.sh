#!/usr/bin/env bash

bench_external_default_suite() {
  printf '%s\n' "core"
}

bench_external_suite_names() {
  printf '%s\n' \
    "core" \
    "full" \
    "generality" \
    "coverage" \
    "numeric-structural" \
    "primitive-math" \
    "fixed-width" \
    "nominal-numeric" \
    "wide-integer-records" \
    "binary-event-log" \
    "backup-dedup" \
    "word-frequency" \
    "document-audit" \
    "lexical-rollup" \
    "channel-rollup" \
    "future-pipeline" \
    "future-await-race" \
    "await-channel-mux" \
    "mutex-ledger" \
    "mutex-await-journal" \
    "mutex-work-queue" \
    "concurrency" \
    "text-bytes" \
    "bulk-output" \
    "regex-text" \
    "regex-stream" \
    "log-routing-redaction" \
    "config-validation-extraction" \
    "unicode-scalars" \
    "array-slice-window" \
    "dependency-plan" \
    "discrete-event-simulation" \
    "inventory-reconciliation" \
    "option-result-config" \
    "concurrent-text-index" \
    "validated-job-pipeline" \
    "dependency-wave-validation" \
    "concurrent-event-routing" \
    "concurrent-document-pipeline" \
    "concurrent-stencil-reduction" \
    "concurrent-signal-dispatch" \
    "concurrent-transform-chain" \
    "concurrent-policy-callbacks" \
    "concurrent-graph-visitors" \
    "concurrent-audio-voices" \
    "concurrent-packet-codecs" \
    "concurrent-scene-tiles" \
    "concurrent-tree-folds" \
    "concurrent-state-machines" \
    "concurrent-stateful-pipeline" \
    "manifest-normalization" \
    "policy-record-dispatch" \
    "sensor-calibration" \
    "legacy-sudoku" \
    "sudoku-masks"
}

bench_external_diagnostic_suite_names() {
  printf '%s\n' "legacy-sudoku"
}

bench_external_suite_is_diagnostic() {
  case "$1" in
    legacy-sudoku)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

bench_external_suite_csv() {
  case "$1" in
    ""|core)
      printf '%s\n' "fib,binarytrees,matrixmultiply,quicksort,sudoku_masks,i_before_e"
      ;;
    fixed-width)
      printf '%s\n' "fixed_width_128"
      ;;
    nominal-numeric)
      printf '%s\n' "fixed_width_128,rational_series,wide_integer_records"
      ;;
    wide-integer-records)
      printf '%s\n' "wide_integer_records"
      ;;
    binary-event-log)
      printf '%s\n' "binary_event_log"
      ;;
    backup-dedup)
      printf '%s\n' "backup_dedup"
      ;;
    primitive-math)
      printf '%s\n' "distance_field,rms_norm,nbody"
      ;;
    word-frequency)
      printf '%s\n' "word_frequency"
      ;;
    document-audit)
      printf '%s\n' "document_audit"
      ;;
    lexical-rollup)
      printf '%s\n' "lexical_rollup"
      ;;
    channel-rollup)
      printf '%s\n' "channel_rollup"
      ;;
    future-pipeline)
      printf '%s\n' "future_pipeline"
      ;;
    future-await-race)
      printf '%s\n' "future_await_race"
      ;;
    await-channel-mux)
      printf '%s\n' "await_channel_mux"
      ;;
    mutex-ledger)
      printf '%s\n' "mutex_ledger"
      ;;
    mutex-await-journal)
      printf '%s\n' "mutex_await_journal"
      ;;
    mutex-work-queue)
      printf '%s\n' "mutex_work_queue"
      ;;
    concurrency)
      printf '%s\n' "binarytrees,channel_rollup,future_pipeline,future_await_race,await_channel_mux,mutex_ledger,mutex_await_journal,mutex_work_queue,concurrent_text_index,validated_job_pipeline,dependency_wave_validation,concurrent_event_routing,concurrent_document_pipeline,concurrent_stencil_reduction,concurrent_signal_dispatch,concurrent_transform_chain,concurrent_policy_callbacks,concurrent_graph_visitors,concurrent_audio_voices,concurrent_packet_codecs,concurrent_scene_tiles,concurrent_tree_folds,concurrent_state_machines,concurrent_stateful_pipeline"
      ;;
    full|generality)
      printf '%s\n' "fib,binarytrees,matrixmultiply,quicksort,sudoku_masks,i_before_e,base64,json,monte_carlo_pi,pidigits,mandelbrot,reverse_complement,k_nucleotide,nbody,tapelang_alphabet,distance_field,rms_norm,fasta_generation"
      ;;
    coverage)
      # Every portable candidate-selection application. The historical
      # scan-based Sudoku source remains directly addressable through
      # `legacy-sudoku`, but is not comparable with the exact-cover Go/Python
      # references and cannot complete within the bounded scorecard protocol.
      printf '%s\n' "fib,binarytrees,matrixmultiply,quicksort,sudoku_masks,i_before_e,base64,backup_dedup,binary_event_log,json,monte_carlo_pi,pidigits,mandelbrot,reverse_complement,k_nucleotide,nbody,tapelang_alphabet,distance_field,rms_norm,fasta_generation,fixed_width_128,rational_series,wide_integer_records,word_frequency,document_audit,lexical_rollup,channel_rollup,future_pipeline,future_await_race,await_channel_mux,mutex_ledger,mutex_await_journal,mutex_work_queue,regex_suffix_audit,regex_set_audit,regex_stream_audit,log_routing_redaction,config_validation_extraction,unicode_scalar_pipeline,array_slice_window,dependency_plan,discrete_event_simulation,inventory_reconciliation,option_result_config,concurrent_text_index,validated_job_pipeline,dependency_wave_validation,concurrent_event_routing,concurrent_document_pipeline,concurrent_stencil_reduction,concurrent_signal_dispatch,concurrent_transform_chain,concurrent_policy_callbacks,concurrent_graph_visitors,concurrent_audio_voices,concurrent_packet_codecs,concurrent_scene_tiles,concurrent_tree_folds,concurrent_state_machines,concurrent_stateful_pipeline,manifest_normalization,policy_record_dispatch,sensor_calibration"
      ;;
    numeric-structural)
      printf '%s\n' "fib,binarytrees,matrixmultiply,mandelbrot,monte_carlo_pi,pidigits,nbody,distance_field,rms_norm"
      ;;
    text-bytes)
      printf '%s\n' "base64,binary_event_log,json,i_before_e,reverse_complement,k_nucleotide,quicksort,tapelang_alphabet,fasta_generation"
      ;;
    bulk-output)
      printf '%s\n' "mandelbrot,reverse_complement,fasta_generation"
      ;;
    regex-text)
      printf '%s\n' "regex_suffix_audit,regex_set_audit,regex_stream_audit,log_routing_redaction,config_validation_extraction"
      ;;
    regex-stream)
      printf '%s\n' "regex_stream_audit"
      ;;
    log-routing-redaction)
      printf '%s\n' "log_routing_redaction"
      ;;
    config-validation-extraction)
      printf '%s\n' "config_validation_extraction"
      ;;
    unicode-scalars)
      printf '%s\n' "unicode_scalar_pipeline"
      ;;
    array-slice-window)
      printf '%s\n' "array_slice_window"
      ;;
    dependency-plan)
      printf '%s\n' "dependency_plan"
      ;;
    discrete-event-simulation)
      printf '%s\n' "discrete_event_simulation"
      ;;
    inventory-reconciliation)
      printf '%s\n' "inventory_reconciliation"
      ;;
    option-result-config)
      printf '%s\n' "option_result_config"
      ;;
    concurrent-text-index)
      printf '%s\n' "concurrent_text_index"
      ;;
    validated-job-pipeline)
      printf '%s\n' "validated_job_pipeline"
      ;;
    dependency-wave-validation)
      printf '%s\n' "dependency_wave_validation"
      ;;
    concurrent-event-routing)
      printf '%s\n' "concurrent_event_routing"
      ;;
    concurrent-document-pipeline)
      printf '%s\n' "concurrent_document_pipeline"
      ;;
    concurrent-stencil-reduction)
      printf '%s\n' "concurrent_stencil_reduction"
      ;;
    concurrent-signal-dispatch)
      printf '%s\n' "concurrent_signal_dispatch"
      ;;
    concurrent-transform-chain)
      printf '%s\n' "concurrent_transform_chain"
      ;;
    concurrent-policy-callbacks)
      printf '%s\n' "concurrent_policy_callbacks"
      ;;
    concurrent-graph-visitors)
      printf '%s\n' "concurrent_graph_visitors"
      ;;
    concurrent-audio-voices)
      printf '%s\n' "concurrent_audio_voices"
      ;;
    concurrent-packet-codecs)
      printf '%s\n' "concurrent_packet_codecs"
      ;;
    concurrent-scene-tiles)
      printf '%s\n' "concurrent_scene_tiles"
      ;;
    concurrent-tree-folds)
      printf '%s\n' "concurrent_tree_folds"
      ;;
    concurrent-state-machines)
      printf '%s\n' "concurrent_state_machines"
      ;;
    concurrent-stateful-pipeline)
      printf '%s\n' "concurrent_stateful_pipeline"
      ;;
    manifest-normalization)
      printf '%s\n' "manifest_normalization"
      ;;
    policy-record-dispatch)
      printf '%s\n' "policy_record_dispatch"
      ;;
    sensor-calibration)
      printf '%s\n' "sensor_calibration"
      ;;
    legacy-sudoku)
      printf '%s\n' "sudoku"
      ;;
    sudoku-masks)
      printf '%s\n' "sudoku_masks"
      ;;
    *)
      return 1
      ;;
  esac
}

bench_external_target() {
  local root="$1"
  local bench="$2"
  case "$bench" in
    fib)
      printf '%s\n' "$root/examples/benchmarks/fib.able"
      ;;
    fixed_width_128)
      printf '%s\n' "$root/examples/benchmarks/fixed_width_128/fixed_width_128.able"
      ;;
    rational_series)
      printf '%s\n' "$root/examples/benchmarks/rational_series/rational_series.able"
      ;;
    wide_integer_records)
      printf '%s\n' "$root/examples/benchmarks/wide_integer_records/wide_integer_records.able"
      ;;
    word_frequency)
      printf '%s\n' "$root/examples/benchmarks/word_frequency/word_frequency.able"
      ;;
    document_audit)
      printf '%s\n' "$root/examples/benchmarks/document_audit/document_audit.able"
      ;;
    lexical_rollup)
      printf '%s\n' "$root/examples/benchmarks/lexical_rollup/lexical_rollup.able"
      ;;
    channel_rollup)
      printf '%s\n' "$root/examples/benchmarks/channel_rollup/channel_rollup.able"
      ;;
    future_pipeline)
      printf '%s\n' "$root/examples/benchmarks/future_pipeline/future_pipeline.able"
      ;;
    future_await_race)
      printf '%s\n' "$root/examples/benchmarks/future_await_race/future_await_race.able"
      ;;
    await_channel_mux)
      printf '%s\n' "$root/examples/benchmarks/await_channel_mux/await_channel_mux.able"
      ;;
    mutex_ledger)
      printf '%s\n' "$root/examples/benchmarks/mutex_ledger/mutex_ledger.able"
      ;;
    mutex_await_journal)
      printf '%s\n' "$root/examples/benchmarks/mutex_await_journal/mutex_await_journal.able"
      ;;
    mutex_work_queue)
      printf '%s\n' "$root/examples/benchmarks/mutex_work_queue/mutex_work_queue.able"
      ;;
    regex_suffix_audit)
      printf '%s\n' "$root/examples/benchmarks/regex_suffix_audit/regex_suffix_audit.able"
      ;;
    regex_set_audit)
      printf '%s\n' "$root/examples/benchmarks/regex_set_audit/regex_set_audit.able"
      ;;
    regex_stream_audit)
      printf '%s\n' "$root/examples/benchmarks/regex_stream_audit/regex_stream_audit.able"
      ;;
    log_routing_redaction)
      printf '%s\n' "$root/examples/benchmarks/log_routing_redaction/log_routing_redaction.able"
      ;;
    config_validation_extraction)
      printf '%s\n' "$root/examples/benchmarks/config_validation_extraction/config_validation_extraction.able"
      ;;
    unicode_scalar_pipeline)
      printf '%s\n' "$root/examples/benchmarks/unicode_scalar_pipeline/unicode_scalar_pipeline.able"
      ;;
    array_slice_window)
      printf '%s\n' "$root/examples/benchmarks/array_slice_window/array_slice_window.able"
      ;;
    dependency_plan)
      printf '%s\n' "$root/examples/benchmarks/dependency_plan/dependency_plan.able"
      ;;
    discrete_event_simulation)
      printf '%s\n' "$root/examples/benchmarks/discrete_event_simulation/discrete_event_simulation.able"
      ;;
    inventory_reconciliation)
      printf '%s\n' "$root/examples/benchmarks/inventory_reconciliation/inventory_reconciliation.able"
      ;;
    option_result_config)
      printf '%s\n' "$root/examples/benchmarks/option_result_config/option_result_config.able"
      ;;
    concurrent_text_index)
      printf '%s\n' "$root/examples/benchmarks/concurrent_text_index/concurrent_text_index.able"
      ;;
    validated_job_pipeline)
      printf '%s\n' "$root/examples/benchmarks/validated_job_pipeline/validated_job_pipeline.able"
      ;;
    dependency_wave_validation)
      printf '%s\n' "$root/examples/benchmarks/dependency_wave_validation/dependency_wave_validation.able"
      ;;
    concurrent_event_routing)
      printf '%s\n' "$root/examples/benchmarks/concurrent_event_routing/concurrent_event_routing.able"
      ;;
    concurrent_document_pipeline)
      printf '%s\n' "$root/examples/benchmarks/concurrent_document_pipeline/concurrent_document_pipeline.able"
      ;;
    concurrent_stencil_reduction)
      printf '%s\n' "$root/examples/benchmarks/concurrent_stencil_reduction/concurrent_stencil_reduction.able"
      ;;
    concurrent_signal_dispatch)
      printf '%s\n' "$root/examples/benchmarks/concurrent_signal_dispatch/concurrent_signal_dispatch.able"
      ;;
    concurrent_transform_chain)
      printf '%s\n' "$root/examples/benchmarks/concurrent_transform_chain/concurrent_transform_chain.able"
      ;;
    concurrent_policy_callbacks)
      printf '%s\n' "$root/examples/benchmarks/concurrent_policy_callbacks/concurrent_policy_callbacks.able"
      ;;
    concurrent_graph_visitors)
      printf '%s\n' "$root/examples/benchmarks/concurrent_graph_visitors/concurrent_graph_visitors.able"
      ;;
    concurrent_audio_voices)
      printf '%s\n' "$root/examples/benchmarks/concurrent_audio_voices/concurrent_audio_voices.able"
      ;;
    concurrent_packet_codecs)
      printf '%s\n' "$root/examples/benchmarks/concurrent_packet_codecs/concurrent_packet_codecs.able"
      ;;
    concurrent_scene_tiles)
      printf '%s\n' "$root/examples/benchmarks/concurrent_scene_tiles/concurrent_scene_tiles.able"
      ;;
    concurrent_tree_folds)
      printf '%s\n' "$root/examples/benchmarks/concurrent_tree_folds/concurrent_tree_folds.able"
      ;;
    concurrent_state_machines)
      printf '%s\n' "$root/examples/benchmarks/concurrent_state_machines/concurrent_state_machines.able"
      ;;
    concurrent_stateful_pipeline)
      printf '%s\n' "$root/examples/benchmarks/concurrent_stateful_pipeline/concurrent_stateful_pipeline.able"
      ;;
    manifest_normalization)
      printf '%s\n' "$root/examples/benchmarks/manifest_normalization/manifest_normalization.able"
      ;;
    policy_record_dispatch)
      printf '%s\n' "$root/examples/benchmarks/policy_record_dispatch/policy_record_dispatch.able"
      ;;
    sensor_calibration)
      printf '%s\n' "$root/examples/benchmarks/sensor_calibration/sensor_calibration.able"
      ;;
    binarytrees)
      printf '%s\n' "$root/examples/benchmarks/binarytrees.able"
      ;;
    matrixmultiply)
      printf '%s\n' "$root/examples/benchmarks/matrixmultiply.able"
      ;;
    base64)
      printf '%s\n' "$root/examples/benchmarks/base64/base64.able"
      ;;
    binary_event_log)
      printf '%s\n' "$root/examples/benchmarks/binary_event_log/binary_event_log.able"
      ;;
    backup_dedup)
      printf '%s\n' "$root/examples/benchmarks/backup_dedup/backup_dedup.able"
      ;;
    mandelbrot)
      printf '%s\n' "$root/examples/benchmarks/mandelbrot/mandelbrot.able"
      ;;
    json)
      printf '%s\n' "$root/examples/benchmarks/json/json.able"
      ;;
    monte_carlo_pi)
      printf '%s\n' "$root/examples/benchmarks/monte_carlo_pi/monte_carlo_pi.able"
      ;;
    nbody)
      printf '%s\n' "$root/examples/benchmarks/nbody.able"
      ;;
    distance_field)
      printf '%s\n' "$root/examples/benchmarks/distance_field/distance_field.able"
      ;;
    rms_norm)
      printf '%s\n' "$root/examples/benchmarks/rms_norm/rms_norm.able"
      ;;
    fasta_generation)
      printf '%s\n' "$root/examples/benchmarks/fasta_generation/fasta_generation.able"
      ;;
    pidigits)
      printf '%s\n' "$root/examples/benchmarks/pidigits/pidigits.able"
      ;;
    k_nucleotide)
      printf '%s\n' "$root/examples/benchmarks/k_nucleotide/k_nucleotide.able"
      ;;
    reverse_complement)
      printf '%s\n' "$root/examples/benchmarks/reverse_complement/reverse_complement.able"
      ;;
    tapelang_alphabet)
      printf '%s\n' "$root/examples/benchmarks/tapelang_alphabet/tapelang_alphabet.able"
      ;;
    i_before_e)
      printf '%s\n' "$root/examples/benchmarks/i_before_e/i_before_e.able"
      ;;
    quicksort)
      printf '%s\n' "$root/examples/benchmarks/quicksort/quicksort.able"
      ;;
    sudoku)
      printf '%s\n' "$root/examples/benchmarks/sudoku/sudoku.able"
      ;;
    sudoku_masks)
      printf '%s\n' "$root/examples/benchmarks/sudoku_masks/sudoku_masks.able"
      ;;
    *)
      return 1
      ;;
  esac
}

bench_external_program_args() {
  case "$1" in
    matrixmultiply)
      # The external suite Dockerfiles all run the canonical 1000x1000 input;
      # pass the same input to fresh local Go and Able comparison processes.
      printf '%s\n' "1000"
      ;;
    i_before_e)
      printf '%s\n' "wordlist.txt"
      ;;
    pidigits)
      printf '%s\n' "10000"
      ;;
    k_nucleotide)
      printf '%s\n' "knucleotide-input.fasta"
      ;;
    reverse_complement)
      printf '%s\n' "reverse-complement-input.fasta"
      ;;
    tapelang_alphabet)
      printf '%s\n' "benchmark.tape"
      ;;
    word_frequency)
      printf '%s\n' "corpus.md"
      ;;
    document_audit)
      printf '%s\n' "word-frequency/corpus.md"
      ;;
    lexical_rollup)
      printf '%s\n' "i-before-e/wordlist.txt"
      ;;
    channel_rollup|concurrent_text_index)
      printf '%s\n' "i-before-e/wordlist.txt"
      ;;
    concurrent_event_routing)
      printf '%s\n' "events.txt"
      ;;
    validated_job_pipeline)
      printf '%s\n' "jobs.txt"
      ;;
    binary_event_log)
      printf '%s\n' "events.bin"
      ;;
    backup_dedup)
      printf '%s\n' "word-frequency/corpus.md"
      ;;
    concurrent_document_pipeline)
      printf '%s\n' "documents.txt"
      ;;
    concurrent_stencil_reduction)
      printf '%s\n' "samples.txt"
      ;;
    concurrent_signal_dispatch)
      printf '%s\n' "signals.txt"
      ;;
    concurrent_transform_chain)
      printf '%s\n' "samples.txt"
      ;;
    concurrent_policy_callbacks)
      printf '%s\n' "policies.txt"
      ;;
    manifest_normalization)
      printf '%s\n' "manifests.txt"
      ;;
    policy_record_dispatch)
      printf '%s\n' "records.txt"
      ;;
    sensor_calibration)
      printf '%s\n' "readings.txt"
      ;;
    regex_suffix_audit)
      printf '%s\n' "i-before-e/wordlist.txt"
      ;;
    regex_set_audit)
      printf '%s\n' "i-before-e/wordlist.txt"
      ;;
    regex_stream_audit)
      printf '%s\n' "i-before-e/wordlist.txt"
      ;;
    log_routing_redaction)
      printf '%s\n' "logs.txt"
      ;;
    config_validation_extraction)
      printf '%s\n' "deployments.cfg"
      ;;
    *)
      return 0
      ;;
  esac
}

bench_external_executor() {
  case "$1" in
    binarytrees|channel_rollup|future_pipeline|future_await_race|await_channel_mux|mutex_ledger|mutex_await_journal|mutex_work_queue|concurrent_text_index|validated_job_pipeline|dependency_wave_validation|concurrent_event_routing|concurrent_document_pipeline|concurrent_stencil_reduction|concurrent_signal_dispatch|concurrent_transform_chain|concurrent_policy_callbacks|concurrent_graph_visitors|concurrent_audio_voices|concurrent_packet_codecs|concurrent_scene_tiles|concurrent_tree_folds|concurrent_state_machines|concurrent_stateful_pipeline)
      printf '%s\n' "goroutine"
      ;;
    *)
      return 0
      ;;
  esac
}

# Performance contracts are mode-aware. Compiled applications whose matched
# sources intentionally exercise parallel scheduling receive four logical
# CPUs. Interpreter comparisons remain single-CPU so Able bytecode, Python,
# and Ruby measure the same serial execution budget even when the Able program
# uses the goroutine executor for language-level concurrency semantics.
bench_external_logical_cpu_budget() {
  local benchmark="$1"
  local mode="$2"
  case "$mode" in
    compiled)
      case "$benchmark" in
        binarytrees|channel_rollup|future_pipeline|future_await_race|await_channel_mux|mutex_ledger|mutex_await_journal|mutex_work_queue|concurrent_text_index|validated_job_pipeline|dependency_wave_validation|concurrent_event_routing|concurrent_document_pipeline|concurrent_stencil_reduction|concurrent_signal_dispatch|concurrent_transform_chain|concurrent_policy_callbacks|concurrent_graph_visitors|concurrent_audio_voices|concurrent_packet_codecs|concurrent_scene_tiles|concurrent_tree_folds|concurrent_state_machines|concurrent_stateful_pipeline)
          printf '%s\n' "4"
          ;;
        *)
          printf '%s\n' "1"
          ;;
      esac
      ;;
    bytecode|bytecode-prechecked|bytecode-runtime|treewalker)
      printf '%s\n' "1"
      ;;
    *)
      return 1
      ;;
  esac
}

bench_external_executor_policy() {
  local executor
  executor="$(bench_external_executor "$1" || true)"
  printf '%s\n' "${executor:-serial}"
}

bench_external_effective_cpu_pool() {
  local requested_pool="$1"
  if [[ -n "$requested_pool" ]]; then
    printf '%s\n' "$requested_pool"
    return 0
  fi
  command -v taskset >/dev/null 2>&1 || return 1
  taskset -pc "$$" 2>/dev/null | awk -F: 'NR == 1 { sub(/^[[:space:]]+/, "", $2); print $2 }'
}

bench_external_expand_cpu_pool() {
  local raw_pool="$1"
  local part start end cpu
  local -A seen=()
  local expanded=()
  IFS=',' read -r -a parts <<< "$raw_pool"
  for part in "${parts[@]}"; do
    [[ -n "$part" ]] || return 1
    if [[ "$part" =~ ^([0-9]+)-([0-9]+)$ ]]; then
      start="${BASH_REMATCH[1]}"
      end="${BASH_REMATCH[2]}"
      ((10#$start <= 10#$end)) || return 1
      for ((cpu = 10#$start; cpu <= 10#$end; cpu++)); do
        if [[ -z "${seen[$cpu]:-}" ]]; then
          seen[$cpu]=1
          expanded+=("$cpu")
        fi
      done
    elif [[ "$part" =~ ^[0-9]+$ ]]; then
      cpu=$((10#$part))
      if [[ -z "${seen[$cpu]:-}" ]]; then
        seen[$cpu]=1
        expanded+=("$cpu")
      fi
    else
      return 1
    fi
  done
  ((${#expanded[@]} > 0)) || return 1
  printf '%s\n' "${expanded[@]}"
}

bench_external_resolve_cpu_affinity() {
  local cpu_pool="$1"
  local logical_cpu_budget="$2"
  [[ "$logical_cpu_budget" =~ ^[1-9][0-9]*$ ]] || return 1
  local cpus=()
  while IFS= read -r cpu; do
    [[ -n "$cpu" ]] || continue
    cpus+=("$cpu")
  done < <(bench_external_expand_cpu_pool "$cpu_pool")
  ((${#cpus[@]} >= logical_cpu_budget)) || return 1
  local selected=("${cpus[@]:0:logical_cpu_budget}")
  local IFS=','
  printf '%s\n' "${selected[*]}"
}

# These applications have a sibling run.able (and, for some, sibling inputs)
# with the same package name as their canonical Able entry source. The sibling
# benchmark directory must not also become a source root.
bench_external_source_root_only() {
  case "$1" in
    fixed_width_128|rational_series|wide_integer_records|backup_dedup|future_pipeline|future_await_race|await_channel_mux|mutex_ledger|mutex_await_journal|mutex_work_queue|regex_suffix_audit|regex_set_audit|regex_stream_audit|log_routing_redaction|config_validation_extraction|unicode_scalar_pipeline|array_slice_window|dependency_plan|discrete_event_simulation|inventory_reconciliation|option_result_config|distance_field|rms_norm|fasta_generation|concurrent_text_index|validated_job_pipeline|dependency_wave_validation|concurrent_event_routing|concurrent_stencil_reduction|concurrent_signal_dispatch|concurrent_transform_chain|concurrent_policy_callbacks|concurrent_graph_visitors|concurrent_audio_voices|concurrent_packet_codecs|concurrent_scene_tiles|concurrent_tree_folds|concurrent_state_machines|concurrent_stateful_pipeline|policy_record_dispatch|sensor_calibration)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

bench_external_result_name() {
  case "$1" in
    i_before_e)
      printf '%s\n' "i-before-e"
      ;;
    k_nucleotide)
      printf '%s\n' "k-nucleotide"
      ;;
    reverse_complement)
      printf '%s\n' "reverse-complement"
      ;;
    tapelang_alphabet)
      printf '%s\n' "tapelang-alphabet"
      ;;
    distance_field)
      printf '%s\n' "distance-field"
      ;;
    rms_norm)
      printf '%s\n' "rms-norm"
      ;;
    fasta_generation)
      printf '%s\n' "fasta-generation"
      ;;
    fixed_width_128)
      printf '%s\n' "fixed-width-128"
      ;;
    rational_series)
      printf '%s\n' "rational-series"
      ;;
    wide_integer_records)
      printf '%s\n' "wide-integer-records"
      ;;
    binary_event_log)
      printf '%s\n' "binary-event-log"
      ;;
    backup_dedup)
      printf '%s\n' "backup-dedup"
      ;;
    word_frequency)
      printf '%s\n' "word-frequency"
      ;;
    document_audit)
      printf '%s\n' "document-audit"
      ;;
    lexical_rollup)
      printf '%s\n' "lexical-rollup"
      ;;
    channel_rollup)
      printf '%s\n' "channel-rollup"
      ;;
    future_pipeline)
      printf '%s\n' "future-pipeline"
      ;;
    future_await_race)
      printf '%s\n' "future-await-race"
      ;;
    await_channel_mux)
      printf '%s\n' "await-channel-mux"
      ;;
    mutex_ledger)
      printf '%s\n' "mutex-ledger"
      ;;
    mutex_await_journal)
      printf '%s\n' "mutex-await-journal"
      ;;
    mutex_work_queue)
      printf '%s\n' "mutex-work-queue"
      ;;
    regex_suffix_audit)
      printf '%s\n' "regex-suffix-audit"
      ;;
    regex_set_audit)
      printf '%s\n' "regex-set-audit"
      ;;
    regex_stream_audit)
      printf '%s\n' "regex-stream-audit"
      ;;
    log_routing_redaction)
      printf '%s\n' "log-routing-redaction"
      ;;
    config_validation_extraction)
      printf '%s\n' "config-validation-extraction"
      ;;
    unicode_scalar_pipeline)
      printf '%s\n' "unicode-scalar-pipeline"
      ;;
    array_slice_window)
      printf '%s\n' "array-slice-window"
      ;;
    dependency_plan)
      printf '%s\n' "dependency-plan"
      ;;
    discrete_event_simulation)
      printf '%s\n' "discrete-event-simulation"
      ;;
    inventory_reconciliation)
      printf '%s\n' "inventory-reconciliation"
      ;;
    option_result_config)
      printf '%s\n' "option-result-config"
      ;;
    concurrent_text_index)
      printf '%s\n' "concurrent-text-index"
      ;;
    validated_job_pipeline)
      printf '%s\n' "validated-job-pipeline"
      ;;
    dependency_wave_validation)
      printf '%s\n' "dependency-wave-validation"
      ;;
    concurrent_event_routing)
      printf '%s\n' "concurrent-event-routing"
      ;;
    concurrent_document_pipeline)
      printf '%s\n' "concurrent-document-pipeline"
      ;;
    concurrent_stencil_reduction)
      printf '%s\n' "concurrent-stencil-reduction"
      ;;
    concurrent_signal_dispatch)
      printf '%s\n' "concurrent-signal-dispatch"
      ;;
    concurrent_transform_chain)
      printf '%s\n' "concurrent-transform-chain"
      ;;
    concurrent_policy_callbacks)
      printf '%s\n' "concurrent-policy-callbacks"
      ;;
    concurrent_graph_visitors)
      printf '%s\n' "concurrent-graph-visitors"
      ;;
    concurrent_audio_voices)
      printf '%s\n' "concurrent-audio-voices"
      ;;
    concurrent_packet_codecs)
      printf '%s\n' "concurrent-packet-codecs"
      ;;
    concurrent_scene_tiles)
      printf '%s\n' "concurrent-scene-tiles"
      ;;
    concurrent_tree_folds)
      printf '%s\n' "concurrent-tree-folds"
      ;;
    concurrent_state_machines)
      printf '%s\n' "concurrent-state-machines"
      ;;
    concurrent_stateful_pipeline)
      printf '%s\n' "concurrent-stateful-pipeline"
      ;;
    manifest_normalization)
      printf '%s\n' "manifest-normalization"
      ;;
    policy_record_dispatch)
      printf '%s\n' "policy-record-dispatch"
      ;;
    sensor_calibration)
      printf '%s\n' "sensor-calibration"
      ;;
    sudoku_masks)
      printf '%s\n' "sudoku-masks"
      ;;
    *)
      printf '%s\n' "$1"
      ;;
  esac
}

bench_external_suite_dir() {
  local repo="$1"
  local bench="$2"
  printf '%s\n' "$repo/$(bench_external_result_name "$bench")"
}

# Most applications use assets colocated with their implementation. The shared-
# input applications deliberately reuse sibling-suite inputs, so their catalog
# arguments are repository-relative and must run from the benchmark root.
bench_external_run_dir() {
  local repo="$1"
  local bench="$2"
  case "$bench" in
    backup_dedup|document_audit|lexical_rollup|channel_rollup|regex_suffix_audit|regex_set_audit|regex_stream_audit|sudoku_masks|concurrent_text_index)
      printf '%s\n' "$repo"
      ;;
    *)
      bench_external_suite_dir "$repo" "$bench"
      ;;
  esac
}

# The catalog is normally sourced by benchmark scripts. Its small direct CLI
# keeps non-shell report tooling on the same authoritative target mapping
# rather than duplicating benchmark-to-source paths in another language.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  catalog_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  case "${1:-}" in
    targets)
      shift
      (($# > 0)) || {
        echo "usage: $0 targets BENCHMARK [...]" >&2
        exit 2
      }
      for benchmark in "$@"; do
        target="$(bench_external_target "$catalog_dir" "$benchmark")" || {
          echo "bench_external_catalog: unknown benchmark: $benchmark" >&2
          exit 2
        }
        printf '%s\t%s\n' "$benchmark" "$target"
      done
      ;;
    contracts)
      shift
      (($# > 1)) || {
        echo "usage: $0 contracts BENCHMARK_REPO BENCHMARK [...]" >&2
        exit 2
      }
      benchmark_repo="$(realpath "$1")"
      shift
      [[ -d "$benchmark_repo" ]] || {
        echo "bench_external_catalog: benchmark repository not found: $benchmark_repo" >&2
        exit 2
      }
      for benchmark in "$@"; do
        run_dir="$(bench_external_run_dir "$benchmark_repo" "$benchmark")"
        suite_dir="$(bench_external_suite_dir "$benchmark_repo" "$benchmark")"
        verifier="$suite_dir/verify.rb"
        [[ -d "$run_dir" ]] || {
          echo "bench_external_catalog: missing run directory for $benchmark: $run_dir" >&2
          exit 2
        }
        arguments=()
        while IFS= read -r argument; do
          [[ -n "$argument" ]] || continue
          arguments+=("$argument")
        done < <(bench_external_program_args "$benchmark")
        joined_arguments=""
        for argument in "${arguments[@]}"; do
          joined_arguments+="${joined_arguments:+$'\x1f'}$argument"
        done
        printf '%s\t%s\t%s\t%s\n' "$benchmark" "$run_dir" "$verifier" "$joined_arguments"
      done
      ;;
    execution-contracts)
      shift
      (($# > 1)) || {
        echo "usage: $0 execution-contracts MODE BENCHMARK [...]" >&2
        exit 2
      }
      mode="$1"
      shift
      for benchmark in "$@"; do
        logical_cpu_budget="$(bench_external_logical_cpu_budget "$benchmark" "$mode")" || {
          echo "bench_external_catalog: unsupported execution mode: $mode" >&2
          exit 2
        }
        executor_policy="$(bench_external_executor_policy "$benchmark")"
        printf '%s\t%s\t%s\t%s\n' \
          "$benchmark" "$mode" "$logical_cpu_budget" "$executor_policy"
      done
      ;;
    resolve-cpus)
      shift
      (($# == 2)) || {
        echo "usage: $0 resolve-cpus CPU_POOL LOGICAL_CPU_BUDGET" >&2
        exit 2
      }
      bench_external_resolve_cpu_affinity "$1" "$2" || {
        echo "bench_external_catalog: CPU pool cannot satisfy logical CPU budget" >&2
        exit 2
      }
      ;;
    *)
      echo "usage: $0 {targets|contracts|execution-contracts|resolve-cpus} ..." >&2
      exit 2
      ;;
  esac
fi
