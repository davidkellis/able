package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

const nominalEffectSchemaVersion = 1

// NominalEffectReport is an opt-in compiler diagnostic. It records conservative
// parameter effects but does not authorize or select a generated carrier.
type NominalEffectReport struct {
	SchemaVersion int                            `json:"schema_version"`
	Callables     []NominalCallableEffectSummary `json:"callables"`
	NominalTypes  []NominalTypeEffectSummary     `json:"nominal_types"`
	Totals        NominalEffectTotals            `json:"totals"`
}

type NominalEffectTotals struct {
	Callables                   int `json:"callables"`
	Parameters                  int `json:"parameters"`
	NominalParameters           int `json:"nominal_parameters"`
	ReadOnlyNonEscapingNominals int `json:"read_only_non_escaping_nominal_parameters"`
	MutationParameters          int `json:"mutation_parameters"`
	CaptureParameters           int `json:"capture_parameters"`
	ReturnAliasParameters       int `json:"return_alias_parameters"`
	UnknownCallParameters       int `json:"unknown_call_parameters"`
}

type NominalCallableEffectSummary struct {
	Callable        string                          `json:"callable"`
	Package         string                          `json:"package"`
	Kind            string                          `json:"kind"`
	GeneratedGoName string                          `json:"generated_go_name,omitempty"`
	Parameters      []NominalParameterEffectSummary `json:"nominal_parameters"`
}

type NominalParameterEffectSummary struct {
	Index                int      `json:"index"`
	Name                 string   `json:"name"`
	Type                 string   `json:"type"`
	Nominal              string   `json:"nominal"`
	FieldOrIndexWrite    bool     `json:"field_or_index_write"`
	CapturedOrStored     bool     `json:"captured_or_stored"`
	ReturnAlias          bool     `json:"return_alias"`
	UnknownOrDynamicCall bool     `json:"unknown_or_dynamic_call"`
	ReadOnlyNonEscaping  bool     `json:"read_only_non_escaping"`
	Blockers             []string `json:"blockers,omitempty"`
}

type NominalTypeEffectSummary struct {
	Package                         string `json:"package"`
	Nominal                         string `json:"nominal"`
	ParameterInstances              int    `json:"parameter_instances"`
	ReadOnlyNonEscapingInstances    int    `json:"read_only_non_escaping_instances"`
	MutationInstances               int    `json:"mutation_instances"`
	CaptureInstances                int    `json:"capture_instances"`
	ReturnAliasInstances            int    `json:"return_alias_instances"`
	UnknownCallInstances            int    `json:"unknown_call_instances"`
	AllInstancesReadOnlyNonEscaping bool   `json:"all_instances_read_only_non_escaping"`
}

type nominalParameterEffect uint8

const (
	nominalEffectMutation nominalParameterEffect = 1 << iota
	nominalEffectCapture
	nominalEffectReturnAlias
	nominalEffectUnknownCall
)

type nominalEffectEdge struct {
	sources     []int
	callee      *nominalEffectCallable
	calleeParam int
}

type nominalEffectCallable struct {
	info             *functionInfo
	parent           *nominalEffectCallable
	body             ast.Node
	params           []*ast.FunctionParameter
	paramTypes       []ast.TypeExpression
	ctx              *compileContext
	name             string
	kind             string
	direct           []nominalParameterEffect
	effects          []nominalParameterEffect
	typeAmbiguous    []bool
	aliases          map[string]map[int]struct{}
	callableBindings map[string][]*nominalEffectCallable
	edges            []nominalEffectEdge
}

type nominalEffectAnalyzer struct {
	g          *generator
	callables  []*nominalEffectCallable
	byInfo     map[*functionInfo]*nominalEffectCallable
	byLambda   map[*ast.LambdaExpression]*nominalEffectCallable
	byLocalDef map[*ast.FunctionDefinition]*nominalEffectCallable
}

func (g *generator) resolveNominalParameterEffects() *NominalEffectReport {
	analyzer := g.resolvedNominalEffectAnalyzer()
	if analyzer == nil {
		return nil
	}
	return analyzer.report()
}

