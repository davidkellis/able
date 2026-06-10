package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// bytecodeFrameLayout describes a slot-indexed local variable layout for a
// function body. When non-nil on a bytecodeProgram, the VM uses a flat
// []Value array instead of map-based Environment lookups for locals.
type bytecodeFrameLayout struct {
	slotCount              int // total slots needed (params + locals); set after lowering
	paramSlots             int // number of param slots (always indices 0..paramSlots-1)
	paramTypes             []ast.TypeExpression
	paramSimpleTypes       []string // cached simple type names for params (empty = non-simple)
	paramSimpleChecks      []bytecodeSimpleTypeCheck
	paramKinds             []bytecodeCellKind // cached typed-cell kind for params
	paramExactStructDef    []*runtime.StructDefinitionValue
	paramNeedsCoercion     []bool             // cached "may need runtime coercion" flags for inline arg setup
	anyParamCoercion       bool               // true when any parameter may need runtime coercion
	anyExplicitCoercion    bool               // true when any non-receiver param may need runtime coercion
	methodShorthand        bool               // true when the declaration used implicit self shorthand
	selfCallSlot           int                // reserved slot for recursive self-call fast path; -1 when disabled
	returnType             ast.TypeExpression // declared return type (for coercion on inline return)
	returnSimpleType       string             // cached simple type name for inline return coercion checks
	returnSimpleCheck      bytecodeSimpleTypeCheck
	returnNullableSimple   string
	returnExactStructDef   *runtime.StructDefinitionValue
	returnTypeUsesGenerics bool
	returnTypeHasAlias     bool
	returnCanonicalType    ast.TypeExpression
	usesImplicitMember     bool               // true if body references #member syntax
	needsEnvScopes         bool               // true if body has definitions needing env registration
	preservesControlFlow   bool               // true when execution cannot mutate loop/iterator stacks
	selfCallOneArgFast     bool               // true when one-arg self-call inline can skip declaration shape checks
	firstParamType         ast.TypeExpression // cached first parameter type for self-call inline checks/coercion
	firstParamSimple       string             // cached simple type name for first parameter (empty for non-simple)
	slotKinds              []bytecodeCellKind // typed-cell kind by slot after lowering finalizes locals
	hasTypedSlots          bool
	i32RegisterFrame       bool
	i32FrameProof          *bytecodeI32FrameProof // typechecker-backed VM-v2 eligibility metadata
}

// analyzeFrameLayout inspects a function definition and returns a
// bytecodeFrameLayout if the function body is eligible for slot-indexed
// locals. Returns nil if any bail-out condition is detected, in which
// case the function falls back to map-based Environment storage.
func analyzeFrameLayout(i *Interpreter, def *ast.FunctionDefinition) *bytecodeFrameLayout {
	return analyzeFrameLayoutWithEnv(i, def, nil)
}

func analyzeFrameLayoutWithEnv(i *Interpreter, def *ast.FunctionDefinition, env *runtime.Environment) *bytecodeFrameLayout {
	return analyzeFrameLayoutWithEnvAndMethodSet(i, def, env, nil)
}

