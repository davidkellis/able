package compiler

import (
	"sort"
	"strconv"

	"able/interpreter-go/pkg/ast"
)

const nominalOwnershipSchemaVersion = 1

// NominalOwnershipReport is an opt-in proof report. Eligible sites have a
// complete typed callee set, a fresh successor result, and an unconditional
// caller replacement with no syntactically retained alias. The report never
// changes carrier selection or generated execution.
type NominalOwnershipReport struct {
	SchemaVersion int                                `json:"schema_version"`
	Successors    []NominalOwnershipSuccessorSummary `json:"successors"`
	CallSites     []NominalOwnershipCallSite         `json:"call_sites"`
	Totals        NominalOwnershipTotals             `json:"totals"`
}

type NominalOwnershipTotals struct {
	Callables              int `json:"callables"`
	FreshSuccessors        int `json:"fresh_successors"`
	CandidateCallSites     int `json:"candidate_call_sites"`
	EligibleTransfers      int `json:"eligible_transfers"`
	DirectTransfers        int `json:"direct_transfers"`
	EmbeddedFieldTransfers int `json:"embedded_field_transfers"`
	InterfaceTransfers     int `json:"interface_transfers"`
}

type NominalOwnershipSuccessorSummary struct {
	Callable        string   `json:"callable"`
	Package         string   `json:"package"`
	Kind            string   `json:"kind"`
	GeneratedGoName string   `json:"generated_go_name,omitempty"`
	ParameterIndex  int      `json:"parameter_index"`
	ParameterName   string   `json:"parameter_name"`
	Nominal         string   `json:"nominal"`
	ResultType      string   `json:"result_type"`
	ResultPath      []string `json:"result_path,omitempty"`
	Fresh           bool     `json:"fresh"`
	ReadOnly        bool     `json:"read_only_non_escaping"`
}

type NominalOwnershipCallSite struct {
	Caller        string   `json:"caller"`
	Package       string   `json:"package"`
	SourceBinding string   `json:"source_binding"`
	Nominal       string   `json:"nominal"`
	Method        string   `json:"method,omitempty"`
	Dispatch      string   `json:"dispatch"`
	Targets       []string `json:"targets"`
	ResultPath    []string `json:"result_path,omitempty"`
	Replacement   string   `json:"replacement"`
	Eligible      bool     `json:"eligible"`
	Blockers      []string `json:"blockers,omitempty"`
}

type nominalOwnershipValue struct {
	source int
	path   []string
}

type nominalOwnershipAnalyzer struct {
	effects        *nominalEffectAnalyzer
	byParam        map[*nominalEffectCallable]map[int]nominalOwnershipValue
	callSites      []NominalOwnershipCallSite
	executionSites map[*ast.FunctionCall]*nominalOwnershipExecutionSite
}

func (g *generator) nominalOwnershipAnalyzer() *nominalOwnershipAnalyzer {
	effects := g.resolvedNominalEffectAnalyzer()
	if effects == nil {
		return nil
	}
	analyzer := &nominalOwnershipAnalyzer{
		effects:        effects,
		byParam:        make(map[*nominalEffectCallable]map[int]nominalOwnershipValue),
		executionSites: make(map[*ast.FunctionCall]*nominalOwnershipExecutionSite),
	}
	analyzer.closeSuccessorFixedPoint()
	analyzer.collectCallSites()
	return analyzer
}

func (g *generator) resolveNominalOwnership() *NominalOwnershipReport {
	analyzer := g.nominalOwnershipAnalyzer()
	if analyzer == nil {
		return nil
	}
	return analyzer.report()
}

func (a *nominalOwnershipAnalyzer) closeSuccessorFixedPoint() {
	for changed := true; changed; {
		changed = false
		for _, callable := range a.effects.callables {
			values := a.successorsForCallable(callable)
			for source, value := range values {
				if source < 0 || source >= len(callable.effects) || callable.effects[source] != 0 {
					continue
				}
				current := a.byParam[callable]
				if current == nil {
					current = make(map[int]nominalOwnershipValue)
					a.byParam[callable] = current
				}
				if existing, ok := current[source]; ok && ownershipPathsEqual(existing.path, value.path) {
					continue
				}
				current[source] = value
				changed = true
			}
		}
	}
}

