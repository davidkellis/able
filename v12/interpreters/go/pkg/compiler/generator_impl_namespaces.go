package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type namedImplNamespaceInfo struct {
	Package       string
	Name          string
	InterfaceName string
	TargetType    ast.TypeExpression
	TargetKey     string
	Invalid       bool
	Methods       []*implMethodInfo
}

type namedImplMethodOverloadGroup struct {
	MethodName string
	Entries    []*implMethodInfo
	MinArity   int
}

func (g *generator) implDefinitionPackageMap() map[*ast.ImplementationDefinition]string {
	out := make(map[*ast.ImplementationDefinition]string)
	if g == nil {
		return out
	}
	for _, info := range g.implDefinitions {
		if info == nil || info.Definition == nil || strings.TrimSpace(info.Package) == "" {
			continue
		}
		out[info.Definition] = info.Package
	}
	return out
}

func (g *generator) implNamespacePackage(impl *implMethodInfo, pkgByDef map[*ast.ImplementationDefinition]string) string {
	if impl == nil {
		return ""
	}
	if impl.ImplDefinition != nil {
		if pkg, ok := pkgByDef[impl.ImplDefinition]; ok && strings.TrimSpace(pkg) != "" {
			return pkg
		}
	}
	if impl.Info != nil && strings.TrimSpace(impl.Info.Package) != "" {
		return impl.Info.Package
	}
	return ""
}

func (g *generator) namedImplNamespacesByPackage() map[string][]*namedImplNamespaceInfo {
	out := make(map[string][]*namedImplNamespaceInfo)
	if g == nil || len(g.implMethodList) == 0 {
		return out
	}
	pkgByDef := g.implDefinitionPackageMap()
	groups := make(map[string]*namedImplNamespaceInfo)
	for _, impl := range g.sortedImplMethodInfos() {
		if impl == nil || impl.Info == nil || strings.TrimSpace(impl.ImplName) == "" {
			continue
		}
		pkgName := g.implNamespacePackage(impl, pkgByDef)
		if strings.TrimSpace(pkgName) == "" {
			continue
		}
		key := pkgName + "\x00" + impl.ImplName
		group := groups[key]
		if group == nil {
			group = &namedImplNamespaceInfo{
				Package:       pkgName,
				Name:          impl.ImplName,
				InterfaceName: impl.InterfaceName,
				TargetType:    impl.TargetType,
				TargetKey:     typeExpressionToString(impl.TargetType),
				Invalid:       false,
				Methods:       make([]*implMethodInfo, 0, 4),
			}
			groups[key] = group
		} else if group.InterfaceName != impl.InterfaceName || group.TargetKey != typeExpressionToString(impl.TargetType) {
			group.Invalid = true
		}
		group.Methods = append(group.Methods, impl)
	}
	for _, group := range groups {
		if group == nil || strings.TrimSpace(group.Package) == "" {
			continue
		}
		sort.Slice(group.Methods, func(i, j int) bool {
			left := group.Methods[i]
			right := group.Methods[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.MethodName != right.MethodName {
				return left.MethodName < right.MethodName
			}
			if left.Info == nil || right.Info == nil {
				return left.Info != nil
			}
			return left.Info.GoName < right.Info.GoName
		})
		out[group.Package] = append(out[group.Package], group)
	}
	for pkgName := range out {
		sort.Slice(out[pkgName], func(i, j int) bool {
			left := out[pkgName][i]
			right := out[pkgName][j]
			if left == nil || right == nil {
				return left != nil
			}
			return left.Name < right.Name
		})
	}
	return out
}

func (g *generator) sortedNamedImplNamespacesForPackage(pkgName string) []*namedImplNamespaceInfo {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	byPkg := g.namedImplNamespacesByPackage()
	infos := byPkg[pkgName]
	if len(infos) == 0 {
		return nil
	}
	out := make([]*namedImplNamespaceInfo, len(infos))
	copy(out, infos)
	return out
}

func (g *generator) namedImplDispatchParamTypes(method *implMethodInfo) []ast.TypeExpression {
	if method == nil || method.Info == nil || method.Info.Definition == nil {
		return nil
	}
	def := method.Info.Definition
	expectsSelf := methodDefinitionExpectsSelf(def)
	params := methodDefinitionParamTypes(def, method.TargetType, expectsSelf)
	interfaceBindings := make(map[string]ast.TypeExpression)
	for idx, gp := range method.InterfaceGenerics {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			continue
		}
		if idx >= len(method.InterfaceArgs) || method.InterfaceArgs[idx] == nil {
			continue
		}
		interfaceBindings[gp.Name.Name] = method.InterfaceArgs[idx]
	}
	if len(interfaceBindings) == 0 {
		return params
	}
	out := make([]ast.TypeExpression, len(params))
	for idx, paramType := range params {
		out[idx] = substituteTypeParams(paramType, interfaceBindings)
	}
	return out
}

