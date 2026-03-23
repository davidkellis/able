package compiler

import "able/interpreter-go/pkg/ast"

type paramFactObservation struct {
	Seen  bool
	Valid bool
	Fact  integerFact
}

func (g *generator) resolveStaticFunctionIntegerFacts() {
	if g == nil {
		return
	}
	for _, info := range g.allFunctionInfos() {
		if info != nil {
			info.ParamFacts = nil
		}
	}
	for {
		observations := make(map[*functionInfo]map[string]paramFactObservation)
		for _, info := range g.sortedFunctionInfos() {
			if info == nil || !info.Compileable || info.Definition == nil || info.Definition.Body == nil {
				continue
			}
			ctx := newCompileContext(g, info, g.functionsForCompileContext(info), g.overloadsForPackage(info.Package), info.Package, g.compileContextGenericNames(info))
			if implInfo, ok := g.implMethodByInfo[info]; ok && implInfo != nil && implInfo.IsDefault {
				ctx.implSiblings = g.implSiblingsForFunction(info)
			}
			g.collectIntegerFactsFromStatements(ctx, info.Definition.Body.Body, observations)
		}
		if !g.applyStaticFunctionIntegerFacts(observations) {
			return
		}
	}
}

func (g *generator) applyStaticFunctionIntegerFacts(observations map[*functionInfo]map[string]paramFactObservation) bool {
	changed := false
	for _, info := range g.allFunctionInfos() {
		if info == nil {
			continue
		}
		next := make(map[string]integerFact)
		for _, param := range info.Params {
			if param.Name == "" {
				continue
			}
			paramObs, ok := observations[info][param.Name]
			if !ok || !paramObs.Seen || !paramObs.Valid || !paramObs.Fact.hasUsefulFact() {
				continue
			}
			next[param.Name] = paramObs.Fact
		}
		if !integerFactMapEqual(info.ParamFacts, next) {
			changed = true
			if len(next) == 0 {
				info.ParamFacts = nil
			} else {
				info.ParamFacts = next
			}
		}
	}
	return changed
}

func integerFactMapEqual(left map[string]integerFact, right map[string]integerFact) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || leftValue != rightValue {
			return false
		}
	}
	return true
}

func (g *generator) collectIntegerFactsFromStatements(ctx *compileContext, statements []ast.Statement, observations map[*functionInfo]map[string]paramFactObservation) {
	if g == nil || ctx == nil {
		return
	}
	for _, stmt := range statements {
		g.collectIntegerFactsFromStatement(ctx, stmt, observations)
	}
}

func (g *generator) collectIntegerFactsFromStatement(ctx *compileContext, stmt ast.Statement, observations map[*functionInfo]map[string]paramFactObservation) {
	if g == nil || ctx == nil || stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.AssignmentExpression:
		g.collectIntegerFactsFromAssignment(ctx, s, observations)
	case *ast.BlockExpression:
		g.collectIntegerFactsFromStatements(ctx.child(), s.Body, observations)
	case *ast.IfExpression:
		g.collectIntegerFactsFromExpr(ctx, s.IfCondition, observations)
		g.collectIntegerFactsFromStatements(ctx.child(), s.IfBody.Body, observations)
		for _, clause := range s.ElseIfClauses {
			if clause == nil {
				continue
			}
			g.collectIntegerFactsFromExpr(ctx, clause.Condition, observations)
			g.collectIntegerFactsFromStatements(ctx.child(), clause.Body.Body, observations)
		}
		if s.ElseBody != nil {
			g.collectIntegerFactsFromStatements(ctx.child(), s.ElseBody.Body, observations)
		}
	case *ast.LoopExpression:
		loopCtx := ctx.child()
		if varName, boundExpr, ok := g.matchCountedLoopGuardFromStatements(s.Body.Body); ok {
			if binding, found := loopCtx.lookup(varName); found {
				g.seedCountedLoopIntegerFact(loopCtx, binding, boundExpr)
			}
		}
		g.collectIntegerFactsFromStatements(loopCtx, s.Body.Body, observations)
	case *ast.WhileLoop:
		g.collectIntegerFactsFromExpr(ctx, s.Condition, observations)
		g.collectIntegerFactsFromStatements(ctx.child(), s.Body.Body, observations)
	case *ast.ForLoop:
		g.collectIntegerFactsFromExpr(ctx, s.Iterable, observations)
		g.collectIntegerFactsFromStatements(ctx.child(), s.Body.Body, observations)
	case *ast.ReturnStatement:
		g.collectIntegerFactsFromExpr(ctx, s.Argument, observations)
	case *ast.RaiseStatement:
		g.collectIntegerFactsFromExpr(ctx, s.Expression, observations)
	case *ast.YieldStatement:
		g.collectIntegerFactsFromExpr(ctx, s.Expression, observations)
	case *ast.FunctionDefinition:
		// Local function bodies do not affect outer callsite facts.
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			g.collectIntegerFactsFromExpr(ctx, expr, observations)
		}
	}
}

