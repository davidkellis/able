package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"math"
)

func generatedBoundsObservation(statement *ast.IfStmt, env rangeEnv, evaluate func(ast.Expr, rangeEnv) intRange) (relationalObservation, bool) {
	if statement == nil {
		return relationalObservation{}, false
	}
	index, length, kind, ok := generatedBoundsCondition(statement.Cond)
	if !ok {
		return relationalObservation{}, false
	}
	indexRange := evaluate(index, env)
	lengthRange := evaluate(length, env)
	safe := indexRange.Known && lengthRange.Known &&
		indexRange.Min >= 0 && lengthRange.Min >= 0 &&
		indexRange.Max < lengthRange.Min
	return relationalObservation{
		Kind:   kind,
		Pos:    statement,
		Index:  formatGoExpr(index),
		Length: formatGoExpr(length),
		Safe:   safe,
		Reason: boundsReason(indexRange, lengthRange),
	}, true
}

func generatedBoundsCondition(expr ast.Expr) (ast.Expr, ast.Expr, string, bool) {
	binary, ok := unparen(expr).(*ast.BinaryExpr)
	if !ok {
		return nil, nil, "", false
	}
	switch binary.Op {
	case token.LOR:
		index, zeroOK := comparisonWithZero(binary.X, token.LSS)
		rightIndex, length, lengthOK := comparisonPair(binary.Y, token.GEQ)
		if zeroOK && lengthOK && sameGoExpr(index, rightIndex) {
			return index, length, "strict-array-bounds", true
		}
	case token.LAND:
		index, zeroOK := comparisonWithZero(binary.X, token.GEQ)
		rightIndex, length, lengthOK := comparisonPair(binary.Y, token.LSS)
		if zeroOK && lengthOK && sameGoExpr(index, rightIndex) {
			return index, length, "safe-array-bounds", true
		}
	}
	return nil, nil, "", false
}

func comparisonWithZero(expr ast.Expr, operator token.Token) (ast.Expr, bool) {
	binary, ok := unparen(expr).(*ast.BinaryExpr)
	if !ok || binary.Op != operator {
		return nil, false
	}
	value, ok := integerLiteral(binary.Y)
	return binary.X, ok && value == 0
}

func comparisonPair(expr ast.Expr, operator token.Token) (ast.Expr, ast.Expr, bool) {
	binary, ok := unparen(expr).(*ast.BinaryExpr)
	if !ok || binary.Op != operator {
		return nil, nil, false
	}
	return binary.X, binary.Y, true
}

func unparen(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}

func formatGoExpr(expr ast.Expr) string {
	var out bytes.Buffer
	if expr != nil && printer.Fprint(&out, token.NewFileSet(), expr) == nil {
		return out.String()
	}
	return ""
}

func sameGoExpr(left, right ast.Expr) bool {
	return formatGoExpr(left) == formatGoExpr(right)
}

func boundsReason(index, length intRange) string {
	return describeRange(index) + "; length " + describeRange(length)
}

func nonnegativeBitwiseUpper(left, right int64) (int64, bool) {
	if left < 0 || right < 0 {
		return 0, false
	}
	maxValue := left
	if right > maxValue {
		maxValue = right
	}
	if maxValue == 0 {
		return 0, true
	}
	if maxValue == math.MaxInt64 {
		return math.MaxInt64, true
	}
	bits := uint(0)
	for value := maxValue; value > 0; value >>= 1 {
		bits++
	}
	if bits >= 63 {
		return math.MaxInt64, true
	}
	return int64(1)<<bits - 1, true
}
