#!/usr/bin/env bash

bench_fixture_all_csv() {
  find "$1/fixtures/bench" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort | paste -sd, -
}

bench_fixture_suite_names() {
  printf '%s\n' \
    "fixture-core" \
    "fixture-full" \
    "fixture-generality" \
    "fixture-collections" \
    "fixture-text" \
    "fixture-algorithms" \
    "fixture-concurrency" \
    "fixture-numeric" \
    "fixture-external-small"
}

bench_fixture_suite_csv() {
  local root="$1"
  case "$2" in
    fixture-core)
      printf '%s\n' "array_map_i32_small,channel_roundtrip_i32_small,dijkstra_heap_small,future_fanout_i32_small,graph_bfs_small,hash_set_i32_small,iterator_match_identifier_small,md5_hex_small,queue_i32_small,regex_is_match_small,string_builder_small,union_find_small,word_count_small"
      ;;
    fixture-full)
      bench_fixture_all_csv "$root"
      ;;
    fixture-generality)
      printf '%s\n' "array_map_i32_small,boolean_reconciliation_small,channel_roundtrip_i32_small,deque_i32_small,dijkstra_heap_small,future_fanout_i32_small,heap_i32_small,nbody_small,persistent_set_i32_small,persistent_sorted_set_i32_small,random_lcg_i64_small,regex_is_match_small,string_builder_small,sum_u32_small,word_count_small,zigzag_char_small"
      ;;
    fixture-collections)
      printf '%s\n' "array_filter_i32_small,array_fold_i32_small,array_map_i32_small,bit_set_small,concurrent_queue_i32_small,deque_i32_small,hash_set_i32_small,hashmap_i32_small,heap_i32_small,iterator_match_identifier_small,lazy_seq_cache_i32_small,lazy_seq_take_i32_small,linked_list_enumerable_i32_small,linked_list_for_i32_small,linked_list_iterator_collect_i64_small,linked_list_iterator_filter_map_i64_small,linked_list_iterator_pipeline_i64_small,list_i32_small,persistent_map_i32_small,persistent_map_string_small,persistent_queue_i32_small,persistent_set_i32_small,persistent_sorted_set_i32_small,queue_i32_small,tree_map_i32_small,tree_set_i32_small,vector_i32_small"
      ;;
    fixture-text)
      printf '%s\n' "ascii_lower_small,automata_dfa_small,base64_roundtrip_small,byte_histogram_small,i_before_e_small,json_means_small,k_nucleotide_small,md5_hex_small,persistent_map_string_small,regex_is_match_small,reverse_complement_small,string_builder_small,string_contains_small,string_split_join_small,tapelang_small,word_count_small,zigzag_char_small"
      ;;
    fixture-algorithms)
      printf '%s\n' "binarytrees_small,dijkstra_heap_small,fib_i32_small,graph_bfs_small,knapsack_i32_small,levenshtein_small,mandelbrot_small,matrixmultiply_f64_small,monte_carlo_pi_small,nbody_small,pidigits_small,quicksort_file_small,sieve_count,sieve_full,sudoku_file_small,sum_u32_small,toposort_small,union_find_small"
      ;;
    fixture-concurrency)
      printf '%s\n' "await_batch_i64_small,binarytrees_small,channel_pipeline_i32_small,channel_roundtrip_i32_small,concurrent_queue_i32_small,future_fanout_i32_small,future_yield_i32_small,mutex_counter_i32_small"
      ;;
    fixture-numeric)
      printf '%s\n' "bigint_add_mul_small,bigint_ref_newton_small,biguint_add_mul_small,fib_i32_small,int128_accumulate_small,mandelbrot_small,matrixmultiply_f64_small,monte_carlo_pi_small,nbody_small,pidigits_small,random_lcg_i64_small,rational_series_small,sum_u32_small,uint128_accumulate_small"
      ;;
    fixture-external-small)
      printf '%s\n' "base64_roundtrip_small,binarytrees_small,i_before_e_small,json_means_small,k_nucleotide_small,mandelbrot_small,monte_carlo_pi_small,nbody_small,pidigits_small,quicksort_file_small,reverse_complement_small,sudoku_file_small,tapelang_small"
      ;;
    *)
      return 1
      ;;
  esac
}

bench_fixture_dir() {
  local dir="$1/fixtures/bench/$2"
  [[ -d "$dir" ]] || return 1
  printf '%s\n' "$dir"
}

bench_fixture_entry() {
  local dir
  dir="$(bench_fixture_dir "$1" "$2")" || return 1
  [[ -f "$dir/main.able" ]] || return 1
  printf '%s\n' "main.able"
}

bench_fixture_program_args() {
  local root="$1"
  local bench="$2"
  local dir
  dir="$(bench_fixture_dir "$root" "$bench")" || return 1
  case "$bench" in
    tapelang_small)
      printf '%s\n' "$dir/benchmark.tape"
      ;;
    i_before_e_small)
      printf '%s\n' "$dir/words.txt"
      ;;
    quicksort_file_small)
      printf '%s\n' "$dir/numbers.txt"
      ;;
    sudoku_file_small)
      printf '%s\n' "$dir/sudoku.txt"
      ;;
    k_nucleotide_small)
      printf '%s\n' "$dir/knucleotide-input.fasta"
      ;;
    reverse_complement_small)
      printf '%s\n' "$dir/reverse-complement-input.fasta"
      ;;
    *)
      return 0
      ;;
  esac
}
