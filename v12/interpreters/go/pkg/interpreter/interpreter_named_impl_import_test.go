package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestStaticImportRejectsNamedImplementationBindingCollision(t *testing.T) {
	interp := New()
	registerNamedImplementationNamespace(interp, "first", "Fancy", false)
	registerNamedImplementationNamespace(interp, "second", "Fancy", false)

	module := ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
		ast.Imp([]interface{}{"second"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
	}, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatal("expected named implementation collision to fail")
	} else if got, want := err.Error(), "Import error: named implementation binding 'Fancy' conflicts with an existing import; use a selector alias"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func TestStaticImportAllowsNamedImplementationSelectorRename(t *testing.T) {
	interp := New()
	registerNamedImplementationNamespace(interp, "first", "Fancy", false)
	registerNamedImplementationNamespace(interp, "second", "Fancy", false)

	module := ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
		ast.Imp([]interface{}{"second"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", "OtherFancy")}, nil),
	}, nil)
	_, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("renamed imports failed: %v", err)
	}
	for _, name := range []string{"Fancy", "OtherFancy"} {
		value, err := env.Get(name)
		if err != nil {
			t.Fatalf("missing renamed implementation binding %q: %v", name, err)
		}
		if _, ok := value.(runtime.ImplementationNamespaceValue); !ok {
			t.Fatalf("binding %q has type %T, want ImplementationNamespaceValue", name, value)
		}
	}
}

func TestStaticWildcardImportRejectsNamedImplementationBindingCollision(t *testing.T) {
	interp := New()
	registerNamedImplementationNamespace(interp, "first", "Fancy", false)
	registerNamedImplementationNamespace(interp, "second", "Fancy", false)

	module := ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, true, nil, nil),
		ast.Imp([]interface{}{"second"}, true, nil, nil),
	}, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatal("expected wildcard named implementation collision to fail")
	} else if got, want := err.Error(), "Import error: named implementation binding 'Fancy' conflicts with an existing import; use a selector alias"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func TestStaticImportRejectsPrivateNamedImplementation(t *testing.T) {
	interp := New()
	registerNamedImplementationNamespace(interp, "dep", "Hidden", true)

	module := ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"dep"}, false, []*ast.ImportSelector{ast.ImpSel("Hidden", nil)}, nil),
	}, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatal("expected private named implementation import to fail")
	} else if got, want := err.Error(), "Import error: implementation 'Hidden' is private"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func registerNamedImplementationNamespace(interp *Interpreter, pkg, name string, private bool) {
	interp.RegisterPackageSymbol(pkg, name, runtime.ImplementationNamespaceValue{
		Name:      ast.ID(name),
		Methods:   map[string]runtime.Value{},
		IsPrivate: private,
	})
}
