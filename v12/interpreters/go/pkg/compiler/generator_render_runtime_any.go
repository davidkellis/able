package compiler

import (
	"bytes"
	"fmt"
	"sort"
)

func (g *generator) renderRuntimeAnyHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_any_to_value(v any) runtime.Value {\n")
	fmt.Fprintf(buf, "\treturn __able_any_to_value_seen(v, nil)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_any_to_value_seen(v any, seen map[any]runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif v == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif rv, ok := v.(runtime.Value); ok {\n")
	fmt.Fprintf(buf, "\t\treturn rv\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch val := v.(type) {\n")
	fmt.Fprintf(buf, "\tcase bool:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToBool(val)\n")
	fmt.Fprintf(buf, "\tcase string:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToString(val)\n")
	fmt.Fprintf(buf, "\tcase int:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(val), runtime.IntegerType(\"isize\"))\n")
	fmt.Fprintf(buf, "\tcase int8:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(val), runtime.IntegerType(\"i8\"))\n")
	fmt.Fprintf(buf, "\tcase int16:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(val), runtime.IntegerType(\"i16\"))\n")
	fmt.Fprintf(buf, "\tcase int32:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(val), runtime.IntegerType(\"i32\"))\n")
	fmt.Fprintf(buf, "\tcase int64:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(val), runtime.IntegerType(\"i64\"))\n")
	fmt.Fprintf(buf, "\tcase uint:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(val), runtime.IntegerType(\"usize\"))\n")
	fmt.Fprintf(buf, "\tcase uint8:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(val), runtime.IntegerType(\"u8\"))\n")
	fmt.Fprintf(buf, "\tcase uint16:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(val), runtime.IntegerType(\"u16\"))\n")
	fmt.Fprintf(buf, "\tcase uint32:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(val), runtime.IntegerType(\"u32\"))\n")
	fmt.Fprintf(buf, "\tcase uint64:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(val), runtime.IntegerType(\"u64\"))\n")
	fmt.Fprintf(buf, "\tcase float32:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToFloat32(val)\n")
	fmt.Fprintf(buf, "\tcase float64:\n")
	fmt.Fprintf(buf, "\t\treturn bridge.ToFloat64(val)\n")
	fmt.Fprintf(buf, "\tcase struct{}:\n")
	fmt.Fprintf(buf, "\t\treturn runtime.VoidValue{}\n")
	for _, spec := range nativeNullableSpecs {
		// Go aliases rune to int32, so both cannot coexist as distinct
		// pointer cases inside one type switch.
		if spec.PtrType == "*rune" {
			continue
		}
		fmt.Fprintf(buf, "\tcase %s:\n", spec.PtrType)
		fmt.Fprintf(buf, "\t\treturn __able_nullable_%s_to_value(val)\n", spec.HelperStem)
	}
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil || info.GoType == "" {
			continue
		}
		fmt.Fprintf(buf, "\tcase %s:\n", info.GoType)
		fmt.Fprintf(buf, "\t\tif val == nil { return runtime.NilValue{} }\n")
		fmt.Fprintf(buf, "\t\trv, err := %s(__able_runtime, val)\n", info.ToRuntimeHelper)
		fmt.Fprintf(buf, "\t\tif err != nil { panic(err) }\n")
		fmt.Fprintf(buf, "\t\treturn rv\n")
	}
	for _, key := range g.sortedNativeCallableKeys() {
		info := g.nativeCallables[key]
		if info == nil || info.GoType == "" {
			continue
		}
		fmt.Fprintf(buf, "\tcase %s:\n", info.GoType)
		fmt.Fprintf(buf, "\t\tif val == nil { return runtime.NilValue{} }\n")
		fmt.Fprintf(buf, "\t\trv, err := %s(__able_runtime, val)\n", info.ToRuntimeHelper)
		fmt.Fprintf(buf, "\t\tif err != nil { panic(err) }\n")
		fmt.Fprintf(buf, "\t\treturn rv\n")
	}
	for _, spec := range g.sortedMonoArraySpecs() {
		fmt.Fprintf(buf, "\tcase *%s:\n", spec.GoName)
		fmt.Fprintf(buf, "\t\tif val == nil { return runtime.NilValue{} }\n")
		fmt.Fprintf(buf, "\t\trv, err := %s(__able_runtime, val)\n", spec.ToRuntimeHelper)
		fmt.Fprintf(buf, "\t\tif err != nil { panic(err) }\n")
		fmt.Fprintf(buf, "\t\treturn rv\n")
	}
	for _, key := range g.sortedNativeUnionKeys() {
		info := g.nativeUnions[key]
		if info == nil || info.GoType == "" {
			continue
		}
		fmt.Fprintf(buf, "\tcase %s:\n", info.GoType)
		fmt.Fprintf(buf, "\t\trv, err := %s(__able_runtime, val)\n", info.ToRuntimeHelper)
		fmt.Fprintf(buf, "\t\tif err != nil { panic(err) }\n")
		fmt.Fprintf(buf, "\t\treturn rv\n")
	}
	structNames := make([]string, 0, len(g.structs)+len(g.specializedStructs))
	for _, info := range g.allStructInfos() {
		if info != nil && info.Supported && info.GoName != "" {
			structNames = append(structNames, info.GoName)
		}
	}
	sort.Strings(structNames)
	for _, goName := range structNames {
		fmt.Fprintf(buf, "\tcase *%s:\n", goName)
		fmt.Fprintf(buf, "\t\tif val == nil { return runtime.NilValue{} }\n")
		fmt.Fprintf(buf, "\t\trv, err := __able_struct_%s_to_seen(__able_runtime, val, seen)\n", goName)
		fmt.Fprintf(buf, "\t\tif err != nil { panic(err) }\n")
		fmt.Fprintf(buf, "\t\treturn rv\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"__able_any_to_value: unsupported type %%T\", v))\n")
	fmt.Fprintf(buf, "}\n\n")
}