func (a *nominalOwnershipAnalyzer) successorsForCallable(callable *nominalEffectCallable) map[int]nominalOwnershipValue {
	if callable == nil || callable.body == nil {
		return nil
	}
	if ownershipBodyHasExplicitReturn(callable.body) {
		return nil
	}
	env := make(map[string]nominalOwnershipValue, len(callable.params))
	for index, param := range callable.params {
		if name := functionParameterName(param); name != "" {
			env[name] = nominalOwnershipValue{source: index}
		}
	}
	value, ok := a.successorForNode(callable, callable.body, env)
	if !ok {
		return nil
	}
	return map[int]nominalOwnershipValue{value.source: value}
}

func ownershipBodyHasExplicitReturn(body ast.Node) bool {
	found := false
	ast.Walk(body, func(node ast.Node) bool {
		if found {
			return false
		}
		if node != body {
			switch node.(type) {
			case *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
				return false
			}
		}
		if _, ok := node.(*ast.ReturnStatement); ok {
			found = true
			return false
		}
		return true
	})
	return found
}

func (a *nominalOwnershipAnalyzer) successorForNode(
	callable *nominalEffectCallable,
	node ast.Node,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipValue, bool) {
	switch current := node.(type) {
	case *ast.BlockExpression:
		return a.successorForBlock(callable, current, env)
	case ast.Expression:
		return a.successorForExpression(callable, current, env)
	default:
		return nominalOwnershipValue{}, false
	}
}

func (a *nominalOwnershipAnalyzer) successorForBlock(
	callable *nominalEffectCallable,
	block *ast.BlockExpression,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipValue, bool) {
	if block == nil || len(block.Body) == 0 {
		return nominalOwnershipValue{}, false
	}
	for index, statement := range block.Body {
		if assignment, ok := statement.(*ast.AssignmentExpression); ok && assignment != nil {
			if identifier, ok := assignment.Left.(*ast.Identifier); ok && identifier != nil {
				value, proven := a.successorForExpression(callable, assignment.Right, env)
				if proven {
					env[identifier.Name] = value
				} else {
					delete(env, identifier.Name)
				}
			}
		}
		if index == len(block.Body)-1 {
			expression, ok := statement.(ast.Expression)
			if !ok {
				return nominalOwnershipValue{}, false
			}
			return a.successorForExpression(callable, expression, env)
		}
	}
	return nominalOwnershipValue{}, false
}

func (a *nominalOwnershipAnalyzer) successorForExpression(
	callable *nominalEffectCallable,
	expr ast.Expression,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipValue, bool) {
	switch current := expr.(type) {
	case *ast.Identifier:
		value, ok := env[current.Name]
		return value, ok
	case *ast.BlockExpression:
		return a.successorForBlock(callable, current, cloneOwnershipEnv(env))
	case *ast.IfExpression:
		return a.successorForIf(callable, current, env)
	case *ast.FunctionCall:
		claim, ok := a.callClaim(callable, current, env)
		if !ok {
			return nominalOwnershipValue{}, false
		}
		return nominalOwnershipValue{source: claim.source.source, path: append([]string(nil), claim.path...)}, true
	case *ast.StructLiteral:
		return a.successorForStruct(callable, current, env)
	default:
		return nominalOwnershipValue{}, false
	}
}

func (a *nominalOwnershipAnalyzer) successorForIf(
	callable *nominalEffectCallable,
	expr *ast.IfExpression,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipValue, bool) {
	if expr == nil || expr.ElseBody == nil {
		return nominalOwnershipValue{}, false
	}
	first, ok := a.successorForBlock(callable, expr.IfBody, cloneOwnershipEnv(env))
	if !ok {
		return nominalOwnershipValue{}, false
	}
	for _, clause := range expr.ElseIfClauses {
		if clause == nil {
			return nominalOwnershipValue{}, false
		}
		value, proven := a.successorForBlock(callable, clause.Body, cloneOwnershipEnv(env))
		if !proven || !ownershipValuesEqual(first, value) {
			return nominalOwnershipValue{}, false
		}
	}
	last, ok := a.successorForBlock(callable, expr.ElseBody, cloneOwnershipEnv(env))
	if !ok || !ownershipValuesEqual(first, last) {
		return nominalOwnershipValue{}, false
	}
	return first, true
}

