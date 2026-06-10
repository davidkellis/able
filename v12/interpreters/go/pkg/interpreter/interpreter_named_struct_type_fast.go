package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func fastExactNamedStructTypeMatch(i *Interpreter, typeExpr ast.TypeExpression, value runtime.Value) (bool, bool) {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return false, false
	}
	name := normalizeKernelAliasName(simple.Name.Name)
	if name == "" || fastNamedStructTypeNameIsNonNominal(i, name) {
		return false, false
	}
	if _, ok := exactNamedStructCoercionValueForName(value, name); ok {
		return true, true
	}
	return false, false
}

func fastNamedStructTypeNameIsNonNominal(i *Interpreter, name string) bool {
	switch name {
	case "Array", "Channel", "Error", "Future", "HashMap", "IoHandle", "Iterator", "IteratorEnd", "Mutex", "ProcHandle", "Self", "String", "bool", "char", "nil", "void":
		return true
	}
	if _, ok := lookupIntegerInfo(runtime.IntegerType(name)); ok {
		return true
	}
	if name == "f32" || name == "f64" {
		return true
	}
	if i != nil {
		if _, ok := i.interfaces[name]; ok {
			return true
		}
		if _, ok := i.unionDefinitions[name]; ok {
			return true
		}
	}
	return false
}

func exactNamedStructValueHasName(value runtime.Value, name string) bool {
	_, ok := exactNamedStructCoercionValueForName(value, name)
	return ok
}

func exactNamedStructValueMatchesDefinition(value runtime.Value, def *runtime.StructDefinitionValue, name string) bool {
	_, ok := exactNamedStructCoercionValueForDefinition(value, def, name)
	return ok
}

func exactNamedStructCoercionValueForName(value runtime.Value, name string) (runtime.Value, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			return value, v.Definition.Node.ID.Name == name
		}
	case *runtime.StructDefinitionValue:
		if v != nil && isSingletonStructDef(v.Node) && v.Node != nil && v.Node.ID != nil {
			return value, v.Node.ID.Name == name
		}
	case runtime.StructDefinitionValue:
		return exactNamedStructCoercionValueForName(&v, name)
	}
	if errVal, ok := asErrorValue(value); ok && errVal.Payload != nil {
		if payload, ok := errVal.Payload["value"]; ok && payload != nil {
			if coerced, ok := exactNamedStructCoercionValueForName(payload, name); ok {
				return coerced, true
			}
		}
	}
	return nil, false
}

func exactNamedStructCoercionValueForDefinition(value runtime.Value, def *runtime.StructDefinitionValue, name string) (runtime.Value, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && v.Definition != nil {
			if def != nil && (v.Definition == def || (v.Definition.Node != nil && def.Node != nil && v.Definition.Node == def.Node)) {
				return value, true
			}
			if name != "" && v.Definition.Node != nil && v.Definition.Node.ID != nil && v.Definition.Node.ID.Name == name {
				return value, true
			}
		}
	case *runtime.StructDefinitionValue:
		if v != nil && isSingletonStructDef(v.Node) {
			if def != nil && (v == def || (v.Node != nil && def.Node != nil && v.Node == def.Node)) {
				return value, true
			}
			if name != "" && v.Node != nil && v.Node.ID != nil && v.Node.ID.Name == name {
				return value, true
			}
		}
	case runtime.StructDefinitionValue:
		return exactNamedStructCoercionValueForDefinition(&v, def, name)
	}
	if errVal, ok := asErrorValue(value); ok && errVal.Payload != nil {
		if payload, ok := errVal.Payload["value"]; ok && payload != nil {
			if coerced, ok := exactNamedStructCoercionValueForDefinition(payload, def, name); ok {
				return coerced, true
			}
		}
	}
	return nil, false
}

func inlineExactNamedStructNoCoercion(i *Interpreter, typeName string, value runtime.Value) bool {
	name := normalizeKernelAliasName(typeName)
	if name == "" || fastNamedStructTypeNameIsNonNominal(i, name) {
		return false
	}
	return exactNamedStructValueHasName(value, name)
}

func inlineExactNamedStructNoCoercionCached(def *runtime.StructDefinitionValue, typeName string, value runtime.Value) bool {
	name := normalizeKernelAliasName(typeName)
	if def == nil && name == "" {
		return false
	}
	return exactNamedStructValueMatchesDefinition(value, def, name)
}

func inlineExactNamedStructNoCoercionBytecodeExactDef(def *runtime.StructDefinitionValue, value runtime.Value) bool {
	if def == nil {
		return false
	}
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil {
			return false
		}
		return v.Definition == def || (v.Definition.Node != nil && def.Node != nil && v.Definition.Node == def.Node)
	case *runtime.StructDefinitionValue:
		if v == nil || !isSingletonStructDef(v.Node) {
			return false
		}
		return v == def || (v.Node != nil && def.Node != nil && v.Node == def.Node)
	case runtime.StructDefinitionValue:
		return inlineExactNamedStructNoCoercionBytecodeExactDef(def, &v)
	default:
		return false
	}
}

func inlineCoercionUnnecessaryWithInterpreter(i *Interpreter, typeExpr ast.TypeExpression, val runtime.Value) bool {
	if inlineCoercionUnnecessary(typeExpr, val) {
		return true
	}
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	return inlineExactNamedStructNoCoercion(i, simple.Name.Name, val)
}

func inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(i *Interpreter, typeName string, val runtime.Value) bool {
	if inlineCoercionUnnecessaryBySimpleType(typeName, val) {
		return true
	}
	return inlineExactNamedStructNoCoercion(i, typeName, val)
}

func exactNamedStructDefinitionForTypeExpr(i *Interpreter, env *runtime.Environment, typeExpr ast.TypeExpression) *runtime.StructDefinitionValue {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
		return nil
	}
	name := normalizeKernelAliasName(simple.Name.Name)
	if name == "" || fastNamedStructTypeNameIsNonNominal(i, name) || env == nil {
		return nil
	}
	def, ok := env.StructDefinition(name)
	if !ok || def == nil || def.Node == nil || def.Node.Kind != ast.StructKindNamed || len(def.Node.GenericParams) != 0 {
		return nil
	}
	return def
}
