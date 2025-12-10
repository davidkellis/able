package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func buildTypeMismatchModule() *ast.Module {
	badFn := ast.Fn(
		"bad",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	return ast.NewModule([]ast.Statement{badFn}, nil, nil)
}

func TestInterpreterTypecheckerDiagnosticsNonStrict(t *testing.T) {
	module := buildTypeMismatchModule()
	interp := New()
	before := interp.GlobalEnvironment().Snapshot()

	interp.EnableTypechecker(TypecheckConfig{})

	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("unexpected interpreter error: %v", err)
	}

	diags := interp.TypecheckDiagnostics()
	if len(diags) == 0 {
		t.Fatalf("expected typechecker diagnostics")
	}
	if want := "return expects String"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected first diagnostic to mention %q, got %q", want, diags[0].Message)
	}

	after := interp.GlobalEnvironment().Snapshot()
	if _, ok := after["bad"]; !ok {
		t.Fatalf("expected interpreter to register function despite diagnostics")
	}
	if _, ok := before["bad"]; ok {
		t.Fatalf("expected function to be absent before evaluation")
	}
}

func TestInterpreterTypecheckerStrictModePreventsEvaluation(t *testing.T) {
	module := buildTypeMismatchModule()
	interp := New()
	before := interp.GlobalEnvironment().Snapshot()

	interp.EnableTypechecker(TypecheckConfig{FailFast: true})

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected interpreter to stop on typechecker diagnostics")
	}
	if want := "return expects String"; !strings.Contains(err.Error(), want) {
		t.Fatalf("expected interpreter error to mention %q, got %q", want, err.Error())
	}

	diags := interp.TypecheckDiagnostics()
	if len(diags) == 0 {
		t.Fatalf("expected stored typechecker diagnostics when evaluation aborted")
	}

	after := interp.GlobalEnvironment().Snapshot()
	if _, ok := after["bad"]; ok {
		t.Fatalf("expected interpreter to skip module execution when fail-fast is enabled")
	}

	// Ensure baseline bindings remain unchanged.
	if len(after) != len(before) {
		t.Fatalf("expected global environment to remain unchanged, before=%d after=%d", len(before), len(after))
	}
}
