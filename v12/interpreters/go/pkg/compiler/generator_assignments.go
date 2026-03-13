package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileAssignment(ctx *compileContext, assign *ast.AssignmentExpression) ([]string, string, string, bool) {
	if assign == nil {
		ctx.setReason("missing assignment")
		return nil, "", "", false
	}
	if implicitTarget, ok := assign.Left.(*ast.ImplicitMemberExpression); ok {
		if ctx == nil || !ctx.hasImplicitReceiver || ctx.implicitReceiver.Name == "" {
			ctx.setReason("implicit member assignment requires receiver")
			return nil, "", "", false
		}
		receiver := ast.NewIdentifier(ctx.implicitReceiver.Name)
		memberExpr := ast.NewMemberAccessExpression(receiver, implicitTarget.Member)
		synthetic := ast.NewAssignmentExpression(assign.Operator, memberExpr, assign.Right)
		return g.compileAssignment(ctx, synthetic)
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
		valueConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("index assignment value unsupported")
			return nil, "", "", false
		}
		objLines, objExpr, objType, ok := g.compileExprLines(ctx, indexTarget.Object, "")
		if !ok {
			return nil, "", "", false
		}
		if assign.Operator == ast.AssignmentAssign && g.isArrayStructType(objType) {
			idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, indexTarget.Index, "")
			if !ok {
				return nil, "", "", false
			}
			valueTemp := ctx.newTemp()
			objTemp := ctx.newTemp()
			idxTemp := ctx.newTemp()
			indexTemp := ctx.newTemp()
			lengthTemp := ctx.newTemp()
			resultTemp := ctx.newTemp()
			lines := append([]string{}, valueLines...)
			lines = append(lines, valueConvLines...)
			lines = append(lines, objLines...)
			lines = append(lines, idxLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
			lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
			lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
			if !ok {
				ctx.setReason("index assignment index unsupported")
				return nil, "", "", false
			}
			lines = append(lines, fmt.Sprintf("%s := len(%s.Elements)", lengthTemp, objTemp))
			lines = append(lines, fmt.Sprintf("%s := %s", resultTemp, valueTemp))
			lines = append(lines, fmt.Sprintf("if %s < 0 || %s >= %s { %s = __able_index_error(%s, %s) } else { %s.Elements[%s] = %s }", indexTemp, indexTemp, lengthTemp, resultTemp, indexTemp, lengthTemp, objTemp, indexTemp, valueTemp))
			lines = append(lines, fmt.Sprintf("__able_struct_Array_sync(%s)", objTemp))
			return lines, resultTemp, "runtime.Value", true
		}
		objConvLines, objRuntime, ok := g.runtimeValueLines(ctx, objExpr, objType)
		if !ok {
			ctx.setReason("index assignment target unsupported")
			return nil, "", "", false
		}
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, indexTarget.Index, "")
		if !ok {
			return nil, "", "", false
		}
		idxConvLines, idxRuntime, ok := g.runtimeValueLines(ctx, idxExpr, idxType)
		if !ok {
			ctx.setReason("index assignment index unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		objTemp := ctx.newTemp()
		idxTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, valueConvLines...)
		lines = append(lines, objLines...)
		lines = append(lines, objConvLines...)
		lines = append(lines, idxLines...)
		lines = append(lines, idxConvLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", idxTemp, idxRuntime))
		if assign.Operator == ast.AssignmentAssign {
			resultTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s, %s := __able_index_set(%s, %s, %s)", resultTemp, controlTemp, objTemp, idxTemp, valueTemp))
			controlLines, ok := g.controlCheckLines(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, controlLines...)
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
		currentControlTemp := ctx.newTemp()
		computedControlTemp := ctx.newTemp()
		resultControlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := __able_index(%s, %s)", currentTemp, currentControlTemp, objTemp, idxTemp))
		controlLines, ok := g.controlCheckLines(ctx, currentControlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("%s, %s := __able_binary_op(%q, %s, %s)", computedTemp, computedControlTemp, op, currentTemp, valueTemp))
		controlLines, ok = g.controlCheckLines(ctx, computedControlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("%s, %s := __able_index_set(%s, %s, %s)", resultTemp, resultControlTemp, objTemp, idxTemp, computedTemp))
		controlLines, ok = g.controlCheckLines(ctx, resultControlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return lines, computedTemp, "runtime.Value", true
	}
	if pattern, ok := assign.Left.(ast.Pattern); ok {
		if !isSimpleAssignmentPattern(pattern) {
			return g.compilePatternAssignment(ctx, assign, pattern)
		}
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
		objLines, objExpr, objType, ok := g.compileExprLines(ctx, memberTarget.Object, "")
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
				field, ok := g.structFieldForMember(info, memberTarget.Member)
				if !ok {
					ctx.setReason("unsupported member assignment target")
					return nil, "", "", false
				}
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
				lines = append(lines, objLines...)
				lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
				currentTemp := ctx.newTemp()
				computedTemp := ctx.newTemp()
				needsAddr := true
				if baseName, ok := g.structBaseName(objType); ok && objType != baseName {
					needsAddr = false
				}
				nodeName := g.diagNodeName(assign, "*ast.AssignmentExpression", "assign")
				if g.isAddressableMemberObject(memberTarget.Object) && needsAddr {
					objTemp := ctx.newTemp()
					lines = append(lines, fmt.Sprintf("%s := &%s", objTemp, objExpr))
					lines = append(lines, fmt.Sprintf("%s := %s.%s", currentTemp, objTemp, field.GoName))
					opLines, opExpr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, field.GoType, valueTemp, valueType, field.GoType, nodeName)
					if !ok {
						return nil, "", "", false
					}
					if resultType != field.GoType {
						ctx.setReason("member assignment type mismatch")
						return nil, "", "", false
					}
					lines = append(lines, opLines...)
					lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, opExpr))
					lines = append(lines, fmt.Sprintf("%s.%s = %s", objTemp, field.GoName, computedTemp))
					return lines, computedTemp, field.GoType, true
				}
				objTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objExpr))
				lines = append(lines, fmt.Sprintf("%s := %s.%s", currentTemp, objTemp, field.GoName))
				opLines, opExpr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, field.GoType, valueTemp, valueType, field.GoType, nodeName)
				if !ok {
					return nil, "", "", false
				}
				if resultType != field.GoType {
					ctx.setReason("member assignment type mismatch")
					return nil, "", "", false
				}
				lines = append(lines, opLines...)
				lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, opExpr))
				lines = append(lines, fmt.Sprintf("%s.%s = %s", objTemp, field.GoName, computedTemp))
				return lines, computedTemp, field.GoType, true
			}
			field, ok := g.structFieldForMember(info, memberTarget.Member)
			if !ok {
				ctx.setReason("unsupported member assignment target")
				return nil, "", "", false
			}
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
			lines = append(lines, objLines...)
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
		objCategory := g.typeCategory(objType)
		if objCategory != "runtime" && objCategory != "any" {
			ctx.setReason("unsupported member assignment target")
			return nil, "", "", false
		}
		// Invalidate CSE extraction cache — __able_member_set modifies
		// the underlying struct, making any cached extraction stale.
		if ctx.originExtractions != nil {
			if objIdent, ok := memberTarget.Object.(*ast.Identifier); ok && objIdent != nil {
				delete(ctx.originExtractions, objIdent.Name)
			}
		}
		valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
		if !ok {
			return nil, "", "", false
		}
		valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("member assignment value unsupported")
			return nil, "", "", false
		}
		memberRuntime, ok := g.memberAssignmentRuntimeValue(ctx, memberTarget.Member)
		if !ok {
			ctx.setReason("unsupported member assignment target")
			return nil, "", "", false
		}
		objConvLines, objRuntime, ok := g.runtimeValueLines(ctx, objExpr, objType)
		if !ok {
			ctx.setReason("member assignment target unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		objTemp := ctx.newTemp()
		memberTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, valConvLines...)
		lines = append(lines, objLines...)
		lines = append(lines, objConvLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objRuntime))
		lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberRuntime))
		// Helper to invalidate CSE cache after mutation — the right side may
		// have re-populated the cache with a pre-mutation extraction.
		invalidateAfterMemberSet := func() {
			if ctx.originExtractions != nil {
				if objIdent, ok := memberTarget.Object.(*ast.Identifier); ok && objIdent != nil {
					delete(ctx.originExtractions, objIdent.Name)
				}
			}
		}
		if assign.Operator == ast.AssignmentAssign {
			lines, resultTemp, ok = g.appendRuntimeMemberSetControlLines(ctx, lines, objTemp, memberTemp, valueTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
			invalidateAfterMemberSet()
			return lines, resultTemp, "runtime.Value", true
		}
		op, ok := binaryOpForAssignment(assign.Operator)
		if !ok {
			ctx.setReason("unsupported member assignment operator")
			return nil, "", "", false
		}
		computedTemp := ctx.newTemp()
		computedControlTemp := ctx.newTemp()
		lines, currentTemp, ok := g.appendRuntimeMemberGetControlLines(ctx, lines, objTemp, memberTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("%s, %s := __able_binary_op(%q, %s, %s)", computedTemp, computedControlTemp, op, currentTemp, valueTemp))
		controlLines, ok := g.controlCheckLines(ctx, computedControlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines, resultTemp, ok = g.appendRuntimeMemberSetControlLines(ctx, lines, objTemp, memberTemp, computedTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		invalidateAfterMemberSet()
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
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, typeAnnotation)
			if !ok {
				ctx.setReason("unsupported type annotation")
				return nil, "", "", false
			}
			if mapped != goType {
				ctx.setReason("assignment type mismatch")
				return nil, "", "", false
			}
		}
		expectedTypeExpr := typeAnnotation
		if expectedTypeExpr == nil {
			expectedTypeExpr = existing.TypeExpr
		}
		previousExpectedTypeExpr := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = expectedTypeExpr
		valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, goType, assign.Right)
		ctx.expectedTypeExpr = previousExpectedTypeExpr
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
		nodeName := g.diagNodeName(assign, "*ast.AssignmentExpression", "assign")
		opLines, opExpr, resultType, ok := g.compileBinaryOperation(ctx, op, currentTemp, goType, valueTemp, valueType, goType, nodeName)
		if !ok {
			return nil, "", "", false
		}
		if resultType != goType {
			ctx.setReason("assignment type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, opLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", computedTemp, opExpr))
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
	useEnvSet := assign.Operator == ast.AssignmentAssign && !exists && (typeAnnotation == nil || g.hasModuleBindingName(ctx.packageName, name))
	var goType string
	if typeAnnotation != nil {
		mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, typeAnnotation)
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
		expectedTypeExpr := typeAnnotation
		if expectedTypeExpr == nil && exists {
			expectedTypeExpr = existing.TypeExpr
		}
		previousExpectedTypeExpr := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = expectedTypeExpr
		compiledLines, compiled, _, ok := g.compileTailExpression(ctx, goType, assign.Right)
		ctx.expectedTypeExpr = previousExpectedTypeExpr
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
	assignmentTypeExpr := typeAnnotation
	if assignmentTypeExpr == nil && exists {
		assignmentTypeExpr = existing.TypeExpr
	}
	if ifaceType, ok := g.interfaceTypeExpr(assignmentTypeExpr); ok && goType == "runtime.Value" {
		ifaceLines, coerced, ok := g.interfaceReturnExprLines(ctx, expr, ifaceType, ctx.genericNames)
		if !ok {
			ctx.setReason("unsupported interface assignment coercion")
			return nil, "", "", false
		}
		exprLines = append(exprLines, ifaceLines...)
		expr = coerced
	}
	if useEnvSet {
		valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("env assignment value unsupported")
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(assign, "*ast.AssignmentExpression", "assign")
		resultTemp := ctx.newTemp()
		lines := append([]string{}, exprLines...)
		lines = append(lines, valConvLines...)
		lines = append(lines, fmt.Sprintf("%s := __able_env_set(%q, %s, %s)", resultTemp, name, valueRuntime, nodeName))
		return lines, resultTemp, "runtime.Value", true
	}
	originStructType := ""
	goName := existing.GoName
	if declaring {
		goName = sanitizeIdent(name)
		ctx.locals[name] = paramInfo{Name: name, GoName: goName, GoType: goType, TypeExpr: typeAnnotation, OriginGoType: originStructType}
	} else {
		// Invalidate CSE extraction cache on reassignment.
		if ctx.originExtractions != nil {
			delete(ctx.originExtractions, name)
		}
	}
	if !declaring && typeAnnotation != nil {
		updated := existing
		updated.TypeExpr = typeAnnotation
		if ctx.locals == nil {
			ctx.locals = make(map[string]paramInfo)
		}
		ctx.locals[name] = updated
	}
	line := ""
	if declaring {
		line = fmt.Sprintf("var %s %s = %s", goName, goType, expr)
	} else {
		line = fmt.Sprintf("%s = %s", goName, expr)
	}
	lines := append(exprLines, line)
	if typeAnnotation != nil && (goType == "runtime.Value" || goType == "any") {
		typeExpr, ok := g.renderTypeExpression(typeAnnotation)
		if ok {
			g.needsAst = true
			checkOk := ctx.newTemp()
			resultTemp := ctx.newTemp()
			castSubject := goName
			if goType == "any" {
				convTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, goName))
				castSubject = convTemp
			}
			controlTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("_, %s, %s := __able_try_cast(%s, %s)", checkOk, controlTemp, castSubject, typeExpr))
			controlLines, ok := g.controlCheckLines(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, controlLines...)
			lines = append(lines, fmt.Sprintf("var %s runtime.Value", resultTemp))
			lines = append(lines, fmt.Sprintf("if %s { %s = %s } else { %s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"} }", checkOk, resultTemp, castSubject, resultTemp))
			return lines, resultTemp, "runtime.Value", true
		}
	}
	return lines, goName, goType, true
}

