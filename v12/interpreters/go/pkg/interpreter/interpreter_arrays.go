package interpreter

import (
	"fmt"
	"math"
	"weak"

	"able/interpreter-go/pkg/runtime"
)

type arrayState = runtime.ArrayState

// pruneArrayHandleTracking drops completed ArrayValue views without retaining
// them. It also keeps the single-view path compact after a shared view expires.
// Callers hold the interpreter array mutex.
func pruneArrayHandleTracking(tracking *arrayHandleTracking) int {
	if tracking == nil {
		return 0
	}
	if tracking.many == nil {
		arr := tracking.single.Value()
		if arr == nil {
			tracking.single = weak.Pointer[runtime.ArrayValue]{}
			return 0
		}
		arr.TrackedAliases = false
		return 1
	}

	live := 0
	var only *runtime.ArrayValue
	for ref := range tracking.many {
		arr := ref.Value()
		if arr == nil {
			delete(tracking.many, ref)
			continue
		}
		live++
		only = arr
	}
	switch live {
	case 0:
		tracking.single = weak.Pointer[runtime.ArrayValue]{}
		tracking.many = nil
	case 1:
		tracking.single = weak.Make(only)
		tracking.many = nil
		only.TrackedAliases = false
	default:
		tracking.single = weak.Pointer[runtime.ArrayValue]{}
		for ref := range tracking.many {
			if arr := ref.Value(); arr != nil {
				arr.TrackedAliases = true
			} else {
				delete(tracking.many, ref)
			}
		}
	}
	return live
}

func (i *Interpreter) trackArrayValue(handle int64, arr *runtime.ArrayValue) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.trackArrayValueLocked(handle, arr)
}

func (i *Interpreter) trackArrayValueLocked(handle int64, arr *runtime.ArrayValue) {
	if arr == nil || handle == 0 {
		return
	}
	if arr.Handle != 0 && arr.Handle != handle {
		i.untrackArrayValueLocked(arr.Handle, arr)
	}
	if i.arraysByHandle == nil {
		i.arraysByHandle = make(map[int64]arrayHandleTracking)
	}
	_ = runtime.ArrayStoreTrackArrayValueLease(arr, handle)
	arr.Handle = handle
	arr.TrackedHandle = handle
	tracking := i.arraysByHandle[handle]
	pruneArrayHandleTracking(&tracking)
	switch {
	case tracking.many != nil:
		arr.TrackedAliases = true
		tracking.many[weak.Make(arr)] = struct{}{}
	case tracking.single.Value() == nil:
		arr.TrackedAliases = false
		tracking.single = weak.Make(arr)
	case tracking.single.Value() == arr:
		arr.TrackedAliases = false
	default:
		previous := tracking.single.Value()
		previous.TrackedAliases = true
		arr.TrackedAliases = true
		tracking.many = map[weak.Pointer[runtime.ArrayValue]]struct{}{
			weak.Make(previous): {},
			weak.Make(arr):      {},
		}
		tracking.single = weak.Pointer[runtime.ArrayValue]{}
	}
	i.arraysByHandle[handle] = tracking
}

func (i *Interpreter) untrackArrayValue(handle int64, arr *runtime.ArrayValue) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.untrackArrayValueLocked(handle, arr)
}

func (i *Interpreter) untrackArrayValueLocked(handle int64, arr *runtime.ArrayValue) {
	if i == nil || arr == nil || handle == 0 || i.arraysByHandle == nil {
		return
	}
	tracking, ok := i.arraysByHandle[handle]
	if !ok {
		return
	}
	pruneArrayHandleTracking(&tracking)
	switch {
	case tracking.single.Value() == arr:
		tracking.single = weak.Pointer[runtime.ArrayValue]{}
		arr.TrackedAliases = false
	case tracking.many != nil:
		delete(tracking.many, weak.Make(arr))
		arr.TrackedAliases = false
	}
	if pruneArrayHandleTracking(&tracking) == 0 {
		if arr.TrackedHandle == handle {
			arr.TrackedHandle = 0
		}
		arr.TrackedAliases = false
		delete(i.arraysByHandle, handle)
		return
	}
	i.arraysByHandle[handle] = tracking
}

