package flow

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
)

func typeNameOrDynamic(expression ast.TypeExpression) string {
	if expression == nil {
		return "dynamic"
	}
	return typeName(expression)
}

func typeName(expression ast.TypeExpression) string {
	switch value := expression.(type) {
	case *ast.SimpleTypeExpression:
		if value.Name != nil {
			return value.Name.Name
		}
	case *ast.GenericTypeExpression:
		arguments := make([]string, len(value.Arguments))
		for index, argument := range value.Arguments {
			arguments[index] = typeName(argument)
		}
		return typeName(value.Base) + "<" + strings.Join(arguments, ",") + ">"
	case *ast.NullableTypeExpression:
		return typeName(value.InnerType) + "?"
	case *ast.ResultTypeExpression:
		return typeName(value.InnerType) + "!"
	case *ast.UnionTypeExpression:
		members := make([]string, len(value.Members))
		for index, member := range value.Members {
			members[index] = typeName(member)
		}
		return strings.Join(members, "|")
	case *ast.WildcardTypeExpression:
		return "_"
	}
	return "dynamic"
}

func collectImports(module *ast.Module) map[string]string {
	result := make(map[string]string)
	for _, statement := range module.Imports {
		parts := make([]string, 0, len(statement.PackagePath))
		for _, part := range statement.PackagePath {
			if part != nil {
				parts = append(parts, part.Name)
			}
		}
		path := strings.Join(parts, ".")
		qualifier := ""
		if statement.Alias != nil {
			qualifier = statement.Alias.Name
		} else if len(parts) != 0 {
			qualifier = parts[len(parts)-1]
		}
		if qualifier != "" {
			result[qualifier] = path
		}
		for _, selector := range statement.Selectors {
			if selector == nil || selector.Name == nil {
				continue
			}
			name := selector.Name.Name
			if selector.Alias != nil {
				name = selector.Alias.Name
			}
			result[name] = path
		}
	}
	return result
}

func collectFunctionReturns(module *ast.Module) map[string]string {
	result := make(map[string]string)
	for _, statement := range module.Body {
		function, ok := statement.(*ast.FunctionDefinition)
		if !ok || function.ID == nil {
			continue
		}
		result[function.ID.Name] = typeNameOrDynamic(function.ReturnType)
	}
	return result
}

func collectGlobalTypes(module *ast.Module) map[string]string {
	result := make(map[string]string)
	for _, statement := range module.Body {
		assignment, ok := statement.(*ast.AssignmentExpression)
		if !ok {
			continue
		}
		name, annotation, ok := assignmentTargetInfo(assignment.Left)
		if !ok {
			continue
		}
		if annotation != nil {
			result[name] = typeName(annotation)
			continue
		}
		result[name] = literalTypeName(assignment.Right)
	}
	return result
}

func assignmentTargetInfo(target ast.AssignmentTarget) (string, ast.TypeExpression, bool) {
	switch value := target.(type) {
	case *ast.Identifier:
		return value.Name, nil, true
	case *ast.TypedPattern:
		name, ok := patternBindingName(value.Pattern)
		return name, value.TypeAnnotation, ok
	default:
		return "", nil, false
	}
}

func patternBindingName(pattern ast.Pattern) (string, bool) {
	identifier, ok := pattern.(*ast.Identifier)
	if !ok || identifier == nil {
		return "", false
	}
	return identifier.Name, true
}

func literalTypeName(expression ast.Expression) string {
	switch value := expression.(type) {
	case *ast.IntegerLiteral:
		return integerTypeName(value.IntegerType)
	case *ast.FloatLiteral:
		if value.FloatType != nil && *value.FloatType == ast.FloatTypeF32 {
			return "f32"
		}
		return "f64"
	case *ast.StringLiteral:
		return "String"
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.CharLiteral:
		return "char"
	case *ast.NilLiteral:
		return "nil"
	default:
		return "dynamic"
	}
}

func integerTypeName(value *ast.IntegerType) string {
	if value == nil {
		return "i32"
	}
	return string(*value)
}

