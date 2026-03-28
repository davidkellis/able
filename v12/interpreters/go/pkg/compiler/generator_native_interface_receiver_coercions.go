package compiler

import (
	"bytes"
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) nativeInterfaceDefaultReceiverInfo(receiverGoType string, method *nativeInterfaceGenericMethod, bindings map[string]ast.TypeExpression) (ast.TypeExpression, string, map[string]ast.TypeExpression, bool) {
	if g == nil || method == nil || receiverGoType == "" {
		return nil, "", nil, false
	}
	receiverTypeExpr, ok := g.typeExprForGoType(receiverGoType)
	if !ok || receiverTypeExpr == nil {
		return nil, "", nil, false
	}
	mergedBindings := g.nativeInterfaceGenericDefaultMethodBindings(method, bindings)
	receiverIsNativeInterface := g.nativeInterfaceInfoForGoType(receiverGoType) != nil
	if receiverBindings, ok := g.nativeInterfaceDefaultReceiverBindings(receiverTypeExpr, method); ok {
		if mergedBindings == nil {
			mergedBindings = make(map[string]ast.TypeExpression, len(receiverBindings))
		}
		if !g.mergeConcreteBindings(method.InterfacePackage, mergedBindings, receiverBindings) {
			return nil, "", nil, false
		}
	} else if receiverIsNativeInterface {
		return nil, "", nil, false
	}
	if concreteReceiverExpr, concreteReceiverGoType, ok := g.nativeInterfaceConcreteStructReceiverFromBindings(receiverGoType, receiverTypeExpr, method.InterfacePackage, mergedBindings); ok {
		receiverTypeExpr = concreteReceiverExpr
		receiverGoType = concreteReceiverGoType
	} else if g.isNativeStructPointerType(receiverGoType) {
		structInfo := g.structInfoByGoName(receiverGoType)
		pkgName := method.InterfacePackage
		if structInfo != nil && structInfo.Package != "" {
			pkgName = structInfo.Package
		}
		if receiverTypeExpr != nil && !g.typeExprFullyBound(pkgName, receiverTypeExpr) {
			return nil, "", nil, false
		}
	}
	if g.nativeInterfaceInfoForGoType(receiverGoType) == nil {
		concreteBindings, ok := g.nativeInterfaceConcreteReceiverBindings(receiverGoType, method, mergedBindings)
		if !ok {
			return nil, "", nil, false
		}
		mergedBindings = concreteBindings
		if concreteReceiverExpr, concreteReceiverGoType, ok := g.nativeInterfaceConcreteStructReceiverFromBindings(receiverGoType, receiverTypeExpr, method.InterfacePackage, mergedBindings); ok {
			receiverTypeExpr = concreteReceiverExpr
			receiverGoType = concreteReceiverGoType
		} else if g.isNativeStructPointerType(receiverGoType) {
			structInfo := g.structInfoByGoName(receiverGoType)
			pkgName := method.InterfacePackage
			if structInfo != nil && structInfo.Package != "" {
				pkgName = structInfo.Package
			}
			if receiverTypeExpr != nil && !g.typeExprFullyBound(pkgName, receiverTypeExpr) {
				return nil, "", nil, false
			}
		}
	}
	iface := g.interfaces[method.InterfaceName]
	if iface != nil {
		for name, expr := range g.interfaceSelfTypeBindings(iface, receiverTypeExpr) {
			if expr == nil {
				continue
			}
			if mergedBindings == nil {
				mergedBindings = make(map[string]ast.TypeExpression)
			}
			mergedBindings[name] = normalizeTypeExprForPackage(g, method.InterfacePackage, expr)
		}
		preserveStructReceiver := g.preserveNativeInterfaceDefaultStructReceiver(method.InterfacePackage, receiverGoType, receiverTypeExpr)
		if iface.SelfTypePattern != nil {
			concreteReceiver := normalizeTypeExprForPackage(g, method.InterfacePackage, substituteTypeParams(iface.SelfTypePattern, mergedBindings))
			if generic, ok := concreteReceiver.(*ast.GenericTypeExpression); ok && generic != nil {
				if base, ok := generic.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && isBuiltinMappedType(base.Name.Name) {
					concreteReceiver = ast.Ty(base.Name.Name)
				}
			}
			mapper := NewTypeMapper(g, method.InterfacePackage)
			if mapped, ok := mapper.Map(concreteReceiver); ok && mapped != "" {
				if recovered, ok := g.recoverRepresentableCarrierType(method.InterfacePackage, concreteReceiver, mapped); ok && recovered != "" && recovered != "runtime.Value" && recovered != "any" {
					if g.keepActualDefaultReceiverForDifferentStructFamily(receiverGoType, recovered) {
						return receiverTypeExpr, receiverGoType, mergedBindings, true
					}
					if preserveStructReceiver && g.preferActualDefaultReceiverGoType(receiverGoType, recovered) {
						return receiverTypeExpr, receiverGoType, mergedBindings, true
					}
					return concreteReceiver, recovered, mergedBindings, true
				}
				if g.nativeInterfaceInfoForGoType(receiverGoType) != nil && (mapped == "runtime.Value" || mapped == "any") {
					return receiverTypeExpr, receiverGoType, mergedBindings, true
				}
				if g.keepActualDefaultReceiverForDifferentStructFamily(receiverGoType, mapped) {
					return receiverTypeExpr, receiverGoType, mergedBindings, true
				}
				if preserveStructReceiver && g.preferActualDefaultReceiverGoType(receiverGoType, mapped) {
					return receiverTypeExpr, receiverGoType, mergedBindings, true
				}
				return concreteReceiver, mapped, mergedBindings, true
			}
		}
		if preserveStructReceiver {
			return receiverTypeExpr, receiverGoType, mergedBindings, true
		}
	}
	return receiverTypeExpr, receiverGoType, mergedBindings, true
}

