package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// Checker traverses Able AST nodes and records diagnostics.
type Checker struct {
	infer                InferenceMap
	global               *Environment
	nodeOrigins          map[ast.Node]string
	returnTypeStack      []Type
	functionGenericStack []functionGenericContext
	rescueDepth          int
	loopDepth            int
	loopResultStack      []Type
	breakpointStack      []string
	asyncDepth           int
	implementations      []ImplementationSpec
	methodSets           []MethodSetSpec
	obligations          []ConstraintObligation
	constraintStack      []map[string][]Type
	allowDynamicLookups  bool
	preludeEnv           *Environment
	preludeImpls         []ImplementationSpec
	preludeMethodSets    []MethodSetSpec
	preludeImplCount     int
	preludeMethodCount   int
	publicDeclarations   []exportRecord
	localTypeNames       map[string]struct{}

	builtinImplementations []ImplementationSpec
	pendingDiagnostics     []Diagnostic
	duplicateFunctions     map[*ast.FunctionDefinition]struct{}
}

// DiagnosticSeverity conveys the diagnostic level.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
)

// DiagnosticNote captures secondary context for a diagnostic.
type DiagnosticNote struct {
	Message string
	Node    ast.Node
}

// Diagnostic represents a type-checking error or warning.
type Diagnostic struct {
	Severity DiagnosticSeverity
	Message  string
	Node     ast.Node
	Notes    []DiagnosticNote
}

type exportRecord struct {
	name string
	node ast.Node
}

type functionGenericContext struct {
	def      *ast.FunctionDefinition
	label    string
	inferred map[string]*ast.GenericParameter
}

// ExportedSymbol describes a public binding produced by a module.
type ExportedSymbol struct {
	Name string
	Type Type
	Node ast.Node
}

// New returns a checker instance.
func New() *Checker {
	c := &Checker{
		infer:           make(InferenceMap),
		global:          NewEnvironment(nil),
		nodeOrigins:     nil,
		returnTypeStack: nil,
		rescueDepth:     0,
	}
	c.initBuiltinInterfaces()
	return c
}

// SetNodeOrigins attaches origin metadata for diagnostics.
func (c *Checker) SetNodeOrigins(origins map[ast.Node]string) {
	c.nodeOrigins = origins
}

// SetPrelude seeds the checker with bindings and implementation metadata that
// should be visible before processing the next module.
func (c *Checker) SetPrelude(env *Environment, impls []ImplementationSpec, methods []MethodSetSpec) {
	if env != nil {
		c.preludeEnv = env.Clone()
	} else {
		c.preludeEnv = nil
	}
	c.preludeImpls = append(c.preludeImpls[:0], impls...)
	c.preludeMethodSets = append(c.preludeMethodSets[:0], methods...)
}

// CheckModule performs typechecking on a module AST and returns diagnostics.
func (c *Checker) CheckModule(module *ast.Module) ([]Diagnostic, error) {
	if module == nil {
		return nil, fmt.Errorf("typechecker: module is nil")
	}
	// Reset inference map between runs.
	c.infer = make(InferenceMap)
	c.returnTypeStack = nil
	c.functionGenericStack = nil
	c.rescueDepth = 0
	c.loopDepth = 0
	c.loopResultStack = nil
	c.breakpointStack = nil
	c.asyncDepth = 0
	c.implementations = nil
	c.methodSets = nil
	c.obligations = nil
	c.constraintStack = nil
	c.allowDynamicLookups = false
	c.publicDeclarations = nil
	c.preludeImplCount = 0
	c.preludeMethodCount = 0
	c.pendingDiagnostics = nil
	c.duplicateFunctions = nil
	declDiags := c.collectDeclarations(module)
	var diagnostics []Diagnostic
	diagnostics = append(diagnostics, declDiags...)

	activeBuiltins := c.builtinImplsForModule(module)
	builtinCount := len(activeBuiltins)
	if builtinCount > 0 {
		base := make([]ImplementationSpec, builtinCount)
		copy(base, activeBuiltins)
		c.implementations = append(base, c.implementations...)
		c.preludeImplCount = builtinCount
	} else {
		c.preludeImplCount = 0
	}
	if len(c.preludeImpls) > 0 {
		base := make([]ImplementationSpec, len(c.preludeImpls))
		copy(base, c.preludeImpls)
		c.implementations = append(base, c.implementations...)
		c.preludeImplCount += len(base)
	}
	if len(c.preludeMethodSets) > 0 {
		base := make([]MethodSetSpec, len(c.preludeMethodSets))
		copy(base, c.preludeMethodSets)
		c.methodSets = append(base, c.methodSets...)
		c.preludeMethodCount = len(base)
	} else {
		c.preludeMethodCount = 0
	}

	env := c.global.Extend()
	c.applyImports(env, module.Imports)
	for _, stmt := range module.Body {
		stDiags := c.checkStatement(env, stmt)
		diagnostics = append(diagnostics, stDiags...)
	}

	constraintDiags := c.resolveObligations()
	diagnostics = append(diagnostics, constraintDiags...)

	implDiags := c.validateImplementations()
	diagnostics = append(diagnostics, implDiags...)

	if len(c.pendingDiagnostics) > 0 {
		diagnostics = append(diagnostics, c.pendingDiagnostics...)
	}
	return diagnostics, nil
}

