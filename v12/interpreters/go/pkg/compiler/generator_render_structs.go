package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderStructs(buf *bytes.Buffer) {
	if len(g.structs) == 0 {
		return
	}
	for _, name := range g.sortedStructNames() {
		info := g.structs[name]
		if info == nil {
			continue
		}
		fmt.Fprintf(buf, "type %s struct {\n", info.GoName)
		for _, field := range info.Fields {
			fmt.Fprintf(buf, "\t%s %s\n", field.GoName, field.GoType)
		}
		if info.Name == "Array" {
			fmt.Fprintf(buf, "\tElements []runtime.Value\n")
		}
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderStructConverters(buf *bytes.Buffer) {
	if len(g.structs) == 0 {
		return
	}
	for _, name := range g.sortedStructNames() {
		info := g.structs[name]
		if info == nil {
			continue
		}
		g.renderStructFrom(buf, info)
		g.renderStructTo(buf, info)
		g.renderStructApply(buf, info)
		if info.Name == "Array" {
			g.renderArrayStructHelpers(buf)
		}
	}
}

func (g *generator) renderStructFrom(buf *bytes.Buffer, info *structInfo) {
	fmt.Fprintf(buf, "func __able_struct_%s_from(value runtime.Value) (*%s, error) {\n", info.GoName, info.GoName)
	fmt.Fprintf(buf, "\tout := &%s{}\n", info.GoName)
	fmt.Fprintf(buf, "\tcurrent := __able_unwrap_interface(value)\n")
	fmt.Fprintf(buf, "\tif _, isNil := current.(runtime.NilValue); isNil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	if info.Name == "Array" {
		fmt.Fprintf(buf, "\tif raw, ok, nilPtr := __able_runtime_array_value(current); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif raw.Handle != 0 {\n")
		fmt.Fprintf(buf, "\t\t\tstate, err := runtime.ArrayStoreState(raw.Handle)\n")
		fmt.Fprintf(buf, "\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tout.Elements = make([]runtime.Value, len(state.Values), state.Capacity)\n")
		fmt.Fprintf(buf, "\t\t\tcopy(out.Elements, state.Values)\n")
		fmt.Fprintf(buf, "\t\t} else {\n")
		fmt.Fprintf(buf, "\t\t\tout.Elements = make([]runtime.Value, len(raw.Elements), cap(raw.Elements))\n")
		fmt.Fprintf(buf, "\t\t\tcopy(out.Elements, raw.Elements)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tout.Storage_handle = raw.Handle\n")
		fmt.Fprintf(buf, "\t\t__able_struct_Array_sync(out)\n")
		fmt.Fprintf(buf, "\t\treturn out, nil\n")
		fmt.Fprintf(buf, "\t}\n")
		// Struct instance fallback: interpreter-created Array struct instances use storage_handle.
		fmt.Fprintf(buf, "\tif inst, ok := current.(*runtime.StructInstanceValue); ok && inst != nil {\n")
		fmt.Fprintf(buf, "\t\tif inst.Definition != nil && inst.Definition.Node != nil && inst.Definition.Node.ID != nil && inst.Definition.Node.ID.Name == \"Array\" {\n")
		fmt.Fprintf(buf, "\t\t\tif lengthVal, ok := inst.Fields[\"length\"]; ok {\n")
		fmt.Fprintf(buf, "\t\t\t\tlength, err := bridge.AsInt(lengthVal, 32)\n")
		fmt.Fprintf(buf, "\t\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t\tout.Length = int32(length)\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tif capacityVal, ok := inst.Fields[\"capacity\"]; ok {\n")
		fmt.Fprintf(buf, "\t\t\t\tcapacity, err := bridge.AsInt(capacityVal, 32)\n")
		fmt.Fprintf(buf, "\t\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t\tout.Capacity = int32(capacity)\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tif handleVal, ok := inst.Fields[\"storage_handle\"]; ok {\n")
		fmt.Fprintf(buf, "\t\t\t\thandle, err := __able_array_handle_from_value(handleVal)\n")
		fmt.Fprintf(buf, "\t\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t\tout.Storage_handle = handle\n")
		fmt.Fprintf(buf, "\t\t\t\tif handle != 0 {\n")
		fmt.Fprintf(buf, "\t\t\t\t\tstate, err := runtime.ArrayStoreState(handle)\n")
		fmt.Fprintf(buf, "\t\t\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t\t\tout.Elements = make([]runtime.Value, len(state.Values), state.Capacity)\n")
		fmt.Fprintf(buf, "\t\t\t\t\tcopy(out.Elements, state.Values)\n")
		fmt.Fprintf(buf, "\t\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tif out.Storage_handle == 0 && out.Capacity < out.Length {\n")
		fmt.Fprintf(buf, "\t\t\t\tout.Capacity = out.Length\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tif out.Storage_handle == 0 && out.Capacity > 0 {\n")
		fmt.Fprintf(buf, "\t\t\t\tout.Elements = make([]runtime.Value, int(out.Length), int(out.Capacity))\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\t__able_struct_Array_sync(out)\n")
		fmt.Fprintf(buf, "\t\t\treturn out, nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn out, fmt.Errorf(\"expected Array value\")\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.Name == "IteratorEnd" {
		fmt.Fprintf(buf, "\tif _, ok := current.(runtime.IteratorEndValue); ok {\n")
		fmt.Fprintf(buf, "\t\treturn out, nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif raw, ok := current.(*runtime.IteratorEndValue); ok {\n")
		fmt.Fprintf(buf, "\t\tif raw == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn out, nil\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	if info.Kind == ast.StructKindSingleton || (info.Kind != ast.StructKindPositional && len(info.Fields) == 0) {
		fmt.Fprintf(buf, "\tif def, ok, nilPtr := __able_runtime_struct_definition_value(current); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != %q {\n", info.Name)
		fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn out, nil\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tinst, ok := current.(*runtime.StructInstanceValue)\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != %q {\n", info.Name)
	fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	if len(info.Fields) == 0 {
		fmt.Fprintf(buf, "\treturn out, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.Kind == ast.StructKindPositional {
		fmt.Fprintf(buf, "\tif len(inst.Positional) < %d {\n", len(info.Fields))
		fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"missing positional fields for %s\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		for idx, field := range info.Fields {
			g.renderFieldFromPositional(buf, field, idx)
		}
	} else {
		fmt.Fprintf(buf, "\tif inst.Fields == nil {\n")
		fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"missing fields for %s\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		for _, field := range info.Fields {
			g.renderFieldFromNamed(buf, field)
		}
	}
	fmt.Fprintf(buf, "\treturn out, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderFieldFromNamed(buf *bytes.Buffer, field fieldInfo) {
	fmt.Fprintf(buf, "\t{\n")
	fmt.Fprintf(buf, "\t\tfieldValue, ok := inst.Fields[%q]\n", field.Name)
	fmt.Fprintf(buf, "\t\tif !ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"missing field %s\")\n", field.Name)
	fmt.Fprintf(buf, "\t\t}\n")
	g.renderValueConversion(buf, "\t\t", "fieldValue", field.GoType, "out."+field.GoName, "out")
	fmt.Fprintf(buf, "\t}\n")
}

func (g *generator) renderFieldFromPositional(buf *bytes.Buffer, field fieldInfo, idx int) {
	fmt.Fprintf(buf, "\t{\n")
	fmt.Fprintf(buf, "\t\tfieldValue := inst.Positional[%d]\n", idx)
	g.renderValueConversion(buf, "\t\t", "fieldValue", field.GoType, "out."+field.GoName, "out")
	fmt.Fprintf(buf, "\t}\n")
}

func (g *generator) renderStructTo(buf *bytes.Buffer, info *structInfo) {
	fmt.Fprintf(buf, "func __able_struct_%s_to(rt *bridge.Runtime, value *%s) (runtime.Value, error) {\n", info.GoName, info.GoName)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing %s value\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	if info.Name == "Array" {
		fmt.Fprintf(buf, "\tarr, err := __able_struct_Array_runtime_value(value, value.Storage_handle)\n")
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn arr, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.Name == "IteratorEnd" {
		fmt.Fprintf(buf, "\treturn runtime.IteratorEnd, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\tdef, err := rt.StructDefinition(%q)\n", info.Name)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	if info.Kind == ast.StructKindPositional {
		fmt.Fprintf(buf, "\tvalues := make([]runtime.Value, 0, %d)\n", len(info.Fields))
		for _, field := range info.Fields {
			g.renderValueToRuntime(buf, "value."+field.GoName, field.GoType, "values")
		}
		fmt.Fprintf(buf, "\treturn &runtime.StructInstanceValue{Definition: def, Positional: values}, nil\n")
	} else {
		fmt.Fprintf(buf, "\tfields := make(map[string]runtime.Value, %d)\n", len(info.Fields))
		for _, field := range info.Fields {
			g.renderValueToRuntimeNamed(buf, "value."+field.GoName, field.GoType, field.Name)
		}
		fmt.Fprintf(buf, "\treturn &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil\n")
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderStructApply(buf *bytes.Buffer, info *structInfo) {
	fmt.Fprintf(buf, "func __able_struct_%s_apply(rt *bridge.Runtime, target runtime.Value, value *%s) error {\n", info.GoName, info.GoName)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"missing %s value\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\ttargetCurrent := __able_unwrap_interface(target)\n")
	if info.Name == "Array" {
		fmt.Fprintf(buf, "\t__able_struct_Array_sync(value)\n")
		fmt.Fprintf(buf, "\tif raw, ok, nilPtr := __able_runtime_array_value(targetCurrent); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tpreferredHandle := raw.Handle\n")
		fmt.Fprintf(buf, "\t\tif preferredHandle == 0 {\n")
		fmt.Fprintf(buf, "\t\t\tpreferredHandle = value.Storage_handle\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tarr, err := __able_struct_Array_runtime_value(value, preferredHandle)\n")
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\traw.Handle = arr.Handle\n")
		fmt.Fprintf(buf, "\t\traw.Elements = arr.Elements\n")
		fmt.Fprintf(buf, "\t\tvalue.Storage_handle = arr.Handle\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tinst, ok := targetCurrent.(*runtime.StructInstanceValue)\n")
		fmt.Fprintf(buf, "\tif !ok || inst == nil {\n")
		fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != %q {\n", info.Name)
		fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tpreferredHandle := value.Storage_handle\n")
		fmt.Fprintf(buf, "\tif handleVal, ok := inst.Fields[\"storage_handle\"]; ok {\n")
		fmt.Fprintf(buf, "\t\thandle, herr := __able_array_handle_from_value(handleVal)\n")
		fmt.Fprintf(buf, "\t\tif herr == nil && handle != 0 {\n")
		fmt.Fprintf(buf, "\t\t\tpreferredHandle = handle\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif preferredHandle == 0 {\n")
		fmt.Fprintf(buf, "\t\tpreferredHandle = runtime.ArrayStoreNewWithCapacity(__able_struct_Array_capacity_hint(value))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tarr, err := __able_struct_Array_runtime_value(value, preferredHandle)\n")
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn err\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif inst.Fields == nil {\n")
		fmt.Fprintf(buf, "\t\tinst.Fields = make(map[string]runtime.Value, 3)\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tinst.Fields[\"length\"] = bridge.ToInt(int64(len(arr.Elements)), runtime.IntegerI32)\n")
		fmt.Fprintf(buf, "\tinst.Fields[\"capacity\"] = bridge.ToInt(int64(cap(arr.Elements)), runtime.IntegerI32)\n")
		fmt.Fprintf(buf, "\tinst.Fields[\"storage_handle\"] = bridge.ToInt(arr.Handle, runtime.IntegerI64)\n")
		fmt.Fprintf(buf, "\tvalue.Storage_handle = arr.Handle\n")
		fmt.Fprintf(buf, "\treturn nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.Name == "IteratorEnd" {
		fmt.Fprintf(buf, "\tif _, ok := targetCurrent.(runtime.IteratorEndValue); ok {\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif raw, ok := targetCurrent.(*runtime.IteratorEndValue); ok {\n")
		fmt.Fprintf(buf, "\t\tif raw == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tinst, ok := targetCurrent.(*runtime.StructInstanceValue)\n")
	if info.Kind == ast.StructKindSingleton || (info.Kind != ast.StructKindPositional && len(info.Fields) == 0) {
		fmt.Fprintf(buf, "\tif def, ok, nilPtr := __able_runtime_struct_definition_value(targetCurrent); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != %q {\n", info.Name)
		fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tif !ok || inst == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != %q {\n", info.Name)
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tconverted, err := __able_struct_%s_to(rt, value)\n", info.GoName)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tupdated, ok := converted.(*runtime.StructInstanceValue)\n")
	fmt.Fprintf(buf, "\tif !ok || updated == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tinst.Definition = updated.Definition\n")
	fmt.Fprintf(buf, "\tinst.Fields = updated.Fields\n")
	fmt.Fprintf(buf, "\tinst.Positional = updated.Positional\n")
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderArrayStructHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_struct_Array_sync(value *Array) {\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvalue.Length = int32(len(value.Elements))\n")
	fmt.Fprintf(buf, "\tvalue.Capacity = int32(cap(value.Elements))\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_struct_Array_capacity_hint(value *Array) int {\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tcapHint := cap(value.Elements)\n")
	fmt.Fprintf(buf, "\tif capHint < int(value.Capacity) {\n")
	fmt.Fprintf(buf, "\t\tcapHint = int(value.Capacity)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif capHint < len(value.Elements) {\n")
	fmt.Fprintf(buf, "\t\tcapHint = len(value.Elements)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn capHint\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_struct_Array_clone_elements(values []runtime.Value, capacityHint int) []runtime.Value {\n")
	fmt.Fprintf(buf, "\tif capacityHint < len(values) {\n")
	fmt.Fprintf(buf, "\t\tcapacityHint = len(values)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tcloned := make([]runtime.Value, len(values), capacityHint)\n")
	fmt.Fprintf(buf, "\tcopy(cloned, values)\n")
	fmt.Fprintf(buf, "\treturn cloned\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_struct_Array_runtime_value(value *Array, preferredHandle int64) (*runtime.ArrayValue, error) {\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing Array value\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t__able_struct_Array_sync(value)\n")
	fmt.Fprintf(buf, "\tcapHint := __able_struct_Array_capacity_hint(value)\n")
	fmt.Fprintf(buf, "\telems := __able_struct_Array_clone_elements(value.Elements, capHint)\n")
	fmt.Fprintf(buf, "\tif preferredHandle == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn &runtime.ArrayValue{Elements: elems}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tstate, err := runtime.ArrayStoreEnsureHandle(preferredHandle, len(elems), cap(elems))\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tstate.Values = elems\n")
	fmt.Fprintf(buf, "\tstate.Capacity = cap(elems)\n")
	fmt.Fprintf(buf, "\treturn &runtime.ArrayValue{Elements: state.Values, Handle: preferredHandle}, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderValueConversion(buf *bytes.Buffer, indent, valueVar, goType, assignTarget, returnExpr string) {
	if goType == "runtime.ErrorValue" {
		fmt.Fprintf(buf, "%sconverted, ok, nilPtr := __able_runtime_error_value(%s)\n", indent, valueVar)
		fmt.Fprintf(buf, "%sif !ok || nilPtr {\n", indent)
		fmt.Fprintf(buf, "%s\tconverted = bridge.ErrorValue(__able_runtime, %s)\n", indent, valueVar)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
		return
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%sconverted, err := %s(%s)\n", indent, spec.FromRuntimeHelper, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
		return
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%sconverted, err := %s(%s)\n", indent, helper, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%sconverted, err := %s(__able_runtime, %s)\n", indent, callable.FromRuntimeHelper, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "%s%s = %s\n", indent, assignTarget, valueVar)
	case "bool":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsBool(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "string":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsString(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "rune":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsRune(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "float32":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsFloat(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = float32(convertedRaw)\n", indent, assignTarget)
	case "float64":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsFloat(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "int":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = int(convertedRaw)\n", indent, assignTarget)
	case "uint":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = uint(convertedRaw)\n", indent, assignTarget)
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsInt(%s, %d)\n", indent, valueVar, bits)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = %s(convertedRaw)\n", indent, assignTarget, goType)
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsUint(%s, %d)\n", indent, valueVar, bits)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = %s(convertedRaw)\n", indent, assignTarget, goType)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "%sconverted, err := __able_struct_%s_from(%s)\n", indent, baseName, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "union":
		info := g.nativeUnionInfoForGoType(goType)
		if info == nil {
			fmt.Fprintf(buf, "%sreturn %s, fmt.Errorf(\"unsupported field type\")\n", indent, returnExpr)
			return
		}
		fmt.Fprintf(buf, "%sconverted, err := %s(__able_runtime, %s)\n", indent, info.FromRuntimeHelper, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "interface":
		info := g.nativeInterfaceInfoForGoType(goType)
		if info == nil {
			fmt.Fprintf(buf, "%sreturn %s, fmt.Errorf(\"unsupported field type\")\n", indent, returnExpr)
			return
		}
		fmt.Fprintf(buf, "%sconverted, err := %s(__able_runtime, %s)\n", indent, info.FromRuntimeHelper, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "any":
		fmt.Fprintf(buf, "%s%s = %s\n", indent, assignTarget, valueVar)
	default:
		fmt.Fprintf(buf, "%sreturn %s, fmt.Errorf(\"unsupported field type\")\n", indent, returnExpr)
	}
}

func (g *generator) renderValueToRuntime(buf *bytes.Buffer, valueExpr, goType, targetSlice string) {
	if goType == "runtime.ErrorValue" {
		fmt.Fprintf(buf, "\t%s = append(%s, %s)\n", targetSlice, targetSlice, valueExpr)
		return
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", spec.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = append(%s, converted)\n", targetSlice, targetSlice)
		fmt.Fprintf(buf, "\t}\n")
		return
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\t%s = append(%s, %s(%s))\n", targetSlice, targetSlice, helper, valueExpr)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", callable.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = append(%s, converted)\n", targetSlice, targetSlice)
		fmt.Fprintf(buf, "\t}\n")
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t%s = append(%s, %s)\n", targetSlice, targetSlice, valueExpr)
	case "bool":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToBool(%s))\n", targetSlice, targetSlice, valueExpr)
	case "string":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToString(%s))\n", targetSlice, targetSlice, valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToRune(%s))\n", targetSlice, targetSlice, valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToFloat32(%s))\n", targetSlice, targetSlice, valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToFloat64(%s))\n", targetSlice, targetSlice, valueExpr)
	case "int":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")))\n", targetSlice, targetSlice, valueExpr)
	case "int8":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")))\n", targetSlice, targetSlice, valueExpr)
	case "int16":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")))\n", targetSlice, targetSlice, valueExpr)
	case "int32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")))\n", targetSlice, targetSlice, valueExpr)
	case "int64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint8":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint16":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")))\n", targetSlice, targetSlice, valueExpr)
	case "struct":
		fmt.Fprintf(buf, "\t%s = append(%s, __able_any_to_value(%s))\n", targetSlice, targetSlice, valueExpr)
	case "interface":
		info := g.nativeInterfaceInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = append(%s, converted)\n", targetSlice, targetSlice)
		fmt.Fprintf(buf, "\t}\n")
	case "union":
		info := g.nativeUnionInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = append(%s, converted)\n", targetSlice, targetSlice)
		fmt.Fprintf(buf, "\t}\n")
	case "any":
		fmt.Fprintf(buf, "\t%s = append(%s, __able_any_to_value(%s))\n", targetSlice, targetSlice, valueExpr)
	}
}

func (g *generator) renderValueToRuntimeAssign(buf *bytes.Buffer, valueExpr, goType, targetExpr string) {
	if goType == "runtime.ErrorValue" {
		fmt.Fprintf(buf, "\t\t%s = %s\n", targetExpr, valueExpr)
		return
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", spec.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = converted\n", targetExpr)
		return
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\t\t%s = %s(%s)\n", targetExpr, helper, valueExpr)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", callable.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = converted\n", targetExpr)
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t\t%s = %s\n", targetExpr, valueExpr)
	case "bool":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToBool(%s)\n", targetExpr, valueExpr)
	case "string":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToString(%s)\n", targetExpr, valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToRune(%s)\n", targetExpr, valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToFloat32(%s)\n", targetExpr, valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToFloat64(%s)\n", targetExpr, valueExpr)
	case "int":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))\n", targetExpr, valueExpr)
	case "uint":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))\n", targetExpr, valueExpr)
	case "int8":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\"))\n", targetExpr, valueExpr)
	case "int16":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\"))\n", targetExpr, valueExpr)
	case "int32":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\"))\n", targetExpr, valueExpr)
	case "int64":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\"))\n", targetExpr, valueExpr)
	case "uint8":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\"))\n", targetExpr, valueExpr)
	case "uint16":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\"))\n", targetExpr, valueExpr)
	case "uint32":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\"))\n", targetExpr, valueExpr)
	case "uint64":
		fmt.Fprintf(buf, "\t\t%s = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\"))\n", targetExpr, valueExpr)
	case "struct":
		fmt.Fprintf(buf, "\t\t%s = __able_any_to_value(%s)\n", targetExpr, valueExpr)
	case "interface":
		info := g.nativeInterfaceInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = converted\n", targetExpr)
	case "union":
		info := g.nativeUnionInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t%s = converted\n", targetExpr)
	case "any":
		fmt.Fprintf(buf, "\t\t%s = __able_any_to_value(%s)\n", targetExpr, valueExpr)
	}
}

