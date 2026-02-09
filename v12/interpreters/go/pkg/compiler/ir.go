package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

// Typed IR is A-normal form (ANF) with explicit control flow.
// - Evaluation order is the instruction order within a block.
// - Operations that may produce errors expose an explicit error value that must
//   be routed through an IRTerminator (ex: IRCheck) to model error flow.

type IRProgram struct {
	Packages     map[string]*IRPackage
	EntryPackage string
	EntryFunc    string
}

type IRPackage struct {
	Name      string
	Functions map[string]*IRFunction
	Init      *IRFunction
}

type IRFunction struct {
	Name       string
	Package    string
	Params     []IRParam
	Captured   []*IRSlot
	ReturnType typechecker.Type
	Locals     []*IRSlot
	Blocks     map[string]*IRBlock
	EntryLabel string
	Source     ast.Node
}

type IRParam struct {
	Value   *IRValue
	Mutable bool
}

type IRSlot struct {
	Name    string
	Type    typechecker.Type
	Mutable bool
	Source  ast.Node
}

type IRValue struct {
	ID     int
	Name   string
	Type   typechecker.Type
	Source ast.Node
}

type IRValueRef interface {
	irValueRefNode()
	Type() typechecker.Type
}

type IRValueUse struct {
	Value *IRValue
}

func (IRValueUse) irValueRefNode() {}

func (v IRValueUse) Type() typechecker.Type {
	if v.Value == nil {
		return nil
	}
	return v.Value.Type
}

type IRConst struct {
	Literal ast.Literal
	TypeRef typechecker.Type
}

func (IRConst) irValueRefNode() {}

func (c IRConst) Type() typechecker.Type {
	return c.TypeRef
}

type IRVoid struct{}

func (IRVoid) irValueRefNode() {}

func (IRVoid) Type() typechecker.Type { return nil }

type IRGlobal struct {
	Name    string
	TypeRef typechecker.Type
}

func (IRGlobal) irValueRefNode() {}

func (g IRGlobal) Type() typechecker.Type { return g.TypeRef }

type IROp string

const (
	IROpConst       IROp = "const"
	IROpUnary       IROp = "unary"
	IROpBinary      IROp = "binary"
	IROpCall        IROp = "call"
	IROpMethodCall  IROp = "method_call"
	IROpMember      IROp = "member"
	IROpIndex       IROp = "index"
	IROpStructLit   IROp = "struct_lit"
	IROpArrayLit    IROp = "array_lit"
	IROpMapLit      IROp = "map_lit"
	IROpCast        IROp = "cast"
	IROpMatch       IROp = "match"
	IROpSpawn       IROp = "spawn"
	IROpAwait       IROp = "await"
	IROpRaise       IROp = "raise"
	IROpIterator    IROp = "iterator"
	IROpRange       IROp = "range"
	IROpPlaceholder IROp = "placeholder"
	IROpIsNil       IROp = "is_nil"
	IROpIsError     IROp = "is_error"
	IROpAsError     IROp = "as_error"
	IROpPropagate   IROp = "propagate"
)

type IRInstruction interface {
	irInstructionNode()
}

type IRNoop struct {
	Source ast.Node
}

func (*IRNoop) irInstructionNode() {}

// IRCompute represents a pure computation that cannot raise runtime errors.
type IRCompute struct {
	Dest     *IRValue
	Op       IROp
	Args     []IRValueRef
	Operator string
	Source   ast.Node
}

func (*IRCompute) irInstructionNode() {}

// IRInvoke represents a computation that may produce a runtime error.
// Error must be non-nil when the operation can fail.
type IRInvoke struct {
	Value  *IRValue
	Error  *IRValue
	Op     IROp
	Callee IRValueRef
	Args   []IRValueRef
	Source ast.Node
}

func (*IRInvoke) irInstructionNode() {}

// IRLoad reads a local slot into a value.
type IRLoad struct {
	Dest   *IRValue
	Slot   *IRSlot
	Source ast.Node
}

func (*IRLoad) irInstructionNode() {}

