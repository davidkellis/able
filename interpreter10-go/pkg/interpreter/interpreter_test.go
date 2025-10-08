package interpreter

import (
	"math"
	"math/big"
	"strings"
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

func TestImportRejectsPrivateFunction(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, true),
			ast.Fn("public", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"mypkg"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	importModule := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"mypkg"}, false, []*ast.ImportSelector{ast.ImpSel("secret", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private function to fail")
	} else if err.Error() != "Import error: function 'secret' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportMissingSymbolUsesGlobalsMessage(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("visible", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"otherpkg"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	missingImport := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"otherpkg"}, false, []*ast.ImportSelector{ast.ImpSel("unknown", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(missingImport); err == nil {
		t.Fatalf("expected missing symbol import to fail")
	} else if err.Error() != "Import error: symbol 'unknown' from 'otherpkg' not found in globals" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportRejectsPrivateStruct(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef("Secret", nil, ast.StructKindNamed, nil, nil, true),
			ast.StructDef("Public", nil, ast.StructKindNamed, nil, nil, false),
		},
		nil,
		ast.Pkg([]interface{}{"pkg"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	importModule := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg"}, false, []*ast.ImportSelector{ast.ImpSel("Secret", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private struct to fail")
	} else if err.Error() != "Import error: struct 'Secret' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportRejectsPrivateInterface(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Iface("Hidden", nil, nil, nil, nil, nil, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg2"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	importModule := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg2"}, false, []*ast.ImportSelector{ast.ImpSel("Hidden", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private interface to fail")
	} else if err.Error() != "Import error: interface 'Hidden' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestWildcardImportSkipsPrivateSymbols(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("visible", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg3"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	wildcardModule := ast.Mod(
		[]ast.Statement{
			ast.ID("visible"),
			ast.ID("secret"),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg3"}, true, nil, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(wildcardModule); err == nil {
		t.Fatalf("expected lookup of private symbol to fail")
	} else if !strings.Contains(err.Error(), "Undefined variable 'secret'") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDynImportRejectsPrivateFunction(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg4"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	dynModule := ast.Mod(
		[]ast.Statement{
			ast.DynImp([]interface{}{"pkg4"}, false, []*ast.ImportSelector{ast.ImpSel("secret", nil)}, nil),
		},
		nil,
		nil,
	)

	if _, _, err := interp.EvaluateModule(dynModule); err == nil {
		t.Fatalf("expected dyn import of private function to fail")
	} else if err.Error() != "dynimport error: function 'secret' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDynImportWildcardSkipsPrivateSymbols(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("public", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg5"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	module := ast.Mod(
		[]ast.Statement{
			ast.DynImp([]interface{}{"pkg5"}, true, nil, nil),
			ast.ID("public"),
			ast.ID("secret"),
		},
		nil,
		nil,
	)

	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected dyn lookup of private symbol to fail")
	} else if !strings.Contains(err.Error(), "Undefined variable 'secret'") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportReexportChain(t *testing.T) {
	interp := New()
	baseModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("base_value", nil, []ast.Statement{ast.Ret(ast.Int(41))}, nil, nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"static_base"}, false),
	)
	if _, _, err := interp.EvaluateModule(baseModule); err != nil {
		t.Fatalf("base module failed: %v", err)
	}

	middleModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"middle_value",
				nil,
				[]ast.Statement{ast.Ret(ast.Bin("+", ast.Call("base_value"), ast.Int(1)))},
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"static_base"}, false, []*ast.ImportSelector{ast.ImpSel("base_value", nil)}, nil),
		},
		ast.Pkg([]interface{}{"static_middle"}, false),
	)
	if _, _, err := interp.EvaluateModule(middleModule); err != nil {
		t.Fatalf("middle module failed: %v", err)
	}

	entryModule := ast.Mod(
		[]ast.Statement{
			ast.Call("middle_value"),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"static_middle"}, false, []*ast.ImportSelector{ast.ImpSel("middle_value", nil)}, nil),
		},
		nil,
	)

	result, _, err := interp.EvaluateModule(entryModule)
	if err != nil {
		t.Fatalf("entry module failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected integer 42, got %#v", result)
	}
}

