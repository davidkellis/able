package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsTryArrayPushF64AffineProduct(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	expr := ast.Bin("*",
		ast.ID("t"),
		ast.Bin("*",
			ast.NewTypeCastExpression(ast.Bin("-", ast.ID("i"), ast.ID("j")), ast.Ty("f64")),
			ast.NewTypeCastExpression(ast.Bin("+", ast.ID("i"), ast.ID("j")), ast.Ty("f64")),
		),
	)
	def := ast.Fn(
		"fill",
		[]*ast.FunctionParameter{
			ast.Param("row", arrayF64),
			ast.Param("t", ast.Ty("f64")),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("j", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("row"), "push"), expr),
			ast.ID("row"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpTryArrayPushF64AffineProduct) {
		t.Fatalf("expected lowering to emit f64 affine Array.push try opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallMemberArraySlot) {
		t.Fatalf("expected lowering to retain Array.push fallback call")
	}
	if len(program.f64AffinePushes) != 1 {
		t.Fatalf("expected one f64 affine push plan, got %#v", program.f64AffinePushes)
	}
	for ip, plan := range program.f64AffinePushes {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpTryArrayPushF64AffineProduct {
			t.Fatalf("f64 affine push plan attached to wrong ip %d", ip)
		}
		if !plan.validForSlots(4) || program.instructions[ip].target <= ip {
			t.Fatalf("unexpected f64 affine push plan or target: plan=%#v instr=%#v", plan, program.instructions[ip])
		}
	}
}

func TestBytecodeVM_LoweringEmitsTryArrayPushF64NestedGet(t *testing.T) {
	arrayF64 := ast.Gen(ast.Ty("Array"), ast.Ty("f64"))
	arrayArrayF64 := ast.Gen(ast.Ty("Array"), arrayF64)
	arg := ast.Prop(ast.CallExpr(
		ast.Member(
			ast.Prop(ast.CallExpr(ast.Member(ast.ID("b"), "get"), ast.ID("j"))),
			"get",
		),
		ast.ID("i"),
	))
	def := ast.Fn(
		"transpose_cell",
		[]*ast.FunctionParameter{
			ast.Param("row", arrayF64),
			ast.Param("b", arrayArrayF64),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("j", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("row"), "push"), arg),
			ast.ID("row"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpTryArrayPushF64NestedGet) {
		t.Fatalf("expected lowering to emit f64 nested Array.get push try opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallMemberArraySlot) {
		t.Fatalf("expected lowering to retain Array.push fallback call")
	}
	if len(program.f64NestedGetPushes) != 1 {
		t.Fatalf("expected one f64 nested get push plan, got %#v", program.f64NestedGetPushes)
	}
	for ip, plan := range program.f64NestedGetPushes {
		if ip < 0 || ip >= len(program.instructions) || program.instructions[ip].op != bytecodeOpTryArrayPushF64NestedGet {
			t.Fatalf("f64 nested get push plan attached to wrong ip %d", ip)
		}
		if !plan.validForSlots(4) || plan.receiverSlot != 0 || plan.outerSlot != 1 || plan.rowIndexSlot != 3 || plan.colIndexSlot != 2 || program.instructions[ip].target <= ip {
			t.Fatalf("unexpected f64 nested get push plan or target: plan=%#v instr=%#v", plan, program.instructions[ip])
		}
	}
}

func TestBytecodeVM_TryArrayPushF64AffineProductAppendsDirectFloat(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	arr := interp.newArrayValue([]runtime.Value{}, 1)
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpTryArrayPushF64AffineProduct, name: "push", argCount: 1, target: 2},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64AffinePushes: map[int]bytecodeF64AffineProductPushPlan{0: {
			receiverSlot: 0,
			scaleSlot:    1,
			leftSlot:     2,
			rightSlot:    3,
		}},
	}
	vm.storeCachedCanonicalArraySlotCall(program, 0, program.instructions[1], arr, bytecodeMemberMethodFastPathArrayPush)
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		arr,
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(3), runtime.IntegerI32),
	}

	if err := vm.execTryArrayPushF64AffineProduct(program, &program.instructions[0]); err != nil {
		t.Fatalf("try f64 affine push failed: %v", err)
	}
	if vm.ip != 2 {
		t.Fatalf("ip after fast push = %d, want 2", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length after fast push = %d, want 1", len(vm.stack))
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("fast push should leave void result, got %#v", vm.stack[0])
	}
	values, monoF64, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(arr.Handle)
	if err != nil {
		t.Fatalf("read mono f64 values: %v", err)
	}
	if !monoF64 || len(values) != 1 || values[0] != 32 {
		t.Fatalf("expected one mono f64 value 32, mono=%v values=%#v arr=%#v", monoF64, values, arr)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono f64 fast push should detach boxed tracked state, arr=%#v", arr)
	}
	boxed, err := runtime.ArrayStoreRead(arr.Handle, 0)
	if err != nil {
		t.Fatalf("generic mono f64 read: %v", err)
	}
	got, ok := boxed.(runtime.FloatValue)
	if !ok || got.TypeSuffix != runtime.FloatF64 || got.Val != 32 {
		t.Fatalf("generic mono f64 read = %#v, want f64 32", boxed)
	}
}

func TestBytecodeVM_TryArrayPushF64NestedGetAppendsDirectFloat(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	dest := interp.newArrayValue([]runtime.Value{}, 1)
	row := monoF64ArrayValueForTest(t, 3.25, 4.5)
	outer := interp.newArrayValue([]runtime.Value{row}, 1)
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpTryArrayPushF64NestedGet, name: "push", argCount: 1, target: 2},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64NestedGetPushes: map[int]bytecodeF64NestedArrayGetPushPlan{0: {
			receiverSlot: 0,
			outerSlot:    1,
			rowIndexSlot: 2,
			colIndexSlot: 3,
		}},
	}
	vm.storeCachedCanonicalArraySlotCall(program, 0, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		dest,
		outer,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	if err := vm.execTryArrayPushF64NestedGet(program, &program.instructions[0]); err != nil {
		t.Fatalf("try f64 nested get push failed: %v", err)
	}
	if vm.ip != 2 {
		t.Fatalf("ip after fast push = %d, want 2", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length after fast push = %d, want 1", len(vm.stack))
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("fast push should leave void result, got %#v", vm.stack[0])
	}
	values, monoF64, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(dest.Handle)
	if err != nil {
		t.Fatalf("read mono f64 values: %v", err)
	}
	if !monoF64 || len(values) != 1 || values[0] != 4.5 {
		t.Fatalf("expected one mono f64 value 4.5, mono=%v values=%#v dest=%#v", monoF64, values, dest)
	}
}

func TestBytecodeVM_TryArrayPushF64NestedGetFallsThroughOnGuardMiss(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	dest := interp.newArrayValue([]runtime.Value{}, 1)
	outer := interp.newArrayValue([]runtime.Value{}, 1)
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpTryArrayPushF64NestedGet, name: "push", argCount: 1, target: 2},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64NestedGetPushes: map[int]bytecodeF64NestedArrayGetPushPlan{0: {
			receiverSlot: 0,
			outerSlot:    1,
			rowIndexSlot: 2,
			colIndexSlot: 3,
		}},
	}
	vm.storeCachedCanonicalArraySlotCall(program, 0, program.instructions[1], dest, bytecodeMemberMethodFastPathArrayPush)
	vm.storeCachedCanonicalArrayGetCall(program, 0, bytecodeInstruction{name: "get", argCount: 1}, outer)
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		dest,
		outer,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	if err := vm.execTryArrayPushF64NestedGet(program, &program.instructions[0]); err != nil {
		t.Fatalf("guard miss should fall through without error: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after guard miss = %d, want fallback ip 1", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("guard miss should not push a result, stack=%#v", vm.stack)
	}
	if size, err := runtime.ArrayStoreSize(dest.Handle); err != nil || size != 0 {
		t.Fatalf("guard miss should not append, size=%d err=%v", size, err)
	}
}

