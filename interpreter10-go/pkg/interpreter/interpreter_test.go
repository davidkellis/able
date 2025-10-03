package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func keysOf(m map[string]runtime.Value) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestEvaluateStringLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{ast.Str("hello")}, nil, nil)
	val, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateIdentifierLookup(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	global.Define("greeting", runtime.StringValue{Val: "hello"})

	val, err := interp.evaluateExpression(ast.ID("greeting"), global)
	if err != nil {
		t.Fatalf("identifier lookup failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateBlockCreatesScope(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Str("inner")),
		ast.ID("x"),
	)

	val, err := interp.evaluateExpression(block, global)
	if err != nil {
		t.Fatalf("block evaluation failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "inner" {
		t.Fatalf("unexpected block result %#v", val)
	}

	if _, err := global.Get("x"); err == nil {
		t.Fatalf("expected inner binding to stay scoped")
	}
}

func TestEvaluateBinaryAddition(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Int(1)),
		ast.Assign(ast.ID("b"), ast.Int(2)),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	iv, ok := result.(runtime.IntegerValue)
	if !ok || iv.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected integer 3, got %#v", result)
	}
}

func TestStringInterpolationEvaluatesParts(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(2)),
		ast.Interp(
			ast.Str("x = "),
			ast.ID("x"),
			ast.Str(", sum = "),
			ast.Bin("+", ast.Int(3), ast.Int(4)),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %#v", result)
	}
	if str.Val != "x = 2, sum = 7" {
		t.Fatalf("unexpected interpolation output: %q", str.Val)
	}
}

func TestStringInterpolationUsesToStringMethod(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Interp(
							ast.Str("Point("),
							ast.Member(ast.ID("self"), "x"),
							ast.Str(","),
							ast.Member(ast.ID("self"), "y"),
							ast.Str(")"),
						)),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.Interp(
			ast.Str("P= "),
			ast.ID("p"),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %#v", result)
	}
	if str.Val != "P= Point(1,2)" {
		t.Fatalf("unexpected interpolation output: %q", str.Val)
	}
}

func TestMatchExpressionWithIdentifierAndLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(2),
			ast.Mc(ast.LitP(ast.Int(1)), ast.Int(10)),
			ast.Mc(ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(5))),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", result)
	}
}

func TestMatchExpressionStructGuard(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Match(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
			ast.Mc(
				ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("a"), "x", nil),
					ast.FieldP(ast.ID("b"), "y", nil),
				}, false, "Point"),
				ast.Bin("+", ast.ID("a"), ast.ID("b")),
				ast.Bin(">", ast.ID("b"), ast.ID("a")),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected integer 3, got %#v", result)
	}
}

func TestOrElseExpressionHandlesRaise(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.OrElse(
			ast.Prop(ast.Block(ast.Raise(ast.Str("x")))),
			"e",
			ast.Str("handled"),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", result)
	}
}

func TestRescueExpressionTypedPattern(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Rescue(
			ast.Block(
				ast.Raise(ast.Str("boom")),
			),
			ast.Mc(
				ast.TypedP(ast.ID("err"), ast.Ty("Error")),
				ast.Str("caught"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "caught" {
		t.Fatalf("expected 'caught', got %#v", result)
	}
}

func TestEnsureExpressionRunsFinally(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("flag"), ast.Str("")),
		ast.Assign(
			ast.ID("value"),
			ast.Ensure(
				ast.Rescue(
					ast.Block(ast.Raise(ast.Str("oops"))),
					ast.Mc(ast.Wc(), ast.Str("rescued")),
				),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Str("done")),
			),
		),
		ast.ID("value"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := result.(runtime.StringValue); !ok || str.Val != "rescued" {
		t.Fatalf("expected ensure to return 'rescued', got %#v", result)
	}
	rescued, err := env.Get("flag")
	if err != nil {
		t.Fatalf("expected binding for flag: %v", err)
	}
	if str, ok := rescued.(runtime.StringValue); !ok || str.Val != "done" {
		t.Fatalf("expected final flag value 'done', got %#v", rescued)
	}
}

func TestRethrowStatementBubblesToOuterRescue(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("result"),
			ast.Rescue(
				ast.Rescue(
					ast.Block(ast.Raise(ast.Str("oops"))),
					ast.Mc(ast.Wc(), ast.Block(ast.Rethrow())),
				),
				ast.Mc(ast.Wc(), ast.Str("handled")),
			),
		),
		ast.ID("result"),
	}, nil, nil)

	value, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := value.(runtime.StringValue); !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", value)
	}
}

