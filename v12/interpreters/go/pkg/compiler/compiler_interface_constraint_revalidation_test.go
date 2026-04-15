package compiler

import (
	"strings"
	"testing"
)

func genericInterfaceConstraintRevalidationSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"type Labelish = Label",
		"",
		"interface Display for Self {",
		"  fn to_string(self: Self) -> String",
		"}",
		"",
		"interface Echo for Self {",
		"  fn pass<U>(self: Self, value: U) -> U",
		"}",
		"",
		"struct Label { text: String }",
		"struct Box T { value: T }",
		"",
		"impl Display for Label {",
		"  fn to_string(self: Self) -> String { self.text }",
		"}",
		"",
		"impl Echo for Box T where T: Display {",
		"  fn pass<U>(self: Self, value: U) -> U { value }",
		"}",
		"",
		"fn main() -> String {",
		"  echo: Echo = Box { value: Label { text: \"seed\" } }",
		"  value := echo.pass<Labelish>(Label { text: \"ok\" })",
		"  print(value.to_string())",
		"  value.to_string()",
		"}",
		"",
	}, "\n")
}

func TestCompilerGenericInterfaceConstraintRevalidationUsesGeneratedMetadata(t *testing.T) {
	result := compileNoFallbackSource(t, genericInterfaceConstraintRevalidationSource())
	compiledSrc := combinedGeneratedSource(result)

	for _, fragment := range []string{
		"func __able_expand_runtime_type_aliases(",
		"func __able_type_expr_satisfies_interface(",
		"var __able_type_alias_defs = map[string]*ast.TypeAliasDefinition{",
		"var __able_interface_method_names = map[string][]string{",
		"var __able_interface_generic_param_names",
		"var __able_known_type_names = map[string]struct{}{",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generated constraint revalidation support to include %q:\n%s", fragment, compiledSrc)
		}
	}

	for _, fragment := range []string{
		"bridge.EnsureTypeSatisfiesInterface(",
		"bridge.IsKnownConstraintTypeName(",
		"bridge.ExpandTypeAliases(__able_runtime, entry.targetType)",
		"bridge.ExpandTypeAliases(__able_runtime, actual)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected no-bootstrap generic interface revalidation to avoid %q:\n%s", fragment, compiledSrc)
		}
	}

	enforceBody, ok := findCompiledFunction(result, "__able_enforce_constraints")
	if !ok {
		t.Fatalf("could not find generated __able_enforce_constraints helper")
	}
	if !strings.Contains(enforceBody, "__able_enforce_constraints_seen(constraints, bindings, make(map[string]struct{}))") {
		t.Fatalf("expected __able_enforce_constraints to delegate to the seen-aware generated helper:\n%s", enforceBody)
	}

	seenBody, ok := findCompiledFunction(result, "__able_enforce_constraints_seen")
	if !ok {
		t.Fatalf("could not find generated __able_enforce_constraints_seen helper")
	}
	for _, fragment := range []string{
		"__able_expand_runtime_type_aliases(",
		"__able_type_expr_satisfies_interface_seen(",
	} {
		if !strings.Contains(seenBody, fragment) {
			t.Fatalf("expected seen-aware constraint helper to include %q:\n%s", fragment, seenBody)
		}
	}
}

func TestCompilerGenericInterfaceConstraintRevalidationExecutesWithoutFallbacks(t *testing.T) {
	stdout := strings.TrimSpace(compileAndRunSourceWithOptions(t, "ablec-generic-iface-constraint-revalidation-", genericInterfaceConstraintRevalidationSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	}))
	if stdout != "ok" {
		t.Fatalf("expected constrained generic interface dispatch program to print ok, got %q", stdout)
	}
}
