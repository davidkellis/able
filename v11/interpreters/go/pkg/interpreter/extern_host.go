package interpreter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type externTargetState struct {
	preludes   []string
	externs    []*ast.ExternFunctionBody
	externByID map[string]int
}

type externHostPackage struct {
	targets map[ast.HostTarget]*externTargetState
	modules map[ast.HostTarget]*externHostModule
}

type externHostModule struct {
	hash    string
	plugin  *plugin.Plugin
	symbols map[string]reflect.Value
}

const externCacheVersion = "v2"

func (i *Interpreter) isKernelExtern(name string) bool {
	return strings.HasPrefix(name, "__able_")
}

func (i *Interpreter) registerExternStatements(module *ast.Module) {
	if module == nil {
		return
	}
	pkgName := i.currentPackage
	if pkgName == "" {
		pkgName = "<root>"
	}
	if i.externHostPackages == nil {
		i.externHostPackages = make(map[string]*externHostPackage)
	}
	pkg := i.externHostPackages[pkgName]
	if pkg == nil {
		pkg = &externHostPackage{
			targets: make(map[ast.HostTarget]*externTargetState),
			modules: make(map[ast.HostTarget]*externHostModule),
		}
		i.externHostPackages[pkgName] = pkg
	}
	for _, stmt := range module.Body {
		switch s := stmt.(type) {
		case *ast.PreludeStatement:
			if s == nil {
				continue
			}
			state := ensureExternTarget(pkg, s.Target)
			state.preludes = append(state.preludes, s.Code)
		case *ast.ExternFunctionBody:
			if s == nil || s.Signature == nil || s.Signature.ID == nil {
				continue
			}
			name := s.Signature.ID.Name
			if name == "" {
				continue
			}
			state := ensureExternTarget(pkg, s.Target)
			if idx, ok := state.externByID[name]; ok {
				state.externs[idx] = s
			} else {
				state.externByID[name] = len(state.externs)
				state.externs = append(state.externs, s)
			}
		}
	}
}

func ensureExternTarget(pkg *externHostPackage, target ast.HostTarget) *externTargetState {
	if pkg.targets == nil {
		pkg.targets = make(map[ast.HostTarget]*externTargetState)
	}
	state := pkg.targets[target]
	if state == nil {
		state = &externTargetState{externByID: make(map[string]int)}
		pkg.targets[target] = state
	}
	return state
}

func (i *Interpreter) invokeExternHostFunction(pkgName string, def *ast.ExternFunctionBody, args []runtime.Value) (runtime.Value, error) {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return runtime.NilValue{}, nil
	}
	if pkgName == "" {
		pkgName = "<root>"
	}
	pkg := i.externHostPackages[pkgName]
	if pkg == nil {
		return nil, fmt.Errorf("extern package %s is not registered", pkgName)
	}
	targetState := pkg.targets[def.Target]
	if targetState == nil {
		return nil, fmt.Errorf("extern target %s is not registered", def.Target)
	}
	module, err := i.ensureExternHostModule(pkgName, def.Target, targetState, pkg)
	if err != nil {
		return nil, err
	}
	fn, err := module.lookup(def.Signature.ID.Name)
	if err != nil {
		return nil, err
	}
	fnType := fn.Type()
	paramCount := fnType.NumIn()
	if len(args) != paramCount {
		return nil, fmt.Errorf("extern function %s expects %d args, got %d", def.Signature.ID.Name, paramCount, len(args))
	}
	callArgs := make([]reflect.Value, paramCount)
	for idx := 0; idx < paramCount; idx++ {
		hostVal, convErr := i.toHostValue(def.Signature.Params[idx].ParamType, args[idx], fnType.In(idx))
		if convErr != nil {
			return nil, convErr
		}
		callArgs[idx] = hostVal
	}

	var results []reflect.Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("extern panic: %v", r)
			}
		}()
		if err == nil {
			results = fn.Call(callArgs)
		}
	}()
	if err != nil {
		return nil, err
	}
	return i.fromHostResults(def, results)
}

