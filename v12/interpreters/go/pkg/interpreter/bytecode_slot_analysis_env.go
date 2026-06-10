package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// slotEligibleBlock checks a block expression for slot eligibility.
func slotEligibleBlock(block *ast.BlockExpression) bool {
	return slotEligibleBlockWithEnv(block, nil)
}

func slotEligibleBlockWithEnv(block *ast.BlockExpression, env *runtime.Environment) bool {
	if block == nil {
		return true
	}
	scopeEnv := runtime.NewEnvironmentWithValueCapacity(env, 0)
	for _, stmt := range block.Body {
		if !slotEligibleStatementWithEnv(stmt, scopeEnv) {
			return false
		}
		slotEligibleRegisterStructDefinition(scopeEnv, stmt)
	}
	return true
}

func slotEligibleRegisterStructDefinition(env *runtime.Environment, stmt ast.Statement) {
	if env == nil || stmt == nil {
		return
	}
	def, ok := stmt.(*ast.StructDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name == "" {
		return
	}
	env.DefineStruct(def.ID.Name, newStructDefinitionValue(def))
}

// slotEligibleStatement checks a statement for conditions that prevent
// slot-indexed locals.
func slotEligibleStatement(stmt ast.Statement) bool {
	return slotEligibleStatementWithEnv(stmt, nil)
}

func slotEligibleStatementWithEnv(stmt ast.Statement, env *runtime.Environment) bool {
	if stmt == nil {
		return true
	}
	switch s := stmt.(type) {
	case *ast.FunctionDefinition:
		// Nested function definitions capture the env; slot variables
		// would be invisible to the nested function's closure.
		return false
	case *ast.MethodsDefinition:
		return false
	case *ast.ImplementationDefinition:
		return false
	case *ast.ForLoop:
		return slotEligibleForLoopWithEnv(s, env)
	case *ast.WhileLoop:
		if s == nil {
			return true
		}
		return slotEligibleExprWithEnv(s.Condition, env) && slotEligibleBlockWithEnv(s.Body, env)
	case *ast.ReturnStatement:
		if s != nil && s.Argument != nil {
			return slotEligibleExprWithEnv(s.Argument, env)
		}
		return true
	case *ast.YieldStatement:
		if s != nil && s.Expression != nil {
			return slotEligibleExprWithEnv(s.Expression, env)
		}
		return true
	case *ast.RaiseStatement:
		if s != nil && s.Expression != nil {
			return slotEligibleExprWithEnv(s.Expression, env)
		}
		return true
	case *ast.RethrowStatement:
		return true
	case *ast.BreakStatement:
		if s != nil && s.Value != nil {
			return slotEligibleExprWithEnv(s.Value, env)
		}
		return true
	case *ast.ContinueStatement:
		return true
	case *ast.StructDefinition, *ast.UnionDefinition, *ast.TypeAliasDefinition:
		return true
	case *ast.InterfaceDefinition:
		return true
	case *ast.ExternFunctionBody:
		return true
	case *ast.ImportStatement, *ast.DynImportStatement:
		return true
	case *ast.PackageStatement, *ast.PreludeStatement:
		return true
	case ast.Expression:
		return slotEligibleExprWithEnv(s, env)
	default:
		return false
	}
}

// slotEligibleForLoop checks a for loop for slot eligibility.
func slotEligibleForLoop(loop *ast.ForLoop) bool {
	return slotEligibleForLoopWithEnv(loop, nil)
}

func slotEligibleForLoopWithEnv(loop *ast.ForLoop, env *runtime.Environment) bool {
	if loop == nil {
		return true
	}
	// The loop pattern must be a simple identifier.
	if _, ok := loop.Pattern.(*ast.Identifier); !ok {
		return false
	}
	return slotEligibleExprWithEnv(loop.Iterable, env) && slotEligibleBlockWithEnv(loop.Body, env)
}

// slotEligibleExpr checks an expression tree for conditions that prevent
// slot-indexed locals.
func slotEligibleExpr(expr ast.Expression) bool {
	return slotEligibleExprWithEnv(expr, nil)
}

func slotEligibleExprWithEnv(expr ast.Expression, env *runtime.Environment) bool {
	if expr == nil {
		return true
	}
	switch n := expr.(type) {
	// --- Bail-out types ---
	case *ast.LambdaExpression:
		return slotEligibleNonCapturingLambda(n)
	case *ast.SpawnExpression:
		return false
	case *ast.IteratorLiteral:
		return false
	case *ast.RescueExpression:
		return false
	case *ast.EnsureExpression:
		return false
	case *ast.BreakpointExpression:
		return false
	case *ast.OrElseExpression:
		return false
	case *ast.StructLiteral:
		return slotEligibleSimpleNamedStructLiteral(n, env)
	case *ast.MapLiteral:
		// Evaluated via tree-walk (evaluateMapLiteral) which can't
		// see slot variables.
		return false

	// --- Leaf types (always eligible) ---
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.CharLiteral,
		*ast.NilLiteral, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.Identifier, *ast.ImplicitMemberExpression:
		return true

	// --- Container types: recurse into children ---
	case *ast.BinaryExpression:
		if n.Operator == "|>" || n.Operator == "|>>" {
			// Pipe expressions evaluate RHS via tree-walk, which
			// can't see slot variables.
			return false
		}
		return slotEligibleExprWithEnv(n.Left, env) && slotEligibleExprWithEnv(n.Right, env)
	case *ast.UnaryExpression:
		return slotEligibleExprWithEnv(n.Operand, env)
	case *ast.AssignmentExpression:
		return slotEligibleAssignmentWithEnv(n, env)
	case *ast.FunctionCall:
		if !slotEligibleExprWithEnv(n.Callee, env) {
			return false
		}
		for _, arg := range n.Arguments {
			if !slotEligibleExprWithEnv(arg, env) {
				return false
			}
		}
		return true
	case *ast.MemberAccessExpression:
		return slotEligibleExprWithEnv(n.Object, env)
	case *ast.IndexExpression:
		return slotEligibleExprWithEnv(n.Object, env) && slotEligibleExprWithEnv(n.Index, env)
	case *ast.BlockExpression:
		return slotEligibleBlockWithEnv(n, env)
	case *ast.IfExpression:
		if !slotEligibleExprWithEnv(n.IfCondition, env) || !slotEligibleBlockWithEnv(n.IfBody, env) {
			return false
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil {
				if !slotEligibleExprWithEnv(clause.Condition, env) || !slotEligibleBlockWithEnv(clause.Body, env) {
					return false
				}
			}
		}
		return slotEligibleBlockWithEnv(n.ElseBody, env)
	case *ast.MatchExpression:
		return slotEligibleMatchExpressionWithEnv(n, env)
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if !slotEligibleExprWithEnv(el, env) {
				return false
			}
		}
		return true
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if !slotEligibleExprWithEnv(part, env) {
				return false
			}
		}
		return true
	case *ast.TypeCastExpression:
		return slotEligibleExprWithEnv(n.Expression, env)
	case *ast.RangeExpression:
		return slotEligibleExprWithEnv(n.Start, env) && slotEligibleExprWithEnv(n.End, env)
	case *ast.PropagationExpression:
		return slotEligibleExprWithEnv(n.Expression, env)
	case *ast.AwaitExpression:
		return slotEligibleExprWithEnv(n.Expression, env)
	case *ast.LoopExpression:
		return slotEligibleBlockWithEnv(n.Body, env)
	default:
		// Unknown expression type: bail out conservatively.
		return false
	}
}

