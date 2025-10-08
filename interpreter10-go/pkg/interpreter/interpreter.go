package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

const numericEpsilon = 1e-9

// Interpreter drives evaluation of Able v10 AST nodes.
type Interpreter struct {
	global          *runtime.Environment
	inherentMethods map[string]map[string]*runtime.FunctionValue
	interfaces      map[string]*runtime.InterfaceDefinitionValue
	implMethods     map[string][]implEntry
	unnamedImpls    map[string]map[string]map[string]struct{}
	raiseStack      []runtime.Value
	packageRegistry map[string]map[string]runtime.Value
	currentPackage  string
	breakpoints     []string
}

type implEntry struct {
	interfaceName string
	methods       map[string]*runtime.FunctionValue
	definition    *ast.ImplementationDefinition
	argTemplates  []ast.TypeExpression
	genericParams []*ast.GenericParameter
	whereClause   []*ast.WhereClauseConstraint
	unionVariants []string
	defaultOnly   bool
}

type implCandidate struct {
	entry       *implEntry
	bindings    map[string]ast.TypeExpression
	constraints []constraintSpec
	score       int
}

type methodMatch struct {
	candidate implCandidate
	method    *runtime.FunctionValue
}

type constraintSpec struct {
	typeParam string
	ifaceType ast.TypeExpression
}

type typeInfo struct {
	name     string
	typeArgs []ast.TypeExpression
}

type targetVariant struct {
	typeName     string
	argTemplates []ast.TypeExpression
	signature    string
}

func expandImplementationTargetVariants(target ast.TypeExpression) ([]targetVariant, []string, error) {
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return nil, nil, fmt.Errorf("Implementation target requires identifier")
		}
		signature := typeExpressionToString(t)
		return []targetVariant{{typeName: t.Name.Name, argTemplates: nil, signature: signature}}, nil, nil
	case *ast.GenericTypeExpression:
		simple, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || simple.Name == nil {
			return nil, nil, fmt.Errorf("Implementation target requires simple base type")
		}
		signature := typeExpressionToString(t)
		return []targetVariant{{
			typeName:     simple.Name.Name,
			argTemplates: append([]ast.TypeExpression(nil), t.Arguments...),
			signature:    signature,
		}}, nil, nil
	case *ast.UnionTypeExpression:
		var variants []targetVariant
		signatureSet := make(map[string]struct{})
		for _, member := range t.Members {
			childVariants, childSigs, err := expandImplementationTargetVariants(member)
			if err != nil {
				return nil, nil, err
			}
			for _, v := range childVariants {
				if _, seen := signatureSet[v.signature]; seen {
					continue
				}
				signatureSet[v.signature] = struct{}{}
				variants = append(variants, v)
			}
			for _, sig := range childSigs {
				signatureSet[sig] = struct{}{}
			}
		}
		if len(variants) == 0 {
			return nil, nil, fmt.Errorf("Union target must contain at least one concrete type")
		}
		// Build union signature list for penalty/book-keeping
		unionSigs := make([]string, 0, len(signatureSet))
		for sig := range signatureSet {
			unionSigs = append(unionSigs, sig)
		}
		sort.Strings(unionSigs)
		return variants, unionSigs, nil
	default:
		return nil, nil, fmt.Errorf("Implementation target type %T is not supported", target)
	}
}

func identifiersToStrings(ids []*ast.Identifier) []string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		parts = append(parts, id.Name)
	}
	return parts
}

func joinIdentifierNames(ids []*ast.Identifier) string {
	parts := identifiersToStrings(ids)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ".")
}

func (i *Interpreter) qualifiedName(name string) string {
	if i.currentPackage == "" {
		return ""
	}
	return i.currentPackage + "." + name
}

func (i *Interpreter) pushBreakpoint(label string) {
	i.breakpoints = append(i.breakpoints, label)
}

func (i *Interpreter) popBreakpoint() {
	if len(i.breakpoints) == 0 {
		return
	}
	i.breakpoints = i.breakpoints[:len(i.breakpoints)-1]
}

func (i *Interpreter) hasBreakpoint(label string) bool {
	for idx := len(i.breakpoints) - 1; idx >= 0; idx-- {
		if i.breakpoints[idx] == label {
			return true
		}
	}
	return false
}

func (i *Interpreter) registerSymbol(name string, value runtime.Value) {
	if i.currentPackage == "" {
		return
	}
	bucket, ok := i.packageRegistry[i.currentPackage]
	if !ok {
		bucket = make(map[string]runtime.Value)
		i.packageRegistry[i.currentPackage] = bucket
	}
	bucket[name] = value
	if qn := i.qualifiedName(name); qn != "" {
		i.global.Define(qn, value)
	}
}

func isPrivateSymbol(val runtime.Value) bool {
	switch v := val.(type) {
	case *runtime.FunctionValue:
		if fn, ok := v.Declaration.(*ast.FunctionDefinition); ok {
			return fn.IsPrivate
		}
	case *runtime.StructDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.StructDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case *runtime.InterfaceDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.InterfaceDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case *runtime.UnionDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.UnionDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	}
	return false
}

func importPrivacyError(name string, val runtime.Value) error {
	switch v := val.(type) {
	case *runtime.FunctionValue:
		if fn, ok := v.Declaration.(*ast.FunctionDefinition); ok && fn.IsPrivate {
			return fmt.Errorf("Import error: function '%s' is private", name)
		}
	case *runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: struct '%s' is private", name)
		}
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: struct '%s' is private", name)
		}
	case *runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: interface '%s' is private", name)
		}
	case runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: interface '%s' is private", name)
		}
	case *runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: union '%s' is private", name)
		}
	case runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: union '%s' is private", name)
		}
	}
	return fmt.Errorf("Import error: symbol '%s' is private", name)
}

func dynImportPrivacyError(name string, val runtime.Value) error {
	switch v := val.(type) {
	case *runtime.FunctionValue:
		if fn, ok := v.Declaration.(*ast.FunctionDefinition); ok && fn.IsPrivate {
			return fmt.Errorf("dynimport error: function '%s' is private", name)
		}
	case *runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: struct '%s' is private", name)
		}
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: struct '%s' is private", name)
		}
	case *runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: interface '%s' is private", name)
		}
	case runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: interface '%s' is private", name)
		}
	case *runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: union '%s' is private", name)
		}
	case runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: union '%s' is private", name)
		}
	}
	return fmt.Errorf("dynimport error: symbol '%s' is private", name)
}

func copyPublicSymbols(bucket map[string]runtime.Value) map[string]runtime.Value {
	public := make(map[string]runtime.Value)
	for name, val := range bucket {
		if isPrivateSymbol(val) {
			continue
		}
		public[name] = val
	}
	return public
}

// New returns an interpreter with an empty global environment.
func New() *Interpreter {
	return &Interpreter{
		global:          runtime.NewEnvironment(nil),
		inherentMethods: make(map[string]map[string]*runtime.FunctionValue),
		interfaces:      make(map[string]*runtime.InterfaceDefinitionValue),
		implMethods:     make(map[string][]implEntry),
		unnamedImpls:    make(map[string]map[string]map[string]struct{}),
		raiseStack:      make([]runtime.Value, 0),
		packageRegistry: make(map[string]map[string]runtime.Value),
		breakpoints:     make([]string, 0),
	}
}

// GlobalEnvironment returns the interpreter’s global environment.
func (i *Interpreter) GlobalEnvironment() *runtime.Environment {
	return i.global
}

// EvaluateModule executes a module node and returns the last evaluated value and environment.
func (i *Interpreter) EvaluateModule(module *ast.Module) (runtime.Value, *runtime.Environment, error) {
	moduleEnv := i.global
	prevPackage := i.currentPackage
	defer func() { i.currentPackage = prevPackage }()

	if module.Package != nil {
		moduleEnv = runtime.NewEnvironment(i.global)
		pkgName := joinIdentifierNames(module.Package.NamePath)
		i.currentPackage = pkgName
		if _, ok := i.packageRegistry[pkgName]; !ok {
			i.packageRegistry[pkgName] = make(map[string]runtime.Value)
		}
	} else {
		i.currentPackage = ""
	}

	for _, imp := range module.Imports {
		if _, err := i.evaluateImportStatement(imp, moduleEnv); err != nil {
			return nil, nil, err
		}
	}

	var last runtime.Value = runtime.NilValue{}
	for _, stmt := range module.Body {
		val, err := i.evaluateStatement(stmt, moduleEnv)
		if err != nil {
			if rs, ok := err.(raiseSignal); ok {
				return nil, moduleEnv, rs
			}
			if _, ok := err.(returnSignal); ok {
				return nil, nil, fmt.Errorf("return outside function")
			}
			return nil, nil, err
		}
		last = val
	}
	return last, moduleEnv, nil
}

// evaluateStatement currently delegates to expression evaluation for expressions.
func (i *Interpreter) evaluateStatement(node ast.Statement, env *runtime.Environment) (runtime.Value, error) {
	switch n := node.(type) {
	case ast.Expression:
		return i.evaluateExpression(n, env)
	case *ast.StructDefinition:
		return i.evaluateStructDefinition(n, env)
	case *ast.MethodsDefinition:
		return i.evaluateMethodsDefinition(n, env)
	case *ast.InterfaceDefinition:
		return i.evaluateInterfaceDefinition(n, env)
	case *ast.ImplementationDefinition:
		return i.evaluateImplementationDefinition(n, env)
	case *ast.FunctionDefinition:
		return i.evaluateFunctionDefinition(n, env)
	case *ast.WhileLoop:
		return i.evaluateWhileLoop(n, env)
	case *ast.ForLoop:
		return i.evaluateForLoop(n, env)
	case *ast.RaiseStatement:
		return i.evaluateRaiseStatement(n, env)
	case *ast.BreakStatement:
		return i.evaluateBreakStatement(n, env)
	case *ast.ContinueStatement:
		return i.evaluateContinueStatement(n, env)
	case *ast.ReturnStatement:
		return i.evaluateReturnStatement(n, env)
	case *ast.RethrowStatement:
		return i.evaluateRethrowStatement(n, env)
	case *ast.PackageStatement:
		return runtime.NilValue{}, nil
	case *ast.ImportStatement:
		return i.evaluateImportStatement(n, env)
	case *ast.DynImportStatement:
		return i.evaluateDynImportStatement(n, env)
	default:
		return nil, fmt.Errorf("unsupported statement type: %s", n.NodeType())
	}
}