func analyzeFrameLayoutWithEnvAndMethodSet(i *Interpreter, def *ast.FunctionDefinition, env *runtime.Environment, methodSet *runtime.MethodSet) *bytecodeFrameLayout {
	if def == nil || def.Body == nil {
		return nil
	}
	// All params must be simple identifiers (no destructuring patterns).
	for _, param := range def.Params {
		if param == nil {
			return nil
		}
		if _, ok := param.Name.(*ast.Identifier); !ok {
			return nil
		}
	}
	analysisEnv := slotEligibleFunctionEnv(env, def)
	// Walk the body against the real definition environment before admitting a
	// slot layout. A nil-env pass can mistake global singleton pattern names
	// for new bindings, after which lowering must fall back to nested match
	// bytecode that cannot see slot-only params such as method receivers.
	slotEligibleWithoutEnv := slotEligibleBlock(def.Body)
	if !slotEligibleBlockWithEnv(def.Body, analysisEnv) {
		return nil
	}
	if !slotEligibleWithoutEnv && blockHasSlotUnsafePlaceholder(def.Body) {
		return nil
	}
	var firstParamType ast.TypeExpression
	firstParamSimple := ""
	generics := buildFunctionGenericNameSet(def, methodSet)
	paramTypes := make([]ast.TypeExpression, len(def.Params))
	paramSimpleTypes := make([]string, len(def.Params))
	paramSimpleChecks := make([]bytecodeSimpleTypeCheck, len(def.Params))
	paramKinds := make([]bytecodeCellKind, len(def.Params))
	paramExactStructDef := make([]*runtime.StructDefinitionValue, len(def.Params))
	paramNeedsCoercion := make([]bool, len(def.Params))
	anyParamCoercion := false
	anyExplicitCoercion := false
	hasTypedSlots := false
	for idx, param := range def.Params {
		if param == nil {
			continue
		}
		paramType := param.ParamType
		if methodSet != nil {
			paramType = resolveSelfTypeExpr(paramType, methodSet.TargetType)
		}
		if i != nil && paramType != nil {
			paramType = i.canonicalizeTypeExpressionCached(paramType, env, i.typeExpressionReferencesAliasCached(paramType))
		}
		paramTypes[idx] = paramType
		paramSimpleTypes[idx] = cachedSimpleTypeName(paramType)
		paramSimpleChecks[idx] = bytecodeSimpleTypeCheckForName(paramSimpleTypes[idx])
		paramKinds[idx] = bytecodeCellKindForSimpleTypeName(paramSimpleTypes[idx])
		paramExactStructDef[idx] = exactNamedStructDefinitionForTypeExpr(i, env, paramType)
		if paramKinds[idx] != bytecodeCellKindValue {
			hasTypedSlots = true
		}
		noOpCoercion := false
		if i != nil {
			noOpCoercion = i.coerceValueToTypeWouldBeNoOp(paramType)
		}
		paramNeedsCoercion[idx] = paramType != nil && !paramUsesGeneric(paramType, generics) && !noOpCoercion
		if paramNeedsCoercion[idx] {
			anyParamCoercion = true
			if idx > 0 {
				anyExplicitCoercion = true
			}
		}
	}
	if len(def.Params) > 0 && def.Params[0] != nil {
		firstParamType = paramTypes[0]
		firstParamSimple = paramSimpleTypes[0]
	}
	returnSimpleType := cachedSimpleTypeName(def.ReturnType)
	returnNullableSimple := cachedNullableSimpleTypeName(def.ReturnType)
	returnExactStructDef := exactNamedStructDefinitionForTypeExpr(i, env, def.ReturnType)
	returnTypeUsesGenerics := typeExpressionUsesGenerics(def.ReturnType, generics)
	returnTypeHasAlias := false
	returnCanonicalType := def.ReturnType
	if i != nil && def.ReturnType != nil {
		returnTypeHasAlias = i.typeExpressionReferencesAliasCached(def.ReturnType)
		returnCanonicalType = i.canonicalizeTypeExpressionCached(def.ReturnType, env, returnTypeHasAlias)
	}
	selfName := ""
	if def.ID != nil {
		selfName = def.ID.Name
	}
	return &bytecodeFrameLayout{
		paramSlots:             len(def.Params),
		paramTypes:             paramTypes,
		paramSimpleTypes:       paramSimpleTypes,
		paramSimpleChecks:      paramSimpleChecks,
		paramKinds:             paramKinds,
		paramExactStructDef:    paramExactStructDef,
		paramNeedsCoercion:     paramNeedsCoercion,
		anyParamCoercion:       anyParamCoercion,
		anyExplicitCoercion:    anyExplicitCoercion,
		methodShorthand:        def.IsMethodShorthand,
		selfCallSlot:           -1,
		returnType:             def.ReturnType,
		returnSimpleType:       returnSimpleType,
		returnSimpleCheck:      bytecodeSimpleTypeCheckForName(returnSimpleType),
		returnNullableSimple:   returnNullableSimple,
		returnExactStructDef:   returnExactStructDef,
		returnTypeUsesGenerics: returnTypeUsesGenerics,
		returnTypeHasAlias:     returnTypeHasAlias,
		returnCanonicalType:    returnCanonicalType,
		usesImplicitMember:     blockUsesImplicitMember(def.Body),
		needsEnvScopes:         blockNeedsEnvScopes(def.Body),
		preservesControlFlow:   blockPreservesControlFlow(def.Body),
		selfCallOneArgFast:     !def.IsMethodShorthand && len(def.Params) == 1 && len(def.GenericParams) == 0,
		firstParamType:         firstParamType,
		firstParamSimple:       firstParamSimple,
		slotKinds:              append([]bytecodeCellKind(nil), paramKinds...),
		hasTypedSlots:          hasTypedSlots,
		i32RegisterFrame:       blockCanUseI32RegisterFrame(def.Body, selfName),
	}
}

