package main

import "testing"

func TestRelationalCensusCarriesArrayLengthThroughDirectCall(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
type __able_array_i32 struct{ Elements []int32 }
func __able_slice_len[T any](values []T) int { return len(values) }
func __able_compiled_fn_scan(values *__able_array_i32) (int32, *__ableControl) {
	n := int32(__able_slice_len(values.Elements))
	i := int32(0)
	for i < n {
		index := int(i)
		length := __able_slice_len(values.Elements)
		if index < 0 || index >= length { return 0, &__ableControl{} }
		_ = values.Elements[index]
		i++
	}
	return 0, nil
}
func __able_compiled_fn_main() (int32, *__ableControl) {
	raw := &__able_array_i32{}
	values := raw
	tmp := values
	tmp.Elements = append(tmp.Elements, int32(1))
	tmp = values
	tmp.Elements = append(tmp.Elements, int32(2))
	return __able_compiled_fn_scan(values)
}
`)
	scan := reportFunction(t, report, "__able_compiled_fn_scan")
	if len(scan.RelationalBlockers) != 1 || !scan.RelationalBlockers[0].ClosedDirectSafe {
		t.Fatalf("direct-call length and loop interval should prove the guard: %#v", scan.RelationalBlockers)
	}
	if !hasAggregateRange(scan.ClosedAggregateFacts, "values.#length", 2, 2) {
		t.Fatalf("missing closed Array length fact: %#v", scan.ClosedAggregateFacts)
	}
}

func TestRelationalCensusInvalidatesArrayFactsAtUnknownCall(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
type __able_array_i32 struct{ Elements []int32 }
func __able_slice_len[T any](values []T) int { return len(values) }
func unknown(values *__able_array_i32) {}
func __able_compiled_fn_scan(values *__able_array_i32) (int32, *__ableControl) {
	index := int(0)
	length := __able_slice_len(values.Elements)
	if index < 0 || index >= length { return 0, &__ableControl{} }
	return values.Elements[index], nil
}
func __able_compiled_fn_main() (int32, *__ableControl) {
	values := &__able_array_i32{Elements: []int32{1}}
	unknown(values)
	return __able_compiled_fn_scan(values)
}
`)
	scan := reportFunction(t, report, "__able_compiled_fn_scan")
	if len(scan.RelationalBlockers) != 1 || scan.RelationalBlockers[0].ClosedDirectSafe {
		t.Fatalf("unknown aliasing call must invalidate the length fact: %#v", scan.RelationalBlockers)
	}
}

func TestRelationalCensusRequiresEveryDirectCallShape(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
type __able_array_i32 struct{ Elements []int32 }
func __able_slice_len[T any](values []T) int { return len(values) }
func unknownArray() *__able_array_i32 { return nil }
func __able_compiled_fn_scan(values *__able_array_i32) (int32, *__ableControl) {
	index := int(0)
	length := __able_slice_len(values.Elements)
	if index < 0 || index >= length { return 0, &__ableControl{} }
	return values.Elements[index], nil
}
func __able_compiled_fn_main() (int32, *__ableControl) {
	known := &__able_array_i32{Elements: []int32{1}}
	_, control := __able_compiled_fn_scan(known)
	if control != nil { return 0, control }
	unknown := unknownArray()
	return __able_compiled_fn_scan(unknown)
}
`)
	scan := reportFunction(t, report, "__able_compiled_fn_scan")
	if len(scan.RelationalBlockers) != 1 || scan.RelationalBlockers[0].ClosedDirectSafe {
		t.Fatalf("one unknown direct-call shape must block specialization: %#v", scan.RelationalBlockers)
	}
}

func TestRelationalCensusCarriesStructAndBitmaskIntervals(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
type Position struct{ Row int32; Choices int32 }
func __able_compiled_fn_use(pos *Position, choices int32) (int32, *__ableControl) {
	wide := int64(pos.Row) + int64(choices)
	if wide < -2147483648 || wide > 2147483647 { return 0, &__ableControl{} }
	return int32(wide), nil
}
func __able_compiled_fn_main() (int32, *__ableControl) {
	pos := &Position{Row: int32(3), Choices: int32(0)}
	choices := int32(511) ^ int32(7)
	return __able_compiled_fn_use(pos, choices)
}
`)
	use := reportFunction(t, report, "__able_compiled_fn_use")
	if !hasAggregateRange(use.ClosedAggregateFacts, "pos.Row", 3, 3) {
		t.Fatalf("missing struct-field interval: %#v", use.ClosedAggregateFacts)
	}
	if len(use.ClosedDirectParamRanges) != 1 ||
		use.ClosedDirectParamRanges[0].Min != 0 || use.ClosedDirectParamRanges[0].Max != 511 {
		t.Fatalf("missing bitmask domain: %#v", use.ClosedDirectParamRanges)
	}
}

func TestRelationalCensusCarriesReturnedArrayShapeIntoNextCall(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
type __able_array_i32 struct{ Elements []int32 }
func __able_compiled_fn_make() (*__able_array_i32, *__ableControl) {
	if false { return nil, &__ableControl{} }
	values := &__able_array_i32{Elements: []int32{1, 2, 3}}
	return values, nil
}
func __able_compiled_fn_use(values *__able_array_i32) (int32, *__ableControl) { return 0, nil }
func __able_compiled_fn_main() (int32, *__ableControl) {
	raw, control := __able_compiled_fn_make()
	if control != nil { return 0, control }
	values := raw
	return __able_compiled_fn_use(values)
}
`)
	use := reportFunction(t, report, "__able_compiled_fn_use")
	if !hasAggregateRange(use.ClosedAggregateFacts, "values.#length", 3, 3) {
		t.Fatalf("missing returned Array shape at next direct call: %#v", use.ClosedAggregateFacts)
	}
}

func hasAggregateRange(facts []aggregateRange, path string, minValue, maxValue int64) bool {
	for _, fact := range facts {
		if fact.Path == path && fact.Min == minValue && fact.Max == maxValue {
			return true
		}
	}
	return false
}
