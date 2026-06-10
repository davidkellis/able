package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bindingIntent struct {
	declarationNames map[string]struct{}
	allowFallback    bool
}

type patternBinding = runtime.EnvironmentBinding

const inlinePatternBindingStorage = 8

type patternMismatchError struct {
	message string
}

func (e patternMismatchError) Error() string {
	return e.message
}

func asPatternMismatch(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	switch v := err.(type) {
	case patternMismatchError:
		return v.message, true
	case *patternMismatchError:
		if v == nil {
			return "", false
		}
		return v.message, true
	default:
		return "", false
	}
}

func zeroFieldStructPatternTypeName(pattern *ast.StructPattern) string {
	if pattern == nil || pattern.StructType == nil || pattern.StructType.Name == "" || len(pattern.Fields) != 0 {
		return ""
	}
	return normalizeKernelAliasName(pattern.StructType.Name)
}

func zeroFieldStructPatternMatchesValue(pattern *ast.StructPattern, value runtime.Value) bool {
	name := zeroFieldStructPatternTypeName(pattern)
	if name == "" {
		return false
	}
	switch v := value.(type) {
	case runtime.IteratorEndValue, *runtime.IteratorEndValue:
		return name == "IteratorEnd"
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return false
		}
		return normalizeKernelAliasName(v.Definition.Node.ID.Name) == name
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return false
		}
		return normalizeKernelAliasName(v.Node.ID.Name) == name
	case runtime.StructDefinitionValue:
		return zeroFieldStructPatternMatchesValue(pattern, &v)
	default:
		return false
	}
}

func (i *Interpreter) collectPatternBindings(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, bindings []patternBinding) ([]patternBinding, error) {
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil || p.Name == "" || p.Name == "_" {
			return bindings, nil
		}
		bindings = append(bindings, patternBinding{Name: p.Name, Value: value})
		return bindings, nil
	case *ast.WildcardPattern:
		return bindings, nil
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return bindings, fmt.Errorf("invalid literal in pattern: %T", p.Literal)
		}
		litVal, err := i.evaluateExpression(litExpr, env)
		if err != nil {
			return bindings, err
		}
		if !valuesEqual(litVal, value) {
			return bindings, patternMismatchError{message: "pattern literal mismatch"}
		}
		return bindings, nil
	case *ast.StructPattern:
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(*errValPtr)
		}
		if zeroFieldStructPatternMatchesValue(p, value) {
			return bindings, nil
		}
		switch value.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return bindings, patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return bindings, patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return bindings, patternMismatchError{message: "Struct type mismatch in destructuring"}
			}
		}
		if structVal.Positional != nil && !structUsesNamedFieldStorage(structVal) {
			if len(p.Fields) != len(structVal.Positional) {
				return bindings, patternMismatchError{message: "Struct field count mismatch"}
			}
			for idx, field := range p.Fields {
				if field == nil {
					return bindings, fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return bindings, patternMismatchError{message: fmt.Sprintf("missing positional struct value at index %d", idx)}
				}
				var err error
				bindings, err = i.collectPatternBindings(field.Pattern, fieldVal, env, bindings)
				if err != nil {
					return bindings, err
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					bindings = append(bindings, patternBinding{Name: field.Binding.Name, Value: fieldVal})
				}
			}
			return bindings, nil
		}
		if !structUsesNamedFieldStorage(structVal) {
			if len(p.Fields) == 0 {
				return bindings, nil
			}
			return bindings, patternMismatchError{message: "Expected named struct"}
		}
		if structVal.Positional != nil && structVal.Definition != nil && structVal.Definition.Node != nil {
			if plan, ok := i.namedStructPatternPlanCached(p, structVal.Definition.Node); ok && len(plan.fieldOrder) == len(p.Fields) {
				for idx, field := range p.Fields {
					fieldIndex := plan.fieldOrder[idx]
					if fieldIndex < 0 || fieldIndex >= len(structVal.Positional) {
						ok = false
						break
					}
					fieldVal := structVal.Positional[fieldIndex]
					var err error
					bindings, err = i.collectPatternBindings(field.Pattern, fieldVal, env, bindings)
					if err != nil {
						return bindings, err
					}
					if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
						bindings = append(bindings, patternBinding{Name: field.Binding.Name, Value: fieldVal})
					}
				}
				if ok {
					return bindings, nil
				}
			}
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return bindings, fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structNamedFieldValue(structVal, field.FieldName.Name)
			if !ok {
				return bindings, patternMismatchError{message: fmt.Sprintf("Missing field '%s' during destructuring", field.FieldName.Name)}
			}
			var err error
			bindings, err = i.collectPatternBindings(field.Pattern, fieldVal, env, bindings)
			if err != nil {
				return bindings, err
			}
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				bindings = append(bindings, patternBinding{Name: field.Binding.Name, Value: fieldVal})
			}
		}
		return bindings, nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return bindings, patternMismatchError{message: "Cannot destructure non-array value"}
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return bindings, patternMismatchError{message: "Array length mismatch in destructuring"}
		}
		if len(elements) < len(p.Elements) {
			return bindings, patternMismatchError{message: "Array too short for destructuring"}
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return bindings, fmt.Errorf("invalid array pattern at index %d", idx)
			}
			elemVal := elements[idx]
			var err error
			bindings, err = i.collectPatternBindings(elemPattern, elemVal, env, bindings)
			if err != nil {
				return bindings, err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				if rest.Name != "" && rest.Name != "_" {
					restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
					restVal := i.newArrayValue(restElems, len(restElems))
					bindings = append(bindings, patternBinding{Name: rest.Name, Value: restVal})
				}
			case *ast.WildcardPattern:
				// ignore remaining elements
			default:
				return bindings, fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return bindings, patternMismatchError{message: "array length mismatch in destructuring"}
		}
		return bindings, nil
	case *ast.TypedPattern:
		typeExpr := i.canonicalizeTypeExpressionCached(p.TypeAnnotation, env, i.typeExpressionReferencesAliasCached(p.TypeAnnotation))
		if !i.matchesType(typeExpr, value) {
			expected := typeExpressionToString(typeExpr)
			actualExpr := i.typeExpressionFromValue(value)
			actual := value.Kind().String()
			if actualExpr != nil {
				actual = typeExpressionToString(actualExpr)
			}
			return bindings, patternMismatchError{message: fmt.Sprintf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual)}
		}
		coerced, err := i.coerceValueToType(typeExpr, value)
		if err != nil {
			return bindings, err
		}
		return i.collectPatternBindings(p.Pattern, coerced, env, bindings)
	default:
		return bindings, fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func matchPatternClauseBindingEnv(base *runtime.Environment, matchEnv *runtime.Environment, name string, value runtime.Value, valueCapacity int) *runtime.Environment {
	if name == "" || name == "_" {
		return matchEnv
	}
	if matchEnv == nil {
		if valueCapacity < inlinePatternBindingStorage {
			valueCapacity = inlinePatternBindingStorage
		}
		return runtime.NewEnvironmentWithSingleBinding(base, valueCapacity, name, value)
	}
	matchEnv.DefineWithoutMerge(name, value)
	return matchEnv
}

