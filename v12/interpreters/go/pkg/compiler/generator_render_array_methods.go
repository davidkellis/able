package compiler

import (
	"bytes"
	"fmt"
)

func isNativeArrayCoreMethod(method *methodInfo) bool {
	if method == nil || method.TargetName != "Array" {
		return false
	}
	switch method.MethodName {
	case "new", "with_capacity", "size", "len", "capacity", "is_empty", "push", "pop", "clear", "read_slot", "write_slot", "reserve", "clone_shallow", "refresh_metadata":
		return true
	default:
		return false
	}
}

func (g *generator) renderNativeArrayCoreMethod(buf *bytes.Buffer, method *methodInfo, info *functionInfo) {
	fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	arrayType := info.ReturnType
	if method.ExpectsSelf && len(info.Params) > 0 {
		arrayType = info.Params[0].GoType
	}
	arraySpec, monoArray := g.monoArraySpecForGoType(arrayType)
	syncCall := g.staticArraySyncCall(arrayType, "arr")
	selfSyncCall := ""
	if method.ExpectsSelf && len(info.Params) > 0 {
		selfSyncCall = g.staticArraySyncCall(arrayType, info.Params[0].GoName)
	}
	switch method.MethodName {
	case "new":
		if monoArray {
			fmt.Fprintf(buf, "\tarr := &%s{}\n", arraySpec.GoName)
		} else {
			fmt.Fprintf(buf, "\tarr := &Array{Length: 0, Capacity: 0, Storage_handle: int64(0)}\n")
		}
		fmt.Fprintf(buf, "\t%s\n", syncCall)
		fmt.Fprintf(buf, "\treturn arr, nil\n")
	case "with_capacity":
		capacity := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s < 0 {\n", capacity)
		fmt.Fprintf(buf, "\t\treturn nil, __able_control_from_error(fmt.Errorf(\"capacity must be a non-negative integer\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		if monoArray {
			fmt.Fprintf(buf, "\tarr := &%s{Elements: make([]%s, 0, int(%s))}\n", arraySpec.GoName, arraySpec.ElemGoType, capacity)
		} else {
			fmt.Fprintf(buf, "\tarr := &Array{Elements: make([]runtime.Value, 0, int(%s))}\n", capacity)
		}
		fmt.Fprintf(buf, "\t%s\n", syncCall)
		fmt.Fprintf(buf, "\treturn arr, nil\n")
	case "size":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn uint64(0), __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tcount := int32(len(%s.Elements))\n", self)
		fmt.Fprintf(buf, "\tif count <= 0 {\n")
		fmt.Fprintf(buf, "\t\treturn uint64(0), nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn uint64(count), nil\n")
	case "len":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn int32(0), __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn int32(len(%s.Elements)), nil\n", self)
	case "capacity":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn int32(0), __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn int32(cap(%s.Elements)), nil\n", self)
	case "is_empty":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn false, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn len(%s.Elements) == 0, nil\n", self)
	case "push":
		self := info.Params[0].GoName
		value := info.Params[1].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s.Elements = append(%s.Elements, %s)\n", self, self, value)
		fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	case "pop":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		if monoArray && info.ReturnType != "runtime.Value" {
			fmt.Fprintf(buf, "\t\treturn nil, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		} else {
			fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		}
		fmt.Fprintf(buf, "\t}\n")
		nullableInner, nullableReturn := g.nativeNullableValueInnerType(info.ReturnType)
		directNilableCarrier := monoArray && arraySpec != nil && info.ReturnType == arraySpec.ElemGoType && g.isNilableStaticCarrierType(info.ReturnType)
		switch {
		case monoArray && info.ReturnType == "runtime.Value":
			fmt.Fprintf(buf, "\tvar result runtime.Value = runtime.NilValue{}\n")
			fmt.Fprintf(buf, "\tif count := len(%s.Elements); count > 0 {\n", self)
			fmt.Fprintf(buf, "\t\tvalue := %s.Elements[count-1]\n", self)
			fmt.Fprintf(buf, "\t\tresult = %s\n", g.staticArrayResultValueExpr(arrayType, "value"))
			fmt.Fprintf(buf, "\t\t%s.Elements = %s.Elements[:count-1]\n", self, self)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		case directNilableCarrier:
			fmt.Fprintf(buf, "\tvar result %s\n", info.ReturnType)
			fmt.Fprintf(buf, "\tif count := len(%s.Elements); count > 0 {\n", self)
			fmt.Fprintf(buf, "\t\tresult = %s.Elements[count-1]\n", self)
			fmt.Fprintf(buf, "\t\t%s.Elements = %s.Elements[:count-1]\n", self, self)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		case monoArray && nullableReturn && nullableInner == arraySpec.ElemGoType:
			fmt.Fprintf(buf, "\tvar result %s\n", info.ReturnType)
			fmt.Fprintf(buf, "\tif count := len(%s.Elements); count > 0 {\n", self)
			fmt.Fprintf(buf, "\t\tvalue := %s.Elements[count-1]\n", self)
			fmt.Fprintf(buf, "\t\tresult = __able_ptr(value)\n")
			fmt.Fprintf(buf, "\t\t%s.Elements = %s.Elements[:count-1]\n", self, self)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		default:
			fmt.Fprintf(buf, "\tvar result runtime.Value = runtime.NilValue{}\n")
			fmt.Fprintf(buf, "\tif count := len(%s.Elements); count > 0 {\n", self)
			fmt.Fprintf(buf, "\t\tresult = %s.Elements[count-1]\n", self)
			fmt.Fprintf(buf, "\t\t%s.Elements = %s.Elements[:count-1]\n", self, self)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
			fmt.Fprintf(buf, "\tif result == nil {\n")
			fmt.Fprintf(buf, "\t\tresult = runtime.NilValue{}\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\treturn result, nil\n")
	case "clear":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s.Elements = %s.Elements[:0]\n", self, self)
		fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	case "read_slot":
		self := info.Params[0].GoName
		idx := info.Params[1].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		if monoArray && info.ReturnType != "runtime.Value" {
			fmt.Fprintf(buf, "\t\treturn nil, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		} else {
			fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		}
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\ti := int(%s)\n", idx)
		nullableInner, nullableReturn := g.nativeNullableValueInnerType(info.ReturnType)
		directNilableCarrier := monoArray && arraySpec != nil && info.ReturnType == arraySpec.ElemGoType && g.isNilableStaticCarrierType(info.ReturnType)
		switch {
		case monoArray && info.ReturnType == "runtime.Value":
			fmt.Fprintf(buf, "\tif i >= 0 && i < len(%s.Elements) {\n", self)
			fmt.Fprintf(buf, "\t\tvalue := %s.Elements[i]\n", self)
			fmt.Fprintf(buf, "\t\treturn %s, nil\n", g.staticArrayResultValueExpr(arrayType, "value"))
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn runtime.NilValue{}, nil\n")
		case directNilableCarrier:
			fmt.Fprintf(buf, "\tif i >= 0 && i < len(%s.Elements) {\n", self)
			fmt.Fprintf(buf, "\t\treturn %s.Elements[i], nil\n", self)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn nil, nil\n")
		case monoArray && nullableReturn && nullableInner == arraySpec.ElemGoType:
			fmt.Fprintf(buf, "\tif i >= 0 && i < len(%s.Elements) {\n", self)
			fmt.Fprintf(buf, "\t\tvalue := %s.Elements[i]\n", self)
			fmt.Fprintf(buf, "\t\treturn __able_ptr(value), nil\n")
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn nil, nil\n")
		default:
			fmt.Fprintf(buf, "\tif i >= 0 && i < len(%s.Elements) {\n", self)
			fmt.Fprintf(buf, "\t\tif v := %s.Elements[i]; v != nil {\n", self)
			fmt.Fprintf(buf, "\t\t\treturn v, nil\n")
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn runtime.NilValue{}, nil\n")
		}
	case "write_slot":
		self := info.Params[0].GoName
		idx := info.Params[1].GoName
		value := info.Params[2].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\ti := int(%s)\n", idx)
		fmt.Fprintf(buf, "\tif i >= 0 && i < len(%s.Elements) {\n", self)
		fmt.Fprintf(buf, "\t\t%s.Elements[i] = %s\n", self, value)
		fmt.Fprintf(buf, "\t} else if i >= 0 {\n")
		if monoArray {
			fmt.Fprintf(buf, "\t\tfor len(%s.Elements) <= i { %s.Elements = append(%s.Elements, %s) }\n", self, self, self, g.staticArrayZeroValueExpr(arrayType))
		} else {
			fmt.Fprintf(buf, "\t\tfor len(%s.Elements) <= i { %s.Elements = append(%s.Elements, runtime.NilValue{}) }\n", self, self, self)
		}
		fmt.Fprintf(buf, "\t\t%s.Elements[i] = %s\n", self, value)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	case "reserve":
		self := info.Params[0].GoName
		capacity := info.Params[1].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif %s < 0 {\n", capacity)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"capacity must be a non-negative integer\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif int(%s) > cap(%s.Elements) {\n", capacity, self)
		if monoArray {
			fmt.Fprintf(buf, "\t\telems := make([]%s, len(%s.Elements), int(%s))\n", arraySpec.ElemGoType, self, capacity)
		} else {
			fmt.Fprintf(buf, "\t\telems := make([]runtime.Value, len(%s.Elements), int(%s))\n", self, capacity)
		}
		fmt.Fprintf(buf, "\t\tcopy(elems, %s.Elements)\n", self)
		fmt.Fprintf(buf, "\t\t%s.Elements = elems\n", self)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	case "clone_shallow":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn nil, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		if monoArray {
			fmt.Fprintf(buf, "\telems := make([]%s, len(%s.Elements), cap(%s.Elements))\n", arraySpec.ElemGoType, self, self)
			fmt.Fprintf(buf, "\tcopy(elems, %s.Elements)\n", self)
			fmt.Fprintf(buf, "\tcloned := &%s{Elements: elems}\n", arraySpec.GoName)
			fmt.Fprintf(buf, "\t%s(cloned)\n", arraySpec.SyncHelper)
		} else {
			fmt.Fprintf(buf, "\tcloned := &Array{Elements: __able_struct_Array_clone_elements(%s.Elements, __able_struct_Array_capacity_hint(%s))}\n", self, self)
			fmt.Fprintf(buf, "\t__able_struct_Array_sync(cloned)\n")
		}
		fmt.Fprintf(buf, "\treturn cloned, nil\n")
	case "refresh_metadata":
		self := info.Params[0].GoName
		fmt.Fprintf(buf, "\tif %s == nil {\n", self)
		fmt.Fprintf(buf, "\t\treturn struct{}{}, __able_control_from_error(fmt.Errorf(\"missing Array value\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s\n", selfSyncCall)
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	default:
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"compiler: unsupported native Array method %s\"))\n", method.MethodName)
	}
	fmt.Fprintf(buf, "}\n\n")
}
