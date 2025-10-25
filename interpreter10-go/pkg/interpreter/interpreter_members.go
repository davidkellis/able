package interpreter

import (
	"fmt"
	"math"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateMemberAccess(expr *ast.MemberAccessExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	return i.memberAccessOnValue(obj, expr.Member, env)
}

func (i *Interpreter) memberAccessOnValue(obj runtime.Value, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	switch v := obj.(type) {
	case *runtime.StructDefinitionValue:
		return i.structDefinitionMember(v, member)
	case runtime.StructDefinitionValue:
		return i.structDefinitionMember(&v, member)
	case runtime.PackageValue:
		return i.packageMemberAccess(v, member)
	case *runtime.PackageValue:
		return i.packageMemberAccess(*v, member)
	case runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(v, member)
	case *runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(*v, member)
	case runtime.DynPackageValue:
		return i.dynPackageMemberAccess(v, member)
	case *runtime.DynPackageValue:
		return i.dynPackageMemberAccess(*v, member)
	case *runtime.StructInstanceValue:
		return i.structInstanceMember(v, member, env)
	case *runtime.InterfaceValue:
		return i.interfaceMember(v, member)
	case *runtime.ArrayValue:
		i.ensureArrayBuiltins()
		return i.arrayMember(v, member)
	case *runtime.HashMapValue:
		i.ensureHashMapBuiltins()
		return i.hashMapMember(v, member)
	case *runtime.HasherValue:
		return i.hasherMember(v, member)
	case *runtime.ProcHandleValue:
		return i.procHandleMember(v, member)
	case *runtime.FutureValue:
		return i.futureMember(v, member)
	case *runtime.IteratorValue:
		return i.iteratorMember(v, member)
	default:
		if ident, ok := member.(*ast.Identifier); ok {
			if bound, ok := i.tryUfcs(env, ident.Name, obj); ok {
				return bound, nil
			}
		}
		return nil, fmt.Errorf("Member access only supported on structs in this milestone")
	}
}

func (i *Interpreter) evaluateImplicitMemberExpression(expr *ast.ImplicitMemberExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr == nil || expr.Member == nil {
		return nil, fmt.Errorf("Implicit member requires identifier")
	}
	state := i.stateFromEnv(env)
	receiver, ok := state.currentImplicitReceiver()
	if !ok {
		return nil, fmt.Errorf("Implicit member '#%s' requires enclosing function with a first parameter", expr.Member.Name)
	}
	return i.memberAccessOnValue(receiver, expr.Member, env)
}

func (i *Interpreter) evaluateIndexExpression(expr *ast.IndexExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	idxVal, err := i.evaluateExpression(expr.Index, env)
	if err != nil {
		return nil, err
	}
	arr, err := toArrayValue(obj)
	if err != nil {
		return nil, err
	}
	idx, err := indexFromValue(idxVal)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(arr.Elements) {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	val := arr.Elements[idx]
	if val == nil {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	return val, nil
}

func toArrayValue(val runtime.Value) (*runtime.ArrayValue, error) {
	switch v := val.(type) {
	case *runtime.ArrayValue:
		return v, nil
	default:
		return nil, fmt.Errorf("Indexing is only supported on arrays")
	}
}

func indexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || !v.Val.IsInt64() {
			return 0, fmt.Errorf("Array index must be within int range")
		}
		return int(v.Val.Int64()), nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("Array index must be a number")
		}
		idx := int(math.Trunc(v.Val))
		return idx, nil
	default:
		return 0, fmt.Errorf("Array index must be a number")
	}
}

func (i *Interpreter) structInstanceMember(inst *runtime.StructInstanceValue, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
	if ident, ok := member.(*ast.Identifier); ok {
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		if val, ok := inst.Fields[ident.Name]; ok {
			return val, nil
		}
		if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			if bound, ok := i.tryUfcs(env, ident.Name, inst); ok {
				return bound, nil
			}
			return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
		}
		typeName := inst.Definition.Node.ID.Name
		if bucket, ok := i.inherentMethods[typeName]; ok {
			if method, ok := bucket[ident.Name]; ok {
				if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
					return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
				}
				return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
			}
		}
		method, err := i.selectStructMethod(inst, ident.Name)
		if err != nil {
			return nil, err
		}
		if method != nil {
			return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
		}
		if bound, ok := i.tryUfcs(env, ident.Name, inst); ok {
			return bound, nil
		}
		return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
	}
	if intLit, ok := member.(*ast.IntegerLiteral); ok {
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if intLit.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(intLit.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		return inst.Positional[idx], nil
	}
	return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
}

