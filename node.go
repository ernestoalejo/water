package main

import (
	"fmt"
)

type Node interface {
	String() string
}

// ========================================================

type ListNode struct {
	Nodes []Node
}

func (n *ListNode) String() string {
	return fmt.Sprintf("list node containing %d nodes", len(n.Nodes))
}

// ========================================================

type CallNode struct {
	Name string
	Args []Node
}

func (n *CallNode) String() string {
	return fmt.Sprintf("call node to function %s with %d args", n.Name, len(n.Args))
}

// ========================================================

type NumberNode struct {
	Text string

	IsInt, IsUint bool

	Int64  int64
	Uint64 uint64
}

func (n *NumberNode) String() string {
	return fmt.Sprintf("number node with the value of %s", n.Text)
}

// ========================================================

type StringNode struct {
	Text string
}

func (n *StringNode) String() string {
	return fmt.Sprintf("string node containing ```%s```", n.Text)
}

// ========================================================

type VarNode struct {
	Name string
}

func (n *VarNode) String() string {
	return fmt.Sprintf("variable node with name %s", n.Name)
}

// ========================================================

type DefineNode struct {
	Variable *VarNode
	Value    Node
}

func (n *DefineNode) String() string {
	return fmt.Sprintf("define a %s", n.Variable)
}

// ========================================================

type SetNode struct {
	Variable *VarNode
	Value    Node
}

func (n *SetNode) String() string {
	return fmt.Sprintf("set a %s", n.Variable)
}

// ========================================================

type IfNode struct {
	Test   Node
	Conseq Node
	Alt    Node
}

func (n *IfNode) String() string {
	return fmt.Sprintf("if node")
}

// ========================================================

type BoolNode struct {
	Value bool
}

func (n *BoolNode) String() string {
	return fmt.Sprintf("bool node with the value of %v", n.Value)
}

// ========================================================

type BeginNode struct {
	Nodes []Node
}

func (n *BeginNode) String() string {
	return fmt.Sprintf("begin node with %d things to execute", len(n.Nodes))
}

// ========================================================

type LambdaNode struct {
	Args []Node // always a *VarNode
	Body Node   // always a *CallNode
}

func (n *LambdaNode) String() string {
	return fmt.Sprintf("lambda node with arity %d", len(n.Args))
}
