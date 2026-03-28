package compiler

import (
	"strings"
	"testing"
)

func TestCompilerKernelHasherInterfaceAdapterKeepsNativeStateCarrier(t *testing.T) {
	result := compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{KernelHasher, Hasher}",
		"",
		"fn probe(hasher: Hasher) -> u64 {",
		"  hasher.write_u16(515_u16)",
		"  hasher.finish()",
		"}",
		"",
		"fn main() -> u64 {",
		"  probe(KernelHasher.new())",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "Value *KernelHasher") {
		t.Fatalf("expected Hasher adapter to retain native *KernelHasher state:\n%s", compiledSrc)
	}

	probeBody, ok := findCompiledFunction(result, "__able_compiled_fn_probe")
	if !ok {
		t.Fatalf("could not find compiled probe")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_member_get_method(",
		"__able_iface_Hasher_to_runtime_value(",
	} {
		if strings.Contains(probeBody, fragment) {
			t.Fatalf("expected Hasher calls in probe to stay static:\n%s", probeBody)
		}
	}
}

func TestCompilerKernelHasherThroughHasherExecutes(t *testing.T) {
	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-kernel-hasher-iface", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{KernelHasher, Hasher}",
		"",
		"fn probe(hasher: Hasher) -> u64 {",
		"  hasher.write_u16(515_u16)",
		"  hasher.finish()",
		"}",
		"",
		"fn main() -> void {",
		"  print(probe(KernelHasher.new()))",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	}))

	if stdout != "592598317564770290" {
		t.Fatalf("expected KernelHasher through Hasher to produce 592598317564770290, got %q", stdout)
	}
}
