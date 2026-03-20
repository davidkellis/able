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
		fmt.Fprintf(buf, "\tLength         int32\n")
		fmt.Fprintf(buf, "\tCapacity       int32\n")
		fmt.Fprintf(buf, "\tStorage_handle int64\n")
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
		g.renderMonoArraySync(buf, spec)
		g.renderMonoArrayFrom(buf, spec)
		g.renderMonoArrayTo(buf, spec)
	}
}

func (g *generator) renderMonoArraySync(buf *bytes.Buffer, spec monoArraySpec) {
	fmt.Fprintf(buf, "func %s(value *%s) {\n", spec.SyncHelper, spec.GoName)
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvalue.Length = int32(len(value.Elements))\n")
	fmt.Fprintf(buf, "\tvalue.Capacity = int32(cap(value.Elements))\n")
	fmt.Fprintf(buf, "\tif value.Storage_handle == 0 && value.Capacity < value.Length {\n")
	fmt.Fprintf(buf, "\t\tvalue.Capacity = value.Length\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderMonoArrayFrom(buf *bytes.Buffer, spec monoArraySpec) {
	fmt.Fprintf(buf, "func %s(value runtime.Value) (*%s, error) {\n", spec.FromRuntimeHelper, spec.GoName)
	fmt.Fprintf(buf, "\tbase, err := __able_struct_Array_from(value)\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif base == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tresult := &%s{Length: base.Length, Capacity: base.Capacity, Storage_handle: base.Storage_handle, Elements: make([]%s, len(base.Elements), cap(base.Elements))}\n", spec.GoName, spec.ElemGoType)
	fmt.Fprintf(buf, "\tfor idx, raw := range base.Elements {\n")
	g.renderMonoArrayElemFrom(buf, spec)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t%s(result)\n", spec.SyncHelper)
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
	fmt.Fprintf(buf, "\t%s(value)\n", spec.SyncHelper)
	fmt.Fprintf(buf, "\telems := make([]runtime.Value, len(value.Elements), cap(value.Elements))\n")
	fmt.Fprintf(buf, "\tfor idx, raw := range value.Elements {\n")
	g.renderMonoArrayElemTo(buf, spec)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbase := &Array{Length: value.Length, Capacity: value.Capacity, Storage_handle: value.Storage_handle, Elements: elems}\n")
	fmt.Fprintf(buf, "\t__able_struct_Array_sync(base)\n")
	fmt.Fprintf(buf, "\treturn __able_struct_Array_to(rt, base)\n")
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
