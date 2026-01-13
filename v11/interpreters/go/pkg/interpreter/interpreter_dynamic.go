package interpreter

import (
	"fmt"
	"math/big"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser"
	"able/interpreter-go/pkg/parser/language"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) initDynamicBuiltins() {
	defMethod := runtime.NativeFunctionValue{
		Name:  "dyn.Package.def",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return runtime.ErrorValue{Message: "dyn.Package.def expects code"}, nil
			}
			pkgName, ok := dynPackageName(args[0])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.def called on non-dyn package"}, nil
			}
			code, ok := asStringValue(i, args[1])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.def expects String"}, nil
			}
			return i.evaluateDynamicDefinition(pkgName, code), nil
		},
	}
	i.dynPackageDefMethod = defMethod

	evalMethod := runtime.NativeFunctionValue{
		Name:  "dyn.Package.eval",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return runtime.ErrorValue{Message: "dyn.Package.eval expects code"}, nil
			}
			pkgName, ok := dynPackageName(args[0])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.eval called on non-dyn package"}, nil
			}
			code, ok := asStringValue(i, args[1])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.eval expects String"}, nil
			}
			return i.evaluateDynamicEval(pkgName, code), nil
		},
	}
	i.dynPackageEvalMethod = evalMethod

	packageFn := runtime.NativeFunctionValue{
		Name:  "dyn.package",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			name, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.package expects String"}, nil
			}
			if _, ok := i.packageRegistry[name]; !ok {
				return runtime.ErrorValue{Message: fmt.Sprintf("dyn.package: package '%s' not found", name)}, nil
			}
			return runtime.DynPackageValue{Name: name}, nil
		},
	}

	defPackageFn := runtime.NativeFunctionValue{
		Name:  "dyn.def_package",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			name, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.def_package expects String"}, nil
			}
			i.ensureDynamicPackage(name)
			return runtime.DynPackageValue{Name: name}, nil
		},
	}

	evalFn := runtime.NativeFunctionValue{
		Name:  "dyn.eval",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			code, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.eval expects String"}, nil
			}
			return i.evaluateDynamicEval("dyn.eval", code), nil
		},
	}

	dynPkg := runtime.PackageValue{
		Name:      "dyn",
		NamePath:  []string{"dyn"},
		IsPrivate: false,
		Public: map[string]runtime.Value{
			"package":     packageFn,
			"def_package": defPackageFn,
			"eval":        evalFn,
		},
	}
	i.global.Define("dyn", dynPkg)
}

func firstArg(args []runtime.Value) runtime.Value {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

func asStringValue(i *Interpreter, val runtime.Value) (string, bool) {
	for {
		if iface, ok := val.(*runtime.InterfaceValue); ok && iface != nil {
			val = iface.Underlying
			continue
		}
		if iface, ok := val.(runtime.InterfaceValue); ok {
			val = iface.Underlying
			continue
		}
		break
	}
	str, err := i.coerceStringValue(val)
	if err != nil {
		return "", false
	}
	return str, true
}

func dynPackageName(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.DynPackageValue:
		if v.Name != "" {
			return v.Name, true
		}
		if len(v.NamePath) > 0 {
			return strings.Join(v.NamePath, "."), true
		}
	case *runtime.DynPackageValue:
		if v == nil {
			return "", false
		}
		if v.Name != "" {
			return v.Name, true
		}
		if len(v.NamePath) > 0 {
			return strings.Join(v.NamePath, "."), true
		}
	}
	return "", false
}

func (i *Interpreter) ensureDynamicPackage(name string) {
	if name == "" {
		return
	}
	if _, ok := i.packageRegistry[name]; !ok {
		i.packageRegistry[name] = make(map[string]runtime.Value)
	}
	if _, ok := i.packageMetadata[name]; !ok {
		parts := strings.Split(name, ".")
		i.packageMetadata[name] = packageMeta{namePath: parts, isPrivate: false}
	}
}

