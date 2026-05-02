package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) collectMethodsDefinition(def *ast.MethodsDefinition, mapper *TypeMapper, pkgName string) {
	if def == nil || def.TargetType == nil || mapper == nil {
		return
	}
	targetType := g.methodsTargetTypeExpr(pkgName, def)
	targetName, ok := g.methodTargetName(targetType)
	if !ok || targetName == "" {
		return
	}
	if info, ok := g.structInfoForTypeName(pkgName, targetName); ok && (info == nil || !info.Supported) {
		return
	}
	if g.methods == nil {
		g.methods = make(map[string]map[string][]*methodInfo)
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil || fn.ID.Name == "" {
			continue
		}
		methodName := fn.ID.Name
		expectsSelf := methodDefinitionExpectsSelf(fn)
		goName := g.mangler.unique(fmt.Sprintf("method_%s_%s", sanitizeIdent(targetName), sanitizeIdent(methodName)))
		info := &functionInfo{
			Name:       fmt.Sprintf("%s.%s", targetName, methodName),
			Package:    pkgName,
			GoName:     goName,
			Definition: fn,
		}
		g.fillMethodInfo(info, mapper, targetType, expectsSelf)
		method := &methodInfo{
			TargetName:  targetName,
			TargetType:  targetType,
			MethodName:  methodName,
			ExpectsSelf: expectsSelf,
			Info:        info,
		}
		if expectsSelf && len(info.Params) > 0 {
			method.ReceiverType = info.Params[0].GoType
		}
		if g.methods[targetName] == nil {
			g.methods[targetName] = make(map[string][]*methodInfo)
		}
		g.methods[targetName][methodName] = append(g.methods[targetName][methodName], method)
		g.methodList = append(g.methodList, method)
	}
}

func (g *generator) methodsTargetTypeExpr(pkgName string, def *ast.MethodsDefinition) ast.TypeExpression {
	if g == nil || def == nil || def.TargetType == nil {
		return nil
	}
	targetType := normalizeTypeExprForPackage(g, pkgName, def.TargetType)
	if len(def.GenericParams) == 0 {
		return targetType
	}
	if _, ok := targetType.(*ast.GenericTypeExpression); ok {
		return targetType
	}
	base, ok := targetType.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil || base.Name.Name == "" {
		return targetType
	}
	args := make([]ast.TypeExpression, 0, len(def.GenericParams))
	for _, gp := range def.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return targetType
		}
		args = append(args, ast.Ty(gp.Name.Name))
	}
	return ast.NewGenericTypeExpression(ast.Ty(base.Name.Name), args)
}

func (g *generator) expandTypeAliasForPackage(pkgName string, expr ast.TypeExpression) ast.TypeExpression {
	_, expanded := g.expandTypeAliasContextForPackage(pkgName, expr)
	return expanded
}

func (g *generator) expandTypeAliasOnceForPackage(pkgName string, expr ast.TypeExpression) (ast.TypeExpression, string, string, bool) {
	if g == nil || expr == nil {
		return expr, pkgName, "", false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || strings.TrimSpace(t.Name.Name) == "" {
			return expr, pkgName, "", false
		}
		aliasPkg, aliasName, target, params, ok := g.typeAliasTargetForPackage(pkgName, t.Name.Name)
		if !ok || target == nil {
			return expr, pkgName, "", false
		}
		if len(params) != 0 {
			return expr, pkgName, "", false
		}
		return normalizeTypeExprForPackage(g, aliasPkg, target), aliasPkg, aliasName, true
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr, pkgName, "", false
		}
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil || strings.TrimSpace(base.Name.Name) == "" {
			return expr, pkgName, "", false
		}
		aliasPkg, aliasName, target, params, ok := g.typeAliasTargetForPackage(pkgName, base.Name.Name)
		if !ok || target == nil {
			return expr, pkgName, "", false
		}
		if len(params) == 0 {
			args := make([]ast.TypeExpression, 0, len(t.Arguments))
			for _, arg := range t.Arguments {
				if arg == nil {
					return expr, pkgName, "", false
				}
				args = append(args, normalizeTypeExprForPackage(g, pkgName, arg))
			}
			expanded := ast.NewGenericTypeExpression(cloneTypeExpr(target), args)
			return normalizeTypeExprForPackage(g, aliasPkg, expanded), aliasPkg, aliasName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">", true
		}
		if len(params) != len(t.Arguments) {
			return expr, pkgName, "", false
		}
		bindings := make(map[string]ast.TypeExpression, len(params))
		for idx, gp := range params {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" || t.Arguments[idx] == nil {
				return expr, pkgName, "", false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, aliasPkg, t.Arguments[idx])
		}
		expanded := substituteTypeParams(target, bindings)
		return normalizeTypeExprForPackage(g, aliasPkg, expanded), aliasPkg, aliasName + "<" + normalizeTypeExprListKey(g, aliasPkg, t.Arguments) + ">", true
	default:
		return expr, pkgName, "", false
	}
}

