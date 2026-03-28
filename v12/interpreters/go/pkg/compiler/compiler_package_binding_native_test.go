package compiler

import (
	"strings"
	"testing"
)

func TestCompilerInterfaceDefinitionModuleBindingAvoidsInstanceCoercion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface KernelEq for T {",
		"  fn eq(self: Self, other: Self) -> bool",
		"}",
		"",
		"type Eq = KernelEq;",
		"EqAlias := KernelEq",
		"",
		"fn main() -> bool { true }",
		"",
	}, "\n"))

	initSrc, ok := result.Files["compiled_package_inits.go"]
	if !ok {
		t.Fatalf("expected compiled package init output")
	}
	initText := string(initSrc)
	for _, fragment := range []string{
		"__able_iface_Eq_from_value(",
		"__able_iface_Eq_to_runtime_value(",
		"__able_iface_KernelEq_from_value(",
		"__able_iface_KernelEq_to_runtime_value(",
	} {
		if strings.Contains(initText, fragment) {
			t.Fatalf("expected interface definition module binding to avoid %q:\n%s", fragment, initText)
		}
	}
	if !strings.Contains(initText, "__able_env_set(\"EqAlias\", __able_env_get(\"KernelEq\"") {
		t.Fatalf("expected interface definition module binding to stay on direct runtime symbol assignment:\n%s", initText)
	}
}

func TestCompilerImportedInterfaceDefinitionModuleBindingAvoidsInstanceCoercion(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.core.interfaces",
			"",
			"fn main() -> bool {",
			"  true",
			"}",
			"",
		}, "\n"),
		"kernel/module.able": strings.Join([]string{
			"interface Eq for T {",
			"  fn eq(self: Self, other: Self) -> bool",
			"}",
			"",
		}, "\n"),
		"core/interfaces/module.able": strings.Join([]string{
			"import demo.kernel.{Eq::KernelEq}",
			"",
			"type Eq = KernelEq;",
			"Eq := KernelEq",
			"",
		}, "\n"),
	})

	initSrc, ok := result.Files["compiled_package_inits.go"]
	if !ok {
		t.Fatalf("expected compiled package init output")
	}
	initText := string(initSrc)
	for _, fragment := range []string{
		"__able_iface_Eq_from_value(",
		"__able_iface_Eq_to_runtime_value(",
	} {
		if strings.Contains(initText, fragment) {
			t.Fatalf("expected imported interface definition module binding to avoid %q:\n%s", fragment, initText)
		}
	}
	if !strings.Contains(initText, "__able_env_set(\"Eq\", __able_env_get(\"KernelEq\"") {
		t.Fatalf("expected imported interface definition module binding to stay on direct runtime symbol assignment:\n%s", initText)
	}
}