func (a *nominalOwnershipAnalyzer) successorForStruct(
	callable *nominalEffectCallable,
	literal *ast.StructLiteral,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipValue, bool) {
	if literal == nil || len(literal.FunctionalUpdateSources) != 0 {
		return nominalOwnershipValue{}, false
	}
	literalType := a.effects.g.inferredExpressionTypeExpr(callable.ctx, literal)
	if literalType == nil && literal.StructType != nil {
		literalType = ast.Ty(literal.StructType.Name)
	}
	literalInfo, literalNominal := a.effects.g.structInfoForTypeExpr(callablePackage(callable), literalType)
	contained := make(map[int]struct{})
	for _, source := range callable.containedExpressionSources(literal) {
		contained[source] = struct{}{}
	}
	for _, field := range literal.Fields {
		if field == nil {
			continue
		}
		if value, ok := a.successorForExpression(callable, field.Value, env); ok {
			contained[value.source] = struct{}{}
		}
	}
	for source := range contained {
		if source < 0 || source >= len(callable.params) {
			continue
		}
		sourceType := callableParameterType(callable, source, callable.params[source])
		sourceInfo, sourceNominal := a.effects.g.structInfoForTypeExpr(callablePackage(callable), sourceType)
		if literalNominal && sourceNominal && literalInfo != nil && sourceInfo != nil &&
			literalInfo.GoName == sourceInfo.GoName {
			return nominalOwnershipValue{source: source}, true
		}
	}
	if literalNominal && literalInfo != nil {
		matching := -1
		for source, param := range callable.params {
			sourceType := callableParameterType(callable, source, param)
			sourceInfo, sourceNominal := a.effects.g.structInfoForTypeExpr(callablePackage(callable), sourceType)
			if !sourceNominal || sourceInfo == nil || sourceInfo.GoName != literalInfo.GoName {
				continue
			}
			if matching >= 0 {
				matching = -2
				break
			}
			matching = source
		}
		if matching >= 0 {
			// A fresh result may replace its sole same-nominal input even when
			// construction does not read that input. Multiple matching inputs
			// are intentionally ambiguous and remain unproven.
			return nominalOwnershipValue{source: matching}, true
		}
	}
	var result nominalOwnershipValue
	found := false
	for index, field := range literal.Fields {
		if field == nil {
			continue
		}
		value, ok := a.successorForExpression(callable, field.Value, env)
		if !ok {
			continue
		}
		fieldName := ownershipFieldName(literal, field, index)
		if fieldName == "" {
			return nominalOwnershipValue{}, false
		}
		value.path = append([]string{fieldName}, value.path...)
		if found && !ownershipValuesEqual(result, value) {
			return nominalOwnershipValue{}, false
		}
		result, found = value, true
	}
	return result, found
}

func ownershipFieldName(literal *ast.StructLiteral, field *ast.StructFieldInitializer, index int) string {
	if field != nil && field.Name != nil && field.Name.Name != "" {
		return field.Name.Name
	}
	if field != nil && field.IsShorthand {
		if identifier, ok := field.Value.(*ast.Identifier); ok && identifier != nil {
			return identifier.Name
		}
	}
	if literal != nil && literal.IsPositional {
		return "#" + strconv.Itoa(index)
	}
	return ""
}

type nominalOwnershipCallClaim struct {
	source    nominalOwnershipValue
	path      []string
	targets   []*nominalEffectCallable
	call      *ast.FunctionCall
	parameter int
	argument  int
	dispatch  string
	method    string
	depth     int
}

func (a *nominalOwnershipAnalyzer) callClaim(
	caller *nominalEffectCallable,
	call *ast.FunctionCall,
	env map[string]nominalOwnershipValue,
) (nominalOwnershipCallClaim, bool) {
	if caller == nil || call == nil {
		return nominalOwnershipCallClaim{}, false
	}
	if target, offset, _ := a.effects.resolveCallTarget(caller, call); target != nil {
		if memberAccess, memberCall := call.Callee.(*ast.MemberAccessExpression); memberCall &&
			memberAccess != nil && offset == 1 {
			if source, ok := ownershipValueForExpression(memberAccess.Object, env); ok {
				if claim, proven := a.claimForTargets(source, 0, []*nominalEffectCallable{target}); proven {
					claim.call = call
					claim.argument = -1
					claim.dispatch = "direct"
					claim.method = ownershipCallMethod(call)
					return claim, true
				}
			}
		}
		for index, argument := range call.Arguments {
			source, ok := ownershipValueForExpression(argument, env)
			if !ok {
				continue
			}
			if claim, proven := a.claimForTargets(source, index+offset, []*nominalEffectCallable{target}); proven {
				claim.call = call
				claim.argument = index
				claim.dispatch = "direct"
				claim.method = ownershipCallMethod(call)
				return claim, true
			}
		}
		return nominalOwnershipCallClaim{}, false
	}
	targets, offset, ok := a.interfaceCallTargets(caller, call)
	if !ok || len(targets) == 0 {
		return nominalOwnershipCallClaim{}, false
	}
	for index, argument := range call.Arguments {
		source, present := ownershipValueForExpression(argument, env)
		if !present {
			continue
		}
		if claim, proven := a.claimForTargets(source, index+offset, targets); proven {
			claim.call = call
			claim.argument = index
			claim.dispatch = "native-interface"
			claim.method = ownershipCallMethod(call)
			return claim, true
		}
	}
	return nominalOwnershipCallClaim{}, false
}