func (g *generator) typeAliasTargetForPackage(pkgName string, aliasName string) (string, string, ast.TypeExpression, []*ast.GenericParameter, bool) {
	if g == nil {
		return "", "", nil, nil, false
	}
	aliasPkg := strings.TrimSpace(pkgName)
	aliasName = strings.TrimSpace(aliasName)
	if aliasName == "" {
		return "", "", nil, nil, false
	}
	if target, params, ok := g.lookupTypeAlias(aliasPkg, aliasName); ok {
		return aliasPkg, aliasName, target, params, true
	}
	sourcePkg, sourceName := g.importedSelectorSourceTypeAlias(aliasPkg, aliasName)
	if sourcePkg == "" || sourceName == "" {
		return "", "", nil, nil, false
	}
	target, params, ok := g.lookupTypeAlias(sourcePkg, sourceName)
	if !ok {
		return sourcePkg, sourceName, ast.NewSimpleTypeExpression(ast.NewIdentifier(sourceName)), nil, true
	}
	return sourcePkg, sourceName, target, params, true
}

func (g *generator) lookupTypeAlias(pkgName string, aliasName string) (ast.TypeExpression, []*ast.GenericParameter, bool) {
	if g == nil || g.typeAliases == nil {
		return nil, nil, false
	}
	perPkg := g.typeAliases[pkgName]
	if perPkg == nil {
		return nil, nil, false
	}
	target, ok := perPkg[aliasName]
	if !ok || target == nil {
		return nil, nil, false
	}
	var params []*ast.GenericParameter
	if g.typeAliasGenericParams != nil && g.typeAliasGenericParams[pkgName] != nil {
		params = g.typeAliasGenericParams[pkgName][aliasName]
	}
	return target, params, true
}

func (g *generator) importedSelectorSourceTypeAlias(pkgName string, localName string) (string, string) {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(localName) == "" {
		return "", ""
	}
	for _, binding := range g.staticImports[pkgName] {
		if binding.Kind != staticImportBindingSelector {
			continue
		}
		if strings.TrimSpace(binding.LocalName) != strings.TrimSpace(localName) {
			continue
		}
		source := strings.TrimSpace(binding.SourceName)
		if source == "" {
			continue
		}
		return strings.TrimSpace(binding.SourcePackage), source
	}
	return "", ""
}

func (g *generator) methodTargetName(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return "", false
		}
		return t.Name.Name, true
	case *ast.GenericTypeExpression:
		if t == nil || t.Base == nil {
			return "", false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			return base.Name.Name, true
		}
	}
	return "", false
}

func methodDefinitionExpectsSelf(def *ast.FunctionDefinition) bool {
	if def == nil {
		return false
	}
	if def.IsMethodShorthand {
		return true
	}
	if len(def.Params) == 0 {
		return false
	}
	first := def.Params[0]
	if first == nil {
		return false
	}
	if ident, ok := first.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
		return true
	}
	if simple, ok := first.ParamType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
		return simple.Name.Name == "Self"
	}
	return false
}

