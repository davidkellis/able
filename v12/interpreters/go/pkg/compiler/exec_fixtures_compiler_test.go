package compiler

import (
	"bytes"
	"context"
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
	"able/interpreter-go/pkg/stdlibpath"
	"able/interpreter-go/pkg/typechecker"
)

const compilerFixtureEnv = "ABLE_COMPILER_EXEC_FIXTURES"
const compilerNoBootstrapFixtureEnv = "ABLE_COMPILER_NO_BOOTSTRAP_FIXTURES"

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
			t.Parallel()
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

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-fixture")

	comp := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: requireNoFallbacksForFixtureGates(t),
	})
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath)
	cmd.Env = applyFixtureEnv(os.Environ(), manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("compiled fixture timed out after 60s")
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
			t.Fatalf("stdout mismatch: expected %v, got %v\nstderr: %s", expectedStdout, actualStdout, stderr.String())
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
	buf.WriteString("\t\"able/interpreter-go/pkg/compiler/bridge\"\n")
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
	buf.WriteString("\trt, err := RegisterIn(interp, entryEnv)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tprintBoundaryMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcount := int64(0)\n")
	buf.WriteString("\t\texplicitCount := int64(0)\n")
	buf.WriteString("\t\tif rt != nil {\n")
	buf.WriteString("\t\t\tcount = __able_boundary_fallback_count_get()\n")
	buf.WriteString("\t\t\texplicitCount = __able_boundary_explicit_count_get()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_FALLBACK_CALLS=%d\\n\", count)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_EXPLICIT_CALLS=%d\\n\", explicitCount)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_FALLBACK_NAMES=%s\\n\", __able_boundary_fallback_snapshot())\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_EXPLICIT_NAMES=%s\\n\", __able_boundary_explicit_snapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif os.Getenv(\"ABLE_COMPILER_STRICT_DISPATCH_MARKER\") != \"\" {\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_STRICT=%v\\n\", __able_interface_dispatch_strict)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tprintGlobalLookupMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_GLOBAL_LOOKUP_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcalls := bridge.GlobalLookupFallbackStats()\n")
	buf.WriteString("\t\tenvCalls, registryCalls := bridge.GlobalLookupFallbackBucketStats()\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS=%d\\n\", calls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS=%d\\n\", envCalls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS=%d\\n\", registryCalls)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_NAMES=%s\\n\", bridge.GlobalLookupFallbackSnapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tprintInterfaceLookupMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_INTERFACE_LOOKUP_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcalls, ifaceCalls := bridge.MemberGetPreferMethodsStats()\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_CALLS=%d\\n\", calls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS=%d\\n\", ifaceCalls)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_NAMES=%s\\n\", bridge.MemberGetPreferMethodsSnapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tbridge.ResetMemberGetPreferMethodsCounters()\n")
	buf.WriteString("\tbridge.ResetGlobalLookupFallbackCounters()\n")
	buf.WriteString("\tif entryEnv == nil {\n\t\tentryEnv = interp.GlobalEnvironment()\n\t}\n")
	buf.WriteString("\tif err := RunRegisteredMain(rt, interp, entryEnv); err != nil {\n")
	buf.WriteString("\t\tprintBoundaryMarkers()\n")
	buf.WriteString("\t\tprintGlobalLookupMarkers()\n")
	buf.WriteString("\t\tprintInterfaceLookupMarkers()\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tprintBoundaryMarkers()\n")
	buf.WriteString("\tprintGlobalLookupMarkers()\n")
	buf.WriteString("\tprintInterfaceLookupMarkers()\n")
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
	buf.WriteString("\tcase runtime.IntegerValue:\n\t\treturn v.String()\n")
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
	fixtures := []string(nil)
	if raw := strings.TrimSpace(os.Getenv(compilerFixtureEnv)); raw != "" {
		if strings.EqualFold(raw, "all") {
			fixtures = collectExecFixtures(t, root)
		} else {
			parts := strings.FieldsFunc(raw, func(r rune) bool {
				return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
			})
			fixtures = make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed == "" {
					continue
				}
				fixtures = append(fixtures, trimmed)
			}
		}
	} else {
		fixtures = defaultCompilerExecFixtures()
	}
	return applyCompilerFixtureBatch(t, fixtures, compilerExecBatchIndexEnv, compilerExecBatchCountEnv)
}

