package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type interfaceDispatchEntry struct {
	Impl             *implMethodInfo
	TargetExpr       string
	InterfaceArgs    string
	GenericNames     string
	Constraints      string
	TargetKey        string
	InterfaceArgsKey string
	ConstraintKey    string
}

type interfaceDispatchGroup struct {
	InterfaceName    string
	MethodName       string
	TargetExpr       string
	InterfaceArgs    string
	GenericNames     string
	Constraints      string
	TargetKey        string
	InterfaceArgsKey string
	ConstraintKey    string
	IsPrivate        bool
	Entries          []*implMethodInfo
}

func (g *generator) implMethodDispatchable(impl *implMethodInfo) bool {
	if impl == nil || impl.Info == nil || !impl.Info.Compileable {
		return false
	}
	if impl.ImplName != "" {
		return false
	}
	if impl.TargetType == nil {
		return false
	}
	return true
}

func (g *generator) interfaceDispatchStrict() bool {
	if g == nil {
		return false
	}
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil {
			continue
		}
		if impl.ImplName != "" {
			continue
		}
		if !impl.Info.Compileable {
			return false
		}
	}
	return true
}

func (g *generator) interfaceDispatchEntries() []*interfaceDispatchEntry {
	if g == nil {
		return nil
	}
	counts := make(map[string]int)
	entries := make([]*interfaceDispatchEntry, 0)
	for _, impl := range g.sortedImplMethodInfos() {
		if !g.implMethodDispatchable(impl) {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(impl.TargetType)
		if !ok {
			continue
		}
		ifaceArgsExpr, ok := g.renderTypeExpressionList(impl.InterfaceArgs)
		if !ok {
			continue
		}
		constraintsExpr, constraintKey, ok := g.renderConstraintSpecs(impl.ImplGenerics, impl.WhereClause)
		if !ok {
			continue
		}
		targetKey := typeExpressionToString(impl.TargetType)
		ifaceKey := ""
		if len(impl.InterfaceArgs) > 0 {
			parts := make([]string, 0, len(impl.InterfaceArgs))
			for _, arg := range impl.InterfaceArgs {
				parts = append(parts, typeExpressionToString(arg))
			}
			ifaceKey = strings.Join(parts, "|")
		}
		genericNames := make([]string, 0, len(impl.ImplGenerics))
		for _, gp := range impl.ImplGenerics {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				continue
			}
			genericNames = append(genericNames, gp.Name.Name)
		}
		sort.Strings(genericNames)
		genericExpr := "nil"
		if len(genericNames) > 0 {
			quoted := make([]string, 0, len(genericNames))
			for _, name := range genericNames {
				quoted = append(quoted, fmt.Sprintf("%q", name))
			}
			genericExpr = fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", "))
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s", impl.InterfaceName, impl.MethodName, targetKey, ifaceKey, constraintKey)
		counts[key]++
		entries = append(entries, &interfaceDispatchEntry{
			Impl:             impl,
			TargetExpr:       targetExpr,
			InterfaceArgs:    ifaceArgsExpr,
			GenericNames:     genericExpr,
			Constraints:      constraintsExpr,
			TargetKey:        targetKey,
			InterfaceArgsKey: ifaceKey,
			ConstraintKey:    constraintKey,
		})
	}
	if len(entries) == 0 {
		return nil
	}
	filtered := make([]*interfaceDispatchEntry, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s", entry.Impl.InterfaceName, entry.Impl.MethodName, entry.TargetKey, entry.InterfaceArgsKey, entry.ConstraintKey)
		if counts[key] != 1 {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func (g *generator) interfaceDispatchGroups() []*interfaceDispatchGroup {
	if g == nil {
		return nil
	}
	groups := make(map[string]*interfaceDispatchGroup)
	keys := make([]string, 0)
	for _, impl := range g.sortedImplMethodInfos() {
		if !g.implMethodDispatchable(impl) {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(impl.TargetType)
		if !ok {
			continue
		}
		ifaceArgsExpr, ok := g.renderTypeExpressionList(impl.InterfaceArgs)
		if !ok {
			continue
		}
		constraintsExpr, constraintKey, ok := g.renderConstraintSpecs(impl.ImplGenerics, impl.WhereClause)
		if !ok {
			continue
		}
		targetKey := typeExpressionToString(impl.TargetType)
		ifaceKey := ""
		if len(impl.InterfaceArgs) > 0 {
			parts := make([]string, 0, len(impl.InterfaceArgs))
			for _, arg := range impl.InterfaceArgs {
				parts = append(parts, typeExpressionToString(arg))
			}
			ifaceKey = strings.Join(parts, "|")
		}
		genericNames := make([]string, 0, len(impl.ImplGenerics))
		for _, gp := range impl.ImplGenerics {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				continue
			}
			genericNames = append(genericNames, gp.Name.Name)
		}
		sort.Strings(genericNames)
		genericExpr := "nil"
		if len(genericNames) > 0 {
			quoted := make([]string, 0, len(genericNames))
			for _, name := range genericNames {
				quoted = append(quoted, fmt.Sprintf("%q", name))
			}
			genericExpr = fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", "))
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s", impl.InterfaceName, impl.MethodName, targetKey, ifaceKey, constraintKey)
		group := groups[key]
		if group == nil {
			group = &interfaceDispatchGroup{
				InterfaceName:    impl.InterfaceName,
				MethodName:       impl.MethodName,
				TargetExpr:       targetExpr,
				InterfaceArgs:    ifaceArgsExpr,
				GenericNames:     genericExpr,
				Constraints:      constraintsExpr,
				TargetKey:        targetKey,
				InterfaceArgsKey: ifaceKey,
				ConstraintKey:    constraintKey,
				IsPrivate:        false,
				Entries:          make([]*implMethodInfo, 0, 1),
			}
			groups[key] = group
			keys = append(keys, key)
		}
		if impl.Info != nil && impl.Info.Definition != nil && impl.Info.Definition.IsPrivate {
			group.IsPrivate = true
		}
		group.Entries = append(group.Entries, impl)
	}
	if len(groups) == 0 {
		return nil
	}
	sort.Strings(keys)
	results := make([]*interfaceDispatchGroup, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		sort.Slice(group.Entries, func(i, j int) bool {
			left := group.Entries[i]
			right := group.Entries[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.Info == nil || right.Info == nil {
				return left.Info != nil
			}
			return left.Info.GoName < right.Info.GoName
		})
		results = append(results, group)
	}
	return results
}

type constraintSpec struct {
	subject ast.TypeExpression
	iface   ast.TypeExpression
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
			specs = append(specs, constraintSpec{
				subject: ast.NewSimpleTypeExpression(gp.Name),
				iface:   constraint.InterfaceType,
			})
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
			specs = append(specs, constraintSpec{
				subject: clause.TypeParam,
				iface:   constraint.InterfaceType,
			})
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
		parts = append(parts, fmt.Sprintf("%s->%s", typeExpressionToString(spec.subject), typeExpressionToString(spec.iface)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func (g *generator) renderConstraintSpecs(generics []*ast.GenericParameter, whereClause []*ast.WhereClauseConstraint) (string, string, bool) {
	specs := collectConstraintSpecs(generics, whereClause)
	if len(specs) == 0 {
		return "nil", "<none>", true
	}
	entries := make([]string, 0, len(specs))
	for _, spec := range specs {
		subjectExpr, ok := g.renderTypeExpression(spec.subject)
		if !ok {
			return "", "", false
		}
		ifaceExpr, ok := g.renderTypeExpression(spec.iface)
		if !ok {
			return "", "", false
		}
		entries = append(entries, fmt.Sprintf("{subject: %s, iface: %s}", subjectExpr, ifaceExpr))
	}
	return fmt.Sprintf("[]__able_interface_constraint_spec{%s}", strings.Join(entries, ", ")), constraintSignature(specs), true
}

func (g *generator) interfaceSearchMap() map[string][]string {
	if g == nil || len(g.interfaces) == 0 {
		return nil
	}
	names := make([]string, 0, len(g.interfaces))
	for name := range g.interfaces {
		names = append(names, name)
	}
	sort.Strings(names)
	search := make(map[string][]string, len(names))
	for _, name := range names {
		search[name] = g.interfaceSearchNames(name, make(map[string]struct{}))
	}
	return search
}

func (g *generator) interfaceSearchNames(name string, visited map[string]struct{}) []string {
	return g.interfaceSearchNamesForPackage("", name, visited)
}

func (g *generator) interfaceSearchNamesForPackage(pkgName string, name string, visited map[string]struct{}) []string {
	if name == "" {
		return nil
	}
	key := strings.TrimSpace(pkgName) + "::" + name
	if _, seen := visited[key]; seen {
		return nil
	}
	visited[key] = struct{}{}
	names := []string{name}
	iface, ifacePkg, ok := g.interfaceDefinitionForPackage(pkgName, name)
	if !ok || iface == nil {
		return names
	}
	for _, base := range iface.BaseInterfaces {
		baseName, ok := typeExprBaseName(base)
		if !ok || baseName == "" {
			continue
		}
		basePkg := ifacePkg
		if resolvedPkg, _, _, _, ok := interfaceExprInfo(g, ifacePkg, base); ok && resolvedPkg != "" {
			basePkg = resolvedPkg
		}
		names = append(names, g.interfaceSearchNamesForPackage(basePkg, baseName, visited)...)
	}
	return names
}

func (g *generator) interfaceBaseTypesMap() map[string][]ast.TypeExpression {
	if g == nil || len(g.interfaces) == 0 {
		return nil
	}
	baseTypes := make(map[string][]ast.TypeExpression)
	for name, iface := range g.interfaces {
		if iface == nil || len(iface.BaseInterfaces) == 0 {
			continue
		}
		baseTypes[name] = append([]ast.TypeExpression(nil), iface.BaseInterfaces...)
	}
	return baseTypes
}