func TestImportReexportMultiHop(t *testing.T) {
	interp := New()
	baseModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("base_value", nil, []ast.Statement{ast.Ret(ast.Int(40))}, nil, nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"static_base_multihop"}, false),
	)
	if _, _, err := interp.EvaluateModule(baseModule); err != nil {
		t.Fatalf("base module failed: %v", err)
	}

	middleModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"middle_value",
				nil,
				[]ast.Statement{ast.Ret(ast.Bin("+", ast.Call("base_value"), ast.Int(1)))},
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"static_base_multihop"}, false, []*ast.ImportSelector{ast.ImpSel("base_value", nil)}, nil),
		},
		ast.Pkg([]interface{}{"static_middle_multihop"}, false),
	)
	if _, _, err := interp.EvaluateModule(middleModule); err != nil {
		t.Fatalf("middle module failed: %v", err)
	}

	topModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"top_value",
				nil,
				[]ast.Statement{ast.Ret(ast.Bin("+", ast.Call("middle_value"), ast.Int(1)))},
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"static_middle_multihop"}, false, []*ast.ImportSelector{ast.ImpSel("middle_value", nil)}, nil),
		},
		ast.Pkg([]interface{}{"static_top_multihop"}, false),
	)
	if _, _, err := interp.EvaluateModule(topModule); err != nil {
		t.Fatalf("top module failed: %v", err)
	}

	entryModule := ast.Mod(
		[]ast.Statement{
			ast.Call("top_value"),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"static_top_multihop"}, false, []*ast.ImportSelector{ast.ImpSel("top_value", nil)}, nil),
		},
		nil,
	)

	result, _, err := interp.EvaluateModule(entryModule)
	if err != nil {
		t.Fatalf("entry module failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected integer 42, got %#v", result)
	}
}

func TestImportPackageAliasExposesPublicSymbols(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("public_fn", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
			ast.Fn("secret_fn", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_alias"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	aliasModule := ast.Mod(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_alias"}, false, nil, "Alias"),
		},
		nil,
	)

	if _, env, err := interp.EvaluateModule(aliasModule); err != nil {
		t.Fatalf("alias module failed: %v", err)
	} else if env == nil {
		t.Fatalf("expected module environment")
	} else {
		val, err := env.Get("Alias")
		if err != nil {
			t.Fatalf("alias not defined: %v", err)
		}
		pkgVal, ok := val.(runtime.PackageValue)
		if !ok {
			t.Fatalf("expected PackageValue, got %#v", val)
		}
		if len(pkgVal.NamePath) != 1 || pkgVal.NamePath[0] != "pkg_alias" {
			t.Fatalf("unexpected NamePath: %#v", pkgVal.NamePath)
		}
		if _, ok := pkgVal.Public["secret_fn"]; ok {
			t.Fatalf("private symbol leaked in package alias")
		}
		fnVal, ok := pkgVal.Public["public_fn"].(*runtime.FunctionValue)
		if !ok || fnVal == nil {
			t.Fatalf("public function missing from package alias")
		}
	}
}

func TestDynImportAliasAndSelectors(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("f", nil, []ast.Statement{ast.Ret(ast.Int(11))}, nil, nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"dynp"}, false),
	)
	if _, _, err := interp.EvaluateModule(packageModule); err != nil {
		t.Fatalf("package module evaluation failed: %v", err)
	}

	module := ast.Mod(
		[]ast.Statement{
			ast.DynImp([]interface{}{"dynp"}, false, []*ast.ImportSelector{ast.ImpSel("f", "ff")}, nil),
			ast.DynImp([]interface{}{"dynp"}, false, nil, "D"),
			ast.Assign(ast.ID("x"), ast.Call("ff")),
			ast.Assign(ast.ID("y"), ast.CallExpr(ast.Member(ast.ID("D"), "f"))),
			ast.Bin("+", ast.ID("x"), ast.ID("y")),
		},
		nil,
		nil,
	)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("dyn import module failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(22)) != 0 {
		t.Fatalf("expected integer 22, got %#v", result)
	}
	if env == nil {
		t.Fatalf("expected module environment")
	}
	packageHandle, err := env.Get("D")
	if err != nil {
		t.Fatalf("dyn package alias missing: %v", err)
	}
	dynPkg, ok := packageHandle.(runtime.DynPackageValue)
	if !ok {
		t.Fatalf("expected DynPackageValue, got %#v", packageHandle)
	}
	if len(dynPkg.NamePath) != 1 || dynPkg.NamePath[0] != "dynp" {
		t.Fatalf("unexpected dyn package NamePath: %#v", dynPkg.NamePath)
	}
	if dynPkg.Name != "dynp" {
		t.Fatalf("unexpected dyn package Name: %s", dynPkg.Name)
	}
}

func TestRescueExpressionPattern(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Int(42))),
		ast.Mc(ast.Wc(), ast.Int(7)),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("rescue evaluation failed: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", val)
	}
}

func TestOrElseExpressionBindsError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.OrElse(
		ast.Prop(ast.Block(ast.Raise(ast.Str("x")))),
		"e",
		ast.Str("handled"),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("or else evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "handled" {
		t.Fatalf("expected string 'handled', got %#v", val)
	}
}

