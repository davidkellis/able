package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

type generator struct {
	opts                                Options
	structs                             map[string]*structInfo
	specializedStructs                  map[string]*structInfo
	typeAliases                         map[string]map[string]ast.TypeExpression
	typeAliasGenericParams              map[string]map[string][]*ast.GenericParameter
	unions                              map[string]*ast.UnionDefinition
	unionPackages                       map[string]string
	nativeUnions                        map[string]*nativeUnionInfo
	nativeUnionExprIndex                map[string]*nativeUnionInfo
	nativeUnionRendered                 map[string]struct{}
	nativeCallables                     map[string]*nativeCallableInfo
	nativeInterfaces                    map[string]*nativeInterfaceInfo
	nativeInterfaceExplicitAdapters     map[string]map[string]*nativeInterfaceAdapter
	nativeInterfaceDefaultByInfo        map[*functionInfo]*nativeInterfaceDefaultMethodInfo
	nativeInterfaceDefaultSiblingCache  map[string]map[string]implSiblingInfo
	nativeInterfaceGenericDispatches    map[string]*nativeInterfaceGenericDispatchInfo
	nativeInterfaceRenderedAdapters     map[string]struct{}
	nativeInterfaceRenderedInfos        map[string]struct{}
	nativeInterfaceRenderedDispatches   map[string]struct{}
	nativeInterfaceRenderedApplyHelpers map[string]struct{}
	iteratorCollectMonoArrays           map[string]*iteratorCollectMonoArrayInfo
	monoArraySpecs                      map[string]*monoArraySpec
	nativeInterfaceBuilding             map[string]struct{}
	nativeInterfaceRefreshing           map[string]struct{}
	nativeInterfaceSpecializing         map[string]struct{}
	nativeInterfaceImplBindingCache     map[string]nativeInterfaceImplBindingCacheEntry
	normalizedTypeExprCache             map[string]ast.TypeExpression
	normalizedTypeExprPackageCache      map[string]string
	normalizedTypeExprPackagesByExpr    map[ast.TypeExpression]string
	nativeInterfaceImplCandidateCache   []nativeInterfaceImplCandidate
	nativeInterfaceImplCandidateCounts  [2]int
	nativeInterfaceAdapterVersion       int
	bodyCompilationDepth                int
	interfaces                          map[string]*ast.InterfaceDefinition
	interfacesByPackage                 map[string]map[string]*ast.InterfaceDefinition
	interfacePackages                   map[string]string
	staticImports                       map[string][]staticImportBinding
	functions                           map[string]map[string]*functionInfo
	overloads                           map[string]map[string]*overloadInfo
	packages                            []string
	entryPackage                        string
	methods                             map[string]map[string][]*methodInfo
	methodList                          []*methodInfo
	implMethodList                      []*implMethodInfo
	implDefinitions                     []*implDefinitionInfo
	implMethodByInfo                    map[*functionInfo]*implMethodInfo
	implMethodsBySignature              map[string][]*implMethodInfo
	specializedFunctions                []*functionInfo
	specializedFunctionIndex            map[string]*functionInfo
	nominalCoercions                    map[string]*nominalCoercionInfo
	warnings                            []string
	fallbacks                           []FallbackInfo
	mangler                             *nameMangler
	needsAst                            bool
	needsIterator                       bool
	needsStrconv                        bool
	needsStringFromByteArray            bool
	awaitExprs                          []string
	awaitNames                          map[*ast.AwaitExpression]string
	diagNodes                           []diagNodeInfo
	diagNames                           map[ast.Node]string
	nodeOrigins                         map[ast.Node]string
	packageEnvVars                      map[string]string
	packageBootstrappedVars             map[string]string
	packageEnvOrder                     []string
	packageInitOrder                    []string
	packageInitStatements               map[string][]ast.Statement
	packageInitCompiled                 map[string][]string
	hasDynamicFeature                   bool
	moduleBindings                      map[string][]moduleBinding // package -> bindings
	moduleBindingNames                  map[string]map[string]struct{}
	moduleMutableBindingNames           map[string]map[string]struct{}
	evaluatedConstants                  map[string]*evaluatedConst // "pkg::name" -> value
	staticCallableNames                 map[string]map[string]struct{}
	externCallables                     map[string]map[string]struct{}
	externBodies                        map[string]map[string][]*ast.ExternFunctionBody
	goPreludes                          map[string][]string
	goPreludeImports                    []string
	goPreludeDecls                      []string
	inferredTypes                       map[string]typechecker.InferenceMap
}

