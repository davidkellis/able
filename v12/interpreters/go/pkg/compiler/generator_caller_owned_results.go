package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

const callerOwnedResultSuffix = "_into"

func callerOwnedResultVariantName(name string) string {
	if name == "" {
		return ""
	}
	return name + callerOwnedResultSuffix
}

func (g *generator) callerOwnedResultInfo(info *functionInfo) *structInfo {
	if g == nil || info == nil || g.callerOwnedResults == nil {
		return nil
	}
	if result := g.callerOwnedResults[info]; result != nil {
		return result
	}
	for candidate, result := range g.callerOwnedResults {
		if candidate != nil && result != nil && candidate.Definition == info.Definition && candidate.ReturnType == info.ReturnType {
			return result
		}
	}
	return nil
}

func (g *generator) resolveCallerOwnedResults() {
	if g == nil {
		return
	}
	g.callerOwnedResults = make(map[*functionInfo]*structInfo)
	for {
		progress := false
		candidates := g.sortedFunctionInfos()
		seen := make(map[*functionInfo]struct{}, len(candidates))
		for _, info := range candidates {
			seen[info] = struct{}{}
		}
		for _, method := range g.sortedMethodInfos() {
			if method == nil || method.Info == nil {
				continue
			}
			if _, ok := seen[method.Info]; ok {
				continue
			}
			seen[method.Info] = struct{}{}
			candidates = append(candidates, method.Info)
		}
		beforeFunctions := len(candidates)
		for _, info := range candidates {
			if info == nil || !info.Compileable || info.Definition == nil || info.ExternBody != nil {
				continue
			}
			if g.callerOwnedResults[info] != nil {
				continue
			}
			resultInfo := g.smallCallerOwnedResultStruct(info.ReturnType)
			if resultInfo == nil || statementsContainReturn(info.Definition.Body.Body) {
				continue
			}
			ctx := g.compileBodyContext(info)
			ctx.analysisOnly = true
			_, resultExpr, ok := g.compileBody(ctx, info)
			if !ok {
				continue
			}
			fresh := strings.HasPrefix(strings.TrimSpace(resultExpr), "&"+resultInfo.GoName+"{")
			if source := ctx.resultSources[resultExpr]; source != nil {
				fresh = g.callerOwnedResultInfo(source) == resultInfo
			}
			if fresh {
				g.callerOwnedResults[info] = resultInfo
				progress = true
			}
		}
		if !progress && len(candidates) == beforeFunctions {
			return
		}
	}
}

func (g *generator) smallCallerOwnedResultStruct(goType string) *structInfo {
	if g == nil || !strings.HasPrefix(goType, "*") {
		return nil
	}
	info := g.structInfoByGoName(goType)
	if info == nil || !info.Supported || info.Kind == ast.StructKindSingleton || len(info.Fields) == 0 {
		return nil
	}
	words := 0
	for _, field := range info.Fields {
		if !field.Supported {
			return nil
		}
		fieldWords, ok := callerOwnedScalarWords(field.GoType)
		if !ok {
			return nil
		}
		words += fieldWords
		if words > 2 {
			return nil
		}
	}
	return info
}

func callerOwnedScalarWords(goType string) (int, bool) {
	switch goType {
	case "bool", "byte", "rune", "int", "uint",
		"int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "runtime.Int128", "runtime.Uint128":
		if goType == "runtime.Int128" || goType == "runtime.Uint128" {
			return 2, true
		}
		return 1, true
	default:
		return 0, false
	}
}

func statementsContainReturn(statements []ast.Statement) bool {
	for _, statement := range statements {
		if statementContainsReturn(statement) {
			return true
		}
	}
	return false
}