func TestBytecodeVM_TryArrayPushF64AffineProductFallsThroughOnGuardMiss(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpTryArrayPushF64AffineProduct, name: "push", argCount: 1, target: 2},
			{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1},
		},
		f64AffinePushes: map[int]bytecodeF64AffineProductPushPlan{0: {
			receiverSlot: 0,
			scaleSlot:    1,
			leftSlot:     2,
			rightSlot:    3,
		}},
	}
	vm.currentProgram = program
	vm.ip = 0
	vm.slots = []runtime.Value{
		runtime.NilValue{},
		runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		runtime.NewBigIntValue(big.NewInt(5), runtime.IntegerI32),
		runtime.NewBigIntValue(big.NewInt(3), runtime.IntegerI32),
	}

	if err := vm.execTryArrayPushF64AffineProduct(program, &program.instructions[0]); err != nil {
		t.Fatalf("guard miss should fall through without error: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after guard miss = %d, want fallback ip 1", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("guard miss should not push a result, stack=%#v", vm.stack)
	}
}

func TestBytecodeVM_F64AffineProductPushMatcherRejectsMismatchedTerms(t *testing.T) {
	ctx := &bytecodeLoweringContext{
		frameLayout: &bytecodeFrameLayout{},
		slotScopes: []map[string]int{{
			"row": 0,
			"t":   1,
			"i":   2,
			"j":   3,
			"k":   4,
		}},
	}
	expr := ast.Bin("*",
		ast.ID("t"),
		ast.Bin("*",
			ast.NewTypeCastExpression(ast.Bin("+", ast.ID("i"), ast.ID("j")), ast.Ty("f64")),
			ast.NewTypeCastExpression(ast.Bin("-", ast.ID("i"), ast.ID("k")), ast.Ty("f64")),
		),
	)
	if _, ok := bytecodeF64AffineProductPushPlanForCall(ctx, ast.Member(ast.ID("row"), "push"), expr); ok {
		t.Fatalf("mismatched affine product terms should not match")
	}
}
