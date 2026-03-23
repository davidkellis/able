package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeInterfaceGenericDispatchInfo struct {
	Key              string
	GoName           string
	Package          string
	InterfaceKey     string
	InterfaceGoType  string
	Method           *nativeInterfaceGenericMethod
	ParamTypeExprs   []ast.TypeExpression
	ParamGoTypes     []string
	ReturnTypeExpr   ast.TypeExpression
	ReturnGoType     string
	Cases            []nativeInterfaceGenericDispatchCase
	RuntimeDefault   *functionInfo
	MonoArrayCollect *iteratorCollectMonoArrayInfo
}

type nativeInterfaceGenericDispatchCase struct {
	AdapterType string
	GoType      string
	Impl        *nativeInterfaceAdapterMethod
}

func (g *generator) nativeInterfaceAdapterMethodSpecificity(method *nativeInterfaceAdapterMethod) int {
	if g == nil || method == nil {
		return 0
	}
	score := 0
	for _, goType := range method.CompiledParamGoTypes {
		if goType != "" && goType != "runtime.Value" && goType != "any" {
			score += 2
		}
	}
	if method.CompiledReturnGoType != "" && method.CompiledReturnGoType != "runtime.Value" && method.CompiledReturnGoType != "any" {
		score += 3
	}
	return score
}

func (g *generator) ensureNativeInterfaceGenericDispatchInfo(ctx *compileContext, call *ast.FunctionCall, expected string, info *nativeInterfaceInfo, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string, bindings map[string]ast.TypeExpression) (*nativeInterfaceGenericDispatchInfo, bool) {
	if g == nil || ctx == nil || call == nil || info == nil || method == nil || returnGoType == "" {
		return nil, false
	}
	if len(paramTypeExprs) != len(paramGoTypes) {
		return nil, false
	}
	cases := g.nativeInterfaceGenericDispatchCases(ctx, call, expected, info, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType)
	if len(cases) == 0 {
		return nil, false
	}
	parts := []string{
		info.Key,
		method.Name,
		normalizeTypeExprListKey(g, method.InterfacePackage, paramTypeExprs),
		normalizeTypeExprString(g, method.InterfacePackage, returnTypeExpr),
	}
	key := strings.Join(parts, "::")
	if existing, ok := g.nativeInterfaceGenericDispatches[key]; ok && existing != nil {
		return existing, true
	}
	var runtimeDefault *functionInfo
	var monoArrayCollect *iteratorCollectMonoArrayInfo
	if method.DefaultDefinition != nil {
		if info.RuntimeIteratorAdapter != "" {
			if defaultInfo, ok := g.ensureSpecializedNativeInterfaceGenericDefaultMethod(method, info.RuntimeIteratorAdapter, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings); ok && defaultInfo != nil && defaultInfo.Compileable {
				runtimeDefault = defaultInfo
			}
		}
		if runtimeDefault == nil {
			// Reuse the shared interface-carrier default helper first. Runtime
			// adapters already satisfy the native interface type, so dispatch should
			// stay on the same compiled helper instead of forcing a receiver-specific
			// specialization.
			if defaultInfo, ok := g.ensureSpecializedNativeInterfaceGenericDefaultMethod(method, info.GoType, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings); ok && defaultInfo != nil && defaultInfo.Compileable {
				runtimeDefault = defaultInfo
			} else if info.RuntimeAdapter != "" {
				if defaultInfo, ok := g.ensureSpecializedNativeInterfaceGenericDefaultMethod(method, info.RuntimeAdapter, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings); ok && defaultInfo != nil && defaultInfo.Compileable {
					runtimeDefault = defaultInfo
				}
			}
		}
	}
	if method.Name == "collect" {
		if collectInfo, ok := g.ensureIteratorCollectMonoArrayInfo(method, info.GoType, returnGoType); ok && collectInfo != nil {
			monoArrayCollect = collectInfo
		}
	}
	dispatch := &nativeInterfaceGenericDispatchInfo{
		Key:              key,
		GoName:           g.mangler.unique(fmt.Sprintf("iface_%s_%s_dispatch", sanitizeIdent(method.InterfaceName), sanitizeIdent(method.Name))),
		Package:          method.InterfacePackage,
		InterfaceKey:     info.Key,
		InterfaceGoType:  info.GoType,
		Method:           method,
		ParamTypeExprs:   append([]ast.TypeExpression(nil), paramTypeExprs...),
		ParamGoTypes:     append([]string(nil), paramGoTypes...),
		ReturnTypeExpr:   returnTypeExpr,
		ReturnGoType:     returnGoType,
		Cases:            cases,
		RuntimeDefault:   runtimeDefault,
		MonoArrayCollect: monoArrayCollect,
	}
	g.nativeInterfaceGenericDispatches[key] = dispatch
	return dispatch, true
}

