package bridge

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

const NativeIntBits = strconv.IntSize

type Runtime struct {
	interp    *interpreter.Interpreter
	mu        sync.RWMutex
	originals map[string]runtime.Value
	structs   map[string]*runtime.StructDefinitionValue
	env       *runtime.Environment
}

func New(interp *interpreter.Interpreter) *Runtime {
	return &Runtime{
		interp:    interp,
		originals: make(map[string]runtime.Value),
		structs:   make(map[string]*runtime.StructDefinitionValue),
	}
}

func (r *Runtime) SetEnv(env *runtime.Environment) {
	if r == nil || env == nil {
		return
	}
	r.mu.Lock()
	r.env = env
	r.mu.Unlock()
}

func (r *Runtime) Env() *runtime.Environment {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	env := r.env
	r.mu.RUnlock()
	return env
}

func (r *Runtime) SwapEnv(env *runtime.Environment) *runtime.Environment {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	prev := r.env
	r.env = env
	r.mu.Unlock()
	return prev
}

func (r *Runtime) currentEnv() *runtime.Environment {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	env := r.env
	r.mu.RUnlock()
	if env == nil && r.interp != nil {
		env = r.interp.GlobalEnvironment()
	}
	return env
}

func (r *Runtime) RegisterOriginal(name string, value runtime.Value) {
	if r == nil || name == "" || value == nil {
		return
	}
	r.mu.Lock()
	if _, exists := r.originals[name]; !exists {
		r.originals[name] = value
	}
	r.mu.Unlock()
}

func (r *Runtime) CallOriginal(name string, args []runtime.Value) (runtime.Value, error) {
	if r == nil || r.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	r.mu.RLock()
	orig, ok := r.originals[name]
	r.mu.RUnlock()
	if !ok || orig == nil {
		return nil, fmt.Errorf("compiler bridge: original function %s not found", name)
	}
	env := r.currentEnv()
	return r.interp.CallFunctionIn(orig, args, env)
}

func (r *Runtime) Call(name string, args []runtime.Value) (runtime.Value, error) {
	if r == nil || r.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := r.currentEnv()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	value, err := env.Get(name)
	if err != nil {
		return nil, err
	}
	return r.interp.CallFunctionIn(value, args, env)
}

func Get(rt *Runtime, name string) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	value, err := env.Get(name)
	if err == nil {
		return value, nil
	}
	if env != rt.interp.GlobalEnvironment() {
		if fallback := rt.interp.GlobalEnvironment(); fallback != nil {
			if value, err := fallback.Get(name); err == nil {
				return value, nil
			}
		}
	}
	return nil, err
}

func Assign(rt *Runtime, name string, value runtime.Value) error {
	if rt == nil || rt.interp == nil {
		return fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	if env == nil {
		return fmt.Errorf("compiler bridge: missing global environment")
	}
	if env.AssignExisting(name, value) {
		return nil
	}
	env.Define(name, value)
	return nil
}

func (r *Runtime) StructDefinition(name string) (*runtime.StructDefinitionValue, error) {
	if r == nil || r.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	r.mu.RLock()
	if def, ok := r.structs[name]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	env := r.currentEnv()
	r.mu.RUnlock()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	def, ok := env.StructDefinition(name)
	if (!ok || def == nil) && env != r.interp.GlobalEnvironment() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			if alt, found := fallback.StructDefinition(name); found && alt != nil {
				def, ok = alt, true
			}
		}
	}
	if !ok || def == nil {
		return nil, fmt.Errorf("compiler bridge: struct %s not found", name)
	}
	r.mu.Lock()
	r.structs[name] = def
	r.mu.Unlock()
	return def, nil
}

func Index(rt *Runtime, obj runtime.Value, idx runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.IndexGet(obj, idx, nil)
}

func IndexAssign(rt *Runtime, obj runtime.Value, idx runtime.Value, value runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.IndexAssign(obj, idx, value, nil)
}