func TestErrorPayloadDestructuring(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Rescue(
			ast.Block(ast.Raise(ast.Str("fail"))),
			ast.Mc(
				ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("value"), "value", nil),
				}, false, nil),
				ast.Str("handled"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := result.(runtime.StringValue); !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", result)
	}
}

func TestPackageWildcardImport(t *testing.T) {
	interp := New()
	modulePkg := ast.Mod([]ast.Statement{
		ast.Fn(
			"hello",
			[]*ast.FunctionParameter{},
			[]ast.Statement{ast.Ret(ast.Str("hi"))},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Fn(
			"secret",
			[]*ast.FunctionParameter{},
			[]ast.Statement{ast.Ret(ast.Str("nope"))},
			nil,
			nil,
			nil,
			false,
			true,
		),
		ast.StructDef("Thing", nil, ast.StructKindNamed, nil, nil, false),
		ast.StructDef("Hidden", nil, ast.StructKindNamed, nil, nil, true),
		ast.Iface("PublicI", nil, nil, nil, nil, nil, false),
		ast.Iface("HiddenI", nil, nil, nil, nil, nil, true),
	}, nil, ast.Pkg([]interface{}{"example"}, false))

	if _, _, err := interp.EvaluateModule(modulePkg); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}
	if bucket, ok := interp.packageRegistry["example"]; ok {
		if _, ok := bucket["Thing"]; !ok {
			t.Fatalf("expected package registry to include struct 'Thing'")
		}
	}

	moduleUse := ast.Mod(
		[]ast.Statement{ast.Call("hello")},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"example"}, true, nil, nil)},
		nil,
	)

	result, env, err := interp.EvaluateModule(moduleUse)
	if err != nil {
		t.Fatalf("consumer module evaluation failed: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "hi" {
		t.Fatalf("expected 'hi', got %#v", result)
	}
	if _, err := env.Get("secret"); err == nil {
		t.Fatalf("expected private symbol to remain unavailable")
	}
}

func TestPackageAliasFiltersPrivate(t *testing.T) {
	interp := New()
	modulePkg := ast.Mod([]ast.Statement{
		ast.Fn(
			"hello",
			[]*ast.FunctionParameter{},
			[]ast.Statement{ast.Ret(ast.Str("hi"))},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Fn(
			"secret",
			[]*ast.FunctionParameter{},
			[]ast.Statement{ast.Ret(ast.Str("nope"))},
			nil,
			nil,
			nil,
			false,
			true,
		),
		ast.StructDef("Thing", nil, ast.StructKindNamed, nil, nil, false),
		ast.StructDef("Hidden", nil, ast.StructKindNamed, nil, nil, true),
		ast.Iface("PublicI", nil, nil, nil, nil, nil, false),
		ast.Iface("HiddenI", nil, nil, nil, nil, nil, true),
	}, nil, ast.Pkg([]interface{}{"example"}, false))

	if _, _, err := interp.EvaluateModule(modulePkg); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	moduleAlias := ast.Mod(
		[]ast.Statement{
			ast.Assign(ast.ID("result"), ast.CallExpr(ast.Member(ast.ID("pkg"), "hello"))),
			ast.ID("result"),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"example"}, false, nil, "pkg")},
		nil,
	)

	value, env, err := interp.EvaluateModule(moduleAlias)
	if err != nil {
		t.Fatalf("alias module evaluation failed: %v", err)
	}
	strResult, ok := value.(runtime.StringValue)
	if !ok || strResult.Val != "hi" {
		t.Fatalf("expected alias call to return 'hi', got %#v", value)
	}
	pkgValRaw, err := env.Get("pkg")
	if err != nil {
		t.Fatalf("expected package binding: %v", err)
	}
	pkgVal, ok := pkgValRaw.(runtime.PackageValue)
	if !ok {
		t.Fatalf("expected PackageValue, got %#v", pkgValRaw)
	}
	if _, ok := pkgVal.Public["hello"]; !ok {
		t.Fatalf("expected public symbol 'hello' in package")
	}
	if _, ok := pkgVal.Public["secret"]; ok {
		t.Fatalf("did not expect private symbol 'secret' to be exported")
	}
	if _, ok := pkgVal.Public["Thing"]; !ok {
		t.Fatalf("expected public struct 'Thing' to be exported; got keys %#v", keysOf(pkgVal.Public))
	}

	if _, err := interp.evaluateExpression(ast.Member(ast.ID("pkg"), "secret"), env); err == nil {
		t.Fatalf("expected private function access via alias to fail")
	}
	if _, err := interp.evaluateExpression(ast.Member(ast.ID("pkg"), "Hidden"), env); err == nil {
		t.Fatalf("expected private struct access via alias to fail")
	}

	if _, err := interp.evaluateExpression(ast.Member(ast.ID("pkg"), "HiddenI"), env); err == nil {
		t.Fatalf("expected private interface access via alias to fail")
	}

	if val, err := interp.evaluateExpression(ast.Member(ast.ID("pkg"), "Thing"), env); err != nil {
		t.Fatalf("expected public struct access via alias to succeed: %v", err)
	} else if _, ok := val.(*runtime.StructDefinitionValue); !ok {
		t.Fatalf("expected struct definition through alias, got %#v", val)
	}
}

func TestImportingPrivateFunctionFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Fn(
			"secret",
			[]*ast.FunctionParameter{},
			[]ast.Statement{ast.Ret(ast.Int(1))},
			nil,
			nil,
			nil,
			false,
			true,
		),
	}, nil, ast.Pkg([]interface{}{"core"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	importModule := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"core"}, false, []*ast.ImportSelector{ast.ImpSel("secret", "alias")}, nil)},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected private import to fail")
	} else {
		expected := "Import error: function 'secret' is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}
}

func TestImportSelectorAliasBindsFunction(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Fn(
			"foo",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(42))},
			nil,
			nil,
			nil,
			false,
			false,
		),
	}, nil, ast.Pkg([]interface{}{"core"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	selectors := []*ast.ImportSelector{ast.ImpSel("foo", "bar")}
	consumer := ast.Mod(
		[]ast.Statement{
			ast.Assign(ast.ID("result"), ast.Call("bar")),
			ast.ID("result"),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"core"}, false, selectors, nil)},
		nil,
	)

	value, env, err := interp.EvaluateModule(consumer)
	if err != nil {
		t.Fatalf("consumer module evaluation failed: %v", err)
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected 42 from alias call, got %#v", value)
	}

	aliasBinding, err := env.Get("bar")
	if err != nil {
		t.Fatalf("expected alias binding for 'bar': %v", err)
	}
	if aliasBinding.Kind() != runtime.KindFunction {
		t.Fatalf("expected alias to bind a function, got %#v", aliasBinding)
	}

	if _, err := env.Get("foo"); err == nil {
		t.Fatalf("original symbol 'foo' should not be bound after alias import")
	}

	resultBinding, err := env.Get("result")
	if err != nil {
		t.Fatalf("expected module binding for 'result': %v", err)
	}
	resultVal, ok := resultBinding.(runtime.IntegerValue)
	if !ok || resultVal.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected result binding 42, got %#v", resultBinding)
	}
}

func TestImportingPrivateStructFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.StructDef("Hidden", nil, ast.StructKindNamed, nil, nil, true),
		ast.StructDef("Public", nil, ast.StructKindNamed, nil, nil, false),
	}, nil, ast.Pkg([]interface{}{"pkg"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	privateImport := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"pkg"}, false, []*ast.ImportSelector{ast.ImpSel("Hidden", nil)}, nil)},
		nil,
	)

	if _, _, err := interp.EvaluateModule(privateImport); err == nil {
		t.Fatalf("expected private struct import to fail")
	} else {
		expected := "Import error: struct 'Hidden' is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	publicImport := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"pkg"}, false, []*ast.ImportSelector{ast.ImpSel("Public", "Pub")}, nil)},
		nil,
	)

	_, env, err := interp.EvaluateModule(publicImport)
	if err != nil {
		t.Fatalf("public struct import should succeed: %v", err)
	}
	val, err := env.Get("Pub")
	if err != nil {
		t.Fatalf("expected alias binding: %v", err)
	}
	if _, ok := val.(*runtime.StructDefinitionValue); !ok {
		t.Fatalf("expected StructDefinitionValue, got %#v", val)
	}
}

func TestPrivateStaticMethodNotAccessible(t *testing.T) {
	interp := New()
	defModule := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden_static",
					nil,
					[]ast.Statement{
						ast.Ret(
							ast.StructLit(
								[]*ast.StructFieldInitializer{
									ast.FieldInit(ast.Int(0), "x"),
									ast.FieldInit(ast.Int(0), "y"),
								},
								false,
								"Point",
								nil,
								nil,
							),
						),
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
				ast.Fn(
					"origin",
					nil,
					[]ast.Statement{
						ast.Ret(
							ast.StructLit(
								[]*ast.StructFieldInitializer{
									ast.FieldInit(ast.Int(0), "x"),
									ast.FieldInit(ast.Int(0), "y"),
								},
								false,
								"Point",
								nil,
								nil,
							),
						),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(defModule); err != nil {
		t.Fatalf("setup module evaluation failed: %v", err)
	}

	callHidden := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("Point"), "hidden_static")),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private static method call to fail")
	} else {
		expected := "Method 'hidden_static' on Point is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	callPublic := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("Point"), "origin")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(callPublic)
	if err != nil {
		t.Fatalf("public static method call failed: %v", err)
	}
	if _, ok := result.(*runtime.StructInstanceValue); !ok {
		t.Fatalf("expected struct instance result, got %#v", result)
	}
}

func TestPrivateInstanceMethodNotAccessible(t *testing.T) {
	interp := New()
	defModule := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Int(1))},
					nil,
					nil,
					nil,
					false,
					true,
				),
				ast.Fn(
					"get",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Member(ast.ID("self"), "value")),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(defModule); err != nil {
		t.Fatalf("setup module evaluation failed: %v", err)
	}

	instanceModule := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("counter"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(5), "value")},
				false,
				"Counter",
				nil,
				nil,
			),
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(instanceModule); err != nil {
		t.Fatalf("instance setup failed: %v", err)
	}

	callHidden := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("counter"), "hidden")),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private instance method call to fail")
	} else {
		expected := "Method 'hidden' on Counter is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	callPublic := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("counter"), "get")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(callPublic)
	if err != nil {
		t.Fatalf("public instance method call failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok || intResult.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected 5 from get(), got %#v", result)
	}
}

func TestImportingPrivateInterfaceFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Iface("HiddenI", nil, nil, nil, nil, nil, true),
		ast.Iface("PublicI", nil, nil, nil, nil, nil, false),
	}, nil, ast.Pkg([]interface{}{"pkg"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	privateImport := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"pkg"}, false, []*ast.ImportSelector{ast.ImpSel("HiddenI", nil)}, nil)},
		nil,
	)

	if _, _, err := interp.EvaluateModule(privateImport); err == nil {
		t.Fatalf("expected private interface import to fail")
	} else {
		expected := "Import error: interface 'HiddenI' is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	publicImport := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"pkg"}, false, []*ast.ImportSelector{ast.ImpSel("PublicI", "PI")}, nil)},
		nil,
	)

	_, env, err := interp.EvaluateModule(publicImport)
	if err != nil {
		t.Fatalf("public interface import should succeed: %v", err)
	}
	val, err := env.Get("PI")
	if err != nil {
		t.Fatalf("expected alias binding: %v", err)
	}
	if _, ok := val.(*runtime.InterfaceDefinitionValue); !ok {
		t.Fatalf("expected InterfaceDefinitionValue, got %#v", val)
	}
}

