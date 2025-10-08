package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

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
