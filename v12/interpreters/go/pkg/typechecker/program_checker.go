package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

var reexportedSymbols = map[string]string{
	"able.collections.array.Array":        "able.kernel.Array",
	"able.collections.range.Range":        "able.kernel.Range",
	"able.collections.range.RangeFactory": "able.kernel.RangeFactory",
	"able.core.numeric.Ratio":             "able.kernel.Ratio",
	"able.concurrency.Channel":            "able.kernel.Channel",
	"able.concurrency.Mutex":              "able.kernel.Mutex",
	"able.concurrency.Awaitable":          "able.kernel.Awaitable",
	"able.concurrency.AwaitWaker":         "able.kernel.AwaitWaker",
	"able.concurrency.AwaitRegistration":  "able.kernel.AwaitRegistration",
}

// ProgramChecker coordinates typechecking across dependency-ordered modules.
type ProgramChecker struct {
	exports map[string]*packageExports
}

// NewProgramChecker constructs a session that can typecheck entire programs.
func NewProgramChecker() *ProgramChecker {
	return &ProgramChecker{
		exports: make(map[string]*packageExports),
	}
}

// Check walks every module in the supplied program, returning diagnostics and package summaries.
func (pc *ProgramChecker) Check(program *driver.Program) (CheckResult, error) {
	if program == nil {
		return CheckResult{}, fmt.Errorf("typechecker: program is nil")
	}
	var diagnostics []ModuleDiagnostic
	seenAliases := make(map[string]aliasDeclInfo)
	for _, mod := range program.Modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		env, impls, methods, importDiags := pc.buildPrelude(mod.AST.Imports, mod.Package)
		checker := New()
		checker.SetPrelude(env, impls, methods)
		checker.SetNodeOrigins(mod.NodeOrigins)

		moduleDiags, err := checker.CheckModule(mod.AST)
		if err != nil {
			return CheckResult{Diagnostics: diagnostics}, err
		}

		for _, diag := range importDiags {
			diagnostics = append(diagnostics, ModuleDiagnostic{
				Package:    mod.Package,
				Files:      mod.Files,
				Diagnostic: diag,
				Source:     pc.hintForNode(mod, diag.Node),
			})
		}
		for _, diag := range moduleDiags {
			diagnostics = append(diagnostics, ModuleDiagnostic{
				Package:    mod.Package,
				Files:      mod.Files,
				Diagnostic: diag,
				Source:     pc.hintForNode(mod, diag.Node),
			})
		}
		for _, diag := range pc.collectAliasDuplicateDiagnostics(mod, seenAliases) {
			diagnostics = append(diagnostics, diag)
		}

		pc.captureExports(mod, checker)
	}
	return CheckResult{
		Diagnostics: diagnostics,
		Packages:    pc.clonePackageSummaries(),
	}, nil
}

func (pc *ProgramChecker) collectAliasDuplicateDiagnostics(mod *driver.Module, seen map[string]aliasDeclInfo) []ModuleDiagnostic {
	if mod == nil || mod.AST == nil || len(mod.AST.Body) == 0 || seen == nil {
		return nil
	}
	var diags []ModuleDiagnostic
	for _, stmt := range mod.AST.Body {
		def, ok := stmt.(*ast.TypeAliasDefinition)
		if !ok || def == nil || def.ID == nil || def.ID.Name == "" {
			continue
		}
		name := def.ID.Name
		key := name
		if mod.Package != "" {
			key = mod.Package + "::" + name
		}
		var origin string
		if mod.NodeOrigins != nil {
			if path, ok := mod.NodeOrigins[def]; ok {
				origin = path
			}
		}
		if prev, ok := seen[key]; ok {
			if prev.path != "" && origin != "" && prev.path == origin {
				continue
			}
			location := formatNodeLocation(prev.node, prev.origins)
			msg := fmt.Sprintf("typechecker: duplicate declaration '%s' (previous declaration at %s)", name, location)
			diags = append(diags, ModuleDiagnostic{
				Package:    mod.Package,
				Files:      mod.Files,
				Diagnostic: Diagnostic{Message: msg, Node: def},
				Source:     pc.hintForNode(mod, def),
			})
			continue
		}
		seen[key] = aliasDeclInfo{node: def, origins: mod.NodeOrigins, path: origin}
	}
	return diags
}

