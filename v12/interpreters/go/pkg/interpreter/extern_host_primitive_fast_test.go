package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBuildExternFastInvokerForPrimitiveIntegerSignature(t *testing.T) {
	interp := New()
	sig := ast.Fn(
		"combine_like",
		[]*ast.FunctionParameter{
			ast.Param("dst", ast.Ty("i32")),
			ast.Param("lhs", ast.Ty("i32")),
			ast.Param("rhs", ast.Ty("i64")),
		},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	def := ast.Extern(ast.HostTargetGo, sig, `return dst + lhs + int32(rhs)`)
	invoker := buildExternFastInvoker(def, func(dst, lhs int32, rhs int64) int32 {
		return dst + lhs + int32(rhs)
	})
	if invoker == nil {
		t.Fatalf("expected fast invoker for primitive integer signature")
	}

	got, err := invoker(interp, []runtime.Value{
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
		runtime.NewSmallInt(4, runtime.IntegerI64),
	})
	if err != nil {
		t.Fatalf("run primitive invoker: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", got)
	}
	if !intVal.IsSmall() || intVal.TypeSuffix != runtime.IntegerI32 || intVal.Int64Fast() != 9 {
		t.Fatalf("unexpected primitive invoker result %#v", intVal)
	}
}

func TestBuildExternFastInvokerForPrimitiveIntegerReturnKinds(t *testing.T) {
	toI64 := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"to_i64_like",
			[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
			nil,
			ast.Ty("i64"),
			nil,
			nil,
			false,
			false,
		),
		`return int64(value) * 2`,
	)
	toString := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"to_string_like",
			[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`return ""`,
	)

	i64Invoker := buildExternFastInvoker(toI64, func(value int32) int64 { return int64(value) * 2 })
	if i64Invoker == nil {
		t.Fatalf("expected i32->i64 fast invoker")
	}
	gotI64, err := i64Invoker(New(), []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)})
	if err != nil {
		t.Fatalf("invoke i32->i64 fast invoker: %v", err)
	}
	intVal, ok := gotI64.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", gotI64)
	}
	if !intVal.IsSmall() || intVal.TypeSuffix != runtime.IntegerI64 || intVal.Int64Fast() != 14 {
		t.Fatalf("unexpected i64 result %#v", intVal)
	}

	stringInvoker := buildExternFastInvoker(toString, func(value int32) string { return "n=" + string(rune('0'+value)) })
	if stringInvoker == nil {
		t.Fatalf("expected i32->String fast invoker")
	}
	gotString, err := stringInvoker(New(), []runtime.Value{runtime.NewSmallInt(5, runtime.IntegerI32)})
	if err != nil {
		t.Fatalf("invoke i32->String fast invoker: %v", err)
	}
	str, ok := gotString.(runtime.StringValue)
	if !ok || str.Val != "n=5" {
		t.Fatalf("unexpected string result %#v", gotString)
	}
}