// evaluateExpression handles literals, identifiers, and blocks.
func (i *Interpreter) evaluateExpression(node ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	switch n := node.(type) {
	case *ast.StringLiteral:
		return runtime.StringValue{Val: n.Value}, nil
	case *ast.BooleanLiteral:
		return runtime.BoolValue{Val: n.Value}, nil
	case *ast.CharLiteral:
		if len(n.Value) == 0 {
			return nil, fmt.Errorf("empty char literal")
		}
		return runtime.CharValue{Val: []rune(n.Value)[0]}, nil
	case *ast.NilLiteral:
		return runtime.NilValue{}, nil
	case *ast.IntegerLiteral:
		suffix := runtime.IntegerI32
		if n.IntegerType != nil {
			suffix = runtime.IntegerType(*n.IntegerType)
		}
		return runtime.IntegerValue{Val: runtime.CloneBigInt(bigFromLiteral(n.Value)), TypeSuffix: suffix}, nil
	case *ast.FloatLiteral:
		suffix := runtime.FloatF64
		if n.FloatType != nil {
			suffix = runtime.FloatType(*n.FloatType)
		}
		return runtime.FloatValue{Val: n.Value, TypeSuffix: suffix}, nil
	case *ast.ArrayLiteral:
		values := make([]runtime.Value, 0, len(n.Elements))
		for _, el := range n.Elements {
			val, err := i.evaluateExpression(el, env)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		return &runtime.ArrayValue{Elements: values}, nil
	case *ast.StringInterpolation:
		var builder strings.Builder
		for _, part := range n.Parts {
			if part == nil {
				return nil, fmt.Errorf("string interpolation contains nil part")
			}
			if lit, ok := part.(*ast.StringLiteral); ok {
				builder.WriteString(lit.Value)
				continue
			}
			val, err := i.evaluateExpression(part, env)
			if err != nil {
				return nil, err
			}
			str, err := i.stringifyValue(val, env)
			if err != nil {
				return nil, err
			}
			builder.WriteString(str)
		}
		return runtime.StringValue{Val: builder.String()}, nil
	case *ast.BreakpointExpression:
		return i.evaluateBreakpointExpression(n, env)
	case *ast.RangeExpression:
		start, err := i.evaluateExpression(n.Start, env)
		if err != nil {
			return nil, err
		}
		endExpr, err := i.evaluateExpression(n.End, env)
		if err != nil {
			return nil, err
		}
		if !isNumericValue(start) || !isNumericValue(endExpr) {
			return nil, fmt.Errorf("Range boundaries must be numeric")
		}
		return &runtime.RangeValue{Start: start, End: endExpr, Inclusive: n.Inclusive}, nil
	case *ast.StructLiteral:
		return i.evaluateStructLiteral(n, env)
	case *ast.MatchExpression:
		return i.evaluateMatchExpression(n, env)
	case *ast.PropagationExpression:
		return i.evaluatePropagationExpression(n, env)
	case *ast.OrElseExpression:
		return i.evaluateOrElseExpression(n, env)
	case *ast.EnsureExpression:
		return i.evaluateEnsureExpression(n, env)
	case *ast.MemberAccessExpression:
		return i.evaluateMemberAccess(n, env)
	case *ast.IndexExpression:
		return i.evaluateIndexExpression(n, env)
	case *ast.UnaryExpression:
		return i.evaluateUnaryExpression(n, env)
	case *ast.Identifier:
		val, err := env.Get(n.Name)
		if err != nil {
			return nil, err
		}
		return val, nil
	case *ast.FunctionCall:
		return i.evaluateFunctionCall(n, env)
	case *ast.BinaryExpression:
		return i.evaluateBinaryExpression(n, env)
	case *ast.AssignmentExpression:
		return i.evaluateAssignment(n, env)
	case *ast.BlockExpression:
		return i.evaluateBlock(n, env)
	case *ast.IfExpression:
		return i.evaluateIfExpression(n, env)
	case *ast.RescueExpression:
		return i.evaluateRescueExpression(n, env)
	default:
		return nil, fmt.Errorf("unsupported expression type: %s", n.NodeType())
	}
}

func (i *Interpreter) evaluateBlock(block *ast.BlockExpression, env *runtime.Environment) (runtime.Value, error) {
	scope := runtime.NewEnvironment(env)
	var result runtime.Value = runtime.NilValue{}
	for _, stmt := range block.Body {
		val, err := i.evaluateStatement(stmt, scope)
		if err != nil {
			if _, ok := err.(returnSignal); ok {
				return nil, err
			}
			return nil, err
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) evaluateWhileLoop(loop *ast.WhileLoop, env *runtime.Environment) (runtime.Value, error) {
	var result runtime.Value = runtime.NilValue{}
	for {
		cond, err := i.evaluateExpression(loop.Condition, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			return result, nil
		}
		val, err := i.evaluateBlock(loop.Body, env)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label != "" {
					return nil, sig
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("Labeled continue not supported")
				}
				continue
			case raiseSignal:
				return nil, sig
			case returnSignal:
				return nil, sig
			default:
				return nil, err
			}
		}
		result = val
	}
}

func (i *Interpreter) evaluateRaiseStatement(stmt *ast.RaiseStatement, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(stmt.Expression, env)
	if err != nil {
		return nil, err
	}
	errVal := makeErrorValue(val)
	return nil, raiseSignal{value: errVal}
}

func (i *Interpreter) evaluateReturnStatement(stmt *ast.ReturnStatement, env *runtime.Environment) (runtime.Value, error) {
	var result runtime.Value = runtime.NilValue{}
	if stmt.Argument != nil {
		val, err := i.evaluateExpression(stmt.Argument, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return nil, returnSignal{value: result}
}

func (i *Interpreter) evaluateForLoop(loop *ast.ForLoop, env *runtime.Environment) (runtime.Value, error) {
	iterable, err := i.evaluateExpression(loop.Iterable, env)
	if err != nil {
		return nil, err
	}
	bodyEnvBase := runtime.NewEnvironment(env)

	var values []runtime.Value
	switch it := iterable.(type) {
	case *runtime.ArrayValue:
		values = it.Elements
	case *runtime.RangeValue:
		startVal, err := rangeEndpoint(it.Start)
		if err != nil {
			return nil, err
		}
		endVal, err := rangeEndpoint(it.End)
		if err != nil {
			return nil, err
		}
		step := 1
		if endVal < startVal {
			step = -1
		}
		values = make([]runtime.Value, 0)
		for v := startVal; ; v += step {
			if step > 0 {
				if it.Inclusive {
					if v > endVal {
						break
					}
				} else if v >= endVal {
					break
				}
			} else {
				if it.Inclusive {
					if v < endVal {
						break
					}
				} else if v <= endVal {
					break
				}
			}
			values = append(values, runtime.IntegerValue{Val: big.NewInt(int64(v)), TypeSuffix: runtime.IntegerI32})
		}
	default:
		return nil, fmt.Errorf("for-loop iterable must be array or range, got %s", iterable.Kind())
	}

	var result runtime.Value = runtime.NilValue{}
	for _, el := range values {
		iterEnv := runtime.NewEnvironment(bodyEnvBase)
		if err := i.assignPattern(loop.Pattern, el, iterEnv, true); err != nil {
			return nil, err
		}
		val, err := i.evaluateBlock(loop.Body, iterEnv)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label != "" {
					return nil, sig
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("Labeled continue not supported")
				}
				continue
			case raiseSignal:
				return nil, sig
			case returnSignal:
				return nil, sig
			default:
				return nil, err
			}
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) evaluateBreakStatement(stmt *ast.BreakStatement, env *runtime.Environment) (runtime.Value, error) {
	var val runtime.Value = runtime.NilValue{}
	if stmt.Value != nil {
		var err error
		val, err = i.evaluateExpression(stmt.Value, env)
		if err != nil {
			return nil, err
		}
	}
	label := ""
	if stmt.Label != nil {
		label = stmt.Label.Name
		if !i.hasBreakpoint(label) {
			return nil, fmt.Errorf("Unknown break label '%s'", label)
		}
	}
	return nil, breakSignal{label: label, value: val}
}

func (i *Interpreter) evaluateContinueStatement(stmt *ast.ContinueStatement, env *runtime.Environment) (runtime.Value, error) {
	label := ""
	if stmt.Label != nil {
		label = stmt.Label.Name
		return nil, fmt.Errorf("Labeled continue not supported")
	}
	return nil, continueSignal{label: label}
}

func (i *Interpreter) evaluateBreakpointExpression(expr *ast.BreakpointExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr.Label == nil {
		return nil, fmt.Errorf("Breakpoint expression requires label")
	}
	label := expr.Label.Name
	i.pushBreakpoint(label)
	defer i.popBreakpoint()
	for {
		val, err := i.evaluateBlock(expr.Body, env)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label == label {
					return sig.value, nil
				}
				return nil, sig
			case continueSignal:
				if sig.label == label {
					continue
				}
				return nil, sig
			default:
				return nil, err
			}
		}
		if val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}
}

func (i *Interpreter) evaluateAssignment(assign *ast.AssignmentExpression, env *runtime.Environment) (runtime.Value, error) {
	value, err := i.evaluateExpression(assign.Right, env)
	if err != nil {
		return nil, err
	}
	binaryOp, isCompound := binaryOpForAssignment(assign.Operator)

	switch lhs := assign.Left.(type) {
	case *ast.Identifier:
		switch assign.Operator {
		case ast.AssignmentDeclare:
			env.Define(lhs.Name, value)
		case ast.AssignmentAssign:
			if err := env.Assign(lhs.Name, value); err != nil {
				return nil, err
			}
		default:
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			current, err := env.Get(lhs.Name)
			if err != nil {
				return nil, err
			}
			computed, err := applyBinaryOperator(binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			if err := env.Assign(lhs.Name, computed); err != nil {
				return nil, err
			}
			return computed, nil
		}
	case *ast.MemberAccessExpression:
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on member access")
		}
		target, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		switch inst := target.(type) {
		case *runtime.StructInstanceValue:
			switch member := lhs.Member.(type) {
			case *ast.Identifier:
				if inst.Fields == nil {
					return nil, fmt.Errorf("Expected named struct instance")
				}
				current, ok := inst.Fields[member.Name]
				if !ok {
					return nil, fmt.Errorf("No field named '%s'", member.Name)
				}
				if assign.Operator == ast.AssignmentAssign {
					inst.Fields[member.Name] = value
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				computed, err := applyBinaryOperator(binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				inst.Fields[member.Name] = computed
				return computed, nil
			case *ast.IntegerLiteral:
				if inst.Positional == nil {
					return nil, fmt.Errorf("Expected positional struct instance")
				}
				if member.Value == nil {
					return nil, fmt.Errorf("Struct field index out of bounds")
				}
				idx := int(member.Value.Int64())
				if idx < 0 || idx >= len(inst.Positional) {
					return nil, fmt.Errorf("Struct field index out of bounds")
				}
				if assign.Operator == ast.AssignmentAssign {
					inst.Positional[idx] = value
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				current := inst.Positional[idx]
				computed, err := applyBinaryOperator(binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				inst.Positional[idx] = computed
				return computed, nil
			default:
				return nil, fmt.Errorf("Unsupported member assignment target %s", member.NodeType())
			}
		case *runtime.ArrayValue:
			arrayVal := inst
			intMember, ok := lhs.Member.(*ast.IntegerLiteral)
			if !ok {
				return nil, fmt.Errorf("Array member assignment requires integer member")
			}
			if intMember.Value == nil {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			idx := int(intMember.Value.Int64())
			if idx < 0 || idx >= len(arrayVal.Elements) {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			if assign.Operator == ast.AssignmentAssign {
				arrayVal.Elements[idx] = value
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			current := arrayVal.Elements[idx]
			computed, err := applyBinaryOperator(binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			arrayVal.Elements[idx] = computed
			return computed, nil
		default:
			return nil, fmt.Errorf("Member assignment requires struct or array")
		}
	case *ast.IndexExpression:
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on index assignment")
		}
		arrObj, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		arr, err := toArrayValue(arrObj)
		if err != nil {
			return nil, err
		}
		idxVal, err := i.evaluateExpression(lhs.Index, env)
		if err != nil {
			return nil, err
		}
		idx, err := indexFromValue(idxVal)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx >= len(arr.Elements) {
			return nil, fmt.Errorf("Array index out of bounds")
		}
		if assign.Operator == ast.AssignmentAssign {
			arr.Elements[idx] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		current := arr.Elements[idx]
		computed, err := applyBinaryOperator(binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		arr.Elements[idx] = computed
		return computed, nil
	case ast.Pattern:
		if isCompound {
			return nil, fmt.Errorf("compound assignment not supported with patterns")
		}
		isDeclaration := assign.Operator == ast.AssignmentDeclare
		if !isDeclaration && assign.Operator != ast.AssignmentAssign {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		if err := i.assignPattern(lhs, value, env, isDeclaration); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported assignment target %s", lhs.NodeType())
	}

	return value, nil
}

func binaryOpForAssignment(op ast.AssignmentOperator) (string, bool) {
	switch op {
	case ast.AssignmentAdd:
		return "+", true
	case ast.AssignmentSub:
		return "-", true
	case ast.AssignmentMul:
		return "*", true
	case ast.AssignmentDiv:
		return "/", true
	case ast.AssignmentMod:
		return "%", true
	case ast.AssignmentBitAnd:
		return "&", true
	case ast.AssignmentBitOr:
		return "|", true
	case ast.AssignmentBitXor:
		return "^", true
	case ast.AssignmentShiftL:
		return "<<", true
	case ast.AssignmentShiftR:
		return ">>", true
	default:
		return "", false
	}
}

func (i *Interpreter) evaluateIfExpression(expr *ast.IfExpression, env *runtime.Environment) (runtime.Value, error) {
	cond, err := i.evaluateExpression(expr.IfCondition, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return i.evaluateBlock(expr.IfBody, env)
	}
	for _, clause := range expr.OrClauses {
		if clause.Condition != nil {
			clauseCond, err := i.evaluateExpression(clause.Condition, env)
			if err != nil {
				return nil, err
			}
			if !isTruthy(clauseCond) {
				continue
			}
		}
		return i.evaluateBlock(clause.Body, env)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateMatchExpression(expr *ast.MatchExpression, env *runtime.Environment) (runtime.Value, error) {
	subject, err := i.evaluateExpression(expr.Subject, env)
	if err != nil {
		return nil, err
	}
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseEnv, matched := i.matchPattern(clause.Pattern, subject, env)
		if !matched {
			continue
		}
		if clause.Guard != nil {
			guardVal, err := i.evaluateExpression(clause.Guard, clauseEnv)
			if err != nil {
				return nil, err
			}
			if !isTruthy(guardVal) {
				continue
			}
		}
		return i.evaluateExpression(clause.Body, clauseEnv)
	}
	return nil, fmt.Errorf("Non-exhaustive match")
}

func (i *Interpreter) evaluateRescueExpression(expr *ast.RescueExpression, env *runtime.Environment) (runtime.Value, error) {
	result, err := i.evaluateExpression(expr.MonitoredExpression, env)
	if err == nil {
		return result, nil
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		return nil, err
	}
	for _, clause := range expr.Clauses {
		clauseEnv, matched := i.matchPattern(clause.Pattern, rs.value, env)
		if !matched {
			continue
		}
		i.raiseStack = append(i.raiseStack, rs.value)
		if clause.Guard != nil {
			guardVal, gErr := i.evaluateExpression(clause.Guard, clauseEnv)
			if gErr != nil {
				i.raiseStack = i.raiseStack[:len(i.raiseStack)-1]
				return nil, gErr
			}
			if !isTruthy(guardVal) {
				i.raiseStack = i.raiseStack[:len(i.raiseStack)-1]
				continue
			}
		}
		result, bodyErr := i.evaluateExpression(clause.Body, clauseEnv)
		i.raiseStack = i.raiseStack[:len(i.raiseStack)-1]
		if bodyErr != nil {
			return nil, bodyErr
		}
		return result, nil
	}
	return nil, rs
}

func (i *Interpreter) evaluatePropagationExpression(expr *ast.PropagationExpression, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		return nil, err
	}
	if errVal, ok := val.(runtime.ErrorValue); ok {
		return nil, raiseSignal{value: errVal}
	}
	return val, nil
}

func (i *Interpreter) evaluateOrElseExpression(expr *ast.OrElseExpression, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		if rs, ok := err.(raiseSignal); ok {
			handlerEnv := runtime.NewEnvironment(env)
			if expr.ErrorBinding != nil {
				handlerEnv.Define(expr.ErrorBinding.Name, rs.value)
			}
			result, handlerErr := i.evaluateBlock(expr.Handler, handlerEnv)
			if handlerErr != nil {
				return nil, handlerErr
			}
			if result == nil {
				return runtime.NilValue{}, nil
			}
			return result, nil
		}
		return nil, err
	}
	return val, nil
}

func (i *Interpreter) evaluateEnsureExpression(expr *ast.EnsureExpression, env *runtime.Environment) (runtime.Value, error) {
	var tryResult runtime.Value = runtime.NilValue{}
	val, err := i.evaluateExpression(expr.TryExpression, env)
	if err == nil {
		if val != nil {
			tryResult = val
		}
	}
	if expr.EnsureBlock != nil {
		if _, ensureErr := i.evaluateBlock(expr.EnsureBlock, env); ensureErr != nil {
			return nil, ensureErr
		}
	}
	if err != nil {
		return nil, err
	}
	if tryResult == nil {
		return runtime.NilValue{}, nil
	}
	return tryResult, nil
}

func (i *Interpreter) evaluateImportStatement(imp *ast.ImportStatement, env *runtime.Environment) (runtime.Value, error) {
	return i.processImport(imp.PackagePath, imp.IsWildcard, imp.Selectors, imp.Alias, env, false)
}

func (i *Interpreter) evaluateDynImportStatement(imp *ast.DynImportStatement, env *runtime.Environment) (runtime.Value, error) {
	return i.processImport(imp.PackagePath, imp.IsWildcard, imp.Selectors, imp.Alias, env, true)
}

func (i *Interpreter) processImport(packagePath []*ast.Identifier, isWildcard bool, selectors []*ast.ImportSelector, alias *ast.Identifier, env *runtime.Environment, dynamic bool) (runtime.Value, error) {
	pkgParts := identifiersToStrings(packagePath)
	pkgName := strings.Join(pkgParts, ".")

	if dynamic {
		return i.processDynImport(pkgName, pkgParts, isWildcard, selectors, alias, env)
	}

	if alias != nil && !isWildcard && len(selectors) == 0 {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		public := copyPublicSymbols(bucket)
		env.Define(alias.Name, runtime.PackageValue{NamePath: pkgParts, Public: public})
		return runtime.NilValue{}, nil
	}

	if isWildcard {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		for name, val := range bucket {
			if isPrivateSymbol(val) {
				continue
			}
			env.Define(name, val)
		}
		return runtime.NilValue{}, nil
	}

	if len(selectors) > 0 {
		for _, sel := range selectors {
			if sel == nil || sel.Name == nil {
				return nil, fmt.Errorf("Import selector missing name")
			}
			original := sel.Name.Name
			aliasName := original
			if sel.Alias != nil {
				aliasName = sel.Alias.Name
			}
			val, err := i.lookupImportSymbol(pkgName, original)
			if err != nil {
				return nil, err
			}
			if isPrivateSymbol(val) {
				return nil, importPrivacyError(original, val)
			}
			env.Define(aliasName, val)
		}
		return runtime.NilValue{}, nil
	}

	if pkgName != "" && alias == nil {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		public := copyPublicSymbols(bucket)
		env.Define(pkgName, runtime.PackageValue{NamePath: pkgParts, Public: public})
	}

	return runtime.NilValue{}, nil
}

func (i *Interpreter) processDynImport(pkgName string, pkgParts []string, isWildcard bool, selectors []*ast.ImportSelector, alias *ast.Identifier, env *runtime.Environment) (runtime.Value, error) {
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dynimport error: package '%s' not found", pkgName)
	}

	if alias != nil && !isWildcard && len(selectors) == 0 {
		env.Define(alias.Name, runtime.DynPackageValue{NamePath: pkgParts, Name: pkgName})
		return runtime.NilValue{}, nil
	}

	if isWildcard {
		for name, val := range bucket {
			if isPrivateSymbol(val) {
				continue
			}
			env.Define(name, runtime.DynRefValue{Package: pkgName, Name: name})
		}
		return runtime.NilValue{}, nil
	}

	if len(selectors) > 0 {
		for _, sel := range selectors {
			if sel == nil || sel.Name == nil {
				return nil, fmt.Errorf("dynimport selector missing name")
			}
			original := sel.Name.Name
			aliasName := original
			if sel.Alias != nil {
				aliasName = sel.Alias.Name
			}
			sym, ok := bucket[original]
			if !ok {
				return nil, fmt.Errorf("dynimport error: '%s' not found in '%s'", original, pkgName)
			}
			if isPrivateSymbol(sym) {
				return nil, dynImportPrivacyError(original, sym)
			}
			env.Define(aliasName, runtime.DynRefValue{Package: pkgName, Name: original})
		}
		return runtime.NilValue{}, nil
	}

	if pkgName != "" && alias == nil {
		env.Define(pkgName, runtime.DynPackageValue{NamePath: pkgParts, Name: pkgName})
	}

	return runtime.NilValue{}, nil
}

func (i *Interpreter) lookupImportSymbol(pkgName, symbol string) (runtime.Value, error) {
	if pkgName != "" {
		if bucket, ok := i.packageRegistry[pkgName]; ok {
			if val, ok := bucket[symbol]; ok {
				return val, nil
			}
		}
		if val, err := i.global.Get(pkgName + "." + symbol); err == nil {
			return val, nil
		}
	}
	if val, err := i.global.Get(symbol); err == nil {
		return val, nil
	}
	if pkgName != "" {
		return nil, fmt.Errorf("Import error: symbol '%s' from '%s' not found in globals", symbol, pkgName)
	}
	return nil, fmt.Errorf("Import error: symbol '%s' not found in globals", symbol)
}

func (i *Interpreter) evaluateRethrowStatement(_ *ast.RethrowStatement, env *runtime.Environment) (runtime.Value, error) {
	_ = env
	if len(i.raiseStack) == 0 {
		return nil, fmt.Errorf("rethrow outside rescue")
	}
	current := i.raiseStack[len(i.raiseStack)-1]
	return nil, raiseSignal{value: current}
}

func (i *Interpreter) evaluateUnaryExpression(expr *ast.UnaryExpression, env *runtime.Environment) (runtime.Value, error) {
	operand, err := i.evaluateExpression(expr.Operand, env)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case "-":
		switch v := operand.(type) {
		case runtime.IntegerValue:
			neg := new(big.Int).Neg(v.Val)
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, nil
		case runtime.FloatValue:
			return runtime.FloatValue{Val: -v.Val, TypeSuffix: v.TypeSuffix}, nil
		default:
			return nil, fmt.Errorf("unary '-' not supported for %T", operand)
		}
	case "!":
		if bv, ok := operand.(runtime.BoolValue); ok {
			return runtime.BoolValue{Val: !bv.Val}, nil
		}
		return nil, fmt.Errorf("unary '!' expects bool, got %T", operand)
	default:
		return nil, fmt.Errorf("unsupported unary operator %s", expr.Operator)
	}
}

func (i *Interpreter) evaluateBinaryExpression(expr *ast.BinaryExpression, env *runtime.Environment) (runtime.Value, error) {
	leftVal, err := i.evaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case "&&":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		if !lb.Val {
			return runtime.BoolValue{Val: false}, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	case "||":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		if lb.Val {
			return runtime.BoolValue{Val: true}, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	default:
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		if expr.Operator == "+" {
			if ls, ok := leftVal.(runtime.StringValue); ok {
				rs, ok := rightVal.(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("Arithmetic requires numeric operands")
				}
				return runtime.StringValue{Val: ls.Val + rs.Val}, nil
			}
			if _, ok := rightVal.(runtime.StringValue); ok {
				return nil, fmt.Errorf("Arithmetic requires numeric operands")
			}
		}
		return applyBinaryOperator(expr.Operator, leftVal, rightVal)
	}
}

func applyBinaryOperator(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	switch op {
	case "+", "-", "*", "/":
		return evaluateArithmetic(op, left, right)
	case "%":
		return evaluateModulo(left, right)
	case "<", "<=", ">", ">=":
		return evaluateComparison(op, left, right)
	case "==":
		return runtime.BoolValue{Val: valuesEqual(left, right)}, nil
	case "!=":
		return runtime.BoolValue{Val: !valuesEqual(left, right)}, nil
	case "&", "|", "^", "<<", ">>":
		return evaluateBitwise(op, left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", op)
	}
}

func isNumericValue(val runtime.Value) bool {
	switch val.(type) {
	case runtime.IntegerValue, runtime.FloatValue:
		return true
	default:
		return false
	}
}

func numericToFloat(val runtime.Value) (float64, error) {
	switch v := val.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case runtime.IntegerValue:
		int32Val, err := int32FromIntegerValue(v)
		if err != nil {
			return 0, err
		}
		return float64(int32Val), nil
	default:
		return 0, fmt.Errorf("Arithmetic requires numeric operands")
	}
}

func evaluateModulo(left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if lv, ok := left.(runtime.IntegerValue); ok {
		if rv, ok := right.(runtime.IntegerValue); ok {
			if rv.Val == nil || rv.Val.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result := new(big.Int).Rem(runtime.CloneBigInt(lv.Val), rv.Val)
			return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	if rightFloat == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return runtime.FloatValue{Val: math.Mod(leftFloat, rightFloat), TypeSuffix: runtime.FloatF64}, nil
}

func evaluateBitwise(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires i32 operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires i32 operands")
	}
	l, err := int32FromIntegerValue(lv)
	if err != nil {
		return nil, err
	}
	r, err := int32FromIntegerValue(rv)
	if err != nil {
		return nil, err
	}
	var result int32
	switch op {
	case "&":
		result = l & r
	case "|":
		result = l | r
	case "^":
		result = l ^ r
	case "<<":
		if r < 0 || r >= 32 {
			return nil, fmt.Errorf("shift out of range")
		}
		result = l << uint(r)
	case ">>":
		if r < 0 || r >= 32 {
			return nil, fmt.Errorf("shift out of range")
		}
		result = l >> uint(r)
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
	return runtime.IntegerValue{Val: big.NewInt(int64(result)), TypeSuffix: runtime.IntegerI32}, nil
}

func int32FromIntegerValue(val runtime.IntegerValue) (int32, error) {
	if val.TypeSuffix != runtime.IntegerI32 {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	if val.Val == nil || !val.Val.IsInt64() {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	raw := val.Val.Int64()
	if raw < math.MinInt32 || raw > math.MaxInt32 {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	return int32(raw), nil
}

func (i *Interpreter) evaluateFunctionCall(call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, error) {
	calleeVal, err := i.evaluateExpression(call.Callee, env)
	if err != nil {
		return nil, err
	}
	var injected []runtime.Value
	var funcValue *runtime.FunctionValue
	switch fn := calleeVal.(type) {
	case runtime.NativeFunctionValue:
		args := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env}
		return fn.Impl(ctx, args)
	case *runtime.NativeFunctionValue:
		args := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env}
		return fn.Impl(ctx, args)
	case runtime.NativeBoundMethodValue:
		args := make([]runtime.Value, 0, len(call.Arguments)+1)
		args = append(args, fn.Receiver)
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env}
		return fn.Method.Impl(ctx, args)
	case *runtime.NativeBoundMethodValue:
		args := make([]runtime.Value, 0, len(call.Arguments)+1)
		args = append(args, fn.Receiver)
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env}
		return fn.Method.Impl(ctx, args)
	case runtime.DynRefValue:
		resolved, resErr := i.resolveDynRef(fn)
		if resErr != nil {
			return nil, resErr
		}
		funcValue = resolved
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, resErr := i.resolveDynRef(*fn)
		if resErr != nil {
			return nil, resErr
		}
		funcValue = resolved
	case *runtime.BoundMethodValue:
		funcValue = fn.Method
		injected = append(injected, fn.Receiver)
	case runtime.BoundMethodValue:
		funcValue = fn.Method
		injected = append(injected, fn.Receiver)
	case *runtime.FunctionValue:
		funcValue = fn
	default:
		return nil, fmt.Errorf("calling non-function value of kind %s", calleeVal.Kind())
	}
	if funcValue == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	args := make([]runtime.Value, 0, len(injected)+len(call.Arguments))
	args = append(args, injected...)
	for _, argExpr := range call.Arguments {
		val, err := i.evaluateExpression(argExpr, env)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	return i.invokeFunction(funcValue, args, call)
}

func (i *Interpreter) invokeFunction(fn *runtime.FunctionValue, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, error) {
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.Body == nil {
			return runtime.NilValue{}, nil
		}
		if call != nil {
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		if len(args) != len(decl.Params) {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, len(decl.Params), len(args))
		}
		localEnv := runtime.NewEnvironment(fn.Closure)
		if call != nil {
			i.bindTypeArgumentsIfAny(decl, call, localEnv)
		}
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("function parameter %d is nil", idx)
			}
			if err := i.assignPattern(param.Name, args[idx], localEnv, true); err != nil {
				return nil, err
			}
		}
		result, err := i.evaluateBlock(decl.Body, localEnv)
		if err != nil {
			if ret, ok := err.(returnSignal); ok {
				if ret.value == nil {
					return runtime.NilValue{}, nil
				}
				return ret.value, nil
			}
			return nil, err
		}
		if result == nil {
			return runtime.NilValue{}, nil
		}
		return result, nil
	default:
		return nil, fmt.Errorf("calling unsupported function declaration %T", fn.Declaration)
	}
}

func (i *Interpreter) enforceGenericConstraintsIfAny(funcNode ast.Node, call *ast.FunctionCall) error {
	if funcNode == nil || call == nil {
		return nil
	}
	generics, whereClause := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return nil
	}
	name := functionNameForErrors(funcNode)
	if len(call.TypeArguments) != len(generics) {
		return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", name, len(generics), len(call.TypeArguments))
	}
	constraints := collectConstraintSpecs(generics, whereClause)
	if len(constraints) == 0 {
		return nil
	}
	typeArgMap, err := mapTypeArguments(generics, call.TypeArguments, fmt.Sprintf("calling %s", name))
	if err != nil {
		return err
	}
	return i.enforceConstraintSpecs(constraints, typeArgMap)
}

func (i *Interpreter) bindTypeArgumentsIfAny(funcNode ast.Node, call *ast.FunctionCall, env *runtime.Environment) {
	if funcNode == nil || call == nil {
		return
	}
	generics, _ := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return
	}
	count := len(generics)
	if len(call.TypeArguments) < count {
		count = len(call.TypeArguments)
	}
	for idx := 0; idx < count; idx++ {
		gp := generics[idx]
		if gp == nil || gp.Name == nil {
			continue
		}
		ta := call.TypeArguments[idx]
		if ta == nil {
			continue
		}
		name := gp.Name.Name + "_type"
		value := runtime.StringValue{Val: typeExpressionToString(ta)}
		env.Define(name, value)
	}
}

func extractFunctionGenerics(funcNode ast.Node) ([]*ast.GenericParameter, []*ast.WhereClauseConstraint) {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		return fn.GenericParams, fn.WhereClause
	case *ast.LambdaExpression:
		return fn.GenericParams, fn.WhereClause
	default:
		return nil, nil
	}
}

func functionNameForErrors(funcNode ast.Node) string {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		if fn.ID != nil && fn.ID.Name != "" {
			return fn.ID.Name
		}
	case *ast.LambdaExpression:
		return "(lambda)"
	}
	return "(lambda)"
}

func collectConstraintSpecs(generics []*ast.GenericParameter, whereClause []*ast.WhereClauseConstraint) []constraintSpec {
	var specs []constraintSpec
	for _, gp := range generics {
		if gp == nil || gp.Name == nil {
			continue
		}
		for _, constraint := range gp.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			specs = append(specs, constraintSpec{typeParam: gp.Name.Name, ifaceType: constraint.InterfaceType})
		}
	}
	for _, clause := range whereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		for _, constraint := range clause.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			specs = append(specs, constraintSpec{typeParam: clause.TypeParam.Name, ifaceType: constraint.InterfaceType})
		}
	}
	return specs
}

func constraintSignature(specs []constraintSpec) string {
	if len(specs) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(specs))
	for _, spec := range specs {
		parts = append(parts, fmt.Sprintf("%s->%s", spec.typeParam, typeExpressionToString(spec.ifaceType)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func (i *Interpreter) registerUnnamedImpl(ifaceName string, variant targetVariant, unionSignatures []string, baseConstraintSig string, targetDescription string) error {
	key := ifaceName + "::" + variant.typeName
	bucket, ok := i.unnamedImpls[key]
	if !ok {
		bucket = make(map[string]map[string]struct{})
		i.unnamedImpls[key] = bucket
	}
	templateKey := "<none>"
	if len(variant.argTemplates) > 0 {
		parts := make([]string, 0, len(variant.argTemplates))
		for _, tmpl := range variant.argTemplates {
			parts = append(parts, typeExpressionToString(tmpl))
		}
		templateKey = strings.Join(parts, "|")
	}
	if len(unionSignatures) > 0 {
		prefix := strings.Join(unionSignatures, "::")
		templateKey = prefix + "::" + templateKey
	}
	constraintKey := baseConstraintSig
	if len(unionSignatures) > 0 {
		prefix := strings.Join(unionSignatures, "::")
		constraintKey = prefix + "::" + baseConstraintSig
	}
	constraintSet, ok := bucket[templateKey]
	if !ok {
		constraintSet = make(map[string]struct{})
		bucket[templateKey] = constraintSet
	}
	if _, exists := constraintSet[constraintKey]; exists {
		return fmt.Errorf("Unnamed impl for (%s, %s) already exists", ifaceName, targetDescription)
	}
	constraintSet[constraintKey] = struct{}{}
	return nil
}

func genericNameSet(params []*ast.GenericParameter) map[string]struct{} {
	set := make(map[string]struct{})
	for _, gp := range params {
		if gp == nil || gp.Name == nil {
			continue
		}
		set[gp.Name.Name] = struct{}{}
	}
	return set
}

func measureTemplateSpecificity(expr ast.TypeExpression, genericNames map[string]struct{}) int {
	if expr == nil {
		return 0
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0
		}
		if _, ok := genericNames[t.Name.Name]; ok {
			return 0
		}
		return 1
	case *ast.GenericTypeExpression:
		score := measureTemplateSpecificity(t.Base, genericNames)
		for _, arg := range t.Arguments {
			score += measureTemplateSpecificity(arg, genericNames)
		}
		return score
	case *ast.NullableTypeExpression:
		return measureTemplateSpecificity(t.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return measureTemplateSpecificity(t.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		score := 0
		for _, member := range t.Members {
			score += measureTemplateSpecificity(member, genericNames)
		}
		return score
	default:
		return 0
	}
}

func computeImplSpecificity(entry *implEntry, bindings map[string]ast.TypeExpression, constraints []constraintSpec) int {
	genericNames := genericNameSet(entry.genericParams)
	concreteScore := 0
	for _, tmpl := range entry.argTemplates {
		concreteScore += measureTemplateSpecificity(tmpl, genericNames)
	}
	constraintScore := len(constraints)
	bindingScore := len(bindings)
	unionPenalty := len(entry.unionVariants)
	defaultPenalty := 0
	if entry.defaultOnly {
		defaultPenalty = 1
	}
	return concreteScore*100 + constraintScore*10 + bindingScore - unionPenalty - defaultPenalty
}

func typeExpressionsEqual(a, b ast.TypeExpression) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch ta := a.(type) {
	case *ast.SimpleTypeExpression:
		tb, ok := b.(*ast.SimpleTypeExpression)
		if !ok {
			return false
		}
		if ta.Name == nil || tb.Name == nil {
			return ta.Name == nil && tb.Name == nil
		}
		return ta.Name.Name == tb.Name.Name
	case *ast.GenericTypeExpression:
		tb, ok := b.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !typeExpressionsEqual(ta.Base, tb.Base) {
			return false
		}
		if len(ta.Arguments) != len(tb.Arguments) {
			return false
		}
		for idx := range ta.Arguments {
			if !typeExpressionsEqual(ta.Arguments[idx], tb.Arguments[idx]) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		tb, ok := b.(*ast.NullableTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEqual(ta.InnerType, tb.InnerType)
	case *ast.ResultTypeExpression:
		tb, ok := b.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEqual(ta.InnerType, tb.InnerType)
	case *ast.FunctionTypeExpression:
		tb, ok := b.(*ast.FunctionTypeExpression)
		if !ok {
			return false
		}
		if len(ta.ParamTypes) != len(tb.ParamTypes) {
			return false
		}
		for idx := range ta.ParamTypes {
			if !typeExpressionsEqual(ta.ParamTypes[idx], tb.ParamTypes[idx]) {
				return false
			}
		}
		return typeExpressionsEqual(ta.ReturnType, tb.ReturnType)
	case *ast.UnionTypeExpression:
		tb, ok := b.(*ast.UnionTypeExpression)
		if !ok || len(ta.Members) != len(tb.Members) {
			return false
		}
		for idx := range ta.Members {
			if !typeExpressionsEqual(ta.Members[idx], tb.Members[idx]) {
				return false
			}
		}
		return true
	case *ast.WildcardTypeExpression:
		_, ok := b.(*ast.WildcardTypeExpression)
		return ok
	default:
		return false
	}
}

func matchTypeExpressionTemplate(template, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	switch t := template.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return actual == nil
		}
		name := t.Name.Name
		if _, isGeneric := genericNames[name]; isGeneric {
			if existing, ok := bindings[name]; ok {
				return typeExpressionsEqual(existing, actual)
			}
			bindings[name] = actual
			return true
		}
		return typeExpressionsEqual(template, actual)
	case *ast.GenericTypeExpression:
		other, ok := actual.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !matchTypeExpressionTemplate(t.Base, other.Base, genericNames, bindings) {
			return false
		}
		if len(t.Arguments) != len(other.Arguments) {
			return false
		}
		for idx := range t.Arguments {
			if !matchTypeExpressionTemplate(t.Arguments[idx], other.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		other, ok := actual.(*ast.NullableTypeExpression)
		if !ok {
			return false
		}
		return matchTypeExpressionTemplate(t.InnerType, other.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		other, ok := actual.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return matchTypeExpressionTemplate(t.InnerType, other.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		other, ok := actual.(*ast.UnionTypeExpression)
		if !ok || len(t.Members) != len(other.Members) {
			return false
		}
		for idx := range t.Members {
			if !matchTypeExpressionTemplate(t.Members[idx], other.Members[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	default:
		return typeExpressionsEqual(template, actual)
	}
}

func (i *Interpreter) matchImplEntry(entry *implEntry, info typeInfo) (map[string]ast.TypeExpression, bool) {
	if entry == nil {
		return nil, false
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := genericNameSet(entry.genericParams)
	if len(entry.argTemplates) > 0 {
		if len(info.typeArgs) != len(entry.argTemplates) {
			return nil, false
		}
		for idx, tmpl := range entry.argTemplates {
			if !matchTypeExpressionTemplate(tmpl, info.typeArgs[idx], genericNames, bindings) {
				return nil, false
			}
		}
	} else if len(info.typeArgs) > 0 {
		return nil, false
	}
	for _, gp := range entry.genericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		if _, ok := bindings[gp.Name.Name]; !ok {
			return nil, false
		}
	}
	return bindings, true
}

func (i *Interpreter) collectImplCandidates(info typeInfo, interfaceFilter string) ([]implCandidate, error) {
	if info.name == "" {
		return nil, nil
	}
	entries := i.implMethods[info.name]
	matches := make([]implCandidate, 0)
	var constraintErr error
	for idx := range entries {
		entry := &entries[idx]
		if interfaceFilter != "" && entry.interfaceName != interfaceFilter {
			continue
		}
		bindings, ok := i.matchImplEntry(entry, info)
		if !ok {
			continue
		}
		constraints := collectConstraintSpecs(entry.genericParams, entry.whereClause)
		if len(constraints) > 0 {
			if err := i.enforceConstraintSpecs(constraints, bindings); err != nil {
				if constraintErr == nil {
					constraintErr = err
				}
				continue
			}
		}
		score := computeImplSpecificity(entry, bindings, constraints)
		matches = append(matches, implCandidate{
			entry:       entry,
			bindings:    bindings,
			constraints: constraints,
			score:       score,
		})
	}
	if len(matches) == 0 {
		return nil, constraintErr
	}
	return matches, nil
}

func (i *Interpreter) compareMethodMatches(a, b implCandidate) int {
	if a.score > b.score {
		return 1
	}
	if a.score < b.score {
		return -1
	}
	aUnion := a.entry.unionVariants
	bUnion := b.entry.unionVariants
	if len(aUnion) > 0 && len(bUnion) == 0 {
		return -1
	}
	if len(aUnion) == 0 && len(bUnion) > 0 {
		return 1
	}
	if len(aUnion) > 0 && len(bUnion) > 0 {
		if isProperSubset(aUnion, bUnion) {
			return 1
		}
		if isProperSubset(bUnion, aUnion) {
			return -1
		}
		if len(aUnion) != len(bUnion) {
			if len(aUnion) < len(bUnion) {
				return 1
			}
			return -1
		}
	}
	aConstraints := i.buildConstraintKeySet(a.constraints)
	bConstraints := i.buildConstraintKeySet(b.constraints)
	if isConstraintSuperset(aConstraints, bConstraints) {
		return 1
	}
	if isConstraintSuperset(bConstraints, aConstraints) {
		return -1
	}
	return 0
}

func (i *Interpreter) buildConstraintKeySet(constraints []constraintSpec) map[string]struct{} {
	set := make(map[string]struct{})
	for _, c := range constraints {
		if c.ifaceType == nil {
			continue
		}
		expressions := i.collectInterfaceConstraintStrings(c.ifaceType, make(map[string]struct{}))
		for _, expr := range expressions {
			key := fmt.Sprintf("%s->%s", c.typeParam, expr)
			set[key] = struct{}{}
		}
	}
	return set
}

func (i *Interpreter) collectInterfaceConstraintStrings(typeExpr ast.TypeExpression, memo map[string]struct{}) []string {
	if typeExpr == nil {
		return nil
	}
	key := typeExpressionToString(typeExpr)
	if _, seen := memo[key]; seen {
		return nil
	}
	memo[key] = struct{}{}
	results := []string{key}
	if simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
		if iface, exists := i.interfaces[simple.Name.Name]; exists && iface.Node != nil {
			for _, base := range iface.Node.BaseInterfaces {
				results = append(results, i.collectInterfaceConstraintStrings(base, memo)...)
			}
		}
	}
	return results
}

func isConstraintSuperset(a, b map[string]struct{}) bool {
	if len(a) <= len(b) {
		return false
	}
	for key := range b {
		if _, ok := a[key]; !ok {
			return false
		}
	}
	return true
}

func isProperSubset(a, b []string) bool {
	if len(a) == 0 {
		return len(b) > 0
	}
	setA := make(map[string]struct{}, len(a))
	for _, val := range a {
		setA[val] = struct{}{}
	}
	setB := make(map[string]struct{}, len(b))
	for _, val := range b {
		setB[val] = struct{}{}
	}
	if len(setA) >= len(setB) {
		return false
	}
	for val := range setA {
		if _, ok := setB[val]; !ok {
			return false
		}
	}
	return true
}

func (i *Interpreter) selectBestMethodCandidate(matches []methodMatch) (*methodMatch, []methodMatch) {
	if len(matches) == 0 {
		return nil, nil
	}
	bestIdx := 0
	contenders := []int{0}
	for idx := 1; idx < len(matches); idx++ {
		cmp := i.compareMethodMatches(matches[idx].candidate, matches[bestIdx].candidate)
		if cmp > 0 {
			bestIdx = idx
			contenders = []int{idx}
		} else if cmp == 0 {
			reverse := i.compareMethodMatches(matches[bestIdx].candidate, matches[idx].candidate)
			if reverse < 0 {
				bestIdx = idx
				contenders = []int{idx}
			} else if reverse == 0 {
				contenders = append(contenders, idx)
			}
		}
	}
	if len(contenders) > 1 {
		ambiguous := make([]methodMatch, 0, len(contenders))
		for _, idx := range contenders {
			ambiguous = append(ambiguous, matches[idx])
		}
		return nil, ambiguous
	}
	return &matches[bestIdx], nil
}

func (i *Interpreter) selectBestCandidate(matches []implCandidate) (*implCandidate, []implCandidate) {
	if len(matches) == 0 {
		return nil, nil
	}
	bestIdx := 0
	contenders := []int{0}
	for idx := 1; idx < len(matches); idx++ {
		cmp := i.compareMethodMatches(matches[idx], matches[bestIdx])
		if cmp > 0 {
			bestIdx = idx
			contenders = []int{idx}
		} else if cmp == 0 {
			reverse := i.compareMethodMatches(matches[bestIdx], matches[idx])
			if reverse < 0 {
				bestIdx = idx
				contenders = []int{idx}
			} else if reverse == 0 {
				contenders = append(contenders, idx)
			}
		}
	}
	if len(contenders) > 1 {
		ambiguous := make([]implCandidate, 0, len(contenders))
		for _, idx := range contenders {
			ambiguous = append(ambiguous, matches[idx])
		}
		return nil, ambiguous
	}
	return &matches[bestIdx], nil
}

func (i *Interpreter) typeInfoFromStructInstance(inst *runtime.StructInstanceValue) (typeInfo, bool) {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return typeInfo{}, false
	}
	info := typeInfo{name: inst.Definition.Node.ID.Name}
	if len(inst.TypeArguments) > 0 {
		info.typeArgs = append([]ast.TypeExpression(nil), inst.TypeArguments...)
	}
	return info, true
}

func typeInfoToString(info typeInfo) string {
	if info.name == "" {
		return "<unknown>"
	}
	if len(info.typeArgs) == 0 {
		return info.name
	}
	parts := make([]string, 0, len(info.typeArgs))
	for _, arg := range info.typeArgs {
		parts = append(parts, typeExpressionToString(arg))
	}
	return fmt.Sprintf("%s<%s>", info.name, strings.Join(parts, ", "))
}

func descriptionsFromMethodMatches(matches []methodMatch) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.candidate.entry == nil || match.candidate.entry.definition == nil {
			continue
		}
		desc := typeExpressionToString(match.candidate.entry.definition.TargetType)
		set[desc] = struct{}{}
	}
	return sortedKeys(set)
}

func descriptionsFromCandidates(matches []implCandidate) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.entry == nil || match.entry.definition == nil {
			continue
		}
		desc := typeExpressionToString(match.entry.definition.TargetType)
		set[desc] = struct{}{}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapTypeArguments(generics []*ast.GenericParameter, provided []ast.TypeExpression, context string) (map[string]ast.TypeExpression, error) {
	result := make(map[string]ast.TypeExpression)
	if len(generics) == 0 {
		return result, nil
	}
	if len(provided) != len(generics) {
		return nil, fmt.Errorf("Type arguments count mismatch %s: expected %d, got %d", context, len(generics), len(provided))
	}
	for idx, gp := range generics {
		if gp == nil || gp.Name == nil {
			continue
		}
		ta := provided[idx]
		if ta == nil {
			return nil, fmt.Errorf("Missing type argument for '%s' required by %s", gp.Name.Name, context)
		}
		result[gp.Name.Name] = ta
	}
	return result, nil
}

func (i *Interpreter) enforceConstraintSpecs(constraints []constraintSpec, typeArgMap map[string]ast.TypeExpression) error {
	for _, spec := range constraints {
		actual, ok := typeArgMap[spec.typeParam]
		if !ok {
			return fmt.Errorf("Missing type argument for '%s' required by constraints", spec.typeParam)
		}
		tInfo, ok := parseTypeExpression(actual)
		if !ok {
			continue
		}
		if err := i.ensureTypeSatisfiesInterface(tInfo, spec.ifaceType, spec.typeParam, make(map[string]struct{})); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) ensureTypeSatisfiesInterface(tInfo typeInfo, ifaceExpr ast.TypeExpression, context string, visited map[string]struct{}) error {
	ifaceInfo, ok := parseTypeExpression(ifaceExpr)
	if !ok {
		return nil
	}
	if _, seen := visited[ifaceInfo.name]; seen {
		return nil
	}
	visited[ifaceInfo.name] = struct{}{}
	ifaceDef, ok := i.interfaces[ifaceInfo.name]
	if !ok {
		return fmt.Errorf("Unknown interface '%s' in constraint on '%s'", ifaceInfo.name, context)
	}
	if ifaceDef.Node != nil {
		for _, base := range ifaceDef.Node.BaseInterfaces {
			if err := i.ensureTypeSatisfiesInterface(tInfo, base, context, visited); err != nil {
				return err
			}
		}
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			methodName := sig.Name.Name
			if !i.typeHasMethod(tInfo, methodName, ifaceInfo.name) {
				return fmt.Errorf("Type '%s' does not satisfy interface '%s': missing method '%s'", tInfo.name, ifaceInfo.name, methodName)
			}
		}
	}
	return nil
}

func (i *Interpreter) typeHasMethod(info typeInfo, methodName, ifaceName string) bool {
	if info.name == "" {
		return false
	}
	if bucket, ok := i.inherentMethods[info.name]; ok {
		if _, exists := bucket[methodName]; exists {
			return true
		}
	}
	method, err := i.findMethod(info, methodName, ifaceName)
	return err == nil && method != nil
}

func parseTypeExpression(expr ast.TypeExpression) (typeInfo, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: t.Name.Name, typeArgs: nil}, true
	case *ast.GenericTypeExpression:
		tInfo, ok := parseTypeExpression(t.Base)
		if !ok {
			return typeInfo{}, false
		}
		tInfo.typeArgs = append([]ast.TypeExpression(nil), t.Arguments...)
		return tInfo, true
	default:
		return typeInfo{}, false
	}
}

func typeExpressionToString(expr ast.TypeExpression) string {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "<?>"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		base := typeExpressionToString(t.Base)
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			args = append(args, typeExpressionToString(arg))
		}
		return fmt.Sprintf("%s<%s>", base, strings.Join(args, ", "))
	case *ast.NullableTypeExpression:
		return typeExpressionToString(t.InnerType) + "?"
	case *ast.FunctionTypeExpression:
		parts := make([]string, 0, len(t.ParamTypes))
		for _, p := range t.ParamTypes {
			parts = append(parts, typeExpressionToString(p))
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(parts, ", "), typeExpressionToString(t.ReturnType))
	case *ast.UnionTypeExpression:
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, typeExpressionToString(member))
		}
		return strings.Join(parts, " | ")
	default:
		return "<?>"
	}
}

// bigFromLiteral normalizes numeric literals (number or bigint) to *big.Int.
func bigFromLiteral(val interface{}) *big.Int {
	switch v := val.(type) {
	case int:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case float64:
		return big.NewInt(int64(v))
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
		return big.NewInt(0)
	case *big.Int:
		return runtime.CloneBigInt(v)
	default:
		return big.NewInt(0)
	}
}

func evaluateArithmetic(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if lv, ok := left.(runtime.IntegerValue); ok {
		if rv, ok := right.(runtime.IntegerValue); ok {
			result := new(big.Int)
			switch op {
			case "+":
				result.Add(lv.Val, rv.Val)
			case "-":
				result.Sub(lv.Val, rv.Val)
			case "*":
				result.Mul(lv.Val, rv.Val)
			case "/":
				if rv.Val.Sign() == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result.Quo(lv.Val, rv.Val)
			default:
				return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
			}
			return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	var val float64
	switch op {
	case "+":
		val = leftFloat + rightFloat
	case "-":
		val = leftFloat - rightFloat
	case "*":
		val = leftFloat * rightFloat
	case "/":
		if rightFloat == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		val = leftFloat / rightFloat
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
	return runtime.FloatValue{Val: val, TypeSuffix: runtime.FloatF64}, nil
}

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	if isNumericValue(left) && isNumericValue(right) {
		lf, err := numericToFloat(left)
		if err != nil {
			return false
		}
		rf, err := numericToFloat(right)
		if err != nil {
			return false
		}
		return math.Abs(lf-rf) < numericEpsilon
	}
	switch lv := left.(type) {
	case runtime.StringValue:
		if rv, ok := right.(runtime.StringValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.BoolValue:
		if rv, ok := right.(runtime.BoolValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.CharValue:
		if rv, ok := right.(runtime.CharValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.NilValue:
		_, ok := right.(runtime.NilValue)
		return ok
	case runtime.IntegerValue:
		if rv, ok := right.(runtime.IntegerValue); ok {
			return lv.Val.Cmp(rv.Val) == 0
		}
	case runtime.FloatValue:
		if rv, ok := right.(runtime.FloatValue); ok {
			return math.Abs(lv.Val-rv.Val) < numericEpsilon
		}
	}
	return false
}

func evaluateComparison(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	cmp := 0
	diff := leftFloat - rightFloat
	if math.Abs(diff) < numericEpsilon {
		cmp = 0
	} else if diff < 0 {
		cmp = -1
	} else {
		cmp = 1
	}
	return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
}

func comparisonOp(op string, cmp int) bool {
	switch op {
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	default:
		return false
	}
}

func isTruthy(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.BoolValue:
		return v.Val
	case runtime.NilValue:
		return false
	default:
		return true
	}
}

type breakSignal struct {
	label string
	value runtime.Value
}

func (b breakSignal) Error() string {
	if b.label != "" {
		return fmt.Sprintf("break %s", b.label)
	}
	return "break"
}

type continueSignal struct {
	label string
}

func (c continueSignal) Error() string {
	if c.label != "" {
		return fmt.Sprintf("continue %s", c.label)
	}
	return "continue"
}

type raiseSignal struct {
	value runtime.Value
}

func (r raiseSignal) Error() string {
	if errVal, ok := r.value.(runtime.ErrorValue); ok {
		return errVal.Message
	}
	return valueToString(r.value)
}

type returnSignal struct {
	value runtime.Value
}

func (r returnSignal) Error() string {
	return "return"
}

func makeErrorValue(val runtime.Value) runtime.ErrorValue {
	if errVal, ok := val.(runtime.ErrorValue); ok {
		return errVal
	}
	message := valueToString(val)
	payload := map[string]runtime.Value{
		"value": val,
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

func (i *Interpreter) stringifyValue(val runtime.Value, env *runtime.Environment) (string, error) {
	_ = env
	if inst, ok := val.(*runtime.StructInstanceValue); ok {
		if str, ok := i.invokeStructToString(inst); ok {
			return str, nil
		}
	}
	return valueToString(val), nil
}

func (i *Interpreter) invokeStructToString(inst *runtime.StructInstanceValue) (string, bool) {
	if inst == nil {
		return "", false
	}
	typeName := structTypeName(inst)
	if typeName == "" {
		return "", false
	}
	if bucket, ok := i.inherentMethods[typeName]; ok {
		if method := bucket["to_string"]; method != nil {
			if str, ok := i.callStringMethod(method, inst); ok {
				return str, true
			}
		}
	}
	if method, err := i.selectStructMethod(inst, "to_string"); err == nil && method != nil {
		if str, ok := i.callStringMethod(method, inst); ok {
			return str, true
		}
	}
	return "", false
}

func (i *Interpreter) callStringMethod(fn *runtime.FunctionValue, receiver runtime.Value) (string, bool) {
	if fn == nil {
		return "", false
	}
	result, err := i.invokeFunction(fn, []runtime.Value{receiver}, nil)
	if err != nil {
		return "", false
	}
	if result == nil {
		return "", false
	}
	if strVal, ok := result.(runtime.StringValue); ok {
		return strVal.Val, true
	}
	return "", false
}

func structTypeName(inst *runtime.StructInstanceValue) string {
	if inst == nil {
		return ""
	}
	if inst.Definition != nil && inst.Definition.Node != nil && inst.Definition.Node.ID != nil {
		return inst.Definition.Node.ID.Name
	}
	return ""
}

func simpleTypeName(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			return t.Name.Name, true
		}
	}
	return "", false
}

func valueToString(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.CharValue:
		return string(v.Val)
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.NilValue:
		return "nil"
	case *runtime.ArrayValue:
		parts := make([]string, 0, len(v.Elements))
		for _, el := range v.Elements {
			parts = append(parts, valueToString(el))
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	case *runtime.RangeValue:
		start := valueToString(v.Start)
		end := valueToString(v.End)
		delim := "..."
		if v.Inclusive {
			delim = ".."
		}
		return fmt.Sprintf("%s%s%s", start, delim, end)
	case *runtime.StructInstanceValue:
		return structInstanceToString(v)
	case runtime.StructDefinitionValue:
		name := "struct"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case *runtime.StructDefinitionValue:
		name := "struct"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case *runtime.InterfaceDefinitionValue:
		name := "interface"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case runtime.InterfaceDefinitionValue:
		name := "interface"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.InterfaceValue:
		name := "interface"
		if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil {
			name = v.Interface.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.FunctionValue:
		return "<function>"
	case runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case *runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case runtime.BoundMethodValue:
		return "<bound method>"
	case *runtime.BoundMethodValue:
		return "<bound method>"
	case runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case *runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case runtime.PackageValue:
		return fmt.Sprintf("<package %s>", strings.Join(v.NamePath, "::"))
	case runtime.ErrorValue:
		return v.Message
	default:
		if val == nil {
			return "<nil>"
		}
		return fmt.Sprintf("[%s]", val.Kind())
	}
}

func structInstanceToString(inst *runtime.StructInstanceValue) string {
	if inst == nil {
		return "<struct> {}"
	}
	name := structTypeName(inst)
	if name == "" {
		name = "<struct>"
	}
	if inst.Positional != nil {
		parts := make([]string, 0, len(inst.Positional))
		for _, el := range inst.Positional {
			parts = append(parts, valueToString(el))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	if inst.Fields != nil {
		keys := make([]string, 0, len(inst.Fields))
		for k := range inst.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, valueToString(inst.Fields[k])))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	return fmt.Sprintf("%s { }", name)
}

func (i *Interpreter) evaluateFunctionDefinition(def *ast.FunctionDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Function definition requires identifier")
	}
	if err := i.validateGenericConstraints(def); err != nil {
		return nil, err
	}
	fnVal := &runtime.FunctionValue{Declaration: def, Closure: env}
	env.Define(def.ID.Name, fnVal)
	i.registerSymbol(def.ID.Name, fnVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, fnVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateStructDefinition(def *ast.StructDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Struct definition requires identifier")
	}
	structVal := &runtime.StructDefinitionValue{Node: def}
	env.Define(def.ID.Name, structVal)
	i.registerSymbol(def.ID.Name, structVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, structVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateInterfaceDefinition(def *ast.InterfaceDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Interface definition requires identifier")
	}
	ifaceVal := &runtime.InterfaceDefinitionValue{Node: def, Env: env}
	env.Define(def.ID.Name, ifaceVal)
	i.interfaces[def.ID.Name] = ifaceVal
	i.registerSymbol(def.ID.Name, ifaceVal)
	if qn := i.qualifiedName(def.ID.Name); qn != "" {
		i.global.Define(qn, ifaceVal)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateImplementationDefinition(def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.InterfaceName == nil {
		return nil, fmt.Errorf("Implementation requires interface name")
	}
	ifaceName := def.InterfaceName.Name
	ifaceDef, ok := i.interfaces[ifaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", ifaceName)
	}
	variants, unionSignatures, err := expandImplementationTargetVariants(def.TargetType)
	if err != nil {
		return nil, err
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("Implementation target must reference at least one concrete type")
	}
	methods := make(map[string]*runtime.FunctionValue)
	hasExplicit := false
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Implementation method requires identifier")
		}
		methods[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
		hasExplicit = true
	}
	if ifaceDef.Node != nil {
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			name := sig.Name.Name
			if _, ok := methods[name]; ok {
				continue
			}
			if sig.DefaultImpl == nil {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			methods[name] = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
		}
	}
	constraintSpecs := collectConstraintSpecs(def.GenericParams, def.WhereClause)
	baseConstraintSig := constraintSignature(constraintSpecs)
	targetDescription := typeExpressionToString(def.TargetType)
	for _, variant := range variants {
		if def.ImplName == nil {
			if err := i.registerUnnamedImpl(ifaceName, variant, unionSignatures, baseConstraintSig, targetDescription); err != nil {
				return nil, err
			}
			entry := implEntry{
				interfaceName: ifaceName,
				methods:       methods,
				definition:    def,
				argTemplates:  variant.argTemplates,
				genericParams: def.GenericParams,
				whereClause:   def.WhereClause,
				defaultOnly:   !hasExplicit,
			}
			if len(unionSignatures) > 0 {
				entry.unionVariants = append([]string(nil), unionSignatures...)
			}
			i.implMethods[variant.typeName] = append(i.implMethods[variant.typeName], entry)
		}
	}
	if def.ImplName != nil {
		name := def.ImplName.Name
		implVal := runtime.ImplementationNamespaceValue{
			Name:          def.ImplName,
			InterfaceName: def.InterfaceName,
			TargetType:    def.TargetType,
			Methods:       methods,
		}
		env.Define(name, implVal)
		i.registerSymbol(name, implVal)
		if qn := i.qualifiedName(name); qn != "" {
			i.global.Define(qn, implVal)
		}
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) validateGenericConstraints(def *ast.FunctionDefinition) error {
	if def == nil || len(def.GenericParams) == 0 {
		return nil
	}
	for _, param := range def.GenericParams {
		if param == nil || param.Name == nil {
			continue
		}
		for _, constraint := range param.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			ifaceName, ok := simpleTypeName(constraint.InterfaceType)
			if !ok || ifaceName == "" {
				return fmt.Errorf("Unknown interface in constraint on '%s'", param.Name.Name)
			}
			if _, exists := i.interfaces[ifaceName]; !exists {
				return fmt.Errorf("Unknown interface '%s' in constraint on '%s'", ifaceName, param.Name.Name)
			}
		}
	}
	return nil
}

func (i *Interpreter) evaluateMethodsDefinition(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	simpleType, ok := def.TargetType.(*ast.SimpleTypeExpression)
	if !ok || simpleType.Name == nil {
		return nil, fmt.Errorf("MethodsDefinition requires simple target type")
	}
	typeName := simpleType.Name.Name
	bucket, ok := i.inherentMethods[typeName]
	if !ok {
		bucket = make(map[string]*runtime.FunctionValue)
		i.inherentMethods[typeName] = bucket
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Method definition requires identifier")
		}
		bucket[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateStructLiteral(lit *ast.StructLiteral, env *runtime.Environment) (runtime.Value, error) {
	if lit.StructType == nil {
		return nil, fmt.Errorf("Struct literal requires explicit struct type in this milestone")
	}
	structName := lit.StructType.Name
	defValue, err := env.Get(structName)
	if err != nil {
		return nil, err
	}
	structDefVal, err := toStructDefinitionValue(defValue, structName)
	if err != nil {
		return nil, err
	}
	structDef := structDefVal.Node
	if structDef == nil {
		return nil, fmt.Errorf("struct definition '%s' unavailable", structName)
	}
	explicitTypeArgs := append([]ast.TypeExpression(nil), lit.TypeArguments...)
	if lit.IsPositional {
		if structDef.Kind != ast.StructKindPositional && structDef.Kind != ast.StructKindSingleton {
			return nil, fmt.Errorf("Positional struct literal not allowed for struct '%s'", structName)
		}
		if len(lit.Fields) != len(structDef.Fields) {
			return nil, fmt.Errorf("Struct '%s' expects %d fields, got %d", structName, len(structDef.Fields), len(lit.Fields))
		}
		values := make([]runtime.Value, len(lit.Fields))
		for idx, field := range lit.Fields {
			val, err := i.evaluateExpression(field.Value, env)
			if err != nil {
				return nil, err
			}
			values[idx] = val
		}
		typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, nil)
		if err != nil {
			return nil, err
		}
		return &runtime.StructInstanceValue{Definition: structDefVal, Positional: values, TypeArguments: typeArgs}, nil
	}
	if structDef.Kind == ast.StructKindPositional && lit.FunctionalUpdateSource == nil {
		return nil, fmt.Errorf("Named struct literal not allowed for positional struct '%s'", structName)
	}
	if lit.FunctionalUpdateSource != nil && structDef.Kind == ast.StructKindPositional {
		return nil, fmt.Errorf("Functional update only supported for named structs")
	}
	fields := make(map[string]runtime.Value)
	var baseStruct *runtime.StructInstanceValue
	if lit.FunctionalUpdateSource != nil {
		base, err := i.evaluateExpression(lit.FunctionalUpdateSource, env)
		if err != nil {
			return nil, err
		}
		var ok bool
		baseStruct, ok = base.(*runtime.StructInstanceValue)
		if !ok {
			return nil, fmt.Errorf("Functional update source must be a struct instance")
		}
		if baseStruct.Definition == nil || baseStruct.Definition.Node == nil || baseStruct.Definition.Node.ID == nil || baseStruct.Definition.Node.ID.Name != structName {
			return nil, fmt.Errorf("Functional update source must be same struct type")
		}
		if baseStruct.Fields == nil {
			return nil, fmt.Errorf("Functional update only supported for named structs")
		}
		for k, v := range baseStruct.Fields {
			fields[k] = v
		}
	}
	for _, f := range lit.Fields {
		name := ""
		if f.Name != nil {
			name = f.Name.Name
		} else if f.IsShorthand {
			if ident, ok := f.Value.(*ast.Identifier); ok {
				name = ident.Name
			}
		}
		if name == "" {
			return nil, fmt.Errorf("Named struct field initializer must have a field name")
		}
		val, err := i.evaluateExpression(f.Value, env)
		if err != nil {
			return nil, err
		}
		fields[name] = val
	}
	if structDef.Kind == ast.StructKindNamed {
		required := make(map[string]struct{}, len(structDef.Fields))
		for _, defField := range structDef.Fields {
			if defField.Name != nil {
				required[defField.Name.Name] = struct{}{}
			}
		}
		for k := range fields {
			delete(required, k)
		}
		if len(required) > 0 {
			for missing := range required {
				return nil, fmt.Errorf("Missing field '%s' for struct '%s'", missing, structName)
			}
		}
	}
	typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, baseStruct)
	if err != nil {
		return nil, err
	}
	return &runtime.StructInstanceValue{Definition: structDefVal, Fields: fields, TypeArguments: typeArgs}, nil
}

func (i *Interpreter) resolveStructTypeArguments(def *ast.StructDefinition, explicit []ast.TypeExpression, base *runtime.StructInstanceValue) ([]ast.TypeExpression, error) {
	if def == nil {
		return nil, fmt.Errorf("Struct definition missing")
	}
	structName := "<anonymous>"
	if def.ID != nil && def.ID.Name != "" {
		structName = def.ID.Name
	}
	genericCount := len(def.GenericParams)
	if genericCount == 0 {
		if len(explicit) > 0 {
			return nil, fmt.Errorf("Type '%s' does not accept type arguments", structName)
		}
		if base != nil && len(base.TypeArguments) > 0 {
			return nil, fmt.Errorf("Type '%s' does not accept type arguments", structName)
		}
		return nil, nil
	}
	if len(explicit) > 0 {
		if len(explicit) != genericCount {
			return nil, fmt.Errorf("Type '%s' expects %d type arguments, got %d", structName, genericCount, len(explicit))
		}
		return append([]ast.TypeExpression(nil), explicit...), nil
	}
	if base != nil {
		if len(base.TypeArguments) != genericCount {
			return nil, fmt.Errorf("Type '%s' expects %d type arguments, got %d", structName, genericCount, len(base.TypeArguments))
		}
		return append([]ast.TypeExpression(nil), base.TypeArguments...), nil
	}
	return nil, fmt.Errorf("Type '%s' requires type arguments", structName)
}

func (i *Interpreter) evaluateMemberAccess(expr *ast.MemberAccessExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	switch v := obj.(type) {
	case *runtime.StructDefinitionValue:
		return i.structDefinitionMember(v, expr.Member)
	case runtime.StructDefinitionValue:
		return i.structDefinitionMember(&v, expr.Member)
	case runtime.PackageValue:
		return i.packageMemberAccess(v, expr.Member)
	case *runtime.PackageValue:
		return i.packageMemberAccess(*v, expr.Member)
	case runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(v, expr.Member)
	case *runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(*v, expr.Member)
	case runtime.DynPackageValue:
		return i.dynPackageMemberAccess(v, expr.Member)
	case *runtime.DynPackageValue:
		return i.dynPackageMemberAccess(*v, expr.Member)
	case *runtime.StructInstanceValue:
		return i.structInstanceMember(v, expr.Member, env)
	case *runtime.InterfaceValue:
		return i.interfaceMember(v, expr.Member)
	default:
		if ident, ok := expr.Member.(*ast.Identifier); ok {
			if bound, ok := i.tryUfcs(env, ident.Name, obj); ok {
				return bound, nil
			}
		}
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
}

func (i *Interpreter) evaluateIndexExpression(expr *ast.IndexExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	idxVal, err := i.evaluateExpression(expr.Index, env)
	if err != nil {
		return nil, err
	}
	arr, err := toArrayValue(obj)
	if err != nil {
		return nil, err
	}
	idx, err := indexFromValue(idxVal)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(arr.Elements) {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	val := arr.Elements[idx]
	if val == nil {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	return val, nil
}

func toArrayValue(val runtime.Value) (*runtime.ArrayValue, error) {
	switch v := val.(type) {
	case *runtime.ArrayValue:
		return v, nil
	default:
		return nil, fmt.Errorf("Indexing is only supported on arrays")
	}
}

func indexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || !v.Val.IsInt64() {
			return 0, fmt.Errorf("Array index must be within int range")
		}
		return int(v.Val.Int64()), nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("Array index must be a number")
		}
		idx := int(math.Trunc(v.Val))
		return idx, nil
	default:
		return 0, fmt.Errorf("Array index must be a number")
	}
}

func (i *Interpreter) structInstanceMember(inst *runtime.StructInstanceValue, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
	if ident, ok := member.(*ast.Identifier); ok {
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		if val, ok := inst.Fields[ident.Name]; ok {
			return val, nil
		}
		if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			if bound, ok := i.tryUfcs(env, ident.Name, inst); ok {
				return bound, nil
			}
			return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
		}
		typeName := inst.Definition.Node.ID.Name
		if bucket, ok := i.inherentMethods[typeName]; ok {
			if method, ok := bucket[ident.Name]; ok {
				if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
					return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
				}
				return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
			}
		}
		method, err := i.selectStructMethod(inst, ident.Name)
		if err != nil {
			return nil, err
		}
		if method != nil {
			return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
		}
		if bound, ok := i.tryUfcs(env, ident.Name, inst); ok {
			return bound, nil
		}
		return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
	}
	if intLit, ok := member.(*ast.IntegerLiteral); ok {
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if intLit.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(intLit.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		return inst.Positional[idx], nil
	}
	return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
}

func (i *Interpreter) tryUfcs(env *runtime.Environment, funcName string, receiver runtime.Value) (runtime.Value, bool) {
	if env == nil {
		return nil, false
	}
	val, err := env.Get(funcName)
	if err != nil {
		return nil, false
	}
	if fn, ok := val.(*runtime.FunctionValue); ok {
		return &runtime.BoundMethodValue{Receiver: receiver, Method: fn}, true
	}
	return nil, false
}

func (i *Interpreter) structDefinitionMember(def *runtime.StructDefinitionValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Static access expects identifier member")
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing identifier")
	}
	typeName := def.Node.ID.Name
	bucket := i.inherentMethods[typeName]
	if bucket == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	method, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
	}
	return method, nil
}

func (i *Interpreter) packageMemberAccess(pkg runtime.PackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Package member access expects identifier")
	}
	if pkg.Public == nil {
		return nil, fmt.Errorf("Package has no public members")
	}
	val, ok := pkg.Public[ident.Name]
	if !ok {
		pkgName := strings.Join(pkg.NamePath, ".")
		if pkgName == "" {
			pkgName = "<package>"
		}
		return nil, fmt.Errorf("No public member '%s' on package %s", ident.Name, pkgName)
	}
	return val, nil
}

func (i *Interpreter) dynPackageMemberAccess(pkg runtime.DynPackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Dyn package member access expects identifier")
	}
	pkgName := pkg.Name
	if pkgName == "" {
		pkgName = strings.Join(pkg.NamePath, ".")
	}
	if pkgName == "" {
		return nil, fmt.Errorf("Dyn package missing name")
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' not found", pkgName)
	}
	sym, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' has no member '%s'", pkgName, ident.Name)
	}
	if isPrivateSymbol(sym) {
		return nil, fmt.Errorf("dyn package '%s' member '%s' is private", pkgName, ident.Name)
	}
	return runtime.DynRefValue{Package: pkgName, Name: ident.Name}, nil
}

func (i *Interpreter) implNamespaceMember(ns runtime.ImplementationNamespaceValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Impl namespace member access expects identifier")
	}
	if ns.Methods == nil {
		return nil, fmt.Errorf("Impl namespace has no methods")
	}
	method, ok := ns.Methods[ident.Name]
	if !ok {
		name := "<impl>"
		if ns.Name != nil {
			name = ns.Name.Name
		}
		return nil, fmt.Errorf("No method '%s' on impl %s", ident.Name, name)
	}
	return method, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := ""
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		ifaceName = val.Interface.Node.ID.Name
	}
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	var method *runtime.FunctionValue
	if val.Methods != nil {
		method = val.Methods[ident.Name]
	}
	if method == nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			resolved, err := i.findMethod(info, ident.Name, ifaceName)
			if err != nil {
				return nil, err
			}
			method = resolved
			if method != nil {
				if val.Methods == nil {
					val.Methods = make(map[string]*runtime.FunctionValue)
				}
				val.Methods[ident.Name] = method
			}
		}
	}
	if method == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, ifaceName)
	}
	return &runtime.BoundMethodValue{Receiver: val.Underlying, Method: method}, nil
}

