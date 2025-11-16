package ast

import "testing"

func TestCopySpansPropagatesNestedNodes(t *testing.T) {
	src := sampleModule(true)
	dst := sampleModule(false)

	CopySpans(dst, src)

	if got, want := dst.Span(), src.Span(); got != want {
		t.Fatalf("module span mismatch: got %+v, want %+v", got, want)
	}

	srcFn, _ := src.Body[0].(*FunctionDefinition)
	dstFn, _ := dst.Body[0].(*FunctionDefinition)
	if got, want := dstFn.Span(), srcFn.Span(); got != want {
		t.Fatalf("function span mismatch: got %+v, want %+v", got, want)
	}
	if got, want := dstFn.ID.Span(), srcFn.ID.Span(); got != want {
		t.Fatalf("function identifier span mismatch: got %+v, want %+v", got, want)
	}

	srcParam := srcFn.Params[0]
	dstParam := dstFn.Params[0]
	if got, want := dstParam.Span(), srcParam.Span(); got != want {
		t.Fatalf("param span mismatch: got %+v, want %+v", got, want)
	}
	srcParamID, _ := srcParam.Name.(*Identifier)
	dstParamID, _ := dstParam.Name.(*Identifier)
	if got, want := dstParamID.Span(), srcParamID.Span(); got != want {
		t.Fatalf("param identifier span mismatch: got %+v, want %+v", got, want)
	}

	srcBody := srcFn.Body
	dstBody := dstFn.Body
	if got, want := dstBody.Span(), srcBody.Span(); got != want {
		t.Fatalf("block span mismatch: got %+v, want %+v", got, want)
	}
	srcStmt := srcBody.Body[0].(Expression)
	dstStmt := dstBody.Body[0].(Expression)
	if got, want := dstStmt.Span(), srcStmt.Span(); got != want {
		t.Fatalf("block statement span mismatch: got %+v, want %+v", got, want)
	}
}

func TestCopySpansHandlesNilAndMismatchedShapes(t *testing.T) {
	dst := NewIdentifier("dst")
	src := NewIdentifier("src")
	SetSpan(src, Span{Start: Position{Line: 5, Column: 7}})

	CopySpans(dst, src)
	if got, want := dst.Span(), src.Span(); got != want {
		t.Fatalf("identifier span mismatch: got %+v, want %+v", got, want)
	}

	CopySpans(dst, nil)
	CopySpans(nil, src)
	CopySpans(nil, nil)
}

func sampleModule(withSpans bool) *Module {
	pkg := NewPackageStatement([]*Identifier{NewIdentifier("pkg")}, false)
	fnName := NewIdentifier("format_point")
	paramName := NewIdentifier("value")
	param := NewFunctionParameter(paramName, nil)
	bodyIdent := NewIdentifier("value")
	body := NewBlockExpression([]Statement{bodyIdent})
	fn := NewFunctionDefinition(fnName, []*FunctionParameter{param}, body, nil, nil, nil, false, false)
	module := NewModule([]Statement{fn}, nil, pkg)

	if withSpans {
		SetSpan(module, Span{Start: Position{Line: 1, Column: 1}})
		SetSpan(fn, Span{Start: Position{Line: 2, Column: 1}})
		SetSpan(fnName, Span{Start: Position{Line: 2, Column: 4}})
		SetSpan(param, Span{Start: Position{Line: 2, Column: 16}})
		SetSpan(paramName, Span{Start: Position{Line: 2, Column: 16}})
		SetSpan(body, Span{Start: Position{Line: 3, Column: 1}})
		SetSpan(bodyIdent, Span{Start: Position{Line: 3, Column: 3}})
	}

	return module
}
