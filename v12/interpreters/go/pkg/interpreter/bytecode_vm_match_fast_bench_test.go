package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func BenchmarkBytecodeVMJumpIfNotTypedPattern(b *testing.B) {
	bench := func(b *testing.B, planned bool) {
		interp := NewBytecode()
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
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
		value := &runtime.StructInstanceValue{
			Definition:    def,
			Fields:        map[string]runtime.Value{"value": runtime.NewSmallInt(7, runtime.IntegerI32)},
			TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
		}
		instr := &bytecodeInstruction{
			op:       bytecodeOpJumpIfNotTypedPattern,
			typeExpr: ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			target:   3,
		}
		if planned {
			instr.genericStructMatch = bytecodeGenericStructPatternPlanForTypeExprWithDefinition(instr.typeExpr, def)
		}

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			vm.stack = vm.stack[:0]
			vm.stack = append(vm.stack, value)
			vm.ip = 0
			if err := vm.execJumpIfNotTypedPattern(instr); err != nil {
				b.Fatalf("execJumpIfNotTypedPattern: %v", err)
			}
			if vm.ip != 1 || len(vm.stack) != 1 || vm.stack[0] != value {
				b.Fatalf("unexpected match state: ip=%d stack=%#v", vm.ip, vm.stack)
			}
		}
	}

	b.Run("generic_named_struct", func(b *testing.B) {
		bench(b, false)
	})
	b.Run("generic_named_struct_planned", func(b *testing.B) {
		bench(b, true)
	})
}
