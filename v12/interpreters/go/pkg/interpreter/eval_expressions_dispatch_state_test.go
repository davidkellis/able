package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestEvaluateExpression_IdentifierKeepsPayloadStateLazy(t *testing.T) {
	interp := New()
	payload := &asyncContextPayload{}
	env := runtime.NewEnvironment(nil)
	env.SetRuntimeData(payload)
	env.Define("value", runtime.NewSmallInt(7, runtime.IntegerI32))

	if payload.state != nil {
		t.Fatalf("payload state initialized before expression evaluation")
	}

	got, err := interp.evaluateExpression(ast.ID("value"), env)
	if err != nil {
		t.Fatalf("evaluateExpression failed: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 7 {
		t.Fatalf("expected 7, got %#v", got)
	}
	if payload.state != nil {
		t.Fatalf("expected payload state to remain nil for identifier lookup")
	}
}

func TestEvaluateExpression_IdentifierErrorPreservesCallStack(t *testing.T) {
	interp := New()
	payload := &asyncContextPayload{}
	env := runtime.NewEnvironment(nil)
	env.SetRuntimeData(payload)
	root := runtimeDiagnosticRoot()
	if root == "" {
		t.Fatalf("expected diagnostic root path")
	}

	path := filepath.Join(root, "v12/fixtures/exec/11_03_raise_exit_unhandled/main.able")
	errorNode := ast.ID("missing")
	callNode := ast.Call("outer")
	ast.SetSpan(errorNode, ast.Span{
		Start: ast.Position{Line: 6, Column: 3},
		End:   ast.Position{Line: 6, Column: 10},
	})
	ast.SetSpan(callNode, ast.Span{
		Start: ast.Position{Line: 10, Column: 5},
		End:   ast.Position{Line: 10, Column: 10},
	})
	interp.SetNodeOrigins(map[ast.Node]string{
		errorNode: path,
		callNode:  path,
	})

	interp.PushCallFrame(env, callNode)
	defer interp.PopCallFrame(env)

	_, err := interp.evaluateExpression(errorNode, env)
	if err == nil {
		t.Fatalf("expected missing identifier error")
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
}
