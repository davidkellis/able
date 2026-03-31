package compiler

import (
	"sort"

	"able/interpreter-go/pkg/ast"
)

type concreteInterfaceMethodCandidate struct {
	method         *methodInfo
	impl           *implMethodInfo
	receiverExact  bool
	targetScore    int
	isConcrete     bool
	isBuiltin      bool
	constraintKeys map[string]struct{}
	unionVariants  []string
}

func (g *generator) compileableInterfaceMethodForConcreteReceiverExpr(ctx *compileContext, expr ast.Expression, goType string, methodName string) *methodInfo {
	if g == nil || goType == "" || goType == "runtime.Value" || goType == "any" || methodName == "" {
		return nil
	}
	actualExpr, _ := g.inferExpressionTypeExpr(ctx, expr, goType)
	if actualExpr == nil {
		actualExpr, _ = g.typeExprForGoType(goType)
	}
	if actualExpr != nil && ctx != nil {
		actualExpr = normalizeTypeExprForPackage(g, ctx.packageName, actualExpr)
	}
	if !g.typeExprCompatibleWithCarrier(ctx, actualExpr, goType) {
		actualExpr = nil
		if fallbackExpr, ok := g.typeExprForGoType(goType); ok && fallbackExpr != nil {
			if ctx != nil {
				fallbackExpr = normalizeTypeExprForPackage(g, ctx.packageName, fallbackExpr)
			}
			if g.typeExprCompatibleWithCarrier(ctx, fallbackExpr, goType) {
				actualExpr = fallbackExpr
			}
		}
	}
	var found *concreteInterfaceMethodCandidate
	for _, impl := range g.implMethodList {
		candidate := g.concreteInterfaceMethodCandidate(ctx, impl, actualExpr, goType, methodName)
		if candidate == nil {
			continue
		}
		if found == nil {
			found = candidate
			continue
		}
		switch g.compareConcreteInterfaceMethodCandidates(found, candidate) {
		case -1:
			found = candidate
		case 0:
			if found.method == nil || candidate.method == nil || found.method.Info != candidate.method.Info {
				return nil
			}
		}
	}
	if found == nil {
		return nil
	}
	if found.impl != nil && actualExpr != nil {
		genericNames := g.implSpecializationGenericNames(found.method)
		if !g.typeExprHasGeneric(actualExpr, genericNames) {
			bindings := cloneTypeBindings(found.method.Info.TypeBindings)
			if bindings == nil {
				bindings = make(map[string]ast.TypeExpression)
			}
			if g.seedImplBindingsFromConcreteTarget(found.method, found.impl, actualExpr, bindings) {
				bindings = g.normalizeConcreteTypeBindings(found.method.Info.Package, bindings, genericNames)
				if len(bindings) > 0 {
					if specialized, ok := g.ensureSpecializedImplMethod(found.method, found.impl, bindings); ok && specialized != nil && specialized.Info != nil && specialized.Info.Compileable {
						return specialized
					}
				}
			}
		}
	}
	return found.method
}

