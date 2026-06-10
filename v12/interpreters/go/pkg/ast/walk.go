package ast

import "reflect"

// Walk visits every node reachable from root once. Returning false from visit
// skips that node's descendants; it does not stop traversal of sibling nodes.
func Walk(root Node, visit func(Node) bool) {
	if root == nil || visit == nil {
		return
	}
	walkNode(root, visit, make(map[Node]struct{}))
}

func walkNode(node Node, visit func(Node) bool, visited map[Node]struct{}) {
	if node == nil {
		return
	}
	if _, ok := visited[node]; ok {
		return
	}
	visited[node] = struct{}{}
	if !visit(node) {
		return
	}

	value := reflect.ValueOf(node)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		walkValue(value.Elem(), visit, visited)
		return
	}
	walkValue(value, visit, visited)
}

func walkValue(value reflect.Value, visit func(Node) bool, visited map[Node]struct{}) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		if value.CanInterface() {
			if node, ok := value.Interface().(Node); ok {
				walkNode(node, visit, visited)
				return
			}
		}
		walkValue(value.Elem(), visit, visited)
		return
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return
		}
		walkValue(value.Elem(), visit, visited)
		return
	}
	if value.Kind() == reflect.Struct && value.CanAddr() {
		pointer := value.Addr()
		if pointer.IsValid() && pointer.CanInterface() {
			if node, ok := pointer.Interface().(Node); ok {
				walkNode(node, visit, visited)
			}
		}
	}
	if value.CanInterface() && value.Kind() != reflect.Struct {
		if node, ok := value.Interface().(Node); ok {
			walkNode(node, visit, visited)
			return
		}
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			walkValue(value.Field(i), visit, visited)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			walkValue(value.Index(i), visit, visited)
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			walkValue(value.MapIndex(key), visit, visited)
		}
	}
}
