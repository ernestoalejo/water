package main

import (
	"fmt"
	"strconv"
)

type Node interface {
	String() string
}

// ========================================================

type ListNode struct {
	Nodes []Node
}

func newList() *ListNode {
	return &ListNode{}
}

func (n *ListNode) String() string {
	return fmt.Sprintf("list node containing %d nodes", len(n.Nodes))
}

// ========================================================

type CallNode struct {
	Name string
	Args []Node
}

func newCall(name string) *CallNode {
	return &CallNode{
		Name: name,
		Args: []Node{},
	}
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

func newNumber(text string) (*NumberNode, error) {
	n := &NumberNode{Text: text}

	u, err := strconv.ParseUint(text, 0, 64)
	if err == nil {
		n.IsUint = true
		n.Uint64 = u
	}

	i, err := strconv.ParseInt(text, 0, 64)
	if err == nil {
		n.IsInt = true
		n.Int64 = i

		if i == 0 {
			n.IsUint = true
			n.Uint64 = u
		}
	}

	if !n.IsUint || !n.IsInt {
		return nil, fmt.Errorf("illegal number syntax: %s", text)
	}

	return n, nil
}

func (n *NumberNode) String() string {
	return fmt.Sprintf("number node with the value of %s", n.Text)
}

// ========================================================

type StringNode struct {
	Text string
}

func newString(text string) (*StringNode, error) {
	n := new(StringNode)

	var err error
	n.Text, err = strconv.Unquote(text)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n *StringNode) String() string {
	return fmt.Sprintf("string node containing ```%s```", n.Text)
}

// ========================================================

type VarNode struct {
	Name string
}

func newVar(name string) *VarNode {
	return &VarNode{
		Name: name,
	}
}

func (n *VarNode) String() string {
	return fmt.Sprintf("variable node with name %s", n.Name)
}

// ========================================================

type DefineNode struct {
	Variable *VarNode
	Value    Node
}

func newDefine(variable *VarNode, value Node) *DefineNode {
	return &DefineNode{
		Variable: variable,
		Value:    value,
	}
}

func (n *DefineNode) String() string {
	return fmt.Sprintf("define a %s", n.Variable)
}

// ========================================================

type SetNode struct {
	Variable *VarNode
	Value    Node
}

func newSet(variable *VarNode, value Node) *SetNode {
	return &SetNode{
		Variable: variable,
		Value:    value,
	}
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

func newIf(test, consequence, alternative Node) *IfNode {
	return &IfNode{
		Test:   test,
		Conseq: consequence,
		Alt:    alternative,
	}
}

func (n *IfNode) String() string {
	return fmt.Sprintf("if node")
}

// ========================================================

type BoolNode struct {
	Value bool
}

func newBool(value bool) *BoolNode {
	return &BoolNode{
		Value: value,
	}
}

func (n *BoolNode) String() string {
	return fmt.Sprintf("bool node with the value of %v", n.Value)
}

// ========================================================

type BeginNode struct {
	Nodes []Node
}

func newBegin(nodes []Node) *BeginNode {
	return &BeginNode{
		Nodes: nodes,
	}
}

func (n *BeginNode) String() string {
	return fmt.Sprintf("begin node with %d things to execute", len(n.Nodes))
}
