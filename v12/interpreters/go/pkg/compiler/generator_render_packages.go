package compiler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (g *generator) registrationPackageList() []string {
	if g == nil {
		return nil
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
	return packageList
}

func (g *generator) packageRegistrarFuncName(pkgName string, idx int) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return fmt.Sprintf("__able_register_compiled_package_%d", idx)
	}
	return fmt.Sprintf("__able_register_compiled_package_%s_%d", sanitizeIdent(trimmed), idx)
}

func (g *generator) packageCallableRegistrarFuncName(pkgName string, idx int) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return fmt.Sprintf("__able_register_compiled_package_callables_%d", idx)
	}
	return fmt.Sprintf("__able_register_compiled_package_callables_%s_%d", sanitizeIdent(trimmed), idx)
}

func (g *generator) packageMethodImplRegistrarFuncName(pkgName string, idx int) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return fmt.Sprintf("__able_register_compiled_package_methods_impls_%d", idx)
	}
	return fmt.Sprintf("__able_register_compiled_package_methods_impls_%s_%d", sanitizeIdent(trimmed), idx)
}

func (g *generator) packageDefinitionRegistrarFuncName(pkgName string, idx int) string {
	trimmed := strings.TrimSpace(pkgName)
	if trimmed == "" {
		return fmt.Sprintf("__able_register_compiled_package_defs_%d", idx)
	}
	return fmt.Sprintf("__able_register_compiled_package_defs_%s_%d", sanitizeIdent(trimmed), idx)
}

