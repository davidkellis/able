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

type patternBinding struct {
	name  string
	value runtime.Value
}

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

func (i *Interpreter) collectPatternBindings(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, bindings *[]patternBinding) error {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil || p.Name == "" || p.Name == "_" {
			return nil
		}
		*bindings = append(*bindings, patternBinding{name: p.Name, value: value})
		return nil
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
			return patternMismatchError{message: "pattern literal mismatch"}
		}
		return nil
	case *ast.StructPattern:
		switch value.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			if p.StructType != nil && p.StructType.Name == "IteratorEnd" && len(p.Fields) == 0 {
				return nil
			}
			return patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = i.errorValueToStructInstance(*errValPtr)
		}
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return patternMismatchError{message: "Cannot destructure non-struct value"}
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return patternMismatchError{message: "Struct type mismatch in destructuring"}
			}
		}
		if structVal.Positional != nil {
			if len(p.Fields) != len(structVal.Positional) {
				return patternMismatchError{message: "Struct field count mismatch"}
			}
			for idx, field := range p.Fields {
				if field == nil {
					return fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return patternMismatchError{message: fmt.Sprintf("missing positional struct value at index %d", idx)}
				}
				if err := i.collectPatternBindings(field.Pattern, fieldVal, env, bindings); err != nil {
					return err
				}
				if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
					*bindings = append(*bindings, patternBinding{name: field.Binding.Name, value: fieldVal})
				}
			}
			return nil
		}
		if structVal.Fields == nil {
			if len(p.Fields) == 0 {
				return nil
			}
			return patternMismatchError{message: "Expected named struct"}
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structVal.Fields[field.FieldName.Name]
			if !ok {
				return patternMismatchError{message: fmt.Sprintf("Missing field '%s' during destructuring", field.FieldName.Name)}
			}
			if err := i.collectPatternBindings(field.Pattern, fieldVal, env, bindings); err != nil {
				return err
			}
			if field.Binding != nil && field.Binding.Name != "" && field.Binding.Name != "_" {
				*bindings = append(*bindings, patternBinding{name: field.Binding.Name, value: fieldVal})
			}
		}
		return nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return patternMismatchError{message: "Cannot destructure non-array value"}
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return patternMismatchError{message: "Array length mismatch in destructuring"}
		}
		if len(elements) < len(p.Elements) {
			return patternMismatchError{message: "Array too short for destructuring"}
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return fmt.Errorf("invalid array pattern at index %d", idx)
			}
			elemVal := elements[idx]
			if err := i.collectPatternBindings(elemPattern, elemVal, env, bindings); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				if rest.Name != "" && rest.Name != "_" {
					restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
					restVal := i.newArrayValue(restElems, len(restElems))
					*bindings = append(*bindings, patternBinding{name: rest.Name, value: restVal})
				}
			case *ast.WildcardPattern:
				// ignore remaining elements
			default:
				return fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return patternMismatchError{message: "array length mismatch in destructuring"}
		}
		return nil
	case *ast.TypedPattern:
		if !i.matchesType(p.TypeAnnotation, value) {
			expected := typeExpressionToString(p.TypeAnnotation)
			actualExpr := i.typeExpressionFromValue(value)
			actual := value.Kind().String()
			if actualExpr != nil {
				actual = typeExpressionToString(actualExpr)
			}
			return patternMismatchError{message: fmt.Sprintf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual)}
		}
		coerced, err := i.coerceValueToType(p.TypeAnnotation, value)
		if err != nil {
			return err
		}
		return i.collectPatternBindings(p.Pattern, coerced, env, bindings)
	default:
		return fmt.Errorf("unsupported pattern %s", pattern.NodeType())
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
	bindings := make([]patternBinding, 0)
	if err := i.collectPatternBindings(pattern, value, env, &bindings); err != nil {
		if msg, ok := asPatternMismatch(err); ok {
			return runtime.ErrorValue{Message: msg}, nil
		}
		return nil, err
	}
	for _, binding := range bindings {
		if err := declareOrAssign(env, binding.name, binding.value, isDeclaration, intent); err != nil {
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
	bindings := make([]patternBinding, 0)
	if err := i.collectPatternBindings(pattern, value, env, &bindings); err != nil {
		if msg, ok := asPatternMismatch(err); ok {
			return runtime.ErrorValue{Message: msg}, nil
		}
		return nil, err
	}
	for _, binding := range bindings {
		env.DefineWithoutMerge(binding.name, binding.value)
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
		if structVal.Positional != nil {
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
		if structVal.Fields == nil {
			if len(p.Fields) == 0 {
				return nil
			}
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
		if !i.matchesType(p.TypeAnnotation, value) {
			expected := typeExpressionToString(p.TypeAnnotation)
			actualExpr := i.typeExpressionFromValue(value)
			actual := value.Kind().String()
			if actualExpr != nil {
				actual = typeExpressionToString(actualExpr)
			}
			return fmt.Errorf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual)
		}
		coerced, err := i.coerceValueToType(p.TypeAnnotation, value)
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

func (i *Interpreter) matchPatternFast(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool, bool) {
	if pattern == nil || base == nil {
		return nil, false, false
	}
	if ident, ok := pattern.(*ast.Identifier); ok && ident != nil {
		if ident.Name == "" || ident.Name == "_" {
			return runtime.NewEnvironment(base), true, true
		}
		if existing, ok := base.Lookup(ident.Name); ok {
			switch defVal := existing.(type) {
			case *runtime.StructDefinitionValue:
				if isSingletonStructDef(defVal.Node) {
					if valuesEqual(existing, value) {
						return runtime.NewEnvironment(base), true, true
					}
					return nil, false, true
				}
			case runtime.StructDefinitionValue:
				if isSingletonStructDef(defVal.Node) {
					if valuesEqual(existing, value) {
						return runtime.NewEnvironment(base), true, true
					}
					return nil, false, true
				}
			}
		}
		matchEnv := runtime.NewEnvironmentWithValueCapacity(base, 1)
		matchEnv.DefineWithoutMerge(ident.Name, value)
		return matchEnv, true, true
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return runtime.NewEnvironment(base), true, true
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return nil, false, true
		}
		litVal, err := i.evaluateExpression(litExpr, base)
		if err != nil || !valuesEqual(litVal, value) {
			return nil, false, true
		}
		return runtime.NewEnvironment(base), true, true
	case *ast.TypedPattern:
		coerced, ok := i.matchTypedPatternValue(p.TypeAnnotation, value)
		if !ok {
			return nil, false, true
		}
		return i.matchPatternFast(p.Pattern, coerced, base)
	default:
		return nil, false, false
	}
}

func (i *Interpreter) matchTypedPatternValue(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool) {
	if coerced, ok := matchTypedPatternExactPrimitive(typeExpr, value); ok {
		return coerced, true
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

func (i *Interpreter) matchPattern(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool) {
	if pattern == nil {
		return nil, false
	}
	if matchEnv, matched, handled := i.matchPatternFast(pattern, value, base); handled {
		return matchEnv, matched
	}
	bindings := make([]patternBinding, 0, 4)
	if err := i.collectPatternBindings(pattern, value, base, &bindings); err != nil {
		return nil, false
	}
	matchEnv := runtime.NewEnvironmentWithValueCapacity(base, len(bindings))
	for _, binding := range bindings {
		matchEnv.DefineWithoutMerge(binding.name, binding.value)
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