func (i *Interpreter) ensureExternHostModule(pkgName string, target ast.HostTarget, state *externTargetState, pkg *externHostPackage) (*externHostModule, error) {
	hash := hashExternState(target, state)
	if existing := pkg.modules[target]; existing != nil && existing.hash == hash && existing.plugin != nil {
		return existing, nil
	}
	module, err := buildExternModule(pkgName, target, state, hash)
	if err != nil {
		return nil, err
	}
	pkg.modules[target] = module
	return module, nil
}

func hashExternState(target ast.HostTarget, state *externTargetState) string {
	hasher := sha256.New()
	hasher.Write([]byte("target:"))
	hasher.Write([]byte(target))
	hasher.Write([]byte("\nversion:"))
	hasher.Write([]byte(externCacheVersion))
	hasher.Write([]byte("\n"))
	for _, prelude := range state.preludes {
		hasher.Write([]byte("prelude:"))
		hasher.Write([]byte(prelude))
		hasher.Write([]byte("\n"))
	}
	for _, extern := range state.externs {
		if extern == nil || extern.Signature == nil || extern.Signature.ID == nil {
			continue
		}
		hasher.Write([]byte("extern:"))
		hasher.Write([]byte(externSignatureKey(extern)))
		hasher.Write([]byte("\n"))
		hasher.Write([]byte(extern.Body))
		hasher.Write([]byte("\n"))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func externSignatureKey(extern *ast.ExternFunctionBody) string {
	if extern == nil || extern.Signature == nil || extern.Signature.ID == nil {
		return "<missing>"
	}
	params := make([]string, len(extern.Signature.Params))
	for idx, param := range extern.Signature.Params {
		if param == nil {
			params[idx] = "_"
			continue
		}
		params[idx] = typeKey(param.ParamType)
	}
	return fmt.Sprintf("%s(%s)->%s", extern.Signature.ID.Name, strings.Join(params, ","), typeKey(extern.Signature.ReturnType))
}

func typeKey(expr ast.TypeExpression) string {
	if expr == nil {
		return "void"
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "_"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		args := make([]string, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = typeKey(arg)
		}
		return fmt.Sprintf("%s<%s>", typeKey(t.Base), strings.Join(args, ","))
	case *ast.NullableTypeExpression:
		return "?" + typeKey(t.InnerType)
	case *ast.ResultTypeExpression:
		return "!" + typeKey(t.InnerType)
	case *ast.UnionTypeExpression:
		members := make([]string, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = typeKey(member)
		}
		return strings.Join(members, "|")
	case *ast.FunctionTypeExpression:
		params := make([]string, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = typeKey(param)
		}
		return fmt.Sprintf("(%s)->%s", strings.Join(params, ","), typeKey(t.ReturnType))
	case *ast.WildcardTypeExpression:
		return "_"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func buildExternModule(pkgName string, target ast.HostTarget, state *externTargetState, hash string) (*externHostModule, error) {
	cacheDir := filepath.Join(os.TempDir(), "able-v11-extern-go", sanitizePackageName(pkgName), hash)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("extern cache mkdir: %w", err)
	}
	sourcePath := filepath.Join(cacheDir, "extern.go")
	pluginPath := filepath.Join(cacheDir, "extern.so")

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		source, err := renderGoHostModule(state)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
			return nil, fmt.Errorf("extern cache write: %w", err)
		}
	}

	modPath := filepath.Join(cacheDir, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		moduleLine := fmt.Sprintf("module able_extern_%s\n\ngo 1.22\n", hash)
		if err := os.WriteFile(modPath, []byte(moduleLine), 0o644); err != nil {
			return nil, fmt.Errorf("extern cache write go.mod: %w", err)
		}
	}

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, sourcePath)
		cmd.Dir = cacheDir
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("extern host build failed: %w\n%s", err, output.String())
		}
	}

	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("extern host plugin open: %w", err)
	}
	return &externHostModule{
		hash:    hash,
		plugin:  plug,
		symbols: make(map[string]reflect.Value),
	}, nil
}

