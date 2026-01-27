package interpreter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func hashExternState(target ast.HostTarget, state *externTargetState, salt string) string {
	hasher := sha256.New()
	if salt != "" {
		hasher.Write([]byte("salt:"))
		hasher.Write([]byte(salt))
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
