package typechecker

import (
	"fmt"
	"math/big"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkExpression(env *Environment, expr ast.Expression) ([]Diagnostic, Type) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		suffix := "i32"
		explicit := false
		if e.IntegerType != nil {
			suffix = string(*e.IntegerType)
			explicit = true
		}
		literal := new(big.Int)
		if e.Value != nil {
			literal = new(big.Int).Set(e.Value)
		}
		typ := IntegerType{Suffix: suffix, Literal: literal, Explicit: explicit}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.FloatLiteral:
		suffix := "f64"
		if e.FloatType != nil {
			suffix = string(*e.FloatType)
		}
		typ := FloatType{Suffix: suffix}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.BooleanLiteral:
		typ := PrimitiveType{Kind: PrimitiveBool}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.NilLiteral:
		typ := PrimitiveType{Kind: PrimitiveNil}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.StringLiteral:
		typ := PrimitiveType{Kind: PrimitiveString}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.CharLiteral:
		typ := PrimitiveType{Kind: PrimitiveChar}
		c.infer.set(e, typ)
		return nil, typ
	case *ast.IteratorLiteral:
		diags, iteratorType := c.checkIteratorLiteral(env, e)
		c.infer.set(e, iteratorType)
		return diags, iteratorType
	case *ast.LoopExpression:
		diags, loopType := c.checkLoopExpression(env, e)
		c.infer.set(e, loopType)
		return diags, loopType
	case *ast.ImplicitMemberExpression:
		// Placeholder-based member access (e.g., within pipe shorthand). Without
		// full context we treat it as unknown.
		c.infer.set(e, UnknownType{})
		return nil, UnknownType{}
	case *ast.PlaceholderExpression:
		c.infer.set(e, UnknownType{})
		return nil, UnknownType{}
	case *ast.TopicReferenceExpression:
		c.infer.set(e, UnknownType{})
		return nil, UnknownType{}
	case *ast.ArrayLiteral:
		return c.checkArrayLiteral(env, e)
	case *ast.MapLiteral:
		var diags []Diagnostic
		var keyType Type = UnknownType{}
		var valueType Type = UnknownType{}
		for _, element := range e.Elements {
			switch entry := element.(type) {
			case *ast.MapLiteralEntry:
				keyDiags, inferredKey := c.checkExpression(env, entry.Key)
				diags = append(diags, keyDiags...)
				valueDiags, inferredValue := c.checkExpression(env, entry.Value)
				diags = append(diags, valueDiags...)
				var mergeDiags []Diagnostic
				keyType, mergeDiags = mergeMapComponentType(keyType, inferredKey, "map key", entry.Key)
				diags = append(diags, mergeDiags...)
				valueType, mergeDiags = mergeMapComponentType(valueType, inferredValue, "map value", entry.Value)
				diags = append(diags, mergeDiags...)
			case *ast.MapLiteralSpread:
				spreadDiags, spreadType := c.checkExpression(env, entry.Expression)
				diags = append(diags, spreadDiags...)
				if mapType, ok := spreadType.(MapType); ok {
					var mergeDiags []Diagnostic
					keyType, mergeDiags = mergeMapComponentType(keyType, mapType.Key, "map key", entry.Expression)
					diags = append(diags, mergeDiags...)
					valueType, mergeDiags = mergeMapComponentType(valueType, mapType.Value, "map value", entry.Expression)
					diags = append(diags, mergeDiags...)
				} else if !isUnknownType(spreadType) {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: map spread expects Map, got %s", spreadType.Name()),
						Node:    entry.Expression,
					})
				}
			default:
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: unsupported map literal element %T", element),
					Node:    e,
				})
			}
		}
		resultType := MapType{Key: keyType, Value: valueType}
		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.BlockExpression:
		blockEnv := env.Extend()
		var (
			diags      []Diagnostic
			resultType Type = UnknownType{}
		)
		for idx, stmt := range e.Body {
			switch s := stmt.(type) {
			case *ast.AssignmentExpression:
				assignDiags := c.checkStatement(blockEnv, s)
				diags = append(diags, assignDiags...)
				if idx == len(e.Body)-1 {
					if inferred, ok := c.infer.get(s.Right); ok {
						resultType = inferred
					}
				}
			case ast.Expression:
				exprDiags, exprType := c.checkExpression(blockEnv, s)
				diags = append(diags, exprDiags...)
				if idx == len(e.Body)-1 {
					resultType = exprType
				}
			default:
				diags = append(diags, c.checkStatement(blockEnv, s)...)
			}
		}
		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.IfExpression:
		var diags []Diagnostic
		condDiags, condType := c.checkExpression(env, e.IfCondition)
		diags = append(diags, condDiags...)
		if !typeAssignable(condType, PrimitiveType{Kind: PrimitiveBool}) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: if condition must be bool",
				Node:    e.IfCondition,
			})
		}

		branchTypes := make([]Type, 0, 1+len(e.OrClauses))
		if e.IfBody != nil {
			bodyDiags, bodyType := c.checkExpression(env, e.IfBody)
			diags = append(diags, bodyDiags...)
			branchTypes = append(branchTypes, bodyType)
		} else {
			branchTypes = append(branchTypes, UnknownType{})
		}

		for _, clause := range e.OrClauses {
			if clause == nil {
				continue
			}
			if clause.Condition != nil {
				orCondDiags, orCondType := c.checkExpression(env, clause.Condition)
				diags = append(diags, orCondDiags...)
				if !typeAssignable(orCondType, PrimitiveType{Kind: PrimitiveBool}) {
					diags = append(diags, Diagnostic{
						Message: "typechecker: if-or condition must be bool",
						Node:    clause.Condition,
					})
				}
			}
			if clause.Body != nil {
				bodyDiags, bodyType := c.checkExpression(env, clause.Body)
				diags = append(diags, bodyDiags...)
				branchTypes = append(branchTypes, bodyType)
			} else {
				branchTypes = append(branchTypes, UnknownType{})
			}
		}

		resultType := mergeBranchTypes(branchTypes)
		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.UnaryExpression:
		return c.checkUnaryExpression(env, e)
	case *ast.BinaryExpression:
		return c.checkBinaryExpression(env, e)
	case *ast.LambdaExpression:
		return c.checkLambdaExpression(env, e)
	case *ast.FunctionCall:
		var diags []Diagnostic
		var builtinName string
		if ident, ok := e.Callee.(*ast.Identifier); ok && ident != nil {
			builtinName = ident.Name
		}
		inPipeline := c.inPipeContext() || hasPipelinePlaceholder(e.Arguments)
		calleeDiags, calleeType := c.checkExpression(env, e.Callee)
		diags = append(diags, calleeDiags...)
		if inPipeline {
			calleeType = UnknownType{}
		}

		argTypes := make([]Type, len(e.Arguments))
		for i, arg := range e.Arguments {
			argDiags, argType := c.checkExpression(env, arg)
			diags = append(diags, argDiags...)
			argTypes[i] = argType
		}

		resultType := Type(UnknownType{})
		if fnType, ok := calleeType.(FunctionType); ok {
			if len(e.TypeArguments) > 0 && len(fnType.TypeParams) != len(e.TypeArguments) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: function expects %d type arguments, got %d", len(fnType.TypeParams), len(e.TypeArguments)),
					Node:    e,
				})
			}
			instantiated, instDiags := c.instantiateFunctionCall(fnType, e, argTypes)
			diags = append(diags, instDiags...)
			if len(instantiated.Obligations) > 0 {
				c.obligations = append(c.obligations, instantiated.Obligations...)
			}
			if len(instantiated.Params) != len(argTypes) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: function expects %d arguments, got %d", len(instantiated.Params), len(argTypes)),
					Node:    e,
				})
			} else {
				for i, expected := range instantiated.Params {
					if !typeAssignable(argTypes[i], expected) {
						if msg, ok := literalMismatchMessage(argTypes[i], expected); ok {
							diags = append(diags, Diagnostic{
								Message: fmt.Sprintf("typechecker: %s", msg),
								Node:    e.Arguments[i],
							})
						} else {
							diags = append(diags, Diagnostic{
								Message: fmt.Sprintf("typechecker: argument %d has type %s, expected %s", i+1, typeName(argTypes[i]), typeName(expected)),
								Node:    e.Arguments[i],
							})
						}
					}
				}
			}
			resultType = instantiated.Return
		} else if inPipeline {
			resultType = UnknownType{}
		} else if !isUnknownType(calleeType) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: cannot call non-function value",
				Node:    e.Callee,
			})
		}

		diags = append(diags, c.checkBuiltinCallContext(builtinName, e)...)

		c.infer.set(e, resultType)
		return diags, resultType
	case *ast.AssignmentExpression:
		diags := c.checkStatement(env, e)
		c.infer.set(e, UnknownType{})
		return diags, UnknownType{}
	case *ast.MemberAccessExpression:
		return c.checkMemberAccess(env, e)
	case *ast.IndexExpression:
		return c.checkIndexExpression(env, e)
	case *ast.StructLiteral:
		return c.checkStructLiteral(env, e)
	case *ast.MatchExpression:
		return c.checkMatchExpression(env, e)
	case *ast.RangeExpression:
		return c.checkRangeExpression(env, e)
	case *ast.RescueExpression:
		return c.checkRescueExpression(env, e)
	case *ast.OrElseExpression:
		return c.checkOrElseExpression(env, e)
	case *ast.EnsureExpression:
		return c.checkEnsureExpression(env, e)
	case *ast.BreakpointExpression:
		return c.checkBreakpointExpression(env, e)
	case *ast.StringInterpolation:
		return c.checkStringInterpolation(env, e)
	case *ast.ProcExpression:
		return c.checkProcExpression(env, e)
	case *ast.SpawnExpression:
		return c.checkSpawnExpression(env, e)
	case *ast.AwaitExpression:
		c.infer.set(e, UnknownType{})
		return nil, UnknownType{}
	case *ast.PropagationExpression:
		return c.checkPropagationExpression(env, e)
	case *ast.Identifier:
		if typ, ok := env.Lookup(e.Name); ok {
			c.infer.set(e, typ)
			return nil, typ
		}
		if c.allowDynamicLookups {
			c.infer.set(e, UnknownType{})
			return nil, UnknownType{}
		}
		diag := Diagnostic{Message: fmt.Sprintf("typechecker: undefined identifier '%s'", e.Name), Node: expr}
		return []Diagnostic{diag}, UnknownType{}
	default:
		diag := Diagnostic{Message: fmt.Sprintf("typechecker: unsupported expression %T", expr), Node: expr}
		return []Diagnostic{diag}, UnknownType{}
	}
}

