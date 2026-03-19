package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type nativeUnionPatternTarget struct {
	GoType          string
	Member          *nativeUnionMember
	InterfaceBranch bool
}

func (g *generator) nativeCarrierImplementsInterface(goType string, interfaceName string) bool {
	if g == nil || goType == "" || interfaceName == "" {
		return false
	}
	if goType == "runtime.ErrorValue" && interfaceName == "Error" {
		return true
	}
	for _, impl := range g.implDefinitions {
		if impl == nil || impl.Definition == nil || impl.Definition.InterfaceName == nil {
			continue
		}
		if impl.Definition.InterfaceName.Name != interfaceName || impl.Definition.TargetType == nil {
			continue
		}
		targetType := g.expandTypeAliasForPackage(impl.Package, impl.Definition.TargetType)
		if targetType == nil {
			targetType = impl.Definition.TargetType
		}
		mapped, ok := g.mapTypeExpressionInPackage(impl.Package, targetType)
		if ok && mapped == goType {
			return true
		}
	}
	return false
}

func (g *generator) nativeUnionInterfacePatternMember(subjectType string, patternType ast.TypeExpression, pkgName string) (*nativeUnionMember, bool) {
	info := g.nativeUnionInfoForGoType(subjectType)
	if info == nil {
		return nil, false
	}
	_ = pkgName
	simple, ok := patternType.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return nil, false
	}
	interfaceName := simple.Name.Name
	if !g.isInterfaceName(interfaceName) && interfaceName != "Error" {
		return nil, false
	}
	var matched *nativeUnionMember
	for _, member := range info.Members {
		if member == nil || !g.nativeCarrierImplementsInterface(member.GoType, interfaceName) {
			continue
		}
		if matched != nil {
			return nil, false
		}
		matched = member
	}
	return matched, matched != nil
}

func (g *generator) resolveNativeUnionTypedPattern(subjectType string, patternType ast.TypeExpression, pkgName string) (nativeUnionPatternTarget, bool) {
	if mapped, ok := g.nativeUnionPatternMemberType(subjectType, patternType, pkgName); ok {
		if mapped == subjectType {
			return nativeUnionPatternTarget{GoType: mapped}, true
		}
		union := g.nativeUnionInfoForGoType(subjectType)
		if union == nil {
			return nativeUnionPatternTarget{}, false
		}
		member, ok := g.nativeUnionMember(union, mapped)
		if !ok {
			return nativeUnionPatternTarget{}, false
		}
		return nativeUnionPatternTarget{GoType: mapped, Member: member}, true
	}
	member, ok := g.nativeUnionInterfacePatternMember(subjectType, patternType, pkgName)
	if !ok || member == nil {
		return nativeUnionPatternTarget{}, false
	}
	return nativeUnionPatternTarget{
		GoType:          member.GoType,
		Member:          member,
		InterfaceBranch: true,
	}, true
}

func (g *generator) resolveNativeUnionTypedPatternInContext(ctx *compileContext, subjectType string, patternType ast.TypeExpression) (nativeUnionPatternTarget, bool) {
	if ctx == nil {
		return g.resolveNativeUnionTypedPattern(subjectType, patternType, "")
	}
	return g.resolveNativeUnionTypedPattern(subjectType, g.typeExprInContext(ctx, patternType), ctx.packageName)
}

func nativeUnionWholeValueBinding(pattern ast.Pattern) bool {
	ident, ok := pattern.(*ast.Identifier)
	return ok && ident != nil && ident.Name != "" && ident.Name != "_"
}

func (g *generator) compileNativeUnionTypedPatternCondition(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern) ([]string, string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, "", false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		ctx.setReason("typed pattern type mismatch")
		return nil, "", false
	}
	if target.Member == nil {
		return g.compileMatchPatternCondition(ctx, pattern.Pattern, subjectTemp, subjectType)
	}
	okTemp := ctx.newTemp()
	castTemp := ctx.newTemp()
	innerCondLines, innerCond, ok := g.compileMatchPatternCondition(ctx, pattern.Pattern, castTemp, target.GoType)
	if !ok {
		return nil, "", false
	}
	if innerCond == "true" && len(innerCondLines) == 0 {
		return []string{fmt.Sprintf("_, %s := %s(%s)", okTemp, target.Member.UnwrapHelper, subjectTemp)}, okTemp, true
	}
	condTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := %s(%s)", castTemp, okTemp, target.Member.UnwrapHelper, subjectTemp),
		fmt.Sprintf("var %s bool", condTemp),
		fmt.Sprintf("if %s {", okTemp),
	}
	lines = append(lines, indentLines(innerCondLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", condTemp, innerCond))
	lines = append(lines, "}")
	return lines, condTemp, true
}

func (g *generator) compileNativeUnionTypedPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern) ([]string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		ctx.setReason("typed pattern type mismatch")
		return nil, false
	}
	if target.Member == nil {
		return g.compileMatchPatternBindings(ctx, pattern.Pattern, subjectTemp, subjectType)
	}
	convertedTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", convertedTemp, target.Member.UnwrapHelper, subjectTemp)}
	bindSubject := convertedTemp
	bindType := target.GoType
	if target.InterfaceBranch && nativeUnionWholeValueBinding(pattern.Pattern) {
		runtimeLines, runtimeExpr, ok := g.runtimeValueLines(ctx, convertedTemp, target.GoType)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		lines = append(lines, runtimeLines...)
		bindSubject = runtimeExpr
		bindType = "runtime.Value"
	}
	bindLines, ok := g.compileMatchPatternBindings(ctx, pattern.Pattern, bindSubject, bindType)
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines = append(lines, bindLines...)
	return lines, true
}

func (g *generator) compileNativeUnionTypedAssignmentPatternBindings(ctx *compileContext, subjectTemp string, subjectType string, pattern *ast.TypedPattern, mode patternBindingMode) ([]string, bool) {
	if pattern == nil || pattern.TypeAnnotation == nil {
		ctx.setReason("missing typed pattern annotation")
		return nil, false
	}
	target, ok := g.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		ctx.setReason("typed pattern type mismatch")
		return nil, false
	}
	if target.Member == nil {
		return g.compileAssignmentPatternBindings(ctx, pattern.Pattern, subjectTemp, subjectType, mode)
	}
	convertedTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s, _ := %s(%s)", convertedTemp, target.Member.UnwrapHelper, subjectTemp)}
	bindSubject := convertedTemp
	bindType := target.GoType
	if target.InterfaceBranch && nativeUnionWholeValueBinding(pattern.Pattern) {
		runtimeLines, runtimeExpr, ok := g.runtimeValueLines(ctx, convertedTemp, target.GoType)
		if !ok {
			ctx.setReason("typed pattern type mismatch")
			return nil, false
		}
		lines = append(lines, runtimeLines...)
		bindSubject = runtimeExpr
		bindType = "runtime.Value"
	}
	bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern.Pattern, bindSubject, bindType, mode)
	if !ok {
		return nil, false
	}
	if len(bindLines) == 0 {
		return nil, true
	}
	lines = append(lines, bindLines...)
	return lines, true
}
