package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
)

func (ctx *parseContext) parseNumberLiteral(node *sitter.Node) (ast.Expression, error) {
	content := sliceContent(node, ctx.source)
	if content == "" {
		return nil, fmt.Errorf("parser: empty number literal")
	}

	base := content
	var intType *ast.IntegerType
	var floatType *ast.FloatType

	if idx := strings.LastIndex(content, "_"); idx > 0 {
		suffix := content[idx+1:]
		if isNumericSuffix(suffix) {
			base = content[:idx]
			switch suffix {
			case "i8":
				t := ast.IntegerTypeI8
				intType = &t
			case "i16":
				t := ast.IntegerTypeI16
				intType = &t
			case "i32":
				t := ast.IntegerTypeI32
				intType = &t
			case "i64":
				t := ast.IntegerTypeI64
				intType = &t
			case "i128":
				t := ast.IntegerTypeI128
				intType = &t
			case "u8":
				t := ast.IntegerTypeU8
				intType = &t
			case "u16":
				t := ast.IntegerTypeU16
				intType = &t
			case "u32":
				t := ast.IntegerTypeU32
				intType = &t
			case "u64":
				t := ast.IntegerTypeU64
				intType = &t
			case "u128":
				t := ast.IntegerTypeU128
				intType = &t
			case "f32":
				t := ast.FloatTypeF32
				floatType = &t
			case "f64":
				t := ast.FloatTypeF64
				floatType = &t
			}
		}
	}

	sanitized := strings.ReplaceAll(base, "_", "")
	lower := strings.ToLower(base)

	hasBasePrefix := strings.HasPrefix(lower, "0b") || strings.HasPrefix(lower, "0o") || strings.HasPrefix(lower, "0x")
	isHexLiteral := strings.HasPrefix(lower, "0x")
	if hasBasePrefix && !isHexLiteral {
		end := int(node.EndByte())
		if end < len(ctx.source) {
			switch ctx.source[end] {
			case 'e', 'E':
				return nil, fmt.Errorf("parser: invalid number literal %q", content+string(ctx.source[end]))
			}
		}
	}
	if hasBasePrefix && !isHexLiteral && strings.ContainsAny(base, "eE") {
		return nil, fmt.Errorf("parser: invalid number literal %q", content)
	}

	hasExponent := !hasBasePrefix && strings.ContainsAny(base, "eE")
	hasDecimal := strings.Contains(base, ".")
	if hasDecimal || hasExponent || floatType != nil {
		value, err := strconv.ParseFloat(sanitized, 64)
		if err != nil {
			return nil, fmt.Errorf("parser: invalid number literal %q", content)
		}
		if floatType != nil {
			return annotateExpression(ast.FltTyped(value, floatType), node), nil
		}
		return annotateExpression(ast.Flt(value), node), nil
	}

	var (
		intBase = 10
		digits  = sanitized
	)

	switch {
	case strings.HasPrefix(lower, "0b"):
		intBase = 2
		digits = strings.TrimPrefix(strings.ReplaceAll(lower, "_", ""), "0b")
	case strings.HasPrefix(lower, "0o"):
		intBase = 8
		digits = strings.TrimPrefix(strings.ReplaceAll(lower, "_", ""), "0o")
	case strings.HasPrefix(lower, "0x"):
		intBase = 16
		digits = strings.TrimPrefix(strings.ReplaceAll(lower, "_", ""), "0x")
	default:
		digits = sanitized
	}

	if digits == "" {
		return nil, fmt.Errorf("parser: invalid number literal %q", content)
	}

	value := new(big.Int)
	if _, ok := value.SetString(digits, intBase); !ok {
		return nil, fmt.Errorf("parser: invalid number literal %q", content)
	}
	return annotateExpression(ast.IntBig(value, intType), node), nil
}

func isNumericSuffix(s string) bool {
	switch s {
	case "i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128",
		"f32", "f64":
		return true
	default:
		return false
	}
}

func (ctx *parseContext) parseStringLiteral(node *sitter.Node) (ast.Expression, error) {
	raw := sliceContent(node, ctx.source)
	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		return nil, fmt.Errorf("parser: invalid string literal %q: %w", raw, err)
	}
	return annotateExpression(ast.Str(unquoted), node), nil
}

func (ctx *parseContext) parseCharLiteral(node *sitter.Node) (ast.Expression, error) {
	raw := sliceContent(node, ctx.source)
	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		return nil, fmt.Errorf("parser: invalid character literal %q: %w", raw, err)
	}
	if len(unquoted) == 0 {
		return nil, fmt.Errorf("parser: empty character literal")
	}
	if len([]rune(unquoted)) != 1 {
		return nil, fmt.Errorf("parser: character literal %q must resolve to a single rune", raw)
	}
	return annotateExpression(ast.Chr(unquoted), node), nil
}
