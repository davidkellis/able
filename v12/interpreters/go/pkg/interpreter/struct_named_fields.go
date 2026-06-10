package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"sync"
)

var (
	structDefinitionFieldIndexCacheMu sync.RWMutex
	structDefinitionFieldIndexCache   = make(map[*ast.StructDefinition]map[string]int)
)

func structDefinitionFieldIndices(def *ast.StructDefinition) map[string]int {
	if def == nil {
		return nil
	}
	structDefinitionFieldIndexCacheMu.RLock()
	if cached, ok := structDefinitionFieldIndexCache[def]; ok {
		structDefinitionFieldIndexCacheMu.RUnlock()
		return cached
	}
	structDefinitionFieldIndexCacheMu.RUnlock()
	created := buildStructDefinitionFieldIndices(def)
	structDefinitionFieldIndexCacheMu.Lock()
	if existing, ok := structDefinitionFieldIndexCache[def]; ok {
		structDefinitionFieldIndexCacheMu.Unlock()
		return existing
	}
	structDefinitionFieldIndexCache[def] = created
	structDefinitionFieldIndexCacheMu.Unlock()
	return created
}

func buildStructDefinitionFieldIndices(def *ast.StructDefinition) map[string]int {
	if def == nil {
		return nil
	}
	created := make(map[string]int, len(def.Fields))
	for idx, field := range def.Fields {
		if field == nil || field.Name == nil || field.Name.Name == "" {
			continue
		}
		created[field.Name.Name] = idx
	}
	return created
}

func newStructDefinitionValue(def *ast.StructDefinition) *runtime.StructDefinitionValue {
	val := &runtime.StructDefinitionValue{Node: def}
	if def != nil && def.Kind == ast.StructKindNamed && len(def.Fields) > 4 {
		val.NamedFieldIndices = buildStructDefinitionFieldIndices(def)
	}
	return val
}

func structDefinitionNamedFieldIndex(def *ast.StructDefinition, name string) (int, bool) {
	if def == nil || name == "" {
		return 0, false
	}
	switch len(def.Fields) {
	case 1:
		field := def.Fields[0]
		if field != nil && field.Name != nil && field.Name.Name == name {
			return 0, true
		}
		return 0, false
	case 2:
		field0 := def.Fields[0]
		if field0 != nil && field0.Name != nil && field0.Name.Name == name {
			return 0, true
		}
		field1 := def.Fields[1]
		if field1 != nil && field1.Name != nil && field1.Name.Name == name {
			return 1, true
		}
		return 0, false
	case 3:
		field0 := def.Fields[0]
		if field0 != nil && field0.Name != nil && field0.Name.Name == name {
			return 0, true
		}
		field1 := def.Fields[1]
		if field1 != nil && field1.Name != nil && field1.Name.Name == name {
			return 1, true
		}
		field2 := def.Fields[2]
		if field2 != nil && field2.Name != nil && field2.Name.Name == name {
			return 2, true
		}
		return 0, false
	case 4:
		field0 := def.Fields[0]
		if field0 != nil && field0.Name != nil && field0.Name.Name == name {
			return 0, true
		}
		field1 := def.Fields[1]
		if field1 != nil && field1.Name != nil && field1.Name.Name == name {
			return 1, true
		}
		field2 := def.Fields[2]
		if field2 != nil && field2.Name != nil && field2.Name.Name == name {
			return 2, true
		}
		field3 := def.Fields[3]
		if field3 != nil && field3.Name != nil && field3.Name.Name == name {
			return 3, true
		}
		return 0, false
	}
	idx, ok := structDefinitionFieldIndices(def)[name]
	return idx, ok
}

func structUsesNamedFieldStorage(inst *runtime.StructInstanceValue) bool {
	if inst == nil {
		return false
	}
	if inst.Fields != nil {
		return true
	}
	if inst.Definition == nil || inst.Definition.Node == nil {
		return false
	}
	return inst.Definition.Node.Kind == ast.StructKindNamed
}

func newNamedStructInstancePositionalStorage(def *runtime.StructDefinitionValue, typeArgs []ast.TypeExpression) (*runtime.StructInstanceValue, bool) {
	if def == nil || def.Node == nil || def.Node.Kind != ast.StructKindNamed {
		return nil, false
	}
	inst, _ := runtime.NewStructInstancePositionalSized(def, len(def.Node.Fields), typeArgs)
	return inst, true
}

func structNamedFieldIndex(inst *runtime.StructInstanceValue, name string) (int, bool) {
	if inst == nil || name == "" || !structUsesNamedFieldStorage(inst) {
		return 0, false
	}
	if inst.Definition == nil || inst.Definition.Node == nil {
		return 0, false
	}
	if inst.Definition.NamedFieldIndices != nil {
		idx, ok := inst.Definition.NamedFieldIndices[name]
		return idx, ok
	}
	return structDefinitionNamedFieldIndex(inst.Definition.Node, name)
}

func structNamedFieldValue(inst *runtime.StructInstanceValue, name string) (runtime.Value, bool) {
	if inst == nil || name == "" {
		return nil, false
	}
	if inst.Fields != nil {
		val, ok := inst.Fields[name]
		return val, ok
	}
	if inst.Positional == nil {
		return nil, false
	}
	if !structUsesNamedFieldStorage(inst) {
		return nil, false
	}
	if inst.Definition == nil || inst.Definition.Node == nil {
		return nil, false
	}
	idx, ok := 0, false
	if inst.Definition.NamedFieldIndices != nil {
		idx, ok = inst.Definition.NamedFieldIndices[name]
	} else {
		idx, ok = structDefinitionNamedFieldIndex(inst.Definition.Node, name)
	}
	if !ok || idx < 0 || idx >= len(inst.Positional) {
		return nil, false
	}
	return inst.Positional[idx], true
}

func structSetNamedFieldValue(inst *runtime.StructInstanceValue, name string, value runtime.Value) bool {
	if inst == nil || name == "" {
		return false
	}
	if inst.Fields != nil {
		if _, ok := inst.Fields[name]; !ok {
			return false
		}
		inst.Fields[name] = value
		return true
	}
	if inst.Positional == nil {
		return false
	}
	if !structUsesNamedFieldStorage(inst) {
		return false
	}
	idx, ok := structNamedFieldIndex(inst, name)
	if !ok || idx < 0 || idx >= len(inst.Positional) {
		return false
	}
	inst.Positional[idx] = value
	return true
}

func structNamedFieldCount(inst *runtime.StructInstanceValue) int {
	if inst == nil || !structUsesNamedFieldStorage(inst) {
		return 0
	}
	if inst.Fields != nil {
		return len(inst.Fields)
	}
	if inst.Positional != nil {
		return len(inst.Positional)
	}
	return 0
}

func structCopyNamedFields(inst *runtime.StructInstanceValue) (map[string]runtime.Value, bool) {
	if inst == nil || !structUsesNamedFieldStorage(inst) {
		return nil, false
	}
	if inst.Fields != nil {
		copied := make(map[string]runtime.Value, len(inst.Fields))
		for name, value := range inst.Fields {
			copied[name] = value
		}
		return copied, true
	}
	def := inst.Definition
	if def == nil || def.Node == nil || inst.Positional == nil {
		return nil, false
	}
	copied := make(map[string]runtime.Value, len(def.Node.Fields))
	for _, field := range def.Node.Fields {
		if field == nil || field.Name == nil {
			continue
		}
		value, ok := structNamedFieldValue(inst, field.Name.Name)
		if !ok {
			return nil, false
		}
		copied[field.Name.Name] = value
	}
	return copied, true
}
