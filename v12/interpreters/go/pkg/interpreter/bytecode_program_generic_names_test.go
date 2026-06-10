package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeInlineReturnGenericNamesSkipsUnusedReturnGenerics(t *testing.T) {
	names := map[string]struct{}{"T": {}}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnTypeUsesGenerics: false,
		},
		returnGenericNames:       names,
		returnGenericNamesCached: true,
	}
	if got := bytecodeInlineReturnGenericNames(nil, program); got != nil {
		t.Fatalf("unused return generics = %#v, want nil", got)
	}

	program.frameLayout.returnTypeUsesGenerics = true
	if got := bytecodeInlineReturnGenericNames(nil, program); len(got) != 1 {
		t.Fatalf("used return generics = %#v, want cached map", got)
	} else if _, ok := got["T"]; !ok {
		t.Fatalf("used return generics = %#v, want T", got)
	}
}

func TestBytecodeInlineReturnGenericNamesSkipsUnusedSlotlessReturnGenerics(t *testing.T) {
	fnDef := ast.Fn(
		"concrete",
		nil,
		[]ast.Statement{ast.Int(1)},
		ast.Ty("i32"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	program := &bytecodeProgram{}
	fn := &runtime.FunctionValue{Declaration: fnDef}
	setFunctionBytecodeProgram(fn, program)

	if !program.returnTypeMetadataCached {
		t.Fatalf("expected return metadata to be cached")
	}
	if program.returnTypeUsesGenerics {
		t.Fatalf("returnTypeUsesGenerics = true, want false")
	}
	if got := bytecodeInlineReturnGenericNames(fn, program); got != nil {
		t.Fatalf("unused slotless return generics = %#v, want nil", got)
	}
}

func TestBytecodeInlineReturnGenericNamesKeepsUsedSlotlessReturnGenerics(t *testing.T) {
	fnDef := ast.Fn(
		"identity",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	program := &bytecodeProgram{}
	fn := &runtime.FunctionValue{Declaration: fnDef}
	setFunctionBytecodeProgram(fn, program)

	if !program.returnTypeMetadataCached {
		t.Fatalf("expected return metadata to be cached")
	}
	if !program.returnTypeUsesGenerics {
		t.Fatalf("returnTypeUsesGenerics = false, want true")
	}
	if got := bytecodeInlineReturnGenericNames(fn, program); len(got) != 1 {
		t.Fatalf("used slotless return generics = %#v, want cached map", got)
	} else if _, ok := got["T"]; !ok {
		t.Fatalf("used slotless return generics = %#v, want T", got)
	}
}
