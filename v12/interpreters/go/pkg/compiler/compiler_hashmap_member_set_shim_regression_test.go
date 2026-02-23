package compiler

import (
	"strings"
	"testing"
)

func TestCompilerMemberSetHashMapHandleUsesSetterBranch(t *testing.T) {
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

	if strings.Contains(compiledSrc, "val, ok := inst.Fields[\"handle\"]") {
		t.Fatalf("expected legacy HashMap.handle member_set read shim branch to be removed")
	}
	if !strings.Contains(compiledSrc, "inst.Fields[\"handle\"] = value") {
		t.Fatalf("expected HashMap.handle member_set setter assignment branch to remain")
	}
	if !strings.Contains(compiledSrc, "hash map handle must be positive") {
		t.Fatalf("expected HashMap.handle setter validation branch to remain")
	}
}
