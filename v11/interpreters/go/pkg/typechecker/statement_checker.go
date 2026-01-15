package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) checkStatement(env *Environment, stmt ast.Statement) []Diagnostic {
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		if idx, ok := s.Left.(*ast.IndexExpression); ok {
			var diags []Diagnostic
			rhsDiags, typ := c.checkExpression(env, s.Right)
			diags = append(diags, rhsDiags...)
			if typ == nil {
				typ = UnknownType{}
			}
			diags = append(diags, c.checkIndexAssignment(env, idx, typ, s.Operator)...)
			c.infer.set(s, UnknownType{})
			return diags
		}
		if member, ok := s.Left.(*ast.MemberAccessExpression); ok {
			var diags []Diagnostic
			if s.Operator == ast.AssignmentDeclare {
				diags = append(diags, Diagnostic{
					Message: "typechecker: cannot declare new binding on member assignment",
					Node:    s,
				})
			}
			memberDiags, memberType := c.checkMemberAccess(env, member)
			diags = append(diags, memberDiags...)
			rhsDiags, rhsType := c.checkExpression(env, s.Right)
			diags = append(diags, rhsDiags...)
			if memberType == nil {
				memberType = UnknownType{}
			}
			if rhsType == nil {
				rhsType = UnknownType{}
			}
			if !isUnknownType(memberType) && !isUnknownType(rhsType) && !typeAssignable(rhsType, memberType) {
				if msg, ok := literalMismatchMessage(rhsType, memberType); ok {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: %s", msg),
						Node:    s.Right,
					})
				} else {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: cannot assign %s to member (expected %s)", typeName(rhsType), typeName(memberType)),
						Node:    s,
					})
				}
			}
			c.infer.set(s, UnknownType{})
			return diags
		}
		var diags []Diagnostic
		var intent *patternIntent
		if s.Operator == ast.AssignmentDeclare {
			newNames, hasAny := analyzeAssignmentTargets(env, s.Left)
			if hasAny && len(newNames) == 0 {
				diags = append(diags, Diagnostic{
					Message: "typechecker: ':=' requires at least one new binding",
					Node:    s.Left,
				})
			}
			intent = &patternIntent{declarationNames: newNames}
			diags = append(diags, c.bindPattern(env, s.Left, UnknownType{}, true, intent)...)
		}
		expectedType := Type(UnknownType{})
		if typed, ok := s.Left.(*ast.TypedPattern); ok && typed.TypeAnnotation != nil {
			expectedType = c.resolveTypeReference(typed.TypeAnnotation)
		} else if s.Operator == ast.AssignmentAssign {
			if ident, ok := s.Left.(*ast.Identifier); ok && ident.Name != "" {
				if existing, ok := env.Lookup(ident.Name); ok {
					expectedType = existing
				}
			}
		}
		rhsDiags, typ := c.checkExpressionWithExpectedType(env, s.Right, expectedType)
		diags = append(diags, rhsDiags...)
		if typ == nil {
			typ = UnknownType{}
		}
		if s.Operator == ast.AssignmentDeclare {
			return append(diags, c.bindPattern(env, s.Left, typ, true, intent)...)
		}
		if s.Operator == ast.AssignmentAssign {
			assignIntent := &patternIntent{allowFallback: true}
			return append(diags, c.bindPattern(env, s.Left, typ, false, assignIntent)...)
		}
		return diags
	case *ast.WhileLoop:
		diags, _ := c.checkWhileLoop(env, s)
		return diags
	case *ast.ForLoop:
		diags, _ := c.checkForLoop(env, s)
		return diags
	case *ast.RaiseStatement:
		return c.checkRaiseStatement(env, s)
	case *ast.RethrowStatement:
		return c.checkRethrowStatement(s)
	case *ast.BreakStatement:
		return c.checkBreakStatement(env, s)
	case *ast.ContinueStatement:
		return c.checkContinueStatement(s)
	case *ast.StructDefinition:
		return c.checkLocalTypeDeclaration(identifierName(s.ID), s)
	case *ast.UnionDefinition:
		return c.checkLocalTypeDeclaration(identifierName(s.ID), s)
	case *ast.InterfaceDefinition:
		return c.checkLocalTypeDeclaration(identifierName(s.ID), s)
	case *ast.TypeAliasDefinition:
		return c.checkLocalTypeDeclaration(identifierName(s.ID), s)
	case *ast.DynImportStatement:
		placeholder := Type(UnknownType{})
		if s.IsWildcard {
			c.allowDynamicLookups = true
		}
		if s.Alias != nil && s.Alias.Name != "" {
			env.Define(s.Alias.Name, placeholder)
		}
		for _, sel := range s.Selectors {
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
		return nil
	case *ast.FunctionDefinition:
		return c.checkFunctionDefinition(env, s)
	case *ast.ImplementationDefinition:
		return c.checkImplementationDefinition(env, s)
	case *ast.MethodsDefinition:
		return c.checkMethodsDefinition(env, s)
	case *ast.ExternFunctionBody:
		if s == nil || s.Signature == nil {
			return nil
		}
		return c.checkFunctionDefinition(env, s.Signature)
	case *ast.PreludeStatement:
		return nil
	case *ast.ReturnStatement:
		return c.checkReturnStatement(env, s)
	case ast.Expression:
		diags, _ := c.checkExpression(env, s)
		return diags
	default:
		return []Diagnostic{{Message: fmt.Sprintf("typechecker: unsupported statement %T", stmt), Node: stmt}}
	}
}

func (c *Checker) checkLocalTypeDeclaration(name string, node ast.Node) []Diagnostic {
	if name == "" {
		return nil
	}
	if len(c.functionGenericStack) == 0 {
		return nil
	}
	current := c.functionGenericStack[len(c.functionGenericStack)-1]
	if len(current.inferred) == 0 {
		return nil
	}
	param, ok := current.inferred[name]
	if !ok {
		return nil
	}
	location := formatNodeLocation(param, c.nodeOrigins)
	msg := fmt.Sprintf("typechecker: cannot redeclare inferred type parameter '%s' inside %s (inferred at %s)", name, current.label, location)
	return []Diagnostic{{Message: msg, Node: node}}
}

func analyzeAssignmentTargets(env *Environment, target ast.AssignmentTarget) (map[string]struct{}, bool) {
	names := make(map[string]struct{})
	collectAssignmentTargetIdentifiers(target, names)
	newNames := make(map[string]struct{})
	for name := range names {
		if !env.HasInCurrentScope(name) {
			newNames[name] = struct{}{}
		}
	}
	return newNames, len(names) > 0
}

func collectAssignmentTargetIdentifiers(target ast.AssignmentTarget, into map[string]struct{}) {
	switch t := target.(type) {
	case *ast.Identifier:
		if t.Name != "" {
			into[t.Name] = struct{}{}
		}
	case *ast.StructPattern:
		for _, field := range t.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" {
				into[field.Binding.Name] = struct{}{}
			}
			if inner, ok := field.Pattern.(ast.AssignmentTarget); ok {
				collectAssignmentTargetIdentifiers(inner, into)
			}
		}
	case *ast.ArrayPattern:
		for _, elem := range t.Elements {
			if inner, ok := elem.(ast.AssignmentTarget); ok {
				collectAssignmentTargetIdentifiers(inner, into)
			}
		}
		if rest, ok := t.RestPattern.(*ast.Identifier); ok && rest.Name != "" {
			into[rest.Name] = struct{}{}
		}
	case *ast.TypedPattern:
		if inner, ok := t.Pattern.(ast.AssignmentTarget); ok {
			collectAssignmentTargetIdentifiers(inner, into)
		}
	}
}