func (i *Interpreter) matchPatternIntoClauseEnv(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, matchEnv *runtime.Environment, captureBindings bool, scopeCapacity int) (*runtime.Environment, error) {
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil {
			return matchEnv, nil
		}
		if !captureBindings {
			return matchEnv, nil
		}
		return matchPatternClauseBindingEnv(base, matchEnv, p.Name, value, scopeCapacity), nil
	case *ast.WildcardPattern:
		return matchEnv, nil
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid literal in pattern: %T", p.Literal)
		}
		litVal, err := i.evaluateExpression(litExpr, base)
		if err != nil {
			return nil, err
		}
		if !valuesEqual(litVal, value) {
			return nil, patternMismatchError{message: "pattern literal mismatch"}
		}
		return matchEnv, nil
	case *ast.StructPattern:
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(*errValPtr)
		}
		if zeroFieldStructPatternMatchesValue(p, value) {
			return matchEnv, nil
		}
		switch value.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return nil, patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return nil, patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return nil, patternMismatchError{message: "Struct type mismatch in destructuring"}
			}
		}
		if structVal.Positional != nil && !structUsesNamedFieldStorage(structVal) {
			if len(p.Fields) != len(structVal.Positional) {
				return nil, patternMismatchError{message: "Struct field count mismatch"}
			}
			for idx, field := range p.Fields {
				if field == nil {
					return nil, fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return nil, patternMismatchError{message: fmt.Sprintf("missing positional struct value at index %d", idx)}
				}
				var err error
				matchEnv, err = i.matchPatternIntoClauseEnv(field.Pattern, fieldVal, base, matchEnv, captureBindings, scopeCapacity)
				if err != nil {
					return nil, err
				}
				if captureBindings && field.Binding != nil {
					matchEnv = matchPatternClauseBindingEnv(base, matchEnv, field.Binding.Name, fieldVal, scopeCapacity)
				}
			}
			return matchEnv, nil
		}
		if !structUsesNamedFieldStorage(structVal) {
			if len(p.Fields) == 0 {
				return matchEnv, nil
			}
			return nil, patternMismatchError{message: "Expected named struct"}
		}
		if structVal.Positional != nil && structVal.Definition != nil && structVal.Definition.Node != nil {
			if plan, ok := i.namedStructPatternPlanCached(p, structVal.Definition.Node); ok && len(plan.fieldOrder) == len(p.Fields) {
				for idx, field := range p.Fields {
					fieldIndex := plan.fieldOrder[idx]
					if fieldIndex < 0 || fieldIndex >= len(structVal.Positional) {
						ok = false
						break
					}
					fieldVal := structVal.Positional[fieldIndex]
					var err error
					matchEnv, err = i.matchPatternIntoClauseEnv(field.Pattern, fieldVal, base, matchEnv, captureBindings, scopeCapacity)
					if err != nil {
						return nil, err
					}
					if captureBindings && field.Binding != nil {
						matchEnv = matchPatternClauseBindingEnv(base, matchEnv, field.Binding.Name, fieldVal, scopeCapacity)
					}
				}
				if ok {
					return matchEnv, nil
				}
			}
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return nil, fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structNamedFieldValue(structVal, field.FieldName.Name)
			if !ok {
				return nil, patternMismatchError{message: fmt.Sprintf("Missing field '%s' during destructuring", field.FieldName.Name)}
			}
			var err error
			matchEnv, err = i.matchPatternIntoClauseEnv(field.Pattern, fieldVal, base, matchEnv, captureBindings, scopeCapacity)
			if err != nil {
				return nil, err
			}
			if captureBindings && field.Binding != nil {
				matchEnv = matchPatternClauseBindingEnv(base, matchEnv, field.Binding.Name, fieldVal, scopeCapacity)
			}
		}
		return matchEnv, nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return nil, patternMismatchError{message: "Cannot destructure non-array value"}
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return nil, patternMismatchError{message: "Array length mismatch in destructuring"}
		}
		if len(elements) < len(p.Elements) {
			return nil, patternMismatchError{message: "Array too short for destructuring"}
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return nil, fmt.Errorf("invalid array pattern at index %d", idx)
			}
			var err error
			matchEnv, err = i.matchPatternIntoClauseEnv(elemPattern, elements[idx], base, matchEnv, captureBindings, scopeCapacity)
			if err != nil {
				return nil, err
			}
		}
		if p.RestPattern != nil && captureBindings {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				if rest.Name != "" && rest.Name != "_" {
					restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
					restVal := i.newArrayValue(restElems, len(restElems))
					matchEnv = matchPatternClauseBindingEnv(base, matchEnv, rest.Name, restVal, scopeCapacity)
				}
			case *ast.WildcardPattern:
			default:
				return nil, fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return nil, patternMismatchError{message: "array length mismatch in destructuring"}
		}
		return matchEnv, nil
	case *ast.TypedPattern:
		typeExpr := i.canonicalizeTypeExpressionCached(p.TypeAnnotation, base, i.typeExpressionReferencesAliasCached(p.TypeAnnotation))
		if !i.matchesType(typeExpr, value) {
			expected := typeExpressionToString(typeExpr)
			actualExpr := i.typeExpressionFromValue(value)
			actual := value.Kind().String()
			if actualExpr != nil {
				actual = typeExpressionToString(actualExpr)
			}
			return nil, patternMismatchError{message: fmt.Sprintf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual)}
		}
		coerced, err := i.coerceValueToType(typeExpr, value)
		if err != nil {
			return nil, err
		}
		return i.matchPatternIntoClauseEnv(p.Pattern, coerced, base, matchEnv, captureBindings, scopeCapacity)
	default:
		return nil, fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func (i *Interpreter) assignPatternExpression(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, op ast.AssignmentOperator) (runtime.Value, error) {
	if pattern == nil {
		return nil, fmt.Errorf("missing assignment pattern")
	}
	if env == nil {
		return nil, fmt.Errorf("missing assignment environment")
	}
	switch op {
	case ast.AssignmentDeclare, ast.AssignmentAssign:
	default:
		return nil, fmt.Errorf("unsupported assignment operator %s", op)
	}
	var intent *bindingIntent
	isDeclaration := op == ast.AssignmentDeclare
	if isDeclaration {
		newNames, hasAny := analyzePatternDeclarationNames(env, pattern)
		if !hasAny || len(newNames) == 0 {
			return nil, fmt.Errorf(":= requires at least one new binding")
		}
		intent = &bindingIntent{declarationNames: newNames}
	} else {
		intent = &bindingIntent{allowFallback: true}
	}
	var inline [inlinePatternBindingStorage]patternBinding
	bindings := inline[:0]
	var err error
	bindings, err = i.collectPatternBindings(pattern, value, env, bindings)
	if err != nil {
		if msg, ok := asPatternMismatch(err); ok {
			return runtime.ErrorValue{Message: msg}, nil
		}
		return nil, err
	}
	for _, binding := range bindings {
		if err := declareOrAssign(env, binding.Name, binding.Value, isDeclaration, intent); err != nil {
			return nil, err
		}
	}
	if value == nil {
		value = runtime.NilValue{}
	}
	return value, nil
}

