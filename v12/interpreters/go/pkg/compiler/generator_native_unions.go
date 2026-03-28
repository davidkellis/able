package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeUnionMember struct {
	GoType       string
	TypeExpr     ast.TypeExpression
	KeyPart      string
	Token        string
	WrapperType  string
	WrapHelper   string
	UnwrapHelper string
}

type nativeUnionInfo struct {
	Key                  string
	GoType               string
	TypeExpr             ast.TypeExpression
	MarkerMethod         string
	FromRuntimeHelper    string
	FromRuntimePanic     string
	ToRuntimeHelper      string
	ToRuntimePanic       string
	TypeString           string
	OrderedMemberGoTypes []string
	Members              []*nativeUnionMember
}

func (g *generator) nativeUnionInfoForGoType(goType string) *nativeUnionInfo {
	if g == nil || goType == "" {
		return nil
	}
	for _, info := range g.nativeUnions {
		if info != nil && info.GoType == goType {
			return info
		}
	}
	return nil
}

func (g *generator) nativeUnionMember(info *nativeUnionInfo, goType string) (*nativeUnionMember, bool) {
	if info == nil || goType == "" {
		return nil, false
	}
	for _, member := range info.Members {
		if member != nil && member.GoType == goType {
			return member, true
		}
	}
	return nil, false
}

func isNilTypeExpression(expr ast.TypeExpression) bool {
	simple, ok := expr.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "nil"
}

func nativeUnionNullableInnerTypeExpr(members []ast.TypeExpression) (ast.TypeExpression, bool) {
	if len(members) != 2 {
		return nil, false
	}
	if isNilTypeExpression(members[0]) {
		return members[1], true
	}
	if isNilTypeExpression(members[1]) {
		return members[0], true
	}
	return nil, false
}

func (g *generator) expandedUnionMembersInPackage(pkgName string, expr ast.TypeExpression) (string, []ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return "", nil, false
	}
	switch t := expr.(type) {
	case *ast.UnionTypeExpression:
		return pkgName, t.Members, t != nil
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return "", nil, false
		}
		def, ok := g.unions[t.Name.Name]
		if !ok || def == nil || len(def.GenericParams) > 0 {
			return "", nil, false
		}
		unionPkg := pkgName
		if recorded, ok := g.unionPackages[t.Name.Name]; ok && recorded != "" {
			unionPkg = recorded
		}
		return unionPkg, nativeUnionDefinitionMembers(def), true
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", nil, false
		}
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil || base.Name.Name == "" {
			return "", nil, false
		}
		def, ok := g.unions[base.Name.Name]
		if !ok || def == nil || len(def.GenericParams) == 0 || len(def.GenericParams) != len(t.Arguments) {
			return "", nil, false
		}
		unionPkg := pkgName
		if recorded, ok := g.unionPackages[base.Name.Name]; ok && recorded != "" {
			unionPkg = recorded
		}
		bindings := make(map[string]ast.TypeExpression, len(def.GenericParams))
		for idx, gp := range def.GenericParams {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" || idx >= len(t.Arguments) || t.Arguments[idx] == nil {
				return "", nil, false
			}
			bindings[gp.Name.Name] = t.Arguments[idx]
		}
		raw := nativeUnionDefinitionMembers(def)
		out := make([]ast.TypeExpression, 0, len(raw))
		for _, member := range raw {
			if member == nil {
				return "", nil, false
			}
			out = append(out, substituteTypeParams(member, bindings))
		}
		return unionPkg, out, true
	default:
		return "", nil, false
	}
}

func (g *generator) nativeUnionRuntimeMemberAcceptsActual(member *nativeUnionMember, actual string) bool {
	if g == nil || member == nil || member.GoType != "runtime.Value" || actual == "" {
		return false
	}
	if actual == "runtime.Value" || actual == "any" {
		return true
	}
	if ifaceExpr, ok := g.interfaceTypeExpr(member.TypeExpr); ok {
		if name, ok := typeExprBaseName(ifaceExpr); ok {
			return g.nativeCarrierImplementsInterface(actual, name)
		}
		return false
	}
	switch member.TypeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.FunctionTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		// Residual runtime.Value branches produced by generic/open members accept
		// any concrete static payload.
		return true
	}
	return false
}

func (g *generator) nativeUnionRuntimeMemberRequiresMatch(member *nativeUnionMember) bool {
	if g == nil || member == nil || member.GoType != "runtime.Value" || member.TypeExpr == nil {
		return false
	}
	if _, ok := g.interfaceTypeExpr(member.TypeExpr); ok {
		return true
	}
	switch member.TypeExpr.(type) {
	case *ast.FunctionTypeExpression:
		return true
	case *ast.GenericTypeExpression:
		return true
	case *ast.UnionTypeExpression:
		return true
	case *ast.ResultTypeExpression:
		return true
	}
	return false
}

