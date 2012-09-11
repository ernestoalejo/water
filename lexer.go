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
)

var itemNames = map[itemType]string{
	itemError:      "ERROR",
	itemEOF:        "EOF",
	itemLeftParen:  "(",
	itemRightParen: ")",
	itemCall:   "call",
	itemNumber:     "number",
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
		state: lexText,
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

func lexText(l *lexer) stateFn {
	for {
		c := l.next()
		if c == '(' {
			return lexLeftParen
		} else if c == eof {
			break
		} else if isSpace(c) {
			l.ignore()
			continue
		} else {
			return l.errorf("input not expected: %s", l.input[l.start:l.pos])
		}
	}

	l.emit(itemEOF)
	return nil
}

func lexLeftParen(l *lexer) stateFn {
	l.emit(itemLeftParen)
	return lexInsideParen
}

func lexRightParen(l *lexer) stateFn {
	l.emit(itemRightParen)
	return lexText
}

func lexInsideParen(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		return l.errorf("unclosed parenthesis")

	case isSpace(r):
		l.ignore()

	case r == '+' || r == '-':
		if c := l.peek(); isSpace(c) || c == ')' {
			l.backup()
			return lexCall
		}
		fallthrough

	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber

	case r == ')':
		return lexRightParen

	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}

	return lexInsideParen
}

func lexCall(l *lexer) stateFn {
	l.next()
	l.emit(itemCall)

	return lexInsideParen
}

func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %s", l.input[l.start:l.pos])
	}

	l.emit(itemNumber)

	return lexInsideParen
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