// PackageExports returns a shallow copy of the exported symbol table for the specified package.
func (pc *ProgramChecker) PackageExports(pkg string) map[string]Type {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.symbols) == 0 {
		return nil
	}
	out := make(map[string]Type, len(rec.symbols))
	for name, typ := range rec.symbols {
		out[name] = typ
	}
	return out
}

// PackageStructs returns named struct type metadata exported by the given package.
func (pc *ProgramChecker) PackageStructs(pkg string) map[string]StructType {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.structs) == 0 {
		return nil
	}
	out := make(map[string]StructType, len(rec.structs))
	for name, typ := range rec.structs {
		out[name] = typ
	}
	return out
}

// PackageInterfaces returns interface metadata exported by the given package.
func (pc *ProgramChecker) PackageInterfaces(pkg string) map[string]InterfaceType {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.interfaces) == 0 {
		return nil
	}
	out := make(map[string]InterfaceType, len(rec.interfaces))
	for name, typ := range rec.interfaces {
		out[name] = typ
	}
	return out
}

// PackageFunctions returns exported top-level function signatures for the package.
func (pc *ProgramChecker) PackageFunctions(pkg string) map[string]FunctionType {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.functions) == 0 {
		return nil
	}
	out := make(map[string]FunctionType, len(rec.functions))
	for name, typ := range rec.functions {
		out[name] = typ
	}
	return out
}

// Exports returns a snapshot of all package export tables collected during the last Check run.
func (pc *ProgramChecker) Exports() map[string]map[string]Type {
	if pc == nil {
		return nil
	}
	if len(pc.exports) == 0 {
		return nil
	}
	out := make(map[string]map[string]Type, len(pc.exports))
	for pkg, rec := range pc.exports {
		if rec == nil || len(rec.symbols) == 0 {
			continue
		}
		dup := make(map[string]Type, len(rec.symbols))
		for name, typ := range rec.symbols {
			dup[name] = typ
		}
		out[pkg] = dup
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// PackageImplementations returns exported implementation specs for the given package.
func (pc *ProgramChecker) PackageImplementations(pkg string) []ImplementationSpec {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.impls) == 0 {
		return nil
	}
	out := make([]ImplementationSpec, len(rec.impls))
	copy(out, rec.impls)
	return out
}

// PackageMethodSets returns exported method-set specs for the given package.
func (pc *ProgramChecker) PackageMethodSets(pkg string) []MethodSetSpec {
	if pc == nil {
		return nil
	}
	rec, ok := pc.exports[pkg]
	if !ok || rec == nil || len(rec.methodSets) == 0 {
		return nil
	}
	out := make([]MethodSetSpec, len(rec.methodSets))
	copy(out, rec.methodSets)
	return out
}

func (pc *ProgramChecker) clonePackageSummaries() map[string]PackageSummary {
	if len(pc.exports) == 0 {
		return map[string]PackageSummary{}
	}
	result := make(map[string]PackageSummary, len(pc.exports))
	for name, rec := range pc.exports {
		if rec == nil {
			continue
		}
		symbols := make(map[string]ExportedSymbolSummary, len(rec.symbols))
		for symName, typ := range rec.symbols {
			symbols[symName] = ExportedSymbolSummary{
				Type:       formatType(typ),
				Visibility: "public",
			}
		}

		privateSymbols := make(map[string]ExportedSymbolSummary, len(rec.private))
		for symName, typ := range rec.private {
			privateSymbols[symName] = ExportedSymbolSummary{
				Type:       formatType(typ),
				Visibility: "private",
			}
		}

		structs := make(map[string]ExportedStructSummary, len(rec.structs))
		for structName, structType := range rec.structs {
			structs[structName] = summarizeStructType(structType)
		}

		interfaces := make(map[string]ExportedInterfaceSummary, len(rec.interfaces))
		for interfaceName, ifaceType := range rec.interfaces {
			interfaces[interfaceName] = summarizeInterfaceType(ifaceType)
		}

		functions := make(map[string]ExportedFunctionSummary, len(rec.functions))
		for fnName, fnType := range rec.functions {
			functions[fnName] = summarizeFunctionType(fnType)
		}

		impls := make([]ExportedImplementationSummary, 0, len(rec.impls))
		for _, impl := range rec.impls {
			impls = append(impls, summarizeImplementation(impl))
		}

		methodSets := make([]ExportedMethodSetSummary, 0, len(rec.methodSets))
		for _, set := range rec.methodSets {
			methodSets = append(methodSets, summarizeMethodSet(set))
		}

		summary := PackageSummary{
			Name:            name,
			Visibility:      rec.visibility,
			Symbols:         symbols,
			PrivateSymbols:  privateSymbols,
			Structs:         structs,
			Interfaces:      interfaces,
			Functions:       functions,
			Implementations: impls,
			MethodSets:      methodSets,
		}
		if summary.Visibility == "" {
			summary.Visibility = "public"
		}
		if summary.Symbols == nil {
			summary.Symbols = map[string]ExportedSymbolSummary{}
		}
		if summary.PrivateSymbols == nil {
			summary.PrivateSymbols = map[string]ExportedSymbolSummary{}
		}
		if summary.Structs == nil {
			summary.Structs = map[string]ExportedStructSummary{}
		}
		if summary.Interfaces == nil {
			summary.Interfaces = map[string]ExportedInterfaceSummary{}
		}
		if summary.Functions == nil {
			summary.Functions = map[string]ExportedFunctionSummary{}
		}
		if summary.Implementations == nil {
			summary.Implementations = []ExportedImplementationSummary{}
		}
		if summary.MethodSets == nil {
			summary.MethodSets = []ExportedMethodSetSummary{}
		}
		result[name] = summary
	}
	return result
}

func (pc *ProgramChecker) resolveReexportTarget(target string) (Type, bool) {
	if target == "" {
		return nil, false
	}
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return nil, false
	}
	pkg := strings.Join(parts[:len(parts)-1], ".")
	sym := parts[len(parts)-1]
	export, ok := pc.exports[pkg]
	if !ok || export == nil || len(export.symbols) == 0 {
		return nil, false
	}
	typ, ok := export.symbols[sym]
	if !ok || typ == nil {
		return nil, false
	}
	return typ, true
}

