package compiler

import (
	"sort"

	"able/interpreter-go/pkg/ast"
)

type nominalOwnershipFlow struct {
	nextID   int
	depth    int
	env      map[string]nominalOwnershipValue
	names    map[int]string
	nominals map[int]string
	origins  map[int]string
	pending  map[string]nominalOwnershipCallClaim
}

func (a *nominalOwnershipAnalyzer) collectCallSites() {
	for _, callable := range a.effects.callables {
		if callable == nil || callable.body == nil {
			continue
		}
		flow := &nominalOwnershipFlow{
			env:      make(map[string]nominalOwnershipValue),
			names:    make(map[int]string),
			nominals: make(map[int]string),
			origins:  make(map[int]string),
			pending:  make(map[string]nominalOwnershipCallClaim),
		}
		for index, param := range callable.params {
			name := functionParameterName(param)
			typeExpr := callableParameterType(callable, index, param)
			info, ok := a.effects.g.structInfoForTypeExpr(callablePackage(callable), typeExpr)
			if name == "" || !ok || info == nil {
				continue
			}
			flow.bindFreshOrigin(name, info.Name, "parameter")
		}
		a.collectCallSitesInNode(callable, callable.body, flow)
	}
}

func (flow *nominalOwnershipFlow) bindFresh(name, nominal string) nominalOwnershipValue {
	return flow.bindFreshOrigin(name, nominal, "local-fresh")
}

func (flow *nominalOwnershipFlow) bindFreshOrigin(name, nominal, origin string) nominalOwnershipValue {
	flow.nextID++
	value := nominalOwnershipValue{source: flow.nextID}
	flow.env[name] = value
	flow.names[value.source] = name
	flow.nominals[value.source] = nominal
	flow.origins[value.source] = origin
	delete(flow.pending, name)
	return value
}

func (flow *nominalOwnershipFlow) clone() *nominalOwnershipFlow {
	return &nominalOwnershipFlow{
		nextID:   flow.nextID,
		depth:    flow.depth,
		env:      cloneOwnershipEnv(flow.env),
		names:    cloneStringMap(flow.names),
		nominals: cloneStringMap(flow.nominals),
		origins:  cloneStringMap(flow.origins),
		pending:  cloneOwnershipClaims(flow.pending),
	}
}

func (flow *nominalOwnershipFlow) nested() *nominalOwnershipFlow {
	result := flow.clone()
	result.depth++
	return result
}

