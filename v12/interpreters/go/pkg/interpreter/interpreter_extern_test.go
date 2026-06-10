package interpreter

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

const externPluginTestChildEnv = "ABLE_EXTERN_PLUGIN_TEST_CHILD"

const externPluginInitialBuildHelperEnv = "ABLE_EXTERN_PLUGIN_INITIAL_BUILD_HELPER"

// runExternPluginTestInChild runs one plugin-loading case in an otherwise
// fresh process. Go plugins are process-global and cannot be unloaded, so this
// keeps independent integration cases from sharing loader state while their
// direct host-boundary behavior remains covered.
func runExternPluginTestInChild(t *testing.T) bool {
	t.Helper()
	if os.Getenv(externPluginTestChildEnv) == t.Name() {
		return false
	}
	cmd := exec.Command(os.Args[0], "-test.run=^"+regexp.QuoteMeta(t.Name())+"$")
	cmd.Env = append(os.Environ(), externPluginTestChildEnv+"="+t.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run isolated extern plugin test: %v\n%s", err, output)
	}
	return true
}

func TestExternHostCacheRootUsesConfiguredDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(externCacheDirEnv, dir)
	if got := externHostCacheRoot(); got != dir {
		t.Fatalf("extern cache root = %q, want %q", got, dir)
	}
}

func TestExternHostCacheRootNormalizesRelativeDirectory(t *testing.T) {
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	defer os.Chdir(previousDir)
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Setenv(externCacheDirEnv, "relative-extern-cache")
	if got, want := externHostCacheRoot(), filepath.Join(workingDir, "relative-extern-cache"); got != want {
		t.Fatalf("extern cache root = %q, want %q", got, want)
	}
}

func TestPrewarmExternHostModulesBuildsWithoutEvaluatingModule(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	cacheDir := t.TempDir()
	t.Setenv(externCacheDirEnv, cacheDir)

	extern := ast.Extern(
		ast.HostTargetGo,
		ast.Fn("cached_value", nil, nil, ast.Ty("i32"), nil, nil, false, false),
		"return int32(7)",
	)
	module := ast.Mod([]ast.Statement{extern}, nil, ast.Pkg([]interface{}{"sample", "host"}, false))
	program := &driver.Program{Modules: []*driver.Module{{
		Package: "sample.host",
		AST:     module,
	}}}

	result, err := New().PrewarmExternHostModules(program)
	if err != nil {
		t.Fatalf("PrewarmExternHostModules: %v", err)
	}
	if result.Modules != 1 {
		t.Fatalf("prewarmed modules = %d, want 1", result.Modules)
	}
	foundPlugin := false
	err = filepath.WalkDir(cacheDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Ext(path) == ".so" {
			foundPlugin = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk extern cache: %v", err)
	}
	if !foundPlugin {
		t.Fatalf("prewarm did not build a plugin under %s", cacheDir)
	}
}

func TestProgramExternHostImageSharesOnePluginAcrossPackages(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	cacheDir := t.TempDir()
	t.Setenv(externCacheDirEnv, cacheDir)

	left := &driver.Module{
		Package: "sample.left",
		AST: ast.Mod([]ast.Statement{
			ast.Prelude(ast.HostTargetGo, `import "strings"`),
			ast.Extern(ast.HostTargetGo, ast.Fn("value", nil, nil, ast.Ty("i32"), nil, nil, false, false), `return int32(len(strings.Repeat("x", 17)))`),
		}, nil, ast.Pkg([]interface{}{"sample", "left"}, false)),
	}
	right := &driver.Module{
		Package: "sample.right",
		AST: ast.Mod([]ast.Statement{
			ast.Extern(ast.HostTargetGo, ast.Fn("value", nil, nil, ast.Ty("i32"), nil, nil, false, false), "return int32(29)"),
		}, nil, ast.Pkg([]interface{}{"sample", "right"}, false)),
	}
	program := &driver.Program{Entry: right, Modules: []*driver.Module{left, right}}
	interp := New()
	result, err := interp.PrewarmExternHostModules(program)
	if err != nil {
		t.Fatalf("PrewarmExternHostModules: %v", err)
	}
	if result.Modules != 1 {
		t.Fatalf("prewarmed host images = %d, want 1", result.Modules)
	}

	leftModule := interp.externHostPackages["sample.left"].modules[ast.HostTargetGo]
	rightModule := interp.externHostPackages["sample.right"].modules[ast.HostTargetGo]
	if leftModule == nil || rightModule == nil || leftModule.plugin != rightModule.plugin {
		t.Fatal("expected both packages to share one host image plugin")
	}

	_, leftEnv, err := interp.EvaluateModule(left.AST)
	if err != nil {
		t.Fatalf("evaluate left module: %v", err)
	}
	leftValue, err := interp.evaluateExpression(ast.Call("value"), leftEnv)
	if err != nil {
		t.Fatalf("call left extern: %v", err)
	}
	_, rightEnv, err := interp.EvaluateModule(right.AST)
	if err != nil {
		t.Fatalf("evaluate right module: %v", err)
	}
	rightValue, err := interp.evaluateExpression(ast.Call("value"), rightEnv)
	if err != nil {
		t.Fatalf("call right extern: %v", err)
	}
	for name, value := range map[string]runtime.Value{"left": leftValue, "right": rightValue} {
		integer, ok := value.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("%s extern value type = %T, want IntegerValue", name, value)
		}
		want := int64(17)
		if name == "right" {
			want = 29
		}
		if integer.Int64Fast() != want {
			t.Fatalf("%s extern value = %d, want %d", name, integer.Int64Fast(), want)
		}
	}
}