func TestEnsureRunsOnError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("flag"), ast.Str("")), env); err != nil {
		t.Fatalf("failed to initialise flag: %v", err)
	}
	expr := ast.Ensure(
		ast.Rescue(
			ast.Block(ast.Raise(ast.Str("oops"))),
			ast.Mc(ast.Wc(), ast.Str("rescued")),
		),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Str("done")),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("ensure evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "rescued" {
		t.Fatalf("expected string 'rescued', got %#v", val)
	}
	flagVal, err := env.Get("flag")
	if err != nil {
		t.Fatalf("expected flag binding: %v", err)
	}
	flagStr, ok := flagVal.(runtime.StringValue)
	if !ok || flagStr.Val != "done" {
		t.Fatalf("expected flag 'done', got %#v", flagVal)
	}
}

func TestRethrowPropagatesToOuterRescue(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	inner := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("oops"))),
		ast.Mc(ast.Wc(), ast.Block(ast.Rethrow())),
	)
	outer := ast.Rescue(
		inner,
		ast.Mc(ast.Wc(), ast.Str("handled")),
	)
	val, err := interp.evaluateExpression(outer, env)
	if err != nil {
		t.Fatalf("outer rescue failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "handled" {
		t.Fatalf("expected string 'handled', got %#v", val)
	}
}

func TestRaiseConvertsValueToErrorStruct(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateStatement(ast.Raise(ast.Int(5)), env)
	if err == nil {
		t.Fatalf("expected raise to escape")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %#v", rs.value)
	}
	if errVal.Message == "" {
		t.Fatalf("expected error message")
	}
	if _, ok := errVal.Payload["value"]; !ok {
		t.Fatalf("expected payload with original value")
	}
}

func TestRescueGuardSkipsClause(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("boom"))),
		ast.Mc(ast.ID("msg"), ast.Str("guard"), ast.Bin("==", ast.ID("msg"), ast.Str("skip"))),
		ast.Mc(ast.Wc(), ast.Str("fallback")),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("rescue evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "fallback" {
		t.Fatalf("expected fallback string, got %#v", val)
	}
}

func TestRescueNoMatchPropagatesError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("boom"))),
		ast.Mc(ast.LitP(ast.Str("ok")), ast.Str("handled")),
	)
	_, err := interp.evaluateExpression(expr, env)
	if err == nil {
		t.Fatalf("expected error to propagate")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %#v", rs.value)
	}
	if errVal.Message != "boom" {
		t.Fatalf("expected message 'boom', got %q", errVal.Message)
	}
}

func TestPropagationExpressionRaisesOnErrorValue(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	env.Define("err", runtime.ErrorValue{Message: "bad"})
	expr := ast.Prop(ast.ID("err"))
	_, err := interp.evaluateExpression(expr, env)
	if err == nil {
		t.Fatalf("expected propagation to raise")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok || errVal.Message != "bad" {
		t.Fatalf("expected ErrorValue 'bad', got %#v", rs.value)
	}
}

func TestPropagationExpressionPassesThroughNonError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Prop(ast.Int(9))
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected integer 9, got %#v", val)
	}
}

func TestBitshiftRangeDiagnostics(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{
			name: "NegativeShift",
			expr: ast.Bin("<<", ast.Int(1), ast.Int(-1)),
			msg:  "shift out of range",
		},
		{
			name: "LargeShift",
			expr: ast.Bin(">>", ast.Int(1), ast.Int(32)),
			msg:  "shift out of range",
		},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestBitwiseRequiresI32Operands(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("&", ast.Int(1), ast.Flt(1.0)),
		ast.Bin("|", ast.Flt(1.0), ast.Int(1)),
		ast.Bin("^", ast.Flt(1.0), ast.Flt(1.0)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "Bitwise requires i32 operands" {
			t.Fatalf("case %d: expected 'Bitwise requires i32 operands', got %q", idx, err.Error())
		}
	}
}

func TestDivisionByZeroDiagnostics(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("/", ast.Int(4), ast.Int(0)),
		ast.Bin("%", ast.Int(4), ast.Int(0)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "division by zero" {
			t.Fatalf("case %d: expected 'division by zero', got %q", idx, err.Error())
		}
	}
}

func TestArithmeticMixedNumericTypes(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect float64
	}{
		{"AddMixed", ast.Bin("+", ast.Int(1), ast.Flt(1.5)), 2.5},
		{"SubMixed", ast.Bin("-", ast.Flt(2.5), ast.Int(1)), 1.5},
		{"MulMixed", ast.Bin("*", ast.Int(2), ast.Flt(3.5)), 7},
		{"DivMixed", ast.Bin("/", ast.Flt(7), ast.Int(2)), 3.5},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		fv, ok := val.(runtime.FloatValue)
		if !ok {
			t.Fatalf("%s: expected float result, got %#v", tc.name, val)
		}
		if math.Abs(fv.Val-tc.expect) > 1e-9 {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, fv.Val)
		}
	}
}

