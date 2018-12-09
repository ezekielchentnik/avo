package avo

import (
	"github.com/mmcloughlin/avo/gotypes"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

type Asm interface {
	Asm() string
}

// GoType represents a Golang type.
type GoType interface{}

// Parameter represents a parameter to an assembly function.
type Parameter struct {
	Name string
	Type GoType
}

type Operand interface {
	Asm
}

type Node interface {
	node()
}

type Label string

func (l Label) node() {}

// Instruction is a single instruction in a function.
type Instruction struct {
	Opcode   string
	Operands []operand.Op

	Inputs  []operand.Op
	Outputs []operand.Op

	IsTerminal    bool
	IsBranch      bool
	IsConditional bool

	// CFG.
	Pred []*Instruction
	Succ []*Instruction

	// LiveIn/LiveOut are sets of live register IDs pre/post execution.
	LiveIn  reg.Set
	LiveOut reg.Set
}

func (i *Instruction) node() {}

func (i Instruction) TargetLabel() *Label {
	if !i.IsBranch {
		return nil
	}
	if len(i.Operands) == 0 {
		return nil
	}
	if ref, ok := i.Operands[0].(operand.LabelRef); ok {
		lbl := Label(ref)
		return &lbl
	}
	return nil
}

func (i Instruction) Registers() []reg.Register {
	var rs []reg.Register
	for _, op := range i.Operands {
		rs = append(rs, operand.Registers(op)...)
	}
	return rs
}

func (i Instruction) InputRegisters() []reg.Register {
	var rs []reg.Register
	for _, op := range i.Inputs {
		rs = append(rs, operand.Registers(op)...)
	}
	for _, op := range i.Outputs {
		if operand.IsMem(op) {
			rs = append(rs, operand.Registers(op)...)
		}
	}
	return rs
}

func (i Instruction) OutputRegisters() []reg.Register {
	var rs []reg.Register
	for _, op := range i.Outputs {
		if r, ok := op.(reg.Register); ok {
			rs = append(rs, r)
		}
	}
	return rs
}

// File represents an assembly file.
type File struct {
	Functions []*Function
}

func NewFile() *File {
	return &File{}
}

// Function represents an assembly function.
type Function struct {
	Name      string
	Signature *gotypes.Signature

	Nodes []Node

	// LabelTarget maps from label name to the following instruction.
	LabelTarget map[Label]*Instruction

	// Register allocation.
	Allocation reg.Allocation
}

func NewFunction(name string) *Function {
	return &Function{
		Name:      name,
		Signature: gotypes.NewSignatureVoid(),
	}
}

func (f *Function) SetSignature(s *gotypes.Signature) {
	f.Signature = s
}

func (f *Function) AddInstruction(i *Instruction) {
	f.AddNode(i)
}

func (f *Function) AddLabel(l Label) {
	f.AddNode(l)
}

func (f *Function) AddNode(n Node) {
	f.Nodes = append(f.Nodes, n)
}

// Instructions returns just the list of instruction nodes.
func (f *Function) Instructions() []*Instruction {
	var is []*Instruction
	for _, n := range f.Nodes {
		i, ok := n.(*Instruction)
		if ok {
			is = append(is, i)
		}
	}
	return is
}

// Stub returns the Go function declaration.
func (f *Function) Stub() string {
	return "func " + f.Name + f.Signature.String()
}

// FrameBytes returns the size of the stack frame in bytes.
func (f *Function) FrameBytes() int {
	// TODO(mbm): implement Function.FrameBytes()
	return 0
}

// ArgumentBytes returns the size of the arguments in bytes.
func (f *Function) ArgumentBytes() int {
	return f.Signature.Bytes()
}