func updateArrayElementTypeTokenForWrite(state *arrayState, idx int, value runtime.Value) {
	if state == nil {
		return
	}
	state.Revision++
	if idx == 0 {
		if updateKnownI32FirstElementMetadataForWrite(state, value) {
			return
		}
		token, ok := bytecodeIndexValueTypeToken(value)
		if !ok {
			token, ok = bytecodeArrayElementTypeTokenFromValues(state.Values)
		}
		state.ElementTypeToken = token
		state.ElementTypeTokenKnown = ok
		updateTrackedArrayI32RawCacheForWrite(state, idx, value)
		return
	}
	if state.ElementTypeTokenKnown {
		updateTrackedArrayI32RawCacheForWrite(state, idx, value)
		return
	}
	token, ok := bytecodeArrayElementTypeTokenFromValues(state.Values)
	state.ElementTypeToken = token
	state.ElementTypeTokenKnown = ok
	updateTrackedArrayI32RawCacheForWrite(state, idx, value)
}

func updateKnownI32FirstElementMetadataForWrite(state *arrayState, value runtime.Value) bool {
	if state == nil ||
		!state.ElementTypeTokenKnown ||
		state.ElementTypeToken != bytecodeIndexTypeI32 ||
		len(state.Values) == 0 ||
		len(state.CachedI32Values) != len(state.Values) ||
		len(state.CachedI32ValuesValid) != len(state.Values) {
		return false
	}
	raw, ok := bytecodeDirectSmallI32Value(value)
	if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
		return false
	}
	if !state.CachedI32ValuesValid[0] {
		state.CachedI32ValuesCount++
	}
	state.CachedI32Values[0] = int32(raw)
	state.CachedI32ValuesValid[0] = true
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
	return true
}

func updateArrayElementTypeTokenForLength(state *arrayState) {
	if state == nil {
		return
	}
	token, ok := bytecodeArrayElementTypeTokenFromValues(state.Values)
	state.ElementTypeToken = token
	state.ElementTypeTokenKnown = ok
	if !ok || token != bytecodeIndexTypeI32 {
		clearTrackedArrayI32RawCache(state)
		return
	}
	if !reconcileTrackedArrayI32RawCacheLength(state, -1) {
		return
	}
}

func updateTrackedArrayMetadataForSwap(state *arrayState, first int, second int) {
	if state == nil {
		return
	}
	state.Revision += 2
	if !state.ElementTypeTokenKnown || first == 0 || second == 0 {
		token, ok := bytecodeArrayElementTypeTokenFromValues(state.Values)
		state.ElementTypeToken = token
		state.ElementTypeTokenKnown = ok
		refreshTrackedArrayI32RawCache(state)
		return
	}
	if state.ElementTypeToken == bytecodeIndexTypeI32 {
		swapTrackedArrayI32RawCache(state, first, second)
	}
}

func arrayStateWriteKeepsMaterializedValues(state *arrayState, value runtime.Value) {
	if state == nil || !state.ValuesMaterialized {
		return
	}
	if bytecodeIsRawIntegerCarrier(value) {
		state.ValuesMaterialized = false
		return
	}
	switch value.(type) {
	case bytecodeRawF32SlotValue, bytecodeRawF64SlotValue:
		state.ValuesMaterialized = false
	}
}

func materializeArrayStateValues(state *arrayState) {
	if state == nil || state.ValuesMaterialized {
		return
	}
	for idx, value := range state.Values {
		state.Values[idx] = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	}
	state.ValuesMaterialized = true
}

func syncTrackedArrayValue(arr *runtime.ArrayValue, handle int64, state *arrayState) {
	if arr == nil || state == nil {
		return
	}
	arr.Handle = handle
	arr.TrackedHandle = handle
	arr.State = state
	arr.Elements = state.Values
}

func (i *Interpreter) syncArrayHandleState(handle int64, state *arrayState) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.syncArrayHandleStateLocked(handle, state)
}