func TestModuloMixedNumericTypes(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect float64
	}{
		{"ModuloIntFloat", ast.Bin("%", ast.Int(5), ast.Flt(2)), 1},
		{"ModuloFloatInt", ast.Bin("%", ast.Flt(5), ast.Int(2)), 1},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		fv, ok := val.(runtime.FloatValue)
		if ok {
			if math.Abs(fv.Val-tc.expect) > 1e-9 {
				t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, fv.Val)
			}
			continue
		}
		iv, ok := val.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("%s: unexpected result %#v", tc.name, val)
		}
		if iv.Val.Cmp(bigInt(int64(tc.expect))) != 0 {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, iv.Val)
		}
	}
}

func TestMixedNumericComparisons(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect bool
	}{
		{"LessMixed", ast.Bin("<", ast.Int(2), ast.Flt(3.5)), true},
		{"GreaterMixed", ast.Bin(">", ast.Flt(3.5), ast.Int(2)), true},
		{"EqualMixed", ast.Bin("==", ast.Int(3), ast.Flt(3)), true},
		{"NotEqualMixed", ast.Bin("!=", ast.Int(3), ast.Flt(4)), true},
		{"LessMixedFalse", ast.Bin("<", ast.Flt(5), ast.Int(4)), false},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		bv, ok := val.(runtime.BoolValue)
		if !ok {
			t.Fatalf("%s: expected bool, got %#v", tc.name, val)
		}
		if bv.Val != tc.expect {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, bv.Val)
		}
	}
}

func TestUnsupportedBinaryOperatorError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{"StringTimes", ast.Bin("*", ast.Str("hi"), ast.Str("there")), "Arithmetic requires numeric operands"},
		{"BoolPlus", ast.Bin("+", ast.Bool(true), ast.Bool(false)), "Arithmetic requires numeric operands"},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestStringConcatRequiresStrings(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateExpression(ast.Bin("+", ast.Str("hi"), ast.Int(2)), env)
	if err == nil {
		t.Fatalf("expected concat error")
	}
	if err.Error() != "Arithmetic requires numeric operands" {
		t.Fatalf("expected string concat error, got %q", err.Error())
	}
}

func TestComparisonTypeErrors(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("<", ast.Str("a"), ast.Int(1)),
		ast.Bin("<", ast.Bool(true), ast.Int(1)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "Arithmetic requires numeric operands" {
			t.Fatalf("case %d: unexpected error %v", idx, err)
		}
	}
}

func TestFunctionDefinitionUnknownInterfaceConstraint(t *testing.T) {
	interp := New()
	fn := ast.Fn(
		"identity",
		[]*ast.FunctionParameter{ast.Param("value", nil)},
		[]ast.Statement{ast.ID("value")},
		nil,
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Error"))),
		},
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{fn}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected unknown interface error")
	} else if err.Error() != "Unknown interface 'Error' in constraint on 'T'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setupShowPoint(t *testing.T, interp *Interpreter) {
	t.Helper()
	show := ast.Iface(
		"Show",
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
	)
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{show}, nil, nil)); err != nil {
		t.Fatalf("interface evaluation failed: %v", err)
	}
	point := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{point}, nil, nil)); err != nil {
		t.Fatalf("struct evaluation failed: %v", err)
	}
	toString := ast.Fn(
		"to_string",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
		[]ast.Statement{
			ast.Ret(
				ast.Interp(
					ast.Str("Point("),
					ast.Member(ast.ID("self"), "x"),
					ast.Str(", "),
					ast.Member(ast.ID("self"), "y"),
					ast.Str(")"),
				),
			),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	acceptShow := ast.Fn(
		"accept_show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Point")),
			ast.Param("x", nil),
		},
		[]ast.Statement{
			ast.Ret(ast.CallExpr(ast.Member(ast.ID("x"), "to_string"))),
		},
		ast.Ty("string"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{toString, acceptShow},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{methods}, nil, nil)); err != nil {
		t.Fatalf("methods evaluation failed: %v", err)
	}
}

func TestFunctionCallGenericConstraintSatisfied(t *testing.T) {
	interp := New()
	setupShowPoint(t, interp)
	module := ast.Mod(
		[]ast.Statement{
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
			ast.CallT(
				ast.Member(ast.ID("p"), "accept_show"),
				[]ast.TypeExpression{ast.Ty("Point")},
				ast.ID("p"),
			),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("call evaluation failed: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %T", result)
	}
	if str.Val != "Point(1, 2)" {
		t.Fatalf("expected Point(1, 2), got %s", str.Val)
	}
}

func TestFunctionCallGenericConstraintViolation(t *testing.T) {
	interp := New()
	setupShowPoint(t, interp)
	module := ast.Mod(
		[]ast.Statement{
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
			ast.CallT(
				ast.Member(ast.ID("p"), "accept_show"),
				[]ast.TypeExpression{ast.Ty("i32")},
				ast.Int(3),
			),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected constraint violation")
	} else if err.Error() != "Type 'i32' does not satisfy interface 'Show': missing method 'to_string'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUfcsOnPrimitiveValue(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"add",
				[]*ast.FunctionParameter{
					ast.Param("a", nil),
					ast.Param("b", nil),
				},
				[]ast.Statement{
					ast.Ret(ast.Bin("+", ast.ID("a"), ast.ID("b"))),
				},
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.CallExpr(ast.Member(ast.Int(4), "add"), ast.Int(5)),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	iv, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}
	if iv.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected 9, got %v", iv.Val)
	}
}

