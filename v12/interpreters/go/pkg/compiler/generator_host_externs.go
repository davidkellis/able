package compiler

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) collectGoPrelude(pkgName string, stmt *ast.PreludeStatement) {
	if g == nil || stmt == nil || stmt.Target != ast.HostTargetGo {
		return
	}
	pkgName = strings.TrimSpace(pkgName)
	code := strings.TrimSpace(stmt.Code)
	if pkgName == "" || code == "" {
		return
	}
	if g.goPreludes == nil {
		g.goPreludes = make(map[string][]string)
	}
	g.goPreludes[pkgName] = append(g.goPreludes[pkgName], code)
}

func (g *generator) collectGoExternBody(pkgName string, body *ast.ExternFunctionBody) {
	if g == nil || body == nil || body.Target != ast.HostTargetGo || body.Signature == nil || body.Signature.ID == nil {
		return
	}
	name := strings.TrimSpace(body.Signature.ID.Name)
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" || name == "" {
		return
	}
	if strings.HasPrefix(name, "__able_") {
		return
	}
	if g.externBodies == nil {
		g.externBodies = make(map[string]map[string][]*ast.ExternFunctionBody)
	}
	if g.externBodies[pkgName] == nil {
		g.externBodies[pkgName] = make(map[string][]*ast.ExternFunctionBody)
	}
	g.externBodies[pkgName][name] = append(g.externBodies[pkgName][name], body)
}

func (g *generator) externBodiesForPackage(pkgName string) map[string][]*ast.ExternFunctionBody {
	if g == nil || g.externBodies == nil {
		return nil
	}
	return g.externBodies[strings.TrimSpace(pkgName)]
}

func (g *generator) hasCompiledGoExterns() bool {
	if g == nil {
		return false
	}
	for _, pkgFuncs := range g.functions {
		for _, info := range pkgFuncs {
			if info != nil && info.ExternBody != nil {
				return true
			}
		}
	}
	for _, pkgOverloads := range g.overloads {
		for _, overload := range pkgOverloads {
			if overload == nil {
				continue
			}
			for _, entry := range overload.Entries {
				if entry != nil && entry.ExternBody != nil {
					return true
				}
			}
		}
	}
	return false
}

func (g *generator) prepareGoHostSupport() error {
	if g == nil {
		return nil
	}
	g.goPreludeImports = nil
	g.goPreludeDecls = nil
	if !g.hasCompiledGoExterns() {
		return nil
	}
	importSet := make(map[string]struct{})
	declSet := make(map[string]struct{})
	pkgNames := make([]string, 0, len(g.goPreludes))
	for pkgName := range g.goPreludes {
		if len(g.externBodiesForPackage(pkgName)) == 0 {
			continue
		}
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)
	for _, pkgName := range pkgNames {
		for _, code := range g.goPreludes[pkgName] {
			imports, decls, err := parseGoPrelude(code)
			if err != nil {
				return fmt.Errorf("compiler: parse go prelude for package %s: %w", pkgName, err)
			}
			for _, imp := range imports {
				importSet[imp] = struct{}{}
			}
			for _, decl := range decls {
				if _, exists := declSet[decl]; exists {
					continue
				}
				declSet[decl] = struct{}{}
				g.goPreludeDecls = append(g.goPreludeDecls, decl)
			}
		}
	}
	for imp := range importSet {
		g.goPreludeImports = append(g.goPreludeImports, imp)
	}
	sort.Strings(g.goPreludeImports)
	return nil
}

