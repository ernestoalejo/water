package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strconv"
)

func Parse(r io.Reader) (l *ListNode, err error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p := &parser{
		Root: new(ListNode),
		lex:  NewLexer(string(contents)),
	}

	defer p.recover(&err)
	go p.lex.emitItems()

	for {
		item := p.peek()
		if item.t == itemEOF {
			break
		}

		p.Root.Nodes = append(p.Root.Nodes, p.parseExpression())
	}

	l = p.Root

	return
}

// ========================================================

type parser struct {
	Root *ListNode

	lex *lexer

	token  item
	stored bool
}

func (p *parser) expect(expected itemType, context string) item {
	token := p.next()
	if token.t != expected {
		p.errorf("expected %s in %s; got %s", expected, context, token)
	}

	return token
}

func (p *parser) peek() item {
	i := p.next()
	p.backup()
	return i
}

func (p *parser) backup() {
	p.stored = true
}

func (p *parser) next() item {
	if p.stored {
		p.stored = false
		return p.token
	}

	p.token = <-p.lex.items
	return p.token
}

func (p *parser) errorf(format string, args ...interface{}) {
	p.Root = nil
	panic(fmt.Sprintf(format, args...))
}

func (p *parser) recover(errp *error) {
	if e := recover(); e != nil {
		*errp = fmt.Errorf("%s", e)
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
	}
}

func (p *parser) parseCall() Node {
	p.expect(itemLeftParen, "call")

	name := p.peek().value

	// Parse some call-like structures that are treated in a
	// different way by the lang
	switch name {
	case "define":
		return p.parseDefine()

	case "set":
		return p.parseSet()

	case "if":
		return p.parseIf()

	case "begin":
		return p.parseBegin()
	}

	c := &CallNode{
		Name: p.next().value,
		Args: make([]Node, 0),
	}
	for {
		if p.peek().t == itemRightParen {
			p.next()
			return c
		}

		c.Args = append(c.Args, p.parseExpression())
	}

	panic("not reached")
}

func (p *parser) parseNumber() Node {
	item := p.expect(itemNumber, "number")

	n := &NumberNode{Text: item.value}

	sign := 1
	if item.value[0] == '+' || item.value[0] == '-' {
		if item.value[0] == '-' {
			sign = -1
		}

		item.value = item.value[1:]
	}

	u, err := strconv.ParseUint(item.value, 0, 64)
	if err == nil {
		n.IsUint = true
		n.Uint64 = uint64(sign) * u
	}

	i, err := strconv.ParseInt(item.value, 0, 64)
	if err == nil {
		n.IsInt = true
		n.Int64 = int64(sign) * i

		if i == 0 {
			n.IsUint = true
			n.Uint64 = uint64(sign) * u
		}
	}

	if !n.IsUint || !n.IsInt {
		p.errorf("illegal number syntax: %s", item.value)
	}

	return n
}

func (p *parser) parseString() Node {
	item := p.expect(itemString, "string")

	n := new(StringNode)

	var err error
	n.Text, err = strconv.Unquote(item.value)
	if err != nil {
		p.errorf("cannot unquote the string literal: %s", err)
	}

	return n
}

func (p *parser) parseDefine() Node {
	p.expect(itemCall, "define")
	name := p.parseVar()
	init := p.parseExpression()
	p.expect(itemRightParen, "define")

	return &DefineNode{
		Variable: name.(*VarNode),
		Value:    init,
	}
}

func (p *parser) parseSet() Node {
	p.expect(itemCall, "set")
	name := p.parseVar()
	init := p.parseExpression()
	p.expect(itemRightParen, "set")

	return &SetNode{
		Variable: name.(*VarNode),
		Value:    init,
	}
}

func (p *parser) parseVar() Node {
	item := p.expect(itemVar, "var")
	return &VarNode{Name: item.value}
}

func (p *parser) parseIf() Node {
	p.expect(itemCall, "if")
	n := &IfNode{
		Test:   p.parseExpression(),
		Conseq: p.parseExpression(),
		Alt:    p.parseExpression(),
	}
	p.expect(itemRightParen, "if")

	return n
}

func (p *parser) parseBool() Node {
	it := p.expect(itemBool, "bool")

	if it.value == "#t" || it.value == "#f" {
		return &BoolNode{Value: it.value == "#t"}
	}

	p.errorf("incorrect boolean value, should be #t or #f: ", it)
	panic("not reached")
}

func (p *parser) parseExpression() Node {
	switch item := p.peek(); item.t {
	case itemNumber:
		return p.parseNumber()

	case itemString:
		return p.parseString()

	case itemLeftParen:
		return p.parseCall()

	case itemBool:
		return p.parseBool()

	case itemVar:
		return p.parseVar()

	default:
		p.errorf("cannot use this kind of value as a expression: %s", item)
	}

	panic("not reached")
}

func (p *parser) parseBegin() Node {
	p.expect(itemCall, "begin")

	nodes := make([]Node, 0)
	for {
		if item := p.peek(); item.t == itemRightParen {
			break
		}

		nodes = append(nodes, p.parseExpression())
	}

	p.expect(itemRightParen, "begin")

	if len(nodes) == 0 {
		p.errorf("begin sentence without expressions")
	}

	return &BeginNode{Nodes: nodes}
}
