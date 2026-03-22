package bridge

import (
	"fmt"
	goruntime "runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

const NativeIntBits = strconv.IntSize

var (
	memberGetPreferMethodsCalls          int64
	memberGetPreferMethodsInterfaceCalls int64
	memberGetPreferMethodsMu             sync.Mutex
	memberGetPreferMethodsNames          map[string]int64
	globalLookupFallbackCalls            int64
	globalLookupFallbackEnvCalls         int64
	globalLookupFallbackRegistryCalls    int64
	globalLookupFallbackMu               sync.Mutex
	globalLookupFallbackNames            map[string]int64
)

type Runtime struct {
	interp                      *interpreter.Interpreter
	mu                          sync.RWMutex
	originals                   map[string]runtime.Value
	structs                     map[string]*runtime.StructDefinitionValue
	env                         *runtime.Environment
	envByGID                    sync.Map
	concurrent                  int32 // atomic: 0 = single goroutine (fast path), 1 = concurrent
	resolver                    QualifiedCallableResolver
	globalLookupFallbackEnabled bool
}

type QualifiedCallableResolver func(name string, env *runtime.Environment) (runtime.Value, bool, error)

func New(interp *interpreter.Interpreter) *Runtime {
	return &Runtime{
		interp:                      interp,
		originals:                   make(map[string]runtime.Value),
		structs:                     make(map[string]*runtime.StructDefinitionValue),
		globalLookupFallbackEnabled: true,
	}
}

func (r *Runtime) SetGlobalLookupFallbackEnabled(enabled bool) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.globalLookupFallbackEnabled = enabled
	r.mu.Unlock()
}

func (r *Runtime) globalLookupFallback() bool {
	if r == nil {
		return true
	}
	r.mu.RLock()
	enabled := r.globalLookupFallbackEnabled
	r.mu.RUnlock()
	return enabled
}

func ExecutorKind(r *Runtime) string {
	if r == nil || r.interp == nil {
		return "serial"
	}
	return r.interp.ExecutorKind()
}

// HasInterpreter reports whether runtime bridge helpers can delegate to interpreter semantics.
func HasInterpreter(r *Runtime) bool {
	return r != nil && r.interp != nil
}

// MarkConcurrent switches the runtime to per-goroutine environment tracking.
// Must be called before spawning goroutines that use the runtime.
func (r *Runtime) MarkConcurrent() {
	if r == nil {
		return
	}
	atomic.StoreInt32(&r.concurrent, 1)
}

func (r *Runtime) isConcurrent() bool {
	return r != nil && atomic.LoadInt32(&r.concurrent) != 0
}

func (r *Runtime) SetEnv(env *runtime.Environment) {
	if r == nil || env == nil {
		return
	}
	r.mu.Lock()
	r.env = env
	r.mu.Unlock()
	if r.isConcurrent() {
		r.envByGID.Store(currentGID(), env)
	}
}

