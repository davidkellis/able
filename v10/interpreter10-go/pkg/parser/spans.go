package parser

import (
	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func spanFromNode(node *sitter.Node) ast.Span {
	if node == nil {
		return ast.Span{}
	}
	start := node.StartPosition()
	end := node.EndPosition()
	return ast.Span{
		Start: ast.Position{Line: int(start.Row) + 1, Column: int(start.Column) + 1},
		End:   ast.Position{Line: int(end.Row) + 1, Column: int(end.Column) + 1},
	}
}

func annotateSpan(node ast.Node, tsNode *sitter.Node) {
	if node == nil || tsNode == nil {
		return
	}
	ast.SetSpan(node, spanFromNode(tsNode))
}

func annotateStatement(stmt ast.Statement, tsNode *sitter.Node) ast.Statement {
	annotateSpan(stmt, tsNode)
	return stmt
}

func annotateExpression(expr ast.Expression, tsNode *sitter.Node) ast.Expression {
	annotateSpan(expr, tsNode)
	return expr
}

func annotatePattern(pattern ast.Pattern, tsNode *sitter.Node) ast.Pattern {
	annotateSpan(pattern, tsNode)
	return pattern
}

func annotateTypeExpression(typ ast.TypeExpression, tsNode *sitter.Node) ast.TypeExpression {
	annotateSpan(typ, tsNode)
	return typ
}

func isZeroSpan(span ast.Span) bool {
	return span.Start.Line == 0 && span.Start.Column == 0 && span.End.Line == 0 && span.End.Column == 0
}

func comparePosition(a, b ast.Position) int {
	switch {
	case a.Line < b.Line:
		return -1
	case a.Line > b.Line:
		return 1
	case a.Column < b.Column:
		return -1
	case a.Column > b.Column:
		return 1
	default:
		return 0
	}
}

func annotateCompositeExpression(expr ast.Expression, start ast.Node, tsNode *sitter.Node) ast.Expression {
	if expr == nil || tsNode == nil {
		return expr
	}
	span := spanFromNode(tsNode)
	if start != nil {
		startSpan := start.Span()
		if !isZeroSpan(startSpan) {
			span.Start = startSpan.Start
		}
	}
	ast.SetSpan(expr, span)
	return expr
}

func extendExpressionToNode(expr ast.Expression, tsNode *sitter.Node) {
	if expr == nil || tsNode == nil {
		return
	}
	newSpan := spanFromNode(tsNode)
	current := expr.Span()
	if !isZeroSpan(current) {
		newSpan.Start = current.Start
		if comparePosition(current.End, newSpan.End) > 0 {
			newSpan.End = current.End
		}
	}
	ast.SetSpan(expr, newSpan)
}