func hasPipelinePlaceholder(args []ast.Expression) bool {
	for _, arg := range args {
		switch arg.(type) {
		case *ast.TopicReferenceExpression, *ast.PlaceholderExpression:
			return true
		}
	}
	return false
}

func (c *Checker) checkStatement(env *Environment, stmt ast.Statement) []Diagnostic {
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
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
		rhsDiags, typ := c.checkExpression(env, s.Right)
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

func (c *Checker) checkIteratorLiteral(env *Environment, lit *ast.IteratorLiteral) ([]Diagnostic, Type) {
	if lit == nil {
		return nil, IteratorType{Element: UnknownType{}}
	}
	var expected Type = UnknownType{}
	if lit.ElementType != nil {
		expected = c.resolveTypeReference(lit.ElementType)
		if expected == nil {
			expected = UnknownType{}
		}
	}
	bodyEnv := env
	if bodyEnv == nil {
		bodyEnv = c.global.Extend()
	} else {
		bodyEnv = env.Extend()
	}
	if lit.Binding != nil && lit.Binding.Name != "" {
		bodyEnv.Define(lit.Binding.Name, UnknownType{})
	}
	bodyEnv.Define("gen", UnknownType{})
	var inferred Type = UnknownType{}
	var diags []Diagnostic
	for _, stmt := range lit.Body {
		if stmt == nil {
			continue
		}
		if yieldStmt, ok := stmt.(*ast.YieldStatement); ok {
			yieldDiags, yieldType := c.checkIteratorYield(bodyEnv, yieldStmt, expected)
			diags = append(diags, yieldDiags...)
			inferred = mergeIteratorElementType(inferred, yieldType)
			continue
		}
		diags = append(diags, c.checkStatement(bodyEnv, stmt)...)
	}
	elementType := expected
	if elementType == nil || isUnknownType(elementType) {
		elementType = inferred
	}
	if elementType == nil {
		elementType = UnknownType{}
	}
	return diags, IteratorType{Element: elementType}
}

func (c *Checker) checkIteratorYield(env *Environment, stmt *ast.YieldStatement, expected Type) ([]Diagnostic, Type) {
	var diags []Diagnostic
	valueType := Type(PrimitiveType{Kind: PrimitiveNil})
	if stmt.Expression != nil {
		exprDiags, exprType := c.checkExpression(env, stmt.Expression)
		diags = append(diags, exprDiags...)
		if exprType != nil {
			valueType = exprType
		} else {
			valueType = UnknownType{}
		}
	}
	if expected == nil || isUnknownType(expected) {
		return diags, valueType
	}
	if typeAssignable(valueType, expected) {
		return diags, valueType
	}
	actual := typeName(valueType)
	if actual == "" {
		actual = "unknown"
	}
	message := fmt.Sprintf(
		"typechecker: iterator annotation expects elements of type %s, got %s",
		typeName(expected),
		actual,
	)
	if msg, ok := literalMismatchMessage(valueType, expected); ok {
		message = fmt.Sprintf("typechecker: %s", msg)
	}
	diags = append(diags, Diagnostic{
		Message: message,
		Node:    stmt,
	})
	return diags, valueType
}

func mergeIteratorElementType(current Type, next Type) Type {
	if next == nil || isUnknownType(next) {
		return current
	}
	if current == nil || isUnknownType(current) {
		return next
	}
	if typeAssignable(next, current) {
		return current
	}
	if typeAssignable(current, next) {
		return next
	}
	return UnknownType{}
}

func mergeMapComponentType(current Type, candidate Type, label string, node ast.Node) (Type, []Diagnostic) {
	if current == nil || isUnknownType(current) {
		return candidate, nil
	}
	if candidate == nil || isUnknownType(candidate) {
		return current, nil
	}
	if typeAssignable(candidate, current) {
		return current, nil
	}
	if typeAssignable(current, candidate) {
		return candidate, nil
	}
	diag := Diagnostic{
		Message: fmt.Sprintf("typechecker: %s expects type %s, got %s", label, current.Name(), candidate.Name()),
		Node:    node,
	}
	return current, []Diagnostic{diag}
}
