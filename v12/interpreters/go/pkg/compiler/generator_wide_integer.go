package compiler

import (
	"fmt"
	"math/big"
)

func (g *generator) integerLiteralFitsType(value *big.Int, goType string) bool {
	if value == nil || !g.isIntegerType(goType) {
		return false
	}
	bits := g.intBits(goType)
	if g.isSignedIntegerType(goType) {
		min := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)))
		max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)), big.NewInt(1))
		return value.Cmp(min) >= 0 && value.Cmp(max) <= 0
	}
	if value.Sign() < 0 {
		return false
	}
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	return value.Cmp(max) <= 0
}

func (g *generator) integerLiteralExprForType(value *big.Int, goType string) (string, bool) {
	if value == nil || !g.isWideIntegerType(goType) {
		return "", false
	}
	pattern := new(big.Int).Set(value)
	if goType == "runtime.Int128" {
		min := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
		max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
		if pattern.Cmp(min) < 0 || pattern.Cmp(max) > 0 {
			return "", false
		}
		if pattern.Sign() < 0 {
			pattern.Add(pattern, new(big.Int).Lsh(big.NewInt(1), 128))
		}
	} else if pattern.Sign() < 0 || pattern.BitLen() > 128 {
		return "", false
	}
	low := pattern.Uint64()
	high := new(big.Int).Rsh(new(big.Int).Set(pattern), 64).Uint64()
	return fmt.Sprintf("%s{High: uint64(%d), Low: uint64(%d)}", goType, high, low), true
}

func (g *generator) compileCheckedWideIntegerBinaryExpression(ctx *compileContext, left string, right string, op string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	method := map[string]string{"+": "AddChecked", "-": "SubChecked", "*": "MulChecked"}[op]
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
		fmt.Sprintf("%s, %s := (%s).%s(%s)", resultTemp, okTemp, leftTemp, method, rightTemp),
		fmt.Sprintf("var %s *__ableControl", controlTemp),
		fmt.Sprintf("if !%s { %s = __able_raise_overflow(%s) }", okTemp, controlTemp, nodeName),
	}
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp
}

func (g *generator) compileWideDivModExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	quotientTemp := ctx.newTemp()
	remainderTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
		fmt.Sprintf("var %s *__ableControl", controlTemp),
	}
	if operandType == "runtime.Uint128" {
		okTemp := ctx.newTemp()
		assignment := fmt.Sprintf("%s, %s, %s := (%s).DivMod(%s)", quotientTemp, remainderTemp, okTemp, leftTemp, rightTemp)
		if op == "//" {
			assignment = fmt.Sprintf("%s, _, %s := (%s).DivMod(%s)", quotientTemp, okTemp, leftTemp, rightTemp)
		} else {
			assignment = fmt.Sprintf("_, %s, %s := (%s).DivMod(%s)", remainderTemp, okTemp, leftTemp, rightTemp)
		}
		lines = append(lines,
			assignment,
			fmt.Sprintf("if !%s { %s = __able_raise_division_by_zero(%s) }", okTemp, controlTemp, nodeName),
		)
	} else {
		nonzeroTemp := ctx.newTemp()
		inRangeTemp := ctx.newTemp()
		assignment := fmt.Sprintf("%s, %s, %s, %s := (%s).DivMod(%s)", quotientTemp, remainderTemp, nonzeroTemp, inRangeTemp, leftTemp, rightTemp)
		if op == "//" {
			assignment = fmt.Sprintf("%s, _, %s, %s := (%s).DivMod(%s)", quotientTemp, nonzeroTemp, inRangeTemp, leftTemp, rightTemp)
		} else {
			assignment = fmt.Sprintf("_, %s, %s, %s := (%s).DivMod(%s)", remainderTemp, nonzeroTemp, inRangeTemp, leftTemp, rightTemp)
		}
		lines = append(lines,
			assignment,
			fmt.Sprintf("if !%s { %s = __able_raise_division_by_zero(%s) } else if !%s { %s = __able_raise_overflow(%s) }", nonzeroTemp, controlTemp, nodeName, inRangeTemp, controlTemp, nodeName),
		)
	}
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	if op == "//" {
		return lines, quotientTemp
	}
	return lines, remainderTemp
}

func (g *generator) compileWidePowExpression(ctx *compileContext, left string, right string, operandType string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
		fmt.Sprintf("var %s *__ableControl", controlTemp),
	}
	if operandType == "runtime.Uint128" {
		okTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := (%s).PowChecked(%s)", resultTemp, okTemp, leftTemp, rightTemp),
			fmt.Sprintf("if !%s { %s = __able_raise_overflow(%s) }", okTemp, controlTemp, nodeName),
		)
	} else {
		statusTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := (%s).PowChecked(%s)", resultTemp, statusTemp, leftTemp, rightTemp),
			fmt.Sprintf("if %s == 1 { %s = __able_raise_runtime_error(%s, \"Negative integer exponent is not supported\") } else if %s == 2 { %s = __able_raise_overflow(%s) }", statusTemp, controlTemp, nodeName, statusTemp, controlTemp, nodeName),
		)
	}
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp
}