func (i *Interpreter) resolveDynRef(ref runtime.DynRefValue) (*runtime.FunctionValue, error) {
	bucket, ok := i.packageRegistry[ref.Package]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	val, ok := bucket[ref.Name]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	if fn, ok := val.(*runtime.FunctionValue); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("dyn ref '%s.%s' is not callable", ref.Package, ref.Name)
}

func toStructDefinitionValue(val runtime.Value, name string) (*runtime.StructDefinitionValue, error) {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		return v, nil
	case runtime.StructDefinitionValue:
		return &v, nil
	default:
		return nil, fmt.Errorf("'%s' is not a struct type", name)
	}
}

func (i *Interpreter) assignPattern(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, isDeclaration bool) error {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return declareOrAssign(env, p.Name, value, isDeclaration)
	case *ast.WildcardPattern:
		return nil
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return fmt.Errorf("invalid literal in pattern: %T", p.Literal)
		}
		litVal, err := i.evaluateExpression(litExpr, env)
		if err != nil {
			return err
		}
		if !valuesEqual(litVal, value) {
			return fmt.Errorf("pattern literal mismatch")
		}
		return nil
	case *ast.StructPattern:
		if errVal, ok := value.(runtime.ErrorValue); ok {
			value = errorValueToStructInstance(errVal)
		}
		if errValPtr, ok := value.(*runtime.ErrorValue); ok {
			value = errorValueToStructInstance(*errValPtr)
		}
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return fmt.Errorf("Cannot destructure non-struct value")
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return fmt.Errorf("Struct type mismatch in destructuring")
			}
		}
		if p.IsPositional {
			if structVal.Positional == nil {
				return fmt.Errorf("Expected positional struct")
			}
			if len(p.Fields) != len(structVal.Positional) {
				return fmt.Errorf("Struct field count mismatch")
			}
			for idx, field := range p.Fields {
				if field == nil {
					return fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return fmt.Errorf("missing positional struct value at index %d", idx)
				}
				if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
					return err
				}
				if field.Binding != nil {
					if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if structVal.Fields == nil {
			return fmt.Errorf("Expected named struct")
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return fmt.Errorf("Named struct pattern missing field name")
			}
			fieldVal, ok := structVal.Fields[field.FieldName.Name]
			if !ok {
				return fmt.Errorf("Missing field '%s' during destructuring", field.FieldName.Name)
			}
			if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
				return err
			}
			if field.Binding != nil {
				if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
					return err
				}
			}
		}
		return nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return fmt.Errorf("Cannot destructure non-array value")
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return fmt.Errorf("Array length mismatch in destructuring")
		}
		if len(elements) < len(p.Elements) {
			return fmt.Errorf("Array too short for destructuring")
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return fmt.Errorf("invalid array pattern at index %d", idx)
			}
			elemVal := elements[idx]
			if err := i.assignPattern(elemPattern, elemVal, env, isDeclaration); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
				restVal := &runtime.ArrayValue{Elements: restElems}
				if err := declareOrAssign(env, rest.Name, restVal, isDeclaration); err != nil {
					return err
				}
			case *ast.WildcardPattern:
				// ignore remaining elements
			default:
				return fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return fmt.Errorf("array length mismatch in destructuring")
		}
		return nil
	case *ast.TypedPattern:
		if !i.matchesType(p.TypeAnnotation, value) {
			return fmt.Errorf("Typed pattern mismatch in assignment")
		}
		coerced, err := i.coerceValueToType(p.TypeAnnotation, value)
		if err != nil {
			return err
		}
		return i.assignPattern(p.Pattern, coerced, env, isDeclaration)
	default:
		return fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func errorValueToStructInstance(err runtime.ErrorValue) *runtime.StructInstanceValue {
	fields := make(map[string]runtime.Value)
	if err.Payload != nil {
		for k, v := range err.Payload {
			fields[k] = v
		}
	}
	fields["message"] = runtime.StringValue{Val: err.Message}
	return &runtime.StructInstanceValue{Fields: fields}
}

