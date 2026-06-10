package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerBuildsLargeI128AndU128Literals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large integer literal build test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-intlit")

	source := `fn main() -> i32 {
  min := -9223372036854775808_i128
  max := 9223372036854775807_i128
  unsigned_max := 18446744073709551615_i128
  huge_unsigned := 340282366920938463463374607431768211455_u128
  if min < max && unsigned_max > 0_i128 && huge_unsigned > 0_u128 { 0 } else { 1 }
}
`
	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outputDir := filepath.Join(workDir, "out")
	comp := New(Options{
		PackageName:        "main",
		EmitMain:           true,
		EntryPath:          entryPath,
		RequireNoFallbacks: true,
	})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-bin")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}
}

func TestCompilerExecutesI128AndU128PrimitiveOperationsWithoutFallbacks(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-wide-int-ops", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  print(18446744073709551616_i128 + 5_i128)",
		"  print(123456789_u128 * 987654321_u128)",
		"  print(-7_i128 // 3_i128)",
		"  print(-7_i128 % 3_i128)",
		"  shifted := 1_u128 .<< 100_u128",
		"  print(shifted)",
		"  print(shifted .>> 96_u128)",
		"  print((-1_i128) .>> 127_i128)",
		"  print((-1_i128) as u128)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})

	want := strings.Join([]string{
		"18446744073709551621",
		"121932631112635269",
		"-3",
		"2",
		"1267650600228229401496703205376",
		"16",
		"-1",
		"340282366920938463463374607431768211455",
		"",
	}, "\n")
	if stdout != want {
		t.Fatalf("wide primitive output mismatch:\nwant:\n%s\ngot:\n%s", want, stdout)
	}
}

func TestCompilerWideIntegerModuleConstantsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"SIGNED_MIN: i128 := -9223372036854775808_i128",
		"UNSIGNED_MASK: u128 := 18446744073709551615_u128",
		"",
		"fn signed() -> i128 { SIGNED_MIN }",
		"fn masked(value: u128) -> u128 { value .& UNSIGNED_MASK }",
		"",
	}, "\n"))

	for _, function := range []string{"__able_compiled_fn_signed", "__able_compiled_fn_masked"} {
		body, ok := findCompiledFunction(result, function)
		if !ok {
			t.Fatalf("could not find %s", function)
		}
		for _, dynamic := range []string{"runtime.NewBigIntValue", "runtime.Int128FromValue", "runtime.Uint128FromValue", "new(big.Int)"} {
			if strings.Contains(body, dynamic) {
				t.Fatalf("expected wide module constant in %s to stay native, found %q:\n%s", function, dynamic, body)
			}
		}
	}
}
