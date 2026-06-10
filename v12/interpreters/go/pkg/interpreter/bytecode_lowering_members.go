package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeStoreNamedStructMemberPlan(ctx *bytecodeLoweringContext, ip int, member *ast.MemberAccessExpression) {
	if ctx == nil || member == nil || ip < 0 {
		return
	}
	memberName := bytecodeIdentifierMemberName(member.Member)
	if memberName == "" {
		return
	}
	def := bytecodeNominalNamedStructDefinitionForExpr(ctx, member.Object)
	if def == nil {
		return
	}
	fieldIndex, ok := bytecodeNamedStructFieldIndex(def, memberName)
	if !ok {
		return
	}
	bytecodeStoreNamedStructFieldMemberPlan(ctx, ip, def, fieldIndex)
}

func bytecodeLoadSlotStructFieldInstruction(ctx *bytecodeLoweringContext, member *ast.MemberAccessExpression) (bytecodeInstruction, bytecodeNamedStructMemberPlan, bool) {
	if ctx == nil || member == nil || member.Safe {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	objectIdent, ok := member.Object.(*ast.Identifier)
	if !ok || objectIdent == nil {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	slot, ok := ctx.lookupSlot(objectIdent.Name)
	if !ok {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	fieldName := bytecodeIdentifierMemberName(member.Member)
	if fieldName == "" {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	def := ctx.slotExactStructDef(slot)
	if def == nil {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	fieldIndex, ok := bytecodeNamedStructFieldIndex(def, fieldName)
	if !ok {
		return bytecodeInstruction{}, bytecodeNamedStructMemberPlan{}, false
	}
	instr := bytecodeInstruction{
		op:     bytecodeOpLoadSlotStructField,
		target: slot,
		name:   fieldName,
		node:   member,
	}
	plan := bytecodeNamedStructMemberPlan{
		definition: def,
		fieldIndex: fieldIndex,
	}
	return instr, plan, true
}

func bytecodeNominalNamedStructDefinitionForExpr(ctx *bytecodeLoweringContext, expr ast.Expression) *runtime.StructDefinitionValue {
	if ctx == nil || expr == nil {
		return nil
	}
	switch n := expr.(type) {
	case *ast.Identifier:
		if slot, ok := ctx.lookupSlot(n.Name); ok {
			return ctx.slotExactStructDef(slot)
		}
	case *ast.TypeCastExpression:
		return bytecodeNominalNamedStructDefinitionForTypeExpr(ctx, n.TargetType)
	case *ast.StructLiteral:
		if n.StructType != nil {
			return bytecodeNamedStructLiteralDefinition(ctx, n.StructType.Name)
		}
	}
	return nil
}

func bytecodeNominalNamedStructDefinitionForTypeExpr(ctx *bytecodeLoweringContext, typeExpr ast.TypeExpression) *runtime.StructDefinitionValue {
	name := bytecodeNominalNamedStructTypeName(ctx, typeExpr)
	if name == "" {
		return nil
	}
	if fastNamedStructTypeNameIsNonNominal(nil, name) {
		return nil
	}
	return bytecodeNamedStructLiteralDefinition(ctx, name)
}

func bytecodeNominalNamedStructTypeName(ctx *bytecodeLoweringContext, typeExpr ast.TypeExpression) string {
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return ""
		}
		switch t.Name.Name {
		case "Self", "SelfType":
			if ctx != nil && ctx.methodSet != nil && ctx.methodSet.TargetType != nil {
				return bytecodeNominalNamedStructTypeName(ctx, ctx.methodSet.TargetType)
			}
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		if t == nil {
			return ""
		}
		return bytecodeNominalNamedStructTypeName(ctx, t.Base)
	default:
		return ""
	}
}

func bytecodeNamedStructDefinitionForTypeExpr(ctx *bytecodeLoweringContext, typeExpr ast.TypeExpression) *runtime.StructDefinitionValue {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return nil
	}
	return bytecodeNamedStructLiteralDefinition(ctx, simple.Name.Name)
}

func bytecodeNamedStructFieldIndex(def *runtime.StructDefinitionValue, name string) (int, bool) {
	if def == nil || def.Node == nil || name == "" {
		return 0, false
	}
	for idx, field := range def.Node.Fields {
		if field == nil || field.Name == nil {
			continue
		}
		if field.Name.Name == name {
			return idx, true
		}
	}
	return 0, false
}
