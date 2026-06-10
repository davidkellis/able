package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMStackDiagnosticsTracksDepthsAndCapacityGrowth(t *testing.T) {
	interp := &Interpreter{bytecodeStatsEnabled: true}
	vm := &bytecodeVM{interp: interp}
	vm.stack = make([]runtime.Value, 3, 4)
	vm.callFrameKinds = make([]bytecodeCallFrameKind, 2)
	vm.selfFastMinimalSuffix = 1
	vm.recordBytecodeStackDiagnostics()

	vm.stack = append(vm.stack, runtime.NilValue{}, runtime.NilValue{})
	vm.callFrameKinds = append(vm.callFrameKinds, bytecodeCallFrameKindFull)
	vm.recordBytecodeStackDiagnostics()

	stats := interp.BytecodeStats()
	if stats.ValueStackMaxDepth != 5 {
		t.Fatalf("value stack max depth = %d, want 5", stats.ValueStackMaxDepth)
	}
	if stats.ValueStackMaxCapacity < 5 {
		t.Fatalf("value stack max capacity = %d, want at least 5", stats.ValueStackMaxCapacity)
	}
	if stats.ValueStackCapacityGrowths != 1 {
		t.Fatalf("value stack capacity growths = %d, want 1", stats.ValueStackCapacityGrowths)
	}
	if stats.CallFrameMaxDepth != 4 {
		t.Fatalf("call frame max depth = %d, want 4", stats.CallFrameMaxDepth)
	}

	vm = &bytecodeVM{interp: interp}
	constInstr := &bytecodeInstruction{op: bytecodeOpConst}
	popInstr := &bytecodeInstruction{op: bytecodeOpPop}
	vm.beginBytecodeInstructionDiagnostics(bytecodeOpConst, 3, constInstr)
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.beginBytecodeInstructionDiagnostics(bytecodeOpPop, 4, popInstr)
	vm.stack = vm.stack[:0]
	vm.finishBytecodeInstructionDiagnostics()
	stats = interp.BytecodeStats()
	if got := stats.ValueStackDeltasByName[bytecodeOpName(bytecodeOpConst)]; got != 1 {
		t.Fatalf("const value-stack delta = %d, want 1", got)
	}
	if got := stats.ValueStackDeltasByName[bytecodeOpName(bytecodeOpPop)]; got != -1 {
		t.Fatalf("pop value-stack delta = %d, want -1", got)
	}
	if len(stats.TopValueStackPeakSites) != 1 || stats.TopValueStackPeakSites[0].Op != int(bytecodeOpConst) || stats.TopValueStackPeakSites[0].IP != 3 || stats.TopValueStackPeakSites[0].Growth != 1 {
		t.Fatalf("peak sites = %+v, want const at ip 3 with growth 1", stats.TopValueStackPeakSites)
	}
	if len(stats.TopValueStackDeltaSites) != 1 || stats.TopValueStackDeltaSites[0].Op != int(bytecodeOpConst) || stats.TopValueStackDeltaSites[0].IP != 3 || stats.TopValueStackDeltaSites[0].Delta != 1 {
		t.Fatalf("positive delta sites = %+v, want const at ip 3 with delta 1", stats.TopValueStackDeltaSites)
	}

	callNode := ast.ID("call")
	interp.SetNodeOrigins(map[ast.Node]string{callNode: "call-balance.able"})
	callVM := &bytecodeVM{interp: interp, stack: []runtime.Value{runtime.NilValue{}, runtime.NilValue{}, runtime.NilValue{}}}
	callInstr := &bytecodeInstruction{op: bytecodeOpCallMember, argCount: 1, name: "write", node: callNode}
	region := callVM.beginBytecodeCallOperandRegion(callInstr)
	if !region.valid || region.base != 1 || region.operandValues != 2 || region.expectedResults != 1 {
		t.Fatalf("call region = %+v, want member receiver/argument base", region)
	}
	callVM.stack = append(callVM.stack[:1], runtime.VoidValue{})
	callVM.completeBytecodeCallOperandRegion(region, len(callVM.stack))
	if got := interp.BytecodeStats().TopCallOperandBalances; len(got) != 0 {
		t.Fatalf("balanced call recorded imbalance: %+v", got)
	}

	callVM.stack = []runtime.Value{runtime.NilValue{}, runtime.NilValue{}, runtime.NilValue{}}
	region = callVM.beginBytecodeCallOperandRegion(callInstr)
	callVM.stack = append(callVM.stack[:1], runtime.VoidValue{}, runtime.NilValue{})
	callVM.completeBytecodeCallOperandRegion(region, len(callVM.stack))
	stats = interp.BytecodeStats()
	if len(stats.TopCallOperandBalances) != 1 {
		t.Fatalf("call operand balances = %+v, want one entry", stats.TopCallOperandBalances)
	}
	callBalance := stats.TopCallOperandBalances[0]
	if callBalance.Origin != "call-balance.able" || callBalance.Violations != 1 || callBalance.OperandValues != 2 || callBalance.ExpectedResults != 1 || callBalance.ActualResults != 2 || callBalance.Excess != 1 || callBalance.MaxExcess != 1 {
		t.Fatalf("call operand balance = %+v, want one excess result at call site", callBalance)
	}

	jumpNode := ast.ID("loop")
	interp.SetNodeOrigins(map[ast.Node]string{jumpNode: "loop-balance.able"})
	jumpVM := &bytecodeVM{interp: interp, currentProgram: &bytecodeProgram{}, stack: []runtime.Value{runtime.NilValue{}, runtime.NilValue{}}}
	jumpInstr := &bytecodeInstruction{op: bytecodeOpJump, target: 4, node: jumpNode}
	jumpVM.ip = 10
	jumpVM.beginBytecodeInstructionDiagnostics(bytecodeOpJump, 10, jumpInstr)
	jumpVM.ip = 4
	jumpVM.beginBytecodeInstructionDiagnostics(bytecodeOpConst, 4, &bytecodeInstruction{op: bytecodeOpConst})
	jumpVM.stack = append(jumpVM.stack, runtime.NilValue{}, runtime.NilValue{}, runtime.NilValue{})
	jumpVM.ip = 10
	jumpVM.beginBytecodeInstructionDiagnostics(bytecodeOpJump, 10, jumpInstr)
	jumpVM.ip = 4
	jumpVM.beginBytecodeInstructionDiagnostics(bytecodeOpConst, 4, &bytecodeInstruction{op: bytecodeOpConst})
	stats = interp.BytecodeStats()
	if len(stats.TopLoopBackedgeBalances) != 1 {
		t.Fatalf("loop backedge balances = %+v, want one entry", stats.TopLoopBackedgeBalances)
	}
	loopBalance := stats.TopLoopBackedgeBalances[0]
	if loopBalance.Origin != "loop-balance.able" || loopBalance.Backedges != 2 || loopBalance.Baseline != 2 || loopBalance.Excess != 3 || loopBalance.MaxExcess != 3 {
		t.Fatalf("loop backedge balance = %+v, want baseline and excess at jump site", loopBalance)
	}

	pooledVM := &bytecodeVM{bytecodeStatsInlineCallOperands: []bytecodeCallOperandRegion{{valid: true}}}
	pooledVM.resetForRun(interp, nil)
	if len(pooledVM.bytecodeStatsInlineCallOperands) != 0 {
		t.Fatalf("pooled VM retained deferred call operand diagnostics")
	}

	frameVM := &bytecodeVM{interp: interp, stack: make([]runtime.Value, 2)}
	frameVM.pushCallFrame(1, nil, nil, nil, nil, 0, 0, false, false)
	if got := frameVM.topInlineFrameStackBase(); got != 2 {
		t.Fatalf("inline frame stack base = %d, want 2", got)
	}
	returnNode := ast.ID("result")
	interp.SetNodeOrigins(map[ast.Node]string{returnNode: "stack-balance.able"})
	interp.recordBytecodeInlineFrameBalance(&bytecodeInstruction{node: returnNode}, 2, 5)
	stats = interp.BytecodeStats()
	if len(stats.TopInlineFrameBalances) != 1 {
		t.Fatalf("inline frame balances = %+v, want one entry", stats.TopInlineFrameBalances)
	}
	balance := stats.TopInlineFrameBalances[0]
	if balance.Origin != "stack-balance.able" || balance.Excess != 3 || balance.Max != 3 || balance.Returns != 1 {
		t.Fatalf("inline frame balance = %+v, want origin/excess/max/returns", balance)
	}

	interp.ResetBytecodeStats()
	stats = interp.BytecodeStats()
	if stats.ValueStackMaxDepth != 0 || stats.ValueStackMaxCapacity != 0 || stats.ValueStackCapacityGrowths != 0 || stats.CallFrameMaxDepth != 0 || len(stats.TopValueStackPeakSites) != 0 || len(stats.TopValueStackDeltaSites) != 0 || len(stats.TopCallOperandBalances) != 0 || len(stats.TopLoopBackedgeBalances) != 0 || len(stats.TopInlineFrameBalances) != 0 {
		t.Fatalf("stack diagnostics were not reset: %+v", stats)
	}
}

