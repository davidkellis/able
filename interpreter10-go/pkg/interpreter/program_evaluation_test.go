package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
)

func TestInterpreterEvaluateProgramTypecheckFailure(t *testing.T) {
	depModule := &driver.Module{
		Package: "dep",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false),
			},
			nil,
			ast.Pkg([]interface{}{"dep"}, false),
		),
		Files: []string{"dep/lib.able"},
	}
	mainModule := &driver.Module{
		Package: "root",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("shout", nil, []ast.Statement{
					ast.Ret(ast.Bin("+", ast.CallExpr(ast.Member(ast.ID("dep"), ast.ID("provide"))), ast.Str("!"))),
				}, ast.Ty("string"), nil, nil, false, false),
			},
			[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, nil)},
			ast.Pkg([]interface{}{"root"}, false),
		),
		Files:   []string{"root/main.able"},
		Imports: []string{"dep"},
	}
	program := &driver.Program{
		Entry:   mainModule,
		Modules: []*driver.Module{depModule, mainModule},
	}

	interp := New()
	value, entryEnv, diags, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
	if value != nil {
		t.Fatalf("expected nil value when diagnostics produced")
	}
	if entryEnv != nil {
		t.Fatalf("expected nil entry environment when diagnostics produced")
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for type mismatch")
	}
	if want := "requires both operands"; !strings.Contains(diags[0].Diagnostic.Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Diagnostic.Message)
	}
}

func TestInterpreterEvaluateProgramSuccess(t *testing.T) {
	depModule := &driver.Module{
		Package: "dep",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Str("dep"))}, ast.Ty("string"), nil, nil, false, false),
			},
			nil,
			ast.Pkg([]interface{}{"dep"}, false),
		),
		Files: []string{"dep/lib.able"},
	}
	mainModule := &driver.Module{
		Package: "root",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("shout", nil, []ast.Statement{
					ast.Ret(ast.Bin("+", ast.CallExpr(ast.Member(ast.ID("dep"), ast.ID("provide"))), ast.Str("!"))),
				}, ast.Ty("string"), nil, nil, false, false),
			},
			[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, nil)},
			ast.Pkg([]interface{}{"root"}, false),
		),
		Files:   []string{"root/main.able"},
		Imports: []string{"dep"},
	}
	program := &driver.Program{
		Entry:   mainModule,
		Modules: []*driver.Module{depModule, mainModule},
	}

	interp := New()
	value, entryEnv, diags, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if entryEnv == nil {
		t.Fatalf("expected non-nil entry environment")
	}
	if value == nil {
		t.Fatalf("expected non-nil value from entry module")
	}
	if _, err := entryEnv.Get("shout"); err != nil {
		t.Fatalf("entry environment missing shout: %v", err)
	}
}

func TestInterpreterEvaluateProgramAllowsDiagnostics(t *testing.T) {
	depModule := &driver.Module{
		Package: "dep",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Str("dep"))}, ast.Ty("string"), nil, nil, false, false),
			},
			nil,
			ast.Pkg([]interface{}{"dep"}, false),
		),
		Files: []string{"dep/lib.able"},
	}
	mainModule := &driver.Module{
		Package: "root",
		AST: ast.Mod(
			[]ast.Statement{
				ast.Fn("shout", nil, []ast.Statement{
					ast.Ret(ast.CallExpr(ast.Member(ast.ID("dep"), ast.ID("provide")))),
				}, ast.Ty("i32"), nil, nil, false, false),
			},
			[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, nil)},
			ast.Pkg([]interface{}{"root"}, false),
		),
		Files:   []string{"root/main.able"},
		Imports: []string{"dep"},
	}
	program := &driver.Program{
		Entry:   mainModule,
		Modules: []*driver.Module{depModule, mainModule},
	}

	interp := New()
	value, entryEnv, diags, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		AllowDiagnostics: true,
	})
	if err != nil {
		t.Fatalf("EvaluateProgram error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics when allowing typecheck failures")
	}
	if entryEnv == nil {
		t.Fatalf("expected entry environment even when diagnostics present")
	}
	if value == nil {
		t.Fatalf("expected entry value even when diagnostics present")
	}
	if want := "return expects Int:i32"; !strings.Contains(diags[0].Diagnostic.Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, diags[0].Diagnostic.Message)
	}
}