func (g *generator) concreteInterfaceMethodCandidate(ctx *compileContext, impl *implMethodInfo, actualExpr ast.TypeExpression, goType string, methodName string) *concreteInterfaceMethodCandidate {
	if g == nil || impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" {
		return nil
	}
	if impl.MethodName != methodName || len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType == "" {
		return nil
	}
	receiverType := g.implReceiverGoType(impl)
	receiverCompatible := g.receiverGoTypeCompatible(receiverType, goType)
	if !receiverCompatible && (actualExpr == nil || impl.TargetType == nil) {
		return nil
	}
	genericNames := g.callableGenericNames(impl.Info)
	currentGenericNames := genericNames
	if ctx != nil {
		currentGenericNames = mergeGenericNameSets(currentGenericNames, ctx.genericNames)
	}
	if actualExpr != nil && impl.TargetType != nil {
		targetType := normalizeTypeExprForPackage(g, impl.Info.Package, impl.TargetType)
		if !g.typeExprHasGeneric(targetType, genericNames) &&
			g.usesNominalStructCarrier(impl.Info.Package, targetType) &&
			!g.nominalTargetTypeExprCompatible(impl.Info.Package, actualExpr, impl.TargetType) {
			return nil
		}
		if receiverType != goType && g.typeExprHasGeneric(actualExpr, currentGenericNames) {
			return nil
		}
		bindings := cloneTypeBindings(impl.Info.TypeBindings)
		if bindings == nil {
			bindings = make(map[string]ast.TypeExpression)
		}
		if !g.specializedTypeTemplateMatches(impl.Info.Package, impl.TargetType, actualExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil
		}
		if !g.implConstraintsSatisfied(impl.Info.Package, impl, bindings) {
			return nil
		}
	} else if !receiverCompatible {
		return nil
	}
	targetName, _ := g.methodTargetName(impl.TargetType)
	method := &methodInfo{
		TargetName:   targetName,
		TargetType:   impl.TargetType,
		MethodName:   methodName,
		ReceiverType: receiverType,
		ExpectsSelf:  true,
		Info:         impl.Info,
	}
	targetExpr := normalizeTypeExprForPackage(g, impl.Info.Package, impl.TargetType)
	return &concreteInterfaceMethodCandidate{
		method:         method,
		impl:           impl,
		receiverExact:  receiverType == goType,
		targetScore:    g.typeExprTemplateSpecificity(targetExpr, genericNames),
		isConcrete:     !g.typeExprHasGeneric(targetExpr, genericNames),
		isBuiltin:      targetName != "" && isBuiltinMappedType(targetName),
		constraintKeys: g.constraintKeySetForImpl(impl),
		unionVariants:  g.expandUnionTargetSignatures(targetExpr),
	}
}

func (g *generator) implReceiverGoType(impl *implMethodInfo) string {
	if g == nil || impl == nil || impl.Info == nil {
		return ""
	}
	return g.nativeInterfaceImplWitnessGoType(impl.Info, impl, impl.Info.TypeBindings)
}

func (g *generator) implConstraintsSatisfied(pkgName string, impl *implMethodInfo, bindings map[string]ast.TypeExpression) bool {
	return g.implConstraintsSatisfiedSeen(pkgName, impl, bindings, make(map[string]struct{}))
}

func (g *generator) implConstraintsSatisfiedSeen(pkgName string, impl *implMethodInfo, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || impl == nil {
		return false
	}
	specs := collectConstraintSpecs(impl.ImplGenerics, impl.WhereClause)
	if len(specs) == 0 {
		return true
	}
	for _, spec := range specs {
		if spec.subject == nil || spec.iface == nil {
			return false
		}
		subject := normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(spec.subject, bindings))
		ifaceExpr := normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(spec.iface, bindings))
		if subject == nil || ifaceExpr == nil {
			return false
		}
		if g.typeExprHasGeneric(subject, nil) || g.typeExprHasGeneric(ifaceExpr, nil) {
			return false
		}
		subjectGoType, ok := g.lowerCarrierTypeInPackage(pkgName, subject)
		subjectGoType, ok = g.recoverRepresentableCarrierType(pkgName, subject, subjectGoType)
		if !ok || subjectGoType == "" || subjectGoType == "runtime.Value" || subjectGoType == "any" {
			return false
		}
		ifaceInfo, ok := g.ensureNativeInterfaceInfo(pkgName, ifaceExpr)
		if !ok || ifaceInfo == nil {
			return false
		}
		if !g.nativeInterfaceAcceptsActualSeen(ifaceInfo, subjectGoType, seen) {
			return false
		}
	}
	return true
}

