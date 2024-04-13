package main

import (
	"strconv"
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
	Eof TokenType = iota
	Whitespace
	String
	Number
	Keyword
	Identifier
	Operator
	Wildcard
	Comma
	LeftParenthesis
	RightParenthesis
	UnknownTokenType
)

type Token struct {
	Type  TokenType
	Value interface{}
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
		if token.Type == Eof {
			break
		}
		if token.Type == Whitespace {
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
		return l.createToken(Eof)
	// PARENTHESIS
	case l.matchChar('('):
		return l.createToken(LeftParenthesis)
	case l.matchChar(')'):
		return l.createToken(RightParenthesis)
	// WHITESPACE
	case l.matchCharFunc(unicode.IsSpace):
		for l.matchCharFunc(unicode.IsSpace) {
			continue
		}
		return l.createToken(Whitespace)
	// NUMBER
	case l.matchCharFunc(unicode.IsNumber):
		for l.matchCharFunc(unicode.IsNumber) {
			continue
		}
		return l.createToken(Number)
	// IDENTIFIER OR KEYWORD
	case l.matchCharFunc(isLetterOrUnderscore):
		for l.matchCharFunc(isAlphanumericOrUnderscore) {
			continue
		}
		if stringIsKeyword(l.currString()) {
			return l.createToken(Keyword)
		}
		return l.createToken(Identifier)
	// STRING
	case l.matchChar('\''):
		for l.matchCharFunc(func(a rune) bool { return a != '\'' }) {
			continue
		}
		l.cursor++
		value := l.input[l.currTokenStart+1 : l.cursor-1]
		return Token{Type: String, Value: value}
	// COMMA
	case l.matchChar(','):
		return l.createToken(Comma)
	// WILDCARD
	case l.matchChar('*'):
		return l.createToken(Wildcard)
	// OPERATOR
	case stringIsOperator(l.input[l.currTokenStart : l.cursor+1]):
		for stringIsOperator(l.input[l.currTokenStart : l.cursor+1]) {
			l.cursor++
		}
		return l.createToken(Operator)
	default:
		return l.createToken(UnknownTokenType)
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
	if tokenType == Number {
		value, _ := strconv.Atoi(l.currString())
		return Token{Type: tokenType, Value: value}
	}
	return Token{Type: tokenType, Value: l.currString()}
}
