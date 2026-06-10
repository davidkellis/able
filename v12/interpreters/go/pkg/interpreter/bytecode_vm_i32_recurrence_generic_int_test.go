package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64UntypedConstRangeSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
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
		t.Fatalf("expected recurrence kernel for generic Int i64 untyped const-range source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64UntypedOutOfOrderMixedSeedSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1 }
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
		t.Fatalf("expected recurrence kernel for generic Int i64 untyped mixed-seed source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI64UntypedLessThanConstSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n < 3_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i64 untyped less-than const source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI32MixedRecursiveKindsSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1_i64) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 mixed-recursive-kinds source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI32ThreeKindRecursiveSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 0 }
  fib(n - 1_i64) + fib(n - 2_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 three-kind recursive source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI32SparseThreeKindRecursiveSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 2000_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 sparse three-kind recursive source shape")
	}
}

func TestBytecodeVMDetectsI32RecurrenceKernelForIntI32SpillHeavyThreeKindRecursiveSource(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1001_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 spill-heavy three-kind recursive source shape")
	}
}

func TestBytecodeVMI32RecurrenceKernelIntI32SpillHeavyThreeKindRecursivePrefersPagedDPHeuristic(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1001_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 spill-heavy three-kind recursive source shape")
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(runtime.IntegerI32, program.i32RecurrenceKernel.firstSubKind, program.i32RecurrenceKernel.secondSubKind)
	if !ok {
		t.Fatalf("expected kind set for generic Int i32 spill-heavy three-kind recursive source shape")
	}
	if !program.i32RecurrenceKernel.genericPagedDPPreferred(bytecodeI32RecurrenceDPMaxInput+424, len(kindSet.kinds)) {
		t.Fatalf("expected spill-heavy close-step source to prefer paged DP above the dense-DP budget")
	}
}

func TestBytecodeVMI32RecurrenceKernelIntI32SparseThreeKindRecursiveKeepsMemoHeuristic(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 2000_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 sparse three-kind recursive source shape")
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(runtime.IntegerI32, program.i32RecurrenceKernel.firstSubKind, program.i32RecurrenceKernel.secondSubKind)
	if !ok {
		t.Fatalf("expected kind set for generic Int i32 sparse three-kind recursive source shape")
	}
	if program.i32RecurrenceKernel.genericPagedDPPreferred(bytecodeI32RecurrenceDPMaxInput+424, len(kindSet.kinds)) {
		t.Fatalf("expected large-step sparse source to stay off the paged DP heuristic")
	}
}

func TestBytecodeVMI32RecurrenceKernelIntI32BoundaryThreeKindRecursivePrefersPagedDPHeuristic(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1049_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 boundary three-kind recursive source shape")
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(runtime.IntegerI32, program.i32RecurrenceKernel.firstSubKind, program.i32RecurrenceKernel.secondSubKind)
	if !ok {
		t.Fatalf("expected kind set for generic Int i32 boundary three-kind recursive source shape")
	}
	if !program.i32RecurrenceKernel.genericPagedDPPreferred(bytecodeI32RecurrenceDPMaxInput+424, len(kindSet.kinds)) {
		t.Fatalf("expected 1000/1049 source to stay on the paged DP side of the density cutoff")
	}
}

func TestBytecodeVMI32RecurrenceKernelIntI32BoundaryThreeKindRecursiveKeepsMemoHeuristic(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1051_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 boundary three-kind recursive source shape")
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(runtime.IntegerI32, program.i32RecurrenceKernel.firstSubKind, program.i32RecurrenceKernel.secondSubKind)
	if !ok {
		t.Fatalf("expected kind set for generic Int i32 boundary three-kind recursive source shape")
	}
	if program.i32RecurrenceKernel.genericPagedDPPreferred(bytecodeI32RecurrenceDPMaxInput+424, len(kindSet.kinds)) {
		t.Fatalf("expected 1000/1051 source to stay on the memo side of the density cutoff")
	}
}

func TestBytecodeVMI32RecurrenceKernelIntI64UntypedConstRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 untyped const-range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 55)
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64UntypedConstRangeDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1 }
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
		t.Fatalf("expected recurrence kernel for generic Int i64 untyped const-range source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI64)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int i64 untyped const-range fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int i64 untyped const-range to use direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64UntypedOutOfOrderMixedSeedParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 untyped mixed-seed recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI64, 55)
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI64UntypedOutOfOrderMixedSeedNegativeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(-1_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 untyped mixed-seed negative recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI64, -1)
	assertIntValue(t, got, runtime.IntegerI64, -1)
}

