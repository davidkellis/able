package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

type generator struct {
	opts              Options
	structs           map[string]*structInfo
	unions            map[string]*ast.UnionDefinition
	unionPackages     map[string]string
	interfaces        map[string]*ast.InterfaceDefinition
	interfacePackages map[string]string
	staticImports     map[string][]staticImportBinding
	functions         map[string]map[string]*functionInfo
	overloads         map[string]map[string]*overloadInfo
	packages          []string
	entryPackage      string
	methods           map[string]map[string][]*methodInfo
	methodList        []*methodInfo
	implMethodList    []*implMethodInfo
	implDefinitions   []*implDefinitionInfo
	implMethodByInfo  map[*functionInfo]*implMethodInfo
	warnings          []string
	fallbacks         []FallbackInfo
	mangler           *nameMangler
	needsAst          bool
	needsIterator     bool
	awaitExprs        []string
	awaitNames        map[*ast.AwaitExpression]string
	diagNodes         []diagNodeInfo
	diagNames         map[ast.Node]string
	nodeOrigins       map[ast.Node]string
	packageEnvVars    map[string]string
	packageEnvOrder   []string
	hasDynamicFeature bool
}

type diagNodeInfo struct {
	Name       string
	GoType     string
	Span       ast.Span
	Origin     string
	CallName   string
	CallMember string
}

type compileContext struct {
	params              map[string]paramInfo
	locals              map[string]paramInfo
	functions           map[string]*functionInfo
	overloads           map[string]*overloadInfo
	packageName         string
	parent              *compileContext
	temps               *int
	reason              string
	loopDepth           int
	rethrowVar          string
	rethrowErrVar       string
	breakpoints         map[string]int
	implicitReceiver    paramInfo
	hasImplicitReceiver bool
	placeholderParams   map[int]paramInfo
	inPlaceholder       bool
	returnType          string
	returnTypeExpr      ast.TypeExpression
	genericNames        map[string]struct{}
}

func newGenerator(opts Options) *generator {
	return &generator{
		opts:              opts,
		structs:           make(map[string]*structInfo),
		unions:            make(map[string]*ast.UnionDefinition),
		unionPackages:     make(map[string]string),
		interfaces:        make(map[string]*ast.InterfaceDefinition),
		interfacePackages: make(map[string]string),
		staticImports:     make(map[string][]staticImportBinding),
		functions:         make(map[string]map[string]*functionInfo),
		overloads:         make(map[string]map[string]*overloadInfo),
		methods:           make(map[string]map[string][]*methodInfo),
		mangler:           newNameMangler(),
		awaitNames:        make(map[*ast.AwaitExpression]string),
		implMethodByInfo:  make(map[*functionInfo]*implMethodInfo),
	}
}

func (g *generator) setDynamicFeatureReport(report *DynamicFeatureReport) {
	if g == nil {
		return
	}
	g.hasDynamicFeature = report != nil && report.UsesDynamic()
}

func (g *generator) ensurePackageEnvVars() {
	if g.packageEnvVars != nil {
		return
	}
	names := g.collectPackageNames()
	g.packageEnvVars = make(map[string]string, len(names))
	g.packageEnvOrder = names
	for idx, name := range names {
		g.packageEnvVars[name] = fmt.Sprintf("__able_pkg_env_%d", idx)
	}
}

func (g *generator) packageEnvVar(name string) (string, bool) {
	if g == nil {
		return "", false
	}
	g.ensurePackageEnvVars()
	envVar, ok := g.packageEnvVars[name]
	return envVar, ok
}

func (g *generator) collectPackageNames() []string {
	seen := make(map[string]struct{})
	var names []string
	add := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for _, name := range g.packages {
		add(name)
	}
	add(g.entryPackage)
	if len(names) == 0 {
		for pkg := range g.functions {
			add(pkg)
		}
		for pkg := range g.overloads {
			add(pkg)
		}
	}
	sort.Strings(names)
	return names
}