func (g *generator) nativeUnionWrapExpr(expected, actual, expr string) (string, bool) {
	info := g.nativeUnionInfoForGoType(expected)
	if info == nil {
		return "", false
	}
	member, ok := g.nativeUnionMember(info, actual)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s(%s)", member.WrapHelper, expr), true
}

func (g *generator) nativeUnionPatternMemberType(subjectType string, patternType ast.TypeExpression, pkgName string) (string, bool) {
	info := g.nativeUnionInfoForGoType(subjectType)
	if info == nil || patternType == nil {
		return "", false
	}
	mapped, ok := g.lowerCarrierTypeInPackage(pkgName, patternType)
	if !ok || mapped == "" {
		return "", false
	}
	if mapped == subjectType {
		return mapped, true
	}
	if _, ok := g.nativeUnionMember(info, mapped); ok {
		return mapped, true
	}
	for _, member := range info.Members {
		if member == nil || member.GoType == "" {
			continue
		}
		if g.receiverGoTypeCompatible(mapped, member.GoType) {
			return member.GoType, true
		}
	}
	return "", false
}

func (g *generator) nativeUnionScalarFromRuntimeHelper(goType string) (string, bool) {
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		return helper, true
	}
	switch goType {
	case "bool":
		return "bridge.AsBool", true
	case "string":
		return "bridge.AsString", true
	case "rune":
		return "bridge.AsRune", true
	case "float64":
		return "bridge.AsFloat", true
	}
	return "", false
}

func (g *generator) nativeUnionTypeToken(goType string) string {
	if goType == "" {
		return "value"
	}
	token := goType
	if strings.HasPrefix(token, "*") {
		token = "ptr_" + strings.TrimPrefix(token, "*")
	}
	if strings.HasPrefix(token, "[]") {
		token = "slice_" + strings.TrimPrefix(token, "[]")
	}
	return sanitizeIdent(token)
}

func (g *generator) nativeUnionTypeExprInPackage(pkgName string, expr ast.TypeExpression) (*nativeUnionInfo, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	normalized := normalizeTypeExprForPackage(g, pkgName, expr)
	exprKey := normalizeTypeExprString(g, pkgName, normalized)
	if exprKey != "" {
		if info, ok := g.nativeUnionExprIndex[exprKey]; ok && info != nil {
			return info, true
		}
	}
	if unionPkg, members, ok := g.expandedUnionMembersInPackage(pkgName, expr); ok {
		if _, nullable := nativeUnionNullableInnerTypeExpr(members); nullable {
			return nil, false
		}
		info, ok := g.ensureNativeUnionInfo(unionPkg, members)
		if ok && info != nil && exprKey != "" {
			g.nativeUnionExprIndex[exprKey] = info
		}
		return info, ok
	}
	if t, ok := expr.(*ast.ResultTypeExpression); ok {
		info, ok := g.ensureNativeResultUnionInfo(pkgName, t)
		if ok && info != nil && exprKey != "" {
			g.nativeUnionExprIndex[exprKey] = info
		}
		return info, ok
	}
	return nil, false
}

func (g *generator) nativeUnionExpectedTypeForExpr(ctx *compileContext, expected string, expr ast.Expression) string {
	info := g.nativeUnionInfoForGoType(expected)
	if ctx == nil || expr == nil {
		return expected
	}
	switch e := expr.(type) {
	case *ast.StructLiteral:
		if e == nil || e.StructType == nil || e.StructType.Name == "" {
			return expected
		}
		if info != nil {
			var matched string
			for _, member := range info.Members {
				if member == nil {
					continue
				}
				memberInfo := g.structInfoByGoName(member.GoType)
				if memberInfo == nil || memberInfo.Name != e.StructType.Name {
					continue
				}
				if matched != "" && matched != member.GoType {
					matched = ""
					break
				}
				matched = member.GoType
			}
			if matched != "" {
				return matched
			}
		}
		structTypeExpr := g.staticStructLiteralTypeExpr(ctx, e, "")
		if structTypeExpr == nil {
			structTypeExpr = ast.Ty(e.StructType.Name)
		}
		mapped, ok := g.lowerCarrierType(ctx, structTypeExpr)
		if !ok {
			return expected
		}
		if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil && g.nativeInterfaceAcceptsActual(iface, mapped) {
			return mapped
		}
		if innerType, nullable := g.nativeNullableValueInnerType(expected); nullable && innerType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(mapped) {
			return mapped
		}
		if expected == "runtime.ErrorValue" && g.isNativeErrorCarrierType(mapped) {
			return mapped
		}
		if info != nil {
			if _, ok := g.nativeUnionMember(info, mapped); ok {
				return mapped
			}
			for _, member := range info.Members {
				if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, mapped) {
					return mapped
				}
			}
			if _, ok := g.nativeUnionMember(info, "runtime.ErrorValue"); ok && g.isNativeErrorCarrierType(mapped) {
				return mapped
			}
		}
	case *ast.ArrayLiteral:
		if info != nil {
			for _, member := range info.Members {
				if member != nil && g.isMonoArrayType(member.GoType) {
					return member.GoType
				}
			}
		}
		if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil && g.nativeInterfaceAcceptsActual(iface, "*Array") {
			return "*Array"
		}
		if info != nil {
			if _, ok := g.nativeUnionMember(info, "*Array"); ok {
				return "*Array"
			}
			for _, member := range info.Members {
				if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, "*Array") {
					return "*Array"
				}
			}
		}
	}
	return expected
}

