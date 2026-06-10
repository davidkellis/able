package runtime

// CompiledThunk executes a compiled Able body using the provided environment.
// It lives in the shared value layer because both generated code and the
// interpreter attach these thunks to runtime function values.
type CompiledThunk func(env *Environment, args []Value) (Value, error)
