package parser

import (
	"fmt"
	"github.com/cowlet/moncow/ast"
	"github.com/cowlet/moncow/lexer"
	"github.com/cowlet/moncow/token"
)

type Parser struct {
	l            *lexer.Lexer
	currentToken token.Token
	peekToken    token.Token
	errors       []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	/* Read two tokens into current and peek */
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) tokenError(t token.TokenType) {
	msg := fmt.Sprintf("Expected token type %s, got %s instead",
		t, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) validateToken(t token.TokenType) token.Token {
	tok := p.currentToken
	if p.currentToken.Type != t {
		p.tokenError(t)
	}
	p.nextToken()
	return tok
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	/* Expect LET, IDENT, ASSIGN, <expression>, SEMI */
	let := p.validateToken(token.LET)
	name := p.parseIdentifier()
	p.validateToken(token.ASSIGN)

	/* TODO: revisit */
	for p.currentToken.Type != token.SEMI {
		p.nextToken()
	}
	//expression := p.parseExpression()
	return &ast.LetStatement{Token: let, Name: name} // skip Value
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	name := p.validateToken(token.IDENT)
	ident := &ast.Identifier{Token: name, Value: name.Literal}
	return ident
}

func (p *Parser) parseStatement() ast.Statement {
	fmt.Printf("Parsing statement beginning '%s'\n", p.currentToken.Type)
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	default:
		return nil
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currentToken.Type != token.EOF {
		statement := p.parseStatement()

		if statement != nil {
			fmt.Printf("Parsed statement %#v\n", statement)
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}
	return program
}
