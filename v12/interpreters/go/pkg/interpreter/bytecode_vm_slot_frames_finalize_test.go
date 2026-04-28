package interpreter

import (
	"errors"
	"testing"

	"able/interpreter-go/pkg/runtime"
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
	vm.pushCallFrame(0, nil, callerSlots, env, nil, 0, 0, false, false)

	runErr := errors.New("boom")
	vm.finishRunResumable(&runErr)

	if vm.slots != nil {
		t.Fatalf("expected top-level slots to be released after non-yield run exit")
	}
	if len(vm.callFrames) != 0 {
		t.Fatalf("expected inline call frames to be cleared, got %d", len(vm.callFrames))
	}
	if len(vm.callFrameKinds) != 0 {
		t.Fatalf("expected inline call frame kinds to be cleared, got %d", len(vm.callFrameKinds))
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

func TestBytecodeVMAcquireSlotFramePrefillsHotBatchForSmallLayouts(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	acquired := vm.acquireSlotFrame(3)
	if len(acquired) != 3 {
		t.Fatalf("expected acquired slot frame length 3, got %d", len(acquired))
	}
	if vm.slotFrameHotSize != 3 {
		t.Fatalf("expected hot slot frame size 3, got %d", vm.slotFrameHotSize)
	}
	if len(vm.slotFrameHotPool) != bytecodeSlotFrameBatchSize-1 {
		t.Fatalf("expected hot slot frame pool size %d after batch prefill, got %d", bytecodeSlotFrameBatchSize-1, len(vm.slotFrameHotPool))
	}
}

func TestBytecodeVMAcquireSlotFrame2UsesHotBatch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	acquired := vm.acquireSlotFrame2()
	if len(acquired) != 2 {
		t.Fatalf("expected acquired slot frame length 2, got %d", len(acquired))
	}
	if vm.slotFrameHotSize != 2 {
		t.Fatalf("expected hot slot frame size 2, got %d", vm.slotFrameHotSize)
	}
	if len(vm.slotFrameHotPool) != bytecodeSlotFrameBatchSize-1 {
		t.Fatalf("expected hot slot frame pool size %d after size-2 prefill, got %d", bytecodeSlotFrameBatchSize-1, len(vm.slotFrameHotPool))
	}

	vm.releaseSlotFrame(acquired)
	reacquired := vm.acquireSlotFrame2()
	if len(reacquired) != 2 || &reacquired[0] != &acquired[0] {
		t.Fatalf("expected size-2 frame to round-trip through hot pool")
	}
}

func TestBytecodeVMReleaseSlotFrame2ClearsAndReusesHotPool(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	acquired := vm.acquireSlotFrame2()
	acquired[0] = runtime.StringValue{Val: "left"}
	acquired[1] = runtime.StringValue{Val: "right"}

	vm.releaseSlotFrame2(acquired)
	if acquired[0] != nil || acquired[1] != nil {
		t.Fatalf("expected releaseSlotFrame2 to clear both slots, got %#v", acquired)
	}

	reacquired := vm.acquireSlotFrame2()
	if len(reacquired) != 2 || &reacquired[0] != &acquired[0] {
		t.Fatalf("expected size-2 frame to round-trip through dedicated release path")
	}
}

func TestBytecodeVMAcquireSlotFrameSpillsOldHotBatchOnSizeChange(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	_ = vm.acquireSlotFrame(2)
	if len(vm.slotFrameHotPool) == 0 {
		t.Fatalf("expected initial hot pool prefill for size 2")
	}

	_ = vm.acquireSlotFrame(3)
	if vm.slotFrameHotSize != 3 {
		t.Fatalf("expected hot slot frame size to switch to 3, got %d", vm.slotFrameHotSize)
	}
	if len(vm.slotFrameHotPool) != bytecodeSlotFrameBatchSize-1 {
		t.Fatalf("expected refreshed hot slot frame pool size %d for size 3, got %d", bytecodeSlotFrameBatchSize-1, len(vm.slotFrameHotPool))
	}
	if vm.slotFramePool == nil || len(vm.slotFramePool[2]) == 0 {
		t.Fatalf("expected prior size-2 hot frames to spill into general pool")
	}
}

func TestBytecodeVMFinishRunResumableReleasesUnwoundSelfFastCallFrames(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	callerSlots := vm.acquireSlotFrame(2)
	calleeSlots := vm.acquireSlotFrame(1)
	vm.slots = calleeSlots
	vm.currentProgram = &bytecodeProgram{}
	vm.pushCallFrame(0, vm.currentProgram, callerSlots, env, nil, 0, 0, false, true)

	runErr := errors.New("boom")
	vm.finishRunResumable(&runErr)

	if vm.slots != nil {
		t.Fatalf("expected top-level slots to be released after non-yield run exit")
	}
	if len(vm.selfFastCallFrames) != 0 {
		t.Fatalf("expected self-fast call frames to be cleared, got %d", len(vm.selfFastCallFrames))
	}
	if len(vm.callFrameKinds) != 0 {
		t.Fatalf("expected inline call frame kinds to be cleared, got %d", len(vm.callFrameKinds))
	}

	reacquiredCallee := vm.acquireSlotFrame(1)
	if len(reacquiredCallee) == 0 || &reacquiredCallee[0] != &calleeSlots[0] {
		t.Fatalf("expected callee slot frame to be returned to pool during self-fast unwind")
	}

	reacquiredCaller := vm.acquireSlotFrame(2)
	if len(reacquiredCaller) == 0 || &reacquiredCaller[0] != &callerSlots[0] {
		t.Fatalf("expected caller slot frame to be returned to pool after self-fast unwind")
	}
}

func TestBytecodeVMFinishRunResumableReleasesUnwoundMinimalSelfFastCallFrames(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	callerSlots := vm.acquireSlotFrame(2)
	calleeSlots := vm.acquireSlotFrame(1)
	vm.slots = calleeSlots
	vm.currentProgram = &bytecodeProgram{}
	vm.pushCallFrame(0, vm.currentProgram, callerSlots, env, nil, 0, 0, false, true)

	if len(vm.selfFastMinimal) != 1 {
		t.Fatalf("expected minimal self-fast frame to be used, got %d", len(vm.selfFastMinimal))
	}
	if len(vm.selfFastCallFrames) != 0 {
		t.Fatalf("expected full self-fast frame stack to remain empty, got %d", len(vm.selfFastCallFrames))
	}

	runErr := errors.New("boom")
	vm.finishRunResumable(&runErr)

	if vm.slots != nil {
		t.Fatalf("expected top-level slots to be released after non-yield run exit")
	}
	if len(vm.selfFastMinimal) != 0 {
		t.Fatalf("expected minimal self-fast call frames to be cleared, got %d", len(vm.selfFastMinimal))
	}
	if len(vm.callFrameKinds) != 0 {
		t.Fatalf("expected inline call frame kinds to be cleared, got %d", len(vm.callFrameKinds))
	}

	reacquiredCallee := vm.acquireSlotFrame(1)
	if len(reacquiredCallee) == 0 || &reacquiredCallee[0] != &calleeSlots[0] {
		t.Fatalf("expected callee slot frame to be returned to pool during minimal self-fast unwind")
	}

	reacquiredCaller := vm.acquireSlotFrame(2)
	if len(reacquiredCaller) == 0 || &reacquiredCaller[0] != &callerSlots[0] {
		t.Fatalf("expected caller slot frame to be returned to pool after minimal self-fast unwind")
	}
}

func TestBytecodeVMPushSelfFastMinimalCallFrameUsesMinimalStacks(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	callerSlots := vm.acquireSlotFrame(2)
	vm.pushSelfFastMinimalCallFrame(7, callerSlots)

	if len(vm.callFrameKinds) != 0 {
		t.Fatalf("expected minimal self-fast frame to stay out of callFrameKinds, got %#v", vm.callFrameKinds)
	}
	if len(vm.selfFastMinimal) != 1 {
		t.Fatalf("expected one minimal self-fast frame, got %d", len(vm.selfFastMinimal))
	}
	if vm.selfFastMinimalSuffix != 1 {
		t.Fatalf("expected one unmaterialized minimal self-fast frame, got %d", vm.selfFastMinimalSuffix)
	}
	if len(vm.selfFastCallFrames) != 0 {
		t.Fatalf("expected full self-fast frame stack to remain empty, got %d", len(vm.selfFastCallFrames))
	}
	if vm.selfFastMinimal[0].returnIP != 7 {
		t.Fatalf("expected returnIP=7, got %d", vm.selfFastMinimal[0].returnIP)
	}
	if len(vm.selfFastMinimal[0].slots) == 0 || &vm.selfFastMinimal[0].slots[0] != &callerSlots[0] {
		t.Fatalf("expected minimal self-fast frame to keep caller slot slice")
	}
}

func TestBytecodeVMPushCallFrameMaterializesMinimalSelfFastSuffix(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	callerSlots := vm.acquireSlotFrame(2)
	fullCallerSlots := vm.acquireSlotFrame(3)
	vm.pushSelfFastMinimalCallFrame(7, callerSlots)
	vm.pushCallFrame(9, nil, fullCallerSlots, env, nil, 0, 0, false, false)

	if vm.selfFastMinimalSuffix != 0 {
		t.Fatalf("expected minimal self-fast suffix to materialize before full frame push, got %d", vm.selfFastMinimalSuffix)
	}
	if len(vm.callFrameKinds) != 2 {
		t.Fatalf("expected materialized minimal kind plus full kind, got %d", len(vm.callFrameKinds))
	}
	if vm.callFrameKinds[0] != bytecodeCallFrameKindSelfFastMinimal || vm.callFrameKinds[1] != bytecodeCallFrameKindFull {
		t.Fatalf("unexpected materialized call frame kinds: %#v", vm.callFrameKinds)
	}
	if len(vm.selfFastMinimal) != 1 {
		t.Fatalf("expected minimal self-fast frame payload to remain available, got %d", len(vm.selfFastMinimal))
	}
	if len(vm.callFrames) != 1 {
		t.Fatalf("expected one full call frame, got %d", len(vm.callFrames))
	}
}
