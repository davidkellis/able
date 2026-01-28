package runtime

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"able/interpreter-go/pkg/ast"
)

// Kind identifies the runtime value category.
type Kind int

const (
	KindString Kind = iota
	KindBool
	KindChar
	KindNil
	KindVoid
	KindInteger
	KindFloat
	KindArray
	KindHashMap
	KindHasher
	KindFunction
	KindNativeFunction
	KindFunctionOverload
	KindStructDefinition
	KindTypeRef
	KindStructInstance
	KindInterfaceDefinition
	KindInterfaceValue
	KindUnionDefinition
	KindPackage
	KindDynPackage
	KindDynRef
	KindError
	KindHostHandle
	KindBoundMethod
	KindNativeBoundMethod
	KindImplementationNamespace
	KindFuture
	KindIterator
	KindIteratorEnd
	KindPartialFunction
)

func (k Kind) String() string {
	switch k {
	case KindString:
		return "String"
	case KindBool:
		return "bool"
	case KindChar:
		return "char"
	case KindNil:
		return "nil"
	case KindVoid:
		return "void"
	case KindInteger:
		return "integer"
	case KindFloat:
		return "float"
	case KindArray:
		return "array"
	case KindHashMap:
		return "hash_map"
	case KindHasher:
		return "hasher"
	case KindFunction:
		return "function"
	case KindNativeFunction:
		return "native_function"
	case KindFunctionOverload:
		return "function_overload"
	case KindStructDefinition:
		return "struct_def"
	case KindTypeRef:
		return "type_ref"
	case KindStructInstance:
		return "struct_instance"
	case KindInterfaceDefinition:
		return "interface_def"
	case KindInterfaceValue:
		return "interface_value"
	case KindUnionDefinition:
		return "union_def"
	case KindPackage:
		return "package"
	case KindDynPackage:
		return "dyn_package"
	case KindDynRef:
		return "dyn_ref"
	case KindError:
		return "error"
	case KindHostHandle:
		return "host_handle"
	case KindBoundMethod:
		return "bound_method"
	case KindNativeBoundMethod:
		return "native_bound_method"
	case KindImplementationNamespace:
		return "impl_namespace"
	case KindFuture:
		return "future"
	case KindIterator:
		return "iterator"
	case KindIteratorEnd:
		return "iterator_end"
	case KindPartialFunction:
		return "partial_function"
	default:
		return fmt.Sprintf("unknown_kind_%d", int(k))
	}
}

// Value is the shared behaviour for all runtime values.
type Value interface {
	Kind() Kind
}

// TypeRefValue represents a type reference bound for generic static calls.
type TypeRefValue struct {
	TypeName string
	TypeArgs []ast.TypeExpression
}

func (v TypeRefValue) Kind() Kind { return KindTypeRef }

//-----------------------------------------------------------------------------
// Scalars
//-----------------------------------------------------------------------------

type StringValue struct {
	Val string
}

func (v StringValue) Kind() Kind { return KindString }

type BoolValue struct {
	Val bool
}

func (v BoolValue) Kind() Kind { return KindBool }

type CharValue struct {
	Val rune
}

func (v CharValue) Kind() Kind { return KindChar }

type NilValue struct{}

func (NilValue) Kind() Kind { return KindNil }

type VoidValue struct{}

func (VoidValue) Kind() Kind { return KindVoid }

// Integer sub-types mirror the spec’s suffix set.
type IntegerType string

const (
	IntegerI8   IntegerType = "i8"
	IntegerI16  IntegerType = "i16"
	IntegerI32  IntegerType = "i32"
	IntegerI64  IntegerType = "i64"
	IntegerI128 IntegerType = "i128"
	IntegerU8   IntegerType = "u8"
	IntegerU16  IntegerType = "u16"
	IntegerU32  IntegerType = "u32"
	IntegerU64  IntegerType = "u64"
	IntegerU128 IntegerType = "u128"
)

type IntegerValue struct {
	Val        *big.Int
	TypeSuffix IntegerType
}

func (v IntegerValue) Kind() Kind { return KindInteger }