func TestBytecodeVMI32RecurrenceKernelIntI64UntypedLessThanConstParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n < 3_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i64 untyped less-than const recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 55)
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32ConstRangeParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i32 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int explicit i32 const-range recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 55)
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32ConstRangeBaseCaseParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i32 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(1_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int explicit i32 const-range base-case mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 1)
	assertIntValue(t, got, runtime.IntegerI32, 1)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32ConstRangeUsesDirectFastPathForI64Arg(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i32 }
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
		t.Fatalf("expected recurrence kernel for generic Int explicit i32 const-range source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI64)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected explicit i32 const-range fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected explicit i32 const-range with i64 arg to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32MixedSeedParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i32 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(10_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int explicit i32 mixed-seed recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI64, 55)
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32MixedSeedSecondBaseCaseParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i32 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
fib(2_i64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int explicit i32 mixed-seed second-base mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 1)
	assertIntValue(t, got, runtime.IntegerI32, 1)
}

func TestBytecodeVMI32RecurrenceKernelIntExplicitI32MixedSeedUsesDirectFastPathForI64Arg(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i32 }
  if n <= 1_i64 { return n }
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
		t.Fatalf("expected recurrence kernel for generic Int explicit i32 mixed-seed source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI64)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected explicit i32 mixed-seed fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected explicit i32 mixed-seed with i64 arg to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI32MixedRecursiveKindsParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1_i64) + fib(n - 2)
}
fib(10_i32)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i32 mixed-recursive-kinds recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI64, 55)
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI32MixedRecursiveKindsBaseCaseParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1_i64) + fib(n - 2)
}
fib(1_i32)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i32 mixed-recursive-kinds base-case mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 1)
	assertIntValue(t, got, runtime.IntegerI32, 1)
}

func TestBytecodeVMI32RecurrenceKernelIntI32MixedRecursiveKindsUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1_i64) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int i32 mixed-recursive-kinds source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI32)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int i32 mixed-recursive-kinds fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int i32 mixed-recursive-kinds source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI64, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntI32MixedRecursiveKindsOversizeUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 0 }
  fib(n - 1_i64) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int oversize mixed-recursive-kinds source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(bytecodeI32RecurrenceDPMaxInput+123, runtime.IntegerI32)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int oversize mixed-recursive-kinds fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int oversize mixed-recursive-kinds source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32ThreeKindRecursiveParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 0 }
  fib(n - 1_i64) + fib(n - 2_i128)
}
fib(10_i32)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i32 three-kind recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 0)
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32ThreeKindRecursiveOversizeUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 0 }
  fib(n - 1_i64) + fib(n - 2_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int oversize three-kind recursive source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(bytecodeI32RecurrenceDPMaxInput+123, runtime.IntegerI32)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int oversize three-kind recursive fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int oversize three-kind recursive source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32SparseThreeKindRecursiveParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 2000_i128)
}
fib(5000_i32)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i32 sparse three-kind recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 0)
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32SparseThreeKindRecursiveOversizeUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 2000_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int oversize sparse three-kind recursive source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(bytecodeI32RecurrenceDPMaxInput+424, runtime.IntegerI32)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int oversize sparse three-kind recursive fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int oversize sparse three-kind recursive source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32SpillHeavyThreeKindRecursiveParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1001_i128)
}
fib(5000_i32)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int i32 spill-heavy three-kind recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI32, 0)
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntI32SpillHeavyThreeKindRecursiveOversizeUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 2000_i64 { return 0 }
  fib(n - 1000_i64) + fib(n - 1001_i128)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int oversize spill-heavy three-kind recursive source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(bytecodeI32RecurrenceDPMaxInput+424, runtime.IntegerI32)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int oversize spill-heavy three-kind recursive fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int oversize spill-heavy three-kind recursive source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 0)
}

func TestBytecodeVMI32RecurrenceKernelIntU64UntypedCurrentReturnUsesDirectFastPath(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
`)
	def := firstFunctionDefinition(t, module)
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected recurrence kernel for generic Int u64 untyped current-return source shape")
	}

	activeProgram := program
	instructions := program.instructions
	var validatedIntConsts []bool
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerU64)}

	fastReturned, got, err := vm.tryExecI32RecurrenceProgram(&activeProgram, &instructions, &validatedIntConsts, &slotConstIntImmTable, false)
	if err != nil {
		t.Fatalf("unexpected generic Int u64 untyped current-return fast-path error: %v", err)
	}
	if !fastReturned {
		t.Fatalf("expected generic Int u64 untyped current-return source to use the direct recurrence fast path")
	}
	assertIntValue(t, got, runtime.IntegerI128, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntU64UntypedCurrentReturnParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(10_u64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int u64 untyped current-return recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerI128, 55)
	assertIntValue(t, got, runtime.IntegerI128, 55)
}

func TestBytecodeVMI32RecurrenceKernelIntU64UntypedCurrentReturnBaseCaseParity(t *testing.T) {
	module := parseRecurrenceSourceModule(t, `fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
fib(1_u64)
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic Int u64 untyped current-return base-case mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, want, runtime.IntegerU64, 1)
	assertIntValue(t, got, runtime.IntegerU64, 1)
}
