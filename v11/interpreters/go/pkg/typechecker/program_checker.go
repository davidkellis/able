package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

// ModuleDiagnostic ties a diagnostic to the package/files that produced it.
type ModuleDiagnostic struct {
	Package    string
	Files      []string
	Diagnostic Diagnostic
	Source     SourceHint
}

// SourceHint provides a best-effort reference to the originating file.
type SourceHint struct {
	Path   string
	Line   int
	Column int
}

// DescribeModuleDiagnostic formats a module diagnostic for human-readable output.
func DescribeModuleDiagnostic(diag ModuleDiagnostic) string {
	message := diag.Diagnostic.Message
	if diag.Package != "" {
		message = fmt.Sprintf("%s: %s", diag.Package, message)
	}
	if diag.Source.Path != "" {
		switch {
		case diag.Source.Line > 0 && diag.Source.Column > 0:
			message = fmt.Sprintf("%s (%s:%d:%d)", message, diag.Source.Path, diag.Source.Line, diag.Source.Column)
		case diag.Source.Line > 0:
			message = fmt.Sprintf("%s (%s:%d)", message, diag.Source.Path, diag.Source.Line)
		default:
			message = fmt.Sprintf("%s (%s)", message, diag.Source.Path)
		}
	} else if len(diag.Files) > 0 {
		message = fmt.Sprintf("%s (e.g., %s)", message, diag.Files[0])
	}
	return message
}

// ExportedSymbolSummary summarises a binding exposed by a package.
type ExportedSymbolSummary struct {
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
}

// ExportedGenericParamSummary summarises a generic parameter and its constraints.
type ExportedGenericParamSummary struct {
	Name        string   `json:"name"`
	Constraints []string `json:"constraints,omitempty"`
}

// ExportedWhereConstraintSummary records a where-clause requirement.
type ExportedWhereConstraintSummary struct {
	TypeParam   string   `json:"typeParam"`
	Constraints []string `json:"constraints,omitempty"`
}

// ExportedObligationSummary captures solver obligations that arose while collecting exports.
type ExportedObligationSummary struct {
	Owner      string `json:"owner,omitempty"`
	TypeParam  string `json:"typeParam"`
	Constraint string `json:"constraint"`
	Subject    string `json:"subject"`
	Context    string `json:"context,omitempty"`
}

// ExportedFunctionSummary describes the callable surface of a function or method.
type ExportedFunctionSummary struct {
	Parameters  []string                         `json:"parameters,omitempty"`
	ReturnType  string                           `json:"returnType"`
	TypeParams  []ExportedGenericParamSummary    `json:"typeParams,omitempty"`
	Where       []ExportedWhereConstraintSummary `json:"where,omitempty"`
	Obligations []ExportedObligationSummary      `json:"obligations,omitempty"`
}

// ExportedStructSummary summarises a public struct definition.
type ExportedStructSummary struct {
	TypeParams []ExportedGenericParamSummary    `json:"typeParams,omitempty"`
	Fields     map[string]string                `json:"fields,omitempty"`
	Positional []string                         `json:"positional,omitempty"`
	Where      []ExportedWhereConstraintSummary `json:"where,omitempty"`
}

// ExportedInterfaceSummary summarises a public interface definition.
type ExportedInterfaceSummary struct {
	TypeParams []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Methods    map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where      []ExportedWhereConstraintSummary   `json:"where,omitempty"`
}

// ExportedImplementationSummary summarises a public impl block.
type ExportedImplementationSummary struct {
	ImplName      string                             `json:"implName,omitempty"`
	InterfaceName string                             `json:"interface"`
	Target        string                             `json:"target"`
	InterfaceArgs []string                           `json:"interfaceArgs,omitempty"`
	TypeParams    []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Methods       map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where         []ExportedWhereConstraintSummary   `json:"where,omitempty"`
	Obligations   []ExportedObligationSummary        `json:"obligations,omitempty"`
}

