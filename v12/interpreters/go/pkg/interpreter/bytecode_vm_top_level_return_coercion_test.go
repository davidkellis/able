package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMRunTopLevelReturnCoercesFrameLayoutReturnType(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("u64"),
			returnSimpleType:  "u64",
			returnSimpleCheck: bytecodeSimpleTypeCheckU64,
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(7, runtime.IntegerI32)},
			{op: bytecodeOpReturn},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	assertIntValue(t, got, runtime.IntegerU64, 7)
}

func TestBytecodeVMRunTopLevelReturnSignalCoercesFrameLayoutReturnType(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	returnNode := ast.Ret(ast.Int(7))
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("u64"),
			returnSimpleType:  "u64",
			returnSimpleCheck: bytecodeSimpleTypeCheckU64,
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(7, runtime.IntegerI32)},
			{op: bytecodeOpReturn, node: returnNode},
		},
	}

	_, err := vm.run(program)
	ret, ok := err.(returnSignal)
	if !ok {
		t.Fatalf("vm.run() error = %T, want returnSignal", err)
	}
	assertIntValue(t, ret.value, runtime.IntegerU64, 7)
}

func TestBytecodeVMRunTopLevelReturnUsesProgramGenericNames(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType: ast.Ty("T"),
		},
		returnGenericNames:       map[string]struct{}{"T": {}},
		returnGenericNamesCached: true,
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(7, runtime.IntegerI32)},
			{op: bytecodeOpReturn},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	assertIntValue(t, got, runtime.IntegerI32, 7)
}

func TestBytecodeVMRunTopLevelReturnBinaryIntAddCoercesFrameLayoutReturnType(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("u64"),
			returnSimpleType:  "u64",
			returnSimpleCheck: bytecodeSimpleTypeCheckU64,
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(19, runtime.IntegerI32)},
			{op: bytecodeOpConst, value: runtime.NewSmallInt(23, runtime.IntegerI32)},
			{op: bytecodeOpReturnBinaryIntAddI32, operator: "+"},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	assertIntValue(t, got, runtime.IntegerU64, 42)
}

func TestBytecodeVMRunTopLevelReturnCoercesNullableSimpleRawInteger(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:           ast.Nullable(ast.Ty("u8")),
			returnNullableSimple: "u8",
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: bytecodeRawU8ResultValue(65)},
			{op: bytecodeOpReturn},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	assertIntValue(t, got, runtime.IntegerU8, 65)
}

func TestBytecodeVMRunTopLevelReturnCoercesNullableSimpleNil(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:           ast.Nullable(ast.Ty("u8")),
			returnNullableSimple: "u8",
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NilValue{}},
			{op: bytecodeOpReturn},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	if !isNilRuntimeValue(got) {
		t.Fatalf("nullable simple nil return = %#v, want nil", got)
	}
}

func TestBytecodeVMRunTopLevelReturnNullableSimpleMismatchUsesGenericError(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:           ast.Nullable(ast.Ty("u8")),
			returnNullableSimple: "u8",
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.StringValue{Val: "no"}},
			{op: bytecodeOpReturn},
		},
	}

	_, err := vm.run(program)
	if err == nil {
		t.Fatalf("vm.run() error = nil, want nullable return mismatch")
	}
}

func TestBytecodeVMRunTopLevelReturnDirectVoidUsesFrameLayoutFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:       ast.Ty("void"),
			returnSimpleType: "void",
		},
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(7, runtime.IntegerI32)},
			{op: bytecodeOpReturn},
		},
	}

	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("vm.run() error = %v", err)
	}
	if _, ok := got.(runtime.VoidValue); !ok {
		t.Fatalf("expected direct void return coercion to produce void, got %#v", got)
	}
}