func slotEligibleMatchExpression(n *ast.MatchExpression) bool {
	return slotEligibleMatchExpressionWithEnv(n, nil)
}

func slotEligibleMatchExpressionWithEnv(n *ast.MatchExpression, env *runtime.Environment) bool {
	if n == nil {
		return true
	}
	if !bytecodeCanLowerSlotMatchInEnv(n, env) || !slotEligibleExprWithEnv(n.Subject, env) {
		return false
	}
	for _, clause := range n.Clauses {
		if clause == nil {
			continue
		}
		if !slotEligibleExprWithEnv(clause.Body, env) {
			return false
		}
	}
	return true
}

// slotEligibleAssignment checks an assignment expression for slot eligibility.
func slotEligibleAssignment(n *ast.AssignmentExpression) bool {
	return slotEligibleAssignmentWithEnv(n, nil)
}

func slotEligibleAssignmentWithEnv(n *ast.AssignmentExpression, env *runtime.Environment) bool {
	if n == nil {
		return true
	}
	// Index and member assignments are fine (they don't create local bindings).
	if _, ok := n.Left.(*ast.IndexExpression); ok {
		return slotEligibleExprWithEnv(n.Right, env)
	}
	if _, ok := n.Left.(*ast.MemberAccessExpression); ok {
		return slotEligibleExprWithEnv(n.Right, env)
	}
	if _, ok := n.Left.(*ast.ImplicitMemberExpression); ok {
		return slotEligibleExprWithEnv(n.Right, env)
	}
	// Simple identifier targets (including typed identifier patterns): fine.
	if _, ok := resolveAssignmentTargetName(n.Left); ok {
		return slotEligibleExprWithEnv(n.Right, env)
	}
	// Anything else (destructuring pattern) is a bail-out.
	return false
}