func (pc *ProgramChecker) resolveReexportedSymbol(pkgName, symbol string) (Type, bool) {
	target, ok := reexportedSymbols[pkgName+"."+symbol]
	if !ok {
		return nil, false
	}
	return pc.resolveReexportTarget(target)
}

func (pc *ProgramChecker) addReexportsToEnv(pkgName string, env *Environment) {
	if env == nil || pkgName == "" {
		return
	}
	prefix := pkgName + "."
	for fq, target := range reexportedSymbols {
		if !strings.HasPrefix(fq, prefix) {
			continue
		}
		name := strings.TrimPrefix(fq, prefix)
		if env.HasInCurrentScope(name) {
			continue
		}
		if typ, ok := pc.resolveReexportTarget(target); ok {
			env.Define(name, typ)
		}
	}
}

func (pc *ProgramChecker) addReexportsToSymbols(pkgName string, symbols map[string]Type) {
	if pkgName == "" || symbols == nil {
		return
	}
	prefix := pkgName + "."
	for fq, target := range reexportedSymbols {
		if !strings.HasPrefix(fq, prefix) {
			continue
		}
		name := strings.TrimPrefix(fq, prefix)
		if _, exists := symbols[name]; exists {
			continue
		}
		if typ, ok := pc.resolveReexportTarget(target); ok {
			symbols[name] = typ
		}
	}
}

