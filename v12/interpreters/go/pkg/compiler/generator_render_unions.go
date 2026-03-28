package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderNativeUnions(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.nativeUnions) == 0 {
		return
	}
	if g.nativeUnionRendered == nil {
		g.nativeUnionRendered = make(map[string]struct{})
	}
	for _, key := range g.sortedNativeUnionKeys() {
		info := g.nativeUnions[key]
		if info == nil {
			continue
		}
		if _, ok := g.nativeUnionRendered[key]; ok {
			continue
		}
		fmt.Fprintf(buf, "type %s interface {\n", info.GoType)
		fmt.Fprintf(buf, "\t%s()\n", info.MarkerMethod)
		fmt.Fprintf(buf, "}\n\n")
		for _, member := range info.Members {
			fmt.Fprintf(buf, "type %s struct {\n", member.WrapperType)
			fmt.Fprintf(buf, "\tValue %s\n", member.GoType)
			fmt.Fprintf(buf, "}\n\n")
			fmt.Fprintf(buf, "func (%s) %s() {}\n\n", member.WrapperType, info.MarkerMethod)
			fmt.Fprintf(buf, "func %s(value %s) %s {\n", member.WrapHelper, member.GoType, info.GoType)
			fmt.Fprintf(buf, "\treturn %s{Value: value}\n", member.WrapperType)
			fmt.Fprintf(buf, "}\n\n")
			fmt.Fprintf(buf, "func %s(value %s) (%s, bool) {\n", member.UnwrapHelper, info.GoType, member.GoType)
			fmt.Fprintf(buf, "\traw, ok := value.(%s)\n", member.WrapperType)
			fmt.Fprintf(buf, "\tif !ok {\n")
			fmt.Fprintf(buf, "\t\tvar zero %s\n", member.GoType)
			fmt.Fprintf(buf, "\t\treturn zero, false\n")
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn raw.Value, true\n")
			fmt.Fprintf(buf, "}\n\n")
		}
		g.renderNativeUnionFromRuntimeHelper(buf, info)
		g.renderNativeUnionToRuntimeHelper(buf, info)
		g.nativeUnionRendered[key] = struct{}{}
	}
}

