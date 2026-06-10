package main

import (
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

const (
	aggregateLength  = ".#length"
	aggregateElement = ".#element"
)

func (run *rangeRun) aggregateRoot(expr ast.Expr) string {
	switch typed := unparen(expr).(type) {
	case *ast.Ident:
		if root := run.aliases[typed.Name]; root != "" {
			return root
		}
		return typed.Name
	case *ast.SelectorExpr:
		if typed.Sel.Name == "Elements" {
			return run.aggregateRoot(typed.X)
		}
	case *ast.IndexExpr:
		if selector, ok := unparen(typed.X).(*ast.SelectorExpr); ok && selector.Sel.Name == "Elements" {
			if root := run.aggregateRoot(selector.X); root != "" {
				return root + aggregateElement
			}
		}
	}
	return ""
}

func importAggregateFacts(env rangeEnv, root string, facts map[string]intRange) {
	for suffix, value := range facts {
		env[root+suffix] = value
	}
}

func exportAggregateFacts(env rangeEnv, root string) map[string]intRange {
	if root == "" {
		return nil
	}
	facts := make(map[string]intRange)
	for key, value := range env {
		if strings.HasPrefix(key, root+".") {
			facts[strings.TrimPrefix(key, root)] = value
		}
	}
	if len(facts) == 0 {
		return nil
	}
	return facts
}

func copyAggregateFacts(env rangeEnv, target, source string) {
	if target == "" || source == "" || target == source {
		return
	}
	clearAggregateFacts(env, target)
	importAggregateFacts(env, target, exportAggregateFacts(env, source))
}

func clearAggregateFacts(env rangeEnv, root string) {
	for key := range env {
		if strings.HasPrefix(key, root+".") {
			delete(env, key)
		}
	}
}

func (run *rangeRun) initializeAggregateTarget(target string, expr ast.Expr, env rangeEnv) bool {
	if target == "" || expr == nil {
		return false
	}
	if source := run.aggregateRoot(expr); source != "" {
		run.aliases[target] = source
		return len(exportAggregateFacts(env, source)) > 0
	}
	unary, ok := unparen(expr).(*ast.UnaryExpr)
	if !ok || unary.Op != token.AND {
		return false
	}
	literal, ok := unparen(unary.X).(*ast.CompositeLit)
	if !ok {
		return false
	}
	run.aliases[target] = target
	clearAggregateFacts(env, target)
	typeName := calledName(literal.Type)
	if strings.HasPrefix(typeName, "__able_array_") {
		env[target+aggregateLength] = knownRange(0, 0)
	}
	for _, element := range literal.Elts {
		keyValue, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		field := identName(keyValue.Key)
		if field == "Elements" {
			if length, ok := staticSliceLength(keyValue.Value); ok {
				env[target+aggregateLength] = knownRange(length, length)
			}
			continue
		}
		if field != "" {
			env[target+"."+field] = run.evalExpr(keyValue.Value, env)
		}
	}
	return true
}

func staticSliceLength(expr ast.Expr) (int64, bool) {
	composite, ok := unparen(expr).(*ast.CompositeLit)
	if ok {
		return int64(len(composite.Elts)), true
	}
	call, ok := unparen(expr).(*ast.CallExpr)
	if !ok || calledName(call.Fun) != "make" || len(call.Args) < 2 {
		return 0, false
	}
	return integerLiteral(call.Args[1])
}

func (run *rangeRun) assignAggregate(statement *ast.AssignStmt, env rangeEnv) {
	if statement == nil {
		return
	}
	if len(statement.Lhs) == 1 && len(statement.Rhs) == 1 {
		if run.assignAppend(statement.Lhs[0], statement.Rhs[0], env) {
			return
		}
		if selector, ok := unparen(statement.Lhs[0]).(*ast.SelectorExpr); ok && selector.Sel.Name != "Elements" {
			if root := run.aggregateRoot(selector.X); root != "" {
				key := root + "." + selector.Sel.Name
				value := run.evalExpr(statement.Rhs[0], env)
				if previous, exists := env[key]; exists {
					env[key] = unionRange(previous, value)
				} else {
					env[key] = value
				}
			}
			return
		}
	}
	for idx, lhs := range statement.Lhs {
		name := identName(lhs)
		if name == "" || name == "_" || idx >= len(statement.Rhs) {
			continue
		}
		run.initializeAggregateTarget(name, statement.Rhs[idx], env)
	}
}

func (run *rangeRun) assignAppend(lhs, rhs ast.Expr, env rangeEnv) bool {
	selector, ok := unparen(lhs).(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Elements" {
		return false
	}
	call, ok := unparen(rhs).(*ast.CallExpr)
	if !ok || calledName(call.Fun) != "append" || len(call.Args) < 2 {
		return false
	}
	root := run.aggregateRoot(selector.X)
	if root == "" || run.aggregateRoot(call.Args[0]) != root {
		return false
	}
	lengthKey := root + aggregateLength
	if length := env[lengthKey]; length.Known {
		env[lengthKey] = addRange(length, knownRange(int64(len(call.Args)-1), int64(len(call.Args)-1)))
	}
	for _, valueExpr := range call.Args[1:] {
		value := run.evalExpr(valueExpr, env)
		elementKey := root + aggregateElement
		if previous, exists := env[elementKey]; exists {
			env[elementKey] = unionRange(previous, value)
		} else {
			env[elementKey] = value
		}
		if nested := run.aggregateRoot(valueExpr); nested != "" {
			copyAggregateFacts(env, root+aggregateElement, nested)
		}
	}
	return true
}

func (run *rangeRun) aggregateCallFacts(call *ast.CallExpr, env rangeEnv) []map[string]intRange {
	facts := make([]map[string]intRange, len(call.Args))
	for idx, argument := range call.Args {
		facts[idx] = exportAggregateFacts(env, run.aggregateRoot(argument))
	}
	return facts
}

func (run *rangeRun) applyCallAggregateEffects(call *ast.CallExpr, env rangeEnv) {
	name := calledName(call.Fun)
	callee := run.functions[name]
	if !run.useClosedSummaries || callee == nil || !callee.Direct {
		return
	}
	for idx, argument := range call.Args {
		if idx >= len(callee.closedParamFacts) {
			continue
		}
		root := run.aggregateRoot(argument)
		if root == "" {
			continue
		}
		clearAggregateFacts(env, root)
		importAggregateFacts(env, root, callee.closedParamFacts[idx])
	}
}

func (run *rangeRun) invalidateUnknownCallFacts(call *ast.CallExpr, env rangeEnv) {
	if call == nil {
		return
	}
	name := calledName(call.Fun)
	if name == "append" || name == "len" || name == "cap" || name == "__able_slice_len" {
		return
	}
	if callee := run.functions[name]; callee != nil && callee.Direct {
		return
	}
	for _, argument := range call.Args {
		if root := run.aggregateRoot(argument); root != "" {
			clearAggregateFacts(env, root)
		}
	}
}

func (run *rangeRun) assignCallReturnFacts(target string, call *ast.CallExpr, env rangeEnv) {
	if target == "" || call == nil || !run.useClosedSummaries {
		return
	}
	callee := run.functions[calledName(call.Fun)]
	if callee == nil || !callee.Direct || len(callee.closedReturnFacts) == 0 {
		return
	}
	run.aliases[target] = target
	clearAggregateFacts(env, target)
	importAggregateFacts(env, target, callee.closedReturnFacts)
}

func (run *rangeRun) observeReturnFacts(expr ast.Expr, env rangeEnv) {
	root := run.aggregateRoot(expr)
	facts := exportAggregateFacts(env, root)
	if !run.returnFactsSeen {
		run.returnFacts = cloneAggregateFacts(facts)
		run.returnFactsSeen = true
		return
	}
	run.returnFacts, _ = mergeCompleteAggregateFacts(run.returnFacts, facts)
}

func (run *rangeRun) parameterOutputFacts(params []functionParam) []map[string]intRange {
	facts := make([]map[string]intRange, len(params))
	for idx, param := range params {
		facts[idx] = exportAggregateFacts(run.env, run.aggregateRoot(&ast.Ident{Name: param.Name}))
	}
	return facts
}

func mergeAggregateFacts(left, right map[string]intRange) (map[string]intRange, bool) {
	if len(left) == 0 {
		copy := cloneAggregateFacts(right)
		return copy, len(copy) > 0
	}
	merged := cloneAggregateFacts(left)
	changed := false
	for key, rightValue := range right {
		if leftValue, ok := merged[key]; ok {
			value := unionRange(leftValue, rightValue)
			if !rangeEqual(value, leftValue) {
				merged[key] = value
				changed = true
			}
		} else {
			merged[key] = rightValue
			changed = true
		}
	}
	return merged, changed
}

func mergeCompleteAggregateFacts(left, right map[string]intRange) (map[string]intRange, bool) {
	merged := make(map[string]intRange, len(left)+len(right))
	changed := false
	for key, leftValue := range left {
		rightValue, exists := right[key]
		if !exists {
			merged[key] = intRange{}
			if leftValue.Known {
				changed = true
			}
			continue
		}
		value := unionRange(leftValue, rightValue)
		merged[key] = value
		if !rangeEqual(value, leftValue) {
			changed = true
		}
	}
	for key, rightValue := range right {
		if _, exists := left[key]; exists {
			continue
		}
		merged[key] = intRange{}
		if rightValue.Known {
			changed = true
		}
	}
	return merged, changed
}

func mergeAggregateObservations(paramCount int, bySite map[string][]map[string]intRange) []map[string]intRange {
	merged := make([]map[string]intRange, paramCount)
	seen := make([]bool, paramCount)
	for _, arguments := range bySite {
		for idx := 0; idx < paramCount; idx++ {
			var facts map[string]intRange
			if idx < len(arguments) {
				facts = arguments[idx]
			}
			if !seen[idx] {
				merged[idx] = cloneAggregateFacts(facts)
				seen[idx] = true
				continue
			}
			merged[idx], _ = mergeCompleteAggregateFacts(merged[idx], facts)
		}
	}
	return merged
}

func cloneAggregateFacts(source map[string]intRange) map[string]intRange {
	if len(source) == 0 {
		return nil
	}
	copy := make(map[string]intRange, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}

func aggregateFactsEqual(left, right map[string]intRange) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if !rangeEqual(value, right[key]) {
			return false
		}
	}
	return true
}

