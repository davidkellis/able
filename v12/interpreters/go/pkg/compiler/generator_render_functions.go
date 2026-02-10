package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderCompiledFunctions(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable {
			continue
		}
		ctx := newCompileContext(info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package)
		lines, retExpr, ok := g.compileBody(ctx, info)
		if !ok {
			if info.Reason == "" {
				reason := ctx.reason
				if reason == "" {
					reason = "unsupported function body"
				}
				info.Reason = reason
			}
			info.Compileable = false
			continue
		}
		fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
		}
		resultName := "__able_result"
		fmt.Fprintf(buf, ") (%s %s) {\n", resultName, info.ReturnType)
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			fmt.Fprintf(buf, "\tif __able_runtime != nil && %s != nil {\n", envVar)
			fmt.Fprintf(buf, "\t\tprevEnv := __able_runtime.SwapEnv(%s)\n", envVar)
			fmt.Fprintf(buf, "\t\tdefer __able_runtime.SwapEnv(prevEnv)\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		recoverValue := fmt.Sprintf("val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", info.ReturnType, resultName)
		if info.ReturnType == "runtime.Value" {
			recoverValue = fmt.Sprintf("if ret.value == nil { %s = runtime.NilValue{}; return }; val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", resultName, info.ReturnType, resultName)
		}
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tif ret, ok := recovered.(__able_return); ok {\n")
		fmt.Fprintf(buf, "\t\t\t\t%s\n", recoverValue)
		fmt.Fprintf(buf, "\t\t\t\treturn\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tpanic(recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		for _, line := range lines {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
		fmt.Fprintf(buf, "\treturn %s\n", retExpr)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderCompiledMethods(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil || !method.Info.Compileable {
			continue
		}
		info := method.Info
		ctx := newCompileContext(info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package)
		lines, retExpr, ok := g.compileBody(ctx, info)
		if !ok {
			if info.Reason == "" {
				reason := ctx.reason
				if reason == "" {
					reason = "unsupported method body"
				}
				info.Reason = reason
			}
			info.Compileable = false
			continue
		}
		fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
		}
		resultName := "__able_result"
		fmt.Fprintf(buf, ") (%s %s) {\n", resultName, info.ReturnType)
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			fmt.Fprintf(buf, "\tif __able_runtime != nil && %s != nil {\n", envVar)
			fmt.Fprintf(buf, "\t\tprevEnv := __able_runtime.SwapEnv(%s)\n", envVar)
			fmt.Fprintf(buf, "\t\tdefer __able_runtime.SwapEnv(prevEnv)\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		recoverValue := fmt.Sprintf("val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", info.ReturnType, resultName)
		if info.ReturnType == "runtime.Value" {
			recoverValue = fmt.Sprintf("if ret.value == nil { %s = runtime.NilValue{}; return }; val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", resultName, info.ReturnType, resultName)
		}
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tif ret, ok := recovered.(__able_return); ok {\n")
		fmt.Fprintf(buf, "\t\t\t\t%s\n", recoverValue)
		fmt.Fprintf(buf, "\t\t\t\treturn\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tpanic(recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		for _, line := range lines {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
		fmt.Fprintf(buf, "\treturn %s\n", retExpr)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderWrappers(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil {
			continue
		}
		if !info.Compileable && !info.HasOriginal {
			continue
		}
		genericNames := g.callableGenericNames(info)
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tresult = nil\n")
		fmt.Fprintf(buf, "\t\t\terr = bridge.Recover(rt, ctx, recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		fmt.Fprintf(buf, "\tif rt != nil && ctx != nil && ctx.Env != nil {\n")
		fmt.Fprintf(buf, "\t\tprevEnv := rt.SwapEnv(ctx.Env)\n")
		fmt.Fprintf(buf, "\t\tdefer rt.SwapEnv(prevEnv)\n")
		fmt.Fprintf(buf, "\t}\n")
		if g.hasOptionalLastParam(info) && info.Arity > 0 {
			fmt.Fprintf(buf, "\tif len(args) == %d {\n", info.Arity-1)
			fmt.Fprintf(buf, "\t\targs = append(args, runtime.NilValue{})\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		if info.Compileable {
			fmt.Fprintf(buf, "\tif len(args) != %d {\n", info.Arity)
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"arity mismatch calling %s: expected %d, got %%d\", len(args))\n", info.Name, info.Arity)
			fmt.Fprintf(buf, "\t}\n")
			for idx, param := range info.Params {
				argName := fmt.Sprintf("arg%d", idx)
				fmt.Fprintf(buf, "\t%sValue := args[%d]\n", argName, idx)
				g.renderArgConversion(buf, argName, param, info.Name, genericNames)
			}
			fmt.Fprintf(buf, "\tcompiledResult := __able_compiled_%s(", info.GoName)
			for i, param := range info.Params {
				if i > 0 {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%s", param.GoName)
			}
			fmt.Fprintf(buf, ")\n")
			g.renderReturnConversion(buf, "compiledResult", info.ReturnType, info.Definition.ReturnType, info.Name, genericNames)
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		if info.HasOriginal {
			qualified := info.Name
			if info.QualifiedName != "" {
				qualified = info.QualifiedName
			}
			fmt.Fprintf(buf, "\treturn rt.CallOriginal(%q, args)\n", qualified)
		} else {
			fmt.Fprintf(buf, "\treturn nil, fmt.Errorf(\"compiler: missing compiled implementation for %s\")\n", info.Name)
		}
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMethodWrappers(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		info := method.Info
		genericNames := g.methodGenericNames(method)
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tresult = nil\n")
		fmt.Fprintf(buf, "\t\t\terr = bridge.Recover(rt, ctx, recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		fmt.Fprintf(buf, "\tif rt != nil && ctx != nil && ctx.Env != nil {\n")
		fmt.Fprintf(buf, "\t\tprevEnv := rt.SwapEnv(ctx.Env)\n")
		fmt.Fprintf(buf, "\t\tdefer rt.SwapEnv(prevEnv)\n")
		fmt.Fprintf(buf, "\t}\n")
		if g.hasOptionalLastParam(info) && info.Arity > 0 {
			fmt.Fprintf(buf, "\tif len(args) == %d {\n", info.Arity-1)
			fmt.Fprintf(buf, "\t\targs = append(args, runtime.NilValue{})\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\tif len(args) != %d {\n", info.Arity)
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"arity mismatch calling %s: expected %d, got %%d\", len(args))\n", info.Name, info.Arity)
		fmt.Fprintf(buf, "\t}\n")
		for idx, param := range info.Params {
			argName := fmt.Sprintf("arg%d", idx)
			fmt.Fprintf(buf, "\t%sValue := args[%d]\n", argName, idx)
			g.renderArgConversion(buf, argName, param, info.Name, genericNames)
		}
		fmt.Fprintf(buf, "\tcompiledResult := __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s", param.GoName)
		}
		fmt.Fprintf(buf, ")\n")
		if method.ExpectsSelf && len(info.Params) > 0 {
			recv := info.Params[0]
			if g.typeCategory(recv.GoType) == "struct" {
				baseName, ok := g.structBaseName(recv.GoType)
				if !ok {
					baseName = strings.TrimPrefix(recv.GoType, "*")
				}
				fmt.Fprintf(buf, "\tif err := __able_struct_%s_apply(rt, arg0Value, %s); err != nil {\n", baseName, recv.GoName)
				fmt.Fprintf(buf, "\t\treturn nil, err\n")
				fmt.Fprintf(buf, "\t}\n")
			}
		}
		g.renderReturnConversion(buf, "compiledResult", info.ReturnType, info.Definition.ReturnType, info.Name, genericNames)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderFunctionThunks(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable {
			continue
		}
		fmt.Fprintf(buf, "func __able_function_thunk_%s(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tctx := &runtime.NativeCallContext{Env: env}\n")
		fmt.Fprintf(buf, "\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMethodThunks(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		info := method.Info
		fmt.Fprintf(buf, "func __able_method_thunk_%s(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tctx := &runtime.NativeCallContext{Env: env}\n")
		fmt.Fprintf(buf, "\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderRegister(buf *bytes.Buffer) {
	methodCounts := make(map[string]int)
	methodOverloads := g.methodOverloadGroups()
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		key := fmt.Sprintf("%s|%s|%t", method.TargetName, method.MethodName, method.ExpectsSelf)
		methodCounts[key]++
	}
	fmt.Fprintf(buf, "func Register(interp *interpreter.Interpreter) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\treturn RegisterIn(interp, nil)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RegisterIn(interp *interpreter.Interpreter, env *runtime.Environment) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\tif interp == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing interpreter\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tentryEnv := env\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(buf, "\t\tentryEnv = interp.GlobalEnvironment()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing environment\")\n")
	fmt.Fprintf(buf, "\t}\n")
	if g.entryPackage != "" {
		fmt.Fprintf(buf, "\tif entryEnv == interp.GlobalEnvironment() {\n")
		fmt.Fprintf(buf, "\t\tif pkgEnv := interp.PackageEnvironment(%q); pkgEnv != nil {\n", g.entryPackage)
		fmt.Fprintf(buf, "\t\t\tentryEnv = pkgEnv\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\trt := bridge.New(interp)\n")
	fmt.Fprintf(buf, "\t__able_runtime = rt\n")
	fmt.Fprintf(buf, "\trt.SetEnv(entryEnv)\n")
	if envVar, ok := g.packageEnvVar(g.entryPackage); ok {
		fmt.Fprintf(buf, "\t%s = entryEnv\n", envVar)
	}
	if len(g.diagNodes) > 0 {
		fmt.Fprintf(buf, "\t__able_register_diag_nodes()\n")
	}
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(method.TargetType)
		if !ok {
			return
		}
		paramExprs, ok := g.renderMethodParamTypes(method)
		if !ok {
			return
		}
		fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledMethodOverload(%q, %q, %t, %s, %s, __able_method_thunk_%s); err != nil {\n", method.TargetName, method.MethodName, method.ExpectsSelf, targetExpr, paramExprs, method.Info.GoName)
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
		key := fmt.Sprintf("%s|%s|%t", method.TargetName, method.MethodName, method.ExpectsSelf)
		if methodCounts[key] == 1 && method.Info.Arity >= 0 {
			arity := method.Info.Arity
			minArgs := arity
			if g.hasOptionalLastParam(method.Info) && arity > 0 {
				minArgs = arity - 1
			}
			if method.ExpectsSelf {
				arity--
				minArgs--
			}
			if arity < 0 {
				arity = 0
			}
			if minArgs < 0 {
				minArgs = 0
			}
			fmt.Fprintf(buf, "\t__able_register_compiled_method(%q, %q, %t, %d, %d, __able_wrap_%s)\n", method.TargetName, method.MethodName, method.ExpectsSelf, arity, minArgs, method.Info.GoName)
		}
	}
	implEnvIndex := 0
	for _, implMethod := range g.sortedImplMethodInfos() {
		if implMethod == nil || implMethod.Info == nil || !implMethod.Info.Compileable {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(implMethod.TargetType)
		if !ok {
			return
		}
		ifaceArgsExpr, ok := g.renderTypeExpressionList(implMethod.InterfaceArgs)
		if !ok {
			return
		}
		paramExprs, ok := g.renderImplMethodParamTypes(implMethod)
		if !ok {
			return
		}
		constraintKey := constraintSignature(collectConstraintSpecs(implMethod.ImplGenerics, implMethod.WhereClause))
		if implMethod.ImplName != "" {
			envExpr := "entryEnv"
			if implMethod.Info.Package != g.entryPackage {
				envExpr = fmt.Sprintf("implEnv%d", implEnvIndex)
				implEnvIndex++
				fmt.Fprintf(buf, "\t%s := interp.PackageEnvironment(%q)\n", envExpr, implMethod.Info.Package)
				fmt.Fprintf(buf, "\tif %s == nil {\n", envExpr)
				fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing environment for package %s\")\n", implMethod.Info.Package)
				fmt.Fprintf(buf, "\t}\n")
			}
			fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledImplNamespaceMethod(%s, %q, %q, %s, __able_function_thunk_%s); err != nil {\n",
				envExpr, implMethod.ImplName, implMethod.MethodName, paramExprs, implMethod.Info.GoName)
			fmt.Fprintf(buf, "\t\treturn nil, err\n")
			fmt.Fprintf(buf, "\t}\n")
			continue
		}
		fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledImplMethodOverload(%q, %s, %s, %q, %q, %q, %s, __able_function_thunk_%s); err != nil {\n",
			implMethod.InterfaceName, targetExpr, ifaceArgsExpr, constraintKey, implMethod.ImplName, implMethod.MethodName, paramExprs, implMethod.Info.GoName)
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	for _, group := range g.interfaceDispatchGroups() {
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		entryInfo := group.Entries[0]
		if entryInfo == nil || entryInfo.Info == nil {
			continue
		}
		info := entryInfo.Info
		arity := info.Arity
		if methodDefinitionExpectsSelf(info.Definition) && arity > 0 {
			arity--
		}
		if len(group.Entries) == 1 {
			if arity < 0 {
				arity = -1
			}
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\tfn := &runtime.NativeFunctionValue{Name: %q, Arity: %d}\n", entryInfo.MethodName, arity)
			fmt.Fprintf(buf, "\t\tfn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			fmt.Fprintf(buf, "\t\t\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\t__able_register_interface_dispatch(%q, %q, %s, %s, %s, %s, fn, %t, false)\n", group.InterfaceName, group.MethodName, group.TargetExpr, group.InterfaceArgs, group.GenericNames, group.Constraints, group.IsPrivate)
			fmt.Fprintf(buf, "\t}\n")
			continue
		}
		minArgs := -1
		for _, entry := range group.Entries {
			if entry == nil || entry.Info == nil || entry.Info.Definition == nil {
				continue
			}
			def := entry.Info.Definition
			expectsSelf := methodDefinitionExpectsSelf(def)
			paramTypes := methodDefinitionParamTypes(def, entry.TargetType, expectsSelf)
			paramCount := len(paramTypes)
			entryMin := paramCount
			if len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1]) {
				entryMin--
			}
			if entryMin < 0 {
				entryMin = 0
			}
			if minArgs < 0 || entryMin < minArgs {
				minArgs = entryMin
			}
		}
		fmt.Fprintf(buf, "\t{\n")
		fmt.Fprintf(buf, "\t\tfn := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", group.MethodName)
		fmt.Fprintf(buf, "\t\tfn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
		if minArgs > 0 {
			fmt.Fprintf(buf, "\t\t\tif len(args) < %d {\n", minArgs)
			fmt.Fprintf(buf, "\t\t\t\targsCopy := append([]runtime.Value{}, args...)\n")
			fmt.Fprintf(buf, "\t\t\t\treturn runtime.PartialFunctionValue{Target: fn, BoundArgs: argsCopy}, nil\n")
			fmt.Fprintf(buf, "\t\t\t}\n")
		}
		fmt.Fprintf(buf, "\t\t\tbestIdx := -1\n")
		fmt.Fprintf(buf, "\t\t\tbestScore := 0.0\n")
		fmt.Fprintf(buf, "\t\t\tties := 0\n")
		fmt.Fprintf(buf, "\t\t\tvar bestArgs []runtime.Value\n")
		for idx, entry := range group.Entries {
			if entry == nil || entry.Info == nil || entry.Info.Definition == nil {
				continue
			}
			def := entry.Info.Definition
			expectsSelf := methodDefinitionExpectsSelf(def)
			paramTypes := methodDefinitionParamTypes(def, entry.TargetType, expectsSelf)
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
			fmt.Fprintf(buf, "\t\t\t\t\t\t\tbestIdx = %d\n", idx)
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
		fmt.Fprintf(buf, "\t\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", group.MethodName)
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tif ties > 1 {\n")
		fmt.Fprintf(buf, "\t\t\t\treturn nil, fmt.Errorf(\"Ambiguous overload for %s\")\n", group.MethodName)
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tswitch bestIdx {\n")
		for idx, entry := range group.Entries {
			if entry == nil || entry.Info == nil {
				continue
			}
			fmt.Fprintf(buf, "\t\t\tcase %d:\n", idx)
			fmt.Fprintf(buf, "\t\t\t\treturn __able_wrap_%s(__able_runtime, ctx, bestArgs)\n", entry.Info.GoName)
		}
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", group.MethodName)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t__able_register_interface_dispatch(%q, %q, %s, %s, %s, %s, fn, %t, false)\n", group.InterfaceName, group.MethodName, group.TargetExpr, group.InterfaceArgs, group.GenericNames, group.Constraints, group.IsPrivate)
		fmt.Fprintf(buf, "\t}\n")
	}
	if g.interfaceDispatchStrict() {
		fmt.Fprintf(buf, "\t__able_interface_dispatch_strict = true\n")
	}
	for _, group := range methodOverloads {
		if group == nil {
			continue
		}
		minArgs := group.MinArity
		if minArgs < 0 {
			minArgs = 0
		}
		if group.ExpectsSelf {
			minArgs--
		}
		if minArgs < 0 {
			minArgs = 0
		}
		wrapperName := g.methodOverloadWrapperName(group.TargetName, group.MethodName, group.ExpectsSelf)
		valueName := g.methodOverloadValueName(group.TargetName, group.MethodName, group.ExpectsSelf)
		fmt.Fprintf(buf, "\t__able_register_compiled_method(%q, %q, %t, -1, %d, %s)\n", group.TargetName, group.MethodName, group.ExpectsSelf, minArgs, wrapperName)
		fmt.Fprintf(buf, "\tif entry := __able_lookup_compiled_method(%q, %q, %t); entry != nil {\n", group.TargetName, group.MethodName, group.ExpectsSelf)
		fmt.Fprintf(buf, "\t\t%s = entry.fn\n", valueName)
		fmt.Fprintf(buf, "\t}\n")
	}
	packageList := g.packages
	if len(packageList) == 0 {
		for pkgName := range g.functions {
			packageList = append(packageList, pkgName)
		}
		for pkgName := range g.overloads {
			packageList = append(packageList, pkgName)
		}
		sort.Strings(packageList)
	}
	pkgIndex := 0
	for _, pkgName := range packageList {
		pkgEnvVar := "entryEnv"
		if pkgName != g.entryPackage {
			pkgEnvVar = fmt.Sprintf("pkgEnv%d", pkgIndex)
			pkgIndex++
			fmt.Fprintf(buf, "\t%s := interp.PackageEnvironment(%q)\n", pkgEnvVar, pkgName)
			fmt.Fprintf(buf, "\tif %s == nil {\n", pkgEnvVar)
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing environment for package %s\")\n", pkgName)
			fmt.Fprintf(buf, "\t}\n")
		}
		if envVar, ok := g.packageEnvVar(pkgName); ok {
			fmt.Fprintf(buf, "\t%s = %s\n", envVar, pkgEnvVar)
		}
		for _, name := range g.sortedCallableNames(pkgName) {
			if overload, ok := g.overloads[pkgName][name]; ok && overload != nil {
				qualified := overload.QualifiedName
				if qualified == "" {
					qualified = qualifiedName(pkgName, name)
				}
				fmt.Fprintf(buf, "\tif original, err := %s.Get(%q); err == nil {\n", pkgEnvVar, name)
				fmt.Fprintf(buf, "\t\trt.RegisterOriginal(%q, original)\n", qualified)
				fmt.Fprintf(buf, "\t}\n")
				fmt.Fprintf(buf, "\t{\n")
				fmt.Fprintf(buf, "\t\toverloadFn := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", name)
				fmt.Fprintf(buf, "\t\toverloadFn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
				fmt.Fprintf(buf, "\t\t\treturn %s(overloadFn, ctx, args, nil)\n", g.overloadWrapperName(pkgName, name))
				fmt.Fprintf(buf, "\t\t}\n")
				fmt.Fprintf(buf, "\t\t%s = overloadFn\n", g.overloadValueName(pkgName, name))
				fmt.Fprintf(buf, "\t}\n")
				fmt.Fprintf(buf, "\t__able_register_compiled_call(%s, %q, -1, %d, func(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", pkgEnvVar, name, overload.MinArity)
				fmt.Fprintf(buf, "\t\treturn %s(%s, ctx, args, nil)\n", g.overloadWrapperName(pkgName, name), g.overloadValueName(pkgName, name))
				fmt.Fprintf(buf, "\t})\n")
				for _, entry := range overload.Entries {
					if entry == nil || !entry.Compileable {
						continue
					}
					paramExprs, ok := g.renderFunctionParamTypes(entry)
					if !ok {
						return
					}
					fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledFunctionOverload(%s, %q, %s, __able_function_thunk_%s); err != nil {\n", pkgEnvVar, name, paramExprs, entry.GoName)
					fmt.Fprintf(buf, "\t\treturn nil, err\n")
					fmt.Fprintf(buf, "\t}\n")
				}
				continue
			}
			info := g.functions[pkgName][name]
			if info == nil {
				continue
			}
			qualified := info.QualifiedName
			if qualified == "" {
				qualified = qualifiedName(pkgName, info.Name)
			}
			fmt.Fprintf(buf, "\tif original, err := %s.Get(%q); err == nil {\n", pkgEnvVar, info.Name)
			fmt.Fprintf(buf, "\t\trt.RegisterOriginal(%q, original)\n", qualified)
			fmt.Fprintf(buf, "\t}\n")
			if info.Compileable {
				paramExprs, ok := g.renderFunctionParamTypes(info)
				if !ok {
					return
				}
				fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledFunctionOverload(%s, %q, %s, __able_function_thunk_%s); err != nil {\n", pkgEnvVar, info.Name, paramExprs, info.GoName)
				fmt.Fprintf(buf, "\t\treturn nil, err\n")
				fmt.Fprintf(buf, "\t}\n")
			}
			if info.Arity >= 0 {
				minArgs := info.Arity
				if g.hasOptionalLastParam(info) && info.Arity > 0 {
					minArgs = info.Arity - 1
				}
				fmt.Fprintf(buf, "\t__able_register_compiled_call(%s, %q, %d, %d, __able_wrap_%s)\n", pkgEnvVar, info.Name, info.Arity, minArgs, info.GoName)
			}
		}
	}
	fmt.Fprintf(buf, "\treturn rt, nil\n")
	fmt.Fprintf(buf, "}\n")
}

func (g *generator) renderArgConversion(buf *bytes.Buffer, argName string, param paramInfo, funcName string, genericNames map[string]struct{}) {
	goType := param.GoType
	target := param.GoName
	if g.typeCategory(goType) == "runtime" {
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
			if !g.typeExprHasGeneric(ifaceType, genericNames) {
				rendered, ok := g.renderTypeExpression(ifaceType)
				if ok {
					okName := fmt.Sprintf("%sOk", argName)
					expected := typeExpressionToString(ifaceType)
					fmt.Fprintf(buf, "\t%s, %s, err := bridge.MatchType(rt, %s, %sValue)\n", target, okName, rendered, argName)
					fmt.Fprintf(buf, "\tif err != nil {\n")
					fmt.Fprintf(buf, "\t\treturn nil, err\n")
					fmt.Fprintf(buf, "\t}\n")
					fmt.Fprintf(buf, "\tif !%s {\n", okName)
					fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch calling %s: expected %s\")\n", funcName, expected)
					fmt.Fprintf(buf, "\t}\n")
					return
				}
			}
		}
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	case "bool":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsBool(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "string":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsString(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "rune":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsRune(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "float32":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsFloat(%sValue)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := float32(%sRaw)\n", target, argName)
	case "float64":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsFloat(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "int":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsInt(%sValue, bridge.NativeIntBits)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := int(%sRaw)\n", target, argName)
	case "uint":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsUint(%sValue, bridge.NativeIntBits)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := uint(%sRaw)\n", target, argName)
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsInt(%sValue, %d)\n", argName, argName, bits)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := %s(%sRaw)\n", target, goType, argName)
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsUint(%sValue, %d)\n", argName, argName, bits)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := %s(%sRaw)\n", target, goType, argName)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\t%s, err := __able_struct_%s_from(%sValue)\n", target, baseName, argName)
		g.renderConvertErr(buf)
	default:
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	}
}

func (g *generator) renderReturnConversion(buf *bytes.Buffer, resultName, goType string, returnType ast.TypeExpression, funcName string, genericNames map[string]struct{}) {
	if g.typeCategory(goType) == "runtime" {
		if ifaceType, ok := g.interfaceTypeExpr(returnType); ok {
			if !g.typeExprHasGeneric(ifaceType, genericNames) {
				rendered, ok := g.renderTypeExpression(ifaceType)
				if ok {
					okName := fmt.Sprintf("%sOk", resultName)
					expected := typeExpressionToString(ifaceType)
					fmt.Fprintf(buf, "\t%s, %s, err := bridge.MatchType(rt, %s, %s)\n", resultName, okName, rendered, resultName)
					fmt.Fprintf(buf, "\tif err != nil {\n")
					fmt.Fprintf(buf, "\t\treturn nil, err\n")
					fmt.Fprintf(buf, "\t}\n")
					fmt.Fprintf(buf, "\tif !%s {\n", okName)
					fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"return type mismatch in %s: expected %s\")\n", funcName, expected)
					fmt.Fprintf(buf, "\t}\n")
					fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
					return
				}
			}
		}
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	case "void":
		fmt.Fprintf(buf, "\t_ = %s\n", resultName)
		fmt.Fprintf(buf, "\treturn runtime.VoidValue{}, nil\n")
	case "bool":
		fmt.Fprintf(buf, "\treturn bridge.ToBool(%s), nil\n", resultName)
	case "string":
		fmt.Fprintf(buf, "\treturn bridge.ToString(%s), nil\n", resultName)
	case "rune":
		fmt.Fprintf(buf, "\treturn bridge.ToRune(%s), nil\n", resultName)
	case "float32":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat32(%s), nil\n", resultName)
	case "float64":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat64(%s), nil\n", resultName)
	case "int":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")), nil\n", resultName)
	case "uint":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")), nil\n", resultName)
	case "int8":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")), nil\n", resultName)
	case "int16":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")), nil\n", resultName)
	case "int32":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")), nil\n", resultName)
	case "int64":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")), nil\n", resultName)
	case "uint8":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")), nil\n", resultName)
	case "uint16":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")), nil\n", resultName)
	case "uint32":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")), nil\n", resultName)
	case "uint64":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")), nil\n", resultName)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\treturn __able_struct_%s_to(rt, %s)\n", baseName, resultName)
	default:
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	}
}