func parseGoPrelude(code string) ([]string, []string, error) {
	src := "package __able_prelude\n\n" + code + "\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "prelude.go", src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	imports := make([]string, 0, len(file.Imports))
	decls := make([]string, 0, len(file.Decls))
	renames := make(map[string]string)
	for _, imp := range file.Imports {
		if imp == nil || imp.Path == nil {
			continue
		}
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return nil, nil, err
		}
		name := defaultGoImportName(path)
		if imp.Name != nil && imp.Name.Name != "" {
			name = imp.Name.Name
		}
		spec := strconv.Quote(path)
		if alias, ok := preludeImportAlias(name, path); ok {
			if name != "_" && name != "." {
				renames[name] = alias
			}
			spec = alias + " " + spec
		} else if imp.Name != nil && imp.Name.Name != "" {
			spec = imp.Name.Name + " " + spec
		}
		imports = append(imports, spec)
	}
	if len(renames) > 0 {
		for _, decl := range file.Decls {
			renamePreludeImportSelectors(decl, renames)
		}
	}
	for _, decl := range file.Decls {
		if gen, ok := decl.(*goast.GenDecl); ok && gen != nil && gen.Tok == token.IMPORT {
			continue
		}
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, decl); err != nil {
			return nil, nil, err
		}
		decls = append(decls, strings.TrimSpace(buf.String()))
	}
	return imports, decls, nil
}

func defaultGoImportName(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 && idx < len(path)-1 {
		return path[idx+1:]
	}
	return path
}

func preludeImportAlias(name string, path string) (string, bool) {
	switch name {
	case "runtime":
		if path != "able/interpreter-go/pkg/runtime" {
			return "goruntime", true
		}
	case "ast":
		if path != "able/interpreter-go/pkg/ast" {
			return "goastlib", true
		}
	case "bridge":
		if path != "able/interpreter-go/pkg/compiler/bridge" {
			return "gobridge", true
		}
	case "interpreter":
		if path != "able/interpreter-go/pkg/interpreter" {
			return "gointerpreter", true
		}
	}
	return "", false
}

func renamePreludeImportSelectors(node goast.Node, renames map[string]string) {
	if node == nil || len(renames) == 0 {
		return
	}
	goast.Inspect(node, func(curr goast.Node) bool {
		sel, ok := curr.(*goast.SelectorExpr)
		if !ok || sel == nil {
			return true
		}
		ident, ok := sel.X.(*goast.Ident)
		if !ok || ident == nil {
			return true
		}
		if replacement, ok := renames[ident.Name]; ok {
			ident.Name = replacement
		}
		return true
	})
}

func (g *generator) renderGoPreludeDecls(buf *bytes.Buffer) {
	if g == nil || buf == nil || !g.hasCompiledGoExterns() {
		return
	}
	fmt.Fprintf(buf, "type IoHandle = any\n")
	fmt.Fprintf(buf, "type ProcHandle = any\n\n")
	fmt.Fprintf(buf, "type __able_host_error struct{ message string }\n")
	fmt.Fprintf(buf, "func (e __able_host_error) Error() string { return e.message }\n")
	fmt.Fprintf(buf, "func host_error[T any](message string) (result T, err error) {\n")
	fmt.Fprintf(buf, "\terr = __able_host_error{message: message}\n")
	fmt.Fprintf(buf, "\treturn result, err\n")
	fmt.Fprintf(buf, "}\n\n")
	for _, decl := range g.goPreludeDecls {
		fmt.Fprintf(buf, "%s\n\n", decl)
	}
}

func (g *generator) externBodyCompileable(info *functionInfo) bool {
	if info == nil || info.ExternBody == nil || info.Definition == nil {
		if info != nil {
			info.Reason = "missing extern body"
		}
		return false
	}
	for _, param := range info.Params {
		expr := normalizeTypeExprForPackage(g, info.Package, param.TypeExpr)
		if _, ok := g.renderTypeExpression(expr); !ok {
			info.Reason = "unsupported extern parameter type"
			return false
		}
		if _, err := goExternHostType(expr, false); err != nil {
			info.Reason = err.Error()
			return false
		}
	}
	retExpr := normalizeTypeExprForPackage(g, info.Package, info.Definition.ReturnType)
	if _, ok := g.renderTypeExpression(retExpr); !ok {
		info.Reason = "unsupported extern return type"
		return false
	}
	if resultExpr, ok := retExpr.(*ast.ResultTypeExpression); ok && resultExpr != nil {
		if _, err := goExternHostType(normalizeTypeExprForPackage(g, info.Package, resultExpr.InnerType), false); err != nil {
			info.Reason = err.Error()
			return false
		}
		return true
	}
	if _, err := goExternHostType(retExpr, true); err != nil {
		info.Reason = err.Error()
		return false
	}
	return true
}

