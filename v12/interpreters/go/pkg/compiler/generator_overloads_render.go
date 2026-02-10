package compiler

import (
	"bytes"
	"fmt"
	"sort"
)

func (g *generator) renderOverloadDispatchers(buf *bytes.Buffer) {
	if g == nil {
		return
	}
	if len(g.overloads) == 0 {
		g.renderMethodOverloadDispatchers(buf)
		return
	}
	packages := g.packages
	if len(packages) == 0 {
		for pkg := range g.overloads {
			packages = append(packages, pkg)
		}
		sort.Strings(packages)
	}
	for _, pkgName := range packages {
		pkgOverloads := g.overloads[pkgName]
		if len(pkgOverloads) == 0 {
			continue
		}
		names := make([]string, 0, len(pkgOverloads))
		for name := range pkgOverloads {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			overload := pkgOverloads[name]
			if overload == nil || len(overload.Entries) == 0 {
				continue
			}
			wrapperName := g.overloadWrapperName(pkgName, name)
			callName := g.overloadCallName(pkgName, name)
			valueName := g.overloadValueName(pkgName, name)
			fmt.Fprintf(buf, "var %s *runtime.NativeFunctionValue\n\n", valueName)
			fmt.Fprintf(buf, "func %s(self *runtime.NativeFunctionValue, ctx *runtime.NativeCallContext, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, error) {\n", wrapperName)
			fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"compiler: missing runtime\")\n")
			fmt.Fprintf(buf, "\t}\n")
			if overload.MinArity > 0 {
				fmt.Fprintf(buf, "\tif len(args) < %d {\n", overload.MinArity)
				fmt.Fprintf(buf, "\t\targsCopy := append([]runtime.Value{}, args...)\n")
				fmt.Fprintf(buf, "\t\treturn &runtime.PartialFunctionValue{Target: self, BoundArgs: argsCopy}, nil\n")
				fmt.Fprintf(buf, "\t}\n")
			}
			fmt.Fprintf(buf, "\tbestIdx := -1\n")
			fmt.Fprintf(buf, "\tbestScore := 0.0\n")
			fmt.Fprintf(buf, "\tties := 0\n")
			fmt.Fprintf(buf, "\tvar bestArgs []runtime.Value\n")
			for idx, entry := range overload.Entries {
				if entry == nil || entry.Definition == nil {
					continue
				}
				def := entry.Definition
				paramCount := len(def.Params)
				optionalLast := paramCount > 0 && isNullableParam(def.Params[paramCount-1])
				fmt.Fprintf(buf, "\t{\n")
				fmt.Fprintf(buf, "\t\tif len(args) == %d", paramCount)
				if optionalLast {
					fmt.Fprintf(buf, " || len(args) == %d", paramCount-1)
				}
				fmt.Fprintf(buf, " {\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\tmissingOptional := len(args) == %d\n", paramCount-1)
				}
				fmt.Fprintf(buf, "\t\t\tcompatible := true\n")
				fmt.Fprintf(buf, "\t\t\tscore := 0.0\n")
				fmt.Fprintf(buf, "\t\t\tcoercedArgs := make([]runtime.Value, len(args))\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\tif missingOptional { score -= 0.5 }\n")
				}
				generics := genericNameSet(def.GenericParams)
				for pIdx, param := range def.Params {
					fmt.Fprintf(buf, "\t\t\tif len(args) > %d {\n", pIdx)
					if param == nil {
						fmt.Fprintf(buf, "\t\t\t\tcompatible = false\n")
					} else if param.ParamType != nil && !typeExprUsesGeneric(param.ParamType, generics) {
						typeExpr, ok := g.renderTypeExpression(param.ParamType)
						if !ok {
							fmt.Fprintf(buf, "\t\t\t\tcompatible = false\n")
						} else {
							spec := parameterSpecificity(param.ParamType, generics)
							fmt.Fprintf(buf, "\t\t\t\tif compatible {\n")
							fmt.Fprintf(buf, "\t\t\t\t\tcoerced, ok, err := bridge.MatchType(__able_runtime, %s, args[%d])\n", typeExpr, pIdx)
							fmt.Fprintf(buf, "\t\t\t\t\tif err != nil { return nil, err }\n")
							fmt.Fprintf(buf, "\t\t\t\t\tif !ok { compatible = false } else { coercedArgs[%d] = coerced; score += %d }\n", pIdx, spec)
							fmt.Fprintf(buf, "\t\t\t\t}\n")
						}
					} else {
						fmt.Fprintf(buf, "\t\t\t\tcoercedArgs[%d] = args[%d]\n", pIdx, pIdx)
					}
					fmt.Fprintf(buf, "\t\t\t}\n")
				}
				fmt.Fprintf(buf, "\t\t\tif compatible {\n")
				if optionalLast {
					fmt.Fprintf(buf, "\t\t\t\tif missingOptional { coercedArgs = append(coercedArgs, runtime.NilValue{}) }\n")
				}
				fmt.Fprintf(buf, "\t\t\t\tif bestIdx < 0 || score > bestScore {\n")
				fmt.Fprintf(buf, "\t\t\t\t\tbestIdx = %d\n", idx)
				fmt.Fprintf(buf, "\t\t\t\t\tbestScore = score\n")
				fmt.Fprintf(buf, "\t\t\t\t\tbestArgs = coercedArgs\n")
				fmt.Fprintf(buf, "\t\t\t\t\tties = 1\n")
				fmt.Fprintf(buf, "\t\t\t\t} else if score == bestScore {\n")
				fmt.Fprintf(buf, "\t\t\t\t\tties++\n")
				fmt.Fprintf(buf, "\t\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t\t}\n")
				fmt.Fprintf(buf, "\t\t}\n")
				fmt.Fprintf(buf, "\t}\n")
			}
			fmt.Fprintf(buf, "\tif bestIdx < 0 {\n")
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", name)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\tif ties > 1 {\n")
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"Ambiguous overload for %s\")\n", name)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\tswitch bestIdx {\n")
			for idx, entry := range overload.Entries {
				if entry == nil {
					continue
				}
				fmt.Fprintf(buf, "\tcase %d:\n", idx)
				fmt.Fprintf(buf, "\t\treturn __able_wrap_%s(__able_runtime, ctx, bestArgs)\n", entry.GoName)
			}
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", name)
			fmt.Fprintf(buf, "}\n\n")

			fmt.Fprintf(buf, "func %s(args []runtime.Value, call *ast.FunctionCall) runtime.Value {\n", callName)
			fmt.Fprintf(buf, "\tval, err := %s(%s, nil, args, call)\n", wrapperName, valueName)
			fmt.Fprintf(buf, "\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\tif call != nil && __able_runtime != nil {\n")
			fmt.Fprintf(buf, "\t\t\tbridge.RaiseRuntimeErrorWithContext(__able_runtime, call, err)\n")
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\t__able_panic_on_error(err)\n")
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\tif val == nil {\n")
			fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\treturn val\n")
			fmt.Fprintf(buf, "}\n\n")
		}
	}
	g.renderMethodOverloadDispatchers(buf)
}

