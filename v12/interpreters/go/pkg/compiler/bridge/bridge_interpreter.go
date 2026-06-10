package bridge

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// Interpreter is the dynamic-runtime surface used by compiled code. Keeping
// this contract at the bridge boundary lets static executables use the bridge
// without depending on the concrete tree-walker/bytecode implementation.
type Interpreter interface {
	AddNodeOrigin(node ast.Node, origin string)
	AppendRuntimeCallFrame(err error, call *ast.FunctionCall) error
	ApplyBinaryOperator(op string, left runtime.Value, right runtime.Value) (runtime.Value, error)
	ApplyUnaryOperator(operator string, operand runtime.Value) (runtime.Value, error)
	ArrayElements(arr *runtime.ArrayValue) ([]runtime.Value, error)
	AttachRuntimeContextWithCallStack(err error, node ast.Node, env *runtime.Environment, callNodes []*ast.FunctionCall) error
	AwaitIterable(expr *ast.AwaitExpression, iterable runtime.Value, env *runtime.Environment) (runtime.Value, error)
	CallFunctionIn(value runtime.Value, args []runtime.Value, env *runtime.Environment) (runtime.Value, error)
	CallFunctionInWithCallNode(value runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error)
	CallStaticGenericUnionMember(obj runtime.Value, memberName string, args []runtime.Value, call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, bool, error)
	CallStaticGenericUnionMemberFromCandidates(candidates []runtime.Value, obj runtime.Value, memberName string, args []runtime.Value, call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, bool, error)
	CastValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error)
	CoerceValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error)
	EnsureTypeSatisfiesInterface(subject ast.TypeExpression, ifaceExpr ast.TypeExpression, context string) error
	EvaluateRangeValues(start, end runtime.Value, inclusive bool, env *runtime.Environment) (runtime.Value, error)
	EvaluateStatementIn(stmt ast.Statement, env *runtime.Environment) (runtime.Value, error)
	ExecutorKind() string
	ExpandTypeAliases(expr ast.TypeExpression) ast.TypeExpression
	GlobalEnvironment() *runtime.Environment
	HashMapHashValue(val runtime.Value) (uint64, error)
	HashMapKeysEqual(a, b runtime.Value) (bool, error)
	IndexAssign(obj runtime.Value, idx runtime.Value, value runtime.Value, env *runtime.Environment) (runtime.Value, error)
	IndexGet(obj runtime.Value, idx runtime.Value, env *runtime.Environment) (runtime.Value, error)
	IsErrorValue(val runtime.Value) bool
	IsKnownConstraintTypeName(name string) bool
	IsTruthy(val runtime.Value) bool
	LookupStructDefinition(name string) (*runtime.StructDefinitionValue, bool)
	LookupUnionDefinition(name string) (*runtime.UnionDefinitionValue, bool)
	MakeErrorValue(val runtime.Value, env *runtime.Environment) runtime.ErrorValue
	MatchesType(typeExpr ast.TypeExpression, value runtime.Value) bool
	MemberAssign(obj runtime.Value, member runtime.Value, value runtime.Value, env *runtime.Environment) (runtime.Value, error)
	MemberGet(obj runtime.Value, member runtime.Value, env *runtime.Environment) (runtime.Value, error)
	MemberGetPreferMethods(obj runtime.Value, member runtime.Value, env *runtime.Environment) (runtime.Value, error)
	RaiseValue(value runtime.Value, env *runtime.Environment) error
	PackageEnvironment(name string) *runtime.Environment
	RegisterCompiledFunctionOverload(env *runtime.Environment, name string, paramTypes []ast.TypeExpression, thunk runtime.CompiledThunk) error
	RegisterCompiledImplMethodOverload(interfaceName string, targetType ast.TypeExpression, interfaceArgs []ast.TypeExpression, constraintSig string, implName string, methodName string, paramTypes []ast.TypeExpression, thunk runtime.CompiledThunk) error
	RegisterCompiledImplNamespaceMethod(env *runtime.Environment, implName string, methodName string, paramTypes []ast.TypeExpression, thunk runtime.CompiledThunk) error
	RegisterCompiledMethodOverload(typeName, methodName string, expectsSelf bool, targetType ast.TypeExpression, paramTypes []ast.TypeExpression, thunk runtime.CompiledThunk) error
	RegisterImplementationDefinitionIn(def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error)
	RegisterInterfaceDefinition(name string, def *runtime.InterfaceDefinitionValue)
	RegisterMethodsDefinitionIn(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error)
	RegisterPackageSymbol(pkgName string, name string, val runtime.Value)
	RegisterStaticCallReceiverType(call *ast.FunctionCall, receiverType ast.TypeExpression)
	RegisterTypeAlias(name string, alias *ast.TypeAliasDefinition)
	RegisterUnionDefinition(name string, def *runtime.UnionDefinitionValue)
	ReserveNodeOrigins(capacity int)
	ResolveIteratorValue(iterable runtime.Value, env *runtime.Environment) (*runtime.IteratorValue, error)
	RunCompiledFuture(env *runtime.Environment, task func(*runtime.Environment) (runtime.Value, error)) *runtime.FutureValue
	SeedStructDefinitions(dst *runtime.Environment) int
	SetCompiledImplChecker(checker func(typeName string, interfaceName string) bool)
	SetCompiledInstanceMethodResolver(resolver func(typeName string, methodName string) (runtime.Value, bool))
	SetCompiledInterfaceMemberResolver(resolver func(receiver runtime.Value, methodName string) (runtime.Value, bool))
	SetInterfaceMethodResolver(resolver func(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, bool))
	StandardDivisionByZeroErrorValue() runtime.ErrorValue
	StandardOverflowErrorValue(operation string) runtime.ErrorValue
	StandardShiftOutOfRangeErrorValue(shift int64) runtime.ErrorValue
	Stringify(val runtime.Value, env *runtime.Environment) (string, error)
	TypeExpressionFromValue(value runtime.Value) ast.TypeExpression
}

func materializeBoundaryValue(value runtime.Value) runtime.Value {
	if materializer, ok := value.(runtime.RuntimeValueMaterializer); ok {
		return materializer.MaterializeRuntimeValue()
	}
	return value
}

func materializeBoundaryValues(values []runtime.Value) []runtime.Value {
	for index, value := range values {
		if _, ok := value.(runtime.RuntimeValueMaterializer); !ok {
			continue
		}
		materialized := make([]runtime.Value, len(values))
		copy(materialized, values[:index])
		for remaining := index; remaining < len(values); remaining++ {
			materialized[remaining] = materializeBoundaryValue(values[remaining])
		}
		return materialized
	}
	return values
}