func statementContainsReturn(statement ast.Statement) bool {
	switch node := statement.(type) {
	case *ast.ReturnStatement:
		return true
	case *ast.BlockExpression:
		return node != nil && statementsContainReturn(node.Body)
	case *ast.IfExpression:
		if node == nil {
			return false
		}
		if node.IfBody != nil && statementsContainReturn(node.IfBody.Body) {
			return true
		}
		for _, clause := range node.ElseIfClauses {
			if clause != nil && clause.Body != nil && statementsContainReturn(clause.Body.Body) {
				return true
			}
		}
		return node.ElseBody != nil && statementsContainReturn(node.ElseBody.Body)
	case *ast.WhileLoop:
		return node != nil && node.Body != nil && statementsContainReturn(node.Body.Body)
	case *ast.ForLoop:
		return node != nil && node.Body != nil && statementsContainReturn(node.Body.Body)
	case *ast.LoopExpression:
		return node != nil && node.Body != nil && statementsContainReturn(node.Body.Body)
	case *ast.MatchExpression:
		if node == nil {
			return false
		}
		for _, clause := range node.Clauses {
			if clause != nil && expressionContainsReturn(clause.Body) {
				return true
			}
		}
	case *ast.RescueExpression:
		if node == nil || expressionContainsReturn(node.MonitoredExpression) {
			return node != nil
		}
		for _, clause := range node.Clauses {
			if clause != nil && expressionContainsReturn(clause.Body) {
				return true
			}
		}
	case *ast.EnsureExpression:
		return node != nil && (expressionContainsReturn(node.TryExpression) ||
			(node.EnsureBlock != nil && statementsContainReturn(node.EnsureBlock.Body)))
	case *ast.FunctionDefinition:
		return false
	default:
		if expression, ok := statement.(ast.Expression); ok {
			return expressionContainsReturn(expression)
		}
	}
	return false
}

func expressionContainsReturn(expression ast.Expression) bool {
	switch node := expression.(type) {
	case *ast.BlockExpression, *ast.IfExpression, *ast.LoopExpression,
		*ast.MatchExpression, *ast.RescueExpression, *ast.EnsureExpression:
		return statementContainsReturn(node.(ast.Statement))
	default:
		return false
	}
}

func (g *generator) renderCallerOwnedResultVariant(buf *bytes.Buffer, info *functionInfo) {
	resultInfo := g.callerOwnedResultInfo(info)
	if g == nil || buf == nil || info == nil || resultInfo == nil {
		return
	}
	ctx := g.compileBodyContext(info)
	ctx.callerOwnedResultSlot = "__able_result"
	lines, resultExpr, ok := g.compileBody(ctx, info)
	if !ok {
		return
	}
	bodyName := g.compiledBodyName(info)
	entryName := g.compiledEntryName(info)
	if g.executionContextsEnabled() {
		bodyName = g.compiledContextBodyName(info)
		entryName = g.compiledContextEntryName(info)
	}
	bodyName = callerOwnedResultVariantName(bodyName)
	entryName = callerOwnedResultVariantName(entryName)

	g.writeCallerOwnedResultSignature(buf, bodyName, info, resultInfo)
	if g.executionContextsEnabled() {
		fmt.Fprintln(buf, "\t_ = __able_exec_ctx")
	}
	for _, line := range lines {
		fmt.Fprintf(buf, "\t%s\n", line)
	}
	fmt.Fprintf(buf, "\t*__able_result = *%s\n", resultExpr)
	fmt.Fprintln(buf, "\treturn __able_result, nil")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)

	g.writeCallerOwnedResultSignature(buf, entryName, info, resultInfo)
	if g.executionContextsEnabled() {
		fmt.Fprintln(buf, "\t_ = __able_exec_ctx")
		if envVar, ok := g.packageEnvVar(info.Package); ok {
			writeExecutionContextPackageEnv(buf, "\t", "__able_exec_ctx", "__able_runtime", envVar)
		}
	} else if envVar, ok := g.packageEnvVar(info.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	args := make([]string, 0, len(info.Params)+1)
	for _, param := range info.Params {
		args = append(args, param.GoName)
	}
	args = append(args, "__able_result")
	fmt.Fprintf(buf, "\treturn %s(%s)\n", bodyName, g.compiledCallArgs(&compileContext{executionContextExpr: "__able_exec_ctx"}, args))
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

func (g *generator) writeCallerOwnedResultSignature(buf *bytes.Buffer, name string, info *functionInfo, resultInfo *structInfo) {
	fmt.Fprintf(buf, "func %s(", name)
	for index, param := range info.Params {
		if index > 0 {
			fmt.Fprint(buf, ", ")
		}
		fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
	}
	if len(info.Params) > 0 {
		fmt.Fprint(buf, ", ")
	}
	fmt.Fprintf(buf, "__able_result *%s", resultInfo.GoName)
	if g.executionContextsEnabled() {
		fmt.Fprintf(buf, ", __able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (*%s, *__ableControl) {\n", resultInfo.GoName)
}