func TestDynImportSelectorsAndAlias(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(11))},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Fn(
			"hidden",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(1))},
			nil,
			nil,
			nil,
			false,
			true,
		),
	}, nil, ast.Pkg([]interface{}{"dynp"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	selectorsModule := ast.Mod([]ast.Statement{
		ast.DynImp([]interface{}{"dynp"}, false, []*ast.ImportSelector{ast.ImpSel("f", "ff")}, nil),
		ast.Assign(ast.ID("x"), ast.Call("ff")),
	}, nil, nil)

	value, _, err := interp.EvaluateModule(selectorsModule)
	if err != nil {
		t.Fatalf("dyn import selectors failed: %v", err)
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(11)) != 0 {
		t.Fatalf("expected 11 from dyn selector, got %#v", value)
	}

	aliasModule := ast.Mod([]ast.Statement{
		ast.DynImp([]interface{}{"dynp"}, false, nil, "D"),
		ast.Assign(ast.ID("y"), ast.CallExpr(ast.Member(ast.ID("D"), "f"))),
	}, nil, nil)

	aliasResult, env, err := interp.EvaluateModule(aliasModule)
	if err != nil {
		t.Fatalf("dyn import alias failed: %v", err)
	}
	aliasInt, ok := aliasResult.(runtime.IntegerValue)
	if !ok || aliasInt.Val.Cmp(bigInt(11)) != 0 {
		t.Fatalf("expected 11 from dyn alias, got %#v", aliasResult)
	}
	aliasPkg, err := env.Get("D")
	if err != nil {
		t.Fatalf("expected dyn package alias binding: %v", err)
	}
	dynPkg, ok := aliasPkg.(runtime.DynPackageValue)
	if !ok {
		t.Fatalf("expected DynPackageValue for alias, got %#v", aliasPkg)
	}
	if dynPkg.Name != "dynp" {
		t.Fatalf("expected dyn package name 'dynp', got %q", dynPkg.Name)
	}
	if len(dynPkg.NamePath) != 1 || dynPkg.NamePath[0] != "dynp" {
		t.Fatalf("expected name path ['dynp'], got %#v", dynPkg.NamePath)
	}
}

func TestDynImportWildcardSkipsPrivate(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(11))},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Fn(
			"hidden",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(1))},
			nil,
			nil,
			nil,
			false,
			true,
		),
	}, nil, ast.Pkg([]interface{}{"dynp"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	wildcardModule := ast.Mod([]ast.Statement{
		ast.DynImp([]interface{}{"dynp"}, true, nil, nil),
		ast.Assign(ast.ID("z"), ast.Call("f")),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(wildcardModule)
	if err != nil {
		t.Fatalf("dyn wildcard import failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(11)) != 0 {
		t.Fatalf("expected 11 from dyn wildcard, got %#v", result)
	}
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	if _, err := env.Get("hidden"); err == nil {
		t.Fatalf("expected hidden not to be imported via wildcard")
	}

	moduleCheck := ast.Mod([]ast.Statement{ast.ID("hidden")}, nil, nil)
	if _, _, err := interp.EvaluateModule(moduleCheck); err == nil {
		t.Fatalf("expected evaluating hidden after dyn wildcard to fail")
	}
}

func TestDynImportPrivateSelectorFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod([]ast.Statement{
		ast.Fn(
			"hidden",
			nil,
			[]ast.Statement{ast.Ret(ast.Int(1))},
			nil,
			nil,
			nil,
			false,
			true,
		),
	}, nil, ast.Pkg([]interface{}{"dynp"}, false))

	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package evaluation failed: %v", err)
	}

	privateSelector := ast.Mod([]ast.Statement{
		ast.DynImp([]interface{}{"dynp"}, false, []*ast.ImportSelector{ast.ImpSel("hidden", nil)}, nil),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(privateSelector); err == nil {
		t.Fatalf("expected dynimport of private symbol to fail")
	} else {
		expected := "dynimport error: function 'hidden' is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}
}

func TestDestructuringAssignmentArrayPattern(t *testing.T) {
	interp := New()
	patternWithRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, ast.PatternFrom("rest"))
	patternNoRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))),
		ast.Assign(patternWithRest, ast.ID("arr")),
		ast.AssignOp(ast.AssignmentAssign, patternNoRest, ast.Arr(ast.Int(4), ast.Int(5))),
		ast.ID("rest"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first, err := env.Get("first")
	if err != nil {
		t.Fatalf("expected binding for first: %v", err)
	}
	firstInt, ok := first.(runtime.IntegerValue)
	if !ok || firstInt.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected first == 4, got %#v", first)
	}
	second, err := env.Get("second")
	if err != nil {
		t.Fatalf("expected binding for second: %v", err)
	}
	secondInt, ok := second.(runtime.IntegerValue)
	if !ok || secondInt.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected second == 5, got %#v", second)
	}
	if _, err := env.Get("rest"); err != nil {
		t.Fatalf("expected binding for rest: %v", err)
	}
	restVal, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected rest array, got %#v", result)
	}
	if len(restVal.Elements) != 1 {
		t.Fatalf("expected rest length 1, got %d", len(restVal.Elements))
	}
	if restElem, ok := restVal.Elements[0].(runtime.IntegerValue); !ok || restElem.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected rest element 3, got %#v", restVal.Elements[0])
	}
}

