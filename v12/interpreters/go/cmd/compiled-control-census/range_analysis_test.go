package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRangeClosureDistinguishesUniversalAndCallSiteSafety(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
func __able_compiled_fn_bounded(n int32) (int32, *__ableControl) {
	wide := int64(n) - int64(1)
	if wide < -2147483648 || wide > 2147483647 {
		return 0, __able_raise_overflow()
	}
	return int32(wide), nil
}
func __able_compiled_fn_main() (int32, *__ableControl) {
	value, control := __able_compiled_fn_bounded(int32(4))
	if control != nil { return 0, control }
	return value, nil
}
func __able_raise_overflow() *__ableControl { return &__ableControl{} }
`)
	bounded := reportFunction(t, report, "__able_compiled_fn_bounded")
	if bounded.UniversalRangeFree {
		t.Fatalf("full i32 input must retain subtraction overflow: %#v", bounded)
	}
	if !bounded.ClosedDirectRangeFree || bounded.RangeClass != "call-site-specializable" {
		t.Fatalf("bounded internal call should be specializable: %#v", bounded)
	}
	if len(bounded.ClosedDirectParamRanges) != 1 || bounded.ClosedDirectParamRanges[0].Min != 4 || bounded.ClosedDirectParamRanges[0].Max != 4 {
		t.Fatalf("unexpected closed parameter range: %#v", bounded.ClosedDirectParamRanges)
	}
}

func TestRangeClosureProvesLocalLoopShiftUniversallySafe(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
func __able_compiled_fn_pixel() (int32, *__ableControl) {
	bit := int32(0)
	for bit < int32(8) {
		shift := int32(7) - bit
		value, control := __able_shift_left_signed(int64(1), int64(shift), 32)
		if control != nil { return 0, control }
		_ = value
		bit++
	}
	return 0, nil
}
func __able_compiled_fn_main() (int32, *__ableControl) { return __able_compiled_fn_pixel() }
func __able_shift_left_signed(value, shift int64, bits int) (int64, *__ableControl) {
	if shift < 0 || shift >= int64(bits) { return 0, &__ableControl{} }
	return value << uint(shift), nil
}
`)
	pixel := reportFunction(t, report, "__able_compiled_fn_pixel")
	if !pixel.UniversalRangeFree || pixel.RangeClass != "universally-range-safe" {
		t.Fatalf("bounded local shift should be universally safe: %#v", pixel)
	}
	if len(pixel.PrimitiveRangeBlockers) != 1 || !pixel.PrimitiveRangeBlockers[0].UniversalSafe {
		t.Fatalf("unexpected shift blocker report: %#v", pixel.PrimitiveRangeBlockers)
	}
}

func TestRangeClosureCoversEveryLoopCallArgument(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
func __able_compiled_fn_leaf(n int32) (int32, *__ableControl) { return n, nil }
func __able_compiled_fn_main() (int32, *__ableControl) {
	i := int32(0)
	for i < int32(8) {
		_, control := __able_compiled_fn_leaf(i)
		if control != nil { return 0, control }
		i++
	}
	return 0, nil
}
`)
	leaf := reportFunction(t, report, "__able_compiled_fn_leaf")
	if len(leaf.ClosedDirectParamRanges) != 1 || leaf.ClosedDirectParamRanges[0].Min != 0 || leaf.ClosedDirectParamRanges[0].Max != 7 {
		t.Fatalf("loop call range must cover [0,7]: %#v", leaf.ClosedDirectParamRanges)
	}
}

func TestRangeClosureDoesNotFreezeLoopCarriedShiftValue(t *testing.T) {
	report := analyzeRangeFixture(t, `package main
type __ableControl struct{}
func __able_compiled_fn_mask(k int32) (uint64, *__ableControl) {
	mask := uint64(0)
	i := int32(0)
	for i < k {
		next, control := __able_shift_left_unsigned(mask, uint64(2), 64)
		if control != nil { return 0, control }
		mask = next | uint64(3)
		i++
	}
	return mask, nil
}
func __able_compiled_fn_main() (uint64, *__ableControl) { return __able_compiled_fn_mask(int32(40)) }
func __able_shift_left_unsigned(value, shift uint64, bits int) (uint64, *__ableControl) {
	if shift >= uint64(bits) { return 0, &__ableControl{} }
	return value << shift, nil
}
`)
	mask := reportFunction(t, report, "__able_compiled_fn_mask")
	if mask.UniversalRangeFree || mask.ClosedDirectRangeFree {
		t.Fatalf("loop-carried mask growth must remain unproven: %#v", mask)
	}
	if len(mask.PrimitiveRangeBlockers) != 1 || mask.PrimitiveRangeBlockers[0].ClosedDirectSafe {
		t.Fatalf("unexpected loop-carried blocker report: %#v", mask.PrimitiveRangeBlockers)
	}
}

func analyzeRangeFixture(t *testing.T, source string) *censusReport {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "compiled.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := analyzeDirectory(directory)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func reportFunction(t *testing.T, report *censusReport, name string) *functionEffect {
	t.Helper()
	for _, effect := range report.Functions {
		if effect.Name == name {
			return effect
		}
	}
	t.Fatalf("missing function %s", name)
	return nil
}