func slotEligibleFunctionEnv(parent *runtime.Environment, def *ast.FunctionDefinition) *runtime.Environment {
	env := runtime.NewEnvironmentWithValueCapacity(parent, len(def.Params))
	for _, param := range def.Params {
		if param == nil {
			continue
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			continue
		}
		env.DefineWithoutMerge(ident.Name, runtime.NilValue{})
	}
	return env
}

func blockCanUseI32RegisterFrame(block *ast.BlockExpression, selfName string) bool {
	if block == nil {
		return true
	}
	for _, stmt := range block.Body {
		if !stmtCanUseI32RegisterFrame(stmt, selfName) {
			return false
		}
	}
	return true
}

func stmtCanUseI32RegisterFrame(stmt ast.Statement, selfName string) bool {
	if stmt == nil {
		return true
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		return exprCanUseI32RegisterFrame(s.Argument, selfName)
	case *ast.BreakStatement:
		return exprCanUseI32RegisterFrame(s.Value, selfName)
	case *ast.ContinueStatement:
		return true
	case *ast.WhileLoop:
		return exprCanUseI32RegisterFrame(s.Condition, selfName) && blockCanUseI32RegisterFrame(s.Body, selfName)
	case *ast.LoopExpression:
		return blockCanUseI32RegisterFrame(s.Body, selfName)
	case *ast.IfExpression:
		return exprCanUseI32RegisterFrame(s.IfCondition, selfName) &&
			blockCanUseI32RegisterFrame(s.IfBody, selfName) &&
			elseIfClausesCanUseI32RegisterFrame(s.ElseIfClauses, selfName) &&
			blockCanUseI32RegisterFrame(s.ElseBody, selfName)
	case ast.Expression:
		return exprCanUseI32RegisterFrame(s, selfName)
	default:
		return false
	}
}

func elseIfClausesCanUseI32RegisterFrame(clauses []*ast.ElseIfClause, selfName string) bool {
	for _, clause := range clauses {
		if clause == nil {
			continue
		}
		if !exprCanUseI32RegisterFrame(clause.Condition, selfName) || !blockCanUseI32RegisterFrame(clause.Body, selfName) {
			return false
		}
	}
	return true
}

