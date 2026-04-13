package compiler

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCompilerArrayHelperFixtureIndexErrorStaysBoundaryClean(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	runCompilerNoBootstrapBoundaryAuditFixture(t, root, "06_12_02_stdlib_array_helpers")
}

func TestCompilerStructErrorPayloadHelpersUseSharedStructInstanceUnwrap(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_02_stdlib_array_helpers")

	tryFrom, ok := findCompiledFunction(result, "__able_struct_IndexError_try_from")
	if !ok {
		t.Fatalf("could not find __able_struct_IndexError_try_from")
	}
	if !strings.Contains(tryFrom, "inst := __able_struct_instance(current)") {
		t.Fatalf("expected IndexError try_from helper to unwrap error payloads through __able_struct_instance:\n%s", tryFrom)
	}
	if strings.Contains(tryFrom, "current.(*runtime.StructInstanceValue)") {
		t.Fatalf("expected IndexError try_from helper to avoid direct raw struct assertions:\n%s", tryFrom)
	}

	from, ok := findCompiledFunction(result, "__able_struct_IndexError_from")
	if !ok {
		t.Fatalf("could not find __able_struct_IndexError_from")
	}
	if !strings.Contains(from, "inst := __able_struct_instance(current)") {
		t.Fatalf("expected IndexError from helper to unwrap error payloads through __able_struct_instance:\n%s", from)
	}
	if strings.Contains(from, "current.(*runtime.StructInstanceValue)") {
		t.Fatalf("expected IndexError from helper to avoid direct raw struct assertions:\n%s", from)
	}
}
