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

	switch token := p.peek(); token.t {
	case itemOperator:
		return p.parseOperator()
	}

	panic("not reached")
}

func (p *parser) parseOperator() Node {
	op := newOperator(p.next().value)
	for {
		switch p.peek().t {
		case itemRightParen:
			p.next()
			if len(op.Values) == 0 {
				p.errorf("values expected for the operator: %s", op.Operator)
			}
			return op

		case itemNumber:
			op.Values = append(op.Values, p.parseNumber())
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

func Parse(r io.Reader) (p *parser, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s", e)
			if _, ok := e.(runtime.Error); ok {
				panic(e)
			}
		}
	}()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p = &parser{
		Root: newList(),
		lex:  NewLexer(string(contents)),
	}

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
