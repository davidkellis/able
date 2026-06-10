package runtime

// RuntimeValueMaterializer lets a compiled native carrier defer construction
// of its semantic runtime representation until a dynamic consumer needs it.
type RuntimeValueMaterializer interface {
	Value
	MaterializeRuntimeValue() Value
}

// NativeAwaitableValue is the compiled carrier for the language-kernel
// Awaitable protocol. Dynamic consumers continue to see the materialized
// runtime value through RuntimeValueMaterializer.
type NativeAwaitableValue interface {
	RuntimeValueMaterializer
	NativeAwaitableIsReady(*NativeCallContext) (bool, error)
	NativeAwaitableRegister(*NativeCallContext, Value) (Value, error)
	NativeAwaitableCommit(*NativeCallContext) (Value, error)
	NativeAwaitableIsDefault() bool
}
