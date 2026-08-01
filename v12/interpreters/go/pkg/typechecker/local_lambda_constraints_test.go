package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func localIdentityLambda() *ast.LambdaExpression {
	return ast.NewLambdaExpression(
		[]*ast.FunctionParameter{ast.Param("value", nil)},
		ast.ID("value"),
		nil,
		nil,
		nil,
		false,
	)
}

func callableConsumer(name, scalar string) *ast.FunctionDefinition {
	return ast.Fn(
		name,
		[]*ast.FunctionParameter{
			ast.Param("callback", ast.FnType(
				[]ast.TypeExpression{ast.Ty(scalar)},
				ast.Ty(scalar),
			)),
		},
		[]ast.Statement{ast.CallExpr(ast.ID("callback"), ast.Int(1))},
		ast.Ty(scalar),
		nil,
		nil,
		false,
		false,
	)
}

func checkLocalLambdaModule(t *testing.T, body ...ast.Statement) (*Checker, []Diagnostic) {
	t.Helper()
	checker := New()
	module := ast.NewModule(body, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("check module: %v", err)
	}
	return checker, diags
}

func TestLocalLambdaCompatibleCallableConstraintsInferOneSignature(t *testing.T) {
	lambda := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		callableConsumer("apply_i32", "i32"),
		ast.Assign(ast.ID("callback"), lambda),
		ast.Call("apply_i32", ast.ID("callback")),
		ast.Call("apply_i32", ast.ID("callback")),
	)
	if len(diags) != 0 {
		t.Fatalf("expected compatible constraints, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected inferred local lambda signature")
	}
	want := FunctionType{
		Params: []Type{IntegerType{Suffix: "i32"}},
		Return: IntegerType{Suffix: "i32"},
	}
	if !exactTypeEquivalent(inferred, want) {
		t.Fatalf("expected fn(i32) -> i32, got %s", formatTypeForReturnDiagnostic(inferred))
	}
}

func TestLocalLambdaImplicitDeclarationUsesCallableConstraint(t *testing.T) {
	lambda := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		callableConsumer("apply_i64", "i64"),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("callback"), lambda),
		ast.Call("apply_i64", ast.ID("callback")),
	)
	if len(diags) != 0 {
		t.Fatalf("expected implicit declaration constraint, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected inferred implicitly declared lambda signature")
	}
	if got := formatTypeForReturnDiagnostic(inferred); got != "fn(i64) -> i64" {
		t.Fatalf("expected fn(i64) -> i64, got %s", got)
	}
}

func TestLocalLambdaConflictingCallableConstraintsAreCoded(t *testing.T) {
	_, diags := checkLocalLambdaModule(
		t,
		callableConsumer("apply_i32", "i32"),
		callableConsumer("apply_i64", "i64"),
		ast.Assign(ast.ID("callback"), localIdentityLambda()),
		ast.Call("apply_i32", ast.ID("callback")),
		ast.Call("apply_i64", ast.ID("callback")),
	)
	var matched *Diagnostic
	for idx := range diags {
		if diags[idx].Code == DiagnosticCodeCallableSignatureMismatch {
			matched = &diags[idx]
			break
		}
	}
	if matched == nil {
		t.Fatalf("expected coded callable conflict, got %v", diags)
	}
	if !strings.Contains(matched.Message, "fn(i32) -> i32 and fn(i64) -> i64") {
		t.Fatalf("unexpected conflict diagnostic: %s", matched.Message)
	}
	if len(matched.Notes) != 1 {
		t.Fatalf("expected first-constraint note, got %v", matched.Notes)
	}
}

func TestLocalLambdaDirectInvocationInfersCompleteSignature(t *testing.T) {
	lambda := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		ast.Assign(ast.ID("callback"), lambda),
		ast.CallExpr(ast.ID("callback"), ast.Int(1)),
		ast.CallExpr(ast.ID("callback"), ast.Int(2)),
	)
	if len(diags) != 0 {
		t.Fatalf("expected compatible direct invocations, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected inferred direct-invocation signature")
	}
	if got := formatTypeForReturnDiagnostic(inferred); got != "fn(i32) -> i32" {
		t.Fatalf("expected fn(i32) -> i32, got %s", got)
	}
}

func TestLocalLambdaTypedBindingUsesCallableContext(t *testing.T) {
	lambda := localIdentityLambda()
	binding := ast.TypedP(
		ast.ID("callback"),
		ast.FnType(
			[]ast.TypeExpression{ast.Ty("i64")},
			ast.Ty("i64"),
		),
	)
	checker, diags := checkLocalLambdaModule(
		t,
		ast.Assign(binding, lambda),
	)
	if len(diags) != 0 {
		t.Fatalf("expected typed binding constraint, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected typed binding to infer the lambda signature")
	}
	if got := formatTypeForReturnDiagnostic(inferred); got != "fn(i64) -> i64" {
		bindingType, _ := checker.infer.get(binding)
		t.Fatalf(
			"expected fn(i64) -> i64, got %s (binding %s)",
			got,
			formatTypeForReturnDiagnostic(bindingType),
		)
	}
}