func newGenerator(opts Options) *generator {
	// Compiler-native array carriers are now the default static lowering path.
	// Keep the option field for compatibility while older call sites/tests are
	// updated, but do not let the generic *Array carrier remain the default.
	opts.ExperimentalMonoArrays = true
	return &generator{
		opts:                                opts,
		structs:                             make(map[string]*structInfo),
		specializedStructs:                  make(map[string]*structInfo),
		typeAliases:                         make(map[string]map[string]ast.TypeExpression),
		typeAliasGenericParams:              make(map[string]map[string][]*ast.GenericParameter),
		unions:                              make(map[string]*ast.UnionDefinition),
		unionPackages:                       make(map[string]string),
		nativeUnions:                        make(map[string]*nativeUnionInfo),
		nativeUnionExprIndex:                make(map[string]*nativeUnionInfo),
		nativeUnionRendered:                 make(map[string]struct{}),
		nativeCallables:                     make(map[string]*nativeCallableInfo),
		nativeInterfaces:                    make(map[string]*nativeInterfaceInfo),
		nativeInterfaceExplicitAdapters:     make(map[string]map[string]*nativeInterfaceAdapter),
		nativeInterfaceDefaultByInfo:        make(map[*functionInfo]*nativeInterfaceDefaultMethodInfo),
		nativeInterfaceDefaultSiblingCache:  make(map[string]map[string]implSiblingInfo),
		nativeInterfaceGenericDispatches:    make(map[string]*nativeInterfaceGenericDispatchInfo),
		nativeInterfaceRenderedAdapters:     make(map[string]struct{}),
		nativeInterfaceRenderedInfos:        make(map[string]struct{}),
		nativeInterfaceRenderedDispatches:   make(map[string]struct{}),
		nativeInterfaceRenderedApplyHelpers: make(map[string]struct{}),
		iteratorCollectMonoArrays:           make(map[string]*iteratorCollectMonoArrayInfo),
		monoArraySpecs:                      make(map[string]*monoArraySpec),
		nativeInterfaceBuilding:             make(map[string]struct{}),
		nativeInterfaceRefreshing:           make(map[string]struct{}),
		nativeInterfaceSpecializing:         make(map[string]struct{}),
		nativeInterfaceImplBindingCache:     make(map[string]nativeInterfaceImplBindingCacheEntry),
		normalizedTypeExprCache:             make(map[string]ast.TypeExpression),
		normalizedTypeExprPackageCache:      make(map[string]string),
		normalizedTypeExprPackagesByExpr:    make(map[ast.TypeExpression]string),
		nativeInterfaceImplCandidateCounts:  [2]int{-1, -1},
		nativeInterfaceAdapterVersion:       1,
		interfaces:                          make(map[string]*ast.InterfaceDefinition),
		interfacesByPackage:                 make(map[string]map[string]*ast.InterfaceDefinition),
		interfacePackages:                   make(map[string]string),
		staticImports:                       make(map[string][]staticImportBinding),
		functions:                           make(map[string]map[string]*functionInfo),
		overloads:                           make(map[string]map[string]*overloadInfo),
		methods:                             make(map[string]map[string][]*methodInfo),
		mangler:                             newNameMangler(),
		awaitNames:                          make(map[*ast.AwaitExpression]string),
		implMethodByInfo:                    make(map[*functionInfo]*implMethodInfo),
		implMethodsBySignature:              make(map[string][]*implMethodInfo),
		specializedFunctionIndex:            make(map[string]*functionInfo),
		nominalCoercions:                    make(map[string]*nominalCoercionInfo),
		packageInitStatements:               make(map[string][]ast.Statement),
		packageInitCompiled:                 make(map[string][]string),
		moduleBindingNames:                  make(map[string]map[string]struct{}),
		moduleMutableBindingNames:           make(map[string]map[string]struct{}),
		externCallables:                     make(map[string]map[string]struct{}),
		externBodies:                        make(map[string]map[string][]*ast.ExternFunctionBody),
		goPreludes:                          make(map[string][]string),
	}
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
	g.nativeUnionExprIndex = make(map[string]*nativeUnionInfo)
	g.nativeUnionRendered = make(map[string]struct{})
	g.nativeCallables = make(map[string]*nativeCallableInfo)
	g.nativeInterfaces = make(map[string]*nativeInterfaceInfo)
	g.nativeInterfaceExplicitAdapters = make(map[string]map[string]*nativeInterfaceAdapter)
	g.nativeInterfaceDefaultByInfo = make(map[*functionInfo]*nativeInterfaceDefaultMethodInfo)
	g.nativeInterfaceDefaultSiblingCache = make(map[string]map[string]implSiblingInfo)
	g.nativeInterfaceGenericDispatches = make(map[string]*nativeInterfaceGenericDispatchInfo)
	g.nativeInterfaceRenderedAdapters = make(map[string]struct{})
	g.nativeInterfaceRenderedInfos = make(map[string]struct{})
	g.nativeInterfaceRenderedDispatches = make(map[string]struct{})
	g.nativeInterfaceRenderedApplyHelpers = make(map[string]struct{})
	g.iteratorCollectMonoArrays = make(map[string]*iteratorCollectMonoArrayInfo)
	g.monoArraySpecs = make(map[string]*monoArraySpec)
	g.nativeInterfaceBuilding = make(map[string]struct{})
	g.nativeInterfaceRefreshing = make(map[string]struct{})
	g.nativeInterfaceSpecializing = make(map[string]struct{})
	g.nativeInterfaceImplBindingCache = make(map[string]nativeInterfaceImplBindingCacheEntry)
	g.normalizedTypeExprCache = make(map[string]ast.TypeExpression)
	g.normalizedTypeExprPackageCache = make(map[string]string)
	g.normalizedTypeExprPackagesByExpr = make(map[ast.TypeExpression]string)
	g.nativeInterfaceImplCandidateCache = nil
	g.nativeInterfaceImplCandidateCounts = [2]int{-1, -1}
	g.nativeInterfaceAdapterVersion = 1
	g.interfaces = make(map[string]*ast.InterfaceDefinition)
	g.interfacesByPackage = make(map[string]map[string]*ast.InterfaceDefinition)
	g.interfacePackages = make(map[string]string)
	g.staticCallableNames = nil
	g.externCallables = make(map[string]map[string]struct{})
	g.externBodies = make(map[string]map[string][]*ast.ExternFunctionBody)
	g.goPreludes = make(map[string][]string)
	g.goPreludeImports = nil
	g.goPreludeDecls = nil
	g.implMethodsBySignature = make(map[string][]*implMethodInfo)
	g.specializedFunctions = nil
	g.specializedFunctionIndex = make(map[string]*functionInfo)
	g.nominalCoercions = make(map[string]*nominalCoercionInfo)
	g.packageInitOrder = nil
	g.packageInitStatements = make(map[string][]ast.Statement)
	g.packageInitCompiled = make(map[string][]string)
	g.invalidatePackageEnvVars()
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
			if g.interfacesByPackage != nil {
				if g.interfacesByPackage[module.Package] == nil {
					g.interfacesByPackage[module.Package] = make(map[string]*ast.InterfaceDefinition)
				}
				if _, exists := g.interfacesByPackage[module.Package][name]; !exists {
					g.interfacesByPackage[module.Package][name] = def
				}
			}
			if _, exists := g.interfaces[name]; !exists {
				g.interfaces[name] = def
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
	g.ensureBuiltinDivModStruct()

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
			fieldTypeExpr := normalizeTypeExprForPackage(g, info.Package, field.FieldType)
			goType, ok := mapper.Map(field.FieldType)
			goType, ok = g.recoverRepresentableCarrierType(info.Package, fieldTypeExpr, goType)
			if !ok {
				supported = false
			}
			fields = append(fields, fieldInfo{
				Name:      fieldName,
				GoName:    goFieldName,
				GoType:    goType,
				TypeExpr:  fieldTypeExpr,
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
			g.invalidatePackageEnvVars()
		}
		mapper := NewTypeMapper(g, pkgName)

		functions := make(map[string][]*ast.FunctionDefinition)
		for _, stmt := range module.AST.Body {
			switch def := stmt.(type) {
			case *ast.PreludeStatement:
				g.collectGoPrelude(pkgName, def)
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
				g.collectGoExternBody(pkgName, def)
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
		g.collectMutableModuleBindings(module.AST.Body, pkgName)

		if g.functions[pkgName] == nil {
			g.functions[pkgName] = make(map[string]*functionInfo)
		}
		if g.overloads[pkgName] == nil {
			g.overloads[pkgName] = make(map[string]*overloadInfo)
		}

		externs := g.externBodiesForPackage(pkgName)
		callableNames := make(map[string]struct{}, len(functions)+len(externs))
		for name := range functions {
			callableNames[name] = struct{}{}
		}
		for name := range externs {
			callableNames[name] = struct{}{}
		}
		callableList := make([]string, 0, len(callableNames))
		for name := range callableNames {
			callableList = append(callableList, name)
		}
		sort.Strings(callableList)
		for _, name := range callableList {
			defs := functions[name]
			externDefs := externs[name]
			qualified := qualifiedName(pkgName, name)
			if len(externDefs) > 0 {
				if len(externDefs) != 1 {
					entries := make([]*functionInfo, 0, len(externDefs))
					minArity := -1
					for idx, def := range externDefs {
						if def == nil || def.Signature == nil || def.Signature.ID == nil {
							continue
						}
						info := &functionInfo{
							Name:          name,
							Package:       pkgName,
							QualifiedName: qualified,
							GoName:        g.mangler.unique(fmt.Sprintf("fn_%s_extern_overload_%d", sanitizeIdent(name), idx)),
							Definition:    def.Signature,
							ExternBody:    def,
							HasOriginal:   true,
						}
						g.fillFunctionInfo(info, mapper)
						entries = append(entries, info)
						if arity := len(info.Params); arity >= 0 {
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
				def := externDefs[0]
				if def == nil || def.Signature == nil || def.Signature.ID == nil {
					continue
				}
				info := &functionInfo{
					Name:          name,
					Package:       pkgName,
					QualifiedName: qualified,
					GoName:        g.mangler.unique("fn_" + sanitizeIdent(name)),
					Definition:    def.Signature,
					ExternBody:    def,
					HasOriginal:   true,
				}
				g.fillFunctionInfo(info, mapper)
				g.functions[pkgName][name] = info
				continue
			}
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
	g.collectPackageInitStatements(program)
	g.collectDefaultImplMethods()
	sort.Strings(g.packages)
	g.invalidatePackageEnvVars()
	g.resolveCompileabilityFixedPoint()
	g.detectAstNeeds()
	return nil
}

func (g *generator) fillFunctionInfo(info *functionInfo, mapper *TypeMapper) {
	if info == nil || info.Definition == nil {
		return
	}
	g.invalidateFunctionDerivedInfo(info)
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
		goType, ok = g.recoverRepresentableCarrierType(info.Package, param.ParamType, goType)
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
	retExpr := g.functionDeclaredOrInferredReturnTypeExpr(info)
	retType, ok := mapper.Map(retExpr)
	retType, ok = g.recoverRepresentableCarrierType(info.Package, retExpr, retType)
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

func (g *generator) compileBody(ctx *compileContext, info *functionInfo) ([]string, string, bool) {
	if g != nil {
		g.bodyCompilationDepth++
		defer func() {
			g.bodyCompilationDepth--
		}()
	}
	if info == nil || info.Definition == nil || info.Definition.Body == nil {
		ctx.setReason("missing function body")
		return nil, "", false
	}
	statements := info.Definition.Body.Body
	if len(statements) == 0 {
		if g.isVoidType(info.ReturnType) {
			return nil, "struct{}{}", true
		}
		if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, info.ReturnType); ok {
			return nil, successExpr, true
		}
		ctx.setReason("empty body requires void return")
		return nil, "", false
	}
	ctx.blockStatements = statements
	lines := make([]string, 0, len(statements))
	for idx, stmt := range statements {
		ctx.statementIndex = idx
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
			if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, info.ReturnType); ok {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", false
				}
				lines = append(lines, stmtLines...)
				return lines, successExpr, true
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
	if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, info.ReturnType); ok {
		return lines, successExpr, true
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
		if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, returnType); ok {
			return lines, successExpr, true
		}
		ctx.setReason("missing return expression")
		return nil, "", false
	}
	if g.isVoidType(returnType) {
		stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", ret.Argument)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, "", false
			}
		}
		return lines, "struct{}{}", true
	}
	previousExpectedTypeExpr := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = g.concretizedExpectedTypeExpr(ctx, returnType, ctx.returnTypeExpr)
	exprLines, expr, exprType, ok := g.compileTailExpression(ctx, returnType, ret.Argument)
	ctx.expectedTypeExpr = previousExpectedTypeExpr
	if !ok {
		return nil, "", false
	}
	if returnType == "runtime.Value" {
		if ifaceType, ok := g.interfaceTypeExpr(ctx.returnTypeExpr); ok {
			if exprType != "runtime.Value" {
				convLines, converted, ok := g.lowerRuntimeValue(ctx, expr, exprType)
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
		stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, stmtLines...)
		if valueExpr != "" {
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, "", false
			}
		}
		return lines, "struct{}{}", true
	}
	previousExpectedTypeExpr := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = g.concretizedExpectedTypeExpr(ctx, returnType, ctx.returnTypeExpr)
	stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, expr)
	ctx.expectedTypeExpr = previousExpectedTypeExpr
	if !ok {
		return nil, "", false
	}
	if returnType == "runtime.Value" && valueType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("return type mismatch")
			return nil, "", false
		}
		stmtLines = append(stmtLines, convLines...)
		valueExpr = converted
		valueType = "runtime.Value"
	} else if returnType != "" && returnType != "runtime.Value" && returnType != "any" && returnType != valueType && g.canCoerceStaticExpr(returnType, valueType) {
		coercedLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
		if !ok {
			ctx.setReason("assignment return type mismatch")
			return nil, "", false
		}
		stmtLines = coercedLines
		valueExpr = coercedExpr
		valueType = coercedType
	} else if !g.typeMatches(returnType, valueType) {
		if returnType != "" && returnType != "runtime.Value" && returnType != "any" && g.canCoerceStaticExpr(returnType, valueType) {
			coercedLines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, stmtLines, valueExpr, valueType, returnType)
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
		lines, valueExpr, valueType, ok := g.compileAssignment(ctx, s)
		if !ok {
			return nil, false
		}
		if valueExpr != "" {
			if name, _, ok := g.assignmentTargetName(s.Left); ok && name != "" {
				lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
				return lines, true
			}
			lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
			if !ok {
				return nil, false
			}
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
					pipeLines, pipeValue, pipeType, ok := g.compilePipeExpression(ctx, pipeExpr, "")
					if !ok {
						return nil, false
					}
					lines := append([]string{}, assignLines...)
					lines = append(lines, pipeLines...)
					if pipeValue != "" {
						lines, ok = g.discardStatementResult(ctx, lines, pipeValue, pipeType)
						if !ok {
							return nil, false
						}
					}
					return lines, true
				}
			}
		}
		if expr, ok := stmt.(ast.Expression); ok {
			valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr)
			if !ok {
				return nil, false
			}
			if valueExpr == "" {
				return valueLines, true
			}
			lines, ok := g.discardStatementResult(ctx, valueLines, valueExpr, valueType)
			if !ok {
				return nil, false
			}
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
				if successExpr, ok := g.nativeResultVoidSuccessExpr(ctx, ctx.returnType); ok {
					return []string{fmt.Sprintf("return %s, nil", successExpr)}, true
				}
				expected := typeExpressionToString(ctx.returnTypeExpr)
				if expected == "" || expected == "<?>" {
					expected = typeNameFromGoType(ctx.returnType)
				}
				nodeName := g.diagNodeName(s, "*ast.ReturnStatement", "return")
				ctrlExpr := fmt.Sprintf("__able_raise_return_type_mismatch(%s, %q, %q)", nodeName, expected, "void")
				lines, ok := g.lowerControlTransfer(ctx, ctrlExpr)
				if !ok {
					return nil, false
				}
				return lines, true
			}
			return []string{"return struct{}{}, nil"}, true
		}
		if g.isVoidType(ctx.returnType) {
			var lines []string
			stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", s.Argument)
			if !ok {
				return nil, false
			}
			lines = append(lines, stmtLines...)
			if valueExpr != "" {
				lines, ok = g.discardStatementResult(ctx, lines, valueExpr, valueType)
				if !ok {
					return nil, false
				}
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
			valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr)
			if !ok {
				return nil, false
			}
			if valueExpr == "" {
				return valueLines, true
			}
			lines, ok := g.discardStatementResult(ctx, valueLines, valueExpr, valueType)
			if !ok {
				return nil, false
			}
			return lines, true
		}
		ctx.setReason("unsupported statement")
		return nil, false
	}
}
