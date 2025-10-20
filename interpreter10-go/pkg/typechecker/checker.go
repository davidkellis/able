package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

// Checker traverses Able AST nodes and records diagnostics.
type Checker struct {
	infer               InferenceMap
	global              *Environment
	returnTypeStack     []Type
	rescueDepth         int
	loopDepth           int
	loopResultStack     []Type
	breakpointStack     []string
	asyncDepth          int
	implementations     []ImplementationSpec
	methodSets          []MethodSetSpec
	obligations         []ConstraintObligation
	constraintStack     []map[string][]Type
	allowDynamicLookups bool
}

// Diagnostic represents a type-checking error or warning.
type Diagnostic struct {
	Message string
	Node    ast.Node
}

// New returns a checker instance.
func New() *Checker {
	return &Checker{
		infer:           make(InferenceMap),
		global:          NewEnvironment(nil),
		returnTypeStack: nil,
		rescueDepth:     0,
	}
}

// CheckModule performs typechecking on a module AST and returns diagnostics.
func (c *Checker) CheckModule(module *ast.Module) ([]Diagnostic, error) {
	if module == nil {
		return nil, fmt.Errorf("typechecker: module is nil")
	}
	// Reset inference map between runs.
	c.infer = make(InferenceMap)
	c.returnTypeStack = nil
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
	declDiags := c.collectDeclarations(module)
	var diagnostics []Diagnostic
	diagnostics = append(diagnostics, declDiags...)

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
		if imp.IsWildcard {
			c.allowDynamicLookups = true
		}
		if imp.Alias != nil && imp.Alias.Name != "" {
			env.Define(imp.Alias.Name, placeholder)
			continue
		}
		for _, sel := range imp.Selectors {
			if sel == nil {
				continue
			}
			if sel.Alias != nil && sel.Alias.Name != "" {
				env.Define(sel.Alias.Name, placeholder)
				continue
			}
			if sel.Name != nil && sel.Name.Name != "" {
				env.Define(sel.Name.Name, placeholder)
			}
		}
	}
}
