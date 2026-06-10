package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ExactGenericStructPatternMatchesOpenTypeArgument(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := bytecodeGenericStructPatternValue("Box", ast.Ty("i32"))

	got, matched, decided := vm.bytecodeMatchExactGenericStructPattern(ast.Gen(ast.Ty("Box"), ast.Ty("T")), value)
	if !decided {
		t.Fatalf("expected generic struct fast path to decide")
	}
	if !matched || got != value {
		t.Fatalf("expected open generic struct match, got matched=%v value=%#v", matched, got)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternUsesSimpleExactMatch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := runtime.NewSmallInt(7, runtime.IntegerI64)
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("i64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckI64,
		target:          9,
	}
	vm.stack = append(vm.stack, value)

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 || vm.stack[0] != value {
		t.Fatalf("stack = %#v, want exact matched value", vm.stack)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternExactIteratorEndDefMatchesRuntimeSentinel(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	iterEndDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("IteratorEnd", nil, ast.StructKindNamed, nil, nil, false),
	}
	instr := &bytecodeInstruction{
		op:       bytecodeOpJumpIfNotTypedPattern,
		typeExpr: ast.Ty("IteratorEnd"),
		value:    iterEndDef,
		target:   9,
	}
	vm.stack = append(vm.stack, runtime.IteratorEnd)

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	if _, ok := vm.stack[0].(runtime.IteratorEndValue); !ok {
		t.Fatalf("stack = %#v, want IteratorEnd sentinel", vm.stack)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternUsesSimpleIteratorEndCheck(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("IteratorEnd"),
		typeSimpleCheck: bytecodeSimpleTypeCheckIteratorEnd,
		target:          9,
	}
	vm.stack = append(vm.stack, runtime.IteratorEnd)

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 || len(vm.stack) != 1 {
		t.Fatalf("matched state ip=%d stack=%#v", vm.ip, vm.stack)
	}
	if _, ok := vm.stack[0].(runtime.IteratorEndValue); !ok {
		t.Fatalf("stack = %#v, want IteratorEnd sentinel", vm.stack)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternSimpleIteratorEndMissJumps(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("IteratorEnd"),
		typeSimpleCheck: bytecodeSimpleTypeCheckIteratorEnd,
		target:          9,
	}
	vm.stack = append(vm.stack, runtime.NewSmallInt(1, runtime.IntegerI64))

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 9 || len(vm.stack) != 0 {
		t.Fatalf("miss state ip=%d stack=%#v", vm.ip, vm.stack)
	}
}

func TestBytecodeMatchSimpleIteratorEndDefersInterfaceWrapper(t *testing.T) {
	wrapped := &runtime.InterfaceValue{Underlying: runtime.IteratorEnd}
	if _, _, decided := bytecodeMatchSimpleIteratorEndTypedPattern(wrapped); decided {
		t.Fatal("simple IteratorEnd check should defer interface wrappers")
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternSimpleMissJumps(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("i64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckI64,
		target:          9,
	}
	vm.stack = append(vm.stack, runtime.StringValue{Val: "no"})

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("vm ip = %d, want jump target 9", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack = %#v, want popped subject only", vm.stack)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternCoercesIntegerWithSimpleCheck(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("i64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckI64,
		target:          9,
	}
	vm.stack = append(vm.stack, runtime.NewSmallInt(7, runtime.IntegerI32))

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok || got.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("stack result = %#v, want coerced i64 integer", vm.stack[0])
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternKeepsRawIntegerCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	source := &bytecodeRawIntegerSlotCell{Raw: 7, TypeSuffix: runtime.IntegerU32}
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("u64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckU64,
		target:          9,
	}
	vm.stack = append(vm.stack, source)

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	source.Raw = 99
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU64 || raw != 7 {
		t.Fatalf("stack result = %#v, want raw u64 7 snapshot", vm.stack[0])
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("stack result = %#v, wanted raw integer carrier", vm.stack[0])
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternExactRawI64HotPathIsAllocationFree(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = make([]runtime.Value, 0, 1)
	source := &bytecodeRawI64SlotCell{Val: 7}
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("i64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckI64,
		target:          9,
	}
	runRawI64TypedPatternJump(t, vm, instr, source)

	allocs := testing.AllocsPerRun(1000, func() {
		runRawI64TypedPatternJump(t, vm, instr, source)
	})
	if allocs != 0 {
		t.Fatalf("expected exact raw i64 typed-pattern hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternGenericRawI64HotPathIsAllocationFree(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = make([]runtime.Value, 0, 1)
	source := &bytecodeRawI64SlotCell{Val: 7}
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("T"),
		typeSimpleCheck: bytecodeSimpleTypeCheckUnknown,
		target:          9,
	}
	runRawI64TypedPatternJump(t, vm, instr, source)

	allocs := testing.AllocsPerRun(1000, func() {
		runRawI64TypedPatternJump(t, vm, instr, source)
	})
	if allocs != 0 {
		t.Fatalf("expected generic raw i64 typed-pattern hot path allocations to be zero, got %.2f", allocs)
	}
}

func runRawI64TypedPatternJump(t *testing.T, vm *bytecodeVM, instr *bytecodeInstruction, source *bytecodeRawI64SlotCell) {
	t.Helper()
	vm.ip = 0
	source.Val = 7
	vm.stack = vm.stack[:0]
	vm.stack = append(vm.stack, source)
	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	got := vm.stack[0]
	if got == source {
		t.Fatalf("typed pattern reused mutable source cell")
	}
	source.Val = 99
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerI64 || raw != 7 {
		t.Fatalf("stack result = %#v, want raw i64 snapshot 7", got)
	}
	if _, boxed := got.(runtime.IntegerValue); boxed {
		t.Fatalf("stack result = %#v, wanted raw integer carrier", got)
	}
}

func TestBytecodeMatchExactStructDefinitionRawI64MissHotPathIsAllocationFree(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false),
	}
	source := &bytecodeRawI64SlotCell{Val: 7}
	if got, matched := bytecodeMatchExactStructDefinition(def, source); matched || got != nil {
		t.Fatalf("raw i64 should not match exact struct, got matched=%v value=%#v", matched, got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if got, matched := bytecodeMatchExactStructDefinition(def, source); matched || got != nil {
			t.Fatalf("raw i64 should not match exact struct, got matched=%v value=%#v", matched, got)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw i64 exact-struct miss allocations to be zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternCoercesFloatWithSimpleCheck(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("f64"),
		typeSimpleCheck: bytecodeSimpleTypeCheckF64,
		target:          9,
	}
	vm.stack = append(vm.stack, runtime.NewSmallInt(7, runtime.IntegerI32))

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	coerced, ok := vm.stack[0].(runtime.FloatValue)
	if !ok || coerced.TypeSuffix != runtime.FloatF64 || coerced.Val != 7 {
		t.Fatalf("stack result = %#v, want coerced f64 float", vm.stack[0])
	}
}

func TestBytecodeVM_JumpIfNotTypedPatternFallsBackForNominalString(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"String",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("u8"), "byte")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{"byte": runtime.NewSmallInt(65, runtime.IntegerU8)},
	}
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfNotTypedPattern,
		typeExpr:        ast.Ty("String"),
		typeSimpleCheck: bytecodeSimpleTypeCheckString,
		target:          9,
	}
	vm.stack = append(vm.stack, value)

	if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
		t.Fatalf("execJumpIfNotTypedPattern failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("vm ip = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 1 || vm.stack[0] != value {
		t.Fatalf("stack = %#v, want nominal String value via fallback", vm.stack)
	}
}

func TestBytecodeVM_ExactGenericStructPatternPlanMatchesOpenTypeArgument(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := bytecodeGenericStructPatternValue("Box", ast.Ty("i32"))
	plan := bytecodeGenericStructPatternPlanForTypeExpr(ast.Gen(ast.Ty("Box"), ast.Ty("T")))

	got, matched, decided := vm.bytecodeMatchExactGenericStructPatternPlan(plan, value)
	if !decided {
		t.Fatalf("expected planned generic struct fast path to decide")
	}
	if !matched || got != value {
		t.Fatalf("expected planned open generic struct match, got matched=%v value=%#v", matched, got)
	}
}

func TestBytecodeVM_ExactGenericStructPatternPlanPrecomputesOpenParam(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
	}
	plan := bytecodeGenericStructPatternPlanForTypeExprWithDefinition(ast.Gen(ast.Ty("Box"), ast.Ty("T")), def)
	if plan == nil || len(plan.args) != 1 {
		t.Fatalf("expected single-arg generic struct plan, got %#v", plan)
	}
	if !plan.args[0].simpleOpenParam {
		t.Fatalf("expected planned generic struct arg to precompute open parameter")
	}
}

func TestBytecodeVM_ExactGenericStructPatternMatchesKnownTypeArgument(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := bytecodeGenericStructPatternValue("Box", ast.Ty("i32"))

	got, matched, decided := vm.bytecodeMatchExactGenericStructPattern(ast.Gen(ast.Ty("Box"), ast.Ty("i32")), value)
	if !decided {
		t.Fatalf("expected generic struct fast path to decide")
	}
	if !matched || got != value {
		t.Fatalf("expected exact generic struct match, got matched=%v value=%#v", matched, got)
	}
}

func TestBytecodeVM_ExactGenericStructPatternRejectsKnownTypeArgumentMismatch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := bytecodeGenericStructPatternValue("Box", ast.Ty("i32"))

	_, matched, decided := vm.bytecodeMatchExactGenericStructPattern(ast.Gen(ast.Ty("Box"), ast.Ty("u32")), value)
	if !decided {
		t.Fatalf("expected generic struct fast path to decide")
	}
	if matched {
		t.Fatalf("expected exact generic struct mismatch")
	}
}

func bytecodeGenericStructPatternValue(name string, args ...ast.TypeExpression) *runtime.StructInstanceValue {
	return &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				name,
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
		},
		Fields:        map[string]runtime.Value{"value": runtime.NewSmallInt(7, runtime.IntegerI32)},
		TypeArguments: args,
	}
}
