package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileAssignment(ctx *compileContext, assign *ast.AssignmentExpression) ([]string, string, string, bool) {
	if assign == nil {
		ctx.setReason("missing assignment")
		return nil, "", "", false
	}
	if indexTarget, ok := assign.Left.(*ast.IndexExpression); ok {
		if assign.Operator == ast.AssignmentDeclare {
			ctx.setReason("index assignment cannot declare")
			return nil, "", "", false
		}
		op, isCompound := binaryOpForAssignment(assign.Operator)
		valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
		if !ok {
			return nil, "", "", false
		}
		valueRuntime, ok := g.runtimeValueExpr(valueExpr, valueType)
		if !ok {
			ctx.setReason("index assignment value unsupported")
			return nil, "", "", false
		}
		objExpr, objType, ok := g.compileExpr(ctx, indexTarget.Object, "")
		if !ok {
			return nil, "", "", false
		}
		objRuntime, ok := g.runtimeValueExpr(objExpr, objType)
		if !ok {
			ctx.setReason("index assignment target unsupported")
			return nil, "", "", false
		}
		idxExpr, idxType, ok := g.compileExpr(ctx, indexTarget.Index, "")
		if !ok {
			return nil, "", "", false
		}
		idxRuntime, ok := g.runtimeValueExpr(idxExpr, idxType)
		if !ok {
			ctx.setReason("index assignment index unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", idxTemp, idxRuntime))
		if assign.Operator == ast.AssignmentAssign {
			resultTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_index_set(%s, %s, %s)", resultTemp, objTemp, idxTemp, valueTemp))
			lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
			return lines, resultTemp, "runtime.Value", true
		}
		if !isCompound {
			ctx.setReason("unsupported index assignment operator")
			return nil, "", "", false
		}
		currentTemp := ctx.newTemp()
		computedTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_index(%s, %s)", currentTemp, objTemp, idxTemp))
		lines = append(lines, fmt.Sprintf("%s := __able_binary_op(%q, %s, %s)", computedTemp, op, currentTemp, valueTemp))
		lines = append(lines, fmt.Sprintf("%s := __able_index_set(%s, %s, %s)", resultTemp, objTemp, idxTemp, computedTemp))
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return lines, computedTemp, "runtime.Value", true
	}
	if memberTarget, ok := assign.Left.(*ast.MemberAccessExpression); ok {
		if assign.Operator == ast.AssignmentDeclare {
			ctx.setReason("member assignment cannot declare")
			return nil, "", "", false
		}
		if memberTarget.Safe {
			ctx.setReason("safe member assignment unsupported")
			return nil, "", "", false
		}
		objExpr, objType, ok := g.compileExpr(ctx, memberTarget.Object, "")
		if !ok {
			return nil, "", "", false
		}
		if info := g.structInfoByGoName(objType); info != nil {
			if assign.Operator != ast.AssignmentAssign {
				op, ok := binaryOpForAssignment(assign.Operator)
				if !ok {
					ctx.setReason("unsupported member assignment operator")
					return nil, "", "", false
				}
				memberIdent, ok := memberTarget.Member.(*ast.Identifier)
				if !ok || memberIdent == nil || memberIdent.Name == "" {
					ctx.setReason("unsupported member assignment target")
					return nil, "", "", false
				}
				field := g.fieldInfo(info, memberIdent.Name)
				if field == nil {
					ctx.setReason("unknown struct field")
					return nil, "", "", false
				}
				valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, field.GoType, assign.Right)
				if !ok {
					return nil, "", "", false
				}
				if !g.typeMatches(field.GoType, valueType) {
					ctx.setReason("member assignment type mismatch")
					return nil, "", "", false
				}
				valueTemp := ctx.newTemp()
				lines := append([]string{}, valueLines...)
				lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
				currentTemp := ctx.newTemp()
				computedTemp := ctx.newTemp()
				if g.isAddressableMemberObject(memberTarget.Object) {
					objTemp := ctx.newTemp()
					lines = append(lines, fmt.Sprintf("%s := &%s", objTemp, objExpr))
					lines = append(lines, fmt.Sprintf("%s := %s.%s", currentTemp, objTemp, field.GoName))
					expr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, field.GoType, valueTemp, valueType, field.GoType)
					if !ok {
						return nil, "", "", false
					}
					if resultType != field.GoType {
						ctx.setReason("member assignment type mismatch")
						return nil, "", "", false
					}
					lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, expr))
					lines = append(lines, fmt.Sprintf("%s.%s = %s", objTemp, field.GoName, computedTemp))
					return lines, computedTemp, field.GoType, true
				}
				objTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
				lines = append(lines, fmt.Sprintf("%s := %s.%s", currentTemp, objTemp, field.GoName))
				expr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, field.GoType, valueTemp, valueType, field.GoType)
				if !ok {
					return nil, "", "", false
				}
				if resultType != field.GoType {
					ctx.setReason("member assignment type mismatch")
					return nil, "", "", false
				}
				lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, expr))
				lines = append(lines, fmt.Sprintf("%s.%s = %s", objTemp, field.GoName, computedTemp))
				return lines, computedTemp, field.GoType, true
			}
			memberIdent, ok := memberTarget.Member.(*ast.Identifier)
			if !ok || memberIdent == nil || memberIdent.Name == "" {
				ctx.setReason("unsupported member assignment target")
				return nil, "", "", false
			}
			field := g.fieldInfo(info, memberIdent.Name)
			if field == nil {
				ctx.setReason("unknown struct field")
				return nil, "", "", false
			}
			valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, field.GoType, assign.Right)
			if !ok {
				return nil, "", "", false
			}
			if !g.typeMatches(field.GoType, valueType) {
				ctx.setReason("member assignment type mismatch")
				return nil, "", "", false
			}
			valueTemp := ctx.newTemp()
			lines := append([]string{}, valueLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
			targetExpr := objExpr
			if !g.isAddressableMemberObject(memberTarget.Object) {
				objTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
				targetExpr = objTemp
			}
			lines = append(lines, fmt.Sprintf("%s.%s = %s", targetExpr, field.GoName, valueTemp))
			return lines, valueTemp, field.GoType, true
		}
		if g.typeCategory(objType) != "runtime" {
			ctx.setReason("unsupported member assignment target")
			return nil, "", "", false
		}
		valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
		if !ok {
			return nil, "", "", false
		}
		valueRuntime, ok := g.runtimeValueExpr(valueExpr, valueType)
		if !ok {
			ctx.setReason("member assignment value unsupported")
			return nil, "", "", false
		}
		memberRuntime, ok := g.memberAssignmentRuntimeValue(ctx, memberTarget.Member)
		if !ok {
			ctx.setReason("unsupported member assignment target")
			return nil, "", "", false
		}
		objRuntime, ok := g.runtimeValueExpr(objExpr, objType)
		if !ok {
			ctx.setReason("member assignment target unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		objTemp := ctx.newTemp()
		memberTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberRuntime))
		if assign.Operator == ast.AssignmentAssign {
			lines = append(lines, fmt.Sprintf("%s := __able_member_set(%s, %s, %s)", resultTemp, objTemp, memberTemp, valueTemp))
			lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
			return lines, resultTemp, "runtime.Value", true
		}
		op, ok := binaryOpForAssignment(assign.Operator)
		if !ok {
			ctx.setReason("unsupported member assignment operator")
			return nil, "", "", false
		}
		currentTemp := ctx.newTemp()
		computedTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_member_get(%s, %s)", currentTemp, objTemp, memberTemp))
		lines = append(lines, fmt.Sprintf("%s := __able_binary_op(%q, %s, %s)", computedTemp, op, currentTemp, valueTemp))
		lines = append(lines, fmt.Sprintf("%s := __able_member_set(%s, %s, %s)", resultTemp, objTemp, memberTemp, computedTemp))
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return lines, computedTemp, "runtime.Value", true
	}
	if assign.Operator != ast.AssignmentDeclare && assign.Operator != ast.AssignmentAssign {
		op, ok := binaryOpForAssignment(assign.Operator)
		if !ok {
			ctx.setReason("unsupported assignment operator")
			return nil, "", "", false
		}
		name, typeAnnotation, ok := g.assignmentTargetName(assign.Left)
		if !ok {
			ctx.setReason("unsupported assignment target")
			return nil, "", "", false
		}
		if name == "" {
			ctx.setReason("missing assignment identifier")
			return nil, "", "", false
		}
		existing, exists := ctx.lookup(name)
		if !exists {
			ctx.setReason("compound assignment requires existing binding")
			return nil, "", "", false
		}
		goType := existing.GoType
		if typeAnnotation != nil {
			mapped, ok := g.mapTypeExpression(typeAnnotation)
			if !ok {
				ctx.setReason("unsupported type annotation")
				return nil, "", "", false
			}
			if mapped != goType {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
		}
		valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, goType, assign.Right)
		if !ok {
			return nil, "", "", false
		}
		if !g.typeMatches(goType, valueType) {
			ctx.setReason("assignment type mismatch")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		currentTemp := ctx.newTemp()
		computedTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
		lines = append(lines, fmt.Sprintf("%s := %s", currentTemp, existing.GoName))
		expr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, goType, valueTemp, valueType, goType)
		if !ok {
			return nil, "", "", false
		}
		if resultType != goType {
			ctx.setReason("assignment type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, expr))
		lines = append(lines, fmt.Sprintf("%s = %s", existing.GoName, computedTemp))
		return lines, computedTemp, goType, true
	}
	name, typeAnnotation, ok := g.assignmentTargetName(assign.Left)
	if !ok {
		ctx.setReason("unsupported assignment target")
		return nil, "", "", false
	}
	if name == "" {
		ctx.setReason("missing assignment identifier")
		return nil, "", "", false
	}
	existing, exists := ctx.lookup(name)
	_, currentExists := ctx.lookupCurrent(name)
	if assign.Operator == ast.AssignmentDeclare && currentExists {
		ctx.setReason(":= requires new binding")
		return nil, "", "", false
	}
	declaring := assign.Operator == ast.AssignmentDeclare || !exists
	var goType string
	if typeAnnotation != nil {
		mapped, ok := g.mapTypeExpression(typeAnnotation)
		if !ok {
			ctx.setReason("unsupported type annotation")
			return nil, "", "", false
		}
		goType = mapped
		if !declaring && exists && existing.GoType != goType {
			ctx.setReason("assignment type mismatch")
			return nil, "", "", false
		}
	}
	if !declaring && goType == "" && exists {
		goType = existing.GoType
	}
	var expr string
	var exprLines []string
	if goType != "" {
		compiledLines, compiled, _, ok := g.compileTailExpression(ctx, goType, assign.Right)
		if !ok {
			return nil, "", "", false
		}
		exprLines = compiledLines
		expr = compiled
	} else {
		compiledLines, compiled, inferredType, ok := g.compileTailExpression(ctx, "", assign.Right)
		if !ok {
			return nil, "", "", false
		}
		exprLines = compiledLines
		expr = compiled
		goType = inferredType
		if goType == "" {
			ctx.setReason("could not infer assignment type")
			return nil, "", "", false
		}
	}
	goName := existing.GoName
	if declaring {
		goName = sanitizeIdent(name)
		ctx.locals[name] = paramInfo{Name: name, GoName: goName, GoType: goType}
	}
	line := ""
	if declaring {
		line = fmt.Sprintf("var %s %s = %s", goName, goType, expr)
	} else {
		line = fmt.Sprintf("%s = %s", goName, expr)
	}
	lines := append(exprLines, line)
	return lines, goName, goType, true
}

func binaryOpForAssignment(op ast.AssignmentOperator) (string, bool) {
	switch op {
	case ast.AssignmentAdd:
		return "+", true
	case ast.AssignmentSub:
		return "-", true
	case ast.AssignmentMul:
		return "*", true
	case ast.AssignmentDiv:
		return "/", true
	case ast.AssignmentMod:
		return "%", true
	case ast.AssignmentBitAnd:
		return ".&", true
	case ast.AssignmentBitOr:
		return ".|", true
	case ast.AssignmentBitXor:
		return ".^", true
	case ast.AssignmentShiftL:
		return ".<<", true
	case ast.AssignmentShiftR:
		return ".>>", true
	default:
		return "", false
	}
}

func (g *generator) assignmentTargetName(target ast.AssignmentTarget) (string, ast.TypeExpression, bool) {
	switch t := target.(type) {
	case *ast.Identifier:
		if t == nil {
			return "", nil, false
		}
		return t.Name, nil, true
	case *ast.TypedPattern:
		if t == nil {
			return "", nil, false
		}
		if ident, ok := t.Pattern.(*ast.Identifier); ok && ident != nil {
			return ident.Name, t.TypeAnnotation, true
		}
		return "", nil, false
	default:
		return "", nil, false
	}
}

func (g *generator) isAddressableMemberObject(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e != nil
	case *ast.MemberAccessExpression:
		if e == nil || e.Safe {
			return false
		}
		return g.isAddressableMemberObject(e.Object)
	default:
		return false
	}
}

func (g *generator) memberAssignmentRuntimeValue(ctx *compileContext, member ast.Expression) (string, bool) {
	switch m := member.(type) {
	case *ast.Identifier:
		if m == nil || m.Name == "" {
			return "", false
		}
		return fmt.Sprintf("bridge.ToString(%q)", m.Name), true
	case *ast.IntegerLiteral:
		expr, goType, ok := g.compileIntegerLiteral(ctx, m, "")
		if !ok {
			return "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			return "", false
		}
		return valueExpr, true
	default:
		return "", false
	}
}
