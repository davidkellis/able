package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/typechecker"
)

const compilerFixtureEnv = "ABLE_COMPILER_EXEC_FIXTURES"

func TestCompilerExecFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler fixture parity in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveCompilerFixtures(t, root)
	if len(fixtures) == 0 {
		t.Skip("no compiler fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		dir := filepath.Join(root, filepath.FromSlash(rel))
		t.Run(filepath.ToSlash(rel), func(t *testing.T) {
			runCompilerExecFixture(t, dir, rel)
		})
	}
}

func runCompilerExecFixture(t *testing.T, dir string, rel string) {
	t.Helper()
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	expectedTypecheck := manifest.Expect.TypecheckDiagnostics
	if expectedTypecheck != nil {
		check, err := interpreter.TypecheckProgram(program)
		if err != nil {
			t.Fatalf("typecheck program: %v", err)
		}
		formatted := formatModuleDiagnostics(check.Diagnostics)
		if len(expectedTypecheck) == 0 {
			if len(formatted) != 0 {
				t.Fatalf("typecheck diagnostics mismatch: expected none, got %v", formatted)
			}
		} else {
			expectedKeys := diagnosticKeys(expectedTypecheck)
			actualKeys := diagnosticKeys(formatted)
			if len(expectedKeys) != len(actualKeys) {
				t.Fatalf("typecheck diagnostics mismatch: expected %v, got %v", expectedTypecheck, formatted)
			}
			for i := range expectedKeys {
				if expectedKeys[i] != actualKeys[i] {
					t.Fatalf("typecheck diagnostics mismatch: expected %v, got %v", expectedTypecheck, formatted)
				}
			}
		}
	}
	if expectedTypecheck != nil && len(expectedTypecheck) > 0 {
		return
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir tmp: %v", err)
	}
	workDir, err := os.MkdirTemp(tmpRoot, "ablec-fixture-")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(workDir) })

	comp := New(Options{PackageName: "main"})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	harness := compilerHarnessSource(entryPath, searchPaths, manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-fixture")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	cmd := exec.Command(binPath)
	cmd.Env = applyFixtureEnv(os.Environ(), manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	start := time.Now()
	runErr := cmd.Run()
	if time.Since(start) > time.Minute {
		t.Fatalf("fixture runtime exceeded 1 minute")
	}

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run error: %v", runErr)
		}
	}

	actualStdout := splitLines(stdout.String())
	actualStderr := splitLines(stderr.String())
	expected := manifest.Expect
	if expected.Stdout != nil {
		expectedStdout := expandFixtureLines(expected.Stdout)
		if !reflect.DeepEqual(actualStdout, expectedStdout) {
			t.Fatalf("stdout mismatch: expected %v, got %v", expectedStdout, actualStdout)
		}
	}
	if expected.Stderr != nil {
		expectedStderr := expandFixtureLines(expected.Stderr)
		if !reflect.DeepEqual(actualStderr, expectedStderr) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expectedStderr, actualStderr)
		}
	}
	if expected.Exit != nil {
		if exitCode != *expected.Exit {
			t.Fatalf("exit code mismatch: expected %d, got %d", *expected.Exit, exitCode)
		}
	} else if exitCode != 0 {
		t.Fatalf("exit code mismatch: expected 0, got %d", exitCode)
	}
}

