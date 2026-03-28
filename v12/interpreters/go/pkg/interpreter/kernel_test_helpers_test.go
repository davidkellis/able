package interpreter

import (
	"sort"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func loadKernelModule(t *testing.T, interp *Interpreter) {
	t.Helper()
	mod, err := parseSourceModule(kernelEntry)
	if err != nil {
		t.Fatalf("load kernel: %v", err)
	}
	mod.Package = ast.Pkg([]interface{}{"able", "kernel"}, false)
	if _, _, err := interp.EvaluateModule(mod); err != nil {
		t.Fatalf("evaluate kernel: %v", err)
	}
	if _, ok := interp.packageRegistry["able.kernel"]; !ok {
		keys := make([]string, 0, len(interp.packageRegistry))
		for key := range interp.packageRegistry {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		t.Fatalf("kernel package missing (packages: %v)", keys)
	}
	val, err := interp.GlobalEnvironment().Get("able.kernel.KernelHasher.new")
	if err != nil {
		t.Fatalf("kernel missing KernelHasher.new: %v", err)
	}
	if overload, ok := val.(*runtime.FunctionOverloadValue); ok {
		t.Fatalf("kernel KernelHasher.new overloaded (%d entries)", len(overload.Overloads))
	}
}

func TestNewKernelHasherFallsBackToPackageRegistry(t *testing.T) {
	interp := New()
	loadKernelModule(t, interp)

	if _, err := interp.GlobalEnvironment().Get("KernelHasher.new"); err == nil {
		t.Fatalf("expected unqualified KernelHasher.new to be absent from the global environment for this regression")
	}

	value, err := interp.newKernelHasher()
	if err != nil {
		t.Fatalf("newKernelHasher should resolve through package registry fallback: %v", err)
	}
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		t.Fatalf("expected KernelHasher struct instance, got %T", value)
	}
	if name := structInstanceName(inst); name != "KernelHasher" {
		t.Fatalf("expected KernelHasher instance, got %q", name)
	}
}
