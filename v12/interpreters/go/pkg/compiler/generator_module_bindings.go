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

func (g *generator) markMutableModuleBindingName(pkgName string, name string) {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return
	}
	perPkg := g.moduleMutableBindingNames[pkgName]
	if perPkg == nil {
		perPkg = make(map[string]struct{})
		g.moduleMutableBindingNames[pkgName] = perPkg
	}
	perPkg[name] = struct{}{}
}

func (g *generator) isMutableModuleBindingName(pkgName string, name string) bool {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return false
	}
	perPkg := g.moduleMutableBindingNames[pkgName]
	if perPkg == nil {
		return false
	}
	_, ok := perPkg[name]
	return ok
}

func (g *generator) collectMutableModuleBindings(stmts []ast.Statement, pkgName string) {
	if g == nil || len(stmts) == 0 || strings.TrimSpace(pkgName) == "" {
		return
	}
	for _, stmt := range stmts {
		fn, ok := stmt.(*ast.FunctionDefinition)
		if !ok || fn == nil || fn.Body == nil {
			continue
		}
		g.collectMutableModuleBindingsFromBlock(fn.Body, pkgName)
	}
}

func (g *generator) collectMutableModuleBindingsFromBlock(block *ast.BlockExpression, pkgName string) {
	if g == nil || block == nil {
		return
	}
	for _, stmt := range block.Body {
		g.collectMutableModuleBindingsFromNode(stmt, pkgName)
	}
}

func (g *generator) collectMutableModuleBindingsFromNode(node ast.Node, pkgName string) {
	if g == nil || node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.AssignmentExpression:
		if n == nil {
			return
		}
		if name, _, ok := g.assignmentTargetName(n.Left); ok && g.hasModuleBindingName(pkgName, name) {
			g.markMutableModuleBindingName(pkgName, name)
		}
		g.collectMutableModuleBindingsFromNode(n.Right, pkgName)
	case *ast.BlockExpression:
		g.collectMutableModuleBindingsFromBlock(n, pkgName)
	case *ast.IfExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.IfCondition, pkgName)
		g.collectMutableModuleBindingsFromBlock(n.IfBody, pkgName)
		for _, clause := range n.ElseIfClauses {
			if clause == nil {
				continue
			}
			g.collectMutableModuleBindingsFromNode(clause.Condition, pkgName)
			g.collectMutableModuleBindingsFromBlock(clause.Body, pkgName)
		}
		if n.ElseBody != nil {
			g.collectMutableModuleBindingsFromBlock(n.ElseBody, pkgName)
		}
	case *ast.MatchExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Subject, pkgName)
		for _, clause := range n.Clauses {
			if clause == nil || clause.Body == nil {
				continue
			}
			g.collectMutableModuleBindingsFromNode(clause.Body, pkgName)
		}
	case *ast.FunctionCall:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Callee, pkgName)
		for _, arg := range n.Arguments {
			g.collectMutableModuleBindingsFromNode(arg, pkgName)
		}
	case *ast.MemberAccessExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Object, pkgName)
		g.collectMutableModuleBindingsFromNode(n.Member, pkgName)
	case *ast.ReturnStatement:
		if n != nil {
			g.collectMutableModuleBindingsFromNode(n.Argument, pkgName)
		}
	case *ast.RaiseStatement:
		if n != nil {
			g.collectMutableModuleBindingsFromNode(n.Expression, pkgName)
		}
	case *ast.WhileLoop:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Condition, pkgName)
		g.collectMutableModuleBindingsFromBlock(n.Body, pkgName)
	case *ast.ForLoop:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Iterable, pkgName)
		g.collectMutableModuleBindingsFromBlock(n.Body, pkgName)
	case *ast.RescueExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.MonitoredExpression, pkgName)
		for _, clause := range n.Clauses {
			if clause == nil || clause.Body == nil {
				continue
			}
			g.collectMutableModuleBindingsFromNode(clause.Body, pkgName)
		}
	case *ast.OrElseExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.Expression, pkgName)
		if n.Handler != nil {
			g.collectMutableModuleBindingsFromBlock(n.Handler, pkgName)
		}
	case *ast.EnsureExpression:
		if n == nil {
			return
		}
		g.collectMutableModuleBindingsFromNode(n.TryExpression, pkgName)
		if n.EnsureBlock != nil {
			g.collectMutableModuleBindingsFromBlock(n.EnsureBlock, pkgName)
		}
	case *ast.LoopExpression:
		if n != nil && n.Body != nil {
			g.collectMutableModuleBindingsFromBlock(n.Body, pkgName)
		}
	case *ast.BreakpointExpression:
		if n != nil && n.Body != nil {
			g.collectMutableModuleBindingsFromBlock(n.Body, pkgName)
		}
	case *ast.FunctionDefinition:
		// Nested function bodies should not make outer reads static.
		if n != nil && n.Body != nil {
			g.collectMutableModuleBindingsFromBlock(n.Body, pkgName)
		}
	}
}

func (g *generator) moduleBindingByName(pkgName string, name string) (moduleBinding, bool) {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return moduleBinding{}, false
	}
	for _, binding := range g.moduleBindings[pkgName] {
		if strings.TrimSpace(binding.Name) == name {
			return binding, true
		}
	}
	return moduleBinding{}, false
}

func goTypeForIntegerSuffix(suffix string) (string, bool) {
	switch suffix {
	case "runtime.IntegerI8":
		return "int8", true
	case "runtime.IntegerI16":
		return "int16", true
	case "runtime.IntegerI32":
		return "int32", true
	case "runtime.IntegerI64":
		return "int64", true
	case "runtime.IntegerU8":
		return "uint8", true
	case "runtime.IntegerU16":
		return "uint16", true
	case "runtime.IntegerU32":
		return "uint32", true
	case "runtime.IntegerU64":
		return "uint64", true
	}
	return "", false
}

func (g *generator) evaluatedConstantExpr(pkgName string, name string) (string, string, bool) {
	if g == nil || g.evaluatedConstants == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return "", "", false
	}
	c, ok := g.evaluatedConstants[pkgName+"::"+name]
	if !ok || c == nil || c.val == nil {
		return "", "", false
	}
	goType, ok := goTypeForIntegerSuffix(c.suffix)
	if !ok || goType == "" {
		return "", "", false
	}
	return goType + "(" + c.val.String() + ")", goType, true
}

func (g *generator) compileStaticModuleBindingIdentifier(ctx *compileContext, ident *ast.Identifier, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || ident == nil || strings.TrimSpace(ident.Name) == "" {
		return nil, "", "", false
	}
	if !g.hasModuleBindingName(ctx.packageName, ident.Name) {
		return nil, "", "", false
	}
	if g.isMutableModuleBindingName(ctx.packageName, ident.Name) {
		return nil, "", "", false
	}
	if expr, goType, ok := g.evaluatedConstantExpr(ctx.packageName, ident.Name); ok {
		return g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, goType, expected)
	}
	binding, ok := g.moduleBindingByName(ctx.packageName, ident.Name)
	if !ok || strings.TrimSpace(binding.GoValue) == "" {
		return nil, "", "", false
	}
	if expected == "" || expected == "runtime.Value" {
		return nil, binding.GoValue, "runtime.Value", true
	}
	if expected == "any" {
		return nil, "any(" + binding.GoValue + ")", "any", true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, binding.GoValue, expected)
	if !ok {
		return nil, "", "", false
	}
	return convLines, converted, expected, true
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
