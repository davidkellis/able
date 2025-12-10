package ast

import "math/big"

// Identifier and literal helpers.

func ID(name string) *Identifier {
	return NewIdentifier(name)
}

func Str(value string) *StringLiteral {
	return NewStringLiteral(value)
}

func Int(value int64) *IntegerLiteral {
	return IntTyped(value, nil)
}

func IntTyped(value int64, integerType *IntegerType) *IntegerLiteral {
	return NewIntegerLiteral(big.NewInt(value), integerType)
}

func IntBig(value *big.Int, integerType *IntegerType) *IntegerLiteral {
	return NewIntegerLiteral(new(big.Int).Set(value), integerType)
}

func Flt(value float64) *FloatLiteral {
	return FltTyped(value, nil)
}

func FltTyped(value float64, floatType *FloatType) *FloatLiteral {
	return NewFloatLiteral(value, floatType)
}

func Bool(value bool) *BooleanLiteral {
	return NewBooleanLiteral(value)
}

func Nil() *NilLiteral {
	return NewNilLiteral()
}

func Chr(value string) *CharLiteral {
	return NewCharLiteral(value)
}

func Arr(elements ...Expression) *ArrayLiteral {
	return NewArrayLiteral(elements)
}

// Type expression helpers.

func Ty(name string) *SimpleTypeExpression {
	return NewSimpleTypeExpression(ID(name))
}

func TyID(id *Identifier) *SimpleTypeExpression {
	return NewSimpleTypeExpression(id)
}

func Gen(base TypeExpression, args ...TypeExpression) *GenericTypeExpression {
	return NewGenericTypeExpression(base, args)
}

func FnType(params []TypeExpression, returnType TypeExpression) *FunctionTypeExpression {
	return NewFunctionTypeExpression(params, returnType)
}

func Nullable(inner TypeExpression) *NullableTypeExpression {
	return NewNullableTypeExpression(inner)
}

func Result(inner TypeExpression) *ResultTypeExpression {
	return NewResultTypeExpression(inner)
}

func UnionT(members ...TypeExpression) *UnionTypeExpression {
	return NewUnionTypeExpression(members)
}

func WildT() *WildcardTypeExpression {
	return NewWildcardTypeExpression()
}

func InterfaceConstr(interfaceType TypeExpression) *InterfaceConstraint {
	return NewInterfaceConstraint(interfaceType)
}

func GenericParam(name string, constraints ...*InterfaceConstraint) *GenericParameter {
	return NewGenericParameter(ID(name), constraints)
}

func WhereConstraint(typeParam string, constraints ...*InterfaceConstraint) *WhereClauseConstraint {
	return NewWhereClauseConstraint(ID(typeParam), constraints)
}

// Pattern helpers.

func Wc() *WildcardPattern {
	return NewWildcardPattern()
}

func LitP(l Literal) *LiteralPattern {
	return NewLiteralPattern(l)
}

func TypedP(pattern Pattern, typeAnnotation TypeExpression) *TypedPattern {
	return NewTypedPattern(pattern, typeAnnotation)
}

func FieldP(pattern Pattern, fieldName interface{}, binding interface{}, typeAnnotation ...TypeExpression) *StructPatternField {
	var ann TypeExpression
	if len(typeAnnotation) > 0 {
		ann = typeAnnotation[0]
	}
	return NewStructPatternField(pattern, identifierPtr(fieldName), identifierPtr(binding), ann)
}

func StructP(fields []*StructPatternField, isPositional bool, structType interface{}) *StructPattern {
	return NewStructPattern(fields, isPositional, identifierPtr(structType))
}

func ArrP(elements []Pattern, rest Pattern) *ArrayPattern {
	return NewArrayPattern(elements, rest)
}

func PatternFrom(value interface{}) Pattern {
	switch v := value.(type) {
	case string:
		return ID(v)
	case *Identifier:
		return v
	case Pattern:
		return v
	default:
		panic("ast: unsupported pattern value")
	}
}

// Expression helpers.

