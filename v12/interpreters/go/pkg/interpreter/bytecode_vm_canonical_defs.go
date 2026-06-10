package interpreter

import (
	"math"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func isCanonicalAbleStdlibOrigin(origin string, relative string) bool {
	if origin == "" || relative == "" {
		return false
	}
	origin = filepath.ToSlash(origin)
	relative = strings.TrimPrefix(filepath.ToSlash(relative), "/")
	return hasCanonicalPathSuffix(origin, "/able-stdlib/src/", relative) ||
		hasCanonicalPathSuffix(origin, "/pkg/src/", relative) ||
		hasCanonicalVersionedStdlibPath(origin, relative)
}

func hasCanonicalPathSuffix(origin string, base string, relative string) bool {
	if relative == "" || !strings.HasSuffix(origin, relative) {
		return false
	}
	prefixLen := len(origin) - len(relative)
	return prefixLen >= len(base) && strings.HasSuffix(origin[:prefixLen], base)
}

func hasCanonicalVersionedStdlibPath(origin string, relative string) bool {
	if relative == "" || !strings.HasSuffix(origin, relative) {
		return false
	}
	prefix := origin[:len(origin)-len(relative)]
	marker := "/pkg/src/able/"
	idx := strings.LastIndex(prefix, marker)
	if idx < 0 {
		return false
	}
	versionPart := strings.TrimSuffix(prefix[idx+len(marker):], "/src/")
	return versionPart != "" && !strings.Contains(versionPart, "/")
}

func isCanonicalAbleKernelOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	origin = filepath.ToSlash(origin)
	return strings.HasSuffix(origin, "/v12/kernel/src/kernel.able") ||
		strings.HasSuffix(origin, "/kernel/src/kernel.able")
}

func (vm *bytecodeVM) bytecodeCanonicalStringBuilderBuffer(value runtime.Value) (*runtime.ArrayValue, bool) {
	inst, ok := vm.bytecodeCanonicalStringBuilderInstance(value)
	if !ok {
		return nil, false
	}
	bufferValue, ok := structNamedFieldValue(inst, "buffer")
	if !ok {
		return nil, false
	}
	buffer, ok := bufferValue.(*runtime.ArrayValue)
	return buffer, ok && buffer != nil
}

func (vm *bytecodeVM) bytecodeCanonicalStringBuilderInstance(value runtime.Value) (*runtime.StructInstanceValue, bool) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil || vm == nil {
		return nil, false
	}
	if def, ok := vm.canonicalStringBuilderDefinition(); ok && inst.Definition == def {
		return inst, true
	}
	if vm.isCanonicalStructDefinition(inst.Definition, "StringBuilder", "text/string.able") {
		return inst, true
	}
	return nil, false
}

func (vm *bytecodeVM) canonicalStringBuilderDefinition() (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringBuilderDefSet {
		return vm.stringBuilderDef, vm.stringBuilderDef != nil
	}
	vm.stringBuilderDefSet = true
	def, ok := vm.lookupCanonicalStructDefinition("StringBuilder", "text/string.able")
	if ok {
		vm.stringBuilderDef = def
	}
	return def, ok
}

func (vm *bytecodeVM) canonicalStringStructValue(bytes *runtime.ArrayValue, length int) (runtime.Value, bool) {
	if vm == nil || bytes == nil || length < 0 || length > math.MaxInt32 {
		return nil, false
	}
	def, ok := vm.canonicalStringDefinition()
	if !ok || def == nil || def.Node == nil {
		return nil, false
	}
	lengthValue := boxedOrSmallIntegerValue(runtime.IntegerI32, int64(length))
	if def.Node.Kind == ast.StructKindPositional {
		return &runtime.StructInstanceValue{
			Definition: def,
			Positional: []runtime.Value{bytes, lengthValue},
		}, true
	}
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields: map[string]runtime.Value{
			"bytes":     bytes,
			"len_bytes": lengthValue,
		},
	}, true
}

func (vm *bytecodeVM) canonicalStringDefinition() (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringDefSet {
		return vm.stringDef, vm.stringDef != nil
	}
	vm.stringDefSet = true
	def, ok := vm.lookupCanonicalStructDefinition("String", "text/string.able")
	if ok {
		vm.stringDef = def
	}
	return def, ok
}