func cloneStringMap(source map[int]string) map[int]string {
	result := make(map[int]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneOwnershipClaims(source map[string]nominalOwnershipCallClaim) map[string]nominalOwnershipCallClaim {
	result := make(map[string]nominalOwnershipCallClaim, len(source))
	for key, value := range source {
		value.path = append([]string(nil), value.path...)
		value.targets = append([]*nominalEffectCallable(nil), value.targets...)
		result[key] = value
	}
	return result
}

func (a *nominalOwnershipAnalyzer) collectCallSitesInNode(
	callable *nominalEffectCallable,
	node ast.Node,
	flow *nominalOwnershipFlow,
) {
	switch current := node.(type) {
	case *ast.BlockExpression:
		if current == nil {
			return
		}
		for _, statement := range current.Body {
			a.collectCallSitesInStatement(callable, statement, flow)
		}
	case ast.Expression:
		a.collectNestedOwnershipBlocks(callable, current, flow)
	}
}

func (a *nominalOwnershipAnalyzer) collectCallSitesInStatement(
	callable *nominalEffectCallable,
	statement ast.Statement,
	flow *nominalOwnershipFlow,
) {
	if assignment, ok := statement.(*ast.AssignmentExpression); ok && assignment != nil {
		a.collectAssignmentCallSite(callable, assignment, flow)
	}
	a.collectNestedOwnershipBlocks(callable, statement, flow)
}

func (a *nominalOwnershipAnalyzer) collectNestedOwnershipBlocks(
	callable *nominalEffectCallable,
	node ast.Node,
	flow *nominalOwnershipFlow,
) {
	switch current := node.(type) {
	case *ast.LoopExpression:
		if current == nil {
			return
		}
		a.collectCallSitesInNode(callable, current.Body, flow.nested())
	case *ast.WhileLoop:
		if current == nil {
			return
		}
		a.collectCallSitesInNode(callable, current.Body, flow.nested())
	case *ast.ForLoop:
		if current == nil {
			return
		}
		a.collectCallSitesInNode(callable, current.Body, flow.nested())
	case *ast.IfExpression:
		if current == nil {
			return
		}
		a.collectCallSitesInNode(callable, current.IfBody, flow.nested())
		for _, clause := range current.ElseIfClauses {
			if clause != nil {
				a.collectCallSitesInNode(callable, clause.Body, flow.nested())
			}
		}
		a.collectCallSitesInNode(callable, current.ElseBody, flow.nested())
	case *ast.MatchExpression:
		if current == nil {
			return
		}
		for _, clause := range current.Clauses {
			if clause != nil {
				a.collectCallSitesInNode(callable, clause.Body, flow.nested())
			}
		}
	}
}

func (a *nominalOwnershipAnalyzer) collectAssignmentCallSite(
	callable *nominalEffectCallable,
	assignment *ast.AssignmentExpression,
	flow *nominalOwnershipFlow,
) {
	left, ok := assignment.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name == "" {
		return
	}
	if call, ok := assignment.Right.(*ast.FunctionCall); ok && call != nil {
		if claim, proven := a.callClaim(callable, call, flow.env); proven {
			if assignment.Operator == ast.AssignmentAssign &&
				len(claim.path) == 0 &&
				ownershipBindingMatches(flow.env[left.Name], claim.source) {
				a.recordCallSite(callable, left.Name, flow, claim, "direct-replacement", nil)
				nominal := flow.nominals[claim.source.source]
				flow.bindFresh(left.Name, nominal)
				return
			}
			if assignment.Operator == ast.AssignmentDeclare {
				claim.depth = flow.depth
				flow.pending[left.Name] = claim
				delete(flow.env, left.Name)
				return
			}
			a.recordCallSite(callable, flow.names[claim.source.source], flow, claim, "not-unconditionally-replaced",
				[]string{"result-not-unconditionally-replacing-source"})
			return
		}
	}
	if base, path, ok := ownershipMemberPath(assignment.Right); ok {
		if claim, exists := flow.pending[base]; exists &&
			assignment.Operator == ast.AssignmentAssign &&
			ownershipPathsEqual(path, claim.path) &&
			ownershipBindingMatches(flow.env[left.Name], claim.source) {
			var blockers []string
			if claim.depth != flow.depth {
				blockers = append(blockers, "conditional-or-nonstraight-replacement")
			}
			a.recordCallSite(callable, left.Name, flow, claim, "embedded-field-replacement", blockers)
			nominal := flow.nominals[claim.source.source]
			flow.bindFresh(left.Name, nominal)
			delete(flow.pending, base)
			return
		}
	}
	if right, ok := assignment.Right.(*ast.Identifier); ok && right != nil {
		if value, exists := flow.env[right.Name]; exists {
			flow.env[left.Name] = value
			delete(flow.pending, left.Name)
			return
		}
	}
	if _, freshLiteral := assignment.Right.(*ast.StructLiteral); freshLiteral {
		if info, ok := a.nominalExpressionInfo(callable, assignment.Right); ok {
			flow.bindFresh(left.Name, info.Name)
			return
		}
	}
	delete(flow.env, left.Name)
	delete(flow.pending, left.Name)
}

func ownershipBindingMatches(left, right nominalOwnershipValue) bool {
	return left.source != 0 && left.source == right.source && len(left.path) == 0 && len(right.path) == 0
}

func ownershipMemberPath(expr ast.Expression) (string, []string, bool) {
	switch current := expr.(type) {
	case *ast.Identifier:
		if current == nil || current.Name == "" {
			return "", nil, false
		}
		return current.Name, nil, true
	case *ast.MemberAccessExpression:
		if current == nil || current.Safe {
			return "", nil, false
		}
		base, path, ok := ownershipMemberPath(current.Object)
		member, memberOK := current.Member.(*ast.Identifier)
		if !ok || !memberOK || member == nil || member.Name == "" {
			return "", nil, false
		}
		return base, append(path, member.Name), true
	default:
		return "", nil, false
	}
}

func (a *nominalOwnershipAnalyzer) nominalExpressionInfo(
	callable *nominalEffectCallable,
	expr ast.Expression,
) (*structInfo, bool) {
	if callable == nil || callable.ctx == nil || expr == nil {
		return nil, false
	}
	typeExpr := a.effects.g.inferredExpressionTypeExpr(callable.ctx, expr)
	if typeExpr == nil {
		if literal, ok := expr.(*ast.StructLiteral); ok && literal != nil && literal.StructType != nil {
			typeExpr = ast.Ty(literal.StructType.Name)
		}
	}
	return a.effects.g.structInfoForTypeExpr(callablePackage(callable), typeExpr)
}

func (a *nominalOwnershipAnalyzer) recordCallSite(
	callable *nominalEffectCallable,
	sourceName string,
	flow *nominalOwnershipFlow,
	claim nominalOwnershipCallClaim,
	replacement string,
	blockers []string,
) {
	if sourceName == "" {
		sourceName = flow.names[claim.source.source]
	}
	if flow.origins[claim.source.source] != "local-fresh" {
		blockers = append(blockers, "source-not-locally-fresh")
	}
	blockers = append(blockers, a.bindingBlockers(callable, sourceName)...)
	blockers = uniqueOwnershipStrings(blockers)
	targets := make([]string, 0, len(claim.targets))
	for _, target := range claim.targets {
		if target != nil {
			targets = append(targets, target.name)
		}
	}
	sort.Strings(targets)
	site := NominalOwnershipCallSite{
		Caller:        callable.name,
		Package:       callablePackage(callable),
		SourceBinding: sourceName,
		Nominal:       flow.nominals[claim.source.source],
		Method:        claim.method,
		Dispatch:      claim.dispatch,
		Targets:       targets,
		ResultPath:    append([]string(nil), claim.path...),
		Replacement:   replacement,
		Eligible:      len(blockers) == 0,
		Blockers:      blockers,
	}
	a.callSites = append(a.callSites, site)
	if site.Eligible && claim.call != nil {
		a.executionSites[claim.call] = &nominalOwnershipExecutionSite{
			call:            claim.call,
			caller:          callable,
			sourceParameter: claim.parameter,
			sourceArgument:  claim.argument,
			path:            append([]string(nil), claim.path...),
			targets:         append([]*nominalEffectCallable(nil), claim.targets...),
			dispatch:        claim.dispatch,
			method:          claim.method,
		}
	}
}

func (a *nominalOwnershipAnalyzer) bindingBlockers(
	callable *nominalEffectCallable,
	name string,
) []string {
	if callable == nil || callable.body == nil || name == "" {
		return []string{"missing-source-binding"}
	}
	blockers := make(map[string]struct{})
	ast.Walk(callable.body, func(node ast.Node) bool {
		if node != callable.body {
			switch node.(type) {
			case *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
				if astNodeReferencesIdentifier(node, name) {
					blockers["captured-by-nested-callable"] = struct{}{}
				}
				return false
			}
		}
		switch current := node.(type) {
		case *ast.AssignmentExpression:
			right, direct := current.Right.(*ast.Identifier)
			if direct && right != nil && right.Name == name {
				if left, ok := current.Left.(*ast.Identifier); ok && left != nil && left.Name != name {
					blockers["retained-alias"] = struct{}{}
				}
				switch current.Left.(type) {
				case *ast.MemberAccessExpression, *ast.IndexExpression:
					blockers["stored-alias"] = struct{}{}
				}
			}
		case *ast.ReturnStatement:
			if ownershipIsIdentifier(current.Argument, name) {
				blockers["returned-alias"] = struct{}{}
			}
		case *ast.StructFieldInitializer:
			if ownershipIsIdentifier(current.Value, name) {
				blockers["stored-in-aggregate"] = struct{}{}
			}
		case *ast.ArrayLiteral:
			for _, element := range current.Elements {
				if ownershipIsIdentifier(element, name) {
					blockers["stored-in-aggregate"] = struct{}{}
				}
			}
		case *ast.FunctionCall:
			for index, argument := range current.Arguments {
				if ownershipIsIdentifier(argument, name) &&
					!a.callArgumentReadOnly(callable, current, index) {
					blockers["captured-or-dynamic-call"] = struct{}{}
				}
			}
		}
		return true
	})
	result := make([]string, 0, len(blockers))
	for blocker := range blockers {
		result = append(result, blocker)
	}
	sort.Strings(result)
	return result
}

func ownershipIsIdentifier(expr ast.Expression, name string) bool {
	identifier, ok := expr.(*ast.Identifier)
	return ok && identifier != nil && identifier.Name == name
}

func (a *nominalOwnershipAnalyzer) callArgumentReadOnly(
	callable *nominalEffectCallable,
	call *ast.FunctionCall,
	argument int,
) bool {
	target, offset, _ := a.effects.resolveCallTarget(callable, call)
	if target != nil {
		index := argument + offset
		return index >= 0 && index < len(target.effects) && target.effects[index] == 0
	}
	targets, interfaceOffset, ok := a.interfaceCallTargets(callable, call)
	if !ok {
		return false
	}
	index := argument + interfaceOffset
	for _, candidate := range targets {
		if candidate == nil || index < 0 || index >= len(candidate.effects) || candidate.effects[index] != 0 {
			return false
		}
	}
	return len(targets) > 0
}

func uniqueOwnershipStrings(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (a *nominalOwnershipAnalyzer) report() *NominalOwnershipReport {
	report := &NominalOwnershipReport{
		SchemaVersion: nominalOwnershipSchemaVersion,
		CallSites:     append([]NominalOwnershipCallSite(nil), a.callSites...),
	}
	report.Totals.Callables = len(a.effects.callables)
	for callable, summaries := range a.byParam {
		for index, value := range summaries {
			if callable == nil || index < 0 || index >= len(callable.params) {
				continue
			}
			typeExpr := callableParameterType(callable, index, callable.params[index])
			info, ok := a.effects.g.structInfoForTypeExpr(callablePackage(callable), typeExpr)
			if !ok || info == nil {
				continue
			}
			resultType := ast.TypeExpression(nil)
			if callable.info != nil {
				resultType = a.effects.g.functionDeclaredOrInferredReturnTypeExpr(callable.info)
			} else if callable.ctx != nil {
				resultType = a.effects.g.inferredExpressionTypeExpr(callable.ctx, callableTailExpression(callable.body))
			}
			summary := NominalOwnershipSuccessorSummary{
				Callable:       callable.name,
				Package:        callablePackage(callable),
				Kind:           callable.kind,
				ParameterIndex: index,
				ParameterName:  functionParameterName(callable.params[index]),
				Nominal:        info.Name,
				ResultType:     typeExpressionToString(resultType),
				ResultPath:     append([]string(nil), value.path...),
				Fresh:          true,
				ReadOnly:       callable.effects[index] == 0,
			}
			if callable.info != nil {
				summary.GeneratedGoName = callable.info.GoName
			}
			report.Successors = append(report.Successors, summary)
		}
	}
	sort.Slice(report.Successors, func(i, j int) bool {
		if report.Successors[i].Package != report.Successors[j].Package {
			return report.Successors[i].Package < report.Successors[j].Package
		}
		if report.Successors[i].Callable != report.Successors[j].Callable {
			return report.Successors[i].Callable < report.Successors[j].Callable
		}
		return report.Successors[i].ParameterIndex < report.Successors[j].ParameterIndex
	})
	sort.Slice(report.CallSites, func(i, j int) bool {
		if report.CallSites[i].Package != report.CallSites[j].Package {
			return report.CallSites[i].Package < report.CallSites[j].Package
		}
		if report.CallSites[i].Caller != report.CallSites[j].Caller {
			return report.CallSites[i].Caller < report.CallSites[j].Caller
		}
		if report.CallSites[i].SourceBinding != report.CallSites[j].SourceBinding {
			return report.CallSites[i].SourceBinding < report.CallSites[j].SourceBinding
		}
		return report.CallSites[i].Method < report.CallSites[j].Method
	})
	report.Totals.FreshSuccessors = len(report.Successors)
	report.Totals.CandidateCallSites = len(report.CallSites)
	for _, site := range report.CallSites {
		if !site.Eligible {
			continue
		}
		report.Totals.EligibleTransfers++
		switch site.Replacement {
		case "direct-replacement":
			report.Totals.DirectTransfers++
		case "embedded-field-replacement":
			report.Totals.EmbeddedFieldTransfers++
		}
		if site.Dispatch == "native-interface" {
			report.Totals.InterfaceTransfers++
		}
	}
	return report
}
