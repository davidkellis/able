package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
)

type rangeEnv map[string]intRange

type rangeRun struct {
	fset               *token.FileSet
	functions          map[string]*functionEffect
	effect             *functionEffect
	env                rangeEnv
	blockers           []blockerObservation
	relations          []relationalObservation
	calls              []rangeCall
	returned           intRange
	returnedSet        bool
	useClosedSummaries bool
	returnFacts        map[string]intRange
	returnFactsSeen    bool
	aliases            map[string]string
	recognized         map[token.Pos]bool
}

func newRangeRun(fset *token.FileSet, functions map[string]*functionEffect, effect *functionEffect, env rangeEnv) *rangeRun {
	run := &rangeRun{
		fset: fset, functions: functions, effect: effect, env: cloneRangeEnv(env),
		aliases: make(map[string]string), recognized: make(map[token.Pos]bool),
	}
	for _, param := range functionParameters(effect.decl) {
		run.aliases[param.Name] = param.Name
	}
	return run
}

func (run *rangeRun) analyze() {
	if run == nil || run.effect == nil || run.effect.decl == nil || run.effect.decl.Body == nil {
		return
	}
	run.env, _ = run.analyzeBlock(run.effect.decl.Body.List, run.env)
}

func (run *rangeRun) blockersWithMissingPrimitiveCalls() []blockerObservation {
	if run == nil || run.effect == nil || run.effect.decl == nil {
		return nil
	}
	blockers := append([]blockerObservation(nil), run.blockers...)
	ast.Inspect(run.effect.decl.Body, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok || run.recognized[call.Pos()] {
			return true
		}
		name := calledName(call.Fun)
		if !isPrimitiveRangeHelper(name) {
			return true
		}
		blockers = append(blockers, blockerObservation{
			Kind:   "unresolved-primitive-control",
			Helper: name,
			Pos:    call,
			Safe:   false,
			Reason: "primitive control call was not in a supported generated shape",
		})
		return true
	})
	return blockers
}

func (run *rangeRun) analyzeBlock(statements []ast.Stmt, env rangeEnv) (rangeEnv, bool) {
	current := cloneRangeEnv(env)
	for _, statement := range statements {
		var falls bool
		current, falls = run.analyzeStmt(statement, current)
		if !falls {
			return current, false
		}
	}
	return current, true
}

func (run *rangeRun) analyzeStmt(statement ast.Stmt, env rangeEnv) (rangeEnv, bool) {
	switch typed := statement.(type) {
	case *ast.AssignStmt:
		run.analyzeAssignment(typed, env)
		return env, true
	case *ast.DeclStmt:
		run.analyzeDeclaration(typed, env)
		return env, true
	case *ast.ExprStmt:
		run.evalExpr(typed.X, env)
		return env, true
	case *ast.ReturnStmt:
		successReturn := len(typed.Results) < 2 || identName(typed.Results[len(typed.Results)-1]) == "nil"
		for idx, result := range typed.Results {
			value := run.evalExpr(result, env)
			if idx == 0 && successReturn {
				if !run.returnedSet {
					run.returned = value
				} else {
					run.returned = unionRange(run.returned, value)
				}
				run.returnedSet = true
				run.observeReturnFacts(result, env)
			}
		}
		return env, false
	case *ast.BranchStmt:
		return env, typed.Tok != token.BREAK && typed.Tok != token.CONTINUE && typed.Tok != token.GOTO
	case *ast.IncDecStmt:
		name := identName(typed.X)
		value := env[name]
		if value.Known {
			delta := knownRange(1, 1)
			if typed.Tok == token.INC {
				env[name] = addRange(value, delta)
			} else {
				env[name] = subtractRange(value, delta)
			}
		}
		return env, true
	case *ast.IfStmt:
		return run.analyzeIf(typed, env)
	case *ast.ForStmt:
		return run.analyzeFor(typed, env)
	case *ast.RangeStmt:
		run.evalExpr(typed.X, env)
		bodyEnv := cloneRangeEnv(env)
		run.analyzeBlock(typed.Body.List, bodyEnv)
		return env, true
	case *ast.BlockStmt:
		return run.analyzeBlock(typed.List, env)
	case *ast.LabeledStmt:
		return run.analyzeStmt(typed.Stmt, env)
	case *ast.SwitchStmt:
		if typed.Init != nil {
			env, _ = run.analyzeStmt(typed.Init, env)
		}
		run.evalExpr(typed.Tag, env)
		return run.analyzeCaseClauses(typed.Body, env), true
	default:
		return env, true
	}
}

