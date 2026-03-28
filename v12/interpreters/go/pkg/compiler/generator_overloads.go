package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type methodOverloadGroup struct {
	TargetName  string
	MethodName  string
	ExpectsSelf bool
	Entries     []*methodInfo
	MinArity    int
}

type publicPackageMethodCallableGroup struct {
	PackageName string
	MethodName  string
	Entries     []*methodInfo
	MinArity    int
}

func (g *generator) allFunctionInfos() []*functionInfo {
	if g == nil {
		return nil
	}
	total := 0
	for _, pkgFuncs := range g.functions {
		total += len(pkgFuncs)
	}
	for _, pkgOverloads := range g.overloads {
		for _, info := range pkgOverloads {
			if info != nil {
				total += len(info.Entries)
			}
		}
	}
	for _, impl := range g.implMethodList {
		if impl != nil && impl.Info != nil {
			total++
		}
	}
	total += len(g.specializedFunctions)
	all := make([]*functionInfo, 0, total)
	functionPackages := sortedMapKeys(g.functions)
	for _, pkgName := range functionPackages {
		pkgFuncs := g.functions[pkgName]
		for _, name := range sortedMapKeys(pkgFuncs) {
			info := pkgFuncs[name]
			if info != nil {
				all = append(all, info)
			}
		}
	}
	overloadPackages := sortedMapKeys(g.overloads)
	for _, pkgName := range overloadPackages {
		pkgOverloads := g.overloads[pkgName]
		for _, name := range sortedMapKeys(pkgOverloads) {
			overload := pkgOverloads[name]
			if overload == nil {
				continue
			}
			for _, entry := range overload.Entries {
				if entry != nil {
					all = append(all, entry)
				}
			}
		}
	}
	for _, impl := range g.implMethodList {
		if impl != nil && impl.Info != nil {
			all = append(all, impl.Info)
		}
	}
	all = append(all, g.specializedFunctions...)
	return all
}

func (g *generator) sortedFunctionInfos() []*functionInfo {
	all := g.allFunctionInfos()
	if len(all) == 0 {
		return nil
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i] == nil || all[j] == nil {
			return false
		}
		if all[i].Package != all[j].Package {
			return all[i].Package < all[j].Package
		}
		if all[i].GoName == all[j].GoName {
			return all[i].Name < all[j].Name
		}
		return all[i].GoName < all[j].GoName
	})
	return all
}

func sortedMapKeys[T any](m map[string]T) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *generator) sortedCallableNames(pkgName string) []string {
	if g == nil {
		return nil
	}
	funcs := g.functions[pkgName]
	overloads := g.overloads[pkgName]
	names := make([]string, 0, len(funcs)+len(overloads))
	for name := range funcs {
		names = append(names, name)
	}
	for name := range overloads {
		if _, exists := funcs[name]; !exists {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (g *generator) hasFunctions() bool {
	if g == nil {
		return false
	}
	if len(g.methodList) > 0 {
		return true
	}
	if len(g.implMethodList) > 0 {
		return true
	}
	for _, pkgFuncs := range g.functions {
		if len(pkgFuncs) > 0 {
			return true
		}
	}
	for _, pkgOverloads := range g.overloads {
		if len(pkgOverloads) > 0 {
			return true
		}
	}
	return false
}

func (g *generator) functionsForPackage(pkgName string) map[string]*functionInfo {
	if g == nil {
		return nil
	}
	return g.functions[pkgName]
}

func (g *generator) functionsForCompileContext(info *functionInfo) map[string]*functionInfo {
	if g == nil || info == nil {
		return g.functionsForPackage("")
	}
	base := g.functionsForPackage(info.Package)
	name := strings.TrimSpace(info.Name)
	if name == "" {
		return base
	}
	if existing, ok := base[name]; ok && existing == info {
		return base
	}
	if len(base) == 0 {
		return map[string]*functionInfo{name: info}
	}
	overlay := make(map[string]*functionInfo, len(base)+1)
	for key, value := range base {
		overlay[key] = value
	}
	overlay[name] = info
	return overlay
}

func (g *generator) overloadsForPackage(pkgName string) map[string]*overloadInfo {
	if g == nil {
		return nil
	}
	return g.overloads[pkgName]
}

func minArgsForDefinition(def *ast.FunctionDefinition) int {
	if def == nil {
		return 0
	}
	paramCount := len(def.Params)
	if paramCount == 0 {
		return 0
	}
	if isNullableParam(def.Params[paramCount-1]) {
		return paramCount - 1
	}
	return paramCount
}

func minArgsForMethod(method *methodInfo) int {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return 0
	}
	def := method.Info.Definition
	paramCount := len(def.Params)
	if method.ExpectsSelf && def.IsMethodShorthand {
		paramCount++
	}
	if paramCount == 0 {
		return 0
	}
	if len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1]) {
		return paramCount - 1
	}
	return paramCount
}

func isNullableParam(param *ast.FunctionParameter) bool {
	if param == nil || param.ParamType == nil {
		return false
	}
	_, ok := param.ParamType.(*ast.NullableTypeExpression)
	return ok
}

