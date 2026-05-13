package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsFloatAddMulSlotUpdate(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("s"), ast.Flt(0)),
			ast.Assign(ast.ID("a"), ast.Flt(2)),
			ast.Assign(ast.ID("b"), ast.Flt(3)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*", ast.ID("a"), ast.ID("b")))),
			ast.ID("s"),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddMul) {
		t.Fatalf("expected lowering to emit fused float add-mul slot update")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotFloatAddMul {
			if instr.target < 0 || instr.name != "s" {
				t.Fatalf("unexpected fused update instruction: %#v", instr)
			}
		}
	}
}

func TestBytecodeVM_LoweringEmitsFloatAddMulArrayGetSlotUpdate(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	def := ast.Fn(
		"step",
		[]*ast.FunctionParameter{
			ast.Param("ai", arrayF64),
			ast.Param("cj", arrayF64),
			ast.Param("k", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("s"), ast.Flt(1.5)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*",
				ast.Prop(ast.CallExpr(ast.Member(ast.ID("ai"), "get"), ast.ID("k"))),
				ast.Prop(ast.CallExpr(ast.Member(ast.ID("cj"), "get"), ast.ID("k"))),
			))),
			ast.ID("s"),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAddMulArrayGet) {
		t.Fatalf("expected lowering to emit fused float add-mul Array.get slot update")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCallMemberArrayGet) {
		t.Fatalf("fused Array.get update should avoid standalone Array.get member calls")
	}
}

func TestBytecodeVM_LoweringEmitsF64DotLoopPlan(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	def := ast.Fn(
		"dot",
		[]*ast.FunctionParameter{
			ast.Param("ai", arrayF64),
			ast.Param("cj", arrayF64),
			ast.Param("n", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("s"), ast.Flt(0)),
			ast.Assign(ast.ID("k"), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("k"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*",
					ast.Prop(ast.CallExpr(ast.Member(ast.ID("ai"), "get"), ast.ID("k"))),
					ast.Prop(ast.CallExpr(ast.Member(ast.ID("cj"), "get"), ast.ID("k"))),
				))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("k"), ast.Bin("+", ast.ID("k"), ast.Int(1))),
			),
			ast.ID("s"),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if len(program.f64DotLoops) != 1 {
		t.Fatalf("expected one f64 dot-loop plan, got %#v", program.f64DotLoops)
	}
	for ip, plan := range program.f64DotLoops {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpLoopEnter {
			t.Fatalf("f64 dot-loop plan attached to non-loop-enter ip %d", ip)
		}
		if plan.accumulatorSlot < 0 || plan.indexSlot < 0 || plan.boundSlot < 0 || plan.leftReceiverSlot < 0 || plan.rightReceiverSlot < 0 || plan.successTarget <= ip {
			t.Fatalf("unexpected f64 dot-loop plan: %#v", plan)
		}
	}
}

func TestBytecodeVM_LoweringF64DotLoopPlansResultAppend(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	def := ast.Fn(
		"dot_row",
		[]*ast.FunctionParameter{
			ast.Param("ai", arrayF64),
			ast.Param("cj", arrayF64),
			ast.Param("n", ast.Ty("i32")),
			ast.Param("di", arrayF64),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("s"), ast.Flt(0)),
			ast.Assign(ast.ID("k"), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("k"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*",
					ast.Prop(ast.CallExpr(ast.Member(ast.ID("ai"), "get"), ast.ID("k"))),
					ast.Prop(ast.CallExpr(ast.Member(ast.ID("cj"), "get"), ast.ID("k"))),
				))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("k"), ast.Bin("+", ast.ID("k"), ast.Int(1))),
			),
			ast.CallExpr(ast.Member(ast.ID("di"), "push"), ast.ID("s")),
			ast.ID("di"),
		},
		arrayF64,
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if len(program.f64DotLoops) != 1 {
		t.Fatalf("expected one f64 dot-loop plan, got %#v", program.f64DotLoops)
	}
	for ip, plan := range program.f64DotLoops {
		if !plan.resultAppend {
			t.Fatalf("expected f64 dot-loop result append plan: %#v", plan)
		}
		if plan.resultReceiverSlot < 0 || plan.resultPushIP <= ip || plan.resultTarget <= plan.resultPushIP {
			t.Fatalf("unexpected result append plan: ip=%d plan=%#v", ip, plan)
		}
		if got := program.instructions[plan.resultPushIP]; got.op != bytecodeOpCallMemberArraySlot || got.name != "push" || got.argCount != 1 {
			t.Fatalf("result append push ip points at %#v", got)
		}
	}
}