func (run *rangeRun) analyzeIf(statement *ast.IfStmt, env rangeEnv) (rangeEnv, bool) {
	current := cloneRangeEnv(env)
	if statement.Init != nil {
		current, _ = run.analyzeStmt(statement.Init, current)
	}
	run.evalExpr(statement.Cond, current)
	if observation, ok := generatedBoundsObservation(statement, current, run.evalExpr); ok {
		run.relations = append(run.relations, observation)
	}
	if blocker, _, _, ok := run.inlineOverflowBlocker(statement, current); ok {
		run.blockers = append(run.blockers, blocker)
	}
	trueEnv := refineRangeEnv(current, statement.Cond, true)
	falseEnv := refineRangeEnv(current, statement.Cond, false)
	thenEnv, thenFalls := run.analyzeBlock(statement.Body.List, trueEnv)
	if statement.Else == nil {
		if !thenFalls {
			return falseEnv, true
		}
		return mergeRangeEnvs(thenEnv, falseEnv), true
	}
	elseEnv, elseFalls := run.analyzeStmt(statement.Else, falseEnv)
	switch {
	case thenFalls && elseFalls:
		return mergeRangeEnvs(thenEnv, elseEnv), true
	case thenFalls:
		return thenEnv, true
	case elseFalls:
		return elseEnv, true
	default:
		return current, false
	}
}

func (run *rangeRun) analyzeFor(statement *ast.ForStmt, env rangeEnv) (rangeEnv, bool) {
	current := cloneRangeEnv(env)
	if statement.Init != nil {
		current, _ = run.analyzeStmt(statement.Init, current)
	}
	bodyEnv := cloneRangeEnv(current)
	inductionName := ""
	if variable, values, ok := loopInductionRange(statement, current); ok {
		inductionName = variable
		bodyEnv[variable] = values
	}
	invalidateLoopCarriedRanges(statement.Body, bodyEnv, inductionName)
	if statement.Cond != nil {
		run.evalExpr(statement.Cond, current)
		bodyEnv = refineRangeEnv(bodyEnv, statement.Cond, true)
	}
	bodyResult, _ := run.analyzeBlock(statement.Body.List, bodyEnv)
	if statement.Post != nil {
		run.analyzeStmt(statement.Post, bodyResult)
	}
	if statement.Cond != nil {
		result := refineRangeEnv(current, statement.Cond, false)
		mergeLoopAggregateFacts(result, current, bodyResult, statement, inductionName)
		return result, true
	}
	return current, true
}

func invalidateLoopCarriedRanges(body *ast.BlockStmt, env rangeEnv, inductionName string) {
	if body == nil || len(env) == 0 {
		return
	}
	ast.Inspect(body, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}
		switch typed := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range typed.Lhs {
				name := identName(lhs)
				if name != "" && name != inductionName {
					if _, tracked := env[name]; tracked {
						env[name] = intRange{}
					}
				}
			}
		case *ast.IncDecStmt:
			name := identName(typed.X)
			if name != "" && name != inductionName {
				if _, tracked := env[name]; tracked {
					env[name] = intRange{}
				}
			}
		}
		return true
	})
}