func MemberAssign(rt *Runtime, obj runtime.Value, member runtime.Value, value runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.MemberAssign(obj, member, value, nil)
}

func MemberGet(rt *Runtime, obj runtime.Value, member runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.MemberGet(obj, member, env)
}

func MemberGetPreferMethods(rt *Runtime, obj runtime.Value, member runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.MemberGetPreferMethods(obj, member, env)
}

func CallValue(rt *Runtime, fn runtime.Value, args []runtime.Value) (runtime.Value, error) {
	return CallValueWithNode(rt, fn, args, nil)
}

func CallValueWithNode(rt *Runtime, fn runtime.Value, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	val, err := rt.interp.CallFunctionInWithCallNode(fn, args, env, call)
	if err != nil && call != nil {
		err = rt.interp.AttachRuntimeContext(err, call, env)
	}
	return val, err
}

func CallNamed(rt *Runtime, name string, args []runtime.Value) (runtime.Value, error) {
	return CallNamedWithNode(rt, name, args, nil)
}

func CallNamedWithNode(rt *Runtime, name string, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	value, err := env.Get(name)
	if err == nil {
		val, callErr := rt.interp.CallFunctionInWithCallNode(value, args, env, call)
		if callErr != nil && call != nil {
			callErr = rt.interp.AttachRuntimeContext(callErr, call, env)
		}
		return val, callErr
	}
	if dot := strings.Index(name, "."); dot > 0 && dot < len(name)-1 {
		head := name[:dot]
		tail := name[dot+1:]
		receiver, recvErr := env.Get(head)
		if recvErr != nil {
			if def, ok := env.StructDefinition(head); ok {
				receiver = def
			} else {
				receiver = runtime.TypeRefValue{TypeName: head}
			}
		}
		member := runtime.StringValue{Val: tail}
		candidate, err := rt.interp.MemberGetPreferMethods(receiver, member, env)
		if err != nil {
			return nil, err
		}
		val, callErr := rt.interp.CallFunctionInWithCallNode(candidate, args, env, call)
		if callErr != nil && call != nil {
			callErr = rt.interp.AttachRuntimeContext(callErr, call, env)
		}
		return val, callErr
	}
	return nil, err
}

func Stringify(rt *Runtime, value runtime.Value) (string, error) {
	if rt == nil || rt.interp == nil {
		return "", fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.Stringify(value, env)
}

func IsError(rt *Runtime, value runtime.Value) bool {
	if value == nil {
		return false
	}
	if rt == nil || rt.interp == nil {
		switch value.(type) {
		case runtime.ErrorValue, *runtime.ErrorValue:
			return true
		default:
			return false
		}
	}
	return rt.interp.IsErrorValue(value)
}

func IsTruthy(rt *Runtime, value runtime.Value) bool {
	if value == nil {
		return false
	}
	if rt == nil || rt.interp == nil {
		switch v := value.(type) {
		case runtime.BoolValue:
			return v.Val
		case *runtime.BoolValue:
			return v != nil && v.Val
		case runtime.NilValue, *runtime.NilValue:
			return false
		case runtime.ErrorValue, *runtime.ErrorValue:
			return false
		case runtime.InterfaceValue:
			if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
				return false
			}
		case *runtime.InterfaceValue:
			if v != nil && v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
				return false
			}
		}
		return true
	}
	return rt.interp.IsTruthy(value)
}

func ErrorValue(rt *Runtime, value runtime.Value) runtime.ErrorValue {
	switch v := value.(type) {
	case runtime.ErrorValue:
		return v
	case *runtime.ErrorValue:
		if v != nil {
			return *v
		}
	}
	if rt == nil || rt.interp == nil {
		payload := map[string]runtime.Value{}
		if value != nil {
			payload["value"] = value
		}
		return runtime.ErrorValue{Message: fmt.Sprintf("%v", value), Payload: payload}
	}
	env := rt.currentEnv()
	return rt.interp.MakeErrorValue(value, env)
}

