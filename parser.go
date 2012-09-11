package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
)

type parser struct {
	Root *ListNode

	lex *lexer

	token  item
	stored bool
}

func Parse(r io.Reader) (p *parser, err error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p = &parser{
		Root: newList(),
		lex:  NewLexer(string(contents)),
	}

	defer p.recover(&err)
	go p.lex.emitItems()

	for {
		item := p.peek()
		if item.t == itemEOF {
			break
		}

		p.Root.Nodes = append(p.Root.Nodes, p.parseAction())
	}

	return
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

func (p *parser) parseAction() Node {
	p.expect(itemLeftParen, "action")

	token := p.peek()
	switch token.t {
	case itemCall:
		return p.parseCall()
	}

	p.errorf("token not expected: %s", token)
	return nil
}

func (p *parser) parseCall() Node {
	c := newCall(p.next().value)
	for {
		switch p.peek().t {
		case itemRightParen:
			p.next()
			return c

		case itemNumber:
			c.Args = append(c.Args, p.parseNumber())
		}
	}

	panic("not reached")
}

func (p *parser) parseNumber() Node {
	item := p.expect(itemNumber, "number")

	n, err := newNumber(item.value)
	if err != nil {
		p.errorf("wrong number value %s: %s", item.value, err)
	}

	return n
}

func (p *parser) recover(errp *error) {
	if e := recover(); e != nil {
		*errp = fmt.Errorf("%s", e)
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
	}
}