func TestBytecodeVM_FloatAddMulSlotUpdateParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("s"), ast.Flt(1.5)),
				ast.Assign(ast.ID("a"), ast.Flt(2)),
				ast.Assign(ast.ID("b"), ast.Flt(3)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*", ast.ID("a"), ast.ID("b")))),
				ast.ID("s"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float add-mul slot update mismatch: got=%#v want=%#v", got, want)
	}
	floatVal, ok := got.(runtime.FloatValue)
	if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7.5 {
		t.Fatalf("unexpected float result: %#v", got)
	}
}

func TestBytecodeVM_FloatAddMulArrayGetSlotUpdateFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
	}, 1)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
	}, 1)
	if _, err := interp.ensureArrayState(left, 0); err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	if _, err := interp.ensureArrayState(right, 0); err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	program, instr := floatAddMulArrayGetProgramForTest()
	vm.currentProgram = program
	vm.slots = []runtime.Value{runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}}
	vm.stack = []runtime.Value{
		left,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		right,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr)
	if err != nil {
		t.Fatalf("fused Array.get update failed: %v", err)
	}
	if handled {
		t.Fatalf("unexpected inline return from successful fused Array.get update")
	}
	if vm.ip != 1 {
		t.Fatalf("ip after fused Array.get update = %d, want 1", vm.ip)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 7.5)
	if len(vm.stack) != 1 {
		t.Fatalf("stack length after fused Array.get update = %d, want 1", len(vm.stack))
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF64, 7.5)
}

func TestBytecodeVM_F64DotLoopFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 4, TypeSuffix: runtime.FloatF64},
	}, 3)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 5, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 6, TypeSuffix: runtime.FloatF64},
		runtime.FloatValue{Val: 7, TypeSuffix: runtime.FloatF64},
	}, 3)
	leftState, err := interp.ensureArrayState(left, 0)
	if err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	rightState, err := interp.ensureArrayState(right, 0)
	if err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	leftState.ElementTypeToken = bytecodeIndexTypeF64
	leftState.ElementTypeTokenKnown = true
	rightState.ElementTypeToken = bytecodeIndexTypeF64
	rightState.ElementTypeTokenKnown = true

	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{{
			op:           bytecodeOpLoopEnter,
			loopBreak:    1,
			loopContinue: 0,
		}},
		f64DotLoops: map[int]bytecodeF64DotLoopPlan{0: {
			accumulatorSlot:   0,
			indexSlot:         1,
			boundSlot:         2,
			leftReceiverSlot:  3,
			rightReceiverSlot: 4,
			successTarget:     2,
		}},
	}
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(3), runtime.IntegerI32),
		left,
		right,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 dot-loop fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 dot-loop fast path to complete")
	}
	if vm.ip != 2 {
		t.Fatalf("ip after f64 dot-loop = %d, want 2", vm.ip)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 57.5)
	cell, ok := vm.slots[0].(*runtime.FloatValue)
	if !ok || cell == nil {
		t.Fatalf("f64 dot-loop accumulator = %#v, want owned pointer", vm.slots[0])
	}
	assertIntValue(t, vm.slots[1], runtime.IntegerI32, 3)
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("stack after f64 dot-loop = %#v, want nil loop result", vm.stack)
	}
	vm.stack = nil
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("load f64 dot-loop accumulator: %v", err)
	}
	if vm.stack[0] == cell {
		t.Fatalf("load should snapshot owned f64 accumulator, got cell pointer")
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF64, 57.5)
}