func compilerHarnessSource(entryPath string, searchPaths []driver.SearchPath, executorName string) string {
	var buf strings.Builder
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"os\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/driver\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/interpreter\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/runtime\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func main() {\n")
	buf.WriteString("\tsearchPaths := []driver.SearchPath{\n")
	for _, sp := range searchPaths {
		buf.WriteString(fmt.Sprintf("\t\t{Path: %q, Kind: %s},\n", sp.Path, searchPathKindLiteral(sp.Kind)))
	}
	buf.WriteString("\t}\n")
	buf.WriteString(fmt.Sprintf("\tentry := %q\n", entryPath))
	buf.WriteString("\tloader, err := driver.NewLoader(searchPaths)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tdefer loader.Close()\n")
	buf.WriteString("\tprogram, err := loader.Load(entry)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString(fmt.Sprintf("\texecutor := selectFixtureExecutor(%q)\n", executorName))
	buf.WriteString("\tinterp := interpreter.NewWithExecutor(executor)\n")
	buf.WriteString("\tinterp.SetArgs(os.Args[1:])\n")
	buf.WriteString("\tregisterPrint(interp)\n")
	buf.WriteString("\tmode := resolveFixtureTypecheckMode()\n")
	buf.WriteString("\t_, entryEnv, _, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{SkipTypecheck: mode == typecheckModeOff, AllowDiagnostics: mode != typecheckModeOff})\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tif _, err := RegisterIn(interp, entryEnv); err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tif entryEnv == nil {\n\t\tentryEnv = interp.GlobalEnvironment()\n\t}\n")
	buf.WriteString("\tmainValue, err := entryEnv.Get(\"main\")\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tif _, err := interp.CallFunction(mainValue, nil); err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func selectFixtureExecutor(name string) interpreter.Executor {\n")
	buf.WriteString("\tswitch strings.ToLower(strings.TrimSpace(name)) {\n")
	buf.WriteString("\tcase \"\", \"serial\":\n")
	buf.WriteString("\t\treturn interpreter.NewSerialExecutor(nil)\n")
	buf.WriteString("\tcase \"goroutine\":\n")
	buf.WriteString("\t\treturn interpreter.NewGoroutineExecutor(nil)\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"unknown fixture executor %q\\n\", name)\n")
	buf.WriteString("\t\tos.Exit(1)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn nil\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func registerPrint(interp *interpreter.Interpreter) {\n")
	buf.WriteString("\tprintFn := runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\tName:  \"print\",\n\t\tArity: 1,\n")
	buf.WriteString("\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\tvar parts []string\n")
	buf.WriteString("\t\t\tfor _, arg := range args {\n\t\t\t\tparts = append(parts, formatRuntimeValue(arg))\n\t\t\t}\n")
	buf.WriteString("\t\t\tfmt.Fprintln(os.Stdout, strings.Join(parts, \" \"))\n")
	buf.WriteString("\t\t\treturn runtime.NilValue{}, nil\n\t\t},\n\t}\n")
	buf.WriteString("\tinterp.GlobalEnvironment().Define(\"print\", printFn)\n}\n\n")
	buf.WriteString("func formatRuntimeValue(val runtime.Value) string {\n")
	buf.WriteString("\tswitch v := val.(type) {\n")
	buf.WriteString("\tcase runtime.StringValue:\n\t\treturn v.Val\n")
	buf.WriteString("\tcase runtime.BoolValue:\n\t\tif v.Val { return \"true\" }; return \"false\"\n")
	buf.WriteString("\tcase runtime.VoidValue:\n\t\treturn \"void\"\n")
	buf.WriteString("\tcase runtime.IntegerValue:\n\t\treturn v.Val.String()\n")
	buf.WriteString("\tcase runtime.FloatValue:\n\t\treturn fmt.Sprintf(\"%g\", v.Val)\n")
	buf.WriteString("\tdefault:\n\t\treturn fmt.Sprintf(\"[%s]\", v.Kind())\n\t}\n}\n")
	buf.WriteString("type fixtureTypecheckMode int\n\n")
	buf.WriteString("const (\n")
	buf.WriteString("\ttypecheckModeOff fixtureTypecheckMode = iota\n")
	buf.WriteString("\ttypecheckModeWarn\n")
	buf.WriteString("\ttypecheckModeStrict\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func resolveFixtureTypecheckMode() fixtureTypecheckMode {\n")
	buf.WriteString("\tif _, ok := os.LookupEnv(\"ABLE_TYPECHECK_FIXTURES\"); !ok {\n")
	buf.WriteString("\t\treturn typecheckModeStrict\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tmode := strings.TrimSpace(strings.ToLower(os.Getenv(\"ABLE_TYPECHECK_FIXTURES\")))\n")
	buf.WriteString("\tswitch mode {\n")
	buf.WriteString("\tcase \"\", \"0\", \"off\", \"false\":\n")
	buf.WriteString("\t\treturn typecheckModeOff\n")
	buf.WriteString("\tcase \"strict\", \"fail\", \"error\", \"1\", \"true\":\n")
	buf.WriteString("\t\treturn typecheckModeStrict\n")
	buf.WriteString("\tcase \"warn\", \"warning\":\n")
	buf.WriteString("\t\treturn typecheckModeWarn\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn typecheckModeWarn\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String()
}

func searchPathKindLiteral(kind driver.RootKind) string {
	switch kind {
	case driver.RootStdlib:
		return "driver.RootStdlib"
	default:
		return "driver.RootUser"
	}
}

func applyFixtureEnv(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key := entry
		if idx := strings.Index(entry, "="); idx >= 0 {
			key = entry[:idx]
		}
		if _, ok := overrides[key]; ok {
			seen[key] = struct{}{}
			out = append(out, fmt.Sprintf("%s=%s", key, overrides[key]))
			continue
		}
		out = append(out, entry)
	}
	for key, value := range overrides {
		if _, ok := seen[key]; ok {
			continue
		}
		out = append(out, fmt.Sprintf("%s=%s", key, value))
	}
	return out
}

