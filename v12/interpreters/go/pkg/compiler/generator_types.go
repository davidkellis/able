package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) isVoidType(goType string) bool {
	return goType == "struct{}"
}

func (g *generator) isStringType(goType string) bool {
	return goType == "string"
}

func (g *generator) isIntegerType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func (g *generator) isSignedIntegerType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64":
		return true
	}
	return false
}

func (g *generator) isUnsignedIntegerType(goType string) bool {
	switch goType {
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func (g *generator) isFloatType(goType string) bool {
	return goType == "float32" || goType == "float64"
}

func (g *generator) isNumericType(goType string) bool {
	return g.isIntegerType(goType) || g.isFloatType(goType)
}

func (g *generator) isEqualityComparable(goType string) bool {
	return g.isNumericType(goType) || g.isStringType(goType) || goType == "bool" || goType == "rune"
}

func (g *generator) isOrderedComparable(goType string) bool {
	return g.isNumericType(goType) || g.isStringType(goType) || goType == "rune"
}

func (g *generator) structBaseName(goType string) (string, bool) {
	if info := g.structInfoByGoName(goType); info != nil {
		return info.Name, true
	}
	return "", false
}

func (g *generator) structHelperName(goType string) (string, bool) {
	if info := g.structInfoByGoName(goType); info != nil {
		if info.GoName != "" {
			return info.GoName, true
		}
		if info.Name != "" {
			return info.Name, true
		}
	}
	return "", false
}

func (g *generator) sameNominalStructFamily(left string, right string) bool {
	if g == nil || left == "" || right == "" {
		return false
	}
	if strings.HasPrefix(left, "*") != strings.HasPrefix(right, "*") {
		return false
	}
	leftInfo := g.structInfoByGoName(left)
	rightInfo := g.structInfoByGoName(right)
	if leftInfo == nil || rightInfo == nil {
		return false
	}
	if leftInfo.Package != rightInfo.Package || leftInfo.Name == "" || leftInfo.Name != rightInfo.Name {
		return false
	}
	if leftInfo.TypeExpr == nil || rightInfo.TypeExpr == nil {
		return true
	}
	leftExpr := normalizeTypeExprString(g, leftInfo.Package, leftInfo.TypeExpr)
	rightExpr := normalizeTypeExprString(g, rightInfo.Package, rightInfo.TypeExpr)
	return leftExpr != "" && leftExpr == rightExpr
}

func (g *generator) nominalStructCarrierCoercible(expected string, actual string) bool {
	if g == nil || expected == "" || actual == "" {
		return false
	}
	if g.sameNominalStructFamily(expected, actual) {
		return true
	}
	if g.receiverNominalFamilyCompatible(expected, actual) {
		return true
	}
	return g.receiverNominalFamilyCompatible(actual, expected)
}

func (g *generator) recoverRepresentableCarrierType(pkgName string, expr ast.TypeExpression, mapped string) (string, bool) {
	if g == nil || expr == nil {
		return mapped, mapped != ""
	}
	if mapped != "" && mapped != "runtime.Value" && mapped != "any" {
		return mapped, true
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if ifaceExpr, ok := g.interfaceTypeExpr(expr); ok {
		if ifacePkg, _, _, _, ok := interfaceExprInfo(g, pkgName, ifaceExpr); ok {
			if info, ok := g.ensureNativeInterfaceInfo(ifacePkg, ifaceExpr); ok && info != nil && info.GoType != "" {
				return info.GoType, true
			}
		}
		if info, ok := g.ensureNativeInterfaceInfo(pkgName, ifaceExpr); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if fnExpr, ok := expr.(*ast.FunctionTypeExpression); ok {
		if info, ok := g.ensureNativeCallableInfo(pkgName, fnExpr); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if _, members, ok := g.expandedUnionMembersInPackage(pkgName, expr); ok {
		if info, ok := g.ensureNativeUnionInfo(pkgName, members); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if resultExpr, ok := expr.(*ast.ResultTypeExpression); ok {
		if info, ok := g.ensureNativeResultUnionInfo(pkgName, resultExpr); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if nullableExpr, ok := expr.(*ast.NullableTypeExpression); ok {
		mapper := NewTypeMapper(g, pkgName)
		if nullableType, ok := mapper.mapNullableType(nullableExpr); ok && nullableType != "" && nullableType != "runtime.Value" && nullableType != "any" {
			return nullableType, true
		}
	}
	ctx := &compileContext{packageName: pkgName}
	recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, expr)
	if !ok || recovered == "" {
		return mapped, mapped != ""
	}
	return recovered, true
}

func (g *generator) typeExprAllowsNilInPackage(pkgName string, expr ast.TypeExpression) bool {
	if g == nil || expr == nil {
		return false
	}
	normalized := normalizeTypeExprForPackage(g, pkgName, expr)
	switch t := normalized.(type) {
	case *ast.NullableTypeExpression:
		return t != nil
	case *ast.UnionTypeExpression:
		if t == nil {
			return false
		}
		_, ok := nativeUnionNullableInnerTypeExpr(t.Members)
		return ok
	default:
		return false
	}
}

func (g *generator) refreshRepresentableFunctionInfo(info *functionInfo) {
	if g == nil || info == nil {
		return
	}
	if info.cachedCarrier {
		return
	}
	for idx := range info.Params {
		param := &info.Params[idx]
		if param == nil {
			continue
		}
		if derived := g.functionParamTypeExpr(info, idx); derived != nil {
			param.TypeExpr = derived
		}
		if param.TypeExpr == nil {
			continue
		}
		recovered, ok := g.recoverRepresentableCarrierType(info.Package, param.TypeExpr, param.GoType)
		if ok && recovered != "" && (param.GoType == "" || param.GoType == "runtime.Value" || param.GoType == "any" || !g.typeMatches(param.GoType, recovered)) {
			param.GoType = recovered
			param.Supported = true
		}
	}
	retExpr := g.functionReturnTypeExpr(info)
	if recovered, ok := g.recoverRepresentableCarrierType(info.Package, retExpr, info.ReturnType); ok && recovered != "" {
		if info.ReturnType == "" || info.ReturnType == "runtime.Value" || info.ReturnType == "any" || !g.typeMatches(info.ReturnType, recovered) {
			info.ReturnType = recovered
		}
	}
	info.cachedCarrier = true
}

func (g *generator) functionParamTypeExpr(info *functionInfo, idx int) ast.TypeExpression {
	if g == nil || info == nil || info.Definition == nil || idx < 0 || idx >= len(info.Definition.Params) {
		return nil
	}
	param := info.Definition.Params[idx]
	if param == nil {
		return nil
	}
	paramType := param.ParamType
	if impl := g.implMethodByInfo[info]; impl != nil {
		contextBindings := g.compileContextTypeBindings(info)
		concreteTarget := g.specializedImplTargetType(impl, contextBindings)
		if concreteTarget == nil {
			concreteTarget = impl.TargetType
		}
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && (ident.Name == "self" || ident.Name == "Self") {
			if idx >= 0 && idx < len(info.Params) && info.Params[idx].TypeExpr != nil {
				if concreteParam := normalizeTypeExprForPackage(g, info.Package, info.Params[idx].TypeExpr); concreteParam != nil {
					return concreteParam
				}
			}
			if concreteTarget != nil {
				return normalizeTypeExprForPackage(g, info.Package, concreteTarget)
			}
		}
		interfaceBindings := g.implTypeBindings(info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
		selfTarget := g.implSelfTargetType(impl.Info.Package, concreteTarget, interfaceBindings)
		allBindings := g.mergeImplSelfTargetBindings(info.Package, concreteTarget, selfTarget, interfaceBindings)
		for name, expr := range contextBindings {
			if expr == nil {
				continue
			}
			if name == "Self" && selfTarget != nil {
				continue
			}
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings[name] = normalizeTypeExprForPackage(g, info.Package, expr)
		}
		if selfTarget != nil {
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings["Self"] = normalizeTypeExprForPackage(g, info.Package, selfTarget)
		}
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && (ident.Name == "self" || ident.Name == "Self") {
				paramType = selfTarget
			}
		}
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && (ident.Name == "self" || ident.Name == "Self") {
			paramType = selfTarget
		}
		paramType = resolveSelfTypeExpr(paramType, selfTarget)
		paramType = substituteTypeParams(paramType, allBindings)
	}
	paramType = substituteTypeParams(paramType, g.compileContextTypeBindings(info))
	return normalizeTypeExprForPackage(g, info.Package, paramType)
}

func (g *generator) isNativeStructPointerType(goType string) bool {
	return strings.HasPrefix(goType, "*") && g.structInfoByGoName(goType) != nil
}

func (g *generator) isNilableStaticCarrierType(goType string) bool {
	if goType == "" || goType == "runtime.Value" || goType == "any" {
		return false
	}
	if strings.HasPrefix(goType, "*") {
		return true
	}
	if g.nativeInterfaceInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeCallableInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return true
	}
	return false
}

func (g *generator) intBits(goType string) int {
	switch goType {
	case "int8", "uint8":
		return 8
	case "int16", "uint16":
		return 16
	case "int32", "uint32":
		return 32
	case "int64", "uint64":
		return 64
	}
	return 64
}

func (g *generator) nativeIntegerWidenExpr(expr string, srcType string, targetType string) (string, bool) {
	if srcType == targetType {
		return expr, true
	}
	if !g.isIntegerType(srcType) || !g.isIntegerType(targetType) {
		return "", false
	}
	if g.isSignedIntegerType(srcType) != g.isSignedIntegerType(targetType) {
		return "", false
	}
	if g.intBits(srcType) > g.intBits(targetType) {
		return "", false
	}
	return fmt.Sprintf("%s(%s)", targetType, expr), true
}

func (g *generator) nativePrimitiveCastExpr(expr string, srcType string, targetType string) (string, bool) {
	if srcType == targetType {
		return expr, true
	}
	if widened, ok := g.nativeIntegerWidenExpr(expr, srcType, targetType); ok {
		return widened, true
	}
	if g.isIntegerType(srcType) && g.isFloatType(targetType) {
		return fmt.Sprintf("%s(%s)", targetType, expr), true
	}
	if g.isFloatType(srcType) && g.isFloatType(targetType) {
		return fmt.Sprintf("%s(%s)", targetType, expr), true
	}
	return "", false
}

func (g *generator) nativeFloatToIntBounds(targetType string) (string, string, bool) {
	if !g.isIntegerType(targetType) {
		return "", "", false
	}
	if g.isSignedIntegerType(targetType) {
		bits := g.intBits(targetType)
		if bits >= 64 {
			return "-math.Ldexp(1, 63)", "math.Ldexp(1, 63)", true
		}
		upper := int64(1) << uint(bits-1)
		return fmt.Sprintf("-%d.0", upper), fmt.Sprintf("%d.0", upper), true
	}
	bits := g.intBits(targetType)
	if bits >= 64 {
		return "0.0", "math.Ldexp(1, 64)", true
	}
	upper := uint64(1) << uint(bits)
	return "0.0", fmt.Sprintf("%d.0", upper), true
}

func (g *generator) nativePrimitiveCastLines(ctx *compileContext, nodeExpr string, expr string, srcType string, targetType string) ([]string, string, string, bool) {
	if g != nil && ctx != nil && expr != "" && g.isIntegerType(srcType) && g.isIntegerType(targetType) {
		valueTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valueTemp, expr),
			fmt.Sprintf("%s := %s(%s)", resultTemp, targetType, valueTemp),
		}
		return lines, resultTemp, targetType, true
	}
	if directExpr, ok := g.nativePrimitiveCastExpr(expr, srcType, targetType); ok {
		return nil, directExpr, targetType, true
	}
	if g == nil || ctx == nil || expr == "" || !g.isFloatType(srcType) || !g.isIntegerType(targetType) {
		return nil, "", "", false
	}
	lowerBound, upperBound, ok := g.nativeFloatToIntBounds(targetType)
	if !ok {
		return nil, "", "", false
	}
	if nodeExpr == "" {
		nodeExpr = "nil"
	}
	floatTemp := ctx.newTemp()
	truncTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	floatExpr := expr
	if srcType != "float64" {
		floatExpr = fmt.Sprintf("float64(%s)", expr)
	}
	overflowTransfer, ok := g.lowerControlTransfer(ctx, fmt.Sprintf("__able_raise_overflow(%s)", nodeExpr))
	if !ok {
		return nil, "", "", false
	}
	lines := []string{
		fmt.Sprintf("%s := %s", floatTemp, floatExpr),
		fmt.Sprintf("if math.IsNaN(%s) || math.IsInf(%s, 0) {", floatTemp, floatTemp),
	}
	lines = append(lines, indentLines(overflowTransfer, 1)...)
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("%s := math.Trunc(%s)", truncTemp, floatTemp))
	lines = append(lines, fmt.Sprintf("if %s < %s || %s >= %s {", truncTemp, lowerBound, truncTemp, upperBound))
	lines = append(lines, indentLines(overflowTransfer, 1)...)
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("%s := %s(%s)", resultTemp, targetType, truncTemp))
	return lines, resultTemp, targetType, true
}

func (g *generator) integerTypeSuffix(goType string) (string, bool) {
	switch goType {
	case "int8":
		return "i8", true
	case "int16":
		return "i16", true
	case "int32":
		return "i32", true
	case "int64":
		return "i64", true
	case "uint8":
		return "u8", true
	case "uint16":
		return "u16", true
	case "uint32":
		return "u32", true
	case "uint64":
		return "u64", true
	case "int":
		return "isize", true
	case "uint":
		return "usize", true
	default:
		return "", false
	}
}

func (g *generator) isUntypedNumericLiteral(expr ast.Expression) bool {
	switch lit := expr.(type) {
	case *ast.IntegerLiteral:
		return lit != nil && lit.IntegerType == nil
	case *ast.FloatLiteral:
		return lit != nil && lit.FloatType == nil
	default:
		return false
	}
}

func (g *generator) inferIntegerLiteralType(lit *ast.IntegerLiteral) string {
	if lit == nil || lit.IntegerType == nil {
		return "int32"
	}
	switch *lit.IntegerType {
	case ast.IntegerTypeI8:
		return "int8"
	case ast.IntegerTypeI16:
		return "int16"
	case ast.IntegerTypeI32:
		return "int32"
	case ast.IntegerTypeI64:
		return "int64"
	case ast.IntegerTypeI128:
		return "runtime.Value"
	case ast.IntegerTypeU8:
		return "uint8"
	case ast.IntegerTypeU16:
		return "uint16"
	case ast.IntegerTypeU32:
		return "uint32"
	case ast.IntegerTypeU64:
		return "uint64"
	case ast.IntegerTypeU128:
		return "runtime.Value"
	default:
		return "int32"
	}
}

func (g *generator) inferFloatLiteralType(lit *ast.FloatLiteral) string {
	if lit == nil || lit.FloatType == nil {
		return "float64"
	}
	switch *lit.FloatType {
	case ast.FloatTypeF32:
		return "float32"
	case ast.FloatTypeF64:
		return "float64"
	default:
		return "float64"
	}
}

func (g *generator) mapTypeExpression(expr ast.TypeExpression) (string, bool) {
	return g.mapTypeExpressionInPackage("", expr)
}

func (g *generator) mapTypeExpressionInPackage(pkgName string, expr ast.TypeExpression) (string, bool) {
	mapper := NewTypeMapper(g, pkgName)
	return mapper.Map(expr)
}

func (g *generator) interfaceTypeExpr(expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if expr == nil {
		return nil, false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return nil, false
		}
		if g.isInterfaceName(t.Name.Name) {
			return expr, true
		}
	case *ast.GenericTypeExpression:
		if t == nil {
			return nil, false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			if g.isInterfaceName(base.Name.Name) {
				return expr, true
			}
		}
	}
	return nil, false
}

