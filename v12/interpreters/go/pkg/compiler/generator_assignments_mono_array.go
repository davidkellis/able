package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) monoArrayMetadataFieldType(memberName string) (string, bool) {
	switch memberName {
	case "length", "capacity":
		return "int32", true
	case "storage_handle":
		return "int64", true
	default:
		return "", false
	}
}

func (g *generator) compileMonoArrayMetadataAssignment(
	ctx *compileContext,
	assign *ast.AssignmentExpression,
	target *ast.MemberAccessExpression,
	objLines []string,
	objExpr string,
	objType string,
) ([]string, string, string, bool) {
	if g == nil || ctx == nil || assign == nil || target == nil || objExpr == "" || !g.isMonoArrayType(objType) {
		return nil, "", "", false
	}
	memberName := g.memberName(target.Member)
	fieldType, ok := g.monoArrayMetadataFieldType(memberName)
	if !ok {
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, fieldType, assign.Right)
	if !ok {
		return nil, "", "", false
	}
	lines := append([]string{}, valueLines...)
	lines = append(lines, objLines...)
	objTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
	if assign.Operator != ast.AssignmentAssign {
		op, ok := binaryOpForAssignment(assign.Operator)
		if !ok {
			ctx.setReason("unsupported member assignment operator")
			return nil, "", "", false
		}
		currentExpr, currentType, ok := g.monoArrayFieldAccessExpr(objTemp, objType, memberName)
		if !ok {
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(assign, "*ast.AssignmentExpression", "assign")
		opLines, opExpr, resultType, ok := g.compileBinaryOperation(ctx, op, currentExpr, currentType, valueExpr, valueType, fieldType, nodeName)
		if !ok {
			return nil, "", "", false
		}
		if !g.typeMatches(fieldType, resultType) {
			ctx.setReason("member assignment type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, opLines...)
		valueExpr = opExpr
		valueType = resultType
	}
	if !g.typeMatches(fieldType, valueType) {
		ctx.setReason("member assignment type mismatch")
		return nil, "", "", false
	}
	valueTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	switch memberName {
	case "length":
		targetLen := ctx.newTemp()
		storedTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := int(%s)", targetLen, valueTemp),
			fmt.Sprintf("if %s < 0 { %s = 0 }", targetLen, targetLen),
			fmt.Sprintf("if %s <= len(%s.Elements) {", targetLen, objTemp),
			fmt.Sprintf("\t%s.Elements = %s.Elements[:%s]", objTemp, objTemp, targetLen),
			"} else {",
			fmt.Sprintf("\tfor len(%s.Elements) < %s { %s.Elements = append(%s.Elements, %s) }", objTemp, targetLen, objTemp, objTemp, g.staticArrayZeroValueExpr(objType)),
			"}",
			fmt.Sprintf("%s := int32(%s)", storedTemp, targetLen),
		)
		return lines, storedTemp, "int32", true
	case "capacity":
		targetCap := ctx.newTemp()
		storedTemp := ctx.newTemp()
		reservedTemp := ctx.newTemp()
		elemType := g.staticArrayElemGoType(objType)
		lines = append(lines,
			fmt.Sprintf("%s := int(%s)", targetCap, valueTemp),
			fmt.Sprintf("if %s < 0 { %s = 0 }", targetCap, targetCap),
			fmt.Sprintf("if %s < len(%s.Elements) { %s = len(%s.Elements) }", targetCap, objTemp, targetCap, objTemp),
			fmt.Sprintf("if %s != cap(%s.Elements) {", targetCap, objTemp),
			fmt.Sprintf("\t%s := make([]%s, len(%s.Elements), %s)", reservedTemp, elemType, objTemp, targetCap),
			fmt.Sprintf("\tcopy(%s, %s.Elements)", reservedTemp, objTemp),
			fmt.Sprintf("\t%s.Elements = %s", objTemp, reservedTemp),
			"}",
			fmt.Sprintf("%s := int32(%s)", storedTemp, targetCap),
		)
		return lines, storedTemp, "int32", true
	case "storage_handle":
		lines = append(lines,
			fmt.Sprintf("_ = %s", objTemp),
			fmt.Sprintf("_ = %s", valueTemp),
		)
		return lines, "int64(0)", "int64", true
	default:
		return nil, "", "", false
	}
}