func exprCanUseI32RegisterFrame(expr ast.Expression, selfName string) bool {
	if expr == nil {
		return true
	}
	switch n := expr.(type) {
	case *ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return true
	case *ast.BinaryExpression:
		return exprCanUseI32RegisterFrame(n.Left, selfName) && exprCanUseI32RegisterFrame(n.Right, selfName)
	case *ast.UnaryExpression:
		return exprCanUseI32RegisterFrame(n.Operand, selfName)
	case *ast.TypeCastExpression:
		return exprCanUseI32RegisterFrame(n.Expression, selfName)
	case *ast.AssignmentExpression:
		if indexExpr, ok := n.Left.(*ast.IndexExpression); ok {
			if n.Operator != ast.AssignmentAssign {
				return false
			}
			return exprCanUseI32RegisterFrame(indexExpr.Object, selfName) &&
				exprCanUseI32RegisterFrame(indexExpr.Index, selfName) &&
				exprCanUseI32RegisterFrame(n.Right, selfName)
		}
		if _, ok := resolveAssignmentTargetName(n.Left); !ok {
			return false
		}
		if n.Operator == ast.AssignmentDeclare {
			if _, typed := n.Left.(*ast.TypedPattern); !typed {
				return false
			}
		}
		return exprCanUseI32RegisterFrame(n.Right, selfName)
	case *ast.FunctionCall:
		if len(n.TypeArguments) > 0 {
			return false
		}
		if member, ok := n.Callee.(*ast.MemberAccessExpression); ok && member != nil {
			return canonicalArraySlotCallCanUseI32RegisterFrame(n, member, selfName)
		}
		ident, ok := n.Callee.(*ast.Identifier)
		if !ok || ident == nil {
			return false
		}
		if selfName != "" && ident.Name == selfName {
			return false
		}
		for _, arg := range n.Arguments {
			if !exprCanUseI32RegisterFrame(arg, selfName) {
				return false
			}
		}
		return true
	case *ast.IndexExpression:
		return exprCanUseI32RegisterFrame(n.Object, selfName) &&
			exprCanUseI32RegisterFrame(n.Index, selfName)
	case *ast.BlockExpression:
		return blockCanUseI32RegisterFrame(n, selfName)
	case *ast.IfExpression:
		return stmtCanUseI32RegisterFrame(n, selfName)
	case *ast.LoopExpression:
		return blockCanUseI32RegisterFrame(n.Body, selfName)
	default:
		return false
	}
}

func canonicalArraySlotCallCanUseI32RegisterFrame(call *ast.FunctionCall, member *ast.MemberAccessExpression, selfName string) bool {
	if call == nil || member == nil || member.Safe {
		return false
	}
	switch bytecodeIdentifierMemberName(member.Member) {
	case "read_slot":
		if len(call.Arguments) != 1 {
			return false
		}
		return exprCanUseI32RegisterFrame(member.Object, selfName) &&
			exprCanUseI32RegisterFrame(call.Arguments[0], selfName)
	case "write_slot":
		if len(call.Arguments) != 2 {
			return false
		}
		return exprCanUseI32RegisterFrame(member.Object, selfName) &&
			exprCanUseI32RegisterFrame(call.Arguments[0], selfName) &&
			exprCanUseI32RegisterFrame(call.Arguments[1], selfName)
	default:
		return false
	}
}

// blockUsesImplicitMember returns true if the block contains any
// ImplicitMemberExpression or ImplicitMemberSetExpression.
func blockUsesImplicitMember(block *ast.BlockExpression) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if stmtUsesImplicitMember(stmt) {
			return true
		}
	}
	return false
}

func stmtUsesImplicitMember(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.ForLoop:
		return exprUsesImplicitMember(s.Iterable) || blockUsesImplicitMember(s.Body)
	case *ast.WhileLoop:
		if s == nil {
			return false
		}
		return exprUsesImplicitMember(s.Condition) || blockUsesImplicitMember(s.Body)
	case *ast.ReturnStatement:
		if s != nil {
			return exprUsesImplicitMember(s.Argument)
		}
	case ast.Expression:
		return exprUsesImplicitMember(s)
	}
	return false
}

