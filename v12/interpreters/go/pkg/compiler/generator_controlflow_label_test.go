package compiler

import "testing"

func TestLinesReferenceLabelRequiresExactLabelName(t *testing.T) {
	lines := []string{
		"break __able_tmp_12",
		"if done { continue __able_tmp_12 }",
	}
	if linesReferenceLabel(lines, "__able_tmp_1") {
		t.Fatal("temp-label prefix must not count as a reference")
	}
	if !linesReferenceLabel(lines, "__able_tmp_12") {
		t.Fatal("exact break/continue label references must be recognized")
	}
}

func TestLinesReferenceLabelAcceptsDelimitedReference(t *testing.T) {
	lines := []string{
		"if failed { result = runtime.VoidValue{}; break __able_tmp_1 }",
		"continue __able_tmp_1 // resume",
	}
	if !linesReferenceLabel(lines, "__able_tmp_1") {
		t.Fatal("delimited label references must be recognized")
	}
}