func (i *Interpreter) evaluateDynamicDefinition(pkgName, source string) runtime.Value {
	if pkgName == "" {
		return runtime.ErrorValue{Message: "dyn.def requires package name"}
	}
	i.ensureDynamicPackage(pkgName)
	mod, err := parseDynamicModule(source)
	if err != nil {
		return runtime.ErrorValue{Message: fmt.Sprintf("dyn.def parse error: %s", err.Error())}
	}
	baseParts := strings.Split(pkgName, ".")
	var targetParts []string
	if mod.Package != nil {
		pkgParts := identifiersToStrings(mod.Package.NamePath)
		targetParts = resolveDynamicPackage(baseParts, pkgParts)
	} else {
		targetParts = baseParts
	}
	mod.Package = ast.NewPackageStatement(stringsToIdentifiers(targetParts), false)

	prevDynamic := i.dynamicDefinitionMode
	prevTypecheck := i.typecheckerEnabled
	prevStrict := i.typecheckerStrict
	i.dynamicDefinitionMode = true
	i.typecheckerEnabled = false
	i.typecheckerStrict = false
	_, _, evalErr := i.EvaluateModule(mod)
	i.dynamicDefinitionMode = prevDynamic
	i.typecheckerEnabled = prevTypecheck
	i.typecheckerStrict = prevStrict

	if evalErr != nil {
		switch v := evalErr.(type) {
		case raiseSignal:
			return i.makeErrorValue(v.value, i.global)
		default:
			return runtime.ErrorValue{Message: fmt.Sprintf("dyn.def error: %s", evalErr.Error())}
		}
	}
	return runtime.NilValue{}
}

func (i *Interpreter) evaluateDynamicEval(pkgName, source string) runtime.Value {
	if pkgName == "" {
		return runtime.ErrorValue{Message: "dyn.eval requires package name"}
	}
	tree, root, err := parseDynamicTree([]byte(source))
	if err != nil {
		return runtime.ErrorValue{Message: fmt.Sprintf("dyn.eval parse error: %s", err.Error())}
	}
	if root == nil || root.Kind() != "source_file" {
		if tree != nil {
			tree.Close()
		}
		return i.makeParseErrorValue(parseErrorInfo{
			message:      fmt.Sprintf("parse error: root %s", nodeKind(root)),
			startByte:    0,
			endByte:      0,
			isIncomplete: false,
		})
	}
	if root.HasError() {
		info := parseErrorInfoFromTree(root, []byte(source))
		if tree != nil {
			tree.Close()
		}
		return i.makeParseErrorValue(info)
	}
	if tree != nil {
		tree.Close()
	}
	mod, err := parseDynamicModule(source)
	if err != nil {
		return i.makeParseErrorValue(parseErrorInfo{
			message:      fmt.Sprintf("parse error: %s", err.Error()),
			startByte:    0,
			endByte:      0,
			isIncomplete: false,
		})
	}
	baseParts := strings.Split(pkgName, ".")
	var targetParts []string
	if mod.Package != nil {
		pkgParts := identifiersToStrings(mod.Package.NamePath)
		targetParts = resolveDynamicPackage(baseParts, pkgParts)
	} else {
		targetParts = baseParts
	}
	mod.Package = ast.NewPackageStatement(stringsToIdentifiers(targetParts), false)

	prevDynamic := i.dynamicDefinitionMode
	prevTypecheck := i.typecheckerEnabled
	prevStrict := i.typecheckerStrict
	i.dynamicDefinitionMode = true
	i.typecheckerEnabled = false
	i.typecheckerStrict = false
	result, _, evalErr := i.EvaluateModule(mod)
	i.dynamicDefinitionMode = prevDynamic
	i.typecheckerEnabled = prevTypecheck
	i.typecheckerStrict = prevStrict

	if evalErr != nil {
		switch v := evalErr.(type) {
		case raiseSignal:
			return i.makeErrorValue(v.value, i.global)
		default:
			return runtime.ErrorValue{Message: fmt.Sprintf("dyn.eval error: %s", evalErr.Error())}
		}
	}
	if result == nil {
		return runtime.NilValue{}
	}
	return result
}

func parseDynamicModule(source string) (*ast.Module, error) {
	p, err := parser.NewModuleParser()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	return p.ParseModule([]byte(source))
}

