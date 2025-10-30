package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func parseNumberLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	content := sliceContent(node, source)
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

	if strings.ContainsAny(base, ".") || strings.ContainsAny(base, "eE") || floatType != nil {
		value, err := strconv.ParseFloat(sanitized, 64)
		if err != nil {
			return nil, fmt.Errorf("parser: invalid number literal %q", content)
		}
		if floatType != nil {
			return ast.FltTyped(value, floatType), nil
		}
		return ast.Flt(value), nil
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
	return ast.IntBig(value, intType), nil
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

func parseStringLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := sliceContent(node, source)
	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		return nil, fmt.Errorf("parser: invalid string literal %q: %w", raw, err)
	}
	return ast.Str(unquoted), nil
}

func parseCharLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := sliceContent(node, source)
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
	return ast.Chr(unquoted), nil
}