func (g *generator) namedImplMethodOverloadGroups(info *namedImplNamespaceInfo) []*namedImplMethodOverloadGroup {
	if g == nil || info == nil || len(info.Methods) == 0 {
		return nil
	}
	groups := make(map[string]*namedImplMethodOverloadGroup)
	order := make([]string, 0)
	for _, method := range info.Methods {
		if method == nil || strings.TrimSpace(method.MethodName) == "" {
			continue
		}
		group := groups[method.MethodName]
		if group == nil {
			group = &namedImplMethodOverloadGroup{
				MethodName: method.MethodName,
				Entries:    make([]*implMethodInfo, 0, 1),
				MinArity:   -1,
			}
			groups[method.MethodName] = group
			order = append(order, method.MethodName)
		}
		group.Entries = append(group.Entries, method)
		if method.Info != nil && method.Info.Definition != nil {
			paramCount := len(g.namedImplDispatchParamTypes(method))
			minArgs := paramCount
			def := method.Info.Definition
			if len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1]) {
				minArgs--
			}
			if minArgs < 0 {
				minArgs = 0
			}
			if group.MinArity < 0 || minArgs < group.MinArity {
				group.MinArity = minArgs
			}
		}
	}
	sort.Strings(order)
	out := make([]*namedImplMethodOverloadGroup, 0, len(order))
	for _, name := range order {
		group := groups[name]
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		sort.Slice(group.Entries, func(i, j int) bool {
			left := group.Entries[i]
			right := group.Entries[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.Info == nil || right.Info == nil {
				return left.Info != nil
			}
			return left.Info.GoName < right.Info.GoName
		})
		out = append(out, group)
	}
	return out
}