// ExportedMethodSetSummary summarises a public methods block.
type ExportedMethodSetSummary struct {
	TypeParams  []ExportedGenericParamSummary      `json:"typeParams,omitempty"`
	Target      string                             `json:"target"`
	Methods     map[string]ExportedFunctionSummary `json:"methods,omitempty"`
	Where       []ExportedWhereConstraintSummary   `json:"where,omitempty"`
	Obligations []ExportedObligationSummary        `json:"obligations,omitempty"`
}

// PackageSummary captures the public API surface exported by a package.
type PackageSummary struct {
	Name            string                              `json:"name"`
	Visibility      string                              `json:"visibility"`
	Symbols         map[string]ExportedSymbolSummary    `json:"symbols"`
	PrivateSymbols  map[string]ExportedSymbolSummary    `json:"privateSymbols"`
	Structs         map[string]ExportedStructSummary    `json:"structs"`
	Interfaces      map[string]ExportedInterfaceSummary `json:"interfaces"`
	Functions       map[string]ExportedFunctionSummary  `json:"functions"`
	Implementations []ExportedImplementationSummary     `json:"implementations"`
	MethodSets      []ExportedMethodSetSummary          `json:"methodSets"`
}

// CheckResult aggregates diagnostics and package summaries for a program check.
type CheckResult struct {
	Diagnostics []ModuleDiagnostic
	Packages    map[string]PackageSummary
}

type packageExports struct {
	name       string
	visibility string
	symbols    map[string]Type
	private    map[string]Type
	impls      []ImplementationSpec
	methodSets []MethodSetSpec
	structs    map[string]StructType
	interfaces map[string]InterfaceType
	functions  map[string]FunctionType
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

		pc.captureExports(mod, checker)
	}
	return CheckResult{
		Diagnostics: diagnostics,
		Packages:    pc.clonePackageSummaries(),
	}, nil
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
		if _, already := seen[pkgName]; !already {
			if len(export.impls) > 0 {
				impls = append(impls, export.impls...)
			}
			if len(export.methodSets) > 0 {
				methods = append(methods, export.methodSets...)
			}
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
		}
		pkgType := PackageType{
			Package: pkgName,
			Symbols: nil,
		}
		if export != nil {
			pkgType.Symbols = export.symbols
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
		case ProcType:
			return "Proc"
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
	for _, impl := range export.impls {
		privacy := make(map[string]bool)
		if impl.Definition != nil {
			for _, fn := range impl.Definition.Definitions {
				if fn == nil || fn.ID == nil {
					continue
				}
				privacy[fn.ID.Name] = fn.IsPrivate
			}
		}
		for name, fn := range impl.Methods {
			export.functions[name] = fn
			if privacy[name] {
				export.private[name] = fn
				delete(export.symbols, name)
			} else {
				export.symbols[name] = fn
				delete(export.private, name)
			}
		}
	}
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
	}
	if node != nil && mod.NodeOrigins != nil {
		if path, ok := mod.NodeOrigins[node]; ok && path != "" {
			hint.Path = path
			return hint
		}
	}
	if len(mod.Files) > 0 {
		hint.Path = mod.Files[0]
	}
	return hint
}

func summarizeStructType(src StructType) ExportedStructSummary {
	summary := ExportedStructSummary{
		TypeParams: summarizeGenericParams(src.TypeParams),
		Fields:     summarizeTypeMap(src.Fields),
		Positional: summarizeTypeSlice(src.Positional),
		Where:      summarizeWhereConstraints(src.Where),
	}
	if summary.Fields == nil {
		summary.Fields = map[string]string{}
	}
	if summary.Positional == nil {
		summary.Positional = []string{}
	}
	if summary.TypeParams == nil {
		summary.TypeParams = []ExportedGenericParamSummary{}
	}
	if summary.Where == nil {
		summary.Where = []ExportedWhereConstraintSummary{}
	}
	return summary
}

func summarizeInterfaceType(src InterfaceType) ExportedInterfaceSummary {
	methods := make(map[string]ExportedFunctionSummary, len(src.Methods))
	for name, fn := range src.Methods {
		methods[name] = summarizeFunctionType(fn)
	}
	if methods == nil {
		methods = map[string]ExportedFunctionSummary{}
	}
	return ExportedInterfaceSummary{
		TypeParams: summarizeGenericParams(src.TypeParams),
		Methods:    methods,
		Where:      summarizeWhereConstraints(src.Where),
	}
}

