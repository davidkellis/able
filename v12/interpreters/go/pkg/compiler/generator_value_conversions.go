package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) expectRuntimeValueExpr(valueExpr string, expected string) (string, bool) {
	if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil {
		return fmt.Sprintf("%s(%s)", iface.FromRuntimePanic, valueExpr), true
	}
	if callable := g.nativeCallableInfoForGoType(expected); callable != nil {
		return fmt.Sprintf("%s(%s)", callable.FromRuntimePanic, valueExpr), true
	}
	if union := g.nativeUnionInfoForGoType(expected); union != nil {
		return fmt.Sprintf("%s(%s)", union.FromRuntimePanic, valueExpr), true
	}
	if expected == "runtime.ErrorValue" {
		return fmt.Sprintf("func() runtime.ErrorValue { if v, ok, nilPtr := __able_runtime_error_value(%s); ok && !nilPtr { return v }; return bridge.ErrorValue(__able_runtime, %s) }()", valueExpr, valueExpr), true
	}
	if helper, ok := g.nativeNullableFromRuntimePanicHelper(expected); ok {
		return fmt.Sprintf("%s(%s)", helper, valueExpr), true
	}
	switch g.typeCategory(expected) {
	case "bool":
		return fmt.Sprintf("func() bool { val := %s; v, err := bridge.AsBool(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "string":
		return fmt.Sprintf("func() string { val := %s; v, err := bridge.AsString(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "rune":
		return fmt.Sprintf("func() rune { val := %s; v, err := bridge.AsRune(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "float32":
		return fmt.Sprintf("func() float32 { val := %s; v, err := bridge.AsFloat(val); if err != nil { panic(err) }; return float32(v) }()", valueExpr), true
	case "float64":
		return fmt.Sprintf("func() float64 { val := %s; v, err := bridge.AsFloat(val); if err != nil { panic(err) }; return v }()", valueExpr), true
	case "int":
		return fmt.Sprintf("func() int { val := %s; v, err := bridge.AsInt(val, bridge.NativeIntBits); if err != nil { panic(err) }; return int(v) }()", valueExpr), true
	case "uint":
		return fmt.Sprintf("func() uint { val := %s; v, err := bridge.AsUint(val, bridge.NativeIntBits); if err != nil { panic(err) }; return uint(v) }()", valueExpr), true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(expected)
		return fmt.Sprintf("func() %s { val := %s; v, err := bridge.AsInt(val, %d); if err != nil { panic(err) }; return %s(v) }()", expected, valueExpr, bits, expected), true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(expected)
		return fmt.Sprintf("func() %s { val := %s; v, err := bridge.AsUint(val, %d); if err != nil { panic(err) }; return %s(v) }()", expected, valueExpr, bits, expected), true
	case "struct":
		baseName, ok := g.structBaseName(expected)
		if !ok {
			baseName = strings.TrimPrefix(expected, "*")
		}
		return fmt.Sprintf("func() %s { val := %s; v, err := __able_struct_%s_from(val); if err != nil { panic(err) }; return v }()", expected, valueExpr, baseName), true
	}
	return "", false
}

func (g *generator) expectRuntimeValueExprLines(ctx *compileContext, valueExpr string, expected string) ([]string, string, bool) {
	if expected == "any" {
		return nil, valueExpr, true
	}
	if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil {
		valTemp := ctx.newTemp()
		vTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", vTemp, errTemp, iface.FromRuntimeHelper, valTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	}
	if callable := g.nativeCallableInfoForGoType(expected); callable != nil {
		valTemp := ctx.newTemp()
		vTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", vTemp, errTemp, callable.FromRuntimeHelper, valTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	}
	if union := g.nativeUnionInfoForGoType(expected); union != nil {
		valTemp := ctx.newTemp()
		vTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", vTemp, errTemp, union.FromRuntimeHelper, valTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	}
	if expected == "runtime.ErrorValue" {
		valTemp := ctx.newTemp()
		vTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		nilPtrTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s, %s := __able_runtime_error_value(%s)", vTemp, okTemp, nilPtrTemp, valTemp),
			fmt.Sprintf("if !%s || %s { %s = bridge.ErrorValue(__able_runtime, %s) }", okTemp, nilPtrTemp, vTemp, valTemp),
		}
		return lines, vTemp, true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(expected); ok {
		valTemp := ctx.newTemp()
		vTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := %s(%s)", vTemp, errTemp, helper, valTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	}
	valTemp := ctx.newTemp()
	vTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	switch g.typeCategory(expected) {
	case "bool":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsBool(%s)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	case "string":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsString(%s)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	case "rune":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsRune(%s)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	case "float32":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsFloat(%s)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("float32(%s)", vTemp), true
	case "float64":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsFloat(%s)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	case "int":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsInt(%s, bridge.NativeIntBits)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("int(%s)", vTemp), true
	case "uint":
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsUint(%s, bridge.NativeIntBits)", vTemp, errTemp, valTemp),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("uint(%s)", vTemp), true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(expected)
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsInt(%s, %d)", vTemp, errTemp, valTemp, bits),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("%s(%s)", expected, vTemp), true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(expected)
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := bridge.AsUint(%s, %d)", vTemp, errTemp, valTemp, bits),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("%s(%s)", expected, vTemp), true
	case "struct":
		baseName, ok := g.structBaseName(expected)
		if !ok {
			baseName = strings.TrimPrefix(expected, "*")
		}
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valTemp, valueExpr),
			fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", vTemp, errTemp, baseName, valTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, vTemp, true
	}
	return nil, "", false
}

func (g *generator) memberName(member ast.Expression) string {
	if ident, ok := member.(*ast.Identifier); ok && ident != nil {
		return ident.Name
	}
	return ""
}

func (g *generator) stringBuiltinFieldAccess(objectExpr, fieldName string) (string, string, bool) {
	switch fieldName {
	case "len_bytes":
		return fmt.Sprintf("uint(len(%s))", objectExpr), "uint", true
	}
	return "", "", false
}

func (g *generator) isArrayStructType(goType string) bool {
	if g == nil || g.typeCategory(goType) != "struct" {
		return false
	}
	baseName, ok := g.structBaseName(goType)
	return ok && baseName == "Array"
}

func (g *generator) structInfoByGoName(goName string) *structInfo {
	if goName == "" {
		return nil
	}
	if strings.HasPrefix(goName, "*") {
		goName = strings.TrimPrefix(goName, "*")
	}
	for _, info := range g.structs {
		if info != nil && info.GoName == goName {
			return info
		}
	}
	return nil
}

func (g *generator) nativeIndexIntExpr(expr string, goType string) (string, bool) {
	switch g.typeCategory(goType) {
	case "int", "int8", "int16", "int32", "uint8", "uint16":
		return fmt.Sprintf("int(%s)", expr), true
	case "int64":
		return fmt.Sprintf("func() int { raw := int64(%s); idx := int(raw); if int64(idx) != raw { panic(fmt.Errorf(\"index overflows native int\")) }; return idx }()", expr), true
	case "uint", "uint32", "uint64":
		return fmt.Sprintf("func() int { raw := uint64(%s); idx := int(raw); if uint64(idx) != raw { panic(fmt.Errorf(\"index overflows native int\")) }; return idx }()", expr), true
	default:
		return "", false
	}
}

func (g *generator) appendIndexIntLines(ctx *compileContext, lines []string, idxExpr string, idxType string, idxTemp string, indexTemp string) ([]string, bool) {
	lines = append(lines, fmt.Sprintf("%s := %s", idxTemp, idxExpr))
	if nativeExpr, ok := g.nativeIndexIntExpr(idxTemp, idxType); ok {
		lines = append(lines, fmt.Sprintf("%s := %s", indexTemp, nativeExpr))
		return lines, true
	}
	idxConvLines, idxValueExpr, ok := g.runtimeValueLines(ctx, idxTemp, idxType)
	if !ok {
		return nil, false
	}
	lines = append(lines, idxConvLines...)
	idxRuntimeTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", idxRuntimeTemp, idxValueExpr))
	idxRawTemp := ctx.newTemp()
	idxErrTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := bridge.AsInt(%s, 64)", idxRawTemp, idxErrTemp, idxRuntimeTemp))
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, idxErrTemp))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, false
	}
	lines = append(lines, controlLines...)
	lines = append(lines, fmt.Sprintf("%s := int(%s)", indexTemp, idxRawTemp))
	return lines, true
}

func (g *generator) runtimeValueExpr(expr string, goType string) (string, bool) {
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		return fmt.Sprintf("%s(%s)", iface.ToRuntimePanic, expr), true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		return fmt.Sprintf("%s(%s)", callable.ToRuntimePanic, expr), true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		return fmt.Sprintf("%s(%s)", union.ToRuntimePanic, expr), true
	}
	if goType == "runtime.ErrorValue" {
		return expr, true
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		return fmt.Sprintf("%s(%s)", helper, expr), true
	}
	switch g.typeCategory(goType) {
	case "runtime":
		return expr, true
	case "any":
		return fmt.Sprintf("__able_any_to_value(%s)", expr), true
	case "void":
		return fmt.Sprintf("func() runtime.Value { _ = %s; return runtime.VoidValue{} }()", expr), true
	case "bool":
		return fmt.Sprintf("bridge.ToBool(%s)", expr), true
	case "string":
		return fmt.Sprintf("bridge.ToString(%s)", expr), true
	case "rune":
		return fmt.Sprintf("bridge.ToRune(%s)", expr), true
	case "float32":
		return fmt.Sprintf("bridge.ToFloat32(%s)", expr), true
	case "float64":
		return fmt.Sprintf("bridge.ToFloat64(%s)", expr), true
	case "int":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))", expr), true
	case "uint":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))", expr), true
	case "int8":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\"))", expr), true
	case "int16":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\"))", expr), true
	case "int32":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\"))", expr), true
	case "int64":
		return fmt.Sprintf("bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\"))", expr), true
	case "uint8":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\"))", expr), true
	case "uint16":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\"))", expr), true
	case "uint32":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\"))", expr), true
	case "uint64":
		return fmt.Sprintf("bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\"))", expr), true
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		if strings.HasPrefix(goType, "*") {
			return fmt.Sprintf("__able_any_to_value(%s)", expr), true
		}
		return fmt.Sprintf("func() runtime.Value { if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }; v, err := __able_struct_%s_to(__able_runtime, %s); if err != nil { panic(err) }; return v }()", baseName, expr), true
	default:
		return "", false
	}
}

func (g *generator) runtimeValueLines(ctx *compileContext, expr string, goType string) ([]string, string, bool) {
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convTemp, errTemp, iface.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convTemp, true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convTemp, errTemp, callable.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convTemp, true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convTemp, errTemp, union.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convTemp, true
	}
	if goType == "runtime.ErrorValue" {
		return nil, expr, true
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		convTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s(%s)", convTemp, helper, expr),
		}
		return lines, convTemp, true
	}
	switch g.typeCategory(goType) {
	case "void":
		return []string{fmt.Sprintf("_ = %s", expr)}, "runtime.VoidValue{}", true
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		if strings.HasPrefix(goType, "*") {
			convTemp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, expr),
			}
			return lines, convTemp, true
		}
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			"if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }",
			fmt.Sprintf("%s, %s := __able_struct_%s_to(__able_runtime, %s)", convTemp, errTemp, baseName, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convTemp, true
	default:
		converted, ok := g.runtimeValueExpr(expr, goType)
		return nil, converted, ok
	}
}