func (g *generator) collect(program *driver.Program) error {
	if program == nil || program.Entry == nil || program.Entry.AST == nil {
		return fmt.Errorf("compiler: missing entry module")
	}
	g.entryPackage = program.Entry.Package
	g.packages = nil
	g.staticImports = make(map[string][]staticImportBinding)
	g.unions = make(map[string]*ast.UnionDefinition)
	g.unionPackages = make(map[string]string)
	g.interfaces = make(map[string]*ast.InterfaceDefinition)
	g.interfacePackages = make(map[string]string)
	if g.nodeOrigins == nil {
		g.nodeOrigins = make(map[ast.Node]string)
	}
	if program.Entry.NodeOrigins != nil {
		for node, origin := range program.Entry.NodeOrigins {
			if _, exists := g.nodeOrigins[node]; !exists {
				g.nodeOrigins[node] = origin
			}
		}
	}
	for _, module := range program.Modules {
		if module == nil || module.NodeOrigins == nil {
			continue
		}
		for node, origin := range module.NodeOrigins {
			if _, exists := g.nodeOrigins[node]; !exists {
				g.nodeOrigins[node] = origin
			}
		}
	}
	modules := make([]*driver.Module, 0, len(program.Modules)+1)
	if program.Entry != nil {
		modules = append(modules, program.Entry)
	}
	modules = append(modules, program.Modules...)
	seenModules := make(map[*driver.Module]struct{})
	uniqueModules := make([]*driver.Module, 0, len(modules))
	for _, module := range modules {
		if module == nil || module.AST == nil {
			continue
		}
		if _, ok := seenModules[module]; ok {
			continue
		}
		seenModules[module] = struct{}{}
		uniqueModules = append(uniqueModules, module)
	}

	for _, module := range uniqueModules {
		for _, stmt := range module.AST.Body {
			def, ok := stmt.(*ast.UnionDefinition)
			if !ok || def == nil || def.ID == nil {
				continue
			}
			name := def.ID.Name
			if name == "" {
				continue
			}
			if _, exists := g.unions[name]; exists {
				continue
			}
			g.unions[name] = def
			g.unionPackages[name] = module.Package
		}
	}

	for _, module := range uniqueModules {
		for _, stmt := range module.AST.Body {
			def, ok := stmt.(*ast.InterfaceDefinition)
			if !ok || def == nil || def.ID == nil {
				continue
			}
			name := def.ID.Name
			if name == "" {
				continue
			}
			if _, exists := g.interfaces[name]; exists {
				continue
			}
			g.interfaces[name] = def
			if g.interfacePackages != nil {
				g.interfacePackages[name] = module.Package
			}
		}
	}

	for _, module := range uniqueModules {
		for _, stmt := range module.AST.Body {
			def, ok := stmt.(*ast.StructDefinition)
			if !ok || def == nil || def.ID == nil {
				continue
			}
			name := def.ID.Name
			if name == "" {
				continue
			}
			if _, exists := g.structs[name]; exists {
				g.warnings = append(g.warnings, fmt.Sprintf("compiler: duplicate struct %s; skipping", name))
				continue
			}
			goName := g.mangler.unique(exportIdent(name))
			g.structs[name] = &structInfo{
				Name:    name,
				Package: module.Package,
				GoName:  goName,
				Kind:    def.Kind,
				Node:    def,
			}
		}
	}

	for _, info := range g.structs {
		mapper := NewTypeMapper(g.structs, info.Package)
		fields := make([]fieldInfo, 0, len(info.Node.Fields))
		supported := true
		for idx, field := range info.Node.Fields {
			fieldName := ""
			if field.Name != nil {
				fieldName = field.Name.Name
			} else {
				fieldName = fmt.Sprintf("field_%d", idx+1)
			}
			goFieldName := exportIdent(fieldName)
			goType, ok := mapper.Map(field.FieldType)
			if !ok {
				supported = false
			}
			fields = append(fields, fieldInfo{
				Name:      fieldName,
				GoName:    goFieldName,
				GoType:    goType,
				Supported: ok,
			})
		}
		info.Fields = fields
		info.Supported = supported
	}

	seenPackages := make(map[string]struct{})
	for _, module := range uniqueModules {
		g.collectStaticImportsForPackage(module.Package, module.AST.Imports)
		pkgName := module.Package
		if _, ok := seenPackages[pkgName]; !ok {
			seenPackages[pkgName] = struct{}{}
			g.packages = append(g.packages, pkgName)
		}
		mapper := NewTypeMapper(g.structs, pkgName)

		functions := make(map[string][]*ast.FunctionDefinition)
		for _, stmt := range module.AST.Body {
			switch def := stmt.(type) {
			case *ast.FunctionDefinition:
				if def == nil || def.ID == nil {
					continue
				}
				name := def.ID.Name
				if name == "" {
					continue
				}
				functions[name] = append(functions[name], def)
			case *ast.MethodsDefinition:
				if def == nil {
					continue
				}
				g.collectMethodsDefinition(def, mapper, pkgName)
			case *ast.ImplementationDefinition:
				if def == nil {
					continue
				}
				g.collectImplDefinition(def, mapper, pkgName)
			}
		}

		if g.functions[pkgName] == nil {
			g.functions[pkgName] = make(map[string]*functionInfo)
		}
		if g.overloads[pkgName] == nil {
			g.overloads[pkgName] = make(map[string]*overloadInfo)
		}

		for name, defs := range functions {
			qualified := qualifiedName(pkgName, name)
			if len(defs) != 1 {
				entries := make([]*functionInfo, 0, len(defs))
				minArity := -1
				for idx, def := range defs {
					if def == nil {
						continue
					}
					info := &functionInfo{
						Name:          name,
						Package:       pkgName,
						QualifiedName: qualified,
						GoName:        g.mangler.unique(fmt.Sprintf("fn_%s_overload_%d", sanitizeIdent(name), idx)),
						Definition:    def,
						HasOriginal:   true,
					}
					g.fillFunctionInfo(info, mapper)
					entries = append(entries, info)
					if arity := minArgsForDefinition(def); arity >= 0 {
						if minArity < 0 || arity < minArity {
							minArity = arity
						}
					}
				}
				if len(entries) > 0 {
					g.overloads[pkgName][name] = &overloadInfo{
						Name:          name,
						Package:       pkgName,
						QualifiedName: qualified,
						Entries:       entries,
						MinArity:      minArity,
					}
				}
				continue
			}
			info := &functionInfo{
				Name:          name,
				Package:       pkgName,
				QualifiedName: qualified,
				GoName:        g.mangler.unique("fn_" + sanitizeIdent(name)),
				Definition:    defs[0],
				HasOriginal:   true,
			}
			g.fillFunctionInfo(info, mapper)
			g.functions[pkgName][name] = info
		}
	}
	g.collectDefaultImplMethods()
	sort.Strings(g.packages)
	g.resolveCompileableFunctions()
	g.resolveCompileableMethods()
	g.detectAstNeeds()
	return nil
}

