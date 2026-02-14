package compiler

func defaultCompilerHeavyAuditFixtures() []string {
	return []string{
		"12_08_blocking_io_concurrency",
		"13_06_stdlib_package_resolution",
		"14_02_regex_core_match_streaming",
		"14_01_language_interfaces_index_apply_iterable",
		"06_01_compiler_struct_positional",
		"10_17_interface_overload_dispatch",
		"06_12_04_stdlib_numbers_bigint",
		"06_12_10_stdlib_collections_list_vector",
		"06_12_18_stdlib_collections_array_range",
		"06_12_19_stdlib_concurrency_channel_mutex_queue",
		"06_12_20_stdlib_math_core_numeric",
		"06_12_21_stdlib_fs_path",
		"06_12_26_stdlib_test_harness_reporters",
	}
}