func (g *generator) preferActualDefaultReceiverGoType(actualGoType string, candidateGoType string) bool {
	if g == nil || actualGoType == "" || candidateGoType == "" {
		return false
	}
	if actualGoType == candidateGoType {
		return true
	}
	if g.receiverNominalFamilyCompatible(candidateGoType, actualGoType) {
		return true
	}
	return g.sameNominalStructFamily(candidateGoType, actualGoType)
}

func (g *generator) keepActualDefaultReceiverForDifferentStructFamily(actualGoType string, candidateGoType string) bool {
	if g == nil || actualGoType == "" || candidateGoType == "" {
		return false
	}
	if !g.isNativeStructPointerType(actualGoType) || !g.isNativeStructPointerType(candidateGoType) {
		return false
	}
	if actualGoType == candidateGoType {
		return true
	}
	if g.receiverNominalFamilyCompatible(candidateGoType, actualGoType) {
		return false
	}
	if g.sameNominalStructFamily(candidateGoType, actualGoType) {
		return false
	}
	return true
}

func (g *generator) nativeInterfaceDefaultReceiverBindings(receiverTypeExpr ast.TypeExpression, method *nativeInterfaceGenericMethod) (map[string]ast.TypeExpression, bool) {
	if g == nil || method == nil || receiverTypeExpr == nil {
		return nil, false
	}
	expectedExpr := nativeInterfaceInstantiationExpr(method.InterfaceName, method.InterfaceArgs)
	genericNames := nativeInterfaceGenericNameSet(method.GenericParams)
	genericNames = mergeGenericNameSets(genericNames, g.typeExprVariableNames(expectedExpr))
	return g.nativeInterfaceImplBindingsForTarget(
		method.InterfacePackage,
		receiverTypeExpr,
		genericNames,
		method.InterfacePackage,
		method.InterfaceName,
		method.InterfaceArgs,
		make(map[string]struct{}),
	)
}

func (g *generator) nativeInterfaceConcreteStructReceiverFromBindings(receiverGoType string, receiverTypeExpr ast.TypeExpression, fallbackPkg string, bindings map[string]ast.TypeExpression) (ast.TypeExpression, string, bool) {
	if g == nil || receiverGoType == "" || receiverTypeExpr == nil || !g.isNativeStructPointerType(receiverGoType) || len(bindings) == 0 {
		return nil, "", false
	}
	structInfo := g.structInfoByGoName(receiverGoType)
	pkgName := fallbackPkg
	if structInfo != nil && structInfo.Package != "" {
		pkgName = structInfo.Package
	}
	concreteExpr := g.reconstructGenericStructTargetFromBindings(pkgName, receiverTypeExpr, bindings)
	if concreteExpr == nil {
		concreteExpr = normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(receiverTypeExpr, bindings))
	}
	if concreteExpr == nil || !g.typeExprFullyBound(pkgName, concreteExpr) {
		return nil, "", false
	}
	if normalizeTypeExprString(g, pkgName, concreteExpr) == normalizeTypeExprString(g, pkgName, receiverTypeExpr) {
		return concreteExpr, receiverGoType, true
	}
	if goType, ok := g.nativeStructCarrierTypeForExpr(pkgName, concreteExpr); ok && goType != "" {
		return concreteExpr, goType, true
	}
	mapper := NewTypeMapper(g, pkgName)
	goType, ok := mapper.Map(concreteExpr)
	if !ok || goType == "" || goType == "runtime.Value" || goType == "any" {
		return nil, "", false
	}
	return concreteExpr, goType, true
}