func loopInductionRange(statement *ast.ForStmt, env rangeEnv) (string, intRange, bool) {
	if statement == nil {
		return "", intRange{}, false
	}
	condition := statement.Cond
	if condition == nil && statement.Body != nil && len(statement.Body.List) > 0 {
		if guard, ok := statement.Body.List[0].(*ast.IfStmt); ok && guard.Body != nil && blockHasBreak(guard.Body) {
			condition = inverseLoopGuard(guard.Cond)
		}
	}
	binary, ok := condition.(*ast.BinaryExpr)
	if !ok {
		return "", intRange{}, false
	}
	name := identName(binary.X)
	bound, ok := integerLiteral(binary.Y)
	if !ok {
		boundRange := env[identName(binary.Y)]
		if boundRange.Known && boundRange.Min == boundRange.Max {
			bound = boundRange.Min
			ok = true
		}
	}
	if !ok || name == "" {
		return "", intRange{}, false
	}
	current := env[name]
	if !current.Known {
		return "", intRange{}, false
	}
	maxValue := int64(0)
	switch binary.Op {
	case token.LSS:
		if bound == math.MinInt64 {
			return "", intRange{}, false
		}
		maxValue = bound - 1
	case token.LEQ:
		maxValue = bound
	default:
		return "", intRange{}, false
	}
	minValue := int64(math.MinInt64)
	if loopHasSinglePositiveUpdate(statement, name) {
		minValue = current.Min
	}
	if minValue > maxValue {
		return "", intRange{}, false
	}
	return name, knownRange(minValue, maxValue), true
}

func loopHasSinglePositiveUpdate(statement *ast.ForStmt, name string) bool {
	if statement == nil || name == "" {
		return false
	}
	writes := 0
	positive := false
	observe := func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.IncDecStmt:
			if identName(typed.X) == name {
				writes++
				positive = typed.Tok == token.INC
			}
		case *ast.AssignStmt:
			for _, lhs := range typed.Lhs {
				if identName(lhs) == name {
					writes++
					positive = directPositiveUpdate(name, typed)
				}
			}
		}
		return true
	}
	if statement.Body != nil {
		ast.Inspect(statement.Body, observe)
	}
	if statement.Post != nil {
		ast.Inspect(statement.Post, observe)
	}
	return writes == 1 && positive
}

func directPositiveUpdate(name string, statement *ast.AssignStmt) bool {
	if statement == nil || len(statement.Lhs) != 1 || len(statement.Rhs) != 1 || identName(statement.Lhs[0]) != name {
		return false
	}
	binary, ok := statement.Rhs[0].(*ast.BinaryExpr)
	if !ok || binary.Op != token.ADD || identName(binary.X) != name {
		return false
	}
	increment, ok := integerLiteral(binary.Y)
	return ok && increment > 0
}

func inverseLoopGuard(expr ast.Expr) ast.Expr {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return expr
	}
	copy := *binary
	copy.Op = inverseComparison(binary.Op)
	return &copy
}

func blockHasBreak(block *ast.BlockStmt) bool {
	if block == nil {
		return false
	}
	for _, statement := range block.List {
		branch, ok := statement.(*ast.BranchStmt)
		if ok && branch.Tok == token.BREAK {
			return true
		}
	}
	return false
}

func (run *rangeRun) analyzeCaseClauses(body *ast.BlockStmt, env rangeEnv) rangeEnv {
	if body == nil {
		return env
	}
	var merged rangeEnv
	for _, statement := range body.List {
		clause, ok := statement.(*ast.CaseClause)
		if !ok {
			continue
		}
		for _, expr := range clause.List {
			run.evalExpr(expr, env)
		}
		result, _ := run.analyzeBlock(clause.Body, cloneRangeEnv(env))
		if merged == nil {
			merged = result
		} else {
			merged = mergeRangeEnvs(merged, result)
		}
	}
	if merged == nil {
		return env
	}
	return merged
}

func (run *rangeRun) analyzeDeclaration(statement *ast.DeclStmt, env rangeEnv) {
	declaration, ok := statement.Decl.(*ast.GenDecl)
	if !ok {
		return
	}
	for _, spec := range declaration.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for idx, name := range valueSpec.Names {
			if idx < len(valueSpec.Values) {
				env[name.Name] = run.evalExpr(valueSpec.Values[idx], env)
				run.initializeAggregateTarget(name.Name, valueSpec.Values[idx], env)
				if call, ok := unparen(valueSpec.Values[idx]).(*ast.CallExpr); ok {
					run.assignCallReturnFacts(name.Name, call, env)
				}
			} else {
				env[name.Name] = intRange{}
			}
		}
	}
}

