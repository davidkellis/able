package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAnalyzeFrameLayoutResolvesMethodSetSelfParamMetadata(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"eq",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("other", ast.Ty("Self")),
		},
		[]ast.Statement{ast.Bool(true)},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	methodSet := &runtime.MethodSet{TargetType: ast.Ty("Bool")}

	layout := analyzeFrameLayoutWithEnvAndMethodSet(interp, def, interp.GlobalEnvironment(), methodSet)
	if layout == nil {
		t.Fatal("expected method frame layout")
	}
	for idx := range def.Params {
		if got := layout.paramSimpleTypes[idx]; got != "Bool" {
			t.Fatalf("param %d simple type = %q, want Bool", idx, got)
		}
		if got := layout.paramSimpleChecks[idx]; got != bytecodeSimpleTypeCheckBool {
			t.Fatalf("param %d simple check = %d, want Bool check", idx, got)
		}
		if !layout.paramNeedsCoercion[idx] {
			t.Fatalf("param %d should retain runtime coercion eligibility", idx)
		}
	}
}

func TestAnalyzeFrameLayoutResolvesMethodSetSelfToExactNamedStruct(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	keyDef := ast.StructDef(
		"Key",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(keyDef, env); err != nil {
		t.Fatalf("evaluate Key definition: %v", err)
	}
	def := ast.Fn(
		"eq",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("other", ast.Ty("Self")),
		},
		[]ast.Statement{ast.Bool(true)},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	methodSet := &runtime.MethodSet{TargetType: ast.Ty("Key")}

	layout := analyzeFrameLayoutWithEnvAndMethodSet(interp, def, env, methodSet)
	if layout == nil {
		t.Fatal("expected method frame layout")
	}
	for idx, exactDef := range layout.paramExactStructDef {
		if exactDef == nil || exactDef.Node != keyDef {
			t.Fatalf("param %d did not cache the Key definition", idx)
		}
	}
}

func TestAnalyzeFrameLayoutResolvedGenericSelfKeepsGenericSafety(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"same",
		[]*ast.FunctionParameter{ast.Param("other", ast.Ty("Self"))},
		[]ast.Statement{ast.Bool(true)},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	methodSet := &runtime.MethodSet{
		TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		GenericParams: []*ast.GenericParameter{ast.GenericParam("T", nil)},
	}

	layout := analyzeFrameLayoutWithEnvAndMethodSet(interp, def, nil, methodSet)
	if layout == nil {
		t.Fatal("expected generic method frame layout")
	}
	if !typeExpressionsEqual(layout.paramTypes[0], methodSet.TargetType) {
		t.Fatalf("resolved param type = %s, want %s", typeExpressionToString(layout.paramTypes[0]), typeExpressionToString(methodSet.TargetType))
	}
	if layout.paramNeedsCoercion[0] || layout.anyParamCoercion {
		t.Fatal("generic Self target should retain the existing generic coercion safety path")
	}
}

func TestBytecodeSimpleTypeCheckRecognizesCharValues(t *testing.T) {
	check := bytecodeSimpleTypeCheckForName("char")
	if check != bytecodeSimpleTypeCheckChar {
		t.Fatalf("char simple check = %d, want %d", check, bytecodeSimpleTypeCheckChar)
	}
	charValue := runtime.CharValue{Val: 'a'}
	if !inlineCoercionUnnecessaryBySimpleCheck(check, charValue) {
		t.Fatal("char value should satisfy the cached char check")
	}
	if !inlineCoercionUnnecessaryBySimpleCheck(check, &charValue) {
		t.Fatal("char pointer should satisfy the cached char check")
	}
	if inlineCoercionUnnecessaryBySimpleCheck(check, runtime.StringValue{Val: "a"}) {
		t.Fatal("string value should not satisfy the cached char check")
	}
}
