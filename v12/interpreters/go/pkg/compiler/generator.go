package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

type generator struct {
	opts                      Options
	structs                   map[string]*structInfo
	specializedStructs        map[string]*structInfo
	typeAliases               map[string]map[string]ast.TypeExpression
	typeAliasGenericParams    map[string]map[string][]*ast.GenericParameter
	unions                    map[string]*ast.UnionDefinition
	unionPackages             map[string]string
	nativeUnions              map[string]*nativeUnionInfo
	nativeCallables           map[string]*nativeCallableInfo
	nativeInterfaces          map[string]*nativeInterfaceInfo
	iteratorCollectMonoArrays map[string]*iteratorCollectMonoArrayInfo
	monoArraySpecs            map[string]*monoArraySpec
	nativeInterfaceBuilding   map[string]struct{}
	nativeInterfaceRefreshing map[string]struct{}
	interfaces                map[string]*ast.InterfaceDefinition
	interfacePackages         map[string]string
	staticImports             map[string][]staticImportBinding
	functions                 map[string]map[string]*functionInfo
	overloads                 map[string]map[string]*overloadInfo
	packages                  []string
	entryPackage              string
	methods                   map[string]map[string][]*methodInfo
	methodList                []*methodInfo
	implMethodList            []*implMethodInfo
	implDefinitions           []*implDefinitionInfo
	implMethodByInfo          map[*functionInfo]*implMethodInfo
	specializedFunctions      []*functionInfo
	specializedFunctionIndex  map[string]*functionInfo
	warnings                  []string
	fallbacks                 []FallbackInfo
	mangler                   *nameMangler
	needsAst                  bool
	needsIterator             bool
	needsStrconv              bool
	needsStringFromByteArray  bool
	awaitExprs                []string
	awaitNames                map[*ast.AwaitExpression]string
	diagNodes                 []diagNodeInfo
	diagNames                 map[ast.Node]string
	nodeOrigins               map[ast.Node]string
	packageEnvVars            map[string]string
	packageEnvOrder           []string
	hasDynamicFeature         bool
	moduleBindings            map[string][]moduleBinding // package -> bindings
	moduleBindingNames        map[string]map[string]struct{}
	evaluatedConstants        map[string]*evaluatedConst // "pkg::name" -> value
	staticCallableNames       map[string]map[string]struct{}
	externCallables           map[string]map[string]struct{}
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
	Info   *functionInfo
}

type compileContext struct {
	params                map[string]paramInfo
	locals                map[string]paramInfo
	integerFacts          map[string]integerFact
	functions             map[string]*functionInfo
	overloads             map[string]*overloadInfo
	packageName           string
	parent                *compileContext
	temps                 *int
	reason                string
	loopDepth             int
	loopLabel             string
	loopBreakValueTemp    string
	rethrowVar            string
	rethrowErrVar         string
	breakpoints           map[string]int
	breakpointGoLabels    map[string]string
	breakpointResultTemps map[string]string
	implicitReceiver      paramInfo
	hasImplicitReceiver   bool
	placeholderParams     map[int]paramInfo
	inPlaceholder         bool
	returnType            string
	returnTypeExpr        ast.TypeExpression
	expectedTypeExpr      ast.TypeExpression
	controlMode           string
	controlCaptureVar     string
	controlCaptureLabel   string
	controlCaptureBreak   bool
	rethrowControlVar     string
	genericNames          map[string]struct{}
	typeBindings          map[string]ast.TypeExpression
	implSiblings          map[string]implSiblingInfo
	originExtractions     map[string]string // CSE cache: Able variable name → Go extraction temp
}

