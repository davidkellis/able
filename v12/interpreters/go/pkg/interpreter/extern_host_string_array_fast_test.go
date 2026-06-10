package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func testStringArrayValue(values ...string) runtime.Value {
	elements := make([]runtime.Value, len(values))
	for idx, value := range values {
		elements[idx] = runtime.StringValue{Val: value}
	}
	return &runtime.ArrayValue{Elements: elements}
}

func TestBuildExternFastInvokerForStringStringArrayStringUnionF64(t *testing.T) {
	def := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"means_like",
			[]*ast.FunctionParameter{
				ast.Param("text", ast.Ty("String")),
				ast.Param("array_key", ast.Ty("String")),
				ast.Param("field_names", ast.Gen(ast.Ty("Array"), ast.Ty("String"))),
			},
			nil,
			&ast.UnionTypeExpression{
				Members: []ast.TypeExpression{
					ast.Ty("JsonError"),
					ast.Gen(ast.Ty("Array"), ast.Ty("f64")),
				},
			},
			nil,
			nil,
			false,
			false,
		),
		`return nil`,
	)

	invoker := buildExternFastInvoker(def, func(text string, key string, fields []string) interface{} {
		if text != "json" || key != "coordinates" || len(fields) != 3 || fields[0] != "x" || fields[1] != "y" || fields[2] != "z" {
			return map[string]any{"message": "bad args"}
		}
		return []float64{1.5, 2.5, 3.5}
	})
	if invoker == nil {
		t.Fatalf("expected String,String,Array String fast invoker")
	}

	got, err := invoker(New(), []runtime.Value{
		runtime.StringValue{Val: "json"},
		runtime.StringValue{Val: "coordinates"},
		testStringArrayValue("x", "y", "z"),
	})
	if err != nil {
		t.Fatalf("invoke fast invoker: %v", err)
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 means, got %d", len(arr.Elements))
	}
	for idx, want := range []float64{1.5, 2.5, 3.5} {
		floatVal, ok := arr.Elements[idx].(runtime.FloatValue)
		if !ok {
			t.Fatalf("mean %d type = %T", idx, arr.Elements[idx])
		}
		if floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != want {
			t.Fatalf("mean %d = %#v, want %f", idx, floatVal, want)
		}
	}
}

func TestExternModuleBuildsFastInvokerForUnionStringSignature(t *testing.T) {
	interp := New()
	sig := ast.Fn(
		"read_text_like",
		[]*ast.FunctionParameter{ast.Param("path", ast.Ty("String"))},
		nil,
		&ast.UnionTypeExpression{
			Members: []ast.TypeExpression{
				ast.Ty("IOError"),
				ast.Ty("String"),
			},
		},
		nil,
		nil,
		false,
		false,
	)
	def := ast.Extern(ast.HostTargetGo, sig, `return path + ".txt"`)

	invoker := buildExternFastInvoker(def, func(path string) interface{} { return path + ".txt" })
	if invoker == nil {
		t.Fatalf("expected fast invoker for union String signature")
	}
	got, err := invoker(interp, []runtime.Value{runtime.StringValue{Val: "sample"}})
	if err != nil {
		t.Fatalf("invoke union String fast invoker: %v", err)
	}
	str, ok := got.(runtime.StringValue)
	if !ok || str.Val != "sample.txt" {
		t.Fatalf("unexpected union String result %#v", got)
	}
}