func TestExternHostPluginRebuildsInvalidArtifact(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	cacheDir := os.Getenv("ABLE_EXTERN_PLUGIN_REBUILD_CACHE")
	if cacheDir == "" {
		cacheDir = t.TempDir()
		t.Setenv("ABLE_EXTERN_PLUGIN_REBUILD_CACHE", cacheDir)
	}
	t.Setenv(externCacheDirEnv, cacheDir)

	state := &externTargetState{
		externs: []*ast.ExternFunctionBody{ast.Extern(
			ast.HostTargetGo,
			ast.Fn("rebuild_cached_value", nil, nil, ast.Ty("i32"), nil, nil, false, false),
			"return int32(9)",
		)},
		externByID: map[string]int{"rebuild_cached_value": 0},
	}
	hash := cachedExternStateHash(ast.HostTargetGo, state, externHostCacheScope())
	if os.Getenv(externPluginInitialBuildHelperEnv) == "1" {
		if _, err := buildExternModule("rebuild.host", ast.HostTargetGo, state, hash); err != nil {
			t.Fatalf("build initial plugin: %v", err)
		}
		return
	}
	if os.Getenv("ABLE_EXTERN_PLUGIN_REBUILD_HELPER") == "1" {
		if _, err := buildExternModule("rebuild.host", ast.HostTargetGo, state, hash); err != nil {
			t.Fatalf("rebuild invalid plugin: %v", err)
		}
		return
	}

	initialCmd := exec.Command(os.Args[0], "-test.run=^TestExternHostPluginRebuildsInvalidArtifact$")
	initialCmd.Env = append(os.Environ(), externPluginInitialBuildHelperEnv+"=1")
	if output, err := initialCmd.CombinedOutput(); err != nil {
		t.Fatalf("run initial-plugin build helper: %v\n%s", err, output)
	}
	pluginPath := filepath.Join(cacheDir, "rebuild_host", hash, "extern.so")
	if err := os.WriteFile(pluginPath, []byte("not a Go plugin"), 0o644); err != nil {
		t.Fatalf("corrupt plugin artifact: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestExternHostPluginRebuildsInvalidArtifact$")
	cmd.Env = append(os.Environ(), "ABLE_EXTERN_PLUGIN_REBUILD_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run invalid-plugin rebuild helper: %v\n%s", err, output)
	}
	marker, err := os.ReadFile(filepath.Join(cacheDir, "rebuild_host", hash, "plugin-artifact"))
	if err != nil {
		t.Fatalf("read rebuilt plugin marker: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(marker)), "extern-rebuilt-") {
		t.Fatalf("rebuilt plugin marker = %q, want extern-rebuilt-*", marker)
	}
}

func TestExternHandlersRegisterNativeFunctions(t *testing.T) {
	interp := New()
	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, ast.Pkg([]interface{}{"host"}, false))

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if _, ok := val.(*runtime.NativeFunctionValue); !ok {
		t.Fatalf("expected native function, got %#v", val)
	}
	native, ok := val.(*runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("expected native function pointer, got %#v", val)
	}
	if !native.BorrowArgs {
		t.Fatalf("expected extern native wrapper to borrow args")
	}
	if !native.SkipContext {
		t.Fatalf("expected extern native wrapper to skip native call context setup")
	}
	bucket := interp.packageRegistry["host"]
	if bucket == nil || bucket["now_nanos"] == nil {
		t.Fatalf("package registry missing now_nanos")
	}
}

