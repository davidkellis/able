package parser

import "able/interpreter-go/pkg/ast"

// NormalizeFixtureModule cleans up optional fields so exported fixture JSON stays stable.
func NormalizeFixtureModule(mod *ast.Module) {
	if mod == nil {
		return
	}
	for _, stmt := range mod.Body {
		normalizeFixtureNode(stmt)
	}
}

func normalizeFixtureNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.FunctionDefinition:
		if len(n.GenericParams) == 0 {
			n.GenericParams = nil
		}
		if len(n.WhereClause) == 0 {
			n.WhereClause = nil
		}
		if n.Body != nil {
			normalizeFixtureNode(n.Body)
		}
	case *ast.FunctionSignature:
		if len(n.GenericParams) == 0 {
			n.GenericParams = nil
		}
		if len(n.WhereClause) == 0 {
			n.WhereClause = nil
		}
		if n.DefaultImpl != nil {
			normalizeFixtureNode(n.DefaultImpl)
		}
	case *ast.MethodsDefinition:
		for _, def := range n.Definitions {
			normalizeFixtureNode(def)
		}
	case *ast.InterfaceDefinition:
		if len(n.GenericParams) == 0 {
			n.GenericParams = nil
		}
		if len(n.WhereClause) == 0 {
			n.WhereClause = nil
		}
		for _, sig := range n.Signatures {
			normalizeFixtureNode(sig)
		}
	case *ast.ExternFunctionBody:
		if len(n.Signature.GenericParams) == 0 {
			n.Signature.GenericParams = nil
		}
		if len(n.Signature.WhereClause) == 0 {
			n.Signature.WhereClause = nil
		}
	case *ast.StructDefinition:
		if len(n.GenericParams) == 0 {
			n.GenericParams = nil
		}
		if len(n.WhereClause) == 0 {
			n.WhereClause = nil
		}
	case *ast.StructLiteral:
		for _, field := range n.Fields {
			if field != nil {
				normalizeFixtureNode(field.Value)
			}
		}
	case *ast.StructFieldInitializer:
		normalizeFixtureNode(n.Value)
	case *ast.FunctionCall:
		if len(n.TypeArguments) == 0 {
			n.TypeArguments = nil
		}
		for _, arg := range n.Arguments {
			normalizeFixtureNode(arg)
		}
	case *ast.AssignmentExpression:
		normalizeFixtureNode(n.Right)
	case *ast.BlockExpression:
		for _, stmt := range n.Body {
			normalizeFixtureNode(stmt)
		}
	case *ast.ArrayLiteral:
		for _, elem := range n.Elements {
			normalizeFixtureNode(elem)
		}
	case *ast.GenericTypeExpression:
		if len(n.Arguments) == 0 {
			n.Arguments = nil
		}
	case *ast.NullableTypeExpression:
		if n.InnerType != nil {
			normalizeFixtureNode(n.InnerType)
		}
	case *ast.ResultTypeExpression:
		if n.InnerType != nil {
			normalizeFixtureNode(n.InnerType)
		}
	case *ast.UnionTypeExpression:
		for _, member := range n.Members {
			normalizeFixtureNode(member)
		}
	case *ast.MatchExpression:
		normalizeFixtureNode(n.Subject)
		for _, clause := range n.Clauses {
			normalizeFixtureNode(clause)
		}
	case *ast.RescueExpression:
		normalizeFixtureNode(n.MonitoredExpression)
		for _, clause := range n.Clauses {
			normalizeFixtureNode(clause)
		}
	case *ast.EnsureExpression:
		normalizeFixtureNode(n.TryExpression)
		if n.EnsureBlock != nil {
			normalizeFixtureNode(n.EnsureBlock)
		}
	case *ast.BreakpointExpression:
		if n.Body != nil {
			normalizeFixtureNode(n.Body)
		}
	case *ast.MatchClause:
		if n.Guard != nil {
			normalizeFixtureNode(n.Guard)
		}
		normalizeFixtureNode(n.Body)
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			normalizeFixtureNode(part)
		}
	case *ast.ArrayPattern:
		if n.RestPattern != nil {
			normalizeFixtureNode(n.RestPattern)
		}
		for _, elem := range n.Elements {
			normalizeFixtureNode(elem)
		}
	case *ast.StructPattern:
		for _, field := range n.Fields {
			if field.Pattern != nil {
				normalizeFixtureNode(field.Pattern)
			}
			if field.TypeAnnotation != nil {
				normalizeFixtureNode(field.TypeAnnotation)
			}
		}
	case *ast.TypedPattern:
		if n.Pattern != nil {
			normalizeFixtureNode(n.Pattern)
		}
		if n.TypeAnnotation != nil {
			normalizeFixtureNode(n.TypeAnnotation)
		}
	case *ast.FunctionParameter:
		if n.ParamType != nil {
			normalizeFixtureNode(n.ParamType)
		}
	case *ast.ImplementationDefinition:
		if len(n.GenericParams) == 0 {
			n.GenericParams = nil
		}
		if len(n.InterfaceArgs) == 0 {
			n.InterfaceArgs = nil
		}
		if len(n.WhereClause) == 0 {
			n.WhereClause = nil
		}
		for _, def := range n.Definitions {
			normalizeFixtureNode(def)
		}
	case *ast.FunctionTypeExpression:
		if len(n.ParamTypes) == 0 {
			n.ParamTypes = nil
		}
		if n.ReturnType != nil {
			normalizeFixtureNode(n.ReturnType)
		}
	}
}
