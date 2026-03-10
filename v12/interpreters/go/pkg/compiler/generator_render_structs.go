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
		fmt.Fprintf(buf, "\t\tstate, handle, err := runtime.ArrayStoreEnsure(raw, len(raw.Elements))\n")
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn out, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tout.Length = int32(len(state.Values))\n")
		fmt.Fprintf(buf, "\t\tout.Capacity = int32(state.Capacity)\n")
		fmt.Fprintf(buf, "\t\tout.Storage_handle = bridge.ToInt(handle, runtime.IntegerI64)\n")
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
		fmt.Fprintf(buf, "\tif raw, ok, nilPtr := __able_runtime_array_value(targetCurrent); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\thandle, err := __able_array_handle_from_value(value.Storage_handle)\n")
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif _, err := runtime.ArrayStoreEnsureHandle(handle, int(value.Length), int(value.Capacity)); err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif err := runtime.ArrayStoreSetLength(handle, int(value.Length)); err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif err := runtime.ArrayStoreReserve(handle, int(value.Capacity)); err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tstate, err := runtime.ArrayStoreEnsureHandle(handle, int(value.Length), int(value.Capacity))\n")
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\traw.Handle = handle\n")
		fmt.Fprintf(buf, "\t\traw.Elements = state.Values\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tinst, ok := targetCurrent.(*runtime.StructInstanceValue)\n")
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

func (g *generator) renderValueConversion(buf *bytes.Buffer, indent, valueVar, goType, assignTarget, returnExpr string) {
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
	case "any":
		fmt.Fprintf(buf, "%s%s = %s\n", indent, assignTarget, valueVar)
	default:
		fmt.Fprintf(buf, "%sreturn %s, fmt.Errorf(\"unsupported field type\")\n", indent, returnExpr)
	}
}

func (g *generator) renderValueToRuntime(buf *bytes.Buffer, valueExpr, goType, targetSlice string) {
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
	case "any":
		fmt.Fprintf(buf, "\t%s = append(%s, __able_any_to_value(%s))\n", targetSlice, targetSlice, valueExpr)
	}
}

func (g *generator) renderValueToRuntimeNamed(buf *bytes.Buffer, valueExpr, goType, fieldName string) {
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