func resolveCompilerFixtures(t *testing.T, root string) []string {
	if raw := strings.TrimSpace(os.Getenv(compilerFixtureEnv)); raw != "" {
		if strings.EqualFold(raw, "all") {
			return collectExecFixtures(t, root)
		}
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
		})
		fixtures := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			fixtures = append(fixtures, trimmed)
		}
		return fixtures
	}
	return []string{
		"15_01_program_entry_hello_world",
		"15_02_entry_args_signature",
		"15_03_exit_status_return_value",
		"15_04_background_work_flush",
		"16_01_host_interop_inline_extern",
		"02_lexical_comments_identifiers",
		"03_blocks_expr_separation",
		"04_01_type_inference_constraints",
		"06_05_control_flow_expr_value",
		"06_02_block_expression_value_scope",
		"06_01_compiler_if_block_exprs",
		"06_01_compiler_index_statement",
		"06_01_compiler_index_assignment",
		"06_01_compiler_index_assignment_value",
		"06_01_compiler_array_struct_literal",
		"06_01_compiler_struct_update",
		"06_01_compiler_struct_positional",
		"06_01_compiler_map_literal",
		"06_01_compiler_map_literal_spread",
		"06_01_compiler_map_literal_typed",
		"06_01_compiler_member_assignment",
		"06_01_compiler_division_by_zero",
		"06_01_compiler_division_ops",
		"06_01_compiler_bitwise_shift",
		"06_01_compiler_divmod",
		"06_01_compiler_shift_out_of_range",
		"06_01_compiler_compound_assignment",
		"06_01_compiler_dynamic_member_compound",
		"06_01_compiler_dynamic_member_access",
		"06_01_compiler_bound_method_value",
		"06_01_compiler_safe_navigation",
		"06_01_compiler_breakpoint",
		"06_01_compiler_loops",
		"06_01_compiler_for_loop",
		"06_01_compiler_for_loop_pattern",
		"06_01_compiler_for_loop_pattern_mismatch",
		"06_01_compiler_for_loop_struct_pattern",
		"06_01_compiler_for_loop_pattern_guard",
		"06_01_compiler_for_loop_typed_pattern",
		"06_01_compiler_for_loop_typed_pattern_mismatch",
		"06_01_compiler_nullable_return",
		"06_01_compiler_union_return",
		"06_01_compiler_result_return",
		"06_01_compiler_union_param",
		"06_01_compiler_nullable_param",
		"06_01_compiler_struct_param_bridge",
		"06_01_compiler_match_patterns",
		"06_01_compiler_assignment_patterns",
		"06_01_compiler_assignment_pattern_errors",
		"06_01_compiler_assignment_pattern_typed_mismatch",
		"06_01_compiler_assignment_pattern_rest_mismatch",
		"06_01_compiler_assignment_pattern_struct_mismatch",
		"06_01_compiler_assignment_pattern_positional_mismatch",
		"06_01_compiler_iterator_literal",
		"06_01_compiler_spawn_await",
		"06_01_compiler_await_future",
		"06_01_compiler_placeholder_lambda",
		"06_01_compiler_pipe",
		"06_01_compiler_rescue",
		"06_01_compiler_ensure_rethrow",
		"06_01_compiler_ensure_error_passthrough",
		"06_01_compiler_raise_error_interface",
		"06_01_compiler_raise_non_error",
		"06_01_compiler_or_else",
		"06_01_compiler_or_else_mixed",
		"06_01_compiler_or_else_struct_mix",
		"06_01_compiler_or_else_error_union",
		"06_01_compiler_string_interpolation",
		"06_01_compiler_string_interpolation_display",
		"06_01_compiler_method_call",
		"06_01_compiler_type_qualified_method",
		"06_01_compiler_lambda_closure",
		"06_01_compiler_verbose_anonymous_fn",
		"07_01_function_definition_generics_inference",
		"07_02_lambdas_closures_capture",
		"07_02_01_verbose_anonymous_fn",
		"07_02_bytecode_lambda_calls",
		"07_03_explicit_return_flow",
		"07_04_apply_callable_interface",
		"07_04_trailing_lambda_method_syntax",
		"07_05_partial_application",
		"07_06_shorthand_member_placeholder_lambdas",
		"07_07_bytecode_implicit_iterator",
		"07_08_bytecode_placeholder_lambda",
		"07_09_bytecode_iterator_yield",
		"07_07_overload_resolution_runtime",
		"07_08_return_context_generic_call_inference",
		"06_01_literals_array_map_inference",
		"06_01_literals_numeric_contextual",
		"06_01_literals_numeric_contextual_diag",
		"06_01_literals_string_char_escape",
		"06_01_bytecode_map_spread",
		"06_07_generator_yield_iterator_end",
		"06_07_iterator_pipeline",
		"06_08_array_ops_mutability",
		"06_02_bytecode_unary_range_cast",
		"06_09_lexical_trailing_commas_line_join",
		"06_10_dynamic_metaprogramming_package_object",
		"06_03_safe_navigation_nil_short_circuit",
		"06_04_function_call_eval_order_trailing_lambda",
		"06_06_string_interpolation",
		"06_03_cast_semantics",
		"06_03_cast_error_payload_recovery",
		"06_03_operator_precedence_associativity",
		"06_03_operator_overloading_interfaces",
		"14_01_operator_interfaces_arithmetic_comparison",
		"14_01_language_interfaces_index_apply_iterable",
		"14_02_hash_eq_primitives",
		"14_02_hash_eq_float",
		"14_02_hash_eq_custom",
		"14_02_regex_core_match_streaming",
		"10_01_interface_defaults_composites",
		"10_02_impl_specificity_named_overrides",
		"10_02_impl_where_clause",
		"10_03_interface_type_dynamic_dispatch",
		"10_04_interface_dispatch_defaults_generics",
		"10_05_interface_named_impl_defaults",
		"10_06_interface_generic_param_dispatch",
		"10_07_interface_default_chain",
		"10_08_interface_default_override",
		"10_09_interface_named_impl_inherent",
		"10_10_interface_inheritance_defaults",
		"10_11_interface_generic_args_dispatch",
		"10_12_interface_union_target_dispatch",
		"10_13_interface_param_generic_args",
		"10_14_interface_return_generic_args",
		"10_15_interface_default_generic_method",
		"10_16_interface_value_storage",
		"13_01_package_structure_modules",
		"13_02_packages_visibility_diag",
		"13_03_package_config_prelude",
		"13_04_import_alias_selective_dynimport",
		"13_05_dynimport_interface_dispatch",
		"13_06_stdlib_package_resolution",
		"13_07_search_path_env_override",
		"12_01_bytecode_spawn_basic",
		"12_01_bytecode_await_default",
		"12_02_async_spawn_combo",
		"12_02_future_fairness_cancellation",
		"12_03_spawn_future_status_error",
		"12_04_future_handle_value_view",
		"12_05_concurrency_channel_ping_pong",
		"12_05_mutex_lock_unlock",
		"12_06_await_fairness_cancellation",
		"12_07_channel_mutex_error_types",
		"12_08_blocking_io_concurrency",
		"06_11_truthiness_boolean_context",
		"06_12_01_stdlib_string_helpers",
		"06_12_02_stdlib_array_helpers",
		"06_12_03_stdlib_numeric_ratio_divmod",
		"08_01_if_truthiness_value",
		"08_01_control_flow_fizzbuzz",
		"08_01_bytecode_if_indexing",
		"08_01_bytecode_match_basic",
		"08_01_bytecode_match_subject",
		"08_01_match_guards_exhaustiveness",
		"08_01_union_match_basic",
		"08_02_bytecode_loop_basics",
		"08_02_loop_expression_break_value",
		"08_02_numeric_sum_loop",
		"08_02_range_inclusive_exclusive",
		"08_02_while_continue_break",
		"08_03_breakpoint_nonlocal_jump",
		"09_00_methods_generics_imports_combo",
		"09_00_bytecode_member_calls",
		"09_02_methods_instance_vs_static",
		"09_04_methods_ufcs_basics",
		"09_05_method_set_generics_where",
		"04_02_primitives_truthiness_numeric",
		"04_02_primitives_truthiness_numeric_diag",
		"04_03_type_expression_syntax",
		"04_03_type_expression_arity_diag",
		"04_03_type_expression_associativity_diag",
		"04_04_reserved_underscore_types",
		"04_05_01_struct_singleton_usage",
		"04_05_02_struct_named_update_mutation",
		"04_05_02_struct_named_update_mutation_diag",
		"04_05_03_struct_positional_named_tuple",
		"04_05_04_struct_literal_generic_inference",
		"05_00_mutability_declaration_vs_assignment",
		"05_02_array_nested_patterns",
		"05_02_identifier_wildcard_typed_patterns",
		"05_02_struct_pattern_rename_typed",
		"05_03_assignment_evaluation_order",
		"05_03_bytecode_assignment_patterns",
		"04_06_01_union_payload_patterns",
		"04_06_02_nullable_truthiness",
		"04_06_03_union_construction_result_option",
		"04_06_04_union_guarded_match_exhaustive",
		"04_06_04_union_guarded_match_exhaustive_diag",
		"04_07_02_alias_generic_substitution",
		"04_07_03_alias_scope_visibility_imports",
		"04_07_04_alias_methods_impls_interaction",
		"04_07_05_alias_recursion_termination",
		"04_07_06_alias_reexport_methods_impls",
		"04_07_types_alias_union_generic_combo",
		"11_00_errors_match_loop_combo",
		"11_01_return_statement_type_enforcement",
		"11_01_return_statement_typecheck_diag",
		"11_02_bytecode_or_else_basic",
		"11_02_option_result_or_handlers",
		"11_02_option_result_propagation",
		"11_03_raise_exit_unhandled",
		"11_03_bytecode_ensure_basic",
		"11_03_bytecode_rescue_basic",
		"11_03_rescue_ensure",
		"11_03_rescue_rethrow_standard_errors",
	}
}