func TestExternIgnoresNonGoTargets(t *testing.T) {
	interp := New()
	sig := ast.Fn("ts_only", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetTypeScript, sig, "")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	if _, err := env.Get("ts_only"); err == nil {
		t.Fatalf("expected ts_only to be undefined for go target")
	}
}

func TestExternPreservesExistingBinding(t *testing.T) {
	interp := New()
	existing := &runtime.NativeFunctionValue{Name: "now_nanos", Arity: 0}
	interp.global.Define("now_nanos", existing)

	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if val != existing {
		t.Fatalf("expected existing binding to remain")
	}
}

func TestExternStructArrayFieldCoercesIntoHostMap(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()

	specDef := ast.StructDef("Spec", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Gen(ast.Ty("Array"), ast.Ty("String")), "args"),
	}, ast.StructKindNamed, nil, nil, false)
	sig := ast.Fn(
		"accept_spec",
		[]*ast.FunctionParameter{ast.Param("spec", ast.Ty("Spec"))},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	mod := ast.Mod([]ast.Statement{
		specDef,
		ast.Extern(ast.HostTargetGo, sig, `
specMap, ok := spec.(map[string]any)
if !ok {
	return int32(-1)
}
args, ok := specMap["args"].([]any)
if !ok {
	return int32(-2)
}
return int32(len(args))
`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("accept_spec",
		ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Arr(ast.Str("a"), ast.Str("b")), "args"),
		}, false, "Spec", nil, nil),
	), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}

	iv, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if iv.BigInt().Int64() != 2 {
		t.Fatalf("expected extern to receive two args, got %s", iv.BigInt().String())
	}
}

func TestExternStructNullableArrayFieldCoercesIntoHostMap(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()

	specDef := ast.StructDef("Spec", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Nullable(ast.Gen(ast.Ty("Array"), ast.Ty("String"))), "args"),
	}, ast.StructKindNamed, nil, nil, false)
	sig := ast.Fn(
		"accept_spec",
		[]*ast.FunctionParameter{ast.Param("spec", ast.Ty("Spec"))},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	mod := ast.Mod([]ast.Statement{
		specDef,
		ast.Extern(ast.HostTargetGo, sig, `
specMap, ok := spec.(map[string]any)
if !ok {
	return int32(-1)
}
raw, ok := specMap["args"]
if !ok {
	return int32(-2)
}
if raw == nil {
	return int32(-3)
}
args, ok := raw.([]any)
if !ok {
	return int32(-4)
}
return int32(len(args))
`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("accept_spec",
		ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Arr(ast.Str("x"), ast.Str("y"), ast.Str("z")), "args"),
		}, false, "Spec", nil, nil),
	), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}

	iv, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if iv.BigInt().Int64() != 3 {
		t.Fatalf("expected extern to receive three args, got %s", iv.BigInt().String())
	}
}

