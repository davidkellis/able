# Portable Feature-Interaction Matrix

- Families: `11`
- Pairwise interactions: `55`
- Baseline zero-coverage pairs: `0`
- Current zero-coverage pairs: `0`
- Improved pairs: `45`
- Baseline excludes: `policy_record_dispatch`
- Intentionally excluded families: `packages_modules_imports`

The baseline column removes only the named newly added applications from the
current manifest. Counts therefore measure exact interaction coverage gained
by the tranche without maintaining a second stale coverage manifest.

| Left family | Right family | Baseline | Current | Delta | Current applications |
| --- | --- | ---: | ---: | ---: | --- |
| `concurrency` | `control_flow` | 2 | 2 | +0 | `concurrent_event_routing`, `dependency_wave_validation` |
| `concurrency` | `expressions_arrays_text_files` | 2 | 2 | +0 | `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `functions_closures_callables` | 3 | 3 | +0 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `inherent_methods` | 5 | 5 | +0 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `interfaces_implementations_dispatch` | 4 | 4 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `lexical_blocks_bindings_patterns` | 1 | 1 | +0 | `concurrent_event_routing` |
| `concurrency` | `option_result_exceptions` | 3 | 3 | +0 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `program_entry` | 2 | 2 | +0 | `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `stdlib_protocols_regex` | 4 | 4 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `types_nominals_generics_unions` | 5 | 5 | +0 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `control_flow` | `expressions_arrays_text_files` | 4 | 5 | +1 | `concurrent_event_routing`, `distance_field`, `fasta_generation`, `policy_record_dispatch`, `rms_norm` |
| `control_flow` | `functions_closures_callables` | 2 | 3 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch` |
| `control_flow` | `inherent_methods` | 2 | 3 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch` |
| `control_flow` | `interfaces_implementations_dispatch` | 2 | 3 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch` |
| `control_flow` | `lexical_blocks_bindings_patterns` | 4 | 5 | +1 | `concurrent_event_routing`, `fib`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `control_flow` | `option_result_exceptions` | 2 | 3 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch` |
| `control_flow` | `program_entry` | 3 | 4 | +1 | `concurrent_event_routing`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `control_flow` | `stdlib_protocols_regex` | 3 | 4 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `fasta_generation`, `policy_record_dispatch` |
| `control_flow` | `types_nominals_generics_unions` | 2 | 3 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `functions_closures_callables` | 4 | 5 | +1 | `concurrent_event_routing`, `document_audit`, `lexical_rollup`, `policy_record_dispatch`, `word_frequency` |
| `expressions_arrays_text_files` | `inherent_methods` | 2 | 3 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `interfaces_implementations_dispatch` | 5 | 6 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `document_audit`, `lexical_rollup`, `policy_record_dispatch`, `word_frequency` |
| `expressions_arrays_text_files` | `lexical_blocks_bindings_patterns` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `option_result_exceptions` | 5 | 6 | +1 | `base64`, `concurrent_event_routing`, `config_validation_extraction`, `json`, `log_routing_redaction`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `program_entry` | 7 | 8 | +1 | `base64`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `i_before_e`, `log_routing_redaction`, `policy_record_dispatch`, `reverse_complement` |
| `expressions_arrays_text_files` | `stdlib_protocols_regex` | 11 | 12 | +1 | `array_slice_window`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `document_audit`, `fasta_generation`, `lexical_rollup`, `log_routing_redaction`, `policy_record_dispatch`, `regex_stream_audit`, `unicode_scalar_pipeline`, `word_frequency` |
| `expressions_arrays_text_files` | `types_nominals_generics_unions` | 2 | 3 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `functions_closures_callables` | `inherent_methods` | 3 | 4 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `functions_closures_callables` | `interfaces_implementations_dispatch` | 6 | 7 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `policy_record_dispatch`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `lexical_blocks_bindings_patterns` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `functions_closures_callables` | `option_result_exceptions` | 3 | 4 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `functions_closures_callables` | `program_entry` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `functions_closures_callables` | `stdlib_protocols_regex` | 7 | 8 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `policy_record_dispatch`, `regex_suffix_audit`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `types_nominals_generics_unions` | 3 | 4 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `interfaces_implementations_dispatch` | 6 | 7 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `lexical_blocks_bindings_patterns` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `inherent_methods` | `option_result_exceptions` | 4 | 5 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `program_entry` | 2 | 3 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `inherent_methods` | `stdlib_protocols_regex` | 5 | 6 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `types_nominals_generics_unions` | 10 | 11 | +1 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `fixed_width_128`, `option_result_config`, `policy_record_dispatch`, `rational_series`, `validated_job_pipeline`, `wide_integer_records` |
| `interfaces_implementations_dispatch` | `lexical_blocks_bindings_patterns` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `interfaces_implementations_dispatch` | `option_result_exceptions` | 4 | 5 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `interfaces_implementations_dispatch` | `program_entry` | 2 | 3 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `interfaces_implementations_dispatch` | `stdlib_protocols_regex` | 10 | 11 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `document_audit`, `inventory_reconciliation`, `k_nucleotide`, `lexical_rollup`, `policy_record_dispatch`, `validated_job_pipeline`, `word_frequency` |
| `interfaces_implementations_dispatch` | `types_nominals_generics_unions` | 8 | 9 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `lexical_blocks_bindings_patterns` | `option_result_exceptions` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `lexical_blocks_bindings_patterns` | `program_entry` | 3 | 4 | +1 | `concurrent_event_routing`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `lexical_blocks_bindings_patterns` | `stdlib_protocols_regex` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `lexical_blocks_bindings_patterns` | `types_nominals_generics_unions` | 1 | 2 | +1 | `concurrent_event_routing`, `policy_record_dispatch` |
| `option_result_exceptions` | `program_entry` | 4 | 5 | +1 | `base64`, `concurrent_event_routing`, `config_validation_extraction`, `log_routing_redaction`, `policy_record_dispatch` |
| `option_result_exceptions` | `stdlib_protocols_regex` | 5 | 6 | +1 | `concurrent_event_routing`, `config_validation_extraction`, `dependency_wave_validation`, `log_routing_redaction`, `policy_record_dispatch`, `validated_job_pipeline` |
| `option_result_exceptions` | `types_nominals_generics_unions` | 4 | 5 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `program_entry` | `stdlib_protocols_regex` | 4 | 5 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `log_routing_redaction`, `policy_record_dispatch` |
| `program_entry` | `types_nominals_generics_unions` | 2 | 3 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `stdlib_protocols_regex` | `types_nominals_generics_unions` | 7 | 8 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `policy_record_dispatch`, `validated_job_pipeline` |
