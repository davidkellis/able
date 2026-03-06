package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"sync"
)

var (
	cachedWildcardTypeExpression ast.TypeExpression = ast.NewWildcardTypeExpression()

	cachedSimpleTypeExpressionsMu sync.RWMutex
	cachedSimpleTypeExpressions   = map[string]ast.TypeExpression{
		"String":      ast.Ty("String"),
		"bool":        ast.Ty("bool"),
		"char":        ast.Ty("char"),
		"nil":         ast.Ty("nil"),
		"void":        ast.Ty("void"),
		"IteratorEnd": ast.Ty("IteratorEnd"),
		"Array":       ast.Ty("Array"),
		"Iterator":    ast.Ty("Iterator"),
		"Error":       ast.Ty("Error"),
	}

	cachedIntegerTypeExpressions = map[runtime.IntegerType]ast.TypeExpression{
		runtime.IntegerI8:    ast.Ty(string(runtime.IntegerI8)),
		runtime.IntegerI16:   ast.Ty(string(runtime.IntegerI16)),
		runtime.IntegerI32:   ast.Ty(string(runtime.IntegerI32)),
		runtime.IntegerI64:   ast.Ty(string(runtime.IntegerI64)),
		runtime.IntegerI128:  ast.Ty(string(runtime.IntegerI128)),
		runtime.IntegerU8:    ast.Ty(string(runtime.IntegerU8)),
		runtime.IntegerU16:   ast.Ty(string(runtime.IntegerU16)),
		runtime.IntegerU32:   ast.Ty(string(runtime.IntegerU32)),
		runtime.IntegerU64:   ast.Ty(string(runtime.IntegerU64)),
		runtime.IntegerU128:  ast.Ty(string(runtime.IntegerU128)),
		runtime.IntegerIsize: ast.Ty(string(runtime.IntegerIsize)),
		runtime.IntegerUsize: ast.Ty(string(runtime.IntegerUsize)),
	}

	cachedFloatTypeExpressions = map[runtime.FloatType]ast.TypeExpression{
		runtime.FloatF32: ast.Ty(string(runtime.FloatF32)),
		runtime.FloatF64: ast.Ty(string(runtime.FloatF64)),
	}

	cachedGenericTypeExpressionsMu sync.RWMutex
	cachedArrayTypeExpressions     = make(map[string]ast.TypeExpression)
	cachedIteratorTypeExpression   = ast.Gen(ast.Ty("Iterator"), cachedWildcardTypeExpression)

	cachedTypeInfosMu    sync.RWMutex
	cachedArrayTypeInfos = map[string]typeInfo{
		"*": {name: "Array", typeArgs: []ast.TypeExpression{cachedWildcardTypeExpression}},
	}
	cachedIteratorTypeInfo = typeInfo{name: "Iterator", typeArgs: []ast.TypeExpression{cachedWildcardTypeExpression}}
)

func cachedSimpleTypeExpression(name string) ast.TypeExpression {
	cachedSimpleTypeExpressionsMu.RLock()
	expr, ok := cachedSimpleTypeExpressions[name]
	cachedSimpleTypeExpressionsMu.RUnlock()
	if ok {
		return expr
	}
	created := ast.Ty(name)
	cachedSimpleTypeExpressionsMu.Lock()
	if existing, ok := cachedSimpleTypeExpressions[name]; ok {
		cachedSimpleTypeExpressionsMu.Unlock()
		return existing
	}
	cachedSimpleTypeExpressions[name] = created
	cachedSimpleTypeExpressionsMu.Unlock()
	return created
}

func cachedIntegerTypeExpression(kind runtime.IntegerType) ast.TypeExpression {
	if expr, ok := cachedIntegerTypeExpressions[kind]; ok {
		return expr
	}
	return ast.Ty(string(kind))
}

func cachedFloatTypeExpression(kind runtime.FloatType) ast.TypeExpression {
	if expr, ok := cachedFloatTypeExpressions[kind]; ok {
		return expr
	}
	return ast.Ty(string(kind))
}