func TestExternTargetHashCacheReusesAndInvalidates(t *testing.T) {
	interp := New()
	modA := ast.Mod([]ast.Statement{
		ast.Extern(ast.HostTargetGo, ast.Fn("fast", nil, nil, ast.Ty("i32"), nil, nil, false, false), "return int32(1)"),
	}, nil, ast.Pkg([]interface{}{"host"}, false))

	if _, _, err := interp.EvaluateModule(modA); err != nil {
		t.Fatalf("evaluate module A: %v", err)
	}

	pkg := interp.externHostPackages["host"]
	if pkg == nil {
		t.Fatalf("expected extern host package")
	}
	state := pkg.targets[ast.HostTargetGo]
	if state == nil {
		t.Fatalf("expected go extern target state")
	}
	if state.hashValid {
		t.Fatalf("hash should start invalid before first cached lookup")
	}

	firstHash := cachedExternStateHash(ast.HostTargetGo, state, externHostCacheScope())
	if firstHash == "" {
		t.Fatalf("expected cached hash")
	}
	if !state.hashValid {
		t.Fatalf("hash should be valid after caching")
	}
	if got := cachedExternStateHash(ast.HostTargetGo, state, externHostCacheScope()); got != firstHash {
		t.Fatalf("cached hash mismatch: got %q want %q", got, firstHash)
	}

	modB := ast.Mod([]ast.Statement{
		ast.Extern(ast.HostTargetGo, ast.Fn("fast", nil, nil, ast.Ty("i32"), nil, nil, false, false), "return int32(2)"),
	}, nil, ast.Pkg([]interface{}{"host"}, false))
	interp.currentPackage = "host"
	interp.registerExternStatements(modB)
	if state.hashValid {
		t.Fatalf("hash should be invalidated after extern re-registration")
	}
	secondHash := cachedExternStateHash(ast.HostTargetGo, state, externHostCacheScope())
	if secondHash == "" {
		t.Fatalf("expected recomputed hash")
	}
	if secondHash == firstHash {
		t.Fatalf("expected hash to change after extern body update")
	}
}

func TestExternModuleBuildsFastInvokerForHotStringSignatures(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()
	replaceSig := ast.Fn(
		"replace_like",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("String")),
			ast.Param("old", ast.Ty("String")),
			ast.Param("new", ast.Ty("String")),
		},
		nil,
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	splitSig := ast.Fn(
		"split_like",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("String"))},
		nil,
		ast.Gen(ast.Ty("Array"), ast.Ty("String")),
		nil,
		nil,
		false,
		false,
	)
	replaceExtern := ast.Extern(ast.HostTargetGo, replaceSig, `return strings.ReplaceAll(value, old, new)`)
	splitExtern := ast.Extern(ast.HostTargetGo, splitSig, `return []string{value, value}`)
	mod := ast.Mod([]ast.Statement{
		ast.Prelude(ast.HostTargetGo, `import "strings"`),
		replaceExtern,
		splitExtern,
	}, nil, ast.Pkg([]interface{}{"host"}, false))

	if _, _, err := interp.EvaluateModule(mod); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	pkg := interp.externHostPackages["host"]
	if pkg == nil {
		t.Fatalf("expected extern host package")
	}
	state := pkg.targets[ast.HostTargetGo]
	if state == nil {
		t.Fatalf("expected go extern target state")
	}
	module, err := interp.ensureExternHostModule("host", ast.HostTargetGo, state, pkg)
	if err != nil {
		t.Fatalf("ensure extern host module: %v", err)
	}

	replaceInvoker, err := module.lookupInvoker(replaceExtern)
	if err != nil {
		t.Fatalf("lookup replace invoker: %v", err)
	}
	if replaceInvoker == nil {
		t.Fatalf("expected fast invoker for replace_like")
	}
	gotReplace, err := replaceInvoker(interp, []runtime.Value{
		runtime.StringValue{Val: "ceiling"},
		runtime.StringValue{Val: "cei"},
		runtime.StringValue{Val: ""},
	})
	if err != nil {
		t.Fatalf("run replace invoker: %v", err)
	}
	if str, ok := gotReplace.(runtime.StringValue); !ok || str.Val != "ling" {
		t.Fatalf("unexpected replace invoker result %#v", gotReplace)
	}

	splitInvoker, err := module.lookupInvoker(splitExtern)
	if err != nil {
		t.Fatalf("lookup split invoker: %v", err)
	}
	if splitInvoker == nil {
		t.Fatalf("expected fast invoker for split_like")
	}
	gotSplit, err := splitInvoker(interp, []runtime.Value{runtime.StringValue{Val: "x"}})
	if err != nil {
		t.Fatalf("run split invoker: %v", err)
	}
	arr, ok := gotSplit.(*runtime.ArrayValue)
	if !ok || len(arr.Elements) != 2 {
		t.Fatalf("unexpected split invoker result %#v", gotSplit)
	}
	first, ok := arr.Elements[0].(runtime.StringValue)
	if !ok || first.Val != "x" {
		t.Fatalf("unexpected split first element %#v", arr.Elements[0])
	}
}