func (a *nominalEffectAnalyzer) collectRootCallables() {
	for _, info := range a.g.environmentEffectFunctionInfos() {
		if info == nil || !info.Compileable || info.Definition == nil ||
			info.Definition.Body == nil || info.ExternBody != nil {
			continue
		}
		callable := &nominalEffectCallable{
			info:             info,
			body:             info.Definition.Body,
			params:           info.Definition.Params,
			ctx:              a.g.compileBodyContext(info),
			name:             info.QualifiedName,
			kind:             "function",
			callableBindings: make(map[string][]*nominalEffectCallable),
		}
		if callable.name == "" {
			callable.name = qualifiedName(info.Package, info.Name)
		}
		if a.g.implMethodByInfo[info] != nil || a.rootCallableIsMethod(info) {
			callable.kind = "method"
		}
		callable.paramTypes = make([]ast.TypeExpression, len(info.Params))
		for index := range info.Params {
			callable.paramTypes[index] = info.Params[index].TypeExpr
		}
		callable.init()
		a.byInfo[info] = callable
		a.callables = append(a.callables, callable)
	}
}

func (a *nominalEffectAnalyzer) rootCallableIsMethod(info *functionInfo) bool {
	for _, method := range a.g.methodList {
		if method != nil && method.Info == info {
			return true
		}
	}
	return false
}

func (c *nominalEffectCallable) init() {
	count := len(c.params)
	c.direct = make([]nominalParameterEffect, count)
	c.effects = make([]nominalParameterEffect, count)
	c.typeAmbiguous = make([]bool, count)
	c.aliases = make(map[string]map[int]struct{}, count)
	for index, param := range c.params {
		if name := functionParameterName(param); name != "" {
			c.aliases[name] = map[int]struct{}{index: {}}
		}
	}
}

func functionParameterName(param *ast.FunctionParameter) string {
	if param == nil {
		return ""
	}
	identifier, _ := param.Name.(*ast.Identifier)
	if identifier == nil {
		return ""
	}
	return identifier.Name
}

func (a *nominalEffectAnalyzer) collectNestedCallables(parent *nominalEffectCallable) {
	if parent == nil || parent.body == nil {
		return
	}
	a.walkCallableBody(parent, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.AssignmentExpression:
			identifier, _ := current.Left.(*ast.Identifier)
			if identifier == nil || identifier.Name == "" {
				return true
			}
			switch right := current.Right.(type) {
			case *ast.LambdaExpression:
				target := a.ensureLambdaCallable(parent, right)
				a.addCallableBinding(parent, identifier.Name, target)
			case *ast.Identifier:
				if target, ok := a.resolveLocalCallable(parent, right.Name); ok {
					a.addCallableBinding(parent, identifier.Name, target)
				}
			}
		case *ast.FunctionDefinition:
			if current.ID == nil || current.ID.Name == "" {
				return false
			}
			target := a.ensureLocalFunctionCallable(parent, current)
			a.addCallableBinding(parent, current.ID.Name, target)
			return false
		case *ast.LambdaExpression:
			a.ensureLambdaCallable(parent, current)
			return false
		}
		return true
	})
}

func (a *nominalEffectAnalyzer) addCallableBinding(parent *nominalEffectCallable, name string, target *nominalEffectCallable) {
	if parent == nil || name == "" || target == nil {
		return
	}
	for _, existing := range parent.callableBindings[name] {
		if existing == target {
			return
		}
	}
	parent.callableBindings[name] = append(parent.callableBindings[name], target)
}

func (a *nominalEffectAnalyzer) ensureLambdaCallable(parent *nominalEffectCallable, lambda *ast.LambdaExpression) *nominalEffectCallable {
	if lambda == nil {
		return nil
	}
	if existing := a.byLambda[lambda]; existing != nil {
		return existing
	}
	paramTypes := a.lambdaParameterTypes(parent, lambda)
	ctx := parent.ctx.closureChild()
	a.seedCallableContext(ctx, lambda.Params, paramTypes)
	target := &nominalEffectCallable{
		parent:           parent,
		body:             lambda.Body,
		params:           lambda.Params,
		paramTypes:       paramTypes,
		ctx:              ctx,
		name:             fmt.Sprintf("%s::lambda@%d", parent.name, len(a.callables)+1),
		kind:             "lambda",
		callableBindings: make(map[string][]*nominalEffectCallable),
	}
	target.init()
	a.byLambda[lambda] = target
	a.callables = append(a.callables, target)
	return target
}

