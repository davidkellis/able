package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CanonicalStdlibStructFastPathDetection(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	randomStruct := ast.StructDef("Random", nil, ast.StructKindNamed, nil, nil, false)
	int128Struct := ast.StructDef("Int128", nil, ast.StructKindNamed, nil, nil, false)
	uint128Struct := ast.StructDef("UInt128", nil, ast.StructKindNamed, nil, nil, false)

	randomNextI64 := ast.Fn("next_i64", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
	}, []ast.Statement{ast.Int(0)}, ast.Ty("i64"), nil, nil, false, false)
	int128Add := ast.Fn("add", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("other", ast.Ty("Int128")),
	}, []ast.Statement{ast.Int(0)}, ast.Ty("Int128"), nil, nil, false, false)
	uint128Mul := ast.Fn("mul", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("other", ast.Ty("UInt128")),
	}, []ast.Statement{ast.Int(0)}, ast.Ty("UInt128"), nil, nil, false, false)
	customNextI64 := ast.Fn("next_i64", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
	}, []ast.Statement{ast.Int(0)}, ast.Ty("i64"), nil, nil, false, false)

	interp.SetNodeOrigins(map[ast.Node]string{
		randomStruct:  "/tmp/able-stdlib/src/random.able",
		int128Struct:  "/tmp/able-stdlib/src/numbers/int128.able",
		uint128Struct: "/tmp/able-stdlib/src/numbers/uint128.able",
		randomNextI64: "/tmp/able-stdlib/src/random.able",
		int128Add:     "/tmp/able-stdlib/src/numbers/int128.able",
		uint128Mul:    "/tmp/able-stdlib/src/numbers/uint128.able",
		customNextI64: "/tmp/project/random.able",
	})

	randomKey := bytecodeMemberMethodCacheKey{
		member:        "next_i64",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverStruct,
		structDef:     &runtime.StructDefinitionValue{Node: randomStruct},
	}
	if got := vm.memberMethodFastPathFor(randomKey, &runtime.FunctionValue{Declaration: randomNextI64}); got != bytecodeMemberMethodFastPathRandomNextI64 {
		t.Fatalf("Random.next_i64 fast path = %d, want RandomNextI64", got)
	}
	if got := vm.memberMethodFastPathFor(randomKey, &runtime.FunctionValue{Declaration: customNextI64}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom Random.next_i64 fast path = %d, want none", got)
	}

	int128Key := bytecodeMemberMethodCacheKey{
		member:        "add",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverStruct,
		structDef:     &runtime.StructDefinitionValue{Node: int128Struct},
	}
	if got := vm.memberMethodFastPathFor(int128Key, &runtime.FunctionValue{Declaration: int128Add}); got != bytecodeMemberMethodFastPathInt128Add {
		t.Fatalf("Int128.add fast path = %d, want Int128Add", got)
	}

	uint128Key := bytecodeMemberMethodCacheKey{
		member:        "mul",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverStruct,
		structDef:     &runtime.StructDefinitionValue{Node: uint128Struct},
	}
	if got := vm.memberMethodFastPathFor(uint128Key, &runtime.FunctionValue{Declaration: uint128Mul}); got != bytecodeMemberMethodFastPathUInt128Mul {
		t.Fatalf("UInt128.mul fast path = %d, want UInt128Mul", got)
	}
}

func TestBytecodeVM_CanonicalStdlibNumericStructMethodsUseFastPathsFromSource(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")

	source := `
import able.numbers.int128.{Int128}
import able.numbers.uint128.{UInt128}

fn main() -> i64 {
  total := 0_i64
  lhs := Int128.zero()
  lhs = lhs.add(Int128.from_i64(7_i64))
  lhs = lhs.mul(Int128.from_i64(5_i64))
  lhs = lhs.rem(Int128.from_i64(11_i64))
  total = total + lhs.to_i64()!

  rhs := UInt128.zero()
  rhs = rhs.add(UInt128.from_u64(7_u64))
  rhs = rhs.mul(UInt128.from_u64(5_u64))
  rhs = rhs.rem(UInt128.from_u64(11_u64))
  total = total + (rhs.to_u64()! as i64)

  total
}

main()
`

	program := mustLoadAbleProgramFromSource(t, source)
	want, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("tree evaluation failed: %v", err)
	}
	interp := NewBytecode()
	got, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("canonical stdlib struct trace fixture mismatch: got=%#v want=%#v", got, want)
	}

	snapshot := interp.BytecodeTrace(0)
	// The canonical constructors are ordinary Able source methods: Int128's
	// from_i64 delegates to from_i128, and UInt128's from_u64 builds its named
	// fields directly. Keep this assertion on the shared arithmetic and checked
	// conversion operations that the VM actually recognizes as fast paths.
	wantDispatches := map[string]bool{
		"int128_add_fast":     false,
		"int128_mul_fast":     false,
		"int128_rem_fast":     false,
		"int128_to_i64_fast":  false,
		"uint128_add_fast":    false,
		"uint128_mul_fast":    false,
		"uint128_rem_fast":    false,
		"uint128_to_u64_fast": false,
	}
	for _, entry := range snapshot.Entries {
		if _, ok := wantDispatches[entry.Dispatch]; ok {
			wantDispatches[entry.Dispatch] = true
		}
	}
	for dispatch, seen := range wantDispatches {
		if !seen {
			t.Fatalf("missing trace dispatch %q in %#v", dispatch, snapshot.Entries)
		}
	}
}