func TestExternSmallUnsignedResultsStaySmallInt(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()
	sig := ast.Fn("read_u64", nil, nil, ast.Ty("u64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{
		ast.Extern(ast.HostTargetGo, sig, `return uint64(42)`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("read_u64"), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if !intVal.IsSmall() {
		t.Fatalf("expected small integer result for u64 42")
	}
	if intVal.Int64Fast() != 42 || intVal.TypeSuffix != runtime.IntegerU64 {
		t.Fatalf("unexpected integer result %#v", intVal)
	}
}

func TestExternSmallUnsignedArrayElementsStaySmallInt(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()
	sig := ast.Fn("read_bytes", nil, nil, ast.Gen(ast.Ty("Array"), ast.Ty("u8")), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{
		ast.Extern(ast.HostTargetGo, sig, `return []uint8{1, 2, 3}`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("read_bytes"), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}
	arr, ok := value.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", value)
	}
	if size, err := runtime.ArrayStoreSize(arr.Handle); err != nil {
		t.Fatalf("array size: %v", err)
	} else if size != 3 {
		t.Fatalf("expected 3 stored elements, got %d", size)
	}
	for idx, want := range []int64{1, 2, 3} {
		elem, err := runtime.ArrayStoreRead(arr.Handle, idx)
		if err != nil {
			t.Fatalf("array read %d: %v", idx, err)
		}
		intVal, ok := elem.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("element %d type = %T, want runtime.IntegerValue", idx, elem)
		}
		if !intVal.IsSmall() {
			t.Fatalf("element %d should stay small-int, got %#v", idx, intVal)
		}
		if intVal.TypeSuffix != runtime.IntegerU8 || intVal.Int64Fast() != want {
			t.Fatalf("unexpected element %d value %#v", idx, intVal)
		}
	}
	elements, err := interp.ArrayElements(arr)
	if err != nil {
		t.Fatalf("materialize array elements: %v", err)
	}
	if len(elements) != 3 {
		t.Fatalf("expected 3 materialized elements, got %d", len(elements))
	}
}

func TestExternLargeU64StillFallsBackToBigInt(t *testing.T) {
	if runExternPluginTestInChild(t) {
		return
	}
	interp := New()
	sig := ast.Fn("read_large_u64", nil, nil, ast.Ty("u64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{
		ast.Extern(ast.HostTargetGo, sig, `return uint64(1) << 63`),
	}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("read_large_u64"), env)
	if err != nil {
		t.Fatalf("call extern: %v", err)
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", value)
	}
	if intVal.IsSmall() {
		t.Fatalf("expected large u64 to fall back to big integer")
	}
	if got := intVal.BigInt().Uint64(); got != uint64(math.MaxInt64)+1 {
		t.Fatalf("unexpected large u64 value %d", got)
	}
}