func (a *nominalEffectAnalyzer) ensureLocalFunctionCallable(parent *nominalEffectCallable, def *ast.FunctionDefinition) *nominalEffectCallable {
	if def == nil {
		return nil
	}
	if existing := a.byLocalDef[def]; existing != nil {
		return existing
	}
	paramTypes := make([]ast.TypeExpression, len(def.Params))
	for index, param := range def.Params {
		if param != nil {
			paramTypes[index] = param.ParamType
		}
	}
	ctx := parent.ctx.closureChild()
	a.seedCallableContext(ctx, def.Params, paramTypes)
	name := "<local>"
	if def.ID != nil && def.ID.Name != "" {
		name = def.ID.Name
	}
	target := &nominalEffectCallable{
		parent:           parent,
		body:             def.Body,
		params:           def.Params,
		paramTypes:       paramTypes,
		ctx:              ctx,
		name:             parent.name + "::" + name,
		kind:             "local-function",
		callableBindings: make(map[string][]*nominalEffectCallable),
	}
	target.init()
	a.byLocalDef[def] = target
	a.callables = append(a.callables, target)
	return target
}

func (a *nominalEffectAnalyzer) lambdaParameterTypes(parent *nominalEffectCallable, lambda *ast.LambdaExpression) []ast.TypeExpression {
	result := make([]ast.TypeExpression, len(lambda.Params))
	if parent != nil && parent.ctx != nil {
		if inferred, ok := a.g.inferredExpressionTypeExpr(parent.ctx, lambda).(*ast.FunctionTypeExpression); ok && inferred != nil {
			for index := range result {
				if index < len(inferred.ParamTypes) {
					result[index] = inferred.ParamTypes[index]
				}
			}
		}
	}
	for index, param := range lambda.Params {
		if param != nil && param.ParamType != nil {
			result[index] = param.ParamType
		}
	}
	return result
}

func (a *nominalEffectAnalyzer) seedCallableContext(ctx *compileContext, params []*ast.FunctionParameter, types []ast.TypeExpression) {
	if ctx == nil {
		return
	}
	for index, param := range params {
		name := functionParameterName(param)
		if name == "" {
			continue
		}
		var typeExpr ast.TypeExpression
		if index < len(types) {
			typeExpr = types[index]
		}
		goType, _ := a.g.lowerCarrierType(ctx, typeExpr)
		ctx.locals[name] = paramInfo{
			Name:     name,
			GoName:   safeParamName(name, index),
			GoType:   goType,
			TypeExpr: typeExpr,
		}
	}
}

func (a *nominalEffectAnalyzer) collectAliases(callable *nominalEffectCallable) {
	if callable == nil {
		return
	}
	for changed := true; changed; {
		changed = false
		a.walkCallableBody(callable, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignmentExpression)
			if !ok || assignment == nil {
				return true
			}
			identifier, _ := assignment.Left.(*ast.Identifier)
			if identifier == nil || identifier.Name == "" {
				return true
			}
			sources := a.expressionAliasSources(callable, assignment.Right)
			if len(sources) == 0 {
				return true
			}
			if mergeEffectSources(callable.aliases, identifier.Name, sources) {
				changed = true
			}
			return true
		})
	}
}

func mergeEffectSources(target map[string]map[int]struct{}, name string, sources []int) bool {
	if name == "" || len(sources) == 0 {
		return false
	}
	set := target[name]
	if set == nil {
		set = make(map[int]struct{}, len(sources))
		target[name] = set
	}
	changed := false
	for _, source := range sources {
		if _, exists := set[source]; exists {
			continue
		}
		set[source] = struct{}{}
		changed = true
	}
	return changed
}

func (c *nominalEffectCallable) directExpressionSources(expr ast.Expression) []int {
	identifier, _ := expr.(*ast.Identifier)
	if identifier == nil {
		return nil
	}
	return sortedEffectSources(c.aliases[identifier.Name])
}