func (g *generator) expectedUnionMemberTypeExpr(pkgName string, expectedTypeExpr ast.TypeExpression, expectedGoType string, nominalName string) ast.TypeExpression {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(nominalName) == "" {
		return nil
	}
	candidates := make([]ast.TypeExpression, 0, 2)
	if expectedTypeExpr != nil {
		candidates = append(candidates, expectedTypeExpr)
	}
	if expectedGoType != "" && expectedGoType != "runtime.Value" && expectedGoType != "any" {
		if expr, ok := g.typeExprForGoType(expectedGoType); ok && expr != nil {
			candidates = append(candidates, expr)
		}
	}
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, candidate)
		key := normalizeTypeExprString(g, pkgName, normalized)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		_, members, ok := g.expandedUnionMembersInPackage(pkgName, normalized)
		if !ok {
			continue
		}
		for _, member := range members {
			member = normalizeTypeExprForPackage(g, pkgName, member)
			if member == nil {
				continue
			}
			baseName, ok := typeExprBaseName(member)
			if !ok || baseName != nominalName {
				continue
			}
			if info, ok := g.structInfoForTypeExpr(pkgName, member); ok && info != nil && info.Name == nominalName {
				return member
			}
		}
	}
	return nil
}

func (g *generator) isNativeErrorCarrierType(goType string) bool {
	if goType == "runtime.ErrorValue" {
		return true
	}
	return g.nativeCarrierImplementsInterface(goType, "Error")
}

func (g *generator) nativeUnionOrElseMembers(goType string) (success *nativeUnionMember, failure *nativeUnionMember, ok bool) {
	info := g.nativeUnionInfoForGoType(goType)
	if info == nil || len(info.Members) != 2 {
		return nil, nil, false
	}
	for _, member := range info.Members {
		if member == nil {
			return nil, nil, false
		}
		if g.isNativeErrorCarrierType(member.GoType) {
			if failure != nil {
				return nil, nil, false
			}
			failure = member
			continue
		}
		if success != nil {
			return nil, nil, false
		}
		success = member
	}
	if success == nil || failure == nil {
		return nil, nil, false
	}
	return success, failure, true
}

func nativeUnionDefinitionMembers(def *ast.UnionDefinition) []ast.TypeExpression {
	if def == nil {
		return nil
	}
	if len(def.Variants) == 1 {
		if unionExpr, ok := def.Variants[0].(*ast.UnionTypeExpression); ok && unionExpr != nil {
			return unionExpr.Members
		}
	}
	return def.Variants
}

