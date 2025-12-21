package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestAliasMethodsPropagate(t *testing.T) {
	bagAlias := ast.NewTypeAliasDefinition(
		ast.ID("Bag"),
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		true,
	)
	strListAlias := ast.NewTypeAliasDefinition(
		ast.ID("StrList"),
		ast.Gen(ast.Ty("Array"), ast.Ty("String")),
		nil,
		nil,
		true,
	)

	headFn := ast.Fn(
		"head",
		[]*ast.FunctionParameter{ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Self"))},
		[]ast.Statement{
			ast.Ret(ast.Index(ast.ID("self"), ast.Int(0))),
		},
		ast.Nullable(ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Gen(ast.Ty("Bag"), ast.Ty("T")),
		[]*ast.FunctionDefinition{headFn},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
	)

	displayFn := ast.Fn(
		"to_string",
		[]*ast.FunctionParameter{ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Self"))},
		[]ast.Statement{ast.Ret(ast.Str("<strlist>"))},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	displayImpl := ast.Impl(ast.ID("Display"), ast.Ty("StrList"), []*ast.FunctionDefinition{displayFn}, nil, nil, nil, nil, false)

	arrAssign := ast.Assign(ast.ID("arr"), ast.Arr(ast.Str("left"), ast.Str("right")))
	headAssign := ast.Assign(ast.ID("headVal"), ast.CallExpr(ast.Member(ast.ID("arr"), ast.ID("head"))))
	strAssign := ast.Assign(ast.ID("strVal"), ast.CallExpr(ast.Member(ast.ID("arr"), ast.ID("to_string"))))

	module := ast.NewModule([]ast.Statement{
		bagAlias,
		strListAlias,
		methods,
		displayImpl,
		arrAssign,
		headAssign,
		strAssign,
	}, nil, nil)

	i := New()
	_, env, err := i.EvaluateModule(module)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	headVal, err := env.Get("headVal")
	if err != nil {
		t.Fatalf("headVal missing: %v", err)
	}
	headStr, ok := headVal.(runtime.StringValue)
	if !ok || headStr.Val != "left" {
		t.Fatalf("expected headVal to be 'left', got %#v", headVal)
	}

	strVal, err := env.Get("strVal")
	if err != nil {
		t.Fatalf("strVal missing: %v", err)
	}
	strOut, ok := strVal.(runtime.StringValue)
	if !ok || strOut.Val != "<strlist>" {
		t.Fatalf("expected strVal to be '<strlist>', got %#v", strVal)
	}
}

func TestAliasMethodsUseCanonicalTypeName(t *testing.T) {
	interp := New()
	env := runtime.NewEnvironment(nil)

	widget := ast.NewStructDefinition(
		ast.ID("Widget"),
		[]*ast.StructFieldDefinition{ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("value"))},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(widget, env); err != nil {
		t.Fatalf("evaluate struct: %v", err)
	}
	orig, err := env.Get("Widget")
	if err != nil {
		t.Fatalf("widget binding missing: %v", err)
	}
	env.Define("AliasWidget", orig)

	describeFn := ast.Fn(
		"describe",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
		[]ast.Statement{ast.Ret(ast.Str("ok"))},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	methods := ast.Methods(ast.Ty("AliasWidget"), []*ast.FunctionDefinition{describeFn}, nil, nil)
	if _, err := interp.evaluateMethodsDefinition(methods, env); err != nil {
		t.Fatalf("evaluate methods: %v", err)
	}

	bucket, ok := interp.inherentMethods["Widget"]
	if !ok {
		t.Fatalf("expected methods bucket for Widget")
	}
	if _, exists := bucket["describe"]; !exists {
		t.Fatalf("expected describe to register under canonical type name, got %#v", bucket)
	}
}