func (a *nominalEffectAnalyzer) expressionAliasSources(callable *nominalEffectCallable, expr ast.Expression) []int {
	direct := callable.directExpressionSources(expr)
	if len(direct) > 0 || callable == nil || callable.ctx == nil || expr == nil {
		return direct
	}
	expressionType := a.g.inferredExpressionTypeExpr(callable.ctx, expr)
	if expressionType == nil {
		return nil
	}
	contained := callable.containedExpressionSources(expr)
	result := make([]int, 0, len(contained))
	for _, source := range contained {
		if source < 0 || source >= len(callable.paramTypes) {
			continue
		}
		parameterType := a.effectConcreteType(callable, source)
		if parameterType != nil &&
			a.effectTypesEqual(callable, expressionType, callable, parameterType) {
			result = append(result, source)
		}
	}
	return result
}

func sortedEffectSources(set map[int]struct{}) []int {
	result := make([]int, 0, len(set))
	for index := range set {
		result = append(result, index)
	}
	sort.Ints(result)
	return result
}

func (a *nominalEffectAnalyzer) collectDirectEffects(callable *nominalEffectCallable) {
	if callable == nil || callable.body == nil {
		return
	}
	a.collectNestedCaptures(callable)
	a.walkCallableBody(callable, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.AssignmentExpression:
			a.collectAssignmentEffects(callable, current)
		case *ast.ReturnStatement:
			direct := a.returnAliasSources(callable, current.Argument)
			callable.addDirect(direct, nominalEffectReturnAlias)
		case *ast.SpawnExpression:
			callable.addDirect(callable.containedExpressionSources(current.Expression), nominalEffectCapture)
		case *ast.StructFieldInitializer:
			callable.addDirect(a.expressionAliasSources(callable, current.Value), nominalEffectCapture)
		case *ast.ArrayLiteral:
			for _, element := range current.Elements {
				callable.addDirect(a.expressionAliasSources(callable, element), nominalEffectCapture)
			}
		case *ast.MapLiteralEntry:
			callable.addDirect(a.expressionAliasSources(callable, current.Key), nominalEffectCapture)
			callable.addDirect(a.expressionAliasSources(callable, current.Value), nominalEffectCapture)
		case *ast.MapLiteralSpread:
			callable.addDirect(a.expressionAliasSources(callable, current.Expression), nominalEffectCapture)
		case *ast.FunctionCall:
			a.collectCallEffects(callable, current)
		}
		return true
	})
	if tail := callableTailExpression(callable.body); tail != nil {
		direct := a.returnAliasSources(callable, tail)
		callable.addDirect(direct, nominalEffectReturnAlias)
	}
}

func (a *nominalEffectAnalyzer) returnAliasSources(
	callable *nominalEffectCallable,
	expr ast.Expression,
) []int {
	switch expr.(type) {
	case *ast.StructLiteral, *ast.ArrayLiteral, *ast.MapLiteral:
		// Aggregate literals create a fresh root. Directly embedded nominal
		// identities are already classified as captures by their field or
		// element nodes; reads used to construct the root are not return aliases.
		return nil
	default:
		return a.expressionAliasSources(callable, expr)
	}
}

func (a *nominalEffectAnalyzer) collectNestedCaptures(callable *nominalEffectCallable) {
	for _, nested := range a.callables {
		if nested == nil || nested.parent != callable || nested.body == nil {
			continue
		}
		for name, sources := range callable.aliases {
			if nestedBindsName(nested, name) || !astNodeReferencesIdentifier(nested.body, name) {
				continue
			}
			callable.addDirect(sortedEffectSources(sources), nominalEffectCapture)
		}
	}
}

func nestedBindsName(callable *nominalEffectCallable, name string) bool {
	for _, param := range callable.params {
		if functionParameterName(param) == name {
			return true
		}
	}
	return astNodeDeclaresIdentifier(callable.body, name)
}