func typeExprUsesGeneric(typeExpr ast.TypeExpression, generics map[string]struct{}) bool {
	if typeExpr == nil || len(generics) == 0 {
		return false
	}
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return false
		}
		_, ok := generics[t.Name.Name]
		return ok
	case *ast.GenericTypeExpression:
		if typeExprUsesGeneric(t.Base, generics) {
			return true
		}
		for _, arg := range t.Arguments {
			if typeExprUsesGeneric(arg, generics) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if typeExprUsesGeneric(t.ReturnType, generics) {
			return true
		}
		for _, param := range t.ParamTypes {
			if typeExprUsesGeneric(param, generics) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return typeExprUsesGeneric(t.InnerType, generics)
	case *ast.ResultTypeExpression:
		return typeExprUsesGeneric(t.InnerType, generics)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if typeExprUsesGeneric(member, generics) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func parameterSpecificity(typeExpr ast.TypeExpression, generics map[string]struct{}) int {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return 0
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0
		}
		if _, ok := generics[t.Name.Name]; ok {
			return 1
		}
		return 3
	case *ast.NullableTypeExpression:
		return 1 + parameterSpecificity(t.InnerType, generics)
	case *ast.GenericTypeExpression:
		score := 2 + parameterSpecificity(t.Base, generics)
		for _, arg := range t.Arguments {
			score += parameterSpecificity(arg, generics)
		}
		return score
	case *ast.FunctionTypeExpression, *ast.UnionTypeExpression:
		return 2
	default:
		return 1
	}
}

