package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type implMethodContext struct {
	implName      string
	interfaceName string
	target        ast.TypeExpression
	methods       map[string]runtime.Value
}

func (i *Interpreter) resolveMethodFromPool(env *runtime.Environment, funcName string, receiver runtime.Value, ifaceFilter string) (runtime.Value, error) {
	functionCandidates := make([]*runtime.FunctionValue, 0)
	nativeCandidates := make([]runtime.Value, 0)
	seenFuncs := make(map[*runtime.FunctionValue]struct{})
	seenNatives := make(map[string]struct{})
	var scopeCallable runtime.Value
	var scopeSet map[*runtime.FunctionValue]struct{}
	nameInScope := false
	if env != nil {
		if val, err := env.Get(funcName); err == nil && isCallableRuntimeValue(val) {
			nameInScope = true
			scopeCallable = val
			scopeSet = functionScopeSet(val)
		}
	}
	var info typeInfo
	var hasInfo bool
	if receiver != nil {
		info, hasInfo = i.getTypeInfoForValue(receiver)
	}
	typeNameInScope := false
	if hasInfo && env != nil {
		for _, name := range i.canonicalTypeNames(info.name) {
			if env.Has(name) {
				typeNameInScope = true
				break
			}
		}
	}
	allowInherent := nameInScope || typeNameInScope || isPrimitiveReceiver(receiver)

	var addCallable func(method runtime.Value, privacyContext string) error
	addCallable = func(method runtime.Value, privacyContext string) error {
		switch fn := method.(type) {
		case *runtime.FunctionValue:
			if fn == nil {
				return nil
			}
			if fnDecl, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDecl != nil && fnDecl.IsPrivate {
				name := privacyContext
				if name == "" {
					name = "<unknown>"
				}
				return fmt.Errorf("Method '%s' on %s is private", funcName, name)
			}
			if _, ok := seenFuncs[fn]; ok {
				return nil
			}
			seenFuncs[fn] = struct{}{}
			functionCandidates = append(functionCandidates, fn)
		case *runtime.FunctionOverloadValue:
			if fn == nil {
				return nil
			}
			for _, entry := range fn.Overloads {
				if entry == nil {
					continue
				}
				if err := addCallable(entry, privacyContext); err != nil {
					return err
				}
			}
		case runtime.NativeFunctionValue:
			key := fn.Name
			if key == "" {
				key = fmt.Sprintf("%p", &fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn})
		case *runtime.NativeFunctionValue:
			if fn == nil {
				return nil
			}
			key := fn.Name
			if key == "" {
				key = fmt.Sprintf("%p", fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, runtime.NativeBoundMethodValue{Receiver: receiver, Method: *fn})
		case runtime.NativeBoundMethodValue:
			key := fn.Method.Name
			if key == "" {
				key = fmt.Sprintf("%p", &fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, fn)
		case *runtime.NativeBoundMethodValue:
			if fn == nil {
				return nil
			}
			key := fn.Method.Name
			if key == "" {
				key = fmt.Sprintf("%p", fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, *fn)
		}
		return nil
	}

	if env != nil {
		if data := env.RuntimeData(); data != nil {
			if ctx, ok := data.(*implMethodContext); ok && ctx != nil && ctx.target != nil && receiver != nil {
				if ifaceFilter == "" || ifaceFilter == ctx.interfaceName {
					if i.matchesType(ctx.target, receiver) {
						if method := ctx.methods[funcName]; method != nil {
							if callable, ok := i.selectUfcsCallable(method, receiver, true, nil); ok {
								if err := checkPrivateMethod(funcName, ctx.implName, callable); err != nil {
									return nil, err
								}
								return &runtime.BoundMethodValue{Receiver: receiver, Method: callable}, nil
							}
						}
					}
				}
			}
		}
	}

	if hasInfo {
		if allowInherent {
			for _, name := range i.canonicalTypeNames(info.name) {
				if bucket, ok := i.inherentMethods[name]; ok {
					if method := bucket[funcName]; method != nil {
						if callable, ok := i.selectUfcsCallable(method, receiver, true, nil); ok {
							if err := addCallable(callable, name); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
		existing := len(functionCandidates) + len(nativeCandidates)
		if method, err := i.findMethod(info, funcName, ifaceFilter); err == nil && method != nil {
			if callable, ok := i.selectUfcsCallable(method, receiver, true, nil); ok {
				if err := addCallable(callable, info.name); err != nil {
					return nil, err
				}
			}
		} else if err != nil {
			if existing == 0 {
				return nil, err
			}
		}
	}

	hasMethodCandidate := len(functionCandidates)+len(nativeCandidates) > 0

	if env != nil && nameInScope && !hasMethodCandidate {
		if scopeCallable != nil {
			if callable, ok := i.selectUfcsCallable(scopeCallable, receiver, false, scopeSet); ok {
				if err := addCallable(callable, ""); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(functionCandidates) > 0 && len(nativeCandidates) > 0 {
		return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
	}
	if len(functionCandidates) > 0 {
		var callable runtime.Value
		if len(functionCandidates) == 1 {
			callable = functionCandidates[0]
		} else {
			callable = &runtime.FunctionOverloadValue{Overloads: functionCandidates}
		}
		return &runtime.BoundMethodValue{Receiver: receiver, Method: callable}, nil
	}
	if len(nativeCandidates) > 0 {
		if len(nativeCandidates) > 1 {
			return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
		}
		return nativeCandidates[0], nil
	}
	return nil, nil
}

func functionScopeSet(scope runtime.Value) map[*runtime.FunctionValue]struct{} {
	overloads := runtime.FlattenFunctionOverloads(scope)
	if len(overloads) == 0 {
		return nil
	}
	set := make(map[*runtime.FunctionValue]struct{}, len(overloads))
	for _, fn := range overloads {
		if fn != nil {
			set[fn] = struct{}{}
		}
	}
	return set
}

func (i *Interpreter) methodTargetMatchesReceiver(fn *runtime.FunctionValue, receiver runtime.Value) bool {
	if fn == nil || fn.MethodSet == nil || fn.MethodSet.TargetType == nil {
		return true
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil {
		return true
	}
	target := fn.MethodSet.TargetType
	var targetName string
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			targetName = t.Name.Name
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name != nil {
			targetName = base.Name.Name
		}
	}
	if targetName == "" || fn.Closure == nil {
		return true
	}
	def, ok := fn.Closure.StructDefinition(targetName)
	if !ok || def == nil {
		return true
	}
	return def == inst.Definition
}

func (i *Interpreter) selectUfcsCallable(method runtime.Value, receiver runtime.Value, requireSelf bool, scopeSet map[*runtime.FunctionValue]struct{}) (runtime.Value, bool) {
	switch fn := method.(type) {
	case *runtime.FunctionValue:
		if scopeSet != nil {
			if _, ok := scopeSet[fn]; !ok {
				return nil, false
			}
		}
		if (!requireSelf || functionExpectsSelf(fn)) && i.functionFirstParamMatches(fn, receiver) {
			if !requireSelf && fn.TypeQualified {
				return nil, false
			}
			if !i.methodTargetMatchesReceiver(fn, receiver) {
				return nil, false
			}
			return fn, true
		}
		return nil, false
	case *runtime.FunctionOverloadValue:
		filtered := make([]*runtime.FunctionValue, 0, len(fn.Overloads))
		for _, entry := range fn.Overloads {
			if entry == nil {
				continue
			}
			if scopeSet != nil {
				if _, ok := scopeSet[entry]; !ok {
					continue
				}
			}
			if !requireSelf && entry.TypeQualified {
				continue
			}
			if requireSelf && !functionExpectsSelf(entry) {
				continue
			}
			if !i.functionFirstParamMatches(entry, receiver) {
				continue
			}
			if !i.methodTargetMatchesReceiver(entry, receiver) {
				continue
			}
			filtered = append(filtered, entry)
		}
		if len(filtered) > 1 {
			targetFiltered := make([]*runtime.FunctionValue, 0, len(filtered))
			for _, entry := range filtered {
				if i.methodTargetMatchesReceiver(entry, receiver) {
					targetFiltered = append(targetFiltered, entry)
				}
			}
			if len(targetFiltered) == 1 {
				return targetFiltered[0], true
			}
			if len(targetFiltered) > 0 {
				filtered = targetFiltered
			}
		}
		if len(filtered) > 1 {
			if recvType := i.typeExpressionFromValue(receiver); recvType != nil {
				narrowed := make([]*runtime.FunctionValue, 0, len(filtered))
				for _, entry := range filtered {
					if def, ok := entry.Declaration.(*ast.FunctionDefinition); ok && def != nil {
						if len(def.Params) > 0 && def.Params[0] != nil && def.Params[0].ParamType != nil {
							if typeExpressionsEqual(def.Params[0].ParamType, recvType) {
								narrowed = append(narrowed, entry)
							}
						}
					}
				}
				if len(narrowed) == 1 {
					return narrowed[0], true
				}
				if len(narrowed) > 0 {
					filtered = narrowed
				}
			}
		}
		if len(filtered) == 0 {
			return nil, false
		}
		if len(filtered) == 1 {
			return filtered[0], true
		}
		return &runtime.FunctionOverloadValue{Overloads: filtered}, true
	default:
		return nil, false
	}
}

func checkPrivateMethod(funcName, context string, callable runtime.Value) error {
	name := context
	if name == "" {
		name = "<unknown>"
	}
	switch fn := callable.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil
		}
		if def, ok := fn.Declaration.(*ast.FunctionDefinition); ok && def != nil && def.IsPrivate {
			return fmt.Errorf("Method '%s' on %s is private", funcName, name)
		}
	case *runtime.FunctionOverloadValue:
		if fn == nil {
			return nil
		}
		for _, entry := range fn.Overloads {
			if entry == nil {
				continue
			}
			if def, ok := entry.Declaration.(*ast.FunctionDefinition); ok && def != nil && def.IsPrivate {
				return fmt.Errorf("Method '%s' on %s is private", funcName, name)
			}
		}
	}
	return nil
}

func functionExpectsSelf(fn *runtime.FunctionValue) bool {
	if fn == nil {
		return false
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil {
		return false
	}
	if decl.IsMethodShorthand {
		return true
	}
	if len(decl.Params) == 0 {
		return false
	}
	first := decl.Params[0]
	if first == nil {
		return false
	}
	if ident, ok := first.Name.(*ast.Identifier); ok && strings.EqualFold(ident.Name, "self") {
		return true
	}
	if simple, ok := first.ParamType.(*ast.SimpleTypeExpression); ok && simple.Name != nil && simple.Name.Name == "Self" {
		return true
	}
	return false
}

func (i *Interpreter) functionFirstParamMatches(fn *runtime.FunctionValue, receiver runtime.Value) bool {
	if fn == nil || fn.Declaration == nil {
		return false
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.IsMethodShorthand {
			return true
		}
		if len(decl.Params) == 0 {
			return false
		}
		first := decl.Params[0]
		if first == nil {
			return false
		}
		if first.ParamType == nil {
			return true
		}
		return i.matchesType(first.ParamType, receiver)
	case *ast.LambdaExpression:
		if len(decl.Params) == 0 {
			return false
		}
		first := decl.Params[0]
		if first == nil {
			return false
		}
		if first.ParamType == nil {
			return true
		}
		return i.matchesType(first.ParamType, receiver)
	default:
		return false
	}
}
