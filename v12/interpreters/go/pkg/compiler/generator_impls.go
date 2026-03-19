package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) collectImplDefinition(def *ast.ImplementationDefinition, mapper *TypeMapper, pkgName string) {
	if def == nil || mapper == nil {
		return
	}
	if def.InterfaceName == nil || def.InterfaceName.Name == "" {
		return
	}
	if def.TargetType == nil {
		return
	}
	targetType := g.expandTypeAliasForPackage(pkgName, def.TargetType)
	if targetType == nil {
		targetType = def.TargetType
	}
	if g.implDefinitions == nil {
		g.implDefinitions = make([]*implDefinitionInfo, 0, 4)
	}
	g.implDefinitions = append(g.implDefinitions, &implDefinitionInfo{
		Definition: def,
		Package:    pkgName,
	})
	if len(def.Definitions) == 0 {
		return
	}
	if g.implMethodList == nil {
		g.implMethodList = make([]*implMethodInfo, 0, len(def.Definitions))
	}
	ifaceName := def.InterfaceName.Name
	targetDesc := typeExpressionToString(targetType)
	implName := ""
	if def.ImplName != nil {
		implName = def.ImplName.Name
	}
	var ifaceGenerics []*ast.GenericParameter
	if iface := g.interfaces[ifaceName]; iface != nil {
		ifaceGenerics = iface.GenericParams
	}
	interfaceBindings := implInterfaceTypeBindings(ifaceGenerics, def.InterfaceArgs)
	for idx, fn := range def.Definitions {
		if fn == nil || fn.ID == nil || fn.ID.Name == "" {
			continue
		}
		methodName := fn.ID.Name
		info := &functionInfo{
			Name:        fmt.Sprintf("impl %s for %s.%s", ifaceName, targetDesc, methodName),
			Package:     pkgName,
			GoName:      g.mangler.unique(fmt.Sprintf("impl_%s_%s_%d", sanitizeIdent(ifaceName), sanitizeIdent(methodName), idx)),
			Definition:  fn,
			HasOriginal: false,
		}
		g.fillImplMethodInfo(info, mapper, targetType, interfaceBindings)
		implInfo := &implMethodInfo{
			InterfaceName:     ifaceName,
			InterfaceArgs:     def.InterfaceArgs,
			InterfaceGenerics: ifaceGenerics,
			TargetType:        targetType,
			ImplName:          implName,
			ImplGenerics:      def.GenericParams,
			WhereClause:       def.WhereClause,
			MethodName:        methodName,
			Info:              info,
			ImplDefinition:    def,
		}
		g.implMethodList = append(g.implMethodList, implInfo)
		if g.implMethodByInfo != nil {
			g.implMethodByInfo[info] = implInfo
		}
	}
}

