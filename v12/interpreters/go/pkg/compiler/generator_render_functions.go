package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) renderCompiledFunctions(buf *bytes.Buffer) {
	g.renderCompiledBodies(buf)
}

func (g *generator) renderCompiledFunctionFallback(buf *bytes.Buffer, info *functionInfo) {
	if info == nil {
		return
	}
	g.refreshRepresentableFunctionInfo(info)
	fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
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
	zeroExpr, _ := g.zeroValueExpr(info.ReturnType)
	if info.ReturnType == "struct{}" {
		fmt.Fprintf(buf, "\t_, control := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn struct{}{}, control\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.ReturnType == "runtime.Value" {
		fmt.Fprintf(buf, "\tval, control := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	converted, ok := g.expectRuntimeValueExpr(callExpr, info.ReturnType)
	if !ok {
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"compiler: missing fallback conversion for %s\"))\n", info.Name)
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\tval, control := %s\n", callExpr)
	fmt.Fprintf(buf, "\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn %s, nil\n", strings.ReplaceAll(converted, callExpr, "val"))
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderCompiledMethods(buf *bytes.Buffer) {
	g.renderCompiledBodies(buf)
}

func (g *generator) renderCompiledMethodFallback(buf *bytes.Buffer, method *methodInfo) {
	if method == nil || method.Info == nil {
		return
	}
	info := method.Info
	g.refreshRepresentableFunctionInfo(info)
	fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
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
		callExpr = fmt.Sprintf("__able_method_call(%s, %q, %s)", runtimeArgs[0], method.MethodName, argsExpr)
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
	zeroExpr, _ := g.zeroValueExpr(info.ReturnType)
	if info.ReturnType == "struct{}" {
		fmt.Fprintf(buf, "\t_, control := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn struct{}{}, control\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.ReturnType == "runtime.Value" {
		fmt.Fprintf(buf, "\tval, control := %s\n", callExpr)
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	converted, ok := g.expectRuntimeValueExpr(callExpr, info.ReturnType)
	if !ok {
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"compiler: missing method fallback conversion for %s\"))\n", info.Name)
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\tval, control := %s\n", callExpr)
	fmt.Fprintf(buf, "\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn %s, nil\n", strings.ReplaceAll(converted, callExpr, "val"))
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderWrappers(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil {
			continue
		}
		g.refreshRepresentableFunctionInfo(info)
		if info.InternalOnly {
			continue
		}
		if !info.Compileable && !info.HasOriginal {
			continue
		}
		genericNames := g.callableGenericNames(info)
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "rt", "ctx.Env", "ctx != nil")
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
				g.renderArgConversion(buf, argName, param, info.Name, info.Package, genericNames)
			}
			fmt.Fprintf(buf, "\tcompiledResult, control := __able_compiled_%s(", info.GoName)
			for i, param := range info.Params {
				if i > 0 {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%s", param.GoName)
			}
			fmt.Fprintf(buf, ")\n")
			fmt.Fprintf(buf, "\tif control != nil {\n")
			fmt.Fprintf(buf, "\t\treturn nil, __able_control_to_error(rt, ctx, control)\n")
			fmt.Fprintf(buf, "\t}\n")
			if methodDefinitionExpectsSelf(info.Definition) && len(info.Params) > 0 {
				recv := info.Params[0]
				if g.typeCategory(recv.GoType) == "struct" {
					baseName, ok := g.structHelperName(recv.GoType)
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
		g.refreshRepresentableFunctionInfo(info)
		genericNames := g.methodGenericNames(method)
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "rt", "ctx.Env", "ctx != nil")
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
			g.renderArgConversion(buf, argName, param, info.Name, info.Package, genericNames)
		}
		fmt.Fprintf(buf, "\tcompiledResult, control := __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s", param.GoName)
		}
		fmt.Fprintf(buf, ")\n")
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, __able_control_to_error(rt, ctx, control)\n")
		fmt.Fprintf(buf, "\t}\n")
		if method.ExpectsSelf && len(info.Params) > 0 {
			recv := info.Params[0]
			if g.typeCategory(recv.GoType) == "struct" {
				baseName, ok := g.structHelperName(recv.GoType)
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
	seedStructNames := g.sortedUniqueStructNames()
	seenSeedStruct := make(map[string]struct{}, len(seedStructNames)+1)
	for _, name := range seedStructNames {
		seenSeedStruct[name] = struct{}{}
	}
	if _, ok := seenSeedStruct["Array"]; !ok {
		seedStructNames = append(seedStructNames, "Array")
		sort.Strings(seedStructNames)
	}

	fmt.Fprintf(buf, "func Register(interp *interpreter.Interpreter) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\treturn RegisterIn(interp, nil)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RegisterIn(interp *interpreter.Interpreter, env *runtime.Environment) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\tentryEnv := env\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil && interp != nil {\n")
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
		fmt.Fprintf(buf, "\tif interp != nil && entryEnv == interp.GlobalEnvironment() {\n")
		fmt.Fprintf(buf, "\t\tif pkgEnv := interp.PackageEnvironment(%q); pkgEnv != nil {\n", g.entryPackage)
		fmt.Fprintf(buf, "\t\t\tentryEnv = pkgEnv\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\trt := bridge.New(interp)\n")
	fmt.Fprintf(buf, "\t__able_runtime = rt\n")
	fmt.Fprintf(buf, "\trt.SetEnv(entryEnv)\n")
	fmt.Fprintf(buf, "\t__able_seed_entry_struct_defs(interp, entryEnv)\n")
	if envVar, ok := g.packageEnvVar(g.entryPackage); ok {
		fmt.Fprintf(buf, "\t%s = entryEnv\n", envVar)
	}
	if len(g.diagNodes) > 0 {
		fmt.Fprintf(buf, "\t__able_register_diag_nodes()\n")
	}
	fmt.Fprintf(buf, "\t__able_register_builtin_compiled_calls(entryEnv, interp)\n")
	fmt.Fprintf(buf, "\t__able_register_builtin_compiled_methods()\n")
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_method_impl_packages(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_interface_dispatch(rt); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := __able_register_compiled_packages(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := __able_run_compiled_package_inits(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\trt.SetQualifiedCallableResolver(__able_resolve_qualified_callable)\n")
	fmt.Fprintf(buf, "\tif interp != nil {\n")
	fmt.Fprintf(buf, "\t\tinterp.SetInterfaceMethodResolver(func(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, bool) {\n")
	fmt.Fprintf(buf, "\t\t\treturn __able_resolve_compiled_interface_method(rt, receiver, interfaceName, methodName)\n")
	fmt.Fprintf(buf, "\t\t})\n")
	fmt.Fprintf(buf, "\t\tinterp.SetCompiledImplChecker(func(typeName string, interfaceName string) bool {\n")
	fmt.Fprintf(buf, "\t\t\tentries, ok := __able_interface_dispatch[interfaceName]\n")
	fmt.Fprintf(buf, "\t\t\tif !ok {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tfor _, methodEntries := range entries {\n")
	fmt.Fprintf(buf, "\t\t\t\tfor _, entry := range methodEntries {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tif simple, ok := entry.targetType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == typeName {\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\t\tfor _, v := range entry.unionVariants {\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\tif v == typeName {\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\t\treturn true\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn false\n")
	fmt.Fprintf(buf, "\t\t})\n")
	fmt.Fprintf(buf, "\t\tinterp.SetCompiledInstanceMethodResolver(func(typeName string, methodName string) (runtime.Value, bool) {\n")
	fmt.Fprintf(buf, "\t\t\tentry := __able_lookup_compiled_method(typeName, methodName, true)\n")
	fmt.Fprintf(buf, "\t\t\tif entry == nil || entry.fn == nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn entry.fn, true\n")
	fmt.Fprintf(buf, "\t\t})\n")
	fmt.Fprintf(buf, "\t\tinterp.SetCompiledInterfaceMemberResolver(func(receiver runtime.Value, methodName string) (runtime.Value, bool) {\n")
	fmt.Fprintf(buf, "\t\t\tmethod, ok, err := __able_interface_dispatch_member(receiver, methodName)\n")
	fmt.Fprintf(buf, "\t\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn method, ok\n")
	fmt.Fprintf(buf, "\t\t})\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn rt, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_seed_entry_struct_defs(interp *interpreter.Interpreter, entryEnv *runtime.Environment) {\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif interp != nil {\n")
	fmt.Fprintf(buf, "\t\t_ = interp.SeedStructDefinitions(entryEnv)\n")
	for _, name := range seedStructNames {
		fmt.Fprintf(buf, "\t\tif _, ok := entryEnv.StructDefinition(%q); !ok {\n", name)
		fmt.Fprintf(buf, "\t\t\tif def, found := interp.LookupStructDefinition(%q); found && def != nil {\n", name)
		fmt.Fprintf(buf, "\t\t\t\tentryEnv.DefineStruct(%q, def)\n", name)
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t}\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	// Direct struct definition seeding for no-bootstrap mode
	for _, name := range seedStructNames {
		info, ok := g.structInfoByNameUnique(name)
		if !ok {
			continue
		}
		if info == nil {
			continue
		}
		defExpr, ok := g.renderStructDefinitionExpr(info)
		if !ok {
			continue
		}
		fmt.Fprintf(buf, "\tif _, ok := entryEnv.StructDefinition(%q); !ok {\n", name)
		fmt.Fprintf(buf, "\t\tentryEnv.DefineStruct(%q, %s)\n", name, defExpr)
		fmt.Fprintf(buf, "\t}\n")
	}
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
	fmt.Fprintf(buf, "\tif entryEnv == nil && rt != nil {\n")
	fmt.Fprintf(buf, "\t\tentryEnv = rt.Env()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif entryEnv == nil && interp != nil {\n")
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
	fmt.Fprintf(buf, "\t\terr := func() (err error) {\n")
	fmt.Fprintf(buf, "\t\t\tdefer func() {\n")
	fmt.Fprintf(buf, "\t\t\t\tif r := recover(); r != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\t\terr = bridge.Recover(rt, ctx, r)\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t}()\n")
	fmt.Fprintf(buf, "\t\t\t_, err = entry.fn.Impl(ctx, nil)\n")
	fmt.Fprintf(buf, "\t\t\treturn err\n")
	fmt.Fprintf(buf, "\t\t}()\n")
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmainValue, err := entryEnv.Get(\"main\")\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"entry module does not define a main function\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif interp == nil {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"entry module does not define a main function\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t_, err = interp.CallFunctionIn(mainValue, nil, entryEnv)\n")
	fmt.Fprintf(buf, "\treturn err\n")
	fmt.Fprintf(buf, "}\n")
}

func (g *generator) renderArgConversion(buf *bytes.Buffer, argName string, param paramInfo, funcName string, pkgName string, genericNames map[string]struct{}) {
	goType := param.GoType
	target := param.GoName
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(%sValue)\n", target, spec.FromRuntimeHelper, argName)
		g.renderConvertErr(buf)
		if g.staticArrayArgRequiresValue(pkgName, param.TypeExpr, goType) {
			fmt.Fprintf(buf, "\tif %s == nil {\n", target)
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch calling %s: expected %s\")\n", funcName, typeExpressionToString(param.TypeExpr))
			fmt.Fprintf(buf, "\t}\n")
		}
		return
	}
	if goType == "runtime.ErrorValue" {
		okName := fmt.Sprintf("%sOk", argName)
		nilPtrName := fmt.Sprintf("%sNilPtr", argName)
		fmt.Fprintf(buf, "\t%sCoerced, %s, err := bridge.MatchType(rt, ast.Ty(%q), %sValue)\n", argName, okName, "Error", argName)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif !%s {\n", okName)
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch calling %s: expected Error\")\n", funcName)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s, %s, %s := __able_runtime_error_value(%sCoerced)\n", target, okName, nilPtrName, argName)
		fmt.Fprintf(buf, "\tif !%s || %s {\n", okName, nilPtrName)
		fmt.Fprintf(buf, "\t\t%s = bridge.ErrorValue(rt, %sCoerced)\n", target, argName)
		fmt.Fprintf(buf, "\t}\n")
		return
	}
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
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\t%s, err := %s(%sValue)\n", target, helper, argName)
		g.renderConvertErr(buf)
		return
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(rt, %sValue)\n", target, iface.FromRuntimeHelper, argName)
		g.renderConvertErr(buf)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(rt, %sValue)\n", target, callable.FromRuntimeHelper, argName)
		g.renderConvertErr(buf)
		return
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(rt, %sValue)\n", target, union.FromRuntimeHelper, argName)
		g.renderConvertErr(buf)
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	case "any":
		// runtime.Value satisfies any — direct assignment.
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
		baseName, ok := g.structHelperName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\t%s, err := __able_struct_%s_from(%sValue)\n", target, baseName, argName)
		g.renderConvertErr(buf)
		if g.structArgRequiresValue(pkgName, param.TypeExpr, goType) {
			fmt.Fprintf(buf, "\tif %s == nil {\n", target)
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch calling %s: expected %s\")\n", funcName, typeExpressionToString(param.TypeExpr))
			fmt.Fprintf(buf, "\t}\n")
		}
	default:
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	}
}

func (g *generator) renderReturnConversion(buf *bytes.Buffer, resultName, goType string, returnType ast.TypeExpression, funcName string, genericNames map[string]struct{}) {
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\treturn %s(rt, %s)\n", spec.ToRuntimeHelper, resultName)
		return
	}
	if goType == "runtime.ErrorValue" {
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
		return
	}
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
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\treturn %s(%s), nil\n", helper, resultName)
		return
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "\treturn %s(rt, %s)\n", iface.ToRuntimeHelper, resultName)
		return
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\treturn %s(rt, %s)\n", callable.ToRuntimeHelper, resultName)
		return
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "\treturn %s(rt, %s)\n", union.ToRuntimeHelper, resultName)
		return
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	case "any":
		fmt.Fprintf(buf, "\treturn __able_any_to_value(%s), nil\n", resultName)
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
		lines, ok := g.structReturnConversionLines(resultName, goType, "rt")
		if !ok {
			fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
			return
		}
		for _, line := range lines {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
	default:
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	}
}
