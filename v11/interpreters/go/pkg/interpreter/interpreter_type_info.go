package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) getTypeInfoForValue(value runtime.Value) (typeInfo, bool) {
	switch v := value.(type) {
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return typeInfo{}, false
		}
		return typeInfo{name: v.Node.ID.Name}, true
	case runtime.StructDefinitionValue:
		return i.getTypeInfoForValue(&v)
	case *runtime.StructInstanceValue:
		return i.typeInfoFromStructInstance(v)
	case *runtime.InterfaceValue:
		return i.getTypeInfoForValue(v.Underlying)
	case runtime.StringValue, runtime.BoolValue, runtime.CharValue, runtime.NilValue, runtime.VoidValue,
		runtime.IntegerValue, *runtime.IntegerValue,
		runtime.FloatValue, *runtime.FloatValue,
		*runtime.ArrayValue,
		*runtime.IteratorValue,
		runtime.IteratorEndValue,
		*runtime.IteratorEndValue:
		typeExpr := i.typeExpressionFromValue(value)
		if info, ok := parseTypeExpression(typeExpr); ok {
			return info, true
		}
		return typeInfo{}, false
	case runtime.ErrorValue:
		if v.TypeName != nil && v.TypeName.Name != "" {
			return typeInfo{name: v.TypeName.Name}, true
		}
		return typeInfo{name: "Error"}, true
	case *runtime.ErrorValue:
		if v == nil {
			return typeInfo{}, false
		}
		if v.TypeName != nil && v.TypeName.Name != "" {
			return typeInfo{name: v.TypeName.Name}, true
		}
		return typeInfo{name: "Error"}, true
	default:
		return typeInfo{}, false
	}
}

func (i *Interpreter) typeExpressionFromValue(value runtime.Value) ast.TypeExpression {
	switch v := value.(type) {
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return nil
		}
		return ast.Ty(v.Node.ID.Name)
	case runtime.StructDefinitionValue:
		return i.typeExpressionFromValue(&v)
	case runtime.StringValue:
		return ast.Ty("String")
	case runtime.BoolValue:
		return ast.Ty("bool")
	case runtime.CharValue:
		return ast.Ty("char")
	case runtime.NilValue:
		return ast.Ty("nil")
	case runtime.VoidValue:
		return ast.Ty("void")
	case runtime.IteratorEndValue:
		return ast.Ty("IteratorEnd")
	case *runtime.IteratorEndValue:
		if v == nil {
			return nil
		}
		return ast.Ty("IteratorEnd")
	case runtime.IntegerValue:
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.IntegerValue:
		if v == nil {
			return nil
		}
		return ast.Ty(string(v.TypeSuffix))
	case runtime.FloatValue:
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.FloatValue:
		if v == nil {
			return nil
		}
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.HostHandleValue:
		if v == nil {
			return nil
		}
		if v.HandleType == "" {
			return nil
		}
		return ast.Ty(v.HandleType)
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil
		}
		base := ast.Ty(v.Definition.Node.ID.Name)
		generics := v.Definition.Node.GenericParams
		if len(generics) > 0 {
			typeArgs := v.TypeArguments
			needsInference := len(typeArgs) != len(generics)
			if !needsInference {
				genericNames := genericNameSet(generics)
				for _, arg := range typeArgs {
					if arg == nil {
						needsInference = true
						break
					}
					if _, ok := arg.(*ast.WildcardTypeExpression); ok {
						needsInference = true
						break
					}
					if simple, ok := arg.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
						if _, ok := genericNames[simple.Name.Name]; ok {
							needsInference = true
							break
						}
					}
				}
			}
			if needsInference {
				typeArgs = i.inferStructTypeArguments(v.Definition.Node, v.Fields, v.Positional)
			}
			if len(typeArgs) > 0 {
				return ast.Gen(base, typeArgs...)
			}
		}
		return base
	case *runtime.InterfaceValue:
		if v == nil || v.Interface == nil || v.Interface.Node == nil || v.Interface.Node.ID == nil {
			return nil
		}
		return ast.Ty(v.Interface.Node.ID.Name)
	case runtime.InterfaceValue:
		return i.typeExpressionFromValue(&v)
	case *runtime.ArrayValue:
		if v == nil {
			return nil
		}
		var elemType ast.TypeExpression
		for _, el := range v.Elements {
			inferred := i.typeExpressionFromValue(el)
			if inferred == nil {
				continue
			}
			if elemType == nil {
				elemType = inferred
				continue
			}
			if !typeExpressionsEqual(elemType, inferred) {
				elemType = ast.NewWildcardTypeExpression()
				break
			}
		}
		if elemType == nil {
			elemType = ast.NewWildcardTypeExpression()
		}
		return ast.Gen(ast.Ty("Array"), elemType)
	case *runtime.IteratorValue:
		if v == nil {
			return nil
		}
		return ast.Ty("Iterator")
	case runtime.ErrorValue:
		if v.TypeName != nil {
			return ast.Ty(v.TypeName.Name)
		}
		return ast.Ty("Error")
	default:
		return nil
	}
}

func (i *Interpreter) canonicalTypeNames(name string) []string {
	seen := make(map[string]struct{})
	add := func(n string) {
		if n == "" {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
	}
	add(name)
	if name != "" {
		if expanded := expandTypeAliases(ast.Ty(name), i.typeAliases, nil); expanded != nil {
			if info, ok := parseTypeExpression(expanded); ok && info.name != "" {
				add(info.name)
			}
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	return names
}