func TestBytecodeVM_F64DotLoopResultAppendFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := monoF64ArrayValueForTest(t, 2, 3, 4)
	right := monoF64ArrayValueForTest(t, 5, 6, 7)
	dest := interp.newArrayValue([]runtime.Value{}, 3)
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{
				op:           bytecodeOpLoopEnter,
				loopBreak:    1,
				loopContinue: 0,
			},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64DotLoops: map[int]bytecodeF64DotLoopPlan{0: {
			accumulatorSlot:    0,
			indexSlot:          1,
			boundSlot:          2,
			leftReceiverSlot:   3,
			rightReceiverSlot:  4,
			successTarget:      2,
			resultAppend:       true,
			resultReceiverSlot: 5,
			resultPushIP:       1,
			resultTarget:       4,
		}},
	}
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(3), runtime.IntegerI32),
		left,
		right,
		dest,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)
	vm.storeCachedCanonicalArraySlotCall(program, 1, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 result append dot-loop fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 result append dot-loop fast path to complete")
	}
	if vm.ip != 4 {
		t.Fatalf("ip after result append dot-loop = %d, want 4", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack after result append dot-loop = %#v, want empty", vm.stack)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 57.5)
	assertIntValue(t, vm.slots[1], runtime.IntegerI32, 3)
	values, mono, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 1 || values[0] != 57.5 {
		t.Fatalf("result row values=%#v mono=%v, want [57.5]", values, mono)
	}
}

func TestBytecodeVM_F64DotLoopReadsMonoF64Arrays(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := monoF64ArrayValueForTest(t, 2, 3, 4)
	right := monoF64ArrayValueForTest(t, 5, 6, 7)
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{{
			op:           bytecodeOpLoopEnter,
			loopBreak:    1,
			loopContinue: 0,
		}},
		f64DotLoops: map[int]bytecodeF64DotLoopPlan{0: {
			accumulatorSlot:   0,
			indexSlot:         1,
			boundSlot:         2,
			leftReceiverSlot:  3,
			rightReceiverSlot: 4,
			successTarget:     2,
		}},
	}
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(3), runtime.IntegerI32),
		left,
		right,
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	programPtr := program
	instructions := program.instructions
	handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
	if err != nil {
		t.Fatalf("f64 mono dot-loop fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected f64 mono dot-loop fast path to complete")
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 57.5)
	assertIntValue(t, vm.slots[1], runtime.IntegerI32, 3)
	if left.State != nil || right.State != nil || left.Elements != nil || right.Elements != nil {
		t.Fatalf("mono f64 dot-loop should not materialize boxed array state")
	}
	if len(vm.f64ArrayCache) != 0 {
		t.Fatalf("mono f64 dot-loop should not populate boxed row cache, got %d entries", len(vm.f64ArrayCache))
	}
}

func TestBytecodeVM_F64DotLoopRowCacheInvalidatesOnTrackedWrite(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
	}, 1)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 5, TypeSuffix: runtime.FloatF64},
	}, 1)
	leftState, err := interp.ensureArrayState(left, 0)
	if err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	rightState, err := interp.ensureArrayState(right, 0)
	if err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	leftState.ElementTypeToken = bytecodeIndexTypeF64
	leftState.ElementTypeTokenKnown = true
	rightState.ElementTypeToken = bytecodeIndexTypeF64
	rightState.ElementTypeTokenKnown = true
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{{
			op:           bytecodeOpLoopEnter,
			loopBreak:    1,
			loopContinue: 0,
		}},
		f64DotLoops: map[int]bytecodeF64DotLoopPlan{0: {
			accumulatorSlot:   0,
			indexSlot:         1,
			boundSlot:         2,
			leftReceiverSlot:  3,
			rightReceiverSlot: 4,
			successTarget:     2,
		}},
	}
	vm.currentProgram = program
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)
	run := func(want float64) {
		t.Helper()
		vm.ip = 0
		vm.stack = vm.stack[:0]
		vm.slots = []runtime.Value{
			runtime.FloatValue{Val: 0, TypeSuffix: runtime.FloatF64},
			runtime.NewBigIntValue(big.NewInt(0), runtime.IntegerI32),
			runtime.NewBigIntValue(big.NewInt(1), runtime.IntegerI32),
			left,
			right,
		}
		programPtr := program
		instructions := program.instructions
		handled, err := vm.execLoopEnterOpcode(&programPtr, &instructions, nil, nil, &program.instructions[0])
		if err != nil {
			t.Fatalf("f64 dot-loop fast path failed: %v", err)
		}
		if !handled {
			t.Fatalf("expected f64 dot-loop fast path to complete")
		}
		assertFloatValue(t, vm.slots[0], runtime.FloatF64, want)
	}
	run(10)
	if len(vm.f64ArrayCache) != 2 {
		t.Fatalf("expected f64 row cache entries for both rows, got %d", len(vm.f64ArrayCache))
	}
	replacement := runtime.FloatValue{Val: 10, TypeSuffix: runtime.FloatF64}
	leftState.Values[0] = replacement
	interp.syncTrackedArrayWrite(left, leftState, 0, replacement)
	run(50)
}