func (g *generator) fillMethodInfo(info *functionInfo, mapper *TypeMapper, target ast.TypeExpression, expectsSelf bool) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	g.invalidateFunctionDerivedInfo(info)
	def := info.Definition
	bindings := cloneTypeBindings(info.TypeBindings)
	concreteTarget := target
	if len(bindings) > 0 {
		concreteTarget = substituteTypeParams(concreteTarget, bindings)
	}
	concreteTarget = normalizeTypeExprForPackage(g, info.Package, concreteTarget)
	params := make([]paramInfo, 0, len(def.Params)+1)
	supported := true
	paramIndex := 0
	if expectsSelf && def.IsMethodShorthand {
		selfGo, ok := g.mapMethodType(mapper, concreteTarget, concreteTarget)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      "self",
			GoName:    safeParamName("self", paramIndex),
			GoType:    selfGo,
			TypeExpr:  concreteTarget,
			Supported: ok,
		})
		paramIndex++
	}
	for _, param := range def.Params {
		if param == nil {
			supported = false
			continue
		}
		name := fmt.Sprintf("arg%d", paramIndex)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			supported = false
		}
		paramType := param.ParamType
		if paramType == nil && strings.EqualFold(name, "self") {
			paramType = concreteTarget
		}
		paramType = resolveSelfTypeExpr(paramType, concreteTarget)
		if len(bindings) > 0 {
			paramType = substituteTypeParams(paramType, bindings)
		}
		paramType = normalizeTypeExprForPackage(g, info.Package, paramType)
		goType, ok := g.mapMethodType(mapper, paramType, concreteTarget)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    safeParamName(name, paramIndex),
			GoType:    goType,
			TypeExpr:  paramType,
			Supported: ok,
		})
		paramIndex++
	}
	retExpr := resolveSelfTypeExpr(def.ReturnType, concreteTarget)
	if len(bindings) > 0 {
		retExpr = substituteTypeParams(retExpr, bindings)
	}
	retExpr = normalizeTypeExprForPackage(g, info.Package, retExpr)
	retType := ""
	ok := false
	if forcedType, forced := g.staticMethodNominalStructReturnType(info.Package, concreteTarget, expectsSelf, retExpr); forced {
		retType = forcedType
		ok = true
	}
	if !ok {
		retType, ok = g.mapMethodType(mapper, retExpr, concreteTarget)
	}
	if !ok || retType == "" {
		supported = false
	}
	info.Params = params
	info.ReturnType = retType
	info.SupportedTypes = supported
	info.Arity = len(params)
	if !supported {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
	}
}

func (g *generator) staticMethodNominalStructReturnType(pkgName string, target ast.TypeExpression, expectsSelf bool, retExpr ast.TypeExpression) (string, bool) {
	if g == nil || expectsSelf || retExpr == nil {
		return "", false
	}
	targetName, ok := g.methodTargetName(target)
	if !ok || targetName == "" {
		return "", false
	}
	returnBase, ok := typeExprBaseName(retExpr)
	if !ok || returnBase == "" {
		return "", false
	}
	if returnBase != targetName {
		return "", false
	}
	// Skip for types that have builtin Go mappings (String→string, Bool→bool, etc.)
	// — their return types should use the builtin mapping, not the struct pointer.
	if isBuiltinMappedType(targetName) {
		return "", false
	}
	if recovered, ok := g.recoverRepresentableCarrierType(pkgName, retExpr, ""); ok && recovered != "" && recovered != "runtime.Value" && recovered != "any" {
		return recovered, true
	}
	if carrierType, ok := g.nativeStructCarrierTypeForExpr(pkgName, retExpr); ok && carrierType != "" {
		return carrierType, true
	}
	if carrierType, ok := g.nativeStructCarrierTypeForExpr(pkgName, target); ok && carrierType != "" {
		return carrierType, true
	}
	return "", false
}

