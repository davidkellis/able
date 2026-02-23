package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestLookupStructDefinitionPrefersPackageStructOverSameNameFunction(t *testing.T) {
	interp := New()
	pkgName := "compiled_tests"

	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("CountingReporter", nil, ast.StructKindNamed, nil, nil, false),
	}
	pkgEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	pkgEnv.DefineStruct("CountingReporter", def)
	interp.packageEnvs[pkgName] = pkgEnv
	interp.packageRegistry[pkgName] = map[string]runtime.Value{
		"CountingReporter": runtime.NativeFunctionValue{
			Name:  "CountingReporter",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				return runtime.NilValue{}, nil
			},
		},
	}

	if got, ok := interp.LookupStructDefinition("CountingReporter"); !ok || got != def {
		t.Fatalf("LookupStructDefinition(simple) = (%v, %t), want (%v, true)", got, ok, def)
	}
	if got, ok := interp.LookupStructDefinition("compiled_tests.CountingReporter"); !ok || got != def {
		t.Fatalf("LookupStructDefinition(qualified) = (%v, %t), want (%v, true)", got, ok, def)
	}
}

func TestSeedStructDefinitionsCopiesKnownStructsIntoDestinationEnv(t *testing.T) {
	interp := New()
	pkgName := "compiled_tests"

	reporterDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("CountingReporter", nil, ast.StructKindNamed, nil, nil, false),
	}
	personDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Person", nil, ast.StructKindNamed, nil, nil, false),
	}

	pkgEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	pkgEnv.DefineStruct("CountingReporter", reporterDef)
	interp.packageEnvs[pkgName] = pkgEnv
	interp.packageRegistry[pkgName] = map[string]runtime.Value{
		"CountingReporter": reporterDef,
		"Person":           personDef,
	}

	dst := runtime.NewEnvironment(interp.GlobalEnvironment())
	if seeded := interp.SeedStructDefinitions(dst); seeded < 2 {
		t.Fatalf("SeedStructDefinitions seeded %d entries, want at least 2", seeded)
	}
	if got, ok := dst.StructDefinition("CountingReporter"); !ok || got != reporterDef {
		t.Fatalf("dst.StructDefinition(CountingReporter) = (%v, %t), want (%v, true)", got, ok, reporterDef)
	}
	if got, ok := dst.StructDefinition("Person"); !ok || got != personDef {
		t.Fatalf("dst.StructDefinition(Person) = (%v, %t), want (%v, true)", got, ok, personDef)
	}
}
