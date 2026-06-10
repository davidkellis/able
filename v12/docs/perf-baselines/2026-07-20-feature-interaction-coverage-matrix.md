# Portable Feature-Interaction Matrix

- Families: `11`
- Pairwise interactions: `55`
- Baseline zero-coverage pairs: `29`
- Current zero-coverage pairs: `0`
- Improved pairs: `55`
- Baseline excludes: `concurrent_event_routing, concurrent_text_index, dependency_wave_validation, validated_job_pipeline`
- Intentionally excluded families: `packages_modules_imports`

The baseline column removes only the named newly added applications from the
current manifest. Counts therefore measure exact interaction coverage gained
by the tranche without maintaining a second stale coverage manifest.

| Left family | Right family | Baseline | Current | Delta | Current applications |
| --- | --- | ---: | ---: | ---: | --- |
| `concurrency` | `control_flow` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `concurrency` | `expressions_arrays_text_files` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `functions_closures_callables` | 0 | 3 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `inherent_methods` | 1 | 5 | +4 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `interfaces_implementations_dispatch` | 0 | 4 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `lexical_blocks_bindings_patterns` | 0 | 1 | +1 | `concurrent_event_routing` |
| `concurrency` | `option_result_exceptions` | 0 | 3 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `program_entry` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `stdlib_protocols_regex` | 0 | 4 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `types_nominals_generics_unions` | 1 | 5 | +4 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `control_flow` | `expressions_arrays_text_files` | 3 | 4 | +1 | `concurrent_event_routing`, `distance_field`, `fasta_generation`, `rms_norm` |
| `control_flow` | `functions_closures_callables` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `control_flow` | `inherent_methods` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `control_flow` | `interfaces_implementations_dispatch` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `control_flow` | `lexical_blocks_bindings_patterns` | 3 | 4 | +1 | `concurrent_event_routing`, `fib`, `quicksort`, `sudoku_masks` |
| `control_flow` | `option_result_exceptions` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `control_flow` | `program_entry` | 2 | 3 | +1 | `concurrent_event_routing`, `quicksort`, `sudoku_masks` |
| `control_flow` | `stdlib_protocols_regex` | 1 | 3 | +2 | `concurrent_event_routing`, `dependency_wave_validation`, `fasta_generation` |
| `control_flow` | `types_nominals_generics_unions` | 0 | 2 | +2 | `concurrent_event_routing`, `dependency_wave_validation` |
| `expressions_arrays_text_files` | `functions_closures_callables` | 3 | 4 | +1 | `concurrent_event_routing`, `document_audit`, `lexical_rollup`, `word_frequency` |
| `expressions_arrays_text_files` | `inherent_methods` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `expressions_arrays_text_files` | `interfaces_implementations_dispatch` | 3 | 5 | +2 | `concurrent_event_routing`, `concurrent_text_index`, `document_audit`, `lexical_rollup`, `word_frequency` |
| `expressions_arrays_text_files` | `lexical_blocks_bindings_patterns` | 0 | 1 | +1 | `concurrent_event_routing` |
| `expressions_arrays_text_files` | `option_result_exceptions` | 4 | 5 | +1 | `base64`, `concurrent_event_routing`, `config_validation_extraction`, `json`, `log_routing_redaction` |
| `expressions_arrays_text_files` | `program_entry` | 5 | 7 | +2 | `base64`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `i_before_e`, `log_routing_redaction`, `reverse_complement` |
| `expressions_arrays_text_files` | `stdlib_protocols_regex` | 9 | 11 | +2 | `array_slice_window`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `document_audit`, `fasta_generation`, `lexical_rollup`, `log_routing_redaction`, `regex_stream_audit`, `unicode_scalar_pipeline`, `word_frequency` |
| `expressions_arrays_text_files` | `types_nominals_generics_unions` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `functions_closures_callables` | `inherent_methods` | 0 | 3 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `functions_closures_callables` | `interfaces_implementations_dispatch` | 3 | 6 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `lexical_blocks_bindings_patterns` | 0 | 1 | +1 | `concurrent_event_routing` |
| `functions_closures_callables` | `option_result_exceptions` | 0 | 3 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `functions_closures_callables` | `program_entry` | 0 | 1 | +1 | `concurrent_event_routing` |
| `functions_closures_callables` | `stdlib_protocols_regex` | 4 | 7 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `regex_suffix_audit`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `types_nominals_generics_unions` | 0 | 3 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `inherent_methods` | `interfaces_implementations_dispatch` | 2 | 6 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `option_result_config`, `validated_job_pipeline` |
| `inherent_methods` | `lexical_blocks_bindings_patterns` | 0 | 1 | +1 | `concurrent_event_routing` |
| `inherent_methods` | `option_result_exceptions` | 1 | 4 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `validated_job_pipeline` |
| `inherent_methods` | `program_entry` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `inherent_methods` | `stdlib_protocols_regex` | 1 | 5 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `validated_job_pipeline` |
| `inherent_methods` | `types_nominals_generics_unions` | 6 | 10 | +4 | `binarytrees`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `fixed_width_128`, `option_result_config`, `rational_series`, `validated_job_pipeline`, `wide_integer_records` |
| `interfaces_implementations_dispatch` | `lexical_blocks_bindings_patterns` | 0 | 1 | +1 | `concurrent_event_routing` |
| `interfaces_implementations_dispatch` | `option_result_exceptions` | 1 | 4 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `validated_job_pipeline` |
| `interfaces_implementations_dispatch` | `program_entry` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `interfaces_implementations_dispatch` | `stdlib_protocols_regex` | 6 | 10 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `document_audit`, `inventory_reconciliation`, `k_nucleotide`, `lexical_rollup`, `validated_job_pipeline`, `word_frequency` |
| `interfaces_implementations_dispatch` | `types_nominals_generics_unions` | 4 | 8 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `option_result_config`, `validated_job_pipeline` |
| `lexical_blocks_bindings_patterns` | `option_result_exceptions` | 0 | 1 | +1 | `concurrent_event_routing` |
| `lexical_blocks_bindings_patterns` | `program_entry` | 2 | 3 | +1 | `concurrent_event_routing`, `quicksort`, `sudoku_masks` |
| `lexical_blocks_bindings_patterns` | `stdlib_protocols_regex` | 0 | 1 | +1 | `concurrent_event_routing` |
| `lexical_blocks_bindings_patterns` | `types_nominals_generics_unions` | 0 | 1 | +1 | `concurrent_event_routing` |
| `option_result_exceptions` | `program_entry` | 3 | 4 | +1 | `base64`, `concurrent_event_routing`, `config_validation_extraction`, `log_routing_redaction` |
| `option_result_exceptions` | `stdlib_protocols_regex` | 2 | 5 | +3 | `concurrent_event_routing`, `config_validation_extraction`, `dependency_wave_validation`, `log_routing_redaction`, `validated_job_pipeline` |
| `option_result_exceptions` | `types_nominals_generics_unions` | 1 | 4 | +3 | `concurrent_event_routing`, `dependency_wave_validation`, `option_result_config`, `validated_job_pipeline` |
| `program_entry` | `stdlib_protocols_regex` | 2 | 4 | +2 | `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `log_routing_redaction` |
| `program_entry` | `types_nominals_generics_unions` | 0 | 2 | +2 | `concurrent_event_routing`, `concurrent_text_index` |
| `stdlib_protocols_regex` | `types_nominals_generics_unions` | 3 | 7 | +4 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `validated_job_pipeline` |