func slotEligibleSimpleNamedStructLiteral(lit *ast.StructLiteral, env *runtime.Environment) bool {
	if lit == nil || env == nil || lit.StructType == nil || lit.StructType.Name == "" {
		return false
	}
	def, ok := env.StructDefinition(lit.StructType.Name)
	if !ok || def == nil || def.Node == nil || !slotEligibleSimpleNamedStructLiteralForDefinition(lit, def.Node) {
		return false
	}
	for _, field := range lit.Fields {
		if !slotEligibleExprWithEnv(field.Value, env) {
			return false
		}
	}
	return true
}

func slotEligibleSimpleNamedStructLiteralForDefinition(lit *ast.StructLiteral, def *ast.StructDefinition) bool {
	if def == nil || !simpleNamedStructLiteralSyntacticEligible(lit) {
		return false
	}
	if isSingletonStructDef(def) {
		return true
	}
	return simpleNamedStructLiteralEligibleForDefinition(lit, def)
}

func blockHasSlotUnsafePlaceholder(block *ast.BlockExpression) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if stmtHasSlotUnsafePlaceholder(stmt) {
			return true
		}
	}
	return false
}

func stmtHasSlotUnsafePlaceholder(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.ForLoop:
		return exprHasSlotUnsafePlaceholder(s.Iterable) || blockHasSlotUnsafePlaceholder(s.Body)
	case *ast.WhileLoop:
		if s == nil {
			return false
		}
		return exprHasSlotUnsafePlaceholder(s.Condition) || blockHasSlotUnsafePlaceholder(s.Body)
	case *ast.ReturnStatement:
		if s == nil {
			return false
		}
		return exprHasSlotUnsafePlaceholder(s.Argument)
	case *ast.YieldStatement:
		if s == nil {
			return false
		}
		return exprHasSlotUnsafePlaceholder(s.Expression)
	case *ast.BreakStatement:
		if s == nil {
			return false
		}
		return exprHasSlotUnsafePlaceholder(s.Value)
	case ast.Expression:
		return exprHasSlotUnsafePlaceholder(s)
	default:
		return false
	}
}

func exprHasSlotUnsafePlaceholder(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch n := expr.(type) {
	case *ast.PlaceholderExpression:
		return true
	case *ast.BinaryExpression:
		return exprHasSlotUnsafePlaceholder(n.Left) || exprHasSlotUnsafePlaceholder(n.Right)
	case *ast.UnaryExpression:
		return exprHasSlotUnsafePlaceholder(n.Operand)
	case *ast.AssignmentExpression:
		return exprHasSlotUnsafePlaceholder(n.Right)
	case *ast.FunctionCall:
		if exprHasSlotUnsafePlaceholder(n.Callee) {
			return true
		}
		for _, arg := range n.Arguments {
			if exprHasSlotUnsafePlaceholder(arg) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		return exprHasSlotUnsafePlaceholder(n.Object)
	case *ast.IndexExpression:
		return exprHasSlotUnsafePlaceholder(n.Object) || exprHasSlotUnsafePlaceholder(n.Index)
	case *ast.BlockExpression:
		return blockHasSlotUnsafePlaceholder(n)
	case *ast.IfExpression:
		if exprHasSlotUnsafePlaceholder(n.IfCondition) || blockHasSlotUnsafePlaceholder(n.IfBody) {
			return true
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil && (exprHasSlotUnsafePlaceholder(clause.Condition) || blockHasSlotUnsafePlaceholder(clause.Body)) {
				return true
			}
		}
		return blockHasSlotUnsafePlaceholder(n.ElseBody)
	case *ast.MatchExpression:
		if exprHasSlotUnsafePlaceholder(n.Subject) {
			return true
		}
		for _, clause := range n.Clauses {
			if clause == nil {
				continue
			}
			if exprHasSlotUnsafePlaceholder(clause.Guard) || exprHasSlotUnsafePlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if exprHasSlotUnsafePlaceholder(el) {
				return true
			}
		}
		return false
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if exprHasSlotUnsafePlaceholder(part) {
				return true
			}
		}
		return false
	case *ast.TypeCastExpression:
		return exprHasSlotUnsafePlaceholder(n.Expression)
	case *ast.RangeExpression:
		return exprHasSlotUnsafePlaceholder(n.Start) || exprHasSlotUnsafePlaceholder(n.End)
	case *ast.PropagationExpression:
		return exprHasSlotUnsafePlaceholder(n.Expression)
	case *ast.AwaitExpression:
		return exprHasSlotUnsafePlaceholder(n.Expression)
	case *ast.LoopExpression:
		return blockHasSlotUnsafePlaceholder(n.Body)
	case *ast.StructLiteral:
		for _, field := range n.Fields {
			if field != nil && exprHasSlotUnsafePlaceholder(field.Value) {
				return true
			}
		}
		for _, src := range n.FunctionalUpdateSources {
			if exprHasSlotUnsafePlaceholder(src) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
