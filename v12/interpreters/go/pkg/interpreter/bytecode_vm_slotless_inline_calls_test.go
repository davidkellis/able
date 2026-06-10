package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_InlinesSlotlessResolvedFunctionFromCallName(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	fnDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	fnProgram := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpLoadName, name: "x", nameSimple: true},
			{op: bytecodeOpReturn},
		},
	})
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	setFunctionBytecodeProgram(fn, fnProgram)
	env.Define("id", fn)

	callNode := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(41)}, nil, false)
	callerProgram := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(41, runtime.IntegerI32)},
			{op: bytecodeOpCallName, name: "id", nameSimple: true, argCount: 1, node: callNode},
			{op: bytecodeOpReturn},
		},
	})
	got, err := newBytecodeVM(interp, env).run(callerProgram)
	if err != nil {
		t.Fatalf("bytecode run failed: %v", err)
	}
	want := runtime.NewSmallInt(41, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("slotless inline call result = %#v, want %#v", got, want)
	}
	stats := interp.BytecodeStats()
	if stats.CallNameInlineResolvedHits != 1 {
		t.Fatalf("CallNameInlineResolvedHits = %d, want 1; stats=%#v", stats.CallNameInlineResolvedHits, stats)
	}
	if stats.CallNameResolvedFunctionHits != 0 {
		t.Fatalf("CallNameResolvedFunctionHits = %d, want 0; stats=%#v", stats.CallNameResolvedFunctionHits, stats)
	}
}

func TestBytecodeVM_SlotlessInlineCallCoercesReturnType(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	fnDef := ast.Fn(
		"widen",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	fnProgram := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpLoadName, name: "x", nameSimple: true},
			{op: bytecodeOpReturn},
		},
	})
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	setFunctionBytecodeProgram(fn, fnProgram)
	env.Define("widen", fn)

	callNode := ast.NewFunctionCall(ast.ID("widen"), []ast.Expression{ast.Int(41)}, nil, false)
	callerProgram := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(41, runtime.IntegerI32)},
			{op: bytecodeOpCallName, name: "widen", nameSimple: true, argCount: 1, node: callNode},
			{op: bytecodeOpReturn},
		},
	})
	got, err := newBytecodeVM(interp, env).run(callerProgram)
	if err != nil {
		t.Fatalf("bytecode run failed: %v", err)
	}
	want := runtime.NewSmallInt(41, runtime.IntegerI64)
	if !valuesEqual(got, want) {
		t.Fatalf("slotless inline coerced return = %#v, want %#v", got, want)
	}
	stats := interp.BytecodeStats()
	if stats.CallNameInlineResolvedHits != 1 {
		t.Fatalf("CallNameInlineResolvedHits = %d, want 1; stats=%#v", stats.CallNameInlineResolvedHits, stats)
	}
}