func exprUsesImplicitMember(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.ImplicitMemberExpression:
		return true
	case *ast.BinaryExpression:
		return exprUsesImplicitMember(n.Left) || exprUsesImplicitMember(n.Right)
	case *ast.UnaryExpression:
		return exprUsesImplicitMember(n.Operand)
	case *ast.AssignmentExpression:
		if impl, ok := n.Left.(*ast.ImplicitMemberExpression); ok && impl != nil {
			return true
		}
		return exprUsesImplicitMember(n.Right)
	case *ast.FunctionCall:
		if exprUsesImplicitMember(n.Callee) {
			return true
		}
		for _, arg := range n.Arguments {
			if exprUsesImplicitMember(arg) {
				return true
			}
		}
	case *ast.MemberAccessExpression:
		return exprUsesImplicitMember(n.Object)
	case *ast.IndexExpression:
		return exprUsesImplicitMember(n.Object) || exprUsesImplicitMember(n.Index)
	case *ast.BlockExpression:
		return blockUsesImplicitMember(n)
	case *ast.IfExpression:
		if exprUsesImplicitMember(n.IfCondition) || blockUsesImplicitMember(n.IfBody) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil && (exprUsesImplicitMember(clause.Condition) || blockUsesImplicitMember(clause.Body)) {
				return true
			}
		}
		return blockUsesImplicitMember(n.ElseBody)
	case *ast.MatchExpression:
		if exprUsesImplicitMember(n.Subject) {
			return true
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if exprUsesImplicitMember(clause.Guard) || exprUsesImplicitMember(clause.Body) {
				return true
			}
		}
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if exprUsesImplicitMember(el) {
				return true
			}
		}
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if exprUsesImplicitMember(part) {
				return true
			}
		}
	case *ast.TypeCastExpression:
		return exprUsesImplicitMember(n.Expression)
	case *ast.RangeExpression:
		return exprUsesImplicitMember(n.Start) || exprUsesImplicitMember(n.End)
	case *ast.PropagationExpression:
		return exprUsesImplicitMember(n.Expression)
	case *ast.AwaitExpression:
		return exprUsesImplicitMember(n.Expression)
	case *ast.LoopExpression:
		return blockUsesImplicitMember(n.Body)
	}
	return false
}

type bytecodeSimpleTypeCheck uint8

const (
	bytecodeSimpleTypeCheckUnknown bytecodeSimpleTypeCheck = iota
	bytecodeSimpleTypeCheckAnyInteger
	bytecodeSimpleTypeCheckI8
	bytecodeSimpleTypeCheckI16
	bytecodeSimpleTypeCheckI32
	bytecodeSimpleTypeCheckI64
	bytecodeSimpleTypeCheckI128
	bytecodeSimpleTypeCheckU8
	bytecodeSimpleTypeCheckU16
	bytecodeSimpleTypeCheckU32
	bytecodeSimpleTypeCheckU64
	bytecodeSimpleTypeCheckU128
	bytecodeSimpleTypeCheckIsize
	bytecodeSimpleTypeCheckUsize
	bytecodeSimpleTypeCheckAnyFloat
	bytecodeSimpleTypeCheckF32
	bytecodeSimpleTypeCheckF64
	bytecodeSimpleTypeCheckString
	bytecodeSimpleTypeCheckBool
	bytecodeSimpleTypeCheckChar
	bytecodeSimpleTypeCheckIteratorEnd
)

func bytecodeSimpleTypeCheckForName(typeName string) bytecodeSimpleTypeCheck {
	switch typeName {
	case "Int":
		return bytecodeSimpleTypeCheckAnyInteger
	case "i8":
		return bytecodeSimpleTypeCheckI8
	case "i16":
		return bytecodeSimpleTypeCheckI16
	case "i32":
		return bytecodeSimpleTypeCheckI32
	case "i64":
		return bytecodeSimpleTypeCheckI64
	case "i128":
		return bytecodeSimpleTypeCheckI128
	case "u8":
		return bytecodeSimpleTypeCheckU8
	case "u16":
		return bytecodeSimpleTypeCheckU16
	case "u32":
		return bytecodeSimpleTypeCheckU32
	case "u64":
		return bytecodeSimpleTypeCheckU64
	case "u128":
		return bytecodeSimpleTypeCheckU128
	case "isize":
		return bytecodeSimpleTypeCheckIsize
	case "usize":
		return bytecodeSimpleTypeCheckUsize
	case "Float":
		return bytecodeSimpleTypeCheckAnyFloat
	case "f32":
		return bytecodeSimpleTypeCheckF32
	case "f64":
		return bytecodeSimpleTypeCheckF64
	case "String", "string":
		return bytecodeSimpleTypeCheckString
	case "Bool", "bool":
		return bytecodeSimpleTypeCheckBool
	case "char":
		return bytecodeSimpleTypeCheckChar
	case "IteratorEnd":
		return bytecodeSimpleTypeCheckIteratorEnd
	default:
		return bytecodeSimpleTypeCheckUnknown
	}
}

