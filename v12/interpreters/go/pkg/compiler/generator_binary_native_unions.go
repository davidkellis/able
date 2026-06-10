package compiler

import "fmt"

func (g *generator) compileStaticNativeUnionEqualityComparison(ctx *compileContext, op string, leftExpr string, leftType string, rightExpr string, rightType string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || leftType == "" || rightType == "" || leftType != rightType {
		return nil, "", "", false
	}
	if op != "==" && op != "!=" {
		return nil, "", "", false
	}
	if expected != "" && expected != "bool" {
		return nil, "", "", false
	}
	union := g.nativeUnionInfoForGoType(leftType)
	if union == nil || !g.nativeUnionSupportsStaticEquality(union) {
		return nil, "", "", false
	}

	resultTemp := ctx.newTemp()
	matchedTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := false", resultTemp),
		fmt.Sprintf("%s := false", matchedTemp),
	}
	for _, member := range union.Members {
		if member == nil || member.UnwrapHelper == "" || !g.nativeUnionMemberSupportsStaticEquality(member) {
			return nil, "", "", false
		}
		leftTemp := ctx.newTemp()
		leftOK := ctx.newTemp()
		rightTemp := ctx.newTemp()
		rightOK := ctx.newTemp()
		eqExpr, ok := g.nativeUnionMemberStaticEqualityExpr(member, leftTemp, rightTemp)
		if !ok || eqExpr == "" {
			return nil, "", "", false
		}
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(%s)", leftTemp, leftOK, member.UnwrapHelper, leftExpr),
			fmt.Sprintf("%s, %s := %s(%s)", rightTemp, rightOK, member.UnwrapHelper, rightExpr),
			fmt.Sprintf("_ = %s", leftTemp),
			fmt.Sprintf("_ = %s", rightTemp),
			fmt.Sprintf("if !%s && %s && %s {", matchedTemp, leftOK, rightOK),
			fmt.Sprintf("\t%s = %s", resultTemp, eqExpr),
			fmt.Sprintf("\t%s = true", matchedTemp),
			"}",
		)
	}
	if op == "!=" {
		lines = append(lines, fmt.Sprintf("if %s { %s = !%s } else { %s = true }", matchedTemp, resultTemp, resultTemp, resultTemp))
	}
	return lines, resultTemp, "bool", true
}

func (g *generator) nativeUnionSupportsStaticEquality(info *nativeUnionInfo) bool {
	if g == nil || info == nil || len(info.Members) == 0 {
		return false
	}
	for _, member := range info.Members {
		if !g.nativeUnionMemberSupportsStaticEquality(member) {
			return false
		}
	}
	return true
}

func (g *generator) nativeUnionMemberSupportsStaticEquality(member *nativeUnionMember) bool {
	if g == nil || member == nil || member.GoType == "" {
		return false
	}
	if member.GoType == "runtime.ErrorValue" {
		// Error values do not compare equal in the reference interpreter,
		// including two values carrying the same error payload. Keeping that
		// rule in the native union comparison avoids boxing an otherwise
		// concrete Result solely because ErrorValue contains a Go map.
		return true
	}
	if g.isEqualityComparable(member.GoType) || member.GoType == "struct{}" {
		return true
	}
	return g.isSingletonNominalCarrierType(member.GoType)
}

func (g *generator) nativeUnionMemberStaticEqualityExpr(member *nativeUnionMember, leftExpr string, rightExpr string) (string, bool) {
	if g == nil || member == nil || leftExpr == "" || rightExpr == "" {
		return "", false
	}
	switch {
	case member.GoType == "runtime.ErrorValue":
		return "false", true
	case g.isEqualityComparable(member.GoType):
		return fmt.Sprintf("(%s == %s)", leftExpr, rightExpr), true
	case member.GoType == "struct{}":
		return "true", true
	case g.isSingletonNominalCarrierType(member.GoType):
		return "true", true
	default:
		return "", false
	}
}

func (g *generator) isSingletonNominalCarrierType(goType string) bool {
	if g == nil || goType == "" || g.typeCategory(goType) != "struct" {
		return false
	}
	info := g.structInfoByGoName(goType)
	if info == nil {
		return false
	}
	return len(info.Fields) == 0
}
