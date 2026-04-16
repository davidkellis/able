package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAnalyzeFrameLayoutCachesParamSimpleTypes(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("String")),
			ast.Param("c", nil),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramSimpleTypes), 3; got != want {
		t.Fatalf("unexpected param simple type count: got=%d want=%d", got, want)
	}
	if got := layout.paramSimpleTypes[0]; got != "i32" {
		t.Fatalf("unexpected first param simple type: got=%q want=%q", got, "i32")
	}
	if got := layout.paramSimpleTypes[1]; got != "String" {
		t.Fatalf("unexpected second param simple type: got=%q want=%q", got, "String")
	}
	if got := layout.paramSimpleTypes[2]; got != "" {
		t.Fatalf("unexpected third param simple type: got=%q want empty", got)
	}
}

func TestAnalyzeFrameLayoutCachesParamCoercionMetadata(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("c", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramTypes), 3; got != want {
		t.Fatalf("unexpected param type count: got=%d want=%d", got, want)
	}
	if got, want := len(layout.paramNeedsCoercion), 3; got != want {
		t.Fatalf("unexpected coercion metadata count: got=%d want=%d", got, want)
	}
	if !layout.paramNeedsCoercion[0] {
		t.Fatalf("expected concrete primitive param to retain coercion check metadata")
	}
	if layout.paramNeedsCoercion[1] {
		t.Fatalf("expected Array param to skip runtime coercion metadata")
	}
	if layout.paramNeedsCoercion[2] {
		t.Fatalf("expected generic param to skip runtime coercion metadata")
	}
	if !layout.anyParamCoercion {
		t.Fatalf("expected layout to record at least one coercion-bearing parameter")
	}
	if layout.anyExplicitCoercion {
		t.Fatalf("expected only the first parameter to require coercion in this layout")
	}
}

func TestAnalyzeFrameLayoutCachesNoCoercionSummaryFlags(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("b", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("void"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if layout.paramNeedsCoercion[0] || layout.paramNeedsCoercion[1] {
		t.Fatalf("expected both parameters to skip runtime coercion metadata")
	}
	if layout.anyParamCoercion {
		t.Fatalf("expected layout to record no coercion-bearing parameters")
	}
	if layout.anyExplicitCoercion {
		t.Fatalf("expected layout to record no explicit coercion-bearing parameters")
	}
}

func TestInlineCoercionUnnecessaryAcceptsBoxedPrimitivePointers(t *testing.T) {
	intVal := runtime.IntegerValue{TypeSuffix: runtime.IntegerI32}
	floatVal := runtime.FloatValue{TypeSuffix: runtime.FloatF64}
	stringVal := runtime.StringValue{Val: "x"}
	boolVal := runtime.BoolValue{Val: true}

	if !inlineCoercionUnnecessary(ast.Ty("i32"), &intVal) {
		t.Fatalf("expected boxed integer pointer to match i32")
	}
	if !inlineCoercionUnnecessary(ast.Ty("f64"), &floatVal) {
		t.Fatalf("expected boxed float pointer to match f64")
	}
	if !inlineCoercionUnnecessary(ast.Ty("String"), &stringVal) {
		t.Fatalf("expected boxed string pointer to match String")
	}
	if !inlineCoercionUnnecessary(ast.Ty("Bool"), &boolVal) {
		t.Fatalf("expected boxed bool pointer to match Bool")
	}
}

func TestInlineCoerceValueBySimpleTypeIntegerWidening(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	coerced, ok, err := inlineCoerceValueBySimpleType("i64", value)
	if err != nil {
		t.Fatalf("unexpected coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected integer widening to be handled")
	}
	intVal, ok := coerced.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected widened integer value, got %T", coerced)
	}
	if intVal.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("unexpected widened suffix: got=%s want=%s", intVal.TypeSuffix, runtime.IntegerI64)
	}
	if got, fits := intVal.ToInt64(); !fits || got != 7 {
		t.Fatalf("unexpected widened integer payload: got=%d fits=%v", got, fits)
	}
}

func TestInlineCoerceValueBySimpleTypeIntegerToFloat(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	coerced, ok, err := inlineCoerceValueBySimpleType("f64", value)
	if err != nil {
		t.Fatalf("unexpected coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected integer-to-float coercion to be handled")
	}
	floatVal, ok := coerced.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float value, got %T", coerced)
	}
	if floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
		t.Fatalf("unexpected float coercion result: got=%#v", floatVal)
	}
}
