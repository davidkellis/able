package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

type closedBenchmarkSourceAuditCase struct {
	name      string
	entryPath string
	functions []string
}

// The three tests keep this audit under the project’s one-minute test bound
// even with a cold compiler cache. Together they cover the source bodies that
// dominate the established static compiled benchmark families. Each excludes
// explicit file, argument, output, and scheduler boundaries.
func TestCompilerClosedBenchmarkNumericComputeKernelsStayStatic(t *testing.T) {
	auditClosedBenchmarkComputeKernels(t, []closedBenchmarkSourceAuditCase{
		{"fib", "fib.able", []string{"fib"}},
		{"matrixmultiply", "matrixmultiply.able", []string{"build_matrix", "matmul"}},
		{"pidigits", "pidigits/pidigits.able", []string{"parse_positive_i32", "next_term", "eliminate_digit"}},
	})
}

func TestCompilerClosedBenchmarkStructuralComputeKernelsStayStatic(t *testing.T) {
	auditClosedBenchmarkComputeKernels(t, []closedBenchmarkSourceAuditCase{
		{"binarytrees", "binarytrees.able", []string{"make_tree", "check_tree"}},
		{"quicksort", "quicksort/quicksort.able", []string{"parse_numbers", "swap", "quicksort"}},
	})
}

func TestCompilerClosedBenchmarkSearchComputeKernelsStayStatic(t *testing.T) {
	auditClosedBenchmarkComputeKernels(t, []closedBenchmarkSourceAuditCase{
		{"sudoku", "sudoku/sudoku.able", []string{"parse_board", "is_valid", "find_empty", "solve"}},
		{"i_before_e", "i_before_e/i_before_e.able", []string{"is_valid"}},
	})
}

func TestCompilerInferredNilLoopReturnUsesNativeCarrier(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn complete(n: i32) {",
		"  if n < 0 { return }",
		"  loop {",
		"    if n == 0 { break }",
		"    n = n - 1",
		"  }",
		"  nil",
		"}",
		"",
		"fn main() -> void {",
		"  complete(2)",
		"  complete(-1)",
		"  print(\"ok\")",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackSource(t, source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_complete")
	if !ok {
		t.Fatal("could not find compiled inferred-nil function")
	}
	if !strings.Contains(body, ") (runtime.NilValue, *__ableControl)") {
		t.Fatalf("expected inferred-nil function to use the native nil carrier:\n%s", body)
	}
	for _, fragment := range []string{"runtime.Value", "__able_raise_return_type_mismatch"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inferred-nil function to avoid %q:\n%s", fragment, body)
		}
	}

	stdout := compileAndRunSourceWithOptions(t, "ablec-inferred-nil-loop-", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if stdout != "ok\n" {
		t.Fatalf("expected inferred-nil function to complete normally, got %q", stdout)
	}
}

func TestCompilerInferredNilWithValueBranchPreservesValue(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn make_id(id) {",
		"  if id > 0 { id } else { nil }",
		"}",
		"",
		"fn main() -> void {",
		"  print(make_id(7))",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackSource(t, source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_make_id")
	if !ok {
		t.Fatal("could not find compiled inferred nullable function")
	}
	if strings.Contains(body, ") (runtime.NilValue, *__ableControl)") {
		t.Fatalf("expected inferred nullable function to preserve its value branch:\n%s", body)
	}

	stdout := compileAndRunSourceWithOptions(t, "ablec-inferred-nil-value-branch-", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if stdout != "7\n" {
		t.Fatalf("expected inferred nullable function to preserve 7, got %q", stdout)
	}
}

func TestCompilerInferredImplicitVoidUsesStaticCarrier(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn visit(values: Array i32, n: i32) {",
		"  if n < 0 { return }",
		"  if n > 0 {",
		"    values.write_slot(0, n)",
		"    visit(values, 0)",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  values: Array i32 := [0]",
		"  visit(values, -1)",
		"  visit(values, 7)",
		"  print(values.read_slot(0))",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackSource(t, source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_visit")
	if !ok {
		t.Fatal("could not find compiled inferred-void function")
	}
	if !strings.Contains(body, ") (struct{}, *__ableControl)") {
		t.Fatalf("expected inferred side-effect-only function to use the static void carrier:\n%s", body)
	}
	for _, fragment := range []string{"runtime.Value", "__able_any_to_value(", "__able_raise_return_type_mismatch"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inferred side-effect-only function to avoid %q:\n%s", fragment, body)
		}
	}

	stdout := compileAndRunSourceWithOptions(t, "ablec-inferred-void-", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if stdout != "7\n" {
		t.Fatalf("expected inferred-void function to preserve effects, got %q", stdout)
	}
}

func TestCompilerDiscardedLoopBreakValueStillEvaluates(t *testing.T) {
	stdout := compileAndRunSourceWithOptions(t, "ablec-discarded-loop-break-value-", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  loop { break print(\"inside\") }",
		"  print(\"after\")",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if stdout != "inside\nafter\n" {
		t.Fatalf("expected discarded break value to run before the loop exits, got %q", stdout)
	}
}

func auditClosedBenchmarkComputeKernels(t *testing.T, cases []closedBenchmarkSourceAuditCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := compileClosedBenchmarkSource(t, tc.entryPath)
			for _, function := range tc.functions {
				body, ok := findCompiledFunction(result, "__able_compiled_fn_"+function)
				if !ok {
					t.Fatalf("could not find compiled compute kernel %q", function)
				}
				assertClosedBenchmarkKernelStatic(t, tc.name, function, body)
			}
		})
	}
}

func compileClosedBenchmarkSource(t *testing.T, relEntryPath string) *Result {
	t.Helper()
	entryPath := filepath.Join(repositoryRoot(), "v12", "examples", "benchmarks", filepath.FromSlash(relEntryPath))
	if _, err := os.Stat(entryPath); err != nil {
		t.Fatalf("benchmark entry %q: %v", relEntryPath, err)
	}
	searchPaths, err := buildExecSearchPaths(entryPath, filepath.Dir(entryPath), interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("benchmark search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("benchmark loader: %v", err)
	}
	t.Cleanup(func() { loader.Close() })
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load benchmark %q: %v", relEntryPath, err)
	}
	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile benchmark %q without fallback: %v", relEntryPath, err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("benchmark %q emitted compiler fallbacks: %v", relEntryPath, result.Fallbacks)
	}
	return result
}

func assertClosedBenchmarkKernelStatic(t *testing.T, benchmark string, function string, body string) {
	t.Helper()
	for _, fragment := range []string{
		"runtime.Value",
		"[]runtime.Value",
		"__able_call_value(",
		"__able_call_value_fast(",
		"__able_call_named(",
		"__able_method_call_node(",
		"__able_member_get(",
		"__able_member_set(",
		"__able_any_to_value(",
		"bridge.PushCallFrame(__able_runtime,",
		"bridge.PopCallFrame(__able_runtime)",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected %s/%s compute kernel to avoid dynamic fragment %q:\n%s", benchmark, function, fragment, body)
		}
	}
}
