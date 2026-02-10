package compiler

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIREmitFunctionSimple(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main(x: i32) -> i32 {\n  x + 1\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionIf(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main(x: i32) -> i32 {\n  if x > 0 { 1 } else { 2 }\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionLiterals(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "package demo\n\nstruct Point { x: i32, y: i32 }\n\nfn main() -> void {\n  arr := [1, 2]\n  msg := `hi ${1}`\n  point := Point { x: 1, y: 2 }\n  arr = arr\n  msg = msg\n  point = point\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionCast(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main(x: i32) -> i32 {\n  (x as i64) as i32\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionSpawn(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  handle := spawn { 1 }\n  handle = handle\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	if !strings.Contains(string(src), "__able_spawn") {
		t.Fatalf("expected spawn codegen to call __able_spawn")
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionAwait(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  _ = await [1]\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	if !strings.Contains(string(src), "bridge.Await") {
		t.Fatalf("expected await codegen to call bridge.Await")
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitFunctionIteratorLiteral(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void {\n  iter := Iterator { gen.yield(1) }\n  iter = iter\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	if !strings.Contains(string(src), "__able_new_iterator") {
		t.Fatalf("expected iterator literal to use __able_new_iterator")
	}
	if !strings.Contains(string(src), "iterator.next re-entered while suspended at yield") {
		t.Fatalf("expected iterator helper to enforce re-entrancy errors")
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}

func TestIREmitPatternAssignmentExistingBinding(t *testing.T) {
	root := t.TempDir()
	writeTestPackage(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "struct Point { x: i32 }\n\nfn main() -> void {\n  x := 1\n  Point { x } = Point { x: 2 }\n  x = x\n}\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	program := loadTestProgram(t, entryPath)
	def := findFunction(t, program.Entry.AST, "main")

	fn, err := LowerFunction(def, program.Entry.Package, nil)
	if err != nil {
		t.Fatalf("lower function: %v", err)
	}
	src, err := EmitIRFunction(fn, IRGoOptions{PackageName: "main"})
	if err != nil {
		t.Fatalf("emit IR: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ir_generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("parse generated source: %v", err)
	}
}