func renderGoHostModule(state *externTargetState) (string, error) {
	var builder strings.Builder
	builder.WriteString("package main\n\n")
	if state != nil {
		for _, prelude := range state.preludes {
			builder.WriteString(prelude)
			builder.WriteString("\n")
		}
	}
	if needsBigInt(state) {
		builder.WriteString("import \"math/big\"\n\n")
	}
	builder.WriteString("type IoHandle = interface{}\n")
	builder.WriteString("type ProcHandle = interface{}\n")
	builder.WriteString("type hostError struct{ message string }\n")
	builder.WriteString("func (e hostError) Error() string { return e.message }\n")
	builder.WriteString("func host_error[T any](message string) (T, error) { var zero T; return zero, hostError{message} }\n\n")

	for _, extern := range state.externs {
		if extern == nil || extern.Signature == nil || extern.Signature.ID == nil {
			continue
		}
		fn, err := renderGoExternFunction(extern)
		if err != nil {
			return "", err
		}
		builder.WriteString(fn)
		builder.WriteString("\n\n")
	}
	return builder.String(), nil
}

func renderGoExternFunction(extern *ast.ExternFunctionBody) (string, error) {
	name := extern.Signature.ID.Name
	params := make([]string, 0, len(extern.Signature.Params))
	argNames := make([]string, 0, len(extern.Signature.Params))
	for idx, param := range extern.Signature.Params {
		paramName := externParamName(param, idx)
		typ, err := goTypeForExpr(param.ParamType)
		if err != nil {
			return "", err
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, typ))
		argNames = append(argNames, paramName)
	}
	retType, err := goReturnTypeForExpr(extern.Signature.ReturnType)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("func ")
	builder.WriteString(name)
	builder.WriteString("(")
	builder.WriteString(strings.Join(params, ", "))
	builder.WriteString(")")
	if retType != "" {
		builder.WriteString(" ")
		builder.WriteString(retType)
	}
	builder.WriteString(" {\n")
	body := strings.TrimSpace(extern.Body)
	if body != "" {
		for _, line := range strings.Split(body, "\n") {
			builder.WriteString("\t")
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("}")
	builder.WriteString("\n\n")
	builder.WriteString("func ")
	builder.WriteString(externSymbolName(name))
	builder.WriteString("(")
	builder.WriteString(strings.Join(params, ", "))
	builder.WriteString(")")
	if retType != "" {
		builder.WriteString(" ")
		builder.WriteString(retType)
	}
	builder.WriteString(" {\n\t")
	if retType != "" {
		builder.WriteString("return ")
	}
	builder.WriteString(name)
	builder.WriteString("(")
	builder.WriteString(strings.Join(argNames, ", "))
	builder.WriteString(")\n}")
	return builder.String(), nil
}

func externParamName(param *ast.FunctionParameter, idx int) string {
	if param == nil {
		return fmt.Sprintf("arg%d", idx)
	}
	if id, ok := param.Name.(*ast.Identifier); ok && id != nil {
		return id.Name
	}
	return fmt.Sprintf("arg%d", idx)
}

func goReturnTypeForExpr(expr ast.TypeExpression) (string, error) {
	if expr == nil {
		return "", nil
	}
	switch t := expr.(type) {
	case *ast.ResultTypeExpression:
		inner, err := goTypeForExpr(t.InnerType)
		if err != nil {
			return "", err
		}
		if inner == "" {
			inner = "struct{}"
		}
		return fmt.Sprintf("(%s, error)", inner), nil
	default:
		return goTypeForExpr(expr)
	}
}

func goTypeForExpr(expr ast.TypeExpression) (string, error) {
	if expr == nil {
		return "", nil
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		switch name {
		case "String":
			return "string", nil
		case "bool":
			return "bool", nil
		case "char":
			return "rune", nil
		case "void":
			return "", nil
		case "IoHandle", "ProcHandle":
			return name, nil
		case "i8":
			return "int8", nil
		case "i16":
			return "int16", nil
		case "i32":
			return "int32", nil
		case "i64":
			return "int64", nil
		case "u8":
			return "uint8", nil
		case "u16":
			return "uint16", nil
		case "u32":
			return "uint32", nil
		case "u64":
			return "uint64", nil
		case "i128", "u128":
			return "*big.Int", nil
		case "f32":
			return "float32", nil
		case "f64":
			return "float64", nil
		default:
			return "interface{}", nil
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil {
			if normalizeKernelAliasName(base.Name.Name) == "Array" {
				elemType := "interface{}"
				if len(t.Arguments) > 0 {
					if argType, err := goTypeForExpr(t.Arguments[0]); err == nil && argType != "" {
						elemType = argType
					}
				}
				return "[]" + elemType, nil
			}
		}
		return "interface{}", nil
	case *ast.NullableTypeExpression:
		inner, err := goTypeForExpr(t.InnerType)
		if err != nil {
			return "", err
		}
		if inner == "" {
			inner = "struct{}"
		}
		return "*" + inner, nil
	default:
		return "interface{}", nil
	}
}

func needsBigInt(state *externTargetState) bool {
	if state == nil {
		return false
	}
	for _, ext := range state.externs {
		if ext == nil || ext.Signature == nil {
			continue
		}
		if typeUsesBigInt(ext.Signature.ReturnType) {
			return true
		}
		for _, param := range ext.Signature.Params {
			if param != nil && typeUsesBigInt(param.ParamType) {
				return true
			}
		}
	}
	return false
}

func typeUsesBigInt(expr ast.TypeExpression) bool {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return false
		}
		name := normalizeKernelAliasName(t.Name.Name)
		return name == "i128" || name == "u128"
	case *ast.GenericTypeExpression:
		if typeUsesBigInt(t.Base) {
			return true
		}
		for _, arg := range t.Arguments {
			if typeUsesBigInt(arg) {
				return true
			}
		}
	case *ast.NullableTypeExpression:
		return typeUsesBigInt(t.InnerType)
	case *ast.ResultTypeExpression:
		return typeUsesBigInt(t.InnerType)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if typeUsesBigInt(member) {
				return true
			}
		}
	}
	return false
}

