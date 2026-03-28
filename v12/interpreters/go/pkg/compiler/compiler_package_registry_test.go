package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticMethodRegistrarSeedsPackageRegistry(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_02_stdlib_array_helpers")

	var methodImplSrc string
	for name, src := range result.Files {
		if strings.HasPrefix(name, "compiled_pkg_methods_impls_able_kernel_") {
			methodImplSrc = string(src)
			break
		}
	}
	if methodImplSrc == "" {
		t.Fatalf("expected compiled kernel package method impl output")
	}
	if !strings.Contains(methodImplSrc, `interp.RegisterPackageSymbol("able.kernel", "KernelHasher.new", entry.fn)`) {
		t.Fatalf("expected static method registrar to seed package registry for KernelHasher.new:\n%s", methodImplSrc)
	}
}