func (pc *ProgramChecker) buildPrelude(imports []*ast.ImportStatement, currentPackage string) (*Environment, []ImplementationSpec, []MethodSetSpec, []Diagnostic) {
	if len(imports) == 0 {
		return nil, nil, nil, nil
	}
	env := NewEnvironment(nil)
	var (
		impls    []ImplementationSpec
		methods  []MethodSetSpec
		diags    []Diagnostic
		hasScope bool
	)
	seen := make(map[string]struct{})
	implSeen := make(map[string]struct{})
	var collectImpls func(pkg string)
	collectImpls = func(pkg string) {
		if pkg == "" {
			return
		}
		if _, ok := implSeen[pkg]; ok {
			return
		}
		implSeen[pkg] = struct{}{}
		export, ok := pc.exports[pkg]
		if !ok || export == nil {
			return
		}
		if len(export.impls) > 0 {
			impls = append(impls, export.impls...)
		}
		if len(export.methodSets) > 0 {
			methods = append(methods, export.methodSets...)
		}
		for _, dep := range export.imports {
			collectImpls(dep)
		}
	}

	for _, imp := range imports {
		if imp == nil {
			continue
		}
		pkgName := joinImportPath(imp.PackagePath)
		export, ok := pc.exports[pkgName]
		if !ok {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: import references unknown package '%s'", pkgName),
				Node:    imp,
			})
			continue
		}
		if export == nil {
			continue
		}
		if export.visibility == "private" && pkgName != currentPackage {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: package '%s' is private", pkgName),
				Node:    imp,
			})
			continue
		}
		collectImpls(pkgName)
		if _, already := seen[pkgName]; !already {
			seen[pkgName] = struct{}{}
		}

		if imp.IsWildcard {
			for name, typ := range export.symbols {
				if typ == nil {
					typ = UnknownType{}
				}
				env.Define(name, typ)
				hasScope = true
			}
			pc.addReexportsToEnv(pkgName, env)
			continue
		}

		if len(imp.Selectors) > 0 {
			for _, sel := range imp.Selectors {
				if sel == nil || sel.Name == nil {
					continue
				}
				alias := sel.Name.Name
				if sel.Alias != nil && sel.Alias.Name != "" {
					alias = sel.Alias.Name
				}
				if typ, exists := export.symbols[sel.Name.Name]; exists && typ != nil {
					env.Define(alias, typ)
					hasScope = true
					continue
				}
				if typ, ok := pc.resolveReexportedSymbol(pkgName, sel.Name.Name); ok && typ != nil {
					env.Define(alias, typ)
					hasScope = true
					continue
				}
				if export.private != nil {
					if _, exists := export.private[sel.Name.Name]; exists {
						diags = append(diags, Diagnostic{
							Message: fmt.Sprintf("typechecker: package '%s' symbol '%s' is private", pkgName, sel.Name.Name),
							Node:    sel,
						})
						env.Define(alias, UnknownType{})
						hasScope = true
						continue
					}
				}
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: package '%s' has no symbol '%s'", pkgName, sel.Name.Name),
					Node:    sel,
				})
				env.Define(alias, UnknownType{})
				hasScope = true
			}
			continue
		}

		alias := pkgName
		if imp.Alias != nil && imp.Alias.Name != "" {
			alias = imp.Alias.Name
		} else if tail := lastImportSegment(imp.PackagePath); tail != "" {
			alias = tail
		}
		pkgType := PackageType{
			Package: pkgName,
			Symbols: nil,
		}
		if export != nil {
			symbols := make(map[string]Type, len(export.symbols))
			for name, typ := range export.symbols {
				symbols[name] = typ
			}
			pc.addReexportsToSymbols(pkgName, symbols)
			pkgType.Symbols = symbols
			pkgType.PrivateSymbols = export.private
		} else {
			pkgType.Symbols = map[string]Type{}
			pkgType.PrivateSymbols = map[string]Type{}
		}
		env.Define(alias, pkgType)
		hasScope = true
	}

	if !hasScope {
		env = nil
	}
	return env, impls, methods, diags
}