func collectExecFixtures(t *testing.T, root string) []string {
	t.Helper()
	if root == "" {
		return nil
	}
	var dirs []string
	var walk func(string)
	walk = func(current string) {
		entries, err := os.ReadDir(current)
		if err != nil {
			return
		}
		hasManifest := false
		for _, entry := range entries {
			if entry.Type().IsRegular() && entry.Name() == "manifest.json" {
				hasManifest = true
				break
			}
		}
		if hasManifest {
			rel, err := filepath.Rel(root, current)
			if err == nil {
				dirs = append(dirs, filepath.ToSlash(rel))
			}
		}
		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(current, entry.Name()))
			}
		}
	}
	walk(root)
	sort.Strings(dirs)
	return dirs
}

func splitLines(raw string) []string {
	trimmed := strings.TrimRight(raw, "\n")
	if strings.TrimSpace(trimmed) == "" {
		return []string{}
	}
	return strings.Split(trimmed, "\n")
}

func expandFixtureLines(lines []string) []string {
	if len(lines) == 0 {
		return []string{}
	}
	var out []string
	for _, raw := range lines {
		trimmed := strings.TrimRight(raw, "\n")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		out = append(out, strings.Split(trimmed, "\n")...)
	}
	return out
}

func shouldSkipTarget(skip []string, target string) bool {
	if len(skip) == 0 {
		return false
	}
	target = strings.ToLower(target)
	for _, entry := range skip {
		if strings.ToLower(strings.TrimSpace(entry)) == target {
			return true
		}
	}
	return false
}

