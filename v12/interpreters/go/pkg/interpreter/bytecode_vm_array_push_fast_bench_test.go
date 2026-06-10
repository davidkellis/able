package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var benchmarkBytecodeArrayPushProgramSink *bytecodeProgram
var benchmarkBytecodeArrayPushBoolSink bool
var benchmarkBytecodeArrayPushErrorSource error
var benchmarkBytecodeArrayPushIntSink int
var benchmarkBytecodeArrayPushModeSource = "array_write_slot_tracked_fast"
var benchmarkBytecodeArrayPushValueSink runtime.Value

func benchmarkMonoCharArrayWithCapacity(b *testing.B, capacity int) *runtime.ArrayValue {
	b.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityChar(capacity)
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func benchmarkArrayPushVM(b *testing.B, includePop bool) (*bytecodeVM, *bytecodeProgram, *bytecodeInstruction, *runtime.ArrayValue) {
	b.Helper()
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instructions := []bytecodeInstruction{
		{
			op:             bytecodeOpCallMemberArraySlot,
			name:           "push",
			argCount:       1,
			memberFastPath: bytecodeMemberMethodFastPathArrayPush,
		},
	}
	if includePop {
		instructions = append(instructions, bytecodeInstruction{op: bytecodeOpPop})
	}
	program := &bytecodeProgram{instructions: instructions}
	vm.currentProgram = program
	instr := &program.instructions[0]
	arr := benchmarkMonoCharArrayWithCapacity(b, b.N+1)
	vm.storeCachedCanonicalArraySlotCallForArray(program, 0, arr, bytecodeMemberMethodFastPathArrayPush)
	return vm, program, instr, arr
}

func benchmarkTrackedArraySlotVM(b *testing.B) (*bytecodeVM, *runtime.ArrayValue, runtime.Value) {
	b.Helper()
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(11, runtime.IntegerI32),
	}, 1)
	return vm, arr, runtime.NewSmallInt(0, runtime.IntegerI32)
}

func benchmarkArraySlotCallVM(b *testing.B, name string, argCount int, kind bytecodeMemberMethodFastPathKind) (*bytecodeVM, *bytecodeProgram, *bytecodeInstruction, *runtime.ArrayValue, runtime.Value, runtime.Value) {
	b.Helper()
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{
			op:             bytecodeOpCallMemberArraySlot,
			name:           name,
			argCount:       argCount,
			memberFastPath: kind,
		},
	}}
	vm.currentProgram = program
	instr := &program.instructions[0]
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(11, runtime.IntegerI32),
	}, 1)
	indexVal := runtime.NewSmallInt(0, runtime.IntegerI32)
	writeVal := runtime.NewSmallInt(17, runtime.IntegerI32)
	vm.storeCachedCanonicalArraySlotCallForArray(program, 0, arr, kind)
	return vm, program, instr, arr, indexVal, writeVal
}

