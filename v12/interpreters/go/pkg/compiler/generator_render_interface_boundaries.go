package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

func (g *generator) renderNativeInterfaceDirectCoercionControl(buf *bytes.Buffer, target string, expr string, actualGoType string, expectedGoType string, failureGoType string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualGoType == "" || expectedGoType == "" {
		return false
	}
	if actualGoType == expectedGoType {
		fmt.Fprintf(buf, "\t%s := %s\n", target, expr)
		return true
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actualGoType)
	expectedInfo := g.nativeInterfaceInfoForGoType(expectedGoType)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	zeroExpr, ok := g.zeroValueExpr(failureGoType)
	if !ok {
		return false
	}
	fmt.Fprintf(buf, "\tvar %s %s\n", target, expectedGoType)
	defaultLine := fmt.Sprintf("\t\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion from %s to %s\"))", zeroExpr, actualGoType, expectedGoType)
	return g.renderNativeInterfaceDirectCoercionSwitch(buf, "\t", target, expr, actualInfo, expectedInfo, defaultLine)
}

func (g *generator) renderNativeInterfaceDirectCoercionError(buf *bytes.Buffer, indent string, target string, expr string, actualGoType string, expectedGoType string, failureExpr string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualGoType == "" || expectedGoType == "" || failureExpr == "" {
		return false
	}
	if actualGoType == expectedGoType {
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actualGoType)
	expectedInfo := g.nativeInterfaceInfoForGoType(expectedGoType)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	fmt.Fprintf(buf, "%svar %s %s\n", indent, target, expectedGoType)
	fmt.Fprintf(buf, "%sif %s == nil {\n", indent, expr)
	fmt.Fprintf(buf, "%s\t%s = %s(nil)\n", indent, target, expectedGoType)
	fmt.Fprintf(buf, "%s} else {\n", indent)
	defaultLine := fmt.Sprintf("%s\t\treturn %s, fmt.Errorf(\"unsupported native interface conversion from %s to %s\")", indent, failureExpr, actualGoType, expectedGoType)
	if !g.renderNativeInterfaceDirectCoercionSwitch(buf, indent+"\t", target, expr, actualInfo, expectedInfo, defaultLine) {
		return false
	}
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceDirectCoercionSwitch(buf *bytes.Buffer, indent string, target string, expr string, actualInfo *nativeInterfaceInfo, expectedInfo *nativeInterfaceInfo, defaultLine string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualInfo == nil || expectedInfo == nil || defaultLine == "" {
		return false
	}
	fmt.Fprintf(buf, "%sswitch typed := %s.(type) {\n", indent, expr)
	for _, actualAdapter := range g.nativeInterfaceKnownAdapters(actualInfo) {
		if actualAdapter == nil || actualAdapter.AdapterType == "" || actualAdapter.GoType == "" {
			continue
		}
		expectedAdapter, ok := g.nativeInterfaceAdapterForActual(expectedInfo, actualAdapter.GoType)
		if !ok || expectedAdapter == nil || expectedAdapter.WrapHelper == "" {
			continue
		}
		fmt.Fprintf(buf, "%scase %s:\n", indent, actualAdapter.AdapterType)
		fmt.Fprintf(buf, "%s\t%s = %s(typed.Value)\n", indent, target, expectedAdapter.WrapHelper)
	}
	if actualInfo.RuntimeAdapter != "" && expectedInfo.RuntimeWrapHelper != "" {
		fmt.Fprintf(buf, "%scase %s:\n", indent, actualInfo.RuntimeAdapter)
		fmt.Fprintf(buf, "%s\t%s = %s(typed.Value)\n", indent, target, expectedInfo.RuntimeWrapHelper)
	}
	fmt.Fprintf(buf, "%sdefault:\n", indent)
	fmt.Fprintf(buf, "%s\n", defaultLine)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceBoundaryHelpers(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	renderedType, ok := g.renderTypeExpression(info.TypeExpr)
	if !ok {
		return
	}
	baseName, _ := typeExprBaseName(info.TypeExpr)
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value runtime.Value) (%s, bool, error) {\n", info.TryFromRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbase := __able_unwrap_interface(value)\n")
	fmt.Fprintf(buf, "\t_ = base\n")
	renderedAdapterTypes := make(map[string]struct{})
	if baseName == "Iterator" {
		fmt.Fprintf(buf, "\tif iter, ok, nilPtr := __able_runtime_iterator_value(value); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, false, nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn %s(iter), true, nil\n", info.RuntimeWrapHelper)
		fmt.Fprintf(buf, "\t}\n")
	}
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.TypeExpr == nil {
			continue
		}
		if adapter.GoType != "" {
			renderedAdapterTypes[adapter.GoType] = struct{}{}
		}
		renderedAdapterType, ok := g.renderTypeExpression(adapter.TypeExpr)
		if !ok {
			continue
		}
		fmt.Fprintf(buf, "\tif coerced, ok, err := bridge.MatchType(rt, %s, base); err != nil {\n", renderedAdapterType)
		fmt.Fprintf(buf, "\t\treturn nil, false, err\n")
		fmt.Fprintf(buf, "\t} else if ok {\n")
		if g.renderNativeInterfaceRuntimeToGoValueTryError(buf, "converted", "coerced", adapter.GoType, "\t\t") {
			fmt.Fprintf(buf, "\t\treturn %s(converted), true, nil\n", adapter.WrapHelper)
		} else {
			fmt.Fprintf(buf, "\t\t_ = coerced\n")
		}
		fmt.Fprintf(buf, "\t}\n")
	}
	for _, actualKey := range g.sortedNativeInterfaceKeys() {
		actualInfo := g.nativeInterfaces[actualKey]
		if actualInfo == nil || actualInfo.Key == info.Key {
			continue
		}
		if !g.nativeInterfaceDirectAdapterPossible(actualInfo, info) {
			continue
		}
		for _, actualAdapter := range g.nativeInterfaceKnownAdapters(actualInfo) {
			if actualAdapter == nil || actualAdapter.TypeExpr == nil || actualAdapter.GoType == "" {
				continue
			}
			if _, seen := renderedAdapterTypes[actualAdapter.GoType]; seen {
				continue
			}
			actualInterfaceAdapter, ok := g.nativeInterfaceAdapterForActual(info, actualInfo.GoType)
			if !ok || actualInterfaceAdapter == nil || actualInterfaceAdapter.WrapHelper == "" {
				continue
			}
			renderedActualType, ok := g.renderTypeExpression(actualAdapter.TypeExpr)
			if !ok {
				continue
			}
			renderedAdapterTypes[actualAdapter.GoType] = struct{}{}
			fmt.Fprintf(buf, "\tif coerced, ok, err := bridge.MatchType(rt, %s, base); err != nil {\n", renderedActualType)
			fmt.Fprintf(buf, "\t\treturn nil, false, err\n")
			fmt.Fprintf(buf, "\t} else if ok {\n")
			if g.renderNativeInterfaceRuntimeToGoValueTryError(buf, "converted", "coerced", actualAdapter.GoType, "\t\t") {
				fmt.Fprintf(buf, "\t\treturn %s(%s(converted)), true, nil\n", actualInterfaceAdapter.WrapHelper, actualAdapter.WrapHelper)
			} else {
				fmt.Fprintf(buf, "\t\t_ = coerced\n")
			}
			fmt.Fprintf(buf, "\t}\n")
		}
	}
	fmt.Fprintf(buf, "\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn %s(coerced), true, nil\n", info.RuntimeWrapHelper)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value runtime.Value) (%s, error) {\n", info.FromRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tconverted, ok, err := %s(rt, value)\n", info.TryFromRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch: expected %s\")\n", info.TypeString)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value runtime.Value) %s {\n", info.FromRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.FromRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value %s) (runtime.Value, error) {\n", info.ToRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value.%s(rt)\n", info.ToRuntimeMethod)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value %s) runtime.Value {\n", info.ToRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.ToRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceApplyRuntimeHelper(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	if info == nil {
		return
	}
	fmt.Fprintf(buf, "func %s(value %s, runtimeValue runtime.Value) error {\n", info.ApplyRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch typed := value.(type) {\n")
	renderedCase := false
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || !strings.HasPrefix(adapter.GoType, "*") || g.typeCategory(adapter.GoType) != "struct" {
			continue
		}
		baseName, ok := g.structHelperName(adapter.GoType)
		if !ok {
			baseName = strings.TrimPrefix(adapter.GoType, "*")
		}
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", adapter.AdapterType)
		fmt.Fprintf(buf, "\t\tif typed.Value == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		if g.isArrayStructType(adapter.GoType) {
			fmt.Fprintf(buf, "\t\tvar converted *Array\n")
			fmt.Fprintf(buf, "\t\tvar err error\n")
			for _, line := range g.runtimeValueToGenericArrayBoundaryLines("converted", "err", "runtimeValue", true) {
				fmt.Fprintf(buf, "\t\t%s\n", line)
			}
		} else {
			fmt.Fprintf(buf, "\t\tconverted, err := __able_struct_%s_from(runtimeValue)\n", baseName)
		}
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif converted == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t*typed.Value = *converted\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
	}
	fmt.Fprintf(buf, "\tdefault:\n")
	if renderedCase {
		fmt.Fprintf(buf, "\t\treturn nil\n")
	} else {
		fmt.Fprintf(buf, "\t\t_ = typed\n")
		fmt.Fprintf(buf, "\t\t_ = runtimeValue\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceRuntimeToGoValueTryError(buf *bytes.Buffer, target string, expr string, goType string, indent string) bool {
	rawVar := sanitizeIdent(target) + "_raw"
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, spec.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, iface.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, callable.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, union.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, helper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "%s%s := bridge.ErrorValue(__able_runtime, %s)\n", indent, target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "%svar %s any = %s\n", indent, target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "%s_ = %s\n", indent, expr)
		fmt.Fprintf(buf, "%s%s := struct{}{}\n", indent, target)
		return true
	case "bool":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsBool(%s)\n", indent, target, expr)
	case "string":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsString(%s)\n", indent, target, expr)
	case "rune":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsRune(%s)\n", indent, target, expr)
	case "float32":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := float32(%s)\n", indent, target, rawVar)
		return true
	case "float64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, target, expr)
	case "int":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := int(%s)\n", indent, target, rawVar)
		return true
	case "uint":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := uint(%s)\n", indent, target, rawVar)
		return true
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	default:
		if g.typeCategory(goType) == "struct" {
			if g.isArrayStructType(goType) {
				errVar := sanitizeIdent(target) + "_err"
				fmt.Fprintf(buf, "%svar %s *Array\n", indent, target)
				fmt.Fprintf(buf, "%svar %s error\n", indent, errVar)
				for _, line := range g.runtimeValueToGenericArrayBoundaryLines(target, errVar, expr, true) {
					fmt.Fprintf(buf, "%s%s\n", indent, line)
				}
				fmt.Fprintf(buf, "%sif %s != nil {\n", indent, errVar)
				fmt.Fprintf(buf, "%s\treturn nil, false, %s\n", indent, errVar)
				fmt.Fprintf(buf, "%s}\n", indent)
				return true
			}
			baseName, ok := g.structHelperName(goType)
			if !ok {
				baseName = strings.TrimPrefix(goType, "*")
			}
			fmt.Fprintf(buf, "%s%s, err := __able_struct_%s_from(%s)\n", indent, target, baseName, expr)
			fmt.Fprintf(buf, "%sif err != nil {\n", indent)
			fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
			fmt.Fprintf(buf, "%s}\n", indent)
			return true
		}
		fmt.Fprintf(buf, "%sreturn nil, false, fmt.Errorf(\"unsupported native interface conversion to %s\")\n", indent, goType)
		return false
	}
	fmt.Fprintf(buf, "%sif err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn nil, false, err\n", indent)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceRuntimeToGoValueError(buf *bytes.Buffer, target string, expr string, goType string, indent string) bool {
	rawVar := sanitizeIdent(target) + "_raw"
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, spec.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, iface.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, callable.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, union.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, helper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "%s%s := bridge.ErrorValue(__able_runtime, %s)\n", indent, target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "%svar %s any = %s\n", indent, target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "%s_ = %s\n", indent, expr)
		fmt.Fprintf(buf, "%s%s := struct{}{}\n", indent, target)
		return true
	case "bool":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsBool(%s)\n", indent, target, expr)
	case "string":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsString(%s)\n", indent, target, expr)
	case "rune":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsRune(%s)\n", indent, target, expr)
	case "float32":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := float32(%s)\n", indent, target, rawVar)
		return true
	case "float64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, target, expr)
	case "int":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := int(%s)\n", indent, target, rawVar)
		return true
	case "uint":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := uint(%s)\n", indent, target, rawVar)
		return true
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	default:
		if g.typeCategory(goType) == "struct" {
			if g.isArrayStructType(goType) {
				errVar := sanitizeIdent(target) + "_err"
				fmt.Fprintf(buf, "%svar %s *Array\n", indent, target)
				fmt.Fprintf(buf, "%svar %s error\n", indent, errVar)
				for _, line := range g.runtimeValueToGenericArrayBoundaryLines(target, errVar, expr, true) {
					fmt.Fprintf(buf, "%s%s\n", indent, line)
				}
				fmt.Fprintf(buf, "%sif %s != nil {\n", indent, errVar)
				fmt.Fprintf(buf, "%s\treturn nil, %s\n", indent, errVar)
				fmt.Fprintf(buf, "%s}\n", indent)
				return true
			}
			baseName, ok := g.structHelperName(goType)
			if !ok {
				baseName = strings.TrimPrefix(goType, "*")
			}
			fmt.Fprintf(buf, "%s%s, err := __able_struct_%s_from(%s)\n", indent, target, baseName, expr)
			fmt.Fprintf(buf, "%sif err != nil {\n", indent)
			fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
			fmt.Fprintf(buf, "%s}\n", indent)
			return true
		}
		fmt.Fprintf(buf, "%sreturn nil, fmt.Errorf(\"unsupported native interface conversion to %s\")\n", indent, goType)
		return false
	}
	fmt.Fprintf(buf, "%sif err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceGoToRuntimeValueControl(buf *bytes.Buffer, target string, expr string, goType string, returnType string) {
	fmt.Fprintf(buf, "\t")
	if g.renderNativeInterfaceGoToRuntimeValue(buf, target, expr, goType, "\t", false, returnType) {
		return
	}
}

func (g *generator) renderNativeInterfaceGoToRuntimeValueError(buf *bytes.Buffer, target string, expr string, goType string) {
	g.renderNativeInterfaceGoToRuntimeValue(buf, target, expr, goType, "\t", true, "")
}

func (g *generator) renderNativeInterfaceGoToRuntimeValue(buf *bytes.Buffer, target string, expr string, goType string, indent string, returnError bool, returnType string) bool {
	failErr := func(errExpr string) {
		if returnError {
			fmt.Fprintf(buf, "%sif err != nil {\n", indent)
			fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
			fmt.Fprintf(buf, "%s}\n", indent)
			return
		}
		zeroExpr, ok := g.zeroValueExpr(returnType)
		if !ok {
			fmt.Fprintf(buf, "%svar zero %s\n", indent, returnType)
			zeroExpr = "zero"
		}
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn %s, __able_control_from_error(%s)\n", indent, zeroExpr, errExpr)
		fmt.Fprintf(buf, "%s}\n", indent)
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, spec.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, iface.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, callable.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, union.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, helper, expr)
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "%s%s := __able_any_to_value(%s)\n", indent, target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "%s_ = %s\n", indent, expr)
		fmt.Fprintf(buf, "%s%s := runtime.VoidValue{}\n", indent, target)
		return true
	case "bool":
		fmt.Fprintf(buf, "%s%s := bridge.ToBool(%s)\n", indent, target, expr)
		return true
	case "string":
		fmt.Fprintf(buf, "%s%s := bridge.ToString(%s)\n", indent, target, expr)
		return true
	case "rune":
		fmt.Fprintf(buf, "%s%s := bridge.ToRune(%s)\n", indent, target, expr)
		return true
	case "float32":
		fmt.Fprintf(buf, "%s%s := bridge.ToFloat32(%s)\n", indent, target, expr)
		return true
	case "float64":
		fmt.Fprintf(buf, "%s%s := bridge.ToFloat64(%s)\n", indent, target, expr)
		return true
	case "int":
		fmt.Fprintf(buf, "%s%s := bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))\n", indent, target, expr)
		return true
	case "uint":
		fmt.Fprintf(buf, "%s%s := bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))\n", indent, target, expr)
		return true
	case "int8", "int16", "int32", "int64":
		suffix, _ := g.integerTypeSuffix(goType)
		fmt.Fprintf(buf, "%s%s := bridge.ToInt(int64(%s), runtime.IntegerType(%q))\n", indent, target, expr, suffix)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		suffix, _ := g.integerTypeSuffix(goType)
		fmt.Fprintf(buf, "%s%s := bridge.ToUint(uint64(%s), runtime.IntegerType(%q))\n", indent, target, expr, suffix)
		return true
	}
	if g.typeCategory(goType) == "struct" {
		baseName, ok := g.structHelperName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "%s%s, err := __able_struct_%s_to(__able_runtime, %s)\n", indent, target, baseName, expr)
		failErr("err")
		return true
	}
	fmt.Fprintf(buf, "%svar %s runtime.Value\n", indent, target)
	fmt.Fprintf(buf, "%s_ = %s\n", indent, target)
	if returnError {
		fmt.Fprintf(buf, "%sreturn nil, fmt.Errorf(\"unsupported native interface conversion from %s\")\n", indent, goType)
	} else {
		zeroExpr, ok := g.zeroValueExpr(returnType)
		if !ok {
			fmt.Fprintf(buf, "%svar zero %s\n", indent, returnType)
			zeroExpr = "zero"
		}
		fmt.Fprintf(buf, "%sreturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion from %s\"))\n", indent, zeroExpr, goType)
	}
	return false
}

