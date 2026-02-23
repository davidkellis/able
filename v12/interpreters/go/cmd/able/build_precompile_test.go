package main

import (
	"os"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestResolveBuildPrecompileStdlibFromEnvExplicitValues(t *testing.T) {
	t.Setenv("ABLE_BUILD_PRECOMPILE_STDLIB", "")
	value, err := resolveBuildPrecompileStdlibFromEnv()
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if value {
		t.Fatalf("expected explicit empty env value to disable precompile toggle")
	}

	t.Setenv("ABLE_BUILD_PRECOMPILE_STDLIB", "true")
	value, err = resolveBuildPrecompileStdlibFromEnv()
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if !value {
		t.Fatalf("expected true env to enable precompile")
	}
}

func TestResolveBuildPrecompileStdlibFromEnvMissingDefaultsTrue(t *testing.T) {
	if err := os.Unsetenv("ABLE_BUILD_PRECOMPILE_STDLIB"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	value, err := resolveBuildPrecompileStdlibFromEnv()
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if !value {
		t.Fatalf("expected missing env to default to true")
	}
}

func TestResolveBuildPrecompileStdlibFromEnvInvalid(t *testing.T) {
	t.Setenv("ABLE_BUILD_PRECOMPILE_STDLIB", "sometimes")
	if _, err := resolveBuildPrecompileStdlibFromEnv(); err == nil {
		t.Fatalf("expected parse error for invalid env value")
	}
}

func TestDiscoverPrecompilePackagesIncludesStdlibAndKernel(t *testing.T) {
	searchPaths := []driver.SearchPath{
		{Path: repoStdlibPath(t), Kind: driver.RootStdlib},
		{Path: repoKernelPath(t), Kind: driver.RootStdlib},
	}
	packages, err := discoverPrecompilePackages(searchPaths, false)
	if err != nil {
		t.Fatalf("discover precompile packages: %v", err)
	}
	if !containsPath(packages, "able.io") {
		t.Fatalf("expected able.io in precompile package list")
	}
	if !containsPath(packages, "able.io.path") {
		t.Fatalf("expected able.io.path in precompile package list")
	}
	if !containsPath(packages, "able.io.temp") {
		t.Fatalf("expected able.io.temp in precompile package list")
	}
	if !containsPath(packages, "able.os") {
		t.Fatalf("expected able.os in precompile package list")
	}
	if !containsPath(packages, "able.process") {
		t.Fatalf("expected able.process in precompile package list")
	}
	if !containsPath(packages, "able.term") {
		t.Fatalf("expected able.term in precompile package list")
	}
	if !containsPath(packages, "able.fs") {
		t.Fatalf("expected able.fs in precompile package list")
	}
	if !containsPath(packages, "able.spec") {
		t.Fatalf("expected able.spec in precompile package list")
	}
	if !containsPath(packages, "able.collections.enumerable") {
		t.Fatalf("expected able.collections.enumerable in precompile package list")
	}
	if !containsPath(packages, "able.collections.list") {
		t.Fatalf("expected able.collections.list in precompile package list")
	}
	if !containsPath(packages, "able.collections.vector") {
		t.Fatalf("expected able.collections.vector in precompile package list")
	}
	if !containsPath(packages, "able.collections.tree_map") {
		t.Fatalf("expected able.collections.tree_map in precompile package list")
	}
	if !containsPath(packages, "able.collections.tree_set") {
		t.Fatalf("expected able.collections.tree_set in precompile package list")
	}
	if !containsPath(packages, "able.collections.persistent_map") {
		t.Fatalf("expected able.collections.persistent_map in precompile package list")
	}
	if !containsPath(packages, "able.collections.persistent_sorted_set") {
		t.Fatalf("expected able.collections.persistent_sorted_set in precompile package list")
	}
	if !containsPath(packages, "able.collections.persistent_queue") {
		t.Fatalf("expected able.collections.persistent_queue in precompile package list")
	}
	if !containsPath(packages, "able.collections.linked_list") {
		t.Fatalf("expected able.collections.linked_list in precompile package list")
	}
	if !containsPath(packages, "able.collections.lazy_seq") {
		t.Fatalf("expected able.collections.lazy_seq in precompile package list")
	}
	if !containsPath(packages, "able.collections.hash_map") {
		t.Fatalf("expected able.collections.hash_map in precompile package list")
	}
	if !containsPath(packages, "able.collections.hash_set") {
		t.Fatalf("expected able.collections.hash_set in precompile package list")
	}
	if !containsPath(packages, "able.collections.deque") {
		t.Fatalf("expected able.collections.deque in precompile package list")
	}
	if !containsPath(packages, "able.collections.queue") {
		t.Fatalf("expected able.collections.queue in precompile package list")
	}
	if !containsPath(packages, "able.collections.array") {
		t.Fatalf("expected able.collections.array in precompile package list")
	}
	if !containsPath(packages, "able.collections.range") {
		t.Fatalf("expected able.collections.range in precompile package list")
	}
	if !containsPath(packages, "able.collections.bit_set") {
		t.Fatalf("expected able.collections.bit_set in precompile package list")
	}
	if !containsPath(packages, "able.collections.heap") {
		t.Fatalf("expected able.collections.heap in precompile package list")
	}
	if !containsPath(packages, "able.concurrency") {
		t.Fatalf("expected able.concurrency in precompile package list")
	}
	if !containsPath(packages, "able.concurrency.concurrent_queue") {
		t.Fatalf("expected able.concurrency.concurrent_queue in precompile package list")
	}
	if !containsPath(packages, "able.math") {
		t.Fatalf("expected able.math in precompile package list")
	}
	if !containsPath(packages, "able.core.numeric") {
		t.Fatalf("expected able.core.numeric in precompile package list")
	}
	if !containsPath(packages, "able.text.string") {
		t.Fatalf("expected able.text.string in precompile package list")
	}
	if !containsPath(packages, "able.text.regex") {
		t.Fatalf("expected able.text.regex in precompile package list")
	}
	if !containsPath(packages, "able.text.ascii") {
		t.Fatalf("expected able.text.ascii in precompile package list")
	}
	if !containsPath(packages, "able.text.automata") {
		t.Fatalf("expected able.text.automata in precompile package list")
	}
	if !containsPath(packages, "able.text.automata_dsl") {
		t.Fatalf("expected able.text.automata_dsl in precompile package list")
	}
	if !containsPath(packages, "able.test.protocol") {
		t.Fatalf("expected able.test.protocol in precompile package list")
	}
	if !containsPath(packages, "able.test.harness") {
		t.Fatalf("expected able.test.harness in precompile package list")
	}
	if !containsPath(packages, "able.test.reporters") {
		t.Fatalf("expected able.test.reporters in precompile package list")
	}
	if !containsPath(packages, "able.numbers.bigint") {
		t.Fatalf("expected able.numbers.bigint in precompile package list")
	}
	if !containsPath(packages, "able.numbers.biguint") {
		t.Fatalf("expected able.numbers.biguint in precompile package list")
	}
	if !containsPath(packages, "able.numbers.int128") {
		t.Fatalf("expected able.numbers.int128 in precompile package list")
	}
	if !containsPath(packages, "able.numbers.uint128") {
		t.Fatalf("expected able.numbers.uint128 in precompile package list")
	}
	if !containsPath(packages, "able.numbers.rational") {
		t.Fatalf("expected able.numbers.rational in precompile package list")
	}
	if !containsPath(packages, "able.numbers.primitives") {
		t.Fatalf("expected able.numbers.primitives in precompile package list")
	}
	if !containsPath(packages, "able.kernel") {
		t.Fatalf("expected able.kernel in precompile package list")
	}
}

func TestParseBuildArgumentsPrecompileStdlibFlagOverridesEnv(t *testing.T) {
	t.Setenv("ABLE_BUILD_PRECOMPILE_STDLIB", "true")
	config, _, err := parseBuildArguments([]string{"--no-precompile-stdlib", "main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if config.PrecompileStdlib {
		t.Fatalf("expected --no-precompile-stdlib to disable precompile")
	}
}