func (i *Interpreter) assignPatternForLoop(pattern ast.Pattern, value runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	if pattern == nil {
		return nil, fmt.Errorf("missing assignment pattern")
	}
	if env == nil {
		return nil, fmt.Errorf("missing assignment environment")
	}
	var inline [inlinePatternBindingStorage]patternBinding
	bindings := inline[:0]
	var err error
	bindings, err = i.collectPatternBindings(pattern, value, env, bindings)
	if err != nil {
		if msg, ok := asPatternMismatch(err); ok {
			return runtime.ErrorValue{Message: msg}, nil
		}
		return nil, err
	}
	if len(bindings) > 0 {
		env.DefineWithoutMergeBindings(bindings)
	}
	if value == nil {
		value = runtime.NilValue{}
	}
	return value, nil
}

func (i *Interpreter) assignPattern(
	pattern ast.Pattern,
	value runtime.Value,
	env *runtime.Environment,
	isDeclaration bool,
	intent *bindingIntent,
) error {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil {
			return nil
		}
		return declareOrAssign(env, p.Name, value, isDeclaration, intent)
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
		switch value.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			if p.StructType != nil && p.StructType.Name == "IteratorEnd" && len(p.Fields) == 0 {
				return nil
			}
			return fmt.Errorf("Cannot destructure non-struct value")
		}
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(*errValPtr)
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
		if structVal.Positional != nil && !structUsesNamedFieldStorage(structVal) {
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
				if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration, intent); err != nil {
					return err
				}
				if field.Binding != nil {
					if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration, intent); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if !structUsesNamedFieldStorage(structVal) {
			if len(p.Fields) == 0 {
				return nil
			}
			return fmt.Errorf("Expected named struct")
		}
		if structVal.Positional != nil && structVal.Definition != nil && structVal.Definition.Node != nil {
			if plan, ok := i.namedStructPatternPlanCached(p, structVal.Definition.Node); ok && len(plan.fieldOrder) == len(p.Fields) {
				for idx, field := range p.Fields {
					fieldIndex := plan.fieldOrder[idx]
					if fieldIndex < 0 || fieldIndex >= len(structVal.Positional) {
						ok = false
						break
					}
					fieldVal := structVal.Positional[fieldIndex]
					if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration, intent); err != nil {
						return err
					}
					if field.Binding != nil {
						if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration, intent); err != nil {
							return err
						}
					}
				}
				if ok {
					return nil
				}
			}
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structNamedFieldValue(structVal, field.FieldName.Name)
			if !ok {
				return fmt.Errorf("Missing field '%s' during destructuring", field.FieldName.Name)
			}
			if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration, intent); err != nil {
				return err
			}
			if field.Binding != nil {
				if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration, intent); err != nil {
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
			if err := i.assignPattern(elemPattern, elemVal, env, isDeclaration, intent); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
				restVal := i.newArrayValue(restElems, len(restElems))
				if err := declareOrAssign(env, rest.Name, restVal, isDeclaration, intent); err != nil {
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
		typeExpr := i.canonicalizeTypeExpressionCached(p.TypeAnnotation, env, i.typeExpressionReferencesAliasCached(p.TypeAnnotation))
		if !i.matchesType(typeExpr, value) {
			expected := typeExpressionToString(typeExpr)
			actualExpr := i.typeExpressionFromValue(value)
			actual := value.Kind().String()
			if actualExpr != nil {
				actual = typeExpressionToString(actualExpr)
			}
			return fmt.Errorf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual)
		}
		coerced, err := i.coerceValueToType(typeExpr, value)
		if err != nil {
			return err
		}
		return i.assignPattern(p.Pattern, coerced, env, isDeclaration, intent)
	default:
		return fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func (i *Interpreter) errorValueToStructInstance(err runtime.ErrorValue) *runtime.StructInstanceValue {
	fields := make(map[string]runtime.Value)
	if err.Payload != nil {
		for k, v := range err.Payload {
			fields[k] = v
		}
	}
	fields["message"] = runtime.StringValue{Val: err.Message}
	inst := &runtime.StructInstanceValue{Fields: fields}
	if i != nil && err.TypeName != nil && err.TypeName.Name != "" {
		if def, ok := i.lookupStructDefinition(err.TypeName.Name); ok && def != nil {
			inst.Definition = def
		}
	}
	return inst
}

func bindinglessPatternEnv(base *runtime.Environment, reuseBase bool, scopeCapacity int) *runtime.Environment {
	if reuseBase {
		return base
	}
	if scopeCapacity > 0 {
		return runtime.NewEnvironmentWithValueCapacity(base, scopeCapacity)
	}
	return runtime.NewEnvironment(base)
}

func bindinglessClauseScopeEnv(base *runtime.Environment, plan clauseScopePlan) *runtime.Environment {
	if !plan.needsLocalScope {
		return base
	}
	return runtime.NewEnvironmentWithValueCapacity(base, plan.localBindingCapacity)
}

func (i *Interpreter) matchPatternFastWithScopeReuse(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, captureBindings bool, reuseBaseForBindingless bool, scopeCapacity int) (*runtime.Environment, bool, bool) {
	if pattern == nil || base == nil {
		return nil, false, false
	}
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	if ident, ok := pattern.(*ast.Identifier); ok && ident != nil {
		if ident.Name == "" || ident.Name == "_" {
			return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
		}
		if existing, ok := base.Lookup(ident.Name); ok {
			switch defVal := existing.(type) {
			case *runtime.StructDefinitionValue:
				if isSingletonStructDef(defVal.Node) {
					if valuesEqual(existing, value) {
						return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
					}
					return nil, false, true
				}
			case runtime.StructDefinitionValue:
				if isSingletonStructDef(defVal.Node) {
					if valuesEqual(existing, value) {
						return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
					}
					return nil, false, true
				}
			}
		}
		if !captureBindings {
			return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
		}
		if scopeCapacity < 1 {
			scopeCapacity = 1
		}
		matchEnv := runtime.NewEnvironmentWithSingleBinding(base, scopeCapacity, ident.Name, value)
		return matchEnv, true, true
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return nil, false, true
		}
		litVal, err := i.evaluateExpression(litExpr, base)
		if err != nil || !valuesEqual(litVal, value) {
			return nil, false, true
		}
		return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
	case *ast.StructPattern:
		if zeroFieldStructPatternMatchesValue(p, value) {
			return bindinglessPatternEnv(base, reuseBaseForBindingless, scopeCapacity), true, true
		}
		return nil, false, false
	case *ast.TypedPattern:
		coerced, ok := i.matchTypedPatternValueInEnv(p.TypeAnnotation, value, base)
		if !ok {
			return nil, false, true
		}
		return i.matchPatternFastWithScopeReuse(p.Pattern, coerced, base, captureBindings, reuseBaseForBindingless, scopeCapacity)
	default:
		return nil, false, false
	}
}