func sanitizePackageName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "pkg"
	}
	out := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			out.WriteRune(r)
		} else {
			out.WriteRune('_')
		}
	}
	return out.String()
}

func externSymbolName(name string) string {
	return "AbleExtern_" + sanitizeSymbolName(name)
}

func sanitizeSymbolName(name string) string {
	if name == "" {
		return "fn"
	}
	out := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			out.WriteRune(r)
		} else {
			out.WriteRune('_')
		}
	}
	return out.String()
}

func (m *externHostModule) lookup(name string) (reflect.Value, error) {
	if m == nil || m.plugin == nil {
		return reflect.Value{}, fmt.Errorf("extern host module not initialized")
	}
	if name == "" {
		return reflect.Value{}, fmt.Errorf("extern function name is empty")
	}
	if m.symbols == nil {
		m.symbols = make(map[string]reflect.Value)
	}
	if sym, ok := m.symbols[name]; ok {
		return sym, nil
	}
	symbolName := externSymbolName(name)
	raw, err := m.plugin.Lookup(symbolName)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("extern lookup %s: %w", name, err)
	}
	fn := reflect.ValueOf(raw)
	if fn.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("extern symbol %s is not a function", name)
	}
	m.symbols[name] = fn
	return fn, nil
}

func (i *Interpreter) toHostValue(typeExpr ast.TypeExpression, value runtime.Value, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}
	for {
		switch v := value.(type) {
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v != nil {
				value = v.Underlying
				continue
			}
		}
		break
	}
	hostVal, err := i.coerceRuntimeToHost(typeExpr, value, targetType)
	if err != nil {
		return reflect.Value{}, err
	}
	rv := reflect.ValueOf(hostVal)
	if !rv.IsValid() {
		return reflect.Zero(targetType), nil
	}
	if rv.Type().AssignableTo(targetType) {
		return rv, nil
	}
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType), nil
	}
	return reflect.Value{}, fmt.Errorf("extern argument cannot convert %s to %s", rv.Type(), targetType)
}

