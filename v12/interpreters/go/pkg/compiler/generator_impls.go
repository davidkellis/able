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
	targetDesc := typeExpressionToString(def.TargetType)
	implName := ""
	if def.ImplName != nil {
		implName = def.ImplName.Name
	}
	var ifaceGenerics []*ast.GenericParameter
	if iface := g.interfaces[ifaceName]; iface != nil {
		ifaceGenerics = iface.GenericParams
	}
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
		g.fillImplMethodInfo(info, mapper, def.TargetType)
		implInfo := &implMethodInfo{
			InterfaceName:     ifaceName,
			InterfaceArgs:     def.InterfaceArgs,
			InterfaceGenerics: ifaceGenerics,
			TargetType:        def.TargetType,
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
		mapper := NewTypeMapper(g.structs, pkgName)
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
				Name:        fmt.Sprintf("impl %s for %s.%s", ifaceName, typeExpressionToString(def.TargetType), sig.Name.Name),
				Package:     pkgName,
				GoName:      g.mangler.unique(fmt.Sprintf("impl_%s_%s_default_%d", sanitizeIdent(ifaceName), sanitizeIdent(sig.Name.Name), idx)),
				Definition:  defaultDef,
				HasOriginal: false,
			}
			g.fillImplMethodInfo(info, mapper, def.TargetType)
			implInfo := &implMethodInfo{
				InterfaceName:     ifaceName,
				InterfaceArgs:     def.InterfaceArgs,
				InterfaceGenerics: iface.GenericParams,
				TargetType:        def.TargetType,
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

func (g *generator) fillImplMethodInfo(info *functionInfo, mapper *TypeMapper, target ast.TypeExpression) {
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
	retType, ok := mapper.Map(retExpr)
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