func (g *generator) nativeInterfaceGenericDispatchCases(ctx *compileContext, call *ast.FunctionCall, expected string, info *nativeInterfaceInfo, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string) []nativeInterfaceGenericDispatchCase {
	if g == nil || ctx == nil || call == nil || info == nil || method == nil {
		return nil
	}
	specializationExpected := expected
	if specializationExpected == "" || specializationExpected == "runtime.Value" || specializationExpected == "any" {
		if returnGoType != "" && returnGoType != "runtime.Value" && returnGoType != "any" {
			specializationExpected = returnGoType
		}
	}
	var cases []nativeInterfaceGenericDispatchCase
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		impl := g.nativeInterfaceSpecializedGenericMethodImpl(ctx, call, specializationExpected, adapter.GoType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType)
		if impl == nil {
			continue
		}
		cases = append(cases, nativeInterfaceGenericDispatchCase{
			AdapterType: adapter.AdapterType,
			GoType:      adapter.GoType,
			Impl:        impl,
		})
	}
	return cases
}

func (g *generator) sortedNativeInterfaceGenericDispatchKeys() []string {
	if g == nil || len(g.nativeInterfaceGenericDispatches) == 0 {
		return nil
	}
	keys := make([]string, 0, len(g.nativeInterfaceGenericDispatches))
	for key := range g.nativeInterfaceGenericDispatches {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *generator) nativeInterfaceSpecializedGenericMethodImpl(ctx *compileContext, call *ast.FunctionCall, expected string, goType string, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string) *nativeInterfaceAdapterMethod {
	if g == nil || ctx == nil || call == nil || goType == "" || method == nil {
		return nil
	}
	receiverTypeExpr, ok := g.typeExprForGoType(goType)
	if !ok || receiverTypeExpr == nil {
		return nil
	}
	var found *nativeInterfaceAdapterMethod
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || !info.Compileable || impl.ImplName != "" {
			continue
		}
		if impl.MethodName != method.Name || len(info.Params) == 0 {
			continue
		}
		bindings, ok := g.nativeInterfaceGenericImplBindings(impl, method)
		if !ok {
			continue
		}
		if info.Params[0].GoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || len(info.Params) == 0 || info.Params[0].GoType != goType {
				continue
			}
		}
		targetName, _ := typeExprBaseName(impl.TargetType)
		methodInfo := &methodInfo{
			TargetName:   targetName,
			TargetType:   impl.TargetType,
			MethodName:   impl.MethodName,
			ReceiverType: info.Params[0].GoType,
			ExpectsSelf:  true,
			Info:         info,
		}
		bindings, ok = g.specializedImplMethodBindings(ctx, call, methodInfo, impl, receiverTypeExpr, expected)
		if !ok || len(bindings) == 0 {
			continue
		}
		specialized, ok := g.ensureSpecializedImplMethod(methodInfo, impl, bindings)
		if !ok || specialized == nil || specialized.Info == nil || len(specialized.Info.Params) == 0 || specialized.Info.Params[0].GoType != goType {
			continue
		}
		implParamTypeExprs, implParamGoTypes, implReturnTypeExpr, implReturnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok || optionalLast || len(implParamGoTypes) != len(paramGoTypes) {
			continue
		}
		matches := true
		leftVars := make(map[string]string)
		rightVars := make(map[string]string)
		for idx := range paramGoTypes {
			if g.canCoerceStaticExpr(paramGoTypes[idx], implParamGoTypes[idx]) {
				continue
			}
			if !g.typeExprEquivalentModuloGenerics(paramTypeExprs[idx], implParamTypeExprs[idx], leftVars, rightVars) {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		if !g.canCoerceStaticExpr(returnGoType, implReturnGoType) &&
			!g.typeExprEquivalentModuloGenerics(returnTypeExpr, implReturnTypeExpr, leftVars, rightVars) {
			continue
		}
		candidate := &nativeInterfaceAdapterMethod{
			Info:                 specialized.Info,
			CompiledReturnGoType: specialized.Info.ReturnType,
			ParamGoTypes:         implParamGoTypes,
			ReturnGoType:         implReturnGoType,
		}
		for idx := 1; idx < len(specialized.Info.Params); idx++ {
			candidate.CompiledParamGoTypes = append(candidate.CompiledParamGoTypes, specialized.Info.Params[idx].GoType)
		}
		if found != nil && found.Info != candidate.Info {
			if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
				continue
			}
			foundScore := g.nativeInterfaceAdapterMethodSpecificity(found)
			candidateScore := g.nativeInterfaceAdapterMethodSpecificity(candidate)
			switch {
			case candidateScore > foundScore:
				found = candidate
				continue
			case candidateScore < foundScore:
				continue
			default:
				return nil
			}
		}
		found = candidate
	}
	return found
}