func monoF64ArrayValueForTest(t *testing.T, values ...float64) *runtime.ArrayValue {
	t.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityF64(len(values))
	for idx, value := range values {
		if err := runtime.ArrayStoreMonoWriteF64(handle, idx, value); err != nil {
			t.Fatalf("write mono f64 value %d: %v", idx, err)
		}
	}
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func TestBytecodeVM_FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
	}, 1)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
	}, 1)
	if _, err := interp.ensureArrayState(left, 0); err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	if _, err := interp.ensureArrayState(right, 0); err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	program, instr := floatAddMulArrayGetProgramForTest()
	instr.discardResult = true
	vm.currentProgram = program
	vm.slots = []runtime.Value{runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	vm.ip = 0
	vm.stack = floatAddMulArrayGetOperandsForTest(left, right)
	if handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr); err != nil || handled {
		t.Fatalf("first raw fused update handled=%v err=%v", handled, err)
	}
	cell, ok := vm.slots[0].(*runtime.FloatValue)
	if !ok || cell == nil {
		t.Fatalf("slot after raw fused update = %#v, want owned float cell", vm.slots[0])
	}
	if cell.Val != 7.5 || cell.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("raw fused update cell = %#v, want f64 7.5", cell)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded raw fused update stack = %#v, want empty", vm.stack)
	}

	vm.ip = 0
	vm.stack = floatAddMulArrayGetOperandsForTest(left, right)
	if handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr); err != nil || handled {
		t.Fatalf("second raw fused update handled=%v err=%v", handled, err)
	}
	if vm.slots[0] != cell {
		t.Fatalf("raw fused update should reuse owned cell")
	}
	if cell.Val != 13.5 {
		t.Fatalf("raw fused update reused cell value = %v, want 13.5", cell.Val)
	}

	vm.stack = nil
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("load raw accumulator slot: %v", err)
	}
	snapshot, ok := vm.stack[0].(runtime.FloatValue)
	if !ok || snapshot.Val != 13.5 || snapshot.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("loaded raw accumulator = %#v, want f64 value copy 13.5", vm.stack[0])
	}

	vm.ip = 0
	vm.stack = floatAddMulArrayGetOperandsForTest(left, right)
	if handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr); err != nil || handled {
		t.Fatalf("third raw fused update handled=%v err=%v", handled, err)
	}
	if cell.Val != 19.5 {
		t.Fatalf("third raw fused update cell value = %v, want 19.5", cell.Val)
	}
	if snapshot.Val != 13.5 {
		t.Fatalf("loaded accumulator snapshot changed to %v, want 13.5", snapshot.Val)
	}
}

