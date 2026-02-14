package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderCompiledFunctions(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable {
			continue
		}
		ctx := newCompileContext(info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
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
			g.renderCompiledFunctionFallback(buf, info)
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

func (g *generator) renderCompiledFunctionFallback(buf *bytes.Buffer, info *functionInfo) {
	if info == nil {
		return
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
	qualified := info.Name
	if info.QualifiedName != "" {
		qualified = info.QualifiedName
	}
	runtimeArgs := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		argExpr, ok := g.runtimeValueExpr(param.GoName, param.GoType)
		if !ok {
			argExpr = "runtime.NilValue{}"
		}
		runtimeArgs = append(runtimeArgs, argExpr)
	}
	argList := "nil"
	if len(runtimeArgs) > 0 {
		argList = "[]runtime.Value{" + strings.Join(runtimeArgs, ", ") + "}"
	}
	callExpr := fmt.Sprintf("__able_call_named(%q, %s, nil)", qualified, argList)
	if info.ReturnType == "struct{}" {
		fmt.Fprintf(buf, "\t_ = %s\n", callExpr)
		fmt.Fprintf(buf, "\treturn struct{}{}\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.ReturnType == "runtime.Value" {
		fmt.Fprintf(buf, "\tval := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	converted, ok := g.expectRuntimeValueExpr(callExpr, info.ReturnType)
	if !ok {
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"compiler: missing fallback conversion for %s\"))\n", info.Name)
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\treturn %s\n", converted)
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderCompiledMethods(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil || !method.Info.Compileable {
			continue
		}
		info := method.Info
		ctx := newCompileContext(info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
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
			g.renderCompiledMethodFallback(buf, method)
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

func (g *generator) renderCompiledMethodFallback(buf *bytes.Buffer, method *methodInfo) {
	if method == nil || method.Info == nil {
		return
	}
	info := method.Info
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
	runtimeArgs := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		argExpr, ok := g.runtimeValueExpr(param.GoName, param.GoType)
		if !ok {
			argExpr = "runtime.NilValue{}"
		}
		runtimeArgs = append(runtimeArgs, argExpr)
	}
	callExpr := ""
	if method.ExpectsSelf && len(runtimeArgs) > 0 {
		argsExpr := "nil"
		if len(runtimeArgs) > 1 {
			argsExpr = "[]runtime.Value{" + strings.Join(runtimeArgs[1:], ", ") + "}"
		}
		callExpr = fmt.Sprintf("__able_call_value(__able_member_get_method(%s, bridge.ToString(%q)), %s, nil)", runtimeArgs[0], method.MethodName, argsExpr)
	} else {
		target := method.MethodName
		if method.TargetName != "" {
			target = method.TargetName + "." + method.MethodName
		}
		argsExpr := "nil"
		if len(runtimeArgs) > 0 {
			argsExpr = "[]runtime.Value{" + strings.Join(runtimeArgs, ", ") + "}"
		}
		callExpr = fmt.Sprintf("__able_call_named(%q, %s, nil)", target, argsExpr)
	}
	if info.ReturnType == "struct{}" {
		fmt.Fprintf(buf, "\t_ = %s\n", callExpr)
		fmt.Fprintf(buf, "\treturn struct{}{}\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.ReturnType == "runtime.Value" {
		fmt.Fprintf(buf, "\tval := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	converted, ok := g.expectRuntimeValueExpr(callExpr, info.ReturnType)
	if !ok {
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"compiler: missing method fallback conversion for %s\"))\n", info.Name)
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\treturn %s\n", converted)
	fmt.Fprintf(buf, "}\n\n")
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
			if methodDefinitionExpectsSelf(info.Definition) && len(info.Params) > 0 {
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
			continue
		}
		if info.HasOriginal {
			qualified := info.Name
			if info.QualifiedName != "" {
				qualified = info.QualifiedName
			}
			fmt.Fprintf(buf, "\t__able_mark_boundary_explicit(\"call_original\", %q)\n", qualified)
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

func (g *generator) renderRegister(buf *bytes.Buffer) {
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
	fmt.Fprintf(buf, "\t__able_bootstrapped_metadata := false\n")
	fmt.Fprintf(buf, "\tif existingMain, err := entryEnv.Get(\"main\"); err == nil {\n")
	fmt.Fprintf(buf, "\t\tswitch existingMain.(type) {\n")
	fmt.Fprintf(buf, "\t\tcase *runtime.FunctionValue, *runtime.FunctionOverloadValue:\n")
	fmt.Fprintf(buf, "\t\t\t__able_bootstrapped_metadata = true\n")
	fmt.Fprintf(buf, "\t\t}\n")
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
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_method_impl_packages(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_interface_dispatch(rt); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_packages(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn rt, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RunMain(interp *interpreter.Interpreter) error {\n")
	fmt.Fprintf(buf, "\treturn RunMainIn(interp, nil)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RunMainIn(interp *interpreter.Interpreter, env *runtime.Environment) error {\n")
	fmt.Fprintf(buf, "\trt, err := RegisterIn(interp, env)\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tentryEnv := env\n")
	fmt.Fprintf(buf, "\tif rt != nil && rt.Env() != nil {\n")
	fmt.Fprintf(buf, "\t\tentryEnv = rt.Env()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn RunRegisteredMain(rt, interp, entryEnv)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RunRegisteredMain(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment) error {\n")
	fmt.Fprintf(buf, "\tif interp == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"missing interpreter\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil && rt != nil {\n")
	fmt.Fprintf(buf, "\t\tentryEnv = rt.Env()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(buf, "\t\tentryEnv = interp.GlobalEnvironment()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"missing environment\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif rt != nil {\n")
	fmt.Fprintf(buf, "\t\trt.SetEnv(entryEnv)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entry := __able_lookup_compiled_call(entryEnv, \"main\"); entry != nil && entry.fn != nil {\n")
	fmt.Fprintf(buf, "\t\tvar state any\n")
	fmt.Fprintf(buf, "\t\tstate = entryEnv.RuntimeData()\n")
	fmt.Fprintf(buf, "\t\tctx := &runtime.NativeCallContext{Env: entryEnv, State: state}\n")
	fmt.Fprintf(buf, "\t\t_, err := entry.fn.Impl(ctx, nil)\n")
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmainValue, err := entryEnv.Get(\"main\")\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t_, err = interp.CallFunctionIn(mainValue, nil, entryEnv)\n")
	fmt.Fprintf(buf, "\treturn err\n")
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
