package interpreter

import (
	"sort"

	"able/interpreter-go/pkg/ast"
)

type bytecodeTraceKey struct {
	Op       string
	Name     string
	Lookup   string
	Dispatch string
	Origin   string
	Line     int
	Column   int
}

type BytecodeTraceEntry struct {
	Hits     uint64 `json:"hits"`
	Op       string `json:"op"`
	Name     string `json:"name,omitempty"`
	Lookup   string `json:"lookup,omitempty"`
	Dispatch string `json:"dispatch,omitempty"`
	Origin   string `json:"origin,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

type BytecodeTraceSnapshot struct {
	Enabled   bool                 `json:"enabled"`
	TotalHits uint64               `json:"total_hits"`
	Entries   []BytecodeTraceEntry `json:"entries"`
}

func (i *Interpreter) recordBytecodeCallTrace(op string, name string, lookup string, dispatch string, node ast.Node) {
	if i == nil || !i.bytecodeTraceEnabled {
		return
	}
	key := bytecodeTraceKey{
		Op:       op,
		Name:     name,
		Lookup:   lookup,
		Dispatch: dispatch,
	}
	if node != nil {
		if span := node.Span(); span != (ast.Span{}) {
			key.Line = span.Start.Line
			key.Column = span.Start.Column
		}
		if i.nodeOrigins != nil {
			key.Origin = i.nodeOrigins[node]
		}
	}
	i.bytecodeTraceMu.Lock()
	if i.bytecodeTraceCounts == nil {
		i.bytecodeTraceCounts = make(map[bytecodeTraceKey]uint64, 16)
	}
	i.bytecodeTraceCounts[key]++
	i.bytecodeTraceMu.Unlock()
}

func (i *Interpreter) BytecodeTrace(limit int) BytecodeTraceSnapshot {
	snapshot := BytecodeTraceSnapshot{}
	if i == nil {
		return snapshot
	}
	snapshot.Enabled = i.bytecodeTraceEnabled
	if !i.bytecodeTraceEnabled {
		return snapshot
	}
	i.bytecodeTraceMu.Lock()
	defer i.bytecodeTraceMu.Unlock()

	if len(i.bytecodeTraceCounts) == 0 {
		return snapshot
	}
	snapshot.Entries = make([]BytecodeTraceEntry, 0, len(i.bytecodeTraceCounts))
	for key, hits := range i.bytecodeTraceCounts {
		snapshot.TotalHits += hits
		snapshot.Entries = append(snapshot.Entries, BytecodeTraceEntry{
			Hits:     hits,
			Op:       key.Op,
			Name:     key.Name,
			Lookup:   key.Lookup,
			Dispatch: key.Dispatch,
			Origin:   key.Origin,
			Line:     key.Line,
			Column:   key.Column,
		})
	}
	sort.Slice(snapshot.Entries, func(a, b int) bool {
		left := snapshot.Entries[a]
		right := snapshot.Entries[b]
		if left.Hits != right.Hits {
			return left.Hits > right.Hits
		}
		if left.Op != right.Op {
			return left.Op < right.Op
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Lookup != right.Lookup {
			return left.Lookup < right.Lookup
		}
		if left.Dispatch != right.Dispatch {
			return left.Dispatch < right.Dispatch
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Column < right.Column
	})
	if limit > 0 && len(snapshot.Entries) > limit {
		snapshot.Entries = snapshot.Entries[:limit]
	}
	return snapshot
}

func (i *Interpreter) ResetBytecodeTrace() {
	if i == nil || !i.bytecodeTraceEnabled {
		return
	}
	i.bytecodeTraceMu.Lock()
	clear(i.bytecodeTraceCounts)
	i.bytecodeTraceMu.Unlock()
}
