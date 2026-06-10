package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func BenchmarkBytecodeVMArraySlotCallCacheFast(b *testing.B) {
	b.Run("array_slot_kind_selection_read", func(b *testing.B) {
		_, _, instr, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
			if !ok || kind != bytecodeMemberMethodFastPathArrayReadSlot {
				b.Fatalf("read_slot kind = (%v, %v), want read/true", kind, ok)
			}
			benchmarkBytecodeArrayPushIntSink = int(kind)
		}
	})

	b.Run("array_slot_kind_selection_write", func(b *testing.B) {
		_, _, instr, _, _, _ := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
			if !ok || kind != bytecodeMemberMethodFastPathArrayWriteSlot {
				b.Fatalf("write_slot kind = (%v, %v), want write/true", kind, ok)
			}
			benchmarkBytecodeArrayPushIntSink = int(kind)
		}
	})

	b.Run("array_slot_receiver_validation_read", func(b *testing.B) {
		_, _, instr, arr, indexVal, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)
		stack := []runtime.Value{arr, indexVal}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			if len(stack) < instr.argCount+1 {
				b.Fatalf("benchmark stack underflow")
			}
			receiverIndex := len(stack) - instr.argCount - 1
			receiver := stack[receiverIndex]
			loadedArr, ok := receiver.(*runtime.ArrayValue)
			if !ok || loadedArr == nil {
				b.Fatalf("receiver validation failed")
			}
			benchmarkBytecodeArrayPushValueSink = loadedArr
			benchmarkBytecodeArrayPushIntSink = receiverIndex
		}
	})

	b.Run("array_slot_receiver_validation_write", func(b *testing.B) {
		_, _, instr, arr, indexVal, writeVal := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)
		stack := []runtime.Value{arr, indexVal, writeVal}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			if len(stack) < instr.argCount+1 {
				b.Fatalf("benchmark stack underflow")
			}
			receiverIndex := len(stack) - instr.argCount - 1
			receiver := stack[receiverIndex]
			loadedArr, ok := receiver.(*runtime.ArrayValue)
			if !ok || loadedArr == nil {
				b.Fatalf("receiver validation failed")
			}
			benchmarkBytecodeArrayPushValueSink = loadedArr
			benchmarkBytecodeArrayPushIntSink = receiverIndex
		}
	})

	b.Run("array_slot_cache_hit_prefix_read", func(b *testing.B) {
		vm, program, instr, arr, indexVal, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = append(vm.stack[:0], arr, indexVal)
			kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
			if !ok {
				b.Fatalf("read kind selection failed")
			}
			if len(vm.stack) < instr.argCount+1 {
				b.Fatalf("benchmark stack underflow")
			}
			receiverIndex := len(vm.stack) - instr.argCount - 1
			receiver := vm.stack[receiverIndex]
			loadedArr, arrOK := receiver.(*runtime.ArrayValue)
			if !arrOK || loadedArr == nil {
				b.Fatalf("receiver validation failed")
			}
			argBase := receiverIndex + 1
			hit := vm.interp != nil &&
				!vm.hasRuntimeData() &&
				vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, 0, kind)
			if !hit {
				b.Fatalf("array slot cache prefix missed")
			}
			benchmarkBytecodeArrayPushIntSink = argBase
			benchmarkBytecodeArrayPushValueSink = loadedArr
		}
	})

	b.Run("array_slot_cache_hit_prefix_write", func(b *testing.B) {
		vm, program, instr, arr, indexVal, writeVal := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = append(vm.stack[:0], arr, indexVal, writeVal)
			kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
			if !ok {
				b.Fatalf("write kind selection failed")
			}
			if len(vm.stack) < instr.argCount+1 {
				b.Fatalf("benchmark stack underflow")
			}
			receiverIndex := len(vm.stack) - instr.argCount - 1
			receiver := vm.stack[receiverIndex]
			loadedArr, arrOK := receiver.(*runtime.ArrayValue)
			if !arrOK || loadedArr == nil {
				b.Fatalf("receiver validation failed")
			}
			argBase := receiverIndex + 1
			hit := vm.interp != nil &&
				!vm.hasRuntimeData() &&
				vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, 0, kind)
			if !hit {
				b.Fatalf("array slot cache prefix missed")
			}
			benchmarkBytecodeArrayPushIntSink = argBase
			benchmarkBytecodeArrayPushValueSink = loadedArr
		}
	})

	b.Run("array_slot_no_runtime_data_versions", func(b *testing.B) {
		vm, _, _, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			globalRevision, methodVersion, ok := vm.noRuntimeDataGlobalAndMethodVersions()
			if !ok {
				b.Fatalf("noRuntimeDataGlobalAndMethodVersions missed")
			}
			benchmarkBytecodeArrayPushIntSink = int(globalRevision + methodVersion)
		}
	})

	b.Run("array_slot_direct_cache_lookup_read_with_versions", func(b *testing.B) {
		vm, program, _, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)
		globalRevision, methodVersion, ok := vm.noRuntimeDataGlobalAndMethodVersions()
		if !ok {
			b.Fatalf("noRuntimeDataGlobalAndMethodVersions missed")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeArrayPushBoolSink = vm.lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions(
				program,
				0,
				bytecodeMemberMethodFastPathArrayReadSlot,
				globalRevision,
				methodVersion,
			)
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("read lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions returned false")
			}
		}
	})

	b.Run("array_slot_cache_hit_prefix_read_with_versions", func(b *testing.B) {
		vm, program, instr, arr, indexVal, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.stack = append(vm.stack[:0], arr, indexVal)
			kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
			if !ok {
				b.Fatalf("read kind selection failed")
			}
			if len(vm.stack) < instr.argCount+1 {
				b.Fatalf("benchmark stack underflow")
			}
			receiverIndex := len(vm.stack) - instr.argCount - 1
			receiver := vm.stack[receiverIndex]
			loadedArr, arrOK := receiver.(*runtime.ArrayValue)
			if !arrOK || loadedArr == nil {
				b.Fatalf("receiver validation failed")
			}
			argBase := receiverIndex + 1
			globalRevision, methodVersion, noRuntimeData := vm.noRuntimeDataGlobalAndMethodVersions()
			hit := noRuntimeData &&
				vm.lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions(program, 0, kind, globalRevision, methodVersion)
			if !hit {
				b.Fatalf("array slot cache prefix missed")
			}
			benchmarkBytecodeArrayPushIntSink = argBase
			benchmarkBytecodeArrayPushValueSink = loadedArr
		}
	})

	b.Run("array_slot_direct_cache_lookup_read", func(b *testing.B) {
		vm, program, _, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeArrayPushBoolSink = vm.lookupCachedCanonicalArraySlotCallForArrayValidated(
				program,
				0,
				bytecodeMemberMethodFastPathArrayReadSlot,
			)
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("read lookupCachedCanonicalArraySlotCallForArrayValidated returned false")
			}
		}
	})

	b.Run("array_slot_direct_cache_lookup_write", func(b *testing.B) {
		vm, program, _, _, _, _ := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeArrayPushBoolSink = vm.lookupCachedCanonicalArraySlotCallForArrayValidated(
				program,
				0,
				bytecodeMemberMethodFastPathArrayWriteSlot,
			)
			if !benchmarkBytecodeArrayPushBoolSink {
				b.Fatalf("write lookupCachedCanonicalArraySlotCallForArrayValidated returned false")
			}
		}
	})

	b.Run("exec_call_member_array_slot_read_cached", func(b *testing.B) {
		vm, program, instr, arr, indexVal, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal)
			newProg, err := vm.execCallMemberArraySlot(instr, program)
			if err != nil {
				b.Fatalf("execCallMemberArraySlot read failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("exec_call_member_array_slot_write_cached", func(b *testing.B) {
		vm, program, instr, arr, indexVal, writeVal := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, writeVal)
			newProg, err := vm.execCallMemberArraySlot(instr, program)
			if err != nil {
				b.Fatalf("execCallMemberArraySlot write failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("opcode_dispatch_array_slot_switch", func(b *testing.B) {
		_, _, instr, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			if instr == nil {
				b.Fatalf("missing benchmark instruction")
			}
			switch instr.op {
			case bytecodeOpCallMemberArraySlot:
				benchmarkBytecodeArrayPushBoolSink = true
			default:
				b.Fatalf("unexpected opcode %v", instr.op)
			}
		}
	})

	b.Run("run_loop_call_opcode_fetch_switch", func(b *testing.B) {
		_, program, _, _, _, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)
		instructions := program.instructions

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			ip := 0
			instr := &instructions[ip]
			switch instr.op {
			case bytecodeOpCall, bytecodeOpCallName, bytecodeOpCallMember, bytecodeOpCallMemberArrayGet, bytecodeOpCallMemberNext, bytecodeOpCallMemberArrayNew, bytecodeOpCallMemberArraySlot, bytecodeOpCallSelf, bytecodeOpCallSelfIntSubSlotConst:
				benchmarkBytecodeArrayPushBoolSink = instr.op == bytecodeOpCallMemberArraySlot
			default:
				b.Fatalf("unexpected opcode %v", instr.op)
			}
		}
	})

	b.Run("exec_call_opcode_array_slot_read_cached", func(b *testing.B) {
		vm, program, instr, arr, indexVal, _ := benchmarkArraySlotCallVM(b, "read_slot", 1, bytecodeMemberMethodFastPathArrayReadSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal)
			newProg, err := vm.execCallOpcode(instr, nil, program)
			if err != nil {
				b.Fatalf("execCallOpcode read failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})

	b.Run("exec_call_opcode_array_slot_write_cached", func(b *testing.B) {
		vm, program, instr, arr, indexVal, writeVal := benchmarkArraySlotCallVM(b, "write_slot", 2, bytecodeMemberMethodFastPathArrayWriteSlot)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = 0
			vm.stack = append(vm.stack[:0], arr, indexVal, writeVal)
			newProg, err := vm.execCallOpcode(instr, nil, program)
			if err != nil {
				b.Fatalf("execCallOpcode write failed: %v", err)
			}
			benchmarkBytecodeArrayPushProgramSink = newProg
			benchmarkBytecodeArrayPushValueSink = vm.stack[0]
		}
	})
}