func (check bytecodeSimpleTypeCheck) integerType() (runtime.IntegerType, bool) {
	switch check {
	case bytecodeSimpleTypeCheckI8:
		return runtime.IntegerI8, true
	case bytecodeSimpleTypeCheckI16:
		return runtime.IntegerI16, true
	case bytecodeSimpleTypeCheckI32:
		return runtime.IntegerI32, true
	case bytecodeSimpleTypeCheckI64:
		return runtime.IntegerI64, true
	case bytecodeSimpleTypeCheckI128:
		return runtime.IntegerI128, true
	case bytecodeSimpleTypeCheckU8:
		return runtime.IntegerU8, true
	case bytecodeSimpleTypeCheckU16:
		return runtime.IntegerU16, true
	case bytecodeSimpleTypeCheckU32:
		return runtime.IntegerU32, true
	case bytecodeSimpleTypeCheckU64:
		return runtime.IntegerU64, true
	case bytecodeSimpleTypeCheckU128:
		return runtime.IntegerU128, true
	case bytecodeSimpleTypeCheckIsize:
		return runtime.IntegerIsize, true
	case bytecodeSimpleTypeCheckUsize:
		return runtime.IntegerUsize, true
	default:
		return "", false
	}
}

func (check bytecodeSimpleTypeCheck) floatType() (runtime.FloatType, bool) {
	switch check {
	case bytecodeSimpleTypeCheckF32:
		return runtime.FloatF32, true
	case bytecodeSimpleTypeCheckF64:
		return runtime.FloatF64, true
	default:
		return "", false
	}
}

func inlineCoercionUnnecessaryBySimpleCheck(check bytecodeSimpleTypeCheck, val runtime.Value) bool {
	switch check {
	case bytecodeSimpleTypeCheckAnyInteger:
		return isIntegerValue(val)
	case bytecodeSimpleTypeCheckAnyFloat:
		switch v := val.(type) {
		case runtime.FloatValue:
			return true
		case *runtime.FloatValue:
			return v != nil
		default:
			return false
		}
	case bytecodeSimpleTypeCheckString:
		switch v := val.(type) {
		case runtime.StringValue:
			return true
		case *runtime.StringValue:
			return v != nil
		default:
			return false
		}
	case bytecodeSimpleTypeCheckBool:
		switch v := val.(type) {
		case runtime.BoolValue:
			return true
		case *runtime.BoolValue:
			return v != nil
		default:
			return false
		}
	case bytecodeSimpleTypeCheckChar:
		switch v := val.(type) {
		case runtime.CharValue:
			return true
		case *runtime.CharValue:
			return v != nil
		default:
			return false
		}
	}
	if kind, ok := check.integerType(); ok {
		if rawKind, _, ok := bytecodeRawIntegerValueInfo(val); ok {
			return rawKind == kind
		}
		switch v := val.(type) {
		case runtime.IntegerValue:
			return v.TypeSuffix == kind
		case *runtime.IntegerValue:
			return v != nil && v.TypeSuffix == kind
		default:
			return false
		}
	}
	if kind, ok := check.floatType(); ok {
		switch v := val.(type) {
		case runtime.FloatValue:
			return v.TypeSuffix == kind
		case *runtime.FloatValue:
			return v != nil && v.TypeSuffix == kind
		default:
			return false
		}
	}
	return false
}

