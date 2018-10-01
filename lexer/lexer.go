package lexer

import (
	"fmt"
	"github.com/cowlet/moncow/token"
	"unicode"
	"unicode/utf8"
)

var _ = fmt.Printf // TODO: delete when done

type Lexer struct {
	input        string
	position     int  // current position in input (current rune start)
	readPosition int  // current reading position in input (after current rune)
	ch           rune // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readRune() // initialize the character pointers
	return l
}

func (l *Lexer) readRune() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
		l.readPosition += 1
	} else {
		runeValue, width := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = runeValue
		l.position = l.readPosition
		l.readPosition += width
	}
}

func (l *Lexer) peekRune() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	runeValue, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return runeValue
}

func (l *Lexer) readIdentifier() string {
	startPos := l.position
	for isLetter(l.ch) {
		l.readRune()
	}
	return l.input[startPos:l.position]
}

func (l *Lexer) readNumber() string {
	startPos := l.position
	for isDigit(l.ch) {
		l.readRune()
	}
	return l.input[startPos:l.position]
}

func (l *Lexer) skipWhitespace() {
	for unicode.Is(unicode.White_Space, l.ch) {
		l.readRune()
	}
}

func isLetter(ch rune) bool {
	return unicode.In(ch, unicode.Letter, unicode.Symbol) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.In(ch, unicode.Number)
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	//fmt.Printf("Inspecting rune %#U\n", l.ch)

	switch l.ch {
	case '=':
		if l.peekRune() == '=' {
			l.readRune() // consume one here, the other below
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekRune() == '=' {
			l.readRune() // consume one here, the other below
			tok = token.Token{Type: token.NEQ, Literal: "!="}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case ';':
		tok = newToken(token.SEMI, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.MULT, l.ch)
	case '/':
		tok = newToken(token.DIV, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok = token.Token{Type: token.INT, Literal: l.readNumber()}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readRune()
	return tok
}
