package compiler

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

type generator struct {
	opts              Options
	structs           map[string]*structInfo
	typeAliases           map[string]map[string]ast.TypeExpression
	typeAliasGenericParams map[string]map[string][]*ast.GenericParameter
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
	hasDynamicFeature  bool
	moduleBindings     map[string][]moduleBinding // package -> bindings
	evaluatedConstants map[string]*evaluatedConst // "pkg::name" -> value
}

type moduleBinding struct {
	Name    string
	GoValue string // Go literal expression (e.g., "uint64(14695981039346656037)")
	GoType  string // Go type (e.g., "uint64")
}

type evaluatedConst struct {
	val    *big.Int
	suffix string
}

type diagNodeInfo struct {
	Name       string
	GoType     string
	Span       ast.Span
	Origin     string
	CallName   string
	CallMember string
}

type implSiblingInfo struct {
	GoName string
	Arity  int
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
	implSiblings        map[string]implSiblingInfo
}

func newGenerator(opts Options) *generator {
	return &generator{
		opts:              opts,
		structs:           make(map[string]*structInfo),
		typeAliases:            make(map[string]map[string]ast.TypeExpression),
		typeAliasGenericParams: make(map[string]map[string][]*ast.GenericParameter),
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
	g.typeAliases = make(map[string]map[string]ast.TypeExpression)
	g.typeAliasGenericParams = make(map[string]map[string][]*ast.GenericParameter)
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
			alias, ok := stmt.(*ast.TypeAliasDefinition)
			if !ok || alias == nil || alias.ID == nil || strings.TrimSpace(alias.ID.Name) == "" || alias.TargetType == nil {
				continue
			}
			if g.typeAliases[module.Package] == nil {
				g.typeAliases[module.Package] = make(map[string]ast.TypeExpression)
			}
			if _, exists := g.typeAliases[module.Package][alias.ID.Name]; !exists {
				g.typeAliases[module.Package][alias.ID.Name] = alias.TargetType
				if len(alias.GenericParams) > 0 {
					if g.typeAliasGenericParams[module.Package] == nil {
						g.typeAliasGenericParams[module.Package] = make(map[string][]*ast.GenericParameter)
					}
					g.typeAliasGenericParams[module.Package][alias.ID.Name] = alias.GenericParams
				}
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
			case *ast.AssignmentExpression:
				if def == nil || def.Operator != ast.AssignmentDeclare {
					continue
				}
				g.collectModuleBinding(def, pkgName)
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
	case *ast.FunctionDefinition:
		return g.compileLocalFunctionDefinitionStatement(ctx, s)
	case *ast.StructDefinition:
		return g.compileLocalStructDefinitionStatement(ctx, s)
	case *ast.UnionDefinition:
		return g.compileLocalUnionDefinitionStatement(ctx, s)
	case *ast.InterfaceDefinition:
		return g.compileLocalInterfaceDefinitionStatement(ctx, s)
	case *ast.TypeAliasDefinition:
		return g.compileLocalTypeAliasDefinitionStatement(ctx, s)
	case *ast.MethodsDefinition:
		return g.compileLocalMethodsDefinitionStatement(ctx, s)
	case *ast.ImplementationDefinition:
		return g.compileLocalImplementationDefinitionStatement(ctx, s)
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

func (g *generator) collectModuleBinding(assign *ast.AssignmentExpression, pkgName string) {
	if g == nil || assign == nil || assign.Right == nil {
		return
	}
	name := ""
	switch lhs := assign.Left.(type) {
	case *ast.Identifier:
		if lhs != nil {
			name = lhs.Name
		}
	case *ast.TypedPattern:
		if lhs != nil {
			if ident, ok := lhs.Pattern.(*ast.Identifier); ok && ident != nil {
				name = ident.Name
			}
		}
	}
	if name == "" {
		return
	}
	// Try to evaluate as a constant integer first, to track for later resolution.
	if val, suffix, ok := g.evalConstInt(assign.Right, pkgName); ok {
		if g.evaluatedConstants == nil {
			g.evaluatedConstants = make(map[string]*evaluatedConst)
		}
		g.evaluatedConstants[pkgName+"::"+name] = &evaluatedConst{val: val, suffix: suffix}
	}
	goValue, goType := g.literalToRuntimeExpr(assign.Right, pkgName)
	if goValue == "" {
		return
	}
	if g.moduleBindings == nil {
		g.moduleBindings = make(map[string][]moduleBinding)
	}
	g.moduleBindings[pkgName] = append(g.moduleBindings[pkgName], moduleBinding{
		Name:    name,
		GoValue: goValue,
		GoType:  goType,
	})
}

func (g *generator) literalToRuntimeExpr(expr ast.Expression, pkgName string) (string, string) {
	switch lit := expr.(type) {
	case *ast.IntegerLiteral:
		if lit == nil || lit.Value == nil {
			return "", ""
		}
		valStr := lit.Value.String()
		bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", valStr)
		suffix := integerTypeSuffix(lit.IntegerType)
		return fmt.Sprintf("runtime.IntegerValue{Val: %s, TypeSuffix: %s}", bigExpr, suffix), "runtime.Value"
	case *ast.FloatLiteral:
		if lit == nil {
			return "", ""
		}
		suffix := "runtime.FloatF64"
		if lit.FloatType != nil && *lit.FloatType == ast.FloatTypeF32 {
			suffix = "runtime.FloatF32"
		}
		return fmt.Sprintf("runtime.FloatValue{Val: %v, TypeSuffix: %s}", lit.Value, suffix), "runtime.Value"
	case *ast.StringLiteral:
		if lit == nil {
			return "", ""
		}
		return fmt.Sprintf("runtime.StringValue{Val: %q}", lit.Value), "runtime.Value"
	case *ast.BooleanLiteral:
		if lit == nil {
			return "", ""
		}
		return fmt.Sprintf("runtime.BoolValue{Val: %t}", lit.Value), "runtime.Value"
	case *ast.UnaryExpression:
		if lit == nil || lit.Operator != "-" {
			return "", ""
		}
		val, suffix, ok := g.evalConstInt(lit, pkgName)
		if ok {
			bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", val.String())
			return fmt.Sprintf("runtime.IntegerValue{Val: %s, TypeSuffix: %s}", bigExpr, suffix), "runtime.Value"
		}
		return "", ""
	case *ast.BinaryExpression:
		if lit == nil {
			return "", ""
		}
		val, suffix, ok := g.evalConstInt(lit, pkgName)
		if !ok {
			return "", ""
		}
		bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", val.String())
		return fmt.Sprintf("runtime.IntegerValue{Val: %s, TypeSuffix: %s}", bigExpr, suffix), "runtime.Value"
	}
	return "", ""
}

// evalConstInt evaluates a constant integer expression at compile time.
// Handles integer literals, negated integer literals, binary arithmetic/bitwise,
// and identifier references to previously evaluated constants in the same package.
func (g *generator) evalConstInt(expr ast.Expression, pkgName string) (*big.Int, string, bool) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		if e == nil || e.Value == nil {
			return nil, "", false
		}
		return new(big.Int).Set(e.Value), integerTypeSuffix(e.IntegerType), true
	case *ast.Identifier:
		if e == nil || g == nil || g.evaluatedConstants == nil {
			return nil, "", false
		}
		key := pkgName + "::" + e.Name
		if c, ok := g.evaluatedConstants[key]; ok {
			return new(big.Int).Set(c.val), c.suffix, true
		}
		return nil, "", false
	case *ast.UnaryExpression:
		if e == nil || e.Operator != "-" {
			return nil, "", false
		}
		val, suffix, ok := g.evalConstInt(e.Operand, pkgName)
		if !ok {
			return nil, "", false
		}
		return val.Neg(val), suffix, true
	case *ast.BinaryExpression:
		if e == nil {
			return nil, "", false
		}
		left, lSuffix, lok := g.evalConstInt(e.Left, pkgName)
		right, _, rok := g.evalConstInt(e.Right, pkgName)
		if !lok || !rok {
			return nil, "", false
		}
		switch e.Operator {
		case "+":
			return left.Add(left, right), lSuffix, true
		case "-":
			return left.Sub(left, right), lSuffix, true
		case "*":
			return left.Mul(left, right), lSuffix, true
		case "/":
			if right.Sign() == 0 {
				return nil, "", false
			}
			return left.Div(left, right), lSuffix, true
		case ".<<", "<<":
			shift := right.Int64()
			if shift < 0 || shift > 64 {
				return nil, "", false
			}
			return left.Lsh(left, uint(shift)), lSuffix, true
		case ".>>", ">>":
			shift := right.Int64()
			if shift < 0 || shift > 64 {
				return nil, "", false
			}
			return left.Rsh(left, uint(shift)), lSuffix, true
		case ".&", "&":
			return left.And(left, right), lSuffix, true
		case ".|", "|":
			return left.Or(left, right), lSuffix, true
		case ".^", "^":
			return left.Xor(left, right), lSuffix, true
		}
		return nil, "", false
	}
	return nil, "", false
}

func integerTypeSuffix(intType *ast.IntegerType) string {
	if intType != nil {
		switch *intType {
		case ast.IntegerTypeU64:
			return "runtime.IntegerU64"
		case ast.IntegerTypeI64:
			return "runtime.IntegerI64"
		case ast.IntegerTypeU32:
			return "runtime.IntegerU32"
		case ast.IntegerTypeI32:
			return "runtime.IntegerI32"
		case ast.IntegerTypeU16:
			return "runtime.IntegerU16"
		case ast.IntegerTypeI16:
			return "runtime.IntegerI16"
		case ast.IntegerTypeU8:
			return "runtime.IntegerU8"
		case ast.IntegerTypeI8:
			return "runtime.IntegerI8"
		case ast.IntegerTypeI128:
			return "runtime.IntegerI128"
		case ast.IntegerTypeU128:
			return "runtime.IntegerU128"
		}
	}
	return "runtime.IntegerI32"
}
