package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBytecodeVM_LoweringSkipsFloatStoreFusionForIntegerReassignments(t *testing.T) {
	ops := []string{"+", "-", "*"}

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			def := ast.Fn(
				"step",
				[]*ast.FunctionParameter{
					ast.Param("a", ast.Ty("i32")),
					ast.Param("b", ast.Ty("i32")),
				},
				[]ast.Statement{
					ast.Assign(ast.ID("x"), ast.ID("a")),
					ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Bin(op, ast.ID("x"), ast.ID("b"))),
					ast.ID("x"),
				},
				ast.Ty("i32"),
				nil,
				nil,
				false,
				false,
			)

			program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
			if err != nil {
				t.Fatalf("bytecode lowering failed: %v", err)
			}
			if bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatBinary) {
				t.Fatalf("unexpected float binary store fusion for integer reassignment")
			}

			module := ast.Mod([]ast.Statement{
				def,
				ast.Call("step", ast.Int(9), ast.Int(4)),
			}, nil, nil)

			want := mustEvalModule(t, New(), module)
			got := runBytecodeModule(t, module)
			if !valuesEqual(got, want) {
				t.Fatalf("bytecode integer reassignment mismatch: got=%#v want=%#v", got, want)
			}
		})
	}
}
