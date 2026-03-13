package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

// TestCompilerStructMethodStaticDispatch verifies that method calls on struct
// locals (declared via :=) use direct compiled method calls instead of
// __able_member_get_method + __able_call_value dynamic dispatch.
func TestCompilerStructMethodStaticDispatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #inc() -> i32 { #value + 1 }",
		"}",
		"",
		"fn bump() -> i32 {",
		"  counter := Counter { value: 2 }",
		"  counter.inc()",
		"}",
		"",
		"fn main() -> void {",
		"  print(bump())",
		"}",
		"",
	}, "\n")
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	compiledBytes, ok := result.Files["compiled.go"]
	if !ok || len(compiledBytes) == 0 {
		t.Fatalf("compiled.go not found in output")
	}
	compiled := string(compiledBytes)

	// Extract the bump() function body.
	funcBody := extractCompiledFunctionBody(compiled, "fn_bump")
	if funcBody == "" {
		t.Fatalf("fn_bump not found in compiled output")
	}

	// The method call should use direct compiled dispatch, not dynamic.
	if strings.Contains(funcBody, "__able_member_get_method") {
		t.Errorf("fn_bump should not use __able_member_get_method:\n%s", funcBody)
	}
	if strings.Contains(funcBody, "__able_call_value") {
		t.Errorf("fn_bump should not use __able_call_value:\n%s", funcBody)
	}
	if !strings.Contains(funcBody, "__able_compiled_method_Counter_inc") {
		t.Errorf("fn_bump should call __able_compiled_method_Counter_inc:\n%s", funcBody)
	}
	if strings.Contains(funcBody, "__able_struct_Counter_from") || strings.Contains(funcBody, "__able_struct_Counter_to") {
		t.Errorf("fn_bump should keep native Counter locals instead of runtime extract/writeback:\n%s", funcBody)
	}
	if !strings.Contains(funcBody, "var counter *Counter =") {
		t.Errorf("fn_bump should keep counter as a native *Counter local:\n%s", funcBody)
	}
}

// TestCompilerStructFieldStaticAccess verifies that field access on struct
// locals uses direct Go field access instead of __able_member_get.
func TestCompilerStructFieldStaticAccess(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"fn sum() -> i32 {",
		"  p := Point { x: 3, y: 4 }",
		"  p.x + p.y",
		"}",
		"",
		"fn main() -> void {",
		"  print(sum())",
		"}",
		"",
	}, "\n")
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	compiledBytes, ok := result.Files["compiled.go"]
	if !ok || len(compiledBytes) == 0 {
		t.Fatalf("compiled.go not found in output")
	}
	compiled := string(compiledBytes)

	funcBody := extractCompiledFunctionBody(compiled, "fn_sum")
	if funcBody == "" {
		t.Fatalf("fn_sum not found in compiled output")
	}

	if strings.Contains(funcBody, "__able_member_get(") {
		t.Errorf("fn_sum should not use __able_member_get:\n%s", funcBody)
	}
	if !strings.Contains(funcBody, ".X") || !strings.Contains(funcBody, ".Y") {
		t.Errorf("fn_sum should access .X and .Y directly:\n%s", funcBody)
	}
	if strings.Contains(funcBody, "__able_struct_Point_from") || strings.Contains(funcBody, "__able_struct_Point_to") {
		t.Errorf("fn_sum should keep native Point locals instead of runtime extract/writeback:\n%s", funcBody)
	}
	if !strings.Contains(funcBody, "var p *Point =") {
		t.Errorf("fn_sum should keep p as a native *Point local:\n%s", funcBody)
	}
}

// TestCompilerDefaultImplSiblingDirectCall verifies that default interface
// method implementations calling sibling methods use direct compiled calls.
func TestCompilerDefaultImplSiblingDirectCall(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"interface MyEq {",
		"  fn eq(self: Self, other: Self) -> bool",
		"  fn ne(self: Self, other: Self) -> bool { !self.eq(other) }",
		"}",
		"",
		"impl MyEq for Point {",
		"  fn eq(self: Self, other: Self) -> bool { self.x == other.x && self.y == other.y }",
		"}",
		"",
		"fn main() -> void {",
		"  a := Point { x: 1, y: 2 }",
		"  b := Point { x: 1, y: 3 }",
		"  print(a.ne(b))",
		"}",
		"",
	}, "\n")
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	compiledBytes, ok := result.Files["compiled.go"]
	if !ok || len(compiledBytes) == 0 {
		t.Fatalf("compiled.go not found in output")
	}
	compiled := string(compiledBytes)

	neBody := extractCompiledFunctionBody(compiled, "impl_MyEq_ne_default")
	if neBody == "" {
		t.Fatalf("ne default impl not found in compiled output")
	}

	if strings.Contains(neBody, "__able_impl_self_method") {
		t.Errorf("ne default should not use __able_impl_self_method:\n%s", neBody)
	}
	if strings.Contains(neBody, "__able_call_value") {
		t.Errorf("ne default should not use __able_call_value:\n%s", neBody)
	}
	if strings.Contains(neBody, "__able_member_get_method") {
		t.Errorf("ne default should not use __able_member_get_method:\n%s", neBody)
	}
	if !strings.Contains(neBody, "__able_compiled_impl_MyEq_eq") {
		t.Errorf("ne default should call __able_compiled_impl_MyEq_eq directly:\n%s", neBody)
	}
}

func TestCompilerStructFunctionParamAndReturnStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"fn make_point() -> Point {",
		"  Point { x: 3, y: 4 }",
		"}",
		"",
		"fn sum_point(p: Point) -> i32 {",
		"  p.x + p.y",
		"}",
		"",
		"fn main() -> i32 {",
		"  p := make_point()",
		"  sum_point(p)",
		"}",
		"",
	}, "\n"))

	compiled := string(result.Files["compiled.go"])
	if !strings.Contains(compiled, "func __able_compiled_fn_make_point() (*Point, *__ableControl)") {
		t.Fatalf("make_point should return native *Point:\n%s", compiled)
	}
	if !strings.Contains(compiled, "func __able_compiled_fn_sum_point(p *Point) (int32, *__ableControl)") {
		t.Fatalf("sum_point should accept native *Point:\n%s", compiled)
	}

	mainBody := extractCompiledFunctionBody(compiled, "fn_main")
	if mainBody == "" {
		t.Fatalf("fn_main not found in compiled output")
	}
	if strings.Contains(mainBody, "__able_struct_Point_from") || strings.Contains(mainBody, "__able_struct_Point_to") {
		t.Fatalf("fn_main should keep Point on the native static path:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "var p *Point =") {
		t.Fatalf("fn_main should keep p as a native *Point local:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_fn_make_point()") || !strings.Contains(mainBody, "__able_compiled_fn_sum_point(") {
		t.Fatalf("fn_main should call compiled native functions directly:\n%s", mainBody)
	}

	sumBody := extractCompiledFunctionBody(compiled, "fn_sum_point")
	if sumBody == "" {
		t.Fatalf("fn_sum_point not found in compiled output")
	}
	if strings.Contains(sumBody, "__able_struct_Point_from") || strings.Contains(sumBody, "__able_member_get(") {
		t.Fatalf("fn_sum_point should use native Point access:\n%s", sumBody)
	}
	if !strings.Contains(sumBody, "p.X") || !strings.Contains(sumBody, "p.Y") {
		t.Fatalf("fn_sum_point should use native field access:\n%s", sumBody)
	}
}

func TestCompilerStructMutationAcrossStaticFunctionCallStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Counter { value: i32 }",
		"",
		"fn bump(counter: Counter) -> void {",
		"  counter.value = counter.value + 1",
		"}",
		"",
		"fn main() -> i32 {",
		"  counter := Counter { value: 1 }",
		"  bump(counter)",
		"  counter.value",
		"}",
		"",
	}, "\n"))

	compiled := string(result.Files["compiled.go"])
	if !strings.Contains(compiled, "func __able_compiled_fn_bump(counter *Counter) (struct{}, *__ableControl)") {
		t.Fatalf("bump should accept native *Counter:\n%s", compiled)
	}

	mainBody := extractCompiledFunctionBody(compiled, "fn_main")
	if mainBody == "" {
		t.Fatalf("fn_main not found in compiled output")
	}
	if strings.Contains(mainBody, "__able_struct_Counter_from") || strings.Contains(mainBody, "__able_struct_Counter_to") {
		t.Fatalf("fn_main should keep Counter on the native static path:\n%s", mainBody)
	}
	if !strings.Contains(mainBody, "__able_compiled_fn_bump(counter)") {
		t.Fatalf("fn_main should call bump with the native pointer local:\n%s", mainBody)
	}

	bumpBody := extractCompiledFunctionBody(compiled, "fn_bump")
	if bumpBody == "" {
		t.Fatalf("fn_bump not found in compiled output")
	}
	if strings.Contains(bumpBody, "__able_struct_Counter_from") || strings.Contains(bumpBody, "__able_member_set(") {
		t.Fatalf("fn_bump should mutate Counter natively:\n%s", bumpBody)
	}
	if !strings.Contains(bumpBody, "counter.Value =") {
		t.Fatalf("fn_bump should mutate the native Counter field directly:\n%s", bumpBody)
	}
}

func TestCompilerStructWrapperReturnUsesExplicitStructConverter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"fn make_point() -> Point {",
		"  Point { x: 3, y: 4 }",
		"}",
		"",
	}, "\n"))

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_make_point")
	if !ok {
		t.Fatalf("could not find wrapper for make_point")
	}
	if !strings.Contains(wrapBody, "return __able_struct_Point_to(rt, compiledResult)") {
		t.Fatalf("expected wrapper to use explicit Point boundary conversion:\n%s", wrapBody)
	}
	if strings.Contains(wrapBody, "__able_any_to_value(compiledResult)") {
		t.Fatalf("wrapper should not route Point return through __able_any_to_value:\n%s", wrapBody)
	}
}

// extractCompiledFunctionBody finds the func definition for a compiled function
// containing the given name fragment and returns its body.
func extractCompiledFunctionBody(code string, nameFragment string) string {
	marker := "func __able_compiled_" + nameFragment
	idx := strings.Index(code, marker)
	if idx < 0 {
		return ""
	}
	lineEnd := strings.Index(code[idx:], "\n")
	if lineEnd < 0 {
		return ""
	}
	signature := code[idx : idx+lineEnd]
	braceStart := strings.LastIndex(signature, "{")
	if braceStart < 0 {
		return ""
	}
	start := idx + braceStart
	depth := 0
	for i := start; i < len(code); i++ {
		switch code[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return code[start : i+1]
			}
		}
	}
	return ""
}