func TestBytecodeVM_CanonicalRandomFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	randomStruct := ast.StructDef("Random", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i64"), "state"),
	}, ast.StructKindNamed, nil, nil, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		randomStruct: "/tmp/able-stdlib/src/random.able",
	})
	randomDef := &runtime.StructDefinitionValue{Node: randomStruct}

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{randomDef, runtime.NewSmallInt(123_456_789, runtime.IntegerI64)}
	_, handled, err := vm.execStaticCanonicalStructMemberFast(
		bytecodeInstruction{name: "seeded", argCount: 1},
		vm.stack[0],
		0,
		1,
		nil,
	)
	if err != nil || !handled {
		t.Fatalf("Random.seeded fast path failed: handled=%v err=%v", handled, err)
	}
	rng, ok := vm.stack[0].(*runtime.StructInstanceValue)
	if !ok || rng == nil {
		t.Fatalf("Random.seeded result = %#v, want struct instance", vm.stack[0])
	}
	state, ok := bytecodeStructI64Field(rng, "state")
	if !ok || state != 123_456_789 {
		t.Fatalf("Random.seeded state = %d ok=%v, want 123456789", state, ok)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{rng}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathRandomNextI64,
		bytecodeInstruction{name: "next_i64", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil || !handled {
		t.Fatalf("Random.next_i64 fast path failed: handled=%v err=%v", handled, err)
	}
	next1 := (123_456_789 * 48271) % 2147483647
	if got, ok := vm.stack[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI64 || got.Int64Fast() != int64(next1) {
		t.Fatalf("Random.next_i64 result = %#v, want i64 %d", vm.stack[0], next1)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{rng}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathRandomNextI32,
		bytecodeInstruction{name: "next_i32", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil || !handled {
		t.Fatalf("Random.next_i32 fast path failed: handled=%v err=%v", handled, err)
	}
	next2 := (next1 * 48271) % 2147483647
	if got, ok := vm.stack[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 || got.Int64Fast() != int64(next2) {
		t.Fatalf("Random.next_i32 result = %#v, want i32 %d", vm.stack[0], next2)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{rng}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathRandomNextF64,
		bytecodeInstruction{name: "next_f64", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil || !handled {
		t.Fatalf("Random.next_f64 fast path failed: handled=%v err=%v", handled, err)
	}
	next3 := (next2 * 48271) % 2147483647
	result, ok := vm.stack[0].(runtime.FloatValue)
	if !ok || result.TypeSuffix != runtime.FloatF64 || result.Val != float64(next3)/2147483647.0 {
		t.Fatalf("Random.next_f64 result = %#v, want f64 %v", vm.stack[0], float64(next3)/2147483647.0)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{randomDef}
	_, handled, err = vm.execStaticCanonicalStructMemberFast(
		bytecodeInstruction{name: "default", argCount: 0},
		vm.stack[0],
		0,
		1,
		nil,
	)
	if err != nil || !handled {
		t.Fatalf("Random.default fast path failed: handled=%v err=%v", handled, err)
	}
	defaultRng, ok := vm.stack[0].(*runtime.StructInstanceValue)
	if !ok || defaultRng == nil {
		t.Fatalf("Random.default result = %#v, want struct instance", vm.stack[0])
	}
	if got, ok := bytecodeStructI64Field(defaultRng, "state"); !ok || got != 1 {
		t.Fatalf("Random.default state = %d ok=%v, want 1", got, ok)
	}
}

func TestBytecodeVM_CanonicalUInt128FromI64FastPathFallsBackForNegative(t *testing.T) {
	interp := NewBytecode()
	randomStruct := ast.StructDef("UInt128", nil, ast.StructKindNamed, nil, nil, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		randomStruct: "/tmp/able-stdlib/src/numbers/uint128.able",
	})

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{
		&runtime.StructDefinitionValue{Node: randomStruct},
		runtime.NewSmallInt(-1, runtime.IntegerI64),
	}
	_, handled, err := vm.execStaticCanonicalStructMemberFast(
		bytecodeInstruction{name: "from_i64", argCount: 1},
		vm.stack[0],
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("UInt128.from_i64 negative fast path returned error instead of falling back: %v", err)
	}
	if handled {
		t.Fatalf("UInt128.from_i64 negative fast path should fall back")
	}
}

func TestBytecodeVM_CanonicalInt128ToI64FastPathFallsBackOutOfRange(t *testing.T) {
	interp := NewBytecode()
	int128Struct := ast.StructDef("Int128", nil, ast.StructKindNamed, nil, nil, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		int128Struct: "/tmp/able-stdlib/src/numbers/int128.able",
	})

	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: int128Struct},
		Fields: map[string]runtime.Value{
			"high": runtime.NewBigIntValue(newBigUint64(1), runtime.IntegerU64),
			"low":  runtime.NewSmallInt(0, runtime.IntegerU64),
		},
	}

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = []runtime.Value{inst}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathInt128ToI64,
		bytecodeInstruction{name: "to_i64", argCount: 0},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("Int128.to_i64 out-of-range fast path returned error instead of falling back: %v", err)
	}
	if handled {
		t.Fatalf("Int128.to_i64 out-of-range fast path should fall back")
	}
}

func newBigUint64(value uint64) *big.Int {
	return new(big.Int).SetUint64(value)
}