func (run *rangeRun) analyzeAssignment(statement *ast.AssignStmt, env rangeEnv) {
	run.assignAggregate(statement, env)
	if len(statement.Rhs) == 1 {
		if call, ok := statement.Rhs[0].(*ast.CallExpr); ok {
			if result, handled := run.analyzePrimitiveCall(call, env); handled {
				if len(statement.Lhs) > 0 {
					if name := identName(statement.Lhs[0]); name != "" && name != "_" {
						env[name] = result
					}
				}
				for idx := 1; idx < len(statement.Lhs); idx++ {
					if name := identName(statement.Lhs[idx]); name != "" && name != "_" {
						env[name] = intRange{}
					}
				}
				return
			}
			run.applyCallAggregateEffects(call, env)
			if len(statement.Lhs) > 0 {
				run.assignCallReturnFacts(identName(statement.Lhs[0]), call, env)
			}
		}
	}
	values := make([]intRange, len(statement.Rhs))
	for idx, expr := range statement.Rhs {
		values[idx] = run.evalExpr(expr, env)
	}
	if len(statement.Rhs) == 1 && len(statement.Lhs) > 1 {
		values = make([]intRange, len(statement.Lhs))
	}
	for idx, expr := range statement.Lhs {
		name := identName(expr)
		if name == "" || name == "_" {
			continue
		}
		if idx < len(values) {
			env[name] = values[idx]
		} else {
			env[name] = intRange{}
		}
	}
}

func (run *rangeRun) analyzePrimitiveCall(call *ast.CallExpr, env rangeEnv) (intRange, bool) {
	name := calledName(call.Fun)
	if !isPrimitiveRangeHelper(name) || name == "__able_raise_overflow" {
		return intRange{}, false
	}
	args := make([]intRange, len(call.Args))
	for idx, arg := range call.Args {
		args[idx] = run.evalExpr(arg, env)
	}
	result, safe, reason := evaluatePrimitiveHelper(name, call.Args, args)
	run.recognized[call.Pos()] = true
	run.blockers = append(run.blockers, blockerObservation{Kind: primitiveKind(name), Helper: name, Pos: call, Safe: safe, Reason: reason})
	return result, true
}

func (run *rangeRun) inlineOverflowBlocker(statement *ast.IfStmt, env rangeEnv) (blockerObservation, string, intRange, bool) {
	if statement == nil || statement.Body == nil {
		return blockerObservation{}, "", intRange{}, false
	}
	var raiseCall *ast.CallExpr
	ast.Inspect(statement.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok && calledName(call.Fun) == "__able_raise_overflow" {
			raiseCall = call
			return false
		}
		return true
	})
	if raiseCall == nil {
		return blockerObservation{}, "", intRange{}, false
	}
	variable, lower, upper, ok := overflowConditionBounds(statement.Cond)
	if !ok {
		return blockerObservation{}, "", intRange{}, false
	}
	value := env[variable]
	safe := value.Known && value.Min >= lower && value.Max <= upper
	reason := fmt.Sprintf("%s must remain within [%d,%d]", describeRange(value), lower, upper)
	run.recognized[raiseCall.Pos()] = true
	return blockerObservation{Kind: "inline-signed-bounds", Helper: "__able_raise_overflow", Pos: raiseCall, Safe: safe, Reason: reason}, variable, knownRange(lower, upper), true
}