func defaultCompilerExecFixtures() []string {
	return []string{
		"15_01_program_entry_hello_world",
		"15_02_entry_args_signature",
		"15_03_exit_status_return_value",
		"15_04_background_work_flush",
		"16_01_host_interop_inline_extern",
		"06_01_compiler_array_struct_literal",
		"06_01_compiler_assignment_pattern_errors",
		"06_01_compiler_assignment_pattern_positional_mismatch",
		"06_01_compiler_assignment_pattern_rest_mismatch",
		"06_01_compiler_assignment_patterns",
		"06_01_compiler_assignment_pattern_struct_mismatch",
		"06_01_compiler_assignment_pattern_typed_mismatch",
		"06_01_compiler_await_future",
		"06_01_compiler_bitwise_shift",
		"06_01_compiler_bound_method_value",
		"06_01_compiler_breakpoint",
		"06_01_compiler_compound_assignment",
		"06_01_compiler_compound_assignment_overflow",
		"06_01_compiler_division_by_zero",
		"06_01_compiler_division_ops",
		"06_01_compiler_divmod",
		"06_01_compiler_divmod_overflow",
		"06_01_compiler_dynamic_member_access",
		"06_01_compiler_dynamic_member_compound",
		"06_01_compiler_ensure_error_passthrough",
		"06_01_compiler_ensure_rethrow",
		"06_01_compiler_for_loop",
		"06_01_compiler_for_loop_pattern",
		"06_01_compiler_for_loop_pattern_guard",
		"06_01_compiler_for_loop_pattern_mismatch",
		"06_01_compiler_for_loop_struct_pattern",
		"06_01_compiler_for_loop_typed_pattern",
		"06_01_compiler_for_loop_typed_pattern_mismatch",
		"06_01_compiler_if_block_exprs",
		"06_01_compiler_index_assignment",
		"06_01_compiler_index_assignment_value",
		"06_01_compiler_index_statement",
		"06_01_compiler_integer_overflow",
		"06_01_compiler_integer_overflow_mul",
		"06_01_compiler_integer_overflow_sub",
		"06_01_compiler_iterator_literal",
		"06_01_compiler_lambda_closure",
		"06_01_compiler_loops",
		"06_01_compiler_map_literal",
		"06_01_compiler_map_literal_spread",
		"06_01_compiler_map_literal_typed",
		"06_01_compiler_match_patterns",
		"06_01_compiler_member_assignment",
		"06_01_compiler_method_call",
		"06_01_compiler_nullable_param",
		"06_01_compiler_nullable_return",
		"06_01_compiler_or_else",
		"06_01_compiler_or_else_error_union",
		"06_01_compiler_or_else_mixed",
		"06_01_compiler_or_else_struct_mix",
		"06_01_compiler_pipe",
		"06_01_compiler_placeholder_lambda",
		"06_01_compiler_pow_negative_exponent",
		"06_01_compiler_pow_overflow",
		"06_01_compiler_raise_error_interface",
		"06_01_compiler_raise_non_error",
		"06_01_compiler_rescue",
		"06_01_compiler_result_return",
		"06_01_compiler_safe_navigation",
		"06_01_compiler_shift_out_of_range",
		"06_01_compiler_spawn_await",
		"06_01_compiler_string_interpolation",
		"06_01_compiler_string_interpolation_display",
		"06_01_compiler_struct_param_bridge",
		"06_01_compiler_struct_positional",
		"06_01_compiler_struct_update",
		"06_01_compiler_type_qualified_method",
		"06_01_compiler_unary_overflow",
		"06_01_compiler_union_param",
		"06_01_compiler_union_return",
		"06_01_compiler_verbose_anonymous_fn",
		"07_07_overload_resolution_runtime",
		"07_10_iterator_reentrancy",
		"08_03_breakpoint_nonlocal_jump",
		"10_17_interface_overload_dispatch",
		"11_02_option_result_or_handlers",
		"12_02_future_fairness_cancellation",
		"12_08_blocking_io_concurrency",
		"13_04_import_alias_selective_dynimport",
		"13_05_dynimport_interface_dispatch",
		"13_06_stdlib_package_resolution",
		"14_01_language_interfaces_index_apply_iterable",
		"14_02_regex_core_match_streaming",
		"06_12_01_stdlib_string_helpers",
		"06_12_02_stdlib_array_helpers",
		"06_12_03_stdlib_numeric_ratio_divmod",
		"06_12_04_stdlib_numbers_bigint",
		"06_12_05_stdlib_numbers_biguint",
		"06_12_06_stdlib_numbers_int128",
		"06_12_07_stdlib_numbers_uint128",
		"06_12_08_stdlib_numbers_rational",
		"06_12_09_stdlib_numbers_primitives",
		"06_12_10_stdlib_collections_list_vector",
		"06_12_11_stdlib_collections_tree_map_set",
		"06_12_12_stdlib_collections_persistent_map_set",
		"06_12_13_stdlib_collections_persistent_sorted_queue",
		"06_12_14_stdlib_collections_linked_list_lazy_seq",
		"06_12_15_stdlib_collections_hash_map_set",
		"06_12_16_stdlib_collections_deque_queue",
		"06_12_17_stdlib_collections_bit_set_heap",
		"06_12_18_stdlib_collections_array_range",
		"06_12_19_stdlib_concurrency_channel_mutex_queue",
		"06_12_20_stdlib_math_core_numeric",
		"06_12_21_stdlib_fs_path",
		"06_12_22_stdlib_io_temp",
		"06_12_23_stdlib_os",
		"06_12_24_stdlib_process",
		"06_12_25_stdlib_term",
		"06_12_26_stdlib_test_harness_reporters",
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
		if len(roots) > 0 {
			return roots
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
	if installed := stdlibpath.ResolveInstalledSrc(); installed != "" {
		add(installed)
		return roots
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		if len(roots) > 0 {
			return roots
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