func (pc *ProgramChecker) captureExports(mod *driver.Module, checker *Checker) {
	if mod == nil {
		return
	}
	var methodQualifier func(Type) string
	methodQualifier = func(t Type) string {
		switch tv := t.(type) {
		case StructType:
			return tv.StructName
		case StructInstanceType:
			return tv.StructName
		case ArrayType:
			return "Array"
		case RangeType:
			return "Range"
		case IteratorType:
			return "Iterator"
		case FutureType:
			return "Future"
		case NullableType:
			return methodQualifier(tv.Inner)
		case AppliedType:
			return methodQualifier(tv.Base)
		case AliasType:
			return methodQualifier(tv.Target)
		default:
			return typeName(t)
		}
	}
	visibility := "public"
	if mod.AST != nil && mod.AST.Package != nil && mod.AST.Package.IsPrivate {
		visibility = "private"
	}
	export := &packageExports{
		name:       mod.Package,
		visibility: visibility,
		symbols:    make(map[string]Type),
		private:    make(map[string]Type),
		structs:    make(map[string]StructType),
		interfaces: make(map[string]InterfaceType),
		functions:  make(map[string]FunctionType),
	}
	if mod.AST != nil && len(mod.AST.Imports) > 0 {
		seen := make(map[string]struct{})
		for _, imp := range mod.AST.Imports {
			if imp == nil {
				continue
			}
			pkgName := joinImportPath(imp.PackagePath)
			if pkgName == "" {
				continue
			}
			if _, ok := seen[pkgName]; ok {
				continue
			}
			seen[pkgName] = struct{}{}
			export.imports = append(export.imports, pkgName)
		}
	}
	for _, sym := range checker.ExportedSymbols() {
		export.symbols[sym.Name] = sym.Type
		switch t := sym.Type.(type) {
		case StructType:
			export.structs[sym.Name] = t
		case InterfaceType:
			export.interfaces[sym.Name] = t
		case FunctionType:
			export.functions[sym.Name] = t
		}
	}
	for _, impl := range checker.ModuleImplementations() {
		if impl.Definition != nil && impl.Definition.IsPrivate {
			continue
		}
		export.impls = append(export.impls, impl)
	}
	// Note: impl methods are not exported as standalone symbols. They are
	// resolved via interface/method lookup instead of importable names.
	for _, methods := range checker.ModuleMethodSets() {
		export.methodSets = append(export.methodSets, methods)
	}

	methodPrivacy := func(def *ast.MethodsDefinition) map[string]bool {
		privacy := make(map[string]bool)
		if def == nil {
			return privacy
		}
		for _, fn := range def.Definitions {
			if fn == nil || fn.ID == nil {
				continue
			}
			privacy[fn.ID.Name] = fn.IsPrivate
		}
		return privacy
	}
	for _, spec := range export.methodSets {
		qualifier := methodQualifier(spec.Target)
		privacy := methodPrivacy(spec.Definition)
		for name, fn := range spec.Methods {
			key := name
			if spec.TypeQualified != nil && spec.TypeQualified[name] && qualifier != "" {
				key = fmt.Sprintf("%s.%s", qualifier, name)
			}
			export.functions[key] = fn
			if _, exists := export.symbols[key]; exists {
				continue
			}
			if privacy[name] {
				export.private[key] = fn
				delete(export.symbols, key)
			} else {
				export.symbols[key] = fn
				delete(export.private, key)
			}
		}
	}

	recordPrivate := func(name string) {
		if name == "" || export.private == nil || checker == nil || checker.global == nil {
			return
		}
		if _, exists := export.private[name]; exists {
			return
		}
		if typ, ok := checker.global.Lookup(name); ok && typ != nil {
			export.private[name] = typ
		}
	}
	if mod.AST != nil {
		for _, stmt := range mod.AST.Body {
			switch def := stmt.(type) {
			case *ast.StructDefinition:
				if def == nil || def.ID == nil || !def.IsPrivate {
					continue
				}
				recordPrivate(def.ID.Name)
			case *ast.InterfaceDefinition:
				if def == nil || def.ID == nil || !def.IsPrivate {
					continue
				}
				recordPrivate(def.ID.Name)
			case *ast.FunctionDefinition:
				if def == nil || def.ID == nil {
					continue
				}
				if def.IsPrivate {
					recordPrivate(def.ID.Name)
				}
			case *ast.MethodsDefinition:
				for _, fn := range def.Definitions {
					if fn == nil || fn.ID == nil || !fn.IsPrivate {
						continue
					}
					recordPrivate(fn.ID.Name)
				}
			case *ast.TypeAliasDefinition:
				if def == nil || def.ID == nil || !def.IsPrivate {
					continue
				}
				recordPrivate(def.ID.Name)
			}
		}
	}

	pc.exports[mod.Package] = export
}

func joinImportPath(parts []*ast.Identifier) string {
	if len(parts) == 0 {
		return ""
	}
	names := make([]string, 0, len(parts))
	for _, id := range parts {
		if id == nil || id.Name == "" {
			continue
		}
		names = append(names, id.Name)
	}
	return strings.Join(names, ".")
}

func (pc *ProgramChecker) hintForNode(mod *driver.Module, node ast.Node) SourceHint {
	if mod == nil {
		return SourceHint{}
	}
	hint := SourceHint{}
	if node != nil {
		span := node.Span()
		if span.Start.Line > 0 && span.Start.Column > 0 {
			hint.Line = span.Start.Line
			hint.Column = span.Start.Column
		}
		if span.End.Line > 0 && span.End.Column > 0 {
			hint.EndLine = span.End.Line
			hint.EndColumn = span.End.Column
		}
	}
	if node != nil && mod.NodeOrigins != nil {
		if path, ok := mod.NodeOrigins[node]; ok && path != "" {
			hint.Path = path
			return hint
		}
	}
	if node != nil && hint.Line == 0 && hint.Column == 0 && hint.EndLine == 0 && hint.EndColumn == 0 {
		return hint
	}
	if len(mod.Files) > 0 {
		hint.Path = mod.Files[0]
	}
	return hint
}