// Float sub-types.
type FloatType string

const (
	FloatF32 FloatType = "f32"
	FloatF64 FloatType = "f64"
)

type FloatValue struct {
	Val        float64
	TypeSuffix FloatType
}

func (v FloatValue) Kind() Kind { return KindFloat }

//-----------------------------------------------------------------------------
// Collections and ranges
//-----------------------------------------------------------------------------

type ArrayValue struct {
	Elements []Value
	Handle   int64
}

func (v *ArrayValue) Kind() Kind { return KindArray }

type HashMapEntry struct {
	Key   Value
	Value Value
	Hash  uint64
}

type HashMapValue struct {
	Entries []HashMapEntry
}

func (v *HashMapValue) Kind() Kind { return KindHashMap }

// HostHandleValue carries opaque host handles across extern boundaries.
type HostHandleValue struct {
	HandleType string
	Value      any
}

func (v *HostHandleValue) Kind() Kind { return KindHostHandle }

type HasherValue struct {
	state uint64
}

func (v *HasherValue) Kind() Kind { return KindHasher }

// IteratorValue represents a lazily evaluated iterator produced by generator literals.
type IteratorValue struct {
	mu     sync.Mutex
	next   func() (Value, bool, error)
	closer func()
	closed bool
}

// NewIteratorValue constructs an iterator with the provided driver function.
func NewIteratorValue(step func() (Value, bool, error), finalize func()) *IteratorValue {
	if step == nil {
		step = func() (Value, bool, error) { return IteratorEnd, true, nil }
	}
	return &IteratorValue{next: step, closer: finalize}
}

func (v *IteratorValue) Kind() Kind { return KindIterator }

// Next advances the iterator. The bool result reports whether iteration has completed.
func (v *IteratorValue) Next() (Value, bool, error) {
	if v == nil {
		return IteratorEnd, true, nil
	}
	v.mu.Lock()
	if v.closed {
		v.mu.Unlock()
		return IteratorEnd, true, nil
	}
	step := v.next
	v.mu.Unlock()
	if step == nil {
		return IteratorEnd, true, nil
	}
	return step()
}

// Close releases any resources held by the iterator.
func (v *IteratorValue) Close() {
	if v == nil {
		return
	}
	v.mu.Lock()
	if v.closed {
		v.mu.Unlock()
		return
	}
	v.closed = true
	closer := v.closer
	v.mu.Unlock()
	if closer != nil {
		closer()
	}
}

// IteratorEndValue is a sentinel returned once an iterator is exhausted.
type IteratorEndValue struct{}

func (IteratorEndValue) Kind() Kind { return KindIteratorEnd }

// IteratorEnd is the singleton sentinel shared by all iterators.
var IteratorEnd = IteratorEndValue{}

//-----------------------------------------------------------------------------
// Functions & closures
//-----------------------------------------------------------------------------

type FunctionValue struct {
	Declaration    ast.Node // LambdaExpression or FunctionDefinition
	Closure        *Environment
	MethodPriority float64
	TypeQualified  bool
	MethodSet      *MethodSet
	Bytecode       any // interpreter-specific compiled program (e.g., *bytecodeProgram)
}

func (v *FunctionValue) Kind() Kind { return KindFunction }

type MethodSet struct {
	TargetType    ast.TypeExpression
	GenericParams []*ast.GenericParameter
	WhereClause   []*ast.WhereClauseConstraint
}

// FunctionOverloadValue aggregates multiple function declarations under a single name.
type FunctionOverloadValue struct {
	Overloads []*FunctionValue
}

func (v *FunctionOverloadValue) Kind() Kind { return KindFunctionOverload }

// NativeCallContext provides hooks for native functions. Fields will be
// populated as interpreter functionality grows.
type NativeCallContext struct {
	Env   *Environment
	State any
}

type NativeFunc func(*NativeCallContext, []Value) (Value, error)

type NativeFunctionValue struct {
	Name  string
	Arity int
	Impl  NativeFunc
}

func (v NativeFunctionValue) Kind() Kind { return KindNativeFunction }