func (g *generator) diagNodeName(node ast.Node, goType string, prefix string) string {
	if node == nil {
		return "nil"
	}
	if g.diagNames == nil {
		g.diagNames = make(map[ast.Node]string)
	}
	if name, ok := g.diagNames[node]; ok {
		return name
	}
	name := fmt.Sprintf("__able_%s_node_%d", prefix, len(g.diagNodes))
	info := diagNodeInfo{
		Name:   name,
		GoType: goType,
		Span:   node.Span(),
	}
	if call, ok := node.(*ast.FunctionCall); ok && call != nil {
		switch callee := call.Callee.(type) {
		case *ast.Identifier:
			info.CallName = callee.Name
		case *ast.MemberAccessExpression:
			if member, ok := callee.Member.(*ast.Identifier); ok && member != nil {
				info.CallMember = member.Name
			}
		}
	}
	if g.nodeOrigins != nil {
		if origin, ok := g.nodeOrigins[node]; ok {
			info.Origin = origin
		}
	}
	g.diagNodes = append(g.diagNodes, info)
	g.diagNames[node] = name
	g.needsAst = true
	return name
}

func (g *generator) fillFunctionInfo(info *functionInfo, mapper *TypeMapper) {
	if info == nil || info.Definition == nil {
		return
	}
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	if def.IsMethodShorthand {
		supported = false
	}
	for idx, param := range def.Params {
		name := fmt.Sprintf("arg%d", idx)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			supported = false
		}
		goName := safeParamName(name, idx)
		goType, ok := mapper.Map(param.ParamType)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    goName,
			GoType:    goType,
			TypeExpr:  param.ParamType,
			Supported: ok,
		})
	}
	retType, ok := mapper.Map(def.ReturnType)
	if !ok || retType == "" {
		supported = false
	}
	info.Params = params
	info.ReturnType = retType
	info.SupportedTypes = supported
	info.Arity = len(params)

	if !supported {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
	}
}

func (g *generator) bodyCompileable(info *functionInfo, retType string) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	def := info.Definition
	if def.Body == nil {
		info.Reason = "missing function body"
		return false
	}
	ctx := newCompileContext(info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
	_, _, ok := g.compileBody(ctx, info)
	if !ok {
		info.Reason = ctx.reason
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
	}
	return ok
}

