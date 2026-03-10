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
	// CSE: struct should be extracted at most once (p.x + p.y reuses extraction).
	fromCount := strings.Count(funcBody, "__able_struct_Point_from")
	if fromCount > 1 {
		t.Errorf("fn_sum should extract struct at most once, got %d extractions:\n%s", fromCount, funcBody)
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

// extractCompiledFunctionBody finds the func definition for a compiled function
// containing the given name fragment and returns its body.
func extractCompiledFunctionBody(code string, nameFragment string) string {
	marker := "func __able_compiled_" + nameFragment
	idx := strings.Index(code, marker)
	if idx < 0 {
		return ""
	}
	braceStart := strings.Index(code[idx:], "{")
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