func genericNameSet(params []*ast.GenericParameter) map[string]struct{} {
	names := make(map[string]struct{}, len(params))
	for _, gp := range params {
		if gp == nil || gp.Name == nil {
			continue
		}
		names[gp.Name.Name] = struct{}{}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func combinedGenericNameSet(params ...[]*ast.GenericParameter) map[string]struct{} {
	names := make(map[string]struct{})
	for _, list := range params {
		for _, gp := range list {
			if gp == nil || gp.Name == nil {
				continue
			}
			names[gp.Name.Name] = struct{}{}
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func (g *generator) overloadWrapperName(pkgName string, name string) string {
	return fmt.Sprintf("__able_overload_%s", g.overloadBase(pkgName, name))
}

func (g *generator) overloadCallName(pkgName string, name string) string {
	return fmt.Sprintf("__able_call_overload_%s", g.overloadBase(pkgName, name))
}

func (g *generator) overloadValueName(pkgName string, name string) string {
	return fmt.Sprintf("__able_overload_value_%s", g.overloadBase(pkgName, name))
}

func (g *generator) overloadBase(pkgName string, name string) string {
	safeName := sanitizeIdent(name)
	if pkgName == "" {
		return safeName
	}
	return fmt.Sprintf("%s_%s", sanitizeIdent(pkgName), safeName)
}

func (g *generator) methodOverloadWrapperName(targetName string, methodName string, expectsSelf bool) string {
	return fmt.Sprintf("__able_method_overload_%s", g.methodOverloadBase(targetName, methodName, expectsSelf))
}

func (g *generator) methodOverloadValueName(targetName string, methodName string, expectsSelf bool) string {
	return fmt.Sprintf("__able_method_overload_value_%s", g.methodOverloadBase(targetName, methodName, expectsSelf))
}

func (g *generator) methodOverloadBase(targetName string, methodName string, expectsSelf bool) string {
	base := fmt.Sprintf("%s_%s", sanitizeIdent(targetName), sanitizeIdent(methodName))
	if expectsSelf {
		return base + "_self"
	}
	return base + "_static"
}

func (g *generator) publicPackageMethodCallableWrapperName(pkgName string, methodName string) string {
	return fmt.Sprintf("__able_public_package_method_%s", g.publicPackageMethodCallableBase(pkgName, methodName))
}

func (g *generator) publicPackageMethodCallableBase(pkgName string, methodName string) string {
	base := fmt.Sprintf("%s_%s", sanitizeIdent(pkgName), sanitizeIdent(methodName))
	return strings.Trim(base, "_")
}

func (g *generator) methodOverloadKey(targetName string, methodName string, expectsSelf bool) string {
	return fmt.Sprintf("%s|%s|%t", targetName, methodName, expectsSelf)
}

func (g *generator) methodOverloadGroups() []*methodOverloadGroup {
	if g == nil || len(g.methodList) == 0 {
		return nil
	}
	totalCounts := make(map[string]int)
	invalid := make(map[string]bool)
	groups := make(map[string]*methodOverloadGroup)
	for _, method := range g.methodList {
		if method == nil || method.Info == nil {
			continue
		}
		key := g.methodOverloadKey(method.TargetName, method.MethodName, method.ExpectsSelf)
		totalCounts[key]++
		if !g.registerableMethod(method) {
			invalid[key] = true
			continue
		}
		group := groups[key]
		if group == nil {
			group = &methodOverloadGroup{
				TargetName:  method.TargetName,
				MethodName:  method.MethodName,
				ExpectsSelf: method.ExpectsSelf,
				MinArity:    -1,
			}
			groups[key] = group
		}
		group.Entries = append(group.Entries, method)
		if minArgs := minArgsForMethod(method); minArgs >= 0 {
			if group.MinArity < 0 || minArgs < group.MinArity {
				group.MinArity = minArgs
			}
		}
	}
	result := make([]*methodOverloadGroup, 0, len(groups))
	for key, group := range groups {
		if invalid[key] {
			continue
		}
		if totalCounts[key] <= 1 {
			continue
		}
		if len(group.Entries) != totalCounts[key] {
			continue
		}
		sort.Slice(group.Entries, func(i, j int) bool {
			left := group.Entries[i]
			right := group.Entries[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.Info == nil || right.Info == nil {
				return left.Info != nil
			}
			return left.Info.GoName < right.Info.GoName
		})
		result = append(result, group)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.TargetName != right.TargetName {
			return left.TargetName < right.TargetName
		}
		if left.MethodName != right.MethodName {
			return left.MethodName < right.MethodName
		}
		if left.ExpectsSelf != right.ExpectsSelf {
			return left.ExpectsSelf
		}
		return false
	})
	return result
}

func (g *generator) publicPackageMethodCallableGroups(pkgName string) []*publicPackageMethodCallableGroup {
	if g == nil || strings.TrimSpace(pkgName) == "" || len(g.methodList) == 0 {
		return nil
	}
	blockedNames := make(map[string]struct{})
	for _, name := range g.sortedCallableNames(pkgName) {
		blockedNames[strings.TrimSpace(name)] = struct{}{}
	}
	for _, name := range g.sortedPublicStructNames(pkgName) {
		blockedNames[strings.TrimSpace(name)] = struct{}{}
	}
	for _, name := range g.sortedPublicInterfaceNames(pkgName) {
		blockedNames[strings.TrimSpace(name)] = struct{}{}
	}
	for _, name := range g.sortedPublicUnionNames(pkgName) {
		blockedNames[strings.TrimSpace(name)] = struct{}{}
	}
	for _, name := range g.sortedPublicImplNamespaceNames(pkgName) {
		blockedNames[strings.TrimSpace(name)] = struct{}{}
	}

	groups := make(map[string]*publicPackageMethodCallableGroup)
	for _, method := range g.methodList {
		if method == nil || method.Info == nil || method.Info.Definition == nil {
			continue
		}
		if method.Info.Package != pkgName {
			continue
		}
		if method.Info.Definition.IsPrivate {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		name := strings.TrimSpace(method.MethodName)
		if name == "" {
			continue
		}
		if _, blocked := blockedNames[name]; blocked {
			continue
		}
		group := groups[name]
		if group == nil {
			group = &publicPackageMethodCallableGroup{
				PackageName: pkgName,
				MethodName:  name,
				MinArity:    -1,
			}
			groups[name] = group
		}
		group.Entries = append(group.Entries, method)
		if minArgs := minArgsForMethod(method); minArgs >= 0 {
			if group.MinArity < 0 || minArgs < group.MinArity {
				group.MinArity = minArgs
			}
		}
	}
	if len(groups) == 0 {
		return nil
	}
	result := make([]*publicPackageMethodCallableGroup, 0, len(groups))
	for _, group := range groups {
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		sort.Slice(group.Entries, func(i, j int) bool {
			left := group.Entries[i]
			right := group.Entries[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.Info == nil || right.Info == nil {
				return left.Info != nil
			}
			return left.Info.GoName < right.Info.GoName
		})
		result = append(result, group)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.PackageName != right.PackageName {
			return left.PackageName < right.PackageName
		}
		return left.MethodName < right.MethodName
	})
	return result
}

func (g *generator) sortedPublicMethodCallableNames(pkgName string) []string {
	groups := g.publicPackageMethodCallableGroups(pkgName)
	if len(groups) == 0 {
		return nil
	}
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		if group == nil || strings.TrimSpace(group.MethodName) == "" {
			continue
		}
		names = append(names, group.MethodName)
	}
	sort.Strings(names)
	return names
}

func (g *generator) compileOverloadCall(ctx *compileContext, call *ast.FunctionCall, expected string, name string, callNode string) ([]string, string, string, bool) {
	return g.compileResolvedOverloadCall(ctx, call, expected, ctx.packageName, name, callNode)
}

func (g *generator) compileResolvedOverloadCall(ctx *compileContext, call *ast.FunctionCall, expected string, pkgName string, name string, callNode string) ([]string, string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return nil, "", "", false
	}
	if !g.isStaticallyKnownExpectedType(expected) {
		ctx.setReason("call return type mismatch")
		return nil, "", "", false
	}
	var lines []string
	args := make([]string, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		argLines, expr, goType, ok := g.compileExprLines(ctx, arg, "")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		argConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return nil, "", "", false
		}
		lines = append(lines, argConvLines...)
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}
	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}
	callExpr := fmt.Sprintf("%s(%s, %s)", g.overloadCallName(pkgName, name), argList, callNode)
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		return lines, "struct{}{}", "struct{}", true
	}
	resultTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, resultTemp, "runtime.Value", true
}