func (g *generator) collectIntegerFactsFromAssignment(ctx *compileContext, assign *ast.AssignmentExpression, observations map[*functionInfo]map[string]paramFactObservation) {
	if g == nil || ctx == nil || assign == nil {
		return
	}
	g.collectIntegerFactsFromExpr(ctx, assign.Right, observations)
	switch left := assign.Left.(type) {
	case *ast.MemberAccessExpression:
		g.collectIntegerFactsFromExpr(ctx, left.Object, observations)
		g.collectIntegerFactsFromExpr(ctx, left.Member, observations)
	case *ast.IndexExpression:
		g.collectIntegerFactsFromExpr(ctx, left.Object, observations)
		g.collectIntegerFactsFromExpr(ctx, left.Index, observations)
	}
	if assign.Operator != ast.AssignmentDeclare && assign.Operator != ast.AssignmentAssign {
		if name, _, ok := g.assignmentTargetName(assign.Left); ok && name != "" {
			if binding, ok := ctx.lookup(name); ok {
				ctx.clearIntegerFact(binding.GoName)
			}
		}
		return
	}
	name, typeAnnotation, ok := g.assignmentTargetName(assign.Left)
	if !ok || name == "" {
		return
	}
	existing, exists := ctx.lookup(name)
	goType := ""
	if typeAnnotation != nil {
		if mapped, ok := g.lowerCarrierType(ctx, typeAnnotation); ok {
			goType = mapped
		}
	} else if exists {
		goType = existing.GoType
	} else if inferred, ok := g.inferAssignmentIntegerGoType(ctx, assign.Right); ok {
		goType = inferred
	}
	goName := existing.GoName
	if assign.Operator == ast.AssignmentDeclare || !exists {
		goName = sanitizeIdent(name)
		ctx.setLocalBinding(name, paramInfo{Name: name, GoName: goName, GoType: goType, TypeExpr: typeAnnotation})
	}
	if goType == "" {
		ctx.clearIntegerFact(goName)
		return
	}
	g.refreshIntegerFactForBinding(ctx, paramInfo{Name: name, GoName: goName, GoType: goType}, assign.Right)
}

func (g *generator) inferAssignmentIntegerGoType(ctx *compileContext, expr ast.Expression) (string, bool) {
	if g == nil || expr == nil {
		return "", false
	}
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return g.inferIntegerLiteralType(e), true
	case *ast.Identifier:
		if ctx == nil || e == nil || e.Name == "" {
			return "", false
		}
		binding, ok := ctx.lookup(e.Name)
		return binding.GoType, ok
	case *ast.TypeCastExpression:
		if e == nil || e.TargetType == nil {
			return "", false
		}
		return g.lowerCarrierType(ctx, e.TargetType)
	default:
		return "", false
	}
}