// Bound methods capture `self` and a callable.
type BoundMethodValue struct {
	Receiver Value
	Method   Value
}

func (v BoundMethodValue) Kind() Kind { return KindBoundMethod }

type NativeBoundMethodValue struct {
	Receiver Value
	Method   NativeFunctionValue
}

func (v NativeBoundMethodValue) Kind() Kind { return KindNativeBoundMethod }

type PartialFunctionValue struct {
	Target    Value
	BoundArgs []Value
	Call      *ast.FunctionCall
}

func (v PartialFunctionValue) Kind() Kind { return KindPartialFunction }

// IsFunctionLike reports whether the value is a function or overload set.
func IsFunctionLike(v Value) bool {
	switch v.(type) {
	case *FunctionValue, *FunctionOverloadValue:
		return true
	default:
		return false
	}
}

// MergeFunctionValues collapses two function-like values into an overload set.
// Returns (nil, false) when either input is not function-like.
func MergeFunctionValues(existing, incoming Value) (Value, bool) {
	if !IsFunctionLike(existing) || !IsFunctionLike(incoming) {
		return nil, false
	}
	all := make([]*FunctionValue, 0, 2)
	all = append(all, FlattenFunctionOverloads(existing)...)
	all = append(all, FlattenFunctionOverloads(incoming)...)
	return &FunctionOverloadValue{Overloads: all}, true
}

// FlattenFunctionOverloads returns the concrete functions contained in a function-like value.
func FlattenFunctionOverloads(v Value) []*FunctionValue {
	switch fn := v.(type) {
	case *FunctionValue:
		if fn == nil {
			return nil
		}
		return []*FunctionValue{fn}
	case *FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) == 0 {
			return nil
		}
		out := make([]*FunctionValue, 0, len(fn.Overloads))
		for _, f := range fn.Overloads {
			if f != nil {
				out = append(out, f)
			}
		}
		return out
	default:
		return nil
	}
}

//-----------------------------------------------------------------------------
// Structs, unions, interfaces
//-----------------------------------------------------------------------------

type StructDefinitionValue struct {
	Node *ast.StructDefinition
}

func (v StructDefinitionValue) Kind() Kind { return KindStructDefinition }

type StructInstanceValue struct {
	Definition    *StructDefinitionValue
	Fields        map[string]Value
	Positional    []Value
	TypeArguments []ast.TypeExpression
}

func (v *StructInstanceValue) Kind() Kind { return KindStructInstance }

type UnionDefinitionValue struct {
	Node *ast.UnionDefinition
}

func (v UnionDefinitionValue) Kind() Kind { return KindUnionDefinition }

type InterfaceDefinitionValue struct {
	Node *ast.InterfaceDefinition
	Env  *Environment
}

func (v InterfaceDefinitionValue) Kind() Kind { return KindInterfaceDefinition }

type InterfaceValue struct {
	Interface     *InterfaceDefinitionValue
	Underlying    Value
	Methods       map[string]Value
	InterfaceArgs []ast.TypeExpression
}

func (v InterfaceValue) Kind() Kind { return KindInterfaceValue }

type ImplementationNamespaceValue struct {
	Name          *ast.Identifier
	InterfaceName *ast.Identifier
	TargetType    ast.TypeExpression
	Methods       map[string]Value
}

func (v ImplementationNamespaceValue) Kind() Kind { return KindImplementationNamespace }

//-----------------------------------------------------------------------------
// Packages & errors
//-----------------------------------------------------------------------------

type PackageValue struct {
	Name      string
	NamePath  []string
	IsPrivate bool
	Public    map[string]Value
}

func (v PackageValue) Kind() Kind { return KindPackage }

type DynPackageValue struct {
	Name      string
	NamePath  []string
	IsPrivate bool
}

func (v DynPackageValue) Kind() Kind { return KindDynPackage }

type DynRefValue struct {
	Package string
	Name    string
}

func (v DynRefValue) Kind() Kind { return KindDynRef }

type ErrorValue struct {
	TypeName *ast.Identifier
	Payload  map[string]Value
	Message  string
}

func (v ErrorValue) Kind() Kind { return KindError }

