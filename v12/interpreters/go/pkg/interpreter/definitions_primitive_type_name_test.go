package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCanonicalTypeNameSkipsEnvironmentForReservedScalarPrimitives(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	shadow := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Shadow", nil, ast.StructKindNamed, nil, nil, false),
	}
	primitiveNames := []string{
		"bool", "char", "nil", "void",
		"i8", "i16", "i32", "i64", "i128", "isize",
		"u8", "u16", "u32", "u64", "u128", "usize",
		"f32", "f64",
	}
	for _, name := range primitiveNames {
		env.Define(name, shadow)
		if got := canonicalTypeName(env, name); got != name {
			t.Fatalf("canonicalTypeName(%q) = %q, want reserved primitive name", name, got)
		}
	}
}

func TestCanonicalTypeNameStillResolvesNominalBindings(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	env.Define("Alias", &runtime.StructDefinitionValue{
		Node: ast.StructDef("Target", nil, ast.StructKindNamed, nil, nil, false),
	})
	if got := canonicalTypeName(env, "Alias"); got != "Target" {
		t.Fatalf("canonicalTypeName(Alias) = %q, want Target", got)
	}
}