func TestForLoopArrayPattern(t *testing.T) {
	interp := New()
	pattern := ast.ArrP([]ast.Pattern{ast.PatternFrom("x"), ast.PatternFrom("y")}, nil)
	pairs := ast.Arr(
		ast.Arr(ast.Int(1), ast.Int(2)),
		ast.Arr(ast.Int(3), ast.Int(4)),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("pairs"), pairs),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(pattern, ast.ID("pairs"), ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
		)),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum, ok := result.(runtime.IntegerValue)
	if !ok || sum.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected sum 4, got %#v", result)
	}
}

func TestForLoopContinueSkipsElements(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.Iff(
					ast.Bin("==", ast.ID("x"), ast.Int(2)),
					ast.Block(ast.Cont(nil)),
				),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected 4 from continue loop, got %#v", result)
	}
}

func TestBreakpointLabeledBreak(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.Breakpoint(
			"exit",
			ast.Block(
				ast.ForLoopPattern(
					ast.ID("n"),
					ast.Range(ast.Int(1), ast.Int(5), true),
					ast.Block(
						ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("n"))),
						ast.Iff(
							ast.Bin("==", ast.ID("n"), ast.Int(3)),
							ast.Block(ast.Brk("exit", ast.Str("done"))),
						),
					),
				),
				ast.Str("fallthrough"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "done" {
		t.Fatalf("expected 'done', got %#v", result)
	}
}

func TestStructLiteralNamed(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "x"),
					ast.FieldInit(ast.Int(4), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.Member(ast.ID("p"), "x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected struct field x == 3, got %#v", result)
	}
	structVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected binding for p: %v", err)
	}
	instance, ok := structVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", structVal)
	}
	if instance.Fields == nil {
		t.Fatalf("expected named struct fields map")
	}
	if field, ok := instance.Fields["y"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected struct field y == 4, got %#v", instance.Fields["y"])
	}
}

func TestStructLiteralPositional(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Pair",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), nil),
				ast.FieldDef(ast.Ty("i32"), nil),
			},
			ast.StructKindPositional,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("pair"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(7), nil),
					ast.FieldInit(ast.Int(9), nil),
				},
				true,
				"Pair",
				nil,
				nil,
			),
		),
		ast.Member(ast.ID("pair"), ast.Int(1)),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected positional field 1 == 9, got %#v", result)
	}
}

func TestStructMemberAssignmentMutation(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(0), "x"),
					ast.FieldInit(ast.Int(0), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.AssignMember(ast.ID("p"), "x", ast.Int(5)),
		ast.Member(ast.ID("p"), "x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected updated field x == 5, got %#v", result)
	}
	structVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected struct binding for p: %v", err)
	}
	inst, ok := structVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", structVal)
	}
	if field, ok := inst.Fields["y"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("unexpected change to y field: %#v", inst.Fields["y"])
	}
}

