package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// declarationCollector walks statements to populate the global environment.
type declarationCollector struct {
	env         *Environment
	origins     map[ast.Node]string
	declNodes   map[string]ast.Node
	diags       []Diagnostic
	impls       []ImplementationSpec
	methodSets  []MethodSetSpec
	obligations []ConstraintObligation
	exports     []exportRecord
	duplicates  map[*ast.FunctionDefinition]struct{}
}

func (c *Checker) collectDeclarations(module *ast.Module) []Diagnostic {
	builtinEnv := NewEnvironment(nil)
	registerBuiltins(builtinEnv)
	rootEnv := NewEnvironment(builtinEnv)
	if c.preludeEnv != nil {
		c.preludeEnv.ForEach(func(name string, typ Type) {
			rootEnv.Define(name, typ)
		})
	}
	collector := &declarationCollector{
		env:        rootEnv,
		origins:    c.nodeOrigins,
		declNodes:  make(map[string]ast.Node),
		duplicates: make(map[*ast.FunctionDefinition]struct{}),
	}
	// Register built-in primitives in the global scope for convenience.
	collector.env.Define("true", PrimitiveType{Kind: PrimitiveBool})
	collector.env.Define("false", PrimitiveType{Kind: PrimitiveBool})

	for _, stmt := range module.Body {
		collector.registerTypeDeclaration(stmt)
	}
	for _, stmt := range module.Body {
		collector.visitStatement(stmt)
	}

	// Store the global environment for later lookups.
	c.global = collector.env
	c.implementations = collector.impls
	c.methodSets = collector.methodSets
	c.obligations = collector.obligations
	c.publicDeclarations = collector.exports
	c.duplicateFunctions = collector.duplicates
	return collector.diags
}

func (c *declarationCollector) registerExternFunction(def *ast.ExternFunctionBody) {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return
	}
	sig := def.Signature
	name := sig.ID.Name
	owner := fmt.Sprintf("extern fn %s", functionName(sig))
	fnType := c.functionTypeFromDefinition(sig, nil, owner, sig)
	if prev, exists := c.declNodes[name]; exists {
		if existing, ok := c.env.Lookup(name); ok {
			if prevFn, ok := existing.(FunctionType); ok && typesEquivalentForSignature(prevFn, fnType) {
				return
			}
		}
		location := formatNodeLocation(prev, c.origins)
		msg := fmt.Sprintf("typechecker: duplicate declaration '%s' (previous declaration at %s)", name, location)
		c.diags = append(c.diags, Diagnostic{Message: msg, Node: def})
		if c.duplicates != nil {
			c.duplicates[def.Signature] = struct{}{}
		}
		return
	}
	c.env.Define(name, fnType)
	c.declNodes[name] = sig
	if shouldExportTopLevel(sig) {
		c.exports = append(c.exports, exportRecord{name: name, node: sig})
	}
}

func (c *declarationCollector) registerTypeDeclaration(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.StructDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			fields, positional := c.collectStructFields(s, paramScope)
			structType := StructType{
				StructName: s.ID.Name,
				TypeParams: params,
				Fields:     fields,
				Positional: positional,
				Where:      where,
			}
			c.declare(s.ID.Name, structType, s)
		}
	case *ast.UnionDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			unionType := UnionType{
				UnionName:  s.ID.Name,
				TypeParams: params,
				Where:      where,
				Variants:   make([]Type, 0, len(s.Variants)),
			}
			if len(s.Variants) > 0 {
				if paramScope == nil {
					paramScope = make(map[string]Type)
				}
				for _, variant := range s.Variants {
					if variant == nil {
						continue
					}
					unionType.Variants = append(unionType.Variants, c.resolveTypeExpression(variant, paramScope))
				}
			}
			c.declare(s.ID.Name, unionType, s)
		}
	case *ast.InterfaceDefinition:
		if s.ID != nil {
			params, paramScope := c.convertGenericParams(s.GenericParams)
			where := c.convertWhereClause(s.WhereClause, paramScope)
			if paramScope == nil {
				paramScope = make(map[string]Type)
			}
			if _, exists := paramScope[s.ID.Name]; !exists {
				paramScope[s.ID.Name] = InterfaceType{InterfaceName: s.ID.Name}
			}
			if _, exists := paramScope["Self"]; !exists {
				paramScope["Self"] = TypeParameterType{ParameterName: "Self"}
			}
			methods := c.collectInterfaceMethods(s, paramScope)
			ifaceType := InterfaceType{
				InterfaceName:   s.ID.Name,
				TypeParams:      params,
				Where:           where,
				Methods:         methods,
				SelfTypePattern: s.SelfTypePattern,
			}
			c.declare(s.ID.Name, ifaceType, s)
		}
	case *ast.TypeAliasDefinition:
		c.collectTypeAliasDefinition(s)
	}
}

func (c *declarationCollector) visitStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		if s.ID != nil {
			owner := fmt.Sprintf("fn %s", functionName(s))
			sig := c.functionTypeFromDefinition(s, nil, owner, s)
			c.declare(s.ID.Name, sig, s)
		}
	case *ast.ExternFunctionBody:
		c.registerExternFunction(s)
	case *ast.ImplementationDefinition:
		spec, diags := c.collectImplementationDefinition(s)
		c.diags = append(c.diags, diags...)
		if spec != nil {
			c.impls = append(c.impls, *spec)
		}
	case *ast.MethodsDefinition:
		spec, diags := c.collectMethodsDefinition(s)
		c.diags = append(c.diags, diags...)
		if spec != nil {
			c.methodSets = append(c.methodSets, *spec)
		}
	}
}

func (c *declarationCollector) declare(name string, typ Type, node ast.Node) {
	if name == "" || node == nil {
		return
	}
	if prev, exists := c.declNodes[name]; exists {
		location := formatNodeLocation(prev, c.origins)
		msg := fmt.Sprintf("typechecker: duplicate declaration '%s' (previous declaration at %s)", name, location)
		c.diags = append(c.diags, Diagnostic{Message: msg, Node: node})
		if fn, ok := node.(*ast.FunctionDefinition); ok {
			if c.duplicates != nil {
				c.duplicates[fn] = struct{}{}
			}
		}
		return
	}
	c.env.Define(name, typ)
	c.declNodes[name] = node
	if shouldExportTopLevel(node) {
		c.exports = append(c.exports, exportRecord{name: name, node: node})
	}
}

func shouldExportTopLevel(node ast.Node) bool {
	switch def := node.(type) {
	case *ast.StructDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.UnionDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.InterfaceDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.FunctionDefinition:
		if def == nil || def.ID == nil {
			return false
		}
		return !def.IsPrivate
	case *ast.TypeAliasDefinition:
		return def != nil && def.ID != nil && !def.IsPrivate
	case *ast.ImplementationDefinition:
		return def != nil && def.ImplName != nil && def.ImplName.Name != "" && !def.IsPrivate
	default:
		return false
	}
}

func formatNodeLocation(node ast.Node, origins map[ast.Node]string) string {
	if node == nil {
		return "<unknown location>"
	}
	path := "<unknown file>"
	if origins != nil {
		if origin, ok := origins[node]; ok && origin != "" {
			path = normalizeDiagnosticPath(origin)
		}
	}
	span := node.Span()
	line := span.Start.Line
	column := span.Start.Column
	if line <= 0 {
		line = 0
	}
	if column <= 0 {
		column = 0
	}
	return fmt.Sprintf("%s:%d:%d", path, line, column)
}