func (g *generator) compareConcreteInterfaceMethodCandidates(a *concreteInterfaceMethodCandidate, b *concreteInterfaceMethodCandidate) int {
	if a == nil || b == nil {
		switch {
		case a == nil && b != nil:
			return -1
		case a != nil && b == nil:
			return 1
		default:
			return 0
		}
	}
	if a.receiverExact && !b.receiverExact {
		return 1
	}
	if b.receiverExact && !a.receiverExact {
		return -1
	}
	if a.isConcrete && !b.isConcrete {
		return 1
	}
	if b.isConcrete && !a.isConcrete {
		return -1
	}
	if constraintKeySetSuperset(a.constraintKeys, b.constraintKeys) {
		return 1
	}
	if constraintKeySetSuperset(b.constraintKeys, a.constraintKeys) {
		return -1
	}
	aUnion := a.unionVariants
	bUnion := b.unionVariants
	if len(aUnion) > 0 && len(bUnion) == 0 {
		return -1
	}
	if len(aUnion) == 0 && len(bUnion) > 0 {
		return 1
	}
	if len(aUnion) > 0 && len(bUnion) > 0 {
		if stringSliceProperSubset(aUnion, bUnion) {
			return 1
		}
		if stringSliceProperSubset(bUnion, aUnion) {
			return -1
		}
		if len(aUnion) != len(bUnion) {
			if len(aUnion) < len(bUnion) {
				return 1
			}
			return -1
		}
	}
	if a.targetScore > b.targetScore {
		return 1
	}
	if a.targetScore < b.targetScore {
		return -1
	}
	if a.isBuiltin != b.isBuiltin {
		if a.isBuiltin {
			return -1
		}
		return 1
	}
	return 0
}

func (g *generator) typeExprTemplateSpecificity(expr ast.TypeExpression, genericNames map[string]struct{}) int {
	if g == nil || expr == nil {
		return 0
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return 0
		}
		if _, ok := genericNames[t.Name.Name]; ok {
			return 0
		}
		return 1
	case *ast.GenericTypeExpression:
		score := g.typeExprTemplateSpecificity(t.Base, genericNames)
		for _, arg := range t.Arguments {
			score += g.typeExprTemplateSpecificity(arg, genericNames)
		}
		return score
	case *ast.NullableTypeExpression:
		return g.typeExprTemplateSpecificity(t.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return g.typeExprTemplateSpecificity(t.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		score := 0
		for _, member := range t.Members {
			score += g.typeExprTemplateSpecificity(member, genericNames)
		}
		return score
	default:
		return 0
	}
}

func (g *generator) constraintKeySetForImpl(impl *implMethodInfo) map[string]struct{} {
	if g == nil || impl == nil {
		return nil
	}
	specs := collectConstraintSpecs(impl.ImplGenerics, impl.WhereClause)
	if len(specs) == 0 {
		return nil
	}
	keys := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		if spec.subject == nil || spec.iface == nil {
			continue
		}
		subject := normalizeTypeExprString(g, impl.Info.Package, spec.subject)
		iface := normalizeTypeExprString(g, impl.Info.Package, spec.iface)
		if subject == "" || iface == "" {
			continue
		}
		keys[subject+"->"+iface] = struct{}{}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func constraintKeySetSuperset(a map[string]struct{}, b map[string]struct{}) bool {
	if len(a) <= len(b) {
		return false
	}
	for key := range b {
		if _, ok := a[key]; !ok {
			return false
		}
	}
	return true
}

func stringSliceProperSubset(a []string, b []string) bool {
	if len(a) == 0 {
		return len(b) > 0
	}
	setA := make(map[string]struct{}, len(a))
	for _, value := range a {
		setA[value] = struct{}{}
	}
	setB := make(map[string]struct{}, len(b))
	for _, value := range b {
		setB[value] = struct{}{}
	}
	if len(setA) >= len(setB) {
		return false
	}
	for value := range setA {
		if _, ok := setB[value]; !ok {
			return false
		}
	}
	return true
}

func (g *generator) expandUnionTargetSignatures(expr ast.TypeExpression) []string {
	if g == nil || expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.UnionTypeExpression:
		var variants []string
		seen := make(map[string]struct{})
		for _, member := range t.Members {
			for _, sig := range g.expandUnionTargetSignatures(member) {
				if _, ok := seen[sig]; ok {
					continue
				}
				seen[sig] = struct{}{}
				variants = append(variants, sig)
			}
		}
		sort.Strings(variants)
		return variants
	default:
		sig := typeExpressionToString(expr)
		if sig == "" {
			return nil
		}
		return []string{sig}
	}
}