func (i *Interpreter) syncArrayHandleStateLocked(handle int64, state *arrayState) {
	if i == nil || state == nil || i.arraysByHandle == nil || handle == 0 {
		return
	}
	tracking, ok := i.arraysByHandle[handle]
	if !ok {
		return
	}
	if tracking.many == nil {
		arr := tracking.single.Value()
		if arr == nil {
			delete(i.arraysByHandle, handle)
			return
		}
		arr.TrackedAliases = false
		syncTrackedArrayValue(arr, handle, state)
		return
	}
	for ref := range tracking.many {
		arr := ref.Value()
		if arr == nil {
			delete(tracking.many, ref)
			continue
		}
		syncTrackedArrayValue(arr, handle, state)
	}
	if pruneArrayHandleTracking(&tracking) == 0 {
		delete(i.arraysByHandle, handle)
		return
	}
	i.arraysByHandle[handle] = tracking
}

func invalidateTrackedArrayValue(arr *runtime.ArrayValue, handle int64) {
	if arr == nil {
		return
	}
	arr.Handle = handle
	arr.TrackedHandle = handle
	arr.State = nil
	arr.Elements = nil
}

func (i *Interpreter) invalidateArrayHandleValueView(handle int64) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.invalidateArrayHandleValueViewLocked(handle)
}

func (i *Interpreter) invalidateArrayHandleValueViewLocked(handle int64) {
	if i == nil || i.arraysByHandle == nil || handle == 0 {
		return
	}
	tracking, ok := i.arraysByHandle[handle]
	if !ok {
		return
	}
	if tracking.many == nil {
		arr := tracking.single.Value()
		if arr == nil {
			delete(i.arraysByHandle, handle)
			return
		}
		arr.TrackedAliases = false
		invalidateTrackedArrayValue(arr, handle)
		return
	}
	for ref := range tracking.many {
		arr := ref.Value()
		if arr == nil {
			delete(tracking.many, ref)
			continue
		}
		invalidateTrackedArrayValue(arr, handle)
	}
	if pruneArrayHandleTracking(&tracking) == 0 {
		delete(i.arraysByHandle, handle)
		return
	}
	i.arraysByHandle[handle] = tracking
}

func (i *Interpreter) syncTrackedArrayState(arr *runtime.ArrayValue, state *arrayState) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.syncTrackedArrayStateLocked(arr, state)
}

func (i *Interpreter) syncTrackedArrayStateLocked(arr *runtime.ArrayValue, state *arrayState) {
	if i == nil || arr == nil || state == nil {
		return
	}
	handle := arr.Handle
	if handle == 0 {
		handle = arr.TrackedHandle
	}
	syncTrackedArrayValue(arr, handle, state)
	if !arr.TrackedAliases || handle == 0 {
		return
	}
	i.syncArrayHandleStateLocked(handle, state)
}

func (i *Interpreter) syncTrackedArrayWrite(arr *runtime.ArrayValue, state *arrayState, idx int, value runtime.Value) {
	if i == nil || arr == nil || state == nil {
		return
	}
	updateArrayElementTypeTokenForWrite(state, idx, value)
	arrayStateWriteKeepsMaterializedValues(state, value)
	i.syncTrackedArrayState(arr, state)
}

func (i *Interpreter) syncArrayHandleWrite(handle int64, state *arrayState, idx int, value runtime.Value) {
	if i == nil || state == nil {
		return
	}
	updateArrayElementTypeTokenForWrite(state, idx, value)
	state.ValuesMaterialized = true
	i.syncArrayHandleState(handle, state)
}

func (i *Interpreter) syncArrayHandleWriteAfterStore(handle int64, idx int, value runtime.Value) {
	if i == nil || handle == 0 {
		return
	}
	if _, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle); err == nil && ok {
		i.invalidateArrayHandleValueView(handle)
		return
	}
	if state, err := runtime.ArrayStoreState(handle); err == nil {
		i.syncArrayHandleWrite(handle, state, idx, value)
	}
}

func (i *Interpreter) syncArrayHandleLength(handle int64, state *arrayState) {
	if i == nil || state == nil {
		return
	}
	updateArrayElementTypeTokenForLength(state)
	i.syncArrayHandleState(handle, state)
}

func (i *Interpreter) syncArrayHandleMetadata(handle int64, state *arrayState) {
	if i == nil || state == nil {
		return
	}
	i.syncArrayHandleState(handle, state)
}