func TestBytecodeVM_StoreSlotFloatReusesOwnedCellAcrossReinitialization(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NilValue{}}

	vm.stack = []runtime.Value{runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF64}}
	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: 0}); err != nil {
		t.Fatalf("initial f64 store slot: %v", err)
	}
	cell, ok := vm.slots[0].(*runtime.FloatValue)
	if !ok || cell == nil || cell.Val != 1.25 || cell.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("initial f64 store slot = %#v, want owned f64 cell 1.25", vm.slots[0])
	}
	if snapshot, ok := vm.stack[0].(runtime.FloatValue); !ok || snapshot.Val != 1.25 || snapshot.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("store expression result = %#v, want f64 value snapshot", vm.stack[0])
	}

	vm.stack = []runtime.Value{runtime.FloatValue{Val: 0, TypeSuffix: runtime.FloatF64}}
	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlotNew, target: 0}); err != nil {
		t.Fatalf("reinitialize f64 store slot: %v", err)
	}
	if vm.slots[0] != cell || cell.Val != 0 || cell.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("reinitialized f64 store slot = %#v cell=%#v, want same cell at 0", vm.slots[0], cell)
	}

	vm.slots[0] = runtime.NilValue{}
	vm.stack = []runtime.Value{runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64}}
	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlot, target: 0}); err != nil {
		t.Fatalf("reuse after slot clear: %v", err)
	}
	if vm.slots[0] != cell || cell.Val != 2.5 || cell.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("store after clear = %#v cell=%#v, want same cell at 2.5", vm.slots[0], cell)
	}

	vm.stack = nil
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("load owned f64 slot: %v", err)
	}
	if snapshot, ok := vm.stack[0].(runtime.FloatValue); !ok || snapshot.Val != 2.5 || snapshot.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("loaded owned f64 slot = %#v, want value snapshot 2.5", vm.stack[0])
	}
}

func TestBytecodeVM_FloatAddMulArrayGetSlotUpdatePropagatesNil(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue(nil, 0)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
	}, 1)
	leftState, err := interp.ensureArrayState(left, 0)
	if err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	leftState.ElementTypeToken = bytecodeIndexTypeF64
	leftState.ElementTypeTokenKnown = true
	if _, err := interp.ensureArrayState(right, 0); err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	program, instr := floatAddMulArrayGetProgramForTest()
	vm.currentProgram = program
	vm.slots = []runtime.Value{runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}}
	vm.stack = []runtime.Value{
		left,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		right,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr)
	if handled {
		t.Fatalf("top-level nil propagation should return a signal, not inline-handle")
	}
	ret, ok := err.(returnSignal)
	if !ok {
		t.Fatalf("nil Array.get propagation error = %T, want returnSignal", err)
	}
	if !isNilRuntimeValue(ret.value) {
		t.Fatalf("nil Array.get propagated value = %#v, want nil", ret.value)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack after nil propagation = %#v, want empty", vm.stack)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 1.5)
}

func TestBytecodeVM_FloatAddMulArrayGetSlotUpdateDoesNotSkipErrorResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.ErrorValue{Message: "bad"},
	}, 1)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
	}, 1)
	leftState, err := interp.ensureArrayState(left, 0)
	if err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	leftState.ElementTypeToken = bytecodeIndexTypeF64
	leftState.ElementTypeTokenKnown = true
	if _, err := interp.ensureArrayState(right, 0); err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	program, instr := floatAddMulArrayGetProgramForTest()
	vm.currentProgram = program
	vm.slots = []runtime.Value{runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}}
	vm.stack = []runtime.Value{
		left,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		right,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr)
	if handled {
		t.Fatalf("error propagation should not inline-handle")
	}
	raised, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("error Array.get propagation error = %T, want raiseSignal", err)
	}
	if got := raised.Error(); got != "bad" {
		t.Fatalf("raised error = %q, want bad", got)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack after error propagation = %#v, want empty", vm.stack)
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 1.5)
}