func (r *Runtime) Env() *runtime.Environment {
	if r == nil {
		return nil
	}
	if r.isConcurrent() {
		if env, ok := r.envByGID.Load(currentGID()); ok {
			if typed, ok := env.(*runtime.Environment); ok && typed != nil {
				return typed
			}
		}
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
	if !r.isConcurrent() {
		// Fast path: single goroutine, simple swap
		r.mu.Lock()
		prev := r.env
		if env != nil {
			r.env = env
		}
		r.mu.Unlock()
		return prev
	}
	gid := currentGID()
	var prev *runtime.Environment
	if existing, ok := r.envByGID.Load(gid); ok {
		if typed, ok := existing.(*runtime.Environment); ok {
			prev = typed
		}
	} else {
		r.mu.RLock()
		prev = r.env
		r.mu.RUnlock()
	}
	if env == nil {
		r.envByGID.Delete(gid)
	} else {
		r.envByGID.Store(gid, env)
	}
	return prev
}

func (r *Runtime) currentEnv() *runtime.Environment {
	if r == nil {
		return nil
	}
	if r.isConcurrent() {
		if env, ok := r.envByGID.Load(currentGID()); ok {
			if typed, ok := env.(*runtime.Environment); ok && typed != nil {
				return typed
			}
		}
	}
	r.mu.RLock()
	env := r.env
	r.mu.RUnlock()
	if env == nil && r.interp != nil {
		env = r.interp.GlobalEnvironment()
	}
	return env
}

func currentGID() uint64 {
	var buf [64]byte
	n := goruntime.Stack(buf[:], false)
	if n <= 10 {
		return 0
	}
	var id uint64
	for i := 10; i < n; i++ {
		c := buf[i]
		if c < '0' || c > '9' {
			break
		}
		id = id*10 + uint64(c-'0')
	}
	return id
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

func (r *Runtime) SetQualifiedCallableResolver(resolver QualifiedCallableResolver) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.resolver = resolver
	r.mu.Unlock()
}

func (r *Runtime) resolveQualifiedCallable(name string, env *runtime.Environment) (runtime.Value, bool, error) {
	if r == nil {
		return nil, false, nil
	}
	r.mu.RLock()
	resolver := r.resolver
	r.mu.RUnlock()
	if resolver == nil {
		return nil, false, nil
	}
	return resolver(name, env)
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
	if err != nil && env != r.interp.GlobalEnvironment() && r.globalLookupFallback() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			if alt, altErr := fallback.Get(name); altErr == nil {
				recordGlobalLookupFallback("call", name)
				value, err = alt, nil
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return r.interp.CallFunctionIn(value, args, env)
}

func Get(rt *Runtime, name string) (runtime.Value, error) {
	if rt == nil {
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
	if rt.interp != nil && env != rt.interp.GlobalEnvironment() && rt.globalLookupFallback() {
		if fallback := rt.interp.GlobalEnvironment(); fallback != nil {
			if value, err := fallback.Get(name); err == nil {
				recordGlobalLookupFallback("get", name)
				return value, nil
			}
		}
	}
	return nil, err
}

func Assign(rt *Runtime, name string, value runtime.Value) error {
	if rt == nil {
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
	if r == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := r.currentEnv()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	cacheKey := structCacheKey(env, name)
	r.mu.RLock()
	if def, ok := r.structs[cacheKey]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	r.mu.RUnlock()
	def, ok := env.StructDefinition(name)
	if (!ok || def == nil) && r.interp != nil {
		if seeded, found := r.interp.LookupStructDefinition(name); found && seeded != nil {
			def, ok = seeded, true
			env.DefineStruct(name, seeded)
			if seeded.Node != nil && seeded.Node.ID != nil {
				if canonical := strings.TrimSpace(seeded.Node.ID.Name); canonical != "" && canonical != name {
					env.DefineStruct(canonical, seeded)
				}
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && env != r.interp.GlobalEnvironment() && r.globalLookupFallback() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			if alt, found := fallback.StructDefinition(name); found && alt != nil {
				recordGlobalLookupFallback("struct_global", name)
				def, ok = alt, true
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && r.globalLookupFallback() {
		if alt, found := r.interp.LookupStructDefinition(name); found && alt != nil {
			recordGlobalLookupFallback("struct_registry", name)
			def, ok = alt, true
		}
	}
	if !ok || def == nil {
		return nil, fmt.Errorf("compiler bridge: struct %s not found", name)
	}
	r.mu.Lock()
	r.structs[cacheKey] = def
	r.mu.Unlock()
	return def, nil
}

func structCacheKey(env *runtime.Environment, name string) string {
	if env == nil {
		return "<nil>:" + name
	}
	return fmt.Sprintf("%p:%s", env, name)
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

func HashMapHashValue(rt *Runtime, val runtime.Value) (uint64, error) {
	if rt == nil || rt.interp == nil {
		return 0, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.HashMapHashValue(val)
}

func HashMapKeysEqual(rt *Runtime, a runtime.Value, b runtime.Value) (bool, error) {
	if rt == nil || rt.interp == nil {
		return false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.HashMapKeysEqual(a, b)
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
	atomic.AddInt64(&memberGetPreferMethodsCalls, 1)
	if isInterfaceReceiver(obj) {
		atomic.AddInt64(&memberGetPreferMethodsInterfaceCalls, 1)
	}
	if name := memberGetPreferMethodsName(obj, member); name != "" {
		memberGetPreferMethodsMu.Lock()
		if memberGetPreferMethodsNames == nil {
			memberGetPreferMethodsNames = make(map[string]int64)
		}
		memberGetPreferMethodsNames[name]++
		memberGetPreferMethodsMu.Unlock()
	}
	env := rt.currentEnv()
	return rt.interp.MemberGetPreferMethods(obj, member, env)
}

func ResetMemberGetPreferMethodsCounters() {
	atomic.StoreInt64(&memberGetPreferMethodsCalls, 0)
	atomic.StoreInt64(&memberGetPreferMethodsInterfaceCalls, 0)
	memberGetPreferMethodsMu.Lock()
	memberGetPreferMethodsNames = nil
	memberGetPreferMethodsMu.Unlock()
}

func MemberGetPreferMethodsStats() (calls int64, interfaceCalls int64) {
	calls = atomic.LoadInt64(&memberGetPreferMethodsCalls)
	interfaceCalls = atomic.LoadInt64(&memberGetPreferMethodsInterfaceCalls)
	return calls, interfaceCalls
}

func MemberGetPreferMethodsSnapshot() string {
	memberGetPreferMethodsMu.Lock()
	defer memberGetPreferMethodsMu.Unlock()
	if len(memberGetPreferMethodsNames) == 0 {
		return ""
	}
	keys := make([]string, 0, len(memberGetPreferMethodsNames))
	for name := range memberGetPreferMethodsNames {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", name, memberGetPreferMethodsNames[name]))
	}
	return strings.Join(parts, ",")
}

func recordGlobalLookupFallback(kind string, name string) {
	key := strings.TrimSpace(kind)
	if key == "" {
		key = "unknown"
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName != "" {
		key = key + ":" + trimmedName
	}
	atomic.AddInt64(&globalLookupFallbackCalls, 1)
	if key == "struct_registry" || strings.HasPrefix(key, "struct_registry:") {
		atomic.AddInt64(&globalLookupFallbackRegistryCalls, 1)
	} else {
		atomic.AddInt64(&globalLookupFallbackEnvCalls, 1)
	}
	globalLookupFallbackMu.Lock()
	if globalLookupFallbackNames == nil {
		globalLookupFallbackNames = make(map[string]int64)
	}
	globalLookupFallbackNames[key]++
	globalLookupFallbackMu.Unlock()
}

func ResetGlobalLookupFallbackCounters() {
	atomic.StoreInt64(&globalLookupFallbackCalls, 0)
	atomic.StoreInt64(&globalLookupFallbackEnvCalls, 0)
	atomic.StoreInt64(&globalLookupFallbackRegistryCalls, 0)
	globalLookupFallbackMu.Lock()
	globalLookupFallbackNames = nil
	globalLookupFallbackMu.Unlock()
}

func GlobalLookupFallbackStats() int64 {
	return atomic.LoadInt64(&globalLookupFallbackCalls)
}

func GlobalLookupFallbackBucketStats() (envCalls int64, registryCalls int64) {
	envCalls = atomic.LoadInt64(&globalLookupFallbackEnvCalls)
	registryCalls = atomic.LoadInt64(&globalLookupFallbackRegistryCalls)
	return envCalls, registryCalls
}

func GlobalLookupFallbackSnapshot() string {
	globalLookupFallbackMu.Lock()
	defer globalLookupFallbackMu.Unlock()
	if len(globalLookupFallbackNames) == 0 {
		return ""
	}
	keys := make([]string, 0, len(globalLookupFallbackNames))
	for name := range globalLookupFallbackNames {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", name, globalLookupFallbackNames[name]))
	}
	return strings.Join(parts, ",")
}

func isInterfaceReceiver(value runtime.Value) bool {
	switch value.(type) {
	case runtime.InterfaceValue, *runtime.InterfaceValue:
		return true
	default:
		return false
	}
}

func memberGetPreferMethodsName(obj runtime.Value, member runtime.Value) string {
	memberName := ""
	switch typed := member.(type) {
	case runtime.StringValue:
		memberName = strings.TrimSpace(typed.Val)
	case *runtime.StringValue:
		if typed == nil {
			return ""
		}
		memberName = strings.TrimSpace(typed.Val)
	default:
		return ""
	}
	if memberName == "" {
		return ""
	}
	receiverName := memberGetPreferMethodsReceiverName(obj)
	if receiverName == "" {
		return memberName
	}
	return receiverName + "." + memberName
}

func memberGetPreferMethodsReceiverName(value runtime.Value) string {
	for value != nil {
		switch typed := value.(type) {
		case runtime.InterfaceValue:
			value = typed.Underlying
			continue
		case *runtime.InterfaceValue:
			if typed == nil {
				return ""
			}
			value = typed.Underlying
			continue
		}
		break
	}
	switch typed := value.(type) {
	case *runtime.StructInstanceValue:
		if typed == nil || typed.Definition == nil || typed.Definition.Node == nil || typed.Definition.Node.ID == nil {
			return "*struct"
		}
		return typed.Definition.Node.ID.Name
	case runtime.IntegerValue:
		return string(typed.TypeSuffix)
	case *runtime.IntegerValue:
		if typed == nil {
			return "*int"
		}
		return string(typed.TypeSuffix)
	case runtime.FloatValue:
		return string(typed.TypeSuffix)
	case *runtime.FloatValue:
		if typed == nil {
			return "*float"
		}
		return string(typed.TypeSuffix)
	case runtime.ImplementationNamespaceValue:
		if typed.Name != nil && typed.Name.Name != "" {
			return "impl:" + typed.Name.Name
		}
		return "impl"
	case *runtime.ImplementationNamespaceValue:
		if typed == nil {
			return "*impl"
		}
		if typed.Name != nil && typed.Name.Name != "" {
			return "impl:" + typed.Name.Name
		}
		return "impl"
	case runtime.TypeRefValue:
		return typed.TypeName
	case *runtime.TypeRefValue:
		if typed == nil {
			return "*type"
		}
		return typed.TypeName
	case runtime.StringValue:
		return "String"
	case *runtime.StringValue:
		return "String"
	case runtime.BoolValue:
		return "bool"
	case *runtime.BoolValue:
		return "bool"
	case runtime.NilValue, *runtime.NilValue:
		return "nil"
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprintf("%T", value)
	}
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
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	value, err := env.Get(name)
	if err != nil && env != rt.interp.GlobalEnvironment() && rt.globalLookupFallback() {
		if fallback := rt.interp.GlobalEnvironment(); fallback != nil {
			if alt, altErr := fallback.Get(name); altErr == nil {
				recordGlobalLookupFallback("call_named", name)
				value, err = alt, nil
			}
		}
	}
	if err == nil {
		val, callErr := rt.interp.CallFunctionInWithCallNode(value, args, env, call)
		if callErr != nil && call != nil {
			callErr = rt.interp.AttachRuntimeContext(callErr, call, env)
		}
		return val, callErr
	}
	if dot := strings.Index(name, "."); dot > 0 && dot < len(name)-1 {
		if resolved, ok, resolveErr := rt.resolveQualifiedCallable(name, env); resolveErr != nil {
			return nil, resolveErr
		} else if ok && resolved != nil {
			val, callErr := rt.interp.CallFunctionInWithCallNode(resolved, args, env, call)
			if callErr != nil && call != nil {
				callErr = rt.interp.AttachRuntimeContext(callErr, call, env)
			}
			return val, callErr
		}
		head := name[:dot]
		tail := name[dot+1:]
		receiver, recvErr := env.Get(head)
		if recvErr != nil && rt.globalLookupFallback() {
			if fallback := rt.interp.GlobalEnvironment(); fallback != nil && fallback != env {
				if alt, altErr := fallback.Get(head); altErr == nil {
					recordGlobalLookupFallback("call_named_head", head)
					receiver, recvErr = alt, nil
				}
			}
		}
		if recvErr != nil {
			if def, ok := env.StructDefinition(head); ok {
				receiver = def
				recvErr = nil
			} else if rt.globalLookupFallback() {
				if fallback := rt.interp.GlobalEnvironment(); fallback != nil && fallback != env {
					if def, ok := fallback.StructDefinition(head); ok {
						recordGlobalLookupFallback("call_named_head_struct", head)
						receiver = def
						recvErr = nil
					}
				}
			}
		}
		if recvErr != nil {
			receiver = runtime.TypeRefValue{TypeName: head}
		}
		if receiver == nil {
			receiver = runtime.TypeRefValue{TypeName: head}
		}
		member := runtime.StringValue{Val: tail}
		candidate, err := MemberGetPreferMethods(rt, receiver, member)
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
		if prev, swapped := SwapEnvIfNeeded(rt, taskEnv); swapped {
			defer RestoreEnvIfNeeded(rt, prev, swapped)
		}
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
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		coerced, ok := matchTypeWithoutInterpreter(typeExpr, value)
		if !ok {
			return nil, fmt.Errorf("cannot cast value to requested type")
		}
		if coerced == nil {
			return runtime.NilValue{}, nil
		}
		return coerced, nil
	}
	return rt.interp.CastValueToType(typeExpr, value)
}

// MatchType checks whether a value matches a type expression and returns the coerced value when it does.
func MatchType(rt *Runtime, typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool, error) {
	if rt == nil {
		return nil, false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		coerced, ok := matchTypeWithoutInterpreter(typeExpr, value)
		if !ok {
			return nil, false, nil
		}
		if coerced == nil {
			coerced = runtime.NilValue{}
		}
		return coerced, true, nil
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

// TypeExpressionFromValue exposes runtime type expression inference for compiler helpers.
func TypeExpressionFromValue(rt *Runtime, value runtime.Value) (ast.TypeExpression, error) {
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		return staticTypeExpressionFromValue(value), nil
	}
	return rt.interp.TypeExpressionFromValue(value), nil
}

// ExpandTypeAliases expands type aliases using the interpreter alias table.
func ExpandTypeAliases(rt *Runtime, expr ast.TypeExpression) (ast.TypeExpression, error) {
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		return expr, nil
	}
	return rt.interp.ExpandTypeAliases(expr), nil
}

// EnsureTypeSatisfiesInterface checks interface constraints using the interpreter.
func EnsureTypeSatisfiesInterface(rt *Runtime, subject ast.TypeExpression, iface ast.TypeExpression, context string) error {
	if rt == nil {
		return fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		// Static no-bootstrap mode cannot enforce dynamic interface constraints at runtime.
		return nil
	}
	return rt.interp.EnsureTypeSatisfiesInterface(subject, iface, context)
}

// IsKnownConstraintTypeName reports if a type name is known for constraint enforcement.
func IsKnownConstraintTypeName(rt *Runtime, name string) bool {
	if rt == nil || rt.interp == nil {
		return false
	}
	return rt.interp.IsKnownConstraintTypeName(name)
}