func (g *generator) renderNativeInterfaceRuntimeToGoValueControl(buf *bytes.Buffer, target string, expr string, goType string, returnType string) bool {
	zeroExpr, zeroOK := g.zeroValueExpr(returnType)
	if !zeroOK {
		fmt.Fprintf(buf, "\tvar zero %s\n", returnType)
		zeroExpr = "zero"
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(%s)\n", target, spec.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, iface.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, callable.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, union.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\t%s, err := %s(%s)\n", target, helper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "\t%s := %s\n", target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "\t%s := bridge.ErrorValue(__able_runtime, %s)\n", target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "\tvar %s any = %s\n", target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "\tif errVal, ok, nilPtr := __able_runtime_error_value(%s); ok || nilPtr {\n", expr)
		fmt.Fprintf(buf, "\t\tif ok {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, __able_raise_control(nil, errVal)\n", zeroExpr)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t_ = %s\n", expr)
		fmt.Fprintf(buf, "\t%s := struct{}{}\n", target)
		return true
	case "bool":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsBool(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "string":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsString(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "rune":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsRune(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "float32":
		fmt.Fprintf(buf, "\traw, err := bridge.AsFloat(%s)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := float32(raw)\n", target)
		return true
	case "float64":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsFloat(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "int":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := int(raw)\n", target)
		return true
	case "uint":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := uint(raw)\n", target)
		return true
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(%s, %d)\n", expr, g.intBits(goType))
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := %s(raw)\n", target, goType)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(%s, %d)\n", expr, g.intBits(goType))
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := %s(raw)\n", target, goType)
		return true
	default:
		if g.typeCategory(goType) == "struct" {
			if g.isArrayStructType(goType) {
				errVar := sanitizeIdent(target) + "_err"
				fmt.Fprintf(buf, "\tvar %s *Array\n", target)
				fmt.Fprintf(buf, "\tvar %s error\n", errVar)
				for _, line := range g.runtimeValueToGenericArrayBoundaryLines(target, errVar, expr, true) {
					fmt.Fprintf(buf, "\t%s\n", line)
				}
				fmt.Fprintf(buf, "\tif %s != nil {\n", errVar)
				fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(%s)\n", zeroExpr, errVar)
				fmt.Fprintf(buf, "\t}\n")
				return true
			}
			baseName, ok := g.structHelperName(goType)
			if !ok {
				baseName = strings.TrimPrefix(goType, "*")
			}
			fmt.Fprintf(buf, "\t%s, err := __able_struct_%s_from(%s)\n", target, baseName, expr)
			fmt.Fprintf(buf, "\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			return true
		}
		fmt.Fprintf(buf, "\tvar %s %s\n", target, goType)
		fmt.Fprintf(buf, "\t_ = %s\n", target)
		fmt.Fprintf(buf, "\t_ = %s\n", expr)
		fmt.Fprintf(buf, "\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion to %s\"))\n", zeroExpr, goType)
		return false
	}
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	return true
}
