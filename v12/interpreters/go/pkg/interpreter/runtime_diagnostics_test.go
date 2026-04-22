package interpreter

import (
	"fmt"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestRuntimeDiagnosticsFormatting(t *testing.T) {
	interp := New()
	state := newEvalState()
	root := runtimeDiagnosticRoot()
	if root == "" {
		t.Fatalf("expected diagnostic root path")
	}

	path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
	errorNode := ast.ID("boom")
	callNode := ast.Call("boom")
	ast.SetSpan(errorNode, ast.Span{
		Start: ast.Position{Line: 6, Column: 3},
		End:   ast.Position{Line: 6, Column: 7},
	})
	ast.SetSpan(callNode, ast.Span{
		Start: ast.Position{Line: 10, Column: 5},
		End:   ast.Position{Line: 10, Column: 9},
	})

	interp.SetNodeOrigins(map[ast.Node]string{
		errorNode: path,
		callNode:  path,
	})
	state.pushCallFrame(callNode)

	var err error = raiseSignal{
		value: runtime.ErrorValue{Message: "boom"},
	}
	err = interp.attachRuntimeContext(err, errorNode, state)

	diag := interp.BuildRuntimeDiagnostic(err)
	got := DescribeRuntimeDiagnostic(diag)
	expectedPath := normalizeRuntimePath(path)
	expected := fmt.Sprintf(
		"runtime: %s:6:3 boom\nnote: %s:10:5 called from here",
		expectedPath,
		expectedPath,
	)
	if got != expected {
		t.Fatalf("unexpected diagnostic output:\nexpected: %s\ngot: %s", expected, got)
	}
}

func TestAttachRuntimeContextWithExplicitCallStack(t *testing.T) {
	interp := New()
	root := runtimeDiagnosticRoot()
	if root == "" {
		t.Fatalf("expected diagnostic root path")
	}

	path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
	errorNode := ast.ID("boom")
	callNode := ast.Call("boom")
	ast.SetSpan(errorNode, ast.Span{
		Start: ast.Position{Line: 6, Column: 3},
		End:   ast.Position{Line: 6, Column: 7},
	})
	ast.SetSpan(callNode, ast.Span{
		Start: ast.Position{Line: 10, Column: 5},
		End:   ast.Position{Line: 10, Column: 9},
	})

	interp.SetNodeOrigins(map[ast.Node]string{
		errorNode: path,
		callNode:  path,
	})

	var err error = raiseSignal{
		value: runtime.ErrorValue{Message: "boom"},
	}
	err = interp.AttachRuntimeContextWithCallStack(err, errorNode, nil, []*ast.FunctionCall{callNode})

	diag := interp.BuildRuntimeDiagnostic(err)
	got := DescribeRuntimeDiagnostic(diag)
	expectedPath := normalizeRuntimePath(path)
	expected := fmt.Sprintf(
		"runtime: %s:6:3 boom\nnote: %s:10:5 called from here",
		expectedPath,
		expectedPath,
	)
	if got != expected {
		t.Fatalf("unexpected explicit-stack diagnostic output:\nexpected: %s\ngot: %s", expected, got)
	}
}

func TestAppendRuntimeCallFrame(t *testing.T) {
	interp := New()
	root := runtimeDiagnosticRoot()
	if root == "" {
		t.Fatalf("expected diagnostic root path")
	}

	path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
	errorNode := ast.ID("boom")
	innerCall := ast.Call("inner")
	outerCall := ast.Call("outer")
	ast.SetSpan(errorNode, ast.Span{
		Start: ast.Position{Line: 6, Column: 3},
		End:   ast.Position{Line: 6, Column: 7},
	})
	ast.SetSpan(innerCall, ast.Span{
		Start: ast.Position{Line: 10, Column: 5},
		End:   ast.Position{Line: 10, Column: 10},
	})
	ast.SetSpan(outerCall, ast.Span{
		Start: ast.Position{Line: 14, Column: 7},
		End:   ast.Position{Line: 14, Column: 12},
	})

	interp.SetNodeOrigins(map[ast.Node]string{
		errorNode: path,
		innerCall: path,
		outerCall: path,
	})

	var err error = raiseSignal{
		value: runtime.ErrorValue{Message: "boom"},
	}
	err = interp.AttachRuntimeContextWithCallStack(err, errorNode, nil, nil)
	err = interp.AppendRuntimeCallFrame(err, innerCall)
	err = interp.AppendRuntimeCallFrame(err, outerCall)

	diag := interp.BuildRuntimeDiagnostic(err)
	got := DescribeRuntimeDiagnostic(diag)
	expectedPath := normalizeRuntimePath(path)
	expected := fmt.Sprintf(
		"runtime: %s:6:3 boom\nnote: %s:10:5 called from here\nnote: %s:14:7 called from here",
		expectedPath,
		expectedPath,
		expectedPath,
	)
	if got != expected {
		t.Fatalf("unexpected appended-call-frame diagnostic output:\nexpected: %s\ngot: %s", expected, got)
	}
}

func TestAttachRuntimeContextPreservesCallStackAcrossUnwind(t *testing.T) {
	interp := New()
	state := newEvalState()
	root := runtimeDiagnosticRoot()
	if root == "" {
		t.Fatalf("expected diagnostic root path")
	}

	path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
	errorNode := ast.ID("boom")
	callNode := ast.Call("boom")
	ast.SetSpan(errorNode, ast.Span{
		Start: ast.Position{Line: 6, Column: 3},
		End:   ast.Position{Line: 6, Column: 7},
	})
	ast.SetSpan(callNode, ast.Span{
		Start: ast.Position{Line: 10, Column: 5},
		End:   ast.Position{Line: 10, Column: 9},
	})

	interp.SetNodeOrigins(map[ast.Node]string{
		errorNode: path,
		callNode:  path,
	})

	state.pushCallFrame(callNode)
	var err error = raiseSignal{
		value: runtime.ErrorValue{Message: "boom"},
	}
	err = interp.attachRuntimeContext(err, errorNode, state)
	state.popCallFrame()

	diag := interp.BuildRuntimeDiagnostic(err)
	got := DescribeRuntimeDiagnostic(diag)
	expectedPath := normalizeRuntimePath(path)
	expected := fmt.Sprintf(
		"runtime: %s:6:3 boom\nnote: %s:10:5 called from here",
		expectedPath,
		expectedPath,
	)
	if got != expected {
		t.Fatalf("unexpected unwound diagnostic output:\nexpected: %s\ngot: %s", expected, got)
	}
}

func TestReturnCoercionErrorAttachesRuntimeContextLazily(t *testing.T) {
	for _, tc := range []struct {
		name   string
		create func() *Interpreter
	}{
		{name: "treewalker", create: New},
		{name: "bytecode", create: NewBytecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := tc.create()
			root := runtimeDiagnosticRoot()
			if root == "" {
				t.Fatalf("expected diagnostic root path")
			}

			path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
			retStmt := ast.Ret(ast.Str("oops"))
			callNode := ast.Call("bad")
			ast.SetSpan(retStmt, ast.Span{
				Start: ast.Position{Line: 6, Column: 3},
				End:   ast.Position{Line: 6, Column: 16},
			})
			ast.SetSpan(callNode, ast.Span{
				Start: ast.Position{Line: 10, Column: 5},
				End:   ast.Position{Line: 10, Column: 8},
			})

			module := ast.Mod([]ast.Statement{
				ast.Fn("bad", nil, []ast.Statement{retStmt}, ast.Ty("i32"), nil, nil, false, false),
			}, nil, nil)

			interp.SetNodeOrigins(map[ast.Node]string{
				retStmt:  path,
				callNode: path,
			})

			_, env, err := interp.EvaluateModule(module)
			if err != nil {
				t.Fatalf("evaluate module: %v", err)
			}
			badFn, err := env.Get("bad")
			if err != nil {
				t.Fatalf("lookup bad: %v", err)
			}

			_, err = interp.CallFunctionInWithCallNode(badFn, nil, env, callNode)
			if err == nil {
				t.Fatalf("expected return coercion error")
			}

			diag := interp.BuildRuntimeDiagnostic(err)
			if got := normalizeRuntimePath(diag.Location.Path); got != normalizeRuntimePath(path) {
				t.Fatalf("runtime diagnostic path = %q, want %q", got, normalizeRuntimePath(path))
			}
			if diag.Location.Line != 6 || diag.Location.Column != 3 {
				t.Fatalf("runtime diagnostic location = %d:%d, want 6:3", diag.Location.Line, diag.Location.Column)
			}
			if len(diag.Notes) != 1 {
				t.Fatalf("expected one call note, got %d (%#v)", len(diag.Notes), diag.Notes)
			}
			note := diag.Notes[0]
			if got := normalizeRuntimePath(note.Location.Path); got != normalizeRuntimePath(path) {
				t.Fatalf("note path = %q, want %q", got, normalizeRuntimePath(path))
			}
			if note.Location.Line != 10 || note.Location.Column != 5 {
				t.Fatalf("note location = %d:%d, want 10:5", note.Location.Line, note.Location.Column)
			}
		})
	}
}
