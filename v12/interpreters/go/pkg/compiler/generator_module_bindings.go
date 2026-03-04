package compiler

import (
	"fmt"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type moduleBinding struct {
	Name    string
	GoValue string // Go literal expression (e.g., "uint64(14695981039346656037)")
	GoType  string // Go type (e.g., "uint64")
}

type evaluatedConst struct {
	val    *big.Int
	suffix string
}

func (g *generator) collectModuleBinding(assign *ast.AssignmentExpression, pkgName string) {
	if g == nil || assign == nil || assign.Right == nil {
		return
	}
	name, _, ok := g.assignmentTargetName(assign.Left)
	if !ok || name == "" {
		return
	}
	// Try to evaluate as a constant integer first, to track for later resolution.
	if val, suffix, ok := g.evalConstInt(assign.Right, pkgName); ok {
		if g.evaluatedConstants == nil {
			g.evaluatedConstants = make(map[string]*evaluatedConst)
		}
		g.evaluatedConstants[pkgName+"::"+name] = &evaluatedConst{val: val, suffix: suffix}
	}
	goValue, goType := g.literalToRuntimeExpr(assign.Right, pkgName)
	if goValue == "" {
		return
	}
	if g.moduleBindings == nil {
		g.moduleBindings = make(map[string][]moduleBinding)
	}
	g.moduleBindings[pkgName] = append(g.moduleBindings[pkgName], moduleBinding{
		Name:    name,
		GoValue: goValue,
		GoType:  goType,
	})
}

func (g *generator) collectModuleBindingName(assign *ast.AssignmentExpression, pkgName string) {
	if g == nil || assign == nil {
		return
	}
	name, _, ok := g.assignmentTargetName(assign.Left)
	if !ok || strings.TrimSpace(name) == "" {
		return
	}
	perPkg := g.moduleBindingNames[pkgName]
	if perPkg == nil {
		perPkg = make(map[string]struct{})
		g.moduleBindingNames[pkgName] = perPkg
	}
	perPkg[name] = struct{}{}
}

func (g *generator) hasModuleBindingName(pkgName string, name string) bool {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return false
	}
	perPkg := g.moduleBindingNames[pkgName]
	if perPkg == nil {
		return false
	}
	_, ok := perPkg[name]
	return ok
}

func (g *generator) literalToRuntimeExpr(expr ast.Expression, pkgName string) (string, string) {
	switch lit := expr.(type) {
	case *ast.IntegerLiteral:
		if lit == nil || lit.Value == nil {
			return "", ""
		}
		valStr := lit.Value.String()
		bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", valStr)
		suffix := integerTypeSuffix(lit.IntegerType)
		return fmt.Sprintf("runtime.NewBigIntValue(%s, %s)", bigExpr, suffix), "runtime.Value"
	case *ast.FloatLiteral:
		if lit == nil {
			return "", ""
		}
		suffix := "runtime.FloatF64"
		if lit.FloatType != nil && *lit.FloatType == ast.FloatTypeF32 {
			suffix = "runtime.FloatF32"
		}
		return fmt.Sprintf("runtime.FloatValue{Val: %v, TypeSuffix: %s}", lit.Value, suffix), "runtime.Value"
	case *ast.StringLiteral:
		if lit == nil {
			return "", ""
		}
		return fmt.Sprintf("runtime.StringValue{Val: %q}", lit.Value), "runtime.Value"
	case *ast.BooleanLiteral:
		if lit == nil {
			return "", ""
		}
		return fmt.Sprintf("runtime.BoolValue{Val: %t}", lit.Value), "runtime.Value"
	case *ast.UnaryExpression:
		if lit == nil || lit.Operator != "-" {
			return "", ""
		}
		val, suffix, ok := g.evalConstInt(lit, pkgName)
		if ok {
			bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", val.String())
			return fmt.Sprintf("runtime.NewBigIntValue(%s, %s)", bigExpr, suffix), "runtime.Value"
		}
		return "", ""
	case *ast.BinaryExpression:
		if lit == nil {
			return "", ""
		}
		val, suffix, ok := g.evalConstInt(lit, pkgName)
		if !ok {
			return "", ""
		}
		bigExpr := fmt.Sprintf("func() *big.Int { v, _ := new(big.Int).SetString(%q, 10); return v }()", val.String())
		return fmt.Sprintf("runtime.NewBigIntValue(%s, %s)", bigExpr, suffix), "runtime.Value"
	}
	return "", ""
}