func DivisionByZeroError(rt *Runtime) runtime.Value {
	if rt == nil || rt.interp == nil {
		return runtime.ErrorValue{Message: "division by zero"}
	}
	return rt.interp.StandardDivisionByZeroErrorValue()
}

func OverflowError(rt *Runtime, operation string) runtime.Value {
	message := operation
	if message == "" {
		message = "integer overflow"
	}
	if rt == nil || rt.interp == nil {
		return runtime.ErrorValue{Message: message}
	}
	return rt.interp.StandardOverflowErrorValue(operation)
}

func ShiftOutOfRangeError(rt *Runtime, shift int64) runtime.Value {
	if rt == nil || rt.interp == nil {
		return runtime.ErrorValue{Message: "shift out of range"}
	}
	return rt.interp.StandardShiftOutOfRangeErrorValue(shift)
}

func ApplyBinaryOperator(rt *Runtime, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.ApplyBinaryOperator(op, left, right)
}

func ApplyUnaryOperator(rt *Runtime, op string, operand runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.ApplyUnaryOperator(op, operand)
}

func Range(rt *Runtime, start runtime.Value, end runtime.Value, inclusive bool) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.EvaluateRangeValues(start, end, inclusive, env)
}

func ResolveIterator(rt *Runtime, iterable runtime.Value) (*runtime.IteratorValue, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.ResolveIteratorValue(iterable, env)
}

func Spawn(rt *Runtime, task func(*runtime.Environment) (runtime.Value, error)) (*runtime.FutureValue, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if task == nil {
		return nil, fmt.Errorf("compiler bridge: missing task")
	}
	env := rt.currentEnv()
	future := rt.interp.RunCompiledFuture(env, func(taskEnv *runtime.Environment) (runtime.Value, error) {
		prev := rt.SwapEnv(taskEnv)
		defer rt.SwapEnv(prev)
		return task(taskEnv)
	})
	if future == nil {
		return nil, fmt.Errorf("compiler bridge: spawn failed")
	}
	return future, nil
}

func Await(rt *Runtime, expr *ast.AwaitExpression, iterable runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if expr == nil {
		return nil, fmt.Errorf("compiler bridge: missing await expression")
	}
	env := rt.currentEnv()
	return rt.interp.AwaitIterable(expr, iterable, env)
}

func ArrayElements(rt *Runtime, arr *runtime.ArrayValue) ([]runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.ArrayElements(arr)
}

func Cast(rt *Runtime, typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.CastValueToType(typeExpr, value)
}