func (i *Interpreter) coerceRuntimeToHost(typeExpr ast.TypeExpression, value runtime.Value, targetType reflect.Type) (any, error) {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
		}
	} else {
		return i.coerceRuntimeByKind(value, targetType)
	}
	switch t := typeExpr.(type) {
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			if i.matchesType(member, value) {
				return i.coerceRuntimeToHost(member, value, targetType)
			}
		}
		return i.coerceRuntimeByKind(value, targetType)
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		if targetType.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("extern nullable type expects pointer target")
		}
		elemVal, err := i.coerceRuntimeToHost(t.InnerType, value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		if elemVal != nil {
			ev := reflect.ValueOf(elemVal)
			if ev.IsValid() && ev.Type().ConvertibleTo(targetType.Elem()) {
				ptr.Elem().Set(ev.Convert(targetType.Elem()))
			}
		}
		return ptr.Interface(), nil
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && normalizeKernelAliasName(base.Name.Name) == "Array" {
			arr, err := i.toArrayValue(value)
			if err != nil {
				return nil, err
			}
			elemType := targetType.Elem()
			slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			for idx, elem := range arr.Elements {
				hostElem, err := i.coerceRuntimeToHost(elemExpr, elem, elemType)
				if err != nil {
					return nil, err
				}
				ev := reflect.ValueOf(hostElem)
				if ev.IsValid() && ev.Type().ConvertibleTo(elemType) {
					slice.Index(idx).Set(ev.Convert(elemType))
				}
			}
			return slice.Interface(), nil
		}
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		switch name {
		case "String":
			return i.coerceStringValue(value)
		case "bool":
			if v, ok := value.(runtime.BoolValue); ok {
				return v.Val, nil
			}
		case "char":
			if v, ok := value.(runtime.CharValue); ok {
				return v.Val, nil
			}
		case "IoHandle", "ProcHandle":
			if hv, ok := value.(*runtime.HostHandleValue); ok && hv != nil {
				return hv.Value, nil
			}
			return nil, nil
		case "f32", "f64":
			if f, ok := value.(runtime.FloatValue); ok {
				return f.Val, nil
			}
			if f, ok := value.(*runtime.FloatValue); ok && f != nil {
				return f.Val, nil
			}
			if iv, ok := value.(runtime.IntegerValue); ok {
				return bigIntToFloat(iv.Val), nil
			}
			if iv, ok := value.(*runtime.IntegerValue); ok && iv != nil {
				return bigIntToFloat(iv.Val), nil
			}
		case "i8", "i16", "i32", "i64":
			return coerceIntValue(value, name)
		case "u8", "u16", "u32", "u64":
			return coerceUintValue(value, name)
		case "i128", "u128":
			return coerceBigInt(value)
		}
		if def, ok := i.lookupStructDefinition(name); ok {
			return i.structToHostValue(def, value)
		}
	}
	if targetType.Kind() == reflect.Interface {
		return value, nil
	}
	return nil, fmt.Errorf("unsupported extern argument type")
}