func TestArrayIndexRead(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Arr(ast.Int(10), ast.Int(20), ast.Int(30))),
		ast.Index(ast.ID("a"), ast.Int(1)),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(20)) != 0 {
		t.Fatalf("expected index read 20, got %#v", result)
	}
}

func TestArrayIndexAssignment(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("a"), ast.Int(1)), ast.Int(9)),
		ast.Index(ast.ID("a"), ast.Int(1)),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected updated index value 9, got %#v", result)
	}
	arrVal, err := env.Get("a")
	if err != nil {
		t.Fatalf("expected array binding for 'a': %v", err)
	}
	arr, ok := arrVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array value, got %#v", arrVal)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected array length 2, got %d", len(arr.Elements))
	}
	if elem, ok := arr.Elements[1].(runtime.IntegerValue); !ok || elem.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected element 1 == 9, got %#v", arr.Elements[1])
	}
}

func TestCompoundAssignments(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(2)),
		ast.AssignOp(ast.AssignmentAdd, ast.ID("x"), ast.Int(3)),
		ast.AssignOp(ast.AssignmentShiftL, ast.ID("x"), ast.Int(1)),
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.AssignOp(ast.AssignmentAdd, ast.Member(ast.ID("p"), "x"), ast.Int(4)),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(3), ast.Int(4))),
		ast.AssignOp(ast.AssignmentMul, ast.Index(ast.ID("arr"), ast.Int(1)), ast.Int(2)),
		ast.ID("x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intVal, ok := result.(runtime.IntegerValue); !ok || intVal.Val.Cmp(bigInt(10)) != 0 {
		t.Fatalf("expected x == 10, got %#v", result)
	}
	pVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected struct binding for 'p': %v", err)
	}
	inst, ok := pVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", pVal)
	}
	if field, ok := inst.Fields["x"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected struct field x == 5, got %#v", inst.Fields["x"])
	}
	arrVal, err := env.Get("arr")
	if err != nil {
		t.Fatalf("expected array binding for 'arr': %v", err)
	}
	arr, ok := arrVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array value, got %#v", arrVal)
	}
	if elem, ok := arr.Elements[1].(runtime.IntegerValue); !ok || elem.Val.Cmp(bigInt(8)) != 0 {
		t.Fatalf("expected array element 1 == 8, got %#v", arr.Elements[1])
	}
}

func TestStructLiteralMissingFieldError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(3), "x"),
			},
			false,
			"Point",
			nil,
			nil,
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected missing field error")
	}
	if got := err.Error(); got != "Missing field 'y' for struct 'Point'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructLiteralPositionalArityError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Pair",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), nil),
				ast.FieldDef(ast.Ty("i32"), nil),
			},
			ast.StructKindPositional,
			nil,
			nil,
			false,
		),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(7), nil),
			},
			true,
			"Pair",
			nil,
			nil,
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected arity error")
	}
	if got := err.Error(); got != "Struct 'Pair' expects 2 fields, got 1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructFunctionalUpdate(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"User",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "id"),
				ast.FieldDef(ast.Ty("string"), "name"),
				ast.FieldDef(ast.Ty("bool"), "active"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("base"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "id"),
					ast.FieldInit(ast.Str("Alice"), "name"),
					ast.FieldInit(ast.Bool(true), "active"),
				},
				false,
				"User",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("updated"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("Bob"), "name"),
				},
				false,
				"User",
				ast.ID("base"),
				nil,
			),
		),
		ast.Member(ast.ID("updated"), "name"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nameVal, ok := result.(runtime.StringValue)
	if !ok || nameVal.Val != "Bob" {
		t.Fatalf("expected updated.name == Bob, got %#v", result)
	}
	baseVal, err := env.Get("base")
	if err != nil {
		t.Fatalf("expected binding for base: %v", err)
	}
	baseStruct, ok := baseVal.(*runtime.StructInstanceValue)
	if !ok || baseStruct.Fields == nil {
		t.Fatalf("expected named struct base, got %#v", baseVal)
	}
	if field, ok := baseStruct.Fields["name"].(runtime.StringValue); !ok || field.Val != "Alice" {
		t.Fatalf("base struct mutated unexpectedly: %#v", baseStruct.Fields["name"])
	}
}