// Copy of exec fixture search path resolution helpers (kept local to avoid import cycles).

func buildExecSearchPaths(entryPath string, fixtureDir string, manifest interpreter.FixtureManifest) ([]driver.SearchPath, error) {
	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, err
	}
	entryDir := filepath.Dir(entryAbs)

	manifestRoot := findFixtureManifestRoot(entryDir)
	ablePathEnv := resolveFixtureEnv("ABLE_PATH", manifest.Env, os.Getenv("ABLE_PATH"))
	ableModulePathsEnv := resolveFixtureEnv("ABLE_MODULE_PATHS", manifest.Env, os.Getenv("ABLE_MODULE_PATHS"))

	var paths []driver.SearchPath
	seen := map[string]struct{}{}
	add := func(candidate string, kind driver.RootKind) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		paths = append(paths, driver.SearchPath{Path: abs, Kind: kind})
	}

	for _, extra := range []string{manifestRoot, entryDir} {
		add(extra, driver.RootUser)
	}
	if cwd, err := os.Getwd(); err == nil {
		add(cwd, driver.RootUser)
	}
	for _, entry := range resolveFixturePathList(ablePathEnv, fixtureDir) {
		add(entry, driver.RootUser)
	}
	for _, entry := range resolveFixturePathList(ableModulePathsEnv, fixtureDir) {
		add(entry, driver.RootUser)
	}
	for _, entry := range findKernelRoots(entryDir) {
		add(entry, driver.RootStdlib)
	}
	for _, entry := range findStdlibRoots(entryDir) {
		add(entry, driver.RootStdlib)
	}
	if cwd, err := os.Getwd(); err == nil {
		for _, entry := range findKernelRoots(cwd) {
			add(entry, driver.RootStdlib)
		}
		for _, entry := range findStdlibRoots(cwd) {
			add(entry, driver.RootStdlib)
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, entry := range findKernelRoots(exeDir) {
			add(entry, driver.RootStdlib)
		}
		for _, entry := range findStdlibRoots(exeDir) {
			add(entry, driver.RootStdlib)
		}
	}
	return paths, nil
}