func isBuiltinMappedType(name string) bool {
	switch name {
	case "bool", "Bool", "String", "string", "char", "Char",
		"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64",
		"isize", "usize", "f32", "f64", "void", "Void":
		return true
	}
	return false
}

func resolveSelfTypeExpr(expr ast.TypeExpression, target ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return expr
	}
	if s, ok := expr.(*ast.SimpleTypeExpression); ok && s != nil && s.Name != nil {
		if s.Name.Name == "Self" {
			return target
		}
	}
	return expr
}

func (g *generator) mapMethodType(mapper *TypeMapper, expr ast.TypeExpression, target ast.TypeExpression) (string, bool) {
	if mapper == nil {
		return "", false
	}
	mappedExpr := resolveSelfTypeExpr(expr, target)
	mapped, ok := mapper.Map(mappedExpr)
	if g == nil {
		return mapped, ok
	}
	return g.recoverRepresentableCarrierType(mapper.packageName, mappedExpr, mapped)
}

func (g *generator) resolveCompileableMethods() {
	ordered := g.sortedMethodInfos()
	pending := make([]*methodInfo, 0, len(ordered))
	for _, method := range ordered {
		if method == nil || method.Info == nil {
			continue
		}
		if !method.Info.SupportedTypes {
			method.Info.Compileable = false
			continue
		}
		if !method.Info.Compileable {
			pending = append(pending, method)
		}
	}
	for {
		progress := false
		for _, method := range pending {
			if method == nil || method.Info == nil {
				continue
			}
			if method.Info.Compileable {
				continue
			}
			if ok := g.bodyCompileable(method.Info, method.Info.ReturnType); ok {
				method.Info.Compileable = true
				method.Info.Reason = ""
				progress = true
			}
		}
		if !progress {
			break
		}
	}
	for _, method := range pending {
		if method == nil || method.Info == nil {
			continue
		}
		if method.Info.Compileable {
			continue
		}
		if method.Info.Reason == "" {
			method.Info.Reason = "unsupported method body"
		}
		method.Info.Compileable = false
	}
	g.touchNativeInterfaceAdapters()
}

func (g *generator) sortedMethodInfos() []*methodInfo {
	if len(g.methodList) == 0 {
		return nil
	}
	infos := make([]*methodInfo, 0, len(g.methodList))
	infos = append(infos, g.methodList...)
	sortMethodInfos(infos)
	return infos
}