func summarizeFunctionType(src FunctionType) ExportedFunctionSummary {
	return ExportedFunctionSummary{
		Parameters:  summarizeTypeSlice(src.Params),
		ReturnType:  formatType(src.Return),
		TypeParams:  summarizeGenericParams(src.TypeParams),
		Where:       summarizeWhereConstraints(src.Where),
		Obligations: summarizeObligations(src.Obligations),
	}
}

func summarizeImplementation(src ImplementationSpec) ExportedImplementationSummary {
	return ExportedImplementationSummary{
		ImplName:      src.ImplName,
		InterfaceName: src.InterfaceName,
		Target:        formatType(src.Target),
		InterfaceArgs: summarizeTypeSlice(src.InterfaceArgs),
		TypeParams:    summarizeGenericParams(src.TypeParams),
		Methods:       summarizeFunctionMap(src.Methods),
		Where:         summarizeWhereConstraints(src.Where),
		Obligations:   summarizeObligations(src.Obligations),
	}
}

func summarizeMethodSet(src MethodSetSpec) ExportedMethodSetSummary {
	qualifier := typeName(src.Target)
	methods := src.Methods
	if len(src.TypeQualified) > 0 && qualifier != "" {
		remapped := make(map[string]FunctionType, len(src.Methods))
		for name, fn := range src.Methods {
			key := name
			if src.TypeQualified != nil && src.TypeQualified[name] && qualifier != "" {
				key = fmt.Sprintf("%s.%s", qualifier, name)
			}
			remapped[key] = fn
		}
		methods = remapped
	}
	return ExportedMethodSetSummary{
		TypeParams:  summarizeGenericParams(src.TypeParams),
		Target:      formatType(src.Target),
		Methods:     summarizeFunctionMap(methods),
		Where:       summarizeWhereConstraints(src.Where),
		Obligations: summarizeObligations(src.Obligations),
	}
}

func summarizeGenericParams(params []GenericParamSpec) []ExportedGenericParamSummary {
	if len(params) == 0 {
		return nil
	}
	out := make([]ExportedGenericParamSummary, len(params))
	for i, param := range params {
		out[i] = ExportedGenericParamSummary{
			Name:        param.Name,
			Constraints: summarizeTypeSlice(param.Constraints),
		}
		if len(out[i].Constraints) == 0 {
			out[i].Constraints = nil
		}
	}
	return out
}

func summarizeWhereConstraints(constraints []WhereConstraintSpec) []ExportedWhereConstraintSummary {
	if len(constraints) == 0 {
		return nil
	}
	out := make([]ExportedWhereConstraintSummary, len(constraints))
	for i, constraint := range constraints {
		out[i] = ExportedWhereConstraintSummary{
			TypeParam:   constraint.TypeParam,
			Constraints: summarizeTypeSlice(constraint.Constraints),
		}
		if len(out[i].Constraints) == 0 {
			out[i].Constraints = nil
		}
	}
	return out
}

func summarizeObligations(obligations []ConstraintObligation) []ExportedObligationSummary {
	if len(obligations) == 0 {
		return nil
	}
	out := make([]ExportedObligationSummary, len(obligations))
	for i, ob := range obligations {
		out[i] = ExportedObligationSummary{
			Owner:      ob.Owner,
			TypeParam:  ob.TypeParam,
			Constraint: formatType(ob.Constraint),
			Subject:    formatType(ob.Subject),
			Context:    ob.Context,
		}
	}
	return out
}

func summarizeTypeMap(src map[string]Type) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for name, typ := range src {
		out[name] = formatType(typ)
	}
	return out
}

func summarizeFunctionMap(src map[string]FunctionType) map[string]ExportedFunctionSummary {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]ExportedFunctionSummary, len(src))
	for name, fn := range src {
		out[name] = summarizeFunctionType(fn)
	}
	return out
}

func summarizeTypeSlice(src []Type) []string {
	if len(src) == 0 {
		return nil
	}
	out := make([]string, len(src))
	for i, typ := range src {
		out[i] = formatType(typ)
	}
	return out
}