func (i *Interpreter) matchPattern(pattern ast.Pattern, value runtime.Value, base *runtime.Environment) (*runtime.Environment, bool) {
	if pattern == nil {
		return nil, false
	}
	matchEnv := runtime.NewEnvironment(base)
	if err := i.assignPattern(pattern, value, matchEnv, true); err != nil {
		return nil, false
	}
	return matchEnv, true
}

func declareOrAssign(env *runtime.Environment, name string, value runtime.Value, isDeclaration bool) error {
	if isDeclaration {
		env.Define(name, value)
		return nil
	}
	return env.Assign(name, value)
}

func (i *Interpreter) getTypeInfoForValue(value runtime.Value) (typeInfo, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		return i.typeInfoFromStructInstance(v)
	case *runtime.InterfaceValue:
		return i.getTypeInfoForValue(v.Underlying)
	default:
		return typeInfo{}, false
	}
}

func (i *Interpreter) lookupImplEntry(info typeInfo, interfaceName string) (*implCandidate, error) {
	matches, err := i.collectImplCandidates(info, interfaceName)
	if len(matches) == 0 {
		return nil, err
	}
	best, ambiguous := i.selectBestCandidate(matches)
	if ambiguous != nil {
		detail := descriptionsFromCandidates(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Ambiguous impl for interface '%s' on type '%s' (candidates: %s)", interfaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	return best, nil
}

func (i *Interpreter) findMethod(info typeInfo, methodName string, interfaceFilter string) (*runtime.FunctionValue, error) {
	matches, err := i.collectImplCandidates(info, interfaceFilter)
	if len(matches) == 0 {
		return nil, err
	}
	methodMatches := make([]methodMatch, 0, len(matches))
	for _, cand := range matches {
		method := cand.entry.methods[methodName]
		if method == nil {
			if ifaceDef, ok := i.interfaces[cand.entry.interfaceName]; ok && ifaceDef.Node != nil {
				for _, sig := range ifaceDef.Node.Signatures {
					if sig == nil || sig.Name == nil || sig.Name.Name != methodName || sig.DefaultImpl == nil {
						continue
					}
					defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
					method = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
					if cand.entry.methods == nil {
						cand.entry.methods = make(map[string]*runtime.FunctionValue)
					}
					cand.entry.methods[methodName] = method
					break
				}
			}
		}
		if method == nil {
			continue
		}
		methodMatches = append(methodMatches, methodMatch{candidate: cand, method: method})
	}
	if len(methodMatches) == 0 {
		return nil, err
	}
	best, ambiguous := i.selectBestMethodCandidate(methodMatches)
	if ambiguous != nil {
		detail := descriptionsFromMethodMatches(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Ambiguous method '%s' for type '%s' (candidates: %s)", methodName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	if fnDef, ok := best.method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", methodName, info.name)
	}
	return best.method, nil
}

func (i *Interpreter) interfaceMatches(val *runtime.InterfaceValue, interfaceName string) bool {
	if val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == interfaceName {
			return true
		}
	}
	info, ok := i.getTypeInfoForValue(val.Underlying)
	if !ok {
		return false
	}
	entry, err := i.lookupImplEntry(info, interfaceName)
	return err == nil && entry != nil
}

func (i *Interpreter) selectStructMethod(inst *runtime.StructInstanceValue, methodName string) (*runtime.FunctionValue, error) {
	if inst == nil {
		return nil, nil
	}
	info, ok := i.typeInfoFromStructInstance(inst)
	if !ok {
		return nil, nil
	}
	return i.findMethod(info, methodName, "")
}

func (i *Interpreter) matchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := t.Name.Name
		switch name {
		case "string":
			_, ok := value.(runtime.StringValue)
			return ok
		case "bool":
			_, ok := value.(runtime.BoolValue)
			return ok
		case "char":
			_, ok := value.(runtime.CharValue)
			return ok
		case "nil":
			_, ok := value.(runtime.NilValue)
			return ok
		case "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128":
			iv, ok := value.(runtime.IntegerValue)
			if !ok {
				return false
			}
			return string(iv.TypeSuffix) == name
		case "f32", "f64":
			fv, ok := value.(runtime.FloatValue)
			if !ok {
				return false
			}
			return string(fv.TypeSuffix) == name
		case "Error":
			_, ok := value.(runtime.ErrorValue)
			return ok
		default:
			if _, ok := i.interfaces[name]; ok {
				switch v := value.(type) {
				case *runtime.InterfaceValue:
					return i.interfaceMatches(v, name)
				case runtime.InterfaceValue:
					return i.interfaceMatches(&v, name)
				default:
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						return false
					}
					if candidate, err := i.lookupImplEntry(info, name); err == nil && candidate != nil {
						return true
					}
					return false
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					return structVal.Definition.Node.ID.Name == name
				}
			}
			return false
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name.Name == "Array" {
			arr, ok := value.(*runtime.ArrayValue)
			if !ok {
				return false
			}
			if len(t.Arguments) == 0 {
				return true
			}
			elemType := t.Arguments[0]
			for _, el := range arr.Elements {
				if !i.matchesType(elemType, el) {
					return false
				}
			}
			return true
		}
		return true
	case *ast.FunctionTypeExpression:
		_, ok := value.(*runtime.FunctionValue)
		return ok
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return true
		}
		return i.matchesType(t.InnerType, value)
	case *ast.ResultTypeExpression:
		return i.matchesType(t.InnerType, value)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if i.matchesType(member, value) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (i *Interpreter) coerceValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			name := t.Name.Name
			if _, ok := i.interfaces[name]; ok {
				return i.coerceToInterfaceValue(name, value)
			}
		}
	}
	return value, nil
}

