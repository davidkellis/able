package runtime

func seedArrayI32CacheFromMonoValues(state *ArrayState, values []int32) {
	if state == nil {
		return
	}
	length := len(values)
	if length == 0 {
		state.CachedI32Values = nil
		state.CachedI32ValuesValid = nil
		state.CachedI32ValuesCount = 0
		state.CachedI32ValuesKnown = true
		return
	}
	raws := make([]int32, length)
	copy(raws, values)
	valid := make([]bool, length)
	for idx := range valid {
		valid[idx] = true
	}
	state.CachedI32Values = raws
	state.CachedI32ValuesValid = valid
	state.CachedI32ValuesCount = length
	state.CachedI32ValuesKnown = true
}

func clearArrayI32Cache(state *ArrayState) {
	if state == nil {
		return
	}
	state.CachedI32Values = nil
	state.CachedI32ValuesValid = nil
	state.CachedI32ValuesCount = 0
	state.CachedI32ValuesKnown = false
}

func arrayStateI32CacheHoleValue(value Value) bool {
	if value == nil {
		return true
	}
	switch value.(type) {
	case NilValue, *NilValue:
		return true
	default:
		return false
	}
}

func invalidateArrayI32CacheSlot(state *ArrayState, index int) {
	if state == nil || index < 0 || index >= len(state.CachedI32Values) || index >= len(state.CachedI32ValuesValid) {
		return
	}
	if state.CachedI32ValuesValid[index] {
		state.CachedI32Values[index] = 0
		state.CachedI32ValuesValid[index] = false
		state.CachedI32ValuesCount--
		if state.CachedI32ValuesCount < 0 {
			state.CachedI32ValuesCount = 0
		}
	}
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.CachedI32ValuesValid)
}

func updateArrayI32CacheForDynamicWrite(state *ArrayState, oldLength int, index int, value Value) {
	if state == nil || index < 0 || index >= len(state.Values) {
		return
	}
	cacheLength := len(state.CachedI32Values)
	if cacheLength != len(state.CachedI32ValuesValid) {
		clearArrayI32Cache(state)
		return
	}
	cacheActive := cacheLength > 0 || state.CachedI32ValuesCount > 0 || state.CachedI32ValuesKnown
	if !cacheActive {
		return
	}
	if cacheLength == oldLength && oldLength != len(state.Values) {
		adjustArrayI32CacheLength(state, oldLength, len(state.Values))
		cacheLength = len(state.CachedI32Values)
	}
	if cacheLength != len(state.Values) || len(state.CachedI32ValuesValid) != len(state.Values) {
		clearArrayI32Cache(state)
		return
	}
	if arrayStateI32CacheHoleValue(value) {
		invalidateArrayI32CacheSlot(state, index)
		return
	}
	raw, err := int32FromValue(value)
	if err != nil {
		clearArrayI32Cache(state)
		return
	}
	if !state.CachedI32ValuesValid[index] {
		state.CachedI32ValuesCount++
	}
	state.CachedI32Values[index] = raw
	state.CachedI32ValuesValid[index] = true
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
}

func adjustArrayI32CacheLength(state *ArrayState, oldLength int, newLength int) {
	if state == nil || len(state.CachedI32Values) == 0 || len(state.CachedI32ValuesValid) == 0 {
		return
	}
	if newLength <= oldLength {
		if newLength <= len(state.CachedI32Values) && newLength <= len(state.CachedI32ValuesValid) {
			removedValidCount := 0
			for _, valid := range state.CachedI32ValuesValid[newLength:oldLength] {
				if valid {
					removedValidCount++
				}
			}
			clear(state.CachedI32Values[newLength:oldLength])
			clear(state.CachedI32ValuesValid[newLength:oldLength])
			state.CachedI32Values = state.CachedI32Values[:newLength]
			state.CachedI32ValuesValid = state.CachedI32ValuesValid[:newLength]
			state.CachedI32ValuesCount -= removedValidCount
			if state.CachedI32ValuesCount < 0 {
				state.CachedI32ValuesCount = 0
			}
			if state.CachedI32ValuesCount > newLength {
				state.CachedI32ValuesCount = newLength
			}
			state.CachedI32ValuesKnown = state.CachedI32ValuesCount == newLength
			return
		}
		clearArrayI32Cache(state)
		return
	}
	if oldLength < 0 || oldLength > len(state.CachedI32Values) || oldLength > len(state.CachedI32ValuesValid) {
		clearArrayI32Cache(state)
		return
	}
	if newLength <= cap(state.CachedI32Values) && newLength <= cap(state.CachedI32ValuesValid) {
		prevLength := len(state.CachedI32Values)
		state.CachedI32Values = state.CachedI32Values[:newLength]
		state.CachedI32ValuesValid = state.CachedI32ValuesValid[:newLength]
		clear(state.CachedI32Values[prevLength:newLength])
		clear(state.CachedI32ValuesValid[prevLength:newLength])
	} else {
		grownValues := make([]int32, newLength)
		copy(grownValues, state.CachedI32Values[:oldLength])
		grownValid := make([]bool, newLength)
		copy(grownValid, state.CachedI32ValuesValid[:oldLength])
		state.CachedI32Values = grownValues
		state.CachedI32ValuesValid = grownValid
	}
	if state.CachedI32ValuesCount > newLength {
		state.CachedI32ValuesCount = newLength
	}
	state.CachedI32ValuesKnown = false
}