func TestBytecodeVMLoopFinalConditionalStatementKeepsValueStackBounded(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Loop(
			ast.Iff(ast.Bin(">=", ast.ID("i"), ast.Int(1024)), ast.Brk(nil, nil)),
			bytecodeConditionalIncrement(),
		),
		ast.ID("i"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
	stats := interp.BytecodeStats()
	if stats.ValueStackMaxDepth > 32 {
		t.Fatalf("value stack max depth = %d, want at most 32", stats.ValueStackMaxDepth)
	}
}

func TestBytecodeVMWhileFinalConditionalStatementKeepsValueStackBounded(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Wloop(
			ast.Bin("<", ast.ID("i"), ast.Int(1024)),
			bytecodeConditionalIncrement(),
		),
		ast.ID("i"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
	if stats := interp.BytecodeStats(); stats.ValueStackMaxDepth > 32 {
		t.Fatalf("value stack max depth = %d, want at most 32", stats.ValueStackMaxDepth)
	}
}

func TestBytecodeVMLoopNestedFinalConditionalStatementKeepsValueStackBounded(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	inc := func() ast.Statement {
		return ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(1)))
	}
	nested := ast.NewIfExpression(
		ast.Bin("<", ast.ID("i"), ast.Int(3072)),
		ast.Block(inc()),
		nil,
		ast.Block(inc()),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Loop(
			ast.Iff(ast.Bin(">=", ast.ID("i"), ast.Int(4096)), ast.Brk(nil, nil)),
			ast.NewIfExpression(
				ast.Bin("<", ast.ID("i"), ast.Int(1024)),
				ast.Block(inc()),
				nil,
				ast.Block(nested),
			),
		),
		ast.ID("i"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
	if stats := interp.BytecodeStats(); stats.ValueStackMaxDepth > 32 {
		t.Fatalf("value stack max depth = %d, want at most 32", stats.ValueStackMaxDepth)
	}
}

func bytecodeConditionalIncrement() *ast.IfExpression {
	return ast.NewIfExpression(
		ast.Bin("<", ast.ID("i"), ast.Int(512)),
		ast.Block(ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(1)))),
		nil,
		ast.Block(ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(1)))),
	)
}
