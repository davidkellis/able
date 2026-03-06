package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// bytecodeFrameLayout describes a slot-indexed local variable layout for a
// function body. When non-nil on a bytecodeProgram, the VM uses a flat
// []Value array instead of map-based Environment lookups for locals.
type bytecodeFrameLayout struct {
	slotCount          int                // total slots needed (params + locals); set after lowering
	paramSlots         int                // number of param slots (always indices 0..paramSlots-1)
	selfCallSlot       int                // reserved slot for recursive self-call fast path; -1 when disabled
	returnType         ast.TypeExpression // declared return type (for coercion on inline return)
	usesImplicitMember bool               // true if body references #member syntax
	needsEnvScopes     bool               // true if body has definitions needing env registration
	selfCallOneArgFast bool               // true when one-arg self-call inline can skip declaration shape checks
	firstParamType     ast.TypeExpression // cached first parameter type for self-call inline checks/coercion
	firstParamSimple   string             // cached simple type name for first parameter (empty for non-simple)
}

// analyzeFrameLayout inspects a function definition and returns a
// bytecodeFrameLayout if the function body is eligible for slot-indexed
// locals. Returns nil if any bail-out condition is detected, in which
// case the function falls back to map-based Environment storage.
func analyzeFrameLayout(def *ast.FunctionDefinition) *bytecodeFrameLayout {
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
	// Walk the body checking for bail-out conditions.
	if !slotEligibleBlock(def.Body) {
		return nil
	}
	var firstParamType ast.TypeExpression
	firstParamSimple := ""
	if len(def.Params) > 0 && def.Params[0] != nil {
		firstParamType = def.Params[0].ParamType
		firstParamSimple = cachedSimpleTypeName(firstParamType)
	}
	return &bytecodeFrameLayout{
		paramSlots:         len(def.Params),
		selfCallSlot:       -1,
		returnType:         def.ReturnType,
		usesImplicitMember: blockUsesImplicitMember(def.Body),
		needsEnvScopes:     blockNeedsEnvScopes(def.Body),
		selfCallOneArgFast: !def.IsMethodShorthand && len(def.Params) == 1 && len(def.GenericParams) == 0,
		firstParamType:     firstParamType,
		firstParamSimple:   firstParamSimple,
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
	case *ast.LoopExpression:
		return blockNeedsEnvScopes(n.Body)
	}
	return false
}

// slotEligibleBlock checks a block expression for slot eligibility.
func slotEligibleBlock(block *ast.BlockExpression) bool {
	if block == nil {
		return true
	}
	for _, stmt := range block.Body {
		if !slotEligibleStatement(stmt) {
			return false
		}
	}
	return true
}

// slotEligibleStatement checks a statement for conditions that prevent
// slot-indexed locals.
func slotEligibleStatement(stmt ast.Statement) bool {
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
		return slotEligibleForLoop(s)
	case *ast.WhileLoop:
		if s == nil {
			return true
		}
		return slotEligibleExpr(s.Condition) && slotEligibleBlock(s.Body)
	case *ast.ReturnStatement:
		if s != nil && s.Argument != nil {
			return slotEligibleExpr(s.Argument)
		}
		return true
	case *ast.YieldStatement:
		if s != nil && s.Expression != nil {
			return slotEligibleExpr(s.Expression)
		}
		return true
	case *ast.RaiseStatement:
		// RaiseStatement uses bytecodeOpRaise which evaluates subexpression
		// via tree-walk (evalExpressionBytecode), so slot variables in the
		// raise expression would not be found. Bail out.
		return false
	case *ast.RethrowStatement:
		return true
	case *ast.BreakStatement:
		if s != nil && s.Value != nil {
			return slotEligibleExpr(s.Value)
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
		return slotEligibleExpr(s)
	default:
		return false
	}
}

// slotEligibleForLoop checks a for loop for slot eligibility.
func slotEligibleForLoop(loop *ast.ForLoop) bool {
	if loop == nil {
		return true
	}
	// The loop pattern must be a simple identifier.
	if _, ok := loop.Pattern.(*ast.Identifier); !ok {
		return false
	}
	return slotEligibleExpr(loop.Iterable) && slotEligibleBlock(loop.Body)
}