func (i *Interpreter) coerceToInterfaceValue(interfaceName string, value runtime.Value) (runtime.Value, error) {
	if ifaceVal, ok := value.(*runtime.InterfaceValue); ok {
		if i.interfaceMatches(ifaceVal, interfaceName) {
			return value, nil
		}
	}
	if ifaceVal, ok := value.(runtime.InterfaceValue); ok {
		if i.interfaceMatches(&ifaceVal, interfaceName) {
			return value, nil
		}
	}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", interfaceName)
	}
	info, ok := i.getTypeInfoForValue(value)
	if !ok {
		return nil, fmt.Errorf("Value does not implement interface %s", interfaceName)
	}
	candidate, err := i.lookupImplEntry(info, interfaceName)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.entry == nil {
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Type '%s' does not implement interface %s", typeDesc, interfaceName)
	}
	methods := make(map[string]*runtime.FunctionValue, len(candidate.entry.methods))
	for name, fn := range candidate.entry.methods {
		methods[name] = fn
	}
	return &runtime.InterfaceValue{Interface: ifaceDef, Underlying: value, Methods: methods}, nil
}

func rangeEndpoint(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		return int(v.Val.Int64()), nil
	case runtime.FloatValue:
		return int(v.Val), nil
	default:
		return 0, fmt.Errorf("range endpoint must be numeric, got %s", val.Kind())
	}
}