func resolveFixtureEnv(key string, env map[string]string, fallback string) string {
	if env == nil {
		return fallback
	}
	if value, ok := env[key]; ok {
		return value
	}
	return fallback
}

func resolveFixturePathList(raw string, baseDir string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	rawParts := strings.Split(raw, string(os.PathListSeparator))
	parts := make([]string, 0, len(rawParts))
	for _, entry := range rawParts {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		if !filepath.IsAbs(trimmed) {
			trimmed = filepath.Join(baseDir, filepath.FromSlash(trimmed))
		}
		parts = append(parts, trimmed)
	}
	return parts
}

func findFixtureManifestRoot(start string) string {
	dir := start
	for {
		candidate := filepath.Join(dir, "package.yml")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func findKernelRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "kernel", "src"),
			filepath.Join(dir, "v12", "kernel", "src"),
			filepath.Join(dir, "ablekernel", "src"),
			filepath.Join(dir, "able_kernel", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

func findStdlibRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "v12", "stdlib", "src"),
			filepath.Join(dir, "stdlib", "v12", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

// Typecheck diagnostic formatting helpers (copied to keep expectations stable).

type diagKey struct {
	path    string
	line    int
	message string
}

func diagnosticKeys(entries []string) []diagKey {
	if len(entries) == 0 {
		return nil
	}
	keys := make([]diagKey, 0, len(entries))
	for _, entry := range entries {
		key := parseDiagnosticKey(entry)
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		if keys[i].line != keys[j].line {
			return keys[i].line < keys[j].line
		}
		return keys[i].message < keys[j].message
	})
	return keys
}

func parseDiagnosticKey(entry string) diagKey {
	trimmed := entry
	severityPrefix := ""
	if strings.HasPrefix(trimmed, "warning: ") {
		severityPrefix = "warning: "
		trimmed = strings.TrimPrefix(trimmed, "warning: ")
	}
	trimmed = strings.TrimPrefix(trimmed, "typechecker: ")
	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		return diagKey{message: severityPrefix + trimmed}
	}
	location := parts[0]
	message := parts[1]
	if !strings.HasPrefix(message, "typechecker:") {
		message = "typechecker: " + message
	}
	if severityPrefix != "" {
		message = severityPrefix + message
	}
	path := location
	line := 0
	if colon := strings.LastIndex(location, ":"); colon > 0 {
		path = location[:colon]
		suffix := location[colon+1:]
		if inner := strings.Index(suffix, ":"); inner >= 0 {
			suffix = suffix[:inner]
		}
		if parsed, err := strconv.Atoi(suffix); err == nil {
			line = parsed
		}
	}
	return diagKey{path: path, line: line, message: message}
}