func (i *Interpreter) coerceRuntimeByKind(value runtime.Value, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return i.coerceStringValue(value)
	case reflect.Bool:
		if v, ok := value.(runtime.BoolValue); ok {
			return v.Val, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num, err := toInt64(value)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(targetType).Interface(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num, err := toUint64(value)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(num).Convert(targetType).Interface(), nil
	case reflect.Float32, reflect.Float64:
		if f, ok := value.(runtime.FloatValue); ok {
			return reflect.ValueOf(f.Val).Convert(targetType).Interface(), nil
		}
		if f, ok := value.(*runtime.FloatValue); ok && f != nil {
			return reflect.ValueOf(f.Val).Convert(targetType).Interface(), nil
		}
	case reflect.Interface:
		return value, nil
	case reflect.Pointer:
		if _, ok := value.(runtime.NilValue); ok {
			return nil, nil
		}
		elemVal, err := i.coerceRuntimeByKind(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(targetType.Elem())
		ev := reflect.ValueOf(elemVal)
		if ev.IsValid() && ev.Type().ConvertibleTo(targetType.Elem()) {
			ptr.Elem().Set(ev.Convert(targetType.Elem()))
		}
		return ptr.Interface(), nil
	case reflect.Slice:
		arr, err := i.toArrayValue(value)
		if err != nil {
			return nil, err
		}
		elemType := targetType.Elem()
		slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
		for idx, elem := range arr.Elements {
			hostElem, err := i.coerceRuntimeByKind(elem, elemType)
			if err != nil {
				return nil, err
			}
			ev := reflect.ValueOf(hostElem)
			if ev.IsValid() && ev.Type().ConvertibleTo(elemType) {
				slice.Index(idx).Set(ev.Convert(elemType))
			}
		}
		return slice.Interface(), nil
	}
	return nil, fmt.Errorf("unsupported extern argument type")
}

func (i *Interpreter) fromHostResults(def *ast.ExternFunctionBody, results []reflect.Value) (runtime.Value, error) {
	if def == nil || def.Signature == nil {
		return runtime.NilValue{}, nil
	}
	if ret, ok := def.Signature.ReturnType.(*ast.ResultTypeExpression); ok {
		if len(results) != 2 {
			return nil, fmt.Errorf("extern result expects two return values")
		}
		errVal := results[1]
		if !errVal.IsNil() {
			if err, ok := errVal.Interface().(error); ok {
				return nil, raiseSignal{value: runtime.ErrorValue{Message: err.Error()}}
			}
			return nil, raiseSignal{value: runtime.ErrorValue{Message: "extern host error"}}
		}
		return i.fromHostValue(ret.InnerType, results[0])
	}
	if len(results) == 0 || def.Signature.ReturnType == nil {
		return runtime.VoidValue{}, nil
	}
	return i.fromHostValue(def.Signature.ReturnType, results[0])
}

func (i *Interpreter) fromHostValue(typeExpr ast.TypeExpression, value reflect.Value) (runtime.Value, error) {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
		}
	}
	for value.IsValid() && value.Kind() == reflect.Interface {
		if value.IsNil() {
			value = reflect.Value{}
			break
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		if _, ok := typeExpr.(*ast.NullableTypeExpression); ok {
			return runtime.NilValue{}, nil
		}
		if simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple != nil {
			if normalizeKernelAliasName(simple.Name.Name) == "void" {
				return runtime.VoidValue{}, nil
			}
		}
		if _, ok := typeExpr.(*ast.UnionTypeExpression); !ok {
			return nil, fmt.Errorf("extern value is nil")
		}
	}
	switch t := typeExpr.(type) {
	case *ast.UnionTypeExpression:
		var lastErr error
		for _, member := range t.Members {
			if member == nil {
				continue
			}
			converted, err := i.fromHostValue(member, value)
			if err == nil {
				return converted, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return runtime.NilValue{}, nil
	case *ast.NullableTypeExpression:
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return runtime.NilValue{}, nil
			}
			return i.fromHostValue(t.InnerType, value.Elem())
		}
		if value.Kind() == reflect.Invalid {
			return runtime.NilValue{}, nil
		}
		return i.fromHostValue(t.InnerType, value)
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && normalizeKernelAliasName(base.Name.Name) == "Array" {
			if value.Kind() != reflect.Slice {
				return nil, fmt.Errorf("extern expected slice result")
			}
			var elemExpr ast.TypeExpression
			if len(t.Arguments) > 0 {
				elemExpr = t.Arguments[0]
			}
			elements := make([]runtime.Value, value.Len())
			for idx := 0; idx < value.Len(); idx++ {
				elemVal, err := i.fromHostValue(elemExpr, value.Index(idx))
				if err != nil {
					return nil, err
				}
				elements[idx] = elemVal
			}
			return i.newArrayValue(elements, len(elements)), nil
		}
		return nil, fmt.Errorf("unsupported extern return type")
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		switch name {
		case "String":
			return runtime.StringValue{Val: fmt.Sprint(value.Interface())}, nil
		case "bool":
			if value.Kind() == reflect.Bool {
				return runtime.BoolValue{Val: value.Bool()}, nil
			}
		case "char":
			switch value.Kind() {
			case reflect.Int32, reflect.Int, reflect.Int64:
				return runtime.CharValue{Val: rune(value.Int())}, nil
			}
		case "IoHandle", "ProcHandle":
			return &runtime.HostHandleValue{HandleType: name, Value: value.Interface()}, nil
		case "f32", "f64":
			if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
				return runtime.FloatValue{Val: value.Convert(reflect.TypeOf(float64(0))).Float(), TypeSuffix: runtime.FloatType(name)}, nil
			}
		case "i8", "i16", "i32", "i64":
			if value.Kind() >= reflect.Int && value.Kind() <= reflect.Int64 {
				return runtime.IntegerValue{Val: bigIntFromInt(value.Int()), TypeSuffix: runtime.IntegerType(name)}, nil
			}
		case "u8", "u16", "u32", "u64":
			if value.Kind() >= reflect.Uint && value.Kind() <= reflect.Uint64 {
				return runtime.IntegerValue{Val: bigIntFromUint(value.Uint()), TypeSuffix: runtime.IntegerType(name)}, nil
			}
		case "i128", "u128":
			if value.Kind() == reflect.Pointer {
				if bi, ok := value.Interface().(*big.Int); ok && bi != nil {
					return runtime.IntegerValue{Val: new(big.Int).Set(bi), TypeSuffix: runtime.IntegerType(name)}, nil
				}
			}
		case "void":
			return runtime.VoidValue{}, nil
		}
		if def, ok := i.lookupStructDefinition(name); ok {
			return i.structFromHostValue(def, value)
		}
	}
	return nil, fmt.Errorf("unsupported extern return type")
}

