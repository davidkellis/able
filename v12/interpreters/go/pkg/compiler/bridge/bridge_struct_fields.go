package bridge

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// structNamedFieldValue reads either supported representation of a named Able
// struct at a compiler/runtime boundary.
func structNamedFieldValue(inst *runtime.StructInstanceValue, name string) (runtime.Value, bool) {
	if inst == nil || name == "" {
		return nil, false
	}
	if inst.Fields != nil {
		if value, ok := inst.Fields[name]; ok {
			return value, true
		}
	}
	if inst.Positional == nil || inst.Definition == nil || inst.Definition.Node == nil ||
		inst.Definition.Node.Kind != ast.StructKindNamed {
		return nil, false
	}
	idx, ok := inst.Definition.NamedFieldIndices[name]
	if !ok {
		for fieldIndex, field := range inst.Definition.Node.Fields {
			if field != nil && field.Name != nil && field.Name.Name == name {
				idx, ok = fieldIndex, true
				break
			}
		}
	}
	if !ok || idx < 0 || idx >= len(inst.Positional) {
		return nil, false
	}
	return inst.Positional[idx], true
}
