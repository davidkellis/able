package ast

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
