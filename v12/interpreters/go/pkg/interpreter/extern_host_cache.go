//go:build !(js && wasm)

package interpreter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"

	"able/interpreter-go/pkg/ast"
)

const externCacheDirEnv = "ABLE_EXTERN_CACHE_DIR"

func externHostCacheScope() string {
	return strings.Join([]string{
		"go=" + runtime.Version(),
		"goos=" + runtime.GOOS,
		"goarch=" + runtime.GOARCH,
		"goexperiment=" + os.Getenv("GOEXPERIMENT"),
		"goflags=" + os.Getenv("GOFLAGS"),
	}, "\n")
}

func hashExternState(target ast.HostTarget, state *externTargetState, scope string) string {
	hasher := sha256.New()
	if scope != "" {
		hasher.Write([]byte("scope:"))
		hasher.Write([]byte(scope))
		hasher.Write([]byte("\n"))
	}
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

func cachedExternStateHash(target ast.HostTarget, state *externTargetState, scope string) string {
	if state == nil {
		return ""
	}
	if state.hashValid && state.hashScope == scope {
		return state.cachedHash
	}
	hash := hashExternState(target, state, scope)
	state.cachedHash = hash
	state.hashScope = scope
	state.hashValid = true
	return hash
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

func externImageSymbolName(packageKey, name string) string {
	return "AbleImageExtern_" + sanitizeSymbolName(packageKey) + "_" + sanitizeSymbolName(name)
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