func (a *nominalEffectAnalyzer) collectAssignmentEffects(callable *nominalEffectCallable, assignment *ast.AssignmentExpression) {
	if assignment == nil {
		return
	}
	switch left := assignment.Left.(type) {
	case *ast.MemberAccessExpression:
		callable.addDirect(a.expressionAliasSources(callable, left.Object), nominalEffectMutation)
		callable.addDirect(a.expressionAliasSources(callable, assignment.Right), nominalEffectCapture)
	case *ast.IndexExpression:
		callable.addDirect(a.expressionAliasSources(callable, left.Object), nominalEffectMutation)
		callable.addDirect(a.expressionAliasSources(callable, assignment.Right), nominalEffectCapture)
	}
}

func (a *nominalEffectAnalyzer) collectCallEffects(callable *nominalEffectCallable, call *ast.FunctionCall) {
	if call == nil {
		return
	}
	target, offset, receiverSources := a.resolveCallTarget(callable, call)
	allSources := append([]int(nil), receiverSources...)
	kernelName := ""
	if identifier, ok := call.Callee.(*ast.Identifier); ok && identifier != nil {
		kernelName = identifier.Name
	}
	if target != nil && len(receiverSources) > 0 {
		a.addEdgeOrUnknown(callable, receiverSources, target, 0)
	}
	for index, argument := range call.Arguments {
		direct := a.expressionAliasSources(callable, argument)
		if len(direct) == 0 {
			continue
		}
		if effect, known := kernelNominalArgumentEffect(kernelName, index); known {
			callable.addDirect(direct, effect)
			continue
		}
		allSources = append(allSources, direct...)
		if target == nil {
			continue
		}
		a.addEdgeOrUnknown(callable, direct, target, index+offset)
	}
	if target == nil {
		callable.addDirect(uniqueEffectSources(allSources), nominalEffectUnknownCall)
	}
}

// kernelNominalArgumentEffect describes language-kernel storage boundaries,
// independent of the source nominal type flowing through them. These are
// semantic intrinsics, not named-container lowering rules.
func kernelNominalArgumentEffect(name string, argument int) (nominalParameterEffect, bool) {
	switch name {
	case "__able_array_write":
		if argument == 2 {
			return nominalEffectCapture, true
		}
	}
	return 0, false
}

func (a *nominalEffectAnalyzer) addEdgeOrUnknown(caller *nominalEffectCallable, sources []int, callee *nominalEffectCallable, param int) {
	if callee == nil || param < 0 || param >= len(callee.params) {
		caller.addDirect(sources, nominalEffectUnknownCall)
		return
	}
	caller.edges = append(caller.edges, nominalEffectEdge{
		sources:     append([]int(nil), sources...),
		callee:      callee,
		calleeParam: param,
	})
}

func (a *nominalEffectAnalyzer) resolveCallTarget(callable *nominalEffectCallable, call *ast.FunctionCall) (*nominalEffectCallable, int, []int) {
	switch callee := call.Callee.(type) {
	case *ast.LambdaExpression:
		return a.byLambda[callee], 0, nil
	case *ast.Identifier:
		if target, ok := a.resolveLocalCallable(callable, callee.Name); ok {
			return target, 0, nil
		}
		info, overload, ok := a.g.resolveStaticCallable(callable.ctx, callee.Name)
		if !ok || info == nil || overload != nil {
			return nil, 0, nil
		}
		return a.byInfo[info], 0, nil
	case *ast.MemberAccessExpression:
		member, _ := callee.Member.(*ast.Identifier)
		if member == nil || member.Name == "" || callee.Safe {
			return nil, 0, a.expressionAliasSources(callable, callee.Object)
		}
		receiverSources := a.expressionAliasSources(callable, callee.Object)
		goType := ""
		probe := callable.ctx.probeChild()
		probe.analysisOnly = true
		if _, _, compiledType, ok := a.g.compileDispatchReceiverExpr(probe, callee.Object); ok {
			goType = compiledType
		}
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			typeExpr := a.g.inferredExpressionTypeExpr(callable.ctx, callee.Object)
			typeExpr = callable.ctx.substituteTypeBindings(typeExpr)
			goType, _ = a.g.lowerCarrierType(callable.ctx, typeExpr)
		}
		method := a.g.methodForReceiver(goType, member.Name)
		if method == nil {
			if resolved, ok := a.g.resolveStaticMethodCallForCall(callable.ctx, call, callee.Object, member.Name); ok {
				method = resolved
			}
		}
		if method == nil || method.Info == nil {
			return nil, 0, a.expressionAliasSources(callable, callee.Object)
		}
		target := a.byInfo[method.Info]
		if method.ExpectsSelf {
			return target, 1, receiverSources
		}
		return target, 0, nil
	default:
		return nil, 0, a.expressionAliasSources(callable, call.Callee)
	}
}

