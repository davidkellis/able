package typechecker

import "able/interpreter-go/pkg/ast"

// adoptNumericLiteralContext resolves an unsuffixed integer literal against a
// concrete numeric expectation. The literal remains distinguishable from an
// explicitly suffixed source, but its inference fact carries the runtime and
// compiler carrier selected by the valid context.
func (c *Checker) adoptNumericLiteralContext(expr ast.Expression, actual Type, expected Type) (Type, bool) {
	literal, ok := expr.(*ast.IntegerLiteral)
	if !ok || literal == nil || literal.IntegerType != nil {
		return actual, false
	}
	source, ok := actual.(IntegerType)
	if !ok || source.Literal == nil {
		return actual, false
	}
	expected = normalizeSpecialType(expected)
	if !literalAssignableTo(source, expected) {
		return actual, false
	}
	var adopted Type
	switch target := expected.(type) {
	case IntegerType:
		adopted = IntegerType{
			Suffix:  target.Suffix,
			Literal: source.Literal,
		}
	case FloatType:
		adopted = target
	default:
		return actual, false
	}
	c.infer.set(literal, adopted)
	return adopted, true
}