func (g *generator) isResultVoidTypeExpr(expr ast.TypeExpression) bool {
	res, ok := expr.(*ast.ResultTypeExpression)
	if !ok || res == nil || res.InnerType == nil {
		return false
	}
	inner, ok := res.InnerType.(*ast.SimpleTypeExpression)
	if !ok || inner == nil || inner.Name == nil {
		return false
	}
	return inner.Name.Name == "void" || inner.Name.Name == "Void"
}

func (g *generator) isInterfaceName(name string) bool {
	if name == "" || g == nil || g.interfaces == nil {
		return false
	}
	_, ok := g.interfaces[name]
	return ok
}

func (g *generator) renderTypeExpression(expr ast.TypeExpression) (string, bool) {
	if expr == nil {
		return "", false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "", false
		}
		return fmt.Sprintf("ast.Ty(%q)", t.Name.Name), true
	case *ast.GenericTypeExpression:
		if t == nil {
			return "", false
		}
		baseExpr, ok := g.renderTypeExpression(t.Base)
		if !ok {
			return "", false
		}
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			rendered, ok := g.renderTypeExpression(arg)
			if !ok {
				return "", false
			}
			args = append(args, rendered)
		}
		if len(args) == 0 {
			return fmt.Sprintf("ast.Gen(%s)", baseExpr), true
		}
		return fmt.Sprintf("ast.Gen(%s, %s)", baseExpr, strings.Join(args, ", ")), true
	case *ast.FunctionTypeExpression:
		if t == nil {
			return "", false
		}
		params := make([]string, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			rendered, ok := g.renderTypeExpression(param)
			if !ok {
				return "", false
			}
			params = append(params, rendered)
		}
		ret, ok := g.renderTypeExpression(t.ReturnType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.FnType([]ast.TypeExpression{%s}, %s)", strings.Join(params, ", "), ret), true
	case *ast.NullableTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := g.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Nullable(%s)", inner), true
	case *ast.ResultTypeExpression:
		if t == nil {
			return "", false
		}
		inner, ok := g.renderTypeExpression(t.InnerType)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("ast.Result(%s)", inner), true
	case *ast.UnionTypeExpression:
		if t == nil {
			return "", false
		}
		members := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			rendered, ok := g.renderTypeExpression(member)
			if !ok {
				return "", false
			}
			members = append(members, rendered)
		}
		return fmt.Sprintf("ast.UnionT(%s)", strings.Join(members, ", ")), true
	case *ast.WildcardTypeExpression:
		return "ast.WildT()", true
	default:
		return "", false
	}
}

