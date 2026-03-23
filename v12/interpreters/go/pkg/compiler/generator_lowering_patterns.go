package compiler

import "able/interpreter-go/pkg/ast"

// lowerJoinCarrier is the canonical join-synthesis entrypoint for static
// control-flow result carriers.
func (g *generator) lowerJoinCarrier(ctx *compileContext, types ...string) (string, bool) {
	return g.joinResultType(ctx, types...)
}

// lowerJoinCarrierFromBranches is the canonical join-synthesis entrypoint when
// branch metadata is available.
func (g *generator) lowerJoinCarrierFromBranches(ctx *compileContext, branches []joinBranchInfo) (string, bool) {
	return g.joinResultTypeFromBranches(ctx, branches)
}

// lowerJoinCarrierAllowNil is the canonical join-synthesis entrypoint for
// joins that must model nil-capable branches explicitly.
func (g *generator) lowerJoinCarrierAllowNil(ctx *compileContext, types []string, nilFlags []bool) (string, bool) {
	return g.joinResultTypeAllowNil(ctx, types, nilFlags)
}

// lowerStaticPatternCompatible is the canonical static typed-pattern
// compatibility entrypoint.
func (g *generator) lowerStaticPatternCompatible(subjectType string, patternType string) bool {
	return g.staticTypedPatternCompatible(subjectType, patternType)
}

// lowerPatternRuntimeTypeCheck is the canonical runtime type-check synthesis
// entrypoint for typed patterns that still operate on dynamic subjects.
func (g *generator) lowerPatternRuntimeTypeCheck(ctx *compileContext, expr ast.TypeExpression, subjectTemp string) ([]string, string, bool) {
	return g.runtimeTypeCheckForTypeExpression(ctx, expr, subjectTemp)
}

// lowerNativeUnionPatternMemberType is the canonical pattern carrier lookup
// entrypoint for native union typed-pattern narrowing.
func (g *generator) lowerNativeUnionPatternMemberType(ctx *compileContext, subjectType string, patternType ast.TypeExpression) (string, bool) {
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	return g.nativeUnionPatternMemberType(subjectType, patternType, pkgName)
}

// lowerNativeUnionPatternMemberTypeInPackage is the canonical package-scoped
// pattern carrier lookup entrypoint for native union typed-pattern narrowing.
func (g *generator) lowerNativeUnionPatternMemberTypeInPackage(pkgName string, subjectType string, patternType ast.TypeExpression) (string, bool) {
	return g.nativeUnionPatternMemberType(subjectType, patternType, pkgName)
}
