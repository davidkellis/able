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
	interp                      Interpreter
	mu                          sync.RWMutex
	originals                   map[string]runtime.Value
	structs                     map[structDefinitionCacheKey]*runtime.StructDefinitionValue
	qualifiedStructs            map[string]*runtime.StructDefinitionValue
	executorKind                string
	env                         atomic.Pointer[runtime.Environment]
	envByGID                    sync.Map
	callFrames                  []*ast.FunctionCall
	callFramesByGID             sync.Map
	concurrent                  int32 // atomic: 0 = single goroutine (fast path), 1 = concurrent
	resolver                    QualifiedCallableResolver
	globalLookupFallbackEnabled bool
}

type QualifiedCallableResolver func(name string, env *runtime.Environment) (runtime.Value, bool, error)

func New(interp Interpreter) *Runtime {
	return &Runtime{
		interp:                      interp,
		originals:                   make(map[string]runtime.Value),
		structs:                     make(map[structDefinitionCacheKey]*runtime.StructDefinitionValue),
		qualifiedStructs:            make(map[string]*runtime.StructDefinitionValue),
		globalLookupFallbackEnabled: true,
	}
}

// RegisterQualifiedStructDefinition records a package-owned nominal struct for
// standalone compiled binaries, which do not create an interpreter registry.
func (r *Runtime) RegisterQualifiedStructDefinition(pkgName, name string, def *runtime.StructDefinitionValue) {
	if r == nil || def == nil {
		return
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return
	}
	qualifiedName := pkgName + "." + name
	r.mu.Lock()
	r.qualifiedStructs[qualifiedName] = def
	r.mu.Unlock()
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
	if r == nil {
		return "serial"
	}
	r.mu.RLock()
	kind := r.executorKind
	interp := r.interp
	r.mu.RUnlock()
	if kind != "" {
		return kind
	}
	if interp == nil {
		return "serial"
	}
	return interp.ExecutorKind()
}

func (r *Runtime) SetExecutorKind(kind string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.executorKind = kind
	r.mu.Unlock()
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
	r.env.Store(env)
	if r.isConcurrent() {
		r.envByGID.Store(currentGID(), env)
	}
}

func (r *Runtime) Env() *runtime.Environment {
	if r == nil {
		return nil
	}
	if r.isConcurrent() {
		if env := r.goroutineEnv(currentGID()); env != nil {
			return env
		}
	}
	return r.env.Load()
}

func (r *Runtime) SwapEnv(env *runtime.Environment) *runtime.Environment {
	if r == nil {
		return nil
	}
	if !r.isConcurrent() {
		prev := r.env.Load()
		if env != nil {
			r.env.Store(env)
		}
		return prev
	}
	gid := currentGID()
	prev := r.goroutineEnv(gid)
	if prev == nil {
		prev = r.env.Load()
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
		if env := r.goroutineEnv(currentGID()); env != nil {
			return env
		}
	}
	env := r.env.Load()
	if env == nil && r.interp != nil {
		env = r.interp.GlobalEnvironment()
	}
	return env
}