func TestStructFunctionalUpdateWrongType(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"A",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"B",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("a"),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(10), "x")}, false, "A", nil, nil),
		),
		ast.Assign(
			ast.ID("b"),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(20), "y")}, false, "B", nil, nil),
		),
		ast.StructLit([]*ast.StructFieldInitializer{}, false, "A", ast.ID("b"), nil),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected functional update type mismatch error")
	}
	if got := err.Error(); got != "Functional update source must be same struct type" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructStaticMethod(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"origin",
					[]*ast.FunctionParameter{},
					[]ast.Statement{
						ast.Ret(ast.StructLit(
							[]*ast.StructFieldInitializer{
								ast.FieldInit(ast.Int(0), "x"),
								ast.FieldInit(ast.Int(0), "y"),
							},
							false,
							"Point",
							nil,
							nil,
						)),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.CallExpr(ast.Member(ast.ID("Point"), "origin")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inst, ok := result.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", result)
	}
	if inst.Fields == nil {
		t.Fatalf("expected named struct instance fields")
	}
	if x, ok := inst.Fields["x"].(runtime.IntegerValue); !ok || x.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("expected x == 0, got %#v", inst.Fields["x"])
	}
	if y, ok := inst.Fields["y"].(runtime.IntegerValue); !ok || y.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("expected y == 0, got %#v", inst.Fields["y"])
	}
}

func TestStructStaticMethodPrivateError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden_static",
					[]*ast.FunctionParameter{},
					[]ast.Statement{
						ast.Ret(ast.StructLit(
							[]*ast.StructFieldInitializer{
								ast.FieldInit(ast.Int(0), "x"),
								ast.FieldInit(ast.Int(0), "y"),
							},
							false,
							"Point",
							nil,
							nil,
						)),
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
			},
			nil,
			nil,
		),
		ast.CallExpr(ast.Member(ast.ID("Point"), "hidden_static")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected private static method error")
	}
	if got := err.Error(); got != "Method 'hidden_static' on Point is private" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructInstanceMethodCall(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"get",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Member(ast.ID("self"), "value")),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("c"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(5), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("c"), "get")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected result 5, got %#v", result)
	}
}

func TestStructInstanceMethodPrivateError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Int(1)),
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("c"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("c"), "hidden")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected private instance method error")
	}
	if got := err.Error(); got != "Method 'hidden' on Counter is private" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterfaceDynamicDispatch(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Display",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("point"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "x"),
					ast.FieldInit(ast.Int(3), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("value"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "point" {
		t.Fatalf("expected interface dispatch to return 'point', got %#v", result)
	}
}

func TestInterfaceAssignmentMissingImplementation(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Display",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected failure when assigning struct without impl to interface")
	}
	if got := err.Error(); got != "Typed pattern mismatch in assignment" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestInterfaceUnionDispatchPrefersSpecific(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Fancy",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Fancy"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Basic"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
					[]ast.Statement{ast.Ret(ast.Str("basic"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("union"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("items"),
			ast.Arr(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil),
			),
		),
		ast.Assign(ast.ID("buffer"), ast.Str("")),
		ast.ForLoopPattern(
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.ID("items"),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("buffer"),
					ast.Bin(
						"+",
						ast.ID("buffer"),
						ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
					),
				),
			),
		),
		ast.ID("buffer"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "fancybasic" {
		t.Fatalf("expected fancybasic, got %#v", result)
	}
}

func TestInterfaceDefaultMethodFallback(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Speakable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Str("default"))),
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Bot",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "default" {
		t.Fatalf("expected default, got %#v", result)
	}
}

func TestInterfaceDefaultMethodOverride(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Speakable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Str("default"))),
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Bot",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("custom"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "custom" {
		t.Fatalf("expected custom, got %#v", result)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