func cloneAggregateFactSlices(source []map[string]intRange) []map[string]intRange {
	copy := make([]map[string]intRange, len(source))
	for idx, facts := range source {
		copy[idx] = cloneAggregateFacts(facts)
	}
	return copy
}

func aggregateFactSlicesEqual(left, right []map[string]intRange) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if !aggregateFactsEqual(left[idx], right[idx]) {
			return false
		}
	}
	return true
}

func widenChangingAggregateFacts(previous, next map[string]intRange) map[string]intRange {
	widened := cloneAggregateFacts(next)
	for key, value := range widened {
		if prior, exists := previous[key]; exists && !rangeEqual(prior, value) {
			widened[key] = intRange{}
		}
	}
	return widened
}

func widenChangingAggregateFactSlices(previous, next []map[string]intRange) []map[string]intRange {
	widened := cloneAggregateFactSlices(next)
	for idx := range widened {
		var prior map[string]intRange
		if idx < len(previous) {
			prior = previous[idx]
		}
		widened[idx] = widenChangingAggregateFacts(prior, widened[idx])
	}
	return widened
}

func aggregateRangeReports(params []functionParam, facts []map[string]intRange) []aggregateRange {
	var reports []aggregateRange
	for idx, param := range params {
		if idx < len(facts) {
			reports = append(reports, aggregateFactReports(param.Name, facts[idx])...)
		}
	}
	return reports
}

