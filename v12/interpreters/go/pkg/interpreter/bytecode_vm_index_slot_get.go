package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) lookupDirectCompatibleArrayIndexGetSlot(obj runtime.Value) (*runtime.ArrayValue, int64, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, 0, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(obj)
	if !ok || arr == nil {
		return nil, 0, false
	}
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	return vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(
		bytecodeIndexMethodCacheGet,
		arr,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	)
}

func (vm *bytecodeVM) execArrayIndexGetSlot(instr *bytecodeInstruction) error {
	if vm == nil || vm.interp == nil || instr == nil {
		return fmt.Errorf("bytecode array index slot missing VM or instruction")
	}
	objSlot, idxSlot := instr.argCount, instr.loopBreak
	if objSlot < 0 || objSlot >= len(vm.slots) || idxSlot < 0 || idxSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array index slot out of range")
	}
	obj := vm.slots[objSlot]
	statsEnabled := vm.interp.bytecodeStatsEnabled
	if statsEnabled {
		vm.interp.recordBytecodeArrayIndexSlotLookup()
	}
	if !vm.hasI32RegisterFrame() {
		idxVal := vm.slots[idxSlot]
		var (
			result                     runtime.Value
			err                        error
			elementToken               uint16
			elementTokenKnown          bool
			elementTokenExactPrimitive bool
			fallbackReason             = bytecodeArrayIndexSlotFallbackFastDisabled
		)
		if vm.interp.canUseDirectArrayIndexGetFastPath() {
			fallbackReason = bytecodeArrayIndexSlotFallbackReceiverMiss
			if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok && arr != nil {
				fallbackReason = bytecodeArrayIndexSlotFallbackIndexMiss
				if idx, small := bytecodeDirectSmallArrayIndex(idxVal); small {
					fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
					if state, tracked := bytecodeTrackedArrayState(arr); tracked {
						if idx < 0 || idx >= len(state.Values) {
							result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
						} else if val := state.Values[idx]; val != nil {
							result = val
						} else {
							result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
						}
						if statsEnabled {
							vm.interp.recordBytecodeArrayIndexSlotTrackedHit()
						}
						vm.appendStackValue(result)
						if vm.canSkipArrayGetSuccessPropagation(result, state.ElementTypeToken, state.ElementTypeTokenKnown) {
							vm.ip += 2
							return nil
						}
						vm.ip++
						return nil
					}
					handle, handleOK, handleErr := vm.arrayHandleFast(arr)
					if handleErr != nil {
						err = handleErr
					} else if handleOK {
						if fast, token, tokenKnown, handled, directErr := vm.appendDirectMonoUnsignedArrayIndexGetAtHandle(handle, idx); handled {
							if statsEnabled {
								vm.interp.recordBytecodeArrayIndexSlotMonoUnsignedHit()
							}
							if directErr != nil {
								err = directErr
							} else {
								if vm.canSkipExactPrimitiveArrayGetSuccessPropagation(fast, token, tokenKnown) {
									vm.ip += 2
									return nil
								}
								if vm.canSkipArrayGetSuccessPropagation(fast, token, tokenKnown) {
									vm.ip += 2
									return nil
								}
								vm.ip++
								return nil
							}
						} else {
							result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
							elementTokenExactPrimitive = true
							if statsEnabled && (result != nil || err != nil) {
								vm.interp.recordBytecodeArrayIndexSlotDirectHit()
							}
						}
					} else {
						fallbackReason = bytecodeArrayIndexSlotFallbackHandleMiss
						result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, 0)
						elementTokenExactPrimitive = true
						if statsEnabled && (result != nil || err != nil) {
							vm.interp.recordBytecodeArrayIndexSlotDirectHit()
						}
					}
				} else if value, token, tokenKnown, handled, directErr := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, 0); handled {
					result, err = value, directErr
					elementToken, elementTokenKnown = token, tokenKnown
					elementTokenExactPrimitive = true
					if statsEnabled {
						vm.interp.recordBytecodeArrayIndexSlotDirectHit()
					}
				} else {
					fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
				}
			}
		} else if arr, handle, ok := vm.lookupDirectCompatibleArrayIndexGetSlot(obj); ok {
			fallbackReason = bytecodeArrayIndexSlotFallbackIndexMiss
			if idx, small := bytecodeDirectSmallArrayIndex(idxVal); small {
				fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
				result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
				elementTokenExactPrimitive = true
				if statsEnabled && (result != nil || err != nil) {
					vm.interp.recordBytecodeArrayIndexSlotDirectHit()
				}
			} else if value, token, tokenKnown, handled, directErr := vm.resolveDirectArrayIndexGetWithValidatedHandleAndToken(arr, idxVal, handle); handled {
				result, err = value, directErr
				elementToken, elementTokenKnown = token, tokenKnown
				elementTokenExactPrimitive = true
				if statsEnabled {
					vm.interp.recordBytecodeArrayIndexSlotDirectHit()
				}
			} else {
				fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
			}
		}
		if result == nil && err == nil {
			if statsEnabled {
				vm.interp.recordBytecodeArrayIndexSlotFallback(fallbackReason)
			}
			elementTokenExactPrimitive = false
			result, err = vm.resolveIndexGet(obj, idxVal)
		}
		if err != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return err
		}
		vm.appendStackValue(result)
		if !elementTokenKnown && vm.hasFollowingSuccessPropagation(result) {
			if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok {
				elementToken, elementTokenKnown = vm.arrayElementTypeTokenForPropagation(arr)
			}
		}
		if elementTokenExactPrimitive && vm.canSkipExactPrimitiveArrayGetSuccessPropagation(result, elementToken, elementTokenKnown) {
			vm.ip += 2
			return nil
		}
		if vm.canSkipArrayGetSuccessPropagation(result, elementToken, elementTokenKnown) {
			vm.ip += 2
			return nil
		}
		vm.ip++
		return nil
	}
	var (
		result                     runtime.Value
		err                        error
		idxVal                     runtime.Value
		elementToken               uint16
		elementTokenKnown          bool
		elementTokenExactPrimitive bool
		fallbackReason             = bytecodeArrayIndexSlotFallbackFastDisabled
	)
	if vm.interp.canUseDirectArrayIndexGetFastPath() {
		fallbackReason = bytecodeArrayIndexSlotFallbackReceiverMiss
		if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok && arr != nil {
			fallbackReason = bytecodeArrayIndexSlotFallbackIndexMiss
			if idx, small := vm.slotDirectSmallArrayIndexValidated(idxSlot); small {
				fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
				if state, tracked := bytecodeTrackedArrayState(arr); tracked {
					if idx < 0 || idx >= len(state.Values) {
						result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
					} else if val := state.Values[idx]; val != nil {
						result = val
					} else {
						result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
					}
					if statsEnabled {
						vm.interp.recordBytecodeArrayIndexSlotTrackedHit()
					}
					vm.appendStackValue(result)
					if vm.canSkipArrayGetSuccessPropagation(result, state.ElementTypeToken, state.ElementTypeTokenKnown) {
						vm.ip += 2
						return nil
					}
					vm.ip++
					return nil
				}
				handle, handleOK, handleErr := vm.arrayHandleFast(arr)
				if handleErr != nil {
					err = handleErr
				} else if handleOK {
					if fast, token, tokenKnown, handled, directErr := vm.appendDirectMonoUnsignedArrayIndexGetAtHandle(handle, idx); handled {
						if statsEnabled {
							vm.interp.recordBytecodeArrayIndexSlotMonoUnsignedHit()
						}
						if directErr != nil {
							err = directErr
						} else {
							if vm.canSkipExactPrimitiveArrayGetSuccessPropagation(fast, token, tokenKnown) {
								vm.ip += 2
								return nil
							}
							if vm.canSkipArrayGetSuccessPropagation(fast, token, tokenKnown) {
								vm.ip += 2
								return nil
							}
							vm.ip++
							return nil
						}
					} else {
						result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
						elementTokenExactPrimitive = true
						if statsEnabled && (result != nil || err != nil) {
							vm.interp.recordBytecodeArrayIndexSlotDirectHit()
						}
					}
				} else {
					fallbackReason = bytecodeArrayIndexSlotFallbackHandleMiss
					result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, 0)
					elementTokenExactPrimitive = true
					if statsEnabled && (result != nil || err != nil) {
						vm.interp.recordBytecodeArrayIndexSlotDirectHit()
					}
				}
			} else if idxVal = vm.slotMaterializedValue(idxSlot); idxVal != nil {
				fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
				if value, token, tokenKnown, handled, directErr := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, 0); handled {
					result, err = value, directErr
					elementToken, elementTokenKnown = token, tokenKnown
					elementTokenExactPrimitive = true
					if statsEnabled {
						vm.interp.recordBytecodeArrayIndexSlotDirectHit()
					}
				}
			}
		}
	} else if arr, handle, ok := vm.lookupDirectCompatibleArrayIndexGetSlot(obj); ok {
		fallbackReason = bytecodeArrayIndexSlotFallbackIndexMiss
		if idx, small := vm.slotDirectSmallArrayIndexValidated(idxSlot); small {
			fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
			result, elementToken, elementTokenKnown, err = vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
			elementTokenExactPrimitive = true
			if statsEnabled && (result != nil || err != nil) {
				vm.interp.recordBytecodeArrayIndexSlotDirectHit()
			}
		} else if idxVal = vm.slotMaterializedValue(idxSlot); idxVal != nil {
			fallbackReason = bytecodeArrayIndexSlotFallbackDirectMiss
			if value, token, tokenKnown, handled, directErr := vm.resolveDirectArrayIndexGetWithValidatedHandleAndToken(arr, idxVal, handle); handled {
				result, err = value, directErr
				elementToken, elementTokenKnown = token, tokenKnown
				elementTokenExactPrimitive = true
				if statsEnabled {
					vm.interp.recordBytecodeArrayIndexSlotDirectHit()
				}
			}
		}
	}
	if result == nil && err == nil {
		if statsEnabled {
			vm.interp.recordBytecodeArrayIndexSlotFallback(fallbackReason)
		}
		elementTokenExactPrimitive = false
		if idxVal == nil {
			idxVal = vm.slotMaterializedValue(idxSlot)
		}
		result, err = vm.resolveIndexGet(obj, idxVal)
	}
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.appendStackValue(result)
	if !elementTokenKnown && vm.hasFollowingSuccessPropagation(result) {
		if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok {
			elementToken, elementTokenKnown = vm.arrayElementTypeTokenForPropagation(arr)
		}
	}
	if elementTokenExactPrimitive && vm.canSkipExactPrimitiveArrayGetSuccessPropagation(result, elementToken, elementTokenKnown) {
		vm.ip += 2
		return nil
	}
	if vm.canSkipArrayGetSuccessPropagation(result, elementToken, elementTokenKnown) {
		vm.ip += 2
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) appendDirectMonoUnsignedArrayIndexGetAt(arr *runtime.ArrayValue, idx int) (runtime.Value, uint16, bool, bool, error) {
	if vm == nil || arr == nil {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, true, err
	}
	if !ok {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	return vm.appendDirectMonoUnsignedArrayIndexGetAtHandle(handle, idx)
}

func (vm *bytecodeVM) appendDirectMonoUnsignedArrayIndexGetAtHandle(handle int64, idx int) (runtime.Value, uint16, bool, bool, error) {
	if vm == nil || handle == 0 {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	var info runtime.ArrayStoreMonoPrimitiveReadInfo
	if ok, err := runtime.ArrayStoreMonoPrimitiveReadInfoInto(handle, idx, &info); err != nil {
		return nil, bytecodeIndexTypeUnknown, false, true, err
	} else if ok && info.InBounds {
		switch info.Kind {
		case runtime.ArrayStoreMonoPrimitiveReadU32:
			return vm.appendRawIntegerStack(runtime.IntegerU32, int64(uint32(info.Uint64))), bytecodeIndexTypeU32, true, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadU64:
			if info.Uint64 <= math.MaxInt64 {
				return vm.appendRawIntegerStack(runtime.IntegerU64, int64(info.Uint64)), bytecodeIndexTypeU64, true, true, nil
			}
			value := bytecodeUnsignedIntegerValue(runtime.IntegerU64, info.Uint64)
			vm.appendStackValue(value)
			return value, bytecodeIndexTypeU64, true, true, nil
		}
	}
	return nil, bytecodeIndexTypeUnknown, false, false, nil
}