func (i *Interpreter) matchPatternFast(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool, bool) {
	return i.matchPatternFastWithScopeReuse(pattern, value, base, true, false, patternBindingCapacity(pattern))
}

func (i *Interpreter) matchTypedPatternValue(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool) {
	return i.matchTypedPatternValueInEnv(typeExpr, value, nil)
}

func (i *Interpreter) matchTypedPatternValueInEnv(typeExpr ast.TypeExpression, value runtime.Value, env *runtime.Environment) (runtime.Value, bool) {
	typeExpr = i.canonicalizeTypeExpressionCached(typeExpr, env, i.typeExpressionReferencesAliasCached(typeExpr))
	if coerced, ok := matchTypedPatternExactPrimitive(typeExpr, value); ok {
		return coerced, true
	}
	if coerced, ok := matchTypedPatternExactNamedStruct(typeExpr, value); ok {
		return coerced, true
	}
	if coerced, ok, err := i.tryFastSimpleTypeCoercion(typeExpr, value); ok {
		if err != nil {
			return nil, false
		}
		return coerced, true
	}
	if matched, ok := i.matchesTypeWithoutRuntimeValue(typeExpr); ok {
		if !matched {
			return nil, false
		}
		return bytecodeStackSnapshotValue(value), true
	}
	if !i.matchesType(typeExpr, value) {
		return nil, false
	}
	coerced, err := i.coerceValueToType(typeExpr, value)
	if err != nil {
		return nil, false
	}
	return coerced, true
}