func Un(operator UnaryOperator, operand Expression) *UnaryExpression {
	return NewUnaryExpression(operator, operand)
}

func Bin(operator string, left, right Expression) *BinaryExpression {
	return NewBinaryExpression(operator, left, right)
}

func CallExpr(callee Expression, args ...Expression) *FunctionCall {
	return NewFunctionCall(callee, args, nil, false)
}

func Call(name string, args ...Expression) *FunctionCall {
	return CallExpr(ID(name), args...)
}

func CallT(callee Expression, typeArgs []TypeExpression, args ...Expression) *FunctionCall {
	return NewFunctionCall(callee, args, typeArgs, false)
}

func CallTS(name string, typeArgs []TypeExpression, args ...Expression) *FunctionCall {
	return CallT(ID(name), typeArgs, args...)
}

func Block(statements ...Statement) *BlockExpression {
	return NewBlockExpression(statements)
}

func Assign(left AssignmentTarget, value Expression) *AssignmentExpression {
	return AssignOp(AssignmentDeclare, left, value)
}

func AssignOp(op AssignmentOperator, left AssignmentTarget, value Expression) *AssignmentExpression {
	return NewAssignmentExpression(op, left, value)
}

func AssignMember(object Expression, member interface{}, value Expression) *AssignmentExpression {
	memberExpr := memberExpression(member)
	return AssignOp(AssignmentAssign, NewMemberAccessExpression(object, memberExpr), value)
}

func AssignIndex(object Expression, index Expression, value Expression) *AssignmentExpression {
	return AssignOp(AssignmentAssign, NewIndexExpression(object, index), value)
}

func Range(start, end Expression, inclusive bool) *RangeExpression {
	return NewRangeExpression(start, end, inclusive)
}

func Interp(parts ...Expression) *StringInterpolation {
	return NewStringInterpolation(parts)
}

func Member(object Expression, member interface{}) *MemberAccessExpression {
	return NewMemberAccessExpression(object, memberExpression(member))
}

func Index(object Expression, index Expression) *IndexExpression {
	return NewIndexExpression(object, index)
}

func Lam(params []*FunctionParameter, body Expression) *LambdaExpression {
	return NewLambdaExpression(params, body, nil, nil, nil, false)
}

func ImplicitMember(name interface{}) *ImplicitMemberExpression {
	return NewImplicitMemberExpression(identifierPtr(name))
}

func Placeholder() *PlaceholderExpression {
	return NewPlaceholderExpression(nil)
}

func PlaceholderN(index int) *PlaceholderExpression {
	idx := index
	return NewPlaceholderExpression(&idx)
}

func LamBlock(params []*FunctionParameter, body *BlockExpression) *LambdaExpression {
	return NewLambdaExpression(params, body, nil, nil, nil, true)
}

func Proc(expr Expression) *ProcExpression {
	return NewProcExpression(expr)
}

func Spawn(expr Expression) *SpawnExpression {
	return NewSpawnExpression(expr)
}

func Await(expr Expression) *AwaitExpression {
	return NewAwaitExpression(expr)
}

func Prop(expr Expression) *PropagationExpression {
	return NewPropagationExpression(expr)
}

func OrElse(expr Expression, errorBinding interface{}, handlerStatements ...Statement) *OrElseExpression {
	return OrElseBlock(expr, Block(handlerStatements...), errorBinding)
}

func OrElseBlock(expr Expression, handler *BlockExpression, errorBinding interface{}) *OrElseExpression {
	return NewOrElseExpression(expr, handler, identifierPtr(errorBinding))
}

func Breakpoint(label interface{}, body *BlockExpression) *BreakpointExpression {
	return NewBreakpointExpression(identifierPtr(label), body)
}

func Bp(label interface{}, statements ...Statement) *BreakpointExpression {
	return Breakpoint(label, Block(statements...))
}

// Control flow helpers.

func ElseIf(body *BlockExpression, condition Expression) *ElseIfClause {
	return NewElseIfClause(body, condition)
}

func IfExpr(condition Expression, body *BlockExpression, elseIfClauses ...*ElseIfClause) *IfExpression {
	return NewIfExpression(condition, body, elseIfClauses, nil)
}

