package compiler

import (
	"strings"
	"unicode"
)

var goKeywords = map[string]struct{}{
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
}

func sanitizeIdent(name string) string {
	if name == "" {
		return "_"
	}
	var b strings.Builder
	for i, r := range name {
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			if i == 0 && unicode.IsDigit(r) {
				b.WriteByte('_')
			}
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	out := b.String()
	if _, ok := goKeywords[out]; ok {
		return "_" + out
	}
	return out
}

func exportIdent(name string) string {
	safe := sanitizeIdent(name)
	if safe == "" {
		return "X"
	}
	return strings.ToUpper(safe[:1]) + safe[1:]
}

type nameMangler struct {
	seen map[string]int
}

func newNameMangler() *nameMangler {
	return &nameMangler{seen: make(map[string]int)}
}

func (m *nameMangler) unique(base string) string {
	if m == nil {
		return base
	}
	if base == "" {
		base = "_"
	}
	count := m.seen[base]
	m.seen[base] = count + 1
	if count == 0 {
		return base
	}
	return base + "_" + string('a'+rune(count-1))
}