func (i *Interpreter) iteratorMember(iter *runtime.IteratorValue, member ast.Expression) (runtime.Value, error) {
	if iter == nil {
		return nil, fmt.Errorf("iterator receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("iterator member access expects identifier")
	}
	switch ident.Name {
	case "next":
		fn := runtime.NativeFunctionValue{
			Name:  "iterator.next",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("next expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.IteratorValue)
				if !ok {
					return nil, fmt.Errorf("next receiver must be an iterator")
				}
				value, done, err := receiver.Next()
				if err != nil {
					return nil, err
				}
				if done {
					return runtime.IteratorEnd, nil
				}
				if value == nil {
					return runtime.NilValue{}, nil
				}
				return value, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: iter, Method: fn}, nil
	default:
		return nil, fmt.Errorf("iterator has no member '%s'", ident.Name)
	}
}

func (i *Interpreter) tryUfcs(env *runtime.Environment, funcName string, receiver runtime.Value) (runtime.Value, bool) {
	if env == nil {
		return nil, false
	}
	val, err := env.Get(funcName)
	if err != nil {
		return nil, false
	}
	if fn, ok := val.(*runtime.FunctionValue); ok {
		return &runtime.BoundMethodValue{Receiver: receiver, Method: fn}, true
	}
	return nil, false
}

func (i *Interpreter) structDefinitionMember(def *runtime.StructDefinitionValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Static access expects identifier member")
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing identifier")
	}
	typeName := def.Node.ID.Name
	bucket := i.inherentMethods[typeName]
	if bucket == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	method, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
	}
	return method, nil
}

func (i *Interpreter) packageMemberAccess(pkg runtime.PackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Package member access expects identifier")
	}
	if pkg.Public == nil {
		return nil, fmt.Errorf("Package has no public members")
	}
	val, ok := pkg.Public[ident.Name]
	if !ok {
		pkgName := pkg.Name
		if pkgName == "" {
			pkgName = strings.Join(pkg.NamePath, ".")
		}
		if pkgName == "" {
			pkgName = "<package>"
		}
		return nil, fmt.Errorf("No public member '%s' on package %s", ident.Name, pkgName)
	}
	return val, nil
}

func (i *Interpreter) dynPackageMemberAccess(pkg runtime.DynPackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Dyn package member access expects identifier")
	}
	pkgName := pkg.Name
	if pkgName == "" {
		pkgName = strings.Join(pkg.NamePath, ".")
	}
	if pkgName == "" {
		return nil, fmt.Errorf("Dyn package missing name")
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' not found", pkgName)
	}
	sym, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' has no member '%s'", pkgName, ident.Name)
	}
	if isPrivateSymbol(sym) {
		return nil, fmt.Errorf("dyn package '%s' member '%s' is private", pkgName, ident.Name)
	}
	return runtime.DynRefValue{Package: pkgName, Name: ident.Name}, nil
}

func (i *Interpreter) implNamespaceMember(ns runtime.ImplementationNamespaceValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Impl namespace member access expects identifier")
	}
	if ns.Methods == nil {
		return nil, fmt.Errorf("Impl namespace has no methods")
	}
	method, ok := ns.Methods[ident.Name]
	if !ok {
		name := "<impl>"
		if ns.Name != nil {
			name = ns.Name.Name
		}
		return nil, fmt.Errorf("No method '%s' on impl %s", ident.Name, name)
	}
	return method, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := ""
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		ifaceName = val.Interface.Node.ID.Name
	}
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	var method *runtime.FunctionValue
	if val.Methods != nil {
		method = val.Methods[ident.Name]
	}
	if method == nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			resolved, err := i.findMethod(info, ident.Name, ifaceName)
			if err != nil {
				return nil, err
			}
			method = resolved
			if method != nil {
				if val.Methods == nil {
					val.Methods = make(map[string]*runtime.FunctionValue)
				}
				val.Methods[ident.Name] = method
			}
		}
	}
	if method == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, ifaceName)
	}
	return &runtime.BoundMethodValue{Receiver: val.Underlying, Method: method}, nil
}

func (i *Interpreter) resolveDynRef(ref runtime.DynRefValue) (*runtime.FunctionValue, error) {
	bucket, ok := i.packageRegistry[ref.Package]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	val, ok := bucket[ref.Name]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	if fn, ok := val.(*runtime.FunctionValue); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("dyn ref '%s.%s' is not callable", ref.Package, ref.Name)
}

func toStructDefinitionValue(val runtime.Value, name string) (*runtime.StructDefinitionValue, error) {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		return v, nil
	case runtime.StructDefinitionValue:
		return &v, nil
	default:
		return nil, fmt.Errorf("'%s' is not a struct type", name)
	}
}
