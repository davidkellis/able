package compiler

import "able/interpreter-go/pkg/ast"

type diagNodeInfo struct {
	Name       string
	GoType     string
	Span       ast.Span
	Origin     string
	CallName   string
	CallMember string
}

type implSiblingInfo struct {
	GoName string
	Arity  int
	Info   *functionInfo
}

type controlFlowResultProbe struct {
	branchTypes     []string
	branchTypeExprs []ast.TypeExpression
	sawNil          bool
}

type nominalCoercionOrigin struct {
	Expr   string
	GoType string
}

type compileContext struct {
	params                 map[string]paramInfo
	locals                 map[string]paramInfo
	integerFacts           map[string]integerFact
	functions              map[string]*functionInfo
	overloads              map[string]*overloadInfo
	packageName            string
	parent                 *compileContext
	temps                  *int
	reason                 string
	loopDepth              int
	loopLabel              string
	loopBreakValueTemp     string
	loopBreakValueType     string
	loopBreakProbe         *controlFlowResultProbe
	rethrowVar             string
	rethrowErrVar          string
	breakpoints            map[string]int
	breakpointGoLabels     map[string]string
	breakpointResultTemps  map[string]string
	breakpointResultTypes  map[string]string
	breakpointResultProbes map[string]*controlFlowResultProbe
	implicitReceiver       paramInfo
	hasImplicitReceiver    bool
	placeholderParams      map[int]paramInfo
	inPlaceholder          bool
	returnType             string
	returnTypeExpr         ast.TypeExpression
	expectedTypeExpr       ast.TypeExpression
	controlMode            string
	controlCaptureVar      string
	controlCaptureLabel    string
	controlCaptureBreak    bool
	rethrowControlVar      string
	genericNames           map[string]struct{}
	typeBindings           map[string]ast.TypeExpression
	implSiblings           map[string]implSiblingInfo
	originExtractions      map[string]string // CSE cache: Able variable name → Go extraction temp
	coercedNominalOrigins  map[string]nominalCoercionOrigin
	analysisOnly           bool
}