func TestUfcsOnStructInstance(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Point",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("i32"), "x"),
				},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Fn(
				"move",
				[]*ast.FunctionParameter{
					ast.Param("p", nil),
					ast.Param("dx", nil),
				},
				[]ast.Statement{
					ast.AssignMember(ast.ID("p"), "x", ast.Bin("+", ast.Member(ast.ID("p"), "x"), ast.ID("dx"))),
					ast.Ret(ast.ID("p")),
				},
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.Assign(
				ast.ID("p"),
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
				}, false, "Point", nil, nil),
			),
			ast.CallExpr(ast.Member(ast.ID("p"), "move"), ast.Int(3)),
		},
		nil,
		nil,
	)
	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	if _, ok := result.(*runtime.StructInstanceValue); !ok {
		t.Fatalf("expected struct instance result, got %T", result)
	}
	val, err := env.Get("p")
	if err != nil {
		t.Fatalf("env lookup failed: %v", err)
	}
	inst, ok := val.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance for p, got %T", val)
	}
	field, ok := inst.Fields["x"]
	if !ok {
		t.Fatalf("field x missing on struct instance")
	}
	iv, ok := field.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer field, got %T", field)
	}
	if iv.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected updated x=4, got %v", iv.Val)
	}
}

func TestNamedImplDisambiguation(t *testing.T) {
	interp := New()
	defs := []ast.Statement{
		ast.StructDef("Service", nil, ast.StructKindNamed, nil, nil, false),
		ast.Iface(
			"A",
			[]*ast.FunctionSignature{
				ast.FnSig("act", []*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))}, ast.Ty("string"), nil, nil, nil),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Iface(
			"B",
			[]*ast.FunctionSignature{
				ast.FnSig("act", []*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))}, ast.Ty("string"), nil, nil, nil),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"A",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("A"))},
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
			"B",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("B"))},
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
			"A",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("A.named"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			"ActA",
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"B",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("B.named"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			"ActB",
			nil,
			nil,
			nil,
			false,
		),
	}
	if _, _, err := interp.EvaluateModule(ast.Mod(defs, nil, nil)); err != nil {
		t.Fatalf("definitions failed: %v", err)
	}
	serviceLiteral := ast.StructLit(nil, false, "Service", nil, nil)
	ambiguous := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(serviceLiteral, "act")),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(ambiguous); err == nil {
		t.Fatalf("expected ambiguity error")
	} else if !strings.Contains(err.Error(), "Ambiguous method 'act' for type 'Service'") {
		t.Fatalf("unexpected error: %v", err)
	}
	callA := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("ActA"), "act"), ast.StructLit(nil, false, "Service", nil, nil)),
	}, nil, nil)
	valA, _, err := interp.EvaluateModule(callA)
	if err != nil {
		t.Fatalf("ActA call failed: %v", err)
	}
	strA, ok := valA.(runtime.StringValue)
	if !ok || strA.Val != "A.named" {
		t.Fatalf("expected A.named, got %#v", valA)
	}
	callB := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("ActB"), "act"), ast.StructLit(nil, false, "Service", nil, nil)),
	}, nil, nil)
	valB, _, err := interp.EvaluateModule(callB)
	if err != nil {
		t.Fatalf("ActB call failed: %v", err)
	}
	strB, ok := valB.(runtime.StringValue)
	if !ok || strB.Val != "B.named" {
		t.Fatalf("expected B.named, got %#v", valB)
	}
}

