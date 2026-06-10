package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeLoweringSkipsEnvScopeForPureIfBlocks(t *testing.T) {
	expr := ast.IfExpr(ast.Bool(true), ast.Block(ast.Int(1)))
	expr.ElseBody = ast.Block(ast.Int(2))

	program, err := NewBytecode().lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode: %v", err)
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpEnterScope); got != 0 {
		t.Fatalf("runtime EnterScope count = %d, want 0", got)
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpExitScope); got != 0 {
		t.Fatalf("runtime ExitScope count = %d, want 0", got)
	}
}

func TestBytecodeLoweringEmitsEnvScopeForIfBlockDeclaration(t *testing.T) {
	expr := ast.IfExpr(
		ast.Bool(true),
		ast.Block(
			ast.Assign(ast.ID("x"), ast.Int(1)),
			ast.ID("x"),
		),
	)
	expr.ElseBody = ast.Block(ast.Int(2))

	program, err := NewBytecode().lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode: %v", err)
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpEnterScope); got != 1 {
		t.Fatalf("runtime EnterScope count = %d, want 1", got)
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpExitScope); got != 1 {
		t.Fatalf("runtime ExitScope count = %d, want 1", got)
	}
	enter := bytecodeFirstRuntimeScopeInstruction(program, bytecodeOpEnterScope)
	if enter == nil {
		t.Fatalf("expected runtime EnterScope instruction")
	}
	if !enter.transientScope {
		t.Fatalf("expected local-only if block scope to be marked transient")
	}
}

func TestBytecodeLoweringEmitsRuntimeScopeOnlyForDirectEnvBlockInSlotFunction(t *testing.T) {
	ifExpr := ast.IfExpr(ast.Bool(true), ast.Block(ast.Int(1)))
	ifExpr.ElseBody = ast.Block(
		ast.StructDef("Marker", nil, ast.StructKindNamed, nil, nil, false),
		ast.Int(2),
	)
	def := ast.Fn(
		"pick",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{ifExpr},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecode: %v", err)
	}
	if program.frameLayout == nil {
		t.Fatalf("expected slot frame layout")
	}
	if !program.frameLayout.needsEnvScopes {
		t.Fatalf("expected layout to keep global env-scope flag for nested struct definition")
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpEnterScope); got != 1 {
		t.Fatalf("runtime EnterScope count = %d, want 1", got)
	}
	if got := bytecodeRuntimeScopeOpCount(program, bytecodeOpExitScope); got != 1 {
		t.Fatalf("runtime ExitScope count = %d, want 1", got)
	}
	enter := bytecodeFirstRuntimeScopeInstruction(program, bytecodeOpEnterScope)
	if enter == nil {
		t.Fatalf("expected runtime EnterScope instruction")
	}
	if enter.transientScope {
		t.Fatalf("expected direct-env block scope to remain non-transient")
	}
}

func TestBytecodeLoweringKeepsNonTransientEnvScopeForLambdaCaptureBlock(t *testing.T) {
	expr := ast.IfExpr(
		ast.Bool(true),
		ast.Block(
			ast.Assign(ast.ID("x"), ast.Int(7)),
			ast.Lam(nil, ast.ID("x")),
		),
	)
	expr.ElseBody = ast.Block(ast.Int(0))

	program, err := NewBytecode().lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode: %v", err)
	}
	enter := bytecodeFirstRuntimeScopeInstruction(program, bytecodeOpEnterScope)
	if enter == nil {
		t.Fatalf("expected runtime EnterScope instruction")
	}
	if enter.transientScope {
		t.Fatalf("expected lambda-capturing block scope to remain non-transient")
	}
}

func TestBytecodeVMLoopContinueWithoutRuntimeScopes(t *testing.T) {
	breakIf := ast.IfExpr(
		ast.Bin(">=", ast.ID("n"), ast.Int(3)),
		ast.Block(ast.Brk(nil, ast.ID("n"))),
	)
	loopBody := ast.Block(
		breakIf,
		ast.AssignOp(ast.AssignmentAssign, ast.ID("n"), ast.Bin("+", ast.ID("n"), ast.Int(1))),
		ast.Cont(nil),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("n"), ast.Int(0)),
		ast.Loop(loopBody.Body...),
	}, nil, nil)

	got := runBytecodeModule(t, module)
	intVal, ok := got.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("loop result = %#v, want 3", got)
	}
}

func bytecodeRuntimeScopeOpCount(program *bytecodeProgram, target bytecodeOp) int {
	if program == nil {
		return 0
	}
	count := 0
	for _, instr := range program.instructions {
		if instr.op == target && instr.argCount > 0 {
			count++
		}
		if instr.program != nil {
			count += bytecodeRuntimeScopeOpCount(instr.program, target)
		}
	}
	return count
}

func bytecodeFirstRuntimeScopeInstruction(program *bytecodeProgram, target bytecodeOp) *bytecodeInstruction {
	if program == nil {
		return nil
	}
	for idx := range program.instructions {
		instr := &program.instructions[idx]
		if instr.op == target && instr.argCount > 0 {
			return instr
		}
		if instr.program != nil {
			if nested := bytecodeFirstRuntimeScopeInstruction(instr.program, target); nested != nil {
				return nested
			}
		}
	}
	return nil
}