func isFunctionTypeExpr(expr ast.TypeExpression) bool {
	_, ok := expr.(*ast.FunctionTypeExpression)
	return ok
}

func typeExpressionToString(expr ast.TypeExpression) string {
	return typeExpressionToStringSeen(expr, make(map[uintptr]struct{}))
}

func typeExpressionToStringSeen(expr ast.TypeExpression, seen map[uintptr]struct{}) string {
	if expr == nil {
		return "<?>"
	}
	if key, ok := typeExpressionIdentity(expr); ok {
		if _, exists := seen[key]; exists {
			return "<cycle>"
		}
		seen[key] = struct{}{}
		defer delete(seen, key)
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "<?>"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		base := typeExpressionToStringSeen(t.Base, seen)
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			args = append(args, typeExpressionToStringSeen(arg, seen))
		}
		return fmt.Sprintf("%s<%s>", base, strings.Join(args, ", "))
	case *ast.NullableTypeExpression:
		return typeExpressionToStringSeen(t.InnerType, seen) + "?"
	case *ast.ResultTypeExpression:
		return fmt.Sprintf("Result<%s>", typeExpressionToStringSeen(t.InnerType, seen))
	case *ast.FunctionTypeExpression:
		parts := make([]string, 0, len(t.ParamTypes))
		for _, p := range t.ParamTypes {
			parts = append(parts, typeExpressionToStringSeen(p, seen))
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(parts, ", "), typeExpressionToStringSeen(t.ReturnType, seen))
	case *ast.UnionTypeExpression:
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, typeExpressionToStringSeen(member, seen))
		}
		return strings.Join(parts, " | ")
	default:
		return "<?>"
	}
}