func integerAux(value *ast.IntegerType) uint32 {
	if value == nil {
		return 0
	}
	order := []ast.IntegerType{
		ast.IntegerTypeI8, ast.IntegerTypeI16, ast.IntegerTypeI32, ast.IntegerTypeI64, ast.IntegerTypeI128,
		ast.IntegerTypeU8, ast.IntegerTypeU16, ast.IntegerTypeU32, ast.IntegerTypeU64, ast.IntegerTypeU128,
	}
	for index, candidate := range order {
		if *value == candidate {
			return uint32(index + 1)
		}
	}
	return math.MaxUint32
}

func floatAux(value *ast.FloatType) uint32 {
	if value != nil && *value == ast.FloatTypeF32 {
		return 1
	}
	return 2
}

func signedMagnitude(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return []byte{0}
	}
	abs := new(big.Int).Abs(value).Bytes()
	sign := byte(0)
	if value.Sign() < 0 {
		sign = 1
	}
	return append([]byte{sign}, abs...)
}

func (lower *lowerer) emitLiteral(node ast.Expression) (uint32, bool, error) {
	var tag, aux uint32
	var data []byte
	typeName := literalTypeName(node)
	switch value := node.(type) {
	case *ast.IntegerLiteral:
		tag, aux, data = semanticabi.TagKindInteger, integerAux(value.IntegerType), signedMagnitude(value.Value)
	case *ast.FloatLiteral:
		tag, aux = semanticabi.TagKindFloat, floatAux(value.FloatType)
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, math.Float64bits(value.Value))
	case *ast.StringLiteral:
		tag, data = semanticabi.TagKindString, []byte(value.Value)
	case *ast.BooleanLiteral:
		tag, data = semanticabi.TagKindBool, []byte{0}
		if value.Value {
			data[0] = 1
		}
	case *ast.CharLiteral:
		tag, data = semanticabi.TagKindChar, []byte(value.Value)
	case *ast.NilLiteral:
		tag = semanticabi.TagKindNil
	default:
		return 0, false, nil
	}
	typeID := lower.tables.typeIndex(typeName)
	register := lower.newRegister(typeID)
	constant := lower.tables.constant(tag, aux, data)
	if err := lower.emit(node, semanticabi.OpLoadConst, register, constant); err != nil {
		return 0, true, err
	}
	return register, true, nil
}

func (lower *lowerer) emitVoid(node ast.Node) (uint32, error) {
	typeID := lower.tables.typeIndex("void")
	register := lower.newRegister(typeID)
	constant := lower.tables.constant(semanticabi.TagKindVoid, 0, nil)
	return register, lower.emit(node, semanticabi.OpLoadConst, register, constant)
}

func (lower *lowerer) emitNil(node ast.Node) (uint32, error) {
	typeID := lower.tables.typeIndex("nil")
	register := lower.newRegister(typeID)
	constant := lower.tables.constant(semanticabi.TagKindNil, 0, nil)
	return register, lower.emit(node, semanticabi.OpLoadConst, register, constant)
}

func (lower *lowerer) dynamicType() uint32 { return lower.tables.typeIndex("dynamic") }
func (lower *lowerer) boolType() uint32    { return lower.tables.typeIndex("bool") }

func (lower *lowerer) joinTypes(left, right uint32) uint32 {
	if left == right {
		return left
	}
	return lower.dynamicType()
}

func expressionName(expression ast.Expression) string {
	switch value := expression.(type) {
	case *ast.Identifier:
		return value.Name
	case *ast.MemberAccessExpression:
		left, right := expressionName(value.Object), expressionName(value.Member)
		if left != "" && right != "" {
			return left + "." + right
		}
	}
	return ""
}

func (lower *lowerer) hostCapability(callee ast.Expression) (string, string, bool) {
	name := expressionName(callee)
	switch name {
	case "print":
		return "able.host.print", "void", true
	}
	qualified := name
	if qualifier, member, ok := splitQualified(name); ok {
		if packageName, imported := lower.imports[qualifier]; imported {
			qualified = packageName + "." + member
		}
	}
	returnType, ok := lower.hostFunctions[qualified]
	return qualified, returnType, ok
}

func appendUnique(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}

func requireIdentifier(expression ast.Expression) (*ast.Identifier, error) {
	identifier, ok := expression.(*ast.Identifier)
	if !ok || identifier == nil {
		return nil, fmt.Errorf("semanticabi flow: expected identifier, got %T", expression)
	}
	return identifier, nil
}
