package main

import (
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input          string
	cursor         int
	currTokenStart int
}

type TokenType uint

const (
	EOF TokenType = iota
	WHITESPACE
	STRING
	NUMBER
	KEYWORD
	IDENTIFIER
	OPERATOR
	WILDCARD
	COMMA
	UNKNOWN
)

type Token struct {
	Type  TokenType
	Value string
}

func NewLexer() Lexer {
	return Lexer{}
}

func (l *Lexer) Scan(input string) []Token {
	var tokens []Token
	l.input = input
	l.cursor = 0
	for {
		token := l.scanNext()
		if token.Type == EOF {
			break
		}
		if token.Type == WHITESPACE {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func (l *Lexer) scanNext() Token {
	l.currTokenStart = l.cursor
	switch {
	// EOF
	case l.cursor >= len(l.input):
		return l.createToken(EOF)
	// WHITESPACE
	case l.matchCharFunc(unicode.IsSpace):
		for l.matchCharFunc(unicode.IsSpace) {
			continue
		}
		return l.createToken(WHITESPACE)
	// NUMBER
	case l.matchCharFunc(unicode.IsNumber):
		for l.matchCharFunc(unicode.IsNumber) {
			continue
		}
		return l.createToken(NUMBER)
	// IDENTIFIER OR KEYWORD
	case l.matchCharFunc(isLetterOrUnderscore):
		for l.matchCharFunc(isAlphanumericOrUnderscore) {
			continue
		}
		if stringIsKeyword(l.currString()) {
			return l.createToken(KEYWORD)
		}
		return l.createToken(IDENTIFIER)
	// STRING
	case l.matchChar('\''):
		for l.matchCharFunc(func(a rune) bool { return a != '\'' }) {
			continue
		}
		l.cursor++
		value := l.input[l.currTokenStart+1 : l.cursor-1]
		return Token{Type: STRING, Value: value}
	// COMMA
	case l.matchChar(','):
		return l.createToken(COMMA)
	// WILDCARD
	case l.matchChar('*'):
		return l.createToken(WILDCARD)
	// OPERATOR
	case stringIsOperator(l.input[l.currTokenStart : l.cursor+1]):
		for stringIsOperator(l.input[l.currTokenStart : l.cursor+1]) {
			l.cursor++
		}
		return l.createToken(OPERATOR)
	default:
		return l.createToken(UNKNOWN)
	}
}

func (l *Lexer) matchChar(value rune) bool {
	if l.cursor >= len(l.input) {
		return false
	}
	char, _ := utf8.DecodeRuneInString(l.input[l.cursor:])
	if char == value {
		l.cursor++
		return true
	}
	return false
}

func (l *Lexer) matchCharFunc(cb func(char rune) bool) bool {
	if l.cursor >= len(l.input) {
		return false
	}
	char, _ := utf8.DecodeRuneInString(l.input[l.cursor:])
	if cb(char) {
		l.cursor++
		return true
	}
	return false
}

func (l Lexer) currString() string {
	return l.input[l.currTokenStart:l.cursor]
}

func (l Lexer) createToken(tokenType TokenType) Token {
	return Token{Type: tokenType, Value: l.currString()}
}