func (g *generator) compilePatternAssignment(ctx *compileContext, assign *ast.AssignmentExpression, pattern ast.Pattern) ([]string, string, string, bool) {
	if assign == nil {
		ctx.setReason("missing assignment")
		return nil, "", "", false
	}
	if assign.Operator != ast.AssignmentDeclare && assign.Operator != ast.AssignmentAssign {
		ctx.setReason("compound assignment not supported with patterns")
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", assign.Right)
	if !ok {
		return nil, "", "", false
	}
	mode := patternBindingMode{declare: assign.Operator == ast.AssignmentDeclare}
	if mode.declare {
		newNames := map[string]struct{}{}
		collectPatternBindingNames(pattern, newNames)
		if len(newNames) == 0 {
			ctx.setReason(":= requires new binding")
			return nil, "", "", false
		}
		filtered := map[string]struct{}{}
		for name := range newNames {
			if _, ok := ctx.lookupCurrent(name); !ok {
				filtered[name] = struct{}{}
			}
		}
		if len(filtered) == 0 {
			ctx.setReason(":= requires new binding")
			return nil, "", "", false
		}
		mode.newNames = filtered
	}

	if valueType == "runtime.Value" || valueType == "any" {
		valConvLines, valueRuntime, ok := g.runtimeValueLines(ctx, valueExpr, valueType)
		if !ok {
			ctx.setReason("pattern assignment value unsupported")
			return nil, "", "", false
		}
		valueTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines, valConvLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueRuntime))
		condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, valueTemp, "runtime.Value")
		if !ok {
			return nil, "", "", false
		}
		bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern, valueTemp, "runtime.Value", mode)
		if !ok {
			return nil, "", "", false
		}
		declLines, assignLines := splitPatternBindingLines(bindLines)
		lines = append(lines, declLines...)
		lines = append(lines, condLines...)
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s runtime.Value", resultTemp))
		if cond != "true" {
			lines = append(lines, fmt.Sprintf("if !(%s) { %s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"} } else {", cond, resultTemp))
			lines = append(lines, assignLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, valueTemp))
			lines = append(lines, "}")
		} else {
			lines = append(lines, assignLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, valueTemp))
		}
		return lines, resultTemp, "runtime.Value", true
	}

	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	condLines, cond, ok := g.compileMatchPatternCondition(ctx, pattern, valueTemp, valueType)
	if !ok {
		return nil, "", "", false
	}
	bindLines, ok := g.compileAssignmentPatternBindings(ctx, pattern, valueTemp, valueType, mode)
	if !ok {
		return nil, "", "", false
	}
	declLines, assignLines := splitPatternBindingLines(bindLines)
	lines = append(lines, declLines...)
	lines = append(lines, condLines...)
	resultTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s runtime.Value", resultTemp))
	resultLines, resultExpr, ok := g.runtimeValueLines(ctx, valueTemp, valueType)
	if !ok {
		ctx.setReason("pattern assignment value unsupported")
		return nil, "", "", false
	}
	if cond != "true" {
		lines = append(lines, fmt.Sprintf("if !(%s) { %s = runtime.ErrorValue{Message: \"pattern assignment mismatch\"} } else {", cond, resultTemp))
		lines = append(lines, assignLines...)
		lines = append(lines, resultLines...)
		lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, resultExpr))
		lines = append(lines, "}")
	} else {
		lines = append(lines, assignLines...)
		lines = append(lines, resultLines...)
		lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, resultExpr))
	}
	return lines, resultTemp, "runtime.Value", true
}

func splitPatternBindingLines(lines []string) ([]string, []string) {
	if len(lines) == 0 {
		return nil, nil
	}
	decls := make([]string, 0, len(lines))
	assigns := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var ") {
			if idx := strings.Index(trimmed, " = "); idx != -1 {
				decl := strings.TrimSpace(trimmed[:idx])
				expr := strings.TrimSpace(trimmed[idx+3:])
				fields := strings.Fields(decl)
				if len(fields) >= 2 {
					name := fields[1]
					decls = append(decls, decl)
					assigns = append(assigns, fmt.Sprintf("%s = %s", name, expr))
					continue
				}
			}
			decls = append(decls, line)
			continue
		}
		if strings.HasPrefix(trimmed, "_ = ") || strings.HasPrefix(trimmed, "_=") {
			decls = append(decls, line)
			continue
		}
		assigns = append(assigns, line)
	}
	return decls, assigns
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

func isSimpleAssignmentPattern(pattern ast.Pattern) bool {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return true
	case *ast.TypedPattern:
		if p == nil || p.Pattern == nil {
			return false
		}
		if _, ok := p.Pattern.(*ast.Identifier); ok {
			return true
		}
	}
	return false
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
