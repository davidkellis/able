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

var reservedIdents = map[string]struct{}{
	"ast":         {},
	"bridge":      {},
	"errors":      {},
	"fmt":         {},
	"interpreter": {},
	"runtime":     {},
	"sync":        {},
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
	if _, ok := reservedIdents[out]; ok {
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
	return base + "_" + alphaSuffix(count-1)
}

func alphaSuffix(index int) string {
	if index < 0 {
		return "a"
	}
	index++
	buf := make([]byte, 0, 4)
	for index > 0 {
		index--
		buf = append(buf, byte('a'+(index%26)))
		index /= 26
	}
	for left, right := 0, len(buf)-1; left < right; left, right = left+1, right-1 {
		buf[left], buf[right] = buf[right], buf[left]
	}
	return string(buf)
}