func (g *generator) renderRegisterPackageRegistrars(buf *bytes.Buffer) bool {
	if g == nil || buf == nil {
		return false
	}
	packageList := g.registrationPackageList()
	fmt.Fprintf(buf, "func __able_register_compiled_method_impl_packages(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n")
	for idx, pkgName := range packageList {
		fmt.Fprintf(buf, "\tif err := %s(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n", g.packageMethodImplRegistrarFuncName(pkgName, idx))
		fmt.Fprintf(buf, "\t\treturn err\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_register_compiled_packages(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n")
	for idx, pkgName := range packageList {
		fmt.Fprintf(buf, "\tif err := %s(rt, interp, entryEnv, __able_bootstrapped_metadata); err != nil {\n", g.packageRegistrarFuncName(pkgName, idx))
		fmt.Fprintf(buf, "\t\treturn err\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\t__able_seed_no_bootstrap_imports(__able_bootstrapped_metadata)\n")
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
	return true
}

func (g *generator) renderRegisterPackageRegistrar(buf *bytes.Buffer, pkgName string, idx int) bool {
	if g == nil || buf == nil {
		return false
	}
	fnName := g.packageRegistrarFuncName(pkgName, idx)
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n", fnName)
	fmt.Fprintf(buf, "\t_ = __able_bootstrapped_metadata\n")
	fmt.Fprintf(buf, "\t_ = rt\n")
	fmt.Fprintf(buf, "\t_ = interp\n")
	fmt.Fprintf(buf, "\t_ = entryEnv\n")
	if pkgName == g.entryPackage {
		fmt.Fprintf(buf, "\tpkgEnv := entryEnv\n")
	} else {
		fmt.Fprintf(buf, "\tpkgEnv := interp.PackageEnvironment(%q)\n", pkgName)
		fmt.Fprintf(buf, "\tif pkgEnv == nil {\n")
		fmt.Fprintf(buf, "\t\tpkgEnv = runtime.NewEnvironment(entryEnv)\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	if envVar, ok := g.packageEnvVar(pkgName); ok {
		fmt.Fprintf(buf, "\t%s = pkgEnv\n", envVar)
	}
	fmt.Fprintf(buf, "\tif err := %s(rt, interp, pkgEnv, __able_bootstrapped_metadata); err != nil {\n", g.packageDefinitionRegistrarFuncName(pkgName, idx))
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := %s(rt, interp, pkgEnv, __able_bootstrapped_metadata); err != nil {\n", g.packageCallableRegistrarFuncName(pkgName, idx))
	fmt.Fprintf(buf, "\t\treturn err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
	return true
}

func (g *generator) renderCompiledPackageCallableFiles() (map[string][]byte, error) {
	if g == nil || !g.hasFunctions() {
		return nil, nil
	}
	out := make(map[string][]byte)
	packageList := g.registrationPackageList()
	for idx, pkgName := range packageList {
		fileName := fmt.Sprintf("compiled_pkg_callables_%s_%d.go", sanitizeIdent(strings.TrimSpace(pkgName)), idx)
		if strings.TrimSpace(pkgName) == "" {
			fileName = fmt.Sprintf("compiled_pkg_callables_entry_%d.go", idx)
		}
		src, err := g.renderCompiledPackageCallableFile(pkgName, idx)
		if err != nil {
			return nil, err
		}
		out[fileName] = src
	}
	return out, nil
}

func (g *generator) renderCompiledPackageCallableFile(pkgName string, idx int) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := []string{
		"able/interpreter-go/pkg/ast",
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var _ = ast.NewIdentifier\n\n")
	fnName := g.packageCallableRegistrarFuncName(pkgName, idx)
	fmt.Fprintf(&buf, "func %s(rt *bridge.Runtime, interp *interpreter.Interpreter, pkgEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n", fnName)
	fmt.Fprintf(&buf, "\t_ = rt\n")
	fmt.Fprintf(&buf, "\t_ = interp\n")
	fmt.Fprintf(&buf, "\t_ = pkgEnv\n")
	fmt.Fprintf(&buf, "\t_ = __able_bootstrapped_metadata\n")
	for _, name := range g.sortedCallableNames(pkgName) {
		if overload, ok := g.overloads[pkgName][name]; ok && overload != nil {
			qualified := overload.QualifiedName
			if qualified == "" {
				qualified = qualifiedName(pkgName, name)
			}
			fmt.Fprintf(&buf, "\tif original, err := pkgEnv.Get(%q); err == nil {\n", name)
			fmt.Fprintf(&buf, "\t\trt.RegisterOriginal(%q, original)\n", qualified)
			fmt.Fprintf(&buf, "\t}\n")
			fmt.Fprintf(&buf, "\t{\n")
			fmt.Fprintf(&buf, "\t\toverloadFn := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", name)
			fmt.Fprintf(&buf, "\t\toverloadFn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			fmt.Fprintf(&buf, "\t\t\treturn %s(overloadFn, ctx, args, nil)\n", g.overloadWrapperName(pkgName, name))
			fmt.Fprintf(&buf, "\t\t}\n")
			fmt.Fprintf(&buf, "\t\t%s = overloadFn\n", g.overloadValueName(pkgName, name))
			fmt.Fprintf(&buf, "\t}\n")
			for _, entry := range overload.Entries {
				if entry == nil || !entry.Compileable {
					continue
				}
				paramExprs, ok := g.renderFunctionParamTypes(entry)
				if !ok {
					return nil, fmt.Errorf("compiler: render param types for overload %s in package %s", name, pkgName)
				}
				fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
				fmt.Fprintf(&buf, "\t\tif _, err := pkgEnv.Get(%q); err == nil {\n", name)
				fmt.Fprintf(&buf, "\t\t\tif err := interp.RegisterCompiledFunctionOverload(pkgEnv, %q, %s, __able_function_thunk_%s); err != nil {\n", name, paramExprs, entry.GoName)
				fmt.Fprintf(&buf, "\t\t\t\treturn err\n")
				fmt.Fprintf(&buf, "\t\t\t}\n")
				fmt.Fprintf(&buf, "\t\t}\n")
				fmt.Fprintf(&buf, "\t}\n")
			}
			fmt.Fprintf(&buf, "\t__able_register_compiled_call(pkgEnv, %q, -1, %d, %q, func(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", name, overload.MinArity, "")
			fmt.Fprintf(&buf, "\t\treturn %s(%s, ctx, args, nil)\n", g.overloadWrapperName(pkgName, name), g.overloadValueName(pkgName, name))
			fmt.Fprintf(&buf, "\t})\n")
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
		fmt.Fprintf(&buf, "\tif original, err := pkgEnv.Get(%q); err == nil {\n", info.Name)
		fmt.Fprintf(&buf, "\t\trt.RegisterOriginal(%q, original)\n", qualified)
		fmt.Fprintf(&buf, "\t}\n")
		if info.Compileable {
			paramExprs, ok := g.renderFunctionParamTypes(info)
			if !ok {
				return nil, fmt.Errorf("compiler: render param types for function %s in package %s", info.Name, pkgName)
			}
			fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
			fmt.Fprintf(&buf, "\t\tif _, err := pkgEnv.Get(%q); err == nil {\n", info.Name)
			fmt.Fprintf(&buf, "\t\t\tif err := interp.RegisterCompiledFunctionOverload(pkgEnv, %q, %s, __able_function_thunk_%s); err != nil {\n", info.Name, paramExprs, info.GoName)
			fmt.Fprintf(&buf, "\t\t\t\treturn err\n")
			fmt.Fprintf(&buf, "\t\t\t}\n")
			fmt.Fprintf(&buf, "\t\t}\n")
			fmt.Fprintf(&buf, "\t\t}\n")
		}
		if info.Arity >= 0 {
			minArgs := info.Arity
			if g.hasOptionalLastParam(info) && info.Arity > 0 {
				minArgs = info.Arity - 1
			}
			ufcsTarget := g.ufcsTargetName(info)
			fmt.Fprintf(&buf, "\t__able_register_compiled_call(pkgEnv, %q, %d, %d, %q, __able_wrap_%s)\n", info.Name, info.Arity, minArgs, ufcsTarget, info.GoName)
		}
	}
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}

func (g *generator) methodRegistrationKeyCounts() map[string]int {
	counts := make(map[string]int)
	if g == nil {
		return counts
	}
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		key := fmt.Sprintf("%s|%s|%t", method.TargetName, method.MethodName, method.ExpectsSelf)
		counts[key]++
	}
	return counts
}

func (g *generator) ufcsTargetName(info *functionInfo) string {
	if g == nil || info == nil || info.Definition == nil || len(info.Params) == 0 {
		return ""
	}
	if info.Definition.IsMethodShorthand {
		return ""
	}
	targetExpr := g.expandTypeAliasForPackage(info.Package, info.Params[0].TypeExpr)
	targetName, ok := g.methodTargetName(targetExpr)
	if !ok {
		return ""
	}
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		return ""
	}
	if _, ok := g.structs[targetName]; ok {
		return targetName
	}
	switch targetName {
	case "String", "Array", "HashMap", "Error", "Iterator",
		"bool", "char", "nil", "void",
		"i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128",
		"f32", "f64":
		return targetName
	default:
		return ""
	}
}

func (g *generator) renderCompiledPackageMethodImplFiles() (map[string][]byte, error) {
	if g == nil || !g.hasFunctions() {
		return nil, nil
	}
	out := make(map[string][]byte)
	packageList := g.registrationPackageList()
	methodCounts := g.methodRegistrationKeyCounts()
	for idx, pkgName := range packageList {
		fileName := fmt.Sprintf("compiled_pkg_methods_impls_%s_%d.go", sanitizeIdent(strings.TrimSpace(pkgName)), idx)
		if strings.TrimSpace(pkgName) == "" {
			fileName = fmt.Sprintf("compiled_pkg_methods_impls_entry_%d.go", idx)
		}
		src, err := g.renderCompiledPackageMethodImplFile(pkgName, idx, methodCounts)
		if err != nil {
			return nil, err
		}
		out[fileName] = src
	}
	return out, nil
}

func (g *generator) renderCompiledPackageMethodImplFile(pkgName string, idx int, methodCounts map[string]int) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)
	imports := []string{
		"able/interpreter-go/pkg/ast",
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/interpreter",
		"able/interpreter-go/pkg/runtime",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var _ = ast.NewIdentifier\n\n")
	fnName := g.packageMethodImplRegistrarFuncName(pkgName, idx)
	fmt.Fprintf(&buf, "func %s(rt *bridge.Runtime, interp *interpreter.Interpreter, entryEnv *runtime.Environment, __able_bootstrapped_metadata bool) error {\n", fnName)
	fmt.Fprintf(&buf, "\t_ = rt\n")
	fmt.Fprintf(&buf, "\t_ = interp\n")
	fmt.Fprintf(&buf, "\t_ = entryEnv\n")
	fmt.Fprintf(&buf, "\t_ = __able_bootstrapped_metadata\n")
	fmt.Fprintf(&buf, "\tpkgEnv := entryEnv\n")
	if pkgName != g.entryPackage {
		fmt.Fprintf(&buf, "\tpkgEnv = interp.PackageEnvironment(%q)\n", pkgName)
		fmt.Fprintf(&buf, "\tif pkgEnv == nil {\n")
		fmt.Fprintf(&buf, "\t\tpkgEnv = runtime.NewEnvironment(entryEnv)\n")
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\t_ = pkgEnv\n")
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil || method.Info.Package != pkgName {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(method.TargetType)
		if !ok {
			return nil, fmt.Errorf("compiler: render method target type %s.%s", method.TargetName, method.MethodName)
		}
		paramExprs, ok := g.renderMethodParamTypes(method)
		if !ok {
			return nil, fmt.Errorf("compiler: render method params %s.%s", method.TargetName, method.MethodName)
		}
		fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
		fmt.Fprintf(&buf, "\t\tif err := interp.RegisterCompiledMethodOverload(%q, %q, %t, %s, %s, __able_method_thunk_%s); err != nil {\n", method.TargetName, method.MethodName, method.ExpectsSelf, targetExpr, paramExprs, method.Info.GoName)
		fmt.Fprintf(&buf, "\t\t\treturn err\n")
		fmt.Fprintf(&buf, "\t\t}\n")
		fmt.Fprintf(&buf, "\t}\n")
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
			fmt.Fprintf(&buf, "\t__able_register_compiled_method(%q, %q, %t, %d, %d, __able_wrap_%s)\n", method.TargetName, method.MethodName, method.ExpectsSelf, arity, minArgs, method.Info.GoName)
		}
	}
	for _, implMethod := range g.sortedImplMethodInfos() {
		if implMethod == nil || implMethod.Info == nil || implMethod.Info.Package != pkgName || !implMethod.Info.Compileable {
			continue
		}
		if implMethod.IsDefault {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(implMethod.TargetType)
		if !ok {
			return nil, fmt.Errorf("compiler: render impl target type %s.%s", implMethod.InterfaceName, implMethod.MethodName)
		}
		ifaceArgsExpr, ok := g.renderTypeExpressionList(implMethod.InterfaceArgs)
		if !ok {
			return nil, fmt.Errorf("compiler: render impl interface args %s.%s", implMethod.InterfaceName, implMethod.MethodName)
		}
		paramExprs, ok := g.renderImplMethodParamTypes(implMethod)
		if !ok {
			return nil, fmt.Errorf("compiler: render impl params %s.%s", implMethod.InterfaceName, implMethod.MethodName)
		}
		constraintKey := constraintSignature(collectConstraintSpecs(implMethod.ImplGenerics, implMethod.WhereClause))
		if implMethod.ImplName != "" {
			fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
			fmt.Fprintf(&buf, "\t\tif err := interp.RegisterCompiledImplNamespaceMethod(pkgEnv, %q, %q, %s, __able_function_thunk_%s); err != nil {\n", implMethod.ImplName, implMethod.MethodName, paramExprs, implMethod.Info.GoName)
			fmt.Fprintf(&buf, "\t\t\treturn err\n")
			fmt.Fprintf(&buf, "\t\t}\n")
			fmt.Fprintf(&buf, "\t}\n")
			continue
		}
		fmt.Fprintf(&buf, "\tif __able_bootstrapped_metadata {\n")
		fmt.Fprintf(&buf, "\t\tif err := interp.RegisterCompiledImplMethodOverload(%q, %s, %s, %q, %q, %q, %s, __able_function_thunk_%s); err != nil {\n",
			implMethod.InterfaceName, targetExpr, ifaceArgsExpr, constraintKey, implMethod.ImplName, implMethod.MethodName, paramExprs, implMethod.Info.GoName)
		fmt.Fprintf(&buf, "\t\t\treturn err\n")
		fmt.Fprintf(&buf, "\t\t}\n")
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}
