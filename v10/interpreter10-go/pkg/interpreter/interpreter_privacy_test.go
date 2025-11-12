package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestImportPrivateFunctionWithAliasFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_priv_func"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	importModule := ast.Mod(
		nil,
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_priv_func"}, false, []*ast.ImportSelector{ast.ImpSel("secret", "alias")}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private function to fail")
	} else if err.Error() != "Import error: function 'secret' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportPrivateStructFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef("SecretStruct", nil, ast.StructKindNamed, nil, nil, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_priv_struct"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	importModule := ast.Mod(
		nil,
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_priv_struct"}, false, []*ast.ImportSelector{ast.ImpSel("SecretStruct", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private struct to fail")
	} else if err.Error() != "Import error: struct 'SecretStruct' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportPublicStructAliasSucceeds(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef("PublicStruct", nil, ast.StructKindNamed, nil, nil, false),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_pub_struct"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	importModule := ast.Mod(
		nil,
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_pub_struct"}, false, []*ast.ImportSelector{ast.ImpSel("PublicStruct", "Pub")}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err != nil {
		t.Fatalf("expected public struct import with alias to succeed: %v", err)
	}
	if _, err := interp.global.Get("Pub"); err != nil {
		t.Fatalf("expected alias binding, got error: %v", err)
	}
}

func TestImportPrivateInterfaceFails(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Iface("HiddenIface", nil, nil, nil, nil, nil, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_priv_iface"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	importModule := ast.Mod(
		nil,
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_priv_iface"}, false, []*ast.ImportSelector{ast.ImpSel("HiddenIface", nil)}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err == nil {
		t.Fatalf("expected import of private interface to fail")
	} else if err.Error() != "Import error: interface 'HiddenIface' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportPublicInterfaceAliasSucceeds(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Iface("VisibleIface", nil, nil, nil, nil, nil, false),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_pub_iface"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	importModule := ast.Mod(
		nil,
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_pub_iface"}, false, []*ast.ImportSelector{ast.ImpSel("VisibleIface", "VI")}, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(importModule); err != nil {
		t.Fatalf("expected public interface import with alias to succeed: %v", err)
	}
	if _, err := interp.global.Get("VI"); err != nil {
		t.Fatalf("expected alias binding, got error: %v", err)
	}
}

func TestWildcardImportSkipsPrivateSymbols(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("visible", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
			ast.Fn("hidden", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_wildcard"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	module := ast.Mod(
		[]ast.Statement{
			ast.ID("visible"),
			ast.ID("hidden"),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkg_wildcard"}, true, nil, nil),
		},
		nil,
	)

	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected lookup of private symbol to fail")
	} else if err.Error() != "Undefined variable 'hidden'" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDynImportRejectsPrivateFunction(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("secret_dyn", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_dyn_priv"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	module := ast.Mod(
		[]ast.Statement{
			ast.DynImp([]interface{}{"pkg_dyn_priv"}, false, []*ast.ImportSelector{ast.ImpSel("secret_dyn", nil)}, nil),
		},
		nil,
		nil,
	)

	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected dyn import of private function to fail")
	} else if err.Error() != "dynimport error: function 'secret_dyn' is private" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDynImportWildcardSkipsPrivateSymbols(t *testing.T) {
	interp := New()
	packageModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("public_dyn", nil, []ast.Statement{ast.Ret(ast.Int(1))}, nil, nil, nil, false, false),
			ast.Fn("secret_dyn", nil, []ast.Statement{ast.Ret(ast.Int(2))}, nil, nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"pkg_dyn_wild"}, false),
	)
	mustEvalModule(t, interp, packageModule)

	module := ast.Mod(
		[]ast.Statement{
			ast.DynImp([]interface{}{"pkg_dyn_wild"}, true, nil, nil),
			ast.ID("public_dyn"),
			ast.ID("secret_dyn"),
		},
		nil,
		nil,
	)

	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected dyn lookup of private symbol to fail")
	} else if err.Error() != "Undefined variable 'secret_dyn'" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPrivateStaticMethodInaccessible(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, ast.Mod(
		[]ast.Statement{
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
		},
		nil,
		nil,
	))

	mustEvalModule(t, interp, ast.Mod(
		[]ast.Statement{
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
		},
		nil,
		nil,
	))

	callHidden := ast.Mod(
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("Point"), "hidden_static")),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private static method access to fail")
	} else if err.Error() != "Method 'hidden_static' on Point is private" {
		t.Fatalf("unexpected error message: %v", err)
	}

	callOrigin := ast.Mod(
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("Point"), "origin")),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(callOrigin)
	if err != nil {
		t.Fatalf("public static method call failed: %v", err)
	}
	if _, ok := result.(*runtime.StructInstanceValue); !ok {
		t.Fatalf("expected struct instance result, got %#v", result)
	}
}

func TestPrivateInstanceMethodInaccessible(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, ast.Mod(
		[]ast.Statement{
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
		},
		nil,
		nil,
	))

	mustEvalModule(t, interp, ast.Mod(
		[]ast.Statement{
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
		},
		nil,
		nil,
	))

	inst := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(5), "value"),
		},
		false,
		"Counter",
		nil,
		nil,
	)

	callHidden := ast.Mod(
		[]ast.Statement{
			ast.CallExpr(ast.Member(inst, "hidden")),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private instance method access to fail")
	} else if err.Error() != "Method 'hidden' on Counter is private" {
		t.Fatalf("unexpected error message: %v", err)
	}

	callGet := ast.Mod(
		[]ast.Statement{
			ast.CallExpr(ast.Member(inst, "get")),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(callGet)
	if err != nil {
		t.Fatalf("public instance method call failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected integer 5 result, got %#v", result)
	}
}
