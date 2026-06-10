package main

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type functionRangeInfo struct {
	Params              []functionParam
	Universal           []blockerObservation
	Closed              []blockerObservation
	ClosedParams        []intRange
	ClosedReachable     bool
	UniversalHelpers    map[string]bool
	ClosedHelpers       map[string]bool
	UniversalRelations  []relationalObservation
	ClosedRelations     []relationalObservation
	ClosedAggregateArgs []map[string]intRange
}

func analyzeRangeEffects(fset *token.FileSet, functions map[string]*functionEffect) {
	infos := make(map[string]*functionRangeInfo, len(functions))
	for name, effect := range functions {
		infos[name] = &functionRangeInfo{Params: functionParameters(effect.decl)}
	}

	for name, effect := range functions {
		info := infos[name]
		env := make(rangeEnv)
		for _, param := range info.Params {
			if param.Integer {
				env[param.Name] = param.Full
			}
		}
		run := newRangeRun(fset, functions, effect, env)
		run.analyze()
		info.Universal = run.blockersWithMissingPrimitiveCalls()
		info.UniversalHelpers = safeHelpers(info.Universal)
		info.UniversalRelations = run.relations
	}

	resolveClosedDirectRanges(fset, functions, infos)
	universalFree := resolveRangeClosure(functions, infos, false)
	closedFree := resolveRangeClosure(functions, infos, true)

	for name, effect := range functions {
		info := infos[name]
		effect.UniversalRangeFree = universalFree[name]
		effect.ClosedDirectReachable = info.ClosedReachable
		effect.ClosedDirectRangeFree = info.ClosedReachable && closedFree[name]
		switch {
		case effect.ControlFree:
			effect.RangeClass = "control-free"
		case effect.UniversalRangeFree:
			effect.RangeClass = "universally-range-safe"
		case effect.ClosedDirectRangeFree:
			effect.RangeClass = "call-site-specializable"
		case effect.ClosedDirectReachable:
			effect.RangeClass = "reachable-unproven"
		default:
			effect.RangeClass = "not-main-reachable"
		}
		effect.PrimitiveRangeBlockers = mergeBlockerReports(fset, effect, info.Universal, info.Closed)
		effect.RelationalBlockers = mergeRelationalReports(fset, info.UniversalRelations, info.ClosedRelations)
		if info.ClosedReachable {
			for idx, param := range info.Params {
				if !param.Integer || idx >= len(info.ClosedParams) || !info.ClosedParams[idx].Known {
					continue
				}
				observed := info.ClosedParams[idx]
				effect.ClosedDirectParamRanges = append(effect.ClosedDirectParamRanges, parameterRange{
					Name: param.Name,
					Type: param.Type,
					Min:  observed.Min,
					Max:  observed.Max,
				})
			}
			effect.ClosedAggregateFacts = aggregateRangeReports(info.Params, info.ClosedAggregateArgs)
			effect.ClosedReturnFacts = aggregateFactReports("return", effect.closedReturnFacts)
		}
	}
}

func functionParameters(decl *ast.FuncDecl) []functionParam {
	if decl == nil || decl.Type == nil || decl.Type.Params == nil {
		return nil
	}
	var params []functionParam
	for _, field := range decl.Type.Params.List {
		if field == nil {
			continue
		}
		typeName := integerTypeName(field.Type)
		full, integer := fullIntegerRange(typeName)
		for _, name := range field.Names {
			if name != nil {
				params = append(params, functionParam{Name: name.Name, Type: typeName, Integer: integer, Full: full})
			}
		}
	}
	return params
}