func cachedArrayTypeExpression(elemType ast.TypeExpression) ast.TypeExpression {
	if elemType == nil {
		elemType = cachedWildcardTypeExpression
	}
	key := "*"
	cacheable := false
	if simple, ok := elemType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name != "" {
		key = simple.Name.Name
		cacheable = true
	} else if _, ok := elemType.(*ast.WildcardTypeExpression); ok {
		cacheable = true
	}
	if !cacheable {
		return ast.Gen(cachedSimpleTypeExpression("Array"), elemType)
	}
	cachedGenericTypeExpressionsMu.RLock()
	cached, ok := cachedArrayTypeExpressions[key]
	cachedGenericTypeExpressionsMu.RUnlock()
	if ok {
		return cached
	}
	cachedElem := elemType
	if key != "*" {
		cachedElem = cachedSimpleTypeExpression(key)
	} else {
		cachedElem = cachedWildcardTypeExpression
	}
	created := ast.Gen(cachedSimpleTypeExpression("Array"), cachedElem)
	cachedGenericTypeExpressionsMu.Lock()
	if existing, ok := cachedArrayTypeExpressions[key]; ok {
		cachedGenericTypeExpressionsMu.Unlock()
		return existing
	}
	cachedArrayTypeExpressions[key] = created
	cachedGenericTypeExpressionsMu.Unlock()
	return created
}

func cachedArrayTypeInfo(elemType ast.TypeExpression) typeInfo {
	if elemType == nil {
		return cachedArrayTypeInfos["*"]
	}
	key := ""
	if _, ok := elemType.(*ast.WildcardTypeExpression); ok {
		key = "*"
	} else if simple, ok := elemType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name != "" {
		key = simple.Name.Name
	}
	if key == "" {
		return typeInfo{name: "Array", typeArgs: []ast.TypeExpression{elemType}}
	}
	cachedTypeInfosMu.RLock()
	info, ok := cachedArrayTypeInfos[key]
	cachedTypeInfosMu.RUnlock()
	if ok {
		return info
	}
	cachedElem := cachedSimpleTypeExpression(key)
	created := typeInfo{name: "Array", typeArgs: []ast.TypeExpression{cachedElem}}
	cachedTypeInfosMu.Lock()
	if existing, ok := cachedArrayTypeInfos[key]; ok {
		cachedTypeInfosMu.Unlock()
		return existing
	}
	cachedArrayTypeInfos[key] = created
	cachedTypeInfosMu.Unlock()
	return created
}

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
	case runtime.StringValue, *runtime.StringValue:
		if ptr, ok := v.(*runtime.StringValue); ok && ptr == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: "String"}, true
	case runtime.BoolValue, *runtime.BoolValue:
		return typeInfo{name: "bool"}, true
	case runtime.CharValue, *runtime.CharValue:
		return typeInfo{name: "char"}, true
	case runtime.NilValue:
		return typeInfo{name: "nil"}, true
	case runtime.VoidValue, *runtime.VoidValue:
		return typeInfo{name: "void"}, true
	case runtime.IntegerValue:
		return typeInfo{name: string(v.TypeSuffix)}, true
	case *runtime.IntegerValue:
		if v == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: string(v.TypeSuffix)}, true
	case runtime.FloatValue:
		return typeInfo{name: string(v.TypeSuffix)}, true
	case *runtime.FloatValue:
		if v == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: string(v.TypeSuffix)}, true
	case *runtime.ArrayValue:
		if v == nil {
			return typeInfo{}, false
		}
		elemType := cachedWildcardTypeExpression
		if len(v.Elements) > 0 {
			if inferred := i.typeExpressionFromValueWithSeen(v.Elements[0], nil); inferred != nil {
				elemType = inferred
			}
		}
		return cachedArrayTypeInfo(elemType), true
	case *runtime.IteratorValue:
		if v == nil {
			return typeInfo{}, false
		}
		return cachedIteratorTypeInfo, true
	case runtime.IteratorEndValue:
		return typeInfo{name: "IteratorEnd"}, true
	case *runtime.IteratorEndValue:
		if v == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: "IteratorEnd"}, true
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
	return i.typeExpressionFromValueWithSeen(value, nil)
}