// MatchType checks whether a value matches a type expression and returns the coerced value when it does.
func MatchType(rt *Runtime, typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool, error) {
	if rt == nil || rt.interp == nil {
		return nil, false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if !rt.interp.MatchesType(typeExpr, value) {
		return nil, false, nil
	}
	coerced, err := rt.interp.CoerceValueToType(typeExpr, value)
	if err != nil {
		return nil, false, err
	}
	if coerced == nil {
		coerced = runtime.NilValue{}
	}
	return coerced, true, nil
}

// Raise panics with the provided value so compiled code can signal a runtime error.
func Raise(value runtime.Value) {
	panic(value)
}

// RaiseWithContext raises a value with attached runtime diagnostics.
func RaiseWithContext(rt *Runtime, node ast.Node, value runtime.Value) {
	if rt == nil || rt.interp == nil {
		panic(value)
	}
	env := rt.currentEnv()
	err := interpreter.Raise(rt.interp, value, env)
	if node != nil {
		err = rt.interp.AttachRuntimeContext(err, node, env)
	}
	panic(err)
}

// RaiseRuntimeErrorWithContext attaches runtime diagnostics to an error and panics.
func RaiseRuntimeErrorWithContext(rt *Runtime, node ast.Node, err error) {
	if err == nil {
		return
	}
	if rt == nil || rt.interp == nil {
		panic(err)
	}
	env := rt.currentEnv()
	panic(rt.interp.AttachRuntimeContext(err, node, env))
}

// RegisterNodeOrigin wires a node origin path for compiled diagnostics.
func RegisterNodeOrigin(rt *Runtime, node ast.Node, origin string) {
	if rt == nil || rt.interp == nil || node == nil || origin == "" {
		return
	}
	rt.interp.AddNodeOrigin(node, origin)
}

// PushCallFrame records a call expression in the interpreter's runtime state.
func PushCallFrame(rt *Runtime, call *ast.FunctionCall) {
	if rt == nil || rt.interp == nil || call == nil {
		return
	}
	env := rt.currentEnv()
	rt.interp.PushCallFrame(env, call)
}

// PopCallFrame removes the most recent call expression frame.
func PopCallFrame(rt *Runtime) {
	if rt == nil || rt.interp == nil {
		return
	}
	env := rt.currentEnv()
	rt.interp.PopCallFrame(env)
}

// Recover converts a recovered panic into a runtime error compatible with the interpreter.
func Recover(rt *Runtime, ctx *runtime.NativeCallContext, recovered any) error {
	if recovered == nil {
		return nil
	}
	if err, ok := recovered.(error); ok {
		return err
	}
	if val, ok := recovered.(runtime.Value); ok {
		var env *runtime.Environment
		if ctx != nil {
			env = ctx.Env
		}
		if rt == nil || rt.interp == nil {
			return fmt.Errorf("panic: %v", val)
		}
		return interpreter.Raise(rt.interp, val, env)
	}
	return fmt.Errorf("panic: %v", recovered)
}

func AsString(value runtime.Value) (string, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val, nil
	case *runtime.StringValue:
		if v == nil {
			return "", fmt.Errorf("expected String, got nil")
		}
		return v.Val, nil
	case *runtime.StructInstanceValue:
		return stringFromStruct(v)
	default:
		return "", fmt.Errorf("expected String, got %T", value)
	}
}

func stringFromStruct(inst *runtime.StructInstanceValue) (string, error) {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return "", fmt.Errorf("expected String, got nil")
	}
	if inst.Definition.Node.ID.Name != "String" {
		return "", fmt.Errorf("expected String, got %T", inst)
	}
	var bytesVal runtime.Value
	if inst.Fields != nil {
		if field, ok := inst.Fields["bytes"]; ok {
			bytesVal = field
		}
	}
	if bytesVal == nil && len(inst.Positional) > 0 {
		bytesVal = inst.Positional[0]
	}
	if bytesVal == nil {
		return "", fmt.Errorf("string bytes are missing")
	}
	arr, err := arrayValueFromRuntime(bytesVal)
	if err != nil {
		return "", err
	}
	if arr == nil {
		return "", fmt.Errorf("string bytes are missing")
	}
	bytes := make([]byte, len(arr.Elements))
	maxByte := big.NewInt(0xff)
	for idx, elem := range arr.Elements {
		intVal, err := extractInteger(elem)
		if err != nil {
			return "", err
		}
		if intVal.Sign() < 0 || intVal.Cmp(maxByte) > 0 {
			return "", fmt.Errorf("string byte out of range")
		}
		bytes[idx] = byte(intVal.Int64())
	}
	return string(bytes), nil
}

func ToString(value string) runtime.Value {
	return runtime.StringValue{Val: value}
}

func AsBool(value runtime.Value) (bool, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.BoolValue:
		return v.Val, nil
	case *runtime.BoolValue:
		if v == nil {
			return false, fmt.Errorf("expected bool, got nil")
		}
		return v.Val, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", value)
	}
}

func ToBool(value bool) runtime.Value {
	return runtime.BoolValue{Val: value}
}

func AsRune(value runtime.Value) (rune, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.CharValue:
		return v.Val, nil
	case *runtime.CharValue:
		if v == nil {
			return 0, fmt.Errorf("expected char, got nil")
		}
		return v.Val, nil
	default:
		return 0, fmt.Errorf("expected char, got %T", value)
	}
}