//-----------------------------------------------------------------------------
// Concurrency handles (Future/spawn)
//-----------------------------------------------------------------------------

type FutureStatus int

const (
	FuturePending FutureStatus = iota
	FutureResolved
	FutureCancelled
	FutureFailed
)

type FutureValue struct {
	mu              sync.Mutex
	status          FutureStatus
	result          Value
	err             Value // usually ErrorValue wrapping FutureError
	cancelRequested bool
	ctx             context.Context
	cancel          context.CancelFunc
	started         bool
	done            *sync.Cond
	awaiters        []func()
}

func NewFuture() *FutureValue {
	return NewFutureWithContext(context.Background(), nil)
}

func NewFutureWithContext(ctx context.Context, cancel context.CancelFunc) *FutureValue {
	if ctx == nil {
		ctx = context.Background()
	}
	h := &FutureValue{
		ctx:    ctx,
		cancel: cancel,
	}
	h.done = sync.NewCond(&h.mu)
	return h
}

func (v *FutureValue) Kind() Kind { return KindFuture }

func (v *FutureValue) Context() context.Context {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.ctx
}

func (v *FutureValue) CancelFunc() context.CancelFunc {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.cancel
}

func (v *FutureValue) MarkStarted() {
	v.mu.Lock()
	v.started = true
	v.mu.Unlock()
}

func (v *FutureValue) Started() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.started
}

func (v *FutureValue) Status() FutureStatus {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.status
}

func (v *FutureValue) Snapshot() (Value, Value, FutureStatus) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.result, v.err, v.status
}

func (v *FutureValue) AddAwaiter(cb func()) {
	if cb == nil {
		return
	}
	v.mu.Lock()
	if v.status != FuturePending {
		v.mu.Unlock()
		cb()
		return
	}
	v.awaiters = append(v.awaiters, cb)
	v.mu.Unlock()
}

func (v *FutureValue) Await() (Value, Value, FutureStatus) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for v.status == FuturePending {
		v.done.Wait()
	}
	return v.result, v.err, v.status
}

func (v *FutureValue) Resolve(val Value) {
	var awaiters []func()
	v.mu.Lock()
	if v.status == FuturePending {
		v.status = FutureResolved
		v.result = val
		awaiters = v.awaiters
		v.awaiters = nil
		v.done.Broadcast()
	}
	v.mu.Unlock()
	for _, cb := range awaiters {
		if cb != nil {
			cb()
		}
	}
}

func (v *FutureValue) Fail(err Value) {
	var awaiters []func()
	v.mu.Lock()
	if v.status == FuturePending {
		v.status = FutureFailed
		v.err = err
		awaiters = v.awaiters
		v.awaiters = nil
		v.done.Broadcast()
	}
	v.mu.Unlock()
	for _, cb := range awaiters {
		if cb != nil {
			cb()
		}
	}
}

func (v *FutureValue) Cancel(err Value) {
	var awaiters []func()
	v.mu.Lock()
	if v.status == FuturePending {
		v.status = FutureCancelled
		v.err = err
		v.cancelRequested = true
		awaiters = v.awaiters
		v.awaiters = nil
		v.done.Broadcast()
	}
	v.mu.Unlock()
	for _, cb := range awaiters {
		if cb != nil {
			cb()
		}
	}
}

func (v *FutureValue) RequestCancel() {
	v.mu.Lock()
	already := v.cancelRequested
	cancel := v.cancel
	v.cancelRequested = true
	v.mu.Unlock()
	if !already && cancel != nil {
		cancel()
	}
}

func (v *FutureValue) CancelRequested() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.cancelRequested
}

//-----------------------------------------------------------------------------
// Utility helpers
//-----------------------------------------------------------------------------

// CloneBigInt copies the provided big.Int pointer, tolerating nil.
func CloneBigInt(src *big.Int) *big.Int {
	if src == nil {
		return nil
	}
	return new(big.Int).Set(src)
}

// TimestampValue is a helper for debugging.
type TimestampValue struct {
	Time time.Time
}

func (TimestampValue) Kind() Kind { return KindNativeFunction }