func sortMethodInfos(infos []*methodInfo) {
	if len(infos) == 0 {
		return
	}
	sort.Slice(infos, func(i, j int) bool {
		left := infos[i]
		right := infos[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.TargetName != right.TargetName {
			return left.TargetName < right.TargetName
		}
		if left.MethodName != right.MethodName {
			return left.MethodName < right.MethodName
		}
		if left.Info == nil || right.Info == nil {
			return left.Info != nil
		}
		return left.Info.GoName < right.Info.GoName
	})
}

func (g *generator) methodForTypeName(typeName string, methodName string, expectsSelf bool) *methodInfo {
	if g == nil || typeName == "" || methodName == "" {
		return nil
	}
	typeBucket := g.methods[typeName]
	if typeBucket == nil {
		return nil
	}
	entries := typeBucket[methodName]
	if len(entries) != 1 {
		return nil
	}
	method := entries[0]
	if method == nil || method.Info == nil || !method.Info.Compileable {
		return nil
	}
	if method.ExpectsSelf != expectsSelf {
		return nil
	}
	return method
}

func (g *generator) methodForTypeNameInPackage(pkgName string, typeName string, methodName string, expectsSelf bool) *methodInfo {
	if g == nil || strings.TrimSpace(typeName) == "" || strings.TrimSpace(methodName) == "" {
		return nil
	}
	info, ok := g.structInfoForTypeName(pkgName, typeName)
	if !ok || info == nil {
		return nil
	}
	return g.methodForStruct(info, methodName, expectsSelf)
}

func (g *generator) methodForStruct(info *structInfo, methodName string, expectsSelf bool) *methodInfo {
	if g == nil || info == nil || strings.TrimSpace(methodName) == "" {
		return nil
	}
	typeBucket := g.methods[info.Name]
	if len(typeBucket) == 0 {
		return nil
	}
	entries := typeBucket[methodName]
	if len(entries) == 0 {
		return nil
	}
	var found *methodInfo
	for _, method := range entries {
		if method == nil || method.Info == nil || !method.Info.Compileable {
			continue
		}
		if method.ExpectsSelf != expectsSelf {
			continue
		}
		if method.Info.Package != info.Package {
			continue
		}
		if found != nil && found != method {
			return nil
		}
		found = method
	}
	return found
}

func (g *generator) methodForReceiver(goType string, methodName string) *methodInfo {
	if g == nil || goType == "" || methodName == "" {
		return nil
	}
	info := g.structInfoByGoName(goType)
	if info == nil && g.isMonoArrayType(goType) {
		info, _ = g.structInfoByNameUnique("Array")
	}
	if info != nil && info.Name != "" {
		// Look up by struct Able name. Unlike methodForStruct, skip the
		// package check: methods may be defined in a different package
		// (e.g., able.collections.array extends able.kernel.Array) and
		// since we resolved the struct by concrete GoType, there is no
		// ambiguity.
		typeBucket := g.methods[info.Name]
		if len(typeBucket) > 0 {
			entries := typeBucket[methodName]
			var found *methodInfo
			bestScore := 0
			for _, method := range entries {
				if method == nil || method.Info == nil || !method.Info.Compileable {
					continue
				}
				if !method.ExpectsSelf {
					continue
				}
				score := g.receiverMethodMatchScore(method.ReceiverType, goType)
				if method.ReceiverType != "" && score == 0 {
					continue
				}
				if score > bestScore {
					found = method
					bestScore = score
					continue
				}
				if score < bestScore {
					continue
				}
				if found != nil && found != method {
					return nil // ambiguous
				}
				found = method
			}
			if found != nil {
				return found
			}
		}
		return nil
	}
	// For primitive types (bool, int32, string, etc.) search by receiver Go type.
	return g.methodForReceiverGoType(goType, methodName)
}

func (g *generator) hasUncompiledReceiverMethod(goType string, methodName string) bool {
	if g == nil || goType == "" || methodName == "" {
		return false
	}
	info := g.structInfoByGoName(goType)
	if info == nil && g.isMonoArrayType(goType) {
		info, _ = g.structInfoByNameUnique("Array")
	}
	if info != nil && info.Name != "" {
		for _, method := range g.methods[info.Name][methodName] {
			if method == nil || method.Info == nil || method.Info.Compileable || !method.ExpectsSelf {
				continue
			}
			if method.ReceiverType == "" || g.receiverMethodMatchScore(method.ReceiverType, goType) > 0 {
				return true
			}
		}
		return false
	}
	if goType == "runtime.Value" || strings.HasPrefix(goType, "*") {
		return false
	}
	for _, typeBucket := range g.methods {
		for _, method := range typeBucket[methodName] {
			if method == nil || method.Info == nil || method.Info.Compileable || !method.ExpectsSelf {
				continue
			}
			if method.ReceiverType == goType {
				return true
			}
		}
	}
	return false
}

func (g *generator) receiverMethodMatchScore(expectedReceiverType string, actualGoType string) int {
	if g == nil || expectedReceiverType == "" || actualGoType == "" {
		return 0
	}
	if expectedReceiverType == actualGoType {
		return 3
	}
	if g.typeMatches(expectedReceiverType, actualGoType) {
		return 2
	}
	if expectedInfo := g.structInfoByGoName(expectedReceiverType); expectedInfo != nil {
		if actualInfo := g.structInfoByGoName(actualGoType); actualInfo != nil &&
			expectedInfo.Package == actualInfo.Package &&
			expectedInfo.Name != "" &&
			expectedInfo.Name == actualInfo.Name {
			if g.receiverNominalFamilyCompatible(expectedReceiverType, actualGoType) {
				return 1
			}
			return 0
		}
	}
	if g.isArrayStructType(expectedReceiverType) && g.isStaticArrayType(actualGoType) {
		return 1
	}
	return 0
}

func (g *generator) methodForReceiverGoType(goType string, methodName string) *methodInfo {
	if g == nil {
		return nil
	}
	// Only match concrete primitive Go types, not generic carriers.
	if goType == "runtime.Value" || strings.HasPrefix(goType, "*") {
		return nil
	}
	var found *methodInfo
	for _, typeBucket := range g.methods {
		entries := typeBucket[methodName]
		for _, method := range entries {
			if method == nil || method.Info == nil || !method.Info.Compileable {
				continue
			}
			if !method.ExpectsSelf || method.ReceiverType != goType {
				continue
			}
			if found != nil && found != method {
				return nil // ambiguous
			}
			found = method
		}
	}
	if found != nil {
		return found
	}
	// Also search impl method list (impl I for T { ... } methods).
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		if impl.MethodName != methodName {
			continue
		}
		if len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType != goType {
			continue
		}
		m := &methodInfo{
			TargetName:   typeExpressionToString(impl.TargetType),
			TargetType:   impl.TargetType,
			MethodName:   impl.MethodName,
			ExpectsSelf:  true,
			Info:         impl.Info,
			ReceiverType: goType,
		}
		if found != nil {
			return nil // ambiguous
		}
		found = m
	}
	return found
}