func (run *rangeRun) evalExpr(expr ast.Expr, env rangeEnv) intRange {
	if expr == nil {
		return intRange{}
	}
	if value, ok := integerLiteral(expr); ok {
		return knownRange(value, value)
	}
	switch typed := expr.(type) {
	case *ast.Ident:
		return env[typed.Name]
	case *ast.ParenExpr:
		return run.evalExpr(typed.X, env)
	case *ast.UnaryExpr:
		value := run.evalExpr(typed.X, env)
		if typed.Op == token.ADD {
			return value
		}
		if typed.Op == token.SUB && value.Known && value.Min != math.MinInt64 && value.Max != math.MinInt64 {
			return knownRange(-value.Max, -value.Min)
		}
		return intRange{}
	case *ast.BinaryExpr:
		left := run.evalExpr(typed.X, env)
		right := run.evalExpr(typed.Y, env)
		switch typed.Op {
		case token.ADD:
			return addRange(left, right)
		case token.SUB:
			return subtractRange(left, right)
		case token.MUL:
			return multiplyRange(left, right)
		case token.QUO:
			if left.Known && right.Known && left.Min >= 0 && right.Min > 0 {
				return knownRange(left.Min/right.Max, left.Max/right.Min)
			}
		case token.AND:
			if left.Known && right.Known && left.Min >= 0 && right.Min >= 0 {
				maxValue := left.Max
				if right.Max < maxValue {
					maxValue = right.Max
				}
				return knownRange(0, maxValue)
			}
		case token.OR, token.XOR:
			if left.Known && right.Known && left.Min >= 0 && right.Min >= 0 {
				maxValue, ok := nonnegativeBitwiseUpper(left.Max, right.Max)
				if ok {
					return knownRange(0, maxValue)
				}
			}
		}
		return intRange{}
	case *ast.CallExpr:
		name := calledName(typed.Fun)
		if name == "__able_slice_len" && len(typed.Args) == 1 {
			if root := run.aggregateRoot(typed.Args[0]); root != "" {
				return env[root+aggregateLength]
			}
			return intRange{}
		}
		if (name == "uint64" || name == "uint") && len(typed.Args) == 1 {
			value := run.evalExpr(typed.Args[0], env)
			if value.Known && value.Min >= 0 {
				return value
			}
			return intRange{}
		}
		if full, ok := fullIntegerRange(name); ok && len(typed.Args) == 1 {
			value := run.evalExpr(typed.Args[0], env)
			if value.Known && value.Min >= full.Min && value.Max <= full.Max {
				return value
			}
			return full
		}
		args := make([]intRange, len(typed.Args))
		for idx, arg := range typed.Args {
			args[idx] = run.evalExpr(arg, env)
		}
		if callee := run.functions[name]; callee != nil && callee.Direct {
			run.calls = append(run.calls, rangeCall{
				Callee: name, Args: args, AggregateFacts: run.aggregateCallFacts(typed, env),
			})
			run.applyCallAggregateEffects(typed, env)
			if run.useClosedSummaries && callee.closedReturnSet {
				return callee.closedReturnRange
			}
		}
		run.invalidateUnknownCallFacts(typed, env)
		return intRange{}
	case *ast.IndexExpr:
		run.evalExpr(typed.X, env)
		run.evalExpr(typed.Index, env)
		if root := run.aggregateRoot(typed); root != "" {
			return env[root]
		}
	case *ast.SelectorExpr:
		run.evalExpr(typed.X, env)
		if root := run.aggregateRoot(typed.X); root != "" && typed.Sel.Name != "Elements" {
			return env[root+"."+typed.Sel.Name]
		}
	}
	return intRange{}
}

func overflowConditionBounds(expr ast.Expr) (string, int64, int64, bool) {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != token.LOR {
		return "", 0, 0, false
	}
	leftName, lower, leftOK := comparisonBound(binary.X, token.LSS)
	rightName, upper, rightOK := comparisonBound(binary.Y, token.GTR)
	if !leftOK || !rightOK || leftName != rightName {
		return "", 0, 0, false
	}
	return leftName, lower, upper, true
}

func comparisonBound(expr ast.Expr, operator token.Token) (string, int64, bool) {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != operator {
		return "", 0, false
	}
	name := identName(binary.X)
	value, ok := integerLiteral(binary.Y)
	return name, value, ok && name != ""
}

func cloneRangeEnv(source rangeEnv) rangeEnv {
	copy := make(rangeEnv, len(source))
	for name, value := range source {
		copy[name] = value
	}
	return copy
}

func mergeRangeEnvs(left, right rangeEnv) rangeEnv {
	merged := make(rangeEnv)
	for name, leftValue := range left {
		rightValue, ok := right[name]
		if !ok {
			continue
		}
		merged[name] = unionRange(leftValue, rightValue)
	}
	return merged
}