func TestLocalCallableAnnotationUsesLexicalRenamedType(t *testing.T) {
	checker := New()
	env := checker.global.Extend()
	remoteThing := StructType{
		StructName: "Thing",
		Fields: map[string]Type{
			"remote": IntegerType{Suffix: "i32"},
		},
	}
	env.Define("RemoteThing", remoteThing)

	resolved := checker.resolveLocalTypeReference(
		env,
		ast.FnType(nil, ast.Ty("RemoteThing")),
	)
	callable, ok := resolved.(FunctionType)
	if !ok {
		t.Fatalf("expected callable type, got %s", formatTypeForReturnDiagnostic(resolved))
	}
	result, ok := callable.Return.(StructType)
	if !ok {
		t.Fatalf("expected renamed struct result, got %s", formatTypeForReturnDiagnostic(callable.Return))
	}
	if _, ok := result.Fields["remote"]; !ok {
		t.Fatalf("expected lexical RemoteThing type, got %#v", result)
	}
}

func TestLocalLambdaReturnContextConstrainsBinding(t *testing.T) {
	lambda := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		ast.Fn(
			"identity_factory",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("callback"), lambda),
				ast.Ret(ast.ID("callback")),
			},
			ast.FnType(
				[]ast.TypeExpression{ast.Ty("i64")},
				ast.Ty("i64"),
			),
			nil,
			nil,
			false,
			false,
		),
	)
	if len(diags) != 0 {
		t.Fatalf("expected return constraint, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected return context to infer the local lambda signature")
	}
	if got := formatTypeForReturnDiagnostic(inferred); got != "fn(i64) -> i64" {
		t.Fatalf("expected fn(i64) -> i64, got %s", got)
	}
}

func TestLocalLambdaExplicitDynamicUseRemainsUnconstrained(t *testing.T) {
	lambda := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		ast.Fn(
			"retain",
			[]*ast.FunctionParameter{ast.Param("value", ast.Ty("any"))},
			[]ast.Statement{ast.ID("value")},
			ast.Ty("any"),
			nil,
			nil,
			false,
			false,
		),
		ast.Assign(ast.ID("callback"), lambda),
		ast.Call("retain", ast.ID("callback")),
	)
	if len(diags) != 0 {
		t.Fatalf("expected explicit dynamic use to remain valid, got %v", diags)
	}
	inferred, ok := checker.infer.get(lambda)
	if !ok {
		t.Fatal("expected initial erased lambda inference")
	}
	if complete, ok := inferred.(FunctionType); !ok || completeMonomorphicFunctionType(complete) {
		t.Fatalf("expected unresolved callable signature, got %s", formatTypeForReturnDiagnostic(inferred))
	}
}

func TestLocalLambdaConstraintsRespectLexicalShadowing(t *testing.T) {
	outer := localIdentityLambda()
	inner := localIdentityLambda()
	checker, diags := checkLocalLambdaModule(
		t,
		callableConsumer("apply_i32", "i32"),
		callableConsumer("apply_i64", "i64"),
		ast.Assign(ast.ID("callback"), outer),
		ast.Block(
			ast.Assign(ast.ID("callback"), inner),
			ast.Call("apply_i64", ast.ID("callback")),
		),
		ast.Call("apply_i32", ast.ID("callback")),
	)
	if len(diags) != 0 {
		t.Fatalf("expected shadowed lambdas to constrain independently, got %v", diags)
	}
	if got, ok := checker.infer.get(outer); !ok ||
		formatTypeForReturnDiagnostic(got) != "fn(i32) -> i32" {
		t.Fatalf("expected outer i32 signature, got %v", got)
	}
	if got, ok := checker.infer.get(inner); !ok ||
		formatTypeForReturnDiagnostic(got) != "fn(i64) -> i64" {
		t.Fatalf("expected inner i64 signature, got %v", got)
	}
}

func TestLocalLambdaReassignmentClearsOriginalConstraintSource(t *testing.T) {
	_, diags := checkLocalLambdaModule(
		t,
		callableConsumer("apply_i32", "i32"),
		callableConsumer("apply_i64", "i64"),
		ast.Assign(ast.ID("callback"), localIdentityLambda()),
		ast.Call("apply_i32", ast.ID("callback")),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("callback"), localIdentityLambda()),
		ast.Call("apply_i64", ast.ID("callback")),
	)
	for _, diagnostic := range diags {
		if diagnostic.Code == DiagnosticCodeCallableSignatureMismatch &&
			strings.Contains(diagnostic.Message, "local lambda 'callback'") {
			t.Fatalf("reassignment must not constrain the original lambda: %v", diags)
		}
	}
}
