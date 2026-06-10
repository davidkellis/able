package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileDynImportStatement(ctx *compileContext, stmt *ast.DynImportStatement) ([]string, bool) {
	if g == nil || ctx == nil || stmt == nil {
		if ctx != nil {
			ctx.setReason("missing dynimport statement")
		}
		return nil, false
	}
	stmtExpr, ok := renderDynImportStatementExpr(stmt)
	if !ok {
		ctx.setReason("invalid dynimport statement")
		return nil, false
	}
	valueTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s, %s := bridge.EvaluateStatement(__able_runtime, %s, __able_dynamic_scope_env)", valueTemp, errTemp, stmtExpr),
		fmt.Sprintf("if %s != nil {", errTemp),
	}
	transferLines, ok := g.lowerControlTransfer(ctx, g.runtimeErrorControlExpr("nil", errTemp))
	if !ok {
		return nil, false
	}
	lines = append(lines, indentLines(transferLines, 1)...)
	lines = append(lines,
		"}",
		fmt.Sprintf("_ = %s", valueTemp),
	)
	return lines, true
}

func renderDynImportStatementExpr(stmt *ast.DynImportStatement) (string, bool) {
	if stmt == nil || len(stmt.PackagePath) == 0 {
		return "", false
	}
	path := make([]string, 0, len(stmt.PackagePath))
	for _, ident := range stmt.PackagePath {
		if ident == nil || strings.TrimSpace(ident.Name) == "" {
			return "", false
		}
		path = append(path, fmt.Sprintf("ast.NewIdentifier(%q)", ident.Name))
	}
	selectors := make([]string, 0, len(stmt.Selectors))
	for _, selector := range stmt.Selectors {
		if selector == nil || selector.Name == nil || strings.TrimSpace(selector.Name.Name) == "" {
			return "", false
		}
		aliasExpr := "nil"
		if selector.Alias != nil && strings.TrimSpace(selector.Alias.Name) != "" {
			aliasExpr = fmt.Sprintf("ast.NewIdentifier(%q)", selector.Alias.Name)
		}
		selectors = append(selectors, fmt.Sprintf("ast.NewImportSelector(ast.NewIdentifier(%q), %s)", selector.Name.Name, aliasExpr))
	}
	selectorExpr := "nil"
	if len(selectors) > 0 {
		selectorExpr = "[]*ast.ImportSelector{" + strings.Join(selectors, ", ") + "}"
	}
	aliasExpr := "nil"
	if stmt.Alias != nil && strings.TrimSpace(stmt.Alias.Name) != "" {
		aliasExpr = fmt.Sprintf("ast.NewIdentifier(%q)", stmt.Alias.Name)
	}
	return fmt.Sprintf(
		"ast.NewDynImportStatement([]*ast.Identifier{%s}, %t, %s, %s)",
		strings.Join(path, ", "),
		stmt.IsWildcard,
		selectorExpr,
		aliasExpr,
	), true
}

func (g *generator) dynamicImportScopePrefix(info *functionInfo) []string {
	if g == nil || info == nil || info.Definition == nil || info.Definition.Body == nil {
		return nil
	}
	hasDynImport := false
	ast.Walk(info.Definition.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.DynImportStatement); ok {
			hasDynImport = true
			return false
		}
		return !hasDynImport
	})
	if !hasDynImport {
		return nil
	}
	parentEnv := "__able_runtime.Env()"
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		parentEnv = envVar
	}
	return []string{
		"var __able_dynamic_scope_env *runtime.Environment",
		"if __able_runtime != nil {",
		fmt.Sprintf("\t__able_dynamic_scope_env = runtime.NewEnvironment(%s)", parentEnv),
		"\tif __able_dynamic_prev_env, __able_dynamic_swapped_env := bridge.SwapEnvIfNeeded(__able_runtime, __able_dynamic_scope_env); __able_dynamic_swapped_env {",
		"\t\tdefer bridge.RestoreEnvIfNeeded(__able_runtime, __able_dynamic_prev_env, __able_dynamic_swapped_env)",
		"\t}",
		"}",
	}
}
