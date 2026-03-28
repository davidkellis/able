package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type nominalCoercionInfo struct {
	Key            string
	HelperName     string
	SeenHelperName string
	ActualGoType   string
	ExpectedGoType string
}

func (g *generator) ensureNominalCoercionInfo(actual string, expected string) *nominalCoercionInfo {
	if g == nil || actual == "" || expected == "" || actual == expected || !g.nominalStructCarrierCoercible(expected, actual) {
		return nil
	}
	if g.nominalCoercions == nil {
		g.nominalCoercions = make(map[string]*nominalCoercionInfo)
	}
	key := actual + "->" + expected
	if existing := g.nominalCoercions[key]; existing != nil {
		return existing
	}
	actualBase := strings.TrimPrefix(actual, "*")
	expectedBase := strings.TrimPrefix(expected, "*")
	info := &nominalCoercionInfo{
		Key:            key,
		HelperName:     g.mangler.unique(fmt.Sprintf("__able_nominal_coerce_%s_to_%s", sanitizeIdent(actualBase), sanitizeIdent(expectedBase))),
		SeenHelperName: g.mangler.unique(fmt.Sprintf("__able_nominal_coerce_%s_to_%s_seen", sanitizeIdent(actualBase), sanitizeIdent(expectedBase))),
		ActualGoType:   actual,
		ExpectedGoType: expected,
	}
	g.nominalCoercions[key] = info
	actualInfo := g.structInfoByGoName(actual)
	expectedInfo := g.structInfoByGoName(expected)
	if actualInfo == nil || expectedInfo == nil {
		return info
	}
	for _, expectedField := range expectedInfo.Fields {
		actualField := g.fieldInfo(actualInfo, expectedField.Name)
		if actualField == nil {
			continue
		}
		if actualField.GoType == expectedField.GoType {
			continue
		}
		if g.nominalStructCarrierCoercible(expectedField.GoType, actualField.GoType) {
			g.ensureNominalCoercionInfo(actualField.GoType, expectedField.GoType)
		}
	}
	return info
}

func (g *generator) sortedNominalCoercions() []*nominalCoercionInfo {
	if g == nil || len(g.nominalCoercions) == 0 {
		return nil
	}
	infos := make([]*nominalCoercionInfo, 0, len(g.nominalCoercions))
	for _, info := range g.nominalCoercions {
		if info != nil {
			infos = append(infos, info)
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].ActualGoType != infos[j].ActualGoType {
			return infos[i].ActualGoType < infos[j].ActualGoType
		}
		if infos[i].ExpectedGoType != infos[j].ExpectedGoType {
			return infos[i].ExpectedGoType < infos[j].ExpectedGoType
		}
		return infos[i].HelperName < infos[j].HelperName
	})
	return infos
}

func (g *generator) renderNominalCoercions(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	infos := g.sortedNominalCoercions()
	if len(infos) == 0 {
		return
	}
	for _, info := range infos {
		g.renderNominalCoercion(buf, info)
	}
}

