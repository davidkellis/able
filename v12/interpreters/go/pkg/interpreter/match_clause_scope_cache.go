package interpreter

import "able/interpreter-go/pkg/ast"

type clauseScopePlan struct {
	needsLocalScope       bool
	capturePatternBinding bool
	patternBindingCount   int
	localBindingCapacity  int
	transientScopeEnvOK   bool
	transientBindingSetOK bool
	transientSingleBindOK bool
}

func buildClauseScopePlans(i *Interpreter, clauses []*ast.MatchClause) []clauseScopePlan {
	if len(clauses) == 0 {
		return nil
	}
	plans := make([]clauseScopePlan, len(clauses))
	for idx, clause := range clauses {
		if clause == nil {
			continue
		}
		if i == nil {
			plans[idx] = clauseLocalScopePlan(clause.Pattern, clause.Guard, clause.Body)
			continue
		}
		plans[idx] = i.matchClauseScopePlan(clause)
	}
	return plans
}

func (i *Interpreter) matchClauseScopePlan(clause *ast.MatchClause) clauseScopePlan {
	if i == nil || clause == nil {
		return clauseScopePlan{}
	}
	if i.envSingleThread {
		if plan, ok := i.matchClauseScopePlanCache[clause]; ok {
			return plan
		}
		plan := clauseLocalScopePlan(clause.Pattern, clause.Guard, clause.Body)
		if i.matchClauseScopePlanCache == nil {
			i.matchClauseScopePlanCache = make(map[*ast.MatchClause]clauseScopePlan)
		}
		i.matchClauseScopePlanCache[clause] = plan
		return plan
	}
	i.matchClauseScopePlanCacheMu.RLock()
	plan, ok := i.matchClauseScopePlanCache[clause]
	i.matchClauseScopePlanCacheMu.RUnlock()
	if ok {
		return plan
	}
	plan = clauseLocalScopePlan(clause.Pattern, clause.Guard, clause.Body)
	i.matchClauseScopePlanCacheMu.Lock()
	if cached, ok := i.matchClauseScopePlanCache[clause]; ok {
		i.matchClauseScopePlanCacheMu.Unlock()
		return cached
	}
	if i.matchClauseScopePlanCache == nil {
		i.matchClauseScopePlanCache = make(map[*ast.MatchClause]clauseScopePlan)
	}
	i.matchClauseScopePlanCache[clause] = plan
	i.matchClauseScopePlanCacheMu.Unlock()
	return plan
}

func (i *Interpreter) matchExpressionClausePlans(expr *ast.MatchExpression) []clauseScopePlan {
	if i == nil || expr == nil {
		return buildClauseScopePlans(i, nil)
	}
	if i.envSingleThread {
		if plans, ok := i.matchExpressionClausePlansCache[expr]; ok {
			return plans
		}
		plans := buildClauseScopePlans(i, expr.Clauses)
		i.matchExpressionClausePlansCache[expr] = plans
		return plans
	}
	i.matchExpressionClausePlansCacheMu.RLock()
	plans, ok := i.matchExpressionClausePlansCache[expr]
	i.matchExpressionClausePlansCacheMu.RUnlock()
	if ok {
		return plans
	}
	plans = buildClauseScopePlans(i, expr.Clauses)
	i.matchExpressionClausePlansCacheMu.Lock()
	if cached, ok := i.matchExpressionClausePlansCache[expr]; ok {
		i.matchExpressionClausePlansCacheMu.Unlock()
		return cached
	}
	i.matchExpressionClausePlansCache[expr] = plans
	i.matchExpressionClausePlansCacheMu.Unlock()
	return plans
}

func (i *Interpreter) rescueExpressionClausePlans(expr *ast.RescueExpression) []clauseScopePlan {
	if i == nil || expr == nil {
		return buildClauseScopePlans(i, nil)
	}
	if i.envSingleThread {
		if plans, ok := i.rescueExpressionClausePlansCache[expr]; ok {
			return plans
		}
		plans := buildClauseScopePlans(i, expr.Clauses)
		i.rescueExpressionClausePlansCache[expr] = plans
		return plans
	}
	i.rescueExpressionClausePlansCacheMu.RLock()
	plans, ok := i.rescueExpressionClausePlansCache[expr]
	i.rescueExpressionClausePlansCacheMu.RUnlock()
	if ok {
		return plans
	}
	plans = buildClauseScopePlans(i, expr.Clauses)
	i.rescueExpressionClausePlansCacheMu.Lock()
	if cached, ok := i.rescueExpressionClausePlansCache[expr]; ok {
		i.rescueExpressionClausePlansCacheMu.Unlock()
		return cached
	}
	i.rescueExpressionClausePlansCache[expr] = plans
	i.rescueExpressionClausePlansCacheMu.Unlock()
	return plans
}
