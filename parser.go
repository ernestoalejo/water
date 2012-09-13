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

	token := p.peek()
	switch token.t {
	case itemCall:
		return p.parseCall()

	case itemDefine:
		return p.parseDefine()
	}

	p.errorf("token not expected: %s", token)
	return nil
}

func (p *parser) parseCall() Node {
	c := newCall(p.next().value)
	for {
		switch p.peek().t {
		// The right delimiter finishes the call parsing
		case itemRightParen:
			p.next()
			return c

		// Parse a number as the arg
		case itemNumber:
			c.Args = append(c.Args, p.parseNumber())

		// Parse a string as the arg
		case itemString:
			c.Args = append(c.Args, p.parseString())

		// Parse a var as the arg
		case itemVar:
			c.Args = append(c.Args, p.parseVar())

		case itemLeftParen:
			p.next()

			if p.peek().t == itemDefine {
				// The args it's the result of a definition
				c.Args = append(c.Args, p.parseDefine())
			} else {
				// The arg it's another function call
				c.Args = append(c.Args, p.parseCall())
			}

		default:
			p.errorf("unexpected token in call to %s: %s", c.Name, p.peek().t)
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

func (p *parser) parseString() Node {
	item := p.expect(itemString, "string")

	n, err := newString(item.value)
	if err != nil {
		p.errorf("%s", err)
	}

	return n
}

func (p *parser) parseDefine() Node {
	p.expect(itemDefine, "define")
	name := p.expect(itemVar, "define")

	var init Node
	switch item := p.peek(); item.t {
	case itemNumber:
		init = p.parseNumber()

	case itemString:
		init = p.parseString()

	case itemLeftParen:
		p.next()
		init = p.parseCall()

	default:
		p.errorf("cannot init a variable with this kind of value: %s", item.t)
	}

	p.expect(itemRightParen, "define")

	return newDefine(newVar(name.value), init)
}

func (p *parser) parseVar() Node {
	v := p.expect(itemVar, "var")
	return newVar(v.value)
}

func (p *parser) recover(errp *error) {
	if e := recover(); e != nil {
		*errp = fmt.Errorf("%s", e)
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
	}
}

func Parse(r io.Reader) (l *ListNode, err error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p := &parser{
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

	l = p.Root

	return
}
