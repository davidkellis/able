package compiler

func (g *generator) nativeInterfaceBoundarySiblingInfos(info *nativeInterfaceInfo) []*nativeInterfaceInfo {
	if g == nil || info == nil || info.TypeExpr == nil {
		return nil
	}
	ifacePkg, ifaceName, _, _, ok := interfaceExprInfo(g, "", info.TypeExpr)
	if !ok || ifacePkg == "" || ifaceName == "" {
		return nil
	}
	seen := map[string]struct{}{}
	infos := make([]*nativeInterfaceInfo, 0)
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		fn := candidateInfo.info
		if impl == nil || fn == nil || impl.ImplName != "" {
			continue
		}
		actualExpr := g.implConcreteInterfaceExpr(impl, fn.TypeBindings)
		if actualExpr == nil {
			continue
		}
		actualPkg, actualName, _, _, ok := interfaceExprInfo(g, fn.Package, actualExpr)
		if !ok || actualPkg != ifacePkg || actualName != ifaceName {
			continue
		}
		if !g.typeExprFullyBound(actualPkg, actualExpr) {
			continue
		}
		actualInfo, ok := g.ensureNativeInterfaceInfo(actualPkg, actualExpr)
		if !ok || actualInfo == nil || actualInfo.Key == info.Key {
			continue
		}
		if _, ok := seen[actualInfo.Key]; ok {
			continue
		}
		seen[actualInfo.Key] = struct{}{}
		infos = append(infos, actualInfo)
	}
	return infos
}