func (r *Runtime) goroutineEnv(gid uint64) *runtime.Environment {
	if r == nil {
		return nil
	}
	if env, ok := r.envByGID.Load(gid); ok {
		if typed, ok := env.(*runtime.Environment); ok && typed != nil {
			return typed
		}
	}
	return nil
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
	args = materializeBoundaryValues(args)
	value, err := r.interp.CallFunctionIn(orig, args, env)
	return materializeBoundaryValue(value), err
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
	args = materializeBoundaryValues(args)
	result, err := r.interp.CallFunctionIn(value, args, env)
	return materializeBoundaryValue(result), err
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
		return materializeBoundaryValue(value), nil
	}
	if rt.interp != nil && env != rt.interp.GlobalEnvironment() && rt.globalLookupFallback() {
		if fallback := rt.interp.GlobalEnvironment(); fallback != nil {
			if value, err := fallback.Get(name); err == nil {
				recordGlobalLookupFallback("get", name)
				return materializeBoundaryValue(value), nil
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
	value = materializeBoundaryValue(value)
	if env.AssignExisting(name, value) {
		return nil
	}
	env.Define(name, value)
	return nil
}

func Index(rt *Runtime, obj runtime.Value, idx runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	value, err := rt.interp.IndexGet(materializeBoundaryValue(obj), materializeBoundaryValue(idx), nil)
	return materializeBoundaryValue(value), err
}

func IndexAssign(rt *Runtime, obj runtime.Value, idx runtime.Value, value runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	result, err := rt.interp.IndexAssign(
		materializeBoundaryValue(obj),
		materializeBoundaryValue(idx),
		materializeBoundaryValue(value),
		nil,
	)
	return materializeBoundaryValue(result), err
}

func HashMapHashValue(rt *Runtime, val runtime.Value) (uint64, error) {
	if rt == nil {
		return 0, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		if hash, ok, err := primitiveHashMapHash(val); ok || err != nil {
			return hash, err
		}
		return 0, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.HashMapHashValue(val)
}

func HashMapKeysEqual(rt *Runtime, a runtime.Value, b runtime.Value) (bool, error) {
	if rt == nil {
		return false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		if equal, ok := primitiveHashMapKeyEqual(a, b); ok {
			return equal, nil
		}
		return false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	return rt.interp.HashMapKeysEqual(a, b)
}

func MemberAssign(rt *Runtime, obj runtime.Value, member runtime.Value, value runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	result, err := rt.interp.MemberAssign(
		materializeBoundaryValue(obj),
		materializeBoundaryValue(member),
		materializeBoundaryValue(value),
		nil,
	)
	return materializeBoundaryValue(result), err
}

func MemberGet(rt *Runtime, obj runtime.Value, member runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	value, err := rt.interp.MemberGet(materializeBoundaryValue(obj), materializeBoundaryValue(member), env)
	return materializeBoundaryValue(value), err
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
	value, err := rt.interp.MemberGetPreferMethods(materializeBoundaryValue(obj), materializeBoundaryValue(member), env)
	return materializeBoundaryValue(value), err
}

// CallStaticGenericUnionMember preserves a checked static receiver type for
// the interpreter-backed fallback of generic named-union dispatch. Standalone
// compiled binaries use their generated method table first; all other calls
// continue through generated fast dispatch.
func CallStaticGenericUnionMember(rt *Runtime, obj runtime.Value, memberName string, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, bool, error) {
	if rt == nil || rt.interp == nil {
		// Standalone compiled binaries deliberately omit the interpreter
		// bootstrap. Let their generated native-method table handle the call;
		// the static interpreter path below is only needed when that table is
		// not the active dispatch mechanism.
		return nil, false, nil
	}
	materializedArgs := make([]runtime.Value, len(args))
	for index, arg := range args {
		materializedArgs[index] = materializeBoundaryValue(arg)
	}
	originals := rt.originalCallablesNamed(memberName)
	var value runtime.Value
	var handled bool
	var err error
	if len(originals) > 0 {
		value, handled, err = rt.interp.CallStaticGenericUnionMemberFromCandidates(
			originals,
			materializeBoundaryValue(obj),
			memberName,
			materializedArgs,
			call,
			rt.currentEnv(),
		)
	} else {
		value, handled, err = rt.interp.CallStaticGenericUnionMember(
			materializeBoundaryValue(obj),
			memberName,
			materializedArgs,
			call,
			rt.currentEnv(),
		)
	}
	return materializeBoundaryValue(value), handled, err
}

func (r *Runtime) originalCallablesNamed(name string) []runtime.Value {
	if r == nil || name == "" {
		return nil
	}
	suffix := "." + name
	r.mu.RLock()
	keys := make([]string, 0, len(r.originals))
	for qualified := range r.originals {
		if strings.HasSuffix(qualified, suffix) {
			keys = append(keys, qualified)
		}
	}
	sort.Strings(keys)
	values := make([]runtime.Value, 0, len(keys))
	for _, key := range keys {
		if value := r.originals[key]; value != nil {
			values = append(values, value)
		}
	}
	r.mu.RUnlock()
	return values
}

// RegisterStaticCallReceiverType associates a generated call node with the
// checked receiver type that its source call had before Go lowering.
func RegisterStaticCallReceiverType(rt *Runtime, call *ast.FunctionCall, receiverType ast.TypeExpression) {
	if rt == nil || rt.interp == nil {
		return
	}
	rt.interp.RegisterStaticCallReceiverType(call, receiverType)
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
	val, err := rt.interp.CallFunctionInWithCallNode(materializeBoundaryValue(fn), materializeBoundaryValues(args), env, call)
	if err != nil && call != nil {
		err = attachRuntimeContext(rt, err, call, env)
	}
	return materializeBoundaryValue(val), err
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
	args = materializeBoundaryValues(args)
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
			callErr = attachRuntimeContext(rt, callErr, call, env)
		}
		return materializeBoundaryValue(val), callErr
	}
	if dot := strings.Index(name, "."); dot > 0 && dot < len(name)-1 {
		if resolved, ok, resolveErr := rt.resolveQualifiedCallable(name, env); resolveErr != nil {
			return nil, resolveErr
		} else if ok && resolved != nil {
			val, callErr := rt.interp.CallFunctionInWithCallNode(resolved, args, env, call)
			if callErr != nil && call != nil {
				callErr = attachRuntimeContext(rt, callErr, call, env)
			}
			return materializeBoundaryValue(val), callErr
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
			callErr = attachRuntimeContext(rt, callErr, call, env)
		}
		return materializeBoundaryValue(val), callErr
	}
	return nil, err
}

func Stringify(rt *Runtime, value runtime.Value) (string, error) {
	if rt == nil {
		return "", fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		if rendered, ok := stringifyPrimitive(value); ok {
			return rendered, nil
		}
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
		return runtime.ErrorValue{Message: fallbackValueToString(value), Payload: payload}
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
