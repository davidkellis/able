package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

var alwaysAvailableNamedCallables = map[string]struct{}{
	"print":                 {},
	"future_yield":          {},
	"future_cancelled":      {},
	"future_flush":          {},
	"future_pending_tasks":  {},
	"__able_await_default":  {},
	"__able_await_sleep_ms": {},
}

func (g *generator) mayResolveStaticNamedCall(ctx *compileContext, name string) bool {
	if g == nil || ctx == nil {
		return false
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false
	}
	if _, ok := alwaysAvailableNamedCallables[trimmed]; ok {
		return true
	}
	if _, ok := g.runtimeHelperImpl(trimmed); ok {
		return true
	}
	if strings.HasPrefix(trimmed, "__able_") {
		return true
	}
	if currentPkg, targetPkg, targetName, ok := g.resolveQualifiedStaticCallable(ctx, trimmed); ok {
		return g.callableAccessibleFromPackage(currentPkg, targetPkg, targetName)
	}
	pkgName := strings.TrimSpace(ctx.packageName)
	if pkgName == "" {
		return false
	}
	set := g.staticCallableNameSet(pkgName)
	if len(set) == 0 {
		return false
	}
	_, ok := set[trimmed]
	return ok
}

func (g *generator) mayResolveStaticUFCSCall(ctx *compileContext, call *ast.FunctionCall, name string) bool {
	if g == nil || ctx == nil || call == nil {
		return false
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(call.Arguments) == 0 {
		return false
	}
	if !g.hasCompileableInstanceMethodNamed(trimmed) {
		return false
	}
	receiverType := g.ufcsReceiverGoType(ctx, call.Arguments[0])
	if receiverType == "" || receiverType == "runtime.Value" {
		return true
	}
	return g.methodForReceiver(receiverType, trimmed) != nil
}

func (g *generator) hasCompileableInstanceMethodNamed(name string) bool {
	if g == nil {
		return false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, typeBucket := range g.methods {
		if len(typeBucket) == 0 {
			continue
		}
		entries := typeBucket[name]
		for _, entry := range entries {
			if entry == nil || !entry.ExpectsSelf || entry.Info == nil || !entry.Info.Compileable {
				continue
			}
			return true
		}
	}
	return false
}

func (g *generator) ufcsReceiverGoType(ctx *compileContext, receiver ast.Expression) string {
	if g == nil || ctx == nil || receiver == nil {
		return ""
	}
	ident, ok := receiver.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return ""
	}
	binding, ok := ctx.lookup(ident.Name)
	if !ok {
		return ""
	}
	goType := strings.TrimSpace(binding.GoType)
	if goType != "" && goType != "runtime.Value" {
		return goType
	}
	if binding.TypeExpr == nil {
		return ""
	}
	mapped, ok := g.lowerCarrierType(ctx, binding.TypeExpr)
	if !ok {
		return ""
	}
	mapped = strings.TrimSpace(mapped)
	if mapped == "" || mapped == "runtime.Value" {
		return ""
	}
	return mapped
}

func (g *generator) resolveQualifiedStaticCallable(ctx *compileContext, name string) (currentPkg string, targetPkg string, callable string, ok bool) {
	if g == nil || ctx == nil {
		return "", "", "", false
	}
	currentPkg = strings.TrimSpace(ctx.packageName)
	if currentPkg == "" {
		return "", "", "", false
	}
	head, tail, qualified := splitQualifiedCallable(name)
	if !qualified {
		return "", "", "", false
	}
	if g.callableExists(head, tail) {
		return currentPkg, head, tail, true
	}
	for _, binding := range g.staticImportsForPackage(currentPkg) {
		if binding.Kind != staticImportBindingPackage {
			continue
		}
		if strings.TrimSpace(binding.LocalName) != head {
			continue
		}
		if strings.TrimSpace(binding.SourcePackage) == "" {
			continue
		}
		if g.callableExists(binding.SourcePackage, tail) {
			return currentPkg, binding.SourcePackage, tail, true
		}
	}
	return "", "", "", false
}

func splitQualifiedCallable(name string) (pkg string, callable string, ok bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", "", false
	}
	dot := strings.LastIndex(trimmed, ".")
	if dot <= 0 || dot >= len(trimmed)-1 {
		return "", "", false
	}
	pkg = strings.TrimSpace(trimmed[:dot])
	callable = strings.TrimSpace(trimmed[dot+1:])
	if pkg == "" || callable == "" {
		return "", "", false
	}
	return pkg, callable, true
}