func TestUnnamedImplDuplicateRejected(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.Iface(
				"M",
				[]*ast.FunctionSignature{
					ast.FnSig("id", nil, ast.Ty("Self"), nil, nil, nil),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
			ast.Impl(
				"M",
				ast.Ty("i32"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"id",
						nil,
						[]ast.Statement{
							ast.Ret(ast.Int(0)),
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
				nil,
				nil,
				false,
			),
			ast.Impl(
				"M",
				ast.Ty("i32"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"id",
						nil,
						[]ast.Statement{
							ast.Ret(ast.Int(1)),
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
				nil,
				nil,
				false,
			),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected duplicate unnamed impl error")
	} else if err.Error() != "Unnamed impl for (M, i32) already exists" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogicalOperandErrors(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{"LogicalAndLeft", ast.Bin("&&", ast.Int(1), ast.Bool(true)), "Logical operands must be bool"},
		{"LogicalAndRight", ast.Bin("&&", ast.Bool(true), ast.Int(1)), "Logical operands must be bool"},
		{"LogicalOrLeft", ast.Bin("||", ast.Int(0), ast.Bool(false)), "Logical operands must be bool"},
		{"LogicalOrRight", ast.Bin("||", ast.Bool(false), ast.Int(1)), "Logical operands must be bool"},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestRangeBoundariesMustBeNumeric(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateExpression(ast.Range(ast.Bool(true), ast.Int(5), true), env)
	if err == nil {
		t.Fatalf("expected range error")
	}
	if err.Error() != "Range boundaries must be numeric" {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = interp.evaluateExpression(ast.Range(ast.Int(1), ast.Bool(false), false), env)
	if err == nil {
		t.Fatalf("expected range error for end")
	}
	if err.Error() != "Range boundaries must be numeric" {
		t.Fatalf("unexpected error for end: %v", err)
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

func TestInterfaceDynamicDispatchUsesUnderlyingImpl(t *testing.T) {
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
					[]ast.Statement{
						ast.Ret(
							ast.Interp(
								ast.Str("Point("),
								ast.Member(ast.ID("self"), "x"),
								ast.Str(", "),
								ast.Member(ast.ID("self"), "y"),
								ast.Str(")"),
							),
						),
					},
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
	if !ok || str.Val != "Point(2, 3)" {
		t.Fatalf("expected Point(2, 3), got %#v", result)
	}
}

func TestInterfaceDynamicDispatchPrefersMostSpecificImpl(t *testing.T) {
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
				ast.StructLit(
					[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")},
					false,
					"Fancy",
					nil,
					nil,
				),
				ast.StructLit(
					[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")},
					false,
					"Basic",
					nil,
					nil,
				),
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

func TestImplResolutionPrefersStricterConstraints(t *testing.T) {
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
		ast.Iface(
			"Copyable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("Self"),
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
			"Item",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("T"), "value"),
			},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Ty("Item"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Item"))},
					[]ast.Statement{ast.Ret(ast.Str("Item"))},
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
			"Copyable",
			ast.Ty("Item"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Item"))},
					[]ast.Statement{ast.Ret(ast.ID("self"))},
					ast.Ty("Item"),
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
			"Display",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("generic"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("copyable"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")), ast.InterfaceConstr(ast.Ty("Copyable")))},
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("item"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(3), "value")},
				false,
				"Item",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("wrapper"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.ID("item"), "value")},
				false,
				"Wrapper",
				nil,
				[]ast.TypeExpression{ast.Ty("Item")},
			),
		),
		ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "copyable" {
		t.Fatalf("expected copyable, got %#v", result)
	}
}

func TestImplResolutionAmbiguousMultiTraitConstraints(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
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
		ast.Iface(
			"Copyable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("Self"),
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
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("T"), "value"),
			},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("Point"))},
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
			"Copyable",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.ID("self"))},
					ast.Ty("Point"),
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
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("show-constrained"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("copy-constrained"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Copyable")))},
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("point"),
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
		ast.Assign(
			ast.ID("wrapper"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.ID("point"), "value")},
				false,
				"Wrapper",
				nil,
				[]ast.TypeExpression{ast.Ty("Point")},
			),
		),
		ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected ambiguity error, got nil")
	}
	if !strings.Contains(err.Error(), "Ambiguous method 'to_string'") {
		t.Fatalf("expected ambiguous method error, got %v", err)
	}
}

func TestImplResolutionInherentMethodsTakePrecedence(t *testing.T) {
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
			"Bot",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "id")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("beep inherent"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("beep impl"))},
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
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(42), "id")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "beep inherent" {
		t.Fatalf("expected beep inherent, got %#v", result)
	}
}

func TestImplResolutionMoreSpecificImplWinsOverGeneric(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
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
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("Point inherent"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("generic"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("Point")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("Point")))},
					[]ast.Statement{ast.Ret(ast.Str("specific"))},
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
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(
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
							"value",
						),
					},
					false,
					"Wrapper",
					nil,
					[]ast.TypeExpression{ast.Ty("Point")},
				),
				"to_string",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "specific" {
		t.Fatalf("expected specific, got %#v", result)
	}
}