func goExternHostType(expr ast.TypeExpression, allowVoidAny bool) (string, error) {
	if expr == nil {
		return "any", nil
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "", fmt.Errorf("compiler: unsupported extern host type")
		}
		switch normalizeExternKernelTypeName(t.Name.Name) {
		case "String":
			return "string", nil
		case "bool":
			return "bool", nil
		case "char":
			return "rune", nil
		case "void":
			if allowVoidAny {
				return "any", nil
			}
			return "", nil
		case "IoHandle", "ProcHandle":
			return "any", nil
		case "i8":
			return "int8", nil
		case "i16":
			return "int16", nil
		case "i32":
			return "int32", nil
		case "i64":
			return "int64", nil
		case "u8":
			return "uint8", nil
		case "u16":
			return "uint16", nil
		case "u32":
			return "uint32", nil
		case "u64":
			return "uint64", nil
		case "i128", "u128":
			return "*big.Int", nil
		case "f32":
			return "float32", nil
		case "f64":
			return "float64", nil
		default:
			return "any", nil
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && normalizeExternKernelTypeName(base.Name.Name) == "Array" {
			elemType := "any"
			if len(t.Arguments) > 0 {
				mapped, err := goExternHostType(t.Arguments[0], false)
				if err != nil {
					return "", err
				}
				if strings.TrimSpace(mapped) != "" {
					elemType = mapped
				}
			}
			return "[]" + elemType, nil
		}
		return "any", nil
	case *ast.NullableTypeExpression:
		inner, err := goExternHostType(t.InnerType, false)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(inner) == "" {
			inner = "any"
		}
		return "*" + inner, nil
	case *ast.ResultTypeExpression:
		return "any", nil
	case *ast.UnionTypeExpression, *ast.FunctionTypeExpression, *ast.WildcardTypeExpression:
		return "any", nil
	default:
		return "any", nil
	}
}

func normalizeExternKernelTypeName(name string) string {
	switch strings.TrimSpace(name) {
	case "string":
		return "String"
	default:
		return strings.TrimSpace(name)
	}
}

func (g *generator) externHostFunctionName(info *functionInfo) string {
	if info == nil {
		return "__able_host_extern"
	}
	return "__able_host_" + info.GoName
}

func (g *generator) renderGoExternHostFunction(buf *bytes.Buffer, info *functionInfo) error {
	if g == nil || buf == nil || info == nil || info.ExternBody == nil || info.Definition == nil {
		return nil
	}
	extern := info.ExternBody
	fnName := g.externHostFunctionName(info)
	params := make([]string, 0, len(info.Definition.Params))
	argNames := make([]string, 0, len(info.Definition.Params))
	for idx, param := range info.Definition.Params {
		paramName := fmt.Sprintf("arg%d", idx)
		if param != nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && strings.TrimSpace(ident.Name) != "" {
				paramName = ident.Name
			}
		}
		hostType, err := goExternHostType(normalizeTypeExprForPackage(g, info.Package, param.ParamType), false)
		if err != nil {
			return err
		}
		if strings.TrimSpace(hostType) == "" {
			hostType = "any"
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, hostType))
		argNames = append(argNames, paramName)
	}
	retExpr := normalizeTypeExprForPackage(g, info.Package, info.Definition.ReturnType)
	switch ret := retExpr.(type) {
	case *ast.ResultTypeExpression:
		innerType, err := goExternHostType(normalizeTypeExprForPackage(g, info.Package, ret.InnerType), false)
		if err != nil {
			return err
		}
		if strings.TrimSpace(innerType) == "" {
			innerType = "any"
		}
		fmt.Fprintf(buf, "func %s(%s) (__able_result %s, __able_err error) {\n", fnName, strings.Join(params, ", "), innerType)
	default:
		hostType, err := goExternHostType(retExpr, true)
		if err != nil {
			return err
		}
		if strings.TrimSpace(hostType) == "" {
			hostType = "any"
		}
		fmt.Fprintf(buf, "func %s(%s) (__able_result %s) {\n", fnName, strings.Join(params, ", "), hostType)
	}
	body := strings.TrimSpace(extern.Body)
	if body != "" {
		for _, line := range strings.Split(body, "\n") {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
	}
	fmt.Fprintf(buf, "\treturn\n")
	fmt.Fprintf(buf, "}\n\n")
	return nil
}