func (g *generator) renderValueToRuntimeNamed(buf *bytes.Buffer, valueExpr, goType, fieldName string) {
	if goType == "runtime.ErrorValue" {
		fmt.Fprintf(buf, "\tfields[%q] = %s\n", fieldName, valueExpr)
		return
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", spec.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tfields[%q] = converted\n", fieldName)
		fmt.Fprintf(buf, "\t}\n")
		return
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\tfields[%q] = %s(%s)\n", fieldName, helper, valueExpr)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", callable.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tfields[%q] = converted\n", fieldName)
		fmt.Fprintf(buf, "\t}\n")
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\tfields[%q] = %s\n", fieldName, valueExpr)
	case "bool":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToBool(%s)\n", fieldName, valueExpr)
	case "string":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToString(%s)\n", fieldName, valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToRune(%s)\n", fieldName, valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToFloat32(%s)\n", fieldName, valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToFloat64(%s)\n", fieldName, valueExpr)
	case "int":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))\n", fieldName, valueExpr)
	case "uint":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))\n", fieldName, valueExpr)
	case "int8":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\"))\n", fieldName, valueExpr)
	case "int16":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\"))\n", fieldName, valueExpr)
	case "int32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\"))\n", fieldName, valueExpr)
	case "int64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\"))\n", fieldName, valueExpr)
	case "uint8":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\"))\n", fieldName, valueExpr)
	case "uint16":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\"))\n", fieldName, valueExpr)
	case "uint32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\"))\n", fieldName, valueExpr)
	case "uint64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\"))\n", fieldName, valueExpr)
	case "struct":
		fmt.Fprintf(buf, "\tfields[%q] = __able_any_to_value(%s)\n", fieldName, valueExpr)
	case "interface":
		info := g.nativeInterfaceInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tfields[%q] = converted\n", fieldName)
		fmt.Fprintf(buf, "\t}\n")
	case "union":
		info := g.nativeUnionInfoForGoType(goType)
		if info == nil {
			return
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tconverted, err := %s(rt, %s)\n", info.ToRuntimeHelper, valueExpr)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tfields[%q] = converted\n", fieldName)
		fmt.Fprintf(buf, "\t}\n")
	case "any":
		fmt.Fprintf(buf, "\tfields[%q] = __able_any_to_value(%s)\n", fieldName, valueExpr)
	}
}

func (g *generator) renderConvertErr(buf *bytes.Buffer) {
	g.renderConvertErrWith(buf, "\t", "nil")
}

func (g *generator) renderConvertErrWith(buf *bytes.Buffer, indent string, returnExpr string) {
	fmt.Fprintf(buf, "%sif err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn %s, err\n", indent, returnExpr)
	fmt.Fprintf(buf, "%s}\n", indent)
}
