package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) addExternCallable(pkgName string, name string) {
	if g == nil {
		return
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return
	}
	if g.externCallables == nil {
		g.externCallables = make(map[string]map[string]struct{})
	}
	bucket := g.externCallables[pkgName]
	if bucket == nil {
		bucket = make(map[string]struct{})
		g.externCallables[pkgName] = bucket
	}
	bucket[name] = struct{}{}
}

func (g *generator) externCallableExists(pkgName string, name string) bool {
	if g == nil || g.externCallables == nil {
		return false
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return false
	}
	bucket := g.externCallables[pkgName]
	if len(bucket) == 0 {
		return false
	}
	_, ok := bucket[name]
	return ok
}

func (g *generator) externCallableNames(pkgName string) []string {
	if g == nil || g.externCallables == nil {
		return nil
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" {
		return nil
	}
	bucket := g.externCallables[pkgName]
	if len(bucket) == 0 {
		return nil
	}
	names := make([]string, 0, len(bucket))
	for name := range bucket {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		names = append(names, trimmed)
	}
	sort.Strings(names)
	return names
}

func (g *generator) diagNodeName(node ast.Node, goType string, prefix string) string {
	if node == nil {
		return "nil"
	}
	if g.diagNames == nil {
		g.diagNames = make(map[ast.Node]string)
	}
	if name, ok := g.diagNames[node]; ok {
		return name
	}
	name := fmt.Sprintf("__able_%s_node_%d", prefix, len(g.diagNodes))
	info := diagNodeInfo{
		Name:   name,
		GoType: goType,
		Span:   node.Span(),
	}
	if call, ok := node.(*ast.FunctionCall); ok && call != nil {
		switch callee := call.Callee.(type) {
		case *ast.Identifier:
			info.CallName = callee.Name
		case *ast.MemberAccessExpression:
			if member, ok := callee.Member.(*ast.Identifier); ok && member != nil {
				info.CallMember = member.Name
			}
		}
	}
	if g.nodeOrigins != nil {
		if origin, ok := g.nodeOrigins[node]; ok {
			info.Origin = origin
		}
	}
	g.diagNodes = append(g.diagNodes, info)
	g.diagNames[node] = name
	g.needsAst = true
	return name
}
