package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestLowerIfExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main(x: i32) -> i32 {\n  if x > 0 {\n    1\n  } else {\n    2\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "entry")
	assertHasBlock(t, fn, "if_then")
	assertHasBlock(t, fn, "if_else")
	assertHasBlock(t, fn, "if_join")
}

func TestLowerMatchExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main(x: i32) -> i32 {\n  x match {\n    case 0 => 1,\n    case _ => 2\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "entry")
	assertHasBlock(t, fn, "match_case_0")
	assertHasBlock(t, fn, "match_case_1")
	assertHasBlock(t, fn, "match_join")
	assertHasBlock(t, fn, "match_nomatch")
}

func TestLowerWhileLoopCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  count := 0\n  while count < 3 {\n    count = count + 1\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "while_cond")
	assertHasBlock(t, fn, "while_body")
	assertHasBlock(t, fn, "while_exit")
}

func TestLowerLoopExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> i32 {\n  count := 0\n  loop {\n    if count == 3 { break count }\n    count = count + 1\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "loop_body")
	assertHasBlock(t, fn, "loop_exit")
}

func TestLowerForLoopCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  sum := 0\n  arr := [1, 2, 3]\n  for v in arr {\n    sum = sum + v\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "for_next")
	assertHasBlock(t, fn, "for_body")
	assertHasBlock(t, fn, "for_exit")
}

func TestLowerOrElseExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  value := (if true { nil } else { 1 }) or { err => 9 }\n  value = value\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "or_fail_error")
	assertHasBlock(t, fn, "or_fail_nil")
	assertHasBlock(t, fn, "or_join")
}

func TestLowerAssignmentAllowsImplicitBinding(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  implicit = 7\n  implicit = implicit\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
}

func TestLowerRescueEnsureExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  result := \"\"\n  do {\n    raise \"boom\"\n  } rescue {\n    case _ => { result = \"rescued\" }\n  } ensure {\n    result = result\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "rescue_handler")
	assertHasBlock(t, fn, "rescue_case_0")
	assertHasBlock(t, fn, "ensure_block")
	assertHasBlock(t, fn, "ensure_ok")
	assertHasBlock(t, fn, "ensure_err")
	assertHasBlock(t, fn, "ensure_join")
}

func TestLowerSpawnAwaitDoesNotError(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  worker := spawn { 21 }\n  res := await [worker]\n  res = res\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
}

func TestLowerBreakpointExpressionCreatesBlocks(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> i32 {\n  breakpoint 'done {\n    break 'done 7\n  }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
	assertHasBlock(t, fn, "breakpoint_body")
	assertHasBlock(t, fn, "breakpoint_exit")
}

func TestLowerLiteralExpressions(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  arr := [1, 2]\n  map := #{ \"a\": 1, \"b\": 2 }\n  msg := `hi ${arr.size()}`\n  iter := Iterator { gen.yield(1) }\n  arr = arr\n  map = map\n  msg = msg\n  iter = iter\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	if err := ValidateIR(fn); err != nil {
		t.Fatalf("validate IR: %v", err)
	}
}

func writeTestPackage(t *testing.T, root, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: "+name+"\n"), 0o644); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
}

func loadTestProgram(t *testing.T, entryPath string) *driver.Program {
	t.Helper()
	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	return program
}

func findFunction(t *testing.T, module *ast.Module, name string) *ast.FunctionDefinition {
	t.Helper()
	if module == nil {
		t.Fatalf("missing module")
	}
	for _, stmt := range module.Body {
		if def, ok := stmt.(*ast.FunctionDefinition); ok && def != nil && def.ID != nil && def.ID.Name == name {
			return def
		}
	}
	t.Fatalf("missing function %s", name)
	return nil
}

func assertHasBlock(t *testing.T, fn *IRFunction, label string) {
	t.Helper()
	if fn == nil || fn.Blocks == nil {
		t.Fatalf("missing blocks")
	}
	if _, ok := fn.Blocks[label]; !ok {
		t.Fatalf("missing block %s", label)
	}
}
