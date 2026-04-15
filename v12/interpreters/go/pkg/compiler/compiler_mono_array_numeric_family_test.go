package compiler

import (
	"strings"
	"testing"
)

func TestCompilerExperimentalMonoArraysRemainingNumericFamilyUsesSpecializedWrappers(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  bytes: Array i8 = [1_i8, 2_i8]",
		"  shorts: Array i16 = [3_i16, 4_i16]",
		"  smalls: Array u16 = [5_u16, 6_u16]",
		"  counts: Array u32 = [7_u32, 8_u32]",
		"  totals: Array u64 = [9_u64, 10_u64]",
		"  spans: Array isize = [11 as isize, 12 as isize]",
		"  widths: Array usize = [13 as usize, 14 as usize]",
		"  ratios: Array f32 = [1.25_f32, 2.5_f32]",
		"  _ = bytes",
		"  _ = shorts",
		"  _ = smalls",
		"  _ = counts",
		"  _ = totals",
		"  _ = spans",
		"  _ = widths",
		"  _ = ratios",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_array_i8 struct {",
		"Elements []int8",
		"type __able_array_i16 struct {",
		"Elements []int16",
		"type __able_array_u16 struct {",
		"Elements []uint16",
		"type __able_array_u32 struct {",
		"Elements []uint32",
		"type __able_array_u64 struct {",
		"Elements []uint64",
		"type __able_array_isize struct {",
		"Elements []int",
		"type __able_array_usize struct {",
		"Elements []uint",
		"type __able_array_f32 struct {",
		"Elements []float32",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected remaining numeric mono-array lowering to contain %q", fragment)
		}
	}
}

func TestCompilerExperimentalMonoArraysRemainingNumericFamilyExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  ok := true",
		"  bytes: Array i8 = [1_i8, 2_i8]",
		"  bytes.push(3_i8)",
		"  ok = ok && bytes[2]! == 3_i8",
		"  shorts: Array i16 = [4_i16]",
		"  shorts.push(5_i16)",
		"  ok = ok && shorts.get(1)! == 5_i16",
		"  smalls: Array u16 = [6_u16]",
		"  smalls[0] = 7_u16",
		"  ok = ok && smalls[0]! == 7_u16",
		"  counts: Array u32 = [8_u32]",
		"  counts.push(9_u32)",
		"  ok = ok && counts[1]! == 9_u32",
		"  totals: Array u64 = [10_u64]",
		"  totals.push(11_u64)",
		"  ok = ok && totals[1]! == 11_u64",
		"  spans: Array isize = [12 as isize]",
		"  spans.push(13 as isize)",
		"  ok = ok && spans[1]! == 13 as isize",
		"  widths: Array usize = [14 as usize]",
		"  widths.push(15 as usize)",
		"  ok = ok && widths[1]! == 15 as usize",
		"  ratios: Array f32 = [1.25_f32]",
		"  ratios.push(2.5_f32)",
		"  ok = ok && ratios[1]! == 2.5_f32",
		"  if ok { print(\"ok\") } else { print(\"bad\") }",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunSourceWithOptions(t, "ablec-mono-array-numeric-family-", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("expected remaining numeric mono-array program to print ok, got %q", stdout)
	}
}
