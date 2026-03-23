package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) inferLoopExpressionResultType(ctx *compileContext, loop *ast.LoopExpression, expected string) string {
	if g == nil || ctx == nil || loop == nil || loop.Body == nil {
		return "runtime.Value"
	}
	if g.hasExplicitStaticExpectedType(expected) {
		return expected
	}
	probeCtx := ctx.probeChild()
	if probeCtx == nil {
		return "runtime.Value"
	}
	probe := &controlFlowResultProbe{}
	probeCtx.loopDepth++
	probeCtx.loopLabel = "__able_probe_loop"
	probeCtx.loopBreakValueTemp = "__able_probe_loop_result"
	probeCtx.loopBreakValueType = "runtime.Value"
	probeCtx.loopBreakProbe = probe
	if _, ok := g.compileBlockStatement(probeCtx, loop.Body); !ok {
		return "runtime.Value"
	}
	if resultType, ok := g.resolveControlFlowResultType(ctx, "", nil, nil, false, probe); ok {
		return resultType
	}
	return "runtime.Value"
}

func (g *generator) inferBreakpointExpressionResultType(ctx *compileContext, expr *ast.BreakpointExpression, expected string) string {
	if g == nil || ctx == nil || expr == nil || expr.Body == nil || expr.Label == nil || expr.Label.Name == "" {
		return "runtime.Value"
	}
	if g.hasExplicitStaticExpectedType(expected) {
		return expected
	}
	probeCtx := ctx.probeChild()
	if probeCtx == nil {
		return "runtime.Value"
	}
	label := expr.Label.Name
	probe := &controlFlowResultProbe{}
	probeCtx.pushBreakpoint(label)
	if probeCtx.breakpointGoLabels == nil {
		probeCtx.breakpointGoLabels = make(map[string]string)
	}
	if probeCtx.breakpointResultTemps == nil {
		probeCtx.breakpointResultTemps = make(map[string]string)
	}
	if probeCtx.breakpointResultTypes == nil {
		probeCtx.breakpointResultTypes = make(map[string]string)
	}
	if probeCtx.breakpointResultProbes == nil {
		probeCtx.breakpointResultProbes = make(map[string]*controlFlowResultProbe)
	}
	probeCtx.breakpointGoLabels[label] = "__able_probe_breakpoint"
	probeCtx.breakpointResultTemps[label] = "__able_probe_breakpoint_result"
	probeCtx.breakpointResultTypes[label] = "runtime.Value"
	probeCtx.breakpointResultProbes[label] = probe
	normalType := "struct{}"
	var normalExpr ast.Expression
	var normalTypeExpr ast.TypeExpression
	stmts := expr.Body.Body
	for idx, stmt := range stmts {
		isLast := idx == len(stmts)-1
		if isLast {
			if tailExpr, ok := stmt.(ast.Expression); ok {
				_, _, tailType, ok := g.compileTailExpression(probeCtx, "", tailExpr)
				if !ok {
					return "runtime.Value"
				}
				normalType = tailType
				normalExpr = tailExpr
				normalTypeExpr, _ = g.inferExpressionTypeExpr(probeCtx, tailExpr, tailType)
				break
			}
		}
		if _, ok := g.compileStatement(probeCtx, stmt); !ok {
			return "runtime.Value"
		}
	}
	probeCtx.popBreakpoint(label)
	if resultType, ok := g.resolveControlFlowResultType(ctx, normalType, normalExpr, normalTypeExpr, true, probe); ok {
		return resultType
	}
	return "runtime.Value"
}

func (g *generator) hasExplicitStaticExpectedType(expected string) bool {
	return expected != "" && expected != "runtime.Value" && expected != "any"
}

func (g *generator) resolveControlFlowResultType(ctx *compileContext, normalType string, normalExpr ast.Expression, normalTypeExpr ast.TypeExpression, hasNormal bool, probe *controlFlowResultProbe) (string, bool) {
	if g == nil {
		return "", false
	}
	branches := make([]joinBranchInfo, 0, 1+len(probe.branchTypes))
	sawVoid := false
	addType := func(goType string) bool {
		switch {
		case goType == "":
			return false
		case g.isVoidType(goType):
			sawVoid = true
			return true
		}
		return true
	}
	if hasNormal && !addType(normalType) {
		return "", false
	}
	if hasNormal && !g.isVoidType(normalType) {
		branches = append(branches, joinBranchInfo{GoType: normalType, Expr: normalExpr, TypeExpr: normalTypeExpr})
	}
	if probe != nil {
		for idx, goType := range probe.branchTypes {
			if !addType(goType) {
				return "", false
			}
			if !g.isVoidType(goType) {
				var typeExpr ast.TypeExpression
				if idx < len(probe.branchTypeExprs) {
					typeExpr = probe.branchTypeExprs[idx]
				}
				branches = append(branches, joinBranchInfo{GoType: goType, TypeExpr: typeExpr})
			}
		}
	}
	if sawVoid {
		if len(branches) == 0 && (probe == nil || !probe.sawNil) {
			return "struct{}", true
		}
		return "", false
	}
	if probe != nil && probe.sawNil {
		branches = append(branches, joinBranchInfo{SawNil: true})
	}
	return g.lowerJoinCarrierFromBranches(ctx, branches)
}

func (g *generator) controlFlowResultExpr(ctx *compileContext, resultType string, expr string, exprType string) ([]string, string, bool) {
	if resultType == "" {
		return nil, "", false
	}
	if expr == "" || exprType == "" {
		return nil, "", false
	}
	return g.coerceJoinBranch(ctx, resultType, expr, exprType)
}

func (g *generator) controlFlowNilResultExpr(resultType string) (string, bool) {
	if resultType == "" {
		return "", false
	}
	if resultType == "runtime.Value" {
		return "runtime.NilValue{}", true
	}
	return g.typedNilExpr(resultType)
}
