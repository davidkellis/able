package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInterfaceDispatchMemberMatchResolution(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_interface_dispatch_match_entry(receiver runtime.Value, receiverTypeName string, entry __able_interface_dispatch_entry) (runtime.Value, bool, error) {") {
		t.Fatalf("expected shared interface dispatch match helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_interface_dispatch_member(")
	if start < 0 {
		t.Fatalf("expected __able_interface_dispatch_member helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "\tif len(matches) == 0 {")
	if end < 0 {
		t.Fatalf("expected __able_interface_dispatch_member match segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "if receiverTypeName != \"\" {") {
		t.Fatalf("expected inline receiverTypeName match branch to be removed from __able_interface_dispatch_member")
	}
	if strings.Contains(segment, "if len(entry.genericNames) > 0 {") {
		t.Fatalf("expected inline generic match branch to be removed from __able_interface_dispatch_member")
	}
	if strings.Contains(segment, "coerced, ok, err := bridge.MatchType(__able_runtime, entry.targetType, receiver)") {
		t.Fatalf("expected inline bridge.MatchType branch to be removed from __able_interface_dispatch_member")
	}
	if !strings.Contains(segment, "if matchedReceiver, ok, err := __able_interface_dispatch_match_entry(receiver, receiverTypeName, entry); err != nil {") {
		t.Fatalf("expected __able_interface_dispatch_member to use shared match helper path")
	}
}