func (g *generator) collectDefaultImplMethods() {
	if g == nil || len(g.implDefinitions) == 0 {
		return
	}
	for _, entry := range g.implDefinitions {
		if entry == nil || entry.Definition == nil {
			continue
		}
		def := entry.Definition
		if def.InterfaceName == nil || def.InterfaceName.Name == "" || def.TargetType == nil {
			continue
		}
		ifaceName := def.InterfaceName.Name
		iface := g.interfaces[ifaceName]
		if iface == nil || len(iface.Signatures) == 0 {
			continue
		}
		explicit := make(map[string]struct{}, len(def.Definitions))
		for _, fn := range def.Definitions {
			if fn == nil || fn.ID == nil || fn.ID.Name == "" {
				continue
			}
			explicit[fn.ID.Name] = struct{}{}
		}
		implName := ""
		if def.ImplName != nil {
			implName = def.ImplName.Name
		}
		pkgName := g.interfacePackages[ifaceName]
		if pkgName == "" {
			pkgName = entry.Package
		}
		mapper := NewTypeMapper(g, pkgName)
		targetType := g.expandTypeAliasForPackage(entry.Package, def.TargetType)
		if targetType == nil {
			targetType = def.TargetType
		}
		if g.implMethodList == nil {
			g.implMethodList = make([]*implMethodInfo, 0, len(iface.Signatures))
		}
		for idx, sig := range iface.Signatures {
			if sig == nil || sig.Name == nil || sig.Name.Name == "" || sig.DefaultImpl == nil {
				continue
			}
			if _, ok := explicit[sig.Name.Name]; ok {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			info := &functionInfo{
				Name:        fmt.Sprintf("impl %s for %s.%s", ifaceName, typeExpressionToString(targetType), sig.Name.Name),
				Package:     pkgName,
				GoName:      g.mangler.unique(fmt.Sprintf("impl_%s_%s_default_%d", sanitizeIdent(ifaceName), sanitizeIdent(sig.Name.Name), idx)),
				Definition:  defaultDef,
				HasOriginal: false,
			}
			interfaceBindings := implInterfaceTypeBindings(iface.GenericParams, def.InterfaceArgs)
			g.fillImplMethodInfo(info, mapper, targetType, interfaceBindings)
			implInfo := &implMethodInfo{
				InterfaceName:     ifaceName,
				InterfaceArgs:     def.InterfaceArgs,
				InterfaceGenerics: iface.GenericParams,
				TargetType:        targetType,
				ImplName:          implName,
				IsDefault:         true,
				ImplGenerics:      def.GenericParams,
				WhereClause:       def.WhereClause,
				MethodName:        sig.Name.Name,
				Info:              info,
				ImplDefinition:    def,
			}
			g.implMethodList = append(g.implMethodList, implInfo)
			if g.implMethodByInfo != nil {
				g.implMethodByInfo[info] = implInfo
			}
		}
	}
}

func implInterfaceTypeBindings(params []*ast.GenericParameter, args []ast.TypeExpression) map[string]ast.TypeExpression {
	if len(params) == 0 || len(args) == 0 {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression, len(args))
	for idx, gp := range params {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || idx >= len(args) || args[idx] == nil {
			continue
		}
		bindings[gp.Name.Name] = args[idx]
	}
	if len(bindings) == 0 {
		return nil
	}
	return bindings
}

func (g *generator) fillImplMethodInfo(info *functionInfo, mapper *TypeMapper, target ast.TypeExpression, interfaceBindings map[string]ast.TypeExpression) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	if def.IsMethodShorthand {
		supported = false
	}
	for idx, param := range def.Params {
		if param == nil {
			supported = false
			continue
		}
		name := fmt.Sprintf("arg%d", idx)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			supported = false
		}
		paramType := param.ParamType
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil {
				if ident.Name == "self" || ident.Name == "Self" {
					paramType = target
				}
			}
		}
		paramType = resolveSelfTypeExpr(paramType, target)
		paramType = substituteTypeParams(paramType, interfaceBindings)
		goType, ok := mapper.Map(paramType)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    safeParamName(name, idx),
			GoType:    goType,
			TypeExpr:  paramType,
			Supported: ok,
		})
	}
	retExpr := resolveSelfTypeExpr(def.ReturnType, target)
	retExpr = substituteTypeParams(retExpr, interfaceBindings)
	expectsSelf := methodDefinitionExpectsSelf(def)
	retType := ""
	ok := false
	if forcedType, forced := g.staticMethodNominalStructReturnType(info.Package, target, expectsSelf, retExpr); forced {
		retType = forcedType
		ok = true
	}
	if !ok {
		retType, ok = mapper.Map(retExpr)
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

func (g *generator) sortedImplMethodInfos() []*implMethodInfo {
	if g == nil || len(g.implMethodList) == 0 {
		return nil
	}
	infos := make([]*implMethodInfo, 0, len(g.implMethodList))
	infos = append(infos, g.implMethodList...)
	sort.Slice(infos, func(i, j int) bool {
		left := infos[i]
		right := infos[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.InterfaceName != right.InterfaceName {
			return left.InterfaceName < right.InterfaceName
		}
		if left.ImplName != right.ImplName {
			return left.ImplName < right.ImplName
		}
		if left.MethodName != right.MethodName {
			return left.MethodName < right.MethodName
		}
		if left.Info == nil || right.Info == nil {
			return left.Info != nil
		}
		return left.Info.GoName < right.Info.GoName
	})
	return infos
}

func (g *generator) implSiblingsForDefault(target *implMethodInfo) map[string]implSiblingInfo {
	if g == nil || target == nil {
		return nil
	}
	targetTypeStr := typeExpressionToString(target.TargetType)
	targetTypeName, targetTypeNameOK := g.methodTargetName(target.TargetType)
	siblings := make(map[string]implSiblingInfo)
	ambiguous := make(map[string]struct{})
	allowedInterfaces := g.interfaceFamilyNames(target.InterfaceName)
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		if _, ok := allowedInterfaces[impl.InterfaceName]; !ok {
			continue
		}
		// For named impls, match by impl name; for unnamed, match by target type.
		if target.ImplName != "" {
			if impl.ImplName != target.ImplName {
				continue
			}
		}
		if implTypeName, ok := g.methodTargetName(impl.TargetType); targetTypeNameOK && ok {
			if implTypeName != targetTypeName {
				continue
			}
		} else if typeExpressionToString(impl.TargetType) != targetTypeStr {
			continue
		}
		if impl.MethodName == target.MethodName {
			continue
		}
		candidate := implSiblingInfo{
			GoName: impl.Info.GoName,
			Arity:  impl.Info.Arity,
			Info:   impl.Info,
		}
		if existing, ok := siblings[impl.MethodName]; ok && existing.Info != candidate.Info {
			delete(siblings, impl.MethodName)
			ambiguous[impl.MethodName] = struct{}{}
			continue
		}
		if _, blocked := ambiguous[impl.MethodName]; blocked {
			continue
		}
		siblings[impl.MethodName] = candidate
	}
	if len(siblings) == 0 {
		return nil
	}
	return siblings
}

func (g *generator) interfaceFamilyNames(interfaceName string) map[string]struct{} {
	if g == nil || interfaceName == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var visit func(string)
	visit = func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		iface := g.interfaces[name]
		if iface == nil {
			return
		}
		pkgName := g.interfacePackages[name]
		for _, baseExpr := range iface.BaseInterfaces {
			_, baseName, _, _, ok := interfaceExprInfo(g, pkgName, baseExpr)
			if !ok {
				continue
			}
			visit(baseName)
		}
	}
	visit(interfaceName)
	return seen
}