// evalConstInt evaluates a constant integer expression at compile time.
// Handles integer literals, negated integer literals, binary arithmetic/bitwise,
// and identifier references to previously evaluated constants in the same package.
func (g *generator) evalConstInt(expr ast.Expression, pkgName string) (*big.Int, string, bool) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		if e == nil || e.Value == nil {
			return nil, "", false
		}
		return new(big.Int).Set(e.Value), integerTypeSuffix(e.IntegerType), true
	case *ast.Identifier:
		if e == nil || g == nil || g.evaluatedConstants == nil {
			return nil, "", false
		}
		key := pkgName + "::" + e.Name
		if c, ok := g.evaluatedConstants[key]; ok {
			return new(big.Int).Set(c.val), c.suffix, true
		}
		return nil, "", false
	case *ast.UnaryExpression:
		if e == nil || e.Operator != "-" {
			return nil, "", false
		}
		val, suffix, ok := g.evalConstInt(e.Operand, pkgName)
		if !ok {
			return nil, "", false
		}
		return val.Neg(val), suffix, true
	case *ast.BinaryExpression:
		if e == nil {
			return nil, "", false
		}
		left, lSuffix, lok := g.evalConstInt(e.Left, pkgName)
		right, _, rok := g.evalConstInt(e.Right, pkgName)
		if !lok || !rok {
			return nil, "", false
		}
		switch e.Operator {
		case "+":
			return left.Add(left, right), lSuffix, true
		case "-":
			return left.Sub(left, right), lSuffix, true
		case "*":
			return left.Mul(left, right), lSuffix, true
		case "/":
			if right.Sign() == 0 {
				return nil, "", false
			}
			return left.Div(left, right), lSuffix, true
		case ".<<", "<<":
			shift := right.Int64()
			if shift < 0 || shift > 64 {
				return nil, "", false
			}
			return left.Lsh(left, uint(shift)), lSuffix, true
		case ".>>", ">>":
			shift := right.Int64()
			if shift < 0 || shift > 64 {
				return nil, "", false
			}
			return left.Rsh(left, uint(shift)), lSuffix, true
		case ".&", "&":
			return left.And(left, right), lSuffix, true
		case ".|", "|":
			return left.Or(left, right), lSuffix, true
		case ".^", "^":
			return left.Xor(left, right), lSuffix, true
		}
		return nil, "", false
	}
	return nil, "", false
}

func integerTypeSuffix(intType *ast.IntegerType) string {
	if intType != nil {
		switch *intType {
		case ast.IntegerTypeU64:
			return "runtime.IntegerU64"
		case ast.IntegerTypeI64:
			return "runtime.IntegerI64"
		case ast.IntegerTypeU32:
			return "runtime.IntegerU32"
		case ast.IntegerTypeI32:
			return "runtime.IntegerI32"
		case ast.IntegerTypeU16:
			return "runtime.IntegerU16"
		case ast.IntegerTypeI16:
			return "runtime.IntegerI16"
		case ast.IntegerTypeU8:
			return "runtime.IntegerU8"
		case ast.IntegerTypeI8:
			return "runtime.IntegerI8"
		case ast.IntegerTypeI128:
			return "runtime.IntegerI128"
		case ast.IntegerTypeU128:
			return "runtime.IntegerU128"
		}
	}
	return "runtime.IntegerI32"
}