func (i *Interpreter) typeExpressionFromValueWithSeen(value runtime.Value, seen map[*runtime.StructInstanceValue]struct{}) ast.TypeExpression {
	switch v := value.(type) {
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return nil
		}
		return cachedSimpleTypeExpression(v.Node.ID.Name)
	case runtime.StructDefinitionValue:
		return i.typeExpressionFromValueWithSeen(&v, seen)
	case runtime.StringValue:
		return cachedSimpleTypeExpression("String")
	case runtime.BoolValue:
		return cachedSimpleTypeExpression("bool")
	case runtime.CharValue:
		return cachedSimpleTypeExpression("char")
	case runtime.NilValue:
		return cachedSimpleTypeExpression("nil")
	case runtime.VoidValue:
		return cachedSimpleTypeExpression("void")
	case runtime.IteratorEndValue:
		return cachedSimpleTypeExpression("IteratorEnd")
	case *runtime.IteratorEndValue:
		if v == nil {
			return nil
		}
		return cachedSimpleTypeExpression("IteratorEnd")
	case runtime.IntegerValue:
		return cachedIntegerTypeExpression(v.TypeSuffix)
	case *runtime.IntegerValue:
		if v == nil {
			return nil
		}
		return cachedIntegerTypeExpression(v.TypeSuffix)
	case runtime.FloatValue:
		return cachedFloatTypeExpression(v.TypeSuffix)
	case *runtime.FloatValue:
		if v == nil {
			return nil
		}
		return cachedFloatTypeExpression(v.TypeSuffix)
	case *runtime.HostHandleValue:
		if v == nil {
			return nil
		}
		if v.HandleType == "" {
			return nil
		}
		return cachedSimpleTypeExpression(v.HandleType)
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil
		}
		base := cachedSimpleTypeExpression(v.Definition.Node.ID.Name)
		if seen == nil {
			seen = make(map[*runtime.StructInstanceValue]struct{})
		}
		if _, ok := seen[v]; ok {
			return base
		}
		seen[v] = struct{}{}
		defer delete(seen, v)
		if v.Definition.Node.ID.Name == "Array" {
			if arr, err := i.arrayValueFromStructFields(v.Fields); err == nil && arr != nil {
				if inferred := i.typeExpressionFromValueWithSeen(arr, seen); inferred != nil {
					return inferred
				}
			}
			return cachedArrayTypeExpression(cachedWildcardTypeExpression)
		}
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
				typeArgs = i.inferStructTypeArgumentsWithSeen(v.Definition.Node, v.Fields, v.Positional, seen)
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
		return cachedSimpleTypeExpression(v.Interface.Node.ID.Name)
	case runtime.InterfaceValue:
		return i.typeExpressionFromValueWithSeen(&v, seen)
	case *runtime.ArrayValue:
		if v == nil {
			return nil
		}
		var elemType ast.TypeExpression
		if len(v.Elements) > 0 {
			elemType = i.typeExpressionFromValueWithSeen(v.Elements[0], seen)
		}
		if elemType == nil {
			elemType = cachedWildcardTypeExpression
		}
		return cachedArrayTypeExpression(elemType)
	case *runtime.IteratorValue:
		if v == nil {
			return nil
		}
		return cachedIteratorTypeExpression
	case runtime.ErrorValue:
		if v.TypeName != nil {
			return cachedSimpleTypeExpression(v.TypeName.Name)
		}
		return cachedSimpleTypeExpression("Error")
	default:
		return nil
	}
}

func (i *Interpreter) canonicalTypeNames(name string) []string {
	if name == "" {
		return nil
	}
	i.typeAliasCacheMu.RLock()
	cached, ok := i.typeAliasBaseCache[name]
	i.typeAliasCacheMu.RUnlock()
	if ok {
		return cached
	}
	expanded := aliasBaseTypeName(name, i.typeAliases)
	var result []string
	if expanded == "" || expanded == name {
		result = []string{name}
	} else {
		result = []string{name, expanded}
	}
	i.typeAliasCacheMu.Lock()
	if i.typeAliasBaseCache == nil {
		i.typeAliasBaseCache = make(map[string][]string)
	}
	if existing, ok := i.typeAliasBaseCache[name]; ok {
		i.typeAliasCacheMu.Unlock()
		return existing
	}
	i.typeAliasBaseCache[name] = result
	i.typeAliasCacheMu.Unlock()
	return result
}

func aliasBaseTypeName(name string, aliases map[string]*ast.TypeAliasDefinition) string {
	if name == "" || len(aliases) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, 4)
	current := name
	for {
		if _, ok := seen[current]; ok {
			return ""
		}
		seen[current] = struct{}{}
		alias, ok := aliases[current]
		if !ok || alias == nil || alias.TargetType == nil {
			return ""
		}
		switch t := alias.TargetType.(type) {
		case *ast.SimpleTypeExpression:
			if t == nil || t.Name == nil || t.Name.Name == "" {
				return ""
			}
			next := t.Name.Name
			if _, chained := aliases[next]; chained {
				current = next
				continue
			}
			return next
		case *ast.GenericTypeExpression:
			if t != nil {
				if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil && base.Name.Name != "" {
					return base.Name.Name
				}
			}
			if info, ok := parseTypeExpression(alias.TargetType); ok {
				return info.name
			}
			return ""
		default:
			if info, ok := parseTypeExpression(alias.TargetType); ok {
				return info.name
			}
			return ""
		}
	}
}
