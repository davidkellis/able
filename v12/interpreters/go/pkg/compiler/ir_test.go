package compiler

import "testing"

func TestIRBlockRejectsLateInstructions(t *testing.T) {
	block := &IRBlock{Label: "entry"}
	if err := block.Append(&IRNoop{}); err != nil {
		t.Fatalf("append noop: %v", err)
	}
	if err := block.SetTerminator(&IRReturn{}); err != nil {
		t.Fatalf("set terminator: %v", err)
	}
	if err := block.Append(&IRNoop{}); err == nil {
		t.Fatalf("expected error when appending after terminator")
	}
}

func TestIRBlockRejectsMultipleTerminators(t *testing.T) {
	block := &IRBlock{Label: "entry"}
	if err := block.SetTerminator(&IRReturn{}); err != nil {
		t.Fatalf("set terminator: %v", err)
	}
	if err := block.SetTerminator(&IRJump{Target: "next"}); err == nil {
		t.Fatalf("expected error when setting second terminator")
	}
}