func resolveClosedDirectRanges(fset *token.FileSet, functions map[string]*functionEffect, infos map[string]*functionRangeInfo) {
	root := functions["__able_compiled_fn_main"]
	if root == nil || !root.Direct {
		return
	}
	rootInfo := infos[root.Name]
	rootInfo.ClosedReachable = true
	rootInfo.ClosedParams = make([]intRange, len(rootInfo.Params))
	rootInfo.ClosedAggregateArgs = make([]map[string]intRange, len(rootInfo.Params))
	queue := []string{root.Name}
	inQueue := map[string]bool{root.Name: true}
	iterations := make(map[string]int)
	aggregateObservations := make(map[string]map[string][]map[string]intRange)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		inQueue[name] = false
		effect := functions[name]
		info := infos[name]
		iterations[name]++
		env := make(rangeEnv)
		for idx, param := range info.Params {
			if param.Integer && idx < len(info.ClosedParams) {
				env[param.Name] = info.ClosedParams[idx]
			}
			if idx < len(info.ClosedAggregateArgs) {
				importAggregateFacts(env, param.Name, info.ClosedAggregateArgs[idx])
			}
		}
		run := newRangeRun(fset, functions, effect, env)
		run.useClosedSummaries = true
		run.analyze()
		info.Closed = run.blockersWithMissingPrimitiveCalls()
		info.ClosedHelpers = safeHelpers(info.Closed)
		info.ClosedRelations = run.relations
		returnChanged := effect.closedReturnSet != run.returnedSet ||
			(effect.closedReturnSet && !rangeEqual(effect.closedReturnRange, run.returned))
		effect.closedReturnSet = run.returnedSet
		effect.closedReturnRange = run.returned
		nextReturnFacts := cloneAggregateFacts(run.returnFacts)
		paramFacts := run.parameterOutputFacts(info.Params)
		nextParamFacts := cloneAggregateFactSlices(paramFacts)
		if iterations[name] > 64 {
			nextReturnFacts = widenChangingAggregateFacts(effect.closedReturnFacts, nextReturnFacts)
			nextParamFacts = widenChangingAggregateFactSlices(effect.closedParamFacts, nextParamFacts)
		}
		factsChanged := !aggregateFactsEqual(effect.closedReturnFacts, nextReturnFacts) ||
			!aggregateFactSlicesEqual(effect.closedParamFacts, nextParamFacts)
		effect.closedReturnFacts = nextReturnFacts
		effect.closedParamFacts = nextParamFacts
		for _, bySite := range aggregateObservations {
			for site := range bySite {
				if strings.HasPrefix(site, name+":") {
					delete(bySite, site)
				}
			}
		}
		for callIndex, call := range run.calls {
			callee := functions[call.Callee]
			calleeInfo := infos[call.Callee]
			if callee == nil || calleeInfo == nil || !callee.Direct {
				continue
			}
			if !calleeInfo.ClosedReachable {
				calleeInfo.ClosedReachable = true
				calleeInfo.ClosedParams = make([]intRange, len(calleeInfo.Params))
				calleeInfo.ClosedAggregateArgs = make([]map[string]intRange, len(calleeInfo.Params))
			}
			if aggregateObservations[call.Callee] == nil {
				aggregateObservations[call.Callee] = make(map[string][]map[string]intRange)
			}
			site := name + ":" + strconv.Itoa(callIndex)
			aggregateObservations[call.Callee][site] = cloneAggregateFactSlices(call.AggregateFacts)
			changed := false
			for idx, param := range calleeInfo.Params {
				if !param.Integer {
					continue
				}
				observed := param.Full
				if idx < len(call.Args) && call.Args[idx].Known {
					observed = intersectRange(call.Args[idx], param.Full)
				}
				if !calleeInfo.ClosedParams[idx].Known {
					calleeInfo.ClosedParams[idx] = observed
					changed = true
					continue
				}
				merged := unionRange(calleeInfo.ClosedParams[idx], observed)
				if !rangeEqual(merged, calleeInfo.ClosedParams[idx]) {
					calleeInfo.ClosedParams[idx] = merged
					changed = true
				}
			}
			nextAggregateArgs := mergeAggregateObservations(
				len(calleeInfo.Params), aggregateObservations[call.Callee],
			)
			if !aggregateFactSlicesEqual(calleeInfo.ClosedAggregateArgs, nextAggregateArgs) {
				calleeInfo.ClosedAggregateArgs = nextAggregateArgs
				changed = true
			}
			if changed && iterations[call.Callee] > 4096 {
				for idx, param := range calleeInfo.Params {
					if param.Integer {
						calleeInfo.ClosedParams[idx] = param.Full
					}
				}
			}
			if (changed || iterations[call.Callee] == 0) && !inQueue[call.Callee] {
				queue = append(queue, call.Callee)
				inQueue[call.Callee] = true
			}
		}
		if returnChanged || factsChanged {
			for callerName, callerInfo := range infos {
				if callerInfo.ClosedReachable && callerName != name && !inQueue[callerName] {
					queue = append(queue, callerName)
					inQueue[callerName] = true
				}
			}
		}
	}
}