func matchTypedPatternExactPrimitive(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool) {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return nil, false
	}
	name := normalizeKernelAliasName(simple.Name.Name)
	switch v := value.(type) {
	case runtime.IntegerValue:
		if string(v.TypeSuffix) == name {
			return value, true
		}
	case *runtime.IntegerValue:
		if v != nil && string(v.TypeSuffix) == name {
			return value, true
		}
	case runtime.FloatValue:
		if string(v.TypeSuffix) == name {
			return value, true
		}
	case *runtime.FloatValue:
		if v != nil && string(v.TypeSuffix) == name {
			return value, true
		}
	case runtime.StringValue:
		if name == "String" {
			return value, true
		}
	case *runtime.StringValue:
		if v != nil && name == "String" {
			return value, true
		}
	case runtime.BoolValue:
		if name == "bool" || name == "Bool" {
			return value, true
		}
	case *runtime.BoolValue:
		if v != nil && (name == "bool" || name == "Bool") {
			return value, true
		}
	case runtime.CharValue:
		if name == "char" {
			return value, true
		}
	case *runtime.CharValue:
		if v != nil && name == "char" {
			return value, true
		}
	case runtime.NilValue:
		if name == "nil" {
			return value, true
		}
	case runtime.VoidValue:
		if name == "void" {
			return value, true
		}
	case *runtime.VoidValue:
		if v != nil && name == "void" {
			return value, true
		}
	case runtime.IteratorEndValue:
		if name == "IteratorEnd" {
			return value, true
		}
	case *runtime.IteratorEndValue:
		if v != nil && name == "IteratorEnd" {
			return value, true
		}
	}
	return nil, false
}