func resolveDynamicPackage(base, target []string) []string {
	if len(target) == 0 {
		return base
	}
	if len(target) > 0 && target[0] == "root" {
		return target
	}
	if len(base) > 0 && len(target) >= len(base) {
		matches := true
		for i := range base {
			if target[i] != base[i] {
				matches = false
				break
			}
		}
		if matches {
			return target
		}
	}
	combined := make([]string, 0, len(base)+len(target))
	combined = append(combined, base...)
	combined = append(combined, target...)
	return combined
}

func stringsToIdentifiers(parts []string) []*ast.Identifier {
	idents := make([]*ast.Identifier, 0, len(parts))
	for _, part := range parts {
		idents = append(idents, ast.NewIdentifier(part))
	}
	return idents
}

type parseErrorInfo struct {
	message      string
	startByte    uint
	endByte      uint
	isIncomplete bool
}

func parseDynamicTree(source []byte) (*sitter.Tree, *sitter.Node, error) {
	lang := language.Able()
	if lang == nil {
		return nil, nil, fmt.Errorf("parser: able language not available")
	}
	p := sitter.NewParser()
	if err := p.SetLanguage(lang); err != nil {
		p.Close()
		return nil, nil, fmt.Errorf("parser: %w", err)
	}
	tree := p.Parse(source, nil)
	p.Close()
	if tree == nil {
		return nil, nil, fmt.Errorf("parser: no tree")
	}
	return tree, tree.RootNode(), nil
}

func nodeKind(node *sitter.Node) string {
	if node == nil {
		return "<nil>"
	}
	return node.Kind()
}

func parseErrorInfoFromTree(root *sitter.Node, source []byte) parseErrorInfo {
	endByte := lastNonWhitespaceByte(source)
	var firstError *sitter.Node
	var incomplete *sitter.Node
	stack := []*sitter.Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == nil {
			continue
		}
		isMissing := node.IsMissing()
		isError := node.IsError()
		if isMissing || isError {
			if firstError == nil {
				firstError = node
			}
			start := node.StartByte()
			end := node.EndByte()
			if (isMissing && start >= endByte) || (isError && end >= endByte) {
				incomplete = node
			}
		}
		for i := uint(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child != nil {
				stack = append(stack, child)
			}
		}
	}
	target := incomplete
	if target == nil {
		target = firstError
	}
	if target == nil {
		target = root
	}
	start := uint(0)
	end := uint(0)
	if target != nil {
		start = target.StartByte()
		end = target.EndByte()
	}
	return parseErrorInfo{
		message:      "parse error: syntax errors",
		startByte:    start,
		endByte:      end,
		isIncomplete: incomplete != nil,
	}
}

func lastNonWhitespaceByte(source []byte) uint {
	if len(source) == 0 {
		return 0
	}
	idx := len(source)
	for idx > 0 {
		b := source[idx-1]
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' && b != '\v' && b != '\f' {
			break
		}
		idx -= 1
	}
	if idx < 0 {
		idx = 0
	}
	return uint(idx)
}

func (i *Interpreter) makeParseErrorValue(info parseErrorInfo) runtime.Value {
	spanDef := i.resolveStandardErrorStruct("Span")
	parseDef := i.resolveStandardErrorStruct("ParseError")
	span := &runtime.StructInstanceValue{
		Definition: spanDef,
		Fields: map[string]runtime.Value{
			"start": runtime.IntegerValue{Val: big.NewInt(int64(info.startByte)), TypeSuffix: runtime.IntegerU64},
			"end":   runtime.IntegerValue{Val: big.NewInt(int64(info.endByte)), TypeSuffix: runtime.IntegerU64},
		},
	}
	parseVal := &runtime.StructInstanceValue{
		Definition: parseDef,
		Fields: map[string]runtime.Value{
			"message":       runtime.StringValue{Val: info.message},
			"span":          span,
			"is_incomplete": runtime.BoolValue{Val: info.isIncomplete},
		},
	}
	return runtime.ErrorValue{
		Message: info.message,
		Payload: map[string]runtime.Value{"value": parseVal},
	}
}