func (g *generator) nativeInterfaceConcreteReceiverBindings(receiverGoType string, method *nativeInterfaceGenericMethod, bindings map[string]ast.TypeExpression) (map[string]ast.TypeExpression, bool) {
	if g == nil || method == nil || receiverGoType == "" {
		return nil, false
	}
	expectedExpr := nativeInterfaceInstantiationExpr(method.InterfaceName, method.InterfaceArgs)
	genericNames := nativeInterfaceGenericNameSet(method.GenericParams)
	genericNames = mergeGenericNameSets(genericNames, g.typeExprVariableNames(expectedExpr))
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || impl.ImplName != "" || impl.InterfaceName != method.InterfaceName {
			continue
		}
		merged := cloneTypeBindings(bindings)
		if merged == nil {
			merged = make(map[string]ast.TypeExpression)
		}
		concreteInfo, concreteBindings, ok := g.nativeInterfaceConcreteImplInfo(receiverGoType, impl, merged)
		if !ok || concreteInfo == nil {
			continue
		}
		actualExpr := g.implConcreteInterfaceExpr(impl, concreteBindings)
		if actualExpr == nil {
			continue
		}
		if _, ok := g.nativeInterfaceImplBindingsForTarget(
			concreteInfo.Package,
			actualExpr,
			genericNames,
			method.InterfacePackage,
			method.InterfaceName,
			method.InterfaceArgs,
			make(map[string]struct{}),
		); !ok {
			continue
		}
		concreteBindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(concreteInfo, concreteBindings)
		if !ok {
			continue
		}
		return concreteBindings, true
	}
	return nil, false
}

func (g *generator) preserveNativeInterfaceDefaultStructReceiver(pkgName string, receiverGoType string, receiverTypeExpr ast.TypeExpression) bool {
	if g == nil || !g.isNativeStructPointerType(receiverGoType) {
		return false
	}
	if receiverTypeExpr == nil {
		return true
	}
	baseName, ok := typeExprBaseName(receiverTypeExpr)
	if !ok || !isBuiltinMappedType(baseName) {
		return true
	}
	mapper := NewTypeMapper(g, pkgName)
	mapped, ok := mapper.Map(receiverTypeExpr)
	if !ok || mapped == "" {
		return true
	}
	return mapped == receiverGoType
}

func (g *generator) prepareConcreteNativeInterfaceReceiver(ctx *compileContext, receiverExpr string, receiverType string, impl *nativeInterfaceAdapterMethod) ([]string, string, string, bool) {
	if g == nil || ctx == nil || receiverExpr == "" || receiverType == "" || impl == nil || impl.Info == nil || len(impl.Info.Params) == 0 {
		return nil, "", "", false
	}
	expectedReceiverType := g.canonicalMethodReceiverGoType(impl.Info, receiverType)
	if expectedReceiverType == "" {
		return nil, receiverExpr, receiverType, true
	}
	return g.prepareStaticCallArg(ctx, receiverExpr, receiverType, expectedReceiverType)
}

func (g *generator) canonicalMethodReceiverGoType(info *functionInfo, actualGoType string) string {
	if g == nil || info == nil || len(info.Params) == 0 {
		return ""
	}
	expectedReceiverType := info.Params[0].GoType
	if actualGoType == "" {
		return expectedReceiverType
	}
	meta := g.nativeInterfaceDefaultByInfo[info]
	if meta == nil {
		return expectedReceiverType
	}
	method := &nativeInterfaceGenericMethod{
		Name:             meta.MethodName,
		InterfaceName:    meta.InterfaceName,
		InterfacePackage: meta.InterfacePackage,
		InterfaceArgs:    cloneTypeExprSlice(meta.InterfaceArgs),
	}
	_, canonicalGoType, _, ok := g.nativeInterfaceDefaultReceiverInfo(actualGoType, method, nil)
	if !ok || canonicalGoType == "" {
		return expectedReceiverType
	}
	return canonicalGoType
}

func (g *generator) renderNativeInterfaceReceiverCoercionControl(buf *bytes.Buffer, indent string, target string, expr string, actualGoType string, expectedGoType string, failureGoType string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualGoType == "" || expectedGoType == "" || failureGoType == "" {
		return false
	}
	if actualGoType == expectedGoType {
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	}
	if !g.sameNominalStructFamily(expectedGoType, actualGoType) {
		return false
	}
	info := g.ensureNominalCoercionInfo(actualGoType, expectedGoType)
	if info == nil {
		return false
	}
	zeroExpr, ok := g.zeroValueExpr(failureGoType)
	if !ok {
		return false
	}
	fmt.Fprintf(buf, "%s%s, __able_receiver_coerce_err := %s(%s)\n", indent, target, info.HelperName, expr)
	fmt.Fprintf(buf, "%sif __able_receiver_coerce_err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn %s, __able_control_from_error(__able_receiver_coerce_err)\n", indent, zeroExpr)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}
