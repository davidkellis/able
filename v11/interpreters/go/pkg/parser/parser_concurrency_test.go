package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseSpawnExpressionForms(t *testing.T) {
	source := `fn worker(n: i32) -> i32 {
  n
}

handle := spawn do {
  worker(1)
}

inline := spawn {
  worker(2)
}

call := spawn worker(3)
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	workerFn := ast.NewFunctionDefinition(
		ast.ID("worker"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("n"), ast.Ty("i32")),
		},
		ast.Block(ast.ID("n")),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	handleAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("handle"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("worker"),
					[]ast.Expression{ast.Int(1)},
					nil,
					false,
				),
			),
		),
	)

	inlineAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("inline"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("worker"),
					[]ast.Expression{ast.Int(2)},
					nil,
					false,
				),
			),
		),
	)

	callAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("call"),
		ast.NewSpawnExpression(
			ast.NewFunctionCall(
				ast.ID("worker"),
				[]ast.Expression{ast.Int(3)},
				nil,
				false,
			),
		),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			workerFn,
			handleAssign,
			inlineAssign,
			callAssign,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseSpawnExpressionFormsWithCallTargets(t *testing.T) {
	source := `fn task() -> i32 {
  1
}

fn run(n: i32) -> i32 {
  n
}

future := spawn do {
  task()
}

inline := spawn {
  task()
}

call := spawn run(3)
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	taskFn := ast.NewFunctionDefinition(
		ast.ID("task"),
		nil,
		ast.Block(ast.Int(1)),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	runFn := ast.NewFunctionDefinition(
		ast.ID("run"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("n"), ast.Ty("i32")),
		},
		ast.Block(ast.ID("n")),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	futureAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("future"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("task"),
					[]ast.Expression{},
					nil,
					false,
				),
			),
		),
	)

	inlineAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("inline"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("task"),
					[]ast.Expression{},
					nil,
					false,
				),
			),
		),
	)

	callAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("call"),
		ast.NewSpawnExpression(
			ast.NewFunctionCall(
				ast.ID("run"),
				[]ast.Expression{ast.Int(3)},
				nil,
				false,
			),
		),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			taskFn,
			runFn,
			futureAssign,
			inlineAssign,
			callAssign,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseSpawnHelpers(t *testing.T) {
	source := `handle := spawn do {
  future_yield()
  0
}

future_yield()
isCancelled := future_cancelled()
future_flush(handle)
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	handleAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("handle"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("future_yield"),
					[]ast.Expression{},
					nil,
					false,
				),
				ast.Int(0),
			),
		),
	)

	yieldCall := ast.NewFunctionCall(
		ast.ID("future_yield"),
		[]ast.Expression{},
		nil,
		false,
	)

	cancelAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("isCancelled"),
		ast.NewFunctionCall(
			ast.ID("future_cancelled"),
			[]ast.Expression{},
			nil,
			false,
		),
	)

	flushCall := ast.NewFunctionCall(
		ast.ID("future_flush"),
		[]ast.Expression{ast.ID("handle")},
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			handleAssign,
			yieldCall,
			cancelAssign,
			flushCall,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseChannelAndMutexHelpers(t *testing.T) {
	source := `channel := __able_channel_new(2)
__able_channel_send(channel, 1)
value := __able_channel_receive(channel)
trySend := __able_channel_try_send(channel, 3)
maybeValue := __able_channel_try_receive(channel)
isClosed := __able_channel_is_closed(channel)
__able_channel_close(channel)

mutex := __able_mutex_new()
__able_mutex_lock(mutex)
__able_mutex_unlock(mutex)
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	channelAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("channel"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_new"),
			[]ast.Expression{ast.Int(2)},
			nil,
			false,
		),
	)

	channelSend := ast.NewFunctionCall(
		ast.ID("__able_channel_send"),
		[]ast.Expression{ast.ID("channel"), ast.Int(1)},
		nil,
		false,
	)

	valueAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("value"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_receive"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	trySendAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("trySend"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_try_send"),
			[]ast.Expression{ast.ID("channel"), ast.Int(3)},
			nil,
			false,
		),
	)

	tryReceiveAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("maybeValue"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_try_receive"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	isClosedAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("isClosed"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_is_closed"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	closeCall := ast.NewFunctionCall(
		ast.ID("__able_channel_close"),
		[]ast.Expression{ast.ID("channel")},
		nil,
		false,
	)

	mutexAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("mutex"),
		ast.NewFunctionCall(
			ast.ID("__able_mutex_new"),
			[]ast.Expression{},
			nil,
			false,
		),
	)

	mutexLock := ast.NewFunctionCall(
		ast.ID("__able_mutex_lock"),
		[]ast.Expression{ast.ID("mutex")},
		nil,
		false,
	)

	mutexUnlock := ast.NewFunctionCall(
		ast.ID("__able_mutex_unlock"),
		[]ast.Expression{ast.ID("mutex")},
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			channelAssign,
			channelSend,
			valueAssign,
			trySendAssign,
			tryReceiveAssign,
			isClosedAssign,
			closeCall,
			mutexAssign,
			mutexLock,
			mutexUnlock,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}
