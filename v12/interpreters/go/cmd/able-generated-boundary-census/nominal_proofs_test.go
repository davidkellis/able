package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeNominalProofsFailClosed(t *testing.T) {
	dir := t.TempDir()
	source := `package main

type Safe struct { Value int64 }
type Mutated struct { Count int64 }
type Exposed struct { Code int64 }
type Compared struct { ID int64 }
type UnknownCall struct { Item int64 }
type OpaqueHandle struct { Handle int64 }
type Nested struct { Value *Safe }
type __able_union__Safe_or_error struct { Value *Safe }

func __able_union__Safe_or_error_wrap_ptr_Safe(value *Safe) __able_union__Safe_or_error {
	return __able_union__Safe_or_error{Value: value}
}

func __able_struct_Exposed_to(value *Exposed) any { return value }
func sink(value *UnknownCall) {}

func __able_compiled_fn_main() {
	safe := &Safe{Value: 1}
	_ = __able_union__Safe_or_error_wrap_ptr_Safe(safe)
	mutated := &Mutated{Count: 1}
	mutated.Count = 2
	exposed := &Exposed{Code: 1}
	_ = __able_struct_Exposed_to(exposed)
	compared := &Compared{ID: 1}
	_ = compared == compared
	unknown := &UnknownCall{Item: 1}
	callback := sink
	callback(unknown)
	handle := &OpaqueHandle{Handle: 1}
	external(handle.Handle)
	_ = &Nested{Value: safe}
}

var external = func(value int64) {}
`
	path := filepath.Join(dir, "compiled.go")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := analyze(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.NominalProofs["Safe"].Eligible {
		t.Fatalf("safe proof = %#v", result.NominalProofs["Safe"])
	}
	for name, blocker := range map[string]string{
		"Mutated":      "reachable-field-mutation",
		"Exposed":      "runtime-or-host-identity-exposure",
		"Compared":     "pointer-identity-observation",
		"UnknownCall":  "unknown-mutation-capable-call",
		"OpaqueHandle": "opaque-field-boundary",
		"Nested":       "non-primitive-field-carrier",
	} {
		proof := result.NominalProofs[name]
		if proof.Eligible || !containsString(proof.Blockers, blocker) {
			t.Fatalf("%s proof = %#v, want blocker %q", name, proof, blocker)
		}
	}
	sites := result.NominalProofs["UnknownCall"].UnknownCallSites
	if len(sites) != 1 {
		t.Fatalf("unknown call sites = %#v, want one", sites)
	}
	site := sites[0]
	if site.Caller != "__able_compiled_fn_main" ||
		site.Callee != "callback" ||
		site.File != "compiled.go" ||
		len(site.ArgumentIndexes) != 1 ||
		site.ArgumentIndexes[0] != 0 {
		t.Fatalf("unknown call site = %#v", site)
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
