package compiler

import (
	"math"

	"able/interpreter-go/pkg/ast"
)

const boundedRecursiveReturnFactMaxParam = 4096

type recursiveReturnTerm struct {
	SelfCall  bool
	Decrement int64
	Constant  int64
}

type integerReturnRangeFact struct {
	ParamName   string
	ParamGoName string
	MaxParam    int64
	BaseMax     int64
	MaxByParam  []int64
}

func (fact *integerReturnRangeFact) maxReturnForParamMax(maxParam int64) (int64, bool) {
	if fact == nil {
		return 0, false
	}
	if maxParam < 0 {
		return fact.BaseMax, true
	}
	if maxParam > fact.MaxParam || int(maxParam) >= len(fact.MaxByParam) {
		return 0, false
	}
	return fact.MaxByParam[maxParam], true
}

func (g *generator) resolveStaticFunctionIntegerReturnFacts() {
	if g == nil {
		return
	}
	for _, info := range g.allFunctionInfos() {
		if info == nil {
			continue
		}
		info.ReturnFact = integerFact{}
		info.HasReturnFact = false
		info.ReturnRange = nil
	}
	for _, info := range g.sortedFunctionInfos() {
		fact, returnRange, ok := g.inferBoundedSelfRecursiveReturnFact(info)
		if !ok || !fact.hasUsefulFact() {
			continue
		}
		info.ReturnFact = fact
		info.HasReturnFact = true
		info.ReturnRange = returnRange
	}
}

func (g *generator) inferBoundedSelfRecursiveReturnFact(info *functionInfo) (integerFact, *integerReturnRangeFact, bool) {
	if g == nil || info == nil || !info.Compileable || info.Definition == nil || info.Definition.Body == nil {
		return integerFact{}, nil, false
	}
	if len(info.Params) != 1 || !g.isSignedIntegerType(info.ReturnType) {
		return integerFact{}, nil, false
	}
	upperBound, ok := g.signedIntegerUpperBound(info.ReturnType)
	if !ok {
		return integerFact{}, nil, false
	}
	param := info.Params[0]
	if param.Name == "" || !g.isSignedIntegerType(param.GoType) {
		return integerFact{}, nil, false
	}
	paramFact, ok := g.observedNonSelfParamFact(info, param.Name)
	if !ok {
		paramFact, ok = info.ParamFacts[param.Name]
	}
	if !ok || !paramFact.HasMax || paramFact.MaxInclusive < 0 || paramFact.MaxInclusive > boundedRecursiveReturnFactMaxParam {
		return integerFact{}, nil, false
	}
	statements := info.Definition.Body.Body
	if len(statements) < 2 {
		return integerFact{}, nil, false
	}
	finalExpr, ok := finalReturnExpression(statements[len(statements)-1])
	if !ok {
		return integerFact{}, nil, false
	}
	baseMax, threshold, ok := boundedRecursiveBaseCase(statements[:len(statements)-1], param.Name)
	if !ok || baseMax < 0 || baseMax > upperBound || threshold < 0 || threshold > paramFact.MaxInclusive {
		return integerFact{}, nil, false
	}
	ctx := g.compileBodyContext(info)
	terms, selfCalls, ok := g.boundedRecursiveReturnTerms(ctx, info, param.Name, finalExpr)
	if !ok || selfCalls == 0 {
		return integerFact{}, nil, false
	}
	values := make([]int64, int(paramFact.MaxInclusive)+1)
	maxResult := baseMax
	for n := int64(0); n <= paramFact.MaxInclusive; n++ {
		if n <= threshold {
			values[n] = baseMax
			continue
		}
		var result int64
		for _, term := range terms {
			if !term.SelfCall {
				next, ok := addInt64NoOverflow(result, term.Constant)
				if !ok || next > upperBound {
					return integerFact{}, nil, false
				}
				result = next
				continue
			}
			arg := n - term.Decrement
			if arg < 0 || arg >= n || arg > paramFact.MaxInclusive {
				return integerFact{}, nil, false
			}
			next, ok := addInt64NoOverflow(result, values[arg])
			if !ok || next > upperBound {
				return integerFact{}, nil, false
			}
			result = next
		}
		values[n] = result
		if result > maxResult {
			maxResult = result
		}
	}
	maxByParam := make([]int64, len(values))
	prefixMax := baseMax
	for idx, value := range values {
		if value > prefixMax {
			prefixMax = value
		}
		maxByParam[idx] = prefixMax
	}
	returnRange := &integerReturnRangeFact{
		ParamName:   param.Name,
		ParamGoName: param.GoName,
		MaxParam:    paramFact.MaxInclusive,
		BaseMax:     baseMax,
		MaxByParam:  maxByParam,
	}
	return integerFact{NonNegative: true, HasMax: true, MaxInclusive: maxResult}, returnRange, true
}