func matchTypedPatternExactNamedStruct(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool) {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return nil, false
	}
	name := simple.Name.Name
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil, false
		}
		if len(v.Definition.Node.GenericParams) != 0 {
			return nil, false
		}
		if v.Definition.Node.ID.Name == name {
			return value, true
		}
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return nil, false
		}
		if v.Node.ID.Name == name {
			return value, true
		}
	case runtime.StructDefinitionValue:
		return matchTypedPatternExactNamedStruct(typeExpr, &v)
	}
	return nil, false
}

func (i *Interpreter) matchPattern(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool) {
	return i.matchPatternForClause(pattern, value, base, clauseScopePlan{
		needsLocalScope:       true,
		capturePatternBinding: true,
		localBindingCapacity:  patternBindingCapacity(pattern),
	})
}

func (i *Interpreter) matchPatternForClause(pattern ast.Pattern, value runtime.Value, base *runtime.Environment, plan clauseScopePlan) (*runtime.Environment, bool) {
	if pattern == nil {
		return nil, false
	}
	if matchEnv, matched, handled := i.matchPatternFastWithScopeReuse(pattern, value, base, plan.capturePatternBinding, true, plan.localBindingCapacity); handled {
		if !matched {
			return nil, false
		}
		if matchEnv == base {
			return bindinglessClauseScopeEnv(base, plan), true
		}
		return matchEnv, true
	}
	matchEnv, err := i.matchPatternIntoClauseEnv(pattern, value, base, nil, plan.capturePatternBinding, plan.localBindingCapacity)
	if err != nil {
		return nil, false
	}
	if matchEnv == nil {
		return bindinglessClauseScopeEnv(base, plan), true
	}
	return matchEnv, true
}

func declareOrAssign(env *runtime.Environment, name string, value runtime.Value, isDeclaration bool, intent *bindingIntent) error {
	if isDeclaration {
		if intent == nil || intent.declarationNames == nil {
			env.Define(name, value)
			return nil
		}
		if _, ok := intent.declarationNames[name]; ok {
			env.Define(name, value)
			return nil
		}
		return env.Assign(name, value)
	}
	if err := env.Assign(name, value); err != nil {
		if intent != nil && intent.allowFallback {
			env.Define(name, value)
			return nil
		}
		return err
	}
	return nil
}