func (a *nominalEffectAnalyzer) resolveLocalCallable(callable *nominalEffectCallable, name string) (*nominalEffectCallable, bool) {
	for current := callable; current != nil; current = current.parent {
		targets, exists := current.callableBindings[name]
		if !exists {
			continue
		}
		if len(targets) != 1 || targets[0] == nil {
			return nil, false
		}
		return targets[0], true
	}
	return nil, false
}

func (c *nominalEffectCallable) containedExpressionSources(expr ast.Expression) []int {
	if expr == nil {
		return nil
	}
	set := make(map[int]struct{})
	ast.Walk(expr, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.LambdaExpression, *ast.FunctionDefinition, *ast.IteratorLiteral:
			return false
		case *ast.Identifier:
			for source := range c.aliases[current.Name] {
				set[source] = struct{}{}
			}
		}
		return true
	})
	return sortedEffectSources(set)
}

func (c *nominalEffectCallable) addDirect(sources []int, effect nominalParameterEffect) {
	for _, source := range sources {
		if source >= 0 && source < len(c.direct) {
			c.direct[source] |= effect
		}
	}
}

func uniqueEffectSources(sources []int) []int {
	set := make(map[int]struct{}, len(sources))
	for _, source := range sources {
		set[source] = struct{}{}
	}
	return sortedEffectSources(set)
}

func callableTailExpression(body ast.Node) ast.Expression {
	switch current := body.(type) {
	case *ast.BlockExpression:
		if current == nil || len(current.Body) == 0 {
			return nil
		}
		expression, _ := current.Body[len(current.Body)-1].(ast.Expression)
		return expression
	case ast.Expression:
		return current
	default:
		return nil
	}
}

func (a *nominalEffectAnalyzer) walkCallableBody(callable *nominalEffectCallable, visit func(ast.Node) bool) {
	if callable == nil || callable.body == nil {
		return
	}
	ast.Walk(callable.body, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		if node != callable.body {
			switch node.(type) {
			case *ast.LambdaExpression, *ast.FunctionDefinition, *ast.IteratorLiteral:
				visit(node)
				return false
			}
		}
		return visit(node)
	})
}

func (a *nominalEffectAnalyzer) closeFixedPoint() {
	for _, callable := range a.callables {
		copy(callable.effects, callable.direct)
	}
	for changed := true; changed; {
		changed = false
		for _, caller := range a.callables {
			for _, edge := range caller.edges {
				if edge.callee == nil || edge.calleeParam < 0 || edge.calleeParam >= len(edge.callee.effects) {
					continue
				}
				effect := edge.callee.effects[edge.calleeParam]
				for _, source := range edge.sources {
					if source < 0 || source >= len(caller.effects) {
						continue
					}
					next := caller.effects[source] | effect
					if next != caller.effects[source] {
						caller.effects[source] = next
						changed = true
					}
				}
			}
		}
	}
}

func (a *nominalEffectAnalyzer) closeParameterTypeFixedPoint() {
	for changed := true; changed; {
		changed = false
		for _, caller := range a.callables {
			for _, edge := range caller.edges {
				if edge.callee == nil || edge.calleeParam < 0 ||
					edge.calleeParam >= len(edge.callee.paramTypes) {
					continue
				}
				calleeType := a.effectConcreteType(edge.callee, edge.calleeParam)
				for _, source := range edge.sources {
					if source < 0 || source >= len(caller.paramTypes) {
						continue
					}
					callerType := a.effectConcreteType(caller, source)
					switch {
					case callerType == nil && calleeType != nil && !caller.typeAmbiguous[source]:
						caller.paramTypes[source] = calleeType
						changed = true
					case calleeType == nil && callerType != nil && !edge.callee.typeAmbiguous[edge.calleeParam]:
						edge.callee.paramTypes[edge.calleeParam] = callerType
						changed = true
					case callerType != nil && calleeType != nil &&
						!a.effectTypesEqual(caller, callerType, edge.callee, calleeType):
						if !caller.typeAmbiguous[source] {
							caller.typeAmbiguous[source] = true
							caller.paramTypes[source] = nil
							changed = true
						}
						if !edge.callee.typeAmbiguous[edge.calleeParam] {
							edge.callee.typeAmbiguous[edge.calleeParam] = true
							edge.callee.paramTypes[edge.calleeParam] = nil
							changed = true
						}
					}
				}
			}
		}
	}
}