func newGenerator(opts Options) *generator {
	return &generator{
		opts:                      opts,
		structs:                   make(map[string]*structInfo),
		specializedStructs:        make(map[string]*structInfo),
		typeAliases:               make(map[string]map[string]ast.TypeExpression),
		typeAliasGenericParams:    make(map[string]map[string][]*ast.GenericParameter),
		unions:                    make(map[string]*ast.UnionDefinition),
		unionPackages:             make(map[string]string),
		nativeUnions:              make(map[string]*nativeUnionInfo),
		nativeCallables:           make(map[string]*nativeCallableInfo),
		nativeInterfaces:          make(map[string]*nativeInterfaceInfo),
		iteratorCollectMonoArrays: make(map[string]*iteratorCollectMonoArrayInfo),
		monoArraySpecs:            make(map[string]*monoArraySpec),
		nativeInterfaceBuilding:   make(map[string]struct{}),
		nativeInterfaceRefreshing: make(map[string]struct{}),
		interfaces:                make(map[string]*ast.InterfaceDefinition),
		interfacePackages:         make(map[string]string),
		staticImports:             make(map[string][]staticImportBinding),
		functions:                 make(map[string]map[string]*functionInfo),
		overloads:                 make(map[string]map[string]*overloadInfo),
		methods:                   make(map[string]map[string][]*methodInfo),
		mangler:                   newNameMangler(),
		awaitNames:                make(map[*ast.AwaitExpression]string),
		implMethodByInfo:          make(map[*functionInfo]*implMethodInfo),
		specializedFunctionIndex:  make(map[string]*functionInfo),
		moduleBindingNames:        make(map[string]map[string]struct{}),
		externCallables:           make(map[string]map[string]struct{}),
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
	g.specializedStructs = make(map[string]*structInfo)
	g.typeAliases = make(map[string]map[string]ast.TypeExpression)
	g.typeAliasGenericParams = make(map[string]map[string][]*ast.GenericParameter)
	g.unions = make(map[string]*ast.UnionDefinition)
	g.unionPackages = make(map[string]string)
	g.nativeUnions = make(map[string]*nativeUnionInfo)
	g.nativeCallables = make(map[string]*nativeCallableInfo)
	g.nativeInterfaces = make(map[string]*nativeInterfaceInfo)
	g.iteratorCollectMonoArrays = make(map[string]*iteratorCollectMonoArrayInfo)
	g.monoArraySpecs = make(map[string]*monoArraySpec)
	g.nativeInterfaceBuilding = make(map[string]struct{})
	g.nativeInterfaceRefreshing = make(map[string]struct{})
	g.interfaces = make(map[string]*ast.InterfaceDefinition)
	g.interfacePackages = make(map[string]string)
	g.staticCallableNames = nil
	g.externCallables = make(map[string]map[string]struct{})
	g.specializedFunctions = nil
	g.specializedFunctionIndex = make(map[string]*functionInfo)
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
			qualified := qualifiedName(module.Package, name)
			if _, exists := g.structs[qualified]; exists {
				continue
			}
			goName := g.mangler.unique(exportIdent(name))
			g.structs[qualified] = &structInfo{
				Name:     name,
				Package:  module.Package,
				GoName:   goName,
				TypeExpr: genericStructTypeExprForDefinition(def),
				Kind:     def.Kind,
				Node:     def,
			}
		}
	}
	g.ensureBuiltinArrayStruct()

	for _, info := range g.structs {
		mapper := NewTypeMapper(g, info.Package)
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
				TypeExpr:  normalizeTypeExprForPackage(g, info.Package, field.FieldType),
				Supported: ok,
			})
		}
		info.Fields = fields
		info.Supported = supported
		// Override Array so compiled static code keeps native slice-backed storage
		// while preserving the spec-visible metadata fields.
		if info.Name == "Array" {
			info.Fields = []fieldInfo{
				{
					Name:      "length",
					GoName:    "Length",
					GoType:    "int32",
					TypeExpr:  ast.Ty("i32"),
					Supported: true,
				},
				{
					Name:      "capacity",
					GoName:    "Capacity",
					GoType:    "int32",
					TypeExpr:  ast.Ty("i32"),
					Supported: true,
				},
				{
					Name:      "storage_handle",
					GoName:    "Storage_handle",
					GoType:    "int64",
					TypeExpr:  ast.Ty("i64"),
					Supported: true,
				},
			}
			info.Supported = true
		}
	}

	seenPackages := make(map[string]struct{})
	for _, module := range uniqueModules {
		g.collectStaticImportsForPackage(module.Package, module.AST.Imports)
		pkgName := module.Package
		if _, ok := seenPackages[pkgName]; !ok {
			seenPackages[pkgName] = struct{}{}
			g.packages = append(g.packages, pkgName)
		}
		mapper := NewTypeMapper(g, pkgName)

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
			case *ast.ExternFunctionBody:
				if def == nil || def.Signature == nil || def.Signature.ID == nil {
					continue
				}
				g.addExternCallable(pkgName, def.Signature.ID.Name)
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
				if def == nil {
					continue
				}
				if def.Operator == ast.AssignmentDeclare || def.Operator == ast.AssignmentAssign {
					g.collectModuleBindingName(def, pkgName)
				}
				if def.Operator == ast.AssignmentDeclare {
					g.collectModuleBinding(def, pkgName)
				}
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
	ctx := newCompileContext(g, info, g.functionsForPackage(info.Package), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
	if implInfo, ok := g.implMethodByInfo[info]; ok && implInfo != nil && implInfo.IsDefault {
		ctx.implSiblings = g.implSiblingsForFunction(info)
	}
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
		if (returnType == "runtime.Value" || returnType == "any") && g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
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
				convLines, converted, ok := g.runtimeValueLines(ctx, expr, exprType)
				if !ok {
					ctx.setReason("return type mismatch")
					return nil, "", false
				}
				exprLines = append(exprLines, convLines...)
				expr = converted
			}
			ifaceLines, coerced, ok := g.interfaceReturnExprLines(ctx, expr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			exprLines = append(exprLines, ifaceLines...)
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
	if returnType == "runtime.Value" && valueType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("return type mismatch")
			return nil, "", false
		}
		stmtLines = append(stmtLines, convLines...)
		valueExpr = converted
		valueType = "runtime.Value"
	} else if returnType != "" && returnType != "runtime.Value" && returnType != "any" && returnType != valueType && g.canCoerceStaticExpr(returnType, valueType) {
		coercedLines, coercedExpr, coercedType, ok := g.coerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
		if !ok {
			ctx.setReason("assignment return type mismatch")
			return nil, "", false
		}
		stmtLines = coercedLines
		valueExpr = coercedExpr
		valueType = coercedType
	} else if !g.typeMatches(returnType, valueType) {
		if returnType != "" && returnType != "runtime.Value" && returnType != "any" && g.canCoerceStaticExpr(returnType, valueType) {
			coercedLines, coercedExpr, coercedType, ok := g.coerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
			if !ok {
				ctx.setReason("assignment return type mismatch")
				return nil, "", false
			}
			stmtLines = coercedLines
			valueExpr = coercedExpr
			valueType = coercedType
		} else {
			ctx.setReason("assignment return type mismatch")
			return nil, "", false
		}
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			ifaceLines, coerced, ok := g.interfaceReturnExprLines(ctx, valueExpr, ifaceType, ctx.genericNames)
			if !ok {
				ctx.setReason("return type mismatch")
				return nil, "", false
			}
			stmtLines = append(stmtLines, ifaceLines...)
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
					pipeLines, pipeValue, _, ok := g.compilePipeExpression(ctx, pipeExpr, "")
					if !ok {
						return nil, false
					}
					lines := append([]string{}, assignLines...)
					lines = append(lines, pipeLines...)
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
					return []string{"return runtime.VoidValue{}, nil"}, true
				}
				expected := typeExpressionToString(ctx.returnTypeExpr)
				if expected == "" || expected == "<?>" {
					expected = typeNameFromGoType(ctx.returnType)
				}
				nodeName := g.diagNodeName(s, "*ast.ReturnStatement", "return")
				ctrlExpr := fmt.Sprintf("__able_raise_return_type_mismatch(%s, %q, %q)", nodeName, expected, "void")
				lines, ok := g.controlTransferLines(ctx, ctrlExpr)
				if !ok {
					return nil, false
				}
				return lines, true
			}
			return []string{"return struct{}{}, nil"}, true
		}
		if g.isVoidType(ctx.returnType) {
			var lines []string
			stmtLines, valueExpr, _, ok := g.compileTailExpression(ctx, "", s.Argument)
			if !ok {
				return nil, false
			}
			lines = append(lines, stmtLines...)
			if valueExpr != "" {
				lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
			}
			lines = append(lines, "return struct{}{}, nil")
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
		lines = append(lines, fmt.Sprintf("return %s, nil", valueExpr))
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
