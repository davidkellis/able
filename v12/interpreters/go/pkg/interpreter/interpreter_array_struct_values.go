package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) arrayValueFromStructFields(fields map[string]runtime.Value) (*runtime.ArrayValue, error) {
	handle, length, capacity := arrayStructMetadataFromFields(fields)
	return i.arrayValueFromStructMetadata(handle, length, capacity)
}

func arrayStructMetadataFromFields(fields map[string]runtime.Value) (int64, int, int) {
	var handle int64
	var length int
	var capacity int
	if fields != nil {
		if hv, ok := fields["storage_handle"]; ok {
			if h, err := hostIntegerToInt64(hv); err == nil {
				handle = h
			}
		}
		if lv, ok := fields["length"]; ok {
			if l, err := arrayIndexFromValue(lv); err == nil {
				length = l
			}
		}
		if cv, ok := fields["capacity"]; ok {
			if c, err := arrayIndexFromValue(cv); err == nil {
				capacity = c
			}
		}
	}
	return handle, length, capacity
}

func (i *Interpreter) arrayValueFromStructFieldValues(fields []*ast.StructFieldDefinition, values []runtime.Value) (*runtime.ArrayValue, error) {
	handle, length, capacity := arrayStructMetadataFromFieldValues(fields, values)
	return i.arrayValueFromStructMetadata(handle, length, capacity)
}

func arrayStructMetadataFromFieldValues(fields []*ast.StructFieldDefinition, values []runtime.Value) (int64, int, int) {
	var handle int64
	var length int
	var capacity int
	for idx, field := range fields {
		if field == nil || field.Name == nil || idx >= len(values) {
			continue
		}
		switch field.Name.Name {
		case "storage_handle":
			if h, err := hostIntegerToInt64(values[idx]); err == nil {
				handle = h
			}
		case "length":
			if l, err := arrayIndexFromValue(values[idx]); err == nil {
				length = l
			}
		case "capacity":
			if c, err := arrayIndexFromValue(values[idx]); err == nil {
				capacity = c
			}
		}
	}
	return handle, length, capacity
}

func (i *Interpreter) arrayValueFromStructInstance(inst *runtime.StructInstanceValue) (*runtime.ArrayValue, error) {
	if inst == nil {
		return nil, nil
	}
	var handle int64
	var length int
	var capacity int
	if inst.Fields != nil {
		handle, length, capacity = arrayStructMetadataFromFields(inst.Fields)
	} else if inst.Definition != nil && inst.Definition.Node != nil {
		handle, length, capacity = arrayStructMetadataFromFieldValues(inst.Definition.Node.Fields, inst.Positional)
	}
	if handle != 0 {
		if err := runtime.ArrayStoreTrackStructInstanceLease(inst, handle); err != nil {
			return nil, err
		}
	}
	return i.arrayValueFromStructMetadata(handle, length, capacity)
}

func (i *Interpreter) arrayValueFromStructMetadata(handle int64, length int, capacity int) (*runtime.ArrayValue, error) {
	if capacity < length {
		capacity = length
	}
	if handle != 0 {
		return i.arrayValueFromHandle(handle, length, capacity)
	}
	return i.newArrayValue(make([]runtime.Value, length, capacity), capacity), nil
}