func (a *nominalEffectAnalyzer) effectConcreteType(callable *nominalEffectCallable, index int) ast.TypeExpression {
	if callable == nil || index < 0 || index >= len(callable.paramTypes) ||
		callable.typeAmbiguous[index] {
		return nil
	}
	typeExpr := callable.paramTypes[index]
	if typeExpr == nil {
		return nil
	}
	pkgName := callablePackage(callable)
	typeExpr = normalizeTypeExprForPackage(a.g, pkgName, typeExpr)
	if !a.g.typeExprFullyBound(pkgName, typeExpr) {
		return nil
	}
	return typeExpr
}

func (a *nominalEffectAnalyzer) effectTypesEqual(
	leftCallable *nominalEffectCallable,
	left ast.TypeExpression,
	rightCallable *nominalEffectCallable,
	right ast.TypeExpression,
) bool {
	leftKey := normalizeTypeExprIdentityKey(a.g, callablePackage(leftCallable), left)
	rightKey := normalizeTypeExprIdentityKey(a.g, callablePackage(rightCallable), right)
	return leftKey != "" && leftKey == rightKey
}

func (a *nominalEffectAnalyzer) resetEffectsWithInferredParameterTypes() {
	for _, callable := range a.callables {
		callable.direct = make([]nominalParameterEffect, len(callable.params))
		callable.effects = make([]nominalParameterEffect, len(callable.params))
		callable.edges = nil
		if callable.ctx == nil {
			continue
		}
		for index, param := range callable.params {
			name := functionParameterName(param)
			if name == "" || index >= len(callable.paramTypes) {
				continue
			}
			typeExpr := callable.paramTypes[index]
			goType, _ := a.g.lowerCarrierType(callable.ctx, typeExpr)
			if binding, ok := callable.ctx.lookupCurrent(name); ok {
				binding.TypeExpr = typeExpr
				binding.GoType = goType
				callable.ctx.setLocalBinding(name, binding)
			}
		}
	}
}

