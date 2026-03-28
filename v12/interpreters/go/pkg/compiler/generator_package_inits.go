package compiler

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func (g *generator) collectPackageInitStatements(program *driver.Program) {
	if g == nil || program == nil {
		return
	}
	seen := make(map[string]struct{})
	modules := make([]*driver.Module, 0, len(program.Modules)+1)
	modules = append(modules, program.Modules...)
	if program.Entry != nil {
		modules = append(modules, program.Entry)
	}
	for _, module := range modules {
		if module == nil || module.AST == nil {
			continue
		}
		stmts := g.moduleInitStatements(module)
		if len(stmts) == 0 {
			continue
		}
		if _, ok := seen[module.Package]; !ok {
			seen[module.Package] = struct{}{}
			g.packageInitOrder = append(g.packageInitOrder, module.Package)
		}
		g.packageInitStatements[module.Package] = stmts
	}
}

func (g *generator) moduleInitStatements(module *driver.Module) []ast.Statement {
	if g == nil || module == nil || module.AST == nil {
		return nil
	}
	var out []ast.Statement
	for _, stmt := range module.AST.Body {
		if stmt == nil || !g.isModuleInitStatement(module.Package, stmt) {
			continue
		}
		out = append(out, g.rewriteModuleInitStatement(module.Package, stmt))
	}
	return out
}

func (g *generator) isModuleInitStatement(pkgName string, stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.FunctionDefinition,
		*ast.StructDefinition,
		*ast.UnionDefinition,
		*ast.InterfaceDefinition,
		*ast.TypeAliasDefinition,
		*ast.MethodsDefinition,
		*ast.ImplementationDefinition,
		*ast.ExternFunctionBody,
		*ast.PreludeStatement,
		*ast.ImportStatement,
		*ast.DynImportStatement,
		*ast.PackageStatement:
		return false
	case *ast.AssignmentExpression:
		if s == nil {
			return false
		}
		if s.Operator == ast.AssignmentDeclare {
			if _, goType := g.literalToRuntimeExpr(s.Right, pkgName); goType != "" {
				return false
			}
		}
		return true
	default:
		return true
	}
}

func (g *generator) rewriteModuleInitStatement(pkgName string, stmt ast.Statement) ast.Statement {
	assign, ok := stmt.(*ast.AssignmentExpression)
	if !ok || assign == nil || assign.Operator != ast.AssignmentDeclare {
		return stmt
	}
	if _, _, ok := g.assignmentTargetName(assign.Left); !ok {
		return stmt
	}
	return ast.NewAssignmentExpression(ast.AssignmentAssign, assign.Left, assign.Right)
}
