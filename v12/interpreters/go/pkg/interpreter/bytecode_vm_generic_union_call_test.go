package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeGenericUnionCallCacheInvalidatesMemberChanges(t *testing.T) {
	root := runtime.NewEnvironment(nil)
	callable := &runtime.FunctionValue{}
	root.Define("map", callable)
	interp := NewBytecode()
	interp.global = root
	interp.envSingleThread = true
	vm := newBytecodeVM(interp, root)
	program := &bytecodeProgram{}

	if _, ok := vm.storeCachedGenericUnionCall(program, 3, "map", callable); !ok {
		t.Fatal("expected cache store")
	}
	if _, ok := vm.lookupCachedGenericUnionCall(program, 3, "map"); !ok {
		t.Fatal("expected initial cache hit")
	}
	root.Assign("map", &runtime.FunctionValue{})
	if _, ok := vm.lookupCachedGenericUnionCall(program, 3, "map"); ok {
		t.Fatal("owner assignment must invalidate cached method")
	}

	root.Define("map", callable)
	if _, ok := vm.storeCachedGenericUnionCall(program, 3, "map", callable); !ok {
		t.Fatal("expected cache restore")
	}
	child := runtime.NewEnvironment(root)
	child.Define("map", &runtime.FunctionValue{})
	vm.env = child
	if _, ok := vm.lookupCachedGenericUnionCall(program, 3, "map"); ok {
		t.Fatal("lexical shadow must invalidate cached method")
	}
}

func TestBytecodeLoweringUsesGenericUnionOpcodeOnlyForProvenCall(t *testing.T) {
	call := ast.NewFunctionCall(ast.Member(ast.Int(1), "map"), nil, nil, false)
	member := call.Callee.(*ast.MemberAccessExpression)
	interp := NewBytecode()
	interp.bytecodeMethodSelections = bytecodeMethodSelections{
		member: {GenericNamedUnion: true},
	}
	if !interp.bytecodeGenericUnionMethodCallProven(call) {
		t.Fatal("test selection was not recognized")
	}
	program, err := interp.lowerExpressionToBytecode(call)
	if err != nil {
		t.Fatalf("lower proven call: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallGenericUnionMember) {
		t.Fatalf("proven call did not receive generic-union opcode %d: %#v", bytecodeOpCallGenericUnionMember, program.instructions)
	}

	interp.bytecodeMethodSelections = nil
	ordinaryCall := ast.NewFunctionCall(ast.Member(ast.Int(1), "map"), nil, nil, false)
	program, err = interp.lowerExpressionToBytecode(ordinaryCall)
	if err != nil {
		t.Fatalf("lower ordinary call: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCallGenericUnionMember) {
		t.Fatal("ordinary call received generic-union opcode")
	}
}