func (c *Checker) applyImports(env *Environment, imports []*ast.ImportStatement) {
	if env == nil || len(imports) == 0 {
		return
	}
	placeholder := Type(UnknownType{})
	for _, imp := range imports {
		if imp == nil {
			continue
		}
		if imp.Alias != nil && imp.Alias.Name != "" {
			if _, exists := env.Lookup(imp.Alias.Name); exists {
				continue
			}
			env.Define(imp.Alias.Name, placeholder)
			continue
		}
		if !imp.IsWildcard && len(imp.Selectors) == 0 {
			if alias := lastImportSegment(imp.PackagePath); alias != "" {
				if _, exists := env.Lookup(alias); !exists {
					env.Define(alias, placeholder)
				}
			}
			continue
		}
		for _, sel := range imp.Selectors {
			if sel == nil {
				continue
			}
			if sel.Alias != nil && sel.Alias.Name != "" {
				if _, exists := env.Lookup(sel.Alias.Name); exists {
					continue
				}
				env.Define(sel.Alias.Name, placeholder)
				continue
			}
			if sel.Name != nil && sel.Name.Name != "" {
				if _, exists := env.Lookup(sel.Name.Name); exists {
					continue
				}
				env.Define(sel.Name.Name, placeholder)
			}
		}
	}
}

func lastImportSegment(parts []*ast.Identifier) string {
	if len(parts) == 0 {
		return ""
	}
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != nil && parts[i].Name != "" {
			return parts[i].Name
		}
	}
	return ""
}

func (c *Checker) addDiagnostic(diag Diagnostic) {
	if diag.Message == "" {
		return
	}
	c.pendingDiagnostics = append(c.pendingDiagnostics, diag)
}

func (c *Checker) verifyAliasConstraints(alias AliasType, subst map[string]Type, node ast.Node) {
	if len(alias.Obligations) == 0 {
		return
	}
	obligations := alias.Obligations
	if len(subst) > 0 {
		obligations = substituteObligations(obligations, subst)
	}
	ok, detail, _ := c.obligationSetSatisfied(obligations)
	if ok {
		return
	}
	message := detail
	if message == "" {
		message = fmt.Sprintf("type alias '%s' constraints not satisfied", alias.AliasName)
	} else {
		message = fmt.Sprintf("type alias '%s': %s", alias.AliasName, detail)
	}
	diagNode := node
	if diagNode == nil {
		diagNode = alias.Definition
	}
	c.addDiagnostic(Diagnostic{
		Message: message,
		Node:    diagNode,
	})
}

// ExportedSymbols returns the public bindings declared in the last module that was checked.
func (c *Checker) ExportedSymbols() []ExportedSymbol {
	if len(c.publicDeclarations) == 0 || c.global == nil {
		return nil
	}
	out := make([]ExportedSymbol, 0, len(c.publicDeclarations))
	for _, rec := range c.publicDeclarations {
		if rec.name == "" {
			continue
		}
		typ, ok := c.global.Lookup(rec.name)
		if !ok || typ == nil {
			continue
		}
		out = append(out, ExportedSymbol{
			Name: rec.name,
			Type: typ,
			Node: rec.node,
		})
	}
	return out
}

// ModuleImplementations returns the implementation specs declared in the last module, excluding prelude entries.
func (c *Checker) ModuleImplementations() []ImplementationSpec {
	total := len(c.implementations)
	if total == 0 || c.preludeImplCount >= total {
		return nil
	}
	count := total - c.preludeImplCount
	out := make([]ImplementationSpec, count)
	copy(out, c.implementations[c.preludeImplCount:])
	return out
}

func (c *Checker) builtinImplsForModule(module *ast.Module) []ImplementationSpec {
	if len(c.builtinImplementations) == 0 {
		return nil
	}
	disableDisplay := moduleDefinesInterface(module, "Display")
	disableClone := moduleDefinesInterface(module, "Clone")
	disableOrd := moduleDefinesInterface(module, "Ord")
	disableError := moduleDefinesInterface(module, "Error")
	if !disableDisplay && !disableClone && !disableOrd && !disableError {
		return c.builtinImplementations
	}
	filtered := make([]ImplementationSpec, 0, len(c.builtinImplementations))
	for _, impl := range c.builtinImplementations {
		switch impl.InterfaceName {
		case "Display":
			if disableDisplay {
				continue
			}
		case "Clone":
			if disableClone {
				continue
			}
		case "Ord":
			if disableOrd {
				continue
			}
		case "Error":
			if disableError {
				continue
			}
		}
		filtered = append(filtered, impl)
	}
	return filtered
}

func moduleDefinesInterface(module *ast.Module, name string) bool {
	if module == nil || name == "" {
		return false
	}
	for _, stmt := range module.Body {
		if iface, ok := stmt.(*ast.InterfaceDefinition); ok && iface != nil && iface.ID != nil && iface.ID.Name == name {
			return true
		}
	}
	return false
}

// ModuleMethodSets returns the method-set specs declared in the last module, excluding prelude entries.
func (c *Checker) ModuleMethodSets() []MethodSetSpec {
	total := len(c.methodSets)
	if total == 0 || c.preludeMethodCount >= total {
		return nil
	}
	count := total - c.preludeMethodCount
	out := make([]MethodSetSpec, count)
	copy(out, c.methodSets[c.preludeMethodCount:])
	return out
}

// GlobalEnvironment exposes the checker’s global environment (read-only).
func (c *Checker) GlobalEnvironment() *Environment {
	return c.global
}
