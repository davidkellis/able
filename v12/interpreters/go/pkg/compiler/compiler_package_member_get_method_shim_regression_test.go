package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesPackagePublicMemberGetMethodShim(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = 1",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "case runtime.PackageValue:") {
		t.Fatalf("expected value-form PackageValue member_get_method fast path to remain for strict lookup bypass")
	}
	if strings.Contains(compiledSrc, "case *runtime.PackageValue:") {
		t.Fatalf("expected pointer-form PackageValue member_get_method shim branch to be removed")
	}
	if !strings.Contains(compiledSrc, "val, err := bridge.MemberGetPreferMethods(__able_runtime, obj, member)") {
		t.Fatalf("expected member_get_method to continue using bridge.MemberGetPreferMethods path")
	}
}