func (g *generator) collectIntegerFactsFromExpr(ctx *compileContext, expr ast.Expression, observations map[*functionInfo]map[string]paramFactObservation) {
	if g == nil || ctx == nil || expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.FunctionCall:
		for _, arg := range e.Arguments {
			g.collectIntegerFactsFromExpr(ctx, arg, observations)
		}
		g.collectIntegerFactsFromExpr(ctx, e.Callee, observations)
		g.observeStaticFunctionCallIntegerFacts(ctx, e, observations)
	case *ast.BinaryExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Left, observations)
		g.collectIntegerFactsFromExpr(ctx, e.Right, observations)
	case *ast.UnaryExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Operand, observations)
	case *ast.TypeCastExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Expression, observations)
	case *ast.MemberAccessExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Object, observations)
		g.collectIntegerFactsFromExpr(ctx, e.Member, observations)
	case *ast.IndexExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Object, observations)
		g.collectIntegerFactsFromExpr(ctx, e.Index, observations)
	case *ast.BlockExpression:
		g.collectIntegerFactsFromStatements(ctx.child(), e.Body, observations)
	case *ast.IfExpression:
		g.collectIntegerFactsFromStatement(ctx, e, observations)
	case *ast.MatchExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Subject, observations)
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			g.collectIntegerFactsFromExpr(ctx, clause.Guard, observations)
			g.collectIntegerFactsFromExpr(ctx.child(), clause.Body, observations)
		}
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil {
				g.collectIntegerFactsFromExpr(ctx, field.Value, observations)
			}
		}
	case *ast.ArrayLiteral:
		for _, element := range e.Elements {
			g.collectIntegerFactsFromExpr(ctx, element, observations)
		}
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			g.collectIntegerFactsFromExpr(ctx, part, observations)
		}
	case *ast.RangeExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Start, observations)
		g.collectIntegerFactsFromExpr(ctx, e.End, observations)
	case *ast.LambdaExpression:
		g.collectIntegerFactsFromExpr(ctx.child(), e.Body, observations)
	case *ast.IteratorLiteral:
		g.collectIntegerFactsFromStatements(ctx.child(), e.Body, observations)
	case *ast.PropagationExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Expression, observations)
	case *ast.OrElseExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Expression, observations)
		if e.Handler != nil {
			g.collectIntegerFactsFromStatements(ctx.child(), e.Handler.Body, observations)
		}
	case *ast.AwaitExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Expression, observations)
	case *ast.SpawnExpression:
		g.collectIntegerFactsFromExpr(ctx, e.Expression, observations)
	case *ast.RescueExpression:
		g.collectIntegerFactsFromExpr(ctx, e.MonitoredExpression, observations)
		for _, clause := range e.Clauses {
			if clause != nil {
				g.collectIntegerFactsFromExpr(ctx.child(), clause.Body, observations)
			}
		}
	case *ast.EnsureExpression:
		g.collectIntegerFactsFromExpr(ctx, e.TryExpression, observations)
		if e.EnsureBlock != nil {
			g.collectIntegerFactsFromStatements(ctx.child(), e.EnsureBlock.Body, observations)
		}
	}
}

func (g *generator) observeStaticFunctionCallIntegerFacts(ctx *compileContext, call *ast.FunctionCall, observations map[*functionInfo]map[string]paramFactObservation) {
	if g == nil || ctx == nil || call == nil || len(call.TypeArguments) != 0 {
		return
	}
	callee, ok := call.Callee.(*ast.Identifier)
	if !ok || callee == nil || callee.Name == "" {
		return
	}
	info, overload, ok := g.resolveStaticCallable(ctx, callee.Name)
	if !ok || overload != nil || info == nil || !info.Compileable {
		return
	}
	if observations[info] == nil {
		observations[info] = make(map[string]paramFactObservation)
	}
	for idx, param := range info.Params {
		if idx >= len(call.Arguments) {
			break
		}
		argFact, ok := g.exprIntegerFact(ctx, call.Arguments[idx])
		current := observations[info][param.Name]
		observations[info][param.Name] = mergeObservedParamFact(current, argFact, ok)
	}
}

func mergeObservedParamFact(current paramFactObservation, fact integerFact, ok bool) paramFactObservation {
	if !current.Seen {
		current.Seen = true
		if !ok || !fact.hasUsefulFact() {
			current.Valid = false
			return current
		}
		current.Valid = true
		current.Fact = fact
		return current
	}
	if !current.Valid {
		return current
	}
	if !ok || !fact.hasUsefulFact() {
		current.Valid = false
		current.Fact = integerFact{}
		return current
	}
	current.Fact = intersectIntegerFacts(current.Fact, fact)
	if !current.Fact.hasUsefulFact() {
		current.Valid = false
	}
	return current
}

func intersectIntegerFacts(left integerFact, right integerFact) integerFact {
	merged := integerFact{
		NonNegative: left.NonNegative && right.NonNegative,
	}
	if left.HasMax && right.HasMax {
		merged.HasMax = true
		if left.MaxInclusive >= right.MaxInclusive {
			merged.MaxInclusive = left.MaxInclusive
		} else {
			merged.MaxInclusive = right.MaxInclusive
		}
	}
	return merged
}