func (a *nominalEffectAnalyzer) report() *NominalEffectReport {
	report := &NominalEffectReport{SchemaVersion: nominalEffectSchemaVersion}
	report.Totals.Callables = len(a.callables)
	typeSummaries := make(map[string]*NominalTypeEffectSummary)
	for _, callable := range a.callables {
		report.Totals.Parameters += len(callable.params)
		summary := NominalCallableEffectSummary{
			Callable: callable.name,
			Package:  callablePackage(callable),
			Kind:     callable.kind,
		}
		if callable.info != nil {
			summary.GeneratedGoName = callable.info.GoName
		}
		for index, param := range callable.params {
			typeExpr := callableParameterType(callable, index, param)
			info, ok := a.g.structInfoForTypeExpr(summary.Package, typeExpr)
			if !ok || info == nil || info.Node == nil {
				continue
			}
			effect := callable.effects[index]
			parameter := nominalParameterSummary(index, functionParameterName(param), typeExpr, info, effect)
			summary.Parameters = append(summary.Parameters, parameter)
			report.Totals.NominalParameters++
			if parameter.ReadOnlyNonEscaping {
				report.Totals.ReadOnlyNonEscapingNominals++
			}
			if parameter.FieldOrIndexWrite {
				report.Totals.MutationParameters++
			}
			if parameter.CapturedOrStored {
				report.Totals.CaptureParameters++
			}
			if parameter.ReturnAlias {
				report.Totals.ReturnAliasParameters++
			}
			if parameter.UnknownOrDynamicCall {
				report.Totals.UnknownCallParameters++
			}
			key := info.Package + "::" + info.Name
			aggregate := typeSummaries[key]
			if aggregate == nil {
				aggregate = &NominalTypeEffectSummary{
					Package: info.Package,
					Nominal: info.Name,
				}
				typeSummaries[key] = aggregate
			}
			aggregate.ParameterInstances++
			if parameter.ReadOnlyNonEscaping {
				aggregate.ReadOnlyNonEscapingInstances++
			}
			if parameter.FieldOrIndexWrite {
				aggregate.MutationInstances++
			}
			if parameter.CapturedOrStored {
				aggregate.CaptureInstances++
			}
			if parameter.ReturnAlias {
				aggregate.ReturnAliasInstances++
			}
			if parameter.UnknownOrDynamicCall {
				aggregate.UnknownCallInstances++
			}
		}
		if len(summary.Parameters) > 0 {
			report.Callables = append(report.Callables, summary)
		}
	}
	sort.Slice(report.Callables, func(i, j int) bool {
		if report.Callables[i].Package != report.Callables[j].Package {
			return report.Callables[i].Package < report.Callables[j].Package
		}
		return report.Callables[i].Callable < report.Callables[j].Callable
	})
	for _, summary := range typeSummaries {
		summary.AllInstancesReadOnlyNonEscaping =
			summary.ParameterInstances > 0 &&
				summary.ParameterInstances == summary.ReadOnlyNonEscapingInstances
		report.NominalTypes = append(report.NominalTypes, *summary)
	}
	sort.Slice(report.NominalTypes, func(i, j int) bool {
		if report.NominalTypes[i].Package != report.NominalTypes[j].Package {
			return report.NominalTypes[i].Package < report.NominalTypes[j].Package
		}
		return report.NominalTypes[i].Nominal < report.NominalTypes[j].Nominal
	})
	return report
}

func callablePackage(callable *nominalEffectCallable) string {
	for current := callable; current != nil; current = current.parent {
		if current.info != nil {
			return current.info.Package
		}
		if current.ctx != nil && current.ctx.packageName != "" {
			return current.ctx.packageName
		}
	}
	return ""
}

func callableParameterType(callable *nominalEffectCallable, index int, param *ast.FunctionParameter) ast.TypeExpression {
	var result ast.TypeExpression
	if index < len(callable.paramTypes) {
		result = callable.paramTypes[index]
	}
	if result == nil && param != nil {
		result = param.ParamType
	}
	if callable.ctx != nil {
		result = callable.ctx.substituteTypeBindings(result)
	}
	return result
}

func nominalParameterSummary(index int, name string, typeExpr ast.TypeExpression, info *structInfo, effect nominalParameterEffect) NominalParameterEffectSummary {
	result := NominalParameterEffectSummary{
		Index:                index,
		Name:                 name,
		Type:                 typeExpressionToString(typeExpr),
		Nominal:              info.Name,
		FieldOrIndexWrite:    effect&nominalEffectMutation != 0,
		CapturedOrStored:     effect&nominalEffectCapture != 0,
		ReturnAlias:          effect&nominalEffectReturnAlias != 0,
		UnknownOrDynamicCall: effect&nominalEffectUnknownCall != 0,
	}
	if result.FieldOrIndexWrite {
		result.Blockers = append(result.Blockers, "field-or-index-write")
	}
	if result.CapturedOrStored {
		result.Blockers = append(result.Blockers, "captured-or-stored")
	}
	if result.ReturnAlias {
		result.Blockers = append(result.Blockers, "return-alias")
	}
	if result.UnknownOrDynamicCall {
		result.Blockers = append(result.Blockers, "unknown-or-dynamic-call")
	}
	result.ReadOnlyNonEscaping = len(result.Blockers) == 0
	return result
}

func (r *NominalEffectReport) String() string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf(
		"callables=%d nominal_parameters=%d read_only_non_escaping=%d",
		r.Totals.Callables,
		r.Totals.NominalParameters,
		r.Totals.ReadOnlyNonEscapingNominals,
	))
}