// slotEligibleExpr checks an expression tree for conditions that prevent
// slot-indexed locals.
func slotEligibleExpr(expr ast.Expression) bool {
	if expr == nil {
		return true
	}
	switch n := expr.(type) {
	// --- Bail-out types ---
	case *ast.LambdaExpression:
		return false
	case *ast.SpawnExpression:
		return false
	case *ast.MatchExpression:
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
		// Evaluated via tree-walk (evaluateStructLiteral) which can't
		// see slot variables.
		return false
	case *ast.MapLiteral:
		// Evaluated via tree-walk (evaluateMapLiteral) which can't
		// see slot variables.
		return false

	// --- Leaf types (always eligible) ---
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.CharLiteral,
		*ast.NilLiteral, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.Identifier, *ast.ImplicitMemberExpression,
		*ast.PlaceholderExpression:
		return true

	// --- Container types: recurse into children ---
	case *ast.BinaryExpression:
		if n.Operator == "|>" || n.Operator == "|>>" {
			// Pipe expressions evaluate RHS via tree-walk, which
			// can't see slot variables.
			return false
		}
		return slotEligibleExpr(n.Left) && slotEligibleExpr(n.Right)
	case *ast.UnaryExpression:
		return slotEligibleExpr(n.Operand)
	case *ast.AssignmentExpression:
		return slotEligibleAssignment(n)
	case *ast.FunctionCall:
		if !slotEligibleExpr(n.Callee) {
			return false
		}
		for _, arg := range n.Arguments {
			if !slotEligibleExpr(arg) {
				return false
			}
		}
		return true
	case *ast.MemberAccessExpression:
		return slotEligibleExpr(n.Object)
	case *ast.IndexExpression:
		return slotEligibleExpr(n.Object) && slotEligibleExpr(n.Index)
	case *ast.BlockExpression:
		return slotEligibleBlock(n)
	case *ast.IfExpression:
		if !slotEligibleExpr(n.IfCondition) || !slotEligibleBlock(n.IfBody) {
			return false
		}
		for _, clause := range n.ElseIfClauses {
			if clause != nil {
				if !slotEligibleExpr(clause.Condition) || !slotEligibleBlock(clause.Body) {
					return false
				}
			}
		}
		return slotEligibleBlock(n.ElseBody)
	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			if !slotEligibleExpr(el) {
				return false
			}
		}
		return true
	case *ast.StringInterpolation:
		for _, part := range n.Parts {
			if !slotEligibleExpr(part) {
				return false
			}
		}
		return true
	case *ast.TypeCastExpression:
		return slotEligibleExpr(n.Expression)
	case *ast.RangeExpression:
		return slotEligibleExpr(n.Start) && slotEligibleExpr(n.End)
	case *ast.PropagationExpression:
		return slotEligibleExpr(n.Expression)
	case *ast.AwaitExpression:
		return slotEligibleExpr(n.Expression)
	case *ast.LoopExpression:
		return slotEligibleBlock(n.Body)
	default:
		// Unknown expression type: bail out conservatively.
		return false
	}
}

// slotEligibleAssignment checks an assignment expression for slot eligibility.
func slotEligibleAssignment(n *ast.AssignmentExpression) bool {
	if n == nil {
		return true
	}
	// Index and member assignments are fine (they don't create local bindings).
	if _, ok := n.Left.(*ast.IndexExpression); ok {
		return slotEligibleExpr(n.Right)
	}
	if _, ok := n.Left.(*ast.MemberAccessExpression); ok {
		return slotEligibleExpr(n.Right)
	}
	if _, ok := n.Left.(*ast.ImplicitMemberExpression); ok {
		return slotEligibleExpr(n.Right)
	}
	// Simple identifier targets (including typed identifier patterns): fine.
	if _, ok := resolveAssignmentTargetName(n.Left); ok {
		return slotEligibleExpr(n.Right)
	}
	// Anything else (destructuring pattern) is a bail-out.
	return false
}
