package interpreter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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
