package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) assignPattern(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, isDeclaration bool) error {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return declareOrAssign(env, p.Name, value, isDeclaration)
	case *ast.WildcardPattern:
		return nil
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return fmt.Errorf("invalid literal in pattern: %T", p.Literal)
		}
		litVal, err := i.evaluateExpression(litExpr, env)
		if err != nil {
			return err
		}
		if !valuesEqual(litVal, value) {
			return fmt.Errorf("pattern literal mismatch")
		}
		return nil
	case *ast.StructPattern:
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = errorValueToStructInstance(*errValPtr)
		}
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return fmt.Errorf("Cannot destructure non-struct value")
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return fmt.Errorf("Struct type mismatch in destructuring")
			}
		}
		if p.IsPositional {
			if structVal.Positional == nil {
				return fmt.Errorf("Expected positional struct")
			}
			if len(p.Fields) != len(structVal.Positional) {
				return fmt.Errorf("Struct field count mismatch")
			}
			for idx, field := range p.Fields {
				if field == nil {
					return fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return fmt.Errorf("missing positional struct value at index %d", idx)
				}
				if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
					return err
				}
				if field.Binding != nil {
					if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if structVal.Fields == nil {
			return fmt.Errorf("Expected named struct")
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structVal.Fields[field.FieldName.Name]
			if !ok {
				return fmt.Errorf("Missing field '%s' during destructuring", field.FieldName.Name)
			}
			if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
				return err
			}
			if field.Binding != nil {
				if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
					return err
				}
			}
		}
		return nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return fmt.Errorf("Cannot destructure non-array value")
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return fmt.Errorf("Array length mismatch in destructuring")
		}
		if len(elements) < len(p.Elements) {
			return fmt.Errorf("Array too short for destructuring")
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return fmt.Errorf("invalid array pattern at index %d", idx)
			}
			elemVal := elements[idx]
			if err := i.assignPattern(elemPattern, elemVal, env, isDeclaration); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
				restVal := &runtime.ArrayValue{Elements: restElems}
				if err := declareOrAssign(env, rest.Name, restVal, isDeclaration); err != nil {
					return err
				}
			case *ast.WildcardPattern:
				// ignore remaining elements
			default:
				return fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return fmt.Errorf("array length mismatch in destructuring")
		}
		return nil
	case *ast.TypedPattern:
		if !i.matchesType(p.TypeAnnotation, value) {
			return fmt.Errorf("Typed pattern mismatch in assignment")
		}
		coerced, err := i.coerceValueToType(p.TypeAnnotation, value)
		if err != nil {
			return err
		}
		return i.assignPattern(p.Pattern, coerced, env, isDeclaration)
	default:
		return fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func errorValueToStructInstance(err runtime.ErrorValue) *runtime.StructInstanceValue {
	fields := make(map[string]runtime.Value)
	if err.Payload != nil {
		for k, v := range err.Payload {
			fields[k] = v
		}
	}
	fields["message"] = runtime.StringValue{Val: err.Message}
	return &runtime.StructInstanceValue{Fields: fields}
}

func (i *Interpreter) matchPattern(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool) {
	if pattern == nil {
		return nil, false
	}
	matchEnv := runtime.NewEnvironment(base)
	if err := i.assignPattern(pattern, value, matchEnv, true); err != nil {
		return nil, false
	}
	return matchEnv, true
}

func declareOrAssign(env *runtime.Environment, name string, value runtime.Value, isDeclaration bool) error {
	if isDeclaration {
		env.Define(name, value)
		return nil
	}
	return env.Assign(name, value)
}