// inlineCoercionUnnecessary returns true when the value trivially matches
// the declared type, allowing the inline call path to skip the expensive
// coerceValueToType / coerceReturnValue calls.
func inlineCoercionUnnecessary(typeExpr ast.TypeExpression, val runtime.Value) bool {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	name := simple.Name.Name
	switch v := val.(type) {
	case runtime.IntegerValue:
		// "Int" is not a fixed-width type; coercion is always a no-op.
		// For fixed-width types (i32, i64, etc.), skip only when suffix matches.
		if name == "Int" {
			return true
		}
		return string(v.TypeSuffix) == name
	case runtime.FloatValue:
		if name == "Float" {
			return true
		}
		return string(v.TypeSuffix) == name
	case runtime.StringValue:
		return name == "String"
	case runtime.BoolValue:
		return name == "Bool"
	case *runtime.IntegerValue:
		if v == nil {
			return false
		}
		if name == "Int" {
			return true
		}
		return string(v.TypeSuffix) == name
	case *runtime.FloatValue:
		if v == nil {
			return false
		}
		if name == "Float" {
			return true
		}
		return string(v.TypeSuffix) == name
	case *runtime.StringValue:
		return v != nil && name == "String"
	case *runtime.BoolValue:
		return v != nil && name == "Bool"
	}
	if kind, _, ok := bytecodeRawIntegerValueInfo(val); ok {
		if name == "Int" {
			return true
		}
		return string(kind) == name
	}
	return false
}

func inlineCoercionUnnecessaryBySimpleType(typeName string, val runtime.Value) bool {
	if typeName == "" {
		return false
	}
	switch v := val.(type) {
	case runtime.IntegerValue:
		if typeName == "Int" {
			return true
		}
		return string(v.TypeSuffix) == typeName
	case runtime.FloatValue:
		if typeName == "Float" {
			return true
		}
		return string(v.TypeSuffix) == typeName
	case runtime.StringValue:
		return typeName == "String"
	case runtime.BoolValue:
		return typeName == "Bool"
	case *runtime.IntegerValue:
		if v == nil {
			return false
		}
		if typeName == "Int" {
			return true
		}
		return string(v.TypeSuffix) == typeName
	case *runtime.FloatValue:
		if v == nil {
			return false
		}
		if typeName == "Float" {
			return true
		}
		return string(v.TypeSuffix) == typeName
	case *runtime.StringValue:
		return v != nil && typeName == "String"
	case *runtime.BoolValue:
		return v != nil && typeName == "Bool"
	default:
		return false
	}
}

func cachedSimpleTypeName(typeExpr ast.TypeExpression) string {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return ""
	}
	return simple.Name.Name
}

func cachedNullableSimpleTypeName(typeExpr ast.TypeExpression) string {
	nullable, ok := typeExpr.(*ast.NullableTypeExpression)
	if !ok || nullable == nil {
		return ""
	}
	return cachedSimpleTypeName(nullable.InnerType)
}

// blockNeedsEnvScopes returns true if the block contains statements that
// register definitions in the environment (struct defs, imports, etc.),
// meaning EnterScope/ExitScope cannot be skipped.
func blockNeedsEnvScopes(block *ast.BlockExpression) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		switch s := stmt.(type) {
		case *ast.StructDefinition, *ast.UnionDefinition, *ast.TypeAliasDefinition,
			*ast.InterfaceDefinition, *ast.ExternFunctionBody,
			*ast.ImportStatement, *ast.DynImportStatement:
			return true
		case *ast.ForLoop:
			if blockNeedsEnvScopes(s.Body) {
				return true
			}
		case *ast.WhileLoop:
			if blockNeedsEnvScopes(s.Body) {
				return true
			}
		case ast.Expression:
			if exprNeedsEnvScopes(s) {
				return true
			}
		}
	}
	return false
}

func blockPreservesControlFlow(block *ast.BlockExpression) bool {
	if block == nil {
		return true
	}
	for _, stmt := range block.Body {
		if !stmtPreservesControlFlow(stmt) {
			return false
		}
	}
	return true
}