func TestBytecodeVM_FusedArrayGetFloatForTokenRejectsStaleToken(t *testing.T) {
	f64Value := runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64}
	got, ok := bytecodeFusedArrayGetFloatForToken(f64Value, bytecodeIndexTypeF64, true)
	if !ok || got.value != 2.5 || got.kind != runtime.FloatF64 {
		t.Fatalf("f64 token/value extraction = %#v, %v; want f64 2.5", got, ok)
	}

	f32Value := &runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF32}
	got, ok = bytecodeFusedArrayGetFloatForToken(f32Value, bytecodeIndexTypeF32, true)
	if !ok || got.value != 1.25 || got.kind != runtime.FloatF32 {
		t.Fatalf("f32 token/value extraction = %#v, %v; want f32 1.25", got, ok)
	}

	if got, ok := bytecodeFusedArrayGetFloatForToken(f32Value, bytecodeIndexTypeF64, true); ok {
		t.Fatalf("stale f64 token accepted f32 value: %#v", got)
	}
	if got, ok := bytecodeFusedArrayGetFloatForToken(runtime.ErrorValue{Message: "bad"}, bytecodeIndexTypeF64, true); ok {
		t.Fatalf("f64 token accepted error value: %#v", got)
	}
	if got, ok := bytecodeFusedArrayGetFloatForToken(f64Value, bytecodeIndexTypeF64, false); ok {
		t.Fatalf("unknown token accepted f64 value: %#v", got)
	}
}

func TestBytecodeVM_FloatAddMulArrayGetSlotUpdatePointerIndexFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	left := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
	}, 1)
	right := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64},
	}, 1)
	if _, err := interp.ensureArrayState(left, 0); err != nil {
		t.Fatalf("ensure left array state: %v", err)
	}
	if _, err := interp.ensureArrayState(right, 0); err != nil {
		t.Fatalf("ensure right array state: %v", err)
	}
	program, instr := floatAddMulArrayGetProgramForTest()
	instr.discardResult = true
	vm.currentProgram = program
	vm.slots = []runtime.Value{runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}}
	leftIndex := runtime.NewSmallInt(0, runtime.IntegerI32)
	rightIndex := runtime.NewSmallInt(0, runtime.IntegerI32)
	vm.stack = []runtime.Value{left, &leftIndex, right, &rightIndex}
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, left)

	handled, err := execFloatAddMulArrayGetUpdateForTest(t, vm, program, instr)
	if err != nil {
		t.Fatalf("fused Array.get update failed: %v", err)
	}
	if handled {
		t.Fatalf("unexpected inline return from successful fused Array.get update")
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 7.5)
	if len(vm.stack) != 0 {
		t.Fatalf("discarded pointer-index fused update stack = %#v, want empty", vm.stack)
	}
}

func TestBytecodeVM_FloatAddMulSlotUpdateFallbackParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("s"), ast.Int(1)),
				ast.Assign(ast.ID("a"), ast.Int(2)),
				ast.Assign(ast.ID("b"), ast.Int(3)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*", ast.ID("a"), ast.ID("b")))),
				ast.ID("s"),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode add-mul fallback mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 7)
}

func TestBytecodeVM_FloatAddMulSlotUpdatePreservesRHSOrder(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("s"), ast.Flt(1)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Bin("+", ast.ID("s"), ast.Bin("*",
					ast.AssignOp(ast.AssignmentAssign, ast.ID("s"), ast.Flt(10)),
					ast.Flt(2),
				))),
				ast.ID("s"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode add-mul evaluation order mismatch: got=%#v want=%#v", got, want)
	}
	floatVal, ok := got.(runtime.FloatValue)
	if !ok || floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 21 {
		t.Fatalf("unexpected order-sensitive result: %#v", got)
	}
}

func floatAddMulArrayGetProgramForTest() (*bytecodeProgram, *bytecodeInstruction) {
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{
		op:       bytecodeOpStoreSlotFloatAddMulArrayGet,
		target:   0,
		name:     "s",
		operator: "+",
	}}}
	return program, &program.instructions[0]
}

func floatAddMulArrayGetOperandsForTest(left *runtime.ArrayValue, right *runtime.ArrayValue) []runtime.Value {
	return []runtime.Value{
		left,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		right,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
}

func execFloatAddMulArrayGetUpdateForTest(t *testing.T, vm *bytecodeVM, program *bytecodeProgram, instr *bytecodeInstruction) (bool, error) {
	t.Helper()
	programPtr := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	return vm.execStoreSlotFloatAddMulArrayGet(&programPtr, &instructions, &validatedIntConsts, &slotConstIntImmTable, instr)
}
