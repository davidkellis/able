package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeProgramReachAttributesTopLevelExecution(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	fn := ast.Fn("work", nil, nil, ast.Ty("i32"), nil, nil, false, false)
	ast.SetSpan(fn, ast.Span{Start: ast.Position{Line: 7, Column: 3}})
	interp.nodeOrigins = make(map[ast.Node]string)
	interp.nodeOrigins[fn] = "/tmp/work.able"
	program := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: runtime.NewSmallInt(4, runtime.IntegerI32)},
		{op: bytecodeOpReturn},
	}})
	program = interp.annotateBytecodeProgramReach(program, "function", "work", fn)

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	if _, err := vm.run(program); err != nil {
		t.Fatalf("run: %v", err)
	}
	stats := interp.BytecodeStats()
	if len(stats.ProgramReach) != 1 {
		t.Fatalf("program rows = %d, want 1", len(stats.ProgramReach))
	}
	row := stats.ProgramReach[0]
	if row.Name != "work" || row.Origin != "/tmp/work.able" || row.Line != 7 || row.Column != 3 {
		t.Fatalf("unexpected identity: %+v", row)
	}
	if row.Entries != 1 || row.DynamicInstructions != 2 || row.DynamicPrimitiveEligible != 1 || row.DynamicEffectBoundaries != 1 {
		t.Fatalf("unexpected dynamic reach: %+v", row)
	}
}

func TestBytecodeProgramReachAnnotationIsOptIn(t *testing.T) {
	interp := NewBytecode()
	program := interp.annotateBytecodeProgramReach(&bytecodeProgram{}, "function", "work", nil)
	if program.reach != nil {
		t.Fatalf("reach metadata should be absent when bytecode stats are disabled")
	}
}

func TestBytecodeProgramReachCountsInlineProgramEntryOnce(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	caller := interp.annotateBytecodeProgramReach(&bytecodeProgram{}, "function", "caller", nil)
	callee := interp.annotateBytecodeProgramReach(&bytecodeProgram{}, "function", "callee", nil)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = caller
	vm.pushCallFrame(1, caller, nil, interp.GlobalEnvironment(), nil, 0, 0, false, false)
	vm.ip = 0
	program := caller
	instructions := caller.instructions
	var validated []bool
	var immediates *bytecodeSlotConstIntImmediateTable
	vm.switchRunProgram(&program, &instructions, &validated, &immediates, callee)
	vm.switchRunProgram(&program, &instructions, &validated, &immediates, callee)

	stats := interp.BytecodeStats()
	if len(stats.ProgramReach) != 1 || stats.ProgramReach[0].Name != "callee" || stats.ProgramReach[0].Entries != 1 {
		t.Fatalf("unexpected inline entries: %+v", stats.ProgramReach)
	}
}

func TestBytecodeProgramReachClassifiesPrimitiveSpanAndBackedge(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	program := interp.annotateBytecodeProgramReach(&bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpConst, value: runtime.IntegerValue{TypeSuffix: runtime.IntegerI32}},
		{op: bytecodeOpStoreSlotI32},
		{op: bytecodeOpJump, target: 0},
		{op: bytecodeOpCallName},
	}}, "function", "loop", nil)
	interp.recordBytecodeProgramEntry(program)
	for ip := range program.instructions {
		interp.recordBytecodeProgramInstruction(program, ip, &program.instructions[ip])
	}

	row := interp.BytecodeStats().ProgramReach[0]
	if row.StaticPrimitiveEligible != 3 || row.StaticEffectBoundaries != 1 || row.StaticBackedges != 1 {
		t.Fatalf("unexpected static classification: %+v", row)
	}
	if row.MaxStaticPrimitiveSpan != 3 || row.DynamicBackedges != 1 {
		t.Fatalf("unexpected primitive span/backedge reach: %+v", row)
	}
}

func TestBytecodeProgramReachResetAllowsReregistration(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	program := interp.annotateBytecodeProgramReach(&bytecodeProgram{}, "function", "main", nil)
	interp.recordBytecodeProgramEntry(program)
	interp.ResetBytecodeStats()
	if rows := interp.BytecodeStats().ProgramReach; len(rows) != 0 {
		t.Fatalf("rows after reset: %+v", rows)
	}
	interp.recordBytecodeProgramEntry(program)
	rows := interp.BytecodeStats().ProgramReach
	if len(rows) != 1 || rows[0].Entries != 1 {
		t.Fatalf("rows after reregistration: %+v", rows)
	}
}

func TestBytecodeProgramReachAttributesScalarProofGaps(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true
	definition := &runtime.StructDefinitionValue{Node: ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{ast.NewStructFieldDefinition(ast.Ty("i64"), ast.ID("x"))},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)}
	cast := ast.NewTypeCastExpression(ast.Int(1), ast.Ty("f64"))
	compare := ast.Bin("<", ast.Int(1), ast.Int(2))
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpLoadSlot, target: 0},
			{op: bytecodeOpLoadSlotStructField, target: 1},
			{op: bytecodeOpCast, node: cast},
			{op: bytecodeOpBinary, operator: "<", node: compare},
			{op: bytecodeOpLoadSlot, target: 1},
		},
		frameLayout: &bytecodeFrameLayout{slotKinds: []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindValue}},
		namedStructMembers: map[int]bytecodeNamedStructMemberPlan{
			1: {definition: definition, fieldIndex: 0},
		},
	}
	checks := []bytecodeSimpleTypeCheck{
		bytecodeSimpleTypeCheckF64,
		bytecodeSimpleTypeCheckUnknown,
		bytecodeSimpleTypeCheckUnknown,
		bytecodeSimpleTypeCheckUnknown,
		bytecodeSimpleTypeCheckUnknown,
	}
	program = interp.annotateBytecodeProgramReachWithScalarChecks(program, "function", "proofs", nil, checks)
	interp.recordBytecodeProgramEntry(program)
	for ip := range program.instructions {
		interp.recordBytecodeProgramInstruction(program, ip, &program.instructions[ip])
	}

	row := interp.BytecodeStats().ProgramReach[0]
	if row.StaticPrimitiveEligible != 4 || row.DynamicPrimitiveEligible != 4 {
		t.Fatalf("primitive reach = static %d dynamic %d, want 4/4", row.StaticPrimitiveEligible, row.DynamicPrimitiveEligible)
	}
	got := make(map[string]BytecodeScalarProofSnapshot, len(row.ScalarProofs))
	for _, proof := range row.ScalarProofs {
		got[proof.Proof] = proof
	}
	for _, name := range []string{
		"primitive-slot-float",
		"primitive-field-integer",
		"primitive-numeric-cast",
		"primitive-integer-compare",
		"unproven-or-boxed",
	} {
		proof, ok := got[name]
		if !ok || proof.StaticInstructions != 1 || proof.DynamicInstructions != 1 {
			t.Fatalf("proof %q = %+v, present=%v; all=%+v", name, proof, ok, row.ScalarProofs)
		}
	}
}