func (g *generator) namedImplNamespacesSeedable() bool {
	if g == nil {
		return false
	}
	byPkg := g.namedImplNamespacesByPackage()
	for _, infos := range byPkg {
		for _, info := range infos {
			if info == nil {
				continue
			}
			if info.Invalid || strings.TrimSpace(info.Package) == "" || strings.TrimSpace(info.Name) == "" || strings.TrimSpace(info.InterfaceName) == "" || info.TargetType == nil {
				return false
			}
			if _, ok := g.renderTypeExpression(info.TargetType); !ok {
				return false
			}
			methods := make(map[string]struct{})
			for _, method := range info.Methods {
				if method == nil || method.Info == nil || !method.Info.Compileable {
					return false
				}
				if strings.TrimSpace(method.MethodName) == "" || method.Info.Definition == nil || strings.TrimSpace(method.Info.GoName) == "" {
					return false
				}
				methods[method.MethodName] = struct{}{}
			}
			if len(methods) == 0 {
				return false
			}
			for _, group := range g.namedImplMethodOverloadGroups(info) {
				if group == nil {
					continue
				}
				for _, method := range group.Entries {
					if method == nil || method.Info == nil || method.Info.Definition == nil {
						return false
					}
					generics := combinedGenericNameSet(method.ImplGenerics, method.InterfaceGenerics, method.Info.Definition.GenericParams)
					for _, paramType := range g.namedImplDispatchParamTypes(method) {
						if paramType == nil || typeExprUsesGeneric(paramType, generics) {
							continue
						}
						if _, ok := g.renderTypeExpression(paramType); !ok {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func (g *generator) renderNamedImplNamespaceSeeds(buf *bytes.Buffer, pkgEnvVar string, pkgName string) {
	if g == nil || buf == nil || strings.TrimSpace(pkgEnvVar) == "" || strings.TrimSpace(pkgName) == "" {
		return
	}
	infos := g.sortedNamedImplNamespacesForPackage(pkgName)
	for idx, info := range infos {
		if info == nil || info.Invalid || strings.TrimSpace(info.Name) == "" || strings.TrimSpace(info.InterfaceName) == "" || info.TargetType == nil {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(info.TargetType)
		if !ok {
			continue
		}
		methodsVar := fmt.Sprintf("__able_impl_methods_%s_%d", sanitizeIdent(info.Name), idx)
		fmt.Fprintf(buf, "\tif _, err := %s.Get(%q); err != nil {\n", pkgEnvVar, info.Name)
		fmt.Fprintf(buf, "\t\t%s := map[string]runtime.Value{}\n", methodsVar)
		methodGroups := g.namedImplMethodOverloadGroups(info)
		for mIdx, group := range methodGroups {
			if group == nil || strings.TrimSpace(group.MethodName) == "" || len(group.Entries) == 0 {
				continue
			}
			methodVar := fmt.Sprintf("__able_impl_method_%s_%d_%d", sanitizeIdent(info.Name), idx, mIdx)
			if len(group.Entries) == 1 {
				method := group.Entries[0]
				if method == nil || method.Info == nil || strings.TrimSpace(method.Info.GoName) == "" {
					continue
				}
				arity := method.Info.Arity
				if arity < 0 {
					arity = -1
				}
				fmt.Fprintf(buf, "\t\t%s := &runtime.NativeFunctionValue{Name: %q, Arity: %d}\n", methodVar, group.MethodName, arity)
				fmt.Fprintf(buf, "\t\t%s.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", methodVar)
				fmt.Fprintf(buf, "\t\t\tif ctx == nil {\n")
				fmt.Fprintf(buf, "\t\t\t\tctx = &runtime.NativeCallContext{Env: %s}\n", pkgEnvVar)
				fmt.Fprintf(buf, "\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", method.Info.GoName)
				fmt.Fprintf(buf, "\t\t}\n")
				fmt.Fprintf(buf, "\t\t%s[%q] = %s\n", methodsVar, group.MethodName, methodVar)
				continue
			}
			displayName := fmt.Sprintf("%s.%s", info.Name, group.MethodName)
			fmt.Fprintf(buf, "\t\t%s := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", methodVar, group.MethodName)
			fmt.Fprintf(buf, "\t\t%s.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", methodVar)
			fmt.Fprintf(buf, "\t\t\tif ctx == nil {\n")
			fmt.Fprintf(buf, "\t\t\t\tctx = &runtime.NativeCallContext{Env: %s}\n", pkgEnvVar)
			fmt.Fprintf(buf, "\t\t\t}\n")
			if group.MinArity > 0 {
				fmt.Fprintf(buf, "\t\t\tif len(args) < %d {\n", group.MinArity)
				fmt.Fprintf(buf, "\t\t\t\targsCopy := append([]runtime.Value{}, args...)\n")
				fmt.Fprintf(buf, "\t\t\t\treturn &runtime.PartialFunctionValue{Target: %s, BoundArgs: argsCopy}, nil\n", methodVar)
				fmt.Fprintf(buf, "\t\t\t}\n")
			}
			fmt.Fprintf(buf, "\t\t\tbestIdx := -1\n")
			fmt.Fprintf(buf, "\t\t\tbestScore := 0.0\n")
			fmt.Fprintf(buf, "\t\t\tties := 0\n")
			fmt.Fprintf(buf, "\t\t\tvar bestArgs []runtime.Value\n")
			for entryIdx, entry := range group.Entries {
				if entry == nil || entry.Info == nil || entry.Info.Definition == nil {
					continue
				}
				def := entry.Info.Definition
				paramTypes := g.namedImplDispatchParamTypes(entry)
				paramCount := len(paramTypes)
				optionalLast := len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1])
				generics := combinedGenericNameSet(entry.ImplGenerics, entry.InterfaceGenerics, def.GenericParams)
				fmt.Fprintf(buf, "\t\t\t{\n")
				fmt.Fprintf(buf, "\t\t\t\tif len(args) == %d", paramCount)
				if optionalLast {
					fmt.Fprintf(buf, " || len(args) == %d", paramCount-1)
				}
				fmt.Fprintf(buf, " {\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\t\t\tmissingOptional := len(args) == %d\n", paramCount-1)
				}
				fmt.Fprintf(buf, "\t\t\t\t\tcompatible := true\n")
				fmt.Fprintf(buf, "\t\t\t\t\tscore := 0.0\n")
				fmt.Fprintf(buf, "\t\t\t\t\tcoercedArgs := make([]runtime.Value, len(args))\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\t\t\tif missingOptional { score -= 0.5 }\n")
				}
				for pIdx, paramType := range paramTypes {
					fmt.Fprintf(buf, "\t\t\t\t\tif len(args) > %d {\n", pIdx)
					if paramType == nil {
						fmt.Fprintf(buf, "\t\t\t\t\t\tcompatible = false\n")
					} else if typeExprUsesGeneric(paramType, generics) {
						fmt.Fprintf(buf, "\t\t\t\t\t\tcoercedArgs[%d] = args[%d]\n", pIdx, pIdx)
					} else if typeExpr, ok := g.renderTypeExpression(paramType); ok {
						spec := parameterSpecificity(paramType, generics)
						fmt.Fprintf(buf, "\t\t\t\t\t\tif compatible {\n")
						fmt.Fprintf(buf, "\t\t\t\t\t\t\tcoerced, ok, err := bridge.MatchType(__able_runtime, %s, args[%d])\n", typeExpr, pIdx)
						fmt.Fprintf(buf, "\t\t\t\t\t\t\tif err != nil { return nil, err }\n")
						fmt.Fprintf(buf, "\t\t\t\t\t\t\tif !ok { compatible = false } else { coercedArgs[%d] = coerced; score += %d }\n", pIdx, spec)
						fmt.Fprintf(buf, "\t\t\t\t\t\t}\n")
					} else {
						fmt.Fprintf(buf, "\t\t\t\t\t\tcompatible = false\n")
					}
					fmt.Fprintf(buf, "\t\t\t\t\t}\n")
				}
				fmt.Fprintf(buf, "\t\t\t\t\tif compatible {\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\t\t\t\tif missingOptional { coercedArgs = append(coercedArgs, runtime.NilValue{}) }\n")
				}
				fmt.Fprintf(buf, "\t\t\t\t\t\tif bestIdx < 0 || score > bestScore {\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t\tbestIdx = %d\n", entryIdx)
				fmt.Fprintf(buf, "\t\t\t\t\t\t\tbestScore = score\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t\tbestArgs = coercedArgs\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t\tties = 1\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t} else if score == bestScore {\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t\tties++\n")
				fmt.Fprintf(buf, "\t\t\t\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t\t}\n")
			}
			fmt.Fprintf(buf, "\t\t\tif bestIdx < 0 {\n")
			fmt.Fprintf(buf, "\t\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", displayName)
			fmt.Fprintf(buf, "\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t\tif ties > 1 {\n")
			fmt.Fprintf(buf, "\t\t\t\treturn nil, fmt.Errorf(\"Ambiguous overload for %s\")\n", displayName)
			fmt.Fprintf(buf, "\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t\tswitch bestIdx {\n")
			for entryIdx, entry := range group.Entries {
				if entry == nil || entry.Info == nil || strings.TrimSpace(entry.Info.GoName) == "" {
					continue
				}
				fmt.Fprintf(buf, "\t\t\tcase %d:\n", entryIdx)
				fmt.Fprintf(buf, "\t\t\t\treturn __able_wrap_%s(__able_runtime, ctx, bestArgs)\n", entry.Info.GoName)
			}
			fmt.Fprintf(buf, "\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", displayName)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\t%s[%q] = %s\n", methodsVar, group.MethodName, methodVar)
		}
		fmt.Fprintf(buf, "\t\t%s.Define(%q, runtime.ImplementationNamespaceValue{Name: ast.NewIdentifier(%q), InterfaceName: ast.NewIdentifier(%q), TargetType: %s, Methods: %s})\n",
			pkgEnvVar, info.Name, info.Name, info.InterfaceName, targetExpr, methodsVar)
		fmt.Fprintf(buf, "\t}\n")
	}
}