func (i *Interpreter) syncArrayValues(handle int64, state *arrayState) {
	if i == nil {
		return
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.syncArrayValuesLocked(handle, state)
}

func (i *Interpreter) syncArrayValuesLocked(handle int64, state *arrayState) {
	if state == nil {
		return
	}
	materializeArrayStateValues(state)
	state.Revision++
	state.ValuesMaterialized = true
	token, ok := bytecodeArrayElementTypeTokenFromValues(state.Values)
	state.ElementTypeToken = token
	state.ElementTypeTokenKnown = ok
	refreshTrackedArrayI32RawCache(state)
	i.syncArrayHandleStateLocked(handle, state)
}

func (i *Interpreter) ensureArrayBuiltins() {
	if i.arrayReady {
		return
	}
	i.initArrayBuiltins()
}

func (i *Interpreter) arrayStateForHandle(handle int64) (*arrayState, error) {
	return runtime.ArrayStoreState(handle)
}

func ensureArrayCapacity(state *arrayState, minimum int) bool {
	return runtime.ArrayEnsureCapacity(state, minimum)
}

func setArrayLength(state *arrayState, length int) {
	runtime.ArraySetLength(state, length)
}

func (i *Interpreter) ensureArrayState(arr *runtime.ArrayValue, capacityHint int) (*arrayState, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	i.ensureArrayBuiltins()
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	if arr.State != nil && arr.Handle != 0 && arr.TrackedHandle == arr.Handle && capacityHint <= arr.State.Capacity {
		materializeArrayStateValues(arr.State)
		arr.Elements = arr.State.Values
		return arr.State, nil
	}
	for idx, value := range arr.Elements {
		arr.Elements[idx] = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	}
	state, handle, err := runtime.ArrayStoreEnsure(arr, capacityHint)
	if err != nil {
		return nil, err
	}
	arr.State = state
	i.trackArrayValueLocked(handle, arr)
	i.syncArrayValuesLocked(handle, state)
	return state, nil
}

func (i *Interpreter) ensureArrayStateForMetadata(arr *runtime.ArrayValue, capacityHint int) (*arrayState, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	i.ensureArrayBuiltins()
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	if state := arrayCurrentTrackedState(arr); state != nil && capacityHint <= state.Capacity {
		return state, nil
	}
	handle := arrayValueHandle(arr)
	if handle != 0 {
		state, err := runtime.ArrayStoreEnsureHandle(handle, len(arr.Elements), capacityHint)
		if err != nil {
			return nil, err
		}
		arr.State = state
		i.trackArrayValueLocked(handle, arr)
		i.syncTrackedArrayStateLocked(arr, state)
		return state, nil
	}
	state, handle, err := runtime.ArrayStoreEnsure(arr, capacityHint)
	if err != nil {
		return nil, err
	}
	arr.State = state
	i.trackArrayValueLocked(handle, arr)
	i.syncTrackedArrayStateLocked(arr, state)
	return state, nil
}

// ArrayElements exposes array state access for compiled interop.
func (i *Interpreter) ArrayElements(arr *runtime.ArrayValue) ([]runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	state, err := i.ensureArrayState(arr, 0)
	if err != nil {
		return nil, err
	}
	return state.Values, nil
}

func (i *Interpreter) arrayValueFromHandle(handle int64, lengthHint int, capacityHint int) (*runtime.ArrayValue, error) {
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	i.ensureArrayBuiltins()
	arr, state, err := runtime.ArrayStoreValueViewFromHandle(handle, lengthHint, capacityHint)
	if err != nil {
		return nil, err
	}
	i.arrayMu.Lock()
	defer i.arrayMu.Unlock()
	i.trackArrayValueLocked(handle, arr)
	if state != nil {
		arr.State = state
		i.syncArrayValuesLocked(handle, state)
	}
	return arr, nil
}

func (i *Interpreter) newArrayValue(elements []runtime.Value, capacityHint int) *runtime.ArrayValue {
	if capacityHint < len(elements) {
		capacityHint = len(elements)
	}
	arr := &runtime.ArrayValue{Elements: elements}
	if _, err := i.ensureArrayState(arr, capacityHint); err != nil {
		return arr
	}
	return arr
}

func (i *Interpreter) newU8ArrayValueFromString(text string) *runtime.ArrayValue {
	arr := runtime.ArrayStoreMonoValueFromU8String(text)
	if arr != nil {
		i.trackArrayValue(arr.Handle, arr)
	}
	return arr
}

func (i *Interpreter) newU8ArrayValueFromBytes(data []byte) *runtime.ArrayValue {
	arr := runtime.ArrayStoreMonoValueFromU8Bytes(data)
	if arr != nil {
		i.trackArrayValue(arr.Handle, arr)
	}
	return arr
}

func (i *Interpreter) newOwnedU8ArrayValueFromBytes(data []byte) *runtime.ArrayValue {
	arr := runtime.ArrayStoreMonoValueFromOwnedU8Bytes(data)
	if arr != nil {
		i.trackArrayValue(arr.Handle, arr)
	}
	return arr
}

func monoArrayHandleForGenericElementType(env *runtime.Environment, capacity int) (int64, bool) {
	if env == nil {
		return 0, false
	}
	raw, ok := env.Lookup("T")
	if !ok {
		return 0, false
	}
	var ref runtime.TypeRefValue
	switch typed := raw.(type) {
	case runtime.TypeRefValue:
		ref = typed
	case *runtime.TypeRefValue:
		if typed == nil {
			return 0, false
		}
		ref = *typed
	default:
		return 0, false
	}
	if ref.TypeName == "" || len(ref.TypeArgs) != 0 {
		return 0, false
	}
	switch ref.TypeName {
	case "i32":
		return runtime.ArrayStoreMonoNewWithCapacityI32(capacity), true
	case "i64":
		return runtime.ArrayStoreMonoNewWithCapacityI64(capacity), true
	case "bool":
		return runtime.ArrayStoreMonoNewWithCapacityBool(capacity), true
	case "char":
		return runtime.ArrayStoreMonoNewWithCapacityChar(capacity), true
	case "u8":
		return runtime.ArrayStoreMonoNewWithCapacityU8(capacity), true
	case "u32":
		return runtime.ArrayStoreMonoNewWithCapacityU32(capacity), true
	case "u64":
		return runtime.ArrayStoreMonoNewWithCapacityU64(capacity), true
	case "f64":
		return runtime.ArrayStoreMonoNewWithCapacityF64(capacity), true
	default:
		return 0, false
	}
}

func arrayHandleForCapacity(env *runtime.Environment, capacity int) int64 {
	if handle, ok := monoArrayHandleForGenericElementType(env, capacity); ok {
		return handle
	}
	if capacity <= 0 {
		return runtime.ArrayStoreNew()
	}
	return runtime.ArrayStoreNewReservedCapacity(capacity)
}

func (i *Interpreter) initArrayBuiltins() {
	if i.arrayReady {
		return
	}
	if i.arraysByHandle == nil {
		i.arraysByHandle = make(map[int64]arrayHandleTracking)
	}

	parseArrayHandle := func(val runtime.Value) (int64, error) {
		n, err := hostIntegerToInt64(val)
		if err != nil {
			return 0, fmt.Errorf("array handle must be an integer")
		}
		return n, nil
	}

	arrayNewHandle := runtime.NativeFunctionValue{
		Name:       "__able_array_new",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_array_new expects no arguments")
			}
			var env *runtime.Environment
			if ctx != nil {
				env = ctx.Env
			}
			handle := arrayHandleForCapacity(env, 0)
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arrayWithCapacity := runtime.NativeFunctionValue{
		Name:       "__able_array_with_capacity",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_with_capacity expects capacity argument")
			}
			capacity, err := arrayIndexFromValue(args[0])
			if err != nil {
				return nil, fmt.Errorf("capacity must be a non-negative integer")
			}
			if capacity < 0 {
				capacity = 0
			}
			var env *runtime.Environment
			if ctx != nil {
				env = ctx.Env
			}
			handle := arrayHandleForCapacity(env, capacity)
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arraySize := runtime.NativeFunctionValue{
		Name:       "__able_array_size",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_size expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			size, err := runtime.ArrayStoreSize(handle)
			if err != nil {
				return nil, err
			}
			sizeVal := int64(size)
			if boxed, ok := runtime.BoxedArrayMetadataU64Value(sizeVal); ok {
				return boxed, nil
			}
			return runtime.NewSmallInt(sizeVal, runtime.IntegerU64), nil
		},
	}

	arrayCapacity := runtime.NativeFunctionValue{
		Name:       "__able_array_capacity",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_capacity expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			capacity, err := runtime.ArrayStoreCapacity(handle)
			if err != nil {
				return nil, err
			}
			capacityVal := int64(capacity)
			if boxed, ok := runtime.BoxedArrayMetadataU64Value(capacityVal); ok {
				return boxed, nil
			}
			return runtime.NewSmallInt(capacityVal, runtime.IntegerU64), nil
		},
	}

	arraySetLen := runtime.NativeFunctionValue{
		Name:       "__able_array_set_len",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_set_len expects handle and length")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			length, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, fmt.Errorf("length must be a non-negative integer")
			}
			if err := runtime.ArrayStoreSetLength(handle, length); err != nil {
				return nil, err
			}
			if state, err := runtime.ArrayStoreState(handle); err == nil {
				i.syncArrayHandleLength(handle, state)
			}
			return runtime.NilValue{}, nil
		},
	}

	arrayRead := runtime.NativeFunctionValue{
		Name:       "__able_array_read",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_read expects handle and index")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			idx, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, err
			}
			val, err := runtime.ArrayStoreRead(handle, idx)
			if err != nil {
				return nil, err
			}
			return val, nil
		},
	}

	arrayWrite := runtime.NativeFunctionValue{
		Name:       "__able_array_write",
		Arity:      3,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__able_array_write expects handle, index, and value")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			idx, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, err
			}
			if idx < 0 {
				return nil, fmt.Errorf("index must be non-negative")
			}
			if err := runtime.ArrayStoreWrite(handle, idx, args[2]); err != nil {
				return nil, err
			}
			i.syncArrayHandleWriteAfterStore(handle, idx, args[2])
			return runtime.NilValue{}, nil
		},
	}

	arrayReserve := runtime.NativeFunctionValue{
		Name:       "__able_array_reserve",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_reserve expects handle and capacity")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			minCapacity, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, fmt.Errorf("capacity must be a non-negative integer")
			}
			if err := runtime.ArrayStoreReserve(handle, minCapacity); err != nil {
				return nil, err
			}
			if state, err := runtime.ArrayStoreState(handle); err == nil {
				i.syncArrayHandleMetadata(handle, state)
			}
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arrayClone := runtime.NativeFunctionValue{
		Name:       "__able_array_clone",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_clone expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			newHandle, err := runtime.ArrayStoreClone(handle)
			if err != nil {
				return nil, err
			}
			return runtime.NewSmallInt(newHandle, runtime.IntegerI64), nil
		},
	}

	arrayPkg := &runtime.PackageValue{
		Name:   "Array",
		Public: make(map[string]runtime.Value),
	}

	arrayNew := runtime.NativeFunctionValue{
		Name:       "Array.new",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			capacity := 0
			if len(args) > 1 {
				return nil, fmt.Errorf("Array.new expects zero or one argument")
			}
			if len(args) == 1 {
				val, err := arrayIndexFromValue(args[0])
				if err != nil {
					return nil, fmt.Errorf("Array.new capacity must be a non-negative integer")
				}
				if val < 0 {
					return nil, fmt.Errorf("Array.new capacity must be non-negative")
				}
				capacity = val
			}
			if capacity < 0 {
				capacity = 0
			}
			var env *runtime.Environment
			if ctx != nil {
				env = ctx.Env
			}
			if handle, ok := monoArrayHandleForGenericElementType(env, capacity); ok {
				return i.arrayValueFromHandle(handle, 0, capacity)
			}
			return i.newArrayValue(make([]runtime.Value, 0, capacity), capacity), nil
		},
	}

	arrayPkg.Public["new"] = arrayNew
	i.global.Define("__able_array_new", arrayNewHandle)
	i.global.Define("__able_array_with_capacity", arrayWithCapacity)
	i.global.Define("__able_array_size", arraySize)
	i.global.Define("__able_array_capacity", arrayCapacity)
	i.global.Define("__able_array_set_len", arraySetLen)
	i.global.Define("__able_array_read", arrayRead)
	i.global.Define("__able_array_write", arrayWrite)
	i.global.Define("__able_array_reserve", arrayReserve)
	i.global.Define("__able_array_clone", arrayClone)
	i.global.Define("Array", arrayPkg)
	i.arrayReady = true
}