func BenchmarkBytecodeVMArrayPushMemberFast(b *testing.B) {
	b.Run("runtime_append_char_if_mono", func(b *testing.B) {
		arr := benchmarkMonoCharArrayWithCapacity(b, b.N+1)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			ok, err := runtime.ArrayStoreAppendCharIfMono(arr.Handle, rune('a'+idx%26))
			if err != nil || !ok {
				b.Fatalf("ArrayStoreAppendCharIfMono = (%v, %v), want ok/nil", ok, err)
			}
			benchmarkBytecodeArrayPushBoolSink = ok
		}
	})

	b.Run("append_array_char_value_fast", func(b *testing.B) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
		arr := benchmarkMonoCharArrayWithCapacity(b, b.N+1)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeArrayPushBoolSink = vm.appendArrayCharValueFast(arr, rune('a'+idx%26))
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("appendArrayCharValueFast returned false")
			}
		}
	})

	b.Run("exec_push_member_fast_pop", func(b *testing.B) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
		program := &bytecodeProgram{instructions: []bytecodeInstruction{
			{op: bytecodeOpCallMember, name: "push", argCount: 1},
			{op: bytecodeOpPop},
		}}
		vm.currentProgram = program
		instr := &program.instructions[0]
		arr := benchmarkMonoCharArrayWithCapacity(b, b.N+1)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, runtime.CharValue{Val: rune('a' + idx%26)})
			newProg, handled, err := vm.execArrayPushMemberFast(instr.name, instr.argCount, instr.node, 0, 1, nil)
			if err != nil || !handled {
				b.Fatalf("execArrayPushMemberFast = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
		}
	})

	b.Run("finish_array_read_slot_member_fast_tracked", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		instr := bytecodeInstruction{name: "read_slot", argCount: 1}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal)
			newProg, handled, err := vm.finishArrayReadSlotMemberFast(instr.name, instr.argCount, instr.node, arr, 0, 1, nil)
			if err != nil || !handled {
				b.Fatalf("finishArrayReadSlotMemberFast = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("read_array_slot_value_fast_tracked_ignore_mode", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			result, _, handled, err := vm.readArraySlotValueFast(arr, indexVal)
			if err != nil || !handled {
				b.Fatalf("readArraySlotValueFast = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushValueSink = result
		}
	})

	b.Run("read_array_slot_value_fast_checked_tracked_ignore_mode", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			result, _, handled, err := vm.readArraySlotValueFastChecked(arr, indexVal)
			if err != nil || !handled {
				b.Fatalf("readArraySlotValueFastChecked = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushValueSink = result
		}
	})

	b.Run("read_shell_guard_stack_load", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		argCount := 1
		receiverIndex := 0
		argBase := 1

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = append(vm.stack[:0], arr, indexVal)
			ok := !(vm == nil || arr == nil || argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack))
			if !ok {
				b.Fatalf("read guard rejected benchmark inputs")
			}
			benchmarkBytecodeArrayPushValueSink = vm.stack[argBase]
		}
	})

	b.Run("read_shell_tracked_state_lookup", func(b *testing.B) {
		_, arr, indexVal := benchmarkTrackedArraySlotVM(b)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			state, tracked := bytecodeTrackedArrayState(arr)
			if !tracked {
				b.Fatalf("benchmark array is not tracked")
			}
			slot, ok := arraySlotIndexSmall(indexVal)
			if !ok || slot >= len(state.Values) {
				b.Fatalf("tracked lookup missed benchmark slot")
			}
			result := state.Values[slot]
			if result == nil {
				result = runtime.NilValue{}
			}
			benchmarkBytecodeArrayPushValueSink = result
		}
	})

	b.Run("read_shell_result_completion", func(b *testing.B) {
		vm, arr, _ := benchmarkTrackedArraySlotVM(b)
		result := runtime.Value(runtime.NewSmallInt(11, runtime.IntegerI32))
		receiverIndex := 0

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr)
			vm.stack = vm.stack[:receiverIndex]
			vm.stack = append(vm.stack, result)
			vm.ip++
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("read_shell_guard_lookup_result_completion", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		argCount := 1
		receiverIndex := 0
		argBase := 1

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal)
			if vm == nil || arr == nil || argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
				b.Fatalf("read guard rejected benchmark inputs")
			}
			loadedIndex := vm.stack[argBase]
			state, tracked := bytecodeTrackedArrayState(arr)
			if !tracked {
				b.Fatalf("benchmark array is not tracked")
			}
			slot, ok := arraySlotIndexSmall(loadedIndex)
			if !ok || slot >= len(state.Values) {
				b.Fatalf("tracked lookup missed benchmark slot")
			}
			result := state.Values[slot]
			if result == nil {
				result = runtime.NilValue{}
			}
			vm.stack = vm.stack[:receiverIndex]
			vm.stack = append(vm.stack, result)
			vm.ip++
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("finish_array_write_slot_member_fast_tracked", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		instr := bytecodeInstruction{name: "write_slot", argCount: 2}
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, value)
			newProg, handled, err := vm.finishArrayWriteSlotMemberFast(instr.name, instr.argCount, instr.node, arr, 0, 1, nil)
			if err != nil || !handled {
				b.Fatalf("finishArrayWriteSlotMemberFast = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("write_array_slot_value_fast_tracked", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			mode, handled, err := vm.writeArraySlotValueFast(arr, indexVal, value)
			if err != nil || !handled || mode != "array_write_slot_tracked_fast" {
				b.Fatalf("writeArraySlotValueFast = (%q, %v, %v), want tracked/handled/nil", mode, handled, err)
			}
			benchmarkBytecodeArrayPushBoolSink = handled
		}
	})

	b.Run("write_array_slot_value_fast_tracked_ignore_mode", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			_, handled, err := vm.writeArraySlotValueFast(arr, indexVal, value)
			if err != nil || !handled {
				b.Fatalf("writeArraySlotValueFast = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushBoolSink = handled
		}
	})

	b.Run("write_array_slot_value_fast_checked_tracked_ignore_mode", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			_, handled, err := vm.writeArraySlotValueFastChecked(arr, indexVal, value)
			if err != nil || !handled {
				b.Fatalf("writeArraySlotValueFastChecked = (%v, %v), want handled/nil", handled, err)
			}
			benchmarkBytecodeArrayPushBoolSink = handled
		}
	})

	b.Run("write_shell_guard_stack_load", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)
		argCount := 2
		receiverIndex := 0
		argBase := 1

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = append(vm.stack[:0], arr, indexVal, value)
			ok := !(vm == nil || arr == nil || argCount != 2 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase+1 >= len(vm.stack) || vm.interp == nil)
			if !ok {
				b.Fatalf("write guard rejected benchmark inputs")
			}
			loadedIndex := vm.stack[argBase]
			loadedValue := vm.stack[argBase+1]
			benchmarkBytecodeArrayPushBoolSink = loadedIndex == indexVal
			benchmarkBytecodeArrayPushValueSink = loadedValue
		}
	})

	b.Run("write_shell_no_error_no_fallback_fast_void", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)
		receiverIndex := 0
		mode := benchmarkBytecodeArrayPushModeSource
		handled := true

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, value)
			err := benchmarkBytecodeArrayPushErrorSource
			if err != nil {
				vm.stack = vm.stack[:receiverIndex]
				newProg, finishErr := vm.finishCompletedCall(nil, err, nil, nil)
				if finishErr != nil {
					b.Fatalf("finishCompletedCall(error) failed: %v", finishErr)
				}
				benchmarkBytecodeArrayPushProgramSink = newProg
				continue
			}
			if !handled {
				b.Fatalf("write shell unexpectedly fell back")
			}
			if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
				vm.interp.recordBytecodeCallTrace("call_member", "write_slot", "resolved_method", mode, nil)
			}
			vm.stack = vm.stack[:receiverIndex]
			newProg, finishErr := vm.finishCompletedVoidCallFast()
			if finishErr != nil {
				b.Fatalf("finishCompletedVoidCallFast failed: %v", finishErr)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("write_shell_guard_stack_branch_fast_void", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)
		argCount := 2
		receiverIndex := 0
		argBase := 1
		mode := benchmarkBytecodeArrayPushModeSource
		handled := true

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, value)
			if vm == nil || arr == nil || argCount != 2 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase+1 >= len(vm.stack) || vm.interp == nil {
				b.Fatalf("write guard rejected benchmark inputs")
			}
			benchmarkBytecodeArrayPushValueSink = vm.stack[argBase+1]
			err := benchmarkBytecodeArrayPushErrorSource
			if err != nil {
				vm.stack = vm.stack[:receiverIndex]
				newProg, finishErr := vm.finishCompletedCall(nil, err, nil, nil)
				if finishErr != nil {
					b.Fatalf("finishCompletedCall(error) failed: %v", finishErr)
				}
				benchmarkBytecodeArrayPushProgramSink = newProg
				continue
			}
			if !handled {
				b.Fatalf("write shell unexpectedly fell back")
			}
			if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
				vm.interp.recordBytecodeCallTrace("call_member", "write_slot", "resolved_method", mode, nil)
			}
			vm.stack = vm.stack[:receiverIndex]
			newProg, finishErr := vm.finishCompletedVoidCallFast()
			if finishErr != nil {
				b.Fatalf("finishCompletedVoidCallFast failed: %v", finishErr)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("finish_completed_void_call", func(b *testing.B) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
		var void runtime.Value = runtime.VoidValue{}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = vm.stack[:0]
			newProg, err := vm.finishCompletedCall(void, nil, nil, nil)
			if err != nil {
				b.Fatalf("finishCompletedCall(void) failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("write_shell_trace_disabled_branch", func(b *testing.B) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
		mode := "array_write_slot_tracked_fast"

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
				vm.interp.recordBytecodeCallTrace("call_member", "write_slot", "resolved_method", mode, nil)
			}
			benchmarkBytecodeArrayPushBoolSink = interp.bytecodeTraceEnabled
		}
	})

	b.Run("write_shell_stack_truncate", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)
		vm.stack = append(vm.stack[:0], arr, indexVal, value)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = vm.stack[:3]
			vm.stack = vm.stack[:0]
			benchmarkBytecodeArrayPushIntSink = len(vm.stack)
		}
	})

	b.Run("write_shell_truncate_finish_void", func(b *testing.B) {
		vm, arr, indexVal := benchmarkTrackedArraySlotVM(b)
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)
		var void runtime.Value = runtime.VoidValue{}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, value)
			vm.stack = vm.stack[:0]
			newProg, err := vm.finishCompletedCall(void, nil, nil, nil)
			if err != nil {
				b.Fatalf("finishCompletedCall(void) failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("finish_completed_void_call_fast", func(b *testing.B) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = vm.stack[:0]
			if _, err := vm.finishCompletedVoidCallFast(); err != nil {
				b.Fatalf("finishCompletedVoidCallFast failed: %v", err)
			}
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("array_slot_index_small", func(b *testing.B) {
		_, _, indexVal := benchmarkTrackedArraySlotVM(b)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			slot, ok := arraySlotIndexSmall(indexVal)
			if !ok || slot != 0 {
				b.Fatalf("arraySlotIndexSmall = (%d, %v), want 0/true", slot, ok)
			}
			benchmarkBytecodeArrayPushIntSink = slot
		}
	})

	b.Run("raw_tracked_state_replacement", func(b *testing.B) {
		_, arr, _ := benchmarkTrackedArraySlotVM(b)
		state, tracked := bytecodeTrackedArrayState(arr)
		if !tracked {
			b.Fatalf("benchmark array is not tracked")
		}
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			state.Values[0] = value
			benchmarkBytecodeArrayPushValueSink = state.Values[0]
		}
	})

	b.Run("sync_unaliased_tracked_array_write", func(b *testing.B) {
		_, arr, _ := benchmarkTrackedArraySlotVM(b)
		state, tracked := bytecodeTrackedArrayState(arr)
		if !tracked {
			b.Fatalf("benchmark array is not tracked")
		}
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, 0, value) {
				b.Fatalf("bytecodeSyncUnaliasedTrackedArrayWrite returned false")
			}
			benchmarkBytecodeArrayPushBoolSink = true
		}
	})

	b.Run("update_array_element_type_token_i32_write", func(b *testing.B) {
		_, arr, _ := benchmarkTrackedArraySlotVM(b)
		state, tracked := bytecodeTrackedArrayState(arr)
		if !tracked {
			b.Fatalf("benchmark array is not tracked")
		}
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			updateArrayElementTypeTokenForWrite(state, 0, value)
			benchmarkBytecodeArrayPushIntSink = int(state.Revision)
		}
	})

	b.Run("array_state_write_keeps_materialized_values", func(b *testing.B) {
		_, arr, _ := benchmarkTrackedArraySlotVM(b)
		state, tracked := bytecodeTrackedArrayState(arr)
		if !tracked {
			b.Fatalf("benchmark array is not tracked")
		}
		var value runtime.Value = runtime.NewSmallInt(17, runtime.IntegerI32)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			arrayStateWriteKeepsMaterializedValues(state, value)
			benchmarkBytecodeArrayPushBoolSink = state.ValuesMaterialized
		}
	})

	b.Run("cache_lookup_direct_push", func(b *testing.B) {
		vm, program, _, arr := benchmarkArrayPushVM(b, true)
		_ = arr

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeArrayPushBoolSink = vm.lookupCachedCanonicalArraySlotCallForArrayValidated(
				program,
				0,
				bytecodeMemberMethodFastPathArrayPush,
			)
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("lookupCachedCanonicalArraySlotCallForArrayValidated returned false at %d", idx)
			}
		}
	})

	b.Run("can_skip_following_pop", func(b *testing.B) {
		vm, _, _, arr := benchmarkArrayPushVM(b, true)
		_ = arr

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			benchmarkBytecodeArrayPushBoolSink = vm.canSkipFollowingPop(nil)
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("canSkipFollowingPop returned false at %d", idx)
			}
		}
	})

	b.Run("exec_call_member_array_slot_push_pop", func(b *testing.B) {
		vm, program, instr, arr := benchmarkArrayPushVM(b, true)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, runtime.CharValue{Val: rune('a' + idx%26)})
			newProg, err := vm.execCallMemberArraySlot(instr, program)
			if err != nil {
				b.Fatalf("execCallMemberArraySlot push+pop failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
		}
	})

	b.Run("exec_call_opcode_array_slot_push_pop", func(b *testing.B) {
		vm, program, _, arr := benchmarkArrayPushVM(b, true)
		instr := &program.instructions[0]

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, runtime.CharValue{Val: rune('a' + idx%26)})
			newProg, err := vm.execCallOpcode(instr, nil, program)
			if err != nil {
				b.Fatalf("execCallOpcode array-slot push+pop failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
		}
	})

	b.Run("exec_call_member_array_slot_push_no_pop", func(b *testing.B) {
		vm, program, instr, arr := benchmarkArrayPushVM(b, false)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, runtime.CharValue{Val: rune('a' + idx%26)})
			newProg, err := vm.execCallMemberArraySlot(instr, program)
			if err != nil {
				b.Fatalf("execCallMemberArraySlot push failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
		}
	})
}
