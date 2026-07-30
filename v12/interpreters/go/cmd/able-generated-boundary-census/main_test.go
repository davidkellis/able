package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeSeparatesCompiledBodiesFromRuntimeWrappers(t *testing.T) {
	dir := t.TempDir()
	source := `package main

import (
	"able/interpreter-go/pkg/compiler/bridge"
	"able/interpreter-go/pkg/runtime"
)

type Record struct { Value int64 }
type __able_iface_Policy interface{}

func __able_compiled_fn_main(value runtime.Value) {
	_ = bridge.ToInt(1, runtime.IntegerType("i64"))
	_, _ = __able_struct_Record_from(value)
	_ = __able_iface_Policy_wrap_ptr_Record(&Record{Value: 1})
	_, _ = __able_method_call_node(value, "run", nil, nil)
}

func __able_compiled_entry_fn_main(value runtime.Value) {
	__able_compiled_fn_main(value)
}

func __able_wrap_fn_main(value runtime.Value) {
	__able_compiled_entry_fn_main(value)
}
`
	path := filepath.Join(dir, "compiled.go")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := analyze(dir)
	if err != nil {
		t.Fatal(err)
	}
	compiled := result.Scopes["compiled_body"]
	if compiled.Functions != 1 {
		t.Fatalf("compiled functions = %d, want 1", compiled.Functions)
	}
	if compiled.RuntimeValueTypes != 1 {
		t.Fatalf("compiled runtime.Value sites = %d, want 1", compiled.RuntimeValueTypes)
	}
	for category, want := range map[string]int{
		"bridge_encode":             1,
		"struct_runtime_conversion": 1,
		"native_interface_adapter":  1,
		"erased_or_dynamic_call":    1,
	} {
		if got := compiled.BoundaryCategories[category]; got != want {
			t.Fatalf("%s sites = %d, want %d", category, got, want)
		}
	}
	if got := compiled.HeapNominalLiterals["Record"]; got != 1 {
		t.Fatalf("Record heap literals = %d, want 1", got)
	}
	if got := result.Scopes["entry_wrapper"].Functions; got != 1 {
		t.Fatalf("entry wrappers = %d, want 1", got)
	}
	if got := result.Scopes["runtime_wrapper"].Functions; got != 1 {
		t.Fatalf("runtime wrappers = %d, want 1", got)
	}
	if got := result.Scopes["main_direct_reachable"].Functions; got != 1 {
		t.Fatalf("main-direct reachable functions = %d, want 1", got)
	}
}
