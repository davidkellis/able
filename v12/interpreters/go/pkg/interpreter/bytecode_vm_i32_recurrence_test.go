package interpreter

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMDetectsI32RecurrenceKernel(t *testing.T) {
	program := lowerI32RecurrenceTestProgram(t)
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected i32 recurrence kernel")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntSlotBase(t *testing.T) {
	program := lowerI32RecurrenceIntSlotBaseProgram(t)
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected i32 recurrence kernel for Int slot base")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64Source(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i64-literal source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64ConstRangeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i64 const-range source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64OutOfOrderMixedSeedSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i64 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i64 out-of-order mixed-seed source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI64Source(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i64 source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI64UntypedBaseReturnSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i64 source with untyped base return")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactIsizeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: isize) -> isize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact isize source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactU64Source(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: u64) -> u64 {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact u64 source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactUsizeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: usize) -> usize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact usize source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32LessThanConstSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n < 3 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 less-than const source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32LessThanSlotBaseSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n < 2 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 less-than slot-base source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI64LessThanConstSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n < 3_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i64 less-than const source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32GreaterThanConstLeftSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if 3 > n { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 const-left greater-than source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32GreaterThanSlotBaseConstLeftSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if 2 > n { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 const-left greater-than slot-base source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32EqualityBasePrefixSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n == 1 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 equality-base prefix source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32EqualityBasePrefixCurrentReturnSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return n }
  if n == 1 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 equality-base prefix current-return source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32EqualityPrefixPlusRangeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 equality-prefix plus range source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32EqualityPrefixPlusCurrentRangeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 equality-prefix plus current-range source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForExactI32OutOfOrderEqualityPlusCurrentRangeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 2 { return 1 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for exact i32 out-of-order equality plus current-range source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForAdditionalExactIntegerWidths(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		callArg  int64
	}{
		{name: "i8", typeName: "i8", callArg: 10},
		{name: "i16", typeName: "i16", callArg: 20},
		{name: "u8", typeName: "u8", callArg: 10},
		{name: "u16", typeName: "u16", callArg: 20},
		{name: "u32", typeName: "u32", callArg: 30},
		{name: "i128", typeName: "i128", callArg: 30},
		{name: "u128", typeName: "u128", callArg: 30},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			module := parseRecurrenceSourceModule(t, exactIntegerRecurrenceSource(tc.typeName, tc.callArg))
			def := firstFunctionDefinition(t, module)
			program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
			if err != nil {
				t.Fatalf("bytecode lowering failed: %v", err)
			}
			if program.i32RecurrenceKernel == nil {
				t.Fatalf("expected recurrence kernel for exact %s source shape", tc.typeName)
			}
		})
	}
}

func TestBytecodeVMI32RecurrenceKernelParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		i32RecurrenceTestFunction("fib"),
		ast.Call("fib", ast.Int(10)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntSlotBaseParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		intRecurrenceSlotBaseTestFunction("fib"),
		ast.Call("fib", ast.Int(10)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode Int recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64Parity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64ConstRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 const-range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64ConstRangeDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i64 const-range source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI64)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int i64 const-range fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int i64 const-range to use direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64OutOfOrderMixedSeedParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i64 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 out-of-order mixed-seed recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64OutOfOrderMixedSeedNegativeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i64 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(-1_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 out-of-order mixed-seed negative recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, -1)
}

func TestBytecodeVMI32RecurrenceKernelExactI64Parity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i64 recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI64UntypedBaseReturnParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i64 untyped-base recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactIsizeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: isize) -> isize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact isize recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerIsize, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactU64Parity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: u64) -> u64 {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact u64 recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerU64, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactUsizeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: usize) -> usize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact usize recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerUsize, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32LessThanConstParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n < 3 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 less-than const recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32LessThanSlotBaseParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n < 2 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 less-than slot-base recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI64LessThanConstParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i64) -> i64 {
  if n < 3_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i64 less-than const recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32GreaterThanConstLeftParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if 3 > n { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 const-left greater-than recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32GreaterThanSlotBaseConstLeftParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if 2 > n { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 const-left greater-than slot-base recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32EqualityBasePrefixParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n == 1 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 equality-base prefix recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32EqualityBasePrefixCurrentReturnParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return n }
  if n == 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 equality-base prefix current-return recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32EqualityPrefixPlusRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 equality-prefix plus range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32EqualityPrefixPlusCurrentRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 equality-prefix plus current-range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32OutOfOrderEqualityPlusCurrentRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 2 { return 1 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 out-of-order equality plus current-range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelExactI32EqualityPrefixPlusRangeNegativeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(-1)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 equality-prefix plus range negative recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 1)
}

func TestBytecodeVMI32RecurrenceKernelExactI32OutOfOrderEqualityPlusCurrentRangeNegativeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: i32) -> i32 {
  if n == 2 { return 1 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(-1)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode exact i32 out-of-order equality plus current-range negative recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, -1)
}