func (g *generator) registerableMethod(method *methodInfo) bool {
	if method == nil || method.Info == nil || !method.Info.Compileable {
		return false
	}
	key, ok := g.methodSignatureKey(method)
	if !ok {
		return false
	}
	count := 0
	for _, other := range g.methodList {
		if other == nil || other.Info == nil || !other.Info.Compileable {
			continue
		}
		if other.TargetName != method.TargetName || other.MethodName != method.MethodName || other.ExpectsSelf != method.ExpectsSelf {
			continue
		}
		otherKey, ok := g.methodSignatureKey(other)
		if !ok {
			continue
		}
		if otherKey == key {
			count++
			if count > 1 {
				return false
			}
		}
	}
	return count == 1
}

func (g *generator) methodSignatureKey(method *methodInfo) (string, bool) {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return "", false
	}
	target := method.TargetType
	if target == nil && method.TargetName != "" {
		target = ast.NewSimpleTypeExpression(ast.NewIdentifier(method.TargetName))
	}
	def := method.Info.Definition
	parts := make([]string, 0, len(def.Params)+1)
	if method.ExpectsSelf && def.IsMethodShorthand {
		parts = append(parts, typeExpressionToString(resolveSelfTypeExpr(target, target)))
	}
	for _, param := range def.Params {
		if param == nil {
			parts = append(parts, "<?>")
			continue
		}
		paramType := param.ParamType
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
				paramType = target
			}
		}
		paramType = resolveSelfTypeExpr(paramType, target)
		parts = append(parts, typeExpressionToString(paramType))
	}
	return fmt.Sprintf("%s|%t|%s|%s", method.TargetName, method.ExpectsSelf, typeExpressionToString(resolveSelfTypeExpr(target, target)), strings.Join(parts, ",")), true
}

func methodDefinitionParamTypes(def *ast.FunctionDefinition, target ast.TypeExpression, expectsSelf bool) []ast.TypeExpression {
	if def == nil {
		return nil
	}
	params := make([]ast.TypeExpression, 0, len(def.Params)+1)
	if expectsSelf && def.IsMethodShorthand {
		params = append(params, resolveSelfTypeExpr(target, target))
	}
	for _, param := range def.Params {
		if param == nil {
			params = append(params, nil)
			continue
		}
		paramType := param.ParamType
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
				paramType = target
			}
		}
		paramType = resolveSelfTypeExpr(paramType, target)
		params = append(params, paramType)
	}
	return params
}
