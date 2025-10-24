package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/typechecker"
)

// ModuleDiagnostic represents a diagnostic associated with a module in a program.
type ModuleDiagnostic struct {
	Package    string
	Files      []string
	Diagnostic typechecker.Diagnostic
}

// DescribeModuleDiagnostic formats a module diagnostic for display.
func DescribeModuleDiagnostic(diag ModuleDiagnostic) string {
	message := diag.Diagnostic.Message
	if diag.Package != "" {
		message = fmt.Sprintf("%s: %s", diag.Package, message)
	}
	if len(diag.Files) > 0 {
		message = fmt.Sprintf("%s (e.g., %s)", message, diag.Files[0])
	}
	return message
}

// TypecheckProgram applies the Able typechecker across all modules in the provided program,
// returning diagnostics ordered by module evaluation order. An empty slice indicates success.
func TypecheckProgram(program *driver.Program) ([]ModuleDiagnostic, error) {
	tc := newProgramTypechecker()
	return tc.Check(program)
}

type packageExports struct {
	name       string
	symbols    map[string]typechecker.Type
	impls      []typechecker.ImplementationSpec
	methodSets []typechecker.MethodSetSpec
}

type programTypechecker struct {
	exports map[string]*packageExports
}

func newProgramTypechecker() *programTypechecker {
	return &programTypechecker{
		exports: make(map[string]*packageExports),
	}
}

func (pt *programTypechecker) Check(program *driver.Program) ([]ModuleDiagnostic, error) {
	if program == nil {
		return nil, fmt.Errorf("typechecker: program is nil")
	}
	var results []ModuleDiagnostic
	for _, mod := range program.Modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		env, impls, methods, importDiags := pt.buildPrelude(mod.AST.Imports)
		checker := typechecker.New()
		checker.SetPrelude(env, impls, methods)

		moduleDiags, err := checker.CheckModule(mod.AST)
		if err != nil {
			return results, err
		}

		for _, diag := range importDiags {
			results = append(results, ModuleDiagnostic{
				Package:    mod.Package,
				Files:      mod.Files,
				Diagnostic: diag,
			})
		}
		for _, diag := range moduleDiags {
			results = append(results, ModuleDiagnostic{
				Package:    mod.Package,
				Files:      mod.Files,
				Diagnostic: diag,
			})
		}

		pt.captureExports(mod, checker)
	}
	return results, nil
}

func (pt *programTypechecker) buildPrelude(imports []*ast.ImportStatement) (*typechecker.Environment, []typechecker.ImplementationSpec, []typechecker.MethodSetSpec, []typechecker.Diagnostic) {
	if len(imports) == 0 {
		return nil, nil, nil, nil
	}
	env := typechecker.NewEnvironment(nil)
	var (
		impls    []typechecker.ImplementationSpec
		methods  []typechecker.MethodSetSpec
		diags    []typechecker.Diagnostic
		hasScope bool
	)
	seen := make(map[string]struct{})

	for _, imp := range imports {
		if imp == nil {
			continue
		}
		pkgName := joinImportPath(imp.PackagePath)
		export, ok := pt.exports[pkgName]
		if ok {
			if _, already := seen[pkgName]; !already {
				if len(export.impls) > 0 {
					impls = append(impls, export.impls...)
				}
				if len(export.methodSets) > 0 {
					methods = append(methods, export.methodSets...)
				}
				seen[pkgName] = struct{}{}
			}
		} else {
			diags = append(diags, typechecker.Diagnostic{
				Message: fmt.Sprintf("typechecker: import references unknown package '%s'", pkgName),
				Node:    imp,
			})
		}

		if imp.IsWildcard {
			if ok {
				for name, typ := range export.symbols {
					if typ == nil {
						typ = typechecker.UnknownType{}
					}
					env.Define(name, typ)
					hasScope = true
				}
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
				if ok {
					if typ, exists := export.symbols[sel.Name.Name]; exists && typ != nil {
						env.Define(alias, typ)
						hasScope = true
						continue
					}
					diags = append(diags, typechecker.Diagnostic{
						Message: fmt.Sprintf("typechecker: package '%s' has no symbol '%s'", pkgName, sel.Name.Name),
						Node:    sel,
					})
				}
				env.Define(alias, typechecker.UnknownType{})
				hasScope = true
			}
			continue
		}

		alias := pkgName
		if imp.Alias != nil && imp.Alias.Name != "" {
			alias = imp.Alias.Name
		}
		pkgType := typechecker.PackageType{
			Package: pkgName,
			Symbols: nil,
		}
		if ok {
			pkgType.Symbols = export.symbols
		} else {
			pkgType.Symbols = map[string]typechecker.Type{}
		}
		env.Define(alias, pkgType)
		hasScope = true
	}

	if !hasScope {
		env = nil
	}
	return env, impls, methods, diags
}

func (pt *programTypechecker) captureExports(mod *driver.Module, checker *typechecker.Checker) {
	if mod == nil {
		return
	}
	pkg := mod.Package
	export := &packageExports{
		name:    pkg,
		symbols: make(map[string]typechecker.Type),
	}
	for _, sym := range checker.ExportedSymbols() {
		export.symbols[sym.Name] = sym.Type
	}
	for _, impl := range checker.ModuleImplementations() {
		if impl.Definition != nil && impl.Definition.IsPrivate {
			continue
		}
		export.impls = append(export.impls, impl)
	}
	for _, methods := range checker.ModuleMethodSets() {
		export.methodSets = append(export.methodSets, methods)
	}
	pt.exports[pkg] = export
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