func typeExpressionIdentity(expr ast.TypeExpression) (uintptr, bool) {
	value := reflect.ValueOf(expr)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return 0, false
	}
	return value.Pointer(), true
}

func typeNameFromGoType(goType string) string {
	switch goType {
	case "bool":
		return "bool"
	case "string":
		return "String"
	case "rune":
		return "char"
	case "int8":
		return "i8"
	case "int16":
		return "i16"
	case "int32":
		return "i32"
	case "int64":
		return "i64"
	case "uint8":
		return "u8"
	case "uint16":
		return "u16"
	case "uint32":
		return "u32"
	case "uint64":
		return "u64"
	case "int":
		return "isize"
	case "uint":
		return "usize"
	case "float32":
		return "f32"
	case "float64":
		return "f64"
	case "struct{}":
		return "void"
	}
	if strings.HasPrefix(goType, "*") {
		return strings.TrimPrefix(goType, "*")
	}
	return goType
}

func (g *generator) hasOptionalLastParam(info *functionInfo) bool {
	if info == nil || info.Definition == nil {
		return false
	}
	params := info.Definition.Params
	if len(params) == 0 {
		return false
	}
	last := params[len(params)-1]
	if last == nil || last.ParamType == nil {
		return false
	}
	_, ok := last.ParamType.(*ast.NullableTypeExpression)
	return ok
}