func (g *generator) renderNominalCoercion(buf *bytes.Buffer, info *nominalCoercionInfo) {
	if g == nil || buf == nil || info == nil {
		return
	}
	actualInfo := g.structInfoByGoName(info.ActualGoType)
	expectedInfo := g.structInfoByGoName(info.ExpectedGoType)
	if actualInfo == nil || expectedInfo == nil {
		return
	}
	fmt.Fprintf(buf, "func %s(value %s) (%s, error) {\n", info.HelperName, info.ActualGoType, info.ExpectedGoType)
	fmt.Fprintf(buf, "\treturn %s(value, map[any]any{})\n", info.SeenHelperName)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value %s, seen map[any]any) (%s, error) {\n", info.SeenHelperName, info.ActualGoType, info.ExpectedGoType)
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif existing, ok := seen[value]; ok {\n")
	fmt.Fprintf(buf, "\t\tif converted, ok := existing.(%s); ok {\n", info.ExpectedGoType)
	fmt.Fprintf(buf, "\t\t\treturn converted, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tout := &%s{}\n", expectedInfo.GoName)
	fmt.Fprintf(buf, "\tseen[value] = out\n")
	tempCounter := 0
	for _, expectedField := range expectedInfo.Fields {
		actualField := g.fieldInfo(actualInfo, expectedField.Name)
		if actualField == nil {
			continue
		}
		g.renderNominalCoercionField(buf, info, actualField, expectedField, &tempCounter)
	}
	fmt.Fprintf(buf, "\treturn out, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNominalCoercionField(buf *bytes.Buffer, info *nominalCoercionInfo, actualField *fieldInfo, expectedField fieldInfo, tempCounter *int) {
	if g == nil || buf == nil || info == nil || actualField == nil || tempCounter == nil {
		return
	}
	actualExpr := "value." + actualField.GoName
	assignTarget := "out." + expectedField.GoName
	if actualField.GoType == expectedField.GoType {
		fmt.Fprintf(buf, "\t%s = %s\n", assignTarget, actualExpr)
		return
	}
	if g.nominalStructCarrierCoercible(expectedField.GoType, actualField.GoType) {
		nested := g.ensureNominalCoercionInfo(actualField.GoType, expectedField.GoType)
		if nested != nil {
			errName := fmt.Sprintf("__able_nominal_err_%d", *tempCounter)
			valueName := fmt.Sprintf("__able_nominal_value_%d", *tempCounter)
			*tempCounter++
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\t%s, %s := %s(%s, seen)\n", valueName, errName, nested.SeenHelperName, actualExpr)
			fmt.Fprintf(buf, "\t\tif %s != nil {\n", errName)
			fmt.Fprintf(buf, "\t\t\treturn nil, %s\n", errName)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\t%s = %s\n", assignTarget, valueName)
			fmt.Fprintf(buf, "\t}\n")
			return
		}
	}
	if g.renderNativeInterfaceDirectCoercionError(buf, "\t", fmt.Sprintf("__able_nominal_iface_%d", *tempCounter), actualExpr, actualField.GoType, expectedField.GoType, "nil") {
		fmt.Fprintf(buf, "\t%s = __able_nominal_iface_%d\n", assignTarget, *tempCounter)
		*tempCounter++
		return
	}
	runtimeValueName := fmt.Sprintf("__able_nominal_runtime_%d", *tempCounter)
	*tempCounter++
	fmt.Fprintf(buf, "\tvar %s runtime.Value\n", runtimeValueName)
	fmt.Fprintf(buf, "\t{\n")
	if g.nominalFieldNeedsRuntimeContext(actualField.GoType) {
		fmt.Fprintf(buf, "\t\trt := __able_runtime\n")
	}
	g.renderValueToRuntimeAssign(buf, actualExpr, actualField.GoType, runtimeValueName)
	fmt.Fprintf(buf, "\t}\n")
	if expectedField.GoType == "runtime.Value" {
		fmt.Fprintf(buf, "\t%s = %s\n", assignTarget, runtimeValueName)
	} else {
		fmt.Fprintf(buf, "\t{\n")
		g.renderValueConversion(buf, "\t\t", runtimeValueName, expectedField.GoType, assignTarget, "nil")
		fmt.Fprintf(buf, "\t}\n")
	}
}

func (g *generator) nominalFieldNeedsRuntimeContext(goType string) bool {
	if g == nil || goType == "" {
		return false
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		return true
	}
	switch g.typeCategory(goType) {
	case "interface", "union":
		return true
	default:
		return false
	}
}

func (g *generator) coerceNominalStructFamilyLines(ctx *compileContext, expr string, actual string, expected string) ([]string, string, bool) {
	if g == nil || ctx == nil || expr == "" || !g.nominalStructCarrierCoercible(expected, actual) {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	info := g.ensureNominalCoercionInfo(actual, expected)
	if info == nil {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := %s(%s)", resultTemp, errTemp, info.HelperName, expr),
		fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
	}
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp, true
}
