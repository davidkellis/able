package main

import (
	"fmt"
	"go/ast"
	"math"
	"strconv"
	"strings"
)

type intRange struct {
	Known bool
	Min   int64
	Max   int64
}

type parameterRange struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Min  int64  `json:"min"`
	Max  int64  `json:"max"`
}

type aggregateRange struct {
	Path string `json:"path"`
	Min  int64  `json:"min"`
	Max  int64  `json:"max"`
}

type primitiveBlocker struct {
	Kind             string `json:"kind"`
	Helper           string `json:"helper"`
	File             string `json:"file"`
	Line             int    `json:"line"`
	UniversalSafe    bool   `json:"universal_safe"`
	ClosedDirectSafe bool   `json:"closed_direct_safe"`
	UniversalReason  string `json:"universal_reason"`
	ClosedReason     string `json:"closed_direct_reason"`
}

type relationalBlocker struct {
	Kind             string `json:"kind"`
	File             string `json:"file"`
	Line             int    `json:"line"`
	IndexExpression  string `json:"index_expression"`
	LengthExpression string `json:"length_expression"`
	UniversalSafe    bool   `json:"universal_safe"`
	ClosedDirectSafe bool   `json:"closed_direct_safe"`
	UniversalReason  string `json:"universal_reason"`
	ClosedReason     string `json:"closed_direct_reason"`
}

type functionParam struct {
	Name    string
	Type    string
	Integer bool
	Full    intRange
}

type rangeCall struct {
	Callee         string
	Args           []intRange
	AggregateFacts []map[string]intRange
}

type blockerObservation struct {
	Kind   string
	Helper string
	Pos    ast.Node
	Safe   bool
	Reason string
}

type relationalObservation struct {
	Kind   string
	Pos    ast.Node
	Index  string
	Length string
	Safe   bool
	Reason string
}

func knownRange(minValue, maxValue int64) intRange {
	if minValue > maxValue {
		return intRange{}
	}
	return intRange{Known: true, Min: minValue, Max: maxValue}
}

func unionRange(left, right intRange) intRange {
	if !left.Known || !right.Known {
		return intRange{}
	}
	if right.Min < left.Min {
		left.Min = right.Min
	}
	if right.Max > left.Max {
		left.Max = right.Max
	}
	return left
}

func intersectRange(left, right intRange) intRange {
	if !left.Known {
		return right
	}
	if !right.Known {
		return left
	}
	if right.Min > left.Min {
		left.Min = right.Min
	}
	if right.Max < left.Max {
		left.Max = right.Max
	}
	if left.Min > left.Max {
		return intRange{}
	}
	return left
}

func rangeEqual(left, right intRange) bool {
	return left == right
}

func fullIntegerRange(typeName string) (intRange, bool) {
	switch typeName {
	case "int8":
		return knownRange(math.MinInt8, math.MaxInt8), true
	case "int16":
		return knownRange(math.MinInt16, math.MaxInt16), true
	case "int32":
		return knownRange(math.MinInt32, math.MaxInt32), true
	case "int64", "int":
		return knownRange(math.MinInt64, math.MaxInt64), true
	case "uint8":
		return knownRange(0, math.MaxUint8), true
	case "uint16":
		return knownRange(0, math.MaxUint16), true
	case "uint32":
		return knownRange(0, math.MaxUint32), true
	default:
		return intRange{}, false
	}
}

func integerTypeName(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident == nil {
		return ""
	}
	return ident.Name
}

func integerLiteral(expr ast.Expr) (int64, bool) {
	if paren, ok := expr.(*ast.ParenExpr); ok {
		return integerLiteral(paren.X)
	}
	if call, ok := expr.(*ast.CallExpr); ok && len(call.Args) == 1 {
		if _, integer := fullIntegerRange(calledName(call.Fun)); integer {
			return integerLiteral(call.Args[0])
		}
	}
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op.String() == "-" {
		value, ok := integerLiteral(unary.X)
		if !ok || value == math.MinInt64 {
			return 0, false
		}
		return -value, true
	}
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit == nil {
		return 0, false
	}
	text := strings.ReplaceAll(lit.Value, "_", "")
	value, err := strconv.ParseInt(text, 0, 64)
	return value, err == nil
}

func addRange(left, right intRange) intRange {
	if !left.Known || !right.Known {
		return intRange{}
	}
	minValue, ok := add64(left.Min, right.Min)
	if !ok {
		return intRange{}
	}
	maxValue, ok := add64(left.Max, right.Max)
	if !ok {
		return intRange{}
	}
	return knownRange(minValue, maxValue)
}

func subtractRange(left, right intRange) intRange {
	if !left.Known || !right.Known {
		return intRange{}
	}
	minValue, ok := sub64(left.Min, right.Max)
	if !ok {
		return intRange{}
	}
	maxValue, ok := sub64(left.Max, right.Min)
	if !ok {
		return intRange{}
	}
	return knownRange(minValue, maxValue)
}

func multiplyRange(left, right intRange) intRange {
	if !left.Known || !right.Known {
		return intRange{}
	}
	values := [4]int64{}
	pairs := [][2]int64{{left.Min, right.Min}, {left.Min, right.Max}, {left.Max, right.Min}, {left.Max, right.Max}}
	for idx, pair := range pairs {
		value, ok := mulSigned64(pair[0], pair[1])
		if !ok {
			return intRange{}
		}
		values[idx] = value
	}
	minValue, maxValue := values[0], values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	return knownRange(minValue, maxValue)
}

func add64(left, right int64) (int64, bool) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, false
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, false
	}
	return left + right, true
}

func sub64(left, right int64) (int64, bool) {
	if right == math.MinInt64 {
		if left >= 0 {
			return 0, false
		}
		return left - right, true
	}
	return add64(left, -right)
}

func mul64(left, right int64) (int64, bool) {
	if left < 0 || right < 0 {
		return 0, false
	}
	if left != 0 && right > math.MaxInt64/left {
		return 0, false
	}
	return left * right, true
}

func mulSigned64(left, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}
	if left == -1 && right == math.MinInt64 || right == -1 && left == math.MinInt64 {
		return 0, false
	}
	result := left * right
	return result, result/right == left
}

func describeRange(value intRange) string {
	if !value.Known {
		return "range is unknown"
	}
	return fmt.Sprintf("range [%d,%d]", value.Min, value.Max)
}
