package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderMonoArrayTypes(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	specs := g.sortedMonoArraySpecs()
	if len(specs) == 0 {
		return
	}
	for _, spec := range specs {
		fmt.Fprintf(buf, "type %s struct {\n", spec.GoName)
		fmt.Fprintf(buf, "\tElements       []%s\n", spec.ElemGoType)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMonoArrayConverters(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	specs := g.sortedMonoArraySpecs()
	if len(specs) == 0 {
		return
	}
	for _, spec := range specs {
		g.renderMonoArrayFrom(buf, spec)
		g.renderMonoArrayTo(buf, spec)
	}
}

func (g *generator) renderMonoArrayFrom(buf *bytes.Buffer, spec monoArraySpec) {
	fmt.Fprintf(buf, "func %s(value runtime.Value) (*%s, error) {\n", spec.FromRuntimeHelper, spec.GoName)
	fmt.Fprintf(buf, "\tcurrent := __able_unwrap_interface(value)\n")
	fmt.Fprintf(buf, "\tif _, isNil := current.(runtime.NilValue); isNil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvar sourceValues []runtime.Value\n")
	fmt.Fprintf(buf, "\tvar err error\n")
	fmt.Fprintf(buf, "\tif raw, ok, nilPtr := __able_runtime_array_value(current); ok || nilPtr {\n")
	fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"expected Array value\")\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif raw.Handle != 0 {\n")
	fmt.Fprintf(buf, "\t\t\tstate, err := runtime.ArrayStoreState(raw.Handle)\n")
	fmt.Fprintf(buf, "\t\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tsourceValues = make([]runtime.Value, len(state.Values), state.Capacity)\n")
	fmt.Fprintf(buf, "\t\t\tcopy(sourceValues, state.Values)\n")
	fmt.Fprintf(buf, "\t\t} else {\n")
	fmt.Fprintf(buf, "\t\t\tsourceValues = make([]runtime.Value, len(raw.Elements), cap(raw.Elements))\n")
	fmt.Fprintf(buf, "\t\t\tcopy(sourceValues, raw.Elements)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t} else if inst, ok := current.(*runtime.StructInstanceValue); ok && inst != nil {\n")
	fmt.Fprintf(buf, "\t\tsourceValues, _, _, _, err = __able_array_struct_instance_state(inst)\n")
	fmt.Fprintf(buf, "\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t} else {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"expected Array value\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tresult := &%s{Elements: make([]%s, len(sourceValues), cap(sourceValues))}\n", spec.GoName, spec.ElemGoType)
	fmt.Fprintf(buf, "\tfor idx, raw := range sourceValues {\n")
	g.renderMonoArrayElemFrom(buf, spec)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn result, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderMonoArrayTo(buf *bytes.Buffer, spec monoArraySpec) {
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value *%s) (runtime.Value, error) {\n", spec.ToRuntimeHelper, spec.GoName)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\telems := make([]runtime.Value, len(value.Elements), cap(value.Elements))\n")
	fmt.Fprintf(buf, "\tfor idx, raw := range value.Elements {\n")
	g.renderMonoArrayElemTo(buf, spec)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn &runtime.ArrayValue{Elements: elems}, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderMonoArrayElemFrom(buf *bytes.Buffer, spec monoArraySpec) {
	if buf == nil {
		return
	}
	g.renderValueConversion(buf, "\t\t", "raw", spec.ElemGoType, "result.Elements[idx]", "nil")
}

func (g *generator) renderMonoArrayElemTo(buf *bytes.Buffer, spec monoArraySpec) {
	if buf == nil {
		return
	}
	g.renderValueToRuntimeAssign(buf, "raw", spec.ElemGoType, "elems[idx]")
}
