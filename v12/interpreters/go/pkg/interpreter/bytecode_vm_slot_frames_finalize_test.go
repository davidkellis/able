package interpreter

import (
	"errors"
	"testing"
)

func TestBytecodeVMReleaseCompletedRunFramesReleasesActiveSlots(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	released := vm.acquireSlotFrame(3)
	vm.slots = released
	vm.releaseCompletedRunFrames()

	if vm.slots != nil {
		t.Fatalf("expected active slots to be cleared after completed-run cleanup")
	}
	reacquired := vm.acquireSlotFrame(3)
	if len(reacquired) != len(released) {
		t.Fatalf("expected slot frame length %d, got %d", len(released), len(reacquired))
	}
	if len(reacquired) == 0 {
		t.Fatalf("expected non-empty slot frame for reuse assertion")
	}
	if &reacquired[0] != &released[0] {
		t.Fatalf("expected completed-run cleanup to return active slot frame to pool")
	}
}

func TestBytecodeVMFinishRunResumableReleasesUnwoundCallFrames(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	callerSlots := vm.acquireSlotFrame(2)
	calleeSlots := vm.acquireSlotFrame(1)
	vm.slots = calleeSlots
	vm.callFrames = append(vm.callFrames, bytecodeCallFrame{
		returnIP: 0,
		slots:    callerSlots,
		env:      env,
	})

	runErr := errors.New("boom")
	vm.finishRunResumable(&runErr)

	if vm.slots != nil {
		t.Fatalf("expected top-level slots to be released after non-yield run exit")
	}
	if len(vm.callFrames) != 0 {
		t.Fatalf("expected inline call frames to be cleared, got %d", len(vm.callFrames))
	}

	reacquiredCallee := vm.acquireSlotFrame(1)
	if len(reacquiredCallee) == 0 || &reacquiredCallee[0] != &calleeSlots[0] {
		t.Fatalf("expected callee slot frame to be returned to pool during unwind")
	}

	reacquiredCaller := vm.acquireSlotFrame(2)
	if len(reacquiredCaller) == 0 || &reacquiredCaller[0] != &callerSlots[0] {
		t.Fatalf("expected caller slot frame to be returned to pool after unwind")
	}
}
