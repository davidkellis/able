package runtime

import "able/interpreter-go/pkg/ast"

func NewStructInstancePositionalSized(def *StructDefinitionValue, fieldCount int, typeArgs []ast.TypeExpression) (*StructInstanceValue, []Value) {
	inst := &StructInstanceValue{
		Definition:    def,
		TypeArguments: typeArgs,
	}
	if fieldCount <= 0 {
		if fieldCount == 0 {
			inst.Positional = inst.inlinePositional[:0]
		}
		return inst, inst.Positional
	}
	if fieldCount <= len(inst.inlinePositional) {
		inst.Positional = inst.inlinePositional[:fieldCount]
		return inst, inst.Positional
	}
	inst.Positional = make([]Value, fieldCount)
	return inst, inst.Positional
}