func (g *generator) renderMethodOverloadDispatchers(buf *bytes.Buffer) {
	groups := g.methodOverloadGroups()
	if len(groups) == 0 {
		return
	}
	for _, group := range groups {
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		wrapperName := g.methodOverloadWrapperName(group.TargetName, group.MethodName, group.ExpectsSelf)
		valueName := g.methodOverloadValueName(group.TargetName, group.MethodName, group.ExpectsSelf)
		displayName := fmt.Sprintf("%s.%s", group.TargetName, group.MethodName)
		fmt.Fprintf(buf, "var %s *runtime.NativeFunctionValue\n\n", valueName)
		fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", wrapperName)
		fmt.Fprintf(buf, "\tif rt == nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"compiler: missing runtime\")\n")
		fmt.Fprintf(buf, "\t}\n")
		if group.MinArity > 0 {
			fmt.Fprintf(buf, "\tif len(args) < %d {\n", group.MinArity)
			fmt.Fprintf(buf, "\t\tif %s == nil {\n", valueName)
			fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"compiler: missing overload value for %s\")\n", displayName)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\targsCopy := append([]runtime.Value{}, args...)\n")
			fmt.Fprintf(buf, "\t\treturn &runtime.PartialFunctionValue{Target: %s, BoundArgs: argsCopy}, nil\n", valueName)
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\tbestIdx := -1\n")
		fmt.Fprintf(buf, "\tbestScore := 0.0\n")
		fmt.Fprintf(buf, "\tties := 0\n")
		fmt.Fprintf(buf, "\tvar bestArgs []runtime.Value\n")
		for idx, method := range group.Entries {
			if method == nil || method.Info == nil || method.Info.Definition == nil {
				continue
			}
			def := method.Info.Definition
			paramTypes := methodDefinitionParamTypes(def, method.TargetType, method.ExpectsSelf)
			paramCount := len(paramTypes)
			optionalLast := len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1])
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\tif len(args) == %d", paramCount)
			if optionalLast {
				fmt.Fprintf(buf, " || len(args) == %d", paramCount-1)
			}
			fmt.Fprintf(buf, " {\n")
			if optionalLast {
				fmt.Fprintf(buf, "\t\t\tmissingOptional := len(args) == %d\n", paramCount-1)
			}
			fmt.Fprintf(buf, "\t\t\tcompatible := true\n")
			fmt.Fprintf(buf, "\t\t\tscore := 0.0\n")
			fmt.Fprintf(buf, "\t\t\tcoercedArgs := make([]runtime.Value, len(args))\n")
			if optionalLast {
				fmt.Fprintf(buf, "\t\t\tif missingOptional { score -= 0.5 }\n")
			}
			generics := g.methodGenericNames(method)
			for pIdx, paramType := range paramTypes {
				fmt.Fprintf(buf, "\t\t\tif len(args) > %d {\n", pIdx)
				if paramType != nil && !typeExprUsesGeneric(paramType, generics) {
					typeExpr, ok := g.renderTypeExpression(paramType)
					if !ok {
						fmt.Fprintf(buf, "\t\t\t\tcompatible = false\n")
					} else {
						spec := parameterSpecificity(paramType, generics)
						fmt.Fprintf(buf, "\t\t\t\tif compatible {\n")
						fmt.Fprintf(buf, "\t\t\t\t\tcoerced, ok, err := bridge.MatchType(rt, %s, args[%d])\n", typeExpr, pIdx)
						fmt.Fprintf(buf, "\t\t\t\t\tif err != nil { return nil, err }\n")
						fmt.Fprintf(buf, "\t\t\t\t\tif !ok { compatible = false } else { coercedArgs[%d] = coerced; score += %d }\n", pIdx, spec)
						fmt.Fprintf(buf, "\t\t\t\t}\n")
					}
				} else {
					fmt.Fprintf(buf, "\t\t\t\tcoercedArgs[%d] = args[%d]\n", pIdx, pIdx)
				}
				fmt.Fprintf(buf, "\t\t\t}\n")
			}
			fmt.Fprintf(buf, "\t\t\tif compatible {\n")
			if optionalLast {
				fmt.Fprintf(buf, "\t\t\t\tif missingOptional { coercedArgs = append(coercedArgs, runtime.NilValue{}) }\n")
			}
			fmt.Fprintf(buf, "\t\t\t\tif bestIdx < 0 || score > bestScore {\n")
			fmt.Fprintf(buf, "\t\t\t\t\tbestIdx = %d\n", idx)
			fmt.Fprintf(buf, "\t\t\t\t\tbestScore = score\n")
			fmt.Fprintf(buf, "\t\t\t\t\tbestArgs = coercedArgs\n")
			fmt.Fprintf(buf, "\t\t\t\t\tties = 1\n")
			fmt.Fprintf(buf, "\t\t\t\t} else if score == bestScore {\n")
			fmt.Fprintf(buf, "\t\t\t\t\tties++\n")
			fmt.Fprintf(buf, "\t\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\tif bestIdx < 0 {\n")
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", displayName)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif ties > 1 {\n")
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"Ambiguous overload for %s\")\n", displayName)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tswitch bestIdx {\n")
		for idx, method := range group.Entries {
			if method == nil || method.Info == nil {
				continue
			}
			fmt.Fprintf(buf, "\tcase %d:\n", idx)
			fmt.Fprintf(buf, "\t\treturn __able_wrap_%s(rt, ctx, bestArgs)\n", method.Info.GoName)
		}
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", displayName)
		fmt.Fprintf(buf, "}\n\n")
	}
}