func (g *generator) renderNativeUnionFromRuntimeHelper(buf *bytes.Buffer, info *nativeUnionInfo) {
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value runtime.Value) (%s, error) {\n", info.FromRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	var fallbackRuntimeMember *nativeUnionMember
	for _, member := range info.Members {
		renderedType, ok := g.renderTypeExpression(member.TypeExpr)
		if !ok {
			continue
		}
		if member.GoType == "runtime.Value" {
			if g.nativeUnionRuntimeMemberRequiresMatch(member) {
				fmt.Fprintf(buf, "\t{\n")
				fmt.Fprintf(buf, "\t\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
				fmt.Fprintf(buf, "\t\tif err != nil {\n")
				fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
				fmt.Fprintf(buf, "\t\t}\n")
				fmt.Fprintf(buf, "\t\tif ok {\n")
				fmt.Fprintf(buf, "\t\t\treturn %s(coerced), nil\n", member.WrapHelper)
				fmt.Fprintf(buf, "\t\t}\n")
				fmt.Fprintf(buf, "\t}\n")
			} else if fallbackRuntimeMember == nil {
				fallbackRuntimeMember = member
			}
			continue
		}
		if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil {
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
			fmt.Fprintf(buf, "\t\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\tif ok {\n")
			fmt.Fprintf(buf, "\t\t\treturn %s(%s(coerced)), nil\n", member.WrapHelper, iface.RuntimeWrapHelper)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
			continue
		}
		if member.GoType == "runtime.ErrorValue" {
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
			fmt.Fprintf(buf, "\t\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\tif ok {\n")
			fmt.Fprintf(buf, "\t\t\treturn %s(bridge.ErrorValue(rt, coerced)), nil\n", member.WrapHelper)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
			continue
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif ok {\n")
		switch {
		case g.isMonoArrayType(member.GoType):
			spec, _ := g.monoArraySpecForGoType(member.GoType)
			fmt.Fprintf(buf, "\t\t\tconverted, err := %s(coerced)\n", spec.FromRuntimeHelper)
		case member.GoType == "struct{}":
			fmt.Fprintf(buf, "\t\t\tconverted := struct{}{}\n")
			fmt.Fprintf(buf, "\t\t\t_ = coerced\n")
		case g.typeCategory(member.GoType) == "struct":
			helperName, ok := g.structHelperName(member.GoType)
			if !ok || helperName == "" {
				helperName = member.Token
			}
			fmt.Fprintf(buf, "\t\t\tconverted, err := __able_struct_%s_from(coerced)\n", helperName)
		case g.nativeUnionInfoForGoType(member.GoType) != nil:
			inner := g.nativeUnionInfoForGoType(member.GoType)
			fmt.Fprintf(buf, "\t\t\tconverted, err := %s(rt, coerced)\n", inner.FromRuntimeHelper)
		case g.isNativeNullableValueType(member.GoType):
			if helper, ok := g.nativeNullableFromRuntimeHelper(member.GoType); ok {
				fmt.Fprintf(buf, "\t\t\tconverted, err := %s(coerced)\n", helper)
			}
		case member.GoType == "bool":
			fmt.Fprintf(buf, "\t\t\tconverted, err := bridge.AsBool(coerced)\n")
		case member.GoType == "string":
			fmt.Fprintf(buf, "\t\t\tconverted, err := bridge.AsString(coerced)\n")
		case member.GoType == "rune":
			fmt.Fprintf(buf, "\t\t\tconverted, err := bridge.AsRune(coerced)\n")
		case member.GoType == "float32":
			fmt.Fprintf(buf, "\t\t\traw, err := bridge.AsFloat(coerced)\n")
			fmt.Fprintf(buf, "\t\t\tconverted := float32(raw)\n")
		case member.GoType == "float64":
			fmt.Fprintf(buf, "\t\t\tconverted, err := bridge.AsFloat(coerced)\n")
		case member.GoType == "int":
			fmt.Fprintf(buf, "\t\t\traw, err := bridge.AsInt(coerced, bridge.NativeIntBits)\n")
			fmt.Fprintf(buf, "\t\t\tconverted := int(raw)\n")
		case member.GoType == "uint":
			fmt.Fprintf(buf, "\t\t\traw, err := bridge.AsUint(coerced, bridge.NativeIntBits)\n")
			fmt.Fprintf(buf, "\t\t\tconverted := uint(raw)\n")
		case g.isSignedIntegerType(member.GoType):
			fmt.Fprintf(buf, "\t\t\traw, err := bridge.AsInt(coerced, %d)\n", g.intBits(member.GoType))
			fmt.Fprintf(buf, "\t\t\tconverted := %s(raw)\n", member.GoType)
		case g.isUnsignedIntegerType(member.GoType):
			fmt.Fprintf(buf, "\t\t\traw, err := bridge.AsUint(coerced, %d)\n", g.intBits(member.GoType))
			fmt.Fprintf(buf, "\t\t\tconverted := %s(raw)\n", member.GoType)
		default:
			fmt.Fprintf(buf, "\t\t\tvar converted %s\n", member.GoType)
			fmt.Fprintf(buf, "\t\t\tvar err error\n")
			fmt.Fprintf(buf, "\t\t\t_ = converted\n")
			fmt.Fprintf(buf, "\t\t\t_ = err\n")
			fmt.Fprintf(buf, "\t\t\t_ = coerced\n")
			fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"unsupported union member type %s\")\n", member.GoType)
		}
		fmt.Fprintf(buf, "\t\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\treturn %s(converted), nil\n", member.WrapHelper)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	if fallbackRuntimeMember != nil {
		fmt.Fprintf(buf, "\treturn %s(value), nil\n", fallbackRuntimeMember.WrapHelper)
	}
	fmt.Fprintf(buf, "\treturn nil, fmt.Errorf(\"type mismatch: expected %s\")\n", info.TypeString)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value runtime.Value) %s {\n", info.FromRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.FromRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeUnionToRuntimeHelper(buf *bytes.Buffer, info *nativeUnionInfo) {
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value %s) (runtime.Value, error) {\n", info.ToRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing union value\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch raw := value.(type) {\n")
	for _, member := range info.Members {
		fmt.Fprintf(buf, "\tcase %s:\n", member.WrapperType)
		switch {
		case member.GoType == "runtime.Value":
			fmt.Fprintf(buf, "\t\treturn raw.Value, nil\n")
		case member.GoType == "struct{}":
			fmt.Fprintf(buf, "\t\treturn runtime.VoidValue{}, nil\n")
		case g.isMonoArrayType(member.GoType):
			spec, _ := g.monoArraySpecForGoType(member.GoType)
			fmt.Fprintf(buf, "\t\treturn %s(rt, raw.Value)\n", spec.ToRuntimeHelper)
		case g.nativeInterfaceInfoForGoType(member.GoType) != nil:
			iface := g.nativeInterfaceInfoForGoType(member.GoType)
			fmt.Fprintf(buf, "\t\treturn %s(rt, raw.Value)\n", iface.ToRuntimeHelper)
		case member.GoType == "runtime.ErrorValue":
			fmt.Fprintf(buf, "\t\treturn raw.Value, nil\n")
		case g.typeCategory(member.GoType) == "struct":
			helperName, ok := g.structHelperName(member.GoType)
			if !ok || helperName == "" {
				helperName = member.Token
			}
			fmt.Fprintf(buf, "\t\treturn __able_struct_%s_to(rt, raw.Value)\n", helperName)
		case g.nativeUnionInfoForGoType(member.GoType) != nil:
			inner := g.nativeUnionInfoForGoType(member.GoType)
			fmt.Fprintf(buf, "\t\treturn %s(rt, raw.Value)\n", inner.ToRuntimeHelper)
		case g.isNativeNullableValueType(member.GoType):
			if helper, ok := g.nativeNullableToRuntimeHelper(member.GoType); ok {
				fmt.Fprintf(buf, "\t\treturn %s(raw.Value), nil\n", helper)
			}
		case member.GoType == "bool":
			fmt.Fprintf(buf, "\t\treturn bridge.ToBool(raw.Value), nil\n")
		case member.GoType == "string":
			fmt.Fprintf(buf, "\t\treturn bridge.ToString(raw.Value), nil\n")
		case member.GoType == "rune":
			fmt.Fprintf(buf, "\t\treturn bridge.ToRune(raw.Value), nil\n")
		case member.GoType == "float32":
			fmt.Fprintf(buf, "\t\treturn bridge.ToFloat32(raw.Value), nil\n")
		case member.GoType == "float64":
			fmt.Fprintf(buf, "\t\treturn bridge.ToFloat64(raw.Value), nil\n")
		case member.GoType == "int":
			fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(raw.Value), runtime.IntegerType(\"isize\")), nil\n")
		case member.GoType == "uint":
			fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(raw.Value), runtime.IntegerType(\"usize\")), nil\n")
		case g.isSignedIntegerType(member.GoType):
			suffix, _ := g.integerTypeSuffix(member.GoType)
			fmt.Fprintf(buf, "\t\treturn bridge.ToInt(int64(raw.Value), runtime.IntegerType(%q)), nil\n", suffix)
		case g.isUnsignedIntegerType(member.GoType):
			suffix, _ := g.integerTypeSuffix(member.GoType)
			fmt.Fprintf(buf, "\t\treturn bridge.ToUint(uint64(raw.Value), runtime.IntegerType(%q)), nil\n", suffix)
		default:
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"unsupported union member type %s\")\n", member.GoType)
		}
	}
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"unsupported union carrier %%T\", value)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value %s) runtime.Value {\n", info.ToRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.ToRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}