func TestBytecodeVMI32RecurrenceKernelAdditionalExactIntegerWidthParity(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		callArg  int64
		wantType runtime.IntegerType
		wantRaw  int64
	}{
		{name: "i8", typeName: "i8", callArg: 10, wantType: runtime.IntegerI8, wantRaw: 55},
		{name: "i16", typeName: "i16", callArg: 20, wantType: runtime.IntegerI16, wantRaw: 6765},
		{name: "u8", typeName: "u8", callArg: 10, wantType: runtime.IntegerU8, wantRaw: 55},
		{name: "u16", typeName: "u16", callArg: 20, wantType: runtime.IntegerU16, wantRaw: 6765},
		{name: "u32", typeName: "u32", callArg: 30, wantType: runtime.IntegerU32, wantRaw: 832040},
		{name: "i128", typeName: "i128", callArg: 30, wantType: runtime.IntegerI128, wantRaw: 832040},
		{name: "u128", typeName: "u128", callArg: 30, wantType: runtime.IntegerU128, wantRaw: 832040},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			module := parseRecurrenceSourceModule(t, exactIntegerRecurrenceSource(tc.typeName, tc.callArg))
			want := mustEvalModule(t, New(), module)
			got := runBytecodeModule(t, module)
			if !valuesEqual(got, want) {
				t.Fatalf("bytecode exact %s recurrence mismatch: got=%#v want=%#v", tc.typeName, got, want)
			}
			assertIntValue(t, got, tc.wantType, tc.wantRaw)
		})
	}
}

func TestBytecodeVMI32RecurrenceKernelOverflowParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"boom",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{
				ast.IfExpr(
					ast.Bin("<=", ast.ID("n"), ast.Int(0)),
					ast.Block(ast.Ret(ast.Int(math.MaxInt32))),
				),
				ast.Bin(
					"+",
					ast.Call("boom", ast.Bin("-", ast.ID("n"), ast.Int(1))),
					ast.Call("boom", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("boom", ast.Int(1)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "integer overflow") {
		t.Fatalf("expected tree integer overflow, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "integer overflow") {
		t.Fatalf("expected bytecode integer overflow, got: %v", byteErr)
	}
}

func TestBytecodeVMI32RecurrenceKernelNegativeTermParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"weird",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{
				ast.IfExpr(
					ast.Bin("<=", ast.ID("n"), ast.Int(0)),
					ast.Block(ast.Ret(ast.Int(1))),
				),
				ast.Bin(
					"+",
					ast.Call("weird", ast.Bin("-", ast.ID("n"), ast.Int(2))),
					ast.Call("weird", ast.Bin("-", ast.ID("n"), ast.Int(3))),
				),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("weird", ast.Int(5)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode weird recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 5)
}

func lowerI32RecurrenceTestProgram(t *testing.T) *bytecodeProgram {
	t.Helper()
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(i32RecurrenceTestFunction("fib"))
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	return program
}

func lowerI32RecurrenceIntSlotBaseProgram(t *testing.T) *bytecodeProgram {
	t.Helper()
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(intRecurrenceSlotBaseTestFunction("fib"))
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	return program
}

func i32RecurrenceTestFunction(name string) *ast.FunctionDefinition {
	return ast.Fn(
		name,
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
}

func intRecurrenceSlotBaseTestFunction(name string) *ast.FunctionDefinition {
	return ast.Fn(
		name,
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("Int"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(1)),
				ast.Block(ast.Ret(ast.ID("n"))),
			),
			ast.Bin(
				"+",
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("Int"),
		nil,
		nil,
		false,
		false,
	)
}

func parseRecurrenceSourceModule(t *testing.T, src string) *ast.Module {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "recurrence-*.able")
	if err != nil {
		t.Fatalf("create temp source: %v", err)
	}
	path := tmpFile.Name()
	defer os.Remove(path)
	if _, err := tmpFile.WriteString(src); err != nil {
		tmpFile.Close()
		t.Fatalf("write temp source: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp source: %v", err)
	}
	module, err := parseSourceModule(path)
	if err != nil {
		t.Fatalf("parse temp source: %v", err)
	}
	return module
}

func exactIntegerRecurrenceSource(typeName string, callArg int64) string {
	return fmt.Sprintf(`fn fib(n: %s) -> %s {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
fib(%d)
`, typeName, typeName, callArg)
}

func firstFunctionDefinition(t *testing.T, module *ast.Module) *ast.FunctionDefinition {
	t.Helper()
	if module == nil {
		t.Fatalf("nil module")
	}
	for _, stmt := range module.Body {
		if def, ok := stmt.(*ast.FunctionDefinition); ok && def != nil {
			return def
		}
	}
	t.Fatalf("module missing function definition")
	return nil
}