func (g *generator) ensureNativeUnionInfo(pkgName string, members []ast.TypeExpression) (*nativeUnionInfo, bool) {
	if g == nil || len(members) < 2 {
		return nil, false
	}
	mapper := NewTypeMapper(g, pkgName)
	memberSpecs := make([]*nativeUnionMember, 0, len(members))
	keyParts := make([]string, 0, len(members))
	typeParts := make([]string, 0, len(members))
	seen := make(map[string]struct{}, len(members))
	for _, memberExpr := range members {
		if memberExpr == nil {
			return nil, false
		}
		memberType, ok := mapper.Map(memberExpr)
		if !ok || memberType == "" {
			return nil, false
		}
		if memberType == "any" || memberType == "runtime.Value" {
			if g.typeExprIsConcreteInPackage(pkgName, memberExpr) {
				return nil, false
			}
			memberType = "runtime.Value"
		}
		if _, exists := seen[memberType]; exists {
			return nil, false
		}
		seen[memberType] = struct{}{}
		keyPart := memberType
		if memberType == "runtime.Value" {
			keyPart = memberType + "<" + typeExpressionToString(memberExpr) + ">"
		}
		memberSpecs = append(memberSpecs, &nativeUnionMember{
			GoType:   memberType,
			TypeExpr: memberExpr,
			KeyPart:  keyPart,
			Token:    g.nativeUnionTypeToken(memberType),
		})
		keyParts = append(keyParts, keyPart)
		typeParts = append(typeParts, typeExpressionToString(memberExpr))
	}
	sort.Strings(keyParts)
	key := strings.Join(keyParts, "|")
	if info, ok := g.nativeUnions[key]; ok && info != nil {
		return info, true
	}
	baseName := "__able_union_" + strings.Join(keyParts, "_or_")
	baseName = sanitizeIdent(baseName)
	info := &nativeUnionInfo{
		Key:                  key,
		GoType:               baseName,
		TypeExpr:             ast.NewUnionTypeExpression(append([]ast.TypeExpression(nil), members...)),
		MarkerMethod:         baseName + "_marker",
		FromRuntimeHelper:    baseName + "_from_value",
		FromRuntimePanic:     baseName + "_from_runtime_value_or_panic",
		ToRuntimeHelper:      baseName + "_to_value",
		ToRuntimePanic:       baseName + "_to_runtime_value_or_panic",
		TypeString:           strings.Join(typeParts, " | "),
		OrderedMemberGoTypes: append([]string(nil), keyParts...),
		Members:              memberSpecs,
	}
	for _, member := range info.Members {
		member.WrapperType = info.GoType + "_variant_" + member.Token
		member.WrapHelper = info.GoType + "_wrap_" + member.Token
		member.UnwrapHelper = info.GoType + "_as_" + member.Token
	}
	g.nativeUnions[key] = info
	return info, true
}

func (g *generator) ensureNativeResultUnionInfo(pkgName string, result *ast.ResultTypeExpression) (*nativeUnionInfo, bool) {
	if g == nil || result == nil || result.InnerType == nil {
		return nil, false
	}
	mapper := NewTypeMapper(g, pkgName)
	innerType, ok := mapper.Map(result.InnerType)
	if !ok || innerType == "" {
		return nil, false
	}
	if innerType == "any" || innerType == "runtime.Value" {
		if g.typeExprIsConcreteInPackage(pkgName, result.InnerType) {
			return nil, false
		}
		innerType = "runtime.Value"
	}
	keyParts := []string{"runtime.ErrorValue", innerType}
	if innerType == "runtime.Value" {
		keyParts[1] = innerType + "<" + typeExpressionToString(result.InnerType) + ">"
	}
	sort.Strings(keyParts)
	key := strings.Join(keyParts, "|")
	if info, ok := g.nativeUnions[key]; ok && info != nil {
		return info, true
	}
	memberSpecs := []*nativeUnionMember{
		{
			GoType:   "runtime.ErrorValue",
			TypeExpr: ast.Ty("Error"),
			KeyPart:  "runtime.ErrorValue",
			Token:    g.nativeUnionTypeToken("runtime_ErrorValue"),
		},
		{
			GoType:   innerType,
			TypeExpr: result.InnerType,
			KeyPart: func() string {
				if innerType == "runtime.Value" {
					return innerType + "<" + typeExpressionToString(result.InnerType) + ">"
				}
				return innerType
			}(),
			Token: g.nativeUnionTypeToken(innerType),
		},
	}
	baseName := "__able_union_" + strings.Join(keyParts, "_or_")
	baseName = sanitizeIdent(baseName)
	info := &nativeUnionInfo{
		Key:                  key,
		GoType:               baseName,
		TypeExpr:             ast.NewResultTypeExpression(result.InnerType),
		MarkerMethod:         baseName + "_marker",
		FromRuntimeHelper:    baseName + "_from_value",
		FromRuntimePanic:     baseName + "_from_runtime_value_or_panic",
		ToRuntimeHelper:      baseName + "_to_value",
		ToRuntimePanic:       baseName + "_to_runtime_value_or_panic",
		TypeString:           "Error | " + typeExpressionToString(result.InnerType),
		OrderedMemberGoTypes: append([]string(nil), keyParts...),
		Members:              memberSpecs,
	}
	for _, member := range info.Members {
		member.WrapperType = info.GoType + "_variant_" + member.Token
		member.WrapHelper = info.GoType + "_wrap_" + member.Token
		member.UnwrapHelper = info.GoType + "_as_" + member.Token
	}
	g.nativeUnions[key] = info
	return info, true
}

func (g *generator) sortedNativeUnionKeys() []string {
	if g == nil || len(g.nativeUnions) == 0 {
		return nil
	}
	keys := make([]string, 0, len(g.nativeUnions))
	for key := range g.nativeUnions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
