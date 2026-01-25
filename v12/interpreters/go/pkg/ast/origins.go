package ast

import "reflect"

// AnnotateOrigins assigns the provided source path to every node reachable from root.
// The table map may be nil; when provided it is populated with node -> path entries.
func AnnotateOrigins(root Node, path string, table map[Node]string) {
	if root == nil || path == "" {
		return
	}
	if table == nil {
		table = make(map[Node]string)
	}
	annotateOrigins(root, path, table, make(map[Node]struct{}))
}

func annotateOrigins(node Node, path string, table map[Node]string, visited map[Node]struct{}) {
	if node == nil {
		return
	}
	if _, ok := visited[node]; ok {
		return
	}
	visited[node] = struct{}{}
	if _, ok := table[node]; !ok {
		table[node] = path
	}
	val := reflect.ValueOf(node)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		annotateValue(val.Elem(), path, table, visited)
		return
	}
	annotateValue(val, path, table, visited)
}

func annotateValue(val reflect.Value, path string, table map[Node]string, visited map[Node]struct{}) {
	if !val.IsValid() {
		return
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		if node, ok := val.Interface().(Node); ok {
			annotateOrigins(node, path, table, visited)
			return
		}
		annotateValue(val.Elem(), path, table, visited)
		return
	}
	if val.Kind() == reflect.Interface {
		if val.IsNil() {
			return
		}
		annotateValue(val.Elem(), path, table, visited)
		return
	}
	if val.Kind() == reflect.Struct && val.CanAddr() {
		ptr := val.Addr()
		if ptr.IsValid() && ptr.CanInterface() {
			if node, ok := ptr.Interface().(Node); ok {
				annotateOrigins(node, path, table, visited)
			}
		}
	}
	if val.CanInterface() && val.Kind() != reflect.Struct {
		if node, ok := val.Interface().(Node); ok {
			annotateOrigins(node, path, table, visited)
			return
		}
	}
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			annotateValue(val.Field(i), path, table, visited)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			annotateValue(val.Index(i), path, table, visited)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			annotateValue(val.MapIndex(key), path, table, visited)
		}
	}
}
