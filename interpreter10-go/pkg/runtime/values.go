package runtime

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"able/interpreter10-go/pkg/ast"
)

// Kind identifies the runtime value category.
type Kind int

const (
	KindString Kind = iota
	KindBool
	KindChar
	KindNil
	KindInteger
	KindFloat
	KindArray
	KindRange
	KindFunction
	KindNativeFunction
	KindStructDefinition
	KindStructInstance
	KindInterfaceDefinition
	KindInterfaceValue
	KindUnionDefinition
	KindPackage
	KindDynPackage
	KindDynRef
	KindError
	KindBoundMethod
	KindNativeBoundMethod
	KindImplementationNamespace
	KindProcHandle
	KindFuture
)

func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindChar:
		return "char"
	case KindNil:
		return "nil"
	case KindInteger:
		return "integer"
	case KindFloat:
		return "float"
	case KindArray:
		return "array"
	case KindRange:
		return "range"
	case KindFunction:
		return "function"
	case KindNativeFunction:
		return "native_function"
	case KindStructDefinition:
		return "struct_def"
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
	case KindBoundMethod:
		return "bound_method"
	case KindNativeBoundMethod:
		return "native_bound_method"
	case KindImplementationNamespace:
		return "impl_namespace"
	case KindProcHandle:
		return "proc_handle"
	case KindFuture:
		return "future"
	default:
		return fmt.Sprintf("unknown_kind_%d", int(k))
	}
}

// Value is the shared behaviour for all runtime values.
type Value interface {
	Kind() Kind
}

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
}

func (v *ArrayValue) Kind() Kind { return KindArray }

// RangeValue matches Able’s numeric range semantics.
type RangeValue struct {
	Start     Value
	End       Value
	Inclusive bool
}

func (v RangeValue) Kind() Kind { return KindRange }

//-----------------------------------------------------------------------------
// Functions & closures
//-----------------------------------------------------------------------------

type FunctionValue struct {
	Declaration ast.Node // LambdaExpression or FunctionDefinition
	Closure     *Environment
}

func (v *FunctionValue) Kind() Kind { return KindFunction }

// NativeCallContext provides hooks for native functions. Fields will be
// populated as interpreter functionality grows.
type NativeCallContext struct {
	Env *Environment
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
	Method   *FunctionValue
}

func (v BoundMethodValue) Kind() Kind { return KindBoundMethod }

type NativeBoundMethodValue struct {
	Receiver Value
	Method   NativeFunctionValue
}

func (v NativeBoundMethodValue) Kind() Kind { return KindNativeBoundMethod }

//-----------------------------------------------------------------------------
// Structs, unions, interfaces
//-----------------------------------------------------------------------------

type StructDefinitionValue struct {
	Node *ast.StructDefinition
}

func (v StructDefinitionValue) Kind() Kind { return KindStructDefinition }

type StructInstanceValue struct {
	Definition *StructDefinitionValue
	Fields     map[string]Value
	Positional []Value
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
	Interface  *InterfaceDefinitionValue
	Underlying Value
	Methods    map[string]*FunctionValue
}

func (v InterfaceValue) Kind() Kind { return KindInterfaceValue }

type ImplementationNamespaceValue struct {
	Name          *ast.Identifier
	InterfaceName *ast.Identifier
	TargetType    ast.TypeExpression
	Definitions   []*FunctionValue
}

func (v ImplementationNamespaceValue) Kind() Kind { return KindImplementationNamespace }

//-----------------------------------------------------------------------------
// Packages & errors
//-----------------------------------------------------------------------------

type PackageValue struct {
	NamePath []string
	Public   map[string]Value
}

func (v PackageValue) Kind() Kind { return KindPackage }

type DynPackageValue struct {
	NamePath []string
	Name     string
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
// Concurrency handles (proc/spawn)
//-----------------------------------------------------------------------------

type ProcStatus int

const (
	ProcPending ProcStatus = iota
	ProcResolved
	ProcCancelled
	ProcFailed
)

type ProcHandleValue struct {
	mu              sync.Mutex
	status          ProcStatus
	result          Value
	err             Value // usually ErrorValue wrapping ProcError
	cancelRequested bool
	done            *sync.Cond
}

func NewProcHandle() *ProcHandleValue {
	h := &ProcHandleValue{}
	h.done = sync.NewCond(&h.mu)
	return h
}

func (v *ProcHandleValue) Kind() Kind { return KindProcHandle }

func (v *ProcHandleValue) Status() ProcStatus {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.status
}

func (v *ProcHandleValue) Await() (Value, Value, ProcStatus) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for v.status == ProcPending {
		v.done.Wait()
	}
	return v.result, v.err, v.status
}

func (v *ProcHandleValue) Resolve(val Value) {
	v.mu.Lock()
	if v.status == ProcPending {
		v.status = ProcResolved
		v.result = val
		v.done.Broadcast()
	}
	v.mu.Unlock()
}

func (v *ProcHandleValue) Fail(err Value) {
	v.mu.Lock()
	if v.status == ProcPending {
		v.status = ProcFailed
		v.err = err
		v.done.Broadcast()
	}
	v.mu.Unlock()
}

func (v *ProcHandleValue) Cancel(err Value) {
	v.mu.Lock()
	if v.status == ProcPending {
		v.status = ProcCancelled
		v.err = err
		v.cancelRequested = true
		v.done.Broadcast()
	}
	v.mu.Unlock()
}

func (v *ProcHandleValue) RequestCancel() {
	v.mu.Lock()
	v.cancelRequested = true
	v.mu.Unlock()
}

func (v *ProcHandleValue) CancelRequested() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.cancelRequested
}

// FutureValue memoizes a deferred computation triggered on first observe.
type FutureValue struct {
	once   sync.Once
	done   chan struct{}
	value  Value
	err    Value
	runner func() (Value, Value)
}

func NewFutureValue(runner func() (Value, Value)) *FutureValue {
	return &FutureValue{
		done:   make(chan struct{}),
		runner: runner,
	}
}

func (v *FutureValue) Kind() Kind { return KindFuture }

func (v *FutureValue) force() (Value, Value) {
	v.once.Do(func() {
		if v.runner == nil {
			close(v.done)
			return
		}
		val, err := v.runner()
		v.value = val
		v.err = err
		close(v.done)
	})
	<-v.done
	return v.value, v.err
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
