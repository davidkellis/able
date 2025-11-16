package ast

import "reflect"

// SetSpan annotates the node with the provided span.
func SetSpan(node Node, span Span) {
	if node == nil {
		return
	}
	if setter, ok := node.(interface{ setSpan(Span) }); ok {
		setter.setSpan(span)
	}
}

// ZeroSpan returns an empty span value.
func ZeroSpan() Span {
	return Span{}
}

// CopySpans traverses src and applies any available span metadata to the corresponding
// nodes in dst. Mismatched shapes are ignored so partially decoded fixtures still load.
func CopySpans(dst, src Node) {
	copySpans(dst, src, make(map[Node]struct{}))
}

func copySpans(dst, src Node, visited map[Node]struct{}) {
	if dst == nil || src == nil {
		return
	}
	if _, ok := visited[dst]; ok {
		return
	}
	visited[dst] = struct{}{}
	if span := src.Span(); span != (Span{}) {
		SetSpan(dst, span)
	}
	copySpanValue(derefValue(reflect.ValueOf(dst)), derefValue(reflect.ValueOf(src)), visited)
}

func copySpanValue(dstVal, srcVal reflect.Value, visited map[Node]struct{}) {
	if !dstVal.IsValid() || !srcVal.IsValid() {
		return
	}
	if dstVal.Type() != srcVal.Type() {
		return
	}
	switch dstVal.Kind() {
	case reflect.Pointer:
		if dstVal.IsNil() || srcVal.IsNil() {
			return
		}
		if dstNode, ok := dstVal.Interface().(Node); ok {
			if srcNode, ok := srcVal.Interface().(Node); ok {
				copySpans(dstNode, srcNode, visited)
				return
			}
		}
		copySpanValue(dstVal.Elem(), srcVal.Elem(), visited)
	case reflect.Interface:
		if dstVal.IsNil() || srcVal.IsNil() {
			return
		}
		copySpanValue(dstVal.Elem(), srcVal.Elem(), visited)
	case reflect.Struct:
		numFields := dstVal.NumField()
		for i := 0; i < numFields; i++ {
			copySpanValue(dstVal.Field(i), srcVal.Field(i), visited)
		}
	case reflect.Slice, reflect.Array:
		length := dstVal.Len()
		if other := srcVal.Len(); other < length {
			length = other
		}
		for i := 0; i < length; i++ {
			copySpanValue(dstVal.Index(i), srcVal.Index(i), visited)
		}
	case reflect.Map:
		iter := dstVal.MapRange()
		for iter.Next() {
			key := iter.Key()
			dstElem := iter.Value()
			srcElem := srcVal.MapIndex(key)
			if !srcElem.IsValid() {
				continue
			}
			copySpanValue(dstElem, srcElem, visited)
		}
	}
}

func derefValue(val reflect.Value) reflect.Value {
	for val.IsValid() && (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) {
		if val.IsNil() {
			return val
		}
		val = val.Elem()
	}
	return val
}