func (g *generator) resolveCompileableFunctions() {
	pending := make(map[*functionInfo]struct{})
	for _, info := range g.allFunctionInfos() {
		if info == nil {
			continue
		}
		if !info.SupportedTypes {
			info.Compileable = false
			continue
		}
		pending[info] = struct{}{}
	}
	for {
		progress := false
		for info := range pending {
			if info.Compileable {
				delete(pending, info)
				continue
			}
			if ok := g.bodyCompileable(info, info.ReturnType); ok {
				info.Compileable = true
				info.Reason = ""
				progress = true
				delete(pending, info)
			}
		}
		if !progress {
			break
		}
	}
	for info := range pending {
		if info == nil {
			continue
		}
		if info.Reason == "" {
			info.Reason = "unsupported function body"
		}
		info.Compileable = false
	}
}

func (g *generator) collectFallbacks() []FallbackInfo {
	if g == nil {
		return nil
	}
	fallbacks := make([]FallbackInfo, 0, len(g.fallbacks))
	fallbacks = append(fallbacks, g.fallbacks...)
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || info.Compileable {
			continue
		}
		reason := info.Reason
		if reason == "" {
			reason = "unsupported function body"
		}
		name := info.Name
		if info.QualifiedName != "" {
			name = info.QualifiedName
		}
		fallbacks = append(fallbacks, FallbackInfo{
			Name:   name,
			Reason: reason,
		})
	}
	return fallbacks
}

func (g *generator) compileBody(ctx *compileContext, info *functionInfo) ([]string, string, bool) {
	if info == nil || info.Definition == nil || info.Definition.Body == nil {
		ctx.setReason("missing function body")
		return nil, "", false
	}
	statements := info.Definition.Body.Body
	if len(statements) == 0 {
		if g.isVoidType(info.ReturnType) {
			return nil, "struct{}{}", true
		}
		ctx.setReason("empty body requires void return")
		return nil, "", false
	}
	lines := make([]string, 0, len(statements))
	for idx, stmt := range statements {
		isLast := idx == len(statements)-1
		if ret, ok := stmt.(*ast.ReturnStatement); ok {
			return g.compileReturnStatement(ctx, info.ReturnType, ret, lines)
		}
		if isLast {
			if raiseStmt, ok := stmt.(*ast.RaiseStatement); ok {
				stmtLines, ok := g.compileRaiseStatement(ctx, raiseStmt)
				if !ok {
					return nil, "", false
				}
				lines = append(lines, stmtLines...)
				retExpr, ok := g.zeroValueExpr(info.ReturnType)
				if !ok {
					ctx.setReason("missing return expression")
					return nil, "", false
				}
				return lines, retExpr, true
			}
			if rethrowStmt, ok := stmt.(*ast.RethrowStatement); ok {
				stmtLines, ok := g.compileRethrowStatement(ctx, rethrowStmt)
				if !ok {
					return nil, "", false
				}
				lines = append(lines, stmtLines...)
				retExpr, ok := g.zeroValueExpr(info.ReturnType)
				if !ok {
					ctx.setReason("missing return expression")
					return nil, "", false
				}
				return lines, retExpr, true
			}
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				return g.compileImplicitReturn(ctx, info.ReturnType, expr, lines)
			}
			if g.isVoidType(info.ReturnType) {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", false
				}
				lines = append(lines, stmtLines...)
				return lines, "struct{}{}", true
			}
			ctx.setReason("missing return expression")
			return nil, "", false
		}
		stmtLines, ok := g.compileStatement(ctx, stmt)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
	}
	ctx.setReason("missing return expression")
	return nil, "", false
}

func (g *generator) compileReturnStatement(ctx *compileContext, returnType string, ret *ast.ReturnStatement, lines []string) ([]string, string, bool) {
	if ret == nil {
		ctx.setReason("missing return")
		return nil, "", false
	}
	if ret.Argument == nil {
		if g.isVoidType(returnType) {
			return lines, "struct{}{}", true
		}
		if returnType == "runtime.Value" && g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
			return lines, "runtime.VoidValue{}", true
		}
		ctx.setReason("missing return expression")
		return nil, "", false
	}
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", ret.Argument)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
		}
		return lines, "struct{}{}", true
	}
	exprLines, expr, exprType, ok := g.compileTailExpression(ctx, returnType, ret.Argument)
	if !ok {
		return nil, "", false
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			if exprType != "runtime.Value" {
				converted, ok := g.runtimeValueExpr(expr, exprType)
				if !ok {
					ctx.setReason("return type mismatch")
					return nil, "", false
				}
				expr = converted
			}
			coerced, ok := g.interfaceReturnExpr(expr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			expr = coerced
		}
	}
	lines = append(lines, exprLines...)
	return lines, expr, true
}