func resolveRangeClosure(functions map[string]*functionEffect, infos map[string]*functionRangeInfo, closed bool) map[string]bool {
	free := make(map[string]bool, len(functions))
	for name, effect := range functions {
		info := infos[name]
		eligible := len(effect.Hazards) == 0
		if closed && !info.ClosedReachable {
			eligible = false
		}
		free[name] = eligible
	}
	changed := true
	for changed {
		changed = false
		for name, effect := range functions {
			if !free[name] {
				continue
			}
			helpers := infos[name].UniversalHelpers
			if closed {
				helpers = infos[name].ClosedHelpers
			}
			for _, dependency := range effect.Dependencies {
				if helpers[dependency] {
					continue
				}
				if !free[dependency] {
					free[name] = false
					changed = true
					break
				}
			}
		}
	}
	return free
}

func safeHelpers(blockers []blockerObservation) map[string]bool {
	seen := make(map[string]bool)
	safe := make(map[string]bool)
	for _, blocker := range blockers {
		if !seen[blocker.Helper] {
			seen[blocker.Helper] = true
			safe[blocker.Helper] = true
		}
		if !blocker.Safe {
			safe[blocker.Helper] = false
		}
	}
	return safe
}

func mergeBlockerReports(fset *token.FileSet, effect *functionEffect, universal, closed []blockerObservation) []primitiveBlocker {
	type pair struct {
		universal *blockerObservation
		closed    *blockerObservation
	}
	byKey := make(map[string]*pair)
	add := func(observation blockerObservation, isClosed bool) {
		position := fset.Position(observation.Pos.Pos())
		key := observation.Helper + ":" + position.Filename + ":" + strconv.Itoa(position.Line) + ":" + strconv.Itoa(position.Column)
		entry := byKey[key]
		if entry == nil {
			entry = &pair{}
			byKey[key] = entry
		}
		copy := observation
		if isClosed {
			entry.closed = &copy
		} else {
			entry.universal = &copy
		}
	}
	for _, blocker := range universal {
		add(blocker, false)
	}
	for _, blocker := range closed {
		add(blocker, true)
	}
	var reports []primitiveBlocker
	for _, entry := range byKey {
		observation := entry.universal
		if observation == nil {
			observation = entry.closed
		}
		position := fset.Position(observation.Pos.Pos())
		report := primitiveBlocker{Kind: observation.Kind, Helper: observation.Helper, File: filepath.Base(position.Filename), Line: position.Line}
		if entry.universal != nil {
			report.UniversalSafe = entry.universal.Safe
			report.UniversalReason = entry.universal.Reason
		}
		if entry.closed != nil {
			report.ClosedDirectSafe = entry.closed.Safe
			report.ClosedReason = entry.closed.Reason
		}
		reports = append(reports, report)
	}
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].File != reports[j].File {
			return reports[i].File < reports[j].File
		}
		if reports[i].Line != reports[j].Line {
			return reports[i].Line < reports[j].Line
		}
		return reports[i].Helper < reports[j].Helper
	})
	return reports
}

func mergeRelationalReports(fset *token.FileSet, universal, closed []relationalObservation) []relationalBlocker {
	type pair struct {
		universal *relationalObservation
		closed    *relationalObservation
	}
	byKey := make(map[string]*pair)
	add := func(observation relationalObservation, isClosed bool) {
		position := fset.Position(observation.Pos.Pos())
		key := observation.Kind + ":" + position.Filename + ":" + strconv.Itoa(position.Line) + ":" + strconv.Itoa(position.Column)
		entry := byKey[key]
		if entry == nil {
			entry = &pair{}
			byKey[key] = entry
		}
		copy := observation
		if isClosed {
			entry.closed = &copy
		} else {
			entry.universal = &copy
		}
	}
	for _, observation := range universal {
		add(observation, false)
	}
	for _, observation := range closed {
		add(observation, true)
	}
	reports := make([]relationalBlocker, 0, len(byKey))
	for _, entry := range byKey {
		observation := entry.universal
		if observation == nil {
			observation = entry.closed
		}
		position := fset.Position(observation.Pos.Pos())
		report := relationalBlocker{
			Kind: observation.Kind, File: filepath.Base(position.Filename), Line: position.Line,
			IndexExpression: observation.Index, LengthExpression: observation.Length,
		}
		if entry.universal != nil {
			report.UniversalSafe = entry.universal.Safe
			report.UniversalReason = entry.universal.Reason
		}
		if entry.closed != nil {
			report.ClosedDirectSafe = entry.closed.Safe
			report.ClosedReason = entry.closed.Reason
		}
		reports = append(reports, report)
	}
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].File != reports[j].File {
			return reports[i].File < reports[j].File
		}
		if reports[i].Line != reports[j].Line {
			return reports[i].Line < reports[j].Line
		}
		return reports[i].Kind < reports[j].Kind
	})
	return reports
}