func (a *nominalOwnershipAnalyzer) claimForTargets(
	source nominalOwnershipValue,
	parameter int,
	targets []*nominalEffectCallable,
) (nominalOwnershipCallClaim, bool) {
	var path []string
	for index, target := range targets {
		summary := a.byParam[target]
		value, ok := summary[parameter]
		if !ok {
			return nominalOwnershipCallClaim{}, false
		}
		if index == 0 {
			path = append([]string(nil), value.path...)
		} else if !ownershipPathsEqual(path, value.path) {
			return nominalOwnershipCallClaim{}, false
		}
	}
	return nominalOwnershipCallClaim{
		source:    source,
		path:      path,
		targets:   append([]*nominalEffectCallable(nil), targets...),
		parameter: parameter,
	}, true
}

func (a *nominalOwnershipAnalyzer) interfaceCallTargets(
	caller *nominalEffectCallable,
	call *ast.FunctionCall,
) ([]*nominalEffectCallable, int, bool) {
	memberAccess, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || memberAccess == nil || memberAccess.Safe {
		return nil, 0, false
	}
	member, ok := memberAccess.Member.(*ast.Identifier)
	if !ok || member == nil || member.Name == "" {
		return nil, 0, false
	}
	receiverType := a.effects.g.inferredExpressionTypeExpr(caller.ctx, memberAccess.Object)
	baseName, ok := typeExprBaseName(receiverType)
	if !ok {
		return nil, 0, false
	}
	_, interfacePackage, ok := a.effects.g.interfaceDefinitionForPackage(callablePackage(caller), baseName)
	if !ok {
		return nil, 0, false
	}
	var targets []*nominalEffectCallable
	seen := make(map[*nominalEffectCallable]struct{})
	for _, candidate := range a.effects.g.nativeInterfaceImplCandidates() {
		if candidate.impl == nil || candidate.info == nil ||
			candidate.impl.MethodName != member.Name {
			continue
		}
		_, candidatePackage, resolved := a.effects.g.interfaceDefinitionForPackage(
			candidate.info.Package,
			candidate.impl.InterfaceName,
		)
		if !resolved || candidatePackage != interfacePackage ||
			candidate.impl.InterfaceName != baseName {
			continue
		}
		target := a.effects.byInfo[candidate.info]
		if target == nil {
			return nil, 0, false
		}
		if _, exists := seen[target]; exists {
			continue
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].name < targets[j].name })
	return targets, 1, len(targets) > 0
}

func ownershipValueForExpression(expr ast.Expression, env map[string]nominalOwnershipValue) (nominalOwnershipValue, bool) {
	identifier, ok := expr.(*ast.Identifier)
	if !ok || identifier == nil {
		return nominalOwnershipValue{}, false
	}
	value, ok := env[identifier.Name]
	return value, ok
}

func cloneOwnershipEnv(source map[string]nominalOwnershipValue) map[string]nominalOwnershipValue {
	result := make(map[string]nominalOwnershipValue, len(source))
	for name, value := range source {
		value.path = append([]string(nil), value.path...)
		result[name] = value
	}
	return result
}

func ownershipValuesEqual(left, right nominalOwnershipValue) bool {
	return left.source == right.source && ownershipPathsEqual(left.path, right.path)
}

func ownershipPathsEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func ownershipCallMethod(call *ast.FunctionCall) string {
	if call == nil {
		return ""
	}
	if memberAccess, ok := call.Callee.(*ast.MemberAccessExpression); ok && memberAccess != nil {
		if member, ok := memberAccess.Member.(*ast.Identifier); ok && member != nil {
			return member.Name
		}
	}
	if identifier, ok := call.Callee.(*ast.Identifier); ok && identifier != nil {
		return identifier.Name
	}
	return ""
}
