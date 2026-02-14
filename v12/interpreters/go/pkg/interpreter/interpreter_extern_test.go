package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestExternHandlersRegisterNativeFunctions(t *testing.T) {
	interp := New()
	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, ast.Pkg([]interface{}{"host"}, false))

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if _, ok := val.(*runtime.NativeFunctionValue); !ok {
		t.Fatalf("expected native function, got %#v", val)
	}
	bucket := interp.packageRegistry["host"]
	if bucket == nil || bucket["now_nanos"] == nil {
		t.Fatalf("package registry missing now_nanos")
	}
}

func TestExternIgnoresNonGoTargets(t *testing.T) {
	interp := New()
	sig := ast.Fn("ts_only", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetTypeScript, sig, "")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	if _, err := env.Get("ts_only"); err == nil {
		t.Fatalf("expected ts_only to be undefined for go target")
	}
}

func TestExternPreservesExistingBinding(t *testing.T) {
	interp := New()
	existing := &runtime.NativeFunctionValue{Name: "now_nanos", Arity: 0}
	interp.global.Define("now_nanos", existing)

	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if val != existing {
		t.Fatalf("expected existing binding to remain")
	}
}

func TestExternStructArrayFieldCoercesIntoHostMap(t *testing.T) {
	interp := New()

	specDef := ast.StructDef("Spec", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Gen(ast.Ty("Array"), ast.Ty("String")), "args"),
	}, ast.StructKindNamed, nil, nil, false)
	sig := ast.Fn(
		"accept_spec",
		[]*ast.FunctionParameter{ast.Param("spec", ast.Ty("Spec"))},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	mod := ast.Mod([]ast.Statement{
		specDef,
		ast.Extern(ast.HostTargetGo, sig, `
specMap, ok := spec.(map[string]any)
if !ok {
	return int32(-1)
}
args, ok := specMap["args"].([]any)
if !ok {
	return int32(-2)
}
return int32(len(args))
`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("accept_spec",
		ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Arr(ast.Str("a"), ast.Str("b")), "args"),
		}, false, "Spec", nil, nil),
	), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}

	iv, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if iv.Val.Int64() != 2 {
		t.Fatalf("expected extern to receive two args, got %s", iv.Val.String())
	}
}

func TestExternStructNullableArrayFieldCoercesIntoHostMap(t *testing.T) {
	interp := New()

	specDef := ast.StructDef("Spec", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Nullable(ast.Gen(ast.Ty("Array"), ast.Ty("String"))), "args"),
	}, ast.StructKindNamed, nil, nil, false)
	sig := ast.Fn(
		"accept_spec",
		[]*ast.FunctionParameter{ast.Param("spec", ast.Ty("Spec"))},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	mod := ast.Mod([]ast.Statement{
		specDef,
		ast.Extern(ast.HostTargetGo, sig, `
specMap, ok := spec.(map[string]any)
if !ok {
	return int32(-1)
}
raw, ok := specMap["args"]
if !ok {
	return int32(-2)
}
if raw == nil {
	return int32(-3)
}
args, ok := raw.([]any)
if !ok {
	return int32(-4)
}
return int32(len(args))
`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("accept_spec",
		ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Arr(ast.Str("x"), ast.Str("y"), ast.Str("z")), "args"),
		}, false, "Spec", nil, nil),
	), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}

	iv, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if iv.Val.Int64() != 3 {
		t.Fatalf("expected extern to receive three args, got %s", iv.Val.String())
	}
}
