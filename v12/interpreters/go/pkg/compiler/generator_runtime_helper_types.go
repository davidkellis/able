package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) runtimeHelperResultTypeExpr(name string) ast.TypeExpression {
	switch name {
	case "__able_String_from_builtin":
		return ast.Gen(ast.Ty("Array"), ast.Ty("u8"))
	case "__able_String_to_builtin":
		return ast.Ty("String")
	case "__able_char_from_codepoint":
		return ast.Ty("char")
	case "__able_char_to_codepoint":
		return ast.Ty("i32")
	case "__able_f32_bits":
		return ast.Ty("u32")
	case "__able_f64_bits":
		return ast.Ty("u64")
	case "__able_u64_mul":
		return ast.Ty("u64")
	case "__able_ratio_from_float":
		return ast.Ty("Ratio")
	default:
		return nil
	}
}