func (i *Interpreter) lookupStructDefinition(name string) (*runtime.StructDefinitionValue, bool) {
	if i == nil || i.global == nil {
		return nil, false
	}
	if def, ok := i.global.StructDefinition(name); ok && def != nil {
		return def, true
	}
	if val, err := i.global.Get(name); err == nil {
		if def, conv := toStructDefinitionValue(val, name); conv == nil && def != nil {
			return def, true
		}
	}
	return nil, false
}

func (i *Interpreter) structToHostValue(def *runtime.StructDefinitionValue, value runtime.Value) (any, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing")
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		return def.Node.ID.Name, nil
	}
	var inst *runtime.StructInstanceValue
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		inst = v
	case *runtime.StructDefinitionValue:
		if v != nil && v.Node != nil && v.Node.ID != nil && v.Node.ID.Name == def.Node.ID.Name {
			return def.Node.ID.Name, nil
		}
	}
	if inst == nil {
		return nil, fmt.Errorf("expected %s struct instance", def.Node.ID.Name)
	}
	if def.Node.Kind == ast.StructKindPositional && len(inst.Positional) > 0 {
		out := make([]any, len(inst.Positional))
		for idx, elem := range inst.Positional {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			hostElem, err := i.coerceRuntimeToHost(fieldType, elem, reflect.TypeOf((*any)(nil)).Elem())
			if err != nil {
				return nil, err
			}
			out[idx] = hostElem
		}
		return out, nil
	}
	if inst.Fields == nil {
		return nil, fmt.Errorf("expected %s struct fields", def.Node.ID.Name)
	}
	out := make(map[string]any)
	for _, field := range def.Node.Fields {
		if field == nil || field.Name == nil {
			continue
		}
		fieldName := field.Name.Name
		fieldVal, ok := inst.Fields[fieldName]
		if !ok {
			if field.FieldType != nil {
				if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
					out[fieldName] = nil
					continue
				}
			}
			return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
		}
		hostVal, err := i.coerceRuntimeToHost(field.FieldType, fieldVal, reflect.TypeOf((*any)(nil)).Elem())
		if err != nil {
			return nil, err
		}
		out[fieldName] = hostVal
	}
	return out, nil
}