func (g *generator) observedNonSelfParamFact(target *functionInfo, paramName string) (integerFact, bool) {
	if g == nil || target == nil || paramName == "" {
		return integerFact{}, false
	}
	observations := make(map[*functionInfo]map[string]paramFactObservation)
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || info == target || !info.Compileable || info.Definition == nil || info.Definition.Body == nil {
			continue
		}
		ctx := g.compileBodyContext(info)
		g.collectIntegerFactsFromStatements(ctx, info.Definition.Body.Body, observations)
	}
	paramObs, ok := observations[target][paramName]
	if !ok || !paramObs.Seen || !paramObs.Valid || !paramObs.Fact.hasUsefulFact() {
		return integerFact{}, false
	}
	return paramObs.Fact, true
}

func finalReturnExpression(stmt ast.Statement) (ast.Expression, bool) {
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s == nil || s.Argument == nil {
			return nil, false
		}
		return s.Argument, true
	case ast.Expression:
		return s, s != nil
	default:
		return nil, false
	}
}

func boundedRecursiveBaseCase(statements []ast.Statement, paramName string) (int64, int64, bool) {
	if len(statements) != 1 {
		return 0, 0, false
	}
	ifExpr, ok := statements[0].(*ast.IfExpression)
	if !ok || ifExpr == nil || len(ifExpr.ElseIfClauses) != 0 || ifExpr.ElseBody != nil || !blockAlwaysExits(ifExpr.IfBody) {
		return 0, 0, false
	}
	threshold, ok := paramUpperGuardLimit(ifExpr.IfCondition, paramName)
	if !ok {
		return 0, 0, false
	}
	baseMax, ok := nonNegativeLiteralReturnMax(ifExpr.IfBody)
	return baseMax, threshold, ok
}

func paramUpperGuardLimit(expr ast.Expression, paramName string) (int64, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if !ok || binary == nil || paramName == "" {
		return 0, false
	}
	if ident, ok := binary.Left.(*ast.Identifier); ok && ident != nil && ident.Name == paramName {
		value, ok := integerLiteralInt64(binary.Right)
		if !ok {
			return 0, false
		}
		switch binary.Operator {
		case "<=":
			return value, true
		case "<":
			if value == math.MinInt64 {
				return 0, false
			}
			return value - 1, true
		}
	}
	if ident, ok := binary.Right.(*ast.Identifier); ok && ident != nil && ident.Name == paramName {
		value, ok := integerLiteralInt64(binary.Left)
		if !ok {
			return 0, false
		}
		switch binary.Operator {
		case ">=":
			return value, true
		case ">":
			if value == math.MinInt64 {
				return 0, false
			}
			return value - 1, true
		}
	}
	return 0, false
}

func nonNegativeLiteralReturnMax(block *ast.BlockExpression) (int64, bool) {
	if block == nil || len(block.Body) != 1 {
		return 0, false
	}
	ret, ok := block.Body[0].(*ast.ReturnStatement)
	if !ok || ret == nil {
		return 0, false
	}
	value, ok := integerLiteralInt64(ret.Argument)
	if !ok || value < 0 {
		return 0, false
	}
	return value, true
}

func (g *generator) boundedRecursiveReturnTerms(ctx *compileContext, info *functionInfo, paramName string, expr ast.Expression) ([]recursiveReturnTerm, int, bool) {
	binary, ok := expr.(*ast.BinaryExpression)
	if ok && binary != nil && binary.Operator == "+" {
		leftTerms, leftCalls, ok := g.boundedRecursiveReturnTerms(ctx, info, paramName, binary.Left)
		if !ok {
			return nil, 0, false
		}
		rightTerms, rightCalls, ok := g.boundedRecursiveReturnTerms(ctx, info, paramName, binary.Right)
		if !ok {
			return nil, 0, false
		}
		return append(leftTerms, rightTerms...), leftCalls + rightCalls, true
	}
	if value, ok := integerLiteralInt64(expr); ok {
		if value < 0 {
			return nil, 0, false
		}
		return []recursiveReturnTerm{{Constant: value}}, 0, true
	}
	decrement, ok := g.selfRecursiveArgumentDecrement(ctx, info, paramName, expr)
	if !ok || decrement <= 0 {
		return nil, 0, false
	}
	return []recursiveReturnTerm{{SelfCall: true, Decrement: decrement}}, 1, true
}

func (g *generator) selfRecursiveArgumentDecrement(ctx *compileContext, info *functionInfo, paramName string, expr ast.Expression) (int64, bool) {
	call, ok := expr.(*ast.FunctionCall)
	if !ok || call == nil || len(call.TypeArguments) != 0 || len(call.Arguments) != 1 {
		return 0, false
	}
	callee, ok := call.Callee.(*ast.Identifier)
	if !ok || callee == nil || callee.Name == "" {
		return 0, false
	}
	resolved, overload, ok := g.resolveStaticCallable(ctx, callee.Name)
	if !ok || overload != nil || resolved != info {
		return 0, false
	}
	binary, ok := call.Arguments[0].(*ast.BinaryExpression)
	if !ok || binary == nil || binary.Operator != "-" {
		return 0, false
	}
	left, ok := binary.Left.(*ast.Identifier)
	if !ok || left == nil || left.Name != paramName {
		return 0, false
	}
	decrement, ok := positiveIntegerLiteralValue(binary.Right)
	if !ok {
		return 0, false
	}
	return decrement, true
}
