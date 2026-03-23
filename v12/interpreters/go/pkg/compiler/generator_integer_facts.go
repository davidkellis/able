package compiler

import "able/interpreter-go/pkg/ast"

type integerFact struct {
	NonNegative  bool
	HasMax       bool
	MaxInclusive int64
}

func cloneIntegerFacts(src map[string]integerFact) map[string]integerFact {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]integerFact, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func (f integerFact) hasUsefulFact() bool {
	return f.NonNegative || f.HasMax
}

func (c *compileContext) setLocalBinding(name string, info paramInfo) {
	if c == nil || name == "" {
		return
	}
	if c.locals == nil {
		c.locals = make(map[string]paramInfo)
	}
	c.locals[name] = info
	if info.GoName != "" {
		c.clearIntegerFact(info.GoName)
	}
}

func (c *compileContext) integerFactForGoName(goName string) (integerFact, bool) {
	if c == nil || goName == "" || c.integerFacts == nil {
		return integerFact{}, false
	}
	fact, ok := c.integerFacts[goName]
	return fact, ok
}

func (c *compileContext) setIntegerFact(goName string, fact integerFact) {
	if c == nil || goName == "" {
		return
	}
	if c.integerFacts == nil {
		c.integerFacts = make(map[string]integerFact)
	}
	c.integerFacts[goName] = fact
}

func (c *compileContext) clearIntegerFact(goName string) {
	if c == nil || goName == "" || c.integerFacts == nil {
		return
	}
	delete(c.integerFacts, goName)
}

func (g *generator) refreshIntegerFactForBinding(ctx *compileContext, binding paramInfo, expr ast.Expression) {
	if g == nil || ctx == nil || binding.GoName == "" {
		return
	}
	if !g.isIntegerType(binding.GoType) {
		ctx.clearIntegerFact(binding.GoName)
		return
	}
	fact, ok := g.exprIntegerFact(ctx, expr)
	if g.isUnsignedIntegerType(binding.GoType) {
		if !ok {
			fact = integerFact{}
		}
		fact.NonNegative = true
		ok = fact.hasUsefulFact()
	}
	if ok && fact.hasUsefulFact() {
		ctx.setIntegerFact(binding.GoName, fact)
		return
	}
	ctx.clearIntegerFact(binding.GoName)
}

func (g *generator) exprProvenNonNegative(ctx *compileContext, expr ast.Expression) bool {
	fact, ok := g.exprIntegerFact(ctx, expr)
	return ok && fact.NonNegative
}

func (g *generator) exprIntegerFact(ctx *compileContext, expr ast.Expression) (integerFact, bool) {
	if g == nil || expr == nil {
		return integerFact{}, false
	}
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		if e == nil || e.Value == nil || !e.Value.IsInt64() {
			return integerFact{}, false
		}
		value := e.Value.Int64()
		return integerFact{
			NonNegative:  value >= 0,
			HasMax:       true,
			MaxInclusive: value,
		}, true
	case *ast.Identifier:
		if ctx == nil || e == nil || e.Name == "" {
			return integerFact{}, false
		}
		binding, ok := ctx.lookup(e.Name)
		if !ok {
			return integerFact{}, false
		}
		if g.isUnsignedIntegerType(binding.GoType) {
			if fact, ok := ctx.integerFactForGoName(binding.GoName); ok {
				fact.NonNegative = true
				return fact, true
			}
			return integerFact{NonNegative: true}, true
		}
		fact, ok := ctx.integerFactForGoName(binding.GoName)
		return fact, ok && fact.hasUsefulFact()
	case *ast.TypeCastExpression:
		return g.exprIntegerFactForCast(ctx, e)
	case *ast.UnaryExpression:
		if e == nil {
			return integerFact{}, false
		}
		switch e.Operator {
		case "+":
			return g.exprIntegerFact(ctx, e.Operand)
		case "-":
			if lit, ok := e.Operand.(*ast.IntegerLiteral); ok && lit != nil && lit.Value != nil {
				if lit.Value.Sign() == 0 {
					return integerFact{NonNegative: true, HasMax: true, MaxInclusive: 0}, true
				}
			}
		}
		return integerFact{}, false
	case *ast.BinaryExpression:
		if e == nil {
			return integerFact{}, false
		}
		return g.exprIntegerFactForBinary(ctx, e)
	default:
		return integerFact{}, false
	}
}