func (vm *bytecodeVM) canonicalStringBytesIteratorInterfaceValue(iter *runtime.StructInstanceValue) (runtime.Value, bool, error) {
	if iter == nil {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringBytesIteratorDefinition()
	if !ok || iter.Definition != iterDef {
		return nil, false, nil
	}
	ifaceDef, ok := vm.canonicalIteratorInterfaceDefinition()
	if !ok {
		return nil, false, nil
	}
	nextMethod, ok, err := vm.canonicalStringBytesIteratorNextMethod()
	if err != nil || !ok {
		return nil, ok, err
	}
	return &runtime.InterfaceValue{
		Interface:  ifaceDef,
		Underlying: iter,
		Methods: map[string]runtime.Value{
			"next":     nextMethod,
			"iterator": bytecodeIteratorSelfNativeMethod(),
		},
		InterfaceArgs: bytecodeStringBytesIteratorTypeArgs,
	}, true, nil
}

func (vm *bytecodeVM) canonicalIteratorInterfaceDefinition() (*runtime.InterfaceDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringBytesIteratorInterfaceDefSet {
		return vm.stringBytesIteratorInterfaceDef, vm.stringBytesIteratorInterfaceDef != nil
	}
	vm.stringBytesIteratorInterfaceDefSet = true
	def, ok := vm.interp.interfaces["Iterator"]
	if !ok || def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != "Iterator" {
		return nil, false
	}
	if !isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def.Node], "core/iteration.able") {
		return nil, false
	}
	vm.stringBytesIteratorInterfaceDef = def
	return def, true
}

func (vm *bytecodeVM) canonicalStringBytesIteratorNextMethod() (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	version := vm.bytecodeMethodCacheVersion()
	globalRev := vm.bytecodeGlobalRevision()
	if vm.stringBytesIteratorNextSet &&
		vm.stringBytesIteratorNextVersion == version &&
		vm.stringBytesIteratorNextGlobalRev == globalRev {
		return vm.stringBytesIteratorNextMethod, vm.stringBytesIteratorNextMethod != nil, nil
	}
	method, err := vm.interp.findMethod(
		typeInfo{name: "RawStringBytesIter"},
		"next",
		"Iterator",
		bytecodeStringBytesIteratorTypeArgs,
	)
	if err != nil {
		return nil, false, err
	}
	if method == nil {
		vm.stringBytesIteratorNextMethod = nil
		vm.stringBytesIteratorNextVersion = version
		vm.stringBytesIteratorNextGlobalRev = globalRev
		vm.stringBytesIteratorNextSet = true
		return nil, false, nil
	}
	fn := firstFunction(method)
	if fn == nil {
		return nil, false, nil
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "next" ||
		!isStringByteIteratorNextReturnType(def.ReturnType) ||
		!isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def], "text/string.able") {
		return nil, false, nil
	}
	vm.stringBytesIteratorNextMethod = method
	vm.stringBytesIteratorNextVersion = version
	vm.stringBytesIteratorNextGlobalRev = globalRev
	vm.stringBytesIteratorNextSet = true
	return method, true, nil
}

func (vm *bytecodeVM) canonicalStringBytesIteratorDefinition() (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringBytesIterDefSet {
		return vm.stringBytesIterDef, vm.stringBytesIterDef != nil
	}
	vm.stringBytesIterDefSet = true
	def, ok := vm.lookupCanonicalStructDefinition("RawStringBytesIter", "text/string.able")
	if ok {
		vm.stringBytesIterDef = def
	}
	return def, ok
}

func (vm *bytecodeVM) lookupCanonicalStructDefinition(name string, originRelative string) (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil || name == "" || originRelative == "" {
		return nil, false
	}
	if def, ok := vm.interp.lookupStructDefinition(name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
		return def, true
	}
	seen := make(map[string]struct{}, len(vm.interp.packageEnvs)+len(vm.interp.dynamicPackageEnvs)+len(vm.interp.packageRegistry))
	for pkgName := range vm.interp.packageEnvs {
		seen[pkgName] = struct{}{}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	for pkgName := range vm.interp.dynamicPackageEnvs {
		if _, ok := seen[pkgName]; ok {
			continue
		}
		seen[pkgName] = struct{}{}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	for pkgName := range vm.interp.packageRegistry {
		if _, ok := seen[pkgName]; ok {
			continue
		}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	return nil, false
}

func (vm *bytecodeVM) isCanonicalStructDefinition(def *runtime.StructDefinitionValue, name string, originRelative string) bool {
	if vm == nil || vm.interp == nil || def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != name {
		return false
	}
	return isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def.Node], originRelative)
}