func TestImplResolutionWhereClauseSupersetPreferred(t *testing.T) {
	buildModule := func(wrapper ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"Readable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"read",
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
			ast.Iface(
				"Writable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"write",
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
			ast.Iface(
				"Show",
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Wrap",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
			ast.Impl(
				"Readable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("read-fancy"))},
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
				"Writable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"write",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("write-fancy"))},
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
				"Readable",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("read-basic"))},
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
				ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrap"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "read")),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable"))),
				},
				false,
			),
			ast.Impl(
				"Show",
				ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrap"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "write")),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Writable"))),
				},
				false,
			),
			ast.Assign(ast.ID("wrapper"), wrapper),
			ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
		}, nil, nil)
	}

	fancyInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	fancyWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(fancyInner, "value")},
		false,
		"Wrap",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyWrapper)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "write-fancy" {
		t.Fatalf("expected write-fancy, got %#v", resultFancy)
	}

	basicInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	basicWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(basicInner, "value")},
		false,
		"Wrap",
		nil,
		[]ast.TypeExpression{ast.Ty("Basic")},
	)
	moduleBasic := buildModule(basicWrapper)
	interpBasic := New()
	resultBasic, _, err := interpBasic.EvaluateModule(moduleBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strBasic, ok := resultBasic.(runtime.StringValue)
	if !ok || strBasic.Val != "read-basic" {
		t.Fatalf("expected read-basic, got %#v", resultBasic)
	}
}

func TestImplResolutionWhereClauseMultiParamPreferred(t *testing.T) {
	buildModule := func(pair ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"Readable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"read",
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
			ast.Iface(
				"Writable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"write",
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
			ast.Iface(
				"Combine",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"combine",
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
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Pair",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("A"), "left"),
					ast.FieldDef(ast.Ty("B"), "right"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				false,
			),
			ast.Impl(
				"Readable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("read-fancy"))},
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
				"Writable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"write",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("write-fancy"))},
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
				"Readable",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("read-basic"))},
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
				"Combine",
				ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"combine",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")))},
						[]ast.Statement{
							ast.Ret(
								ast.Bin(
									"+",
									ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "left"), "read")),
									ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "right"), "read")),
								),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("A", ast.InterfaceConstr(ast.Ty("Readable"))),
					ast.WhereConstraint("B", ast.InterfaceConstr(ast.Ty("Readable"))),
				},
				false,
			),
			ast.Impl(
				"Combine",
				ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"combine",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "right"), "write")),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("A", ast.InterfaceConstr(ast.Ty("Readable"))),
					ast.WhereConstraint("B", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Writable"))),
				},
				false,
			),
			ast.Assign(ast.ID("pair"), pair),
			ast.CallExpr(ast.Member(ast.ID("pair"), "combine")),
		}, nil, nil)
	}

	fancy := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	basic := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	fancyPair := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(fancy, "left"),
			ast.FieldInit(fancy, "right"),
		},
		false,
		"Pair",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy"), ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyPair)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "write-fancy" {
		t.Fatalf("expected write-fancy, got %#v", resultFancy)
	}

	mixedPair := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(fancy, "left"),
			ast.FieldInit(basic, "right"),
		},
		false,
		"Pair",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy"), ast.Ty("Basic")},
	)
	moduleMixed := buildModule(mixedPair)
	interpMixed := New()
	resultMixed, _, err := interpMixed.EvaluateModule(moduleMixed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strMixed, ok := resultMixed.(runtime.StringValue)
	if !ok || strMixed.Val != "read-fancyread-basic" {
		t.Fatalf("expected read-fancyread-basic, got %#v", resultMixed)
	}
}

func TestImplResolutionUnionSpecificityPrefersSubset(t *testing.T) {
	buildModule := func(lit ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"Show",
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Extra",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Impl(
				"Show",
				ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", nil)},
						[]ast.Statement{ast.Ret(ast.Str("pair"))},
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
				ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic"), ast.Ty("Extra")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", nil)},
						[]ast.Statement{ast.Ret(ast.Str("triple"))},
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
			ast.CallExpr(ast.Member(lit, "to_string")),
		}, nil, nil)
	}

	fancyLiteral := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	moduleFancy := buildModule(fancyLiteral)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "pair" {
		t.Fatalf("expected pair, got %#v", resultFancy)
	}

	extraLiteral := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("e"), "label")}, false, "Extra", nil, nil)
	moduleExtra := buildModule(extraLiteral)
	interpExtra := New()
	resultExtra, _, err := interpExtra.EvaluateModule(moduleExtra)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strExtra, ok := resultExtra.(runtime.StringValue)
	if !ok || strExtra.Val != "triple" {
		t.Fatalf("expected triple, got %#v", resultExtra)
	}
}

func TestImplResolutionUnionAmbiguousWithoutSubset(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
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
			"Fancy",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Extra",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("pair"))},
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
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Extra")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("other"))},
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
		ast.CallExpr(ast.Member(
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
			"to_string",
		)),
	}, nil, nil)

	interp := New()
	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "Ambiguous method 'to_string'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterfaceDynamicValueUsesUnionImpl(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
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
			"Fancy",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
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
					"to_string",
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
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
		),
		ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "union" {
		t.Fatalf("expected union, got %#v", result)
	}
}

