package driver

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func joinIdentifiers(ids []*ast.Identifier) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == nil || id.Name == "" {
			continue
		}
		parts = append(parts, id.Name)
	}
	return strings.Join(parts, ".")
}

func buildIdentifiers(parts []string) []*ast.Identifier {
	out := make([]*ast.Identifier, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, ast.NewIdentifier(part))
	}
	return out
}

func copyIdentifiers(ids []*ast.Identifier) []*ast.Identifier {
	if len(ids) == 0 {
		return nil
	}
	out := make([]*ast.Identifier, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		out = append(out, ast.NewIdentifier(id.Name))
	}
	return out
}

func importKey(imp *ast.ImportStatement) string {
	if imp == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(joinIdentifiers(imp.PackagePath))
	sb.WriteString("|")
	if imp.IsWildcard {
		sb.WriteString("*")
	}
	sb.WriteString("|")
	if imp.Alias != nil {
		sb.WriteString(imp.Alias.Name)
	}
	if len(imp.Selectors) > 0 {
		sb.WriteString("|")
		for _, sel := range imp.Selectors {
			if sel == nil || sel.Name == nil {
				continue
			}
			sb.WriteString(sel.Name.Name)
			if sel.Alias != nil {
				sb.WriteString("::")
				sb.WriteString(sel.Alias.Name)
			}
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func sanitizeSegment(seg string) string {
	seg = strings.TrimSpace(seg)
	seg = strings.ReplaceAll(seg, "-", "_")
	return seg
}