func ToRune(value rune) runtime.Value {
	return runtime.CharValue{Val: value}
}

func AsFloat(value runtime.Value) (float64, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case *runtime.FloatValue:
		if v == nil {
			return 0, fmt.Errorf("expected float, got nil")
		}
		return v.Val, nil
	case runtime.IntegerValue:
		if v.Val == nil {
			return 0, fmt.Errorf("expected float, got nil")
		}
		return bigIntToFloat(v.Val), nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil {
			return 0, fmt.Errorf("expected float, got nil")
		}
		return bigIntToFloat(v.Val), nil
	default:
		return 0, fmt.Errorf("expected float, got %T", value)
	}
}

func ToFloat64(value float64) runtime.Value {
	return runtime.FloatValue{Val: value, TypeSuffix: runtime.FloatF64}
}

func ToFloat32(value float32) runtime.Value {
	return runtime.FloatValue{Val: float64(value), TypeSuffix: runtime.FloatF32}
}

func AsInt(value runtime.Value, bits int) (int64, error) {
	val, err := extractInteger(value)
	if err != nil {
		return 0, err
	}
	min, max := signedRange(bits)
	if val.Cmp(min) < 0 || val.Cmp(max) > 0 {
		return 0, fmt.Errorf("integer %s overflows %d-bit signed", val.String(), bits)
	}
	return val.Int64(), nil
}

func AsUint(value runtime.Value, bits int) (uint64, error) {
	val, err := extractInteger(value)
	if err != nil {
		return 0, err
	}
	if val.Sign() < 0 {
		return 0, fmt.Errorf("integer %s is negative for unsigned", val.String())
	}
	_, max := unsignedRange(bits)
	if val.Cmp(max) > 0 {
		return 0, fmt.Errorf("integer %s overflows %d-bit unsigned", val.String(), bits)
	}
	return val.Uint64(), nil
}

func ToInt(value int64, suffix runtime.IntegerType) runtime.Value {
	return runtime.IntegerValue{
		Val:        big.NewInt(value),
		TypeSuffix: suffix,
	}
}

func ToUint(value uint64, suffix runtime.IntegerType) runtime.Value {
	val := new(big.Int)
	val.SetUint64(value)
	return runtime.IntegerValue{
		Val:        val,
		TypeSuffix: suffix,
	}
}

func unwrapInterface(value runtime.Value) runtime.Value {
	for {
		switch v := value.(type) {
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v != nil {
				value = v.Underlying
				continue
			}
		}
		break
	}
	return value
}

func arrayValueFromRuntime(value runtime.Value) (*runtime.ArrayValue, error) {
	switch v := value.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return nil, fmt.Errorf("string bytes are missing")
		}
		return v, nil
	default:
		return nil, fmt.Errorf("string bytes must be an array")
	}
}

func bigIntToFloat(val *big.Int) float64 {
	if val == nil {
		return 0
	}
	f := new(big.Float).SetInt(val)
	result, _ := f.Float64()
	return result
}

func extractInteger(value runtime.Value) (*big.Int, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val == nil {
			return nil, fmt.Errorf("expected integer, got nil")
		}
		return v.Val, nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil {
			return nil, fmt.Errorf("expected integer, got nil")
		}
		return v.Val, nil
	default:
		return nil, fmt.Errorf("expected integer, got %T", value)
	}
}

func signedRange(bits int) (*big.Int, *big.Int) {
	if bits <= 0 || bits > 64 {
		bits = 64
	}
	one := big.NewInt(1)
	max := new(big.Int).Lsh(one, uint(bits-1))
	max.Sub(max, one)
	min := new(big.Int).Neg(new(big.Int).Lsh(one, uint(bits-1)))
	return min, max
}

func unsignedRange(bits int) (*big.Int, *big.Int) {
	if bits <= 0 || bits > 64 {
		bits = 64
	}
	one := big.NewInt(1)
	max := new(big.Int).Lsh(one, uint(bits))
	max.Sub(max, one)
	return big.NewInt(0), max
}