// IRStore writes a value into a local slot.
type IRStore struct {
	Slot   *IRSlot
	Value  IRValueRef
	Source ast.Node
}

func (*IRStore) irInstructionNode() {}

// IRDestructure represents pattern destructuring with explicit error flow.
type IRDestructure struct {
	Pattern  ast.Pattern
	Value    IRValueRef
	Error    *IRValue
	Source   ast.Node
	Bindings map[ast.Node]*IRSlot
}

func (*IRDestructure) irInstructionNode() {}

// IRIterNext advances an iterator, producing a value, done flag, and error.
type IRIterNext struct {
	Iterator IRValueRef
	Value    *IRValue
	Done     *IRValue
	Error    *IRValue
	Source   ast.Node
}

func (*IRIterNext) irInstructionNode() {}

// IRSpawn represents a deferred spawn with an explicit body function.
type IRSpawn struct {
	Value    *IRValue
	Error    *IRValue
	Body     *IRFunction
	Captures []*IRSlot
	Source   ast.Node
}

func (*IRSpawn) irInstructionNode() {}

type IRArrayLiteral struct {
	Dest     *IRValue
	Elements []IRValueRef
	Source   ast.Node
}

func (*IRArrayLiteral) irInstructionNode() {}

type IRStructField struct {
	Name        string
	Value       IRValueRef
	IsShorthand bool
}

type IRStructLiteral struct {
	Dest          *IRValue
	StructName    string
	Positional    bool
	Fields        []IRStructField
	Updates       []IRValueRef
	TypeArguments []ast.TypeExpression
	Source        ast.Node
}

func (*IRStructLiteral) irInstructionNode() {}

type IRMapElement interface {
	irMapElementNode()
}

type IRMapEntry struct {
	Key   IRValueRef
	Value IRValueRef
}

func (IRMapEntry) irMapElementNode() {}

type IRMapSpread struct {
	Value IRValueRef
}

func (IRMapSpread) irMapElementNode() {}

type IRMapLiteral struct {
	Dest     *IRValue
	Elements []IRMapElement
	Source   ast.Node
}

func (*IRMapLiteral) irInstructionNode() {}

type IRStringInterpolation struct {
	Dest   *IRValue
	Parts  []IRValueRef
	Source ast.Node
}

func (*IRStringInterpolation) irInstructionNode() {}

type IRIteratorLiteral struct {
	Dest        *IRValue
	BindingName string
	Body        *IRFunction
	Captures    []*IRSlot
	Source      ast.Node
}

func (*IRIteratorLiteral) irInstructionNode() {}

type IRBlock struct {
	Label        string
	Instructions []IRInstruction
	Terminator   IRTerminator
}

func (b *IRBlock) Append(instr IRInstruction) error {
	if b.Terminator != nil {
		return fmt.Errorf("compiler: cannot append instruction after terminator")
	}
	b.Instructions = append(b.Instructions, instr)
	return nil
}

func (b *IRBlock) SetTerminator(term IRTerminator) error {
	if term == nil {
		return fmt.Errorf("compiler: terminator is nil")
	}
	if b.Terminator != nil {
		return fmt.Errorf("compiler: terminator already set")
	}
	b.Terminator = term
	return nil
}

type IRTerminator interface {
	irTerminatorNode()
}

type IRReturn struct {
	Value  IRValueRef
	Source ast.Node
}

func (*IRReturn) irTerminatorNode() {}

type IRJump struct {
	Target string
}

func (*IRJump) irTerminatorNode() {}

type IRBranch struct {
	Condition  IRValueRef
	TrueLabel  string
	FalseLabel string
}

func (*IRBranch) irTerminatorNode() {}

// IRCheck branches based on an error value produced by IRInvoke/IRDestructure.
type IRCheck struct {
	Error    IRValueRef
	OkLabel  string
	ErrLabel string
}

func (*IRCheck) irTerminatorNode() {}

type IRUnreachable struct{}

func (*IRUnreachable) irTerminatorNode() {}
