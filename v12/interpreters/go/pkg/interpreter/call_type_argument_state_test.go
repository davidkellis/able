package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestPopulateCallTypeArguments_RecomputesInferredTypeArgumentsForPolymorphicCallSite(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}); err != nil {
		t.Fatalf("populate i32 type args: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("first inferred type arg = %s, want i32", got)
	}
	firstVersion := interp.inferredCallTypeArgumentVersion(call)
	if firstVersion == 0 {
		t.Fatalf("expected inferred call type-argument version to be recorded")
	}

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.StringValue{Val: "hello"}}); err != nil {
		t.Fatalf("populate String type args: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "String" {
		t.Fatalf("second inferred type arg = %s, want String", got)
	}
	if got := interp.inferredCallTypeArgumentVersion(call); got <= firstVersion {
		t.Fatalf("expected inferred call type-argument version to advance, got %d after %d", got, firstVersion)
	}
}