func aggregateFactReports(root string, facts map[string]intRange) []aggregateRange {
	var reports []aggregateRange
	for suffix, value := range facts {
		if value.Known {
			reports = append(reports, aggregateRange{Path: root + suffix, Min: value.Min, Max: value.Max})
		}
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Path < reports[j].Path })
	return reports
}

func mergeLoopAggregateFacts(result, before, body rangeEnv, statement *ast.ForStmt, inductionName string) {
	iterations, exact := exactLoopIterations(statement, before, inductionName)
	for key, bodyValue := range body {
		if !strings.Contains(key, ".#") && !strings.Contains(key, ".") {
			continue
		}
		beforeValue, existed := before[key]
		if strings.HasSuffix(key, aggregateLength) && exact && existed &&
			beforeValue.Known && bodyValue.Known {
			deltaMin, minOK := sub64(bodyValue.Min, beforeValue.Min)
			deltaMax, maxOK := sub64(bodyValue.Max, beforeValue.Max)
			if minOK && maxOK && deltaMin == deltaMax && deltaMin >= 0 {
				minGrowth, minMulOK := mul64(deltaMin, iterations)
				maxGrowth, maxMulOK := mul64(deltaMax, iterations)
				if minMulOK && maxMulOK {
					minValue, minAddOK := add64(beforeValue.Min, minGrowth)
					maxValue, maxAddOK := add64(beforeValue.Max, maxGrowth)
					if minAddOK && maxAddOK {
						result[key] = knownRange(minValue, maxValue)
						continue
					}
				}
			}
		}
		if existed {
			result[key] = unionRange(beforeValue, bodyValue)
		} else if exact && iterations > 0 {
			result[key] = bodyValue
		}
	}
}

func exactLoopIterations(statement *ast.ForStmt, env rangeEnv, inductionName string) (int64, bool) {
	if inductionName == "" {
		return 0, false
	}
	_, values, ok := loopInductionRange(statement, env)
	if !ok || !values.Known {
		return 0, false
	}
	count, ok := add64(values.Max-values.Min, 1)
	return count, ok && count >= 0
}