func Iff(condition Expression, statements ...Statement) *IfExpression {
	return IfExpr(condition, Block(statements...))
}

func Mc(pattern Pattern, body Expression, guard ...Expression) *MatchClause {
	var g Expression
	if len(guard) > 0 {
		g = guard[0]
	}
	return NewMatchClause(pattern, body, g)
}

func Match(subject Expression, clauses ...*MatchClause) *MatchExpression {
	return NewMatchExpression(subject, clauses)
}

func While(condition Expression, body *BlockExpression) *WhileLoop {
	return NewWhileLoop(condition, body)
}

func IteratorLit(statements ...Statement) *IteratorLiteral {
	return NewIteratorLiteral(statements)
}

func Wloop(condition Expression, statements ...Statement) *WhileLoop {
	return While(condition, Block(statements...))
}

func LoopExpr(body *BlockExpression) *LoopExpression {
	return NewLoopExpression(body)
}

func Loop(statements ...Statement) *LoopExpression {
	return LoopExpr(Block(statements...))
}

func ForLoopPattern(pattern Pattern, iterable Expression, body *BlockExpression) *ForLoop {
	return NewForLoop(pattern, iterable, body)
}

func ForIn(pattern interface{}, iterable Expression, statements ...Statement) *ForLoop {
	return ForLoopPattern(PatternFrom(pattern), iterable, Block(statements...))
}

func Brk(label interface{}, value Expression) *BreakStatement {
	return NewBreakStatement(identifierPtr(label), value)
}

func Cont(label interface{}) *ContinueStatement {
	return NewContinueStatement(identifierPtr(label))
}

// Error handling helpers.

func Yield(expr Expression) *YieldStatement {
	return NewYieldStatement(expr)
}

func Raise(expression Expression) *RaiseStatement {
	return NewRaiseStatement(expression)
}

func Rescue(monitored Expression, clauses ...*MatchClause) *RescueExpression {
	return NewRescueExpression(monitored, clauses)
}

func Ensure(expr Expression, statements ...Statement) *EnsureExpression {
	return NewEnsureExpression(expr, Block(statements...))
}

func Rethrow() *RethrowStatement {
	return NewRethrowStatement()
}

// Match / clause helpers reuse MatchClause and Match above.

// Definition helpers.

func FieldDef(fieldType TypeExpression, name interface{}) *StructFieldDefinition {
	return NewStructFieldDefinition(fieldType, identifierPtr(name))
}

func StructDef(name interface{}, fields []*StructFieldDefinition, kind StructKind, generics []*GenericParameter, whereClause []*WhereClauseConstraint, isPrivate bool) *StructDefinition {
	return NewStructDefinition(identifierPtr(name), fields, kind, generics, whereClause, isPrivate)
}

func FieldInit(value Expression, name interface{}) *StructFieldInitializer {
	return NewStructFieldInitializer(value, identifierPtr(name), false)
}

func ShorthandField(name interface{}) *StructFieldInitializer {
	id := identifierPtr(name)
	return NewStructFieldInitializer(id, id, true)
}

func StructLit(fields []*StructFieldInitializer, isPositional bool, structType interface{}, functionalUpdateSources []Expression, typeArgs []TypeExpression) *StructLiteral {
	return NewStructLiteral(fields, isPositional, identifierPtr(structType), functionalUpdateSources, typeArgs)
}

func MapEntry(key, value Expression) *MapLiteralEntry {
	return NewMapLiteralEntry(key, value)
}

func MapSpread(expr Expression) *MapLiteralSpread {
	return NewMapLiteralSpread(expr)
}

func MapLit(elements []MapLiteralElement) *MapLiteral {
	return NewMapLiteral(elements)
}

func UnionDef(name interface{}, variants []TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, isPrivate bool) *UnionDefinition {
	return NewUnionDefinition(identifierPtr(name), variants, generics, whereClause, isPrivate)
}