func (g *generator) renderCompiledExternFunctionBody(buf *bytes.Buffer, info *functionInfo) {
	if g == nil || buf == nil || info == nil || info.ExternBody == nil || info.Definition == nil {
		return
	}
	if err := g.renderGoExternHostFunction(buf, info); err != nil {
		panic(err)
	}
	bodyName := g.compiledBodyName(info)
	fmt.Fprintf(buf, "func %s(", bodyName)
	for i, param := range info.Params {
		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnType)
	zeroExpr, ok := g.zeroValueExpr(info.ReturnType)
	if !ok {
		zeroExpr = "runtime.NilValue{}"
	}
	for idx, param := range info.Params {
		runtimeExpr, ok := g.runtimeValueExpr(param.GoName, param.GoType)
		if !ok {
			panic(fmt.Errorf("compiler: unsupported extern argument carrier %s", param.GoType))
		}
		typeExpr := normalizeTypeExprForPackage(g, info.Package, param.TypeExpr)
		typeExprCode, ok := g.renderTypeExpression(typeExpr)
		if !ok {
			panic(fmt.Errorf("compiler: render extern arg type %s", param.Name))
		}
		hostType, err := goExternHostType(typeExpr, false)
		if err != nil {
			panic(err)
		}
		if strings.TrimSpace(hostType) == "" {
			hostType = "any"
		}
		if directExpr, ok := directHostArgExpr(param.GoType, hostType, param.GoName); ok {
			fmt.Fprintf(buf, "\t__able_host_arg_%d := %s\n", idx, directExpr)
		} else {
			fmt.Fprintf(buf, "\t__able_host_arg_%d, err := bridge.RuntimeValueToHost[%s](%s, %s)\n", idx, hostType, typeExprCode, runtimeExpr)
			fmt.Fprintf(buf, "\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
		}
	}
	hostArgs := make([]string, 0, len(info.Params))
	for idx := range info.Params {
		hostArgs = append(hostArgs, fmt.Sprintf("__able_host_arg_%d", idx))
	}
	retExpr := normalizeTypeExprForPackage(g, info.Package, info.Definition.ReturnType)
	switch ret := retExpr.(type) {
	case *ast.ResultTypeExpression:
		innerExpr := normalizeTypeExprForPackage(g, info.Package, ret.InnerType)
		innerExprCode, ok := g.renderTypeExpression(innerExpr)
		if !ok {
			panic(fmt.Errorf("compiler: render extern result inner type %s", info.Name))
		}
		fmt.Fprintf(buf, "\t__able_host_result, __able_host_err := %s(%s)\n", g.externHostFunctionName(info), strings.Join(hostArgs, ", "))
		fmt.Fprintf(buf, "\t__able_runtime_result, err := bridge.HostResultToRuntime(__able_runtime, %s, __able_host_result, __able_host_err)\n", innerExprCode)
	default:
		retExprCode, ok := g.renderTypeExpression(retExpr)
		if !ok {
			panic(fmt.Errorf("compiler: render extern return type %s", info.Name))
		}
		hostType, err := goExternHostType(retExpr, true)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(buf, "\t__able_host_result := %s(%s)\n", g.externHostFunctionName(info), strings.Join(hostArgs, ", "))
		if directExpr, ok := directHostReturnExpr(info.ReturnType, hostType, "__able_host_result"); ok {
			fmt.Fprintf(buf, "\treturn %s, nil\n", directExpr)
			fmt.Fprintf(buf, "}\n\n")
			g.renderCompiledExternEntryWrapper(buf, info, bodyName)
			return
		}
		if emitted, terminal := g.renderDirectHostReturnIfPossible(buf, info, retExpr, "__able_host_result"); terminal {
			fmt.Fprintf(buf, "}\n\n")
			g.renderCompiledExternEntryWrapper(buf, info, bodyName)
			return
		} else if emitted {
			fmt.Fprintf(buf, "\t// Host result did not match a native fast-path shape; use the semantic bridge.\n")
		}
		fmt.Fprintf(buf, "\t__able_runtime_result, err := bridge.HostValueToRuntime(__able_runtime, %s, __able_host_result)\n", retExprCode)
	}
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	if info.ReturnType == "struct{}" {
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.ReturnType == "runtime.Value" {
		fmt.Fprintf(buf, "\tif __able_runtime_result == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn __able_runtime_result, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	converted, ok := g.expectRuntimeValueExpr("__able_runtime_result", info.ReturnType)
	if !ok {
		panic(fmt.Errorf("compiler: missing extern conversion for %s", info.Name))
	}
	fmt.Fprintf(buf, "\treturn %s, nil\n", converted)
	fmt.Fprintf(buf, "}\n\n")
	g.renderCompiledExternEntryWrapper(buf, info, bodyName)
}

func (g *generator) renderCompiledExternEntryWrapper(buf *bytes.Buffer, info *functionInfo, bodyName string) {
	if g == nil || buf == nil || info == nil {
		return
	}
	entryName := g.compiledEntryName(info)
	fmt.Fprintf(buf, "func %s(", entryName)
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
	args := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	fmt.Fprintf(buf, "\treturn %s(%s)\n", bodyName, strings.Join(args, ", "))
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderDirectHostReturnIfPossible(buf *bytes.Buffer, info *functionInfo, retExpr ast.TypeExpression, hostResult string) (bool, bool) {
	if g == nil || buf == nil || info == nil || retExpr == nil || hostResult == "" {
		return false, false
	}
	if spec, ok := g.monoArraySpecForArrayTypeExpr(info.Package, retExpr); ok && spec != nil && info.ReturnType == spec.GoType {
		renderDirectHostSliceToMonoArrayReturn(buf, spec, hostResult)
		return true, true
	}
	union := g.nativeUnionInfoForGoType(info.ReturnType)
	if union == nil {
		return false, false
	}
	emitted := false
	for _, member := range union.Members {
		if member == nil {
			continue
		}
		spec, ok := g.monoArraySpecForGoType(member.GoType)
		if !ok || spec == nil {
			continue
		}
		if g.typeExprIncludesNilInPackage(info.Package, member.TypeExpr) && g.goTypeHasNilZeroValue(member.GoType) {
			fmt.Fprintf(buf, "\tif %s == nil {\n", hostResult)
			fmt.Fprintf(buf, "\t\treturn %s(nil), nil\n", member.WrapHelper)
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\tif __able_host_slice, ok := %s.([]%s); ok {\n", hostResult, spec.ElemGoType)
		fmt.Fprintf(buf, "\t\treturn %s(&%s{Elements: append([]%s(nil), __able_host_slice...)}), nil\n", member.WrapHelper, spec.GoName, spec.ElemGoType)
		fmt.Fprintf(buf, "\t}\n")
		emitted = true
	}
	return emitted, false
}

func renderDirectHostSliceToMonoArrayReturn(buf *bytes.Buffer, spec *monoArraySpec, hostResult string) {
	fmt.Fprintf(buf, "\treturn &%s{Elements: append([]%s(nil), %s...)}, nil\n", spec.GoName, spec.ElemGoType, hostResult)
}

func directHostArgExpr(goType string, hostType string, expr string) (string, bool) {
	goType = strings.TrimSpace(goType)
	hostType = strings.TrimSpace(hostType)
	if goType == "" || hostType == "" || expr == "" || goType != hostType {
		return "", false
	}
	if !isDirectHostScalarType(goType) {
		return "", false
	}
	return expr, true
}

func directHostReturnExpr(goType string, hostType string, expr string) (string, bool) {
	goType = strings.TrimSpace(goType)
	hostType = strings.TrimSpace(hostType)
	if goType == "" || hostType == "" || expr == "" || goType != hostType {
		return "", false
	}
	if !isDirectHostScalarType(goType) {
		return "", false
	}
	return expr, true
}

func isDirectHostScalarType(goType string) bool {
	switch goType {
	case "bool", "string", "rune",
		"float32", "float64",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}
