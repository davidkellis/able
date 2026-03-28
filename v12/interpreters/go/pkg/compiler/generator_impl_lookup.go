package compiler

import (
	"strconv"
	"strings"
)

func (g *generator) implMethodInfoForFunction(info *functionInfo) *implMethodInfo {
	if g == nil || info == nil {
		return nil
	}
	if g.implMethodByInfo != nil {
		if impl := g.implMethodByInfo[info]; impl != nil {
			return impl
		}
	}
	var found *implMethodInfo
	foundKey := ""
	for _, impl := range g.implMethodsBySignature[functionInfoSignatureKey(info)] {
		key := g.implMethodCanonicalKey(impl)
		if found != nil && key != foundKey {
			return nil
		}
		found = impl
		foundKey = key
	}
	if found != nil && g.implMethodByInfo != nil {
		g.implMethodByInfo[info] = found
	}
	return found
}

func functionInfoSignatureKey(info *functionInfo) string {
	if info == nil {
		return ""
	}
	logicalName := strings.TrimSpace(info.Name)
	if info.QualifiedName != "" {
		logicalName = strings.TrimSpace(info.QualifiedName)
	}
	if logicalName == "" {
		logicalName = functionInfoSourceName(info)
	}
	parts := []string{
		info.Package,
		logicalName,
		functionInfoSourceName(info),
		info.ReturnType,
	}
	for _, param := range info.Params {
		parts = append(parts, param.GoType)
	}
	return strings.Join(parts, "|")
}

func (g *generator) implMethodCanonicalKey(impl *implMethodInfo) string {
	if impl == nil {
		return ""
	}
	pkgName := ""
	if impl.Info != nil {
		pkgName = impl.Info.Package
	}
	return pkgName + "::" +
		impl.InterfaceName + "::" +
		impl.MethodName + "::" +
		normalizeTypeExprString(g, pkgName, impl.TargetType) + "::" +
		constraintSignature(collectConstraintSpecs(impl.ImplGenerics, impl.WhereClause)) + "::" +
		impl.ImplName + "::" +
		strconv.FormatBool(impl.IsDefault)
}