func (i *Interpreter) structFromHostValue(def *runtime.StructDefinitionValue, value reflect.Value) (runtime.Value, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing")
	}
	if def.Node.Kind == ast.StructKindSingleton || len(def.Node.Fields) == 0 {
		return &runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{}}, nil
	}
	if value.Kind() == reflect.Invalid {
		return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
	}
	if value.Kind() == reflect.String && value.String() == def.Node.ID.Name {
		return &runtime.StructInstanceValue{Definition: def, Fields: map[string]runtime.Value{}}, nil
	}
	if def.Node.Kind == ast.StructKindPositional && value.Kind() == reflect.Slice {
		positional := make([]runtime.Value, value.Len())
		for idx := 0; idx < value.Len(); idx++ {
			var fieldType ast.TypeExpression
			if idx < len(def.Node.Fields) && def.Node.Fields[idx] != nil {
				fieldType = def.Node.Fields[idx].FieldType
			}
			elem, err := i.fromHostValue(fieldType, value.Index(idx))
			if err != nil {
				return nil, err
			}
			positional[idx] = elem
		}
		return &runtime.StructInstanceValue{Definition: def, Positional: positional}, nil
	}
	fields := make(map[string]runtime.Value)
	if value.Kind() == reflect.Map {
		for _, field := range def.Node.Fields {
			if field == nil || field.Name == nil {
				continue
			}
			fieldName := field.Name.Name
			key := reflect.ValueOf(fieldName)
			if key.Type().AssignableTo(value.Type().Key()) {
				key = key.Convert(value.Type().Key())
			} else if key.Type().ConvertibleTo(value.Type().Key()) {
				key = key.Convert(value.Type().Key())
			} else {
				continue
			}
			entry := value.MapIndex(key)
			if !entry.IsValid() {
				if field.FieldType != nil {
					if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := i.fromHostValue(field.FieldType, entry)
			if err != nil {
				return nil, err
			}
			fields[fieldName] = converted
		}
		return &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil
	}
	if value.Kind() == reflect.Struct {
		for _, field := range def.Node.Fields {
			if field == nil || field.Name == nil {
				continue
			}
			fieldName := field.Name.Name
			entry := value.FieldByName(fieldName)
			if !entry.IsValid() {
				if field.FieldType != nil {
					if _, ok := field.FieldType.(*ast.NullableTypeExpression); ok {
						fields[fieldName] = runtime.NilValue{}
						continue
					}
				}
				return nil, fmt.Errorf("missing %s.%s", def.Node.ID.Name, fieldName)
			}
			converted, err := i.fromHostValue(field.FieldType, entry)
			if err != nil {
				return nil, err
			}
			fields[fieldName] = converted
		}
		return &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil
	}
	return nil, fmt.Errorf("expected %s struct value", def.Node.ID.Name)
}

func (i *Interpreter) coerceStringValue(value runtime.Value) (string, error) {
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val, nil
	case *runtime.StringValue:
		if v == nil {
			return "", fmt.Errorf("string value is nil")
		}
		return v.Val, nil
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return "", fmt.Errorf("string value is invalid")
		}
		if v.Definition.Node.ID.Name != "String" {
			return "", fmt.Errorf("expected String struct")
		}
		var bytesVal runtime.Value
		if v.Fields != nil {
			bytesVal = v.Fields["bytes"]
		}
		if bytesVal == nil && len(v.Positional) > 0 {
			bytesVal = v.Positional[0]
		}
		arr, err := i.toArrayValue(bytesVal)
		if err != nil {
			return "", err
		}
		if _, err := i.ensureArrayState(arr, 0); err != nil {
			return "", err
		}
		buf := make([]byte, len(arr.Elements))
		for idx, elem := range arr.Elements {
			num, err := toInt64(elem)
			if err != nil {
				return "", err
			}
			if num < 0 || num > 0xff {
				return "", fmt.Errorf("string byte out of range")
			}
			buf[idx] = byte(num)
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("expected String value")
}

func coerceIntValue(value runtime.Value, kind string) (any, error) {
	num, err := toInt64(value)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "i8":
		return int8(num), nil
	case "i16":
		return int16(num), nil
	case "i32":
		return int32(num), nil
	case "i64":
		return int64(num), nil
	}
	return num, nil
}

func coerceUintValue(value runtime.Value, kind string) (any, error) {
	num, err := toUint64(value)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "u8":
		return uint8(num), nil
	case "u16":
		return uint16(num), nil
	case "u32":
		return uint32(num), nil
	case "u64":
		return uint64(num), nil
	}
	return num, nil
}

func coerceBigInt(value runtime.Value) (any, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val == nil {
			return nil, fmt.Errorf("integer is nil")
		}
		return new(big.Int).Set(v.Val), nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil {
			return nil, fmt.Errorf("integer is nil")
		}
		return new(big.Int).Set(v.Val), nil
	}
	return nil, fmt.Errorf("expected integer")
}

func toInt64(value runtime.Value) (int64, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil && v.Val.IsInt64() {
			return v.Val.Int64(), nil
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil && v.Val.IsInt64() {
			return v.Val.Int64(), nil
		}
	}
	return 0, fmt.Errorf("expected integer")
}

func toUint64(value runtime.Value) (uint64, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil && v.Val.IsUint64() {
			return v.Val.Uint64(), nil
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil && v.Val.IsUint64() {
			return v.Val.Uint64(), nil
		}
	}
	return 0, fmt.Errorf("expected unsigned integer")
}

func bigIntFromInt(val int64) *big.Int {
	return big.NewInt(val)
}

func bigIntFromUint(val uint64) *big.Int {
	b := new(big.Int)
	b.SetUint64(val)
	return b
}
