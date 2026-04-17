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

func TestLookupStructDefinitionSupportsDeepQualifiedPackageNames(t *testing.T) {
	interp := New()

	parentPkg := "able.spec"
	parentDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("CustomMatcher", nil, ast.StructKindNamed, nil, nil, false),
	}
	parentEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	parentEnv.DefineStruct("CustomMatcher", parentDef)
	interp.packageEnvs[parentPkg] = parentEnv
	interp.packageRegistry[parentPkg] = map[string]runtime.Value{"CustomMatcher": parentDef}

	deepPkg := "able.spec.assertions"
	deepDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("CustomMatcher", nil, ast.StructKindNamed, nil, nil, false),
	}
	deepEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	deepEnv.DefineStruct("CustomMatcher", deepDef)
	interp.packageEnvs[deepPkg] = deepEnv
	interp.packageRegistry[deepPkg] = map[string]runtime.Value{"CustomMatcher": deepDef}

	if got, ok := interp.LookupStructDefinition("able.spec.assertions.CustomMatcher"); !ok || got != deepDef {
		t.Fatalf("LookupStructDefinition(deep qualified) = (%v, %t), want (%v, true)", got, ok, deepDef)
	}
}

func TestLookupStructDefinitionSupportsCollapsedVisiblePackageAlias(t *testing.T) {
	interp := New()
	interp.currentPackage = "demo.demo"

	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false),
	}
	pkgEnv := runtime.NewEnvironment(interp.GlobalEnvironment())
	pkgEnv.DefineStruct("Box", def)
	interp.packageEnvs["demo.demo"] = pkgEnv
	interp.packageRegistry["demo.demo"] = map[string]runtime.Value{"Box": def}

	if got, ok := interp.LookupStructDefinition("demo.Box"); !ok || got != def {
		t.Fatalf("LookupStructDefinition(collapsed visible package alias) = (%v, %t), want (%v, true)", got, ok, def)
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
