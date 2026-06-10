# Portable Feature-Interaction Matrix

- Families: `11`
- Pairwise interactions: `55`
- Baseline zero-coverage pairs: `0`
- Current zero-coverage pairs: `0`
- Baseline minimum pair depth: `3`
- Current minimum pair depth: `3`
- Baseline depth-one pairs: `0`
- Current depth-one pairs: `0`
- Improved pairs: `36`
- Baseline excludes: `manifest_normalization`
- Baseline removes memberships: `none`
- Intentionally excluded families: `packages_modules_imports`

The baseline column removes only the named applications or individual family
memberships from the current manifest. Counts therefore measure the exact
interaction coverage gained by the tranche without maintaining a second stale
coverage manifest.

| Left family | Right family | Baseline | Current | Delta | Current applications |
| --- | --- | ---: | ---: | ---: | --- |
| `concurrency` | `control_flow` | 4 | 4 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation` |
| `concurrency` | `expressions_arrays_text_files` | 3 | 3 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `functions_closures_callables` | 4 | 4 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `inherent_methods` | 6 | 6 | +0 | `binarytrees`, `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `interfaces_implementations_dispatch` | 4 | 4 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `lexical_blocks_bindings_patterns` | 4 | 4 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation` |
| `concurrency` | `option_result_exceptions` | 4 | 4 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `program_entry` | 3 | 3 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index` |
| `concurrency` | `stdlib_protocols_regex` | 5 | 5 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `concurrency` | `types_nominals_generics_unions` | 6 | 6 | +0 | `binarytrees`, `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `validated_job_pipeline` |
| `control_flow` | `expressions_arrays_text_files` | 7 | 8 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `distance_field`, `fasta_generation`, `manifest_normalization`, `policy_record_dispatch`, `rms_norm` |
| `control_flow` | `functions_closures_callables` | 4 | 5 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `control_flow` | `inherent_methods` | 5 | 5 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `policy_record_dispatch` |
| `control_flow` | `interfaces_implementations_dispatch` | 4 | 5 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `control_flow` | `lexical_blocks_bindings_patterns` | 8 | 9 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `fib`, `manifest_normalization`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `control_flow` | `option_result_exceptions` | 4 | 5 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `control_flow` | `program_entry` | 6 | 7 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `control_flow` | `stdlib_protocols_regex` | 6 | 7 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `fasta_generation`, `manifest_normalization`, `policy_record_dispatch` |
| `control_flow` | `types_nominals_generics_unions` | 5 | 6 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `functions_closures_callables` | 6 | 7 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `document_audit`, `lexical_rollup`, `manifest_normalization`, `policy_record_dispatch`, `word_frequency` |
| `expressions_arrays_text_files` | `inherent_methods` | 4 | 4 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `interfaces_implementations_dispatch` | 6 | 7 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `document_audit`, `lexical_rollup`, `manifest_normalization`, `policy_record_dispatch`, `word_frequency` |
| `expressions_arrays_text_files` | `lexical_blocks_bindings_patterns` | 4 | 5 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `option_result_exceptions` | 7 | 8 | +1 | `base64`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `json`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch` |
| `expressions_arrays_text_files` | `program_entry` | 9 | 10 | +1 | `base64`, `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `i_before_e`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch`, `reverse_complement` |
| `expressions_arrays_text_files` | `stdlib_protocols_regex` | 13 | 14 | +1 | `array_slice_window`, `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `document_audit`, `fasta_generation`, `lexical_rollup`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch`, `regex_stream_audit`, `unicode_scalar_pipeline`, `word_frequency` |
| `expressions_arrays_text_files` | `types_nominals_generics_unions` | 4 | 5 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch` |
| `functions_closures_callables` | `inherent_methods` | 5 | 5 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `functions_closures_callables` | `interfaces_implementations_dispatch` | 7 | 8 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `lexical_blocks_bindings_patterns` | 4 | 5 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `functions_closures_callables` | `option_result_exceptions` | 4 | 5 | +1 | `concurrent_event_routing`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline` |
| `functions_closures_callables` | `program_entry` | 3 | 4 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `manifest_normalization`, `policy_record_dispatch` |
| `functions_closures_callables` | `stdlib_protocols_regex` | 9 | 10 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `document_audit`, `lexical_rollup`, `manifest_normalization`, `policy_record_dispatch`, `regex_suffix_audit`, `validated_job_pipeline`, `word_frequency` |
| `functions_closures_callables` | `types_nominals_generics_unions` | 5 | 6 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `interfaces_implementations_dispatch` | 7 | 7 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `lexical_blocks_bindings_patterns` | 5 | 5 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `policy_record_dispatch` |
| `inherent_methods` | `option_result_exceptions` | 6 | 6 | +0 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `program_entry` | 4 | 4 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `policy_record_dispatch` |
| `inherent_methods` | `stdlib_protocols_regex` | 7 | 7 | +0 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `policy_record_dispatch`, `validated_job_pipeline` |
| `inherent_methods` | `types_nominals_generics_unions` | 12 | 12 | +0 | `binarytrees`, `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `fixed_width_128`, `option_result_config`, `policy_record_dispatch`, `rational_series`, `validated_job_pipeline`, `wide_integer_records` |
| `interfaces_implementations_dispatch` | `lexical_blocks_bindings_patterns` | 4 | 5 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `interfaces_implementations_dispatch` | `option_result_exceptions` | 6 | 7 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `interfaces_implementations_dispatch` | `program_entry` | 3 | 4 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch` |
| `interfaces_implementations_dispatch` | `stdlib_protocols_regex` | 11 | 12 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `document_audit`, `inventory_reconciliation`, `k_nucleotide`, `lexical_rollup`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline`, `word_frequency` |
| `interfaces_implementations_dispatch` | `types_nominals_generics_unions` | 9 | 10 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `manifest_normalization`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `lexical_blocks_bindings_patterns` | `option_result_exceptions` | 4 | 5 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `lexical_blocks_bindings_patterns` | `program_entry` | 6 | 7 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch`, `quicksort`, `sudoku_masks` |
| `lexical_blocks_bindings_patterns` | `stdlib_protocols_regex` | 5 | 6 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `lexical_blocks_bindings_patterns` | `types_nominals_generics_unions` | 5 | 6 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `policy_record_dispatch` |
| `option_result_exceptions` | `program_entry` | 6 | 7 | +1 | `base64`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch` |
| `option_result_exceptions` | `stdlib_protocols_regex` | 7 | 8 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `dependency_wave_validation`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline` |
| `option_result_exceptions` | `types_nominals_generics_unions` | 6 | 7 | +1 | `concurrent_event_routing`, `concurrent_text_index`, `dependency_wave_validation`, `manifest_normalization`, `option_result_config`, `policy_record_dispatch`, `validated_job_pipeline` |
| `program_entry` | `stdlib_protocols_regex` | 6 | 7 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `config_validation_extraction`, `log_routing_redaction`, `manifest_normalization`, `policy_record_dispatch` |
| `program_entry` | `types_nominals_generics_unions` | 4 | 5 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `manifest_normalization`, `policy_record_dispatch` |
| `stdlib_protocols_regex` | `types_nominals_generics_unions` | 9 | 10 | +1 | `concurrent_document_pipeline`, `concurrent_event_routing`, `concurrent_text_index`, `dependency_plan`, `dependency_wave_validation`, `inventory_reconciliation`, `k_nucleotide`, `manifest_normalization`, `policy_record_dispatch`, `validated_job_pipeline` |
