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
	targetName, ok := g.methodTargetName(def.TargetType)
	if !ok || targetName == "" {
		return
	}
	info := g.structs[targetName]
	if info == nil || !info.Supported {
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
		g.fillMethodInfo(info, mapper, def.TargetType, expectsSelf)
		method := &methodInfo{
			TargetName:  targetName,
			TargetType:  def.TargetType,
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
	if ident, ok := first.Name.(*ast.Identifier); ok && ident != nil {
		return strings.EqualFold(ident.Name, "self")
	}
	return false
}

func (g *generator) fillMethodInfo(info *functionInfo, mapper *TypeMapper, target ast.TypeExpression, expectsSelf bool) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params)+1)
	supported := true
	paramIndex := 0
	if expectsSelf && def.IsMethodShorthand {
		selfGo, ok := g.mapMethodType(mapper, target, target)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      "self",
			GoName:    safeParamName("self", paramIndex),
			GoType:    selfGo,
			TypeExpr:  target,
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
			paramType = target
		}
		paramType = resolveSelfTypeExpr(paramType, target)
		goType, ok := g.mapMethodType(mapper, paramType, target)
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
	retExpr := resolveSelfTypeExpr(def.ReturnType, target)
	retType, ok := g.mapMethodType(mapper, retExpr, target)
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
	return mapper.Map(mappedExpr)
}

func (g *generator) resolveCompileableMethods() {
	for _, method := range g.methodList {
		if method == nil || method.Info == nil {
			continue
		}
		if !method.Info.SupportedTypes {
			method.Info.Compileable = false
			continue
		}
		if ok := g.bodyCompileable(method.Info, method.Info.ReturnType); ok {
			method.Info.Compileable = true
			method.Info.Reason = ""
			continue
		}
		if method.Info.Reason == "" {
			method.Info.Reason = "unsupported method body"
		}
		method.Info.Compileable = false
	}
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

func (g *generator) methodForReceiver(goType string, methodName string) *methodInfo {
	if g == nil || goType == "" || methodName == "" {
		return nil
	}
	info := g.structInfoByGoName(goType)
	if info == nil || info.Name == "" {
		return nil
	}
	method := g.methodForTypeName(info.Name, methodName, true)
	if method == nil {
		return nil
	}
	if method.ReceiverType != "" && method.ReceiverType != goType {
		return nil
	}
	return method
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