func formatModuleDiagnostics(diags []interpreter.ModuleDiagnostic) []string {
	if len(diags) == 0 {
		return nil
	}
	msgs := make([]string, len(diags))
	for i, diag := range diags {
		msgs[i] = formatModuleDiagnostic(diag)
	}
	return msgs
}

func formatModuleDiagnostic(diag interpreter.ModuleDiagnostic) string {
	location := formatSourceHint(diag.Source)
	prefix := "typechecker: "
	if diag.Diagnostic.Severity == typechecker.SeverityWarning {
		prefix = "warning: typechecker: "
	}
	if location != "" {
		return fmt.Sprintf("%s%s %s", prefix, location, diag.Diagnostic.Message)
	}
	return fmt.Sprintf("%s%s", prefix, diag.Diagnostic.Message)
}

func formatSourceHint(hint typechecker.SourceHint) string {
	path := normalizeSourcePath(strings.TrimSpace(hint.Path))
	line := hint.Line
	column := hint.Column
	switch {
	case path != "" && line > 0 && column > 0:
		return fmt.Sprintf("%s:%d:%d", path, line, column)
	case path != "" && line > 0:
		return fmt.Sprintf("%s:%d", path, line)
	case path != "":
		return path
	case line > 0 && column > 0:
		return fmt.Sprintf("line %d, column %d", line, column)
	case line > 0:
		return fmt.Sprintf("line %d", line)
	default:
		return ""
	}
}

func normalizeSourcePath(raw string) string {
	if raw == "" {
		return ""
	}
	path := raw
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	root := repositoryRoot()
	anchors := []string{}
	if root != "" {
		anchors = append(anchors, root)
		anchors = append(anchors, filepath.Join(root, "v12"))
	}
	for _, anchor := range anchors {
		if anchor == "" {
			continue
		}
		rel, err := filepath.Rel(anchor, path)
		if err != nil {
			continue
		}
		rel = filepath.Clean(rel)
		if rel == "." || rel == "" {
			path = rel
			break
		}
		if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
			continue
		}
		path = rel
		break
	}
	return filepath.ToSlash(path)
}

var (
	repoRootOnce sync.Once
	repoRootPath string
	repoRootErr  error
)

func repositoryRoot() string {
	repoRootOnce.Do(func() {
		start := ""
		if _, file, _, ok := runtime.Caller(0); ok {
			start = filepath.Dir(file)
		} else if wd, err := os.Getwd(); err == nil {
			start = wd
		}
		dir := start
		for i := 0; i < 10 && dir != "" && dir != string(filepath.Separator); i++ {
			if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
				repoRootPath = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if repoRootPath == "" {
			repoRootErr = fmt.Errorf("repository root not found from %s", start)
		}
	})
	if repoRootErr != nil {
		return ""
	}
	return repoRootPath
}