func stmtPreservesControlFlow(stmt ast.Statement) bool {
	if stmt == nil {
		return true
	}
	switch s := stmt.(type) {
	case *ast.ForLoop, *ast.WhileLoop:
		return false
	case *ast.ReturnStatement:
		if s != nil {
			return exprPreservesControlFlow(s.Argument)
		}
	case *ast.BreakStatement:
		if s != nil {
			return exprPreservesControlFlow(s.Value)
		}
	case *ast.ContinueStatement:
		return true
	case ast.Expression:
		return exprPreservesControlFlow(s)
	}
	return false
}

func exprPreservesControlFlow(expr ast.Expression) bool {
	if expr == nil {
		return true
	}
	switch n := expr.(type) {
	case *ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return true
	case *ast.BinaryExpression:
		return exprPreservesControlFlow(n.Left) && exprPreservesControlFlow(n.Right)
	case *ast.UnaryExpression:
		return exprPreservesControlFlow(n.Operand)
	case *ast.AssignmentExpression:
		return exprPreservesControlFlow(n.Right)
	case *ast.FunctionCall:
		if !exprPreservesControlFlow(n.Callee) {
			return false
		}
		for _, arg := range n.Arguments {
			if !exprPreservesControlFlow(arg) {
				return false
			}
		}
		return true
	case *ast.MemberAccessExpression:
		return exprPreservesControlFlow(n.Object)
	case *ast.IndexExpression:
		return exprPreservesControlFlow(n.Object) && exprPreservesControlFlow(n.Index)
	case *ast.BlockExpression:
		return blockPreservesControlFlow(n)
	case *ast.IfExpression:
		if !exprPreservesControlFlow(n.IfCondition) || !blockPreservesControlFlow(n.IfBody) || !blockPreservesControlFlow(n.ElseBody) {
			return false
		}
		for _, clause := range n.ElseIfClauses {
			if clause == nil {
				continue
			}
			if !exprPreservesControlFlow(clause.Condition) || !blockPreservesControlFlow(clause.Body) {
				return false
			}
		}
		return true
	case *ast.MatchExpression:
		if !exprPreservesControlFlow(n.Subject) {
			return false
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if !exprPreservesControlFlow(clause.Guard) || !exprPreservesControlFlow(clause.Body) {
				return false
			}
		}
		return true
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if !exprPreservesControlFlow(el) {
				return false
			}
		}
		return true
	case *ast.StructLiteral:
		for _, field := range n.Fields {
			if field != nil && !exprPreservesControlFlow(field.Value) {
				return false
			}
		}
		for _, src := range n.FunctionalUpdateSources {
			if !exprPreservesControlFlow(src) {
				return false
			}
		}
		return true
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if !exprPreservesControlFlow(part) {
				return false
			}
		}
		return true
	case *ast.TypeCastExpression:
		return exprPreservesControlFlow(n.Expression)
	case *ast.RangeExpression:
		return exprPreservesControlFlow(n.Start) && exprPreservesControlFlow(n.End)
	case *ast.PropagationExpression:
		return exprPreservesControlFlow(n.Expression)
	case *ast.AwaitExpression:
		return exprPreservesControlFlow(n.Expression)
	case *ast.LoopExpression, *ast.IteratorLiteral:
		return false
	default:
		return false
	}
}

func exprNeedsEnvScopes(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.BlockExpression:
		return blockNeedsEnvScopes(n)
	case *ast.IfExpression:
		if blockNeedsEnvScopes(n.IfBody) || blockNeedsEnvScopes(n.ElseBody) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil && blockNeedsEnvScopes(clause.Body) {
				return true
			}
		}
	case *ast.MatchExpression:
		if exprNeedsEnvScopes(n.Subject) {
			return true
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if exprNeedsEnvScopes(clause.Guard) || exprNeedsEnvScopes(clause.Body) {
				return true
			}
		}
	case *ast.LoopExpression:
		return blockNeedsEnvScopes(n.Body)
	}
	return false
}