func TestImplResolutionInterfaceInheritancePrefersDeeper(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
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
		ast.Iface(
			"FancyShow",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			ast.Ty("Show"),
			nil,
			[]ast.TypeExpression{ast.Ty("Show")},
			false,
		),
		ast.Iface(
			"ShinyShow",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"shine",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			ast.Ty("Show"),
			nil,
			[]ast.TypeExpression{ast.Ty("FancyShow")},
			false,
		),
		ast.StructDef(
			"FancyBase",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"FancySpecial",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Wrap",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("FancyBase"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyBase"))},
					[]ast.Statement{ast.Ret(ast.Str("base"))},
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
			"FancyShow",
			ast.Ty("FancyBase"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyBase"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy-base"))},
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
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("special"))},
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
			"FancyShow",
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy-special"))},
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
			"ShinyShow",
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"shine",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("shine"))},
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
			ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Wrap"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "fancy")),
						),
					},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("FancyShow"))),
			},
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Wrap"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "shine")),
						),
					},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("ShinyShow"))),
			},
			false,
		),
		ast.CallExpr(ast.Member(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(
						ast.StructLit(
							[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("s"), "label")},
							false,
							"FancySpecial",
							nil,
							nil,
						),
						"value",
					),
				},
				false,
				"Wrap",
				nil,
				[]ast.TypeExpression{ast.Ty("FancySpecial")},
			),
			"to_string",
		)),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "shine" {
		t.Fatalf("expected shine, got %#v", result)
	}
}

func TestImplResolutionNestedGenericConstraintsPreferred(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Readable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"read",
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
		ast.Iface(
			"Comparable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"cmp",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self")), ast.Param("other", ast.Ty("Self"))},
					ast.Ty("i32"),
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
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"show",
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
			"Container",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.StructDef(
			"FancyNum",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Readable",
			ast.Ty("FancyNum"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"read",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyNum"))},
					[]ast.Statement{
						ast.Ret(
							ast.Interp(
								ast.Str("#"),
								ast.Member(ast.ID("self"), "value"),
							),
						),
					},
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
			"Comparable",
			ast.Ty("FancyNum"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"cmp",
					[]*ast.FunctionParameter{
						ast.Param("self", ast.Ty("FancyNum")),
						ast.Param("other", ast.Ty("FancyNum")),
					},
					[]ast.Statement{ast.Ret(ast.Int(0))},
					ast.Ty("i32"),
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
			ast.Gen(ast.Ty("Container"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"show",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Container"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "read")),
						),
					},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable"))),
			},
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Container"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"show",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Container"))},
					[]ast.Statement{ast.Ret(ast.Str("comparable"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Comparable"))),
			},
			false,
		),
		ast.CallExpr(ast.Member(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(
						ast.StructLit(
							[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(42), "value")},
							false,
							"FancyNum",
							nil,
							nil,
						),
						"value",
					),
				},
				false,
				"Container",
				nil,
				[]ast.TypeExpression{ast.Ty("FancyNum")},
			),
			"show",
		)),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "comparable" {
		t.Fatalf("expected comparable, got %#v", result)
	}
}

func TestImplResolutionSupersetConstraintsPreferred(t *testing.T) {
	buildModule := func(wrapper ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"TraitA",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"trait_a",
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
			ast.Iface(
				"TraitB",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"trait_b",
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
			ast.Iface(
				"Show",
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Wrapper",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
			ast.Impl(
				"TraitA",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_a",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("A:Fancy"))},
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
				"TraitB",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_b",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("B:Fancy"))},
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
				"TraitA",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_a",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("A:Basic"))},
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
				ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "trait_a")),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("TraitA")))},
				nil,
				nil,
				false,
			),
			ast.Impl(
				"Show",
				ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "trait_b")),
							),
						},
						ast.Ty("string"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("TraitA")), ast.InterfaceConstr(ast.Ty("TraitB")))},
				nil,
				nil,
				false,
			),
			ast.Assign(ast.ID("wrapper"), wrapper),
			ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
		}, nil, nil)
	}

	fancyInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	fancyWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(fancyInner, "value")},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyWrapper)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "B:Fancy" {
		t.Fatalf("expected B:Fancy, got %#v", resultFancy)
	}

	basicInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	basicWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(basicInner, "value")},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("Basic")},
	)
	moduleBasic := buildModule(basicWrapper)
	interpBasic := New()
	resultBasic, _, err := interpBasic.EvaluateModule(moduleBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strBasic, ok := resultBasic.(runtime.StringValue)
	if !ok || strBasic.Val != "A:Basic" {
		t.Fatalf("expected A:Basic, got %#v", resultBasic)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