func Param(name interface{}, paramType TypeExpression) *FunctionParameter {
	return NewFunctionParameter(PatternFrom(name), paramType)
}

func Fn(name interface{}, params []*FunctionParameter, body []Statement, returnType TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, isMethodShorthand, isPrivate bool) *FunctionDefinition {
	return NewFunctionDefinition(identifierPtr(name), params, Block(body...), returnType, generics, whereClause, isMethodShorthand, isPrivate)
}

func FnSig(name interface{}, params []*FunctionParameter, returnType TypeExpression, generics []*GenericParameter, whereClause []*WhereClauseConstraint, defaultImpl *BlockExpression) *FunctionSignature {
	return NewFunctionSignature(identifierPtr(name), params, returnType, generics, whereClause, defaultImpl)
}

func Iface(name interface{}, signatures []*FunctionSignature, generics []*GenericParameter, selfTypePattern TypeExpression, whereClause []*WhereClauseConstraint, baseInterfaces []TypeExpression, isPrivate bool) *InterfaceDefinition {
	return NewInterfaceDefinition(identifierPtr(name), signatures, generics, selfTypePattern, whereClause, baseInterfaces, isPrivate)
}

func Impl(interfaceName interface{}, targetType TypeExpression, definitions []*FunctionDefinition, implName interface{}, generics []*GenericParameter, interfaceArgs []TypeExpression, whereClause []*WhereClauseConstraint, isPrivate bool) *ImplementationDefinition {
	return NewImplementationDefinition(identifierPtr(interfaceName), targetType, definitions, identifierPtr(implName), generics, interfaceArgs, whereClause, isPrivate)
}

func Methods(targetType TypeExpression, definitions []*FunctionDefinition, generics []*GenericParameter, whereClause []*WhereClauseConstraint) *MethodsDefinition {
	return NewMethodsDefinition(targetType, definitions, generics, whereClause)
}

// Packages & imports.

func Pkg(namePath []interface{}, isPrivate bool) *PackageStatement {
	ids := make([]*Identifier, len(namePath))
	for i, v := range namePath {
		ids[i] = identifierPtr(v)
	}
	return NewPackageStatement(ids, isPrivate)
}

func ImpSel(name interface{}, alias interface{}) *ImportSelector {
	return NewImportSelector(identifierPtr(name), identifierPtr(alias))
}

func Imp(packagePath []interface{}, isWildcard bool, selectors []*ImportSelector, alias interface{}) *ImportStatement {
	ids := make([]*Identifier, len(packagePath))
	for i, v := range packagePath {
		ids[i] = identifierPtr(v)
	}
	return NewImportStatement(ids, isWildcard, selectors, identifierPtr(alias))
}

func DynImp(packagePath []interface{}, isWildcard bool, selectors []*ImportSelector, alias interface{}) *DynImportStatement {
	ids := make([]*Identifier, len(packagePath))
	for i, v := range packagePath {
		ids[i] = identifierPtr(v)
	}
	return NewDynImportStatement(ids, isWildcard, selectors, identifierPtr(alias))
}

// Module & statements.

func Mod(body []Statement, imports []*ImportStatement, pkg *PackageStatement) *Module {
	return NewModule(body, imports, pkg)
}

func Ret(argument Expression) *ReturnStatement {
	return NewReturnStatement(argument)
}

// Host interop helpers.

func Prelude(target HostTarget, code string) *PreludeStatement {
	return NewPreludeStatement(target, code)
}

func Extern(target HostTarget, signature *FunctionDefinition, body string) *ExternFunctionBody {
	return NewExternFunctionBody(target, signature, body)
}

// Internal helper utilities.

func identifierPtr(value interface{}) *Identifier {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		return ID(v)
	case *Identifier:
		return v
	default:
		panic("ast: expected string or *Identifier")
	}
}

func memberExpression(member interface{}) Expression {
	switch v := member.(type) {
	case string:
		return ID(v)
	case int:
		return Int(int64(v))
	case int64:
		return Int(v)
	case *Identifier:
		return v
	case Expression:
		return v
	default:
		panic("ast: unsupported member expression")
	}
}
