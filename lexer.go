package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

type stateFn func(*lexer) stateFn

// ========================================================

type item struct {
	t     itemType
	value string
}

func (i item) String() string {
	return fmt.Sprintf("%s => %s", i.t, i.value)
}

// ========================================================

type itemType int

func (i itemType) String() string {
	name, ok := itemNames[i]
	if ok {
		return name
	}

	return fmt.Sprintf("%d", int(i))
}

const (
	itemError itemType = iota
	itemEOF
	itemLeftParen
	itemRightParen
	itemCall
	itemNumber
	itemString
	itemVar
	itemDefine
)

var itemNames = map[itemType]string{
	itemError:      "ERROR",
	itemEOF:        "EOF",
	itemLeftParen:  "(",
	itemRightParen: ")",
	itemCall:       "call",
	itemNumber:     "number",
	itemString:     "string",
	itemVar:        "variable",
	itemDefine:     "define",
}

// ========================================================

type lexer struct {
	input             string
	state             stateFn
	pos, start, width int
	items             chan item
}

func NewLexer(input string) *lexer {
	return &lexer{
		input: input,
		state: lexCode,
		items: make(chan item),
	}
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		l.start = l.pos
		return eof
	}

	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width

	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()

	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) scanNumber() bool {
	l.accept("+-")

	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits += "abcdefABCDEF"
	}
	l.acceptRun(digits)

	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}

	return true
}

func (l *lexer) emitItems() {
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.items)
}

func lexLeftParen(l *lexer) stateFn {
	l.emit(itemLeftParen)
	return lexCall
}

func lexRightParen(l *lexer) stateFn {
	l.emit(itemRightParen)
	return lexCode
}

func lexCode(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil

	case isSpace(r):
		l.ignore()
		return lexCode

	case r == '+' || r == '-':
		if c := l.peek(); '0' <= c && c <= '9' {
			l.backup()
			return lexNumber
		}

		l.backup()
		return lexVar

	case '0' <= r && r <= '9':
		l.backup()
		return lexNumber

	case r == ')':
		return lexRightParen

	case r == '(':
		return lexLeftParen

	case r == '"' || r == '\'':
		l.backup()
		return lexString

	default:
		l.backup()
		return lexVar
	}

	panic("not reached")
}

func lexCall(l *lexer) stateFn {
	// Ignore all the spaces before the function name
	for {
		r := l.next()
		if !isSpace(r) {
			l.backup()
			break
		}
	}

	if strings.HasPrefix(l.input[l.start:], "define") {
		l.pos += len("define")
		l.emit(itemDefine)
		return lexCode
	}

	// Scan the name
	r := l.next()
	for r != ' ' && r != ')' {
		r = l.next()

		if r == eof {
			return l.errorf("eof not expected inside a method call")
		}
	}
	l.backup()

	// If there's no name, it's an illegal call
	if l.start == l.pos {
		return l.errorf("illegal function name")
	}

	// Emit the correct token
	l.emit(itemCall)

	return lexCode
}

func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %s", l.input[l.start:l.pos])
	}

	l.emit(itemNumber)

	return lexCode
}

func lexString(l *lexer) stateFn {
	delim := l.next()
	for {
		r := l.next()
		if r == delim {
			break
		} else if r == eof {
			l.errorf("eof not expected inside a string")
		}
	}

	l.emit(itemString)
	return lexCode
}

func lexVar(l *lexer) stateFn {
	// Scan the name
	r := l.next()
	for r != ' ' && r != ')' {
		r = l.next()

		if r == eof {
			return l.errorf("eof not expected inside a variable name")
		}
	}
	l.backup()

	// If there's no name, it's an illegal variable
	if l.start == l.pos {
		return l.errorf("illegal variable name")
	}

	// Emit the correct token
	l.emit(itemVar)

	return lexCode
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\n', '\t', '\r':
		return true
	}
	return false
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