func (g *generator) exprIntegerFactForCast(ctx *compileContext, expr *ast.TypeCastExpression) (integerFact, bool) {
	if g == nil || expr == nil || expr.TargetType == nil {
		return integerFact{}, false
	}
	targetType, ok := g.lowerCarrierType(ctx, expr.TargetType)
	if !ok || !g.isIntegerType(targetType) {
		return integerFact{}, false
	}
	if fact, ok := g.exprIntegerFact(ctx, expr.Expression); ok && fact.hasUsefulFact() {
		if g.isUnsignedIntegerType(targetType) {
			fact.NonNegative = true
		}
		if fact.HasMax {
			if maxBound, ok := g.signedIntegerUpperBound(targetType); ok {
				if fact.MaxInclusive > maxBound {
					return integerFact{}, false
				}
			}
			if maxBound, ok := g.unsignedIntegerUpperBound(targetType); ok {
				if fact.MaxInclusive < 0 || fact.MaxInclusive > maxBound {
					return integerFact{}, false
				}
			}
		}
		return fact, true
	}
	binary, ok := expr.Expression.(*ast.BinaryExpression)
	if !ok || binary == nil {
		return integerFact{}, false
	}
	switch binary.Operator {
	case "/", "//":
		leftFact, ok := g.exprIntegerFact(ctx, binary.Left)
		if !ok || !leftFact.NonNegative || !leftFact.HasMax {
			return integerFact{}, false
		}
		divisor, ok := positiveIntegerLiteralValue(binary.Right)
		if !ok || divisor == 0 {
			return integerFact{}, false
		}
		return integerFact{
			NonNegative:  true,
			HasMax:       true,
			MaxInclusive: leftFact.MaxInclusive / divisor,
		}, true
	default:
		return integerFact{}, false
	}
}

func (g *generator) exprIntegerFactForBinary(ctx *compileContext, expr *ast.BinaryExpression) (integerFact, bool) {
	if g == nil || expr == nil {
		return integerFact{}, false
	}
	leftFact, leftOK := g.exprIntegerFact(ctx, expr.Left)
	rightFact, rightOK := g.exprIntegerFact(ctx, expr.Right)
	switch expr.Operator {
	case "+":
		if !leftOK || !rightOK || !leftFact.NonNegative || !rightFact.NonNegative || !leftFact.HasMax || !rightFact.HasMax {
			return integerFact{}, false
		}
		sum, ok := addInt64NoOverflow(leftFact.MaxInclusive, rightFact.MaxInclusive)
		if !ok {
			return integerFact{}, false
		}
		return integerFact{NonNegative: true, HasMax: true, MaxInclusive: sum}, true
	case "-":
		if !leftOK {
			return integerFact{}, false
		}
		fact := integerFact{HasMax: leftFact.HasMax, MaxInclusive: leftFact.MaxInclusive}
		return fact, fact.hasUsefulFact()
	case "*":
		if !leftOK || !rightOK || !leftFact.NonNegative || !rightFact.NonNegative || !leftFact.HasMax || !rightFact.HasMax {
			return integerFact{}, false
		}
		product, ok := mulInt64NoOverflow(leftFact.MaxInclusive, rightFact.MaxInclusive)
		if !ok {
			return integerFact{}, false
		}
		return integerFact{NonNegative: true, HasMax: true, MaxInclusive: product}, true
	default:
		return integerFact{}, false
	}
}

func positiveIntegerLiteralValue(expr ast.Expression) (int64, bool) {
	lit, ok := expr.(*ast.IntegerLiteral)
	if !ok || lit == nil || lit.Value == nil || !lit.Value.IsInt64() {
		return 0, false
	}
	value := lit.Value.Int64()
	if value <= 0 {
		return 0, false
	}
	return value, true
}

func addInt64NoOverflow(a int64, b int64) (int64, bool) {
	if b > 0 && a > (1<<63-1)-b {
		return 0, false
	}
	if b < 0 && a < (-1<<63)-b {
		return 0, false
	}
	return a + b, true
}

func mulInt64NoOverflow(a int64, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	if a < 0 || b < 0 {
		return 0, false
	}
	if a > (1<<63-1)/b {
		return 0, false
	}
	return a * b, true
}

func (g *generator) signedIntegerUpperBound(goType string) (int64, bool) {
	switch goType {
	case "int8":
		return 127, true
	case "int16":
		return 32767, true
	case "int32":
		return 2147483647, true
	case "int64":
		return 1<<63 - 1, true
	default:
		return 0, false
	}
}

func (g *generator) unsignedIntegerUpperBound(goType string) (int64, bool) {
	switch goType {
	case "uint8":
		return 255, true
	case "uint16":
		return 65535, true
	case "uint32":
		return 4294967295, true
	default:
		return 0, false
	}
}