func (g *generator) callableExists(pkgName string, name string) bool {
	if g == nil {
		return false
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return false
	}
	if pkgFuncs := g.functions[pkgName]; pkgFuncs != nil {
		if info := pkgFuncs[name]; info != nil {
			return true
		}
	}
	if pkgOverloads := g.overloads[pkgName]; pkgOverloads != nil {
		if overload := pkgOverloads[name]; overload != nil {
			return true
		}
	}
	if g.externCallableExists(pkgName, name) {
		return true
	}
	return false
}

func (g *generator) callableAccessibleFromPackage(currentPkg string, targetPkg string, name string) bool {
	if g == nil {
		return false
	}
	currentPkg = strings.TrimSpace(currentPkg)
	targetPkg = strings.TrimSpace(targetPkg)
	name = strings.TrimSpace(name)
	if currentPkg == "" || targetPkg == "" || name == "" {
		return false
	}
	return g.callableExists(targetPkg, name)
}

func (g *generator) resolveStaticCallable(ctx *compileContext, name string) (*functionInfo, *overloadInfo, bool) {
	if g == nil || ctx == nil {
		return nil, nil, false
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, nil, false
	}
	if info, ok := ctx.functions[trimmed]; ok && info != nil {
		return info, nil, true
	}
	if overload, ok := ctx.overloads[trimmed]; ok && overload != nil {
		return nil, overload, true
	}
	if _, targetPkg, targetName, ok := g.resolveQualifiedStaticCallable(ctx, trimmed); ok {
		if info := g.functions[targetPkg][targetName]; info != nil {
			return info, nil, true
		}
		if overload := g.overloads[targetPkg][targetName]; overload != nil {
			return nil, overload, true
		}
	}

	var resolvedInfo *functionInfo
	var resolvedOverload *overloadInfo
	recordResolution := func(info *functionInfo, overload *overloadInfo) bool {
		if info != nil {
			if resolvedOverload != nil {
				return false
			}
			if resolvedInfo != nil && resolvedInfo != info {
				return false
			}
			resolvedInfo = info
			return true
		}
		if overload != nil {
			if resolvedInfo != nil {
				return false
			}
			if resolvedOverload != nil && resolvedOverload != overload {
				return false
			}
			resolvedOverload = overload
			return true
		}
		return true
	}

	for _, binding := range g.staticImportsForPackage(strings.TrimSpace(ctx.packageName)) {
		targetPkg := strings.TrimSpace(binding.SourcePackage)
		targetName := ""
		switch binding.Kind {
		case staticImportBindingSelector:
			if strings.TrimSpace(binding.LocalName) != trimmed {
				continue
			}
			targetName = strings.TrimSpace(binding.SourceName)
		case staticImportBindingWildcard:
			if !g.callableExists(targetPkg, trimmed) {
				continue
			}
			targetName = trimmed
		default:
			continue
		}
		if targetPkg == "" || targetName == "" {
			continue
		}
		if info := g.functions[targetPkg][targetName]; info != nil {
			if !recordResolution(info, nil) {
				return nil, nil, false
			}
			continue
		}
		if overload := g.overloads[targetPkg][targetName]; overload != nil {
			if !recordResolution(nil, overload) {
				return nil, nil, false
			}
		}
	}

	if resolvedInfo != nil || resolvedOverload != nil {
		return resolvedInfo, resolvedOverload, true
	}
	return nil, nil, false
}

func (g *generator) staticCallableNameSet(pkgName string) map[string]struct{} {
	if g == nil {
		return nil
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" {
		return nil
	}
	if g.staticCallableNames == nil {
		g.staticCallableNames = make(map[string]map[string]struct{})
	}
	if cached, ok := g.staticCallableNames[pkgName]; ok {
		return cached
	}

	set := make(map[string]struct{})
	for _, name := range g.sortedCallableNames(pkgName) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	for _, name := range g.externCallableNames(pkgName) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}

	for _, binding := range g.staticImportsForPackage(pkgName) {
		sourcePkg := strings.TrimSpace(binding.SourcePackage)
		if sourcePkg == "" {
			continue
		}
		switch binding.Kind {
		case staticImportBindingSelector:
			localName := strings.TrimSpace(binding.LocalName)
			sourceName := strings.TrimSpace(binding.SourceName)
			if localName == "" || sourceName == "" {
				continue
			}
			set[localName] = struct{}{}
		case staticImportBindingWildcard:
			for _, exported := range g.sortedPublicCallableNames(sourcePkg) {
				trimmed := strings.TrimSpace(exported)
				if trimmed == "" {
					continue
				}
				set[trimmed] = struct{}{}
			}
		}
	}

	g.staticCallableNames[pkgName] = set
	return set
}
