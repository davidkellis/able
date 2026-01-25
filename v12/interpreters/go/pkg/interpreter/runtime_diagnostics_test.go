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