func (g *generator) compileImplicitReturn(ctx *compileContext, returnType string, expr ast.Expression, lines []string) ([]string, string, bool) {
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", expr)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
		}
		return lines, "struct{}{}", true
	}
	stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, expr)
	if !ok {
		return nil, "", false
	}
	if !g.typeMatches(returnType, valueType) {
		ctx.setReason("assignment return type mismatch")
		return nil, "", false
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			if valueType != "runtime.Value" {
				converted, ok := g.runtimeValueExpr(valueExpr, valueType)
				if !ok {
					ctx.setReason("return type mismatch")
					return nil, "", false
				}
				valueExpr = converted
			}
			coerced, ok := g.interfaceReturnExpr(valueExpr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			valueExpr = coerced
		}
	}
	lines = append(lines, stmtLines...)
	return lines, valueExpr, true
}

func (g *generator) compileStatement(ctx *compileContext, stmt ast.Statement) ([]string, bool) {
	if stmt == nil {
		ctx.setReason("missing statement")
		return nil, false
	}
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		lines, valueExpr, _, ok := g.compileAssignment(ctx, s)
		if !ok {
			return nil, false
		}
		if valueExpr != "" {
			lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
		}
		return lines, true
	case *ast.BinaryExpression:
		if s.Operator == "|>" || s.Operator == "|>>" {
			if assign, ok := s.Left.(*ast.AssignmentExpression); ok && assign != nil && assign.Operator == ast.AssignmentDeclare {
				if name, _, ok := g.assignmentTargetName(assign.Left); ok && name != "" {
					assignLines, _, _, ok := g.compileAssignment(ctx, assign)
					if !ok {
						return nil, false
					}
					pipeExpr := ast.NewBinaryExpression(s.Operator, ast.NewIdentifier(name), s.Right)
					pipeValue, _, ok := g.compilePipeExpression(ctx, pipeExpr, "")
					if !ok {
						return nil, false
					}
					lines := append([]string{}, assignLines...)
					if pipeValue != "" {
						lines = append(lines, fmt.Sprintf("_ = %s", pipeValue))
					}
					return lines, true
				}
			}
		}
		if expr, ok := stmt.(ast.Expression); ok {
			valueLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", expr)
			if !ok {
				return nil, false
			}
			if valueExpr == "" {
				return valueLines, true
			}
			lines := append(valueLines, fmt.Sprintf("_ = %s", valueExpr))
			return lines, true
		}
		ctx.setReason("unsupported statement")
		return nil, false
	case *ast.WhileLoop:
		return g.compileWhileLoop(ctx, s)
	case *ast.ForLoop:
		return g.compileForLoop(ctx, s)
	case *ast.BreakStatement:
		return g.compileBreakStatement(ctx, s)
	case *ast.ContinueStatement:
		return g.compileContinueStatement(ctx, s)
	case *ast.RaiseStatement:
		return g.compileRaiseStatement(ctx, s)
	case *ast.RethrowStatement:
		return g.compileRethrowStatement(ctx, s)
	case *ast.YieldStatement:
		return g.compileYieldStatement(ctx, s)
	case *ast.ReturnStatement:
		if ctx == nil || ctx.returnType == "" {
			ctx.setReason("return outside function")
			return nil, false
		}
		if s.Argument == nil {
			if !g.isVoidType(ctx.returnType) {
				if ctx.returnType == "runtime.Value" && g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
					return []string{"panic(__able_return{value: runtime.VoidValue{}})"}, true
				}
				expected := typeExpressionToString(ctx.returnTypeExpr)
				if expected == "" || expected == "<?>" {
					expected = typeNameFromGoType(ctx.returnType)
				}
				nodeName := g.diagNodeName(s, "*ast.ReturnStatement", "return")
				return []string{fmt.Sprintf("__able_raise_return_type_mismatch(%s, %q, %q)", nodeName, expected, "void")}, true
			}
			return []string{"panic(__able_return{value: struct{}{}})"}, true
		}
		if g.isVoidType(ctx.returnType) {
			lines := []string{}
			stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", s.Argument)
			if !ok {
				return nil, false
			}
			lines = append(lines, stmtLines...)
			if valueExpr != "" {
				lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
			}
			lines = append(lines, "panic(__able_return{value: struct{}{}})")
			return lines, true
		}
		stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, ctx.returnType, s.Argument)
		if !ok {
			return nil, false
		}
		if !g.typeMatches(ctx.returnType, valueType) {
			ctx.setReason("return type mismatch")
			return nil, false
		}
		lines := append([]string{}, stmtLines...)
		lines = append(lines, fmt.Sprintf("panic(__able_return{value: %s})", valueExpr))
		return lines, true
	case *ast.IfExpression:
		return g.compileIfStatement(ctx, s)
	case *ast.MatchExpression:
		return g.compileMatchStatement(ctx, s)
	case *ast.BlockExpression:
		return g.compileBlockStatement(ctx, s)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			valueLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", expr)
			if !ok {
				return nil, false
			}
			if valueExpr == "" {
				return valueLines, true
			}
			lines := append(valueLines, fmt.Sprintf("_ = %s", valueExpr))
			return lines, true
		}
		ctx.setReason("unsupported statement")
		return nil, false
	}
}

func newCompileContext(info *functionInfo, functions map[string]*functionInfo, overloads map[string]*overloadInfo, packageName string, genericNames map[string]struct{}) *compileContext {
	counter := 0
	ctx := &compileContext{
		params:       make(map[string]paramInfo),
		locals:       make(map[string]paramInfo),
		functions:    functions,
		overloads:    overloads,
		packageName:  packageName,
		temps:        &counter,
		loopDepth:    0,
		breakpoints:  make(map[string]int),
		genericNames: genericNames,
	}
	if info != nil {
		ctx.returnType = info.ReturnType
		if info.Definition != nil {
			ctx.returnTypeExpr = info.Definition.ReturnType
		}
		for _, param := range info.Params {
			if param.Name == "" {
				continue
			}
			ctx.params[param.Name] = param
		}
		if len(info.Params) > 0 {
			ctx.implicitReceiver = info.Params[0]
			ctx.hasImplicitReceiver = true
		}
	}
	return ctx
}

func qualifiedName(pkg string, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func (c *compileContext) setReason(reason string) {
	if c == nil || reason == "" {
		return
	}
	if c.reason == "" {
		c.reason = reason
	}
}

func (c *compileContext) lookup(name string) (paramInfo, bool) {
	if c == nil {
		return paramInfo{}, false
	}
	if local, ok := c.locals[name]; ok {
		return local, true
	}
	if c.parent != nil {
		return c.parent.lookup(name)
	}
	if param, ok := c.params[name]; ok {
		return param, true
	}
	return paramInfo{}, false
}

func (c *compileContext) lookupCurrent(name string) (paramInfo, bool) {
	if c == nil {
		return paramInfo{}, false
	}
	if local, ok := c.locals[name]; ok {
		return local, true
	}
	if c.parent == nil {
		if param, ok := c.params[name]; ok {
			return param, true
		}
	}
	return paramInfo{}, false
}

func (c *compileContext) child() *compileContext {
	if c == nil {
		return nil
	}
	return &compileContext{
		locals:              make(map[string]paramInfo),
		functions:           c.functions,
		overloads:           c.overloads,
		packageName:         c.packageName,
		parent:              c,
		temps:               c.temps,
		loopDepth:           c.loopDepth,
		rethrowVar:          c.rethrowVar,
		rethrowErrVar:       c.rethrowErrVar,
		breakpoints:         c.breakpoints,
		implicitReceiver:    c.implicitReceiver,
		hasImplicitReceiver: c.hasImplicitReceiver,
		placeholderParams:   c.placeholderParams,
		inPlaceholder:       c.inPlaceholder,
		returnType:          c.returnType,
		returnTypeExpr:      c.returnTypeExpr,
		genericNames:        c.genericNames,
	}
}

func (c *compileContext) pushBreakpoint(label string) {
	if c == nil || label == "" {
		return
	}
	if c.breakpoints == nil {
		c.breakpoints = make(map[string]int)
	}
	c.breakpoints[label]++
}

func (c *compileContext) popBreakpoint(label string) {
	if c == nil || label == "" || c.breakpoints == nil {
		return
	}
	count := c.breakpoints[label]
	if count <= 1 {
		delete(c.breakpoints, label)
		return
	}
	c.breakpoints[label] = count - 1
}

func (c *compileContext) hasBreakpoint(label string) bool {
	if c == nil || label == "" || c.breakpoints == nil {
		return false
	}
	return c.breakpoints[label] > 0
}

func (c *compileContext) newTemp() string {
	if c == nil || c.temps == nil {
		return "__able_tmp"
	}
	for {
		name := fmt.Sprintf("__able_tmp_%d", *c.temps)
		*c.temps++
		if _, exists := c.lookup(name); !exists {
			return name
		}
	}
}
